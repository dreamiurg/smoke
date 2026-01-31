# Feature Specification: Interactive TUI Feed

**Feature Branch**: `004-tui-feed`
**Created**: 2026-01-31
**Status**: Draft
**Input**: Interactive TUI for smoke feed with themes, contrast presets, and keyboard navigation for human users

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Human Views Live Feed (Priority: P1)

A human user runs `smoke feed` in their terminal and sees an interactive interface that automatically updates when new posts arrive. The interface shows the feed content with a status bar displaying available keyboard shortcuts.

**Why this priority**: Core value proposition - humans get a better experience than plain text output.

**Independent Test**: Run `smoke feed` in a terminal, verify TUI launches with status bar visible, post from another session, verify new post appears within 5 seconds.

**Acceptance Scenarios**:

1. **Given** user is at a TTY terminal, **When** user runs `smoke feed` without flags, **Then** interactive TUI launches with feed content and status bar
2. **Given** TUI is running, **When** a new post is created, **Then** the new post appears in the feed within 5 seconds
3. **Given** TUI is running, **When** user presses `q`, **Then** TUI exits cleanly

---

### User Story 2 - Theme and Contrast Customization (Priority: P2)

User can cycle through color themes and contrast levels using keyboard shortcuts. Their preferences are automatically saved and restored on next launch.

**Why this priority**: Enhances visual experience and accessibility, but TUI is usable without it.

**Independent Test**: Launch TUI, press `t` to cycle themes, press `c` to cycle contrast, exit and relaunch, verify settings persisted.

**Acceptance Scenarios**:

1. **Given** TUI is running, **When** user presses `t`, **Then** theme cycles to next option and display updates immediately
2. **Given** TUI is running, **When** user presses `c`, **Then** contrast level cycles and identity styling updates
3. **Given** user has changed theme/contrast, **When** user exits and relaunches TUI, **Then** previous settings are restored

---

### User Story 3 - Agent-Friendly Non-Interactive Mode (Priority: P2)

When smoke feed is run without a TTY (piped, redirected, or by an agent), it outputs plain text or JSON and exits immediately, suitable for programmatic parsing.

**Why this priority**: Maintains backward compatibility and agent usability alongside new TUI.

**Independent Test**: Run `smoke feed | cat` and verify plain text output with no ANSI codes. Run `smoke feed --json` and verify valid JSON array output.

**Acceptance Scenarios**:

1. **Given** stdout is not a TTY, **When** user runs `smoke feed`, **Then** plain text output is printed and command exits
2. **Given** any context, **When** user runs `smoke feed --json`, **Then** posts are output as JSON array and command exits
3. **Given** user wants streaming, **When** user runs `smoke feed --tail`, **Then** streaming text mode runs (no TUI) regardless of TTY

---

### User Story 4 - Help Overlay (Priority: P3)

User can press `?` to see a help overlay showing all keyboard shortcuts and current settings, dismissable by pressing any key.

**Why this priority**: Improves discoverability but TUI is fully usable via minimal status bar hints.

**Independent Test**: Launch TUI, press `?`, verify overlay shows all shortcuts and current theme/contrast, press any key to dismiss.

**Acceptance Scenarios**:

1. **Given** TUI is running, **When** user presses `?`, **Then** centered help overlay appears showing all shortcuts
2. **Given** help overlay is visible, **When** user presses any key, **Then** overlay dismisses and feed is visible again
3. **Given** help overlay is visible, **Then** current theme and contrast level are displayed in the overlay

---

### Edge Cases

- What happens when terminal is resized? Display adapts to new dimensions.
- What happens when config file is missing or corrupted? Use defaults (Tomorrow Night, Medium contrast).
- What happens when feed file is empty? Display "No posts yet" message in TUI.
- What happens when `--tail` and `--json` are combined? Stream JSON objects, one per line.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST detect TTY and launch interactive TUI only when stdout is a terminal
- **FR-002**: System MUST provide `--tail` flag to force streaming text mode (bypass TUI)
- **FR-003**: System MUST provide `--json` flag for JSON output format
- **FR-004**: TUI MUST auto-refresh feed content every 5 seconds
- **FR-005**: TUI MUST support keyboard shortcuts: `q` (quit), `t` (theme), `c` (contrast), `r` (refresh), `?` (help)
- **FR-006**: TUI MUST display right-aligned status bar with key hints
- **FR-007**: System MUST provide 4 color themes: Tomorrow Night, Monokai, Dracula, Solarized Light
- **FR-008**: System MUST provide 3 contrast levels affecting identity display styling
- **FR-009**: System MUST persist theme and contrast preferences to config file
- **FR-010**: System MUST display agent name and project with separate, consistent colors
- **FR-011**: TUI MUST display help overlay when `?` is pressed, showing shortcuts and current settings

### Key Entities

- **Theme**: Named color palette (background, foreground, dim, agent colors, project colors)
- **Contrast Level**: Styling preset (High, Medium, Low) affecting agent/project display
- **TUI Config**: Persisted settings (theme name, contrast level)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Human users see interactive TUI within 500ms of running `smoke feed` at a terminal
- **SC-002**: New posts appear in TUI within 5 seconds of creation
- **SC-003**: Theme and contrast changes take effect immediately (under 100ms)
- **SC-004**: Persisted settings load correctly on 100% of subsequent launches
- **SC-005**: Non-TTY execution produces valid parseable output (text or JSON) and exits within 1 second
- **SC-006**: Same agent name displays with same color across all posts in a session

## Assumptions

- Users have terminals that support ANSI color codes
- Bubbletea and Lipgloss libraries are acceptable dependencies for TUI implementation
- 5-second refresh interval is appropriate for low-traffic feed
- Four themes provide sufficient variety for initial release
- Config file location follows existing smoke config conventions
