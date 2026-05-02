# Spec Draft: Restore

## Summary

Extract DOFS backup archives onto a new machine, bootstrap the database, register items from linkable categories, and optionally create all symlinks — the complete migration landing pad.

## User Story

As a developer setting up a new machine, I want to point DOFS at my backup archives and have it recreate my entire setup: extract files, initialize the database, and create all symlinks.

## Scope

### `dofs restore [--link] <archives...>`
- Accepts one or more archive file paths (glob-friendly)
- For each archive:
    1. Parse category name from filename (`dofs-backup-<category>-YYYYMMDD-HHMMSS.tar.gz`)
    2. Create `<root>/<category>/` directory
    3. Extract archive contents into it
- After all archives extracted:
    4. Create/open SQLite database
    5. Walk linkable categories (`home/`, `secrets/`)
    6. For each item found: derive target path (`~/.<name>`), insert into `managed_links`
    7. If `--link`: create all missing symlinks

### Archive Extraction
- Regular files: create with original permissions
- Symlinks: recreate as symlinks
- Directories: create with original permissions
- **Zip-slip protection**: validate all paths stay within the target directory

### Safety
- Does not overwrite existing files by default
- Category parsed from filename — malformed filenames produce clear errors
- DB registration uses the same logic as `dofs init` for deriving target paths

## Dependencies

- Spec 1 (Foundation & Init) — config, DB, model, migration
- Spec 2 (Link Management) — symlink creation logic
- Spec 5 (Backup) — produces the archives this command consumes

## Key Design Decisions

- Restore is a first-class command, not just "extract and run init"
- Category parsed from filename avoids needing metadata inside the archive
- Linkable category registration happens automatically — no manual `--add` for each item
- `--link` is opt-in: you might want to extract and inspect before creating symlinks

## Acceptance Criteria

- `dofs restore /tmp/dofs-backup-home-*.tar.gz` extracts into `<root>/home/`
- Multiple archives in one invocation all get extracted correctly
- Items in `home/` and `secrets/` are registered in `managed_links`
- `--link` creates symlinks for all registered items
- Zip-slip attempts are rejected with an error
- Malformed archive filenames produce a clear error
- Running restore twice doesn't fail on existing DB records
- `--dry-run` shows what would be extracted and linked without doing it
