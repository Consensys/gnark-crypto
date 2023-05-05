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

package bw6761

import (
	"github.com/consensys/gnark-crypto/ecc/bw6-761/fp"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
	"math/rand"
	"testing"
)

func TestG2SqrtRatio(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	gen := GenFp()

	properties.Property("G2SqrtRatio must square back to the right value", prop.ForAll(
		func(u fp.Element, v fp.Element) bool {

			var seen fp.Element
			qr := g2SqrtRatio(&seen, &u, &v) == 0

			seen.
				Square(&seen).
				Mul(&seen, &v)

			var ref fp.Element
			if qr {
				ref = u
			} else {
				g2MulByZ(&ref, &u)
			}

			return seen.Equal(&ref)
		}, gen, gen))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestHashToFpG2(t *testing.T) {
	for _, c := range encodeToG2Vector.cases {
		elems, err := fp.Hash([]byte(c.msg), encodeToG2Vector.dst, 1)
		if err != nil {
			t.Error(err)
		}
		g2TestMatchCoord(t, "u", c.msg, c.u, g2CoordAt(elems, 0))
	}

	for _, c := range hashToG2Vector.cases {
		elems, err := fp.Hash([]byte(c.msg), hashToG2Vector.dst, 2*1)
		if err != nil {
			t.Error(err)
		}
		g2TestMatchCoord(t, "u0", c.msg, c.u0, g2CoordAt(elems, 0))
		g2TestMatchCoord(t, "u1", c.msg, c.u1, g2CoordAt(elems, 1))
	}
}

func TestMapToCurve2(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	properties.Property("[G2] mapping output must be on curve", prop.ForAll(
		func(a fp.Element) bool {

			g := MapToCurve2(&a)

			if !isOnE2Prime(g) {
				t.Log("Mapping output not on E' curve")
				return false
			}
			g2Isogeny(&g)

			if !g.IsOnCurve() {
				t.Log("Isogeny∘SSWU output not on curve")
				return false
			}

			return true
		},
		GenFp(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))

	for _, c := range encodeToG2Vector.cases {
		var u fp.Element
		g2CoordSetString(&u, c.u)
		q := MapToCurve2(&u)
		g2Isogeny(&q)
		g2TestMatchPoint(t, "Q", c.msg, c.Q, &q)
	}

	for _, c := range hashToG2Vector.cases {
		var u fp.Element
		g2CoordSetString(&u, c.u0)
		q := MapToCurve2(&u)
		g2Isogeny(&q)
		g2TestMatchPoint(t, "Q0", c.msg, c.Q0, &q)

		g2CoordSetString(&u, c.u1)
		q = MapToCurve2(&u)
		g2Isogeny(&q)
		g2TestMatchPoint(t, "Q1", c.msg, c.Q1, &q)
	}
}

func TestMapToG2(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	properties.Property("[G2] mapping to curve should output point on the curve", prop.ForAll(
		func(a fp.Element) bool {
			g := MapToG2(a)
			return g.IsInSubGroup()
		},
		GenFp(),
	))

	properties.Property("[G2] mapping to curve should be deterministic", prop.ForAll(
		func(a fp.Element) bool {
			g1 := MapToG2(a)
			g2 := MapToG2(a)
			return g1.Equal(&g2)
		},
		GenFp(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestEncodeToG2(t *testing.T) {
	t.Parallel()
	for _, c := range encodeToG2Vector.cases {
		p, err := EncodeToG2([]byte(c.msg), encodeToG2Vector.dst)
		if err != nil {
			t.Fatal(err)
		}
		g2TestMatchPoint(t, "P", c.msg, c.P, &p)
	}
}

func TestHashToG2(t *testing.T) {
	t.Parallel()
	for _, c := range hashToG2Vector.cases {
		p, err := HashToG2([]byte(c.msg), hashToG2Vector.dst)
		if err != nil {
			t.Fatal(err)
		}
		g2TestMatchPoint(t, "P", c.msg, c.P, &p)
	}
}

func BenchmarkEncodeToG2(b *testing.B) {
	const size = 54
	bytes := make([]byte, size)
	dst := encodeToG2Vector.dst
	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		bytes[rand.Int()%size] = byte(rand.Int()) //#nosec G404 weak rng is fine here

		if _, err := EncodeToG2(bytes, dst); err != nil {
			b.Fail()
		}
	}
}

func BenchmarkHashToG2(b *testing.B) {
	const size = 54
	bytes := make([]byte, size)
	dst := hashToG2Vector.dst
	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		bytes[rand.Int()%size] = byte(rand.Int()) //#nosec G404 weak rng is fine here

		if _, err := HashToG2(bytes, dst); err != nil {
			b.Fail()
		}
	}
}

// TODO: Crude. Do something clever in Jacobian
func isOnE2Prime(p G2Affine) bool {

	var A, B fp.Element

	A.SetString(
		"6429719010846137499474887978131198018330761288163789627290055406883908067119696591103101123992665411263189240363728172709848698522760005194862816392151436104205214136976570209818204605171075531070134198773930389453798390056516896",
	)

	B.SetString(
		"5348306863922295212600474030012704926780090705412552782187041272079620891140642329199277344037019889626771397168162938103438296026272884909103171857394985776682488984714989551922989188985164920238405955336107390943902906254560160",
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
func g2CoordSetString(z *fp.Element, s string) {
	z.SetString(s)
}

func g2CoordAt(slice []fp.Element, i int) fp.Element {
	return slice[i]
}

func g2TestMatchCoord(t *testing.T, coordName string, msg string, expectedStr string, seen fp.Element) {
	var expected fp.Element

	g2CoordSetString(&expected, expectedStr)

	if !expected.Equal(&seen) {
		t.Errorf("mismatch on \"%s\", %s:\n\texpected %s\n\tsaw      %s", msg, coordName, expected.String(), &seen)
	}
}

func g2TestMatchPoint(t *testing.T, pointName string, msg string, expected point, seen *G2Affine) {
	g2TestMatchCoord(t, pointName+".x", msg, expected.x, seen.X)
	g2TestMatchCoord(t, pointName+".y", msg, expected.y, seen.Y)
}

var encodeToG2Vector encodeTestVector
var hashToG2Vector hashTestVector
