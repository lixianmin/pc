package gateway

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPidfile_ReadWrite(t *testing.T) {
	tests := []struct {
		name    string
		pid     int
		wantErr bool
	}{
		{
			name:    "write and read valid pid",
			pid:     12345,
			wantErr: false,
		},
		{
			name:    "write and read pid 1",
			pid:     1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			pidfilePath := filepath.Join(tmpDir, "test.pid")
			pf := NewPidfile(pidfilePath)

			// Test Write
			err := pf.Write(tt.pid)
			if (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify file exists
			if !pf.Exists() {
				t.Error("Exists() = false, want true after Write")
			}

			// Test Read
			got, err := pf.Read()
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.pid {
				t.Errorf("Read() = %v, want %v", got, tt.pid)
			}
		})
	}
}

func TestPidfile_Read_NotExist(t *testing.T) {
	tmpDir := t.TempDir()
	pidfilePath := filepath.Join(tmpDir, "nonexistent.pid")
	pf := NewPidfile(pidfilePath)

	_, err := pf.Read()
	if err == nil {
		t.Error("Read() should return error for non-existent file")
	}
}

func TestPidfile_Read_InvalidContent(t *testing.T) {
	tmpDir := t.TempDir()
	pidfilePath := filepath.Join(tmpDir, "invalid.pid")

	// Write invalid content
	if err := os.WriteFile(pidfilePath, []byte("not-a-number"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	pf := NewPidfile(pidfilePath)
	_, err := pf.Read()
	if err == nil {
		t.Error("Read() should return error for invalid content")
	}
}

func TestPidfile_Remove(t *testing.T) {
	tmpDir := t.TempDir()
	pidfilePath := filepath.Join(tmpDir, "test.pid")
	pf := NewPidfile(pidfilePath)

	// Write then remove
	if err := pf.Write(12345); err != nil {
		t.Fatalf("Write() failed: %v", err)
	}

	if !pf.Exists() {
		t.Fatal("File should exist after Write")
	}

	if err := pf.Remove(); err != nil {
		t.Errorf("Remove() error = %v", err)
	}

	if pf.Exists() {
		t.Error("File should not exist after Remove")
	}

	// Remove non-existent file should not error
	if err := pf.Remove(); err != nil {
		t.Errorf("Remove() on non-existent file should not error: %v", err)
	}
}

func TestPidfile_Exists(t *testing.T) {
	tmpDir := t.TempDir()
	pidfilePath := filepath.Join(tmpDir, "test.pid")
	pf := NewPidfile(pidfilePath)

	if pf.Exists() {
		t.Error("Exists() = true for non-existent file")
	}

	if err := pf.Write(12345); err != nil {
		t.Fatalf("Write() failed: %v", err)
	}

	if !pf.Exists() {
		t.Error("Exists() = false for existing file")
	}
}

func TestPidfile_Cleanup(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(pf *Pidfile) error
		wantErr   bool
		wantExist bool
	}{
		{
			name: "cleanup non-existent file",
			setup: func(pf *Pidfile) error {
				return nil
			},
			wantErr:   false,
			wantExist: false,
		},
		{
			name: "cleanup stale pidfile with invalid pid",
			setup: func(pf *Pidfile) error {
				// Write a pid that is definitely not running
				return pf.Write(99999)
			},
			wantErr:   false,
			wantExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			pidfilePath := filepath.Join(tmpDir, "test.pid")
			pf := NewPidfile(pidfilePath)

			if err := tt.setup(pf); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			err := pf.Cleanup()
			if (err != nil) != tt.wantErr {
				t.Errorf("Cleanup() error = %v, wantErr %v", err, tt.wantErr)
			}

			exists := pf.Exists()
			if exists != tt.wantExist {
				t.Errorf("Exists() = %v, want %v", exists, tt.wantExist)
			}
		})
	}
}

func TestPidfile_IsRunning_CurrentProcess(t *testing.T) {
	tmpDir := t.TempDir()
	pidfilePath := filepath.Join(tmpDir, "test.pid")
	pf := NewPidfile(pidfilePath)

	// Write current process pid
	currentPid := os.Getpid()
	if err := pf.Write(currentPid); err != nil {
		t.Fatalf("Write() failed: %v", err)
	}

	// Current process should be running
	if !pf.IsRunning() {
		t.Error("IsRunning() = false for current process")
	}
}

func TestPidfile_IsRunning_InvalidPid(t *testing.T) {
	tmpDir := t.TempDir()
	pidfilePath := filepath.Join(tmpDir, "test.pid")
	pf := NewPidfile(pidfilePath)

	// Write a very high pid that definitely doesn't exist
	if err := pf.Write(999999); err != nil {
		t.Fatalf("Write() failed: %v", err)
	}

	// This pid should not be running
	if pf.IsRunning() {
		t.Error("IsRunning() = true for non-existent pid")
	}
}

func TestNewPidfile(t *testing.T) {
	path := "/tmp/test.pid"
	pf := NewPidfile(path)

	if pf == nil {
		t.Fatal("NewPidfile() returned nil")
	}

	if pf.GetPath() != path {
		t.Errorf("GetPath() = %v, want %v", pf.GetPath(), path)
	}
}
