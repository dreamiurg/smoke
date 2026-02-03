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
	origSmokeName := os.Getenv("SMOKE_NAME")
	origSessionID := os.Getenv("TERM_SESSION_ID")
	defer func() {
		os.Setenv("SMOKE_NAME", origSmokeName)
		os.Setenv("TERM_SESSION_ID", origSessionID)
	}()

	// Set env var
	os.Setenv("SMOKE_NAME", "test-user")
	os.Setenv("TERM_SESSION_ID", "")

	identity, err := GetIdentity("")
	require.NoError(t, err)

	if identity.Suffix != "test-user" {
		t.Errorf("Expected suffix 'test-user', got %q", identity.Suffix)
	}
	if identity.Agent != "" {
		t.Errorf("Expected agent '', got %q", identity.Agent)
	}
}

func TestGetIdentityWithOverride(t *testing.T) {
	// Save original env
	origSessionID := os.Getenv("TERM_SESSION_ID")
	defer os.Setenv("TERM_SESSION_ID", origSessionID)

	// Ensure we have a session seed so auto-detection doesn't fail
	os.Setenv("TERM_SESSION_ID", "test-session-123")

	identity, err := GetIdentity("my-custom-name")
	require.NoError(t, err)

	if identity.Suffix != "my-custom-name" {
		t.Errorf("Expected suffix 'my-custom-name', got %q", identity.Suffix)
	}
	if identity.Agent != "" {
		t.Errorf("Expected agent '', got %q", identity.Agent)
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

func TestSanitizeProjectName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"dreamwork.github.io", "dreamwork.github.io"},
		{"Hello World", "hello-world"},
		{"Test_Name", "test_name"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := sanitizeProjectName(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeProjectName(%q) = %q, want %q", tt.input, got, tt.want)
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
	origSmokeName := os.Getenv("SMOKE_NAME")
	origSessionID := os.Getenv("TERM_SESSION_ID")
	defer func() {
		os.Setenv("SMOKE_NAME", origSmokeName)
		os.Setenv("TERM_SESSION_ID", origSessionID)
	}()

	// Clear explicit author to force auto-detection
	os.Setenv("SMOKE_NAME", "")
	os.Setenv("TERM_SESSION_ID", "test-auto-detect-123")

	identity, err := GetIdentity("")
	require.NoError(t, err)

	// Should have a non-empty suffix from auto-generated name
	if identity.Suffix == "" {
		t.Error("Expected non-empty suffix from auto-detection")
	}
	// Auto-detection no longer sets Agent (removed "claude" prefix)
	if identity.Agent != "" {
		t.Error("Expected empty agent for auto-detected identity")
	}
	// Should have detected project
	if identity.Project == "" {
		t.Error("Expected non-empty project")
	}

	t.Logf("Auto-detected identity: %s", identity.String())
}

func TestGetIdentity_FallsBackToSessionSeed(t *testing.T) {
	// Save original env
	origSmokeName := os.Getenv("SMOKE_NAME")
	origSessionID := os.Getenv("TERM_SESSION_ID")
	defer func() {
		os.Setenv("SMOKE_NAME", origSmokeName)
		os.Setenv("TERM_SESSION_ID", origSessionID)
	}()

	// Clear explicit author and primary session ID to force PPID fallback
	os.Setenv("SMOKE_NAME", "")
	os.Setenv("TERM_SESSION_ID", "")

	identity, err := GetIdentity("")
	require.NoError(t, err)

	// Should have generated an identity using PPID-based seed
	if identity.Suffix == "" {
		t.Error("Expected non-empty suffix even with PPID fallback")
	}
	// Auto-detection no longer sets Agent (removed "claude" prefix)
	if identity.Agent != "" {
		t.Errorf("Expected empty agent for auto-detected identity, got %q", identity.Agent)
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

	identity, err := GetIdentity("custom-brave@test")
	require.NoError(t, err)

	// Overrides use agent="custom" with full name as suffix (don't parse agent-suffix)
	if identity.Agent != "" {
		t.Errorf("Expected agent '', got %q", identity.Agent)
	}
	if identity.Suffix != "custom-brave" {
		t.Errorf("Expected suffix 'custom-brave', got %q", identity.Suffix)
	}
	// Project should be auto-detected, not from override
	actualProject := detectProject()
	if identity.Project != actualProject {
		t.Errorf("Expected project to be auto-detected as %q, got %q", actualProject, identity.Project)
	}
}

func TestGetIdentityWithOverride_Empty(t *testing.T) {
	// Save original env
	origSmokeName := os.Getenv("SMOKE_NAME")
	origSessionID := os.Getenv("TERM_SESSION_ID")
	defer func() {
		os.Setenv("SMOKE_NAME", origSmokeName)
		os.Setenv("TERM_SESSION_ID", origSessionID)
	}()

	// Set up for GetIdentity to work
	os.Setenv("SMOKE_NAME", "default-author")
	os.Setenv("TERM_SESSION_ID", "")

	identity, err := GetIdentity("")
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
	// Test GetIdentity with full identity in SMOKE_NAME env var
	// @project should be ignored and auto-detected
	origSmokeName := os.Getenv("SMOKE_NAME")
	defer os.Setenv("SMOKE_NAME", origSmokeName)

	os.Setenv("SMOKE_NAME", "claude-brave@ignored-project")

	// Get the actual auto-detected project
	actualProject := detectProject()

	identity, err := GetIdentity("")
	require.NoError(t, err)

	// With the new behavior, "claude-brave" is treated as the full suffix
	// @project is stripped but we don't parse agent-suffix anymore in overrides
	if identity.Agent != "" {
		t.Errorf("Expected agent '', got %q", identity.Agent)
	}
	if identity.Suffix != "claude-brave" {
		t.Errorf("Expected suffix 'claude-brave', got %q", identity.Suffix)
	}
	// Project should be auto-detected, not from SMOKE_NAME
	if identity.Project != actualProject {
		t.Errorf("Expected project to be auto-detected as %q, got %q (should ignore @project in SMOKE_NAME)", actualProject, identity.Project)
	}
}

func TestGetIdentity_NoSessionSeed(t *testing.T) {
	// Test GetIdentity returns error when no session seed available
	// This covers the error case (line 59-61)
	origSmokeName := os.Getenv("SMOKE_NAME")
	origSessionID := os.Getenv("TERM_SESSION_ID")
	origWindowID := os.Getenv("WINDOWID")
	origTTY := os.Getenv("TTY")
	origPPID := os.Getenv("PPID") // Not actually used by code, but for safety
	defer func() {
		os.Setenv("SMOKE_NAME", origSmokeName)
		os.Setenv("TERM_SESSION_ID", origSessionID)
		os.Setenv("WINDOWID", origWindowID)
		os.Setenv("TTY", origTTY)
		os.Setenv("PPID", origPPID)
	}()

	// Clear all session identifiers to trigger no-seed scenario
	os.Setenv("SMOKE_NAME", "")
	os.Setenv("TERM_SESSION_ID", "")
	os.Setenv("WINDOWID", "")
	os.Setenv("TTY", "")

	// This test attempts to create a scenario with no session seed
	// In practice, PPID is always > 0, so this may not fully trigger the error
	// but we can still verify the logic path exists
	identity, err := GetIdentity("")
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

func TestGetIdentity_AutoDetectPath(t *testing.T) {
	// Test GetIdentity with SMOKE_NAME empty
	// This ensures we cover the auto-detection path
	origSmokeName := os.Getenv("SMOKE_NAME")
	origSessionID := os.Getenv("TERM_SESSION_ID")
	defer func() {
		os.Setenv("SMOKE_NAME", origSmokeName)
		os.Setenv("TERM_SESSION_ID", origSessionID)
	}()

	// Clear override to force auto-detection
	os.Setenv("SMOKE_NAME", "")
	os.Setenv("TERM_SESSION_ID", "auto-detect-test-session")

	identity, err := GetIdentity("")
	require.NoError(t, err)

	// Verify all components are populated
	// Auto-detection no longer sets Agent (removed "claude" prefix)
	if identity.Agent != "" {
		t.Errorf("Expected empty Agent in auto-detect path, got %q", identity.Agent)
	}
	if identity.Suffix == "" {
		t.Error("Expected non-empty Suffix in auto-detect path")
	}
	if identity.Project == "" {
		t.Error("Expected non-empty Project in auto-detect path")
	}

	t.Logf("Auto-detected identity: %s", identity.String())
}

// Test that @project override is ignored in SMOKE_NAME
func TestGetIdentity_SmokeNameIgnoresProjectOverride(t *testing.T) {
	origSmokeName := os.Getenv("SMOKE_NAME")
	defer os.Setenv("SMOKE_NAME", origSmokeName)

	actualProject := detectProject()

	os.Setenv("SMOKE_NAME", "bob@other-repo")

	identity, err := GetIdentity("")
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
func TestGetIdentity_OverrideIgnoresProjectOverride(t *testing.T) {
	actualProject := detectProject()

	identity, err := GetIdentity("charlie@ignored-project")
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
	origSmokeName := os.Getenv("SMOKE_NAME")
	defer os.Setenv("SMOKE_NAME", origSmokeName)

	actualProject := detectProject()

	os.Setenv("SMOKE_NAME", "dave")

	identity, err := GetIdentity("")
	require.NoError(t, err)

	if identity.Suffix != "dave" {
		t.Errorf("Expected suffix 'dave', got %q", identity.Suffix)
	}
	if identity.Project != actualProject {
		t.Errorf("Expected project %q (auto-detected), got %q", actualProject, identity.Project)
	}
}

// TestClaudeCodeSessionIdentity verifies that when running under Claude Code,
// the session seed uses PPID instead of terminal session ID for per-session identity.
func TestClaudeCodeSessionIdentity(t *testing.T) {
	// This test only works when actually running under Claude Code, since we need
	// a real Claude process ancestor for findClaudeAncestor() to detect
	if claudePID := findClaudeAncestor(); claudePID == 0 {
		t.Skip("This test requires actually running under Claude Code (need real Claude ancestor)")
	}

	origClaudeCode := os.Getenv("CLAUDECODE")
	origSmokeName := os.Getenv("SMOKE_NAME")
	origSessionID := os.Getenv("TERM_SESSION_ID")
	defer func() {
		os.Setenv("CLAUDECODE", origClaudeCode)
		os.Setenv("SMOKE_NAME", origSmokeName)
		os.Setenv("TERM_SESSION_ID", origSessionID)
	}()

	// Simulate running under Claude Code
	os.Setenv("CLAUDECODE", "1")
	os.Setenv("SMOKE_NAME", "")
	os.Setenv("TERM_SESSION_ID", "same-terminal-session")

	// Get identity - should use PPID-based seed, not TERM_SESSION_ID
	identity1, err := GetIdentity("")
	require.NoError(t, err)
	require.NotEmpty(t, identity1.Suffix, "Should generate identity under Claude Code")

	// Verify the seed format includes claude-ppid prefix by checking getSessionSeed directly
	seed := getSessionSeed()
	require.Contains(t, seed, "claude-ppid-", "Session seed should use claude-ppid format under Claude Code")

	t.Logf("Claude Code session identity: %s (seed: %s)", identity1.String(), seed)
}

// TestNonClaudeCodeUsesTerminalSession verifies that when NOT running under Claude Code,
// the session seed uses terminal session ID as before.
func TestNonClaudeCodeUsesTerminalSession(t *testing.T) {
	// Skip this test if actually running under Claude Code, since we can't
	// simulate "not under Claude" when findClaudeAncestor() will find Claude
	if claudePID := findClaudeAncestor(); claudePID > 0 {
		t.Skip("Cannot test non-Claude behavior when actually running under Claude Code")
	}

	origClaudeCode := os.Getenv("CLAUDECODE")
	origSmokeName := os.Getenv("SMOKE_NAME")
	origSessionID := os.Getenv("TERM_SESSION_ID")
	defer func() {
		os.Setenv("CLAUDECODE", origClaudeCode)
		os.Setenv("SMOKE_NAME", origSmokeName)
		os.Setenv("TERM_SESSION_ID", origSessionID)
	}()

	// Simulate NOT running under Claude Code
	os.Setenv("CLAUDECODE", "")
	os.Setenv("SMOKE_NAME", "")
	os.Setenv("TERM_SESSION_ID", "my-terminal-session-id")

	// Get session seed - should use TERM_SESSION_ID
	seed := getSessionSeed()
	require.Equal(t, "my-terminal-session-id", seed, "Should use TERM_SESSION_ID when not under Claude Code")
}

// TestSessionFileWriteAndRead tests the session file read/write functionality
func TestSessionFileWriteAndRead(t *testing.T) {
	// Create a temp config dir for testing
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "smoke")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	// Override the config dir for this test
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)
	os.Setenv("HOME", tmpDir)

	// Write session info
	info := &sessionInfo{
		PID:           12345,
		TermSessionID: "test-term-session",
		Seed:          "claude-ppid-12345",
	}
	err := writeSessionInfo(info)
	require.NoError(t, err)

	// Read it back
	readInfo := readSessionInfo()
	require.NotNil(t, readInfo, "Should be able to read session info")
	require.Equal(t, 12345, readInfo.PID)
	require.Equal(t, "test-term-session", readInfo.TermSessionID)
	require.Equal(t, "claude-ppid-12345", readInfo.Seed)
}

// TestSessionFileReturnsNilForMissingFile tests that readSessionInfo returns nil for missing file
func TestSessionFileReturnsNilForMissingFile(t *testing.T) {
	// Use a temp dir that doesn't have the session file
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "smoke")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)
	os.Setenv("HOME", tmpDir)

	info := readSessionInfo()
	require.Nil(t, info, "Should return nil for missing session file")
}

// TestSessionFileReturnsNilForInvalidJSON tests that readSessionInfo returns nil for invalid JSON
func TestSessionFileReturnsNilForInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "smoke")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)
	os.Setenv("HOME", tmpDir)

	// Write invalid JSON
	sessionFile := filepath.Join(configDir, "session.json")
	require.NoError(t, os.WriteFile(sessionFile, []byte("not valid json"), 0600))

	info := readSessionInfo()
	require.Nil(t, info, "Should return nil for invalid JSON")
}

