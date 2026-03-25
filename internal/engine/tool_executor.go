package engine

import (
	pkgtypes "github.com/lixianmin/pc/pkg/types"
)

type PluginManager interface {
	ListPlugins() []*pkgtypes.Plugin
	CallPlugin(plugin *pkgtypes.Plugin, method string, params any) (any, error)
}
