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
)

func TestEvaluate(t *testing.T) {

	// monomial
	{
		var m monomial
		m.coeff.SetInt64(3)
		m.exponents = make([]int, 3)
		m.exponents[0] = 1
		m.exponents[1] = 2
		m.exponents[2] = 3

		x := make([]fr.Element, 3)
		x[0].SetUint64(2)
		x[1].SetUint64(2)
		x[2].SetUint64(2)

		var res fr.Element
		res.SetUint64(192)

		r := m.evaluate(x)

		if !r.Equal(&res) {
			t.Fatal("evaluation monomial failed")
		}
	}

	// multivariate polynomial
	{
		p := make(multivariatePolynomial, 3)

		p[0].coeff.SetUint64(1)
		p[0].exponents = []int{1, 1, 1}

		p[1].coeff.SetUint64(2)
		p[1].exponents = []int{2, 2, 2}

		p[2].coeff.SetUint64(0)
		p[2].exponents = []int{3, 2, 1}

		x := make([]fr.Element, 3)
		x[0].SetUint64(2)
		x[1].SetUint64(2)
		x[2].SetUint64(2)

		var res fr.Element
		res.SetUint64(136)

		r := p.evaluate(x)
		if !r.Equal(&res) {
			t.Fatal("evaluation multivariate polynomial failed")
		}

	}

}
