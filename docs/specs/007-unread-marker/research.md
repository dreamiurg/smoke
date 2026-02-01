# Research: Unread Messages Marker

**Feature**: 007-unread-marker
**Date**: 2026-02-01

## Research Questions

### 1. Read State Storage Format

**Question**: What format should be used to store the last-read position per identity?

**Decision**: YAML file with map of identity → post ID

**Rationale**:
- YAML already used for TUIConfig in the project
- Human-readable for debugging
- Simple key-value mapping sufficient
- Post ID is unique and stable (smk-XXXXXX format)

**Alternatives Considered**:
- JSON: Works but project uses YAML for config
- Separate file per identity: Unnecessary complexity
- Store in TUIConfig: Mixes concerns (display settings vs. read state)

### 2. Read Position Tracking

**Question**: Should we track by post ID or timestamp?

**Decision**: Track by post ID

**Rationale**:
- Post IDs are unique and immutable
- Timestamps could have duplicates (same-second posts)
- IDs are already used for thread linking (ParentID)
- No timezone issues

**Alternatives Considered**:
- Timestamp: Risk of same-second collisions
- Index position: Changes when posts are deleted/filtered
- Combination: Unnecessary complexity

### 3. Keybinding Assignment

**Question**: Which keys should be assigned for mark-as-read operations?

**Decision**:
- `m` - Mark all as read (mnemonic: "mark")
- `M` (Shift+M) - Mark read to current scroll position

**Rationale**:
- `m` is unused in current TUI
- Follows existing pattern (lowercase = primary, uppercase = variant)
- Mnemonic is intuitive
- Doesn't conflict with vim-style navigation (j/k/g/G)

**Alternatives Considered**:
- `Space` for mark-all: Too easy to hit accidentally while scrolling
- `r` for mark-read: Already used for refresh
- `Enter` for mark: Could be confusing (no action occurs)

### 4. Separator Visual Design

**Question**: How should the unread separator be styled?

**Decision**: Follow existing day separator pattern with distinct label

**Design**: `──────── NEW ────────` (full width, centered label)

**Rationale**:
- Consistent with existing `formatDaySeparator()` styling
- Uses theme colors (TextMuted for separator line)
- Distinct text makes it obvious
- "NEW" is short and clear (vs. "NEW MESSAGES" or "UNREAD")

**Alternatives Considered**:
- Different color: Theme consistency more important
- Emoji marker: Constitution discourages emojis
- Blinking/bold: Too intrusive, accessibility concern

### 5. First-Time User Experience

**Question**: What happens when a user opens TUI for the first time (no read state)?

**Decision**: Show no separator, mark current feed as implicitly read on first exit

**Rationale**:
- Showing all posts as "new" on first visit provides no value
- Avoids confusing marker when there's nothing to compare against
- Natural transition: first visit establishes baseline

**Alternatives Considered**:
- Show all as unread: Useless information
- Prompt user: Violates zero-config principle
- Show separator at top: Confusing

### 6. Auto-Refresh Behavior

**Question**: How should the separator behave when new posts arrive via auto-refresh?

**Decision**: Separator stays at original position; new posts accumulate above/below

**Rationale**:
- Jumping separator is disorienting
- User can see new posts accumulating
- Mark-as-read is explicit action
- Matches Slack/Discord behavior

**Implementation**: Store `lastReadPostID` at TUI start, don't update during session

### 7. Status Bar Display

**Question**: Should unread count be shown in status bar?

**Decision**: Yes, show "N new" when unread posts exist

**Rationale**:
- Provides at-a-glance count without scrolling
- Useful when separator is off-screen
- Non-intrusive addition to existing status bar

**Format**: Add to status bar: `(m)ark: 5 new` (or just status items when none)

## Technical Decisions

### File Location

Read state file: `~/.config/smoke/readstate.yaml`

```yaml
# Example content
identities:
  claude-swift-fox@smoke: "smk-a1b2c3"
  claude-bold-wolf@myproject: "smk-d4e5f6"
updated: "2026-02-01T10:30:00Z"
```

### Identity Resolution

Use existing identity pattern from feed commands:
1. Check SMOKE_NAME environment variable
2. Auto-generate from session context (existing behavior)
3. Key read state by full identity including suffix (author@project)

### Edge Cases

| Case | Behavior |
|------|----------|
| Empty feed | No separator, nothing to mark |
| All posts read | No separator |
| Feed file deleted | Treat as first-time (no separator) |
| Read state file corrupted | Ignore, treat as first-time |
| Post ID in read state no longer exists | Find closest older post, or treat as all-read |
| Multiple identities | Each identity has independent read state |
