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
	"math/bits"

	"github.com/consensys/gnark-crypto/ecc/bw6-756/fr"
)

// Expression represents a multivariate polynomial.
type Expression func(x ...fr.Element) fr.Element

// Evaluate evaluates f on each entry of x. The returned value is
// the vector of evaluations of e on x.
// The form of the result is form.
// The Size field of the result is the same as the one of x[0].
// The BlindedSize field of the result is the same as Size.
// The Shift field of the result is 0.
func Evaluate(f Expression, form Form, x ...*WrappedPolynomial) (WrappedPolynomial, error) {

	var res WrappedPolynomial

	// check that the sizes are consistent
	n := len(x[0].P.Coefficients)
	for i := 1; i < len(x); i++ {
		if n != len(x[i].P.Coefficients) {
			return res, ErrInconsistentSize
		}
	}

	r := make([]fr.Element, n)
	nbVariables := len(x)
	vx := make([]fr.Element, nbVariables)

	if form.Layout == Regular {
		for i := 0; i < n; i++ {
			for j := 0; j < nbVariables; j++ {
				vx[j] = x[j].GetCoeff(i)
			}
			r[i] = f(vx...)
		}
	} else {
		nn := uint64(64 - bits.TrailingZeros(uint(n)))
		for i := 0; i < n; i++ {
			for j := 0; j < nbVariables; j++ {
				vx[j] = x[j].GetCoeff(i)
			}
			iRev := bits.Reverse64(uint64(i)) >> nn
			r[iRev] = f(vx...)
		}
	}

	res.P = NewPolynomial(r, form)
	res.Size = x[0].Size
	res.BlindedSize = x[0].Size
	res.Shift = 0

	return res, nil
}
