// Package config provides configuration and initialization management for smoke.
// It handles directory paths, feed storage, and smoke initialization state.
package config

import (
	"errors"
	"fmt"
	"hash/fnv"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dreamiurg/smoke/internal/identity"
)

// ErrNoIdentity is returned when identity cannot be determined
var ErrNoIdentity = errors.New("cannot determine identity. Use --as flag or set SMOKE_AUTHOR")

// Identity represents the agent's identity for posting
type Identity struct {
	Agent   string // Agent type (e.g., "claude", "unknown")
	Suffix  string // Adjective-animal suffix (e.g., "swift-fox")
	Project string // Project name (e.g., "smoke")
}

// String returns the full identity string: agent-suffix@project or suffix@project
func (i *Identity) String() string {
	if i.Agent == "" {
		return fmt.Sprintf("%s@%s", i.Suffix, i.Project)
	}
	return fmt.Sprintf("%s-%s@%s", i.Agent, i.Suffix, i.Project)
}

// GetIdentity resolves the agent identity from environment, session, and optional override.
// If override is provided, it takes precedence. Otherwise, checks env vars (BD_ACTOR, then SMOKE_AUTHOR),
// then falls back to auto-detection.
func GetIdentity(override string) (*Identity, error) {
	// Use override if provided
	author := override

	// Otherwise check env vars (BD_ACTOR takes precedence, then SMOKE_AUTHOR)
	if author == "" {
		author = os.Getenv("BD_ACTOR")
		if author == "" {
			author = os.Getenv("SMOKE_AUTHOR")
		}
	}

	// If we have an explicit author (from override or env), use as custom identity
	if author != "" {
		// Strip @project if present (always ignore it)
		name := author
		if idx := strings.Index(author, "@"); idx != -1 {
			name = author[:idx] // Take only the name part before @
		}

		project := detectProject() // ALWAYS auto-detect

		// Use as custom identity (don't try to split agent-suffix for overrides)
		return &Identity{
			Agent:   "",
			Suffix:  sanitizeName(name),
			Project: project,
		}, nil
	}

	// Auto-detect identity
	seed := getSessionSeed()
	project := detectProject()

	if seed == "" {
		return nil, ErrNoIdentity
	}

	// Select pattern deterministically based on seed
	pattern := identity.SelectPattern(seed)

	// Generate suffix using the selected pattern
	suffix, err := identity.GenerateWithPattern(seed, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to generate suffix: %w", err)
	}

	// Apply style formatting deterministically based on seed
	words := splitSuffixIntoWords(suffix)
	styleFunc := selectStyleFunc(seed)
	styledSuffix := styleFunc(words)

	// Return identity without agent prefix (remove "claude" from output)
	return &Identity{
		Agent:   "",
		Suffix:  styledSuffix,
		Project: project,
	}, nil
}

// parseFullIdentity parses "agent-suffix@project" or "name@project" format
// NOTE: @project is ALWAYS auto-detected and cannot be overridden
func parseFullIdentity(s string) (*Identity, error) {
	parts := strings.SplitN(s, "@", 2)

	// Extract name part (before @, or whole string if no @)
	agentSuffix := parts[0]

	// ALWAYS auto-detect project, ignore @ override
	project := detectProject()

	// Split agent-suffix (e.g., "claude-swift-fox" -> "claude", "swift-fox")
	firstDash := strings.Index(agentSuffix, "-")
	if firstDash == -1 {
		// Simple name without dash (e.g., "ember@testrig")
		// Use as suffix only, no agent prefix
		return &Identity{
			Agent:   "",
			Suffix:  sanitizeName(agentSuffix),
			Project: project,
		}, nil
	}

	return &Identity{
		Agent:   sanitizeName(agentSuffix[:firstDash]),
		Suffix:  sanitizeName(agentSuffix[firstDash+1:]),
		Project: project,
	}, nil
}

// detectAgent determines the agent type from environment
func detectAgent() string {
	// Check for Claude Code
	home, err := os.UserHomeDir()
	if err != nil {
		return "unknown"
	}

	claudeDir := filepath.Join(home, ".claude")
	if _, err := os.Stat(claudeDir); err == nil {
		return "claude"
	}

	return "unknown"
}

