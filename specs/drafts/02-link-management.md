# Spec Draft: Link Verification & Management

## Summary

The core symlink lifecycle: verify tracked links against the filesystem, create missing symlinks, repair broken ones, register new items, and stop tracking items. This is the primary value loop of DOFS.

## User Stories

- As a developer, I want to see the status of all my tracked symlinks at a glance so I can spot problems.
- As a developer, I want to fix broken or missing symlinks with one command after an OS reinstall.
- As a developer, I want to register a new config file to track and create its symlink.
- As a developer, I want to stop tracking a link without deleting my files.

## Scope

### `dofs link` (read-only verification)
- For each record in `managed_links`:
    - `os.Lstat(target_path)` — does anything exist?
    - `os.Readlink(target_path)` — what does the symlink point to?
    - `os.Stat(target_path)` — does the symlink target actually exist?
    - Resolve and compare actual vs expected paths
- Status determination:
    - `ok` — symlink exists, points to correct target
    - `missing` — nothing exists at target path
    - `broken` — symlink exists but target doesn't (dangling)
    - `conflict` — regular file/directory exists where symlink should be
    - `wrong_target` — symlink points somewhere unexpected
- Table output with name, category, status, paths
- Update `last_status` and `last_checked_at` in DB

### `dofs link --create`
- Create symlinks for all items with status `missing`
- Ensure parent directories exist
- Idempotent

### `dofs link --repair`
- Fix items with status `broken` or `wrong_target`
- Remove the incorrect symlink and recreate it

### `dofs link --add <name>`
- Register a new item in `managed_links`
- Default category: `home`
- `--category` flag to specify `home` or `secrets`
- `--target` flag for custom target paths (name mappings like `gpg` → `~/.gnupg`)
- Source path derived from: `<root>/<category>/<name>`
- Target path derived from: `~/<name>` (or `--target` value)

### `dofs link --remove <name>`
- Remove record from `managed_links`
- Does NOT delete the symlink or any files (safety first)

### `dofs link --force`
- Combined with `--create`: overwrite conflicts
- Backs up the existing file to `<target>.dofs-backup.<timestamp>` before overwriting
- Never destroys user data without a backup

### Symlink Resolution
- Absolute symlinks: direct path comparison after `filepath.Abs()`
- Relative symlinks: resolve against the directory containing the symlink
- All comparisons after normalization

### Dry-run
- All write operations (`--create`, `--repair`, `--add`, `--remove`, `--force`) respect `--dry-run`

## Dependencies

- Spec 1 (Foundation & Init) — needs config, DB, model types

## Key Design Decisions

- Read-only verification is the default (no flags) — safe to run anytime
- `--remove` only stops tracking; it never deletes files or symlinks
- `--force` always creates a backup before overwriting — zero data loss
- Status is always re-derived from the filesystem, never cached

## Acceptance Criteria

- `dofs link` correctly identifies all 5 status states
- `dofs link --create` creates symlinks only for `missing` items
- `dofs link --repair` fixes `broken` and `wrong_target` items
- `dofs link --add` with `--target` correctly handles name mappings
- `dofs link --remove` deletes the DB record but leaves filesystem untouched
- `--force` creates a backup file before overwriting a conflict
- `--dry-run` on all write operations shows what would happen without doing it
- Relative and absolute symlinks both resolve correctly
