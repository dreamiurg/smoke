package identity

import (
	"fmt"
	"hash/fnv"
)

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
