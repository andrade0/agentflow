package input

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// CompletionType represents the type of completion
type CompletionType int

const (
	CompletionNone CompletionType = iota
	CompletionCommand
	CompletionFile
)

// Completion represents a single completion suggestion
type Completion struct {
	Value       string
	Display     string
	Description string
	Type        CompletionType
}

// Completer provides autocomplete suggestions
type Completer struct {
	commands []Completion
}

// NewCompleter creates a new Completer with default commands
func NewCompleter() *Completer {
	return &Completer{
		commands: []Completion{
			{Value: "/help", Display: "/help", Description: "Show help message", Type: CompletionCommand},
			{Value: "/quit", Display: "/quit", Description: "Exit the session", Type: CompletionCommand},
			{Value: "/exit", Display: "/exit", Description: "Exit the session", Type: CompletionCommand},
			{Value: "/clear", Display: "/clear", Description: "Clear conversation", Type: CompletionCommand},
			{Value: "/model", Display: "/model", Description: "Show/change model", Type: CompletionCommand},
			{Value: "/provider", Display: "/provider", Description: "Show/change provider", Type: CompletionCommand},
			{Value: "/skills", Display: "/skills", Description: "List available skills", Type: CompletionCommand},
			{Value: "/status", Display: "/status", Description: "Show session status", Type: CompletionCommand},
			{Value: "/history", Display: "/history", Description: "Show conversation stats", Type: CompletionCommand},
			{Value: "/compact", Display: "/compact", Description: "Compact conversation", Type: CompletionCommand},
		},
	}
}

// Complete returns completion suggestions based on the input
func (c *Completer) Complete(input string, cursorPos int) []Completion {
	if input == "" {
		return nil
	}

	// Extract the word at cursor position
	word, wordStart := c.wordAtCursor(input, cursorPos)

	// Check for command completion (starts with /)
	if strings.HasPrefix(input, "/") && wordStart == 0 {
		return c.completeCommands(word)
	}

	// Check for file completion (starts with @)
	if strings.HasPrefix(word, "@") {
		return c.completeFiles(strings.TrimPrefix(word, "@"))
	}

	// Check for @ anywhere in input
	atIdx := strings.LastIndex(input[:cursorPos], "@")
	if atIdx >= 0 {
		// Check if @ is at word boundary
		if atIdx == 0 || input[atIdx-1] == ' ' {
			partial := input[atIdx+1 : cursorPos]
			return c.completeFiles(partial)
		}
	}

	return nil
}

// wordAtCursor extracts the word at the cursor position
func (c *Completer) wordAtCursor(input string, cursorPos int) (string, int) {
	if cursorPos > len(input) {
		cursorPos = len(input)
	}

	// Find word start
	start := cursorPos
	for start > 0 && !isWordSeparator(input[start-1]) {
		start--
	}

	return input[start:cursorPos], start
}

func isWordSeparator(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n'
}

// completeCommands returns command completions
func (c *Completer) completeCommands(prefix string) []Completion {
	var results []Completion
	prefixLower := strings.ToLower(prefix)

	for _, cmd := range c.commands {
		if strings.HasPrefix(strings.ToLower(cmd.Value), prefixLower) {
			results = append(results, cmd)
		}
	}

	return results
}

// completeFiles returns file completions
func (c *Completer) completeFiles(prefix string) []Completion {
	var results []Completion

	// Determine base path
	dir := "."
	base := prefix

	if strings.Contains(prefix, "/") {
		dir = filepath.Dir(prefix)
		base = filepath.Base(prefix)
	}

	// Read directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	baseLower := strings.ToLower(base)

	for _, entry := range entries {
		name := entry.Name()
		nameLower := strings.ToLower(name)

		// Skip hidden files unless prefix starts with .
		if strings.HasPrefix(name, ".") && !strings.HasPrefix(base, ".") {
			continue
		}

		if strings.HasPrefix(nameLower, baseLower) {
			fullPath := filepath.Join(dir, name)
			display := "@" + fullPath

			desc := "file"
			if entry.IsDir() {
				desc = "directory"
				display += "/"
			}

			results = append(results, Completion{
				Value:       "@" + fullPath,
				Display:     display,
				Description: desc,
				Type:        CompletionFile,
			})
		}
	}

	// Sort: directories first, then alphabetically
	sort.Slice(results, func(i, j int) bool {
		if results[i].Description != results[j].Description {
			return results[i].Description == "directory"
		}
		return results[i].Value < results[j].Value
	})

	// Limit results
	if len(results) > 10 {
		results = results[:10]
	}

	return results
}

// AddCommand adds a custom command to the completer
func (c *Completer) AddCommand(value, description string) {
	c.commands = append(c.commands, Completion{
		Value:       value,
		Display:     value,
		Description: description,
		Type:        CompletionCommand,
	})
}
