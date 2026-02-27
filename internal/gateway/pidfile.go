package gateway

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Pidfile manages the pidfile for the daemon.
type Pidfile struct {
	path string
}

// NewPidfile creates a new pidfile manager.
func NewPidfile(path string) *Pidfile {
	return &Pidfile{path: path}
}

// GetPath returns the pidfile path.
func (my *Pidfile) GetPath() string {
	return my.path
}

// Exists checks if the pidfile exists.
func (my *Pidfile) Exists() bool {
	_, err := os.Stat(my.path)
	return err == nil
}

// Read reads the pid from the pidfile.
func (my *Pidfile) Read() (int, error) {
	data, err := os.ReadFile(my.path)
	if err != nil {
		return 0, fmt.Errorf("failed to read pidfile: %w", err)
	}

	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return 0, fmt.Errorf("invalid pid in pidfile: %w", err)
	}

	return pid, nil
}

// Write writes the pid to the pidfile.
func (my *Pidfile) Write(pid int) error {
	// Ensure directory exists
	dir := filepath.Dir(my.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create pidfile directory: %w", err)
	}

	data := strconv.Itoa(pid) + "\n"
	if err := os.WriteFile(my.path, []byte(data), 0644); err != nil {
		return fmt.Errorf("failed to write pidfile: %w", err)
	}

	return nil
}

// Remove removes the pidfile.
func (my *Pidfile) Remove() error {
	if err := os.Remove(my.path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove pidfile: %w", err)
	}
	return nil
}

// IsRunning checks if the process with the pid in the pidfile is running.
func (my *Pidfile) IsRunning() bool {
	pid, err := my.Read()
	if err != nil {
		return false
	}

	// On Unix, sending signal 0 checks if process exists
	return isProcessRunning(pid)
}

// Cleanup removes stale pidfile if the process is not running.
func (my *Pidfile) Cleanup() error {
	if !my.Exists() {
		return nil
	}

	if my.IsRunning() {
		return fmt.Errorf("daemon is already running")
	}

	// Process not running, remove stale pidfile
	return my.Remove()
}
