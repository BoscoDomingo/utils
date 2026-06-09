package invocation

import (
	"strings"
	"testing"
)

func testSupportsBackend(name string) bool {
	for _, supported := range []string{"opencode", "pi", "claude", "copilot", "gemini", "qwen"} {
		if name == supported {
			return true
		}
	}
	return false
}

func assertNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertErrorContains(t *testing.T, err error, substring string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error containing %q", substring)
	}
	if !strings.Contains(err.Error(), substring) {
		t.Fatalf("expected error %q to contain %q", err.Error(), substring)
	}
}

func assertEqual[T comparable](t *testing.T, got T, want T) {
	t.Helper()

	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}
