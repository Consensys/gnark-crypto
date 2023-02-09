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

// Code generated by consensys/gnark-crypto DO NOT EDIT

package iop

import (
	"math/big"
	"math/bits"

	"github.com/consensys/gnark-crypto/ecc/bls24-315/fr"
	"github.com/consensys/gnark-crypto/ecc/bls24-315/fr/fft"
)

// DivideByXMinusOne
// The input must be in LagrangeCoset.
// The result is in Canonical Regular.
func DivideByXMinusOne(a WrappedPolynomial, domains [2]*fft.Domain) (*WrappedPolynomial, error) {

	// check that the basis is LagrangeCoset
	if a.Basis != LagrangeCoset {
		return nil, ErrMustBeLagrangeCoset
	}

	// prepare the evaluations of x^n-1 on the big domain's coset
	xnMinusOneInverseLagrangeCoset := evaluateXnMinusOneDomainBigCoset(domains)

	rho := len(a.Coefficients) / a.size

	nbElmts := len(a.Coefficients)

	var p Polynomial
	p.Form = Form{Layout: BitReverse, Basis: LagrangeCoset}
	p.Coefficients = make([]fr.Element, len(a.Coefficients))
	res := NewWrappedPolynomial(&p)
	res.size = a.size
	res.blindedSize = a.blindedSize

	nn := uint64(64 - bits.TrailingZeros(uint(nbElmts)))
	for i := 0; i < len(a.Coefficients); i++ {

		iRev := bits.Reverse64(uint64(i)) >> nn
		c := a.GetCoeff(i)
		res.Coefficients[iRev].
			Mul(&c, &xnMinusOneInverseLagrangeCoset[i%rho])
	}

	res.ToCanonical(domains[1])

	return res, nil

}

// evaluateXnMinusOneDomainBigCoset evalutes Xᵐ-1 on DomainBig coset
func evaluateXnMinusOneDomainBigCoset(domains [2]*fft.Domain) []fr.Element {

	ratio := domains[1].Cardinality / domains[0].Cardinality

	res := make([]fr.Element, ratio)

	expo := big.NewInt(int64(domains[0].Cardinality))
	res[0].Exp(domains[1].FrMultiplicativeGen, expo)

	var t fr.Element
	t.Exp(domains[1].Generator, big.NewInt(int64(domains[0].Cardinality)))

	for i := 1; i < int(ratio); i++ {
		res[i].Mul(&res[i-1], &t)
	}

	var one fr.Element
	one.SetOne()
	for i := 0; i < int(ratio); i++ {
		res[i].Sub(&res[i], &one)
	}

	res = fr.BatchInvert(res)

	return res
}
