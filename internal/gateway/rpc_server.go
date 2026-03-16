package gateway

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/lixianmin/got/loom"
	"github.com/lixianmin/logo"
	"github.com/lixianmin/pc/internal/engine"
	"github.com/lixianmin/pc/internal/plugin"
	"github.com/lixianmin/pc/pkg/protocol"
)

type RPCHandler func(ctx context.Context, params json.RawMessage) (interface{}, error)

type RPCServer struct {
	socketPath string
	listener   net.Listener
	handlers   map[string]RPCHandler
	engine     *engine.Engine
	pluginMgr  *plugin.PluginManager
}

func NewRPCServer(socketPath string, eng *engine.Engine, pluginMgr *plugin.PluginManager) *RPCServer {
	server := &RPCServer{
		socketPath: socketPath,
		handlers:   make(map[string]RPCHandler),
		engine:     eng,
		pluginMgr:  pluginMgr,
	}

	server.registerHandlers()

	return server
}

func (my *RPCServer) registerHandlers() {
	my.handlers[string(protocol.RPCMethodProcessMessage)] = my.handleProcessMessage
	my.handlers[string(protocol.RPCMethodProcessMessageStream)] = my.handleProcessMessageStream
	my.handlers[string(protocol.RPCMethodGetStatus)] = my.handleGetStatus
	my.handlers[string(protocol.RPCMethodListSkills)] = my.handleListSkills
	my.handlers[string(protocol.RPCMethodExecuteSkill)] = my.handleExecuteSkill
	my.handlers[string(protocol.RPCMethodListTasks)] = my.handleListTasks
	my.handlers[string(protocol.RPCMethodAddTask)] = my.handleAddTask
	my.handlers[string(protocol.RPCMethodCompleteTask)] = my.handleCompleteTask
	my.handlers[string(protocol.RPCMethodDeleteTask)] = my.handleDeleteTask
}

func (my *RPCServer) Start() error {
	if err := os.RemoveAll(my.socketPath); err != nil {
		return fmt.Errorf("failed to remove existing socket: %w", err)
	}

	listener, err := net.Listen("unix", my.socketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on socket: %w", err)
	}

	my.listener = listener

	if err := os.Chmod(my.socketPath, 0600); err != nil {
		listener.Close()
		return fmt.Errorf("failed to set socket permissions: %w", err)
	}

	logo.Info("RPC server listening on:", my.socketPath)

	loom.Go(func(later loom.Later) {
		my.acceptConnections(later)
	})

	return nil
}

func (my *RPCServer) Stop() error {
	if my.listener != nil {
		return my.listener.Close()
	}
	return nil
}

func (my *RPCServer) acceptConnections(later loom.Later) {
	for {
		conn, err := my.listener.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				logo.Warn("Temporary accept error:", err)
				continue
			}
			logo.Info("RPC server stopped accepting connections")
			return
		}

		loom.Go(func(later loom.Later) {
			my.handleConnection(conn)
		})
	}
}

func (my *RPCServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		lengthBuf := make([]byte, 4)
		totalRead := 0
		for totalRead < 4 {
			n, err := reader.Read(lengthBuf[totalRead:])
			if err != nil {
				return
			}
			totalRead += n
		}

		length := int32(lengthBuf[0])<<24 | int32(lengthBuf[1])<<16 | int32(lengthBuf[2])<<8 | int32(lengthBuf[3])
		if length <= 0 || length > 10*1024*1024 {
			logo.Error("Invalid request length:", length)
			return
		}

		reqData := make([]byte, length)
		totalRead = 0
		for totalRead < int(length) {
			n, err := reader.Read(reqData[totalRead:])
			if err != nil {
				return
			}
			totalRead += n
		}

		var req protocol.RPCRequest
		if err := json.Unmarshal(reqData, &req); err != nil {
			logo.Error("Failed to parse request:", err)
			my.sendError(conn, "", protocol.RPCErrorCodeParseError, "parse error")
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)

		resp := my.handleRequest(ctx, &req)
		cancel()

		respData, err := json.Marshal(resp)
		if err != nil {
			logo.Error("Failed to marshal response:", err)
			continue
		}

		respLength := int32(len(respData))
		lengthBytes := []byte{
			byte(respLength >> 24),
			byte(respLength >> 16),
			byte(respLength >> 8),
			byte(respLength),
		}
		if _, err := conn.Write(lengthBytes); err != nil {
			logo.Error("Failed to write length prefix:", err)
			return
		}

		if _, err := conn.Write(respData); err != nil {
			logo.Error("Failed to write response:", err)
			return
		}
	}
}

