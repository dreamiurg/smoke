package feed

import (
	"sort"
	"time"
)

// FilterRecent filters posts to those within the specified time window.
// It returns posts created within the last 'window' duration from now,
// sorted by timestamp newest first. Future posts are excluded.
// If the feed is empty, returns an empty slice with no error.
func FilterRecent(posts []*Post, window time.Duration) ([]*Post, error) {
	if len(posts) == 0 {
		return []*Post{}, nil
	}

	now := time.Now().UTC()
	cutoff := now.Add(-window)

	var filtered []*Post

	for _, post := range posts {
		// Parse the CreatedAt timestamp
		createdTime, err := post.GetCreatedTime()
		if err != nil {
			// Skip posts with invalid timestamps
			continue
		}

		// Exclude future posts (createdTime > now)
		if createdTime.After(now) {
			continue
		}

		// Include posts within the time window (createdTime >= cutoff)
		// Allow a 1-second grace period for boundary conditions to handle timing differences
		gracePeriod := 1 * time.Second
		if !createdTime.Before(cutoff.Add(-gracePeriod)) {
			filtered = append(filtered, post)
		}
	}

	// Sort by timestamp newest first
	sort.Slice(filtered, func(i, j int) bool {
		timeI, _ := filtered[i].GetCreatedTime()
		timeJ, _ := filtered[j].GetCreatedTime()
		return timeI.After(timeJ)
	})

	return filtered, nil
}

// GetRecentPosts is a convenience function that filters posts to recent ones
// within a specified duration and returns them.
func GetRecentPosts(posts []*Post, duration time.Duration) ([]*Post, error) {
	return FilterRecent(posts, duration)
}
