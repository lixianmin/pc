package debug

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/lixianmin/got/loom"
	"github.com/oklog/ulid/v2"
)

type PromptStats struct {
	Total   int `json:"total"`
	System  int `json:"system"`
	History int `json:"history"`
	User    int `json:"user"`
}

type PromptRecord struct {
	RequestID string          `json:"request_id"`
	Timestamp string          `json:"timestamp"`
	SessionID string          `json:"session_id"`
	Messages  []MessageRecord `json:"messages"`
	Stats     PromptStats     `json:"stats"`
}

type MessageRecord struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type PromptRecorder struct {
	promptsDir     string
	maxPromptFiles int
}

func NewPromptRecorder(promptsDir string, maxPromptFiles int) *PromptRecorder {
	if maxPromptFiles <= 0 {
		maxPromptFiles = 100
	}
	return &PromptRecorder{
		promptsDir:     promptsDir,
		maxPromptFiles: maxPromptFiles,
	}
}

func (my *PromptRecorder) IsEnabled() bool {
	return my.promptsDir != ""
}

func (my *PromptRecorder) Record(sessionId string, messages []map[string]string) (string, error) {
	if !my.IsEnabled() {
		return "", nil
	}

	if err := os.MkdirAll(my.promptsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create prompts dir: %w", err)
	}

	requestId := my.generateRequestId()
	record := my.buildRecord(requestId, sessionId, messages)

	if err := my.writeRecord(record); err != nil {
		return "", err
	}

	loom.Go(func(later loom.Later) {
		my.cleanupOldFiles()
	})

	return requestId, nil
}

func (my *PromptRecorder) generateRequestId() string {
	return ulid.Make().String()[:8]
}

func (my *PromptRecorder) buildRecord(requestId, sessionId string, messages []map[string]string) PromptRecord {
	var stats PromptStats
	var msgRecords []MessageRecord

	for _, msg := range messages {
		role := msg["role"]
		content := msg["content"]
		contentLen := len(content)

		stats.Total += contentLen
		switch role {
		case "system":
			stats.System += contentLen
		case "user":
			if strings.HasPrefix(content, "Tool result:") {
				stats.History += contentLen
			} else {
				stats.User += contentLen
			}
		case "assistant":
			stats.History += contentLen
		}

		msgRecords = append(msgRecords, MessageRecord{
			Role:    role,
			Content: content,
		})
	}

	return PromptRecord{
		RequestID: requestId,
		Timestamp: time.Now().Format(time.RFC3339),
		SessionID: sessionId,
		Messages:  msgRecords,
		Stats:     stats,
	}
}

func (my *PromptRecorder) writeRecord(record PromptRecord) error {
	filename := fmt.Sprintf("%s_%s.log",
		time.Now().Format("2006-01-02_150405"),
		record.RequestID)
	path := filepath.Join(my.promptsDir, filename)

	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal record: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write record: %w", err)
	}

	return nil
}

func (my *PromptRecorder) cleanupOldFiles() {
	entries, err := os.ReadDir(my.promptsDir)
	if err != nil {
		return
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".log") {
			files = append(files, entry.Name())
		}
	}

	if len(files) <= my.maxPromptFiles {
		return
	}

	sort.Strings(files)

	for i := 0; i < len(files)-my.maxPromptFiles; i++ {
		path := filepath.Join(my.promptsDir, files[i])
		os.Remove(path)
	}
}
