package gateway

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/lixianmin/logo"
	"github.com/lixianmin/pc/internal/engine"
	"github.com/lixianmin/pc/internal/plugin"
	"github.com/lixianmin/pc/pkg/protocol"
)

// RPCHandler is a function that handles RPC requests.
type RPCHandler func(ctx context.Context, params json.RawMessage) (interface{}, error)

// RPCServer represents the RPC server.
type RPCServer struct {
	socketPath string
	listener   net.Listener
	handlers   map[string]RPCHandler
	engine     *engine.Engine
	pluginMgr  *plugin.PluginManager
}

// NewRPCServer creates a new RPC server.
func NewRPCServer(socketPath string, eng *engine.Engine, pluginMgr *plugin.PluginManager) *RPCServer {
	server := &RPCServer{
		socketPath: socketPath,
		handlers:   make(map[string]RPCHandler),
		engine:     eng,
		pluginMgr:  pluginMgr,
	}

	// Register handlers
	server.registerHandlers()

	return server
}

// registerHandlers registers RPC method handlers.
func (my *RPCServer) registerHandlers() {
	my.handlers[string(protocol.RPCMethodProcessMessage)] = my.handleProcessMessage
	my.handlers[string(protocol.RPCMethodGetStatus)] = my.handleGetStatus
	my.handlers[string(protocol.RPCMethodListSkills)] = my.handleListSkills
	my.handlers[string(protocol.RPCMethodExecuteSkill)] = my.handleExecuteSkill
	my.handlers[string(protocol.RPCMethodListTasks)] = my.handleListTasks
	my.handlers[string(protocol.RPCMethodAddTask)] = my.handleAddTask
	my.handlers[string(protocol.RPCMethodCompleteTask)] = my.handleCompleteTask
}

// Start starts the RPC server.
func (my *RPCServer) Start() error {
	// Remove existing socket file
	if err := os.RemoveAll(my.socketPath); err != nil {
		return fmt.Errorf("failed to remove existing socket: %w", err)
	}

	// Create listener
	listener, err := net.Listen("unix", my.socketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on socket: %w", err)
	}

	my.listener = listener

	// Set socket permissions (only owner can access)
	if err := os.Chmod(my.socketPath, 0600); err != nil {
		listener.Close()
		return fmt.Errorf("failed to set socket permissions: %w", err)
	}

	logo.Info("RPC server listening on:", my.socketPath)

	// Accept connections
	go my.acceptConnections()

	return nil
}

// Stop stops the RPC server.
func (my *RPCServer) Stop() error {
	if my.listener != nil {
		return my.listener.Close()
	}
	return nil
}

// acceptConnections accepts incoming connections.
func (my *RPCServer) acceptConnections() {
	for {
		conn, err := my.listener.Accept()
		if err != nil {
			// Check if listener was closed
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				logo.Warn("Temporary accept error:", err)
				continue
			}
			// Listener closed
			logo.Info("RPC server stopped accepting connections")
			return
		}

		go my.handleConnection(conn)
	}
}

// handleConnection handles a client connection.
func (my *RPCServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		// Read request (line-delimited)
		reqData, err := reader.ReadBytes('\n')
		if err != nil {
			// Connection closed or error
			return
		}

		// Parse request
		var req protocol.RPCRequest
		if err := json.Unmarshal(reqData, &req); err != nil {
			logo.Error("Failed to parse request:", err)
			my.sendError(conn, "", protocol.RPCErrorCodeParseError, "parse error")
			continue
		}

		// Create timeout context for request handling
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		// Handle request
		resp := my.handleRequest(ctx, &req)
		cancel()

		// Send response
		respData, err := json.Marshal(resp)
		if err != nil {
			logo.Error("Failed to marshal response:", err)
			continue
		}

		respData = append(respData, '\n')
		if _, err := conn.Write(respData); err != nil {
			logo.Error("Failed to write response:", err)
			return
		}
	}
}

