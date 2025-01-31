// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package fft

import (
	"math/big"
	"strconv"
	"testing"

	"github.com/consensys/gnark-crypto/field/babybear"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"encoding/binary"
	"fmt"
)

func TestFFT(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 6
	properties := gopter.NewProperties(parameters)

	for maxSize := 2; maxSize <= 1<<10; maxSize <<= 1 {

		domainWithPrecompute := NewDomain(uint64(maxSize))
		domainWithoutPrecompute := NewDomain(uint64(maxSize), WithoutPrecompute())

		for domainName, domain := range map[string]*Domain{
			"with precompute":    domainWithPrecompute,
			"without precompute": domainWithoutPrecompute,
		} {
			domainName := domainName
			domain := domain
			t.Logf("domain: %s", domainName)
			properties.Property("DIF FFT should be consistent with dual basis", prop.ForAll(

				// checks that a random evaluation of a dual function eval(gen**ithpower) is consistent with the FFT result
				func(ithpower int) bool {

					pol := make([]babybear.Element, maxSize)
					backupPol := make([]babybear.Element, maxSize)

					for i := 0; i < maxSize; i++ {
						pol[i].SetRandom()
					}
					copy(backupPol, pol)

					domain.FFT(pol, DIF)
					BitReverse(pol)

					sample := domain.Generator
					sample.Exp(sample, big.NewInt(int64(ithpower)))

					eval := evaluatePolynomial(backupPol, sample)

					return eval.Equal(&pol[ithpower])

				},
				gen.IntRange(0, maxSize-1),
			))

			properties.Property("DIF FFT on cosets should be consistent with dual basis", prop.ForAll(

				// checks that a random evaluation of a dual function eval(gen**ithpower) is consistent with the FFT result
				func(ithpower int) bool {

					pol := make([]babybear.Element, maxSize)
					backupPol := make([]babybear.Element, maxSize)

					for i := 0; i < maxSize; i++ {
						pol[i].SetRandom()
					}
					copy(backupPol, pol)

					domain.FFT(pol, DIF, OnCoset())
					BitReverse(pol)

					sample := domain.Generator
					sample.Exp(sample, big.NewInt(int64(ithpower))).
						Mul(&sample, &domain.FrMultiplicativeGen)

					eval := evaluatePolynomial(backupPol, sample)

					return eval.Equal(&pol[ithpower])

				},
				gen.IntRange(0, maxSize-1),
			))

			properties.Property("DIT FFT should be consistent with dual basis", prop.ForAll(

				// checks that a random evaluation of a dual function eval(gen**ithpower) is consistent with the FFT result
				func(ithpower int) bool {

					pol := make([]babybear.Element, maxSize)
					backupPol := make([]babybear.Element, maxSize)

					for i := 0; i < maxSize; i++ {
						pol[i].SetRandom()
					}
					copy(backupPol, pol)

					BitReverse(pol)
					domain.FFT(pol, DIT)

					sample := domain.Generator
					sample.Exp(sample, big.NewInt(int64(ithpower)))

					eval := evaluatePolynomial(backupPol, sample)

					return eval.Equal(&pol[ithpower])

				},
				gen.IntRange(0, maxSize-1),
			))

			properties.Property("bitReverse(DIF FFT(DIT FFT (bitReverse))))==id", prop.ForAll(

				func() bool {

					pol := make([]babybear.Element, maxSize)
					backupPol := make([]babybear.Element, maxSize)

					for i := 0; i < maxSize; i++ {
						pol[i].SetRandom()
					}
					copy(backupPol, pol)

					BitReverse(pol)
					domain.FFT(pol, DIT)
					domain.FFTInverse(pol, DIF)
					BitReverse(pol)

					check := true
					for i := 0; i < len(pol); i++ {
						check = check && pol[i].Equal(&backupPol[i])
					}
					return check
				},
			))

			for nbCosets := 2; nbCosets < 5; nbCosets++ {
				properties.Property(fmt.Sprintf("bitReverse(DIF FFT(DIT FFT (bitReverse))))==id on %d cosets", nbCosets), prop.ForAll(

					func() bool {

						pol := make([]babybear.Element, maxSize)
						backupPol := make([]babybear.Element, maxSize)

						for i := 0; i < maxSize; i++ {
							pol[i].SetRandom()
						}
						copy(backupPol, pol)

						check := true

						for i := 1; i <= nbCosets; i++ {

							BitReverse(pol)
							domain.FFT(pol, DIT, OnCoset())
							domain.FFTInverse(pol, DIF, OnCoset())
							BitReverse(pol)

							for i := 0; i < len(pol); i++ {
								check = check && pol[i].Equal(&backupPol[i])
							}
						}

						return check
					},
				))
			}

			properties.Property("DIT FFT(DIF FFT)==id", prop.ForAll(

				func() bool {

					pol := make([]babybear.Element, maxSize)
					backupPol := make([]babybear.Element, maxSize)

					for i := 0; i < maxSize; i++ {
						pol[i].SetRandom()
					}
					copy(backupPol, pol)

					domain.FFTInverse(pol, DIF)
					domain.FFT(pol, DIT)

					check := true
					for i := 0; i < len(pol); i++ {
						check = check && (pol[i] == backupPol[i])
					}
					return check
				},
			))

			properties.Property("DIT FFT(DIF FFT)==id on cosets", prop.ForAll(

				func() bool {

					pol := make([]babybear.Element, maxSize)
					backupPol := make([]babybear.Element, maxSize)

					for i := 0; i < maxSize; i++ {
						pol[i].SetRandom()
					}
					copy(backupPol, pol)

					domain.FFTInverse(pol, DIF, OnCoset())
					domain.FFT(pol, DIT, OnCoset())

					for i := 0; i < len(pol); i++ {
						if !(pol[i].Equal(&backupPol[i])) {
							return false
						}
					}

					// compute with nbTasks == 1
					domain.FFTInverse(pol, DIF, OnCoset(), WithNbTasks(1))
					domain.FFT(pol, DIT, OnCoset(), WithNbTasks(1))

					for i := 0; i < len(pol); i++ {
						if !(pol[i].Equal(&backupPol[i])) {
							return false
						}
					}

					return true
				},
			))
		}
		properties.TestingRun(t, gopter.ConsoleReporter(false))
	}

}
func FuzzFFTAvx512(f *testing.F) {
	if !supportAVX512 {
		f.Skip("AVX512 not supported")
	}

	domain := NewDomain(512)

	q := babybear.Modulus()
	qUuint32 := uint32(q.Uint64())
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 512*babybear.Bytes {
			t.Skip("not enough data")
		}

		var a0, a1, a2, a3 [512]babybear.Element

		for i := range a0 {
			a0[i][0] = binary.LittleEndian.Uint32(data[i*babybear.Bytes:])
			a0[i][0] %= qUuint32
		}

		copy(a1[:], a0[:])
		copy(a2[:], a0[:])
		copy(a3[:], a0[:])

		// check that the AVX512 and generic implementations match for innerDIFWithTwiddles
		innerDIFWithTwiddles(a0[:], domain.twiddles[0], 0, 256, 256)
		innerDIFWithTwiddlesGeneric(a1[:], domain.twiddles[0], 0, 256, 256)

		for i := range a0 {
			if !a0[i].Equal(&a1[i]) {
				t.Fatalf("innerDIFWithTwiddles mismatch at index %d: got %v, want %v", i, a0[i], a1[i])
			}
		}

		// do the same thing with the kernel of size 256
		kerDIFNP_256generic(a2[:], domain.twiddles, 1)
		kerDIFNP_256(a3[:], domain.twiddles, 1)

		for i := range a2 {
			if !a2[i].Equal(&a3[i]) {
				t.Fatalf("kerDIFNP_256 mismatch at index %d: got %v, want %v", i, a2[i], a3[i])
			}
		}
	})
}

