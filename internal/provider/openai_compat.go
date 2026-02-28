package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/agentflow/agentflow/pkg/types"
)

// OpenAICompatProvider implements Provider for OpenAI-compatible APIs
type OpenAICompatProvider struct {
	name    string
	baseURL string
	apiKey  string
	models  []string
	client  *http.Client
}

// NewOpenAICompat creates a generic OpenAI-compatible provider
func NewOpenAICompat(name string, cfg Config) *OpenAICompatProvider {
	return &OpenAICompatProvider{
		name:    name,
		baseURL: strings.TrimSuffix(cfg.BaseURL, "/"),
		apiKey:  cfg.APIKey,
		models:  cfg.Models,
		client: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

func (o *OpenAICompatProvider) Name() string {
	return o.name
}

func (o *OpenAICompatProvider) Models() []string {
	return o.models
}

func (o *OpenAICompatProvider) SupportsModel(model string) bool {
	for _, m := range o.models {
		if m == model {
			return true
		}
	}
	return false
}

// OpenAI API types
type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	Temperature float64         `json:"temperature,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int           `json:"index"`
		Message      openAIMessage `json:"message"`
		Delta        openAIMessage `json:"delta"`
		FinishReason string        `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func (o *OpenAICompatProvider) Complete(ctx context.Context, req types.CompletionRequest) (*types.CompletionResponse, error) {
	msgs := make([]openAIMessage, len(req.Messages))
	for i, m := range req.Messages {
		msgs[i] = openAIMessage{Role: m.Role, Content: m.Content}
	}

	oaiReq := openAIRequest{
		Model:       req.Model,
		Messages:    msgs,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		Stream:      false,
	}

	body, err := json.Marshal(oaiReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if o.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+o.apiKey)
	}

	resp, err := o.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%s error %d: %s", o.name, resp.StatusCode, string(respBody))
	}

	var oaiResp openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&oaiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(oaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &types.CompletionResponse{
		Content:      oaiResp.Choices[0].Message.Content,
		Model:        oaiResp.Model,
		FinishReason: oaiResp.Choices[0].FinishReason,
		TokensUsed:   oaiResp.Usage.TotalTokens,
	}, nil
}

func (o *OpenAICompatProvider) Stream(ctx context.Context, req types.CompletionRequest) (<-chan types.StreamChunk, error) {
	msgs := make([]openAIMessage, len(req.Messages))
	for i, m := range req.Messages {
		msgs[i] = openAIMessage{Role: m.Role, Content: m.Content}
	}

	oaiReq := openAIRequest{
		Model:       req.Model,
		Messages:    msgs,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		Stream:      true,
	}

	body, err := json.Marshal(oaiReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	if o.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+o.apiKey)
	}

	resp, err := o.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("%s error: status %d", o.name, resp.StatusCode)
	}

	chunks := make(chan types.StreamChunk)
	go func() {
		defer close(chunks)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				chunks <- types.StreamChunk{Done: true}
				return
			}

			var chunk openAIResponse
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}
			if len(chunk.Choices) > 0 {
				chunks <- types.StreamChunk{
					Content: chunk.Choices[0].Delta.Content,
					Done:    chunk.Choices[0].FinishReason != "",
				}
			}
		}
		if err := scanner.Err(); err != nil {
			chunks <- types.StreamChunk{Error: err}
		}
	}()

	return chunks, nil
}
