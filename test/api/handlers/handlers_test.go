package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/sblizard/vector-db/internal/api/handlers"
	"github.com/sblizard/vector-db/internal/engine"
	"github.com/sblizard/vector-db/internal/storage"
)

func setupTestHandlers(t *testing.T) (*handlers.UpsertHandler, *handlers.ReadHandler, *handlers.DeleteHandler, func()) {
	tmpDir := t.TempDir()
	store := storage.NewMetaStore(tmpDir)
	layout := storage.NewLayout(tmpDir)
	engine := engine.NewEngine(store, layout)

	upsertHandler := handlers.NewUpsertHandler(engine)
	readHandler := handlers.NewReadHandler(engine)
	deleteHandler := handlers.NewDeleteHandler(engine)

	cleanup := func() {
		_ = store.Close()
	}

	return upsertHandler, readHandler, deleteHandler, cleanup
}

func TestUpsertHandler_Insert(t *testing.T) {
	upsertHandler, _, _, cleanup := setupTestHandlers(t)
	defer cleanup()

	req := handlers.UpsertRequest{
		ID:       "vec1",
		Vector:   []float32{1.0, 2.0, 3.0},
		Metadata: map[string]interface{}{"type": "test"},
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/upsert", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	upsertHandler.Upsert(w, httpReq)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var resp handlers.UpsertResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)

	if resp.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", resp.Status)
	}
}

func TestUpsertHandler_Update(t *testing.T) {
	upsertHandler, _, _, cleanup := setupTestHandlers(t)
	defer cleanup()

	// First insert
	req1 := handlers.UpsertRequest{
		ID:       "vec1",
		Vector:   []float32{1.0, 2.0, 3.0},
		Metadata: map[string]interface{}{"type": "test"},
	}
	body1, _ := json.Marshal(req1)
	httpReq1 := httptest.NewRequest("POST", "/upsert", bytes.NewBuffer(body1))
	w1 := httptest.NewRecorder()
	upsertHandler.Upsert(w1, httpReq1)

	// Then update
	req2 := handlers.UpsertRequest{
		ID:       "vec1",
		Vector:   []float32{10.0, 20.0, 30.0},
		Metadata: map[string]interface{}{"type": "updated"},
	}
	body2, _ := json.Marshal(req2)
	httpReq2 := httptest.NewRequest("POST", "/upsert", bytes.NewBuffer(body2))
	w2 := httptest.NewRecorder()
	upsertHandler.Upsert(w2, httpReq2)

	if w2.Code != http.StatusOK {
		t.Errorf("Expected status %d for update, got %d", http.StatusOK, w2.Code)
	}

	var resp handlers.UpsertResponse
	_ = json.NewDecoder(w2.Body).Decode(&resp)

	if resp.Message != "Vector updated successfully" {
		t.Errorf("Expected update message, got '%s'", resp.Message)
	}
}

func TestUpsertHandler_InvalidJSON(t *testing.T) {
	upsertHandler, _, _, cleanup := setupTestHandlers(t)
	defer cleanup()

	httpReq := httptest.NewRequest("POST", "/upsert", bytes.NewBufferString("invalid json"))
	w := httptest.NewRecorder()

	upsertHandler.Upsert(w, httpReq)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpsertHandler_MissingID(t *testing.T) {
	upsertHandler, _, _, cleanup := setupTestHandlers(t)
	defer cleanup()

	req := handlers.UpsertRequest{
		ID:     "",
		Vector: []float32{1.0, 2.0, 3.0},
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/upsert", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	upsertHandler.Upsert(w, httpReq)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var resp handlers.UpsertResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)

	if resp.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", resp.Status)
	}
}

func TestUpsertHandler_EmptyVector(t *testing.T) {
	upsertHandler, _, _, cleanup := setupTestHandlers(t)
	defer cleanup()

	req := handlers.UpsertRequest{
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
	_, readHandler, _, cleanup := setupTestHandlers(t)
	defer cleanup()

	httpReq := httptest.NewRequest("GET", "/vectors", nil)
	w := httptest.NewRecorder()

	readHandler.GetAllVectors(w, httpReq)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d for empty store, got %d", http.StatusOK, w.Code)
	}

	var resp handlers.GetAllResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)

	if len(resp.Vectors) != 0 {
		t.Errorf("Expected empty vector array, got %d vectors", len(resp.Vectors))
	}
}

func TestReadHandler_GetAllVectors_WithData(t *testing.T) {
	upsertHandler, readHandler, _, cleanup := setupTestHandlers(t)
	defer cleanup()

	// Insert test vectors
	vectors := []handlers.UpsertRequest{
		{ID: "vec1", Vector: []float32{1.0, 2.0}, Metadata: map[string]interface{}{"label": "a"}},
		{ID: "vec2", Vector: []float32{3.0, 4.0}, Metadata: map[string]interface{}{"label": "b"}},
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

	var resp handlers.GetAllResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)

	if len(resp.Vectors) != len(vectors) {
		t.Errorf("Expected %d vectors, got %d", len(vectors), len(resp.Vectors))
	}
}

// Delete Handler Tests

