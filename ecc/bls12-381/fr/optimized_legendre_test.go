// Copyright 2020-2023 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fr

import (
	"math/big"
	"testing"
)

func TestLegendreOptimized(t *testing.T) {
	var zero, one Element
	zero.SetZero()
	one.SetOne()

	// Create test elements
	// We'll test a set of special values plus some random values
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
	// For the BLS12-381 scalar field, we compute these values using the standard Legendre method
	// to determine their expected status
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

// Helper function to generate elements with known Legendre symbols
func generateElementWithLegendre(legendre int) Element {
	// Get field modulus
	p := Modulus()
	_ = new(big.Int).Sub(p, big.NewInt(1)) // p-1, might be used for future extensions

	var res Element

	// Keep generating random elements until we find one with the desired Legendre symbol
	for {
		res.SetRandom()
		if res.IsZero() {
			continue
		}

		if res.Legendre() == legendre {
			return res
		}
	}
}
