# qq Agent CLI Wrapper Implementation Plan

Status: superseded by [`ai_llm/qq/spec.md`](../../go/tools/qq/spec.md).

## Documentation Reviewed

- `AGENTS.md`
- `README.md`
- Attempted `docs/INDEX.md` (absent)
- `docs/decisions/2026-05-31-model-and-harness-agnostic-agents.md`
- `ai_llm/qq/bin/qq`
- `ai_llm/qq/tests/test_qq.py`
- `ai_llm/qq/README.md`
- `.gitignore`
- `.tmp/lightweight-iceye-workflow/20260617T145146Z_qq_agent_cli_wrapper/qq-agent-cli-handoff-20260617-1747.md`
- `lw-plan-implementation` skill
- `validate-scope` checklist

Repo findings:

- No `docs/INDEX.md`, `docs/**/INDEX.md`, or `tools/**/docs/INDEX.md` exists.
- No Go or Rust tooling was found anywhere in the repo: no `go.mod`, `Cargo.toml`, `*.go`, or `*.rs`.
- No directory-specific `AGENTS.md` exists under `ai_llm/qq`.
- Existing repo style is self-contained utility folders, with no unified build system.
- Existing `qq` production code is already Bash; the Python dependency is only the current test file.

## Scope Validation

Lightweight baseline: user's request to plan, not implement, the `qq` shell helper.

| Baseline objective                      | Plan coverage                                                          | Class |
|-----------------------------------------|------------------------------------------------------------------------|-------|
| Plan only; no production implementation | Plan file plus state update only                                       | PASS  |
| Identify likely home, rules, tooling    | Recommends self-contained `ai_llm/qq` and notes no local shell tooling | PASS  |
| Tests first                             | First chunk writes black-box tests before script                       | PASS  |
| Fixed backend priority list             | Acceptance criteria and chunks use supplied v1 list                    | PASS  |
| Piping and stdin merge                  | Acceptance criteria and tests cover args-only, stdin-only, and both    | PASS  |
| Conservative product defaults           | Silent backend, fallback only on missing command, clear separator      | PASS  |
| Pitfall analysis                        | Dedicated pitfalls section                                             | PASS  |
| Update ephemeral state                  | State update planned with phase status and plan path                   | PASS  |
| Python removal                          | Replace Python tests with Go stdlib tests and remove test exceptions   | PASS  |

## Confirmed Product Decisions

- Location: create a self-contained helper under `ai_llm/qq/`.
- Shape: small executable script at `ai_llm/qq/bin/qq`, plus tests and a short README.
- Shell integration: users can symlink or alias the script; do not make v1 only a shell function.
- Backend selection: stay silent on success. Print selected backend only in verbose/debug mode later,
  not in v1 default output.
- Fallback: try next backend only when the backend binary is missing. If an installed backend exits
  non-zero, stop and return that failure so auth, quota, config, or model errors are not hidden.
- Prompt merge:
  - args only: prompt is joined CLI args.
  - stdin only: prompt is stdin.
  - args plus stdin: prompt is args, blank line, `Context:`, blank line, stdin.
- Amazon Q Developer CLI is included in v1 as the first Tier 2 backend:
  `q chat --non-interactive "prompt"`. The helper name remains `qq`.

## No-Python Implementation Decision

Superseded on 2026-06-17: `qq` is now moving to a Go runtime using Cobra for
CLI/completion support and Bubble Tea/Bubbles for interactive provider
selection. Python remains out of scope for `ai_llm/qq`; this section is kept as
historical planning context only.

Recommendation: keep `ai_llm/qq/bin/qq` as the Bash runtime and replace the Python black-box tests
with Go stdlib tests.

Rationale:

- The utility is shell-shaped: PATH lookup, TTY/stdin handling, argv dispatch, and stdout/stderr
  passthrough are simpler and more transparent in Bash than in a compiled binary.
- Python is not needed for runtime or tests. Go's `testing`, `os/exec`, `os`, and `t.TempDir`
  cover the black-box test harness without third-party packages.
- Go is the better no-Python test harness than Rust here because it has less project scaffolding,
  faster compile/test cycles, and simpler temp-process tests for a tiny CLI wrapper.
