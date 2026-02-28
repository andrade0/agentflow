package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	// DefaultMaxSessions is the default number of sessions to keep
	DefaultMaxSessions = 50
)

// Manager handles session persistence
type Manager struct {
	dir         string
	maxSessions int
}

// NewManager creates a session manager
func NewManager(dir string) *Manager {
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".agentflow", "sessions")
	}
	return &Manager{
		dir:         dir,
		maxSessions: DefaultMaxSessions,
	}
}

// SetMaxSessions sets the maximum number of sessions to keep
func (m *Manager) SetMaxSessions(max int) {
	m.maxSessions = max
}

// ensureDir creates the sessions directory if needed
func (m *Manager) ensureDir() error {
	return os.MkdirAll(m.dir, 0755)
}

// sessionPath returns the file path for a session
func (m *Manager) sessionPath(id string) string {
	return filepath.Join(m.dir, id+".json")
}

// Save persists a session to disk
func (m *Manager) Save(s *Session) error {
	if err := m.ensureDir(); err != nil {
		return fmt.Errorf("create sessions dir: %w", err)
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}

	path := m.sessionPath(s.ID)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write session: %w", err)
	}

	// Cleanup old sessions
	m.cleanup()

	return nil
}

// Get retrieves a session by ID
func (m *Manager) Get(id string) (*Session, error) {
	path := m.sessionPath(id)
	return m.loadFromPath(path)
}

// GetByNameOrID finds a session by name or ID prefix
func (m *Manager) GetByNameOrID(query string) (*Session, error) {
	sessions, err := m.List()
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	for _, s := range sessions {
		// Exact ID match
		if s.ID == query {
			return s, nil
		}
		// ID prefix match
		if strings.HasPrefix(s.ID, query) {
			return s, nil
		}
		// Name match (case-insensitive)
		if s.Name != "" && strings.ToLower(s.Name) == query {
			return s, nil
		}
	}

	return nil, fmt.Errorf("session not found: %s", query)
}

// Delete removes a session
func (m *Manager) Delete(id string) error {
	path := m.sessionPath(id)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// Rename renames a session
func (m *Manager) Rename(id, name string) error {
	s, err := m.Get(id)
	if err != nil {
		return err
	}
	s.Name = name
	return m.Save(s)
}

// List returns all sessions sorted by last update (newest first)
func (m *Manager) List() ([]*Session, error) {
	if err := m.ensureDir(); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(m.dir)
	if err != nil {
		return nil, fmt.Errorf("read sessions dir: %w", err)
	}

	sessions := make([]*Session, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		path := filepath.Join(m.dir, entry.Name())
		s, err := m.loadFromPath(path)
		if err != nil {
			continue // Skip invalid sessions
		}
		sessions = append(sessions, s)
	}

	// Sort by UpdatedAt descending
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].UpdatedAt.After(sessions[j].UpdatedAt)
	})

	return sessions, nil
}

// ListForWorkdir returns sessions for a specific workdir
func (m *Manager) ListForWorkdir(workdir string) ([]*Session, error) {
	all, err := m.List()
	if err != nil {
		return nil, err
	}

	var filtered []*Session
	for _, s := range all {
		if s.Workdir == workdir {
			filtered = append(filtered, s)
		}
	}
	return filtered, nil
}

// GetLatest returns the most recent session for a workdir
func (m *Manager) GetLatest(workdir string) (*Session, error) {
	sessions, err := m.ListForWorkdir(workdir)
	if err != nil {
		return nil, err
	}
	if len(sessions) == 0 {
		return nil, fmt.Errorf("no sessions found for %s", workdir)
	}
	return sessions[0], nil
}

// loadFromPath loads a session from a file path
func (m *Manager) loadFromPath(path string) (*Session, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read session: %w", err)
	}

	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}

	return &s, nil
}

// cleanup removes old sessions beyond maxSessions
func (m *Manager) cleanup() {
	sessions, err := m.List()
	if err != nil || len(sessions) <= m.maxSessions {
		return
	}

	// Delete oldest sessions
	for _, s := range sessions[m.maxSessions:] {
		os.Remove(m.sessionPath(s.ID))
	}
}

// Count returns the total number of sessions
func (m *Manager) Count() (int, error) {
	sessions, err := m.List()
	if err != nil {
		return 0, err
	}
	return len(sessions), nil
}

// Dir returns the sessions directory path
func (m *Manager) Dir() string {
	return m.dir
}
