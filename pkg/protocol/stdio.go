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
func (p *StdioProtocol) Connect() error {
	// Setup pipes
	stdin, err := p.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	p.stdin = stdin

	stdout, err := p.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	p.stdout = stdout

	stderr, err := p.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	p.stderr = stderr

	// Start the command
	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Start response reader
	go p.readResponses()
	go p.readErrors()

	return nil
}

// Call invokes a method with parameters and returns the result.
func (p *StdioProtocol) Call(method string, params any) (any, error) {
	if p.closed {
		return nil, fmt.Errorf("protocol is closed")
	}

	// Create request
	req := NewRequest(method, params)

	// Create response channel
	respChan := make(chan *Response, 1)

	// Register pending request
	p.mu.Lock()
	p.pendingReqs[req.ID] = respChan
	p.mu.Unlock()

	defer func() {
		p.mu.Lock()
		delete(p.pendingReqs, req.ID)
		p.mu.Unlock()
	}()

	// Encode and send request
	data, err := req.Encode()
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}

	if _, err := p.stdin.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	// Wait for response with timeout
	select {
	case resp := <-respChan:
		if resp.IsError() {
			return nil, fmt.Errorf("plugin error: %s", resp.Error.Message)
		}
		return resp.Result, nil
	case <-time.After(p.timeout):
		return nil, fmt.Errorf("timeout waiting for response")
	}
}

// Close terminates the connection.
func (p *StdioProtocol) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}
	p.closed = true

	// Close pipes
	if p.stdin != nil {
		p.stdin.Close()
	}
	if p.stdout != nil {
		p.stdout.Close()
	}
	if p.stderr != nil {
		p.stderr.Close()
	}

	// Close response channel
	close(p.responseChan)

	return nil
}

// readResponses reads responses from stdout.
func (p *StdioProtocol) readResponses() {
	scanner := bufio.NewScanner(p.stdout)

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
		p.mu.RLock()
		respChan, ok := p.pendingReqs[resp.ID]
		p.mu.RUnlock()

		if ok {
			respChan <- resp
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error reading stdout: %v\n", err)
	}
}

// readErrors reads errors from stderr.
func (p *StdioProtocol) readErrors() {
	scanner := bufio.NewScanner(p.stderr)

	for scanner.Scan() {
		// Output stderr to our stderr
		fmt.Fprintln(os.Stderr, scanner.Text())
	}
}

// SetTimeout sets the timeout for calls.
func (p *StdioProtocol) SetTimeout(timeout time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.timeout = timeout
}

// GetTimeout returns the current timeout.
func (p *StdioProtocol) GetTimeout() time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.timeout
}
