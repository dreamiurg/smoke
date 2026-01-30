package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetIdentity(t *testing.T) {
	// Save original env
	origBDActor := os.Getenv("BD_ACTOR")
	origSmokeAuthor := os.Getenv("SMOKE_AUTHOR")
	defer func() {
		os.Setenv("BD_ACTOR", origBDActor)
		os.Setenv("SMOKE_AUTHOR", origSmokeAuthor)
	}()

	tests := []struct {
		name        string
		bdActor     string
		smokeAuthor string
		wantAuthor  string
		wantRig     string
		wantErr     bool
	}{
		{
			name:       "BD_ACTOR rig/role/name format",
			bdActor:    "smoke/crew/ember",
			wantAuthor: "ember",
			wantRig:    "smoke",
			wantErr:    false,
		},
		{
			name:       "BD_ACTOR rig/name format",
			bdActor:    "calle/witness",
			wantAuthor: "witness",
			wantRig:    "calle",
			wantErr:    false,
		},
		{
			name:       "BD_ACTOR name only",
			bdActor:    "refinery",
			wantAuthor: "refinery",
			wantErr:    false,
		},
		{
			name:        "SMOKE_AUTHOR fallback",
			bdActor:     "",
			smokeAuthor: "testuser",
			wantAuthor:  "testuser",
			wantErr:     false,
		},
		{
			name:    "no identity",
			bdActor: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("BD_ACTOR", tt.bdActor)
			os.Setenv("SMOKE_AUTHOR", tt.smokeAuthor)

			got, err := GetIdentity()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetIdentity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if got.Author != tt.wantAuthor {
				t.Errorf("GetIdentity().Author = %v, want %v", got.Author, tt.wantAuthor)
			}
			if tt.wantRig != "" && got.Rig != tt.wantRig {
				t.Errorf("GetIdentity().Rig = %v, want %v", got.Rig, tt.wantRig)
			}
		})
	}
}

func TestGetIdentityWithOverrides(t *testing.T) {
	// Save original env
	origBDActor := os.Getenv("BD_ACTOR")
	defer os.Setenv("BD_ACTOR", origBDActor)

	tests := []struct {
		name           string
		bdActor        string
		authorOverride string
		rigOverride    string
		wantAuthor     string
		wantRig        string
		wantErr        bool
	}{
		{
			name:           "override author",
			bdActor:        "smoke/crew/ember",
			authorOverride: "custom",
			wantAuthor:     "custom",
			wantRig:        "smoke",
			wantErr:        false,
		},
		{
			name:        "override rig",
			bdActor:     "smoke/crew/ember",
			rigOverride: "custom-rig",
			wantAuthor:  "ember",
			wantRig:     "custom-rig",
			wantErr:     false,
		},
		{
			name:           "override both",
			bdActor:        "smoke/crew/ember",
			authorOverride: "custom-author",
			rigOverride:    "custom-rig",
			wantAuthor:     "custom-author",
			wantRig:        "custom-rig",
			wantErr:        false,
		},
		{
			name:           "override with no env",
			bdActor:        "",
			authorOverride: "manual",
			wantAuthor:     "manual",
			wantErr:        false,
		},
		{
			name:    "no identity no override",
			bdActor: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("BD_ACTOR", tt.bdActor)

			got, err := GetIdentityWithOverrides(tt.authorOverride, tt.rigOverride)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetIdentityWithOverrides() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if got.Author != tt.wantAuthor {
				t.Errorf("GetIdentityWithOverrides().Author = %v, want %v", got.Author, tt.wantAuthor)
			}
			if tt.wantRig != "" && got.Rig != tt.wantRig {
				t.Errorf("GetIdentityWithOverrides().Rig = %v, want %v", got.Rig, tt.wantRig)
			}
		})
	}
}

func TestIdentityNameSanitization(t *testing.T) {
	// Save original env
	origBDActor := os.Getenv("BD_ACTOR")
	defer os.Setenv("BD_ACTOR", origBDActor)

	tests := []struct {
		name       string
		bdActor    string
		wantAuthor string
		wantRig    string
	}{
		{
			name:       "mixed case",
			bdActor:    "Smoke/Crew/Ember",
			wantAuthor: "ember",
			wantRig:    "smoke",
		},
		{
			name:       "with spaces",
			bdActor:    "my rig/my name",
			wantAuthor: "my-name",
			wantRig:    "my-rig",
		},
		{
			name:       "with special chars",
			bdActor:    "rig@test/user!name",
			wantAuthor: "username",
			wantRig:    "rigtest",
		},
		{
			name:       "whitespace trim",
			bdActor:    "  smoke  /  ember  ",
			wantAuthor: "ember",
			wantRig:    "smoke",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("BD_ACTOR", tt.bdActor)

			got, err := GetIdentity()
			if err != nil {
				t.Errorf("GetIdentity() unexpected error: %v", err)
				return
			}
			if got.Author != tt.wantAuthor {
				t.Errorf("GetIdentity().Author = %v, want %v", got.Author, tt.wantAuthor)
			}
			if got.Rig != tt.wantRig {
				t.Errorf("GetIdentity().Rig = %v, want %v", got.Rig, tt.wantRig)
			}
		})
	}
}

func TestInferRig(t *testing.T) {
	// Create a temporary Gas Town structure
	tmpDir := t.TempDir()
	gasTownRoot := filepath.Join(tmpDir, "my-town")
	beadsDir := filepath.Join(gasTownRoot, ".beads")

	if err := os.MkdirAll(beadsDir, 0755); err != nil {
		t.Fatalf("Failed to create .beads dir: %v", err)
	}

	// Change to Gas Town directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	if err := os.Chdir(gasTownRoot); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}

	// Save original env
	origBDActor := os.Getenv("BD_ACTOR")
	origSmokeAuthor := os.Getenv("SMOKE_AUTHOR")
	defer func() {
		os.Setenv("BD_ACTOR", origBDActor)
		os.Setenv("SMOKE_AUTHOR", origSmokeAuthor)
	}()

	// Test with SMOKE_AUTHOR which should infer rig
	os.Setenv("BD_ACTOR", "")
	os.Setenv("SMOKE_AUTHOR", "testuser")

	got, err := GetIdentity()
	if err != nil {
		t.Errorf("GetIdentity() unexpected error: %v", err)
		return
	}
	if got.Rig != "my-town" {
		t.Errorf("GetIdentity().Rig = %v, want my-town", got.Rig)
	}
}
