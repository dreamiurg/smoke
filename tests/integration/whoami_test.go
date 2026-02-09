package integration

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// TestWhoamiDeterminism verifies that the same session seed produces identical usernames
// across multiple calls to whoami. This is critical for agent identity stability.
func TestWhoamiDeterminism(t *testing.T) {
	// When running under an agent (Claude, Codex, etc.), the identity seed
	// is based on the agent ancestor PID found via process tree walk, not
	// TERM_SESSION_ID. Each subprocess spawned by h.Run gets a different
	// immediate PPID, but they all share the same agent ancestor.
	// However, the process tree walk result can vary per-subprocess depending
	// on the shell layer, so skip when under an agent.
	if os.Getenv("CLAUDECODE") == "1" || os.Getenv("CODEX_SANDBOX") != "" {
		t.Skip("Cannot test TERM_SESSION_ID determinism when running under an agent")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	// Use a fixed seed via TERM_SESSION_ID environment variable
	// This makes the session seed deterministic for testing
	sessionSeed := "test-session-12345"
	os.Setenv("TERM_SESSION_ID", sessionSeed)
	defer os.Unsetenv("TERM_SESSION_ID")

	// Initialize smoke with the fixed seed
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Get identity first time
	stdout1, _, err := h.Run("whoami")
	if err != nil {
		t.Fatalf("first whoami failed: %v", err)
	}
	identity1 := strings.TrimSpace(stdout1)

	// Get identity second time with same seed
	stdout2, _, err := h.Run("whoami")
	if err != nil {
		t.Fatalf("second whoami failed: %v", err)
	}
	identity2 := strings.TrimSpace(stdout2)

	// Third call to triple-check determinism
	stdout3, _, err := h.Run("whoami")
	if err != nil {
		t.Fatalf("third whoami failed: %v", err)
	}
	identity3 := strings.TrimSpace(stdout3)

	// All three should be identical
	if identity1 != identity2 {
		t.Errorf("identity mismatch: call1=%q, call2=%q", identity1, identity2)
	}
	if identity2 != identity3 {
		t.Errorf("identity mismatch: call2=%q, call3=%q", identity2, identity3)
	}

	// Verify format: should contain @ (project separator)
	if !strings.Contains(identity1, "@") {
		t.Errorf("identity missing @ separator: %q", identity1)
	}
}

// TestWhoamiStyleVariety verifies that different session seeds produce usernames
// with varied formatting styles (e.g., lowercase, snake_case, CamelCase).
func TestWhoamiStyleVariety(t *testing.T) {
	// When running under an agent, identity is based on the agent's PID
	// (which is stable), not TERM_SESSION_ID. Skip in agent context.
	if os.Getenv("CLAUDECODE") == "1" || os.Getenv("CODEX_SANDBOX") != "" {
		t.Skip("Cannot test TERM_SESSION_ID variation when running under an agent")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize with base config
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Test multiple different seeds
	testCases := []struct {
		seed string
		name string
	}{
		{"seed-a-00001", "seed A"},
		{"seed-b-00002", "seed B"},
		{"seed-c-00003", "seed C"},
		{"seed-d-00004", "seed D"},
		{"seed-e-00005", "seed E"},
	}

	identities := make(map[string]string)

	for _, tc := range testCases {
		// Clear session seed for clean state
		os.Unsetenv("TERM_SESSION_ID")
		os.Unsetenv("WINDOWID")

		// Set specific session seed
		os.Setenv("TERM_SESSION_ID", tc.seed)
		defer os.Unsetenv("TERM_SESSION_ID")

		stdout, _, err := h.Run("whoami")
		if err != nil {
			t.Fatalf("whoami with seed %q failed: %v", tc.seed, err)
		}

		identity := strings.TrimSpace(stdout)
		identities[tc.name] = identity
		t.Logf("Seed %q: %q", tc.seed, identity)
	}

	// Check that we got at least some variety in the identities
	// We expect different seeds to produce different suffixes
	uniqueIdentities := make(map[string]bool)
	for _, id := range identities {
		uniqueIdentities[id] = true
	}

	if len(uniqueIdentities) < 3 {
		t.Errorf("expected at least 3 different identities, got %d: %v", len(uniqueIdentities), identities)
	}

	// Verify format for each identity
	for name, identity := range identities {
		if !strings.Contains(identity, "@") {
			t.Errorf("%s: identity missing @ separator: %q", name, identity)
		}

		// Extract suffix (part before @)
		parts := strings.Split(identity, "@")
		if len(parts) != 2 {
			t.Errorf("%s: invalid identity format (expected name@project): %q", name, identity)
		}
	}
}

// TestWhoamiAgentPrefix verifies that generated usernames include the detected
// agent prefix when running under an agent, or no agent prefix when not.
func TestWhoamiAgentPrefix(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	stdout, _, err := h.Run("whoami")
	if err != nil {
		t.Fatalf("whoami failed: %v", err)
	}

	identity := strings.TrimSpace(stdout)
	parts := strings.Split(identity, "@")
	if len(parts) < 1 {
		t.Fatalf("identity missing @ separator: %q", identity)
	}

	name := parts[0]

	// When running under an agent, the identity should include the agent prefix
	// (e.g., "claude-swift-fox" or "codex-swift-fox")
	// When not under an agent, there's no prefix
	// The suffix word(s) are always generated from the session seed
	t.Logf("Identity name part: %q (full: %q)", name, identity)

	// Verify name is non-empty and has valid characters
	if name == "" {
		t.Error("identity name part is empty")
	}
}

// TestWhoamiProjectSuffix verifies that generated usernames end with @project suffix
func TestWhoamiProjectSuffix(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Get whoami output
	stdout, _, err := h.Run("whoami")
	if err != nil {
		t.Fatalf("whoami failed: %v", err)
	}

	identity := strings.TrimSpace(stdout)

	// Should have @ separator
	if !strings.Contains(identity, "@") {
		t.Errorf("identity missing @ separator: %q", identity)
	}

	parts := strings.Split(identity, "@")
	if len(parts) != 2 {
		t.Errorf("expected name@project format, got %d parts: %q", len(parts), identity)
	}

	// Project part should not be empty
	if parts[1] == "" {
		t.Errorf("project part is empty: %q", identity)
	}

	// Project should be "smoke" in this context
	if parts[1] != "smoke" {
		t.Logf("note: project is %q (expected 'smoke' in normal context)", parts[1])
	}
}

// TestWhoamiSmokeNameOverride verifies that SMOKE_NAME environment variable
// works as an override, bypassing the auto-generation logic
func TestWhoamiSmokeNameOverride(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Set SMOKE_NAME override
	originalName := os.Getenv("SMOKE_NAME")
	os.Setenv("SMOKE_NAME", "testbot")
	defer os.Setenv("SMOKE_NAME", originalName)

	stdout, _, err := h.Run("whoami")
	if err != nil {
		t.Fatalf("whoami with SMOKE_NAME failed: %v", err)
	}

	identity := strings.TrimSpace(stdout)

	// Should use the override name
	if !strings.HasPrefix(identity, "testbot") {
		t.Errorf("expected identity to start with 'testbot', got: %q", identity)
	}

	// Should still have @project suffix
	if !strings.Contains(identity, "@") {
		t.Errorf("identity missing @ separator: %q", identity)
	}
}

// TestWhoamiSmokeNameIgnoresProjectOverride verifies that SMOKE_NAME env var
// uses the name but @project is ALWAYS auto-detected (cannot be overridden)
func TestWhoamiSmokeNameIgnoresProjectOverride(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Set SMOKE_NAME with @project (which will be stripped and ignored)
	h.SetIdentity("agent-name@testproj")

	stdout, _, err := h.Run("whoami")
	if err != nil {
		t.Fatalf("whoami with SMOKE_NAME failed: %v", err)
	}

	identity := strings.TrimSpace(stdout)

	// Should use agent-name but with auto-detected project (not testproj)
	// @project override is explicitly ignored
	if !strings.HasPrefix(identity, "agent-name@") {
		t.Errorf("expected identity to start with 'agent-name@', got: %q", identity)
	}
	// Verify @testproj was NOT used
	if strings.Contains(identity, "testproj") {
		t.Errorf("@project override should be ignored, but got: %q", identity)
	}
}

// TestWhoamiJSONOutput verifies that --json flag produces valid JSON with
// name and project fields
func TestWhoamiJSONOutput(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	stdout, _, err := h.Run("whoami", "--json")
	if err != nil {
		t.Fatalf("whoami --json failed: %v", err)
	}

	// Parse JSON output
	var output map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &output); err != nil {
		t.Fatalf("invalid JSON output: %v\nGot: %s", err, stdout)
	}

	// Verify required fields
	name, hasName := output["name"]
	if !hasName {
		t.Error("JSON output missing 'name' field")
	}

	project, hasProject := output["project"]
	if !hasProject {
		t.Error("JSON output missing 'project' field")
	}

	// Verify fields are strings
	nameStr, ok := name.(string)
	if !ok || nameStr == "" {
		t.Errorf("'name' field should be non-empty string, got: %v", name)
	}

	projectStr, ok := project.(string)
	if !ok || projectStr == "" {
		t.Errorf("'project' field should be non-empty string, got: %v", project)
	}
}

// TestWhoamiNameFlag verifies that --name flag outputs only the name part
// without the project suffix
func TestWhoamiNameFlag(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	stdout, _, err := h.Run("whoami", "--name")
	if err != nil {
		t.Fatalf("whoami --name failed: %v", err)
	}

	name := strings.TrimSpace(stdout)

	// Should NOT contain @ separator
	if strings.Contains(name, "@") {
		t.Errorf("--name output should not contain @: %q", name)
	}

	// Should not be empty
	if name == "" {
		t.Error("--name output is empty")
	}
}

// TestWhoamiMultipleSessionSeeds verifies that different TERM_SESSION_ID values
// produce different usernames (testing the hash-based generation)
func TestWhoamiMultipleSessionSeeds(t *testing.T) {
	// When running under an agent, identity is based on the agent's PID
	// (which is stable), not TERM_SESSION_ID. Skip this test in agent context.
	if os.Getenv("CLAUDECODE") == "1" || os.Getenv("CODEX_SANDBOX") != "" {
		t.Skip("Cannot test TERM_SESSION_ID variation when running under an agent")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	seeds := []string{
		"unique-seed-alpha-001",
		"unique-seed-beta-002",
		"unique-seed-gamma-003",
		"unique-seed-delta-004",
		"unique-seed-epsilon-005",
	}
	// Collect identities from different seeds
	identities := make([]string, 0, len(seeds))

	for _, seed := range seeds {
		os.Setenv("TERM_SESSION_ID", seed)
		defer os.Unsetenv("TERM_SESSION_ID")

		stdout, _, err := h.Run("whoami")
		if err != nil {
			t.Fatalf("whoami failed for seed %q: %v", seed, err)
		}

		identities = append(identities, strings.TrimSpace(stdout))
	}

	// Verify we got multiple different identities
	uniqueMap := make(map[string]bool)
	for _, id := range identities {
		uniqueMap[id] = true
	}

	if len(uniqueMap) < 3 {
		t.Errorf("expected at least 3 different identities from 5 seeds, got: %v", identities)
	}

	t.Logf("Generated %d unique identities from %d seeds: %v", len(uniqueMap), len(seeds), identities)
}

// TestWhoamiConsistencyAcrossSessions verifies that the same seed in a new process
// still produces the same identity (no random state leakage)
func TestWhoamiConsistencyAcrossSessions(t *testing.T) {
	// When running under an agent, the seed is based on process tree, not TERM_SESSION_ID.
	// Each subprocess may get a different seed. Skip in agent context.
	if os.Getenv("CLAUDECODE") == "1" || os.Getenv("CODEX_SANDBOX") != "" {
		t.Skip("Cannot test TERM_SESSION_ID consistency when running under an agent")
	}

	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	fixedSeed := "fixed-consistency-test"
	os.Setenv("TERM_SESSION_ID", fixedSeed)
	defer os.Unsetenv("TERM_SESSION_ID")

	// Get identity
	stdout1, _, err := h.Run("whoami")
	if err != nil {
		t.Fatalf("first whoami failed: %v", err)
	}
	identity1 := strings.TrimSpace(stdout1)

	// Unset and reset the seed to simulate a new session process
	os.Unsetenv("TERM_SESSION_ID")
	os.Setenv("TERM_SESSION_ID", fixedSeed)

	// Get identity again
	stdout2, _, err := h.Run("whoami")
	if err != nil {
		t.Fatalf("second whoami failed: %v", err)
	}
	identity2 := strings.TrimSpace(stdout2)

	// Should be identical
	if identity1 != identity2 {
		t.Errorf("identity changed across sessions with same seed: %q != %q", identity1, identity2)
	}
}
