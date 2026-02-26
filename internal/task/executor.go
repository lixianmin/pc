package task

import (
	"context"
	"fmt"
	"time"
)

// StepExecutor is the interface for executing a single step
type StepExecutor interface {
	Execute(ctx context.Context, step *Step) error
}

// TaskExecutor handles task execution with retry and progress tracking
type TaskExecutor struct{}

// NewTaskExecutor creates a new task executor
func NewTaskExecutor() *TaskExecutor {
	return &TaskExecutor{}
}

// Execute executes all steps of a task in order
func (my *TaskExecutor) Execute(ctx context.Context, task *Task, executor StepExecutor) error {
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}

	task.UpdateState(StateInProgress)

	for i := range task.Steps {
		step := &task.Steps[i]
		if step.Done {
			continue
		}

		if err := my.ExecuteStep(ctx, step, executor); err != nil {
			return fmt.Errorf("failed to execute step '%s': %w", step.Title, err)
		}
	}

	// Check if all steps are done
	allDone := true
	for _, step := range task.Steps {
		if !step.Done {
			allDone = false
			break
		}
	}

	if allDone {
		task.UpdateState(StateCompleted)
	}

	return nil
}

// ExecuteStep executes a single step
func (my *TaskExecutor) ExecuteStep(ctx context.Context, step *Step, executor StepExecutor) error {
	if step.Done {
		return nil
	}

	// Check context cancellation before execution
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if err := executor.Execute(ctx, step); err != nil {
		return err
	}

	// Check context cancellation after execution
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return nil
}

// ExecuteWithRetry executes a task with retry on failure
func (my *TaskExecutor) ExecuteWithRetry(ctx context.Context, task *Task, executor StepExecutor, maxAttempts int) error {
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := my.Execute(ctx, task, executor); err != nil {
			lastErr = err
			if attempt < maxAttempts {
				// Wait before retry with exponential backoff
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}
			return fmt.Errorf("failed after %d attempts: %w", maxAttempts, lastErr)
		}
		return nil
	}

	return lastErr
}

// ExecuteWithTimeout executes a task with a timeout
func (my *TaskExecutor) ExecuteWithTimeout(ctx context.Context, task *Task, executor StepExecutor, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return my.Execute(ctx, task, executor)
}

// Progress represents the execution progress of a task
type Progress struct {
	Total      int
	Done       int
	Percentage float64
}

// GetProgress returns the progress of a task
func GetProgress(task *Task) Progress {
	total := len(task.Steps)
	done := 0
	for _, step := range task.Steps {
		if step.Done {
			done++
		}
	}

	percentage := 0.0
	if total > 0 {
		percentage = float64(done) * 100 / float64(total)
	}

	return Progress{
		Total:      total,
		Done:       done,
		Percentage: percentage,
	}
}
