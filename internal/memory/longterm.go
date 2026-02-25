package memory

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
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
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Tags      []string  `json:"tags"`
	Priority  int       `json:"priority"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
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
func (s *LongTermService) Store(content string, tags []string, priority int) (string, error) {
	if content == "" {
		return "", fmt.Errorf("content cannot be empty")
	}

	id := uuid.New().String()
	entry := MemoryEntry{
		ID:        id,
		Content:   content,
		Tags:      tags,
		Priority:  priority,
		CreatedAt: time.Now(),
	}

	s.entries = append(s.entries, entry)
	return id, nil
}

// Retrieve retrieves memories by query.
func (s *LongTermService) Retrieve(query string, limit int) ([]MemoryEntry, error) {
	if limit <= 0 {
		limit = 10 // default limit
	}

	var results []MemoryEntry

	// If query is empty, return all entries sorted by priority
	if query == "" {
		for _, entry := range s.entries {
			results = append(results, entry)
			if len(results) >= limit {
				break
			}
		}
		return results, nil
	}

	// Search by content or tags (case-insensitive)
	queryLower := strings.ToLower(query)
	for _, entry := range s.entries {
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
func (s *LongTermService) RetrieveByTimeRange(start, end time.Time) ([]MemoryEntry, error) {
	var results []MemoryEntry

	for _, entry := range s.entries {
		// Check if entry was created within the time range
		if (entry.CreatedAt.Equal(start) || entry.CreatedAt.After(start)) &&
			(entry.CreatedAt.Equal(end) || entry.CreatedAt.Before(end)) {
			results = append(results, entry)
		}
	}

	return results, nil
}

// Delete deletes a memory by ID.
func (s *LongTermService) Delete(id string) error {
	for i, entry := range s.entries {
		if entry.ID == id {
			// Remove entry from slice
			s.entries = append(s.entries[:i], s.entries[i+1:]...)
			return nil
		}
	}
	// Don't return error if memory not found (idempotent)
	return nil
}

// Cleanup removes expired memories.
func (s *LongTermService) Cleanup() error {
	now := time.Now()
	var remaining []MemoryEntry

	for _, entry := range s.entries {
		// If entry has expiration date and it's not expired, keep it
		if entry.ExpiresAt.IsZero() || entry.ExpiresAt.After(now) {
			remaining = append(remaining, entry)
		}
	}

	s.entries = remaining
	return nil
}

// Close closes the memory service.
func (s *LongTermService) Close() error {
	s.entries = []MemoryEntry{}
	return nil
}
