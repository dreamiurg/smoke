# Research: TUI Header and Status Bar Redesign

**Date**: 2026-01-31
**Feature**: 005-tui-redesign

## Theme Selection Research

### Decision: 8 Themes Selected
- Dracula, GitHub, Catppuccin, Solarized, Nord, Gruvbox, One Dark, Tokyo Night

### Rationale
- Mix of high-contrast (Dracula, GitHub, One Dark, Gruvbox) and pastel (Catppuccin, Nord, Tokyo Night, Solarized)
- All are well-documented with established color palettes
- Cover both warm (Gruvbox) and cool (Nord) preferences
- Include both legacy favorites (Solarized) and modern trends (Catppuccin, Tokyo Night)

### Alternatives Considered
- Full Abacus theme set (22 themes) - rejected as too many to maintain
- Minimal set (4 themes) - rejected as insufficient variety

## Theme Interface Research

### Decision: Simplified 6-Color Interface
```go
type Theme interface {
    Text() lipgloss.AdaptiveColor           // Primary text
    TextMuted() lipgloss.AdaptiveColor      // Timestamps, secondary
    BackgroundSecondary() lipgloss.AdaptiveColor // Header/status bars
    Accent() lipgloss.AdaptiveColor         // Version badge, highlights
    Error() lipgloss.AdaptiveColor          // Error indicators
    AgentColors() []lipgloss.Color          // 5 colors for agent names
    Name() string                           // Theme identifier
    DisplayName() string                    // Human-readable name
}
```

### Rationale
- Smoke's TUI is simpler than Abacus - doesn't need 16 semantic colors
- Covers all actual use cases: text, backgrounds, accents, errors, agent colors
- AdaptiveColor provides automatic light/dark terminal adaptation
- Fewer colors = fewer values to maintain per theme

### Alternatives Considered
- Full Abacus interface (16 colors) - rejected as over-engineered
- Even simpler (3 colors) - rejected as insufficient for proper theming

## Locale Time Formatting Research

### Decision: Use Go's time.Local with time.Kitchen or custom format
```go
time.Now().Local().Format("15:04")  // 24-hour default
// Or detect locale preference via environment
```

### Rationale
- Go standard library handles timezone automatically via time.Local
- No external dependencies needed
- System locale typically determines 12/24 hour preference

### Alternatives Considered
- External locale library - rejected per constitution (minimal dependencies)
- Fixed 24-hour format - acceptable fallback

## Layout Implementation Research

### Decision: Three-section vertical layout with lipgloss
```go
func (m Model) View() string {
    header := m.renderHeader()      // Fixed at top
    content := m.renderContent()    // Scrollable middle
    status := m.renderStatus()      // Fixed at bottom

    // Calculate content height = terminal height - 2 (header + status)
    contentHeight := m.height - 2

    return lipgloss.JoinVertical(lipgloss.Left, header, content, status)
}
```

### Rationale
- lipgloss.JoinVertical handles layout composition
- Content area receives remaining height after header/status
- Scrolling handled by limiting content lines to available height

### Alternatives Considered
- Viewport component for scrolling - may add later if needed
- Manual string concatenation - less maintainable

## Auto-Refresh State Persistence Research

### Decision: Add to existing TUIConfig
```go
type TUIConfig struct {
    Theme      string `json:"theme"`
    Contrast   string `json:"contrast"`
    Style      string `json:"style"`
    AutoRefresh bool  `json:"auto_refresh"` // NEW
}
```

### Rationale
- Follows existing pattern in tui.go
- Minimal change to existing config structure
- JSON serialization already handled

### Alternatives Considered
- Separate config file - rejected as unnecessary complexity
- Environment variable - rejected as not persistent across sessions
