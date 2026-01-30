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
	gasTownRoot string
	binPath     string
	origDir     string
	origBDActor string
}

func NewTestHelper(t *testing.T) *TestHelper {
	t.Helper()

	tmpDir := t.TempDir()
	gasTownRoot := filepath.Join(tmpDir, "testtown")
	beadsDir := filepath.Join(gasTownRoot, ".beads")

	if err := os.MkdirAll(beadsDir, 0755); err != nil {
		t.Fatalf("Failed to create test Gas Town: %v", err)
	}

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

	return &TestHelper{
		t:           t,
		tmpDir:      tmpDir,
		gasTownRoot: gasTownRoot,
		binPath:     binPath,
		origDir:     origDir,
		origBDActor: origBDActor,
	}
}

func (h *TestHelper) Cleanup() {
	os.Chdir(h.origDir)
	os.Setenv("BD_ACTOR", h.origBDActor)
}

func (h *TestHelper) ChDir() {
	if err := os.Chdir(h.gasTownRoot); err != nil {
		h.t.Fatalf("Failed to chdir: %v", err)
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
	cmd.Dir = h.gasTownRoot
	cmd.Env = os.Environ()

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

	// Check .smoke directory created
	smokeDir := filepath.Join(h.gasTownRoot, ".smoke")
	if _, err := os.Stat(smokeDir); os.IsNotExist(err) {
		t.Error(".smoke directory not created")
	}

	// Check feed.jsonl created
	feedFile := filepath.Join(smokeDir, "feed.jsonl")
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

func TestSmokeInitNotGasTown(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	if h.binPath == "" {
		t.Skip("smoke binary not found. Set SMOKE_BIN or build with 'make build'")
	}

	// Create a non-Gas Town directory
	notGasTown := filepath.Join(h.tmpDir, "notgastown")
	os.MkdirAll(notGasTown, 0755)

	// Try to init in non-Gas Town
	cmd := exec.Command(h.binPath, "init")
	cmd.Dir = notGasTown
	cmd.Env = os.Environ()

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Error("smoke init should fail in non-Gas Town directory")
	}

	if !strings.Contains(stderr.String(), "not in a Gas Town") {
		t.Errorf("Expected 'not in a Gas Town' error: %s", stderr.String())
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
	h.SetIdentity("testtown/crew/testuser")

	// Post a message
	stdout, _, err := h.Run("post", "hello from integration test")
	if err != nil {
		t.Fatalf("smoke post failed: %v", err)
	}

	if !strings.Contains(stdout, "Posted smk-") {
		t.Errorf("post output missing confirmation: %s", stdout)
	}

	// Verify post in feed
	feedFile := filepath.Join(h.gasTownRoot, ".smoke", "feed.jsonl")
	content, _ := os.ReadFile(feedFile)
	if !strings.Contains(string(content), "hello from integration test") {
		t.Errorf("Post not found in feed file: %s", content)
	}
}

func TestSmokePostNoIdentity(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Clear identity
	os.Unsetenv("BD_ACTOR")
	os.Unsetenv("SMOKE_AUTHOR")

	// Post should fail
	_, stderr, err := h.Run("post", "test message")
	if err == nil {
		t.Error("smoke post should fail without identity")
	}

	if !strings.Contains(stderr, "cannot determine identity") {
		t.Errorf("Expected identity error: %s", stderr)
	}
}

func TestSmokePostTooLong(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	h.SetIdentity("testtown/crew/testuser")

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

	h.SetIdentity("testtown/crew/testuser")

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

	h.SetIdentity("testtown/crew/ember")

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
	if !strings.Contains(stdout, "ember@testtown") {
		t.Errorf("feed missing author: %s", stdout)
	}
}

func TestSmokeFeedLimit(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	h.SetIdentity("testtown/crew/ember")

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
	count := strings.Count(stdout, "ember@testtown")
	if count != 2 {
		t.Errorf("feed -n 2 returned %d posts, want 2", count)
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
	h.SetIdentity("testtown/crew/ember")
	h.Run("post", "ember's post")

	h.SetIdentity("testtown/crew/witness")
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

	h.SetIdentity("testtown/crew/ember")
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

	h.SetIdentity("testtown/crew/ember")

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

	h.SetIdentity("testtown/crew/witness")

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

	h.SetIdentity("testtown/crew/ember")

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

	h.SetIdentity("testtown/crew/ember")

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

	if !strings.Contains(stdout, "Internal social feed") {
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

	h.SetIdentity("testtown/crew/ember")
	h.Run("post", "box drawing test")

	// Read feed
	stdout, _, err := h.Run("feed")
	if err != nil {
		t.Fatalf("smoke feed failed: %v", err)
	}

	// Check for box-drawing characters
	if !strings.Contains(stdout, "╭") {
		t.Errorf("feed missing top-left corner (╭): %s", stdout)
	}
	if !strings.Contains(stdout, "╯") {
		t.Errorf("feed missing bottom-right corner (╯): %s", stdout)
	}
	if !strings.Contains(stdout, "│") {
		t.Errorf("feed missing vertical border (│): %s", stdout)
	}
}

func TestSmokeFeedReplyIndent(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	h.SetIdentity("testtown/crew/ember")
	stdout, _, _ := h.Run("post", "parent post")

	// Extract post ID
	var postID string
	for _, p := range strings.Fields(stdout) {
		if strings.HasPrefix(p, "smk-") {
			postID = p
			break
		}
	}

	h.SetIdentity("testtown/crew/witness")
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
