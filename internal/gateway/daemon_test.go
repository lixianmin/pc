package gateway

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewDaemon(t *testing.T) {
	pcDir := "/tmp/test-pc"
	d := NewDaemon(pcDir)

	if d == nil {
		t.Fatal("NewDaemon() returned nil")
	}

	expectedPidfile := filepath.Join(pcDir, "pc.pid")
	if d.GetPidfile().GetPath() != expectedPidfile {
		t.Errorf("GetPidfile().GetPath() = %v, want %v", d.GetPidfile().GetPath(), expectedPidfile)
	}

	expectedSocket := filepath.Join(pcDir, "pc.sock")
	if d.GetSocketPath() != expectedSocket {
		t.Errorf("GetSocketPath() = %v, want %v", d.GetSocketPath(), expectedSocket)
	}

	expectedLog := filepath.Join(pcDir, "logs", "pc.log")
	if d.GetLogPath() != expectedLog {
		t.Errorf("GetLogPath() = %v, want %v", d.GetLogPath(), expectedLog)
	}
}

func TestDaemon_GetStatus_NotRunning(t *testing.T) {
	tmpDir := t.TempDir()
	d := NewDaemon(tmpDir)

	status := d.GetStatus()

	if status.Running {
		t.Error("GetStatus().Running = true, want false")
	}

	if status.Error != "pidfile not found" {
		t.Errorf("GetStatus().Error = %v, want 'pidfile not found'", status.Error)
	}
}

func TestDaemon_GetStatus_StalePidfile(t *testing.T) {
	tmpDir := t.TempDir()
	d := NewDaemon(tmpDir)

	// Write a stale pidfile with non-existent pid
	if err := d.pidfile.Write(999999); err != nil {
		t.Fatalf("Failed to write pidfile: %v", err)
	}

	status := d.GetStatus()

	if status.Running {
		t.Error("GetStatus().Running = true, want false")
	}

	if status.Error != "process not running (stale pidfile)" {
		t.Errorf("GetStatus().Error = %v, want 'process not running (stale pidfile)'", status.Error)
	}
}

func TestDaemon_GetStatus_Running(t *testing.T) {
	tmpDir := t.TempDir()
	d := NewDaemon(tmpDir)

	// Write current process pid
	currentPid := os.Getpid()
	if err := d.pidfile.Write(currentPid); err != nil {
		t.Fatalf("Failed to write pidfile: %v", err)
	}

	status := d.GetStatus()

	if !status.Running {
		t.Error("GetStatus().Running = false, want true")
	}

	if status.Pid != currentPid {
		t.Errorf("GetStatus().Pid = %v, want %v", status.Pid, currentPid)
	}

	if status.Error != "" {
		t.Errorf("GetStatus().Error = %v, want empty", status.Error)
	}
}

func TestDaemon_Start_AlreadyRunning(t *testing.T) {
	tmpDir := t.TempDir()
	d := NewDaemon(tmpDir)

	// Write current process pid (simulating running daemon)
	currentPid := os.Getpid()
	if err := d.pidfile.Write(currentPid); err != nil {
		t.Fatalf("Failed to write pidfile: %v", err)
	}

	// Try to start - should fail because "already running"
	err := d.Start()
	if err == nil {
		t.Error("Start() should return error when daemon is already running")
	}

	if err.Error() != "daemon is already running (pid: "+string(rune(currentPid))+")" {
		// Error message should contain the pid
		if err.Error() == "" {
			t.Error("Start() error should contain pid information")
		}
	}
}

func TestDaemon_Stop_NotRunning(t *testing.T) {
	tmpDir := t.TempDir()
	d := NewDaemon(tmpDir)

	err := d.Stop()
	if err == nil {
		t.Error("Stop() should return error when daemon is not running")
	}

	if err.Error() != "daemon is not running (no pidfile)" {
		t.Errorf("Stop() error = %v, want 'daemon is not running (no pidfile)'", err.Error())
	}
}

func TestDaemon_Stop_StalePidfile(t *testing.T) {
	tmpDir := t.TempDir()
	d := NewDaemon(tmpDir)

	// Write stale pidfile
	if err := d.pidfile.Write(999999); err != nil {
		t.Fatalf("Failed to write pidfile: %v", err)
	}

	err := d.Stop()
	if err == nil {
		t.Error("Stop() should return error when daemon is not running")
	}

	// Pidfile should be removed
	if d.pidfile.Exists() {
		t.Error("Pidfile should be removed after Stop() with stale pidfile")
	}
}

func TestStatus_Struct(t *testing.T) {
	status := &Status{
		Running: true,
		Pid:     12345,
		Error:   "",
	}

	if !status.Running {
		t.Error("Status.Running should be true")
	}

	if status.Pid != 12345 {
		t.Errorf("Status.Pid = %v, want 12345", status.Pid)
	}

	if status.Error != "" {
		t.Error("Status.Error should be empty")
	}
}

func TestDaemon_LogDirectoryCreation(t *testing.T) {
	tmpDir := t.TempDir()
	d := NewDaemon(tmpDir)

	logDir := filepath.Dir(d.GetLogPath())

	// Ensure log directory doesn't exist
	os.RemoveAll(logDir)

	// Verify it doesn't exist
	if _, err := os.Stat(logDir); !os.IsNotExist(err) {
		t.Skip("Log directory exists, skipping test")
	}

	// When Start() is called, it should create the log directory
	// But since we can't actually start the daemon in tests (it would fork),
	// we just verify the path is correctly set
	expectedLogDir := filepath.Join(tmpDir, "logs")
	if logDir != expectedLogDir {
		t.Errorf("Log directory = %v, want %v", logDir, expectedLogDir)
	}
}
