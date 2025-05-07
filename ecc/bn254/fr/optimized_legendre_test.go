// Copyright 2020-2023 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fr

import (
	"testing"
)

func TestLegendreOptimized(t *testing.T) {
	var zero, one Element
	zero.SetZero()
	one.SetOne()

	// Create test elements
	var two, three, four, five, seven, negOne Element
	two.SetUint64(2)
	three.SetUint64(3)
	four.SetUint64(4)
	five.SetUint64(5)
	seven.SetUint64(7)
	negOne.Neg(&one)

	// Add some random elements
	var rnd1, rnd2, rnd3 Element
	rnd1.SetRandom()
	rnd2.SetRandom()
	rnd3.SetRandom()

	// Known specific values (quadratic residue status is field-specific)
	tests := []struct {
		name    string
		element Element
	}{
		{"zero", zero},
		{"one", one},
		{"two", two},
		{"three", three},
		{"four", four},
		{"five", five},
		{"seven", seven},
		{"negOne", negOne},
		{"random1", rnd1},
		{"random2", rnd2},
		{"random3", rnd3},
	}

	for _, test := range tests {
		// Compute results using both methods
		standard := test.element.Legendre()
		optimized := test.element.LegendreOptimized()

		if standard != optimized {
			t.Errorf("Legendre mismatch for %s: standard=%d, optimized=%d",
				test.name, standard, optimized)
		}
	}
}

// TestRandomLegendreComparison tests a large number of random elements
func TestRandomLegendreComparison(t *testing.T) {
	const numTests = 100

	for i := 0; i < numTests; i++ {
		var e Element
		e.SetRandom()

		standard := e.Legendre()
		optimized := e.LegendreOptimized()

		if standard != optimized {
			t.Errorf("Legendre mismatch for random element #%d: standard=%d, optimized=%d",
				i+1, standard, optimized)
		}
	}
}

// TestSqrtOptimized tests the optimized square root implementation
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

	// Test cases
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

// BenchmarkLegendreSymbol benchmarks both Legendre symbol implementations
func BenchmarkLegendreSymbol(b *testing.B) {
	// Setup - generate a random element
	var e Element
	e.SetRandom()

	// Benchmark the standard implementation
	b.Run("Standard", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			e.Legendre()
		}
	})

	// Benchmark the optimized implementation
	b.Run("Optimized", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			e.LegendreOptimized()
		}
	})
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
