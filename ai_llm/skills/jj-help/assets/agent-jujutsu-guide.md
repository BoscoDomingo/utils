# Jujutsu Quick Guide

Use Jujutsu (`jj`) for version control when it is available, except when the current checkout is a linked Git worktree. In a linked Git worktree, do not use `jj`; use basic Git commands instead.

This guide is meant for agents that know Git better than Jujutsu. Use built-in `jj` commands, not local aliases.

## Mental Model

- `jj` can use a Git repository as storage, but it has its own commit model and commands.
- Linked Git worktrees are an exception: if you are already inside one, stay in Git mode and do not run `jj`.
- Detached HEAD is normal. Do not treat it as a problem by itself.
- The working copy is a mutable commit named `@`.
- The parent of the working-copy commit is `@-`.
- Branch-like refs are called bookmarks.
- There is no Git-style staging area. Use filesets, `--interactive`, `jj split`, or `jj squash` instead.
- Most commands snapshot the working copy automatically before they run.
- A repo can have multiple working copies attached to it. These are workspaces, not Git worktrees. See the Workspaces section.
- Do not assume Git worktree behavior. jj workspaces are not branch-bound and share the full revision history.

## Safety Rules

- First determine whether the current directory is a linked Git worktree. If it is, do not use `jj`; use the Git Worktree Exception below.
- Prefer read-only inspection first: `jj status`, `jj diff`, `jj show`, `jj log`.
- Do not run `git checkout -b`, `git switch -c`, `git pull`, or `git push` unless you are in the Git Worktree Exception.
- Use `jj git fetch` instead of `git pull`.
- Use `jj git push` instead of `git push`.
- Use bookmarks instead of Git branches.
- Do not assume user aliases exist.
- Ask before mutating operations if the repo policy requires approval.
- Treat commits, rebases, splits, squashes, abandons, restores, bookmark moves, and pushes as history/state-changing operations.

## Git Worktree Exception

If the current checkout is a linked Git worktree (created with `git worktree`), DO NOT use `jj`. Treat it as a Git-owned workspace and use conservative/basic Git commands instead.

Detect a linked Git worktree with read-only Git inspection:

```sh
git rev-parse --is-inside-work-tree
git rev-parse --path-format=absolute --git-dir
git rev-parse --path-format=absolute --git-common-dir
```

If `--git-dir` and `--git-common-dir` differ, you are in a linked Git worktree. Another useful check is:

```sh
git worktree list --porcelain
```

In this mode:

- Do not run `jj status`, `jj log`, `jj new`, `jj workspace`, `jj bookmark`, `jj git fetch`, or other `jj` commands.
- Inspect with `git status --short --branch`, `git diff`, `git log --oneline --decorate -n 20`, and `git branch --show-current`.
- Commit with normal Git staging and commit commands (`git add`, `git commit`) if approval policy allows.
- Fetch with `git fetch`; avoid `git pull` unless the user explicitly asks and repo policy permits it.
- Push with `git push` only when explicitly requested/approved.
- Do not try to translate the worktree into a jj workspace. Finish or hand off using the Git worktree workflow already in use.

## Inspect State

```sh
jj status
jj diff
jj show
jj log
jj log --summary
jj op log
```

Useful revsets:

```sh
jj log -r '::'
jj log -r 'empty() & mutable()'
jj log -r 'divergent()'
jj log -r 'remote_bookmarks()..'
jj log -r '(remote_bookmarks()..@)::'
```

Meanings:

- `::` shows all ancestors and descendants reachable from visible heads.
- `empty() & mutable()` shows empty mutable commits.
- `divergent()` shows divergent changes.
- `remote_bookmarks()..` shows local work not reachable from remote bookmarks.
- `(remote_bookmarks()..@)::` shows non-pushed history related to the current working copy.

## Fetch

```sh
jj git fetch
```

Do not use `git pull`. In `jj`, fetch remote state first, then inspect/rebase/bookmark as needed.

