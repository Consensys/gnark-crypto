// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package extensions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestE4CbrtOnCubicResidues(t *testing.T) {
	for i := 0; i < 128; i++ {
		var a, x, got, check E4
		a.MustSetRandom()
		x.Square(&a).Mul(&x, &a)
		require.NotNil(t, got.Cbrt(&x))
		check.Square(&got).Mul(&check, &got)
		require.True(t, check.Equal(&x))
	}
}

func TestE4CbrtRejectsNonResidues(t *testing.T) {
	var x, got E4
	for i := 0; i < 256; i++ {
		x.MustSetRandom()
		if got.Cbrt(&x) == nil {
			return
		}
	}
	t.Fatal("failed to find an E4 non-cube in 256 samples")
}

func BenchmarkE4Cbrt(b *testing.B) {
	var a, x E4
	a.MustSetRandom()
	x.Square(&a).Mul(&x, &a)
	var z E4
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if z.Cbrt(&x) == nil {
			b.Fatal("expected cubic residue to have a cube root")
		}
	}
}
