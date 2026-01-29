// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

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
// where s is the cofactor (x^5-x^4-x) (El Housni and Guillevic)
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

	// Hard part (up to permutation)
	// (x₀^5-x₀^4-x₀)(p²-p+1)/r
	// Algorithm 4.5 from https://yelhousni.github.io/phd.pdf
	var a, b, c, d, e, f, g, h, i, t, mp GT
	mp.Frobenius(&result)
	a.ExptMinus1Squared(&mp)
	a.ExptSquarePlus1(&a)
	a.Mul(&result, &a)
	t.Conjugate(&mp)
	b.ExptPlus1(&a).
		Mul(&b, &t)
	t.CyclotomicSquare(&a).
		Mul(&t, &a)
	a.Conjugate(&t)
	c.ExptMinus1Div3(&b)
	d.ExptMinus1(&c)
	d.ExptSquarePlus1(&d)
	e.ExptMinus1Squared(&d)
	e.ExptSquarePlus1(&e)
	e.Mul(&e, &d)
	f.ExptPlus1(&e).
		Mul(&f, &c).
		Conjugate(&f).
		Mul(&f, &d)
	g.Mul(&f, &d).
		Conjugate(&g)
	h.ExptPlus1(&g).
		Mul(&h, &c).
		Mul(&h, &b)
	// ht = −7, hy = −1
	// c1 = (ht-hy)/2 = -3
	i.Expc1(&f).
		Mul(&i, &e)
	// c2 = (ht^2+3*hy^2)/4 = 13
	t.CyclotomicSquare(&i).
		Mul(&t, &i).
		Mul(&t, &b)
	i.Expc2(&h).
		Mul(&i, &t)
	result.Mul(&a, &i)

	return result
}

