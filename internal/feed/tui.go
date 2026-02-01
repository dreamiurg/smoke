package feed

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dreamiurg/smoke/internal/config"
)

// Help overlay dimensions
const (
	helpBoxInnerWidth = 35 // Content width inside the help box
	helpBoxPadding    = 4  // Additional width for borders and padding
)

// Model is the Bubbletea model for the TUI feed.
type Model struct {
	posts       []*Post
	theme       *Theme
	contrast    *ContrastLevel
	layout      *LayoutStyle
	showHelp    bool
	autoRefresh bool
	newestOnTop bool // Sort order: true=newest first, false=oldest first
	width       int
	height      int
	store       *Store
	config      *config.TUIConfig
	version     string
	err         error
}

// tickMsg is sent every 5 seconds for auto-refresh
type tickMsg time.Time

// clockTickMsg is sent every second for clock updates
type clockTickMsg time.Time

// loadPostsMsg is sent when posts are loaded
type loadPostsMsg struct {
	posts []*Post
	err   error
}

// NewModel creates a new TUI model with the given store, theme, contrast, layout, and version.
func NewModel(store *Store, theme *Theme, contrast *ContrastLevel, layout *LayoutStyle, cfg *config.TUIConfig, version string) Model {
	return Model{
		theme:       theme,
		contrast:    contrast,
		layout:      layout,
		autoRefresh: cfg.AutoRefresh,
		newestOnTop: cfg.NewestOnTop,
		store:       store,
		config:      cfg,
		version:     version,
	}
}

// tickCmd returns a command that ticks every 5 seconds for auto-refresh
func tickCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// clockTickCmd returns a command that ticks every second for clock updates
func clockTickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return clockTickMsg(t)
	})
}

// loadPostsCmd loads posts from the store
func (m Model) loadPostsCmd() tea.Msg {
	posts, err := m.store.ReadAll()
	return loadPostsMsg{posts: posts, err: err}
}

// Init initializes the model and returns the initial command.
func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{m.loadPostsCmd, clockTickCmd()}
	if m.autoRefresh {
		cmds = append(cmds, tickCmd())
	}
	return tea.Batch(cmds...)
}

// Update handles incoming messages and updates the model state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle any-key-to-dismiss when help is visible
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "r":
			return m, m.loadPostsCmd

		case "a":
			// Toggle auto-refresh
			m.autoRefresh = !m.autoRefresh
			m.config.AutoRefresh = m.autoRefresh
			m.err = config.SaveTUIConfig(m.config)
			if m.autoRefresh {
				return m, tickCmd()
			}
			return m, nil

		case "s":
			// Toggle sort order
			m.newestOnTop = !m.newestOnTop
			m.config.NewestOnTop = m.newestOnTop
			m.err = config.SaveTUIConfig(m.config)
			return m, nil

		case "l":
			m.config.Layout = NextLayout(m.config.Layout)
			m.layout = GetLayout(m.config.Layout)
			m.err = config.SaveTUIConfig(m.config)
			return m, nil

		case "L":
			m.config.Layout = PrevLayout(m.config.Layout)
			m.layout = GetLayout(m.config.Layout)
			m.err = config.SaveTUIConfig(m.config)
			return m, nil

		case "t":
			m.config.Theme = NextTheme(m.config.Theme)
			m.theme = GetTheme(m.config.Theme)
			m.err = config.SaveTUIConfig(m.config)
			return m, nil

		case "T":
			m.config.Theme = PrevTheme(m.config.Theme)
			m.theme = GetTheme(m.config.Theme)
			m.err = config.SaveTUIConfig(m.config)
			return m, nil

		case "c":
			m.config.Contrast = NextContrastLevel(m.config.Contrast)
			m.contrast = GetContrastLevel(m.config.Contrast)
			m.err = config.SaveTUIConfig(m.config)
			return m, nil

		case "C":
			m.config.Contrast = PrevContrastLevel(m.config.Contrast)
			m.contrast = GetContrastLevel(m.config.Contrast)
			m.err = config.SaveTUIConfig(m.config)
			return m, nil

		case "?":
			m.showHelp = !m.showHelp
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tickMsg:
		if m.autoRefresh {
			return m, tea.Batch(m.loadPostsCmd, tickCmd())
		}
		return m, nil

	case clockTickMsg:
		// Just trigger a re-render for clock update
		return m, clockTickCmd()

	case loadPostsMsg:
		if msg.err == nil {
			m.posts = msg.posts
		}
		return m, nil
	}

	return m, nil
}

