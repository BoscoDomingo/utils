package app

import (
	"context"
	"fmt"
	"io"
	"os/exec"

	"github.com/BoscoDomingo/utils/go/tools/qq/internal/backend"
	"github.com/BoscoDomingo/utils/go/tools/qq/internal/invocation"
)

type App struct {
	stdout       io.Writer
	stderr       io.Writer
	stdin        io.Reader
	stdinIsTTY   bool
	ttyAvailable bool
	lookPath     func(string) (string, error)
	runner       backend.CommandRunner
	selector     BackendSelector
}

func (app *App) Run(ctx context.Context, args []string) error {
	stdinText, hasStdin, err := app.readPromptStdin(app.stdinReadMode(args))
	if err != nil {
		return app.printError(ExitUsage, fmt.Sprintf("read stdin: %v", err))
	}

	invocationResult, err := invocation.Parse(args, stdinText, hasStdin, backend.IsSupported)
	if err != nil {
		return app.printError(ExitUsage, err.Error())
	}

	backendName := invocationResult.BackendName
	if invocationResult.NeedsSelector {
		backendName, err = app.selectBackend(ctx)
		if err != nil {
			return err
		}
	}

	if backendName == "" {
		backendName, err = app.firstInstalledBackend()
		if err != nil {
			return app.printNoBackendError()
		}
	} else if _, err := app.lookPath(backendName); err != nil {
		return app.printError(
			ExitBackendUnavailable,
			fmt.Sprintf("Selected backend not found: %s", backendName),
		)
	}

	spec, ok := backend.CommandForBackend(
		backendName,
		invocationResult.Prompt,
		modelInfo(invocationResult.Model),
	)
	if !ok {
		return app.printError(
			ExitBackendUnavailable,
			fmt.Sprintf("unsupported backend: %s", backendName),
		)
	}

	err = app.runner.Run(ctx, spec, app.stdout, app.stderr)
	if err != nil {
		return err
	}

	return nil
}

func (app *App) selectBackend(ctx context.Context) (string, error) {
	if !app.ttyAvailable {
		return "", app.printError(ExitUsage, "interactive backend selection requires a TTY")
	}
	if app.selector == nil {
		return "", app.printError(ExitUsage, "backend selection failed: no selector configured")
	}

	backends := app.installedBackends()
	if len(backends) == 0 {
		return "", app.printNoBackendError()
	}

	backendName, err := app.selector.Select(ctx, backends)
	if err != nil {
		return "", app.printError(ExitUsage, fmt.Sprintf("backend selection failed: %v", err))
	}
	return backendName, nil
}

func (app *App) firstInstalledBackend() (string, error) {
	for _, item := range backend.Supported() {
		if _, err := app.lookPath(item.Name()); err == nil {
			return item.Name(), nil
		}
	}

	return "", exec.ErrNotFound
}

func (app *App) installedBackends() []backend.Backend {
	var installed []backend.Backend
	for _, item := range backend.Supported() {
		if _, err := app.lookPath(item.Name()); err == nil {
			installed = append(installed, item)
		}
	}
	return installed
}

func modelInfo(selector string) *backend.LLMInfo {
	if selector == "" {
		return nil
	}
	return &backend.LLMInfo{Raw: selector}
}
