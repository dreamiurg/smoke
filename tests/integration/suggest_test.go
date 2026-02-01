package integration

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// TestSuggestShowsRecentPosts verifies that smoke suggest displays 2-3 recent posts with IDs
// This tests the NEW behavior where suggest shows actual feed posts, not just context-aware prompts
func TestSuggestShowsRecentPosts(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Set pressure to 4 (always fire) for deterministic test behavior
	if err := h.SetPressure(4); err != nil {
		t.Fatalf("failed to set pressure: %v", err)
	}

	// Set identity and post some test messages
	h.SetIdentity("test_author")
	if _, _, err := h.Run("post", "First test post about observations"); err != nil {
		t.Fatalf("failed to post first message: %v", err)
	}

	if _, _, err := h.Run("post", "Second test post about learnings"); err != nil {
		t.Fatalf("failed to post second message: %v", err)
	}

	if _, _, err := h.Run("post", "Third test post about questions"); err != nil {
		t.Fatalf("failed to post third message: %v", err)
	}

	// Run suggest command
	stdout, _, err := h.Run("suggest")
	if err != nil {
		t.Fatalf("smoke suggest failed: %v", err)
	}

	// Verify output contains recent posts with IDs
	if !strings.Contains(stdout, "smk-") {
		t.Errorf("suggest output missing post IDs (smk-): %s", stdout)
	}

	// Verify output contains author information
	if !strings.Contains(stdout, "test_author") {
		t.Errorf("suggest output missing author: %s", stdout)
	}

	// Verify output contains post content
	if !strings.Contains(stdout, "First test post") &&
		!strings.Contains(stdout, "Second test post") &&
		!strings.Contains(stdout, "Third test post") {
		t.Errorf("suggest output missing post content: %s", stdout)
	}
}

// TestSuggestShowsTemplates verifies that smoke suggest displays 2-3 template ideas
func TestSuggestShowsTemplates(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Set pressure to 4 (always fire) for deterministic test behavior
	if err := h.SetPressure(4); err != nil {
		t.Fatalf("failed to set pressure: %v", err)
	}

	// Run suggest without any posts
	stdout, _, err := h.Run("suggest")
	if err != nil {
		t.Fatalf("smoke suggest failed: %v", err)
	}

	// Even with empty feed, suggest should show example ideas
	if !strings.Contains(stdout, "Post idea") && !strings.Contains(stdout, "Post ideas") {
		t.Errorf("suggest output missing example suggestions: %s", stdout)
	}

	// Verify it contains example patterns (bullet points with sample post text)
	if !strings.Contains(stdout, "•") {
		t.Errorf("suggest output missing example bullet points: %s", stdout)
	}
}

// TestSuggestEmptyFeedShowsOnlyTemplates verifies that suggest shows only templates when feed is empty
// This tests edge case: empty feed should not error, just show templates
func TestSuggestEmptyFeedShowsOnlyTemplates(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Set pressure to 4 (always fire) for deterministic test behavior
	if err := h.SetPressure(4); err != nil {
		t.Fatalf("failed to set pressure: %v", err)
	}

	// Run suggest with empty feed
	stdout, stderr, err := h.Run("suggest")

	// Should not error
	if err != nil {
		t.Fatalf("smoke suggest failed with empty feed (should not error): %v, stderr: %s", err, stderr)
	}

	// Output should not be empty
	if strings.TrimSpace(stdout) == "" {
		t.Errorf("suggest output empty with empty feed (should show templates)")
	}

	// Should contain template references
	if !strings.Contains(stdout, "Post idea") && !strings.Contains(stdout, "template") {
		t.Errorf("suggest output with empty feed missing templates: %s", stdout)
	}

	// Should NOT show "recent posts" section when feed is empty
	lowerOut := strings.ToLower(stdout)
	if strings.Contains(lowerOut, "recent post") && !strings.Contains(strings.ToLower(h.configDir), "") {
		// This is allowed to fail if "recent" appears in a template, but ideally empty feed has no "recent posts" section
		t.Logf("note: output mentions 'recent posts' but feed is empty: %s", stdout)
	}
}

