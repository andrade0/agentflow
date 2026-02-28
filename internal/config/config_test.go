package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("expected non-nil config")
	}

	if _, ok := cfg.Providers["ollama"]; !ok {
		t.Error("expected ollama provider")
	}

	if cfg.Defaults.Main == "" {
		t.Error("expected default main model")
	}
}

func TestLoad(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configContent := `
providers:
  ollama:
    base_url: http://localhost:11434
    models:
      - llama3.3:latest
  groq:
    api_key: test-key
    models:
      - llama-3.3-70b-versatile

defaults:
  main: groq/llama-3.3-70b-versatile
  subagent: ollama/llama3.3:latest
  reviewer: groq/llama-3.3-70b-versatile

skills:
  paths:
    - skills
    - ~/.agentflow/skills
`

	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// Check providers
	if len(cfg.Providers) != 2 {
		t.Errorf("expected 2 providers, got %d", len(cfg.Providers))
	}

	ollama := cfg.Providers["ollama"]
	if ollama.BaseURL != "http://localhost:11434" {
		t.Errorf("ollama base_url = %q", ollama.BaseURL)
	}
	if len(ollama.Models) != 1 {
		t.Errorf("ollama models = %d", len(ollama.Models))
	}

	groq := cfg.Providers["groq"]
	if groq.APIKey != "test-key" {
		t.Errorf("groq api_key = %q", groq.APIKey)
	}

	// Check defaults
	if cfg.Defaults.Main != "groq/llama-3.3-70b-versatile" {
		t.Errorf("defaults.main = %q", cfg.Defaults.Main)
	}

	// Check skills
	if len(cfg.Skills.Paths) != 2 {
		t.Errorf("skills.paths = %d", len(cfg.Skills.Paths))
	}
}

func TestLoad_EnvExpansion(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-env-test")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set test env var
	os.Setenv("TEST_API_KEY", "secret-key-123")
	defer os.Unsetenv("TEST_API_KEY")

	configContent := `
providers:
  groq:
    api_key: ${TEST_API_KEY}
    models:
      - test-model
defaults:
  main: groq/test-model
`

	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Providers["groq"].APIKey != "secret-key-123" {
		t.Errorf("api_key not expanded: %q", cfg.Providers["groq"].APIKey)
	}
}

func TestConfig_Save(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-save-test")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := DefaultConfig()
	savePath := filepath.Join(tmpDir, "subdir", "config.yaml")

	if err := cfg.Save(savePath); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Check file exists
	if _, err := os.Stat(savePath); err != nil {
		t.Errorf("saved file not found: %v", err)
	}

	// Load and verify
	loaded, err := Load(savePath)
	if err != nil {
		t.Fatalf("Load after save: %v", err)
	}

	if loaded.Defaults.Main != cfg.Defaults.Main {
		t.Errorf("loaded main = %q, want %q", loaded.Defaults.Main, cfg.Defaults.Main)
	}
}

func TestConfig_BuildRegistry(t *testing.T) {
	cfg := &Config{
		Providers: map[string]ProviderConfig{
			"ollama": {
				BaseURL: "http://localhost:11434",
				Models:  []string{"llama3.3"},
			},
			"groq": {
				APIKey: "test-key",
				Models: []string{"mixtral-8x7b"},
			},
			"together": {
				APIKey: "test-key",
				Models: []string{"llama-70b"},
			},
			"custom": {
				BaseURL: "https://custom.api.com/v1",
				APIKey:  "key",
				Models:  []string{"custom-model"},
			},
		},
	}

	registry := cfg.BuildRegistry()

	// Check all providers registered
	providers := registry.List()
	if len(providers) != 4 {
		t.Errorf("expected 4 providers, got %d", len(providers))
	}

	// Check specific providers
	ollama, ok := registry.Get("ollama")
	if !ok {
		t.Error("ollama not registered")
	} else if ollama.Name() != "ollama" {
		t.Errorf("ollama.Name() = %q", ollama.Name())
	}

	groq, ok := registry.Get("groq")
	if !ok {
		t.Error("groq not registered")
	} else if groq.Name() != "groq" {
		t.Errorf("groq.Name() = %q", groq.Name())
	}

	custom, ok := registry.Get("custom")
	if !ok {
		t.Error("custom not registered")
	} else if custom.Name() != "custom" {
		t.Errorf("custom.Name() = %q", custom.Name())
	}
}

func TestLoadDefault_NoConfig(t *testing.T) {
	// Save current directory
	cwd, _ := os.Getwd()
	
	// Change to temp directory with no config
	tmpDir, _ := os.MkdirTemp("", "no-config")
	defer os.RemoveAll(tmpDir)
	os.Chdir(tmpDir)
	defer os.Chdir(cwd)

	cfg, err := LoadDefault()
	if err != nil {
		t.Fatalf("LoadDefault: %v", err)
	}

	// Should return default config
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.Defaults.Main == "" {
		t.Error("expected default main model")
	}
}
