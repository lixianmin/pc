package task

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewHeartbeatScheduler(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "create new scheduler",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewHeartbeatScheduler()
			if s == nil {
				t.Error("NewHeartbeatScheduler() returned nil")
			}
		})
	}
}

func TestHeartbeatTask_IsDue(t *testing.T) {
	tests := []struct {
		name      string
		task      HeartbeatTask
		now       time.Time
		wantDue   bool
	}{
		{
			name: "task is due",
			task: HeartbeatTask{
				Schedule:    "0 * * * *",
				LastRun:     time.Now().Add(-2 * time.Hour).UnixMilli(),
			},
			now:     time.Now(),
			wantDue: true,
		},
		{
			name: "task not due",
			task: HeartbeatTask{
				Schedule:    "0 0 * * *",
				LastRun:     time.Now().Add(-1 * time.Hour).UnixMilli(),
			},
			now:     time.Now(),
			wantDue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			due := tt.task.IsDue(tt.now)
			if due != tt.wantDue {
				t.Errorf("IsDue() = %v, want %v", due, tt.wantDue)
			}
		})
	}
}

func TestHeartbeatScheduler_LoadFromFile(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantCount int
		wantErr   bool
	}{
		{
			name: "load valid heartbeat file",
			content: `# Heartbeat Tasks

## Daily Health Check
schedule: "0 2 * * *"
description: Check project health
onFailure: notify

## Weekly Report
schedule: "0 10 * * 0"
description: Generate weekly report
onFailure: ignore
`,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "load empty file",
			content:   "# Heartbeat Tasks\n",
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "heartbeat.md")
			os.WriteFile(filePath, []byte(tt.content), 0644)

			s := NewHeartbeatScheduler()
			tasks, err := s.LoadFromFile(filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(tasks) != tt.wantCount {
				t.Errorf("LoadFromFile() task count = %v, want %v", len(tasks), tt.wantCount)
			}
		})
	}
}

func TestHeartbeatScheduler_StartAndStop(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "start and stop scheduler",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewHeartbeatScheduler()
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			err := s.Start(ctx)
			if err != nil {
				t.Errorf("Start() error = %v", err)
			}

			// Wait for context to timeout
			<-ctx.Done()
			s.Stop()
		})
	}
}

func TestHeartbeatScheduler_AddTask(t *testing.T) {
	tests := []struct {
		name    string
		task    HeartbeatTask
		wantErr bool
	}{
		{
			name: "add valid task",
			task: HeartbeatTask{
				Name:        "Test Task",
				Schedule:    "0 * * * *",
				Description: "Test description",
				OnFailure:   "notify",
			},
			wantErr: false,
		},
		{
			name: "add task with empty schedule",
			task: HeartbeatTask{
				Name:     "Invalid Task",
				Schedule: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewHeartbeatScheduler()
			err := s.AddTask(tt.task)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddTask() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHeartbeatScheduler_RemoveTask(t *testing.T) {
	tests := []struct {
		name      string
		taskName  string
		setupFunc func(*HeartbeatScheduler)
		wantErr   bool
	}{
		{
			name:     "remove existing task",
			taskName: "Test Task",
			setupFunc: func(s *HeartbeatScheduler) {
				s.AddTask(HeartbeatTask{
					Name:     "Test Task",
					Schedule: "0 * * * *",
				})
			},
			wantErr: false,
		},
		{
			name:     "remove non-existent task",
			taskName: "NonExistent",
			setupFunc: func(s *HeartbeatScheduler) {},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewHeartbeatScheduler()
			tt.setupFunc(s)

			err := s.RemoveTask(tt.taskName)
			if (err != nil) != tt.wantErr {
				t.Errorf("RemoveTask() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
