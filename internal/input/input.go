package input

import (
	"strings"

	"github.com/agentflow/agentflow/internal/history"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Mode represents the current input mode
type Mode int

const (
	ModeNormal Mode = iota
	ModeReverseSearch
	ModeAutocomplete
)

// Styles
var (
	searchPromptStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F59E0B")).
				Bold(true)

	searchMatchStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#F59E0B")).
				Foreground(lipgloss.Color("#000000"))

	completionStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7C3AED")).
			Padding(0, 1)

	completionSelectedStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#7C3AED")).
				Foreground(lipgloss.Color("#FFFFFF"))

	completionDescStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6B7280"))
)

// Model represents the enhanced input model
type Model struct {
	textarea textarea.Model
	history  *history.History
	completer *Completer

	// State
	mode              Mode
	searchQuery       string
	searchResults     []history.SearchResult
	searchIndex       int
	completions       []Completion
	completionIndex   int
	savedInput        string // Input saved before entering search mode
	multilineEnabled  bool
	width             int
}

// SubmitMsg is sent when the user submits input
type SubmitMsg struct {
	Value     string
	IsBash    bool // True if input starts with !
}

// New creates a new enhanced input model
func New(workdir string) Model {
	ta := textarea.New()
	ta.Placeholder = "Type a message... (Ctrl+Enter to send, /help for commands)"
	ta.Focus()
	ta.Prompt = "│ "
	ta.CharLimit = 8192
	ta.SetWidth(80)
	ta.SetHeight(3)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false

	hist, _ := history.New(workdir)

	return Model{
		textarea:         ta,
		history:          hist,
		completer:        NewCompleter(),
		mode:             ModeNormal,
		multilineEnabled: true,
		completions:      nil,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.textarea.SetWidth(msg.Width - 4)
		return m, nil
	}

	// Forward to textarea in normal mode
	if m.mode == ModeNormal {
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd
	}

	return m, nil
}

// handleKey processes key input
func (m Model) handleKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	key := msg.String()

	// Global keys
	switch key {
	case "ctrl+c":
		if m.mode != ModeNormal {
			m.cancelMode()
			return m, nil
		}
		return m, tea.Quit
	}

	// Mode-specific handling
	switch m.mode {
	case ModeReverseSearch:
		return m.handleReverseSearchKey(msg)
	case ModeAutocomplete:
		return m.handleAutocompleteKey(msg)
	default:
		return m.handleNormalKey(msg)
	}
}

// handleNormalKey handles keys in normal mode
func (m Model) handleNormalKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "ctrl+r":
		// Enter reverse search mode
		m.mode = ModeReverseSearch
		m.searchQuery = ""
		m.searchResults = nil
		m.searchIndex = 0
		m.savedInput = m.textarea.Value()
		return m, nil

	case "ctrl+enter", "ctrl+s":
		// Submit
		return m.submit()

	case "up":
		// History previous
		if m.textarea.Line() == 0 {
			if prev, ok := m.history.Previous(); ok {
				m.textarea.SetValue(prev)
				m.textarea.CursorEnd()
			}
			return m, nil
		}

	case "down":
		// History next
		lines := strings.Split(m.textarea.Value(), "\n")
		if m.textarea.Line() == len(lines)-1 {
			if next, ok := m.history.Next(); ok {
				m.textarea.SetValue(next)
				m.textarea.CursorEnd()
			} else {
				m.textarea.SetValue("")
			}
			return m, nil
		}

	case "tab":
		// Trigger autocomplete
		input := m.textarea.Value()
		cursorPos := m.getCursorPosition()
		completions := m.completer.Complete(input, cursorPos)
		if len(completions) > 0 {
			if len(completions) == 1 {
				// Single completion - apply directly
				m.applyCompletion(completions[0])
			} else {
				// Multiple completions - show popup
				m.mode = ModeAutocomplete
				m.completions = completions
				m.completionIndex = 0
			}
		}
		return m, nil

	case "alt+enter", "alt+j":
		// Insert newline (Option+Enter on macOS)
		m.textarea.InsertString("\n")
		return m, nil

	case "enter":
		// Check for backslash continuation
		input := m.textarea.Value()
		if strings.HasSuffix(input, "\\") {
			// Remove backslash and add newline
			m.textarea.SetValue(strings.TrimSuffix(input, "\\") + "\n")
			m.textarea.CursorEnd()
			return m, nil
		}
		// If multiline and not empty, check if we should submit or add newline
		if m.multilineEnabled && input != "" && !strings.HasPrefix(input, "/") && !strings.HasPrefix(input, "!") {
			// In a multiline context, enter adds a newline
			// Use Ctrl+Enter to submit
			m.textarea.InsertString("\n")
			return m, nil
		}
		// Single line commands submit on enter
		if strings.HasPrefix(input, "/") || strings.HasPrefix(input, "!") {
			return m.submit()
		}
	}

	// Forward to textarea
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

