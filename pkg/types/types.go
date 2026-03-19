package types

import "context"

type ParamSchema struct {
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Default     any    `json:"default,omitempty"`
}

type BuiltinTool interface {
	Name() string
	Description() string
	Parameters() map[string]ParamSchema
	Execute(ctx context.Context, params map[string]any) (string, error)
}

// PluginType represents the type of a plugin.
type PluginType string

const (
	// PluginTypeLLM is the LLM provider plugin type.
	PluginTypeLLM PluginType = "llm"
	// PluginTypeSearch is the web search plugin type.
	PluginTypeSearch PluginType = "search"
	// PluginTypeChannel is the communication channel plugin type.
	PluginTypeChannel PluginType = "channel"
	// PluginTypeTool is the tool plugin type.
	PluginTypeTool PluginType = "tool"
)

// PluginTypeFromString converts a string to PluginType.
// Returns PluginTypeTool for unknown types.
func PluginTypeFromString(s string) PluginType {
	switch s {
	case string(PluginTypeLLM):
		return PluginTypeLLM
	case string(PluginTypeSearch):
		return PluginTypeSearch
	case string(PluginTypeChannel):
		return PluginTypeChannel
	case string(PluginTypeTool):
		return PluginTypeTool
	default:
		return PluginTypeTool
	}
}

// String returns the string representation of PluginType.
func (my PluginType) String() string {
	return string(my)
}

// IsValid returns true if the plugin type is valid.
func (my PluginType) IsValid() bool {
	switch my {
	case PluginTypeLLM, PluginTypeSearch, PluginTypeChannel, PluginTypeTool:
		return true
	default:
		return false
	}
}

// Plugin represents a loaded plugin with its metadata.
type Plugin struct {
	Name        string       `yaml:"name"`    // Plugin name
	Type        PluginType   `yaml:"type"`    // Plugin type (llm, search, channel, tool)
	Enabled     bool         `yaml:"enabled"` // Whether the plugin is enabled
	Version     string       `yaml:"version"` // Plugin version
	Entry       string       `yaml:"entry"`   // Entry point executable path
	Path        string       // Full path to the plugin directory
	Permissions []Permission `yaml:"permissions"` // Declared permissions
}

// PermissionType represents the type of permission.
type PermissionType string

const (
	// PermissionNetwork allows network access.
	PermissionNetwork PermissionType = "network"
	// PermissionFilesystem allows filesystem access.
	PermissionFilesystem PermissionType = "filesystem"
	// PermissionSystem allows system-level access.
	PermissionSystem PermissionType = "system"
)

// Permission represents a plugin permission.
type Permission struct {
	Type PermissionType `yaml:"type"` // Permission type
}

// PermissionTypeFromString converts a string to PermissionType.
// Returns PermissionNetwork for unknown types.
func PermissionTypeFromString(s string) PermissionType {
	switch s {
	case string(PermissionNetwork):
		return PermissionNetwork
	case string(PermissionFilesystem):
		return PermissionFilesystem
	case string(PermissionSystem):
		return PermissionSystem
	default:
		return PermissionNetwork
	}
}

// String returns the string representation of PermissionType.
func (my PermissionType) String() string {
	return string(my)
}

// IsValid returns true if the permission type is valid.
func (my PermissionType) IsValid() bool {
	switch my {
	case PermissionNetwork, PermissionFilesystem, PermissionSystem:
		return true
	default:
		return false
	}
}
