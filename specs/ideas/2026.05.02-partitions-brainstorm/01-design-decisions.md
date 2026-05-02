# Design Decisions

## Decision 1: Three-layer hierarchy

**Context:** DOFS currently has a flat list of nine hardcoded categories.
The user needs configurable organization that supports single-user and
multi-user setups without different code paths or migration tooling.

**Decision:** DOFS adopts a three-layer hierarchy: Filesystem (installation)
to Disk to Partition. Every installation has at least one disk. Partitions
live inside disks. This structure is always the same regardless of complexity;
a solo developer has one disk with a few partitions, a multi-user workstation
has several disks with different ownership.

**Rationale:** A single consistent structure eliminates format modes,
conversion commands, and conditional code paths. The "always at least one
disk" rule means adding a second disk is just `dofs disk create`, not a
format migration. The cost (one extra directory level) is minimal compared
to the complexity of supporting two layouts.

**Trade-off acknowledged:** Solo developers get a directory layer (e.g.,
`dofs/main/home/` instead of `dofs/home/`) that adds no value for their
use case. Accepted because structural consistency across all setups outweighs
the minor path depth increase.


## Decision 2: Cascading configuration with four levels

**Context:** Partitions need per-instance settings (encryption, backup rules,
linking behavior), but requiring full configuration on every partition is
tedious and error-prone. Most partitions within a disk share the same
ownership and similar backup rules. Additionally, folders within a partition
may need different settings (e.g., different link targets for SSH keys vs.
GPG keys within a vault partition).

**Decision:** Configuration cascades through four levels via config files:
`.dofs.yaml` (filesystem-level defaults), `.disk.yaml` (disk-level overrides),
`.partition.yaml` (partition-level overrides), `.folder.yaml` (folder-level
overrides, recursive). Each level inherits from its parent and can override
any setting. Missing config files at any level simply mean "inherit everything
from above."

**Rationale:** This matches how developers already think about configuration
(global git config, per-repo config, per-worktree config). It minimizes
repetition while keeping overrides explicit and local to where they apply.
Extending the cascade into folders avoids splitting conceptually related data
into separate partitions solely because of differing link targets or
permissions.

**Rejected:** Flat per-partition config with no inheritance. Would require
copying identical settings across every partition in a disk. Violates DRY
and makes bulk changes painful.


## Decision 3: Three setting groups

**Context:** Partitions have various properties that need configuration.
These properties serve different purposes and have different portability
characteristics.

**Decision:** Settings are organized into three groups:

- **Security:** encrypted (boolean), encryption-key (e.g., age public key),
  default file mode, owner uid/gid.
- **Backup:** exclusion patterns, schedule, target storage location,
  retention policy.
- **Linking:** enabled (boolean), target path (default: `~/`).

Each group can be set at any level of the cascade (FS, disk, partition).

**Rationale:** Grouping by concern makes config files readable and makes it
clear which settings interact. Security settings gate access and encryption.
Backup settings control archival. Linking settings control symlink behavior.
These concerns are orthogonal and don't cross-contaminate.


## Decision 4: Portable vs. machine-local configuration split

**Context:** Some settings are inherent to the data (this partition is
encrypted, these files should be excluded from backup) and must travel with
backups for restore on a new machine. Other settings are specific to the
machine (backup target path, schedule timing, resolved uid/gid mappings).

**Decision:** The config files (`.dofs.yaml`, `.disk.yaml`,
`.partition.yaml`) are portable; they live inside the DOFS root and are
included in backups. Machine-local settings (backup target paths, schedule
timing, and other installation-specific bindings) live separately, either
in the database or a machine-local config outside the DOFS root.

**Rationale:** Portable config ensures that restoring a backup on a new
machine reconstructs the full partition definitions, security settings, and
linking rules. Machine-local settings are excluded because they would be
wrong on a different machine (different mount points, different schedules,
different uids).

**Trade-off acknowledged:** The split means two places to look for
configuration. Accepted because conflating portable and machine-local
settings would cause restore to apply stale machine bindings from the
source machine.


## Decision 5: Ownership-based disk binding

**Context:** In a multi-user setup, DOFS needs to know which disks to
process for the current user. Two disks owned by different users share
a DOFS root, but each user should only see and manage their own disks
(plus any shared ones).

**Decision:** Disk binding is ownership-based. DOFS processes disks that
the current user has filesystem-level read access to. No explicit binding
configuration is needed; standard Unix permissions (user ownership and
group membership) determine visibility.

**Rationale:** This leverages the OS permission model rather than
reinventing it. Tar archives preserve uid/gid, so ownership information
travels with backups. The filesystem-level group (e.g., `dofs` group with
a declared gid) enables shared disks that multiple users can access.

**Note:** The FS-level config (`.dofs.yaml`) must declare the shared group
name and gid so that `dofs init` on a new machine can instruct the user to
create the group before proceeding. Tar preserves the gid on files, but
the group definition itself needs to be explicit.


## Decision 6: Conflict resolution — error on collision

**Context:** When two disks contain partitions that would link the same
file to the same target (e.g., `work/home/.gitconfig` and
`shared/home/.gitconfig` both targeting `~/.gitconfig`), DOFS needs a
resolution strategy.

**Decision:** DOFS errors on symlink target collision. If two partitions
across different disks would create symlinks to the same target path,
`dofs link` refuses and reports the conflict with both sources.

**Rationale:** Silently picking a winner (via priority ordering or
last-write-wins) would hide a real configuration problem. The user should
know about and resolve conflicts explicitly. More sophisticated resolution
(disk priority ordering, per-link overrides) can be designed once real
usage patterns reveal what users actually need.

**Trade-off acknowledged:** This is deliberately conservative. Users who
hit this will need to manually restructure (move the file to one disk,
remove from the other). Accepted because premature conflict resolution
design would likely be wrong without real-world usage data.


