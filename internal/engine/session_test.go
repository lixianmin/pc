package engine

import (
	"testing"
)

func TestNewSession(t *testing.T) {
	tests := []struct {
		name      string
		sessionId string
	}{
		{
			name:      "create new session",
			sessionId: "test-session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewSession(tt.sessionId)
			if got == nil {
				t.Error("NewSession() returned nil")
			}
			if got.Id != tt.sessionId {
				t.Errorf("NewSession().Id = %v, want %v", got.Id, tt.sessionId)
			}
			if got.Messages == nil {
				t.Error("NewSession().Messages is nil")
			}
			if len(got.Messages) != 0 {
				t.Errorf("NewSession().Messages length = %v, want 0", len(got.Messages))
			}
		})
	}
}

func TestSessionAddMessage(t *testing.T) {
	tests := []struct {
		name    string
		role    string
		content string
		wantErr bool
	}{
		{
			name:    "add user message",
			role:    "user",
			content: "Hello, world!",
			wantErr: false,
		},
		{
			name:    "add assistant message",
			role:    "assistant",
			content: "Hi there!",
			wantErr: false,
		},
		{
			name:    "add message with empty content",
			role:    "user",
			content: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSession("test-session")
			err := s.AddMessage(tt.role, tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				msgs := s.GetMessages()
				if len(msgs) != 1 {
					t.Errorf("message count = %v, want 1", len(msgs))
				}
				if msgs[0].Role != tt.role {
					t.Errorf("message role = %v, want %v", msgs[0].Role, tt.role)
				}
				if msgs[0].Content != tt.content {
					t.Errorf("message content = %v, want %v", msgs[0].Content, tt.content)
				}
			}
		})
	}
}

func TestSessionGetMessages(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*Session)
		wantCount int
	}{
		{
			name:      "get messages from empty session",
			setup:     func(s *Session) {},
			wantCount: 0,
		},
		{
			name: "get messages from session with messages",
			setup: func(s *Session) {
				s.AddMessage("user", "Hello")
				s.AddMessage("assistant", "Hi")
			},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSession("test-session")
			tt.setup(s)

			msgs := s.GetMessages()
			if len(msgs) != tt.wantCount {
				t.Errorf("GetMessages() count = %v, want %v", len(msgs), tt.wantCount)
			}
		})
	}
}

func TestSessionGetMessagesReturnsCopy(t *testing.T) {
	s := NewSession("test-session")
	s.AddMessage("user", "Hello")

	msgs := s.GetMessages()
	msgs[0].Content = "Modified"

	originalMsgs := s.GetMessages()
	if originalMsgs[0].Content == "Modified" {
		t.Error("GetMessages() returned reference instead of copy")
	}
}

func TestSessionSetContextLimit(t *testing.T) {
	tests := []struct {
		name    string
		limit   int
		wantErr bool
	}{
		{
			name:    "set valid context limit",
			limit:   10,
			wantErr: false,
		},
		{
			name:    "set zero limit",
			limit:   0,
			wantErr: false,
		},
		{
			name:    "set negative limit",
			limit:   -1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSession("test-session")
			err := s.SetContextLimit(tt.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetContextLimit() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && s.Limit != tt.limit {
				t.Errorf("Session.Limit = %v, want %v", s.Limit, tt.limit)
			}
		})
	}
}

func TestSessionContextLimit(t *testing.T) {
	s := NewSession("test-session")
	s.SetContextLimit(3)

	// Add 5 messages, only 3 should be kept
	for i := 0; i < 5; i++ {
		s.AddMessage("user", "message")
	}

	msgs := s.GetMessages()
	if len(msgs) != 3 {
		t.Errorf("message count = %v, want 3", len(msgs))
	}
}

func TestSessionClear(t *testing.T) {
	s := NewSession("test-session")
	s.AddMessage("user", "Hello")
	s.AddMessage("assistant", "Hi")

	s.Clear()

	msgs := s.GetMessages()
	if len(msgs) != 0 {
		t.Errorf("message count after Clear() = %v, want 0", len(msgs))
	}
}
