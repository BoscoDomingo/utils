package cli

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"testing"

	appcore "github.com/BoscoDomingo/utils/go/tools/qq/internal/app"
	"github.com/BoscoDomingo/utils/go/tools/qq/internal/backend"
)

type commandResult struct {
	code   int
	stdout string
	stderr string
	err    error
}

type testCommandApp struct {
	app *appcore.App
}

func newTestCommandApp(t *testing.T, installed []string) *testCommandApp {
	t.Helper()

	runner := newFakeRunner()
	app := appcore.New(appcore.AppOptions{
		Stdout:       &bytes.Buffer{},
		Stderr:       &bytes.Buffer{},
		StdinIsTTY:   true,
		TTYAvailable: true,
		LookPath:     fakeLookPath(installed),
		Runner:       runner,
	})

	return &testCommandApp{app: app}
}

func (app *testCommandApp) execute(ctx context.Context, args ...string) commandResult {
	app.app.Stdout().Reset()
	app.app.Stderr().Reset()

	cmd := New(app.app, args)
	cmd.SetArgs(args)
	err := cmd.ExecuteContext(ctx)

	return commandResult{
		code:   appcore.ExitCode(err),
		stdout: app.app.Stdout().String(),
		stderr: app.app.Stderr().String(),
		err:    err,
	}
}

type fakeRunner struct{}

func newFakeRunner() *fakeRunner {
	return &fakeRunner{}
}

func (runner *fakeRunner) Run(
	_ context.Context,
	spec backend.CommandSpec,
	stdout *bytes.Buffer,
	stderr *bytes.Buffer,
) error {
	fmt.Fprintf(stdout, "answer from %s\n", spec.Name)
	return nil
}

func fakeLookPath(installed []string) func(string) (string, error) {
	installedSet := map[string]bool{}
	for _, name := range installed {
		installedSet[name] = true
	}

	return func(name string) (string, error) {
		if installedSet[name] {
			return "/fake/" + name, nil
		}
		return "", exec.ErrNotFound
	}
}

func assertExitCode(t *testing.T, result commandResult, expected int) {
	t.Helper()

	if result.code != expected {
		t.Fatalf(
			"exit code mismatch: got %d want %d stdout=%q stderr=%q err=%v",
			result.code,
			expected,
			result.stdout,
			result.stderr,
			result.err,
		)
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
