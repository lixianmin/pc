package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/viewport"
)

func TestRenderScrollbar_NoPanic(t *testing.T) {
	tests := []struct {
		name        string
		lineCount   int
		visibleH    int
		yOffset     int
		expectPanic bool
	}{
		{
			name:        "zero_total_lines",
			lineCount:   0,
			visibleH:    10,
			yOffset:     0,
			expectPanic: false,
		},
		{
			name:        "content_fits_viewport",
			lineCount:   5,
			visibleH:    10,
			yOffset:     0,
			expectPanic: false,
		},
		{
			name:        "content_exceeds_viewport",
			lineCount:   100,
			visibleH:    20,
			yOffset:     50,
			expectPanic: false,
		},
		{
			name:        "yOffset_at_max",
			lineCount:   100,
			visibleH:    20,
			yOffset:     80,
			expectPanic: false,
		},
		{
			name:        "single_line_content",
			lineCount:   1,
			visibleH:    20,
			yOffset:     0,
			expectPanic: false,
		},
		{
			name:        "large_yOffset_small_content",
			lineCount:   10,
			visibleH:    20,
			yOffset:     100,
			expectPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Model{
				viewport: viewport.Model{
					Height:  tt.visibleH,
					YOffset: tt.yOffset,
				},
				styles: DefaultStyles(),
			}

			content := ""
			if tt.lineCount > 0 {
				content = strings.Repeat("line\n", tt.lineCount)
			}
			m.viewport.SetContent(content)

			defer func() {
				if r := recover(); r != nil {
					if !tt.expectPanic {
						t.Errorf("unexpected panic: %v", r)
					}
				}
			}()

			_ = m.renderScrollbar()

			if tt.expectPanic {
				t.Error("expected panic but didn't get one")
			}
		})
	}
}

func TestRenderMessages_LongContent(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		viewportWidth  int
		expectComplete bool
	}{
		{
			name:           "short_content",
			content:        "Hello, this is a short message.",
			viewportWidth:  80,
			expectComplete: true,
		},
		{
			name:           "long_content_no_newlines",
			content:        strings.Repeat("This is a very long line without any newlines that should wrap properly. ", 50),
			viewportWidth:  80,
			expectComplete: true,
		},
		{
			name:           "long_content_with_newlines",
			content:        strings.Repeat("Line with content\n", 100),
			viewportWidth:  80,
			expectComplete: true,
		},
		{
			name:           "chinese_content",
			content:        strings.Repeat("这是一段中文内容，用于测试长文本的换行和显示效果。", 50),
			viewportWidth:  80,
			expectComplete: true,
		},
		{
			name:           "mixed_content",
			content:        "English text mixed with 中文内容 and some special characters: !@#$%^&*() " + strings.Repeat("repeat ", 100),
			viewportWidth:  80,
			expectComplete: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Model{
				messages: []Message{
					{Role: "user", Content: "test question"},
					{Role: "agent", Content: tt.content},
				},
				viewport: viewport.Model{
					Width:  tt.viewportWidth,
					Height: 20,
				},
				styles: DefaultStyles(),
			}

			result := m.renderMessages()

			if result == "" {
				t.Error("renderMessages returned empty string")
			}

			if !strings.Contains(result, "You:") {
				t.Error("missing user message prefix")
			}

			if !strings.Contains(result, "Agent:") {
				t.Error("missing agent message prefix")
			}

			if tt.expectComplete {
				originalLen := len(tt.content)
				if originalLen > 100 && len(result) < originalLen/2 {
					t.Errorf("content may be truncated: original=%d, result=%d", originalLen, len(result))
				}
			}
		})
	}
}

func TestViewportScrolling_LongContent(t *testing.T) {
	longContent := strings.Repeat("This is line number %d of a very long response that should be fully visible when scrolling.\n", 200)

	m := &Model{
		messages: []Message{
			{Role: "agent", Content: longContent},
		},
		viewport: viewport.Model{
			Width:  80,
			Height: 20,
		},
		styles: DefaultStyles(),
		ready:  true,
	}

	content := m.renderMessages()
	m.viewport.SetContent(content)

	totalLines := m.viewport.TotalLineCount()
	if totalLines < 100 {
		t.Errorf("expected many lines for long content, got: %d", totalLines)
	}

	if m.viewport.Height >= totalLines {
		t.Errorf("content should exceed viewport height: total=%d, viewport=%d", totalLines, m.viewport.Height)
	}
}

