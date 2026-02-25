package memory

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
func (s *Service) AddMessage(sessionId, role, content string) error {
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
	if limit, ok := s.limits[sessionId]; ok && limit > 0 {
		sessions := s.sessions[sessionId]
		if len(sessions) >= limit {
			// Keep only the last (limit-1) messages
			start := len(sessions) - (limit - 1)
			s.sessions[sessionId] = append(sessions[start:], msg)
		} else {
			s.sessions[sessionId] = append(s.sessions[sessionId], msg)
		}
	} else {
		s.sessions[sessionId] = append(s.sessions[sessionId], msg)
	}

	return nil
}

// GetMessages returns all messages for a session.
func (s *Service) GetMessages(sessionId string) ([]Message, error) {
	messages, ok := s.sessions[sessionId]
	if !ok {
		return []Message{}, nil
	}
	// Return a copy to avoid external modification
	result := make([]Message, len(messages))
	copy(result, messages)
	return result, nil
}

// SetContextLimit sets the context window limit for a session.
func (s *Service) SetContextLimit(sessionId string, limit int) error {
	if limit < 0 {
		return fmt.Errorf("limit cannot be negative")
	}
	s.limits[sessionId] = limit
	return nil
}

// ClearSession clears all messages for a session.
func (s *Service) ClearSession(sessionId string) error {
	s.sessions[sessionId] = []Message{}
	return nil
}

// Close closes the memory service.
func (s *Service) Close() error {
	// Clear all sessions
	for key := range s.sessions {
		delete(s.sessions, key)
	}
	for key := range s.limits {
		delete(s.limits, key)
	}
	return nil
}
