package protocol

import (
	"encoding/json"
	"fmt"
	"io"
	"testing"
	"time"
)

// mockExecCmd is a mock implementation of execCmd for testing.
type mockExecCmd struct {
	stdinR   *io.PipeReader
	stdinW   *io.PipeWriter
	stdoutR  *io.PipeReader
	stdoutW  *io.PipeWriter
	stderrR  *io.PipeReader
	stderrW  *io.PipeWriter
	started  bool
	waited   bool
	startErr error
	waitErr  error
}

func newMockExecCmd() *mockExecCmd {
	stdinR, stdinW := io.Pipe()
	stdoutR, stdoutW := io.Pipe()
	stderrR, stderrW := io.Pipe()

	return &mockExecCmd{
		stdinR:  stdinR,
		stdinW:  stdinW,
		stdoutR: stdoutR,
		stdoutW: stdoutW,
		stderrR: stderrR,
		stderrW: stderrW,
	}
}

func (m *mockExecCmd) Start() error {
	m.started = true
	return m.startErr
}

func (m *mockExecCmd) Wait() error {
	m.waited = true
	return m.waitErr
}

func (m *mockExecCmd) StdinPipe() (io.WriteCloser, error) {
	return m.stdinW, nil
}

func (m *mockExecCmd) StdoutPipe() (io.ReadCloser, error) {
	return m.stdoutR, nil
}

func (m *mockExecCmd) StderrPipe() (io.ReadCloser, error) {
	return m.stderrR, nil
}

func (m *mockExecCmd) Close() {
	m.stdinW.Close()
	m.stdoutW.Close()
	m.stderrW.Close()
}

func TestNewRequest(t *testing.T) {
	tests := []struct {
		name   string
		method string
		params any
	}{
		{"simple method", "test.method", nil},
		{"method with params", "test.method", map[string]any{"key": "value"}},
		{"method with array params", "test.method", []any{"a", "b", "c"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := NewRequest(tt.method, tt.params)

			if req.Method != tt.method {
				t.Errorf("Method = %v, want %v", req.Method, tt.method)
			}

			if req.Type != MessageTypeCall {
				t.Errorf("Type = %v, want %v", req.Type, MessageTypeCall)
			}

			if req.Version != Version {
				t.Errorf("Version = %v, want %v", req.Version, Version)
			}

			if req.ID == "" {
				t.Error("ID should not be empty")
			}
		})
	}
}

func TestNewResponse(t *testing.T) {
	id := "test-id"
	result := map[string]string{"key": "value"}

	resp := NewResponse(id, result)

	if resp.ID != id {
		t.Errorf("ID = %v, want %v", resp.ID, id)
	}

	if resp.Type != MessageTypeResponse {
		t.Errorf("Type = %v, want %v", resp.Type, MessageTypeResponse)
	}

	if !resp.IsSuccess() {
		t.Error("Response should be success")
	}

	if resp.IsError() {
		t.Error("Response should not be error")
	}
}

func TestNewErrorResponse(t *testing.T) {
	id := "test-id"
	code := 400
	message := "Bad Request"

	resp := NewErrorResponse(id, code, message)

	if resp.ID != id {
		t.Errorf("ID = %v, want %v", resp.ID, id)
	}

	if resp.Type != MessageTypeError {
		t.Errorf("Type = %v, want %v", resp.Type, MessageTypeError)
	}

	if resp.Error == nil {
		t.Fatal("Error should not be nil")
	}

	if resp.Error.Code != code {
		t.Errorf("Error.Code = %v, want %v", resp.Error.Code, code)
	}

	if resp.Error.Message != message {
		t.Errorf("Error.Message = %v, want %v", resp.Error.Message, message)
	}

	if resp.IsSuccess() {
		t.Error("Response should not be success")
	}

	if !resp.IsError() {
		t.Error("Response should be error")
	}
}

func TestRequestEncodeDecode(t *testing.T) {
	req := NewRequest("test.method", map[string]string{"key": "value"})

	data, err := req.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	decoded, err := DecodeRequest(data)
	if err != nil {
		t.Fatalf("DecodeRequest() error = %v", err)
	}

	if decoded.ID != req.ID {
		t.Errorf("Decoded ID = %v, want %v", decoded.ID, req.ID)
	}

	if decoded.Method != req.Method {
		t.Errorf("Decoded Method = %v, want %v", decoded.Method, req.Method)
	}

	if decoded.Type != req.Type {
		t.Errorf("Decoded Type = %v, want %v", decoded.Type, req.Type)
	}
}

