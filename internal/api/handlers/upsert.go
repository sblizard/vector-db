package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sblizard/vector-db/internal/storage"
)

type UpsertHandler struct {
	storage *storage.MetaStore
	layout  *storage.Layout
}

func NewUpsertHandler(store *storage.MetaStore, layout *storage.Layout) *UpsertHandler {
	return &UpsertHandler{
		storage: store,
		layout:  layout,
	}
}

func (h *UpsertHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	var req UpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.ID == "" {
		http.Error(w, "Vector ID is required", http.StatusBadRequest)
		return
	}

	if len(req.Vector) == 0 {
		http.Error(w, "Vector data is required", http.StatusBadRequest)
		return
	}

	fmt.Printf("Received upsert: ID=%s, VectorDim=%d\n", req.ID, len(req.Vector))

	vecPath := h.layout.VectorFile(0)

	existingIndex, err := h.storage.GetIndex(req.ID)
	isUpdate := err == nil

	if isUpdate {
		if len(req.Vector) != existingIndex.Dim {
			http.Error(w, fmt.Sprintf("Vector dimension mismatch: expected %d, got %d", existingIndex.Dim, len(req.Vector)), http.StatusBadRequest)
			return
		}

		if err := storage.WriteVectorAt(vecPath, req.Vector, existingIndex.Position); err != nil {
			http.Error(w, fmt.Sprintf("Failed to update vector: %v", err), http.StatusInternalServerError)
			return
		}
	} else {
		position, err := h.storage.GetNextPosition(len(req.Vector))
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get next position: %v", err), http.StatusInternalServerError)
			return
		}

		if err := storage.AppendVector(vecPath, req.Vector); err != nil {
			http.Error(w, fmt.Sprintf("Failed to store vector: %v", err), http.StatusInternalServerError)
			return
		}

		index := storage.VectorIndex{
			Position: position,
			Dim:      len(req.Vector),
		}
		if err := h.storage.PutIndex(req.ID, index); err != nil {
			http.Error(w, fmt.Sprintf("Failed to store index: %v", err), http.StatusInternalServerError)
			return
		}
	}

	metadataBytes, err := json.Marshal(req.Metadata)
	if err != nil {
		http.Error(w, "Failed to serialize metadata", http.StatusInternalServerError)
		return
	}

	if err := h.storage.Put(req.ID, metadataBytes); err != nil {
		http.Error(w, fmt.Sprintf("Failed to store metadata: %v", err), http.StatusInternalServerError)
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
	json.NewEncoder(w).Encode(resp)

	fmt.Printf("Successfully upserted vector: ID=%s, Dim=%d, IsUpdate=%v\n", req.ID, len(req.Vector), isUpdate)
}
