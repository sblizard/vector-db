package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sblizard/vector-db/internal/storage"
)

type ReadHandler struct {
	storage *storage.MetaStore
	layout  *storage.Layout
}

func NewReadHandler(store *storage.MetaStore, layout *storage.Layout) *ReadHandler {
	return &ReadHandler{
		storage: store,
		layout:  layout,
	}
}

func (h *ReadHandler) GetAllVectors(w http.ResponseWriter, r *http.Request) {
	metaEntries, err := h.storage.GetAll()
	if err != nil {
		http.Error(w, "Failed to retrieve vectors", http.StatusInternalServerError)
		return
	}

	indices, err := h.storage.GetAllIndices()
	if err != nil {
		http.Error(w, "Failed to retrieve indices", http.StatusInternalServerError)
		return
	}

	if len(indices) == 0 {
		http.Error(w, "No vectors stored", http.StatusNotFound)
		return
	}

	vectors := make([]StoredVector, 0, len(indices))

	for id, index := range indices {
		normalizedVector, err := storage.ReadVectorAt(h.layout.VectorFile(0), index.Dim, index.Position)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read vector %s: %v", id, err), http.StatusInternalServerError)
			return
		}

		var metadata map[string]interface{}
		var originalVector []float32

		if metaBytes, ok := metaEntries[id]; ok {
			_ = json.Unmarshal(metaBytes, &metadata)
			originalVector = extractOriginalVector(metaBytes)
		}

		stringMetadata := make(map[string]string)
		for k, v := range metadata {
			if k != "original_vector" {
				if strVal, ok := v.(string); ok {
					stringMetadata[k] = strVal
				}
			}
		}

		vectors = append(vectors, StoredVector{
			ID:             id,
			Vector:         normalizedVector,
			OriginalVector: originalVector,
			Metadata:       stringMetadata,
		})
	}

	resp := GetAllResponse{Vectors: vectors}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)

	fmt.Printf("Returned %d vectors with metadata\n", len(vectors))
}

func extractOriginalVector(metaBytes []byte) []float32 {
	var metadata map[string]interface{}
	if err := json.Unmarshal(metaBytes, &metadata); err != nil {
		return nil
	}

	origVec, exists := metadata["original_vector"]
	if !exists {
		return nil
	}

	vecArray, ok := origVec.([]interface{})
	if !ok {
		return nil
	}

	originalVector := make([]float32, len(vecArray))
	for i, v := range vecArray {
		if floatVal, ok := v.(float64); ok {
			originalVector[i] = float32(floatVal)
		}
	}

	return originalVector
}
