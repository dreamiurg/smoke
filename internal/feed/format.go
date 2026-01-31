package feed

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

// FormatOptions controls how posts are displayed
type FormatOptions struct {
	Oneline       bool      // Single-line compact format
	Quiet         bool      // Suppress headers and formatting
	ColorMode     ColorMode // Color output mode (Auto, Always, Never)
	TerminalWidth int       // Terminal width for wrapping (0 = auto-detect)
}

// getTerminalWidth returns the effective terminal width from options
func (o FormatOptions) getTerminalWidth() int {
	if o.TerminalWidth > 0 {
		return o.TerminalWidth
	}
	return GetTerminalWidth()
}

// FormatPost formats a single post for display
func FormatPost(w io.Writer, post *Post, opts FormatOptions) {
	cw := NewColorWriter(w, opts.ColorMode)
	if opts.Oneline {
		formatOneline(w, post, cw)
	} else {
		formatCompact(w, post, cw, opts.getTerminalWidth())
	}
}

// FormatFeed formats a list of posts with threading
func FormatFeed(w io.Writer, posts []*Post, opts FormatOptions, total int) {
	if len(posts) == 0 {
		if !opts.Quiet {
			_, _ = fmt.Fprintln(w, "No posts yet. Be the first! Try: smoke post \"hello world\"")
		}
		return
	}

	// Reset timestamp tracking for fresh feed display
	ResetTimestamp()

	cw := NewColorWriter(w, opts.ColorMode)

	// Build thread structure
	threads := buildThreads(posts)

	// Display threads
	termWidth := opts.getTerminalWidth()
	for i, thread := range threads {
		if opts.Oneline {
			formatOneline(w, thread.post, cw)
			for _, reply := range thread.replies {
				formatOneline(w, reply, cw)
			}
		} else {
			formatCompact(w, thread.post, cw, termWidth)
			for _, reply := range thread.replies {
				formatReply(w, thread.post, reply, cw, termWidth)
			}
			// Add blank line between threads (not after last one)
			if i < len(threads)-1 {
				_, _ = fmt.Fprintln(w)
			}
		}
	}

	// Footer
	if !opts.Quiet && total > len(posts) {
		_, _ = fmt.Fprintf(w, "\nShowing %d of %d posts. Use -n to see more.\n", len(posts), total)
	}
}

// FormatTailHeader prints the tail mode header
func FormatTailHeader(w io.Writer) {
	_, _ = fmt.Fprintln(w, "Watching for new posts... (Ctrl+C to stop)")
	_, _ = fmt.Fprintln(w)
}

type thread struct {
	post    *Post
	replies []*Post
}

// buildThreads groups replies under their parent posts
func buildThreads(posts []*Post) []thread {
	// Separate posts and replies
	postMap := make(map[string]*Post)
	replyMap := make(map[string][]*Post)
	var topLevelPosts []*Post

	for _, p := range posts {
		postMap[p.ID] = p
		if p.IsReply() {
			replyMap[p.ParentID] = append(replyMap[p.ParentID], p)
		} else {
			topLevelPosts = append(topLevelPosts, p)
		}
	}

	// Sort top-level posts by time (most recent first)
	sort.Slice(topLevelPosts, func(i, j int) bool {
		ti, errI := topLevelPosts[i].GetCreatedTime()
		tj, errJ := topLevelPosts[j].GetCreatedTime()
		if errI != nil || errJ != nil {
			return false
		}
		return ti.After(tj)
	})

	// Build threads
	threads := make([]thread, 0, len(topLevelPosts))
	for _, post := range topLevelPosts {
		t := thread{post: post}
		if replies, ok := replyMap[post.ID]; ok {
			// Sort replies by time (oldest first)
			sort.Slice(replies, func(i, j int) bool {
				ti, errI := replies[i].GetCreatedTime()
				tj, errJ := replies[j].GetCreatedTime()
				if errI != nil || errJ != nil {
					return false
				}
				return ti.Before(tj)
			})
			t.replies = replies
		}
		threads = append(threads, t)
	}

	return threads
}

// MinAuthorColumnWidth is the minimum width for identity column (right-aligned)
// Format: agent-adjective-animal@project (e.g., claude-swift-fox@smoke)
const MinAuthorColumnWidth = 28

