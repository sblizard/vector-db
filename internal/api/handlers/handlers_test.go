package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sblizard/vector-db/internal/storage"
)

func setupTestHandlers(t *testing.T) (*UpsertHandler, *ReadHandler, func()) {
	tmpDir := t.TempDir()
	store := storage.NewMetaStore(tmpDir)
	layout := storage.NewLayout(tmpDir)

	upsertHandler := NewUpsertHandler(store, layout)
	readHandler := NewReadHandler(store, layout)

	cleanup := func() {
		_ = store.Close()
	}

	return upsertHandler, readHandler, cleanup
}

func TestUpsertHandler_Insert(t *testing.T) {
	upsertHandler, _, cleanup := setupTestHandlers(t)
	defer cleanup()

	req := UpsertRequest{
		ID:       "vec1",
		Vector:   []float32{1.0, 2.0, 3.0},
		Metadata: map[string]string{"type": "test"},
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/upsert", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	upsertHandler.Upsert(w, httpReq)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var resp UpsertResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)

	if resp.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", resp.Status)
	}
}

func TestUpsertHandler_Update(t *testing.T) {
	upsertHandler, _, cleanup := setupTestHandlers(t)
	defer cleanup()

	// First insert
	req1 := UpsertRequest{
		ID:       "vec1",
		Vector:   []float32{1.0, 2.0, 3.0},
		Metadata: map[string]string{"type": "test"},
	}
	body1, _ := json.Marshal(req1)
	httpReq1 := httptest.NewRequest("POST", "/upsert", bytes.NewBuffer(body1))
	w1 := httptest.NewRecorder()
	upsertHandler.Upsert(w1, httpReq1)

	// Then update
	req2 := UpsertRequest{
		ID:       "vec1",
		Vector:   []float32{10.0, 20.0, 30.0},
		Metadata: map[string]string{"type": "updated"},
	}
	body2, _ := json.Marshal(req2)
	httpReq2 := httptest.NewRequest("POST", "/upsert", bytes.NewBuffer(body2))
	w2 := httptest.NewRecorder()
	upsertHandler.Upsert(w2, httpReq2)

	if w2.Code != http.StatusOK {
		t.Errorf("Expected status %d for update, got %d", http.StatusOK, w2.Code)
	}

	var resp UpsertResponse
	_ = json.NewDecoder(w2.Body).Decode(&resp)

	if resp.Message != "Vector updated successfully" {
		t.Errorf("Expected update message, got '%s'", resp.Message)
	}
}

func TestUpsertHandler_InvalidJSON(t *testing.T) {
	upsertHandler, _, cleanup := setupTestHandlers(t)
	defer cleanup()

	httpReq := httptest.NewRequest("POST", "/upsert", bytes.NewBufferString("invalid json"))
	w := httptest.NewRecorder()

	upsertHandler.Upsert(w, httpReq)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpsertHandler_MissingID(t *testing.T) {
	upsertHandler, _, cleanup := setupTestHandlers(t)
	defer cleanup()

	req := UpsertRequest{
		ID:     "",
		Vector: []float32{1.0, 2.0, 3.0},
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/upsert", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	upsertHandler.Upsert(w, httpReq)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpsertHandler_EmptyVector(t *testing.T) {
	upsertHandler, _, cleanup := setupTestHandlers(t)
	defer cleanup()

	req := UpsertRequest{
		ID:     "vec1",
		Vector: []float32{},
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/upsert", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	upsertHandler.Upsert(w, httpReq)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestReadHandler_GetAllVectors_Empty(t *testing.T) {
	_, readHandler, cleanup := setupTestHandlers(t)
	defer cleanup()

	httpReq := httptest.NewRequest("GET", "/vectors", nil)
	w := httptest.NewRecorder()

	readHandler.GetAllVectors(w, httpReq)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d for empty store, got %d", http.StatusNotFound, w.Code)
	}
}

func TestReadHandler_GetAllVectors_WithData(t *testing.T) {
	upsertHandler, readHandler, cleanup := setupTestHandlers(t)
	defer cleanup()

	// Insert test vectors
	vectors := []UpsertRequest{
		{ID: "vec1", Vector: []float32{1.0, 2.0}, Metadata: map[string]string{"label": "a"}},
		{ID: "vec2", Vector: []float32{3.0, 4.0}, Metadata: map[string]string{"label": "b"}},
	}

	for _, req := range vectors {
		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("POST", "/upsert", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		upsertHandler.Upsert(w, httpReq)
	}

	// Get all vectors
	httpReq := httptest.NewRequest("GET", "/vectors", nil)
	w := httptest.NewRecorder()
	readHandler.GetAllVectors(w, httpReq)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp GetAllResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)

	if len(resp.Vectors) != len(vectors) {
		t.Errorf("Expected %d vectors, got %d", len(vectors), len(resp.Vectors))
	}
}
