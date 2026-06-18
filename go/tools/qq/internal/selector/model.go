package selector

import (
	"github.com/BoscoDomingo/utils/go/tools/qq/internal/backend"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
)

type selectorModel struct {
	list      list.Model
	selected  string
	cancelled bool
}

func newModel(backends []backend.Backend, delegate list.ItemDelegate) selectorModel {
	items := make([]list.Item, len(backends))
	for index, item := range newItems(backends) {
		items[index] = item
	}

	backendList := list.New(items, delegate, 80, 16)
	backendList.Title = "Select qq backend"
	backendList.SetShowStatusBar(false)
	backendList.SetFilteringEnabled(false)

	return selectorModel{list: backendList}
}

func (model selectorModel) Init() tea.Cmd {
	return nil
}

func (model selectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			item, ok := model.list.SelectedItem().(backendItem)
			if ok {
				model.selected = item.name
			}
			return model, tea.Quit
		case "esc", "q", "ctrl+c":
			model.cancelled = true
			return model, tea.Quit
		}
	}

	var cmd tea.Cmd
	model.list, cmd = model.list.Update(msg)
	return model, cmd
}

func (model selectorModel) View() tea.View {
	return tea.NewView(model.list.View())
}
