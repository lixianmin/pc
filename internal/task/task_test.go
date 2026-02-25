package task

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewTask(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		wantErr bool
	}{
		{
			name:    "create valid task",
			title:   "Deploy v1.0.0",
			wantErr: false,
		},
		{
			name:    "create task with empty title",
			title:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := NewTask(tt.title)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if task.Id == "" {
					t.Error("NewTask() task.Id is empty")
				}
				if task.Title != tt.title {
					t.Errorf("NewTask() task.Title = %v, want %v", task.Title, tt.title)
				}
				if task.State != StatePending {
					t.Errorf("NewTask() task.State = %v, want %v", task.State, StatePending)
				}
				if task.CreateAt == 0 {
					t.Error("NewTask() task.CreateAt is zero")
				}
			}
		})
	}
}

func TestTask_AddStep(t *testing.T) {
	tests := []struct {
		name        string
		stepTitle   string
		wantErr     bool
		wantStepLen int
	}{
		{
			name:        "add valid step",
			stepTitle:   "Code review",
			wantErr:     false,
			wantStepLen: 1,
		},
		{
			name:        "add empty step",
			stepTitle:   "",
			wantErr:     true,
			wantStepLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, _ := NewTask("Test Task")
			err := task.AddStep(tt.stepTitle)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddStep() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(task.Steps) != tt.wantStepLen {
				t.Errorf("AddStep() len(Steps) = %v, want %v", len(task.Steps), tt.wantStepLen)
			}
		})
	}
}

func TestTask_UpdateState(t *testing.T) {
	tests := []struct {
		name      string
		newState  State
		wantState State
	}{
		{
			name:      "update to in_progress",
			newState:  StateInProgress,
			wantState: StateInProgress,
		},
		{
			name:      "update to completed",
			newState:  StateCompleted,
			wantState: StateCompleted,
		},
		{
			name:      "update to pending",
			newState:  StatePending,
			wantState: StatePending,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, _ := NewTask("Test Task")
			task.UpdateState(tt.newState)
			if task.State != tt.wantState {
				t.Errorf("UpdateState() task.State = %v, want %v", task.State, tt.wantState)
			}
			if tt.newState == StateCompleted && task.CompleteAt == 0 {
				t.Error("UpdateState() to completed should set CompleteAt")
			}
		})
	}
}

func TestNewManager(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
	}{
		{
			name:     "create new manager",
			filePath: filepath.Join(t.TempDir(), "task_list.md"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewManager(tt.filePath)
			if manager == nil {
				t.Error("NewManager() returned nil")
			}
			if manager.filePath != tt.filePath {
				t.Errorf("NewManager() filePath = %v, want %v", manager.filePath, tt.filePath)
			}
		})
	}
}

func TestManager_AddTask(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		wantErr bool
	}{
		{
			name:    "add valid task",
			title:   "Deploy v1.0.0",
			wantErr: false,
		},
		{
			name:    "add empty title task",
			title:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			manager := NewManager(filepath.Join(tmpDir, "task_list.md"))

			task, err := manager.AddTask(tt.title)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if task == nil {
					t.Error("AddTask() returned nil task")
					return
				}
				if task.Title != tt.title {
					t.Errorf("AddTask() task.Title = %v, want %v", task.Title, tt.title)
				}
			}
		})
	}
}

func TestManager_GetTask(t *testing.T) {
	tests := []struct {
		name      string
		taskTitle string
		setupFunc func(*Manager) string // returns task ID
		wantErr   bool
	}{
		{
			name:      "get existing task",
			taskTitle: "Test Task",
			setupFunc: func(m *Manager) string {
				task, _ := m.AddTask("Test Task")
				return task.Id
			},
			wantErr: false,
		},
		{
			name:      "get non-existent task",
			taskTitle: "",
			setupFunc: func(m *Manager) string {
				return "non-existent-id"
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			manager := NewManager(filepath.Join(tmpDir, "task_list.md"))

			taskId := tt.setupFunc(manager)
			task, err := manager.GetTask(taskId)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && task.Id != taskId {
				t.Errorf("GetTask() task.Id = %v, want %v", task.Id, taskId)
			}
		})
	}
}

