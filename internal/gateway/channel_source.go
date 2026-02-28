package gateway

import (
	"fmt"
	"sync"
	"time"
)

// ChannelInputSource implements InputSource for Channel plugins (Telegram, Discord, etc.)
type ChannelInputSource struct {
	channelType string
	messageCh   chan *RoutedMessage
	responseCh  chan *RoutedResponse
	running     bool
	mu          sync.RWMutex
}

// NewChannelInputSource creates a new channel input source
func NewChannelInputSource(channelType string) *ChannelInputSource {
	return &ChannelInputSource{
		channelType: channelType,
		messageCh:   make(chan *RoutedMessage, 100),
		responseCh:  make(chan *RoutedResponse, 100),
	}
}

// Start starts the channel input source
func (c *ChannelInputSource) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return fmt.Errorf("%s input source already running", c.channelType)
	}

	c.running = true

	return nil
}

// Stop stops the channel input source
func (c *ChannelInputSource) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return nil
	}

	c.running = false

	return nil
}

// Type returns the source type (e.g., "telegram", "discord")
func (c *ChannelInputSource) Type() string {
	return c.channelType
}

// MessageChannel returns the channel for receiving messages
func (c *ChannelInputSource) MessageChannel() <-chan *RoutedMessage {
	return c.messageCh
}

// SendResponse sends a response back to the channel
func (c *ChannelInputSource) SendResponse(response *RoutedResponse) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.running {
		return fmt.Errorf("%s input source not running", c.channelType)
	}

	select {
	case c.responseCh <- response:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout sending response")
	}
}

// ReceiveMessage receives a message from the channel plugin and routes it to the engine
func (c *ChannelInputSource) ReceiveMessage(sourceID, userID, content, context string) {
	c.mu.RLock()
	running := c.running
	c.mu.RUnlock()

	if !running {
		return
	}

	msg := &RoutedMessage{
		SourceType: c.channelType,
		SourceID:   sourceID,
		Content:    content,
		Context:    context,
		UserID:     userID,
		Timestamp:  time.Now().Unix(),
	}

	select {
	case c.messageCh <- msg:
	case <-time.After(5 * time.Second):
		// Channel full, drop message
	}
}

// IsRunning returns whether the input source is running
func (c *ChannelInputSource) IsRunning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.running
}

// GetResponseChannel returns the response channel for this source
// This is used by the channel plugin adapter to receive responses
func (c *ChannelInputSource) GetResponseChannel() <-chan *RoutedResponse {
	return c.responseCh
}
