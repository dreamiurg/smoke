package config

import (
	"errors"
	"fmt"
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

// GetIdentity resolves the agent identity from environment and session
func GetIdentity() (*Identity, error) {
	// Check for explicit override first (BD_ACTOR takes precedence, then SMOKE_AUTHOR)
	author := os.Getenv("BD_ACTOR")
	if author == "" {
		author = os.Getenv("SMOKE_AUTHOR")
	}

	if author != "" {
		// Parse if it's a full identity (contains @)
		if strings.Contains(author, "@") {
			return parseFullIdentity(author)
		}
		// Otherwise use as-is with detected project
		project := detectProject()
		return &Identity{
			Agent:   "custom",
			Suffix:  sanitizeName(author),
			Project: project,
		}, nil
	}

	// Auto-detect identity
	agent := detectAgent()
	seed := getSessionSeed()
	project := detectProject()

	if seed == "" {
		return nil, ErrNoIdentity
	}

	suffix := identity.Generate(seed)

	return &Identity{
		Agent:   agent,
		Suffix:  suffix,
		Project: project,
	}, nil
}

// GetIdentityWithOverride resolves identity with optional --as override
func GetIdentityWithOverride(authorOverride string) (*Identity, error) {
	if authorOverride != "" {
		// Parse override as full identity or suffix
		if strings.Contains(authorOverride, "@") {
			return parseFullIdentity(authorOverride)
		}
		project := detectProject()
		return &Identity{
			Agent:   "custom",
			Suffix:  sanitizeName(authorOverride),
			Project: project,
		}, nil
	}
	return GetIdentity()
}

// parseFullIdentity parses "agent-suffix@project" or "name@project" format
func parseFullIdentity(s string) (*Identity, error) {
	parts := strings.SplitN(s, "@", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid identity format: %s", s)
	}

	agentSuffix := parts[0]
	project := sanitizeName(parts[1])

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
		// If we can't get home directory, use "unknown" as fallback
		home = "unknown"
	}
	if home != "" && home != "unknown" {
		claudeDir := filepath.Join(home, ".claude")
		if _, err := os.Stat(claudeDir); err == nil {
			return "claude"
		}
	}

	return "unknown"
}

// getSessionSeed returns a stable seed for the current session
func getSessionSeed() string {
	// Try various session identifiers in order of preference
	signals := []string{
		os.Getenv("TERM_SESSION_ID"), // macOS Terminal
		os.Getenv("WINDOWID"),        // X11
	}

	for _, sig := range signals {
		if sig != "" {
			return sig
		}
	}

	// Fallback: PPID + TTY
	ppid := os.Getppid()
	tty := os.Getenv("TTY")
	if tty == "" {
		// Try to get TTY another way
		if ttyname, err := os.Readlink("/dev/fd/0"); err == nil {
			tty = ttyname
		}
	}

	if ppid > 0 {
		return fmt.Sprintf("%d-%s", ppid, tty)
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