func TestDeleteHandler_DeleteExistingVector(t *testing.T) {
	upsertHandler, readHandler, deleteHandler, cleanup := setupTestHandlers(t)
	defer cleanup()

	// Insert a vector
	req := handlers.UpsertRequest{
		ID:       "vec1",
		Vector:   []float32{1.0, 2.0, 3.0},
		Metadata: map[string]interface{}{"type": "test", "label": "delete_me"},
	}
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/upsert", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	upsertHandler.Upsert(w, httpReq)

	// Delete the vector
	httpReq = httptest.NewRequest("DELETE", "/vector/vec1", nil)
	httpReq = mux.SetURLVars(httpReq, map[string]string{"id": "vec1"})
	w = httptest.NewRecorder()
	deleteHandler.DeleteVectorByIDHandler(w, httpReq)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp handlers.DeleteResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)

	if resp.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", resp.Status)
	}

	if resp.Message != "Vector with ID vec1 deleted successfully" {
		t.Errorf("Expected delete success message, got '%s'", resp.Message)
	}

	// Verify vector is deleted by trying to get all vectors
	httpReq = httptest.NewRequest("GET", "/vectors", nil)
	w = httptest.NewRecorder()
	readHandler.GetAllVectors(w, httpReq)

	var getAllResp handlers.GetAllResponse
	_ = json.NewDecoder(w.Body).Decode(&getAllResp)

	if len(getAllResp.Vectors) != 0 {
		t.Errorf("Expected 0 vectors after deletion, got %d", len(getAllResp.Vectors))
	}
}

func TestDeleteHandler_DeleteNonExistentVector(t *testing.T) {
	_, _, deleteHandler, cleanup := setupTestHandlers(t)
	defer cleanup()

	// Try to delete a vector that doesn't exist
	httpReq := httptest.NewRequest("DELETE", "/vector/nonexistent", nil)
	httpReq = mux.SetURLVars(httpReq, map[string]string{"id": "nonexistent"})
	w := httptest.NewRecorder()
	deleteHandler.DeleteVectorByIDHandler(w, httpReq)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d for non-existent vector, got %d", http.StatusNotFound, w.Code)
	}
}

func TestDeleteHandler_DeleteMultipleVectors(t *testing.T) {
	upsertHandler, readHandler, deleteHandler, cleanup := setupTestHandlers(t)
	defer cleanup()

	// Insert multiple vectors
	vectors := []handlers.UpsertRequest{
		{ID: "vec1", Vector: []float32{1.0, 2.0, 3.0}, Metadata: map[string]interface{}{"type": "A"}},
		{ID: "vec2", Vector: []float32{4.0, 5.0, 6.0}, Metadata: map[string]interface{}{"type": "B"}},
		{ID: "vec3", Vector: []float32{7.0, 8.0, 9.0}, Metadata: map[string]interface{}{"type": "C"}},
	}

	for _, req := range vectors {
		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("POST", "/upsert", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		upsertHandler.Upsert(w, httpReq)
	}

	// Delete vec2
	httpReq := httptest.NewRequest("DELETE", "/vector/vec2", nil)
	httpReq = mux.SetURLVars(httpReq, map[string]string{"id": "vec2"})
	w := httptest.NewRecorder()
	deleteHandler.DeleteVectorByIDHandler(w, httpReq)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Verify only 2 vectors remain
	httpReq = httptest.NewRequest("GET", "/vectors", nil)
	w = httptest.NewRecorder()
	readHandler.GetAllVectors(w, httpReq)

	var getAllResp handlers.GetAllResponse
	_ = json.NewDecoder(w.Body).Decode(&getAllResp)

	if len(getAllResp.Vectors) != 2 {
		t.Errorf("Expected 2 vectors after deletion, got %d", len(getAllResp.Vectors))
	}

	// Verify vec2 is not in the results
	for _, vec := range getAllResp.Vectors {
		if vec.ID == "vec2" {
			t.Errorf("vec2 should have been deleted but was found in results")
		}
	}
}

func TestDeleteHandler_DeleteAndReInsert(t *testing.T) {
	upsertHandler, _, deleteHandler, cleanup := setupTestHandlers(t)
	defer cleanup()

	// Insert a vector
	req := handlers.UpsertRequest{
		ID:       "vec1",
		Vector:   []float32{1.0, 2.0, 3.0},
		Metadata: map[string]interface{}{"version": "1"},
	}
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/upsert", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	upsertHandler.Upsert(w, httpReq)

	// Delete the vector
	httpReq = httptest.NewRequest("DELETE", "/vector/vec1", nil)
	httpReq = mux.SetURLVars(httpReq, map[string]string{"id": "vec1"})
	w = httptest.NewRecorder()
	deleteHandler.DeleteVectorByIDHandler(w, httpReq)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d for delete, got %d", http.StatusOK, w.Code)
	}

	// Re-insert with same ID but different data
	req2 := handlers.UpsertRequest{
		ID:       "vec1",
		Vector:   []float32{10.0, 20.0, 30.0},
		Metadata: map[string]interface{}{"version": "2"},
	}
	body2, _ := json.Marshal(req2)
	httpReq = httptest.NewRequest("POST", "/upsert", bytes.NewBuffer(body2))
	w = httptest.NewRecorder()
	upsertHandler.Upsert(w, httpReq)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d for re-insert, got %d", http.StatusCreated, w.Code)
	}

	var resp handlers.UpsertResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)

	if resp.Status != "success" {
		t.Errorf("Expected success for re-insert, got '%s'", resp.Status)
	}
}

func TestDeleteHandler_EmptyID(t *testing.T) {
	_, _, deleteHandler, cleanup := setupTestHandlers(t)
	defer cleanup()

	// Try to delete with empty ID
	httpReq := httptest.NewRequest("DELETE", "/vector/", nil)
	httpReq = mux.SetURLVars(httpReq, map[string]string{"id": ""})
	w := httptest.NewRecorder()
	deleteHandler.DeleteVectorByIDHandler(w, httpReq)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for empty ID, got %d", http.StatusBadRequest, w.Code)
	}
}
