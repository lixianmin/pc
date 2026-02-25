package engine

import (
	"context"
	"fmt"

	"github.com/lixianmin/pc/internal/memory"
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
	// TODO: Add plugin manager
	sessions map[string]bool
	memory   memory.MemoryService
}

// NewEngine creates a new core engine.
func NewEngine() *Engine {
	return &Engine{
		sessions: make(map[string]bool),
		memory:   memory.NewService(),
	}
}

// ProcessMessage processes an incoming message.
func (e *Engine) ProcessMessage(ctx context.Context, sessionId, message string) (string, error) {
	// Validate inputs
	if sessionId == "" {
		return "", fmt.Errorf("session ID cannot be empty")
	}
	if message == "" {
		return "", fmt.Errorf("message cannot be empty")
	}

	// Check if session exists
	if !e.sessions[sessionId] {
		return "", fmt.Errorf("session not found: %s", sessionId)
	}

	// Save user message to memory
	if err := e.memory.AddMessage(sessionId, "user", message); err != nil {
		return "", fmt.Errorf("failed to save user message: %w", err)
	}

	// TODO: Implement actual message processing with LLM plugin
	// For now, return a mock response and save it to memory
	response := fmt.Sprintf("Echo: %s", message)

	// Save assistant response to memory
	if err := e.memory.AddMessage(sessionId, "assistant", response); err != nil {
		return "", fmt.Errorf("failed to save assistant response: %w", err)
	}

	return response, nil
}

// CreateSession creates a new session.
func (e *Engine) CreateSession(sessionId string) error {
	if sessionId == "" {
		return fmt.Errorf("session ID cannot be empty")
	}
	e.sessions[sessionId] = true
	return nil
}

// CloseSession closes a session.
func (e *Engine) CloseSession(sessionId string) error {
	delete(e.sessions, sessionId)
	return nil
}

// Close closes the engine.
func (e *Engine) Close() error {
	if e.memory != nil {
		return e.memory.Close()
	}
	return nil
}
