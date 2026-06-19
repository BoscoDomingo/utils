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

// LLMInfo describes the backend-specific data for a model.
type LLMInfo struct {
	Provider string
	ID       string
	Raw      string
}

type Backend interface {
	Name() backendName
	Args(prompt string, model *LLMInfo) []string
	Env() []string
}

type ModelLister interface {
	ListModels(context.Context) ([]LLMInfo, error)
}

type ModelProviderLister interface {
	ListModelProviders(context.Context) ([]string, error)
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
	"gemini",
}

type commandTemplate struct {
	Args          []string
	Env           []string
	ModelArgs     func([]string, *LLMInfo) []string
	ModelListArgs []string
}

type commandBackend struct {
	name     backendName
	template commandTemplate
}

func (backend commandBackend) Name() backendName {
	return backend.name
}

func (backend commandBackend) Args(prompt string, model *LLMInfo) []string {
	args := renderPrompt(backend.template.Args, prompt)
	if model == nil || model.Raw == "" || backend.template.ModelArgs == nil {
		return args
	}

	return backend.template.ModelArgs(args, model)
}

func (backend commandBackend) Env() []string {
	return append([]string(nil), backend.template.Env...)
}

var supportedBackends = map[backendName]commandTemplate{
	"opencode": {
		Args:          []string{"run", promptPlaceholder},
		Env:           []string{"OPENCODE_DISABLE_EXTERNAL_SKILLS=1"},
		ModelArgs:     insertModelArgs(1),
		ModelListArgs: []string{"models"},
	},
	"pi": {
		Args:          []string{"-p", promptPlaceholder},
		ModelArgs:     insertModelArgs(0),
		ModelListArgs: []string{"--list-models"},
	},
	"codex": {
		Args:      []string{"exec", "--ephemeral", promptPlaceholder},
		ModelArgs: insertModelArgs(2),
	},
	"claude": {
		Args:      []string{"--safe-mode", "-p", promptPlaceholder},
		ModelArgs: insertModelArgs(1),
	},
	"agent": {
		Args:          []string{"--mode", "ask", "-p", promptPlaceholder},
		ModelArgs:     insertModelArgs(2),
		ModelListArgs: []string{"models"},
	},
	"gemini": {
		Args:      []string{"-p", promptPlaceholder},
		ModelArgs: insertModelArgs(0),
	},
}

func Supported() []Backend {
	backends := make([]Backend, len(backendPriority))
	for index, name := range backendPriority {
		item := commandBackend{name: name, template: supportedBackends[name]}
		if len(item.template.ModelListArgs) > 0 {
			backends[index] = modelListingBackend{commandBackend: item}
			continue
		}
		backends[index] = item
	}
	return backends
}

func CommandForBackend(backendName backendName, prompt string, model *LLMInfo) (CommandSpec, bool) {
	template, ok := supportedBackends[backendName]
	if !ok {
		return CommandSpec{}, false
	}

	backend := commandBackend{name: backendName, template: template}
	return CommandSpec{
		Name: backendName,
		Args: backend.Args(prompt, model),
		Env:  backend.Env(),
	}, true
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

func renderPrompt(template []string, prompt string) []string {
	args := make([]string, len(template))
	for index, arg := range template {
		if arg == promptPlaceholder {
			args[index] = prompt
			continue
		}
		args[index] = arg
	}
	return args
}

func insertModelArgs(index int) func([]string, *LLMInfo) []string {
	return func(args []string, model *LLMInfo) []string {
		modelArgs := []string{"--model", model.Raw}
		result := make([]string, 0, len(args)+len(modelArgs))
		result = append(result, args[:index]...)
		result = append(result, modelArgs...)
		result = append(result, args[index:]...)
		return result
	}
}
