---
name: commit
description: Use when the user asks to commit changes, create commits, split work into logical commits, or run /commit. Inspects Git/Jujutsu state, follows existing commit conventions, and commits only intended changes.
---

# Commit

Use this skill when the user asks to commit changes, create commits, split work into logical chunks, or run `/commit`.

## Goals

- Create clean, logical commits that match the repository's existing conventions.
- Prefer Jujutsu only when it is available and there are signs the repo uses it.
- Preserve unrelated user or concurrent-agent changes.
- Avoid committing secrets or unintended files.

## Initial inspection

Before committing, inspect:

- `git status --short`
- `git diff`
- `git diff --cached`
- `git log --oneline -10`
- Whether `jj` is installed with `command -v jj`
- Whether the repo appears Jujutsu-backed, such as a `.jj/` directory or successful read-only `jj status`

Use read-only Jujutsu commands freely when helpful: `jj status`, `jj diff`, `jj log`, `jj show`, and `jj op log`.

## Choosing Git or Jujutsu

- Use Jujutsu when `jj` is available and the checkout is a Jujutsu repo.
- Use Git when Jujutsu is unavailable, `jj status` reports no repo, or there are no signs Jujutsu has been used here.
- If using Jujutsu would require a mutating operation, ask for explicit approval before running it. Mutating Jujutsu operations include commits, working-copy movement (`jj new`, `jj edit`, `jj next`, `jj prev`), rebases, squashes, abandons, bookmark changes, undo, and `jj git push`.

## Commit conventions

- Follow the repository's recent commit style from `git log --oneline -10`.
- If no clear convention exists, use Conventional Commits by default.
- Keep messages concise and specific.
- Prefer scopes when they improve clarity, such as `docs(decisions): ...` or `feat(rules): ...`.

## Splitting commits

- Break changes into logical chunks, not file-by-file busywork.
- Separate documentation/process decisions from executable or behavioral changes when that makes the history clearer.
- Keep generated harness-specific artifacts with the source artifact that makes them useful, unless the user asks otherwise.
- Do not stage unrelated files or changes that were not part of the requested work.

## Git workflow

1. Review the complete working tree and staged state.
2. Decide the logical commit groups.
3. Stage only the files for the next group with explicit paths.
4. Inspect `git diff --cached --stat` and the staged diff for that group.
5. Commit with a message matching repo conventions.
6. Repeat for each logical group.
7. Finish with `git status --short` and `git log --oneline -N`, where `N` covers the new commits.

## Jujutsu workflow

1. Use read-only inspection to understand the working copy and operation history.
2. Explain the intended logical commit split.
3. Ask for explicit approval before any mutating Jujutsu operation.
4. After approval, create commits using the minimal mutating operations needed.
5. Finish with read-only status/log verification.

## Safety rules

- Never use destructive commands such as `git reset --hard`, `git checkout --`, `jj undo`, or abandoning changes unless the user explicitly asks and approves.
- Never amend existing commits unless the user explicitly asks.
- Never force-push unless the user explicitly asks and approves.
- Never run `git commit`, `jj commit`, or any equivalent commit-creating operation without first showing what will be committed and receiving user approval.
- If hooks fail, fix the issue and create the commit normally; do not bypass hooks.
- If unrelated changes are present, leave them unstaged and mention them in the final summary.
- If a file contains both requested and unrelated changes, either stage the whole file only when all changes belong together or ask before doing partial staging.

## Final response

Report:

- Which VCS was used and why.
- The commits created, including short hashes and messages.
- Whether the working tree is clean or what remains uncommitted.
