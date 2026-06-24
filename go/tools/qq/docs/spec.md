# qq specification

`qq` is a terminal helper for quick AI-agent questions. It routes a prompt to a
supported agent CLI, streams the answer on stdout, and avoids TUIs or disruptive
interactive flows in the default path.

## Context

I wanted something simple but a bit more sophisticated that could deal with each tools idiosyncrasies than the quick-and-dirty alternative:

```sh
qq() {
  local -a backend
  backend=(claude --safe-mode -p)

  if [[ "$1" == "-b" ]]; then
    shift
    [[ -n "$1" ]] || { print -u2 "usage: qq [-b 'cmd args'] prompt"; return 2; }
    backend=(${(z)1})
    shift
  fi

  [[ $# -gt 0 ]] || { print -u2 "usage: qq [-b 'cmd args'] prompt"; return 2; }

  "${backend[@]}" "$*"
}
```

## Goals

Primary UX:

```bash
qq "how can I quick untar a file?"
echo "context" | qq "summarize this"
```

Design intent:

- Quick Q&A, not editing. Prefer safe/read-only/ask modes where a backend supports them.
- Fixed first-class backend set and priority list for v1; no user config file.
- Piping must work. The wrapper reads stdin and merges it into the final prompt.
- Success output is backend stdout only, streamed as the backend writes it. Wrapper diagnostics go to stderr.
- Answers should be concise by default; prefer a single-line command or answer.
- Name is `qq`, not `q`, to avoid colliding with Amazon Q Developer CLI's `q` binary.

## Architecture

Runtime is a Go CLI under `go/tools/qq/cmd/qq` (`go install ./cmd/qq` from the tool module root).

| Layer                 | Responsibility                                                          |
|-----------------------|-------------------------------------------------------------------------|
| `cmd/qq/main.go`      | Executable wiring for `go install ./cmd/qq`                             |
| `internal/cli`        | Cobra root command, flags, and shell completion generation              |
| `internal/app`        | Orchestration: stdin handling, invocation parsing, and backend dispatch |
| `internal/invocation` | Pure provider flag parsing and prompt construction                      |
| `internal/backend`    | Backend registry, argv rendering, runner abstraction, and runner errors |
| `internal/selector`   | Interactive backend picker using Charm v2 on `/dev/tty`                 |
| `testdata`            | Fake backend fixtures for real-binary E2E tests                         |

Dependencies:

