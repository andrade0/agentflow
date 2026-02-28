package types

import (
	"encoding/json"
	"testing"
)

func TestMessage_JSON(t *testing.T) {
	msg := Message{
		Role:    "user",
		Content: "Hello, world!",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded Message
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.Role != msg.Role {
		t.Errorf("Role = %q, want %q", decoded.Role, msg.Role)
	}
	if decoded.Content != msg.Content {
		t.Errorf("Content = %q, want %q", decoded.Content, msg.Content)
	}
}

func TestCompletionRequest_JSON(t *testing.T) {
	req := CompletionRequest{
		Model: "llama3.3",
		Messages: []Message{
			{Role: "system", Content: "You are helpful."},
			{Role: "user", Content: "Hi"},
		},
		Temperature: 0.7,
		MaxTokens:   1024,
		Stream:      false,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded CompletionRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.Model != req.Model {
		t.Errorf("Model = %q", decoded.Model)
	}
	if len(decoded.Messages) != 2 {
		t.Errorf("Messages len = %d", len(decoded.Messages))
	}
	if decoded.Temperature != 0.7 {
		t.Errorf("Temperature = %f", decoded.Temperature)
	}
}

func TestCompletionRequest_OmitEmpty(t *testing.T) {
	req := CompletionRequest{
		Model:    "test",
		Messages: []Message{{Role: "user", Content: "hi"}},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	// Should not contain temperature, max_tokens, stream when zero
	str := string(data)
	if contains(str, "temperature") {
		t.Error("should omit temperature when zero")
	}
	if contains(str, "max_tokens") {
		t.Error("should omit max_tokens when zero")
	}
	if contains(str, "stream") {
		t.Error("should omit stream when false")
	}
}

func TestCompletionResponse_JSON(t *testing.T) {
	resp := CompletionResponse{
		Content:      "Hello!",
		Model:        "llama3.3",
		FinishReason: "stop",
		TokensUsed:   50,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded CompletionResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.Content != resp.Content {
		t.Errorf("Content = %q", decoded.Content)
	}
	if decoded.TokensUsed != 50 {
		t.Errorf("TokensUsed = %d", decoded.TokensUsed)
	}
}

func TestProviderType_Constants(t *testing.T) {
	types := []ProviderType{
		ProviderOllama,
		ProviderGroq,
		ProviderTogether,
		ProviderOpenAI,
	}

	expected := []string{"ollama", "groq", "together", "openai"}

	for i, pt := range types {
		if string(pt) != expected[i] {
			t.Errorf("ProviderType[%d] = %q, want %q", i, pt, expected[i])
		}
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
