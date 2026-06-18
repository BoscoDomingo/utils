package cli

import (
	"context"
	"strings"
	"testing"
)

func TestBackendFlagCompletionSuggestions(t *testing.T) {
	t.Parallel()

	for _, args := range [][]string{
		{"__complete", "--backend="},
		{"__complete", "--provider="},
		{"__complete", "--backend", ""},
		{"__complete", "--provider", ""},
		{"__complete", "-b", ""},
		{"__complete", "-p", ""},
		{"__completeNoDesc", "-b", ""},
	} {
		args := args
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			t.Parallel()

			app := newTestCommandApp(t, nil)

			result := app.execute(context.Background(), args...)

			assertExitCode(t, result, 0)
			assertContains(t, result.stdout, "opencode")
			assertContains(t, result.stdout, "claude")
			assertContains(t, result.stdout, ":4")
		})
	}
}

func TestFlagNameCompletionAfterDash(t *testing.T) {
	t.Parallel()

	app := newTestCommandApp(t, nil)
	result := app.execute(context.Background(), "__complete", "-")

	assertExitCode(t, result, 0)
	assertContains(t, result.stdout, "--backend")
	assertContains(t, result.stdout, "-b")
	assertContains(t, result.stdout, "--provider")
	assertContains(t, result.stdout, "-p")
	assertContains(t, result.stdout, ":4")
}

func TestPromptCompletionDisablesFileCompletion(t *testing.T) {
	t.Parallel()

	app := newTestCommandApp(t, nil)

	result := app.execute(context.Background(), "__complete", "prompt")

	assertExitCode(t, result, 0)
	assertContains(t, result.stdout, ":4")
}

func TestBackendFlagsUseSelectorSentinelNoOptDefault(t *testing.T) {
	t.Parallel()

	cmd := New(newTestCommandApp(t, nil).app, nil)

	for _, name := range []string{"backend", "provider"} {
		flag := cmd.Flags().Lookup(name)
		if flag == nil {
			t.Fatalf("missing %s flag", name)
		}
		assertEqual(t, flag.NoOptDefVal, selectorFlagValue)
	}
}

func TestCompletionCommandsGenerateShellScripts(t *testing.T) {
	t.Parallel()

	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		shell := shell
		t.Run(shell, func(t *testing.T) {
			t.Parallel()

			app := newTestCommandApp(t, nil)

			result := app.execute(context.Background(), "completion", shell)

			assertExitCode(t, result, 0)
			assertContains(t, result.stdout, "qq")
		})
	}
}