// View renders the current state of the model as a string.
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing...\n"
	}

	if m.showHelp {
		return m.renderHelpOverlay()
	}

	// Render three sections: header, content, status bar
	header := m.renderHeader()
	statusBar := m.renderStatusBar()

	// Calculate available height for content (total - header - status)
	availableHeight := m.height - 2 // 1 for header, 1 for status

	content := m.renderContent(availableHeight)

	return header + "\n" + content + statusBar
}

// renderHeader creates the header bar with version, stats, and clock
func (m Model) renderHeader() string {
	// Calculate stats
	stats := ComputeStats(m.posts)

	// Format version badge
	versionStr := "[smoke]"
	if m.version != "" {
		versionStr = fmt.Sprintf("[smoke v%s]", m.version)
	}

	// Format stats: N posts | M agents | P projects
	statsStr := fmt.Sprintf("%d posts | %d agents | %d projects",
		stats.PostCount, stats.AgentCount, stats.ProjectCount)

	// Format clock in brackets
	clockStr := fmt.Sprintf("[%s]", time.Now().Local().Format("15:04"))

	// Build header: version + stats on left, clock on right
	leftContent := versionStr + " " + statsStr
	rightContent := clockStr

	// Calculate spacing
	spacing := m.width - len(leftContent) - len(rightContent)
	if spacing < 1 {
		spacing = 1
	}

	headerText := leftContent + strings.Repeat(" ", spacing) + rightContent

	// Style with theme colors
	style := lipgloss.NewStyle().
		Foreground(m.theme.Text).
		Background(m.theme.BackgroundSecondary).
		Width(m.width)

	return style.Render(headerText)
}

// renderStatusBar creates the status bar showing settings and keybindings
func (m Model) renderStatusBar() string {
	// Build status items
	autoStr := "OFF"
	if m.autoRefresh {
		autoStr = "ON"
	}

	sortStr := "old→new"
	if m.newestOnTop {
		sortStr = "new→old"
	}

	layoutName := "comfy"
	if m.layout != nil {
		layoutName = m.layout.Name
	}

	parts := []string{
		fmt.Sprintf("(a)uto: %s", autoStr),
		fmt.Sprintf("(s)ort: %s", sortStr),
		fmt.Sprintf("(l)ayout: %s", layoutName),
		fmt.Sprintf("(t)heme: %s", m.theme.Name),
		fmt.Sprintf("(c)ontrast: %s", m.contrast.Name),
		"(?) help",
		"(q)uit",
	}

	statusText := strings.Join(parts, "  ")

	// Show error if save failed
	if m.err != nil {
		statusText = "⚠ config error  " + statusText
	}

	// Style with theme colors (matching header)
	style := lipgloss.NewStyle().
		Foreground(m.theme.Text).
		Background(m.theme.BackgroundSecondary).
		Width(m.width)

	return style.Render(statusText)
}

// renderContent renders the feed content area
func (m Model) renderContent(availableHeight int) string {
	var content strings.Builder

	if len(m.posts) == 0 {
		content.WriteString("No posts yet. Exit TUI (q) and try: smoke post \"hello world\"\n")
	} else {
		threads := buildThreads(m.posts)

		// Reverse thread order if newestOnTop is true
		// Default (newestOnTop=false): oldest at top, newest at bottom
		// newestOnTop=true: newest at top, oldest at bottom
		if m.newestOnTop {
			for i, j := 0, len(threads)-1; i < j; i, j = i+1, j-1 {
				threads[i], threads[j] = threads[j], threads[i]
			}
		}

		lineCount := 0

		for i, thread := range threads {
			if lineCount >= availableHeight {
				break
			}

			// Format main post using current style
			postLines := m.formatPost(thread.post)
			for _, line := range postLines {
				if lineCount >= availableHeight {
					break
				}
				content.WriteString(line)
				content.WriteString("\n")
				lineCount++
			}

			// Format replies (indented)
			for _, reply := range thread.replies {
				if lineCount >= availableHeight {
					break
				}
				replyLines := m.formatReply(reply)
				for _, line := range replyLines {
					if lineCount >= availableHeight {
						break
					}
					content.WriteString(line)
					content.WriteString("\n")
					lineCount++
				}
			}

			// Blank line between threads
			if i < len(threads)-1 && lineCount < availableHeight {
				content.WriteString("\n")
				lineCount++
			}
		}
	}

	return content.String()
}

