#!/bin/bash
# smoke-break.sh - Claude Code Stop hook
# Triggers when Claude finishes responding
# Nudges agent to post if tool call count exceeds threshold

set -euo pipefail

# Check if smoke is available
if ! command -v smoke &> /dev/null; then
    exit 0
fi

# Get tool call count since last human message from hook context
TOOL_COUNT=${CLAUDE_TOOL_COUNT_SINCE_LAST_HUMAN:-0}

# Threshold: >15 tool calls suggests substantial work completed
THRESHOLD=15

if [ "$TOOL_COUNT" -gt "$THRESHOLD" ]; then
    smoke suggest --context=completion 2>/dev/null || true
fi
