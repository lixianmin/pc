package tui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHistory_Add(t *testing.T) {
	h := NewHistory("")

	// Add items
	h.Add("first")
	h.Add("second")
	h.Add("third")

	if len(h.items) != 3 {
		t.Errorf("expected 3 items, got %d", len(h.items))
	}

	// Add duplicate (should not be added)
	h.Add("third")
	if len(h.items) != 3 {
		t.Errorf("expected 3 items after duplicate, got %d", len(h.items))
	}

	// Add empty string (should not be added)
	h.Add("  ")
	if len(h.items) != 3 {
		t.Errorf("expected 3 items after empty, got %d", len(h.items))
	}
}

func TestHistory_Navigation(t *testing.T) {
	h := NewHistory("")
	h.Add("first")
	h.Add("second")
	h.Add("third")

	tests := []struct {
		name     string
		action   func() string
		expected string
	}{
		{
			name: "previous from end returns last",
			action: func() string {
				return h.Previous()
			},
			expected: "third",
		},
		{
			name: "previous again returns second",
			action: func() string {
				return h.Previous()
			},
			expected: "second",
		},
		{
			name: "next returns third",
			action: func() string {
				return h.Next()
			},
			expected: "third",
		},
		{
			name: "next at end returns empty",
			action: func() string {
				return h.Next()
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.action()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestHistory_NavigationBounds(t *testing.T) {
	h := NewHistory("")

	// Navigation on empty history
	if h.Previous() != "" {
		t.Error("expected empty string for Previous on empty history")
	}
	if h.Next() != "" {
		t.Error("expected empty string for Next on empty history")
	}

	// Add one item
	h.Add("only")

	// Previous beyond start stays at first
	h.Previous()
	h.Previous()
	h.Previous()
	if h.index != 0 {
		t.Errorf("expected index 0, got %d", h.index)
	}
}

func TestHistory_LoadSave(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()
	historyFile := filepath.Join(tempDir, "history")

	// Create history and add items
	h1 := NewHistory(historyFile)
	h1.Add("line1")
	h1.Add("line2")
	h1.Add("line3")

	// Save
	if err := h1.Save(); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	// Load into new history
	h2 := NewHistory(historyFile)
	if err := h2.Load(); err != nil {
		t.Fatalf("failed to load: %v", err)
	}

	if len(h2.items) != 3 {
		t.Errorf("expected 3 items after load, got %d", len(h2.items))
	}

	// Verify order
	if h2.items[0] != "line1" || h2.items[1] != "line2" || h2.items[2] != "line3" {
		t.Errorf("items not in expected order: %v", h2.items)
	}
}

func TestHistory_LoadNonExistent(t *testing.T) {
	h := NewHistory("/nonexistent/path/history")
	if err := h.Load(); err != nil {
		t.Errorf("expected no error for non-existent file, got %v", err)
	}
}

func TestHistory_MaxSize(t *testing.T) {
	h := NewHistory("")
	h.maxSize = 5

	// Add more than maxSize items
	for i := 0; i < 10; i++ {
		h.Add(string(rune('a' + i)))
	}

	if len(h.items) != 5 {
		t.Errorf("expected 5 items (maxSize), got %d", len(h.items))
	}

	// Should contain the last 5 items (f-j)
	expected := []string{"f", "g", "h", "i", "j"}
	for i, exp := range expected {
		if h.items[i] != exp {
			t.Errorf("expected item %d to be %q, got %q", i, exp, h.items[i])
		}
	}
}

func TestHistory_Reset(t *testing.T) {
	h := NewHistory("")
	h.Add("first")
	h.Add("second")

	// Move index back
	h.Previous()
	if h.index != 1 {
		t.Errorf("expected index 1 after Previous, got %d", h.index)
	}

	// Reset
	h.Reset()
	if h.index != 2 {
		t.Errorf("expected index 2 after Reset, got %d", h.index)
	}
}

func TestHistory_EmptyPath(t *testing.T) {
	h := NewHistory("")
	h.Add("item")

	// Save should succeed without error
	if err := h.Save(); err != nil {
		t.Errorf("expected no error for empty path save, got %v", err)
	}

	// Load should succeed without error
	if err := h.Load(); err != nil {
		t.Errorf("expected no error for empty path load, got %v", err)
	}
}

func TestHistory_SaveCreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()
	nestedDir := filepath.Join(tempDir, "nested", "dir")
	historyFile := filepath.Join(nestedDir, "history")

	h := NewHistory(historyFile)
	h.Add("item")

	// Directory doesn't exist yet
	if _, err := os.Stat(nestedDir); !os.IsNotExist(err) {
		t.Fatal("directory should not exist initially")
	}

	// Save should create directory
	if err := h.Save(); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	// Directory should now exist
	if _, err := os.Stat(nestedDir); os.IsNotExist(err) {
		t.Error("directory should exist after save")
	}
}
