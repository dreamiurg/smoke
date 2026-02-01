# Feature Specification: TUI Header and Status Bar Redesign

**Feature Branch**: `005-tui-redesign`
**Created**: 2026-01-31
**Status**: Draft
**Input**: Redesign smoke feed TUI with Abacus-style layout featuring fixed header/status bars, theme system expansion, and auto-refresh toggle.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - View Feed with Context Information (Priority: P1)

As a user viewing the smoke feed, I want to see contextual information about the feed (version, post count, unique agents, unique projects, current time) in a persistent header bar so I can understand the feed's scope at a glance.

**Why this priority**: The header provides essential context that helps users understand what they're viewing. Without this information, users lack awareness of feed activity and scope.

**Independent Test**: Can be fully tested by launching the TUI and verifying the header displays version, statistics, and clock. Delivers immediate value by providing feed context.

**Acceptance Scenarios**:

1. **Given** the TUI is running, **When** I view the feed, **Then** I see a header bar at the top showing "SMOKE v[version]", post count, unique agent count, unique project count, and current time
2. **Given** the TUI is running with posts from 5 different agents across 3 projects, **When** I view the header, **Then** I see "Agents: 5" and "Projects: 3"
3. **Given** the TUI is running, **When** time passes, **Then** the clock in the header updates to reflect current system time in locale format

---

### User Story 2 - View and Understand Current Settings (Priority: P1)

As a user, I want to see my current TUI settings (auto-refresh state, style, contrast) and theme in a persistent status bar so I know what configuration is active and what keys control each setting.

**Why this priority**: The status bar is essential for discoverability - users need to know what settings are active and how to change them. This is foundational for the entire TUI experience.

**Independent Test**: Can be fully tested by launching the TUI and verifying the status bar displays all settings with their keybindings. Delivers immediate value by showing current state.

**Acceptance Scenarios**:

1. **Given** the TUI is running, **When** I view the feed, **Then** I see a status bar at the bottom showing "(a) auto: [ON/OFF]  (s) style: [current]  (t) theme: [current]  (c) contrast: [current]  (?) help  (q) quit"
2. **Given** auto-refresh is enabled, **When** I view the status bar, **Then** I see "(a) auto: ON"
3. **Given** the theme is set to "dracula", **When** I view the status bar, **Then** I see "(t) theme: dracula"

---

### User Story 3 - Toggle Auto-Refresh (Priority: P2)

As a user, I want to toggle automatic feed refresh on/off using the (a) key so I can control whether the feed updates automatically or only when I manually refresh.

**Why this priority**: Auto-refresh toggle gives users control over feed behavior, which is important for both convenience (automatic updates) and focus (stopping distracting updates).

**Independent Test**: Can be fully tested by pressing (a) and verifying the status bar updates and refresh behavior changes accordingly.

**Acceptance Scenarios**:

1. **Given** auto-refresh is ON, **When** I press (a), **Then** auto-refresh turns OFF and the status bar shows "(a) auto: OFF"
2. **Given** auto-refresh is OFF, **When** I press (a), **Then** auto-refresh turns ON and the status bar shows "(a) auto: ON"
3. **Given** auto-refresh is ON, **When** 5 seconds pass, **Then** the feed refreshes automatically
4. **Given** auto-refresh is OFF, **When** 5 seconds pass, **Then** the feed does NOT refresh automatically

---

### User Story 4 - Switch Between Themes (Priority: P2)

As a user, I want to cycle through available themes using the (t) key so I can choose a color scheme that suits my preferences and terminal environment.

**Why this priority**: Theme support allows users to customize their experience and ensures readability across different terminal backgrounds (light/dark).

**Independent Test**: Can be fully tested by pressing (t) repeatedly and verifying theme changes are reflected in both the UI appearance and status bar.

**Acceptance Scenarios**:

1. **Given** the current theme is "dracula", **When** I press (t), **Then** the theme changes to the next theme in the cycle and the status bar updates
2. **Given** I cycle through all themes, **When** I reach the last theme, **Then** pressing (t) returns to the first theme
3. **Given** I change themes, **When** I restart the TUI, **Then** my theme preference is preserved

---

### User Story 5 - Readable Status Bars Across All Themes (Priority: P2)

As a user, I want the header and status bars to be readable regardless of which theme I select so I can always see the information clearly.

**Why this priority**: Contrast issues make the TUI unusable. Every theme must provide readable header/status bars.

**Independent Test**: Can be fully tested by cycling through all themes and verifying text is readable against the bar backgrounds in each theme.

**Acceptance Scenarios**:

