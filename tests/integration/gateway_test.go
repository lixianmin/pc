package integration

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lixianmin/pc/internal/gateway"
	"github.com/lixianmin/pc/internal/engine"
	"github.com/lixianmin/pc/internal/plugin"
)

// TestGatewayStartup tests the gateway daemon startup sequence
func TestGatewayStartup(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "pc-integration-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	socketPath := filepath.Join(tempDir, "pc.sock")

	// Create plugin manager (without actual plugins)
	pluginMgr, err := plugin.NewPluginManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create plugin manager: %v", err)
	}

	// Create engine
	eng := engine.NewEngine(pluginMgr)

	// Create and start RPC server
	rpcServer := gateway.NewRPCServer(socketPath, eng, pluginMgr)
	if err := rpcServer.Start(); err != nil {
		t.Fatalf("Failed to start RPC server: %v", err)
	}
	defer rpcServer.Stop()

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)

	// Create RPC client
	client := gateway.NewRPCClient(socketPath)
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect to RPC server: %v", err)
	}
	defer client.Close()

	// Test GetStatus
	status, err := client.GetStatus()
	if err != nil {
		t.Errorf("GetStatus failed: %v", err)
	}
	if !status.Running {
		t.Error("Gateway should be running")
	}
	if status.Pid == 0 {
		t.Error("PID should be set")
	}
}

// TestTUIAndChannelCoexistence tests that TUI and Channel inputs can coexist
func TestTUIAndChannelCoexistence(t *testing.T) {
	// Create input source registry
	registry := gateway.NewInputSourceRegistry()
	router := gateway.NewMessageRouter()

	// Create TUI source
	tuiSource := gateway.NewTUIInputSource("/tmp/test.sock")

	// Create Channel source
	telegramSource := gateway.NewChannelInputSource("telegram")

	// Register sources
	if err := registry.Register("tui", tuiSource); err != nil {
		t.Fatalf("Failed to register TUI source: %v", err)
	}
	if err := registry.Register("telegram", telegramSource); err != nil {
		t.Fatalf("Failed to register telegram source: %v", err)
	}

	// Start sources
	if err := tuiSource.Start(); err != nil {
		t.Fatalf("Failed to start TUI source: %v", err)
	}
	defer tuiSource.Stop()

	if err := telegramSource.Start(); err != nil {
		t.Fatalf("Failed to start telegram source: %v", err)
	}
	defer telegramSource.Stop()

	// Setup message tracking
	var tuiMessages, channelMessages int
	router.RegisterHandler("tui", func(msg *gateway.RoutedMessage) {
		tuiMessages++
	})
	router.RegisterHandler("telegram", func(msg *gateway.RoutedMessage) {
		channelMessages++
	})

	// Simulate messages
	tuiSource.ReceiveMessage("session-1", "user-1", "Hello from TUI", "")
	telegramSource.ReceiveMessage("chat-1", "user-2", "Hello from Telegram", "")

	// Route messages
	for i := 0; i < 2; i++ {
		select {
		case msg := <-tuiSource.MessageChannel():
			router.Route(msg)
		case msg := <-telegramSource.MessageChannel():
			router.Route(msg)
		case <-time.After(100 * time.Millisecond):
		}
	}

	time.Sleep(50 * time.Millisecond)

	// Verify both messages were handled
	if tuiMessages != 1 {
		t.Errorf("Expected 1 TUI message, got %d", tuiMessages)
	}
	if channelMessages != 1 {
		t.Errorf("Expected 1 Channel message, got %d", channelMessages)
	}
}

// TestMessageRouting tests end-to-end message routing through the gateway
func TestMessageRouting(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pc-routing-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	socketPath := filepath.Join(tempDir, "pc.sock")

	// Setup components
	pluginMgr, err := plugin.NewPluginManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create plugin manager: %v", err)
	}
	eng := engine.NewEngine(pluginMgr)
	rpcServer := gateway.NewRPCServer(socketPath, eng, pluginMgr)

	if err := rpcServer.Start(); err != nil {
		t.Fatalf("Failed to start RPC server: %v", err)
	}
	defer rpcServer.Stop()

	time.Sleep(100 * time.Millisecond)

	// Connect client
	client := gateway.NewRPCClient(socketPath)
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Create session and send message
	sessionID := "test-session"
	if err := eng.CreateSession(sessionID); err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	response, err := client.ProcessMessage(sessionID, "Hello")
	if err != nil {
		t.Errorf("ProcessMessage failed: %v", err)
	}

	// Response might be empty since no LLM plugin is configured
	_ = response
}

// TestRPCClientServerCommunication tests RPC client-server communication
func TestRPCClientServerCommunication(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pc-rpc-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	socketPath := filepath.Join(tempDir, "pc.sock")

	pluginMgr, err := plugin.NewPluginManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create plugin manager: %v", err)
	}
	eng := engine.NewEngine(pluginMgr)
	rpcServer := gateway.NewRPCServer(socketPath, eng, pluginMgr)

	if err := rpcServer.Start(); err != nil {
		t.Fatalf("Failed to start RPC server: %v", err)
	}
	defer rpcServer.Stop()

	time.Sleep(100 * time.Millisecond)

	client := gateway.NewRPCClient(socketPath)
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Test various RPC methods
	tests := []struct {
		name   string
		method func() error
	}{
		{
			name: "GetStatus",
			method: func() error {
				_, err := client.GetStatus()
				return err
			},
		},
		{
			name: "ListSkills",
			method: func() error {
				_, err := client.ListSkills()
				return err
			},
		},
		{
			name: "ListTasks",
			method: func() error {
				_, err := client.ListTasks("")
				return err
			},
		},
		{
			name: "AddTask",
			method: func() error {
				_, err := client.AddTask("Test Task", "Description")
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.method(); err != nil {
				t.Errorf("%s failed: %v", tt.name, err)
			}
		})
	}
}
