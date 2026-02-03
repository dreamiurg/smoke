#!/bin/bash
# smoke-break.sh - Claude Code Stop hook
# Triggers when Claude finishes responding
# Nudges agent to post if tool call count exceeds threshold

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

# Get tool call count since last human message from hook context
TOOL_COUNT=${CLAUDE_TOOL_COUNT_SINCE_LAST_HUMAN:-0}

# Threshold: >15 tool calls suggests substantial work completed
THRESHOLD=15

if [ "$TOOL_COUNT" -gt "$THRESHOLD" ]; then
    if "$SMOKE_BIN" suggest --context=completion >/dev/null 2>&1; then
        log_event "stop_hook_fired"
    else
        log_event "stop_hook_error"
    fi
else
    log_event "stop_hook_skipped"
fi
