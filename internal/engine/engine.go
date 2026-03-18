package engine

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lixianmin/logo"
	"github.com/lixianmin/pc/internal/plugin"
	"github.com/lixianmin/pc/pkg/types"
)

type PluginCaller interface {
	ListPlugins() []*types.Plugin
	CallPlugin(plugin *types.Plugin, method string, params any) (any, error)
}

type llmCallback func(ctx context.Context, session *Session, systemPrompt string) (string, error)

// Engine is the core engine implementation.
type Engine struct {
	pluginManager *plugin.PluginManager
	mockCaller    PluginCaller
	sessions      map[string]*Session
	llmPlugin     *types.Plugin
	systemPrompt  string
	maxIterations int
	toolTimeout   time.Duration
	llmCallback   llmCallback
}

// NewEngine creates a new core engine.
func NewEngine(pm *plugin.PluginManager) *Engine {
	return &Engine{
		pluginManager: pm,
		sessions:      make(map[string]*Session),
		maxIterations: 10,
		toolTimeout:   30 * time.Second,
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

// SetLLMCallback sets a custom LLM callback for testing.
func (my *Engine) SetLLMCallback(cb llmCallback) {
	my.llmCallback = cb
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
func (my *Engine) SetLLMPlugin(p *types.Plugin) {
	my.llmPlugin = p
}

// SetSystemPrompt sets the complete system prompt for LLM calls.
func (my *Engine) SetSystemPrompt(prompt string) {
	my.systemPrompt = prompt
}

// ProcessMessage processes an incoming message with ReAct loop.
func (my *Engine) ProcessMessage(ctx context.Context, sessionId, message string) (string, error) {
	if sessionId == "" {
		return "", fmt.Errorf("session ID cannot be empty")
	}

	if message == "" {
		return "", fmt.Errorf("message cannot be empty")
	}

	session, exists := my.sessions[sessionId]
	if !exists {
		return "", fmt.Errorf("session not found: %s", sessionId)
	}

	logo.Info("[Session:", sessionId, "] User:", message)

	if err := session.AddMessage("user", message); err != nil {
		return "", fmt.Errorf("failed to save user message: %w", err)
	}

	var response string
	hasLLM := (my.pluginManager != nil && my.llmPlugin != nil) || (my.llmCallback != nil)
	if hasLLM {
		logo.Info("[Session:", sessionId, "] Starting ReAct loop")
		resp, err := my.reactLoop(ctx, session, message)
		if err != nil {
			logo.Error("[Session:", sessionId, "] ReAct loop failed:", err)
			return "", fmt.Errorf("failed to process message: %w", err)
		}
		response = resp
		logo.Info("[Session:", sessionId, "] ReAct loop completed, response length:", len(response))
	} else {
		response = fmt.Sprintf("Echo: %s", message)
		logo.Info("[Session:", sessionId, "] No LLM plugin, echoing")
	}

	if err := session.AddMessage("assistant", response); err != nil {
		return "", fmt.Errorf("failed to save assistant response: %w", err)
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

		response, err := my.callLLM(ctx, session, currentMessage)
		if err != nil {
			return "", fmt.Errorf("LLM call failed at iteration %d: %w", iteration+1, err)
		}

		toolCalls, err := ParseToolCalls(response)
		if err != nil {
			logo.Warn("[ReAct] Failed to parse tool calls, returning response:", err)
			return response, nil
		}

		if len(toolCalls) == 0 {
			logo.Info("[ReAct] No tool calls found, returning final response")
			return response, nil
		}

		logo.Info("[ReAct] Found", len(toolCalls), "tool call(s)")

		if err := session.AddMessage("assistant", response); err != nil {
			return "", fmt.Errorf("failed to save assistant message: %w", err)
		}

		toolCtx, cancel := context.WithTimeout(ctx, my.toolTimeout)
		defer cancel()

		var executor *ToolExecutor
		if my.mockCaller != nil {
			executor = NewToolExecutor(my.mockCaller)
		} else {
			executor = NewToolExecutor(my.pluginManager)
		}
		results := executor.ExecuteMultiple(toolCtx, toolCalls)

		toolResultsMessage := FormatToolResults(results)
		if err := session.AddMessage("system", toolResultsMessage); err != nil {
			return "", fmt.Errorf("failed to save tool results: %w", err)
		}

		currentMessage = toolResultsMessage
	}

	return "", fmt.Errorf("exceeded maximum iterations (%d)", my.maxIterations)
}

// callLLM calls the LLM plugin to generate a response.
func (my *Engine) callLLM(ctx context.Context, session *Session, message string) (string, error) {
	logo.Info("[Engine.callLLM] Starting LLM call, message length:", len(message))

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

	if my.llmCallback != nil {
		return my.llmCallback(ctx, session, systemPrompt)
	}

	params := map[string]any{
		"messages": messages,
	}

	logo.Info("[Engine.callLLM] Calling plugin manager with", len(messages), "messages")

	result, err := my.pluginManager.CallPlugin(my.llmPlugin, "complete", params)
	if err != nil {
		logo.Error("[Engine.callLLM] Plugin call failed:", err)
		return "", err
	}

	resultMap, ok := result.(map[string]any)
	if !ok {
		logo.Error("[Engine.callLLM] Unexpected result type:", resultMap)
		return "", fmt.Errorf("unexpected LLM response format")
	}

	content, ok := resultMap["content"].(string)
	if !ok {
		logo.Error("[Engine.callLLM] Missing content in result:", resultMap)
		return "", fmt.Errorf("LLM response missing content")
	}

	logo.Info("[Engine.callLLM] LLM call completed, response length:", len(content))
	return content, nil
}

func (my *Engine) buildDynamicSystemPrompt() string {
	var parts []string

	if my.systemPrompt != "" {
		parts = append(parts, my.systemPrompt)
	}

	availableTools := my.getAvailableTools()
	if len(availableTools) > 0 {
		parts = append(parts, "", my.buildToolGuide(availableTools))
	}

	return strings.Join(parts, "\n")
}

func (my *Engine) getAvailableTools() []ToolInfo {
	if my.mockCaller != nil {
		plugins := my.mockCaller.ListPlugins()
		var tools []ToolInfo
		for _, p := range plugins {
			if p.Type == types.PluginTypeTool && p.Enabled {
				tools = append(tools, ToolInfo{
					Name:        p.Name,
					Description: fmt.Sprintf("%s tool", p.Name),
					Type:        string(p.Type),
				})
			}
		}
		return tools
	}

	if my.pluginManager == nil {
		return nil
	}

	plugins := my.pluginManager.ListPlugins()
	var tools []ToolInfo
	for _, p := range plugins {
		if p.Type == types.PluginTypeTool && p.Enabled {
			tools = append(tools, ToolInfo{
				Name:        p.Name,
				Description: fmt.Sprintf("%s tool", p.Name),
				Type:        string(p.Type),
			})
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

type ToolInfo struct {
	Name        string
	Description string
	Type        string
}

// CreateSession creates a new session.
func (my *Engine) CreateSession(sessionId string) error {
	if sessionId == "" {
		return fmt.Errorf("session ID cannot be empty")
	}
	my.sessions[sessionId] = NewSession(sessionId)
	return nil
}

// CloseSession closes a session.
func (my *Engine) CloseSession(sessionId string) error {
	delete(my.sessions, sessionId)
	return nil
}

// Close closes the engine.
func (my *Engine) Close() error {
	// Clear all sessions
	for key := range my.sessions {
		delete(my.sessions, key)
	}
	return nil
}
