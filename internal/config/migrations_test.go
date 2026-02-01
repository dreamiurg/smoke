package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T006: Test helpers

// setupTestConfig creates a temp config directory with a config.yaml file
// and overrides HOME to redirect config path. Returns cleanup function.
func setupTestConfig(t *testing.T, content string) func() {
	t.Helper()

	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".config", "smoke")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err, "Failed to create config directory")

	configPath := filepath.Join(configDir, "config.yaml")
	err = os.WriteFile(configPath, []byte(content), 0644)
	require.NoError(t, err, "Failed to write test config file")

	// Override HOME to redirect GetConfigPath()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)

	return func() {
		os.Setenv("HOME", oldHome)
	}
}

// T007: Test GetPendingMigrations returns empty when up to date

func TestGetPendingMigrations_UpToDate(t *testing.T) {
	content := `_schema_version: 1
pressure: 2
`
	cleanup := setupTestConfig(t, content)
	defer cleanup()

	pending, err := GetPendingMigrations()
	require.NoError(t, err)
	assert.Empty(t, pending, "Expected no pending migrations when schema is up to date")
}

// T008: Test GetPendingMigrations returns migrations when _schema_version missing

func TestGetPendingMigrations_NoSchemaVersion(t *testing.T) {
	content := `custom_field: value
`
	cleanup := setupTestConfig(t, content)
	defer cleanup()

	pending, err := GetPendingMigrations()
	require.NoError(t, err)
	assert.NotEmpty(t, pending, "Expected pending migrations when _schema_version is missing")

	// Should return all registered migrations
	assert.GreaterOrEqual(t, len(pending), 1, "Expected at least one migration")
}

// T009: Test ApplyMigrations adds missing fields

func TestApplyMigrations_AddsMissingFields(t *testing.T) {
	content := `_schema_version: 0
custom_field: my_value
`
	cleanup := setupTestConfig(t, content)
	defer cleanup()

	applied, err := ApplyMigrations(false)
	require.NoError(t, err)
	assert.NotEmpty(t, applied, "Expected at least one migration to be applied")

	// Verify the config was updated
	configMap, err := GetConfigAsMap()
	require.NoError(t, err)

	// Check that pressure field was added
	pressure, ok := configMap["pressure"]
	assert.True(t, ok, "Expected pressure field to be added")
	assert.Equal(t, 2, pressure, "Expected pressure to be set to default value 2")

	// Check that schema version was updated
	version, err := GetSchemaVersion(configMap)
	require.NoError(t, err)
	assert.Equal(t, CurrentSchemaVersion, version, "Expected schema version to be updated to current")
}

// T010: Test ApplyMigrations preserves existing values

func TestApplyMigrations_PreservesExistingValues(t *testing.T) {
	content := `_schema_version: 0
custom_field: my_value
existing_setting: 42
`
	cleanup := setupTestConfig(t, content)
	defer cleanup()

	applied, err := ApplyMigrations(false)
	require.NoError(t, err)
	assert.NotEmpty(t, applied, "Expected at least one migration to be applied")

	// Verify the config was updated
	configMap, err := GetConfigAsMap()
	require.NoError(t, err)

	// Check that existing values were preserved
	customField, ok := configMap["custom_field"]
	assert.True(t, ok, "Expected custom_field to be preserved")
	assert.Equal(t, "my_value", customField, "Expected custom_field value to be unchanged")

	existingSetting, ok := configMap["existing_setting"]
	assert.True(t, ok, "Expected existing_setting to be preserved")
	assert.Equal(t, 42, existingSetting, "Expected existing_setting value to be unchanged")
}

// Additional test: Verify dry run doesn't modify config

func TestApplyMigrations_DryRun(t *testing.T) {
	content := `_schema_version: 0
custom_field: my_value
`
	cleanup := setupTestConfig(t, content)
	defer cleanup()

	// Get config path to read original content
	configPath, err := GetConfigPath()
	require.NoError(t, err)

	// Get original content
	originalContent, err := os.ReadFile(configPath)
	require.NoError(t, err)

	applied, err := ApplyMigrations(true)
	require.NoError(t, err)
	assert.NotEmpty(t, applied, "Expected migrations to be detected in dry run")

	// Verify the config file wasn't modified
	currentContent, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, string(originalContent), string(currentContent), "Expected config to be unchanged in dry run")
}

// Additional test: Verify GetSchemaVersion handles missing field

func TestGetSchemaVersion_Missing(t *testing.T) {
	configMap := map[string]interface{}{
		"custom_field": "value",
	}

	version, err := GetSchemaVersion(configMap)
	require.NoError(t, err)
	assert.Equal(t, 0, version, "Expected version 0 when _schema_version is missing")
}

// Additional test: Verify GetSchemaVersion handles valid version

func TestGetSchemaVersion_Valid(t *testing.T) {
	configMap := map[string]interface{}{
		"_schema_version": 1,
		"custom_field":    "value",
	}

	version, err := GetSchemaVersion(configMap)
	require.NoError(t, err)
	assert.Equal(t, 1, version, "Expected version 1 from config")
}

// Additional test: Table-driven tests for multiple migration scenarios

func TestApplyMigrations_MultipleScenarios(t *testing.T) {
	tests := []struct {
		name           string
		configContent  string
		expectApplied  bool
		expectPressure bool
	}{
		{
			name:           "no migrations needed",
			configContent:  "_schema_version: 1\npressure: 2\n",
			expectApplied:  false,
			expectPressure: true,
		},
		{
			name:           "version 0 needs migration",
			configContent:  "_schema_version: 0\n",
			expectApplied:  true,
			expectPressure: true,
		},
		{
			name:           "missing version needs migration",
			configContent:  "custom: value\n",
			expectApplied:  true,
			expectPressure: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupTestConfig(t, tt.configContent)
			defer cleanup()

			applied, err := ApplyMigrations(false)
			require.NoError(t, err)

			if tt.expectApplied {
				assert.NotEmpty(t, applied, "Expected migrations to be applied")
			} else {
				assert.Empty(t, applied, "Expected no migrations to be applied")
			}

			if tt.expectPressure {
				configMap, err := GetConfigAsMap()
				require.NoError(t, err)
				_, hasPressure := configMap["pressure"]
				assert.True(t, hasPressure, "Expected pressure field to exist")
			}
		})
	}
}
