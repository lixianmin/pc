package gateway

import (
	"testing"
	"time"
)

func TestInputSourceRegistry_Register(t *testing.T) {
	registry := NewInputSourceRegistry()

	tests := []struct {
		name       string
		sourceType string
		source     InputSource
		wantErr    bool
	}{
		{
			name:       "register TUI source",
			sourceType: "tui",
			source:     &mockInputSource{name: "tui"},
			wantErr:    false,
		},
		{
			name:       "register Channel source",
			sourceType: "telegram",
			source:     &mockInputSource{name: "telegram"},
			wantErr:    false,
		},
		{
			name:       "register duplicate source",
			sourceType: "tui",
			source:     &mockInputSource{name: "tui2"},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.Register(tt.sourceType, tt.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInputSourceRegistry_Unregister(t *testing.T) {
	registry := NewInputSourceRegistry()

	// Register a source
	source := &mockInputSource{name: "tui"}
	registry.Register("tui", source)

	// Unregister it
	err := registry.Unregister("tui")
	if err != nil {
		t.Errorf("Unregister() error = %v", err)
	}

	// Unregister non-existent
	err = registry.Unregister("nonexistent")
	if err == nil {
		t.Error("Unregister() should error for non-existent source")
	}
}

func TestInputSourceRegistry_Get(t *testing.T) {
	registry := NewInputSourceRegistry()
	source := &mockInputSource{name: "tui"}
	registry.Register("tui", source)

	tests := []struct {
		name       string
		sourceType string
		wantFound  bool
	}{
		{
			name:       "get existing source",
			sourceType: "tui",
			wantFound:  true,
		},
		{
			name:       "get non-existent source",
			sourceType: "unknown",
			wantFound:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, found := registry.Get(tt.sourceType)
			if found != tt.wantFound {
				t.Errorf("Get() found = %v, want %v", found, tt.wantFound)
			}
			if found && got == nil {
				t.Error("Get() returned nil source when found")
			}
		})
	}
}

func TestInputSourceRegistry_List(t *testing.T) {
	registry := NewInputSourceRegistry()

	// Initially empty
	if len(registry.List()) != 0 {
		t.Error("List() should return empty initially")
	}

	// Add sources
	registry.Register("tui", &mockInputSource{name: "tui"})
	registry.Register("telegram", &mockInputSource{name: "telegram"})

	sources := registry.List()
	if len(sources) != 2 {
		t.Errorf("List() returned %d sources, want 2", len(sources))
	}
}

func TestMessageRouter_Route(t *testing.T) {
	router := NewMessageRouter()

	// Register a handler
	var receivedMsg *RoutedMessage
	router.RegisterHandler("tui", func(msg *RoutedMessage) {
		receivedMsg = msg
	})

	// Route a message
	msg := &RoutedMessage{
		SourceType: "tui",
		SourceID:   "session-1",
		Content:    "Hello",
		UserID:     "user-1",
	}

	router.Route(msg)

	// Wait for async handling
	time.Sleep(10 * time.Millisecond)

	if receivedMsg == nil {
		t.Error("Handler was not called")
		return
	}

	if receivedMsg.Content != "Hello" {
		t.Errorf("Handler received %q, want %q", receivedMsg.Content, "Hello")
	}
}

func TestMessageRouter_Route_NoHandler(t *testing.T) {
	router := NewMessageRouter()

	// Route without handler should not panic
	msg := &RoutedMessage{
		SourceType: "unknown",
		Content:    "Hello",
	}

	router.Route(msg) // Should not panic
}

func TestRoutedMessage_ToEngineMessage(t *testing.T) {
	tests := []struct {
		name     string
		msg      *RoutedMessage
		expected string
	}{
		{
			name: "simple message",
			msg: &RoutedMessage{
				Content: "Hello",
			},
			expected: "Hello",
		},
		{
			name: "message with context",
			msg: &RoutedMessage{
				Content:    "Review this",
				Context:    "File: main.go\n```\npackage main\n```",
			},
			expected: "Review this\n\n---\nContext:\nFile: main.go\n```\npackage main\n```",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.msg.ToEngineMessage()
			if got != tt.expected {
				t.Errorf("ToEngineMessage() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestRoutedMessage_BuildResponse(t *testing.T) {
	msg := &RoutedMessage{
		SourceType: "tui",
		SourceID:   "session-1",
		Content:    "Hello",
	}

	response := msg.BuildResponse("Hi there!")

	if response.SourceType != "tui" {
		t.Errorf("Response.SourceType = %q, want %q", response.SourceType, "tui")
	}

	if response.SourceID != "session-1" {
		t.Errorf("Response.SourceID = %q, want %q", response.SourceID, "session-1")
	}

	if response.Content != "Hi there!" {
		t.Errorf("Response.Content = %q, want %q", response.Content, "Hi there!")
	}
}

// mockInputSource is a mock implementation of InputSource for testing
type mockInputSource struct {
	name      string
	started   bool
	messageCh chan *RoutedMessage
}

func (m *mockInputSource) Start() error {
	m.started = true
	m.messageCh = make(chan *RoutedMessage, 10)
	return nil
}

func (m *mockInputSource) Stop() error {
	m.started = false
	if m.messageCh != nil {
		close(m.messageCh)
	}
	return nil
}

func (m *mockInputSource) Type() string {
	return m.name
}

func (m *mockInputSource) MessageChannel() <-chan *RoutedMessage {
	return m.messageCh
}

func (m *mockInputSource) SendResponse(response *RoutedResponse) error {
	return nil
}

func TestInputSourceRegistry_WithMultipleSources(t *testing.T) {
	// Create registry and router
	registry := NewInputSourceRegistry()
	router := NewMessageRouter()

	// Register mock input sources
	tuiSource := &mockInputSource{name: "tui"}
	channelSource := &mockInputSource{name: "telegram"}

	registry.Register("tui", tuiSource)
	registry.Register("telegram", channelSource)

	// Verify sources are registered
	if len(registry.List()) != 2 {
		t.Errorf("Expected 2 input sources, got %d", len(registry.List()))
	}

	// Start sources
	tuiSource.Start()
	channelSource.Start()

	if !tuiSource.started || !channelSource.started {
		t.Error("Input sources should be started")
	}

	// Test router
	var receivedTUI, receivedChannel bool
	router.RegisterHandler("tui", func(msg *RoutedMessage) {
		receivedTUI = true
	})
	router.RegisterHandler("telegram", func(msg *RoutedMessage) {
		receivedChannel = true
	})

	// Send messages
	router.Route(&RoutedMessage{SourceType: "tui", Content: "Hello"})
	router.Route(&RoutedMessage{SourceType: "telegram", Content: "Hi"})

	// Wait for async handling
	time.Sleep(10 * time.Millisecond)

	if !receivedTUI {
		t.Error("TUI handler should be called")
	}
	if !receivedChannel {
		t.Error("Channel handler should be called")
	}

	// Stop sources
	tuiSource.Stop()
	channelSource.Stop()

	if tuiSource.started || channelSource.started {
		t.Error("Input sources should be stopped")
	}
}

