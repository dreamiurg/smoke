package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConfigDir(t *testing.T) {
	configDir, err := GetConfigDir()
	require.NoError(t, err)

	assert.NotEmpty(t, configDir)

	// Should end with .config/smoke
	assert.Equal(t, "smoke", filepath.Base(configDir))
}

func TestGetFeedPath(t *testing.T) {
	// Save and restore SMOKE_FEED env var
	origSmokeFeed := os.Getenv("SMOKE_FEED")
	defer os.Setenv("SMOKE_FEED", origSmokeFeed)

	t.Run("SMOKE_FEED override", func(t *testing.T) {
		customPath := "/some/custom/feed.jsonl"
		os.Setenv("SMOKE_FEED", customPath)

		got, err := GetFeedPath()
		assert.NoError(t, err)
		assert.Equal(t, customPath, got)
	})

	t.Run("no override uses global config", func(t *testing.T) {
		os.Setenv("SMOKE_FEED", "")

		got, err := GetFeedPath()
		assert.NoError(t, err)

		// Should be in ~/.config/smoke/
		assert.Equal(t, "feed.jsonl", filepath.Base(got))
	})
}

func TestGetConfigPath(t *testing.T) {
	got, err := GetConfigPath()
	require.NoError(t, err)

	assert.Equal(t, "config.yaml", filepath.Base(got))
}

func TestIsSmokeInitialized(t *testing.T) {
	// Use temp directory as HOME
	tmpHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	// Clear any SMOKE_FEED override
	oldSmokeFeed := os.Getenv("SMOKE_FEED")
	os.Setenv("SMOKE_FEED", "")
	defer os.Setenv("SMOKE_FEED", oldSmokeFeed)

	// Test before initialization
	initialized, err := IsSmokeInitialized()
	assert.NoError(t, err)
	assert.False(t, initialized)

	// Create smoke directory and feed file
	smokeDir := filepath.Join(tmpHome, ".config", "smoke")
	mkdirErr := os.MkdirAll(smokeDir, 0755)
	require.NoError(t, mkdirErr)
	feedPath := filepath.Join(smokeDir, "feed.jsonl")
	writeErr := os.WriteFile(feedPath, []byte{}, 0644)
	require.NoError(t, writeErr)

	// Test after initialization
	initialized, err = IsSmokeInitialized()
	assert.NoError(t, err)
	assert.True(t, initialized)
}

func TestEnsureInitialized(t *testing.T) {
	// Use temp directory as HOME
	tmpHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	// Clear any SMOKE_FEED override
	oldSmokeFeed := os.Getenv("SMOKE_FEED")
	os.Setenv("SMOKE_FEED", "")
	defer os.Setenv("SMOKE_FEED", oldSmokeFeed)

	// Should return error when not initialized
	err := EnsureInitialized()
	assert.Equal(t, ErrNotInitialized, err)

	// Initialize
	smokeDir := filepath.Join(tmpHome, ".config", "smoke")
	os.MkdirAll(smokeDir, 0755)
	os.WriteFile(filepath.Join(smokeDir, "feed.jsonl"), []byte{}, 0644)

	// Should return nil when initialized
	err = EnsureInitialized()
	assert.NoError(t, err)
}
