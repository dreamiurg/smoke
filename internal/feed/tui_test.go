package feed

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dreamiurg/smoke/internal/config"
)

func TestNewModel(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	theme := GetTheme("tomorrow-night")
	contrast := GetContrastLevel("medium")
	cfg := &config.TUIConfig{
		Theme:    "tomorrow-night",
		Contrast: "medium",
	}

	model := NewModel(store, theme, contrast, cfg)

	if model.theme != theme {
		t.Error("NewModel() did not set theme")
	}
	if model.contrast != contrast {
		t.Error("NewModel() did not set contrast")
	}
	if model.store != store {
		t.Error("NewModel() did not set store")
	}
	if model.config != cfg {
		t.Error("NewModel() did not set config")
	}
	if model.showHelp {
		t.Error("NewModel() should initialize with showHelp=false")
	}
}

func TestModelInit(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	theme := GetTheme("tomorrow-night")
	contrast := GetContrastLevel("medium")
	cfg := &config.TUIConfig{
		Theme:    "tomorrow-night",
		Contrast: "medium",
	}

	model := NewModel(store, theme, contrast, cfg)
	cmd := model.Init()

	if cmd == nil {
		t.Error("Init() should return a command")
	}
}

func TestModelUpdate_QuitKeys(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	theme := GetTheme("tomorrow-night")
	contrast := GetContrastLevel("medium")
	cfg := &config.TUIConfig{
		Theme:    "tomorrow-night",
		Contrast: "medium",
	}

	tests := []struct {
		name string
		key  string
	}{
		{"quit on q", "q"},
		{"quit on ctrl+c", "ctrl+c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel(store, theme, contrast, cfg)
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			if tt.key == "ctrl+c" {
				msg = tea.KeyMsg{Type: tea.KeyCtrlC}
			}

			_, cmd := model.Update(msg)
			if cmd == nil {
				t.Error("Update() should return quit command")
			}
		})
	}
}

func TestModelUpdate_RefreshKey(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	theme := GetTheme("tomorrow-night")
	contrast := GetContrastLevel("medium")
	cfg := &config.TUIConfig{
		Theme:    "tomorrow-night",
		Contrast: "medium",
	}

	model := NewModel(store, theme, contrast, cfg)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")}

	_, cmd := model.Update(msg)
	if cmd == nil {
		t.Error("Update(r) should return refresh command")
	}
}

func TestModelUpdate_ThemeCycling(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	theme := GetTheme("tomorrow-night")
	contrast := GetContrastLevel("medium")
	cfg := &config.TUIConfig{
		Theme:    "tomorrow-night",
		Contrast: "medium",
	}

	model := NewModel(store, theme, contrast, cfg)
	initialTheme := model.config.Theme

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")}
	updated, _ := model.Update(msg)
	updatedModel := updated.(Model)

	if updatedModel.config.Theme == initialTheme {
		t.Error("Update(t) should cycle theme")
	}
	if updatedModel.theme.Name != updatedModel.config.Theme {
		t.Error("Update(t) should update model theme to match config")
	}
}

func TestModelUpdate_ContrastCycling(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	theme := GetTheme("tomorrow-night")
	contrast := GetContrastLevel("medium")
	cfg := &config.TUIConfig{
		Theme:    "tomorrow-night",
		Contrast: "medium",
	}

	model := NewModel(store, theme, contrast, cfg)
	initialContrast := model.config.Contrast

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")}
	updated, _ := model.Update(msg)
	updatedModel := updated.(Model)

	if updatedModel.config.Contrast == initialContrast {
		t.Error("Update(c) should cycle contrast")
	}
	if updatedModel.contrast.Name != updatedModel.config.Contrast {
		t.Error("Update(c) should update model contrast to match config")
	}
}

func TestModelUpdate_HelpToggle(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	theme := GetTheme("tomorrow-night")
	contrast := GetContrastLevel("medium")
	cfg := &config.TUIConfig{
		Theme:    "tomorrow-night",
		Contrast: "medium",
	}

	model := NewModel(store, theme, contrast, cfg)

	// Toggle help on
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")}
	updated, _ := model.Update(msg)
	updatedModel := updated.(Model)

	if !updatedModel.showHelp {
		t.Error("Update(?) should toggle help on")
	}

	// Toggle help off
	updated, _ = updatedModel.Update(msg)
	updatedModel = updated.(Model)

	if updatedModel.showHelp {
		t.Error("Update(?) should toggle help off")
	}
}

