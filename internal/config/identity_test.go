package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetIdentity_WithSmokeAuthor(t *testing.T) {
	// Save original env
	origSmokeAuthor := os.Getenv("SMOKE_AUTHOR")
	origSessionID := os.Getenv("TERM_SESSION_ID")
	defer func() {
		os.Setenv("SMOKE_AUTHOR", origSmokeAuthor)
		os.Setenv("TERM_SESSION_ID", origSessionID)
	}()

	// Set env var
	os.Setenv("SMOKE_AUTHOR", "test-user")
	os.Setenv("TERM_SESSION_ID", "")

	identity, err := GetIdentity()
	if err != nil {
		t.Fatalf("GetIdentity failed: %v", err)
	}

	if identity.Suffix != "test-user" {
		t.Errorf("Expected suffix 'test-user', got %q", identity.Suffix)
	}
	if identity.Agent != "custom" {
		t.Errorf("Expected agent 'custom', got %q", identity.Agent)
	}
}

func TestGetIdentityWithOverride(t *testing.T) {
	// Save original env
	origSessionID := os.Getenv("TERM_SESSION_ID")
	defer os.Setenv("TERM_SESSION_ID", origSessionID)

	// Ensure we have a session seed so auto-detection doesn't fail
	os.Setenv("TERM_SESSION_ID", "test-session-123")

	identity, err := GetIdentityWithOverride("my-custom-name")
	if err != nil {
		t.Fatalf("GetIdentityWithOverride failed: %v", err)
	}

	if identity.Suffix != "my-custom-name" {
		t.Errorf("Expected suffix 'my-custom-name', got %q", identity.Suffix)
	}
	if identity.Agent != "custom" {
		t.Errorf("Expected agent 'custom', got %q", identity.Agent)
	}
}

func TestParseFullIdentity(t *testing.T) {
	tests := []struct {
		input   string
		agent   string
		suffix  string
		project string
	}{
		{"claude-swift-fox@smoke", "claude", "swift-fox", "smoke"},
		{"unknown-calm-owl@myproject", "unknown", "calm-owl", "myproject"},
		{"custom@test", "custom", "unknown", "test"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			id, err := parseFullIdentity(tt.input)
			if err != nil {
				t.Fatalf("parseFullIdentity failed: %v", err)
			}
			if id.Agent != tt.agent {
				t.Errorf("Agent: got %q, want %q", id.Agent, tt.agent)
			}
			if id.Suffix != tt.suffix {
				t.Errorf("Suffix: got %q, want %q", id.Suffix, tt.suffix)
			}
			if id.Project != tt.project {
				t.Errorf("Project: got %q, want %q", id.Project, tt.project)
			}
		})
	}
}

