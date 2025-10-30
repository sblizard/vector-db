package engine

import (
	"github.com/sblizard/vector-db/internal/storage"
)

type Engine struct {
	storage *storage.MetaStore
	layout  *storage.Layout
	topK    *int
	dim     *int
}

func NewEngine(store *storage.MetaStore, layout *storage.Layout, topK int, dim int) *Engine {
	return &Engine{
		storage: store,
		layout:  layout,
		topK:    &topK,
		dim:     &dim,
	}
}

func (e *Engine) GetDim() int {
	if e.dim == nil {
		return 0
	}
	return *e.dim
}

func (e *Engine) GetTopK() int {
	if e.topK == nil {
		return 0
	}
	return *e.topK
}
