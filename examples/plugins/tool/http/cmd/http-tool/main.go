package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/lixianmin/pc/examples/plugins/tool/http"
	"github.com/lixianmin/pc/pkg/protocol"
)

const version = "1.0.0"

func main() {
	config := http.Config{
		Timeout:     30 * time.Second,
		MaxBodySize: 1024 * 1024,
	}
	tool := http.NewHTTPTool(config)

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

func handleRequest(tool *http.HTTPTool, req *protocol.Request) *protocol.Response {
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
	return protocol.NewResponse(req.Id, map[string]any{
		"status":  "initialized",
		"version": version,
	})
}

func handleCall(tool *http.HTTPTool, req *protocol.Request) *protocol.Response {
	var params struct {
		Action  string            `json:"action"`
		URL     string            `json:"url"`
		Method  string            `json:"method"`
		Headers map[string]string `json:"headers"`
		Body    string            `json:"body"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		return protocol.NewErrorResponse(req.Id, -32602, fmt.Sprintf("invalid params: %v", err))
	}

	switch params.Action {
	case "get":
		result, err := tool.Get(http.GetParams{
			URL:     params.URL,
			Headers: params.Headers,
		})
		if err != nil {
			return protocol.NewErrorResponse(req.Id, -32603, err.Error())
		}
		return protocol.NewResponse(req.Id, result)
	case "post":
		result, err := tool.Post(http.PostParams{
			URL:     params.URL,
			Headers: params.Headers,
			Body:    params.Body,
		})
		if err != nil {
			return protocol.NewErrorResponse(req.Id, -32603, err.Error())
		}
		return protocol.NewResponse(req.Id, result)
	case "request":
		result, err := tool.Request(http.RequestParams{
			URL:     params.URL,
			Method:  params.Method,
			Headers: params.Headers,
			Body:    params.Body,
		})
		if err != nil {
			return protocol.NewErrorResponse(req.Id, -32603, err.Error())
		}
		return protocol.NewResponse(req.Id, result)
	default:
		return protocol.NewErrorResponse(req.Id, -32602, fmt.Sprintf("unknown action: %s", params.Action))
	}
}

func handleInfo(req *protocol.Request) *protocol.Response {
	return protocol.NewResponse(req.Id, map[string]any{
		"name":    "http",
		"type":    "tool",
		"version": version,
		"description": "Make HTTP requests",
		"actions": []map[string]string{
			{
				"name":        "get",
				"description": "Make HTTP GET request",
				"params":      "url: request URL, headers: optional headers",
			},
			{
				"name":        "post",
				"description": "Make HTTP POST request",
				"params":      "url: request URL, body: request body, headers: optional headers",
			},
			{
				"name":        "request",
				"description": "Make HTTP request with custom method",
				"params":      "url: request URL, method: HTTP method, body: optional body, headers: optional headers",
			},
		},
	})
}
