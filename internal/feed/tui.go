package feed

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	xansi "github.com/charmbracelet/x/ansi"
	"github.com/muesli/reflow/truncate"

	"github.com/dreamiurg/smoke/internal/config"
	"github.com/dreamiurg/smoke/internal/logging"
)

// Help overlay dimensions
const (
	helpBoxInnerWidth = 35 // Content width inside the help box
)

// Model is the Bubbletea model for the TUI feed.
type Model struct {
	posts             []*Post
	theme             *Theme
	contrast          *ContrastLevel
	layout            *LayoutStyle
	showHelp          bool
	autoRefresh       bool
	scrollOffset      int  // Number of lines scrolled from top
	initialScrollDone bool // Track if initial scroll position has been set
	width             int
	height            int
	store             *Store
	config            *config.TUIConfig
	pressure          int // Current pressure level (0-4)
	version           string
	nudgeCount        int // Nudges since last mark-read
	unreadAgentCount  int // Unique agents in unread posts
	err               error
	// Unread tracking fields
	lastReadPostID string // Post ID marking read/unread boundary (set at TUI start)
	unreadCount    int    // Count of unread posts (for status bar display)
	lastReadAt     time.Time

	// Cursor selection state
	selectedPostIndex int     // Index of selected post in displayedPosts
	displayedPosts    []*Post // Posts in display order

	// Copy menu state
	showCopyMenu     bool   // Whether copy menu is visible
	copyMenuIndex    int    // Currently highlighted menu option (0-2)
	copyConfirmation string // Confirmation message after copy

	// Delete confirmation state
	deleteArmed  bool
	deletePostID string
	deleteNotice string
}

// tickMsg is sent every 5 seconds for auto-refresh
type tickMsg time.Time

// clockTickMsg is sent every second for clock updates
type clockTickMsg time.Time

// loadPostsMsg is sent when posts are loaded
type loadPostsMsg struct {
	posts      []*Post
	nudgeCount int
	err        error
}

// contentLine tracks a rendered line and its associated post index (-1 for non-post lines)
type contentLine struct {
	text      string
	postIndex int // Index into displayedPosts, -1 non-post, -2 unread separator
}

const unreadSeparatorIndex = -2

type overlayBox struct {
	lines []string
	top   int
	left  int
}

// NewModel creates a new TUI model with the given store, theme, contrast, layout, and version.
func NewModel(store *Store, theme *Theme, contrast *ContrastLevel, layout *LayoutStyle, cfg *config.TUIConfig, version string) Model {
	// Load last read state
	state, err := config.LoadReadState()
	lastReadID := ""
	lastReadAt := time.Time{}
	if err == nil && state != nil {
		lastReadID = state.LastReadPostID
		lastReadAt = state.Updated
	}

	return Model{
		theme:          theme,
		contrast:       contrast,
		layout:         layout,
		autoRefresh:    cfg.AutoRefresh,
		store:          store,
		config:         cfg,
		pressure:       config.GetPressure(),
		version:        version,
		lastReadPostID: lastReadID,
		lastReadAt:     lastReadAt,
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
	nudgeCount := countAgentNudgesSince(m.lastReadAt)
	return loadPostsMsg{posts: posts, nudgeCount: nudgeCount, err: err}
}

// countAgentNudgesSince counts suggest commands from agent sessions in smoke.log after a timestamp.
func countAgentNudgesSince(since time.Time) int {
	logPath, err := config.GetLogPath()
	if err != nil {
		return 0
	}

	f, err := os.Open(logPath)
	if err != nil {
		return 0
	}
	defer func() {
		_ = f.Close()
	}()

	type cmdObj struct {
		Name string `json:"name"`
	}
	type ctxObj struct {
		Env    string `json:"env"`
		Agent  string `json:"agent"`
		Caller string `json:"caller"`
	}
	type entry struct {
		Time string          `json:"time"`
		Msg  string          `json:"msg"`
		Cmd  json.RawMessage `json:"cmd"`
		Ctx  ctxObj          `json:"ctx"`
	}

	count := 0
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		var e entry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			continue
		}
		if !since.IsZero() {
			if e.Time == "" {
				continue
			}
			ts, err := time.Parse(time.RFC3339Nano, e.Time)
			if err != nil || ts.Before(since) {
				continue
			}
		}
		if e.Msg != "command started" && e.Msg != "command invoked" {
			continue
		}

		cmdName := ""
		var cmdNameStr string
		if err := json.Unmarshal(e.Cmd, &cmdNameStr); err == nil {
			cmdName = cmdNameStr
		} else {
			var obj cmdObj
			if err := json.Unmarshal(e.Cmd, &obj); err == nil {
				cmdName = obj.Name
			}
		}
		if cmdName == "suggest" {
			agent := strings.ToLower(e.Ctx.Agent)
			caller := strings.ToLower(e.Ctx.Caller)
			switch caller {
			case "claude", "codex", "gemini":
				count++
				continue
			}
			if agent != "" && agent != "human" {
				count++
				continue
			}
			if e.Ctx.Env == "claude_code" {
				count++
			}
		}
	}
	if scanErr := scanner.Err(); scanErr != nil {
		logging.LogWarn("failed to scan smoke log", "error", scanErr)
	}

	return count
}

