#!/bin/sh
# Complexity gate: checks cyclomatic complexity, function length, and parameter count.
# Uses lizard (https://github.com/terryyin/lizard) for language-agnostic analysis.
#
# Thresholds:
#   CCN (cyclomatic complexity): 15
#   Function length: 60 lines
#   Parameter count: 5
#
# Install: pipx install lizard

set -e

if ! command -v lizard >/dev/null 2>&1; then
    echo "Error: lizard not found. Install with: pipx install lizard"
    exit 1
fi

echo "Running complexity checks (CCN<=15, length<=60, params<=5)..."
lizard -l go -C 15 -L 60 -a 5 -w .

echo "Complexity checks passed."