// TimeColumnWidth is the width of the timestamp column (HH:MM)
const TimeColumnWidth = 5

// MinContentWidth is the minimum content width before we stop trying to wrap nicely
const MinContentWidth = 30

// lastTimestamp tracks the previous timestamp to avoid repetition
var lastTimestamp string

// AuthorLayout contains calculated layout for author column
type AuthorLayout struct {
	Padding  int // Number of spaces to prepend for right-alignment
	ColWidth int // Total column width
}

// CalculateAuthorLayout computes author column layout with right-alignment.
// Returns padding and total column width based on author length and minimum width.
func CalculateAuthorLayout(authorLen, minWidth int) AuthorLayout {
	if authorLen < minWidth {
		return AuthorLayout{
			Padding:  minWidth - authorLen,
			ColWidth: minWidth,
		}
	}
	return AuthorLayout{
		Padding:  0,
		ColWidth: authorLen,
	}
}

// ContentLayout contains calculated layout for content area
type ContentLayout struct {
	Start int // Column where content starts
	Width int // Available width for content
}

// CalculateContentLayout computes content area dimensions.
// Format: [time][space][padding+author][space 2][content]
func CalculateContentLayout(prefixWidth, authorColWidth, termWidth, minWidth int) ContentLayout {
	start := prefixWidth + 1 + authorColWidth + 2
	width := termWidth - start - 1 // -1 for safety margin
	if width < minWidth {
		width = minWidth
	}
	return ContentLayout{Start: start, Width: width}
}

// ResetTimestamp resets the timestamp tracking (call at start of feed)
func ResetTimestamp() {
	lastTimestamp = ""
}

// formatTimestamp returns the timestamp string for a post, or "??:??" on error
func formatTimestamp(post *Post) string {
	t, err := post.GetCreatedTime()
	if err != nil {
		return "??:??"
	}
	return t.Local().Format("15:04")
}

// formatCompact formats a post with right-aligned author@project and smart timestamps
// Format: 14:32  author@project  content (timestamp only shown when it changes)
func formatCompact(w io.Writer, post *Post, cw *ColorWriter, termWidth int) {
	timeStr := formatTimestamp(post)

	// Only show timestamp if different from previous
	var timeColumn string
	if timeStr != lastTimestamp {
		timeColumn = cw.Dim(timeStr)
		lastTimestamp = timeStr
	} else {
		timeColumn = "     " // 5 spaces to match timestamp width
	}

	// Build identity display with right-alignment
	// Author field contains full identity: agent-adjective-animal@project
	authorLayout := CalculateAuthorLayout(len(post.Author), MinAuthorColumnWidth)

	padding := ""
	if authorLayout.Padding > 0 {
		padding = fmt.Sprintf("%*s", authorLayout.Padding, "")
	}

	identity := cw.AuthorColorize(post.Author)
	authorRig := padding + identity

	// Calculate actual content start position and available width
	contentLayout := CalculateContentLayout(TimeColumnWidth, authorLayout.ColWidth, termWidth, MinContentWidth)

	// Wrap content if needed
	contentLines := wrapText(post.Content, contentLayout.Width)
	for i, line := range contentLines {
		highlightedLine := HighlightAll(line, cw.ColorEnabled)
		if i == 0 {
			// First line: full format
			_, _ = fmt.Fprintf(w, "%s %s  %s\n", timeColumn, authorRig, highlightedLine)
		} else {
			// Continuation lines: indent to align with content
			indent := fmt.Sprintf("%*s", contentLayout.Start, "")
			_, _ = fmt.Fprintf(w, "%s%s\n", indent, highlightedLine)
		}
	}
}

