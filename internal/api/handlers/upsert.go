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
}

func NewUpsertHandler(engine *engine.Engine) *UpsertHandler {
	return &UpsertHandler{
		engine: engine,
	}
}

func (h *UpsertHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	var req UpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.ID == "" {
		req.ID = uuid.New().String()
	}

	if len(req.Vector) == 0 {
		http.Error(w, "Vector data is required", http.StatusBadRequest)
		return
	}

	fmt.Printf("Received upsert: ID=%s, VectorDim=%d\n", req.ID, len(req.Vector))

	isUpdate, err := h.engine.Upsert(req.ID, req.Vector, req.Metadata)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to upsert vector: %v", err), http.StatusInternalServerError)
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
