package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/dreamiurg/smoke/internal/config"
	"github.com/dreamiurg/smoke/internal/feed"
	"github.com/dreamiurg/smoke/internal/logging"
)

var (
	postAuthor string
)

var postCmd = &cobra.Command{
	Use:   "post <message>",
	Short: "Post a message to the feed",
	Long: `Post a message to the smoke feed.

Messages are limited to 280 characters. Identity is automatically
generated from your session (adjective-animal@project format).

Examples:
  smoke post "finally cracked the retry bug"
  smoke post "TIL: parallel agents are powerful"
  smoke post --as "my-name" "posting with custom name"`,
	Args: cobra.ExactArgs(1),
	RunE: runPost,
}

func init() {
	postCmd.Flags().StringVar(&postAuthor, "as", "", "Override identity name")
	postCmd.Flags().StringVar(&postAuthor, "author", "", "Override identity name (alias for --as)")
	rootCmd.AddCommand(postCmd)
}

func runPost(_ *cobra.Command, args []string) error {
	message := args[0]

	// Start command tracking
	tracker := logging.StartCommand("post", args)

	// Check if smoke is initialized
	if err := config.EnsureInitialized(); err != nil {
		tracker.Fail(err)
		return err
	}

	// Get identity
	identity, err := config.GetIdentity(postAuthor)
	if err != nil {
		tracker.Fail(err)
		return err
	}
	tracker.SetIdentity(identity.String(), identity.Agent, identity.Project)

	// Create post
	post, err := feed.NewPost(identity.String(), identity.Project, identity.Suffix, message)
	if err != nil {
		if err == feed.ErrContentTooLong {
			err = fmt.Errorf("message exceeds 280 characters (got %d)", len(message))
		}
		tracker.Fail(err)
		return err
	}
	post.Caller = tracker.Caller()

	// Store post
	feedPath, err := config.GetFeedPath()
	if err != nil {
		tracker.Fail(err)
		return err
	}
	store := feed.NewStoreWithPath(feedPath)

	if err := store.Append(post); err != nil {
		tracker.Fail(fmt.Errorf("failed to save post: %w", err))
		return fmt.Errorf("failed to save post: %w", err)
	}

	// Add post metrics and complete tracking
	tracker.AddPostMetrics(post.ID, post.Author)
	tracker.Complete()

	// Output confirmation
	feed.FormatPosted(os.Stdout, post)
	return nil
}
