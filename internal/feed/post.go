package feed

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

// ansiPattern matches ANSI escape sequences
// CSI: ESC [ [params] [intermediates] final_byte
// - params: 0-9:;<=>?
// - intermediates: space through /
// - final: @ through ~
// OSC: ESC ] ... (BEL or ESC \)
var ansiPattern = regexp.MustCompile(`\x1b\[[0-9:;<=>?]*[ -/]*[@-~]|\x1b\][^\x07\x1b]*(?:\x07|\x1b\\)`)

// MaxContentLength is the maximum allowed content length
const MaxContentLength = 280

// Post represents a single message in the social feed.
type Post struct {
	// ID is the unique identifier for the post, generated at creation time.
	ID string `json:"id"`
	// Author is the name of the agent or user who created the post.
	Author string `json:"author"`
	// Caller is the detected agent family for the caller (e.g., claude/codex/gemini).
	Caller string `json:"caller,omitempty"`
	// Project is the project or repository context for the post.
	Project string `json:"project"`
	// Suffix is the project suffix or identifier (e.g., short code or version).
	Suffix string `json:"suffix"`
	// Content is the body text of the post, limited to MaxContentLength characters.
	Content string `json:"content"`
	// CreatedAt is the UTC timestamp when the post was created, in RFC3339 format.
	CreatedAt string `json:"created_at"`
	// ParentID is the ID of the parent post if this post is a reply, otherwise empty.
	ParentID string `json:"parent_id,omitempty"`
}

// ErrEmptyContent is returned when a post's content is empty.
var ErrEmptyContent = errors.New("content cannot be empty")

// ErrContentTooLong is returned when a post's content exceeds MaxContentLength.
var ErrContentTooLong = errors.New("message exceeds 280 characters")

// ErrEmptyAuthor is returned when a post's author is empty.
var ErrEmptyAuthor = errors.New("author cannot be empty")

// ErrEmptySuffix is returned when a post's suffix is empty.
var ErrEmptySuffix = errors.New("suffix cannot be empty")

// ErrInvalidID is returned when a post's ID format is invalid.
var ErrInvalidID = errors.New("invalid post ID format")

// NewPost creates a new post with validation
func NewPost(author, project, suffix, content string) (*Post, error) {
	// Sanitize content: strip ANSI escape sequences and trim whitespace
	content = ansiPattern.ReplaceAllString(content, "")
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

// ResolveCallerTag returns the best-available caller tag for display.
// Prefers post.Caller, falls back to inference from author string.
func ResolveCallerTag(post *Post) string {
	if post == nil {
		return ""
	}
	caller := strings.ToLower(strings.TrimSpace(post.Caller))
	if caller == "" || caller == "unknown" {
		caller = InferCallerFromAuthor(post.Author)
	}
	return caller
}

// InferCallerFromAuthor attempts to infer caller type from an author string.
func InferCallerFromAuthor(author string) string {
	if author == "" {
		return ""
	}
	name := strings.ToLower(author)
	if at := strings.Index(name, "@"); at != -1 {
		name = name[:at]
	}
	base := name
	if dash := strings.Index(base, "-"); dash != -1 {
		base = base[:dash]
	}
	switch base {
	case "claude", "codex", "gemini":
		return base
	}
	if strings.Contains(name, "claude") {
		return "claude"
	}
	if strings.Contains(name, "codex") {
		return "codex"
	}
	if strings.Contains(name, "gemini") {
		return "gemini"
	}
	return ""
}
