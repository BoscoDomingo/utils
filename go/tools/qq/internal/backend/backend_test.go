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

func TestSupportedBackendsKeepPriorityAndArgvTemplates(t *testing.T) {
	t.Parallel()

	expected := []Backend{
		{
			Name: "opencode",
			Args: []string{"run", promptPlaceholder},
			Env:  []string{"OPENCODE_DISABLE_EXTERNAL_SKILLS=1"},
		},
		{Name: "pi", Args: []string{"--model", piDefaultModel, "-p", promptPlaceholder}},
		{Name: "codex", Args: []string{"exec", "--ephemeral", promptPlaceholder}},
		{Name: "claude", Args: []string{"--safe-mode", "-p", promptPlaceholder}},
		{Name: "agent", Args: []string{"--mode", "ask", "-p", promptPlaceholder}},
		{Name: "copilot", Args: []string{"-sp", promptPlaceholder}},
		{Name: "openclaw", Args: []string{"agent", "--agent", "main", "--message", promptPlaceholder}},
		{Name: "gemini", Args: []string{"-p", promptPlaceholder}},
		{Name: "qwen", Args: []string{"-p", promptPlaceholder}},
		{Name: "q", Args: []string{"chat", "--non-interactive", promptPlaceholder}},
		{Name: "kimi", Args: []string{"--quiet", "-p", promptPlaceholder}},
		{Name: "kilo", Args: []string{"run", promptPlaceholder}},
		{Name: "kiro-cli", Args: []string{"chat", "--no-interactive", promptPlaceholder}},
		{Name: "goose", Args: []string{"run", "--no-session", "-t", promptPlaceholder}},
		{Name: "aider", Args: []string{"--message", promptPlaceholder}},
		{Name: "amp", Args: []string{"-x", promptPlaceholder}},
		{Name: "droid", Args: []string{"exec", promptPlaceholder}},
		{Name: "crush", Args: []string{"run", "--quiet", promptPlaceholder}},
		{Name: "cn", Args: []string{"-p", promptPlaceholder, "--silent"}},
		{Name: "roo", Args: []string{"--print", promptPlaceholder}},
	}

	assertBackendsEqual(t, Supported(), expected)
}

func TestSupportedReturnsCopy(t *testing.T) {
	t.Parallel()

	backends := Supported()
	backends[0].Name = "changed"
	backends[0].Args[0] = "changed"
	backends[0].Env[0] = "changed"

	freshBackends := Supported()

	assertEqual(t, freshBackends[0].Name, "opencode")
	assertStringSlicesEqual(t, freshBackends[0].Args, []string{"run", promptPlaceholder})
	assertStringSlicesEqual(t, freshBackends[0].Env, []string{"OPENCODE_DISABLE_EXTERNAL_SKILLS=1"})
}

func TestCommandForBackendRendersPromptAsSingleArgument(t *testing.T) {
	t.Parallel()

	prompt := "literal $(echo bad)"

	for _, item := range Supported() {
		item := item
		t.Run(item.Name, func(t *testing.T) {
			t.Parallel()

			spec, ok := CommandForBackend(item.Name, prompt)

			assertEqual(t, ok, true)
			assertEqual(t, spec.Name, item.Name)
			assertStringSlicesEqual(t, spec.Args, expectedArgv(item.Args, prompt))
			assertStringSlicesEqual(t, spec.Env, item.Env)
		})
	}
}

func TestCommandForBackendReturnsCopies(t *testing.T) {
	t.Parallel()

	spec, ok := CommandForBackend("opencode", "hello")
	assertEqual(t, ok, true)
	spec.Args[0] = "changed"
	spec.Env[0] = "changed"

	freshSpec, ok := CommandForBackend("opencode", "hello")

	assertEqual(t, ok, true)
	assertStringSlicesEqual(t, freshSpec.Args, []string{"run", "hello"})
	assertStringSlicesEqual(t, freshSpec.Env, []string{"OPENCODE_DISABLE_EXTERNAL_SKILLS=1"})
}

func TestCommandForBackendRejectsUnsupportedBackend(t *testing.T) {
	t.Parallel()

	spec, ok := CommandForBackend("missing", "hello")

	assertEqual(t, ok, false)
	assertEqual(t, spec.Name, "")
	assertEqual(t, len(spec.Args), 0)
}

func TestBackendSupportHelpers(t *testing.T) {
	t.Parallel()

	names := SupportedNames()

	assertEqual(t, IsSupported("opencode"), true)
	assertEqual(t, IsSupported("missing"), false)
	assertEqual(t, names[0], "opencode")
	assertEqual(t, names[len(names)-1], "roo")
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
