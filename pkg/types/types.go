package types

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
func (p PluginType) String() string {
	return string(p)
}

// IsValid returns true if the plugin type is valid.
func (p PluginType) IsValid() bool {
	switch p {
	case PluginTypeLLM, PluginTypeSearch, PluginTypeChannel, PluginTypeTool:
		return true
	default:
		return false
	}
}
