package agent_tools

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/lixianmin/got/convert"
	"github.com/lixianmin/pc/pkg/ks"
)

func Edit(ctx context.Context, filePath, oldString, newString string, expectedReplacements int) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", ks.TraceError("ReadFileError", "err", err)
	}

	var contentStr = convert.String(content)
	if !strings.Contains(contentStr, oldString) {
		return "", ks.TraceError("OldStringNotFound", "filePath", filePath, "oldString", oldString)
	}

	if expectedReplacements <= 1 {
		expectedReplacements = 1
	}

	var newContent string
	newContent = strings.Replace(contentStr, oldString, newString, expectedReplacements)

	if err := os.WriteFile(filePath, convert.Bytes(newContent), 0644); err != nil {
		return "", ks.TraceError("WriteFileError", "err", err)
	}

	var result = fmt.Sprintf("Successfully edited %s", filePath)
	return result, nil
}
