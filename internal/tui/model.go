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
	rpcClient    *gateway.RpcClient
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
func NewModel(rpcClient *gateway.RpcClient) *Model {
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
func (my *Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		tea.EnterAltScreen,
		textarea.Blink,
	}

	// Load history
	if err := my.history.Load(); err == nil {
		// History loaded successfully
	}

	// Send initial greeting
	cmds = append(cmds, my.sendMessageCmd("Hello! I'm your PersonalClaw assistant. How can I help you today?", true))

	return tea.Batch(cmds...)
}

// Update handles messages and updates the model.
func (my *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		taCmd tea.Cmd
		vpCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		my.width = msg.Width
		my.height = msg.Height

		if !my.ready {
			// Initialize viewport
			my.viewport = viewport.New(msg.Width, msg.Height-6)
			my.viewport.SetContent(my.renderMessages())
			my.ready = true
		} else {
			my.viewport.Width = msg.Width
			my.viewport.Height = msg.Height - 6
			my.viewport.SetContent(my.renderMessages())
		}

		my.textarea.SetWidth(msg.Width - 4)

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			my.quitting = true
			return my, tea.Quit

		case tea.KeyEnter:
			if msg.Alt {
				// Alt+Enter for newline
				my.textarea.InsertString("\n")
			} else {
				// Send message
				input := strings.TrimSpace(my.textarea.Value())
				if input != "" {
					my.textarea.SetValue("")
					return my, my.handleInput(input)
				}
			}

		case tea.KeyUp:
			if my.textarea.Value() == "" {
				// Load previous history item
				if item := my.history.Previous(); item != "" {
					my.textarea.SetValue(item)
					my.textarea.CursorEnd()
				}
			} else if msg.Alt {
				// Alt+Up: scroll viewport up
				my.viewport.LineUp(1)
				my.userScrolled = true
			}

		case tea.KeyDown:
			if my.textarea.Value() == "" {
				// Load next history item
				if item := my.history.Next(); item != "" {
					my.textarea.SetValue(item)
					my.textarea.CursorEnd()
				} else {
					my.textarea.SetValue("")
				}
			} else if msg.Alt {
				// Alt+Down: scroll viewport down
				my.viewport.LineDown(1)
				if my.viewport.AtBottom() {
					my.userScrolled = false
				} else {
					my.userScrolled = true
				}
			}

		case tea.KeyPgUp:
			my.viewport.ViewUp()
			my.userScrolled = true

		case tea.KeyPgDown:
			my.viewport.ViewDown()
			if my.viewport.AtBottom() {
				my.userScrolled = false
			} else {
				my.userScrolled = true
			}

		case tea.KeyHome:
			my.viewport.GotoTop()
			my.userScrolled = true

		case tea.KeyEnd:
			my.viewport.GotoBottom()
			my.userScrolled = false

		case tea.KeyTab:
			// Tab completion
			input := my.textarea.Value()
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
				completed = my.completeReference(currentWord)
			} else if strings.HasPrefix(currentWord, "/") {
				// Command completion
				completed = my.completeCommand(currentWord)
			}

			if completed != currentWord {
				// Replace the word at cursor
				newInput := input[:wordStart] + completed
				if cursorPos < len(input) {
					newInput += input[cursorPos:]
				}
				my.textarea.SetValue(newInput + " ")
			}
		}

	case responseMsg:
		my.status = "Connected"
		my.addMessage("agent", string(msg))
		if my.ready {
			my.viewport.SetContent(my.renderMessages())
			// Only auto-scroll if user hasn't manually scrolled
			if !my.userScrolled {
				my.viewport.GotoBottom()
			}
		}

	case errorMsg:
		my.status = fmt.Sprintf("Error: %v", msg)
		my.addMessage("agent", fmt.Sprintf("Error: %v", msg))
		if my.ready {
			my.viewport.SetContent(my.renderMessages())
			if !my.userScrolled {
				my.viewport.GotoBottom()
			}
		}

	case streamStartMsg:
		my.isStreaming = true
		my.streamBuffer.Reset()
		my.status = "Streaming..."

	case streamChunkMsg:
		my.streamBuffer.WriteString(msg.content)
		if my.ready {
			lastIdx := len(my.messages) - 1
			if lastIdx >= 0 && my.messages[lastIdx].Role == "agent-streaming" {
				my.messages[lastIdx].Content = my.streamBuffer.String()
			} else {
				my.messages = append(my.messages, Message{
					Role:    "agent-streaming",
					Content: my.streamBuffer.String(),
				})
			}
			my.viewport.SetContent(my.renderMessages())
			if !my.userScrolled {
				my.viewport.GotoBottom()
			}
		}

		if msg.done {
			return my, my.processStreamFinalize()
		}

	case streamDoneMsg:
		my.isStreaming = false
		my.status = "Connected"
		if my.ready {
			lastIdx := len(my.messages) - 1
			if lastIdx >= 0 && my.messages[lastIdx].Role == "agent-streaming" {
				my.messages[lastIdx].Role = "agent"
				my.messages[lastIdx].Content = msg.fullContent
			}
			my.viewport.SetContent(my.renderMessages())
			if !my.userScrolled {
				my.viewport.GotoBottom()
			}
		}

	case statusMsg:
		my.status = string(msg)
	}

	my.textarea, taCmd = my.textarea.Update(msg)
	if my.ready {
		my.viewport, vpCmd = my.viewport.Update(msg)
	}

	return my, tea.Batch(taCmd, vpCmd)
}

