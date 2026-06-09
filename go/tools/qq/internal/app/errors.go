package app

import (
	"errors"
	"fmt"

	"github.com/BoscoDomingo/utils/go/tools/qq/internal/backend"
)

const (
	ExitUsage              = 64
	ExitBackendUnavailable = 127
)

type appError struct {
	code    int
	message string
}

func (err appError) Error() string {
	return err.message
}

func ExitCode(err error) int {
	if err == nil {
		return 0
	}

	var backendErr backend.ExitError
	if errors.As(err, &backendErr) {
		return backendErr.Code
	}

	var appErr appError
	if errors.As(err, &appErr) {
		return appErr.code
	}

	return 1
}

func (app *App) printNoBackendError() error {
	return app.printError(
		ExitBackendUnavailable,
		fmt.Sprintf("No supported backend found. Supported backends: %s", backend.SupportedCSV()),
	)
}

func (app *App) printError(code int, message string) error {
	fmt.Fprintln(app.stderr, message)
	return appError{code: code, message: message}
}
