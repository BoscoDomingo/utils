package selector

import (
	"context"
	"errors"
	"os"

	"github.com/BoscoDomingo/utils/go/tools/qq/internal/backend"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
)

type Selector struct{}

func New() Selector {
	return Selector{}
}

func (selector Selector) Select(_ context.Context, backends []backend.Backend) (string, error) {
	if len(backends) == 0 {
		return "", errors.New("no installed backends available")
	}

	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", err
	}
	defer tty.Close()

	delegate := list.NewDefaultDelegate()
	model := newModel(backends, delegate)
	program := tea.NewProgram(
		model,
		tea.WithInput(tty),
		tea.WithOutput(tty),
	)
	finalModel, err := program.Run()
	if err != nil {
		return "", err
	}

	model, ok := finalModel.(selectorModel)
	if !ok || model.cancelled || model.selected == "" {
		return "", errors.New("selection cancelled")
	}
	return model.selected, nil
}
