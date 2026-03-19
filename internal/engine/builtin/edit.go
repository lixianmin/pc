package builtin

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/lixianmin/pc/pkg/types"
)

type EditTool struct{}

func NewEditTool() *EditTool {
	return &EditTool{}
}

func (my *EditTool) Name() string {
	return "edit"
}

func (my *EditTool) Description() string {
	return "Edit file by string replacement"
}

func (my *EditTool) Parameters() map[string]types.ParamSchema {
	return map[string]types.ParamSchema{
		"path": {
			Type:        "string",
			Required:    true,
			Description: "The absolute path to edit",
		},
		"old_string": {
			Type:        "string",
			Required:    true,
			Description: "The text to replace",
		},
		"new_string": {
			Type:        "string",
			Required:    true,
			Description: "The replacement text",
		},
		"replace_all": {
			Type:        "boolean",
			Required:    false,
			Description: "Replace all occurrences (default false)",
			Default:     false,
		},
	}
}

func (my *EditTool) Execute(ctx context.Context, params map[string]any) (string, error) {
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return "", fmt.Errorf("missing required parameter: path")
	}

	oldString, ok := params["old_string"].(string)
	if !ok {
		return "", fmt.Errorf("missing required parameter: old_string")
	}

	newString, ok := params["new_string"].(string)
	if !ok {
		return "", fmt.Errorf("missing required parameter: new_string")
	}

	replaceAll := params["replace_all"] == true

	content, err := os.ReadFile(path)
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

	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("Successfully edited %s", path), nil
}
