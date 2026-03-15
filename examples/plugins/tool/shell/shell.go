package shell

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Config struct {
	Timeout          time.Duration
	ConfirmDangerous bool
	AllowedCommands  []string
	BlockedCommands  []string
}

type ShellTool struct {
	config Config
}

func NewShellTool(config Config) *ShellTool {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	return &ShellTool{config: config}
}

type CallParams struct {
	Command    string `json:"command"`
	Timeout    int    `json:"timeout,omitempty"`
	WorkingDir string `json:"working_dir,omitempty"`
}

type CallResult struct {
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
	ExitCode   int    `json:"exit_code"`
	DurationMs int64  `json:"duration_ms"`
}

func (my *ShellTool) Execute(params CallParams) (*CallResult, error) {
	if params.Command == "" {
		return nil, fmt.Errorf("command is required")
	}

	if err := my.validateCommand(params.Command); err != nil {
		return nil, err
	}

	timeout := my.config.Timeout
	if params.Timeout > 0 {
		timeout = time.Duration(params.Timeout) * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	start := time.Now()

	cmd := exec.CommandContext(ctx, "sh", "-c", params.Command)
	if params.WorkingDir != "" {
		cmd.Dir = params.WorkingDir
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(start)

	result := &CallResult{
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

func (my *ShellTool) validateCommand(command string) error {
	if len(my.config.AllowedCommands) > 0 {
		allowed := false
		for _, prefix := range my.config.AllowedCommands {
			if strings.HasPrefix(command, prefix) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("command not in allowed list")
		}
	}

	for _, blocked := range my.config.BlockedCommands {
		if strings.Contains(command, blocked) {
			return fmt.Errorf("command contains blocked pattern: %s", blocked)
		}
	}

	return nil
}

func (my *ShellTool) IsDangerous(command string) bool {
	dangerousPatterns := []string{
		"rm -rf /",
		"rm -rf /*",
		"mkfs",
		"dd if=",
		":(){ :|:& };:",
		"> /dev/sda",
		"chmod -R 777 /",
		"chown -R",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(command, pattern) {
			return true
		}
	}
	return false
}

func (my *ShellTool) SetConfig(config Config) {
	my.config = config
}