func TestDetectProject(t *testing.T) {
	project := detectProject()
	if project == "" {
		t.Error("detectProject returned empty string")
	}
	// Should return something (either git repo name or cwd)
	t.Logf("Detected project: %s", project)
}

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Hello World", "hello-world"},
		{"Test_Name", "test_name"},
		{"with spaces", "with-spaces"},
		{"UPPERCASE", "uppercase"},
		{"special!@#chars", "specialchars"},
		{"  trim  ", "trim"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := sanitizeName(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIdentityString(t *testing.T) {
	id := &Identity{
		Agent:   "claude",
		Suffix:  "swift-fox",
		Project: "smoke",
	}

	want := "claude-swift-fox@smoke"
	got := id.String()

	if got != want {
		t.Errorf("Identity.String() = %q, want %q", got, want)
	}
}

func TestGetIdentity_AutoDetect(t *testing.T) {
	// Save original env
	origSmokeAuthor := os.Getenv("SMOKE_AUTHOR")
	origSessionID := os.Getenv("TERM_SESSION_ID")
	defer func() {
		os.Setenv("SMOKE_AUTHOR", origSmokeAuthor)
		os.Setenv("TERM_SESSION_ID", origSessionID)
	}()

	// Clear explicit author to force auto-detection
	os.Setenv("SMOKE_AUTHOR", "")
	os.Setenv("TERM_SESSION_ID", "test-auto-detect-123")

	identity, err := GetIdentity()
	if err != nil {
		t.Fatalf("GetIdentity failed: %v", err)
	}

	// Should have a non-empty suffix from auto-generated name
	if identity.Suffix == "" {
		t.Error("Expected non-empty suffix from auto-detection")
	}
	// Should have detected agent
	if identity.Agent == "" {
		t.Error("Expected non-empty agent")
	}
	// Should have detected project
	if identity.Project == "" {
		t.Error("Expected non-empty project")
	}

	t.Logf("Auto-detected identity: %s", identity.String())
}

func TestGetIdentity_FallsBackToSessionSeed(t *testing.T) {
	// Save original env
	origSmokeAuthor := os.Getenv("SMOKE_AUTHOR")
	origSessionID := os.Getenv("TERM_SESSION_ID")
	defer func() {
		os.Setenv("SMOKE_AUTHOR", origSmokeAuthor)
		os.Setenv("TERM_SESSION_ID", origSessionID)
	}()

	// Clear explicit author and primary session ID to force PPID fallback
	os.Setenv("SMOKE_AUTHOR", "")
	os.Setenv("TERM_SESSION_ID", "")

	identity, err := GetIdentity()
	if err != nil {
		t.Fatalf("GetIdentity should not fail with PPID fallback: %v", err)
	}

	// Should have generated an identity using PPID-based seed
	if identity.Suffix == "" {
		t.Error("Expected non-empty suffix even with PPID fallback")
	}
	if identity.Agent != "claude" && identity.Agent != "unknown" {
		t.Errorf("Expected agent to be either 'claude' or 'unknown', got %q", identity.Agent)
	}
	t.Logf("Identity with PPID fallback: %s", identity.String())
}

func TestParseFullIdentity_AllComponents(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		wantID  *Identity
	}{
		{
			name:  "valid format",
			input: "agent-suffix@project",
			wantID: &Identity{
				Agent:   "agent",
				Suffix:  "suffix",
				Project: "project",
			},
		},
		{
			name:  "no agent dash",
			input: "agent@project",
			wantID: &Identity{
				Agent:   "agent",
				Suffix:  "unknown",
				Project: "project",
			},
		},
		{
			name:    "no project separator",
			input:   "agent-suffix",
			wantErr: true,
		},
		{
			name:  "case insensitive",
			input: "Agent-Suffix@Project",
			wantID: &Identity{
				Agent:   "agent",
				Suffix:  "suffix",
				Project: "project",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := parseFullIdentity(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFullIdentity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if id.Agent != tt.wantID.Agent {
				t.Errorf("Agent: got %q, want %q", id.Agent, tt.wantID.Agent)
			}
			if id.Suffix != tt.wantID.Suffix {
				t.Errorf("Suffix: got %q, want %q", id.Suffix, tt.wantID.Suffix)
			}
			if id.Project != tt.wantID.Project {
				t.Errorf("Project: got %q, want %q", id.Project, tt.wantID.Project)
			}
		})
	}
}

func TestDetectProject_InGitRepo(t *testing.T) {
	// Save original directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get cwd: %v", err)
	}
	defer os.Chdir(originalDir)

	// Create a temporary directory to simulate a git repo
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, "test-repo")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create a .git directory marker
	dotGit := filepath.Join(gitDir, ".git")
	if err := os.MkdirAll(dotGit, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	// Change to the git repo directory
	if err := os.Chdir(gitDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}

	project := detectProject()
	if project != "test-repo" {
		t.Errorf("Expected 'test-repo', got %q", project)
	}
}

func TestGetIdentityWithOverride_FullIdentity(t *testing.T) {
	// Save original env
	origSessionID := os.Getenv("TERM_SESSION_ID")
	defer os.Setenv("TERM_SESSION_ID", origSessionID)

	// Set session for fallback
	os.Setenv("TERM_SESSION_ID", "test-session-456")

	identity, err := GetIdentityWithOverride("custom-brave@test")
	if err != nil {
		t.Fatalf("GetIdentityWithOverride failed: %v", err)
	}

	if identity.Agent != "custom" {
		t.Errorf("Expected agent 'custom', got %q", identity.Agent)
	}
	if identity.Suffix != "brave" {
		t.Errorf("Expected suffix 'brave', got %q", identity.Suffix)
	}
	if identity.Project != "test" {
		t.Errorf("Expected project 'test', got %q", identity.Project)
	}
}

func TestGetIdentityWithOverride_Empty(t *testing.T) {
	// Save original env
	origSmokeAuthor := os.Getenv("SMOKE_AUTHOR")
	origSessionID := os.Getenv("TERM_SESSION_ID")
	defer func() {
		os.Setenv("SMOKE_AUTHOR", origSmokeAuthor)
		os.Setenv("TERM_SESSION_ID", origSessionID)
	}()

	// Set up for GetIdentity to work
	os.Setenv("SMOKE_AUTHOR", "default-author")
	os.Setenv("TERM_SESSION_ID", "")

	identity, err := GetIdentityWithOverride("")
	if err != nil {
		t.Fatalf("GetIdentityWithOverride failed: %v", err)
	}

	// Should fall back to GetIdentity
	if identity.Suffix != "default-author" {
		t.Errorf("Expected suffix 'default-author', got %q", identity.Suffix)
	}
}
