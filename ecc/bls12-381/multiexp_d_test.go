// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package bls12381

import (
	"fmt"
	"math/bits"
	"math/rand/v2"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
)

func TestMultiExpLargeDG1(t *testing.T) {
	const nbSamples = 64

	var samplePoints [nbSamples]G1Affine
	var g G1Jac
	g.Set(&g1Gen)
	for i := 1; i <= nbSamples; i++ {
		samplePoints[i-1].FromJacobian(&g)
		g.AddAssign(&g1Gen)
	}

	samplePoints[rand.N(nbSamples)].SetInfinity() //#nosec G404 weak rng is fine here
	samplePoints[rand.N(nbSamples)].SetInfinity() //#nosec G404 weak rng is fine here

	var sampleScalars [nbSamples]fr.Element
	for i := 0; i < nbSamples; i++ {
		sampleScalars[i].MustSetRandom()
	}

	var got, expected G1Affine
	if _, err := got.MultiExpLargeD(samplePoints[:], sampleScalars[:]); err != nil {
		t.Fatalf("MultiExpLargeD failed: %v", err)
	}
	if _, err := expected.MultiExp(samplePoints[:], sampleScalars[:], ecc.MultiExpConfig{}); err != nil {
		t.Fatalf("MultiExp failed: %v", err)
	}

	if !expected.Equal(&got) {
		t.Fatal("MultiExpLargeD does not match MultiExp")
	}
}

func TestMultiExpSmallDG1(t *testing.T) {
	const nbSamples = 12

	var samplePoints [nbSamples]G1Affine
	var g G1Jac
	g.Set(&g1Gen)
	for i := 1; i <= nbSamples; i++ {
		samplePoints[i-1].FromJacobian(&g)
		g.AddAssign(&g1Gen)
	}

	samplePoints[rand.N(nbSamples)].SetInfinity() //#nosec G404 weak rng is fine here

	var sampleScalars [nbSamples]fr.Element
	for i := 0; i < nbSamples; i++ {
		sampleScalars[i].MustSetRandom()
	}

	var got, expected G1Affine
	if _, err := got.MultiExpSmallD(samplePoints[:], sampleScalars[:]); err != nil {
		t.Fatalf("MultiExpSmallD failed: %v", err)
	}
	if _, err := expected.MultiExp(samplePoints[:], sampleScalars[:], ecc.MultiExpConfig{}); err != nil {
		t.Fatalf("MultiExp failed: %v", err)
	}

	if !expected.Equal(&got) {
		t.Fatal("MultiExpSmallD does not match MultiExp")
	}
}

func BenchmarkMultiExpLargeDG1(b *testing.B) {
	const (
		pow       = (bits.UintSize / 2) - (bits.UintSize / 8) // 24 on 64 bits arch, 12 on 32 bits
		nbSamples = 1 << pow
	)

	var (
		samplePoints  [nbSamples]G1Affine
		sampleScalars [nbSamples]fr.Element
	)

	fillBenchScalars(sampleScalars[:])
	fillBenchBasesG1(samplePoints[:])

	var testPoint G1Affine

	for i := 5; i <= pow; i++ {
		using := 1 << i

		b.Run(fmt.Sprintf("%d points", using), func(b *testing.B) {
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				_, _ = testPoint.MultiExpLargeD(samplePoints[:using], sampleScalars[:using])
			}
		})
	}
}

func BenchmarkMultiExpSmallDG1(b *testing.B) {
	const nbSamples = 1 << 16

	var (
		samplePoints  [nbSamples]G1Affine
		sampleScalars [nbSamples]fr.Element
	)

	fillBenchScalars(sampleScalars[:])
	fillBenchBasesG1(samplePoints[:])

	var testPoint G1Affine
	sizes := []int{2, 4, 8, 12, 16}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("%d points", size), func(b *testing.B) {
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				_, _ = testPoint.MultiExpSmallD(samplePoints[:size], sampleScalars[:size])
			}
		})
	}
}

func BenchmarkMultiExpCompareMethodsG1(b *testing.B) {
	const (
		pow       = (bits.UintSize / 2) - (bits.UintSize / 8) // 24 on 64 bits arch, 12 on 32 bits
		nbSamples = 1 << pow
	)

	var (
		samplePoints  [nbSamples]G1Affine
		sampleScalars [nbSamples]fr.Element
	)

	fillBenchScalars(sampleScalars[:])
	fillBenchBasesG1(samplePoints[:])

	var testPoint G1Affine

	for i := 1; i <= pow; i++ {
		using := 1 << i

		b.Run(fmt.Sprintf("%d points/bucket", using), func(b *testing.B) {
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				testPoint.MultiExp(samplePoints[:using], sampleScalars[:using], ecc.MultiExpConfig{})
			}
		})

		b.Run(fmt.Sprintf("%d points/large-d", using), func(b *testing.B) {
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				_, _ = testPoint.MultiExpLargeD(samplePoints[:using], sampleScalars[:using])
			}
		})

		b.Run(fmt.Sprintf("%d points/small-d", using), func(b *testing.B) {
			if using > maxMultiExpSmallD {
				b.Skipf("small-d precomputation capped at d=%d", maxMultiExpSmallD)
			}
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				_, _ = testPoint.MultiExpSmallD(samplePoints[:using], sampleScalars[:using])
			}
		})
	}
}

