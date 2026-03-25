package agent_tools

import (
	"context"
	"fmt"
	"os"
	"strings"
)

func Edit(ctx context.Context, filePath, oldString, newString string, replaceAll bool) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, oldString) {
		return "", fmt.Errorf("old_string not found in file")
	}

	count := strings.Count(contentStr, oldString)
	if count > 1 && !replaceAll {
		return "", fmt.Errorf("found %d matches, please use replace_all=true or provide more context", count)
	}

	var newContent string
	if replaceAll {
		newContent = strings.ReplaceAll(contentStr, oldString, newString)
	} else {
		newContent = strings.Replace(contentStr, oldString, newString, 1)
	}

	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("Successfully edited %s", filePath), nil
}