// updateUnreadStats updates unread counters and nudges since last read.
func (m *Model) updateUnreadStats(currentNudges int) {
	m.unreadCount = m.countUnread()
	m.unreadAgentCount = m.countUnreadAgents()
	m.nudgeCount = currentNudges
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

		// Handle copy menu key events
		if m.showCopyMenu {
			return m.handleCopyMenuKey(msg)
		}

		// Clear copy confirmation on any key press
		m.copyConfirmation = ""
		m.deleteNotice = ""
		if msg.String() != "d" {
			m.deleteArmed = false
			m.deletePostID = ""
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

		case "up", "k":
			// Move cursor up to previous post
			if m.selectedPostIndex > 0 {
				m.selectedPostIndex--
				m.ensureSelectedVisible()
			}
			return m, nil

		case "down", "j":
			// Move cursor down to next post
			if m.selectedPostIndex < len(m.displayedPosts)-1 {
				m.selectedPostIndex++
				m.ensureSelectedVisible()
			}
			return m, nil

		case "pgup", "ctrl+u":
			// Move selection up by one page (cursor-style)
			m.moveSelectionByPage(-1)
			return m, nil

		case "pgdown", "ctrl+d":
			// Move selection down by one page (cursor-style)
			m.moveSelectionByPage(1)
			return m, nil

		case "home", "g":
			// Jump to top post
			m.moveSelectionToEdge(true)
			return m, nil

		case "end", "G":
			// Jump to bottom post
			m.moveSelectionToEdge(false)
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
			// Open copy menu for selected post
			if len(m.displayedPosts) > 0 && m.selectedPostIndex >= 0 && m.selectedPostIndex < len(m.displayedPosts) {
				m.showCopyMenu = true
				m.copyMenuIndex = 0
			}
			return m, nil

		case "d":
			if len(m.displayedPosts) == 0 || m.selectedPostIndex < 0 || m.selectedPostIndex >= len(m.displayedPosts) {
				m.deleteNotice = "⚠ No post selected"
				return m, nil
			}
			post := m.displayedPosts[m.selectedPostIndex]
			if post == nil {
				m.deleteNotice = "⚠ No post selected"
				return m, nil
			}
			if m.deleteArmed && m.deletePostID == post.ID {
				if err := m.store.DeleteByID(post.ID); err != nil {
					m.deleteNotice = "⚠ Delete failed"
				} else {
					m.deleteNotice = "✓ Deleted post"
					m.deleteArmed = false
					m.deletePostID = ""
					return m, m.loadPostsCmd
				}
				return m, nil
			}
			m.deleteArmed = true
			m.deletePostID = post.ID
			m.deleteNotice = "Press d again to delete"
			return m, nil

		case "+", "=":
			// Increase pressure (clamped)
			if m.pressure < 4 {
				m.pressure++
				m.err = config.SetPressure(m.pressure)
			}
			return m, nil

		case "-":
			// Decrease pressure (clamped)
			if m.pressure > 0 {
				m.pressure--
				m.err = config.SetPressure(m.pressure)
			}
			return m, nil

		case " ", "space":
			// Mark read up to selected post
			if len(m.displayedPosts) > 0 && m.selectedPostIndex >= 0 && m.selectedPostIndex < len(m.displayedPosts) {
				postID := m.displayedPosts[m.selectedPostIndex].ID
				if err := config.SaveLastReadPostID(postID); err == nil {
					m.lastReadPostID = postID
					m.lastReadAt = time.Now()
					m.updateUnreadStats(0)
				} else {
					m.err = err
				}
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
			m.ensureSelectedVisibleWithUnread()
			m.initialScrollDone = true
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
			oldMaxOffset := m.maxScrollOffset()
			wasAtBottom := m.scrollOffset >= oldMaxOffset
			m.posts = msg.posts
			m.updateDisplayedPosts() // Update displayedPosts for cursor navigation
			m.updateUnreadStats(msg.nudgeCount)
			// Set initial selection once we have posts (scroll waits for WindowSizeMsg)
			if !m.initialScrollDone && len(m.posts) > 0 {
				m.initSelectionToUnread()
				if m.height > 0 {
					m.ensureSelectedVisibleWithUnread()
					m.initialScrollDone = true
				}
			} else if len(m.posts) > oldCount && m.height > 0 {
				// Auto-scroll when NEW posts arrive (after initial load) in auto-refresh mode
				if m.autoRefresh && wasAtBottom {
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

	// Render three sections: header, content, status bar
	header := m.renderHeader()
	statusBar := m.renderStatusBar()
	content := m.renderContentBox()

	// Use JoinVertical for seamless background colors
	view := lipgloss.JoinVertical(lipgloss.Left, header, content, statusBar)

	if m.showHelp {
		view = m.applyOverlay(view, m.renderHelpOverlayBox())
	}
	if m.showCopyMenu {
		view = m.applyOverlay(view, m.renderCopyMenuOverlayBox())
	}

	return view
}

func (m Model) applyOverlay(base string, overlay overlayBox) string {
	if len(overlay.lines) == 0 {
		return base
	}

	baseLines := strings.Split(base, "\n")
	if m.height > 0 && len(baseLines) < m.height {
		for len(baseLines) < m.height {
			baseLines = append(baseLines, "")
		}
	}

	for i, line := range overlay.lines {
		target := overlay.top + i
		if target < 0 || target >= len(baseLines) {
			continue
		}
		baseLines[target] = overlayLine(baseLines[target], line, overlay.left, m.width)
	}

	return strings.Join(baseLines, "\n")
}

func overlayLine(base, overlay string, left int, width int) string {
	if left < 0 {
		left = 0
	}

	baseWidth := xansi.StringWidth(base)
	if width <= 0 {
		width = baseWidth
	}
	if baseWidth < width {
		base += strings.Repeat(" ", width-baseWidth)
		baseWidth = width
	}
	if left > baseWidth {
		left = baseWidth
	}

	maxOverlay := baseWidth - left
	if maxOverlay < 0 {
		maxOverlay = 0
	}
	if xansi.StringWidth(overlay) > maxOverlay {
		overlay = xansi.Cut(overlay, 0, maxOverlay)
	}
	overlayWidth := xansi.StringWidth(overlay)

	leftPart := xansi.Cut(base, 0, left)
	rightStart := left + overlayWidth
	rightPart := ""
	if rightStart < baseWidth {
		rightPart = xansi.Cut(base, rightStart, baseWidth)
	}

	return leftPart + overlay + rightPart
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
	base := lipgloss.NewStyle().Background(m.theme.BackgroundSecondary)
	titleStyle := base.Foreground(m.theme.Accent).Bold(true)
	versionStyle := base.Foreground(m.theme.TextMuted)
	statsStyle := base.Foreground(m.theme.Text)
	sepStyle := base.Foreground(m.theme.TextMuted)

	title := titleStyle.Render("SMOKE")
	version := ""
	if m.version != "" {
		version = versionStyle.Render("v" + m.version)
	} else {
		version = versionStyle.Render("vdev")
	}

	statsText := fmt.Sprintf("new %d posts • %d agents • %d nudges",
		m.unreadCount, m.unreadAgentCount, m.nudgeCount)
	stats := statsStyle.Render(statsText)

	leftContent := title + base.Render(" ") + version + base.Render("  ") + stats

	pressure := statsStyle.Render(m.renderPressureIndicator())
	clock := statsStyle.Render("[" + FormatTime(time.Now()) + "]")
	rightContent := pressure + base.Render("  ") + clock

	spacing := m.width - lipgloss.Width(leftContent) - lipgloss.Width(rightContent)
	if spacing < 1 {
		spacing = 1
	}
	gap := sepStyle.Render(strings.Repeat(" ", spacing))

	return leftContent + gap + rightContent
}

// renderStatusBar creates the status bar showing settings and keybindings
func (m Model) renderStatusBar() string {
	base := lipgloss.NewStyle().Background(m.theme.BackgroundSecondary)
	keyStyle := base.Foreground(m.theme.Accent).Bold(true)
	labelStyle := base.Foreground(m.theme.TextMuted)
	valueStyle := base.Foreground(m.theme.Text)
	sep := base.Render("  ")
	width := m.width
	if width <= 0 {
		width = DefaultTerminalWidth
	}

	autoStr := "OFF"
	if m.autoRefresh {
		autoStr = "ON"
	}

	layoutName := "comfy"
	if m.layout != nil {
		layoutName = m.layout.Name
	}

	markValue := "to here"
	if m.unreadCount > 0 {
		markValue = fmt.Sprintf("to here (%d new)", m.unreadCount)
	}

	items := []string{
		keyStyle.Render("Space") + labelStyle.Render(" Read ") + valueStyle.Render(markValue),
		keyStyle.Render("c") + labelStyle.Render(" Copy"),
		keyStyle.Render("r") + labelStyle.Render(" Refresh"),
		keyStyle.Render("a") + labelStyle.Render(" Auto Refresh ") + valueStyle.Render(autoStr),
		keyStyle.Render("l/L") + labelStyle.Render(" Layout ") + valueStyle.Render(layoutName),
		keyStyle.Render("t/T") + labelStyle.Render(" Theme ") + valueStyle.Render(m.theme.Name),
		keyStyle.Render("?") + labelStyle.Render(" Help"),
		keyStyle.Render("q") + labelStyle.Render(" Quit"),
	}

	prefixItems := make([]string, 0, 3)
	if m.copyConfirmation != "" {
		prefixItems = append(prefixItems, valueStyle.Render(m.copyConfirmation))
	}
	if m.deleteNotice != "" {
		prefixItems = append(prefixItems, valueStyle.Render(m.deleteNotice))
	}
	if m.err != nil {
		prefixItems = append(prefixItems, keyStyle.Render("!")+
			labelStyle.Render(" config error"))
	}

	allItems := append([]string{}, prefixItems...)
	allItems = append(allItems, items...)
	statusText := fitStatusLine(allItems, sep, width)
	statusText = clampStatusLine(statusText, width, base)
	return statusText
}

func fitStatusLine(items []string, sep string, width int) string {
	if width <= 0 {
		return strings.Join(items, sep)
	}
	if len(items) == 0 {
		return ""
	}
	joined := strings.Join(items, sep)
	if lipgloss.Width(joined) <= width {
		return joined
	}

	for len(items) > 1 {
		items = items[:len(items)-1]
		joined = strings.Join(items, sep)
		if lipgloss.Width(joined) <= width {
			return joined
		}
	}

	// Single item still too wide; truncate with ellipsis to avoid wrapping.
	return truncate.StringWithTail(items[0], uint(width), "…")
}

func clampStatusLine(text string, width int, base lipgloss.Style) string {
	if width <= 0 {
		return text
	}

	text = strings.ReplaceAll(text, "\n", " ")
	truncated := truncate.StringWithTail(text, uint(width), "")
	visible := lipgloss.Width(truncated)
	if visible < width {
		padding := base.Render(strings.Repeat(" ", width-visible))
		truncated += padding
	}
	return truncated
}

// contentWidth returns the available width inside the content border.
func (m Model) contentWidth() int {
	width := m.width
	if width <= 0 {
		width = DefaultTerminalWidth
	}
	width -= 2 // border
	if width < 1 {
		return 1
	}
	return width
}

// contentHeight returns the available height inside the content border.
func (m Model) contentHeight() int {
	height := m.height - 2 // header + status
	if height <= 2 {
		return 1
	}
	return height - 2 // border
}

// renderContentBox renders the feed content inside a bordered box.
func (m Model) renderContentBox() string {
	availableHeight := m.contentHeight()
	contentWidth := m.contentWidth()
	content := m.renderContent(availableHeight, contentWidth)
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(m.theme.Accent).
		Background(m.theme.Background).
		Width(m.width)
	return boxStyle.Render(content)
}

// buildAllContentLines builds all content lines for the feed (used for scrolling)
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
	termWidth := m.contentWidth()
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
		Foreground(m.theme.DaySeparator).
		Background(m.theme.Background)

	return style.Render(separator)
}

// formatUnreadSeparator creates a styled "NEW" separator line.
// Format: "──── NEW ────" centered with decorative lines
func (m Model) formatUnreadSeparator() string {
	label := "UNREAD"
	termWidth := m.contentWidth()
	if termWidth <= 0 {
		termWidth = DefaultTerminalWidth
	}

	// Build separator: "──── NEW ────"
	minDecor := 4
	labelWithSpace := " " + label + " "
	availableForDecor := termWidth - len(labelWithSpace)

	var leftDecor, rightDecor string
	if availableForDecor >= minDecor*2 {
		decorLen := availableForDecor / 2
		leftDecor = strings.Repeat("─", decorLen)
		rightDecor = strings.Repeat("─", availableForDecor-decorLen)
	} else {
		leftDecor = "──"
		rightDecor = "──"
	}

	separator := leftDecor + labelWithSpace + rightDecor

	// Style with muted text for subtlety
	style := lipgloss.NewStyle().
		Foreground(m.theme.UnreadSeparator).
		Background(m.theme.Background)

	return style.Render(separator)
}

// countUnread counts the number of unread posts based on lastReadPostID.
// Returns 0 if lastReadPostID is empty (first-time user) or not found.
func (m Model) countUnread() int {
	if m.lastReadPostID == "" || len(m.displayedPosts) == 0 {
		return 0
	}

	lastReadIndex := -1
	for i, post := range m.displayedPosts {
		if post.ID == m.lastReadPostID {
			lastReadIndex = i
			break
		}
	}
	if lastReadIndex == -1 {
		return 0
	}
	return len(m.displayedPosts) - lastReadIndex - 1
}

func (m Model) countUnreadAgents() int {
	if m.lastReadPostID == "" || len(m.displayedPosts) == 0 {
		return 0
	}

	lastReadIndex := -1
	for i, post := range m.displayedPosts {
		if post != nil && post.ID == m.lastReadPostID {
			lastReadIndex = i
			break
		}
	}
	if lastReadIndex == -1 {
		return 0
	}

	seen := make(map[string]struct{})
	for i := lastReadIndex + 1; i < len(m.displayedPosts); i++ {
		post := m.displayedPosts[i]
		if post == nil {
			continue
		}
		seen[post.Author] = struct{}{}
	}
	return len(seen)
}

// maxScrollOffset returns the maximum scroll offset based on content size
func (m Model) maxScrollOffset() int {
	allLines := m.buildAllContentLines()
	availableHeight := m.contentHeight()
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
func (m Model) renderContent(availableHeight, availableWidth int) string {
	contentLines := m.buildAllContentLinesWithPosts()
	allLines := make([]string, len(contentLines))
	for i, cl := range contentLines {
		allLines[i] = cl.text
	}

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

	markerLine := -1
	if m.unreadCount > 0 {
		for i, cl := range contentLines {
			if cl.postIndex == unreadSeparatorIndex {
				markerLine = i
				break
			}
		}
		if markerLine == -1 && m.lastReadPostID != "" {
			lastReadIndex := -1
			for i, post := range m.displayedPosts {
				if post.ID == m.lastReadPostID {
					lastReadIndex = i
					break
				}
			}
			if lastReadIndex >= 0 {
				lastLine := -1
				for i, cl := range contentLines {
					if cl.postIndex == lastReadIndex {
						lastLine = i
					}
				}
				if lastLine >= 0 {
					markerLine = lastLine + 1
				}
			}
		}
	}

	unreadAboveCount := 0
	if markerLine >= 0 && markerLine < offset {
		seen := make(map[int]bool)
		start := markerLine + 1
		if start < 0 {
			start = 0
		}
		for i := start; i < offset && i < len(contentLines); i++ {
			if idx := contentLines[i].postIndex; idx >= 0 && !seen[idx] {
				seen[idx] = true
				unreadAboveCount++
			}
		}
	}

	contentHeight := availableHeight
	if contentHeight <= 0 {
		contentHeight = 1
	}

	computeUnreadBelow := func(endIdx int) int {
		if markerLine < 0 || endIdx >= len(contentLines) {
			return 0
		}
		seen := make(map[int]bool)
		start := markerLine + 1
		if start < 0 {
			start = 0
		}
		if start < endIdx {
			start = endIdx
		}
		count := 0
		for i := start; i < len(contentLines); i++ {
			if idx := contentLines[i].postIndex; idx >= 0 && !seen[idx] {
				seen[idx] = true
				count++
			}
		}
		return count
	}

	endIdx := offset + contentHeight
	if endIdx > len(allLines) {
		endIdx = len(allLines)
	}
	unreadBelowCount := computeUnreadBelow(endIdx)
	if unreadBelowCount > 0 && contentHeight > 1 {
		contentHeight--
		endIdx = offset + contentHeight
		if endIdx > len(allLines) {
			endIdx = len(allLines)
		}
		unreadBelowCount = computeUnreadBelow(endIdx)
	}

	// Extract visible lines
	visibleLines := allLines[offset:endIdx]
	if unreadAboveCount > 0 {
		indicator := m.formatUnreadAboveIndicator(unreadAboveCount)
		lines := make([]string, 0, contentHeight)
		lines = append(lines, indicator)
		for i := 0; i < contentHeight-1; i++ {
			if i < len(visibleLines) {
				lines = append(lines, visibleLines[i])
			} else {
				lines = append(lines, "")
			}
		}
		visibleLines = lines
	}
	if unreadBelowCount > 0 && contentHeight > 1 {
		indicator := m.formatUnreadBelowIndicator(unreadBelowCount)
		visibleLines = append(visibleLines, indicator)
	}

	// Style for background padding
	bgStyle := lipgloss.NewStyle().Background(m.theme.Background)

	// Build styled lines - each line gets background applied separately
	// to avoid gaps from newline characters
	styledLines := make([]string, availableHeight)
	for i := 0; i < availableHeight; i++ {
		var line string
		if i < len(visibleLines) {
			line = visibleLines[i]
		}
		// Pad to full width with STYLED spaces (not plain spaces)
		// This ensures background is maintained after any inner ANSI resets
		visibleLen := lipgloss.Width(line)
		if visibleLen < availableWidth {
			// Style the padding separately so it has its own background
			padding := bgStyle.Render(strings.Repeat(" ", availableWidth-visibleLen))
			line += padding
		}
		styledLines[i] = line
	}

	// Use JoinVertical which handles line joining without extra newlines
	return lipgloss.JoinVertical(lipgloss.Left, styledLines...)
}

func (m Model) formatUnreadAboveIndicator(count int) string {
	width := m.contentWidth()
	if width <= 0 {
		width = DefaultTerminalWidth
	}
	label := fmt.Sprintf(" %d UNREAD ABOVE ", count)
	labelWidth := lipgloss.Width(label)
	if labelWidth > width {
		label = truncate.StringWithTail(label, uint(width), "")
		labelWidth = lipgloss.Width(label)
	}

	remaining := width - labelWidth
	left := remaining / 2
	right := remaining - left

	line := strings.Repeat("─", left) + label + strings.Repeat("─", right)
	style := lipgloss.NewStyle().
		Foreground(m.theme.Accent).
		Background(m.theme.BackgroundSecondary).
		Bold(true)
	return style.Render(line)
}

func (m Model) formatUnreadBelowIndicator(count int) string {
	width := m.contentWidth()
	if width <= 0 {
		width = DefaultTerminalWidth
	}
	label := fmt.Sprintf(" %d NEW BELOW ", count)
	labelWidth := lipgloss.Width(label)
	if labelWidth > width {
		label = truncate.StringWithTail(label, uint(width), "")
		labelWidth = lipgloss.Width(label)
	}

	remaining := width - labelWidth
	left := remaining / 2
	right := remaining - left

	line := strings.Repeat("─", left) + label + strings.Repeat("─", right)
	style := lipgloss.NewStyle().
		Foreground(m.theme.Accent).
		Background(m.theme.BackgroundSecondary).
		Bold(true)
	return style.Render(line)
}

// formatPost formats a post according to the current layout
func (m Model) formatPost(post *Post) []string {
	return m.formatPostWithBackground(post, m.theme.Background, false)
}

// formatPostWithBackground formats a post with a custom background.
// When selected is true, timestamp uses accent color for stronger highlight.
func (m Model) formatPostWithBackground(post *Post, background lipgloss.AdaptiveColor, selected bool) []string {
	if m.layout == nil {
		return m.formatPostComfyWithBackground(post, background, selected)
	}
	switch m.layout.Name {
	case "dense":
		return m.formatPostDenseWithBackground(post, background, selected)
	case "relaxed":
		return m.formatPostRelaxedWithBackground(post, background, selected)
	default:
		return m.formatPostComfyWithBackground(post, background, selected)
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
	return m.formatPostDenseWithBackground(post, m.theme.Background, false)
}

func (m Model) formatPostDenseWithBackground(post *Post, background lipgloss.AdaptiveColor, selected bool) []string {
	termWidth := m.contentWidth()
	if termWidth <= 0 {
		termWidth = DefaultTerminalWidth
	}

	timeStr := m.styleTimestampWithBackground(formatTimestamp(post), background, selected)
	identity := m.styleIdentityWithBackground(post, background)

	// Build prefix with styled spaces to avoid black gaps: "HH:MM author: "
	prefix := timeStr + m.styleSpaceWithBackground(" ", background) + identity + m.styleSpaceWithBackground(": ", background)
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
		highlighted := m.styleSpaceWithBackground(HighlightWithThemeAndBackground(line, m.theme, background), background)
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
	return m.formatPostComfyWithBackground(post, m.theme.Background, false)
}

func (m Model) formatPostComfyWithBackground(post *Post, background lipgloss.AdaptiveColor, selected bool) []string {
	termWidth := m.contentWidth()
	if termWidth <= 0 {
		termWidth = DefaultTerminalWidth
	}

	timeStr := m.styleTimestampWithBackground(formatTimestamp(post), background, selected)
	identity := m.styleIdentityWithBackground(post, background)
	callerTag := ResolveCallerTag(post)
	tagLen := 0
	if callerTag != "" {
		tagLen = len(callerTag) + 3 // leading space + brackets
	}

	// Build prefix with styled spaces to avoid black gaps: "HH:MM  author "
	prefix := timeStr + m.styleSpaceWithBackground("  ", background) + identity
	if callerTag != "" {
		prefix += m.styleSpaceWithBackground(" ", background) + m.styleAgentTagWithBackground(callerTag, background)
	}
	prefix += m.styleSpaceWithBackground(" ", background)
	prefixLen := len(formatTimestamp(post)) + 2 + len(post.Author) + 1 + len(post.Suffix) + 1 + tagLen

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
		highlighted := m.styleSpaceWithBackground(HighlightWithThemeAndBackground(line, m.theme, background), background)
		if i == 0 {
			lines = append(lines, prefix+highlighted)
		} else {
			// Continuation lines aligned with content (styled to avoid black gaps)
			lines = append(lines, m.styleSpaceWithBackground(continuationPadding, background)+highlighted)
		}
	}

	return lines
}

// formatPostRelaxed: Most spacious - author on separate line, content below
// Format: HH:MM  author@project
//
//	message on next line...
func (m Model) formatPostRelaxed(post *Post) []string {
	return m.formatPostRelaxedWithBackground(post, m.theme.Background, false)
}

func (m Model) formatPostRelaxedWithBackground(post *Post, background lipgloss.AdaptiveColor, selected bool) []string {
	termWidth := m.contentWidth()
	if termWidth <= 0 {
		termWidth = DefaultTerminalWidth
	}

	timeStr := m.styleTimestampWithBackground(formatTimestamp(post), background, selected)
	identity := m.styleIdentityWithBackground(post, background)
	agentTag := ResolveCallerTag(post)

	// First line: time and identity (styled spaces to avoid black gaps)
	headerLine := timeStr + m.styleSpaceWithBackground("  ", background) + identity
	if agentTag != "" {
		headerLine += m.styleSpaceWithBackground("  ", background) + m.styleAgentTagWithBackground(agentTag, background)
	}

	// Content lines: wrap to full width minus small margin
	contentLines := wrapText(post.Content, termWidth-2)

	// Build result: header + content lines
	lines := make([]string, 0, 1+len(contentLines))
	lines = append(lines, headerLine)
	for _, line := range contentLines {
		lines = append(lines, m.styleSpaceWithBackground(HighlightWithThemeAndBackground(line, m.theme, background), background))
	}

	return lines
}

// styleTimestamp applies theme styling to timestamp
func (m Model) styleTimestamp(s string) string {
	return m.styleTimestampWithBackground(s, m.theme.Background, false)
}

// styleSpace applies theme background to spacing
func (m Model) styleSpace(s string) string {
	return m.styleSpaceWithBackground(s, m.theme.Background)
}

// selectionBackground returns the background color used for selected posts.
func (m Model) selectionBackground() lipgloss.AdaptiveColor {
	return m.theme.BackgroundSecondary
}

// styleTimestampWithBackground applies theme styling to timestamp with custom background.
func (m Model) styleTimestampWithBackground(s string, background lipgloss.AdaptiveColor, selected bool) string {
	foreground := m.theme.TextMuted
	if selected {
		foreground = m.theme.Accent
	}
	style := lipgloss.NewStyle().
		Foreground(foreground).
		Background(background)
	return style.Render(s)
}

// styleSpaceWithBackground applies custom background to spacing.
func (m Model) styleSpaceWithBackground(s string, background lipgloss.AdaptiveColor) string {
	style := lipgloss.NewStyle().Background(background)
	return style.Render(s)
}

// padLineToWidth pads a line to terminal width using a custom background.
func (m Model) padLineToWidth(line string, background lipgloss.AdaptiveColor) string {
	width := m.contentWidth()
	if width <= 0 {
		return line
	}
	visibleLen := lipgloss.Width(line)
	if visibleLen >= width {
		return line
	}
	padding := strings.Repeat(" ", width-visibleLen)
	return line + m.styleSpaceWithBackground(padding, background)
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

// styleIdentityWithBackground formats and styles author@project with custom background.
func (m Model) styleIdentityWithBackground(post *Post, background lipgloss.AdaptiveColor) string {
	return ColorizeIdentityWithBackground(post.Author, m.theme, m.contrast, background)
}

func (m Model) styleAgentTagWithBackground(tag string, background lipgloss.AdaptiveColor) string {
	if tag == "" {
		return ""
	}
	style := lipgloss.NewStyle().
		Foreground(m.theme.TextMuted).
		Background(background)
	return style.Render("[" + tag + "]")
}

// renderHelpOverlayBox creates a centered help overlay box.
func (m Model) renderHelpOverlayBox() overlayBox {
	autoStr := "OFF"
	if m.autoRefresh {
		autoStr = "ON"
	}

	layoutName := "Comfy"
	if m.layout != nil {
		layoutName = m.layout.DisplayName
	}

	type helpRow struct {
		key  string
		desc string
	}

	base := lipgloss.NewStyle().Background(m.theme.BackgroundSecondary)
	keyStyle := base.Foreground(m.theme.Accent).Bold(true)
	descStyle := base.Foreground(m.theme.TextMuted)
	headerStyle := base.Foreground(m.theme.Text).Bold(true)
	titleStyle := base.Foreground(m.theme.Accent).Bold(true)
	dividerStyle := base.Foreground(m.theme.TextMuted)

	padRight := func(s string, width int) string {
		if width <= 0 {
			return s
		}
		space := width - lipgloss.Width(s)
		if space <= 0 {
			return s
		}
		return s + strings.Repeat(" ", space)
	}

	rowWidth := func(block string) int {
		max := 0
		for _, line := range strings.Split(block, "\n") {
			w := lipgloss.Width(line)
			if w > max {
				max = w
			}
		}
		return max
	}

	renderRows := func(rows []helpRow, keyWidth int) string {
		var b strings.Builder
		for _, row := range rows {
			line := keyStyle.Render(padRight(row.key, keyWidth)) + descStyle.Render(" "+row.desc)
			b.WriteString(line)
			b.WriteString("\n")
		}
		return b.String()
	}

	renderSection := func(title string, rows []helpRow, keyWidth int) string {
		var b strings.Builder
		b.WriteString(headerStyle.Render(title))
		b.WriteString("\n")
		b.WriteString(dividerStyle.Render(strings.Repeat("─", lipgloss.Width(title))))
		b.WriteString("\n")
		b.WriteString(renderRows(rows, keyWidth))
		return b.String()
	}

	leftRows := []helpRow{
		{key: "↑/k", desc: "Select previous post"},
		{key: "↓/j", desc: "Select next post"},
		{key: "PgUp", desc: "Select previous page"},
		{key: "PgDn", desc: "Select next page"},
		{key: "Home", desc: "Top post"},
		{key: "End", desc: "Bottom post"},
		{key: "g/G", desc: "Top/bottom post"},
	}
	shareRows := []helpRow{
		{key: "c", desc: "Copy selected post"},
	}
	readRows := []helpRow{
		{key: "Space", desc: "Mark read to here"},
		{key: "d d", desc: "Delete selected post"},
		{key: "q", desc: "Quit"},
	}
	settingsRows := []helpRow{
		{key: "a", desc: "Toggle auto-refresh"},
		{key: "l/L", desc: "Cycle layout"},
		{key: "t/T", desc: "Cycle theme"},
		{key: "+/-", desc: "Adjust pressure"},
		{key: "r", desc: "Refresh now"},
	}

	pressureLevel := config.GetPressureLevel(m.pressure)
	currentRows := []helpRow{
		{key: "Auto:", desc: autoStr},
		{key: "Layout:", desc: layoutName},
		{key: "Theme:", desc: m.theme.DisplayName},
		{key: "Pressure:", desc: pressureLevel.Label},
	}

	leftKeyWidth := 5
	rightKeyWidth := 7

	leftColumn := strings.Builder{}
	leftColumn.WriteString(renderSection("NAVIGATION", leftRows, leftKeyWidth))
	leftColumn.WriteString("\n")
	leftColumn.WriteString(renderSection("SHARE", shareRows, leftKeyWidth))
	leftColumn.WriteString("\n")
	leftColumn.WriteString(renderSection("READ STATUS", readRows, leftKeyWidth))

	rightColumn := strings.Builder{}
	rightColumn.WriteString(renderSection("SETTINGS", settingsRows, rightKeyWidth))
	rightColumn.WriteString("\n")
	rightColumn.WriteString(renderSection("CURRENT SETTINGS", currentRows, rightKeyWidth))

	leftBlock := leftColumn.String()
	rightBlock := rightColumn.String()
	leftWidth := rowWidth(leftBlock)
	rightWidth := rowWidth(rightBlock)
	leftBlock = m.fillBackgroundBlock(leftBlock, leftWidth, m.theme.BackgroundSecondary)
	rightBlock = m.fillBackgroundBlock(rightBlock, rightWidth, m.theme.BackgroundSecondary)
	gap := base.Render(strings.Repeat(" ", 4))

	columns := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftBlock,
		gap,
		rightBlock,
	)

	title := titleStyle.Render("Smoke Feed Help")
	divider := dividerStyle.Render(strings.Repeat("─", lipgloss.Width("Smoke Feed Help")))

	helpContent := strings.Builder{}
	helpContent.WriteString(title)
	helpContent.WriteString("\n")
	helpContent.WriteString(divider)
	helpContent.WriteString("\n")
	helpContent.WriteString(columns)
	helpContent.WriteString("\n")
	helpContent.WriteString(descStyle.Render("Press any key to close"))
	helpContent.WriteString("\n")

	helpWidth := helpBoxInnerWidth
	if m.width > 0 {
		maxWidth := m.width - 8
		if maxWidth > helpWidth {
			helpWidth = maxWidth
		}
	}

	helpStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Accent).
		Background(m.theme.BackgroundSecondary).
		Padding(0, 2).
		Width(helpWidth)

	contentWithBackground := m.fillBackgroundBlock(helpContent.String(), helpWidth, m.theme.BackgroundSecondary)
	styledBox := helpStyle.Render(contentWithBackground)

	boxHeight := strings.Count(styledBox, "\n") + 1
	boxWidth := lipgloss.Width(styledBox)
	topPadding := (m.height - boxHeight) / 2
	leftPadding := (m.width - boxWidth) / 2

	if leftPadding < 0 {
		leftPadding = 0
	}
	if topPadding < 0 {
		topPadding = 0
	}

	return overlayBox{
		lines: strings.Split(styledBox, "\n"),
		top:   topPadding,
		left:  leftPadding,
	}
}

// renderHelpOverlay returns a string-rendered overlay (used in tests).
func (m Model) renderHelpOverlay() string {
	return m.renderOverlayBoxString(m.renderHelpOverlayBox())
}

func (m Model) renderOverlayBoxString(box overlayBox) string {
	var result strings.Builder
	for i := 0; i < box.top; i++ {
		result.WriteString("\n")
	}
	for _, line := range box.lines {
		if box.left > 0 {
			result.WriteString(strings.Repeat(" ", box.left))
		}
		result.WriteString(line)
		result.WriteString("\n")
	}
	return result.String()
}

// fillBackgroundBlock pads each line to width using background-colored spaces.
// This avoids black gaps when ANSI styles reset within a line.
func (m Model) fillBackgroundBlock(content string, width int, background lipgloss.AdaptiveColor) string {
	if width <= 0 {
		return content
	}
	bgStyle := lipgloss.NewStyle().Background(background)
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		visible := lipgloss.Width(line)
		if visible < width {
			line += bgStyle.Render(strings.Repeat(" ", width-visible))
		}
		lines[i] = line
	}
	return strings.Join(lines, "\n")
}

// updateDisplayedPosts updates the displayedPosts slice in display order.
// This is called when posts are loaded.
func (m *Model) updateDisplayedPosts() {
	if len(m.posts) == 0 {
		m.displayedPosts = nil
		m.selectedPostIndex = 0
		return
	}

	// Build threads and flatten to display order
	threads := buildThreads(m.posts)
	for i, j := 0, len(threads)-1; i < j; i, j = i+1, j-1 {
		threads[i], threads[j] = threads[j], threads[i]
	}

	// Flatten threads to posts in display order (main posts only, not replies)
	m.displayedPosts = make([]*Post, 0, len(threads))
	for _, thread := range threads {
		m.displayedPosts = append(m.displayedPosts, thread.post)
	}

	// Clamp selection index
	if m.selectedPostIndex >= len(m.displayedPosts) {
		m.selectedPostIndex = len(m.displayedPosts) - 1
	}
	if m.selectedPostIndex < 0 {
		m.selectedPostIndex = 0
	}
}

// initSelectionToUnread moves selection to the first unread post if available.
// Falls back to latest post when no unread marker exists.
func (m *Model) initSelectionToUnread() {
	if len(m.displayedPosts) == 0 {
		m.selectedPostIndex = 0
		return
	}

	if m.lastReadPostID == "" {
		m.selectedPostIndex = len(m.displayedPosts) - 1
		return
	}

	lastReadIndex := -1
	for i, post := range m.displayedPosts {
		if post.ID == m.lastReadPostID {
			lastReadIndex = i
			break
		}
	}

	if lastReadIndex == -1 {
		m.selectedPostIndex = len(m.displayedPosts) - 1
		return
	}

	if lastReadIndex < len(m.displayedPosts)-1 {
		// Select first unread post (right after marker)
		m.selectedPostIndex = lastReadIndex + 1
		return
	}

	// No unread posts
	m.selectedPostIndex = lastReadIndex
}

// ensureSelectedVisible adjusts scroll offset to keep the selected post visible.
func (m *Model) ensureSelectedVisible() {
	if len(m.displayedPosts) == 0 || m.contentHeight() <= 0 {
		return
	}

	// Build content lines with post tracking
	contentLines := m.buildAllContentLinesWithPosts()

	// Find the line range for the selected post
	var firstLine, lastLine int
	found := false
	for i, cl := range contentLines {
		if cl.postIndex == m.selectedPostIndex {
			if !found {
				firstLine = i
				found = true
			}
			lastLine = i
		}
	}

	if !found {
		return
	}

	availableHeight := m.contentHeight()

	// Scroll to keep selected post visible
	if firstLine < m.scrollOffset {
		m.scrollOffset = firstLine
	} else if lastLine >= m.scrollOffset+availableHeight {
		m.scrollOffset = lastLine - availableHeight + 1
	}

	// Clamp scroll offset
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

// ensureSelectedVisibleWithUnread keeps the selected post visible and tries to include
// the unread separator when selecting the first unread post.
func (m *Model) ensureSelectedVisibleWithUnread() {
	if len(m.displayedPosts) == 0 || m.contentHeight() <= 0 {
		return
	}

	contentLines := m.buildAllContentLinesWithPosts()
	var firstLine, lastLine int
	found := false
	unreadLine := -1
	for i, cl := range contentLines {
		if cl.postIndex == unreadSeparatorIndex && unreadLine == -1 {
			unreadLine = i
		}
		if cl.postIndex == m.selectedPostIndex {
			if !found {
				firstLine = i
				found = true
			}
			lastLine = i
		}
	}
	if !found {
		return
	}

	availableHeight := m.contentHeight()
	minLine := firstLine
	if unreadLine != -1 && unreadLine < firstLine {
		minLine = unreadLine
	}

	if minLine < m.scrollOffset {
		m.scrollOffset = minLine
	}
	if lastLine >= m.scrollOffset+availableHeight {
		m.scrollOffset = lastLine - availableHeight + 1
	}

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

// moveSelectionByPage moves the selection by one page up/down.
// direction: -1 for up, +1 for down.
func (m *Model) moveSelectionByPage(direction int) {
	if len(m.displayedPosts) == 0 {
		return
	}
	pageSize := m.contentHeight()
	if pageSize < 1 {
		pageSize = 1
	}
	m.moveSelectionByLines(direction * pageSize)
}

// moveSelectionByLines moves selection by a line delta and keeps it visible.
func (m *Model) moveSelectionByLines(delta int) {
	if len(m.displayedPosts) == 0 {
		return
	}

	contentLines := m.buildAllContentLinesWithPosts()
	if len(contentLines) == 0 {
		return
	}

	currentLine := -1
	for i, cl := range contentLines {
		if cl.postIndex == m.selectedPostIndex {
			currentLine = i
			break
		}
	}
	if currentLine == -1 {
		return
	}

	targetLine := currentLine + delta
	if targetLine < 0 {
		targetLine = 0
	}
	if targetLine >= len(contentLines) {
		targetLine = len(contentLines) - 1
	}

	newIndex := m.selectedPostIndex
	if delta < 0 {
		for i := targetLine; i >= 0; i-- {
			if contentLines[i].postIndex >= 0 {
				newIndex = contentLines[i].postIndex
				break
			}
		}
	} else {
		for i := targetLine; i < len(contentLines); i++ {
			if contentLines[i].postIndex >= 0 {
				newIndex = contentLines[i].postIndex
				break
			}
		}
	}

	m.selectedPostIndex = newIndex
	m.ensureSelectedVisible()
}

// moveSelectionToEdge jumps selection to the top or bottom post.
func (m *Model) moveSelectionToEdge(top bool) {
	if len(m.displayedPosts) == 0 {
		return
	}
	if top {
		m.selectedPostIndex = 0
		m.scrollOffset = 0
	} else {
		m.selectedPostIndex = len(m.displayedPosts) - 1
		m.scrollOffset = m.maxScrollOffset()
	}
	m.ensureSelectedVisible()
}

// buildAllContentLinesWithPosts builds content lines with post index tracking.
func (m Model) buildAllContentLinesWithPosts() []contentLine {
	if len(m.posts) == 0 {
		return []contentLine{{text: "No posts yet. Exit TUI (q) and try: smoke post \"hello world\"", postIndex: -1}}
	}

	threads := buildThreads(m.posts)
	for i, j := 0, len(threads)-1; i < j; i, j = i+1, j-1 {
		threads[i], threads[j] = threads[j], threads[i]
	}

	var lines []contentLine
	var lastDay time.Time
	var separatorInserted bool
	var pendingUnreadSeparator bool
	postIndex := 0
	hasUnread := false
	if m.lastReadPostID != "" && len(threads) > 0 {
		lastThreadID := threads[len(threads)-1].post.ID
		hasUnread = m.lastReadPostID != lastThreadID
	}

	for i, thread := range threads {
		// Get post time for day separator
		postTime, err := thread.post.GetCreatedTime()
		if err == nil {
			localTime := postTime.Local()
			postDay := time.Date(localTime.Year(), localTime.Month(), localTime.Day(), 0, 0, 0, 0, localTime.Location())
			if lastDay.IsZero() || !postDay.Equal(lastDay) {
				if i > 0 {
					lines = append(lines, contentLine{text: "", postIndex: -1})
				}
				lines = append(lines, contentLine{text: m.formatDaySeparator(localTime), postIndex: -1})
				lastDay = postDay
			}
		}

		// Format main post with selection indicator
		isSelected := postIndex == m.selectedPostIndex
		postLines := m.formatPostWithSelection(thread.post, isSelected)
		for _, line := range postLines {
			lines = append(lines, contentLine{text: line, postIndex: postIndex})
		}

		// Format replies (not selectable, use -1)
		for _, reply := range thread.replies {
			replyLines := m.formatReplyWithSelection(reply, false)
			for _, line := range replyLines {
				lines = append(lines, contentLine{text: line, postIndex: -1})
			}
		}

		// Insert separator AFTER the last-read thread when there are unread posts
		if hasUnread && !separatorInserted && m.lastReadPostID != "" {
			if thread.post.ID == m.lastReadPostID {
				pendingUnreadSeparator = true
			}
		}

		// Blank line between threads
		if i < len(threads)-1 {
			if pendingUnreadSeparator && !separatorInserted {
				lines = append(lines, contentLine{text: m.formatUnreadSeparator(), postIndex: unreadSeparatorIndex})
				separatorInserted = true
				pendingUnreadSeparator = false
			} else {
				lines = append(lines, contentLine{text: "", postIndex: -1})
			}
		}

		postIndex++
	}

	return lines
}

// formatPostWithSelection formats a post with optional selection indicator.
func (m Model) formatPostWithSelection(post *Post, isSelected bool) []string {
	if !isSelected {
		return m.formatPost(post)
	}

	lines := m.formatPostWithBackground(post, m.selectionBackground(), true)
	for i, line := range lines {
		lines[i] = m.padLineToWidth(line, m.selectionBackground())
	}
	return lines
}

// formatReplyWithSelection formats a reply with optional selection indicator.
func (m Model) formatReplyWithSelection(reply *Post, isSelected bool) []string {
	// Replies are not selectable in the current UI.
	return m.formatReply(reply)
}

// handleCopyMenuKey handles key events when the copy menu is visible.
func (m Model) handleCopyMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.showCopyMenu = false
		return m, nil

	case "up", "k":
		if m.copyMenuIndex > 0 {
			m.copyMenuIndex--
		}
		return m, nil

	case "down", "j":
		if m.copyMenuIndex < 2 {
			m.copyMenuIndex++
		}
		return m, nil

	case "enter", " ":
		m.showCopyMenu = false
		m.executeCopyAction()
		return m, nil

	case "1":
		m.showCopyMenu = false
		m.copyMenuIndex = 0
		m.executeCopyAction()
		return m, nil

	case "2":
		m.showCopyMenu = false
		m.copyMenuIndex = 1
		m.executeCopyAction()
		return m, nil

	case "3":
		m.showCopyMenu = false
		m.copyMenuIndex = 2
		m.executeCopyAction()
		return m, nil
	}

	return m, nil
}

// executeCopyAction performs the copy operation based on copyMenuIndex.
func (m *Model) executeCopyAction() {
	if m.selectedPostIndex < 0 || m.selectedPostIndex >= len(m.displayedPosts) {
		m.copyConfirmation = "⚠ No post selected"
		return
	}

	post := m.displayedPosts[m.selectedPostIndex]

	switch m.copyMenuIndex {
	case 0: // Text
		text := FormatPostAsText(post)
		if err := CopyTextToClipboard(text); err != nil {
			m.copyConfirmation = "⚠ Copy failed"
		} else {
			m.copyConfirmation = "✓ Copied text"
		}

	case 1: // Square image
		data, err := RenderShareCard(post, m.theme, SquareImage)
		if err != nil {
			m.copyConfirmation = "⚠ Render failed"
			return
		}
		if err := CopyImageToClipboard(data); err != nil {
			m.copyConfirmation = "⚠ Copy failed"
		} else {
			m.copyConfirmation = "✓ Copied square image"
		}

	case 2: // Landscape image
		data, err := RenderShareCard(post, m.theme, LandscapeImage)
		if err != nil {
			m.copyConfirmation = "⚠ Render failed"
			return
		}
		if err := CopyImageToClipboard(data); err != nil {
			m.copyConfirmation = "⚠ Copy failed"
		} else {
			m.copyConfirmation = "✓ Copied landscape image"
		}
	}
}

// renderCopyMenuOverlayBox renders the copy menu as a centered overlay box.
func (m Model) renderCopyMenuOverlayBox() overlayBox {
	menuItems := []string{
		"1. Text",
		"2. Square (1200×1200)",
		"3. Landscape (1200×630)",
	}

	base := lipgloss.NewStyle().Background(m.theme.BackgroundSecondary)
	titleStyle := base.Foreground(m.theme.Accent).Bold(true)
	itemStyle := base.Foreground(m.theme.Text)
	selectedStyle := base.Foreground(m.theme.Background).Background(m.theme.Accent).Bold(true)
	hintStyle := base.Foreground(m.theme.TextMuted)

	menuWidth := 32

	var menuContent strings.Builder
	menuContent.WriteString(titleStyle.Width(menuWidth).Align(lipgloss.Center).Render("Copy Post"))
	menuContent.WriteString("\n\n")

	for i, item := range menuItems {
		if i == m.copyMenuIndex {
			menuContent.WriteString(selectedStyle.Width(menuWidth).Render("  " + item))
		} else {
			menuContent.WriteString(itemStyle.Width(menuWidth).Render("  " + item))
		}
		menuContent.WriteString("\n")
	}

	menuContent.WriteString("\n")
	menuContent.WriteString(hintStyle.Width(menuWidth).Render("  ↑/↓ navigate · Enter select"))
	menuContent.WriteString("\n")
	menuContent.WriteString(hintStyle.Width(menuWidth).Render("  Esc/q to cancel"))
	menuContent.WriteString("\n")

	menuStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Accent).
		Background(m.theme.BackgroundSecondary).
		Padding(1, 2).
		Width(menuWidth)

	contentWithBackground := m.fillBackgroundBlock(menuContent.String(), menuWidth, m.theme.BackgroundSecondary)
	styledBox := menuStyle.Render(contentWithBackground)

	boxHeight := strings.Count(styledBox, "\n") + 1
	boxWidth := lipgloss.Width(styledBox)
	topPadding := (m.height - boxHeight) / 2
	leftPadding := (m.width - boxWidth) / 2

	if leftPadding < 0 {
		leftPadding = 0
	}
	if topPadding < 0 {
		topPadding = 0
	}

	return overlayBox{
		lines: strings.Split(styledBox, "\n"),
		top:   topPadding,
		left:  leftPadding,
	}
}

// renderCopyMenuOverlay returns a string-rendered overlay (used in tests).
func (m Model) renderCopyMenuOverlay() string {
	return m.renderOverlayBoxString(m.renderCopyMenuOverlayBox())
}
