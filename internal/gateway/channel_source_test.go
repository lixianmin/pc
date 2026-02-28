package gateway

import (
	"testing"
	"time"
)

func TestChannelInputSource_StartStop(t *testing.T) {
	source := NewChannelInputSource("telegram")

	// Test Start
	err := source.Start()
	if err != nil {
		t.Errorf("Start() error = %v", err)
	}

	if !source.IsRunning() {
		t.Error("Source should be running after Start()")
	}

	// Test Stop
	err = source.Stop()
	if err != nil {
		t.Errorf("Stop() error = %v", err)
	}

	if source.IsRunning() {
		t.Error("Source should not be running after Stop()")
	}
}

func TestChannelInputSource_Type(t *testing.T) {
	source := NewChannelInputSource("telegram")
	if source.Type() != "telegram" {
		t.Errorf("Type() = %q, want %q", source.Type(), "telegram")
	}

	discordSource := NewChannelInputSource("discord")
	if discordSource.Type() != "discord" {
		t.Errorf("Type() = %q, want %q", discordSource.Type(), "discord")
	}
}

func TestChannelInputSource_MessageChannel(t *testing.T) {
	source := NewChannelInputSource("telegram")
	source.Start()
	defer source.Stop()

	ch := source.MessageChannel()
	if ch == nil {
		t.Error("MessageChannel() should not return nil")
	}
}

func TestChannelInputSource_SendResponse(t *testing.T) {
	source := NewChannelInputSource("telegram")
	source.Start()
	defer source.Stop()

	response := &RoutedResponse{
		SourceType: "telegram",
		SourceID:   "chat-123",
		Content:    "Hello from agent",
	}

	err := source.SendResponse(response)
	if err != nil {
		t.Errorf("SendResponse() error = %v", err)
	}

	// Verify response is stored
	select {
	case <-source.responseCh:
		// Response stored
	case <-time.After(100 * time.Millisecond):
		t.Error("Response should be stored")
	}
}

func TestChannelInputSource_ReceiveMessage(t *testing.T) {
	source := NewChannelInputSource("telegram")
	source.Start()
	defer source.Stop()

	// Simulate receiving a message from channel plugin
	source.ReceiveMessage("chat-123", "user-456", "Hello bot", "")

	// Verify message is routed
	select {
	case msg := <-source.MessageChannel():
		if msg.SourceType != "telegram" {
			t.Errorf("SourceType = %q, want %q", msg.SourceType, "telegram")
		}
		if msg.SourceID != "chat-123" {
			t.Errorf("SourceID = %q, want %q", msg.SourceID, "chat-123")
		}
		if msg.UserID != "user-456" {
			t.Errorf("UserID = %q, want %q", msg.UserID, "user-456")
		}
		if msg.Content != "Hello bot" {
			t.Errorf("Content = %q, want %q", msg.Content, "Hello bot")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Message should be routed to MessageChannel")
	}
}

func TestChannelInputSource_MultipleChannels(t *testing.T) {
	// Create multiple channel sources
	telegram := NewChannelInputSource("telegram")
	discord := NewChannelInputSource("discord")

	telegram.Start()
	discord.Start()
	defer telegram.Stop()
	defer discord.Stop()

	// Send messages to both channels
	telegram.ReceiveMessage("tg-chat-1", "user-1", "Hello from Telegram", "")
	discord.ReceiveMessage("dc-chat-1", "user-2", "Hello from Discord", "")

	// Collect messages
	var telegramReceived, discordReceived bool
	timeout := time.After(200 * time.Millisecond)

	for i := 0; i < 2; i++ {
		select {
		case msg := <-telegram.MessageChannel():
			if msg.Content == "Hello from Telegram" {
				telegramReceived = true
			}
		case msg := <-discord.MessageChannel():
			if msg.Content == "Hello from Discord" {
				discordReceived = true
			}
		case <-timeout:
			// Timeout
		}
	}

	if !telegramReceived {
		t.Error("Telegram message not received")
	}
	if !discordReceived {
		t.Error("Discord message not received")
	}
}

func TestChannelInputSource_ResponseRouting(t *testing.T) {
	source := NewChannelInputSource("telegram")
	source.Start()
	defer source.Stop()

	// Setup response collector
	responses := make(map[string]string)
	go func() {
		for resp := range source.responseCh {
			responses[resp.SourceID] = resp.Content
		}
	}()

	// Send responses to different chats
	source.SendResponse(&RoutedResponse{
		SourceType: "telegram",
		SourceID:   "chat-1",
		Content:    "Response to chat 1",
	})

	source.SendResponse(&RoutedResponse{
		SourceType: "telegram",
		SourceID:   "chat-2",
		Content:    "Response to chat 2",
	})

	time.Sleep(50 * time.Millisecond)

	if responses["chat-1"] != "Response to chat 1" {
		t.Errorf("chat-1 response = %q, want %q", responses["chat-1"], "Response to chat 1")
	}

	if responses["chat-2"] != "Response to chat 2" {
		t.Errorf("chat-2 response = %q, want %q", responses["chat-2"], "Response to chat 2")
	}
}

func TestChannelInputSource_WithContext(t *testing.T) {
	source := NewChannelInputSource("telegram")
	source.Start()
	defer source.Stop()

	// Receive message with context
	source.ReceiveMessage("chat-123", "user-456", "Review this code", "File: main.go\n```\npackage main\n```")

	select {
	case msg := <-source.MessageChannel():
		if msg.Context == "" {
			t.Error("Context should not be empty")
		}
		if msg.Content != "Review this code" {
			t.Errorf("Content = %q, want %q", msg.Content, "Review this code")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Message with context should be received")
	}
}
