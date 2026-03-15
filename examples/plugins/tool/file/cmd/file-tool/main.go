package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/lixianmin/pc/examples/plugins/tool/file"
	"github.com/lixianmin/pc/pkg/protocol"
)

const version = "1.0.0"

func main() {
	config := file.Config{
		Timeout:          30 * time.Second,
		ConfirmDangerous: true,
		MaxFileSize:      10 * 1024 * 1024,
		AllowedPaths: []string{
			"/tmp",
			"~/Documents",
			"~/Desktop",
		},
		BlockedPaths: []string{
			"/etc/passwd",
			"/etc/shadow",
			"~/.ssh",
		},
	}
	tool := file.NewFileTool(config)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		req, err := protocol.DecodeRequest([]byte(line))
		if err != nil {
			resp := protocol.NewErrorResponse("", -32700, "Parse error")
			data, _ := resp.Encode()
			fmt.Println(string(data))
			continue
		}

		resp := handleRequest(tool, req)
		data, err := resp.Encode()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to encode response: %v\n", err)
			continue
		}
		fmt.Println(string(data))
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
		os.Exit(1)
	}
}

func handleRequest(tool *file.FileTool, req *protocol.Request) *protocol.Response {
	switch req.Method {
	case "initialize":
		return handleInitialize(req)
	case "call":
		return handleCall(tool, req)
	case "info":
		return handleInfo(req)
	default:
		return protocol.NewErrorResponse(req.Id, -32601, fmt.Sprintf("unknown method: %s", req.Method))
	}
}

func handleInitialize(req *protocol.Request) *protocol.Response {
	var initConfig struct {
		Timeout          string   `json:"timeout"`
		ConfirmDangerous bool     `json:"confirm_dangerous"`
		AllowedPaths     []string `json:"allowed_paths"`
		BlockedPaths     []string `json:"blocked_paths"`
	}

	if err := json.Unmarshal(req.Params, &initConfig); err == nil {
		if initConfig.Timeout != "" {
			if d, err := time.ParseDuration(initConfig.Timeout); err == nil {
				_ = d
			}
		}
	}

	return protocol.NewResponse(req.Id, map[string]any{
		"status":  "initialized",
		"version": version,
	})
}

func handleCall(tool *file.FileTool, req *protocol.Request) *protocol.Response {
	var params struct {
		Action  string `json:"action"`
		Path    string `json:"path"`
		Content string `json:"content"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		return protocol.NewErrorResponse(req.Id, -32602, fmt.Sprintf("invalid params: %v", err))
	}

	switch params.Action {
	case "read":
		return handleRead(tool, req, params.Path)
	case "write":
		return handleWrite(tool, req, params.Path, params.Content)
	case "delete":
		return handleDelete(tool, req, params.Path)
	case "list":
		return handleList(tool, req, params.Path)
	default:
		return protocol.NewErrorResponse(req.Id, -32602, fmt.Sprintf("unknown action: %s", params.Action))
	}
}

func handleRead(tool *file.FileTool, req *protocol.Request, path string) *protocol.Response {
	result, err := tool.Read(file.ReadParams{Path: path})
	if err != nil {
		return protocol.NewErrorResponse(req.Id, -32603, err.Error())
	}
	return protocol.NewResponse(req.Id, result)
}

func handleWrite(tool *file.FileTool, req *protocol.Request, path, content string) *protocol.Response {
	result, err := tool.Write(file.WriteParams{Path: path, Content: content})
	if err != nil {
		return protocol.NewErrorResponse(req.Id, -32603, err.Error())
	}
	return protocol.NewResponse(req.Id, result)
}

func handleDelete(tool *file.FileTool, req *protocol.Request, path string) *protocol.Response {
	result, err := tool.Delete(file.DeleteParams{Path: path})
	if err != nil {
		return protocol.NewErrorResponse(req.Id, -32603, err.Error())
	}
	return protocol.NewResponse(req.Id, result)
}

func handleList(tool *file.FileTool, req *protocol.Request, path string) *protocol.Response {
	if path == "" {
		path = "."
	}
	result, err := tool.List(file.ListParams{Path: path})
	if err != nil {
		return protocol.NewErrorResponse(req.Id, -32603, err.Error())
	}
	return protocol.NewResponse(req.Id, result)
}

func handleInfo(req *protocol.Request) *protocol.Response {
	return protocol.NewResponse(req.Id, map[string]any{
		"name":        "file",
		"type":        "tool",
		"version":     version,
		"description": "Read, write, delete, and list files",
		"actions": []map[string]string{
			{
				"name":        "read",
				"description": "Read file content",
				"params":      "path: file path to read",
			},
			{
				"name":        "write",
				"description": "Write content to file",
				"params":      "path: file path, content: content to write",
			},
			{
				"name":        "delete",
				"description": "Delete a file",
				"params":      "path: file path to delete",
			},
			{
				"name":        "list",
				"description": "List directory contents",
				"params":      "path: directory path (optional, default: current directory)",
			},
		},
	})
}
