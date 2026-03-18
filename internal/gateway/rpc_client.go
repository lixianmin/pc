package gateway

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/lixianmin/logo"
	"github.com/lixianmin/pc/pkg/protocol"
	"github.com/oklog/ulid/v2"
)

// RPCClient represents an RPC client for communicating with the gateway.
type RPCClient struct {
	socketPath string
	conn       net.Conn
	reader     *bufio.Reader
	timeout    time.Duration
}

// NewRPCClient creates a new RPC client.
func NewRPCClient(socketPath string) *RPCClient {
	return &RPCClient{
		socketPath: socketPath,
		timeout:    120 * time.Second,
	}
}

// SetTimeout sets the request timeout.
func (my *RPCClient) SetTimeout(timeout time.Duration) {
	my.timeout = timeout
}

// Connect connects to the RPC server.
func (my *RPCClient) Connect() error {
	conn, err := net.DialTimeout("unix", my.socketPath, my.timeout)
	if err != nil {
		logo.Error("[RPCClient.Connect] failed to connect to gateway: socketPath=", my.socketPath, ", err=", err)
		return fmt.Errorf("failed to connect to gateway: %w", err)
	}

	my.conn = conn
	my.reader = bufio.NewReader(conn)
	return nil
}

// Close closes the connection.
func (my *RPCClient) Close() error {
	if my.conn != nil {
		return my.conn.Close()
	}
	return nil
}

// Call makes an RPC call.
func (my *RPCClient) Call(method string, params interface{}) (*protocol.RPCResponse, error) {
	if my.conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	// Generate request ID
	requestID := ulid.Make().String()

	// Marshal params
	var paramsJSON []byte
	if params != nil {
		var err error
		paramsJSON, err = json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal params: %w", err)
		}
	}

	// Create request
	req := &protocol.RPCRequest{
		ID:     requestID,
		Method: method,
		Params: paramsJSON,
	}

	// Send request (with length prefix)
	reqData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Write length prefix (4 bytes, big-endian)
	length := int32(len(reqData))
	lengthBytes := []byte{
		byte(length >> 24),
		byte(length >> 16),
		byte(length >> 8),
		byte(length),
	}
	if _, err := my.conn.Write(lengthBytes); err != nil {
		logo.Error("[RPCClient.Call] failed to write length prefix: method=", method, ", err=", err)
		return nil, fmt.Errorf("failed to write length prefix: %w", err)
	}

	// Write request data
	if _, err := my.conn.Write(reqData); err != nil {
		logo.Error("[RPCClient.Call] failed to send request: method=", method, ", err=", err)
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Set read timeout
	if err := my.conn.SetReadDeadline(time.Now().Add(my.timeout)); err != nil {
		return nil, fmt.Errorf("failed to set read deadline: %w", err)
	}

	// Read length prefix (4 bytes)
	lengthBuf := make([]byte, 4)
	totalRead := 0
	for totalRead < 4 {
		n, err := my.reader.Read(lengthBuf[totalRead:])
		if err != nil {
			logo.Error("[RPCClient.Call] failed to read length prefix: method=", method, ", socketPath=", my.socketPath, ", err=", err)
			return nil, fmt.Errorf("failed to read length prefix: %w", err)
		}
		totalRead += n
	}

	// Parse length (big-endian)
	length = int32(lengthBuf[0])<<24 | int32(lengthBuf[1])<<16 | int32(lengthBuf[2])<<8 | int32(lengthBuf[3])
	if length <= 0 || length > 10*1024*1024 { // Max 10MB
		return nil, fmt.Errorf("invalid response length: %d", length)
	}

	// Read response data
	respData := make([]byte, length)
	totalRead = 0
	for totalRead < int(length) {
		n, err := my.reader.Read(respData[totalRead:])
		if err != nil {
			logo.Error("[RPCClient.Call] failed to read response: method=", method, ", err=", err)
			return nil, fmt.Errorf("failed to read response: %w", err)
		}
		totalRead += n
	}

	// Clear deadline
	if err := my.conn.SetReadDeadline(time.Time{}); err != nil {
		return nil, fmt.Errorf("failed to clear read deadline: %w", err)
	}

	// Unmarshal response
	var resp protocol.RPCResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		logo.Error("Failed to unmarshal response:", err)
		logo.Error("Raw response data:", string(respData))
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}

