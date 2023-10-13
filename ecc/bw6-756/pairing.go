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

package bw6756

import (
	"errors"

	"github.com/consensys/gnark-crypto/ecc/bw6-756/fp"
	"github.com/consensys/gnark-crypto/ecc/bw6-756/internal/fptower"
)

// GT target group of the pairing
type GT = fptower.E6

type lineEvaluation struct {
	r0 fp.Element
	r1 fp.Element
	r2 fp.Element
}

// Pair calculates the reduced pairing for a set of points
// ∏ᵢ e(Pᵢ, Qᵢ).
//
// This function doesn't check that the inputs are in the correct subgroup. See IsInSubGroup.
func Pair(P []G1Affine, Q []G2Affine) (GT, error) {
	f, err := MillerLoop(P, Q)
	if err != nil {
		return GT{}, err
	}
	return FinalExponentiation(&f), nil
}

// PairingCheck calculates the reduced pairing for a set of points and returns True if the result is One
// ∏ᵢ e(Pᵢ, Qᵢ) =? 1
//
// This function doesn't check that the inputs are in the correct subgroup. See IsInSubGroup.
func PairingCheck(P []G1Affine, Q []G2Affine) (bool, error) {
	f, err := Pair(P, Q)
	if err != nil {
		return false, err
	}
	var one GT
	one.SetOne()
	return f.Equal(&one), nil
}

// FinalExponentiation computes the exponentiation (∏ᵢ zᵢ)ᵈ
// where d = (p^6-1)/r = (p^6-1)/Φ_6(p) ⋅ Φ_6(p)/r = (p^3-1)(p+1)(p^2 - p +1)/r
// we use instead d=s ⋅ (p^3-1)(p+1)(p^2 - p +1)/r
// where s is the cofactor (x_0+1) (El Housni and Guillevic)
func FinalExponentiation(z *GT, _z ...*GT) GT {

	var result GT
	result.Set(z)

	for _, e := range _z {
		result.Mul(&result, e)
	}

	var buf GT

	// Easy part
	// (p^3-1)(p+1)
	buf.Conjugate(&result)
	result.Inverse(&result)
	buf.Mul(&buf, &result)
	result.Frobenius(&buf).
		Mul(&result, &buf)

	var one GT
	one.SetOne()
	if result.Equal(&one) {
		return result
	}

	// 2. Hard part (up to permutation)
	// (x₀+1)(p²-p+1)/r
	// Algorithm 4.4 from https://yelhousni.github.io/phd.pdf
	var a, b, c, d, e, f, g, h, i, j, k, t GT
	a.ExptMinus1Square(&result)
	t.Frobenius(&result)
	a.Mul(&a, &t)
	b.ExptPlus1(&a)
	t.Conjugate(&result)
	b.Mul(&b, &t)
	t.CyclotomicSquare(&a)
	a.Mul(&a, &t)
	c.ExptMinus1Div3(&b)
	d.ExptMinus1(&c)
	e.ExptMinus1Square(&d)
	e.Mul(&e, &d)
	d.Conjugate(&d)
	f.Mul(&d, &b)
	g.ExptPlus1(&e)
	g.Mul(&g, &f)
	h.Mul(&g, &c)
	i.Mul(&g, &d)
	i.ExptPlus1(&i)
	t.Conjugate(&f)
	i.Mul(&i, &t)
	// ht, hy = -1, -1
	// c1 = (ht+hy)/2 = -1
	j.Conjugate(&h)
	j.Mul(&j, &e)
	k.CyclotomicSquare(&j)
	k.Mul(&k, &j)
	k.Mul(&k, &b)
	// c2 = (ht**2+3*hy**2)/4 = 1
	k.Mul(&k, &i)
	result.Mul(&a, &k)

	return result
}

