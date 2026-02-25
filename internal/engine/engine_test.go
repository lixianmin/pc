package engine

import (
	"context"
	"testing"
)

func TestNewEngine(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "create new engine",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewEngine()
			if got == nil {
				t.Error("NewEngine() returned nil")
			}
		})
	}
}

func TestCreateSession(t *testing.T) {
	tests := []struct {
		name      string
		sessionId string
		wantErr   bool
	}{
		{
			name:      "create session with valid ID",
			sessionId: "test-session-1",
			wantErr:   false,
		},
		{
			name:      "create session with empty ID",
			sessionId: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewEngine()
			err := e.CreateSession(tt.sessionId)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateSession() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCloseSession(t *testing.T) {
	tests := []struct {
		name      string
		sessionId string
		wantErr   bool
	}{
		{
			name:      "close existing session",
			sessionId: "test-session-1",
			wantErr:   false,
		},
		{
			name:      "close non-existent session",
			sessionId: "non-existent",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewEngine()
			if !tt.wantErr && tt.sessionId != "non-existent" {
				_ = e.CreateSession(tt.sessionId)
			}
			err := e.CloseSession(tt.sessionId)
			if (err != nil) != tt.wantErr {
				t.Errorf("CloseSession() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProcessMessage(t *testing.T) {
	tests := []struct {
		name      string
		sessionId string
		message   string
		wantErr   bool
	}{
		{
			name:      "process message with valid session",
			sessionId: "test-session-1",
			message:   "Hello, world!",
			wantErr:   false,
		},
		{
			name:      "process empty message",
			sessionId: "test-session-2",
			message:   "",
			wantErr:   true,
		},
		{
			name:      "process message without session",
			sessionId: "",
			message:   "Hello!",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewEngine()
			ctx := context.Background()

			if tt.sessionId != "" {
				_ = e.CreateSession(tt.sessionId)
			}

			resp, err := e.ProcessMessage(ctx, tt.sessionId, tt.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && resp == "" {
				t.Error("ProcessMessage() returned empty response")
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
			name:    "close engine",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewEngine()
			err := e.Close()
			if (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
