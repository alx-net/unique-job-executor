package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFibonacci(t *testing.T) {

	expected := []int64{0, 1, 1, 2, 3, 5, 8, 13, 21, 34, 55}

	for i, el := range expected {
		actual, err := Fibonacci(int64(i))
		assert.NoError(t, err)
		assert.Equal(t, el, actual)
	}
}

func TestIsPrime(t *testing.T) {

	expected := []int64{2, 3, 5, 7919, 6607}

	for _, el := range expected {
		assert.Equal(t, true, IsPrime(el), el)
	}

	expected = []int64{934 * 458, 8437 * 5920, 5498 * 2097, 7919 * 6607}

	for _, el := range expected {
		assert.Equal(t, false, IsPrime(el), el)
	}
}
