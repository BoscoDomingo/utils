---
name: update-docs
description: Update project documentation after implementation changes are ready to push. Use when finishing a feature branch, before creating a PR, after completing implementation tasks, or when the user says docs need updating.
metadata:
  author: "@BoscoDomingo"
---

# Update Documentation

Analyze recent changes and update all affected documentation. Code is the source of truth; docs exist only to orient readers & agents and reduce onboarding cost.

## Philosophy

- **Concise over expressive.** Every line must justify its token cost.
- **Point, don't duplicate.** Reference code locations instead of restating logic.
- **Agent-friendly.** Keep doc files small so loading them into context is cheap.
- **Additive only when necessary.** Prefer updating existing sections over creating new files.

## Workflow

### 1. Gather Context

Collect information by running these in parallel:

```bash
# Changes against base branch
git diff $(git merge-base HEAD main 2>/dev/null || git merge-base HEAD master 2>/dev/null)..HEAD --stat
git diff $(git merge-base HEAD main 2>/dev/null || git merge-base HEAD master 2>/dev/null)..HEAD

# Commit history for this branch
git log --oneline $(git merge-base HEAD main 2>/dev/null || git merge-base HEAD master 2>/dev/null)..HEAD

# Current state
git status --short
```

Also check for a plan file:

```bash
ls docs/plans/ 2>/dev/null
```

If a plan exists, read it for intended scope. If not, infer scope purely from the diff.

### 2. Launch Sub-Agent

Spawn a `generalPurpose` sub-agent with this prompt (fill in the `{placeholders}`):

````markdown
## Task: Update documentation for recent changes

### Context
{paste git diff --stat output here}
{paste git log --oneline output here}
{if a plan file exists, paste its contents here}

### Changed files
{paste git diff output or a summary if too large}

### Instructions

Update documentation affected by the changes above. Follow these rules strictly:

**What to check:**
1. `README.md` — update if usage, setup, dependencies, or project structure changed
2. `AGENTS.md` (root and any subproject) — update if agent-relevant behavior, commands, or structure changed
3. `docs/` folder — update relevant files if conventions, architecture, or workflows changed
4. `CHANGELOG.md` — append entry if the project maintains one (do not create one)
5. Code comments — only where logic is non-obvious and the diff introduced uncommented complexity

**How to write:**
- Maximum 1-2 sentences per item or change. Use sentence fragments where clear.
- Use tables, bullet lists, and code references over prose.
- Remove any doc content that is now stale or contradicted by the code.
- Do not explain what the code does line-by-line; summarize intent and point to the file/function.
- If a section grows beyond ~30 lines, split into a referenced file under `docs/`.
- AGENTS.md must stay under 200 lines (hard max 250).

**How NOT to write:**
- No "This change introduces..." or "We have added..." preambles.
- No restating of obvious type signatures, function names, or parameter lists.
- No speculative documentation for features not yet implemented.

**Output:**
For each file you update, report:
```
- <filepath>: <1-line summary of what changed>
```

If no documentation updates are needed, report: "No documentation updates required."
````

### 3. Review Sub-Agent Output

After the sub-agent completes:

1. Verify each updated file is under its size limit (AGENTS.md < 250 lines, other docs reasonable).
2. Spot-check that no verbose/duplicate content was added.
3. Report the summary to the user.

## Quick Reference

| Doc file      | Update when...                                     | Size guideline                           |
|---------------|----------------------------------------------------|------------------------------------------|
| README.md     | Usage, setup, deps, or structure changed           | Keep minimal; link to `docs/` for detail |
| AGENTS.md     | Commands, agent behavior, or project shape changed | < 200 lines (max 250)                    |
| docs/         | Conventions, architecture, or workflows changed    | Split at ~30 lines per section           |
| CHANGELOG.md  | Project already maintains one                      | 1-2 lines per entry                      |
| Code comments | Non-obvious logic was introduced                   | Inline only; no narration                |
