---
name: test-gh-actions
description: Run GitHub Actions workflows locally with `act`. Discovers workflows, confirms selection with the user, then executes. Use when a change has been implemented that requires passing CI checks or when the user asks to test CI, validate a workflow, or run GitHub Actions locally.
compatibility: requires `act` (https://nektosact.com) and Docker
metadata:
  author: "@BoscoDomingo"
---

# Test GitHub Actions Locally

Run GitHub Actions workflows locally using [`act`](https://nektosact.com) without pushing to the remote.

---

## When to Use

- User asks to test CI locally or validate a workflow
- User wants to dry-run an action before pushing
- User wants to debug a failing GitHub Actions workflow
- After finishing a task where the code will have to pass CI checks

---

## Prerequisites

Verify `act` is available:

```bash
which act
```

If `act` is not found, **stop immediately** and tell the user to install it from <https://github.com/nektos/act>. Do NOT attempt to install it.

Verify Docker is running:

```bash
docker info >/dev/null 2>&1
```

If Docker is not reachable, stop and tell the user to start Docker first.

---

## Procedure

### Step 1 -- Discover workflows

List all workflow files:

```bash
ls .github/workflows/*.yml .github/workflows/*.yaml 2>/dev/null
```

If no files are found, inform the user there are no workflows to run and stop.

### Step 2 -- Infer relevant workflows

For each workflow file, read the `name:` and `on:` keys. Build a summary table:

| # | File | Name | Triggers | Jobs |
|---|------|------|----------|------|

Rank by relevance:
1. Workflows triggered by `push` or `pull_request` -- most likely what the user wants
2. Workflows with `workflow_dispatch` -- runnable but less common for local testing
3. Workflows triggered only by `schedule` or `release` -- least relevant

### Step 3 -- Confirm with the user

Present the summary table and ask the user which workflow(s) and/or job(s) to run. Default suggestion: all workflows from tier 1 above. If only one workflow exists, still confirm before running.

In Cursor, use `AskQuestion` to present the choices. In Claude Code, ask inline. Use similar tools for other agents.

### Step 4 -- Run

For each confirmed workflow, determine the event to simulate (use the first trigger in the `on:` list, preferring `push`).

```bash
act <event> -W .github/workflows/<file>
```

If the user selected specific jobs rather than full workflows:

```bash
act <event> -W .github/workflows/<file> -j <job_name>
```

If a `.secrets` or `.env` file exists at the repo root, pass it:

```bash
act <event> -W .github/workflows/<file> --secret-file .secrets
```

If `act` prompts for a Docker image size on first run, choose `Medium` unless the user specifies otherwise.

### Step 5 -- Report results

After execution, summarize:
- **Per job**: name, status (pass/fail), duration
- **On failure**: show the failing step name and its last ~30 lines of output
- **On success**: confirm all jobs passed

---

## Edge Cases

- **No Docker**: fail early with a clear message (Step 0 check handles this)
- **Secrets required**: if a workflow references `secrets.*` and no `.secrets` file exists, warn the user and ask if they want to create one or skip
- **Matrix builds**: `act` supports matrix strategies; no special handling needed, but warn the user that matrix runs multiply execution time
- **Composite/reusable actions**: `act` has limited support for local composite actions; if a run fails with a "composite action" error, inform the user of this known limitation
- **Large images**: first run may pull Docker images (~10-20 GB for full runner images); warn the user if this is the first `act` invocation in the repo
