package engine

import (
	"context"
	"fmt"

	"github.com/lixianmin/pc/internal/plugin"
	"github.com/lixianmin/pc/pkg/types"
)

// IEngine is the interface for the core engine.
type IEngine interface {
	// ProcessMessage processes an incoming message and returns the response.
	ProcessMessage(ctx context.Context, sessionId, message string) (string, error)

	// CreateSession creates a new session.
	CreateSession(sessionId string) error

	// CloseSession closes a session.
	CloseSession(sessionId string) error

	// Close closes the engine.
	Close() error
}

// Engine is the core engine implementation.
type Engine struct {
	pluginManager *plugin.PluginManager
	sessions      map[string]*Session
	llmPlugin     *types.Plugin
}

// NewEngine creates a new core engine.
func NewEngine(pm *plugin.PluginManager) *Engine {
	return &Engine{
		pluginManager: pm,
		sessions:      make(map[string]*Session),
	}
}

// SetLLMPlugin sets the LLM plugin to use for generating responses.
func (my *Engine) SetLLMPlugin(p *types.Plugin) {
	my.llmPlugin = p
}

// ProcessMessage processes an incoming message.
func (my *Engine) ProcessMessage(ctx context.Context, sessionId, message string) (string, error) {
	// Validate inputs
	if sessionId == "" {
		return "", fmt.Errorf("session ID cannot be empty")
	}

	if message == "" {
		return "", fmt.Errorf("message cannot be empty")
	}

	// Check if session exists
	session, exists := my.sessions[sessionId]
	if !exists {
		return "", fmt.Errorf("session not found: %s", sessionId)
	}

	// Save user message to session
	if err := session.AddMessage("user", message); err != nil {
		return "", fmt.Errorf("failed to save user message: %w", err)
	}

	// Generate response using LLM plugin if available, otherwise echo
	var response string
	if my.pluginManager != nil && my.llmPlugin != nil {
		resp, err := my.callLLM(ctx, session, message)
		if err != nil {
			return "", fmt.Errorf("failed to call LLM: %w", err)
		}
		response = resp
	} else {
		response = fmt.Sprintf("Echo: %s", message)
	}

	// Save assistant response to session
	if err := session.AddMessage("assistant", response); err != nil {
		return "", fmt.Errorf("failed to save assistant response: %w", err)
	}

	return response, nil
}

// callLLM calls the LLM plugin to generate a response.
func (my *Engine) callLLM(ctx context.Context, session *Session, message string) (string, error) {
	// Get conversation history
	var history = session.GetMessages()

	// Build messages for LLM
	// todo: 每次callLLM的时候都把copy一遍全部的历史消息到map中，感觉非常低效，似乎没有必要做格式转换
	messages := make([]map[string]string, 0, len(history)+1)
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

	// Call LLM plugin
	params := map[string]any{
		"messages": messages,
		// TODO: 如果call llm也是走plugin协议，那么llm应该在插件的代码里定义，而不是在这里硬编码
		"model": "gpt-4o-mini", // Default model
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
