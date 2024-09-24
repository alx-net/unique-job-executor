package utils

import (
	"fmt"
	"math"
	"strings"

	"github.com/JohnCGriffin/overflow"
	"github.com/cespare/xxhash"
)

func Fibonacci(n int64) (int64, error) {
	a, b := int64(0), int64(1)

	if n < 1 {
		return a, nil
	}

	for i := int64(0); i < n; i++ {

		res, ok := overflow.Add64(a, b)

		if !ok {
			return 0, fmt.Errorf("integer overflow")
		}
		b, a = res, b
	}
	return a, nil
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

func HashFromStrings(args ...string) uint64 {
	return xxhash.Sum64([]byte(strings.Join(args, "")))
}