// MillerLoop Optimal Tate alternative (or twisted ate or Eta revisited)
// computes the multi-Miller loop
// ∏ᵢ MillerLoop(Pᵢ, Qᵢ) =
// ∏ᵢ { fᵢ_{x₀+1+λ(x₀³-x₀²-x₀),Pᵢ}(Qᵢ) }
//
// Alg.2 in https://eprint.iacr.org/2021/1359.pdf
// Eq. (6) in https://hackmd.io/@gnark/BW6-761-changes
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
	pProj1 := make([]g1Proj, n)
	p1 := make([]G1Affine, n)
	for k := 0; k < n; k++ {
		p1[k].Y.Neg(&p0[k].Y)
		p1[k].X.Mul(&p0[k].X, &thirdRootOneG2)
		pProj1[k].FromAffine(&p1[k])
	}

	// f_{a0+λ*a1,P}(Q)
	var result GT
	result.SetOne()
	var l, l0 lineEvaluation
	var prodLines [5]fp.Element

	var j int8

	if n >= 1 {
		// i = len(loopCounter0) - 2, separately to avoid an E12 Square
		// (Square(res) = 1² = 1)
		// j = 0
		// k = 0, separately to avoid MulBy034 (res × ℓ)
		// (assign line to res)

		// pProj1[0] ← 2pProj1[0] and l0 the tangent ℓ passing 2pProj1[0]
		pProj1[0].doubleStep(&l0)
		// line evaluation at Q[0] (assign)
		result.B1.A0.Mul(&l0.r1, &q[0].X)
		result.B0.A0.Mul(&l0.r0, &q[0].Y)
		result.B1.A1.Set(&l0.r2)
	}

	// k = 1
	if n >= 2 {
		// pProj1[1] ← 2pProj1[1] and l0 the tangent ℓ passing 2pProj1[1]
		pProj1[1].doubleStep(&l0)
		// line evaluation at Q[0]
		l0.r1.Mul(&l0.r1, &q[1].X)
		l0.r0.Mul(&l0.r0, &q[1].Y)
		// ℓ × res
		prodLines = fptower.Mul034By034(&l0.r0, &l0.r1, &l0.r2, &result.B0.A0, &result.B1.A0, &result.B1.A1)
		result.B0.A0 = prodLines[0]
		result.B0.A1 = prodLines[1]
		result.B0.A2 = prodLines[2]
		result.B1.A0 = prodLines[3]
		result.B1.A1 = prodLines[4]
	}

	// k >= 2
	for k := 2; k < n; k++ {
		// pProj1[1] ← 2pProj1[1] and l0 the tangent ℓ passing 2pProj1[1]
		pProj1[k].doubleStep(&l0)
		// line evaluation at Q[k]
		l0.r1.Mul(&l0.r1, &q[k].X)
		l0.r0.Mul(&l0.r0, &q[k].Y)
		// ℓ × res
		result.MulBy034(&l0.r0, &l0.r1, &l0.r2)
	}

	var tmp G1Affine
	for i := len(loopCounter0) - 3; i >= 1; i-- {
		// (∏ᵢfᵢ)²
		// mutualize the square among n Miller loops
		result.Square(&result)

		j = loopCounter1[i]*3 + loopCounter0[i]

		for k := 0; k < n; k++ {
			// pProj1[1] ← 2pProj1[1] and l0 the tangent ℓ passing 2pProj1[1]
			pProj1[k].doubleStep(&l0)
			// line evaluation at Q[k]
			l0.r1.Mul(&l0.r1, &q[k].X)
			l0.r0.Mul(&l0.r0, &q[k].Y)

			switch j {
			// cases -4, -2, 2, 4 do not occur, given the static loopCounters
			case -3:
				tmp.Neg(&p1[k])
				// pProj1[k] ← pProj1[k]-p1[k] and
				// l the line ℓ passing pProj1[k] and -p1[k]
				pProj1[k].addMixedStep(&l, &tmp)
				// line evaluation at Q[k]
				l.r1.Mul(&l.r1, &q[k].X)
				l.r0.Mul(&l.r0, &q[k].Y)
				// ℓ × ℓ
				prodLines = fptower.Mul034By034(&l.r0, &l.r1, &l.r2, &l0.r0, &l0.r1, &l0.r2)
				// (ℓ × ℓ) × res
				result.MulBy01234(&prodLines)
			case -1:
				tmp.Neg(&p0[k])
				// pProj1[k] ← pProj1[k]-p0[k] and
				// l the line ℓ passing pProj1[k] and -p0[k]
				pProj1[k].addMixedStep(&l, &tmp)
				// line evaluation at Q[k]
				l.r1.Mul(&l.r1, &q[k].X)
				l.r0.Mul(&l.r0, &q[k].Y)
				// ℓ × ℓ
				prodLines = fptower.Mul034By034(&l.r0, &l.r1, &l.r2, &l0.r0, &l0.r1, &l0.r2)
				// (ℓ × ℓ) × res
				result.MulBy01234(&prodLines)
			case 0:
				// ℓ × res
				result.MulBy034(&l0.r0, &l0.r1, &l0.r2)
			case 1:
				// pProj1[k] ← pProj1[k]+p0[k] and
				// l the line ℓ passing pProj1[k] and p0[k]
				pProj1[k].addMixedStep(&l, &p0[k])
				// line evaluation at Q[k]
				l.r1.Mul(&l.r1, &q[k].X)
				l.r0.Mul(&l.r0, &q[k].Y)
				// ℓ × ℓ
				prodLines = fptower.Mul034By034(&l.r0, &l.r1, &l.r2, &l0.r0, &l0.r1, &l0.r2)
				// (ℓ × ℓ) × res
				result.MulBy01234(&prodLines)
			case 3:
				// pProj1[k] ← pProj1[k]+p1[k] and
				// l the line ℓ passing pProj1[k] and p1[k]
				pProj1[k].addMixedStep(&l, &p1[k])
				// line evaluation at Q[k]
				l.r1.Mul(&l.r1, &q[k].X)
				l.r0.Mul(&l.r0, &q[k].Y)
				// ℓ × ℓ
				prodLines = fptower.Mul034By034(&l.r0, &l.r1, &l.r2, &l0.r0, &l0.r1, &l0.r2)
				// (ℓ × ℓ) × res
				result.MulBy01234(&prodLines)
			default:
				return GT{}, errors.New("invalid loopCounter")
			}
		}
	}

	// i = 0, separately to avoid a point addition
	// j = -3
	result.Square(&result)
	for k := 0; k < n; k++ {
		// pProj1[k] ← 2pProj1[k] and l0 the tangent ℓ passing 2pProj1[k]
		pProj1[k].doubleStep(&l0)
		// line evaluation at Q[k]
		l0.r1.Mul(&l0.r1, &q[k].X)
		l0.r0.Mul(&l0.r0, &q[k].Y)

		tmp.Neg(&p1[k])
		// l the line passing pProj1[k] and -p1
		pProj1[k].lineCompute(&l, &tmp)
		// line evaluation at Q[k]
		l.r1.Mul(&l.r1, &q[k].X)
		l.r0.Mul(&l.r0, &q[k].Y)
		// ℓ × ℓ
		prodLines = fptower.Mul034By034(&l.r0, &l.r1, &l.r2, &l0.r0, &l0.r1, &l0.r2)
		// (ℓ × ℓ) × res
		result.MulBy01234(&prodLines)
	}

	return result, nil
}

