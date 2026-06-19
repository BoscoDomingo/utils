package backend

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSupportedBackendsKeepPriorityAndDefaultArgv(t *testing.T) {
	t.Parallel()

	prompt := "hello"
	expected := []backendExpectation{
		{
			name: "opencode",
			args: []string{"run", prompt},
			env:  []string{"OPENCODE_DISABLE_EXTERNAL_SKILLS=1"},
		},
		{name: "pi", args: []string{"-p", prompt}},
		{name: "codex", args: []string{"exec", "--ephemeral", prompt}},
		{name: "claude", args: []string{"--safe-mode", "-p", prompt}},
		{name: "agent", args: []string{"--mode", "ask", "-p", prompt}},
		{name: "gemini", args: []string{"-p", prompt}},
	}

	assertBackendsEqual(t, Supported(), expected, prompt, nil)
}

func TestSupportedBackendsRenderModelOverrides(t *testing.T) {
	t.Parallel()

	prompt := "literal $(echo bad)"
	model := &LLMInfo{
		Provider: "github-copilot",
		ID:       "gpt-5.5",
		Raw:      "github-copilot/gpt-5.5",
	}
	expected := []backendExpectation{
		{
			name: "opencode",
			args: []string{"run", "--model", model.Raw, prompt},
			env:  []string{"OPENCODE_DISABLE_EXTERNAL_SKILLS=1"},
		},
		{name: "pi", args: []string{"--model", model.Raw, "-p", prompt}},
		{name: "codex", args: []string{"exec", "--ephemeral", "--model", model.Raw, prompt}},
		{name: "claude", args: []string{"--safe-mode", "--model", model.Raw, "-p", prompt}},
		{name: "agent", args: []string{"--mode", "ask", "--model", model.Raw, "-p", prompt}},
		{name: "gemini", args: []string{"--model", model.Raw, "-p", prompt}},
	}

	assertBackendsEqual(t, Supported(), expected, prompt, model)
}

func TestDefaultModelPolicy(t *testing.T) {
	t.Parallel()

	prompt := "hello"

	tests := []struct {
		name string
		args []string
	}{
		{name: "opencode", args: []string{"run", prompt}},
		{name: "pi", args: []string{"-p", prompt}},
		{name: "codex", args: []string{"exec", "--ephemeral", prompt}},
		{name: "claude", args: []string{"--safe-mode", "-p", prompt}},
		{name: "agent", args: []string{"--mode", "ask", "-p", prompt}},
		{name: "gemini", args: []string{"-p", prompt}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			spec, ok := CommandForBackend(test.name, prompt, nil)

			assertEqual(t, ok, true)
			assertStringSlicesEqual(t, spec.Args, test.args)
		})
	}
}

func TestNoModelOverridePassesNoModelSelector(t *testing.T) {
	t.Parallel()

	prompt := "hello"

	for _, item := range Supported() {
		item := item
		t.Run(item.Name(), func(t *testing.T) {
			t.Parallel()

			spec, ok := CommandForBackend(item.Name(), prompt, nil)

			assertEqual(t, ok, true)
			for _, arg := range spec.Args {
				if arg == "--model" || arg == "-m" {
					t.Fatalf("unexpected default model selector in %#v", spec.Args)
				}
			}
		})
	}
}

func TestBackendArgsAndEnvReturnCopies(t *testing.T) {
	t.Parallel()

	backend := backendByName(t, "opencode")
	args := backend.Args("hello", nil)
	env := backend.Env()
	args[0] = "changed"
	env[0] = "changed"

	freshBackend := backendByName(t, "opencode")

	assertStringSlicesEqual(t, freshBackend.Args("hello", nil), []string{"run", "hello"})
	assertStringSlicesEqual(t, freshBackend.Env(), []string{"OPENCODE_DISABLE_EXTERNAL_SKILLS=1"})
}

func TestCommandForBackendRendersPromptAsSingleArgument(t *testing.T) {
	t.Parallel()

	prompt := "literal $(echo bad)"

	for _, item := range Supported() {
		item := item
		t.Run(item.Name(), func(t *testing.T) {
			t.Parallel()

			spec, ok := CommandForBackend(item.Name(), prompt, nil)

			assertEqual(t, ok, true)
			assertEqual(t, spec.Name, item.Name())
			assertStringSlicesEqual(t, spec.Args, item.Args(prompt, nil))
			assertStringSlicesEqual(t, spec.Env, item.Env())
		})
	}
}

