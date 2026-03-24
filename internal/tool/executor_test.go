package tool

import (
	"context"
	"testing"
)

type mockTool struct {
	name string
}

func (m *mockTool) Name() string        { return m.name }
func (m *mockTool) Description() string { return "mock tool" }
func (m *mockTool) Execute(ctx context.Context, params map[string]any) (any, error) {
	return "mock result", nil
}

func TestExecutor_Register(t *testing.T) {
	e := NewExecutor()
	tool := &mockTool{name: "test"}

	e.Register(tool)

	if len(e.tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(e.tools))
	}
}

func TestExecutor_Execute(t *testing.T) {
	e := NewExecutor()
	tool := &mockTool{name: "test"}
	e.Register(tool)

	result, err := e.Execute(context.Background(), "test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "mock result" {
		t.Errorf("expected 'mock result', got %v", result)
	}
}

func TestExecutor_Execute_UnknownTool(t *testing.T) {
	e := NewExecutor()

	_, err := e.Execute(context.Background(), "unknown", nil)
	if err == nil {
		t.Fatal("expected error for unknown tool")
	}
}
