package gateway

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var sessionCounter uint64

// TUIInputSource implements InputSource for TUI clients via RPC
type TUIInputSource struct {
	socketPath string
	messageCh  chan *RoutedMessage
	responseCh chan *RoutedResponse
	clients    map[string]chan string // sessionID -> response channel
	clientsMu  sync.RWMutex
	running    bool
	mu         sync.RWMutex
}

// NewTUIInputSource creates a new TUI input source
func NewTUIInputSource(socketPath string) *TUIInputSource {
	return &TUIInputSource{
		socketPath: socketPath,
		messageCh:  make(chan *RoutedMessage, 100),
		responseCh: make(chan *RoutedResponse, 100),
		clients:    make(map[string]chan string),
	}
}

// Start starts the TUI input source
func (t *TUIInputSource) Start() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.running {
		return fmt.Errorf("TUI input source already running")
	}

	t.running = true

	// Start response routing goroutine
	go t.routeResponses()

	return nil
}

// Stop stops the TUI input source
func (t *TUIInputSource) Stop() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.running {
		return nil
	}

	t.running = false

	// Close channels
	close(t.responseCh)

	// Unregister all clients
	t.clientsMu.Lock()
	for sessionID, ch := range t.clients {
		close(ch)
		delete(t.clients, sessionID)
	}
	t.clientsMu.Unlock()

	return nil
}

// Type returns the source type
func (t *TUIInputSource) Type() string {
	return "tui"
}

// MessageChannel returns the channel for receiving messages
func (t *TUIInputSource) MessageChannel() <-chan *RoutedMessage {
	return t.messageCh
}

// SendResponse sends a response back to the TUI client
func (t *TUIInputSource) SendResponse(response *RoutedResponse) error {
	t.mu.RLock()
	running := t.running
	t.mu.RUnlock()

	// If running, route through responseCh
	if running {
		select {
		case t.responseCh <- response:
			return nil
		case <-time.After(5 * time.Second):
			return fmt.Errorf("timeout sending response")
		}
	}

	// If not running, try direct routing to registered clients
	t.clientsMu.RLock()
	ch, exists := t.clients[response.SourceID]
	t.clientsMu.RUnlock()

	if exists {
		select {
		case ch <- response.Content:
			return nil
		case <-time.After(5 * time.Second):
			return fmt.Errorf("timeout sending response")
		}
	}

	return fmt.Errorf("no client registered for session %s", response.SourceID)
}

// RegisterClient registers a TUI client session
func (t *TUIInputSource) RegisterClient(sessionID string, responseCh chan string) {
	t.clientsMu.Lock()
	defer t.clientsMu.Unlock()

	t.clients[sessionID] = responseCh
}

// UnregisterClient unregisters a TUI client session
func (t *TUIInputSource) UnregisterClient(sessionID string) {
	t.clientsMu.Lock()
	defer t.clientsMu.Unlock()

	if ch, exists := t.clients[sessionID]; exists {
		close(ch)
		delete(t.clients, sessionID)
	}
}

// ReceiveMessage receives a message from a TUI client and routes it to the engine
func (t *TUIInputSource) ReceiveMessage(sessionID, userID, content, context string) {
	msg := &RoutedMessage{
		SourceType: "tui",
		SourceID:   sessionID,
		Content:    content,
		Context:    context,
		UserID:     userID,
		Timestamp:  time.Now().Unix(),
	}

	select {
	case t.messageCh <- msg:
	case <-time.After(5 * time.Second):
		// Channel full, drop message
	}
}

// routeResponses routes responses to the appropriate client
func (t *TUIInputSource) routeResponses() {
	for response := range t.responseCh {
		t.clientsMu.RLock()
		ch, exists := t.clients[response.SourceID]
		t.clientsMu.RUnlock()

		if exists {
			select {
			case ch <- response.Content:
			case <-time.After(5 * time.Second):
				// Client not receiving, might be disconnected
			}
		}
	}
}

// IsRunning returns whether the input source is running
func (t *TUIInputSource) IsRunning() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.running
}

// generateSessionID generates a unique session ID
func generateSessionID() string {
	counter := atomic.AddUint64(&sessionCounter, 1)
	return fmt.Sprintf("tui-%d-%d", time.Now().UnixNano(), counter)
}
