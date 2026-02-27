//go:build windows

package gateway

import (
	"syscall"
)

// isProcessRunning checks if a process with the given PID is running on Windows.
func isProcessRunning(pid int) bool {
	// OpenProcess with PROCESS_QUERY_LIMITED_INFORMATION
	handle, err := syscall.OpenProcess(syscall.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(handle)

	// Try to get exit code
	var exitCode uint32
	err = syscall.GetExitCodeProcess(handle, &exitCode)
	if err != nil {
		return false
	}

	// If exit code is STILL_ACTIVE (259), process is running
	return exitCode == 259
}
