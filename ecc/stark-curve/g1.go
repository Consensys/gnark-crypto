// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// FOO

package starkcurve

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fr"
	"github.com/consensys/gnark-crypto/internal/parallel"
)

// G1Affine is a point in affine coordinates (x,y)
type G1Affine struct {
	X, Y fp.Element
}

// G1Jac is a point in Jacobian coordinates (x=X/Z², y=Y/Z³)
type G1Jac struct {
	X, Y, Z fp.Element
}

// g1JacExtended is a point in extended Jacobian coordinates (x=X/ZZ, y=Y/ZZZ, ZZ³=ZZZ²)
type g1JacExtended struct {
	X, Y, ZZ, ZZZ fp.Element
}

// -------------------------------------------------------------------------------------------------
// Affine coordinates

// Set sets p to a in affine coordinates.
func (p *G1Affine) Set(a *G1Affine) *G1Affine {
	p.X, p.Y = a.X, a.Y
	return p
}

// ScalarMultiplication computes and returns p = [s]a
// where p and a are affine points.
func (p *G1Affine) ScalarMultiplication(a *G1Affine, s *big.Int) *G1Affine {
	var _p G1Jac
	_p.FromAffine(a)
	_p.mulWindowed(&_p, s)
	p.FromJacobian(&_p)
	return p
}

// ScalarMultiplicationBase computes and returns p = [s]g
// where g is the affine point generating the prime subgroup.
func (p *G1Affine) ScalarMultiplicationBase(s *big.Int) *G1Affine {
	var _p G1Jac
	_p.mulWindowed(&g1Gen, s)
	p.FromJacobian(&_p)
	return p
}

// Add adds two points in affine coordinates.
// It uses the Jacobian addition with a.Z=b.Z=1 and converts the result to affine coordinates.
//
// https://www.hyperelliptic.org/EFD/g1p/auto-shortw-jacobian-0.html#addition-mmadd-2007-bl
func (p *G1Affine) Add(a, b *G1Affine) *G1Affine {
	var p1, p2 G1Jac
	p1.FromAffine(a)
	p2.FromAffine(b)
	p1.AddAssign(&p2)
	p.FromJacobian(&p1)
	return p
}

// Sub subtracts two points in affine coordinates.
// It uses a similar approach to Add, but negates the second point before adding.
func (p *G1Affine) Sub(a, b *G1Affine) *G1Affine {
	var p1, p2 G1Jac
	p1.FromAffine(a)
	p2.FromAffine(b)
	p1.SubAssign(&p2)
	p.FromJacobian(&p1)
	return p
}

// Equal tests if two points in affine coordinates are equal.
func (p *G1Affine) Equal(a *G1Affine) bool {
	return p.X.Equal(&a.X) && p.Y.Equal(&a.Y)
}

// Neg sets p to the affine negative point -a = (a.X, -a.Y).
func (p *G1Affine) Neg(a *G1Affine) *G1Affine {
	p.X = a.X
	p.Y.Neg(&a.Y)
	return p
}

// FromJacobian converts a point p1 from Jacobian to affine coordinates.
func (p *G1Affine) FromJacobian(p1 *G1Jac) *G1Affine {

	var a, b fp.Element

	if p1.Z.IsZero() {
		p.X.SetZero()
		p.Y.SetZero()
		return p
	}

	a.Inverse(&p1.Z)
	b.Square(&a)
	p.X.Mul(&p1.X, &b)
	p.Y.Mul(&p1.Y, &b).Mul(&p.Y, &a)

	return p
}

// String returns the string representation E(x,y) of the affine point p or "O" if it is infinity.
func (p *G1Affine) String() string {
	if p.IsInfinity() {
		return "O"
	}
	return "E([" + p.X.String() + "," + p.Y.String() + "])"
}

// IsInfinity checks if the affine point p is infinity, which is encoded as (0,0).
// N.B.: (0,0) is not on the STARK curve (Y²=X³+X+B).
func (p *G1Affine) IsInfinity() bool {
	return p.X.IsZero() && p.Y.IsZero()
}

// IsOnCurve returns true if the affine point p in on the curve.
func (p *G1Affine) IsOnCurve() bool {
	var point G1Jac
	point.FromAffine(p)
	return point.IsOnCurve() // call this function to handle infinity point
}