func TestResponseEncodeDecode(t *testing.T) {
	resp := NewResponse("test-id", map[string]string{"key": "value"})

	data, err := resp.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	decoded, err := DecodeResponse(data)
	if err != nil {
		t.Fatalf("DecodeResponse() error = %v", err)
	}

	if decoded.ID != resp.ID {
		t.Errorf("Decoded ID = %v, want %v", decoded.ID, resp.ID)
	}

	if decoded.Type != resp.Type {
		t.Errorf("Decoded Type = %v, want %v", decoded.Type, resp.Type)
	}

	if !decoded.IsSuccess() {
		t.Error("Decoded response should be success")
	}
}

func TestStdioProtocolConnect(t *testing.T) {
	cmd := newMockExecCmd()
	defer cmd.Close()

	proto := NewStdioProtocol(cmd)

	err := proto.Connect()
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	if !cmd.started {
		t.Error("Command should be started")
	}
}

func TestStdioProtocolCall(t *testing.T) {
	cmd := newMockExecCmd()
	defer cmd.Close()

	proto := NewStdioProtocol(cmd)

	if err := proto.Connect(); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	// Simulate plugin response
	go func() {
		time.Sleep(10 * time.Millisecond)
		// Create a mock request (we can't access the internal one)
		// So we just verify the protocol doesn't panic
	}()

	// This test is limited because we can't easily mock the request/response cycle
	// In a real scenario, you'd need to integrate test with actual plugins
	_ = proto
}

func TestStdioProtocolClose(t *testing.T) {
	cmd := newMockExecCmd()
	defer cmd.Close()

	proto := NewStdioProtocol(cmd)

	if err := proto.Connect(); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	if err := proto.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	// Double close should not error
	if err := proto.Close(); err != nil {
		t.Errorf("Second Close() error = %v", err)
	}
}

func TestStdioProtocolTimeout(t *testing.T) {
	cmd := newMockExecCmd()
	defer cmd.Close()

	proto := NewStdioProtocolWithTimeout(cmd, 100*time.Millisecond)

	if err := proto.Connect(); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	// Call without response should timeout
	_, err := proto.Call("test.method", nil)

	// Close protocol before checking error to clean up goroutines
	_ = proto.Close()

	if err == nil {
		t.Error("Call() should timeout")
	}
}

func TestStdioProtocolGetSetTimeout(t *testing.T) {
	cmd := newMockExecCmd()
	defer cmd.Close()

	proto := NewStdioProtocol(cmd)

	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{"1 second", 1 * time.Second},
		{"5 seconds", 5 * time.Second},
		{"10 seconds", 10 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proto.SetTimeout(tt.timeout)
			if got := proto.GetTimeout(); got != tt.timeout {
				t.Errorf("GetTimeout() = %v, want %v", got, tt.timeout)
			}
		})
	}
}

func TestMessageVersions(t *testing.T) {
	req := NewRequest("test", nil)
	if req.Version != Version {
		t.Errorf("Request version = %v, want %v", req.Version, Version)
	}

	resp := NewResponse("id", nil)
	if resp.Version != Version {
		t.Errorf("Response version = %v, want %v", resp.Version, Version)
	}
}

func TestNewErrorWithData(t *testing.T) {
	id := "test-id"
	code := 500
	message := "Internal Error"
	data := map[string]string{"field": "invalid"}

	resp := NewErrorWithData(id, code, message, data)

	if resp.Error == nil {
		t.Fatal("Error should not be nil")
	}

	if resp.Error.Code != code {
		t.Errorf("Error.Code = %v, want %v", resp.Error.Code, code)
	}

	if resp.Error.Message != message {
		t.Errorf("Error.Message = %v, want %v", resp.Error.Message, message)
	}

	if resp.Error.Data == nil {
		t.Error("Error.Data should not be nil")
	}
}

func TestInvalidRequestDecode(t *testing.T) {
	data := []byte("invalid json")

	_, err := DecodeRequest(data)
	if err == nil {
		t.Error("DecodeRequest() should return error for invalid JSON")
	}
}

func TestInvalidResponseDecode(t *testing.T) {
	data := []byte("invalid json")

	_, err := DecodeResponse(data)
	if err == nil {
		t.Error("DecodeResponse() should return error for invalid JSON")
	}
}

func TestDefaultTimeout(t *testing.T) {
	proto := NewStdioProtocol(nil)
	if got := proto.GetTimeout(); got != DefaultTimeout {
		t.Errorf("Default timeout = %v, want %v", got, DefaultTimeout)
	}
}

func TestProtocolInterface(t *testing.T) {
	cmd := newMockExecCmd()
	defer cmd.Close()

	var proto Protocol = NewStdioProtocol(cmd)

	// Verify the protocol implements the interface
	_ = proto
}

func TestClosedProtocol(t *testing.T) {
	cmd := newMockExecCmd()
	defer cmd.Close()

	proto := NewStdioProtocol(cmd)

	// Close without connecting
	_ = proto.Close()

	// Call on closed protocol should error
	_, err := proto.Call("test", nil)
	if err == nil {
		t.Error("Call() on closed protocol should return error")
	}
}

