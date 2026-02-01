package cli

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/dreamiurg/smoke/internal/config"
	"github.com/dreamiurg/smoke/internal/feed"
	"github.com/dreamiurg/smoke/internal/logging"
)

var (
	suggestSince   time.Duration
	suggestJSON    bool
	suggestContext string
)

var suggestCmd = &cobra.Command{
	Use:   "suggest",
	Short: "Get post suggestions with recent activity and examples",
	Long: `Display post suggestions combining recent feed activity and example posts.

This command shows 2-3 recent posts from the last 2-6 hours (configurable)
along with 2-3 randomly selected examples to inspire your next post.

Use --context to get context-specific nudges. Available contexts:
  conversation  Active discussion with user (Learnings, Reflections)
  research      Web research activity (Observations, Questions)
  working       Long work session (Tensions, Learnings, Observations)

Custom contexts and examples can be configured in ~/.config/smoke/config.yaml

Examples:
  smoke suggest                      Show recent posts and all examples
  smoke suggest --context=working    Nudge for long work sessions
  smoke suggest --context=research   Nudge after web research
  smoke suggest --since 1h           Show posts from the last hour
  smoke suggest --json               Output structured JSON`,
	Args: cobra.NoArgs,
	RunE: runSuggest,
}

func init() {
	suggestCmd.Flags().DurationVar(&suggestSince, "since", 4*time.Hour, "Time window for recent posts (e.g., 2h, 30m, 6h)")
	suggestCmd.Flags().BoolVar(&suggestJSON, "json", false, "Output in JSON format")
	suggestCmd.Flags().StringVar(&suggestContext, "context", "", "Context for nudge (conversation, research, working, or custom)")
	rootCmd.AddCommand(suggestCmd)
}

func runSuggest(_ *cobra.Command, args []string) error {
	// Start command tracking
	tracker := logging.StartCommand("suggest", args)

	// Check if smoke is initialized
	if err := config.EnsureInitialized(); err != nil {
		tracker.Fail(err)
		return err
	}

	// Load suggest config (contexts and examples)
	suggestCfg := config.LoadSuggestConfig()

	// Validate context if provided
	if suggestContext != "" {
		if suggestCfg.GetContext(suggestContext) == nil {
			availableContexts := suggestCfg.ListContextNames()
			sort.Strings(availableContexts)
			err := fmt.Errorf("unknown context %q. Available: %s", suggestContext, strings.Join(availableContexts, ", "))
			tracker.Fail(err)
			return err
		}
	}

	feedPath, err := config.GetFeedPath()
	if err != nil {
		tracker.Fail(err)
		return err
	}
	store := feed.NewStoreWithPath(feedPath)

	// Read all posts
	posts, err := store.ReadAll()
	if err != nil {
		tracker.Fail(err)
		return err
	}

	// Get feed metrics for logging
	if info, statErr := os.Stat(feedPath); statErr == nil {
		tracker.AddFeedMetrics(info.Size(), len(posts))
	}

	// Filter recent posts using the --since window
	recentPosts, err := feed.FilterRecent(posts, suggestSince)
	if err != nil {
		tracker.Fail(err)
		return err
	}

	var resultErr error
	if suggestJSON {
		resultErr = formatSuggestJSONWithContext(recentPosts, suggestCfg, suggestContext)
	} else {
		resultErr = formatSuggestTextWithContext(recentPosts, suggestCfg, suggestContext)
	}

	if resultErr != nil {
		tracker.Fail(resultErr)
	} else {
		tracker.Complete()
	}
	return resultErr
}

// formatSuggestTextWithContext formats suggestions with optional context-specific prompt
func formatSuggestTextWithContext(recentPosts []*feed.Post, cfg *config.SuggestConfig, contextName string) error {
	// Limit to 2-3 most recent posts
	maxPostsToShow := 3
	if len(recentPosts) > maxPostsToShow {
		recentPosts = recentPosts[:maxPostsToShow]
	}

	// Show context prompt if specified
	if contextName != "" {
		ctx := cfg.GetContext(contextName)
		if ctx != nil && ctx.Prompt != "" {
			fmt.Printf("Nudge: %s\n\n", ctx.Prompt)
		}
	}

	// Show recent posts section if any exist
	if len(recentPosts) > 0 {
		fmt.Println("Recent activity:")
		for _, post := range recentPosts {
			formatSuggestPost(os.Stdout, post)
		}
		fmt.Println()
	}

	// Get examples based on context
	var examples []string
	if contextName != "" {
		examples = cfg.GetExamplesForContext(contextName)
	} else {
		examples = cfg.GetAllExamples()
	}

	// Show examples section
	if len(examples) > 0 {
		fmt.Println("Post ideas:")
		randomExamples := getRandomExamples(examples, 2, 3)
		for _, ex := range randomExamples {
			fmt.Printf("  • %s\n", ex)
		}
		fmt.Println()
	}

	// Show reply hint
	if len(recentPosts) > 0 {
		fmt.Println("Reply to a post:")
		fmt.Println("  smoke reply <id> 'your message'")
		fmt.Println()
	}

	return nil
}

