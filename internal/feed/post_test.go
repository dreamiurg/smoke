package feed

import (
	"strings"
	"testing"
	"time"
)

func TestNewPost(t *testing.T) {
	tests := []struct {
		name    string
		author  string
		project string
		rig     string
		content string
		wantErr error
	}{
		{
			name:    "valid post",
			author:  "ember",
			project: "smoke",
			rig:     "swift-fox",
			content: "hello world",
			wantErr: nil,
		},
		{
			name:    "empty content",
			author:  "ember",
			project: "smoke",
			rig:     "swift-fox",
			content: "",
			wantErr: ErrEmptyContent,
		},
		{
			name:    "whitespace only content",
			author:  "ember",
			project: "smoke",
			rig:     "swift-fox",
			content: "   ",
			wantErr: ErrEmptyContent,
		},
		{
			name:    "content too long",
			author:  "ember",
			project: "smoke",
			rig:     "swift-fox",
			content: strings.Repeat("a", 281),
			wantErr: ErrContentTooLong,
		},
		{
			name:    "max length content",
			author:  "ember",
			project: "smoke",
			rig:     "swift-fox",
			content: strings.Repeat("a", 280),
			wantErr: nil,
		},
		{
			name:    "empty author",
			author:  "",
			project: "smoke",
			rig:     "swift-fox",
			content: "hello",
			wantErr: ErrEmptyAuthor,
		},
		{
			name:    "empty rig",
			author:  "ember",
			project: "smoke",
			rig:     "",
			content: "hello",
			wantErr: ErrEmptySuffix,
		},
		{
			name:    "content with whitespace trimmed",
			author:  "ember",
			project: "smoke",
			rig:     "swift-fox",
			content: "  hello world  ",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			post, err := NewPost(tt.author, tt.project, tt.rig, tt.content)
			if err != tt.wantErr {
				t.Errorf("NewPost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr != nil {
				return
			}

			// Verify post fields
			if post.Author != tt.author {
				t.Errorf("NewPost().Author = %v, want %v", post.Author, tt.author)
			}
			if post.Project != tt.project {
				t.Errorf("NewPost().Project = %v, want %v", post.Project, tt.project)
			}
			if post.Suffix != tt.rig {
				t.Errorf("NewPost().Rig = %v, want %v", post.Suffix, tt.rig)
			}
			if !ValidateID(post.ID) {
				t.Errorf("NewPost().ID = %v, invalid format", post.ID)
			}
			if post.CreatedAt == "" {
				t.Error("NewPost().CreatedAt is empty")
			}
			if post.ParentID != "" {
				t.Errorf("NewPost().ParentID = %v, want empty", post.ParentID)
			}
		})
	}
}

func TestNewReply(t *testing.T) {
	tests := []struct {
		name     string
		author   string
		project  string
		rig      string
		content  string
		parentID string
		wantErr  bool
	}{
		{
			name:     "valid reply",
			author:   "witness",
			project:  "smoke",
			rig:      "swift-fox",
			content:  "nice!",
			parentID: "smk-abc123",
			wantErr:  false,
		},
		{
			name:     "invalid parent ID",
			author:   "witness",
			project:  "smoke",
			rig:      "swift-fox",
			content:  "nice!",
			parentID: "invalid",
			wantErr:  true,
		},
		{
			name:     "empty parent ID",
			author:   "witness",
			project:  "smoke",
			rig:      "swift-fox",
			content:  "nice!",
			parentID: "",
			wantErr:  true,
		},
		{
			name:     "content too long",
			author:   "witness",
			project:  "smoke",
			rig:      "swift-fox",
			content:  strings.Repeat("a", 281),
			parentID: "smk-abc123",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reply, err := NewReply(tt.author, tt.project, tt.rig, tt.content, tt.parentID)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewReply() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if reply.ParentID != tt.parentID {
				t.Errorf("NewReply().ParentID = %v, want %v", reply.ParentID, tt.parentID)
			}
		})
	}
}

