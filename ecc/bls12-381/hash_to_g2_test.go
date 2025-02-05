// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package bls12381

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fp"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/internal/fptower"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
	"math/rand"
	"strings"
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

	gen := GenE2()

	properties.Property("G2SqrtRatio must square back to the right value", prop.ForAll(
		func(u fptower.E2, v fptower.E2) bool {

			var seen fptower.E2
			qr := G2SqrtRatio(&seen, &u, &v) == 0

			seen.
				Square(&seen).
				Mul(&seen, &v)

			var ref fptower.E2
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
		elems, err := fp.Hash([]byte(c.msg), encodeToG2Vector.dst, 2)
		if err != nil {
			t.Error(err)
		}
		g2TestMatchCoord(t, "u", c.msg, c.u, g2CoordAt(elems, 0))
	}

	for _, c := range hashToG2Vector.cases {
		elems, err := fp.Hash([]byte(c.msg), hashToG2Vector.dst, 2*2)
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
		func(a fptower.E2) bool {

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
		GenE2(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))

	for _, c := range encodeToG2Vector.cases {
		var u fptower.E2
		g2CoordSetString(&u, c.u)
		q := MapToCurve2(&u)
		g2Isogeny(&q)
		g2TestMatchPoint(t, "Q", c.msg, c.Q, &q)
	}

	for _, c := range hashToG2Vector.cases {
		var u fptower.E2
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
		func(a fptower.E2) bool {
			g := MapToG2(a)
			return g.IsInSubGroup()
		},
		GenE2(),
	))

	properties.Property("[G2] mapping to curve should be deterministic", prop.ForAll(
		func(a fptower.E2) bool {
			g1 := MapToG2(a)
			g2 := MapToG2(a)
			return g1.Equal(&g2)
		},
		GenE2(),
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

	var A, B fptower.E2

	A.SetString(
		"0",
		"240",
	)

	B.SetString(
		"1012",
		"1012",
	)

	var LHS fptower.E2
	LHS.
		Square(&p.Y).
		Sub(&LHS, &B)

	var RHS fptower.E2
	RHS.
		Square(&p.X).
		Add(&RHS, &A).
		Mul(&RHS, &p.X)

	return LHS.Equal(&RHS)
}

// Only works on simple extensions (two-story towers)
func g2CoordSetString(z *fptower.E2, s string) {
	ssplit := strings.Split(s, ",")
	if len(ssplit) != 2 {
		panic("not equal to tower size")
	}
	z.SetString(
		ssplit[0],
		ssplit[1],
	)
}

func g2CoordAt(slice []fp.Element, i int) fptower.E2 {
	return fptower.E2{
		A0: slice[i*2+0],
		A1: slice[i*2+1],
	}
}

func g2TestMatchCoord(t *testing.T, coordName string, msg string, expectedStr string, seen fptower.E2) {
	var expected fptower.E2

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
