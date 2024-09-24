package utils

import "math"

func Fibonacci(n int) uint64 {
	a, b := uint64(0), uint64(1)

	if n < 1 {
		return a
	}

	for i := 0; i < n; i++ {
		b, a = a+b, b
	}
	return a
}

func IsPrime(p int64) bool {
	if p < 2 {
		return false
	}

	for i := int64(2); i < int64(math.Sqrt(float64(p)))+1; i++ {
		if p%i == 0 {
			return false
		}
	}

	return true
}
