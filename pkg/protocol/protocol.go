package protocol

// Protocol defines the interface for communicating with plugins.
type Protocol interface {
	// Connect establishes the connection to the plugin.
	Connect() error

	// Call invokes a method with parameters and returns the result.
	Call(method string, params any) (any, error)

	// Close terminates the connection.
	Close() error
}
