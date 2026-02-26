// OpenAI LLM Plugin for PersonalClaw
// This plugin implements the LLM Provider interface using OpenAI API
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	version = "1.0.0"
)

// Request represents a stdio protocol request
type Request struct {
	Version string          `json:"version"`
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

// Response represents a stdio protocol response
type Response struct {
	Version string          `json:"version"`
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Result  any             `json:"result,omitempty"`
	Error   *ErrorDetail    `json:"error,omitempty"`
}

// ErrorDetail represents error information
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Config holds the plugin configuration
type Config struct {
	APIKey string `json:"api_key"`
	Model  string `json:"model"`
	BaseURL string `json:"base_url,omitempty"`
}

// OpenAIRequest represents OpenAI API request
type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream,omitempty"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse represents OpenAI API response
type OpenAIResponse struct {
	ID      string   `json:"id"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice represents a completion choice
type Choice struct {
	Index   int     `json:"index"`
	Message Message `json:"message"`
}

// Usage represents token usage
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ModelInfo represents model information
type ModelInfo struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// Plugin is the OpenAI LLM plugin
type Plugin struct {
	config Config
	client *http.Client
}

// NewPlugin creates a new OpenAI plugin
func NewPlugin() *Plugin {
	return &Plugin{
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Initialize sets up the plugin with configuration
func (p *Plugin) Initialize(config Config) error {
	p.config = config
	if p.config.APIKey == "" {
		return fmt.Errorf("api_key is required")
	}
	if p.config.Model == "" {
		p.config.Model = "gpt-4o-mini"
	}
	if p.config.BaseURL == "" {
		p.config.BaseURL = "https://api.openai.com/v1"
	}
	return nil
}

// Handle processes a request and returns a response
func (p *Plugin) Handle(req Request) Response {
	resp := Response{
		Version: version,
		ID:      req.ID,
		Type:    "response",
	}

	switch req.Method {
	case "complete":
		result, err := p.complete(req.Params)
		if err != nil {
			resp.Error = &ErrorDetail{Code: "COMPLETION_ERROR", Message: err.Error()}
		} else {
			resp.Result = result
		}
	case "stream":
		// For streaming, we would need to send multiple responses
		// For now, return error as streaming is not fully implemented
		resp.Error = &ErrorDetail{Code: "NOT_IMPLEMENTED", Message: "streaming not yet implemented"}
	case "models":
		result, err := p.models()
		if err != nil {
			resp.Error = &ErrorDetail{Code: "MODELS_ERROR", Message: err.Error()}
		} else {
			resp.Result = result
		}
	default:
		resp.Error = &ErrorDetail{Code: "UNKNOWN_METHOD", Message: fmt.Sprintf("unknown method: %s", req.Method)}
	}

	return resp
}

// complete performs text completion
func (p *Plugin) complete(params json.RawMessage) (any, error) {
	var req struct {
		Messages []Message `json:"messages"`
		Model    string    `json:"model,omitempty"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	model := p.config.Model
	if req.Model != "" {
		model = req.Model
	}

	openaiReq := OpenAIRequest{
		Model:    model,
		Messages: req.Messages,
	}

	body, err := json.Marshal(openaiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", p.config.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	httpResp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		var errResp struct {
			Error struct {
				Message string `json:"message"`
				Type    string `json:"type"`
			} `json:"error"`
		}
		json.NewDecoder(httpResp.Body).Decode(&errResp)
		return nil, fmt.Errorf("API error: %s", errResp.Error.Message)
	}

	var openaiResp OpenAIResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&openaiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no completion returned")
	}

	return map[string]any{
		"content": openaiResp.Choices[0].Message.Content,
		"model":   openaiResp.Model,
		"usage":   openaiResp.Usage,
	}, nil
}

// models returns available models
func (p *Plugin) models() (any, error) {
	httpReq, err := http.NewRequest("GET", p.config.BaseURL+"/models", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	httpResp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status %d", httpResp.StatusCode)
	}

	var result struct {
		Data []ModelInfo `json:"data"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Filter to chat completion models
	var models []ModelInfo
	for _, m := range result.Data {
		// Only include GPT models
		if len(m.ID) >= 3 && (m.ID[:3] == "gpt" || m.ID[:3] == "GPT") {
			models = append(models, m)
		}
	}

	return map[string]any{
		"models": models,
	}, nil
}

func main() {
	plugin := NewPlugin()

	// Read configuration from environment or stdin
	// For stdio protocol, config is passed via the first request or environment
	// In production, the plugin manager will send initialize request

	reader := bufio.NewReader(os.Stdin)
	decoder := json.NewDecoder(reader)

	for {
		var req Request
		if err := decoder.Decode(&req); err != nil {
			// EOF or error, exit gracefully
			if err.Error() == "EOF" {
				break
			}
			resp := Response{
				Version: version,
				Type:    "response",
				Error:   &ErrorDetail{Code: "PARSE_ERROR", Message: err.Error()},
			}
			json.NewEncoder(os.Stdout).Encode(resp)
			continue
		}

		// Handle initialize request specially
		if req.Method == "initialize" {
			var initConfig Config
			if err := json.Unmarshal(req.Params, &initConfig); err == nil {
				if err := plugin.Initialize(initConfig); err != nil {
					resp := Response{
						Version: version,
						ID:      req.ID,
						Type:    "response",
						Error:   &ErrorDetail{Code: "INIT_ERROR", Message: err.Error()},
					}
					json.NewEncoder(os.Stdout).Encode(resp)
					continue
				}
			}
			resp := Response{
				Version: version,
				ID:      req.ID,
				Type:    "response",
				Result: map[string]any{
					"status":  "initialized",
					"version": version,
				},
			}
			json.NewEncoder(os.Stdout).Encode(resp)
			continue
		}

		resp := plugin.Handle(req)
		if err := json.NewEncoder(os.Stdout).Encode(resp); err != nil {
			fmt.Fprintf(os.Stderr, "failed to encode response: %v\n", err)
		}
		os.Stdout.Sync()
	}
}
