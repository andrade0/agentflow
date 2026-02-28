package history

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	// MaxEntriesPerWorkdir is the maximum number of entries per working directory
	MaxEntriesPerWorkdir = 1000

	// HistoryDir is the default history directory name
	HistoryDir = ".agentflow/history"
)

// History manages command history persistence
type History struct {
	mu       sync.RWMutex
	entries  []string
	workdir  string
	filePath string
	position int
}

// New creates a new History manager for the given working directory
func New(workdir string) (*History, error) {
	if workdir == "" {
		var err error
		workdir, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	// Get user home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Create history directory if it doesn't exist
	historyDir := filepath.Join(homeDir, HistoryDir)
	if err := os.MkdirAll(historyDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create history directory: %w", err)
	}

	// Create a hash of the workdir for the filename
	hash := sha256.Sum256([]byte(workdir))
	filename := hex.EncodeToString(hash[:8]) + ".txt"
	filePath := filepath.Join(historyDir, filename)

	h := &History{
		entries:  make([]string, 0),
		workdir:  workdir,
		filePath: filePath,
		position: 0,
	}

	// Load existing history
	if err := h.load(); err != nil {
		// Not fatal - just start with empty history
		h.entries = make([]string, 0)
	}

	h.position = len(h.entries)

	return h, nil
}

// load reads history from disk
func (h *History) load() error {
	file, err := os.Open(h.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No history file yet
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			h.entries = append(h.entries, line)
		}
	}

	return scanner.Err()
}

// save writes history to disk
func (h *History) save() error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	file, err := os.Create(h.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, entry := range h.entries {
		// Replace newlines with a special marker for multiline commands
		escaped := strings.ReplaceAll(entry, "\n", "\\n")
		if _, err := writer.WriteString(escaped + "\n"); err != nil {
			return err
		}
	}

	return writer.Flush()
}

// Add adds a new entry to history
func (h *History) Add(entry string) error {
	entry = strings.TrimSpace(entry)
	if entry == "" {
		return nil
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// Don't add duplicates of the last entry
	if len(h.entries) > 0 && h.entries[len(h.entries)-1] == entry {
		h.position = len(h.entries)
		return nil
	}

	h.entries = append(h.entries, entry)

	// Trim to max entries
	if len(h.entries) > MaxEntriesPerWorkdir {
		h.entries = h.entries[len(h.entries)-MaxEntriesPerWorkdir:]
	}

	h.position = len(h.entries)

	return h.save()
}

// Previous returns the previous entry in history
func (h *History) Previous() (string, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.entries) == 0 || h.position <= 0 {
		return "", false
	}

	h.position--
	return h.entries[h.position], true
}

// Next returns the next entry in history
func (h *History) Next() (string, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.position >= len(h.entries)-1 {
		h.position = len(h.entries)
		return "", false
	}

	h.position++
	return h.entries[h.position], true
}

// Reset resets the position to the end
func (h *History) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.position = len(h.entries)
}

// Search searches for entries matching the query (reverse search)
func (h *History) Search(query string) []SearchResult {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if query == "" {
		return nil
	}

	results := make([]SearchResult, 0)
	queryLower := strings.ToLower(query)

	// Search from most recent to oldest
	for i := len(h.entries) - 1; i >= 0; i-- {
		entry := h.entries[i]
		entryLower := strings.ToLower(entry)
		if idx := strings.Index(entryLower, queryLower); idx >= 0 {
			results = append(results, SearchResult{
				Entry:      entry,
				Index:      i,
				MatchStart: idx,
				MatchEnd:   idx + len(query),
			})
		}
	}

	return results
}

// Get returns the entry at the given index
func (h *History) Get(index int) (string, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if index < 0 || index >= len(h.entries) {
		return "", false
	}

	return h.entries[index], true
}

// Len returns the number of entries
func (h *History) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.entries)
}

// All returns all entries
func (h *History) All() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]string, len(h.entries))
	copy(result, h.entries)
	return result
}

// SearchResult represents a search match
type SearchResult struct {
	Entry      string
	Index      int
	MatchStart int
	MatchEnd   int
}
