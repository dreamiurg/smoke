package integration

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestHelper manages test environment
type TestHelper struct {
	t           *testing.T
	tmpDir      string
	configDir   string
	binPath     string
	origDir     string
	origBDActor string
	origEnv     map[string]string
}

func NewTestHelper(t *testing.T) *TestHelper {
	t.Helper()

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "smoke")

	// Find smoke binary
	binPath := os.Getenv("SMOKE_BIN")
	if binPath == "" {
		// Get the current working directory
		cwd, _ := os.Getwd()

		// Try to find it relative to various locations
		possiblePaths := []string{
			filepath.Join(cwd, "bin", "smoke"),
			filepath.Join(cwd, "..", "..", "bin", "smoke"),
			"../../bin/smoke",
			"../bin/smoke",
			"./bin/smoke",
		}
		for _, p := range possiblePaths {
			absPath, _ := filepath.Abs(p)
			if _, err := os.Stat(absPath); err == nil {
				binPath = absPath
				break
			}
		}
	}

	origDir, _ := os.Getwd()
	origBDActor := os.Getenv("BD_ACTOR")

	// Save original HOME
	origEnv := map[string]string{
		"HOME": os.Getenv("HOME"),
	}

	return &TestHelper{
		t:           t,
		tmpDir:      tmpDir,
		configDir:   configDir,
		binPath:     binPath,
		origDir:     origDir,
		origBDActor: origBDActor,
		origEnv:     origEnv,
	}
}

func (h *TestHelper) Cleanup() {
	os.Chdir(h.origDir)
	os.Setenv("BD_ACTOR", h.origBDActor)
	for k, v := range h.origEnv {
		os.Setenv(k, v)
	}
}

func (h *TestHelper) SetIdentity(identity string) {
	os.Setenv("BD_ACTOR", identity)
}

func (h *TestHelper) Run(args ...string) (string, string, error) {
	h.t.Helper()

	if h.binPath == "" {
		h.t.Skip("smoke binary not found. Set SMOKE_BIN or build with 'make build'")
		return "", "", nil
	}

	// Verify binary exists
	if _, err := os.Stat(h.binPath); os.IsNotExist(err) {
		h.t.Skip("smoke binary not found at " + h.binPath + ". Build with 'make build'")
		return "", "", nil
	}

	cmd := exec.Command(h.binPath, args...)
	cmd.Dir = h.tmpDir

	// Set HOME to tmpDir so smoke uses tmpDir/.config/smoke/
	env := os.Environ()
	env = append(env, "HOME="+h.tmpDir)
	cmd.Env = env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func TestSmokeInit(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	stdout, _, err := h.Run("init")
	if err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Check output
	if !strings.Contains(stdout, "Initialized smoke") {
		t.Errorf("init output missing success message: %s", stdout)
	}

	// Check config directory created
	if _, err := os.Stat(h.configDir); os.IsNotExist(err) {
		t.Error("config directory not created")
	}

	// Check feed.jsonl created
	feedFile := filepath.Join(h.configDir, "feed.jsonl")
	if _, err := os.Stat(feedFile); os.IsNotExist(err) {
		t.Error("feed.jsonl not created")
	}
}

func TestSmokeInitIdempotent(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// First init
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("First smoke init failed: %v", err)
	}

	// Second init should not fail
	stdout, _, err := h.Run("init")
	if err != nil {
		t.Fatalf("Second smoke init failed: %v", err)
	}

	if !strings.Contains(stdout, "already initialized") {
		t.Errorf("Expected 'already initialized' message: %s", stdout)
	}
}

func TestSmokeInitSeeds(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	stdout, _, err := h.Run("init")
	if err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Should show seeding message
	if !strings.Contains(stdout, "Seeded 4 example posts") {
		t.Errorf("Expected seeding message in output: %s", stdout)
	}

	// Verify feed has 4 posts
	feedOut, _, err := h.Run("feed")
	if err != nil {
		t.Fatalf("smoke feed failed: %v", err)
	}

	// Check for example authors
	for _, author := range []string{"spark", "ember", "flare", "wisp"} {
		if !strings.Contains(feedOut, author) {
			t.Errorf("Expected author %q in feed output: %s", author, feedOut)
		}
	}
}

