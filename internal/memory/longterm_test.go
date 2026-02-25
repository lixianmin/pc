package memory

import (
	"testing"
	"time"

	"github.com/google/uuid"
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
				return uuid.New().String()
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

