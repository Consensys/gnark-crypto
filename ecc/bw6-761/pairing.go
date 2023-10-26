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

type LineEvaluation struct {
	R0 fp.Element
	R1 fp.Element
	R2 fp.Element
}

func (l *LineEvaluation) Set(line *LineEvaluation) *LineEvaluation {
	l.R0, l.R1, l.R2 = line.R0, line.R1, line.R2
	return l
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

// PairFixedQ calculates the reduced pairing for a set of points
// ∏ᵢ e(Pᵢ, Qᵢ) where Q are fixed points in G2.
//
// This function doesn't check that the inputs are in the correct subgroup. See IsInSubGroup.
func PairFixedQ(P []G1Affine, lines [][2][189]LineEvaluation) (GT, error) {
	f, err := MillerLoopFixedQ(P, lines)
	if err != nil {
		return GT{}, err
	}
	return FinalExponentiation(&f), nil
}

// PairingCheckFixedQ calculates the reduced pairing for a set of points and returns True if the result is One
// ∏ᵢ e(Pᵢ, Qᵢ) =? 1 where Q are fixed points in G2.
//
// This function doesn't check that the inputs are in the correct subgroup. See IsInSubGroup.
func PairingCheckFixedQ(P []G1Affine, lines [][2][189]LineEvaluation) (bool, error) {
	f, err := PairFixedQ(P, lines)
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
	j.Expc1(&h)
	j.Mul(&j, &e)
	k.CyclotomicSquare(&j)
	k.Mul(&k, &j)
	k.Mul(&k, &b)
	t.Expc2(&i)
	k.Mul(&k, &t)
	result.Mul(&a, &k)

	return result
}

// MillerLoop computes the multi-Miller loop
// ∏ᵢ MillerLoop(Pᵢ, Qᵢ) =
// ∏ᵢ { fᵢ_{x₀+1+λ(x₀³-x₀²-x₀),Qᵢ}(Pᵢ) }
//
// Alg.2 in https://eprint.iacr.org/2021/1359.pdf
// Eq. (6') in https://hackmd.io/@gnark/BW6-761-changes
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
	q1Neg := make([]G2Affine, n)
	q0Neg := make([]G2Affine, n)
	for k := 0; k < n; k++ {
		q1[k].Y.Neg(&q0[k].Y)
		q0Neg[k].X.Set(&q0[k].X)
		q0Neg[k].Y.Set(&q1[k].Y)
		q1[k].X.Mul(&q0[k].X, &thirdRootOneG1)
		qProj1[k].FromAffine(&q1[k])
		q1Neg[k].Neg(&q1[k])
	}

	// f_{a0+λ*a1,Q}(P)
	var result GT
	result.SetOne()
	var l, l0 LineEvaluation
	var prodLines [5]fp.Element

	var j int8

	if n >= 1 {
		// i = 188, separately to avoid an E12 Square
		// (Square(res) = 1² = 1)
		// j = 0
		// k = 0, separately to avoid MulBy014 (res × ℓ)
		// (assign line to res)
		// qProj1[0] ← 2qProj1[0] and l0 the tangent ℓ passing 2qProj1[0]
		qProj1[0].doubleStep(&l0)
		// line evaluation at Q[0] (assign)
		result.B0.A0.Set(&l0.R0)
		result.B0.A1.Mul(&l0.R1, &p[0].X)
		result.B1.A1.Mul(&l0.R2, &p[0].Y)
	}

	// k = 1
	if n >= 2 {
		// qProj1[1] ← 2qProj1[1] and l0 the tangent ℓ passing 2qProj1[1]
		qProj1[1].doubleStep(&l0)
		// line evaluation at Q[1]
		l0.R1.Mul(&l0.R1, &p[1].X)
		l0.R2.Mul(&l0.R2, &p[1].Y)
		prodLines = fptower.Mul014By014(&l0.R0, &l0.R1, &l0.R2, &result.B0.A0, &result.B0.A1, &result.B1.A1)
		result.B0.A0 = prodLines[0]
		result.B0.A1 = prodLines[1]
		result.B0.A2 = prodLines[2]
		result.B1.A1 = prodLines[3]
		result.B1.A2 = prodLines[4]
	}

	// k >= 2
	for k := 2; k < n; k++ {
		// qProj1[k] ← 2qProj1[k] and l0 the tangent ℓ passing 2qProj1[k]
		qProj1[k].doubleStep(&l0)
		// line evaluation at Q[k]
		l0.R1.Mul(&l0.R1, &p[k].X)
		l0.R2.Mul(&l0.R2, &p[k].Y)
		// ℓ × res
		result.MulBy014(&l0.R0, &l0.R1, &l0.R2)
	}

	for i := 187; i >= 1; i-- {
		result.Square(&result)

		j = loopCounter1[i]*3 + loopCounter0[i]

		for k := 0; k < n; k++ {
			qProj1[k].doubleStep(&l0)
			l0.R1.Mul(&l0.R1, &p[k].X)
			l0.R2.Mul(&l0.R2, &p[k].Y)

			switch j {
			// cases -4, -2, 2, 4 do not occur, given the static loopCounters
			case -3:
				qProj1[k].addMixedStep(&l, &q1Neg[k])
				l.R1.Mul(&l.R1, &p[k].X)
				l.R2.Mul(&l.R2, &p[k].Y)
				prodLines = fptower.Mul014By014(&l0.R0, &l0.R1, &l0.R2, &l.R0, &l.R1, &l.R2)
				result.MulBy01245(&prodLines)
			case -1:
				qProj1[k].addMixedStep(&l, &q0Neg[k])
				l.R1.Mul(&l.R1, &p[k].X)
				l.R2.Mul(&l.R2, &p[k].Y)
				prodLines = fptower.Mul014By014(&l0.R0, &l0.R1, &l0.R2, &l.R0, &l.R1, &l.R2)
				result.MulBy01245(&prodLines)
			case 0:
				result.MulBy014(&l0.R0, &l0.R1, &l0.R2)
			case 1:
				qProj1[k].addMixedStep(&l, &q0[k])
				l.R1.Mul(&l.R1, &p[k].X)
				l.R2.Mul(&l.R2, &p[k].Y)
				prodLines = fptower.Mul014By014(&l0.R0, &l0.R1, &l0.R2, &l.R0, &l.R1, &l.R2)
				result.MulBy01245(&prodLines)
			case 3:
				qProj1[k].addMixedStep(&l, &q1[k])
				l.R1.Mul(&l.R1, &p[k].X)
				l.R2.Mul(&l.R2, &p[k].Y)
				prodLines = fptower.Mul014By014(&l0.R0, &l0.R1, &l0.R2, &l.R0, &l.R1, &l.R2)
				result.MulBy01245(&prodLines)
			default:
				return GT{}, errors.New("invalid loopCounter")
			}
		}
	}

	// i = 0, separately to avoid a point addition
	// j = -3
	result.Square(&result)
	for k := 0; k < n; k++ {
		qProj1[k].doubleStep(&l0)
		l0.R1.Mul(&l0.R1, &p[k].X)
		l0.R2.Mul(&l0.R2, &p[k].Y)
		qProj1[k].lineCompute(&l, &q1Neg[k])
		l.R1.Mul(&l.R1, &p[k].X)
		l.R2.Mul(&l.R2, &p[k].Y)
		prodLines = fptower.Mul014By014(&l0.R0, &l0.R1, &l0.R2, &l.R0, &l.R1, &l.R2)
		result.MulBy01245(&prodLines)
	}

	return result, nil

}

