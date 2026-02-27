package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lixianmin/pc/internal/gateway"
)

// Message represents a chat message.
type Message struct {
	Role      string    // "user" or "agent"
	Content   string
	Timestamp time.Time
}

// Model represents the TUI application state.
type Model struct {
	// UI components
	viewport viewport.Model
	textarea textarea.Model

	// State
	messages   []Message
	history    *History
	rpcClient  *gateway.RPCClient
	sessionID  string
	status     string
	width      int
	height     int
	quitting   bool
	ready      bool

	// Styles
	styles *Styles
}

// Styles holds the UI styles.
type Styles struct {
	Header       lipgloss.Style
	UserMessage  lipgloss.Style
	AgentMessage lipgloss.Style
	StatusBar    lipgloss.Style
	InputPrompt  lipgloss.Style
	Error        lipgloss.Style
}

// DefaultStyles returns the default styles.
func DefaultStyles() *Styles {
	return &Styles{
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			Padding(0, 1),

		UserMessage: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			PaddingLeft(2),

		AgentMessage: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A8A8A8")).
			PaddingLeft(2),

		StatusBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Padding(0, 1),

		InputPrompt: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true),

		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")),
	}
}

// NewModel creates a new TUI model.
func NewModel(rpcClient *gateway.RPCClient) *Model {
	ta := textarea.New()
	ta.Placeholder = "Type a message or /command..."
	ta.SetWidth(80)
	ta.SetHeight(3)
	ta.Focus()

	// Get history file path
	home, _ := os.UserHomeDir()
	historyFile := filepath.Join(home, ".pc", "history")

	return &Model{
		textarea:  ta,
		history:   NewHistory(historyFile),
		rpcClient: rpcClient,
		sessionID: generateSessionID(),
		messages:  make([]Message, 0),
		styles:    DefaultStyles(),
		status:    "Connected",
	}
}

// Init initializes the model.
func (m *Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		tea.EnterAltScreen,
		textarea.Blink,
	}

	// Load history
	if err := m.history.Load(); err == nil {
		// History loaded successfully
	}

	// Send initial greeting
	cmds = append(cmds, m.sendMessageCmd("Hello! I'm your PersonalClaw assistant. How can I help you today?", true))

	return tea.Batch(cmds...)
}

// Update handles messages and updates the model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		taCmd tea.Cmd
		vpCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if !m.ready {
			// Initialize viewport
			m.viewport = viewport.New(msg.Width, msg.Height-6)
			m.viewport.SetContent(m.renderMessages())
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 6
		}

		m.textarea.SetWidth(msg.Width - 4)

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit

		case tea.KeyEnter:
			if msg.Alt {
				// Alt+Enter for newline
				m.textarea.InsertString("\n")
			} else {
				// Send message
				input := strings.TrimSpace(m.textarea.Value())
				if input != "" {
					m.textarea.SetValue("")
					return m, m.handleInput(input)
				}
			}

		case tea.KeyUp:
			if m.textarea.Value() == "" {
				// Load previous history item
				if item := m.history.Previous(); item != "" {
					m.textarea.SetValue(item)
					m.textarea.CursorEnd()
				}
			}

		case tea.KeyDown:
			if m.textarea.Value() == "" {
				// Load next history item
				if item := m.history.Next(); item != "" {
					m.textarea.SetValue(item)
					m.textarea.CursorEnd()
				} else {
					m.textarea.SetValue("")
				}
			}

		case tea.KeyTab:
			// Tab completion
			input := m.textarea.Value()
			if strings.HasPrefix(input, "/") {
				// Command completion
				completed := m.completeCommand(input)
				if completed != input {
					m.textarea.SetValue(completed + " ")
					m.textarea.CursorEnd()
				}
			}
		}

	case responseMsg:
		m.status = "Connected"
		m.addMessage("agent", string(msg))
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()

	case errorMsg:
		m.status = fmt.Sprintf("Error: %v", msg)
		m.addMessage("agent", fmt.Sprintf("Error: %v", msg))
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()

	case statusMsg:
		m.status = string(msg)
	}

	m.textarea, taCmd = m.textarea.Update(msg)
	if m.ready {
		m.viewport, vpCmd = m.viewport.Update(msg)
	}

	return m, tea.Batch(taCmd, vpCmd)
}

// View renders the UI.
func (m *Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	if m.quitting {
		return "Goodbye!\n"
	}

	// Header
	header := m.styles.Header.Render("PersonalClaw v0.1.0")

	// Status bar
	status := m.styles.StatusBar.Render(fmt.Sprintf("Status: %s | Session: %s", m.status, m.sessionID[:8]))

	// Input area
	prompt := m.styles.InputPrompt.Render("> ")
	input := m.textarea.View()

	// Combine all parts
	return fmt.Sprintf(
		"%s\n%s\n%s\n%s%s",
		header,
		m.viewport.View(),
		status,
		prompt,
		input,
	)
}

