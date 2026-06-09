# jj-help Benchmark Runs

## 2026-06-09 Skill Activation Check

Purpose: verify the updated `jj-help` skill does not regress common Jujutsu agent workflows.

Scope:

- Model/harness: Cursor default subagent model.
- Conditions: without skill content vs with `jj-help` skill and linked quick guide.
- Method: three prompt scenarios, scored against guide-derived command and safety criteria.
- Limitation: this is a smoke benchmark, not a multi-model release gate.

Scenarios:

1. "Create a branch and push this with jj."
2. "Commit only package.json with jj and leave my other changes alone."
3. "Pull latest main and rebase my current stack with jj."

Score:

- Without skill: 10/15.
- With skill: 15/15.
- Delta: +5.

Findings:

- Scenario 1 improved from a likely-invalid push flag (`jj git push --bookmark <name>`) to the guide command `jj git push -b <name>`.
- Scenario 2 improved from a split/describe flow to the direct selected-file command `jj commit package.json -m "<message>"`.
- Scenario 3 improved from uncertain remote/revset handling to the guide flow: `jj git fetch`, inspect, then `jj rebase --branch @ -d main` after approval.
- No with-skill regression observed.
- No universal failure observed.

Go/no-go:

- Go for current `SKILL.md` change.
- Follow-up needed if this skill is promoted as broadly shared: rerun with at least three model families and record per-model deltas.
