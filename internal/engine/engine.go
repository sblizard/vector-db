package engine

import (
	"github.com/sblizard/vector-db/internal/storage"
)

type Engine struct {
	storage *storage.MetaStore
	layout  *storage.Layout
}

func NewEngine(store *storage.MetaStore, layout *storage.Layout) *Engine {
	return &Engine{
		storage: store,
		layout:  layout,
	}
}
