package memory

// Deprecated: Use engine.Session instead. This package will be removed in a future version.
// The session management has been consolidated into the Engine package for better cohesion.

import "fmt"

// MemoryService is the interface for the memory service.
type MemoryService interface {
	// AddMessage adds a message to a session.
	AddMessage(sessionId, role, content string) error

	// GetMessages returns all messages for a session.
	GetMessages(sessionId string) ([]Message, error)

	// SetContextLimit sets the context window limit for a session.
	SetContextLimit(sessionId string, limit int) error

	// ClearSession clears all messages for a session.
	ClearSession(sessionId string) error

	// Close closes the memory service.
	Close() error
}

// Message represents a message in memory.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Service is the memory service implementation.
type Service struct {
	sessions map[string][]Message
	limits   map[string]int
}

// NewService creates a new memory service.
func NewService() *Service {
	return &Service{
		sessions: make(map[string][]Message),
		limits:   make(map[string]int),
	}
}

// AddMessage adds a message to a session.
func (my *Service) AddMessage(sessionId, role, content string) error {
	if sessionId == "" {
		return fmt.Errorf("session ID cannot be empty")
	}
	if content == "" {
		return fmt.Errorf("content cannot be empty")
	}

	msg := Message{
		Role:    role,
		Content: content,
	}

	// Check context limit and trim if needed
	if limit, ok := my.limits[sessionId]; ok && limit > 0 {
		sessions := my.sessions[sessionId]
		if len(sessions) >= limit {
			// Keep only the last (limit-1) messages
			start := len(sessions) - (limit - 1)
			my.sessions[sessionId] = append(sessions[start:], msg)
		} else {
			my.sessions[sessionId] = append(my.sessions[sessionId], msg)
		}
	} else {
		my.sessions[sessionId] = append(my.sessions[sessionId], msg)
	}

	return nil
}

// GetMessages returns all messages for a session.
func (my *Service) GetMessages(sessionId string) ([]Message, error) {
	messages, ok := my.sessions[sessionId]
	if !ok {
		return []Message{}, nil
	}
	// Return a copy to avoid external modification
	result := make([]Message, len(messages))
	copy(result, messages)
	return result, nil
}

// SetContextLimit sets the context window limit for a session.
func (my *Service) SetContextLimit(sessionId string, limit int) error {
	if limit < 0 {
		return fmt.Errorf("limit cannot be negative")
	}
	my.limits[sessionId] = limit
	return nil
}

// ClearSession clears all messages for a session.
func (my *Service) ClearSession(sessionId string) error {
	my.sessions[sessionId] = []Message{}
	return nil
}

// Close closes the memory service.
func (my *Service) Close() error {
	// Clear all sessions
	for key := range my.sessions {
		delete(my.sessions, key)
	}
	for key := range my.limits {
		delete(my.limits, key)
	}
	return nil
}
