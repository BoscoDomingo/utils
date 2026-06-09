package app

import (
	"bytes"
	"io"
	"os"
	"strings"

	"github.com/BoscoDomingo/utils/go/tools/qq/internal/backend"
	"github.com/BoscoDomingo/utils/go/tools/qq/internal/invocation"

	"golang.org/x/sys/unix"
)

type stdinReadMode int

const (
	stdinReadNone stdinReadMode = iota
	stdinReadOptional
	stdinReadRequired
)

func (app *App) stdinReadMode(args []string) stdinReadMode {
	if app.stdinIsTTY {
		return stdinReadNone
	}

	invocationResult, err := invocation.Parse(args, "", false, backend.IsSupported)
	if err == nil && invocationResult.Prompt != "" {
		return stdinReadOptional
	}

	if err != nil && strings.HasPrefix(err.Error(), "Usage:") {
		return stdinReadRequired
	}

	return stdinReadNone
}

func (app *App) readPromptStdin(mode stdinReadMode) (string, bool, error) {
	if mode == stdinReadNone {
		return "", false, nil
	}
	if mode == stdinReadOptional && !stdinHasImmediateInput(app.stdin) {
		return "", false, nil
	}

	data, err := io.ReadAll(app.stdin)
	if err != nil {
		return "", false, err
	}

	return string(data), true, nil
}

func stdinHasImmediateInput(stdin io.Reader) bool {
	file, ok := stdin.(*os.File)
	if !ok {
		switch stdin.(type) {
		case *bytes.Buffer, *bytes.Reader, *strings.Reader:
			return true
		default:
			return false
		}
	}

	info, err := file.Stat()
	if err != nil {
		return false
	}
	if info.Mode().IsRegular() {
		return true
	}

	// Optional stdin must never drain a pipe until EOF is already observable.
	pollFileDescriptors := []unix.PollFd{{
		Fd:     int32(file.Fd()),
		Events: unix.POLLIN | unix.POLLHUP,
	}}
	ready, err := unix.Poll(pollFileDescriptors, 0)
	if err != nil {
		return false
	}
	if ready == 0 {
		return false
	}

	const eofReadyEvents = unix.POLLHUP | unix.POLLERR
	return pollFileDescriptors[0].Revents&eofReadyEvents != 0
}
