package gateway

import (
	"fmt"
	"sync"
)

// InputSource represents an input source (TUI, Channel, etc.)
type InputSource interface {
	// Start starts the input source
	Start() error
	// Stop stops the input source
	Stop() error
	// Type returns the source type (e.g., "tui", "telegram")
	Type() string
	// MessageChannel returns the channel for receiving messages
	MessageChannel() <-chan *RoutedMessage
	// SendResponse sends a response back to the source
	SendResponse(response *RoutedResponse) error
}

// RoutedMessage represents a message from an input source
type RoutedMessage struct {
	SourceType string // "tui", "telegram", etc.
	SourceID   string // Session ID, Chat ID, etc.
	Content    string // Message content
	Context    string // Additional context (e.g., file content from @ references)
	UserID     string // User identifier
	Timestamp  int64  // Unix timestamp
}

// ToEngineMessage converts the routed message to the format expected by Engine
func (my *RoutedMessage) ToEngineMessage() string {
	if my.Context != "" {
		return my.Content + "\n\n---\nContext:\n" + my.Context
	}
	return my.Content
}

// BuildResponse creates a response message for this source
func (my *RoutedMessage) BuildResponse(content string) *RoutedResponse {
	return &RoutedResponse{
		SourceType: my.SourceType,
		SourceID:   my.SourceID,
		Content:    content,
	}
}

// RoutedResponse represents a response to be sent back to an input source
type RoutedResponse struct {
	SourceType string // "tui", "telegram", etc.
	SourceID   string // Session ID, Chat ID, etc.
	Content    string // Response content
}

// InputSourceRegistry manages multiple input sources
type InputSourceRegistry struct {
	sources map[string]InputSource
	mu      sync.RWMutex
}

// NewInputSourceRegistry creates a new input source registry
func NewInputSourceRegistry() *InputSourceRegistry {
	return &InputSourceRegistry{
		sources: make(map[string]InputSource),
	}
}

// Register registers an input source
func (my *InputSourceRegistry) Register(sourceType string, source InputSource) error {
	my.mu.Lock()
	defer my.mu.Unlock()

	if _, exists := my.sources[sourceType]; exists {
		return fmt.Errorf("input source already registered: %s", sourceType)
	}

	my.sources[sourceType] = source
	return nil
}

// Unregister unregisters an input source
func (my *InputSourceRegistry) Unregister(sourceType string) error {
	my.mu.Lock()
	defer my.mu.Unlock()

	if _, exists := my.sources[sourceType]; !exists {
		return fmt.Errorf("input source not found: %s", sourceType)
	}

	delete(my.sources, sourceType)
	return nil
}

// Get gets an input source by type
func (my *InputSourceRegistry) Get(sourceType string) (InputSource, bool) {
	my.mu.RLock()
	defer my.mu.RUnlock()

	source, exists := my.sources[sourceType]
	return source, exists
}

// List returns all registered input source types
func (my *InputSourceRegistry) List() []string {
	my.mu.RLock()
	defer my.mu.RUnlock()

	types := make([]string, 0, len(my.sources))
	for t := range my.sources {
		types = append(types, t)
	}
	return types
}

// MessageHandler is a function that handles routed messages
type MessageHandler func(msg *RoutedMessage)

// MessageRouter routes messages from input sources to handlers
type MessageRouter struct {
	handlers map[string]MessageHandler
	mu       sync.RWMutex
}

// NewMessageRouter creates a new message router
func NewMessageRouter() *MessageRouter {
	return &MessageRouter{
		handlers: make(map[string]MessageHandler),
	}
}

// RegisterHandler registers a handler for a source type
func (my *MessageRouter) RegisterHandler(sourceType string, handler MessageHandler) {
	my.mu.Lock()
	defer my.mu.Unlock()

	my.handlers[sourceType] = handler
}

// UnregisterHandler unregisters a handler for a source type
func (my *MessageRouter) UnregisterHandler(sourceType string) {
	my.mu.Lock()
	defer my.mu.Unlock()

	delete(my.handlers, sourceType)
}

// Route routes a message to its handler
func (my *MessageRouter) Route(msg *RoutedMessage) {
	my.mu.RLock()
	handler, exists := my.handlers[msg.SourceType]
	my.mu.RUnlock()

	if !exists {
		// No handler for this source type
		return
	}

	// Handle in goroutine to avoid blocking
	go handler(msg)
}