// handleRequest handles an RPC request.
func (my *RPCServer) handleRequest(ctx context.Context, req *protocol.RPCRequest) *protocol.RPCResponse {
	// Validate request
	if req.ID == "" {
		logo.Error("RPC request missing ID")
		return protocol.NewRPCErrorResponse("", protocol.RPCErrorCodeInvalidRequest, "missing request id")
	}

	if req.Method == "" {
		logo.Error("RPC request missing method, ID:", req.ID)
		return protocol.NewRPCErrorResponse(req.ID, protocol.RPCErrorCodeInvalidRequest, "missing method")
	}

	// Find handler
	handler, ok := my.handlers[req.Method]
	if !ok {
		logo.Error("RPC method not found:", req.Method, "ID:", req.ID)
		return protocol.NewRPCErrorResponse(req.ID, protocol.RPCErrorCodeMethodNotFound, "method not found: "+req.Method)
	}

	// Call handler
	result, err := handler(ctx, req.Params)
	if err != nil {
		logo.Error("RPC handler error for method:", req.Method, ", error:", err)
		return protocol.NewRPCErrorResponse(req.ID, protocol.RPCErrorCodeInternalError, err.Error())
	}

	return protocol.NewRPCResponse(req.ID, result)
}

// sendError sends an error response.
func (my *RPCServer) sendError(conn net.Conn, id string, code int, message string) {
	resp := protocol.NewRPCErrorResponse(id, code, message)
	respData, _ := json.Marshal(resp)
	respData = append(respData, '\n')
	conn.Write(respData)
}

// handleProcessMessage handles ProcessMessage requests.
func (my *RPCServer) handleProcessMessage(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req protocol.ProcessMessageParams
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if req.SessionID == "" {
		return nil, fmt.Errorf("sessionId is required")
	}

	if req.Message == "" {
		return nil, fmt.Errorf("message is required")
	}

	// Create session if not exists
	if err := my.engine.CreateSession(req.SessionID); err != nil {
		// Session might already exist, ignore error
	}

	// Process message
	response, err := my.engine.ProcessMessage(ctx, req.SessionID, req.Message)
	if err != nil {
		return nil, err
	}

	return &protocol.ProcessMessageResult{Response: response}, nil
}

// handleGetStatus handles GetStatus requests.
func (my *RPCServer) handleGetStatus(ctx context.Context, params json.RawMessage) (interface{}, error) {
	plugins := my.pluginMgr.ListPlugins()

	return &protocol.GatewayStatus{
		Running:     true,
		Pid:         os.Getpid(),
		Version:     "dev", // TODO: Get actual version
		PluginCount: len(plugins),
	}, nil
}

// handleListSkills handles ListSkills requests.
func (my *RPCServer) handleListSkills(ctx context.Context, params json.RawMessage) (interface{}, error) {
	// TODO: Implement skill listing
	// For now, return empty list
	return &protocol.ListSkillsResult{
		Skills: []protocol.SkillInfo{},
	}, nil
}

// handleExecuteSkill handles ExecuteSkill requests.
func (my *RPCServer) handleExecuteSkill(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req protocol.ExecuteSkillParams
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	// TODO: Implement skill execution
	return &protocol.ExecuteSkillResult{
		Result: fmt.Sprintf("Skill %s executed (not implemented)", req.Name),
	}, nil
}

// handleListTasks handles ListTasks requests.
func (my *RPCServer) handleListTasks(ctx context.Context, params json.RawMessage) (interface{}, error) {
	// TODO: Implement task listing
	return &protocol.ListTasksResult{
		Tasks: []protocol.TaskInfo{},
	}, nil
}

// handleAddTask handles AddTask requests.
func (my *RPCServer) handleAddTask(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req protocol.AddTaskParams
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	// TODO: Implement task addition
	return &protocol.AddTaskResult{
		ID: "task-001", // Placeholder
	}, nil
}

// handleCompleteTask handles CompleteTask requests.
func (my *RPCServer) handleCompleteTask(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req protocol.CompleteTaskParams
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if req.ID == "" {
		return nil, fmt.Errorf("id is required")
	}

	// TODO: Implement task completion
	return &protocol.CompleteTaskResult{
		Success: true,
	}, nil
}