// handleInput processes user input.
func (m *Model) handleInput(input string) tea.Cmd {
	// Add to history
	m.history.Add(input)
	m.history.Save()

	// Check for commands
	if strings.HasPrefix(input, "/") {
		return m.handleCommand(input)
	}

	// Regular message
	m.addMessage("user", input)
	m.viewport.SetContent(m.renderMessages())
	m.viewport.GotoBottom()

	// Send to agent
	return m.sendToAgent(input)
}

// handleCommand handles slash commands.
func (m *Model) handleCommand(cmd string) tea.Cmd {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return nil
	}

	command := parts[0]

	switch command {
	case "/quit", "/q":
		m.quitting = true
		return tea.Quit

	case "/clear", "/c":
		m.messages = make([]Message, 0)
		m.viewport.SetContent(m.renderMessages())
		return nil

	case "/skills", "/s":
		return m.listSkills()

	case "/status":
		return m.showStatus()

	case "/help", "/h":
		help := `Available commands:
  /quit, /q     - Exit the TUI
  /clear, /c    - Clear the screen
  /skills, /s   - List available skills
  /status       - Show gateway status
  /help, /h     - Show this help message`
		m.addMessage("agent", help)
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
		return nil

	default:
		m.addMessage("agent", fmt.Sprintf("Unknown command: %s. Type /help for available commands.", command))
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
		return nil
	}
}

// completeCommand provides tab completion for commands.
func (m *Model) completeCommand(input string) string {
	commands := []string{
		"/quit", "/q",
		"/clear", "/c",
		"/skills", "/s",
		"/status",
		"/help", "/h",
	}

	for _, cmd := range commands {
		if strings.HasPrefix(cmd, input) {
			return cmd
		}
	}

	return input
}

// addMessage adds a message to the chat.
func (m *Model) addMessage(role, content string) {
	m.messages = append(m.messages, Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})
}

// renderMessages renders all messages as a string.
func (m *Model) renderMessages() string {
	var b strings.Builder

	for _, msg := range m.messages {
		switch msg.Role {
		case "user":
			b.WriteString(m.styles.UserMessage.Render("You: " + msg.Content))
		case "agent":
			b.WriteString(m.styles.AgentMessage.Render("Agent: " + msg.Content))
		}
		b.WriteString("\n\n")
	}

	return b.String()
}

// sendToAgent sends a message to the agent via RPC.
func (m *Model) sendToAgent(message string) tea.Cmd {
	return func() tea.Msg {
		if m.rpcClient == nil {
			return errorMsg("not connected to gateway")
		}

		response, err := m.rpcClient.ProcessMessage(m.sessionID, message)
		if err != nil {
			return errorMsg(err.Error())
		}

		return responseMsg(response)
	}
}

// listSkills lists available skills.
func (m *Model) listSkills() tea.Cmd {
	return func() tea.Msg {
		if m.rpcClient == nil {
			return errorMsg("not connected to gateway")
		}

		skills, err := m.rpcClient.ListSkills()
		if err != nil {
			return errorMsg(err.Error())
		}

		if len(skills) == 0 {
			return responseMsg("No skills available.")
		}

		var b strings.Builder
		b.WriteString("Available skills:\n")
		for _, skill := range skills {
			b.WriteString(fmt.Sprintf("  - %s: %s\n", skill.Name, skill.Description))
		}

		return responseMsg(b.String())
	}
}

// showStatus shows gateway status.
func (m *Model) showStatus() tea.Cmd {
	return func() tea.Msg {
		if m.rpcClient == nil {
			return errorMsg("not connected to gateway")
		}

		status, err := m.rpcClient.GetStatus()
		if err != nil {
			return errorMsg(err.Error())
		}

		msg := fmt.Sprintf(
			"Gateway Status:\n  Running: %v\n  PID: %d\n  Version: %s\n  Plugins: %d",
			status.Running, status.Pid, status.Version, status.PluginCount,
		)

		return responseMsg(msg)
	}
}

// sendMessageCmd creates a command that adds a message.
func (m *Model) sendMessageCmd(content string, isAgent bool) tea.Cmd {
	return func() tea.Msg {
		if isAgent {
			return responseMsg(content)
		}
		return nil
	}
}

// Message types for commands.
type responseMsg string
type errorMsg string
type statusMsg string

// generateSessionID generates a unique session ID.
func generateSessionID() string {
	return fmt.Sprintf("session-%d", time.Now().UnixNano())
}
