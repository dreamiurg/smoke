# Feature Specification: Smoke Doctor Command

**Feature Branch**: `003-doctor-command`
**Created**: 2026-01-31
**Status**: Draft
**Input**: User description: "Add support for `smoke doctor` that would act and look similar to `bd doctor` and both show status and --fix things as needed"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Health Check Status (Priority: P1)

An agent runs `smoke doctor` to verify their smoke installation is working correctly before posting or reading the feed. The command displays a clear status report showing what's working and what needs attention.

**Why this priority**: This is the core functionality - agents need to quickly diagnose why smoke commands might be failing.

**Independent Test**: Can be fully tested by running `smoke doctor` and verifying it displays accurate status for each check. Delivers immediate value by showing installation health.

**Acceptance Scenarios**:

1. **Given** smoke is properly initialized, **When** agent runs `smoke doctor`, **Then** all checks show green checkmarks (✓) with "OK" status
2. **Given** smoke is not initialized, **When** agent runs `smoke doctor`, **Then** the initialization check shows warning (⚠) with clear message about running `smoke init`
3. **Given** the feed file is corrupted or missing, **When** agent runs `smoke doctor`, **Then** the feed file check shows error with specific problem description

---

### User Story 2 - Auto-Fix Problems (Priority: P2)

An agent encounters issues with their smoke installation and runs `smoke doctor --fix` to automatically repair common problems without manual intervention.

**Why this priority**: Auto-fix reduces friction for agents who encounter problems - they can self-heal without human intervention.

**Independent Test**: Can be tested by introducing a fixable problem (e.g., missing config file), running `--fix`, and verifying the problem is resolved.

**Acceptance Scenarios**:

1. **Given** smoke is not initialized, **When** agent runs `smoke doctor --fix`, **Then** smoke is automatically initialized and success message is displayed
2. **Given** feed file has been accidentally deleted, **When** agent runs `smoke doctor --fix`, **Then** an empty feed file is recreated
3. **Given** config file is missing but feed exists, **When** agent runs `smoke doctor --fix`, **Then** config file is regenerated with default values
4. **Given** all checks pass, **When** agent runs `smoke doctor --fix`, **Then** message indicates "No problems to fix"

---

### User Story 3 - Dry-Run Mode (Priority: P3)

An agent wants to see what `--fix` would do before actually making changes, using `smoke doctor --fix --dry-run` to preview repairs.

**Why this priority**: Provides safety for cautious users who want to understand changes before they happen.

**Independent Test**: Can be tested by introducing problems, running with `--dry-run`, verifying no changes made, then running without `--dry-run` to confirm fixes work.

**Acceptance Scenarios**:

1. **Given** smoke is not initialized, **When** agent runs `smoke doctor --fix --dry-run`, **Then** output shows what would be fixed without making changes
2. **Given** problems exist, **When** agent runs `smoke doctor --fix --dry-run`, **Then** exit code indicates problems exist but no files are modified

---

### Edge Cases

- What happens when config directory exists but is not writable?
- How does doctor handle partial initialization (some files exist, others don't)?
- What happens when feed file exists but is not valid JSONL?
- How does doctor behave when running without any environment (no home directory)?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST display a categorized health check report with clear pass/warn/fail indicators
- **FR-002**: System MUST check if smoke is initialized (config directory and feed file exist)
- **FR-003**: System MUST check if feed file is valid (exists, readable, valid JSONL format)
- **FR-004**: System MUST check if config file exists and is valid
- **FR-005**: System MUST check file permissions on config directory and files
- **FR-006**: System MUST display the smoke version at the top of the report
- **FR-007**: System MUST use consistent visual indicators: ✓ for pass, ⚠ for warning, ✗ for error
- **FR-008**: System MUST group checks into logical categories (e.g., "INSTALLATION", "DATA")
- **FR-009**: When `--fix` flag is provided, system MUST attempt to repair fixable problems
- **FR-010**: When `--dry-run` flag is combined with `--fix`, system MUST show what would be fixed without making changes
- **FR-011**: System MUST return appropriate exit codes: 0 for all pass, 1 for warnings, 2 for errors
- **FR-012**: For each failed check, system MUST display actionable guidance on how to resolve it

### Key Entities

- **HealthCheck**: Represents a single diagnostic check with name, status (pass/warn/fail), message, and optional fix action
- **CheckCategory**: Groups related health checks (e.g., Installation, Data, Permissions)
- **FixAction**: Represents an auto-repair operation with description of what will be changed

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `smoke doctor` completes and displays results in under 1 second for a healthy installation
- **SC-002**: `smoke doctor --fix` successfully repairs 100% of documented fixable issues
- **SC-003**: Output format matches `bd doctor` visual style (checkmarks, indentation, categories)
- **SC-004**: All error messages include specific actionable guidance (e.g., "Run 'smoke init' to fix")
- **SC-005**: Agent can diagnose and fix common issues without human intervention in 90% of cases

## Assumptions

- The `bd doctor` visual format (with ✓/⚠/✗ indicators, categories, and indented details) is the target style
- Fixable issues are limited to: missing initialization, missing config file, missing feed file
- Non-fixable issues (like permission problems) will show guidance but require manual intervention
- The command follows existing smoke CLI patterns (Cobra framework, consistent flag naming)
