# Machine Setup

This guide covers the physical setup of a DOFS-ready machine: partitioning, encryption, permissions, and auto-mounting. These steps are done once per machine and are independent of the `dofs` CLI tool itself.

The examples use Fedora Workstation on a 1 TB NVMe drive, but the concepts apply to any hardware and distro.

## Partition Layout

Instead of letting the installer use the entire disk, create a custom layout that separates your OS from your data:

| Partition | Name | Size | Type | Encrypted | Mount Point |
|-----------|------|------|------|-----------|-------------|
| 1 | `efi` | 512 MB | EFI System Partition | No | `/boot/efi` |
| 2 | `boot-fedora` | 1 GB | ext4 | No | `/boot` |
| 3 | `root-fedora` | 250 GB | ext4 | LUKS2 | `/` |
| 4 | `dofs` | 250 GB | ext4 | LUKS2 | `/mnt/dofs` |
| — | *(free space)* | ~499 GB | — | — | — |

**Why this layout?**

- The **EFI** partition is shared across all OSes — each adds its own bootloader entry.
- Each OS gets its own **boot** and **root** partitions. Name them descriptively (e.g., `boot-fedora`, `root-ubuntu`) so you can tell them apart later.
- The **DOFS** partition is independent and never wiped by OS installs.
- **Free space** is left unallocated for future OS installs or additional data. You can always create new partitions later.

> [!TIP]
> Use the same LUKS passphrase for the root and DOFS partitions. Most distros will unlock all LUKS partitions with a single passphrase prompt at boot if they share the same password.

> [!IMPORTANT]
> If your company requires full-disk encryption, encrypt both root and DOFS. The EFI and boot partitions cannot be encrypted (the bootloader needs to read them).

## Shared Group for Cross-Distro Access

Different distros may assign different UIDs/GIDs to your user. To ensure consistent file access across OSes, create a dedicated group with a fixed GID:

```bash
# Pick a memorable GID (something high to avoid collisions with system groups)
sudo groupadd -g <YOUR_GID> dofs

# Add your user to the group
sudo usermod -aG dofs $(whoami)
```

Run these same two commands on every distro you install. As long as the GID matches, file permissions on the DOFS partition work seamlessly across all of them.

## Mounting the DOFS Partition

```bash
# Create the mount point
sudo mkdir -p /mnt/dofs

# Unlock the LUKS partition (if encrypted)
sudo cryptsetup luksOpen /dev/nvme0n1p4 dofs-crypt

# Mount it
sudo mount /dev/mapper/dofs-crypt /mnt/dofs

# Set ownership and permissions
sudo chown -R $(whoami):dofs /mnt/dofs
sudo chmod -R 2775 /mnt/dofs
```

The `2775` permission sets the **setgid bit**, meaning any new files or directories created inside `/mnt/dofs` automatically inherit the `dofs` group. This is key for cross-distro access.

## Auto-Mount at Boot

To avoid manually mounting on every boot, add entries to `crypttab` and `fstab`.

Get the UUID of your DOFS partition:

```bash
sudo blkid /dev/nvme0n1p4
# Note the UUID value (not PARTUUID)
```

Add the LUKS entry:

```bash
echo 'dofs-crypt UUID=<YOUR_UUID> none luks' | sudo tee -a /etc/crypttab
```

Add the mount entry:

```bash
echo '/dev/mapper/dofs-crypt /mnt/dofs ext4 defaults 0 2' | sudo tee -a /etc/fstab
```

Reboot and verify it mounts automatically. If your root and DOFS partitions share the same LUKS passphrase, you should only be prompted once.

## Adding a Second Distro

This is where DOFS really shines. Say you want to add Ubuntu alongside Fedora:

1. **Create partitions** from your free space:
    - `boot-ubuntu` — 1 GB, ext4
    - `root-ubuntu` — 250 GB, ext4, LUKS2 (same passphrase)

