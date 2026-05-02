# Constitution Input Draft

The following is the natural language input to feed into `/speckit.constitution`:

---

Create principles for DOFS (Developer Optimized File System), a Go CLI tool that separates developer files from the OS through a canonical directory tree with automated symlinking, backup, and restore.

**Code quality**: The codebase follows strict package dependency rules — `model/` is always a leaf package with zero internal imports, all cross-package communication uses model types. Prefer explicit error handling over silent swallowing. Use typed errors for programmatic handling rather than string-wrapped fmt.Errorf. Keep the CGo-free constraint (modernc.org/sqlite) to ensure single-binary cross-compilation.

**Testing standards**: Every spec must include acceptance criteria that map to testable scenarios. Priority test areas: symlink resolution (absolute, relative, double-hop chains), filesystem state detection (5 link statuses), archive round-trips (backup → restore → verify), and zip-slip protection. Integration tests should use temporary directories with real filesystem operations — no mocking symlinks. Test on both Linux and macOS in CI.

**Safety and data integrity**: The filesystem is always the source of truth, the database is a cache. Never delete user files without explicit --force, and always create a backup before overwriting. All mutating operations must respect --dry-run. Operations must be idempotent — running the same command twice is always safe. Scan adoption must be atomic: if symlink creation fails after a move, roll back the move.

**User experience**: CLI output should be concise and scannable — table output for status, clear counts for operations, actionable error messages. Interactive prompts (scan --adopt) must have a quit option and never leave the filesystem in a partial state. The tool should feel fast for the common case (~30 managed symlinks) and provide progress feedback for slow operations (large backups).

**Performance**: Init and link verification should complete in under 1 second for typical setups. Backup exclusions (node_modules, vendor, .venv, etc.) must use fs.SkipDir to avoid walking large dependency trees. SQLite operations use WAL mode for safe concurrent reads.

**Portability**: Linux-first, macOS-compatible. No Windows target. The tool must work across distros with different package managers, init systems, and filesystem layouts. The DOFS root can be on any mounted filesystem — no assumptions about mount points beyond the configured root path.

**Architecture**: Cobra for CLI with clean separation between command definitions (cmd/) and business logic (internal/). Config resolution follows flags > environment variables > defaults. Each command either reads state or mutates it — never both implicitly. New features should map cleanly to a single subcommand.