// handleReverseSearchKey handles keys in reverse search mode
func (m Model) handleReverseSearchKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "ctrl+r":
		// Cycle to next result
		if len(m.searchResults) > 0 {
			m.searchIndex = (m.searchIndex + 1) % len(m.searchResults)
			m.applySearchResult()
		}
		return m, nil

	case "tab", "enter":
		// Accept current result
		m.mode = ModeNormal
		m.history.Reset()
		return m, nil

	case "esc":
		// Cancel search, restore saved input
		m.textarea.SetValue(m.savedInput)
		m.cancelMode()
		return m, nil

	case "backspace":
		if len(m.searchQuery) > 0 {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
			m.updateSearch()
		}
		return m, nil

	default:
		// Add character to search query
		if len(key) == 1 && key[0] >= 32 {
			m.searchQuery += key
			m.updateSearch()
		}
		return m, nil
	}
}

// handleAutocompleteKey handles keys in autocomplete mode
func (m Model) handleAutocompleteKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "tab", "down":
		// Next completion
		m.completionIndex = (m.completionIndex + 1) % len(m.completions)
		return m, nil

	case "up", "shift+tab":
		// Previous completion
		m.completionIndex--
		if m.completionIndex < 0 {
			m.completionIndex = len(m.completions) - 1
		}
		return m, nil

	case "enter":
		// Accept completion
		m.applyCompletion(m.completions[m.completionIndex])
		m.cancelMode()
		return m, nil

	case "esc":
		// Cancel
		m.cancelMode()
		return m, nil

	default:
		// Exit autocomplete and forward key
		m.cancelMode()
		return m.handleNormalKey(msg)
	}
}

// updateSearch updates search results
func (m *Model) updateSearch() {
	m.searchResults = m.history.Search(m.searchQuery)
	m.searchIndex = 0
	m.applySearchResult()
}

// applySearchResult applies the current search result to textarea
func (m *Model) applySearchResult() {
	if len(m.searchResults) > 0 && m.searchIndex < len(m.searchResults) {
		m.textarea.SetValue(m.searchResults[m.searchIndex].Entry)
		m.textarea.CursorEnd()
	}
}

// getCursorPosition calculates the cursor position in the text
func (m Model) getCursorPosition() int {
	value := m.textarea.Value()
	lines := strings.Split(value, "\n")
	currentLine := m.textarea.Line()

	pos := 0
	for i := 0; i < currentLine && i < len(lines); i++ {
		pos += len(lines[i]) + 1 // +1 for newline
	}

	// Add column offset within current line
	if currentLine < len(lines) {
		info := m.textarea.LineInfo()
		pos += info.CharOffset
	}

	return pos
}

// applyCompletion applies a completion to the input
func (m *Model) applyCompletion(comp Completion) {
	input := m.textarea.Value()
	cursorPos := m.getCursorPosition()

	// Find the start of the word to replace
	wordStart := cursorPos
	for wordStart > 0 && wordStart <= len(input) {
		ch := input[wordStart-1]
		if ch == ' ' || ch == '\t' || ch == '\n' {
			break
		}
		wordStart--
	}

	// Build new input
	var newInput string
	if wordStart > 0 {
		newInput = input[:wordStart]
	}
	newInput += comp.Value

	// Add remaining text after cursor
	if cursorPos < len(input) {
		newInput += input[cursorPos:]
	}

	// Add space after command completion
	if comp.Type == CompletionCommand && !strings.HasSuffix(newInput, " ") {
		newInput += " "
	}

	m.textarea.SetValue(newInput)
	m.textarea.CursorEnd()
}