2. **Install Ubuntu** onto `root-ubuntu`. Point `/boot` at `boot-ubuntu`. The installer will detect the existing EFI partition and share it.

3. **After first boot into Ubuntu**, set up DOFS access:

    ```bash
    # Create the same group with the same GID
    sudo groupadd -g <YOUR_GID> dofs
    sudo usermod -aG dofs $(whoami)

    # Mount the DOFS partition
    sudo mkdir -p /mnt/dofs
    sudo cryptsetup luksOpen /dev/nvme0n1p4 dofs-crypt
    sudo mount /dev/mapper/dofs-crypt /mnt/dofs

    # Set up auto-mount (same crypttab/fstab entries as above)

    # Bootstrap DOFS — auto-imports existing symlinks
    dofs init
    dofs link --create
    ```

4. **Done.** Your SSH keys, Git config, Kube config, and everything else is already there. Same files, same permissions, different OS.

## Ventoy: Multi-Boot USB Companion

When setting up new machines or distros, [Ventoy](https://ventoy.net) is invaluable. Install it once on a USB drive or external SSD, then just drop ISO files onto it. Each ISO becomes bootable — no reflashing needed. Keep your Fedora, Ubuntu, and recovery ISOs all on one drive alongside regular files.

> [!IMPORTANT]
> **Secure Boot:** On first boot with Ventoy, you may need to enroll its MOK (Machine Owner Key). Look for `ENROLL_THIS_KEY_IN_MOKMANAGER.cer` in the Ventoy menu and enroll it. This is a one-time step per machine.

## What to Put in DOFS

- SSH keys and config (`~/.ssh`)
- GPG keys (`~/.gnupg`)
- Git configuration (`~/.gitconfig`, `~/.gitignore_global`)
- Kubernetes config (`~/.kube`)
- Shell configuration (`~/.zshrc`, `~/.bashrc`)
- Editor configs (`~/.config/nvim`, `~/.vscode/`)
- AWS / cloud CLI credentials (`~/.aws`)
- Docker config (`~/.docker/config.json` — but not Docker images/volumes)
- Personal scripts and tools

## What NOT to Put in DOFS

- Application caches (`~/.cache`) — OS/version specific
- Runtime state (`~/.local/share` in general) — often not portable
- Entire `~/.config` directory — some configs are distro-specific; symlink individual app configs instead
- Docker images and volumes — tied to the OS install in `/var/lib/docker`
- Package manager data — distro-specific

## Security Considerations

- The DOFS partition should be LUKS2-encrypted at rest.
- SSH keys and GPG keys need proper permissions (`chmod 600` for private keys, `chmod 700` for `.ssh` and `.gnupg` directories).
- If you version-control your DOFS root, exclude `secrets/` in your `.gitignore`.
- The shared group approach means any user in the `dofs` group on any OS can access the files — only add trusted users.

## FAQ

**Q: Can I resize the DOFS partition later?**
A: Yes, but with LUKS2 it requires booting from a live USB and resizing both the LUKS container and the filesystem. It's easier to leave unallocated space during initial setup.

**Q: What filesystem should I use?**
A: **ext4** if you're staying Linux-only. It's the most reliable and widely supported. If you need Windows compatibility, things get more complicated since Windows can't natively read ext4 or LUKS.

**Q: Will my UID mismatch cause issues?**
A: Not if you use the shared group approach. The `dofs` group with a fixed GID and the setgid bit means all files are group-accessible regardless of which user created them.

**Q: Can I use this with NixOS / immutable distros?**
A: Yes. The DOFS partition is just a mounted filesystem. Immutable distros can still mount external partitions and create symlinks in your home directory.

**Q: What about Homebrew / Linuxbrew?**
A: Homebrew installs to `/home/linuxbrew/.linuxbrew` by default and works identically across distros. It's a great way to manage dev tools without relying on distro-specific package managers. Homebrew itself lives on the root partition, but your Homebrew config/taps can be tracked in DOFS.
