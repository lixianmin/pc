package evolution

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// CodeAccessor provides controlled access to code files
type CodeAccessor struct {
	baseDir     string
	allowedDirs []string
}

// FileInfo represents information about a code file
type FileInfo struct {
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	ModTime int64  `json:"mod_time"`
	IsDir   bool   `json:"is_dir"`
}

// NewCodeAccessor creates a new code accessor with security boundaries
func NewCodeAccessor(baseDir string, allowedDirs []string) (*CodeAccessor, error) {
	if strings.TrimSpace(baseDir) == "" {
		return nil, fmt.Errorf("base directory cannot be empty")
	}

	if len(allowedDirs) == 0 {
		return nil, fmt.Errorf("at least one allowed directory must be specified")
	}

	// Ensure base directory exists
	if _, err := os.Stat(baseDir); err != nil {
		return nil, fmt.Errorf("base directory does not exist: %w", err)
	}

	return &CodeAccessor{
		baseDir:     baseDir,
		allowedDirs: allowedDirs,
	}, nil
}

// ScanDirectory scans a directory for files
func (my *CodeAccessor) ScanDirectory(dir string) ([]FileInfo, error) {
	if !my.IsAllowed(dir) {
		return nil, fmt.Errorf("access denied: %s", dir)
	}

	var files []FileInfo
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		files = append(files, FileInfo{
			Path:    path,
			Size:    info.Size(),
			ModTime: info.ModTime().UnixMilli(),
			IsDir:   d.IsDir(),
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan directory: %w", err)
	}

	return files, nil
}

// ReadFile reads a file's content
func (my *CodeAccessor) ReadFile(filePath string) ([]byte, error) {
	// Resolve relative paths
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(my.baseDir, filePath)
	}

	if !my.IsAllowed(filePath) {
		return nil, fmt.Errorf("access denied: %s", filePath)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return content, nil
}

// IsAllowed checks if a path is within allowed directories
func (my *CodeAccessor) IsAllowed(path string) bool {
	// Clean the path to resolve any .. or .
	cleanPath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return false
	}

	for _, allowedDir := range my.allowedDirs {
		allowedAbs, err := filepath.Abs(filepath.Clean(allowedDir))
		if err != nil {
			continue
		}

		// Check if path is within allowed directory
		if strings.HasPrefix(cleanPath, allowedAbs) {
			return true
		}
	}

	return false
}

// ListGoFiles lists all Go files in a directory
func (my *CodeAccessor) ListGoFiles(dir string) ([]string, error) {
	if !my.IsAllowed(dir) {
		return nil, fmt.Errorf("access denied: %s", dir)
	}

	var files []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".go") {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list Go files: %w", err)
	}

	return files, nil
}

// GetFileInfo returns information about a file
func (my *CodeAccessor) GetFileInfo(filePath string) (*FileInfo, error) {
	// Resolve relative paths
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(my.baseDir, filePath)
	}

	if !my.IsAllowed(filePath) {
		return nil, fmt.Errorf("access denied: %s", filePath)
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &FileInfo{
		Path:    filePath,
		Size:    info.Size(),
		ModTime: info.ModTime().UnixMilli(),
		IsDir:   info.IsDir(),
	}, nil
}
