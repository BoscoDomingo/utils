package selector

import (
	"testing"

	"github.com/BoscoDomingo/utils/go/tools/qq/internal/backend"

	tea "charm.land/bubbletea/v2"
)

func updateModel(model selectorModel, msg tea.Msg) (selectorModel, tea.Cmd) {
	updated, cmd := model.Update(msg)
	selector, ok := updated.(selectorModel)
	if !ok {
		panic("selector model update returned wrong model type")
	}
	return selector, cmd
}

func assertQuitCommand(t *testing.T, cmd tea.Cmd) {
	t.Helper()

	if cmd == nil {
		t.Fatal("expected quit command")
	}
}

func testBackends() []backend.Backend {
	return []backend.Backend{
		{Name: "opencode"},
		{Name: "qwen"},
	}
}

func assertEqual[T comparable](t *testing.T, actual T, expected T) {
	t.Helper()

	if actual != expected {
		t.Fatalf("got %v, want %v", actual, expected)
	}
}
