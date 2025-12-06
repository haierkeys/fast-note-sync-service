package util

import "math"

func Ceil(a, b int64) int64 {
	if b == 0 {
		panic("division by zero")
	}
	return int64(math.Ceil(float64(a) / float64(b)))
}