// IsInSubGroup returns true if the affine point p is in the correct subgroup, false otherwise.
func (p *G1Affine) IsInSubGroup() bool {
	var _p G1Jac
	_p.FromAffine(p)
	return _p.IsInSubGroup()
}

// -------------------------------------------------------------------------------------------------
// Jacobian coordinates

// Set sets p to a in Jacobian coordinates.
func (p *G1Jac) Set(q *G1Jac) *G1Jac {
	p.X, p.Y, p.Z = q.X, q.Y, q.Z
	return p
}

// Equal tests if two points in Jacobian coordinates are equal.
func (p *G1Jac) Equal(a *G1Jac) bool {

	if p.Z.IsZero() && a.Z.IsZero() {
		return true
	}
	_p := G1Affine{}
	_p.FromJacobian(p)

	_a := G1Affine{}
	_a.FromJacobian(a)

	return _p.X.Equal(&_a.X) && _p.Y.Equal(&_a.Y)
}

// Neg sets p to the Jacobian negative point -q = (q.X, -q.Y, q.Z).
func (p *G1Jac) Neg(q *G1Jac) *G1Jac {
	*p = *q
	p.Y.Neg(&q.Y)
	return p
}

// SubAssign sets p to p-a in Jacobian coordinates.
// It uses a similar approach to AddAssign, but negates the point a before adding.
func (p *G1Jac) SubAssign(a *G1Jac) *G1Jac {
	var tmp G1Jac
	tmp.Set(a)
	tmp.Y.Neg(&tmp.Y)
	p.AddAssign(&tmp)
	return p
}

// AddAssign sets p to p+a in Jacobian coordinates.
//
// https://hyperelliptic.org/EFD/g1p/auto-shortw-jacobian-3.html#addition-add-2007-bl
func (p *G1Jac) AddAssign(a *G1Jac) *G1Jac {

	// p is infinity, return a
	if p.Z.IsZero() {
		p.Set(a)
		return p
	}

	// a is infinity, return p
	if a.Z.IsZero() {
		return p
	}

	var Z1Z1, Z2Z2, U1, U2, S1, S2, H, I, J, r, V fp.Element
	Z1Z1.Square(&a.Z)
	Z2Z2.Square(&p.Z)
	U1.Mul(&a.X, &Z2Z2)
	U2.Mul(&p.X, &Z1Z1)
	S1.Mul(&a.Y, &p.Z).
		Mul(&S1, &Z2Z2)
	S2.Mul(&p.Y, &a.Z).
		Mul(&S2, &Z1Z1)

	// if p == a, we double instead
	if U1.Equal(&U2) && S1.Equal(&S2) {
		return p.DoubleAssign()
	}

	H.Sub(&U2, &U1)
	I.Double(&H).
		Square(&I)
	J.Mul(&H, &I)
	r.Sub(&S2, &S1).Double(&r)
	V.Mul(&U1, &I)
	p.X.Square(&r).
		Sub(&p.X, &J).
		Sub(&p.X, &V).
		Sub(&p.X, &V)
	p.Y.Sub(&V, &p.X).
		Mul(&p.Y, &r)
	S1.Mul(&S1, &J).Double(&S1)
	p.Y.Sub(&p.Y, &S1)
	p.Z.Add(&p.Z, &a.Z)
	p.Z.Square(&p.Z).
		Sub(&p.Z, &Z1Z1).
		Sub(&p.Z, &Z2Z2).
		Mul(&p.Z, &H)

	return p
}

// AddMixed sets p to p+a in Jacobian coordinates, where a.Z = 1.
//
// http://www.hyperelliptic.org/EFD/g1p/auto-shortw-jacobian-0.html#addition-madd-2007-bl
func (p *G1Jac) AddMixed(a *G1Affine) *G1Jac {

	//if a is infinity return p
	if a.IsInfinity() {
		return p
	}
	// p is infinity, return a
	if p.Z.IsZero() {
		p.X = a.X
		p.Y = a.Y
		p.Z.SetOne()
		return p
	}

	var Z1Z1, U2, S2, H, HH, I, J, r, V fp.Element
	Z1Z1.Square(&p.Z)
	U2.Mul(&a.X, &Z1Z1)
	S2.Mul(&a.Y, &p.Z).
		Mul(&S2, &Z1Z1)

	// if p == a, we double instead
	if U2.Equal(&p.X) && S2.Equal(&p.Y) {
		return p.DoubleAssign()
	}

	H.Sub(&U2, &p.X)
	HH.Square(&H)
	I.Double(&HH).Double(&I)
	J.Mul(&H, &I)
	r.Sub(&S2, &p.Y).Double(&r)
	V.Mul(&p.X, &I)
	p.X.Square(&r).
		Sub(&p.X, &J).
		Sub(&p.X, &V).
		Sub(&p.X, &V)
	J.Mul(&J, &p.Y).Double(&J)
	p.Y.Sub(&V, &p.X).
		Mul(&p.Y, &r)
	p.Y.Sub(&p.Y, &J)
	p.Z.Add(&p.Z, &H)
	p.Z.Square(&p.Z).
		Sub(&p.Z, &Z1Z1).
		Sub(&p.Z, &HH)

	return p
}

