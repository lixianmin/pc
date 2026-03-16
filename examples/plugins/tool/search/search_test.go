package search

import (
	"testing"
)

func TestSearchTool_Search(t *testing.T) {
	tests := []struct {
		name        string
		params      SearchParams
		wantErr     bool
		errContains string
		skipNetwork bool
	}{
		{
			name:        "empty query returns error",
			params:      SearchParams{Query: ""},
			wantErr:     true,
			errContains: "query is required",
		},
		{
			name:        "valid query (network)",
			params:      SearchParams{Query: "golang", Limit: 3},
			wantErr:     false,
			skipNetwork: true,
		},
	}

	tool := NewSearchTool(Config{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipNetwork {
				t.Skip("skipping network test")
			}

			results, err := tool.Search(tt.params)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Search() expected error, got nil")
					return
				}
				if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("Search() error = %v, want containing %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("Search() unexpected error: %v", err)
				return
			}

			if len(results) > tt.params.Limit {
				t.Errorf("Search() returned %d results, want at most %d", len(results), tt.params.Limit)
			}

			for i, r := range results {
				if r.Title == "" {
					t.Errorf("Search() result[%d] has empty title", i)
				}
				if r.URL == "" {
					t.Errorf("Search() result[%d] has empty URL", i)
				}
			}
		})
	}
}

func TestParseDuckDuckGoResults(t *testing.T) {
	tool := NewSearchTool(Config{})

	html := `
		<div class="result">
			<a class="result__a" href="https://example.com/result1">Example Result 1</a>
			<a class="result__snippet">This is snippet 1</a>
		</div>
		<div class="result">
			<a class="result__a" href="https://example.com/result2">Example Result 2</a>
			<a class="result__snippet">This is snippet 2</a>
		</div>
	`

	results := tool.parseDuckDuckGoResults(html, 5)

	if len(results) != 2 {
		t.Errorf("parseDuckDuckGoResults() returned %d results, want 2", len(results))
		return
	}

	if results[0].Title != "Example Result 1" {
		t.Errorf("parseDuckDuckGoResults() results[0].Title = %v, want 'Example Result 1'", results[0].Title)
	}

	if results[0].Snippet != "This is snippet 1" {
		t.Errorf("parseDuckDuckGoResults() results[0].Snippet = %v, want 'This is snippet 1'", results[0].Snippet)
	}
}

func TestNewSearchTool(t *testing.T) {
	tests := []struct {
		name          string
		config        Config
		wantTimeout   bool
		wantMaxRes    int
		wantUserAgent bool
	}{
		{
			name:          "default config",
			config:        Config{},
			wantTimeout:   true,
			wantMaxRes:    5,
			wantUserAgent: true,
		},
		{
			name:          "custom config",
			config:        Config{Timeout: 10000000000, MaxResults: 10, UserAgent: "CustomBot"},
			wantTimeout:   true,
			wantMaxRes:    10,
			wantUserAgent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := NewSearchTool(tt.config)

			if tool.config.Timeout <= 0 {
				t.Errorf("NewSearchTool() timeout = %v, want > 0", tool.config.Timeout)
			}

			if tool.config.MaxResults != tt.wantMaxRes {
				t.Errorf("NewSearchTool() maxResults = %v, want %v", tool.config.MaxResults, tt.wantMaxRes)
			}

			if tool.config.UserAgent == "" {
				t.Errorf("NewSearchTool() userAgent is empty")
			}

			if tool.client == nil {
				t.Errorf("NewSearchTool() client is nil")
			}
		})
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && s[0:len(substr)] == substr
}
