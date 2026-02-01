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

func TestModelUpdate_ContrastCycling(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
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
	if !strings.Contains(result, "(l)ayout:") {
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
	if !strings.Contains(result, "smoke") {
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
	if !strings.Contains(result, "Contrast:") {
		t.Error("renderHelpOverlay() should show current contrast")
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

	result := model.renderContent(20)

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

	if !strings.Contains(result, "[smoke v1.2.3]") {
		t.Errorf("renderHeader() should include version badge [smoke v1.2.3], got: %s", result)
	}
}

func TestRenderHeader_ContainsStats(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
	model.width = 100
	model.posts = []*Post{
		{Author: "agent1@proj1"},
		{Author: "agent2@proj1"},
	}

	result := model.renderHeader()

	if !strings.Contains(result, "2 posts") {
		t.Error("renderHeader() should show post count")
	}
	if !strings.Contains(result, "2 agents") {
		t.Error("renderHeader() should show agent count")
	}
	if !strings.Contains(result, "1 project") {
		t.Error("renderHeader() should show project count")
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

func TestInitialScrollPosition_NewestOnTop(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	theme := GetTheme("dracula")
	contrast := GetContrastLevel("medium")
	layout := GetLayout("comfy")
	cfg := &config.TUIConfig{
		Theme:       "dracula",
		Contrast:    "medium",
		Layout:      "comfy",
		AutoRefresh: false,
		NewestOnTop: true,
	}
	model := NewModel(store, theme, contrast, layout, cfg, "test")

	// Simulate window size message first
	model.width = 80
	model.height = 24

	// Then simulate loading posts
	posts := []*Post{
		{ID: "1", Content: "post 1"},
		{ID: "2", Content: "post 2"},
		{ID: "3", Content: "post 3"},
	}

	// Process WindowSizeMsg
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = updated.(Model)

	// Process loadPostsMsg
	updated, _ = model.Update(loadPostsMsg{posts: posts, err: nil})
	model = updated.(Model)

	// With newestOnTop=true, scroll should be at 0 (top)
	if model.scrollOffset != 0 {
		t.Errorf("initial scroll with newestOnTop=true should be 0, got %d", model.scrollOffset)
	}
}

func TestInitialScrollPosition_OldestOnTop(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	theme := GetTheme("dracula")
	contrast := GetContrastLevel("medium")
	layout := GetLayout("comfy")
	cfg := &config.TUIConfig{
		Theme:       "dracula",
		Contrast:    "medium",
		Layout:      "comfy",
		AutoRefresh: false,
		NewestOnTop: false, // oldest on top
	}
	model := NewModel(store, theme, contrast, layout, cfg, "test")

	// Create many posts to exceed screen height
	var posts []*Post
	for i := 0; i < 50; i++ {
		posts = append(posts, &Post{ID: string(rune('0' + i)), Content: "post content that is reasonably long"})
	}

	// Process WindowSizeMsg first
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 10})
	model = updated.(Model)

	// Then process loadPostsMsg
	updated, _ = model.Update(loadPostsMsg{posts: posts, err: nil})
	model = updated.(Model)

	// With newestOnTop=false, scroll should be at maxScrollOffset (bottom)
	maxOffset := model.maxScrollOffset()
	if model.scrollOffset != maxOffset {
		t.Errorf("initial scroll with newestOnTop=false should be %d, got %d", maxOffset, model.scrollOffset)
	}
}

// TestInitialScrollPosition_PostsBeforeWindowSize tests when loadPostsMsg arrives before WindowSizeMsg
func TestInitialScrollPosition_PostsBeforeWindowSize(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	theme := GetTheme("dracula")
	contrast := GetContrastLevel("medium")
	layout := GetLayout("comfy")
	cfg := &config.TUIConfig{
		Theme:       "dracula",
		Contrast:    "medium",
		Layout:      "comfy",
		AutoRefresh: false,
		NewestOnTop: false, // oldest on top, so should scroll to bottom
	}
	model := NewModel(store, theme, contrast, layout, cfg, "test")

	// Create many posts
	var posts []*Post
	for i := 0; i < 50; i++ {
		posts = append(posts, &Post{ID: string(rune('0' + i)), Content: "post content that is reasonably long"})
	}

	// Process loadPostsMsg FIRST (before window size is known)
	// This simulates posts loading quickly before terminal reports size
	updated, _ := model.Update(loadPostsMsg{posts: posts, err: nil})
	model = updated.(Model)

	t.Logf("After loadPostsMsg: scrollOffset=%d, height=%d, initialScrollDone=%v",
		model.scrollOffset, model.height, model.initialScrollDone)

	// THEN process WindowSizeMsg
	updated, _ = model.Update(tea.WindowSizeMsg{Width: 80, Height: 10})
	model = updated.(Model)

	t.Logf("After WindowSizeMsg: scrollOffset=%d, height=%d, maxOffset=%d",
		model.scrollOffset, model.height, model.maxScrollOffset())

	// With newestOnTop=false, scroll should be at maxScrollOffset (bottom)
	maxOffset := model.maxScrollOffset()
	if model.scrollOffset != maxOffset {
		t.Errorf("initial scroll with newestOnTop=false should be %d, got %d", maxOffset, model.scrollOffset)
	}
}

func TestScrollKeys(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
	model.width = 80
	model.height = 10

	// Add many posts
	var posts []*Post
	for i := 0; i < 50; i++ {
		posts = append(posts, &Post{ID: string(rune('0' + i)), Content: "post content"})
	}
	model.posts = posts
	model.initialScrollDone = true

	tests := []struct {
		name      string
		key       string
		wantDelta int // expected change in scrollOffset
	}{
		{"down arrow", "down", 1},
		{"j key", "j", 1},
		{"up arrow", "up", -1},
		{"k key", "k", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := model
			m.scrollOffset = 5 // start in middle

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
			if updatedModel.scrollOffset != expected {
				t.Errorf("%s: scrollOffset = %d, want %d", tt.name, updatedModel.scrollOffset, expected)
			}
		})
	}
}

func TestSortToggleScrollsToNewest(t *testing.T) {
	store := NewStoreWithPath(t.TempDir() + "/feed.jsonl")
	model := testModel(store)
	model.width = 80
	model.height = 10
	model.newestOnTop = true
	model.scrollOffset = 0
	model.initialScrollDone = true

	// Add many posts
	var posts []*Post
	for i := 0; i < 50; i++ {
		posts = append(posts, &Post{ID: string(rune('0' + i)), Content: "post content"})
	}
	model.posts = posts

	// Toggle sort order with 's'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")}
	updated, _ := model.Update(msg)
	updatedModel := updated.(Model)

	// Now newestOnTop should be false, and scroll should be at bottom
	if updatedModel.newestOnTop {
		t.Error("'s' should toggle newestOnTop to false")
	}
	maxOffset := updatedModel.maxScrollOffset()
	if updatedModel.scrollOffset != maxOffset {
		t.Errorf("after toggling to oldestOnTop, scroll should be at bottom (%d), got %d", maxOffset, updatedModel.scrollOffset)
	}
}
