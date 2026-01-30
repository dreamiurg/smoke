package feed

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestFormatPost(t *testing.T) {
	post := &Post{
		ID:        "smk-abc123",
		Author:    "ember",
		Rig:       "smoke",
		Content:   "hello world",
		CreatedAt: "2026-01-30T09:24:00Z",
	}

	t.Run("default format", func(t *testing.T) {
		var buf bytes.Buffer
		FormatPost(&buf, post, FormatOptions{})

		output := buf.String()
		if !strings.Contains(output, "ember@smoke") {
			t.Errorf("FormatPost() output missing author@rig: %s", output)
		}
		if !strings.Contains(output, "hello world") {
			t.Errorf("FormatPost() output missing content: %s", output)
		}
	})

	t.Run("oneline format", func(t *testing.T) {
		var buf bytes.Buffer
		FormatPost(&buf, post, FormatOptions{Oneline: true})

		output := buf.String()
		if !strings.HasPrefix(output, "smk-abc123") {
			t.Errorf("FormatPost(oneline) output should start with ID: %s", output)
		}
		if strings.Count(output, "\n") > 1 {
			t.Errorf("FormatPost(oneline) output has multiple lines: %s", output)
		}
	})
}

func TestFormatFeed(t *testing.T) {
	posts := []*Post{
		{
			ID:        "smk-aaa111",
			Author:    "ember",
			Rig:       "smoke",
			Content:   "first post",
			CreatedAt: "2026-01-30T09:00:00Z",
		},
		{
			ID:        "smk-bbb222",
			Author:    "witness",
			Rig:       "smoke",
			Content:   "second post",
			CreatedAt: "2026-01-30T09:05:00Z",
		},
	}

	t.Run("default format", func(t *testing.T) {
		var buf bytes.Buffer
		FormatFeed(&buf, posts, FormatOptions{}, 10)

		output := buf.String()
		if !strings.Contains(output, "ember@smoke") {
			t.Errorf("FormatFeed() output missing first author: %s", output)
		}
		if !strings.Contains(output, "witness@smoke") {
			t.Errorf("FormatFeed() output missing second author: %s", output)
		}
	})

	t.Run("shows footer when more posts", func(t *testing.T) {
		var buf bytes.Buffer
		FormatFeed(&buf, posts, FormatOptions{}, 100)

		output := buf.String()
		if !strings.Contains(output, "Showing 2 of 100 posts") {
			t.Errorf("FormatFeed() missing footer: %s", output)
		}
	})

	t.Run("quiet mode no footer", func(t *testing.T) {
		var buf bytes.Buffer
		FormatFeed(&buf, posts, FormatOptions{Quiet: true}, 100)

		output := buf.String()
		if strings.Contains(output, "Showing") {
			t.Errorf("FormatFeed(quiet) should not show footer: %s", output)
		}
	})

	t.Run("empty feed", func(t *testing.T) {
		var buf bytes.Buffer
		FormatFeed(&buf, []*Post{}, FormatOptions{}, 0)

		output := buf.String()
		if !strings.Contains(output, "No posts yet") {
			t.Errorf("FormatFeed() empty should show message: %s", output)
		}
	})

	t.Run("oneline format", func(t *testing.T) {
		var buf bytes.Buffer
		FormatFeed(&buf, posts, FormatOptions{Oneline: true}, 2)

		output := buf.String()
		if !strings.HasPrefix(output, "smk-") {
			t.Errorf("FormatFeed(oneline) should start with ID: %s", output)
		}
	})
}

func TestFormatFeedWithReplies(t *testing.T) {
	posts := []*Post{
		{
			ID:        "smk-parent",
			Author:    "ember",
			Rig:       "smoke",
			Content:   "parent post",
			CreatedAt: "2026-01-30T09:00:00Z",
		},
		{
			ID:        "smk-reply1",
			Author:    "witness",
			Rig:       "smoke",
			Content:   "reply to parent",
			CreatedAt: "2026-01-30T09:05:00Z",
			ParentID:  "smk-parent",
		},
	}

	var buf bytes.Buffer
	FormatFeed(&buf, posts, FormatOptions{}, 2)

	output := buf.String()

	// Reply should be indented
	if !strings.Contains(output, "\\--") {
		t.Errorf("FormatFeed() reply should be indented with \\--: %s", output)
	}
}

func TestFormatTailHeader(t *testing.T) {
	var buf bytes.Buffer
	FormatTailHeader(&buf)

	output := buf.String()
	if !strings.Contains(output, "Watching for new posts") {
		t.Errorf("FormatTailHeader() missing header: %s", output)
	}
	if !strings.Contains(output, "Ctrl+C") {
		t.Errorf("FormatTailHeader() missing Ctrl+C instruction: %s", output)
	}
}

