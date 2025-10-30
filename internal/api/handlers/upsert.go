package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/sblizard/vector-db/internal/engine"
)

type UpsertHandler struct {
	engine *engine.Engine
	dim    int
}

func NewUpsertHandler(engine *engine.Engine) *UpsertHandler {
	return &UpsertHandler{
		engine: engine,
		dim:    engine.GetDim(),
	}
}

func (h *UpsertHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	var req UpsertRequest

	req, paramsErr := h.extractUpsertParams(w, r)
	if paramsErr != nil {
		return
	}

	isUpdate, upsertErr := h.engine.Upsert(req.ID, req.Vector, req.Metadata)

	if upsertErr != nil {
		http.Error(w, fmt.Sprintf("Failed to upsert vector: %v", upsertErr), http.StatusInternalServerError)
		return
	}

	statusCode := 201
	message := "Vector inserted successfully"
	if isUpdate {
		statusCode = 200
		message = "Vector updated successfully"
	}

	resp := UpsertResponse{Status: "success", Message: message, StatusCode: statusCode}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(resp)

	fmt.Printf("Successfully upserted vector: ID=%s, Dim=%d, IsUpdate=%v\n", req.ID, len(req.Vector), isUpdate)
}

func (h *UpsertHandler) extractUpsertParams(w http.ResponseWriter, r *http.Request) (UpsertRequest, error) {
	var req UpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return UpsertRequest{}, err
	}

	if h.dim != len(req.Vector) {
		http.Error(w, fmt.Sprintf("Vector dimension mismatch: expected %d, got %d", h.dim, len(req.Vector)), http.StatusBadRequest)
		return UpsertRequest{}, fmt.Errorf("vector dimension mismatch: expected %d, got %d", h.dim, len(req.Vector))
	}

	if req.ID == "" {
		req.ID = uuid.New().String()
	}

	return req, nil
}
