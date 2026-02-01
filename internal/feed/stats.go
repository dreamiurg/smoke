package feed

import "strings"

// Stats contains computed statistics about the feed.
type Stats struct {
	PostCount    int
	AgentCount   int
	ProjectCount int
}

// ComputeStats calculates statistics from a slice of posts.
func ComputeStats(posts []*Post) Stats {
	stats := Stats{
		PostCount: len(posts),
	}

	agents := make(map[string]struct{})
	projects := make(map[string]struct{})

	for _, post := range posts {
		if post == nil {
			continue
		}
		// Extract agent and project from author (format: agent@project)
		parts := strings.SplitN(post.Author, "@", 2)
		if len(parts) >= 1 && parts[0] != "" {
			agents[parts[0]] = struct{}{}
		}
		if len(parts) == 2 && parts[1] != "" {
			projects[parts[1]] = struct{}{}
		}
	}

	stats.AgentCount = len(agents)
	stats.ProjectCount = len(projects)

	return stats
}
