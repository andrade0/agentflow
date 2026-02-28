package repl

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/agentflow/agentflow/internal/agent"
	"github.com/agentflow/agentflow/internal/config"
	"github.com/agentflow/agentflow/internal/provider"
	"github.com/agentflow/agentflow/internal/skill"
	"github.com/fatih/color"
)

// REPL represents the interactive Read-Eval-Print Loop
type REPL struct {
	config       *config.Config
	provider     provider.Provider
	skillManager *skill.Manager
	agent        *agent.Agent
	history      []Message
	running      bool
}

// Message represents a conversation message
type Message struct {
	Role    string // "user" or "assistant"
	Content string
}

// New creates a new REPL instance
func New(cfg *config.Config) (*REPL, error) {
	// Initialize provider
	p, err := provider.New(cfg.Defaults.Provider, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	// Initialize skill manager
	sm, err := skill.NewManager(cfg.SkillPaths)
	if err != nil {
		return nil, fmt.Errorf("failed to create skill manager: %w", err)
	}

	// Initialize agent
	ag := agent.New(p, sm, cfg)

	return &REPL{
		config:       cfg,
		provider:     p,
		skillManager: sm,
		agent:        ag,
		history:      make([]Message, 0),
		running:      false,
	}, nil
}

// Run starts the interactive REPL session
func (r *REPL) Run(ctx context.Context) error {
	r.running = true

	// Handle Ctrl+C gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n\nSession ended. Goodbye!")
		r.running = false
		os.Exit(0)
	}()

	// Print welcome message
	r.printWelcome()

	// Main REPL loop
	reader := bufio.NewReader(os.Stdin)
	for r.running {
		// Print prompt
		r.printPrompt()

		// Read user input
		input, err := reader.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				fmt.Println("\nSession ended. Goodbye!")
				break
			}
			return fmt.Errorf("failed to read input: %w", err)
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Handle special commands
		if r.handleCommand(input) {
			continue
		}

		// Process the input with the agent
		if err := r.processInput(ctx, input); err != nil {
			color.Red("Error: %v", err)
		}
	}

	return nil
}

// printWelcome prints the welcome message
func (r *REPL) printWelcome() {
	cyan := color.New(color.FgCyan, color.Bold)
	gray := color.New(color.FgHiBlack)

	fmt.Println()
	cyan.Println("╭─────────────────────────────────────────────────────────────╮")
	cyan.Println("│                    AgentFlow v0.1.0                         │")
	cyan.Println("│           Superpowers for everyone 🚀                       │")
	cyan.Println("╰─────────────────────────────────────────────────────────────╯")
	fmt.Println()

	gray.Printf("Provider: %s | Model: %s\n", r.config.Defaults.Provider, r.config.Defaults.Model)
	gray.Println("Type /help for commands, /quit to exit")
	fmt.Println()
}

// printPrompt prints the input prompt
func (r *REPL) printPrompt() {
	green := color.New(color.FgGreen, color.Bold)
	green.Print("You > ")
}

// handleCommand handles special REPL commands
func (r *REPL) handleCommand(input string) bool {
	if !strings.HasPrefix(input, "/") {
		return false
	}

	parts := strings.Fields(input)
	cmd := strings.ToLower(parts[0])

	switch cmd {
	case "/quit", "/exit", "/q":
		fmt.Println("Session ended. Goodbye!")
		r.running = false
		return true

	case "/help", "/h":
		r.printHelp()
		return true

	case "/clear":
		r.history = make([]Message, 0)
		fmt.Println("Conversation cleared.")
		return true

	case "/skills":
		r.listSkills()
		return true

	case "/model":
		if len(parts) > 1 {
			r.config.Defaults.Model = parts[1]
			fmt.Printf("Model changed to: %s\n", parts[1])
		} else {
			fmt.Printf("Current model: %s\n", r.config.Defaults.Model)
		}
		return true

	case "/provider":
		if len(parts) > 1 {
			r.config.Defaults.Provider = parts[1]
			fmt.Printf("Provider changed to: %s\n", parts[1])
		} else {
			fmt.Printf("Current provider: %s\n", r.config.Defaults.Provider)
		}
		return true

	case "/history":
		r.printHistory()
		return true

	case "/compact":
		fmt.Println("Compacting conversation history...")
		// TODO: Implement conversation compaction
		return true

	default:
		color.Yellow("Unknown command: %s (type /help for available commands)", cmd)
		return true
	}
}

// printHelp prints available commands
func (r *REPL) printHelp() {
	cyan := color.New(color.FgCyan)
	gray := color.New(color.FgHiBlack)

	fmt.Println()
	cyan.Println("Available Commands:")
	fmt.Println()
	fmt.Println("  /help, /h        Show this help message")
	fmt.Println("  /quit, /exit, /q Exit the session")
	fmt.Println("  /clear           Clear conversation history")
	fmt.Println("  /skills          List available skills")
	fmt.Println("  /model [name]    Show or change current model")
	fmt.Println("  /provider [name] Show or change current provider")
	fmt.Println("  /history         Show conversation history")
	fmt.Println("  /compact         Compact conversation to save context")
	fmt.Println()
	gray.Println("  Tip: Just type naturally to start working!")
	fmt.Println()
}

// listSkills lists available skills
func (r *REPL) listSkills() {
	skills := r.skillManager.List()
	cyan := color.New(color.FgCyan)

	fmt.Println()
	cyan.Println("Available Skills:")
	fmt.Println()
	for _, s := range skills {
		fmt.Printf("  • %s\n", s.Name)
		color.HiBlack("    %s\n", s.Description)
	}
	fmt.Println()
}

// printHistory prints conversation history
func (r *REPL) printHistory() {
	if len(r.history) == 0 {
		fmt.Println("No conversation history.")
		return
	}

	fmt.Println()
	for _, msg := range r.history {
		if msg.Role == "user" {
			color.Green("You: %s", msg.Content)
		} else {
			color.Cyan("Agent: %s", truncate(msg.Content, 100))
		}
	}
	fmt.Println()
}

// processInput processes user input and generates a response
func (r *REPL) processInput(ctx context.Context, input string) error {
	// Add to history
	r.history = append(r.history, Message{Role: "user", Content: input})

	// Match skill
	matchedSkill := r.skillManager.Match(input)
	if matchedSkill != nil {
		color.HiBlack("\n[Skill: %s]\n", matchedSkill.Name)
	}

	// Generate response with streaming
	cyan := color.New(color.FgCyan, color.Bold)
	cyan.Print("\nAgent > ")

	// Reset color for response
	color.Unset()

	// Stream response
	var fullResponse strings.Builder
	responseChan, err := r.agent.StreamChat(ctx, r.history, matchedSkill)
	if err != nil {
		return err
	}

	for chunk := range responseChan {
		fmt.Print(chunk)
		fullResponse.WriteString(chunk)
	}
	fmt.Println()
	fmt.Println()

	// Add response to history
	r.history = append(r.history, Message{Role: "assistant", Content: fullResponse.String()})

	return nil
}

// truncate truncates a string to maxLen characters
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
