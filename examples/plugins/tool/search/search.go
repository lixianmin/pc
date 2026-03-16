package search

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type Config struct {
	Timeout    time.Duration
	MaxResults int
	UserAgent  string
}

type SearchTool struct {
	config Config
	client *http.Client
}

func NewSearchTool(config Config) *SearchTool {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxResults == 0 {
		config.MaxResults = 5
	}
	if config.UserAgent == "" {
		config.UserAgent = "Mozilla/5.0 (compatible; PersonalClawBot/1.0)"
	}
	return &SearchTool{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

type SearchParams struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

func (t *SearchTool) Search(params SearchParams) ([]SearchResult, error) {
	if params.Query == "" {
		return nil, fmt.Errorf("query is required")
	}

	limit := params.Limit
	if limit <= 0 {
		limit = t.config.MaxResults
	}

	results, err := t.searchDuckDuckGo(params.Query, limit)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	return results, nil
}

func (t *SearchTool) searchDuckDuckGo(query string, limit int) ([]SearchResult, error) {
	searchURL := fmt.Sprintf(
		"https://html.duckduckgo.com/html/?q=%s",
		url.QueryEscape(query),
	)

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", t.config.UserAgent)
	req.Header.Set("Accept", "text/html")

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return t.parseDuckDuckGoResults(string(body), limit), nil
}

func (t *SearchTool) parseDuckDuckGoResults(html string, limit int) []SearchResult {
	var results []SearchResult

	resultRegex := regexp.MustCompile(`<a[^>]*class="result__a"[^>]*href="([^"]*)"[^>]*>([^<]*)</a>`)
	snippetRegex := regexp.MustCompile(`<a[^>]*class="result__snippet"[^>]*>([^<]*)</a>`)

	resultMatches := resultRegex.FindAllStringSubmatch(html, -1)
	snippetMatches := snippetRegex.FindAllStringSubmatch(html, -1)

	for i, match := range resultMatches {
		if i >= limit {
			break
		}

		resultURL := match[1]
		title := strings.TrimSpace(match[2])

		if strings.HasPrefix(resultURL, "//duckduckgo.com/l/?uddg=") {
			if decoded, err := url.QueryUnescape(strings.TrimPrefix(resultURL, "//duckduckgo.com/l/?uddg=")); err == nil {
				resultURL = decoded
				if idx := strings.Index(resultURL, "&rut="); idx > 0 {
					resultURL = resultURL[:idx]
				}
			}
		}

		snippet := ""
		if i < len(snippetMatches) {
			snippet = strings.TrimSpace(snippetMatches[i][1])
		}

		if resultURL != "" && title != "" {
			results = append(results, SearchResult{
				Title:   title,
				URL:     resultURL,
				Snippet: snippet,
			})
		}
	}

	return results
}
