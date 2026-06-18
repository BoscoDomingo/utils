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

// Pi keeps sticky model state; a provider-qualified model avoids OpenAI API fallback.
const piDefaultModel = "github-copilot/gpt-5.5"

type Backend struct {
	Name backendName
	Args []string
	Env  []string
}

type CommandSpec struct {
	Name backendName
	Args []string
	Env  []string
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
	cmd.Env = append(cmd.Environ(), spec.Env...)
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

type commandTemplate struct {
	Args []string
	Env  []string
}

var supportedBackends = map[backendName]commandTemplate{
	"opencode": {
		Args: []string{"run", promptPlaceholder},
		Env:  []string{"OPENCODE_DISABLE_EXTERNAL_SKILLS=1"},
	},
	"pi": {
		Args: []string{"--model", piDefaultModel, "-p", promptPlaceholder},
	},
	"codex":    {Args: []string{"exec", "--ephemeral", promptPlaceholder}},
	"claude":   {Args: []string{"--safe-mode", "-p", promptPlaceholder}},
	"agent":    {Args: []string{"--mode", "ask", "-p", promptPlaceholder}},
	"copilot":  {Args: []string{"-sp", promptPlaceholder}},
	"openclaw": {Args: []string{"agent", "--agent", "main", "--message", promptPlaceholder}},
	"gemini":   {Args: []string{"-p", promptPlaceholder}},
	"qwen":     {Args: []string{"-p", promptPlaceholder}},
	"q":        {Args: []string{"chat", "--non-interactive", promptPlaceholder}},
	"kimi":     {Args: []string{"--quiet", "-p", promptPlaceholder}},
	"kilo":     {Args: []string{"run", promptPlaceholder}},
	"kiro-cli": {Args: []string{"chat", "--no-interactive", promptPlaceholder}},
	"goose":    {Args: []string{"run", "--no-session", "-t", promptPlaceholder}},
	"aider":    {Args: []string{"--message", promptPlaceholder}},
	"amp":      {Args: []string{"-x", promptPlaceholder}},
	"droid":    {Args: []string{"exec", promptPlaceholder}},
	"crush":    {Args: []string{"run", "--quiet", promptPlaceholder}},
	"cn":       {Args: []string{"-p", promptPlaceholder, "--silent"}},
	"roo":      {Args: []string{"--print", promptPlaceholder}},
}

func Supported() []Backend {
	backends := make([]Backend, len(backendPriority))
	for index, name := range backendPriority {
		backends[index] = Backend{
			Name: name,
			Args: append([]string(nil), supportedBackends[name].Args...),
			Env:  append([]string(nil), supportedBackends[name].Env...),
		}
	}
	return backends
}

func CommandForBackend(backendName backendName, prompt string) (CommandSpec, bool) {
	template, ok := supportedBackends[backendName]
	if !ok {
		return CommandSpec{}, false
	}

	args := make([]string, len(template.Args))
	for index, arg := range template.Args {
		if arg == promptPlaceholder {
			args[index] = prompt
			continue
		}
		args[index] = arg
	}

	env := append([]string(nil), template.Env...)
	return CommandSpec{Name: backendName, Args: args, Env: env}, true
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
