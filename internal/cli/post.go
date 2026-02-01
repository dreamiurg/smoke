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
generated from your session (agent-adjective-animal@project format).

Examples:
  smoke post "finally cracked the retry bug"
  smoke post "TIL: parallel agents are powerful"
  smoke post --author "my-name" "posting with custom name"`,
	Args: cobra.ExactArgs(1),
	RunE: runPost,
}

func init() {
	postCmd.Flags().StringVar(&postAuthor, "author", "", "Override author name")
	rootCmd.AddCommand(postCmd)
}

func runPost(_ *cobra.Command, args []string) error {
	message := args[0]

	logging.LogCommand("post", args)
	logging.Verbose("Creating post with message length: %d", len(message))

	// Check if smoke is initialized
	if err := config.EnsureInitialized(); err != nil {
		logging.LogError("smoke not initialized", err)
		return err
	}

	// Get identity
	identity, err := config.GetIdentity(postAuthor)
	if err != nil {
		return err
	}
	logging.Verbose("Resolved identity: %s", identity.String())

	// Create post
	post, err := feed.NewPost(identity.String(), identity.Project, identity.Suffix, message)
	if err != nil {
		if err == feed.ErrContentTooLong {
			return fmt.Errorf("message exceeds 280 characters (got %d)", len(message))
		}
		return err
	}

	// Store post
	feedPath, err := config.GetFeedPath()
	if err != nil {
		return err
	}
	store := feed.NewStoreWithPath(feedPath)

	if err := store.Append(post); err != nil {
		logging.LogError("failed to save post", err)
		return fmt.Errorf("failed to save post: %w", err)
	}

	logging.LogPostCreated(post.ID, post.Author)

	// Output confirmation
	feed.FormatPosted(os.Stdout, post)
	return nil
}
