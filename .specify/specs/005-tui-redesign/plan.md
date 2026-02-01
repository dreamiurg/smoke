# Implementation Plan: TUI Header and Status Bar Redesign

**Branch**: `005-tui-redesign` | **Date**: 2026-01-31 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/docs/specs/005-tui-redesign/spec.md`

## Summary

Redesign the smoke feed TUI with an Abacus-style layout featuring:
- Fixed header bar at top showing version, post/agent/project counts, and locale-formatted clock
- Fixed status bar at bottom showing current settings with keybindings: (a) auto, (s) style, (t) theme, (c) contrast, (?) help, (q) quit
- Auto-refresh toggle functionality
- New simplified theme system with 8 themes (Dracula, GitHub, Catppuccin, Solarized, Nord, Gruvbox, One Dark, Tokyo Night)
- Each theme provides: Text, TextMuted, BackgroundSecondary, Accent, Error, and 5 AgentColors with AdaptiveColor support

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**: github.com/charmbracelet/bubbletea, github.com/charmbracelet/lipgloss
**Storage**: JSONL at ~/.config/smoke/feed.jsonl (existing), config at ~/.config/smoke/tui.json (existing)
**Testing**: go test with -race flag, integration tests via compiled binary
**Target Platform**: macOS, Linux terminals (with ANSI/true color support)
**Project Type**: Single CLI application
**Performance Goals**: TUI must render at 60fps, key response <100ms
**Constraints**: Works with 256-color and true-color terminals, graceful degradation for basic terminals
**Scale/Scope**: ~10 files modified/created, ~1000 lines of code

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Go Simplicity | ✅ PASS | Using standard library for time/locale, lipgloss for styling (already a dependency) |
| II. Agent-First CLI Design | ✅ PASS | TUI is for human consumption of feed; agents interact via CLI commands |
| III. Local-First Storage | ✅ PASS | Config persisted to ~/.config/smoke/tui.json (existing pattern) |
| IV. Test What Matters | ✅ PASS | Will add unit tests for theme colors, integration tests for TUI rendering |
| V. Environment Integration | ✅ PASS | No changes to identity/environment handling |
| VI. Minimal Configuration | ✅ PASS | Sensible defaults for all settings, auto-refresh ON by default |

**Gate Status**: ✅ PASSED - No violations

## Project Structure

### Documentation (this feature)

```text
.specify/specs/005-tui-redesign/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
└── quickstart.md        # Phase 1 output
```

### Source Code (repository root)

```text
cmd/smoke/              # Entry point (no changes needed)
internal/
├── cli/
│   └── feed.go         # TUI command entry (minor changes)
├── config/
│   └── tui.go          # TUI config (add AutoRefresh field)
└── feed/
    ├── tui.go          # Main TUI model (major refactor for header/status/layout)
    ├── themes.go       # Theme definitions (replace with new 8-theme system)
    ├── theme_*.go      # Individual theme files (new)
    ├── stats.go        # Feed statistics computation (new)
    └── styles.go       # Layout styles (existing, minor updates)
```

**Structure Decision**: Single project structure following existing smoke conventions. New theme files follow Abacus pattern of one file per theme for maintainability.

## Complexity Tracking

> No violations - complexity tracking not needed.

## Implementation Phases

### Phase 1: Theme System Refactor
1. Define new Theme interface with simplified color set
2. Create 8 theme files with AdaptiveColor support
3. Update theme manager for cycling through new themes
4. Migrate existing code to use new theme interface

### Phase 2: TUI Layout Refactor
1. Implement fixed header bar component
2. Implement fixed status bar component
3. Refactor View() to use header/content/status layout
4. Ensure content area scrolls independently

### Phase 3: Header Bar Implementation
1. Add version display from build info
2. Implement FeedStats computation (post/agent/project counts)
3. Add locale-aware clock with real-time updates
4. Style header using theme BackgroundSecondary

### Phase 4: Status Bar Implementation
1. Implement status bar with all keybinding displays
2. Add auto-refresh toggle state display
3. Update status bar on setting changes
4. Style status bar matching header

### Phase 5: Auto-Refresh Toggle
1. Add AutoRefresh field to TUIConfig
2. Implement (a) key handler for toggle
3. Conditional tick command based on auto-refresh state
4. Persist auto-refresh state to config

### Phase 6: Testing & Polish
1. Unit tests for theme contrast
2. Unit tests for FeedStats computation
3. Integration tests for key handlers
4. Visual verification across all themes
