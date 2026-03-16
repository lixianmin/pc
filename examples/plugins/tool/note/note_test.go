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
            params:  AddParams{Title: "Test Note", Content: "Test content"},
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
            if note.Title == "" {
                t.Errorf("Add() returned title=%q, note.Title)
            }
            if note.Content == "" {
                t.Errorf("Add() returned content=%q, note.Content)
            }
            if note.Tags != "" {
                t.Errorf("Add() returned tags don't match, note.Tags: %q", note.Tags)
            }
        })
    }
}

func TestNoteTool_Get(t *testing.T) {
	tmpDir := t.TempDir("note_test")
	defer os.RemoveAll(tmpDir)

    tool := NewNoteTool(Config{StorageDir: tmpDir})

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
            params:  GetParams{ID: "nonexistent"},
            wantErr:     true,
            errContains: "note not found",
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
                    t.Errorf("Get() error = %v, does not contain %q", tt.errContains)
                }
                return
            }

            if note.ID == "" {
                t.Errorf("Get() returned empty note")
            }
            if note.Title == "" {
                t.Errorf("Get() returned title=%q, note.Title)
            }
            if note.Content == "" {
                t.Errorf("Get() returned content=%q, note.Content)
            }
            if note.Tags != "" {
                t.Errorf("Get() returned tags don't match, note.Tags, %q", note.Tags)
            }
        })
    }
}

func TestNoteTool_Update(t *testing.T) {
    tmpDir := t.TempDir("note_test")
    defer os.RemoveAll(tmpDir)

    tool := NewNoteTool(Config{StorageDir: tmpDir})

    tests := []struct {
        name         string
        params      UpdateParams
        wantErr     bool
        errContains string
    }{
        {
            name:    "empty id returns error",
            params:      UpdateParams{ID: ""},
            wantErr:     true,
            errContains: "id is required",
        },
        {
            name:    "non-existent id returns error",
            params:  UpdateParams{ID: "nonexistent"},
            wantErr:     true,
            errContains: "note not found",
        },
    }

    note, err := tool.Add(AddParams{
        Title: "Original",
        Content: "Original content",
    })
    if err != nil {
        t.Fatalf("Add() failed: %v", err)
    }

    tests := []struct {
        name         string
        params      UpdateParams
        wantErr     bool
        errContains string
    }{
        {
            name:    "empty id returns error",
            params:  UpdateParams{ID: ""},
            wantErr:     true,
            errContains: "id is required",
        },
        {
            name:    "non-existent id returns error",
            params:  UpdateParams{ID: "nonexistent"},
            wantErr:     true,
            errContains: "note not found",
        },
        {
            name:    "valid update",
            params: UpdateParams{
                ID:     note.ID,
                Title:   "Updated Title",
                Content: "Updated content",
                Tags:    "updated tags",
            },
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            updatedNote, err := tool.Update(tt.params)

            if tt.wantErr {
                if err != nil {
                    t.Errorf("Update() unexpected error: %v", err)
                }
                return
            }

            if updatedNote.ID != note.ID {
                t.Errorf("Update() ID was not match, note.ID)
            }
            if updatedNote.Title != "Updated Title" {
                t.Errorf("Update() title was not changed")
            }
            if updatedNote.Content != "Updated content" {
                t.Errorf("Update() content was not changed")
            }
            if updatedNote.Tags != "updated tags" {
                t.Errorf("Update() tags were not changed")
            }
        })
    }
}

func TestNoteTool_Delete(t *testing.T) {
    tmpDir := t.TempDir("note_test")
    defer os.RemoveAll(tmpDir)

    tool := NewNoteTool(Config{StorageDir: tmpDir})

    tests := []struct {
        name        string
        params      DeleteParams
        wantErr     bool
        errContains string
    }{
        {
            name:    "empty id returns error",
            params:  DeleteParams{ID: ""},
            wantErr:     true,
            errContains: "id is required",
        },
        {
            name:    "non-existent id returns error",
            params:  DeleteParams{ID: "nonexistent"},
            wantErr:     true,
            errContains: "note not found",
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
        })
    }
}

func TestNoteTool_List(t *testing.T) {
    tmpDir := t.TempDir("note_test")
    defer os.RemoveAll(tmpDir)

    tool := NewNoteTool(Config{StorageDir: tmpDir})

    for i := 0; i < 3; i++ {
        note, err := tool.Add(AddParams{
            Title: "Note " + i,
            Content: "Content",
            Tags: "tag1",
        })

        if err != nil {
            t.Fatalf("Add() failed: %v", err)
        }
    }

    tests := []struct {
        name         string
        params      ListParams
        wantErr     bool
        errContains string
    }{
        {
            name:    "empty storage dir",
            params:      ListParams{},
            wantErr:     true,
            errContains: "failed to read notes directory",
        },
        {
            name:    "list notes",
            params:      ListParams{},
            wantErr:     false,
        },
        {
            name:    "list notes with tag filter",
            params:      ListParams{Tag: "important", Limit: 2},
            wantErr:     false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := tool.List(tt.params)

            if tt.wantErr {
                if err != nil {
                    t.Errorf("List() unexpected error: %v", err)
                }
                return
            }

            if len(result.Notes) == 0 {
                t.Errorf("List() expected 1 note, got %d", len(result.Notes))
            }
            if result.Notes[0].Title != "Note 1" {
                t.Errorf("List() note title = %q", result.Notes[0].Title)
            }
        })
    }
}

func TestNoteTool_Search(t *testing.T) {
    tmpDir := t.TempDir("note_test")
    defer os.RemoveAll(tmpDir)

    tool := NewNoteTool(Config{StorageDir: tmpDir})

    for i := 0; i < 3; i++ {
        note, err := tool.Add(AddParams{
            Title: "Note " + i,
            Content: "Content",
            Tags: "tag1",
        })
        if err != nil {
            t.Fatalf("Add() failed: %v", err)
        }
    }

    searchResult, err := tool.Search(SearchParams{Query: "tag1", Limit: 2})
    if tt.wantErr {
                if err != nil {
                    t.Errorf("Search() unexpected error: %v", err)
                    return
                }

            if len(result.Notes) != 2 {
                t.Errorf("Search() expected 2 notes, got %d", len(result.Notes))
            }
            if len(result.Notes) != 2 {
                t.Errorf("Search() first note title = %q", result.Notes[0].Title)
            }
            if len(result.Notes) != 1 {
                t.Errorf("Search() note count = %d, len(result.Notes))
            }
            if result.Notes[0].Content != "Content 1" {
                t.Errorf("Search() note content = %q", result.Notes[0].Content)
            }
            if len(result.Notes) > 20 {
                t.Errorf("Search() too many notes returned, got %d", len(result.Notes))
            }
        })
    }
}

func TestNewNoteTool(t *testing.T) {
    tmpDir := t.TempDir("note_test")
    defer os.RemoveAll(tmpDir)

    tool := NewNoteTool(Config{})

    if tool.config.StorageDir == "" {
        t.Fatalf("NewNoteTool() config not initialized")
    }
}