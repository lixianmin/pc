package engine

import (
	"context"
	"fmt"
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
	// Get conversation history
	var history = session.GetMessages()

	// Build messages for LLM
	// Pre-allocate capacity: system prompt (optional) + history + current message
	capacity := len(history) + 1
	if my.systemPrompt != "" {
		capacity++
	}
	messages := make([]map[string]string, 0, capacity)

	// Add system prompt if available
	if my.systemPrompt != "" {
		messages = append(messages, map[string]string{
			"role":    "system",
			"content": my.systemPrompt,
		})
	}

	// Add conversation history
	for _, msg := range history {
		messages = append(messages, map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}
	// Add current message
	messages = append(messages, map[string]string{
		"role":    "user",
		"content": message,
	})

	// Use llmCallback if available (for testing)
	if my.llmCallback != nil {
		return my.llmCallback(ctx, session, my.systemPrompt)
	}

	// Call LLM plugin
	params := map[string]any{
		"messages": messages,
	}

	result, err := my.pluginManager.CallPlugin(my.llmPlugin, "complete", params)
	if err != nil {
		return "", err
	}

	// Extract response content from result
	// TODO: 如果result的处理逻辑就只是取到其中的content字段，那么是否可以在plugin协议层面做一个约定，直接把content作为结果返回，而不是在这里再解析一次
	resultMap, ok := result.(map[string]any)
	if !ok {
		return "", fmt.Errorf("unexpected LLM response format")
	}

	content, ok := resultMap["content"].(string)
	if !ok {
		return "", fmt.Errorf("LLM response missing content")
	}

	return content, nil
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
