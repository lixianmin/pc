package file

import (
	"fmt"
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

type ReadResult struct {
	Content  string `json:"content"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type,omitempty"`
}

func (t *FileTool) Read(params ReadParams) (*ReadResult, error) {
	path := params.Path

	if err := t.validatePath(path); err != nil {
		return nil, err
	}

	if err := t.checkFileSize(path); err != nil {
		return nil, err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	mimeType := t.detectMimeType(path)

	return &ReadResult{
		Content:  string(content),
		Size:     info.Size(),
		MimeType: mimeType,
	}, nil
}

type WriteParams struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type WriteResult struct {
	Success bool   `json:"success"`
	Size    int    `json:"size"`
	Path    string `json:"path"`
}

func (t *FileTool) Write(params WriteParams) (*WriteResult, error) {
	path := params.Path
	content := params.Content

	if err := t.validatePath(path); err != nil {
		return nil, err
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return nil, err
	}

	return &WriteResult{
		Success: true,
		Size:    len(content),
		Path:    path,
	}, nil
}

type DeleteParams struct {
	Path string `json:"path"`
}

type DeleteResult struct {
	Success bool   `json:"success"`
	Path    string `json:"path"`
}

func (t *FileTool) Delete(params DeleteParams) (*DeleteResult, error) {
	path := params.Path

	if err := t.validatePath(path); err != nil {
		return nil, err
	}

	if err := os.Remove(path); err != nil {
		return nil, err
	}

	return &DeleteResult{
		Success: true,
		Path:    path,
	}, nil
}

type ListParams struct {
	Path string `json:"path"`
}

type ListResult struct {
	Files []FileInfo `json:"files"`
	Path  string     `json:"path"`
}

type FileInfo struct {
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size,omitempty"`
	Mode  string `json:"mode,omitempty"`
}

func (t *FileTool) List(params ListParams) (*ListResult, error) {
	path := params.Path
	if path == "" {
		path = "."
	}

	if err := t.validatePath(path); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, FileInfo{
			Name:  entry.Name(),
			IsDir: entry.IsDir(),
			Size:  info.Size(),
			Mode:  info.Mode().String(),
		})
	}

	return &ListResult{
		Files: files,
		Path:  path,
	}, nil
}

func (t *FileTool) validatePath(path string) error {
	expanded := os.ExpandEnv(path)

	absPath, err := filepath.Abs(expanded)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	if len(t.config.AllowedPaths) > 0 {
		allowed := false
		for _, allowedPath := range t.config.AllowedPaths {
			if strings.HasPrefix(absPath, os.ExpandEnv(allowedPath)) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("path not in allowed list: %s", absPath)
		}
	}

	return nil
}

func (t *FileTool) checkFileSize(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.Size() > t.config.MaxFileSize {
		return fmt.Errorf("file size %d exceeds maximum allowed size %d", info.Size(), t.config.MaxFileSize)
	}

	return nil
}

func (t *FileTool) detectMimeType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".txt":
		return "text/plain"
	case ".json":
		return "application/json"
	case ".md":
		return "text/markdown"
	case ".yml", ".yaml":
		return "text/yaml"
	case ".go":
		return "text/x-go"
	case ".py":
		return "text/x-python"
	case ".js":
		return "application/javascript"
	default:
		return "application/octet-stream"
	}
}
