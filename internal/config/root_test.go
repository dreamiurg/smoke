package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetConfigDir(t *testing.T) {
	configDir, err := GetConfigDir()
	if err != nil {
		t.Fatalf("GetConfigDir() error: %v", err)
	}

	if configDir == "" {
		t.Error("GetConfigDir() returned empty string")
	}

	// Should end with .config/smoke
	if filepath.Base(configDir) != "smoke" {
		t.Errorf("GetConfigDir() should end with 'smoke', got %s", configDir)
	}
}

func TestGetFeedPath(t *testing.T) {
	// Save and restore SMOKE_FEED env var
	origSmokeFeed := os.Getenv("SMOKE_FEED")
	defer os.Setenv("SMOKE_FEED", origSmokeFeed)

	t.Run("SMOKE_FEED override", func(t *testing.T) {
		customPath := "/some/custom/feed.jsonl"
		os.Setenv("SMOKE_FEED", customPath)

		got, err := GetFeedPath()
		if err != nil {
			t.Errorf("GetFeedPath() unexpected error: %v", err)
		}
		if got != customPath {
			t.Errorf("GetFeedPath() = %v, want %v", got, customPath)
		}
	})

	t.Run("no override uses global config", func(t *testing.T) {
		os.Setenv("SMOKE_FEED", "")

		got, err := GetFeedPath()
		if err != nil {
			t.Errorf("GetFeedPath() unexpected error: %v", err)
		}

		// Should be in ~/.config/smoke/
		if filepath.Base(got) != "feed.jsonl" {
			t.Errorf("GetFeedPath() should end with feed.jsonl, got %s", got)
		}
	})
}

func TestGetConfigPath(t *testing.T) {
	got, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() error: %v", err)
	}

	if filepath.Base(got) != "config.yaml" {
		t.Errorf("GetConfigPath() should end with config.yaml, got %s", got)
	}
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
	if err != nil {
		t.Errorf("IsSmokeInitialized() unexpected error: %v", err)
	}
	if initialized {
		t.Error("IsSmokeInitialized() = true, want false (no feed file yet)")
	}

	// Create smoke directory and feed file
	smokeDir := filepath.Join(tmpHome, ".config", "smoke")
	if mkdirErr := os.MkdirAll(smokeDir, 0755); mkdirErr != nil {
		t.Fatalf("Failed to create smoke dir: %v", mkdirErr)
	}
	feedPath := filepath.Join(smokeDir, "feed.jsonl")
	if writeErr := os.WriteFile(feedPath, []byte{}, 0644); writeErr != nil {
		t.Fatalf("Failed to create feed file: %v", writeErr)
	}

	// Test after initialization
	initialized, err = IsSmokeInitialized()
	if err != nil {
		t.Errorf("IsSmokeInitialized() unexpected error: %v", err)
	}
	if !initialized {
		t.Error("IsSmokeInitialized() = false, want true")
	}
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
	if err != ErrNotInitialized {
		t.Errorf("EnsureInitialized() = %v, want ErrNotInitialized", err)
	}

	// Initialize
	smokeDir := filepath.Join(tmpHome, ".config", "smoke")
	os.MkdirAll(smokeDir, 0755)
	os.WriteFile(filepath.Join(smokeDir, "feed.jsonl"), []byte{}, 0644)

	// Should return nil when initialized
	err = EnsureInitialized()
	if err != nil {
		t.Errorf("EnsureInitialized() = %v, want nil", err)
	}
}