func TestModelUpdate_HelpDismissOnAnyKey(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	theme := GetTheme("tomorrow-night")
	contrast := GetContrastLevel("medium")
	cfg := &config.TUIConfig{
		Theme:    "tomorrow-night",
		Contrast: "medium",
	}

	model := NewModel(store, theme, contrast, cfg)
	model.showHelp = true

	// Press any key other than quit
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")}
	updated, cmd := model.Update(msg)
	updatedModel := updated.(Model)

	if updatedModel.showHelp {
		t.Error("Update() should dismiss help on any key")
	}
	if cmd != nil {
		t.Error("Update() with help visible should not quit on non-quit keys")
	}
}

func TestModelUpdate_WindowResize(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	theme := GetTheme("tomorrow-night")
	contrast := GetContrastLevel("medium")
	cfg := &config.TUIConfig{
		Theme:    "tomorrow-night",
		Contrast: "medium",
	}

	model := NewModel(store, theme, contrast, cfg)

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := model.Update(msg)
	updatedModel := updated.(Model)

	if updatedModel.width != 120 {
		t.Errorf("Update(WindowSizeMsg) width = %d, want 120", updatedModel.width)
	}
	if updatedModel.height != 40 {
		t.Errorf("Update(WindowSizeMsg) height = %d, want 40", updatedModel.height)
	}
}

func TestModelUpdate_TickMsg(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	theme := GetTheme("tomorrow-night")
	contrast := GetContrastLevel("medium")
	cfg := &config.TUIConfig{
		Theme:    "tomorrow-night",
		Contrast: "medium",
	}

	model := NewModel(store, theme, contrast, cfg)

	msg := tickMsg(time.Now())
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("Update(tickMsg) should return load command for auto-refresh")
	}
}

func TestModelUpdate_LoadPostsMsg(t *testing.T) {
	tmpDir := t.TempDir()
	feedPath := tmpDir + "/feed.jsonl"

	// Create empty feed file
	if err := os.WriteFile(feedPath, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create feed file: %v", err)
	}

	store := NewStoreWithPath(feedPath)
	theme := GetTheme("tomorrow-night")
	contrast := GetContrastLevel("medium")
	cfg := &config.TUIConfig{
		Theme:    "tomorrow-night",
		Contrast: "medium",
	}

	model := NewModel(store, theme, contrast, cfg)

	// Add a post to the store
	post, _ := NewPost("test-author", "smoke", "test", "test content")
	if err := store.Append(post); err != nil {
		t.Fatalf("Failed to append post: %v", err)
	}

	// Manually create a loadPostsMsg with loaded posts
	posts, err := store.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read posts: %v", err)
	}
	loadMsg := loadPostsMsg{posts: posts, err: nil}

	updated, _ := model.Update(loadMsg)
	updatedModel := updated.(Model)

	if len(updatedModel.posts) != 1 {
		t.Errorf("Update(loadPostsMsg) posts length = %d, want 1", len(updatedModel.posts))
	}
	if len(updatedModel.posts) > 0 && updatedModel.posts[0].Content != "test content" {
		t.Errorf("Update(loadPostsMsg) post content = %q, want %q", updatedModel.posts[0].Content, "test content")
	}
}

func TestModelView_Initializing(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	theme := GetTheme("tomorrow-night")
	contrast := GetContrastLevel("medium")
	cfg := &config.TUIConfig{
		Theme:    "tomorrow-night",
		Contrast: "medium",
	}

	model := NewModel(store, theme, contrast, cfg)
	view := model.View()

	if !strings.Contains(view, "Initializing") {
		t.Error("View() should show 'Initializing' when width/height are 0")
	}
}

func TestModelView_NoPosts(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	theme := GetTheme("tomorrow-night")
	contrast := GetContrastLevel("medium")
	cfg := &config.TUIConfig{
		Theme:    "tomorrow-night",
		Contrast: "medium",
	}

	model := NewModel(store, theme, contrast, cfg)
	model.width = 80
	model.height = 24

	view := model.View()

	if !strings.Contains(view, "No posts yet") {
		t.Error("View() should show 'No posts yet' when there are no posts")
	}
}

