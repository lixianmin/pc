package protocol

import (
	"encoding/json"
	"fmt"
)

// RPCMethod represents available RPC methods.
type RPCMethod string

const (
	// RPCMethodProcessMessage processes a user message.
	RPCMethodProcessMessage RPCMethod = "ProcessMessage"
	// RPCMethodGetStatus returns the gateway status.
	RPCMethodGetStatus RPCMethod = "GetStatus"
	// RPCMethodListSkills lists all available skills.
	RPCMethodListSkills RPCMethod = "ListSkills"
	// RPCMethodExecuteSkill executes a skill.
	RPCMethodExecuteSkill RPCMethod = "ExecuteSkill"
	// RPCMethodListTasks lists tasks.
	RPCMethodListTasks RPCMethod = "ListTasks"
	// RPCMethodAddTask adds a new task.
	RPCMethodAddTask RPCMethod = "AddTask"
	// RPCMethodCompleteTask completes a task.
	RPCMethodCompleteTask RPCMethod = "CompleteTask"
	// RPCMethodDeleteTask deletes a task.
	RPCMethodDeleteTask RPCMethod = "DeleteTask"
)

// RPCRequest represents an RPC request.
type RPCRequest struct {
	ID     string          `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

// RPCResponse represents an RPC response.
type RPCResponse struct {
	ID     string      `json:"id"`
	Result interface{} `json:"result,omitempty"`
	Error  *RPCError   `json:"error,omitempty"`
}

// RPCError represents an RPC error.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// UnmarshalJSON implements custom JSON unmarshaling to handle both string and int codes.
func (my *RPCError) UnmarshalJSON(data []byte) error {
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
			my.Code = RPCErrorCodeInternalError
		}
	default:
		my.Code = RPCErrorCodeInternalError
	}

	my.Message = raw.Message
	return nil
}

// Error implements the error interface.
func (my *RPCError) Error() string {
	return fmt.Sprintf("RPC error %d: %s", my.Code, my.Message)
}

// RPC error codes.
const (
	RPCErrorCodeParseError     = -32700
	RPCErrorCodeInvalidRequest = -32600
	RPCErrorCodeMethodNotFound = -32601
	RPCErrorCodeInvalidParams  = -32602
	RPCErrorCodeInternalError  = -32603
	RPCErrorCodeServerError    = -32000
)

// NewRPCError creates a new RPC error.
func NewRPCError(code int, message string) *RPCError {
	return &RPCError{
		Code:    code,
		Message: message,
	}
}

// NewRPCResponse creates a new successful RPC response.
func NewRPCResponse(id string, result interface{}) *RPCResponse {
	return &RPCResponse{
		ID:     id,
		Result: result,
	}
}

// NewRPCErrorResponse creates a new error RPC response.
func NewRPCErrorResponse(id string, code int, message string) *RPCResponse {
	return &RPCResponse{
		ID:    id,
		Error: NewRPCError(code, message),
	}
}

// IsSuccess returns true if the response is successful.
func (my *RPCResponse) IsSuccess() bool {
	return my.Error == nil
}

// IsError returns true if the response is an error.
func (my *RPCResponse) IsError() bool {
	return my.Error != nil
}

// RPCMessage represents a user message for ProcessMessage.
type ProcessMessageParams struct {
	SessionID string `json:"sessionId"`
	Message   string `json:"message"`
}

// ProcessMessageResult represents the result of ProcessMessage.
type ProcessMessageResult struct {
	Response string `json:"response"`
}

// GatewayStatus represents the gateway status.
type GatewayStatus struct {
	Running    bool   `json:"running"`
	Pid        int    `json:"pid,omitempty"`
	Version    string `json:"version,omitempty"`
	Uptime     int64  `json:"uptime,omitempty"` // seconds
	PluginCount int  `json:"pluginCount,omitempty"`
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
	Name   string                 `json:"name"`
	Params map[string]interface{} `json:"params,omitempty"`
}

// ExecuteSkillResult represents the result of ExecuteSkill.
type ExecuteSkillResult struct {
	Result interface{} `json:"result"`
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