// Double sets p to [2]q in Jacobian coordinates.
//
// https://hyperelliptic.org/EFD/g1p/auto-shortw-jacobian-3.html#doubling-dbl-2007-bl
func (p *G1Jac) Double(q *G1Jac) *G1Jac {
	p.Set(q)
	p.DoubleAssign()
	return p
}

// DoubleAssign doubles p in Jacobian coordinates.
//
// https://hyperelliptic.org/EFD/g1p/auto-shortw-jacobian-3.html#doubling-dbl-2007-bl
func (p *G1Jac) DoubleAssign() *G1Jac {

	var XX, YY, YYYY, ZZ, S, M, T, ZZZZ fp.Element

	XX.Square(&p.X)
	YY.Square(&p.Y)
	YYYY.Square(&YY)
	ZZ.Square(&p.Z)
	S.Add(&p.X, &YY)
	S.Square(&S).
		Sub(&S, &XX).
		Sub(&S, &YYYY).
		Double(&S)
	M.Double(&XX).Add(&M, &XX)
	ZZZZ.Square(&ZZ)
	M.Add(&M, &ZZZZ)
	p.Z.Add(&p.Z, &p.Y).
		Square(&p.Z).
		Sub(&p.Z, &YY).
		Sub(&p.Z, &ZZ)
	T.Square(&M)
	p.X = T
	T.Double(&S)
	p.X.Sub(&p.X, &T)
	p.Y.Sub(&S, &p.X).
		Mul(&p.Y, &M)
	YYYY.Double(&YYYY).Double(&YYYY).Double(&YYYY)
	p.Y.Sub(&p.Y, &YYYY)

	return p
}

// ScalarMultiplication computes and returns p = [s]a
// using a 2-bits windowed double-and-add method.
func (p *G1Jac) ScalarMultiplication(a *G1Jac, s *big.Int) *G1Jac {
	return p.mulWindowed(a, s)
}

// String converts p to affine coordinates and returns its string representation E(x,y) or "O" if it is infinity.
func (p *G1Jac) String() string {
	_p := G1Affine{}
	_p.FromJacobian(p)
	return _p.String()
}

// FromAffine converts a point a from affine to Jacobian coordinates.
func (p *G1Jac) FromAffine(Q *G1Affine) *G1Jac {
	if Q.IsInfinity() {
		p.Z.SetZero()
		p.X.SetOne()
		p.Y.SetOne()
		return p
	}
	p.Z.SetOne()
	p.X.Set(&Q.X)
	p.Y.Set(&Q.Y)
	return p
}

// IsOnCurve returns true if p in on the curve
// Y^2=X^3+X*Z^4+b*Z^6
func (p *G1Jac) IsOnCurve() bool {
	var left, right, tmp, u fp.Element
	left.Square(&p.Y)
	right.Square(&p.X).Mul(&right, &p.X)
	tmp.Square(&p.Z).
		Square(&tmp)
	u.Mul(&p.X, &tmp)
	right.Add(&u, &right)
	tmp.Mul(&tmp, &p.Z).
		Mul(&tmp, &p.Z).
		Mul(&tmp, &bCurveCoeff)
	right.Add(&right, &tmp)
	return left.Equal(&right)
}

// IsInSubGroup returns true if p is on the r-torsion, false otherwise.
// the curve is of prime order i.e. E(𝔽p) is the full group
// so we just check that the point is on the curve.
func (p *G1Jac) IsInSubGroup() bool {

	return p.IsOnCurve()

}

