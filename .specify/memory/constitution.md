# Smoke Constitution

## Core Principles

### I. Go Simplicity

All source code MUST be written in idiomatic Go. Embrace Go's philosophy
of simplicity and readability.

- Standard library preferred over external dependencies
- Minimal dependencies: only add when clear value justifies maintenance cost
- Error handling: explicit, wrapped with context, never silently ignored
- No reflection or code generation unless absolutely necessary
- Follow effective Go and Go code review comments guidelines

**Rationale:** Smoke is a small CLI tool. Keeping dependencies minimal
ensures fast builds, easy maintenance, and predictable behavior. Go's
standard library is battle-tested and sufficient for most needs.

### II. CLI-First Design

Smoke is a command-line tool. Every feature MUST work well in terminal
environments and compose with other Unix tools.

- Output designed for human readability by default
- Machine-readable output via flags (--json, --oneline, --quiet)
- Graceful degradation when piped/redirected (no ANSI when not TTY)
- Exit codes meaningful: 0 success, 1 user error, 2 system error
- Respect standard Unix conventions (stdin, stdout, stderr)

**Rationale:** Gas Town agents use smoke in scripts and interactive
terminals. The tool must work seamlessly in both contexts without
requiring special handling.

### III. Local-First Storage

Smoke stores data locally in the Gas Town directory. No external services,
no network dependencies for core functionality.

- JSONL format for feed storage (append-only, line-based)
- Human-readable data files (can be inspected/edited manually)
- No database dependencies
- Atomic writes where possible (write temp, rename)
- Respect filesystem permissions

**Rationale:** Smoke runs in diverse environments across Gas Town rigs.
Local storage ensures reliability without network dependencies. JSONL
format balances simplicity with append performance.

### IV. Test What Matters

Tests MUST focus on user-visible behavior and edge cases. No test bloat.

- Integration tests for CLI commands (actual binary execution)
- Unit tests for parsing, validation, ID generation
- Table-driven tests for input variations
- Test error paths, not just happy paths
- Coverage target: 50%+ (integration tests count toward user value)

**Rationale:** Smoke is a small tool with clear inputs and outputs.
Testing the CLI end-to-end catches more real bugs than mocking internals.
Coverage exists to find gaps, not as a vanity metric.

### V. Gas Town Integration

Smoke MUST integrate cleanly with Gas Town infrastructure.

- Identity from BD_ACTOR environment variable (fallback SMOKE_AUTHOR)
- Detect Gas Town root via .beads or mayor directory markers
- Store data in .smoke/ directory at Gas Town root
- Post IDs use smk- prefix for namespace clarity

**Rationale:** Smoke is part of the Gas Town ecosystem. Consistent
conventions make it discoverable and predictable for other tools
and agents.

### VI. Minimal Configuration

Smoke SHOULD work with zero configuration. Sensible defaults over options.

- Init creates necessary structure automatically
- Identity from environment (no config file for author)
- No user preferences file (feature flags via CLI only)
- Colors/formatting auto-detected from terminal

**Rationale:** Every configuration option is a decision users must make
and maintain. Smoke should "just work" for Gas Town agents without
requiring setup beyond `smoke init`.

## Architecture Constraints

- **Language:** Go 1.22+
- **CLI framework:** Cobra for commands, standard flag parsing
- **Storage:** JSONL at `<gas-town-root>/.smoke/feed.jsonl`
- **Structure:** cmd/smoke (entry), internal/cli (commands), internal/feed (domain), internal/config (Gas Town detection)
- **Testing:** Go testing + integration tests via compiled binary
- **Build:** Standard `go build`, goreleaser for releases
- **CI:** GitHub Actions, golangci-lint for linting

## Development Workflow Standards

- **Branching:** `main` (production), feature branches for development
- **Commits:** `type: description`. Types: `feat`, `fix`, `chore`, `refactor`, `docs`, `test`, `ci`
- **PRs:** Draft by default, require CI pass before merge
- **Quality gates:** `go vet`, `golangci-lint`, `go test -race`
- **Releases:** Release-please for changelogs, goreleaser for binaries, Homebrew tap for distribution

## Governance

This constitution guides Smoke development. Amendments via PR with rationale.

**Version**: 1.0.0 | **Created**: 2026-01-30
