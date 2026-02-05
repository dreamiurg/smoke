#!/usr/bin/env bash
set -euo pipefail

ensure_gocyclo() {
  if command -v gocyclo >/dev/null 2>&1; then
    return 0
  fi
  echo "gocyclo not found, installing..."
  go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
  if ! command -v gocyclo >/dev/null 2>&1; then
    echo "gocyclo installation failed" >&2
    exit 1
  fi
}

run_check() {
  local limit="$1"
  local label="$2"
  shift 2
  local output
  output=$(gocyclo -over "$limit" "$@" || true)
  if [ -n "$output" ]; then
    echo "Complexity check failed for ${label} (max ${limit}):"
    echo "$output"
    return 1
  fi
  return 0
}

ensure_gocyclo

fail=0

run_check 45 "feed" internal/feed || fail=1
run_check 35 "cli" internal/cli || fail=1
run_check 25 "core" internal/config internal/hooks internal/identity internal/logging cmd tests || fail=1

if [ "$fail" -ne 0 ]; then
  exit 1
fi

echo "Complexity checks passed."
