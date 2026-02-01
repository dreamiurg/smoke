package identity

import (
	"fmt"
	"strings"
)

// Lowercase formats words as a single lowercase string.
// Example: ["telescoped", "seeker"] -> "telescopedseeker"
func Lowercase(words []string) string {
	combined := strings.Join(words, "")
	return strings.ToLower(combined)
}

// SnakeCase formats words with underscores between them, all lowercase.
// Example: ["quantum", "seeker"] -> "quantum_seeker"
func SnakeCase(words []string) string {
	lower := make([]string, 0, len(words))
	for _, w := range words {
		if w != "" {
			lower = append(lower, strings.ToLower(w))
		}
	}
	return strings.Join(lower, "_")
}

// CamelCase formats words with each word capitalized, no separators.
// Example: ["quantum", "seeker"] -> "QuantumSeeker"
func CamelCase(words []string) string {
	if len(words) == 0 {
		return ""
	}
	result := ""
	for _, w := range words {
		if len(w) > 0 {
			result += strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
		}
	}
	return result
}

// LowerCamel formats words with first word lowercase, rest capitalized, no separators.
// Example: ["quantum", "seeker"] -> "quantumSeeker"
func LowerCamel(words []string) string {
	if len(words) == 0 {
		return ""
	}
	// Find first non-empty word
	firstIdx := -1
	for i, w := range words {
		if w != "" {
			firstIdx = i
			break
		}
	}
	if firstIdx == -1 {
		return ""
	}
	result := strings.ToLower(words[firstIdx])
	for i := firstIdx + 1; i < len(words); i++ {
		w := words[i]
		if len(w) > 0 {
			result += strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
		}
	}
	return result
}

// KebabCase formats words with hyphens between them, all lowercase.
// Example: ["quantum", "seeker"] -> "quantum-seeker"
func KebabCase(words []string) string {
	lower := make([]string, 0, len(words))
	for _, w := range words {
		if w != "" {
			lower = append(lower, strings.ToLower(w))
		}
	}
	return strings.Join(lower, "-")
}

// WithNumber formats words in CamelCase and appends a single digit (0-9).
// The digit is deterministic based on word content.
// Example: ["quantum", "seeker"] -> "QuantumSeeker7"
func WithNumber(words []string) string {
	camel := CamelCase(words)
	// Calculate a deterministic number from the words
	hash := 0
	for _, w := range words {
		for _, ch := range w {
			hash += int(ch)
		}
	}
	digit := hash % 10
	return fmt.Sprintf("%s%d", camel, digit)
}