## Commit Current Work

Commit all current working-copy changes:

```sh
jj status
jj diff
jj commit -m "message"
```

`jj commit` updates the description of the current working-copy commit, then creates a new empty working-copy commit on top.

## Commit Only Selected Files Or Hunks

There is no Git staging area. Do not use `git add`.

Commit only specific files/filesets:

```sh
jj status
jj diff
jj commit path/to/file -m "message"
```

Commit selected hunks interactively:

```sh
jj status
jj diff
jj commit --interactive -m "message"
```

With path arguments or `--interactive`, selected changes stay in the current commit and unselected changes move into a new working-copy commit on top.

## Split Mixed Work

Use `jj split` when one revision already contains mixed work and you need to separate it into two commits.

Interactively split the current working-copy commit:

```sh
jj status
jj diff
jj split
```

Split only specific files/filesets into the selected part:

```sh
jj split path/to/file
```

Split another revision:

```sh
jj split -r <rev>
```

Give the selected part a message without opening an editor:

```sh
jj split -m "message"
```

Default behavior: selected changes stay in the original commit, remaining changes go into a new child commit. Use `jj commit --interactive` for partial committing the working copy; use `jj split` for splitting an existing mixed revision.

## Describe A Change

Update a commit message without creating a new change:

```sh
jj describe -m "message"
jj describe -r <rev> -m "message"
```

Some repo policies treat `jj describe` as commit finalization when used to create/update meaningful history; ask first if approval is required.

## Squash / Amend Work

Move working-copy changes into the parent commit:

```sh
jj squash
```

Choose hunks interactively:

```sh
jj squash --interactive
```

Move changes into a specific revision:

```sh
jj squash --into <rev>
```

Automatically move changes into matching mutable commits:

```sh
jj absorb
```

Use `jj squash` for deliberate amend-like workflows. Use `jj absorb` when changes naturally belong to earlier commits in a stack.

## Edit Diffs

Open a diff editor for the current working-copy commit:

```sh
jj diffedit
```

`jj diffedit` is useful when interactive selection is easier in a diff editor than through command prompts.

## Move Around

Create a new working-copy commit on top of a revision:

```sh
jj new <rev>
```

Start new work on top of `main`:

```sh
jj new main
```

Edit an existing revision:

```sh
jj edit <rev>
```

Move to previous/next commit in the current stack:

```sh
jj prev
jj next
```

These commands move the working copy and may leave changes behind in commits. Ask first if policy requires approval.

## Rebase

Move only the current revision:

```sh
jj rebase -r @ -d <destination>
```

Move a whole branch/stack:

```sh
jj rebase --branch <rev> -d <destination>
```

Move descendants from a source:

```sh
jj rebase -s <source> -d <destination>
```

Interactive rebase:

```sh
jj rebase --interactive
```

Common example:

```sh
jj git fetch
jj rebase --branch @ -d main
```

Rebase rewrites history. Ask before running it if approval is required.

## Merge

In `jj`, a merge commit is created by making a new commit with multiple parents:

```sh
jj new <left> <right>
```

Merge `main` into current work:

```sh
jj new main @
```

This creates a new working-copy commit with both `main` and `@` as parents.

## Workspaces

Workspaces let you attach more than one working copy to the same repo. Each workspace is an independent directory on disk, but all workspaces share the same underlying repo, commits, and history.

### Mental Model

- A workspace is an additional, independent working-copy directory backed by the same repo storage.
- Each workspace has its own working-copy commit (`@`) and its own sparse patterns.
- In `jj log`, other workspaces' working-copy commits are shown as `<name>@` (e.g. `default@`, `feature-b@`). The current workspace's commit is the plain `@`.
- All workspaces see all commits. You can rebase, edit, or move any commit from any workspace, regardless of which workspace created it.
- There is no per-workspace branch binding. Workspaces are positions in one shared revision tree, not isolated branch checkouts.

