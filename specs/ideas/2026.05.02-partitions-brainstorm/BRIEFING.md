# Briefing: Partitions and Disks

**Core insight:** DOFS should model itself as an actual filesystem — not
metaphorically, but structurally. An installation is a filesystem, disks
are organizational groups within it, partitions are the user's data
buckets, and folders within partitions can carry their own config overrides.
This eliminates hardcoded categories, enables multi-user setups, and makes
the tool's identity ("Developer Optimized File System") literal.

**Date:** 2026-05-02
**Trigger:** Hardcoded categories surfaced during a bug fix. The fixed
nine-category enum cannot accommodate user-specific needs and blocks
multi-user workstation support.

## Pre-requisite reading

- [Constitution](../../.specify/memory/constitution.md) — especially
  Principle I (filesystem is source of truth) and Principle IV (architecture)
- [Foundation spec](../drafts/01-foundation-and-init.md) — current category
  model being replaced
- [Profiles idea](../profiles.md) — related concept that this design
  partially subsumes

## The Problem

DOFS prescribes nine categories baked into the binary. A developer who
needs a partition for VMs, games, or work-specific projects cannot add one
without recompiling. More critically, the flat category model has no answer
for multi-user workstations where two OS users share a DOFS root with
different ownership and access requirements.

The "profiles" idea explored persona-switching but was really about
activation management on top of a fixed structure. The real problem is
the structure itself.

## What We Designed

A three-layer structural hierarchy (Filesystem, Disk, Partition) that
mirrors actual filesystem concepts. Configuration cascades through four
levels (FS, Disk, Partition, Folder) with inheritance, where folders
within partitions can recursively override settings. Settings group into
three concerns (security, backup, linking). The system handles single-user
and multi-user setups with the same code paths and the same directory
structure.

Key properties:
- Always at least one disk; no "single-disk vs. multi-disk" format modes
- All disks are created empty by default; `--guided` flag triggers an
  interactive partition wizard on any disk; `dofs init` implies `--guided`
- Disk binding is ownership-based (Unix permissions, no config needed)
- Symlink target collisions across disks produce errors (conservative;
  proper resolution deferred to real usage)
- Partition definitions are portable (travel with backups); machine-specific
  settings are local
- Folders within partitions can carry `.folder.yaml` overrides (e.g.,
  different link targets for SSH vs. GPG within a vault partition)

## Decisions

| # | Title | Summary |
|---|-------|---------|
| 1 | Three-layer hierarchy | FS, Disk, Partition. Always at least one disk. One structure for all setups. |
| 2 | Cascading configuration | `.dofs.yaml`, `.disk.yaml`, `.partition.yaml`, `.folder.yaml`. Four levels, inherit and override. |
| 3 | Three setting groups | Security, Backup, Linking. Orthogonal concerns at any cascade level. |
| 4 | Portable vs. machine-local split | Config files travel with backups. Machine bindings stay local. |
| 5 | Ownership-based disk binding | Unix permissions determine which disks a user sees. Shared group declared in FS config. |
| 6 | Error on collision | Symlink target conflicts across disks are errors. Defer resolution design. |
| 7 | Guided disk scaffolding | All disks empty by default. `--guided` triggers partition wizard. `dofs init` implies `--guided`. |
| 8 | Partitions replace categories | Hardcoded Go enum removed. Partitions discovered from filesystem, configured via YAML. |
| 9 | Folder-level config overrides | `.folder.yaml` extends the cascade recursively into subdirectories. Not a managed entity, just a config override point. |

## Next Steps

- Write a formal spec for partitions (phase 1: single-disk, configurable
  partitions, no multi-disk yet)
- Design the config file schema (`.dofs.yaml`, `.disk.yaml`,
  `.partition.yaml`, `.folder.yaml`)
- Plan migration path from current hardcoded categories to dynamic
  partitions
- Update the constitution to reference the FS/Disk/Partition model
- Revisit `profiles.md` — disk-level scoping may obsolete parts of it

## Open Questions

- **Config schema divergence:** Will the four config levels eventually
  need different schemas, or can a single schema with level-appropriate
  defaults suffice?
- **Machine-local storage:** Database or local config file for
  machine-specific settings? Left as implementation detail.
- **Conflict resolution patterns:** What real-world scenarios trigger
  symlink collisions? Design proper resolution once usage data exists.
- **User/group provisioning:** Should DOFS manage group creation directly
  (needs sudo) or just declare requirements and instruct the user?
- **Backup storage backends:** Partitions may need different backup
  strategies (tar, encrypted archive, git, S3, cloud storage). Backend
  is a partition-level setting. Needs design for pluggability, credential
  flow, and how it changes the backup setting group. See open thread in
  design decisions.
- **Encryption scope and engine:** Three distinct encryption scenarios
  (at-rest/LUKS, backup encryption, file-level/SOPS) should not be
  conflated. Which does DOFS own vs. delegate? SOPS has a stable Go
  decrypt API but encryption requires unstable internals. Age is a
  simpler alternative. See open thread in design decisions.

## Links

- [Session context](00-session-context.md)
- [Design decisions](01-design-decisions.md)
