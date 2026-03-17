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
	Role      string // "user" or "agent"
	Content   string
	Timestamp time.Time
}

// Model represents the TUI application state.
type Model struct {
	// UI components
	viewport viewport.Model
	textarea textarea.Model

	// State
	messages     []Message
	history      *History
	rpcClient    *gateway.RPCClient
	sessionID    string
	status       string
	width        int
	height       int
	quitting     bool
	ready        bool
	skillCache   []SkillInfo
	userScrolled bool

	// Streaming state
	isStreaming  bool
	streamBuffer strings.Builder

	// Styles
	styles *Styles
}

// Styles holds the UI styles.
type Styles struct {
	Header         lipgloss.Style
	UserMessage    lipgloss.Style
	AgentMessage   lipgloss.Style
	StatusBar      lipgloss.Style
	InputPrompt    lipgloss.Style
	Error          lipgloss.Style
	ScrollbarTrack lipgloss.Style
	ScrollbarThumb lipgloss.Style
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

		ScrollbarTrack: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#333333")),

		ScrollbarThumb: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")),
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
			} else if msg.Alt {
				// Alt+Up: scroll viewport up
				m.viewport.LineUp(1)
				m.userScrolled = true
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
			} else if msg.Alt {
				// Alt+Down: scroll viewport down
				m.viewport.LineDown(1)
				if m.viewport.AtBottom() {
					m.userScrolled = false
				} else {
					m.userScrolled = true
				}
			}

		case tea.KeyPgUp:
			m.viewport.ViewUp()
			m.userScrolled = true

		case tea.KeyPgDown:
			m.viewport.ViewDown()
			if m.viewport.AtBottom() {
				m.userScrolled = false
			} else {
				m.userScrolled = true
			}

		case tea.KeyHome:
			m.viewport.GotoTop()
			m.userScrolled = true

		case tea.KeyEnd:
			m.viewport.GotoBottom()
			m.userScrolled = false

		case tea.KeyTab:
			// Tab completion
			input := m.textarea.Value()
			cursorPos := len(input) // Use end of input as cursor position for simplicity

			// Find the word at cursor position
			wordStart := cursorPos
			for wordStart > 0 && input[wordStart-1] != ' ' {
				wordStart--
			}
			currentWord := input[wordStart:cursorPos]

			var completed string
			if strings.HasPrefix(currentWord, "@") {
				// Skill or file completion
				completed = m.completeReference(currentWord)
			} else if strings.HasPrefix(currentWord, "/") {
				// Command completion
				completed = m.completeCommand(currentWord)
			}

			if completed != currentWord {
				// Replace the word at cursor
				newInput := input[:wordStart] + completed
				if cursorPos < len(input) {
					newInput += input[cursorPos:]
				}
				m.textarea.SetValue(newInput + " ")
			}
		}

	case responseMsg:
		m.status = "Connected"
		m.addMessage("agent", string(msg))
		if m.ready {
			m.viewport.SetContent(m.renderMessages())
			// Only auto-scroll if user hasn't manually scrolled
			if !m.userScrolled {
				m.viewport.GotoBottom()
			}
		}

	case errorMsg:
		m.status = fmt.Sprintf("Error: %v", msg)
		m.addMessage("agent", fmt.Sprintf("Error: %v", msg))
		if m.ready {
			m.viewport.SetContent(m.renderMessages())
			if !m.userScrolled {
				m.viewport.GotoBottom()
			}
		}

	case streamStartMsg:
		m.isStreaming = true
		m.streamBuffer.Reset()
		m.status = "Streaming..."

	case streamChunkMsg:
		m.streamBuffer.WriteString(msg.content)
		if m.ready {
			lastIdx := len(m.messages) - 1
			if lastIdx >= 0 && m.messages[lastIdx].Role == "agent-streaming" {
				m.messages[lastIdx].Content = m.streamBuffer.String()
			} else {
				m.messages = append(m.messages, Message{
					Role:    "agent-streaming",
					Content: m.streamBuffer.String(),
				})
			}
			m.viewport.SetContent(m.renderMessages())
			if !m.userScrolled {
				m.viewport.GotoBottom()
			}
		}

		if msg.done {
			return m, m.processStreamFinalize()
		}

	case streamDoneMsg:
		m.isStreaming = false
		m.status = "Connected"
		if m.ready {
			lastIdx := len(m.messages) - 1
			if lastIdx >= 0 && m.messages[lastIdx].Role == "agent-streaming" {
				m.messages[lastIdx].Role = "agent"
				m.messages[lastIdx].Content = msg.fullContent
			}
			m.viewport.SetContent(m.renderMessages())
			if !m.userScrolled {
				m.viewport.GotoBottom()
			}
		}

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

	// Render content area with scrollbar
	content := m.renderContentWithScrollbar()

	// Combine all parts
	return fmt.Sprintf(
		"%s\n%s\n%s\n%s%s",
		header,
		content,
		status,
		prompt,
		input,
	)
}

