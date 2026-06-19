package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/BoscoDomingo/utils/go/tools/qq/internal/backend"
)

type commandResult struct {
	code   int
	stdout string
	stderr string
	err    error
}

type testApp struct {
	*App
	runner   *fakeRunner
	selector *fakeSelector
}

func newTestApp(t *testing.T, installed []string) *testApp {
	t.Helper()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	runner := newFakeRunner()
	selector := &fakeSelector{}
	app := New(AppOptions{
		Stdout:       stdout,
		Stderr:       stderr,
		Stdin:        strings.NewReader(""),
		StdinIsTTY:   true,
		TTYAvailable: true,
		LookPath:     fakeLookPath(installed),
		Runner:       runner,
		Selector:     selector,
	})

	return &testApp{App: app, runner: runner, selector: selector}
}

func (app *testApp) execute(ctx context.Context, args ...string) commandResult {
	app.stdout.Reset()
	app.stderr.Reset()

	err := app.Run(ctx, args)

	return commandResult{
		code:   ExitCode(err),
		stdout: app.stdout.String(),
		stderr: app.stderr.String(),
		err:    err,
	}
}

type fakeRunner struct {
	calls     []runnerCall
	exitCodes map[string]int
	stdout    map[string]string
}

type runnerCall struct {
	spec backend.CommandSpec
}

func newFakeRunner() *fakeRunner {
	return &fakeRunner{
		exitCodes: map[string]int{},
		stdout:    map[string]string{},
	}
}

func (runner *fakeRunner) Run(
	_ context.Context,
	spec backend.CommandSpec,
	stdout *bytes.Buffer,
	stderr *bytes.Buffer,
) error {
	runner.calls = append(runner.calls, runnerCall{spec: spec})
	output := runner.stdout[spec.Name]
	if output == "" {
		output = fmt.Sprintf("answer from %s\n", spec.Name)
	}
	stdout.WriteString(output)

	if code := runner.exitCodes[spec.Name]; code != 0 {
		fmt.Fprintf(stderr, "%s failed\n", spec.Name)
		return backend.ExitError{Backend: spec.Name, Code: code}
	}

	return nil
}

type fakeSelector struct {
	selected string
}

func (selector *fakeSelector) Select(
	_ context.Context,
	backends []backend.Backend,
) (string, error) {
	if selector.selected != "" {
		return selector.selected, nil
	}
	if len(backends) == 0 {
		return "", errors.New("no backends to select")
	}
	return backends[0].Name(), nil
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

func assertBackendCalls(t *testing.T, got []runnerCall, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("backend call count mismatch: got %#v, want %#v", got, want)
	}
	for index := range got {
		if got[index].spec.Name != want[index] {
			t.Fatalf("backend calls mismatch: got %#v, want %#v", got, want)
		}
	}
}
