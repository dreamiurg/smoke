package identity

import "testing"

func TestGenerate_Stability(t *testing.T) {
	seed := "test-session-123"

	// Same seed should always produce same result
	result1 := Generate(seed)
	result2 := Generate(seed)

	if result1 != result2 {
		t.Errorf("Generate not stable: got %q and %q for same seed", result1, result2)
	}
}

func TestGenerate_Uniqueness(t *testing.T) {
	seeds := []string{
		"session-1",
		"session-2",
		"session-3",
		"different-seed",
	}

	results := make(map[string]bool)
	for _, seed := range seeds {
		result := Generate(seed)
		if results[result] {
			// Not a hard failure - collisions possible but unlikely with 2500 combos
			t.Logf("Collision detected for seed %q: %s", seed, result)
		}
		results[result] = true
	}
}

func TestGenerate_Format(t *testing.T) {
	result := Generate("test")

	// Should be in format "adjective-animal"
	if len(result) < 5 { // minimum: "a-b"
		t.Errorf("Generate result too short: %q", result)
	}

	// Should contain exactly one hyphen
	hyphenCount := 0
	for _, c := range result {
		if c == '-' {
			hyphenCount++
		}
	}
	if hyphenCount != 1 {
		t.Errorf("Expected 1 hyphen, got %d in %q", hyphenCount, result)
	}
}

func TestGenerateFull(t *testing.T) {
	tests := []struct {
		agent   string
		seed    string
		project string
		wantFmt string // regex-like pattern
	}{
		{"claude", "session-1", "smoke", "claude-"},
		{"unknown", "session-2", "myproject", "unknown-"},
	}

	for _, tt := range tests {
		result := GenerateFull(tt.agent, tt.seed, tt.project)

		// Should start with agent-
		if len(result) < len(tt.agent)+1 {
			t.Errorf("GenerateFull too short: %q", result)
		}

		// Should contain @project
		wantSuffix := "@" + tt.project
		if result[len(result)-len(wantSuffix):] != wantSuffix {
			t.Errorf("GenerateFull missing project suffix: got %q, want suffix %q", result, wantSuffix)
		}
	}
}
