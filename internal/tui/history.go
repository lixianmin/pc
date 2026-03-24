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
func (my *History) Load() error {
	if my.filePath == "" {
		return nil
	}

	file, err := os.Open(my.filePath)
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
			my.items = append(my.items, line)
		}
	}

	if len(my.items) > my.maxSize {
		my.items = my.items[len(my.items)-my.maxSize:]
	}

	my.index = len(my.items)
	return scanner.Err()
}

// Save saves history to file.
func (my *History) Save() error {
	if my.filePath == "" {
		return nil
	}

	dir := filepath.Dir(my.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(my.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, item := range my.items {
		if _, err := writer.WriteString(item + "\n"); err != nil {
			return err
		}
	}

	return writer.Flush()
}

// Add adds a new item to history.
func (my *History) Add(item string) {
	item = strings.TrimSpace(item)
	if item == "" {
		return
	}

	if len(my.items) > 0 && my.items[len(my.items)-1] == item {
		return
	}

	my.items = append(my.items, item)

	if len(my.items) > my.maxSize {
		my.items = my.items[1:]
	}

	my.index = len(my.items)
}

// Previous returns the previous history item.
func (my *History) Previous() string {
	if len(my.items) == 0 {
		return ""
	}

	my.index--
	if my.index < 0 {
		my.index = 0
	}

	return my.items[my.index]
}

func (my *History) Next() string {
	if len(my.items) == 0 {
		return ""
	}

	my.index++
	if my.index >= len(my.items) {
		my.index = len(my.items)
		return ""
	}

	return my.items[my.index]
}

func (my *History) Reset() {
	my.index = len(my.items)
}