func TestModelView_WithPosts(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	theme := GetTheme("tomorrow-night")
	contrast := GetContrastLevel("medium")
	cfg := &config.TUIConfig{
		Theme:    "tomorrow-night",
		Contrast: "medium",
	}

	model := NewModel(store, theme, contrast, cfg)
	model.width = 80
	model.height = 24

	// Add a post
	post, _ := NewPost("test-author", "smoke", "test", "test content")
	model.posts = []*Post{post}

	view := model.View()

	if !strings.Contains(view, "test content") {
		t.Error("View() should show post content")
	}
}

func TestModelView_HelpOverlay(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	theme := GetTheme("tomorrow-night")
	contrast := GetContrastLevel("medium")
	cfg := &config.TUIConfig{
		Theme:    "tomorrow-night",
		Contrast: "medium",
	}

	model := NewModel(store, theme, contrast, cfg)
	model.width = 80
	model.height = 24
	model.showHelp = true

	view := model.View()

	if !strings.Contains(view, "Smoke Feed") {
		t.Error("View() with help should show 'Smoke Feed'")
	}
	if !strings.Contains(view, "Quit") {
		t.Error("View() with help should show 'Quit'")
	}
	if !strings.Contains(view, "Press any key to close") {
		t.Error("View() with help should show dismiss message")
	}
}

func TestFormatPostForTUI(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	theme := GetTheme("tomorrow-night")
	contrast := GetContrastLevel("medium")
	cfg := &config.TUIConfig{
		Theme:    "tomorrow-night",
		Contrast: "medium",
	}

	model := NewModel(store, theme, contrast, cfg)
	model.width = 80

	post := &Post{
		ID:        "smk-test123",
		Author:    "test-author",
		Project:   "smoke",
		Suffix:    "test",
		Content:   "hello world",
		CreatedAt: "2026-01-30T09:24:00Z",
	}

	lines := model.formatPostForTUI(post)

	if len(lines) == 0 {
		t.Error("formatPostForTUI() should return at least one line")
	}

	combined := strings.Join(lines, "\n")
	if !strings.Contains(combined, "hello world") {
		t.Error("formatPostForTUI() should include post content")
	}
}

func TestFormatPostForTUI_LongContent(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	theme := GetTheme("tomorrow-night")
	contrast := GetContrastLevel("medium")
	cfg := &config.TUIConfig{
		Theme:    "tomorrow-night",
		Contrast: "medium",
	}

	model := NewModel(store, theme, contrast, cfg)
	model.width = 80

	longContent := "This is a very long piece of content that should definitely wrap across multiple lines when displayed in the terminal interface because it exceeds the available width"
	post := &Post{
		ID:        "smk-test123",
		Author:    "test-author",
		Project:   "smoke",
		Suffix:    "test",
		Content:   longContent,
		CreatedAt: "2026-01-30T09:24:00Z",
	}

	lines := model.formatPostForTUI(post)

	if len(lines) <= 1 {
		t.Error("formatPostForTUI() should wrap long content to multiple lines")
	}
}

func TestFormatReplyForTUI(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	theme := GetTheme("tomorrow-night")
	contrast := GetContrastLevel("medium")
	cfg := &config.TUIConfig{
		Theme:    "tomorrow-night",
		Contrast: "medium",
	}

	model := NewModel(store, theme, contrast, cfg)
	model.width = 80

	parentPost := &Post{
		ID:        "smk-parent",
		Author:    "parent-author",
		Project:   "smoke",
		Suffix:    "parent",
		Content:   "parent post",
		CreatedAt: "2026-01-30T09:00:00Z",
	}

	replyPost := &Post{
		ID:        "smk-reply",
		Author:    "reply-author",
		Project:   "smoke",
		Suffix:    "reply",
		Content:   "reply content",
		ParentID:  "smk-parent",
		CreatedAt: "2026-01-30T09:05:00Z",
	}

	lines := model.formatReplyForTUI(parentPost, replyPost)

	if len(lines) == 0 {
		t.Error("formatReplyForTUI() should return at least one line")
	}

	combined := strings.Join(lines, "\n")
	if !strings.Contains(combined, "reply content") {
		t.Error("formatReplyForTUI() should include reply content")
	}
	if !strings.Contains(combined, "└─") {
		t.Error("formatReplyForTUI() should include reply tree prefix")
	}
}

func TestStyleTimestamp(t *testing.T) {
	tmpDir := t.TempDir()
	feedPath := tmpDir + "/feed.jsonl"
	store := NewStoreWithPath(feedPath)
	theme := GetTheme("tomorrow-night")
	contrast := GetContrastLevel("medium")
	cfg := &config.TUIConfig{
		Theme:    "tomorrow-night",
		Contrast: "medium",
	}

	model := NewModel(store, theme, contrast, cfg)

	result := model.styleTimestamp("09:24")

	if result == "" {
		t.Error("styleTimestamp() should return styled text")
	}
	// Note: lipgloss may return plain text in test environment without TTY
	// The function call itself is what we're testing, not the exact rendering
}