// View renders the UI.
func (my *Model) View() string {
	if !my.ready {
		return "Loading..."
	}

	if my.quitting {
		return "Goodbye!\n"
	}

	// Header
	header := my.styles.Header.Render("PersonalClaw v0.1.0")

	// Status bar
	status := my.styles.StatusBar.Render(fmt.Sprintf("Status: %s | Session: %s", my.status, my.sessionID[:8]))

	// Input area
	prompt := my.styles.InputPrompt.Render("> ")
	input := my.textarea.View()

	// Render content area with scrollbar
	content := my.renderContentWithScrollbar()

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
func (my *Model) renderContentWithScrollbar() string {
	viewportContent := my.viewport.View()
	scrollbar := my.renderScrollbar()

	// Join viewport and scrollbar horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top, viewportContent, scrollbar)
}

// renderScrollbar renders the scrollbar based on current scroll position.
func (my *Model) renderScrollbar() string {
	// Only show scrollbar if content exceeds viewport height
	totalLines := my.viewport.TotalLineCount()
	visibleLines := my.viewport.Height

	if totalLines <= visibleLines {
		return ""
	}

	// Calculate scrollbar thumb position and size
	scrollPercent := float64(my.viewport.YOffset) / float64(totalLines-visibleLines)
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
			sb.WriteString(my.styles.ScrollbarThumb.Render("█"))
		} else {
			sb.WriteString(my.styles.ScrollbarTrack.Render("░"))
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
func (my *Model) handleInput(input string) tea.Cmd {
	// Reset user scrolled state when user sends a new message
	my.userScrolled = false

	// Add to history
	my.history.Add(input)
	my.history.Save()

	// Check for commands
	if strings.HasPrefix(input, "/") {
		return my.handleCommand(input)
	}

	// Parse input for @ references
	parsed, err := ParseInput(input, my)
	if err != nil {
		my.addMessage("agent", fmt.Sprintf("Error parsing input: %v", err))
		my.viewport.SetContent(my.renderMessages())
		my.viewport.GotoBottom()
		return nil
	}

	// Build full message with context
	fullMessage := parsed.CleanText
	if context := parsed.BuildContext(); context != "" {
		fullMessage += context
	}

	// Display original input to user
	my.addMessage("user", input)
	my.viewport.SetContent(my.renderMessages())
	my.viewport.GotoBottom()

	// Send full message (with context) to agent
	return my.sendToAgent(fullMessage)
}

// handleCommand handles slash commands.
func (my *Model) handleCommand(cmd string) tea.Cmd {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return nil
	}

	command := parts[0]

	switch command {
	case "/quit", "/q", "/exit":
		my.quitting = true
		return tea.Quit

	case "/clear", "/c":
		my.messages = make([]Message, 0)
		my.viewport.SetContent(my.renderMessages())
		return nil

	case "/skills", "/s":
		return my.listSkills()

	case "/status":
		return my.showStatus()

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
		my.addMessage("agent", help)
		my.viewport.SetContent(my.renderMessages())
		my.viewport.GotoBottom()
		return nil

	case "/task":
		return my.handleTaskCommand(parts)

	default:
		my.addMessage("agent", fmt.Sprintf("Unknown command: %s. Type /help for available commands.", command))
		my.viewport.SetContent(my.renderMessages())
		my.viewport.GotoBottom()
		return nil
	}
}

// completeCommand provides tab completion for commands.
func (my *Model) completeCommand(input string) string {
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
func (my *Model) completeReference(input string) string {
	if !strings.HasPrefix(input, "@") {
		return input
	}

	prefix := input[1:] // Remove @

	// Try skill completion first
	if skills := my.getMatchingSkills(prefix); len(skills) > 0 {
		return "@" + skills[0]
	}

	// Try file path completion
	if files := my.getMatchingFiles(prefix); len(files) > 0 {
		return "@" + files[0]
	}

	return input
}

// getMatchingSkills returns skills that match the given prefix.
func (my *Model) getMatchingSkills(prefix string) []string {
	var matches []string

	// Load skills from cache or RPC
	skills := my.loadSkillsCache()

	for _, skill := range skills {
		if strings.HasPrefix(skill.Name, prefix) {
			matches = append(matches, skill.Name)
		}
	}

	return matches
}

// getMatchingFiles returns files that match the given prefix.
func (my *Model) getMatchingFiles(prefix string) []string {
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
func (my *Model) loadSkillsCache() []SkillInfo {
	if my.skillCache != nil {
		return my.skillCache
	}

	// Try to load from RPC if available
	if my.rpcClient != nil {
		skills, err := my.rpcClient.ListSkills()
		if err == nil {
			var cache []SkillInfo
			for _, s := range skills {
				cache = append(cache, SkillInfo{
					Name:        s.Name,
					Description: s.Description,
				})
			}
			my.skillCache = cache
			return cache
		}
	}

	// Return empty cache
	my.skillCache = []SkillInfo{}
	return my.skillCache
}

// GetSkill implements SkillProvider interface.
func (my *Model) GetSkill(name string) (SkillInfo, error) {
	skills := my.loadSkillsCache()
	for _, skill := range skills {
		if skill.Name == name {
			return skill, nil
		}
	}
	return SkillInfo{}, os.ErrNotExist
}

// ListSkills implements SkillProvider interface.
func (my *Model) ListSkills() []SkillInfo {
	return my.loadSkillsCache()
}

// addMessage adds a message to the chat.
func (my *Model) addMessage(role, content string) {
	my.messages = append(my.messages, Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})
}

// renderMessages renders all messages as a string with proper wrapping.
func (my *Model) renderMessages() string {
	var b strings.Builder

	// Calculate available width for message content
	// Subtract padding (2 for left padding) and some margin
	availableWidth := my.viewport.Width - 4
	if availableWidth < 20 {
		availableWidth = 20 // Minimum width
	}

	// Create wrapping styles based on available width
	userStyle := my.styles.UserMessage.Width(availableWidth)
	agentStyle := my.styles.AgentMessage.Width(availableWidth)

	for _, msg := range my.messages {
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
func (my *Model) sendToAgent(message string) tea.Cmd {
	return func() tea.Msg {
		if my.rpcClient == nil {
			return errorMsg("not connected to gateway")
		}

		response, err := my.rpcClient.ProcessMessage(my.sessionID, message)
		if err != nil {
			return errorMsg(err.Error())
		}

		return responseMsg(response)
	}
}

// listSkills lists available skills.
func (my *Model) listSkills() tea.Cmd {
	return func() tea.Msg {
		if my.rpcClient == nil {
			return errorMsg("not connected to gateway")
		}

		skills, err := my.rpcClient.ListSkills()
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
func (my *Model) showStatus() tea.Cmd {
	return func() tea.Msg {
		if my.rpcClient == nil {
			return errorMsg("not connected to gateway")
		}

		status, err := my.rpcClient.GetStatus()
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
func (my *Model) sendMessageCmd(content string, isAgent bool) tea.Cmd {
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

func (my *Model) sendToAgentStream(message string) tea.Cmd {
	return func() tea.Msg {
		if my.rpcClient == nil {
			return errorMsg("not connected to gateway")
		}

		return streamStartMsg{}
	}
}

func generateSessionID() string {
	return fmt.Sprintf("session-%d", time.Now().UnixNano())
}

func (my *Model) processStreamFinalize() tea.Cmd {
	return func() tea.Msg {
		return streamDoneMsg{fullContent: my.streamBuffer.String()}
	}
}

func (my *Model) processStreamChunks(sessionID, message string) tea.Cmd {
	return func() tea.Msg {
		if my.rpcClient == nil {
			return errorMsg("not connected to gateway")
		}

		chunks, err := my.rpcClient.ProcessMessageStream(sessionID, message)
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
func (my *Model) handleTaskCommand(parts []string) tea.Cmd {
	if len(parts) < 2 {
		return func() tea.Msg {
			return responseMsg("Usage: /task <command> [args]\nCommands: list, add, complete, delete")
		}
	}

	subcommand := parts[1]

	switch subcommand {
	case "list":
		return my.listTasks()
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
		return my.addTask(title, description)
	case "complete":
		if len(parts) < 3 {
			return func() tea.Msg {
				return responseMsg("Usage: /task complete <task-id>")
			}
		}
		return my.completeTask(parts[2])
	case "delete":
		if len(parts) < 3 {
			return func() tea.Msg {
				return responseMsg("Usage: /task delete <task-id>")
			}
		}
		return my.deleteTask(parts[2])
	default:
		return func() tea.Msg {
			return responseMsg(fmt.Sprintf("Unknown task command: %s. Available: list, add, complete, delete", subcommand))
		}
	}
}

// listTasks lists all tasks
func (my *Model) listTasks() tea.Cmd {
	return func() tea.Msg {
		if my.rpcClient == nil {
			return errorMsg("not connected to gateway")
		}

		tasks, err := my.rpcClient.ListTasks("")
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
func (my *Model) addTask(title, description string) tea.Cmd {
	return func() tea.Msg {
		if my.rpcClient == nil {
			return errorMsg("not connected to gateway")
		}

		taskID, err := my.rpcClient.AddTask(title, description)
		if err != nil {
			return errorMsg(err.Error())
		}

		return responseMsg(fmt.Sprintf("Task created: %s", taskID))
	}
}

// completeTask marks a task as completed
func (my *Model) completeTask(taskID string) tea.Cmd {
	return func() tea.Msg {
		if my.rpcClient == nil {
			return errorMsg("not connected to gateway")
		}

		err := my.rpcClient.CompleteTask(taskID)
		if err != nil {
			return errorMsg(err.Error())
		}

		return responseMsg(fmt.Sprintf("Task %s completed.", taskID))
	}
}

// deleteTask deletes a task
func (my *Model) deleteTask(taskID string) tea.Cmd {
	return func() tea.Msg {
		if my.rpcClient == nil {
			return errorMsg("not connected to gateway")
		}

		err := my.rpcClient.DeleteTask(taskID)
		if err != nil {
			return errorMsg(err.Error())
		}

		return responseMsg(fmt.Sprintf("Task %s deleted.", taskID))
	}
}

// AddMessage adds a message to the chat (public for testing).
func (my *Model) AddMessage(role, content string) {
	my.addMessage(role, content)
}

// RenderMessages renders all messages as a string (public for testing).
func (my *Model) RenderMessages() string {
	return my.renderMessages()
}

// SetViewportSize sets the viewport dimensions (public for testing).
func (my *Model) SetViewportSize(width, height int) {
	my.viewport.Width = width
	my.viewport.Height = height
}
