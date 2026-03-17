package engine

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestStreamChunk(t *testing.T) {
	tests := []struct {
		name    string
		chunk   StreamChunk
		wantStr string
	}{
		{
			name:    "content chunk",
			chunk:   NewStreamChunk("hello"),
			wantStr: "StreamChunk{Content: \"hello\"}",
		},
		{
			name:    "done chunk",
			chunk:   NewStreamDone(),
			wantStr: "StreamChunk{Done: true}",
		},
		{
			name:    "error chunk",
			chunk:   NewStreamError(fmt.Errorf("test error")),
			wantStr: "StreamChunk{Error: test error}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.chunk.String()
			if got != tt.wantStr {
				t.Errorf("String() = %q, want %q", got, tt.wantStr)
			}
		})
	}
}

func TestEngine_ProcessMessageStream(t *testing.T) {
	tests := []struct {
		name        string
		sessionId   string
		message     string
		wantErr     bool
		errContains string
	}{
		{
			name:        "empty session ID",
			sessionId:   "",
			message:     "hello",
			wantErr:     true,
			errContains: "session ID cannot be empty",
		},
		{
			name:        "empty message",
			sessionId:   "test-session",
			message:     "",
			wantErr:     true,
			errContains: "message cannot be empty",
		},
		{
			name:        "session not found",
			sessionId:   "nonexistent",
			message:     "hello",
			wantErr:     true,
			errContains: "session not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine(nil)

			if tt.sessionId != "nonexistent" && tt.sessionId != "" {
				engine.CreateSession(tt.sessionId)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			ch := engine.ProcessMessageStream(ctx, tt.sessionId, tt.message)

			var gotErr bool
			var errMsg string
			for chunk := range ch {
				if chunk.Error != "" {
					gotErr = true
					errMsg = chunk.Error
				}
			}

			if tt.wantErr {
				if !gotErr {
					t.Errorf("ProcessMessageStream() expected error, got none")
					return
				}
				if tt.errContains != "" && !containsSubstring(errMsg, tt.errContains) {
					t.Errorf("ProcessMessageStream() error = %q, want containing %q", errMsg, tt.errContains)
				}
			}
		})
	}
}

func TestEngine_ProcessMessageStream_Echo(t *testing.T) {
	engine := NewEngine(nil)
	engine.CreateSession("test-echo")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ch := engine.ProcessMessageStream(ctx, "test-echo", "hello world")

	var chunks []StreamChunk
	for chunk := range ch {
		chunks = append(chunks, chunk)
	}

	if len(chunks) < 2 {
		t.Errorf("ProcessMessageStream() got %d chunks, want at least 2", len(chunks))
		return
	}

	if chunks[0].Error != "" {
		t.Errorf("ProcessMessageStream() unexpected error: %s", chunks[0].Error)
		return
	}

	if !chunks[len(chunks)-1].Done {
		t.Errorf("ProcessMessageStream() last chunk should have Done=true")
	}
}
