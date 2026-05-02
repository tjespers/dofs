# DOFS — Developer Optimized File System

A Go CLI tool that manages your developer workstation files through a structured directory layout. Track dotfiles, configs, secrets, and code repos in a canonical tree with automated symlinking — surviving distro hops, dual boots, and fresh installs.

## Quick Start

```bash
# Install
go install github.com/tjespers/dofs@latest

# Bootstrap on an existing setup (auto-imports ~/ symlinks)
dofs init

# Check health
dofs status

# Find and adopt untracked dotfiles
dofs scan --adopt

# Backup for migration
dofs backup -o /path/to/export/

# Restore on a new machine
dofs restore --link /path/to/dofs-backup-*.tar.gz
```

## How It Works

All your userland files live in a single DOFS root (default `~/dofs/data/`) organized by category: `home/`, `secrets/`, `code/`, `documents/`, etc. Items in `home/` and `secrets/` are symlinked back to `~/`. The `dofs` CLI tracks these symlinks in a local SQLite database, detects drift, and automates backup/restore across machines.

## Documentation

- [Overview & Core Concepts](docs/index.md)
- [Machine Setup](docs/machine-setup.md) — partitioning, encryption, dual-boot
- [Getting Started](docs/getting-started.md) — installation and basic workflow
- [CLI Reference](docs/cli-reference.md) — all commands and flags
- [Architecture](docs/architecture.md) — package structure and data flows
- [Known Issues](docs/known-issues.md) — bugs, missing features, future work

## License

Apache-2.0
