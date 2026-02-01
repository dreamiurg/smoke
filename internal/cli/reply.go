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

	logging.LogCommand("reply", args)

	// Check if smoke is initialized
	if err := config.EnsureInitialized(); err != nil {
		logging.LogError("smoke not initialized", err)
		return err
	}

	// Validate parent ID format
	if !feed.ValidateID(parentID) {
		return fmt.Errorf("invalid post ID format: %s", parentID)
	}

	// Get store
	feedPath, err := config.GetFeedPath()
	if err != nil {
		return err
	}
	store := feed.NewStoreWithPath(feedPath)

	// Check if parent post exists
	exists, err := store.Exists(parentID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("post %s not found", parentID)
	}

	// Get identity
	identity, err := config.GetIdentity(replyAuthor)
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
		logging.LogError("failed to save reply", err)
		return fmt.Errorf("failed to save reply: %w", err)
	}

	logging.LogPostCreated(reply.ID, reply.Author)

	// Output confirmation
	feed.FormatReplied(os.Stdout, reply)
	return nil
}
