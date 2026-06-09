package selector

import (
	"fmt"

	"github.com/BoscoDomingo/utils/go/tools/qq/internal/backend"
)

type backendItem backend.Backend

func newItems(backends []backend.Backend) []backendItem {
	items := make([]backendItem, len(backends))
	for index, backend := range backends {
		items[index] = backendItem(backend)
	}
	return items
}

func (item backendItem) Title() string {
	return item.Name
}

func (item backendItem) Description() string {
	return fmt.Sprintf("Run with %s", item.Name)
}

func (item backendItem) FilterValue() string {
	return item.Name
}
