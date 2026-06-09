# qq

Read `docs/spec.md` before behavior changes.

`internal/backend/backend.go` is authoritative for backend priority and argv.

Run checks from `go/tools/qq`; see `README.md` for commands.

Keep default tests deterministic. Real backend E2E is opt-in via
`QQ_REAL_BACKEND_E2E=1`; missing tools skip, installed tools fail on bad output
or non-zero exit.