1. **Given** any theme is selected, **When** I view the header bar, **Then** the text has sufficient contrast against the background to be easily readable
2. **Given** any theme is selected, **When** I view the status bar, **Then** the text has sufficient contrast against the background to be easily readable

---

### Edge Cases

- What happens when the terminal is very narrow (< 60 columns)? Header/status bars should truncate gracefully or wrap appropriately.
- What happens when there are 0 posts? Header should show "Posts: 0" and feed area should show empty state message.
- What happens when terminal doesn't support true color? Themes should degrade gracefully to 256-color or 16-color modes.
- What happens when system locale cannot be determined for clock? Default to 24-hour format (HH:MM).

## Requirements *(mandatory)*

### Functional Requirements

**Header Bar:**
- **FR-001**: System MUST display a fixed header bar at the top of the screen that remains visible during scrolling
- **FR-002**: Header MUST show the smoke version in format "SMOKE v[version]"
- **FR-003**: Header MUST show total post count in format "Posts: [count]"
- **FR-004**: Header MUST show count of unique agents in format "Agents: [count]"
- **FR-005**: Header MUST show count of unique projects in format "Projects: [count]"
- **FR-006**: Header MUST show current time formatted according to system locale
- **FR-007**: Clock MUST update in real-time while TUI is running

**Status Bar:**
- **FR-008**: System MUST display a fixed status bar at the bottom of the screen that remains visible during scrolling
- **FR-009**: Status bar MUST show auto-refresh state with keybinding: "(a) auto: ON" or "(a) auto: OFF"
- **FR-010**: Status bar MUST show current style with keybinding: "(s) style: [style-name]"
- **FR-011**: Status bar MUST show current theme with keybinding: "(t) theme: [theme-name]"
- **FR-012**: Status bar MUST show current contrast with keybinding: "(c) contrast: [contrast-name]"
- **FR-013**: Status bar MUST show help and quit keybindings: "(?) help  (q) quit"

**Auto-Refresh:**
- **FR-014**: System MUST support toggling auto-refresh on/off with the (a) key
- **FR-015**: When auto-refresh is ON, feed MUST refresh every 5 seconds
- **FR-016**: When auto-refresh is OFF, feed MUST NOT refresh automatically
- **FR-017**: Auto-refresh state MUST be persisted across TUI sessions

**Theme System:**
- **FR-018**: System MUST support exactly 8 themes: Dracula, GitHub, Catppuccin, Solarized, Nord, Gruvbox, One Dark, Tokyo Night
- **FR-019**: Each theme MUST define colors for: Text, TextMuted, BackgroundSecondary, Accent, Error, and AgentColors (5 colors for agent name hashing)
- **FR-020**: Each theme MUST provide AdaptiveColor with both light and dark terminal variants
- **FR-021**: Header and status bars MUST use BackgroundSecondary color from current theme
- **FR-022**: Theme selection MUST be persisted across TUI sessions

**Visual Layout:**
- **FR-023**: Header bar, feed content, and status bar MUST be laid out vertically with header at top, content in middle, status at bottom
- **FR-024**: Feed content area MUST scroll independently while header and status bars remain fixed
- **FR-025**: Both bars MUST use matching visual treatment (same BackgroundSecondary)

### Key Entities

- **Theme**: A color scheme defining Text, TextMuted, BackgroundSecondary, Accent, Error, and AgentColors. Each color has light and dark variants for terminal adaptation.
- **FeedStats**: Computed statistics including total post count, unique agent count, and unique project count.
- **TUIConfig**: User preferences including current theme, style, contrast level, and auto-refresh state.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can identify current theme, style, contrast, and auto-refresh state within 2 seconds of looking at the status bar
- **SC-002**: All 8 themes display readable text on header/status bars (verified through visual inspection on both light and dark terminals)
- **SC-003**: Feed statistics in header accurately reflect actual post/agent/project counts (verified against feed data)
- **SC-004**: Clock displays in user's locale format and updates every minute
- **SC-005**: Auto-refresh toggle responds within 100ms of keypress
- **SC-006**: Settings (theme, auto-refresh state) persist correctly across TUI restarts (verified by changing settings, restarting, and confirming values)
- **SC-007**: Header and status bars remain visible and correctly positioned during feed scrolling

## Assumptions

- The existing TUIConfig infrastructure can be extended to store auto-refresh state
- The existing theme system (themes.go, contrast.go) will be refactored to support the new simplified theme interface
- Terminal supports basic ANSI colors at minimum; true color is preferred but not required
- System locale time formatting is available via Go's standard library
- Version string is available from existing build/version infrastructure
