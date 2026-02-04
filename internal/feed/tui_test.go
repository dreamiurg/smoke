package feed

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dreamiurg/smoke/internal/config"
)

// testModel creates a test model with default theme, contrast, layout, and config
func testModel(store *Store) Model {
	theme := GetTheme("dracula")
	contrast := GetContrastLevel("medium")
	layout := GetLayout("comfy")
	cfg := &config.TUIConfig{
		Theme:       "dracula",
		Contrast:    "medium",
		Layout:      "comfy",
		AutoRefresh: true,
	}
	return NewModel(store, theme, contrast, layout, cfg, "test")
}

func TestNewModel(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	theme := GetTheme("dracula")
	contrast := GetContrastLevel("medium")
	layout := GetLayout("comfy")
	cfg := &config.TUIConfig{
		Theme:       "dracula",
		Contrast:    "medium",
		Layout:      "comfy",
		AutoRefresh: true,
	}

	model := NewModel(store, theme, contrast, layout, cfg, "1.0.0")

	if model.theme != theme {
		t.Error("NewModel() did not set theme")
	}
	if model.contrast != contrast {
		t.Error("NewModel() did not set contrast")
	}
	if model.layout != layout {
		t.Error("NewModel() did not set layout")
	}
	if model.store != store {
		t.Error("NewModel() did not set store")
	}
	if model.config != cfg {
		t.Error("NewModel() did not set config")
	}
	if model.version != "1.0.0" {
		t.Error("NewModel() did not set version")
	}
	if model.showHelp {
		t.Error("NewModel() should initialize with showHelp=false")
	}
	if !model.autoRefresh {
		t.Error("NewModel() should initialize autoRefresh from config")
	}
}

func TestModelInit(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
	cmd := model.Init()

	if cmd == nil {
		t.Error("Init() should return a command")
	}
}

func TestModelUpdate_QuitKeys(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")

	tests := []struct {
		name string
		key  string
	}{
		{"quit on q", "q"},
		{"quit on ctrl+c", "ctrl+c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := testModel(store)
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
	model := testModel(store)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")}

	_, cmd := model.Update(msg)
	if cmd == nil {
		t.Error("Update(r) should return refresh command")
	}
}

func TestModelUpdate_ThemeCycling(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
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

func TestModelUpdate_CopyMenu(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)

	// Create and load a post
	post, _ := NewPost("test", "project", "sfx", "hello")
	_ = store.Append(post)
	loadMsg := loadPostsMsg{posts: []*Post{post}}
	updated, _ := model.Update(loadMsg)
	model = updated.(Model)

	// Press 'c' to open copy menu
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")}
	updated, _ = model.Update(msg)
	updatedModel := updated.(Model)

	if !updatedModel.showCopyMenu {
		t.Error("Update(c) should open copy menu when post is selected")
	}
	if updatedModel.copyMenuIndex != 0 {
		t.Error("Update(c) should set copyMenuIndex to 0")
	}
}

func TestModelUpdate_LayoutCycling(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
	initialLayout := model.config.Layout

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")}
	updated, _ := model.Update(msg)
	updatedModel := updated.(Model)

	if updatedModel.config.Layout == initialLayout {
		t.Error("Update(l) should cycle layout")
	}
	if updatedModel.layout.Name != updatedModel.config.Layout {
		t.Error("Update(l) should update model layout to match config")
	}
}

func TestModelUpdate_AutoRefreshToggle(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
	initialAutoRefresh := model.autoRefresh

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}
	updated, _ := model.Update(msg)
	updatedModel := updated.(Model)

	if updatedModel.autoRefresh == initialAutoRefresh {
		t.Error("Update(a) should toggle autoRefresh")
	}
	if updatedModel.config.AutoRefresh != updatedModel.autoRefresh {
		t.Error("Update(a) should update config.AutoRefresh to match model")
	}
}

func TestModelUpdate_HelpToggle(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)

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
	model := testModel(store)
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
	model := testModel(store)

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
	model := testModel(store)
	model.autoRefresh = true

	msg := tickMsg(time.Now())
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("Update(tickMsg) with autoRefresh=true should return load command")
	}
}

