#!/bin/bash
# One-shot patrol check - analyze recent sessions vs smoke feed
# Usage: ./scripts/patrol-check.sh [minutes]
#
# Outputs a report comparing session activity to smoke posts

set -euo pipefail

MINS="${1:-10}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "=== Smoke Patrol Check (last ${MINS} mins) ==="
echo "Time: $(date '+%Y-%m-%d %H:%M:%S')"
echo ""

# Find active sessions
echo "## Active Sessions"
echo ""

total_tools=0
total_sessions=0
sessions_without_posts=0

while IFS= read -r f; do
    [[ -z "$f" ]] && continue

    stats=$("$SCRIPT_DIR/session-stats.sh" "$f" 2>/dev/null || echo '{}')

    tools=$(echo "$stats" | jq -r '.tool_calls // 0')
    humans=$(echo "$stats" | jq -r '.human_messages // 0')
    posts=$(echo "$stats" | jq -r '.smoke_posts // 0')
    since_human=$(echo "$stats" | jq -r '.tools_since_human // 0')

    # Only report sessions with significant activity
    if [[ "$tools" -gt 5 ]]; then
        session_id=$(basename "$f" .jsonl | cut -c1-8)

        # Calculate autonomy ratio
        if [[ "$humans" -gt 0 ]]; then
            ratio=$((tools / humans))
        else
            ratio=$tools
        fi

        # Flag if no posts despite high activity
        flag=""
        if [[ "$posts" -eq 0 ]] && [[ "$tools" -gt 20 ]]; then
            flag=" ⚠️ GAP"
            ((sessions_without_posts++))
        fi

        echo "- $session_id: $tools tools, $humans human, ratio=${ratio}:1, posts=$posts, pending=$since_human$flag"

        total_tools=$((total_tools + tools))
        ((total_sessions++))
    fi
done < <(find ~/.claude/projects -name "*.jsonl" -mmin -"$MINS" ! -path "*/subagents/*" 2>/dev/null)

echo ""
echo "## Smoke Feed (recent)"
echo ""
smoke feed --limit 5 --oneline 2>/dev/null | sed 's/^/- /'

echo ""
echo "## Summary"
echo ""
echo "- Active sessions: $total_sessions"
echo "- Total tool calls: $total_tools"
echo "- Sessions with gaps: $sessions_without_posts"

# Calculate adoption score
feed_posts=$(smoke feed --limit 20 2>/dev/null | grep -c "^smk-" || echo 0)
if [[ "$total_sessions" -gt 0 ]]; then
    echo "- Adoption signal: ${feed_posts} posts visible"
fi