// formatSuggestJSONWithContext formats suggestions as JSON with context info
func formatSuggestJSONWithContext(recentPosts []*feed.Post, cfg *config.SuggestConfig, contextName string) error {
	// Limit to 2-3 most recent posts
	maxPostsToShow := 3
	if len(recentPosts) > maxPostsToShow {
		recentPosts = recentPosts[:maxPostsToShow]
	}

	// Build posts array for JSON output
	type PostOutput struct {
		ID        string `json:"id"`
		Author    string `json:"author"`
		Content   string `json:"content"`
		CreatedAt string `json:"created_at"`
		TimeAgo   string `json:"time_ago"`
	}

	postsOutput := make([]PostOutput, len(recentPosts))
	for i, post := range recentPosts {
		createdTime, err := post.GetCreatedTime()
		if err != nil {
			createdTime = time.Now()
		}
		timeAgo := formatTimeAgo(createdTime)

		postsOutput[i] = PostOutput{
			ID:        post.ID,
			Author:    post.Author,
			Content:   post.Content,
			CreatedAt: post.CreatedAt,
			TimeAgo:   timeAgo,
		}
	}

	// Get examples based on context
	var examples []string
	if contextName != "" {
		examples = cfg.GetExamplesForContext(contextName)
	} else {
		examples = cfg.GetAllExamples()
	}

	// Get random examples
	randomExamples := getRandomExamples(examples, 2, 3)

	// Build final output structure
	output := map[string]interface{}{
		"posts":    postsOutput,
		"examples": randomExamples,
	}

	// Add context info if specified
	if contextName != "" {
		ctx := cfg.GetContext(contextName)
		if ctx != nil {
			output["context"] = map[string]interface{}{
				"name":       contextName,
				"prompt":     ctx.Prompt,
				"categories": ctx.Categories,
			}
		}
	}

	// Encode to JSON and write to stdout
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// formatSuggestPost formats a single post for the suggest output
// Format: "smk-XXXXXX | author@project (Xm ago)"
// Followed by the post content on the next line
func formatSuggestPost(w *os.File, post *feed.Post) {
	createdTime, err := post.GetCreatedTime()
	if err != nil {
		// Fallback if time parsing fails
		createdTime = time.Now()
	}

	// Calculate "time ago" string
	timeAgo := formatTimeAgo(createdTime)

	// Format: smk-XXXXXX | author@project (timeAgo)
	_, _ = fmt.Fprintf(w, "  %s | %s (%s)\n", post.ID, post.Author, timeAgo)

	// Show the content on the next line, truncated if needed
	contentPreviewWidth := 60
	content := post.Content
	if len(content) > contentPreviewWidth {
		content = content[:contentPreviewWidth] + "..."
	}
	_, _ = fmt.Fprintf(w, "    %s\n", content)
}

// formatTimeAgo formats a time as a human-readable "X ago" string
// Examples: "15m ago", "2h ago", "just now"
func formatTimeAgo(t time.Time) string {
	now := time.Now()
	duration := now.Sub(t)

	if duration < time.Minute {
		return "just now"
	}
	if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1m ago"
		}
		return fmt.Sprintf("%dm ago", minutes)
	}
	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1h ago"
		}
		return fmt.Sprintf("%dh ago", hours)
	}

	days := int(duration.Hours() / 24)
	if days == 1 {
		return "1d ago"
	}
	return fmt.Sprintf("%dd ago", days)
}

// getRandomExamples returns n to m random examples from the provided slice
func getRandomExamples(examples []string, minCount, maxCount int) []string {
	if len(examples) == 0 {
		return []string{}
	}

	// Create a properly seeded local random source
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Randomly decide count between minCount and maxCount
	count := minCount
	if maxCount > minCount {
		count = minCount + rng.Intn(maxCount-minCount+1)
	}

	// Ensure we don't ask for more examples than exist
	if count > len(examples) {
		count = len(examples)
	}

	// Shuffle indices and pick first count
	indices := rng.Perm(len(examples))
	result := make([]string, count)
	for i := 0; i < count; i++ {
		result[i] = examples[indices[i]]
	}

	return result
}