// ProcessMessage sends a message to be processed.
func (my *RPCClient) ProcessMessage(sessionID, message string) (string, error) {
	params := &protocol.ProcessMessageParams{
		SessionID: sessionID,
		Message:   message,
	}

	resp, err := my.Call(string(protocol.RPCMethodProcessMessage), params)
	if err != nil {
		return "", err
	}

	if resp.IsError() {
		return "", resp.Error
	}

	var result protocol.ProcessMessageResult
	resultJSON, err := json.Marshal(resp.Result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return result.Response, nil
}

func (my *RPCClient) ProcessMessageStream(sessionID, message string) ([]protocol.ProcessMessageStreamChunk, error) {
	params := &protocol.ProcessMessageParams{
		SessionID: sessionID,
		Message:   message,
	}

	resp, err := my.Call(string(protocol.RPCMethodProcessMessageStream), params)
	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, resp.Error
	}

	resultMap, ok := resp.Result.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected result format")
	}

	chunksRaw, ok := resultMap["chunks"].([]any)
	if !ok {
		return nil, fmt.Errorf("missing chunks in result")
	}

	var chunks []protocol.ProcessMessageStreamChunk
	for _, c := range chunksRaw {
		chunkMap, ok := c.(map[string]any)
		if !ok {
			continue
		}
		chunk := protocol.ProcessMessageStreamChunk{}
		if content, ok := chunkMap["content"].(string); ok {
			chunk.Content = content
		}
		if done, ok := chunkMap["done"].(bool); ok {
			chunk.Done = done
		}
		if errMsg, ok := chunkMap["error"].(string); ok {
			chunk.Error = errMsg
		}
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

func (my *RPCClient) GetStatus() (*protocol.GatewayStatus, error) {
	resp, err := my.Call(string(protocol.RPCMethodGetStatus), nil)
	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, resp.Error
	}

	var status protocol.GatewayStatus
	resultJSON, err := json.Marshal(resp.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	if err := json.Unmarshal(resultJSON, &status); err != nil {
		return nil, fmt.Errorf("failed to unmarshal status: %w", err)
	}

	return &status, nil
}

// ListSkills lists all available skills.
func (my *RPCClient) ListSkills() ([]protocol.SkillInfo, error) {
	resp, err := my.Call(string(protocol.RPCMethodListSkills), nil)
	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, resp.Error
	}

	var result protocol.ListSkillsResult
	resultJSON, err := json.Marshal(resp.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal skills: %w", err)
	}

	return result.Skills, nil
}

// ExecuteSkill executes a skill.
func (my *RPCClient) ExecuteSkill(name string, params map[string]interface{}) (interface{}, error) {
	reqParams := &protocol.ExecuteSkillParams{
		Name:   name,
		Params: params,
	}

	resp, err := my.Call(string(protocol.RPCMethodExecuteSkill), reqParams)
	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, resp.Error
	}

	var result protocol.ExecuteSkillResult
	resultJSON, err := json.Marshal(resp.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return result.Result, nil
}

// ListTasks lists tasks.
func (my *RPCClient) ListTasks(status string) ([]protocol.TaskInfo, error) {
	params := &protocol.ListTasksParams{}
	if status != "" {
		params.Status = status
	}

	resp, err := my.Call(string(protocol.RPCMethodListTasks), params)
	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, resp.Error
	}

	var result protocol.ListTasksResult
	resultJSON, err := json.Marshal(resp.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tasks: %w", err)
	}

	return result.Tasks, nil
}

// AddTask adds a new task.
func (my *RPCClient) AddTask(title, description string) (string, error) {
	params := &protocol.AddTaskParams{
		Title:       title,
		Description: description,
	}

	resp, err := my.Call(string(protocol.RPCMethodAddTask), params)
	if err != nil {
		return "", err
	}

	if resp.IsError() {
		return "", resp.Error
	}

	var result protocol.AddTaskResult
	resultJSON, err := json.Marshal(resp.Result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return result.ID, nil
}

// CompleteTask completes a task.
func (my *RPCClient) CompleteTask(id string) error {
	params := &protocol.CompleteTaskParams{
		ID: id,
	}

	resp, err := my.Call(string(protocol.RPCMethodCompleteTask), params)
	if err != nil {
		return err
	}

	if resp.IsError() {
		return resp.Error
	}

	return nil
}

// DeleteTask deletes a task.
func (my *RPCClient) DeleteTask(id string) error {
	params := &protocol.DeleteTaskParams{
		ID: id,
	}

	resp, err := my.Call(string(protocol.RPCMethodDeleteTask), params)
	if err != nil {
		return err
	}

	if resp.IsError() {
		return resp.Error
	}

	return nil
}
