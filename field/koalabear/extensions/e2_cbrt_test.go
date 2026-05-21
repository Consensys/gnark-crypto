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