func TestSmokeInitDoesNotReseed(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// First init - should seed
	_, _, err := h.Run("init")
	if err != nil {
		t.Fatalf("First init failed: %v", err)
	}

	// Add a user post
	h.SetIdentity("testuser@test")
	h.Run("post", "user post after init")

	// Second init (no --force) - should not reseed
	_, _, err = h.Run("init")
	if err != nil {
		t.Fatalf("Second init failed: %v", err)
	}

	// Feed should have 5 posts (4 seeded + 1 user)
	feedOut, feedErr, err := h.Run("feed", "-n", "10")
	if err != nil {
		t.Fatalf("smoke feed failed: %v (stderr: %s)", err, feedErr)
	}

	// User post should be present
	if !strings.Contains(feedOut, "user post after init") {
		t.Errorf("Expected user post preserved after second init.\nGot stdout: %s\nstderr: %s", feedOut, feedErr)
	}

	// Count posts by checking for authors (more reliable than smk-)
	postCount := 0
	for _, author := range []string{"spark", "ember", "flare", "wisp", "testuser"} {
		if strings.Contains(feedOut, author) {
			postCount++
		}
	}
	if postCount != 5 {
		t.Errorf("Expected 5 posts (4 seeded + 1 user), got %d.\nFeed output: %s", postCount, feedOut)
	}
}

func TestSmokeInitDryRunDoesNotSeed(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Dry-run init
	stdout, _, err := h.Run("init", "--dry-run")
	if err != nil {
		t.Fatalf("smoke init --dry-run failed: %v", err)
	}

	// Should show what would be done
	if !strings.Contains(stdout, "[dry-run]") {
		t.Errorf("Expected dry-run output: %s", stdout)
	}

	// Feed file should not exist or be empty
	// Post should fail because not initialized
	_, _, postErr := h.Run("post", "should fail")
	if postErr == nil {
		t.Error("Expected post to fail after dry-run init, but it succeeded")
	}
}

func TestSmokePost(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Set identity
	h.SetIdentity("testuser@testrig")

	// Post a message
	stdout, _, err := h.Run("post", "hello from integration test")
	if err != nil {
		t.Fatalf("smoke post failed: %v", err)
	}

	if !strings.Contains(stdout, "Posted smk-") {
		t.Errorf("post output missing confirmation: %s", stdout)
	}

	// Verify post in feed
	feedFile := filepath.Join(h.configDir, "feed.jsonl")
	content, _ := os.ReadFile(feedFile)
	if !strings.Contains(string(content), "hello from integration test") {
		t.Errorf("Post not found in feed file: %s", content)
	}
}

func TestSmokePostAutoIdentity(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Clear explicit identity env vars - smoke should auto-generate
	os.Unsetenv("BD_ACTOR")
	os.Unsetenv("SMOKE_AUTHOR")

	// Post should succeed with auto-generated identity
	stdout, _, err := h.Run("post", "test message with auto identity")
	if err != nil {
		t.Fatalf("smoke post with auto identity failed: %v", err)
	}

	if !strings.Contains(stdout, "Posted smk-") {
		t.Errorf("post output missing confirmation: %s", stdout)
	}
}

func TestSmokePostTooLong(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	h.SetIdentity("testuser@testrig")

	// Post a message that's too long
	longMessage := strings.Repeat("a", 281)
	_, stderr, err := h.Run("post", longMessage)
	if err == nil {
		t.Error("smoke post should fail for message > 280 chars")
	}

	if !strings.Contains(stderr, "280 characters") {
		t.Errorf("Expected character limit error: %s", stderr)
	}
}

func TestSmokePostNotInitialized(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	h.SetIdentity("testuser@testrig")

	// Post without init
	_, stderr, err := h.Run("post", "test message")
	if err == nil {
		t.Error("smoke post should fail without init")
	}

	if !strings.Contains(stderr, "not initialized") {
		t.Errorf("Expected 'not initialized' error: %s", stderr)
	}
}