### Workspaces vs Git Worktrees

Agents tend to know Git worktrees and wrongly map them onto jj workspaces. Key rule: if you are already inside a linked Git worktree, do not switch to jj; use basic Git commands for that checkout. Only use jj workspaces in repos that are already being managed with jj.

Key differences:

- A Git worktree is bound to a specific branch; a jj workspace is just another working copy pointing at any revision, with no branch binding.
- Git worktrees cannot share a branch (one branch, one worktree). jj workspaces can point at the same revision freely; two workspaces can sit on the same commit.
- To combine work across Git worktrees you merge branches. With jj you just `jj rebase` or `jj new` commits across the shared tree; there is no separate "merge the worktree" step.
- jj workspace conflicts are handled by jj's normal first-class conflicts (rebase records the conflict in the commit), not by a blocking merge step.
- Practically they serve similar goals (parallel work, long-running builds/tests, agent sandboxes), but do not assume branch-bound worktree semantics.
- Compatibility rule: do not operate on a Git linked worktree with `jj`. Use Git in Git worktrees; use `jj workspace` only for jj-managed workspaces.

### Inspect Workspaces

```sh
jj workspace list
jj workspace root
jj log
```

`jj log` from any workspace shows every workspace's working-copy commit, so you keep context across all of them.

### Create A Workspace

Workspaces should be created as siblings of the main repo directory, not nested inside it (nesting creates a confusing repo-inside-a-repo). Run `jj new` first so the new workspace starts from a shared parent.

```sh
jj new
jj workspace add ../feature-x
```

