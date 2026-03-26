package protocol

import (
	"encoding/json"
	"fmt"
)

// RpcMethod represents available RPC methods.
type RpcMethod string

const (
	// RpcMethodProcessMessage processes a user message.
	RpcMethodProcessMessage RpcMethod = "ProcessMessage"
	// RpcMethodProcessMessageStream processes a user message with streaming.
	RpcMethodProcessMessageStream RpcMethod = "ProcessMessageStream"
	// RpcMethodGetStatus returns the gateway status.
	RpcMethodGetStatus RpcMethod = "GetStatus"
	// RpcMethodListSkills lists all available skills.
	RpcMethodListSkills RpcMethod = "ListSkills"
	// RpcMethodExecuteSkill executes a skill.
	RpcMethodExecuteSkill RpcMethod = "ExecuteSkill"
	// RpcMethodListTasks lists tasks.
	RpcMethodListTasks RpcMethod = "ListTasks"
	// RpcMethodAddTask adds a new task.
	RpcMethodAddTask RpcMethod = "AddTask"
	// RpcMethodCompleteTask completes a task.
	RpcMethodCompleteTask RpcMethod = "CompleteTask"
	// RpcMethodDeleteTask deletes a task.
	RpcMethodDeleteTask RpcMethod = "DeleteTask"
)

// RpcRequest represents an RPC request.
type RpcRequest struct {
	Id     string          `json:"id"`
	Method RpcMethod       `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

// RpcResponse represents an RPC response.
type RpcResponse struct {
	ID     string    `json:"id"`
	Result any       `json:"result,omitempty"`
	Error  *RpcError `json:"error,omitempty"`
}

// RpcError represents an RPC error.
type RpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// UnmarshalJSON implements custom JSON unmarshaling to handle both string and int codes.
func (my *RpcError) UnmarshalJSON(data []byte) error {
	var raw struct {
		Code    interface{} `json:"code"`
		Message string      `json:"message"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	switch v := raw.Code.(type) {
	case float64:
		my.Code = int(v)
	case string:
		// Try to parse as int
		if _, err := fmt.Sscanf(v, "%d", &my.Code); err != nil {
			my.Code = RpcErrorCodeInternalError
		}
	default:
		my.Code = RpcErrorCodeInternalError
	}

	my.Message = raw.Message
	return nil
}

// Error implements the error interface.
func (my *RpcError) Error() string {
	return fmt.Sprintf("RPC error %d: %s", my.Code, my.Message)
}

// RPC error codes.
const (
	RpcErrorCodeParseError     = -32700
	RpcErrorCodeInvalidRequest = -32600
	RpcErrorCodeMethodNotFound = -32601
	RpcErrorCodeInvalidParams  = -32602
	RpcErrorCodeInternalError  = -32603
	RpcErrorCodeServerError    = -32000
)

// NewRpcError creates a new RPC error.
func NewRpcError(code int, message string) *RpcError {
	return &RpcError{
		Code:    code,
		Message: message,
	}
}

// NewRpcResponse creates a new successful RPC response.
func NewRpcResponse(id string, result interface{}) *RpcResponse {
	return &RpcResponse{
		ID:     id,
		Result: result,
	}
}

// NewRpcErrorResponse creates a new error RPC response.
func NewRpcErrorResponse(id string, code int, message string) *RpcResponse {
	return &RpcResponse{
		ID:    id,
		Error: NewRpcError(code, message),
	}
}

// IsSuccess returns true if the response is successful.
func (my *RpcResponse) IsSuccess() bool {
	return my.Error == nil
}

// IsError returns true if the response is an error.
func (my *RpcResponse) IsError() bool {
	return my.Error != nil
}

// RPCMessage represents a user message for ProcessMessage.
type ProcessMessageParams struct {
	SessionId string `json:"sessionId"`
	Message   string `json:"message"`
}

// ProcessMessageResult represents the result of ProcessMessage.
type ProcessMessageResult struct {
	Response string `json:"response"`
}

type StreamChunkType string

const (
	ChunkTypeThinking   StreamChunkType = "thinking"
	ChunkTypeToolCall   StreamChunkType = "tool_call"
	ChunkTypeToolResult StreamChunkType = "tool_result"
	ChunkTypeResponse   StreamChunkType = "response"
	ChunkTypeDone       StreamChunkType = "done"
	ChunkTypeError      StreamChunkType = "error"
)

type ProcessMessageStreamChunk struct {
	Type    StreamChunkType `json:"type"`
	Content string          `json:"content,omitempty"`
	Tool    string          `json:"tool,omitempty"`
	Meta    any             `json:"meta,omitempty"`
	Done    bool            `json:"done"`
	Error   string          `json:"error,omitempty"`
}

// GatewayStatus represents the gateway status.
type GatewayStatus struct {
	Running     bool   `json:"running"`
	Pid         int    `json:"pid,omitempty"`
	Version     string `json:"version,omitempty"`
	Uptime      int64  `json:"uptime,omitempty"` // seconds
	PluginCount int    `json:"pluginCount,omitempty"`
}

// ListSkillsResult represents the result of ListSkills.
type ListSkillsResult struct {
	Skills []SkillInfo `json:"skills"`
}

// SkillInfo represents skill information.
type SkillInfo struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// ExecuteSkillParams represents parameters for ExecuteSkill.
type ExecuteSkillParams struct {
	Name   string         `json:"name"`
	Params map[string]any `json:"params,omitempty"`
}

// ExecuteSkillResult represents the result of ExecuteSkill.
type ExecuteSkillResult struct {
	Result any `json:"result"`
}

// ListTasksParams represents parameters for ListTasks.
type ListTasksParams struct {
	Status string `json:"status,omitempty"`
}

// ListTasksResult represents the result of ListTasks.
type ListTasksResult struct {
	Tasks []TaskInfo `json:"tasks"`
}

// TaskInfo represents task information.
type TaskInfo struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Status      string `json:"status"`
	Description string `json:"description,omitempty"`
}

// AddTaskParams represents parameters for AddTask.
type AddTaskParams struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

// AddTaskResult represents the result of AddTask.
type AddTaskResult struct {
	ID string `json:"id"`
}

// CompleteTaskParams represents parameters for CompleteTask.
type CompleteTaskParams struct {
	ID string `json:"id"`
}

// CompleteTaskResult represents the result of CompleteTask.
type CompleteTaskResult struct {
	Success bool `json:"success"`
}

// DeleteTaskParams represents parameters for DeleteTask.
type DeleteTaskParams struct {
	ID string `json:"id"`
}

// DeleteTaskResult represents the result of DeleteTask.
type DeleteTaskResult struct {
	Success bool `json:"success"`
}
