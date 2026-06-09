package cli

import (
	"context"
	"testing"
)

func TestBackendFlagCompletionSuggestions(t *testing.T) {
	t.Parallel()

	app := newTestCommandApp(t, nil)

	result := app.execute(context.Background(), "__complete", "--backend=")

	assertExitCode(t, result, 0)
	assertContains(t, result.stdout, "opencode")
	assertContains(t, result.stdout, "claude")
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

	for _, shell := range []string{"bash", "zsh", "fish"} {
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
