package tui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/agentflow/agentflow/internal/input"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	// Colors
	primaryColor   = lipgloss.Color("#7C3AED") // Purple
	secondaryColor = lipgloss.Color("#10B981") // Green
	accentColor    = lipgloss.Color("#F59E0B") // Amber
	errorColor     = lipgloss.Color("#EF4444") // Red
	mutedColor     = lipgloss.Color("#6B7280") // Gray
	bgColor        = lipgloss.Color("#1F2937") // Dark gray

	// Styles
	mutedStyle = lipgloss.NewStyle().
			Foreground(mutedColor)
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(bgColor).
			Padding(0, 1)

	statusItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(primaryColor).
			Padding(0, 1)

	statusTextStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Padding(0, 1)

	userStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true)

	assistantStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	skillStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Italic(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor)
)

// Message types
type (
	responseMsg       string
	streamChunkMsg    string
	streamDoneMsg     struct{}
	errorMsg          error
	skillMatchedMsg   string
	tokensUpdatedMsg  int
	clearMsg          struct{}
	bashResultMsg     struct {
		Display string
		Context string
	}
)

// Model represents the TUI state
type Model struct {
	// UI components
	input    input.Model
	viewport viewport.Model
	spinner  spinner.Model

	// State
	messages     []ChatMessage
	streaming    bool
	currentResp  strings.Builder
	width        int
	height       int
	ready        bool
	err          error

	// Stats
	totalTokens   int
	sessionStart  time.Time
	lastSkill     string
	requestCount  int

	// Config
	provider string
	model    string

	// Callbacks
	onSubmit func(string) tea.Cmd
}

// ChatMessage represents a message in the conversation
type ChatMessage struct {
	Role      string // "user", "assistant", "system", "skill"
	Content   string
	Timestamp time.Time
}

