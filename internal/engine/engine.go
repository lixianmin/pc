package engine

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/lixianmin/logo"
	"github.com/lixianmin/pc/baml_client/types"
	"github.com/lixianmin/pc/internal/debug"
	"github.com/lixianmin/pc/internal/plugin"
	"github.com/lixianmin/pc/internal/skill"
	"github.com/lixianmin/pc/internal/task"
	pkgerr "github.com/lixianmin/pc/pkg/error"
	pkgtypes "github.com/lixianmin/pc/pkg/types"
)

type PluginCaller interface {
	ListPlugins() []*pkgtypes.Plugin
	CallPlugin(plugin *pkgtypes.Plugin, method string, params any) (any, error)
}

// Engine is the core engine implementation.
type Engine struct {
	pluginManager  *plugin.PluginManager
	mockCaller     PluginCaller
	sessions       map[string]*Session
	mu             sync.RWMutex
	llmPlugin      *pkgtypes.Plugin
	systemPrompt   string
	maxIterations  int
	toolTimeout    time.Duration
	promptRecorder *debug.PromptRecorder

	// Task management
	taskManager *task.Manager
	decomposer  *task.Decomposer
	taskEnabled bool

	// Skill management
	skillManager *skill.SkillManager

	llmClient *Client
}

// NewEngine creates a new core engine.
func NewEngine(pm *plugin.PluginManager) *Engine {
	return &Engine{
		pluginManager: pm,
		sessions:      make(map[string]*Session),
		maxIterations: 10,
		toolTimeout:   30 * time.Second,
		taskManager:   task.NewManager(""),
		decomposer:    task.NewDecomposer(),
		llmClient:     NewClient(),
	}
}

// NewEngineWithMock creates a new engine with a mock plugin caller for testing.
func NewEngineWithMock(caller PluginCaller) *Engine {
	return &Engine{
		mockCaller:    caller,
		sessions:      make(map[string]*Session),
		maxIterations: 10,
		toolTimeout:   30 * time.Second,
	}
}

// SetPromptRecorder sets the prompt recorder for debugging.
func (my *Engine) SetPromptRecorder(recorder *debug.PromptRecorder) {
	my.promptRecorder = recorder
}

// SetMaxIterations sets the maximum ReAct loop iterations.
func (my *Engine) SetMaxIterations(n int) {
	my.maxIterations = n
}

// SetToolTimeout sets the tool execution timeout.
func (my *Engine) SetToolTimeout(d time.Duration) {
	my.toolTimeout = d
}

// SetLLMPlugin sets the LLM plugin to use for generating responses.
func (my *Engine) SetLLMPlugin(p *pkgtypes.Plugin) {
	my.llmPlugin = p
}

// SetSystemPrompt sets the complete system prompt for LLM calls.
func (my *Engine) SetSystemPrompt(prompt string) {
	my.systemPrompt = prompt
}

// BuildSystemPrompt builds the dynamic system prompt with tools and skills.
func (my *Engine) BuildSystemPrompt() string {
	return my.buildDynamicSystemPrompt()
}

// SetTaskEnabled enables or disables task decomposition.
func (my *Engine) SetTaskEnabled(enabled bool) {
	my.taskEnabled = enabled
}

// GetTaskManager returns the task manager.
func (my *Engine) GetTaskManager() *task.Manager {
	return my.taskManager
}

// SetSkillDir sets the skill directory and loads skills.
func (my *Engine) SetSkillDir(dir string) error {
	my.skillManager = skill.NewSkillManager()
	_, err := my.skillManager.LoadSkills(dir)
	return err
}

// ListSkills returns all loaded skills.
func (my *Engine) ListSkills() []skill.Skill {
	if my.skillManager == nil {
		return nil
	}
	return my.skillManager.ListSkills()
}

// GetSkill returns a skill by name.
func (my *Engine) GetSkill(name string) *skill.Skill {
	return my.skillManager.GetSkill(name)
}

