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

package bw6633

import (
	"errors"

	"github.com/consensys/gnark-crypto/ecc/bw6-633/fp"
	"github.com/consensys/gnark-crypto/ecc/bw6-633/internal/fptower"
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

	// hard part exponent: 3(u+1)(p^2-p+1)/r
	var m [11]GT
	var f10, _m1, _m3, _m4, _m5, _m7, _m8, _m8m5, _m6, f11, f11f10, f12, f1, f1u, f1q, f1a GT
	m[0].Set(&result)
	for i := 1; i < 11; i++ {
		m[i].Expt(&m[i-1])
	}
	result.Mul(&m[3], &m[1]).
		Conjugate(&result).
		Mul(&result, &m[2]).
		Mul(&result, &m[0]).
		CyclotomicSquare(&result).
		Mul(&result, &m[4])
	buf.Frobenius(&m[0]).Conjugate(&buf)
	result.Mul(&result, &buf)
	buf.CyclotomicSquare(&result).
		CyclotomicSquare(&buf).
		CyclotomicSquare(&buf)
	result.Mul(&result, &buf)
	_m1.Conjugate(&m[1])
	_m3.Conjugate(&m[3])
	_m4.Conjugate(&m[4])
	_m5.Conjugate(&m[5])
	_m7.Conjugate(&m[7])
	f10.Mul(&m[4], &_m3).
		CyclotomicSquare(&f10).
		Mul(&f10, &m[2]).
		Mul(&f10, &m[6]).
		Mul(&f10, &_m5).
		CyclotomicSquare(&f10).
		Mul(&f10, &_m1).
		Mul(&f10, &_m5).
		Mul(&f10, &_m7).
		CyclotomicSquare(&f10).
		Mul(&f10, &m[0]).
		Mul(&f10, &m[2]).
		Mul(&f10, &m[3]).
		Mul(&f10, &_m1).
		CyclotomicSquare(&f10).
		Mul(&f10, &m[0]).
		Mul(&f10, &m[8]).
		Mul(&f10, &_m4)
	_m8.Conjugate(&m[8])
	_m6.Conjugate(&m[6])
	_m8m5.Mul(&m[5], &_m8)
	f11.Mul(&m[7], &_m6).
		CyclotomicSquare(&f11).
		Mul(&f11, &m[2]).
		Mul(&f11, &_m3).
		Mul(&f11, &_m8m5).
		CyclotomicSquare(&f11).
		Mul(&f11, &_m8m5).
		Mul(&f11, &m[9]).
		Mul(&f11, &_m1)
	buf.CyclotomicSquare(&f11)
	f11.Mul(&buf, &f11)
	f11f10.Mul(&f11, &f10)
	buf.CyclotomicSquare(&f11f10)
	f11f10.Mul(&f11f10, &buf)
	f12.Mul(&m[0], &m[1]).
		Mul(&f12, &m[2]).
		Mul(&f12, &m[8]).
		Mul(&f12, &m[10])
	buf.CyclotomicSquare(&m[5])
	f12.Mul(&f12, &buf)
	buf.CyclotomicSquare(&m[9]).
		Mul(&buf, &m[6]).
		Mul(&buf, &m[4]).
		Conjugate(&buf)
	f12.Mul(&f12, &buf)
	buf.CyclotomicSquare(&f12). // cyclo exp by 13: (ht**2+3*hy**2)//4
					Mul(&buf, &f12).
					CyclotomicSquare(&buf).
					CyclotomicSquare(&buf)
	f12.Mul(&f12, &buf)
	f1.Mul(&f11f10, &f12)
	f1u.Expt(&f1)
	f1q.Mul(&f1u, &f1).
		Frobenius(&f1q)
	f1a.Conjugate(&f1u).
		Mul(&f1a, &f1).
		Expt(&f1a).
		Expt(&f1a).
		Expt(&f1a).
		Expt(&f1a)
	f1.Conjugate(&f1)
	f1a.Mul(&f1a, &f1)

	result.Mul(&result, &f1a).
		Mul(&result, &f1q)

	return result
}

