package selector

import (
	"context"
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
)

func TestSelectRejectsEmptyBackendsBeforeOpeningTTY(t *testing.T) {
	_, err := New().Select(context.Background(), nil)
	if err == nil {
		t.Fatal("expected empty backend error")
	}
	assertEqual(t, err.Error(), "no installed backends available")
}

func TestModelEnterSelectsFocusedBackend(t *testing.T) {
	model, cmd := updateModel(
		newModel(testBackends(), list.NewDefaultDelegate()),
		tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}),
	)

	assertEqual(t, model.selected, "opencode")
	assertEqual(t, model.cancelled, false)
	assertQuitCommand(t, cmd)
}

func TestModelCancelKeysCancelSelection(t *testing.T) {
	tests := []struct {
		name string
		key  tea.KeyPressMsg
	}{
		{name: "esc", key: tea.KeyPressMsg(tea.Key{Code: tea.KeyEsc})},
		{name: "q", key: tea.KeyPressMsg(tea.Key{Code: 'q', Text: "q"})},
		{name: "ctrl+c", key: tea.KeyPressMsg(tea.Key{Code: 'c', Mod: tea.ModCtrl})},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, cmd := updateModel(newModel(testBackends(), list.NewDefaultDelegate()), tt.key)

			assertEqual(t, model.selected, "")
			assertEqual(t, model.cancelled, true)
			assertQuitCommand(t, cmd)
		})
	}
}

func TestItemTextMatchesBackend(t *testing.T) {
	item := backendItem{name: "gemini"}

	assertEqual(t, item.Title(), "gemini")
	assertEqual(t, item.Description(), "Run with gemini")
	assertEqual(t, item.FilterValue(), "gemini")
}

func TestModelViewIncludesTitle(t *testing.T) {
	view := newModel(testBackends(), list.NewDefaultDelegate()).View()

	if !strings.Contains(view.Content, "Select qq backend") {
		t.Fatalf("expected selector title in view, got %q", view.Content)
	}
}
