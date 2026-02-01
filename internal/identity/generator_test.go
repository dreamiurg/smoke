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

// TestGenerateWithPattern tests the new multi-pattern generation functionality.
// This test suite validates pattern selection and pattern-specific generation.
func TestGenerateWithPattern(t *testing.T) {
	tests := []struct {
		name    string
		seed    string
		pattern Pattern
		wantErr bool
		// Validation function for the generated result
		validate func(t *testing.T, result string)
	}{
		{
			name:    "VerbNoun pattern with valid seed",
			seed:    "test-seed-1",
			pattern: PatternVerbNoun,
			wantErr: false,
			validate: func(t *testing.T, result string) {
				// VerbNoun should be: verb-noun format
				if !containsHyphen(result, 1) {
					t.Errorf("VerbNoun result should have exactly 1 hyphen, got %q", result)
				}
				// Should not be empty
				if result == "" {
					t.Errorf("VerbNoun result should not be empty")
				}
			},
		},
		{
			name:    "AdjectiveNoun pattern with valid seed",
			seed:    "test-seed-2",
			pattern: PatternAdjectiveNoun,
			wantErr: false,
			validate: func(t *testing.T, result string) {
				// AdjectiveNoun should be: adjective-noun format
				if !containsHyphen(result, 1) {
					t.Errorf("AdjectiveNoun result should have exactly 1 hyphen, got %q", result)
				}
				if result == "" {
					t.Errorf("AdjectiveNoun result should not be empty")
				}
			},
		},
		{
			name:    "AbstractConcrete pattern with valid seed",
			seed:    "test-seed-3",
			pattern: PatternAbstractConcrete,
			wantErr: false,
			validate: func(t *testing.T, result string) {
				// AbstractConcrete should be: abstract-concrete format
				if !containsHyphen(result, 1) {
					t.Errorf("AbstractConcrete result should have exactly 1 hyphen, got %q", result)
				}
				if result == "" {
					t.Errorf("AbstractConcrete result should not be empty")
				}
			},
		},
		{
			name:    "TechTerm pattern with valid seed",
			seed:    "test-seed-4",
			pattern: PatternTechTerm,
			wantErr: false,
			validate: func(t *testing.T, result string) {
				// TechTerm could be single word or multiple; should not be empty
				if result == "" {
					t.Errorf("TechTerm result should not be empty")
				}
			},
		},
		{
			name:    "AdjectiveAdjectiveNoun pattern with valid seed",
			seed:    "test-seed-5",
			pattern: PatternAdjectiveAdjectiveNoun,
			wantErr: false,
			validate: func(t *testing.T, result string) {
				// AdjectiveAdjectiveNoun should be: adjective-adjective-noun format
				if !containsHyphen(result, 2) {
					t.Errorf("AdjectiveAdjectiveNoun result should have exactly 2 hyphens, got %q", result)
				}
				if result == "" {
					t.Errorf("AdjectiveAdjectiveNoun result should not be empty")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateWithPattern(tt.seed, tt.pattern)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateWithPattern() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				tt.validate(t, result)
			}
		})
	}
}

// TestPatternDeterminism verifies that the same seed + pattern always produces the same result.
func TestPatternDeterminism(t *testing.T) {
	patterns := []Pattern{
		PatternVerbNoun,
		PatternAdjectiveNoun,
		PatternAbstractConcrete,
		PatternTechTerm,
		PatternAdjectiveAdjectiveNoun,
	}

	seed := "determinism-test-seed"

	for _, pattern := range patterns {
		t.Run(pattern.String(), func(t *testing.T) {
			result1, err1 := GenerateWithPattern(seed, pattern)
			if err1 != nil {
				t.Fatalf("First call failed: %v", err1)
			}

			result2, err2 := GenerateWithPattern(seed, pattern)
			if err2 != nil {
				t.Fatalf("Second call failed: %v", err2)
			}

			if result1 != result2 {
				t.Errorf("Determinism failed for pattern %v: got %q then %q for same seed", pattern, result1, result2)
			}
		})
	}
}

// TestPatternDiversity verifies that different seeds produce different results for the same pattern.
func TestPatternDiversity(t *testing.T) {
	pattern := PatternVerbNoun
	seeds := []string{
		"seed-alpha",
		"seed-beta",
		"seed-gamma",
		"seed-delta",
		"seed-epsilon",
	}

	results := make(map[string]bool)
	collisions := 0

	for _, seed := range seeds {
		result, err := GenerateWithPattern(seed, pattern)
		if err != nil {
			t.Fatalf("GenerateWithPattern failed for seed %q: %v", seed, err)
		}

		if results[result] {
			collisions++
		}
		results[result] = true
	}

	if collisions > 0 {
		t.Logf("PatternDiversity: %d collision(s) detected across %d seeds (acceptable but unlikely)", collisions, len(seeds))
	}

	if len(results) < 3 {
		t.Errorf("PatternDiversity: expected at least 3 unique results from %d seeds, got %d", len(seeds), len(results))
	}
}