// formatPost formats a post according to the current layout
func (m Model) formatPost(post *Post) []string {
	if m.layout == nil {
		return m.formatPostComfy(post)
	}
	switch m.layout.Name {
	case "dense":
		return m.formatPostDense(post)
	case "relaxed":
		return m.formatPostRelaxed(post)
	default:
		return m.formatPostComfy(post)
	}
}

// formatReply formats a reply (indented post)
func (m Model) formatReply(reply *Post) []string {
	lines := m.formatPost(reply)
	indented := make([]string, len(lines))
	for i, line := range lines {
		if i == 0 {
			indented[i] = "  └─ " + line
		} else {
			indented[i] = "     " + line
		}
	}
	return indented
}

// formatPostDense: Most compact - single line with everything inline
// Format: HH:MM author@project: message...
// Continuation lines wrap to column 0 (no alignment padding)
func (m Model) formatPostDense(post *Post) []string {
	termWidth := m.width
	if termWidth <= 0 {
		termWidth = DefaultTerminalWidth
	}

	timeStr := m.styleTimestamp(formatTimestamp(post))
	identity := m.styleIdentity(post)
	prefix := fmt.Sprintf("%s %s: ", timeStr, identity)
	// Calculate raw prefix length (without ANSI codes)
	prefixLen := len(formatTimestamp(post)) + 1 + len(post.Author) + 2 // "HH:MM author: "

	// First line has less space due to prefix
	firstLineWidth := termWidth - prefixLen
	if firstLineWidth < MinContentWidth {
		firstLineWidth = MinContentWidth
	}

	// Continuation lines use full width (wrap to column 0)
	contentLines := wrapTextFirstLineShorter(post.Content, firstLineWidth, termWidth)

	lines := make([]string, 0, len(contentLines))
	for i, line := range contentLines {
		highlighted := HighlightAll(line, true)
		if i == 0 {
			lines = append(lines, prefix+highlighted)
		} else {
			// No padding - wrap to column 0
			lines = append(lines, highlighted)
		}
	}

	return lines
}

// formatPostComfy: Balanced - message starts on same line as identity
// Format: HH:MM  author@project
//
//	message continues here...
func (m Model) formatPostComfy(post *Post) []string {
	termWidth := m.width
	if termWidth <= 0 {
		termWidth = DefaultTerminalWidth
	}

	timeStr := m.styleTimestamp(formatTimestamp(post))
	identity := m.styleIdentity(post)
	// Calculate prefix for first line: "HH:MM  identity "
	prefix := fmt.Sprintf("%s  %s ", timeStr, identity)
	prefixLen := len(formatTimestamp(post)) + 2 + len(post.Author) + 1 + len(post.Suffix) + 1

	contentWidth := termWidth - prefixLen
	if contentWidth < MinContentWidth {
		contentWidth = MinContentWidth
	}
	contentLines := wrapText(post.Content, contentWidth)

	lines := make([]string, 0, len(contentLines))
	for i, line := range contentLines {
		highlighted := HighlightAll(line, true)
		if i == 0 {
			lines = append(lines, prefix+highlighted)
		} else {
			// Continuation lines align with message start
			lines = append(lines, strings.Repeat(" ", prefixLen)+highlighted)
		}
	}

	return lines
}

