package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sblizard/vector-db/internal/engine"
)

type SearchHandler struct {
	engine *engine.Engine
}

func NewSearchHandler(engine *engine.Engine) *SearchHandler {
	return &SearchHandler{
		engine: engine,
	}
}

func (h *SearchHandler) KClosestVectorsBruteHandler(w http.ResponseWriter, r *http.Request) {
	req, err := extractSearchParams(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	query := req.Vector
	k := req.TopK

	closestVectors, err := h.engine.KClosestVectorsBrute(query, k)
	if err != nil {
		http.Error(w, fmt.Sprintf("Search failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(closestVectors); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
	}
}

func extractSearchParams(body io.ReadCloser) (SearchRequest, error) {
	var req SearchRequest
	if err := json.NewDecoder(body).Decode(&req); err != nil {
		return req, err
	}
	if req.TopK <= 0 {
		return req, fmt.Errorf("TopK must be a positive integer")
	}
	if len(req.Vector) == 0 {
		return req, fmt.Errorf("query vector is required")
	}
	return req, nil
}
