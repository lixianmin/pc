package task

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
)

// State represents the state of a task.
type State string

const (
	// StatePending indicates the task is pending.
	StatePending State = "PENDING"
	// StateInProgress indicates the task is in progress.
	StateInProgress State = "IN_PROGRESS"
	// StateCompleted indicates the task is completed.
	StateCompleted State = "COMPLETED"
)

// Step represents a step in a task.
type Step struct {
	Title string `json:"title"`
	Done  bool   `json:"done"`
	Order int    `json:"order"`
}

// Task represents a task in the task list.
type Task struct {
	Id         string `json:"id"`
	Title      string `json:"title"`
	State      State  `json:"state"`
	Steps      []Step `json:"steps"`
	CreateAt   int64  `json:"create_at"`   // Creation time (Unix milliseconds)
	UpdateAt   int64  `json:"update_at"`   // Last update time (Unix milliseconds)
	CompleteAt int64  `json:"complete_at"` // Completion time (Unix milliseconds)
}

// NewTask creates a new task with the given title.
func NewTask(title string) (*Task, error) {
	if strings.TrimSpace(title) == "" {
		return nil, fmt.Errorf("task title cannot be empty")
	}

	id := ulid.Make().String()
	now := time.Now().UnixMilli()

	return &Task{
		Id:       id,
		Title:    title,
		State:    StatePending,
		Steps:    []Step{},
		CreateAt: now,
		UpdateAt: now,
	}, nil
}

// AddStep adds a step to the task.
func (my *Task) AddStep(title string) error {
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("step title cannot be empty")
	}

	my.Steps = append(my.Steps, Step{
		Title: title,
		Done:  false,
		Order: len(my.Steps) + 1,
	})
	return nil
}

// UpdateState updates the state of the task.
func (my *Task) UpdateState(state State) {
	my.State = state
	my.UpdateAt = time.Now().UnixMilli()
	if state == StateCompleted {
		my.CompleteAt = my.UpdateAt
	}
}

// Manager manages tasks and persists them to a file.
type Manager struct {
	tasks    map[string]*Task
	filePath string
}

// NewManager creates a new task manager.
func NewManager(filePath string) *Manager {
	return &Manager{
		tasks:    make(map[string]*Task),
		filePath: filePath,
	}
}

// AddTask creates and adds a new task.
func (my *Manager) AddTask(title string) (*Task, error) {
	task, err := NewTask(title)
	if err != nil {
		return nil, err
	}

	my.tasks[task.Id] = task
	return task, nil
}

// GetTask returns a task by ID.
func (my *Manager) GetTask(id string) (*Task, error) {
	task, ok := my.tasks[id]
	if !ok {
		return nil, fmt.Errorf("task not found: %s", id)
	}
	return task, nil
}

// UpdateTaskState updates the state of a task.
func (my *Manager) UpdateTaskState(id string, state State) error {
	task, ok := my.tasks[id]
	if !ok {
		return fmt.Errorf("task not found: %s", id)
	}

	task.UpdateState(state)
	return nil
}

// ListTasks returns all tasks.
func (my *Manager) ListTasks() []*Task {
	result := make([]*Task, 0, len(my.tasks))
	for _, task := range my.tasks {
		result = append(result, task)
	}
	return result
}

// DeleteTask deletes a task by ID.
func (my *Manager) DeleteTask(id string) error {
	if _, ok := my.tasks[id]; !ok {
		return fmt.Errorf("task not found: %s", id)
	}
	delete(my.tasks, id)
	return nil
}

// CleanupCompleted removes completed tasks older than the given duration.
func (my *Manager) CleanupCompleted(olderThan time.Duration) {
	cutoff := time.Now().Add(-olderThan).UnixMilli()
	for id, task := range my.tasks {
		if task.State == StateCompleted && task.CompleteAt < cutoff {
			delete(my.tasks, id)
		}
	}
}

// Save persists tasks to the file.
func (my *Manager) Save() error {
	// Ensure directory exists
	dir := filepath.Dir(my.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate markdown content
	content := my.generateMarkdown()

	// Write to file
	if err := os.WriteFile(my.filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Load loads tasks from the file.
func (my *Manager) Load() error {
	data, err := os.ReadFile(my.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet, start with empty tasks
			return nil
		}
		return fmt.Errorf("failed to read file: %w", err)
	}

	return my.parseMarkdown(string(data))
}

// generateMarkdown generates markdown content from tasks.
func (my *Manager) generateMarkdown() string {
	var sb strings.Builder
	sb.WriteString("# Task List\n\n")

	for _, task := range my.tasks {
		// Task header with state
		sb.WriteString(fmt.Sprintf("## [%s] %s\n", task.State, task.Title))
		sb.WriteString(fmt.Sprintf("- ID: %s\n", task.Id))
		sb.WriteString(fmt.Sprintf("- CreateAt: %d\n", task.CreateAt))
		sb.WriteString(fmt.Sprintf("- UpdateAt: %d\n", task.UpdateAt))
		if task.State == StateCompleted {
			sb.WriteString(fmt.Sprintf("- CompleteAt: %d\n", task.CompleteAt))
		}

		// Steps
		if len(task.Steps) > 0 {
			sb.WriteString("- Steps:\n")
			for _, step := range task.Steps {
				status := " "
				if step.Done {
					status = "x"
				}
				sb.WriteString(fmt.Sprintf("  - [%s] %s\n", status, step.Title))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// parseMarkdown parses markdown content into tasks.
func (my *Manager) parseMarkdown(content string) error {
	// Simple markdown parser
	lines := strings.Split(content, "\n")
	var currentTask *Task

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Parse task header: ## [STATE] Title
		if strings.HasPrefix(line, "## [") {
			if currentTask != nil {
				my.tasks[currentTask.Id] = currentTask
			}

			// Extract state and title
			endIdx := strings.Index(line, "]")
			if endIdx > 4 {
				stateStr := line[4:endIdx]
				title := strings.TrimSpace(line[endIdx+1:])

				currentTask = &Task{
					Title: title,
					State: State(stateStr),
					Steps: []Step{},
				}
			}
			continue
		}

		// Parse ID
		if strings.HasPrefix(line, "- ID:") {
			if currentTask != nil {
				currentTask.Id = strings.TrimSpace(line[5:])
			}
			continue
		}

		// Parse CreateAt
		if strings.HasPrefix(line, "- CreateAt:") {
			if currentTask != nil {
				fmt.Sscanf(strings.TrimSpace(line[11:]), "%d", &currentTask.CreateAt)
			}
			continue
		}

		// Parse UpdateAt
		if strings.HasPrefix(line, "- UpdateAt:") {
			if currentTask != nil {
				fmt.Sscanf(strings.TrimSpace(line[11:]), "%d", &currentTask.UpdateAt)
			}
			continue
		}

		// Parse CompleteAt
		if strings.HasPrefix(line, "- CompleteAt:") {
			if currentTask != nil {
				fmt.Sscanf(strings.TrimSpace(line[13:]), "%d", &currentTask.CompleteAt)
			}
			continue
		}

		// Parse steps
		if strings.HasPrefix(line, "- [") && currentTask != nil {
			// Step line: - [x] Step title
			done := line[3] == 'x'
			title := strings.TrimSpace(line[5:])
			currentTask.Steps = append(currentTask.Steps, Step{
				Title: title,
				Done:  done,
				Order: len(currentTask.Steps) + 1,
			})
		}
	}

	// Don't forget the last task
	if currentTask != nil && currentTask.Id != "" {
		my.tasks[currentTask.Id] = currentTask
	}

	return nil
}
