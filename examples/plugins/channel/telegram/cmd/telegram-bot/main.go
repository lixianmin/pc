package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/lixianmin/pc/examples/plugins/channel/telegram"
	"github.com/lixianmin/pc/pkg/protocol"
)

func main() {
	bot := createBot()

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

		handleRequest(req, bot)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
		os.Exit(1)
	}
}

func createBot() *telegram.TelegramBot {
	// Load configuration from config.yml
	// For now, use environment variables or defaults
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		botToken = "your-bot-token-here"
	}

	chatID := os.Getenv("TELEGRAM_CHAT_ID")
	if chatID == "" {
		chatID = "your-chat-id-here"
	}

	return telegram.NewTelegramBot(botToken, chatID)
}

func handleRequest(req *protocol.Request, bot *telegram.TelegramBot) {
	var result any
	var err error

	switch req.Method {
	case "start":
		result, err = handleStart(bot)
	case "stop":
		result, err = handleStop(bot)
	case "send":
		result, err = handleSend(req.Params, bot)
	case "listen":
		result, err = handleListen(bot)
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

func handleStart(bot *telegram.TelegramBot) (any, error) {
	err := bot.Start()
	return map[string]any{
		"started": err == nil,
	}, err
}

func handleStop(bot *telegram.TelegramBot) (any, error) {
	err := bot.Stop()
	return map[string]any{
		"stopped": err == nil,
	}, err
}

func handleSend(params any, bot *telegram.TelegramBot) (any, error) {
	// Parse params
	var p struct {
		Message string `json:"message"`
	}

	if err := parseParams(params, &p); err != nil {
		return nil, err
	}

	err := bot.Send(p.Message)
	return map[string]any{
		"sent": err == nil,
	}, err
}

func handleListen(bot *telegram.TelegramBot) (any, error) {
	ch, err := bot.Listen()
	if err != nil {
		return nil, err
	}

	// Collect messages for response
	messages := []string{}
	for msg := range ch {
		messages = append(messages, msg)
	}

	return map[string]any{
		"messages": messages,
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
