package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/lixianmin/pc/examples/plugins/tool/time"
	"github.com/lixianmin/pc/pkg/protocol"
)

const version = "1.0.0"

func main() {
	tool := time.NewTimeTool()

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
}

func handleRequest(tool *time.TimeTool, req *protocol.Request) *protocol.Response {
	switch req.Method {
	case "initialize":
		return protocol.NewResponse(req.Id, map[string]any{
			"status":  "initialized",
			"version": version,
		})
	case "call":
		return handleCall(tool, req)
	case "info":
		return protocol.NewResponse(req.Id, map[string]any{
			"name":    "time",
			"type":    "tool",
			"version": version,
			"description": "Time and date operations",
			"actions": []map[string]string{
				{"name": "now", "description": "Get current time"},
				{"name": "format", "description": "Format unix timestamp"},
				{"name": "parse", "description": "Parse time string"},
			},
		})
	default:
		return protocol.NewErrorResponse(req.Id, -32601, fmt.Sprintf("unknown method: %s", req.Method))
	}
}

func handleCall(tool *time.TimeTool, req *protocol.Request) *protocol.Response {
	var params struct {
		Action string `json:"action"`
		Unix   int64  `json:"unix"`
		Layout string `json:"layout"`
		Value  string `json:"value"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return protocol.NewErrorResponse(req.Id, -32602, fmt.Sprintf("invalid params: %v", err))
	}

	switch params.Action {
	case "now":
		return protocol.NewResponse(req.Id, tool.Now())
	case "format":
		result, err := tool.Format(params.Unix, params.Layout)
		if err != nil {
			return protocol.NewErrorResponse(req.Id, -32603, err.Error())
		}
		return protocol.NewResponse(req.Id, map[string]any{"result": result})
	case "parse":
		result, err := tool.Parse(params.Value, params.Layout)
		if err != nil {
			return protocol.NewErrorResponse(req.Id, -32603, err.Error())
		}
		return protocol.NewResponse(req.Id, result)
	default:
		return protocol.NewErrorResponse(req.Id, -32602, fmt.Sprintf("unknown action: %s", params.Action))
	}
}
