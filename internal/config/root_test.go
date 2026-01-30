package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindGasTownRootFrom(t *testing.T) {
	// Create a temporary Gas Town structure with mayor/town.json (town marker)
	tmpDir := t.TempDir()
	gasTownRoot := filepath.Join(tmpDir, "mytown")
	mayorDir := filepath.Join(gasTownRoot, "mayor")
	rigDir := filepath.Join(gasTownRoot, "smoke")
	rigMayorDir := filepath.Join(rigDir, "mayor") // rig has mayor/ but no town.json
	rigBeadsDir := filepath.Join(rigDir, ".beads")
	subDir := filepath.Join(rigDir, "crew", "ember")

	// Create directories
	if err := os.MkdirAll(mayorDir, 0755); err != nil {
		t.Fatalf("Failed to create mayor dir: %v", err)
	}
	// Create town.json marker file
	townJSON := filepath.Join(mayorDir, "town.json")
	if err := os.WriteFile(townJSON, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to create town.json: %v", err)
	}
	// Create rig mayor without town.json (shouldn't match)
	if err := os.MkdirAll(rigMayorDir, 0755); err != nil {
		t.Fatalf("Failed to create rig mayor dir: %v", err)
	}
	if err := os.MkdirAll(rigBeadsDir, 0755); err != nil {
		t.Fatalf("Failed to create rig .beads dir: %v", err)
	}
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create sub dir: %v", err)
	}

	tests := []struct {
		name    string
		start   string
		want    string
		wantErr bool
	}{
		{
			name:    "from root",
			start:   gasTownRoot,
			want:    gasTownRoot,
			wantErr: false,
		},
		{
			name:    "from rig (has mayor/ but no town.json)",
			start:   rigDir,
			want:    gasTownRoot,
			wantErr: false,
		},
		{
			name:    "from deep subdirectory",
			start:   subDir,
			want:    gasTownRoot,
			wantErr: false,
		},
		{
			name:    "from mayor directory",
			start:   mayorDir,
			want:    gasTownRoot,
			wantErr: false,
		},
		{
			name:    "not in gas town",
			start:   tmpDir,
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindGasTownRootFrom(tt.start)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindGasTownRootFrom() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FindGasTownRootFrom() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFeedPathWithOverride(t *testing.T) {
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

	t.Run("no override falls back to Gas Town", func(t *testing.T) {
		os.Setenv("SMOKE_FEED", "")

		// Create a temporary Gas Town
		tmpDir := t.TempDir()
		gasTownRoot := filepath.Join(tmpDir, "mytown")
		mayorDir := filepath.Join(gasTownRoot, "mayor")
		if err := os.MkdirAll(mayorDir, 0755); err != nil {
			t.Fatalf("Failed to create mayor dir: %v", err)
		}
		townJSON := filepath.Join(mayorDir, "town.json")
		if err := os.WriteFile(townJSON, []byte("{}"), 0644); err != nil {
			t.Fatalf("Failed to create town.json: %v", err)
		}

		// Change to Gas Town
		originalDir, _ := os.Getwd()
		defer func() { _ = os.Chdir(originalDir) }()
		if err := os.Chdir(gasTownRoot); err != nil {
			t.Fatalf("Failed to chdir: %v", err)
		}

		got, err := GetFeedPath()
		if err != nil {
			t.Errorf("GetFeedPath() unexpected error: %v", err)
		}
		// Resolve symlinks for comparison (macOS /var -> /private/var)
		gotResolved, _ := filepath.EvalSymlinks(got)
		wantResolved, _ := filepath.EvalSymlinks(filepath.Join(gasTownRoot, ".smoke", "feed.jsonl"))
		if gotResolved != wantResolved {
			t.Errorf("GetFeedPath() = %v, want %v", got, wantResolved)
		}
	})
}

func TestIsSmokeInitialized(t *testing.T) {
	// Create a temporary Gas Town structure with mayor/town.json marker
	tmpDir := t.TempDir()
	gasTownRoot := filepath.Join(tmpDir, "mytown")
	mayorDir := filepath.Join(gasTownRoot, "mayor")
	smokeDir := filepath.Join(gasTownRoot, ".smoke")

	if err := os.MkdirAll(mayorDir, 0755); err != nil {
		t.Fatalf("Failed to create mayor dir: %v", err)
	}
	// Create town.json marker file
	townJSON := filepath.Join(mayorDir, "town.json")
	if err := os.WriteFile(townJSON, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to create town.json: %v", err)
	}

	// Change to Gas Town directory
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()

	if err := os.Chdir(gasTownRoot); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}

	// Test before initialization
	initialized, err := IsSmokeInitialized()
	if err != nil {
		t.Errorf("IsSmokeInitialized() unexpected error: %v", err)
	}
	if initialized {
		t.Error("IsSmokeInitialized() = true, want false")
	}

	// Create smoke directory and feed file
	if mkdirErr := os.MkdirAll(smokeDir, 0755); mkdirErr != nil {
		t.Fatalf("Failed to create .smoke dir: %v", mkdirErr)
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
