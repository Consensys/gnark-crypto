// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package bls12381

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
