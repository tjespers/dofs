# DOFS — Developer Optimized File System

DOFS is a Go CLI tool that manages a developer's workstation files through a structured directory layout. It replaces the manual process of organizing dotfiles, configs, secrets, and code repos into a canonical tree and symlinking them to their expected OS locations.

## The Problem

Developers accumulate a sprawl of dotfiles, SSH keys, Kubernetes configs, GPG keys, editor settings, and project files scattered across their home directory. Reinstalling your OS, switching distros, or dual-booting means hours of reconfiguring everything. Over time, new dotfiles appear and never get organized, symlinks break, and things get forgotten.

## The Idea

**One directory to rule them all.** All your userland files live in a single directory tree organized by purpose. Symlinks point from where applications expect files (like `~/.ssh`, `~/.kube`, `~/.gitconfig`) back to their real location inside the DOFS root. The `dofs` CLI tracks these symlinks, detects drift, and automates the entire lifecycle — from initial setup to backup and restore on a new machine.

## Core Concepts

- **DOFS root**: A single directory (default `~/dofs/data/`) containing all managed files organized by category.
- **Categories**: Top-level folders in the root — `home`, `secrets`, `code`, `data`, `desktop`, `documents`, `downloads`, `music`, `pictures`.
- **Linkable categories**: `home` and `secrets` — items in these get symlinked to `~/`.
- **Managed links**: Database-tracked symlinks between the DOFS tree and the OS filesystem.
- **The DB is a cache, the filesystem is truth**: Link status is always re-verified against the actual filesystem.

## Why This Works

- **One backup target.** Back up the DOFS root and you've got everything.
- **Distro-independent.** Your data lives on its own partition, untouched by OS installs.
- **Dual/triple boot ready.** Install Fedora, Ubuntu, Arch — each mounts the same DOFS partition, runs `dofs init && dofs link --create`, done.
- **Version controllable.** Track your configs in Git (minus the secrets) for history and sharing across machines.
- **Drift-proof.** The CLI detects new dotfiles, broken symlinks, and untracked items automatically.

## Tech Stack

- **Go** — single portable binary, no runtime dependencies
- **Cobra** — CLI framework with subcommands, help generation, and shell completion
- **SQLite** (via `modernc.org/sqlite`) — CGo-free driver for cross-compilation without a C toolchain
