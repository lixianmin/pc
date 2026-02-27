package tui

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// History manages command history for the TUI.
type History struct {
	filePath string
	items    []string
	index    int
	maxSize  int
}

// NewHistory creates a new history manager.
func NewHistory(filePath string) *History {
	return &History{
		filePath: filePath,
		items:    make([]string, 0),
		index:    -1,
		maxSize:  1000,
	}
}

// Load loads history from file.
func (h *History) Load() error {
	if h.filePath == "" {
		return nil
	}

	file, err := os.Open(h.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			h.items = append(h.items, line)
		}
	}

	// Limit history size
	if len(h.items) > h.maxSize {
		h.items = h.items[len(h.items)-h.maxSize:]
	}

	h.index = len(h.items)
	return scanner.Err()
}

// Save saves history to file.
func (h *History) Save() error {
	if h.filePath == "" {
		return nil
	}

	// Ensure directory exists
	dir := filepath.Dir(h.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(h.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, item := range h.items {
		if _, err := writer.WriteString(item + "\n"); err != nil {
			return err
		}
	}

	return writer.Flush()
}

// Add adds a new item to history.
func (h *History) Add(item string) {
	item = strings.TrimSpace(item)
	if item == "" {
		return
	}

	// Don't add duplicates at the end
	if len(h.items) > 0 && h.items[len(h.items)-1] == item {
		return
	}

	h.items = append(h.items, item)

	// Limit size
	if len(h.items) > h.maxSize {
		h.items = h.items[1:]
	}

	h.index = len(h.items)
}

// Previous returns the previous history item.
func (h *History) Previous() string {
	if len(h.items) == 0 {
		return ""
	}

	h.index--
	if h.index < 0 {
		h.index = 0
	}

	return h.items[h.index]
}

// Next returns the next history item.
func (h *History) Next() string {
	if len(h.items) == 0 {
		return ""
	}

	h.index++
	if h.index >= len(h.items) {
		h.index = len(h.items)
		return ""
	}

	return h.items[h.index]
}

// Reset resets the history navigation index.
func (h *History) Reset() {
	h.index = len(h.items)
}