// ProcessMessage processes an incoming message with ReAct loop.
func (my *Engine) ProcessMessage(ctx context.Context, sessionId, message string) (string, error) {
	if sessionId == "" {
		return "", pkgerr.NewAppError("EngineProcess", "session ID cannot be empty")
	}

	if message == "" {
		return "", pkgerr.NewAppError("EngineProcess", "message cannot be empty")
	}

	my.mu.RLock()
	session, exists := my.sessions[sessionId]
	my.mu.RUnlock()

	if !exists {
		return "", pkgerr.NewAppErrorf("EngineProcess", "session not found: %s", sessionId)
	}

	logo.Info("[Session:", sessionId, "] User:", message)

	// Check for task decomposition trigger
	if my.taskEnabled && my.isGoalMessage(message) {
		response, err := my.handleGoalDecomposition(ctx, sessionId, message)
		if err != nil {
			return "", err
		}
		if response != "" {
			if err := session.AddMessage("assistant", response); err != nil {
				return "", pkgerr.WrapAppError("EngineProcess", "failed to save assistant response", err)
			}
			return response, nil
		}
	}

	if err := session.AddMessage("user", message); err != nil {
		return "", pkgerr.WrapAppError("EngineProcess", "failed to save user message", err)
	}

	var response string
	hasLLM := my.llmClient != nil || (my.pluginManager != nil && my.llmPlugin != nil)
	if hasLLM {
		logo.Info("[Session:", sessionId, "] Starting ReAct loop")
		resp, err := my.reactLoop(ctx, session, message)
		if err != nil {
			logo.Error("[Session:", sessionId, "] ReAct loop failed:", err)
			return "", pkgerr.WrapAppError("EngineProcess", "failed to process message", err)
		}
		response = resp
		logo.Info("[Session:", sessionId, "] ReAct loop completed, response length:", len(response))
	} else {
		response = fmt.Sprintf("Echo: %s", message)
		logo.Info("[Session:", sessionId, "] No LLM plugin, echoing")
	}

	if err := session.AddMessage("assistant", response); err != nil {
		return "", pkgerr.WrapAppError("EngineProcess", "failed to save assistant response", err)
	}

	logResp := response
	if len(logResp) > 100 {
		logResp = logResp[:97] + "..."
	}
	logo.Info("[Session:", sessionId, "] Assistant:", logResp)

	return response, nil
}

// reactLoop implements the ReAct (Reasoning + Acting) loop.
func (my *Engine) reactLoop(ctx context.Context, session *Session, initialMessage string) (string, error) {
	currentMessage := initialMessage

	for iteration := 0; iteration < my.maxIterations; iteration++ {
		logo.Info("[ReAct] Iteration", iteration+1, "/", my.maxIterations)

		result, err := my.callLLM(ctx, session, currentMessage)
		if err != nil {
			return "", pkgerr.WrapAppError("EngineReact", fmt.Sprintf("LLM call failed at iteration %d", iteration+1), err)
		}

		if result.IsChatResponse() {
			logo.Info("[ReAct] ChatResponse received, returning final response")
			return result.AsChatResponse().Content, nil
		}

		toolCall := my.convertToToolCall(result)
		if toolCall == nil {
			logo.Warn("[ReAct] Unknown response type, returning as-is")
			return "[Unknown response type]", nil
		}

		logo.Info("[ReAct] Tool call:", toolCall.Name)

		if err := session.AddMessage("assistant", fmt.Sprintf("[Tool: %s]", toolCall.Name)); err != nil {
			return "", pkgerr.WrapAppError("EngineReact", "failed to save assistant message", err)
		}

		toolCtx, cancel := context.WithTimeout(ctx, my.toolTimeout)
		defer cancel()

		var executor *ToolExecutor
		if my.mockCaller != nil {
			executor = NewToolExecutor(my.mockCaller)
		} else {
			executor = NewToolExecutor(my.pluginManager)
		}
		toolResult := executor.Execute(toolCtx, *toolCall)

		toolResultsMessage := FormatToolResult(toolResult)
		if err := session.AddMessage("system", toolResultsMessage); err != nil {
			return "", pkgerr.WrapAppError("EngineReact", "failed to save tool results", err)
		}

		currentMessage = toolResultsMessage
	}

	return "", pkgerr.NewAppErrorf("EngineReact", "exceeded maximum iterations (%d)", my.maxIterations)
}

