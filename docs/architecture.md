# DOFS Architecture

## Package Dependency Graph

```
main.go
  └── cmd/
       ├── root.go ─────────── config, db
       ├── init.go ─────────── linker, model
       ├── link.go ─────────── linker, model
       ├── backup.go ───────── backup, model
       ├── restore.go ──────── linker, model (+ archive/tar, compress/gzip)
       ├── scan.go ─────────── scanner, model
       └── status.go ───────── status, model
            │
            ▼
       internal/
       ├── config/ ─────────── (no internal deps)
       ├── model/ ──────────── (no internal deps, leaf package)
       ├── db/ ─────────────── model
       ├── linker/ ─────────── config, db, model
       ├── backup/ ─────────── config, db, model
       ├── scanner/ ────────── config, db, model
       └── status/ ─────────── config, db, linker, model
```

Key constraint: `model/` is a leaf package with no internal imports. All cross-package communication goes through model types.

## Data Flow

### Init Flow
```
User runs: dofs init --root ~/dofs/data

1. Validate root directory exists
2. Check expected category subdirectories (warn if missing)
3. Create/open SQLite DB at <root>/.dofs.db
4. Run schema migrations
5. Scan ~/  for symlinks
   For each entry in ~/:
     - Is it a symlink? (os.Lstat → ModeSymlink)
     - Does it point into DOFS root? (filepath.Rel check)
     - Extract category from relative path
     - Insert into managed_links with status=ok
6. Report count
```

### Link Check Flow
```
User runs: dofs link

For each managed_link in DB:
  1. os.Lstat(target_path)
     → NotExist: status=missing
     → Not symlink: status=conflict
  2. os.Readlink(target_path) → actual target
  3. os.Stat(target_path) to check if target exists
     → NotExist: status=broken (dangling symlink)
  4. Resolve actual vs expected paths (handle relative symlinks)
     → Match: status=ok
     → Mismatch: status=wrong_target
  5. Update last_status in DB
```

### Backup Flow
```
User runs: dofs backup code -o /tmp/

1. Resolve source: <root>/code/
2. Generate filename: dofs-backup-code-YYYYMMDD-HHMMSS.tar.gz
3. Insert backup_history (status=in_progress)
4. Create file → gzip.Writer → tar.Writer
5. filepath.WalkDir(source):
   - Directory in exclusion list? → fs.SkipDir
   - Socket/device/pipe? → skip
   - Symlink? → add as symlink entry
   - Regular file? → stream into tar
6. Close writers, stat output file
7. Update backup_history (status=completed, size_bytes)
```

### Restore Flow
```
User runs: dofs restore --link /tmp/dofs-backup-*.tar.gz

For each archive:
  1. Parse category from filename
  2. Create <root>/<category>/ directory
  3. Extract archive contents into it
     - Regular files: create with original permissions
     - Symlinks: recreate
     - Directories: mkdir with original permissions
     - Zip-slip protection on all paths

After extraction:
  4. Create/open SQLite DB
  5. Walk linkable categories (home/, secrets/)
     For each item found:
       - Derive target: ~/.<name>
       - Insert into managed_links
  6. If --link: create all missing symlinks
```

### Scan Flow
```
User runs: dofs scan --adopt

1. Read ~/ entries (depth 1)
2. Filter out:
   - Non-dotfiles (no leading ".")
   - Already tracked in managed_links
   - Builtin ignore list (system files)
   - User ignore patterns from scan_ignores table
   - Symlinks already pointing into DOFS root
3. Present candidates to user
4. For each, prompt: [y/n/i(gnore)/q(uit)]
   - y: os.Rename(~/.<item>, <root>/home/<item>)
         os.Symlink(<root>/home/<item>, ~/.<item>)
         Insert managed_link
         Insert scan_history (adopted)
   - i: Insert scan_ignores pattern
         Insert scan_history (ignored)
   - q: stop
```

## SQLite Schema

```sql
-- Version tracking
schema_version (version INTEGER, applied_at TEXT)

-- Core: tracked symlinks
managed_links (
  id, category, source_name, source_path, target_path UNIQUE,
  last_status DEFAULT 'unknown', created_at, updated_at, last_checked_at
  UNIQUE(category, source_name)
)

-- Scan audit trail
scan_history (id, found_path, disposition, adopted_category, scanned_at, resolved_at)

-- Persistent ignore patterns
scan_ignores (id, pattern UNIQUE, created_at)

-- Backup records
backup_history (id, category, archive_path, size_bytes, started_at, completed_at, status)
```

Migrations are versioned — each migration is an entry in the `migrations` slice in `db/migrations.go`. The current version is applied, and future migrations append to the slice.

## CLI Flag Architecture

Global flags are defined on `rootCmd` in `cmd/root.go` and stored in package-level vars. Two helper functions manage lifecycle:

- `loadConfig()` — resolves flags/env/defaults into a `*config.Config`
- `openStore()` — calls `loadConfig()` then opens the DB
- `openStoreWithConfig()` — opens the DB using an already-loaded config (used by `init` which needs to validate the root before opening the DB)
- `closeStore()` — closes the DB connection

Each subcommand either calls `openStore()` (most commands) or `loadConfig()` + `openStoreWithConfig()` (init, which needs to check the root first).

## Symlink Resolution

The trickiest part of the codebase. Three cases:

1. **Absolute symlink**: `~/.zshrc → /home/user/dofs/data/home/.zshrc`
   - Direct path comparison after `filepath.Abs()`

2. **Relative symlink**: `~/.zshrc → dofs/data/home/.zshrc`
   - Must resolve: `filepath.Abs(filepath.Join(filepath.Dir(target), readlink_result))`
   - Only joins when `!filepath.IsAbs(readlink_result)` — Go's `filepath.Join` does NOT handle absolute second args like Python's `os.path.join`

3. **Double-hop**: `~/.ssh → data/home/.ssh → data/secrets/ssh`
   - Import resolves the full chain via `filepath.EvalSymlinks`
   - Category is determined from the final resolved path

## Backup Exclusions

Exclusion is directory-name-based, not path-based. Any directory anywhere in the tree with a matching name gets skipped via `fs.SkipDir`. The exclusion list covers:

- **JS/Node**: node_modules, .next, .nuxt, .parcel-cache, bower_components
- **PHP**: vendor
- **Python**: .venv, venv, __pycache__, .tox, .eggs, .mypy_cache, .pytest_cache
- **JVM**: .gradle, .maven
- **Rust**: target
- **General**: dist, build, .cache

The `--include-all` flag disables all exclusions.

## Archive Format

- **Format**: tar.gz (standard library `archive/tar` + `compress/gzip`)
- **Contents**: Relative paths from category root (no parent directory in archive)
- **Naming**: `dofs-backup-<category>-YYYYMMDD-HHMMSS.tar.gz`
- **Restore**: Category parsed from filename, contents extracted directly into `<root>/<category>/`
- **Special files**: Sockets, devices, named pipes are skipped during archival
- **Symlinks**: Preserved as symlink entries in the tar (not followed)
