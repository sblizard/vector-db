package storage

import (
	"testing"
)

func TestNewLayout(t *testing.T) {
	baseDir := "/test/data"
	layout := NewLayout(baseDir)

	if layout.BasePath != baseDir {
		t.Errorf("Expected BasePath %s, got %s", baseDir, layout.BasePath)
	}
}

func TestVectorFile(t *testing.T) {
	baseDir := "/test/data"
	layout := NewLayout(baseDir)

	tests := []struct {
		partition int
		expected  string
	}{
		{0, "/test/data/cluster_0.bin"},
		{1, "/test/data/cluster_1.bin"},
		{10, "/test/data/cluster_10.bin"},
	}

	for _, tt := range tests {
		result := layout.VectorFile(tt.partition)
		if result != tt.expected {
			t.Errorf("VectorFile(%d): expected %s, got %s", tt.partition, tt.expected, result)
		}
	}
}

func TestMetaDB(t *testing.T) {
	baseDir := "/test/data"
	layout := NewLayout(baseDir)

	expected := "/test/data/metadata.db"
	result := layout.MetaDB()

	if result != expected {
		t.Errorf("Expected MetaDB %s, got %s", expected, result)
	}
}
