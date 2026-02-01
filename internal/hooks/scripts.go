package hooks

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
)

// GetScriptContent returns the embedded content of a hook script
func GetScriptContent(name string) ([]byte, error) {
	path := fmt.Sprintf("scripts/%s", name)
	return scripts.ReadFile(path)
}

// ListScripts returns all embedded hook scripts with their event mappings
func ListScripts() []HookScript {
	return []HookScript{
		{Name: "smoke-break.sh", Event: EventStop},
		{Name: "smoke-nudge.sh", Event: EventPostToolUse},
	}
}

// compareScriptHash checks if an installed script matches the embedded version
// Returns true if hashes match (or file doesn't exist), false if modified
func compareScriptHash(installedPath string, embeddedContent []byte) (bool, error) {
	// If file doesn't exist, consider it not modified (will be created)
	if _, err := os.Stat(installedPath); os.IsNotExist(err) {
		return true, nil
	}

	// Read installed file
	installedContent, err := os.ReadFile(installedPath)
	if err != nil {
		return false, fmt.Errorf("read installed script: %w", err)
	}

	// Compare SHA256 hashes
	installedHash := sha256.Sum256(installedContent)
	embeddedHash := sha256.Sum256(embeddedContent)

	return installedHash == embeddedHash, nil
}

// scriptExists checks if a script file exists at the given path
func scriptExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// isScriptModified checks if an installed script differs from embedded version
func isScriptModified(installedPath string, embeddedContent []byte) bool {
	if !scriptExists(installedPath) {
		return false // Can't be modified if it doesn't exist
	}

	matches, err := compareScriptHash(installedPath, embeddedContent)
	if err != nil {
		return false // If we can't read it, treat as not modified
	}

	return !matches
}

// getScriptStatus determines the status of a single script
func getScriptStatus(installedPath string, embeddedContent []byte) ScriptStatus {
	if !scriptExists(installedPath) {
		return StatusMissing
	}

	if isScriptModified(installedPath, embeddedContent) {
		return StatusModified
	}

	return StatusOK
}

// writeScript writes script content to disk with executable permissions
func writeScript(path string, content []byte) error {
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create hooks directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, content, 0755); err != nil {
		return fmt.Errorf("write script file: %w", err)
	}

	return nil
}