- [Cobra](https://github.com/spf13/cobra) — CLI, flag completion, `completion` subcommand
- [Bubble Tea v2](https://pkg.go.dev/charm.land/bubbletea/v2) + [Bubbles v2](https://pkg.go.dev/charm.land/bubbles/v2) — interactive backend selector
- `golang.org/x/sys/unix` — non-blocking stdin readiness checks

Implementation uses `exec.Command` with argv slices. Prompts are never passed through shell string interpolation or `eval`.
Backend-specific environment overrides are set only when listed in
[`internal/backend/backend.go`](../internal/backend/backend.go).

## Product decisions

| Topic                       | Decision                                                                                 |
|-----------------------------|------------------------------------------------------------------------------------------|
| Location                    | Self-contained under `go/tools/qq/`                                                                 |
| Distribution                | `go install github.com/BoscoDomingo/utils/go/tools/qq/cmd/qq@latest` or `go install ./cmd/qq` locally |
| Default backend selection   | First installed backend from fixed priority order; Pi is first                           |
| Explicit backend            | `-b`, `--backend`, `-p`, `--provider` flags                                              |
| Explicit model selector     | `--model`, `-m` flags pass the exact selector string to the selected backend              |
| Selector fallback           | When a provider flag is present but no backend value resolves, open interactive selector |
| Success output              | Stream backend stdout directly; no wrapper banners or backend names on stdout            |
| Fallback on missing binary  | Try next backend only when `LookPath` fails for the executable name                      |
| Fallback on backend failure | None. Installed backend non-zero exits stop immediately                                  |
| Prompt merge separator      | Args, blank line, `Context:`, blank line, stdin                                          |
| Answer style                | Backend prompts include a concise-answer instruction; exact argv mapping lives in source |
| First-class backends        | Claude, Cursor Agent, OpenCode, Pi, Codex, and Gemini                                   |
| Python                      | Not used by this tool                                                                    |

## CLI behavior

### Default invocation

`qq "question"` selects the first installed backend from the priority list and
runs it with the prompt.

### Provider flags

Supported flag names (equivalent):

- `-b`, `--backend`
- `-p`, `--provider`

Examples:

```bash
qq -b pi "hey, what's up"
qq -p claude "hey"
qq --backend agent "hey"
qq --provider opencode "hey"
```

### Model selection

`qq --model <selector>` and `qq -m <selector>` pass the exact selector string to
the selected backend using that backend's native model flag:

```bash
qq -b pi -m github-copilot/gpt-5.5 "reply with OK"
qq --provider claude --model claude-sonnet-4 "explain this command"
```

If no model flag is supplied, `qq` passes no model argument. Backend-native
defaults remain responsible for model choice.

### Selector fallback

When a provider flag is present but no usable backend value can be resolved,
`qq` opens an interactive selector:

```bash
qq -b "xyz"          # selector + prompt "xyz"
qq "xyz" -b          # selector + prompt "xyz"
qq -b                # selector; prompt from stdin if available
```

Resolution rules for the token after a provider flag:

| Situation                                                                                                                      | Interpretation                          |
|--------------------------------------------------------------------------------------------------------------------------------|-----------------------------------------|
| Token is a supported backend name                                                                                              | Select that backend                     |
| Token is not a supported backend name, and prompt context exists elsewhere (remaining args, prior prompt args, or piped stdin) | Error: unsupported backend              |
| Token is not a supported backend name, and no other prompt context exists                                                      | Treat token as prompt; request selector |
| Flag present with no value (`-b`, `--backend`, etc.)                                                                           | Request selector                        |

Prompt context examples:

```bash
echo "hello" | qq -b claude    # backend=claude, prompt from stdin
echo "hello" | qq -b           # selector required; stdin is prompt if selector succeeds
qq -b pi "review this"         # backend=pi, prompt="review this"
```

### Interactive selector

When selector is required:

- If `/dev/tty` is available, run Bubble Tea on `/dev/tty` so piped stdin can still supply the prompt.
- If `/dev/tty` is unavailable, fail with a clear error (`interactive backend selection requires a TTY`).
- Selector lists only installed backends (those found via `LookPath`).
- Cancellation exits non-zero.

Shell completion may suggest backend names after provider flags, but runtime only treats a flag value as a backend name when prompt context exists or the value is a known backend per the rules above.

### Prompt construction

| Input        | Prompt                            |
|--------------|-----------------------------------|
| Args only    | Joined CLI args (space-separated) |
| Stdin only   | Stdin contents                    |
| Args + stdin | `{args}\n\nContext:\n\n{stdin}`   |

Empty args and empty stdin exit non-zero with concise usage on stderr.

`qq` also adds a concise-answer instruction in `internal/backend/backend.go`.
Backends with a native append-system-prompt flag receive it there; others receive
an inline `System:` prelude inside the single prompt argument.

### Stdin handling

Stdin behavior is safety-critical for scripting and automation.

| Mode     | When                                                 | Behavior                                            |
|----------|------------------------------------------------------|-----------------------------------------------------|
| None     | Stdin is a TTY                                       | Do not read stdin                                   |
| Optional | Non-TTY stdin and prompt already available from args | Merge stdin only when safe to read without blocking |
| Required | No prompt from args and stdin is not a TTY           | Read stdin (may block until EOF)                    |

Optional stdin is read only when input is immediately drainable:

- In-memory readers (`bytes.Buffer`, `strings.Reader`, etc.)
- Regular files
- Pipes where EOF is already observable (`POLLHUP` or `POLLERR`)

Optional stdin must **not** block on:

- Open pipes with no data
- Open pipes with queued data but writer still open (no EOF yet)

Args-only invocations must never hang waiting for stdin.

## Backend priority and command mapping

Authoritative source: [`internal/backend/backend.go`](../internal/backend/backend.go)
(`supportedBackends`).

Backend priority, argv templates, and backend-specific environment overrides are
intentionally not duplicated here.
Keep behavior changes in source and tests, then update this spec only when the
contract changes.

Missing-command detection checks only the executable name on `PATH`, not subcommands.
Where a backend provides a supported skill-discovery suppression mechanism, `qq`
uses it in the backend command mapping. Cursor Agent currently has no documented
equivalent in this contract, so `qq` does not claim to disable Cursor skills.

## Acceptance criteria

1. `qq "question"` prints only the selected backend's answer to stdout on success.
2. `echo "context" | qq "question"` merges stdin into the prompt sent to the backend.
3. `echo "question" | qq` works with stdin-only input.
4. Empty args plus empty stdin exits non-zero with concise usage on stderr.
5. Backend priority matches `internal/backend/backend.go`.
6. Each backend is invoked per its source command template with the prompt as one
   literal argument.
7. No `eval` or shell-built command strings for prompt execution.
8. If no supported backend binary is found, `qq` exits non-zero and lists supported backends on stderr.
9. If a found backend exits non-zero, `qq` exits non-zero without trying lower-priority installed backends.
10. Success output streams backend stdout/stderr as it is written and does not include wrapper banners, backend names, or progress text from the wrapper.
11. Backend invocations include the concise-answer instruction from `internal/backend/backend.go`.
12. No Python files are used by this tool.
13. Provider flags select backends, open the selector when required, and reject unsupported explicit backends when prompt context exists.
14. `--model` and `-m` pass the exact selector string to the chosen backend; no model flag means no model argv.
15. Args-only invocations do not block on open or partially-filled stdin pipes.
16. Shell completion generates scripts for `bash`, `zsh`, `fish`, and `powershell`; provider flags complete to supported backend names and the model flag avoids file completion.

## Testing

Primary tests are Go tests split by scope:

- root E2E tests in [`e2e_test.go`](../e2e_test.go) run the real `qq` binary
  against fake backend executables under `testdata`
- optional real-backend smoke E2E tests run installed backend CLIs only when
  `QQ_REAL_BACKEND_E2E=1` is set
- root helper files support the E2E harness
- `internal/app` integration tests exercise orchestration through injected dependencies
- focused package unit tests cover `internal/backend`, `internal/cli`,
  `internal/invocation`, and `internal/selector`

Harness:

- Temp directory with fake backend executables prepended to `PATH`
- Fake backends log argv and return deterministic stdout/stderr
- No real AI CLI calls in default automated tests
- Real-backend smoke tests run from an empty temp directory, skip missing backend
  binaries, and fail installed backends that exit non-zero or do not return the
  expected sentinel output
- Selector behavior tested via injectable `BackendSelector`; Bubble Tea wiring smoke-tested only

Coverage includes:

- Full priority order across the first-class backend set
- Model override argv rendering, including no model argv when no override is supplied
- Concise-answer instruction rendering, including native system-prompt flags and inline fallback
- Args-only, stdin-only, and args+stdin prompt merge with exact `Context:` separator
- Shell metacharacters passed literally (no shell execution side effects)
- Missing-command fallback only; non-zero backend stops without fallback
- Stdout streaming before backend process exit
- Provider flag parser semantics and unsupported-backend errors
- Optional stdin non-blocking behavior (open pipe, partial pipe with writer held open)
- Completion registration for provider/backend flags

Verification:

```bash
cd go/tools/qq
go test -race ./...
go vet ./...
go install ./cmd/qq
```

Optional manual checks:

- `qq -b` in a real terminal (interactive selector UX)
- `qq completion zsh` redirected/sourced for your shell

## Pitfalls and edge cases

- **Stdin blocking**: Never read stdin unconditionally. Distinguish TTY, required, and optional modes.
- **Partial pipes**: `POLLIN` alone is not enough for optional reads; wait for EOF signals before `ReadAll`.
- **Model overrides**: `--model` and `-m` pass exact selector strings; no-override paths must not hard-code defaults.
- **Prompt safety**: Pass prompt as argv; never interpolate into shell strings.
- **stderr passthrough**: Do not suppress backend stderr unless the backend flag does.
- **Output streaming**: Do not buffer backend stdout/stderr in the wrapper path.
- **Long-tail CLIs**: Backends outside the first-class set are not supported by v1.
- **Large stdin**: Read into memory in v1; not intended for huge files.
- **Selector vs stdin**: Bubble Tea must use `/dev/tty`, not fd 0, so piped prompts remain available.
- **Completion vs runtime**: Completion may suggest backends even when runtime will treat the next token as a prompt; tests lock the resolution rules.
- **Non-zero fallback**: Falling back after backend failure would hide auth, quota, and config errors.

## Shell completion

Generate completion scripts:

```bash
qq completion bash
qq completion zsh
qq completion fish
qq completion powershell
```

Source or redirect the output for your shell. Provider/backend flags complete to
the supported backend list with `ShellCompDirectiveNoFileComp`; the model flag is
registered but does not enumerate model values.

## Related docs

- Usage and install quick start: [`README.md`](../README.md)
- Historical implementation plan: [`docs/plans/2026-06-17-qq-agent-cli-wrapper.md`](plans/2026-06-17-qq-agent-cli-wrapper.md)
- Repo conventions: [`AGENTS.md`](../../../../AGENTS.md)
