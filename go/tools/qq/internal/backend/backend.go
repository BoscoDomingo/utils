package backend

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type backendName = string

const promptPlaceholder = "{prompt}"

type Backend struct {
	Name backendName
	Args []string
}

type CommandSpec struct {
	Name backendName
	Args []string
}

type CommandRunner interface {
	Run(context.Context, CommandSpec, *bytes.Buffer, *bytes.Buffer) error
}

type ExitError struct {
	Backend backendName
	Code    int
}

func (err ExitError) Error() string {
	return fmt.Sprintf("%s failed", err.Backend)
}

type OSRunner struct{}

func (runner OSRunner) Run(
	ctx context.Context,
	spec CommandSpec,
	stdout *bytes.Buffer,
	stderr *bytes.Buffer,
) error {
	cmd := exec.CommandContext(ctx, string(spec.Name), spec.Args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()
	if err == nil {
		return nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return ExitError{Backend: spec.Name, Code: exitErr.ExitCode()}
	}

	return err
}

var backendPriority = []backendName{
	"opencode",
	"pi",
	"codex",
	"claude",
	"agent",
	"copilot",
	"openclaw",
	"gemini",
	"qwen",
	"q",
	"kimi",
	"kilo",
	"kiro-cli",
	"goose",
	"aider",
	"amp",
	"droid",
	"crush",
	"cn",
	"roo",
}

var supportedBackends = map[backendName][]string{
	"opencode": {"run", promptPlaceholder},
	"pi":       {"-p", promptPlaceholder},
	"codex":    {"exec", "--ephemeral", promptPlaceholder},
	"claude":   {"-p", promptPlaceholder},
	"agent":    {"--mode", "ask", "-p", promptPlaceholder},
	"copilot":  {"-sp", promptPlaceholder},
	"openclaw": {"agent", "--agent", "main", "--message", promptPlaceholder},
	"gemini":   {"-p", promptPlaceholder},
	"qwen":     {"-p", promptPlaceholder},
	"q":        {"chat", "--non-interactive", promptPlaceholder},
	"kimi":     {"--quiet", "-p", promptPlaceholder},
	"kilo":     {"run", promptPlaceholder},
	"kiro-cli": {"chat", "--no-interactive", promptPlaceholder},
	"goose":    {"run", "--no-session", "-t", promptPlaceholder},
	"aider":    {"--message", promptPlaceholder},
	"amp":      {"-x", promptPlaceholder},
	"droid":    {"exec", promptPlaceholder},
	"crush":    {"run", "--quiet", promptPlaceholder},
	"cn":       {"-p", promptPlaceholder, "--silent"},
	"roo":      {"--print", promptPlaceholder},
}

func Supported() []Backend {
	backends := make([]Backend, len(backendPriority))
	for index, name := range backendPriority {
		backends[index] = Backend{
			Name: name,
			Args: append([]string(nil), supportedBackends[name]...),
		}
	}
	return backends
}

func CommandForBackend(backendName backendName, prompt string) (CommandSpec, bool) {
	template, ok := supportedBackends[backendName]
	if !ok {
		return CommandSpec{}, false
	}

	args := make([]string, len(template))
	for index, arg := range template {
		if arg == promptPlaceholder {
			args[index] = prompt
			continue
		}
		args[index] = arg
	}

	return CommandSpec{Name: backendName, Args: args}, true
}

func IsSupported(name backendName) bool {
	_, ok := supportedBackends[name]
	return ok
}

func SupportedNames() []backendName {
	names := append([]backendName(nil), backendPriority...)
	return names
}

func SupportedCSV() string {
	names := SupportedNames()
	parts := make([]string, 0, len(names)+1)
	for _, name := range names {
		parts = append(parts, string(name))
	}
	parts = append(parts, "...")

	return strings.Join(parts, ", ")
}
