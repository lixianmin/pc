package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/lixianmin/pc/examples/plugins/llm/openai"
	"github.com/lixianmin/pc/pkg/protocol"
)

// Alias for compatibility
type LLMClient = openai.GLMClient

func main() {
	client := createClient()

	// Read commands from stdin
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		req, err := protocol.DecodeRequest([]byte(line))
		if err != nil {
			sendError(req.ID, -32700, "Parse error", err)
			continue
		}

		handleRequest(req, client)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
		os.Exit(1)
	}
}

func createClient() *openai.GLMClient {
	// Load configuration from config.yml
	// For now, use environment variables or defaults
	apiKey := os.Getenv("GLM_API_KEY")
	if apiKey == "" {
		apiKey = "your-glm-api-key-here"
	}

	model := os.Getenv("GLM_MODEL")
	if model == "" {
		model = "glm-4.7"
	}

	baseURL := os.Getenv("GLM_API_BASE")
	if baseURL == "" {
		baseURL = "https://open.bigmodel.cn/api/coding/paas/v4"
	}

	return openai.NewGLMClient(apiKey, model, baseURL)
}

func handleRequest(req *protocol.Request, client *openai.GLMClient) {
	var result any
	var err error

	switch req.Method {
	case "complete":
		result, err = handleComplete(req.Params, client)
	case "stream":
		result, err = handleStream(req.Params, client)
	case "models":
		result, err = handleModels(client)
	default:
		sendError(req.ID, -32601, "Method not found", fmt.Errorf("unknown method: %s", req.Method))
		return
	}

	if err != nil {
		sendError(req.ID, -32603, "Internal error", err)
		return
	}

	sendResponse(req.ID, result)
}

func handleComplete(params any, client *openai.GLMClient) (any, error) {
	// Parse params
	var p struct {
		Prompt string                 `json:"prompt"`
		Options map[string]any        `json:"options"`
	}

	if err := parseParams(params, &p); err != nil {
		return nil, err
	}

	result, err := client.Complete(p.Prompt, p.Options)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"text": result,
	}, nil
}

func handleStream(params any, client *openai.GLMClient) (any, error) {
	// Parse params
	var p struct {
		Prompt string                 `json:"prompt"`
		Options map[string]any        `json:"options"`
	}

	if err := parseParams(params, &p); err != nil {
		return nil, err
	}

	ch, err := client.Stream(p.Prompt, p.Options)
	if err != nil {
		return nil, err
	}

	// Collect stream chunks for response
	texts := []string{}
	for text := range ch {
		texts = append(texts, text)
	}

	return map[string]any{
		"texts": texts,
	}, nil
}

func handleModels(client *openai.GLMClient) (any, error) {
	models, err := client.Models()
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"models": models,
	}, nil
}

func parseParams(params any, v any) error {
	data, err := json.Marshal(params)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func sendResponse(id string, result any) {
	resp := protocol.NewResponse(id, result)
	data, _ := resp.Encode()
	fmt.Println(string(data))
}

func sendError(id string, code int, message string, err error) {
	data := map[string]any{
		"error": err.Error(),
	}
	resp := protocol.NewErrorWithData(id, code, message, data)
	respData, _ := resp.Encode()
	fmt.Println(string(respData))
}
