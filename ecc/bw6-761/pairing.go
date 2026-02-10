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
	var l, l0 lineEvaluation
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
		result.B0.A0.Set(&l0.r0)
		result.B0.A1.Mul(&l0.r1, &p[0].X)
		result.B1.A1.Mul(&l0.r2, &p[0].Y)
	}

	// k = 1
	if n >= 2 {
		// qProj1[1] ← 2qProj1[1] and l0 the tangent ℓ passing 2qProj1[1]
		qProj1[1].doubleStep(&l0)
		// line evaluation at Q[1]
		l0.r1.Mul(&l0.r1, &p[1].X)
		l0.r2.Mul(&l0.r2, &p[1].Y)
		prodLines = fptower.Mul014By014(&l0.r0, &l0.r1, &l0.r2, &result.B0.A0, &result.B0.A1, &result.B1.A1)
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
		l0.r1.Mul(&l0.r1, &p[k].X)
		l0.r2.Mul(&l0.r2, &p[k].Y)
		// ℓ × res
		result.MulBy014(&l0.r0, &l0.r1, &l0.r2)
	}

	for i := 187; i >= 1; i-- {
		result.Square(&result)

		j = LoopCounter1[i]*3 + LoopCounter[i]

		for k := 0; k < n; k++ {
			qProj1[k].doubleStep(&l0)
			l0.r1.Mul(&l0.r1, &p[k].X)
			l0.r2.Mul(&l0.r2, &p[k].Y)

			switch j {
			// cases -4, -2, 2, 4 do not occur, given the static LoopCounters
			case -3:
				qProj1[k].addMixedStep(&l, &q1Neg[k])
				l.r1.Mul(&l.r1, &p[k].X)
				l.r2.Mul(&l.r2, &p[k].Y)
				prodLines = fptower.Mul014By014(&l0.r0, &l0.r1, &l0.r2, &l.r0, &l.r1, &l.r2)
				result.MulBy01245(&prodLines)
			case -1:
				qProj1[k].addMixedStep(&l, &q0Neg[k])
				l.r1.Mul(&l.r1, &p[k].X)
				l.r2.Mul(&l.r2, &p[k].Y)
				prodLines = fptower.Mul014By014(&l0.r0, &l0.r1, &l0.r2, &l.r0, &l.r1, &l.r2)
				result.MulBy01245(&prodLines)
			case 0:
				result.MulBy014(&l0.r0, &l0.r1, &l0.r2)
			case 1:
				qProj1[k].addMixedStep(&l, &q0[k])
				l.r1.Mul(&l.r1, &p[k].X)
				l.r2.Mul(&l.r2, &p[k].Y)
				prodLines = fptower.Mul014By014(&l0.r0, &l0.r1, &l0.r2, &l.r0, &l.r1, &l.r2)
				result.MulBy01245(&prodLines)
			case 3:
				qProj1[k].addMixedStep(&l, &q1[k])
				l.r1.Mul(&l.r1, &p[k].X)
				l.r2.Mul(&l.r2, &p[k].Y)
				prodLines = fptower.Mul014By014(&l0.r0, &l0.r1, &l0.r2, &l.r0, &l.r1, &l.r2)
				result.MulBy01245(&prodLines)
			default:
				return GT{}, errors.New("invalid LoopCounter")
			}
		}
	}

	// i = 0, separately to avoid a point addition
	// j = -3
	result.Square(&result)
	for k := 0; k < n; k++ {
		qProj1[k].doubleStep(&l0)
		l0.r1.Mul(&l0.r1, &p[k].X)
		l0.r2.Mul(&l0.r2, &p[k].Y)
		qProj1[k].lineCompute(&l, &q1Neg[k])
		l.r1.Mul(&l.r1, &p[k].X)
		l.r2.Mul(&l.r2, &p[k].Y)
		prodLines = fptower.Mul014By014(&l0.r0, &l0.r1, &l0.r2, &l.r0, &l.r1, &l.r2)
		result.MulBy01245(&prodLines)
	}

	return result, nil

}

// doubleStep doubles a point in Homogenous projective coordinates, and evaluates the line in Miller loop
// https://eprint.iacr.org/2013/722.pdf (Section 4.3)
func (p *g2Proj) doubleStep(evaluations *lineEvaluation) {

	// get some Element from our pool
	var t1, A, B, C, D, E, EE, F, G, H, I, J, K fp.Element
	A.Mul(&p.x, &p.y)
	A.Halve()
	B.Square(&p.y)
	C.Square(&p.z)
	D.Double(&C).
		Add(&D, &C)
	E.Double(&D).Double(&E)
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
	evaluations.r0.Set(&I)
	evaluations.r1.Double(&J).
		Add(&evaluations.r1, &J)
	evaluations.r2.Neg(&H)
}

