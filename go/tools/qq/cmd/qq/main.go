package main

import (
	"bytes"
	"os"
	"os/exec"

	appcore "github.com/BoscoDomingo/utils/go/tools/qq/internal/app"
	"github.com/BoscoDomingo/utils/go/tools/qq/internal/backend"
	"github.com/BoscoDomingo/utils/go/tools/qq/internal/cli"
	"github.com/BoscoDomingo/utils/go/tools/qq/internal/selector"
)

func main() {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
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
		stderr.WriteString(err.Error())
		stderr.WriteByte('\n')
	}

	_, _ = os.Stdout.Write(stdout.Bytes())
	_, _ = os.Stderr.Write(stderr.Bytes())
	if err != nil {
		os.Exit(appcore.ExitCode(err))
	}
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
