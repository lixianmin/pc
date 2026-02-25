package openai

import (
	"testing"
)

func TestNewGLMClient(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		model   string
		baseURL string
		want    *GLMClient
	}{
		{
			name:    "create client with all parameters",
			apiKey:  "test-api-key",
			model:   "glm-4.7",
			baseURL: "https://open.bigmodel.cn/api/coding/paas/v4",
			want: &GLMClient{
				apiKey:  "test-api-key",
				model:   "glm-4.7",
				baseURL: "https://open.bigmodel.cn/api/coding/paas/v4",
			},
		},
		{
			name:    "create client with default baseURL and model",
			apiKey:  "test-api-key",
			model:   "",
			baseURL: "",
			want: &GLMClient{
				apiKey:  "test-api-key",
				model:   "glm-4.7",
				baseURL: "https://open.bigmodel.cn/api/coding/paas/v4",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewGLMClient(tt.apiKey, tt.model, tt.baseURL)
			if got == nil {
				t.Error("NewGLMClient() returned nil")
				return
			}
			if got.apiKey != tt.want.apiKey {
				t.Errorf("NewGLMClient() apiKey = %v, want %v", got.apiKey, tt.want.apiKey)
			}
			if got.model != tt.want.model {
				t.Errorf("NewGLMClient() model = %v, want %v", got.model, tt.want.model)
			}
			if got.baseURL != tt.want.baseURL {
				t.Errorf("NewGLMClient() baseURL = %v, want %v", got.baseURL, tt.want.baseURL)
			}
		})
	}
}

func TestComplete(t *testing.T) {
	tests := []struct {
		name    string
		client  *GLMClient
		prompt  string
		options map[string]any
		wantErr bool
	}{
		{
			name: "complete with valid prompt",
			client: &GLMClient{
				apiKey:  "test-api-key",
				model:   "glm-4.7",
				baseURL: "https://open.bigmodel.cn/api/coding/paas/v4",
			},
			prompt:  "Hello, world!",
			options: map[string]any{},
			wantErr: false,
		},
		{
			name: "complete with empty prompt",
			client: &GLMClient{
				apiKey:  "test-api-key",
				model:   "glm-4.7",
				baseURL: "https://open.bigmodel.cn/api/coding/paas/v4",
			},
			prompt:  "",
			options: map[string]any{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.client.Complete(tt.prompt, tt.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("Complete() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got == "" {
				t.Error("Complete() returned empty result")
			}
		})
	}
}

func TestStream(t *testing.T) {
	tests := []struct {
		name    string
		client  *GLMClient
		prompt  string
		options map[string]any
		wantErr bool
	}{
		{
			name: "stream with valid prompt",
			client: &GLMClient{
				apiKey:  "test-api-key",
				model:   "glm-4.7",
				baseURL: "https://open.bigmodel.cn/api/coding/paas/v4",
			},
			prompt:  "Hello, world!",
			options: map[string]any{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.client.Stream(tt.prompt, tt.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("Stream() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got == nil {
				t.Error("Stream() returned nil channel")
			}
		})
	}
}

func TestModels(t *testing.T) {
	tests := []struct {
		name    string
		client  *GLMClient
		wantErr bool
	}{
		{
			name: "get models list",
			client: &GLMClient{
				apiKey:  "test-api-key",
				model:   "glm-4.7",
				baseURL: "https://open.bigmodel.cn/api/coding/paas/v4",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.client.Models()
			if (err != nil) != tt.wantErr {
				t.Errorf("Models() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && len(got) == 0 {
				t.Error("Models() returned empty list")
			}
		})
	}
}