func TestManager_UpdateTaskState(t *testing.T) {
	tests := []struct {
		name      string
		newState  State
		wantState State
	}{
		{
			name:      "update to in_progress",
			newState:  StateInProgress,
			wantState: StateInProgress,
		},
		{
			name:      "update to completed",
			newState:  StateCompleted,
			wantState: StateCompleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			manager := NewManager(filepath.Join(tmpDir, "task_list.md"))

			task, _ := manager.AddTask("Test Task")
			err := manager.UpdateTaskState(task.Id, tt.newState)
			if err != nil {
				t.Errorf("UpdateTaskState() error = %v", err)
				return
			}

			updated, _ := manager.GetTask(task.Id)
			if updated.State != tt.wantState {
				t.Errorf("UpdateTaskState() state = %v, want %v", updated.State, tt.wantState)
			}
		})
	}
}

func TestManager_ListTasks(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*Manager)
		wantCount int
	}{
		{
			name:      "list empty tasks",
			setupFunc: func(m *Manager) {},
			wantCount: 0,
		},
		{
			name: "list multiple tasks",
			setupFunc: func(m *Manager) {
				m.AddTask("Task 1")
				m.AddTask("Task 2")
				m.AddTask("Task 3")
			},
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			manager := NewManager(filepath.Join(tmpDir, "task_list.md"))
			tt.setupFunc(manager)

			tasks := manager.ListTasks()
			if len(tasks) != tt.wantCount {
				t.Errorf("ListTasks() count = %v, want %v", len(tasks), tt.wantCount)
			}
		})
	}
}

func TestManager_SaveAndLoad(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "save and load tasks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "task_list.md")

			// Create manager and add tasks
			manager1 := NewManager(filePath)
			task1, _ := manager1.AddTask("Task 1")
			task1.AddStep("Step 1")
			task1.AddStep("Step 2")
			manager1.UpdateTaskState(task1.Id, StateInProgress)

			manager1.AddTask("Task 2")

			// Save
			if err := manager1.Save(); err != nil {
				t.Errorf("Save() error = %v", err)
				return
			}

			// Verify file exists
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Error("Save() did not create file")
				return
			}

			// Load with new manager
			manager2 := NewManager(filePath)
			if err := manager2.Load(); err != nil {
				t.Errorf("Load() error = %v", err)
				return
			}

			tasks := manager2.ListTasks()
			if len(tasks) != 2 {
				t.Errorf("Load() task count = %v, want 2", len(tasks))
			}
		})
	}
}

func TestManager_CleanupCompleted(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(*Manager)
		olderThan     time.Duration
		wantRemaining int
	}{
		{
			name: "cleanup old completed tasks",
			setupFunc: func(m *Manager) {
				// Add completed task with old completion time
				task1, _ := m.AddTask("Old Completed Task")
				task1.UpdateState(StateCompleted)
				task1.CompleteAt = time.Now().Add(-48 * time.Hour).UnixMilli()

				// Add completed task with recent completion time
				task2, _ := m.AddTask("Recent Completed Task")
				task2.UpdateState(StateCompleted)
				task2.CompleteAt = time.Now().Add(-1 * time.Hour).UnixMilli()

				// Add pending task
				m.AddTask("Pending Task")
			},
			olderThan:     24 * time.Hour,
			wantRemaining: 2, // recent completed + pending
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			manager := NewManager(filepath.Join(tmpDir, "task_list.md"))
			tt.setupFunc(manager)

			manager.CleanupCompleted(tt.olderThan)

			tasks := manager.ListTasks()
			if len(tasks) != tt.wantRemaining {
				t.Errorf("CleanupCompleted() remaining = %v, want %v", len(tasks), tt.wantRemaining)
			}
		})
	}
}
