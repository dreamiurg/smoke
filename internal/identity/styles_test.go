package identity

import (
	"testing"
)

func TestLowercase(t *testing.T) {
	tests := []struct {
		name     string
		words    []string
		expected string
	}{
		{
			name:     "two words",
			words:    []string{"quantum", "seeker"},
			expected: "quantumseeker",
		},
		{
			name:     "single word",
			words:    []string{"telescoped"},
			expected: "telescoped",
		},
		{
			name:     "three words",
			words:    []string{"bright", "swift", "fox"},
			expected: "brightswiftfox",
		},
		{
			name:     "mixed case",
			words:    []string{"Quantum", "SEEKER"},
			expected: "quantumseeker",
		},
		{
			name:     "nil slice",
			words:    nil,
			expected: "",
		},
		{
			name:     "empty slice",
			words:    []string{},
			expected: "",
		},
		{
			name:     "slice with empty strings",
			words:    []string{"", ""},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Lowercase(tt.words)
			if got != tt.expected {
				t.Errorf("Lowercase(%v) = %s, want %s", tt.words, got, tt.expected)
			}
		})
	}
}

func TestSnakeCase(t *testing.T) {
	tests := []struct {
		name     string
		words    []string
		expected string
	}{
		{
			name:     "two words",
			words:    []string{"quantum", "seeker"},
			expected: "quantum_seeker",
		},
		{
			name:     "single word",
			words:    []string{"telescoped"},
			expected: "telescoped",
		},
		{
			name:     "three words",
			words:    []string{"bright", "swift", "fox"},
			expected: "bright_swift_fox",
		},
		{
			name:     "nil slice",
			words:    nil,
			expected: "",
		},
		{
			name:     "empty slice",
			words:    []string{},
			expected: "",
		},
		{
			name:     "slice with empty strings",
			words:    []string{"", ""},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SnakeCase(tt.words)
			if got != tt.expected {
				t.Errorf("SnakeCase(%v) = %s, want %s", tt.words, got, tt.expected)
			}
		})
	}
}

func TestCamelCase(t *testing.T) {
	tests := []struct {
		name     string
		words    []string
		expected string
	}{
		{
			name:     "two words",
			words:    []string{"quantum", "seeker"},
			expected: "QuantumSeeker",
		},
		{
			name:     "single word",
			words:    []string{"telescoped"},
			expected: "Telescoped",
		},
		{
			name:     "three words",
			words:    []string{"bright", "swift", "fox"},
			expected: "BrightSwiftFox",
		},
		{
			name:     "nil slice",
			words:    nil,
			expected: "",
		},
		{
			name:     "empty slice",
			words:    []string{},
			expected: "",
		},
		{
			name:     "slice with empty strings",
			words:    []string{"", ""},
			expected: "",
		},
		{
			name:     "slice with mixed empty and non-empty",
			words:    []string{"quantum", "", "seeker"},
			expected: "QuantumSeeker",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CamelCase(tt.words)
			if got != tt.expected {
				t.Errorf("CamelCase(%v) = %s, want %s", tt.words, got, tt.expected)
			}
		})
	}
}

func TestLowerCamel(t *testing.T) {
	tests := []struct {
		name     string
		words    []string
		expected string
	}{
		{
			name:     "two words",
			words:    []string{"quantum", "seeker"},
			expected: "quantumSeeker",
		},
		{
			name:     "single word",
			words:    []string{"telescoped"},
			expected: "telescoped",
		},
		{
			name:     "three words",
			words:    []string{"bright", "swift", "fox"},
			expected: "brightSwiftFox",
		},
		{
			name:     "nil slice",
			words:    nil,
			expected: "",
		},
		{
			name:     "empty slice",
			words:    []string{},
			expected: "",
		},
		{
			name:     "slice with empty strings",
			words:    []string{"", ""},
			expected: "",
		},
		{
			name:     "slice with mixed empty and non-empty",
			words:    []string{"quantum", "", "seeker"},
			expected: "quantumSeeker",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LowerCamel(tt.words)
			if got != tt.expected {
				t.Errorf("LowerCamel(%v) = %s, want %s", tt.words, got, tt.expected)
			}
		})
	}
}

func TestKebabCase(t *testing.T) {
	tests := []struct {
		name     string
		words    []string
		expected string
	}{
		{
			name:     "two words",
			words:    []string{"quantum", "seeker"},
			expected: "quantum-seeker",
		},
		{
			name:     "single word",
			words:    []string{"telescoped"},
			expected: "telescoped",
		},
		{
			name:     "three words",
			words:    []string{"bright", "swift", "fox"},
			expected: "bright-swift-fox",
		},
		{
			name:     "nil slice",
			words:    nil,
			expected: "",
		},
		{
			name:     "empty slice",
			words:    []string{},
			expected: "",
		},
		{
			name:     "slice with empty strings",
			words:    []string{"", ""},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := KebabCase(tt.words)
			if got != tt.expected {
				t.Errorf("KebabCase(%v) = %s, want %s", tt.words, got, tt.expected)
			}
		})
	}
}

func TestWithNumber(t *testing.T) {
	tests := []struct {
		name     string
		words    []string
		hasDigit bool
	}{
		{
			name:     "two words",
			words:    []string{"quantum", "seeker"},
			hasDigit: true,
		},
		{
			name:     "single word",
			words:    []string{"telescoped"},
			hasDigit: true,
		},
		{
			name:     "three words",
			words:    []string{"bright", "swift", "fox"},
			hasDigit: true,
		},
		{
			name:     "mixed case",
			words:    []string{"Quantum", "SEEKER"},
			hasDigit: true,
		},
		{
			name:     "single uppercase word",
			words:    []string{"A"},
			hasDigit: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WithNumber(tt.words)
			if len(got) == 0 {
				t.Errorf("WithNumber(%v) returned empty string", tt.words)
			}
			// Verify it ends with a digit
			lastChar := got[len(got)-1]
			if lastChar < '0' || lastChar > '9' {
				t.Errorf("WithNumber(%v) = %s, does not end with digit", tt.words, got)
			}
			// Verify the result starts with CamelCase prefix (at least 2 chars for digit)
			if len(got) < 2 {
				t.Errorf("WithNumber(%v) = %s, too short", tt.words, got)
			}
		})
	}

	// Test determinism
	words := []string{"quantum", "seeker"}
	result1 := WithNumber(words)
	result2 := WithNumber(words)
	if result1 != result2 {
		t.Errorf("WithNumber() not deterministic: %s != %s", result1, result2)
	}
}

func TestWithNumberDeterminism(t *testing.T) {
	// Verify WithNumber produces consistent output for same input
	words := []string{"bright", "swift", "fox"}
	results := make(map[string]int)

	for i := 0; i < 5; i++ {
		result := WithNumber(words)
		results[result]++
	}

	// Should have only 1 unique result after 5 calls
	if len(results) != 1 {
		t.Errorf("WithNumber() not deterministic, got %d different results", len(results))
	}
}