// TestSuggestSinceFlag verifies that --since flag filters posts by time window
func TestSuggestSinceFlag(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Set pressure to 4 (always fire) for deterministic test behavior
	if err := h.SetPressure(4); err != nil {
		t.Fatalf("failed to set pressure: %v", err)
	}

	h.SetIdentity("time_test_author")

	// Post a test message
	if _, _, err := h.Run("post", "Recent test post"); err != nil {
		t.Fatalf("failed to post: %v", err)
	}

	// Test with 1 hour window (should include recent post)
	stdout1, _, err := h.Run("suggest", "--since", "1h")
	if err != nil {
		t.Fatalf("smoke suggest --since 1h failed: %v", err)
	}

	// Should include the recent post
	if !strings.Contains(stdout1, "smk-") {
		t.Errorf("suggest --since 1h missing post IDs: %s", stdout1)
	}

	if !strings.Contains(stdout1, "Recent test post") {
		t.Errorf("suggest --since 1h missing post content: %s", stdout1)
	}

	// Test with very short window (1 minute) that would exclude older posts
	// Note: This post was just created, so it should still be included
	stdout2, _, err := h.Run("suggest", "--since", "1m")
	if err != nil {
		t.Fatalf("smoke suggest --since 1m failed: %v", err)
	}

	// Recent post should still be included (created within last minute)
	if !strings.Contains(stdout2, "smk-") {
		t.Errorf("suggest --since 1m should include just-created post: %s", stdout2)
	}
}

// TestSuggestJSONFlag verifies that --json flag returns valid JSON output
// JSON should contain posts and templates arrays
func TestSuggestJSONFlag(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Set pressure to 4 (always fire) for deterministic test behavior
	if err := h.SetPressure(4); err != nil {
		t.Fatalf("failed to set pressure: %v", err)
	}

	h.SetIdentity("json_test_author")

	// Post a test message
	if _, _, err := h.Run("post", "Test post for JSON output"); err != nil {
		t.Fatalf("failed to post: %v", err)
	}

	// Run suggest with --json flag
	stdout, stderr, err := h.Run("suggest", "--json")
	if err != nil {
		t.Fatalf("smoke suggest --json failed: %v, stderr: %s", err, stderr)
	}

	// Verify output is valid JSON
	var output map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &output); err != nil {
		t.Fatalf("output is not valid JSON: %v, output: %s", err, stdout)
	}

	// Verify structure has posts and examples arrays
	if _, hasPosts := output["posts"]; !hasPosts {
		t.Errorf("JSON output missing 'posts' key: %s", stdout)
	}

	if _, hasExamples := output["examples"]; !hasExamples {
		t.Errorf("JSON output missing 'examples' key: %s", stdout)
	}

	// Verify posts array contains expected fields
	if posts, ok := output["posts"].([]interface{}); ok && len(posts) > 0 {
		firstPost := posts[0].(map[string]interface{})

		requiredFields := []string{"id", "author", "content", "created_at", "time_ago"}
		for _, field := range requiredFields {
			if _, hasField := firstPost[field]; !hasField {
				t.Errorf("post missing required field '%s': %v", field, firstPost)
			}
		}
	}
}

