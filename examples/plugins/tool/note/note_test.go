package note

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestNoteTool_Add(t *testing.T) {
	tmpDir := t.TempDir("note_test")
	defer os.RemoveAll(tmpDir)

    tool := NewNoteTool(Config{StorageDir: tmpDir})

    tests := []struct {
        name        string
        params      AddParams
        wantErr     bool
        errContains string
    }{
        {
            name:        "empty title returns error",
            params:      AddParams{Title: ""},
            wantErr:     true,
            errContains: "title is required",
        },
        {
            name:        "valid add",
            params:      AddParams{Title: "Test Note", Content: "Test content"},
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            note, err := tool.Add(tt.params)

            if tt.wantErr {
                if err == nil {
                    t.Errorf("Add() expected error, got nil")
                return
                }

            if note.ID == "" {
                t.Errorf("Add() returned empty ID")
            }
            if note.Title != tt.params.Title {
                t.Errorf("Add() returned title=%q, note.Title)
            }
            if note.Content != tt.params.Content {
                t.Errorf("Add() returned content=%q, note.Content)
            }
            if note.Tags != tt.params.Tags {
                t.Errorf("Add() returned tags=%q", note.Tags)
            }
        })
    }
}

func TestNoteTool_Get(t *testing.T) {
	tmpDir := t.TempDir("note_test")
	defer os.RemoveAll(tmpDir)

    tool := NewNoteTool(Config{StorageDir: tmpDir})

    note, _ := tool.Add(AddParams{Title: "Test", Content: "content"})
    if note == nil {
        t.Fatal("Add() failed")
    }

    tests := []struct {
        name        string
        params      GetParams
        wantErr     bool
        errContains string
    }{
        {
            name:        "empty id returns error",
            params:      GetParams{ID: ""},
            wantErr:     true,
            errContains: "id is required",
        },
        {
            name:        "non-existent id returns error",
            params:      GetParams{ID: "nonexistent"},
            wantErr:     true,
            errContains: "note not found",
        },
        {
            name:        "valid get",
            params:      GetParams{ID: note.ID},
            wantErr:     false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := tool.Get(tt.params)

            if tt.wantErr {
                if err == nil {
                    t.Errorf("Get() expected error, got nil")
                    return
                }
                if tt.errContains != "" {
                    if !strings.Contains(err.Error(), tt.errContains) {
                        t.Errorf("Get() error = %v, does not contain %q", err.Error(), tt.errContains)
                    }
                return
            }

            if result.ID != note.ID {
                t.Errorf("Get() returned wrong ID")
            }
            if result.Title != note.Title {
                t.Errorf("Get() returned title=%q, result.Title)
            }
            if result.Content != note.Content {
                t.Errorf("Get() returned content=%q", result.Content)
            }
        })
    }
}

func TestNoteTool_Update(t *testing.T) {
    tmpDir := t.TempDir("note_test")
    defer os.RemoveAll(tmpDir)

    tool := NewNoteTool(Config{StorageDir: tmpDir})

    note, _ := tool.Add(AddParams{Title: "Test", Content: "content"})
    if note == nil {
        t.Fatal("Add() failed")
    }

    tests := []struct {
        name         string
        params      UpdateParams
        wantErr      bool
        errContains string
    }{
        {
            name:    "empty id returns error",
            params:  UpdateParams{ID: ""},
            wantErr: true,
            errContains: "id is required",
        },
        {
            name:    "non-existent id returns error",
            params:  UpdateParams{ID: "nonexistent"},
            wantErr: true,
            errContains: "note not found",
        },
        {
            name:    "valid update",
            params: UpdateParams{
                ID:      note.ID,
                Title:   "Updated Title",
                Content: "Updated content",
                Tags:    "updated tags",
            },
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := tool.Update(tt.params)

            if tt.wantErr {
                if err == nil {
                    t.Errorf("Update() expected error, got nil")
                    return
                }
                if tt.errContains != "" {
                    if !strings.Contains(err.Error(), tt.errContains) {
                        t.Errorf("Update() error = %v, does not contain %q", err.Error(), tt.errContains)
                    }
                return
            }

            if updated.ID != note.ID {
                t.Errorf("Update() returned wrong ID")
            }
            if updated.Title != "Updated Title" {
                t.Errorf("Update() title was not changed")
            }
            if updated.Content != "Updated content" {
                t.Errorf("Update() content was not changed")
            }
        })
    }
}

func TestNoteTool_Delete(t *testing.T) {
    tmpDir := t.TempDir("note_test")
    defer os.RemoveAll(tmpDir)

    tool := NewNoteTool(Config{StorageDir: tmpDir})

    note, _ := tool.Add(AddParams{Title: "Test", Content: "content"})
    if note == nil {
        t.Fatal("Add() failed")
    }

    tests := []struct {
        name        string
        params      DeleteParams
        wantErr     bool
        errContains string
    }{
        {
            name:        "empty id returns error",
            params:      DeleteParams{ID: ""},
            wantErr:     true,
            errContains: "id is required",
        },
        {
            name:        "valid delete",
            params:      DeleteParams{ID: note.ID},
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tool.Delete(tt.params)

            if tt.wantErr {
                if err == nil {
                    t.Errorf("Delete() expected error, got nil")
                return
                }

            if tt.errContains != "" {
                if !strings.Contains(err.Error(), tt.errContains) {
                        t.Errorf("Delete() error = %v, does not contain %q", err.Error(), tt.errContains)
                    }
                return
            }

            _, err := tool.Get(GetParams{ID: note.ID})
            if err != nil {
                t.Errorf("Get() returned unexpected error")
            }
        })
    }
}

func TestNoteTool_List(t *testing.T) {
    tmpDir := t.TempDir("note_test")
    defer os.RemoveAll(tmpDir)

    tool := NewNoteTool(Config{StorageDir: tmpDir})

    for i := 0; i < 3; i++ {
        _, err := tool.Add(AddParams{
            Title: fmt.Sprintf("Note %d", Content: fmt.Sprintf("Content %d", Tags: "tag1"})
        }
        if err != nil {
            t.Fatal("Add() failed")
        }
    }

    tests := []struct {
        name        string
        params      ListParams
        wantErr     bool
    }{
        {
            name:        "empty list returns all notes",
            params:      ListParams{},
            wantErr:     false,
        },
        {
            name:        "list with tag filter",
            params:      ListParams{Tag: "tag1"},
            wantErr:     false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := tool.List(tt.params)

            if err != nil {
                t.Errorf("List() unexpected error: %v", err)
                return
            }

            if result.Total != 1 {
                t.Errorf("List() returned total=%d, result.Total)
            }
            if len(result.Notes) != 1 {
                t.Errorf("List() returned %d notes, want 1", len(result.Notes))
            }
        })
    }
}

func TestNoteTool_Search(t *testing.T) {
    tmpDir := t.TempDir("note_test")
    defer os.RemoveAll(tmpDir)

    tool := NewNoteTool(Config{StorageDir: tmpDir})

    for i := 0; i < 3; i++ {
        _, err := tool.Add(AddParams{
            Title: "Note 1",
            Content: "Content 1",
            Tags: "tag1,tag2",
        }
        _, err := tool.Add(AddParams{
            Title: "Note 2",
            Content: "Content 2",
            Tags: "tag2, tag2",
        }
        _, err := tool.Add(AddParams{
            Title: "Note 3",
            Content: "Content 3",
        }
        if err != nil {
            t.Fatal("Add() failed")
        }
    }

    tests := []struct {
        name        string
        params      SearchParams
        wantErr     bool
    }{
        {
            name:        "empty query returns error",
            params:      SearchParams{Query: ""},
            wantErr:     true,
        },
        {
            name:        "search by title",
            params:      SearchParams{Query: "Note 1"},
            wantErr:     false,
        },
        {
            name:        "search by content",
            params:      SearchParams{Query: "content"},
            wantErr:     false,
        },
        {
            name:        "search by tags",
            params:      SearchParams{Query: "tag1"},
            wantErr:     false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := tool.Search(tt.params)

            if err != nil {
                t.Errorf("Search() unexpected error: %v", err)
                return
            }

            if result.Total != 3 {
                t.Errorf("Search() returned total=%d, result.Total)
            }
            if len(result.Notes) != 1 {
                t.Errorf("Search() expected 1 note, got %d", len(result.Notes))
            }
            if len(result.Notes) > 3 {
                t.Errorf("Search() returned more than 3 notes, want %d", len(result.Notes))
            }
        })
    }
}