func TestModelUpdate_TickMsgDisabled(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
	model.autoRefresh = false

	msg := tickMsg(time.Now())
	_, cmd := model.Update(msg)

	if cmd != nil {
		t.Error("Update(tickMsg) with autoRefresh=false should return nil command")
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
	model := testModel(store)

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
	model := testModel(store)
	view := model.View()

	if !strings.Contains(view, "Initializing") {
		t.Error("View() should show 'Initializing' when width/height are 0")
	}
}

func TestModelView_NoPosts(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
	model.width = 80
	model.height = 24

	view := model.View()

	if !strings.Contains(view, "No posts yet") {
		t.Error("View() should show 'No posts yet' when there are no posts")
	}
}

func TestModelView_WithPosts(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
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
	model := testModel(store)
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

func TestModelFormatPost(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
	model.width = 80

	post := &Post{
		ID:        "smk-test123",
		Author:    "test-author",
		Project:   "smoke",
		Suffix:    "test",
		Content:   "hello world",
		CreatedAt: "2026-01-30T09:24:00Z",
	}

	lines := model.formatPost(post)

	if len(lines) == 0 {
		t.Error("formatPost() should return at least one line")
	}

	combined := strings.Join(lines, "\n")
	if !strings.Contains(combined, "hello world") {
		t.Error("formatPost() should include post content")
	}
}

func TestModelFormatPost_LongContent(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
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

	lines := model.formatPost(post)

	if len(lines) <= 1 {
		t.Error("formatPost() should wrap long content to multiple lines")
	}
}

func TestModelFormatReply(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
	model.width = 80

	replyPost := &Post{
		ID:        "smk-reply",
		Author:    "reply-author",
		Project:   "smoke",
		Suffix:    "reply",
		Content:   "reply content",
		ParentID:  "smk-parent",
		CreatedAt: "2026-01-30T09:05:00Z",
	}

	lines := model.formatReply(replyPost)

	if len(lines) == 0 {
		t.Error("formatReply() should return at least one line")
	}

	combined := strings.Join(lines, "\n")
	if !strings.Contains(combined, "reply content") {
		t.Error("formatReply() should include reply content")
	}
	if !strings.Contains(combined, "└─") {
		t.Error("formatReply() should include reply tree prefix")
	}
}

func TestStyleTimestamp(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)

	result := model.styleTimestamp("09:24")

	if result == "" {
		t.Error("styleTimestamp() should return styled text")
	}
}

func TestStyleAuthor(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)

	result := model.styleAuthor("test-author")

	if result == "" {
		t.Error("styleAuthor() should return styled text")
	}
}

func TestStyleIdentity(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)

	post := &Post{Author: "test-author", Suffix: "smoke"}
	result := model.styleIdentity(post)

	if result == "" {
		t.Error("styleIdentity() should return styled text")
	}
}

func TestRenderStatusBar(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
	model.width = 80

	result := model.renderStatusBar()

	if result == "" {
		t.Error("renderStatusBar() should return status bar")
	}
	if !strings.Contains(result, "Layout") {
		t.Error("renderStatusBar() should show layout keybinding")
	}
}

func TestRenderStatusBar_WithError(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
	model.width = 80
	model.err = errors.New("config save failed")

	result := model.renderStatusBar()

	if result == "" {
		t.Error("renderStatusBar() should return status bar even with error")
	}
}

func TestRenderHeader(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
	model.width = 80

	result := model.renderHeader()

	if result == "" {
		t.Error("renderHeader() should return header bar")
	}
	if !strings.Contains(result, "SMOKE") {
		t.Error("renderHeader() should show smoke title")
	}
}

func TestRenderHelpOverlay(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
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
	if !strings.Contains(result, "Auto:") {
		t.Error("renderHelpOverlay() should show auto-refresh status")
	}
	if !strings.Contains(result, "Layout:") {
		t.Error("renderHelpOverlay() should show layout")
	}
	if !strings.Contains(result, "Cycle layout") {
		t.Error("renderHelpOverlay() should show layout cycling keybinding")
	}
	if !strings.Contains(result, "Copy selected post") {
		t.Error("renderHelpOverlay() should show copy keybinding")
	}
}