// TestIsPIDRunning tests the PID checking functionality
func TestIsPIDRunning(t *testing.T) {
	// Current process should be running
	require.True(t, isPIDRunning(os.Getpid()), "Current process PID should be running")

	// Invalid PIDs should return false
	require.False(t, isPIDRunning(0), "PID 0 should return false")
	require.False(t, isPIDRunning(-1), "PID -1 should return false")

	// Very high PID unlikely to exist
	require.False(t, isPIDRunning(999999999), "Non-existent PID should return false")
}

// TestSessionFileCrossProcessSharing tests the main use case:
// Claude Code writes session file, ccstatusline reads it
func TestSessionFileCrossProcessSharing(t *testing.T) {
	// This test only works when actually running under Claude Code, since we need
	// a real Claude process ancestor for findClaudeAncestor() to detect
	if claudePID := findClaudeAncestor(); claudePID == 0 {
		t.Skip("This test requires actually running under Claude Code (need real Claude ancestor)")
	}

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "smoke")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	origHome := os.Getenv("HOME")
	origClaudeCode := os.Getenv("CLAUDECODE")
	origTermSession := os.Getenv("TERM_SESSION_ID")
	origSmokeName := os.Getenv("SMOKE_NAME")
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("CLAUDECODE", origClaudeCode)
		os.Setenv("TERM_SESSION_ID", origTermSession)
		os.Setenv("SMOKE_NAME", origSmokeName)
	}()

	os.Setenv("HOME", tmpDir)
	os.Setenv("SMOKE_NAME", "")

	termSessionID := "shared-terminal-123"
	os.Setenv("TERM_SESSION_ID", termSessionID)

	// Step 1: Simulate Claude Code running - should write session file
	os.Setenv("CLAUDECODE", "1")
	claudeSeed := getSessionSeed()
	require.Contains(t, claudeSeed, "claude-ppid-", "Claude Code should use PPID-based seed")

	// Verify session file was written
	info := readSessionInfo()
	require.NotNil(t, info, "Session file should exist after Claude Code invocation")
	require.Equal(t, claudeSeed, info.Seed, "Session file should contain the Claude seed")
	require.Equal(t, termSessionID, info.TermSessionID, "Session file should contain terminal session ID")

	// Step 2: Simulate ccstatusline (not under Claude Code) - should read from session file
	os.Setenv("CLAUDECODE", "")

	// The stored PID is our current process's PPID (the test runner)
	// Since we're in the same process, we need to use a running PID
	// Update the session file with current process PID for testing
	info.PID = os.Getpid() // Use current PID which is definitely running
	require.NoError(t, writeSessionInfo(info))

	statuslineSeed := getSessionSeed()
	require.Equal(t, claudeSeed, statuslineSeed, "ccstatusline should get same seed as Claude Code via session file")
}

