package memory

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewLongTermService(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "create new long-term service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewLongTermService()
			if got == nil {
				t.Error("NewLongTermService() returned nil")
			}
		})
	}
}

func TestStore(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		tags     []string
		priority int
		wantErr  bool
	}{
		{
			name:     "store valid memory",
			content:  "Buy groceries",
			tags:     []string{"shopping", "todo"},
			priority: 5,
			wantErr:  false,
		},
		{
			name:     "store memory with empty content",
			content:  "",
			tags:     []string{},
			priority: 1,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewLongTermService()
			id, err := s.Store(tt.content, tt.tags, tt.priority)
			if (err != nil) != tt.wantErr {
				t.Errorf("Store() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && id == "" {
				t.Error("Store() returned empty ID")
			}
		})
	}
}

func TestRetrieve(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*LongTermService)
		query   string
		limit   int
		wantMin int
		wantErr bool
	}{
		{
			name: "retrieve by query",
			setup: func(s *LongTermService) {
				s.Store("Buy groceries", []string{"shopping"}, 5)
				s.Store("Call mom", []string{"family"}, 3)
			},
			query:   "shopping",
			limit:   10,
			wantMin: 1,
			wantErr: false,
		},
		{
			name: "retrieve with empty query",
			setup: func(s *LongTermService) {
				s.Store("Test memory", []string{}, 1)
			},
			query:   "",
			limit:   10,
			wantMin: 1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewLongTermService()
			tt.setup(s)

			entries, err := s.Retrieve(tt.query, tt.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("Retrieve() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && len(entries) < tt.wantMin {
				t.Errorf("Retrieve() count = %v, want at least %v", len(entries), tt.wantMin)
			}
		})
	}
}

func TestRetrieveByTimeRange(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		setup   func(*LongTermService)
		start   time.Time
		end     time.Time
		wantMin int
		wantErr bool
	}{
		{
			name: "retrieve by time range",
			setup: func(s *LongTermService) {
				s.Store("Memory 1", []string{}, 1)
			},
			start:   now.Add(-24 * time.Hour),
			end:     now.Add(24 * time.Hour),
			wantMin: 1,
			wantErr: false,
		},
		{
			name: "retrieve outside time range",
			setup: func(s *LongTermService) {
				s.Store("Memory 2", []string{}, 1)
			},
			start:   now.Add(24 * time.Hour),
			end:     now.Add(48 * time.Hour),
			wantMin: 0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewLongTermService()
			tt.setup(s)

			entries, err := s.RetrieveByTimeRange(tt.start, tt.end)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveByTimeRange() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && len(entries) < tt.wantMin {
				t.Errorf("RetrieveByTimeRange() count = %v, want at least %v", len(entries), tt.wantMin)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*LongTermService) string
		wantErr bool
	}{
		{
			name: "delete existing memory",
			setup: func(s *LongTermService) string {
				id, _ := s.Store("Test memory", []string{}, 1)
				return id
			},
			wantErr: false,
		},
		{
			name: "delete non-existent memory",
			setup: func(s *LongTermService) string {
				return "01HABCDEFGHJKMNPQRSTVWXYZ0" // ULID format string
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewLongTermService()
			id := tt.setup(s)

			err := s.Delete(id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCleanup(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*LongTermService)
		wantErr bool
	}{
		{
			name: "cleanup expired memories",
			setup: func(s *LongTermService) {
				s.Store("Expired memory", []string{}, 1)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewLongTermService()
			tt.setup(s)

			err := s.Cleanup()
			if (err != nil) != tt.wantErr {
				t.Errorf("Cleanup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCloseLongTerm(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "close service",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewLongTermService()
			err := s.Close()
			if (err != nil) != tt.wantErr {
				t.Errorf("CloseLongTerm() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPersistentLongTermMemory(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "memory.json")

	t.Run("create and load persistent memory", func(t *testing.T) {
		// Create service with storage
		s1, err := NewPersistentLongTermService(storagePath)
		if err != nil {
			t.Fatalf("NewPersistentLongTermService() error = %v", err)
		}

		// Store some memories
		id1, err := s1.Store("Memory 1", []string{"tag1"}, 5)
		if err != nil {
			t.Fatalf("Store() error = %v", err)
		}

		id2, err := s1.Store("Memory 2", []string{"tag2"}, 3)
		if err != nil {
			t.Fatalf("Store() error = %v", err)
		}

		// Save to disk
		if err := s1.Save(); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		if err := s1.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}

		// Create new service instance and load
		s2, err := NewPersistentLongTermService(storagePath)
		if err != nil {
			t.Fatalf("NewPersistentLongTermService() load error = %v", err)
		}
		defer s2.Close()

		// Verify memories are loaded
		results, err := s2.Retrieve("", 10)
		if err != nil {
			t.Fatalf("Retrieve() error = %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 memories, got %d", len(results))
		}

		// Verify IDs match
		foundIDs := make(map[string]bool)
		for _, r := range results {
			foundIDs[r.ID] = true
		}
		if !foundIDs[id1] {
			t.Errorf("Memory %s not found", id1)
		}
		if !foundIDs[id2] {
			t.Errorf("Memory %s not found", id2)
		}
	})

	t.Run("auto save on close", func(t *testing.T) {
		storagePath := filepath.Join(tmpDir, "memory_auto.json")

		s1, err := NewPersistentLongTermService(storagePath)
		if err != nil {
			t.Fatalf("NewPersistentLongTermService() error = %v", err)
		}

		id, err := s1.Store("Auto save test", []string{"auto"}, 1)
		if err != nil {
			t.Fatalf("Store() error = %v", err)
		}

		// Close should auto-save
		if err := s1.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(storagePath); os.IsNotExist(err) {
			t.Error("Storage file not created after close")
		}

		// Load and verify
		s2, err := NewPersistentLongTermService(storagePath)
		if err != nil {
			t.Fatalf("NewPersistentLongTermService() load error = %v", err)
		}
		defer s2.Close()

		results, err := s2.Retrieve("", 10)
		if err != nil {
			t.Fatalf("Retrieve() error = %v", err)
		}

		found := false
		for _, r := range results {
			if r.ID == id {
				found = true
				break
			}
		}
		if !found {
			t.Error("Memory not found after auto-save")
		}
	})

	t.Run("persist with expiration", func(t *testing.T) {
		storagePath := filepath.Join(tmpDir, "memory_expire.json")

		s, err := NewPersistentLongTermService(storagePath)
		if err != nil {
			t.Fatalf("NewPersistentLongTermService() error = %v", err)
		}
		defer s.Close()

		// Store with future expiration
		futureTime := time.Now().Add(24 * time.Hour).UnixMilli()
		id, err := s.StoreWithExpiration("Expires soon", []string{"temp"}, 1, futureTime)
		if err != nil {
			t.Fatalf("StoreWithExpiration() error = %v", err)
		}

		if err := s.Save(); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		// Load and verify expiration is preserved
		s2, err := NewPersistentLongTermService(storagePath)
		if err != nil {
			t.Fatalf("NewPersistentLongTermService() load error = %v", err)
		}
		defer s2.Close()

		results, err := s2.Retrieve("", 10)
		if err != nil {
			t.Fatalf("Retrieve() error = %v", err)
		}

		if len(results) != 1 {
			t.Fatalf("Expected 1 memory, got %d", len(results))
		}

		if results[0].ID != id {
			t.Errorf("ID mismatch: got %s, want %s", results[0].ID, id)
		}

		if results[0].ExpiresAt != futureTime {
			t.Errorf("ExpiresAt mismatch: got %d, want %d", results[0].ExpiresAt, futureTime)
		}
	})
}

