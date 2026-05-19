// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fft

import (
	"fmt"
	"testing"

	fext "github.com/consensys/gnark-crypto/field/koalabear/extensions"
)

func BenchmarkKoalaBearExtFFTSelectedRange(b *testing.B) {
	for _, logSize := range []int{9, 10, 12, 16, 20, 24} {
		size := 1 << logSize
		b.Run(fmt.Sprintf("e4/2**%d/fft_DIT", logSize), func(b *testing.B) {
			pol := makeE4BenchVector(size)
			domain := NewDomain(uint64(size))
			b.ResetTimer()
			for range b.N {
				domain.FFTExt(pol, DIT)
			}
		})
		b.Run(fmt.Sprintf("e4/2**%d/fft_DIT_coset", logSize), func(b *testing.B) {
			pol := makeE4BenchVector(size)
			domain := NewDomain(uint64(size))
			b.ResetTimer()
			for range b.N {
				domain.FFTExt(pol, DIT, OnCoset())
			}
		})
		b.Run(fmt.Sprintf("e4/2**%d/inverse_DIF", logSize), func(b *testing.B) {
			pol := makeE4BenchVector(size)
			domain := NewDomain(uint64(size))
			b.ResetTimer()
			for range b.N {
				domain.FFTInverseExt(pol, DIF)
			}
		})
		b.Run(fmt.Sprintf("e4/2**%d/inverse_DIF_coset", logSize), func(b *testing.B) {
			pol := makeE4BenchVector(size)
			domain := NewDomain(uint64(size))
			b.ResetTimer()
			for range b.N {
				domain.FFTInverseExt(pol, DIF, OnCoset())
			}
		})
		b.Run(fmt.Sprintf("e6/2**%d/fft_DIT", logSize), func(b *testing.B) {
			pol := makeE6BenchVector(size)
			domain := NewDomain(uint64(size))
			b.ResetTimer()
			for range b.N {
				domain.FFTExt6(pol, DIT)
			}
		})
		b.Run(fmt.Sprintf("e6/2**%d/fft_DIT_coset", logSize), func(b *testing.B) {
			pol := makeE6BenchVector(size)
			domain := NewDomain(uint64(size))
			b.ResetTimer()
			for range b.N {
				domain.FFTExt6(pol, DIT, OnCoset())
			}
		})
		b.Run(fmt.Sprintf("e6/2**%d/inverse_DIF", logSize), func(b *testing.B) {
			pol := makeE6BenchVector(size)
			domain := NewDomain(uint64(size))
			b.ResetTimer()
			for range b.N {
				domain.FFTInverseExt6(pol, DIF)
			}
		})
		b.Run(fmt.Sprintf("e6/2**%d/inverse_DIF_coset", logSize), func(b *testing.B) {
			pol := makeE6BenchVector(size)
			domain := NewDomain(uint64(size))
			b.ResetTimer()
			for range b.N {
				domain.FFTInverseExt6(pol, DIF, OnCoset())
			}
		})
	}
}

func makeE4BenchVector(size int) []fext.E4 {
	pol := make([]fext.E4, size)
	pol[0].MustSetRandom()
	for i := 1; i < size; i++ {
		pol[i] = pol[i-1]
	}
	return pol
}

func makeE6BenchVector(size int) []fext.E6 {
	pol := make([]fext.E6, size)
	pol[0].MustSetRandom()
	for i := 1; i < size; i++ {
		pol[i] = pol[i-1]
	}
	return pol
}
