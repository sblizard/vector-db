package math

import (
	"math"

	"github.com/sblizard/vector-db/internal/util"
)

func EuclideanDistance(a, b []float32) float32 {
	var sum float32
	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}
	return float32(math.Sqrt(float64(sum)))
}

var SafeEuclideanDistance = util.SafeVecOp(func(a, b []float32) ([]float32, error) {
	return []float32{EuclideanDistance(a, b)}, nil
})

func CosineSimilarity(a, b []float32) float32 {
	dot := Dot(a, b)
	return dot / (L2Norm(a) * L2Norm(b))
}

var SafeCosineSimilarity = util.SafeVecOp(func(a, b []float32) ([]float32, error) {
	return []float32{CosineSimilarity(a, b)}, nil
})