// New creates a new TUI model
func New(provider, model string) Model {
	// Get current working directory for history
	workdir, _ := os.Getwd()

	// Create enhanced input
	inp := input.New(workdir)
	inp.SetPlaceholder("Type a message... (Ctrl+Enter to send, /help for commands, ! for bash)")

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(primaryColor)

	vp := viewport.New(80, 20)
	vp.SetContent("")

	return Model{
		input:        inp,
		viewport:     vp,
		spinner:      sp,
		messages:     make([]ChatMessage, 0),
		sessionStart: time.Now(),
		provider:     provider,
		model:        model,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.input.Init(),
		m.spinner.Tick,
	)
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			if m.streaming {
				m.streaming = false
				return m, nil
			}
			// Let input handle ctrl+c in non-normal modes
			if m.input.Mode() != input.ModeNormal {
				m.input, cmd = m.input.Update(msg)
				return m, cmd
			}
			return m, tea.Quit

		case "esc":
			if m.streaming {
				m.streaming = false
				return m, nil
			}
			// Let input handle esc in non-normal modes
			if m.input.Mode() != input.ModeNormal {
				m.input, cmd = m.input.Update(msg)
				return m, cmd
			}
			return m, tea.Quit

		case "ctrl+l":
			m.messages = make([]ChatMessage, 0)
			m.viewport.SetContent("")
			return m, nil

		case "pgup", "pgdown", "ctrl+u", "ctrl+d":
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}

	case input.SubmitMsg:
		return m.handleInputSubmit(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		headerHeight := 3
		footerHeight := 8 // Increased for autocomplete popup
		verticalMargin := headerHeight + footerHeight

		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - verticalMargin
		m.input.SetWidth(msg.Width - 4)

		m.viewport.SetContent(m.renderMessages())
		return m, nil

	case streamChunkMsg:
		m.currentResp.WriteString(string(msg))
		m.updateLastAssistantMessage(m.currentResp.String())
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
		return m, nil

	case streamDoneMsg:
		m.streaming = false
		m.requestCount++
		return m, nil

	case bashResultMsg:
		// Add bash result to conversation
		m.messages = append(m.messages, ChatMessage{
			Role:      "bash",
			Content:   msg.Display,
			Timestamp: time.Now(),
		})
		// Also add to context as a system message
		m.messages = append(m.messages, ChatMessage{
			Role:      "context",
			Content:   msg.Context,
			Timestamp: time.Now(),
		})
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
		return m, nil

	case skillMatchedMsg:
		m.lastSkill = string(msg)
		m.messages = append(m.messages, ChatMessage{
			Role:      "skill",
			Content:   fmt.Sprintf("Skill activated: %s", msg),
			Timestamp: time.Now(),
		})
		m.viewport.SetContent(m.renderMessages())
		return m, nil

	case tokensUpdatedMsg:
		m.totalTokens += int(msg)
		return m, nil

	case errorMsg:
		m.err = msg
		m.streaming = false
		m.messages = append(m.messages, ChatMessage{
			Role:      "system",
			Content:   fmt.Sprintf("Error: %v", msg),
			Timestamp: time.Now(),
		})
		m.viewport.SetContent(m.renderMessages())
		return m, nil

	case clearMsg:
		m.messages = make([]ChatMessage, 0)
		m.viewport.SetContent("")
		return m, nil

	case spinner.TickMsg:
		if m.streaming {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	// Update input
	if !m.streaming {
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Update spinner if streaming
	if m.streaming {
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// handleInputSubmit processes user input from the enhanced input component
func (m Model) handleInputSubmit(msg input.SubmitMsg) (tea.Model, tea.Cmd) {
	inputValue := strings.TrimSpace(msg.Value)
	if inputValue == "" {
		return m, nil
	}

	// Handle bash commands
	if msg.IsBash {
		return m.handleBashCommand(inputValue)
	}

	// Handle commands
	if strings.HasPrefix(inputValue, "/") {
		return m.handleCommand(inputValue)
	}

	// Add user message
	m.messages = append(m.messages, ChatMessage{
		Role:      "user",
		Content:   inputValue,
		Timestamp: time.Now(),
	})

	// Add empty assistant message for streaming
	m.messages = append(m.messages, ChatMessage{
		Role:      "assistant",
		Content:   "",
		Timestamp: time.Now(),
	})

	m.input.Reset()
	m.streaming = true
	m.currentResp.Reset()
	m.viewport.SetContent(m.renderMessages())
	m.viewport.GotoBottom()

	// Trigger the submit callback
	if m.onSubmit != nil {
		return m, m.onSubmit(inputValue)
	}

	return m, nil
}

// handleBashCommand executes a bash command and adds output to context
func (m Model) handleBashCommand(command string) (tea.Model, tea.Cmd) {
	m.input.Reset()
	m.viewport.SetContent(m.renderMessages())

	// Execute bash command asynchronously
	return m, func() tea.Msg {
		result := input.ExecuteBash(context.Background(), command)
		return bashResultMsg{
			Display: input.FormatBashResult(result),
			Context: input.FormatBashResultForContext(result),
		}
	}
}

// handleCommand processes slash commands
func (m Model) handleCommand(input string) (tea.Model, tea.Cmd) {
	parts := strings.Fields(input)
	cmd := strings.ToLower(parts[0])

	switch cmd {
	case "/quit", "/exit", "/q":
		return m, tea.Quit

	case "/help", "/h", "/?":
		help := m.renderHelp()
		m.messages = append(m.messages, ChatMessage{
			Role:      "system",
			Content:   help,
			Timestamp: time.Now(),
		})

	case "/clear", "/c":
		m.messages = make([]ChatMessage, 0)

	case "/model":
		if len(parts) > 1 {
			m.model = parts[1]
			m.messages = append(m.messages, ChatMessage{
				Role:      "system",
				Content:   fmt.Sprintf("Model changed to: %s", m.model),
				Timestamp: time.Now(),
			})
		} else {
			m.messages = append(m.messages, ChatMessage{
				Role:      "system",
				Content:   fmt.Sprintf("Current model: %s", m.model),
				Timestamp: time.Now(),
			})
		}

	case "/provider":
		if len(parts) > 1 {
			m.provider = parts[1]
			m.messages = append(m.messages, ChatMessage{
				Role:      "system",
				Content:   fmt.Sprintf("Provider changed to: %s", m.provider),
				Timestamp: time.Now(),
			})
		} else {
			m.messages = append(m.messages, ChatMessage{
				Role:      "system",
				Content:   fmt.Sprintf("Current provider: %s", m.provider),
				Timestamp: time.Now(),
			})
		}

	case "/status":
		status := m.renderStatus()
		m.messages = append(m.messages, ChatMessage{
			Role:      "system",
			Content:   status,
			Timestamp: time.Now(),
		})

	case "/skills":
		m.messages = append(m.messages, ChatMessage{
			Role:      "system",
			Content:   "Available skills:\n• brainstorming\n• writing-plans\n• subagent-driven-development\n• test-driven-development\n• systematic-debugging\n• verification-before-completion",
			Timestamp: time.Now(),
		})

	case "/compact":
		m.messages = append(m.messages, ChatMessage{
			Role:      "system",
			Content:   "Conversation compacted (not yet implemented)",
			Timestamp: time.Now(),
		})

	case "/history":
		m.messages = append(m.messages, ChatMessage{
			Role:      "system",
			Content:   fmt.Sprintf("Conversation has %d messages", len(m.messages)),
			Timestamp: time.Now(),
		})

	default:
		m.messages = append(m.messages, ChatMessage{
			Role:      "system",
			Content:   fmt.Sprintf("Unknown command: %s (type /help for available commands)", cmd),
			Timestamp: time.Now(),
		})
	}

	m.input.Reset()
	m.viewport.SetContent(m.renderMessages())
	m.viewport.GotoBottom()
	return m, nil
}

// updateLastAssistantMessage updates the last assistant message
func (m *Model) updateLastAssistantMessage(content string) {
	for i := len(m.messages) - 1; i >= 0; i-- {
		if m.messages[i].Role == "assistant" {
			m.messages[i].Content = content
			return
		}
	}
}

// Bash style
var bashStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#22C55E")).
	Bold(true)

var bashOutputStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#A3E635"))

// renderMessages renders all messages
func (m Model) renderMessages() string {
	var sb strings.Builder

	for _, msg := range m.messages {
		switch msg.Role {
		case "user":
			sb.WriteString(userStyle.Render("You") + " ")
			sb.WriteString(mutedStyle.Render(msg.Timestamp.Format("15:04")))
			sb.WriteString("\n")
			sb.WriteString(msg.Content)
			sb.WriteString("\n\n")

		case "assistant":
			sb.WriteString(assistantStyle.Render("Agent") + " ")
			sb.WriteString(mutedStyle.Render(msg.Timestamp.Format("15:04")))
			if m.streaming && msg == m.messages[len(m.messages)-1] {
				sb.WriteString(" " + m.spinner.View())
			}
			sb.WriteString("\n")
			sb.WriteString(msg.Content)
			sb.WriteString("\n\n")

		case "skill":
			sb.WriteString(skillStyle.Render("⚡ " + msg.Content))
			sb.WriteString("\n\n")

		case "bash":
			sb.WriteString(bashStyle.Render("🔧 Bash") + " ")
			sb.WriteString(mutedStyle.Render(msg.Timestamp.Format("15:04")))
			sb.WriteString("\n")
			sb.WriteString(bashOutputStyle.Render(msg.Content))
			sb.WriteString("\n")

		case "context":
			// Context messages are hidden from display but included in conversation
			continue

		case "system":
			sb.WriteString(helpStyle.Render(msg.Content))
			sb.WriteString("\n\n")
		}
	}

	return sb.String()
}

// renderHelp renders help text
func (m Model) renderHelp() string {
	return `
╭───────────────────────────────────────────────────────────────╮
│                      Available Commands                        │
├───────────────────────────────────────────────────────────────┤
│  /help, /h, /?     Show this help message                     │
│  /quit, /exit, /q  Exit the session                           │
│  /clear, /c        Clear conversation history                 │
│  /model [name]     Show or change current model               │
│  /provider [name]  Show or change provider                    │
│  /status           Show session statistics                    │
│  /skills           List available skills                      │
│  /compact          Compact conversation history               │
│  /history          Show conversation stats                    │
├───────────────────────────────────────────────────────────────┤
│                        Keyboard Shortcuts                      │
├───────────────────────────────────────────────────────────────┤
│  Ctrl+Enter        Send message                               │
│  Ctrl+L            Clear screen                               │
│  Ctrl+C / Esc      Cancel / Exit                              │
│  PgUp/PgDown       Scroll history                             │
│  ↑/↓               Navigate command history                   │
│  Ctrl+R            Reverse search history                     │
│  Tab               Autocomplete commands/files                │
│  Alt+Enter         Insert newline (multiline input)           │
│  \ + Enter         Continue on next line                      │
├───────────────────────────────────────────────────────────────┤
│                          Bash Mode                             │
├───────────────────────────────────────────────────────────────┤
│  !command          Execute bash command directly              │
│                    e.g., !git status, !ls -la                 │
│                    Output is added to conversation context    │
├───────────────────────────────────────────────────────────────┤
│                         Autocomplete                           │
├───────────────────────────────────────────────────────────────┤
│  /...              Complete slash commands                    │
│  @...              Complete file paths                        │
╰───────────────────────────────────────────────────────────────╯`
}

// renderStatus renders session status
func (m Model) renderStatus() string {
	duration := time.Since(m.sessionStart).Round(time.Second)
	return fmt.Sprintf(`
Session Status
──────────────
Provider: %s
Model: %s
Duration: %s
Requests: %d
Tokens: ~%d
Last Skill: %s
Messages: %d`,
		m.provider,
		m.model,
		duration,
		m.requestCount,
		m.totalTokens,
		m.lastSkill,
		len(m.messages),
	)
}

// View renders the UI
func (m Model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	// Header with mode indicator
	header := titleStyle.Render("🚀 AgentFlow") + "  "
	switch m.input.Mode() {
	case input.ModeReverseSearch:
		header += helpStyle.Render("Ctrl+R: search • Tab: accept • Esc: cancel")
	case input.ModeAutocomplete:
		header += helpStyle.Render("Tab/↓: next • Enter: accept • Esc: cancel")
	default:
		header += helpStyle.Render("Ctrl+Enter: send • /help • !cmd: bash • Ctrl+R: search")
	}

	// Main content
	content := m.viewport.View()

	// Input area
	inputBox := borderStyle.Render(m.input.View())

	// Status bar
	statusBar := m.renderStatusBar()

	// Combine all parts
	return fmt.Sprintf("%s\n%s\n%s\n%s", header, content, inputBox, statusBar)
}

// renderStatusBar renders the bottom status bar
func (m Model) renderStatusBar() string {
	// Left side: provider/model
	left := statusItemStyle.Render(fmt.Sprintf(" %s/%s ", m.provider, m.model))

	// Center: streaming indicator or skill
	var center string
	if m.streaming {
		center = statusTextStyle.Render(m.spinner.View() + " Generating...")
	} else if m.lastSkill != "" {
		center = statusTextStyle.Render("⚡ " + m.lastSkill)
	}

	// Right side: stats
	duration := time.Since(m.sessionStart).Round(time.Second)
	right := statusTextStyle.Render(fmt.Sprintf("↑%d msgs • %s", len(m.messages), duration))

	// Calculate padding
	totalWidth := m.width
	usedWidth := lipgloss.Width(left) + lipgloss.Width(center) + lipgloss.Width(right)
	padding := totalWidth - usedWidth
	if padding < 0 {
		padding = 0
	}

	spacer := strings.Repeat(" ", padding/2)

	return statusBarStyle.Width(m.width).Render(left + spacer + center + spacer + right)
}

// SetOnSubmit sets the callback for message submission
func (m *Model) SetOnSubmit(fn func(string) tea.Cmd) {
	m.onSubmit = fn
}

// SendStreamChunk sends a chunk to the TUI
func SendStreamChunk(chunk string) tea.Cmd {
	return func() tea.Msg {
		return streamChunkMsg(chunk)
	}
}

// SendStreamDone signals streaming is complete
func SendStreamDone() tea.Cmd {
	return func() tea.Msg {
		return streamDoneMsg{}
	}
}

// SendError sends an error to the TUI
func SendError(err error) tea.Cmd {
	return func() tea.Msg {
		return errorMsg(err)
	}
}

// SendSkillMatched signals a skill was matched
func SendSkillMatched(skill string) tea.Cmd {
	return func() tea.Msg {
		return skillMatchedMsg(skill)
	}
}

// SendTokensUpdated updates token count
func SendTokensUpdated(tokens int) tea.Cmd {
	return func() tea.Msg {
		return tokensUpdatedMsg(tokens)
	}
}
