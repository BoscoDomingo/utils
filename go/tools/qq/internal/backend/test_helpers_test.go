package backend

import (
	"strings"
	"testing"
)

func expectedArgv(template []string, prompt string) []string {
	argv := make([]string, len(template))
	for index, arg := range template {
		if arg == promptPlaceholder {
			argv[index] = prompt
			continue
		}
		argv[index] = arg
	}
	return argv
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func assertBackendsEqual(t *testing.T, got []Backend, want []Backend) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("backend count mismatch: got %#v, want %#v", got, want)
	}
	for index := range got {
		assertEqual(t, got[index].Name, want[index].Name)
		assertStringSlicesEqual(t, got[index].Args, want[index].Args)
		assertStringSlicesEqual(t, got[index].Env, want[index].Env)
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