func PrecomputeLines(Q G2Affine) (PrecomputedLines [2][189]LineEvaluation) {

	// precomputations
	var imQ, imQneg, negQ G2Affine
	var accQ g2Proj
	imQ.Y.Neg(&Q.Y)
	negQ.X.Set(&Q.X)
	negQ.Y.Set(&imQ.Y)
	imQ.X.Mul(&Q.X, &thirdRootOneG1)
	accQ.FromAffine(&imQ)
	imQneg.Neg(&imQ)

	var l LineEvaluation
	for i := 188; i >= 0; i-- {

		accQ.doubleStep(&l)
		PrecomputedLines[0][i].Set(&l)

		switch loopCounter1[i]*3 + loopCounter0[i] {
		// cases -4, -2, 2, 4 do not occur, given the static loopCounters
		case -3:
			accQ.addMixedStep(&l, &imQneg)
			PrecomputedLines[1][i].Set(&l)
		case -1:
			accQ.addMixedStep(&l, &negQ)
			PrecomputedLines[1][i].Set(&l)
		case 0:
			continue
		case 1:
			accQ.addMixedStep(&l, &Q)
			PrecomputedLines[1][i].Set(&l)
		case 3:
			accQ.addMixedStep(&l, &imQ)
			PrecomputedLines[1][i].Set(&l)
		default:
			return [2][189]LineEvaluation{}
		}
	}

	return PrecomputedLines
}

