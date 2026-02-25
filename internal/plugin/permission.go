package plugin

import (
	"fmt"
	"strings"

	"github.com/lixianmin/pc/pkg/types"
)

// ParsePermissions parses permissions from a comma-separated string.
func ParsePermissions(permsStr string) ([]types.Permission, error) {
	if permsStr == "" {
		return []types.Permission{}, nil
	}

	permissionTypes := strings.Split(permsStr, ",")
	permissions := make([]types.Permission, 0, len(permissionTypes))

	for _, pType := range permissionTypes {
		pType = strings.TrimSpace(pType)
		if pType == "" {
			continue
		}

		permType := types.PermissionTypeFromString(pType)
		permissions = append(permissions, types.Permission{Type: permType})
	}

	return permissions, nil
}

// CheckPermission checks if the plugin has all required permissions.
func CheckPermission(pluginPerms []types.Permission, required []types.PermissionType) bool {
	if len(required) == 0 {
		return true
	}

	pluginPermSet := make(map[types.PermissionType]bool)
	for _, perm := range pluginPerms {
		pluginPermSet[perm.Type] = true
	}

	for _, req := range required {
		if !pluginPermSet[req] {
			return false
		}
	}

	return true
}

// FormatPermissionError formats a permission error message.
func FormatPermissionError(pluginPerms []types.Permission, required []types.PermissionType, pluginName string) error {
	missing := []types.PermissionType{}
	pluginPermSet := make(map[types.PermissionType]bool)
	for _, perm := range pluginPerms {
		pluginPermSet[perm.Type] = true
	}

	for _, req := range required {
		if !pluginPermSet[req] {
			missing = append(missing, req)
		}
	}

	if len(missing) == 0 {
		return nil
	}

	return fmt.Errorf("plugin %s missing required permissions: %v", pluginName, missing)
}
