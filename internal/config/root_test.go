package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindGasTownRootFrom(t *testing.T) {
	// Create a temporary Gas Town structure
	tmpDir := t.TempDir()
	gasTownRoot := filepath.Join(tmpDir, "mytown")
	beadsDir := filepath.Join(gasTownRoot, ".beads")
	subDir := filepath.Join(gasTownRoot, "crew", "ember")

	// Create directories
	if err := os.MkdirAll(beadsDir, 0755); err != nil {
		t.Fatalf("Failed to create .beads dir: %v", err)
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
			name:    "from subdirectory",
			start:   subDir,
			want:    gasTownRoot,
			wantErr: false,
		},
		{
			name:    "from beads directory",
			start:   beadsDir,
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

func TestFindGasTownRootFromWithMayor(t *testing.T) {
	// Create a Gas Town with mayor marker
	tmpDir := t.TempDir()
	gasTownRoot := filepath.Join(tmpDir, "mytown")
	mayorDir := filepath.Join(gasTownRoot, "mayor")

	if err := os.MkdirAll(mayorDir, 0755); err != nil {
		t.Fatalf("Failed to create mayor dir: %v", err)
	}

	got, err := FindGasTownRootFrom(gasTownRoot)
	if err != nil {
		t.Errorf("FindGasTownRootFrom() unexpected error: %v", err)
	}
	if got != gasTownRoot {
		t.Errorf("FindGasTownRootFrom() = %v, want %v", got, gasTownRoot)
	}
}

func TestIsSmokeInitialized(t *testing.T) {
	// Create a temporary Gas Town structure
	tmpDir := t.TempDir()
	gasTownRoot := filepath.Join(tmpDir, "mytown")
	beadsDir := filepath.Join(gasTownRoot, ".beads")
	smokeDir := filepath.Join(gasTownRoot, ".smoke")

	if err := os.MkdirAll(beadsDir, 0755); err != nil {
		t.Fatalf("Failed to create .beads dir: %v", err)
	}

	// Change to Gas Town directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

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
	if err := os.MkdirAll(smokeDir, 0755); err != nil {
		t.Fatalf("Failed to create .smoke dir: %v", err)
	}
	feedPath := filepath.Join(smokeDir, "feed.jsonl")
	if err := os.WriteFile(feedPath, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create feed file: %v", err)
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
