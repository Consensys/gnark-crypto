// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package extensions

import "testing"

func TestE2CbrtOnCubicResidues(t *testing.T) {
	for i := 0; i < 128; i++ {
		var a, x, got, check E2
		a.MustSetRandom()
		x.Square(&a).Mul(&x, &a)
		if got.Cbrt(&x) == nil {
			t.Fatal("expected cubic residue to have a cube root")
		}
		check.Square(&got).Mul(&check, &got)
		if !check.Equal(&x) {
			t.Fatal("returned cube root does not verify")
		}
	}
}

// Regression test for the receiver-aliasing bug where z.A0 was written
// before x.A0 was read on the main path (and similarly in the x.A0==0
// branch, where the verify-against-x step ran after z had been overwritten).
// Ensures z == x is safe for E2/E4/E8 cube roots.
func TestE2CbrtReceiverSafeAlias(t *testing.T) {
	for i := 0; i < 64; i++ {
		var a, x, expected, inPlace, check E2
		a.MustSetRandom()
		x.Square(&a).Mul(&x, &a)

		if expected.Cbrt(&x) == nil {
			t.Fatal("expected cubic residue to have a cube root")
		}

		inPlace.Set(&x)
		if inPlace.Cbrt(&inPlace) == nil {
			t.Fatal("in-place Cbrt returned nil on a cubic residue")
		}
		check.Square(&inPlace).Mul(&check, &inPlace)
		if !check.Equal(&x) {
			t.Fatal("in-place Cbrt does not verify against original x")
		}
	}
}

// Exercises the x.A0 == 0 branch (purely imaginary) under z == x aliasing.
func TestE2CbrtReceiverSafeAliasPureImaginary(t *testing.T) {
	for i := 0; i < 64; i++ {
		var a, x, check E2
		a.MustSetRandom()
		a.A0.SetZero()
		x.Square(&a).Mul(&x, &a)
		if x.IsZero() {
			continue
		}
		orig := x
		if x.Cbrt(&x) == nil {
			continue
		}
		check.Square(&x).Mul(&check, &x)
		if !check.Equal(&orig) {
			t.Fatal("in-place Cbrt does not verify on x with A0==0")
		}
	}
}

func TestE2CbrtRejectsNonResidues(t *testing.T) {
	var x, got E2
	for i := 0; i < 256; i++ {
		x.MustSetRandom()
		if got.Cbrt(&x) == nil {
			return
		}
	}
	t.Fatal("failed to find an E2 non-cube in 256 samples")
}

func BenchmarkE2Cbrt(b *testing.B) {
	var a, x E2
	a.MustSetRandom()
	x.Square(&a).Mul(&x, &a)
	var z E2
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if z.Cbrt(&x) == nil {
			b.Fatal("expected cubic residue to have a cube root")
		}
	}
}