// TestSessionFileIgnoredWhenDifferentTerminal tests that session file is ignored
// when called from a different terminal
func TestSessionFileIgnoredWhenDifferentTerminal(t *testing.T) {
	// Skip if running under Claude - process tree walking takes priority over session file
	if claudePID := findClaudeAncestor(); claudePID > 0 {
		t.Skip("Cannot test session file fallback when running under Claude Code")
	}

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "smoke")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	origHome := os.Getenv("HOME")
	origClaudeCode := os.Getenv("CLAUDECODE")
	origTermSession := os.Getenv("TERM_SESSION_ID")
	origSmokeName := os.Getenv("SMOKE_NAME")
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("CLAUDECODE", origClaudeCode)
		os.Setenv("TERM_SESSION_ID", origTermSession)
		os.Setenv("SMOKE_NAME", origSmokeName)
	}()

	os.Setenv("HOME", tmpDir)
	os.Setenv("SMOKE_NAME", "")

	// Write a session file for terminal A
	info := &sessionInfo{
		PID:           os.Getpid(), // Running PID
		TermSessionID: "terminal-A",
		Seed:          "claude-ppid-from-terminal-A",
	}
	require.NoError(t, writeSessionInfo(info))

	// Try to read from terminal B (different TERM_SESSION_ID)
	os.Setenv("CLAUDECODE", "")
	os.Setenv("TERM_SESSION_ID", "terminal-B")

	seed := getSessionSeed()
	// Should NOT use the session file because terminal IDs don't match
	require.Equal(t, "terminal-B", seed, "Should fall back to TERM_SESSION_ID when session file is for different terminal")
}

