// Copyright 2020 ConsenSys AG
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

package bw6761

import (
	"errors"

	"github.com/consensys/gnark-crypto/ecc/bw6-761/fp"
	"github.com/consensys/gnark-crypto/ecc/bw6-761/internal/fptower"
)

// GT target group of the pairing
type GT = fptower.E6

type lineEvaluation struct {
	r0 fp.Element
	r1 fp.Element
	r2 fp.Element
}

// Pair calculates the reduced pairing for a set of points
func Pair(P []G1Affine, Q []G2Affine) (GT, error) {
	f, err := MillerLoop(P, Q)
	if err != nil {
		return GT{}, err
	}
	return FinalExponentiation(&f), nil
}

// PairingCheck calculates the reduced pairing for a set of points and returns True if the result is One
func PairingCheck(P []G1Affine, Q []G2Affine) (bool, error) {
	f, err := Pair(P, Q)
	if err != nil {
		return false, err
	}
	var one GT
	one.SetOne()
	return f.Equal(&one), nil
}

// FinalExponentiation computes the final expo x**(c*(p**3-1)(p+1)(p**2-p+1)/r)
func FinalExponentiation(z *GT, _z ...*GT) GT {

	var result GT
	result.Set(z)

	for _, e := range _z {
		result.Mul(&result, e)
	}

	var buf GT

	// easy part exponent: (p**3 - 1)*(p+1)
	buf.Conjugate(&result)
	result.Inverse(&result)
	buf.Mul(&buf, &result)
	result.Frobenius(&buf).
		Mul(&result, &buf)

		// hard part exponent: 12(u+1)(p**2 - p + 1)/r
	var m1, _m1, m2, _m2, m3, f0, f0_36, g0, g1, _g1, g2, g3, _g3, g4, _g4, g5, _g5, g6, gA, gB, g034, _g1g2, gC, h1, h2, h2g2C, h4 GT
	m1.Expt(&result)
	_m1.Conjugate(&m1)
	m2.Expt(&m1)
	_m2.Conjugate(&m2)
	m3.Expt(&m2)
	f0.Frobenius(&result).
		Mul(&f0, &result).
		Mul(&f0, &m2)
	m2.CyclotomicSquare(&_m1)
	f0.Mul(&f0, &m2)
	f0_36.CyclotomicSquare(&f0).
		CyclotomicSquare(&f0_36).
		CyclotomicSquare(&f0_36).
		Mul(&f0_36, &f0).
		CyclotomicSquare(&f0_36).
		CyclotomicSquare(&f0_36)
	g0.Mul(&result, &m1).
		Frobenius(&g0).
		Mul(&g0, &m3).
		Mul(&g0, &_m2).
		Mul(&g0, &_m1)
	g1.Expt(&g0)
	_g1.Conjugate(&g1)
	g2.Expt(&g1)
	g3.Expt(&g2)
	_g3.Conjugate(&g3)
	g4.Expt(&g3)
	_g4.Conjugate(&g4)
	g5.Expt(&g4)
	_g5.Conjugate(&g5)
	g6.Expt(&g5)
	gA.Mul(&g3, &_g5).
		CyclotomicSquare(&gA).
		Mul(&gA, &g6).
		Mul(&gA, &g1).
		Mul(&gA, &g0)
	g034.Mul(&g0, &g3).
		Mul(&g034, &_g4)
	gB.CyclotomicSquare(&g034).
		Mul(&gB, &g034).
		Mul(&gB, &g5).
		Mul(&gB, &_g1)
	_g1g2.Mul(&_g1, &g2)
	gC.Mul(&_g3, &_g1g2).
		CyclotomicSquare(&gC).
		Mul(&gC, &_g1g2).
		Mul(&gC, &g0).
		CyclotomicSquare(&gC).
		Mul(&gC, &g2).
		Mul(&gC, &g0).
		Mul(&gC, &g4)
		// ht, hy = 13, 9
		// c1 = ht**2+3*hy**2 = 412
	h1.Expc1(&gA)
	// c2 = ht+hy = 22
	h2.Expc2(&gB)
	h2g2C.CyclotomicSquare(&gC).
		Mul(&h2g2C, &h2)
	h4.CyclotomicSquare(&h2g2C).
		Mul(&h4, &h2g2C).
		CyclotomicSquare(&h4)
	result.Mul(&h1, &h4).
		Mul(&result, &f0_36)

	return result
}

