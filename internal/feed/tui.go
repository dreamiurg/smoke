package feed

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dreamiurg/smoke/internal/config"
)

// Model is the Bubbletea model for the TUI feed.
type Model struct {
	posts    []*Post
	theme    *Theme
	contrast *ContrastLevel
	showHelp bool
	width    int
	height   int
	store    *Store
	config   *config.TUIConfig
	err      error
}

// T014: tickMsg is sent every 5 seconds for auto-refresh
type tickMsg time.Time

// loadPostsMsg is sent when posts are loaded
type loadPostsMsg struct {
	posts []*Post
	err   error
}

// NewModel creates a new TUI model with the given store, theme, and contrast level.
func NewModel(store *Store, theme *Theme, contrast *ContrastLevel, cfg *config.TUIConfig) Model {
	return Model{
		posts:    make([]*Post, 0),
		theme:    theme,
		contrast: contrast,
		showHelp: false,
		width:    0,
		height:   0,
		store:    store,
		config:   cfg,
		err:      nil,
	}
}

// T014: tickCmd returns a command that ticks every 5 seconds
func tickCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// loadPostsCmd loads posts from the store
func (m Model) loadPostsCmd() tea.Msg {
	posts, err := m.store.ReadAll()
	return loadPostsMsg{posts: posts, err: err}
}

// Init initializes the model and returns the initial command.
// This includes loading posts and setting up a tick for auto-refresh.
func (m Model) Init() tea.Cmd {
	// Load initial posts and start auto-refresh
	return tea.Batch(m.loadPostsCmd, tickCmd())
}

// Update handles incoming messages and updates the model state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// T041: Handle any-key-to-dismiss when help is visible
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		// T016: Manual refresh on 'r' key
		case "r":
			return m, m.loadPostsCmd

		// T018: Theme cycling (placeholder for later implementation)
		case "t":
			m.config.Theme = NextTheme(m.config.Theme)
			m.theme = GetTheme(m.config.Theme)
			_ = config.SaveTUIConfig(m.config)
			return m, nil

		// T018: Contrast cycling (placeholder for later implementation)
		case "c":
			m.config.Contrast = NextContrastLevel(m.config.Contrast)
			m.contrast = GetContrastLevel(m.config.Contrast)
			_ = config.SaveTUIConfig(m.config)
			return m, nil

		// Help toggle (T038: ? key handler)
		case "?":
			m.showHelp = !m.showHelp
			return m, nil
		}

	// T018: Handle window resize
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	// T015: Handle auto-refresh tick
	case tickMsg:
		return m, m.loadPostsCmd

	// Handle loaded posts
	case loadPostsMsg:
		if msg.err == nil {
			m.posts = msg.posts
		}
		return m, nil
	}

	return m, nil
}

// View renders the current state of the model as a string.
// Returns the layout: feed content area + status bar, or help overlay if visible.
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing...\n"
	}

	// T039-T042: Show help overlay if visible
	if m.showHelp {
		return m.renderHelpOverlay()
	}

	// T013: Render feed content
	var content strings.Builder

	if len(m.posts) == 0 {
		content.WriteString("No posts yet. Exit TUI (q) and try: smoke post \"hello world\"\n")
	} else {
		// Use existing formatting logic adapted for TUI
		// Reset timestamp tracking for fresh rendering
		ResetTimestamp()

		// Build thread structure
		threads := buildThreads(m.posts)

		// Render threads (limited to fit screen)
		// Reserve 1 line for status bar
		availableHeight := m.height - 1
		lineCount := 0

		for i, thread := range threads {
			if lineCount >= availableHeight {
				break
			}

			// Format main post
			postLines := m.formatPostForTUI(thread.post)
			for _, line := range postLines {
				if lineCount >= availableHeight {
					break
				}
				content.WriteString(line)
				content.WriteString("\n")
				lineCount++
			}

			// Format replies
			for _, reply := range thread.replies {
				if lineCount >= availableHeight {
					break
				}
				replyLines := m.formatReplyForTUI(thread.post, reply)
				for _, line := range replyLines {
					if lineCount >= availableHeight {
						break
					}
					content.WriteString(line)
					content.WriteString("\n")
					lineCount++
				}
			}

			// Blank line between threads (not after last)
			if i < len(threads)-1 && lineCount < availableHeight {
				content.WriteString("\n")
				lineCount++
			}
		}
	}

	// T017: Add right-aligned status bar
	statusBar := m.renderStatusBar()

	return content.String() + statusBar
}

