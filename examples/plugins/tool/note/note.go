package note

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	StorageDir string
	MaxNotes   int
}

type NoteTool struct {
	config Config
}

type Note struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Tags      string `json:"tags"`
}

func NewNoteTool(config Config) *NoteTool {
	if config.StorageDir == "" {
		homeDir, _ := os.UserHomeDir()
		config.StorageDir = filepath.Join(homeDir, ".pc", "notes")
	}
	if config.MaxNotes == 0 {
		config.MaxNotes = 1000
	}

	os.MkdirAll(config.StorageDir, 0755)

	return &NoteTool{config: config}
}

type AddParams struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Tags    string `json:"tags"`
}

func (t *NoteTool) Add(params AddParams) (*Note, error) {
	if params.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	now := time.Now().Format(time.RFC3339)
	id := fmt.Sprintf("%d", time.Now().UnixNano())

	note := &Note{
		ID:        id,
		Title:     params.Title,
		Content:   params.Content,
		CreatedAt: now,
		UpdatedAt: now,
		Tags:      params.Tags,
	}

	if err := t.saveNote(note); err != nil {
		return nil, err
	}

	return note, nil
}

type GetParams struct {
	ID string `json:"id"`
}

func (t *NoteTool) Get(params GetParams) (*Note, error) {
	if params.ID == "" {
		return nil, fmt.Errorf("id is required")
	}

	note, err := t.loadNote(params.ID)
	if err != nil {
		return nil, fmt.Errorf("note not found: %w", err)
	}

	return note, nil
}

type UpdateParams struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Tags    string `json:"tags"`
}

func (t *NoteTool) Update(params UpdateParams) (*Note, error) {
	if params.ID == "" {
		return nil, fmt.Errorf("id is required")
	}

	note, err := t.loadNote(params.ID)
	if err != nil {
		return nil, fmt.Errorf("note not found: %w", err)
	}

	if params.Title != "" {
		note.Title = params.Title
	}
	if params.Content != "" {
		note.Content = params.Content
	}
	if params.Tags != "" {
		note.Tags = params.Tags
	}
	note.UpdatedAt = time.Now().Format(time.RFC3339)

	if err := t.saveNote(note); err != nil {
		return nil, err
	}

	return note, nil
}

type DeleteParams struct {
	ID string `json:"id"`
}

func (t *NoteTool) Delete(params DeleteParams) error {
	if params.ID == "" {
		return fmt.Errorf("id is required")
	}

	path := t.notePath(params.ID)
	return os.Remove(path)
}

type ListParams struct {
	Tag   string `json:"tag"`
	Limit int    `json:"limit"`
}

type ListResult struct {
	Notes []NoteSummary `json:"notes"`
	Total int           `json:"total"`
}

type NoteSummary struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
	Tags      string `json:"tags"`
}

func (t *NoteTool) List(params ListParams) (*ListResult, error) {
	limit := params.Limit
	if limit <= 0 {
		limit = 50
	}

	files, err := os.ReadDir(t.config.StorageDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read notes directory: %w", err)
	}

	var notes []NoteSummary
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		id := strings.TrimSuffix(file.Name(), ".json")
		note, err := t.loadNote(id)
		if err != nil {
			continue
		}

		if params.Tag != "" && !strings.Contains(note.Tags, params.Tag) {
			continue
		}

		notes = append(notes, NoteSummary{
			ID:        note.ID,
			Title:     note.Title,
			CreatedAt: note.CreatedAt,
			Tags:      note.Tags,
		})

		if len(notes) >= limit {
			break
		}
	}

	return &ListResult{
		Notes: notes,
		Total: len(notes),
	}, nil
}

type SearchParams struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

type SearchResult struct {
	Notes []Note `json:"notes"`
	Total int    `json:"total"`
}

func (t *NoteTool) Search(params SearchParams) (*SearchResult, error) {
	if params.Query == "" {
		return nil, fmt.Errorf("query is required")
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}

	files, err := os.ReadDir(t.config.StorageDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read notes directory: %w", err)
	}

	query := strings.ToLower(params.Query)
	var notes []Note

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		id := strings.TrimSuffix(file.Name(), ".json")
		note, err := t.loadNote(id)
		if err != nil {
			continue
		}

		if strings.Contains(strings.ToLower(note.Title), query) ||
			strings.Contains(strings.ToLower(note.Content), query) ||
			strings.Contains(strings.ToLower(note.Tags), query) {
			notes = append(notes, *note)
			if len(notes) >= limit {
				break
			}
		}
	}

	return &SearchResult{
		Notes: notes,
		Total: len(notes),
	}, nil
}

func (t *NoteTool) notePath(id string) string {
	return filepath.Join(t.config.StorageDir, id+".json")
}

func (t *NoteTool) saveNote(note *Note) error {
	path := t.notePath(note.ID)
	data := fmt.Sprintf(`{"id":"%s","title":"%s","content":"%s","created_at":"%s","updated_at":"%s","tags":"%s"}`,
		note.ID,
		escapeJSON(note.Title),
		escapeJSON(note.Content),
		note.CreatedAt,
		note.UpdatedAt,
		escapeJSON(note.Tags),
	)
	return os.WriteFile(path, []byte(data), 0644)
}

func (t *NoteTool) loadNote(id string) (*Note, error) {
	path := t.notePath(id)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var note Note
	content := string(data)

	note.ID = extractField(content, "id")
	note.Title = extractField(content, "title")
	note.Content = extractField(content, "content")
	note.CreatedAt = extractField(content, "created_at")
	note.UpdatedAt = extractField(content, "updated_at")
	note.Tags = extractField(content, "tags")

	return &note, nil
}

func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return strings.ReplaceAll(s, "\n", "\\n")
}

func extractField(json, field string) string {
	pattern := `"` + field + `":"`
	start := strings.Index(json, pattern)
	if start == -1 {
		return ""
	}
	start += len(pattern)

	inString := false
	escape := false
	for i := start; i < len(json); i++ {
		c := json[i]
		if escape {
			escape = false
			continue
		}
		if c == '\\' {
			escape = true
			continue
		}
		if c == '"' {
			if !inString {
				inString = true
				continue
			}
			return json[start:i]
		}
	}
	return ""
}
