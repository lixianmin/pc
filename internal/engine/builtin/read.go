package builtin

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ReadTool struct{}

func NewReadTool() *ReadTool {
	return &ReadTool{}
}

func (my *ReadTool) Name() string {
	return "read"
}

func (my *ReadTool) Execute(ctx context.Context, params map[string]any) (string, error) {
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return "", fmt.Errorf("missing required parameter: path")
	}

	offset := 1
	if v, ok := params["offset"].(float64); ok && v > 0 {
		offset = int(v)
	}

	limit := 2000
	if v, ok := params["limit"].(float64); ok && v > 0 {
		limit = int(v)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("file not found: %s", path)
		}
		return "", fmt.Errorf("failed to stat file: %w", err)
	}

	if info.IsDir() {
		return "", fmt.Errorf("is a directory: %s", path)
	}

	file, err := os.Open(absPath)
	if err != nil {
		return "", fmt.Errorf("permission denied: %s", path)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		if lineNum >= offset && (limit <= 0 || len(lines) < limit) {
			lines = append(lines, scanner.Text())
		}
	}

	if scanner.Err() != nil {
		return "", fmt.Errorf("error reading file: %w", scanner.Err())
	}

	return strings.Join(lines, "\n"), nil
}
