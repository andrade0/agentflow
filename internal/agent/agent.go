// Package agent handles the main agent runner and context management
package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/agentflow/agentflow/internal/provider"
	"github.com/agentflow/agentflow/internal/skill"
	"github.com/agentflow/agentflow/pkg/types"
)

// Agent represents an AI agent with context and capabilities
type Agent struct {
	id          string
	provider    provider.Provider
	model       string
	skills      *skill.Loader
	messages    []types.Message
	systemPrompt string
	metadata    map[string]string
	createdAt   time.Time
}

// Config holds agent configuration
type Config struct {
	ID           string
	Provider     provider.Provider
	Model        string
	Skills       *skill.Loader
	SystemPrompt string
	Metadata     map[string]string
}

// New creates a new agent
func New(cfg Config) *Agent {
	if cfg.ID == "" {
		cfg.ID = fmt.Sprintf("agent-%d", time.Now().UnixNano())
	}
	if cfg.Metadata == nil {
		cfg.Metadata = make(map[string]string)
	}

	a := &Agent{
		id:           cfg.ID,
		provider:     cfg.Provider,
		model:        cfg.Model,
		skills:       cfg.Skills,
		systemPrompt: cfg.SystemPrompt,
		metadata:     cfg.Metadata,
		createdAt:    time.Now(),
	}

	// Add system prompt if provided
	if cfg.SystemPrompt != "" {
		a.messages = append(a.messages, types.Message{
			Role:    "system",
			Content: cfg.SystemPrompt,
		})
	}

	return a
}

// ID returns the agent's unique identifier
func (a *Agent) ID() string {
	return a.id
}

// Model returns the model being used
func (a *Agent) Model() string {
	return a.model
}

// AddMessage adds a message to the conversation history
func (a *Agent) AddMessage(role, content string) {
	a.messages = append(a.messages, types.Message{
		Role:    role,
		Content: content,
	})
}

// Messages returns the conversation history
func (a *Agent) Messages() []types.Message {
	return a.messages
}

// ClearHistory clears the conversation history (keeps system prompt)
func (a *Agent) ClearHistory() {
	if a.systemPrompt != "" {
		a.messages = []types.Message{{
			Role:    "system",
			Content: a.systemPrompt,
		}}
	} else {
		a.messages = nil
	}
}

// SetMetadata sets a metadata value
func (a *Agent) SetMetadata(key, value string) {
	a.metadata[key] = value
}

// GetMetadata gets a metadata value
func (a *Agent) GetMetadata(key string) string {
	return a.metadata[key]
}

// Run sends a message and gets a response
func (a *Agent) Run(ctx context.Context, message string) (*types.CompletionResponse, error) {
	// Add user message
	a.AddMessage("user", message)

	// Build request
	req := types.CompletionRequest{
		Model:    a.model,
		Messages: a.messages,
	}

	// Get completion
	resp, err := a.provider.Complete(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("completion: %w", err)
	}

	// Add assistant response to history
	a.AddMessage("assistant", resp.Content)

	return resp, nil
}

// RunWithSkill runs a message with a specific skill context
func (a *Agent) RunWithSkill(ctx context.Context, skillName, message string) (*types.CompletionResponse, error) {
	if a.skills == nil {
		return a.Run(ctx, message)
	}

	sk, ok := a.skills.Get(skillName)
	if !ok {
		return nil, fmt.Errorf("skill not found: %s", skillName)
	}

	// Prepend skill content to message
	enhancedMessage := fmt.Sprintf("# Skill: %s\n\n%s\n\n---\n\n%s", sk.Name, sk.Content, message)
	return a.Run(ctx, enhancedMessage)
}

// Stream sends a message and streams the response
func (a *Agent) Stream(ctx context.Context, message string) (<-chan types.StreamChunk, error) {
	// Add user message
	a.AddMessage("user", message)

	// Build request
	req := types.CompletionRequest{
		Model:    a.model,
		Messages: a.messages,
		Stream:   true,
	}

	// Get stream
	chunks, err := a.provider.Stream(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("stream: %w", err)
	}

	// Wrap to collect full response
	output := make(chan types.StreamChunk)
	go func() {
		defer close(output)
		var fullContent strings.Builder
		for chunk := range chunks {
			if chunk.Error != nil {
				output <- chunk
				return
			}
			fullContent.WriteString(chunk.Content)
			output <- chunk
			if chunk.Done {
				// Add complete response to history
				a.AddMessage("assistant", fullContent.String())
			}
		}
	}()

	return output, nil
}

// Clone creates a new agent with the same configuration but fresh history
func (a *Agent) Clone(newID string) *Agent {
	if newID == "" {
		newID = fmt.Sprintf("%s-clone-%d", a.id, time.Now().UnixNano())
	}

	clone := &Agent{
		id:           newID,
		provider:     a.provider,
		model:        a.model,
		skills:       a.skills,
		systemPrompt: a.systemPrompt,
		metadata:     make(map[string]string),
		createdAt:    time.Now(),
	}

	// Copy metadata
	for k, v := range a.metadata {
		clone.metadata[k] = v
	}

	// Initialize with system prompt
	if a.systemPrompt != "" {
		clone.messages = []types.Message{{
			Role:    "system",
			Content: a.systemPrompt,
		}}
	}

	return clone
}
