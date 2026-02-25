package openai

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/lixianmin/pc/pkg/protocol"
)

func TestPluginIntegration(t *testing.T) {
	// Build the plugin
	buildCmd := exec.Command("go", "build", "-o", "/tmp/openai-llm", "./cmd/openai-llm")
	if err := buildCmd.Run(); err != nil {
		t.Skipf("Failed to build plugin: %v", err)
	}
	defer os.Remove("/tmp/openai-llm")

	tests := []struct {
		name     string
		method   string
		params   any
		wantErr  bool
		validate func(*testing.T, *protocol.Response)
	}{
		{
			name:   "complete method call",
			method: "complete",
			params: map[string]any{
				"prompt": "Hello, world!",
				"options": map[string]any{},
			},
			wantErr: false,
			validate: func(t *testing.T, resp *protocol.Response) {
				if resp.Error != nil {
					t.Errorf("complete() returned error: %v", resp.Error)
				}
				result, ok := resp.Result.(map[string]any)
				if !ok {
					t.Error("complete() result is not a map")
					return
				}
				if _, ok := result["text"]; !ok {
					t.Error("complete() result missing 'text' field")
				}
			},
		},
		{
			name:   "complete with empty prompt",
			method: "complete",
			params: map[string]any{
				"prompt": "",
				"options": map[string]any{},
			},
			wantErr: true,
			validate: func(t *testing.T, resp *protocol.Response) {
				if resp.Error == nil {
					t.Error("complete() with empty prompt should return error")
				}
			},
		},
		{
			name:   "models method call",
			method: "models",
			params: map[string]any{},
			wantErr: false,
			validate: func(t *testing.T, resp *protocol.Response) {
				if resp.Error != nil {
					t.Errorf("models() returned error: %v", resp.Error)
				}
				result, ok := resp.Result.(map[string]any)
				if !ok {
					t.Error("models() result is not a map")
					return
				}
				models, ok := result["models"].([]any)
				if !ok {
					t.Error("models() result missing 'models' field")
					return
				}
				if len(models) == 0 {
					t.Error("models() returned empty list")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := protocol.NewRequest(tt.method, tt.params)
			reqData, err := req.Encode()
			if err != nil {
				t.Fatalf("Failed to encode request: %v", err)
			}

			// Start plugin process
			cmd := exec.Command("/tmp/openai-llm")
			stdin, err := cmd.StdinPipe()
			if err != nil {
				t.Fatalf("Failed to get stdin pipe: %v", err)
			}
			stdout := &bytes.Buffer{}
			cmd.Stdout = stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Start(); err != nil {
				t.Fatalf("Failed to start plugin: %v", err)
			}

			// Send request
			stdin.Write(reqData)
			stdin.Write([]byte("\n"))
			stdin.Close()

			// Wait for response with timeout
			done := make(chan error, 1)
			go func() {
				done <- cmd.Wait()
			}()

			select {
			case <-time.After(5 * time.Second):
				cmd.Process.Kill()
				t.Fatal("Plugin timeout")
			case err := <-done:
				if err != nil {
					t.Errorf("Plugin exited with error: %v", err)
				}
			}

			// Parse response
			output := bytes.TrimSpace(stdout.Bytes())
			if len(output) == 0 {
				t.Fatal("Plugin returned no output")
			}

			resp, err := protocol.DecodeResponse(output)
			if err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			// Validate
			tt.validate(t, resp)
		})
	}
}
