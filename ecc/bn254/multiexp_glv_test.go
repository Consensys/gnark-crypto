// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package bn254

import (
	"fmt"
	"math/big"
	"math/bits"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

func TestMultiExpG1GLV(t *testing.T) {
	sizes := []int{0, 1, 2, 7, 73, 512, 1 << 12}

	for _, size := range sizes {
		t.Run(fmt.Sprintf("%d points", size), func(t *testing.T) {
			points := make([]G1Affine, size)
			scalars := make([]fr.Element, size)

			var g G1Jac
			g.Set(&g1Gen)
			for i := range points {
				points[i].FromJacobian(&g)
				g.AddAssign(&g1Gen)
			}
			fillBenchScalars(scalars)
			if size > 7 {
				points[3].SetInfinity()
				scalars[5].SetZero()
			}

			var expected, got, gotWide, gotWideDense G1Jac
			if _, err := expected.MultiExp(points, scalars, ecc.MultiExpConfig{NbTasks: 2}); err != nil {
				t.Fatal(err)
			}
			if _, err := got.MultiExpGLV(points, scalars, ecc.MultiExpConfig{NbTasks: 2}); err != nil {
				t.Fatal(err)
			}
			if !expected.Equal(&got) {
				t.Fatalf("GLV MSM mismatch for %d points", size)
			}
			if _, err := gotWide.MultiExpGLVWide(points, scalars, ecc.MultiExpConfig{NbTasks: 2}); err != nil {
				t.Fatal(err)
			}
			if !expected.Equal(&gotWide) {
				t.Fatalf("wide GLV MSM mismatch for %d points", size)
			}
			if _, err := gotWideDense.MultiExpGLVWideDense(points, scalars, ecc.MultiExpConfig{NbTasks: 2}); err != nil {
				t.Fatal(err)
			}
			if !expected.Equal(&gotWideDense) {
				t.Fatalf("wide dense GLV MSM mismatch for %d points", size)
			}
		})
	}
}

func TestSplitScalarGLVG1Fixed(t *testing.T) {
	scalars := make([]fr.Element, 1024)
	fillBenchScalars(scalars)
	scalars = append(scalars, fr.Element{}, fr.NewElement(1))

	for i := range scalars {
		var got glvMSMG1Scalar
		splitScalarGLVG1Fixed(scalars[i].Bits(), &got)

		var scalar big.Int
		scalars[i].BigInt(&scalar)
		want := ecc.SplitScalar(&scalar, &glvBasis)

		gotK0 := reconstructGLVG1Component(got.k0, got.extra0)
		gotK1 := reconstructGLVG1Component(got.k1, got.extra1)
		if gotK0.Cmp(&want[0]) != 0 || gotK1.Cmp(&want[1]) != 0 {
			t.Fatalf("split mismatch at index %d", i)
		}

		var gotWide glvMSMG1WideScalar
		splitScalarGLVG1FixedWide(scalars[i].Bits(), &gotWide)
		gotWideK0 := reconstructGLVG1WideComponent(gotWide.k0, gotWide.extra0)
		gotWideK1 := reconstructGLVG1WideComponent(gotWide.k1, gotWide.extra1)
		if gotWideK0.Cmp(&want[0]) != 0 || gotWideK1.Cmp(&want[1]) != 0 {
			t.Fatalf("wide split mismatch at index %d", i)
		}
	}
}

func reconstructGLVG1Component(digits [glvMSMG1Chunks]int16, extra int8) big.Int {
	var res, term big.Int
	for i, digit := range digits {
		if digit == 0 {
			continue
		}
		term.SetInt64(int64(digit))
		term.Lsh(&term, uint(i*glvMSMG1Window))
		res.Add(&res, &term)
	}
	if extra != 0 {
		term.SetInt64(int64(extra))
		term.Lsh(&term, glvMSMG1Chunks*glvMSMG1Window)
		res.Add(&res, &term)
	}
	return res
}

func reconstructGLVG1WideComponent(digits [glvMSMG1WideChunks]int32, extra int8) big.Int {
	var res, term big.Int
	for i, digit := range digits {
		if digit == 0 {
			continue
		}
		term.SetInt64(int64(digit))
		term.Lsh(&term, uint(i*glvMSMG1WideWindow))
		res.Add(&res, &term)
	}
	if extra != 0 {
		term.SetInt64(int64(extra))
		term.Lsh(&term, glvMSMG1WideChunks*glvMSMG1WideWindow)
		res.Add(&res, &term)
	}
	return res
}

func BenchmarkMultiExpG1GLV(b *testing.B) {
	const (
		pow       = (bits.UintSize / 2) - (bits.UintSize / 8)
		nbSamples = 1 << pow
	)

	var (
		samplePoints  [nbSamples]G1Affine
		sampleScalars [nbSamples]fr.Element
	)

	fillBenchScalars(sampleScalars[:])
	fillBenchBasesG1(samplePoints[:])

	for i := 16; i <= pow; i += 2 {
		using := 1 << i

		b.Run(fmt.Sprintf("baseline/%d points", using), func(b *testing.B) {
			var testPoint G1Affine
			b.ResetTimer()
			for range b.N {
				testPoint.MultiExp(samplePoints[:using], sampleScalars[:using], ecc.MultiExpConfig{})
			}
		})

		b.Run(fmt.Sprintf("glv/%d points", using), func(b *testing.B) {
			var testPoint G1Affine
			b.ResetTimer()
			for range b.N {
				testPoint.MultiExpGLV(samplePoints[:using], sampleScalars[:using], ecc.MultiExpConfig{})
			}
		})

		b.Run(fmt.Sprintf("glv-wide/%d points", using), func(b *testing.B) {
			var testPoint G1Affine
			b.ResetTimer()
			for range b.N {
				testPoint.MultiExpGLVWide(samplePoints[:using], sampleScalars[:using], ecc.MultiExpConfig{})
			}
		})

		b.Run(fmt.Sprintf("glv-wide-dense/%d points", using), func(b *testing.B) {
			var testPoint G1Affine
			b.ResetTimer()
			for range b.N {
				testPoint.MultiExpGLVWideDense(samplePoints[:using], sampleScalars[:using], ecc.MultiExpConfig{})
			}
		})
	}
}