func TestRenderHelpOverlay_SmallWindow(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
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
	model := testModel(store)
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
	model := testModel(store)
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

func TestModelUpdate_LoadPostsMsgWithError(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)

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

func TestClockTickCmd(t *testing.T) {
	cmd := clockTickCmd()

	if cmd == nil {
		t.Error("clockTickCmd() should return a command")
	}

	// Execute the command to test the tick message creation
	msg := cmd()
	if _, ok := msg.(clockTickMsg); !ok {
		t.Error("clockTickCmd() should return a command that produces clockTickMsg")
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
	model := testModel(store)

	msg := model.loadPostsCmd()

	loadMsg, ok := msg.(loadPostsMsg)
	if !ok {
		t.Error("loadPostsCmd() should return loadPostsMsg")
	}

	if loadMsg.err != nil {
		t.Errorf("loadPostsCmd() with empty store should not error: %v", loadMsg.err)
	}
}

func TestComputeStats(t *testing.T) {
	posts := []*Post{
		{Author: "agent1@project1"},
		{Author: "agent2@project1"},
		{Author: "agent1@project2"},
		{Author: "agent3@project2"},
	}

	stats := ComputeStats(posts)

	if stats.PostCount != 4 {
		t.Errorf("ComputeStats().PostCount = %d, want 4", stats.PostCount)
	}
	if stats.AgentCount != 3 {
		t.Errorf("ComputeStats().AgentCount = %d, want 3", stats.AgentCount)
	}
	if stats.ProjectCount != 2 {
		t.Errorf("ComputeStats().ProjectCount = %d, want 2", stats.ProjectCount)
	}
}

func TestComputeStats_NilPosts(t *testing.T) {
	posts := []*Post{nil, {Author: "agent@project"}, nil}

	stats := ComputeStats(posts)

	if stats.PostCount != 3 {
		t.Errorf("ComputeStats().PostCount = %d, want 3", stats.PostCount)
	}
	if stats.AgentCount != 1 {
		t.Errorf("ComputeStats().AgentCount = %d, want 1", stats.AgentCount)
	}
}

func TestComputeStats_Empty(t *testing.T) {
	stats := ComputeStats(nil)

	if stats.PostCount != 0 {
		t.Errorf("ComputeStats(nil).PostCount = %d, want 0", stats.PostCount)
	}
	if stats.AgentCount != 0 {
		t.Errorf("ComputeStats(nil).AgentCount = %d, want 0", stats.AgentCount)
	}
	if stats.ProjectCount != 0 {
		t.Errorf("ComputeStats(nil).ProjectCount = %d, want 0", stats.ProjectCount)
	}
}

func TestModelFormatPost_AllLayouts(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	post := &Post{
		ID:        "smk-test123",
		Author:    "test-author",
		Project:   "smoke",
		Suffix:    "test",
		Content:   "hello world",
		CreatedAt: "2026-01-30T09:24:00Z",
	}

	layouts := []string{"dense", "comfy", "relaxed"}

	for _, layoutName := range layouts {
		t.Run(layoutName, func(t *testing.T) {
			model := testModel(store)
			model.width = 80
			model.layout = GetLayout(layoutName)

			lines := model.formatPost(post)

			if len(lines) == 0 {
				t.Errorf("formatPost() with layout %q should return at least one line", layoutName)
			}

			combined := strings.Join(lines, "\n")
			if !strings.Contains(combined, "hello world") {
				t.Errorf("formatPost() with layout %q should include post content", layoutName)
			}
		})
	}
}

func TestFormatPostDense(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
	model.width = 80
	model.layout = GetLayout("dense")

	post := &Post{
		ID:        "smk-test123",
		Author:    "test-author",
		Suffix:    "smoke",
		Content:   "hello world",
		CreatedAt: "2026-01-30T09:24:00Z",
	}

	lines := model.formatPostDense(post)

	if len(lines) == 0 {
		t.Error("formatPostDense() should return at least one line")
	}

	combined := strings.Join(lines, "\n")
	if !strings.Contains(combined, "hello world") {
		t.Error("formatPostDense() should include post content")
	}
	if !strings.Contains(combined, ":") {
		t.Error("formatPostDense() should include colon separator")
	}
}

func TestFormatPostComfy(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
	model.width = 80
	model.layout = GetLayout("comfy")

	post := &Post{
		ID:        "smk-test123",
		Author:    "test-author",
		Suffix:    "smoke",
		Content:   "hello world",
		CreatedAt: "2026-01-30T09:24:00Z",
	}

	lines := model.formatPostComfy(post)

	if len(lines) == 0 {
		t.Error("formatPostComfy() should return at least one line")
	}
	combined := strings.Join(lines, "\n")
	if !strings.Contains(combined, "hello world") {
		t.Error("formatPostComfy() should include post content")
	}
}

func TestFormatPostRelaxed(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
	model.width = 80
	model.layout = GetLayout("relaxed")

	post := &Post{
		ID:        "smk-test123",
		Author:    "test-author",
		Suffix:    "smoke",
		Content:   "hello world",
		CreatedAt: "2026-01-30T09:24:00Z",
	}

	lines := model.formatPostRelaxed(post)

	if len(lines) < 2 {
		t.Error("formatPostRelaxed() should return at least 2 lines (author + content)")
	}
	combined := strings.Join(lines, "\n")
	if !strings.Contains(combined, "hello world") {
		t.Error("formatPostRelaxed() should include post content")
	}
}

func TestGetLayout(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"valid dense", "dense", "dense"},
		{"valid comfy", "comfy", "comfy"},
		{"valid relaxed", "relaxed", "relaxed"},
		{"invalid returns default", "nonexistent", "comfy"},
		{"empty returns default", "", "comfy"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layout := GetLayout(tt.input)
			if layout == nil {
				t.Fatal("GetLayout() returned nil")
			}
			if layout.Name != tt.want {
				t.Errorf("GetLayout(%q).Name = %q, want %q", tt.input, layout.Name, tt.want)
			}
		})
	}
}

