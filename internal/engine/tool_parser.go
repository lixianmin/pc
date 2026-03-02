package engine

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// ToolCall represents a parsed tool call from LLM output.
type ToolCall struct {
	Name   string                 // Tool name
	Params map[string]interface{} // Tool parameters
}

// ToolResult represents the result of a tool execution.
type ToolResult struct {
	Name   string // Tool name
	Output string // Tool output
	Error  error  // Execution error if any
}

// ParseToolCalls parses tool calls from LLM output content.
// It extracts <tool_call> XML tags and returns a list of ToolCall.
//
// Supported format:
//   <tool_call>
//   <name>tool_name</name>
//   <params>{"key": "value"}</params>
//   </tool_call>
//
// Returns empty slice if no tool calls found.
func ParseToolCalls(content string) ([]ToolCall, error) {
	if strings.TrimSpace(content) == "" {
		return nil, nil
	}

	// Regular expression to match tool_call blocks
	// Using (?s) flag to make . match newlines
	re := regexp.MustCompile(`(?s)<tool_call>\s*<name>([^<]+)</name>\s*<params>([^<]*)</params>\s*</tool_call>`)

	matches := re.FindAllStringSubmatch(content, -1)
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

		// Parse JSON params
		var params map[string]interface{}
		if paramsJSON != "" {
			if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
				// If JSON parsing fails, store as raw string
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

// HasToolCalls checks if the content contains any tool calls.
func HasToolCalls(content string) bool {
	if strings.TrimSpace(content) == "" {
		return false
	}

	re := regexp.MustCompile(`(?s)<tool_call>.*?</tool_call>`)
	return re.MatchString(content)
}

// FormatToolResult formats a tool result for inclusion in the conversation.
// This format is used to feed tool execution results back to the LLM.
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

// FormatToolResults formats multiple tool results.
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

// escapeXML escapes special XML characters.
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// unescapeXML unescapes special XML characters.
func unescapeXML(s string) string {
	s = strings.ReplaceAll(s, "&apos;", "'")
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&amp;", "&")
	return s
}
