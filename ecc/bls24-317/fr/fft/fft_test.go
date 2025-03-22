// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package fft

import (
	"math/big"
	"strconv"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bls24-317/fr"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

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

					pol := make([]fr.Element, maxSize)
					backupPol := make([]fr.Element, maxSize)

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

					pol := make([]fr.Element, maxSize)
					backupPol := make([]fr.Element, maxSize)

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

					pol := make([]fr.Element, maxSize)
					backupPol := make([]fr.Element, maxSize)

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

					pol := make([]fr.Element, maxSize)
					backupPol := make([]fr.Element, maxSize)

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

						pol := make([]fr.Element, maxSize)
						backupPol := make([]fr.Element, maxSize)

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

					pol := make([]fr.Element, maxSize)
					backupPol := make([]fr.Element, maxSize)

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

					pol := make([]fr.Element, maxSize)
					backupPol := make([]fr.Element, maxSize)

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

// --------------------------------------------------------------------
// benches

func BenchmarkFFT(b *testing.B) {

	const maxSize = 1 << 20

	pol := make([]fr.Element, maxSize)
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

	pol := make([]fr.Element, maxSize)
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

	pol := make([]fr.Element, maxSize)
	pol[0].SetRandom()
	for i := 1; i < maxSize; i++ {
		pol[i] = pol[i-1]
	}

	domain := NewDomain(maxSize)

	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		domain.FFT(pol, DIT)
	}
}

func BenchmarkFFTDIFReference(b *testing.B) {
	const maxSize = 1 << 20

	pol := make([]fr.Element, maxSize)
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

	pol := make([]fr.Element, maxSize)
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

func evaluatePolynomial(pol []fr.Element, val fr.Element) fr.Element {
	var acc, res, tmp fr.Element
	res.Set(&pol[0])
	acc.Set(&val)
	for i := 1; i < len(pol); i++ {
		tmp.Mul(&acc, &pol[i])
		res.Add(&res, &tmp)
		acc.Mul(&acc, &val)
	}
	return res
}