func (my *Engine) convertToToolCall(result *ChatResult) *ToolCall {
	switch {
	case result.IsBashTool():
		tool := result.AsBashTool()
		return &ToolCall{Name: "bash", Params: map[string]interface{}{"command": tool.Command}}
	case result.IsReadTool():
		tool := result.AsReadTool()
		return &ToolCall{Name: "read", Params: map[string]interface{}{"file_path": tool.File_path}}
	case result.IsWriteTool():
		tool := result.AsWriteTool()
		return &ToolCall{Name: "write", Params: map[string]interface{}{"file_path": tool.File_path, "content": tool.Content}}
	case result.IsEditTool():
		tool := result.AsEditTool()
		return &ToolCall{Name: "edit", Params: map[string]interface{}{"file_path": tool.File_path, "old_string": tool.Old_string, "new_string": tool.New_string}}
	case result.IsWebSearchTool():
		tool := result.AsWebSearchTool()
		return &ToolCall{Name: "web_search", Params: map[string]interface{}{"query": tool.Query}}
	case result.IsWebFetchTool():
		tool := result.AsWebFetchTool()
		return &ToolCall{Name: "web_fetch", Params: map[string]interface{}{"url": tool.Url}}
	case result.IsUseSkill():
		tool := result.AsUseSkill()
		return &ToolCall{Name: "use_skill", Params: map[string]interface{}{"skill_name": tool.Skill_name, "input": tool.Input}}
	default:
		return nil
	}
}