func TestCommandForBackendReturnsCopies(t *testing.T) {
	t.Parallel()

	spec, ok := CommandForBackend("opencode", "hello", nil)
	assertEqual(t, ok, true)
	spec.Args[0] = "changed"
	spec.Env[0] = "changed"

	freshSpec, ok := CommandForBackend("opencode", "hello", nil)

	assertEqual(t, ok, true)
	assertStringSlicesEqual(t, freshSpec.Args, []string{"run", "hello"})
	assertStringSlicesEqual(t, freshSpec.Env, []string{"OPENCODE_DISABLE_EXTERNAL_SKILLS=1"})
}

func TestCommandForBackendRejectsUnsupportedBackend(t *testing.T) {
	t.Parallel()

	spec, ok := CommandForBackend("missing", "hello", nil)

	assertEqual(t, ok, false)
	assertEqual(t, spec.Name, "")
	assertEqual(t, len(spec.Args), 0)
}

func TestBackendSupportHelpers(t *testing.T) {
	t.Parallel()

	names := SupportedNames()

	assertStringSlicesEqual(
		t,
		names,
		[]string{"opencode", "pi", "codex", "claude", "agent", "gemini"},
	)
	assertEqual(t, IsSupported("opencode"), true)
	assertEqual(t, IsSupported("missing"), false)
	assertEqual(t, names[0], "opencode")
	assertEqual(t, names[len(names)-1], "gemini")
	assertEqual(t, SupportedCSV(), strings.Join(append(names, "..."), ", "))
}

func TestOSRunnerPassesPromptAsLiteralArgument(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	backendPath := filepath.Join(workDir, "backend")
	argvPath := filepath.Join(workDir, "argv")
	markerPath := filepath.Join(workDir, "marker")
	prompt := fmt.Sprintf("literal $(touch %s)", markerPath)
	script := fmt.Sprintf(`#!/bin/sh
printf '%%s\n' "$@" > %s
printf 'runner stdout\n'
`, shellQuote(argvPath))
	assertNoError(t, os.WriteFile(backendPath, []byte(script), 0o755))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := OSRunner{}.Run(
		context.Background(),
		CommandSpec{Name: backendPath, Args: []string{"-p", prompt}},
		&stdout,
		&stderr,
	)

	assertNoError(t, err)
	assertEqual(t, stdout.String(), "runner stdout\n")
	assertEqual(t, stderr.String(), "")
	if _, err := os.Stat(markerPath); !os.IsNotExist(err) {
		t.Fatalf("shell metacharacters were executed; marker stat err: %v", err)
	}
	argv, err := os.ReadFile(argvPath)
	assertNoError(t, err)
	assertEqual(t, string(argv), "-p\n"+prompt+"\n")
}

func TestOSRunnerPassesCommandEnvironment(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	backendPath := filepath.Join(workDir, "backend")
	script := `#!/bin/sh
printf '%s\n' "$OPENCODE_DISABLE_EXTERNAL_SKILLS"
`
	assertNoError(t, os.WriteFile(backendPath, []byte(script), 0o755))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := OSRunner{}.Run(
		context.Background(),
		CommandSpec{
			Name: backendPath,
			Env:  []string{"OPENCODE_DISABLE_EXTERNAL_SKILLS=1"},
		},
		&stdout,
		&stderr,
	)

	assertNoError(t, err)
	assertEqual(t, stdout.String(), "1\n")
	assertEqual(t, stderr.String(), "")
}

func TestOSRunnerConvertsNonZeroExitToBackendExitError(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	backendPath := filepath.Join(workDir, "backend")
	script := `#!/bin/sh
printf 'partial stdout\n'
printf 'backend stderr\n' >&2
exit 42
`
	assertNoError(t, os.WriteFile(backendPath, []byte(script), 0o755))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := OSRunner{}.Run(context.Background(), CommandSpec{Name: backendPath}, &stdout, &stderr)

	var exitErr ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected ExitError, got %T %[1]v", err)
	}
	assertEqual(t, exitErr.Backend, backendPath)
	assertEqual(t, exitErr.Code, 42)
	assertEqual(t, stdout.String(), "partial stdout\n")
	assertEqual(t, stderr.String(), "backend stderr\n")
}