func TestStyleAuthor(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	theme := GetTheme("tomorrow-night")
	contrast := GetContrastLevel("medium")
	cfg := &config.TUIConfig{
		Theme:    "tomorrow-night",
		Contrast: "medium",
	}

	model := NewModel(store, theme, contrast, cfg)

	result := model.styleAuthor("test-author")

	if result == "" {
		t.Error("styleAuthor() should return styled text")
	}
	// Note: We can't easily test the exact styling without lipgloss internals,
	// but we can verify it's not empty
}

func TestRenderStatusBar(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	theme := GetTheme("tomorrow-night")
	contrast := GetContrastLevel("medium")
	cfg := &config.TUIConfig{
		Theme:    "tomorrow-night",
		Contrast: "medium",
	}

	model := NewModel(store, theme, contrast, cfg)
	model.width = 80

	result := model.renderStatusBar()

	if result == "" {
		t.Error("renderStatusBar() should return status bar")
	}
	// Status bar should contain help indicators (in some form, styled)
	// We can't easily check styled content, but verify it's not empty
}

func TestRenderStatusBar_WithError(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	theme := GetTheme("tomorrow-night")
	contrast := GetContrastLevel("medium")
	cfg := &config.TUIConfig{
		Theme:    "tomorrow-night",
		Contrast: "medium",
	}

	model := NewModel(store, theme, contrast, cfg)
	model.width = 80
	model.err = errors.New("config save failed")

	result := model.renderStatusBar()

	if result == "" {
		t.Error("renderStatusBar() should return status bar even with error")
	}
}

func TestRenderHelpOverlay(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	theme := GetTheme("tomorrow-night")
	contrast := GetContrastLevel("medium")
	cfg := &config.TUIConfig{
		Theme:    "tomorrow-night",
		Contrast: "medium",
	}

	model := NewModel(store, theme, contrast, cfg)
	model.width = 80
	model.height = 24

	result := model.renderHelpOverlay()

	if result == "" {
		t.Error("renderHelpOverlay() should return help content")
	}
	if !strings.Contains(result, "Smoke Feed") {
		t.Error("renderHelpOverlay() should contain title")
	}
	if !strings.Contains(result, "Theme:") {
		t.Error("renderHelpOverlay() should show current theme")
	}
	if !strings.Contains(result, "Contrast:") {
		t.Error("renderHelpOverlay() should show current contrast")
	}
	if !strings.Contains(result, theme.DisplayName) {
		t.Error("renderHelpOverlay() should show theme display name")
	}
	if !strings.Contains(result, contrast.DisplayName) {
		t.Error("renderHelpOverlay() should show contrast display name")
	}
}

func TestRenderHelpOverlay_SmallWindow(t *testing.T) {
	tmpDir := t.TempDir()
	feedPath := tmpDir + "/feed.jsonl"
	store := NewStoreWithPath(feedPath)
	theme := GetTheme("tomorrow-night")
	contrast := GetContrastLevel("medium")
	cfg := &config.TUIConfig{
		Theme:    "tomorrow-night",
		Contrast: "medium",
	}

	model := NewModel(store, theme, contrast, cfg)
	model.width = 40  // Small width
	model.height = 10 // Small height

	result := model.renderHelpOverlay()

	// Should still render, even if small
	if result == "" {
		t.Error("renderHelpOverlay() should return help content even with small window")
	}
}

func TestModelView_WithManyPosts(t *testing.T) {
	tmpDir := t.TempDir()
	feedPath := tmpDir + "/feed.jsonl"

	// Create empty feed file
	if err := os.WriteFile(feedPath, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create feed file: %v", err)
	}

	store := NewStoreWithPath(feedPath)
	theme := GetTheme("tomorrow-night")
	contrast := GetContrastLevel("medium")
	cfg := &config.TUIConfig{
		Theme:    "tomorrow-night",
		Contrast: "medium",
	}

	model := NewModel(store, theme, contrast, cfg)
	model.width = 80
	model.height = 10 // Small height to trigger truncation

	// Add many posts to exceed available height
	var posts []*Post
	for i := 0; i < 20; i++ {
		post, _ := NewPost("test-author", "smoke", "test", "test content line")
		posts = append(posts, post)
	}
	model.posts = posts

	view := model.View()

	// Should render without error even with many posts
	if view == "" {
		t.Error("View() should render with many posts")
	}
	// Should contain some content
	if !strings.Contains(view, "test content") {
		t.Error("View() should show at least some posts")
	}
}

