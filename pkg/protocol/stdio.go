package protocol

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

const (
	// DefaultTimeout is the default timeout for plugin calls.
	DefaultTimeout = 30 * time.Second
)

// StdioProtocol implements the Protocol interface using stdio (stdin/stdout).
type StdioProtocol struct {
	cmd          execCmd
	stdin        io.WriteCloser
	stdout       io.ReadCloser
	stderr       io.ReadCloser
	responseChan chan *Response
	pendingReqs  map[string]chan *Response
	mu           sync.RWMutex
	closed       bool
	timeout      time.Duration
}

// execCmd is a wrapper for exec.Cmd to avoid importing os/exec in the main package.
type execCmd interface {
	Start() error
	Wait() error
	StdinPipe() (io.WriteCloser, error)
	StdoutPipe() (io.ReadCloser, error)
	StderrPipe() (io.ReadCloser, error)
}

// NewStdioProtocol creates a new StdioProtocol for a command.
func NewStdioProtocol(cmd execCmd) *StdioProtocol {
	return &StdioProtocol{
		cmd:          cmd,
		responseChan: make(chan *Response, 100),
		pendingReqs:  make(map[string]chan *Response),
		timeout:      DefaultTimeout,
	}
}

// NewStdioProtocolWithTimeout creates a new StdioProtocol with a custom timeout.
func NewStdioProtocolWithTimeout(cmd execCmd, timeout time.Duration) *StdioProtocol {
	return &StdioProtocol{
		cmd:          cmd,
		responseChan: make(chan *Response, 100),
		pendingReqs:  make(map[string]chan *Response),
		timeout:      timeout,
	}
}

// Connect establishes the connection to the plugin.
func (my *StdioProtocol) Connect() error {
	// Setup pipes
	stdin, err := my.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	my.stdin = stdin

	stdout, err := my.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	my.stdout = stdout

	stderr, err := my.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	my.stderr = stderr

	// Start the command
	if err := my.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Start response reader
	go my.readResponses()
	go my.readErrors()

	return nil
}

// Call invokes a method with parameters and returns the result.
func (my *StdioProtocol) Call(method string, params any) (any, error) {
	if my.closed {
		return nil, fmt.Errorf("protocol is closed")
	}

	// Create request
	req := NewRequest(method, params)

	// Create response channel
	respChan := make(chan *Response, 1)

	// Register pending request
	my.mu.Lock()
	my.pendingReqs[req.ID] = respChan
	my.mu.Unlock()

	defer func() {
		my.mu.Lock()
		delete(my.pendingReqs, req.ID)
		my.mu.Unlock()
	}()

	// Encode and send request
	data, err := req.Encode()
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}

	if _, err := my.stdin.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	// Wait for response with timeout
	select {
	case resp := <-respChan:
		if resp.IsError() {
			return nil, fmt.Errorf("plugin error: %s", resp.Error.Message)
		}
		return resp.Result, nil
	case <-time.After(my.timeout):
		return nil, fmt.Errorf("timeout waiting for response")
	}
}

// Close terminates the connection.
func (my *StdioProtocol) Close() error {
	my.mu.Lock()
	defer my.mu.Unlock()

	if my.closed {
		return nil
	}
	my.closed = true

	// Close pipes
	if my.stdin != nil {
		my.stdin.Close()
	}
	if my.stdout != nil {
		my.stdout.Close()
	}
	if my.stderr != nil {
		my.stderr.Close()
	}

	// Close response channel
	close(my.responseChan)

	return nil
}

// readResponses reads responses from stdout.
func (my *StdioProtocol) readResponses() {
	scanner := bufio.NewScanner(my.stdout)

	for scanner.Scan() {
		data := scanner.Bytes()

		// Decode response
		resp, err := DecodeResponse(data)
		if err != nil {
			// Log error but continue
			fmt.Fprintf(os.Stderr, "failed to decode response: %v\n", err)
			continue
		}

		// Route response to pending request
		my.mu.RLock()
		respChan, ok := my.pendingReqs[resp.ID]
		my.mu.RUnlock()

		if ok {
			respChan <- resp
		}
	}

	if err := scanner.Err(); err != nil {
		// Ignore closed pipe errors (normal when protocol is closed)
		if err.Error() != "io: read/write on closed pipe" {
			fmt.Fprintf(os.Stderr, "error reading stdout: %v\n", err)
		}
	}
}

// readErrors reads errors from stderr.
func (my *StdioProtocol) readErrors() {
	scanner := bufio.NewScanner(my.stderr)

	for scanner.Scan() {
		// Output stderr to our stderr
		fmt.Fprintln(os.Stderr, scanner.Text())
	}
}

// SetTimeout sets the timeout for calls.
func (my *StdioProtocol) SetTimeout(timeout time.Duration) {
	my.mu.Lock()
	defer my.mu.Unlock()
	my.timeout = timeout
}

// GetTimeout returns the current timeout.
func (my *StdioProtocol) GetTimeout() time.Duration {
	my.mu.RLock()
	defer my.mu.RUnlock()
	return my.timeout
}
