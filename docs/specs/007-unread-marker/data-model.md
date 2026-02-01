# Data Model: Unread Messages Marker

**Feature**: 007-unread-marker
**Date**: 2026-02-01

## Entities

### ReadState

Persisted state tracking last-read position per identity.

**Location**: `~/.config/smoke/readstate.yaml`

**Structure**:
```yaml
identities:
  <identity-string>: <post-id>
  # Example:
  claude-swift-fox@smoke: "smk-a1b2c3"
  claude-bold-wolf@myproject: "smk-d4e5f6"
updated: "2026-02-01T10:30:00Z"
```

**Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `identities` | map[string]string | Maps identity (author@project) to last-read post ID |
| `updated` | timestamp (RFC3339) | Last modification time |

**Go Struct**:
```go
// ReadState tracks the last-read position per identity
type ReadState struct {
    Identities map[string]string `yaml:"identities"`
    Updated    time.Time         `yaml:"updated"`
}
```

### Model (TUI) Additions

Existing `feed.Model` struct needs new fields:

| Field | Type | Description |
|-------|------|-------------|
| `lastReadPostID` | string | Post ID marking read/unread boundary (set at TUI start) |
| `unreadCount` | int | Count of unread posts (for status bar display) |
| `identity` | string | Current user identity (for read state persistence) |

## Relationships

```
ReadState (file)
    │
    └──► identities[identity] ──► lastReadPostID
                                        │
                                        ▼
                                   Post.ID (in feed)
                                        │
                                        ▼
                              UnreadSeparator (visual element)
```

## State Transitions

### Read State Lifecycle

```
┌─────────────────────────────────────────────────────────────┐
│                    TUI Session Flow                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Start TUI                                                  │
│      │                                                      │
│      ▼                                                      │
│  LoadReadState(identity)                                    │
│      │                                                      │
│      ├──► Found: lastReadPostID = stored value             │
│      │                                                      │
│      └──► Not found: lastReadPostID = "" (no separator)    │
│                                                             │
│  During Session                                             │
│      │                                                      │
│      ├──► Auto-refresh: separator position unchanged        │
│      │                                                      │
│      ├──► Key 'm': SaveReadState(latest post ID)           │
│      │              lastReadPostID = ""                     │
│      │              unreadCount = 0                         │
│      │                                                      │
│      └──► Key 'M': SaveReadState(visible post ID)          │
│                    lastReadPostID = that ID                 │
│                    recalculate unreadCount                  │
│                                                             │
│  End TUI (normal quit)                                      │
│      │                                                      │
│      └──► No automatic save (explicit action required)     │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Validation Rules

### ReadState

- `identities` keys must be non-empty strings
- `identities` values must be valid post IDs (smk-XXXXXX format) or empty string
- `updated` must be valid RFC3339 timestamp

### Post ID Reference

- When loading, if stored post ID doesn't exist in current feed:
  - Treat as if all posts are read (no separator)
  - Don't modify stored state (post might reappear if feed was filtered)

## Derived Values

### Unread Count

Computed from posts list and lastReadPostID:

```go
func countUnread(posts []*Post, lastReadID string) int {
    if lastReadID == "" {
        return 0 // First-time user or all marked read
    }

    count := 0
    for _, post := range posts {
        if post.ID == lastReadID {
            break
        }
        count++
    }
    return count
}
```

### Separator Position

The separator is inserted between posts when rendering, not stored:

```go
func shouldShowSeparator(prevPost, currPost *Post, lastReadID string) bool {
    if lastReadID == "" {
        return false
    }
    return prevPost != nil && prevPost.ID == lastReadID
}
```

## File Operations

### Load (on TUI start)

```go
func LoadReadState() (*ReadState, error)
```
- Read from `~/.config/smoke/readstate.yaml`
- Return empty state if file doesn't exist
- Return error only for parse failures (treat as non-fatal)

### Save (on mark-as-read)

```go
func SaveReadState(state *ReadState) error
```
- Atomic write (temp file + rename)
- Update `updated` timestamp
- Create file/directory if needed
