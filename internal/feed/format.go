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
	Oneline bool // Single-line compact format
	Quiet   bool // Suppress headers and formatting
}

// FormatPost formats a single post for display
func FormatPost(w io.Writer, post *Post, opts FormatOptions) {
	if opts.Oneline {
		formatOneline(w, post)
	} else {
		formatDefault(w, post, 0)
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

	// Build thread structure
	threads := buildThreads(posts)

	// Display threads
	for _, thread := range threads {
		if opts.Oneline {
			formatOneline(w, thread.post)
			for _, reply := range thread.replies {
				formatOneline(w, reply)
			}
		} else {
			formatDefault(w, thread.post, 0)
			for _, reply := range thread.replies {
				formatDefault(w, reply, 1)
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

func formatDefault(w io.Writer, post *Post, indent int) {
	t, err := post.GetCreatedTime()
	var timeStr string
	if err != nil {
		timeStr = "??:??"
	} else {
		timeStr = t.Local().Format("15:04")
	}

	prefix := ""
	if indent > 0 {
		prefix = "  " + strings.Repeat("  ", indent-1) + "\\-- "
	}

	_, _ = fmt.Fprintf(w, "%s[%s] %s@%s: %s\n", prefix, timeStr, post.Author, post.Rig, post.Content)
}

func formatOneline(w io.Writer, post *Post) {
	// Truncate content if needed for single line
	content := post.Content
	if len(content) > 60 {
		content = content[:57] + "..."
	}
	_, _ = fmt.Fprintf(w, "%s %s@%s %s\n", post.ID, post.Author, post.Rig, content)
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
