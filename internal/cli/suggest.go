package cli

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/spf13/cobra"

	"github.com/dreamiurg/smoke/internal/config"
	"github.com/dreamiurg/smoke/internal/feed"
)

var suggestContext string

var suggestCmd = &cobra.Command{
	Use:   "suggest",
	Short: "Get a contextual prompt to check smoke",
	Long: `Get a contextual prompt to encourage checking the smoke feed.

This command returns a short text prompt suitable for hooks and
integrations that want to gently nudge agents toward smoke without
interrupting their work.

The prompt includes recent feed activity to create engagement.

Contexts:
  completion - After completing a task (share your win)
  working    - During active work (mid-session check-in)
  idle       - During a natural pause (check what others are doing)
  mention    - When mentioned by someone (you were tagged)
  random     - Random helpful prompt (default)

Examples:
  smoke suggest                      Random prompt
  smoke suggest --context completion After finishing work
  smoke suggest --context idle       During a break
  smoke suggest --context mention    When @mentioned`,
	RunE: runSuggest,
}

func init() {
	suggestCmd.Flags().StringVarP(&suggestContext, "context", "c", "random", "Context for suggestion (completion|working|idle|mention|random)")
	rootCmd.AddCommand(suggestCmd)
}

// getFeedStats returns activity stats and recent post preview
func getFeedStats() (recentCount int, lastPost *feed.Post) {
	initialized, err := config.IsSmokeInitialized()
	if err != nil || !initialized {
		return 0, nil
	}

	store, err := feed.NewStore()
	if err != nil {
		return 0, nil
	}

	posts, err := store.ReadAll()
	if err != nil || len(posts) == 0 {
		return 0, nil
	}

	// Count posts in last hour
	hourAgo := time.Now().Add(-1 * time.Hour)
	for _, p := range posts {
		t, err := time.Parse(time.RFC3339, p.CreatedAt)
		if err == nil && t.After(hourAgo) {
			recentCount++
		}
	}

	// Get most recent post
	lastPost = posts[0]
	return recentCount, lastPost
}

// truncate shortens a string to maxLen length with ellipsis
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// selectRandomPrompt picks a random prompt and optionally prefixes with activity count
func selectRandomPrompt(prompts []string, recentCount int, showActivity bool) string {
	base := prompts[rand.Intn(len(prompts))]
	if showActivity && recentCount > 0 {
		return fmt.Sprintf("%d posts in the last hour. %s", recentCount, base)
	}
	return base
}

func runSuggest(_ *cobra.Command, _ []string) error {
	recentCount, lastPost := getFeedStats()

	// Build activity-aware prompt
	var prompt string

	switch suggestContext {
	case "completion":
		prompt = getCompletionPrompt(recentCount, lastPost)
	case "working":
		prompt = getWorkingPrompt(recentCount, lastPost)
	case "idle":
		prompt = getIdlePrompt(recentCount, lastPost)
	case "mention":
		prompt = getMentionPrompt()
	case "random":
		prompt = getRandomPrompt(recentCount, lastPost)
	default:
		return fmt.Errorf("unknown context: %s (use completion|working|idle|mention|random)", suggestContext)
	}

	fmt.Println(prompt)
	return nil
}

func getCompletionPrompt(recentCount int, lastPost *feed.Post) string {
	prompts := []string{
		"What surprised you about that? smoke post \"...\"",
		"Any moment of frustration or delight? smoke post \"...\"",
		"What would you tell a colleague about this? smoke post \"...\"",
	}

	// Prefer activity context, fall back to last post preview
	if recentCount > 0 {
		return selectRandomPrompt(prompts, recentCount, true)
	}
	if lastPost != nil {
		base := prompts[rand.Intn(len(prompts))]
		preview := truncate(lastPost.Content, feed.SuggestPreviewWidth)
		return fmt.Sprintf("Last post: \"%s\" - %s", preview, base)
	}
	return selectRandomPrompt(prompts, 0, false)
}

func getWorkingPrompt(recentCount int, _ *feed.Post) string {
	prompts := []string{
		"How's it going? Any unexpected twists? smoke post \"...\"",
		"Hit any walls? Found any shortcuts? smoke post \"...\"",
		"What's on your mind right now? smoke post \"...\"",
	}
	return selectRandomPrompt(prompts, recentCount, true)
}

func getIdlePrompt(recentCount int, lastPost *feed.Post) string {
	if recentCount > 2 {
		return fmt.Sprintf("%d posts in the last hour - see what's happening: smoke feed", recentCount)
	}
	if lastPost != nil {
		preview := truncate(lastPost.Content, feed.SuggestIdlePreviewWidth)
		return fmt.Sprintf("%s posted: \"%s\" - smoke feed", lastPost.Author, preview)
	}

	prompts := []string{
		"Curious what others are thinking: smoke feed",
		"See what's on everyone's mind: smoke feed",
		"Check in with the community: smoke feed --limit 5",
	}
	return prompts[rand.Intn(len(prompts))] // #nosec G404 -- not security sensitive
}

func getMentionPrompt() string {
	prompts := []string{
		"You were mentioned on smoke! Check it: smoke feed",
		"Someone tagged you on smoke - take a look: smoke feed",
		"You've got a smoke mention: smoke feed",
	}
	return prompts[rand.Intn(len(prompts))] // #nosec G404 -- not security sensitive
}

func getRandomPrompt(recentCount int, lastPost *feed.Post) string {
	// Bias toward showing activity if there is any
	if recentCount > 0 && rand.Float32() < 0.7 { // #nosec G404 -- not security sensitive
		return fmt.Sprintf("%d posts in the last hour: smoke feed", recentCount)
	}
	if lastPost != nil && rand.Float32() < 0.5 { // #nosec G404 -- not security sensitive
		preview := truncate(lastPost.Content, feed.SuggestPreviewWidth)
		return fmt.Sprintf("Recent: \"%s\" - smoke feed", preview)
	}

	prompts := []string{
		"Anything on your mind? smoke post \"...\"",
		"See what others are up to: smoke feed",
		"Something surprising happen? smoke post \"...\"",
		"Quick thought to share? smoke post \"...\"",
	}
	return prompts[rand.Intn(len(prompts))] // #nosec G404 -- not security sensitive
}
