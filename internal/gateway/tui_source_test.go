package gateway

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTUIInputSource_StartStop(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get home directory")
	}

	pcDir := filepath.Join(home, ".pc")
	socketPath := filepath.Join(pcDir, "test.sock")

	source := NewTUIInputSource(socketPath)

	// Test Start
	// Note: Start will fail because no actual socket server is running,
	// but we can test the initialization logic
	err = source.Start()
	// We expect an error since there's no server, but the source should be initialized
	if source.messageCh == nil {
		t.Error("Message channel should be initialized")
	}

	// Test Stop
	err = source.Stop()
	if err != nil {
		t.Errorf("Stop() error = %v", err)
	}
}

func TestTUIInputSource_Type(t *testing.T) {
	source := &TUIInputSource{}
	if source.Type() != "tui" {
		t.Errorf("Type() = %q, want %q", source.Type(), "tui")
	}
}

func TestTUIInputSource_MessageChannel(t *testing.T) {
	source := &TUIInputSource{
		messageCh: make(chan *RoutedMessage, 10),
	}

	ch := source.MessageChannel()
	if ch == nil {
		t.Error("MessageChannel() should not return nil")
	}
}

func TestTUIInputSource_SendResponse(t *testing.T) {
	source := NewTUIInputSource("/tmp/test.sock")
	source.Start()
	defer source.Stop()

	response := &RoutedResponse{
		SourceType: "tui",
		SourceID:   "session-1",
		Content:    "Hello from agent",
	}

	// SendResponse should store the response for retrieval
	err := source.SendResponse(response)
	if err != nil {
		t.Errorf("SendResponse() error = %v", err)
	}

	// Verify response is stored
	select {
	case resp := <-source.responseCh:
		if resp.Content != "Hello from agent" {
			t.Errorf("Response content = %q, want %q", resp.Content, "Hello from agent")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Response should be stored in responseCh")
	}
}

func TestTUIInputSource_GenerateSessionID(t *testing.T) {
	sessionID1 := generateSessionID()
	sessionID2 := generateSessionID()

	if sessionID1 == "" {
		t.Error("Session ID should not be empty")
	}

	if sessionID1 == sessionID2 {
		t.Error("Session IDs should be unique")
	}

	if len(sessionID1) < 8 {
		t.Error("Session ID should be at least 8 characters")
	}
}

func TestTUIInputSource_RegisterClient(t *testing.T) {
	source := NewTUIInputSource("/tmp/test.sock")

	sessionID := "test-session-1"
	responseCh := make(chan string, 10)

	source.RegisterClient(sessionID, responseCh)

	// Verify client is registered
	source.clientsMu.RLock()
	_, exists := source.clients[sessionID]
	source.clientsMu.RUnlock()

	if !exists {
		t.Error("Client should be registered")
	}

	// Unregister client
	source.UnregisterClient(sessionID)

	source.clientsMu.RLock()
	_, exists = source.clients[sessionID]
	source.clientsMu.RUnlock()

	if exists {
		t.Error("Client should be unregistered")
	}
}

func TestTUIInputSource_RouteResponseToClient(t *testing.T) {
	source := NewTUIInputSource("/tmp/test.sock")

	sessionID := "test-session-1"
	responseCh := make(chan string, 10)

	source.RegisterClient(sessionID, responseCh)

	// Send response to the client
	response := &RoutedResponse{
		SourceType: "tui",
		SourceID:   sessionID,
		Content:    "Test response",
	}

	err := source.SendResponse(response)
	if err != nil {
		t.Errorf("SendResponse() error = %v", err)
	}

	// Wait for response to be routed
	select {
	case resp := <-responseCh:
		if resp != "Test response" {
			t.Errorf("Response = %q, want %q", resp, "Test response")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Response was not routed to client")
	}

	source.UnregisterClient(sessionID)
}