func TestNextLayout(t *testing.T) {
	tests := []struct {
		name    string
		current string
		want    string
	}{
		{"next after dense", "dense", "comfy"},
		{"next after comfy", "comfy", "relaxed"},
		{"next after relaxed wraps", "relaxed", "dense"},
		{"invalid returns first", "nonexistent", "dense"},
		{"empty returns first", "", "dense"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NextLayout(tt.current)
			if got != tt.want {
				t.Errorf("NextLayout(%q) = %q, want %q", tt.current, got, tt.want)
			}
		})
	}
}

func TestRenderContent(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
	model.width = 80
	model.height = 24

	post, _ := NewPost("test-author", "smoke", "test", "test content")
	model.posts = []*Post{post}

	result := model.renderContent(20, model.contentWidth())

	if result == "" {
		t.Error("renderContent() should return content")
	}
	if !strings.Contains(result, "test content") {
		t.Error("renderContent() should include post content")
	}
}

// Regression tests for TUI fixes

func TestRenderHeader_IncludesVersion(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
	model.width = 100
	model.version = "1.2.3"

	result := model.renderHeader()

	if !strings.Contains(result, "v1.2.3") {
		t.Errorf("renderHeader() should include version v1.2.3, got: %s", result)
	}
	if !strings.Contains(result, "SMOKE") {
		t.Errorf("renderHeader() should include SMOKE label, got: %s", result)
	}
}

func TestRenderHeader_ContainsStats(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
	model.width = 100
	model.unreadCount = 2
	model.unreadAgentCount = 2
	model.nudgeCount = 3

	result := model.renderHeader()

	if !strings.Contains(result, "2 posts") {
		t.Error("renderHeader() should show post count")
	}
	if !strings.Contains(result, "2 agents") {
		t.Error("renderHeader() should show agent count")
	}
	if !strings.Contains(result, "3 nudges") {
		t.Error("renderHeader() should show nudge count")
	}
}

