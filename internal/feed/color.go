package feed

import (
	"hash/fnv"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ANSI escape sequences for terminal styling
const (
	Reset = "\033[0m"
	Bold  = "\033[1m"
	Dim   = "\033[2m"
)

// ANSI foreground colors
const (
	FgRed     = "\033[31m"
	FgGreen   = "\033[32m"
	FgYellow  = "\033[33m"
	FgBlue    = "\033[34m"
	FgMagenta = "\033[35m"
	FgCyan    = "\033[36m"
)

// AuthorPalette defines the colors used for author names.
// Excludes black (invisible on dark), white (default text),
// and magenta (reserved for @mentions).
var AuthorPalette = []string{
	FgBlue,
	FgGreen,
	FgYellow,
	FgCyan,
	FgRed,
}

// AuthorColor returns a deterministic color for the given author name.
// The same author always gets the same color.
func AuthorColor(author string) string {
	h := fnv.New32a()
	h.Write([]byte(author))
	idx := h.Sum32() % uint32(len(AuthorPalette))
	return AuthorPalette[idx]
}

// Colorize wraps text with the given ANSI codes and resets afterward.
// If color is empty, returns the text unchanged.
func Colorize(text string, codes ...string) string {
	if len(codes) == 0 {
		return text
	}
	var prefix string
	for _, code := range codes {
		prefix += code
	}
	return prefix + text + Reset
}

// ColorWriter wraps an io.Writer and conditionally applies color.
// When ColorEnabled is false, color functions return plain text.
type ColorWriter struct {
	W            io.Writer
	ColorEnabled bool
}

// NewColorWriter creates a ColorWriter with the given writer and color mode.
func NewColorWriter(w io.Writer, mode ColorMode) *ColorWriter {
	return &ColorWriter{
		W:            w,
		ColorEnabled: ShouldColorize(mode),
	}
}

// Colorize applies color codes only if color is enabled.
func (cw *ColorWriter) Colorize(text string, codes ...string) string {
	if !cw.ColorEnabled {
		return text
	}
	return Colorize(text, codes...)
}

// AuthorColorize returns the colored author name with identity splitting.
// For "agent@project" format, colors the agent name and dims the project.
// If color is disabled, returns the plain author string.
func (cw *ColorWriter) AuthorColorize(author string) string {
	if !cw.ColorEnabled {
		return author
	}

	agent, project := SplitIdentity(author)

	// Color the agent name
	coloredAgent := Colorize(agent, Bold, AuthorColor(agent))

	// If there's a project, dim it
	if project == "" {
		return coloredAgent
	}

	dimmedProject := Colorize(project, Dim)
	return coloredAgent + "@" + dimmedProject
}

// Dim returns dimmed text if color is enabled.
func (cw *ColorWriter) Dim(text string) string {
	if !cw.ColorEnabled {
		return text
	}
	return Colorize(text, Dim)
}

// SplitIdentity splits an identity string into agent and project parts.
// Identity format is "agent@project". If no @ is found, returns the full string as agent.
func SplitIdentity(author string) (agent, project string) {
	parts := strings.Split(author, "@")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return author, ""
}

// ColorizeIdentity applies theme and contrast styling to identity parts.
// Uses lipgloss.Color objects from Theme for proper TUI rendering.
func ColorizeIdentity(author string, theme *Theme, contrast *ContrastLevel) string {
	agent, project := SplitIdentity(author)

	// Build agent style using theme colors
	agentStyle := lipgloss.NewStyle().
		Foreground(theme.AgentColors[hashAgentName(agent)%len(theme.AgentColors)])

	if contrast.AgentBold {
		agentStyle = agentStyle.Bold(true)
	}

	styledAgent := agentStyle.Render(agent)

	// Handle project part if it exists
	if project == "" {
		return styledAgent
	}

	projectStyle := lipgloss.NewStyle()
	if contrast.ProjectColored {
		// Color the project with a secondary theme color
		projectStyle = projectStyle.Foreground(theme.AgentColors[(hashAgentName(project)+1)%len(theme.AgentColors)])
	} else {
		// Dim the project
		projectStyle = projectStyle.Foreground(theme.Dim)
	}

	styledProject := projectStyle.Render(project)

	return styledAgent + "@" + styledProject
}

// hashAgentName computes a deterministic hash for agent name coloring.
func hashAgentName(agent string) int {
	h := fnv.New32a()
	h.Write([]byte(agent))
	return int(h.Sum32())
}
