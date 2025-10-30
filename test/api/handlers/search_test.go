package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sblizard/vector-db/internal/api/handlers"
	"github.com/sblizard/vector-db/internal/engine"
	"github.com/sblizard/vector-db/internal/storage"
)

func setupTestSearchHandler(t *testing.T) (*handlers.UpsertHandler, *handlers.SearchHandler, func()) {
	tmpDir := t.TempDir()
	store := storage.NewMetaStore(tmpDir)
	layout := storage.NewLayout(tmpDir)
	eng := engine.NewEngine(store, layout, 10, 3)

	upsertHandler := handlers.NewUpsertHandler(eng)
	searchHandler := handlers.NewSearchHandler(eng)

	cleanup := func() {
		_ = store.Close()
	}

	return upsertHandler, searchHandler, cleanup
}

func TestSearchHandler_EmptyDatabase(t *testing.T) {
	_, searchHandler, cleanup := setupTestSearchHandler(t)
	defer cleanup()

	req := handlers.SearchRequest{
		Vector: []float32{1.0, 0.0, 0.0},
		TopK:   5,
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/search", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	searchHandler.KClosestVectorsBruteHandler(w, httpReq)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var results []engine.SearchResult
	_ = json.NewDecoder(w.Body).Decode(&results)

	if len(results) != 0 {
		t.Errorf("Expected 0 results from empty database, got %d", len(results))
	}
}

func TestSearchHandler_TopKResults(t *testing.T) {
	upsertHandler, searchHandler, cleanup := setupTestSearchHandler(t)
	defer cleanup()

	// Insert test vectors
	vectors := []handlers.UpsertRequest{
		{ID: "vec1", Vector: []float32{1.0, 0.0, 0.0}, Metadata: map[string]interface{}{"label": "x-axis"}},
		{ID: "vec2", Vector: []float32{0.9, 0.1, 0.0}, Metadata: map[string]interface{}{"label": "near-x"}},
		{ID: "vec3", Vector: []float32{0.0, 1.0, 0.0}, Metadata: map[string]interface{}{"label": "y-axis"}},
		{ID: "vec4", Vector: []float32{0.0, 0.0, 1.0}, Metadata: map[string]interface{}{"label": "z-axis"}},
		{ID: "vec5", Vector: []float32{-1.0, 0.0, 0.0}, Metadata: map[string]interface{}{"label": "neg-x"}},
	}

	for _, vec := range vectors {
		body, _ := json.Marshal(vec)
		httpReq := httptest.NewRequest("POST", "/upsert", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		upsertHandler.Upsert(w, httpReq)
	}

	// Search for top 3 similar to [1, 0, 0]
	searchReq := handlers.SearchRequest{
		Vector: []float32{1.0, 0.0, 0.0},
		TopK:   3,
	}

	body, _ := json.Marshal(searchReq)
	httpReq := httptest.NewRequest("POST", "/search", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	searchHandler.KClosestVectorsBruteHandler(w, httpReq)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var results []engine.SearchResult
	_ = json.NewDecoder(w.Body).Decode(&results)

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	// First result should be vec1 (perfect match)
	if results[0].ID != "vec1" {
		t.Errorf("Expected first result to be 'vec1', got '%s'", results[0].ID)
	}

	// Verify results are ordered by score (descending)
	for i := 0; i < len(results)-1; i++ {
		if results[i].Score < results[i+1].Score {
			t.Errorf("Results not properly ordered: result[%d].Score=%f < result[%d].Score=%f",
				i, results[i].Score, i+1, results[i+1].Score)
		}
	}
}

func TestSearchHandler_MissingVector(t *testing.T) {
	_, searchHandler, cleanup := setupTestSearchHandler(t)
	defer cleanup()

	req := handlers.SearchRequest{
		Vector: []float32{},
		TopK:   5,
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/search", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	searchHandler.KClosestVectorsBruteHandler(w, httpReq)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for missing vector, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestSearchHandler_InvalidJSON(t *testing.T) {
	_, searchHandler, cleanup := setupTestSearchHandler(t)
	defer cleanup()

	httpReq := httptest.NewRequest("POST", "/search", bytes.NewBufferString("invalid json"))
	w := httptest.NewRecorder()

	searchHandler.KClosestVectorsBruteHandler(w, httpReq)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for invalid JSON, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestSearchHandler_TopKLargerThanDatabase(t *testing.T) {
	upsertHandler, searchHandler, cleanup := setupTestSearchHandler(t)
	defer cleanup()

	// Insert only 2 vectors
	vectors := []handlers.UpsertRequest{
		{ID: "vec1", Vector: []float32{1.0, 0.0, 0.0}, Metadata: map[string]interface{}{}},
		{ID: "vec2", Vector: []float32{0.0, 1.0, 0.0}, Metadata: map[string]interface{}{}},
	}

	for _, vec := range vectors {
		body, _ := json.Marshal(vec)
		httpReq := httptest.NewRequest("POST", "/upsert", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		upsertHandler.Upsert(w, httpReq)
	}

	// Request top 10 when only 2 exist
	searchReq := handlers.SearchRequest{
		Vector: []float32{1.0, 0.0, 0.0},
		TopK:   10,
	}

	body, _ := json.Marshal(searchReq)
	httpReq := httptest.NewRequest("POST", "/search", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	searchHandler.KClosestVectorsBruteHandler(w, httpReq)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var results []engine.SearchResult
	_ = json.NewDecoder(w.Body).Decode(&results)

	// Should return all available vectors (2)
	if len(results) != 2 {
		t.Errorf("Expected 2 results (all available), got %d", len(results))
	}
}

func TestSearchHandler_ScoreAccuracy(t *testing.T) {
	upsertHandler, searchHandler, cleanup := setupTestSearchHandler(t)
	defer cleanup()

	// Insert orthogonal vectors
	vectors := []handlers.UpsertRequest{
		{ID: "vec1", Vector: []float32{1.0, 0.0, 0.0}, Metadata: map[string]interface{}{}},
		{ID: "vec2", Vector: []float32{0.0, 1.0, 0.0}, Metadata: map[string]interface{}{}},
	}

	for _, vec := range vectors {
		body, _ := json.Marshal(vec)
		httpReq := httptest.NewRequest("POST", "/upsert", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		upsertHandler.Upsert(w, httpReq)
	}

	// Search with vec1
	searchReq := handlers.SearchRequest{
		Vector: []float32{1.0, 0.0, 0.0},
		TopK:   2,
	}

	body, _ := json.Marshal(searchReq)
	httpReq := httptest.NewRequest("POST", "/search", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	searchHandler.KClosestVectorsBruteHandler(w, httpReq)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var results []engine.SearchResult
	_ = json.NewDecoder(w.Body).Decode(&results)

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	// First should be vec1 with score ~1.0
	if results[0].ID != "vec1" {
		t.Errorf("Expected first result to be 'vec1', got '%s'", results[0].ID)
	}
	if results[0].Score < 0.99 {
		t.Errorf("Expected score ~1.0 for identical vector, got %f", results[0].Score)
	}

	// Second should be vec2 with score ~0 (orthogonal)
	if results[1].ID != "vec2" {
		t.Errorf("Expected second result to be 'vec2', got '%s'", results[1].ID)
	}
	if results[1].Score > 0.01 && results[1].Score < -0.01 {
		t.Errorf("Expected score ~0 for orthogonal vector, got %f", results[1].Score)
	}
}
