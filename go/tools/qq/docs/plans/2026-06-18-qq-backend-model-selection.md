# qq Backend Model Selection Implementation Plan

**Goal:** Implement backend-wide model selection for `qq` from the 2026-06-18 handoff and decision record.

**Architecture:** Keep execution behavior centralized in `internal/backend`, replacing the current static command-template shape with a small execution interface plus optional model-discovery capabilities. Normal prompt execution must not depend on model discovery.

**Tech stack:** Go.

**TDD during implementation:** enforce.

<details>
<summary>Context</summary>

**Prompt:** With go/tools/qq/docs/qq-model-selection-handoff-2026-06-18.md do /lw-plan-implementation

**Documentation read:** `AGENTS.md`; attempted `docs/INDEX.md` (absent); `go/tools/qq/AGENTS.md`; `go/tools/qq/docs/spec.md`; `go/tools/qq/docs/qq-model-selection-handoff-2026-06-18.md`; `go/tools/qq/docs/decisions/2026-06-18-pi-backend-model-selection-policy.md`; existing plan `go/tools/qq/docs/plans/2026-06-18-qq-skill-safe-backend-launch.md`; `lw-plan-implementation` plan template; `validate-scope` checklist.

**Reasoning:** The ADR makes model choice a backend-wide capability, not a Pi-only patch. The smallest coherent change is to narrow supported first-class backends, pass model overrides through the existing CLI/app/backend path, keep model listing optional, and prove the argv contract with fake-backend tests before updating docs.
</details>

---

## Scope

### In scope

- Reduce first-class backend support to Claude, Cursor, OpenCode, Pi, Codex, and Gemini.
- Add per-invocation model selection via `--model` and `-m`, passed to the selected backend using native argv.
- Replace the temporary Pi hard-pin with explicit model override handling. Without a model override, pass no model argv and preserve each backend's own default behavior.
- Introduce the ADR shape: `LLMInfo`, execution `Backend`, optional `ModelLister`, and optional `ModelProviderLister`.
- Add model rendering tests for no override and explicit override behavior across all six supported backends.
- Add model listing/parsing tests for Cursor, OpenCode, and Pi where output shape is known.
- Update `README.md` and `docs/spec.md` after behavior changes without duplicating backend mappings.

### Out of scope

- Prompt difficulty classification or automatic model routing.
- Live backend prompts in default tests.
- Automatic fallback after provider, auth, quota, or backend execution errors.
- Canonicalizing provider or model spelling in `LLMInfo`.
- Renaming the existing Pi-specific ADR filename unless separately requested.
- Editing historical plans or decision records.

## Acceptance Criteria

1. Supported first-class backends are exactly Claude, Cursor, OpenCode, Pi, Codex, and Gemini.
2. `qq --model <selector>` and `qq -m <selector>` pass the exact selector string to the chosen backend using that backend's native model argument.
3. `qq "question"` still selects the first installed backend, passes no model argv, and does not run model discovery.
4. Existing prompt behavior is unchanged for args-only, stdin-only, args plus stdin, and shell metacharacter prompts.
5. Prompt text remains one literal argv argument; no shell interpolation, shell-built command strings, or `eval` are introduced.
6. With no `--model` or `-m`, no backend receives a model selector from `qq`; backend-native defaults remain responsible.
7. Backend argv rendering is covered for Claude, Cursor, OpenCode, Pi, Codex, and Gemini with no model override and with an explicit model override.
8. Model discovery remains optional; execution works for backends that do not implement `ModelLister` or `ModelProviderLister`.
9. Cursor, OpenCode, and Pi model listing/parsing preserve backend-native provider/model spelling and casing.
10. Shell completion reflects the reduced backend list and registers the model flag without changing backend/provider flag semantics.
11. Success output remains backend stdout only, with no wrapper banners or backend-choice text.
12. `README.md` and `docs/spec.md` describe the user-facing model flag, no-override behavior, and reduced backend support while linking to source for authoritative mappings.
13. `gofmt`, `go test -race ./...`, `go vet ./...`, and `go install .` pass from `go/tools/qq`.

### Resolved Audit Decisions

- Model override flags are exactly `--model` and `-m`.
- Without `--model` or `-m`, `qq` passes no model arg to any backend. No Pi default selector or fallback selector is introduced by this plan.
- With `--model` or `-m`, `qq` passes the exact user-supplied selector only through the chosen backend's execution path. This plan treats all six first-class backends as supporting explicit model override; if implementation evidence narrows that support, update this plan and tests before implementation.

## Test Ideas From Acceptance Criteria

- Root fake-backend E2E: `qq -b pi -m github-copilot/gpt-5.5 "hi"` records Pi native model argv with the exact selector and exact prompt as one argument.
- Root fake-backend E2E: normal `qq "hi"` records no model argv and no model-discovery subprocess before backend execution.
- Existing E2E regressions: stdout passthrough, stderr passthrough, non-zero backend stop, stdin merge, and shell metacharacter behavior stay green.
- Backend tests: supported names and priority contain only the six first-class backends.
- Backend tests: all six supported backends render expected argv for nil model with no model selector and explicit `LLMInfo` with the exact selector.
- CLI tests: `--model` and `-m` parse identically and do not consume prompt text incorrectly.
- CLI/completion tests: backend/provider completion uses the reduced list; model flag is registered.
- Parser tests: Cursor, OpenCode, and Pi sample outputs produce expected `LLMInfo` values without normalizing casing.
- Capability tests: execution path accepts backends without model-listing interfaces.

