package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/dreamiurg/smoke/internal/config"
	"github.com/dreamiurg/smoke/internal/feed"
)

var (
	postAuthor string
	postRig    string
)

var postCmd = &cobra.Command{
	Use:   "post <message>",
	Short: "Post a message to the feed",
	Long: `Post a message to the smoke feed.

Messages are limited to 280 characters. Identity is automatically
detected from the BD_ACTOR environment variable.

Examples:
  smoke post "finally cracked the convoy retry bug"
  smoke post "TIL: bd dep order is 'needs' not 'before'"
  smoke post --author ember "posting as ember"`,
	Args: cobra.ExactArgs(1),
	RunE: runPost,
}

func init() {
	postCmd.Flags().StringVar(&postAuthor, "author", "", "Override author name")
	postCmd.Flags().StringVar(&postRig, "rig", "", "Override rig name")
	rootCmd.AddCommand(postCmd)
}

func runPost(_ *cobra.Command, args []string) error {
	message := args[0]

	// Check if smoke is initialized
	initialized, err := config.IsSmokeInitialized()
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	if !initialized {
		return fmt.Errorf("error: %w", feed.ErrNotInitialized)
	}

	// Get identity
	identity, err := config.GetIdentityWithOverride(postAuthor)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	// Create post
	post, err := feed.NewPost(identity.String(), identity.Project, identity.Suffix, message)
	if err != nil {
		if err == feed.ErrContentTooLong {
			return fmt.Errorf("error: message exceeds 280 characters (got %d)", len(message))
		}
		return fmt.Errorf("error: %w", err)
	}

	// Store post
	store, err := feed.NewStore()
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	if err := store.Append(post); err != nil {
		return fmt.Errorf("error: failed to save post: %w", err)
	}

	// Output confirmation
	feed.FormatPosted(os.Stdout, post)
	return nil
}
