# Getting Started

## Installation

```bash
go install github.com/tjespers/dofs@latest
```

Or build from source:

```bash
git clone https://github.com/tjespers/dofs.git
cd dofs
go build -o dofs .
```

## Prerequisites

Your DOFS root directory should already exist with the expected category structure. If you're setting up a machine from scratch, see [Machine Setup](machine-setup.md) first.

The default DOFS root is `~/dofs/data/`. You can override it with `--root`, the `DOFS_ROOT` environment variable, or let it default.

## Bootstrap: `dofs init`

Initialize DOFS on a machine where you already have symlinks pointing into your DOFS root:

```bash
dofs init
```

This will:

1. Validate the root directory exists
2. Check for expected category subdirectories (`home/`, `secrets/`, `code/`, etc.)
3. Create the SQLite database at `<root>/.dofs.db`
4. Scan `~/` for existing symlinks that point into the DOFS root and auto-import them

On a typical setup with ~30 existing symlinks, `dofs init` will find and register all of them automatically — no manual re-registration needed.

## Check Health: `dofs status`

See the current state of everything:

```bash
dofs status
```

This shows a summary of link health (ok, broken, missing, conflicts), untracked items in linkable categories, and backup ages.

## Verify Links: `dofs link`

List all tracked symlinks and their current status:

```bash
dofs link
```

Fix problems:

```bash
# Create any missing symlinks
dofs link --create

# Repair broken symlinks
dofs link --repair
```

## Adopt New Dotfiles: `dofs scan`

Find dotfiles in `~/` that aren't tracked yet:

```bash
# List candidates
dofs scan

# Interactive adoption — move each into DOFS and create symlinks
dofs scan --adopt
```

For each candidate, you can adopt it, ignore it (persistently), or skip it.

## Backup and Restore

### Backup

Create per-category tar.gz archives:

```bash
# Backup everything
dofs backup -o /path/to/export/

# Backup a specific category
dofs backup home -o /path/to/export/

# Preview what would be archived
dofs backup code --dry-run
```

### Restore on a New Machine

```bash
# Extract archives, init DB, register items, create symlinks
dofs restore --link /path/to/dofs-backup-*.tar.gz
```

The `restore` command handles the full lifecycle: extract archives into the DOFS root, initialize the database, register items from linkable categories, and optionally create all symlinks.

## Typical Migration Workflow

```bash
# On the old machine
dofs backup -o /media/usb/dofs-export/

# On the new machine (after partition/mount setup)
dofs restore --link /media/usb/dofs-export/dofs-backup-*.tar.gz
```

## Global Flags

| Flag | Env Var | Default | Description |
|------|---------|---------|-------------|
| `--root` | `DOFS_ROOT` | `~/dofs/data` | Path to the DOFS root directory |
| `--dry-run` | — | `false` | Show what would happen without making changes |
| `--verbose` | — | `false` | Enable detailed output |
