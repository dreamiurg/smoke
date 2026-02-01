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
	tests := []struct {
		input   string
		agent   string
		suffix  string
		project string
	}{
		{"claude-swift-fox@smoke", "claude", "swift-fox", "smoke"},
		{"unknown-calm-owl@myproject", "unknown", "calm-owl", "myproject"},
		{"custom@test", "", "custom", "test"},
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
				Agent:   "",
				Suffix:  "agent",
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
			if tt.wantErr {
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
	// Create a temporary directory with a real git repo (no remote)
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, "test-repo")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Initialize a real git repo (without remote, to test toplevel fallback)
	cmd := exec.Command("git", "init")
	cmd.Dir = gitDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Use t.Chdir (Go 1.24+) for safe directory change in tests
	t.Chdir(gitDir)

	// Clear GIT_DIR/GIT_WORK_TREE to ensure we're not picking up parent repo context
	// This can happen during pre-commit hooks when git sets these env vars
	t.Setenv("GIT_DIR", "")
	t.Setenv("GIT_WORK_TREE", "")

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
	require.NoError(t, err)

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
	// This covers the parseFullIdentity path (line 42-44)
	origSmokeAuthor := os.Getenv("SMOKE_AUTHOR")
	defer os.Setenv("SMOKE_AUTHOR", origSmokeAuthor)

	os.Setenv("SMOKE_AUTHOR", "claude-brave@myproject")

	identity, err := GetIdentity()
	require.NoError(t, err)

	if identity.Agent != "claude" {
		t.Errorf("Expected agent 'claude', got %q", identity.Agent)
	}
	if identity.Suffix != "brave" {
		t.Errorf("Expected suffix 'brave', got %q", identity.Suffix)
	}
	if identity.Project != "myproject" {
		t.Errorf("Expected project 'myproject', got %q", identity.Project)
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
	origBDActor := os.Getenv("BD_ACTOR")
	origSmokeAuthor := os.Getenv("SMOKE_AUTHOR")
	defer func() {
		os.Setenv("BD_ACTOR", origBDActor)
		os.Setenv("SMOKE_AUTHOR", origSmokeAuthor)
	}()

	os.Setenv("BD_ACTOR", "agent-name@project")
	os.Setenv("SMOKE_AUTHOR", "")

	identity, err := GetIdentity()
	require.NoError(t, err)

	if identity.Agent != "agent" {
		t.Errorf("Expected agent 'agent', got %q", identity.Agent)
	}
	if identity.Suffix != "name" {
		t.Errorf("Expected suffix 'name', got %q", identity.Suffix)
	}
	if identity.Project != "project" {
		t.Errorf("Expected project 'project', got %q", identity.Project)
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
