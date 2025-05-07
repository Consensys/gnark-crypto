// Copyright 2020-2023 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package field

import (
	"math/big"
	"testing"
)

// mockElement implements ElementInterface for testing
type mockElement struct {
	value *big.Int
	prime *big.Int // Shared prime across all mockElements
}

// Set sets z to x and returns z
func (z *mockElement) Set(x ElementInterface) ElementInterface {
	xMock := x.(*mockElement)
	if z.value == nil {
		z.value = new(big.Int)
	}
	z.value.Set(xMock.value)
	z.prime = xMock.prime
	return z
}

// IsZero returns true if z equals 0
func (z *mockElement) IsZero() bool {
	return z.value.Sign() == 0
}

// Neg sets z to -x (mod p) and returns z
func (z *mockElement) Neg(x ElementInterface) ElementInterface {
	xMock := x.(*mockElement)
	if z.value == nil {
		z.value = new(big.Int)
	}
	z.value.Neg(xMock.value)
	z.value.Mod(z.value, xMock.prime)
	z.prime = xMock.prime
	return z
}

// Equal returns true if z equals x
func (z *mockElement) Equal(x ElementInterface) bool {
	xMock := x.(*mockElement)
	return z.value.Cmp(xMock.value) == 0
}

// mockInverse computes the modular inverse using a simplified version of Pornin's algorithm
// that also provides information about the Legendre symbol
func mockInverse(z, x ElementInterface) ElementInterface {
	if z == nil {
		z = &mockElement{value: new(big.Int), prime: x.(*mockElement).prime}
	}
	xMock := x.(*mockElement)
	if z.(*mockElement).value == nil {
		z.(*mockElement).value = new(big.Int)
	}
	z.(*mockElement).prime = xMock.prime

	if xMock.value.Sign() == 0 {
		z.(*mockElement).value.SetInt64(0)
		return z
	}

	// Compute (p-1)/2
	exp := new(big.Int).Sub(xMock.prime, big.NewInt(1))
	exp.Rsh(exp, 1)

	// Calculate x^((p-1)/2) to determine if x is a quadratic residue
	legendreResult := new(big.Int).Exp(xMock.value, exp, xMock.prime)

	// For a non-residue, return -1 (represented as p-1 in the field)
	if legendreResult.Cmp(big.NewInt(1)) != 0 {
		z.(*mockElement).value.Sub(xMock.prime, big.NewInt(1))
		return z
	}

	// For a residue, compute the actual inverse
	z.(*mockElement).value.ModInverse(xMock.value, xMock.prime)
	return z
}

// mockNegOne returns -1
func mockNegOne() ElementInterface {
	prime := big.NewInt(11) // Using a small prime for testing
	value := new(big.Int).Sub(prime, big.NewInt(1))
	return &mockElement{value: value, prime: prime}
}

// mockOne returns 1
func mockOne() ElementInterface {
	prime := big.NewInt(11) // Using a small prime for testing
	return &mockElement{value: big.NewInt(1), prime: prime}
}

// mockNew creates a new mockElement with given int64 value
func mockNew(val int64) *mockElement {
	prime := big.NewInt(11) // Using a small prime for testing
	value := new(big.Int).Mod(big.NewInt(val), prime)
	return &mockElement{value: value, prime: prime}
}

// standardLegendre computes the Legendre symbol using exponentiation
func standardLegendre(x *mockElement) int {
	if x.value.Sign() == 0 {
		return 0
	}

	// Compute (p-1)/2
	exp := new(big.Int).Sub(x.prime, big.NewInt(1))
	exp.Rsh(exp, 1)

	// Compute x^((p-1)/2) mod p
	res := new(big.Int).Exp(x.value, exp, x.prime)

	// Compare with -1, 0, and 1
	if res.Cmp(big.NewInt(1)) == 0 {
		return 1
	} else if res.Cmp(new(big.Int).Sub(x.prime, big.NewInt(1))) == 0 {
		return -1
	}

	return 0
}

func TestLegendrePornin(t *testing.T) {
	// Test values for prime p = 11
	// The quadratic residues modulo 11 are 1, 3, 4, 5, 9
	// The quadratic non-residues modulo 11 are 2, 6, 7, 8, 10
	tests := []struct {
		value    int64
		expected int
	}{
		{0, 0},   // 0 is neither residue nor non-residue
		{1, 1},   // 1 is a quadratic residue
		{2, -1},  // 2 is a quadratic non-residue
		{3, 1},   // 3 is a quadratic residue
		{4, 1},   // 4 is a quadratic residue
		{5, 1},   // 5 is a quadratic residue
		{6, -1},  // 6 is a quadratic non-residue
		{7, -1},  // 7 is a quadratic non-residue
		{8, -1},  // 8 is a quadratic non-residue
		{9, 1},   // 9 is a quadratic residue
		{10, -1}, // 10 is a quadratic non-residue
	}

	for _, test := range tests {
		element := mockNew(test.value)

		// Calculate using the optimized method
		optimizedResult := LegendrePornin(element, mockInverse, mockNegOne, mockOne)

		// Calculate using the standard method for comparison
		standardResult := standardLegendre(element)

		if optimizedResult != standardResult {
			t.Errorf("LegendrePornin mismatch for value %d: got %d, want %d",
				test.value, optimizedResult, standardResult)
		}

		if optimizedResult != test.expected {
			t.Errorf("LegendrePornin for value %d: got %d, want %d",
				test.value, optimizedResult, test.expected)
		}
	}
}

// Benchmark comparing the optimized and standard Legendre implementations
func BenchmarkLegendre(b *testing.B) {
	// Setup
	element := mockNew(7) // Use a quadratic non-residue for benchmarking

	b.Run("StandardLegendre", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			standardLegendre(element)
		}
	})

	b.Run("OptimizedLegendre", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			LegendrePornin(element, mockInverse, mockNegOne, mockOne)
		}
	})
}