// callLLM calls the LLM via BAML to generate a response.
func (my *Engine) callLLM(ctx context.Context, session *Session, message string) (*ChatResult, error) {
	startTime := time.Now()

	var systemPrompt = my.buildDynamicSystemPrompt()
	if my.systemPrompt != "" {
		systemPrompt = my.systemPrompt + "\n\n" + systemPrompt
	}

	history := session.GetMessages()
	var messages []types.Message
	for _, msg := range history {
		messages = append(messages, types.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	messages = append(messages, types.Message{
		Role:    "user",
		Content: message,
	})

	result, err := my.llmClient.Chat(ctx, messages, systemPrompt)
	if err != nil {
		logo.Error("[Engine.callLLM] Chat failed:", err)
		return nil, err
	}

	elapsed := time.Since(startTime)
	var totalChars int
	for _, m := range messages {
		totalChars += len(m.Content)
	}
	logo.Info("[Engine.callLLM] LLM call completed, chars=", totalChars, " time=", elapsed)

	return result, nil
}

// callLLMViaPlugin 使用插件调用 LLM（原有实现）
func (my *Engine) callLLMViaPlugin(ctx context.Context, session *Session, message string) (string, error) {
	startTime := time.Now()

	var history = session.GetMessages()

	systemPrompt := my.buildDynamicSystemPrompt()

	capacity := len(history) + 1
	if systemPrompt != "" {
		capacity++
	}
	messages := make([]map[string]string, 0, capacity)

	if systemPrompt != "" {
		messages = append(messages, map[string]string{
			"role":    "system",
			"content": systemPrompt,
		})
	}

	for _, msg := range history {
		messages = append(messages, map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}
	messages = append(messages, map[string]string{
		"role":    "user",
		"content": message,
	})

	var requestId string
	if my.promptRecorder != nil {
		requestId, _ = my.promptRecorder.Record(session.Id, messages)
	}

	params := map[string]any{
		"messages": messages,
	}

	result, err := my.pluginManager.CallPlugin(my.llmPlugin, "complete", params)
	if err != nil {
		logo.Error("[Engine.callLLMViaPlugin] Plugin call failed:", err)
		return "", err
	}

	resultMap, ok := result.(map[string]any)
	if !ok {
		logo.Error("[Engine.callLLMViaPlugin] Unexpected result type:", resultMap)
		return "", pkgerr.NewAppError("EngineLLM", "unexpected LLM response format")
	}

	content, ok := resultMap["content"].(string)
	if !ok {
		logo.Error("[Engine.callLLMViaPlugin] Missing content in result:", resultMap)
		return "", pkgerr.NewAppError("EngineLLM", "LLM response missing content")
	}

	elapsed := time.Since(startTime)
	var totalChars int
	for _, m := range messages {
		totalChars += len(m["content"])
	}

	if requestId != "" {
		logo.Info("[LLM] req=", requestId, " tokens=", totalChars, " time=", elapsed)
	} else {
		logo.Info("[Engine.callLLMViaPlugin] LLM call completed, response length:", len(content))
	}

	return content, nil
}

func (my *Engine) buildDynamicSystemPrompt() string {
	var parts []string

	if my.systemPrompt != "" {
		parts = append(parts, my.systemPrompt)
	}

	var availableTools = my.getAvailableTools()
	if len(availableTools) > 0 {
		parts = append(parts, "", my.buildToolGuide(availableTools))
	}

	availableSkills := my.ListSkills()
	if len(availableSkills) > 0 {
		parts = append(parts, "", my.buildSkillGuide(availableSkills))
	}

	return strings.Join(parts, "\n")
}

func (my *Engine) getAvailableTools() []ToolInfo {
	var tools []ToolInfo

	if my.pluginManager == nil && my.mockCaller == nil {
		return tools
	}

	if my.pluginManager != nil {
		executor := NewToolExecutor(my.pluginManager)
		toolNames := executor.ListAvailableTools()
		logo.Info("[Engine.getAvailableTools] Found", len(toolNames), "tools:", toolNames)

		for _, name := range toolNames {
			tools = append(tools, ToolInfo{
				Name:        name,
				Description: fmt.Sprintf("%s tool", name),
				Type:        "builtin",
			})
		}

		plugins := my.pluginManager.ListPlugins()
		for _, p := range plugins {
			if p.Type == pkgtypes.PluginTypeTool && p.Enabled {
				tools = append(tools, ToolInfo{
					Name:        p.Name,
					Description: fmt.Sprintf("%s tool", p.Name),
					Type:        string(p.Type),
				})
			}
		}
		return tools
	}

	if my.mockCaller != nil {
		plugins := my.mockCaller.ListPlugins()
		for _, p := range plugins {
			if p.Type == pkgtypes.PluginTypeTool && p.Enabled {
				tools = append(tools, ToolInfo{
					Name:        p.Name,
					Description: fmt.Sprintf("%s tool", p.Name),
					Type:        string(p.Type),
				})
			}
		}
	}

	return tools
}

func (my *Engine) buildToolGuide(tools []ToolInfo) string {
	var parts []string

	parts = append(parts, "## 工具使用指南")
	parts = append(parts, "")
	parts = append(parts, "当你需要获取外部信息或执行操作时，可以使用以下工具。")
	parts = append(parts, "")

	parts = append(parts, "### 可用工具")
	for _, t := range tools {
		parts = append(parts, fmt.Sprintf("- **%s**: %s", t.Name, t.Description))
	}

	parts = append(parts, "")
	parts = append(parts, "### 工具调用格式")
	parts = append(parts, "")
	parts = append(parts, "使用以下 XML 格式调用工具：")
	parts = append(parts, "")
	parts = append(parts, "<invoke>")
	parts = append(parts, "<name>工具名</name>")
	parts = append(parts, "<params>{\"参数\": \"值\"}</params>")
	parts = append(parts, "</invoke>")
	parts = append(parts, "")

	parts = append(parts, "### ReAct 工作流程")
	parts = append(parts, "1. 思考：分析用户需求")
	parts = append(parts, "2. 行动：调用工具")
	parts = append(parts, "3. 观察：接收工具结果")
	parts = append(parts, "4. 回复：基于结果回答用户")

	return strings.Join(parts, "\n")
}

func (my *Engine) buildSkillGuide(skills []skill.Skill) string {
	var parts []string

	parts = append(parts, "## 技能使用指南")
	parts = append(parts, "")
	parts = append(parts, "技能是由多个步骤组成的复合能力，当需要执行复杂任务时可以参考。")
	parts = append(parts, "")

	parts = append(parts, "### 可用技能")
	for _, s := range skills {
		description := s.Description
		if s.Description == "" {
			description = "复合任务"
		}
		parts = append(parts, fmt.Sprintf("- **%s**: %s", s.Name, description))
	}

	parts = append(parts, "")
	parts = append(parts, "### 使用方式")
	parts = append(parts, "当需要执行复杂任务时，可以参考技能中的步骤来指导你的操作。")

	return strings.Join(parts, "\n")
}

type ToolInfo struct {
	Name        string
	Description string
	Type        string
}

// FetchSession returns the session if exists, otherwise creates a new one.
func (my *Engine) FetchSession(sessionId string) *Session {
	if sessionId == "" {
		return nil
	}

	my.mu.Lock()
	defer my.mu.Unlock()

	session, exists := my.sessions[sessionId]
	if !exists {
		session = NewSession(sessionId)
		my.sessions[sessionId] = session
	}
	return session
}

// CloseSession closes a session.
func (my *Engine) CloseSession(sessionId string) error {
	my.mu.Lock()
	defer my.mu.Unlock()
	delete(my.sessions, sessionId)
	return nil
}

// Close closes the engine.
func (my *Engine) Close() error {
	my.mu.Lock()
	defer my.mu.Unlock()
	for key := range my.sessions {
		delete(my.sessions, key)
	}
	return nil
}

// isGoalMessage checks if the message is a goal decomposition request.
func (my *Engine) isGoalMessage(message string) bool {
	if strings.HasPrefix(message, "/task ") {
		return true
	}

	goalPrefixes := []string{"我想", "我要", "请帮我", "帮我"}
	for _, prefix := range goalPrefixes {
		if strings.HasPrefix(message, prefix) {
			return true
		}
	}

	return false
}

// handleGoalDecomposition handles goal decomposition and task creation.
func (my *Engine) handleGoalDecomposition(ctx context.Context, sessionId string, message string) (string, error) {
	var goal string
	var err error

	if strings.HasPrefix(message, "/task ") {
		goal, err = my.decomposer.ParseGoal(strings.TrimPrefix(message, "/task "))
	} else {
		goal, err = my.decomposer.ParseGoal(message)
	}

	if err != nil {
		return "", pkgerr.WrapAppError("EngineGoal", "failed to parse goal", err)
	}

	if goal == "" {
		return "", pkgerr.NewAppError("EngineGoal", "goal cannot be empty")
	}

	steps, err := my.decomposer.Decompose(goal)
	if err != nil {
		return "", pkgerr.WrapAppError("EngineGoal", "failed to decompose goal", err)
	}

	newTask, err := my.taskManager.AddTask(goal)
	if err != nil {
		return "", pkgerr.WrapAppError("EngineGoal", "failed to create task", err)
	}

	for _, step := range steps {
		if err := newTask.AddStep(step); err != nil {
			logo.Warn("Failed to add step:", step, err)
		}
	}

	logo.Info("[Session:", sessionId, "] Created task:", goal, "with", len(steps), "steps")

	var responseBuilder strings.Builder
	responseBuilder.WriteString("已为您创建任务：\n\n")
	responseBuilder.WriteString(fmt.Sprintf("**%s**\n\n", goal))
	responseBuilder.WriteString("步骤：\n")
	for i, step := range steps {
		responseBuilder.WriteString(fmt.Sprintf("%d. %s\n", i+1, step))
	}

	return responseBuilder.String(), nil
}
