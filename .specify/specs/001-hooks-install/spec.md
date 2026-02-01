# Feature Specification: Smoke Hooks Installation System

**Feature Branch**: `001-hooks-install`
**Created**: 2026-01-31
**Status**: Draft
**Input**: User description: "Hook installation system for smoke - allows users to install/uninstall Claude Code hooks that nudge agents to post to smoke during natural pauses"

## Design Philosophy

This feature follows smoke's constitution: **"Smoke SHOULD work with zero configuration. Init creates necessary structure automatically."**

Hooks are installed by default during `smoke init` - no separate step required. Users who don't want hooks can uninstall them. This matches the "sensible defaults over options" principle.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Initialize Smoke with Hooks (Priority: P1)

A new user runs `smoke init` to set up smoke. As part of initialization, Claude Code hooks are automatically installed so that agents receive gentle prompts to share observations during natural pauses. The user doesn't need to know hooks exist - they just work.

**Why this priority**: This is the core experience - smoke should work fully out of the box. Hooks are part of "complete initialization," not an optional add-on.

**Independent Test**: Can be fully tested by running `smoke init` on a fresh system and verifying both smoke config AND hooks are in place.

**Acceptance Scenarios**:

1. **Given** smoke is not initialized, **When** user runs `smoke init`, **Then** smoke config is created AND hook scripts are installed AND Claude Code settings are updated
2. **Given** smoke is already initialized but hooks are missing, **When** user runs `smoke init`, **Then** system reports already initialized and suggests `smoke hooks install` to add hooks
3. **Given** Claude Code directory (~/.claude/) doesn't exist, **When** user runs `smoke init`, **Then** system creates the directory and installs hooks

---

### User Story 2 - Reinstall/Repair Hooks (Priority: P2)

A user's hooks have become corrupted, outdated after a smoke upgrade, or were previously uninstalled. They run `smoke hooks install` to restore the hooks to working state.

**Why this priority**: Repair/reinstall is needed for maintenance but not for initial setup (which is handled by init).

**Independent Test**: Can be tested by removing hook files, running install, and verifying hooks are restored.

**Acceptance Scenarios**:

1. **Given** hooks are missing or corrupted, **When** user runs `smoke hooks install`, **Then** hook scripts are written and Claude Code settings are updated
2. **Given** hooks are already installed and current, **When** user runs `smoke hooks install`, **Then** system reports hooks are up to date (idempotent)
3. **Given** user has modified hook scripts, **When** user runs `smoke hooks install --force`, **Then** hooks are overwritten with fresh copies

---

### User Story 3 - Check Hook Installation Status (Priority: P2)

A user wants to verify whether smoke hooks are properly installed and functioning. They run a status command that shows current installation state, hook locations, and any potential issues.

**Why this priority**: Users need to diagnose issues and confirm hooks are working - essential for troubleshooting.

**Independent Test**: Can be tested by running status command in various states (not installed, installed, partially installed) and verifying accurate reporting.

**Acceptance Scenarios**:

1. **Given** hooks are not installed, **When** user runs `smoke hooks status`, **Then** system reports "not installed" with instructions to run `smoke hooks install`
2. **Given** hooks are fully installed, **When** user runs `smoke hooks status`, **Then** system reports "installed" with hook script locations and settings status
3. **Given** hook scripts exist but settings are missing, **When** user runs `smoke hooks status`, **Then** system reports "partially installed" with specific missing components

---

### User Story 4 - Uninstall Smoke Hooks (Priority: P3)

A user decides they don't want smoke nudges during their Claude Code sessions. They run an uninstall command that removes the hooks cleanly without affecting other Claude Code settings or smoke functionality.

**Why this priority**: Important for user control and opt-out, but less frequently used.

**Independent Test**: Can be tested by running uninstall and verifying hooks are removed while other settings remain.

**Acceptance Scenarios**:

1. **Given** hooks are installed, **When** user runs `smoke hooks uninstall`, **Then** hook scripts are removed and Claude Code settings entries are removed
2. **Given** hooks are not installed, **When** user runs `smoke hooks uninstall`, **Then** system reports nothing to uninstall
3. **Given** user has other hooks in Claude Code settings, **When** user runs `smoke hooks uninstall`, **Then** only smoke-related hooks are removed, other hooks remain untouched

---

### Edge Cases

- What happens when Claude Code settings.json has invalid JSON? System reports error and does not modify file; init continues but warns hooks couldn't be installed.
- What happens when user doesn't have write permission to ~/.claude/? System warns hooks couldn't be installed but completes smoke init successfully.
- What happens when hook scripts are modified by user after installation? Status shows "modified" warning; uninstall still removes them; install --force overwrites.
- What happens when smoke binary is not in PATH during hook execution? Hooks gracefully degrade with fallback message (already handled in hook scripts).
- What happens during init if hooks already exist from previous install? Init is idempotent - doesn't duplicate or corrupt existing hooks.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `smoke init` MUST install Claude Code hooks as part of initialization
- **FR-002**: `smoke init` MUST gracefully handle hook installation failures (warn but don't fail init)
- **FR-003**: System MUST provide `smoke hooks install` command to reinstall/repair hooks
- **FR-004**: System MUST provide `smoke hooks uninstall` command to remove hooks
- **FR-005**: System MUST provide `smoke hooks status` command to report installation state
- **FR-006**: System MUST preserve existing Claude Code hooks when installing (merge, not replace)
- **FR-007**: System MUST preserve existing Claude Code hooks when uninstalling (remove only smoke hooks)
- **FR-008**: System MUST create ~/.claude/hooks/ directory if it doesn't exist
- **FR-009**: System MUST create or update ~/.claude/settings.json with hook entries
- **FR-010**: System MUST bundle hook scripts within the smoke binary (embedded assets)
- **FR-011**: System MUST report clear success/failure messages for all operations
- **FR-012**: All hook operations MUST be idempotent

### Key Entities

- **Hook Script**: Bash script that runs on Claude Code events (Stop, PostToolUse). Contains logic to determine when to nudge and what message to show. Two scripts: smoke-break.sh (Stop) and smoke-nudge.sh (PostToolUse).
- **Settings Entry**: JSON configuration in ~/.claude/settings.json that registers hooks with Claude Code, specifying event type and script path.
- **Installation State**: The combination of hook scripts existing on disk and settings entries referencing them. States: installed, not-installed, partially-installed, modified.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: New users get working hooks with just `smoke init` - no additional commands needed
- **SC-002**: Users can verify installation status with clear, actionable output
- **SC-003**: Users can opt-out of hooks with a single uninstall command
- **SC-004**: Reinstall/repair works reliably after corruption or upgrade
- **SC-005**: 100% of hook operations are idempotent (same result when run multiple times)
- **SC-006**: Hook installation failures during init don't prevent smoke from working

## Assumptions

- Users have Claude Code installed (the tool these hooks integrate with)
- Users have write access to their home directory (~/.claude/)
- The hook scripts use bash, which is available on macOS and Linux (primary targets)
- Claude Code's settings.json format follows the documented hook structure
- Hook scripts are small enough to embed in the smoke binary (<10KB each)

## Out of Scope

- Windows support (bash hooks won't work natively)
- Customization of hook thresholds via commands (users can edit scripts post-install)
- Auto-update of hooks when smoke is upgraded (users run `smoke hooks install` to update)
- Integration with other AI coding assistants beyond Claude Code
- Skipping hook installation during init (users can uninstall after)
