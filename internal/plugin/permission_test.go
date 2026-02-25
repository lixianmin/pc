package plugin

import (
	"testing"

	"github.com/lixianmin/pc/pkg/types"
)

func TestParsePermissions(t *testing.T) {
	tests := []struct {
		name         string
		permissions  string
		wantCount    int
		wantTypes    []types.PermissionType
	}{
		{
			name:        "parse single permission",
			permissions: "network",
			wantCount:   1,
			wantTypes:   []types.PermissionType{types.PermissionNetwork},
		},
		{
			name:        "parse multiple permissions",
			permissions: "network,filesystem",
			wantCount:   2,
			wantTypes:   []types.PermissionType{types.PermissionNetwork, types.PermissionFilesystem},
		},
		{
			name:        "parse permissions with duplicates",
			permissions: "network,network",
			wantCount:   2,
			wantTypes:   []types.PermissionType{types.PermissionNetwork, types.PermissionNetwork},
		},
		{
			name:        "parse invalid permission type",
			permissions: "invalid",
			wantCount:   1,
			wantTypes:   []types.PermissionType{types.PermissionNetwork}, // default
		},
		{
			name:        "parse empty permissions",
			permissions: "",
			wantCount:   0,
			wantTypes:   []types.PermissionType{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perms, err := ParsePermissions(tt.permissions)

			if (err != nil) != false {
				t.Errorf("ParsePermissions() unexpected error: %v", err)
			}

			if len(perms) != tt.wantCount {
				t.Errorf("ParsePermissions() count = %d, want %d", len(perms), tt.wantCount)
			}

			// Check permission types
			for i, p := range perms {
				if i < len(tt.wantTypes) {
					if p.Type != tt.wantTypes[i] {
						t.Errorf("ParsePermissions() type[%d] = %v, want %v", i, p.Type, tt.wantTypes[i])
					}
				}
			}
		})
	}
}

func TestCheckPermission(t *testing.T) {
	tests := []struct {
		name        string
		pluginPerms []types.Permission
		required    []types.PermissionType
		result      bool
	}{
		{
			name:        "all permissions granted",
			pluginPerms: []types.Permission{
				{Type: types.PermissionNetwork},
				{Type: types.PermissionFilesystem},
				{Type: types.PermissionSystem},
			},
			required: []types.PermissionType{types.PermissionNetwork, types.PermissionFilesystem},
			result:      true,
		},
		{
			name:        "missing required permission",
			pluginPerms: []types.Permission{
				{Type: types.PermissionNetwork},
			},
			required: []types.PermissionType{types.PermissionNetwork, types.PermissionFilesystem},
			result:      false,
		},
		{
			name:        "extra permission granted",
			pluginPerms: []types.Permission{
				{Type: types.PermissionNetwork},
				{Type: types.PermissionFilesystem},
				{Type: types.PermissionSystem},
			},
			required: []types.PermissionType{types.PermissionNetwork},
			result:      true,
		},
		{
			name:        "no permissions required",
			pluginPerms: []types.Permission{},
			required:    []types.PermissionType{},
			result:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckPermission(tt.pluginPerms, tt.required)

			if result != tt.result {
				t.Errorf("CheckPermission() = %v, want %v", result, tt.result)
			}
		})
	}
}

func TestFormatPermissionError(t *testing.T) {
	tests := []struct {
		name          string
		permissions   []types.Permission
		required      []types.PermissionType
		pluginName    string
	}{
		{
			name:     "missing network permission",
			permissions: []types.Permission{{Type: types.PermissionFilesystem}},
			required:  []types.PermissionType{types.PermissionNetwork},
			pluginName: "test-plugin",
		},
		{
			name:     "multiple missing permissions",
			permissions: []types.Permission{{Type: types.PermissionFilesystem}},
			required: []types.PermissionType{types.PermissionNetwork, types.PermissionSystem},
			pluginName: "test-plugin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := FormatPermissionError(tt.permissions, tt.required, tt.pluginName)

			if err == nil {
				t.Error("FormatPermissionError() should return error for missing permissions")
			}
		})
	}
}