// formatPostRelaxed: Most spacious - author on separate line, content below
// Format: HH:MM  author@project
//
//	message on next line...
func (m Model) formatPostRelaxed(post *Post) []string {
	termWidth := m.width
	if termWidth <= 0 {
		termWidth = DefaultTerminalWidth
	}

	timeStr := m.styleTimestamp(formatTimestamp(post))
	identity := m.styleIdentity(post)
	contentLines := wrapText(post.Content, termWidth-2)

	lines := make([]string, 0, 1+len(contentLines))
	// First line: timestamp and identity only
	lines = append(lines, fmt.Sprintf("%s  %s", timeStr, identity))
	// Content on subsequent lines
	for _, line := range contentLines {
		lines = append(lines, HighlightAll(line, true))
	}

	return lines
}

// styleTimestamp applies theme styling to timestamp
func (m Model) styleTimestamp(s string) string {
	style := lipgloss.NewStyle().Foreground(m.theme.TextMuted)
	return style.Render(s)
}

// styleAuthor applies theme and contrast styling to author name
func (m Model) styleAuthor(author string) string {
	return ColorizeIdentity(author, m.theme, m.contrast)
}

// styleIdentity formats and styles author@project
func (m Model) styleIdentity(post *Post) string {
	// post.Author already contains @project (e.g., "claude-rich-crane@smoke")
	// Use ColorizeIdentity which splits it properly, not ColorizeFullIdentity
	return ColorizeIdentity(post.Author, m.theme, m.contrast)
}

// renderHelpOverlay creates a centered help overlay
func (m Model) renderHelpOverlay() string {
	autoStr := "OFF"
	if m.autoRefresh {
		autoStr = "ON"
	}

	sortStr := "old→new"
	if m.newestOnTop {
		sortStr = "new→old"
	}

	layoutName := "Comfy"
	if m.layout != nil {
		layoutName = m.layout.DisplayName
	}

	helpContent := strings.Builder{}
	helpContent.WriteString("\n")
	helpContent.WriteString("          Smoke Feed\n")
	helpContent.WriteString("\n")
	helpContent.WriteString("   q      Quit\n")
	helpContent.WriteString("   a      Toggle auto-refresh\n")
	helpContent.WriteString("   s      Toggle sort order\n")
	helpContent.WriteString("   l/L    Cycle layout (fwd/back)\n")
	helpContent.WriteString("   t/T    Cycle theme (fwd/back)\n")
	helpContent.WriteString("   c/C    Cycle contrast (fwd/back)\n")
	helpContent.WriteString("   r      Refresh now\n")
	helpContent.WriteString("   ?      Close this help\n")
	helpContent.WriteString("\n")
	helpContent.WriteString(fmt.Sprintf("   Auto: %s\n", autoStr))
	helpContent.WriteString(fmt.Sprintf("   Sort: %s\n", sortStr))
	helpContent.WriteString(fmt.Sprintf("   Layout: %s\n", layoutName))
	helpContent.WriteString(fmt.Sprintf("   Theme: %s\n", m.theme.DisplayName))
	helpContent.WriteString(fmt.Sprintf("   Contrast: %s\n", m.contrast.DisplayName))
	helpContent.WriteString("\n")
	helpContent.WriteString("       Press any key to close\n")

	helpStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Text).
		Padding(1, 2).
		Width(helpBoxInnerWidth)

	styledBox := helpStyle.Render(helpContent.String())

	boxHeight := strings.Count(styledBox, "\n") + 1
	boxWidth := helpBoxInnerWidth + helpBoxPadding
	topPadding := (m.height - boxHeight) / 2
	leftPadding := (m.width - boxWidth) / 2

	if leftPadding < 0 {
		leftPadding = 0
	}
	if topPadding < 0 {
		topPadding = 0
	}

	var result strings.Builder
	for i := 0; i < topPadding; i++ {
		result.WriteString("\n")
	}

	for _, line := range strings.Split(styledBox, "\n") {
		if line != "" || strings.HasSuffix(styledBox, "\n") {
			result.WriteString(strings.Repeat(" ", leftPadding))
			result.WriteString(line)
			result.WriteString("\n")
		}
	}

	return result.String()
}
