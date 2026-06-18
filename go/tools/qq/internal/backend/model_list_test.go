package backend

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestModelListingCapabilitiesAreOptional(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      backendName
		canList   bool
		providers []string
	}{
		{name: "opencode", canList: true, providers: []string{"Anthropic", "GitHub-Copilot"}},
		{name: "pi", canList: true, providers: []string{"GitHub-Copilot", "OpenAI"}},
		{name: "codex", canList: false},
		{name: "claude", canList: false},
		{name: "agent", canList: true, providers: []string{"Cursor", "OpenAI"}},
		{name: "gemini", canList: false},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			item := backendByName(t, test.name)
			lister, canList := item.(ModelLister)
			providerLister, canListProviders := item.(ModelProviderLister)

			assertEqual(t, canList, test.canList)
			assertEqual(t, canListProviders, test.canList)
			if !test.canList {
				return
			}

			models := parseModelList(sampleModelListOutput(test.name))
			assertStringSlicesEqual(t, providersFromModels(models), test.providers)
			_ = lister
			_ = providerLister
		})
	}
}

func TestModelListingRunsBackendNativeCommands(t *testing.T) {
	workDir := t.TempDir()
	originalPath := os.Getenv("PATH")
	t.Setenv("PATH", workDir+string(os.PathListSeparator)+originalPath)

	tests := []struct {
		name     backendName
		wantArgs []string
		output   string
		want     []LLMInfo
	}{
		{
			name:     "opencode",
			wantArgs: []string{"models"},
			output:   "Anthropic/claude-sonnet-4\nGitHub-Copilot/gpt-5.5\n",
			want: []LLMInfo{
				{Provider: "Anthropic", ID: "claude-sonnet-4", Raw: "Anthropic/claude-sonnet-4"},
				{Provider: "GitHub-Copilot", ID: "gpt-5.5", Raw: "GitHub-Copilot/gpt-5.5"},
			},
		},
		{
			name:     "pi",
			wantArgs: []string{"--list-models"},
			output:   "GitHub-Copilot/gpt-5.5\nOpenAI/gpt-5.4-mini\n",
			want: []LLMInfo{
				{Provider: "GitHub-Copilot", ID: "gpt-5.5", Raw: "GitHub-Copilot/gpt-5.5"},
				{Provider: "OpenAI", ID: "gpt-5.4-mini", Raw: "OpenAI/gpt-5.4-mini"},
			},
		},
		{
			name:     "agent",
			wantArgs: []string{"models"},
			output:   "Cursor/auto\nOpenAI/gpt-5.5\n",
			want: []LLMInfo{
				{Provider: "Cursor", ID: "auto", Raw: "Cursor/auto"},
				{Provider: "OpenAI", ID: "gpt-5.5", Raw: "OpenAI/gpt-5.5"},
			},
		},
	}

	for _, test := range tests {
		argvPath := filepath.Join(workDir, string(test.name)+".argv")
		writeFakeModelLister(t, workDir, test.name, argvPath, test.output)

		item := backendByName(t, test.name)
		lister, ok := item.(ModelLister)
		assertEqual(t, ok, true)

		models, err := lister.ListModels(context.Background())

		assertNoError(t, err)
		assertModelsEqual(t, models, test.want)
		argv, err := os.ReadFile(argvPath)
		assertNoError(t, err)
		assertStringSlicesEqual(t, splitLines(string(argv)), test.wantArgs)
	}
}

func TestModelProviderListingPreservesBackendProviderSpelling(t *testing.T) {
	t.Parallel()

	models := parseModelList("GitHub-Copilot/gpt-5.5\nOpenAI/gpt-5.4\nGitHub-Copilot/gpt-4.1\n")

	assertStringSlicesEqual(t, providersFromModels(models), []string{"GitHub-Copilot", "OpenAI"})
}

func sampleModelListOutput(name backendName) string {
	switch name {
	case "opencode":
		return "Anthropic/claude-sonnet-4\nGitHub-Copilot/gpt-5.5\n"
	case "pi":
		return "GitHub-Copilot/gpt-5.5\nOpenAI/gpt-5.4-mini\n"
	case "agent":
		return "Cursor/auto\nOpenAI/gpt-5.5\n"
	default:
		return ""
	}
}

func writeFakeModelLister(
	t *testing.T,
	dir string,
	name backendName,
	argvPath string,
	output string,
) {
	t.Helper()

	script := "#!/bin/sh\n" +
		"printf '%s\\n' \"$@\" > " + shellQuote(argvPath) + "\n" +
		"printf '%s' " + shellQuote(output) + "\n"
	path := filepath.Join(dir, string(name))
	assertNoError(t, os.WriteFile(path, []byte(script), 0o755))
}
