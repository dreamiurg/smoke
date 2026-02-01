package cli

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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
	if err := config.EnsureInitialized(); err != nil {
		return err
	}

	store, err := feed.NewStore()
	if err != nil {
		return err
	}

	if feedTail {
		return runTailMode(store)
	}

	// Launch interactive TUI if stdout is a TTY (terminal), otherwise use plain text output.
	// This provides a better user experience with navigation and formatting when the
	// output is not being piped or redirected.
	if feed.IsTerminal(os.Stdout.Fd()) {
		return runTUIMode(store)
	}

	return runNormalFeed(store)
}

func runNormalFeed(store *feed.Store) error {
	// Read posts sorted by time (most recent first)
	posts, err := store.ReadRecent(0) // 0 = no limit, just sorted
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

	// Limit results (already sorted, so take first N)
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

// runTUIMode launches the interactive TUI feed
func runTUIMode(store *feed.Store) error {
	// Load TUI config (never returns error, gracefully handles all failures)
	cfg := config.LoadTUIConfig()

	theme := feed.GetTheme(cfg.Theme)
	contrast := feed.GetContrastLevel(cfg.Contrast)
	layout := feed.GetLayout(cfg.Layout)

	// Get short version (e.g., "1.3.0" not the full string with commit info)
	version := Version

	// Create model and run
	m := feed.NewModel(store, theme, contrast, layout, cfg, version)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
