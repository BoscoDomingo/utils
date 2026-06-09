# qq Cleanup And Dependency Upgrade Implementation Plan

**Goal:** Reorganize `go/tools/qq` into focused Go packages, split the monolithic test suite into E2E/integration/unit coverage, and upgrade dependencies to latest compatible versions.

**Architecture:** Keep `qq` self-contained inside `go/tools/qq`. Preserve `go install .` from that directory by leaving the executable entrypoint at the module root, with only `main.go` and wiring there. Move reusable behavior into `internal/*` packages for app orchestration, CLI wiring, backend registry/runner behavior, invocation parsing, and selector UI.

**Tech stack:** Go, Cobra, Bubble Tea v2, Bubbles v2, `golang.org/x/sys`.

**TDD during implementation:** enforce

<details>
<summary>Context</summary>

**Prompt:** Significant cleanup in `qq`: organize things into folders/packages, split `qq_test.go` into several files, ensure both E2E tests with mocked backend responses and unit tests via Dependency Injection, preferably move types/structs into their own files, then upgrade all libraries to latest versions because Bubbles has been on v2 for a while.

**Documentation read:** `AGENTS.md`, attempted `docs/INDEX.md` (absent), `docs/decisions/2026-05-31-model-and-harness-agnostic-agents.md`, `docs/plans/2026-06-17-qq-agent-cli-wrapper.md`, `go/tools/qq/README.md`, `go/tools/qq/spec.md`, `.cursor/plans/qq_cleanup_6b4db109.plan.md`, Bubbles v2 docs, and Bubble Tea v2 upgrade notes.

**Reasoning:** The current code already has useful DI (`Runner`, selector, path lookup, stdin/TTY options). The safest path is to add black-box E2E protection first, then extract packages and split tests while preserving behavior, then upgrade dependencies once the new structure gives focused regression coverage.
</details>

---

## Scope

### In Scope

- Keep the executable entrypoint installable with `go install .` from `go/tools/qq`.
- Split current flat `package main` code into focused packages under `go/tools/qq/internal`, leaving only root entrypoint wiring in `package main`.
- Split `qq_test.go` into focused E2E, integration, unit, runner, CLI, and backend mapping tests.
- Add `testdata` fake backend scripts/fixtures for real-binary E2E tests.
- Keep dependency-injected unit/integration tests for app behavior, backend lookup, runner, selector, stdin/TTY, and invocation parsing.
- Move structs/types into focused files near their owning behavior.
- Upgrade direct dependencies to latest compatible versions, including Charm v2 import paths/APIs.
- Update `go/tools/qq/README.md` and canonical `go/tools/qq/spec.md` for the new layout and dependency versions.

### Out Of Scope

- Changing backend priority or argv mappings except where required by existing behavior-preservation tests.
- Calling real AI backend CLIs or network services from tests.
- Rewriting historical docs in `docs/plans` or `docs/decisions`.
- Adding a repo-wide build system.

## Acceptance Criteria

1. `go install .` works from `go/tools/qq`, and `go install github.com/BoscoDomingo/utils/go/tools/qq@latest` remains the documented remote install command.
2. Existing user-visible behavior remains unchanged: prompt merging, backend priority, provider/backend flags, interactive selector trigger, completion generation, exit codes, stdout/stderr passthrough, and no fallback after an installed backend exits non-zero.
3. E2E tests run the real `qq` binary with `PATH` isolated to fake backend executables under `testdata`, with no real backend or network dependency.
4. E2E coverage includes args-only, stdin-only, args plus stdin, explicit backend, priority fallback across missing binaries, backend failure passthrough, no-backend error, and stdout/stderr passthrough.
5. E2E/unit coverage locks safety-critical behavior from `spec.md`: shell metacharacters pass as one literal argv argument, args-only calls never block on open or partially filled stdin pipes, unsupported provider semantics are preserved, empty prompts return usage, provider flags complete to supported backend names, and completion scripts still generate for `bash`, `zsh`, and `fish`.
6. Unit tests cover pure invocation parsing, prompt construction, backend command rendering, supported backend lookup, stdin read-mode decisions, and selector model behavior where practical.
7. DI-based integration tests cover app orchestration with fake path lookup, fake runner, fake selector, and controlled stdin/TTY state.
8. `qq_test.go` is removed or reduced to only intentional shared top-level tests; test helpers live in focused files or `internal/testutil`.
9. Types/structs such as app options, invocation result, backend command specs, runner errors, selector items, and selector model state live in focused files instead of large mixed files.
10. Bubbles and Bubble Tea use v2 module paths/APIs; Cobra, `x/sys`, and transitive deps are refreshed via Go module tooling.
11. `README.md` and `spec.md` reflect the new layout and point to code as the backend mapping source of truth without duplicating additional mapping details.

## Test Ideas From Criteria

