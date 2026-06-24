package main_test

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	appcore "github.com/BoscoDomingo/utils/go/tools/qq/internal/app"
)

func TestE2EArgsOnlyPrompt(t *testing.T) {
	t.Parallel()

	harness := newE2EHarness(t, []string{"opencode"})

	result := harness.run(t, []string{"how", "now"}, nil)

	assertExitStatus(t, result, 0)
	assertEqual(t, result.stdout, "fake stdout from opencode\n")
	assertEqual(t, result.stderr, "")
	assertStringSlicesEqual(t, harness.argv(t, "opencode"), []string{"run", expectedInlinePrompt("how now")})
	assertEqual(t, harness.capturedEnv(t, "opencode")["OPENCODE_DISABLE_EXTERNAL_SKILLS"], "1")
}

func TestE2EStdinOnlyPrompt(t *testing.T) {
	t.Parallel()

	harness := newE2EHarness(t, []string{"opencode"})

	result := harness.run(t, nil, strings.NewReader("from stdin\n"))

	assertExitStatus(t, result, 0)
	assertStringSlicesEqual(t, harness.argv(t, "opencode"), []string{"run", expectedInlinePrompt("from stdin\n")})
}

func TestE2EArgsAndStdinMergePrompt(t *testing.T) {
	t.Parallel()

	harness := newE2EHarness(t, []string{"opencode"})
	stdin := harness.regularStdin(t, "context\n")

	result := harness.run(t, []string{"summarize"}, stdin)

	assertExitStatus(t, result, 0)
	assertStringSlicesEqual(
		t,
		harness.argv(t, "opencode"),
		[]string{"run", expectedInlinePrompt("summarize\n\nContext:\n\ncontext\n")},
	)
	assertEqual(t, harness.capturedEnv(t, "opencode")["OPENCODE_DISABLE_EXTERNAL_SKILLS"], "1")
}

func TestE2EClaudeUsesSafeMode(t *testing.T) {
	t.Parallel()

	harness := newE2EHarness(t, []string{"claude"})
	prompt := "literal $(echo bad)"

	result := harness.run(t, []string{"-b", "claude", prompt}, nil)

	assertExitStatus(t, result, 0)
	assertStringSlicesEqual(
		t,
		harness.argv(t, "claude"),
		[]string{"--safe-mode", "--append-system-prompt", conciseSystemPrompt(), "-p", prompt},
	)
}

func TestE2EDefaultClaudeUsesSafeMode(t *testing.T) {
	t.Parallel()

	harness := newE2EHarness(t, []string{"claude"})

	result := harness.run(t, []string{"hello"}, nil)

	assertExitStatus(t, result, 0)
	assertStringSlicesEqual(
		t,
		harness.argv(t, "claude"),
		[]string{"--safe-mode", "--append-system-prompt", conciseSystemPrompt(), "-p", "hello"},
	)
}

func TestE2EExplicitBackend(t *testing.T) {
	t.Parallel()

	harness := newE2EHarness(t, []string{"opencode", "gemini"})

	result := harness.run(t, []string{"-b", "gemini", "hello"}, nil)

	assertExitStatus(t, result, 0)
	assertEqual(t, result.stdout, "fake stdout from gemini\n")
	assertStringSlicesEqual(t, harness.argv(t, "gemini"), []string{"-p", expectedInlinePrompt("hello")})
	assertEqual(t, harness.backendRan("opencode"), false)
}

func TestE2EPiIsDefaultWhenInstalled(t *testing.T) {
	t.Parallel()

	harness := newE2EHarness(t, []string{"opencode", "pi", "gemini"})

	result := harness.run(t, []string{"hello"}, nil)

	assertExitStatus(t, result, 0)
	assertEqual(t, result.stdout, "fake stdout from pi\n")
	assertStringSlicesEqual(
		t,
		harness.argv(t, "pi"),
		[]string{"--append-system-prompt", conciseSystemPrompt(), "-p", "hello"},
	)
	assertEqual(t, harness.backendRan("opencode"), false)
	assertEqual(t, harness.backendRan("gemini"), false)
}

