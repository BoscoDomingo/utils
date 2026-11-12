---
name: use-jj-workspaces
description: Create a fresh Jujutsu workspace for isolated or parallel work, or reuse one only on explicit request. Use for workspace isolation or the jj equivalent of a Git worktree.
compatibility: Requires jj and filesystem inspection tools. Check installed help for version-dependent options.
---

# Use jj Workspaces

Workspaces isolate files, not history: commits, bookmarks, and operations are
shared. Read applicable repository instructions and jj-help guidance first.
Use `jj workspace`; do not rely on local aliases such as `jj w`.

## Safety

- Obtain express approval before every jj mutation. Explain the base,
  destination, and exact commands. Proposal-only requests permit no writes.
  Never auto-remove or forget workspaces.
- Protect the source workspace: no `jj new`, edit, reset, restore, fetch,
  rebase, abandon, or bookmark movement as setup or cleanup. The only allowed
  source change is an approved commit of current work (step 2).

## Workflow

1. Inspect read-only with `jj --at-op=@ workspace root` and
   `jj --at-op=@ workspace list`; use filesystem inspection for sibling paths.
   Stop for ambiguous operation heads, Git-only worktrees, or conflicting policy.
2. Use the requested base; if unspecified, ask whether to include current work
   or use a named revision. Never base on the source's live `@`: it is rewritten
   as the source keeps changing, leaving the new workspace stale. To include
   current work, ask approval to commit it (`jj commit -m "<message>"`) and use
   the resulting `@-`. Pin exactly one full commit ID as BASE with
   `jj --at-op=@ log --no-graph -r '<revision>' -T 'commit_id ++ "\n"'`.
3. Default to fresh, even if an existing workspace looks reusable. Choose the
   lowest positive N with both name `copy-N` unused and sibling destination
   `<source-name>-copy-N` absent. Derive a quoted absolute path from the source
   root, not the shell directory. If workspaces are very outdated or more than a
   handful, recommend trimming; use judgement or ask if the count is acceptable.
   No hard count/age limit and no automatic cleanup.
4. Ask approval for BASE, name, path, and this command, run from the source
   workspace. Recheck name/path availability first; if occupied, reselect and
   confirm the changed proposal. Do not overwrite paths.

   ```sh
   jj workspace add --name "copy-N" -r "<BASE>" "<absolute-sibling-path>"
   ```

   Explicit `-r` avoids the default of sharing the source `@`'s parents. Respect
   colocation and sparse defaults.
5. Verify the destination root and that its `@`'s parent equals BASE before
   editing; run tools there. Report name/path and BASE. Leave files and
   registration intact at task end or on partial failure; report recovery needs
   instead of repairing them.

## Reuse Only By Request

Ask whether anyone or any background job is using the candidate; never reuse
the source. Verify its root, filesystem work (including untracked and ignored
files), sparse patterns, and colocation. Read-only recorded `@` alone cannot
prove cleanliness; any snapshot requires approval, with
`--config snapshot.auto-update-stale=false`. Require an empty, undescribed,
conflict-free `@` directly above BASE, with no bookmarks, descendants, or other
workspace sharing it. If stale, active, differently based, or uncertain, stop
and ask; do not repair, reposition, or silently substitute.
