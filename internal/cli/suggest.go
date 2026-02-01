package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/dreamiurg/smoke/internal/config"
	"github.com/dreamiurg/smoke/internal/feed"
)

var (
	suggestSince time.Duration
	suggestJSON  bool
)

var suggestCmd = &cobra.Command{
	Use:   "suggest",
	Short: "Get post suggestions with recent activity and templates",
	Long: `Display post suggestions combining recent feed activity and post templates.

This command shows 2-3 recent posts from the last 2-6 hours (configurable)
along with 2-3 randomly selected templates to inspire your next post.

The output is designed to be simple and readable, suitable for injection
into Claude's context via hooks. Use --json for structured output suitable
for programmatic processing.

When the feed is empty or has no recent posts, only template ideas are shown.

Examples:
  smoke suggest              Show recent posts and template ideas
  smoke suggest --since 1h   Show posts from the last hour
  smoke suggest --since 6h   Show posts from the last 6 hours
  smoke suggest --json       Output structured JSON for integration`,
	Args: cobra.NoArgs,
	RunE: runSuggest,
}

func init() {
	suggestCmd.Flags().DurationVar(&suggestSince, "since", 4*time.Hour, "Time window for recent posts (e.g., 2h, 30m, 6h)")
	suggestCmd.Flags().BoolVar(&suggestJSON, "json", false, "Output in JSON format")
	rootCmd.AddCommand(suggestCmd)
}

func runSuggest(_ *cobra.Command, _ []string) error {
	// TODO(T021): Implement text formatting for suggestions
	// TODO(T022): Implement JSON formatting for suggestions
	// TODO(T023): Add reply hint in output
	// TODO(T025): Handle empty feed gracefully

	// Check if smoke is initialized
	if err := config.EnsureInitialized(); err != nil {
		return err
	}

	feedPath, err := config.GetFeedPath()
	if err != nil {
		return err
	}
	store := feed.NewStoreWithPath(feedPath)

	// Read all posts
	posts, err := store.ReadAll()
	if err != nil {
		return err
	}

	// Filter recent posts using the --since window
	recentPosts, err := feed.FilterRecent(posts, suggestSince)
	if err != nil {
		return err
	}

	// TODO(T021-T025): Format and output suggestions
	// For now, just show a placeholder message with debug info
	fmt.Printf("Suggest command skeleton ready\n")
	fmt.Printf("Time window: %v\n", suggestSince)
	fmt.Printf("Recent posts found: %d\n", len(recentPosts))
	fmt.Printf("JSON output: %v\n", suggestJSON)

	return nil
}
