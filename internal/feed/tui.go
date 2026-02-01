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
	posts             []*Post
	theme             *Theme
	contrast          *ContrastLevel
	layout            *LayoutStyle
	showHelp          bool
	autoRefresh       bool
	newestOnTop       bool // Sort order: true=newest first, false=oldest first
	scrollOffset      int  // Number of lines scrolled from top
	initialScrollDone bool // Track if initial scroll position has been set
	width             int
	height            int
	store             *Store
	config            *config.TUIConfig
	pressure          int // Current pressure level (0-4)
	version           string
	err               error
	// Cursor state for post selection (spec 008)
	selectedPostIndex int     // Index of currently selected post in displayedPosts
	displayedPosts    []*Post // Posts currently visible (sorted order)
	// Copy menu state (spec 008)
	showCopyMenu     bool   // Whether copy menu is visible
	copyMenuIndex    int    // Currently highlighted menu option (0=text, 1=square, 2=landscape)
	copyConfirmation string // Confirmation message after copy (e.g., "Copied as text!")
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
		pressure:    config.GetPressure(),
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
			// Toggle sort order and scroll to show newest posts
			m.newestOnTop = !m.newestOnTop
			m.config.NewestOnTop = m.newestOnTop
			m.err = config.SaveTUIConfig(m.config)
			// Update displayedPosts with new sort order
			m.updateDisplayedPosts()
			// Scroll to show newest posts: top if newestOnTop, bottom otherwise
			if m.newestOnTop {
				m.scrollOffset = 0
			} else {
				m.scrollOffset = m.maxScrollOffset()
			}
			return m, nil

		case "up", "k":
			// Move cursor up (select previous post)
			if m.selectedPostIndex > 0 {
				m.selectedPostIndex--
				m.ensureSelectedVisible()
			}
			return m, nil

		case "down", "j":
			// Move cursor down (select next post)
			if m.selectedPostIndex < len(m.displayedPosts)-1 {
				m.selectedPostIndex++
				m.ensureSelectedVisible()
			}
			return m, nil

		case "pgup", "ctrl+u":
			// Scroll up one page
			pageSize := m.height - 2 // Account for header and status bar
			m.scrollOffset -= pageSize
			if m.scrollOffset < 0 {
				m.scrollOffset = 0
			}
			return m, nil

		case "pgdown", "ctrl+d":
			// Scroll down one page
			pageSize := m.height - 2
			maxOffset := m.maxScrollOffset()
			m.scrollOffset += pageSize
			if m.scrollOffset > maxOffset {
				m.scrollOffset = maxOffset
			}
			return m, nil

		case "home", "g":
			// Scroll to top
			m.scrollOffset = 0
			return m, nil

		case "end", "G":
			// Scroll to bottom
			m.scrollOffset = m.maxScrollOffset()
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

		case "c", "C":
			// Open copy menu if posts exist and help not showing
			if len(m.displayedPosts) > 0 && !m.showHelp {
				m.showCopyMenu = true
				m.copyMenuIndex = 0 // Default to first option (text)
				m.copyConfirmation = ""
			}
			return m, nil

		case "+", "=":
			// Increase pressure, clamping at max (like volume)
			if m.pressure < 4 {
				m.pressure++
				m.err = config.SetPressure(m.pressure)
			}
			return m, nil

		case "-":
			// Decrease pressure, clamping at min (like volume)
			if m.pressure > 0 {
				m.pressure--
				m.err = config.SetPressure(m.pressure)
			}
			return m, nil

		case "?":
			m.showHelp = !m.showHelp
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Set initial scroll position once we know the height and have posts
		if !m.initialScrollDone && len(m.posts) > 0 {
			m.initialScrollDone = true
			if m.newestOnTop {
				m.scrollOffset = 0
			} else {
				m.scrollOffset = m.maxScrollOffset()
			}
		}

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
			oldCount := len(m.posts)
			m.posts = msg.posts
			// Update displayedPosts with current sort order
			m.updateDisplayedPosts()
			// Set initial scroll position once we have posts and know height
			if !m.initialScrollDone && m.height > 0 && len(m.posts) > 0 {
				m.initialScrollDone = true
				if m.newestOnTop {
					m.scrollOffset = 0
				} else {
					m.scrollOffset = m.maxScrollOffset()
				}
			} else if len(m.posts) > oldCount && m.height > 0 {
				// Auto-scroll when NEW posts arrive (after initial load)
				if m.newestOnTop {
					m.scrollOffset = 0
				} else {
					m.scrollOffset = m.maxScrollOffset()
				}
			}
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

	if m.showCopyMenu {
		return m.renderCopyMenuOverlay()
	}

	// Render three sections: header, content, status bar
	header := m.renderHeader()
	statusBar := m.renderStatusBar()

	// Calculate available height for content (total - header - status)
	availableHeight := m.height - 2 // 1 for header, 1 for status

	content := m.renderContent(availableHeight)

	// Use JoinVertical for seamless background colors
	return lipgloss.JoinVertical(lipgloss.Left, header, content, statusBar)
}

