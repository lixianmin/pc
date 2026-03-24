package tui

import (
	"testing"
)

func TestHandleCommand_TaskCommands(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantCommand string
	}{
		{
			name:        "task list command",
			input:       "/task list",
			wantCommand: "/task",
		},
		{
			name:        "task add command",
			input:       "/task add Test Task",
			wantCommand: "/task",
		},
		{
			name:        "task complete command",
			input:       "/task complete task-001",
			wantCommand: "/task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := parseCommand(tt.input)
			if len(parts) == 0 {
				t.Error("Expected non-empty command parts")
				return
			}
			if parts[0] != tt.wantCommand {
				t.Errorf("Command = %q, want %q", parts[0], tt.wantCommand)
			}
		})
	}
}

func TestCompleteCommand_Task(t *testing.T) {
	m := &Model{}

	tests := []struct {
		input    string
		expected string
	}{
		{"/ta", "/task"},
		{"/task ", "/task list"},  // Auto-completes to first subcommand
		{"/task l", "/task list"}, // Complete subcommand
	}

	for _, tt := range tests {
		result := m.completeCommand(tt.input)
		if result != tt.expected {
			t.Errorf("completeCommand(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestParseTaskCommand(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantSubcommand string
		wantArgs       []string
	}{
		{
			name:           "list",
			input:          "/task list",
			wantSubcommand: "list",
			wantArgs:       []string{},
		},
		{
			name:           "list with filter",
			input:          "/task list --status pending",
			wantSubcommand: "list",
			wantArgs:       []string{"--status", "pending"},
		},
		{
			name:           "add simple",
			input:          "/task add My Task",
			wantSubcommand: "add",
			wantArgs:       []string{"My", "Task"},
		},
		{
			name:           "add with description",
			input:          "/task add My Task --description \"A description\"",
			wantSubcommand: "add",
			wantArgs:       []string{"My", "Task", "--description", "A description"},
		},
		{
			name:           "complete",
			input:          "/task complete task-001",
			wantSubcommand: "complete",
			wantArgs:       []string{"task-001"},
		},
		{
			name:           "delete",
			input:          "/task delete task-001",
			wantSubcommand: "delete",
			wantArgs:       []string{"task-001"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subcommand, args := parseTaskCommand(tt.input)
			if subcommand != tt.wantSubcommand {
				t.Errorf("Subcommand = %q, want %q", subcommand, tt.wantSubcommand)
			}
			if len(args) != len(tt.wantArgs) {
				t.Errorf("Args length = %d, want %d", len(args), len(tt.wantArgs))
			}
		})
	}
}

// parseCommand is a helper to parse command string into parts
func parseCommand(input string) []string {
	// Simple parsing - in real implementation this is more sophisticated
	parts := make([]string, 0)
	current := ""
	inQuotes := false

	for _, ch := range input {
		switch ch {
		case '"':
			inQuotes = !inQuotes
		case ' ', '\t':
			if !inQuotes && current != "" {
				parts = append(parts, current)
				current = ""
			} else if inQuotes {
				current += string(ch)
			}
		default:
			current += string(ch)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}

// parseTaskCommand parses /task commands
func parseTaskCommand(input string) (string, []string) {
	parts := parseCommand(input)
	if len(parts) < 2 {
		return "", nil
	}
	// Skip "/task"
	if len(parts) < 3 {
		return parts[1], nil
	}
	return parts[1], parts[2:]
}
