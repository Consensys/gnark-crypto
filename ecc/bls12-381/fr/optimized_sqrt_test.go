// Copyright 2020-2023 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fr

import (
	"testing"
)

func TestSqrtOptimized(t *testing.T) {
	var zero, one Element
	zero.SetZero()
	one.SetOne()

	// Create test elements
	var two, three, four, five, seven, nine, sixteen, twentyFive, negOne Element
	two.SetUint64(2)
	three.SetUint64(3)
	four.SetUint64(4)
	five.SetUint64(5)
	seven.SetUint64(7)
	nine.SetUint64(9)
	sixteen.SetUint64(16)
	twentyFive.SetUint64(25)
	negOne.Neg(&one)

	// Test cases with known quadratic residue status
	// Some values might be quadratic residues or non-residues depending on the specific field
	tests := []struct {
		name    string
		element Element
	}{
		{"zero", zero},
		{"one", one},
		{"four", four},
		{"nine", nine},
		{"sixteen", sixteen},
		{"twentyFive", twentyFive},
		{"two", two},
		{"three", three},
		{"five", five},
		{"seven", seven},
		{"negOne", negOne},
	}

	for _, test := range tests {
		// Compute square roots using both methods
		var sqrtStandard, sqrtOptimized Element

		standardResult := sqrtStandard.Sqrt(&test.element)
		optimizedResult := sqrtOptimized.SqrtOptimized(&test.element)

		// Check if both methods agree on whether the element has a square root
		if (standardResult == nil && optimizedResult != nil) ||
			(standardResult != nil && optimizedResult == nil) {
			t.Errorf("Sqrt existence mismatch for %s: standard=%v, optimized=%v",
				test.name, standardResult != nil, optimizedResult != nil)
			continue
		}

		// If there is a square root, verify it's the same with both methods
		if standardResult != nil && optimizedResult != nil {
			if !sqrtStandard.Equal(&sqrtOptimized) {
				t.Errorf("Sqrt value mismatch for %s", test.name)
			}

			// Verify that the square of the result is equal to the original element
			var squared Element
			squared.Square(&sqrtOptimized)
			if !squared.Equal(&test.element) {
				t.Errorf("Sqrt verification failed for %s: square(sqrt(x)) != x", test.name)
			}
		}
	}
}

// TestRandomSqrtComparison tests a large number of random quadratic residues
func TestRandomSqrtComparison(t *testing.T) {
	const numTests = 50

	// Generate quadratic residues by squaring random elements
	for i := 0; i < numTests; i++ {
		var x, squared Element
		x.SetRandom()
		squared.Square(&x) // squared is guaranteed to be a quadratic residue

		// Compute square roots using both methods
		var sqrtStandard, sqrtOptimized Element
		standardResult := sqrtStandard.Sqrt(&squared)
		optimizedResult := sqrtOptimized.SqrtOptimized(&squared)

		// Both methods should find a square root
		if standardResult == nil {
			t.Errorf("Standard Sqrt failed on known quadratic residue #%d", i+1)
			continue
		}

		if optimizedResult == nil {
			t.Errorf("Optimized Sqrt failed on known quadratic residue #%d", i+1)
			continue
		}

		// Verify that both methods give a valid square root (might not be the same one)
		var verifyStandard, verifyOptimized Element
		verifyStandard.Square(&sqrtStandard)
		verifyOptimized.Square(&sqrtOptimized)

		if !verifyStandard.Equal(&squared) {
			t.Errorf("Standard Sqrt verification failed for test #%d", i+1)
		}

		if !verifyOptimized.Equal(&squared) {
			t.Errorf("Optimized Sqrt verification failed for test #%d", i+1)
		}
	}
}

// BenchmarkSqrt benchmarks both square root implementations
func BenchmarkSqrt(b *testing.B) {
	// Generate a quadratic residue
	var x, qr Element
	x.SetRandom()
	qr.Square(&x)

	// Benchmark the standard implementation
	b.Run("Standard", func(b *testing.B) {
		var res Element
		for i := 0; i < b.N; i++ {
			res.Sqrt(&qr)
		}
	})

	// Benchmark the optimized implementation
	b.Run("Optimized", func(b *testing.B) {
		var res Element
		for i := 0; i < b.N; i++ {
			res.SqrtOptimized(&qr)
		}
	})
}
