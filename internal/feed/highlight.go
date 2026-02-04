package feed

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Patterns for detecting hashtags and mentions in text
var (
	// HashtagPattern matches #hashtag (alphanumeric and underscores)
	HashtagPattern = regexp.MustCompile(`(#[a-zA-Z0-9_]+)`)
	// MentionPattern matches @mention (alphanumeric and underscores)
	MentionPattern = regexp.MustCompile(`(@[a-zA-Z0-9_]+)`)
)

// HighlightHashtags colorizes hashtags in dim cyan (muted).
// If colorize is false, returns text unchanged.
// Deprecated: Use HighlightWithTheme instead to include background color.
func HighlightHashtags(text string, colorize bool) string {
	if !colorize {
		return text
	}
	return HashtagPattern.ReplaceAllStringFunc(text, func(match string) string {
		return Colorize(match, Dim, FgCyan)
	})
}

// HighlightMentions colorizes mentions in dim magenta (muted).
// If colorize is false, returns text unchanged.
// Deprecated: Use HighlightWithTheme instead to include background color.
func HighlightMentions(text string, colorize bool) string {
	if !colorize {
		return text
	}
	return MentionPattern.ReplaceAllStringFunc(text, func(match string) string {
		return Colorize(match, Dim, FgMagenta)
	})
}

// HighlightAll applies all highlighting (hashtags and mentions) to text.
// If colorize is false, returns text unchanged.
// Deprecated: Use HighlightWithTheme instead to include background color.
func HighlightAll(text string, colorize bool) string {
	if !colorize {
		return text
	}
	text = HighlightHashtags(text, true)
	text = HighlightMentions(text, true)
	return text
}

// HighlightWithTheme applies highlighting with proper background color from theme.
// This styles ALL text (both highlighted and plain) with background to prevent gaps.
func HighlightWithTheme(text string, theme *Theme) string {
	return HighlightWithThemeAndBackground(text, theme, theme.Background)
}

// HighlightWithThemeAndBackground applies highlighting with a custom background color.
// This styles ALL text (both highlighted and plain) to prevent gaps.
func HighlightWithThemeAndBackground(text string, theme *Theme, background lipgloss.AdaptiveColor) string {
	// Style for plain text: just background
	plainStyle := lipgloss.NewStyle().Background(background)

	// Style for hashtags: dim cyan with theme background
	hashtagStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#56b6c2")). // dim cyan
		Background(background).
		Faint(true)

	// Style for mentions: dim magenta with theme background
	mentionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#c678dd")). // dim magenta
		Background(background).
		Faint(true)

	// Combined pattern for both hashtags and mentions
	combinedPattern := regexp.MustCompile(`(#[a-zA-Z0-9_]+|@[a-zA-Z0-9_]+)`)

	// Find all matches and their positions
	matches := combinedPattern.FindAllStringIndex(text, -1)
	if len(matches) == 0 {
		// No highlights, just apply background to whole text
		return plainStyle.Render(text)
	}

	// Build result by styling each segment
	var result strings.Builder
	lastEnd := 0

	for _, match := range matches {
		start, end := match[0], match[1]

		// Style the plain text before this match
		if start > lastEnd {
			result.WriteString(plainStyle.Render(text[lastEnd:start]))
		}

		// Style the match itself
		matchText := text[start:end]
		if matchText[0] == '#' {
			result.WriteString(hashtagStyle.Render(matchText))
		} else {
			result.WriteString(mentionStyle.Render(matchText))
		}

		lastEnd = end
	}

	// Style any remaining plain text after the last match
	if lastEnd < len(text) {
		result.WriteString(plainStyle.Render(text[lastEnd:]))
	}

	return result.String()
}
