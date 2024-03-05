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

package shplonk

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

func TestBuildVanishingPoly(t *testing.T) {
	s := 10
	x := make([]fr.Element, s)
	for i := 0; i < s; i++ {
		x[i].SetRandom()
	}
	r := buildVanishingPoly(x)

	if len(r) != s+1 {
		t.Fatal("error degree r")
	}

	// check that r(x_{i})=0 for all i
	for i := 0; i < len(x); i++ {
		y := eval(r, x[i])
		if !y.IsZero() {
			t.Fatal("πᵢ(X-xᵢ) at xᵢ should be zero")
		}
	}

	// check that r(y)!=0 for a random point
	var a fr.Element
	a.SetRandom()
	y := eval(r, a)
	if y.IsZero() {
		t.Fatal("πᵢ(X-xᵢ) at r \neq xᵢ should not be zero")
	}
}

func TestMultiplyLinearFactor(t *testing.T) {

	s := 10
	f := make([]fr.Element, s, s+1)
	for i := 0; i < 10; i++ {
		f[i].SetRandom()
	}

	var a, y fr.Element
	a.SetRandom()
	f = multiplyLinearFactor(f, a)
	y = eval(f, a)
	if !y.IsZero() {
		t.Fatal("(X-a)f(X) should be zero at a")
	}
	a.SetRandom()
	y = eval(f, a)
	if y.IsZero() {
		t.Fatal("(X-1)f(X) at a random point should not be zero")
	}

}

func TestNaiveMul(t *testing.T) {

	size := 10
	f := make([]fr.Element, size)
	for i := 0; i < size; i++ {
		f[i].SetRandom()
	}

	nbPoints := 10
	points := make([]fr.Element, nbPoints)
	for i := 0; i < nbPoints; i++ {
		points[i].SetRandom()
	}

	v := buildVanishingPoly(points)
	buf := make([]fr.Element, size+nbPoints-1)
	g := mul(f, v, buf)

	// check that g(x_{i}) = 0
	for i := 0; i < nbPoints; i++ {
		y := eval(g, points[i])
		if !y.IsZero() {
			t.Fatal("f(X)(X-x_{1})..(X-x_{n}) at x_{i} should be zero")
		}
	}

	// check that g(r) != 0 for a random point
	var a fr.Element
	a.SetRandom()
	y := eval(g, a)
	if y.IsZero() {
		t.Fatal("f(X)(X-x_{1})..(X-x_{n}) at a random point should not be zero")
	}

}

func TestDiv(t *testing.T) {

	nbPoints := 10
	s := 10
	f := make([]fr.Element, s, s+nbPoints)
	for i := 0; i < s; i++ {
		f[i].SetRandom()
	}

	// backup
	g := make([]fr.Element, s)
	copy(g, f)

	// successive divions of linear terms
	x := make([]fr.Element, nbPoints)
	for i := 0; i < nbPoints; i++ {
		x[i].SetRandom()
		f = multiplyLinearFactor(f, x[i])
	}
	q := make([][2]fr.Element, nbPoints)
	for i := 0; i < nbPoints; i++ {
		q[i][1].SetOne()
		q[i][0].Neg(&x[i])
		f = div(f, q[i][:])
	}

	// g should be equal to f
	if len(f) != len(g) {
		t.Fatal("lengths don't match")
	}
	for i := 0; i < len(g); i++ {
		if !f[i].Equal(&g[i]) {
			t.Fatal("f(x)(x-a)/(x-a) should be equal to f(x)")
		}
	}

	// division by a degree > 1 polynomial
	for i := 0; i < nbPoints; i++ {
		x[i].SetRandom()
		f = multiplyLinearFactor(f, x[i])
	}
	r := buildVanishingPoly(x)
	f = div(f, r)

	// g should be equal to f
	if len(f) != len(g) {
		t.Fatal("lengths don't match")
	}
	for i := 0; i < len(g); i++ {
		if !f[i].Equal(&g[i]) {
			t.Fatal("f(x)(x-a)/(x-a) should be equal to f(x)")
		}
	}

}
