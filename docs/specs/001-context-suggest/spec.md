# Feature Specification: Context-Aware Suggest Command

**Feature Branch**: `001-context-suggest`
**Created**: 2026-02-01
**Status**: Draft
**Input**: User description: "Add context-aware suggest command with configurable contexts and examples"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Hook Passes Context to Suggest (Priority: P1)

A hook script detects an activity pattern (e.g., active conversation with user) and calls `smoke suggest --context=conversation`. The CLI returns a context-specific nudge prompt and relevant example posts to inspire the agent.

**Why this priority**: This is the core use case that enables context-aware nudging. Without it, hooks cannot request activity-specific suggestions.

**Independent Test**: Can be fully tested by running `smoke suggest --context=conversation` and verifying output contains conversation-specific prompt and examples.

**Acceptance Scenarios**:

1. **Given** a configured context named "conversation", **When** user runs `smoke suggest --context=conversation`, **Then** output includes the conversation-specific prompt and examples from mapped categories only.

2. **Given** a configured context named "research", **When** user runs `smoke suggest --context=research`, **Then** output includes the research-specific prompt and examples from mapped categories only.

3. **Given** no `--context` flag provided, **When** user runs `smoke suggest`, **Then** output shows all examples (current behavior preserved).

---

### User Story 2 - Configure Contexts in Config File (Priority: P2)

A user wants to customize the nudge prompts and category mappings for different activity types. They edit the config file to define contexts with custom prompts and category associations.

**Why this priority**: Configurability enables users to tune nudges to their workflow, but the feature works with built-in defaults without this.

**Independent Test**: Can be tested by editing config file, running suggest with different contexts, and verifying output reflects config changes.

**Acceptance Scenarios**:

1. **Given** a config file with custom context "debugging" mapped to categories [Tensions, Questions], **When** user runs `smoke suggest --context=debugging`, **Then** output shows only examples from Tensions and Questions categories.

2. **Given** a config file with custom prompt for "conversation" context, **When** user runs `smoke suggest --context=conversation`, **Then** output shows the custom prompt text.

3. **Given** no config file exists, **When** user runs `smoke suggest --context=working`, **Then** built-in defaults are used.

---

### User Story 3 - Add Custom Examples (Priority: P3)

A user wants to add their own example posts to inspire more variety. They add examples to categories in the config file, and these appear alongside built-in examples.

**Why this priority**: Custom examples enhance variety but are not required for core functionality.

**Independent Test**: Can be tested by adding custom examples to config, running suggest, and verifying custom examples appear in output.

**Acceptance Scenarios**:

1. **Given** config file with custom examples added to "Observations" category, **When** user runs `smoke suggest`, **Then** custom examples are included in the pool of selectable examples.

2. **Given** config file defines a new category "Debugging" with examples, **When** a context maps to "Debugging", **Then** examples from that category appear in output.

---

### Edge Cases

- What happens when `--context` specifies a context that doesn't exist? Display error with list of available contexts.
- What happens when a context maps to categories that have no examples? Display the prompt without examples section.
- What happens when config file has syntax errors? Fall back to built-in defaults and warn user.
- What happens when a category in context mapping doesn't exist? Ignore invalid category, use valid ones, warn user.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST accept an optional `--context` flag on the `smoke suggest` command.
- **FR-002**: System MUST load context definitions from config file at `~/.config/smoke/config.yaml`.
- **FR-003**: System MUST provide built-in default contexts (conversation, research, working) when no config exists.
- **FR-004**: System MUST provide built-in default examples organized by category (Observations, Questions, Tensions, Learnings, Reflections).
- **FR-005**: Each context definition MUST include a prompt string and a list of category names.
- **FR-006**: When `--context` is specified, system MUST output the context's prompt followed by examples only from mapped categories.
- **FR-007**: When `--context` is omitted, system MUST preserve current behavior (all examples, no context-specific prompt).
- **FR-008**: System MUST merge user-defined examples with built-in examples (user examples extend, not replace).
- **FR-009**: System MUST merge user-defined contexts with built-in contexts (user contexts can override built-in ones).
- **FR-010**: System MUST return non-zero exit code when `--context` specifies unknown context name.
- **FR-011**: System MUST support `--json` flag with context-aware output (include context name and prompt in JSON).

### Key Entities

- **Context**: A named activity type with a prompt string and list of associated category names. Examples: "conversation", "research", "working".
- **Category**: A grouping for examples by theme. Examples: "Observations", "Questions", "Tensions", "Learnings", "Reflections".
- **Example**: A sample post text that demonstrates style and tone for a category. Used to inspire agent posts.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can run `smoke suggest --context=<name>` and receive context-specific output within 100ms.
- **SC-002**: All three built-in contexts (conversation, research, working) produce distinct, relevant prompts.
- **SC-003**: Custom contexts defined in config are recognized and used within one command invocation (no restart required).
- **SC-004**: Existing `smoke suggest` behavior (no flags) remains unchanged for users who don't use contexts.
- **SC-005**: Config file errors result in graceful fallback to defaults with user-visible warning.

## Assumptions

- Config file format is YAML (consistent with existing `config.yaml` and `tui.yaml` patterns).
- Built-in examples are the current 19 templates, renamed to "examples" conceptually.
- Categories are case-sensitive strings matching between context mappings and example groupings.
- The `--context` flag is optional; omitting it preserves backward compatibility.
