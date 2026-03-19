package protocol

import (
	"encoding/json"
	"testing"
)

func TestNewRPCError(t *testing.T) {
	err := NewRpcError(-32700, "parse error")

	if err.Code != -32700 {
		t.Errorf("Code = %d, want -32700", err.Code)
	}

	if err.Message != "parse error" {
		t.Errorf("Message = %q, want 'parse error'", err.Message)
	}

	expected := "RPC error -32700: parse error"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}
}

func TestNewRPCResponse(t *testing.T) {
	result := map[string]string{"key": "value"}
	resp := NewRpcResponse("req-123", result)

	if resp.ID != "req-123" {
		t.Errorf("ID = %q, want 'req-123'", resp.ID)
	}

	if resp.Error != nil {
		t.Error("Error should be nil for successful response")
	}

	if resp.Result == nil {
		t.Error("Result should not be nil")
	}
}

func TestNewRPCErrorResponse(t *testing.T) {
	resp := NewRpcErrorResponse("req-123", -32601, "method not found")

	if resp.ID != "req-123" {
		t.Errorf("ID = %q, want 'req-123'", resp.ID)
	}

	if resp.Result != nil {
		t.Error("Result should be nil for error response")
	}

	if resp.Error == nil {
		t.Fatal("Error should not be nil")
	}

	if resp.Error.Code != -32601 {
		t.Errorf("Error.Code = %d, want -32601", resp.Error.Code)
	}

	if resp.Error.Message != "method not found" {
		t.Errorf("Error.Message = %q, want 'method not found'", resp.Error.Message)
	}
}

