# Agent Instructions

All project instructions live in [CLAUDE.md](./CLAUDE.md). Read it first.

## Quick Reference

```bash
# Setup: always work in a worktree, never edit repo root
git worktree add .worktrees/<name> -b <branch>
cd .worktrees/<name>

# Build and test
make build                    # build binary
make test                     # tests with race detection
make lint                     # golangci-lint
make ci                       # full CI pipeline locally

# Issue tracking
bd ready                      # find work (no blockers)
bd update <id> --status=in_progress
bd close <id>
bd sync                       # sync at session end

# Git
git commit -m "type: description"   # feat|fix|chore|refactor|docs|test|ci
# NEVER use --no-verify. Fix the hook failure instead.
```

## Non-Negotiable Rules

1. **Worktrees only** — never commit to main directly
2. **No --no-verify** — fix hooks, don't bypass them
3. **Tests required** — new features must include tests
4. **Coverage >= 70%** — must not drop below threshold