func TestSmokeFeed(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Note: @testrig is ignored, project is auto-detected
	h.SetIdentity("ember@ignored-testrig")

	// Post a few messages
	h.Run("post", "first post")
	h.Run("post", "second post")
	h.Run("post", "third post")

	// Read feed
	stdout, _, err := h.Run("feed")
	if err != nil {
		t.Fatalf("smoke feed failed: %v", err)
	}

	// Check output contains posts
	if !strings.Contains(stdout, "first post") {
		t.Errorf("feed missing first post: %s", stdout)
	}
	if !strings.Contains(stdout, "third post") {
		t.Errorf("feed missing third post: %s", stdout)
	}
	// Note: project is auto-detected, and "ember@testrig" becomes "ember@<auto-detected>"
	if !strings.Contains(stdout, "ember@") {
		t.Errorf("feed missing author (should be ember@{auto-detected}): %s", stdout)
	}
}

func TestSmokeFeedLimit(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Note: @testrig is ignored, project is auto-detected
	h.SetIdentity("ember@ignored-testrig")

	// Post 5 messages
	for i := 0; i < 5; i++ {
		h.Run("post", "post number "+string(rune('0'+i)))
	}

	// Read feed with limit 2
	stdout, _, err := h.Run("feed", "-n", "2")
	if err != nil {
		t.Fatalf("smoke feed failed: %v", err)
	}

	// Count posts in output (look for author pattern)
	// Note: project is auto-detected, and "ember@testrig" becomes "ember@<auto-detected>"
	count := strings.Count(stdout, "ember@")
	if count != 2 {
		t.Errorf("feed -n 2 returned %d posts, want 2.\nFeed output:\n%s", count, stdout)
	}
}

func TestSmokeFeedAuthorFilter(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Post from different users
	h.SetIdentity("ember@testrig")
	h.Run("post", "ember's post")

	h.SetIdentity("witness@testrig")
	h.Run("post", "witness's post")

	// Filter by author
	stdout, _, err := h.Run("feed", "--author", "ember")
	if err != nil {
		t.Fatalf("smoke feed failed: %v", err)
	}

	if !strings.Contains(stdout, "ember's post") {
		t.Errorf("feed --author ember missing ember's post: %s", stdout)
	}
	if strings.Contains(stdout, "witness's post") {
		t.Errorf("feed --author ember should not show witness's post: %s", stdout)
	}
}

func TestSmokeFeedOneline(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	h.SetIdentity("ember@testrig")
	h.Run("post", "test post")

	// Read feed in oneline format
	stdout, _, err := h.Run("feed", "--oneline")
	if err != nil {
		t.Fatalf("smoke feed failed: %v", err)
	}

	// Should start with post ID
	if !strings.HasPrefix(strings.TrimSpace(stdout), "smk-") {
		t.Errorf("feed --oneline should start with post ID: %s", stdout)
	}
}

func TestSmokeReply(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	h.SetIdentity("ember@testrig")

	// Post original
	stdout, _, _ := h.Run("post", "original post")

	// Extract post ID from output
	parts := strings.Fields(stdout)
	var postID string
	for _, p := range parts {
		if strings.HasPrefix(p, "smk-") {
			postID = p
			break
		}
	}
	if postID == "" {
		t.Fatal("Could not extract post ID from output")
	}

	h.SetIdentity("witness@testrig")

	// Reply
	stdout, _, err := h.Run("reply", postID, "nice post!")
	if err != nil {
		t.Fatalf("smoke reply failed: %v", err)
	}

	if !strings.Contains(stdout, "Replied") {
		t.Errorf("reply output missing confirmation: %s", stdout)
	}
	if !strings.Contains(stdout, postID) {
		t.Errorf("reply output missing parent ID: %s", stdout)
	}
}

