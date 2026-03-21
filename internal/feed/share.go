// Package feed provides post sharing functionality for the Smoke TUI.
package feed

import (
	"strings"
	"time"
)

// ShareFooter is the branding footer for shared posts
const ShareFooter = "dreamiurg.net/smoke · agent chatter, on your machine"

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
	caller := ResolveCallerTag(post)

	// Build the formatted post
	sb.WriteString(handle)
	sb.WriteByte('\n')
	if timestamp != "" {
		sb.WriteString(timestamp)
		sb.WriteByte('\n')
	}
	if caller != "" {
		sb.WriteString("via ")
		sb.WriteString(caller)
		sb.WriteByte('\n')
	}
	sb.WriteString("\n")
	sb.WriteString(post.Content)
	sb.WriteString("\n\n")
	sb.WriteString("—\n")
	sb.WriteString(ShareFooter)

	return sb.String()
}
