package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sblizard/vector-db/internal/api"
	"github.com/sblizard/vector-db/internal/api/handlers"
	"github.com/sblizard/vector-db/internal/storage"
)

func setupTestServer(t *testing.T) (*httptest.Server, func()) {
	tmpDir := t.TempDir()
	store := storage.NewMetaStore(tmpDir)
	layout := storage.NewLayout(tmpDir)

	healthHandler := api.NewHandler(store, layout)
	upsertHandler := handlers.NewUpsertHandler(store, layout)
	readHandler := handlers.NewReadHandler(store, layout)

	router := handlers.NewRouter(healthHandler, upsertHandler, readHandler)
	server := httptest.NewServer(router)

	cleanup := func() {
		server.Close()
		_ = store.Close()
	}

	return server, cleanup
}

func TestIntegration_HealthEndpoint(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	resp, err := http.Get(server.URL + "/health")
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestIntegration_UpsertAndRetrieve(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// Test upsert
	upsertReq := handlers.UpsertRequest{
		ID:       "test_vec",
		Vector:   []float32{1.0, 2.0, 3.0, 4.0},
		Metadata: map[string]string{"category": "test", "label": "sample"},
	}

	body, _ := json.Marshal(upsertReq)
	resp, err := http.Post(server.URL+"/upsert", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Upsert request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status %d for insert, got %d", http.StatusCreated, resp.StatusCode)
	}

	// Test retrieve
	resp, err = http.Get(server.URL + "/vectors")
	if err != nil {
		t.Fatalf("Get vectors request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var getAllResp handlers.GetAllResponse
	_ = json.NewDecoder(resp.Body).Decode(&getAllResp)

	if len(getAllResp.Vectors) != 1 {
		t.Errorf("Expected 1 vector, got %d", len(getAllResp.Vectors))
	}

	vec := getAllResp.Vectors[0]
	if vec.ID != "test_vec" {
		t.Errorf("Expected ID 'test_vec', got '%s'", vec.ID)
	}

	if len(vec.Vector) != 4 {
		t.Errorf("Expected vector length 4, got %d", len(vec.Vector))
	}

	if vec.Metadata["category"] != "test" {
		t.Errorf("Expected metadata category 'test', got '%s'", vec.Metadata["category"])
	}
}

func TestIntegration_MultipleVectors(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// Insert multiple vectors
	vectors := []handlers.UpsertRequest{
		{ID: "vec1", Vector: []float32{1.0, 2.0, 3.0}, Metadata: map[string]string{"label": "a"}},
		{ID: "vec2", Vector: []float32{4.0, 5.0, 6.0}, Metadata: map[string]string{"label": "b"}},
		{ID: "vec3", Vector: []float32{7.0, 8.0, 9.0}, Metadata: map[string]string{"label": "c"}},
	}

	for _, vec := range vectors {
		body, _ := json.Marshal(vec)
		resp, err := http.Post(server.URL+"/upsert", "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("Failed to insert vector %s: %v", vec.ID, err)
		}
		_ = resp.Body.Close()
	}

	// Retrieve all
	resp, err := http.Get(server.URL + "/vectors")
	if err != nil {
		t.Fatalf("Failed to get vectors: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var getAllResp handlers.GetAllResponse
	_ = json.NewDecoder(resp.Body).Decode(&getAllResp)

	if len(getAllResp.Vectors) != 3 {
		t.Errorf("Expected 3 vectors, got %d", len(getAllResp.Vectors))
	}
}

func TestIntegration_UpdateVector(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// Initial insert
	upsertReq := handlers.UpsertRequest{
		ID:       "update_test",
		Vector:   []float32{1.0, 2.0, 3.0},
		Metadata: map[string]string{"version": "1"},
	}
	body, _ := json.Marshal(upsertReq)
	resp, _ := http.Post(server.URL+"/upsert", "application/json", bytes.NewBuffer(body))
	_ = resp.Body.Close()

	// Update
	updateReq := handlers.UpsertRequest{
		ID:       "update_test",
		Vector:   []float32{10.0, 20.0, 30.0},
		Metadata: map[string]string{"version": "2"},
	}
	body, _ = json.Marshal(updateReq)
	resp, err := http.Post(server.URL+"/upsert", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Update request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d for update, got %d", http.StatusOK, resp.StatusCode)
	}

	// Verify update
	resp, err = http.Get(server.URL + "/vectors")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var getAllResp handlers.GetAllResponse
	_ = json.NewDecoder(resp.Body).Decode(&getAllResp)

	if len(getAllResp.Vectors) != 1 {
		t.Errorf("Expected 1 vector after update, got %d", len(getAllResp.Vectors))
	}

	vec := getAllResp.Vectors[0]
	if vec.Vector[0] != 10.0 {
		t.Errorf("Expected updated vector[0] = 10.0, got %f", vec.Vector[0])
	}

	if vec.Metadata["version"] != "2" {
		t.Errorf("Expected updated version '2', got '%s'", vec.Metadata["version"])
	}
}

func TestIntegration_ErrorHandling(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	tests := []struct {
		name           string
		request        handlers.UpsertRequest
		expectedStatus int
	}{
		{
			name:           "missing ID",
			request:        handlers.UpsertRequest{ID: "", Vector: []float32{1.0}},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty vector",
			request:        handlers.UpsertRequest{ID: "test", Vector: []float32{}},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			resp, err := http.Post(server.URL+"/upsert", "application/json", bytes.NewBuffer(body))
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestIntegration_EmptyDatabase(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	resp, err := http.Get(server.URL + "/vectors")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status %d for empty database, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func BenchmarkUpsert(b *testing.B) {
	tmpDir := b.TempDir()
	store := storage.NewMetaStore(tmpDir)
	layout := storage.NewLayout(tmpDir)
	defer func() { _ = store.Close() }()

	healthHandler := api.NewHandler(store, layout)
	upsertHandler := handlers.NewUpsertHandler(store, layout)
	readHandler := handlers.NewReadHandler(store, layout)

	router := handlers.NewRouter(healthHandler, upsertHandler, readHandler)
	server := httptest.NewServer(router)
	defer server.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		upsertReq := handlers.UpsertRequest{
			ID:       fmt.Sprintf("vec_%d", i),
			Vector:   []float32{float32(i), float32(i + 1), float32(i + 2)},
			Metadata: map[string]string{"index": fmt.Sprintf("%d", i)},
		}
		body, _ := json.Marshal(upsertReq)
		resp, _ := http.Post(server.URL+"/upsert", "application/json", bytes.NewBuffer(body))
		_ = resp.Body.Close()
	}
}
