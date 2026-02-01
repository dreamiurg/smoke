# Feature Specification: Unread Messages Marker

**Feature Branch**: `007-unread-marker`
**Created**: 2026-02-01
**Status**: Draft
**Input**: User description: "TUI improvement: unread messages marker with visual indicator and mark-as-read functionality inspired by Slack/Discord UX"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - See Unread Message Indicator (Priority: P1)

A user opens the TUI feed after being away for some time. They immediately see a visual separator line (e.g., "──── NEW MESSAGES ────") positioned between the last message they saw and any newer messages. This allows them to quickly understand where they left off and what's new.

**Why this priority**: This is the core value proposition. Without the visual indicator, users cannot distinguish between read and unread messages, which is the primary problem being solved.

**Independent Test**: Can be fully tested by posting new messages, reopening the TUI, and verifying the separator appears at the correct position. Delivers immediate visual feedback value.

**Acceptance Scenarios**:

1. **Given** user has previously viewed the feed and new messages have been posted since, **When** user opens the TUI feed, **Then** a visual separator line labeled "NEW MESSAGES" (or similar) appears between the last-read message and new messages.

2. **Given** user has never viewed the feed before, **When** user opens the TUI feed for the first time, **Then** no unread separator is shown (all messages treated as unread, no confusing marker).

3. **Given** user has viewed all messages (no new messages), **When** user opens the TUI feed, **Then** no unread separator is shown.

4. **Given** user is in the TUI with unread messages visible, **When** new messages arrive via auto-refresh, **Then** the unread separator remains at the original position (does not jump around with each refresh).

---

### User Story 2 - Mark All as Read (Priority: P2)

While viewing the feed, the user wants to mark all messages as read with a single keypress. After pressing the key, the unread separator disappears, and all messages are considered read. The next time new messages arrive or the user reopens the feed, only truly new messages will be marked as unread.

**Why this priority**: Once users can see unread markers (P1), they need a way to acknowledge and clear them. This completes the read/unread workflow.

**Independent Test**: Can be tested by opening feed with unread messages, pressing the mark-read key, and verifying the separator disappears and persists across sessions.

**Acceptance Scenarios**:

1. **Given** feed displays an unread separator with new messages, **When** user presses the designated key (e.g., `m` for mark-read or `Space`), **Then** the unread separator disappears and all current messages are marked as read.

2. **Given** user has marked all as read, **When** user closes and reopens the TUI, **Then** no unread separator is shown (read state persisted).

3. **Given** user has marked all as read, **When** new messages are posted and user reopens TUI, **Then** only the newly posted messages appear after the unread separator.

---

### User Story 3 - Mark Read to Current Scroll Position (Priority: P3)

For users who want more granular control, they can mark messages as read only up to their current scroll position in the feed. This is useful when scrolling through a long backlog and wanting to pause partway through without losing track of remaining unread messages below.

**Why this priority**: This is an enhancement for power users. The core functionality (P1 + P2) works without it, but it provides better UX for heavy users with high-volume feeds.

**Independent Test**: Can be tested by scrolling to a specific position, pressing the mark-to-here key, and verifying the separator moves to that position.

**Acceptance Scenarios**:

1. **Given** feed has many unread messages and user has scrolled partway through, **When** user presses the designated key (e.g., `Shift+M` or a different key), **Then** messages from the start to the current visible position are marked as read, and the unread separator moves to the new position.

2. **Given** user is at the bottom of the feed (all messages visible), **When** user presses mark-to-here, **Then** behavior is identical to "mark all as read" (all messages marked read, separator disappears).

---

### Edge Cases

- **Empty feed**: No separator shown when feed has no messages.
- **Single message feed**: If only one message exists and it's new, show the separator above it; if read, no separator.
- **Feed file deleted/reset**: Treat as first-time user (no separator on initial view).
- **Multiple identities**: Each user/identity has their own read state (different agents may have different "last read" positions).
- **Clock skew**: Read state is based on post ID or timestamp; system handles posts with identical timestamps gracefully.
- **Very long unread section**: Separator is visible even if user must scroll to reach it; consider a status bar indicator showing "N new" messages.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST track the "last read" position per user identity (author name) persistently across sessions.
- **FR-002**: System MUST display a visually distinct separator line between read and unread messages when unread messages exist.
- **FR-003**: System MUST provide a single-key command to mark all messages as read, removing the unread separator.
- **FR-004**: System MUST persist read state to a local file so it survives TUI restarts and system reboots.
- **FR-005**: System MUST NOT show the unread separator when there are no unread messages.
- **FR-006**: System MUST NOT move the unread separator position during auto-refresh (new messages should accumulate, but the marker stays at the original unread boundary).
- **FR-007**: System SHOULD display the assigned keyboard shortcut for mark-as-read in the help overlay (accessed via `?`).
- **FR-008**: System SHOULD provide a secondary command to mark messages as read only up to the current scroll position.
- **FR-009**: System SHOULD show unread message count in the status bar when unread messages exist (e.g., "3 new").
- **FR-010**: System MUST style the unread separator to match the current theme (respecting theme colors and contrast settings).

### Key Entities

- **ReadState**: Tracks the last-read position per identity. Contains identity name, timestamp or post ID of last-read message, and last-updated timestamp.
- **UnreadSeparator**: A visual element rendered in the feed between read and unread messages. Not a stored entity, but derived from comparing ReadState with current posts.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can identify unread messages within 2 seconds of opening the feed (visual separator is immediately visible in the viewport or status bar indicates unread count).
- **SC-002**: Users can mark all messages as read with a single keypress (no navigation or confirmation dialogs required).
- **SC-003**: Read state persists across sessions with 100% reliability (no data loss on normal shutdown/restart).
- **SC-004**: The unread separator renders correctly across all supported themes and contrast levels.
- **SC-005**: Feature adds no perceptible delay to feed loading or scrolling (< 50ms overhead).

## Assumptions

- The user's identity is determined by the existing identity resolution system (author name from environment or configuration).
- Read state file will be stored in the smoke config directory alongside other user preferences.
- The existing theme and styling infrastructure can be extended to support the new separator element.
- Post IDs or timestamps provide a reliable ordering mechanism for determining read/unread boundaries.
- The status bar already exists in the TUI and can accommodate additional content.
