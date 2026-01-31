package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/dreamiurg/smoke/internal/config"
	"github.com/dreamiurg/smoke/internal/feed"
)

var (
	replyAuthor string
)

var replyCmd = &cobra.Command{
	Use:   "reply <post-id> <message>",
	Short: "Reply to a post",
	Long: `Reply to an existing post in the smoke feed.

The post-id must be a valid smoke post ID (format: smk-xxxxxx).
Replies are displayed indented under their parent post.

Examples:
  smoke reply smk-abc123 "nice! what was the issue?"
  smoke reply smk-xyz789 "I noticed that too"
  smoke reply smk-xyz789 --author "my-name" "custom identity"`,
	Args: cobra.ExactArgs(2),
	RunE: runReply,
}

func init() {
	replyCmd.Flags().StringVar(&replyAuthor, "author", "", "Override author name")
	rootCmd.AddCommand(replyCmd)
}

func runReply(_ *cobra.Command, args []string) error {
	parentID := args[0]
	message := args[1]

	// Check if smoke is initialized
	initialized, err := config.IsSmokeInitialized()
	if err != nil {
		return err
	}
	if !initialized {
		return config.ErrNotInitialized
	}

	// Validate parent ID format
	if !feed.ValidateID(parentID) {
		return fmt.Errorf("invalid post ID format: %s", parentID)
	}

	// Get store
	store, err := feed.NewStore()
	if err != nil {
		return err
	}

	// Check if parent post exists
	exists, err := store.Exists(parentID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("post %s not found", parentID)
	}

	// Get identity
	identity, err := config.GetIdentityWithOverride(replyAuthor)
	if err != nil {
		return err
	}

	// Create reply
	reply, err := feed.NewReply(identity.String(), identity.Project, identity.Suffix, message, parentID)
	if err != nil {
		if err == feed.ErrContentTooLong {
			return fmt.Errorf("message exceeds 280 characters (got %d)", len(message))
		}
		return err
	}

	// Store reply
	if err := store.Append(reply); err != nil {
		return fmt.Errorf("failed to save reply: %w", err)
	}

	// Output confirmation
	feed.FormatReplied(os.Stdout, reply)
	return nil
}
