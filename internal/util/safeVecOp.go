package util

import "errors"

type Number interface {
	~float32 | ~float64 | ~int | ~int64
}

type VecOp[T Number] func(a, b []T) ([]T, error)

func SafeVecOp[T Number](op VecOp[T]) VecOp[T] {
	return func(a, b []T) ([]T, error) {
		if len(a) != len(b) {
			return nil, errors.New("vector length mismatch, ye scurvy dog")
		}
		return op(a, b)
	}
}