func TestModelView_WithReplies(t *testing.T) {
	tmpDir := t.TempDir()
	feedPath := tmpDir + "/feed.jsonl"

	// Create empty feed file
	if err := os.WriteFile(feedPath, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create feed file: %v", err)
	}

	store := NewStoreWithPath(feedPath)
	theme := GetTheme("tomorrow-night")
	contrast := GetContrastLevel("medium")
	cfg := &config.TUIConfig{
		Theme:    "tomorrow-night",
		Contrast: "medium",
	}

	model := NewModel(store, theme, contrast, cfg)
	model.width = 80
	model.height = 24

	// Add a parent post and a reply
	parentPost, _ := NewPost("parent-author", "smoke", "parent", "parent content")
	replyPost, _ := NewPost("reply-author", "smoke", "reply", "reply content")
	replyPost.ParentID = parentPost.ID

	model.posts = []*Post{parentPost, replyPost}

	view := model.View()

	if !strings.Contains(view, "parent content") {
		t.Error("View() should show parent post")
	}
	if !strings.Contains(view, "reply content") {
		t.Error("View() should show reply")
	}
}

func TestModelUpdate_TickMsgAutoRefresh(t *testing.T) {
	tmpDir := t.TempDir()
	feedPath := tmpDir + "/feed.jsonl"

	// Create empty feed file
	if err := os.WriteFile(feedPath, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create feed file: %v", err)
	}

	store := NewStoreWithPath(feedPath)
	theme := GetTheme("tomorrow-night")
	contrast := GetContrastLevel("medium")
	cfg := &config.TUIConfig{
		Theme:    "tomorrow-night",
		Contrast: "medium",
	}

	model := NewModel(store, theme, contrast, cfg)

	// Simulate a tick
	msg := tickMsg(time.Now())
	_, cmd := model.Update(msg)

	// Should return a load command for auto-refresh
	if cmd == nil {
		t.Error("Update(tickMsg) should return a load command")
	}
}

func TestModelUpdate_LoadPostsMsgWithError(t *testing.T) {
	tmpDir := t.TempDir()
	feedPath := tmpDir + "/feed.jsonl"
	store := NewStoreWithPath(feedPath)
	theme := GetTheme("tomorrow-night")
	contrast := GetContrastLevel("medium")
	cfg := &config.TUIConfig{
		Theme:    "tomorrow-night",
		Contrast: "medium",
	}

	model := NewModel(store, theme, contrast, cfg)

	// Create a loadPostsMsg with an error
	loadMsg := loadPostsMsg{
		posts: nil,
		err:   errors.New("test error"),
	}

	updated, _ := model.Update(loadMsg)
	updatedModel := updated.(Model)

	// Posts should remain nil when there's an error
	if updatedModel.posts != nil {
		t.Error("Update(loadPostsMsg) with error should not set posts")
	}
}

func TestTickCmd(t *testing.T) {
	cmd := tickCmd()

	if cmd == nil {
		t.Error("tickCmd() should return a command")
	}

	// Execute the command to test the tick message creation
	msg := cmd()
	if _, ok := msg.(tickMsg); !ok {
		t.Error("tickCmd() should return a command that produces tickMsg")
	}
}

func TestLoadPostsCmd(t *testing.T) {
	tmpDir := t.TempDir()
	feedPath := tmpDir + "/feed.jsonl"

	// Create empty feed file
	if err := os.WriteFile(feedPath, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create feed file: %v", err)
	}

	store := NewStoreWithPath(feedPath)
	theme := GetTheme("tomorrow-night")
	contrast := GetContrastLevel("medium")
	cfg := &config.TUIConfig{
		Theme:    "tomorrow-night",
		Contrast: "medium",
	}

	model := NewModel(store, theme, contrast, cfg)

	msg := model.loadPostsCmd()

	loadMsg, ok := msg.(loadPostsMsg)
	if !ok {
		t.Error("loadPostsCmd() should return loadPostsMsg")
	}

	if loadMsg.err != nil {
		t.Errorf("loadPostsCmd() with empty store should not error: %v", loadMsg.err)
	}
}
