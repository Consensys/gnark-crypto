// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package bls12377

import (
	"errors"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/internal/fptower"
)

// Cubical pairing implementation based on:
// "Biextensions in Pairing-based Cryptography" by Lin, Robert, Zhao, Zheng
// https://eprint.iacr.org/2025/670
//
// For BLS12-377 (D=3, k=12), the cubical pairing formula is:
//   ab,λ(P, Q) = aλ(P, Q)² = Z^((p^12-1)/r)_[z]Q'+P'
//
// where z is the seed, Q' is a point on E'(Fp²) (the twist curve where G2 lives),
// P' = φ⁻¹(P) ∈ E'(Fp¹²), and φ is the D-type sextic twist isomorphism
// φ: E' → E, (x', y') → (x'·v, y'·vw) with inverse φ⁻¹: (x, y) → (x/v, y/(vw)).
//
// Note: BLS12-377's curve admits Montgomery form (has 2-torsion), which could
// enable faster formulas. This implementation uses the j=0 formulas for simplicity.

// cubicalPointE2 represents a point on the Kummer line K = E'/⟨±1⟩
// with projective coordinates (X : Z) over Fp²
type cubicalPointE2 struct {
	X, Z fptower.E2
}

// cubicalPointE12 represents a point on the Kummer line
// with projective coordinates (X : Z) over Fp¹²
type cubicalPointE12 struct {
	X, Z fptower.E12
}

// cDBLE2 performs x-only cubical point doubling on E': y² = x³ + b'
// Algorithm 7 from the paper for j(E) = 0 curves, operating in Fp²
// Cost: 4m₂ + 2s₂
func cDBLE2(p *cubicalPointE2, b *fptower.E2, result *cubicalPointE2) {
	var t1, t2, t3, t4, t5, t6, fourB, tmp fptower.E2

	// t1 = X²
	t1.Square(&p.X)
	// t2 = X³
	t2.Mul(&t1, &p.X)
	// t3 = Z²
	t3.Square(&p.Z)
	// t4 = Z³
	t4.Mul(&t3, &p.Z)

	// 4b
	fourB.Double(b)
	fourB.Double(&fourB)

	// t5 = X³ - 8b·Z³
	tmp.Double(&fourB) // 8b
	tmp.Mul(&tmp, &t4) // 8b·Z³
	t5.Sub(&t2, &tmp)  // X³ - 8b·Z³

	// t6 = 4X³ + 4b·Z³
	t6.Double(&t2)
	t6.Double(&t6)       // 4X³
	tmp.Mul(&fourB, &t4) // 4b·Z³
	t6.Add(&t6, &tmp)    // 4X³ + 4b·Z³

	// X[2]P = X · t5
	result.X.Mul(&p.X, &t5)
	// Z[2]P = Z · t6
	result.Z.Mul(&p.Z, &t6)
}

// cDIFFE2 performs x-only cubical differential addition on E': y² = x³ + b'
// Algorithm 8 from the paper for j(E) = 0 curves, operating in Fp²
// Computes P + Q given P, Q, and 1/X_{P-Q}
// Cost: 6m₂ + 2s₂
func cDIFFE2(p, q *cubicalPointE2, iXPminusQ, b *fptower.E2, result *cubicalPointE2) {
	var t1, t2, t3, t4, t5, t6, t7, fourB, tmp fptower.E2

	// t1 = X_P + Z_P
	t1.Add(&p.X, &p.Z)
	// t2 = X_P - Z_P
	t2.Sub(&p.X, &p.Z)
	// t3 = X_Q + Z_Q
	t3.Add(&q.X, &q.Z)
	// t4 = X_P · X_Q
	t4.Mul(&p.X, &q.X)
	// t5 = Z_P · Z_Q
	t5.Mul(&p.Z, &q.Z)
	// t6 = t1 · t3 - t4 - t5
	t6.Mul(&t1, &t3)
	t6.Sub(&t6, &t4)
	t6.Sub(&t6, &t5)
	// t7 = t2 · t3 - t4 + t5
	t7.Mul(&t2, &t3)
	t7.Sub(&t7, &t4)
	t7.Add(&t7, &t5)

	// 4b
	fourB.Double(b)
	fourB.Double(&fourB)

	// X_{P+Q} = (-4b · t5 · t6 + t4²) · iX_{P-Q}
	tmp.Mul(&fourB, &t5)
	tmp.Mul(&tmp, &t6)
	tmp.Neg(&tmp)        // -4b · t5 · t6
	result.X.Square(&t4) // t4²
	result.X.Add(&result.X, &tmp)
	result.X.Mul(&result.X, iXPminusQ)

	// Z_{P+Q} = t7²
	result.Z.Square(&t7)
}

