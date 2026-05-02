<!--
Sync Impact Report
==================
Version change: N/A → 1.0.0 (initial ratification)

Modified principles: N/A (initial)

Added sections:
  - 7 Core Principles with testable gates (Code Quality,
    Testing Standards, Safety & Data Integrity, User Experience,
    Performance, Portability, Architecture)
  - Technical Constraints
  - Development Standards (Taskfile, linting, goreleaser)
  - Development Workflow
  - Governance (with priority ordering, conflict resolution,
    key tensions, authority hierarchy)

Removed sections: N/A

Templates requiring updates:
  - .specify/templates/plan-template.md ✅ no update needed
    (Constitution Check section uses dynamic placeholder)
  - .specify/templates/spec-template.md ✅ no update needed
  - .specify/templates/tasks-template.md ✅ no update needed
  - .specify/templates/commands/*.md ✅ no command files exist

Follow-up TODOs: none
-->

# DOFS Constitution

## Core Principles

Principles are listed in priority order. When principles conflict,
higher-numbered principles yield to lower-numbered ones. See
**Conflict Resolution** in the Governance section for guidance.

### I. Safety & Data Integrity

Every operation MUST answer: *"What happens if this fails halfway
through?"* If the answer involves data loss or inconsistent state,
the operation needs a rollback path before it ships.

- The filesystem is always the source of truth; the database is a
  cache that MUST be rebuildable from filesystem state.
- User files MUST NOT be deleted without explicit `--force`.
- A backup MUST be created before overwriting any existing file.
- All mutating operations MUST respect `--dry-run`.
- Operations MUST be idempotent — running the same command twice
  MUST produce the same result without side effects.
- Scan adoption MUST be atomic: if symlink creation fails after a
  file move, the move MUST be rolled back.

**Rationale**: A filesystem management tool that loses data is worse
than no tool at all. Idempotency and dry-run support let users build
trust incrementally.

**Testable gate**: Can every mutating command be run with `--dry-run`
and produce zero filesystem changes?

### II. Code Quality

The codebase enforces its invariants structurally — through the
import graph and type system, not through convention or code review
alone.

- The `model/` package MUST be a leaf package with zero internal
  imports. All cross-package communication MUST use model types.
- Errors MUST be handled explicitly — silent swallowing (`_ = err`)
  is prohibited outside of documented cache-only DB writes.
- Typed errors (e.g., `ErrConflict`, `ErrBrokenLink`) MUST be used
  for programmatic error handling. String-wrapped `fmt.Errorf` is
  acceptable only for human-facing messages that are not matched on.
- The CGo-free constraint MUST be maintained: use `modernc.org/sqlite`
  exclusively to ensure single-binary cross-compilation.

**Rationale**: Strict dependency direction prevents import cycles and
keeps the model portable. Typed errors enable callers to make decisions
without brittle string matching. CGo-free means the binary deploys
anywhere without a C toolchain.

**Testable gate**: Can `model/` compile with zero imports from
`internal/`?

### III. Testing Standards

If it touches the filesystem, it gets tested against a real
filesystem. No mocks, no fakes, no in-memory substitutes for
symlinks and file operations.

- Every feature specification MUST include acceptance criteria that
  map directly to testable scenarios.
- Priority test areas MUST include:
  - Symlink resolution (absolute, relative, double-hop chains)
  - Filesystem state detection (all 5 link statuses)
  - Archive round-trips (backup, restore, verify)
  - Zip-slip protection on all extracted paths
- Integration tests MUST use temporary directories with real
  filesystem operations — mocking symlinks is prohibited.
- CI MUST run tests on both Linux and macOS.

**Rationale**: Symlinks and filesystem state are the core domain;
incorrect handling causes data loss. Real filesystem tests catch
platform-specific bugs that mocks hide.

**Testable gate**: Do integration tests run against a real temporary
filesystem with no mocked symlinks?

### IV. Architecture

The CLI is a thin command layer over business logic. Commands parse
flags and call functions — they do not contain logic.

- CLI MUST use Cobra with clean separation between command definitions
  (`cmd/`) and business logic (`internal/`).
- Configuration resolution MUST follow the precedence order:
  flags > environment variables > defaults.
- Each command MUST either read state or mutate it — never both
  implicitly in the same operation.
- New features MUST map cleanly to a single subcommand.

**Rationale**: Separation of concerns keeps the CLI layer thin and
testable. Explicit read-or-write semantics prevent surprising side
effects.

**Testable gate**: Does each subcommand either read state or mutate
it — never both implicitly?

### V. User Experience

A user should never need to inspect the filesystem to understand
what a command did. The output tells the full story.

- CLI output MUST be concise and scannable: table format for status
  output, clear counts for mutating operations, actionable error
  messages that state what went wrong and how to fix it.
- Interactive prompts (e.g., `scan --adopt`) MUST provide a quit
  option and MUST NOT leave the filesystem in a partial state.
- The tool MUST feel fast for the common case (~30 managed symlinks)
  and MUST provide progress feedback for slow operations (large
  backups).

**Rationale**: Developers abandon tools that produce noisy output or
feel sluggish. Filesystem operations that can be interrupted must
leave a clean state.

**Testable gate**: Can a user read a command's output and know what
happened without checking the filesystem manually?

### VI. Performance

The tool MUST be fast enough that running it habitually costs
nothing. If users hesitate to run a command because it might be
slow, the tool has failed.

- `init` and link verification MUST complete in under 1 second for
  typical setups (~30 symlinks).
- Backup exclusions (node_modules, vendor, .venv, etc.) MUST use
  `fs.SkipDir` to avoid walking large dependency trees.
- SQLite MUST use WAL mode for safe concurrent reads.

**Rationale**: Sub-second operations encourage frequent use. SkipDir
prevents multi-minute waits on repos with deep dependency trees.

**Testable gate**: Does `dofs init && dofs link` complete in under
1 second on a typical setup?

### VII. Portability

DOFS works anywhere Linux runs, and also on macOS. No
platform-specific code paths, no distro assumptions, no mount point
conventions.

- Linux is the primary target; macOS MUST be compatible. Windows is
  explicitly out of scope.
- The tool MUST work across distros with different package managers,
  init systems, and filesystem layouts without distro-specific code
  paths.
- The DOFS root MUST be allowed on any mounted filesystem — no
  assumptions about mount points beyond the configured root path.

**Rationale**: The whole point of DOFS is surviving distro hops and
dual boots. Distro-specific logic defeats that purpose.

**Testable gate**: Does the binary build and pass all tests on both
Linux and macOS without platform-specific code paths?

## Technical Constraints

- **Language**: Go (latest stable).
- **CLI framework**: Cobra.
- **Database**: SQLite via `modernc.org/sqlite` (CGo-free).
- **Archive format**: tar.gz via standard library (`archive/tar` +
  `compress/gzip`).
- **Build target**: Single static binary, no runtime dependencies.
- **Platform targets**: Linux (primary), macOS (secondary). No Windows.

## Development Standards

These standards are derived from the repo's existing tooling
configuration, not invented independently.

**Language and structure**:

- Go is the implementation language. All application code lives in
  `internal/` and `cmd/`. The `model/` leaf-package constraint
  (Principle II) is structural.
- Entry point is `cmd/dofs/`. Binary name is `dofs`.

**Task runner**: All development workflows use
[Taskfile](https://taskfile.dev). The canonical commands are:

| Command | Purpose |
|---------|---------|
| `task build` | Build binary to `dist/` |
| `task test` | Run tests |
| `task test:coverage` | Run tests with coverage report |
| `task lint` | Run golangci-lint |
| `task lint:fix` | Run linters and auto-fix |
| `task fmt` | Format code (gofmt) |
| `task vet` | Run go vet |
| `task tidy` | go mod tidy |
| `task check` | Full check: lint + vet + test |
| `task clean` | Remove build artifacts |

**Quality enforcement**: golangci-lint for static analysis. CI MUST
pass linting and tests before merge.

**Releases**: goreleaser handles cross-platform binary builds and
distribution. Snapshot builds via `task release:snapshot`.

**License**: Apache-2.0.

## Development Workflow

- Feature branches MUST follow the `###-feature-name` naming
  convention and be specified through the speckit workflow before
  implementation begins.
- All PRs MUST pass CI (lint, test on Linux + macOS) before merge.
- The `model/` leaf-package constraint MUST be verified in code
  review — any import of an `internal/` package from `model/` is a
  blocking violation.
- `--dry-run` support MUST be verified for every new mutating command
  before the PR is approved.

## Governance

**Authority hierarchy**:

1. This constitution governs development principles, standards, and
   architectural boundaries.
2. Feature specifications (`specs/`) define what each feature does.
   The implementation fulfills the spec; it does not extend or
   contradict it.
3. Documentation (`docs/`) captures current behavior and user-facing
   guidance.

**Principle priority and conflict resolution**: Principles are
ordered I through VII by priority. When two principles conflict:

- Apply the higher-priority principle.
- Document the conflict and the resolution in the relevant PR
  description.
- Common tensions and their resolution:
  - *Safety (I) vs Performance (VI)*: If a performance optimization
    skips a safety check (e.g., omitting backup-before-overwrite),
    safety wins — users trust the tool because it protects their data.
  - *Safety (I) vs UX (V)*: If a safety measure adds friction (e.g.,
    requiring `--force` for destructive operations), safety wins —
    the friction is the feature.
  - *Code Quality (II) vs Architecture (IV)*: If an architectural
    shortcut would violate the leaf-package constraint or introduce
    CGo, code quality wins — the constraint is structural.
  - *Testing (III) vs Performance (VI)*: If testing real filesystem
    operations is slow, testing wins — correctness over speed in the
    test suite.
  - *Architecture (IV) vs UX (V)*: If a UX improvement requires a
    command that both reads and mutates implicitly, architecture
    wins — split it into two explicit operations.

**Key tensions** (acknowledged, not resolved — these are ongoing):

- Safety vs speed: backup-before-overwrite and atomic adoption add
  I/O overhead. The answer is "make it fast enough", not "skip it".
- Portability vs platform features: macOS symlink behavior differs
  from Linux in subtle ways. The answer is real cross-platform tests,
  not ifdef-style code.
- Database-as-cache vs audit trail: the DB stores scan history and
  backup records that are not rebuildable from filesystem state. This
  is an accepted exception — the audit data is useful but not critical.

**Amendments**: Changes to this constitution require:

1. A written proposal describing the change and rationale.
2. An update to this file with version bump per semver:
   - MAJOR: principle removal or backward-incompatible redefinition.
   - MINOR: new principle, new section, or material expansion.
   - PATCH: wording clarification, typo fix, non-semantic refinement.
3. A sync impact review of all dependent templates.

**Compliance**: All PRs and code reviews MUST verify alignment with
these principles. Reviewers SHOULD reference specific principle
numbers (e.g., "Principle I violation: missing --dry-run support").
The constitution is a living document — if a principle consistently
creates friction, that is signal to amend it, not ignore it.

**Version**: 1.0.0 | **Ratified**: 2026-03-08 | **Last Amended**: 2026-03-08
