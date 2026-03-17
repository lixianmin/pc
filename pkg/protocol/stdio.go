package protocol

import (
	"bufio"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/lixianmin/got/loom"
	"github.com/lixianmin/logo"
)

const (
	DefaultTimeout = 60 * time.Second
)

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

type execCmd interface {
	Start() error
	Wait() error
	StdinPipe() (io.WriteCloser, error)
	StdoutPipe() (io.ReadCloser, error)
	StderrPipe() (io.ReadCloser, error)
}

func NewStdioProtocol(cmd execCmd) *StdioProtocol {
	return &StdioProtocol{
		cmd:          cmd,
		responseChan: make(chan *Response, 100),
		pendingReqs:  make(map[string]chan *Response),
		timeout:      DefaultTimeout,
	}
}

func NewStdioProtocolWithTimeout(cmd execCmd, timeout time.Duration) *StdioProtocol {
	return &StdioProtocol{
		cmd:          cmd,
		responseChan: make(chan *Response, 100),
		pendingReqs:  make(map[string]chan *Response),
		timeout:      timeout,
	}
}

func (my *StdioProtocol) Connect() error {
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

	if err := my.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	loom.Go(my.readResponses)
	loom.Go(my.readErrors)

	return nil
}

func (my *StdioProtocol) Call(method string, params any) (any, error) {
	if my.closed {
		return nil, fmt.Errorf("protocol is closed")
	}

	var request = NewRequest(method, params)

	respChan := make(chan *Response, 1)

	my.mu.Lock()
	my.pendingReqs[request.Id] = respChan
	my.mu.Unlock()

	defer func() {
		my.mu.Lock()
		delete(my.pendingReqs, request.Id)
		my.mu.Unlock()
	}()

	data, err := request.Encode()
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}

	logo.Debug("[Protocol] Sending request", "id", request.Id, "method", method)

	data = append(data, '\n')
	if _, err := my.stdin.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	logo.Debug("[Protocol] Waiting for response", "id", request.Id, "timeout", my.timeout, "startTime", time.Now().Format("15:04:05"))

	select {
	case resp := <-respChan:
		logo.Debug("[Protocol] Received response", "id", resp.Id, "endTime", time.Now().Format("15:04:05"))
		if resp.IsError() {
			return nil, fmt.Errorf("plugin error: %s", resp.Error.Message)
		}
		return resp.Result, nil
	case <-time.After(my.timeout):
		logo.Warn("[Protocol] Timeout waiting for response", "id", request.Id, "timeout", my.timeout)
		return nil, fmt.Errorf("timeout waiting for response")
	}
}

func (my *StdioProtocol) Close() error {
	my.mu.Lock()
	defer my.mu.Unlock()

	if my.closed {
		return nil
	}
	my.closed = true

	if my.stdin != nil {
		my.stdin.Close()
	}
	if my.stdout != nil {
		my.stdout.Close()
	}
	if my.stderr != nil {
		my.stderr.Close()
	}

	close(my.responseChan)

	return nil
}

func (my *StdioProtocol) readResponses(later loom.Later) {
	scanner := bufio.NewScanner(my.stdout)

	for scanner.Scan() {
		var data = scanner.Bytes()
		logo.Debug("[Protocol] Raw response", "data", string(data))

		resp, err := DecodeResponse(data)
		if err != nil {
			logo.Warn("[Protocol] Failed to decode response", "error", err)
			continue
		}

		logo.Debug("[Protocol] Decoded response", "id", resp.Id, "type", resp.Type)

		my.mu.RLock()
		respChan, ok := my.pendingReqs[resp.Id]
		my.mu.RUnlock()

		if ok {
			logo.Debug("[Protocol] Routing response", "id", resp.Id)
			respChan <- resp
		} else {
			logo.Warn("[Protocol] No pending request for response", "id", resp.Id)
		}
	}

	if err := scanner.Err(); err != nil {
		if err.Error() != "io: read/write on closed pipe" {
			logo.Warn("[Protocol] Error reading stdout", "error", err)
		}
	}

	logo.Debug("[Protocol] Response reader stopped")
}

func (my *StdioProtocol) readErrors(later loom.Later) {
	scanner := bufio.NewScanner(my.stderr)

	for scanner.Scan() {
		logo.Info("[Plugin stderr]", "line", scanner.Text())
	}
}

func (my *StdioProtocol) SetTimeout(timeout time.Duration) {
	my.mu.Lock()
	defer my.mu.Unlock()
	my.timeout = timeout
}

func (my *StdioProtocol) GetTimeout() time.Duration {
	my.mu.RLock()
	defer my.mu.RUnlock()
	return my.timeout
}
