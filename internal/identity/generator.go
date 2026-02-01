package identity

import (
	"fmt"
	"hash/fnv"
)

// Pattern represents a naming pattern for identity generation.
type Pattern int

const (
	_ Pattern = iota
	PatternVerbNoun
	PatternAdjectiveNoun
	PatternAbstractConcrete
	PatternTechTerm
)

// String returns the string representation of a Pattern.
func (p Pattern) String() string {
	switch p {
	case PatternVerbNoun:
		return "VerbNoun"
	case PatternAdjectiveNoun:
		return "AdjectiveNoun"
	case PatternAbstractConcrete:
		return "AbstractConcrete"
	case PatternTechTerm:
		return "TechTerm"
	default:
		return "Unknown"
	}
}

// Generate creates an adjective-animal identity suffix from a seed string.
// The same seed will always produce the same identity.
func Generate(seed string) string {
	h := fnv.New32a()
	h.Write([]byte(seed))
	hash := h.Sum32()

	adjIdx := hash % uint32(len(Adjectives))
	animalIdx := (hash / uint32(len(Adjectives))) % uint32(len(Animals))

	return fmt.Sprintf("%s-%s", Adjectives[adjIdx], Animals[animalIdx])
}

// GenerateFull creates a full identity string in the format: agent-adjective-animal@project
func GenerateFull(agent, seed, project string) string {
	suffix := Generate(seed)
	return fmt.Sprintf("%s-%s@%s", agent, suffix, project)
}

// SelectPattern selects a naming pattern based on a seed hash.
// Different seeds should produce different (though not necessarily unique) pattern selections.
func SelectPattern(seed string) Pattern {
	h := fnv.New32a()
	h.Write([]byte(seed))
	hash := h.Sum32()

	// Map hash to one of the 4 patterns
	patternIdx := hash % 4
	switch patternIdx {
	case 0:
		return PatternVerbNoun
	case 1:
		return PatternAdjectiveNoun
	case 2:
		return PatternAbstractConcrete
	case 3:
		return PatternTechTerm
	default:
		return PatternVerbNoun
	}
}

// GenerateWithPattern generates an identity using the specified pattern.
// This function will be implemented in T007 to handle all four pattern types.
// For now, it returns a stub that will cause tests to fail.
func GenerateWithPattern(seed string, pattern Pattern) (string, error) {
	// TODO(T007): Implement actual multi-pattern generation
	// This stub ensures tests compile but fail during execution
	return "", fmt.Errorf("pattern %v not implemented", pattern)
}
