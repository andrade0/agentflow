package provider

import (
	"testing"
)

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Fatal("expected non-nil registry")
	}
	if len(r.providers) != 0 {
		t.Errorf("expected empty registry, got %d providers", len(r.providers))
	}
}

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()
	p := NewOllama(Config{BaseURL: "http://localhost:11434"})
	
	r.Register(p)
	
	got, ok := r.Get("ollama")
	if !ok {
		t.Fatal("expected provider to be registered")
	}
	if got.Name() != "ollama" {
		t.Errorf("expected 'ollama', got %q", got.Name())
	}
}

func TestRegistry_List(t *testing.T) {
	r := NewRegistry()
	r.Register(NewOllama(Config{}))
	r.Register(NewGroq(Config{APIKey: "test"}))
	
	list := r.List()
	if len(list) != 2 {
		t.Errorf("expected 2 providers, got %d", len(list))
	}
}

func TestRegistry_ResolveModel(t *testing.T) {
	r := NewRegistry()
	r.Register(NewOllama(Config{}))
	r.Register(NewGroq(Config{APIKey: "test"}))
	
	tests := []struct {
		spec     string
		provider string
		model    string
		ok       bool
	}{
		{"ollama/llama3.3:70b", "ollama", "llama3.3:70b", true},
		{"groq/llama-3.3-70b-versatile", "groq", "llama-3.3-70b-versatile", true},
		{"unknown/model", "", "", false},
		{"nomodel", "", "", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.spec, func(t *testing.T) {
			p, model, ok := r.ResolveModel(tt.spec)
			if ok != tt.ok {
				t.Errorf("ResolveModel(%q): ok = %v, want %v", tt.spec, ok, tt.ok)
				return
			}
			if !ok {
				return
			}
			if p.Name() != tt.provider {
				t.Errorf("provider = %q, want %q", p.Name(), tt.provider)
			}
			if model != tt.model {
				t.Errorf("model = %q, want %q", model, tt.model)
			}
		})
	}
}

func TestOllamaProvider_Name(t *testing.T) {
	p := NewOllama(Config{})
	if p.Name() != "ollama" {
		t.Errorf("expected 'ollama', got %q", p.Name())
	}
}

func TestOllamaProvider_DefaultURL(t *testing.T) {
	p := NewOllama(Config{})
	if p.baseURL != "http://localhost:11434" {
		t.Errorf("expected default URL, got %q", p.baseURL)
	}
}

func TestOllamaProvider_Models(t *testing.T) {
	p := NewOllama(Config{Models: []string{"llama3.3", "codellama"}})
	models := p.Models()
	if len(models) != 2 {
		t.Errorf("expected 2 models, got %d", len(models))
	}
}

func TestGroqProvider(t *testing.T) {
	p := NewGroq(Config{APIKey: "test-key"})
	if p.Name() != "groq" {
		t.Errorf("expected 'groq', got %q", p.Name())
	}
	if p.baseURL != "https://api.groq.com/openai/v1" {
		t.Errorf("expected Groq URL, got %q", p.baseURL)
	}
}

func TestTogetherProvider(t *testing.T) {
	p := NewTogether(Config{APIKey: "test-key"})
	if p.Name() != "together" {
		t.Errorf("expected 'together', got %q", p.Name())
	}
	if p.baseURL != "https://api.together.xyz/v1" {
		t.Errorf("expected Together URL, got %q", p.baseURL)
	}
}

func TestOpenAICompatProvider_SupportsModel(t *testing.T) {
	p := NewOpenAICompat("test", Config{Models: []string{"model-a", "model-b"}})
	
	if !p.SupportsModel("model-a") {
		t.Error("expected model-a to be supported")
	}
	if p.SupportsModel("model-c") {
		t.Error("expected model-c to not be supported")
	}
}
