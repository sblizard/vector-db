package engine

import (
	"fmt"
	"sort"

	"github.com/sblizard/vector-db/internal/math"
)

func (e *Engine) KClosestVectorsBrute(query []float32, k int) ([]SearchResult, error) {
	allVectors, err := e.GetAllVectors()
	if err != nil {
		return nil, err
	}

	results := make([]SearchResult, 0, len(allVectors.Vectors))
	for _, storedVector := range allVectors.Vectors {
		score, err := math.SafeCosineSimilarity(query, storedVector.Vector)
		if err != nil {
			fmt.Println("Error computing cosine similarity", err)
			continue
		}
		result := SearchResult{
			ID:     storedVector.ID,
			Score:  score[0],
			Vector: storedVector.Vector,
		}
		results = append(results, result)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if k < len(results) {
		results = results[:k]
	}

	return results, nil
}
