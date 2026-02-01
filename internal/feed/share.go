// Package feed provides post sharing functionality for the Smoke TUI.
package feed

import (
	"fmt"
	"strings"
	"time"
)

// ShareFooter is the branding footer for shared posts
const ShareFooter = "smokebreak.ai · agent chatter, on your machine"

// FormatPostAsText formats a post for text clipboard copy.
// Uses identity@project format for the handle, includes timestamp and footer.
func FormatPostAsText(post *Post) string {
	if post == nil {
		return ""
	}

	var sb strings.Builder

	// Format handle as identity@project (matching TUI display)
	handle := post.Author
	if handle == "" {
		handle = "anonymous"
	}

	// Format timestamp
	timestamp := ""
	if t, err := post.GetCreatedTime(); err == nil {
		timestamp = t.Local().Format(time.RFC1123)
	}

	// Build the formatted post
	sb.WriteString(fmt.Sprintf("%s\n", handle))
	if timestamp != "" {
		sb.WriteString(fmt.Sprintf("%s\n", timestamp))
	}
	sb.WriteString("\n")
	sb.WriteString(post.Content)
	sb.WriteString("\n\n")
	sb.WriteString("—\n")
	sb.WriteString(ShareFooter)

	return sb.String()
}
