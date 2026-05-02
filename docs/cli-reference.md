# CLI Reference

## Global Flags

These flags are available on all commands:

| Flag | Env Var | Default | Description |
|------|---------|---------|-------------|
| `--root` | `DOFS_ROOT` | `~/dofs/data` | Path to the DOFS root directory |
| `--dry-run` | — | `false` | Show what would happen without making changes |
| `--verbose` | — | `false` | Enable detailed output |

## `dofs init`

Bootstrap the DOFS database and auto-import existing symlinks.

```bash
dofs init [--root PATH]
```

**What it does:**

1. Validates the root directory exists
2. Checks for expected category subdirectories (warns if missing)
3. Creates/opens the SQLite database at `<root>/.dofs.db`
4. Runs schema migrations
5. Scans `~/` for symlinks pointing into the DOFS root
6. Imports found symlinks into the `managed_links` table with status `ok`

**Idempotent**: Safe to run multiple times. Existing records are not duplicated.

## `dofs link`

Manage tracked symlinks.

```bash
# Verify all tracked symlinks
dofs link

# Create missing symlinks
dofs link --create

# Repair broken symlinks
dofs link --repair

# Register a new item to track
dofs link --add <name> [--category home|secrets] [--target PATH]

# Stop tracking a link (does not delete the symlink or files)
dofs link --remove <name>
```

**Link statuses:**

| Status | Meaning |
|--------|---------|
| `ok` | Symlink exists and points to the correct target |
| `missing` | No file/symlink exists at the target path |
| `broken` | Symlink exists but its target doesn't (dangling) |
| `conflict` | A regular file/directory exists where the symlink should be |
| `wrong_target` | Symlink exists but points somewhere else |

**Flags:**

| Flag | Description |
|------|-------------|
| `--create` | Create symlinks for all items with status `missing` |
| `--repair` | Fix items with status `broken` or `wrong_target` |
| `--add <name>` | Register a new item to track |
| `--remove <name>` | Stop tracking an item |
| `--category` | Category for `--add` (default: `home`) |
| `--target` | Custom target path for `--add` (for name mappings like `gpg` → `~/.gnupg`) |
| `--force` | Overwrite conflicts (backs up the existing file first) |

## `dofs status`

Show a full health report.

```bash
dofs status
```

**Output includes:**

- Link summary: counts by status (ok, broken, missing, conflict)
- Untracked items in linkable categories
- Backup ages per category

## `dofs backup`

Create per-category tar.gz archives.

```bash
# Backup all categories
dofs backup -o /path/to/output/

# Backup a specific category
dofs backup home -o /path/to/output/

# Preview without creating archives
dofs backup --dry-run

# Include directories normally excluded (node_modules, vendor, etc.)
dofs backup code --include-all -o /path/to/output/
```

**Archive format:**

- Filename: `dofs-backup-<category>-YYYYMMDD-HHMMSS.tar.gz`
- Contents: relative paths from the category root (no parent directory wrapper)
- Symlinks are preserved as symlink entries (not followed)
- Sockets, devices, and named pipes are skipped

**Default exclusions** (for the `code` category and others with build artifacts):

`node_modules`, `.next`, `.nuxt`, `vendor`, `.venv`, `venv`, `__pycache__`, `.gradle`, `target`, `dist`, `build`, `.cache`, and others. Use `--include-all` to disable.

**Flags:**

| Flag | Description |
|------|-------------|
| `-o, --output` | Output directory for archives |
| `--include-all` | Disable default directory exclusions |

## `dofs restore`

Extract archives and set up DOFS on a new machine.

```bash
# Extract only
dofs restore /path/to/dofs-backup-*.tar.gz

# Extract and create symlinks
dofs restore --link /path/to/dofs-backup-*.tar.gz
```

**What it does:**

1. Parses category from each archive filename
2. Creates `<root>/<category>/` directories
3. Extracts archive contents (with zip-slip protection)
4. Creates/opens the SQLite database
5. Registers items from linkable categories (`home/`, `secrets/`) in the DB
6. If `--link`: creates all missing symlinks

**Flags:**

| Flag | Description |
|------|-------------|
| `--link` | Create symlinks after extraction |

## `dofs scan`

Find adoptable dotfiles in `~/` that aren't tracked by DOFS.

```bash
# List candidates
dofs scan

# Interactive adoption
dofs scan --adopt
```

**Filtering:** The scan excludes:

- Non-dotfiles (no leading `.`)
- Items already tracked in `managed_links`
- Symlinks already pointing into the DOFS root
- Built-in ignore list (system files)
- User-added ignore patterns (from previous `scan --adopt` sessions)

**Interactive adoption prompts:**

| Key | Action |
|-----|--------|
| `y` | Move file into `<root>/home/` and create symlink |
| `n` | Skip (will appear again next scan) |
| `i` | Ignore persistently (won't appear again) |
| `q` | Quit |

**Flags:**

| Flag | Description |
|------|-------------|
| `--adopt` | Enable interactive adoption mode |
