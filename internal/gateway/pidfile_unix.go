//go:build darwin || linux || freebsd || openbsd || netbsd

package gateway

import "syscall"

// isProcessRunning checks if a process with the given PID is running.
func isProcessRunning(pid int) bool {
	// Send signal 0 to check if process exists
	err := syscall.Kill(pid, 0)
	return err == nil
}
