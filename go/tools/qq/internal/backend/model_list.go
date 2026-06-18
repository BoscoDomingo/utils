package backend

import (
	"bytes"
	"context"
	"strings"
)

type modelListingBackend struct {
	commandBackend
}

func (backend modelListingBackend) ListModels(ctx context.Context) ([]LLMInfo, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := OSRunner{}.Run(
		ctx,
		CommandSpec{
			Name: backend.name,
			Args: append([]string(nil), backend.template.ModelListArgs...),
			Env:  backend.Env(),
		},
		&stdout,
		&stderr,
	)
	if err != nil {
		return nil, err
	}

	return parseModelList(stdout.String()), nil
}

func (backend modelListingBackend) ListModelProviders(ctx context.Context) ([]string, error) {
	models, err := backend.ListModels(ctx)
	if err != nil {
		return nil, err
	}
	return providersFromModels(models), nil
}

func parseModelList(output string) []LLMInfo {
	var models []LLMInfo
	for _, line := range strings.Split(output, "\n") {
		selector := modelSelectorFromLine(line)
		if selector == "" {
			continue
		}

		provider, id, ok := strings.Cut(selector, "/")
		if !ok || provider == "" || id == "" {
			continue
		}

		models = append(models, LLMInfo{
			Provider: provider,
			ID:       id,
			Raw:      selector,
		})
	}
	return models
}

func modelSelectorFromLine(line string) string {
	for _, field := range strings.Fields(line) {
		candidate := strings.Trim(field, " \t,;:\"'`()[]")
		candidate = strings.TrimPrefix(candidate, "-")
		candidate = strings.TrimPrefix(candidate, "*")
		if strings.Contains(candidate, "/") {
			return candidate
		}
	}
	return ""
}

func providersFromModels(models []LLMInfo) []string {
	seen := map[string]bool{}
	var providers []string
	for _, model := range models {
		if model.Provider == "" || seen[model.Provider] {
			continue
		}
		seen[model.Provider] = true
		providers = append(providers, model.Provider)
	}
	return providers
}
