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
		// Use temp dir as HOME to test path validation
		tmpHome := t.TempDir()
		oldHome := os.Getenv("HOME")
		os.Setenv("HOME", tmpHome)
		defer os.Setenv("HOME", oldHome)

		customPath := filepath.Join(tmpHome, "custom", "feed.jsonl")
		os.Setenv("SMOKE_FEED", customPath)

		got, err := GetFeedPath()
		assert.NoError(t, err)
		assert.Equal(t, customPath, got)
	})

	t.Run("SMOKE_FEED outside home rejected", func(t *testing.T) {
		customPath := "/some/custom/feed.jsonl"
		os.Setenv("SMOKE_FEED", customPath)

		_, err := GetFeedPath()
		assert.ErrorIs(t, err, ErrInvalidFeedPath)
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

func TestValidateFeedPath(t *testing.T) {
	// Save and restore env vars
	origSmokeFeed := os.Getenv("SMOKE_FEED")
	defer os.Setenv("SMOKE_FEED", origSmokeFeed)

	t.Run("temp directory allowed", func(t *testing.T) {
		tmpPath := filepath.Join(os.TempDir(), "test-feed.jsonl")
		os.Setenv("SMOKE_FEED", tmpPath)

		got, err := GetFeedPath()
		assert.NoError(t, err)
		assert.Contains(t, got, "test-feed.jsonl")
	})

	t.Run("private tmp allowed on macOS", func(t *testing.T) {
		// /private/tmp is the resolved path of /tmp on macOS
		tmpPath := "/private/tmp/test-feed.jsonl"
		os.Setenv("SMOKE_FEED", tmpPath)

		got, err := GetFeedPath()
		assert.NoError(t, err)
		assert.Equal(t, tmpPath, got)
	})

	t.Run("var folders allowed", func(t *testing.T) {
		// /var/folders is used by macOS for temp files
		tmpPath := "/var/folders/xx/test/feed.jsonl"
		os.Setenv("SMOKE_FEED", tmpPath)

		got, err := GetFeedPath()
		assert.NoError(t, err)
		assert.Equal(t, tmpPath, got)
	})

	t.Run("private var folders allowed", func(t *testing.T) {
		// /private/var/folders is the resolved path on macOS
		tmpPath := "/private/var/folders/xx/test/feed.jsonl"
		os.Setenv("SMOKE_FEED", tmpPath)

		got, err := GetFeedPath()
		assert.NoError(t, err)
		assert.Equal(t, tmpPath, got)
	})

	t.Run("TMPDIR path allowed", func(t *testing.T) {
		tmpDir := os.TempDir()
		tmpPath := filepath.Join(tmpDir, "custom", "feed.jsonl")
		os.Setenv("SMOKE_FEED", tmpPath)

		got, err := GetFeedPath()
		assert.NoError(t, err)
		assert.Equal(t, tmpPath, got)
	})

	t.Run("absolute system path rejected", func(t *testing.T) {
		os.Setenv("SMOKE_FEED", "/etc/passwd")

		_, err := GetFeedPath()
		assert.ErrorIs(t, err, ErrInvalidFeedPath)
	})

	t.Run("usr path rejected", func(t *testing.T) {
		os.Setenv("SMOKE_FEED", "/usr/local/feed.jsonl")

		_, err := GetFeedPath()
		assert.ErrorIs(t, err, ErrInvalidFeedPath)
	})
}
