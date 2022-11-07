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
	"errors"
	"math/bits"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/fft"
)

var (
	ErrInconsistantSize       = errors.New("the sizes of the polynomial must be the same as the size of the domain")
	ErrNumberPolynomials      = errors.New("the number of polynomials in the denominator and the numerator must be the same")
	ErrSizeNotPowerOfTwo      = errors.New("the size of the polynomials must be a power of two")
	ErrInconsistantSizeDomain = errors.New("the size of the domain must be consistant with the size of the polynomials")
)

// Build an 'accumulating ratio' polynomial.
// * numerator list of polynomials that will form the numerator of the ratio
// * denominator list of polynomials that will form the denominator of the ratio
// The polynomials in the denominator and the numerator are expected to be of
// the same size and the size must be a power of 2. The polynomials are given as
// pointers in case the caller wants to FFTInv the polynomials during the process.
// * beta variable at which the numerator and denominators are evaluated
// * expectedForm expected form of the resulting polynomial
// * Return: say beta=β, numerator = [P₁,...,P_m], denominator = [Q₁,..,Q_m]. The function
// returns a polynomial whose evaluation on the j-th root of unity is
// (Π_{k<j}Π_{i<m}(\beta-P_i(ω^k)))/(\beta-Q_i(ω^k))
func BuildRatio(numerator, denominator []*Polynomial, beta fr.Element, expectedForm Form, domain *fft.Domain) (Polynomial, error) {

	var res Polynomial

	// check that len(numerator)=len(denominator)
	if len(numerator) != len(denominator) {
		return res, ErrNumberPolynomials
	}
	nbPolynomials := len(numerator)

	// create the domain + some checks on the sizes of the polynomials
	domain, err := buildDomain(numerator, denominator, domain)
	if err != nil {
		return res, err
	}

	// put every polynomials in Lagrange form
	for i := 0; i < nbPolynomials; i++ {
		numerator[i] = toLagrange[getFootPrint(*numerator[i])](numerator[i], domain)
		denominator[i] = toLagrange[getFootPrint(*denominator[i])](denominator[i], domain)
	}

	// build the ratio (careful with the indices of
	// the polynomials which are bit reversed)
	n := len(numerator[0].Coefficients)
	res.Coefficients = make([]fr.Element, n)
	t := make([]fr.Element, n)
	res.Coefficients[0].SetOne()
	t[0].SetOne()
	var a, b, c, d fr.Element

	nn := uint64(64 - bits.TrailingZeros(uint(n)))
	for i := 1; i < n; i++ {

		b.SetOne()
		c.SetOne()

		iMinusOnerev := bits.Reverse64(uint64(i-1)) >> nn

		for j := 0; j < nbPolynomials; j++ {

			if numerator[j].Info.Layout == BitReverse {
				a.Sub(&beta, &numerator[j].Coefficients[iMinusOnerev])
			} else {
				a.Sub(&beta, &numerator[j].Coefficients[i-1])
			}
			b.Mul(&b, &a)

			if denominator[j].Info.Layout == BitReverse {
				c.Sub(&beta, &denominator[j].Coefficients[iMinusOnerev])
			} else {
				c.Sub(&beta, &denominator[j].Coefficients[i-1])
			}
			d.Mul(&d, &c)
		}
		// b = Πₖ (β-P_k(ωⁱ⁻¹))
		// d = Πₖ (β-Q_k(ωⁱ⁻¹))

		res.Coefficients[i].Mul(&res.Coefficients[i-1], &b)
		t[i].Mul(&t[i-1], &d)

	}
	t = fr.BatchInvert(t)
	for i := 1; i < n; i++ {
		res.Coefficients[i].Mul(&res.Coefficients[i], &t[i])
	}

	res.Info = expectedForm

	// at this stage the result is in Lagrange form, Regular layout
	if expectedForm.Basis == Canonical {
		domain.FFTInverse(res.Coefficients, fft.DIF)
		if expectedForm.Layout == Regular {
			fft.BitReverse(res.Coefficients)
		}
		return res, nil
	}

	if expectedForm.Basis == LagrangeCoset {
		domain.FFTInverse(res.Coefficients, fft.DIF)
		domain.FFT(res.Coefficients, fft.DIT, true)
		if expectedForm.Layout == BitReverse {
			fft.BitReverse(res.Coefficients)
		}
		return res, nil
	}

	if expectedForm.Layout == BitReverse {
		fft.BitReverse(res.Coefficients)
	}
	return res, nil
}

func buildDomain(numerator, denominator []*Polynomial, domain *fft.Domain) (*fft.Domain, error) {

	// check sizes between one another
	n := len(numerator[0].Coefficients)
	for i := 0; i < len(numerator); i++ {
		if len(numerator[i].Coefficients) != n {
			return nil, ErrInconsistantSize
		}
		if len(denominator[i].Coefficients) != n {
			return nil, ErrInconsistantSize
		}
	}

	// check if the sizes are a power of 2
	if n&(n-1) != 0 {
		return nil, ErrSizeNotPowerOfTwo
	}

	// check if domain is of the correct size (if not we create it)
	if domain == nil {
		domain = fft.NewDomain(uint64(n))
	}

	// in case domain was not nil, it must match the size of the polynomials.
	if domain.Cardinality != uint64(n) {
		return nil, ErrInconsistantSizeDomain
	}

	return domain, nil
}

func BuildRatioWithPermutation(
	numerator, denominator []*Polynomial,
	beta, gamma fr.Element,
	permutation []int64,
	expectedForm Form,
	domain *fft.Domain) (Polynomial, error) {

	var a Polynomial
	return a, nil

}

// getSupportIdentityPermutation returns the support on which the permutation acts.
// Concrectly it's X evaluated on
// [1,ω,..,ωˢ⁻¹,g,g*ω,..,g*ωˢ⁻¹,..,gⁿ⁻¹,gⁿ⁻¹*ω,..,gⁿ⁻¹*ωˢ⁻¹]
func getSupportIdentityPermutation(n int, domain *fft.Domain) []fr.Element {

	res := make([]fr.Element, uint64(n)*domain.Cardinality)

	res[0].SetOne()
	for i := 1; i < n; i++ {
		res[domain.Cardinality].Mul(&res[uint64(i-1)*domain.Cardinality], &domain.FrMultiplicativeGen)
	}

	for i := uint64(1); i < domain.Cardinality; i++ {
		for j := 0; j < n; j++ {
			res[i+uint64(j)*domain.Cardinality].Mul(&res[i-1], &domain.Generator)
		}
	}

	return res
}