// renderPressureIndicator creates a pressure display in the format: (+/-) Pressure [▓▓░░] ⛅
// Uses filled blocks (▓) for active levels and empty blocks (░) for inactive levels.
func (m Model) renderPressureIndicator() string {
	level := config.GetPressureLevel(m.pressure)

	// Build visual blocks: filled for active levels, empty for inactive
	filled := strings.Repeat("▓", m.pressure)
	empty := strings.Repeat("░", 4-m.pressure)
	blocks := "[" + filled + empty + "]"

	// Format: (+/-) Pressure [blocks] emoji
	return fmt.Sprintf("(+/-) Pressure %s %s", blocks, level.Emoji)
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

	// Format pressure indicator
	pressureStr := m.renderPressureIndicator()

	// Format clock in brackets (locale-aware)
	clockStr := fmt.Sprintf("[%s]", FormatTime(time.Now()))

	// Build header: version + stats on left, pressure + clock on right
	leftContent := versionStr + " " + statsStr
	rightContent := pressureStr + " " + clockStr

	// Calculate spacing using display width (handles emojis correctly)
	spacing := m.width - lipgloss.Width(leftContent) - lipgloss.Width(rightContent)
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

// updateDisplayedPosts updates the displayedPosts slice with posts in current display order.
// Posts are flattened from threads (main post + replies) in sort order.
func (m *Model) updateDisplayedPosts() {
	if len(m.posts) == 0 {
		m.displayedPosts = nil
		m.selectedPostIndex = 0
		return
	}

	threads := buildThreads(m.posts)

	// Reverse thread order if newestOnTop is false (threads come newest-first from buildThreads)
	if !m.newestOnTop {
		for i, j := 0, len(threads)-1; i < j; i, j = i+1, j-1 {
			threads[i], threads[j] = threads[j], threads[i]
		}
	}

	// Flatten threads into displayedPosts (main post followed by its replies)
	m.displayedPosts = make([]*Post, 0, len(m.posts))
	for _, thread := range threads {
		m.displayedPosts = append(m.displayedPosts, thread.post)
		m.displayedPosts = append(m.displayedPosts, thread.replies...)
	}

	// Clamp selectedPostIndex to valid range
	if m.selectedPostIndex >= len(m.displayedPosts) {
		m.selectedPostIndex = len(m.displayedPosts) - 1
	}
	if m.selectedPostIndex < 0 {
		m.selectedPostIndex = 0
	}
}

// ensureSelectedVisible adjusts scrollOffset to keep the selected post visible in the viewport.
func (m *Model) ensureSelectedVisible() {
	if m.height == 0 || len(m.displayedPosts) == 0 {
		return
	}

	// Get content lines to find which lines belong to selected post
	contentLines := m.buildAllContentLinesWithPosts()
	if len(contentLines) == 0 {
		return
	}

	// Find first and last line indices of the selected post
	firstLine := -1
	lastLine := -1
	for i, cl := range contentLines {
		if cl.postIndex == m.selectedPostIndex {
			if firstLine == -1 {
				firstLine = i
			}
			lastLine = i
		}
	}

	if firstLine == -1 {
		// Selected post not found in content (shouldn't happen)
		return
	}

	// Calculate visible range
	availableHeight := m.height - 2 // Account for header and status bar
	if availableHeight <= 0 {
		return
	}

	visibleStart := m.scrollOffset
	visibleEnd := m.scrollOffset + availableHeight - 1

	// Scroll up if selected post starts above visible area
	if firstLine < visibleStart {
		m.scrollOffset = firstLine
	}

	// Scroll down if selected post ends below visible area
	if lastLine > visibleEnd {
		m.scrollOffset = lastLine - availableHeight + 1
	}

	// Clamp scroll offset to valid range
	maxOffset := len(contentLines) - availableHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.scrollOffset > maxOffset {
		m.scrollOffset = maxOffset
	}
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

// contentLine holds a line of content and which post index it belongs to (-1 for non-post content)
type contentLine struct {
	text      string
	postIndex int // Index into displayedPosts, -1 for separators/spacers
}

// buildAllContentLines builds all content lines for the feed (used for scrolling)
// Returns both the text lines and which post each line belongs to for selection highlighting.
func (m Model) buildAllContentLinesWithPosts() []contentLine {
	if len(m.posts) == 0 {
		return []contentLine{{text: "No posts yet. Exit TUI (q) and try: smoke post \"hello world\"", postIndex: -1}}
	}

	threads := buildThreads(m.posts)

	// Reverse thread order if newestOnTop is false (threads come newest-first from buildThreads)
	if !m.newestOnTop {
		for i, j := 0, len(threads)-1; i < j; i, j = i+1, j-1 {
			threads[i], threads[j] = threads[j], threads[i]
		}
	}

	var lines []contentLine
	var lastDay time.Time
	postIdx := 0 // Track index into displayedPosts

	for i, thread := range threads {
		// Get post time for day separator (convert to local time for consistent day comparison)
		postTime, err := thread.post.GetCreatedTime()
		if err == nil {
			localTime := postTime.Local()
			postDay := time.Date(localTime.Year(), localTime.Month(), localTime.Day(), 0, 0, 0, 0, localTime.Location())
			// Check if we need a day separator
			if lastDay.IsZero() || !postDay.Equal(lastDay) {
				if i > 0 {
					// Add blank line before separator (except for first post)
					lines = append(lines, contentLine{text: "", postIndex: -1})
				}
				lines = append(lines, contentLine{text: m.formatDaySeparator(localTime), postIndex: -1})
				lastDay = postDay
			}
		}

		// Format main post - track its index
		currentPostIdx := postIdx
		postIdx++
		postLines := m.formatPost(thread.post)
		for _, line := range postLines {
			lines = append(lines, contentLine{text: line, postIndex: currentPostIdx})
		}

		// Format replies (indented) - each reply is also tracked
		for _, reply := range thread.replies {
			replyIdx := postIdx
			postIdx++
			replyLines := m.formatReply(reply)
			for _, line := range replyLines {
				lines = append(lines, contentLine{text: line, postIndex: replyIdx})
			}
		}

		// Blank line between threads (within same day)
		if i < len(threads)-1 {
			lines = append(lines, contentLine{text: "", postIndex: -1})
		}
	}

	return lines
}

// buildAllContentLines builds all content lines for the feed (used for scrolling)
// This is a convenience wrapper that returns just the text lines.
func (m Model) buildAllContentLines() []string {
	contentLines := m.buildAllContentLinesWithPosts()
	lines := make([]string, len(contentLines))
	for i, cl := range contentLines {
		lines[i] = cl.text
	}
	return lines
}

// formatDaySeparator creates a styled day separator line.
// Format: "──── Today ────" centered with decorative lines
func (m Model) formatDaySeparator(t time.Time) string {
	label := DayLabel(t)
	termWidth := m.width
	if termWidth <= 0 {
		termWidth = DefaultTerminalWidth
	}

	// Build separator: "──── Label ────"
	// Minimum width for label plus surrounding spaces and some decorative chars
	minDecor := 4 // At least 4 dashes on each side
	labelWithSpace := " " + label + " "
	availableForDecor := termWidth - len(labelWithSpace)

	var leftDecor, rightDecor string
	if availableForDecor >= minDecor*2 {
		decorLen := availableForDecor / 2
		leftDecor = strings.Repeat("─", decorLen)
		rightDecor = strings.Repeat("─", availableForDecor-decorLen)
	} else {
		// Terminal too narrow - just show label
		leftDecor = "──"
		rightDecor = "──"
	}

	separator := leftDecor + labelWithSpace + rightDecor

	// Style with muted text color
	style := lipgloss.NewStyle().
		Foreground(m.theme.TextMuted).
		Background(m.theme.Background)

	return style.Render(separator)
}

// maxScrollOffset returns the maximum scroll offset based on content size
func (m Model) maxScrollOffset() int {
	allLines := m.buildAllContentLines()
	availableHeight := m.height - 2 // header + status bar
	if availableHeight <= 0 {
		availableHeight = 1
	}
	maxOffset := len(allLines) - availableHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	return maxOffset
}

// renderContent renders the feed content area with scroll support
func (m Model) renderContent(availableHeight int) string {
	allLines := m.buildAllContentLinesWithPosts()

	// Clamp scroll offset
	maxOffset := len(allLines) - availableHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	offset := m.scrollOffset
	if offset > maxOffset {
		offset = maxOffset
	}
	if offset < 0 {
		offset = 0
	}

	// Extract visible lines
	endIdx := offset + availableHeight
	if endIdx > len(allLines) {
		endIdx = len(allLines)
	}
	visibleLines := allLines[offset:endIdx]

	// Style for background padding
	bgStyle := lipgloss.NewStyle().Background(m.theme.Background)
	// Style for selected post highlight - use secondary background for subtle highlight
	highlightStyle := lipgloss.NewStyle().Background(m.theme.BackgroundSecondary)

	// Build styled lines - each line gets background applied separately
	// to avoid gaps from newline characters
	styledLines := make([]string, availableHeight)
	for i := 0; i < availableHeight; i++ {
		var line string
		var isSelected bool
		if i < len(visibleLines) {
			line = visibleLines[i].text
			isSelected = visibleLines[i].postIndex == m.selectedPostIndex
		}

		// Determine background style based on selection
		var padStyle lipgloss.Style
		if isSelected {
			padStyle = highlightStyle
		} else {
			padStyle = bgStyle
		}

		// Pad to full width with STYLED spaces (not plain spaces)
		// This ensures background is maintained after any inner ANSI resets
		visibleLen := lipgloss.Width(line)
		if visibleLen < m.width {
			// Style the padding separately so it has its own background
			padding := padStyle.Render(strings.Repeat(" ", m.width-visibleLen))
			line += padding
		}

		// Apply selection highlight to the entire line (prepend indicator and wrap)
		if isSelected && len(line) > 0 {
			// Add a subtle indicator and background to the whole line
			line = highlightStyle.Render("▶ ") + line
		} else if len(line) > 0 {
			// Non-selected lines get spacing to align with selected indicator
			line = bgStyle.Render("  ") + line
		}

		styledLines[i] = line
	}

	// Use JoinVertical which handles line joining without extra newlines
	return lipgloss.JoinVertical(lipgloss.Left, styledLines...)
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
			indented[i] = m.styleSpace("  └─ ") + line
		} else {
			indented[i] = m.styleSpace("     ") + line
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

	// Build prefix with styled spaces to avoid black gaps: "HH:MM author: "
	prefix := timeStr + m.styleSpace(" ") + identity + m.styleSpace(": ")
	prefixLen := len(formatTimestamp(post)) + 1 + len(post.Author) + 2

	// Calculate content width for first line
	firstLineWidth := termWidth - prefixLen
	if firstLineWidth < MinContentWidth {
		firstLineWidth = MinContentWidth
	}

	// Wrap text: first line shorter, continuation lines full width
	contentLines := wrapTextFirstLineShorter(post.Content, firstLineWidth, termWidth)

	// Build result lines
	lines := make([]string, 0, len(contentLines))
	for i, line := range contentLines {
		// Apply background to message content (HighlightAll only adds foreground colors)
		highlighted := m.styleSpace(HighlightWithTheme(line, m.theme))
		if i == 0 {
			lines = append(lines, prefix+highlighted)
		} else {
			// Continuation lines at column 0 (no padding)
			lines = append(lines, highlighted)
		}
	}

	return lines
}

// formatPostComfy: Balanced - message starts on same line as identity
// Format: HH:MM  author@project message continues here...
// Continuation lines align with content start
func (m Model) formatPostComfy(post *Post) []string {
	termWidth := m.width
	if termWidth <= 0 {
		termWidth = DefaultTerminalWidth
	}

	timeStr := m.styleTimestamp(formatTimestamp(post))
	identity := m.styleIdentity(post)

	// Build prefix with styled spaces to avoid black gaps: "HH:MM  author "
	prefix := timeStr + m.styleSpace("  ") + identity + m.styleSpace(" ")
	prefixLen := len(formatTimestamp(post)) + 2 + len(post.Author) + 1 + len(post.Suffix) + 1

	// Calculate content width
	contentWidth := termWidth - prefixLen
	if contentWidth < MinContentWidth {
		contentWidth = MinContentWidth
	}

	// Wrap text: all lines same width
	contentLines := wrapText(post.Content, contentWidth)

	// Build result lines with continuation padding
	continuationPadding := strings.Repeat(" ", prefixLen)
	lines := make([]string, 0, len(contentLines))
	for i, line := range contentLines {
		// Apply background to message content (HighlightAll only adds foreground colors)
		highlighted := m.styleSpace(HighlightWithTheme(line, m.theme))
		if i == 0 {
			lines = append(lines, prefix+highlighted)
		} else {
			// Continuation lines aligned with content (styled to avoid black gaps)
			lines = append(lines, m.styleSpace(continuationPadding)+highlighted)
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

	// First line: time and identity (styled spaces to avoid black gaps)
	headerLine := timeStr + m.styleSpace("  ") + identity

	// Content lines: wrap to full width minus small margin
	contentLines := wrapText(post.Content, termWidth-2)

	// Build result: header + content lines
	lines := make([]string, 0, 1+len(contentLines))
	lines = append(lines, headerLine)
	for _, line := range contentLines {
		lines = append(lines, m.styleSpace(HighlightWithTheme(line, m.theme)))
	}

	return lines
}

// styleTimestamp applies theme styling to timestamp
func (m Model) styleTimestamp(s string) string {
	style := lipgloss.NewStyle().
		Foreground(m.theme.TextMuted).
		Background(m.theme.Background)
	return style.Render(s)
}

// styleSpace applies theme background to spacing
func (m Model) styleSpace(s string) string {
	style := lipgloss.NewStyle().Background(m.theme.Background)
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
	helpContent.WriteString("  Navigation\n")
	helpContent.WriteString("   ↑/k    Scroll up\n")
	helpContent.WriteString("   ↓/j    Scroll down\n")
	helpContent.WriteString("   PgUp   Page up\n")
	helpContent.WriteString("   PgDn   Page down\n")
	helpContent.WriteString("   g/G    Top/bottom\n")
	helpContent.WriteString("\n")
	helpContent.WriteString("  Settings\n")
	helpContent.WriteString("   a      Toggle auto-refresh\n")
	helpContent.WriteString("   s      Toggle sort order\n")
	helpContent.WriteString("   l/L    Cycle layout\n")
	helpContent.WriteString("   t/T    Cycle theme\n")
	helpContent.WriteString("   +/-    Adjust pressure\n")
	helpContent.WriteString("   r      Refresh now\n")
	helpContent.WriteString("   q      Quit\n")
	helpContent.WriteString("\n")
	helpContent.WriteString("\n")
	helpContent.WriteString("  Current Settings\n")
	helpContent.WriteString(fmt.Sprintf("   Auto: %s\n", autoStr))
	helpContent.WriteString(fmt.Sprintf("   Sort: %s\n", sortStr))
	helpContent.WriteString(fmt.Sprintf("   Layout: %s\n", layoutName))
	helpContent.WriteString(fmt.Sprintf("   Theme: %s\n", m.theme.DisplayName))
	pressureLevel := config.GetPressureLevel(m.pressure)
	helpContent.WriteString(fmt.Sprintf("   Pressure: %s\n", pressureLevel.Label))
	helpContent.WriteString("\n")
	helpContent.WriteString("      Press any key to close\n")

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

// copyMenuOptions defines the copy format options
var copyMenuOptions = []struct {
	label string
	desc  string
}{
	{"Text", "Copy as formatted text"},
	{"Square", "1200×1200 image for social"},
	{"Landscape", "1200×630 image for Twitter/OG"},
}

// renderCopyMenuOverlay creates a centered copy format menu
func (m Model) renderCopyMenuOverlay() string {
	menuContent := strings.Builder{}
	menuContent.WriteString("\n")
	menuContent.WriteString("      Copy Post\n")
	menuContent.WriteString("\n")

	// Show confirmation message if present
	if m.copyConfirmation != "" {
		menuContent.WriteString(fmt.Sprintf("  ✓ %s\n", m.copyConfirmation))
		menuContent.WriteString("\n")
	}

	// Render menu options
	for i, opt := range copyMenuOptions {
		prefix := "  "
		if i == m.copyMenuIndex {
			prefix = "▶ "
		}
		menuContent.WriteString(fmt.Sprintf("%s%s\n", prefix, opt.label))
		menuContent.WriteString(fmt.Sprintf("    %s\n", opt.desc))
	}

	menuContent.WriteString("\n")
	menuContent.WriteString("  ↑/↓ navigate  Enter select  Esc close\n")

	// Style the menu box
	menuBoxWidth := 36
	menuStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Accent).
		Padding(1, 2).
		Width(menuBoxWidth)

	styledBox := menuStyle.Render(menuContent.String())

	// Center the box on screen
	boxHeight := strings.Count(styledBox, "\n") + 1
	boxWidth := menuBoxWidth + 6 // border + padding
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
