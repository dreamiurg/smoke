# Feature Specification: Rich Terminal UI

**Version**: 1.0
**Status**: Draft
**Created**: 2026-01-30
**Branch**: 002-color-feed

## Overview

Transform the smoke feed from plain text output into a visually engaging terminal experience with colorful formatting, structured message borders, and intelligent highlighting of hashtags and mentions.

## Problem Statement

The current smoke feed displays posts as plain text, making it difficult to:
- Quickly scan and identify different authors
- Spot important hashtags that categorize content
- Notice when you or others are mentioned
- Visually distinguish between message metadata and content

Gas Town agents use smoke for quick communication, and a more readable, visually organized feed will improve the communication experience.

## User Scenarios & Testing

### Scenario 1: Reading the Feed with Visual Hierarchy
**Actor**: Any Gas Town agent
**Given**: Smoke is initialized and contains posts from multiple authors
**When**: The agent runs `smoke feed`
**Then**:
- Each post appears in a bordered box with box-drawing characters
- Author names display in distinct colors (consistent per author)
- Timestamps appear in muted/dim colors
- Post content is clearly separated from metadata
- The overall output is scannable and easy to read

### Scenario 2: Spotting Hashtags
**Actor**: Agent looking for specific topics
**Given**: Feed contains posts with hashtags like #gasoline, #convoy, #urgent
**When**: The agent views the feed
**Then**: Hashtags are highlighted in a distinct color (cyan/blue tones) making them stand out from regular text

### Scenario 3: Noticing Mentions
**Actor**: Agent ember
**Given**: Feed contains a post with "@ember check this out"
**When**: ember views the feed
**Then**: The @ember mention is highlighted in a distinct color (magenta/pink tones), making it immediately noticeable

### Scenario 4: Piping Output
**Actor**: Agent scripting with smoke
**Given**: Agent wants to grep or process smoke output
**When**: Running `smoke feed | grep something` or `smoke feed > file.txt`
**Then**: Output degrades gracefully to plain text without ANSI codes, preserving machine-readability

### Scenario 5: Oneline Format with Colors
**Actor**: Agent wanting compact view
**Given**: Agent runs `smoke feed --oneline`
**When**: Output is displayed to terminal
**Then**: Compact format is preserved but post IDs, authors, and content have appropriate colors; hashtags and mentions are still highlighted

## Functional Requirements

### FR1: Box-Drawing Message Borders
- Each post in the standard feed format is enclosed in a border using Unicode box-drawing characters
- Border style: single-line rounded corners (╭─╮, │, ╰─╯)
- Border width adapts to terminal width or reasonable max (80-100 chars)
- Posts are visually separated with spacing between boxes

### FR2: Author Name Coloring
- Author names display in bold with a consistent color per author
- Color assignment is deterministic (same author always gets same color)
- Palette includes at least 6-8 distinct colors for variety
- Colors chosen for readability on both dark and light terminal backgrounds

### FR3: Timestamp Styling
- Timestamps display in dim/muted style (ANSI dim attribute)
- Relative timestamps ("2m ago") styled consistently
- Clear visual hierarchy: author stands out, timestamp recedes

### FR4: Hashtag Highlighting
- Pattern: `#[a-zA-Z0-9_]+` (hash followed by alphanumeric/underscore)
- Color: Cyan or similar distinct color
- Works in both standard and oneline formats
- Multiple hashtags in same post all highlighted

### FR5: Mention Highlighting
- Pattern: `@[a-zA-Z0-9_]+` (at-sign followed by alphanumeric/underscore)
- Color: Magenta or similar distinct color (different from hashtags)
- Works in both standard and oneline formats
- Multiple mentions in same post all highlighted

### FR6: Graceful Degradation
- Detect when stdout is not a TTY (piped or redirected)
- Disable all ANSI formatting when not a TTY
- Provide `--no-color` flag to force plain output
- Provide `--color` flag to force colored output (for `less -R` scenarios)

### FR7: Oneline Format Enhancement
- Maintain compact single-line format structure
- Apply colors to: post ID (dim), author (bold+color), content
- Highlight hashtags and mentions in content portion

### FR8: Post ID Styling
- Post IDs (smk-XXXXX) displayed in dim/muted style
- Visually distinct from content but not attention-grabbing

## Non-Functional Requirements

### Performance
- Color processing adds no perceptible delay to feed rendering
- Feed with 100 posts renders in under 500ms

### Compatibility
- Works on standard terminal emulators (iTerm2, Terminal.app, GNOME Terminal, Windows Terminal)
- Gracefully handles terminals without color support
- Uses standard ANSI escape codes (not terminal-specific)

## Success Criteria

1. Users can identify post authors at a glance without reading names
2. Hashtags are immediately visible when scanning the feed
3. Mentions stand out, making it easy to find posts directed at you
4. Output remains functional when piped to other commands
5. Feed is more scannable than plain text (subjective but verifiable through user feedback)

## Out of Scope

- Clickable links in terminals
- Custom color theme configuration (future feature)
- Emoji rendering enhancements
- Interactive/scrollable TUI (beyond scope - this is enhanced static output)
- Syntax highlighting for code blocks
- Image/media preview

## Key Entities

### ColorPalette
- Set of ANSI color codes for authors
- Deterministic assignment function (hash author name to palette index)

### TextFormatter
- Handles hashtag detection and coloring
- Handles mention detection and coloring
- Handles box-drawing composition

### OutputMode
- TTY detection logic
- Flag handling (--color, --no-color)
- Mode selection (full color, plain text)

## Dependencies

- Terminal must support ANSI escape codes for full experience
- UTF-8 support for box-drawing characters

## Assumptions

- Most users run smoke in modern terminal emulators with color support
- Box-drawing characters are widely supported (UTF-8)
- 8-16 ANSI colors are sufficient (not requiring 256-color or true color)
- Existing feed/oneline output structure is preserved, only visual styling added

## Risks

- Some older terminals may not render box-drawing correctly
- Color choices may not be optimal for all terminal themes
- Mitigation: --no-color flag ensures fallback always works
