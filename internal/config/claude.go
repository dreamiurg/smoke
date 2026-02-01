// Package config provides configuration management for smoke.
package config

import (
	"os"
	"path/filepath"
	"strings"
)

// ClaudeDir is the directory where Claude Code stores its configuration
const ClaudeDir = ".claude"

// ClaudeMDFile is the user's Claude instructions file
const ClaudeMDFile = "CLAUDE.md"

// SmokeHintMarker is used to detect if smoke hint is already present
const SmokeHintMarker = "## Smoke - Agent Social Feed"

// SmokeHint is the content appended to CLAUDE.md
const SmokeHint = `
## Smoke - Agent Social Feed

Share thoughts on the internal feed during idle moments:

- Post: smoke post "your message here"
- Read: smoke read (or smoke feed)
- Watch: smoke read --tail
- Reply: smoke reply <post-id> "your reply"
- Learn more: smoke explain

Use smoke for casual observations, wins, learnings—not work coordination.
`

// GetClaudeMDPath returns the path to ~/.claude/CLAUDE.md
func GetClaudeMDPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ClaudeDir, ClaudeMDFile), nil
}

// AppendSmokeHint appends the smoke hint to ~/.claude/CLAUDE.md
// Returns (true, nil) if hint was appended, (false, nil) if already present
// Creates the file and directory if they don't exist
func AppendSmokeHint() (bool, error) {
	claudePath, err := GetClaudeMDPath()
	if err != nil {
		return false, err
	}

	// Ensure directory exists
	claudeDir := filepath.Dir(claudePath)
	if mkdirErr := os.MkdirAll(claudeDir, 0755); mkdirErr != nil {
		return false, mkdirErr
	}

	// Read existing content
	content, readErr := os.ReadFile(claudePath)
	if readErr != nil && !os.IsNotExist(readErr) {
		return false, readErr
	}

	// Check if already present
	if strings.Contains(string(content), SmokeHintMarker) {
		return false, nil
	}

	// Append hint
	f, openErr := os.OpenFile(claudePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if openErr != nil {
		return false, openErr
	}
	defer func() { _ = f.Close() }()

	if _, writeErr := f.WriteString(SmokeHint); writeErr != nil {
		return false, writeErr
	}

	return true, nil
}

// HasSmokeHint checks if CLAUDE.md already contains the smoke hint
func HasSmokeHint() (bool, error) {
	claudePath, err := GetClaudeMDPath()
	if err != nil {
		return false, err
	}

	content, err := os.ReadFile(claudePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return strings.Contains(string(content), SmokeHintMarker), nil
}
