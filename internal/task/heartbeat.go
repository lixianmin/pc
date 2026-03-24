package task

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// HeartbeatTask represents a periodic task
type HeartbeatTask struct {
	Name        string `json:"name"`
	Schedule    string `json:"schedule"`
	Description string `json:"description"`
	OnFailure   string `json:"on_failure"`
	LastRun     int64  `json:"last_run"`
	Enabled     bool   `json:"enabled"`
}

// IsDue checks if the task is due to run
func (my *HeartbeatTask) IsDue(now time.Time) bool {
	if my.LastRun == 0 {
		return true
	}

	lastRun := time.UnixMilli(my.LastRun)
	// Simple schedule check - for production, use a proper cron parser
	// For now, check if enough time has passed based on schedule format
	duration := my.parseSchedule()
	return now.After(lastRun.Add(duration))
}

// parseSchedule converts cron-like schedule to duration
// This is a simplified implementation
func (my *HeartbeatTask) parseSchedule() time.Duration {
	// Default to 1 hour
	if my.Schedule == "" {
		return time.Hour
	}

	// Parse simple schedules
	switch my.Schedule {
	case "@hourly", "0 * * * *":
		return time.Hour
	case "@daily", "0 0 * * *":
		return 24 * time.Hour
	case "@weekly", "0 0 * * 0":
		return 7 * 24 * time.Hour
	default:
		// Try to parse duration format like "5m", "1h"
		d, err := time.ParseDuration(my.Schedule)
		if err != nil {
			return time.Hour
		}
		return d
	}
}

// HeartbeatScheduler manages periodic tasks
type HeartbeatScheduler struct {
	mu      sync.Mutex
	tasks   []HeartbeatTask
	running bool
	stopCh  chan struct{}
}

// NewHeartbeatScheduler creates a new heartbeat scheduler
func NewHeartbeatScheduler() *HeartbeatScheduler {
	return &HeartbeatScheduler{
		tasks:  []HeartbeatTask{},
		stopCh: make(chan struct{}),
	}
}

// LoadFromFile loads heartbeat tasks from a markdown file
func (my *HeartbeatScheduler) LoadFromFile(filePath string) ([]HeartbeatTask, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []HeartbeatTask{}, nil
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	tasks, err := my.parseMarkdown(string(data))
	if err != nil {
		return nil, err
	}

	my.tasks = tasks
	return tasks, nil
}

// parseMarkdown parses heartbeat tasks from markdown content
func (my *HeartbeatScheduler) parseMarkdown(content string) ([]HeartbeatTask, error) {
	var tasks []HeartbeatTask
	lines := strings.Split(content, "\n")

	var currentTask *HeartbeatTask
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Parse task header: ## Task Name
		if strings.HasPrefix(line, "## ") {
			if currentTask != nil && currentTask.Schedule != "" {
				tasks = append(tasks, *currentTask)
			}
			currentTask = &HeartbeatTask{
				Name:    strings.TrimSpace(line[3:]),
				Enabled: true,
			}
			continue
		}

		if currentTask == nil {
			continue
		}

		// Parse schedule
		if strings.HasPrefix(line, "schedule:") {
			schedule := strings.TrimSpace(line[9:])
			// Remove quotes if present
			schedule = strings.Trim(schedule, `"'`)
			currentTask.Schedule = schedule
			continue
		}

		// Parse description
		if strings.HasPrefix(line, "description:") {
			currentTask.Description = strings.TrimSpace(line[12:])
			continue
		}

		// Parse onFailure
		if strings.HasPrefix(line, "onFailure:") || strings.HasPrefix(line, "on_failure:") {
			if strings.HasPrefix(line, "onFailure:") {
				currentTask.OnFailure = strings.TrimSpace(line[10:])
			} else {
				currentTask.OnFailure = strings.TrimSpace(line[11:])
			}
			continue
		}
	}

	// Don't forget the last task
	if currentTask != nil && currentTask.Schedule != "" {
		tasks = append(tasks, *currentTask)
	}

	return tasks, nil
}

// Start starts the scheduler
func (my *HeartbeatScheduler) Start(ctx context.Context) error {
	my.mu.Lock()
	defer my.mu.Unlock()
	if my.running {
		return fmt.Errorf("scheduler already running")
	}

	my.running = true
	go my.run(ctx)
	return nil
}

// run is the main scheduler loop
func (my *HeartbeatScheduler) run(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			my.mu.Lock()
			my.running = false
			my.mu.Unlock()
			return
		case <-my.stopCh:
			my.mu.Lock()
			my.running = false
			my.mu.Unlock()
			return
		case <-ticker.C:
			my.checkAndRunTasks()
		}
	}
}

// checkAndRunTasks checks for due tasks and runs them
func (my *HeartbeatScheduler) checkAndRunTasks() {
	now := time.Now()
	for i := range my.tasks {
		if my.tasks[i].Enabled && my.tasks[i].IsDue(now) {
			my.executeTask(&my.tasks[i])
		}
	}
}

// executeTask executes a single task
func (my *HeartbeatScheduler) executeTask(task *HeartbeatTask) {
	task.LastRun = time.Now().UnixMilli()
	// Task execution logic would go here
	// For now, just mark as run
}

// Stop stops the scheduler
func (my *HeartbeatScheduler) Stop() {
	my.mu.Lock()
	defer my.mu.Unlock()
	if my.running {
		my.running = false
		close(my.stopCh)
	}
}

// AddTask adds a new heartbeat task
func (my *HeartbeatScheduler) AddTask(task HeartbeatTask) error {
	if task.Schedule == "" {
		return fmt.Errorf("task schedule cannot be empty")
	}
	if task.Name == "" {
		return fmt.Errorf("task name cannot be empty")
	}

	my.tasks = append(my.tasks, task)
	return nil
}

// RemoveTask removes a task by name
func (my *HeartbeatScheduler) RemoveTask(name string) error {
	for i, task := range my.tasks {
		if task.Name == name {
			my.tasks = append(my.tasks[:i], my.tasks[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("task not found: %s", name)
}

// ListTasks returns all tasks
func (my *HeartbeatScheduler) ListTasks() []HeartbeatTask {
	result := make([]HeartbeatTask, len(my.tasks))
	copy(result, my.tasks)
	return result
}