- Rust would be reasonable if the runtime were ported to a compiled binary for distribution, but
  that is needless complexity for v1 and would add Cargo layout with no existing repo pattern.

Chosen layout:

- Keep shell wrapper: `ai_llm/qq/bin/qq`.
- Add local Go module only for tests: `ai_llm/qq/go.mod`.
- Add Go stdlib black-box tests: `ai_llm/qq/qq_test.go`.
- Remove Python test file: `ai_llm/qq/tests/test_qq.py`.
- Update `.gitignore` to unignore `go.mod` and `qq_test.go`, and stop unignoring the removed
  Python test path.

Planned file changes:

- Add: `ai_llm/qq/go.mod`, `ai_llm/qq/qq_test.go`.
- Change: `ai_llm/qq/bin/qq` only if Go tests expose a behavioral gap.
- Change: `ai_llm/qq/README.md`, `.gitignore`.
- Remove: `ai_llm/qq/tests/test_qq.py`. Remove `ai_llm/qq/tests/` too if it becomes empty.

## Acceptance Criteria

1. `qq "question"` prints only the selected backend's answer to stdout on success.
2. `echo "context" | qq "question"` reads stdin in the wrapper and sends one merged prompt to the
   backend.
3. `echo "question" | qq` works with stdin-only input.
4. Empty args plus empty stdin exits non-zero with concise usage on stderr.
5. Backend priority is fixed for v1, in this exact order:
   `opencode`, `pi`, `codex`, `claude`, `agent`, `copilot`, `openclaw`, `gemini`, `qwen`,
   `q`, `kimi`, `kilo`, `kiro-cli`, `goose`, `aider`, `amp`, `droid`, `crush`, `cn`, `roo`.
6. Each backend is invoked using `Authoritative Backend Command Mapping`.
7. The implementation builds command argv arrays and never uses `eval` or shell-built command
   strings for prompt execution.
8. If no supported backend binary is found, `qq` exits non-zero and lists supported binaries on
   stderr.
9. If a found backend exits non-zero, `qq` exits non-zero without trying lower-priority installed
   backends.
10. Success output does not include wrapper banners, backend names, or progress text generated by
    the wrapper.
11. No Python files, scripts, or test commands remain under `ai_llm/qq`.

## Authoritative Backend Command Mapping

These argv mappings are the v1 source of truth. The prompt is always one literal argv value.

| Priority | Backend    | Argv template                                    |
|----------|------------|--------------------------------------------------|
| 1        | `opencode` | `opencode run <prompt>`                          |
| 2        | `pi`       | `pi -p <prompt>`                                 |
| 3        | `codex`    | `codex exec --ephemeral <prompt>`                |
| 4        | `claude`   | `claude --bare -p <prompt>`                      |
| 5        | `agent`    | `agent --mode ask -p <prompt>`                   |
| 6        | `copilot`  | `copilot -sp <prompt>`                           |
| 7        | `openclaw` | `openclaw agent --agent main --message <prompt>` |
| 8        | `gemini`   | `gemini -p <prompt>`                             |
| 9        | `qwen`     | `qwen -p <prompt>`                               |
| 10       | `q`        | `q chat --non-interactive <prompt>`              |
| 11       | `kimi`     | `kimi --quiet -p <prompt>`                       |
| 12       | `kilo`     | `kilo run <prompt>`                              |
| 13       | `kiro-cli` | `kiro-cli chat --no-interactive <prompt>`        |
| 14       | `goose`    | `goose run --no-session -t <prompt>`             |
| 15       | `aider`    | `aider --message <prompt>`                       |
| 16       | `amp`      | `amp -x <prompt>`                                |
| 17       | `droid`    | `droid exec <prompt>`                            |
| 18       | `crush`    | `crush run --quiet <prompt>`                     |
| 19       | `cn`       | `cn -p <prompt> --silent`                        |
| 20       | `roo`      | `roo --print <prompt>`                           |

## Test-First Strategy

Primary tests: Go stdlib black-box tests in `ai_llm/qq/qq_test.go`.

Test harness:

- Create a temp directory with fake backend executables.
- Put fake backend directory first in `PATH`.
- Run `ai_llm/qq/bin/qq` with `exec.Command`.
- Fake backends record argv and prompt payload to temp files, then print deterministic answers.
- Fake backends should be POSIX shell scripts so the test harness does not reintroduce Python.
- Avoid real AI CLI calls in tests.