## Decision 7: First disk is scaffolded, subsequent disks are empty

**Context:** A new DOFS installation needs some structure to be useful, but
forcing a specific partition layout contradicts the goal of user-manageable
partitions.

**Decision:** All disks are created empty by default, whether it's the
first disk (`dofs init`) or a subsequent one (`dofs disk create`). A
`--guided` flag triggers an interactive wizard that proposes recommended
partitions (home, secrets, code, etc.) for the user to select, deselect,
or extend with custom names. This flag works the same on any disk,
first or Nth. When `dofs init` creates the first disk, it implies
`--guided` automatically since new users benefit from the interactive
partition proposal.

**Rationale:** Treating all disks identically eliminates special-casing
between `dofs init` and `dofs disk create`. The guided wizard is opt-in
because users who know what they want shouldn't be slowed down, while
users who want suggestions can request them on any disk. An empty disk is
a valid disk; DOFS commands handle zero partitions gracefully (empty
status, nothing to backup, nothing to link).

**Rejected:** Scaffolding the first disk differently from subsequent disks.
Would require special-case logic for the first disk and contradicts the
principle that all disks are structurally identical.


## Decision 8: Partitions replace hardcoded categories

**Context:** The current codebase defines categories as a Go enum with nine
constants and two hardcoded slices (`AllCategories`, `LinkableCategories`).
This is used across every layer: model, linker, scanner, backup, restore,
status, and CLI commands.

**Decision:** The hardcoded `Category` type and its constants are removed.
Partitions are discovered from the filesystem (directories within a disk)
and configured via `.partition.yaml`. The `LinkableCategories` concept is
replaced by the `linking.enabled` setting per partition. The `AllCategories`
concept is replaced by directory listing of the disk.

**Rationale:** Hardcoded categories cannot accommodate user-specific needs
(custom partitions for VMs, games, work projects). Dynamic partitions let
the user define their own organizational structure. The filesystem-as-truth
principle (Constitution, Principle I) already supports this; partitions
exist because their directories exist.

**Trade-off acknowledged:** Losing compile-time category validation means
typos in partition names won't be caught by the compiler. Accepted because
the trade-off is inherent to any user-configurable system, and validation
moves to runtime (does this directory exist?).


## Decision 9: Folder-level configuration overrides

**Context:** A partition groups data by what it IS (e.g., a `vault`
partition for all sensitive credentials). But different folders within a
partition may need different deployment targets or permissions. SSH keys
link to `~/.ssh`, GPG keys to `~/.gnupg`, sops config to `/etc/sops`.
Without per-folder overrides, these would need to be separate partitions
solely because they have different link targets, breaking the conceptual
grouping.

**Decision:** A `.folder.yaml` file can be placed in any directory within
a partition to override inherited settings. Nested folders inherit from
their parent folder's resolved config (partition config merged with any
ancestor folder configs). Most folders won't have a `.folder.yaml`; the
override only exists where needed. Folders are not a managed entity like
disks or partitions (there is no `dofs folder create`); they are simply
directories that DOFS recognizes as potential config override points.

**Rationale:** This keeps data organized by concept (all credentials in
one partition) rather than by deployment target (separate partitions for
each link destination). It reuses the same cascade mechanism from Decision
2 without introducing a new structural entity.

**Trade-off acknowledged:** Recursive config merging adds implementation
complexity to the directory walker. Every directory potentially has a
config file to check and merge. Accepted because the merge is a simple
"read if exists, overlay on parent" operation, and the flexibility of
per-folder link targets is essential for real-world use cases like the
vault scenario.


## Open: Backup storage backends

**Status:** Needs further exploration. Captured to preserve the train of
thought.

**Observation:** Backup is not just "make a tar.gz." Different partitions
may need fundamentally different storage backends: plain tar archive,
encrypted archive, VCS (git) for config history, rsync/S3 for off-site
incremental, cloud storage (Google Drive). The backend is a partition-level
setting since folder-level would be chaotic and disk-level is too coarse
when partitions within a disk need different backends. Each backend may
need its own config (S3 bucket, git remote URL, encryption key).

**Observation:** This expands the backup setting group from Decision 3
significantly. It's not just "exclusions and schedule" but "which backend,
with what credentials, to what destination."

**To resolve:** Is the backend a simple enum in partition config, or a
pluggable interface? How do credentials flow (portable vs. machine-local)?
Does this change how `dofs backup` works fundamentally, or is the current
tar.gz just the first backend implementation?


## Open: Encryption — scope and engine

**Status:** Needs further exploration. Captured to preserve the train of
thought.

**Observation:** Three distinct encryption scenarios exist and should not
be conflated:

1. **At rest (LUKS-style):** OS-managed disk encryption. DOFS doesn't
   control this but may need to know about it as metadata.
2. **Backup encryption:** Archives are encrypted before storage. This is
   a property of the backup backend. Age (pure Go, simple) is a candidate
   engine.
3. **File-level (SOPS-style):** Individual files encrypted at rest,
   decrypted on read. SOPS has a stable Go decrypt API
   (`github.com/getsops/sops/v3/decrypt`) but encryption programmatically
   requires unstable internal packages.

**Key question:** Should DOFS manage encryption directly, or declare
encryption requirements and delegate to external tools? The constitution's
safety-first principle (I) and single-binary constraint (II) pull in
opposite directions here — SOPS integration adds dependencies, but
shelling out to external tools breaks the single-binary promise.

**To resolve:** Which of the three scenarios does DOFS actually need to
own? Can at-rest encryption be fully delegated to the OS? Is backup
encryption just a backend concern? Does file-level encryption belong in
DOFS at all, or is it the user's responsibility to use SOPS independently?