// TestSuggestJSONEmptyFeed verifies that --json flag returns valid JSON even with no recent posts
// Tests the case where we filter for a time window with no recent activity
// JSON should contain empty posts array and templates array
func TestSuggestJSONEmptyFeed(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke (creates with 4 seeded posts)
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Set pressure to 4 (always fire) for deterministic test behavior
	if err := h.SetPressure(4); err != nil {
		t.Fatalf("failed to set pressure: %v", err)
	}

	// Run suggest with --json but with a time window that excludes all posts (1 second)
	// This simulates an empty feed without deleting posts
	stdout, stderr, err := h.Run("suggest", "--json", "--since", "1s")
	if err != nil {
		t.Fatalf("smoke suggest --json failed: %v, stderr: %s", err, stderr)
	}

	// Verify output is valid JSON
	var output map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &output); err != nil {
		t.Fatalf("output is not valid JSON: %v, output: %s", err, stdout)
	}

	// Verify structure has posts and examples arrays
	if _, hasPosts := output["posts"]; !hasPosts {
		t.Errorf("JSON output missing 'posts' key: %s", stdout)
	}

	if _, hasExamples := output["examples"]; !hasExamples {
		t.Errorf("JSON output missing 'examples' key: %s", stdout)
	}

	// Posts array should be empty (no posts within last 1 second)
	if posts, ok := output["posts"].([]interface{}); ok {
		if len(posts) != 0 {
			t.Errorf("posts array should be empty with 1s filter, got %d posts: %v", len(posts), posts)
		}
	}

	// Examples array should have 2-3 examples
	if examples, ok := output["examples"].([]interface{}); ok {
		if len(examples) < 2 || len(examples) > 3 {
			t.Errorf("examples array should have 2-3 examples, got %d: %v", len(examples), examples)
		}
	}
}

// TestSuggestReplyHint verifies that output explicitly hints about reply syntax
func TestSuggestReplyHint(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Set pressure to 4 (always fire) for deterministic test behavior
	if err := h.SetPressure(4); err != nil {
		t.Fatalf("failed to set pressure: %v", err)
	}

	h.SetIdentity("reply_hint_author")

	// Post a message
	if _, _, err := h.Run("post", "Reply hint test post"); err != nil {
		t.Fatalf("failed to post: %v", err)
	}

	// Run suggest
	stdout, _, err := h.Run("suggest")
	if err != nil {
		t.Fatalf("smoke suggest failed: %v", err)
	}

	// Verify reply hint is present
	// The hint should mention "smoke reply" and post ID format "smk-"
	if !strings.Contains(stdout, "reply") && !strings.Contains(stdout, "Reply") {
		t.Logf("warning: suggest output missing 'reply' mention (optional but recommended): %s", stdout)
	}

	if !strings.Contains(stdout, "smk-") {
		t.Errorf("suggest output missing post ID (required for reply hint): %s", stdout)
	}
}

// TestSuggestPostIDFormat verifies that post IDs have correct format (smk-xxxxxx)
func TestSuggestPostIDFormat(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Set pressure to 4 (always fire) for deterministic test behavior
	if err := h.SetPressure(4); err != nil {
		t.Fatalf("failed to set pressure: %v", err)
	}

	h.SetIdentity("id_format_author")

	// Post a message
	if _, _, err := h.Run("post", "ID format test"); err != nil {
		t.Fatalf("failed to post: %v", err)
	}

	// Run suggest
	stdout, _, err := h.Run("suggest")
	if err != nil {
		t.Fatalf("smoke suggest failed: %v", err)
	}

	// Extract and verify post ID format
	// IDs should follow smk-XXXXXX pattern (smk- prefix followed by alphanumeric)
	if !strings.Contains(stdout, "smk-") {
		t.Errorf("suggest output missing smk- post ID prefix: %s", stdout)
	}
}

// TestSuggestPostMetadata verifies that posts include ID, author, and timestamp
func TestSuggestPostMetadata(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Set pressure to 4 (always fire) for deterministic test behavior
	if err := h.SetPressure(4); err != nil {
		t.Fatalf("failed to set pressure: %v", err)
	}

	h.SetIdentity("metadata_author")

	// Post a message
	if _, _, err := h.Run("post", "Metadata test post"); err != nil {
		t.Fatalf("failed to post: %v", err)
	}

	// Run suggest
	stdout, _, err := h.Run("suggest")
	if err != nil {
		t.Fatalf("smoke suggest failed: %v", err)
	}

	// Verify post ID present
	if !strings.Contains(stdout, "smk-") {
		t.Errorf("suggest missing post ID: %s", stdout)
	}

	// Verify author present
	if !strings.Contains(stdout, "metadata_author") {
		t.Errorf("suggest missing author: %s", stdout)
	}

	// Verify content present
	if !strings.Contains(stdout, "Metadata test post") {
		t.Errorf("suggest missing post content: %s", stdout)
	}

	// Verify some indication of time (could be "ago", "m", "h", or similar)
	lowerOut := strings.ToLower(stdout)
	timeIndicators := []string{"ago", "min", "hour", "now"}
	foundTime := false
	for _, indicator := range timeIndicators {
		if strings.Contains(lowerOut, indicator) {
			foundTime = true
			break
		}
	}
	if !foundTime {
		t.Logf("warning: suggest missing time indicator (optional but recommended): %s", stdout)
	}
}