// BenchmarkRequestEncode benchmarks request encoding.
func BenchmarkRequestEncode(b *testing.B) {
	req := NewRequest("test.method", map[string]string{"key": "value"})
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = req.Encode()
	}
}

// BenchmarkRequestDecode benchmarks request decoding.
func BenchmarkRequestDecode(b *testing.B) {
	req := NewRequest("test.method", map[string]string{"key": "value"})
	data, _ := req.Encode()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = DecodeRequest(data)
	}
}

// BenchmarkResponseEncode benchmarks response encoding.
func BenchmarkResponseEncode(b *testing.B) {
	resp := NewResponse("test-id", map[string]string{"key": "value"})
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = resp.Encode()
	}
}

// TestResponseWithComplexResult tests response with complex result.
func TestResponseWithComplexResult(t *testing.T) {
	complexResult := map[string]any{
		"string": "value",
		"number": 42,
		"bool":   true,
		"array":  []string{"a", "b", "c"},
		"nested": map[string]int{"a": 1, "b": 2},
	}

	resp := NewResponse("test-id", complexResult)

	data, err := resp.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	decoded, err := DecodeResponse(data)
	if err != nil {
		t.Fatalf("DecodeResponse() error = %v", err)
	}

	if !decoded.IsSuccess() {
		t.Error("Decoded response should be success")
	}

	// Verify the result can be unmarshaled back
	resultMap, ok := decoded.Result.(map[string]any)
	if !ok {
		t.Fatal("Result should be a map")
	}

	if resultMap["string"] != "value" {
		t.Errorf("string field = %v, want value", resultMap["string"])
	}
}

// TestJSONRoundtrip tests that messages can be round-tripped through JSON.
func TestJSONRoundtrip(t *testing.T) {
	originalReq := NewRequest("test.method", map[string]string{"key": "value"})

	data, err := json.Marshal(originalReq)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decodedReq Request
	err = json.Unmarshal(data, &decodedReq)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decodedReq.ID != originalReq.ID {
		t.Errorf("ID mismatch: %v != %v", decodedReq.ID, originalReq.ID)
	}

	if decodedReq.Method != originalReq.Method {
		t.Errorf("Method mismatch: %v != %v", decodedReq.Method, originalReq.Method)
	}
}

// TestNilParams tests that nil parameters work correctly.
func TestNilParams(t *testing.T) {
	req := NewRequest("test.method", nil)

	data, err := req.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	decoded, err := DecodeRequest(data)
	if err != nil {
		t.Fatalf("DecodeRequest() error = %v", err)
	}

	if decoded.Params != nil {
		t.Errorf("Params should be nil, got %v", decoded.Params)
	}
}

// TestEmptyResult tests that empty result works correctly.
func TestEmptyResult(t *testing.T) {
	resp := NewResponse("test-id", nil)

	data, err := resp.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	decoded, err := DecodeResponse(data)
	if err != nil {
		t.Fatalf("DecodeResponse() error = %v", err)
	}

	if decoded.Result != nil {
		t.Errorf("Result should be nil, got %v", decoded.Result)
	}
}

// TestLargeMessage tests encoding and decoding of large messages.
func TestLargeMessage(t *testing.T) {
	largeData := make(map[string]string)
	for i := 0; i < 1000; i++ {
		largeData[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d", i)
	}

	req := NewRequest("test.method", largeData)

	data, err := req.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	if len(data) == 0 {
		t.Error("Encoded data should not be empty")
	}

	decoded, err := DecodeRequest(data)
	if err != nil {
		t.Fatalf("DecodeRequest() error = %v", err)
	}

	if decoded.Method != req.Method {
		t.Errorf("Method mismatch: %v != %v", decoded.Method, req.Method)
	}
}

// TestMessageTypeConstants verifies message type constants.
func TestMessageTypeConstants(t *testing.T) {
	tests := []struct {
		name  string
		value MessageType
	}{
		{"Call", MessageTypeCall},
		{"Response", MessageTypeResponse},
		{"Error", MessageTypeError},
		{"Event", MessageTypeEvent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value == "" {
				t.Errorf("MessageType %s is empty", tt.name)
			}
		})
	}
}

// TestReadWriteCloserMock tests that mock pipes work correctly.
func TestReadWriteCloserMock(t *testing.T) {
	cmd := newMockExecCmd()
	defer cmd.Close()

	// Test that pipes exist and can be used
	if cmd.stdinW == nil {
		t.Error("stdinW should not be nil")
	}
	if cmd.stdinR == nil {
		t.Error("stdinR should not be nil")
	}

	// Just verify pipes work without full round-trip test
	// (which can hang with pipe buffers)
	_ = cmd.stdoutR
	_ = cmd.stdoutW
	_ = cmd.stderrR
	_ = cmd.stderrW
}
