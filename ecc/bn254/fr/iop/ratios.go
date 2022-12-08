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
	"math/bits"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/fft"
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
// (Π_{k<j}Π_{i<m}(β-Pᵢ(ωᵏ)))/(β-Qᵢ(ωᵏ))
func BuildRatioShuffledVectors(numerator, denominator []*Polynomial, beta fr.Element, expectedForm Form, domain *fft.Domain) (Polynomial, error) {

	var res Polynomial

	// check that len(numerator)=len(denominator)
	if len(numerator) != len(denominator) {
		return res, ErrNumberPolynomials
	}
	nbPolynomials := len(numerator)

	// check that the sizes are consistant
	err := checkSize(numerator, denominator)
	if err != nil {
		return res, err
	}

	// create the domain + some checks on the sizes of the polynomials
	n := len(numerator[0].Coefficients)
	domain, err = buildDomain(n, domain)
	if err != nil {
		return res, err
	}

	// put every polynomials in Lagrange form. Also make sure
	// that we don't modify the slices numerator and denominator, but
	// only their entries.
	_numerator := make([]*Polynomial, nbPolynomials)
	_denominator := make([]*Polynomial, nbPolynomials)
	for i := 0; i < nbPolynomials; i++ {
		_numerator[i] = toLagrange(numerator[i], domain)
		_denominator[i] = toLagrange(denominator[i], domain)
	}

	// build the ratio (careful with the indices of
	// the polynomials which are bit reversed)
	res.Coefficients = make([]fr.Element, n)
	t := make([]fr.Element, n)
	res.Coefficients[0].SetOne()
	t[0].SetOne()
	var a, b, c, d fr.Element

	nn := uint64(64 - bits.TrailingZeros(uint(n)))
	for i := 0; i < n-1; i++ {

		b.SetOne()
		d.SetOne()

		iRev := bits.Reverse64(uint64(i)) >> nn

		for j := 0; j < nbPolynomials; j++ {

			if _numerator[j].Info.Layout == BitReverse {
				a.Sub(&beta, &_numerator[j].Coefficients[iRev])
			} else {
				a.Sub(&beta, &_numerator[j].Coefficients[i])
			}
			b.Mul(&b, &a)

			if _denominator[j].Info.Layout == BitReverse {
				c.Sub(&beta, &_denominator[j].Coefficients[iRev])
			} else {
				c.Sub(&beta, &_denominator[j].Coefficients[i])
			}
			d.Mul(&d, &c)
		}
		// b = Πₖ (β-Pₖ(ωⁱ⁻¹))
		// d = Πₖ (β-Qₖ(ωⁱ⁻¹))

		res.Coefficients[i+1].Mul(&res.Coefficients[i], &b)
		t[i+1].Mul(&t[i], &d)

	}

	t = fr.BatchInvert(t)
	for i := 1; i < n; i++ {
		res.Coefficients[i].Mul(&res.Coefficients[i], &t[i])
	}

	res.Info = expectedForm

	// at this stage the result is in Lagrange form, Regular layout
	putInExpectedFormFromLagrangeRegular(&res, domain, expectedForm)

	return res, nil
}

