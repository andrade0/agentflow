// Package types defines shared types for AgentFlow
package types

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`    // system, user, assistant
	Content string `json:"content"` // message content
}

// CompletionRequest is sent to providers
type CompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// CompletionResponse from providers
type CompletionResponse struct {
	Content      string `json:"content"`
	Model        string `json:"model"`
	FinishReason string `json:"finish_reason"`
	TokensUsed   int    `json:"tokens_used"`
}

// StreamChunk for streaming responses
type StreamChunk struct {
	Content string
	Done    bool
	Error   error
}

// ProviderType identifies the LLM provider
type ProviderType string

const (
	ProviderOllama   ProviderType = "ollama"
	ProviderGroq     ProviderType = "groq"
	ProviderTogether ProviderType = "together"
	ProviderOpenAI   ProviderType = "openai"
)
