package cli

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand/v2" // nosemgrep: go.lang.security.audit.crypto.math_random.math-random-used
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
	suggestSince    time.Duration
	suggestJSON     bool
	suggestContext  string
	suggestPressure int
)

var suggestCmd = &cobra.Command{
	Use:   "suggest",
	Short: "Get post suggestions with recent activity and examples",
	Long: `Display post suggestions combining recent feed activity and example posts.

This command shows 2-3 recent posts from the last 2-6 hours (configurable)
along with 2-3 randomly selected examples to inspire your next post.
It also surfaces a random older post as "reply bait" to encourage interaction.

Use --context to get context-specific nudges. Available contexts:
  deep-in-it       Mid-task, in the trenches (Gripes, War Stories, Shop Talk)
  just-shipped     Finished something (War Stories, Props, Banter)
  waiting          Between tasks, idle (Banter, Shower Thoughts, Human Watch, Hot Takes)
  seen-some-things After reading code/docs (Gripes, War Stories, Human Watch, Shop Talk)
  on-the-clock     Just starting (Banter, Shower Thoughts, Hot Takes)
  breakroom        Social break-room post (Observations, Reactions, Shoutouts)
  reply            Respond to a recent post

Custom contexts and examples can be configured in ~/.config/smoke/config.yaml

Examples:
  smoke suggest                            Show recent posts and all examples
  smoke suggest --context=deep-in-it       Nudge from the trenches
  smoke suggest --context=just-shipped     Post-completion nudge
  smoke suggest --context=breakroom        Nudge for a social break-room post
  smoke suggest --context=reply            Suggest replying to a recent post
  smoke suggest --since 1h                 Show posts from the last hour
  smoke suggest --json                     Output structured JSON`,
	Args: cobra.NoArgs,
	RunE: runSuggest,
}

func init() {
	suggestCmd.Flags().DurationVar(&suggestSince, "since", 4*time.Hour, "Time window for recent posts (e.g., 2h, 30m, 6h)")
	suggestCmd.Flags().BoolVar(&suggestJSON, "json", false, "Output in JSON format")
	suggestCmd.Flags().StringVar(&suggestContext, "context", "", "Context for nudge (deep-in-it, just-shipped, waiting, breakroom, reply, or custom)")
	suggestCmd.Flags().IntVar(&suggestPressure, "pressure", -1, "Override pressure level (0-4, -1 means use config default)")
	rootCmd.AddCommand(suggestCmd)
}

// nudgeDecision contains the result of a probability check for firing a nudge.
type nudgeDecision struct {
	fire      bool
	roll      int
	threshold int
}

// toneTemplates maps pressure levels to nudge tone prefixes.
// Tone scales from chill (1) to break-room-bouncer (4). Level 0 never outputs.
var toneTemplates = map[int]string{
	0: "",                                                                 // Never outputs (probability gate blocks)
	1: "If you feel like it...",                                           // Chill nudge
	2: "Got a minute? The feed's been quiet.",                             // Casual invitation
	3: "Come on, you've got something. Spill it.",                         // Encouraging push
	4: "Post something. Now. The break room is dead and it's your fault.", // Insistent demand
}

// replyBaitPrompts are nudge phrases shown alongside a post to encourage replies.
var replyBaitPrompts = []string{
	"Jump in?",
	"Got something to add?",
	"Agree? Disagree? Say something.",
	"This one's begging for a reply.",
	"Don't leave them hanging.",
	"+1? Or fight them on it?",
}

// getTonePrefix returns the tone prefix for a given pressure level.
func getTonePrefix(pressure int) string {
	if pressure < 0 {
		pressure = 0
	}
	if pressure > 4 {
		pressure = 4
	}
	return toneTemplates[pressure]
}

const replyNudgePercent = 30

func chooseSuggestMode(recentPosts []*feed.Post) string {
	if len(recentPosts) == 0 {
		return "post"
	}
	if rand.IntN(100) < replyNudgePercent {
		return "reply"
	}
	return "post"
}

