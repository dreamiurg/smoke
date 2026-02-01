package feed

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
		{
			name:    "content with ANSI stripped",
			author:  "ember",
			project: "smoke",
			rig:     "swift-fox",
			content: "\x1b[2J\x1b[Hhello world",
			wantErr: nil,
		},
		{
			name:    "content only ANSI becomes empty",
			author:  "ember",
			project: "smoke",
			rig:     "swift-fox",
			content: "\x1b[2J\x1b[H",
			wantErr: ErrEmptyContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			post, err := NewPost(tt.author, tt.project, tt.rig, tt.content)
			assert.Equal(t, tt.wantErr, err)
			if tt.wantErr != nil {
				return
			}

			// Verify post fields
			assert.Equal(t, tt.author, post.Author)
			assert.Equal(t, tt.project, post.Project)
			assert.Equal(t, tt.rig, post.Suffix)
			assert.True(t, ValidateID(post.ID))
			assert.NotEmpty(t, post.CreatedAt)
			assert.Empty(t, post.ParentID)
		})
	}
}

func TestNewPost_ANSISanitization(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantContent string
	}{
		{
			name:        "clear screen sequence stripped",
			content:     "\x1b[2J\x1b[Hhello",
			wantContent: "hello",
		},
		{
			name:        "color codes stripped",
			content:     "\x1b[31mred\x1b[0m text",
			wantContent: "red text",
		},
		{
			name:        "terminal title stripped",
			content:     "\x1b]0;Malicious Title\x07hello",
			wantContent: "hello",
		},
		{
			name:        "no ANSI unchanged",
			content:     "hello world",
			wantContent: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			post, err := NewPost("ember", "smoke", "swift-fox", tt.content)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantContent, post.Content)
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
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.parentID, reply.ParentID)
		})
	}
}

func TestPostValidate(t *testing.T) {
	validPost := &Post{
		ID:        "smk-abc123",
		Author:    "ember",
		Project:   "smoke",
		Suffix:    "swift-fox",
		Content:   "hello",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	assert.NoError(t, validPost.Validate())

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
				Suffix:    "swift-fox",
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
				Suffix:    "swift-fox",
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
				Suffix:    "swift-fox",
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
				Suffix:    "",
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
				Suffix:    "swift-fox",
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
				Suffix:    "swift-fox",
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
				Suffix:    "swift-fox",
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
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestPostIsReply(t *testing.T) {
	post := &Post{
		ID:        "smk-abc123",
		Author:    "ember",
		Project:   "smoke",
		Suffix:    "swift-fox",
		Content:   "hello",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	assert.False(t, post.IsReply())

	reply := &Post{
		ID:        "smk-def456",
		Author:    "witness",
		Project:   "smoke",
		Suffix:    "swift-fox",
		Content:   "nice!",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		ParentID:  "smk-abc123",
	}

	assert.True(t, reply.IsReply())
}

func TestPostGetCreatedTime(t *testing.T) {
	now := time.Now().UTC()
	post := &Post{
		ID:        "smk-abc123",
		Author:    "ember",
		Project:   "smoke",
		Suffix:    "swift-fox",
		Content:   "hello",
		CreatedAt: now.Format(time.RFC3339),
	}

	got, err := post.GetCreatedTime()
	assert.NoError(t, err)

	// Compare to second precision
	assert.Equal(t, now.Unix(), got.Unix())

	// Test invalid timestamp
	invalidPost := &Post{
		ID:        "smk-abc123",
		Author:    "ember",
		Project:   "smoke",
		Suffix:    "swift-fox",
		Content:   "hello",
		CreatedAt: "invalid",
	}

	_, err = invalidPost.GetCreatedTime()
	assert.Error(t, err)
}

func TestPostContentLength(t *testing.T) {
	post := &Post{
		ID:        "smk-abc123",
		Author:    "ember",
		Project:   "smoke",
		Suffix:    "swift-fox",
		Content:   "hello world",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	assert.Equal(t, 11, post.ContentLength())
}
