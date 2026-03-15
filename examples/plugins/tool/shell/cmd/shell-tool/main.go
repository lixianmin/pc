package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/lixianmin/pc/examples/plugins/tool/shell"
	"github.com/lixianmin/pc/pkg/protocol"
)

const version = "1.0.0"

func main() {
	config := shell.Config{
		Timeout:          30 * time.Second,
		ConfirmDangerous: true,
		BlockedCommands: []string{
			"rm -rf /",
			"mkfs",
			"dd if=",
		},
	}
	tool := shell.NewShellTool(config)

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

		resp := handleRequest(tool, req, &config)
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

func handleRequest(tool *shell.ShellTool, req *protocol.Request, config *shell.Config) *protocol.Response {
	switch req.Method {
	case "initialize":
		return handleInitialize(req, config, tool)
	case "call":
		return handleCall(tool, req)
	case "info":
		return handleInfo(req)
	default:
		return protocol.NewErrorResponse(req.Id, -32601, fmt.Sprintf("unknown method: %s", req.Method))
	}
}

func handleInitialize(req *protocol.Request, config *shell.Config, tool *shell.ShellTool) *protocol.Response {
	var initConfig struct {
		Timeout          string   `json:"timeout"`
		ConfirmDangerous bool     `json:"confirm_dangerous"`
		AllowedCommands  []string `json:"allowed_commands"`
		BlockedCommands  []string `json:"blocked_commands"`
	}

	if err := json.Unmarshal(req.Params, &initConfig); err == nil {
		if initConfig.Timeout != "" {
			if d, err := time.ParseDuration(initConfig.Timeout); err == nil {
				config.Timeout = d
			}
		}
		config.ConfirmDangerous = initConfig.ConfirmDangerous
		if len(initConfig.AllowedCommands) > 0 {
			config.AllowedCommands = initConfig.AllowedCommands
		}
		if len(initConfig.BlockedCommands) > 0 {
			config.BlockedCommands = initConfig.BlockedCommands
		}
		tool.SetConfig(*config)
	}

	return protocol.NewResponse(req.Id, map[string]any{
		"status":  "initialized",
		"version": version,
	})
}

func handleCall(tool *shell.ShellTool, req *protocol.Request) *protocol.Response {
	var params shell.CallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return protocol.NewErrorResponse(req.Id, -32602, fmt.Sprintf("invalid params: %v", err))
	}

	if tool.IsDangerous(params.Command) {
		return protocol.NewErrorResponse(req.Id, -32603, "dangerous command detected")
	}

	result, err := tool.Execute(params)
	if err != nil {
		return protocol.NewErrorResponse(req.Id, -32603, err.Error())
	}

	return protocol.NewResponse(req.Id, result)
}

func handleInfo(req *protocol.Request) *protocol.Response {
	return protocol.NewResponse(req.Id, map[string]any{
		"name":        "shell",
		"type":        "tool",
		"version":     version,
		"description": "Execute shell commands",
		"params_schema": map[string]any{
			"command":     "The shell command to execute",
			"timeout":     "Optional timeout in seconds (default: 30)",
			"working_dir": "Optional working directory",
		},
	})
}
