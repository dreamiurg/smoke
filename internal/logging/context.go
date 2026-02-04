package logging

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Context captures invocation context for telemetry.
// All fields use consistent naming for log analysis.
type Context struct {
	Identity string // Full identity string "swift-fox@smoke"
	Agent    string // Agent type: "claude", "human", "unknown"
	Caller   string // Caller agent type: claude, codex, gemini, unknown
	Session  string // Session ID for correlation
	Env      string // Environment: "claude_code", "ci", "terminal"
	Project  string // Project name
	Cwd      string // Working directory
}

// CaptureContext gathers invocation context from the environment.
// Call this once at command start.
func CaptureContext() *Context {
	ctx := &Context{
		Env:     detectEnv(),
		Cwd:     getCwd(),
		Session: getSessionID(),
		Caller:  detectCallerAgent(),
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
		slog.String("caller", c.Caller),
		slog.String("session", c.Session),
		slog.String("env", c.Env),
		slog.String("project", c.Project),
		slog.String("cwd", c.Cwd),
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

// detectCallerAgent attempts to identify the calling agent type.
func detectCallerAgent() string {
	if v := strings.TrimSpace(os.Getenv("SMOKE_AGENT")); v != "" {
		return strings.ToLower(v)
	}
	if os.Getenv("CLAUDECODE") == "1" || os.Getenv("CLAUDE_CODE") == "1" {
		return "claude"
	}
	if os.Getenv("CLAUDE_CODE_SUBAGENT_MODEL") != "" ||
		os.Getenv("ANTHROPIC_API_KEY") != "" ||
		os.Getenv("ANTHROPIC_MODEL") != "" ||
		os.Getenv("ANTHROPIC_DEFAULT_OPUS_MODEL") != "" ||
		os.Getenv("ANTHROPIC_DEFAULT_SONNET_MODEL") != "" ||
		os.Getenv("ANTHROPIC_DEFAULT_HAIKU_MODEL") != "" {
		return "claude"
	}
	if os.Getenv("GEMINI_CLI") != "" {
		return "gemini"
	}
	if os.Getenv("CODEX") == "1" || os.Getenv("CODEX_CLI") != "" || os.Getenv("OPENAI_CODEX") != "" ||
		os.Getenv("CODEX_CI") == "1" {
		return "codex"
	}
	if os.Getenv("GEMINI_API_KEY") != "" || os.Getenv("GOOGLE_API_KEY") != "" ||
		os.Getenv("GEMINI_MODEL") != "" || os.Getenv("GOOGLE_CLOUD_PROJECT") != "" ||
		os.Getenv("GOOGLE_CLOUD_LOCATION") != "" {
		return "gemini"
	}
	if os.Getenv("OPENAI_API_KEY") != "" {
		return "codex"
	}
	if findAgentAncestor("claude") {
		return "claude"
	}
	if findAgentAncestor("codex") {
		return "codex"
	}
	if findAgentAncestor("gemini") {
		return "gemini"
	}
	return "unknown"
}

// DetectCallerAgent returns the detected caller agent type.
func DetectCallerAgent() string {
	return detectCallerAgent()
}

// findAgentAncestor walks up the process tree looking for a process name match.
func findAgentAncestor(substr string) bool {
	substr = strings.ToLower(substr)
	pid := os.Getpid()
	visited := make(map[int]bool)

	for pid > 1 && !visited[pid] {
		visited[pid] = true
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
		comm := strings.ToLower(fields[1])
		if strings.Contains(comm, substr) {
			return true
		}
		pid = ppid
	}
	return false
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