// MillerLoopFixedQ computes the multi-Miller loop as in MillerLoop
// but Qᵢ are fixed points in G2 known in advance.
func MillerLoopFixedQ(P []G1Affine, lines [][2][189]LineEvaluation) (GT, error) {
	// check input size match
	n := len(P)
	if n == 0 || n != len(lines) {
		return GT{}, errors.New("invalid inputs sizes")
	}

	// filter infinity points
	p := make([]G1Affine, 0, n)

	for k := 0; k < n; k++ {
		if P[k].IsInfinity() {
			continue
		}
		p = append(p, P[k])
	}

	n = len(p)

	// f_{a0+λ*a1,Q}(P)
	var result GT
	result.SetOne()
	var prodLines [5]fp.Element

	for i := 188; i >= 0; i-- {
		result.Square(&result)

		j := loopCounter1[i]*3 + loopCounter0[i]
		for k := 0; k < n; k++ {
			lines[k][0][i].R1.
				Mul(
					&lines[k][0][i].R1,
					&p[k].X,
				)
			lines[k][0][i].R2.
				Mul(&lines[k][0][i].R2,
					&p[k].Y,
				)
			if j == 0 {
				result.MulBy014(
					&lines[k][0][i].R0,
					&lines[k][0][i].R1,
					&lines[k][0][i].R2,
				)

			} else {
				lines[k][1][i].R1.
					Mul(
						&lines[k][1][i].R1,
						&p[k].X,
					)
				lines[k][1][i].R2.
					Mul(
						&lines[k][1][i].R2,
						&p[k].Y,
					)
				prodLines = fptower.Mul014By014(
					&lines[k][0][i].R0, &lines[k][0][i].R1, &lines[k][0][i].R2,
					&lines[k][1][i].R0, &lines[k][1][i].R1, &lines[k][1][i].R2,
				)
				result.MulBy01245(&prodLines)
			}
		}
	}

	return result, nil

}

// doubleStep doubles a point in Homogenous projective coordinates, and evaluates the line in Miller loop
// https://eprint.iacr.org/2013/722.pdf (Section 4.3)
func (p *g2Proj) doubleStep(evaluations *LineEvaluation) {

	// get some Element from our pool
	var t1, A, B, C, D, E, EE, F, G, H, I, J, K fp.Element
	A.Mul(&p.x, &p.y)
	A.Halve()
	B.Square(&p.y)
	C.Square(&p.z)
	D.Double(&C).
		Add(&D, &C)
	E.Mul(&D, &bTwistCurveCoeff)
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
	evaluations.R0.Set(&I)
	evaluations.R1.Double(&J).
		Add(&evaluations.R1, &J)
	evaluations.R2.Neg(&H)
}

// addMixedStep point addition in Mixed Homogenous projective and Affine coordinates
// https://eprint.iacr.org/2013/722.pdf (Section 4.3)
func (p *g2Proj) addMixedStep(evaluations *LineEvaluation, a *G2Affine) {

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
	evaluations.R0.Set(&J)
	evaluations.R1.Neg(&O)
	evaluations.R2.Set(&L)
}

// lineCompute computes the line through p in Homogenous projective coordinates
// and a in affine coordinates. It does not compute the resulting point p+a.
func (p *g2Proj) lineCompute(evaluations *LineEvaluation, a *G2Affine) {

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
	evaluations.R0.Set(&J)
	evaluations.R1.Neg(&O)
	evaluations.R2.Set(&L)
}
