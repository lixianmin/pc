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
			got := NewEngine(nil)
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
			e := NewEngine(nil)
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
			e := NewEngine(nil)
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
			e := NewEngine(nil)
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

func TestProcessMessageWithContext(t *testing.T) {
	tests := []struct {
		name          string
		sessionId     string
		messages      []struct {
			role    string
			content string
		}
		finalMessage  string
		wantContextLen int
	}{
		{
			name:      "maintain conversation context",
			sessionId: "test-session-context",
			messages: []struct {
				role    string
				content string
			}{
				{role: "user", content: "My name is Alice"},
				{role: "assistant", content: "Hello Alice!"},
			},
			finalMessage:   "What is my name?",
			wantContextLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewEngine(nil)
			ctx := context.Background()
			_ = e.CreateSession(tt.sessionId)

			// Add historical messages
			session := e.sessions[tt.sessionId]
			for _, msg := range tt.messages {
				if msg.role == "user" {
					session.AddMessage(msg.role, msg.content)
				}
			}

			_, _ = e.ProcessMessage(ctx, tt.sessionId, tt.finalMessage)

			// Verify context is maintained
			msgs := session.GetMessages()
			if len(msgs) != tt.wantContextLen {
				t.Errorf("context length = %v, want %v", len(msgs), tt.wantContextLen)
			}
		})
	}
}

func TestEngine_SetSystemPrompt(t *testing.T) {
	tests := []struct {
		name         string
		systemPrompt string
		wantPrompt   string
	}{
		{
			name:         "set system prompt",
			systemPrompt: "你是一个有用的助手。",
			wantPrompt:   "你是一个有用的助手。",
		},
		{
			name:         "set empty system prompt",
			systemPrompt: "",
			wantPrompt:   "",
		},
		{
			name:         "set multi-line system prompt",
			systemPrompt: "你是 Agent。\n\n## 技能\n- 编程\n- 写作",
			wantPrompt:   "你是 Agent。\n\n## 技能\n- 编程\n- 写作",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewEngine(nil)
			e.SetSystemPrompt(tt.systemPrompt)

			if e.systemPrompt != tt.wantPrompt {
				t.Errorf("SetSystemPrompt() = %v, want %v", e.systemPrompt, tt.wantPrompt)
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
			e := NewEngine(nil)
			err := e.Close()
			if (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