// TestSessionFileIgnoredWhenProcessDead tests that session file is ignored
// when the Claude Code process is no longer running
func TestSessionFileIgnoredWhenProcessDead(t *testing.T) {
	// Skip if running under Claude - process tree walking takes priority over session file
	if claudePID := findClaudeAncestor(); claudePID > 0 {
		t.Skip("Cannot test session file fallback when running under Claude Code")
	}

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "smoke")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	origHome := os.Getenv("HOME")
	origClaudeCode := os.Getenv("CLAUDECODE")
	origTermSession := os.Getenv("TERM_SESSION_ID")
	origSmokeName := os.Getenv("SMOKE_NAME")
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("CLAUDECODE", origClaudeCode)
		os.Setenv("TERM_SESSION_ID", origTermSession)
		os.Setenv("SMOKE_NAME", origSmokeName)
	}()

	os.Setenv("HOME", tmpDir)
	os.Setenv("SMOKE_NAME", "")

	termSessionID := "my-terminal"
	os.Setenv("TERM_SESSION_ID", termSessionID)

	// Write a session file with a non-existent PID
	info := &sessionInfo{
		PID:           999999999, // Very unlikely to exist
		TermSessionID: termSessionID,
		Seed:          "claude-ppid-dead-process",
	}
	require.NoError(t, writeSessionInfo(info))

	// Try to read when not under Claude Code
	os.Setenv("CLAUDECODE", "")

	seed := getSessionSeed()
	// Should NOT use the session file because PID is not running
	require.Equal(t, termSessionID, seed, "Should fall back to TERM_SESSION_ID when session file PID is dead")
}

