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
// the same size and the size must be a power of 2.
// * challenge variable at which the numerator and denominators are evaluated
// * expectedForm expected form of the resulting polynomial
// * Return: say challenge=β, numerator = [P₁,...,P_m], denominator = [Q₁,..,Q_m]. The function
// returns a polynomial whose evaluation on the j-th root of unity is
// (Π_{k<j}Π_{i<m}(P_i(ω^k)-β))/(Q_i(ω^k)-β)
func BuildRatio(numerator, denominator []Polynomial, challenge fr.Element, expectedForm Form, domain *fft.Domain) (Polynomial, error) {

	var res Polynomial

	// check that len(numerator)=len(denominator)
	if len(numerator) != len(denominator) {
		return res, ErrNumberPolynomials
	}

	// check sizes between one another
	n := len(numerator[0].Coefficients)
	for i := 0; i < len(numerator); i++ {
		if len(numerator[i].Coefficients) != n {
			return res, ErrInconsistantSize
		}
		if len(denominator[i].Coefficients) != n {
			return res, ErrInconsistantSize
		}
	}

	// check if the sizes are a power of 2
	if n&(n-1) != 0 {
		return res, ErrSizeNotPowerOfTwo
	}

	// check if domain is of the correct size (if not we create it)
	if domain == nil {
		domain = fft.NewDomain(uint64(n))
	}
	if domain.Cardinality != uint64(n) {
		return res, ErrInconsistantSizeDomain
	}

	// put every polynomials in Lagrange from

	// build the ratio

	// put the resulting polynomial in the expected form

}
