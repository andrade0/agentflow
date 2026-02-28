package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/agentflow/agentflow/pkg/types"
)

// OllamaProvider implements the Provider interface for Ollama
type OllamaProvider struct {
	baseURL string
	models  []string
	client  *http.Client
}

// NewOllama creates a new Ollama provider
func NewOllama(cfg Config) *OllamaProvider {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	return &OllamaProvider{
		baseURL: baseURL,
		models:  cfg.Models,
		client: &http.Client{
			Timeout: 5 * time.Minute, // Long timeout for generation
		},
	}
}

func (o *OllamaProvider) Name() string {
	return "ollama"
}

func (o *OllamaProvider) Models() []string {
	return o.models
}

func (o *OllamaProvider) SupportsModel(model string) bool {
	for _, m := range o.models {
		if m == model {
			return true
		}
	}
	return true // Ollama can pull any model
}

// ollamaRequest is the Ollama API request format
type ollamaRequest struct {
	Model    string             `json:"model"`
	Messages []ollamaMessage    `json:"messages"`
	Stream   bool               `json:"stream"`
	Options  *ollamaOptions     `json:"options,omitempty"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaOptions struct {
	Temperature float64 `json:"temperature,omitempty"`
	NumPredict  int     `json:"num_predict,omitempty"`
}

// ollamaResponse is the Ollama API response format
type ollamaResponse struct {
	Model     string        `json:"model"`
	Message   ollamaMessage `json:"message"`
	Done      bool          `json:"done"`
	DoneReason string       `json:"done_reason,omitempty"`
	PromptEvalCount int     `json:"prompt_eval_count,omitempty"`
	EvalCount       int     `json:"eval_count,omitempty"`
}

func (o *OllamaProvider) Complete(ctx context.Context, req types.CompletionRequest) (*types.CompletionResponse, error) {
	// Convert messages to Ollama format
	msgs := make([]ollamaMessage, len(req.Messages))
	for i, m := range req.Messages {
		msgs[i] = ollamaMessage{Role: m.Role, Content: m.Content}
	}

	ollamaReq := ollamaRequest{
		Model:    req.Model,
		Messages: msgs,
		Stream:   false,
	}

	if req.Temperature > 0 || req.MaxTokens > 0 {
		ollamaReq.Options = &ollamaOptions{
			Temperature: req.Temperature,
			NumPredict:  req.MaxTokens,
		}
	}

	body, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama error %d: %s", resp.StatusCode, string(respBody))
	}

	var ollamaResp ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &types.CompletionResponse{
		Content:      ollamaResp.Message.Content,
		Model:        ollamaResp.Model,
		FinishReason: ollamaResp.DoneReason,
		TokensUsed:   ollamaResp.PromptEvalCount + ollamaResp.EvalCount,
	}, nil
}

func (o *OllamaProvider) Stream(ctx context.Context, req types.CompletionRequest) (<-chan types.StreamChunk, error) {
	// Convert messages to Ollama format
	msgs := make([]ollamaMessage, len(req.Messages))
	for i, m := range req.Messages {
		msgs[i] = ollamaMessage{Role: m.Role, Content: m.Content}
	}

	ollamaReq := ollamaRequest{
		Model:    req.Model,
		Messages: msgs,
		Stream:   true,
	}

	if req.Temperature > 0 || req.MaxTokens > 0 {
		ollamaReq.Options = &ollamaOptions{
			Temperature: req.Temperature,
			NumPredict:  req.MaxTokens,
		}
	}

	body, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("ollama error: status %d", resp.StatusCode)
	}

	chunks := make(chan types.StreamChunk)
	go func() {
		defer close(chunks)
		defer resp.Body.Close()

		decoder := json.NewDecoder(resp.Body)
		for {
			var chunk ollamaResponse
			if err := decoder.Decode(&chunk); err != nil {
				if err != io.EOF {
					chunks <- types.StreamChunk{Error: err}
				}
				return
			}
			chunks <- types.StreamChunk{
				Content: chunk.Message.Content,
				Done:    chunk.Done,
			}
			if chunk.Done {
				return
			}
		}
	}()

	return chunks, nil
}