func TestFormatPosted(t *testing.T) {
	post := &Post{
		ID:        "smk-posted",
		Author:    "ember",
		Rig:       "smoke",
		Content:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	var buf bytes.Buffer
	FormatPosted(&buf, post)

	output := buf.String()
	if !strings.Contains(output, "Posted smk-posted") {
		t.Errorf("FormatPosted() output: %s", output)
	}
}

func TestFormatReplied(t *testing.T) {
	reply := &Post{
		ID:        "smk-reply1",
		Author:    "witness",
		Rig:       "smoke",
		Content:   "nice!",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		ParentID:  "smk-parent",
	}

	var buf bytes.Buffer
	FormatReplied(&buf, reply)

	output := buf.String()
	if !strings.Contains(output, "Replied smk-reply1 -> smk-parent") {
		t.Errorf("FormatReplied() output: %s", output)
	}
}

func TestFilterPosts(t *testing.T) {
	now := time.Now()
	posts := []*Post{
		{
			ID:        "smk-aaa111",
			Author:    "ember",
			Rig:       "smoke",
			Content:   "ember smoke post",
			CreatedAt: now.Add(-1 * time.Hour).UTC().Format(time.RFC3339),
		},
		{
			ID:        "smk-bbb222",
			Author:    "witness",
			Rig:       "smoke",
			Content:   "witness smoke post",
			CreatedAt: now.Add(-30 * time.Minute).UTC().Format(time.RFC3339),
		},
		{
			ID:        "smk-ccc333",
			Author:    "ember",
			Rig:       "calle",
			Content:   "ember calle post",
			CreatedAt: now.Add(-10 * time.Minute).UTC().Format(time.RFC3339),
		},
		{
			ID:        "smk-ddd444",
			Author:    "witness",
			Rig:       "calle",
			Content:   "witness calle post",
			CreatedAt: now.Add(-25 * time.Hour).UTC().Format(time.RFC3339), // yesterday
		},
	}

	t.Run("filter by author", func(t *testing.T) {
		result := FilterPosts(posts, FilterCriteria{Author: "ember"})
		if len(result) != 2 {
			t.Errorf("FilterPosts(author=ember) returned %d, want 2", len(result))
		}
	})

	t.Run("filter by rig", func(t *testing.T) {
		result := FilterPosts(posts, FilterCriteria{Rig: "smoke"})
		if len(result) != 2 {
			t.Errorf("FilterPosts(rig=smoke) returned %d, want 2", len(result))
		}
	})

	t.Run("filter by author and rig", func(t *testing.T) {
		result := FilterPosts(posts, FilterCriteria{Author: "ember", Rig: "smoke"})
		if len(result) != 1 {
			t.Errorf("FilterPosts(author=ember, rig=smoke) returned %d, want 1", len(result))
		}
	})

	t.Run("filter by since", func(t *testing.T) {
		result := FilterPosts(posts, FilterCriteria{Since: now.Add(-45 * time.Minute)})
		if len(result) != 2 {
			t.Errorf("FilterPosts(since=45m) returned %d, want 2", len(result))
		}
	})

	t.Run("filter today", func(t *testing.T) {
		result := FilterPosts(posts, FilterCriteria{Today: true})
		if len(result) != 3 {
			t.Errorf("FilterPosts(today) returned %d, want 3", len(result))
		}
	})

	t.Run("no filter", func(t *testing.T) {
		result := FilterPosts(posts, FilterCriteria{})
		if len(result) != 4 {
			t.Errorf("FilterPosts(no filter) returned %d, want 4", len(result))
		}
	})
}

func TestFormatDefaultWithInvalidTime(t *testing.T) {
	post := &Post{
		ID:        "smk-abc123",
		Author:    "ember",
		Rig:       "smoke",
		Content:   "test",
		CreatedAt: "invalid-time",
	}

	var buf bytes.Buffer
	FormatPost(&buf, post, FormatOptions{})

	output := buf.String()
	if !strings.Contains(output, "??:??") {
		t.Errorf("FormatPost() should show ??:?? for invalid time: %s", output)
	}
}

func TestFormatOnelineTruncation(t *testing.T) {
	longContent := strings.Repeat("a", 100)
	post := &Post{
		ID:        "smk-abc123",
		Author:    "ember",
		Rig:       "smoke",
		Content:   longContent,
		CreatedAt: "2026-01-30T09:24:00Z",
	}

	var buf bytes.Buffer
	FormatPost(&buf, post, FormatOptions{Oneline: true})

	output := buf.String()
	if len(output) > 100 {
		// Content should be truncated
		if !strings.Contains(output, "...") {
			t.Errorf("FormatPost(oneline) should truncate long content with ...: %s", output)
		}
	}
}