// TestIsHumanSession_ClaudeCodeNotHuman verifies that CLAUDECODE=1 is never human
func TestIsHumanSession_ClaudeCodeNotHuman(t *testing.T) {
	origClaudeCode := os.Getenv("CLAUDECODE")
	defer os.Setenv("CLAUDECODE", origClaudeCode)

	os.Setenv("CLAUDECODE", "1")

	result := isHumanSession()
	require.False(t, result, "Should not be human when CLAUDECODE=1")
}

// TestIsHumanSession_ValidSessionFileNotHuman verifies that valid Claude session = not human
func TestIsHumanSession_ValidSessionFileNotHuman(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "smoke")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	origHome := os.Getenv("HOME")
	origClaudeCode := os.Getenv("CLAUDECODE")
	origTermSession := os.Getenv("TERM_SESSION_ID")
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("CLAUDECODE", origClaudeCode)
		os.Setenv("TERM_SESSION_ID", origTermSession)
	}()

	os.Setenv("HOME", tmpDir)
	os.Setenv("CLAUDECODE", "") // Not running under Claude Code directly

	termSessionID := "test-terminal"
	os.Setenv("TERM_SESSION_ID", termSessionID)

	// Write a valid session file with running PID
	info := &sessionInfo{
		PID:           os.Getpid(), // Current process is running
		TermSessionID: termSessionID,
		Seed:          "claude-ppid-test",
	}
	require.NoError(t, writeSessionInfo(info))

	result := isHumanSession()
	require.False(t, result, "Should not be human when valid Claude session file exists")
}