// getSessionSeed returns a stable seed for the current session.
// When running under Claude Code, uses PPID for per-session identity.
// Otherwise tries TERM_SESSION_ID and WINDOWID, then falls back to PPID.
func getSessionSeed() string {
	// If running under Claude Code, use PPID for per-session identity.
	// Each Claude Code session has a unique PID, so smoke gets a fresh
	// identity per conversation even in the same terminal.
	if os.Getenv("CLAUDECODE") == "1" {
		ppid := os.Getppid()
		if ppid > 0 {
			return fmt.Sprintf("claude-ppid-%d", ppid)
		}
	}

	// Check terminal session identifiers
	if seed := os.Getenv("TERM_SESSION_ID"); seed != "" {
		return seed
	}
	if seed := os.Getenv("WINDOWID"); seed != "" {
		return seed
	}

	// Fallback to process parent ID (always available)
	ppid := os.Getppid()
	if ppid > 0 {
		return fmt.Sprintf("ppid-%d", ppid)
	}

	return ""
}

// detectProject determines the project name from git remote or cwd
func detectProject() string {
	// Try to get repo name from git remote origin URL
	// This works correctly for both main repo and worktrees
	cmd := exec.Command("git", "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err == nil {
		url := strings.TrimSpace(string(out))
		return sanitizeName(extractRepoName(url))
	}

	// Fallback to git toplevel directory name
	cmd = exec.Command("git", "rev-parse", "--show-toplevel")
	out, err = cmd.Output()
	if err == nil {
		root := strings.TrimSpace(string(out))
		return sanitizeName(filepath.Base(root))
	}

	// Fallback to cwd
	cwd, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return sanitizeName(filepath.Base(cwd))
}

// extractRepoName extracts the repository name from a git URL
// Handles both SSH (git@github.com:user/repo.git) and HTTPS (https://github.com/user/repo.git) formats
func extractRepoName(url string) string {
	// Remove trailing .git if present
	url = strings.TrimSuffix(url, ".git")

	// Try to get the last path component
	// For SSH: git@github.com:user/repo -> repo
	// For HTTPS: https://github.com/user/repo -> repo
	if idx := strings.LastIndex(url, "/"); idx != -1 {
		return url[idx+1:]
	}
	// For SSH with colon: git@github.com:user/repo
	if idx := strings.LastIndex(url, ":"); idx != -1 {
		path := url[idx+1:]
		if slashIdx := strings.LastIndex(path, "/"); slashIdx != -1 {
			return path[slashIdx+1:]
		}
		return path
	}

	return url
}

// sanitizeName removes whitespace and special characters from a name
func sanitizeName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, " ", "-")

	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			result.WriteRune(r)
		}
	}

	return strings.ToLower(result.String())
}

// splitSuffixIntoWords splits a hyphen-separated suffix into words.
// Examples: "swift-fox" -> ["swift", "fox"], "quantum_seeker" -> ["quantum_seeker"], "QuantumSeeker" -> ["QuantumSeeker"]
func splitSuffixIntoWords(suffix string) []string {
	if strings.Contains(suffix, "-") {
		return strings.Split(suffix, "-")
	}
	// If no hyphens, return as single word
	return []string{suffix}
}

// selectStyleFunc selects a style formatting function deterministically based on seed.
// Uses hash-based selection across 6 available styles.
func selectStyleFunc(seed string) func([]string) string {
	h := fnv.New32a()
	h.Write([]byte(seed))
	hash := h.Sum32()

	// 6 style options: Lowercase, SnakeCase, CamelCase, LowerCamel, KebabCase, WithNumber
	styleIdx := hash % 6

	switch styleIdx {
	case 0:
		return identity.Lowercase
	case 1:
		return identity.SnakeCase
	case 2:
		return identity.CamelCase
	case 3:
		return identity.LowerCamel
	case 4:
		return identity.KebabCase
	case 5:
		return identity.WithNumber
	default:
		return identity.Lowercase
	}
}
