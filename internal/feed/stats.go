package feed

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
		agent, project := SplitIdentity(post.Author)
		if agent != "" {
			agents[agent] = struct{}{}
		}
		if project != "" {
			projects[project] = struct{}{}
		}
	}

	stats.AgentCount = len(agents)
	stats.ProjectCount = len(projects)

	return stats
}
