package agent_tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func Bash(ctx context.Context, workdir string, command string) (string, error) {
	if IsDangerousCommand(command) {
		return "", fmt.Errorf("command is potentially dangerous and has been blocked")
	}

	var cmd *exec.Cmd
	if workdir != "" {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
		cmd.Dir = workdir
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()
	if stderr.Len() > 0 {
		output += stderr.String()
	}

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("command timed out")
		}
		return "", fmt.Errorf("command failed: %w", err)
	}

	return output, nil
}

func IsDangerousCommand(command string) bool {
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
