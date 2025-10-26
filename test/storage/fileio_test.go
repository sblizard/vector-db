package storage_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sblizard/vector-db/internal/storage"
)

func TestAppendVector(t *testing.T) {
	tmpDir := t.TempDir()
	vecFile := filepath.Join(tmpDir, "vectors.bin")

	vector := []float32{1.0, 2.0, 3.0, 4.0}

	err := storage.AppendVector(vecFile, vector)
	if err != nil {
		t.Fatalf("AppendVector failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(vecFile); os.IsNotExist(err) {
		t.Fatal("Vector file was not created")
	}

	// Load and verify
	loaded, err := storage.LoadVectors(vecFile, len(vector))
	if err != nil {
		t.Fatalf("LoadVectors failed: %v", err)
	}

	if len(loaded) != 1 {
		t.Fatalf("Expected 1 vector, got %d", len(loaded))
	}

	for i, v := range vector {
		if loaded[0][i] != v {
			t.Errorf("Vector mismatch at index %d: expected %f, got %f", i, v, loaded[0][i])
		}
	}
}

func TestWriteVectorAt(t *testing.T) {
	tmpDir := t.TempDir()
	vecFile := filepath.Join(tmpDir, "vectors.bin")

	// Write first vector
	vector1 := []float32{1.0, 2.0, 3.0}
	err := storage.AppendVector(vecFile, vector1)
	if err != nil {
		t.Fatalf("AppendVector failed: %v", err)
	}

	// Update vector at position 0
	vector2 := []float32{10.0, 20.0, 30.0}
	err = storage.WriteVectorAt(vecFile, vector2, 0)
	if err != nil {
		t.Fatalf("WriteVectorAt failed: %v", err)
	}

	// Verify update
	loaded, err := storage.ReadVectorAt(vecFile, 3, 0)
	if err != nil {
		t.Fatalf("ReadVectorAt failed: %v", err)
	}

	for i, expected := range vector2 {
		if loaded[i] != expected {
			t.Errorf("Vector mismatch at index %d: expected %f, got %f", i, expected, loaded[i])
		}
	}
}

func TestReadVectorAt(t *testing.T) {
	tmpDir := t.TempDir()
	vecFile := filepath.Join(tmpDir, "vectors.bin")

	// Write multiple vectors
	vectors := [][]float32{
		{1.0, 2.0},
		{3.0, 4.0},
		{5.0, 6.0},
	}

	for _, vec := range vectors {
		err := storage.AppendVector(vecFile, vec)
		if err != nil {
			t.Fatalf("AppendVector failed: %v", err)
		}
	}

	// Read second vector (at position 8 bytes)
	loaded, err := storage.ReadVectorAt(vecFile, 2, 8)
	if err != nil {
		t.Fatalf("ReadVectorAt failed: %v", err)
	}

	expected := vectors[1]
	for i, v := range expected {
		if loaded[i] != v {
			t.Errorf("Vector mismatch at index %d: expected %f, got %f", i, v, loaded[i])
		}
	}
}

func TestLoadVectors(t *testing.T) {
	tmpDir := t.TempDir()
	vecFile := filepath.Join(tmpDir, "vectors.bin")

	vectors := [][]float32{
		{1.0, 2.0, 3.0},
		{4.0, 5.0, 6.0},
		{7.0, 8.0, 9.0},
	}

	for _, vec := range vectors {
		err := storage.AppendVector(vecFile, vec)
		if err != nil {
			t.Fatalf("AppendVector failed: %v", err)
		}
	}

	loaded, err := storage.LoadVectors(vecFile, 3)
	if err != nil {
		t.Fatalf("LoadVectors failed: %v", err)
	}

	if len(loaded) != len(vectors) {
		t.Fatalf("Expected %d vectors, got %d", len(vectors), len(loaded))
	}

	for i, vec := range vectors {
		for j, v := range vec {
			if loaded[i][j] != v {
				t.Errorf("Vector mismatch at [%d][%d]: expected %f, got %f", i, j, v, loaded[i][j])
			}
		}
	}
}

func TestLoadVectorsMisaligned(t *testing.T) {
	tmpDir := t.TempDir()
	vecFile := filepath.Join(tmpDir, "vectors.bin")

	// Write 5 floats (not divisible by 3)
	f, _ := os.Create(vecFile)
	defer func() { _ = f.Close() }()

	for _, v := range []float32{1.0, 2.0, 3.0, 4.0, 5.0} {
		_ = storage.AppendVector(vecFile, []float32{v})
	}

	// Should fail with dimension 3
	_, err := storage.LoadVectors(vecFile, 3)
	if err == nil {
		t.Fatal("Expected error for misaligned vector file, got nil")
	}
}