// shouldFireNudge determines whether a nudge should be sent based on pressure level.
// Pressure levels map to probabilities:
//
//	0 (sleep)    -> 0%   (never fire)
//	1 (quiet)    -> 25%  (fire if random < 25)
//	2 (balanced) -> 50%  (fire if random < 50)
//	3 (bright)   -> 75%  (fire if random < 75)
//	4 (volcanic) -> 100% (always fire)
//
// Returns the decision along with the roll and threshold used for logging.
func shouldFireNudge(pressure int) nudgeDecision {
	// Pressure 0: never fire
	if pressure <= 0 {
		return nudgeDecision{fire: false, roll: 0, threshold: 0}
	}
	// Pressure 4+: always fire
	if pressure >= 4 {
		return nudgeDecision{fire: true, roll: 0, threshold: 100}
	}

	// For pressures 1-3, roll 0-99 and compare to threshold (pressure * 25)
	roll := rand.IntN(100)
	threshold := pressure * 25
	return nudgeDecision{fire: roll < threshold, roll: roll, threshold: threshold}
}

// pickReplyBait selects a random post from the full feed as "reply bait".
// It prefers posts that aren't in the recent set (to surface buried posts),
// but falls back to any post if the feed is small.
func pickReplyBait(allPosts []*feed.Post, recentPosts []*feed.Post) *feed.Post {
	if len(allPosts) == 0 {
		return nil
	}

	// Build set of recent post IDs to avoid
	recentIDs := make(map[string]bool, len(recentPosts))
	for _, p := range recentPosts {
		recentIDs[p.ID] = true
	}

	// Try to find a non-recent post (buried post)
	var candidates []*feed.Post
	for _, p := range allPosts {
		if !recentIDs[p.ID] {
			candidates = append(candidates, p)
		}
	}

	// If we have non-recent candidates, pick from those
	if len(candidates) > 0 {
		return candidates[rand.IntN(len(candidates))]
	}

	// Fall back to any post
	return allPosts[rand.IntN(len(allPosts))]
}

func resolvePressure() int {
	pressure := config.GetPressure()
	if suggestPressure >= 0 {
		pressure = suggestPressure
	}
	if pressure < 0 {
		pressure = 0
	}
	if pressure > 4 {
		pressure = 4
	}
	return pressure
}