// MillerLoop Miller loop
func MillerLoop(P []G1Affine, Q []G2Affine) (GT, error) {
	// check input size match
	n := len(P)
	if n == 0 || n != len(Q) {
		return GT{}, errors.New("invalid inputs sizes")
	}

	// filter infinity points
	p := make([]G1Affine, 0, n)
	q0 := make([]G2Affine, 0, n)

	for k := 0; k < n; k++ {
		if P[k].IsInfinity() || Q[k].IsInfinity() {
			continue
		}
		p = append(p, P[k])
		q0 = append(q0, Q[k])
	}

	n = len(p)

	// precomputations
	qProj1 := make([]g2Proj, n)
	q1 := make([]G2Affine, n)
	q01 := make([]G2Affine, n)
	q10 := make([]G2Affine, n)
	qProj01 := make([]g2Proj, n)
	qProj10 := make([]g2Proj, n)
	l01 := make([]lineEvaluation, n)
	l10 := make([]lineEvaluation, n)
	for k := 0; k < n; k++ {
		q1[k].Y.Neg(&q0[k].Y)
		q1[k].X.Mul(&q0[k].X, &thirdRootOneG1)
		qProj1[k].FromAffine(&q1[k])

		// l_{q0,q1}(p)
		qProj01[k].Set(&qProj1[k])
		qProj01[k].AddMixedStep(&l01[k], &q0[k])
		l01[k].r1.Mul(&l01[k].r1, &p[k].X)
		l01[k].r2.Mul(&l01[k].r2, &p[k].Y)

		// l_{q0,-q1}(q)
		qProj10[k].Neg(&qProj1[k])
		qProj10[k].AddMixedStep(&l10[k], &q0[k])
		l10[k].r1.Mul(&l10[k].r1, &p[k].X)
		l10[k].r2.Mul(&l10[k].r2, &p[k].Y)
	}
	BatchProjectiveToAffineG2(qProj01, q01)
	BatchProjectiveToAffineG2(qProj10, q10)

	// f_{a0+lambda*a1,Q}(P)
	var result, ss GT
	result.SetOne()
	var l, l0 lineEvaluation

	var j int8

	// i = 188
	for k := 0; k < n; k++ {
		qProj1[k].DoubleStep(&l0)
		l0.r1.Mul(&l0.r1, &p[k].X)
		l0.r2.Mul(&l0.r2, &p[k].Y)
		result.MulBy014(&l0.r0, &l0.r1, &l0.r2)
	}

	var tmp G2Affine
	for i := 187; i >= 0; i-- {
		result.Square(&result)

		j = loopCounter1[i]*3 + loopCounter0[i]

		for k := 0; k < n; k++ {
			qProj1[k].DoubleStep(&l0)
			l0.r1.Mul(&l0.r1, &p[k].X)
			l0.r2.Mul(&l0.r2, &p[k].Y)

			switch j {
			case -4:
				tmp.Neg(&q01[k])
				qProj1[k].AddMixedStep(&l, &tmp)
				l.r1.Mul(&l.r1, &p[k].X)
				l.r2.Mul(&l.r2, &p[k].Y)
				ss.Mul014By014(&l.r0, &l.r1, &l.r2, &l01[k].r0, &l01[k].r1, &l01[k].r2)
				result.MulBy014(&l0.r0, &l0.r1, &l0.r2).
					Mul(&result, &ss)
			case -3:
				tmp.Neg(&q1[k])
				qProj1[k].AddMixedStep(&l, &tmp)
				l.r1.Mul(&l.r1, &p[k].X)
				l.r2.Mul(&l.r2, &p[k].Y)
				ss.Mul014By014(&l.r0, &l.r1, &l.r2, &l0.r0, &l0.r1, &l0.r2)
				result.Mul(&result, &ss)
			case -2:
				qProj1[k].AddMixedStep(&l, &q10[k])
				l.r1.Mul(&l.r1, &p[k].X)
				l.r2.Mul(&l.r2, &p[k].Y)
				ss.Mul014By014(&l.r0, &l.r1, &l.r2, &l10[k].r0, &l10[k].r1, &l10[k].r2)
				result.MulBy014(&l0.r0, &l0.r1, &l0.r2).
					Mul(&result, &ss)
			case -1:
				tmp.Neg(&q0[k])
				qProj1[k].AddMixedStep(&l, &tmp)
				l.r1.Mul(&l.r1, &p[k].X)
				l.r2.Mul(&l.r2, &p[k].Y)
				ss.Mul014By014(&l.r0, &l.r1, &l.r2, &l0.r0, &l0.r1, &l0.r2)
				result.Mul(&result, &ss)
			case 0:
				result.MulBy014(&l0.r0, &l0.r1, &l0.r2)
			case 1:
				qProj1[k].AddMixedStep(&l, &q0[k])
				l.r1.Mul(&l.r1, &p[k].X)
				l.r2.Mul(&l.r2, &p[k].Y)
				ss.Mul014By014(&l.r0, &l.r1, &l.r2, &l0.r0, &l0.r1, &l0.r2)
				result.Mul(&result, &ss)
			case 2:
				tmp.Neg(&q10[k])
				qProj1[k].AddMixedStep(&l, &tmp)
				l.r1.Mul(&l.r1, &p[k].X)
				l.r2.Mul(&l.r2, &p[k].Y)
				ss.Mul014By014(&l.r0, &l.r1, &l.r2, &l10[k].r0, &l10[k].r1, &l10[k].r2)
				result.MulBy014(&l0.r0, &l0.r1, &l0.r2).
					Mul(&result, &ss)
			case 3:
				qProj1[k].AddMixedStep(&l, &q1[k])
				l.r1.Mul(&l.r1, &p[k].X)
				l.r2.Mul(&l.r2, &p[k].Y)
				ss.Mul014By014(&l.r0, &l.r1, &l.r2, &l0.r0, &l0.r1, &l0.r2)
				result.Mul(&result, &ss)
			case 4:
				qProj1[k].AddMixedStep(&l, &q01[k])
				l.r1.Mul(&l.r1, &p[k].X)
				l.r2.Mul(&l.r2, &p[k].Y)
				ss.Mul014By014(&l.r0, &l.r1, &l.r2, &l01[k].r0, &l01[k].r1, &l01[k].r2)
				result.MulBy014(&l0.r0, &l0.r1, &l0.r2).
					Mul(&result, &ss)
			default:
				return GT{}, errors.New("invalid loopCounter")
			}
		}
	}

	return result, nil

}

