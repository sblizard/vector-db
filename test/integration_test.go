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
		Metadata: map[string]interface{}{"category": "test", "label": "sample"},
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
		{ID: "vec1", Vector: []float32{1.0, 2.0, 3.0}, Metadata: map[string]interface{}{"label": "a"}},
		{ID: "vec2", Vector: []float32{4.0, 5.0, 6.0}, Metadata: map[string]interface{}{"label": "b"}},
		{ID: "vec3", Vector: []float32{7.0, 8.0, 9.0}, Metadata: map[string]interface{}{"label": "c"}},
	}

	for _, vec := range vectors {
		body, _ := json.Marshal(vec)
		resp, err := http.Post(server.URL+"/upsert", "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("Failed to insert vector %s: %v", vec.ID, err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status %d for vector %s, got %d", http.StatusCreated, vec.ID, resp.StatusCode)
		}
		_ = resp.Body.Close()
	}

	// Retrieve all
	resp, err := http.Get(server.URL + "/vectors")
	if err != nil {
		t.Fatalf("Failed to get vectors: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var getAllResp handlers.GetAllResponse
	if err := json.NewDecoder(resp.Body).Decode(&getAllResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(getAllResp.Vectors) != 3 {
		t.Errorf("Expected 3 vectors, got %d", len(getAllResp.Vectors))
		for i, v := range getAllResp.Vectors {
			t.Logf("Vector %d: ID=%s", i, v.ID)
		}
	}
}

func TestIntegration_UpdateVector(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// Initial insert
	upsertReq := handlers.UpsertRequest{
		ID:       "update_test",
		Vector:   []float32{1.0, 2.0, 3.0},
		Metadata: map[string]interface{}{"version": "1"},
	}
	body, _ := json.Marshal(upsertReq)
	resp, _ := http.Post(server.URL+"/upsert", "application/json", bytes.NewBuffer(body))
	_ = resp.Body.Close()

	// Update
	updateReq := handlers.UpsertRequest{
		ID:       "update_test",
		Vector:   []float32{10.0, 20.0, 30.0},
		Metadata: map[string]interface{}{"version": "2"},
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
	if vec.OriginalVector[0] != 10.0 {
		t.Errorf("Expected updated original_vector[0] = 10.0, got %f", vec.OriginalVector[0])
	}

	// Verify normalized vector is actually normalized (unit length)
	normSum := float32(0)
	for _, v := range vec.Vector {
		normSum += v * v
	}
	normLength := float32(0.999) // Allow small floating point error
	if normSum < normLength || normSum > 1.001 {
		t.Errorf("Expected normalized vector to have unit length, got %f", normSum)
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
			expectedStatus: http.StatusCreated,
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

func TestIntegration_VectorNormalization(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// Insert a vector that's not normalized
	upsertReq := handlers.UpsertRequest{
		ID:       "norm_test",
		Vector:   []float32{3.0, 4.0}, // Length = 5.0
		Metadata: map[string]interface{}{"test": "normalization"},
	}
	body, _ := json.Marshal(upsertReq)
	resp, err := http.Post(server.URL+"/upsert", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Upsert request failed: %v", err)
	}
	_ = resp.Body.Close()

	// Retrieve and verify
	resp, err = http.Get(server.URL + "/vectors")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var getAllResp handlers.GetAllResponse
	_ = json.NewDecoder(resp.Body).Decode(&getAllResp)

	if len(getAllResp.Vectors) != 1 {
		t.Fatalf("Expected 1 vector, got %d", len(getAllResp.Vectors))
	}

	vec := getAllResp.Vectors[0]

	// Verify original vector is preserved
	if vec.OriginalVector[0] != 3.0 || vec.OriginalVector[1] != 4.0 {
		t.Errorf("Original vector not preserved: expected [3.0, 4.0], got [%f, %f]",
			vec.OriginalVector[0], vec.OriginalVector[1])
	}

	// Verify normalized vector has unit length (sqrt(x^2 + y^2) = 1)
	var sumSquares float32
	for _, v := range vec.Vector {
		sumSquares += v * v
	}

	// Allow small floating point error
	if sumSquares < 0.999 || sumSquares > 1.001 {
		t.Errorf("Expected normalized vector to have unit length, got length^2 = %f", sumSquares)
	}

	// Verify normalized values are correct: [3/5, 4/5] = [0.6, 0.8]
	expectedNorm := []float32{0.6, 0.8}
	tolerance := float32(0.0001)
	for i, expected := range expectedNorm {
		diff := vec.Vector[i] - expected
		if diff < -tolerance || diff > tolerance {
			t.Errorf("Normalized vector[%d]: expected %f, got %f", i, expected, vec.Vector[i])
		}
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
			Metadata: map[string]interface{}{"index": fmt.Sprintf("%d", i)},
		}
		body, _ := json.Marshal(upsertReq)
		resp, _ := http.Post(server.URL+"/upsert", "application/json", bytes.NewBuffer(body))
		_ = resp.Body.Close()
	}
}
