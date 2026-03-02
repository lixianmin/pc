package gateway

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lixianmin/pc/internal/engine"
	"github.com/lixianmin/pc/internal/plugin"
)

func TestRPCClient_Connect(t *testing.T) {
	tmpDir := t.TempDir()
	socketPath := filepath.Join(tmpDir, "test.sock")

	client := NewRPCClient(socketPath)

	// Try to connect when server is not running
	err := client.Connect()
	if err == nil {
		t.Error("Connect() should fail when server is not running")
	}
}

func TestRPCClient_Call_NotConnected(t *testing.T) {
	tmpDir := t.TempDir()
	socketPath := filepath.Join(tmpDir, "test.sock")

	client := NewRPCClient(socketPath)

	_, err := client.Call("test", nil)
	if err == nil {
		t.Error("Call() should fail when not connected")
	}
}

func TestRPCClient_SetTimeout(t *testing.T) {
	tmpDir := t.TempDir()
	socketPath := filepath.Join(tmpDir, "test.sock")

	client := NewRPCClient(socketPath)

	// Default timeout should be 60 seconds
	if client.timeout != 60*time.Second {
		t.Errorf("Default timeout = %v, want 60s", client.timeout)
	}

	// Set new timeout
	client.SetTimeout(10 * time.Second)
	if client.timeout != 10*time.Second {
		t.Errorf("After SetTimeout, timeout = %v, want 10s", client.timeout)
	}
}

func TestRPCServer_StartStop(t *testing.T) {
	tmpDir := t.TempDir()
	socketPath := filepath.Join(tmpDir, "test.sock")

	// Create mock engine and plugin manager
	pluginMgr, _ := plugin.NewPluginManager(tmpDir)
	eng := engine.NewEngine(pluginMgr)

	server := NewRPCServer(socketPath, eng, pluginMgr)

	// Start server
	err := server.Start()
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Check socket file exists
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		t.Error("Socket file should exist after Start()")
	}

	// Check socket permissions
	info, err := os.Stat(socketPath)
	if err != nil {
		t.Errorf("Failed to stat socket: %v", err)
	} else {
		// Permission should be 0600
		mode := info.Mode().Perm()
		if mode != 0600 {
			t.Errorf("Socket permissions = %o, want 0600", mode)
		}
	}

	// Stop server
	err = server.Stop()
	if err != nil {
		t.Errorf("Stop() error = %v", err)
	}
}

func TestNewRPCClient(t *testing.T) {
	socketPath := "/tmp/test.sock"
	client := NewRPCClient(socketPath)

	if client == nil {
		t.Fatal("NewRPCClient() returned nil")
	}

	if client.socketPath != socketPath {
		t.Errorf("socketPath = %v, want %v", client.socketPath, socketPath)
	}

	if client.timeout != 60*time.Second {
		t.Errorf("timeout = %v, want 60s", client.timeout)
	}
}

func TestNewRPCServer(t *testing.T) {
	socketPath := "/tmp/test.sock"
	tmpDir := t.TempDir()

	pluginMgr, _ := plugin.NewPluginManager(tmpDir)
	eng := engine.NewEngine(pluginMgr)

	server := NewRPCServer(socketPath, eng, pluginMgr)

	if server == nil {
		t.Fatal("NewRPCServer() returned nil")
	}

	if server.socketPath != socketPath {
		t.Errorf("socketPath = %v, want %v", server.socketPath, socketPath)
	}

	if server.engine != eng {
		t.Error("engine not set correctly")
	}

	if server.pluginMgr != pluginMgr {
		t.Error("pluginMgr not set correctly")
	}

	// Check handlers are registered
	expectedHandlers := []string{"ProcessMessage", "GetStatus", "ListSkills", "ExecuteSkill", "ListTasks", "AddTask", "CompleteTask"}
	for _, method := range expectedHandlers {
		if _, ok := server.handlers[method]; !ok {
			t.Errorf("Handler for %s not registered", method)
		}
	}
}