// DoubleStep doubles a point in Homogenous projective coordinates, and evaluates the line in Miller loop
// https://eprint.iacr.org/2013/722.pdf (Section 4.3)
func (p *g2Proj) DoubleStep(evaluations *lineEvaluation) {

	// get some Element from our pool
	var t0, t1, A, B, C, D, E, EE, F, G, H, I, J, K fp.Element
	t0.Mul(&p.x, &p.y)
	A.Mul(&t0, &twoInv)
	B.Square(&p.y)
	C.Square(&p.z)
	D.Double(&C).
		Add(&D, &C)
	E.Mul(&D, &bTwistCurveCoeff)
	F.Double(&E).
		Add(&F, &E)
	G.Add(&B, &F)
	G.Mul(&G, &twoInv)
	H.Add(&p.y, &p.z).
		Square(&H)
	t1.Add(&B, &C)
	H.Sub(&H, &t1)
	I.Sub(&E, &B)
	J.Square(&p.x)
	EE.Square(&E)
	K.Double(&EE).
		Add(&K, &EE)

	// X, Y, Z
	p.x.Sub(&B, &F).
		Mul(&p.x, &A)
	p.y.Square(&G).
		Sub(&p.y, &K)
	p.z.Mul(&B, &H)

	// Line evaluation
	evaluations.r0.Set(&I)
	evaluations.r1.Double(&J).
		Add(&evaluations.r1, &J)
	evaluations.r2.Neg(&H)
}

// AddMixedStep point addition in Mixed Homogenous projective and Affine coordinates
// https://eprint.iacr.org/2013/722.pdf (Section 4.3)
func (p *g2Proj) AddMixedStep(evaluations *lineEvaluation, a *G2Affine) {

	// get some Element from our pool
	var Y2Z1, X2Z1, O, L, C, D, E, F, G, H, t0, t1, t2, J fp.Element
	Y2Z1.Mul(&a.Y, &p.z)
	O.Sub(&p.y, &Y2Z1)
	X2Z1.Mul(&a.X, &p.z)
	L.Sub(&p.x, &X2Z1)
	C.Square(&O)
	D.Square(&L)
	E.Mul(&L, &D)
	F.Mul(&p.z, &C)
	G.Mul(&p.x, &D)
	t0.Double(&G)
	H.Add(&E, &F).
		Sub(&H, &t0)
	t1.Mul(&p.y, &E)

	// X, Y, Z
	p.x.Mul(&L, &H)
	p.y.Sub(&G, &H).
		Mul(&p.y, &O).
		Sub(&p.y, &t1)
	p.z.Mul(&E, &p.z)

	t2.Mul(&L, &a.Y)
	J.Mul(&a.X, &O).
		Sub(&J, &t2)

	// Line evaluation
	evaluations.r0.Set(&J)
	evaluations.r1.Neg(&O)
	evaluations.r2.Set(&L)
}
