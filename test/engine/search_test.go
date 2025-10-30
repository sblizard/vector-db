package engine_test

import (
	"testing"

	"github.com/sblizard/vector-db/internal/engine"
	"github.com/sblizard/vector-db/internal/storage"
)

func setupTestEngine(t *testing.T) (*engine.Engine, func()) {
	tmpDir := t.TempDir()
	store := storage.NewMetaStore(tmpDir)
	layout := storage.NewLayout(tmpDir)
	eng := engine.NewEngine(store, layout, 10, 3)

	cleanup := func() {
		_ = store.Close()
	}

	return eng, cleanup
}

func TestKClosestVectorsBrute_EmptyDatabase(t *testing.T) {
	eng, cleanup := setupTestEngine(t)
	defer cleanup()

	query := []float32{1.0, 0.0, 0.0}
	k := 5

	results, err := eng.KClosestVectorsBrute(query, k)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results from empty database, got %d", len(results))
	}
}

func TestKClosestVectorsBrute_SingleVector(t *testing.T) {
	eng, cleanup := setupTestEngine(t)
	defer cleanup()

	// Insert a vector
	_, err := eng.Upsert("vec1", []float32{1.0, 0.0, 0.0}, map[string]interface{}{"label": "x-axis"})
	if err != nil {
		t.Fatalf("Failed to upsert vector: %v", err)
	}

	// Query with same vector
	query := []float32{1.0, 0.0, 0.0}
	k := 1

	results, err := eng.KClosestVectorsBrute(query, k)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if results[0].ID != "vec1" {
		t.Errorf("Expected ID 'vec1', got '%s'", results[0].ID)
	}

	// Score should be close to 1.0 (perfect similarity)
	if results[0].Score < 0.99 {
		t.Errorf("Expected score close to 1.0, got %f", results[0].Score)
	}
}

func TestKClosestVectorsBrute_TopKOrdering(t *testing.T) {
	eng, cleanup := setupTestEngine(t)
	defer cleanup()

	// Insert vectors with different similarities to query [1, 0, 0]
	vectors := map[string][]float32{
		"vec1": {1.0, 0.0, 0.0},  // Perfect match - score ~1.0
		"vec2": {0.9, 0.1, 0.0},  // High similarity
		"vec3": {0.5, 0.5, 0.0},  // Medium similarity
		"vec4": {0.0, 1.0, 0.0},  // Low similarity (orthogonal)
		"vec5": {-1.0, 0.0, 0.0}, // Opposite direction - negative score
	}

	for id, vec := range vectors {
		_, err := eng.Upsert(id, vec, map[string]interface{}{})
		if err != nil {
			t.Fatalf("Failed to upsert %s: %v", id, err)
		}
	}

	query := []float32{1.0, 0.0, 0.0}
	k := 3

	results, err := eng.KClosestVectorsBrute(query, k)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != k {
		t.Fatalf("Expected %d results, got %d", k, len(results))
	}

	// Check ordering - scores should be descending
	for i := 0; i < len(results)-1; i++ {
		if results[i].Score < results[i+1].Score {
			t.Errorf("Results not properly ordered: result[%d].Score=%f < result[%d].Score=%f",
				i, results[i].Score, i+1, results[i+1].Score)
		}
	}

	// First result should be vec1 (perfect match)
	if results[0].ID != "vec1" {
		t.Errorf("Expected first result to be 'vec1', got '%s'", results[0].ID)
	}

	// Check that we got the top 3
	expectedIDs := map[string]bool{"vec1": true, "vec2": true, "vec3": true}
	for _, result := range results {
		if !expectedIDs[result.ID] {
			t.Errorf("Unexpected ID '%s' in top %d results", result.ID, k)
		}
	}
}

func TestKClosestVectorsBrute_KLargerThanDatabase(t *testing.T) {
	eng, cleanup := setupTestEngine(t)
	defer cleanup()

	// Insert 3 vectors
	for i := 0; i < 3; i++ {
		vec := []float32{float32(i), 1.0, 0.0}
		_, err := eng.Upsert(string(rune('a'+i)), vec, map[string]interface{}{})
		if err != nil {
			t.Fatalf("Failed to upsert vector: %v", err)
		}
	}

	query := []float32{1.0, 1.0, 0.0}
	k := 10 // Ask for more than available

	results, err := eng.KClosestVectorsBrute(query, k)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should return all 3 vectors
	if len(results) != 3 {
		t.Errorf("Expected 3 results (all available), got %d", len(results))
	}
}

func TestKClosestVectorsBrute_KZero(t *testing.T) {
	eng, cleanup := setupTestEngine(t)
	defer cleanup()

	// Insert a vector
	_, err := eng.Upsert("vec1", []float32{1.0, 0.0, 0.0}, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to upsert vector: %v", err)
	}

	query := []float32{1.0, 0.0, 0.0}
	k := 0

	results, err := eng.KClosestVectorsBrute(query, k)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results for k=0, got %d", len(results))
	}
}

