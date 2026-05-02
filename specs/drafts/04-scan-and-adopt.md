# Spec Draft: Scan & Adopt

## Summary

Discover dotfiles in `~/` that aren't managed by DOFS and interactively adopt them — moving files into the DOFS tree, creating symlinks, and registering them. Includes persistent ignore patterns so dismissed items don't reappear.

## User Stories

- As a developer, I want to see which dotfiles in my home directory aren't tracked yet so I can decide what to adopt.
- As a developer, I want to interactively move dotfiles into DOFS with a single prompt per item.
- As a developer, I want to permanently ignore system-managed dotfiles so they don't clutter future scans.

## Scope

### `dofs scan` (list mode)
- Read `~/` entries at depth 1
- Filter out:
    - Non-dotfiles (no leading `.`)
    - Items already in `managed_links`
    - Symlinks already pointing into the DOFS root
    - Built-in ignore list (system-managed files like `.Xauthority`, `.bash_history`, etc.)
    - User-added patterns from `scan_ignores` table
- Display remaining candidates

### `dofs scan --adopt` (interactive mode)
- Present each candidate with a prompt: `[y/n/i/q]`
- **y (adopt)**: `os.Rename(~/.<item>, <root>/home/<item>)` → `os.Symlink(...)` → insert `managed_links` + `scan_history`
- **n (skip)**: skip, will appear again next scan
- **i (ignore)**: insert into `scan_ignores`, insert `scan_history` as ignored
- **q (quit)**: stop processing

### Atomic Adoption
- Move file first, then create symlink
- If symlink creation fails, roll back the move
- Insert DB records after filesystem operations succeed

### Scan History
- Record each adoption/ignore in `scan_history` with disposition and timestamp
- Provides audit trail of what was adopted and when

## Dependencies

- Spec 1 (Foundation & Init) — config, DB, model
- Spec 2 (Link Management) — `managed_links` data for filtering, link creation logic

## Key Design Decisions

- Depth-1 only: scanning recursively into `~/.config/` etc. is out of scope for now
- Default category for adoption is `home` (secrets require manual `dofs link --add`)
- Ignore patterns are exact name matches (not globs) stored in the DB
- Rollback on failure: if symlink creation fails after the move, the file is moved back

## Acceptance Criteria

- `dofs scan` lists only genuine candidates (no already-tracked, no system files)
- `dofs scan --adopt` with `y` moves the file and creates a working symlink
- Adoption is atomic: failure mid-way rolls back the move
- `i` permanently hides the item from future scans
- `q` stops processing remaining candidates
- Already-ignored items don't appear in scan output
- `--dry-run` shows what would be adopted without moving anything
