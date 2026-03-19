package builtin

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lixianmin/pc/pkg/types"
)

type GlobTool struct{}

func NewGlobTool() *GlobTool {
	return &GlobTool{}
}

func (my *GlobTool) Name() string {
	return "glob"
}

func (my *GlobTool) Description() string {
	return "File name pattern matching"
}

func (my *GlobTool) Parameters() map[string]types.ParamSchema {
	return map[string]types.ParamSchema{
		"pattern": {
			Type:        "string",
			Required:    true,
			Description: "The glob pattern (e.g., *.go)",
		},
		"path": {
			Type:        "string",
			Required:    false,
			Description: "The directory to search (default: current directory)",
			Default:     "",
		},
	}
}

func (my *GlobTool) Execute(ctx context.Context, params map[string]any) (string, error) {
	pattern, ok := params["pattern"].(string)
	if !ok || pattern == "" {
		return "", fmt.Errorf("missing required parameter: pattern")
	}

	searchPath, _ := params["path"].(string)
	if searchPath == "" {
		searchPath = "."
	}

	absPath, err := filepath.Abs(searchPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}

	fullPattern := filepath.Join(absPath, pattern)
	matches, err := filepath.Glob(fullPattern)
	if err != nil {
		return "", fmt.Errorf("failed to match pattern: %w", err)
	}

	if len(matches) == 0 {
		return "", nil
	}

	for i, m := range matches {
		rel, err := filepath.Rel(absPath, m)
		if err == nil {
			matches[i] = rel
		}
	}

	return strings.Join(matches, "\n"), nil
}
