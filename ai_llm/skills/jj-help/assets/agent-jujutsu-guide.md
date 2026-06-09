# Jujutsu Quick Guide

Use Jujutsu (`jj`) for version control when it is available. Do not use Git branch workflows unless the user explicitly asks.

This guide is meant for agents that know Git better than Jujutsu. Use built-in `jj` commands, not local aliases.

## Mental Model

- `jj` can use a Git repository as storage, but it has its own commit model and commands.
- Detached HEAD is normal. Do not treat it as a problem by itself.
- The working copy is a mutable commit named `@`.
- The parent of the working-copy commit is `@-`.
- Branch-like refs are called bookmarks.
- There is no Git-style staging area. Use filesets, `--interactive`, `jj split`, or `jj squash` instead.
- Most commands snapshot the working copy automatically before they run.

## Safety Rules

- Prefer read-only inspection first: `jj status`, `jj diff`, `jj show`, `jj log`.
- Do not run `git checkout -b`, `git switch -c`, `git pull`, or `git push`.
- Use `jj git fetch` instead of `git pull`.
- Use `jj git push` instead of `git push`.
- Use bookmarks instead of Git branches.
- Do not assume user aliases exist.
- Ask before mutating operations if the repo policy requires approval.
- Treat commits, rebases, splits, squashes, abandons, restores, bookmark moves, and pushes as history/state-changing operations.

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

## Final Reminders

- Read repo policy before mutating state.
- If unsure, inspect with `jj status`, `jj diff`, `jj log`, and ask.
- Never rely on aliases from someone else's config.
- Avoid Git branch/staging mental models; use working-copy commits, bookmarks, filesets, split, and squash.
