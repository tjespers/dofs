# Constitution Input Draft

The following is the natural language description to feed into `/speckit.constitution`:

---

Build a CLI tool called DOFS (Developer Optimized File System) that manages a developer's workstation files by separating them from the operating system. All user files — dotfiles, configs, secrets, code repos, documents — live in a single canonical directory tree organized by category. Items in linkable categories are symlinked back to where applications expect them in the home directory. The tool tracks these symlinks in a local database, detects drift, and automates the full lifecycle from initial setup through backup and restore on a new machine.

The core problem: developers accumulate files scattered across their home directory that are lost or require manual reconfiguration when switching distros, dual-booting, or reinstalling. DOFS solves this by making the user's data OS-independent — it lives on its own partition, survives OS installs, and can be bootstrapped on a new distro with two commands.

Key principles:
- The filesystem is always the source of truth, not the database. The database is a cache for tracking state.
- Safety first: never delete user files. Conflicts require explicit force flags and always create backups before overwriting.
- All mutating operations must respect dry-run mode.
- Operations should be idempotent — running the same command twice is always safe.
- Single portable binary with no runtime dependencies.

The tool is Linux-first (also targeting macOS) and built for developers who distro-hop, dual-boot, or maintain multiple machines. It is not just a dotfile manager — it's a complete userland file organization system with backup/restore for machine migration.
