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

	_, err := h.engine.DeleteVectorByID(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete vector: %v", err), http.StatusInternalServerError)
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
