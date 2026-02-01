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
	PatternAdjectiveAdjectiveNoun
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
	case PatternAdjectiveAdjectiveNoun:
		return "AdjectiveAdjectiveNoun"
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

	// Map hash to one of the 5 patterns
	patternIdx := hash % 5
	switch patternIdx {
	case 0:
		return PatternVerbNoun
	case 1:
		return PatternAdjectiveNoun
	case 2:
		return PatternAbstractConcrete
	case 3:
		return PatternTechTerm
	case 4:
		return PatternAdjectiveAdjectiveNoun
	default:
		return PatternVerbNoun
	}
}

// GenerateWithPattern generates an identity using the specified pattern.
// Different patterns combine words from different categories:
// - VerbNoun: verb + animal (e.g., "chase-fox")
// - AdjectiveNoun: adjective + animal (e.g., "swift-bear")
// - AbstractConcrete: abstract concept + animal (e.g., "aether-wolf")
// - TechTerm: single tech term (e.g., "lambda")
// - AdjectiveAdjectiveNoun: adjective + adjective + animal (e.g., "swift-clever-fox")
func GenerateWithPattern(seed string, pattern Pattern) (string, error) {
	h := fnv.New32a()
	h.Write([]byte(seed))
	hash := h.Sum32()

	switch pattern {
	case PatternVerbNoun:
		verbIdx := hash % uint32(len(Verbs))
		animalIdx := (hash / uint32(len(Verbs))) % uint32(len(Animals))
		return fmt.Sprintf("%s-%s", Verbs[verbIdx], Animals[animalIdx]), nil

	case PatternAdjectiveNoun:
		adjIdx := hash % uint32(len(Adjectives))
		animalIdx := (hash / uint32(len(Adjectives))) % uint32(len(Animals))
		return fmt.Sprintf("%s-%s", Adjectives[adjIdx], Animals[animalIdx]), nil

	case PatternAbstractConcrete:
		abstractIdx := hash % uint32(len(Abstracts))
		animalIdx := (hash / uint32(len(Abstracts))) % uint32(len(Animals))
		return fmt.Sprintf("%s-%s", Abstracts[abstractIdx], Animals[animalIdx]), nil

	case PatternTechTerm:
		techIdx := hash % uint32(len(TechTerms))
		return TechTerms[techIdx], nil

	case PatternAdjectiveAdjectiveNoun:
		adj1Idx := hash % uint32(len(Adjectives))
		adj2Idx := (hash / uint32(len(Adjectives))) % uint32(len(Adjectives))
		animalIdx := (hash / uint32(len(Adjectives)*len(Adjectives))) % uint32(len(Animals))
		return fmt.Sprintf("%s-%s-%s", Adjectives[adj1Idx], Adjectives[adj2Idx], Animals[animalIdx]), nil

	default:
		return "", fmt.Errorf("invalid pattern: %v", pattern)
	}
}
