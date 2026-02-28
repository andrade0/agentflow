// Package session handles persistent conversation sessions
package session

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/agentflow/agentflow/pkg/types"
)

// Session represents a persistent conversation session
type Session struct {
	ID        string          `json:"id"`
	Name      string          `json:"name,omitempty"`
	Workdir   string          `json:"workdir"`
	Provider  string          `json:"provider"`
	Model     string          `json:"model"`
	Messages  []types.Message `json:"messages"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	Metadata  map[string]any  `json:"metadata,omitempty"`
}

// New creates a new session
func New(workdir, provider, model string) *Session {
	return &Session{
		ID:        generateID(),
		Workdir:   workdir,
		Provider:  provider,
		Model:     model,
		Messages:  make([]types.Message, 0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  make(map[string]any),
	}
}

// generateID creates a short random session ID
func generateID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// AddMessage adds a message to the session
func (s *Session) AddMessage(role, content string) {
	s.Messages = append(s.Messages, types.Message{
		Role:    role,
		Content: content,
	})
	s.UpdatedAt = time.Now()
}

// DisplayName returns the name or a generated display name
func (s *Session) DisplayName() string {
	if s.Name != "" {
		return s.Name
	}
	// Use first user message as preview
	for _, msg := range s.Messages {
		if msg.Role == "user" {
			preview := msg.Content
			if len(preview) > 40 {
				preview = preview[:40] + "..."
			}
			return preview
		}
	}
	return s.ID
}

// Clone creates a fork of this session with a new ID
func (s *Session) Clone() *Session {
	clone := &Session{
		ID:        generateID(),
		Name:      "",
		Workdir:   s.Workdir,
		Provider:  s.Provider,
		Model:     s.Model,
		Messages:  make([]types.Message, len(s.Messages)),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  make(map[string]any),
	}
	copy(clone.Messages, s.Messages)
	for k, v := range s.Metadata {
		clone.Metadata[k] = v
	}
	return clone
}

// MessageCount returns the number of messages
func (s *Session) MessageCount() int {
	return len(s.Messages)
}

// LastActivity returns the last update time
func (s *Session) LastActivity() time.Time {
	return s.UpdatedAt
}