// renderContentWithScrollbar renders the viewport content with a scrollbar.
func (m *Model) renderContentWithScrollbar() string {
	viewportContent := m.viewport.View()
	scrollbar := m.renderScrollbar()

	// Join viewport and scrollbar horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top, viewportContent, scrollbar)
}

// renderScrollbar renders the scrollbar based on current scroll position.
func (m *Model) renderScrollbar() string {
	// Only show scrollbar if content exceeds viewport height
	totalLines := m.viewport.TotalLineCount()
	visibleLines := m.viewport.Height

	if totalLines <= visibleLines {
		return ""
	}

	// Calculate scrollbar thumb position and size
	scrollPercent := float64(m.viewport.YOffset) / float64(totalLines-visibleLines)
	thumbHeight := max(1, visibleLines*visibleLines/totalLines)
	if thumbHeight < 1 {
		thumbHeight = 1
	}

	// Calculate thumb position
	trackHeight := visibleLines
	thumbPos := int(scrollPercent * float64(trackHeight-thumbHeight))

	// Build scrollbar string
	var sb strings.Builder
	for i := 0; i < trackHeight; i++ {
		if i >= thumbPos && i < thumbPos+thumbHeight {
			sb.WriteString(m.styles.ScrollbarThumb.Render("█"))
		} else {
			sb.WriteString(m.styles.ScrollbarTrack.Render("░"))
		}
		if i < trackHeight-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// handleInput processes user input.
func (m *Model) handleInput(input string) tea.Cmd {
	// Reset user scrolled state when user sends a new message
	m.userScrolled = false

	// Add to history
	m.history.Add(input)
	m.history.Save()

	// Check for commands
	if strings.HasPrefix(input, "/") {
		return m.handleCommand(input)
	}

	// Parse input for @ references
	parsed, err := ParseInput(input, m)
	if err != nil {
		m.addMessage("agent", fmt.Sprintf("Error parsing input: %v", err))
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
		return nil
	}

	// Build full message with context
	fullMessage := parsed.CleanText
	if context := parsed.BuildContext(); context != "" {
		fullMessage += context
	}

	// Display original input to user
	m.addMessage("user", input)
	m.viewport.SetContent(m.renderMessages())
	m.viewport.GotoBottom()

	// Send full message (with context) to agent
	return m.sendToAgent(fullMessage)
}

// handleCommand handles slash commands.
func (m *Model) handleCommand(cmd string) tea.Cmd {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return nil
	}

	command := parts[0]

	switch command {
	case "/quit", "/q", "/exit":
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
  /quit, /q, /exit - Exit the TUI
  /clear, /c    - Clear the screen
  /skills, /s   - List available skills
  /status       - Show gateway status
  /task list    - List tasks
  /task add <title>  - Add a new task
  /task complete <id> - Complete a task
  /help, /h     - Show this help message`
		m.addMessage("agent", help)
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
		return nil

	case "/task":
		return m.handleTaskCommand(parts)

	default:
		m.addMessage("agent", fmt.Sprintf("Unknown command: %s. Type /help for available commands.", command))
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
		return nil
	}
}

// completeCommand provides tab completion for commands.
func (m *Model) completeCommand(input string) string {
	// Handle task subcommands
	if strings.HasPrefix(input, "/task ") {
		taskCommands := []string{"list", "add", "complete", "delete"}
		for _, cmd := range taskCommands {
			fullCmd := "/task " + cmd
			if strings.HasPrefix(fullCmd, input) {
				return fullCmd
			}
		}
		return input
	}

	commands := []string{
		"/quit", "/q", "/exit",
		"/clear", "/c",
		"/skills", "/s",
		"/status",
		"/task",
		"/help", "/h",
	}

	for _, cmd := range commands {
		if strings.HasPrefix(cmd, input) {
			return cmd
		}
	}

	return input
}

// completeReference provides tab completion for @ references (skills and files).
func (m *Model) completeReference(input string) string {
	if !strings.HasPrefix(input, "@") {
		return input
	}

	prefix := input[1:] // Remove @

	// Try skill completion first
	if skills := m.getMatchingSkills(prefix); len(skills) > 0 {
		return "@" + skills[0]
	}

	// Try file path completion
	if files := m.getMatchingFiles(prefix); len(files) > 0 {
		return "@" + files[0]
	}

	return input
}

// getMatchingSkills returns skills that match the given prefix.
func (m *Model) getMatchingSkills(prefix string) []string {
	var matches []string

	// Load skills from cache or RPC
	skills := m.loadSkillsCache()

	for _, skill := range skills {
		if strings.HasPrefix(skill.Name, prefix) {
			matches = append(matches, skill.Name)
		}
	}

	return matches
}

// getMatchingFiles returns files that match the given prefix.
func (m *Model) getMatchingFiles(prefix string) []string {
	var matches []string

	// Expand ~ to home directory
	if strings.HasPrefix(prefix, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			prefix = filepath.Join(home, prefix[1:])
		}
	}

	// Get directory and file prefix
	dir := filepath.Dir(prefix)
	filePrefix := filepath.Base(prefix)

	if dir == "" || dir == "." {
		dir = "."
	}

	// Read directory entries
	entries, err := os.ReadDir(dir)
	if err != nil {
		return matches
	}

	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, filePrefix) {
			fullPath := filepath.Join(dir, name)
			// If it's a directory, add trailing slash
			if entry.IsDir() {
				fullPath += "/"
			}
			// Convert back to ~ if it was expanded
			if home, err := os.UserHomeDir(); err == nil && strings.HasPrefix(fullPath, home) {
				fullPath = "~" + fullPath[len(home):]
			}
			matches = append(matches, fullPath)
		}
	}

	return matches
}

// loadSkillsCache loads skills into cache for completion.
func (m *Model) loadSkillsCache() []SkillInfo {
	if m.skillCache != nil {
		return m.skillCache
	}

	// Try to load from RPC if available
	if m.rpcClient != nil {
		skills, err := m.rpcClient.ListSkills()
		if err == nil {
			var cache []SkillInfo
			for _, s := range skills {
				cache = append(cache, SkillInfo{
					Name:        s.Name,
					Description: s.Description,
				})
			}
			m.skillCache = cache
			return cache
		}
	}

	// Return empty cache
	m.skillCache = []SkillInfo{}
	return m.skillCache
}

// GetSkill implements SkillProvider interface.
func (m *Model) GetSkill(name string) (SkillInfo, error) {
	skills := m.loadSkillsCache()
	for _, skill := range skills {
		if skill.Name == name {
			return skill, nil
		}
	}
	return SkillInfo{}, os.ErrNotExist
}

// ListSkills implements SkillProvider interface.
func (m *Model) ListSkills() []SkillInfo {
	return m.loadSkillsCache()
}

// addMessage adds a message to the chat.
func (m *Model) addMessage(role, content string) {
	m.messages = append(m.messages, Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})
}

// renderMessages renders all messages as a string with proper wrapping.
func (m *Model) renderMessages() string {
	var b strings.Builder

	// Calculate available width for message content
	// Subtract padding (2 for left padding) and some margin
	availableWidth := m.viewport.Width - 4
	if availableWidth < 20 {
		availableWidth = 20 // Minimum width
	}

	// Create wrapping styles based on available width
	userStyle := m.styles.UserMessage.Width(availableWidth)
	agentStyle := m.styles.AgentMessage.Width(availableWidth)

	for _, msg := range m.messages {
		switch msg.Role {
		case "user":
			b.WriteString(userStyle.Render("You: " + msg.Content))
		case "agent":
			b.WriteString(agentStyle.Render("Agent: " + msg.Content))
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
type streamChunkMsg struct {
	content string
	done    bool
}
type streamStartMsg struct{}
type streamDoneMsg struct {
	fullContent string
}

func (m *Model) sendToAgentStream(message string) tea.Cmd {
	return func() tea.Msg {
		if m.rpcClient == nil {
			return errorMsg("not connected to gateway")
		}

		return streamStartMsg{}
	}
}

func generateSessionID() string {
	return fmt.Sprintf("session-%d", time.Now().UnixNano())
}

func (m *Model) processStreamFinalize() tea.Cmd {
	return func() tea.Msg {
		return streamDoneMsg{fullContent: m.streamBuffer.String()}
	}
}

func (m *Model) processStreamChunks(sessionID, message string) tea.Cmd {
	return func() tea.Msg {
		if m.rpcClient == nil {
			return errorMsg("not connected to gateway")
		}

		chunks, err := m.rpcClient.ProcessMessageStream(sessionID, message)
		if err != nil {
			return errorMsg(err.Error())
		}

		var fullContent strings.Builder
		for _, chunk := range chunks {
			if chunk.Error != "" {
				return errorMsg(chunk.Error)
			}
			fullContent.WriteString(chunk.Content)
		}

		return streamDoneMsg{fullContent: fullContent.String()}
	}
}

// handleTaskCommand handles /task subcommands
func (m *Model) handleTaskCommand(parts []string) tea.Cmd {
	if len(parts) < 2 {
		return func() tea.Msg {
			return responseMsg("Usage: /task <command> [args]\nCommands: list, add, complete, delete")
		}
	}

	subcommand := parts[1]

	switch subcommand {
	case "list":
		return m.listTasks()
	case "add":
		if len(parts) < 3 {
			return func() tea.Msg {
				return responseMsg("Usage: /task add <title> [--description <desc>]")
			}
		}
		// Join remaining parts as title (until --description if present)
		var title string
		var description string
		for i := 2; i < len(parts); i++ {
			if parts[i] == "--description" && i+1 < len(parts) {
				description = parts[i+1]
				break
			}
			if title != "" {
				title += " "
			}
			title += parts[i]
		}
		return m.addTask(title, description)
	case "complete":
		if len(parts) < 3 {
			return func() tea.Msg {
				return responseMsg("Usage: /task complete <task-id>")
			}
		}
		return m.completeTask(parts[2])
	case "delete":
		if len(parts) < 3 {
			return func() tea.Msg {
				return responseMsg("Usage: /task delete <task-id>")
			}
		}
		return m.deleteTask(parts[2])
	default:
		return func() tea.Msg {
			return responseMsg(fmt.Sprintf("Unknown task command: %s. Available: list, add, complete, delete", subcommand))
		}
	}
}

// listTasks lists all tasks
func (m *Model) listTasks() tea.Cmd {
	return func() tea.Msg {
		if m.rpcClient == nil {
			return errorMsg("not connected to gateway")
		}

		tasks, err := m.rpcClient.ListTasks("")
		if err != nil {
			return errorMsg(err.Error())
		}

		if len(tasks) == 0 {
			return responseMsg("No tasks found.")
		}

		var b strings.Builder
		b.WriteString("Tasks:\n")
		for _, task := range tasks {
			status := "⏳"
			if task.Status == "completed" {
				status = "✅"
			}
			b.WriteString(fmt.Sprintf("  %s %s: %s\n", status, task.ID, task.Title))
		}

		return responseMsg(b.String())
	}
}

// addTask adds a new task
func (m *Model) addTask(title, description string) tea.Cmd {
	return func() tea.Msg {
		if m.rpcClient == nil {
			return errorMsg("not connected to gateway")
		}

		taskID, err := m.rpcClient.AddTask(title, description)
		if err != nil {
			return errorMsg(err.Error())
		}

		return responseMsg(fmt.Sprintf("Task created: %s", taskID))
	}
}

// completeTask marks a task as completed
func (m *Model) completeTask(taskID string) tea.Cmd {
	return func() tea.Msg {
		if m.rpcClient == nil {
			return errorMsg("not connected to gateway")
		}

		err := m.rpcClient.CompleteTask(taskID)
		if err != nil {
			return errorMsg(err.Error())
		}

		return responseMsg(fmt.Sprintf("Task %s completed.", taskID))
	}
}

// deleteTask deletes a task
func (m *Model) deleteTask(taskID string) tea.Cmd {
	return func() tea.Msg {
		if m.rpcClient == nil {
			return errorMsg("not connected to gateway")
		}

		err := m.rpcClient.DeleteTask(taskID)
		if err != nil {
			return errorMsg(err.Error())
		}

		return responseMsg(fmt.Sprintf("Task %s deleted.", taskID))
	}
}

// AddMessage adds a message to the chat (public for testing).
func (m *Model) AddMessage(role, content string) {
	m.addMessage(role, content)
}

// RenderMessages renders all messages as a string (public for testing).
func (m *Model) RenderMessages() string {
	return m.renderMessages()
}

// SetViewportSize sets the viewport dimensions (public for testing).
func (m *Model) SetViewportSize(width, height int) {
	m.viewport.Width = width
	m.viewport.Height = height
}
