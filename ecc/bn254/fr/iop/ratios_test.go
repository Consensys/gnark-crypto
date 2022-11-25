// Copyright 2020 ConsenSys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package iop

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/fft"
)

// getPermutation returns a deterministic permutation
// of n elements where n is even. The result should be
// interpreted as
// a permutation σ(i)=permutation[i]
// g is a generator of ℤ/nℤ
func getPermutation(n, g int) []int {

	res := make([]int, n)
	a := g
	for i := 0; i < n; i++ {
		res[i] = a
		a += g
		a %= n
	}
	return res
}

func getPermutedPolynomials(sizePolynomials, nbPolynomials int) ([]*Polynomial, []*Polynomial, []int) {

	numerator := make([]*Polynomial, nbPolynomials)
	for i := 0; i < nbPolynomials; i++ {
		numerator[i] = new(Polynomial)
		numerator[i].Coefficients = randomVector(sizePolynomials)
		numerator[i].Info.Basis = Lagrange
		numerator[i].Info.Layout = Regular
		numerator[i].Info.Status = Locked
	}

	// get permutation
	sigma := getPermutation(sizePolynomials*nbPolynomials, 3)

	// the denominator is the permuted version of the numerators
	// concatenated
	denominator := make([]*Polynomial, nbPolynomials)
	for i := 0; i < nbPolynomials; i++ {
		denominator[i] = new(Polynomial)
		denominator[i].Coefficients = make([]fr.Element, sizePolynomials)
		denominator[i].Info.Basis = Lagrange
		denominator[i].Info.Layout = Regular
		denominator[i].Info.Status = Locked
	}
	for i := 0; i < len(sigma); i++ {
		id := int(sigma[i] / sizePolynomials)
		od := sigma[i] % sizePolynomials
		in := int(i / sizePolynomials)
		on := i % sizePolynomials
		denominator[in].Coefficients[on].Set(&numerator[id].Coefficients[od])
	}

	return numerator, denominator, sigma

}

func TestBuildRatioShuffledVectors(t *testing.T) {

	// generate random vectors, interpreted in Lagrange form,
	// regular layout. It is enough for this test if TestPutInLagrangeForm
	// passes.
	sizePolynomials := 8
	nbPolynomials := 4
	numerator, denominator, _ := getPermutedPolynomials(sizePolynomials, nbPolynomials)

	// build the ratio polynomial
	expectedForm := Form{Basis: Lagrange, Layout: Regular, Status: Unlocked}
	domain := fft.NewDomain(uint64(sizePolynomials))
	var beta fr.Element
	beta.SetRandom()
	ratio, err := BuildRatioShuffledVectors(numerator, denominator, beta, expectedForm, domain)
	if err != nil {
		t.Fatal()
	}

	// check that the whole product is equal to one
	var a, b, c, d fr.Element
	b.SetOne()
	d.SetOne()
	for i := 0; i < nbPolynomials; i++ {
		a.Sub(&beta, &numerator[i].Coefficients[sizePolynomials-1])
		b.Mul(&a, &b)
		c.Sub(&beta, &denominator[i].Coefficients[sizePolynomials-1])
		d.Mul(&c, &d)
	}
	a.Mul(&b, &ratio.Coefficients[sizePolynomials-1]).
		Div(&a, &d)
	var one fr.Element
	one.SetOne()
	if !a.Equal(&one) {
		t.Fatal("accumulating ratio is not equal to one")
	}

}

func TestBuildRatioSpecificPermutation(t *testing.T) {

	// generate random vectors, interpreted in Lagrange form,
	// regular layout. It is enough for this test if TestPutInLagrangeForm
	// passes.
	sizePolynomials := 8
	nbPolynomials := 4
	numerator, denominator, sigma := getPermutedPolynomials(sizePolynomials, nbPolynomials)

	// build the ratio polynomial
	expectedForm := Form{Basis: Lagrange, Layout: Regular, Status: Unlocked}
	domain := fft.NewDomain(uint64(sizePolynomials))
	var beta, gamma fr.Element
	beta.SetRandom()
	gamma.SetRandom()
	_, err := BuildRatioSpecificPermutation(numerator, denominator, sigma, beta, gamma, expectedForm, domain)
	if err != nil {
		t.Fatal()
	}

}
