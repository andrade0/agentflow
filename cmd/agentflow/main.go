// Package main is the entry point for the agentflow CLI
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/agentflow/agentflow/internal/agent"
	"github.com/agentflow/agentflow/internal/config"
	"github.com/agentflow/agentflow/internal/session"
	"github.com/agentflow/agentflow/internal/skill"
	"github.com/agentflow/agentflow/internal/subagent"
	"github.com/agentflow/agentflow/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	version      = "dev"
	cfgFile      string
	modelSpec    string
	continueFlag bool
	resumeID     string
	forkSession  bool
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "agentflow",
	Short: "AgentFlow - Agentic workflow framework",
	Long: `AgentFlow brings the power of structured AI development to everyone.

Supports free and local models: Ollama, Groq, Together, and any OpenAI-compatible API.
Provides composable skills for brainstorming, planning, TDD, debugging, and more.

Run without arguments to start an interactive session (like Claude Code).`,
	Version: version,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default behavior: start interactive REPL
		return startREPL()
	},
}

func startREPL() error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get provider and model from "provider/model" format
	defaultModel := cfg.Defaults.Main
	if defaultModel == "" {
		defaultModel = "ollama/llama3.3:latest"
	}

	// Extract provider name for display
	providerName := "ollama"
	modelName := defaultModel
	if parts := strings.Split(defaultModel, "/"); len(parts) >= 2 {
		providerName = parts[0]
		modelName = strings.Join(parts[1:], "/")
	}

	// Create TUI
	tuiModel := tui.New(providerName, modelName)

	// Create provider and agent for callbacks
	registry := cfg.BuildRegistry()
	provider, model, ok := registry.ResolveModel(defaultModel)
	if !ok {
		// Fallback to simple model name
		provider, model, _ = registry.ResolveModel(modelName)
	}

	skillLoader := skill.NewLoader(cfg.Skills.Paths)
	if err := skillLoader.Load(); err != nil {
		return fmt.Errorf("load skills: %w", err)
	}

	ag := agent.New(agent.Config{
		Provider: provider,
		Model:    model,
		Skills:   skillLoader,
	})

	// Set up submit callback
	tuiModel.SetOnSubmit(func(input string) tea.Cmd {
		return func() tea.Msg {
			ctx := context.Background()
			
			// Check for skill match
			matchedSkills := skillLoader.Match(input)
			if len(matchedSkills) > 0 {
				// Send skill matched message
				tui.SendSkillMatched(matchedSkills[0].Name)
			}

			// Stream response
			chunks, err := ag.Stream(ctx, input)
			if err != nil {
				return tui.SendError(err)()
			}

			// Process chunks in goroutine
			go func() {
				for chunk := range chunks {
					if chunk.Error != nil {
						// Handle error
						continue
					}
					// This won't work directly - need program reference
					// For now, simplified version
				}
			}()

			return nil
		}
	})

	// Run TUI
	p := tea.NewProgram(tuiModel, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

var runCmd = &cobra.Command{
	Use:   "run [message]",
	Short: "Run a single agent interaction",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()

		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		registry := cfg.BuildRegistry()
		
		// Resolve model
		model := modelSpec
		if model == "" {
			model = cfg.Defaults.Main
		}

		provider, modelName, ok := registry.ResolveModel(model)
		if !ok {
			return fmt.Errorf("unknown model: %s", model)
		}

		// Load skills
		skillLoader := skill.NewLoader(cfg.Skills.Paths)
		if err := skillLoader.Load(); err != nil {
			return fmt.Errorf("load skills: %w", err)
		}

		// Create agent
		a := agent.New(agent.Config{
			Provider: provider,
			Model:    modelName,
			Skills:   skillLoader,
		})

		message := strings.Join(args, " ")
		
		// Check for streaming flag
		stream, _ := cmd.Flags().GetBool("stream")
		if stream {
			chunks, err := a.Stream(ctx, message)
			if err != nil {
				return err
			}
			for chunk := range chunks {
				if chunk.Error != nil {
					return chunk.Error
				}
				fmt.Print(chunk.Content)
			}
			fmt.Println()
		} else {
			resp, err := a.Run(ctx, message)
			if err != nil {
				return err
			}
			fmt.Println(resp.Content)
		}

		return nil
	},
}

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Manage skills",
}

var skillListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available skills",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		loader := skill.NewLoader(cfg.Skills.Paths)
		if err := loader.Load(); err != nil {
			return err
		}

		skills := loader.List()
		if len(skills) == 0 {
			fmt.Println("No skills found")
			return nil
		}

		fmt.Printf("Found %d skill(s):\n\n", len(skills))
		for _, s := range skills {
			fmt.Printf("• %s\n", s.Name)
			if s.Description != "" {
				fmt.Printf("  %s\n", s.Description)
			}
			if len(s.Tags) > 0 {
				fmt.Printf("  Tags: %s\n", strings.Join(s.Tags, ", "))
			}
			fmt.Println()
		}

		return nil
	},
}

