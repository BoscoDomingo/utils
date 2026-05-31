# 2026-05-31 Model- and harness-agnostic agents

Decision: Agents will be model- and harness-agnostic. Each agent will have a canonical source definition with enough metadata to derive native files for each harnesses. 

## Context

We want to add custom subagents to this repo. However, no common spec exists yet, so the target harnesses do not share a single agent-definition format.
The harnesses/tools we plan to support (OpenCode, Claude Code, VS Code/Copilot, Cursor, Codex, and Pi) each have their own expected locations, frontmatter, configuration fields, and capability models. Some support the same agent as both a primary agent and a subagent; others primarily expose subagent-style delegation.

Thus, we need to find a common ground.

## Options considered

1. Maintain only native files for each harness.

2. Use one canonical agent definition where adapted files can be derived from.

## Rationale for decision

A canonical agent definition gives us one source of truth for the agent's identity, description, availability, permissions, target paths, and prompt. Native adapter files can then express the same agent in the format each harness expects. 

This keeps agents intentionally invocable rather than turning them into global instructions. It also lets us support primary-agent usage where available while preserving subagent availability as the baseline requirement.
