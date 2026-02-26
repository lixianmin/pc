package task

import (
	"context"
	"errors"
	"testing"
	"time"
)

// MockStepExecutor is a mock step executor for testing
type MockStepExecutor struct {
	ExecuteFunc func(ctx context.Context, step *Step) error
}

func (m *MockStepExecutor) Execute(ctx context.Context, step *Step) error {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, step)
	}
	return nil
}

func TestNewTaskExecutor(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "create new executor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewTaskExecutor()
			if e == nil {
				t.Error("NewTaskExecutor() returned nil")
			}
		})
	}
}

func TestTaskExecutor_Execute(t *testing.T) {
	tests := []struct {
		name      string
		taskSetup func() *Task
		executor  StepExecutor
		wantState State
		wantErr   bool
	}{
		{
			name: "execute task with all steps",
			taskSetup: func() *Task {
				task, _ := NewTask("Test Task")
				task.AddStep("Step 1")
				task.AddStep("Step 2")
				return task
			},
			executor: &MockStepExecutor{
				ExecuteFunc: func(ctx context.Context, step *Step) error {
					step.Done = true
					return nil
				},
			},
			wantState: StateCompleted,
			wantErr:   false,
		},
		{
			name: "execute task with failing step",
			taskSetup: func() *Task {
				task, _ := NewTask("Test Task")
				task.AddStep("Step 1")
				task.AddStep("Step 2")
				return task
			},
			executor: &MockStepExecutor{
				ExecuteFunc: func(ctx context.Context, step *Step) error {
					if step.Title == "Step 1" {
						return errors.New("step 1 failed")
					}
					step.Done = true
					return nil
				},
			},
			wantState: StateInProgress,
			wantErr:   true,
		},
		{
			name: "execute task with no steps",
			taskSetup: func() *Task {
				task, _ := NewTask("Test Task")
				return task
			},
			executor: &MockStepExecutor{
				ExecuteFunc: func(ctx context.Context, step *Step) error {
					step.Done = true
					return nil
				},
			},
			wantState: StateCompleted,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewTaskExecutor()
			task := tt.taskSetup()

			err := e.Execute(context.Background(), task, tt.executor)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if task.State != tt.wantState {
				t.Errorf("Execute() task.State = %v, want %v", task.State, tt.wantState)
			}
		})
	}
}

func TestTaskExecutor_ExecuteStep(t *testing.T) {
	tests := []struct {
		name     string
		step     *Step
		executor StepExecutor
		wantDone bool
		wantErr  bool
	}{
		{
			name: "execute successful step",
			step: &Step{Title: "Step 1", Done: false},
			executor: &MockStepExecutor{
				ExecuteFunc: func(ctx context.Context, step *Step) error {
					step.Done = true
					return nil
				},
			},
			wantDone: true,
			wantErr:  false,
		},
		{
			name:     "execute already done step",
			step:     &Step{Title: "Step 1", Done: true},
			executor: &MockStepExecutor{},
			wantDone: true,
			wantErr:  false,
		},
		{
			name: "execute failing step",
			step: &Step{Title: "Step 1", Done: false},
			executor: &MockStepExecutor{
				ExecuteFunc: func(ctx context.Context, step *Step) error {
					return errors.New("execution failed")
				},
			},
			wantDone: false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewTaskExecutor()
			err := e.ExecuteStep(context.Background(), tt.step, tt.executor)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteStep() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.step.Done != tt.wantDone {
				t.Errorf("ExecuteStep() step.Done = %v, want %v", tt.step.Done, tt.wantDone)
			}
		})
	}
}

