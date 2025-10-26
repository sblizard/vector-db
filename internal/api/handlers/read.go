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

	vecPath := h.layout.VectorFile(0)
	vectors := make([]StoredVector, 0, len(indices))

	for id, index := range indices {
		vector, err := storage.ReadVectorAt(vecPath, index.Dim, index.Position)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read vector %s: %v", id, err), http.StatusInternalServerError)
			return
		}

		var metadata map[string]string
		if metaBytes, ok := metaEntries[id]; ok {
			_ = json.Unmarshal(metaBytes, &metadata)
		}

		vectors = append(vectors, StoredVector{
			ID:       id,
			Vector:   vector,
			Metadata: metadata,
		})
	}

	resp := GetAllResponse{Vectors: vectors}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)

	fmt.Printf("Returned %d vectors with metadata\n", len(vectors))
}
