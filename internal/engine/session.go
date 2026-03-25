package engine

import (
	"fmt"

	"github.com/lixianmin/pc/baml_client/types"
)

// Message represents a message in a session.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Session represents a conversation session.
type Session struct {
	Id       string
	Messages []Message
	Limit    int
}

// NewSession creates a new session.
func NewSession(sessionId string) *Session {
	return &Session{
		Id:       sessionId,
		Messages: make([]Message, 0),
		Limit:    0, // 0 means no limit
	}
}

// AddMessage adds a message to the session.
func (my *Session) AddMessage(role, content string) *Message {
	var msg = Message{
		Role:    role,
		Content: content,
	}

	// Check context limit and trim if needed
	if my.Limit > 0 && len(my.Messages) >= my.Limit {
		// Keep only the last (limit-1) messages
		start := len(my.Messages) - (my.Limit - 1)
		my.Messages = append(my.Messages[start:], msg)
	} else {
		my.Messages = append(my.Messages, msg)
	}

	return &msg
}

// CloneMessages returns all messages in the session.
func (my *Session) CloneMessages() []Message {
	// Return a copy to avoid external modification
	var result = make([]Message, len(my.Messages))
	copy(result, my.Messages)
	return result
}

func (my *Session) AsBamlMessages() []types.Message {
	var result = make([]types.Message, 0, len(my.Messages))
	for _, msg := range my.Messages {
		result = append(result, types.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	return result
}

// SetContextLimit sets the context window limit.
func (my *Session) SetContextLimit(limit int) error {
	if limit < 0 {
		return fmt.Errorf("limit cannot be negative")
	}
	my.Limit = limit
	return nil
}

// Clear clears all messages in the session.
func (my *Session) Clear() {
	my.Messages = make([]Message, 0)
}
