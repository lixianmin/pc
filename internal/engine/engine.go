package engine

import (
	"context"
	"fmt"
)

// CoreEngine is the interface for the core engine.
type CoreEngine interface {
	// ProcessMessage processes an incoming message and returns the response.
	ProcessMessage(ctx context.Context, sessionID, message string) (string, error)

	// CreateSession creates a new session.
	CreateSession(sessionID string) error

	// CloseSession closes a session.
	CloseSession(sessionID string) error

	// Close closes the engine.
	Close() error
}

// Engine is the core engine implementation.
type Engine struct {
	// TODO: Add plugin manager, memory, etc.
	sessions map[string]bool
}

// NewEngine creates a new core engine.
func NewEngine() *Engine {
	return &Engine{
		sessions: make(map[string]bool),
	}
}

// ProcessMessage processes an incoming message.
func (e *Engine) ProcessMessage(ctx context.Context, sessionID, message string) (string, error) {
	// Validate inputs
	if sessionID == "" {
		return "", fmt.Errorf("session ID cannot be empty")
	}
	if message == "" {
		return "", fmt.Errorf("message cannot be empty")
	}

	// Check if session exists
	if !e.sessions[sessionID] {
		return "", fmt.Errorf("session not found: %s", sessionID)
	}

	// TODO: Implement actual message processing with LLM plugin
	// For now, return a mock response
	return fmt.Sprintf("Echo: %s", message), nil
}

// CreateSession creates a new session.
func (e *Engine) CreateSession(sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}
	e.sessions[sessionID] = true
	return nil
}

// CloseSession closes a session.
func (e *Engine) CloseSession(sessionID string) error {
	delete(e.sessions, sessionID)
	return nil
}

// Close closes the engine.
func (e *Engine) Close() error {
	// TODO: Implement cleanup
	return nil
}