// wrapText wraps text to specified width, breaking on word boundaries
func wrapText(text string, maxWidth int) []string {
	if len(text) <= maxWidth {
		return []string{text}
	}

	var lines []string
	remaining := text

	for len(remaining) > maxWidth {
		// Find last space within maxWidth
		breakPoint := maxWidth
		for breakPoint > 0 && remaining[breakPoint] != ' ' {
			breakPoint--
		}
		if breakPoint == 0 {
			// No space found, force break at maxWidth (but don't exceed remaining length)
			breakPoint = maxWidth
			if breakPoint > len(remaining) {
				breakPoint = len(remaining)
			}
		}

		lines = append(lines, remaining[:breakPoint])
		remaining = remaining[breakPoint:]
		// Skip leading space on next line
		for len(remaining) > 0 && remaining[0] == ' ' {
			remaining = remaining[1:]
		}
	}

	if len(remaining) > 0 {
		lines = append(lines, remaining)
	}

	return lines
}

// formatReply formats a reply with indent (parent already shown in thread)
func formatReply(w io.Writer, _ *Post, reply *Post, cw *ColorWriter, termWidth int) {
	// For replies, always show timestamp (they're responses, timing matters)
	timestamp := cw.Dim(formatTimestamp(reply))

	// Reply prefix: "  └─ " = 5 chars
	const replyPrefix = 5

	// Build identity display - slightly smaller minimum width for reply indent
	minReplyAuthorWidth := MinAuthorColumnWidth - 3
	authorLayout := CalculateAuthorLayout(len(reply.Author), minReplyAuthorWidth)

	padding := ""
	if authorLayout.Padding > 0 {
		padding = fmt.Sprintf("%*s", authorLayout.Padding, "")
	}

	identity := cw.AuthorColorize(reply.Author)
	authorRig := padding + identity

	// Calculate actual content start position and available width
	contentLayout := CalculateContentLayout(replyPrefix+TimeColumnWidth, authorLayout.ColWidth, termWidth, MinContentWidth)

	// Wrap content if needed
	contentLines := wrapText(reply.Content, contentLayout.Width)
	for i, line := range contentLines {
		highlightedLine := HighlightAll(line, cw.ColorEnabled)
		if i == 0 {
			// First line: with tree character
			_, _ = fmt.Fprintf(w, "  └─ %s %s  %s\n", timestamp, authorRig, highlightedLine)
		} else {
			// Continuation lines: indent to align with content
			indent := fmt.Sprintf("%*s", contentLayout.Start, "")
			_, _ = fmt.Fprintf(w, "%s%s\n", indent, highlightedLine)
		}
	}
}

func formatOneline(w io.Writer, post *Post, cw *ColorWriter) {
	// Truncate content if needed for single line
	content := post.Content
	if len(content) > 60 {
		content = content[:57] + "..."
	}
	// Apply highlighting
	content = HighlightAll(content, cw.ColorEnabled)
	id := cw.Dim(post.ID)
	identity := cw.AuthorColorize(post.Author)
	_, _ = fmt.Fprintf(w, "%s %s %s\n", id, identity, content)
}

// FormatPosted outputs the confirmation message after posting
func FormatPosted(w io.Writer, post *Post) {
	_, _ = fmt.Fprintf(w, "Posted %s\n", post.ID)
}

// FormatReplied outputs the confirmation message after replying
func FormatReplied(w io.Writer, post *Post) {
	_, _ = fmt.Fprintf(w, "Replied %s -> %s\n", post.ID, post.ParentID)
}

// FilterCriteria specifies filters to apply when reading posts
type FilterCriteria struct {
	Author string
	Suffix string
	Since  time.Time
	Today  bool
}

// FilterPosts returns posts matching the given criteria
func FilterPosts(posts []*Post, criteria FilterCriteria) []*Post {
	result := make([]*Post, 0, len(posts))

	for _, post := range posts {
		// Author filter (supports substring matching for easier filtering)
		if criteria.Author != "" && !strings.Contains(post.Author, criteria.Author) {
			continue
		}

		// Suffix filter
		if criteria.Suffix != "" && post.Suffix != criteria.Suffix {
			continue
		}

		// Time filters
		if !criteria.Since.IsZero() {
			postTime, err := post.GetCreatedTime()
			if err != nil || postTime.Before(criteria.Since) {
				continue
			}
		}

		if criteria.Today {
			postTime, err := post.GetCreatedTime()
			if err != nil {
				continue
			}
			now := time.Now()
			startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			if postTime.Before(startOfDay) {
				continue
			}
		}

		result = append(result, post)
	}

	return result
}