// cDIFFE12 performs x-only cubical differential addition in Fp¹²
// This is used when one operand involves P' (which lives in Fp¹²)
// Algorithm 8 adapted for Fp¹² arithmetic
func cDIFFE12(pX, pZ, qX, qZ, iXPminusQ *fptower.E12, b *fptower.E2, result *cubicalPointE12) {
	var t1, t2, t3, t4, t5, t6, t7, fourB12, tmp fptower.E12

	// Embed 4b into E12
	var fourB fptower.E2
	fourB.Double(b)
	fourB.Double(&fourB)
	fourB12.C0.B0.Set(&fourB)

	// t1 = X_P + Z_P
	t1.Add(pX, pZ)
	// t2 = X_P - Z_P
	t2.Sub(pX, pZ)
	// t3 = X_Q + Z_Q
	t3.Add(qX, qZ)
	// t4 = X_P · X_Q
	t4.Mul(pX, qX)
	// t5 = Z_P · Z_Q
	t5.Mul(pZ, qZ)
	// t6 = t1 · t3 - t4 - t5
	t6.Mul(&t1, &t3)
	t6.Sub(&t6, &t4)
	t6.Sub(&t6, &t5)
	// t7 = t2 · t3 - t4 + t5
	t7.Mul(&t2, &t3)
	t7.Sub(&t7, &t4)
	t7.Add(&t7, &t5)

	// X_{P+Q} = (-4b · t5 · t6 + t4²) · iX_{P-Q}
	tmp.Mul(&fourB12, &t5)
	tmp.Mul(&tmp, &t6)
	// Negate: -4b · t5 · t6
	tmp.C0.Neg(&tmp.C0)
	tmp.C1.Neg(&tmp.C1)
	result.X.Square(&t4) // t4²
	result.X.Add(&result.X, &tmp)
	result.X.Mul(&result.X, iXPminusQ)

	// Z_{P+Q} = t7²
	result.Z.Square(&t7)
}

// embedE2toE12 embeds an E2 element into E12
// E12 = E6[w]/(w²-v) where E6 = E2[v]/(v³-u)
// An E2 element a is embedded as a + 0·v + 0·v² + 0·w + ...
func embedE2toE12(e2 *fptower.E2, e12 *fptower.E12) {
	*e12 = fptower.E12{}
	e12.C0.B0.Set(e2)
}

