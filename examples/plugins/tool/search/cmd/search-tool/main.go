package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/lixianmin/pc/examples/plugins/tool/search"
	"github.com/lixianmin/pc/pkg/protocol"
)

const version = "1.0.0"

func main() {
	config := search.Config{
		MaxResults: 5,
	}
	tool := search.NewSearchTool(config)

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

func handleRequest(tool *search.SearchTool, req *protocol.Request) *protocol.Response {
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

func handleCall(tool *search.SearchTool, req *protocol.Request) *protocol.Response {
	var params struct {
		Query string `json:"query"`
		Limit int    `json:"limit"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		return protocol.NewErrorResponse(req.Id, -32602, fmt.Sprintf("invalid params: %v", err))
	}

	results, err := tool.Search(search.SearchParams{
		Query: params.Query,
		Limit: params.Limit,
	})
	if err != nil {
		return protocol.NewErrorResponse(req.Id, -32603, err.Error())
	}

	return protocol.NewResponse(req.Id, map[string]any{
		"query":   params.Query,
		"count":   len(results),
		"results": results,
	})
}

func handleInfo(req *protocol.Request) *protocol.Response {
	return protocol.NewResponse(req.Id, map[string]any{
		"name":        "search",
		"type":        "tool",
		"version":     version,
		"description": "Search the web using DuckDuckGo",
		"params": map[string]string{
			"query": "Search query string (required)",
			"limit": "Maximum number of results (optional, default: 5)",
		},
	})
}
