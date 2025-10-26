package storage_test

import (
	"testing"

	"github.com/sblizard/vector-db/internal/storage"
)

func TestMetaStorePutGet(t *testing.T) {
	tmpDir := t.TempDir()
	store := storage.NewMetaStore(tmpDir)
	defer func() { _ = store.Close() }()

	key := "test_key"
	value := []byte("test_value")

	// Put
	err := store.Put(key, value)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Get
	retrieved, err := store.Get(key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(retrieved) != string(value) {
		t.Errorf("Expected %s, got %s", string(value), string(retrieved))
	}
}

func TestMetaStoreDelete(t *testing.T) {
	tmpDir := t.TempDir()
	store := storage.NewMetaStore(tmpDir)
	defer func() { _ = store.Close() }()

	key := "test_key"
	value := []byte("test_value")

	// Put
	err := store.Put(key, value)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Delete
	err = store.Delete(key)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Try to get deleted key
	_, err = store.Get(key)
	if err == nil {
		t.Fatal("Expected error when getting deleted key, got nil")
	}
}

func TestMetaStoreGetAll(t *testing.T) {
	tmpDir := t.TempDir()
	store := storage.NewMetaStore(tmpDir)
	defer func() { _ = store.Close() }()

	entries := map[string][]byte{
		"key1": []byte("value1"),
		"key2": []byte("value2"),
		"key3": []byte("value3"),
	}

	// Put all entries
	for k, v := range entries {
		err := store.Put(k, v)
		if err != nil {
			t.Fatalf("Put failed for %s: %v", k, err)
		}
	}

	// Get all
	all, err := store.GetAll()
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}

	// Verify count
	if len(all) < len(entries) {
		t.Errorf("Expected at least %d entries, got %d", len(entries), len(all))
	}

	// Verify values
	for k, expectedValue := range entries {
		retrievedValue, exists := all[k]
		if !exists {
			t.Errorf("Key %s not found in GetAll result", k)
			continue
		}
		if string(retrievedValue) != string(expectedValue) {
			t.Errorf("Value mismatch for key %s: expected %s, got %s", k, string(expectedValue), string(retrievedValue))
		}
	}
}

func TestMetaStoreGetNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	store := storage.NewMetaStore(tmpDir)
	defer func() { _ = store.Close() }()

	_, err := store.Get("nonexistent_key")
	if err == nil {
		t.Fatal("Expected error when getting nonexistent key, got nil")
	}
}
