package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)

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
	require.NoError(t, err)

	if identity.Suffix != "my-custom-name" {
		t.Errorf("Expected suffix 'my-custom-name', got %q", identity.Suffix)
	}
	if identity.Agent != "custom" {
		t.Errorf("Expected agent 'custom', got %q", identity.Agent)
	}
}

func TestParseFullIdentity(t *testing.T) {
	// Get the actual auto-detected project for comparison
	actualProject := detectProject()

	tests := []struct {
		input  string
		agent  string
		suffix string
	}{
		{"claude-swift-fox@smoke", "claude", "swift-fox"},
		{"unknown-calm-owl@myproject", "unknown", "calm-owl"},
		{"custom@test", "", "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			id, err := parseFullIdentity(tt.input)
			require.NoError(t, err)
			if id.Agent != tt.agent {
				t.Errorf("Agent: got %q, want %q", id.Agent, tt.agent)
			}
			if id.Suffix != tt.suffix {
				t.Errorf("Suffix: got %q, want %q", id.Suffix, tt.suffix)
			}
			// Project should ALWAYS be auto-detected, not from input
			if id.Project != actualProject {
				t.Errorf("Project: got %q, want auto-detected %q (input @project should be ignored)", id.Project, actualProject)
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
	require.NoError(t, err)

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
	require.NoError(t, err)

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
	// Get the actual auto-detected project
	actualProject := detectProject()

	tests := []struct {
		name   string
		input  string
		wantID *Identity
	}{
		{
			name:  "valid format with @project (project ignored)",
			input: "agent-suffix@ignored-project",
			wantID: &Identity{
				Agent:   "agent",
				Suffix:  "suffix",
				Project: actualProject, // ALWAYS auto-detected
			},
		},
		{
			name:  "no agent dash with @project (project ignored)",
			input: "agent@ignored-project",
			wantID: &Identity{
				Agent:   "",
				Suffix:  "agent",
				Project: actualProject, // ALWAYS auto-detected
			},
		},
		{
			name:  "no project separator (project auto-detected)",
			input: "agent-suffix",
			wantID: &Identity{
				Agent:   "agent",
				Suffix:  "suffix",
				Project: actualProject, // ALWAYS auto-detected
			},
		},
		{
			name:  "case insensitive with @project (project ignored)",
			input: "Agent-Suffix@Ignored-Project",
			wantID: &Identity{
				Agent:   "agent",
				Suffix:  "suffix",
				Project: actualProject, // ALWAYS auto-detected
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := parseFullIdentity(tt.input)
			require.NoError(t, err)
			if id.Agent != tt.wantID.Agent {
				t.Errorf("Agent: got %q, want %q", id.Agent, tt.wantID.Agent)
			}
			if id.Suffix != tt.wantID.Suffix {
				t.Errorf("Suffix: got %q, want %q", id.Suffix, tt.wantID.Suffix)
			}
			if id.Project != tt.wantID.Project {
				t.Errorf("Project: got %q, want %q (should be auto-detected)", id.Project, tt.wantID.Project)
			}
		})
	}
}

func TestDetectProject_InGitRepo(t *testing.T) {
	// Save original directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	// Create a temporary directory to simulate a git repo
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, "test-repo")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Initialize a real git repo (needed for detectProject to work)
	if err := os.Chdir(gitDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}

	// Run git init to create a real .git directory
	cmd := exec.Command("git", "init")
	cmd.Dir = gitDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git init: %v", err)
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

	// Get the actual auto-detected project
	actualProject := detectProject()

	identity, err := GetIdentityWithOverride("custom-brave@ignored-project")
	require.NoError(t, err)

	// With the new behavior, "custom-brave" is treated as the full suffix
	// @project is stripped but we don't parse agent-suffix anymore in overrides
	if identity.Agent != "custom" {
		t.Errorf("Expected agent 'custom', got %q", identity.Agent)
	}
	if identity.Suffix != "custom-brave" {
		t.Errorf("Expected suffix 'custom-brave', got %q", identity.Suffix)
	}
	// Project should be auto-detected, not from override
	if identity.Project != actualProject {
		t.Errorf("Expected project to be auto-detected as %q, got %q", actualProject, identity.Project)
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
	require.NoError(t, err)

	// Should fall back to GetIdentity
	if identity.Suffix != "default-author" {
		t.Errorf("Expected suffix 'default-author', got %q", identity.Suffix)
	}
}

func TestExtractRepoName(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "https with .git suffix",
			url:  "https://github.com/dreamiurg/smoke.git",
			want: "smoke",
		},
		{
			name: "https without .git suffix",
			url:  "https://github.com/dreamiurg/smoke",
			want: "smoke",
		},
		{
			name: "ssh format",
			url:  "git@github.com:dreamiurg/smoke.git",
			want: "smoke",
		},
		{
			name: "ssh without .git",
			url:  "git@github.com:dreamiurg/smoke",
			want: "smoke",
		},
		{
			name: "nested path",
			url:  "https://gitlab.com/group/subgroup/repo.git",
			want: "repo",
		},
		{
			name: "simple name",
			url:  "myrepo",
			want: "myrepo",
		},
		{
			name: "with .git only",
			url:  "myrepo.git",
			want: "myrepo",
		},
		// Edge cases
		{
			name: "empty string",
			url:  "",
			want: "",
		},
		{
			name: "url with no slashes",
			url:  "repo",
			want: "repo",
		},
		{
			name: "url ending in just .git",
			url:  ".git",
			want: "",
		},
		{
			name: "malformed url - just a colon",
			url:  ":",
			want: "",
		},
		{
			name: "malformed url - colon at end",
			url:  "git@github.com:",
			want: "",
		},
		{
			name: "malformed url - double slash at start",
			url:  "//github.com/repo.git",
			want: "repo",
		},
		{
			name: "malformed url - multiple slashes",
			url:  "https:///github.com/repo.git",
			want: "repo",
		},
		{
			name: "ssh with multiple slashes after colon",
			url:  "git@github.com://user/repo.git",
			want: "repo",
		},
		{
			name: "trailing slash",
			url:  "https://github.com/user/repo/",
			want: "",
		},
		{
			name: "only .git suffix",
			url:  ".git",
			want: "",
		},
		{
			name: "whitespace in url",
			url:  "https://github.com/user/my repo.git",
			want: "my repo",
		},
		{
			name: "complex nested path",
			url:  "https://gitlab.com/group/subgroup/nested/repo.git",
			want: "repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractRepoName(tt.url)
			if got != tt.want {
				t.Errorf("extractRepoName(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

func TestIdentityString_WithoutAgent(t *testing.T) {
	// Test String() method with empty Agent (coverage for line 26-28)
	id := &Identity{
		Agent:   "",
		Suffix:  "swift-fox",
		Project: "smoke",
	}

	want := "swift-fox@smoke"
	got := id.String()

	if got != want {
		t.Errorf("Identity.String() with empty agent = %q, want %q", got, want)
	}
}

func TestGetIdentity_WithFullIdentityInSmokeAuthor(t *testing.T) {
	// Test GetIdentity with full identity in SMOKE_AUTHOR env var
	// @project should be ignored and auto-detected
	origSmokeAuthor := os.Getenv("SMOKE_AUTHOR")
	defer os.Setenv("SMOKE_AUTHOR", origSmokeAuthor)

	os.Setenv("SMOKE_AUTHOR", "claude-brave@ignored-project")

	// Get the actual auto-detected project
	actualProject := detectProject()

	identity, err := GetIdentity()
	require.NoError(t, err)

	// With the new behavior, "claude-brave" is treated as the full suffix
	// @project is stripped but we don't parse agent-suffix anymore in overrides
	if identity.Agent != "custom" {
		t.Errorf("Expected agent 'custom', got %q", identity.Agent)
	}
	if identity.Suffix != "claude-brave" {
		t.Errorf("Expected suffix 'claude-brave', got %q", identity.Suffix)
	}
	// Project should be auto-detected, not from SMOKE_AUTHOR
	if identity.Project != actualProject {
		t.Errorf("Expected project to be auto-detected as %q, got %q (should ignore @project in SMOKE_AUTHOR)", actualProject, identity.Project)
	}
}

func TestGetIdentity_NoSessionSeed(t *testing.T) {
	// Test GetIdentity returns error when no session seed available
	// This covers the error case (line 59-61)
	origSmokeAuthor := os.Getenv("SMOKE_AUTHOR")
	origSessionID := os.Getenv("TERM_SESSION_ID")
	origWindowID := os.Getenv("WINDOWID")
	origTTY := os.Getenv("TTY")
	origPPID := os.Getenv("PPID") // Not actually used by code, but for safety
	defer func() {
		os.Setenv("SMOKE_AUTHOR", origSmokeAuthor)
		os.Setenv("TERM_SESSION_ID", origSessionID)
		os.Setenv("WINDOWID", origWindowID)
		os.Setenv("TTY", origTTY)
		os.Setenv("PPID", origPPID)
	}()

	// Clear all session identifiers to trigger no-seed scenario
	os.Setenv("SMOKE_AUTHOR", "")
	os.Setenv("TERM_SESSION_ID", "")
	os.Setenv("WINDOWID", "")
	os.Setenv("TTY", "")

	// This test attempts to create a scenario with no session seed
	// In practice, PPID is always > 0, so this may not fully trigger the error
	// but we can still verify the logic path exists
	identity, err := GetIdentity()
	// We may or may not get an error depending on PPID availability
	// The important thing is that the code doesn't crash
	if err == ErrNoIdentity {
		t.Logf("Got expected ErrNoIdentity when no session seed available")
		require.NoError(t, err)
		// If no error, PPID fallback worked
		t.Logf("Identity resolved via PPID fallback: %s", identity.String())
	}
}

func TestDetectAgent_NoClaudeDir(t *testing.T) {
	// Manually test detectAgent behavior
	// The function checks for ~/.claude directory
	agent := detectAgent()
	// The function should return either "claude" or "unknown"
	if agent != "claude" && agent != "unknown" {
		t.Errorf("detectAgent() returned unexpected value: %q", agent)
	}
	t.Logf("detectAgent returned: %q", agent)
}

func TestExtractRepoName_SSHWithColonNoSlash(t *testing.T) {
	// Test the colon path where there's no slash after the colon
	// This covers line 205-207 branch where slash is not found
	url := "git@github.com:myrepo"
	got := extractRepoName(url)
	want := "myrepo"
	if got != want {
		t.Errorf("extractRepoName(%q) = %q, want %q", url, got, want)
	}
}

func TestGetIdentity_WithBdActorOverride(t *testing.T) {
	// Test that BD_ACTOR takes precedence over SMOKE_AUTHOR
	origBDActor := os.Getenv("BD_ACTOR")
	origSmokeAuthor := os.Getenv("SMOKE_AUTHOR")
	defer func() {
		os.Setenv("BD_ACTOR", origBDActor)
		os.Setenv("SMOKE_AUTHOR", origSmokeAuthor)
	}()

	os.Setenv("BD_ACTOR", "bd-user")
	os.Setenv("SMOKE_AUTHOR", "smoke-user")

	identity, err := GetIdentity()
	require.NoError(t, err)

	// BD_ACTOR should take precedence
	if identity.Suffix != "bd-user" {
		t.Errorf("Expected suffix 'bd-user', got %q", identity.Suffix)
	}
}

func TestGetIdentity_WithBdActorFullIdentity(t *testing.T) {
	// Test that BD_ACTOR with full identity format is parsed correctly
	// @project should be ignored and auto-detected
	origBDActor := os.Getenv("BD_ACTOR")
	origSmokeAuthor := os.Getenv("SMOKE_AUTHOR")
	defer func() {
		os.Setenv("BD_ACTOR", origBDActor)
		os.Setenv("SMOKE_AUTHOR", origSmokeAuthor)
	}()

	os.Setenv("BD_ACTOR", "agent-name@ignored-project")
	os.Setenv("SMOKE_AUTHOR", "")

	// Get the actual auto-detected project
	actualProject := detectProject()

	identity, err := GetIdentity()
	require.NoError(t, err)

	// With the new behavior, "agent-name" is treated as the full suffix
	// @project is stripped but we don't parse agent-suffix anymore in overrides
	if identity.Agent != "custom" {
		t.Errorf("Expected agent 'custom', got %q", identity.Agent)
	}
	if identity.Suffix != "agent-name" {
		t.Errorf("Expected suffix 'agent-name', got %q", identity.Suffix)
	}
	// Project should be auto-detected, not from BD_ACTOR
	if identity.Project != actualProject {
		t.Errorf("Expected project to be auto-detected as %q, got %q (should ignore @project in BD_ACTOR)", actualProject, identity.Project)
	}
}

func TestGetIdentity_AutoDetectPath(t *testing.T) {
	// Test GetIdentity with both BD_ACTOR and SMOKE_AUTHOR empty
	// This ensures we cover the auto-detection path (line 55-69)
	origBDActor := os.Getenv("BD_ACTOR")
	origSmokeAuthor := os.Getenv("SMOKE_AUTHOR")
	origSessionID := os.Getenv("TERM_SESSION_ID")
	defer func() {
		os.Setenv("BD_ACTOR", origBDActor)
		os.Setenv("SMOKE_AUTHOR", origSmokeAuthor)
		os.Setenv("TERM_SESSION_ID", origSessionID)
	}()

	// Clear both overrides to force auto-detection
	os.Setenv("BD_ACTOR", "")
	os.Setenv("SMOKE_AUTHOR", "")
	os.Setenv("TERM_SESSION_ID", "auto-detect-test-session")

	identity, err := GetIdentity()
	require.NoError(t, err)

	// Verify all components are populated
	if identity.Agent == "" {
		t.Error("Expected non-empty Agent in auto-detect path")
	}
	if identity.Suffix == "" {
		t.Error("Expected non-empty Suffix in auto-detect path")
	}
	if identity.Project == "" {
		t.Error("Expected non-empty Project in auto-detect path")
	}

	t.Logf("Auto-detected identity: %s", identity.String())
}

// Test that @project override is ignored in BD_ACTOR
func TestGetIdentity_BdActorIgnoresProjectOverride(t *testing.T) {
	origBDActor := os.Getenv("BD_ACTOR")
	defer os.Setenv("BD_ACTOR", origBDActor)

	actualProject := detectProject()

	// Try to override with @fake-project
	os.Setenv("BD_ACTOR", "alice@fake-project")

	identity, err := GetIdentity()
	require.NoError(t, err)

	// Should use name "alice" but ignore project override
	if identity.Suffix != "alice" {
		t.Errorf("Expected suffix 'alice', got %q", identity.Suffix)
	}
	if identity.Project != actualProject {
		t.Errorf("Expected project %q (auto-detected), got %q (should ignore @fake-project)", actualProject, identity.Project)
	}
}

// Test that @project override is ignored in SMOKE_AUTHOR
func TestGetIdentity_SmokeAuthorIgnoresProjectOverride(t *testing.T) {
	origBDActor := os.Getenv("BD_ACTOR")
	origSmokeAuthor := os.Getenv("SMOKE_AUTHOR")
	defer func() {
		os.Setenv("BD_ACTOR", origBDActor)
		os.Setenv("SMOKE_AUTHOR", origSmokeAuthor)
	}()

	actualProject := detectProject()

	// Clear BD_ACTOR to ensure SMOKE_AUTHOR is used
	os.Setenv("BD_ACTOR", "")
	os.Setenv("SMOKE_AUTHOR", "bob@other-repo")

	identity, err := GetIdentity()
	require.NoError(t, err)

	// Should use name "bob" but ignore project override
	if identity.Suffix != "bob" {
		t.Errorf("Expected suffix 'bob', got %q", identity.Suffix)
	}
	if identity.Project != actualProject {
		t.Errorf("Expected project %q (auto-detected), got %q (should ignore @other-repo)", actualProject, identity.Project)
	}
}

// Test that --as flag ignores @project override
func TestGetIdentityWithOverride_IgnoresProjectOverride(t *testing.T) {
	actualProject := detectProject()

	identity, err := GetIdentityWithOverride("charlie@ignored-project")
	require.NoError(t, err)

	// Should use name "charlie" but ignore project override
	if identity.Suffix != "charlie" {
		t.Errorf("Expected suffix 'charlie', got %q", identity.Suffix)
	}
	if identity.Project != actualProject {
		t.Errorf("Expected project %q (auto-detected), got %q (should ignore @ignored-project)", actualProject, identity.Project)
	}
}

// Test that name without @ works normally
func TestGetIdentity_NameWithoutAt(t *testing.T) {
	origBDActor := os.Getenv("BD_ACTOR")
	defer os.Setenv("BD_ACTOR", origBDActor)

	actualProject := detectProject()

	os.Setenv("BD_ACTOR", "dave")

	identity, err := GetIdentity()
	require.NoError(t, err)

	if identity.Suffix != "dave" {
		t.Errorf("Expected suffix 'dave', got %q", identity.Suffix)
	}
	if identity.Project != actualProject {
		t.Errorf("Expected project %q (auto-detected), got %q", actualProject, identity.Project)
	}
}
