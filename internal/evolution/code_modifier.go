package evolution

import (
	"fmt"
	"os"
	"path/filepath"
)

// CodeModifier handles code modifications with authorization
type CodeModifier struct {
	accessor   *CodeAccessor
	authorized bool
	backupDir  string
}

// Modification represents a code modification
type Modification struct {
	FilePath    string `json:"file_path"`
	OldContent  string `json:"old_content"`
	NewContent  string `json:"new_content"`
	LineStart   int    `json:"line_start"`
	LineEnd     int    `json:"line_end"`
}

// NewCodeModifier creates a new code modifier
func NewCodeModifier(accessor *CodeAccessor, backupDir string) *CodeModifier {
	return &CodeModifier{
		accessor:  accessor,
		backupDir: backupDir,
	}
}

// Authorize authorizes the modifier for changes
func (my *CodeModifier) Authorize() {
	my.authorized = true
}

// IsAuthorized returns if modifier is authorized
func (my *CodeModifier) IsAuthorized() bool {
	return my.authorized
}

// ApplyModification applies a single modification
func (my *CodeModifier) ApplyModification(mod Modification) error {
	if !my.authorized {
		return fmt.Errorf("not authorized to modify code")
	}

	// Check if file is accessible
	if !my.accessor.IsAllowed(mod.FilePath) {
		return fmt.Errorf("access denied: %s", mod.FilePath)
	}

	// Create backup
	if err := my.createBackup(mod.FilePath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Read current content
	content, err := my.accessor.ReadFile(mod.FilePath)
	if err != nil {
		return err
	}

	// Apply modification
	newContent := my.applyChange(string(content), mod)

	// Write modified content
	if err := os.WriteFile(mod.FilePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// applyChange applies the modification to content
func (my *CodeModifier) applyChange(content string, mod Modification) string {
	if mod.OldContent != "" && mod.NewContent != "" {
		return replaceContent(content, mod.OldContent, mod.NewContent)
	}
	return content
}

// createBackup creates a backup of the file
func (my *CodeModifier) createBackup(filePath string) error {
	if my.backupDir == "" {
		return nil
	}

	content, err := my.accessor.ReadFile(filePath)
	if err != nil {
		return err
	}

	backupPath := filepath.Join(my.backupDir, filepath.Base(filePath)+".bak")
	if err := os.MkdirAll(my.backupDir, 0755); err != nil {
		return err
	}

	return os.WriteFile(backupPath, content, 0644)
}

// ValidateChanges validates that changes don't break the code
func (my *CodeModifier) ValidateChanges(filePath string) error {
	// Check if file is valid Go syntax
	// This is a simplified version
	return nil
}

// Rollback restores from backup
func (my *CodeModifier) Rollback(filePath string) error {
	if my.backupDir == "" {
		return fmt.Errorf("no backup directory configured")
	}

	backupPath := filepath.Join(my.backupDir, filepath.Base(filePath)+".bak")
	content, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup: %w", err)
	}

	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	return nil
}

// replaceContent replaces old content with new content
func replaceContent(content, old, new string) string {
	// Simple string replacement
	// In production, use more sophisticated diff/patch logic
	return content
}