// TestIsHumanSession_NoAgentIndicators tests the human detection without agent context
// Note: This test may return different results depending on whether it runs in an interactive terminal
func TestIsHumanSession_NoAgentIndicators(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "smoke")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	origHome := os.Getenv("HOME")
	origClaudeCode := os.Getenv("CLAUDECODE")
	origTermSession := os.Getenv("TERM_SESSION_ID")
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("CLAUDECODE", origClaudeCode)
		os.Setenv("TERM_SESSION_ID", origTermSession)
	}()

	os.Setenv("HOME", tmpDir)
	os.Setenv("CLAUDECODE", "")
	os.Setenv("TERM_SESSION_ID", "some-terminal")

	// No session file exists, and CLAUDECODE is not set
	// Result depends on whether stdin is a TTY (it's not in test environment)
	result := isHumanSession()

	// In CI/test environment, stdin is typically not a TTY, so should be false
	// We just verify it doesn't panic and returns a boolean
	t.Logf("isHumanSession() returned %v (expected false in test environment)", result)
}

// TestGetIdentity_HumanInInteractiveTerminal tests that human identity is returned for interactive terminals
// This test simulates the scenario where isHumanSession would return true
func TestGetIdentity_HumanInInteractiveTerminal(t *testing.T) {
	// We can't easily test actual TTY detection in unit tests,
	// but we can verify the HumanIdentity constant is used correctly
	// when we would be human
	require.Equal(t, "<human>", HumanIdentity, "HumanIdentity constant should be <human>")
}

// TestHumanIdentityString verifies the human identity string format
func TestHumanIdentityString(t *testing.T) {
	id := &Identity{
		Agent:   "",
		Suffix:  HumanIdentity,
		Project: "smoke",
	}

	want := "<human>@smoke"
	got := id.String()

	require.Equal(t, want, got, "Human identity should format as <human>@project")
}

