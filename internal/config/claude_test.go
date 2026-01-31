package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetClaudeMDPath(t *testing.T) {
	path, err := GetClaudeMDPath()
	if err != nil {
		t.Fatalf("GetClaudeMDPath() error: %v", err)
	}

	if filepath.Base(path) != "CLAUDE.md" {
		t.Errorf("GetClaudeMDPath() should end with CLAUDE.md, got %s", path)
	}

	if filepath.Base(filepath.Dir(path)) != ".claude" {
		t.Errorf("GetClaudeMDPath() parent should be .claude, got %s", filepath.Dir(path))
	}
}

func TestAppendSmokeHint(t *testing.T) {
	// Use temp directory as HOME
	tmpHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	t.Run("creates file if not exists", func(t *testing.T) {
		appended, err := AppendSmokeHint()
		if err != nil {
			t.Fatalf("AppendSmokeHint() error: %v", err)
		}
		if !appended {
			t.Error("AppendSmokeHint() = false, want true (file created)")
		}

		// Verify file exists
		claudePath := filepath.Join(tmpHome, ".claude", "CLAUDE.md")
		content, err := os.ReadFile(claudePath)
		if err != nil {
			t.Fatalf("Failed to read CLAUDE.md: %v", err)
		}
		if len(content) == 0 {
			t.Error("CLAUDE.md should not be empty")
		}
	})

	t.Run("idempotent - second call returns false", func(t *testing.T) {
		appended, err := AppendSmokeHint()
		if err != nil {
			t.Fatalf("AppendSmokeHint() error: %v", err)
		}
		if appended {
			t.Error("AppendSmokeHint() = true, want false (already present)")
		}
	})
}

func TestAppendSmokeHint_ExistingContent(t *testing.T) {
	// Use temp directory as HOME
	tmpHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	// Create existing CLAUDE.md with some content
	claudeDir := filepath.Join(tmpHome, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude dir: %v", err)
	}

	existingContent := "# My Claude Instructions\n\nSome existing content.\n"
	claudePath := filepath.Join(claudeDir, "CLAUDE.md")
	if err := os.WriteFile(claudePath, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing CLAUDE.md: %v", err)
	}

	// Append smoke hint
	appended, err := AppendSmokeHint()
	if err != nil {
		t.Fatalf("AppendSmokeHint() error: %v", err)
	}
	if !appended {
		t.Error("AppendSmokeHint() = false, want true")
	}

	// Verify content was appended (not replaced)
	content, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("Failed to read CLAUDE.md: %v", err)
	}

	contentStr := string(content)
	if len(contentStr) <= len(existingContent) {
		t.Error("Hint should have been appended to existing content")
	}
	if contentStr[:len(existingContent)] != existingContent {
		t.Error("Existing content should be preserved")
	}
}

func TestHasSmokeHint(t *testing.T) {
	// Use temp directory as HOME
	tmpHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	t.Run("returns false when file missing", func(t *testing.T) {
		has, err := HasSmokeHint()
		if err != nil {
			t.Fatalf("HasSmokeHint() error: %v", err)
		}
		if has {
			t.Error("HasSmokeHint() = true, want false (file missing)")
		}
	})

	t.Run("returns false when hint not present", func(t *testing.T) {
		claudeDir := filepath.Join(tmpHome, ".claude")
		os.MkdirAll(claudeDir, 0755)
		os.WriteFile(filepath.Join(claudeDir, "CLAUDE.md"), []byte("no hint here"), 0644)

		has, err := HasSmokeHint()
		if err != nil {
			t.Fatalf("HasSmokeHint() error: %v", err)
		}
		if has {
			t.Error("HasSmokeHint() = true, want false (no hint)")
		}
	})

	t.Run("returns true when hint present", func(t *testing.T) {
		claudePath := filepath.Join(tmpHome, ".claude", "CLAUDE.md")
		content := "some content\n" + SmokeHintMarker + "\nmore content"
		os.WriteFile(claudePath, []byte(content), 0644)

		has, err := HasSmokeHint()
		if err != nil {
			t.Fatalf("HasSmokeHint() error: %v", err)
		}
		if !has {
			t.Error("HasSmokeHint() = false, want true")
		}
	})
}