// MillerLoop computes the multi-Miller loop
// computes the multi-Miller loop ∏ᵢ MillerLoop(Pᵢ, Qᵢ)
// Alg.2 in https://eprint.iacr.org/2021/1359.pdf
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
	qProj0 := make([]g2Proj, n)
	q1 := make([]G2Affine, n)
	q1Neg := make([]G2Affine, n)
	q0Neg := make([]G2Affine, n)
	for k := 0; k < n; k++ {
		q1[k].Y.Neg(&q0[k].Y)
		q0Neg[k].X.Set(&q0[k].X)
		q0Neg[k].Y.Set(&q1[k].Y)
		q1[k].X.Mul(&q0[k].X, &thirdRootOneG2)
		qProj0[k].FromAffine(&q0[k])
		q1Neg[k].Neg(&q1[k])
	}

	// f_{a0+λ*a1,Q}(P)
	var result GT
	result.SetOne()
	var l, l0 lineEvaluation
	var prodLines [5]fp.Element

	if n >= 1 {
		// i = 157, separately to avoid an E12 Square
		// (Square(res) = 1² = 1)
		// j = 0
		// k = 0, separately to avoid MulBy014 (res × ℓ)
		// (assign line to res)
		// qProj0[0] ← 2qProj0[0] and l0 the tangent ℓ passing 2qProj0[0]
		qProj0[0].doubleStep(&l0)
		// line evaluation at Q[0] (assign)
		result.B0.A0.Set(&l0.r0)
		result.B0.A1.Mul(&l0.r1, &p[0].X)
		result.B1.A1.Mul(&l0.r2, &p[0].Y)
	}

	// k = 1
	if n >= 2 {
		// qProj0[1] ← 2qProj0[1] and l0 the tangent ℓ passing 2qProj0[1]
		qProj0[1].doubleStep(&l0)
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
		// qProj0[k] ← 2qProj0[k] and l0 the tangent ℓ passing 2qProj0[k]
		qProj0[k].doubleStep(&l0)
		// line evaluation at Q[k]
		l0.r1.Mul(&l0.r1, &p[k].X)
		l0.r2.Mul(&l0.r2, &p[k].Y)
		// ℓ × res
		result.MulBy014(&l0.r0, &l0.r1, &l0.r2)
	}

	for i := len(LoopCounter) - 3; i >= 1; i-- {
		// (∏ᵢfᵢ)²
		// mutualize the square among n Miller loops
		result.Square(&result)

		j := LoopCounter[i]*3 + LoopCounter1[i]

		for k := 0; k < n; k++ {
			// qProj0[1] ← 2pProj0[1] and l0 the tangent ℓ qassing 2pProj0[1]
			qProj0[k].doubleStep(&l0)
			// line evaluation at Q[k]
			l0.r1.Mul(&l0.r1, &p[k].X)
			l0.r2.Mul(&l0.r2, &p[k].Y)

			switch j {
			// cases -4, -2, 2, 4 do not occur given the static LoopCounters
			case -3:
				// qProj0[k] ← qProj0[k]-q1[k] and
				// l the line ℓ qassing qProj0[k] and -q1[k]
				qProj0[k].addMixedStep(&l, &q1Neg[k])
				// line evaluation at Q[k]
				l.r1.Mul(&l.r1, &p[k].X)
				l.r2.Mul(&l.r2, &p[k].Y)
				// ℓ × ℓ
				prodLines = fptower.Mul014By014(&l.r0, &l.r1, &l.r2, &l0.r0, &l0.r1, &l0.r2)
				// (ℓ × ℓ) × res
				result.MulBy01245(&prodLines)
			case -1:
				// qProj0[k] ← qProj0[k]-q0[k] and
				// l the line ℓ qassing qProj0[k] and -q0[k]
				qProj0[k].addMixedStep(&l, &q0Neg[k])
				// line evaluation at Q[k]
				l.r1.Mul(&l.r1, &p[k].X)
				l.r2.Mul(&l.r2, &p[k].Y)
				// ℓ × ℓ
				prodLines = fptower.Mul014By014(&l.r0, &l.r1, &l.r2, &l0.r0, &l0.r1, &l0.r2)
				// (ℓ × ℓ) × res
				result.MulBy01245(&prodLines)
			case 0:
				// ℓ × res
				result.MulBy014(&l0.r0, &l0.r1, &l0.r2)
			case 1:
				// qProj0[k] ← qProj0[k]+q0[k] and
				// l the line ℓ qassing qProj0[k] and q0[k]
				qProj0[k].addMixedStep(&l, &q0[k])
				// line evaluation at Q[k]
				l.r1.Mul(&l.r1, &p[k].X)
				l.r2.Mul(&l.r2, &p[k].Y)
				// ℓ × ℓ
				prodLines = fptower.Mul014By014(&l.r0, &l.r1, &l.r2, &l0.r0, &l0.r1, &l0.r2)
				// (ℓ × ℓ) × res
				result.MulBy01245(&prodLines)
			case 3:
				// qProj0[k] ← qProj0[k]+q1[k] and
				// l the line ℓ qassing qProj0[k] and q1[k]
				qProj0[k].addMixedStep(&l, &q1[k])
				// line evaluation at Q[k]
				l.r1.Mul(&l.r1, &p[k].X)
				l.r2.Mul(&l.r2, &p[k].Y)
				// (ℓ × ℓ) × res
				prodLines = fptower.Mul014By014(&l.r0, &l.r1, &l.r2, &l0.r0, &l0.r1, &l0.r2)
				// (ℓ × ℓ) × res
				result.MulBy01245(&prodLines)
			default:
				return GT{}, errors.New("invalid LoopCounter")
			}
		}
	}

	// i = 0, j = 1
	result.Square(&result)
	for k := 0; k < n; k++ {
		// qProj0[1] ← 2pProj0[1] and l0 the tangent ℓ qassing 2pProj0[1]
		qProj0[k].doubleStep(&l0)
		// line evaluation at Q[k]
		l0.r1.Mul(&l0.r1, &p[k].X)
		l0.r2.Mul(&l0.r2, &p[k].Y)
		// qProj0[k] ← qProj0[k]+q0[k] and
		// l the line ℓ qassing qProj0[k] and q0[k]
		qProj0[k].lineCompute(&l, &q0[k])
		// line evaluation at Q[k]
		l.r1.Mul(&l.r1, &p[k].X)
		l.r2.Mul(&l.r2, &p[k].Y)
		// ℓ × ℓ
		prodLines = fptower.Mul014By014(&l.r0, &l.r1, &l.r2, &l0.r0, &l0.r1, &l0.r2)
		// (ℓ × ℓ) × res
		result.MulBy01245(&prodLines)
	}

	// negative x₀
	result.Conjugate(&result)

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
	E.Double(&D).Double(&E).Double(&E)
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
	negQ.X.Set(&Q.X)
	negQ.Y.Set(&imQ.Y)
	imQ.X.Mul(&Q.X, &thirdRootOneG2)
	accQ.Set(&Q)
	imQneg.Neg(&imQ)

	// The loop processes from i=157 down to i=0
	// j = LoopCounter[i]*3 + LoopCounter1[i]
	// j values: -3, -1, 0, 1, 3

	// i=157, i=156, i=155: j=0 (3 consecutive zeros)
	accQ.manyDoubleSteps(3, PrecomputedLines[0][155:158])

	// i=154: j=-1
	accQ.doubleStep(&PrecomputedLines[0][154])
	accQ.addStep(&PrecomputedLines[1][154], &negQ)

	// i=153, i=152: j=0
	accQ.doubleStep(&PrecomputedLines[0][153])
	accQ.doubleStep(&PrecomputedLines[0][152])

	// i=151: j=1
	accQ.doubleStep(&PrecomputedLines[0][151])
	accQ.addStep(&PrecomputedLines[1][151], &Q)

	// i=150, i=149: j=0
	accQ.doubleStep(&PrecomputedLines[0][150])
	accQ.doubleStep(&PrecomputedLines[0][149])

	// i=148: j=-1
	accQ.doubleStep(&PrecomputedLines[0][148])
	accQ.addStep(&PrecomputedLines[1][148], &negQ)

	// i=147: j=0
	accQ.doubleStep(&PrecomputedLines[0][147])

	// i=146: j=1
	accQ.doubleStep(&PrecomputedLines[0][146])
	accQ.addStep(&PrecomputedLines[1][146], &Q)

	// i=145, i=144, i=143: j=0 (3 consecutive zeros)
	accQ.manyDoubleSteps(3, PrecomputedLines[0][143:146])

	// i=142: j=1
	accQ.doubleStep(&PrecomputedLines[0][142])
	accQ.addStep(&PrecomputedLines[1][142], &Q)

	// i=141, i=140: j=0
	accQ.doubleStep(&PrecomputedLines[0][141])
	accQ.doubleStep(&PrecomputedLines[0][140])

	// i=139: j=-1
	accQ.doubleStep(&PrecomputedLines[0][139])
	accQ.addStep(&PrecomputedLines[1][139], &negQ)

	// i=138: j=0
	accQ.doubleStep(&PrecomputedLines[0][138])

	// i=137: j=-1
	accQ.doubleStep(&PrecomputedLines[0][137])
	accQ.addStep(&PrecomputedLines[1][137], &negQ)

	// i=136, i=135, i=134, i=133: j=0 (4 consecutive zeros)
	accQ.manyDoubleSteps(4, PrecomputedLines[0][133:137])

	// i=132: j=-1
	accQ.doubleStep(&PrecomputedLines[0][132])
	accQ.addStep(&PrecomputedLines[1][132], &negQ)

	// i=131: j=0
	accQ.doubleStep(&PrecomputedLines[0][131])

	// i=130: j=-1
	accQ.doubleStep(&PrecomputedLines[0][130])
	accQ.addStep(&PrecomputedLines[1][130], &negQ)

	// i=129: j=0
	accQ.doubleStep(&PrecomputedLines[0][129])

	// i=128: j=1
	accQ.doubleStep(&PrecomputedLines[0][128])
	accQ.addStep(&PrecomputedLines[1][128], &Q)

	// i=127: j=0
	accQ.doubleStep(&PrecomputedLines[0][127])

	// i=126: j=1
	accQ.doubleStep(&PrecomputedLines[0][126])
	accQ.addStep(&PrecomputedLines[1][126], &Q)

	// i=125, i=124, i=123, i=122, i=121, i=120: j=0 (6 consecutive zeros)
	accQ.manyDoubleSteps(6, PrecomputedLines[0][120:126])

	// i=119: j=1
	accQ.doubleStep(&PrecomputedLines[0][119])
	accQ.addStep(&PrecomputedLines[1][119], &Q)

	// i=118, i=117: j=0
	accQ.doubleStep(&PrecomputedLines[0][118])
	accQ.doubleStep(&PrecomputedLines[0][117])

	// i=116: j=-1
	accQ.doubleStep(&PrecomputedLines[0][116])
	accQ.addStep(&PrecomputedLines[1][116], &negQ)

	// i=115: j=0
	accQ.doubleStep(&PrecomputedLines[0][115])

	// i=114: j=1
	accQ.doubleStep(&PrecomputedLines[0][114])
	accQ.addStep(&PrecomputedLines[1][114], &Q)

	// i=113: j=0
	accQ.doubleStep(&PrecomputedLines[0][113])

	// i=112: j=-1
	accQ.doubleStep(&PrecomputedLines[0][112])
	accQ.addStep(&PrecomputedLines[1][112], &negQ)

	// i=111: j=0
	accQ.doubleStep(&PrecomputedLines[0][111])

	// i=110: j=1
	accQ.doubleStep(&PrecomputedLines[0][110])
	accQ.addStep(&PrecomputedLines[1][110], &Q)

	// i=109, i=108: j=0
	accQ.doubleStep(&PrecomputedLines[0][109])
	accQ.doubleStep(&PrecomputedLines[0][108])

	// i=107: j=-1
	accQ.doubleStep(&PrecomputedLines[0][107])
	accQ.addStep(&PrecomputedLines[1][107], &negQ)

	// i=106, i=105, i=104: j=0 (3 consecutive zeros)
	accQ.manyDoubleSteps(3, PrecomputedLines[0][104:107])

	// i=103: j=-1
	accQ.doubleStep(&PrecomputedLines[0][103])
	accQ.addStep(&PrecomputedLines[1][103], &negQ)

	// i=102: j=0
	accQ.doubleStep(&PrecomputedLines[0][102])

	// i=101: j=-1
	accQ.doubleStep(&PrecomputedLines[0][101])
	accQ.addStep(&PrecomputedLines[1][101], &negQ)

	// i=100: j=0
	accQ.doubleStep(&PrecomputedLines[0][100])

	// i=99: j=1
	accQ.doubleStep(&PrecomputedLines[0][99])
	accQ.addStep(&PrecomputedLines[1][99], &Q)

	// i=98, i=97: j=0
	accQ.doubleStep(&PrecomputedLines[0][98])
	accQ.doubleStep(&PrecomputedLines[0][97])

	// i=96: j=-1
	accQ.doubleStep(&PrecomputedLines[0][96])
	accQ.addStep(&PrecomputedLines[1][96], &negQ)

	// i=95: j=0
	accQ.doubleStep(&PrecomputedLines[0][95])

	// i=94: j=-1
	accQ.doubleStep(&PrecomputedLines[0][94])
	accQ.addStep(&PrecomputedLines[1][94], &negQ)

	// i=93, i=92, i=91: j=0 (3 consecutive zeros)
	accQ.manyDoubleSteps(3, PrecomputedLines[0][91:94])

	// i=90: j=1
	accQ.doubleStep(&PrecomputedLines[0][90])
	accQ.addStep(&PrecomputedLines[1][90], &Q)

	// i=89: j=0
	accQ.doubleStep(&PrecomputedLines[0][89])

	// i=88: j=1
	accQ.doubleStep(&PrecomputedLines[0][88])
	accQ.addStep(&PrecomputedLines[1][88], &Q)

	// i=87, i=86: j=0
	accQ.doubleStep(&PrecomputedLines[0][87])
	accQ.doubleStep(&PrecomputedLines[0][86])

	// i=85: j=-1
	accQ.doubleStep(&PrecomputedLines[0][85])
	accQ.addStep(&PrecomputedLines[1][85], &negQ)

	// i=84: j=0
	accQ.doubleStep(&PrecomputedLines[0][84])

	// i=83: j=-1
	accQ.doubleStep(&PrecomputedLines[0][83])
	accQ.addStep(&PrecomputedLines[1][83], &negQ)

	// i=82: j=0
	accQ.doubleStep(&PrecomputedLines[0][82])

	// i=81: j=-1
	accQ.doubleStep(&PrecomputedLines[0][81])
	accQ.addStep(&PrecomputedLines[1][81], &negQ)

	// i=80: j=0
	accQ.doubleStep(&PrecomputedLines[0][80])

	// i=79: j=1
	accQ.doubleStep(&PrecomputedLines[0][79])
	accQ.addStep(&PrecomputedLines[1][79], &Q)

	// i=78, i=77, i=76: j=0 (3 consecutive zeros)
	accQ.manyDoubleSteps(3, PrecomputedLines[0][76:79])

	// i=75: j=-1
	accQ.doubleStep(&PrecomputedLines[0][75])
	accQ.addStep(&PrecomputedLines[1][75], &negQ)

	// i=74: j=0
	accQ.doubleStep(&PrecomputedLines[0][74])

	// i=73: j=1
	accQ.doubleStep(&PrecomputedLines[0][73])
	accQ.addStep(&PrecomputedLines[1][73], &Q)

	// i=72: j=0
	accQ.doubleStep(&PrecomputedLines[0][72])

	// i=71: j=-1
	accQ.doubleStep(&PrecomputedLines[0][71])
	accQ.addStep(&PrecomputedLines[1][71], &negQ)

	// i=70, i=69: j=0
	accQ.doubleStep(&PrecomputedLines[0][70])
	accQ.doubleStep(&PrecomputedLines[0][69])

	// i=68: j=-1
	accQ.doubleStep(&PrecomputedLines[0][68])
	accQ.addStep(&PrecomputedLines[1][68], &negQ)

	// i=67: j=0
	accQ.doubleStep(&PrecomputedLines[0][67])

	// i=66: j=1
	accQ.doubleStep(&PrecomputedLines[0][66])
	accQ.addStep(&PrecomputedLines[1][66], &Q)

	// i=65, i=64: j=0
	accQ.doubleStep(&PrecomputedLines[0][65])
	accQ.doubleStep(&PrecomputedLines[0][64])

	// i=63: j=-1
	accQ.doubleStep(&PrecomputedLines[0][63])
	accQ.addStep(&PrecomputedLines[1][63], &negQ)

	// i=62: j=0
	accQ.doubleStep(&PrecomputedLines[0][62])

	// i=61: j=1
	accQ.doubleStep(&PrecomputedLines[0][61])
	accQ.addStep(&PrecomputedLines[1][61], &Q)

	// i=60, i=59, i=58, i=57: j=0 (4 consecutive zeros)
	accQ.manyDoubleSteps(4, PrecomputedLines[0][57:61])

	// i=56: j=1
	accQ.doubleStep(&PrecomputedLines[0][56])
	accQ.addStep(&PrecomputedLines[1][56], &Q)

	// i=55, i=54: j=0
	accQ.doubleStep(&PrecomputedLines[0][55])
	accQ.doubleStep(&PrecomputedLines[0][54])

	// i=53: j=1
	accQ.doubleStep(&PrecomputedLines[0][53])
	accQ.addStep(&PrecomputedLines[1][53], &Q)

	// i=52, i=51, i=50, i=49, i=48, i=47, i=46: j=0 (7 consecutive zeros)
	accQ.manyDoubleSteps(7, PrecomputedLines[0][46:53])

	// i=45: j=-1
	accQ.doubleStep(&PrecomputedLines[0][45])
	accQ.addStep(&PrecomputedLines[1][45], &negQ)

	// i=44, i=43: j=0
	accQ.doubleStep(&PrecomputedLines[0][44])
	accQ.doubleStep(&PrecomputedLines[0][43])

	// i=42: j=-1
	accQ.doubleStep(&PrecomputedLines[0][42])
	accQ.addStep(&PrecomputedLines[1][42], &negQ)

	// i=41, i=40, i=39, i=38, i=37, i=36, i=35, i=34, i=33: j=0 (9 consecutive zeros)
	accQ.manyDoubleSteps(9, PrecomputedLines[0][33:42])

	// i=32: j=3
	accQ.doubleStep(&PrecomputedLines[0][32])
	accQ.addStep(&PrecomputedLines[1][32], &imQ)

	// i=31: j=0
	accQ.doubleStep(&PrecomputedLines[0][31])

	// i=30: j=-3
	accQ.doubleStep(&PrecomputedLines[0][30])
	accQ.addStep(&PrecomputedLines[1][30], &imQneg)

	// i=29, i=28, i=27, i=26, i=25, i=24, i=23: j=0 (7 consecutive zeros)
	accQ.manyDoubleSteps(7, PrecomputedLines[0][23:30])

	// i=22: j=-3
	accQ.doubleStep(&PrecomputedLines[0][22])
	accQ.addStep(&PrecomputedLines[1][22], &imQneg)

	// i=21: j=0
	accQ.doubleStep(&PrecomputedLines[0][21])

	// i=20: j=3
	accQ.doubleStep(&PrecomputedLines[0][20])
	accQ.addStep(&PrecomputedLines[1][20], &imQ)

	// i=19, i=18, i=17, i=16, i=15, i=14, i=13, i=12, i=11, i=10, i=9, i=8, i=7, i=6, i=5, i=4, i=3, i=2: j=0 (18 consecutive zeros)
	accQ.manyDoubleSteps(18, PrecomputedLines[0][2:20])

	// i=1: j=-3
	accQ.doubleStep(&PrecomputedLines[0][1])
	accQ.addStep(&PrecomputedLines[1][1], &imQneg)

	// i=0: j=1
	accQ.doubleStep(&PrecomputedLines[0][0])
	accQ.addStep(&PrecomputedLines[1][0], &Q)

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

		j := LoopCounter[i]*3 + LoopCounter1[i]
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

	// negative x₀
	result.Conjugate(&result)

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

