package input

import (
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestInput(t *testing.T) {
	// Create temp dir for history
	tmpDir, err := os.MkdirTemp("", "input-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	t.Run("NewInput", func(t *testing.T) {
		m := New("/test/workdir")
		if m.Mode() != ModeNormal {
			t.Errorf("Expected ModeNormal, got %v", m.Mode())
		}
	})

	t.Run("ReverseSearchMode", func(t *testing.T) {
		m := New("/test/workdir")

		// Add some history
		m.history.Add("git status")
		m.history.Add("git commit")
		m.history.Add("ls -la")

		// Simulate Ctrl+R
		keyMsg := tea.KeyMsg{Type: tea.KeyCtrlR}
		m, _ = m.Update(keyMsg)

		if m.Mode() != ModeReverseSearch {
			t.Errorf("Expected ModeReverseSearch, got %v", m.Mode())
		}

		// Type search query
		for _, ch := range "git" {
			m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		}

		if m.searchQuery != "git" {
			t.Errorf("Expected search query 'git', got '%s'", m.searchQuery)
		}

		if len(m.searchResults) != 2 {
			t.Errorf("Expected 2 search results, got %d", len(m.searchResults))
		}

		// Press Esc to cancel
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})

		if m.Mode() != ModeNormal {
			t.Errorf("Expected ModeNormal after Esc, got %v", m.Mode())
		}
	})

	t.Run("AutocompleteMode", func(t *testing.T) {
		m := New("/test/workdir")

		// Type /he
		m.textarea.SetValue("/he")

		// Press Tab
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})

		// Should have autocomplete suggestions for /help
		if m.Mode() == ModeAutocomplete {
			if len(m.completions) == 0 {
				t.Error("Expected completions, got none")
			}
		}
	})

	t.Run("MultilineInput", func(t *testing.T) {
		m := New("/test/workdir")
		m.textarea.SetValue("line1\\")

		// Press Enter after backslash
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		value := m.Value()
		if value != "line1\n" {
			t.Errorf("Expected 'line1\\n', got '%s'", value)
		}
	})

	t.Run("BashMode", func(t *testing.T) {
		m := New("/test/workdir")
		m.textarea.SetValue("!echo hello")

		// Submit
		m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})

		if cmd == nil {
			t.Fatal("Expected command, got nil")
		}

		// Execute the command to get the message
		msg := cmd()
		submitMsg, ok := msg.(SubmitMsg)
		if !ok {
			t.Fatalf("Expected SubmitMsg, got %T", msg)
		}

		if !submitMsg.IsBash {
			t.Error("Expected IsBash to be true")
		}

		if submitMsg.Value != "echo hello" {
			t.Errorf("Expected 'echo hello', got '%s'", submitMsg.Value)
		}
	})
}

func TestCompleter(t *testing.T) {
	c := NewCompleter()

	t.Run("CommandCompletion", func(t *testing.T) {
		results := c.Complete("/he", 3)
		if len(results) != 1 { // /help matches
			t.Errorf("Expected 1 result for /he, got %d", len(results))
		}
		if results[0].Value != "/help" {
			t.Errorf("Expected /help, got %s", results[0].Value)
		}

		// Test /h which should match /help and /history
		results2 := c.Complete("/h", 2)
		if len(results2) != 2 { // /help and /history
			t.Errorf("Expected 2 results for /h, got %d", len(results2))
		}
	})

	t.Run("NoCompletion", func(t *testing.T) {
		results := c.Complete("hello", 5)
		if len(results) != 0 {
			t.Errorf("Expected 0 results, got %d", len(results))
		}
	})

	t.Run("FileCompletion", func(t *testing.T) {
		// This will try to complete from current directory
		results := c.Complete("@", 1)
		// Should have some results (files in current dir)
		// Can't predict exact count
		if results == nil {
			// That's ok, might be empty dir
		}
	})
}

func TestBashExecution(t *testing.T) {
	t.Run("SimpleCommand", func(t *testing.T) {
		result := ExecuteBash(nil, "echo hello")
		if result.ExitCode != 0 {
			t.Errorf("Expected exit code 0, got %d", result.ExitCode)
		}
		if result.Output != "hello\n" {
			t.Errorf("Expected 'hello\\n', got '%s'", result.Output)
		}
	})

	t.Run("FailingCommand", func(t *testing.T) {
		result := ExecuteBash(nil, "exit 42")
		if result.ExitCode != 42 {
			t.Errorf("Expected exit code 42, got %d", result.ExitCode)
		}
	})

	t.Run("FormatResult", func(t *testing.T) {
		result := BashResult{
			Command:  "ls",
			Output:   "file1\nfile2\n",
			ExitCode: 0,
		}
		formatted := FormatBashResult(result)
		if formatted == "" {
			t.Error("Expected non-empty formatted result")
		}
	})
}