// formatPostForTUI formats a single post for TUI display
func (m Model) formatPostForTUI(post *Post) []string {
	var lines []string

	t, err := post.GetCreatedTime()
	var timeStr string
	if err != nil {
		timeStr = "??:??"
	} else {
		timeStr = t.Local().Format("15:04")
	}

	// Only show timestamp if different from previous
	var timeColumn string
	if timeStr != lastTimestamp {
		timeColumn = m.styleTimestamp(timeStr)
		lastTimestamp = timeStr
	} else {
		timeColumn = "     " // 5 spaces
	}

	// Right-align identity
	visibleLen := len(post.Author)
	padding := ""
	if visibleLen < AuthorColumnWidth {
		padding = fmt.Sprintf("%*s", AuthorColumnWidth-visibleLen, "")
	}

	identity := m.styleAuthor(post.Author)
	authorRig := padding + identity

	// Wrap content
	contentLines := wrapText(post.Content, MaxContentWidth)
	for i, line := range contentLines {
		highlightedLine := HighlightAll(line, true) // Always enable color in TUI
		if i == 0 {
			lines = append(lines, fmt.Sprintf("%s %s  %s", timeColumn, authorRig, highlightedLine))
		} else {
			indent := fmt.Sprintf("%*s", ContentColumnStart, "")
			lines = append(lines, fmt.Sprintf("%s%s", indent, highlightedLine))
		}
	}

	return lines
}

// formatReplyForTUI formats a reply for TUI display
func (m Model) formatReplyForTUI(_ *Post, reply *Post) []string {
	var lines []string

	replyTime, err := reply.GetCreatedTime()
	var replyTimeStr string
	if err != nil {
		replyTimeStr = "??:??"
	} else {
		replyTimeStr = replyTime.Local().Format("15:04")
	}

	timestamp := m.styleTimestamp(replyTimeStr)

	// Right-align identity (slightly smaller for indent)
	visibleLen := len(reply.Author)
	replyAuthorWidth := AuthorColumnWidth - 3
	padding := ""
	if visibleLen < replyAuthorWidth {
		padding = fmt.Sprintf("%*s", replyAuthorWidth-visibleLen, "")
	}

	identity := m.styleAuthor(reply.Author)
	authorRig := padding + identity

	// Wrap content
	contentLines := wrapText(reply.Content, MaxContentWidth-5)
	for i, line := range contentLines {
		highlightedLine := HighlightAll(line, true)
		if i == 0 {
			lines = append(lines, fmt.Sprintf("  └─ %s %s  %s", timestamp, authorRig, highlightedLine))
		} else {
			indent := fmt.Sprintf("%*s", ContentColumnStart+5, "")
			lines = append(lines, fmt.Sprintf("%s%s", indent, highlightedLine))
		}
	}

	return lines
}

// styleTimestamp applies theme styling to timestamp
func (m Model) styleTimestamp(s string) string {
	style := lipgloss.NewStyle().Foreground(m.theme.Dim)
	return style.Render(s)
}

// styleAuthor applies theme and contrast styling to author name
func (m Model) styleAuthor(author string) string {
	return ColorizeIdentity(author, m.theme, m.contrast)
}

// renderStatusBar creates the right-aligned status bar
func (m Model) renderStatusBar() string {
	statusText := "q:quit  r:refresh  t:theme  c:contrast  ?:help"

	// Right-align within terminal width
	padding := ""
	if len(statusText) < m.width {
		padding = strings.Repeat(" ", m.width-len(statusText))
	}

	// Apply theme styling
	style := lipgloss.NewStyle().
		Foreground(m.theme.Foreground).
		Background(m.theme.Dim)

	return style.Render(padding + statusText)
}

// renderHelpOverlay creates a centered help overlay with keyboard shortcuts.
// T039-T042: Help overlay with theme and contrast display
func (m Model) renderHelpOverlay() string {
	// Build help content
	helpContent := strings.Builder{}
	helpContent.WriteString("\n")
	helpContent.WriteString("          Smoke Feed\n")
	helpContent.WriteString("\n")
	helpContent.WriteString("   q    Quit\n")
	helpContent.WriteString("   t    Cycle theme\n")
	helpContent.WriteString("   c    Cycle contrast\n")
	helpContent.WriteString("   r    Refresh now\n")
	helpContent.WriteString("   ?    Close this help\n")
	helpContent.WriteString("\n")
	helpContent.WriteString(fmt.Sprintf("   Theme: %s\n", m.theme.DisplayName))
	helpContent.WriteString(fmt.Sprintf("   Contrast: %s\n", m.contrast.DisplayName))
	helpContent.WriteString("\n")
	helpContent.WriteString("       Press any key to close\n")

	// T042: Style help overlay with lipgloss
	helpStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Foreground).
		Padding(1, 2).
		Width(35)

	styledBox := helpStyle.Render(helpContent.String())

	// Center the overlay on screen
	boxHeight := strings.Count(styledBox, "\n") + 1
	boxWidth := 35 + 4 // +4 for padding and borders
	topPadding := (m.height - boxHeight) / 2
	leftPadding := (m.width - boxWidth) / 2

	if leftPadding < 0 {
		leftPadding = 0
	}
	if topPadding < 0 {
		topPadding = 0
	}

	// Build final output with centered box
	var result strings.Builder

	// Add top padding
	for i := 0; i < topPadding; i++ {
		result.WriteString("\n")
	}

	// Add each line with left padding
	for _, line := range strings.Split(styledBox, "\n") {
		if line != "" || strings.HasSuffix(styledBox, "\n") {
			result.WriteString(strings.Repeat(" ", leftPadding))
			result.WriteString(line)
			result.WriteString("\n")
		}
	}

	return result.String()
}
