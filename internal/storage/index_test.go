package storage

import (
	"testing"
)

func TestPutAndGetIndex(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewMetaStore(tmpDir)
	defer func() { _ = store.Close() }()

	index := VectorIndex{
		Position: 1024,
		Dim:      128,
	}

	err := store.PutIndex("vec1", index)
	if err != nil {
		t.Fatalf("PutIndex failed: %v", err)
	}

	retrieved, err := store.GetIndex("vec1")
	if err != nil {
		t.Fatalf("GetIndex failed: %v", err)
	}

	if retrieved.Position != index.Position {
		t.Errorf("Position mismatch: expected %d, got %d", index.Position, retrieved.Position)
	}
	if retrieved.Dim != index.Dim {
		t.Errorf("Dim mismatch: expected %d, got %d", index.Dim, retrieved.Dim)
	}
}

func TestDeleteIndex(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewMetaStore(tmpDir)
	defer func() { _ = store.Close() }()

	index := VectorIndex{Position: 512, Dim: 64}
	id := "vec_to_delete"

	err := store.PutIndex(id, index)
	if err != nil {
		t.Fatalf("PutIndex failed: %v", err)
	}

	err = store.DeleteIndex(id)
	if err != nil {
		t.Fatalf("DeleteIndex failed: %v", err)
	}

	_, err = store.GetIndex(id)
	if err == nil {
		t.Fatal("Expected error when getting deleted index, got nil")
	}
}

func TestGetAllIndices(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewMetaStore(tmpDir)
	defer func() { _ = store.Close() }()

	indices := map[string]VectorIndex{
		"vec1": {Position: 0, Dim: 128},
		"vec2": {Position: 512, Dim: 128},
		"vec3": {Position: 1024, Dim: 256},
	}

	for id, idx := range indices {
		err := store.PutIndex(id, idx)
		if err != nil {
			t.Fatalf("PutIndex failed for %s: %v", id, err)
		}
	}

	all, err := store.GetAllIndices()
	if err != nil {
		t.Fatalf("GetAllIndices failed: %v", err)
	}

	if len(all) != len(indices) {
		t.Errorf("Expected %d indices, got %d", len(indices), len(all))
	}

	for id, expected := range indices {
		retrieved, exists := all[id]
		if !exists {
			t.Errorf("Index %s not found in GetAllIndices result", id)
			continue
		}
		if retrieved.Position != expected.Position || retrieved.Dim != expected.Dim {
			t.Errorf("Index mismatch for %s: expected %+v, got %+v", id, expected, retrieved)
		}
	}
}

func TestGetNextPosition(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewMetaStore(tmpDir)
	defer func() { _ = store.Close() }()

	// Empty store should return 0
	pos, err := store.GetNextPosition(128)
	if err != nil {
		t.Fatalf("GetNextPosition failed: %v", err)
	}
	if pos != 0 {
		t.Errorf("Expected initial position 0, got %d", pos)
	}

	// Add first vector (128 dims = 512 bytes)
	err = store.PutIndex("vec1", VectorIndex{Position: 0, Dim: 128})
	if err != nil {
		t.Fatalf("PutIndex failed: %v", err)
	}

	pos, err = store.GetNextPosition(128)
	if err != nil {
		t.Fatalf("GetNextPosition failed: %v", err)
	}
	expectedPos := int64(128 * 4) // 128 floats * 4 bytes
	if pos != expectedPos {
		t.Errorf("Expected next position %d, got %d", expectedPos, pos)
	}

	// Add second vector
	err = store.PutIndex("vec2", VectorIndex{Position: 512, Dim: 64})
	if err != nil {
		t.Fatalf("PutIndex failed: %v", err)
	}

	pos, err = store.GetNextPosition(128)
	if err != nil {
		t.Fatalf("GetNextPosition failed: %v", err)
	}
	expectedPos = int64(512 + 64*4) // position + (64 floats * 4 bytes)
	if pos != expectedPos {
		t.Errorf("Expected next position %d, got %d", expectedPos, pos)
	}
}

func TestGetIndexNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewMetaStore(tmpDir)
	defer func() { _ = store.Close() }()

	_, err := store.GetIndex("nonexistent_vector")
	if err == nil {
		t.Fatal("Expected error when getting nonexistent index, got nil")
	}
}