func TestRPCResponse_IsSuccess(t *testing.T) {
	tests := []struct {
		name     string
		resp     *RpcResponse
		expected bool
	}{
		{
			name:     "success response",
			resp:     NewRpcResponse("1", "result"),
			expected: true,
		},
		{
			name:     "error response",
			resp:     NewRpcErrorResponse("1", -1, "error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.resp.IsSuccess(); got != tt.expected {
				t.Errorf("IsSuccess() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRPCResponse_IsError(t *testing.T) {
	tests := []struct {
		name     string
		resp     *RpcResponse
		expected bool
	}{
		{
			name:     "success response",
			resp:     NewRpcResponse("1", "result"),
			expected: false,
		},
		{
			name:     "error response",
			resp:     NewRpcErrorResponse("1", -1, "error"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.resp.IsError(); got != tt.expected {
				t.Errorf("IsError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRPCRequest_Marshal(t *testing.T) {
	params := json.RawMessage(`{"key":"value"}`)
	req := &RpcRequest{
		ID:     "req-1",
		Method: "TestMethod",
		Params: params,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded RpcRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.ID != req.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, req.ID)
	}

	if decoded.Method != req.Method {
		t.Errorf("Method = %q, want %q", decoded.Method, req.Method)
	}
}

func TestProcessMessageParams(t *testing.T) {
	params := ProcessMessageParams{
		SessionID: "session-123",
		Message:   "Hello",
	}

	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded ProcessMessageParams
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.SessionID != params.SessionID {
		t.Errorf("SessionID = %q, want %q", decoded.SessionID, params.SessionID)
	}

	if decoded.Message != params.Message {
		t.Errorf("Message = %q, want %q", decoded.Message, params.Message)
	}
}

func TestProcessMessageResult(t *testing.T) {
	result := ProcessMessageResult{
		Response: "Hello there!",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded ProcessMessageResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.Response != result.Response {
		t.Errorf("Response = %q, want %q", decoded.Response, result.Response)
	}
}

func TestGatewayStatus(t *testing.T) {
	status := GatewayStatus{
		Running:     true,
		Pid:         12345,
		Version:     "1.0.0",
		Uptime:      3600,
		PluginCount: 5,
	}

	data, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded GatewayStatus
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.Running != status.Running {
		t.Errorf("Running = %v, want %v", decoded.Running, status.Running)
	}

	if decoded.Pid != status.Pid {
		t.Errorf("Pid = %d, want %d", decoded.Pid, status.Pid)
	}
}

func TestListSkillsResult(t *testing.T) {
	result := ListSkillsResult{
		Skills: []SkillInfo{
			{Name: "skill1", Description: "desc1"},
			{Name: "skill2", Description: "desc2"},
		},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded ListSkillsResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if len(decoded.Skills) != 2 {
		t.Errorf("len(Skills) = %d, want 2", len(decoded.Skills))
	}
}

func TestExecuteSkillParams(t *testing.T) {
	params := ExecuteSkillParams{
		Name:   "test-skill",
		Params: map[string]interface{}{"key": "value"},
	}

	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded ExecuteSkillParams
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.Name != params.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, params.Name)
	}
}

func TestTaskInfo(t *testing.T) {
	task := TaskInfo{
		ID:          "task-1",
		Title:       "Test Task",
		Status:      "pending",
		Description: "A test task",
	}

	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded TaskInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.ID != task.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, task.ID)
	}
}

func TestListTasksResult(t *testing.T) {
	result := ListTasksResult{
		Tasks: []TaskInfo{
			{ID: "1", Title: "Task 1", Status: "pending"},
			{ID: "2", Title: "Task 2", Status: "completed"},
		},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded ListTasksResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if len(decoded.Tasks) != 2 {
		t.Errorf("len(Tasks) = %d, want 2", len(decoded.Tasks))
	}
}

func TestAddTaskParams(t *testing.T) {
	params := AddTaskParams{
		Title:       "New Task",
		Description: "Task description",
	}

	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded AddTaskParams
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.Title != params.Title {
		t.Errorf("Title = %q, want %q", decoded.Title, params.Title)
	}
}

func TestAddTaskResult(t *testing.T) {
	result := AddTaskResult{
		ID: "task-123",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded AddTaskResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.ID != result.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, result.ID)
	}
}

func TestCompleteTaskParams(t *testing.T) {
	params := CompleteTaskParams{
		ID: "task-123",
	}

	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded CompleteTaskParams
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.ID != params.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, params.ID)
	}
}

func TestCompleteTaskResult(t *testing.T) {
	result := CompleteTaskResult{
		Success: true,
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded CompleteTaskResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.Success != result.Success {
		t.Errorf("Success = %v, want %v", decoded.Success, result.Success)
	}
}

func TestRPCError_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		wantCode int
		wantMsg  string
	}{
		{
			name:     "integer code",
			json:     `{"code":-32603,"message":"internal error"}`,
			wantCode: -32603,
			wantMsg:  "internal error",
		},
		{
			name:     "string code",
			json:     `{"code":"-32603","message":"internal error"}`,
			wantCode: -32603,
			wantMsg:  "internal error",
		},
		{
			name:     "float code",
			json:     `{"code":-32603.0,"message":"internal error"}`,
			wantCode: -32603,
			wantMsg:  "internal error",
		},
		{
			name:     "zero code",
			json:     `{"code":0,"message":"no error"}`,
			wantCode: 0,
			wantMsg:  "no error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err RpcError
			if unmarshalErr := json.Unmarshal([]byte(tt.json), &err); unmarshalErr != nil {
				t.Fatalf("Unmarshal error: %v", unmarshalErr)
			}

			if err.Code != tt.wantCode {
				t.Errorf("Code = %d, want %d", err.Code, tt.wantCode)
			}
			if err.Message != tt.wantMsg {
				t.Errorf("Message = %q, want %q", err.Message, tt.wantMsg)
			}
		})
	}
}

func TestRPCError_UnmarshalJSON_InvalidCode(t *testing.T) {
	// Test with invalid code type - should default to internal error
	jsonData := `{"code":{},"message":"unknown error"}`
	var err RpcError
	if unmarshalErr := json.Unmarshal([]byte(jsonData), &err); unmarshalErr != nil {
		t.Fatalf("Unmarshal error: %v", unmarshalErr)
	}

	if err.Code != RpcErrorCodeInternalError {
		t.Errorf("Code = %d, want RPCErrorCodeInternalError (%d)", err.Code, RpcErrorCodeInternalError)
	}
}

func TestRPCErrorResponse_WithStringCode(t *testing.T) {
	// Test full response unmarshal with string code
	jsonData := `{"id":"123","error":{"code":"-32600","message":"invalid request"}}`
	var resp RpcResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if resp.ID != "123" {
		t.Errorf("ID = %q, want %q", resp.ID, "123")
	}
	if resp.Error == nil {
		t.Fatal("Error is nil")
	}
	if resp.Error.Code != -32600 {
		t.Errorf("Error.Code = %d, want -32600", resp.Error.Code)
	}
	if resp.Error.Message != "invalid request" {
		t.Errorf("Error.Message = %q, want %q", resp.Error.Message, "invalid request")
	}
}

func TestRPCErrorResponse_WithIntCode(t *testing.T) {
	// Test full response unmarshal with int code
	jsonData := `{"id":"456","error":{"code":-32603,"message":"internal error"}}`
	var resp RpcResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if resp.ID != "456" {
		t.Errorf("ID = %q, want %q", resp.ID, "456")
	}
	if resp.Error == nil {
		t.Fatal("Error is nil")
	}
	if resp.Error.Code != -32603 {
		t.Errorf("Error.Code = %d, want -32603", resp.Error.Code)
	}
	if resp.Error.Message != "internal error" {
		t.Errorf("Error.Message = %q, want %q", resp.Error.Message, "internal error")
	}
}
