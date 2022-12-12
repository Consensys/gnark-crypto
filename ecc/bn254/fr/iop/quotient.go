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
	"math/big"
	"math/bits"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/fft"
)

// ComputeQuotient returns h(f₁,..,fₙ)/Xⁿ-1 where n=len(f_i).
func ComputeQuotient(entries []*Polynomial, h multivariatePolynomial, expectedForm Form, domains [2]*fft.Domain) (Polynomial, error) {

	var quotientLagrangeCosetBitReverse Polynomial

	// check that the sizes are consistant
	err := checkSize(entries)
	if err != nil {
		return quotientLagrangeCosetBitReverse, nil
	}

	// create the domains for the individual polynomials + for the quotient
	sizeSmall := len(entries[0].Coefficients)
	domains[0], err = buildDomain(sizeSmall, domains[0])
	if err != nil {
		return quotientLagrangeCosetBitReverse, err
	}
	sizeBig := ecc.NextPowerOfTwo(h.degree() * domains[0].Cardinality)
	domains[1], err = buildDomain(int(sizeBig), domains[1])
	if err != nil {
		return quotientLagrangeCosetBitReverse, err
	}

	// put every polynomial in Canonical form. Also make sure
	// that we don't modify the slice entries, but
	// only its entries.
	nbPolynomials := len(entries)
	_entries := make([]*Polynomial, nbPolynomials)
	for i := 0; i < nbPolynomials; i++ {
		_entries[i] = toCanonical(entries[i], domains[0])
	}

	// compute h(f₁,..,fₙ) on a coset
	entriesLagrangeBigDomain := make([]Polynomial, nbPolynomials)
	for i := 0; i < nbPolynomials; i++ {
		entriesLagrangeBigDomain[i].Info.Basis = LagrangeCoset
		entriesLagrangeBigDomain[i].Info.Status = Unlocked
		entriesLagrangeBigDomain[i].Coefficients = make([]fr.Element, sizeBig)
		copy(entriesLagrangeBigDomain[i].Coefficients, _entries[i].Coefficients)
		if _entries[i].Info.Layout == Regular {
			entriesLagrangeBigDomain[i].Info.Layout = BitReverse
			domains[1].FFT(entriesLagrangeBigDomain[i].Coefficients, fft.DIF, true)
		} else {
			entriesLagrangeBigDomain[i].Info.Layout = Regular
			domains[1].FFT(entriesLagrangeBigDomain[i].Coefficients, fft.DIT, true)
		}
	}

	// prepare the evaluations of x^n-1 on the big domain's coset
	xnMinusOneInverseLagrangeCoset := evaluateXnMinusOneDomainBigCoset(domains[1], domains[0])
	ratio := int(domains[1].Cardinality / domains[0].Cardinality)

	// compute the division. We take care of the indices of the
	// polnyomials which are bit reversed.
	// The result is temporarily stored in bit reversed Lagrange form,
	// before it is actually transformed into the expected format.
	nn := uint64(64 - bits.TrailingZeros(uint(sizeBig)))

	nbVars := len(entries)
	x := make([]fr.Element, nbVars)

	quotientLagrangeCosetBitReverse.Coefficients = make([]fr.Element, sizeBig)

	for i := 0; i < int(sizeBig); i++ {

		iRev := bits.Reverse64(uint64(i)) >> nn

		for j := 0; j < nbVars; j++ {

			// set the variable
			if entriesLagrangeBigDomain[j].Info.Layout == Regular {
				x[j].Set(&entriesLagrangeBigDomain[j].Coefficients[j])
			} else {
				x[j].Set(&entriesLagrangeBigDomain[j].Coefficients[iRev])
			}

		}

		// evaluate h on x
		quotientLagrangeCosetBitReverse.Coefficients[iRev] = h.evaluate(x)

		// divide by x^n-1 evaluated on the correct point.
		quotientLagrangeCosetBitReverse.Coefficients[iRev].
			Div(&quotientLagrangeCosetBitReverse.Coefficients[iRev], &xnMinusOneInverseLagrangeCoset[i%ratio])
	}

	// at this stage the result is in Lagrange, bitreversed format.
	// We put it in the expected format.
	putInExpectedFormFromLagrangeCosetBitReversed(&quotientLagrangeCosetBitReverse, domains[1], expectedForm)

	return quotientLagrangeCosetBitReverse, nil
}

// evaluateXnMinusOneDomainBigCoset evalutes Xᵐ-1 on DomainBig coset
func evaluateXnMinusOneDomainBigCoset(domainBig, domainSmall *fft.Domain) []fr.Element {

	ratio := domainBig.Cardinality / domainSmall.Cardinality

	res := make([]fr.Element, ratio)

	expo := big.NewInt(int64(domainSmall.Cardinality))
	res[0].Exp(domainBig.FrMultiplicativeGen, expo)

	var t fr.Element
	t.Exp(domainBig.Generator, big.NewInt(int64(domainSmall.Cardinality)))

	for i := 1; i < int(ratio); i++ {
		res[i].Mul(&res[i-1], &t)
	}

	var one fr.Element
	one.SetOne()
	for i := 0; i < int(ratio); i++ {
		res[i].Sub(&res[i], &one)
	}

	fr.BatchInvert(res)

	return res
}

func putInExpectedFormFromLagrangeCosetBitReversed(p *Polynomial, domain *fft.Domain, expectedForm Form) {

	p.Info = expectedForm

	if expectedForm.Basis == Canonical {
		domain.FFTInverse(p.Coefficients, fft.DIT, true)
		if expectedForm.Layout == BitReverse {
			fft.BitReverse(p.Coefficients)
		}
		return
	}

	if expectedForm.Basis == Lagrange {
		domain.FFTInverse(p.Coefficients, fft.DIT, true)
		domain.FFT(p.Coefficients, fft.DIF)
		if expectedForm.Layout == Regular {
			fft.BitReverse(p.Coefficients)
		}
		return
	}

	if expectedForm.Layout == Regular {
		fft.BitReverse(p.Coefficients)
	}

}
