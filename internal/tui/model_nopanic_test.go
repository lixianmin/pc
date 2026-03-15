package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

func TestModelInitNoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Model.Init() panicked: %v", r)
		}
	}()

	m := NewModel(nil)
	cmd := m.Init()

	if cmd == nil {
		t.Error("Init() should return a command")
	}
}

func TestModelUpdateNoPanic(t *testing.T) {
	tests := []struct {
		name string
		msg  tea.Msg
	}{
		{
			name: "window_size_msg",
			msg:  tea.WindowSizeMsg{Width: 80, Height: 24},
		},
		{
			name: "key_msg_ctrl_c",
			msg:  tea.KeyMsg{Type: tea.KeyCtrlC},
		},
		{
			name: "key_msg_enter",
			msg:  tea.KeyMsg{Type: tea.KeyEnter},
		},
		{
			name: "response_msg",
			msg:  responseMsg("test response"),
		},
		{
			name: "error_msg",
			msg:  errorMsg("test error"),
		},
		{
			name: "status_msg",
			msg:  statusMsg("test status"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Model.Update() panicked with %v: %v", tt.name, r)
				}
			}()

			m := NewModel(nil)
			_, cmd := m.Update(tt.msg)

			t.Logf("Update(%s) executed successfully, cmd=%v", tt.name, cmd != nil)
		})
	}
}

func TestModelViewNoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Model.View() panicked: %v", r)
		}
	}()

	m := NewModel(nil)
	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}

	t.Logf("View() returned: %d bytes", len(view))
}

func TestModelUpdateWithLongContentNoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Model with long content panicked: %v", r)
		}
	}()

	m := NewModel(nil)

	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	longContent := strings.Repeat("This is a very long line that should be handled properly without any panic. ", 1000)
	m.Update(responseMsg(longContent))

	view := m.View()
	if view == "" {
		t.Error("View() should not return empty string even with long content")
	}

	t.Logf("Long content test passed: %d bytes in view", len(view))
}

func TestRenderScrollbarEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		totalLines  int
		visibleH    int
		yOffset     int
		shouldPanic bool
	}{
		{
			name:        "empty_content",
			totalLines:  0,
			visibleH:    20,
			yOffset:     0,
			shouldPanic: false,
		},
		{
			name:        "content_equals_viewport",
			totalLines:  20,
			visibleH:    20,
			yOffset:     0,
			shouldPanic: false,
		},
		{
			name:        "large_content",
			totalLines:  10000,
			visibleH:    20,
			yOffset:     5000,
			shouldPanic: false,
		},
		{
			name:        "very_large_yoffset",
			totalLines:  100,
			visibleH:    20,
			yOffset:     999999,
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.shouldPanic {
						t.Errorf("renderScrollbar() panicked unexpectedly: %v", r)
					}
				}
			}()

			m := NewModel(nil)
			m.viewport = viewport.Model{
				Width:   80,
				Height:  tt.visibleH,
				YOffset: tt.yOffset,
			}

			content := ""
			if tt.totalLines > 0 {
				content = strings.Repeat("line\n", tt.totalLines)
			}
			m.viewport.SetContent(content)

			result := m.renderScrollbar()

			if tt.shouldPanic {
				t.Error("expected panic but didn't get one")
			}

			t.Logf("renderScrollbar(%s) returned: %d bytes", tt.name, len(result))
		})
	}
}

func TestModelCompleteWorkflow(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("complete workflow panicked: %v", r)
		}
	}()

	m := NewModel(nil)

	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	for i := 0; i < 10; i++ {
		userMsg := strings.Repeat("User message content ", 20)
		m.Update(responseMsg(userMsg))

		agentMsg := strings.Repeat("Agent response content ", 30)
		m.Update(responseMsg(agentMsg))
	}

	view := m.View()
	if view == "" {
		t.Error("View() should not be empty after workflow")
	}

	m.viewport.GotoTop()
	topView := m.View()

	m.viewport.GotoBottom()
	bottomView := m.View()

	t.Logf("Complete workflow test passed: top=%d bytes, bottom=%d bytes", len(topView), len(bottomView))
}
