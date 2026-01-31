#!/bin/bash
# Patrol summary - overview for agents resuming patrol
# Usage: ./scripts/patrol-summary.sh
#
# Shows current state and what to investigate

set -euo pipefail

echo "=== Smoke Adoption Patrol Summary ==="
echo "Generated: $(date '+%Y-%m-%d %H:%M:%S')"
echo ""

# Feed stats
echo "## Feed Activity"
total_posts=$(wc -l < ~/.config/smoke/feed.jsonl 2>/dev/null || echo 0)
recent_posts=$(jq -r '.created_at' ~/.config/smoke/feed.jsonl 2>/dev/null | while read ts; do
    post_time=$(date -j -f "%Y-%m-%dT%H:%M:%S" "${ts%%.*}" +%s 2>/dev/null || echo 0)
    hour_ago=$(($(date +%s) - 3600))
    [[ "$post_time" -gt "$hour_ago" ]] && echo "1"
done | wc -l | tr -d ' ')

echo "- Total posts: $total_posts"
echo "- Posts last hour: $recent_posts"
echo ""

# Session activity (last 2 hours)
echo "## Sessions (last 2 hours)"
high_activity=0
gaps=0

for f in $(find ~/.claude/projects -name "*.jsonl" -mmin -120 ! -path "*/subagents/*" 2>/dev/null); do
    tools=$(jq -s '[.[] | select(.type == "assistant") | .message.content[]? | select(.type == "tool_use")] | length' "$f" 2>/dev/null || echo 0)
    posts=$(jq -s '[.[] | select(.type == "assistant") | .message.content[]? | select(.type == "tool_use" and .name == "Bash") | .input.command // "" | select(test("smoke post"))] | length' "$f" 2>/dev/null || echo 0)

    if [[ "$tools" -gt 50 ]]; then
        ((high_activity++))
        if [[ "$posts" -eq 0 ]]; then
            ((gaps++))
            echo "- GAP: $(basename "$f" | cut -c1-8) - $tools tools, 0 posts"
        fi
    fi
done

echo ""
echo "- High-activity sessions (>50 tools): $high_activity"
echo "- Sessions with adoption gaps: $gaps"
echo ""

# Recommendations
echo "## Patrol Actions"
if [[ "$gaps" -gt 0 ]]; then
    echo "1. Investigate gap sessions with: ./scripts/session-stats.sh <id>"
    echo "2. Check if Stop hook fired in those sessions"
    echo "3. Consider adding mid-session prompts"
else
    echo "- No obvious gaps. Monitor for new sessions."
fi
echo ""
echo "4. Run patrol check: ./scripts/patrol-check.sh 10"
echo "5. Log findings: bd update smoke-90a --append-notes=\"...\""
