package logging

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// Context captures invocation context for telemetry.
// All fields use consistent naming for log analysis.
type Context struct {
	Identity string // Full identity string "claude@swift-fox/smoke"
	Agent    string // Agent type: "claude", "human", "unknown"
	Session  string // Session ID for correlation
	Env      string // Environment: "claude_code", "ci", "terminal"
	Project  string // Project name
	Cwd      string // Working directory
	BdActor  string // BD_ACTOR env var if set
}

// CaptureContext gathers invocation context from the environment.
// Call this once at command start.
func CaptureContext() *Context {
	ctx := &Context{
		Env:     detectEnv(),
		Cwd:     getCwd(),
		BdActor: os.Getenv("BD_ACTOR"),
		Session: getSessionID(),
	}
	return ctx
}

// SetIdentity sets identity fields after resolution.
// Called by CLI after identity is resolved.
func (c *Context) SetIdentity(identity, agent, project string) {
	c.Identity = identity
	c.Agent = agent
	c.Project = project
}

// Attrs returns slog attributes for the context group.
func (c *Context) Attrs() slog.Attr {
	attrs := []any{
		slog.String("identity", c.Identity),
		slog.String("agent", c.Agent),
		slog.String("session", c.Session),
		slog.String("env", c.Env),
		slog.String("project", c.Project),
		slog.String("cwd", c.Cwd),
	}
	// Only include bd_actor if set
	if c.BdActor != "" {
		attrs = append(attrs, slog.String("bd_actor", c.BdActor))
	}
	return slog.Group("ctx", attrs...)
}

// detectEnv determines the execution environment.
func detectEnv() string {
	// Claude Code sets CLAUDECODE=1
	if os.Getenv("CLAUDECODE") == "1" {
		return "claude_code"
	}

	// Common CI environment variables
	ciVars := []string{
		"CI",
		"GITHUB_ACTIONS",
		"GITLAB_CI",
		"CIRCLECI",
		"JENKINS_URL",
		"BUILDKITE",
		"TRAVIS",
	}
	for _, v := range ciVars {
		if os.Getenv(v) != "" {
			return "ci"
		}
	}

	// Check if running in a terminal
	if isTerminal() {
		return "terminal"
	}

	return "unknown"
}

// isTerminal checks if stdout is a terminal.
func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// getCwd returns the current working directory, or empty on error.
func getCwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return cwd
}

// getSessionID returns a session identifier for correlation.
// Uses TERM_SESSION_ID, Claude PPID, or falls back to PPID.
func getSessionID() string {
	// If running under Claude Code, use parent PID
	if os.Getenv("CLAUDECODE") == "1" {
		ppid := os.Getppid()
		if ppid > 0 {
			return formatSessionID("claude", ppid)
		}
	}

	// Terminal session ID
	if termID := os.Getenv("TERM_SESSION_ID"); termID != "" {
		// Shorten long UUIDs for readability
		if len(termID) > 16 {
			termID = termID[:16]
		}
		return termID
	}

	// Window ID (X11/Wayland)
	if windowID := os.Getenv("WINDOWID"); windowID != "" {
		return "win-" + windowID
	}

	// Fallback to parent PID
	ppid := os.Getppid()
	if ppid > 0 {
		return formatSessionID("pid", ppid)
	}

	return "unknown"
}

// formatSessionID creates a consistent session ID format.
func formatSessionID(prefix string, pid int) string {
	return prefix + "-" + itoa(pid)
}

// itoa converts int to string without fmt dependency.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	if i < 0 {
		return "-" + itoa(-i)
	}
	var b strings.Builder
	var digits [20]byte
	n := 0
	for i > 0 {
		digits[n] = byte('0' + i%10)
		i /= 10
		n++
	}
	for n > 0 {
		n--
		b.WriteByte(digits[n])
	}
	return b.String()
}

// extractProjectFromCwd attempts to extract project name from cwd.
func extractProjectFromCwd(cwd string) string {
	if cwd == "" {
		return ""
	}
	return filepath.Base(cwd)
}
