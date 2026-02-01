package feed

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFilterRecent(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name      string
		posts     []*Post
		duration  time.Duration
		wantCount int
		wantIDs   []string
		wantErr   bool
	}{
		{
			name:      "empty feed returns empty result",
			posts:     []*Post{},
			duration:  2 * time.Hour,
			wantCount: 0,
			wantIDs:   []string{},
			wantErr:   false,
		},
		{
			name: "single post within 2 hour window",
			posts: []*Post{
				{
					ID:        "smk-post1",
					Author:    "ember",
					Project:   "smoke",
					Suffix:    "swift-fox",
					Content:   "recent post",
					CreatedAt: now.Add(-1 * time.Hour).Format(time.RFC3339),
				},
			},
			duration:  2 * time.Hour,
			wantCount: 1,
			wantIDs:   []string{"smk-post1"},
			wantErr:   false,
		},
		{
			name: "post exactly at boundary (2 hours ago)",
			posts: []*Post{
				{
					ID:        "smk-post2",
					Author:    "ember",
					Project:   "smoke",
					Suffix:    "swift-fox",
					Content:   "boundary post",
					CreatedAt: now.Add(-2 * time.Hour).Format(time.RFC3339),
				},
			},
			duration:  2 * time.Hour,
			wantCount: 1,
			wantIDs:   []string{"smk-post2"},
			wantErr:   false,
		},
		{
			name: "post outside window (beyond 2 hours)",
			posts: []*Post{
				{
					ID:        "smk-post3",
					Author:    "ember",
					Project:   "smoke",
					Suffix:    "swift-fox",
					Content:   "old post",
					CreatedAt: now.Add(-3 * time.Hour).Format(time.RFC3339),
				},
			},
			duration:  2 * time.Hour,
			wantCount: 0,
			wantIDs:   []string{},
			wantErr:   false,
		},
		{
			name: "mixed posts - some within window, some outside",
			posts: []*Post{
				{
					ID:        "smk-post4",
					Author:    "ember",
					Project:   "smoke",
					Suffix:    "swift-fox",
					Content:   "recent post",
					CreatedAt: now.Add(-30 * time.Minute).Format(time.RFC3339),
				},
				{
					ID:        "smk-post5",
					Author:    "witness",
					Project:   "smoke",
					Suffix:    "swift-fox",
					Content:   "old post",
					CreatedAt: now.Add(-3 * time.Hour).Format(time.RFC3339),
				},
				{
					ID:        "smk-post6",
					Author:    "ember",
					Project:   "smoke",
					Suffix:    "swift-fox",
					Content:   "recent post 2",
					CreatedAt: now.Add(-1 * time.Hour).Format(time.RFC3339),
				},
			},
			duration:  2 * time.Hour,
			wantCount: 2,
			wantIDs:   []string{"smk-post4", "smk-post6"},
			wantErr:   false,
		},
		{
			name: "future posts excluded",
			posts: []*Post{
				{
					ID:        "smk-post7",
					Author:    "ember",
					Project:   "smoke",
					Suffix:    "swift-fox",
					Content:   "future post",
					CreatedAt: now.Add(1 * time.Hour).Format(time.RFC3339),
				},
				{
					ID:        "smk-post8",
					Author:    "witness",
					Project:   "smoke",
					Suffix:    "swift-fox",
					Content:   "valid post",
					CreatedAt: now.Add(-30 * time.Minute).Format(time.RFC3339),
				},
			},
			duration:  2 * time.Hour,
			wantCount: 1,
			wantIDs:   []string{"smk-post8"},
			wantErr:   false,
		},
		{
			name: "1 hour window",
			posts: []*Post{
				{
					ID:        "smk-post9",
					Author:    "ember",
					Project:   "smoke",
					Suffix:    "swift-fox",
					Content:   "post within 1h",
					CreatedAt: now.Add(-30 * time.Minute).Format(time.RFC3339),
				},
				{
					ID:        "smk-post10",
					Author:    "witness",
					Project:   "smoke",
					Suffix:    "swift-fox",
					Content:   "post outside 1h",
					CreatedAt: now.Add(-2 * time.Hour).Format(time.RFC3339),
				},
			},
			duration:  1 * time.Hour,
			wantCount: 1,
			wantIDs:   []string{"smk-post9"},
			wantErr:   false,
		},
		{
			name: "6 hour window",
			posts: []*Post{
				{
					ID:        "smk-post11",
					Author:    "ember",
					Project:   "smoke",
					Suffix:    "swift-fox",
					Content:   "post within 6h",
					CreatedAt: now.Add(-3 * time.Hour).Format(time.RFC3339),
				},
				{
					ID:        "smk-post12",
					Author:    "witness",
					Project:   "smoke",
					Suffix:    "swift-fox",
					Content:   "post outside 6h",
					CreatedAt: now.Add(-7 * time.Hour).Format(time.RFC3339),
				},
			},
			duration:  6 * time.Hour,
			wantCount: 1,
			wantIDs:   []string{"smk-post11"},
			wantErr:   false,
		},
		{
			name: "24 hour window",
			posts: []*Post{
				{
					ID:        "smk-post13",
					Author:    "ember",
					Project:   "smoke",
					Suffix:    "swift-fox",
					Content:   "post within 24h",
					CreatedAt: now.Add(-12 * time.Hour).Format(time.RFC3339),
				},
				{
					ID:        "smk-post14",
					Author:    "witness",
					Project:   "smoke",
					Suffix:    "swift-fox",
					Content:   "post outside 24h",
					CreatedAt: now.Add(-25 * time.Hour).Format(time.RFC3339),
				},
			},
			duration:  24 * time.Hour,
			wantCount: 1,
			wantIDs:   []string{"smk-post13"},
			wantErr:   false,
		},
		{
			name: "sorted by timestamp newest first",
			posts: []*Post{
				{
					ID:        "smk-old",
					Author:    "ember",
					Project:   "smoke",
					Suffix:    "swift-fox",
					Content:   "oldest",
					CreatedAt: now.Add(-2 * time.Hour).Format(time.RFC3339),
				},
				{
					ID:        "smk-newest",
					Author:    "witness",
					Project:   "smoke",
					Suffix:    "swift-fox",
					Content:   "newest",
					CreatedAt: now.Add(-1 * time.Minute).Format(time.RFC3339),
				},
				{
					ID:        "smk-middle",
					Author:    "ember",
					Project:   "smoke",
					Suffix:    "swift-fox",
					Content:   "middle",
					CreatedAt: now.Add(-1 * time.Hour).Format(time.RFC3339),
				},
			},
			duration:  2 * time.Hour,
			wantCount: 3,
			wantIDs:   []string{"smk-newest", "smk-middle", "smk-old"},
			wantErr:   false,
		},
		{
			name: "all posts in feed within window",
			posts: []*Post{
				{
					ID:        "smk-p1",
					Author:    "ember",
					Project:   "smoke",
					Suffix:    "swift-fox",
					Content:   "post 1",
					CreatedAt: now.Add(-30 * time.Minute).Format(time.RFC3339),
				},
				{
					ID:        "smk-p2",
					Author:    "witness",
					Project:   "smoke",
					Suffix:    "swift-fox",
					Content:   "post 2",
					CreatedAt: now.Add(-45 * time.Minute).Format(time.RFC3339),
				},
				{
					ID:        "smk-p3",
					Author:    "ember",
					Project:   "smoke",
					Suffix:    "swift-fox",
					Content:   "post 3",
					CreatedAt: now.Add(-1 * time.Hour).Format(time.RFC3339),
				},
			},
			duration:  2 * time.Hour,
			wantCount: 3,
			wantIDs:   []string{"smk-p1", "smk-p2", "smk-p3"},
			wantErr:   false,
		},
		{
			name: "no posts in feed within window",
			posts: []*Post{
				{
					ID:        "smk-old1",
					Author:    "ember",
					Project:   "smoke",
					Suffix:    "swift-fox",
					Content:   "very old post",
					CreatedAt: now.Add(-4 * time.Hour).Format(time.RFC3339),
				},
				{
					ID:        "smk-old2",
					Author:    "witness",
					Project:   "smoke",
					Suffix:    "swift-fox",
					Content:   "another old post",
					CreatedAt: now.Add(-5 * time.Hour).Format(time.RFC3339),
				},
			},
			duration:  2 * time.Hour,
			wantCount: 0,
			wantIDs:   []string{},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FilterRecent(tt.posts, tt.duration)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// Check count
			assert.Equal(t, tt.wantCount, len(result), "expected %d posts, got %d", tt.wantCount, len(result))

			// Check IDs match and are in expected order
			if len(result) > 0 {
				resultIDs := make([]string, len(result))
				for i, post := range result {
					resultIDs[i] = post.ID
				}
				assert.Equal(t, tt.wantIDs, resultIDs)
			}
		})
	}
}

