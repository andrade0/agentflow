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
	"github.com/agentflow/agentflow/internal/session"
	"github.com/agentflow/agentflow/internal/skill"
	"github.com/agentflow/agentflow/pkg/types"
	"github.com/fatih/color"
)

// REPL represents the interactive Read-Eval-Print Loop
type REPL struct {
	config         *config.Config
	registry       *provider.Registry
	provider       provider.Provider
	model          string
	skills         *skill.Loader
	agent          *agent.Agent
	running        bool
	session        *session.Session
	sessionManager *session.Manager
	autoSave       bool
}

// Options configures REPL behavior
type Options struct {
	ContinueLast bool   // Continue last session for current workdir
	ResumeID     string // Resume specific session by ID or name
	ForkSession  bool   // Fork instead of continuing
}

// New creates a new REPL instance
func New(cfg *config.Config) (*REPL, error) {
	return NewWithOptions(cfg, Options{})
}

// NewWithOptions creates a REPL with session options
func NewWithOptions(cfg *config.Config, opts Options) (*REPL, error) {
	// Build provider registry
	registry := cfg.BuildRegistry()

	// Parse default model (format: provider/model)
	defaultModel := cfg.Defaults.Main
	if defaultModel == "" {
		defaultModel = "ollama/llama3.3:latest"
	}

	prov, model, ok := registry.ResolveModel(defaultModel)
	if !ok {
		return nil, fmt.Errorf("unknown model: %s", defaultModel)
	}

	// Load skills
	skillLoader := skill.NewLoader(cfg.Skills.Paths)
	if err := skillLoader.Load(); err != nil {
		return nil, fmt.Errorf("load skills: %w", err)
	}

	// Create agent
	ag := agent.New(agent.Config{
		Provider: prov,
		Model:    model,
		Skills:   skillLoader,
	})

	// Initialize session manager
	sessMgr := session.NewManager("")

	// Get current workdir and provider name
	workdir, _ := os.Getwd()
	providerName := strings.Split(defaultModel, "/")[0]

	// Handle session options
	var sess *session.Session
	var err error
	if opts.ResumeID != "" {
		// Resume specific session
		sess, err = sessMgr.GetByNameOrID(opts.ResumeID)
		if err != nil {
			return nil, fmt.Errorf("resume session: %w", err)
		}
		if opts.ForkSession {
			sess = sess.Clone()
		}
	} else if opts.ContinueLast {
		// Continue last session for this workdir
		sess, err = sessMgr.GetLatest(workdir)
		if err != nil {
			// No existing session, create new
			sess = session.New(workdir, providerName, model)
		} else if opts.ForkSession {
			sess = sess.Clone()
		}
	} else {
		// New session
		sess = session.New(workdir, providerName, model)
	}

	// Restore messages to agent
	for _, msg := range sess.Messages {
		ag.AddMessage(msg.Role, msg.Content)
	}

	return &REPL{
		config:         cfg,
		registry:       registry,
		provider:       prov,
		model:          model,
		skills:         skillLoader,
		agent:          ag,
		running:        false,
		session:        sess,
		sessionManager: sessMgr,
		autoSave:       true,
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

		// Auto-save session after each exchange
		r.autoSaveSession()
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

	gray.Printf("Provider: %s | Model: %s\n", r.session.Provider, r.model)

	// Show session info
	if r.session != nil {
		if len(r.session.Messages) > 0 {
			yellow := color.New(color.FgYellow)
			yellow.Printf("Resumed session: %s (%d messages)\n", r.session.ID, len(r.session.Messages))
		} else {
			gray.Printf("Session: %s\n", r.session.ID)
		}
	}

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
		r.agent.ClearHistory()
		r.session.Messages = nil
		r.autoSaveSession()
		fmt.Println("Conversation cleared.")
		return true

	case "/skills":
		r.listSkills()
		return true

	case "/model":
		if len(parts) > 1 {
			r.changeModel(parts[1])
		} else {
			fmt.Printf("Current model: %s\n", r.model)
		}
		return true

	case "/history":
		r.printHistory()
		return true

	case "/compact":
		fmt.Println("Compacting conversation history...")
		// TODO: Implement conversation compaction
		return true

	case "/sessions":
		r.listSessions()
		return true

	case "/resume":
		if len(parts) > 1 {
			r.resumeSession(parts[1])
		} else {
			r.showSessionPicker()
		}
		return true

	case "/rename":
		if len(parts) > 1 {
			name := strings.Join(parts[1:], " ")
			r.renameSession(name)
		} else {
			fmt.Println("Usage: /rename <name>")
		}
		return true

	case "/session":
		r.showCurrentSession()
		return true

	case "/save":
		r.saveSession()
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
	fmt.Println("  /history         Show conversation history")
	fmt.Println("  /compact         Compact conversation to save context")
	fmt.Println()
	cyan.Println("Session Commands:")
	fmt.Println()
	fmt.Println("  /sessions        List recent sessions")
	fmt.Println("  /session         Show current session info")
	fmt.Println("  /resume [id]     Resume a session (picker if no id)")
	fmt.Println("  /rename <name>   Rename current session")
	fmt.Println("  /save            Force save current session")
	fmt.Println()
	gray.Println("  Tip: Just type naturally to start working!")
	fmt.Println()
}

// listSkills lists available skills
func (r *REPL) listSkills() {
	skills := r.skills.List()
	cyan := color.New(color.FgCyan)

	fmt.Println()
	cyan.Println("Available Skills:")
	fmt.Println()
	for _, s := range skills {
		fmt.Printf("  • %s\n", s.Name)
		if s.Description != "" {
			color.HiBlack("    %s\n", s.Description)
		}
	}
	fmt.Println()
}

// printHistory prints conversation history
func (r *REPL) printHistory() {
	messages := r.agent.Messages()
	if len(messages) == 0 {
		fmt.Println("No conversation history.")
		return
	}

	fmt.Println()
	for _, msg := range messages {
		if msg.Role == "user" {
			color.Green("You: %s", truncate(msg.Content, 100))
		} else if msg.Role == "assistant" {
			color.Cyan("Agent: %s", truncate(msg.Content, 100))
		}
	}
	fmt.Println()
}

// processInput processes user input and generates a response
func (r *REPL) processInput(ctx context.Context, input string) error {
	// Match skill
	matchedSkills := r.skills.Match(input)
	if len(matchedSkills) > 0 {
		color.HiBlack("\n[Skill: %s]\n", matchedSkills[0].Name)
	}

	// Generate response with streaming
	cyan := color.New(color.FgCyan, color.Bold)
	cyan.Print("\nAgent > ")

	// Reset color for response
	color.Unset()

	// Stream response
	var fullResponse strings.Builder
	chunks, err := r.agent.Stream(ctx, input)
	if err != nil {
		return err
	}

	for chunk := range chunks {
		if chunk.Error != nil {
			return chunk.Error
		}
		fmt.Print(chunk.Content)
		fullResponse.WriteString(chunk.Content)
	}
	fmt.Println()
	fmt.Println()

	return nil
}

// changeModel changes the active model
func (r *REPL) changeModel(modelSpec string) {
	prov, model, ok := r.registry.ResolveModel(modelSpec)
	if !ok {
		color.Red("Unknown model: %s", modelSpec)
		return
	}

	r.provider = prov
	r.model = model
	r.agent = agent.New(agent.Config{
		Provider: prov,
		Model:    model,
		Skills:   r.skills,
	})

	// Restore messages
	for _, msg := range r.session.Messages {
		r.agent.AddMessage(msg.Role, msg.Content)
	}

	fmt.Printf("Model changed to: %s\n", model)
}

// truncate truncates a string to maxLen characters
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// listSessions shows recent sessions
func (r *REPL) listSessions() {
	sessions, err := r.sessionManager.List()
	if err != nil {
		color.Red("Error listing sessions: %v", err)
		return
	}

	if len(sessions) == 0 {
		fmt.Println("No saved sessions.")
		return
	}

	cyan := color.New(color.FgCyan)
	gray := color.New(color.FgHiBlack)
	yellow := color.New(color.FgYellow)

	fmt.Println()
	cyan.Println("Recent Sessions:")
	fmt.Println()

	// Show max 10 sessions
	limit := 10
	if len(sessions) < limit {
		limit = len(sessions)
	}

	for i, s := range sessions[:limit] {
		// Mark current session
		marker := " "
		if r.session != nil && s.ID == r.session.ID {
			marker = "*"
			yellow.Printf("%s ", marker)
		} else {
			fmt.Printf("%s ", marker)
		}

		// Session ID and name
		cyan.Printf("[%s]", s.ID)
		if s.Name != "" {
			fmt.Printf(" %s", s.Name)
		}
		fmt.Println()

		// Details
		gray.Printf("    %d messages | %s | %s\n",
			len(s.Messages),
			s.Workdir,
			s.UpdatedAt.Format("Jan 2 15:04"))

		if i < limit-1 {
			fmt.Println()
		}
	}
	fmt.Println()
}

// showCurrentSession shows info about the current session
func (r *REPL) showCurrentSession() {
	if r.session == nil {
		fmt.Println("No active session.")
		return
	}

	cyan := color.New(color.FgCyan)

	fmt.Println()
	cyan.Println("Current Session:")
	fmt.Println()
	fmt.Printf("  ID:       %s\n", r.session.ID)
	if r.session.Name != "" {
		fmt.Printf("  Name:     %s\n", r.session.Name)
	}
	fmt.Printf("  Workdir:  %s\n", r.session.Workdir)
	fmt.Printf("  Provider: %s\n", r.session.Provider)
	fmt.Printf("  Model:    %s\n", r.session.Model)
	fmt.Printf("  Messages: %d\n", len(r.session.Messages))
	fmt.Printf("  Created:  %s\n", r.session.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Updated:  %s\n", r.session.UpdatedAt.Format("2006-01-02 15:04:05"))
	fmt.Println()
}

// resumeSession resumes a specific session
func (r *REPL) resumeSession(idOrName string) {
	sess, err := r.sessionManager.GetByNameOrID(idOrName)
	if err != nil {
		color.Red("Session not found: %s", idOrName)
		return
	}

	r.session = sess

	// Restore to agent
	r.agent.ClearHistory()
	for _, msg := range sess.Messages {
		r.agent.AddMessage(msg.Role, msg.Content)
	}

	color.Green("Resumed session %s (%d messages)", sess.ID, len(sess.Messages))
}

// showSessionPicker shows an interactive session picker
func (r *REPL) showSessionPicker() {
	sessions, err := r.sessionManager.List()
	if err != nil {
		color.Red("Error: %v", err)
		return
	}

	if len(sessions) == 0 {
		fmt.Println("No saved sessions.")
		return
	}

	fmt.Println()
	fmt.Println("Select a session to resume:")
	fmt.Println()

	limit := 10
	if len(sessions) < limit {
		limit = len(sessions)
	}

	for i, s := range sessions[:limit] {
		preview := s.DisplayName()
		fmt.Printf("  %d. [%s] %s\n", i+1, s.ID, preview)
	}

	fmt.Println()
	fmt.Print("Enter number or ID: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return
	}

	// Try as number
	var idx int
	if _, err := fmt.Sscanf(input, "%d", &idx); err == nil && idx > 0 && idx <= limit {
		r.resumeSession(sessions[idx-1].ID)
		return
	}

	// Try as ID
	r.resumeSession(input)
}

// renameSession renames the current session
func (r *REPL) renameSession(name string) {
	if r.session == nil {
		color.Red("No active session.")
		return
	}

	r.session.Name = name
	if err := r.sessionManager.Save(r.session); err != nil {
		color.Red("Error saving session: %v", err)
		return
	}

	color.Green("Session renamed to: %s", name)
}

// saveSession forces a save of the current session
func (r *REPL) saveSession() {
	if r.session == nil {
		color.Red("No active session.")
		return
	}

	if err := r.sessionManager.Save(r.session); err != nil {
		color.Red("Error saving session: %v", err)
		return
	}

	color.Green("Session saved: %s", r.session.ID)
}

// autoSaveSession saves after each exchange
func (r *REPL) autoSaveSession() {
	if !r.autoSave || r.session == nil {
		return
	}

	// Sync agent messages to session
	r.session.Messages = make([]types.Message, 0)
	for _, msg := range r.agent.Messages() {
		r.session.Messages = append(r.session.Messages, msg)
	}
	r.session.UpdatedAt = r.session.LastActivity()

	r.sessionManager.Save(r.session)
}