func TestRenderHeader_ContainsClock(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
	model.width = 100

	result := model.renderHeader()

	// Clock format is [HH:MM]
	if !strings.Contains(result, "[") || !strings.Contains(result, ":") {
		t.Error("renderHeader() should contain clock in [HH:MM] format")
	}
}

func TestInitialSelection_Unread(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	theme := GetTheme("dracula")
	contrast := GetContrastLevel("medium")
	layout := GetLayout("comfy")
	cfg := &config.TUIConfig{
		Theme:       "dracula",
		Contrast:    "medium",
		Layout:      "comfy",
		AutoRefresh: false,
	}
	model := NewModel(store, theme, contrast, layout, cfg, "test")

	now := time.Now().UTC()
	posts := []*Post{
		{ID: "1", Content: "post 1", CreatedAt: now.Add(-3 * time.Minute).Format(time.RFC3339)},
		{ID: "2", Content: "post 2", CreatedAt: now.Add(-2 * time.Minute).Format(time.RFC3339)},
		{ID: "3", Content: "post 3", CreatedAt: now.Add(-1 * time.Minute).Format(time.RFC3339)},
		{ID: "4", Content: "post 4", CreatedAt: now.Format(time.RFC3339)},
	}

	// Mark post 2 as last read, expect selection to move to post 3
	model.lastReadPostID = "2"

	updated, _ := model.Update(loadPostsMsg{posts: posts, err: nil})
	model = updated.(Model)
	updated, _ = model.Update(tea.WindowSizeMsg{Width: 80, Height: 10})
	model = updated.(Model)

	if model.selectedPostIndex != 2 {
		t.Errorf("selectedPostIndex = %d, want 2 (first unread)", model.selectedPostIndex)
	}
}

func TestCursorNavigation(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
	model.width = 80
	model.height = 10

	// Add posts and update displayedPosts
	var posts []*Post
	for i := 0; i < 10; i++ {
		posts = append(posts, &Post{ID: string(rune('0' + i)), Content: "post content"})
	}
	model.posts = posts
	model.updateDisplayedPosts()
	model.initialScrollDone = true

	tests := []struct {
		name      string
		key       string
		wantDelta int // expected change in selectedPostIndex
	}{
		{"down arrow", "down", 1},
		{"j key", "j", 1},
		{"up arrow", "up", -1},
		{"k key", "k", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := model
			m.selectedPostIndex = 5 // start in middle

			var msg tea.KeyMsg
			switch tt.key {
			case "up":
				msg = tea.KeyMsg{Type: tea.KeyUp}
			case "down":
				msg = tea.KeyMsg{Type: tea.KeyDown}
			default:
				msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			}

			updated, _ := m.Update(msg)
			updatedModel := updated.(Model)

			expected := 5 + tt.wantDelta
			if updatedModel.selectedPostIndex != expected {
				t.Errorf("%s: selectedPostIndex = %d, want %d", tt.name, updatedModel.selectedPostIndex, expected)
			}
		})
	}
}

func TestPageNavigationMovesSelection(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
	model.width = 80
	model.height = 10

	// Add posts and update displayedPosts
	var posts []*Post
	for i := 0; i < 30; i++ {
		posts = append(posts, &Post{ID: string(rune('0' + (i % 10))), Content: "post content"})
	}
	model.posts = posts
	model.updateDisplayedPosts()
	model.initialScrollDone = true

	// Page down should move selection forward
	model.selectedPostIndex = 0
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	updatedModel := updated.(Model)
	if updatedModel.selectedPostIndex <= 0 {
		t.Error("pgdown should move selection forward")
	}

	// Page up should move selection backward
	updatedModel.selectedPostIndex = 5
	updated, _ = updatedModel.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	updatedModel = updated.(Model)
	if updatedModel.selectedPostIndex >= 5 {
		t.Error("pgup should move selection backward")
	}
}

