package agent

import (
	"context"
	"testing"

	"github.com/agentflow/agentflow/pkg/types"
)

// mockProvider implements provider.Provider for testing
type mockProvider struct {
	name     string
	response string
	err      error
}

func (m *mockProvider) Name() string { return m.name }
func (m *mockProvider) Models() []string { return []string{"test-model"} }
func (m *mockProvider) SupportsModel(model string) bool { return true }

func (m *mockProvider) Complete(ctx context.Context, req types.CompletionRequest) (*types.CompletionResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &types.CompletionResponse{
		Content:      m.response,
		Model:        req.Model,
		FinishReason: "stop",
		TokensUsed:   100,
	}, nil
}

func (m *mockProvider) Stream(ctx context.Context, req types.CompletionRequest) (<-chan types.StreamChunk, error) {
	ch := make(chan types.StreamChunk)
	go func() {
		ch <- types.StreamChunk{Content: m.response, Done: true}
		close(ch)
	}()
	return ch, nil
}

func TestNew(t *testing.T) {
	p := &mockProvider{name: "test"}
	a := New(Config{
		Provider: p,
		Model:    "test-model",
	})

	if a == nil {
		t.Fatal("expected non-nil agent")
	}

	if a.ID() == "" {
		t.Error("expected non-empty ID")
	}

	if a.Model() != "test-model" {
		t.Errorf("model = %q, want 'test-model'", a.Model())
	}
}

func TestAgent_WithSystemPrompt(t *testing.T) {
	p := &mockProvider{name: "test"}
	a := New(Config{
		Provider:     p,
		Model:        "test-model",
		SystemPrompt: "You are a helpful assistant.",
	})

	messages := a.Messages()
	if len(messages) != 1 {
		t.Errorf("expected 1 message (system), got %d", len(messages))
	}

	if messages[0].Role != "system" {
		t.Errorf("expected system role, got %q", messages[0].Role)
	}
}

func TestAgent_AddMessage(t *testing.T) {
	p := &mockProvider{name: "test"}
	a := New(Config{Provider: p, Model: "test"})

	a.AddMessage("user", "Hello")
	a.AddMessage("assistant", "Hi there!")

	messages := a.Messages()
	if len(messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(messages))
	}

	if messages[0].Role != "user" || messages[0].Content != "Hello" {
		t.Errorf("message[0] = %+v", messages[0])
	}
}

func TestAgent_ClearHistory(t *testing.T) {
	p := &mockProvider{name: "test"}
	a := New(Config{
		Provider:     p,
		Model:        "test",
		SystemPrompt: "System prompt",
	})

	a.AddMessage("user", "Hello")
	a.AddMessage("assistant", "Hi")

	a.ClearHistory()

	messages := a.Messages()
	if len(messages) != 1 {
		t.Errorf("expected 1 message (system), got %d", len(messages))
	}
	if messages[0].Role != "system" {
		t.Error("expected system message preserved")
	}
}

func TestAgent_Run(t *testing.T) {
	p := &mockProvider{name: "test", response: "Hello, human!"}
	a := New(Config{Provider: p, Model: "test-model"})

	resp, err := a.Run(context.Background(), "Hello")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if resp.Content != "Hello, human!" {
		t.Errorf("content = %q", resp.Content)
	}

	// Check messages updated
	messages := a.Messages()
	if len(messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(messages))
	}
	if messages[0].Role != "user" {
		t.Error("expected user message")
	}
	if messages[1].Role != "assistant" {
		t.Error("expected assistant message")
	}
}

func TestAgent_Metadata(t *testing.T) {
	p := &mockProvider{name: "test"}
	a := New(Config{
		Provider: p,
		Model:    "test",
		Metadata: map[string]string{"task": "coding"},
	})

	if a.GetMetadata("task") != "coding" {
		t.Error("expected metadata 'task' = 'coding'")
	}

	a.SetMetadata("project", "agentflow")
	if a.GetMetadata("project") != "agentflow" {
		t.Error("expected metadata 'project' = 'agentflow'")
	}
}

func TestAgent_Clone(t *testing.T) {
	p := &mockProvider{name: "test"}
	a := New(Config{
		ID:           "original",
		Provider:     p,
		Model:        "test-model",
		SystemPrompt: "System",
	})

	a.AddMessage("user", "Hello")
	a.SetMetadata("key", "value")

	clone := a.Clone("clone-1")

	if clone.ID() != "clone-1" {
		t.Errorf("clone ID = %q", clone.ID())
	}

	if clone.Model() != "test-model" {
		t.Errorf("clone model = %q", clone.Model())
	}

	// Clone should have fresh history (only system prompt)
	messages := clone.Messages()
	if len(messages) != 1 {
		t.Errorf("expected 1 message in clone, got %d", len(messages))
	}

	// Metadata should be copied
	if clone.GetMetadata("key") != "value" {
		t.Error("expected metadata copied")
	}

	// Modifying clone metadata shouldn't affect original
	clone.SetMetadata("key", "modified")
	if a.GetMetadata("key") != "value" {
		t.Error("original metadata should not change")
	}
}

func TestAgent_Stream(t *testing.T) {
	p := &mockProvider{name: "test", response: "Streamed response"}
	a := New(Config{Provider: p, Model: "test"})

	chunks, err := a.Stream(context.Background(), "Test message")
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}

	var content string
	for chunk := range chunks {
		if chunk.Error != nil {
			t.Fatalf("chunk error: %v", chunk.Error)
		}
		content += chunk.Content
	}

	if content != "Streamed response" {
		t.Errorf("content = %q", content)
	}
}
