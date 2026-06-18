# qq

`qq` is a small terminal helper for quick AI-agent questions.

```sh
# directly
$ qq "how do I copy a directory to my home directory?"

cp -r /path/to/directory ~/directory

# via stdin piping
$ echo "hi there! how do I remove a directory from my home directory?" | qq

rm -rf ~/directory
```

See [`docs/spec.md`](docs/spec.md) for the full behavior and architecture spec.

## The "quick-and-dirty" alternative (and why this exists)

```sh
alias qq="claude --safe-mode -p"

# or
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

That will likely work for the majority of cases, but I wanted something more
robust and flexible: a wrapper I can tweak on the fly depending on the system
I'm on and adapt to each tool's idiosyncrasies. Hence why I wrote this.

## Installation

Install it with Go:

```bash
go install github.com/BoscoDomingo/utils/go/tools/qq@latest

# from source checkout
git clone https://github.com/BoscoDomingo/utils.git
cd utils/go/tools/qq
go install .
```

## Usage

Examples:

```bash
qq "how do I untar a file?"
echo "context" | qq "summarize this"
qq -b pi "review this error"
qq --provider claude --model claude-sonnet-4 "explain this command"
qq -b pi -m github-copilot/gpt-5.5 "reply with OK"
```

First-class backends are `claude`, `agent` (Cursor Agent), `opencode`, `pi`,
`codex`, and `gemini`. Backend priority and argv mappings live in
[`internal/backend/backend.go`](internal/backend/backend.go); treat that source
as authoritative when behavior and docs differ.

Use `-b`, `--backend`, `-p`, or `--provider` to select a backend. Use `--model`
or `-m` to pass an exact model selector through to the selected backend. When no
model is supplied, `qq` sends no model argument and lets the backend use its own
default.

If the backend flag is present without a backend value, `qq` opens an
interactive selector on `/dev/tty`, so piped stdin can still provide the prompt:

```bash
qq -b "choose interactively for this prompt"
echo "context" | qq -b claude
echo "context" | qq -b
```

Generate shell completion scripts with one of these commands:

```bash
qq completion bash
qq completion zsh
qq completion fish
qq completion powershell
```

Load completions dynamically from your shell startup file:

```bash
# bash: ~/.bashrc
[ -n "$(command -v qq)" ] && source <(qq completion bash)
```

```zsh
# zsh: ~/.zshrc, before plugins that wrap completion such as zsh-autocomplete
[ -n "$(command -v qq)" ] && source <(qq completion zsh)
```

```fish
# fish: ~/.config/fish/config.fish
type -q qq; and qq completion fish | source
```

```powershell
# PowerShell: $PROFILE
if (Get-Command qq -ErrorAction SilentlyContinue) {
    qq completion powershell | Out-String | Invoke-Expression
}
```

For static fish completions:

```fish
mkdir -p ~/.config/fish/completions
qq completion fish > ~/.config/fish/completions/qq.fish
```

If zsh still completes paths after provider flags, clear its completion cache
and restart:

```zsh
rm -f "${ZSH:-$HOME/.oh-my-zsh}/cache/.zcompdump-$HOST"
exec zsh
```

## Development

`qq` keeps only executable wiring in the module root. Behavior is split across
`internal/app`, `internal/backend`, `internal/cli`, `internal/invocation`, and
`internal/selector`. Backend priority and argv mappings live in
[`internal/backend/backend.go`](internal/backend/backend.go); treat that source
as authoritative when behavior and docs differ.

Where a backend exposes a supported non-interactive skill-suppression control,
`qq` applies it in the backend mapping. Backends without such a control still
run with the safest documented non-interactive/ask-style argv known to `qq`.

Automated tests are split by scope:

- root E2E tests use fake backend executables under `testdata`
- optional real-backend E2E smoke tests run installed backend CLIs when
  `QQ_REAL_BACKEND_E2E=1` is set; missing tools skip, installed tools must pass
- `internal/app` integration tests use injected dependencies
- package unit tests cover backend, CLI, invocation, and selector behavior

Run the full local check before changing behavior:

```bash
go test -race ./...
go vet ./...
go install .
```

Run real-backend smoke checks explicitly when you want to verify local backend
CLIs:

```bash
QQ_REAL_BACKEND_E2E=1 go test -run TestE2ERealBackendsReturnExpectedFormat
```

The interactive selector uses Charm v2 modules:
[`charm.land/bubbletea/v2`](https://pkg.go.dev/charm.land/bubbletea/v2) and
[`charm.land/bubbles/v2`](https://pkg.go.dev/charm.land/bubbles/v2).
[`docs/spec.md`](docs/spec.md) remains the canonical behavior spec.