func TestHomeEndMovesSelection(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
	model.width = 80
	model.height = 10

	var posts []*Post
	for i := 0; i < 10; i++ {
		posts = append(posts, &Post{ID: string(rune('0' + i)), Content: "post content"})
	}
	model.posts = posts
	model.updateDisplayedPosts()
	model.initialScrollDone = true

	model.selectedPostIndex = 5
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyHome})
	updatedModel := updated.(Model)
	if updatedModel.selectedPostIndex != 0 {
		t.Error("home should move selection to top post")
	}

	updatedModel.selectedPostIndex = 3
	updated, _ = updatedModel.Update(tea.KeyMsg{Type: tea.KeyEnd})
	updatedModel = updated.(Model)
	if updatedModel.selectedPostIndex != len(updatedModel.displayedPosts)-1 {
		t.Error("end should move selection to bottom post")
	}
}

// TestRenderPressureIndicator tests the pressure display format
func TestRenderPressureIndicator(t *testing.T) {
	tests := []struct {
		name     string
		pressure int
		wantBlks string // Expected block pattern
	}{
		{"level 0", 0, "[░░░░]"},
		{"level 1", 1, "[▓░░░]"},
		{"level 2", 2, "[▓▓░░]"},
		{"level 3", 3, "[▓▓▓░]"},
		{"level 4", 4, "[▓▓▓▓]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
			model := testModel(store)
			model.pressure = tt.pressure

			output := model.renderPressureIndicator()

			if !strings.Contains(output, tt.wantBlks) {
				t.Errorf("renderPressureIndicator() at level %d: got %q, want to contain %q", tt.pressure, output, tt.wantBlks)
			}
		})
	}
}

// TestModelUpdate_PressureIncrease tests + keybinding increases pressure
func TestModelUpdate_PressureIncrease(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
	model.pressure = 2
	initialPressure := model.pressure

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")}
	updated, _ := model.Update(msg)
	updatedModel := updated.(Model)

	if updatedModel.pressure != initialPressure+1 {
		t.Errorf("Update(+) should increase pressure from %d to %d, got %d", initialPressure, initialPressure+1, updatedModel.pressure)
	}
}

// TestModelUpdate_PressureDecrease tests - keybinding decreases pressure
func TestModelUpdate_PressureDecrease(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
	model.pressure = 2
	initialPressure := model.pressure

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("-")}
	updated, _ := model.Update(msg)
	updatedModel := updated.(Model)

	if updatedModel.pressure != initialPressure-1 {
		t.Errorf("Update(-) should decrease pressure from %d to %d, got %d", initialPressure, initialPressure-1, updatedModel.pressure)
	}
}

// TestModelUpdate_PressureClampUp tests clamping at max pressure
func TestModelUpdate_PressureClampUp(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
	model.pressure = 4

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")}
	updated, _ := model.Update(msg)
	updatedModel := updated.(Model)

	if updatedModel.pressure != 4 {
		t.Errorf("Update(+) at level 4 should stay at 4, got %d", updatedModel.pressure)
	}
}

// TestModelUpdate_PressureClampDown tests clamping at min pressure
func TestModelUpdate_PressureClampDown(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
	model.pressure = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("-")}
	updated, _ := model.Update(msg)
	updatedModel := updated.(Model)

	if updatedModel.pressure != 0 {
		t.Errorf("Update(-) at level 0 should stay at 0, got %d", updatedModel.pressure)
	}
}

// TestModelUpdate_PressureEqualSign tests = keybinding also increases pressure
func TestModelUpdate_PressureEqualSign(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
	model.pressure = 2
	initialPressure := model.pressure

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("=")}
	updated, _ := model.Update(msg)
	updatedModel := updated.(Model)

	if updatedModel.pressure != initialPressure+1 {
		t.Errorf("Update(=) should increase pressure from %d to %d, got %d", initialPressure, initialPressure+1, updatedModel.pressure)
	}
}