func TestMultiExpLargeDG2(t *testing.T) {
	const nbSamples = 64

	var samplePoints [nbSamples]G2Affine
	var g G2Jac
	g.Set(&g2Gen)
	for i := 1; i <= nbSamples; i++ {
		samplePoints[i-1].FromJacobian(&g)
		g.AddAssign(&g2Gen)
	}

	samplePoints[rand.N(nbSamples)].SetInfinity() //#nosec G404 weak rng is fine here
	samplePoints[rand.N(nbSamples)].SetInfinity() //#nosec G404 weak rng is fine here

	var sampleScalars [nbSamples]fr.Element
	for i := 0; i < nbSamples; i++ {
		sampleScalars[i].MustSetRandom()
	}

	var got, expected G2Affine
	if _, err := got.MultiExpLargeD(samplePoints[:], sampleScalars[:]); err != nil {
		t.Fatalf("MultiExpLargeD failed: %v", err)
	}
	if _, err := expected.MultiExp(samplePoints[:], sampleScalars[:], ecc.MultiExpConfig{}); err != nil {
		t.Fatalf("MultiExp failed: %v", err)
	}

	if !expected.Equal(&got) {
		t.Fatal("MultiExpLargeD does not match MultiExp")
	}
}

func TestMultiExpSmallDG2(t *testing.T) {
	const nbSamples = 12

	var samplePoints [nbSamples]G2Affine
	var g G2Jac
	g.Set(&g2Gen)
	for i := 1; i <= nbSamples; i++ {
		samplePoints[i-1].FromJacobian(&g)
		g.AddAssign(&g2Gen)
	}

	samplePoints[rand.N(nbSamples)].SetInfinity() //#nosec G404 weak rng is fine here

	var sampleScalars [nbSamples]fr.Element
	for i := 0; i < nbSamples; i++ {
		sampleScalars[i].MustSetRandom()
	}

	var got, expected G2Affine
	if _, err := got.MultiExpSmallD(samplePoints[:], sampleScalars[:]); err != nil {
		t.Fatalf("MultiExpSmallD failed: %v", err)
	}
	if _, err := expected.MultiExp(samplePoints[:], sampleScalars[:], ecc.MultiExpConfig{}); err != nil {
		t.Fatalf("MultiExp failed: %v", err)
	}

	if !expected.Equal(&got) {
		t.Fatal("MultiExpSmallD does not match MultiExp")
	}
}

func BenchmarkMultiExpLargeDG2(b *testing.B) {
	const (
		pow       = (bits.UintSize / 2) - (bits.UintSize / 8) // 24 on 64 bits arch, 12 on 32 bits
		nbSamples = 1 << pow
	)

	var (
		samplePoints  [nbSamples]G2Affine
		sampleScalars [nbSamples]fr.Element
	)

	fillBenchScalars(sampleScalars[:])
	fillBenchBasesG2(samplePoints[:])

	var testPoint G2Affine

	for i := 5; i <= pow; i++ {
		using := 1 << i

		b.Run(fmt.Sprintf("%d points", using), func(b *testing.B) {
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				_, _ = testPoint.MultiExpLargeD(samplePoints[:using], sampleScalars[:using])
			}
		})
	}
}

func BenchmarkMultiExpSmallDG2(b *testing.B) {
	const nbSamples = 1 << 16

	var (
		samplePoints  [nbSamples]G2Affine
		sampleScalars [nbSamples]fr.Element
	)

	fillBenchScalars(sampleScalars[:])
	fillBenchBasesG2(samplePoints[:])

	var testPoint G2Affine
	sizes := []int{2, 4, 8, 12, 16}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("%d points", size), func(b *testing.B) {
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				_, _ = testPoint.MultiExpSmallD(samplePoints[:size], sampleScalars[:size])
			}
		})
	}
}

func BenchmarkMultiExpCompareMethodsG2(b *testing.B) {
	const (
		pow       = (bits.UintSize / 2) - (bits.UintSize / 8) // 24 on 64 bits arch, 12 on 32 bits
		nbSamples = 1 << pow
	)

	var (
		samplePoints  [nbSamples]G2Affine
		sampleScalars [nbSamples]fr.Element
	)

	fillBenchScalars(sampleScalars[:])
	fillBenchBasesG2(samplePoints[:])

	var testPoint G2Affine

	for i := 1; i <= pow; i++ {
		using := 1 << i

		b.Run(fmt.Sprintf("%d points/bucket", using), func(b *testing.B) {
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				testPoint.MultiExp(samplePoints[:using], sampleScalars[:using], ecc.MultiExpConfig{})
			}
		})

		b.Run(fmt.Sprintf("%d points/large-d", using), func(b *testing.B) {
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				_, _ = testPoint.MultiExpLargeD(samplePoints[:using], sampleScalars[:using])
			}
		})

		b.Run(fmt.Sprintf("%d points/small-d", using), func(b *testing.B) {
			if using > maxMultiExpSmallD {
				b.Skipf("small-d precomputation capped at d=%d", maxMultiExpSmallD)
			}
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				_, _ = testPoint.MultiExpSmallD(samplePoints[:using], sampleScalars[:using])
			}
		})
	}
}
