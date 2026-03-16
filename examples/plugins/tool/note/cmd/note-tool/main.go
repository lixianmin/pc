package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/lixianmin/pc/examples/plugins/tool/note"
	"github.com/lixianmin/pc/pkg/protocol"
)

const version = "1.0.0"

func main() {
	config := note.Config{}
	tool := note.NewNoteTool(config)

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

func handleRequest(tool *note.NoteTool, req *protocol.Request) *protocol.Response {
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

func handleCall(tool *note.NoteTool, req *protocol.Request) *protocol.Response {
	var params struct {
		Action  string `json:"action"`
		ID      string `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
		Tags    string `json:"tags"`
		Query   string `json:"query"`
		Tag     string `json:"tag"`
		Limit   int    `json:"limit"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		return protocol.NewErrorResponse(req.Id, -32602, fmt.Sprintf("invalid params: %v", err))
	}

	switch params.Action {
	case "add":
		result, err := tool.Add(note.AddParams{
			Title:   params.Title,
			Content: params.Content,
			Tags:    params.Tags,
		})
		if err != nil {
			return protocol.NewErrorResponse(req.Id, -32603, err.Error())
		}
		return protocol.NewResponse(req.Id, result)

	case "get":
		result, err := tool.Get(note.GetParams{
			ID: params.ID,
		})
		if err != nil {
			return protocol.NewErrorResponse(req.Id, -32603, err.Error())
		}
		return protocol.NewResponse(req.Id, result)

	case "update":
		result, err := tool.Update(note.UpdateParams{
			ID:      params.ID,
			Title:   params.Title,
			Content: params.Content,
			Tags:    params.Tags,
		})
		if err != nil {
			return protocol.NewErrorResponse(req.Id, -32603, err.Error())
		}
		return protocol.NewResponse(req.Id, result)

	case "delete":
		err := tool.Delete(note.DeleteParams{
			ID: params.ID,
		})
		if err != nil {
			return protocol.NewErrorResponse(req.Id, -32603, err.Error())
		}
		return protocol.NewResponse(req.Id, map[string]any{"status": "deleted"})

	case "list":
		result, err := tool.List(note.ListParams{
			Tag:   params.Tag,
			Limit: params.Limit,
		})
		if err != nil {
			return protocol.NewErrorResponse(req.Id, -32603, err.Error())
		}
		return protocol.NewResponse(req.Id, result)

	case "search":
		result, err := tool.Search(note.SearchParams{
			Query: params.Query,
			Limit: params.Limit,
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
		"name":        "note",
		"type":        "tool",
		"version":     version,
		"description": "Manage persistent notes",
		"actions": []map[string]string{
			{
				"name":        "add",
				"description": "Create a new note",
				"params":      "title (required), content, tags",
			},
			{
				"name":        "get",
				"description": "Get a note by ID",
				"params":      "id (required)",
			},
			{
				"name":        "update",
				"description": "Update an existing note",
				"params":      "id (required), title, content, tags",
			},
			{
				"name":        "delete",
				"description": "Delete a note by ID",
				"params":      "id (required)",
			},
			{
				"name":        "list",
				"description": "List all notes",
				"params":      "tag (optional filter), limit (default: 50)",
			},
			{
				"name":        "search",
				"description": "Search notes by query",
				"params":      "query (required), limit (default: 20)",
			},
		},
	})
}