// cubicalLadderWithY computes Z_[z]Q'+P' using the cubical ladder algorithm
// Algorithm 1 from the paper
// Q' is on the twisted curve E'(Fp²), P' = φ⁻¹(P) is computed from P via the inverse twist
// yQ is the y-coordinate of Q' (in E'(Fp²))
// yPprime is y_P/(vw) in Fp¹² (the y-coordinate of P')
func cubicalLadderWithY(qPrime *cubicalPointE2, pPrimeX *fptower.E12, yQ *fptower.E2, yPprime *fptower.E12, b *fptower.E2) fptower.E12 {
	// Initialize:
	// R = Q', S = [2]Q', T = Q' + P'

	var R, S, U cubicalPointE2
	R.X.Set(&qPrime.X)
	R.Z.Set(&qPrime.Z)
	cDBLE2(qPrime, b, &S)

	// Embed Q' coordinates into E12
	var qX12, yQ12 fptower.E12
	embedE2toE12(&qPrime.X, &qX12)
	embedE2toE12(yQ, &yQ12)

	// Compute T = Q' + P' using the standard point addition formula
	// For points (x1, y1) and (x2, y2) on E': y² = x³ + b':
	// λ = (y2 - y1) / (x2 - x1)
	// x3 = λ² - x1 - x2
	//
	// Here: Q' = (qX12, yQ12), P' = (pPrimeX, yPprime)
	// So: λ = (yPprime - yQ12) / (pPrimeX - qX12)

	var lambda, xSum, xDiff, yDiff, TX, TZ fptower.E12

	// yDiff = y_P' - y_Q'
	yDiff.Sub(yPprime, &yQ12)
	// xDiff = x_P' - x_Q'
	xDiff.Sub(pPrimeX, &qX12)
	// lambda = yDiff / xDiff
	lambda.Inverse(&xDiff)
	lambda.Mul(&lambda, &yDiff)

	// x_{Q'+P'} = λ² - x_Q' - x_P'
	xSum.Add(&qX12, pPrimeX)
	TX.Square(&lambda)
	TX.Sub(&TX, &xSum)
	TZ.SetOne()

	// Compute x_{Q'-P'} for the differential when bit=1
	// Q' - P' = Q' + (-P') where -P' = (x_P', -y_P')
	// λ' = (-y_P' - y_Q') / (x_P' - x_Q') = -(y_P' + y_Q') / (x_P' - x_Q')
	// x_{Q'-P'} = λ'² - x_Q' - x_P'

	var lambdaPrime, ySum, xQminusP fptower.E12
	ySum.Add(yPprime, &yQ12)
	ySum.C0.Neg(&ySum.C0) // negate
	ySum.C1.Neg(&ySum.C1)
	lambdaPrime.Inverse(&xDiff)
	lambdaPrime.Mul(&lambdaPrime, &ySum)
	xQminusP.Square(&lambdaPrime)
	xQminusP.Sub(&xQminusP, &xSum)

	// Compute inverses needed for the ladder
	var iXPprime, iXQminusP fptower.E12
	iXPprime.Inverse(pPrimeX)
	iXQminusP.Inverse(&xQminusP)

	// Compute iX_Q' = 1/X_Q' (in E2)
	var iXQprime fptower.E2
	iXQprime.Inverse(&qPrime.X)

	// Now run the ladder
	// LoopCounter has the binary decomposition of x₀
	// We iterate from MSB to LSB

	for i := len(LoopCounter) - 2; i >= 0; i-- {
		// U = cDIFF(S, R, iX_Q') in E2
		cDIFFE2(&S, &R, &iXQprime, b, &U)

		if LoopCounter[i] == 0 {
			// T = cDIFF(T, R, iX_P') in E12
			var RX12, RZ12 fptower.E12
			embedE2toE12(&R.X, &RX12)
			embedE2toE12(&R.Z, &RZ12)

			var newT cubicalPointE12
			cDIFFE12(&TX, &TZ, &RX12, &RZ12, &iXPprime, b, &newT)
			TX.Set(&newT.X)
			TZ.Set(&newT.Z)

			// R = cDBL(R)
			cDBLE2(&R, b, &R)
			// S = U
			S.X.Set(&U.X)
			S.Z.Set(&U.Z)
		} else {
			// T = cDIFF(S, T, iX_{Q'-P'}) in E12
			var SX12, SZ12 fptower.E12
			embedE2toE12(&S.X, &SX12)
			embedE2toE12(&S.Z, &SZ12)

			var newT cubicalPointE12
			cDIFFE12(&SX12, &SZ12, &TX, &TZ, &iXQminusP, b, &newT)
			TX.Set(&newT.X)
			TZ.Set(&newT.Z)

			// S = cDBL(S)
			cDBLE2(&S, b, &S)
			// R = U
			R.X.Set(&U.X)
			R.Z.Set(&U.Z)
		}
	}

	// Return Z_[z]Q'+P'
	// The result is Z_T after the ladder
	return TZ
}

