package storage

import (
	"encoding/binary"
	"fmt"
	"os"
)

func AppendVector(path string, vector []float32) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open vector file: %w", err)
	}
	defer f.Close()

	for _, val := range vector {
		if err := binary.Write(f, binary.LittleEndian, val); err != nil {
			return fmt.Errorf("failed to write vector data: %w", err)
		}
	}
	return nil
}

func WriteVectorAt(path string, vector []float32, position int64) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open vector file: %w", err)
	}
	defer f.Close()

	if _, err := f.Seek(position, 0); err != nil {
		return fmt.Errorf("failed to seek to position: %w", err)
	}

	for _, val := range vector {
		if err := binary.Write(f, binary.LittleEndian, val); err != nil {
			return fmt.Errorf("failed to write vector data: %w", err)
		}
	}
	return nil
}

func ReadVectorAt(path string, dim int, position int64) ([]float32, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open vector file: %w", err)
	}
	defer f.Close()

	if _, err := f.Seek(position, 0); err != nil {
		return nil, fmt.Errorf("failed to seek to position: %w", err)
	}

	vector := make([]float32, dim)
	if err := binary.Read(f, binary.LittleEndian, &vector); err != nil {
		return nil, fmt.Errorf("failed to read vector data: %w", err)
	}

	return vector, nil
}

func LoadVectors(path string, dim int) ([][]float32, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open vector file: %w", err)
	}
	defer f.Close()

	stat, _ := f.Stat()
	totalFloats := int(stat.Size()) / 4
	if totalFloats%dim != 0 {
		return nil, fmt.Errorf("vector file size is not aligned with dimension")
	}

	numVectors := totalFloats / dim
	vectors := make([][]float32, numVectors)

	for i := 0; i < numVectors; i++ {
		vector := make([]float32, dim)
		for j := 0; j < dim; j++ {
			if err := binary.Read(f, binary.LittleEndian, &vector[j]); err != nil {
				return nil, fmt.Errorf("failed to read vector data: %w", err)
			}
		}
		vectors[i] = vector
	}
	return vectors, nil
}
