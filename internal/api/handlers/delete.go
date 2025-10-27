package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sblizard/vector-db/internal/engine"
)

type DeleteHandler struct {
	engine *engine.Engine
}

func NewDeleteHandler(engine *engine.Engine) *DeleteHandler {
	return &DeleteHandler{
		engine: engine,
	}
}

// DeleteVectorByIDHandler deletes a vector and its metadata by ID.
func (h *DeleteHandler) DeleteVectorByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	if id == "" {
		http.Error(w, "Vector ID is required", http.StatusBadRequest)
		return
	}

	deleted, err := h.engine.DeleteVectorByID(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete vector: %v", err), http.StatusInternalServerError)
		return
	}

	if !deleted {
		http.Error(w, fmt.Sprintf("Vector with ID %s not found", id), http.StatusNotFound)
		return
	}

	response := DeleteResponse{
		Status:     "success",
		Message:    fmt.Sprintf("Vector with ID %s deleted successfully", id),
		StatusCode: http.StatusOK,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}
