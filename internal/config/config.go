// Package config handles AgentFlow configuration
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/agentflow/agentflow/internal/provider"
	"gopkg.in/yaml.v3"
)

// Config is the main configuration structure
type Config struct {
	Providers map[string]ProviderConfig `yaml:"providers"`
	Defaults  DefaultsConfig            `yaml:"defaults"`
	Skills    SkillsConfig              `yaml:"skills"`
}

// ProviderConfig holds provider-specific configuration
type ProviderConfig struct {
	BaseURL string   `yaml:"base_url"`
	APIKey  string   `yaml:"api_key"`
	Models  []string `yaml:"models"`
}

// DefaultsConfig holds default model assignments
type DefaultsConfig struct {
	Main     string `yaml:"main"`
	Subagent string `yaml:"subagent"`
	Reviewer string `yaml:"reviewer"`
}

// SkillsConfig holds skill-related configuration
type SkillsConfig struct {
	Paths []string `yaml:"paths"`
}

// Load reads configuration from the given path
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	// Expand environment variables
	expanded := os.ExpandEnv(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &cfg, nil
}

// ConfigSource tracks where configuration was loaded from
var ConfigSource string = ""

// LoadDefault loads configuration from default locations
func LoadDefault() (*Config, error) {
	// Check locations in order
	locations := []string{
		".agentflow/config.yaml",
		".agentflow/config.yml",
	}

	// Add home directory locations
	if home, err := os.UserHomeDir(); err == nil {
		locations = append(locations,
			filepath.Join(home, ".agentflow", "config.yaml"),
			filepath.Join(home, ".agentflow", "config.yml"),
			filepath.Join(home, ".config", "agentflow", "config.yaml"),
		)
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			ConfigSource = loc
			return Load(loc)
		}
	}

	// Return default config if no file found
	ConfigSource = "(default - no config file found)"
	return DefaultConfig(), nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Providers: map[string]ProviderConfig{
			"ollama": {
				BaseURL: "http://localhost:11434",
				Models:  []string{"llama3.3:latest", "codellama:latest"},
			},
		},
		Defaults: DefaultsConfig{
			Main:     "ollama/llama3.3:latest",
			Subagent: "ollama/llama3.3:latest",
			Reviewer: "ollama/llama3.3:latest",
		},
		Skills: SkillsConfig{
			Paths: []string{"skills", ".agentflow/skills"},
		},
	}
}

// Save writes configuration to the given path
func (c *Config) Save(path string) error {
	// Create parent directory if needed
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

// BuildRegistry creates a provider registry from configuration
func (c *Config) BuildRegistry() *provider.Registry {
	registry := provider.NewRegistry()

	for name, cfg := range c.Providers {
		provCfg := provider.Config{
			BaseURL: cfg.BaseURL,
			APIKey:  cfg.APIKey,
			Models:  cfg.Models,
		}

		var p provider.Provider
		switch strings.ToLower(name) {
		case "ollama":
			p = provider.NewOllama(provCfg)
		case "groq":
			p = provider.NewGroq(provCfg)
		case "together":
			p = provider.NewTogether(provCfg)
		default:
			// Generic OpenAI-compatible
			p = provider.NewOpenAICompat(name, provCfg)
		}
		registry.Register(p)
	}

	return registry
}
