package session

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSessionManager(t *testing.T) {
	// Create temp dir
	tmpDir, err := os.MkdirTemp("", "agentflow-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	sessDir := filepath.Join(tmpDir, "sessions")
	mgr := NewManager(sessDir)

	// Test Save and Get
	t.Run("SaveAndGet", func(t *testing.T) {
		s := New("/test/workdir", "ollama", "llama3")
		s.AddMessage("user", "Hello")
		s.AddMessage("assistant", "Hi there!")

		if err := mgr.Save(s); err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		loaded, err := mgr.Get(s.ID)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		if loaded.ID != s.ID {
			t.Errorf("ID mismatch: got %s, want %s", loaded.ID, s.ID)
		}
		if len(loaded.Messages) != 2 {
			t.Errorf("Messages count: got %d, want 2", len(loaded.Messages))
		}
	})

	// Test List
	t.Run("List", func(t *testing.T) {
		// Create additional sessions
		for i := 0; i < 3; i++ {
			s := New("/test/workdir", "ollama", "llama3")
			s.AddMessage("user", "Test")
			mgr.Save(s)
		}

		sessions, err := mgr.List()
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if len(sessions) < 4 {
			t.Errorf("Expected at least 4 sessions, got %d", len(sessions))
		}
	})

	// Test GetLatest
	t.Run("GetLatest", func(t *testing.T) {
		s := New("/specific/workdir", "groq", "mixtral")
		s.AddMessage("user", "Latest message")
		mgr.Save(s)

		latest, err := mgr.GetLatest("/specific/workdir")
		if err != nil {
			t.Fatalf("GetLatest failed: %v", err)
		}

		if latest.ID != s.ID {
			t.Errorf("GetLatest returned wrong session")
		}
	})

	// Test Rename
	t.Run("Rename", func(t *testing.T) {
		s := New("/test", "ollama", "llama3")
		mgr.Save(s)

		if err := mgr.Rename(s.ID, "my-session"); err != nil {
			t.Fatalf("Rename failed: %v", err)
		}

		loaded, _ := mgr.Get(s.ID)
		if loaded.Name != "my-session" {
			t.Errorf("Name not updated: got %s", loaded.Name)
		}
	})

	// Test GetByNameOrID
	t.Run("GetByNameOrID", func(t *testing.T) {
		s := New("/test", "ollama", "llama3")
		s.Name = "named-session"
		mgr.Save(s)

		// By name
		found, err := mgr.GetByNameOrID("named-session")
		if err != nil {
			t.Fatalf("GetByNameOrID by name failed: %v", err)
		}
		if found.ID != s.ID {
			t.Error("Wrong session found by name")
		}

		// By ID prefix
		found, err = mgr.GetByNameOrID(s.ID[:4])
		if err != nil {
			t.Fatalf("GetByNameOrID by prefix failed: %v", err)
		}
		if found.ID != s.ID {
			t.Error("Wrong session found by prefix")
		}
	})

	// Test Delete
	t.Run("Delete", func(t *testing.T) {
		s := New("/test", "ollama", "llama3")
		mgr.Save(s)

		if err := mgr.Delete(s.ID); err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		_, err := mgr.Get(s.ID)
		if err == nil {
			t.Error("Session should not exist after delete")
		}
	})

	// Test cleanup
	t.Run("Cleanup", func(t *testing.T) {
		mgr.SetMaxSessions(5)

		// Create more than max sessions
		for i := 0; i < 10; i++ {
			s := New("/cleanup-test", "ollama", "llama3")
			mgr.Save(s)
		}

		sessions, _ := mgr.List()
		if len(sessions) > 5 {
			t.Errorf("Expected max 5 sessions after cleanup, got %d", len(sessions))
		}
	})
}

func TestSession(t *testing.T) {
	t.Run("DisplayName", func(t *testing.T) {
		s := New("/test", "ollama", "llama3")

		// Empty session
		if s.DisplayName() != s.ID {
			t.Error("Empty session should display ID")
		}

		// With message
		s.AddMessage("user", "Hello world")
		if s.DisplayName() != "Hello world" {
			t.Errorf("Expected 'Hello world', got %s", s.DisplayName())
		}

		// With name
		s.Name = "My Session"
		if s.DisplayName() != "My Session" {
			t.Errorf("Expected 'My Session', got %s", s.DisplayName())
		}
	})

	t.Run("Clone", func(t *testing.T) {
		s := New("/test", "ollama", "llama3")
		s.AddMessage("user", "test")
		s.Metadata["key"] = "value"

		clone := s.Clone()

		if clone.ID == s.ID {
			t.Error("Clone should have different ID")
		}
		if len(clone.Messages) != 1 {
			t.Error("Clone should copy messages")
		}
		if clone.Metadata["key"] != "value" {
			t.Error("Clone should copy metadata")
		}
	})
}
