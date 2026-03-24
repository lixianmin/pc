package engine

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

type SecurityLevel int

const (
	SecurityLevelStrict SecurityLevel = iota
	SecurityLevelModerate
	SecurityLevelPermissive
)

type SecurityConfig struct {
	Level            SecurityLevel
	ConfirmDangerous bool
	AllowedCommands  []string
	BlockedCommands  []string
	AllowedTools     []string
	BlockedTools     []string
	ConfirmCallback  func(toolName string, params map[string]interface{}) (bool, error)
}

type SecurityChecker struct {
	config SecurityConfig
}

func NewSecurityChecker(config SecurityConfig) *SecurityChecker {
	return &SecurityChecker{config: config}
}

func (my *SecurityChecker) CheckToolCall(call ToolCall) error {
	if err := my.checkToolAllowed(call.Name); err != nil {
		return err
	}

	if call.Name == "bash" {
		if err := my.checkShellCommand(call.Params); err != nil {
			return err
		}
	}

	return nil
}

func (my *SecurityChecker) checkToolAllowed(toolName string) error {
	if len(my.config.BlockedTools) > 0 {
		for _, blocked := range my.config.BlockedTools {
			if toolName == blocked {
				return fmt.Errorf("tool '%s' is blocked", toolName)
			}
		}
	}

	if len(my.config.AllowedTools) > 0 {
		allowed := false
		for _, a := range my.config.AllowedTools {
			if toolName == a {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("tool '%s' is not in allowed list", toolName)
		}
	}

	return nil
}

func (my *SecurityChecker) checkShellCommand(params map[string]interface{}) error {
	command, ok := params["command"].(string)
	if !ok {
		return nil
	}

	if len(my.config.BlockedCommands) > 0 {
		for _, blocked := range my.config.BlockedCommands {
			if strings.Contains(command, blocked) {
				return fmt.Errorf("command contains blocked pattern: %s", blocked)
			}
		}
	}

	if my.isDangerousCommand(command) {
		return fmt.Errorf("dangerous command detected: %s", my.maskCommand(command))
	}

	if len(my.config.AllowedCommands) > 0 {
		allowed := false
		for _, prefix := range my.config.AllowedCommands {
			if strings.HasPrefix(command, prefix) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("command not in allowed list: %s", my.maskCommand(command))
		}
	}

	return nil
}

func (my *SecurityChecker) isDangerousCommand(command string) bool {
	dangerousPatterns := []string{
		`rm\s+-rf\s+/`,
		`rm\s+-rf\s+/\*`,
		`mkfs`,
		`dd\s+if=.*\s+of=/dev/`,
		`:$$\)\s*{\s*:\|:&\s*}\s*;:`,
		`>\s*/dev/sd[a-z]`,
		`chmod\s+-R\s+777\s+/`,
		`chown\s+-R\s+.*\s+/`,
		`>\s*/dev/null\s+2>&1\s*;`,
		`curl.*\|\s*(ba)?sh`,
		`wget.*\|\s*(ba)?sh`,
	}

	for _, pattern := range dangerousPatterns {
		matched, _ := regexp.MatchString(pattern, command)
		if matched {
			return true
		}
	}

	dangerousCommands := []string{
		"rm -rf /",
		"rm -rf /*",
		"shutdown",
		"reboot",
		"halt",
		"poweroff",
		"init 0",
		"init 6",
	}

	commandLower := strings.ToLower(command)
	for _, dangerous := range dangerousCommands {
		if strings.Contains(commandLower, dangerous) {
			return true
		}
	}

	return false
}

func (my *SecurityChecker) maskCommand(command string) string {
	if len(command) > 50 {
		return command[:47] + "..."
	}
	return command
}

func (my *SecurityChecker) NeedsConfirmation(call ToolCall) bool {
	if !my.config.ConfirmDangerous {
		return false
	}

	if call.Name == "bash" {
		if command, ok := call.Params["command"].(string); ok {
			return my.isDangerousCommand(command) || my.isPotentiallyDestructive(command)
		}
	}

	return false
}

func (my *SecurityChecker) isPotentiallyDestructive(command string) bool {
	destructivePatterns := []string{
		"rm",
		"rmdir",
		"del",
		"format",
		"erase",
		"truncate",
		"shred",
	}

	commandLower := strings.ToLower(command)
	for _, pattern := range destructivePatterns {
		if strings.Contains(commandLower, pattern) {
			return true
		}
	}

	return false
}

func (my *SecurityChecker) Confirm(ctx context.Context, call ToolCall) (bool, error) {
	if my.config.ConfirmCallback != nil {
		return my.config.ConfirmCallback(call.Name, call.Params)
	}

	return false, fmt.Errorf("confirmation callback not configured")
}

func (my *SecurityChecker) ValidateMultiple(calls []ToolCall) error {
	for i, call := range calls {
		if err := my.CheckToolCall(call); err != nil {
			return fmt.Errorf("tool call %d (%s) failed security check: %w", i+1, call.Name, err)
		}
	}
	return nil
}