// mulWindowed computes the 2-bits windowed double-and-add scalar
// multiplication p=[s]q in Jacobian coordinates.
func (p *G1Jac) mulWindowed(a *G1Jac, s *big.Int) *G1Jac {

	var res G1Jac
	var ops [3]G1Jac

	res.Set(&g1Infinity)
	ops[0].Set(a)
	ops[1].Double(&ops[0])
	ops[2].Set(&ops[0]).AddAssign(&ops[1])

	b := s.Bytes()
	for i := range b {
		w := b[i]
		mask := byte(0xc0)
		for j := 0; j < 4; j++ {
			res.DoubleAssign().DoubleAssign()
			c := (w & mask) >> (6 - 2*j)
			if c != 0 {
				res.AddAssign(&ops[c-1])
			}
			mask = mask >> 2
		}
	}
	p.Set(&res)

	return p

}

// JointScalarMultiplicationBase computes [s1]g+[s2]a using Straus-Shamir technique
// where g is the prime subgroup generator
func (p *G1Jac) JointScalarMultiplicationBase(a *G1Affine, s1, s2 *big.Int) *G1Jac {

	var res, p1, p2 G1Jac
	res.Set(&g1Infinity)
	p1.Set(&g1Gen)
	p2.FromAffine(a)

	var table [15]G1Jac

	var k1, k2 big.Int
	if s1.Sign() == -1 {
		k1.Neg(s1)
		table[0].Neg(&p1)
	} else {
		k1.Set(s1)
		table[0].Set(&p1)
	}
	if s2.Sign() == -1 {
		k2.Neg(s2)
		table[3].Neg(&p2)
	} else {
		k2.Set(s2)
		table[3].Set(&p2)
	}

	// precompute table (2 bits sliding window)
	table[1].Double(&table[0])
	table[2].Set(&table[1]).AddAssign(&table[0])
	table[4].Set(&table[3]).AddAssign(&table[0])
	table[5].Set(&table[3]).AddAssign(&table[1])
	table[6].Set(&table[3]).AddAssign(&table[2])
	table[7].Double(&table[3])
	table[8].Set(&table[7]).AddAssign(&table[0])
	table[9].Set(&table[7]).AddAssign(&table[1])
	table[10].Set(&table[7]).AddAssign(&table[2])
	table[11].Set(&table[7]).AddAssign(&table[3])
	table[12].Set(&table[11]).AddAssign(&table[0])
	table[13].Set(&table[11]).AddAssign(&table[1])
	table[14].Set(&table[11]).AddAssign(&table[2])

	var s [2]fr.Element
	s[0] = s[0].SetBigInt(&k1).Bits()
	s[1] = s[1].SetBigInt(&k2).Bits()

	maxBit := k1.BitLen()
	if k2.BitLen() > maxBit {
		maxBit = k2.BitLen()
	}
	hiWordIndex := (maxBit - 1) / 64

	for i := hiWordIndex; i >= 0; i-- {
		mask := uint64(3) << 62
		for j := 0; j < 32; j++ {
			res.Double(&res).Double(&res)
			b1 := (s[0][i] & mask) >> (62 - 2*j)
			b2 := (s[1][i] & mask) >> (62 - 2*j)
			if b1|b2 != 0 {
				s := (b2<<2 | b1)
				res.AddAssign(&table[s-1])
			}
			mask = mask >> 2
		}
	}

	p.Set(&res)
	return p

}