func TestSmokeReplyInvalidID(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	h.SetIdentity("ember@testrig")

	// Reply to invalid ID
	_, stderr, err := h.Run("reply", "invalid-id", "test reply")
	if err == nil {
		t.Error("smoke reply should fail with invalid ID")
	}

	if !strings.Contains(stderr, "invalid post ID") {
		t.Errorf("Expected 'invalid post ID' error: %s", stderr)
	}
}

func TestSmokeReplyNonExistent(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	h.SetIdentity("ember@testrig")

	// Reply to non-existent ID
	_, stderr, err := h.Run("reply", "smk-notfnd", "test reply")
	if err == nil {
		t.Error("smoke reply should fail with non-existent ID")
	}

	if !strings.Contains(stderr, "not found") {
		t.Errorf("Expected 'not found' error: %s", stderr)
	}
}

func TestSmokeVersion(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	stdout, _, err := h.Run("--version")
	if err != nil {
		t.Fatalf("smoke --version failed: %v", err)
	}

	if !strings.Contains(stdout, "smoke version") {
		t.Errorf("version output: %s", stdout)
	}
}

func TestSmokeHelp(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	stdout, _, err := h.Run("--help")
	if err != nil {
		t.Fatalf("smoke --help failed: %v", err)
	}

	if !strings.Contains(stdout, "Social feed") {
		t.Errorf("help output missing description: %s", stdout)
	}
	if !strings.Contains(stdout, "init") {
		t.Errorf("help output missing init command: %s", stdout)
	}
	if !strings.Contains(stdout, "post") {
		t.Errorf("help output missing post command: %s", stdout)
	}
	if !strings.Contains(stdout, "feed") {
		t.Errorf("help output missing feed command: %s", stdout)
	}
	if !strings.Contains(stdout, "reply") {
		t.Errorf("help output missing reply command: %s", stdout)
	}
}

func TestSmokeFeedBoxDrawing(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Note: @testrig is ignored, project is auto-detected
	h.SetIdentity("ember@ignored-testrig")
	h.Run("post", "box drawing test")

	// Read feed
	stdout, _, err := h.Run("feed")
	if err != nil {
		t.Fatalf("smoke feed failed: %v", err)
	}

	// Check for compact format elements
	// Note: project is auto-detected from git/cwd, not from BD_ACTOR
	if !strings.Contains(stdout, "ember@") {
		t.Errorf("feed missing author@project (should be ember@{auto-detected}): %s", stdout)
	}
	if !strings.Contains(stdout, "box drawing test") {
		t.Errorf("feed missing post content: %s", stdout)
	}
}

func TestSmokeFeedReplyIndent(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	h.SetIdentity("ember@testrig")
	stdout, _, _ := h.Run("post", "parent post")

	// Extract post ID
	var postID string
	for _, p := range strings.Fields(stdout) {
		if strings.HasPrefix(p, "smk-") {
			postID = p
			break
		}
	}

	h.SetIdentity("witness@testrig")
	h.Run("reply", postID, "reply post")

	// Read feed
	stdout, _, err := h.Run("feed")
	if err != nil {
		t.Fatalf("smoke feed failed: %v", err)
	}

	// Replies should use tree-style indent
	if !strings.Contains(stdout, "└─") {
		t.Errorf("feed reply missing tree indent (└─): %s", stdout)
	}
}

// RunWithExitCode runs smoke and returns stdout, stderr, and exit code
func (h *TestHelper) RunWithExitCode(args ...string) (string, string, int) {
	h.t.Helper()

	if h.binPath == "" {
		h.t.Skip("smoke binary not found. Set SMOKE_BIN or build with 'make build'")
		return "", "", 0
	}

	cmd := exec.Command(h.binPath, args...)
	cmd.Dir = h.tmpDir

	env := os.Environ()
	env = append(env, "HOME="+h.tmpDir)
	cmd.Env = env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}
	return stdout.String(), stderr.String(), exitCode
}

