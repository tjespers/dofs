# Spec Draft: Status

## Summary

A single health report command that aggregates link states, identifies untracked items, and shows backup freshness — giving the developer a quick overview of their DOFS health.

## User Story

As a developer, I want to run one command and immediately see if anything needs attention: broken links, untracked dotfiles, or stale backups.

## Scope

### `dofs status`
- **Link summary**: counts grouped by status (ok, broken, missing, conflict, wrong_target)
- **Untracked items**: items found in linkable categories (`home/`, `secrets/`) that are not in `managed_links`
- **Backup ages**: last backup timestamp per category, highlighting stale or never-backed-up categories

### Output Format
- Concise summary suitable for terminal display
- Clear indication of items needing attention vs everything-ok state

## Dependencies

- Spec 1 (Foundation & Init) — config, DB, model
- Spec 2 (Link Management) — link verification logic (reused to get current statuses)

## Key Design Decisions

- Status re-verifies links against the filesystem (doesn't trust cached `last_status`)
- Untracked detection walks linkable category directories and compares against `managed_links`
- Backup age is read from `backup_history` table

## Acceptance Criteria

- Shows correct counts for each link status
- Lists untracked items in linkable categories
- Shows last backup time per category (or "never" if no backup exists)
- Works correctly with zero links, zero backups (fresh state)
- Exits cleanly even if some categories don't exist on disk
