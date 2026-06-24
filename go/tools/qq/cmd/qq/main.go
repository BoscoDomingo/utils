package main

import (
	"os"
	"os/exec"

	appcore "github.com/BoscoDomingo/utils/go/tools/qq/internal/app"
	"github.com/BoscoDomingo/utils/go/tools/qq/internal/backend"
	"github.com/BoscoDomingo/utils/go/tools/qq/internal/cli"
	"github.com/BoscoDomingo/utils/go/tools/qq/internal/selector"
)

func main() {
	stdout := &countingWriter{writer: os.Stdout}
	stderr := &countingWriter{writer: os.Stderr}
	qqApp := appcore.New(appcore.AppOptions{
		Stdout:       stdout,
		Stderr:       stderr,
		Stdin:        os.Stdin,
		StdinIsTTY:   isTerminal(os.Stdin),
		TTYAvailable: canOpenTTY(),
		LookPath:     exec.LookPath,
		Runner:       backend.OSRunner{},
		Selector:     selector.New(),
	})

	args := os.Args[1:]
	cmd := cli.New(qqApp, args)
	cmd.SetArgs(args)
	err := cmd.Execute()
	if err != nil && stderr.Len() == 0 {
		_, _ = stderr.Write([]byte(err.Error() + "\n"))
	}

	if err != nil {
		os.Exit(appcore.ExitCode(err))
	}
}

type countingWriter struct {
	writer *os.File
	count  int
}

func (writer *countingWriter) Write(data []byte) (int, error) {
	written, err := writer.writer.Write(data)
	writer.count += written
	return written, err
}

func (writer *countingWriter) Len() int {
	return writer.count
}

func isTerminal(file *os.File) bool {
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func canOpenTTY() bool {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return false
	}
	_ = tty.Close()
	return true
}
