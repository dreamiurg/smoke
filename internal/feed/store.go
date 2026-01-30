package feed

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/dreamiurg/smoke/internal/config"
)

// ErrNotInitialized is returned when smoke is not initialized
var ErrNotInitialized = errors.New("smoke not initialized. Run 'smoke init' first")

// ErrPostNotFound is returned when a post is not found
var ErrPostNotFound = errors.New("post not found")

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

	// Validate post
	if err := post.Validate(); err != nil {
		return err
	}

	// Check if feed file exists
	if _, err := os.Stat(s.path); os.IsNotExist(err) {
		return ErrNotInitialized
	}

	// Open file for appending
	f, err := os.OpenFile(s.path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open feed file: %w", err)
	}
	defer f.Close()

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
	// Check if feed file exists
	if _, err := os.Stat(s.path); os.IsNotExist(err) {
		return nil, ErrNotInitialized
	}

	f, err := os.Open(s.path)
	if err != nil {
		return nil, fmt.Errorf("failed to open feed file: %w", err)
	}
	defer f.Close()

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
			fmt.Fprintf(os.Stderr, "warning: skipping invalid line %d: %v\n", lineNum, err)
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
		ti, _ := posts[i].GetCreatedTime()
		tj, _ := posts[j].GetCreatedTime()
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
