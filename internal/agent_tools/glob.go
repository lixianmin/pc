package agent_tools

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

func Glob(ctx context.Context, pattern string, searchPath string) (string, error) {
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