func TestTaskExecutor_ExecuteWithRetry(t *testing.T) {
	tests := []struct {
		name         string
		attempts     int
		executorFunc func() StepExecutor
		wantErr      bool
	}{
		{
			name:     "succeed on first attempt",
			attempts: 3,
			executorFunc: func() StepExecutor {
				return &MockStepExecutor{
					ExecuteFunc: func(ctx context.Context, step *Step) error {
						step.Done = true
						return nil
					},
				}
			},
			wantErr: false,
		},
		{
			name:     "succeed on second attempt",
			attempts: 3,
			executorFunc: func() StepExecutor {
				callCount := 0
				return &MockStepExecutor{
					ExecuteFunc: func(ctx context.Context, step *Step) error {
						callCount++
						if callCount < 2 {
							return errors.New("temporary failure")
						}
						step.Done = true
						return nil
					},
				}
			},
			wantErr: false,
		},
		{
			name:     "fail after all retries",
			attempts: 2,
			executorFunc: func() StepExecutor {
				return &MockStepExecutor{
					ExecuteFunc: func(ctx context.Context, step *Step) error {
						return errors.New("persistent failure")
					},
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewTaskExecutor()
			task, _ := NewTask("Test Task")
			task.AddStep("Step 1")

			executor := tt.executorFunc()
			err := e.ExecuteWithRetry(context.Background(), task, executor, tt.attempts)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteWithRetry() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTaskExecutor_GetProgress(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*Task)
		wantTotal int
		wantDone  int
		wantPct   float64
	}{
		{
			name: "no steps",
			setupFunc: func(task *Task) {
				// No steps added
			},
			wantTotal: 0,
			wantDone:  0,
			wantPct:   0,
		},
		{
			name: "all pending",
			setupFunc: func(task *Task) {
				task.AddStep("Step 1")
				task.AddStep("Step 2")
				task.AddStep("Step 3")
			},
			wantTotal: 3,
			wantDone:  0,
			wantPct:   0,
		},
		{
			name: "half done",
			setupFunc: func(task *Task) {
				task.AddStep("Step 1")
				task.AddStep("Step 2")
				task.Steps[0].Done = true
			},
			wantTotal: 2,
			wantDone:  1,
			wantPct:   50,
		},
		{
			name: "all done",
			setupFunc: func(task *Task) {
				task.AddStep("Step 1")
				task.AddStep("Step 2")
				task.Steps[0].Done = true
				task.Steps[1].Done = true
			},
			wantTotal: 2,
			wantDone:  2,
			wantPct:   100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, _ := NewTask("Test Task")
			tt.setupFunc(task)

			progress := GetProgress(task)
			if progress.Total != tt.wantTotal {
				t.Errorf("GetProgress() Total = %v, want %v", progress.Total, tt.wantTotal)
			}
			if progress.Done != tt.wantDone {
				t.Errorf("GetProgress() Done = %v, want %v", progress.Done, tt.wantDone)
			}
			if progress.Percentage != tt.wantPct {
				t.Errorf("GetProgress() Percentage = %v, want %v", progress.Percentage, tt.wantPct)
			}
		})
	}
}

func TestTaskExecutor_ExecuteWithTimeout(t *testing.T) {
	tests := []struct {
		name        string
		timeout     time.Duration
		executorFunc func() StepExecutor
		wantErr     bool
	}{
		{
			name:    "complete before timeout",
			timeout: 5 * time.Second,
			executorFunc: func() StepExecutor {
				return &MockStepExecutor{
					ExecuteFunc: func(ctx context.Context, step *Step) error {
						step.Done = true
						return nil
					},
				}
			},
			wantErr: false,
		},
		{
			name:    "timeout exceeded",
			timeout: 100 * time.Millisecond,
			executorFunc: func() StepExecutor {
				return &MockStepExecutor{
					ExecuteFunc: func(ctx context.Context, step *Step) error {
						time.Sleep(200 * time.Millisecond)
						step.Done = true
						return nil
					},
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewTaskExecutor()
			task, _ := NewTask("Test Task")
			task.AddStep("Step 1")

			executor := tt.executorFunc()
			err := e.ExecuteWithTimeout(context.Background(), task, executor, tt.timeout)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteWithTimeout() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
