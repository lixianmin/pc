package engine

import (
	"context"

	"github.com/lixianmin/pc/pkg/types"
)

type BuiltinTool interface {
	Name() string
	Description() string
	Parameters() map[string]types.ParamSchema
	Execute(ctx context.Context, params map[string]any) (string, error)
}