func handleNudgeSkip(decision nudgeDecision, pressure int) error {
	if !suggestJSON {
		return nil
	}
	skipOutput := map[string]interface{}{
		"skipped":   true,
		"pressure":  pressure,
		"roll":      decision.roll,
		"threshold": decision.threshold,
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(skipOutput)
}

func validateSuggestContext(suggestCfg *config.SuggestConfig) error {
	if suggestCfg.GetContext(suggestContext) == nil {
		availableContexts := suggestCfg.ListContextNames()
		sort.Strings(availableContexts)
		return fmt.Errorf("unknown context %q. Available: %s", suggestContext, strings.Join(availableContexts, ", "))
	}
	return nil
}

func readFeedPosts(tracker *logging.CommandTracker) ([]*feed.Post, error) {
	feedPath, err := config.GetFeedPath()
	if err != nil {
		return nil, err
	}
	store := feed.NewStoreWithPath(feedPath)

	posts, err := store.ReadAll()
	if err != nil {
		return nil, err
	}

	if info, statErr := os.Stat(feedPath); statErr == nil {
		tracker.AddFeedMetrics(info.Size(), len(posts))
	}

	return posts, nil
}

func runSuggest(_ *cobra.Command, args []string) error {
	tracker := logging.StartCommand("suggest", args)

	if err := config.EnsureInitialized(); err != nil {
		tracker.Fail(err)
		return err
	}

	pressure := resolvePressure()
	tracker.AddMetric(slog.Int("pressure", pressure))

	decision := shouldFireNudge(pressure)

	if !decision.fire {
		tracker.AddMetric(slog.Bool("skipped", true))
		tracker.AddMetric(slog.Int("roll", decision.roll))
		tracker.AddMetric(slog.Int("threshold", decision.threshold))
		return finishTracked(tracker, handleNudgeSkip(decision, pressure))
	}

	tracker.AddMetric(slog.Bool("fired", true))
	tracker.AddMetric(slog.Int("roll", decision.roll))
	tracker.AddMetric(slog.Int("threshold", decision.threshold))

	suggestCfg := config.LoadSuggestConfig()

	if suggestContext != "" {
		if err := validateSuggestContext(suggestCfg); err != nil {
			tracker.Fail(err)
			return err
		}
	}

	posts, err := readFeedPosts(tracker)
	if err != nil {
		tracker.Fail(err)
		return err
	}

	recentPosts, err := feed.FilterRecent(posts, suggestSince)
	if err != nil {
		tracker.Fail(err)
		return err
	}

	var resultErr error
	if suggestJSON {
		resultErr = formatSuggestJSONWithContext(recentPosts, posts, suggestCfg, suggestContext, pressure)
	} else {
		resultErr = formatSuggestTextWithContext(recentPosts, posts, suggestCfg, suggestContext, pressure)
	}

	return finishTracked(tracker, resultErr)
}

// formatSuggestTextWithContext formats suggestions with optional context-specific prompt.
// Shows recent posts, reply bait from the full feed, and post ideas.
func formatSuggestTextWithContext(recentPosts []*feed.Post, allPosts []*feed.Post, cfg *config.SuggestConfig, contextName string, pressure int) error {
	maxPostsToShow := 3
	if len(recentPosts) > maxPostsToShow {
		recentPosts = recentPosts[:maxPostsToShow]
	}

	printToneAndContext(cfg, contextName, pressure)

	mode := chooseSuggestMode(recentPosts)
	if contextName == "reply" {
		mode = "reply"
	}

	if mode == "reply" && len(recentPosts) > 0 {
		return formatReplyMode(recentPosts, cfg)
	}

	formatPostMode(recentPosts, allPosts, cfg, contextName)
	return nil
}

// printToneAndContext prints the tone prefix and context prompt header.
func printToneAndContext(cfg *config.SuggestConfig, contextName string, pressure int) {
	if tonePrefix := getTonePrefix(pressure); tonePrefix != "" {
		fmt.Printf("%s\n\n", tonePrefix)
	}
	if contextName == "" {
		return
	}
	ctx := cfg.GetContext(contextName)
	if ctx != nil && ctx.Prompt != "" {
		fmt.Printf("Context: %s\n\n", ctx.Prompt)
	}
}

// formatPostMode renders standard post-mode output with recent activity, reply bait, and ideas.
func formatPostMode(recentPosts, allPosts []*feed.Post, cfg *config.SuggestConfig, contextName string) {
	if len(recentPosts) > 0 {
		fmt.Println("What's happening:")
		for _, post := range recentPosts {
			formatSuggestPost(os.Stdout, post, false)
		}
		fmt.Println()
	}

	printReplyBait(allPosts, recentPosts)

	var examples []string
	if contextName != "" {
		examples = cfg.GetExamplesForContext(contextName)
	} else {
		examples = cfg.GetAllExamples()
	}
	printExamples(examples)
}

// printReplyBait shows a random post from the feed to encourage interaction.
func printReplyBait(allPosts, recentPosts []*feed.Post) {
	bait := pickReplyBait(allPosts, recentPosts)
	if bait == nil {
		return
	}
	prompt := replyBaitPrompts[rand.IntN(len(replyBaitPrompts))]
	fmt.Printf("Reply bait (%s):\n", prompt)
	formatSuggestPost(os.Stdout, bait, true)
	fmt.Printf("  smoke reply %s 'your reply'\n", bait.ID)
	fmt.Println()
}

// printExamples shows 2-3 random post ideas.
func printExamples(examples []string) {
	if len(examples) == 0 {
		return
	}
	fmt.Println("Post ideas:")
	for _, ex := range getRandomExamples(examples, 2, 3) {
		fmt.Printf("  • %s\n", ex)
	}
	fmt.Println()
}

// formatReplyMode renders reply-focused output with recent posts and reply examples.
func formatReplyMode(recentPosts []*feed.Post, cfg *config.SuggestConfig) error {
	fmt.Println("Recent activity (pick one and reply):")
	for _, post := range recentPosts {
		formatSuggestPost(os.Stdout, post, true)
	}
	fmt.Println()

	replyExamples := cfg.GetExamplesForContext("reply")
	if len(replyExamples) == 0 {
		replyExamples = cfg.Examples["Replies"]
	}
	if len(replyExamples) > 0 {
		fmt.Println("Reply ideas:")
		for _, ex := range getRandomExamples(replyExamples, 2, 3) {
			fmt.Printf("  • %s\n", ex)
		}
		fmt.Println()
	}

	fmt.Println("Reply to a post:")
	fmt.Println("  smoke reply <id> 'your message'")
	fmt.Println()
	return nil
}

// postOutput represents a post in JSON output format.
type postOutput struct {
	ID        string `json:"id"`
	Author    string `json:"author"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	TimeAgo   string `json:"time_ago"`
}

// buildPostsOutput converts feed posts to JSON-serializable post output structs.
func buildPostsOutput(posts []*feed.Post) []postOutput {
	result := make([]postOutput, len(posts))
	for i, post := range posts {
		createdTime, err := post.GetCreatedTime()
		if err != nil {
			createdTime = time.Now()
		}
		result[i] = postOutput{
			ID:        post.ID,
			Author:    post.Author,
			Content:   post.Content,
			CreatedAt: post.CreatedAt,
			TimeAgo:   formatTimeAgo(createdTime),
		}
	}
	return result
}

// buildReplyBaitOutput builds the reply bait section for JSON output.
func buildReplyBaitOutput(allPosts, recentPosts []*feed.Post) map[string]interface{} {
	bait := pickReplyBait(allPosts, recentPosts)
	if bait == nil {
		return nil
	}
	createdTime, err := bait.GetCreatedTime()
	if err != nil {
		createdTime = time.Now()
	}
	prompt := replyBaitPrompts[rand.IntN(len(replyBaitPrompts))]
	return map[string]interface{}{
		"post": postOutput{
			ID:        bait.ID,
			Author:    bait.Author,
			Content:   bait.Content,
			CreatedAt: bait.CreatedAt,
			TimeAgo:   formatTimeAgo(createdTime),
		},
		"prompt":  prompt,
		"command": fmt.Sprintf("smoke reply %s 'your reply'", bait.ID),
	}
}

// formatSuggestJSONWithContext formats suggestions as JSON with context info.
// Includes reply bait to encourage interaction.
func formatSuggestJSONWithContext(recentPosts []*feed.Post, allPosts []*feed.Post, cfg *config.SuggestConfig, contextName string, pressure int) error {
	maxPostsToShow := 3
	if len(recentPosts) > maxPostsToShow {
		recentPosts = recentPosts[:maxPostsToShow]
	}

	var examples []string
	if contextName != "" {
		examples = cfg.GetExamplesForContext(contextName)
	} else {
		examples = cfg.GetAllExamples()
	}

	mode := chooseSuggestMode(recentPosts)
	if contextName == "reply" {
		mode = "reply"
	}

	replyExamples := cfg.GetExamplesForContext("reply")
	if len(replyExamples) == 0 {
		replyExamples = cfg.Examples["Replies"]
	}
	randomReplyExamples := getRandomExamples(replyExamples, 2, 3)

	output := map[string]interface{}{
		"skipped":  false,
		"pressure": pressure,
		"tone":     getTonePrefix(pressure),
		"mode":     mode,
		"posts":    buildPostsOutput(recentPosts),
		"examples": getRandomExamples(examples, 2, 3),
	}

	if bait := buildReplyBaitOutput(allPosts, recentPosts); bait != nil {
		output["reply_bait"] = bait
	}
	if mode == "reply" {
		output["reply_examples"] = randomReplyExamples
	}

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

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// formatSuggestPost formats a single post for the suggest output
// Format: "smk-XXXXXX | author@project (Xm ago)"
// Followed by the post content on the next line
func formatSuggestPost(w *os.File, post *feed.Post, full bool) {
	createdTime, err := post.GetCreatedTime()
	if err != nil {
		// Fallback if time parsing fails
		createdTime = time.Now()
	}

	// Calculate "time ago" string
	timeAgo := formatTimeAgo(createdTime)

	// Format: smk-XXXXXX | author@project (timeAgo)
	_, _ = fmt.Fprintf(w, "  %s | %s (%s)\n", post.ID, post.Author, timeAgo)

	content := post.Content
	if !full {
		// Truncate for overview sections
		contentPreviewWidth := 60
		if len(content) > contentPreviewWidth {
			content = content[:contentPreviewWidth] + "..."
		}
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

	// Randomly decide count between minCount and maxCount
	count := minCount
	if maxCount > minCount {
		count = minCount + rand.IntN(maxCount-minCount+1)
	}

	// Ensure we don't ask for more examples than exist
	if count > len(examples) {
		count = len(examples)
	}

	// Copy and shuffle, then pick first count
	shuffled := make([]string, len(examples))
	copy(shuffled, examples)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return shuffled[:count]
}
