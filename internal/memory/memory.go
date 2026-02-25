package memory

import "fmt"

// MemoryService is the interface for the memory service.
type MemoryService interface {
	// AddMessage adds a message to a session.
	AddMessage(sessionID, role, content string) error

	// GetMessages returns all messages for a session.
	GetMessages(sessionID string) ([]Message, error)

	// SetContextLimit sets the context window limit for a session.
	SetContextLimit(sessionID string, limit int) error

	// ClearSession clears all messages for a session.
	ClearSession(sessionID string) error

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
func (s *Service) AddMessage(sessionID, role, content string) error {
	if sessionID == "" {
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
	if limit, ok := s.limits[sessionID]; ok && limit > 0 {
		sessions := s.sessions[sessionID]
		if len(sessions) >= limit {
			// Keep only the last (limit-1) messages
			start := len(sessions) - (limit - 1)
			s.sessions[sessionID] = append(sessions[start:], msg)
		} else {
			s.sessions[sessionID] = append(s.sessions[sessionID], msg)
		}
	} else {
		s.sessions[sessionID] = append(s.sessions[sessionID], msg)
	}

	return nil
}

// GetMessages returns all messages for a session.
func (s *Service) GetMessages(sessionID string) ([]Message, error) {
	messages, ok := s.sessions[sessionID]
	if !ok {
		return []Message{}, nil
	}
	// Return a copy to avoid external modification
	result := make([]Message, len(messages))
	copy(result, messages)
	return result, nil
}

// SetContextLimit sets the context window limit for a session.
func (s *Service) SetContextLimit(sessionID string, limit int) error {
	if limit < 0 {
		return fmt.Errorf("limit cannot be negative")
	}
	s.limits[sessionID] = limit
	return nil
}

// ClearSession clears all messages for a session.
func (s *Service) ClearSession(sessionID string) error {
	s.sessions[sessionID] = []Message{}
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
