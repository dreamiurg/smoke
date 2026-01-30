package feed

import (
	"fmt"
	"io"
	"sort"
	"time"
)

// FormatOptions controls how posts are displayed
type FormatOptions struct {
	Oneline   bool      // Single-line compact format
	Quiet     bool      // Suppress headers and formatting
	ColorMode ColorMode // Color output mode (Auto, Always, Never)
}

// FormatPost formats a single post for display
func FormatPost(w io.Writer, post *Post, opts FormatOptions) {
	cw := NewColorWriter(w, opts.ColorMode)
	if opts.Oneline {
		formatOneline(w, post, cw)
	} else {
		formatCompact(w, post, cw)
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
	for i, thread := range threads {
		if opts.Oneline {
			formatOneline(w, thread.post, cw)
			for _, reply := range thread.replies {
				formatOneline(w, reply, cw)
			}
		} else {
			formatCompact(w, thread.post, cw)
			for _, reply := range thread.replies {
				formatReply(w, thread.post, reply, cw)
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

// AuthorColumnWidth is the fixed width for author@rig column (right-aligned)
const AuthorColumnWidth = 16

// ContentColumnStart is where content begins (for wrapping alignment)
// = 5 (time or spaces) + 1 (space) + AuthorColumnWidth + 2 (spacing)
const ContentColumnStart = 24

// MaxContentWidth is the max width before wrapping (assuming ~100 char terminal)
const MaxContentWidth = 75

// lastTimestamp tracks the previous timestamp to avoid repetition
var lastTimestamp string

// ResetTimestamp resets the timestamp tracking (call at start of feed)
func ResetTimestamp() {
	lastTimestamp = ""
}

// formatCompact formats a post with right-aligned author@rig and smart timestamps
// Format: 14:32  author@rig  content (timestamp only shown when it changes)
func formatCompact(w io.Writer, post *Post, cw *ColorWriter) {
	t, err := post.GetCreatedTime()
	var timeStr string
	if err != nil {
		timeStr = "??:??"
	} else {
		timeStr = t.Local().Format("15:04")
	}

	// Only show timestamp if different from previous
	var timeColumn string
	if timeStr != lastTimestamp {
		timeColumn = cw.Dim(timeStr)
		lastTimestamp = timeStr
	} else {
		timeColumn = "     " // 5 spaces to match timestamp width
	}

	// Build author@rig with right-alignment
	authorRigPlain := post.Author + "@" + post.Rig
	visibleLen := len(authorRigPlain)

	// Right-align: add padding before author@rig
	padding := ""
	if visibleLen < AuthorColumnWidth {
		padding = fmt.Sprintf("%*s", AuthorColumnWidth-visibleLen, "")
	}

	author := cw.AuthorColorize(post.Author)
	rig := cw.Dim("@" + post.Rig)
	authorRig := padding + author + rig

	// Wrap content if needed
	contentLines := wrapText(post.Content, MaxContentWidth)
	for i, line := range contentLines {
		highlightedLine := HighlightAll(line, cw.ColorEnabled)
		if i == 0 {
			// First line: full format
			_, _ = fmt.Fprintf(w, "%s %s  %s\n", timeColumn, authorRig, highlightedLine)
		} else {
			// Continuation lines: indent to align with content
			indent := fmt.Sprintf("%*s", ContentColumnStart, "")
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
			// No space found, force break at maxWidth
			breakPoint = maxWidth
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
func formatReply(w io.Writer, _ *Post, reply *Post, cw *ColorWriter) {
	replyTime, err := reply.GetCreatedTime()
	var replyTimeStr string
	if err != nil {
		replyTimeStr = "??:??"
	} else {
		replyTimeStr = replyTime.Local().Format("15:04")
	}

	// For replies, always show timestamp (they're responses, timing matters)
	timestamp := cw.Dim(replyTimeStr)

	// Build author@rig - slightly smaller width for reply indent
	authorRigPlain := reply.Author + "@" + reply.Rig
	visibleLen := len(authorRigPlain)
	replyAuthorWidth := AuthorColumnWidth - 3

	// Right-align
	padding := ""
	if visibleLen < replyAuthorWidth {
		padding = fmt.Sprintf("%*s", replyAuthorWidth-visibleLen, "")
	}

	author := cw.AuthorColorize(reply.Author)
	rig := cw.Dim("@" + reply.Rig)
	authorRig := padding + author + rig

	// Wrap content if needed (slightly narrower for replies)
	contentLines := wrapText(reply.Content, MaxContentWidth-5)
	for i, line := range contentLines {
		highlightedLine := HighlightAll(line, cw.ColorEnabled)
		if i == 0 {
			// First line: with tree character
			_, _ = fmt.Fprintf(w, "  └─ %s %s  %s\n", timestamp, authorRig, highlightedLine)
		} else {
			// Continuation lines: indent to align with content
			indent := fmt.Sprintf("%*s", ContentColumnStart+5, "")
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
	author := cw.AuthorColorize(post.Author)
	_, _ = fmt.Fprintf(w, "%s %s@%s %s\n", id, author, post.Rig, content)
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
	Rig    string
	Since  time.Time
	Today  bool
}

// FilterPosts returns posts matching the given criteria
func FilterPosts(posts []*Post, criteria FilterCriteria) []*Post {
	result := make([]*Post, 0, len(posts))

	for _, post := range posts {
		// Author filter
		if criteria.Author != "" && post.Author != criteria.Author {
			continue
		}

		// Rig filter
		if criteria.Rig != "" && post.Rig != criteria.Rig {
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
