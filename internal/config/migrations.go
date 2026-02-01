// Package config provides configuration and initialization management for smoke.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// CurrentSchemaVersion is the latest config schema version.
// Increment this when adding new migrations.
const CurrentSchemaVersion = 1

// SchemaVersionKey is the field name used to track schema version in config.
const SchemaVersionKey = "_schema_version"

// Migration represents a single configuration transformation.
type Migration struct {
	Version        int                                       // Sequential migration number (1, 2, 3...)
	Name           string                                    // Short identifier (e.g., "add_pressure_setting")
	Description    string                                    // Human-readable explanation
	NeedsMigration func(config map[string]interface{}) bool  // Returns true if migration should be applied
	Apply          func(config map[string]interface{}) error // Applies the migration to config map
}

// migrations is the ordered list of all migrations.
// New migrations should be appended to this slice with incrementing Version numbers.
var migrations = []Migration{
	{
		Version:     1,
		Name:        "add_pressure_setting",
		Description: "Adds pressure field for suggest nudge frequency (default: 2)",
		NeedsMigration: func(config map[string]interface{}) bool {
			_, hasPressure := config["pressure"]
			return !hasPressure
		},
		Apply: func(config map[string]interface{}) error {
			config["pressure"] = DefaultPressure
			return nil
		},
	},
}

// GetConfigAsMap reads config.yaml and returns it as a map.
// Returns empty map if file doesn't exist (not an error).
// Returns error if file exists but contains invalid YAML.
func GetConfigAsMap() (map[string]interface{}, error) {
	path, err := GetConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]interface{}), nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	// Empty file is valid - return empty map
	if len(data) == 0 {
		return make(map[string]interface{}), nil
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("invalid YAML in config: %w", err)
	}

	// yaml.Unmarshal returns nil map for empty YAML document
	if config == nil {
		return make(map[string]interface{}), nil
	}

	return config, nil
}

// WriteConfigMap writes a config map to config.yaml using atomic write.
// Creates parent directory if needed. Uses 0600 permissions for security.
func WriteConfigMap(config map[string]interface{}) error {
	path, err := GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Atomic write: write to temp file, then rename
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write temp config: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		// Clean up temp file on rename failure
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to rename temp config: %w", err)
	}

	return nil
}

// GetSchemaVersion reads the _schema_version field from config.
// Returns 0 if field is missing (indicates all migrations need to run).
// Returns error if field exists but is not a valid integer.
func GetSchemaVersion(config map[string]interface{}) (int, error) {
	v, exists := config[SchemaVersionKey]
	if !exists {
		return 0, nil
	}

	// yaml.Unmarshal decodes integers as int
	switch version := v.(type) {
	case int:
		return version, nil
	case float64:
		// Handle case where YAML parsed as float
		return int(version), nil
	default:
		return 0, fmt.Errorf("_schema_version must be an integer, got %T", v)
	}
}

// ValidateSchemaVersion checks if the config's schema version is compatible.
// Returns an error if the config is from a future version of smoke.
func ValidateSchemaVersion(config map[string]interface{}) error {
	currentVersion, err := GetSchemaVersion(config)
	if err != nil {
		return err
	}

	if currentVersion > CurrentSchemaVersion {
		return fmt.Errorf("config has schema version %d but smoke only supports up to version %d (config may be from a newer smoke version)",
			currentVersion, CurrentSchemaVersion)
	}

	return nil
}

// GetPendingMigrations reads config and returns migrations that haven't been applied yet.
// Migrations are returned sorted by version ascending.
func GetPendingMigrations() ([]Migration, error) {
	config, err := GetConfigAsMap()
	if err != nil {
		return nil, err
	}
	return GetPendingMigrationsFromMap(config)
}

// GetPendingMigrationsFromMap returns migrations that haven't been applied to the given config map.
// Migrations are returned sorted by version ascending.
func GetPendingMigrationsFromMap(config map[string]interface{}) ([]Migration, error) {
	currentVersion, err := GetSchemaVersion(config)
	if err != nil {
		return nil, err
	}

	var pending []Migration
	for _, m := range migrations {
		if m.Version > currentVersion {
			pending = append(pending, m)
		}
	}

	return pending, nil
}

// ApplyMigrations applies all pending migrations to the config.
// If dryRun is true, returns what would be applied without modifying the config file.
// Returns the list of migrations that were (or would be) applied.
func ApplyMigrations(dryRun bool) ([]Migration, error) {
	config, err := GetConfigAsMap()
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	pending, err := GetPendingMigrationsFromMap(config)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending migrations: %w", err)
	}

	if len(pending) == 0 {
		return nil, nil
	}

	// Apply each migration in order
	for _, m := range pending {
		if m.NeedsMigration(config) {
			if err := m.Apply(config); err != nil {
				return nil, fmt.Errorf("migration %d (%s) failed: %w", m.Version, m.Name, err)
			}
		}
	}

	// Update schema version to the latest applied migration
	config[SchemaVersionKey] = pending[len(pending)-1].Version

	if dryRun {
		return pending, nil
	}

	// Write the updated config
	if err := WriteConfigMap(config); err != nil {
		return nil, fmt.Errorf("failed to write config: %w", err)
	}

	return pending, nil
}

// GetAllMigrations returns all registered migrations.
// Useful for doctor command to show migration status.
func GetAllMigrations() []Migration {
	result := make([]Migration, len(migrations))
	copy(result, migrations)
	return result
}
