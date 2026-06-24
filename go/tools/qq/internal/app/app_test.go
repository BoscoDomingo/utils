package app

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestArgsOnlyPromptDoesNotBlockOnOpenNonTTYStdin(t *testing.T) {
	reader, writer, err := os.Pipe()
	assertNoError(t, err)
	defer func() { _ = reader.Close() }()
	defer func() { _ = writer.Close() }()

	app := newTestApp(t, []string{"opencode"})
	app.stdin = reader
	app.stdinIsTTY = false

	done := make(chan commandResult, 1)
	go func() {
		done <- app.execute(context.Background(), "hello")
	}()

	select {
	case result := <-done:
		assertExitCode(t, result, 0)
		assertEqual(t, app.runner.calls[0].spec.Name, "opencode")
		assertStringSlicesEqual(t, app.runner.calls[0].spec.Args, []string{"run", expectedInlinePrompt("hello")})
	case <-time.After(250 * time.Millisecond):
		_ = reader.Close()
		_ = writer.Close()
		t.Fatal("args-only prompt blocked reading open non-TTY stdin")
	}
}

func TestArgsPromptSkipsPartialPipeStdinUntilWriterCloses(t *testing.T) {
	reader, writer, err := os.Pipe()
	assertNoError(t, err)
	defer func() { _ = reader.Close() }()
	defer func() { _ = writer.Close() }()

	_, err = writer.WriteString("context\n")
	assertNoError(t, err)

	app := newTestApp(t, []string{"opencode"})
	app.stdin = reader
	app.stdinIsTTY = false

	done := make(chan commandResult, 1)
	go func() {
		done <- app.execute(context.Background(), "summarize")
	}()

	select {
	case result := <-done:
		assertExitCode(t, result, 0)
		assertEqual(t, app.runner.calls[0].spec.Name, "opencode")
		assertStringSlicesEqual(t, app.runner.calls[0].spec.Args, []string{"run", expectedInlinePrompt("summarize")})
	case <-time.After(250 * time.Millisecond):
		_ = reader.Close()
		_ = writer.Close()
		select {
		case result := <-done:
			t.Fatalf("partial open pipe returned only after forced close: %#v", result)
		case <-time.After(time.Second):
			t.Fatal("args prompt blocked reading partial pipe stdin")
		}
	}
}

func TestArgsPromptMergesImmediatelyAvailablePipeStdin(t *testing.T) {
	reader, writer, err := os.Pipe()
	assertNoError(t, err)
	_, err = writer.WriteString("context\n")
	assertNoError(t, err)
	assertNoError(t, writer.Close())
	defer func() { _ = reader.Close() }()

	app := newTestApp(t, []string{"opencode"})
	app.stdin = reader
	app.stdinIsTTY = false

	result := app.execute(context.Background(), "summarize")

	assertExitCode(t, result, 0)
	assertStringSlicesEqual(
		t,
		app.runner.calls[0].spec.Args,
		[]string{"run", expectedInlinePrompt("summarize\n\nContext:\n\ncontext\n")},
	)
}

func TestSelectorRequiredWithoutTTYFailsBeforeBackendRun(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, []string{"opencode"})
	app.stdin = strings.NewReader("hello\n")
	app.stdinIsTTY = false
	app.ttyAvailable = false

	result := app.execute(context.Background(), "-b")

	assertExitCode(t, result, ExitUsage)
	assertContains(t, result.stderr, "interactive backend selection requires a TTY")
	assertEqual(t, len(app.runner.calls), 0)
}

func TestSelectorUsesInjectedBackendWithPipedPrompt(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, []string{"claude", "gemini"})
	app.stdin = strings.NewReader("hello\n")
	app.stdinIsTTY = false
	app.ttyAvailable = true
	app.selector.selected = "gemini"

	result := app.execute(context.Background(), "-b")

	assertExitCode(t, result, 0)
	assertEqual(t, result.stdout, "answer from gemini\n")
	assertEqual(t, app.runner.calls[0].spec.Name, "gemini")
	assertStringSlicesEqual(t, app.runner.calls[0].spec.Args, []string{"-p", expectedInlinePrompt("hello\n")})
}

func TestSelectorPromptDoesNotBlockOnOpenNonTTYStdin(t *testing.T) {
	reader, writer, err := os.Pipe()
	assertNoError(t, err)
	defer func() { _ = reader.Close() }()
	defer func() { _ = writer.Close() }()

	app := newTestApp(t, []string{"gemini"})
	app.stdin = reader
	app.stdinIsTTY = false
	app.selector.selected = "gemini"

	done := make(chan commandResult, 1)
	go func() {
		done <- app.execute(context.Background(), "-b", "explain cobra")
	}()

	select {
	case result := <-done:
		assertExitCode(t, result, 0)
		assertEqual(t, app.runner.calls[0].spec.Name, "gemini")
		assertStringSlicesEqual(t, app.runner.calls[0].spec.Args, []string{"-p", expectedInlinePrompt("explain cobra")})
	case <-time.After(250 * time.Millisecond):
		_ = reader.Close()
		_ = writer.Close()
		t.Fatal("selector prompt blocked reading open non-TTY stdin")
	}
}

func TestDefaultBackendPriorityUsesFirstInstalledBackend(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, []string{"opencode", "gemini", "pi"})

	result := app.execute(context.Background(), "hello")

	assertExitCode(t, result, 0)
	assertEqual(t, result.stdout, "answer from pi\n")
	assertEqual(t, app.runner.calls[0].spec.Name, "pi")
}

func TestExplicitBackendDoesNotFallbackWhenMissing(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, []string{"pi"})

	result := app.execute(context.Background(), "-b", "gemini", "hello")

	assertExitCode(t, result, ExitBackendUnavailable)
	assertContains(t, result.stderr, "Selected backend not found: gemini")
	assertEqual(t, len(app.runner.calls), 0)
}

func TestModelFlagPassesSelectorToBackend(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, []string{"pi"})

	result := app.execute(
		context.Background(),
		"-b", "pi", "-m", "github-copilot/gpt-5.4-mini", "hello",
	)

	assertExitCode(t, result, 0)
	assertEqual(t, app.runner.calls[0].spec.Name, "pi")
	assertStringSlicesEqual(
		t,
		app.runner.calls[0].spec.Args,
		[]string{
			"--model", "github-copilot/gpt-5.4-mini",
			"--append-system-prompt", "Be concise; prefer a single-line answer or command. Output only the command when that answers the question.",
			"-p", "hello",
		},
	)
}

func TestNoBackendReturnsClearError(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, nil)

	result := app.execute(context.Background(), "hello")

	assertExitCode(t, result, ExitBackendUnavailable)
	assertContains(t, result.stderr, "No supported backend found")
	assertContains(t, result.stderr, "opencode")
	assertContains(t, result.stderr, "gemini")
	assertEqual(t, len(app.runner.calls), 0)
}

func TestInstalledBackendNonZeroStopsWithoutFallback(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, []string{"opencode", "pi"})
	app.runner.exitCodes["pi"] = 42

	result := app.execute(context.Background(), "hello")

	assertExitCode(t, result, 42)
	assertContains(t, result.stderr, "pi failed")
	assertBackendCalls(t, app.runner.calls, []string{"pi"})
}

func TestSuccessOutputIsExactlyBackendStdout(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, []string{"opencode"})
	app.runner.stdout["opencode"] = "backend stdout only\n"

	result := app.execute(context.Background(), "hello")

	assertExitCode(t, result, 0)
	assertEqual(t, result.stdout, "backend stdout only\n")
	assertEqual(t, result.stderr, "")
}