Name it explicitly (defaults to the destination directory's basename):

```sh
jj workspace add --name feature-x ../feature-x
```

Control the new workspace's parent revision(s):

```sh
jj workspace add -r <rev> ../feature-x
```

With no `-r`, the new workspace's working-copy commit shares the same parent(s) as the current workspace's working-copy commit. With one or more `-r`, it is created as if you had run `jj new r1 r2 ...`.

Control sparse patterns with `--sparse-patterns copy|full|empty` (default `copy`, inherits the current workspace's patterns).

### Work Across Workspaces

- `cd` into a workspace directory and use normal `jj` commands; they operate on the shared repo.
- Move a commit made in one workspace onto another workspace's line with `jj rebase`:

```sh
jj rebase -s <source> -d <destination>
```

- Advance a workspace to a given revision (e.g. to "catch up" the default workspace after rebasing):

```sh
jj edit <rev>
```

Two workspaces can legitimately point at the same revision; `jj log` will show both names on that commit.

### Remove A Workspace

Stop tracking a workspace, then delete its directory:

```sh
jj workspace forget <name>
rm -r ../feature-x
```

`jj workspace forget` only stops tracking the workspace's working-copy commit in the repo; it does not touch the directory on disk. Delete the directory separately. Forgetting does not discard commits that are reachable elsewhere in the tree.

### Rename

```sh
jj workspace rename <new-name>
```

### Stale Workspaces

If a workspace's working copy was left behind by operations run in another workspace, jj reports it as stale. Update it with:

```sh
jj workspace update-stale
```

### Safety Notes

- Creating, forgetting, renaming, and updating-stale workspaces change repo state; follow repo approval policy.
- Confirm which workspace you are in (`jj workspace root`, `jj workspace list`) before mutating, since all workspaces share history and a rebase here affects what others see.
- Prefer creating workspaces as siblings; warn the user before nesting a workspace inside the main repo.

## Bookmarks

List bookmarks:

```sh
jj bookmark list
```

Create a bookmark:

```sh
jj bookmark create <name> -r @
```

Move/set a bookmark:

```sh
jj bookmark set <name> -r @
```

Move a bookmark only if it can advance:

```sh
jj bookmark advance <name>
```

Delete a bookmark:

```sh
jj bookmark delete <name>
```

Use bookmarks for branch-like refs. Do not create Git branches.

## Push

Push all eligible changes according to repo config:

```sh
jj git push
```

Push a specific bookmark:

```sh
jj git push -b <bookmark>
```

Push tracked bookmarks:

```sh
jj git push --tracked
```

Create a new bookmark on push:

```sh
jj git push -N
```

Do not use `git push`. Pushing publishes history; ask first if approval is required.

## Restore / Abandon / Undo

Restore changes:

```sh
jj restore
```

Interactively restore changes:

```sh
jj restore --interactive
```

Abandon a revision:

```sh
jj abandon <rev>
```

Abandon empty mutable commits:

```sh
jj abandon -r 'empty() & mutable()'
```

Undo the last repo operation:

```sh
jj undo
```

These can discard or rewrite work. Ask first if approval is required.

## Conflict Resolution

Inspect conflicts:

```sh
jj status
jj diff
```

Resolve conflicts with the configured merge tool or editor:

```sh
jj resolve
```

After resolving files, inspect again:

```sh
jj status
jj diff
```

## Common Agent Workflows

Inspect before any change:

```sh
jj status
jj diff
jj log --summary
```

Commit all current work:

```sh
jj status
jj diff
jj commit -m "message"
```

Commit only one file:

```sh
jj status
jj diff
jj commit path/to/file -m "message"
```

Commit selected hunks:

```sh
jj status
jj diff
jj commit --interactive -m "message"
```

Split mixed work:

```sh
jj status
jj diff
jj split
```

Fetch and inspect remote-relative work:

```sh
jj git fetch
jj log -r 'remote_bookmarks()..'
```

Create branch-like bookmark and push:

```sh
jj bookmark create <bookmark> -r @
jj git push -b <bookmark>
```

Spin up a parallel working copy (workspace) for concurrent/agent work in a jj-managed repo:

```sh
jj new
jj workspace add ../feature-x
cd ../feature-x
# ...work, then bring it back onto the main line with jj rebase, then:
jj workspace forget feature-x
rm -r ../feature-x
```

If you are already in a Git linked worktree, do not run this workflow; stay with basic Git commands.

## Git Command Translation

| Git habit                  | Jujutsu command                                      |
|----------------------------|------------------------------------------------------|
| `git status`               | `jj status`                                          |
| `git diff`                 | `jj diff`                                            |
| `git log`                  | `jj log`                                             |
| `git show`                 | `jj show`                                            |
| `git add <file>`           | `jj commit <file> -m "message"` or `jj split <file>` |
| `git add -p`               | `jj commit --interactive -m "message"`               |
| `git commit -m "message"`  | `jj commit -m "message"`                             |
| `git commit --amend`       | `jj squash` or `jj describe`                         |
| `git checkout -b <branch>` | `jj bookmark create <bookmark> -r @`                 |
| `git switch <branch>`      | `jj new <bookmark>` or `jj edit <rev>`               |
| `git pull`                 | `jj git fetch`, then inspect/rebase                  |
| `git push`                 | `jj git push`                                        |
| `git rebase`               | `jj rebase`                                          |
| `git merge main`           | `jj new main @`                                      |
| `git worktree add <path>`  | `jj workspace add <path>` only in jj-managed repos; if already in a Git worktree, stay with Git |
| `git worktree list`        | `jj workspace list` for jj workspaces; `git worktree list` for Git worktrees |
| `git worktree remove`      | `jj workspace forget <name>` + `rm -r <path>` for jj workspaces; use Git for Git worktrees |

## Final Reminders

- Read repo policy before mutating state.
- If unsure, inspect with `jj status`, `jj diff`, `jj log`, and ask.
- Never rely on aliases from someone else's config.
- Avoid Git branch/staging mental models; use working-copy commits, bookmarks, filesets, split, and squash.
- Do not map Git worktrees onto jj workspaces; jj workspaces are not branch-bound and share the full revision tree.
- If already inside a linked Git worktree, do not use `jj`; use basic Git commands instead.
