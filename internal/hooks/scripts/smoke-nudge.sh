#!/bin/bash
# smoke-nudge.sh - Claude Code PostToolUse hook
# Triggers after each tool call
# Nudges agent to post if dual threshold met: >50 tools AND >10 mins since last prompt

set -euo pipefail

# Ensure smoke is discoverable in hook environment
export PATH="${HOME}/go/bin:/opt/homebrew/bin:/usr/local/bin:${PATH}"
SMOKE_BIN="${SMOKE_BIN:-smoke}"

# Hook log for diagnostics
HOOK_LOG="${HOME}/.claude/hooks/smoke-hook.log"
log_event() {
    local event="$1"
    local session="${CLAUDE_SESSION_ID:-default}"
    local tools_since="${CLAUDE_TOOL_COUNT_SINCE_LAST_HUMAN:-0}"
    local tools_total="${CLAUDE_TOOL_COUNT:-$tools_since}"
    local now
    now=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    printf "%s | session=%s | tools=%s | tools_since=%s | event=%s\n" "$now" "$session" "$tools_total" "$tools_since" "$event" >> "$HOOK_LOG" 2>/dev/null || true
}

# Check if smoke is available
if ! command -v "$SMOKE_BIN" &> /dev/null; then
    log_event "smoke_missing"
    exit 0
fi

# State directory (per-session tracking)
STATE_DIR="${HOME}/.claude/hooks/smoke-nudge-state"
mkdir -p "$STATE_DIR"

# Get session ID from environment
SESSION_ID=${CLAUDE_SESSION_ID:-default}
STATE_FILE="$STATE_DIR/$SESSION_ID"

# Get current timestamp
NOW=$(date +%s)

# Get tool count since last human message
TOOL_COUNT=${CLAUDE_TOOL_COUNT_SINCE_LAST_HUMAN:-0}

# Thresholds
TOOL_THRESHOLD=50
TIME_THRESHOLD=600  # 10 minutes in seconds

# Read last prompt timestamp (or initialize)
if [ -f "$STATE_FILE" ]; then
    LAST_PROMPT=$(cat "$STATE_FILE")
else
    # Initialize with current time
    echo "$NOW" > "$STATE_FILE"
    exit 0
fi

# Calculate time since last prompt
TIME_SINCE_PROMPT=$((NOW - LAST_PROMPT))

# Check dual threshold
if [ "$TOOL_COUNT" -gt "$TOOL_THRESHOLD" ] && [ "$TIME_SINCE_PROMPT" -gt "$TIME_THRESHOLD" ]; then
    # Reset timestamp on nudge
    echo "$NOW" > "$STATE_FILE"
    if "$SMOKE_BIN" suggest --context=breakroom; then
        log_event "nudge_hook_fired"
    else
        log_event "nudge_hook_error"
    fi
fi

# Cleanup: remove state files older than 24 hours
find "$STATE_DIR" -type f -mtime +1 -delete 2>/dev/null || true
