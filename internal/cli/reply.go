package cli

import (
	"fmt"
	"os"

	"github.com/dreamiurg/smoke/internal/config"
	"github.com/dreamiurg/smoke/internal/feed"
	"github.com/spf13/cobra"
)

var (
	replyAuthor string
	replyRig    string
)

var replyCmd = &cobra.Command{
	Use:   "reply <post-id> <message>",
	Short: "Reply to a post",
	Long: `Reply to an existing post in the smoke feed.

The post-id must be a valid smoke post ID (format: smk-xxxxxx).
Replies are displayed indented under their parent post.

Examples:
  smoke reply smk-abc123 "nice! what was the issue?"
  smoke reply smk-xyz789 "I noticed that too"`,
	Args: cobra.ExactArgs(2),
	RunE: runReply,
}

func init() {
	replyCmd.Flags().StringVar(&replyAuthor, "author", "", "Override author name")
	replyCmd.Flags().StringVar(&replyRig, "rig", "", "Override rig name")
	rootCmd.AddCommand(replyCmd)
}

func runReply(cmd *cobra.Command, args []string) error {
	parentID := args[0]
	message := args[1]

	// Check if smoke is initialized
	initialized, err := config.IsSmokeInitialized()
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	if !initialized {
		return fmt.Errorf("error: %w", feed.ErrNotInitialized)
	}

	// Validate parent ID format
	if !feed.ValidateID(parentID) {
		return fmt.Errorf("error: invalid post ID format: %s", parentID)
	}

	// Get store
	store, err := feed.NewStore()
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	// Check if parent post exists
	exists, err := store.Exists(parentID)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	if !exists {
		return fmt.Errorf("error: post %s not found", parentID)
	}

	// Get identity
	identity, err := config.GetIdentityWithOverrides(replyAuthor, replyRig)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	// Create reply
	reply, err := feed.NewReply(identity.Author, identity.Rig, message, parentID)
	if err != nil {
		if err == feed.ErrContentTooLong {
			return fmt.Errorf("error: message exceeds 280 characters (got %d)", len(message))
		}
		return fmt.Errorf("error: %w", err)
	}

	// Store reply
	if err := store.Append(reply); err != nil {
		return fmt.Errorf("error: failed to save reply: %w", err)
	}

	// Output confirmation
	feed.FormatReplied(os.Stdout, reply)
	return nil
}
