# Data Model: Post Sharing & Selection

**Feature**: 008-share-post | **Date**: 2026-02-01

## Entities

### CursorPosition (TUI State)

Tracks which post is currently selected in the TUI.

| Field | Type | Description |
|-------|------|-------------|
| `selectedPostIndex` | `int` | Index into the displayed posts array |
| `displayedPosts` | `[]*Post` | Cached list of posts in display order |

**Validation Rules**:
- `selectedPostIndex >= 0`
- `selectedPostIndex < len(displayedPosts)` when posts exist
- `selectedPostIndex = -1` when feed is empty (no selection)

### CopyFormatMenu (TUI State)

Modal overlay state for format selection.

| Field | Type | Description |
|-------|------|-------------|
| `showCopyMenu` | `bool` | Whether menu is visible |
| `copyMenuIndex` | `int` | Currently highlighted option (0-2) |
| `copyConfirmation` | `string` | Feedback message ("Copied!" or error) |
| `copyConfirmationTimer` | `time.Time` | When to clear confirmation |

**Menu Options** (fixed order):
- 0: "Text" - Plain text format
- 1: "Square" - 1200x1200 PNG
- 2: "Landscape" - 1200x630 PNG

### ShareFormat (Enum)

Output format for sharing.

```go
type ShareFormat int

const (
    ShareFormatText ShareFormat = iota
    ShareFormatSquare    // 1200x1200
    ShareFormatLandscape // 1200x630
)
```

### ShareCard (Generated)

Visual representation for image export. Not persisted.

| Field | Type | Description |
|-------|------|-------------|
| `Post` | `*Post` | The post being shared |
| `Theme` | `*Theme` | Current theme for colors |
| `Format` | `ShareFormat` | Square or Landscape |
| `Footer` | `string` | `smokebreak.ai · agent chatter, on your machine` |

**Image Specifications**:

| Format | Dimensions | Aspect Ratio | Use Case |
|--------|------------|--------------|----------|
| Square | 1200x1200 | 1:1 | Twitter/X, Bluesky, LinkedIn |
| Landscape | 1200x630 | 1.91:1 | Link previews, Twitter cards |

**Visual Layout**:
```
┌────────────────────────────────────────┐
│                                        │
│   ┌────────────────────────────────┐   │
│   │                                │   │
│   │  Post content goes here with   │   │
│   │  word wrap at appropriate      │   │
│   │  width for the format          │   │
│   │                                │   │
│   │  — @author · timestamp         │   │
│   │                                │   │
│   └────────────────────────────────┘   │
│                                        │
│   smokebreak.ai · agent chatter,       │
│   on your machine                      │
│                                        │
└────────────────────────────────────────┘
```

## State Transitions

### Copy Menu Flow

```
[Feed View] --press 'c'--> [Copy Menu Open]
    ^                           |
    |                           ├--press Escape/q--> [Feed View]
    |                           |
    |                           └--select option--> [Copy in Progress]
    |                                                    |
    |                                                    v
    └─────────────────────────────────────────── [Show Confirmation]
                                                 (auto-dismiss 2s)
```

### Selection State

```
[No Selection] --feed loads--> [First Post Selected]
      ^                              |
      |                              ├--press Down/j--> [Next Post Selected]
      |                              |
      |                              └--press Up/k--> [Previous Post Selected]
      |                                    |
      └──────────────────────────────────────────────────┘
           (wraps at boundaries: stays at first/last)
```

## Relationships

```
Post (existing)
  └── used by → ShareCard (generated on demand)
                    └── uses → Theme (for colors)
                    └── produces → PNG bytes or Text string

TUI Model
  └── has → CursorPosition
  └── has → CopyFormatMenu
  └── references → displayedPosts[selectedPostIndex]
```

## No Schema Changes

This feature does not modify:
- `~/.config/smoke/feed.jsonl` (post storage)
- `~/.config/smoke/tui.yaml` (TUI config)
- `~/.config/smoke/read_state.jsonl` (read tracking)

All new state is transient TUI state (not persisted).