// cancelMode returns to normal mode
func (m *Model) cancelMode() {
	m.mode = ModeNormal
	m.searchQuery = ""
	m.searchResults = nil
	m.completions = nil
}

// submit submits the current input
func (m Model) submit() (Model, tea.Cmd) {
	input := strings.TrimSpace(m.textarea.Value())
	if input == "" {
		return m, nil
	}

	// Add to history
	m.history.Add(input)
	m.history.Reset()

	// Check if it's a bash command
	isBash := strings.HasPrefix(input, "!")
	if isBash {
		input = strings.TrimPrefix(input, "!")
	}

	m.textarea.Reset()
	m.cancelMode()

	return m, func() tea.Msg {
		return SubmitMsg{Value: input, IsBash: isBash}
	}
}

// View renders the input
func (m Model) View() string {
	var sb strings.Builder

	// Render textarea
	sb.WriteString(m.textarea.View())

	// Render mode-specific UI
	switch m.mode {
	case ModeReverseSearch:
		sb.WriteString("\n")
		sb.WriteString(m.renderReverseSearch())

	case ModeAutocomplete:
		sb.WriteString("\n")
		sb.WriteString(m.renderAutocomplete())
	}

	return sb.String()
}

// renderReverseSearch renders the reverse search UI
func (m Model) renderReverseSearch() string {
	prompt := searchPromptStyle.Render("(reverse-i-search)`")
	query := m.searchQuery
	prompt += query
	prompt += searchPromptStyle.Render("': ")

	if len(m.searchResults) > 0 && m.searchIndex < len(m.searchResults) {
		result := m.searchResults[m.searchIndex]
		// Highlight the match
		before := result.Entry[:result.MatchStart]
		match := searchMatchStyle.Render(result.Entry[result.MatchStart:result.MatchEnd])
		after := result.Entry[result.MatchEnd:]
		prompt += before + match + after

		// Show result count
		prompt += lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).
			Render(" (" + string(rune('0'+m.searchIndex+1)) + "/" + string(rune('0'+len(m.searchResults))) + ")")
	} else if m.searchQuery != "" {
		prompt += lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).
			Render("(no match)")
	}

	return prompt
}

// renderAutocomplete renders the autocomplete popup
func (m Model) renderAutocomplete() string {
	var lines []string

	for i, comp := range m.completions {
		line := comp.Display
		if comp.Description != "" {
			line += " " + completionDescStyle.Render("- "+comp.Description)
		}

		if i == m.completionIndex {
			line = completionSelectedStyle.Render("▸ " + line)
		} else {
			line = "  " + line
		}

		lines = append(lines, line)
	}

	return completionStyle.Render(strings.Join(lines, "\n"))
}

// SetWidth sets the input width
func (m *Model) SetWidth(w int) {
	m.width = w
	m.textarea.SetWidth(w - 4)
}

// SetHeight sets the input height
func (m *Model) SetHeight(h int) {
	m.textarea.SetHeight(h)
}

// Focus focuses the input
func (m *Model) Focus() {
	m.textarea.Focus()
}

// Blur blurs the input
func (m *Model) Blur() {
	m.textarea.Blur()
}

// Value returns the current input value
func (m Model) Value() string {
	return m.textarea.Value()
}

// SetValue sets the input value
func (m *Model) SetValue(v string) {
	m.textarea.SetValue(v)
}

// Reset resets the input
func (m *Model) Reset() {
	m.textarea.Reset()
	m.cancelMode()
}

// SetPlaceholder sets the placeholder text
func (m *Model) SetPlaceholder(p string) {
	m.textarea.Placeholder = p
}

// Mode returns the current input mode
func (m Model) Mode() Mode {
	return m.mode
}

// IsMultilineEnabled returns whether multiline is enabled
func (m Model) IsMultilineEnabled() bool {
	return m.multilineEnabled
}

// SetMultilineEnabled enables/disables multiline
func (m *Model) SetMultilineEnabled(enabled bool) {
	m.multilineEnabled = enabled
}

// History returns the history manager
func (m Model) History() *history.History {
	return m.history
}