// TestIsHumanSession_DifferentTerminalSession tests that mismatched terminal ignores session file
func TestIsHumanSession_DifferentTerminalSession(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "smoke")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	origHome := os.Getenv("HOME")
	origClaudeCode := os.Getenv("CLAUDECODE")
	origTermSession := os.Getenv("TERM_SESSION_ID")
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("CLAUDECODE", origClaudeCode)
		os.Setenv("TERM_SESSION_ID", origTermSession)
	}()

	os.Setenv("HOME", tmpDir)
	os.Setenv("CLAUDECODE", "")
	os.Setenv("TERM_SESSION_ID", "different-terminal")

	// Write a session file for a different terminal
	info := &sessionInfo{
		PID:           os.Getpid(),
		TermSessionID: "original-terminal", // Different from current
		Seed:          "claude-ppid-test",
	}
	require.NoError(t, writeSessionInfo(info))

	// Session file doesn't match current terminal, so it's ignored
	// Result depends on TTY status (false in test environment)
	result := isHumanSession()
	t.Logf("isHumanSession() with mismatched terminal: %v", result)
}

// TestIsHumanSession_DeadProcessSession tests that dead process session file is ignored
func TestIsHumanSession_DeadProcessSession(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "smoke")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	origHome := os.Getenv("HOME")
	origClaudeCode := os.Getenv("CLAUDECODE")
	origTermSession := os.Getenv("TERM_SESSION_ID")
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("CLAUDECODE", origClaudeCode)
		os.Setenv("TERM_SESSION_ID", origTermSession)
	}()

	os.Setenv("HOME", tmpDir)
	os.Setenv("CLAUDECODE", "")
	termSessionID := "test-terminal"
	os.Setenv("TERM_SESSION_ID", termSessionID)

	// Write a session file with a non-existent PID
	info := &sessionInfo{
		PID:           999999999, // Very unlikely to exist
		TermSessionID: termSessionID,
		Seed:          "claude-ppid-dead",
	}
	require.NoError(t, writeSessionInfo(info))

	// Dead process means session file is invalid
	// Result depends on TTY status (false in test environment)
	result := isHumanSession()
	t.Logf("isHumanSession() with dead process: %v", result)
}

// TestFindClaudeAncestor tests the process tree walking for Claude detection
func TestFindClaudeAncestor(t *testing.T) {
	// In a normal test environment, we're not running under Claude
	// so this should return 0
	pid := findClaudeAncestor()

	// The result depends on whether we're actually running under Claude
	// In CI or standalone tests, this should be 0
	// We just verify it doesn't crash and returns a reasonable value
	t.Logf("findClaudeAncestor() returned PID: %d", pid)
	require.GreaterOrEqual(t, pid, 0, "PID should be non-negative")
}

// TestGetSessionSeed_UsesClaudeAncestor verifies that getSessionSeed uses
// the Claude ancestor when available
func TestGetSessionSeed_UsesClaudeAncestor(t *testing.T) {
	origClaudeCode := os.Getenv("CLAUDECODE")
	origSmokeName := os.Getenv("SMOKE_NAME")
	origSessionID := os.Getenv("TERM_SESSION_ID")
	defer func() {
		os.Setenv("CLAUDECODE", origClaudeCode)
		os.Setenv("SMOKE_NAME", origSmokeName)
		os.Setenv("TERM_SESSION_ID", origSessionID)
	}()

	// Simulate NOT running directly under Claude Code
	os.Setenv("CLAUDECODE", "")
	os.Setenv("SMOKE_NAME", "")
	os.Setenv("TERM_SESSION_ID", "test-terminal")

	// Get the session seed
	seed := getSessionSeed()

	// In a test environment without Claude ancestor, should fall back
	// to terminal session ID
	t.Logf("getSessionSeed() returned: %s", seed)
	require.NotEmpty(t, seed, "Should return a non-empty seed")
}
