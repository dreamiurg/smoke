package feed

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dreamiurg/smoke/internal/config"
)

func setupTestStore(t *testing.T) (*Store, string) {
	t.Helper()

	tmpDir := t.TempDir()
	feedPath := filepath.Join(tmpDir, "feed.jsonl")

	// Create empty feed file
	err := os.WriteFile(feedPath, []byte{}, 0644)
	require.NoError(t, err)

	return NewStoreWithPath(feedPath), feedPath
}

func TestStoreAppend(t *testing.T) {
	store, _ := setupTestStore(t)

	post := &Post{
		ID:        "smk-abc123",
		Author:    "ember",
		Suffix:    "smoke",
		Content:   "test post",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	err := store.Append(post)
	assert.NoError(t, err)

	// Verify post was written
	posts, err := store.ReadAll()
	require.NoError(t, err)
	assert.Len(t, posts, 1)
	assert.Equal(t, post.ID, posts[0].ID)
}

func TestStoreAppendValidation(t *testing.T) {
	store, _ := setupTestStore(t)

	invalidPost := &Post{
		ID:        "invalid",
		Author:    "ember",
		Suffix:    "smoke",
		Content:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	err := store.Append(invalidPost)
	assert.Error(t, err)
}

func TestStoreAppendNotInitialized(t *testing.T) {
	tmpDir := t.TempDir()
	feedPath := filepath.Join(tmpDir, "nonexistent.jsonl")
	store := NewStoreWithPath(feedPath)

	post := &Post{
		ID:        "smk-abc123",
		Author:    "ember",
		Suffix:    "smoke",
		Content:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	err := store.Append(post)
	assert.Equal(t, config.ErrNotInitialized, err)
}

func TestStoreReadAll(t *testing.T) {
	store, _ := setupTestStore(t)

	// Write multiple posts
	posts := []*Post{
		{
			ID:        "smk-aaa111",
			Author:    "ember",
			Suffix:    "smoke",
			Content:   "first post",
			CreatedAt: "2026-01-30T09:00:00Z",
		},
		{
			ID:        "smk-bbb222",
			Author:    "witness",
			Suffix:    "smoke",
			Content:   "second post",
			CreatedAt: "2026-01-30T09:05:00Z",
		},
		{
			ID:        "smk-ccc333",
			Author:    "ember",
			Suffix:    "calle",
			Content:   "third post",
			CreatedAt: "2026-01-30T09:10:00Z",
		},
	}

	for _, p := range posts {
		err := store.Append(p)
		require.NoError(t, err)
	}

	// Read all
	got, err := store.ReadAll()
	require.NoError(t, err)
	assert.Len(t, got, 3)

	// Verify content
	contentMap := make(map[string]bool)
	for _, p := range got {
		contentMap[p.Content] = true
	}
	for _, p := range posts {
		assert.True(t, contentMap[p.Content], "ReadAll() missing post with content: %s", p.Content)
	}

	// Test reading from non-existent file
	nonExistentStore := NewStoreWithPath(filepath.Join(t.TempDir(), "nonexistent.jsonl"))
	_, err = nonExistentStore.ReadAll()
	assert.Equal(t, config.ErrNotInitialized, err)
}

func TestStoreReadAllSkipsInvalidLines(t *testing.T) {
	store, feedPath := setupTestStore(t)

	// Write a valid post
	post := &Post{
		ID:        "smk-abc123",
		Author:    "ember",
		Suffix:    "smoke",
		Content:   "valid post",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	err := store.Append(post)
	require.NoError(t, err)

	// Manually append invalid JSON
	f, err := os.OpenFile(feedPath, os.O_APPEND|os.O_WRONLY, 0644)
	require.NoError(t, err)
	f.WriteString("invalid json line\n")
	f.WriteString("{\"id\":\"smk-def456\",\"author\":\"witness\",\"rig\":\"smoke\",\"content\":\"another valid\",\"created_at\":\"2026-01-30T10:00:00Z\"}\n")
	f.Close()

	// Read all - should skip invalid line
	posts, err := store.ReadAll()
	require.NoError(t, err)
	assert.Len(t, posts, 2)
}

func TestStoreReadRecent(t *testing.T) {
	store, _ := setupTestStore(t)

	// Write posts with different timestamps
	for i := 0; i < 10; i++ {
		post := &Post{
			ID:        "smk-" + string(rune('a'+i)) + "bcdef",
			Author:    "ember",
			Suffix:    "smoke",
			Content:   "post " + string(rune('0'+i)),
			CreatedAt: time.Now().Add(time.Duration(i) * time.Minute).UTC().Format(time.RFC3339),
		}
		err := store.Append(post)
		require.NoError(t, err)
	}

	// Read recent 5
	posts, err := store.ReadRecent(5)
	require.NoError(t, err)
	assert.Len(t, posts, 5)

	// Verify order (most recent first)
	for i := 1; i < len(posts); i++ {
		ti, _ := posts[i-1].GetCreatedTime()
		tj, _ := posts[i].GetCreatedTime()
		assert.False(t, ti.Before(tj), "ReadRecent() not sorted correctly at index %d", i)
	}
}

func TestStoreFindByID(t *testing.T) {
	store, _ := setupTestStore(t)

	post := &Post{
		ID:        "smk-target",
		Author:    "ember",
		Suffix:    "smoke",
		Content:   "target post",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	err := store.Append(post)
	require.NoError(t, err)

	// Find existing
	found, err := store.FindByID("smk-target")
	assert.NoError(t, err)
	assert.Equal(t, "target post", found.Content)

	// Find non-existing
	_, err = store.FindByID("smk-notfnd")
	assert.Equal(t, ErrPostNotFound, err)
}

func TestStoreExists(t *testing.T) {
	store, _ := setupTestStore(t)

	post := &Post{
		ID:        "smk-exists",
		Author:    "ember",
		Suffix:    "smoke",
		Content:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	err := store.Append(post)
	require.NoError(t, err)

	// Exists
	exists, err := store.Exists("smk-exists")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Does not exist
	exists, err = store.Exists("smk-notfnd")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestStoreCount(t *testing.T) {
	store, _ := setupTestStore(t)

	// Empty store
	count, err := store.Count()
	assert.NoError(t, err)
	assert.Equal(t, 0, count)

	// Add posts
	for i := 0; i < 5; i++ {
		post := &Post{
			ID:        "smk-" + string(rune('a'+i)) + "bcdef",
			Author:    "ember",
			Suffix:    "smoke",
			Content:   "post",
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		}
		store.Append(post)
	}

	count, err = store.Count()
	assert.NoError(t, err)
	assert.Equal(t, 5, count)
}

func TestStorePath(t *testing.T) {
	tmpDir := t.TempDir()
	feedPath := filepath.Join(tmpDir, "custom.jsonl")
	store := NewStoreWithPath(feedPath)

	assert.Equal(t, feedPath, store.Path())
}

func TestNewStore(t *testing.T) {
	// Create a temporary directory and feed file
	tmpDir := t.TempDir()
	feedPath := filepath.Join(tmpDir, "feed.jsonl")

	// Create the feed file
	err := os.WriteFile(feedPath, []byte{}, 0644)
	require.NoError(t, err)

	// Set SMOKE_FEED environment variable to point to test feed
	oldFeed := os.Getenv("SMOKE_FEED")
	defer func() {
		if oldFeed != "" {
			os.Setenv("SMOKE_FEED", oldFeed)
		} else {
			os.Unsetenv("SMOKE_FEED")
		}
	}()

	os.Setenv("SMOKE_FEED", feedPath)

	// Create store with NewStore()
	store, err := NewStore()
	require.NoError(t, err)

	// Verify store is not nil
	assert.NotNil(t, store)

	// Verify store has the correct path
	assert.Equal(t, feedPath, store.Path())

	// Verify store can perform basic operations
	post := &Post{
		ID:        "smk-abc123",
		Author:    "ember",
		Suffix:    "smoke",
		Content:   "test post",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	appendErr := store.Append(post)
	assert.NoError(t, appendErr)

	// Verify post was written
	posts, readErr := store.ReadAll()
	assert.NoError(t, readErr)
	assert.Len(t, posts, 1)
}

func TestCountWithPosts(t *testing.T) {
	tmpDir := t.TempDir()
	feedPath := tmpDir + "/feed.jsonl"

	// Create feed file
	err := os.WriteFile(feedPath, []byte{}, 0644)
	require.NoError(t, err)

	store := NewStoreWithPath(feedPath)

	// Add some posts with valid IDs
	for i := 0; i < 5; i++ {
		var post *Post
		post, err = NewPost("test-author", "smoke", "test", fmt.Sprintf("post %d", i))
		require.NoError(t, err)
		err = store.Append(post)
		require.NoError(t, err)
	}

	count, err := store.Count()
	assert.NoError(t, err)
	assert.Equal(t, 5, count)
}

func TestExistsTrue(t *testing.T) {
	tmpDir := t.TempDir()
	feedPath := tmpDir + "/feed.jsonl"

	// Create feed file
	err := os.WriteFile(feedPath, []byte{}, 0644)
	require.NoError(t, err)

	store := NewStoreWithPath(feedPath)

	// Add a post
	post, err := NewPost("test-author", "smoke", "test", "test content")
	require.NoError(t, err)
	appendErr := store.Append(post)
	require.NoError(t, appendErr)

	exists, existsErr := store.Exists(post.ID)
	assert.NoError(t, existsErr)
	assert.True(t, exists)
}

func TestSeedExamplesEmpty(t *testing.T) {
	store, _ := setupTestStore(t)

	// Seed empty feed
	count, err := store.SeedExamples()
	if err != nil {
		t.Fatalf("SeedExamples() unexpected error: %v", err)
	}

	// Should have seeded 4 posts
	if count != 4 {
		t.Errorf("SeedExamples() returned count %d, want 4", count)
	}

	// Verify posts were written
	posts, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll() unexpected error: %v", err)
	}
	if len(posts) != 4 {
		t.Errorf("ReadAll() returned %d posts, want 4", len(posts))
	}
}

func TestSeedExamplesNonEmpty(t *testing.T) {
	store, _ := setupTestStore(t)

	// Add an existing post first
	existingPost := &Post{
		ID:        "smk-abc123", // Valid 6-char alphanumeric ID
		Author:    "testuser",
		Suffix:    "test",
		Content:   "existing post",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if err := store.Append(existingPost); err != nil {
		t.Fatalf("Append() unexpected error: %v", err)
	}

	// Seed should do nothing
	count, err := store.SeedExamples()
	if err != nil {
		t.Fatalf("SeedExamples() unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("SeedExamples() returned count %d, want 0 (non-empty feed)", count)
	}

	// Verify only original post exists
	posts, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll() unexpected error: %v", err)
	}
	if len(posts) != 1 {
		t.Errorf("ReadAll() returned %d posts, want 1", len(posts))
	}
}

func TestSeedExamplesIdempotent(t *testing.T) {
	store, _ := setupTestStore(t)

	// Seed first time
	count1, err := store.SeedExamples()
	if err != nil {
		t.Fatalf("First SeedExamples() unexpected error: %v", err)
	}
	if count1 != 4 {
		t.Errorf("First SeedExamples() returned count %d, want 4", count1)
	}

	// Seed second time - should be no-op
	count2, err := store.SeedExamples()
	if err != nil {
		t.Fatalf("Second SeedExamples() unexpected error: %v", err)
	}
	if count2 != 0 {
		t.Errorf("Second SeedExamples() returned count %d, want 0", count2)
	}

	// Should still have exactly 4 posts
	posts, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll() unexpected error: %v", err)
	}
	if len(posts) != 4 {
		t.Errorf("ReadAll() returned %d posts after 2 seeds, want 4", len(posts))
	}
}

func TestSeedExamplesContent(t *testing.T) {
	store, _ := setupTestStore(t)

	_, err := store.SeedExamples()
	if err != nil {
		t.Fatalf("SeedExamples() unexpected error: %v", err)
	}

	posts, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll() unexpected error: %v", err)
	}

	// Verify authors
	expectedAuthors := map[string]bool{
		ExampleAuthorSpark: false,
		ExampleAuthorEmber: false,
		ExampleAuthorFlare: false,
		ExampleAuthorWisp:  false,
	}
	for _, p := range posts {
		if _, exists := expectedAuthors[p.Author]; exists {
			expectedAuthors[p.Author] = true
		}
		// Verify suffix
		if p.Suffix != ExampleSuffix {
			t.Errorf("Post suffix = %q, want %q", p.Suffix, ExampleSuffix)
		}
		// Verify ID format
		if len(p.ID) < 4 || p.ID[:4] != "smk-" {
			t.Errorf("Post ID %q doesn't have smk- prefix", p.ID)
		}
		// Verify content length within limits
		if len(p.Content) > MaxContentLength {
			t.Errorf("Post content length %d exceeds max %d", len(p.Content), MaxContentLength)
		}
	}

	// Verify all authors present
	for author, found := range expectedAuthors {
		if !found {
			t.Errorf("Expected author %q not found in seeded posts", author)
		}
	}
}

func TestSeedExamplesTimestamps(t *testing.T) {
	store, _ := setupTestStore(t)

	// Use UTC for consistent comparison
	before := time.Now().UTC().Add(-SeedPostsAgeOffset - time.Minute)
	_, err := store.SeedExamples()
	if err != nil {
		t.Fatalf("SeedExamples() unexpected error: %v", err)
	}
	// Add buffer for test execution time (4 posts, 1 minute apart = ~4 minutes total)
	after := time.Now().UTC().Add(-SeedPostsAgeOffset + 5*time.Minute)

	posts, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll() unexpected error: %v", err)
	}

	for i, p := range posts {
		ts, parseErr := time.Parse(time.RFC3339, p.CreatedAt)
		if parseErr != nil {
			t.Errorf("Post %d: failed to parse timestamp %q: %v", i, p.CreatedAt, parseErr)
			continue
		}
		tsUTC := ts.UTC()
		if tsUTC.Before(before) || tsUTC.After(after) {
			t.Errorf("Post %d: timestamp %v outside expected range [%v, %v]", i, tsUTC, before, after)
		}
	}
}

func TestGetExamplePosts(t *testing.T) {
	examples := GetExamplePosts()

	if len(examples) != 4 {
		t.Errorf("GetExamplePosts() returned %d examples, want 4", len(examples))
	}

	for i, ex := range examples {
		if ex.Author == "" {
			t.Errorf("Example %d has empty author", i)
		}
		if ex.Suffix == "" {
			t.Errorf("Example %d has empty suffix", i)
		}
		if ex.Content == "" {
			t.Errorf("Example %d has empty content", i)
		}
		if len(ex.Content) > MaxContentLength {
			t.Errorf("Example %d content length %d exceeds max %d", i, len(ex.Content), MaxContentLength)
		}
	}
}
