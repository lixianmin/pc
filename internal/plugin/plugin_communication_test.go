package plugin

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/lixianmin/pc/pkg/protocol"
)

func TestPluginCommunication(t *testing.T) {
	pluginPath := findPluginBinary(t)
	if pluginPath == "" {
		t.Skip("Plugin binary not found, skipping integration test")
	}

	t.Run("initialize with config", func(t *testing.T) {
		cmd := exec.Command(pluginPath)
		proto := protocol.NewStdioProtocolWithTimeout(cmd, 5*time.Second)

		if err := proto.Connect(); err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer proto.Close()

		result, err := proto.Call("initialize", map[string]any{
			"api_key":  "test-key",
			"model":    "gpt-4o-mini",
			"base_url": "https://api.openai.com/v1",
		})
		if err != nil {
			t.Errorf("Call(initialize) error = %v", err)
			return
		}

		t.Logf("Result: %+v", result)

		resultMap, ok := result.(map[string]any)
		if !ok {
			t.Fatal("Result is not a map")
		}
		if resultMap["status"] != "initialized" {
			t.Errorf("status = %v, want initialized", resultMap["status"])
		}
	})

	t.Run("unknown method", func(t *testing.T) {
		cmd := exec.Command(pluginPath)
		proto := protocol.NewStdioProtocolWithTimeout(cmd, 5*time.Second)

		if err := proto.Connect(); err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer proto.Close()

		_, err := proto.Call("unknown_method", nil)
		if err == nil {
			t.Error("Expected error for unknown method")
		}
		t.Logf("Error (expected): %v", err)
	})
}

func TestPluginFullCommunicationFlow(t *testing.T) {
	pluginPath := findPluginBinary(t)
	if pluginPath == "" {
		t.Skip("Plugin binary not found, skipping integration test")
	}

	cmd := exec.Command(pluginPath)
	proto := protocol.NewStdioProtocolWithTimeout(cmd, 10*time.Second)

	if err := proto.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer proto.Close()

	initResult, err := proto.Call("initialize", map[string]any{
		"api_key":  "test-key",
		"model":    "gpt-4o-mini",
		"base_url": "https://api.openai.com/v1",
	})
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	t.Logf("Initialize result: %+v", initResult)

	completeResult, err := proto.Call("complete", map[string]any{
		"messages": []map[string]string{
			{"role": "user", "content": "Say hello"},
		},
	})
	if err != nil {
		t.Errorf("Complete failed: %v", err)
	} else {
		t.Logf("Complete result: %+v", completeResult)
	}
}

func TestPluginMessageFormat(t *testing.T) {
	pluginPath := findPluginBinary(t)
	if pluginPath == "" {
		t.Skip("Plugin binary not found")
	}

	req := protocol.NewRequest("initialize", map[string]any{
		"api_key": "test-key",
		"model":   "gpt-4o-mini",
	})

	data, err := req.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	t.Logf("Request JSON: %s", string(data))

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to decode request JSON: %v", err)
	}

	if decoded["method"] != "initialize" {
		t.Errorf("method = %v, want initialize", decoded["method"])
	}
	if decoded["version"] != "1.0" {
		t.Errorf("version = %v, want 1.0", decoded["version"])
	}
	if decoded["type"] != "call" {
		t.Errorf("type = %v, want call", decoded["type"])
	}
}

func TestPluginBinaryExists(t *testing.T) {
	pluginPath := findPluginBinary(t)
	if pluginPath == "" {
		t.Skip("Plugin binary not found")
	}

	t.Logf("Plugin binary found at: %s", pluginPath)

	info, err := os.Stat(pluginPath)
	if err != nil {
		t.Fatalf("Failed to stat plugin binary: %v", err)
	}

	t.Logf("Plugin binary size: %d bytes, mode: %v", info.Size(), info.Mode())
}

func findPluginBinary(t *testing.T) string {
	searchPaths := []string{
		filepath.Join(os.Getenv("HOME"), ".pc/plugins/llm/openai/bin/openai-llm"),
		"/tmp/openai-llm",
		"../examples/plugins/llm/openai/bin/openai-llm",
	}

	for _, path := range searchPaths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			continue
		}
		if _, err := os.Stat(absPath); err == nil {
			return absPath
		}
	}

	t.Log("Attempting to build plugin...")
	pluginSrc := "../examples/plugins/llm/openai/cmd/openai-llm"
	absSrc, err := filepath.Abs(pluginSrc)
	if err != nil {
		t.Logf("Failed to get absolute path for plugin source: %v", err)
		return ""
	}

	if _, err := os.Stat(absSrc); os.IsNotExist(err) {
		t.Logf("Plugin source not found at: %s", absSrc)
		return ""
	}

	outputPath := "/tmp/test-openai-llm"
	buildCmd := exec.Command("go", "build", "-o", outputPath, absSrc)
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		t.Logf("Failed to build plugin: %v", err)
		return ""
	}

	t.Logf("Built plugin at: %s", outputPath)
	return outputPath
}
