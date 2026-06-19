package backend

import (
	"strings"
	"testing"
)

type backendExpectation struct {
	name backendName
	args []string
	env  []string
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func backendByName(t *testing.T, name backendName) Backend {
	t.Helper()

	for _, backend := range Supported() {
		if backend.Name() == name {
			return backend
		}
	}

	t.Fatalf("backend %q not found", name)
	return nil
}

func assertBackendsEqual(
	t *testing.T,
	got []Backend,
	want []backendExpectation,
	prompt string,
	model *LLMInfo,
) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("backend count mismatch: got %#v, want %#v", got, want)
	}
	for index := range got {
		assertEqual(t, got[index].Name(), want[index].name)
		assertStringSlicesEqual(t, got[index].Args(prompt, model), want[index].args)
		assertStringSlicesEqual(t, got[index].Env(), want[index].env)
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertEqual[T comparable](t *testing.T, got T, want T) {
	t.Helper()

	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func assertModelsEqual(t *testing.T, got []LLMInfo, want []LLMInfo) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("model count mismatch: got %#v, want %#v", got, want)
	}
	for index := range got {
		assertEqual(t, got[index].Provider, want[index].Provider)
		assertEqual(t, got[index].ID, want[index].ID)
		assertEqual(t, got[index].Raw, want[index].Raw)
	}
}

func assertStringSlicesEqual(t *testing.T, got []string, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("slice length mismatch: got %#v, want %#v", got, want)
	}
	for index := range got {
		if got[index] != want[index] {
			t.Fatalf("slice mismatch: got %#v, want %#v", got, want)
		}
	}
}

func splitLines(value string) []string {
	trimmed := strings.TrimSuffix(value, "\n")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "\n")
}
