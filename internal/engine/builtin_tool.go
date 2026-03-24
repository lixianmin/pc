package engine

import (
	"context"
)

type BuiltinTool interface {
	Name() string
	Execute(ctx context.Context, params map[string]any) (string, error)
}
