# Known Issues & Future Work

## Bugs / Rough Edges

### 1. Dry-run leaks backup records into DB
`dofs backup --dry-run` still inserts a `backup_history` record with status "dry-run". Over time these accumulate. Fix: check `DryRun` before `InsertBackup`.

### 2. Silent error swallowing in several places
Multiple `_ = store.UpdateLinkStatus(...)` and `_ = store.CompleteBackup(...)` calls. The DB is treated as a cache so this is mostly fine, but verbose mode should log these failures.

### 3. No progress reporting for large backups
Backing up `code/` with many repos gives zero output during what can be a multi-minute operation. Should add a file counter or progress bar.

### 4. Restore doesn't handle re-runs gracefully
Running `dofs restore` twice on the same archives will fail on `InsertLink` due to UNIQUE constraint on `(category, source_name)`. Should use INSERT OR IGNORE or check for existing records.

### 5. `dofs link --add` with `--create` only creates the just-added link
The `--create` path after `--add` passes the category filter, which works, but if other missing links exist in the same category they also get created. This might surprise users. Consider creating only the explicitly added link.

## Missing Features

### Name Mappings
Items where the DOFS name differs from the OS dotfile name (e.g. `secrets/gpg` → `~/.gnupg`, `secrets/ssh` → `~/.ssh`) currently require `--target` on every `--add`. A config file or DB table for persistent name mappings would eliminate this.

### Config File
A `.dofs.yaml` at the DOFS root could declare:
- Name mappings
- Custom exclusion patterns for backup
- Custom category definitions
- Default backup output directory

### Tests
No test files exist. Priority areas for testing:
- `linker.CheckOne()` — various symlink states
- `categoryFromFilename()` — archive name parsing
- `splitPath()` — path splitting utility
- `extractArchive()` — zip-slip protection
- Integration test: backup → restore → verify round-trip

### Verbose/Quiet Consistency
Not all code paths respect `--verbose`. Some operations always print, some never do. A logging abstraction (even just a helper function checking `cfg.Verbose`) would help.

### Shell Completion
Cobra provides `dofs completion bash/zsh/fish` out of the box, but category names for `backup`, `--category` flags etc. could have custom completions registered.

### Periodic/Scheduled Runs
The original vision mentioned "periodically run and auto-symlink". This could be a `dofs watch` or `dofs cron` command, or simply documented as a cron job: `*/30 * * * * dofs link --create --quiet`.

### Link Groups / Profiles
For multi-machine setups (work laptop vs personal laptop), being able to tag links with profiles and only activate certain ones would be useful.

## Architecture Improvements

### Transaction Safety
Scan adoption does: move file → create symlink → insert DB record. If any step fails mid-way, the state is inconsistent. Wrapping the DB operations in a transaction and using a two-phase approach (record intent → execute → confirm) would be safer.

### Concurrency
Multiple `dofs` invocations could race on the DB. SQLite handles basic locking, but long-running backups hold the connection open. Consider using shorter transactions or connection pooling.

### Error Types
Currently all errors are `fmt.Errorf` wrapped strings. Typed errors (e.g. `ErrConflict`, `ErrBrokenLink`) would allow callers to handle specific cases programmatically.
