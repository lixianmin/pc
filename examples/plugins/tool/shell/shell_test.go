package shell

import (
	"testing"
	"time"
)

func TestShellTool_Execute(t *testing.T) {
	tests := []struct {
		name      string
		params    CallParams
		wantErr   bool
		checkFunc func(*testing.T, *CallResult)
	}{
		{
			name: "echo command",
			params: CallParams{
				Command: "echo hello",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, result *CallResult) {
				if result.ExitCode != 0 {
					t.Errorf("expected exit code 0, got %d", result.ExitCode)
				}
				if result.Stdout != "hello\n" {
					t.Errorf("expected stdout 'hello\\n', got %q", result.Stdout)
				}
			},
		},
		{
			name: "command with stderr",
			params: CallParams{
				Command: "ls /nonexistent_dir_12345",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, result *CallResult) {
				if result.ExitCode == 0 {
					t.Error("expected non-zero exit code")
				}
				if result.Stderr == "" {
					t.Error("expected stderr output")
				}
			},
		},
		{
			name: "empty command",
			params: CallParams{
				Command: "",
			},
			wantErr: true,
		},
		{
			name: "blocked command",
			params: CallParams{
				Command: "rm -rf /",
			},
			wantErr: true,
		},
		{
			name: "command with working dir",
			params: CallParams{
				Command:    "pwd",
				WorkingDir: "/tmp",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, result *CallResult) {
				if result.Stdout != "/tmp\n" {
					t.Errorf("expected stdout '/tmp\\n', got %q", result.Stdout)
				}
			},
		},
	}

	config := Config{
		Timeout:          30 * time.Second,
		ConfirmDangerous: true,
		BlockedCommands: []string{
			"rm -rf /",
			"mkfs",
		},
	}

	tool := NewShellTool(config)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tool.Execute(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

func TestShellTool_IsDangerous(t *testing.T) {
	tests := []struct {
		command string
		want    bool
	}{
		{"ls -la", false},
		{"rm -rf /", true},
		{"rm -rf /*", true},
		{"mkfs.ext4 /dev/sda1", true},
		{"dd if=/dev/zero of=/dev/sda", true},
		{"cat /etc/passwd", false},
		{"echo hello", false},
	}

	config := Config{Timeout: 30 * time.Second}
	tool := NewShellTool(config)

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			if got := tool.IsDangerous(tt.command); got != tt.want {
				t.Errorf("IsDangerous(%q) = %v, want %v", tt.command, got, tt.want)
			}
		})
	}
}

func TestShellTool_ValidateCommand(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		command string
		wantErr bool
	}{
		{
			name: "allowed command",
			config: Config{
				AllowedCommands: []string{"ls", "cat"},
			},
			command: "ls -la",
			wantErr: false,
		},
		{
			name: "not allowed command",
			config: Config{
				AllowedCommands: []string{"ls", "cat"},
			},
			command: "rm file.txt",
			wantErr: true,
		},
		{
			name: "blocked command",
			config: Config{
				BlockedCommands: []string{"rm -rf"},
			},
			command: "rm -rf dir",
			wantErr: true,
		},
		{
			name:    "no restrictions",
			config:  Config{},
			command: "any command",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := NewShellTool(tt.config)
			err := tool.validateCommand(tt.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
