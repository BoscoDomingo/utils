package invocation

import "testing"

func TestParseExplicitBackendFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		backend string
		prompt  string
	}{
		{"short backend", []string{"-b", "codex", "hey there"}, "codex", "hey there"},
		{"short provider", []string{"-p", "claude", "hey"}, "claude", "hey"},
		{"long backend", []string{"--backend", "agent", "hey"}, "agent", "hey"},
		{"long provider", []string{"--provider", "opencode", "hey"}, "opencode", "hey"},
		{"equals backend", []string{"--backend=gemini", "hey"}, "gemini", "hey"},
		{"equals provider", []string{"--provider=pi", "hey"}, "pi", "hey"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			invocation, err := Parse(test.args, "", false, testSupportsBackend)

			assertNoError(t, err)
			assertEqual(t, invocation.BackendName, test.backend)
			assertEqual(t, invocation.Prompt, test.prompt)
			assertEqual(t, invocation.NeedsSelector, false)
		})
	}
}

func TestParseSelectorFlagPreservesPrompt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		args   []string
		prompt string
	}{
		{"prompt after flag", []string{"-b", "explain cobra"}, "explain cobra"},
		{"prompt before flag", []string{"explain cobra", "-b"}, "explain cobra"},
		{"long flag before prompt", []string{"--provider", "explain cobra"}, "explain cobra"},
		{"long flag after prompt", []string{"explain cobra", "--backend"}, "explain cobra"},
		{
			"empty equals flag before prompt",
			[]string{"--backend=", "explain cobra"},
			"explain cobra",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			invocation, err := Parse(test.args, "", false, testSupportsBackend)

			assertNoError(t, err)
			assertEqual(t, invocation.BackendName, "")
			assertEqual(t, invocation.NeedsSelector, true)
			assertEqual(t, invocation.Prompt, test.prompt)
		})
	}
}

func TestParseStdinBackendFlagSelectsBackend(t *testing.T) {
	t.Parallel()

	invocation, err := Parse([]string{"-b", "claude"}, "hello\n", true, testSupportsBackend)

	assertNoError(t, err)
	assertEqual(t, invocation.BackendName, "claude")
	assertEqual(t, invocation.NeedsSelector, false)
	assertEqual(t, invocation.Prompt, "hello\n")
}

func TestParseModelFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		args   []string
		model  string
		prompt string
	}{
		{
			name:   "short model",
			args:   []string{"-m", "github-copilot/gpt-5.5", "hey there"},
			model:  "github-copilot/gpt-5.5",
			prompt: "hey there",
		},
		{
			name:   "long model",
			args:   []string{"--model", "anthropic/claude-sonnet-4", "hey"},
			model:  "anthropic/claude-sonnet-4",
			prompt: "hey",
		},
		{
			name:   "equals model",
			args:   []string{"--model=google/gemini-2.5-pro", "hey"},
			model:  "google/gemini-2.5-pro",
			prompt: "hey",
		},
		{
			name:   "model with backend",
			args:   []string{"-b", "pi", "-m", "github-copilot/gpt-5.5", "hey"},
			model:  "github-copilot/gpt-5.5",
			prompt: "hey",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			invocation, err := Parse(test.args, "", false, testSupportsBackend)

			assertNoError(t, err)
			assertEqual(t, invocation.Model, test.model)
			assertEqual(t, invocation.Prompt, test.prompt)
		})
	}
}

func TestParseMissingModelValueReturnsUsageError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{name: "long model missing value", args: []string{"--model"}},
		{name: "long model empty value", args: []string{"--model", "", "hey"}},
		{name: "short model empty value", args: []string{"-m", "", "hey"}},
		{name: "equals model empty value", args: []string{"--model=", "hey"}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := Parse(test.args, "", false, testSupportsBackend)

			assertErrorContains(t, err, "missing model value")
		})
	}
}

func TestParseArgsAndStdinUseContextSeparator(t *testing.T) {
	t.Parallel()

	invocation, err := Parse([]string{"summarize"}, "context\n", true, testSupportsBackend)

	assertNoError(t, err)
	assertEqual(t, invocation.Prompt, "summarize\n\nContext:\n\ncontext\n")
}

func TestUnsupportedExplicitProviderReturnsError(t *testing.T) {
	t.Parallel()

	_, err := Parse([]string{"--backend", "bogus", "prompt"}, "", false, testSupportsBackend)

	assertErrorContains(t, err, "unsupported backend")
}

func TestEmptyPromptReturnsUsageError(t *testing.T) {
	t.Parallel()

	_, err := Parse(nil, "", true, testSupportsBackend)

	assertErrorContains(t, err, "Usage:")
}

func TestBuildPrompt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		argsPrompt string
		stdinText  string
		hasStdin   bool
		want       string
	}{
		{"args only", "question", "", false, "question"},
		{"stdin only", "", "context\n", true, "context\n"},
		{"args and stdin", "summarize", "context\n", true, "summarize\n\nContext:\n\ncontext\n"},
		{"empty stdin with args", "question", "", true, "question"},
		{"no prompt", "", "", false, ""},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			assertEqual(t, buildPrompt(test.argsPrompt, test.stdinText, test.hasStdin), test.want)
		})
	}
}
