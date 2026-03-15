package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	Timeout          time.Duration
	ConfirmDangerous bool
	MaxFileSize      int64
	AllowedPaths     []string
	BlockedPaths     []string
}

type FileTool struct {
	config Config
}

func NewFileTool(config Config) *FileTool {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxFileSize == 0 {
		config.MaxFileSize = 10 * 1024 * 1024
	}
	return &FileTool{config: config}
}

type ReadParams struct {
	Path string `json:"path"`
}

type WriteParams struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type DeleteParams struct {
	Path string `json:"path"`
}

type ListParams struct {
	Path string `json:"path"`
}

type ReadResult struct {
	Content  string `json:"content"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type,omitempty"`
}

type WriteResult struct {
	Success bool   `json:"success"`
	Size    int    `json:"size"`
	Path    string `json:"path"`
}

type DeleteResult struct {
	Success bool   `json:"success"`
	Path    string `json:"path"`
}

type ListResult struct {
	Files []FileInfo `json:"files"`
	Path  string     `json:"path"`
}

type FileInfo struct {
	Name    string `json:"name"`
	IsDir   bool   `json:"is_dir"`
	Size    int64  `json:"size"`
	ModTime string `json:"mod_time"`
}

func (my *FileTool) Read(params ReadParams) (*ReadResult, error) {
	if params.Path == "" {
		return nil, fmt.Errorf("path is required")
	}

	absPath, err := my.validatePath(params.Path)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory, use 'list' action instead")
	}

	if info.Size() > my.config.MaxFileSize {
		return nil, fmt.Errorf("file too large: %d bytes (max: %d)", info.Size(), my.config.MaxFileSize)
	}

	file, err := os.Open(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return &ReadResult{
		Content:  string(content),
		Size:     info.Size(),
		MimeType: my.detectMimeType(absPath),
	}, nil
}

func (my *FileTool) Write(params WriteParams) (*WriteResult, error) {
	if params.Path == "" {
		return nil, fmt.Errorf("path is required")
	}

	absPath, err := my.validatePath(params.Path)
	if err != nil {
		return nil, err
	}

	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(absPath, []byte(params.Content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return &WriteResult{
		Success: true,
		Size:    len(params.Content),
		Path:    absPath,
	}, nil
}

func (my *FileTool) Delete(params DeleteParams) (*DeleteResult, error) {
	if params.Path == "" {
		return nil, fmt.Errorf("path is required")
	}

	absPath, err := my.validatePath(params.Path)
	if err != nil {
		return nil, err
	}

	if my.config.ConfirmDangerous {
		if my.isDangerousPath(absPath) {
			return nil, fmt.Errorf("deleting this path requires confirmation")
		}
	}

	if err := os.Remove(absPath); err != nil {
		return nil, fmt.Errorf("failed to delete file: %w", err)
	}

	return &DeleteResult{
		Success: true,
		Path:    absPath,
	}, nil
}

func (my *FileTool) List(params ListParams) (*ListResult, error) {
	if params.Path == "" {
		params.Path = "."
	}

	absPath, err := my.validatePath(params.Path)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory")
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	files := make([]FileInfo, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, FileInfo{
			Name:    entry.Name(),
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime().Format(time.RFC3339),
		})
	}

	return &ListResult{
		Files: files,
		Path:  absPath,
	}, nil
}

func (my *FileTool) validatePath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	if len(my.config.BlockedPaths) > 0 {
		for _, blocked := range my.config.BlockedPaths {
			if strings.HasPrefix(absPath, blocked) || strings.Contains(absPath, blocked) {
				return "", fmt.Errorf("access to path is blocked: %s", blocked)
			}
		}
	}

	if len(my.config.AllowedPaths) > 0 {
		allowed := false
		for _, allowedPath := range my.config.AllowedPaths {
			expanded := os.ExpandEnv(allowedPath)
			if strings.HasPrefix(absPath, expanded) {
				allowed = true
				break
			}
		}
		if !allowed {
			return "", fmt.Errorf("path not in allowed list: %s", absPath)
		}
	}

	return absPath, nil
}

func (my *FileTool) isDangerousPath(path string) bool {
	dangerousPaths := []string{
		"/etc/passwd",
		"/etc/shadow",
		"/.ssh",
		"/.gnupg",
		"/etc",
		"/usr",
		"/bin",
		"/sbin",
	}

	for _, dangerous := range dangerousPaths {
		if strings.HasPrefix(path, dangerous) {
			return true
		}
	}
	return false
}

func (my *FileTool) detectMimeType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".txt":
		return "text/plain"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".go":
		return "text/x-go"
	case ".py":
		return "text/x-python"
	case ".md":
		return "text/markdown"
	case ".yml", ".yaml":
		return "text/yaml"
	case ".pdf":
		return "application/pdf"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	default:
		return "application/octet-stream"
	}
}

func (my *FileTool) SetConfig(config Config) {
	my.config = config
}