Initial red tests:

1. Selects first installed backend from priority order.
2. Skips missing higher-priority binaries and uses the first present backend.
3. With no Tier 1 backend installed, `q chat --non-interactive` is selected before lower Tier 2
   backends such as `kimi`.
4. Args-only prompt is passed as one prompt argument to backend.
5. Args-only invocation with closed or empty stdin does not block waiting for stdin.
6. Stdin-only prompt is passed as one prompt argument.
7. Args plus stdin use the exact documented `Context:` separator.
8. A prompt containing shell metacharacters, such as `$(touch marker); echo hi`, is passed literally
   and does not execute shell code or create the marker file.
9. Backend non-zero stops without falling back to a lower-priority installed backend, including when
   installed `q` exits non-zero and `kimi` exists.
10. No backend produces non-zero exit and clear stderr.
11. Wrapper success output is exactly backend stdout.
12. Backend command flags match `Authoritative Backend Command Mapping` for every v1 backend.

Verification commands:

- `cd ai_llm/qq && go test ./...`
- `bash -n ai_llm/qq/bin/qq`
- Optional if available: `shellcheck ai_llm/qq/bin/qq`
- Optional host TTY smoke if available: `script -qec 'ai_llm/qq/bin/qq hello' /dev/null`

## Implementation Chunks

| ID | Goal                                                | Primary paths                                                            | Depends on | Parallel-safe | Verify with                              |
|----|-----------------------------------------------------|--------------------------------------------------------------------------|------------|---------------|------------------------------------------|
| T1 | Replace Python tests with Go stdlib black-box tests | `ai_llm/qq/go.mod`, `ai_llm/qq/qq_test.go`, `ai_llm/qq/tests/test_qq.py` | none       | No            | `cd ai_llm/qq && go test ./...`          |
| T2 | Keep/fix Bash wrapper until Go tests pass           | `ai_llm/qq/bin/qq`                                                       | T1         | No            | Go tests plus `bash -n ai_llm/qq/bin/qq` |
| T3 | Update docs and ignore rules for no-Python layout   | `ai_llm/qq/README.md`, `.gitignore`                                      | T2         | No            | docs review plus `git status --short`    |
| T4 | Final verification and cleanup notes                | `ai_llm/qq/*`, `.gitignore`                                              | T1-T3      | No            | all verification commands                |

## Parallelization Notes

- T1 and T2 should be serialized because tests define script behavior.
- T3 touches `.gitignore`, which controls whether new Go files are visible; do not run it in
  parallel with T1.
- T3 can run after T2, but should avoid restating the exact backend
  command table if the script is the source of truth.
- Do not split ownership of `ai_llm/qq/bin/qq` across concurrent implementers.
- Do not touch `docs/decisions/*`; historical docs are read-only for this task.

## Pitfalls And Edge Cases

- Stdin can block if read unconditionally. Read only when fd 0 is not a TTY.
- Missing command detection checks only the executable name, not subcommands such as `opencode run`.
- For Amazon Q Developer CLI, missing-command fallback checks only `q`. If `q` exists but
  `q chat --non-interactive` fails because of auth, config, or unsupported subcommand behavior,
  v1 stops on that failure.
- Prompt must be passed as argv, not interpolated into a shell string.
- Backend stdout should pass through unchanged; wrapper diagnostics go to stderr.
- Some CLIs may emit progress on stderr. Do not suppress stderr unless the backend flag explicitly
  does so; hiding errors makes failures harder to debug.
- `aider` and some agent CLIs are edit-oriented. Keep them lower priority and use only the supplied
  non-interactive prompt flags.
- Large stdin is read into memory in v1. Acceptable for quick Q&A; document as not intended for huge
  files.
- Newline preservation matters for piped context. Tests should assert exact separator and trailing
  newline behavior.
- The `Context:` separator is intentionally plain and stable so users can predict merged prompts.
- Falling back on backend non-zero would hide auth/config/quota failures and may produce confusing
  answers from a different model.
- Go stdlib has no PTY helper. Keep TTY behavior in Bash via `[[ ! -t 0 ]]`; cover no-blocking with
  bounded process tests and use optional `script` smoke testing when the host provides it.
