package tool

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileConfig struct {
	Timeout      time.Duration
	MaxFileSize  int64
	AllowedPaths []string
	BlockedPaths []string
}

type FileTool struct {
	config FileConfig
}

func NewFileTool(config FileConfig) *FileTool {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxFileSize == 0 {
		config.MaxFileSize = 10 * 1024 * 1024
	}
	return &FileTool{config: config}
}

func (t *FileTool) Name() string {
	return "file"
}

func (t *FileTool) Description() string {
	return "Read, write, list, and delete files"
}

type FileInfo struct {
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size"`
	Mode  string `json:"mode"`
}

func (t *FileTool) Execute(ctx context.Context, params map[string]any) (any, error) {
	action, _ := params["action"].(string)
	if action == "" {
		action = "read"
	}

	switch action {
	case "read":
		return t.read(params)
	case "write":
		return t.write(params)
	case "list":
		return t.list(params)
	case "delete":
		return t.delete(params)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

func (t *FileTool) read(params map[string]any) (any, error) {
	path, _ := params["path"].(string)
	if path == "" {
		return nil, fmt.Errorf("path is required")
	}

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

	return map[string]any{
		"content":   string(content),
		"size":      info.Size(),
		"mime_type": t.detectMimeType(path),
	}, nil
}

func (t *FileTool) write(params map[string]any) (any, error) {
	path, _ := params["path"].(string)
	content, _ := params["content"].(string)

	if path == "" {
		return nil, fmt.Errorf("path is required")
	}

	if err := t.validatePath(path); err != nil {
		return nil, err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return nil, err
	}

	return map[string]any{
		"success": true,
		"size":    len(content),
		"path":    path,
	}, nil
}

func (t *FileTool) list(params map[string]any) (any, error) {
	path, _ := params["path"].(string)
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

	return map[string]any{
		"files": files,
		"path":  path,
	}, nil
}

func (t *FileTool) delete(params map[string]any) (any, error) {
	path, _ := params["path"].(string)
	if path == "" {
		return nil, fmt.Errorf("path is required")
	}

	if err := t.validatePath(path); err != nil {
		return nil, err
	}

	if err := os.Remove(path); err != nil {
		return nil, err
	}

	return map[string]any{
		"success": true,
		"path":    path,
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
