package main

import (
	"bytes"
	"os"
	"testing"
)

func TestVersionOutput(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		want     string
		wantExit bool
	}{
		{
			name: "version flag",
			args: []string{"version"},
			want: "PersonalClaw version",
		},
		{
			name: "default output",
			args: []string{},
			want: "PersonalClaw - AI Agent Operating System Kernel",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Redirect stdout to capture output
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Save original args and restore after test
			oldArgs := os.Args
			os.Args = []string{"pc"}
			os.Args = append(os.Args, tt.args...)

			main()

			// Restore stdout and args
			w.Close()
			os.Stdout = old
			os.Args = oldArgs

			// Read captured output
			var buf bytes.Buffer
			buf.ReadFrom(r)
			got := buf.String()

			if got == "" && len(tt.want) > 0 {
				t.Errorf("main() produced no output, want %q", tt.want)
			}
		})
	}
}
