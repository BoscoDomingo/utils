package app

import (
	"bytes"
	"context"
	"io"
	"os/exec"
	"strings"

	"github.com/BoscoDomingo/utils/go/tools/qq/internal/backend"
)

type BackendSelector interface {
	Select(context.Context, []backend.Backend) (string, error)
}

type AppOptions struct {
	Stdout       io.Writer
	Stderr       io.Writer
	Stdin        io.Reader
	StdinIsTTY   bool
	TTYAvailable bool
	LookPath     func(string) (string, error)
	Runner       backend.CommandRunner
	Selector     BackendSelector
}

func New(options AppOptions) *App {
	if options.Stdout == nil {
		options.Stdout = &bytes.Buffer{}
	}
	if options.Stderr == nil {
		options.Stderr = &bytes.Buffer{}
	}
	if options.Stdin == nil {
		options.Stdin = strings.NewReader("")
	}
	if options.LookPath == nil {
		options.LookPath = exec.LookPath
	}
	if options.Runner == nil {
		options.Runner = backend.OSRunner{}
	}

	return &App{
		stdout:       options.Stdout,
		stderr:       options.Stderr,
		stdin:        options.Stdin,
		stdinIsTTY:   options.StdinIsTTY,
		ttyAvailable: options.TTYAvailable,
		lookPath:     options.LookPath,
		runner:       options.Runner,
		selector:     options.Selector,
	}
}

func (app *App) Stdout() io.Writer {
	return app.stdout
}

func (app *App) Stderr() io.Writer {
	return app.stderr
}
