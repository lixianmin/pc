package memory

import (
	"testing"
)

func TestNewService(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "create new service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewService()
			if got == nil {
				t.Error("NewService() returned nil")
			}
		})
	}
}

func TestAddMessage(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
		role      string
		content   string
		wantErr   bool
	}{
		{
			name:      "add user message",
			sessionID: "test-session",
			role:      "user",
			content:   "Hello, world!",
			wantErr:   false,
		},
		{
			name:      "add assistant message",
			sessionID: "test-session",
			role:      "assistant",
			content:   "Hi there!",
			wantErr:   false,
		},
		{
			name:      "add message with empty session",
			sessionID: "",
			role:      "user",
			content:   "Hello!",
			wantErr:   true,
		},
		{
			name:      "add message with empty content",
			sessionID: "test-session",
			role:      "user",
			content:   "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewService()
			err := s.AddMessage(tt.sessionID, tt.role, tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetMessages(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
		setup     func(*Service)
		wantCount int
		wantErr   bool
	}{
		{
			name:      "get messages from existing session",
			sessionID: "test-session",
			setup: func(s *Service) {
				s.AddMessage("test-session", "user", "Hello")
				s.AddMessage("test-session", "assistant", "Hi")
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "get messages from empty session",
			sessionID: "empty-session",
			setup:     func(s *Service) {},
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewService()
			tt.setup(s)

			msgs, err := s.GetMessages(tt.sessionID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMessages() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && len(msgs) != tt.wantCount {
				t.Errorf("GetMessages() count = %v, want %v", len(msgs), tt.wantCount)
			}
		})
	}
}

func TestSetContextLimit(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
		limit     int
		wantErr   bool
	}{
		{
			name:      "set valid context limit",
			sessionID: "test-session",
			limit:     10,
			wantErr:   false,
		},
		{
			name:      "set negative limit",
			sessionID: "test-session",
			limit:     -1,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewService()
			err := s.SetContextLimit(tt.sessionID, tt.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetContextLimit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClearSession(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
		setup     func(*Service)
		wantErr   bool
	}{
		{
			name:      "clear existing session",
			sessionID: "test-session",
			setup: func(s *Service) {
				s.AddMessage("test-session", "user", "Hello")
			},
			wantErr: false,
		},
		{
			name:      "clear non-existent session",
			sessionID: "non-existent",
			setup:     func(s *Service) {},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewService()
			tt.setup(s)

			err := s.ClearSession(tt.sessionID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ClearSession() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				msgs, _ := s.GetMessages(tt.sessionID)
				if len(msgs) != 0 {
					t.Errorf("ClearSession() session not cleared, got %d messages", len(msgs))
				}
			}
		})
	}
}

func TestClose(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "close service",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewService()
			err := s.Close()
			if (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
