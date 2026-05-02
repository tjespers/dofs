# CLAUDE.md — DOFS Contributor Guide

## What is this project?

DOFS (Developer Optimized File System) is a Go CLI tool that manages a developer's workstation files through a structured directory layout. It automates symlinking dotfiles, configs, and secrets from a canonical tree (`~/dofs/data/`) to their expected OS locations (`~/`), and tracks everything in a local SQLite database.

## Build & Run

```bash
# Build
go build -o dofs .

# Or using Taskfile
task build          # Binary goes to dist/dofs

# Run tests
task test

# Lint
task lint

# All checks
task check          # lint + vet + test
```

## Project Structure

```
dofs/
├── main.go                     # Entry point — calls cmd.Execute()
├── cmd/                        # Cobra command definitions
│   ├── root.go                 # Root command, global flags, config/store lifecycle
│   ├── init.go                 # dofs init
│   ├── link.go                 # dofs link
│   ├── backup.go               # dofs backup
│   ├── restore.go              # dofs restore
│   ├── scan.go                 # dofs scan
│   └── status.go               # dofs status
├── internal/
│   ├── config/config.go        # Config resolution (flags > env > defaults)
│   ├── model/model.go          # Shared types and constants (leaf package, no internal deps)
│   ├── db/
│   │   ├── db.go               # SQLite store (CRUD for all tables)
│   │   └── migrations.go       # Versioned schema migrations
│   ├── linker/linker.go        # Symlink verify/create/repair/import logic
│   ├── backup/
│   │   ├── backup.go           # Archive creation with fs.WalkDir
│   │   └── excludes.go         # Default exclusion directory list
│   ├── scanner/scanner.go      # Home dir scanning + adoption workflow
│   └── status/status.go        # Health check aggregation
└── docs/                       # MkDocs documentation site
```

## Key Architecture Rules

- **`model/` is a leaf package** — it has zero internal imports. All cross-package communication uses model types.
- **The DB is a cache, the filesystem is truth** — link status is always re-verified against the actual filesystem, never trusted from the DB alone.
- **Safety first** — never delete user files. Conflicts require `--force` which backs up existing files first. All mutating operations respect `--dry-run`.
- **Idempotent operations** — running `dofs init` or `dofs link --create` multiple times is safe.

## Tech Stack

- **Go** with **Cobra** for CLI
- **SQLite** via `modernc.org/sqlite` (CGo-free, pure Go driver)
- **Taskfile** (`task`) for build automation

## Database

SQLite database at `<root>/.dofs.db` with WAL mode. Four tables:

- `managed_links` — tracked symlinks
- `scan_history` — scan audit trail
- `scan_ignores` — persistent ignore patterns
- `backup_history` — backup records

Migrations are versioned in `internal/db/migrations.go`.

## Conventions

- Config resolution order: CLI flags > environment variables > defaults
- DOFS root default: `~/dofs/data/` (override with `--root` or `DOFS_ROOT`)
- Linkable categories: `home` and `secrets` — items here get symlinked to `~/`
- Archive naming: `dofs-backup-<category>-YYYYMMDD-HHMMSS.tar.gz`
- Archives contain category contents only (no root directory wrapper)

## Current State

- No test files exist yet — this is the top priority for contributions
- See `docs/known-issues.md` for bugs and missing features
- See `docs/architecture.md` for data flows and design details

## Documentation

Documentation lives in `docs/` and is structured for MkDocs. Run `mkdocs serve` to preview locally.
