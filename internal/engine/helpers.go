package engine

func extractOriginalVector(metadata map[string]interface{}) []float32 {
	origVec, exists := metadata["original_vector"]
	if !exists {
		return nil
	}

	vecArray, ok := origVec.([]interface{})
	if !ok {
		return nil
	}

	originalVector := make([]float32, len(vecArray))
	for i, v := range vecArray {
		if floatVal, ok := v.(float64); ok {
			originalVector[i] = float32(floatVal)
		} else {
			return nil
		}
	}

	return originalVector
}
