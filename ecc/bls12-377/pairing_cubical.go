// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package bls12377

import (
	"errors"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fp"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/internal/fptower"
)

// PairCubical computes the reduced Tate pairing using cubical arithmetic on the
// Montgomery model of the curve. For odd r (the BLS12-377 subgroup order), this
// returns the square of the canonical pairing value, which is still a valid
// non-degenerate pairing since r is odd.
func PairCubical(P []G1Affine, Q []G2Affine) (GT, error) {
	n := len(P)
	if n == 0 || n != len(Q) {
		return GT{}, errors.New("invalid inputs sizes")
	}

	// filter infinity points
	p := make([]G1Affine, 0, n)
	q := make([]G2Affine, 0, n)
	for k := 0; k < n; k++ {
		if P[k].IsInfinity() || Q[k].IsInfinity() {
			continue
		}
		p = append(p, P[k])
		q = append(q, Q[k])
	}
	n = len(p)
	if n == 0 {
		var one GT
		one.SetOne()
		return one, nil
	}

	var result GT
	result.SetOne()

	for k := 0; k < n; k++ {
		pairing, err := cubicalPairingSingle(&p[k], &q[k])
		if err != nil {
			return GT{}, err
		}
		result.Mul(&result, &pairing)
	}

	return FinalExponentiation(&result), nil
}

// PairingCheck calculates the reduced Tate pairing using cubical arithmetic on
// the Montgomery model of the curve for a set of points and returns True if
// the result is One
// ∏ᵢ e(Pᵢ, Qᵢ) =? 1
//
// This function doesn't check that the inputs are in the correct subgroup. See IsInSubGroup.
func PairingCubicalCheck(P []G1Affine, Q []G2Affine) (bool, error) {
	f, err := PairCubical(P, Q)
	if err != nil {
		return false, err
	}
	var one GT
	one.SetOne()
	return f.Equal(&one), nil
}

type montgomeryPointX struct {
	X fptower.E12
	Z fptower.E12
}

type e12Point struct {
	X   fptower.E12
	Y   fptower.E12
	Inf bool
}

func cubicalPairingSingle(p *G1Affine, q *G2Affine) (GT, error) {
	pw := weierstrassPointFromG1(p)
	qw := weierstrassPointFromG2(q)

	pMinusQ := weierstrassSub(pw, qw)
	if pMinusQ.Inf {
		return GT{}, errors.New("cubical pairing: invalid P-Q")
	}

	xP := montgomeryX(&pw.X)
	xQ := montgomeryX(&qw.X)
	xPQ := montgomeryX(&pMinusQ.X)

	if xP.IsZero() || xQ.IsZero() || xPQ.IsZero() {
		return GT{}, errors.New("cubical pairing: x-coordinate is zero")
	}

	var inverses [3]fptower.E12
	inverses[0] = xP
	inverses[1] = xQ
	inverses[2] = xPQ
	inv := fptower.BatchInvertE12(inverses[:])
	ixP := inv[0]
	ixQ := inv[1]
	ixPQ := inv[2]

	nP, nPQ := cubicalLadder(&xP, &xQ, &xPQ, &ixP, &ixQ, &ixPQ, fr.Modulus())

	// The Tate pairing from the cubical ladder is computed from the projective
	// coordinates of [r]P and [r]P + Q.
	// According to the paper, the pairing value is:
	//   T_r(P, Q) = Z([r]P + Q) / X([r]P)
	// which gives the square of the Tate pairing for odd r.
	var pairing GT
	pairing.Div(&nPQ.Z, &nP.X)
	return pairing, nil
}

