# 2026-05-31 Cursor-format canonical rules

Decision: Rules will use Cursor's rule format as the canonical source, stored as `.md`, then be adapted into each provider or harness format as needed.

## Context

Rules are similar to agents: each harness has its own file locations, metadata, and activation model. Maintaining only native rule files would duplicate intent and make behavior drift likely.

Cursor's native extension is `.mdc`, but that is a Cursor-invented file format and extension that may not play well with programmatic access outside Cursor. The canonical source should therefore remain a Markdown file while following Cursor's rule structure.

## Options considered

1. Maintain separate native rule files for each harness.

2. Use Cursor's rule format, stored as `.md`, as the canonical rule definition and derive harness-specific files from it.

## Rationale for decision

Cursor's rule format is the most versatile common denominator for the harnesses we care about: Claude, Cursor, OpenCode, Pi, and Codex. It captures the rule body and activation metadata in one portable structure that can be translated into those harnesses' native formats without changing the underlying intent.

This keeps rules toolbelt-oriented: one base rule, many harness projections.
