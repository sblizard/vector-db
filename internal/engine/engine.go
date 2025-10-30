package engine

import (
	"github.com/sblizard/vector-db/internal/storage"
)

type Engine struct {
	storage *storage.MetaStore
	layout  *storage.Layout
	topK    int
	dim     int
}

func NewEngine(store *storage.MetaStore, layout *storage.Layout, topK int, dim int) *Engine {
	return &Engine{
		storage: store,
		layout:  layout,
		topK:    topK,
		dim:     dim,
	}
}

func (e *Engine) GetDim() int {
	return e.dim
}

func (e *Engine) GetTopK() int {
	return e.topK
}