func cubicalLadder(xP, xQ, xPQ, ixP, ixQ, ixPQ *fptower.E12, n *big.Int) (montgomeryPointX, montgomeryPointX) {
	var s0, s1, t montgomeryPointX
	s0.X = cubicalOne
	s0.Z = cubicalZero
	s1.X = *xP
	s1.Z = cubicalOne
	t.X = *xQ
	t.Z = cubicalOne

	if n.Sign() == 0 {
		return s0, t
	}

	var r montgomeryPointX
	for i := n.BitLen() - 1; i >= 0; i-- {
		r = cubicalDiffAdd(&s0, &s1, ixP)

		if n.Bit(i) == 0 {
			t = cubicalDiffAdd(&t, &s0, ixQ)
			s0 = cubicalDouble(&s0)
			s1 = r
		} else {
			t = cubicalDiffAdd(&t, &s1, ixPQ)
			s1 = cubicalDouble(&s1)
			s0 = r
		}
	}

	return s0, t
}

func cubicalDouble(p *montgomeryPointX) montgomeryPointX {
	var t0, t1, t2, t3 fptower.E12
	var x2, z2 fptower.E12

	t0.Add(&p.X, &p.Z)
	t0.Square(&t0)
	t1.Sub(&p.X, &p.Z)
	t1.Square(&t1)
	x2.Mul(&t0, &t1)
	t2.Sub(&t0, &t1)
	mulE12ByFp(&t3, &t2, &montgomeryA24)
	t3.Add(&t3, &t1)
	z2.Mul(&t2, &t3)

	return montgomeryPointX{X: x2, Z: z2}
}

func cubicalDiffAdd(p, q *montgomeryPointX, ixPQ *fptower.E12) montgomeryPointX {
	var t0, t1, t2, t3 fptower.E12
	var x, z fptower.E12

	t0.Sub(&p.X, &p.Z)
	t1.Add(&q.X, &q.Z)
	t0.Mul(&t0, &t1)

	t1.Add(&p.X, &p.Z)
	t2.Sub(&q.X, &q.Z)
	t1.Mul(&t1, &t2)

	t2.Add(&t0, &t1)
	t2.Square(&t2)
	x.Mul(ixPQ, &t2)

	t3.Sub(&t0, &t1)
	t3.Square(&t3)
	z.Set(&t3)

	mulE12ByFp(&x, &x, &invFour)
	mulE12ByFp(&z, &z, &invFour)

	return montgomeryPointX{X: x, Z: z}
}

func weierstrassPointFromG1(p *G1Affine) e12Point {
	if p.IsInfinity() {
		return e12Point{Inf: true}
	}
	var x, y fptower.E12
	x.C0.B0.A0.Set(&p.X)
	y.C0.B0.A0.Set(&p.Y)
	return e12Point{X: x, Y: y}
}

func weierstrassPointFromG2(q *G2Affine) e12Point {
	if q.IsInfinity() {
		return e12Point{Inf: true}
	}
	// For BLS12-377 with D-type twist: E': y² = x³ + 1/u over Fp²
	// The untwist isogeny φ: E'(Fp²) → E(Fp¹²) is:
	//   (x', y') → (ψ²·x', ψ³·y')
	// where ψ⁶ = u (the twist factor).
	//
	// In the tower Fp¹² = Fp⁶[w] where w² = v and Fp⁶ = Fp²[v] where v³ = u:
	//   ψ = w, so ψ² = v and ψ³ = v·w
	//
	// Therefore:
	//   x = v · x'
	//   y = v·w · y'
	var xE6, yE6 fptower.E6
	xE6.B0 = q.X
	yE6.B0 = q.Y
	xE6.MulByNonResidue(&xE6)
	yE6.MulByNonResidue(&yE6)

	var x, y fptower.E12
	x.C0 = xE6
	y.C1 = yE6
	return e12Point{X: x, Y: y}
}

func weierstrassSub(p, q e12Point) e12Point {
	if q.Inf {
		return p
	}
	if p.Inf {
		var negY fptower.E12
		negE12(&negY, &q.Y)
		return e12Point{X: q.X, Y: negY}
	}
	var negY fptower.E12
	negE12(&negY, &q.Y)
	return weierstrassAdd(p, e12Point{X: q.X, Y: negY})
}