func TestE2EExplicitModelOverrideReachesBackend(t *testing.T) {
	t.Parallel()

	harness := newE2EHarness(t, []string{"pi"})
	model := "github-copilot/gpt-5.5"

	result := harness.run(t, []string{"-b", "pi", "-m", model, "hi"}, nil)

	assertExitStatus(t, result, 0)
	assertStringSlicesEqual(
		t,
		harness.argv(t, "pi"),
		[]string{"--model", model, "--append-system-prompt", conciseSystemPrompt(), "-p", "hi"},
	)
}

func TestE2EBackendFailurePassthrough(t *testing.T) {
	t.Parallel()

	harness := newE2EHarness(t, []string{"opencode", "pi"})

	result := harness.run(
		t,
		[]string{"hello"},
		nil,
		"QQ_FAKE_FAIL_BACKEND=pi",
		"QQ_FAKE_EXIT_CODE=42",
		"QQ_FAKE_STDOUT=partial stdout\n",
		"QQ_FAKE_STDERR=backend broke\n",
	)

	assertExitStatus(t, result, 42)
	assertEqual(t, result.stdout, "partial stdout\n")
	assertEqual(t, result.stderr, "backend broke\n")
	assertStringSlicesEqual(
		t,
		harness.argv(t, "pi"),
		[]string{"--append-system-prompt", conciseSystemPrompt(), "-p", "hello"},
	)
	assertEqual(t, harness.backendRan("opencode"), false)
}

func TestE2ENoBackendError(t *testing.T) {
	t.Parallel()

	harness := newE2EHarness(t, nil)

	result := harness.run(t, []string{"hello"}, nil)

	assertExitStatus(t, result, appcore.ExitBackendUnavailable)
	assertEqual(t, result.stdout, "")
	assertContains(t, result.stderr, "No supported backend found")
	assertContains(t, result.stderr, "opencode")
	assertContains(t, result.stderr, "gemini")
}

func TestE2EStdoutStderrPassthrough(t *testing.T) {
	t.Parallel()

	harness := newE2EHarness(t, []string{"opencode"})

	result := harness.run(
		t,
		[]string{"hello"},
		nil,
		"QQ_FAKE_STDOUT=backend stdout\n",
		"QQ_FAKE_STDERR=backend stderr\n",
	)

	assertExitStatus(t, result, 0)
	assertEqual(t, result.stdout, "backend stdout\n")
	assertEqual(t, result.stderr, "backend stderr\n")
}

func TestE2EStreamsBackendStdoutBeforeExit(t *testing.T) {
	t.Parallel()

	harness := newE2EHarness(t, []string{"opencode"})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, qqE2EBinaryPath, "hello")
	cmd.Env = harness.env("QQ_FAKE_STREAM_STDOUT=1", "QQ_FAKE_STREAM_DELAY=1")
	stdout, err := cmd.StdoutPipe()
	assertNoError(t, err)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	assertNoError(t, cmd.Start())
	reader := bufio.NewReader(stdout)
	lineCh := make(chan string, 1)
	errCh := make(chan error, 1)
	go func() {
		line, err := reader.ReadString('\n')
		if err != nil {
			errCh <- err
			return
		}
		lineCh <- line
	}()

	select {
	case line := <-lineCh:
		assertEqual(t, line, "stream first\n")
	case err := <-errCh:
		t.Fatalf("read streamed stdout: %v stderr=%q", err, stderr.String())
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for first streamed stdout chunk")
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		t.Fatalf("qq exited before delayed backend output: err=%v stderr=%q", err, stderr.String())
	case <-time.After(300 * time.Millisecond):
	}

	select {
	case err := <-done:
		assertNoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatalf("qq did not finish after streaming output: stderr=%q", stderr.String())
	}
}

func TestE2ELiteralShellMetacharactersPassAsArgv(t *testing.T) {
	t.Parallel()

	harness := newE2EHarness(t, []string{"opencode"})
	markerPath := filepath.Join(t.TempDir(), "marker")
	prompt := fmt.Sprintf("literal $(touch %s); echo nope", markerPath)

	result := harness.run(t, []string{prompt}, nil)

	assertExitStatus(t, result, 0)
	assertStringSlicesEqual(t, harness.argv(t, "opencode"), []string{"run", expectedInlinePrompt(prompt)})
	if _, err := os.Stat(markerPath); !os.IsNotExist(err) {
		t.Fatalf("shell metacharacters were executed; marker stat err: %v", err)
	}
}
