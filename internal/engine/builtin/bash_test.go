package builtin_test

import (
	"context"
	"strings"
	"testing"

	"github.com/lixianmin/pc/internal/engine/builtin"
)

func TestBashTool_Execute(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]any
		want        string
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid echo command",
			params:      map[string]any{"command": "echo hello"},
			want:        "hello\n",
			wantErr:     false,
			errContains: "",
		},
		{
			name:        "command with workdir",
			params:      map[string]any{"command": "pwd", "workdir": "/tmp"},
			want:        "/tmp\n",
			wantErr:     false,
			errContains: "",
		},
		{
			name:        "timeout",
			params:      map[string]any{"command": "sleep 2", "timeout": 0.1},
			want:        "",
			wantErr:     true,
			errContains: "timed out",
		},
		{
			name:        "missing command parameter",
			params:      map[string]any{},
			want:        "",
			wantErr:     true,
			errContains: "missing required parameter: command",
		},
		{
			name:        "command failure",
			params:      map[string]any{"command": "exit 1"},
			want:        "",
			wantErr:     true,
			errContains: "command failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			tool := builtin.NewBashTool()
			result, err := tool.Execute(ctx, tt.params)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if !strings.Contains(result, tt.want) {
					t.Errorf("expected result containing %q, got %q", tt.want, result)
				}
			}
		})
	}
}

func TestBashTool_IsDangerous(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected bool
	}{
		{"safe command", "ls -la", false},
		{"rm -rf /", "rm -rf /", true},
		{"mkfs", "mkfs.ext4 /dev/sda1", true},
		{"dd if=", "dd if=/dev/zero of=/dev/sda", true},
		{"chmod 777 /", "chmod 777 /", true},
		{"fork bomb", ":(){ :|:& };:", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := builtin.NewBashTool()
			result := tool.IsDangerous(tt.command)
			if result != tt.expected {
				t.Errorf("IsDangerous(%q) = %v, want %v", tt.command, result, tt.expected)
			}
		})
	}
}