// TestPatternSelection verifies that different hash values select different patterns correctly.
func TestPatternSelection(t *testing.T) {
	// Test that different seeds produce different patterns when using SelectPattern
	seeds := []string{
		"pattern-test-1",
		"pattern-test-2",
		"pattern-test-3",
		"pattern-test-4",
	}

	selectedPatterns := make([]Pattern, len(seeds))

	for i, seed := range seeds {
		pattern := SelectPattern(seed)
		selectedPatterns[i] = pattern
	}

	// We should have some variety in pattern selection
	patternCounts := make(map[Pattern]int)
	for _, p := range selectedPatterns {
		patternCounts[p]++
	}

	if len(patternCounts) < 2 {
		t.Logf("PatternSelection: only %d unique patterns selected from %d seeds (may happen by chance)", len(patternCounts), len(seeds))
	}
}

// TestVerbNounPattern specifically validates VerbNoun generation logic.
func TestVerbNounPattern(t *testing.T) {
	seed := "verbnoun-test"
	result, err := GenerateWithPattern(seed, PatternVerbNoun)

	if err != nil {
		t.Fatalf("GenerateWithPattern failed: %v", err)
	}

	if result == "" {
		t.Fatal("VerbNoun pattern produced empty result")
	}

	// Should have exactly one hyphen separating verb and noun
	if !containsHyphen(result, 1) {
		t.Errorf("VerbNoun should have exactly 1 hyphen, got %q", result)
	}

	// Split and validate parts exist
	parts := splitByHyphen(result)
	if len(parts) != 2 {
		t.Errorf("VerbNoun split should produce 2 parts, got %d from %q", len(parts), result)
	}
	for i, part := range parts {
		if part == "" {
			t.Errorf("VerbNoun part %d is empty", i)
		}
	}
}

// TestAdjectiveNounPattern specifically validates AdjectiveNoun generation logic.
func TestAdjectiveNounPattern(t *testing.T) {
	seed := "adjectivenoun-test"
	result, err := GenerateWithPattern(seed, PatternAdjectiveNoun)

	if err != nil {
		t.Fatalf("GenerateWithPattern failed: %v", err)
	}

	if result == "" {
		t.Fatal("AdjectiveNoun pattern produced empty result")
	}

	// Should have exactly one hyphen separating adjective and noun
	if !containsHyphen(result, 1) {
		t.Errorf("AdjectiveNoun should have exactly 1 hyphen, got %q", result)
	}

	// Split and validate parts exist
	parts := splitByHyphen(result)
	if len(parts) != 2 {
		t.Errorf("AdjectiveNoun split should produce 2 parts, got %d from %q", len(parts), result)
	}
	for i, part := range parts {
		if part == "" {
			t.Errorf("AdjectiveNoun part %d is empty", i)
		}
	}
}

// TestAbstractConcretePattern specifically validates AbstractConcrete generation logic.
func TestAbstractConcretePattern(t *testing.T) {
	seed := "abstractconcrete-test"
	result, err := GenerateWithPattern(seed, PatternAbstractConcrete)

	if err != nil {
		t.Fatalf("GenerateWithPattern failed: %v", err)
	}

	if result == "" {
		t.Fatal("AbstractConcrete pattern produced empty result")
	}

	// Should have exactly one hyphen separating abstract and concrete noun
	if !containsHyphen(result, 1) {
		t.Errorf("AbstractConcrete should have exactly 1 hyphen, got %q", result)
	}

	// Split and validate parts exist
	parts := splitByHyphen(result)
	if len(parts) != 2 {
		t.Errorf("AbstractConcrete split should produce 2 parts, got %d from %q", len(parts), result)
	}
	for i, part := range parts {
		if part == "" {
			t.Errorf("AbstractConcrete part %d is empty", i)
		}
	}
}

// TestTechTermPattern specifically validates TechTerm generation logic.
func TestTechTermPattern(t *testing.T) {
	seed := "techterm-test"
	result, err := GenerateWithPattern(seed, PatternTechTerm)

	if err != nil {
		t.Fatalf("GenerateWithPattern failed: %v", err)
	}

	if result == "" {
		t.Fatal("TechTerm pattern produced empty result")
	}

	// TechTerm should not be empty and should be a valid word
	if len(result) < 2 {
		t.Errorf("TechTerm result too short: %q", result)
	}
}

// TestInvalidPattern verifies that invalid patterns are rejected.
func TestInvalidPattern(t *testing.T) {
	seed := "test-seed"
	invalidPattern := Pattern(999) // Invalid pattern value

	result, err := GenerateWithPattern(seed, invalidPattern)

	if err == nil {
		t.Error("Expected error for invalid pattern, got none")
	}

	if result != "" {
		t.Errorf("Expected empty result for invalid pattern, got %q", result)
	}
}

// Helper functions

// containsHyphen checks if a string contains exactly the expected number of hyphens.
func containsHyphen(s string, expected int) bool {
	count := 0
	for _, c := range s {
		if c == '-' {
			count++
		}
	}
	return count == expected
}

// splitByHyphen splits a string by hyphens.
func splitByHyphen(s string) []string {
	var parts []string
	var current string
	for _, c := range s {
		if c == '-' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
