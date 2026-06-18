package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/BoscoDomingo/utils/go/tools/qq/internal/backend"
)

const realBackendE2EEnv = "QQ_REAL_BACKEND_E2E"

const realBackendPrompt = `Reply with exactly this single line and nothing else:
QQ_REAL_BACKEND_E2E_OK

Do not read files. Do not edit files. Do not run commands.`

func TestE2ERealBackendsReturnExpectedFormat(t *testing.T) {
	if os.Getenv(realBackendE2EEnv) != "1" {
		t.Skipf("set %s=1 to run real backend E2E smoke tests", realBackendE2EEnv)
	}

	for _, item := range backend.Supported() {
		name := fmt.Sprint(item.Name())

		t.Run(name, func(t *testing.T) {
			executablePath, err := exec.LookPath(name)
			if err != nil {
				t.Skipf("backend executable %q not found on PATH", name)
			}

			result := runRealBackend(t, name)

			assertExitStatus(t, result, 0)
			assertRealBackendFormat(t, name, executablePath, result)
		})
	}
}

func runRealBackend(t *testing.T, name string) e2eResult {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, qqE2EBinaryPath, "-b", name, realBackendPrompt)
	cmd.Dir = t.TempDir()
	cmd.Env = append(os.Environ(), "NO_COLOR=1")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Fatalf("real backend %q timed out", name)
	}

	return e2eResult{
		code:   commandExitCode(t, err),
		stdout: stdout.String(),
		stderr: stderr.String(),
		err:    err,
	}
}

func assertRealBackendFormat(
	t *testing.T,
	name string,
	executablePath string,
	result e2eResult,
) {
	t.Helper()

	stdout := strings.TrimSpace(result.stdout)
	if stdout != "QQ_REAL_BACKEND_E2E_OK" {
		t.Fatalf(
			"real backend %q at %s returned unexpected stdout format: stdout=%q stderr=%q",
			name,
			executablePath,
			result.stdout,
			result.stderr,
		)
	}
}