func TestPostValidate(t *testing.T) {
	validPost := &Post{
		ID:        "smk-abc123",
		Author:    "ember",
		Project:   "smoke",
		Suffix:       "swift-fox",
		Content:   "hello",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	if err := validPost.Validate(); err != nil {
		t.Errorf("Validate() on valid post returned error: %v", err)
	}

	tests := []struct {
		name    string
		post    *Post
		wantErr error
	}{
		{
			name: "invalid ID",
			post: &Post{
				ID:        "invalid",
				Author:    "ember",
				Project:   "smoke",
				Suffix:       "swift-fox",
				Content:   "hello",
				CreatedAt: time.Now().UTC().Format(time.RFC3339),
			},
			wantErr: ErrInvalidID,
		},
		{
			name: "empty ID",
			post: &Post{
				ID:        "",
				Author:    "ember",
				Project:   "smoke",
				Suffix:       "swift-fox",
				Content:   "hello",
				CreatedAt: time.Now().UTC().Format(time.RFC3339),
			},
			wantErr: ErrInvalidID,
		},
		{
			name: "empty author",
			post: &Post{
				ID:        "smk-abc123",
				Author:    "",
				Project:   "smoke",
				Suffix:       "swift-fox",
				Content:   "hello",
				CreatedAt: time.Now().UTC().Format(time.RFC3339),
			},
			wantErr: ErrEmptyAuthor,
		},
		{
			name: "empty rig",
			post: &Post{
				ID:        "smk-abc123",
				Author:    "ember",
				Project:   "smoke",
				Suffix:       "",
				Content:   "hello",
				CreatedAt: time.Now().UTC().Format(time.RFC3339),
			},
			wantErr: ErrEmptySuffix,
		},
		{
			name: "empty content",
			post: &Post{
				ID:        "smk-abc123",
				Author:    "ember",
				Project:   "smoke",
				Suffix:       "swift-fox",
				Content:   "",
				CreatedAt: time.Now().UTC().Format(time.RFC3339),
			},
			wantErr: ErrEmptyContent,
		},
		{
			name: "content too long",
			post: &Post{
				ID:        "smk-abc123",
				Author:    "ember",
				Project:   "smoke",
				Suffix:       "swift-fox",
				Content:   strings.Repeat("a", 281),
				CreatedAt: time.Now().UTC().Format(time.RFC3339),
			},
			wantErr: ErrContentTooLong,
		},
		{
			name: "invalid parent ID",
			post: &Post{
				ID:        "smk-abc123",
				Author:    "ember",
				Project:   "smoke",
				Suffix:       "swift-fox",
				Content:   "hello",
				CreatedAt: time.Now().UTC().Format(time.RFC3339),
				ParentID:  "invalid",
			},
			wantErr: ErrInvalidID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.post.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPostIsReply(t *testing.T) {
	post := &Post{
		ID:        "smk-abc123",
		Author:    "ember",
		Project:   "smoke",
		Suffix:       "swift-fox",
		Content:   "hello",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	if post.IsReply() {
		t.Error("IsReply() = true for non-reply post")
	}

	reply := &Post{
		ID:        "smk-def456",
		Author:    "witness",
		Project:   "smoke",
		Suffix:       "swift-fox",
		Content:   "nice!",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		ParentID:  "smk-abc123",
	}

	if !reply.IsReply() {
		t.Error("IsReply() = false for reply post")
	}
}

func TestPostGetCreatedTime(t *testing.T) {
	now := time.Now().UTC()
	post := &Post{
		ID:        "smk-abc123",
		Author:    "ember",
		Project:   "smoke",
		Suffix:       "swift-fox",
		Content:   "hello",
		CreatedAt: now.Format(time.RFC3339),
	}

	got, err := post.GetCreatedTime()
	if err != nil {
		t.Errorf("GetCreatedTime() unexpected error: %v", err)
	}

	// Compare to second precision
	if got.Unix() != now.Unix() {
		t.Errorf("GetCreatedTime() = %v, want %v", got, now)
	}

	// Test invalid timestamp
	invalidPost := &Post{
		ID:        "smk-abc123",
		Author:    "ember",
		Project:   "smoke",
		Suffix:       "swift-fox",
		Content:   "hello",
		CreatedAt: "invalid",
	}

	_, err = invalidPost.GetCreatedTime()
	if err == nil {
		t.Error("GetCreatedTime() expected error for invalid timestamp")
	}
}

func TestPostContentLength(t *testing.T) {
	post := &Post{
		ID:        "smk-abc123",
		Author:    "ember",
		Project:   "smoke",
		Suffix:       "swift-fox",
		Content:   "hello world",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	if post.ContentLength() != 11 {
		t.Errorf("ContentLength() = %d, want 11", post.ContentLength())
	}
}
