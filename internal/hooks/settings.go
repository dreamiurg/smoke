package hooks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// readSettings reads and parses Claude Code settings.json
// Returns empty map if file doesn't exist
func readSettings() (map[string]interface{}, error) {
	settingsPath := GetSettingsPath()

	// If file doesn't exist, return empty settings
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		return make(map[string]interface{}), nil
	}

	// Read file
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return nil, fmt.Errorf("read settings file: %w", err)
	}

	// Parse JSON
	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidSettings, err)
	}

	return settings, nil
}

// writeSettings writes settings atomically to settings.json
func writeSettings(settings map[string]interface{}) error {
	settingsPath := GetSettingsPath()

	// Ensure parent directory exists
	dir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("%w: cannot create .claude directory", ErrPermissionDenied)
	}

	// Marshal with indentation for readability
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}

	// Write atomically (temp file + rename)
	tmpPath := settingsPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("%w: cannot write settings", ErrPermissionDenied)
	}

	// On Windows, os.Rename fails if destination exists. Remove first.
	if _, err := os.Stat(settingsPath); err == nil {
		if err := os.Remove(settingsPath); err != nil {
			_ = os.Remove(tmpPath) // Best effort cleanup
			return fmt.Errorf("%w: cannot replace settings", ErrPermissionDenied)
		}
	}
	if err := os.Rename(tmpPath, settingsPath); err != nil {
		_ = os.Remove(tmpPath) // Best effort cleanup
		return fmt.Errorf("%w: cannot update settings", ErrPermissionDenied)
	}

	return nil
}

// BackupSettings creates a timestamped backup of settings.json if it exists.
// Returns the backup path if created, empty string if no backup was needed (file doesn't exist).
func BackupSettings() (string, error) {
	settingsPath := GetSettingsPath()

	// Only backup if settings exists
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		return "", nil
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return "", err
	}

	// Create timestamped backup filename
	timestamp := time.Now().Format("2006-01-02T15-04-05")
	backupPath := fmt.Sprintf("%s.bak.%s", settingsPath, timestamp)

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return "", err
	}

	return backupPath, nil
}

// isSmokeHook checks if a hook command is a smoke hook
func isSmokeHook(command string) bool {
	return strings.Contains(command, "smoke-break.sh") ||
		strings.Contains(command, "smoke-nudge.sh")
}

// addHookToSettings adds or updates a smoke hook in settings
func addHookToSettings(settings map[string]interface{}, event HookEvent, scriptPath string) error {
	// Ensure hooks section exists
	hooks, ok := settings["hooks"].(map[string]interface{})
	if !ok {
		hooks = make(map[string]interface{})
		settings["hooks"] = hooks
	}

	// Get or create event array
	eventKey := string(event)
	eventArray, ok := hooks[eventKey].([]interface{})
	if !ok {
		eventArray = []interface{}{}
	}

	// Check if smoke hook already exists
	hookExists := false
	for _, entry := range eventArray {
		entryMap, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}

		entryHooks, ok := entryMap["hooks"].([]interface{})
		if !ok {
			continue
		}

		for _, hook := range entryHooks {
			hookMap, ok := hook.(map[string]interface{})
			if !ok {
				continue
			}

			if command, ok := hookMap["command"].(string); ok && isSmokeHook(command) {
				// Update existing hook
				hookMap["command"] = scriptPath
				hookExists = true
				break
			}
		}

		if hookExists {
			break
		}
	}

	// If hook doesn't exist, add new entry
	if !hookExists {
		newEntry := map[string]interface{}{
			"matcher": "",
			"hooks": []interface{}{
				map[string]interface{}{
					"type":    "command",
					"command": scriptPath,
				},
			},
		}
		eventArray = append(eventArray, newEntry)
	}

	hooks[eventKey] = eventArray
	return nil
}

// removeHookFromSettings removes smoke hooks from settings
func removeHookFromSettings(settings map[string]interface{}, event HookEvent) error {
	// Get hooks section
	hooks, ok := settings["hooks"].(map[string]interface{})
	if !ok {
		return nil // No hooks section, nothing to remove
	}

	// Get event array
	eventKey := string(event)
	eventArray, ok := hooks[eventKey].([]interface{})
	if !ok {
		return nil // No entries for this event
	}

	// Filter out smoke hooks
	var filtered []interface{}
	for _, entry := range eventArray {
		entryMap, ok := entry.(map[string]interface{})
		if !ok {
			filtered = append(filtered, entry)
			continue
		}

		entryHooks, ok := entryMap["hooks"].([]interface{})
		if !ok {
			filtered = append(filtered, entry)
			continue
		}

		// Filter hooks within entry
		var filteredHooks []interface{}
		for _, hook := range entryHooks {
			hookMap, ok := hook.(map[string]interface{})
			if !ok {
				filteredHooks = append(filteredHooks, hook)
				continue
			}

			command, ok := hookMap["command"].(string)
			if !ok || !isSmokeHook(command) {
				filteredHooks = append(filteredHooks, hook)
			}
		}

		// Keep entry if it still has non-smoke hooks
		if len(filteredHooks) > 0 {
			entryMap["hooks"] = filteredHooks
			filtered = append(filtered, entryMap)
		}
	}

	// Update or remove event array
	if len(filtered) > 0 {
		hooks[eventKey] = filtered
	} else {
		delete(hooks, eventKey)
	}

	return nil
}

// checkHookInSettings checks if a smoke hook is configured in settings
func checkHookInSettings(settings map[string]interface{}, event HookEvent) bool {
	hooks, ok := settings["hooks"].(map[string]interface{})
	if !ok {
		return false
	}

	eventKey := string(event)
	eventArray, ok := hooks[eventKey].([]interface{})
	if !ok {
		return false
	}

	// Check if any entry contains a smoke hook
	for _, entry := range eventArray {
		entryMap, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}

		entryHooks, ok := entryMap["hooks"].([]interface{})
		if !ok {
			continue
		}

		for _, hook := range entryHooks {
			hookMap, ok := hook.(map[string]interface{})
			if !ok {
				continue
			}

			if command, ok := hookMap["command"].(string); ok && isSmokeHook(command) {
				return true
			}
		}
	}

	return false
}
