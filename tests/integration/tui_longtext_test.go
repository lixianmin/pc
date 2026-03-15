package integration

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/lixianmin/pc/internal/tui"
)

func TestTUILongTextDisplay(t *testing.T) {
	tests := []struct {
		name          string
		contentLength int
		contentType   string
	}{
		{
			name:          "short_text",
			contentLength: 100,
			contentType:   "english",
		},
		{
			name:          "medium_text",
			contentLength: 1000,
			contentType:   "english",
		},
		{
			name:          "long_text",
			contentLength: 5000,
			contentType:   "english",
		},
		{
			name:          "very_long_text",
			contentLength: 10000,
			contentType:   "english",
		},
		{
			name:          "chinese_long_text",
			contentLength: 5000,
			contentType:   "chinese",
		},
		{
			name:          "mixed_long_text",
			contentLength: 5000,
			contentType:   "mixed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var content string
			switch tt.contentType {
			case "english":
				content = strings.Repeat("This is a test sentence for long text display verification. ", tt.contentLength/60)
			case "chinese":
				content = strings.Repeat("这是用于测试长文本显示的中文句子。", tt.contentLength/15)
			case "mixed":
				content = strings.Repeat("English and 中文混合 content for testing. ", tt.contentLength/40)
			}

			m := tui.NewModel(nil)
			m.AddMessage("agent", content)
			m.SetViewportSize(80, 20)

			rendered := m.RenderMessages()

			if rendered == "" {
				t.Error("rendered content is empty")
			}

			originalLen := len(content)
			renderedLen := len(rendered)

			if renderedLen < originalLen/2 {
				t.Errorf("content appears truncated: original=%d, rendered=%d", originalLen, renderedLen)
			}

			t.Logf("Content rendered: original=%d, rendered=%d", originalLen, renderedLen)
		})
	}
}

func TestTUIScrollToTop(t *testing.T) {
	lines := make([]string, 200)
	for i := 0; i < 200; i++ {
		lines[i] = fmt.Sprintf("Line %03d: This is test content for scrolling verification.", i)
	}
	content := strings.Join(lines, "\n")

	m := tui.NewModel(nil)
	m.AddMessage("agent", content)
	m.SetViewportSize(80, 20)

	rendered := m.RenderMessages()
	vp := viewport.New(80, 20)
	vp.SetContent(rendered)

	vp.GotoTop()
	topView := vp.View()

	if !strings.Contains(topView, "Line 000") && !strings.Contains(topView, "Line 0") {
		t.Error("top view should contain first line")
	}

	vp.GotoBottom()
	bottomView := vp.View()

	if !strings.Contains(bottomView, "Line 199") && !strings.Contains(bottomView, "Line 19") {
		t.Errorf("bottom view should contain last line, got: %s...", bottomView[:min(100, len(bottomView))])
	}

	t.Logf("Successfully scrolled through %d lines", 200)
}

func TestTUIMultipleLongMessages(t *testing.T) {
	m := tui.NewModel(nil)

	for i := 0; i < 5; i++ {
		content := fmt.Sprintf("Message %d: %s", i, strings.Repeat("Long content line. ", 100))
		m.AddMessage("agent", content)
	}

	rendered := m.RenderMessages()
	m.SetViewportSize(80, 20)
	vp := viewport.New(80, 20)
	vp.SetContent(rendered)

	totalLines := vp.TotalLineCount()
	if totalLines < 50 {
		t.Errorf("expected many lines for multiple long messages, got: %d", totalLines)
	}

	vp.GotoTop()
	topView := vp.View()

	if !strings.Contains(topView, "Message 0") {
		t.Error("top view should contain first message")
	}

	vp.GotoBottom()
	bottomView := vp.View()

	if !strings.Contains(bottomView, "Message 4") && !strings.Contains(bottomView, "content line") {
		t.Errorf("bottom view should contain last message content")
	}

	t.Logf("Multiple long messages: %d total lines", totalLines)
}

func TestTUIViewportPagination(t *testing.T) {
	content := strings.Repeat("Line of content for pagination test.\n", 100)

	m := tui.NewModel(nil)
	m.AddMessage("agent", content)
	rendered := m.RenderMessages()

	vp := viewport.New(80, 20)
	vp.SetContent(rendered)

	vp.GotoTop()
	initialY := vp.YOffset

	vp.ViewDown()
	afterDown := vp.YOffset

	if afterDown <= initialY {
		t.Errorf("ViewDown should increase YOffset: before=%d, after=%d", initialY, afterDown)
	}

	vp.GotoBottom()
	bottomY := vp.YOffset

	if bottomY <= afterDown {
		t.Errorf("GotoBottom should increase YOffset: before=%d, after=%d", afterDown, bottomY)
	}

	vp.GotoTop()
	finalY := vp.YOffset

	if finalY != 0 {
		t.Errorf("GotoTop should reset YOffset to 0, got: %d", finalY)
	}

	t.Logf("Pagination works: top=%d, middle=%d, bottom=%d", initialY, afterDown, bottomY)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