## Files And Areas

- `go/tools/qq/internal/backend/backend.go`
- `go/tools/qq/internal/backend/backend_test.go`
- `go/tools/qq/internal/backend/test_helpers_test.go`
- `go/tools/qq/internal/cli/command.go`
- `go/tools/qq/internal/cli/command_test.go`
- `go/tools/qq/internal/cli/completion.go`
- `go/tools/qq/internal/app/app.go`
- `go/tools/qq/internal/app/options.go`
- `go/tools/qq/internal/app/app_test.go`
- `go/tools/qq/internal/invocation/invocation.go`
- `go/tools/qq/internal/invocation/invocation_test.go`
- `go/tools/qq/e2e_test.go`
- `go/tools/qq/e2e_harness_test.go`
- `go/tools/qq/e2e_real_backend_test.go`
- `go/tools/qq/README.md`
- `go/tools/qq/docs/spec.md`

## Implementation Chunks

- Chunk 1: Add failing tests for the reduced backend set, no-override behavior, exact model override argv rendering, and `--model` / `-m` parsing.
  Verify by running `go test ./internal/backend ./internal/cli ./internal/app ./internal/invocation .` and observing failures tied to the new semantics.

- Chunk 2: Replace static backend command templates with `LLMInfo`, execution `Backend`, and optional model-listing capability interfaces.
  Verify with `go test ./internal/backend`.

- Chunk 3: Reduce `supportedBackends` and backend priority to Claude, Cursor, OpenCode, Pi, Codex, and Gemini.
  Verify with `go test ./internal/backend ./internal/cli`.

- Chunk 4: Add `--model` and `-m` CLI plumbing and pass the selected model through app/invocation/backend execution.
  Verify with `go test ./internal/cli ./internal/app ./internal/invocation`.

- Chunk 5: Implement backend-native model argv rendering for all six supported backends. Remove the temporary Pi hard-pin; no override must pass no model argv, and explicit override must pass the exact selector.
  Verify with `go test ./internal/backend .`.

- Chunk 6: Add optional model listing/parsing support for Cursor, OpenCode, and Pi behind capability interfaces.
  Verify with `go test ./internal/backend`.
  Parallel-safe after Chunk 2 only if isolated to new parser helpers and coordinated with Chunk 5.

- Chunk 7: Update shell completion behavior for the reduced backend list and model flag registration.
  Verify with `go test ./internal/cli`.

- Chunk 8: Update `README.md` and `docs/spec.md` for the reduced backend set and model-selection UX.
  Verify by doc review plus `go test ./...`.
  Parallel-safe after Chunks 3-7 stabilize.

- Chunk 9: Run final cleanup and full verification from `go/tools/qq`.
  Verify with `gofmt -w .`, `go test -race ./...`, `go vet ./...`, and `go install .`.

## Parallelization Notes

- Serialize Chunks 1-5 because they all touch model semantics across `internal/backend`, CLI option flow, and root E2E contracts.
- Chunk 6 can proceed after Chunk 2 only if parser helpers are isolated from Chunk 5's argv rendering work.
- Chunk 7 can proceed after Chunk 3 and Chunk 4 stabilize; it touches `internal/cli` and should not run concurrently with CLI flag plumbing.
- Chunk 8 can proceed after Chunks 3-7 stabilize; docs must reflect source behavior without duplicating backend mappings.

## Verification Strategy

- **Tests:** Write tests from acceptance criteria first; prefer root fake-backend E2E for user-visible behavior, integration tests for CLI/app option flow, and unit tests for backend argv rendering and model parsers.
- **Commands:** Run targeted package tests after each chunk, then full verification from `go/tools/qq`:

```bash
gofmt -w .
go test -race ./...
go vet ./...
go install .
```

- **Manual / doc checks:** Do not run live model prompts by default. Real backend E2E remains opt-in with `QQ_REAL_BACKEND_E2E=1`. Confirm docs link to `internal/backend/backend.go` for authoritative backend mappings rather than duplicating source truth.

## Risks

- Removing long-tail backends is intentionally breaking for users relying on them.
- CLI model-list output can drift by installed version; keep discovery optional and parser tests sample-based.
- Pi model selectors must stay provider-qualified where needed, or the original OpenAI quota path can return.
- Model flag plumbing may collide with backend/provider optional-value parsing if Cobra arg handling is changed carelessly.
- Docs can drift if they restate backend tables; keep source authoritative.

## Scope Validation

Lightweight baseline: `With go/tools/qq/docs/qq-model-selection-handoff-2026-06-18.md do /lw-plan-implementation`

- PASS: Uses the specified handoff as planning input.
- PASS: Produces a lightweight implementation plan under the documented `qq` plans location.
- PASS: Includes acceptance criteria before implementation chunks.
- PASS: Derives black-box test ideas from acceptance criteria.
- PASS: Lists files/areas, verification strategy, risks, and docs read.
