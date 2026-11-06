package utils

import (
	"errors"
	"math"
)

// CalculateEuclideanDistance compares two face embedding vectors
func CalculateEuclideanDistance(vec1, vec2 []float32) (float64, error) {
	if len(vec1) != len(vec2) {
		return 0, errors.New("vector lengths do not match")
	}
	var sum float64
	for i := range vec1 {
		diff := float64(vec1[i] - vec2[i])
		sum += diff * diff
	}
	return math.Sqrt(sum), nil
}
