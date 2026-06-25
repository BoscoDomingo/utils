---
name: test-gh-actions
description: Run GitHub Actions workflows locally with `act`. Discovers workflows, confirms selection with the user, then executes. Use when a change has been implemented that requires passing CI checks or when the user asks to test CI, validate a workflow, or run GitHub Actions locally.
compatibility: requires `act` (https://nektosact.com) and `docker`
metadata:
  author: "@BoscoDomingo"
---

# Test GitHub Actions Locally

## When to Use

- User asks to test CI locally or validate a workflow
- User wants to dry-run an action before pushing
- User wants to debug a failing GitHub Actions workflow
- After finishing a task where the code will have to pass CI checks

## Prerequisites

`act` ships in two officially supported forms — pick whichever is available:

```bash
# Prefer the standalone binary; fall back to the gh CLI extension.
if command -v act >/dev/null 2>&1; then
  ACT_CMD="act"
elif gh act --version >/dev/null 2>&1; then
  ACT_CMD="gh act"
else
  ACT_CMD=""
fi
```

If `ACT_CMD` is empty, **stop immediately** and tell the user to install one of:

- Standalone binary: <https://github.com/nektos/act>
- GitHub CLI extension: install [GitHub CLI](https://cli.github.com/) first if you don't already have `gh`, then run `gh extension install nektos/gh-act` (see <https://github.com/nektos/gh-act>)

Do NOT attempt to install either yourself.

Use `$ACT_CMD` (the chosen invocation) in every example below — the flags are identical between the two forms.

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

Use `AskQuestion`, `AskUserQuestion` or similar tools to present the choices.

### Step 4 -- Run

For each confirmed workflow, determine the event to simulate (use the first trigger in the `on:` list, preferring `pull_request`).

`$ACT_CMD` below is the invocation chosen in Prerequisites (either `act` or `gh act`).

```bash
$ACT_CMD <event> -W .github/workflows/<file>
```

If the user selected specific jobs rather than full workflows:

```bash
$ACT_CMD <event> -W .github/workflows/<file> -j <job_name>
```

If a `.secrets` or `.env` file exists at the repo root, pass each via its matching flag (`--secret-file` for secrets, `--env-file` for environment variables):

```bash
# Include only the flags whose corresponding file actually exists at the repo root.
$ACT_CMD <event> -W .github/workflows/<file> --secret-file .secrets --env-file .env
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
