package gateway

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestInputSourceCoexistence verifies that TUI and Channel sources can coexist
func TestInputSourceCoexistence(t *testing.T) {
	// Create input source registry and message router
	registry := NewInputSourceRegistry()
	router := NewMessageRouter()

	// Create TUI source
	tuiSource := NewTUIInputSource("/tmp/test.sock")

	// Create Channel sources
	telegramSource := NewChannelInputSource("telegram")
	discordSource := NewChannelInputSource("discord")

	// Register all sources
	if err := registry.Register("tui", tuiSource); err != nil {
		t.Fatalf("Failed to register TUI source: %v", err)
	}
	if err := registry.Register("telegram", telegramSource); err != nil {
		t.Fatalf("Failed to register telegram source: %v", err)
	}
	if err := registry.Register("discord", discordSource); err != nil {
		t.Fatalf("Failed to register discord source: %v", err)
	}

	// Verify all sources are registered
	sources := registry.List()
	if len(sources) != 3 {
		t.Errorf("Expected 3 sources, got %d", len(sources))
	}

	// Start all sources
	if err := tuiSource.Start(); err != nil {
		t.Errorf("Failed to start TUI source: %v", err)
	}
	if err := telegramSource.Start(); err != nil {
		t.Errorf("Failed to start telegram source: %v", err)
	}
	if err := discordSource.Start(); err != nil {
		t.Errorf("Failed to start discord source: %v", err)
	}

	// Setup message handlers
	messageCounts := make(map[string]int)
	var countMu sync.Mutex
	tuiResponseCh := make(chan string, 10)

	router.RegisterHandler("tui", func(msg *RoutedMessage) {
		countMu.Lock()
		messageCounts["tui"]++
		countMu.Unlock()
		// Send response back
		response := msg.BuildResponse("TUI response: " + msg.Content)
		tuiSource.SendResponse(response)
	})

	router.RegisterHandler("telegram", func(msg *RoutedMessage) {
		countMu.Lock()
		messageCounts["telegram"]++
		countMu.Unlock()
		response := msg.BuildResponse("Telegram response: " + msg.Content)
		telegramSource.SendResponse(response)
	})

	router.RegisterHandler("discord", func(msg *RoutedMessage) {
		countMu.Lock()
		messageCounts["discord"]++
		countMu.Unlock()
		response := msg.BuildResponse("Discord response: " + msg.Content)
		discordSource.SendResponse(response)
	})

	// Register TUI client
	tuiSource.RegisterClient("tui-session-1", tuiResponseCh)

	// Simulate messages from different sources
	tuiSource.ReceiveMessage("tui-session-1", "user-1", "Hello from TUI", "")
	telegramSource.ReceiveMessage("tg-chat-1", "user-2", "Hello from Telegram", "")
	discordSource.ReceiveMessage("dc-chat-1", "user-3", "Hello from Discord", "")

	// Route all messages
	for i := 0; i < 3; i++ {
		select {
		case msg := <-tuiSource.MessageChannel():
			router.Route(msg)
		case msg := <-telegramSource.MessageChannel():
			router.Route(msg)
		case msg := <-discordSource.MessageChannel():
			router.Route(msg)
		case <-time.After(100 * time.Millisecond):
			// No more messages
		}
	}

	// Wait for routing
	time.Sleep(50 * time.Millisecond)

	// Verify messages were routed
	countMu.Lock()
	if messageCounts["tui"] != 1 {
		t.Errorf("Expected 1 TUI message, got %d", messageCounts["tui"])
	}
	if messageCounts["telegram"] != 1 {
		t.Errorf("Expected 1 Telegram message, got %d", messageCounts["telegram"])
	}
	if messageCounts["discord"] != 1 {
		t.Errorf("Expected 1 Discord message, got %d", messageCounts["discord"])
	}
	countMu.Unlock()

	// Verify TUI response was routed to client
	select {
	case resp := <-tuiResponseCh:
		if resp != "TUI response: Hello from TUI" {
			t.Errorf("TUI response = %q, want %q", resp, "TUI response: Hello from TUI")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("TUI response not received")
	}

	// Cleanup
	tuiSource.UnregisterClient("tui-session-1")
	tuiSource.Stop()
	telegramSource.Stop()
	discordSource.Stop()
}

// TestTUIWithContext verifies TUI messages with file context
func TestTUIWithContext(t *testing.T) {
	tuiSource := NewTUIInputSource("/tmp/test.sock")
	tuiSource.Start()
	defer tuiSource.Stop()

	handlerCalled := false
	router := NewMessageRouter()
	router.RegisterHandler("tui", func(msg *RoutedMessage) {
		handlerCalled = true
		if msg.SourceType != "tui" {
			t.Errorf("SourceType = %q, want %q", msg.SourceType, "tui")
		}
		if msg.SourceID != "session-1" {
			t.Errorf("SourceID = %q, want %q", msg.SourceID, "session-1")
		}
		if msg.UserID != "user-1" {
			t.Errorf("UserID = %q, want %q", msg.UserID, "user-1")
		}
		if msg.Content != "Review this code" {
			t.Errorf("Content = %q, want %q", msg.Content, "Review this code")
		}
		if msg.Context == "" {
			t.Error("Context should not be empty")
		}
	})

	// Send message with context
	tuiSource.ReceiveMessage("session-1", "user-1", "Review this code", "File: main.go\n```\npackage main\n```")

	// Route message
	select {
	case msg := <-tuiSource.MessageChannel():
		router.Route(msg)
	case <-time.After(100 * time.Millisecond):
		t.Error("Message not received")
	}

	time.Sleep(10 * time.Millisecond)

	if !handlerCalled {
		t.Error("Handler should be called")
	}
}

// TestMultipleChannelPlugins verifies multiple channel plugins work together
func TestMultipleChannelPlugins(t *testing.T) {
	registry := NewInputSourceRegistry()

	// Create multiple channel sources
	channels := []string{"telegram", "discord", "slack", "wechat"}
	for _, ch := range channels {
		source := NewChannelInputSource(ch)
		if err := registry.Register(ch, source); err != nil {
			t.Fatalf("Failed to register %s: %v", ch, err)
		}
		if err := source.Start(); err != nil {
			t.Fatalf("Failed to start %s: %v", ch, err)
		}
		defer source.Stop()
	}

	// Verify registration
	if len(registry.List()) != len(channels) {
		t.Errorf("Expected %d channels, got %d", len(channels), len(registry.List()))
	}

	// Test each channel
	for _, chType := range channels {
		source, exists := registry.Get(chType)
		if !exists {
			t.Errorf("Channel %s not found in registry", chType)
			continue
		}

		if source.Type() != chType {
			t.Errorf("Channel type = %q, want %q", source.Type(), chType)
		}

		// Type assert to check running state
		if chSource, ok := source.(*ChannelInputSource); ok {
			if !chSource.IsRunning() {
				t.Errorf("Channel %s should be running", chType)
			}
		}
	}
}

// TestInputSourceLifecycle verifies proper start/stop lifecycle
func TestInputSourceLifecycle(t *testing.T) {
	tuiSource := NewTUIInputSource("/tmp/test.sock")
	telegramSource := NewChannelInputSource("telegram")

	// Initially not running
	if tuiSource.IsRunning() {
		t.Error("TUI source should not be running initially")
	}
	if telegramSource.IsRunning() {
		t.Error("Telegram source should not be running initially")
	}

	// Start
	if err := tuiSource.Start(); err != nil {
		t.Errorf("Failed to start TUI: %v", err)
	}
	if err := telegramSource.Start(); err != nil {
		t.Errorf("Failed to start Telegram: %v", err)
	}

	if !tuiSource.IsRunning() {
		t.Error("TUI source should be running after Start()")
	}
	if !telegramSource.IsRunning() {
		t.Error("Telegram source should be running after Start()")
	}

	// Double start should error
	if err := tuiSource.Start(); err == nil {
		t.Error("Double Start() should return error")
	}
	if err := telegramSource.Start(); err == nil {
		t.Error("Double Start() should return error")
	}

	// Stop
	if err := tuiSource.Stop(); err != nil {
		t.Errorf("Failed to stop TUI: %v", err)
	}
	if err := telegramSource.Stop(); err != nil {
		t.Errorf("Failed to stop Telegram: %v", err)
	}

	if tuiSource.IsRunning() {
		t.Error("TUI source should not be running after Stop()")
	}
	if telegramSource.IsRunning() {
		t.Error("Telegram source should not be running after Stop()")
	}

	// Double stop should not error
	if err := tuiSource.Stop(); err != nil {
		t.Errorf("Double Stop() should not error: %v", err)
	}
	if err := telegramSource.Stop(); err != nil {
		t.Errorf("Double Stop() should not error: %v", err)
	}
}

// TestMessageRoutingConcurrency verifies concurrent message routing
func TestMessageRoutingConcurrency(t *testing.T) {
	tuiSource := NewTUIInputSource("/tmp/test.sock")
	tuiSource.Start()
	defer tuiSource.Stop()

	telegramSource := NewChannelInputSource("telegram")
	telegramSource.Start()
	defer telegramSource.Stop()

	router := NewMessageRouter()

	messageCount := 0
	var mu sync.Mutex

	router.RegisterHandler("tui", func(msg *RoutedMessage) {
		mu.Lock()
		messageCount++
		mu.Unlock()
	})

	router.RegisterHandler("telegram", func(msg *RoutedMessage) {
		mu.Lock()
		messageCount++
		mu.Unlock()
	})

	// Send many messages concurrently
	for i := 0; i < 50; i++ {
		go func(n int) {
			tuiSource.ReceiveMessage(fmt.Sprintf("session-%d", n), "user", "message", "")
		}(i)
		go func(n int) {
			telegramSource.ReceiveMessage(fmt.Sprintf("chat-%d", n), "user", "message", "")
		}(i)
	}

	// Route all messages
	time.Sleep(100 * time.Millisecond)
	for i := 0; i < 100; i++ {
		select {
		case msg := <-tuiSource.MessageChannel():
			router.Route(msg)
		case msg := <-telegramSource.MessageChannel():
			router.Route(msg)
		default:
		}
	}

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	count := messageCount
	mu.Unlock()

	if count != 100 {
		t.Errorf("Expected 100 messages, got %d", count)
	}
}

