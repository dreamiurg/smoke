# Quickstart: Unread Messages Marker

**Feature**: 007-unread-marker
**Date**: 2026-02-01

## Overview

This feature adds visual tracking of unread messages in the TUI feed with:
- A "NEW" separator line between read and unread posts
- Keyboard shortcuts to mark messages as read
- Unread count in status bar
- Persistent read state across sessions

## Implementation Steps

### Step 1: Read State Storage (`internal/config/readstate.go`)

Create new file for read state persistence:

```go
// Key functions to implement:
func GetReadStatePath() string              // ~/.config/smoke/readstate.yaml
func LoadReadState() (*ReadState, error)    // Load from file
func SaveReadState(*ReadState) error        // Atomic save to file
func GetLastReadPostID(identity string) string
func SetLastReadPostID(identity, postID string) error
```

### Step 2: TUI Model Changes (`internal/feed/tui.go`)

Add fields to Model struct:
```go
type Model struct {
    // ... existing fields ...
    lastReadPostID string  // Boundary between read/unread
    unreadCount    int     // For status bar
    identity       string  // Current user identity
}
```

### Step 3: Separator Rendering

Add to `buildAllContentLines()`:
- Check each post against `lastReadPostID`
- Insert separator line when crossing the boundary
- Use existing `formatDaySeparator()` pattern for styling

New function:
```go
func (m Model) formatUnreadSeparator() string
```

### Step 4: Keybindings

Add to `Update()` key handling:
```go
case "m":
    // Mark all as read
    m.lastReadPostID = ""
    m.unreadCount = 0
    SetLastReadPostID(m.identity, latestPostID)
    return m, nil

case "M":
    // Mark to current scroll position
    // Find post at scroll position, save that ID
    return m, nil
```

### Step 5: Status Bar Update

Modify `renderStatusBar()`:
```go
// Add unread indicator
if m.unreadCount > 0 {
    parts = append(parts, fmt.Sprintf("(m)ark: %d new", m.unreadCount))
} else {
    parts = append(parts, "(m)ark")
}
```

### Step 6: Help Overlay Update

Add to `renderHelpOverlay()`:
```go
helpContent.WriteString("   m      Mark all read\n")
helpContent.WriteString("   M      Mark to here\n")
```

## File Changes Summary

| File | Action | Description |
|------|--------|-------------|
| `internal/config/readstate.go` | NEW | Read state persistence |
| `internal/config/readstate_test.go` | NEW | Unit tests |
| `internal/feed/tui.go` | MODIFY | Add separator, keybindings |
| `internal/feed/tui_test.go` | MODIFY | Add tests for new behavior |

## Testing Checklist

- [ ] Read state loads correctly on TUI start
- [ ] Separator appears at correct position
- [ ] `m` key marks all as read (separator disappears)
- [ ] `M` key marks to scroll position
- [ ] Status bar shows unread count
- [ ] Read state persists across sessions
- [ ] No separator on first-time use
- [ ] Auto-refresh doesn't move separator
- [ ] Works with all themes/contrast levels
- [ ] Help overlay shows new keybindings

## Success Verification

```bash
# 1. Start with clean state
rm ~/.config/smoke/readstate.yaml

# 2. Post some messages
smoke post "test message 1"
smoke post "test message 2"

# 3. Open TUI, close it
smoke feed  # No separator on first view
# Press q to quit

# 4. Post more messages
smoke post "test message 3"

# 5. Reopen TUI
smoke feed  # Should see "NEW" separator before message 3

# 6. Press 'm' to mark read
# Separator should disappear

# 7. Quit and reopen
smoke feed  # No separator (all read)
```