// doubleAndAddStep computes 2P+Q in affine coordinates and outputs
// the line evaluations for both the doubling and addition.
// Uses Eisenträger-Lauter-Montgomery formula (Algorithm 4 in https://eprint.iacr.org/2022/1162).
func (p *G2Affine) doubleAndAddStep(evaluations1, evaluations2 *LineEvaluationAff, a *G2Affine) {
	var A, B, A2, B2, X2A2, t, U, AU, invAU, invA, invU, l1, x3, l2, x4, y4 fp.Element

	A.Sub(&p.X, &a.X)
	B.Sub(&p.Y, &a.Y)
	A2.Square(&A)
	B2.Square(&B)
	t.Double(&p.X).Add(&t, &a.X)
	X2A2.Mul(&t, &A2)
	U.Sub(&B2, &X2A2)

	AU.Mul(&A, &U)
	invAU.Inverse(&AU)
	invA.Mul(&U, &invAU)
	invU.Mul(&A, &invAU)

	l1.Mul(&B, &invA)
	x3.Square(&l1)
	x3.Sub(&x3, &p.X)
	x3.Sub(&x3, &a.X)

	evaluations1.R0.Set(&l1)
	evaluations1.R1.Mul(&l1, &p.X)
	evaluations1.R1.Sub(&evaluations1.R1, &p.Y)

	l2.Double(&p.Y)
	l2.Mul(&l2, &A2)
	l2.Mul(&l2, &invU)
	l2.Add(&l2, &l1)
	l2.Neg(&l2)

	x4.Square(&l2)
	x4.Sub(&x4, &p.X)
	x4.Sub(&x4, &x3)

	y4.Sub(&p.X, &x4)
	y4.Mul(&l2, &y4)
	y4.Sub(&y4, &p.Y)

	evaluations2.R0.Set(&l2)
	evaluations2.R1.Mul(&l2, &p.X)
	evaluations2.R1.Sub(&evaluations2.R1, &p.Y)

	p.X.Set(&x4)
	p.Y.Set(&y4)
}

// batchInvertFp computes the batch inverse of a slice of field elements.
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

// manyDoubleSteps performs k consecutive doublings on p and returns the line evaluations.
// It uses a recurrence to compute 2^k*P with a single batch inversion.
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