func TestKClosestVectorsBrute_DifferentDimensions(t *testing.T) {
	eng, cleanup := setupTestEngine(t)
	defer cleanup()

	// Insert 2D vectors
	_, err := eng.Upsert("vec1", []float32{1.0, 0.0}, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to upsert vector: %v", err)
	}

	_, err = eng.Upsert("vec2", []float32{0.0, 1.0}, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to upsert vector: %v", err)
	}

	// Query with 2D vector
	query := []float32{1.0, 0.0}
	k := 2

	results, err := eng.KClosestVectorsBrute(query, k)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// vec1 should be first (perfect match)
	if results[0].ID != "vec1" {
		t.Errorf("Expected first result to be 'vec1', got '%s'", results[0].ID)
	}
}

func TestKClosestVectorsBrute_HighDimensional(t *testing.T) {
	eng, cleanup := setupTestEngine(t)
	defer cleanup()

	// Insert high-dimensional vectors
	dim := 128
	vec1 := make([]float32, dim)
	vec2 := make([]float32, dim)
	vec3 := make([]float32, dim)

	// vec1: all ones
	for i := 0; i < dim; i++ {
		vec1[i] = 1.0
	}

	// vec2: similar to vec1
	for i := 0; i < dim; i++ {
		vec2[i] = 0.9
	}

	// vec3: different pattern
	for i := 0; i < dim; i++ {
		if i%2 == 0 {
			vec3[i] = 1.0
		} else {
			vec3[i] = -1.0
		}
	}

	_, err := eng.Upsert("vec1", vec1, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to upsert vec1: %v", err)
	}

	_, err = eng.Upsert("vec2", vec2, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to upsert vec2: %v", err)
	}

	_, err = eng.Upsert("vec3", vec3, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to upsert vec3: %v", err)
	}

	// Query with vec1
	k := 2
	results, err := eng.KClosestVectorsBrute(vec1, k)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != k {
		t.Fatalf("Expected %d results, got %d", k, len(results))
	}

	// First should be vec1 (perfect match)
	if results[0].ID != "vec1" {
		t.Errorf("Expected first result to be 'vec1', got '%s'", results[0].ID)
	}

	// Second should be vec2 (more similar than vec3)
	if results[1].ID != "vec2" {
		t.Errorf("Expected second result to be 'vec2', got '%s'", results[1].ID)
	}
}

func TestKClosestVectorsBrute_AllIdenticalScores(t *testing.T) {
	eng, cleanup := setupTestEngine(t)
	defer cleanup()

	// Insert vectors that are all equidistant from the query
	// Using orthogonal vectors
	vectors := map[string][]float32{
		"vec1": {1.0, 0.0, 0.0},
		"vec2": {0.0, 1.0, 0.0},
		"vec3": {0.0, 0.0, 1.0},
	}

	for id, vec := range vectors {
		_, err := eng.Upsert(id, vec, map[string]interface{}{})
		if err != nil {
			t.Fatalf("Failed to upsert %s: %v", id, err)
		}
	}

	// Query with a vector that makes all stored vectors equidistant
	query := []float32{1.0, 1.0, 1.0}
	k := 2

	results, err := eng.KClosestVectorsBrute(query, k)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != k {
		t.Errorf("Expected %d results, got %d", k, len(results))
	}

	// All scores should be similar (around 0.577 for 3D orthogonal vectors)
	for i := 0; i < len(results)-1; i++ {
		scoreDiff := results[i].Score - results[i+1].Score
		if scoreDiff < -0.01 || scoreDiff > 0.01 {
			t.Errorf("Scores are not similar: %f vs %f", results[i].Score, results[i+1].Score)
		}
	}
}

func TestKClosestVectorsBrute_ManyVectors(t *testing.T) {
	eng, cleanup := setupTestEngine(t)
	defer cleanup()

	// Insert many vectors
	numVectors := 100
	for i := 0; i < numVectors; i++ {
		vec := []float32{float32(i) / 100.0, float32(100-i) / 100.0, 0.5}
		_, err := eng.Upsert(string(rune(i)), vec, map[string]interface{}{"index": i})
		if err != nil {
			t.Fatalf("Failed to upsert vector %d: %v", i, err)
		}
	}

	query := []float32{0.5, 0.5, 0.5}
	k := 10

	results, err := eng.KClosestVectorsBrute(query, k)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != k {
		t.Errorf("Expected %d results, got %d", k, len(results))
	}

	// Verify ordering
	for i := 0; i < len(results)-1; i++ {
		if results[i].Score < results[i+1].Score {
			t.Errorf("Results not properly ordered at index %d", i)
		}
	}
}
