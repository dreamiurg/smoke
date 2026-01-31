package feed

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestStore(t *testing.T) (*Store, string) {
	t.Helper()

	tmpDir := t.TempDir()
	feedPath := filepath.Join(tmpDir, "feed.jsonl")

	// Create empty feed file
	if err := os.WriteFile(feedPath, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create feed file: %v", err)
	}

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

	if err := store.Append(post); err != nil {
		t.Errorf("Append() unexpected error: %v", err)
	}

	// Verify post was written
	posts, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll() unexpected error: %v", err)
	}
	if len(posts) != 1 {
		t.Errorf("ReadAll() returned %d posts, want 1", len(posts))
	}
	if posts[0].ID != post.ID {
		t.Errorf("ReadAll()[0].ID = %v, want %v", posts[0].ID, post.ID)
	}
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

	if err := store.Append(invalidPost); err == nil {
		t.Error("Append() expected error for invalid post")
	}
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
	if err != ErrNotInitialized {
		t.Errorf("Append() error = %v, want ErrNotInitialized", err)
	}
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
		if err := store.Append(p); err != nil {
			t.Fatalf("Append() unexpected error: %v", err)
		}
	}

	// Read all
	got, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll() unexpected error: %v", err)
	}
	if len(got) != 3 {
		t.Errorf("ReadAll() returned %d posts, want 3", len(got))
	}

	// Verify content
	contentMap := make(map[string]bool)
	for _, p := range got {
		contentMap[p.Content] = true
	}
	for _, p := range posts {
		if !contentMap[p.Content] {
			t.Errorf("ReadAll() missing post with content: %s", p.Content)
		}
	}

	// Test reading from non-existent file
	nonExistentStore := NewStoreWithPath(filepath.Join(t.TempDir(), "nonexistent.jsonl"))
	_, err = nonExistentStore.ReadAll()
	if err != ErrNotInitialized {
		t.Errorf("ReadAll() error = %v, want ErrNotInitialized", err)
	}
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
	if err := store.Append(post); err != nil {
		t.Fatalf("Append() unexpected error: %v", err)
	}

	// Manually append invalid JSON
	f, err := os.OpenFile(feedPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to open feed file: %v", err)
	}
	f.WriteString("invalid json line\n")
	f.WriteString("{\"id\":\"smk-def456\",\"author\":\"witness\",\"rig\":\"smoke\",\"content\":\"another valid\",\"created_at\":\"2026-01-30T10:00:00Z\"}\n")
	f.Close()

	// Read all - should skip invalid line
	posts, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll() unexpected error: %v", err)
	}
	if len(posts) != 2 {
		t.Errorf("ReadAll() returned %d posts, want 2 (skipping invalid)", len(posts))
	}
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
		if err := store.Append(post); err != nil {
			t.Fatalf("Append() unexpected error: %v", err)
		}
	}

	// Read recent 5
	posts, err := store.ReadRecent(5)
	if err != nil {
		t.Fatalf("ReadRecent() unexpected error: %v", err)
	}
	if len(posts) != 5 {
		t.Errorf("ReadRecent(5) returned %d posts, want 5", len(posts))
	}

	// Verify order (most recent first)
	for i := 1; i < len(posts); i++ {
		ti, _ := posts[i-1].GetCreatedTime()
		tj, _ := posts[i].GetCreatedTime()
		if ti.Before(tj) {
			t.Errorf("ReadRecent() not sorted correctly at index %d", i)
		}
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
	if err := store.Append(post); err != nil {
		t.Fatalf("Append() unexpected error: %v", err)
	}

	// Find existing
	found, err := store.FindByID("smk-target")
	if err != nil {
		t.Errorf("FindByID() unexpected error: %v", err)
	}
	if found.Content != "target post" {
		t.Errorf("FindByID().Content = %v, want 'target post'", found.Content)
	}

	// Find non-existing
	_, err = store.FindByID("smk-notfnd")
	if err != ErrPostNotFound {
		t.Errorf("FindByID() error = %v, want ErrPostNotFound", err)
	}
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
	if err := store.Append(post); err != nil {
		t.Fatalf("Append() unexpected error: %v", err)
	}

	// Exists
	exists, err := store.Exists("smk-exists")
	if err != nil {
		t.Errorf("Exists() unexpected error: %v", err)
	}
	if !exists {
		t.Error("Exists() = false, want true")
	}

	// Does not exist
	exists, err = store.Exists("smk-notfnd")
	if err != nil {
		t.Errorf("Exists() unexpected error: %v", err)
	}
	if exists {
		t.Error("Exists() = true, want false")
	}
}

func TestStoreCount(t *testing.T) {
	store, _ := setupTestStore(t)

	// Empty store
	count, err := store.Count()
	if err != nil {
		t.Errorf("Count() unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("Count() = %d, want 0", count)
	}

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
	if err != nil {
		t.Errorf("Count() unexpected error: %v", err)
	}
	if count != 5 {
		t.Errorf("Count() = %d, want 5", count)
	}
}

func TestStorePath(t *testing.T) {
	tmpDir := t.TempDir()
	feedPath := filepath.Join(tmpDir, "custom.jsonl")
	store := NewStoreWithPath(feedPath)

	if store.Path() != feedPath {
		t.Errorf("Path() = %v, want %v", store.Path(), feedPath)
	}
}
