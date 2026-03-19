package integration

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lixianmin/pc/internal/engine"
	"github.com/lixianmin/pc/internal/gateway"
	"github.com/lixianmin/pc/internal/plugin"
)

func TestProcessMessageStream(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
		message   string
		wantErr   bool
	}{
		{
			name:      "basic streaming message",
			sessionID: "stream-session-1",
			message:   "Hello streaming",
			wantErr:   false,
		},
		{
			name:      "empty message",
			sessionID: "stream-session-2",
			message:   "",
			wantErr:   true,
		},
		{
			name:      "long message",
			sessionID: "stream-session-3",
			message:   "This is a longer message that should be processed through the streaming pipeline",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, err := os.MkdirTemp("", "pc-stream-test-*")
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

			client := gateway.NewRpcClient(socketPath)
			if err := client.Connect(); err != nil {
				t.Fatalf("Failed to connect: %v", err)
			}
			defer client.Close()

			chunks, err := client.ProcessMessageStream(tt.sessionID, tt.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessMessageStream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(chunks) == 0 {
					t.Error("ProcessMessageStream() returned empty chunks")
					return
				}

				var fullContent string
				for _, chunk := range chunks {
					fullContent += chunk.Content
					if chunk.Error != "" {
						t.Errorf("Chunk contains error: %s", chunk.Error)
					}
				}
			}
		})
	}
}

func TestProcessMessageStreamWithSession(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pc-stream-session-test-*")
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

	client := gateway.NewRpcClient(socketPath)
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	sessionID := "stream-multi-message"

	chunks1, err := client.ProcessMessageStream(sessionID, "First message")
	if err != nil {
		t.Fatalf("First ProcessMessageStream failed: %v", err)
	}
	for range chunks1 {
	}

	chunks2, err := client.ProcessMessageStream(sessionID, "Second message")
	if err != nil {
		t.Fatalf("Second ProcessMessageStream failed: %v", err)
	}

	var content2 string
	for _, chunk := range chunks2 {
		content2 += chunk.Content
	}

	_ = content2
}

func TestProcessMessageStreamSequential(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pc-stream-sequential-test-*")
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

	client := gateway.NewRpcClient(socketPath)
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	for i := 0; i < 3; i++ {
		sessionID := "sequential-session-" + string(rune('A'+i))
		chunks, err := client.ProcessMessageStream(sessionID, "Sequential message")
		if err != nil {
			t.Errorf("Sequential request %d failed: %v", i, err)
			continue
		}
		if len(chunks) == 0 {
			t.Errorf("Sequential request %d returned empty chunks", i)
		}
	}
}
