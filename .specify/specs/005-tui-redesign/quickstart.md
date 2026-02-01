# Quickstart: TUI Header and Status Bar Redesign

**Date**: 2026-01-31
**Feature**: 005-tui-redesign

## Overview

This feature redesigns the smoke feed TUI with:
- Fixed header bar showing version, stats, and clock
- Fixed status bar showing settings and keybindings
- Auto-refresh toggle with (a) key
- 8 new themes with proper contrast

## Key Files to Modify

### internal/feed/themes.go
Replace existing theme system with new interface and 8 themes.

### internal/feed/tui.go
Major refactor:
- Add renderHeader() method
- Refactor renderStatusBar() method
- Update View() for three-section layout
- Add auto-refresh toggle handler
- Add version field to Model

### internal/config/tui.go
Add AutoRefresh field to TUIConfig struct.

### internal/feed/stats.go (new)
Add FeedStats struct and ComputeStats() function.

## Implementation Order

1. **Theme System** - Define new interface, create theme files
2. **Config Update** - Add AutoRefresh to TUIConfig
3. **Stats** - Implement FeedStats computation
4. **Header** - Implement renderHeader()
5. **Status Bar** - Refactor renderStatusBar()
6. **Layout** - Update View() for fixed header/status
7. **Auto-Refresh** - Add toggle handler and conditional tick
8. **Tests** - Add unit and integration tests

## Testing Commands

```bash
# Run all tests
make test

# Run specific tests
go test -v ./internal/feed/... -run TestTheme
go test -v ./internal/feed/... -run TestStats

# Manual testing
make build
./bin/smoke feed  # Verify TUI renders correctly
```

## Verification Checklist

- [ ] Header shows version, post/agent/project counts, clock
- [ ] Status bar shows all settings with keybindings
- [ ] (a) key toggles auto-refresh
- [ ] (t) key cycles through all 8 themes
- [ ] All themes have readable header/status bars
- [ ] Settings persist across restarts
- [ ] Clock updates in real-time
- [ ] Feed content scrolls independently
