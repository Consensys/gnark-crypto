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

package bls24315

import (
	"github.com/consensys/gnark-crypto/ecc/bls24-315/fp"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
	"math/rand"
	"testing"
)

func TestG1SqrtRatio(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	gen := GenFp()

	properties.Property("G1SqrtRatio must square back to the right value", prop.ForAll(
		func(u fp.Element, v fp.Element) bool {

			var seen fp.Element
			qr := g1SqrtRatio(&seen, &u, &v) == 0

			seen.
				Square(&seen).
				Mul(&seen, &v)

			var ref fp.Element
			if qr {
				ref = u
			} else {
				g1MulByZ(&ref, &u)
			}

			return seen.Equal(&ref)
		}, gen, gen))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestHashToFpG1(t *testing.T) {
	for _, c := range encodeToG1Vector.cases {
		elems, err := fp.Hash([]byte(c.msg), encodeToG1Vector.dst, 1)
		if err != nil {
			t.Error(err)
		}
		g1TestMatchCoord(t, "u", c.msg, c.u, g1CoordAt(elems, 0))
	}

	for _, c := range hashToG1Vector.cases {
		elems, err := fp.Hash([]byte(c.msg), hashToG1Vector.dst, 2*1)
		if err != nil {
			t.Error(err)
		}
		g1TestMatchCoord(t, "u0", c.msg, c.u0, g1CoordAt(elems, 0))
		g1TestMatchCoord(t, "u1", c.msg, c.u1, g1CoordAt(elems, 1))
	}
}

func TestMapToCurve1(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	properties.Property("[G1] mapping output must be on curve", prop.ForAll(
		func(a fp.Element) bool {

			g := MapToCurve1(&a)

			if !isOnE1Prime(g) {
				t.Log("Mapping output not on E' curve")
				return false
			}
			g1Isogeny(&g)

			if !g.IsOnCurve() {
				t.Log("Isogeny∘SSWU output not on curve")
				return false
			}

			return true
		},
		GenFp(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))

	for _, c := range encodeToG1Vector.cases {
		var u fp.Element
		g1CoordSetString(&u, c.u)
		q := MapToCurve1(&u)
		g1Isogeny(&q)
		g1TestMatchPoint(t, "Q", c.msg, c.Q, &q)
	}

	for _, c := range hashToG1Vector.cases {
		var u fp.Element
		g1CoordSetString(&u, c.u0)
		q := MapToCurve1(&u)
		g1Isogeny(&q)
		g1TestMatchPoint(t, "Q0", c.msg, c.Q0, &q)

		g1CoordSetString(&u, c.u1)
		q = MapToCurve1(&u)
		g1Isogeny(&q)
		g1TestMatchPoint(t, "Q1", c.msg, c.Q1, &q)
	}
}

func TestMapToG1(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	properties.Property("[G1] mapping to curve should output point on the curve", prop.ForAll(
		func(a fp.Element) bool {
			g := MapToG1(a)
			return g.IsInSubGroup()
		},
		GenFp(),
	))

	properties.Property("[G1] mapping to curve should be deterministic", prop.ForAll(
		func(a fp.Element) bool {
			g1 := MapToG1(a)
			g2 := MapToG1(a)
			return g1.Equal(&g2)
		},
		GenFp(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestEncodeToG1(t *testing.T) {
	t.Parallel()
	for _, c := range encodeToG1Vector.cases {
		p, err := EncodeToG1([]byte(c.msg), encodeToG1Vector.dst)
		if err != nil {
			t.Fatal(err)
		}
		g1TestMatchPoint(t, "P", c.msg, c.P, &p)
	}
}

func TestHashToG1(t *testing.T) {
	t.Parallel()
	for _, c := range hashToG1Vector.cases {
		p, err := HashToG1([]byte(c.msg), hashToG1Vector.dst)
		if err != nil {
			t.Fatal(err)
		}
		g1TestMatchPoint(t, "P", c.msg, c.P, &p)
	}
}

func BenchmarkEncodeToG1(b *testing.B) {
	const size = 54
	bytes := make([]byte, size)
	dst := encodeToG1Vector.dst
	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		bytes[rand.Int()%size] = byte(rand.Int()) //#nosec G404 weak rng is fine here

		if _, err := EncodeToG1(bytes, dst); err != nil {
			b.Fail()
		}
	}
}

func BenchmarkHashToG1(b *testing.B) {
	const size = 54
	bytes := make([]byte, size)
	dst := hashToG1Vector.dst
	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		bytes[rand.Int()%size] = byte(rand.Int()) //#nosec G404 weak rng is fine here

		if _, err := HashToG1(bytes, dst); err != nil {
			b.Fail()
		}
	}
}

// TODO: Crude. Do something clever in Jacobian
func isOnE1Prime(p G1Affine) bool {

	var A, B fp.Element

	A.SetString(
		"39705142154296798234718093138458736353730097451069869796965271356892223115922042164250209681439",
	)

	B.SetString(
		"22",
	)

	var LHS fp.Element
	LHS.
		Square(&p.Y).
		Sub(&LHS, &B)

	var RHS fp.Element
	RHS.
		Square(&p.X).
		Add(&RHS, &A).
		Mul(&RHS, &p.X)

	return LHS.Equal(&RHS)
}

// Only works on simple extensions (two-story towers)
func g1CoordSetString(z *fp.Element, s string) {
	z.SetString(s)
}

func g1CoordAt(slice []fp.Element, i int) fp.Element {
	return slice[i]
}

func g1TestMatchCoord(t *testing.T, coordName string, msg string, expectedStr string, seen fp.Element) {
	var expected fp.Element

	g1CoordSetString(&expected, expectedStr)

	if !expected.Equal(&seen) {
		t.Errorf("mismatch on \"%s\", %s:\n\texpected %s\n\tsaw      %s", msg, coordName, expected.String(), &seen)
	}
}

func g1TestMatchPoint(t *testing.T, pointName string, msg string, expected point, seen *G1Affine) {
	g1TestMatchCoord(t, pointName+".x", msg, expected.x, seen.X)
	g1TestMatchCoord(t, pointName+".y", msg, expected.y, seen.Y)
}

type hashTestVector struct {
	dst   []byte
	cases []hashTestCase
}

type encodeTestVector struct {
	dst   []byte
	cases []encodeTestCase
}

type point struct {
	x string
	y string
}

type encodeTestCase struct {
	msg string
	P   point  //pY a coordinate of P, the final output
	u   string //u hashed onto the field
	Q   point  //Q map to curve output
}

type hashTestCase struct {
	msg string
	P   point  //pY a coordinate of P, the final output
	u0  string //u0 hashed onto the field
	u1  string //u1 extra hashed onto the field
	Q0  point  //Q0 map to curve output
	Q1  point  //Q1 extra map to curve output
}

var encodeToG1Vector encodeTestVector
var hashToG1Vector hashTestVector
