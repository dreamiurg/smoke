package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// ErrNoIdentity is returned when identity cannot be determined
var ErrNoIdentity = errors.New("cannot determine identity. Set BD_ACTOR or use --author")

// Identity represents the agent's identity
type Identity struct {
	Author string // Agent name (e.g., "ember", "witness")
	Rig    string // Gas Town rig name (e.g., "smoke", "calle")
}

// GetIdentity resolves the agent identity from environment variables
// Priority: BD_ACTOR > SMOKE_AUTHOR > error
func GetIdentity() (*Identity, error) {
	// Try BD_ACTOR first (format: "rig/role/name" or "rig/name")
	if bdActor := os.Getenv("BD_ACTOR"); bdActor != "" {
		return parseIdentity(bdActor)
	}

	// Try SMOKE_AUTHOR as fallback
	if smokeAuthor := os.Getenv("SMOKE_AUTHOR"); smokeAuthor != "" {
		rig, err := inferRig()
		if err != nil {
			rig = "unknown"
		}
		return &Identity{
			Author: sanitizeName(smokeAuthor),
			Rig:    rig,
		}, nil
	}

	return nil, ErrNoIdentity
}

// GetIdentityWithOverrides resolves identity with optional command-line overrides
func GetIdentityWithOverrides(authorOverride, rigOverride string) (*Identity, error) {
	identity, err := GetIdentity()

	// If no identity and no overrides, return error
	if err != nil && authorOverride == "" {
		return nil, err
	}

	// Create identity from overrides if needed
	if identity == nil {
		identity = &Identity{}
	}

	// Apply overrides
	if authorOverride != "" {
		identity.Author = sanitizeName(authorOverride)
	}
	if rigOverride != "" {
		identity.Rig = sanitizeName(rigOverride)
	}

	// Validate we have an author
	if identity.Author == "" {
		return nil, ErrNoIdentity
	}

	// Ensure rig is set
	if identity.Rig == "" {
		rig, err := inferRig()
		if err != nil {
			identity.Rig = "unknown"
		} else {
			identity.Rig = rig
		}
	}

	return identity, nil
}

// parseIdentity parses BD_ACTOR format: "rig/role/name" or "rig/name"
func parseIdentity(bdActor string) (*Identity, error) {
	parts := strings.Split(bdActor, "/")

	switch len(parts) {
	case 3:
		// Format: rig/role/name
		return &Identity{
			Rig:    sanitizeName(parts[0]),
			Author: sanitizeName(parts[2]),
		}, nil
	case 2:
		// Format: rig/name
		return &Identity{
			Rig:    sanitizeName(parts[0]),
			Author: sanitizeName(parts[1]),
		}, nil
	case 1:
		// Just a name
		rig, _ := inferRig()
		if rig == "" {
			rig = "unknown"
		}
		return &Identity{
			Rig:    rig,
			Author: sanitizeName(parts[0]),
		}, nil
	default:
		return nil, errors.New("invalid BD_ACTOR format")
	}
}

// inferRig tries to determine the rig name from the current directory
func inferRig() (string, error) {
	root, err := FindGasTownRoot()
	if err != nil {
		return "", err
	}
	return sanitizeName(filepath.Base(root)), nil
}

// sanitizeName removes whitespace and special characters from a name
func sanitizeName(name string) string {
	// Trim whitespace
	name = strings.TrimSpace(name)

	// Replace spaces with hyphens
	name = strings.ReplaceAll(name, " ", "-")

	// Remove any characters that aren't alphanumeric, hyphen, or underscore
	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			result.WriteRune(r)
		}
	}

	return strings.ToLower(result.String())
}