// TestUpdateDisplayedPosts tests the displayedPosts update logic
func TestUpdateDisplayedPosts(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)

	// Create posts
	var posts []*Post
	for i := 0; i < 5; i++ {
		post, _ := NewPost("author", "project", "sfx", "content")
		posts = append(posts, post)
	}
	model.posts = posts

	t.Run("updates displayedPosts", func(t *testing.T) {
		model.updateDisplayedPosts()

		if len(model.displayedPosts) != 5 {
			t.Errorf("displayedPosts should have 5 posts, got %d", len(model.displayedPosts))
		}
	})

	t.Run("clamps selectedPostIndex", func(t *testing.T) {
		model.selectedPostIndex = 100 // Beyond range
		model.updateDisplayedPosts()

		if model.selectedPostIndex >= len(model.displayedPosts) {
			t.Error("selectedPostIndex should be clamped to valid range")
		}
	})

	t.Run("handles empty posts", func(t *testing.T) {
		model.posts = nil
		model.updateDisplayedPosts()

		if model.displayedPosts != nil {
			t.Error("displayedPosts should be nil when posts is empty")
		}
		if model.selectedPostIndex != 0 {
			t.Error("selectedPostIndex should be 0 when posts is empty")
		}
	})
}

// TestBuildAllContentLinesWithPosts tests content line building with post tracking
func TestBuildAllContentLinesWithPosts(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
	model.width = 80
	model.height = 24

	// Create posts
	var posts []*Post
	for i := 0; i < 3; i++ {
		post, _ := NewPost("author", "project", "sfx", "content")
		posts = append(posts, post)
	}
	model.posts = posts
	model.updateDisplayedPosts()

	lines := model.buildAllContentLinesWithPosts()

	if len(lines) == 0 {
		t.Error("buildAllContentLinesWithPosts should return content lines")
	}

	// Check that some lines have valid post indices
	hasPostLines := false
	for _, line := range lines {
		if line.postIndex >= 0 {
			hasPostLines = true
			break
		}
	}
	if !hasPostLines {
		t.Error("buildAllContentLinesWithPosts should have lines with post indices")
	}
}

// TestFormatPostWithSelection tests selection highlighting
func TestFormatPostWithSelection(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
	model.width = 80

	post, _ := NewPost("author", "project", "sfx", "test content")

	t.Run("selected post is highlighted without layout shift", func(t *testing.T) {
		lines := model.formatPostWithSelection(post, true)
		if len(lines) == 0 {
			t.Fatal("formatPostWithSelection should return lines")
		}
		if strings.HasPrefix(lines[0], "▶") {
			t.Error("selected post should not use a triangle indicator")
		}
		if !strings.Contains(lines[0], "test content") {
			t.Error("selected post should include content")
		}
		if model.width > 0 && lipgloss.Width(lines[0]) != model.contentWidth() {
			t.Error("selected post line should be padded to full width")
		}
	})

	t.Run("unselected post has no indicator", func(t *testing.T) {
		lines := model.formatPostWithSelection(post, false)
		if len(lines) == 0 {
			t.Fatal("formatPostWithSelection should return lines")
		}
		if strings.HasPrefix(lines[0], "▶") {
			t.Error("unselected post should not have selection indicator at start")
		}
	})
}

