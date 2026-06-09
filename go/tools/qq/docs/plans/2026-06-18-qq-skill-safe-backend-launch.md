# qq Skill-Safe Backend Launch Implementation Plan

**Goal:** Ensure `qq` invokes supported agent CLIs without skill auto-discovery where a supported, auth-safe mechanism exists.

**Architecture:** Keep backend behavior centralized in `internal/backend`. Extend the backend command model only as far as needed to support per-backend environment overrides, while preserving argv-only prompt passing and stdout/stderr passthrough.

**Tech stack:** Go.

**TDD during implementation:** enforce.

<details>
<summary>Context</summary>

**Prompt:** After investigating installed agent CLIs, use Claude `--safe-mode` instead of `--bare` because `--bare` breaks OAuth/keychain authentication, then draw an implementation plan.

**Documentation read:** `AGENTS.md`; attempted `docs/INDEX.md` (absent); `docs/decisions/2026-05-31-model-and-harness-agnostic-agents.md`; `go/tools/qq/README.md`; `go/tools/qq/docs/spec.md`; prior plans in `go/tools/qq/docs/plans/`; `lw-plan-implementation` skill; `validate-scope` checklist.

**Reasoning:** `qq` already owns backend argv rendering and runner execution in one small package. The smallest safe change is to add environment support to the command spec, update only backends with known skill-suppression controls, and prove behavior with fake-backend E2E tests before touching docs.
</details>

---

## Scope

### In scope

- Change Claude backend invocation to use auth-safe safe mode.
- Run OpenCode with its documented external-skill suppression environment.
- Add minimal command environment support in the backend runner.
- Add black-box tests proving argv/env behavior without real AI CLI calls.
- Update `README.md` and `docs/spec.md` where the user-facing contract changes.
- Record Cursor Agent's current limitation instead of claiming unsupported skill suppression.

### Out of scope

- Adding user configuration for backend mappings.
- Calling real AI backends in default tests.
- Changing backend priority order.
- Adding a generic skill-discovery abstraction across all possible agent tools.
- Editing historical plans or decision records.

## Acceptance Criteria

1. `qq -b claude "prompt"` invokes Claude in safe mode while preserving the prompt as one literal argv argument.
2. `qq -b opencode "prompt"` runs OpenCode with external skill discovery disabled through environment, while preserving existing argv behavior.
3. Default backend selection applies the same safe backend mapping when Claude or OpenCode is selected by priority.
4. Existing prompt behavior is unchanged for args-only, stdin-only, args plus stdin, and shell metacharacter prompts.
5. Success output remains backend stdout/stderr only; `qq` emits no new success banners or backend-choice text.
6. Backend failure behavior remains unchanged: an installed backend exiting non-zero stops without falling back.
7. Cursor Agent is not documented or treated as skill-auto-discovery-disabled unless a supported CLI/env control is found.
8. `go test -race ./...`, `go vet ./...`, and `go install .` pass from `go/tools/qq`.

## Test Ideas From Criteria

- E2E fake `claude`: assert captured argv includes safe mode, print mode, and the exact prompt as one argument.
- E2E fake `opencode`: assert captured environment contains the external-skill suppression variable and argv still includes `run` plus the exact prompt.
- E2E default priority: install only `claude` or only `opencode` and assert the same safe mapping is used without explicit `-b`.
- E2E stdin merge: pipe context into `qq -b opencode "summarize"` and assert both prompt merge and environment override.
- Existing regression tests: keep literal shell metacharacter, backend failure passthrough, and stdout/stderr passthrough coverage green.
- Unit tests in `internal/backend`: assert command specs copy args/env defensively and render prompt placeholders exactly once.

## Files And Areas

- `go/tools/qq/internal/backend/backend.go`
- `go/tools/qq/internal/backend/*_test.go`
- `go/tools/qq/e2e_test.go`
- `go/tools/qq/e2e_harness_test.go`
- `go/tools/qq/testdata/fake-backend`
- `go/tools/qq/README.md`
- `go/tools/qq/docs/spec.md`

## Implementation Chunks

| ID | Goal                                                       | Primary Paths                                               | Depends On | Parallel-Safe | Verify With                                           |
|----|------------------------------------------------------------|-------------------------------------------------------------|------------|---------------|-------------------------------------------------------|
| C1 | Add backend command environment support and tests          | `internal/backend/backend.go`, `internal/backend/*_test.go` | none       | No            | `go test ./internal/backend`                          |
| C2 | Update Claude and OpenCode backend specs                   | `internal/backend/backend.go`, backend tests                | C1         | No            | `go test ./internal/backend`                          |
| C3 | Extend fake E2E harness to capture backend environment     | `testdata/fake-backend`, `e2e_harness_test.go`              | C1         | No            | focused E2E tests                                     |
| C4 | Add black-box E2E coverage for skill-safe launch behavior  | `e2e_test.go`                                               | C2, C3     | No            | `go test -run TestE2E`                                |
| C5 | Update user-facing docs without duplicating backend tables | `README.md`, `docs/spec.md`                                 | C2         | Yes after C2  | doc review plus `go test ./...`                       |
| C6 | Run full verification                                      | whole `go/tools/qq`                                         | C1-C5      | N/A           | `go test -race ./... && go vet ./... && go install .` |

## Verification Strategy

- **Tests:** Write tests from acceptance criteria first; prefer E2E fake-backend tests for user-visible behavior, then backend unit tests for command rendering/env copying.
- **Commands:** Run targeted package tests after C1/C2, focused E2E after C4, then the full verification suite from `go/tools/qq`.
- **Manual / doc checks:** Confirm `README.md` no longer recommends a `claude --bare -p` quick alternative, and `docs/spec.md` continues to point to `internal/backend/backend.go` as the backend mapping source of truth.

## Risks

- Claude flag ordering could matter. Start with the CLI form selected by the user, and switch to `CLAUDE_CODE_SAFE_MODE=1` only if a real smoke check proves the flag form incompatible.
- OpenCode's environment variable disables external compatible skill directories, not every possible `.opencode` or configured remote skill source. Document that boundary precisely.
- Adding environment to `CommandSpec` can accidentally leak mutable slices between tests and runtime. Copy defensively like argv.
- Cursor Agent currently lacks a supported equivalent. Treat that as a limitation, not an implementation gap to paper over.

## Scope Validation

Lightweight baseline: user asked to draw an implementation plan after selecting Claude `--safe-mode` for skill-safe execution.

| Baseline objective                                        | Plan coverage                                              | Class |
|-----------------------------------------------------------|------------------------------------------------------------|-------|
| Draw an implementation plan                               | Adds a plan under the documented `qq` plans location       | PASS  |
| Use Claude `--safe-mode` instead of `--bare`              | Acceptance criteria, chunks, and risks center on safe mode | PASS  |
| Keep work tied to installed-tool skill-discovery findings | Covers Claude, OpenCode, and Cursor Agent limitation only  | PASS  |
| Plan, not implement behavior now                          | Adds plan document only                                    | PASS  |
