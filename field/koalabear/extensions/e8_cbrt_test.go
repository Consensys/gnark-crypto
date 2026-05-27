// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package extensions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestE8CbrtZero(t *testing.T) {
	var zero, got E8
	require.NotNil(t, got.Cbrt(&zero))
	require.True(t, got.IsZero())
}

func TestE8CbrtOnCubicResidues(t *testing.T) {
	for i := 0; i < 128; i++ {
		var a, x, got, check E8
		a.MustSetRandom()
		x.Square(&a).Mul(&x, &a)
		require.NotNil(t, got.Cbrt(&x))
		check.Square(&got).Mul(&check, &got)
		require.True(t, check.Equal(&x))
	}
}

func TestE8CbrtReceiverSafeAlias(t *testing.T) {
	for i := 0; i < 32; i++ {
		var a, x, inPlace, check E8
		a.MustSetRandom()
		x.Square(&a).Mul(&x, &a)
		inPlace.Set(&x)
		require.NotNil(t, inPlace.Cbrt(&inPlace))
		check.Square(&inPlace).Mul(&check, &inPlace)
		require.True(t, check.Equal(&x), "E8 in-place Cbrt must satisfy z^3 == x")
	}
}

func TestE8CbrtRejectsNonResidues(t *testing.T) {
	var x, got E8
	for i := 0; i < 256; i++ {
		x.MustSetRandom()
		if got.Cbrt(&x) == nil {
			return
		}
	}
	t.Fatal("failed to find an E8 non-cube in 256 samples")
}

func BenchmarkE8Cbrt(b *testing.B) {
	var a, x E8
	a.MustSetRandom()
	x.Square(&a).Mul(&x, &a)
	var z E8
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if z.Cbrt(&x) == nil {
			b.Fatal("expected cubic residue to have a cube root")
		}
	}
}
