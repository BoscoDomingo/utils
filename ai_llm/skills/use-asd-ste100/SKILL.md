---
name: use-asd-ste100
description: >
  Force agent to use ASD-STE100 style for rewriting or producing clear user-facing
  output. Never apply it to internal reasoning, tool calls, code, or quoted
  content.
disable-model-invocation: true
---

# ASD-STE100 Simplified Technical English

Apply this Skill only to visible output intended for the user. Do not apply it
to internal reasoning, hidden analysis, tool calls, commands, code, identifiers,
URLs, structured data, exact errors, or quoted text unless the user asks for
that transformation.

## Core rules

- Use one meaning for each word.
- Use one part of speech for each word.
- Use the same term for the same action.
- Use active voice for instructions.
- Use simple tenses.
- Put one instruction in each sentence.
- Keep instructions to about 20 words and descriptions to about 25 words.
- Keep noun clusters to three words or fewer.
- Keep the subject, verb, and article explicit.
- Use one topic per paragraph.
- Use lists for three or more steps or conditions.
- Define technical terms that are not common English.

Preserve every fact, condition, scope qualifier, uncertainty, safety warning,
number, and required technical term. If a shorter rewrite would lose precision,
keep the longer wording and explain the trade-off.

## Rewrite process

1. Read the input once to understand its meaning.
2. Identify each ambiguity or rule violation.
3. Rewrite the text without changing its meaning.
4. Check that the rewrite preserves all required detail.

When the user asks you to rewrite supplied text, use this format:

| Rule violated | Original | Simplified |
|---|---|---|
| Present perfect tense | "We have received your request." | "We received your request." |

Follow the table with one line about any wording that you did not simplify.
Explain when simplification would remove required precision.

When the user asks for a normal response in ASD-STE100 style, provide the
response directly. Do not add a before-and-after table unless the user asks for
one.

## Boundaries

This Skill is not for creative, marketing, or persuasive text when a specific
voice is important. Follow the user's latest explicit style request when it
conflicts with this Skill.

For exact approved wording, use the
[official ASD-STE100 site](https://www.asd-ste100.org/). This Skill applies the
standard's public principles. It is not a substitute for the official
standard.

## References

- [writing_rules.md](references/writing_rules.md): public rule summary.
- [why_ASD-STE100.md](references/why_ASD-STE100.md): background and scope.
- [sources.md](references/sources.md): source and licence information.
- [before_after.md](assets/examples/before_after.md): rewrite examples.