func TestFilterRecentSortingOrder(t *testing.T) {
	now := time.Now().UTC()

	// Create posts with known timestamps
	posts := []*Post{
		{
			ID:        "smk-first",
			Author:    "ember",
			Project:   "smoke",
			Suffix:    "swift-fox",
			Content:   "first post",
			CreatedAt: now.Add(-3 * time.Minute).Format(time.RFC3339),
		},
		{
			ID:        "smk-second",
			Author:    "witness",
			Project:   "smoke",
			Suffix:    "swift-fox",
			Content:   "second post",
			CreatedAt: now.Add(-2 * time.Minute).Format(time.RFC3339),
		},
		{
			ID:        "smk-third",
			Author:    "ember",
			Project:   "smoke",
			Suffix:    "swift-fox",
			Content:   "third post (newest)",
			CreatedAt: now.Add(-1 * time.Minute).Format(time.RFC3339),
		},
	}

	result, err := FilterRecent(posts, 1*time.Hour)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(result))

	// Verify newest first order
	assert.Equal(t, "smk-third", result[0].ID)
	assert.Equal(t, "smk-second", result[1].ID)
	assert.Equal(t, "smk-first", result[2].ID)
}

func TestFilterRecentEdgeCases(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name      string
		posts     []*Post
		duration  time.Duration
		wantCount int
	}{
		{
			name:      "zero duration window",
			posts:     []*Post{},
			duration:  0 * time.Hour,
			wantCount: 0,
		},
		{
			name: "post exactly now",
			posts: []*Post{
				{
					ID:        "smk-now",
					Author:    "ember",
					Project:   "smoke",
					Suffix:    "swift-fox",
					Content:   "right now",
					CreatedAt: now.Format(time.RFC3339),
				},
			},
			duration:  1 * time.Hour,
			wantCount: 1,
		},
		{
			name: "very large time window (30 days)",
			posts: []*Post{
				{
					ID:        "smk-old-but-in",
					Author:    "ember",
					Project:   "smoke",
					Suffix:    "swift-fox",
					Content:   "old post",
					CreatedAt: now.Add(-29 * 24 * time.Hour).Format(time.RFC3339),
				},
				{
					ID:        "smk-outside-window",
					Author:    "witness",
					Project:   "smoke",
					Suffix:    "swift-fox",
					Content:   "outside window",
					CreatedAt: now.Add(-31 * 24 * time.Hour).Format(time.RFC3339),
				},
			},
			duration:  30 * 24 * time.Hour,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FilterRecent(tt.posts, tt.duration)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantCount, len(result))
		})
	}
}
