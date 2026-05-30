---
name: document-decisions
description: Use when documenting architecture or product decisions in docs/decisions as concise ADR-style records matching this repository's preferred format.
---

# Document Decisions

Use this skill when the user asks to document a decision, write an ADR, record an architectural/product choice, or add a file under `docs/decisions`.

## Location and filename

- By default, store decision records in `docs/decisions/YYYY-MM-DD_decision` unless another convention is used or is provided by the user.

## Format

Decision records should be simple and lightweight. Use [`assets/template.md`](./assets/template.md) as the template.

## Style

- Keep the record concise and practical.
- Preserve the user's wording and intent when they have already phrased the decision.
- Scope the decision narrowly. Do not broaden it into adjacent topics unless the user asks.
- Avoid process-heavy ADR sections such as status, consequences, owners, or dates unless the user explicitly asks for them.
- Use plain language over formal architecture-review language.
- Options considered can be terse; they do not need long pros/cons lists unless useful.
- The rationale should explain why the selected option was chosen, not restate all context.

## Before writing

- If the decision is unclear, ask one or two targeted clarification questions.
- If the user already provided the decision and context, write the file directly.
- Check existing files in `docs/decisions/` to match the repository's current style.

## After writing

- Tell the user the path of the created or updated decision record.
- Mention any assumptions only if they materially affected the text.