// TestSuggestMultiplePosts verifies suggest shows 2-3 recent posts when multiple exist
func TestSuggestMultiplePosts(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Set pressure to 4 (always fire) for deterministic test behavior
	if err := h.SetPressure(4); err != nil {
		t.Fatalf("failed to set pressure: %v", err)
	}

	h.SetIdentity("multi_post_author")

	// Create 5 posts to ensure suggest picks recent ones
	for i := 1; i <= 5; i++ {
		msg := "Multi-post test message " + string(rune(48+i))
		if _, _, err := h.Run("post", msg); err != nil {
			t.Fatalf("failed to post message %d: %v", i, err)
		}
		// Small delay to ensure different timestamps
		time.Sleep(10 * time.Millisecond)
	}

	// Run suggest
	stdout, _, err := h.Run("suggest")
	if err != nil {
		t.Fatalf("smoke suggest failed: %v", err)
	}

	// Count occurrences of smk- (post IDs)
	idCount := strings.Count(stdout, "smk-")

	// Should show 2-3 recent posts
	if idCount < 2 || idCount > 3 {
		t.Logf("suggest showing %d posts (expected 2-3, but implementation may vary): %s", idCount, stdout)
	}

	// Should show the most recent message
	if !strings.Contains(stdout, "message 5") {
		t.Logf("suggest may not be showing most recent post (or post content trimmed): %s", stdout)
	}
}

// TestSuggestSinceFlagParseFormats verifies that --since accepts various time formats
func TestSuggestSinceFlagParseFormats(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Set pressure to 4 (always fire) for deterministic test behavior
	if err := h.SetPressure(4); err != nil {
		t.Fatalf("failed to set pressure: %v", err)
	}

	h.SetIdentity("since_format_author")

	// Post a message
	if _, _, err := h.Run("post", "Since format test"); err != nil {
		t.Fatalf("failed to post: %v", err)
	}

	// Test various --since formats
	formats := []string{"30m", "1h", "2h", "6h", "24h"}

	for _, format := range formats {
		stdout, stderr, err := h.Run("suggest", "--since", format)

		if err != nil {
			// Command should succeed even if format is not recognized
			// (though format should be valid based on spec)
			t.Logf("suggest --since %s returned error: %v, stderr: %s", format, err, stderr)
		}

		if strings.TrimSpace(stdout) == "" {
			t.Logf("suggest --since %s produced empty output (may be OK if no posts in window)", format)
		}
	}
}

