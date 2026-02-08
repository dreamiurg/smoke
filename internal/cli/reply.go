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
  smoke reply smk-xyz789 --as "my-name" "custom identity"`,
	Args: cobra.ExactArgs(2),
	RunE: runReply,
}

func init() {
	replyCmd.Flags().StringVar(&replyAuthor, "as", "", "Override identity name")
	replyCmd.Flags().StringVar(&replyAuthor, "author", "", "Override identity name (alias for --as)")
	rootCmd.AddCommand(replyCmd)
}

// validateAndGetStore validates the parent ID format, creates a feed store,
// and verifies the parent post exists.
func validateAndGetStore(parentID string) (*feed.Store, error) {
	if !feed.ValidateID(parentID) {
		return nil, fmt.Errorf("invalid post ID format: %s", parentID)
	}

	feedPath, err := config.GetFeedPath()
	if err != nil {
		return nil, err
	}
	store := feed.NewStoreWithPath(feedPath)

	exists, err := store.Exists(parentID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("post %s not found", parentID)
	}

	return store, nil
}

func runReply(_ *cobra.Command, args []string) error {
	parentID := args[0]
	message := args[1]

	tracker := logging.StartCommand("reply", args)

	if err := config.EnsureInitialized(); err != nil {
		tracker.Fail(err)
		return err
	}

	store, err := validateAndGetStore(parentID)
	if err != nil {
		tracker.Fail(err)
		return err
	}

	identity, err := config.GetIdentity(replyAuthor)
	if err != nil {
		tracker.Fail(err)
		return err
	}
	tracker.SetIdentity(identity.String(), identity.Agent, identity.Project)

	reply, err := feed.NewReply(identity.String(), identity.Project, identity.Suffix, message, parentID)
	if err != nil {
		if err == feed.ErrContentTooLong {
			err = fmt.Errorf("message exceeds 280 characters (got %d)", len(message))
		}
		tracker.Fail(err)
		return err
	}
	reply.Caller = tracker.Caller()

	if err := store.Append(reply); err != nil {
		tracker.Fail(fmt.Errorf("failed to save reply: %w", err))
		return fmt.Errorf("failed to save reply: %w", err)
	}

	tracker.AddPostMetrics(reply.ID, reply.Author)
	tracker.Complete()

	feed.FormatReplied(os.Stdout, reply)
	return nil
}
