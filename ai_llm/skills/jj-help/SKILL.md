---
name: jj-help
description: Use for Jujutsu (jj) version-control tasks, Git-to-jj translations, bookmarks, working-copy commits, revsets, workspaces (vs Git worktrees), or repos where jj is preferred.
metadata:
  author: "@BoscoDomingo"
---

# Jujutsu (jj) Help

## Use When

- User mentions `jj`, Jujutsu, bookmarks, working-copy commits, revsets, workspaces, or jj/git interop.
- Task involves version control in a repo where `jj` is preferred.
- Prompt asks for Git-style branch, staging, pull, push, amend, rebase, merge, or worktree workflows that need jj translation.
- Task involves multiple working copies, parallel/agent sandboxes, or anything that looks like Git worktrees (use jj workspaces instead).

## Required Workflow

1. Read [Jujutsu (jj) Quick Guide For Agents](./assets/agent-jujutsu-guide.md) before answering or running jj commands.
2. Use the guide as the command source of truth; do not rely on memory or user-local aliases.
3. Inspect state before proposing or running mutating version-control operations.
4. Translate Git-shaped intent into jj concepts using the guide.
5. Follow repo approval policy before any state-changing jj operation.
6. If unsure, stop after read-only inspection and ask the user.

## Integrated Example

User: "Create a branch and push this with jj."

Agent action: read the guide, translate "branch" to a bookmark workflow, inspect current state, ask for required approval before bookmark/push operations, then use built-in `jj` commands from the guide.

Benchmark history: [benchmark-runs.md](./assets/benchmark-runs.md).