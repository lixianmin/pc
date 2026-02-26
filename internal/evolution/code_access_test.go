package evolution

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewCodeAccessor(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func() (string, []string)
		wantErr     bool
	}{
		{
			name: "create valid accessor",
			setupFunc: func() (string, []string) {
				tmpDir := t.TempDir()
				return tmpDir, []string{tmpDir}
			},
			wantErr: false,
		},
		{
			name: "create with empty base dir",
			setupFunc: func() (string, []string) {
				tmpDir := t.TempDir()
				return "", []string{tmpDir}
			},
			wantErr: true,
		},
		{
			name: "create with no allowed dirs",
			setupFunc: func() (string, []string) {
				tmpDir := t.TempDir()
				return tmpDir, []string{}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseDir, allowedDirs := tt.setupFunc()
			accessor, err := NewCodeAccessor(baseDir, allowedDirs)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCodeAccessor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && accessor == nil {
				t.Error("NewCodeAccessor() returned nil")
			}
		})
	}
}

func TestCodeAccessor_ScanDirectory(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() (string, func())
		wantCount int
		wantErr   bool
	}{
		{
			name: "scan directory with files",
			setupFunc: func() (string, func()) {
				tmpDir := t.TempDir()
				os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644)
				os.WriteFile(filepath.Join(tmpDir, "utils.go"), []byte("package main"), 0644)
				// Count: tmpDir (dir) + main.go + utils.go = 3
				return tmpDir, func() {}
			},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name: "scan empty directory",
			setupFunc: func() (string, func()) {
				tmpDir := t.TempDir()
				// Count: tmpDir (dir) = 1
				return tmpDir, func() {}
			},
			wantCount: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, cleanup := tt.setupFunc()
			defer cleanup()

			accessor, _ := NewCodeAccessor(dir, []string{dir})
			files, err := accessor.ScanDirectory(dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScanDirectory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(files) != tt.wantCount {
				t.Errorf("ScanDirectory() file count = %v, want %v", len(files), tt.wantCount)
			}
		})
	}
}

func TestCodeAccessor_ReadFile(t *testing.T) {
	tests := []struct {
		name      string
		fileName  string
		content   string
		wantErr   bool
	}{
		{
			name:     "read existing file",
			fileName: "test.go",
			content:  "package main\n\nfunc main() {}",
			wantErr:  false,
		},
		{
			name:     "read non-existent file",
			fileName: "nonexistent.go",
			content:  "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			accessor, _ := NewCodeAccessor(tmpDir, []string{tmpDir})

			if tt.content != "" {
				os.WriteFile(filepath.Join(tmpDir, tt.fileName), []byte(tt.content), 0644)
			}

			content, err := accessor.ReadFile(tt.fileName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && string(content) != tt.content {
				t.Errorf("ReadFile() content = %v, want %v", string(content), tt.content)
			}
		})
	}
}

func TestCodeAccessor_IsAllowed(t *testing.T) {
	tmpDir := t.TempDir()
	// Create subdirectory for testing
	subDir := filepath.Join(tmpDir, "project")
	os.MkdirAll(subDir, 0755)

	tests := []struct {
		name        string
		path        string
		wantAllowed bool
	}{
		{
			name:        "path in allowed directory",
			path:        filepath.Join(subDir, "main.go"),
			wantAllowed: true,
		},
		{
			name:        "path not in allowed directory",
			path:        "/etc/passwd",
			wantAllowed: false,
		},
		{
			name:        "path traverses outside allowed",
			path:        filepath.Join(subDir, "../../../etc/passwd"),
			wantAllowed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor, err := NewCodeAccessor(subDir, []string{subDir})
			if err != nil {
				t.Fatalf("NewCodeAccessor() error = %v", err)
			}
			allowed := accessor.IsAllowed(tt.path)
			if allowed != tt.wantAllowed {
				t.Errorf("IsAllowed() = %v, want %v", allowed, tt.wantAllowed)
			}
		})
	}
}

func TestCodeAccessor_ListGoFiles(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(string)
		wantCount int
		wantErr   bool
	}{
		{
			name: "list go files",
			setupFunc: func(dir string) {
				os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
				os.WriteFile(filepath.Join(dir, "utils.go"), []byte("package main"), 0644)
				os.WriteFile(filepath.Join(dir, "readme.md"), []byte("# README"), 0644)
			},
			wantCount: 2,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tt.setupFunc(tmpDir)

			accessor, _ := NewCodeAccessor(tmpDir, []string{tmpDir})
			files, err := accessor.ListGoFiles(tmpDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListGoFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(files) != tt.wantCount {
				t.Errorf("ListGoFiles() count = %v, want %v", len(files), tt.wantCount)
			}
		})
	}
}

func TestCodeAccessor_GetFileInfo(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		wantErr  bool
	}{
		{
			name:     "get info of existing file",
			fileName: "test.go",
			wantErr:  false,
		},
		{
			name:     "get info of non-existent file",
			fileName: "nonexistent.go",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			accessor, _ := NewCodeAccessor(tmpDir, []string{tmpDir})

			if tt.fileName == "test.go" {
				os.WriteFile(filepath.Join(tmpDir, tt.fileName), []byte("package main"), 0644)
			}

			info, err := accessor.GetFileInfo(tt.fileName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFileInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && info == nil {
				t.Error("GetFileInfo() returned nil info")
			}
		})
	}
}
