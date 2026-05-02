# Spec Draft: Foundation & Init

## Summary

The foundational layer of DOFS: project scaffolding, configuration resolution, data model, SQLite database with migrations, and the `dofs init` command that bootstraps a machine by auto-importing existing symlinks.

## User Story

As a developer with an existing DOFS directory tree and ~30 symlinks already pointing into it, I want to run a single command that discovers and registers all of them so I can start managing them without manual re-entry.

## Scope

### Config Resolution
- Resolution order: CLI flags > `DOFS_ROOT` env var > default (`~/dofs/data`)
- Global flags: `--root`, `--dry-run`, `--verbose`
- Config struct shared across all commands

### Model / Types
- Leaf package with zero internal imports
- Categories: `home`, `secrets`, `code`, `data`, `desktop`, `documents`, `downloads`, `music`, `pictures`
- Linkable categories: `home`, `secrets`
- Link status enum: `ok`, `missing`, `broken`, `conflict`, `wrong_target`, `unknown`

### SQLite Database
- Location: `<root>/.dofs.db`
- WAL mode for safe concurrent reads
- CGo-free driver (`modernc.org/sqlite`)
- Versioned migrations
- Tables: `schema_version`, `managed_links`, `scan_history`, `scan_ignores`, `backup_history`

### `dofs init`
- Validate root directory exists
- Warn on missing expected category subdirectories
- Create/open DB, run migrations
- Scan `~/` for symlinks pointing into the DOFS root
- Handle absolute symlinks, relative symlinks, and double-hop chains (e.g. `~/.ssh` → `home/.ssh` → `secrets/ssh`)
- Import discovered symlinks into `managed_links` with status `ok`
- Report count of imported links
- Idempotent: safe to run multiple times

## Dependencies

None — this is the foundation.

## Key Design Decisions

- The DB is a cache; the filesystem is the source of truth
- `model/` must remain a leaf package with no internal imports
- Auto-import on init eliminates the cold-start problem for existing setups
- Double-hop symlink resolution uses `filepath.EvalSymlinks` to find the final target

## Acceptance Criteria

- `dofs init` on a machine with existing DOFS symlinks registers all of them
- Running `dofs init` twice does not create duplicates
- Missing root directory produces a clear error
- Missing category subdirectories produce warnings (not errors)
- DB schema is created with all four tables
- Config resolves flags > env > default correctly
