package feed

import (
	"errors"
	"strings"
	"time"
)

// MaxContentLength is the maximum allowed content length
const MaxContentLength = 280

// Post represents a single message in the social feed
type Post struct {
	ID        string `json:"id"`
	Author    string `json:"author"`
	Project   string `json:"project"`
	Suffix    string `json:"suffix"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	ParentID  string `json:"parent_id,omitempty"`
}

// Validation errors
var (
	ErrEmptyContent   = errors.New("content cannot be empty")
	ErrContentTooLong = errors.New("message exceeds 280 characters")
	ErrEmptyAuthor    = errors.New("author cannot be empty")
	ErrEmptySuffix    = errors.New("suffix cannot be empty")
	ErrInvalidID      = errors.New("invalid post ID format")
)

// NewPost creates a new post with validation
func NewPost(author, project, suffix, content string) (*Post, error) {
	// Trim content
	content = strings.TrimSpace(content)

	// Validate content
	if content == "" {
		return nil, ErrEmptyContent
	}
	if len(content) > MaxContentLength {
		return nil, ErrContentTooLong
	}

	// Validate author
	author = strings.TrimSpace(author)
	if author == "" {
		return nil, ErrEmptyAuthor
	}

	// Validate project
	project = strings.TrimSpace(project)

	// Validate suffix
	suffix = strings.TrimSpace(suffix)
	if suffix == "" {
		return nil, ErrEmptySuffix
	}

	// Generate ID
	id, err := GenerateID()
	if err != nil {
		return nil, err
	}

	return &Post{
		ID:        id,
		Author:    author,
		Project:   project,
		Suffix:    suffix,
		Content:   content,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// NewReply creates a new reply post with validation
func NewReply(author, project, suffix, content, parentID string) (*Post, error) {
	post, err := NewPost(author, project, suffix, content)
	if err != nil {
		return nil, err
	}

	// Validate parent ID
	if !ValidateID(parentID) {
		return nil, ErrInvalidID
	}

	post.ParentID = parentID
	return post, nil
}

// Validate checks if a post has valid data
func (p *Post) Validate() error {
	if p.ID == "" || !ValidateID(p.ID) {
		return ErrInvalidID
	}
	if p.Author == "" {
		return ErrEmptyAuthor
	}
	if p.Suffix == "" {
		return ErrEmptySuffix
	}
	if p.Content == "" {
		return ErrEmptyContent
	}
	if len(p.Content) > MaxContentLength {
		return ErrContentTooLong
	}
	if p.ParentID != "" && !ValidateID(p.ParentID) {
		return ErrInvalidID
	}
	return nil
}

// IsReply returns true if this post is a reply
func (p *Post) IsReply() bool {
	return p.ParentID != ""
}

// GetCreatedTime parses and returns the CreatedAt timestamp
func (p *Post) GetCreatedTime() (time.Time, error) {
	return time.Parse(time.RFC3339, p.CreatedAt)
}

// ContentLength returns the length of the content
func (p *Post) ContentLength() int {
	return len(p.Content)
}
