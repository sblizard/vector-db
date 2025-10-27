package math

import "github.com/sblizard/vector-db/internal/util"

func Dot(a, b []float32) float32 {
	var sum float32
	for index := range a[:] {
		sum += a[index] * b[index]
	}
	return sum
}

var SafeDot = util.SafeVecOp(func(a, b []float32) ([]float32, error) {
	return []float32{Dot(a, b)}, nil
})
