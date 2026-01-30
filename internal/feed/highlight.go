package feed

import "regexp"

// Patterns for detecting hashtags and mentions in text
var (
	// HashtagPattern matches #hashtag (alphanumeric and underscores)
	HashtagPattern = regexp.MustCompile(`(#[a-zA-Z0-9_]+)`)
	// MentionPattern matches @mention (alphanumeric and underscores)
	MentionPattern = regexp.MustCompile(`(@[a-zA-Z0-9_]+)`)
)

// HighlightHashtags colorizes hashtags in cyan.
// If colorize is false, returns text unchanged.
func HighlightHashtags(text string, colorize bool) string {
	if !colorize {
		return text
	}
	return HashtagPattern.ReplaceAllStringFunc(text, func(match string) string {
		return Colorize(match, FgCyan)
	})
}

// HighlightMentions colorizes mentions in magenta.
// If colorize is false, returns text unchanged.
func HighlightMentions(text string, colorize bool) string {
	if !colorize {
		return text
	}
	return MentionPattern.ReplaceAllStringFunc(text, func(match string) string {
		return Colorize(match, FgMagenta)
	})
}

// HighlightAll applies all highlighting (hashtags and mentions) to text.
// If colorize is false, returns text unchanged.
func HighlightAll(text string, colorize bool) string {
	if !colorize {
		return text
	}
	text = HighlightHashtags(text, true)
	text = HighlightMentions(text, true)
	return text
}
