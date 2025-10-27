package engine

import (
	"encoding/json"
	"fmt"

	"github.com/sblizard/vector-db/internal/math"
	"github.com/sblizard/vector-db/internal/storage"
)

// DeleteVectorByID deletes a vector and its metadata by ID.
// Returns true if the vector existed and was deleted, false if it didn't exist.
func (e *Engine) DeleteVectorByID(id string) (bool, error) {
	_, err := e.storage.GetIndex(id)
	if err != nil {
		return false, nil
	}

	err = e.storage.DeleteIndex(id)
	if err != nil {
		return false, err
	}

	err = e.storage.Delete(id)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (e *Engine) GetVectorMetadata(id string) (map[string]interface{}, error) {
	metaEntries, err := e.storage.GetAll()
	if err != nil {
		return nil, err
	}

	var metadata map[string]interface{}

	if metaBytes, ok := metaEntries[id]; ok {
		_ = json.Unmarshal(metaBytes, &metadata)
	}

	userMetadata := make(map[string]interface{})
	for k, v := range metadata {
		userMetadata[k] = v
	}

	return userMetadata, nil
}

func (e *Engine) GetAllVectors() (GetAllResponse, error) {
	metaEntries, err := e.storage.GetAll()
	if err != nil {
		fmt.Printf("Error retrieving metadata: %v\n", err)
		return GetAllResponse{}, err
	}

	indices, err := e.storage.GetAllIndices()
	if err != nil {
		fmt.Printf("Error retrieving indices: %v\n", err)
		return GetAllResponse{}, err
	}

	if len(indices) == 0 {
		return GetAllResponse{Vectors: []StoredVector{}}, nil
	}

	vectors := make([]StoredVector, 0, len(indices))

	for id, index := range indices {
		normalizedVector, err := storage.ReadVectorAt(e.layout.VectorFile(0), index.Dim, index.Position)
		if err != nil {
			fmt.Printf("Error reading vector at index: %v\n", err)
			return GetAllResponse{}, err
		}

		var metadata map[string]interface{}
		var originalVector []float32

		if metaBytes, ok := metaEntries[id]; ok {
			_ = json.Unmarshal(metaBytes, &metadata)
			originalVector = extractOriginalVector(metadata)
		}

		userMetadata := make(map[string]interface{})
		for k, v := range metadata {
			if k != "original_vector" {
				userMetadata[k] = v
			}
		}

		vectors = append(vectors, StoredVector{
			ID:             id,
			Vector:         normalizedVector,
			OriginalVector: originalVector,
			Metadata:       userMetadata,
		})
	}

	fmt.Printf("Returned %d vectors with metadata\n", len(vectors))
	return GetAllResponse{Vectors: vectors}, nil
}

func (e *Engine) Upsert(id string, vector []float32, metadata map[string]interface{}) (bool, error) {
	vecPath := e.layout.VectorFile(0)

	existingIndex, err := e.storage.GetIndex(id)
	isUpdate := err == nil

	normalizedVector := math.Normalize(vector)

	if isUpdate {
		if len(vector) != existingIndex.Dim {
			return isUpdate, fmt.Errorf("vector dimension mismatch: expected %d, got %d", existingIndex.Dim, len(vector))
		}

		if err := storage.WriteVectorAt(vecPath, normalizedVector, existingIndex.Position); err != nil {
			return isUpdate, fmt.Errorf("failed to update vector: %v", err)
		}
	} else {
		position, err := e.storage.GetNextPosition(len(vector))
		if err != nil {
			return isUpdate, fmt.Errorf("failed to get next position: %v", err)
		}

		if err := storage.AppendVector(vecPath, normalizedVector); err != nil {
			return isUpdate, fmt.Errorf("failed to store vector: %v", err)
		}

		index := storage.VectorIndex{
			Position: position,
			Dim:      len(vector),
		}
		if err := e.storage.PutIndex(id, index); err != nil {
			return isUpdate, fmt.Errorf("failed to store index: %v", err)
		}
	}

	metadataWithVector := make(map[string]interface{})
	for k, v := range metadata {
		metadataWithVector[k] = v
	}
	metadataWithVector["original_vector"] = vector

	metadataBytes, err := json.Marshal(metadataWithVector)
	if err != nil {
		return isUpdate, fmt.Errorf("failed to serialize metadata: %v", err)
	}

	if err := e.storage.Put(id, metadataBytes); err != nil {
		return isUpdate, fmt.Errorf("failed to store metadata: %v", err)
	}

	fmt.Printf("Successfully upserted vector: ID=%s, Dim=%d, IsUpdate=%v\n", id, len(vector), isUpdate)
	return isUpdate, nil
}

func (e *Engine) DeleteAllVectors() error {
	vecPath := e.layout.VectorFile(0)
	err := storage.DeleteAllVectors(vecPath)
	if err != nil {
		return fmt.Errorf("failed to delete all vectors: %v", err)
	}

	err = e.storage.DeleteAllIndices()
	if err != nil {
		return fmt.Errorf("failed to delete all indices: %v", err)
	}

	err = e.storage.DeleteAllMetadata()
	if err != nil {
		return fmt.Errorf("failed to delete all metadata: %v", err)
	}

	return nil
}
