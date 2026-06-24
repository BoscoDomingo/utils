package main_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var qqE2EBinaryPath string

type e2eHarness struct {
	binDir     string
	captureDir string
}

type e2eResult struct {
	code   int
	stdout string
	stderr string
	err    error
}

func TestMain(m *testing.M) {
	buildDir, err := os.MkdirTemp("", "qq-e2e-build-")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create e2e build dir: %v\n", err)
		os.Exit(1)
	}

	qqE2EBinaryPath = filepath.Join(buildDir, "qq")
	build := exec.Command("go", "build", "-o", qqE2EBinaryPath, ".")
	output, err := build.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "build qq e2e binary: %v\n%s", err, output)
		_ = os.RemoveAll(buildDir)
		os.Exit(1)
	}

	code := m.Run()
	_ = os.RemoveAll(buildDir)
	os.Exit(code)
}

func newE2EHarness(t *testing.T, installedBackends []string) *e2eHarness {
	t.Helper()

	harness := &e2eHarness{
		binDir:     t.TempDir(),
		captureDir: t.TempDir(),
	}
	for _, backend := range installedBackends {
		harness.installFakeBackend(t, backend)
	}

	return harness
}

func (harness *e2eHarness) installFakeBackend(t *testing.T, name string) {
	t.Helper()

	fixture, err := os.ReadFile(filepath.Join("testdata", "fake-backend"))
	assertNoError(t, err)

	path := filepath.Join(harness.binDir, name)
	assertNoError(t, os.WriteFile(path, fixture, 0o755))
}

func (harness *e2eHarness) regularStdin(t *testing.T, content string) *os.File {
	t.Helper()

	path := filepath.Join(t.TempDir(), "stdin")
	assertNoError(t, os.WriteFile(path, []byte(content), 0o600))

	file, err := os.Open(path)
	assertNoError(t, err)
	t.Cleanup(func() {
		assertNoError(t, file.Close())
	})

	return file
}

func (harness *e2eHarness) run(
	t *testing.T,
	args []string,
	stdin io.Reader,
	extraEnv ...string,
) e2eResult {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, qqE2EBinaryPath, args...)
	cmd.Env = harness.env(extraEnv...)
	cmd.Stdin = stdin

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Fatalf("qq e2e command timed out: %v", args)
	}

	return e2eResult{
		code:   commandExitCode(t, err),
		stdout: stdout.String(),
		stderr: stderr.String(),
		err:    err,
	}
}

func (harness *e2eHarness) env(extraEnv ...string) []string {
	env := make([]string, 0, len(os.Environ())+len(extraEnv)+2)
	for _, entry := range os.Environ() {
		key := entry
		if separator := strings.IndexByte(entry, '='); separator >= 0 {
			key = entry[:separator]
		}
		if key == "PATH" || strings.HasPrefix(key, "QQ_FAKE_") {
			continue
		}
		env = append(env, entry)
	}

	env = append(env, "PATH="+harness.binDir)
	env = append(env, "QQ_FAKE_CAPTURE_DIR="+harness.captureDir)
	env = append(env, extraEnv...)
	return env
}

func (harness *e2eHarness) argv(t *testing.T, backend string) []string {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(harness.captureDir, backend+".argv0"))
	assertNoError(t, err)
	if len(data) == 0 {
		return nil
	}

	parts := bytes.Split(data, []byte{0})
	if len(parts[len(parts)-1]) == 0 {
		parts = parts[:len(parts)-1]
	}

	argv := make([]string, len(parts))
	for index, part := range parts {
		argv[index] = string(part)
	}
	return argv
}

func (harness *e2eHarness) capturedEnv(t *testing.T, backend string) map[string]string {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(harness.captureDir, backend+".env"))
	assertNoError(t, err)

	env := map[string]string{}
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			t.Fatalf("malformed captured env line: %q", line)
		}
		env[key] = value
	}
	return env
}

func (harness *e2eHarness) backendRan(backend string) bool {
	_, err := os.Stat(filepath.Join(harness.captureDir, backend+".argv0"))
	return err == nil
}

func commandExitCode(t *testing.T, err error) int {
	t.Helper()

	if err == nil {
		return 0
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}

	t.Fatalf("qq e2e command failed without exit status: %v", err)
	return 1
}

func expectedInlinePrompt(prompt string) string {
	return "System: Be concise; prefer a single-line answer or command. Output only the command when that answers the question.\n\nUser: " + prompt
}

func conciseSystemPrompt() string {
	return "Be concise; prefer a single-line answer or command. Output only the command when that answers the question."
}

func assertExitStatus(t *testing.T, result e2eResult, expected int) {
	t.Helper()

	if result.code != expected {
		t.Fatalf(
			"exit status mismatch: got %d want %d stdout=%q stderr=%q err=%v",
			result.code,
			expected,
			result.stdout,
			result.stderr,
			result.err,
		)
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertContains(t *testing.T, value string, substring string) {
	t.Helper()

	if !strings.Contains(value, substring) {
		t.Fatalf("expected %q to contain %q", value, substring)
	}
}

func assertEqual[T comparable](t *testing.T, got T, want T) {
	t.Helper()

	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func assertStringSlicesEqual(t *testing.T, got []string, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("slice length mismatch: got %#v, want %#v", got, want)
	}
	for index := range got {
		if got[index] != want[index] {
			t.Fatalf("slice mismatch: got %#v, want %#v", got, want)
		}
	}
}
