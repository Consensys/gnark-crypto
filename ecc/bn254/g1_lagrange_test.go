// Copyright 2020 Consensys Software Inc.
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

package bn254

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

func TestToLagrangeG1(t *testing.T) {
	const size = 32

	var samplePoints [size]G1Affine
	var g G1Jac
	g.Set(&g1Gen)
	for i := 1; i <= size; i++ {
		samplePoints[i-1].FromJacobian(&g)
		g.AddAssign(&g1Gen)
	}

	// convert the test SRS to Lagrange form
	pkLagrange := make([]G1Affine, size)
	copy(pkLagrange, samplePoints[:])
	err := ToLagrangeG1(pkLagrange)
	if err != nil {
		t.Fatal(err)
	}

	// generate the Lagrange SRS manually and compare
	w, err := fr.Generator(uint64(size))
	if err != nil {
		t.Fatal(err)
	}

	var li, n, d, one, acc, alpha fr.Element
	alpha.SetBigInt(bAlpha)
	li.SetUint64(uint64(size)).Inverse(&li)
	one.SetOne()
	n.Exp(alpha, big.NewInt(int64(size))).Sub(&n, &one)
	d.Sub(&alpha, &one)
	li.Mul(&li, &n).Div(&li, &d)
	expectedSrsLagrange := make([]G1Affine, size)
	_, _, g1Gen, _ := Generators()
	var s big.Int
	acc.SetOne()
	for i := 0; i < size; i++ {
		li.BigInt(&s)
		expectedSrsLagrange[i].ScalarMultiplication(&g1Gen, &s)

		li.Mul(&li, &w).Mul(&li, &d)
		acc.Mul(&acc, &w)
		d.Sub(&alpha, &acc)
		li.Div(&li, &d)
	}

	for i := 0; i < size; i++ {
		if !expectedSrsLagrange[i].Equal(&pkLagrange[i]) {
			t.Fatal("error lagrange conversion")
		}
	}
}
