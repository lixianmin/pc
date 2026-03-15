package engine

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

type ToolCall struct {
	Name   string
	Params map[string]interface{}
}

type ToolResult struct {
	Name   string
	Output string
	Error  error
}

func ParseToolCalls(content string) ([]ToolCall, error) {
	if strings.TrimSpace(content) == "" {
		return nil, nil
	}

	reInvoke := regexp.MustCompile(`(?s)<invoke>\s*<name>([^<]+)</name>\s*<params>([^<]*)</params>\s*</invoke>`)
	matches := reInvoke.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil, nil
	}

	var toolCalls []ToolCall
	for _, match := range matches {
		if len(match) != 3 {
			continue
		}

		toolName := strings.TrimSpace(match[1])
		paramsJSON := strings.TrimSpace(match[2])

		if toolName == "" {
			continue
		}

		var params map[string]interface{}
		if paramsJSON != "" {
			if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
				params = map[string]interface{}{"_raw": paramsJSON}
			}
		}

		toolCalls = append(toolCalls, ToolCall{
			Name:   toolName,
			Params: params,
		})
	}

	return toolCalls, nil
}

func HasToolCalls(content string) bool {
	if strings.TrimSpace(content) == "" {
		return false
	}

	re := regexp.MustCompile(`(?s)<invoke>.*?</invoke>`)
	return re.MatchString(content)
}

func FormatToolResult(result ToolResult) string {
	var parts []string

	parts = append(parts, "<tool_result>")
	parts = append(parts, fmt.Sprintf("<name>%s</name>", result.Name))

	if result.Error != nil {
		parts = append(parts, fmt.Sprintf("<error>%s</error>", escapeXML(result.Error.Error())))
	} else {
		parts = append(parts, fmt.Sprintf("<output>%s</output>", escapeXML(result.Output)))
	}

	parts = append(parts, "</tool_result>")

	return strings.Join(parts, "\n")
}

func FormatToolResults(results []ToolResult) string {
	if len(results) == 0 {
		return ""
	}

	var parts []string
	for _, result := range results {
		parts = append(parts, FormatToolResult(result))
	}

	return strings.Join(parts, "\n\n")
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

func unescapeXML(s string) string {
	s = strings.ReplaceAll(s, "&apos;", "'")
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&amp;", "&")
	return s
}
