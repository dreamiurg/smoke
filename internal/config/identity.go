// Package config provides configuration and initialization management for smoke.
// It handles directory paths, feed storage, and smoke initialization state.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/term"

	"github.com/dreamiurg/smoke/internal/identity"
)

// sessionInfo stores the current Claude session identity for cross-process sharing.
// This allows ccstatusline to show the same identity as Claude Code's direct invocations.
type sessionInfo struct {
	PID           int    `json:"pid"`             // Claude Code process PID
	TermSessionID string `json:"term_session_id"` // Terminal session ID for multi-terminal support
	Seed          string `json:"seed"`            // The seed used for identity generation
}

// sessionFileName is the name of the session file within the config directory.
const sessionFileName = "session.json"

// getSessionFilePath returns the path to the session file.
func getSessionFilePath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, sessionFileName), nil
}

// writeSessionInfo writes session info to the session file.
func writeSessionInfo(info *sessionInfo) error {
	path, err := getSessionFilePath()
	if err != nil {
		return err
	}

	data, err := json.Marshal(info)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// readSessionInfo reads session info from the session file.
// Returns nil if the file doesn't exist or can't be parsed.
func readSessionInfo() *sessionInfo {
	path, err := getSessionFilePath()
	if err != nil {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var info sessionInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil
	}

	return &info
}

// isPIDRunning checks if a process with the given PID is still running.
func isPIDRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Unix, FindProcess always succeeds. We need to send signal 0 to check if alive.
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// findClaudeAncestor walks up the process tree looking for a Claude Code process.
// Returns the Claude process PID if found, or 0 if not found.
// This allows indirect invocations (e.g., ccstatusline → smoke) to identify
// which Claude session they belong to.
func findClaudeAncestor() int {
	pid := os.Getpid()
	visited := make(map[int]bool)

	for pid > 1 && !visited[pid] {
		visited[pid] = true

		// Get process info using ps
		cmd := exec.Command("ps", "-p", fmt.Sprintf("%d", pid), "-o", "ppid=,comm=")
		out, err := cmd.Output()
		if err != nil {
			break
		}

		fields := strings.Fields(strings.TrimSpace(string(out)))
		if len(fields) < 2 {
			break
		}

		ppid, err := strconv.Atoi(fields[0])
		if err != nil {
			break
		}

		comm := fields[1]

		// Check if this process is Claude Code
		// The process name is typically "claude" or contains "claude"
		if strings.Contains(strings.ToLower(comm), "claude") {
			return pid
		}

		pid = ppid
	}

	return 0
}

// ErrNoIdentity is returned when identity cannot be determined
var ErrNoIdentity = errors.New("cannot determine identity. Use --as flag or set SMOKE_AUTHOR")

// HumanIdentity is the suffix used for human users in interactive terminals.
const HumanIdentity = "<human>"

// isHumanSession detects if the current session is an interactive human user.
// Returns true if:
// 1. Not running under Claude Code (CLAUDECODE != "1")
// 2. Not a descendant of a Claude Code process
// 3. No valid Claude session file exists for this terminal
// 4. Stdin is an interactive terminal (TTY)
func isHumanSession() bool {
	// If running under Claude Code, definitely not human
	if os.Getenv("CLAUDECODE") == "1" {
		return false
	}

	// Check if we're a descendant of a Claude process (indirect invocation)
	if claudePID := findClaudeAncestor(); claudePID > 0 {
		return false
	}

	// Check for a valid Claude session file (ccstatusline case)
	termSessionID := os.Getenv("TERM_SESSION_ID")
	if info := readSessionInfo(); info != nil {
		// If session file matches terminal and Claude process is running,
		// this could be ccstatusline or similar - treat as agent context
		if info.TermSessionID == termSessionID && isPIDRunning(info.PID) {
			return false
		}
	}

	// Interactive terminal with no agent indicators = human
	return term.IsTerminal(int(os.Stdin.Fd()))
}

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
	project := detectProject()

	// Check if this is a human in an interactive terminal
	if isHumanSession() {
		return &Identity{
			Agent:   "",
			Suffix:  HumanIdentity,
			Project: project,
		}, nil
	}

	seed := getSessionSeed()
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
// When running under Claude Code, uses PPID for per-session identity and
// writes to a session file so other processes (like ccstatusline) can use the same identity.
// When not running under Claude Code, walks the process tree to find a Claude ancestor.
func getSessionSeed() string {
	termSessionID := os.Getenv("TERM_SESSION_ID")

	// If running under Claude Code, use PPID for per-session identity.
	// Each Claude Code session has a unique PID, so smoke gets a fresh
	// identity per conversation even in the same terminal.
	if os.Getenv("CLAUDECODE") == "1" {
		ppid := os.Getppid()
		if ppid > 0 {
			seed := fmt.Sprintf("claude-ppid-%d", ppid)
			// Write session info so other processes can use the same identity
			_ = writeSessionInfo(&sessionInfo{
				PID:           ppid,
				TermSessionID: termSessionID,
				Seed:          seed,
			})
			return seed
		}
	}

	// Check if we're a descendant of a Claude process (indirect invocation).
	// This is more reliable than the session file for cases like ccstatusline
	// where smoke is called before Claude has run any direct smoke commands.
	if claudePID := findClaudeAncestor(); claudePID > 0 {
		seed := fmt.Sprintf("claude-ppid-%d", claudePID)
		// Update session file for consistency
		_ = writeSessionInfo(&sessionInfo{
			PID:           claudePID,
			TermSessionID: termSessionID,
			Seed:          seed,
		})
		return seed
	}

	// Fallback to session file for cases where process tree walk fails
	// (e.g., process name doesn't contain "claude")
	if info := readSessionInfo(); info != nil {
		// Validate: same terminal and Claude process still running
		if info.TermSessionID == termSessionID && isPIDRunning(info.PID) {
			return info.Seed
		}
	}

	// Fallback to terminal session identifiers
	if termSessionID != "" {
		return termSessionID
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