// addMixedStep point addition in Mixed Homogenous projective and Affine coordinates
// https://eprint.iacr.org/2013/722.pdf (Section 4.3)
func (p *g2Proj) addMixedStep(evaluations *lineEvaluation, a *G2Affine) {

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

// lineCompute computes the line through p in Homogenous projective coordinates
// and a in affine coordinates. It does not compute the resulting point p+a.
func (p *g2Proj) lineCompute(evaluations *lineEvaluation, a *G2Affine) {

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
	evaluations.r0.Set(&J)
	evaluations.r1.Neg(&O)
	evaluations.r2.Set(&L)
}

// ----------------------
// Fixed-argument pairing
// ----------------------

type LineEvaluationAff struct {
	R0 fp.Element
	R1 fp.Element
}

// PairFixedQ calculates the reduced pairing for a set of points
// ∏ᵢ e(Pᵢ, Qᵢ) where Q are fixed points in G2.
//
// This function doesn't check that the inputs are in the correct subgroup. See IsInSubGroup.
func PairFixedQ(P []G1Affine, lines [][2][len(LoopCounter) - 1]LineEvaluationAff) (GT, error) {
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
func PairingCheckFixedQ(P []G1Affine, lines [][2][len(LoopCounter) - 1]LineEvaluationAff) (bool, error) {
	f, err := PairFixedQ(P, lines)
	if err != nil {
		return false, err
	}
	var one GT
	one.SetOne()
	return f.Equal(&one), nil
}

// PrecomputeLines precomputes the lines for the fixed-argument Miller loop
func PrecomputeLines(Q G2Affine) (PrecomputedLines [2][len(LoopCounter) - 1]LineEvaluationAff) {

	// precomputations
	var accQ, imQ, imQneg, negQ G2Affine
	imQ.Y.Neg(&Q.Y)
	imQ.X.Mul(&Q.X, &thirdRootOneG1)
	negQ.X.Set(&Q.X)
	negQ.Y.Set(&imQ.Y)
	accQ.Set(&imQ)
	imQneg.X.Set(&imQ.X)
	imQneg.Y.Set(&Q.Y)

	// The loop processes i from 188 down to 1, with i=0 handled specially.
	// j = LoopCounter1[i]*3 + LoopCounter[i]
	// j values: 0 at i=188,187; 3 at i=186; 0 at i=185-181; -3 at i=180; etc.

	// i = 188: j = 0
	accQ.doubleStep(&PrecomputedLines[0][188])
	// i = 187: j = 0
	accQ.doubleStep(&PrecomputedLines[0][187])
	// i = 186: j = 3
	accQ.doubleAndAddStep(&PrecomputedLines[0][186], &PrecomputedLines[1][186], &imQ)
	// i = 185 -> 180: 5 consecutive zeros followed by add at i=180
	accQ.manyDoublesAndAdd(5, PrecomputedLines[0][181:186], &PrecomputedLines[0][180], &PrecomputedLines[1][180], &imQneg)
	// i = 179: j = 0
	accQ.doubleStep(&PrecomputedLines[0][179])
	// i = 178: j = -3
	accQ.doubleAndAddStep(&PrecomputedLines[0][178], &PrecomputedLines[1][178], &imQneg)
	// i = 177: j = 0
	accQ.doubleStep(&PrecomputedLines[0][177])
	// i = 176: j = 3
	accQ.doubleAndAddStep(&PrecomputedLines[0][176], &PrecomputedLines[1][176], &imQ)
	// i = 175 -> 172: 3 consecutive zeros followed by add at i=172
	accQ.manyDoublesAndAdd(3, PrecomputedLines[0][173:176], &PrecomputedLines[0][172], &PrecomputedLines[1][172], &imQ)
	// i = 171: j = 0
	accQ.doubleStep(&PrecomputedLines[0][171])
	// i = 170: j = 3
	accQ.doubleAndAddStep(&PrecomputedLines[0][170], &PrecomputedLines[1][170], &imQ)
	// i = 169: j = 0
	accQ.doubleStep(&PrecomputedLines[0][169])
	// i = 168: j = -3
	accQ.doubleAndAddStep(&PrecomputedLines[0][168], &PrecomputedLines[1][168], &imQneg)
	// i = 167: j = 0
	accQ.doubleStep(&PrecomputedLines[0][167])
	// i = 166: j = 3
	accQ.doubleAndAddStep(&PrecomputedLines[0][166], &PrecomputedLines[1][166], &imQ)
	// i = 165: j = 0
	accQ.doubleStep(&PrecomputedLines[0][165])
	// i = 164: j = 0
	accQ.doubleStep(&PrecomputedLines[0][164])
	// i = 163: j = 3
	accQ.doubleAndAddStep(&PrecomputedLines[0][163], &PrecomputedLines[1][163], &imQ)
	// i = 162 -> 159: 3 consecutive zeros followed by add at i=159
	accQ.manyDoublesAndAdd(3, PrecomputedLines[0][160:163], &PrecomputedLines[0][159], &PrecomputedLines[1][159], &imQneg)
	// i = 158: j = 0
	accQ.doubleStep(&PrecomputedLines[0][158])
	// i = 157: j = 3
	accQ.doubleAndAddStep(&PrecomputedLines[0][157], &PrecomputedLines[1][157], &imQ)
	// i = 156: j = 0
	accQ.doubleStep(&PrecomputedLines[0][156])
	// i = 155: j = -3
	accQ.doubleAndAddStep(&PrecomputedLines[0][155], &PrecomputedLines[1][155], &imQneg)
	// i = 154: j = 0
	accQ.doubleStep(&PrecomputedLines[0][154])
	// i = 153: j = 0
	accQ.doubleStep(&PrecomputedLines[0][153])
	// i = 152: j = -3
	accQ.doubleAndAddStep(&PrecomputedLines[0][152], &PrecomputedLines[1][152], &imQneg)
	// i = 151 -> 148: 3 consecutive zeros followed by add at i=148
	accQ.manyDoublesAndAdd(3, PrecomputedLines[0][149:152], &PrecomputedLines[0][148], &PrecomputedLines[1][148], &imQ)
	// i = 147: j = 0
	accQ.doubleStep(&PrecomputedLines[0][147])
	// i = 146: j = 0
	accQ.doubleStep(&PrecomputedLines[0][146])
	// i = 145: j = -3
	accQ.doubleAndAddStep(&PrecomputedLines[0][145], &PrecomputedLines[1][145], &imQneg)
	// i = 144 -> 140: 4 consecutive zeros followed by add at i=140
	accQ.manyDoublesAndAdd(4, PrecomputedLines[0][141:145], &PrecomputedLines[0][140], &PrecomputedLines[1][140], &imQneg)
	// i = 139: j = 0
	accQ.doubleStep(&PrecomputedLines[0][139])
	// i = 138: j = -3
	accQ.doubleAndAddStep(&PrecomputedLines[0][138], &PrecomputedLines[1][138], &imQneg)
	// i = 137 -> 127: 10 consecutive zeros followed by add at i=127
	accQ.manyDoublesAndAdd(10, PrecomputedLines[0][128:138], &PrecomputedLines[0][127], &PrecomputedLines[1][127], &imQ)
	// i = 126 -> 123: 3 consecutive zeros followed by add at i=123
	accQ.manyDoublesAndAdd(3, PrecomputedLines[0][124:127], &PrecomputedLines[0][123], &PrecomputedLines[1][123], &imQ)
	// i = 122: j = 0
	accQ.doubleStep(&PrecomputedLines[0][122])
	// i = 121: j = 3
	accQ.doubleAndAddStep(&PrecomputedLines[0][121], &PrecomputedLines[1][121], &imQ)
	// i = 120: j = 0
	accQ.doubleStep(&PrecomputedLines[0][120])
	// i = 119: j = 0
	accQ.doubleStep(&PrecomputedLines[0][119])
	// i = 118: j = 3
	accQ.doubleAndAddStep(&PrecomputedLines[0][118], &PrecomputedLines[1][118], &imQ)
	// i = 117 -> 114: 3 consecutive zeros followed by add at i=114
	accQ.manyDoublesAndAdd(3, PrecomputedLines[0][115:118], &PrecomputedLines[0][114], &PrecomputedLines[1][114], &imQ)
	// i = 113 -> 110: 3 consecutive zeros followed by add at i=110
	accQ.manyDoublesAndAdd(3, PrecomputedLines[0][111:114], &PrecomputedLines[0][110], &PrecomputedLines[1][110], &imQ)
	// i = 109: j = 0
	accQ.doubleStep(&PrecomputedLines[0][109])
	// i = 108: j = -3
	accQ.doubleAndAddStep(&PrecomputedLines[0][108], &PrecomputedLines[1][108], &imQneg)
	// i = 107 -> 103: 4 consecutive zeros followed by add at i=103
	accQ.manyDoublesAndAdd(4, PrecomputedLines[0][104:108], &PrecomputedLines[0][103], &PrecomputedLines[1][103], &imQneg)
	// i = 102: j = 0
	accQ.doubleStep(&PrecomputedLines[0][102])
	// i = 101: j = 3
	accQ.doubleAndAddStep(&PrecomputedLines[0][101], &PrecomputedLines[1][101], &imQ)
	// i = 100: j = 0
	accQ.doubleStep(&PrecomputedLines[0][100])
	// i = 99: j = -3
	accQ.doubleAndAddStep(&PrecomputedLines[0][99], &PrecomputedLines[1][99], &imQneg)
	// i = 98: j = 0
	accQ.doubleStep(&PrecomputedLines[0][98])
	// i = 97: j = 0
	accQ.doubleStep(&PrecomputedLines[0][97])
	// i = 96: j = 3
	accQ.doubleAndAddStep(&PrecomputedLines[0][96], &PrecomputedLines[1][96], &imQ)
	// i = 95: j = 0
	accQ.doubleStep(&PrecomputedLines[0][95])
	// i = 94: j = 0
	accQ.doubleStep(&PrecomputedLines[0][94])
	// i = 93: j = 3
	accQ.doubleAndAddStep(&PrecomputedLines[0][93], &PrecomputedLines[1][93], &imQ)
	// i = 92 -> 63: 29 consecutive zeros followed by add at i=63
	accQ.manyDoublesAndAdd(29, PrecomputedLines[0][64:93], &PrecomputedLines[0][63], &PrecomputedLines[1][63], &Q)
	// i = 62 -> 58: 4 consecutive zeros followed by add at i=58
	accQ.manyDoublesAndAdd(4, PrecomputedLines[0][59:63], &PrecomputedLines[0][58], &PrecomputedLines[1][58], &Q)
	// i = 57: j = 0
	accQ.doubleStep(&PrecomputedLines[0][57])
	// i = 56: j = 1
	accQ.doubleAndAddStep(&PrecomputedLines[0][56], &PrecomputedLines[1][56], &Q)
	// i = 55 -> 51: 4 consecutive zeros followed by add at i=51
	accQ.manyDoublesAndAdd(4, PrecomputedLines[0][52:56], &PrecomputedLines[0][51], &PrecomputedLines[1][51], &Q)
	// i = 50: j = 0
	accQ.doubleStep(&PrecomputedLines[0][50])
	// i = 49: j = 0
	accQ.doubleStep(&PrecomputedLines[0][49])
	// i = 48: j = 1
	accQ.doubleAndAddStep(&PrecomputedLines[0][48], &PrecomputedLines[1][48], &Q)
	// i = 47: j = 0
	accQ.doubleStep(&PrecomputedLines[0][47])
	// i = 46: j = -1
	accQ.doubleAndAddStep(&PrecomputedLines[0][46], &PrecomputedLines[1][46], &negQ)
	// i = 45 -> 1: 44 consecutive zeros followed by add at i=1
	accQ.manyDoublesAndAdd(44, PrecomputedLines[0][2:46], &PrecomputedLines[0][1], &PrecomputedLines[1][1], &Q)
	// i = 0: j = -3, use tangentCompute (not doubleStep)
	accQ.tangentCompute(&PrecomputedLines[0][0])

	return PrecomputedLines
}

// MillerLoopFixedQ computes the multi-Miller loop as in MillerLoop
// but Qᵢ are fixed points in G2 known in advance.
func MillerLoopFixedQ(P []G1Affine, lines [][2][len(LoopCounter) - 1]LineEvaluationAff) (GT, error) {
	// check input size match
	n := len(P)
	if n == 0 || n != len(lines) {
		return GT{}, errors.New("invalid inputs sizes")
	}

	// no need to filter infinity points:
	// 		1. if Pᵢ=(0,0) then -x/y=1/y=0 by gnark-crypto convention and so
	// 		lines R0 and R1 are 0. It happens that result will stay, through
	// 		the Miller loop, in 𝔽p⁶ because MulBy01(0,0,1),
	// 		Mul01By01(0,0,1,0,0,1) and MulBy01245 set result.C0 to 0. At the
	// 		end result will be in a proper subgroup of Fp¹² so it be reduced to
	// 		1 in FinalExponentiation.
	//
	//      and/or
	//
	// 		2. if Qᵢ=(0,0) then PrecomputeLines(Qᵢ) will return lines R0 and R1
	// 		that are 0 because of gnark-convention (*/0==0) in doubleStep and
	// 		addStep. Similarly to Pᵢ=(0,0) it happens that result be 1
	// 		after the FinalExponentiation.

	// precomputations
	yInv := make([]fp.Element, n)
	xNegOverY := make([]fp.Element, n)
	for k := 0; k < n; k++ {
		yInv[k].Set(&P[k].Y)
	}
	yInv = fp.BatchInvert(yInv)
	for k := 0; k < n; k++ {
		xNegOverY[k].Mul(&P[k].X, &yInv[k]).
			Neg(&xNegOverY[k])
	}

	// f_{a0+λ*a1,Q}(P)
	var result GT
	result.SetOne()
	var prodLines [5]fp.Element

	for i := len(LoopCounter) - 2; i >= 0; i-- {
		result.Square(&result)

		j := LoopCounter1[i]*3 + LoopCounter[i]
		for k := 0; k < n; k++ {
			lines[k][0][i].R1.
				Mul(
					&lines[k][0][i].R1,
					&yInv[k],
				)
			lines[k][0][i].R0.
				Mul(&lines[k][0][i].R0,
					&xNegOverY[k],
				)
			if j == 0 {
				result.MulBy01(
					&lines[k][0][i].R1,
					&lines[k][0][i].R0,
				)

			} else {
				lines[k][1][i].R1.
					Mul(
						&lines[k][1][i].R1,
						&yInv[k],
					)
				lines[k][1][i].R0.
					Mul(
						&lines[k][1][i].R0,
						&xNegOverY[k],
					)
				prodLines = fptower.Mul01By01(
					&lines[k][0][i].R1, &lines[k][0][i].R0,
					&lines[k][1][i].R1, &lines[k][1][i].R0,
				)
				result.MulBy01245(&prodLines)
			}
		}
	}

	return result, nil

}

func (p *G2Affine) doubleStep(evaluations *LineEvaluationAff) {

	var n, d, λ, xr, yr fp.Element
	// λ = 3x²/2y
	n.Square(&p.X)
	λ.Double(&n).
		Add(&λ, &n)
	d.Double(&p.Y)
	λ.Div(&λ, &d)

	// xr = λ²-2x
	xr.Square(&λ).
		Sub(&xr, &p.X).
		Sub(&xr, &p.X)

	// yr = λ(x-xr)-y
	yr.Sub(&p.X, &xr).
		Mul(&yr, &λ).
		Sub(&yr, &p.Y)

	evaluations.R0.Set(&λ)
	evaluations.R1.Mul(&λ, &p.X).
		Sub(&evaluations.R1, &p.Y)

	p.X.Set(&xr)
	p.Y.Set(&yr)
}

func (p *G2Affine) addStep(evaluations *LineEvaluationAff, a *G2Affine) {
	var n, d, λ, λλ, xr, yr fp.Element

	// compute λ = (y2-y1)/(x2-x1)
	n.Sub(&a.Y, &p.Y)
	d.Sub(&a.X, &p.X)
	λ.Div(&n, &d)

	// xr = λ²-x1-x2
	λλ.Square(&λ)
	n.Add(&p.X, &a.X)
	xr.Sub(&λλ, &n)

	// yr = λ(x1-xr) - y1
	yr.Sub(&p.X, &xr).
		Mul(&yr, &λ).
		Sub(&yr, &p.Y)

	evaluations.R0.Set(&λ)
	evaluations.R1.Mul(&λ, &p.X).
		Sub(&evaluations.R1, &p.Y)

	p.X.Set(&xr)
	p.Y.Set(&yr)
}

func (p *G2Affine) doubleAndAddStep(evaluations1, evaluations2 *LineEvaluationAff, a *G2Affine) {
	var A, B, A2, B2, X2A2, t, U, AU, invAU, invA, invU, l1, x3, l2, x4, y4 fp.Element

	// The Eisenträger-Lauter-Montgomery formula for 2P+Q (https://eprint.iacr.org/2003/257)
	// computes both slopes λ1 and λ2 using a single field inversion via batch inversion.
	//
	// Given P = (x1, y1) and Q = (x2, y2), let:
	//   A = x1 - x2
	//   B = y1 - y2
	//   U = B² - (2x1 + x2)·A²
	//
	// Then:
	//   λ1 = B/A                    (slope for P + Q)
	//   λ2 = -λ1 - 2y1·A²/U         (slope for P + (P+Q))
	//
	// We compute 1/A and 1/U using Montgomery's batch inversion:
	//   1/A = U/(A·U) and 1/U = A/(A·U) with a single inversion of A·U.

	// Compute A = x1 - x2 and B = y1 - y2
	A.Sub(&p.X, &a.X)
	B.Sub(&p.Y, &a.Y)

	// Compute A² and B²
	A2.Square(&A)
	B2.Square(&B)

	// Compute U = B² - (2x1 + x2)·A²
	t.Double(&p.X).Add(&t, &a.X)
	X2A2.Mul(&t, &A2)
	U.Sub(&B2, &X2A2)

	// Batch inversion: compute 1/A and 1/U with a single inversion
	AU.Mul(&A, &U)
	invAU.Inverse(&AU)
	invA.Mul(&U, &invAU)
	invU.Mul(&A, &invAU)

	// λ1 = B/A = B·(1/A)
	l1.Mul(&B, &invA)

	// x3 = λ1² - x1 - x2
	x3.Square(&l1)
	x3.Sub(&x3, &p.X)
	x3.Sub(&x3, &a.X)

	// line1 evaluation
	evaluations1.R0.Set(&l1)
	evaluations1.R1.Mul(&l1, &p.X)
	evaluations1.R1.Sub(&evaluations1.R1, &p.Y)

	// λ2 = -λ1 - 2y1·A²/U = -λ1 - 2y1·A²·(1/U)
	l2.Double(&p.Y)
	l2.Mul(&l2, &A2)
	l2.Mul(&l2, &invU)
	l2.Add(&l2, &l1)
	l2.Neg(&l2)

	// x4 = λ2² - x1 - x3
	x4.Square(&l2)
	x4.Sub(&x4, &p.X)
	x4.Sub(&x4, &x3)

	// y4 = λ2·(x1 - x4) - y1
	y4.Sub(&p.X, &x4)
	y4.Mul(&l2, &y4)
	y4.Sub(&y4, &p.Y)

	// line2 evaluation
	evaluations2.R0.Set(&l2)
	evaluations2.R1.Mul(&l2, &p.X)
	evaluations2.R1.Sub(&evaluations2.R1, &p.Y)

	p.X.Set(&x4)
	p.Y.Set(&y4)
}

func batchInvertFp(in []fp.Element) []fp.Element {
	n := len(in)
	if n == 0 {
		return nil
	}
	result := make([]fp.Element, n)
	partials := make([]fp.Element, n)
	partials[0].Set(&in[0])
	for i := 1; i < n; i++ {
		partials[i].Mul(&partials[i-1], &in[i])
	}
	var inv fp.Element
	inv.Inverse(&partials[n-1])
	for i := n - 1; i > 0; i-- {
		result[i].Mul(&inv, &partials[i-1])
		inv.Mul(&inv, &in[i])
	}
	result[0].Set(&inv)
	return result
}

func (p *G2Affine) manyDoubleSteps(k int, evaluations []LineEvaluationAff) {
	if k == 0 {
		return
	}

	// Step 1: Compute A[i], B[i], C[i] using the recurrence
	A := make([]fp.Element, k+1)
	B := make([]fp.Element, k+1)
	C := make([]fp.Element, k+1)

	var tmp fp.Element
	A[0].Set(&p.X)
	C[0].Neg(&p.Y)
	tmp.Square(&p.X)
	B[0].Double(&tmp).Add(&B[0], &tmp) // B[0] = 3x²

	for i := 1; i <= k; i++ {
		var Csq, ACs, eightACs fp.Element
		Csq.Square(&C[i-1])
		ACs.Mul(&A[i-1], &Csq)
		eightACs.Double(&ACs).Double(&eightACs).Double(&eightACs)
		A[i].Square(&B[i-1]).Sub(&A[i], &eightACs)

		tmp.Square(&A[i])
		B[i].Double(&tmp).Add(&B[i], &tmp)

		var C4, fourACs, diff fp.Element
		C4.Square(&Csq)
		fourACs.Double(&ACs).Double(&fourACs)
		diff.Sub(&A[i], &fourACs)
		C[i].Double(&C4).Double(&C[i]).Double(&C[i]) // 8*C[i-1]⁴
		tmp.Mul(&B[i-1], &diff)
		C[i].Add(&C[i], &tmp) // C[i] = 8*C[i-1]⁴ + B[i-1]*(A[i] - 4*A[i-1]*C[i-1]²)
	}

	// Step 2: Compute D[i] = -2*C[i] = 2*y[i] for i = 0..k-1
	D := make([]fp.Element, k)
	for i := 0; i < k; i++ {
		D[i].Double(&C[i]).Neg(&D[i])
	}

	// Step 3: Compute T[i] = D[0]*D[1]*...*D[i] for i = 0..k-1
	T := make([]fp.Element, k)
	T[0].Set(&D[0])
	for i := 1; i < k; i++ {
		T[i].Mul(&T[i-1], &D[i])
	}

	// Step 4: Batch invert T
	invT := batchInvertFp(T)

	// Step 5: Compute line evaluations
	// Fill in REVERSE order: evaluations[k-1] = first doubling, evaluations[0] = k-th doubling
	// This matches the PrecomputeLines loop which goes from high index to low.
	// For i = 0: x[0] = A[0], y[0] = -C[0]
	// For i > 0: x[i] = A[i] / T[i-1]², y[i] = -C[i] / T[i-1]³

	// Step 0: special case since scaling is 1 (goes to evaluations[k-1])
	evaluations[k-1].R0.Mul(&B[0], &invT[0])
	evaluations[k-1].R1.Mul(&B[0], &A[0]).Mul(&evaluations[k-1].R1, &invT[0]).Add(&evaluations[k-1].R1, &C[0])

	// Steps 1 to k-1 (fill in reverse: step i goes to evaluations[k-1-i])
	var invT2, invT3 fp.Element
	for i := 1; i < k; i++ {
		idx := k - 1 - i
		// R0 = B[i] / T[i]
		evaluations[idx].R0.Mul(&B[i], &invT[i])

		// R1 = B[i]*A[i]/(T[i]*T[i-1]²) + C[i]/T[i-1]³
		invT2.Square(&invT[i-1])
		invT3.Mul(&invT2, &invT[i-1])

		var term1, term2 fp.Element
		term1.Mul(&B[i], &A[i]).Mul(&term1, &invT[i]).Mul(&term1, &invT2)
		term2.Mul(&C[i], &invT3)
		evaluations[idx].R1.Add(&term1, &term2)
	}

	// Step 6: Final point coordinates
	// x[k] = A[k] / T[k-1]²
	// y[k] = -C[k] / T[k-1]³
	invT2.Square(&invT[k-1])
	invT3.Mul(&invT2, &invT[k-1])
	p.X.Mul(&A[k], &invT2)
	p.Y.Mul(&C[k], &invT3).Neg(&p.Y)
}

// manyDoublesAndAdd performs k consecutive doublings followed by a doubleAndAdd operation,
// using a single batch inversion for all k+2 line evaluations.
// This combines manyDoubleSteps(k) + doubleAndAddStep into one batch inversion (2 inv → 1 inv).
// Note: evaluations are filled in REVERSE order to match BW6-761's convention.
func (p *G2Affine) manyDoublesAndAdd(k int, doubEvals []LineEvaluationAff, addEval1, addEval2 *LineEvaluationAff, a *G2Affine) {
	if k == 0 {
		p.doubleAndAddStep(addEval1, addEval2, a)
		return
	}

	// Step 1: Compute A[i], B[i], C[i] using the recurrence (same as manyDoubleSteps)
	A := make([]fp.Element, k+1)
	B := make([]fp.Element, k+1)
	C := make([]fp.Element, k+1)

	var tmp fp.Element
	A[0].Set(&p.X)
	C[0].Neg(&p.Y)
	tmp.Square(&p.X)
	B[0].Double(&tmp).Add(&B[0], &tmp) // B[0] = 3x²

	for i := 1; i <= k; i++ {
		var Csq, ACs, eightACs fp.Element
		Csq.Square(&C[i-1])
		ACs.Mul(&A[i-1], &Csq)
		eightACs.Double(&ACs).Double(&eightACs).Double(&eightACs)
		A[i].Square(&B[i-1]).Sub(&A[i], &eightACs)

		tmp.Square(&A[i])
		B[i].Double(&tmp).Add(&B[i], &tmp)

		var C4, fourACs, diff fp.Element
		C4.Square(&Csq)
		fourACs.Double(&ACs).Double(&fourACs)
		diff.Sub(&A[i], &fourACs)
		C[i].Double(&C4).Double(&C[i]).Double(&C[i])
		tmp.Mul(&B[i-1], &diff)
		C[i].Add(&C[i], &tmp)
	}

	// Step 2: Compute D[i] = -2*C[i] for i = 0..k-1
	D := make([]fp.Element, k)
	for i := 0; i < k; i++ {
		D[i].Double(&C[i]).Neg(&D[i])
	}

	// Step 3: Compute partial products T[i] = D[0]*D[1]*...*D[i]
	T := make([]fp.Element, k+1)
	T[0].Set(&D[0])
	for i := 1; i < k; i++ {
		T[i].Mul(&T[i-1], &D[i])
	}

	// Step 4: Compute ELM formula numerators using scaled coordinates
	var S2, S3, A_num, B_num, U_num fp.Element
	S := &T[k-1]
	S2.Square(S)
	S3.Mul(&S2, S)

	A_num.Mul(&a.X, &S2)
	A_num.Sub(&A[k], &A_num)

	B_num.Neg(&C[k])
	tmp.Mul(&a.Y, &S3)
	B_num.Sub(&B_num, &tmp)

	var A_num2, B_num2, coeff fp.Element
	A_num2.Square(&A_num)
	B_num2.Square(&B_num)
	coeff.Double(&A[k])
	tmp.Mul(&a.X, &S2)
	coeff.Add(&coeff, &tmp)
	U_num.Mul(&coeff, &A_num2)
	U_num.Sub(&B_num2, &U_num)

	// Step 5: Extend T for ELM batch inversion
	T[k].Mul(S, &A_num)
	T[k].Mul(&T[k], &U_num)

	// Step 6: Batch invert T[0..k]
	invT := batchInvertFp(T)

	// Step 7: Compute doubling line evaluations (in REVERSE order)
	// doubEvals[k-1] = first doubling, doubEvals[0] = k-th doubling
	doubEvals[k-1].R0.Mul(&B[0], &invT[0])
	doubEvals[k-1].R1.Mul(&B[0], &A[0]).Mul(&doubEvals[k-1].R1, &invT[0]).Add(&doubEvals[k-1].R1, &C[0])

	var invT2, invT3 fp.Element
	for i := 1; i < k; i++ {
		idx := k - 1 - i
		doubEvals[idx].R0.Mul(&B[i], &invT[i])

		invT2.Square(&invT[i-1])
		invT3.Mul(&invT2, &invT[i-1])

		var term1, term2 fp.Element
		term1.Mul(&B[i], &A[i]).Mul(&term1, &invT[i]).Mul(&term1, &invT2)
		term2.Mul(&C[i], &invT3)
		doubEvals[idx].R1.Add(&term1, &term2)
	}

	// Step 8: Compute point coordinates after k doublings
	invT2.Square(&invT[k-1])
	invT3.Mul(&invT2, &invT[k-1])
	var x_P, y_P fp.Element
	x_P.Mul(&A[k], &invT2)
	y_P.Mul(&C[k], &invT3).Neg(&y_P)

	// Step 9: Compute ELM slopes
	var invSA, invSU fp.Element
	invSA.Mul(&U_num, &invT[k])
	invSU.Mul(&A_num, &invT[k])

	var l1 fp.Element
	l1.Mul(&B_num, &invSA)

	var l2 fp.Element
	l2.Double(&C[k])
	l2.Neg(&l2)
	l2.Mul(&l2, &A_num2)
	l2.Mul(&l2, &invSU)
	l2.Add(&l2, &l1)
	l2.Neg(&l2)

	// Step 10: Compute ELM line evaluations
	addEval1.R0.Set(&l1)
	addEval1.R1.Mul(&l1, &x_P)
	addEval1.R1.Sub(&addEval1.R1, &y_P)

	addEval2.R0.Set(&l2)
	addEval2.R1.Mul(&l2, &x_P)
	addEval2.R1.Sub(&addEval2.R1, &y_P)

	// Step 11: Compute final point 2P + Q
	var xPQ fp.Element
	xPQ.Square(&l1)
	xPQ.Sub(&xPQ, &x_P)
	xPQ.Sub(&xPQ, &a.X)

	var x4, y4 fp.Element
	x4.Square(&l2)
	x4.Sub(&x4, &x_P)
	x4.Sub(&x4, &xPQ)

	y4.Sub(&x_P, &x4)
	y4.Mul(&l2, &y4)
	y4.Sub(&y4, &y_P)

	p.X.Set(&x4)
	p.Y.Set(&y4)
}

func (p *G2Affine) tangentCompute(evaluations *LineEvaluationAff) {

	var n, d, λ fp.Element
	// λ = 3x²/2y
	n.Square(&p.X)
	λ.Double(&n).
		Add(&λ, &n)
	d.Double(&p.Y)
	λ.Div(&λ, &d)

	evaluations.R0.Set(&λ)
	evaluations.R1.Mul(&λ, &p.X).
		Sub(&evaluations.R1, &p.Y)
}

// ------------------------
// direct-extension pairing
// ------------------------

// MillerLoopDirect computes the multi-Miller loop using the direct E6 extension
// and returns a towered E6 element. This version corresponds to gnark circuit.
// ∏ᵢ MillerLoop(Pᵢ, Qᵢ) =
// ∏ᵢ { fᵢ_{x₀+1+λ(x₀³-x₀²-x₀),Qᵢ}(Pᵢ) }
//
// Alg.2 in https://eprint.iacr.org/2021/1359.pdf
// Eq. (6') in https://hackmd.io/@gnark/BW6-761-changes
func MillerLoopDirect(P []G1Affine, Q []G2Affine) (GT, error) {
	// check input size match
	n := len(P)
	if n == 0 {
		return GT{}, errors.New("invalid inputs sizes")
	}

	lines := make([][2][len(LoopCounter) - 1]LineEvaluationAff, 0, len(Q))
	for _, qi := range Q {
		lines = append(lines, PrecomputeLines(qi))
	}

	if n != len(Q) {
		return GT{}, errors.New("invalid inputs sizes")
	}

	// no need to filter infinity points:
	// 		1. if Pᵢ=(0,0) then -x/y=1/y=0 by gnark-crypto convention and so
	// 		lines R0 and R1 are 0. It happens that result will stay, through
	// 		the Miller loop, in 𝔽p⁶ because MulBy01(0,0,1),
	// 		Mul01By01(0,0,1,0,0,1) and MulBy01245 set result.C0 to 0. At the
	// 		end result will be in a proper subgroup of Fp¹² so it be reduced to
	// 		1 in FinalExponentiation.
	//
	//      and/or
	//
	// 		2. if Qᵢ=(0,0) then PrecomputeLines(Qᵢ) will return lines R0 and R1
	// 		that are 0 because of gnark-convention (*/0==0) in doubleStep and
	// 		addStep. Similarly to Pᵢ=(0,0) it happens that result be 1
	// 		after the FinalExponentiation.

	// precomputations
	yInv := make([]fp.Element, n)
	xNegOverY := make([]fp.Element, n)
	for k := 0; k < n; k++ {
		yInv[k].Set(&P[k].Y)
	}
	yInv = fp.BatchInvert(yInv)
	for k := 0; k < n; k++ {
		xNegOverY[k].Mul(&P[k].X, &yInv[k]).
			Neg(&xNegOverY[k])
	}

	// f_{a0+λ*a1,Q}(P)
	var result fptower.E6D
	result.SetOne()

	for i := len(LoopCounter) - 2; i >= 0; i-- {
		result.Square(&result)

		j := LoopCounter1[i]*3 + LoopCounter[i]
		for k := 0; k < n; k++ {
			lines[k][0][i].R1.
				Mul(
					&lines[k][0][i].R1,
					&yInv[k],
				)
			lines[k][0][i].R0.
				Mul(&lines[k][0][i].R0,
					&xNegOverY[k],
				)
			result.MulBy023(
				&lines[k][0][i].R1,
				&lines[k][0][i].R0,
			)
			if i > 0 && j != 0 {
				lines[k][1][i].R1.
					Mul(
						&lines[k][1][i].R1,
						&yInv[k],
					)
				lines[k][1][i].R0.
					Mul(
						&lines[k][1][i].R0,
						&xNegOverY[k],
					)
				result.MulBy023(
					&lines[k][1][i].R1,
					&lines[k][1][i].R0,
				)
			}
		}
	}

	return *fptower.ToTower(&result), nil
}