// doubleStep doubles a point in Homogenous projective coordinates, and evaluates the line in Miller loop
// https://eprint.iacr.org/2013/722.pdf (Section 4.3)
func (p *g1Proj) doubleStep(evaluations *lineEvaluation) {

	// get some Element from our pool
	var t1, A, B, C, D, E, EE, F, G, H, I, J, K fp.Element
	A.Mul(&p.x, &p.y)
	A.Halve()
	B.Square(&p.y)
	C.Square(&p.z)
	D.Double(&C).
		Add(&D, &C)
	// E.Mul(&D, &bCurveCoeff)
	E.Set(&D)
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

// addMixedStep point addition in Mixed Homogenous projective and Affine coordinates
// https://eprint.iacr.org/2013/722.pdf (Section 4.3)
func (p *g1Proj) addMixedStep(evaluations *lineEvaluation, a *G1Affine) {

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

// lineCompute computes the line through p in Homogenous projective coordinates
// and a in affine coordinates. It does not compute the resulting point p+a.
func (p *g1Proj) lineCompute(evaluations *lineEvaluation, a *G1Affine) {

	// get some Element from our pool
	var Y2Z1, X2Z1, O, L, t2, J fp.Element
	Y2Z1.Mul(&a.Y, &p.z)
	O.Sub(&p.y, &Y2Z1)
	X2Z1.Mul(&a.X, &p.z)
	L.Sub(&p.x, &X2Z1)
	t2.Mul(&L, &a.Y)
	J.Mul(&a.X, &O).
		Sub(&J, &t2)

	// Line evaluation
	evaluations.r0.Set(&L)
	evaluations.r1.Neg(&O)
	evaluations.r2.Set(&J)
}