// JointScalarMultiplication computes [s1]p1+[s2]p1 using Straus-Shamir technique
// where g is the prime subgroup generator
func (p *G1Jac) JointScalarMultiplication(p1, p2 *G1Jac, s1, s2 *big.Int) *G1Jac {

	var res G1Jac
	res.Set(&g1Infinity)

	var table [15]G1Jac

	var k1, k2 big.Int
	if s1.Sign() == -1 {
		k1.Neg(s1)
		table[0].Neg(p1)
	} else {
		k1.Set(s1)
		table[0].Set(p1)
	}
	if s2.Sign() == -1 {
		k2.Neg(s2)
		table[3].Neg(p2)
	} else {
		k2.Set(s2)
		table[3].Set(p2)
	}

	// precompute table (2 bits sliding window)
	table[1].Double(&table[0])
	table[2].Set(&table[1]).AddAssign(&table[0])
	table[4].Set(&table[3]).AddAssign(&table[0])
	table[5].Set(&table[3]).AddAssign(&table[1])
	table[6].Set(&table[3]).AddAssign(&table[2])
	table[7].Double(&table[3])
	table[8].Set(&table[7]).AddAssign(&table[0])
	table[9].Set(&table[7]).AddAssign(&table[1])
	table[10].Set(&table[7]).AddAssign(&table[2])
	table[11].Set(&table[7]).AddAssign(&table[3])
	table[12].Set(&table[11]).AddAssign(&table[0])
	table[13].Set(&table[11]).AddAssign(&table[1])
	table[14].Set(&table[11]).AddAssign(&table[2])

	var s [2]fr.Element
	s[0] = s[0].SetBigInt(&k1).Bits()
	s[1] = s[1].SetBigInt(&k2).Bits()

	maxBit := k1.BitLen()
	if k2.BitLen() > maxBit {
		maxBit = k2.BitLen()
	}
	hiWordIndex := (maxBit - 1) / 64

	for i := hiWordIndex; i >= 0; i-- {
		mask := uint64(3) << 62
		for j := 0; j < 32; j++ {
			res.Double(&res).Double(&res)
			b1 := (s[0][i] & mask) >> (62 - 2*j)
			b2 := (s[1][i] & mask) >> (62 - 2*j)
			if b1|b2 != 0 {
				s := (b2<<2 | b1)
				res.AddAssign(&table[s-1])
			}
			mask = mask >> 2
		}
	}

	p.Set(&res)
	return p

}

// -------------------------------------------------------------------------------------------------
// extended Jacobian coordinates

// Set sets p to a in extended Jacobian coordinates.
func (p *g1JacExtended) Set(a *g1JacExtended) *g1JacExtended {
	p.X, p.Y, p.ZZ, p.ZZZ = a.X, a.Y, a.ZZ, a.ZZZ
	return p
}

// fromJacExtended converts an extended Jacobian point to an affine point.
func (p *G1Affine) fromJacExtended(Q *g1JacExtended) *G1Affine {
	if Q.ZZ.IsZero() {
		p.X = fp.Element{}
		p.Y = fp.Element{}
		return p
	}
	p.X.Inverse(&Q.ZZ).Mul(&p.X, &Q.X)
	p.Y.Inverse(&Q.ZZZ).Mul(&p.Y, &Q.Y)
	return p
}

// fromJacExtended converts an extended Jacobian point to a Jacobian point.
func (p *G1Jac) fromJacExtended(Q *g1JacExtended) *G1Jac {
	if Q.ZZ.IsZero() {
		p.Set(&g1Infinity)
		return p
	}
	p.X.Mul(&Q.ZZ, &Q.X).Mul(&p.X, &Q.ZZ)
	p.Y.Mul(&Q.ZZZ, &Q.Y).Mul(&p.Y, &Q.ZZZ)
	p.Z.Set(&Q.ZZZ)
	return p
}

// add sets p to p+q in extended Jacobian coordinates.
//
// https://www.hyperelliptic.org/EFD/g1p/auto-shortw-xyzz.html#addition-add-2008-s
func (p *g1JacExtended) add(q *g1JacExtended) *g1JacExtended {
	//if q is infinity return p
	if q.ZZ.IsZero() {
		return p
	}
	// p is infinity, return q
	if p.ZZ.IsZero() {
		p.Set(q)
		return p
	}

	var A, B, U1, U2, S1, S2 fp.Element

	// p2: q, p1: p
	U2.Mul(&q.X, &p.ZZ)
	U1.Mul(&p.X, &q.ZZ)
	A.Sub(&U2, &U1)
	S2.Mul(&q.Y, &p.ZZZ)
	S1.Mul(&p.Y, &q.ZZZ)
	B.Sub(&S2, &S1)

	if A.IsZero() {
		if B.IsZero() {
			return p.double(q)

		}
		p.ZZ = fp.Element{}
		p.ZZZ = fp.Element{}
		return p
	}

	var P, R, PP, PPP, Q, V fp.Element
	P.Sub(&U2, &U1)
	R.Sub(&S2, &S1)
	PP.Square(&P)
	PPP.Mul(&P, &PP)
	Q.Mul(&U1, &PP)
	V.Mul(&S1, &PPP)

	p.X.Square(&R).
		Sub(&p.X, &PPP).
		Sub(&p.X, &Q).
		Sub(&p.X, &Q)
	p.Y.Sub(&Q, &p.X).
		Mul(&p.Y, &R).
		Sub(&p.Y, &V)
	p.ZZ.Mul(&p.ZZ, &q.ZZ).
		Mul(&p.ZZ, &PP)
	p.ZZZ.Mul(&p.ZZZ, &q.ZZZ).
		Mul(&p.ZZZ, &PPP)

	return p
}