- E2E: build or run the root `go/tools/qq` command in tests, prepend a temp fake-backend directory populated from `testdata`, assert exact stdout/stderr/exit code and captured backend argv.
- Integration with mocked externals: instantiate app with fake `LookPath`, fake runner, fake selector, and synthetic stdin/TTY values.
- Unit: test invocation parsing, stdin read-mode decisions, prompt merge separator, backend registry rendering, backend support lookup, selector model enter/cancel updates, unsupported provider errors, empty prompt usage, literal argv handling, and CLI completion results.
- Regression after dependency upgrade: cover Bubble Tea key handling for selector enter/cancel and verify completion commands still generate scripts.

## Files And Areas

- `go/tools/qq/main.go`
- `go/tools/qq/internal/app`
- `go/tools/qq/internal/backend`
- `go/tools/qq/internal/cli`
- `go/tools/qq/internal/invocation`
- `go/tools/qq/internal/selector`
- `go/tools/qq/internal/testutil`
- `go/tools/qq/testdata`
- `go/tools/qq/go.mod`
- `go/tools/qq/go.sum`
- `go/tools/qq/README.md`
- `go/tools/qq/spec.md`

## Implementation Chunks

| ID | Goal                                                                                   | Primary Paths                                            | Depends On | Parallel-Safe                 | Verify With                                                                                          |
|----|----------------------------------------------------------------------------------------|----------------------------------------------------------|------------|-------------------------------|------------------------------------------------------------------------------------------------------|
| C1 | Add real-binary E2E harness and `testdata` fake backend fixtures before moving code    | `qq_test.go`, `testdata`, new E2E test file              | none       | No                            | `cd go/tools/qq && go test ./...`                                                                    |
| C2 | Extract invocation parsing and prompt construction into a pure package with unit tests | `internal/invocation`, invocation tests                  | C1         | Yes, if app code is untouched | `go test ./internal/invocation`                                                                      |
| C3 | Extract backend registry, command rendering, runner, and backend errors with tests     | `internal/backend`, backend tests                        | C1         | Yes, if app code is untouched | `go test ./internal/backend`                                                                         |
| C4 | Extract app orchestration and stdin policy around injected dependencies                | `internal/app`, app integration tests                    | C2, C3     | No                            | `go test ./internal/app`                                                                             |
| C5 | Extract Cobra command construction while preserving root `go install .` entrypoint     | `main.go`, `internal/cli`, CLI/completion tests          | C4         | No                            | `go test ./... && go install .`                                                                      |
| C6 | Extract selector package without dependency upgrades                                   | `internal/selector`, selector tests                      | C5         | No                            | `go test ./internal/selector && go test ./...`                                                       |
| C7 | Migrate Charm dependencies to v2 APIs                                                  | `internal/selector`, `go.mod`, `go.sum`                  | C6         | No                            | `go get charm.land/bubbles/v2@latest charm.land/bubbletea/v2@latest && go mod tidy && go test ./...` |
| C8 | Refresh remaining direct deps and finish test split/helpers                            | `go.mod`, `go.sum`, all `*_test.go`, `internal/testutil` | C7         | No                            | `go get github.com/spf13/cobra@latest golang.org/x/sys@latest && go mod tidy && go test -race ./...` |
| C9 | Update canonical docs                                                                  | `README.md`, `spec.md`                                   | C8         | No                            | `go test -race ./... && go vet ./...`                                                                |

## Verification Strategy

- **Tests:** Written from acceptance criteria first; black-box preference is E2E, then integration with mocked externals, then unit tests for pure logic.
- **Commands:** Run targeted package tests after each extraction, then `cd go/tools/qq && go test ./...`, `go test -race ./...`, `go vet ./...`, and `go install .` before completion.
- **Dependency refresh:** Run `go get charm.land/bubbles/v2@latest charm.land/bubbletea/v2@latest`, adapt selector imports/APIs, then `go mod tidy`; after that run `go get github.com/spf13/cobra@latest golang.org/x/sys@latest`, `go mod tidy`, and full tests. If `@latest` requires a newer Go toolchain than the module can support, stop and report the compatibility constraint instead of pinning silently.
- **Manual / doc checks:** Confirm `README.md` keeps both remote install and from-source `go install .`, `spec.md` matches the new package layout, and docs refer to backend source files instead of duplicating backend mappings.

## Risks

- Bubble Tea/Bubbles v2 changes can affect selector method signatures, key message types, and `View` return values.
- Moving from flat `package main` to `internal/*` can tempt over-exporting. Keep exported APIs minimal and test package-local internals where appropriate.
- E2E tests can become host-dependent if PATH isolation is incomplete. Always use temp directories and fake executables.
- Stdin readiness behavior is fragile around open pipes. Preserve bounded no-blocking tests during the move.
- Backend docs can drift if mappings are duplicated. Keep backend registry code as source of truth.
- Go module `@latest` upgrades may require a newer Go version or introduce Charm v2 transitive changes. Treat that as an explicit compatibility decision, not an accidental downgrade.