// --------------------------------------------------------------------
// benches

func BenchmarkFFT(b *testing.B) {

	const maxSize = 1 << 20

	pol := make([]babybear.Element, maxSize)
	pol[0].SetRandom()
	for i := 1; i < maxSize; i++ {
		pol[i] = pol[i-1]
	}

	for i := 8; i < 20; i++ {
		sizeDomain := 1 << i
		b.Run("fft 2**"+strconv.Itoa(i)+"bits", func(b *testing.B) {
			domain := NewDomain(uint64(sizeDomain))
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				domain.FFT(pol[:sizeDomain], DIT)
			}
		})
		b.Run("fft 2**"+strconv.Itoa(i)+"bits (coset)", func(b *testing.B) {
			domain := NewDomain(uint64(sizeDomain))
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				domain.FFT(pol[:sizeDomain], DIT, OnCoset())
			}
		})
	}

}

func BenchmarkFFTDITCosetReference(b *testing.B) {
	const maxSize = 1 << 20

	pol := make([]babybear.Element, maxSize)
	pol[0].SetRandom()
	for i := 1; i < maxSize; i++ {
		pol[i] = pol[i-1]
	}

	domain := NewDomain(maxSize)

	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		domain.FFT(pol, DIT, OnCoset())
	}
}

func BenchmarkFFTDITReferenceSmall(b *testing.B) {
	const maxSize = 1 << 9

	pol := make([]babybear.Element, maxSize)
	pol[0].SetRandom()
	for i := 1; i < maxSize; i++ {
		pol[i] = pol[i-1]
	}

	domain := NewDomain(maxSize)

	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		domain.FFT(pol, DIT, OnCoset())
	}
}

func BenchmarkFFTDIFReference(b *testing.B) {
	const maxSize = 1 << 20

	pol := make([]babybear.Element, maxSize)
	pol[0].SetRandom()
	for i := 1; i < maxSize; i++ {
		pol[i] = pol[i-1]
	}

	domain := NewDomain(maxSize)

	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		domain.FFT(pol, DIF)
	}
}

func BenchmarkFFTDIFReferenceSmall(b *testing.B) {
	const maxSize = 1 << 9

	pol := make([]babybear.Element, maxSize)
	pol[0].SetRandom()
	for i := 1; i < maxSize; i++ {
		pol[i] = pol[i-1]
	}

	domain := NewDomain(maxSize)

	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		domain.FFT(pol, DIF)
	}
}

func evaluatePolynomial(pol []babybear.Element, val babybear.Element) babybear.Element {
	var acc, res, tmp babybear.Element
	res.Set(&pol[0])
	acc.Set(&val)
	for i := 1; i < len(pol); i++ {
		tmp.Mul(&acc, &pol[i])
		res.Add(&res, &tmp)
		acc.Mul(&acc, &val)
	}
	return res
}
