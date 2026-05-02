# Idea: Profiles

## Concept

Treat the machine as a shared tool used by the developer acting on behalf of different personas (e.g. personal vs work). Activating a profile would scope persona-specific configuration — switching which Git identity, SSH keys, GPG keys, and Kubernetes contexts are active.

## Example Use Cases

- Switching between personal and employer Git/SSH/GPG identities
- Activating a different `~/.kube/config` per client or environment
- Keeping work and personal configs cleanly separated on a single machine

## Persona-Specific Items

- Git config (`~/.gitconfig`)
- SSH keys (`~/.ssh`)
- GPG keys (`~/.gnupg`)
- Kubernetes config (`~/.kube`)

## Open Questions

- **Activation model**: Does `dofs profile <name>` swap symlinks globally? Or are links tagged with profiles and `dofs link --create` only creates the active set?
- **Shared links**: Some links (like `~/.zshrc`) would be common across all profiles. How to distinguish shared vs profile-specific links? Tagging? Separate categories?
- **Storage layout**: Profile-specific items stored under `<root>/profiles/<name>/`? Or stay in `home/`/`secrets/` with metadata tagging?
- **Conflict handling**: What happens to the currently active profile's symlinks when switching? Remove and recreate, or error if conflicts exist?
- **Scope**: Just symlinks, or could profiles also scope things like shell env vars, git includes, etc.?