// PairCubical computes the reduced pairing using cubical arithmetic
// e(P, Q)² = Z^((p^12-1)/r)_[z]Q'+P'
//
// This computes the SQUARE of the standard pairing (level 2 cubical arithmetic).
// The result satisfies: PairCubical(P, Q) = Pair(P, Q)²
//
// This function doesn't check that the inputs are in the correct subgroup.
func PairCubical(P []G1Affine, Q []G2Affine) (GT, error) {
	n := len(P)
	if n == 0 || n != len(Q) {
		return GT{}, errors.New("invalid inputs sizes")
	}

	var result GT
	result.SetOne()

	// The twist isomorphism for BLS12-377 D-twist:
	// E: y² = x³ + 1 over Fp
	// E': y² = x³ + 1/u over Fp² (the twist curve where Q lives)
	//
	// The twist isomorphism is φ: E' → E, (x', y') → (x'·v, y'·vw)
	// where v³ = u and w² = v.
	//
	// The inverse is φ⁻¹: E → E', (x, y) → (x/v, y/(vw))
	// So for P ∈ E(Fp):
	//   x_P' = x_P / v = x_P · v² / u  (in Fp¹² since v is not in Fp)
	//   y_P' = y_P / (vw)              (in Fp¹² since vw is not in Fp)
	//
	// The E12 basis is:
	// - 1   → C0.B0
	// - v   → C0.B1
	// - v²  → C0.B2
	// - w   → C1.B0
	// - vw  → C1.B1
	// - v²w → C1.B2

	// Precompute 1/u in Fp² for computing x_P' = x_P · v² / u
	// In Fp² = Fp[u]/(u²+5), the element u has representation (0, 1)
	var uE2, invU fptower.E2
	uE2.A1.SetOne() // u = (0, 1)
	invU.Inverse(&uE2)

	// Precompute 1/(vw) in E12 for computing y_P' = y_P / (vw)
	var vw, invVW fptower.E12
	vw.C1.B1.SetOne() // vw corresponds to C1.B1
	invVW.Inverse(&vw)

	for k := 0; k < n; k++ {
		if P[k].IsInfinity() || Q[k].IsInfinity() {
			continue
		}

		// Q' is already on the twisted curve E'(Fp²)
		// Q' = Q with normalized Z = 1
		var qPrime cubicalPointE2
		qPrime.X.Set(&Q[k].X)
		qPrime.Z.SetOne()

		// P' = φ⁻¹(P) where φ: E' → E, (x', y') → (x'·v, y'·vw)
		// So φ⁻¹: (x, y) → (x/v, y/(vw))
		//
		// x_P' = x_P / v = x_P · v² / v³ = x_P · v² / u
		// In E12 representation: v² corresponds to C0.B2
		// So x_P' = (x_P / u) in the C0.B2 position

		var xPprime fptower.E12
		// Compute x_P / u in Fp²
		var xPdivU fptower.E2
		xPdivU.A0.Set(&P[k].X)
		xPdivU.Mul(&xPdivU, &invU)
		// x_P' = (x_P/u) · v²: put x_P/u in the v² coefficient (C0.B2)
		xPprime.C0.B2.Set(&xPdivU)

		// y_P' = y_P / (vw)
		// Embed y_P into E12.C0.B0 and multiply by 1/(vw)
		var yPprime fptower.E12
		var yP12 fptower.E12
		yP12.C0.B0.A0.Set(&P[k].Y)
		// y_P' = y_P * (1/vw)
		yPprime.Mul(&yP12, &invVW)

		// Compute the cubical ladder with y-coordinates for proper initialization
		zResult := cubicalLadderWithY(&qPrime, &xPprime, &Q[k].Y, &yPprime, &bTwistCurveCoeff)

		// Accumulate the result
		// For multiple pairings, we multiply the Z values
		var zResultGT GT
		zResultGT.Set(&zResult)
		result.Mul(&result, &zResultGT)
	}

	// Apply final exponentiation
	// x₀ is positive for BLS12-377, so no conjugation needed

	return FinalExponentiation(&result), nil
}

// PairCubicalCheck computes the reduced pairing and returns true if the result is one
// ∏ᵢ e(Pᵢ, Qᵢ)² =? 1
func PairCubicalCheck(P []G1Affine, Q []G2Affine) (bool, error) {
	f, err := PairCubical(P, Q)
	if err != nil {
		return false, err
	}
	var one GT
	one.SetOne()
	return f.Equal(&one), nil
}
