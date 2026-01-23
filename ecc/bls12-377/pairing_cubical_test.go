// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package bls12377

import (
	"math/big"
	"testing"
)

var (
	big2 = big.NewInt(2)
	big3 = big.NewInt(3)
)

func TestCubicalPairing(t *testing.T) {
	// Get generator points
	_, _, g1Aff, g2Aff := Generators()

	P := []G1Affine{g1Aff}
	Q := []G2Affine{g2Aff}

	// Compute standard pairing and square it
	standardPairing, err := Pair(P, Q)
	if err != nil {
		t.Fatalf("Standard pairing failed: %v", err)
	}
	var standardSquared GT
	standardSquared.Square(&standardPairing)

	// Compute cubical pairing (should equal squared standard pairing)
	cubicalPairing, err := PairCubical(P, Q)
	if err != nil {
		t.Fatalf("Cubical pairing failed: %v", err)
	}

	// Compare results
	if !cubicalPairing.Equal(&standardSquared) {
		t.Errorf("Cubical pairing does not equal squared standard pairing")
		t.Logf("Standard pairing squared: %s", standardSquared.String())
		t.Logf("Cubical pairing: %s", cubicalPairing.String())
	}
}

func TestCubicalPairingMultiple(t *testing.T) {
	// Get generator points
	_, _, g1Aff, g2Aff := Generators()

	// Create multiple points by scaling
	var g1Scaled G1Affine
	var g2Scaled G2Affine
	g1Scaled.ScalarMultiplication(&g1Aff, big2)
	g2Scaled.ScalarMultiplication(&g2Aff, big3)

	P := []G1Affine{g1Aff, g1Scaled}
	Q := []G2Affine{g2Aff, g2Scaled}

	// Compute standard pairing product and square it
	standardPairing, err := Pair(P, Q)
	if err != nil {
		t.Fatalf("Standard pairing failed: %v", err)
	}
	var standardSquared GT
	standardSquared.Square(&standardPairing)

	// Compute cubical pairing (should equal squared standard pairing)
	cubicalPairing, err := PairCubical(P, Q)
	if err != nil {
		t.Fatalf("Cubical pairing failed: %v", err)
	}

	// Compare results
	if !cubicalPairing.Equal(&standardSquared) {
		t.Errorf("Cubical pairing does not equal squared standard pairing for multiple inputs")
	}
}

func BenchmarkCubicalPairing(b *testing.B) {
	_, _, g1Aff, g2Aff := Generators()
	P := []G1Affine{g1Aff}
	Q := []G2Affine{g2Aff}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = PairCubical(P, Q)
	}
}

// Tests for fixed-Q cubical pairing

func TestCubicalPairingFixedQ(t *testing.T) {
	_, _, g1Aff, g2Aff := Generators()

	// Precompute Q
	pre := PrecomputeG2Cubical(&g2Aff)

	// Compute fixed-Q pairing
	fixedQPairing, err := PairCubicalFixedQ(&g1Aff, pre)
	if err != nil {
		t.Fatalf("Fixed-Q cubical pairing failed: %v", err)
	}

	// Compute standard cubical pairing for comparison
	P := []G1Affine{g1Aff}
	Q := []G2Affine{g2Aff}
	cubicalPairing, err := PairCubical(P, Q)
	if err != nil {
		t.Fatalf("Cubical pairing failed: %v", err)
	}

	// Compare results
	if !fixedQPairing.Equal(&cubicalPairing) {
		t.Errorf("Fixed-Q cubical pairing does not equal standard cubical pairing")
		t.Logf("Standard cubical: %s", cubicalPairing.String())
		t.Logf("Fixed-Q cubical: %s", fixedQPairing.String())
	}
}

func TestCubicalPairingFixedQMultipleP(t *testing.T) {
	_, _, g1Aff, g2Aff := Generators()

	// Create multiple P points
	var g1Scaled1, g1Scaled2 G1Affine
	g1Scaled1.ScalarMultiplication(&g1Aff, big2)
	g1Scaled2.ScalarMultiplication(&g1Aff, big3)

	P := []G1Affine{g1Aff, g1Scaled1, g1Scaled2}

	// Precompute Q
	pre := PrecomputeG2Cubical(&g2Aff)

	// Compute using fixed-Q for each P individually
	var fixedQResult GT
	fixedQResult.SetOne()
	for _, p := range P {
		pairK, err := PairCubicalFixedQ(&p, pre)
		if err != nil {
			t.Fatalf("Fixed-Q pairing failed: %v", err)
		}
		fixedQResult.Mul(&fixedQResult, &pairK)
	}

	// Compute using standard cubical pairing
	Q := []G2Affine{g2Aff, g2Aff, g2Aff}
	standardResult, err := PairCubical(P, Q)
	if err != nil {
		t.Fatalf("Standard cubical pairing failed: %v", err)
	}

	// Compare
	if !fixedQResult.Equal(&standardResult) {
		t.Errorf("Fixed-Q multi-P result does not match standard cubical pairing")
	}
}

