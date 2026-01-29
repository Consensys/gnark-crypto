// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fptower

import (
	"math/big"
	"testing"
)

func TestCbrtFrobeniusCorrectness(t *testing.T) {
	// First, verify the Frobenius decomposition is mathematically correct
	// by computing x^e directly vs using the optimized chain
	var x E2
	var tmp E2
	tmp.MustSetRandom()
	x.Square(&tmp).Mul(&x, &tmp) // x = tmp³ (a cubic residue)

	// p for BN254
	p, _ := new(big.Int).SetString("21888242871839275222246405745257275088696311157297823662689037894645226208583", 10)

	// e = (2p² + 7) / 27 (correct exponent for E2.Cbrt)
	pSquared := new(big.Int).Mul(p, p)
	e := new(big.Int).Mul(pSquared, big.NewInt(2))
	e.Add(e, big.NewInt(7))
	e.Div(e, big.NewInt(27))

	// Compute x^e directly
	var yDirect E2
	yDirect.Exp(x, e)

	// Compute using the optimized chain
	var yChain E2
	yChain.expByCbrtFrobeniusChain(&x)
	t.Logf("x^e direct: %s", yDirect.String())
	t.Logf("Chain result: %s", yChain.String())
	t.Logf("Equal: %v", yDirect.Equal(&yChain))

	if !yDirect.Equal(&yChain) {
		t.Fatal("Optimized chain gives wrong result!")
	}

	// Test the full CbrtFrobenius
	var y1, y2 E2
	result1 := y1.Cbrt(&x)
	result2 := y2.CbrtFrobenius(&x)

	if result1 == nil {
		t.Fatal("Original Cbrt returned nil for cubic residue")
	}
	if result2 == nil {
		t.Fatal("CbrtFrobenius returned nil for cubic residue")
	}

	// Verify both produce valid cube roots
	var c1, c2 E2
	c1.Square(&y1).Mul(&c1, &y1)
	c2.Square(&y2).Mul(&c2, &y2)
	if !c1.Equal(&x) || !c2.Equal(&x) {
		t.Fatal("Cube root verification failed")
	}
}

func TestCbrtFrobeniusEdgeCases(t *testing.T) {
	// Test with x in Fp (imaginary part = 0)
	var x, y1, y2 E2
	x.A0.SetUint64(27) // 27 = 3³ is a perfect cube
	x.A1.SetZero()

	result1 := y1.Cbrt(&x)
	result2 := y2.CbrtFrobenius(&x)

	if result1 == nil || result2 == nil {
		t.Fatal("Failed to compute cube root of element in Fp")
	}

	// Both should produce a cube root
	var c1, c2 E2
	c1.Square(&y1).Mul(&c1, &y1)
	c2.Square(&y2).Mul(&c2, &y2)

	if !c1.Equal(&x) || !c2.Equal(&x) {
		t.Fatal("Cube root verification failed for element in Fp")
	}
}

func BenchmarkE2CbrtOriginal(b *testing.B) {
	var x, y E2
	x.MustSetRandom()
	// Make x a cubic residue
	var tmp E2
	tmp.MustSetRandom()
	x.Square(&tmp).Mul(&x, &tmp)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		y.Cbrt(&x)
	}
}

func BenchmarkE2CbrtFrobenius(b *testing.B) {
	var x, y E2
	x.MustSetRandom()
	// Make x a cubic residue
	var tmp E2
	tmp.MustSetRandom()
	x.Square(&tmp).Mul(&x, &tmp)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		y.CbrtFrobenius(&x)
	}
}

func BenchmarkE2SqrtTwice(b *testing.B) {
	// Benchmark two Sqrt operations for comparison
	var x, y E2
	x.MustSetRandom()
	// Make x a quadratic residue
	var tmp E2
	tmp.MustSetRandom()
	x.Square(&tmp)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		y.Sqrt(&x)
		y.Sqrt(&y)
	}
}

func TestCbrtFrobeniusNonResidue(t *testing.T) {
	// Test that CbrtFrobenius correctly returns nil for non-cubic residues
	for i := 0; i < 20; i++ {
		var x E2
		x.MustSetRandom()

		result1 := new(E2).Cbrt(&x)
		result2 := new(E2).CbrtFrobenius(&x)

		if (result1 == nil) != (result2 == nil) {
			t.Fatalf("Mismatch for non-residue detection: original=%v, frobenius=%v", result1 != nil, result2 != nil)
		}
	}
}

func TestCbrtFrobeniusExtensive(t *testing.T) {
	// Extensive test with many random cubic residues
	for i := 0; i < 100; i++ {
		var x, y1, y2, tmp E2
		tmp.MustSetRandom()
		x.Square(&tmp).Mul(&x, &tmp) // x = tmp³

		result1 := y1.Cbrt(&x)
		result2 := y2.CbrtFrobenius(&x)

		if result1 == nil || result2 == nil {
			t.Fatal("Failed to compute cube root of cubic residue")
		}

		// Verify both results
		var c1, c2 E2
		c1.Square(&y1).Mul(&c1, &y1)
		c2.Square(&y2).Mul(&c2, &y2)

		if !c1.Equal(&x) || !c2.Equal(&x) {
			t.Fatal("Cube root verification failed")
		}
	}
}
