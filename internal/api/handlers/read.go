package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sblizard/vector-db/internal/engine"
)

type ReadHandler struct {
	engine *engine.Engine
}

func NewReadHandler(engine *engine.Engine) *ReadHandler {
	return &ReadHandler{
		engine: engine,
	}
}

func (h *ReadHandler) GetAllVectors(w http.ResponseWriter, r *http.Request) {
	getAllResp, err := h.engine.GetAllVectors()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get all vectors: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(getAllResp)
}