// double sets p to [2]q in Jacobian extended coordinates.
//
// http://www.hyperelliptic.org/EFD/g1p/auto-shortw-xyzz.html#doubling-dbl-2008-s-1
// N.B.: since we consider any point on Z=0 as the point at infinity
// this doubling formula works for infinity points as well.
func (p *g1JacExtended) double(q *g1JacExtended) *g1JacExtended {
	var Z, U, V, W, S, XX, M fp.Element

	U.Double(&q.Y)
	V.Square(&U)
	W.Mul(&U, &V)
	S.Mul(&q.X, &V)
	XX.Square(&q.X)
	M.Double(&XX).
		Add(&M, &XX)
	Z.Square(&q.ZZ)
	M.Add(&M, &Z)
	U.Mul(&W, &q.Y)

	p.X.Square(&M).
		Sub(&p.X, &S).
		Sub(&p.X, &S)
	p.Y.Sub(&S, &p.X).
		Mul(&p.Y, &M).
		Sub(&p.Y, &U)
	p.ZZ.Mul(&V, &q.ZZ)
	p.ZZZ.Mul(&W, &q.ZZZ)

	return p
}

// subMixed is the same as addMixed, but negates a.Y.
//
// http://www.hyperelliptic.org/EFD/g1p/auto-shortw-xyzz.html#addition-madd-2008-s
func (p *g1JacExtended) subMixed(a *G1Affine) *g1JacExtended {

	//if a is infinity return p
	if a.IsInfinity() {
		return p
	}
	// p is infinity, return a
	if p.ZZ.IsZero() {
		p.X = a.X
		p.Y.Neg(&a.Y)
		p.ZZ.SetOne()
		p.ZZZ.SetOne()
		return p
	}

	var P, R fp.Element

	// p2: a, p1: p
	P.Mul(&a.X, &p.ZZ)
	P.Sub(&P, &p.X)

	R.Mul(&a.Y, &p.ZZZ)
	R.Neg(&R)
	R.Sub(&R, &p.Y)

	if P.IsZero() {
		if R.IsZero() {
			return p.doubleNegMixed(a)

		}
		p.ZZ = fp.Element{}
		p.ZZZ = fp.Element{}
		return p
	}

	var PP, PPP, Q, Q2, RR, X3, Y3 fp.Element

	PP.Square(&P)
	PPP.Mul(&P, &PP)
	Q.Mul(&p.X, &PP)
	RR.Square(&R)
	X3.Sub(&RR, &PPP)
	Q2.Double(&Q)
	p.X.Sub(&X3, &Q2)
	Y3.Sub(&Q, &p.X).Mul(&Y3, &R)
	R.Mul(&p.Y, &PPP)
	p.Y.Sub(&Y3, &R)
	p.ZZ.Mul(&p.ZZ, &PP)
	p.ZZZ.Mul(&p.ZZZ, &PPP)

	return p

}

// addMixed sets p to p+q in extended Jacobian coordinates, where a.ZZ=1.
//
// http://www.hyperelliptic.org/EFD/g1p/auto-shortw-xyzz.html#addition-madd-2008-s
func (p *g1JacExtended) addMixed(a *G1Affine) *g1JacExtended {

	//if a is infinity return p
	if a.IsInfinity() {
		return p
	}
	// p is infinity, return a
	if p.ZZ.IsZero() {
		p.X = a.X
		p.Y = a.Y
		p.ZZ.SetOne()
		p.ZZZ.SetOne()
		return p
	}

	var P, R fp.Element

	// p2: a, p1: p
	P.Mul(&a.X, &p.ZZ)
	P.Sub(&P, &p.X)

	R.Mul(&a.Y, &p.ZZZ)
	R.Sub(&R, &p.Y)

	if P.IsZero() {
		if R.IsZero() {
			return p.doubleMixed(a)

		}
		p.ZZ = fp.Element{}
		p.ZZZ = fp.Element{}
		return p
	}

	var PP, PPP, Q, Q2, RR, X3, Y3 fp.Element

	PP.Square(&P)
	PPP.Mul(&P, &PP)
	Q.Mul(&p.X, &PP)
	RR.Square(&R)
	X3.Sub(&RR, &PPP)
	Q2.Double(&Q)
	p.X.Sub(&X3, &Q2)
	Y3.Sub(&Q, &p.X).Mul(&Y3, &R)
	R.Mul(&p.Y, &PPP)
	p.Y.Sub(&Y3, &R)
	p.ZZ.Mul(&p.ZZ, &PP)
	p.ZZZ.Mul(&p.ZZZ, &PPP)

	return p

}

