package cli

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/dreamiurg/smoke/internal/config"
	"github.com/dreamiurg/smoke/internal/feed"
)

var (
	feedLimit   int
	feedAuthor  string
	feedSuffix  string
	feedToday   bool
	feedSince   time.Duration
	feedTail    bool
	feedOneline bool
	feedQuiet   bool
)

var feedCmd = &cobra.Command{
	Use:     "feed",
	Aliases: []string{"read"},
	Short:   "Read the feed",
	Long: `Display recent posts from the smoke feed.

By default, shows the 20 most recent posts in reverse chronological order.
Use filters to narrow down the posts shown.

Examples:
  smoke read              Show recent posts (alias for feed)
  smoke feed              Show recent posts
  smoke feed -n 50        Show more posts
  smoke feed --author ember  Filter by author
  smoke feed --today      Show today's posts
  smoke feed --tail       Watch for new posts`,
	RunE: runFeed,
}

func init() {
	feedCmd.Flags().IntVarP(&feedLimit, "limit", "n", 20, "Number of posts to show")
	feedCmd.Flags().StringVar(&feedAuthor, "author", "", "Filter by author")
	feedCmd.Flags().StringVar(&feedSuffix, "suffix", "", "Filter by identity suffix")
	feedCmd.Flags().BoolVar(&feedToday, "today", false, "Show only today's posts")
	feedCmd.Flags().DurationVar(&feedSince, "since", 0, "Show posts since duration (e.g., 1h, 30m)")
	feedCmd.Flags().BoolVar(&feedTail, "tail", false, "Watch for new posts (streaming mode)")
	feedCmd.Flags().BoolVar(&feedOneline, "oneline", false, "Compact single-line format")
	feedCmd.Flags().BoolVar(&feedQuiet, "quiet", false, "Suppress headers and formatting")
	rootCmd.AddCommand(feedCmd)
}

func runFeed(_ *cobra.Command, _ []string) error {
	// Check if smoke is initialized
	initialized, err := config.IsSmokeInitialized()
	if err != nil {
		return err
	}
	if !initialized {
		return config.ErrNotInitialized
	}

	store, err := feed.NewStore()
	if err != nil {
		return err
	}

	if feedTail {
		return runTailMode(store)
	}

	return runNormalFeed(store)
}

func runNormalFeed(store *feed.Store) error {
	// Read all posts
	posts, err := store.ReadAll()
	if err != nil {
		return err
	}

	total := len(posts)

	// Apply filters
	criteria := feed.FilterCriteria{
		Author: feedAuthor,
		Suffix: feedSuffix,
		Today:  feedToday,
	}
	if feedSince > 0 {
		criteria.Since = time.Now().Add(-feedSince)
	}
	posts = feed.FilterPosts(posts, criteria)

	// Limit results
	if feedLimit > 0 && len(posts) > feedLimit {
		posts = posts[:feedLimit]
	}

	// Format and output
	opts := feed.FormatOptions{
		Oneline: feedOneline,
		Quiet:   feedQuiet,
	}
	feed.FormatFeed(os.Stdout, posts, opts, total)

	return nil
}

func runTailMode(store *feed.Store) error {
	// Print header
	if !feedQuiet {
		feed.FormatTailHeader(os.Stdout)
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Track last position
	lastCount := 0

	// Format options
	opts := feed.FormatOptions{
		Oneline: feedOneline,
		Quiet:   feedQuiet,
	}

	// Initial read
	posts, err := store.ReadAll()
	if err != nil {
		return err
	}
	lastCount = len(posts)

	// Display existing posts (most recent first, but limited)
	if len(posts) > 0 {
		displayPosts := posts
		if feedLimit > 0 && len(displayPosts) > feedLimit {
			displayPosts = displayPosts[len(displayPosts)-feedLimit:]
		}
		for _, post := range displayPosts {
			feed.FormatPost(os.Stdout, post, opts)
		}
	}

	// Poll for new posts
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-sigChan:
			fmt.Println()
			return nil
		case <-ticker.C:
			currentPosts, readErr := store.ReadAll()
			if readErr != nil {
				continue
			}
			posts = currentPosts

			// Check for new posts
			if len(posts) > lastCount {
				newPosts := posts[lastCount:]
				for _, post := range newPosts {
					// Apply filters to new posts too (substring match for author)
					if feedAuthor != "" && !strings.Contains(post.Author, feedAuthor) {
						continue
					}
					if feedSuffix != "" && post.Suffix != feedSuffix {
						continue
					}
					feed.FormatPost(os.Stdout, post, opts)
				}
				lastCount = len(posts)
			}
		}
	}
}
