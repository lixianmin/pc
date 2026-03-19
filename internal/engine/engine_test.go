package engine

import (
	"context"
	"testing"

	"github.com/lixianmin/pc/internal/task"
)

func TestEngine_TaskDecomposition(t *testing.T) {
	tests := []struct {
		name          string
		sessionId     string
		message       string
		wantTask      bool
		wantErr       bool
		wantTaskTitle string
		wantStepCount int
	}{
		{
			name:          "decompose goal with /task prefix",
			sessionId:     "test-task-1",
			message:       "/task deploy the application",
			wantTask:      true,
			wantErr:       false,
			wantTaskTitle: "deploy the application",
			wantStepCount: 4,
		},
		{
			name:          "decompose goal with 我想 prefix",
			sessionId:     "test-task-2",
			message:       "我想测试这个功能",
			wantTask:      true,
			wantErr:       false,
			wantTaskTitle: "测试这个功能",
			wantStepCount: 4,
		},
		{
			name:      "regular message does not trigger decomposition",
			sessionId: "test-task-3",
			message:   "Hello, how are you?",
			wantTask:  false,
			wantErr:   false,
		},
		{
			name:      "empty goal returns error",
			sessionId: "test-task-4",
			message:   "/task   ",
			wantTask:  false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewEngine(nil)
			e.SetTaskEnabled(true)
			ctx := context.Background()
			_ = e.FetchSession(tt.sessionId)

			resp, err := e.ProcessMessage(ctx, tt.sessionId, tt.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			tm := e.GetTaskManager()
			if tm == nil {
				t.Error("Expected TaskManager to be initialized")
				return
			}

			tasks := tm.ListTasks()
			if tt.wantTask {
				if len(tasks) == 0 {
					t.Error("Expected at least one task to be created")
					return
				}
				var foundTask *task.Task
				for _, tk := range tasks {
					if tk.Title == tt.wantTaskTitle {
						foundTask = tk
						break
					}
				}
				if foundTask == nil {
					t.Errorf("Expected task with title %q not found", tt.wantTaskTitle)
					return
				}
				if len(foundTask.Steps) != tt.wantStepCount {
					t.Errorf("Task steps = %v, want %v", len(foundTask.Steps), tt.wantStepCount)
				}
				if resp == "" {
					t.Error("Expected non-empty response for task decomposition")
					return
				}
			} else if !tt.wantErr {
				if len(tasks) > 0 {
					t.Errorf("Expected no tasks for non-goal message, got %d", len(tasks))
				}
			}
		})
	}
}
