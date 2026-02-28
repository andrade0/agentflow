// Package provider defines the LLM provider interface and implementations
package provider

import (
	"context"

	"github.com/agentflow/agentflow/pkg/types"
)

// Provider is the interface all LLM providers must implement
type Provider interface {
	// Name returns the provider name (e.g., "ollama", "groq")
	Name() string

	// Complete sends a completion request and returns the response
	Complete(ctx context.Context, req types.CompletionRequest) (*types.CompletionResponse, error)

	// Stream sends a completion request and streams the response
	Stream(ctx context.Context, req types.CompletionRequest) (<-chan types.StreamChunk, error)

	// Models returns the list of available models
	Models() []string

	// SupportsModel checks if a model is supported
	SupportsModel(model string) bool
}

// Config holds provider configuration
type Config struct {
	BaseURL string   `yaml:"base_url"`
	APIKey  string   `yaml:"api_key"`
	Models  []string `yaml:"models"`
}

// Registry holds all registered providers
type Registry struct {
	providers map[string]Provider
}

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

// Register adds a provider to the registry
func (r *Registry) Register(p Provider) {
	r.providers[p.Name()] = p
}

// Get retrieves a provider by name
func (r *Registry) Get(name string) (Provider, bool) {
	p, ok := r.providers[name]
	return p, ok
}

// List returns all registered provider names
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// ResolveModel parses "provider/model" format and returns the provider and model
func (r *Registry) ResolveModel(spec string) (Provider, string, bool) {
	for i := 0; i < len(spec); i++ {
		if spec[i] == '/' {
			providerName := spec[:i]
			modelName := spec[i+1:]
			if p, ok := r.providers[providerName]; ok {
				return p, modelName, true
			}
			return nil, "", false
		}
	}
	return nil, "", false
}
