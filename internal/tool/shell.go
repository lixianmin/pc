package tool

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type ShellConfig struct {
	Timeout          time.Duration
	ConfirmDangerous bool
	AllowedCommands  []string
	BlockedCommands  []string
}

type ShellTool struct {
	config ShellConfig
}

func NewShellTool(config ShellConfig) *ShellTool {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	return &ShellTool{config: config}
}

func (t *ShellTool) Name() string {
	return "shell"
}

func (t *ShellTool) Description() string {
	return "Execute shell commands"
}

type ShellResult struct {
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
	ExitCode   int    `json:"exit_code"`
	DurationMs int64  `json:"duration_ms"`
}

func (t *ShellTool) Execute(ctx context.Context, params map[string]any) (any, error) {
	command, _ := params["command"].(string)
	if command == "" {
		return nil, fmt.Errorf("command is required")
	}

	if err := t.validateCommand(command); err != nil {
		return nil, err
	}

	timeout := t.config.Timeout
	if timeoutMs, ok := params["timeout"].(int); ok && timeoutMs > 0 {
		timeout = time.Duration(timeoutMs) * time.Millisecond
	}

	workingDir, _ := params["working_dir"].(string)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(start)

	result := &ShellResult{
		Stdout:     stdout.String(),
		Stderr:     stderr.String(),
		DurationMs: duration.Milliseconds(),
	}

	if ctx.Err() == context.DeadlineExceeded {
		result.ExitCode = -1
		return result, fmt.Errorf("command timed out after %v", timeout)
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = 1
		}
		return result, nil
	}

	result.ExitCode = 0
	return result, nil
}

func (t *ShellTool) validateCommand(command string) error {
	if len(t.config.AllowedCommands) > 0 {
		allowed := false
		for _, prefix := range t.config.AllowedCommands {
			if strings.HasPrefix(command, prefix) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("command not in allowed list")
		}
	}

	for _, blocked := range t.config.BlockedCommands {
		if strings.Contains(command, blocked) {
			return fmt.Errorf("command contains blocked pattern: %s", blocked)
		}
	}

	return nil
}