func weierstrassAdd(p, q e12Point) e12Point {
	if p.Inf {
		return q
	}
	if q.Inf {
		return p
	}
	if p.X.Equal(&q.X) {
		var ySum fptower.E12
		ySum.Add(&p.Y, &q.Y)
		if ySum.IsZero() {
			return e12Point{Inf: true}
		}
		return weierstrassDouble(p)
	}

	var lambda, num, den, denInv fptower.E12
	num.Sub(&q.Y, &p.Y)
	den.Sub(&q.X, &p.X)
	denInv.Inverse(&den)
	lambda.Mul(&num, &denInv)

	var x3, y3 fptower.E12
	x3.Square(&lambda)
	x3.Sub(&x3, &p.X)
	x3.Sub(&x3, &q.X)

	y3.Sub(&p.X, &x3)
	y3.Mul(&y3, &lambda)
	y3.Sub(&y3, &p.Y)

	return e12Point{X: x3, Y: y3}
}

func weierstrassDouble(p e12Point) e12Point {
	if p.Inf {
		return p
	}
	if p.Y.IsZero() {
		return e12Point{Inf: true}
	}

	var xSq, num, den, denInv, lambda fptower.E12
	xSq.Square(&p.X)
	mulE12ByFp(&num, &xSq, &three)
	mulE12ByFp(&den, &p.Y, &two)
	denInv.Inverse(&den)
	lambda.Mul(&num, &denInv)

	var x3, y3 fptower.E12
	x3.Square(&lambda)
	var twoX fptower.E12
	mulE12ByFp(&twoX, &p.X, &two)
	x3.Sub(&x3, &twoX)

	y3.Sub(&p.X, &x3)
	y3.Mul(&y3, &lambda)
	y3.Sub(&y3, &p.Y)

	return e12Point{X: x3, Y: y3}
}

func montgomeryX(xW *fptower.E12) fptower.E12 {
	var t fptower.E12
	t.Set(xW)
	t.C0.B0.A0.Add(&t.C0.B0.A0, &fpOne)

	var out fptower.E12
	mulE12ByFp(&out, &t, &montgomeryU2)
	return out
}

func mulE12ByFp(z, x *fptower.E12, c *fp.Element) *fptower.E12 {
	z.C0.B0.MulByElement(&x.C0.B0, c)
	z.C0.B1.MulByElement(&x.C0.B1, c)
	z.C0.B2.MulByElement(&x.C0.B2, c)
	z.C1.B0.MulByElement(&x.C1.B0, c)
	z.C1.B1.MulByElement(&x.C1.B1, c)
	z.C1.B2.MulByElement(&x.C1.B2, c)
	return z
}

func negE12(z, x *fptower.E12) *fptower.E12 {
	z.C0.Neg(&x.C0)
	z.C1.Neg(&x.C1)
	return z
}

var (
	montgomeryA   fp.Element
	montgomeryA24 fp.Element
	montgomeryU2  fp.Element
	invFour       fp.Element
	fpOne         fp.Element
	two           fp.Element
	three         fp.Element
	cubicalOne    fptower.E12
	cubicalZero   fptower.E12
)

func init() {
	montgomeryA.SetString("30567070899668889872121584789658882274245471728719284894883538395508419196346447682510590835309008936731240225793")

	fpOne.SetOne()
	two.SetUint64(2)
	three.SetUint64(3)

	var tmp fp.Element
	tmp.Set(&montgomeryA)
	tmp.Add(&tmp, &two)
	montgomeryA24 = tmp
	montgomeryA24.Halve()
	montgomeryA24.Halve()

	var invThree fp.Element
	invThree.SetUint64(3)
	invThree.Inverse(&invThree)
	montgomeryU2.Mul(&montgomeryA, &invThree).Neg(&montgomeryU2)

	invFour.SetUint64(4)
	invFour.Inverse(&invFour)

	cubicalOne.SetOne()
	cubicalZero = fptower.E12{}
}