func TestSmokeDoctorHealthy(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Run doctor
	stdout, _, exitCode := h.RunWithExitCode("doctor")

	// Should show all checks passing
	if !strings.Contains(stdout, "smoke doctor") {
		t.Errorf("doctor output missing header: %s", stdout)
	}
	if !strings.Contains(stdout, "INSTALLATION") {
		t.Errorf("doctor output missing INSTALLATION category: %s", stdout)
	}
	if !strings.Contains(stdout, "✓") {
		t.Errorf("doctor output missing pass indicators: %s", stdout)
	}

	// Exit code should be 0 for healthy installation
	if exitCode != 0 {
		t.Errorf("doctor exit code = %d, want 0 for healthy installation", exitCode)
	}
}

func TestSmokeDoctorMissingFeed(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Delete feed file
	feedPath := filepath.Join(h.configDir, "feed.jsonl")
	os.Remove(feedPath)

	// Run doctor
	stdout, _, exitCode := h.RunWithExitCode("doctor")

	// Should show error for missing feed
	if !strings.Contains(stdout, "✗") {
		t.Errorf("doctor output missing error indicator: %s", stdout)
	}
	if !strings.Contains(stdout, "not found") {
		t.Errorf("doctor output missing 'not found' message: %s", stdout)
	}

	// Exit code should be 2 for errors
	if exitCode != 2 {
		t.Errorf("doctor exit code = %d, want 2 for errors", exitCode)
	}
}

func TestSmokeDoctorFix(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Don't initialize - let doctor --fix do it
	// Run doctor --fix
	stdout, _, exitCode := h.RunWithExitCode("doctor", "--fix")

	// Should show fixes applied
	if !strings.Contains(stdout, "Fixed:") {
		t.Errorf("doctor --fix output missing 'Fixed:' message: %s", stdout)
	}
	if !strings.Contains(stdout, "Fixed") && !strings.Contains(stdout, "issue") {
		t.Errorf("doctor --fix output missing fix count: %s", stdout)
	}

	// After fix, exit code should be 0
	if exitCode != 0 {
		t.Errorf("doctor --fix exit code = %d, want 0 after fixes applied", exitCode)
	}

	// Verify files were created
	feedPath := filepath.Join(h.configDir, "feed.jsonl")
	if _, err := os.Stat(feedPath); os.IsNotExist(err) {
		t.Error("doctor --fix did not create feed.jsonl")
	}

	configPath := filepath.Join(h.configDir, "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("doctor --fix did not create config.yaml")
	}
}

func TestSmokeDoctorFixNoProblems(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Run doctor --fix on healthy installation
	stdout, _, exitCode := h.RunWithExitCode("doctor", "--fix")

	// Should show no problems to fix
	if !strings.Contains(stdout, "No problems to fix") {
		t.Errorf("doctor --fix output missing 'No problems to fix' message: %s", stdout)
	}

	// Exit code should be 0
	if exitCode != 0 {
		t.Errorf("doctor --fix exit code = %d, want 0", exitCode)
	}
}

func TestSmokeDoctorDryRun(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Don't initialize - run doctor --fix --dry-run
	stdout, _, exitCode := h.RunWithExitCode("doctor", "--fix", "--dry-run")

	// Should show what would be fixed
	if !strings.Contains(stdout, "Would fix:") {
		t.Errorf("doctor --fix --dry-run output missing 'Would fix:' message: %s", stdout)
	}
	if !strings.Contains(stdout, "would be fixed") {
		t.Errorf("doctor --fix --dry-run output missing summary: %s", stdout)
	}

	// Exit code should be 2 (problems still exist)
	if exitCode != 2 {
		t.Errorf("doctor --fix --dry-run exit code = %d, want 2 (problems remain)", exitCode)
	}
}

func TestSmokeDoctorDryRunNoModify(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Run doctor --fix --dry-run
	h.RunWithExitCode("doctor", "--fix", "--dry-run")

	// Verify files were NOT created
	if _, err := os.Stat(h.configDir); !os.IsNotExist(err) {
		t.Error("doctor --fix --dry-run should not create config directory")
	}

	feedPath := filepath.Join(h.configDir, "feed.jsonl")
	if _, err := os.Stat(feedPath); !os.IsNotExist(err) {
		t.Error("doctor --fix --dry-run should not create feed.jsonl")
	}
}
