# Spec Draft: Backup

## Summary

Create per-category tar.gz archives of the DOFS tree, with smart exclusions for build artifacts and dependency directories. Tracks backup history in the database.

## User Stories

- As a developer, I want to back up my entire DOFS tree (or specific categories) as portable archives for migration or disaster recovery.
- As a developer, I want build artifacts like `node_modules` and `.venv` automatically excluded so backups are fast and small.
- As a developer, I want to preview what would be archived before committing to a potentially large backup.

## Scope

### `dofs backup [category] -o PATH`
- If category specified: archive only that category
- If no category: archive all categories
- Output: `dofs-backup-<category>-YYYYMMDD-HHMMSS.tar.gz` per category

### Archive Format
- tar.gz using standard library (`archive/tar` + `compress/gzip`)
- Contains relative paths from category root (no parent directory wrapper)
- Symlinks preserved as symlink entries (not followed)
- Sockets, devices, named pipes skipped

### Exclusion System
- Directory-name-based exclusion (not path-based)
- Any directory matching an excluded name, anywhere in the tree, is skipped via `fs.SkipDir`
- Default exclusions: `node_modules`, `.next`, `.nuxt`, `.parcel-cache`, `bower_components`, `vendor`, `.venv`, `venv`, `__pycache__`, `.tox`, `.eggs`, `.mypy_cache`, `.pytest_cache`, `.gradle`, `.maven`, `target`, `dist`, `build`, `.cache`
- `--include-all` flag disables all exclusions

### Backup History
- Insert `backup_history` record at start (status: `in_progress`)
- Update on completion (status: `completed`, `size_bytes`, `completed_at`)
- Update on failure (status: `failed`)

### Dry-run
- Shows what categories/files would be archived
- Shows estimated file count
- Does NOT create any archives or DB records

## Dependencies

- Spec 1 (Foundation & Init) — config, DB, model

## Key Design Decisions

- Per-category archives (not one giant archive) for flexibility on restore
- Contents-only format (no root directory) so restore can target any directory structure
- Category name embedded in filename so restore can parse it back
- Exclusion by directory name (not path) keeps the system simple and covers nested occurrences

## Acceptance Criteria

- `dofs backup home -o /tmp/` creates a valid tar.gz with correct contents
- Archive contains relative paths, no category root wrapper
- Excluded directories are not present in the archive
- `--include-all` includes everything
- Symlinks are preserved as symlink entries
- Sockets/devices/pipes are skipped without error
- `backup_history` records are created and updated correctly
- `--dry-run` does not create archives or DB records
- Backup of nonexistent category produces a clear error