// doubleNegMixed works the same as double, but negates q.Y.
func (p *g1JacExtended) doubleNegMixed(q *G1Affine) *g1JacExtended {

	var Z, U, V, W, S, XX, M, S2, L fp.Element

	U.Double(&q.Y)
	U.Neg(&U)
	V.Square(&U)
	W.Mul(&U, &V)
	S.Mul(&q.X, &V)
	XX.Square(&q.X)
	M.Double(&XX).
		Add(&M, &XX)
	Z.Square(&p.ZZ)
	M.Add(&M, &Z)
	S2.Double(&S)
	L.Mul(&W, &q.Y)

	p.X.Square(&M).
		Sub(&p.X, &S2)
	p.Y.Sub(&S, &p.X).
		Mul(&p.Y, &M).
		Add(&p.Y, &L)
	p.ZZ.Set(&V)
	p.ZZZ.Set(&W)

	return p
}

// doubleMixed sets p to [2]a in Jacobian extended coordinates, where a.ZZ=1.
//
// http://www.hyperelliptic.org/EFD/g1p/auto-shortw-xyzz.html#doubling-dbl-2008-s-1
func (p *g1JacExtended) doubleMixed(q *G1Affine) *g1JacExtended {

	var Z, U, V, W, S, XX, M, S2, L fp.Element

	U.Double(&q.Y)
	V.Square(&U)
	W.Mul(&U, &V)
	S.Mul(&q.X, &V)
	XX.Square(&q.X)
	M.Double(&XX).
		Add(&M, &XX)
	Z.Square(&p.ZZ)
	M.Add(&M, &Z)
	S2.Double(&S)
	L.Mul(&W, &q.Y)

	p.X.Square(&M).
		Sub(&p.X, &S2)
	p.Y.Sub(&S, &p.X).
		Mul(&p.Y, &M).
		Sub(&p.Y, &L)
	p.ZZ.Set(&V)
	p.ZZZ.Set(&W)

	return p
}

// BatchJacobianToAffineG1 converts points in Jacobian coordinates to Affine coordinates
// performing a single field inversion using the Montgomery batch inversion trick.
func BatchJacobianToAffineG1(points []G1Jac) []G1Affine {
	result := make([]G1Affine, len(points))
	zeroes := make([]bool, len(points))
	accumulator := fp.One()

	// batch invert all points[].Z coordinates with Montgomery batch inversion trick
	// (stores points[].Z^-1 in result[i].X to avoid allocating a slice of fr.Elements)
	for i := 0; i < len(points); i++ {
		if points[i].Z.IsZero() {
			zeroes[i] = true
			continue
		}
		result[i].X = accumulator
		accumulator.Mul(&accumulator, &points[i].Z)
	}

	var accInverse fp.Element
	accInverse.Inverse(&accumulator)

	for i := len(points) - 1; i >= 0; i-- {
		if zeroes[i] {
			// do nothing, (X=0, Y=0) is infinity point in affine
			continue
		}
		result[i].X.Mul(&result[i].X, &accInverse)
		accInverse.Mul(&accInverse, &points[i].Z)
	}

	// batch convert to affine.
	parallel.Execute(len(points), func(start, end int) {
		for i := start; i < end; i++ {
			if zeroes[i] {
				// do nothing, (X=0, Y=0) is infinity point in affine
				continue
			}
			var a, b fp.Element
			a = result[i].X
			b.Square(&a)
			result[i].X.Mul(&points[i].X, &b)
			result[i].Y.Mul(&points[i].Y, &b).
				Mul(&result[i].Y, &a)
		}
	})

	return result
}