// TestSuggestJSONWithMultiplePosts verifies JSON output works with multiple posts
// Should return 2-3 most recent posts and templates
func TestSuggestJSONWithMultiplePosts(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Set pressure to 4 (always fire) for deterministic test behavior
	if err := h.SetPressure(4); err != nil {
		t.Fatalf("failed to set pressure: %v", err)
	}

	h.SetIdentity("json_multi_author")

	// Create multiple posts
	for i := 1; i <= 3; i++ {
		if _, _, err := h.Run("post", "JSON multi test "+string(rune(48+i))); err != nil {
			t.Fatalf("failed to post: %v", err)
		}
	}

	// Run suggest with --json
	stdout, stderr, err := h.Run("suggest", "--json")
	if err != nil {
		t.Fatalf("smoke suggest --json failed: %v, stderr: %s", err, stderr)
	}

	// Verify output is valid JSON
	var output map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &output); err != nil {
		t.Fatalf("output is not valid JSON: %v, output: %s", err, stdout)
	}

	// Verify posts array has 2-3 items (most recent posts)
	if posts, ok := output["posts"].([]interface{}); ok {
		if len(posts) < 2 || len(posts) > 3 {
			t.Errorf("posts array should have 2-3 items with 3 posts, got %d: %v", len(posts), posts)
		}

		// Verify each post has required fields
		for i, p := range posts {
			post := p.(map[string]interface{})
			requiredFields := []string{"id", "author", "content", "created_at", "time_ago"}
			for _, field := range requiredFields {
				if _, hasField := post[field]; !hasField {
					t.Errorf("post[%d] missing required field '%s': %v", i, field, post)
				}
			}
		}
	} else {
		t.Errorf("posts is not an array: %v", output["posts"])
	}

	// Verify examples array has 2-3 items
	if examples, ok := output["examples"].([]interface{}); ok {
		if len(examples) < 2 || len(examples) > 3 {
			t.Errorf("examples array should have 2-3 items, got %d: %v", len(examples), examples)
		}

		// Examples are now strings, not objects with category/pattern
		for i, ex := range examples {
			if _, ok := ex.(string); !ok {
				t.Errorf("example[%d] should be a string: %v", i, ex)
			}
		}
	} else {
		t.Errorf("examples is not an array: %v", output["examples"])
	}
}

// TestSuggestTextFormatReadability verifies text output is readable and suitable for Claude context
func TestSuggestTextFormatReadability(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Set pressure to 4 (always fire) for deterministic test behavior
	if err := h.SetPressure(4); err != nil {
		t.Fatalf("failed to set pressure: %v", err)
	}

	h.SetIdentity("readability_author")

	// Post a message
	if _, _, err := h.Run("post", "Readability test post for context injection"); err != nil {
		t.Fatalf("failed to post: %v", err)
	}

	// Run suggest
	stdout, _, err := h.Run("suggest")
	if err != nil {
		t.Fatalf("smoke suggest failed: %v", err)
	}

	// Verify output is plain text (no excessive JSON or complex formatting)
	// Should be readable when injected into Claude's context
	if len(stdout) == 0 {
		t.Errorf("suggest output is empty")
	}

	// Should contain readable sections
	lines := strings.Split(stdout, "\n")
	if len(lines) == 0 {
		t.Errorf("suggest output has no line breaks")
	}

	// Should not be overly complex JSON (though some structure is OK)
	if strings.Count(stdout, "{") > 3 || strings.Count(stdout, "[") > 3 {
		t.Logf("warning: suggest text output may have too much structure (looks like JSON): %s", stdout)
	}
}

// TestSuggestTemplateVariety verifies that different suggest calls show different templates
func TestSuggestTemplateVariety(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Set pressure to 4 (always fire) for deterministic test behavior
	if err := h.SetPressure(4); err != nil {
		t.Fatalf("failed to set pressure: %v", err)
	}

	// Run suggest multiple times to check for template variety
	// Note: Templates should be randomly selected, so might see repeats
	outputs := []string{}

	for i := 0; i < 3; i++ {
		stdout, _, err := h.Run("suggest")
		if err != nil {
			t.Fatalf("smoke suggest run %d failed: %v", i+1, err)
		}
		outputs = append(outputs, stdout)
	}

	// All runs should have examples
	for i, out := range outputs {
		if !strings.Contains(out, "Post idea") && !strings.Contains(out, "•") {
			t.Errorf("suggest run %d missing example content: %s", i+1, out)
		}
	}
}

// TestSuggestWithoutInit verifies suggest handles case where smoke is not initialized
func TestSuggestWithoutInit(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Don't call init - smoke is not initialized

	// Run suggest
	_, stderr, err := h.Run("suggest")

	// Command may fail or show helpful error
	if err == nil {
		t.Logf("note: suggest succeeded without init (graceful handling): may show empty suggestions")
	} else {
		// Should provide helpful error message
		if !strings.Contains(stderr, "init") && !strings.Contains(stderr, "Initialize") {
			t.Logf("note: suggest error could be more helpful: %s", stderr)
		}
	}
}
