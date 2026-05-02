# Session Context

**Date:** 2026-05-02
**Trigger:** While debugging a SQLite "unable to open database" error, the
conversation surfaced that DOFS categories are hardcoded as a Go enum in
`internal/model/model.go`. The user wants to evolve these into user-manageable
"partitions" — a better fit for a tool whose identity is "Developer Optimized
File System."

**Starting state:**

- Nine categories are hardcoded as `const` values: home, secrets, code, data,
  desktop, documents, downloads, music, pictures.
- Two hardcoded slices gate behavior: `LinkableCategories` (home, secrets) and
  `AllCategories` (all nine).
- Categories are used across every layer: model, linker, scanner, backup,
  restore, status, and CLI commands.
- No config file exists. All configuration is flags + env + defaults.
- The spec (`01-foundation-and-init.md`) lists the categories as a fixed set.
- The constitution doesn't address category mutability.
- The `profiles.md` idea explores persona-based scoping but doesn't touch
  category/partition management.

**Related prior decisions:** None (first brainstorm session).
