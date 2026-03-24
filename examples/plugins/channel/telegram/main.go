// Telegram Channel Plugin for PersonalClaw
// This plugin implements the Channel interface using Telegram Bot API
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
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
	Version string       `json:"version"`
	ID      string       `json:"id"`
	Type    string       `json:"type"`
	Result  any          `json:"result,omitempty"`
	Error   *ErrorDetail `json:"error,omitempty"`
}

// ErrorDetail represents error information
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Config holds the plugin configuration
type Config struct {
	BotToken string `json:"bot_token"`
	Webhook  string `json:"webhook,omitempty"`
}

// TelegramMessage represents a message from Telegram
type TelegramMessage struct {
	UpdateID int      `json:"update_id"`
	Message  Message  `json:"message"`
}

// Message represents a Telegram message
type Message struct {
	MessageID int    `json:"message_id"`
	From      User   `json:"from"`
	Chat      Chat   `json:"chat"`
	Text      string `json:"text"`
	Date      int    `json:"date"`
}

// User represents a Telegram user
type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// Chat represents a Telegram chat
type Chat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

// Plugin is the Telegram channel plugin
type Plugin struct {
	config     Config
	client     *http.Client
	isRunning  bool
	lastUpdateID int
}

// NewPlugin creates a new Telegram plugin
func NewPlugin() *Plugin {
	return &Plugin{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Initialize sets up the plugin with configuration
func (my *Plugin) Initialize(config Config) error {
	my.config = config
	if my.config.BotToken == "" {
		return fmt.Errorf("bot_token is required")
	}
	return nil
}

// Handle processes a request and returns a response
func (my *Plugin) Handle(req Request) Response {
	resp := Response{
		Version: version,
		ID:      req.ID,
		Type:    "response",
	}

	switch req.Method {
	case "start":
		result, err := my.start()
		if err != nil {
			resp.Error = &ErrorDetail{Code: "START_ERROR", Message: err.Error()}
		} else {
			resp.Result = result
		}
	case "stop":
		result, err := my.stop()
		if err != nil {
			resp.Error = &ErrorDetail{Code: "STOP_ERROR", Message: err.Error()}
		} else {
			resp.Result = result
		}
	case "send":
		result, err := my.send(req.Params)
		if err != nil {
			resp.Error = &ErrorDetail{Code: "SEND_ERROR", Message: err.Error()}
		} else {
			resp.Result = result
		}
	case "listen":
		result, err := my.listen()
		if err != nil {
			resp.Error = &ErrorDetail{Code: "LISTEN_ERROR", Message: err.Error()}
		} else {
			resp.Result = result
		}
	default:
		resp.Error = &ErrorDetail{Code: "UNKNOWN_METHOD", Message: fmt.Sprintf("unknown method: %s", req.Method)}
	}

	return resp
}

// start starts the bot
func (my *Plugin) start() (any, error) {
	if my.isRunning {
		return nil, fmt.Errorf("bot is already running")
	}

	// Validate token by calling getMe
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getMe", my.config.BotToken)
	httpResp, err := my.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid bot token")
	}

	var result struct {
		OK     bool `json:"ok"`
		Result struct {
			Username string `json:"username"`
		} `json:"result"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.OK {
		return nil, fmt.Errorf("failed to start bot")
	}

	my.isRunning = true
	return map[string]any{
		"status":   "started",
		"username": result.Result.Username,
	}, nil
}

// stop stops the bot
func (my *Plugin) stop() (any, error) {
	if !my.isRunning {
		return nil, fmt.Errorf("bot is not running")
	}

	my.isRunning = false
	return map[string]any{
		"status": "stopped",
	}, nil
}

// send sends a message
func (my *Plugin) send(params json.RawMessage) (any, error) {
	if !my.isRunning {
		return nil, fmt.Errorf("bot is not running")
	}

	var req struct {
		ChatID string `json:"chat_id"`
		Text   string `json:"text"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if req.ChatID == "" || req.Text == "" {
		return nil, fmt.Errorf("chat_id and text are required")
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", my.config.BotToken)

	payload := map[string]string{
		"chat_id": req.ChatID,
		"text":    req.Text,
	}

	payloadBytes, _ := json.Marshal(payload)

	httpResp, err := my.client.Post(url, "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to send message: status %d", httpResp.StatusCode)
	}

	return map[string]any{
		"sent": true,
	}, nil
}

// listen listens for incoming messages
func (my *Plugin) listen() (any, error) {
	if !my.isRunning {
		return nil, fmt.Errorf("bot is not running")
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates", my.config.BotToken)
	if my.lastUpdateID > 0 {
		url = fmt.Sprintf("%s?offset=%d", url, my.lastUpdateID+1)
	}

	httpResp, err := my.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get updates: %w", err)
	}
	defer httpResp.Body.Close()

	var result struct {
		OK     bool              `json:"ok"`
		Result []TelegramMessage `json:"result"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.OK {
		return nil, fmt.Errorf("failed to get updates")
	}

	var messages []map[string]any
	for _, update := range result.Result {
		if update.UpdateID > my.lastUpdateID {
			my.lastUpdateID = update.UpdateID
		}

		messages = append(messages, map[string]any{
			"message_id": update.Message.MessageID,
			"chat_id":    strconv.FormatInt(update.Message.Chat.ID, 10),
			"from":       update.Message.From.Username,
			"text":       update.Message.Text,
			"date":       update.Message.Date,
		})
	}

	return map[string]any{
		"messages": messages,
	}, nil
}

func main() {
	plugin := NewPlugin()

	reader := bufio.NewReader(os.Stdin)
	decoder := json.NewDecoder(reader)

	for {
		var req Request
		if err := decoder.Decode(&req); err != nil {
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
