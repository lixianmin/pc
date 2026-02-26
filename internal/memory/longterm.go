package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
)

// LongTermMemory is the interface for long-term memory.
type LongTermMemory interface {
	// Store stores a memory with metadata.
	Store(content string, tags []string, priority int) (string, error)

	// Retrieve retrieves memories by query.
	Retrieve(query string, limit int) ([]MemoryEntry, error)

	// RetrieveByTimeRange retrieves memories within a time range.
	RetrieveByTimeRange(start, end time.Time) ([]MemoryEntry, error)

	// Delete deletes a memory by ID.
	Delete(id string) error

	// Cleanup removes expired memories.
	Cleanup() error

	// Close closes the memory service.
	Close() error
}

// MemoryEntry represents a stored memory entry.
type MemoryEntry struct {
	ID        string   `json:"id"`
	Content   string   `json:"content"`
	Tags      []string `json:"tags"`
	Priority  int      `json:"priority"`
	CreateAt  int64    `json:"create_at"`  // Creation time (Unix milliseconds)
	UpdateAt  int64    `json:"update_at"`  // Last update time (Unix milliseconds)
	ExpiresAt int64    `json:"expires_at"` // Expiration time (0 means never expires, Unix milliseconds)
}

// LongTermService is the long-term memory service implementation.
type LongTermService struct {
	entries []MemoryEntry
}

// NewLongTermService creates a new long-term memory service.
func NewLongTermService() *LongTermService {
	return &LongTermService{
		entries: []MemoryEntry{},
	}
}

// Store stores a memory with metadata.
func (my *LongTermService) Store(content string, tags []string, priority int) (string, error) {
	if content == "" {
		return "", fmt.Errorf("content cannot be empty")
	}

	id := ulid.Make().String()
	now := time.Now().UnixMilli()
	entry := MemoryEntry{
		ID:       id,
		Content:  content,
		Tags:     tags,
		Priority: priority,
		CreateAt: now,
		UpdateAt: now,
	}

	my.entries = append(my.entries, entry)
	return id, nil
}

// Retrieve retrieves memories by query.
func (my *LongTermService) Retrieve(query string, limit int) ([]MemoryEntry, error) {
	if limit <= 0 {
		limit = 10 // default limit
	}

	var results []MemoryEntry

	// If query is empty, return all entries sorted by priority
	if query == "" {
		for _, entry := range my.entries {
			results = append(results, entry)
			if len(results) >= limit {
				break
			}
		}
		return results, nil
	}

	// Search by content or tags (case-insensitive)
	queryLower := strings.ToLower(query)
	for _, entry := range my.entries {
		contentMatch := strings.Contains(strings.ToLower(entry.Content), queryLower)
		tagMatch := false
		for _, tag := range entry.Tags {
			if strings.Contains(strings.ToLower(tag), queryLower) {
				tagMatch = true
				break
			}
		}

		if contentMatch || tagMatch {
			results = append(results, entry)
			if len(results) >= limit {
				break
			}
		}
	}

	return results, nil
}

// RetrieveByTimeRange retrieves memories within a time range.
func (my *LongTermService) RetrieveByTimeRange(start, end time.Time) ([]MemoryEntry, error) {
	var results []MemoryEntry
	startMs := start.UnixMilli()
	endMs := end.UnixMilli()

	for _, entry := range my.entries {
		// Check if entry was created within the time range
		if entry.CreateAt >= startMs && entry.CreateAt <= endMs {
			results = append(results, entry)
		}
	}

	return results, nil
}

// Delete deletes a memory by ID.
func (my *LongTermService) Delete(id string) error {
	for i, entry := range my.entries {
		if entry.ID == id {
			// Remove entry from slice
			my.entries = append(my.entries[:i], my.entries[i+1:]...)
			return nil
		}
	}
	// Don't return error if memory not found (idempotent)
	return nil
}

// Cleanup removes expired memories.
func (my *LongTermService) Cleanup() error {
	now := time.Now().UnixMilli()
	var remaining []MemoryEntry

	for _, entry := range my.entries {
		// If entry has no expiration (0) or it's not expired, keep it
		if entry.ExpiresAt == 0 || entry.ExpiresAt > now {
			remaining = append(remaining, entry)
		}
	}

	my.entries = remaining
	return nil
}

// Close closes the memory service.
func (my *LongTermService) Close() error {
	my.entries = []MemoryEntry{}
	return nil
}

// PersistentLongTermService is a long-term memory service with file persistence.
type PersistentLongTermService struct {
	*LongTermService
	storagePath string
}

// NewPersistentLongTermService creates a new persistent long-term memory service.
// If the storage file exists, it will be loaded automatically.
func NewPersistentLongTermService(storagePath string) (*PersistentLongTermService, error) {
	s := &PersistentLongTermService{
		LongTermService: NewLongTermService(),
		storagePath:     storagePath,
	}

	// Try to load existing data
	if err := s.Load(); err != nil {
		// It's OK if file doesn't exist yet
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load memory: %w", err)
		}
	}

	return s, nil
}

// StoreWithExpiration stores a memory with expiration time.
func (my *PersistentLongTermService) StoreWithExpiration(content string, tags []string, priority int, expiresAt int64) (string, error) {
	if content == "" {
		return "", fmt.Errorf("content cannot be empty")
	}

	id := ulid.Make().String()
	now := time.Now().UnixMilli()
	entry := MemoryEntry{
		ID:        id,
		Content:   content,
		Tags:      tags,
		Priority:  priority,
		CreateAt:  now,
		UpdateAt:  now,
		ExpiresAt: expiresAt,
	}

	my.entries = append(my.entries, entry)
	return id, nil
}

// Save persists memories to the storage file.
func (my *PersistentLongTermService) Save() error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(my.storagePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(my.entries, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal memories: %w", err)
	}

	// Write to file
	if err := os.WriteFile(my.storagePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write storage file: %w", err)
	}

	return nil
}

// Load loads memories from the storage file.
func (my *PersistentLongTermService) Load() error {
	data, err := os.ReadFile(my.storagePath)
	if err != nil {
		return err
	}

	var entries []MemoryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("failed to unmarshal memories: %w", err)
	}

	my.entries = entries
	return nil
}

// Close saves and closes the memory service.
func (my *PersistentLongTermService) Close() error {
	if err := my.Save(); err != nil {
		return fmt.Errorf("failed to save on close: %w", err)
	}
	return my.LongTermService.Close()
}
