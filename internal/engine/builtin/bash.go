package builtin

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type BashTool struct {
	workdir string
	timeout time.Duration
}

func NewBashTool() *BashTool {
	return &BashTool{
		workdir: "",
		timeout: 120 * time.Second,
	}
}

func (my *BashTool) Name() string {
	return "bash"
}

func (my *BashTool) Execute(ctx context.Context, params map[string]any) (string, error) {
	command, ok := params["command"].(string)
	if !ok || command == "" {
		return "", fmt.Errorf("missing required parameter: command")
	}

	workdir, _ := params["workdir"].(string)
	if workdir == "" {
		workdir = my.workdir
	}

	timeoutSec, ok := params["timeout"].(float64)
	if !ok || timeoutSec <= 0 {
		timeoutSec = 120
	}
	timeout := time.Duration(timeoutSec) * time.Second

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

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
			return "", fmt.Errorf("command timed out after %s", timeout)
		}
		return "", fmt.Errorf("command failed: %w", err)
	}

	return output, nil
}

func (my *BashTool) SetWorkDir(dir string) {
	my.workdir = dir
}

func (my *BashTool) IsDangerous(command string) bool {
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