func (my *RPCServer) handleRequest(ctx context.Context, req *protocol.RPCRequest) *protocol.RPCResponse {
	if req.ID == "" {
		logo.Error("RPC request missing ID")
		return protocol.NewRPCErrorResponse("", protocol.RPCErrorCodeInvalidRequest, "missing request id")
	}

	if req.Method == "" {
		logo.Error("RPC request missing method, ID:", req.ID)
		return protocol.NewRPCErrorResponse(req.ID, protocol.RPCErrorCodeInvalidRequest, "missing method")
	}

	handler, ok := my.handlers[req.Method]
	if !ok {
		logo.Error("RPC method not found:", req.Method, "ID:", req.ID)
		return protocol.NewRPCErrorResponse(req.ID, protocol.RPCErrorCodeMethodNotFound, "method not found: "+req.Method)
	}

	result, err := handler(ctx, req.Params)
	if err != nil {
		logo.Error("RPC handler error for method:", req.Method, ", error:", err)
		return protocol.NewRPCErrorResponse(req.ID, protocol.RPCErrorCodeInternalError, err.Error())
	}

	return protocol.NewRPCResponse(req.ID, result)
}

func (my *RPCServer) sendError(conn net.Conn, id string, code int, message string) {
	resp := protocol.NewRPCErrorResponse(id, code, message)
	respData, _ := json.Marshal(resp)
	respData = append(respData, '\n')
	conn.Write(respData)
}

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

	if err := my.engine.CreateSession(req.SessionID); err != nil {
	}

	response, err := my.engine.ProcessMessage(ctx, req.SessionID, req.Message)
	if err != nil {
		return nil, err
	}

	return &protocol.ProcessMessageResult{Response: response}, nil
}

func (my *RPCServer) handleProcessMessageStream(ctx context.Context, params json.RawMessage) (interface{}, error) {
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

	if err := my.engine.CreateSession(req.SessionID); err != nil {
	}

	streamCh := my.engine.ProcessMessageStream(ctx, req.SessionID, req.Message)

	var chunks []protocol.ProcessMessageStreamChunk
	for chunk := range streamCh {
		chunks = append(chunks, protocol.ProcessMessageStreamChunk{
			Content: chunk.Content,
			Done:    chunk.Done,
			Error:   chunk.Error,
		})
	}

	if len(chunks) == 0 {
		return nil, fmt.Errorf("no streaming chunks returned")
	}

	return map[string]any{"chunks": chunks}, nil
}

func (my *RPCServer) handleGetStatus(ctx context.Context, params json.RawMessage) (interface{}, error) {
	plugins := my.pluginMgr.ListPlugins()

	return &protocol.GatewayStatus{
		Running:     true,
		Pid:         os.Getpid(),
		Version:     "dev",
		PluginCount: len(plugins),
	}, nil
}

func (my *RPCServer) handleListSkills(ctx context.Context, params json.RawMessage) (interface{}, error) {
	return &protocol.ListSkillsResult{
		Skills: []protocol.SkillInfo{},
	}, nil
}

func (my *RPCServer) handleExecuteSkill(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req protocol.ExecuteSkillParams
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	return &protocol.ExecuteSkillResult{
		Result: fmt.Sprintf("Skill %s executed (not implemented)", req.Name),
	}, nil
}

func (my *RPCServer) handleListTasks(ctx context.Context, params json.RawMessage) (interface{}, error) {
	return &protocol.ListTasksResult{
		Tasks: []protocol.TaskInfo{},
	}, nil
}

func (my *RPCServer) handleAddTask(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req protocol.AddTaskParams
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	return &protocol.AddTaskResult{
		ID: "task-001",
	}, nil
}

func (my *RPCServer) handleCompleteTask(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req protocol.CompleteTaskParams
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if req.ID == "" {
		return nil, fmt.Errorf("id is required")
	}

	return &protocol.CompleteTaskResult{
		Success: true,
	}, nil
}

func (my *RPCServer) handleDeleteTask(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req protocol.DeleteTaskParams
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if req.ID == "" {
		return nil, fmt.Errorf("id is required")
	}

	return &protocol.DeleteTaskResult{
		Success: true,
	}, nil
}
