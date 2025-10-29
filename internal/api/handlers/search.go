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
	topK   int
}

func NewSearchHandler(engine *engine.Engine, topK int) *SearchHandler {
	return &SearchHandler{
		engine: engine,
		topK:   topK,
	}
}

func (h *SearchHandler) KClosestVectorsBruteHandler(w http.ResponseWriter, r *http.Request) {
	req, err := h.extractSearchParams(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	query := req.Vector
	k := req.TopK
	fmt.Printf("TopK:\n%d\n", k)

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

func (h *SearchHandler) extractSearchParams(body io.ReadCloser) (SearchRequest, error) {
	var req SearchRequest
	if err := json.NewDecoder(body).Decode(&req); err != nil {
		return req, err
	}
	if req.TopK <= 0 {
		req.TopK = h.topK
	}
	if len(req.Vector) == 0 {
		return req, fmt.Errorf("query vector is required")
	}
	return req, nil
}
