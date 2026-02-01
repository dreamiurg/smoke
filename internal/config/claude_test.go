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
		result, err := AppendSmokeHint()
		if err != nil {
			t.Fatalf("AppendSmokeHint() error: %v", err)
		}
		if result == nil || !result.Appended {
			t.Error("AppendSmokeHint() Appended = false, want true (file created)")
		}
		// No backup should be created when file didn't exist
		if result.BackupPath != "" {
			t.Error("AppendSmokeHint() BackupPath should be empty when file didn't exist")
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
		result, err := AppendSmokeHint()
		if err != nil {
			t.Fatalf("AppendSmokeHint() error: %v", err)
		}
		if result == nil || result.Appended {
			t.Error("AppendSmokeHint() Appended = true, want false (already present)")
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
	result, err := AppendSmokeHint()
	if err != nil {
		t.Fatalf("AppendSmokeHint() error: %v", err)
	}
	if result == nil || !result.Appended {
		t.Error("AppendSmokeHint() Appended = false, want true")
	}

	// Backup should be created when modifying existing file
	if result.BackupPath == "" {
		t.Error("AppendSmokeHint() BackupPath should not be empty when modifying existing file")
	}
	// Verify backup file exists
	if _, statErr := os.Stat(result.BackupPath); os.IsNotExist(statErr) {
		t.Errorf("Backup file should exist at %s", result.BackupPath)
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

func TestIsClaudeCodeEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     bool
	}{
		{
			name:     "CLAUDECODE set to 1",
			envValue: "1",
			want:     true,
		},
		{
			name:     "CLAUDECODE empty",
			envValue: "",
			want:     false,
		},
		{
			name:     "CLAUDECODE set to 0",
			envValue: "0",
			want:     false,
		},
		{
			name:     "CLAUDECODE set to other value",
			envValue: "true",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("CLAUDECODE", tt.envValue)
			got := IsClaudeCodeEnvironment()
			if got != tt.want {
				t.Errorf("IsClaudeCodeEnvironment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetClaudeSettingsPath(t *testing.T) {
	got, err := GetClaudeSettingsPath()
	if err != nil {
		t.Fatalf("GetClaudeSettingsPath() error = %v", err)
	}

	// Verify path is not empty
	if got == "" {
		t.Error("GetClaudeSettingsPath() returned empty string")
		return
	}

	// Verify path ends with .claude/settings.json
	if !filepath.IsAbs(got) {
		t.Errorf("GetClaudeSettingsPath() = %q, want absolute path", got)
	}

	// Verify the path structure
	dir := filepath.Dir(got)
	base := filepath.Base(got)
	claudeDir := filepath.Base(dir)

	if base != "settings.json" {
		t.Errorf("GetClaudeSettingsPath() basename = %q, want 'settings.json'", base)
	}
	if claudeDir != ".claude" {
		t.Errorf("GetClaudeSettingsPath() parent dir = %q, want '.claude'", claudeDir)
	}

	// Verify it's in user's home directory
	home, err := os.UserHomeDir()
	if err == nil {
		want := filepath.Join(home, ".claude", "settings.json")
		if got != want {
			t.Errorf("GetClaudeSettingsPath() = %q, want %q", got, want)
		}
	}
}

func TestGetClaudeSettingsPath_NoHomeDir(t *testing.T) {
	// This test verifies the error handling path when UserHomeDir fails
	// In practice, we can't force UserHomeDir to fail reliably across platforms,
	// but we can at least verify the function doesn't panic and returns an error
	// when expected. In normal operation, this should succeed.
	got, err := GetClaudeSettingsPath()

	// In normal operation, should not return error
	if err != nil {
		// If there is an error, path should be empty
		if got != "" {
			t.Errorf("GetClaudeSettingsPath() returned path %q with error %v, want empty string", got, err)
		}
		return
	}

	// If no error, should return absolute path
	if got == "" {
		t.Error("GetClaudeSettingsPath() returned empty string with no error")
	}
	if !filepath.IsAbs(got) {
		t.Errorf("GetClaudeSettingsPath() returned non-absolute path: %q", got)
	}
}

func TestIsSmokeConfiguredInClaude(t *testing.T) {
	// Get testdata directory (relative to package root)
	testdataDir := filepath.Join("..", "..", "testdata", "claude")

	tests := []struct {
		name         string
		settingsFile string
		wantFound    bool
		wantErr      bool
	}{
		{
			name:         "valid settings with smoke",
			settingsFile: "valid-settings.json",
			wantFound:    true,
			wantErr:      false,
		},
		{
			name:         "valid settings without smoke",
			settingsFile: "missing-smoke.json",
			wantFound:    false,
			wantErr:      false,
		},
		{
			name:         "corrupted JSON",
			settingsFile: "corrupted.json",
			wantFound:    false,
			wantErr:      true,
		},
		{
			name:         "minimal settings (empty allow array)",
			settingsFile: "minimal-settings.json",
			wantFound:    false,
			wantErr:      false,
		},
		{
			name:         "non-existent file",
			settingsFile: "does-not-exist.json",
			wantFound:    false,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary home directory for this test
			tmpHome := t.TempDir()
			claudeDir := filepath.Join(tmpHome, ".claude")
			if err := os.MkdirAll(claudeDir, 0755); err != nil {
				t.Fatalf("Failed to create .claude dir: %v", err)
			}

			// Copy test fixture to temp settings.json (if it exists)
			if tt.settingsFile != "does-not-exist.json" {
				srcPath := filepath.Join(testdataDir, tt.settingsFile)
				dstPath := filepath.Join(claudeDir, "settings.json")

				data, err := os.ReadFile(srcPath)
				if err != nil {
					t.Fatalf("Failed to read test fixture: %v", err)
				}
				if err := os.WriteFile(dstPath, data, 0644); err != nil {
					t.Fatalf("Failed to write temp settings: %v", err)
				}
			}

			// Mock GetClaudeSettingsPath by setting HOME
			t.Setenv("HOME", tmpHome)
			t.Setenv("USERPROFILE", tmpHome) // For Windows compatibility

			got, err := IsSmokeConfiguredInClaude()
			if (err != nil) != tt.wantErr {
				t.Errorf("IsSmokeConfiguredInClaude() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantFound {
				t.Errorf("IsSmokeConfiguredInClaude() = %v, want %v", got, tt.wantFound)
			}
		})
	}
}

func TestIsSmokeConfiguredInClaude_CaseInsensitive(t *testing.T) {
	// Create a temporary settings file with uppercase SMOKE
	tmpHome := t.TempDir()
	claudeDir := filepath.Join(tmpHome, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude dir: %v", err)
	}

	settingsPath := filepath.Join(claudeDir, "settings.json")
	settingsContent := `{
  "permissions": {
    "allow": [
      "Bash(SMOKE:*)"
    ]
  }
}`
	if err := os.WriteFile(settingsPath, []byte(settingsContent), 0644); err != nil {
		t.Fatalf("Failed to write settings: %v", err)
	}

	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	got, err := IsSmokeConfiguredInClaude()
	if err != nil {
		t.Errorf("IsSmokeConfiguredInClaude() error = %v", err)
	}
	if !got {
		t.Error("IsSmokeConfiguredInClaude() should find 'SMOKE' (case-insensitive)")
	}
}

func TestIsSmokeConfiguredInClaude_PartialMatch(t *testing.T) {
	// Test that "smoke" is found in strings like "smoketest" or "smoke-test"
	tmpHome := t.TempDir()
	claudeDir := filepath.Join(tmpHome, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude dir: %v", err)
	}

	settingsPath := filepath.Join(claudeDir, "settings.json")
	settingsContent := `{
  "permissions": {
    "allow": [
      "Bash(smoketest:*)"
    ]
  }
}`
	if err := os.WriteFile(settingsPath, []byte(settingsContent), 0644); err != nil {
		t.Fatalf("Failed to write settings: %v", err)
	}

	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	got, err := IsSmokeConfiguredInClaude()
	if err != nil {
		t.Errorf("IsSmokeConfiguredInClaude() error = %v", err)
	}
	if !got {
		t.Error("IsSmokeConfiguredInClaude() should find 'smoke' as substring")
	}
}

func TestIsSmokeConfiguredInClaude_PathError(t *testing.T) {
	// Test behavior when GetClaudeSettingsPath returns an error
	// In practice, os.UserHomeDir() rarely fails in normal operation,
	// but IsSmokeConfiguredInClaude should propagate the error if it does.
	// We can't easily force UserHomeDir to fail without platform-specific tricks,
	// so this test documents the expected behavior: error from GetClaudeSettingsPath
	// should be returned by IsSmokeConfiguredInClaude.
	t.Skip("Cannot reliably test UserHomeDir failure without platform-specific mocking")
}
