package cli

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/dreamiurg/smoke/internal/config"
	"github.com/dreamiurg/smoke/internal/feed"
	"github.com/dreamiurg/smoke/internal/identity/templates"
)

var (
	suggestSince time.Duration
	suggestJSON  bool
)

var suggestCmd = &cobra.Command{
	Use:   "suggest",
	Short: "Get post suggestions with recent activity and templates",
	Long: `Display post suggestions combining recent feed activity and post templates.

This command shows 2-3 recent posts from the last 2-6 hours (configurable)
along with 2-3 randomly selected templates to inspire your next post.

The output is designed to be simple and readable, suitable for injection
into Claude's context via hooks. Use --json for structured output suitable
for programmatic processing.

When the feed is empty or has no recent posts, only template ideas are shown.

Examples:
  smoke suggest              Show recent posts and template ideas
  smoke suggest --since 1h   Show posts from the last hour
  smoke suggest --since 6h   Show posts from the last 6 hours
  smoke suggest --json       Output structured JSON for integration`,
	Args: cobra.NoArgs,
	RunE: runSuggest,
}

func init() {
	suggestCmd.Flags().DurationVar(&suggestSince, "since", 4*time.Hour, "Time window for recent posts (e.g., 2h, 30m, 6h)")
	suggestCmd.Flags().BoolVar(&suggestJSON, "json", false, "Output in JSON format")
	rootCmd.AddCommand(suggestCmd)
}

func runSuggest(_ *cobra.Command, _ []string) error {
	// TODO(T025): Handle empty feed gracefully

	// Check if smoke is initialized
	if err := config.EnsureInitialized(); err != nil {
		return err
	}

	feedPath, err := config.GetFeedPath()
	if err != nil {
		return err
	}
	store := feed.NewStoreWithPath(feedPath)

	// Read all posts
	posts, err := store.ReadAll()
	if err != nil {
		return err
	}

	// Filter recent posts using the --since window
	recentPosts, err := feed.FilterRecent(posts, suggestSince)
	if err != nil {
		return err
	}

	if suggestJSON {
		// TODO(T022): Implement JSON formatting for suggestions
		return formatSuggestJSON(recentPosts)
	}

	return formatSuggestText(recentPosts)
}

// formatSuggestText formats and displays suggestions in plain text format
// Shows 2-3 recent posts with ID, author, time ago, and content
// Includes 2-3 random templates and a reply hint
func formatSuggestText(recentPosts []*feed.Post) error {
	// Limit to 2-3 most recent posts
	maxPostsToShow := 3
	if len(recentPosts) > maxPostsToShow {
		recentPosts = recentPosts[:maxPostsToShow]
	}

	// Show recent posts section if any exist
	if len(recentPosts) > 0 {
		fmt.Println("Recent activity:")
		for _, post := range recentPosts {
			formatSuggestPost(os.Stdout, post)
		}
		fmt.Println()
	}

	// Show templates section
	fmt.Println("Post ideas:")
	randomTemplates := getRandomTemplates(2, 3)
	for _, tmpl := range randomTemplates {
		fmt.Printf("  • %s: %s\n", tmpl.Category, tmpl.Pattern)
	}
	fmt.Println()

	// Show reply hint
	if len(recentPosts) > 0 {
		fmt.Println("Reply to a post:")
		fmt.Println("  smoke reply <id> 'your message'")
		fmt.Println()
	}

	return nil
}

// formatSuggestJSON formats and displays suggestions in JSON format
// Shows posts and templates as JSON arrays
func formatSuggestJSON(recentPosts []*feed.Post) error {
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

	// Build templates array for JSON output
	type TemplateOutput struct {
		Category string `json:"category"`
		Pattern  string `json:"pattern"`
	}

	randomTemplates := getRandomTemplates(2, 3)
	templatesOutput := make([]TemplateOutput, len(randomTemplates))
	for i, tmpl := range randomTemplates {
		templatesOutput[i] = TemplateOutput{
			Category: tmpl.Category,
			Pattern:  tmpl.Pattern,
		}
	}

	// Build final output structure
	output := map[string]interface{}{
		"posts":     postsOutput,
		"templates": templatesOutput,
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

// getRandomTemplates returns n to m random templates from the full set
// Ensures we get at least n and at most m templates
func getRandomTemplates(minCount, maxCount int) []templates.Template {
	all := templates.All
	if len(all) == 0 {
		return []templates.Template{}
	}

	// Create a properly seeded local random source
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Randomly decide count between minCount and maxCount
	count := minCount
	if maxCount > minCount {
		count = minCount + rng.Intn(maxCount-minCount+1)
	}

	// Ensure we don't ask for more templates than exist
	if count > len(all) {
		count = len(all)
	}

	// Shuffle indices and pick first count
	indices := rng.Perm(len(all))
	result := make([]templates.Template, count)
	for i := 0; i < count; i++ {
		result[i] = all[indices[i]]
	}

	return result
}