// TestHandleCopyMenuKey tests copy menu key handling
func TestHandleCopyMenuKey(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)

	// Create and load a post
	post, _ := NewPost("author", "project", "sfx", "content")
	model.posts = []*Post{post}
	model.updateDisplayedPosts()
	model.showCopyMenu = true

	t.Run("escape closes menu", func(t *testing.T) {
		m := model
		msg := tea.KeyMsg{Type: tea.KeyEscape}
		updated, _ := m.handleCopyMenuKey(msg)
		updatedModel := updated.(Model)

		if updatedModel.showCopyMenu {
			t.Error("Escape should close copy menu")
		}
	})

	t.Run("q closes menu", func(t *testing.T) {
		m := model
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
		updated, _ := m.handleCopyMenuKey(msg)
		updatedModel := updated.(Model)

		if updatedModel.showCopyMenu {
			t.Error("q should close copy menu")
		}
	})

	t.Run("down moves menu selection", func(t *testing.T) {
		m := model
		m.copyMenuIndex = 0
		msg := tea.KeyMsg{Type: tea.KeyDown}
		updated, _ := m.handleCopyMenuKey(msg)
		updatedModel := updated.(Model)

		if updatedModel.copyMenuIndex != 1 {
			t.Errorf("Down should move menu index to 1, got %d", updatedModel.copyMenuIndex)
		}
	})

	t.Run("up moves menu selection", func(t *testing.T) {
		m := model
		m.copyMenuIndex = 1
		msg := tea.KeyMsg{Type: tea.KeyUp}
		updated, _ := m.handleCopyMenuKey(msg)
		updatedModel := updated.(Model)

		if updatedModel.copyMenuIndex != 0 {
			t.Errorf("Up should move menu index to 0, got %d", updatedModel.copyMenuIndex)
		}
	})

	t.Run("number keys select option", func(t *testing.T) {
		tests := []struct {
			key       string
			wantIndex int
		}{
			{"1", 0},
			{"2", 1},
			{"3", 2},
		}

		for _, tt := range tests {
			m := model
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			updated, _ := m.handleCopyMenuKey(msg)
			updatedModel := updated.(Model)

			if updatedModel.showCopyMenu {
				t.Errorf("Key %s should close copy menu", tt.key)
			}
			// Note: We can't easily test the actual copy since it requires clipboard
			// but we can verify the menu closed and an action was attempted
		}
	})
}

// TestRenderCopyMenuOverlay tests copy menu rendering
func TestRenderCopyMenuOverlay(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
	model.width = 80
	model.height = 24
	model.showCopyMenu = true
	model.copyMenuIndex = 0

	result := model.renderCopyMenuOverlay()

	if result == "" {
		t.Error("renderCopyMenuOverlay should return content")
	}
	if !strings.Contains(result, "Copy Post") {
		t.Error("renderCopyMenuOverlay should contain title")
	}
	if !strings.Contains(result, "Text") {
		t.Error("renderCopyMenuOverlay should contain Text option")
	}
	if !strings.Contains(result, "Square") {
		t.Error("renderCopyMenuOverlay should contain Square option")
	}
	if !strings.Contains(result, "Landscape") {
		t.Error("renderCopyMenuOverlay should contain Landscape option")
	}
}

// TestCopyConfirmationClears tests that copy confirmation clears on key press
func TestCopyConfirmationClears(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
	model.copyConfirmation = "✓ Copied"

	// Any key should clear the confirmation
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")}
	updated, _ := model.Update(msg)
	updatedModel := updated.(Model)

	if updatedModel.copyConfirmation != "" {
		t.Error("Key press should clear copy confirmation")
	}
}

// TestEnsureSelectedVisible tests auto-scroll behavior
func TestEnsureSelectedVisible(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
	model.width = 80
	model.height = 10 // Small height to force scrolling

	// Create many posts
	var posts []*Post
	for i := 0; i < 20; i++ {
		post, _ := NewPost("author", "project", "sfx", "content line")
		posts = append(posts, post)
	}
	model.posts = posts
	model.updateDisplayedPosts()
	model.initialScrollDone = true

	t.Run("scrolls to keep selected visible", func(t *testing.T) {
		model.selectedPostIndex = 15 // Select a post that's likely off-screen
		model.scrollOffset = 0
		model.ensureSelectedVisible()

		// Scroll offset should have changed to show the selected post
		// We can't predict exact offset, but it should be non-zero
		if model.scrollOffset == 0 && len(model.displayedPosts) > 10 {
			t.Error("ensureSelectedVisible should scroll when selected post is off-screen")
		}
	})

	t.Run("handles empty posts", func(t *testing.T) {
		m := model
		m.displayedPosts = nil
		m.ensureSelectedVisible() // Should not panic
	})
}

// TestCopyMenuDoesNotOpenWithoutPosts tests that copy menu doesn't open without posts
func TestCopyMenuDoesNotOpenWithoutPosts(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
	model.posts = nil
	model.displayedPosts = nil

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")}
	updated, _ := model.Update(msg)
	updatedModel := updated.(Model)

	if updatedModel.showCopyMenu {
		t.Error("Copy menu should not open when there are no posts")
	}
}