// MillerLoop Optimal Tate (or twisted ate or Eta revisited)
// Alg.2 in https://eprint.iacr.org/2021/1359.pdf
func MillerLoop(P []G1Affine, Q []G2Affine) (GT, error) {
	// check input size match
	n := len(P)
	if n == 0 || n != len(Q) {
		return GT{}, errors.New("invalid inputs sizes")
	}

	// filter infinity points
	p0 := make([]G1Affine, 0, n)
	q := make([]G2Affine, 0, n)

	for k := 0; k < n; k++ {
		if P[k].IsInfinity() || Q[k].IsInfinity() {
			continue
		}
		p0 = append(p0, P[k])
		q = append(q, Q[k])
	}

	n = len(q)

	// precomputations
	pProj0 := make([]g1Proj, n)
	p1 := make([]G1Affine, n)
	p01 := make([]G1Affine, n)
	p10 := make([]G1Affine, n)
	pProj01 := make([]g1Proj, n) // P0+P1
	pProj10 := make([]g1Proj, n) // P0-P1
	l01 := make([]lineEvaluation, n)
	l10 := make([]lineEvaluation, n)
	for k := 0; k < n; k++ {
		p1[k].Y.Set(&p0[k].Y)
		p1[k].X.Mul(&p0[k].X, &thirdRootOneG1)
		p0[k].Neg(&p0[k])
		pProj0[k].FromAffine(&p0[k])

		// l_{p0,p1}(q)
		pProj01[k].Set(&pProj0[k])
		pProj01[k].AddMixedStep(&l01[k], &p1[k])
		l01[k].r1.Mul(&l01[k].r1, &q[k].X)
		l01[k].r0.Mul(&l01[k].r0, &q[k].Y)

		// l_{-p0,p1}(q)
		pProj10[k].Neg(&pProj0[k])
		pProj10[k].AddMixedStep(&l10[k], &p1[k])
		l10[k].r1.Mul(&l10[k].r1, &q[k].X)
		l10[k].r0.Mul(&l10[k].r0, &q[k].Y)
	}
	BatchProjectiveToAffineG1(pProj01, p01)
	BatchProjectiveToAffineG1(pProj10, p10)

	// f_{a0+lambda*a1,P}(Q)
	var result, ss GT
	result.SetOne()
	var l, l0 lineEvaluation

	var j int8

	// i = 157
	for k := 0; k < n; k++ {
		pProj0[k].DoubleStep(&l0)
		l0.r1.Mul(&l0.r1, &q[k].X)
		l0.r0.Mul(&l0.r0, &q[k].Y)
		result.MulBy034(&l0.r0, &l0.r1, &l0.r2)
	}

	var tmp G1Affine
	for i := 156; i >= 0; i-- {
		result.Square(&result)

		j = loopCounter0[i]*3 + loopCounter1[i]

		for k := 0; k < n; k++ {
			pProj0[k].DoubleStep(&l0)
			l0.r1.Mul(&l0.r1, &q[k].X)
			l0.r0.Mul(&l0.r0, &q[k].Y)

			switch j {
			case -4:
				tmp.Neg(&p01[k])
				pProj0[k].AddMixedStep(&l, &tmp)
				l.r1.Mul(&l.r1, &q[k].X)
				l.r0.Mul(&l.r0, &q[k].Y)
				ss.Mul034By034(&l.r0, &l.r1, &l.r2, &l01[k].r0, &l01[k].r1, &l01[k].r2)
				result.MulBy034(&l0.r0, &l0.r1, &l0.r2).
					Mul(&result, &ss)
			case -3:
				tmp.Neg(&p1[k])
				pProj0[k].AddMixedStep(&l, &tmp)
				l.r1.Mul(&l.r1, &q[k].X)
				l.r0.Mul(&l.r0, &q[k].Y)
				ss.Mul034By034(&l.r0, &l.r1, &l.r2, &l0.r0, &l0.r1, &l0.r2)
				result.Mul(&result, &ss)
			case -2:
				pProj0[k].AddMixedStep(&l, &p10[k])
				l.r1.Mul(&l.r1, &q[k].X)
				l.r0.Mul(&l.r0, &q[k].Y)
				ss.Mul034By034(&l.r0, &l.r1, &l.r2, &l10[k].r0, &l10[k].r1, &l10[k].r2)
				result.MulBy034(&l0.r0, &l0.r1, &l0.r2).
					Mul(&result, &ss)
			case -1:
				tmp.Neg(&p0[k])
				pProj0[k].AddMixedStep(&l, &tmp)
				l.r1.Mul(&l.r1, &q[k].X)
				l.r0.Mul(&l.r0, &q[k].Y)
				ss.Mul034By034(&l.r0, &l.r1, &l.r2, &l0.r0, &l0.r1, &l0.r2)
				result.Mul(&result, &ss)
			case 0:
				result.MulBy034(&l0.r0, &l0.r1, &l0.r2)
			case 1:
				pProj0[k].AddMixedStep(&l, &p0[k])
				l.r1.Mul(&l.r1, &q[k].X)
				l.r0.Mul(&l.r0, &q[k].Y)
				ss.Mul034By034(&l.r0, &l.r1, &l.r2, &l0.r0, &l0.r1, &l0.r2)
				result.Mul(&result, &ss)
			case 2:
				tmp.Neg(&p10[k])
				pProj0[k].AddMixedStep(&l, &tmp)
				l.r1.Mul(&l.r1, &q[k].X)
				l.r0.Mul(&l.r0, &q[k].Y)
				ss.Mul034By034(&l.r0, &l.r1, &l.r2, &l10[k].r0, &l10[k].r1, &l10[k].r2)
				result.MulBy034(&l0.r0, &l0.r1, &l0.r2).
					Mul(&result, &ss)
			case 3:
				pProj0[k].AddMixedStep(&l, &p1[k])
				l.r1.Mul(&l.r1, &q[k].X)
				l.r0.Mul(&l.r0, &q[k].Y)
				ss.Mul034By034(&l.r0, &l.r1, &l.r2, &l0.r0, &l0.r1, &l0.r2)
				result.Mul(&result, &ss)
			case 4:
				pProj0[k].AddMixedStep(&l, &p01[k])
				l.r1.Mul(&l.r1, &q[k].X)
				l.r0.Mul(&l.r0, &q[k].Y)
				ss.Mul034By034(&l.r0, &l.r1, &l.r2, &l01[k].r0, &l01[k].r1, &l01[k].r2)
				result.MulBy034(&l0.r0, &l0.r1, &l0.r2).
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
func (p *g1Proj) DoubleStep(evaluations *lineEvaluation) {

	// get some Element from our pool
	var t1, A, B, C, D, E, EE, F, G, H, I, J, K fp.Element
	A.Mul(&p.x, &p.y)
	A.Halve()
	B.Square(&p.y)
	C.Square(&p.z)
	D.Double(&C).
		Add(&D, &C)

		// E.Mul(&D, &bCurveCoeff)
	E.Double(&D).
		Double(&E)

	F.Double(&E).
		Add(&F, &E)
	G.Add(&B, &F)
	G.Halve()
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
	evaluations.r0.Neg(&H)
	evaluations.r1.Double(&J).
		Add(&evaluations.r1, &J)
	evaluations.r2.Set(&I)
}

// AddMixedStep point addition in Mixed Homogenous projective and Affine coordinates
// https://eprint.iacr.org/2013/722.pdf (Section 4.3)
func (p *g1Proj) AddMixedStep(evaluations *lineEvaluation, a *G1Affine) {

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
	evaluations.r0.Set(&L)
	evaluations.r1.Neg(&O)
	evaluations.r2.Set(&J)
}
