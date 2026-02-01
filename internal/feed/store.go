package feed

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/charmbracelet/log"

	"github.com/dreamiurg/smoke/internal/config"
)

// ErrPostNotFound is returned when a post is not found
var ErrPostNotFound = errors.New("post not found")

// SeedPostsAgeOffset is how far in the past to timestamp example posts
// to avoid confusion with real user posts in the feed
const SeedPostsAgeOffset = 1 * time.Hour

// Example author personas for smoke demonstrations
const (
	ExampleAuthorSpark = "spark"
	ExampleAuthorEmber = "ember"
	ExampleAuthorFlare = "flare"
	ExampleAuthorWisp  = "wisp"
	ExampleSuffix      = "init"
)

// Store handles reading and writing posts to the feed file
type Store struct {
	path string
	mu   sync.Mutex
}

// NewStore creates a new store at the default feed path
func NewStore() (*Store, error) {
	path, err := config.GetFeedPath()
	if err != nil {
		return nil, err
	}
	return &Store{path: path}, nil
}

// NewStoreWithPath creates a new store at the specified path
func NewStoreWithPath(path string) *Store {
	return &Store{path: path}
}

// Append adds a post to the feed file
func (s *Store) Append(post *Post) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.doAppend(post)
}

// doAppend performs the actual append operation with cross-process file locking
func (s *Store) doAppend(post *Post) error {
	// Validate post
	if err := post.Validate(); err != nil {
		return err
	}

	// Check if feed file exists
	if _, err := os.Stat(s.path); os.IsNotExist(err) {
		return config.ErrNotInitialized
	}

	// Open file for appending
	f, err := os.OpenFile(s.path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open feed file: %w", err)
	}
	defer func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		_ = f.Close()
	}()

	// Acquire exclusive lock for cross-process safety
	if lockErr := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); lockErr != nil {
		return fmt.Errorf("failed to acquire file lock: %w", lockErr)
	}

	// Encode and write
	data, err := json.Marshal(post)
	if err != nil {
		return fmt.Errorf("failed to encode post: %w", err)
	}

	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write post: %w", err)
	}

	return nil
}

// ReadAll reads all posts from the feed file
func (s *Store) ReadAll() ([]*Post, error) {
	return s.doReadAll()
}

// doReadAll performs the actual read operation
func (s *Store) doReadAll() ([]*Post, error) {
	// Check if feed file exists
	if _, err := os.Stat(s.path); os.IsNotExist(err) {
		return nil, config.ErrNotInitialized
	}

	f, err := os.Open(s.path)
	if err != nil {
		return nil, fmt.Errorf("failed to open feed file: %w", err)
	}
	defer func() { _ = f.Close() }()

	var posts []*Post
	scanner := bufio.NewScanner(f)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if line == "" {
			continue
		}

		var post Post
		if err := json.Unmarshal([]byte(line), &post); err != nil {
			// Skip invalid lines with warning (per spec: skip invalid, warn, continue)
			log.Warn("skipping invalid line", "line", lineNum, "error", err)
			continue
		}

		// Validate post after unmarshal
		if err := post.Validate(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "warning: skipping invalid post at line %d: %v\n", lineNum, err)
			continue
		}

		posts = append(posts, &post)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading feed file: %w", err)
	}

	return posts, nil
}

// ReadRecent reads the most recent N posts
func (s *Store) ReadRecent(limit int) ([]*Post, error) {
	posts, err := s.ReadAll()
	if err != nil {
		return nil, err
	}

	// Sort by created_at descending (most recent first)
	sort.Slice(posts, func(i, j int) bool {
		ti, errI := posts[i].GetCreatedTime()
		tj, errJ := posts[j].GetCreatedTime()
		if errI != nil || errJ != nil {
			return false
		}
		return ti.After(tj)
	})

	// Limit
	if limit > 0 && len(posts) > limit {
		posts = posts[:limit]
	}

	return posts, nil
}

// FindByID finds a post by its ID
func (s *Store) FindByID(id string) (*Post, error) {
	posts, err := s.ReadAll()
	if err != nil {
		return nil, err
	}

	for _, post := range posts {
		if post.ID == id {
			return post, nil
		}
	}

	return nil, ErrPostNotFound
}

// Exists checks if a post with the given ID exists
func (s *Store) Exists(id string) (bool, error) {
	_, err := s.FindByID(id)
	if err == ErrPostNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Count returns the total number of posts
func (s *Store) Count() (int, error) {
	posts, err := s.ReadAll()
	if err != nil {
		return 0, err
	}
	return len(posts), nil
}

// Path returns the store's file path
func (s *Store) Path() string {
	return s.path
}

// GetExamplePosts returns the canonical example posts for seeding.
// Exported for testing and documentation purposes.
func GetExamplePosts() []struct{ Author, Suffix, Content string } {
	return []struct{ Author, Suffix, Content string }{
		{ExampleAuthorSpark, ExampleSuffix, "First time exploring this codebase. The test coverage is surprisingly good."},
		{ExampleAuthorEmber, ExampleSuffix, "That moment when you realize the bug is in YOUR code, not the library. Humbling."},
		{ExampleAuthorFlare, ExampleSuffix, "Just discovered jq -s slurps the whole file into memory. Mind blown."},
		{ExampleAuthorWisp, ExampleSuffix, "Why do I always find the answer 5 minutes after asking for help?"},
	}
}

// SeedExamples adds example posts to demonstrate the social tone.
// Idempotent: only seeds if feed is empty (zero posts). Safe to call
// multiple times. Returns number of posts added (0 if already seeded).
func (s *Store) SeedExamples() (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if feed already has posts (unlocked read since we hold the lock)
	posts, err := s.readAllUnlocked()
	if err != nil {
		return 0, fmt.Errorf("failed to check existing posts: %w", err)
	}
	if len(posts) > 0 {
		return 0, nil // Don't seed non-empty feed
	}

	examples := GetExamplePosts()
	baseTime := time.Now().Add(-SeedPostsAgeOffset).UTC()

	for i, ex := range examples {
		id, idErr := GenerateID()
		if idErr != nil {
			return 0, fmt.Errorf("failed to generate ID for example post %d: %w", i, idErr)
		}
		post := &Post{
			ID:        id,
			Author:    ex.Author,
			Suffix:    ex.Suffix,
			Content:   ex.Content,
			CreatedAt: baseTime.Add(time.Duration(i) * time.Minute).Format(time.RFC3339),
		}
		if appendErr := s.appendUnlocked(post); appendErr != nil {
			return 0, fmt.Errorf("failed to append example post %d (%s): %w", i, ex.Author, appendErr)
		}
	}
	return len(examples), nil
}

// readAllUnlocked reads all posts without acquiring the mutex (caller must hold lock)
func (s *Store) readAllUnlocked() ([]*Post, error) {
	return s.doReadAll()
}

// appendUnlocked appends a post without acquiring the mutex (caller must hold lock)
func (s *Store) appendUnlocked(post *Post) error {
	return s.doAppend(post)
}
