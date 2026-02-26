package engine

import (
	"context"
	"fmt"

	"github.com/lixianmin/pc/internal/memory"
	"github.com/lixianmin/pc/internal/plugin"
	"github.com/lixianmin/pc/pkg/types"
)

// CoreEngine is the interface for the core engine.
type CoreEngine interface {
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
	sessions      map[string]bool
	memory        memory.MemoryService
	llmPlugin     *types.Plugin
}

// NewEngine creates a new core engine.
func NewEngine(pm *plugin.PluginManager) *Engine {
	return &Engine{
		pluginManager: pm,
		sessions:      make(map[string]bool),
		memory:        memory.NewService(),
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
	if !my.sessions[sessionId] {
		return "", fmt.Errorf("session not found: %s", sessionId)
	}

	// Save user message to memory
	if err := my.memory.AddMessage(sessionId, "user", message); err != nil {
		return "", fmt.Errorf("failed to save user message: %w", err)
	}

	// Generate response using LLM plugin if available, otherwise echo
	var response string
	if my.pluginManager != nil && my.llmPlugin != nil {
		resp, err := my.callLLM(ctx, sessionId, message)
		if err != nil {
			return "", fmt.Errorf("failed to call LLM: %w", err)
		}
		response = resp
	} else {
		response = fmt.Sprintf("Echo: %s", message)
	}

	// Save assistant response to memory
	if err := my.memory.AddMessage(sessionId, "assistant", response); err != nil {
		return "", fmt.Errorf("failed to save assistant response: %w", err)
	}

	return response, nil
}

// callLLM calls the LLM plugin to generate a response.
func (my *Engine) callLLM(ctx context.Context, sessionId, message string) (string, error) {
	// Get conversation history
	history, err := my.memory.GetMessages(sessionId)
	if err != nil {
		return "", fmt.Errorf("failed to get conversation history: %w", err)
	}

	// Build messages for LLM
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
		"model":    "gpt-4o-mini", // Default model
	}

	result, err := my.pluginManager.CallPlugin(my.llmPlugin, "complete", params)
	if err != nil {
		return "", err
	}

	// Extract response content from result
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
	my.sessions[sessionId] = true
	return nil
}

// CloseSession closes a session.
func (my *Engine) CloseSession(sessionId string) error {
	delete(my.sessions, sessionId)
	return nil
}

// Close closes the engine.
func (my *Engine) Close() error {
	if my.memory != nil {
		return my.memory.Close()
	}
	return nil
}
