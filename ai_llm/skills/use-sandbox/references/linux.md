# Offline Linux Execution

## Prerequisites

Read-only checks; run help only when the tool exists:

```sh
uname -s
command -v bwrap
bwrap --help
command -v timeout
timeout --help
```

Require usable unprivileged user namespaces and all skeleton options. Never substitute best-effort `--*-try`, `--share-net`, `sudo`, or another backend after failure. Use GNU `timeout`, or an existing supervisor with known deadline, forced termination, and descendant cleanup behavior.

- Use an existing trusted minimal runtime with the approved executable, loader, libraries, and trusted probe tools. Never mount host `/`, home, the whole repository, service sockets, or host device trees. Do not execute untrusted code or use `ldd` on it to discover dependencies.
- Runtime mount points `/inputs`, `/work`, `/tmp`, `/home/sandbox`, `/proc`, and `/dev` must exist; add `/out` only if approved. Stage only needed inputs, excluding secrets, `.env`, VCS metadata, sockets, devices, and nested mounts. Read-only mounts still disclose data.
- Review absolute source paths and symlinks; do not dereference links into unrelated host files. Prevent untrusted concurrent writers. Keep runtime and staging unchanged after preflight, or repeat it.
- Require approved invocation-owned staging, a positive deadline, bounded log capture, and a launcher forwarding only intended stdio, not unrelated descriptors or sockets.

## Conditional Command Skeleton

Illustrative Bash, **not a universally safe recipe or tested launcher**. Supply reviewed absolute paths, a prepared runtime, positive `TIME_LIMIT` (e.g. `60s`), and guest executable/arguments in the `COMMAND` array. No `eval` or untrusted host shell interpolation.

```bash
SANDBOX=(
  bwrap
  --unshare-user --unshare-ipc --unshare-pid --unshare-net --unshare-uts
  --disable-userns --cap-drop ALL --new-session --die-with-parent
  --clearenv
  --ro-bind "$RUNTIME_ROOT" /
  --ro-bind "$INPUT_DIR" /inputs
  --tmpfs /work --tmpfs /tmp --tmpfs /home/sandbox
  --proc /proc --dev /dev
  --setenv HOME /home/sandbox --setenv PATH /usr/bin:/bin --setenv LANG C
  --chdir /work
)
# Set COMMAND to the trusted preflight first, then the approved workload.
timeout --signal=TERM --kill-after=5s "$TIME_LIMIT" \
  "${SANDBOX[@]}" -- "${COMMAND[@]}" </dev/null
```

Scratch/home are ephemeral. For approved persistent output, add `--bind "$OUTPUT_DIR" /out` to the array: a separate, empty, invocation-owned directory without links or nested mounts. This grants host writes and can fill storage. Preflight this exact variant.

## Trusted Per-Invocation Preflight

Only with fixture/launch approval, use a trusted probe from the runtime and the same array, environment, descriptors, and supervisor as the workload. It must report each assertion and fail nonzero on failure:

- Read a synthetic staged file with known bytes; its append attempt must fail. A host-readable synthetic canary outside every source mount must be unreadable inside. Compare both host checksums before/after.
- Set a synthetic host token without `--setenv`; confirm it is absent inside. Check the reviewed guest environment allowlist, allowing expected bookkeeping such as `PWD`, but no credential/service variables.
- Compare host and guest `/proc/self/ns/{mnt,user,pid,net}` identifiers; each must differ. Guest `/proc/net/dev` may show only loopback, and `/proc/net/route` no non-loopback route.
- Write/reread synthetic files in each scratch/home path and, if enabled, `/out`; confirm the output reaches only its approved host directory. Inspect mounts and descriptors for unintended access or forwarded sockets.

Missing tools, incomplete results, namespace denial, successful input writes, or changed checksums block execution. This checks configuration, not all escapes or end-to-end connectivity denial.

## Run and Handle Results

- Use a trusted byte-bounded stdout/stderr collector that renders control sequences inert and drains/discards excess or cancels with descendant cleanup. Timeout alone does not bound logs.
- Record exit status separately from truncated output. GNU timeout normally returns 124 on deadline; forced termination can return 137. Report timeout/cancellation as incomplete, and confirm invocation-owned processes exited before cleanup.
- Inspect exports as data, without following symlinks, executing files, applying patches, or loading configuration automatically. Delete only invocation-owned resources.

## Limits and Sources

The skeleton has not been executed or integration-tested; preflight is not an integration-test pass. Supervisor cancellation/descendant cleanup and isolation need runner validation before claiming them tested. This baseline has no syscall allowlist or hard CPU/memory/process/output-disk quotas; stop if required limits are unavailable. Networking requires a separately reviewed policy.

- [Upstream security model and limitations](https://github.com/containers/bubblewrap/blob/main/README.md)
- [Upstream option reference](https://github.com/containers/bubblewrap/blob/main/bwrap.xml)
- [GNU timeout behavior](https://www.gnu.org/software/coreutils/manual/html_node/timeout-invocation.html)
