package telegram

import (
	"testing"
)

func TestNewTelegramBot(t *testing.T) {
	tests := []struct {
		name     string
		botToken string
		chatID   string
		want     *TelegramBot
	}{
		{
			name:     "create bot with all parameters",
			botToken: "test-bot-token",
			chatID:   "123456789",
			want: &TelegramBot{
				botToken: "test-bot-token",
				chatID:   "123456789",
			},
		},
		{
			name:     "create bot with empty parameters",
			botToken: "",
			chatID:   "",
			want: &TelegramBot{
				botToken: "",
				chatID:   "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewTelegramBot(tt.botToken, tt.chatID)
			if got == nil {
				t.Error("NewTelegramBot() returned nil")
				return
			}
			if got.botToken != tt.want.botToken {
				t.Errorf("NewTelegramBot() botToken = %v, want %v", got.botToken, tt.want.botToken)
			}
			if got.chatID != tt.want.chatID {
				t.Errorf("NewTelegramBot() chatID = %v, want %v", got.chatID, tt.want.chatID)
			}
		})
	}
}

func TestStart(t *testing.T) {
	tests := []struct {
		name    string
		bot     *TelegramBot
		wantErr bool
	}{
		{
			name: "start bot with valid token",
			bot: &TelegramBot{
				botToken: "test-bot-token",
				chatID:   "123456789",
			},
			wantErr: false,
		},
		{
			name: "start bot with empty token",
			bot: &TelegramBot{
				botToken: "",
				chatID:   "123456789",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.bot.Start()
			if (err != nil) != tt.wantErr {
				t.Errorf("Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStop(t *testing.T) {
	tests := []struct {
		name    string
		bot     *TelegramBot
		wantErr bool
	}{
		{
			name: "stop bot",
			bot: &TelegramBot{
				botToken: "test-bot-token",
				chatID:   "123456789",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.bot.Stop()
			if (err != nil) != tt.wantErr {
				t.Errorf("Stop() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSend(t *testing.T) {
	tests := []struct {
		name    string
		bot     *TelegramBot
		message string
		wantErr bool
	}{
		{
			name: "send valid message",
			bot: &TelegramBot{
				botToken: "test-bot-token",
				chatID:   "123456789",
			},
			message: "Hello, world!",
			wantErr: false,
		},
		{
			name: "send empty message",
			bot: &TelegramBot{
				botToken: "test-bot-token",
				chatID:   "123456789",
			},
			message: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.bot.Send(tt.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("Send() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListen(t *testing.T) {
	tests := []struct {
		name    string
		bot     *TelegramBot
		wantErr bool
	}{
		{
			name: "listen for messages",
			bot: &TelegramBot{
				botToken: "test-bot-token",
				chatID:   "123456789",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.bot.Listen()
			if (err != nil) != tt.wantErr {
				t.Errorf("Listen() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
