# Implementation Plan: Hooks Installation System

**Branch**: `001-hooks-install` | **Date**: 2026-01-31 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `.specify/specs/001-hooks-install/spec.md`

## Summary

Implement automatic Claude Code hook installation as part of `smoke init`. This adds three new CLI commands (`smoke hooks install|uninstall|status`) and integrates hook installation into the existing init flow. Hook scripts are embedded in the binary using Go's embed directive.

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**: cobra (CLI), embed (assets), encoding/json (settings)
**Storage**: ~/.claude/settings.json (JSON), ~/.claude/hooks/*.sh (scripts)
**Testing**: go test with testify assertions
**Target Platform**: macOS, Linux (bash required for hooks)
**Project Type**: Single CLI application
**Constraints**:
- Must not break existing Claude Code hooks (merge, not replace)
- Hook scripts <10KB each (embed-friendly)
- Zero additional runtime dependencies

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Requirement | Status |
|-----------|-------------|--------|
| I. Go Simplicity | Standard library preferred | PASS - using embed, encoding/json |
| II. Agent-First CLI | Zero config, discoverable | PASS - hooks install automatically |
| III. Local-First Storage | No external services | PASS - local files only |
| IV. Test What Matters | CLI integration tests | PASS - testing install/uninstall flow |
| V. Environment Integration | BD_ACTOR identity | N/A - hooks don't use identity |
| VI. Minimal Configuration | Sensible defaults | PASS - auto-install during init |

**Gate violations**: None

## Project Structure

### Documentation (this feature)

```text
.specify/specs/001-hooks-install/
├── plan.md              # This file
├── research.md          # Research findings
├── data-model.md        # Entity definitions
├── quickstart.md        # Implementation guide
└── contracts/
    ├── cli-interface.md # Command contracts
    └── internal-api.md  # Package API contracts
```

### Source Code (repository root)

```text
cmd/smoke/
└── main.go              # Entry point (no changes)

internal/
├── cli/
│   ├── hooks.go         # NEW: smoke hooks subcommands
│   ├── hooks_test.go    # NEW: CLI tests
│   ├── init.go          # MODIFY: add hook installation
│   └── init_test.go     # MODIFY: test hook integration
│
├── hooks/               # NEW: entire package
│   ├── types.go         # HookEvent, ScriptStatus, InstallState, Status
│   ├── paths.go         # GetHooksDir(), GetSettingsPath()
│   ├── scripts.go       # GetScriptContent(), ListScripts(), hash comparison
│   ├── settings.go      # settings.json read/write, merge logic
│   ├── errors.go        # ErrScriptsModified, ErrPermissionDenied, ErrInvalidSettings
│   ├── hooks.go         # Install, Uninstall, GetStatus
│   ├── embed.go         # Embed directive and asset access
│   ├── hooks_test.go    # Unit tests
│   └── scripts/
│       ├── smoke-break.sh   # Stop hook
│       └── smoke-nudge.sh   # PostToolUse hook

tests/integration/
└── hooks_test.go        # NEW: End-to-end hook tests
```

**Structure Decision**: Single project structure. The new `internal/hooks/` package handles hook logic, `internal/cli/hooks.go` provides commands.

## Implementation Phases

### Phase 1: Core hooks package (internal/hooks/)

Create the hooks package with embedded scripts and core operations.

**Tasks**:
1. Create `internal/hooks/scripts/` with smoke-break.sh and smoke-nudge.sh
2. Create `internal/hooks/embed.go` with embed directives
3. Create `internal/hooks/hooks.go` with Install(), Uninstall(), GetStatus()
4. Create `internal/hooks/settings.go` for settings.json manipulation
5. Create `internal/hooks/hooks_test.go` with comprehensive unit tests

**Acceptance**:
- `go test ./internal/hooks/...` passes
- Scripts are accessible via embed.FS
- Install/Uninstall/Status functions work in isolation

### Phase 2: CLI commands (internal/cli/hooks.go)

Add the `smoke hooks` command group.

**Tasks**:
1. Create `internal/cli/hooks.go` with parent command and subcommands
2. Implement `smoke hooks install [--force]`
3. Implement `smoke hooks uninstall`
4. Implement `smoke hooks status [--json]`
5. Create `internal/cli/hooks_test.go`

**Acceptance**:
- `smoke hooks install` installs hooks
- `smoke hooks uninstall` removes hooks
- `smoke hooks status` reports accurate state
- Commands follow existing CLI patterns

### Phase 3: Init integration (internal/cli/init.go)

Integrate hook installation into `smoke init`.

**Tasks**:
1. Modify `init.go` to call hooks.Install() after smoke setup
2. Handle hook errors gracefully (warn, don't fail per FR-002)
3. Update output to include hook status
4. Add check for "already initialized but hooks missing"
5. Update `init_test.go` with hook integration tests

**Acceptance**:
- `smoke init` installs hooks automatically
- Init succeeds even if hooks fail (with warning)
- "Already initialized" suggests `smoke hooks install` if hooks missing

### Phase 4: Integration testing

End-to-end tests for the complete flow.

**Tasks**:
1. Test fresh init with hooks
2. Test init when hooks already exist
3. Test install/uninstall cycle
4. Test status in all states
5. Test --force flag for modified scripts

**Acceptance**:
- All integration tests pass
- Coverage meets project threshold (70%+)
- `make ci` passes

## Risk Mitigation

| Risk | Mitigation |
|------|-----------|
| Break existing Claude Code hooks | Merge strategy: detect existing, add smoke entries only |
| Invalid settings.json | Backup before modify, handle parse errors |
| Permission issues | Graceful degradation with clear error messages |
| Script hash false positives | Use SHA256, generous comparison |

## Dependencies Between Phases

```
Phase 1 (hooks package)
    │
    ├──▶ Phase 2 (CLI commands)
    │
    └──▶ Phase 3 (init integration)
              │
              └──▶ Phase 4 (integration tests)
```

Phase 2 and 3 can proceed in parallel after Phase 1 completes.