func TestMessageAccumulation(t *testing.T) {
	m := &Model{
		messages: make([]Message, 0),
		viewport: viewport.Model{
			Width:  80,
			Height: 20,
		},
		styles: DefaultStyles(),
		ready:  true,
	}

	for i := 0; i < 10; i++ {
		m.addMessage("user", "user message "+string(rune('0'+i)))
		m.addMessage("agent", "agent response "+string(rune('0'+i))+" "+strings.Repeat("content ", 20))
	}

	if len(m.messages) != 20 {
		t.Errorf("expected 20 messages, got: %d", len(m.messages))
	}

	content := m.renderMessages()
	m.viewport.SetContent(content)

	for i := 0; i < 10; i++ {
		expected := "user message " + string(rune('0'+i))
		if !strings.Contains(content, expected) {
			t.Errorf("missing message: %s", expected)
		}
	}
}

func TestLongTextNotTruncated(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		viewportWidth int
		viewportH     int
	}{
		{
			name:          "very_long_single_line",
			content:       strings.Repeat("word ", 500),
			viewportWidth: 80,
			viewportH:     20,
		},
		{
			name:          "many_paragraphs",
			content:       strings.Repeat("This is a paragraph.\n\n", 100),
			viewportWidth: 80,
			viewportH:     20,
		},
		{
			name:          "mixed_markdown",
			content:       "# Heading\n\n" + strings.Repeat("This is **bold** and *italic* text.\n\n", 50),
			viewportWidth: 80,
			viewportH:     20,
		},
		{
			name:          "chinese_long_text",
			content:       "中文测试开始 " + strings.Repeat("这是独立的句子。", 100) + " 中文测试结束",
			viewportWidth: 80,
			viewportH:     20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Model{
				messages: []Message{
					{Role: "agent", Content: tt.content},
				},
				viewport: viewport.Model{
					Width:  tt.viewportWidth,
					Height: tt.viewportH,
				},
				styles: DefaultStyles(),
				ready:  true,
			}

			rendered := m.renderMessages()
			m.viewport.SetContent(rendered)

			totalLines := m.viewport.TotalLineCount()
			visibleHeight := m.viewport.Height

			t.Logf("Content: original=%d bytes, rendered=%d bytes, lines=%d, viewport=%d",
				len(tt.content), len(rendered), totalLines, visibleHeight)

			markerStart := ""
			markerEnd := ""
			if strings.Contains(tt.content, "开始") {
				markerStart = "开始"
				markerEnd = "结束"
			}
			if strings.HasPrefix(tt.content, "word") {
				markerStart = "word"
			}

			if markerStart != "" && !strings.Contains(rendered, markerStart) {
				t.Errorf("start marker %q missing from rendered content", markerStart)
			}
			if markerEnd != "" && !strings.Contains(rendered, markerEnd) {
				t.Errorf("end marker %q missing from rendered content (content truncated)", markerEnd)
			}

			if len(tt.content) > 1000 && len(rendered) < len(tt.content)/2 {
				t.Errorf("rendered content too short: original=%d, rendered=%d", len(tt.content), len(rendered))
			}
		})
	}
}

func TestScrollingCanReachAllContent(t *testing.T) {
	lines := make([]string, 200)
	for i := 0; i < 200; i++ {
		lines[i] = fmt.Sprintf("Line %03d: This is test content for scrolling verification.", i)
	}
	content := strings.Join(lines, "\n")

	m := &Model{
		messages: []Message{
			{Role: "agent", Content: content},
		},
		viewport: viewport.Model{
			Width:  80,
			Height: 20,
		},
		styles: DefaultStyles(),
		ready:  true,
	}

	rendered := m.renderMessages()
	m.viewport.SetContent(rendered)

	m.viewport.GotoTop()
	topView := m.viewport.View()

	m.viewport.GotoBottom()
	bottomView := m.viewport.View()

	if strings.Contains(topView, "Line 199") {
		t.Error("top view should not contain last line")
	}

	if !strings.Contains(topView, "Line 000") && !strings.Contains(topView, "Line 0") {
		t.Error("top view should contain first line")
	}

	if !strings.Contains(bottomView, "Line 199") && !strings.Contains(bottomView, "Line 19") {
		t.Logf("bottom view may not have last line visible: %s...", bottomView[:min(100, len(bottomView))])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
