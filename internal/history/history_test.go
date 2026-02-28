package history

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHistory(t *testing.T) {
	// Create a temp directory for testing
	tmpDir, err := os.MkdirTemp("", "history-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Override home directory for testing
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	t.Run("NewHistory", func(t *testing.T) {
		h, err := New("/test/workdir")
		if err != nil {
			t.Fatalf("Failed to create history: %v", err)
		}
		if h.Len() != 0 {
			t.Errorf("Expected empty history, got %d entries", h.Len())
		}
	})

	t.Run("AddAndRetrieve", func(t *testing.T) {
		h, _ := New("/test/workdir2")

		h.Add("first command")
		h.Add("second command")
		h.Add("third command")

		if h.Len() != 3 {
			t.Errorf("Expected 3 entries, got %d", h.Len())
		}

		// Test Previous
		cmd, ok := h.Previous()
		if !ok || cmd != "third command" {
			t.Errorf("Expected 'third command', got '%s'", cmd)
		}

		cmd, ok = h.Previous()
		if !ok || cmd != "second command" {
			t.Errorf("Expected 'second command', got '%s'", cmd)
		}

		// Test Next
		cmd, ok = h.Next()
		if !ok || cmd != "third command" {
			t.Errorf("Expected 'third command', got '%s'", cmd)
		}
	})

	t.Run("Search", func(t *testing.T) {
		h, _ := New("/test/workdir3")

		h.Add("git status")
		h.Add("git commit -m 'test'")
		h.Add("ls -la")
		h.Add("git push origin main")

		results := h.Search("git")
		if len(results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(results))
		}

		// Most recent should be first
		if results[0].Entry != "git push origin main" {
			t.Errorf("Expected 'git push origin main', got '%s'", results[0].Entry)
		}
	})

	t.Run("NoDuplicates", func(t *testing.T) {
		h, _ := New("/test/workdir4")

		h.Add("same command")
		h.Add("same command")
		h.Add("same command")

		if h.Len() != 1 {
			t.Errorf("Expected 1 entry (no duplicates), got %d", h.Len())
		}
	})

	t.Run("Persistence", func(t *testing.T) {
		workdir := "/test/persistence"

		h1, _ := New(workdir)
		h1.Add("persistent command")

		// Create new history instance for same workdir
		h2, _ := New(workdir)

		if h2.Len() != 1 {
			t.Errorf("Expected 1 entry from disk, got %d", h2.Len())
		}

		cmd, ok := h2.Previous()
		if !ok || cmd != "persistent command" {
			t.Errorf("Expected 'persistent command', got '%s'", cmd)
		}
	})

	t.Run("MaxEntries", func(t *testing.T) {
		h, _ := New("/test/maxentries")

		// Add more than max entries
		for i := 0; i < MaxEntriesPerWorkdir+100; i++ {
			h.Add("command " + string(rune('0'+i%10)))
		}

		if h.Len() > MaxEntriesPerWorkdir {
			t.Errorf("Expected max %d entries, got %d", MaxEntriesPerWorkdir, h.Len())
		}
	})
}

func TestHistoryDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "history-dir-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	_, err = New("/some/workdir")
	if err != nil {
		t.Fatalf("Failed to create history: %v", err)
	}

	historyDir := filepath.Join(tmpDir, HistoryDir)
	if _, err := os.Stat(historyDir); os.IsNotExist(err) {
		t.Error("History directory was not created")
	}
}
