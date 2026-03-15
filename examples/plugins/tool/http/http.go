package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Config struct {
	Timeout      time.Duration
	MaxBodySize  int64
	AllowedHosts []string
	BlockedHosts []string
}

type HTTPTool struct {
	config Config
	client *http.Client
}

func NewHTTPTool(config Config) *HTTPTool {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxBodySize == 0 {
		config.MaxBodySize = 1024 * 1024
	}
	return &HTTPTool{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

type RequestParams struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

type RequestResult struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Size       int64             `json:"size"`
}

func (t *HTTPTool) Request(params RequestParams) (*RequestResult, error) {
	if params.URL == "" {
		return nil, fmt.Errorf("URL is required")
	}

	if err := t.validateURL(params.URL); err != nil {
		return nil, err
	}

	if params.Method == "" {
		params.Method = "GET"
	}

	var body io.Reader
	if params.Body != "" {
		body = bytes.NewReader([]byte(params.Body))
	}

	req, err := http.NewRequest(params.Method, params.URL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for k, v := range params.Headers {
		req.Header.Set(k, v)
	}

	if params.Body != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, t.config.MaxBodySize))
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	return &RequestResult{
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Body:       string(respBody),
		Size:       int64(len(respBody)),
	}, nil
}

type GetParams struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}

func (t *HTTPTool) Get(params GetParams) (*RequestResult, error) {
	return t.Request(RequestParams{
		URL:     params.URL,
		Method:  "GET",
		Headers: params.Headers,
	})
}

type PostParams struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

func (t *HTTPTool) Post(params PostParams) (*RequestResult, error) {
	return t.Request(RequestParams{
		URL:     params.URL,
		Method:  "POST",
		Headers: params.Headers,
		Body:    params.Body,
	})
}

func (t *HTTPTool) validateURL(url string) error {
	if len(t.config.BlockedHosts) > 0 {
		for _, blocked := range t.config.BlockedHosts {
			if len(url) >= len(blocked) && url[:len(blocked)] == blocked {
				return fmt.Errorf("access to host is blocked: %s", blocked)
			}
		}
	}

	if len(t.config.AllowedHosts) > 0 {
		allowed := false
		for _, a := range t.config.AllowedHosts {
			if len(url) >= len(a) && url[:len(a)] == a {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("host not in allowed list")
		}
	}

	return nil
}
