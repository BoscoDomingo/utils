package selector

import (
	"fmt"

	"github.com/BoscoDomingo/utils/go/tools/qq/internal/backend"
)

type backendItem struct {
	name string
}

func newItems(backends []backend.Backend) []backendItem {
	items := make([]backendItem, len(backends))
	for index, backend := range backends {
		items[index] = backendItem{name: backend.Name()}
	}
	return items
}

func (item backendItem) Title() string {
	return item.name
}

func (item backendItem) Description() string {
	return fmt.Sprintf("Run with %s", item.name)
}

func (item backendItem) FilterValue() string {
	return item.name
}
