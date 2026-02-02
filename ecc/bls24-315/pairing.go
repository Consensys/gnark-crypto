// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package bls24315

import (
	"errors"

	"github.com/consensys/gnark-crypto/ecc/bls24-315/fp"
	"github.com/consensys/gnark-crypto/ecc/bls24-315/internal/fptower"
)

// GT target group of the pairing
type GT = fptower.E24

type lineEvaluation struct {
	r0 fptower.E4
	r1 fptower.E4
	r2 fptower.E4
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
// where d = (p²⁴-1)/r = (p²⁴-1)/Φ₂₄(p) ⋅ Φ₂₄(p)/r = (p¹²-1)(p⁴+1)(p⁸ - p⁴ +1)/r
// we use instead d=s ⋅ (p¹²-1)(p⁴+1)(p⁸ - p⁴ +1)/r
// where s is the cofactor 3 (Hayashida et al.)
func FinalExponentiation(z *GT, _z ...*GT) GT {

	var result GT
	result.Set(z)

	for _, e := range _z {
		result.Mul(&result, e)
	}

	var t [9]GT

	// Easy part
	// (p¹²-1)(p⁴+1)
	t[0].Conjugate(&result)
	result.Inverse(&result)
	t[0].Mul(&t[0], &result)
	result.FrobeniusQuad(&t[0]).
		Mul(&result, &t[0])

	var one GT
	one.SetOne()
	if result.Equal(&one) {
		return result
	}

	// Hard part (up to permutation)
	// Daiki Hayashida and Kenichiro Hayasaka
	// and Tadanori Teruya
	// https://eprint.iacr.org/2020/875.pdf
	// 3(p⁸ - p⁴ +1)/r = (x₀-1)² * (x₀+p) * (x₀²+p²) * (x₀⁴+p⁴-1) + 3
	t[0].CyclotomicSquare(&result)
	t[1].Expt(&result)
	t[2].InverseUnitary(&result)
	t[1].Mul(&t[1], &t[2])
	t[2].Expt(&t[1])
	t[1].InverseUnitary(&t[1])
	t[1].Mul(&t[1], &t[2])
	t[2].Expt(&t[1])
	t[1].Frobenius(&t[1])
	t[1].Mul(&t[1], &t[2])
	result.Mul(&result, &t[0])
	t[0].Expt(&t[1])
	t[2].Expt(&t[0])
	t[0].FrobeniusSquare(&t[1])
	t[2].Mul(&t[0], &t[2])
	t[1].Expt(&t[2])
	t[1].Expt(&t[1])
	t[1].Expt(&t[1])
	t[1].Expt(&t[1])
	t[0].FrobeniusQuad(&t[2])
	t[0].Mul(&t[0], &t[1])
	t[2].InverseUnitary(&t[2])
	t[0].Mul(&t[0], &t[2])
	result.Mul(&result, &t[0])

	return result
}

// MillerLoop computes the multi-Miller loop
// ∏ᵢ MillerLoop(Pᵢ, Qᵢ)
func MillerLoop(P []G1Affine, Q []G2Affine) (GT, error) {
	// check input size match
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

	// projective points for Q
	qProj := make([]g2Proj, n)
	qNeg := make([]G2Affine, n)
	for k := 0; k < n; k++ {
		qProj[k].FromAffine(&q[k])
		qNeg[k].Neg(&q[k])
	}

	var result GT
	result.SetOne()
	var l1, l2 lineEvaluation
	var prodLines [5]fptower.E4

	// Compute ∏ᵢ { fᵢ_{x₀,Q}(P) }
	if n >= 1 {
		// i = 31, separately to avoid an E12 Square
		// (Square(res) = 1² = 1)
		// LoopCounter[31] = 0
		// k = 0, separately to avoid MulBy034 (res × ℓ)
		// (assign line to res)

		// qProj[0] ← 2qProj[0] and l1 the tangent ℓ passing 2qProj[0]	qProj[0].doubleStep(&l1)
		qProj[0].doubleStep(&l1)
		// line evaluation at P[0] (assign)
		result.D0.C0.MulByElement(&l1.r0, &p[0].Y)
		result.D1.C0.MulByElement(&l1.r1, &p[0].X)
		result.D1.C1.Set(&l1.r2)
	}

	if n >= 2 {
		// k = 1, separately to avoid MulBy034 (res × ℓ)
		// (res is also a line at this point, so we use Mul034By034 ℓ × ℓ)

		// qProj[1] ← 2qProj[1] and l1 the tangent ℓ passing 2qProj[1]
		qProj[1].doubleStep(&l1)
		// line evaluation at P[1]
		l1.r0.MulByElement(&l1.r0, &p[1].Y)
		l1.r1.MulByElement(&l1.r1, &p[1].X)
		// ℓ × res
		prodLines = fptower.Mul034By034(&l1.r0, &l1.r1, &l1.r2, &result.D0.C0, &result.D1.C0, &result.D1.C1)
		result.D0.C0 = prodLines[0]
		result.D0.C1 = prodLines[1]
		result.D0.C2 = prodLines[2]
		result.D1.C0 = prodLines[3]
		result.D1.C1 = prodLines[4]
	}

	// k >= 2
	for k := 2; k < n; k++ {
		// qProj[k] ← 2qProj[k] and l1 the tangent ℓ passing 2qProj[k]
		qProj[k].doubleStep(&l1)
		// line evaluation at P[k]
		l1.r0.MulByElement(&l1.r0, &p[k].Y)
		l1.r1.MulByElement(&l1.r1, &p[k].X)
		// ℓ × res
		result.MulBy034(&l1.r0, &l1.r1, &l1.r2)
	}

	// i <= 30
	for i := len(LoopCounter) - 3; i >= 1; i-- {
		// (∏ᵢfᵢ)²
		// mutualize the square among n Miller loops
		result.Square(&result)

		for k := 0; k < n; k++ {
			// qProj[k] ← 2qProj[k] and l1 the tangent ℓ passing 2qProj[k]
			qProj[k].doubleStep(&l1)
			// line evaluation at P[k]
			l1.r0.MulByElement(&l1.r0, &p[k].Y)
			l1.r1.MulByElement(&l1.r1, &p[k].X)

			switch LoopCounter[i] {
			case 0:
				// ℓ × result
				result.MulBy034(&l1.r0, &l1.r1, &l1.r2)
			case 1:
				// qProj[k] ← qProj[k]+Q[k] and
				// l2 the line ℓ passing qProj[k] and Q[k]
				qProj[k].addMixedStep(&l2, &q[k])
				// line evaluation at P[k]
				l2.r0.MulByElement(&l2.r0, &p[k].Y)
				l2.r1.MulByElement(&l2.r1, &p[k].X)
				// ℓ × ℓ
				prodLines = fptower.Mul034By034(&l1.r0, &l1.r1, &l1.r2, &l2.r0, &l2.r1, &l2.r2)
				// (ℓ × ℓ) × result
				result.MulBy01234(&prodLines)
			case -1:
				// qProj[k] ← qProj[k]-Q[k] and
				// l2 the line ℓ passing qProj[k] and -Q[k]
				qProj[k].addMixedStep(&l2, &qNeg[k])
				// line evaluation at P[k]
				l2.r0.MulByElement(&l2.r0, &p[k].Y)
				l2.r1.MulByElement(&l2.r1, &p[k].X)
				// ℓ × ℓ
				prodLines = fptower.Mul034By034(&l1.r0, &l1.r1, &l1.r2, &l2.r0, &l2.r1, &l2.r2)
				// (ℓ × ℓ) × result
				result.MulBy01234(&prodLines)
			}
		}
	}

	// i = 0, separately to avoid a point addition
	// LoopCounter[0] = 1
	result.Square(&result)
	for k := 0; k < n; k++ {
		// qProj[k] ← 2qProj[k] and l1 the tangent ℓ passing 2qProj[k]
		qProj[k].doubleStep(&l1)
		// line evaluation at P[k]
		l1.r0.MulByElement(&l1.r0, &p[k].Y)
		l1.r1.MulByElement(&l1.r1, &p[k].X)

		// l2 the line passing qProj[k] and Q
		qProj[k].lineCompute(&l2, &qNeg[k])
		// line evaluation at P[k]
		l2.r0.MulByElement(&l2.r0, &p[k].Y)
		l2.r1.MulByElement(&l2.r1, &p[k].X)
		// ℓ × ℓ
		prodLines = fptower.Mul034By034(&l1.r0, &l1.r1, &l1.r2, &l2.r0, &l2.r1, &l2.r2)
		// (ℓ × ℓ) × res
		result.MulBy01234(&prodLines)
	}

	// negative x₀
	result.Conjugate(&result)

	return result, nil
}

// doubleStep doubles a point in Homogenous projective coordinates, and evaluates the line in Miller loop
// https://eprint.iacr.org/2013/722.pdf (Section 4.3)
func (p *g2Proj) doubleStep(evaluations *lineEvaluation) {

	// get some Element from our pool
	var t1, A, B, C, D, E, EE, F, G, H, I, J, K fptower.E4
	A.Mul(&p.x, &p.y)
	A.Halve()
	B.Square(&p.y)
	C.Square(&p.z)
	D.Double(&C).
		Add(&D, &C)
	E.MulBybTwistCurveCoeff(&D)
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
func (p *g2Proj) addMixedStep(evaluations *lineEvaluation, a *G2Affine) {

	// get some Element from our pool
	var Y2Z1, X2Z1, O, L, C, D, E, F, G, H, t0, t1, t2, J fptower.E4
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
func (p *g2Proj) lineCompute(evaluations *lineEvaluation, a *G2Affine) {

	// get some Element from our pool
	var Y2Z1, X2Z1, O, L, t2, J fptower.E4
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

// ----------------------
// Fixed-argument pairing
// ----------------------

type LineEvaluationAff struct {
	R0 fptower.E4
	R1 fptower.E4
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

// manyDoubleSteps performs k consecutive doublings on p and returns the line evaluations.
// It uses a recurrence to compute 2^k*P with a single batch inversion.
func (p *G2Affine) manyDoubleSteps(k int, evaluations []LineEvaluationAff) {
	if k == 0 {
		return
	}

	// Step 1: Compute A[i], B[i], C[i] using the recurrence
	A := make([]fptower.E4, k+1)
	B := make([]fptower.E4, k+1)
	C := make([]fptower.E4, k+1)

	var tmp fptower.E4
	A[0].Set(&p.X)
	C[0].Neg(&p.Y)
	tmp.Square(&p.X)
	B[0].Double(&tmp).Add(&B[0], &tmp) // B[0] = 3x²

	for i := 1; i <= k; i++ {
		var Csq, ACs, eightACs fptower.E4
		Csq.Square(&C[i-1])
		ACs.Mul(&A[i-1], &Csq)
		eightACs.Double(&ACs).Double(&eightACs).Double(&eightACs)
		A[i].Square(&B[i-1]).Sub(&A[i], &eightACs)

		tmp.Square(&A[i])
		B[i].Double(&tmp).Add(&B[i], &tmp)

		var C4, fourACs, diff fptower.E4
		C4.Square(&Csq)
		fourACs.Double(&ACs).Double(&fourACs)
		diff.Sub(&A[i], &fourACs)
		C[i].Double(&C4).Double(&C[i]).Double(&C[i]) // 8*C[i-1]⁴
		tmp.Mul(&B[i-1], &diff)
		C[i].Add(&C[i], &tmp) // C[i] = 8*C[i-1]⁴ + B[i-1]*(A[i] - 4*A[i-1]*C[i-1]²)
	}

	// Step 2: Compute D[i] = -2*C[i] = 2*y[i] for i = 0..k-1
	D := make([]fptower.E4, k)
	for i := 0; i < k; i++ {
		D[i].Double(&C[i]).Neg(&D[i])
	}

	// Step 3: Compute T[i] = D[0]*D[1]*...*D[i] for i = 0..k-1
	T := make([]fptower.E4, k)
	T[0].Set(&D[0])
	for i := 1; i < k; i++ {
		T[i].Mul(&T[i-1], &D[i])
	}

	// Step 4: Batch invert T
	invT := fptower.BatchInvertE4(T)

	// Step 5: Compute line evaluations
	// For i = 0: x[0] = A[0], y[0] = -C[0]
	// For i > 0: x[i] = A[i] / T[i-1]², y[i] = -C[i] / T[i-1]³

	// Step 0: special case since scaling is 1
	evaluations[0].R0.Mul(&B[0], &invT[0])
	evaluations[0].R1.Mul(&B[0], &A[0]).Mul(&evaluations[0].R1, &invT[0]).Add(&evaluations[0].R1, &C[0])

	// Steps 1 to k-1
	var invT2, invT3 fptower.E4
	for i := 1; i < k; i++ {
		// R0 = B[i] / T[i]
		evaluations[i].R0.Mul(&B[i], &invT[i])

		// R1 = B[i]*A[i]/(T[i]*T[i-1]²) + C[i]/T[i-1]³
		invT2.Square(&invT[i-1])
		invT3.Mul(&invT2, &invT[i-1])

		var term1, term2 fptower.E4
		term1.Mul(&B[i], &A[i]).Mul(&term1, &invT[i]).Mul(&term1, &invT2)
		term2.Mul(&C[i], &invT3)
		evaluations[i].R1.Add(&term1, &term2)
	}

	// Step 6: Final point coordinates
	// x[k] = A[k] / T[k-1]²
	// y[k] = -C[k] / T[k-1]³
	invT2.Square(&invT[k-1])
	invT3.Mul(&invT2, &invT[k-1])
	p.X.Mul(&A[k], &invT2)
	p.Y.Mul(&C[k], &invT3).Neg(&p.Y)
}

// PrecomputeLines precomputes the lines for the fixed-argument Miller loop
func PrecomputeLines(Q G2Affine) (PrecomputedLines [2][len(LoopCounter) - 1]LineEvaluationAff) {
	var accQ, negQ G2Affine
	accQ.Set(&Q)
	negQ.Neg(&Q)

	for i := len(LoopCounter) - 2; i >= 0; i-- {
		switch LoopCounter[i] {
		case 0:
			accQ.doubleStep(&PrecomputedLines[0][i])
		case 1:
			accQ.doubleAndAddStep(&PrecomputedLines[0][i], &PrecomputedLines[1][i], &Q)
		case -1:
			accQ.doubleAndAddStep(&PrecomputedLines[0][i], &PrecomputedLines[1][i], &negQ)
		}
	}
	return PrecomputedLines
}

// MillerLoopFixedQ computes the multi-Miller loop as in MillerLoop
// but Qᵢ are fixed points in G2 known in advance.
func MillerLoopFixedQ(P []G1Affine, lines [][2][len(LoopCounter) - 1]LineEvaluationAff) (GT, error) {
	n := len(P)
	if n == 0 || n != len(lines) {
		return GT{}, errors.New("invalid inputs sizes")
	}

	// no need to filter infinity points:
	// 		1. if Pᵢ=(0,0) then -x/y=1/y=0 by gnark-crypto convention and so
	// 		lines R0 and R1 are 0. At the end it happens that result will stay
	// 		1 through the Miller loop because MulBy34(1,0,0)==1
	// 		Mul34By34(1,0,0,1,0,0)==1 and MulBy01234(1,0,0,0,0)==1.
	//
	// 		2. if Qᵢ=(0,0) then PrecomputeLines(Qᵢ) will return lines R0 and R1
	// 		that are 0 because of gnark-convention (*/0==0) in doubleStep and
	// 		addStep. Similarly to Pᵢ=(0,0) it happens that result stays 1
	// 		throughout the MillerLoop.

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

	var result GT
	result.SetOne()
	var prodLines [5]E4

	// Compute ∏ᵢ { fᵢ_{x₀,Q}(P) }
	for i := len(LoopCounter) - 2; i >= 0; i-- {
		// mutualize the square among n Miller loops
		// (∏ᵢfᵢ)²
		result.Square(&result)

		for k := 0; k < n; k++ {
			// line evaluation at P[k]
			lines[k][0][i].R0.
				MulByElement(
					&lines[k][0][i].R0,
					&xNegOverY[k],
				)
			lines[k][0][i].R1.
				MulByElement(
					&lines[k][0][i].R1,
					&yInv[k],
				)

			if LoopCounter[i] == 0 {
				// ℓ × res
				result.MulBy34(
					&lines[k][0][i].R0,
					&lines[k][0][i].R1,
				)
			} else {
				// line evaluation at P[k]
				lines[k][1][i].R0.
					MulByElement(
						&lines[k][1][i].R0,
						&xNegOverY[k],
					)
				lines[k][1][i].R1.
					MulByElement(
						&lines[k][1][i].R1,
						&yInv[k],
					)
				// ℓ × ℓ
				prodLines = fptower.Mul34By34(
					&lines[k][0][i].R0, &lines[k][0][i].R1,
					&lines[k][1][i].R0, &lines[k][1][i].R1,
				)
				// (ℓ × ℓ) × res
				result.MulBy01234(&prodLines)
			}
		}
	}

	// negative x₀
	result.Conjugate(&result)

	return result, nil
}

func (p *G2Affine) doubleStep(evaluations *LineEvaluationAff) {

	var n, d, λ, xr, yr fptower.E4
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
	var n, d, λ, λλ, xr, yr fptower.E4

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
	var A, B, A2, B2, X2A2, t, U, AU, invAU, invA, invU, l1, x3, l2, x4, y4 fptower.E4

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
