#!/bin/bash
# Analyze a single Claude session for smoke adoption metrics
# Usage: ./scripts/session-stats.sh <session-file-or-id>
#
# Output: JSON with session metrics

set -euo pipefail

SESSION="${1:-}"

# Find session file
if [[ -z "$SESSION" ]]; then
    echo "Usage: $0 <session-file-or-id>" >&2
    exit 1
fi

# If it's a short ID, find the full path
if [[ ! -f "$SESSION" ]]; then
    SESSION=$(find ~/.claude/projects -name "${SESSION}*.jsonl" ! -path "*/subagents/*" 2>/dev/null | head -1)
fi

if [[ -z "$SESSION" ]] || [[ ! -f "$SESSION" ]]; then
    echo "Session not found: $1" >&2
    exit 1
fi

# Extract metrics using jq
jq -s '
{
  session_id: (.[0].sessionId // "unknown"),

  # Count tool calls
  tool_calls: [.[] | select(.type == "assistant") | .message.content[]? | select(.type == "tool_use")] | length,

  # Count human messages (text input, not tool results)
  human_messages: [.[] | select(.type == "user" and (.message.content | type == "array") and (.message.content | any(.type == "text")))] | length,

  # Count smoke post commands
  smoke_posts: [.[] | select(.type == "assistant") | .message.content[]? | select(.type == "tool_use" and .name == "Bash") | .input.command // "" | select(test("smoke post"))] | length,

  # Tool calls since last human input
  tools_since_human: (
    (to_entries | map(select(.value.type == "user" and (.value.message.content | type == "array") and (.value.message.content | any(.type == "text")))) | last | .key) as $idx |
    [.[($idx // -1) + 1:][] | select(.type == "assistant") | .message.content[]? | select(.type == "tool_use")] | length
  ),

  # Time range
  first_ts: (map(select(.timestamp)) | first | .timestamp // null),
  last_ts: (map(select(.timestamp)) | last | .timestamp // null)
}
' "$SESSION"
