package debug

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPromptRecorder_IsEnabled(t *testing.T) {
	tests := []struct {
		name       string
		promptsDir string
		want       bool
	}{
		{
			name:       "enabled with dir",
			promptsDir: "/tmp/prompts",
			want:       true,
		},
		{
			name:       "disabled with empty dir",
			promptsDir: "",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := NewPromptRecorder(tt.promptsDir, 100)
			if got := recorder.IsEnabled(); got != tt.want {
				t.Errorf("IsEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPromptRecorder_Record_Disabled(t *testing.T) {
	recorder := NewPromptRecorder("", 100)
	messages := []map[string]string{
		{"role": "user", "content": "hello"},
	}

	requestId, err := recorder.Record("session-1", messages)
	if err != nil {
		t.Errorf("Record() error = %v", err)
	}
	if requestId != "" {
		t.Errorf("Record() should return empty requestId when disabled, got %v", requestId)
	}
}

func TestPromptRecorder_Record_Enabled(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "prompts-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	recorder := NewPromptRecorder(tmpDir, 100)
	messages := []map[string]string{
		{"role": "system", "content": "you are an assistant"},
		{"role": "user", "content": "hello"},
		{"role": "assistant", "content": "hi there"},
	}

	requestId, err := recorder.Record("session-1", messages)
	if err != nil {
		t.Errorf("Record() error = %v", err)
	}
	if requestId == "" {
		t.Error("Record() should return non-empty requestId when enabled")
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 file, got %d", len(entries))
	}

	filename := entries[0].Name()
	if !strings.HasSuffix(filename, ".log") {
		t.Errorf("expected log file, got %s", filename)
	}
	if !strings.Contains(filename, requestId) {
		t.Errorf("filename should contain requestId %s, got %s", requestId, filename)
	}
}

func TestPromptRecorder_Record_Content(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "prompts-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	recorder := NewPromptRecorder(tmpDir, 100)
	messages := []map[string]string{
		{"role": "system", "content": "system prompt"},
		{"role": "user", "content": "user message"},
		{"role": "assistant", "content": "assistant response"},
		{"role": "user", "content": "Tool result: tool output"},
	}

	requestId, err := recorder.Record("test-session", messages)
	if err != nil {
		t.Fatalf("Record() error = %v", err)
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tmpDir, entries[0].Name()))
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	var record PromptRecord
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if record.RequestID != requestId {
		t.Errorf("RequestID = %s, want %s", record.RequestID, requestId)
	}
	if record.SessionID != "test-session" {
		t.Errorf("SessionID = %s, want test-session", record.SessionID)
	}
	if len(record.Messages) != 4 {
		t.Errorf("expected 4 messages, got %d", len(record.Messages))
	}

	if record.Stats.System != len("system prompt") {
		t.Errorf("Stats.System = %d, want %d", record.Stats.System, len("system prompt"))
	}
	if record.Stats.User != len("user message") {
		t.Errorf("Stats.User = %d, want %d", record.Stats.User, len("user message"))
	}
	if record.Stats.History != len("assistant response")+len("Tool result: tool output") {
		t.Errorf("Stats.History = %d, want %d", record.Stats.History, len("assistant response")+len("Tool result: tool output"))
	}
}

func TestPromptRecorder_Cleanup(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "prompts-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	recorder := NewPromptRecorder(tmpDir, 3)

	for i := 0; i < 5; i++ {
		messages := []map[string]string{
			{"role": "user", "content": "message"},
		}
		recorder.Record("session", messages)
	}
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}

	logFiles := 0
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".log") {
			logFiles++
		}
	}

	if logFiles > 3 {
		t.Errorf("expected at most 3 files after cleanup, got %d", logFiles)
	}
}

func TestPromptRecorder_MaxPromptFiles_Default(t *testing.T) {
	recorder := NewPromptRecorder("/tmp", 0)
	if recorder.maxPromptFiles != 100 {
		t.Errorf("expected default maxPromptFiles 100, got %d", recorder.maxPromptFiles)
	}

	recorder = NewPromptRecorder("/tmp", -1)
	if recorder.maxPromptFiles != 100 {
		t.Errorf("expected default maxPromptFiles 100, got %d", recorder.maxPromptFiles)
	}
}