func TestCubicalPairingFixedQVsStandard(t *testing.T) {
	_, _, g1Aff, g2Aff := Generators()

	// Precompute Q
	pre := PrecomputeG2Cubical(&g2Aff)

	// Fixed-Q pairing
	fixedQPairing, err := PairCubicalFixedQ(&g1Aff, pre)
	if err != nil {
		t.Fatalf("Fixed-Q cubical pairing failed: %v", err)
	}

	// Standard pairing squared
	P := []G1Affine{g1Aff}
	Q := []G2Affine{g2Aff}
	standardPairing, err := Pair(P, Q)
	if err != nil {
		t.Fatalf("Standard pairing failed: %v", err)
	}
	var standardSquared GT
	standardSquared.Square(&standardPairing)

	// Compare: fixed-Q cubical should equal standard pairing squared
	if !fixedQPairing.Equal(&standardSquared) {
		t.Errorf("Fixed-Q cubical pairing does not equal squared standard pairing")
	}
}

func BenchmarkCubicalPairingFixedQPrecompute(b *testing.B) {
	_, _, _, g2Aff := Generators()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = PrecomputeG2Cubical(&g2Aff)
	}
}

func BenchmarkCubicalPairingFixedQOnline(b *testing.B) {
	_, _, g1Aff, g2Aff := Generators()
	pre := PrecomputeG2Cubical(&g2Aff)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = PairCubicalFixedQ(&g1Aff, pre)
	}
}

func BenchmarkCubicalPairingFixedQMulti(b *testing.B) {
	_, _, g1Aff, g2Aff := Generators()

	// Create 4 distinct P points
	var g1Scaled1, g1Scaled2, g1Scaled3 G1Affine
	g1Scaled1.ScalarMultiplication(&g1Aff, big2)
	g1Scaled2.ScalarMultiplication(&g1Aff, big3)
	g1Scaled3.ScalarMultiplication(&g1Aff, big.NewInt(5))

	P := []G1Affine{g1Aff, g1Scaled1, g1Scaled2, g1Scaled3}
	pre := PrecomputeG2Cubical(&g2Aff)

	b.Run("Batched", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = PairCubicalFixedQMulti(P, pre)
		}
	})

	b.Run("Individual", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var result GT
			result.SetOne()
			for _, p := range P {
				pairK, _ := PairCubicalFixedQ(&p, pre)
				result.Mul(&result, &pairK)
			}
		}
	})
}

// Benchmarks comparing standard MillerLoop FixedQ vs Cubical FixedQ

func BenchmarkFixedQComparison(b *testing.B) {
	_, _, g1Aff, g2Aff := Generators()

	// Standard Miller loop precomputation
	standardLines := PrecomputeLines(g2Aff)
	standardPre := [][2][len(LoopCounter) - 1]LineEvaluationAff{standardLines}

	// Cubical precomputation
	cubicalPre := PrecomputeG2Cubical(&g2Aff)

	b.Run("StandardMillerLoop_Precompute", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = PrecomputeLines(g2Aff)
		}
	})

	b.Run("CubicalMillerLoop_Precompute", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = PrecomputeG2Cubical(&g2Aff)
		}
	})

	b.Run("StandardMillerLoop_Online", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = MillerLoopFixedQ([]G1Affine{g1Aff}, standardPre)
		}
	})

	b.Run("CubicalMillerLoop_Online", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = MillerLoopCubicalFixedQ(&g1Aff, cubicalPre)
		}
	})

	b.Run("StandardPairing_FixedQ", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ml, _ := MillerLoopFixedQ([]G1Affine{g1Aff}, standardPre)
			_ = FinalExponentiation(&ml)
		}
	})

	b.Run("CubicalPairing_FixedQ", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = PairCubicalFixedQ(&g1Aff, cubicalPre)
		}
	})
}
