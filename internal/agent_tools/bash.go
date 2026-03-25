package agent_tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/lixianmin/logo"
	"github.com/lixianmin/pc/pkg/ks"
	"github.com/lixianmin/pc/pkg/tools"
)

func Bash(ctx context.Context, command string) (string, error) {
	if isDangerousCommand(command) {
		return "", ks.TraceError("DangerousCommand", "command", command)
	}

	logo.JsonI("command", command)
	var cmd *exec.Cmd = exec.CommandContext(ctx, "sh", "-c", command)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()
	if stderr.Len() > 0 {
		output += stderr.String()
	}

	if err != nil {
		return "", ks.TraceError("BashError", "command", command, "err", err)
	}

	logo.JsonI("command", command, "output", tools.StrTake(output, 300))
	var result = fmt.Sprintf("bash command: `%s` \n\n bash output: \n```%s``` ", command, output)
	return result, nil
}

func isDangerousCommand(command string) bool {
	dangerousPatterns := []string{
		"rm -rf /",
		"mkfs",
		"dd if=",
		"> /dev/sd",
		":(){ :|:& };:",
		"chmod 777 /",
		"chown -R",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(command, pattern) {
			return true
		}
	}

	return false
}
