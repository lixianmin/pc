package engine

import (
	"errors"
	"testing"
)

func TestParseToolCalls(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantCalls   []ToolCall
		wantErr     bool
	}{
		{
			name:    "单个工具调用",
			content: `<tool_call>
<name>shell</name>
<params>{"command": "ls ~", "description": "List home directory"}</params>
</tool_call>`,
			wantCalls: []ToolCall{
				{
					Name: "shell",
					Params: map[string]interface{}{
						"command":     "ls ~",
						"description": "List home directory",
					},
				},
			},
		},
		{
			name: "多个工具调用",
			content: `<tool_call>
<name>shell</name>
<params>{"command": "pwd"}</params>
</tool_call>
<tool_call>
<name>file</name>
<params>{"action": "read", "path": "/etc/hosts"}</params>
</tool_call>`,
			wantCalls: []ToolCall{
				{
					Name: "shell",
					Params: map[string]interface{}{
						"command": "pwd",
					},
				},
				{
					Name: "file",
					Params: map[string]interface{}{
						"action": "read",
						"path":   "/etc/hosts",
					},
				},
			},
		},
		{
			name:    "带有思考内容的工具调用",
			content: `<thinking>用户想要查看主目录文件</thinking>
<tool_call>
<name>shell</name>
<params>{"command": "ls ~"}</params>
</tool_call>`,
			wantCalls: []ToolCall{
				{
					Name: "shell",
					Params: map[string]interface{}{
						"command": "ls ~",
					},
				},
			},
		},
		{
			name:    "没有工具调用的普通回复",
			content: "这是一个普通的回复，没有工具调用。",
			wantCalls: nil,
		},
		{
			name:    "空内容",
			content: "",
			wantCalls: nil,
		},
		{
			name:    "只有空白字符",
			content: "   \n\t   ",
			wantCalls: nil,
		},
		{
			name:    "格式错误的JSON参数",
			content: `<tool_call>
<name>shell</name>
<params>invalid json</params>
</tool_call>`,
			wantCalls: []ToolCall{
				{
					Name: "shell",
					Params: map[string]interface{}{
						"_raw": "invalid json",
					},
				},
			},
		},
		{
			name:    "空参数",
			content: `<tool_call>
<name>shell</name>
<params></params>
</tool_call>`,
			wantCalls: []ToolCall{
				{
					Name:   "shell",
					Params: map[string]interface{}{},
				},
			},
		},
		{
			name:    "嵌套JSON对象",
			content: `<tool_call>
<name>api</name>
<params>{"url": "https://api.example.com", "headers": {"Authorization": "Bearer token"}, "body": {"key": "value"}}</params>
</tool_call>`,
			wantCalls: []ToolCall{
				{
					Name: "api",
					Params: map[string]interface{}{
						"url":     "https://api.example.com",
						"headers": map[string]interface{}{"Authorization": "Bearer token"},
						"body":    map[string]interface{}{"key": "value"},
					},
				},
			},
		},
		{
			name:    "数组参数",
			content: `<tool_call>
<name>search</name>
<params>{"keywords": ["go", "programming"], "limit": 10}</params>
</tool_call>`,
			wantCalls: []ToolCall{
				{
					Name: "search",
					Params: map[string]interface{}{
						"keywords": []interface{}{"go", "programming"},
						"limit":    float64(10),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseToolCalls(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseToolCalls() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != len(tt.wantCalls) {
				t.Errorf("ParseToolCalls() returned %d calls, want %d", len(got), len(tt.wantCalls))
				return
			}

			for i, want := range tt.wantCalls {
				if got[i].Name != want.Name {
					t.Errorf("ParseToolCalls() call[%d].Name = %v, want %v", i, got[i].Name, want.Name)
				}

				if len(got[i].Params) != len(want.Params) {
					t.Errorf("ParseToolCalls() call[%d].Params length = %d, want %d", i, len(got[i].Params), len(want.Params))
					continue
				}

				for key, wantVal := range want.Params {
					gotVal, exists := got[i].Params[key]
					if !exists {
						t.Errorf("ParseToolCalls() call[%d].Params missing key %q", i, key)
						continue
					}

					// Handle different types
					switch wantV := wantVal.(type) {
					case map[string]interface{}:
						gotV, ok := gotVal.(map[string]interface{})
						if !ok {
							t.Errorf("ParseToolCalls() call[%d].Params[%q] type mismatch", i, key)
						}
						for k, v := range wantV {
							if gotV[k] != v {
								t.Errorf("ParseToolCalls() call[%d].Params[%q][%q] = %v, want %v", i, key, k, gotV[k], v)
							}
						}
					case []interface{}:
						gotV, ok := gotVal.([]interface{})
						if !ok {
							t.Errorf("ParseToolCalls() call[%d].Params[%q] type mismatch", i, key)
						}
						if len(gotV) != len(wantV) {
							t.Errorf("ParseToolCalls() call[%d].Params[%q] length = %d, want %d", i, key, len(gotV), len(wantV))
						}
						for j, v := range wantV {
							if gotV[j] != v {
								t.Errorf("ParseToolCalls() call[%d].Params[%q][%d] = %v, want %v", i, key, j, gotV[j], v)
							}
						}
					default:
						if gotVal != wantVal {
							t.Errorf("ParseToolCalls() call[%d].Params[%q] = %v, want %v", i, key, gotVal, wantVal)
						}
					}
				}
			}
		})
	}
}

func TestHasToolCalls(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{
			name:    "有工具调用",
			content: `<tool_call><name>shell</name><params>{}</params></tool_call>`,
			want:    true,
		},
		{
			name:    "没有工具调用",
			content: "普通回复",
			want:    false,
		},
		{
			name:    "空内容",
			content: "",
			want:    false,
		},
		{
			name:    "只有开始标签",
			content: `<tool_call>`,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasToolCalls(tt.content); got != tt.want {
				t.Errorf("HasToolCalls() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatToolResult(t *testing.T) {
	tests := []struct {
		name   string
		result ToolResult
		want   string
	}{
		{
			name: "成功结果",
			result: ToolResult{
				Name:   "shell",
				Output: "Desktop Documents Downloads",
				Error:  nil,
			},
			want: `<tool_result>
<name>shell</name>
<output>Desktop Documents Downloads</output>
</tool_result>`,
		},
		{
			name: "错误结果",
			result: ToolResult{
				Name:   "shell",
				Output: "",
				Error:  errors.New("command not found"),
			},
			want: `<tool_result>
<name>shell</name>
<error>command not found</error>
</tool_result>`,
		},
		{
			name: "包含特殊字符的输出",
			result: ToolResult{
				Name:   "shell",
				Output: "<script>alert('xss')</script>",
				Error:  nil,
			},
			want: `<tool_result>
<name>shell</name>
<output>&lt;script&gt;alert(&apos;xss&apos;)&lt;/script&gt;</output>
</tool_result>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatToolResult(tt.result)
			if got != tt.want {
				t.Errorf("FormatToolResult() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatToolResults(t *testing.T) {
	results := []ToolResult{
		{Name: "shell", Output: "output1", Error: nil},
		{Name: "file", Output: "output2", Error: nil},
	}

	got := FormatToolResults(results)

	if !contains(got, "<name>shell</name>") {
		t.Error("FormatToolResults() should contain shell result")
	}
	if !contains(got, "<name>file</name>") {
		t.Error("FormatToolResults() should contain file result")
	}

	// Test empty results
	emptyGot := FormatToolResults([]ToolResult{})
	if emptyGot != "" {
		t.Errorf("FormatToolResults([]) = %q, want empty string", emptyGot)
	}
}

func TestEscapeXML(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"<", "&lt;"},
		{">", "&gt;"},
		{"&", "&amp;"},
		{"\"", "&quot;"},
		{"'", "&apos;"},
		{"hello", "hello"},
		{"<div>test</div>", "&lt;div&gt;test&lt;/div&gt;"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := escapeXML(tt.input)
			if got != tt.want {
				t.Errorf("escapeXML(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestUnescapeXML(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"&lt;", "<"},
		{"&gt;", ">"},
		{"&amp;", "&"},
		{"&quot;", "\""},
		{"&apos;", "'"},
		{"hello", "hello"},
		{"&lt;div&gt;test&lt;/div&gt;", "<div>test</div>"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := unescapeXML(tt.input)
			if got != tt.want {
				t.Errorf("unescapeXML(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr, 0))
}

func containsAt(s, substr string, start int) bool {
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
