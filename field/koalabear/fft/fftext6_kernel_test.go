// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fft

import (
	"testing"

	fext "github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/stretchr/testify/require"
)

func TestExt6SmallKernelSpecializations(t *testing.T) {
	domain := NewDomain(1 << 9)

	for i := 0; i < 32; i++ {
		a := make([]fext.E6, 8)
		for j := range a {
			a[j].MustSetRandom()
		}

		t.Run("DIT M2", func(t *testing.T) {
			ref := append([]fext.E6(nil), a[:4]...)
			got := append([]fext.E6(nil), a[:4]...)
			innerDITWithTwiddlesExt6(ref, domain.twiddles[7], 0, 2, 2)
			innerDITWithTwiddlesExt6M2(got, domain.twiddles[7])
			require.Equal(t, ref, got)
		})

		t.Run("DIF M2", func(t *testing.T) {
			ref := append([]fext.E6(nil), a[:4]...)
			got := append([]fext.E6(nil), a[:4]...)
			innerDIFWithTwiddlesExt6(ref, domain.twiddles[7], 0, 2, 2)
			innerDIFWithTwiddlesExt6M2(got, domain.twiddles[7])
			require.Equal(t, ref, got)
		})

		t.Run("DIT M4", func(t *testing.T) {
			ref := append([]fext.E6(nil), a...)
			got := append([]fext.E6(nil), a...)
			innerDITWithTwiddlesExt6(ref, domain.twiddles[6], 0, 4, 4)
			innerDITWithTwiddlesExt6M4(got, domain.twiddles[6])
			require.Equal(t, ref, got)
		})

		t.Run("DIF M4", func(t *testing.T) {
			ref := append([]fext.E6(nil), a...)
			got := append([]fext.E6(nil), a...)
			innerDIFWithTwiddlesExt6(ref, domain.twiddles[6], 0, 4, 4)
			innerDIFWithTwiddlesExt6M4(got, domain.twiddles[6])
			require.Equal(t, ref, got)
		})
	}
}
