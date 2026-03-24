package engine

import (
	"fmt"
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