// BuildRatioSpecificPermutation builds the accumulating ratio polynomial to prove that
// [P₁ ∥ .. ∥ P_{n—] = σ([Q₁ ∥ .. ∥ Qₙ]).
// Namely it returns the polynomial Z whose evaluation on the j-th root of unity is
// Z(ω^j) = Π_{i<j}(Π_{k<n}(P_k(ω^i)+β*u^k+γ))/(Q_k(ω^i)+σ(kn+i)+γ)))
// * numerator list of polynomials that will form the numerator of the ratio
// * denominator list of polynomials that will form the denominator of the ratio
// The polynomials in the denominator and the numerator are expected to be of
// the same size and the size must be a power of 2. The polynomials are given as
// pointers in case the caller wants to FFTInv the polynomials during the process.
// * beta, gamma challenges
// * expectedForm expected form of the resulting polynomial
func BuildRatioSpecificPermutation(
	numerator, denominator []*Polynomial,
	permutation []int,
	beta, gamma fr.Element,
	expectedForm Form,
	domain *fft.Domain) (Polynomial, error) {

	var res Polynomial

	// check that len(numerator)=len(denominator)
	if len(numerator) != len(denominator) {
		return res, ErrNumberPolynomials
	}
	nbPolynomials := len(numerator)

	// check that the sizes are consistant
	err := checkSize(numerator, denominator)
	if err != nil {
		return res, err
	}

	// create the domain + some checks on the sizes of the polynomials
	n := len(numerator[0].Coefficients)
	domain, err = buildDomain(n, domain)
	if err != nil {
		return res, err
	}

	// put every polynomials in Lagrange form. Also make sure
	// that we don't modify the slices numerator and denominator, but
	// only their entries.
	_numerator := make([]*Polynomial, nbPolynomials)
	_denominator := make([]*Polynomial, nbPolynomials)
	for i := 0; i < nbPolynomials; i++ {
		_numerator[i] = toLagrange(numerator[i], domain)
		_denominator[i] = toLagrange(denominator[i], domain)
	}

	// get the support for the permutation
	evaluationIDSmallDomain := getSupportIdentityPermutation(nbPolynomials, domain)

	// build the ratio (careful with the indices of
	// the polynomials which are bit reversed)
	res.Coefficients = make([]fr.Element, n)
	t := make([]fr.Element, n)
	res.Coefficients[0].SetOne()
	t[0].SetOne()
	var a, b, c, d fr.Element

	nn := uint64(64 - bits.TrailingZeros(uint(n)))
	for i := 0; i < n-1; i++ {

		b.SetOne()
		d.SetOne()

		iRev := bits.Reverse64(uint64(i)) >> nn

		for j := 0; j < nbPolynomials; j++ {

			a.Mul(&beta, &evaluationIDSmallDomain[i+j*n]).
				Add(&a, &gamma)
			if _numerator[j].Info.Layout == BitReverse {
				a.Add(&a, &_numerator[j].Coefficients[iRev])
			} else {
				a.Add(&a, &_numerator[j].Coefficients[i])
			}
			b.Mul(&a, &b)

			c.Mul(&beta, &evaluationIDSmallDomain[permutation[i+j*n]]).
				Add(&c, &gamma)
			if _denominator[j].Info.Layout == BitReverse {
				c.Add(&c, &_denominator[j].Coefficients[iRev])
			} else {
				c.Add(&c, &_denominator[j].Coefficients[i])
			}
			d.Mul(&d, &c)
		}

		// b = Πⱼ(Pⱼ(ωⁱ)+β*ωⁱνʲ+γ)
		// d = Πⱼ(Qⱼ(ωⁱ)+β*σ(j*n+i)+γ)

		res.Coefficients[i+1].Mul(&res.Coefficients[i], &b)
		t[i+1].Mul(&t[i], &d)
	}

	t = fr.BatchInvert(t)
	for i := 1; i < n; i++ {
		res.Coefficients[i].Mul(&res.Coefficients[i], &t[i])
	}

	// at this stage the result is in Lagrange form, Regular layout
	putInExpectedFormFromLagrangeRegular(&res, domain, expectedForm)

	return res, nil

}

func putInExpectedFormFromLagrangeRegular(p *Polynomial, domain *fft.Domain, expectedForm Form) {

	p.Info = expectedForm

	if expectedForm.Basis == Canonical {
		domain.FFTInverse(p.Coefficients, fft.DIF)
		if expectedForm.Layout == Regular {
			fft.BitReverse(p.Coefficients)
		}
		return
	}

	if expectedForm.Basis == LagrangeCoset {
		domain.FFTInverse(p.Coefficients, fft.DIF)
		domain.FFT(p.Coefficients, fft.DIT, true)
		if expectedForm.Layout == BitReverse {
			fft.BitReverse(p.Coefficients)
		}
		return
	}

	if expectedForm.Layout == BitReverse {
		fft.BitReverse(p.Coefficients)
	}

}

// check that the polynomials are of the same size.
// It assumes that pols contains slices of the same size.
func checkSize(pols ...[]*Polynomial) error {

	// check sizes between one another
	m := len(pols)
	n := len(pols[0][0].Coefficients)
	for i := 0; i < m; i++ {
		for j := 0; j < len(pols); j++ {
			if len(pols[i][j].Coefficients) != n {
				return ErrInconsistantSize
			}
		}
	}

	return nil
}

// buildDomain builds the fft domain necessary to do FFTs.
// n is the cardinality of the domain, it must be a power of 2.
func buildDomain(n int, domain *fft.Domain) (*fft.Domain, error) {

	// check if the sizes are a power of 2
	if n&(n-1) != 0 {
		return nil, ErrSizeNotPowerOfTwo
	}

	// if the domain doesn't exist we create it.
	if domain == nil {
		domain = fft.NewDomain(uint64(n))
	}

	// in case domain was not nil, it must match the size of the polynomials.
	if domain.Cardinality != uint64(n) {
		return nil, ErrInconsistantSizeDomain
	}

	return domain, nil
}

func BuildRatioShuffledVectorsWithPermutation(
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
// nbCopies is the number of cosets of the roots of unity that are needed, including the set of
// roots of unity itself.
func getSupportIdentityPermutation(nbCopies int, domain *fft.Domain) []fr.Element {

	res := make([]fr.Element, uint64(nbCopies)*domain.Cardinality)
	sizePoly := int(domain.Cardinality)

	res[0].SetOne()
	for i := 0; i < sizePoly-1; i++ {
		res[i+1].Mul(&res[i], &domain.Generator)
	}
	for i := 1; i < nbCopies; i++ {
		copy(res[i*sizePoly:], res[(i-1)*sizePoly:i*int(domain.Cardinality)])
		for j := 0; j < sizePoly; j++ {
			res[i*sizePoly+j].Mul(&res[i*sizePoly+j], &domain.FrMultiplicativeGen)
		}
	}

	return res
}