var skillRunCmd = &cobra.Command{
	Use:   "run [skill] [message]",
	Short: "Run with a specific skill",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()

		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		registry := cfg.BuildRegistry()
		
		model := modelSpec
		if model == "" {
			model = cfg.Defaults.Main
		}

		provider, modelName, ok := registry.ResolveModel(model)
		if !ok {
			return fmt.Errorf("unknown model: %s", model)
		}

		skillLoader := skill.NewLoader(cfg.Skills.Paths)
		if err := skillLoader.Load(); err != nil {
			return err
		}

		a := agent.New(agent.Config{
			Provider: provider,
			Model:    modelName,
			Skills:   skillLoader,
		})

		skillName := args[0]
		message := strings.Join(args[1:], " ")

		resp, err := a.RunWithSkill(ctx, skillName, message)
		if err != nil {
			return err
		}

		fmt.Println(resp.Content)
		return nil
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		fmt.Println("Providers:")
		for name, p := range cfg.Providers {
			fmt.Printf("  %s:\n", name)
			if p.BaseURL != "" {
				fmt.Printf("    URL: %s\n", p.BaseURL)
			}
			if len(p.Models) > 0 {
				fmt.Printf("    Models: %s\n", strings.Join(p.Models, ", "))
			}
		}

		fmt.Println("\nDefaults:")
		fmt.Printf("  Main: %s\n", cfg.Defaults.Main)
		fmt.Printf("  Subagent: %s\n", cfg.Defaults.Subagent)
		fmt.Printf("  Reviewer: %s\n", cfg.Defaults.Reviewer)

		return nil
	},
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration in current directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.DefaultConfig()
		path := ".agentflow/config.yaml"

		if err := cfg.Save(path); err != nil {
			return err
		}

		fmt.Printf("Created %s\n", path)

		// Create skills directory
		if err := os.MkdirAll(".agentflow/skills", 0755); err != nil {
			return err
		}
		fmt.Println("Created .agentflow/skills/")

		return nil
	},
}

var subagentCmd = &cobra.Command{
	Use:   "subagent [task]",
	Short: "Spawn a subagent for a task",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()

		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		registry := cfg.BuildRegistry()
		
		model := modelSpec
		if model == "" {
			model = cfg.Defaults.Subagent
		}

		provider, modelName, ok := registry.ResolveModel(model)
		if !ok {
			return fmt.Errorf("unknown model: %s", model)
		}

		skillLoader := skill.NewLoader(cfg.Skills.Paths)
		if err := skillLoader.Load(); err != nil {
			return err
		}

		pool := subagent.NewPool(subagent.PoolConfig{
			Provider:  provider,
			Model:     modelName,
			Skills:    skillLoader,
			MaxAgents: 5,
		})

		task := subagent.Task{
			ID:          "task-1",
			Description: "Execute user task",
			Message:     strings.Join(args, " "),
		}

		result, err := pool.Spawn(ctx, task)
		if err != nil {
			return err
		}

		if result.Error != nil {
			return result.Error
		}

		fmt.Printf("Agent: %s\n", result.AgentID)
		fmt.Printf("Duration: %v\n", result.Duration)
		fmt.Printf("\n%s\n", result.Response.Content)

		return nil
	},
}

var providersCmd = &cobra.Command{
	Use:   "providers",
	Short: "List configured providers",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		registry := cfg.BuildRegistry()
		providers := registry.List()

		if len(providers) == 0 {
			fmt.Println("No providers configured")
			return nil
		}

		fmt.Println("Configured providers:")
		for _, name := range providers {
			p, _ := registry.Get(name)
			models := p.Models()
			fmt.Printf("  %s: %d model(s)\n", name, len(models))
			for _, m := range models {
				fmt.Printf("    - %s/%s\n", name, m)
			}
		}

		return nil
	},
}

var sessionsCmd = &cobra.Command{
	Use:   "sessions",
	Short: "List saved sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := session.NewManager("")
		sessions, err := mgr.List()
		if err != nil {
			return err
		}

		if len(sessions) == 0 {
			fmt.Println("No saved sessions")
			return nil
		}

		workdir, _ := os.Getwd()
		fmt.Printf("Sessions (%d total):\n\n", len(sessions))

		for _, s := range sessions {
			marker := " "
			if s.Workdir == workdir {
				marker = "*"
			}

			name := s.DisplayName()
			fmt.Printf("%s [%s] %s\n", marker, s.ID, name)
			fmt.Printf("    %d msgs | %s | %s\n",
				len(s.Messages),
				s.Workdir,
				s.UpdatedAt.Format("Jan 2 15:04"))
		}

		fmt.Println("\n* = current directory")
		return nil
	},
}

var sessionDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a session",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := session.NewManager("")
		if err := mgr.Delete(args[0]); err != nil {
			return err
		}
		fmt.Printf("Deleted session: %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")
	rootCmd.PersistentFlags().StringVarP(&modelSpec, "model", "m", "", "model to use (provider/model)")

	// Session flags
	rootCmd.Flags().BoolVarP(&continueFlag, "continue", "c", false, "continue last session for current directory")
	rootCmd.Flags().StringVarP(&resumeID, "resume", "r", "", "resume a specific session by ID or name")
	rootCmd.Flags().BoolVar(&forkSession, "fork-session", false, "fork the session instead of continuing")

	runCmd.Flags().BoolP("stream", "s", false, "stream the response")

	skillCmd.AddCommand(skillListCmd)
	skillCmd.AddCommand(skillRunCmd)

	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configInitCmd)

	sessionsCmd.AddCommand(sessionDeleteCmd)

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(skillCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(subagentCmd)
	rootCmd.AddCommand(providersCmd)
	rootCmd.AddCommand(sessionsCmd)
}

func loadConfig() (*config.Config, error) {
	if cfgFile != "" {
		return config.Load(cfgFile)
	}
	return config.LoadDefault()
}
