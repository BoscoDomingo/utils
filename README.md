# utils
Where I keep all the various utils I've created/found along my journey as a SWE

## Isolation Skills

- [use-jj-workspaces](ai_llm/skills/use-jj-workspaces/SKILL.md): Create fresh workspaces above an explicit, committed base; reuse only on request. Recommend trimming very outdated or numerous workspaces using judgement, not hard limits; mutations and cleanup require express approval.
- [use-sandbox](ai_llm/skills/use-sandbox/SKILL.md): Offline Linux CLI execution uses Bubblewrap with a trusted minimal runtime, timeout supervision, and a required per-invocation preflight; Windows is documentation-only and must not launch. The Linux command skeleton and platform integration tests remain unexecuted.
- Evals: [workspace scenarios](ai_llm/skills/use-jj-workspaces/evals/eval.yaml) and [sandbox scenarios](ai_llm/skills/use-sandbox/evals/eval.yaml) have not been run. Keyword graders are smoke checks; review traces for workspace state and sandbox isolation.
