---
name: use-jj-workspaces
description: Create a fresh Jujutsu workspace for isolated or parallel work, or reuse one only on explicit request. Use for workspace isolation or the jj equivalent of a Git worktree.
compatibility: Requires jj and filesystem inspection tools. Check installed help for version-dependent options.
---

# Use jj Workspaces
Workspaces isolate files, not history: commits, bookmarks, and operations are
shared. Read applicable repository instructions and jj-help guidance first.
Built-in `jj w` is shorthand for `jj workspace`.

## Safety
- Obtain express approval before every jj mutation, including snapshots,
  creation, and update-stale. Explain the base, destination, and exact commands.
  Proposal-only requests permit no writes. Never auto-remove or forget workspaces.
- Protect source/main files and position: no jj new, edit, reset, restore, fetch,
  rebase, abandon, or bookmark movement as setup or cleanup.

## Workflow
1. Inspect with `jj --at-op=@ --ignore-working-copy workspace root` and
   `jj --at-op=@ --ignore-working-copy workspace list`; use filesystem inspection
   for sibling paths. These flags avoid snapshots and operation reconciliation.
   Stop for ambiguous operation heads, Git-only worktrees, or conflicting policy.
2. Use the requested base; if unspecified, ask whether to include current work
   or use a named clean revision. Pin exactly one full commit ID as BASE with
   `jj --at-op=@ --ignore-working-copy log --no-graph -r '<revision>' -T 'commit_id ++ "\n"'`.
   Recorded @ excludes unsnapshotted edits. Including them requires a separately
   approved source snapshot, confirmed inactive source users, and re-pinning BASE;
   disable automatic stale recovery and stop if source recovery is needed.
3. Default to fresh, even if an existing workspace looks reusable. Choose the
   lowest positive N with both name `copy-N` unused and sibling destination
   `<source-name>-copy-N` absent. Derive a quoted absolute path from the source
   root, not the shell directory. If workspaces are very outdated or more than a
   handful, recommend trimming; use judgement or ask if the count is acceptable.
   No hard count/age limit, automatic cleanup, or age-based permission to reuse.
4. Ask approval for the pinned base, name, path, and both commands below. Recheck
   name/path availability before creation; if occupied, reselect and confirm the
   changed proposal. Do not overwrite paths.
   ```sh
   jj -R "<source-root>" --ignore-working-copy workspace add --name "copy-N" -r "commit_id(\"<BASE>\")" "<absolute-sibling-destination>"
   jj -R "<absolute-sibling-destination>" workspace update-stale
   ```

   Explicit -r avoids the default of sharing source @'s parents. The first command
   protects the source but also suppresses destination checkout; the second
   materializes only this new workspace. Respect colocation and sparse defaults.
5. Verify destination root and @'s parent equals BASE before editing; run tools
   there. Report name/path and BASE. Leave files and registration intact at task
   end or on partial failure; report recovery needs instead of repairing them.

## Reuse Only By Request
Ask whether anyone or any background job is using the candidate; never reuse
source/main. Verify its root and shared repository, filesystem work (including
untracked/ignored files), sparse patterns, and colocation. Read-only recorded @
alone cannot prove cleanliness; any snapshot requires approval with automatic
stale recovery disabled. Require empty, undescribed, conflict-free, non-divergent
@ directly above BASE, no references/descendants or other workspace sharing @.
If stale, active, differently based, or uncertain, stop and ask; do not repair,
reposition, or silently substitute.
