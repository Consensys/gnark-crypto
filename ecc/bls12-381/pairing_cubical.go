// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package bls12381

import (
	"errors"

	"github.com/consensys/gnark-crypto/ecc/bls12-381/internal/fptower"
)

// Cubical pairing implementation based on:
// "Biextensions in Pairing-based Cryptography" by Lin, Robert, Zhao, Zheng
// https://eprint.iacr.org/2025/670
//
// For BLS12-381 (D=3, k=12), the cubical pairing formula is:
//   ab,λ(P, Q) = aλ(P, Q)² = Z^((p^12-1)/r)_[z]Q'+P'
//
// where z is the seed, Q' is a point on E'(Fp²) (the twist curve where G2 lives),
// P' = φ⁻¹(P) ∈ E'(Fp¹²), and φ is the M-type sextic twist isomorphism
// φ: E' → E, (x', y') → (x'/v, y'/(vw)) with inverse φ⁻¹: (x, y) → (x·v, y·vw).

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

// mulE12BySparseE2 multiplies a dense E12 by a sparse E12 where only C0.B0 is non-zero
// This is much faster than full E12 multiplication: 6 E2 muls vs ~54 for full mul
// sparse = (a, 0, 0, 0, 0, 0) where a ∈ E2
// result = dense * sparse
func mulE12BySparseE2(dense *fptower.E12, sparse *fptower.E2, result *fptower.E12) {
	// E12 = E6[w]/(w²-v), E6 = E2[v]/(v³-(1+u))
	// sparse embedded: C0 = (sparse, 0, 0), C1 = (0, 0, 0)
	// dense: C0 = (b0, b1, b2), C1 = (b3, b4, b5)
	//
	// result.C0 = sparse.C0 * dense.C0 = (sparse,0,0) * (b0,b1,b2)
	//           = (sparse*b0, sparse*b1, sparse*b2)
	// result.C1 = sparse.C0 * dense.C1 = (sparse,0,0) * (b3,b4,b5)
	//           = (sparse*b3, sparse*b4, sparse*b5)
	result.C0.B0.Mul(sparse, &dense.C0.B0)
	result.C0.B1.Mul(sparse, &dense.C0.B1)
	result.C0.B2.Mul(sparse, &dense.C0.B2)
	result.C1.B0.Mul(sparse, &dense.C1.B0)
	result.C1.B1.Mul(sparse, &dense.C1.B1)
	result.C1.B2.Mul(sparse, &dense.C1.B2)
}

// cDIFFE12_SparseQ performs cDIFF where qX, qZ are sparse (embedded from E2)
// This exploits the sparse structure for significant speedup
// p coordinates are dense (from T), q coordinates are sparse (from precomputation)
// fourB should be precomputed as 4*b
func cDIFFE12_SparseQ(pX, pZ *fptower.E12, qX, qZ *fptower.E2, iXPminusQ *fptower.E12, fourB *fptower.E2, result *cubicalPointE12) {
	var t1, t2, t4, t5, t6, t7, tmp fptower.E12
	var t3, qSum fptower.E2

	// t1 = X_P + Z_P (dense)
	t1.Add(pX, pZ)
	// t2 = X_P - Z_P (dense)
	t2.Sub(pX, pZ)
	// t3 = X_Q + Z_Q (sparse, stays in E2)
	t3.Add(qX, qZ)
	// qSum for later use
	qSum.Set(&t3)

	// t4 = X_P · X_Q (sparse × dense = 6 E2 muls)
	mulE12BySparseE2(pX, qX, &t4)
	// t5 = Z_P · Z_Q (sparse × dense = 6 E2 muls)
	mulE12BySparseE2(pZ, qZ, &t5)

	// t6 = t1 · t3 - t4 - t5 (sparse × dense - dense - dense)
	mulE12BySparseE2(&t1, &qSum, &t6)
	t6.Sub(&t6, &t4)
	t6.Sub(&t6, &t5)

	// t7 = t2 · t3 - t4 + t5 (sparse × dense - dense + dense)
	mulE12BySparseE2(&t2, &qSum, &t7)
	t7.Sub(&t7, &t4)
	t7.Add(&t7, &t5)

	// X_{P+Q} = (-4b · t5 · t6 + t4²) · iX_{P-Q}
	// 4b is sparse, t5 is dense
	mulE12BySparseE2(&t5, fourB, &tmp) // sparse × dense
	tmp.Mul(&tmp, &t6)                 // dense × dense
	// Negate
	tmp.C0.Neg(&tmp.C0)
	tmp.C1.Neg(&tmp.C1)
	result.X.Square(&t4)
	result.X.Add(&result.X, &tmp)
	result.X.Mul(&result.X, iXPminusQ)

	// Z_{P+Q} = t7²
	result.Z.Square(&t7)
}

// cDIFFE12_SparseP performs cDIFF where pX, pZ are sparse (embedded from E2)
// p coordinates are sparse (from precomputation), q coordinates are dense (from T)
// fourB should be precomputed as 4*b
func cDIFFE12_SparseP(pX, pZ *fptower.E2, qX, qZ, iXPminusQ *fptower.E12, fourB *fptower.E2, result *cubicalPointE12) {
	var t3, t4, t5, t6, t7, tmp fptower.E12
	var t1, t2, pSum, pDiff fptower.E2

	// t1 = X_P + Z_P (sparse, stays in E2)
	t1.Add(pX, pZ)
	pSum.Set(&t1)
	// t2 = X_P - Z_P (sparse, stays in E2)
	t2.Sub(pX, pZ)
	pDiff.Set(&t2)
	// t3 = X_Q + Z_Q (dense)
	t3.Add(qX, qZ)

	// t4 = X_P · X_Q (sparse × dense = 6 E2 muls)
	mulE12BySparseE2(qX, pX, &t4)
	// t5 = Z_P · Z_Q (sparse × dense = 6 E2 muls)
	mulE12BySparseE2(qZ, pZ, &t5)

	// t6 = t1 · t3 - t4 - t5 (sparse × dense - dense - dense)
	mulE12BySparseE2(&t3, &pSum, &t6)
	t6.Sub(&t6, &t4)
	t6.Sub(&t6, &t5)

	// t7 = t2 · t3 - t4 + t5 (sparse × dense - dense + dense)
	mulE12BySparseE2(&t3, &pDiff, &t7)
	t7.Sub(&t7, &t4)
	t7.Add(&t7, &t5)

	// X_{P+Q} = (-4b · t5 · t6 + t4²) · iX_{P-Q}
	mulE12BySparseE2(&t5, fourB, &tmp) // sparse × dense
	tmp.Mul(&tmp, &t6)                 // dense × dense
	// Negate
	tmp.C0.Neg(&tmp.C0)
	tmp.C1.Neg(&tmp.C1)
	result.X.Square(&t4)
	result.X.Add(&result.X, &tmp)
	result.X.Mul(&result.X, iXPminusQ)

	// Z_{P+Q} = t7²
	result.Z.Square(&t7)
}

// TwoBitCoeffs holds precomputed sparse coefficients for 2-bit windowed cDIFF
// These are computed from consecutive ladder points (a0, a1) for the (0,0) bit case
type TwoBitCoeffs struct {
	A0      fptower.E2 // a0 (X coordinate of R at first bit)
	A1      fptower.E2 // a1 (X coordinate of R at second bit)
	A0sq    fptower.E2 // a0²
	A0A1    fptower.E2 // a0 * a1
	A0plus1 fptower.E2 // a0 + 1 (since Z=1 for ladder points)
	A1plus1 fptower.E2 // a1 + 1
}

// cDIFF2_SparseQQ performs two consecutive cDIFF operations for (0,0) bits
// This computes T₂ = cDIFF(cDIFF(T₀, R₀, d), R₁, d) more efficiently
// by computing X·Z and Z² only once and using precomputed sparse coefficients
//
// Inputs:
//   - pX, pZ: T₀ coordinates (dense E12)
//   - coeffs: precomputed 2-bit coefficients (sparse)
//   - iXPminusQ: 1/X_{P'} (the differential inverse, sparse in E12)
//   - fourB: 4*b' precomputed
//
// The key optimization is that for bit=0 case with R=(a,1):
//
//	t5·t6 = Z·(X + Z·a) = X·Z + Z²·a
//
// By computing X·Z and Z² once, subsequent operations use sparse×dense
func cDIFF2_SparseQQ(pX, pZ *fptower.E12, coeffs *TwoBitCoeffs, iXPminusQ *fptower.E12, fourB *fptower.E2, result *cubicalPointE12) {
	var XZ, Zsq fptower.E12
	var t4_0, t5t6_0, t7_0, X1num, Z1 fptower.E12
	var t4_1, t5t6_1, t7_1 fptower.E12
	var tmp fptower.E12

	// Phase 1: Compute base products from T₀ (computed ONCE)
	XZ.Mul(pX, pZ) // X·Z (dense × dense) - computed only once!
	Zsq.Square(pZ) // Z² (dense square)

	// === First cDIFF (T₀ + R₀) ===
	// R₀ = (a0, 1), so X_Q = a0, Z_Q = 1
	//
	// From cDIFF formula:
	// t4 = X · X_Q = X · a0
	// t5 = Z · Z_Q = Z · 1 = Z
	// t6 = X·Z_Q + Z·X_Q = X + Z·a0
	// t7 = X·Z_Q - Z·X_Q = X - Z·a0
	//
	// t5·t6 = Z·(X + Z·a0) = XZ + Z²·a0

	// t4_0 = X · a0 (sparse × dense)
	mulE12BySparseE2(pX, &coeffs.A0, &t4_0)

	// t5·t6 = Z·(X + Z·a0) = XZ + Z²·a0
	mulE12BySparseE2(&Zsq, &coeffs.A0, &tmp)
	t5t6_0.Add(&XZ, &tmp)

	// t7_0 = X - Z·a0
	mulE12BySparseE2(pZ, &coeffs.A0, &tmp) // Z·a0
	t7_0.Sub(pX, &tmp)                     // X - Z·a0

	// Z1 = t7_0²
	Z1.Square(&t7_0)

	// X1_num = t4_0² - 4b·t5·t6_0
	mulE12BySparseE2(&t5t6_0, fourB, &tmp)
	X1num.Square(&t4_0)
	X1num.Sub(&X1num, &tmp)

	// X1 = X1_num · d
	var X1 fptower.E12
	X1.Mul(&X1num, iXPminusQ)

	// === Second cDIFF (T₁ + R₁) ===
	// T₁ = (X1, Z1), R₁ = (a1, 1)

	// For the second step, we need new XZ and Zsq from T₁
	var XZ1, Z1sq fptower.E12
	XZ1.Mul(&X1, &Z1) // X1·Z1 (dense × dense)
	Z1sq.Square(&Z1)  // Z1² (dense square)

	// t4_1 = X1 · a1
	mulE12BySparseE2(&X1, &coeffs.A1, &t4_1)

	// t5t6_1 = Z1·(X1 + Z1·a1) = X1·Z1 + Z1²·a1
	mulE12BySparseE2(&Z1sq, &coeffs.A1, &tmp)
	t5t6_1.Add(&XZ1, &tmp)

	// t7_1 = X1 - Z1·a1
	mulE12BySparseE2(&Z1, &coeffs.A1, &tmp)
	t7_1.Sub(&X1, &tmp)

	// Z2 = t7_1²
	result.Z.Square(&t7_1)

	// X2 = (t4_1² - 4b·t5t6_1) · d
	mulE12BySparseE2(&t5t6_1, fourB, &tmp)
	result.X.Square(&t4_1)
	result.X.Sub(&result.X, &tmp)
	result.X.Mul(&result.X, iXPminusQ)
}

// embedE2toE12 embeds an E2 element into E12
// E12 = E6[w]/(w²-v) where E6 = E2[v]/(v³-(1+u))
// An E2 element a is embedded as a + 0·v + 0·v² + 0·w + ...
func embedE2toE12(e2 *fptower.E2, e12 *fptower.E12) {
	*e12 = fptower.E12{}
	e12.C0.B0.Set(e2)
}

// cubicalLadderWithY computes Z_[z]Q'+P' using the cubical ladder algorithm
// Algorithm 1 from the paper
// Q' is on the twisted curve E'(Fp²), P' = φ⁻¹(P) is computed from P via the inverse twist
// yQ is the y-coordinate of Q' (in E'(Fp²))
// yPprime is y_P · vw in Fp¹² (the y-coordinate of P')
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
	// LoopCounter has the binary decomposition of |x₀|
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

	// The twist isomorphism for BLS12-381 M-twist:
	// E: y² = x³ + 4 over Fp
	// E': y² = x³ + 4(1+u) over Fp² (the twist curve where Q lives)
	//
	// The twist isomorphism is φ: E' → E, (x', y') → (x'/v, y'/(vw))
	// where v³ = (1+u) and w² = v.
	//
	// The inverse is φ⁻¹: E → E', (x, y) → (x·v, y·vw)
	// So for P ∈ E(Fp):
	//   x_P' = x_P · v  (in Fp¹² since v is not in Fp)
	//   y_P' = y_P · vw (in Fp¹² since vw is not in Fp)
	//
	// The E12 basis is:
	// - 1   → C0.B0
	// - v   → C0.B1
	// - v²  → C0.B2
	// - w   → C1.B0
	// - vw  → C1.B1
	// - v²w → C1.B2

	for k := 0; k < n; k++ {
		if P[k].IsInfinity() || Q[k].IsInfinity() {
			continue
		}

		// Q' is already on the twisted curve E'(Fp²)
		// Q' = Q with normalized Z = 1
		var qPrime cubicalPointE2
		qPrime.X.Set(&Q[k].X)
		qPrime.Z.SetOne()

		// P' = φ⁻¹(P) where φ: E' → E, (x', y') → (x'/v, y'/(vw))
		// So φ⁻¹: (x, y) → (x·v, y·vw)
		//
		// x_P' = x_P · v
		// In E12 representation: v corresponds to C0.B1
		// So x_P' = x_P in the C0.B1 position

		var xPprime fptower.E12
		// x_P' = x_P · v: put x_P in the v coefficient (C0.B1)
		xPprime.C0.B1.A0.Set(&P[k].X)

		// y_P' = y_P · vw
		// In E12 representation: vw corresponds to C1.B1
		// So y_P' = y_P in the C1.B1 position
		var yPprime fptower.E12
		// y_P' = y_P · vw: put y_P in the vw coefficient (C1.B1)
		yPprime.C1.B1.A0.Set(&P[k].Y)

		// Compute the cubical ladder with y-coordinates for proper initialization
		zResult := cubicalLadderWithY(&qPrime, &xPprime, &Q[k].Y, &yPprime, &bTwistCurveCoeff)

		// Accumulate the result
		// For multiple pairings, we multiply the Z values
		var zResultGT GT
		zResultGT.Set(&zResult)
		result.Mul(&result, &zResultGT)
	}

	// Apply final exponentiation
	// Since x₀ is negative for BLS12-381, we need to conjugate
	result.Conjugate(&result)

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

// ============================================================================
// Fixed-argument cubical pairing (Q precomputed)
// ============================================================================

// G2CubicalPrecompute contains precomputed values for a fixed G2 point Q
// used to accelerate cubical pairing computation when Q is reused.
type G2CubicalPrecompute struct {
	// Q' coordinates on the twist curve E'(Fp²)
	QprimeX fptower.E2
	QprimeY fptower.E2

	// Q' coordinates embedded in E12 (avoids repeated embedding)
	QprimeX12 fptower.E12
	QprimeY12 fptower.E12

	// Precomputed ladder states in E2 (sparse - for optimized multiplication)
	LadderX []fptower.E2
	LadderZ []fptower.E2

	// Which point (R or S) is stored at each index
	// true = R (for bit 0), false = S (for bit 1)
	LadderIsR []bool

	// The twist coefficient b' = 4(1+u)
	BTwist fptower.E2

	// Precomputed 4*b' for cDIFF operations (avoids recomputation)
	FourBTwist fptower.E2

	// Precomputed 2-bit coefficients for consecutive (0,0) bits
	// TwoBitCoeffs[i] contains coefficients for processing bits i and i+1 together
	// Only populated for indices where both bits are 0
	TwoBitCoeffs []TwoBitCoeffs

	// TwoBitValid[i] is true if TwoBitCoeffs[i] is valid (both bits are 0)
	TwoBitValid []bool
}

// PrecomputeG2Cubical precomputes values for a fixed G2 point Q
// to accelerate subsequent cubical pairing computations with varying P.
func PrecomputeG2Cubical(Q *G2Affine) *G2CubicalPrecompute {
	if Q.IsInfinity() {
		return nil
	}

	pre := &G2CubicalPrecompute{}

	// Store Q' = Q (already on twist curve)
	pre.QprimeX.Set(&Q.X)
	pre.QprimeY.Set(&Q.Y)

	// Embed Q' into E12
	embedE2toE12(&Q.X, &pre.QprimeX12)
	embedE2toE12(&Q.Y, &pre.QprimeY12)

	// Store twist coefficient and precompute 4*b'
	pre.BTwist.Set(&bTwistCurveCoeff)
	pre.FourBTwist.Double(&pre.BTwist)
	pre.FourBTwist.Double(&pre.FourBTwist)

	// Compute 1/X_Q' for E2 ladder
	var invXQprime fptower.E2
	invXQprime.Inverse(&Q.X)

	// Run the Q-only ladder to precompute all R and S states
	// Store E2 values directly for sparse multiplication optimization
	numIterations := len(LoopCounter) - 1
	pre.LadderX = make([]fptower.E2, numIterations)
	pre.LadderZ = make([]fptower.E2, numIterations)
	pre.LadderIsR = make([]bool, numIterations)

	// Initialize: R = Q', S = [2]Q'
	var R, S, U cubicalPointE2
	R.X.Set(&Q.X)
	R.Z.SetOne()
	cDBLE2(&R, &pre.BTwist, &S)

	// Store initial state (in E2 for sparse multiplication)
	i0 := len(LoopCounter) - 2
	if LoopCounter[i0] == 0 {
		pre.LadderX[0].Set(&R.X)
		pre.LadderZ[0].Set(&R.Z)
		pre.LadderIsR[0] = true
	} else {
		pre.LadderX[0].Set(&S.X)
		pre.LadderZ[0].Set(&S.Z)
		pre.LadderIsR[0] = false
	}

	// Run the ladder storing the needed point at each iteration
	for idx := 1; idx < numIterations; idx++ {
		i := len(LoopCounter) - 2 - (idx - 1)

		// U = cDIFF(S, R, 1/X_Q')
		cDIFFE2(&S, &R, &invXQprime, &pre.BTwist, &U)

		if LoopCounter[i] == 0 {
			// R = cDBL(R), S = U
			cDBLE2(&R, &pre.BTwist, &R)
			S.X.Set(&U.X)
			S.Z.Set(&U.Z)
		} else {
			// S = cDBL(S), R = U
			cDBLE2(&S, &pre.BTwist, &S)
			R.X.Set(&U.X)
			R.Z.Set(&U.Z)
		}

		// Store the point needed for next iteration (in E2 for sparse multiplication)
		iNext := len(LoopCounter) - 2 - idx
		if iNext >= 0 && LoopCounter[iNext] == 0 {
			pre.LadderX[idx].Set(&R.X)
			pre.LadderZ[idx].Set(&R.Z)
			pre.LadderIsR[idx] = true
		} else {
			pre.LadderX[idx].Set(&S.X)
			pre.LadderZ[idx].Set(&S.Z)
			pre.LadderIsR[idx] = false
		}
	}

	// Normalize all ladder points to affine (Z=1) using batch inversion
	// This enables the simplified 2-bit windowed cDIFF formulas
	// Montgomery's trick: compute all 1/Z_i using only 1 inversion + O(n) multiplications
	if numIterations > 0 {
		// Build products: prod[i] = Z_0 * Z_1 * ... * Z_i
		products := make([]fptower.E2, numIterations)
		products[0].Set(&pre.LadderZ[0])
		for i := 1; i < numIterations; i++ {
			products[i].Mul(&products[i-1], &pre.LadderZ[i])
		}

		// Compute single inversion of the total product
		var invTotal fptower.E2
		invTotal.Inverse(&products[numIterations-1])

		// Compute individual inverses and normalize X coordinates
		// Working backwards: inv[i] = invTotal * (Z_{i+1} * ... * Z_{n-1})
		for i := numIterations - 1; i >= 0; i-- {
			var invZ fptower.E2
			if i == 0 {
				invZ.Set(&invTotal)
			} else {
				invZ.Mul(&invTotal, &products[i-1])
				// Update invTotal for next iteration: invTotal *= Z_i
				invTotal.Mul(&invTotal, &pre.LadderZ[i])
			}
			// Normalize: X = X/Z (now in affine form)
			pre.LadderX[i].Mul(&pre.LadderX[i], &invZ)
			// Set Z = 1 (affine form)
			pre.LadderZ[i].SetOne()
		}
	}

	// Precompute 2-bit coefficients for consecutive (0,0) bits
	// This enables the optimized 2-bit windowed cDIFF
	// Now all ladder points have Z=1 (affine form)
	pre.TwoBitCoeffs = make([]TwoBitCoeffs, numIterations)
	pre.TwoBitValid = make([]bool, numIterations)

	for idx := 0; idx < numIterations-1; idx++ {
		i := len(LoopCounter) - 2 - idx
		iNext := i - 1

		// Check if both this bit and next bit are 0 (the (0,0) case)
		if iNext >= 0 && LoopCounter[i] == 0 && LoopCounter[iNext] == 0 {
			// Both bits are 0, so we can use 2-bit windowed processing
			// a0 = LadderX[idx] (X coordinate of R at first bit, Z=1)
			// a1 = LadderX[idx+1] (X coordinate of R at second bit, Z=1)
			pre.TwoBitValid[idx] = true

			a0 := &pre.LadderX[idx]
			a1 := &pre.LadderX[idx+1]

			pre.TwoBitCoeffs[idx].A0.Set(a0)
			pre.TwoBitCoeffs[idx].A1.Set(a1)
			pre.TwoBitCoeffs[idx].A0sq.Square(a0)
			pre.TwoBitCoeffs[idx].A0A1.Mul(a0, a1)

			// a0 + 1 and a1 + 1 (Z=1 for all normalized ladder points)
			var one fptower.E2
			one.SetOne()
			pre.TwoBitCoeffs[idx].A0plus1.Add(a0, &one)
			pre.TwoBitCoeffs[idx].A1plus1.Add(a1, &one)
		}
	}

	return pre
}

// MillerLoopCubicalFixedQ computes the Miller loop part of the cubical pairing without final exponentiation.
func MillerLoopCubicalFixedQ(P *G1Affine, pre *G2CubicalPrecompute) (GT, error) {
	var result GT

	if P.IsInfinity() || pre == nil {
		result.SetOne()
		return result, nil
	}

	// Compute P' = φ⁻¹(P) for M-type twist
	// x_P' = x_P · v (v coefficient is C0.B1)
	var xPprime fptower.E12
	xPprime.C0.B1.A0.Set(&P.X)

	// y_P' = y_P · vw (vw coefficient is C1.B1)
	var yPprime fptower.E12
	yPprime.C1.B1.A0.Set(&P.Y)

	// Use precomputed Q' embeddings
	qX12 := &pre.QprimeX12
	yQ12 := &pre.QprimeY12

	// Compute T = Q' + P'
	var lambda, xSum, xDiff, yDiff, TX, TZ, invXDiff fptower.E12

	yDiff.Sub(&yPprime, yQ12)
	xDiff.Sub(&xPprime, qX12)
	invXDiff.Inverse(&xDiff) // Compute inverse once and reuse
	lambda.Mul(&invXDiff, &yDiff)

	xSum.Add(qX12, &xPprime)
	TX.Square(&lambda)
	TX.Sub(&TX, &xSum)
	TZ.SetOne()

	// Compute x_{Q'-P'}
	var lambdaPrime, ySum, xQminusP fptower.E12
	ySum.Add(&yPprime, yQ12)
	ySum.C0.Neg(&ySum.C0)
	ySum.C1.Neg(&ySum.C1)
	lambdaPrime.Mul(&invXDiff, &ySum) // Reuse invXDiff
	xQminusP.Square(&lambdaPrime)
	xQminusP.Sub(&xQminusP, &xSum)

	// Compute inverses for the ladder using Montgomery's batch inversion
	// Instead of 2 inversions, we use 1 inversion + 3 multiplications
	// Montgomery's trick: to compute 1/a and 1/b, compute ab, then 1/(ab),
	// then 1/a = b·(1/(ab)) and 1/b = a·(1/(ab))
	var iXPprime, iXQminusP, prod, invProd fptower.E12
	prod.Mul(&xPprime, &xQminusP)     // ab
	invProd.Inverse(&prod)            // 1/(ab)
	iXPprime.Mul(&xQminusP, &invProd) // 1/a = b·(1/(ab))
	iXQminusP.Mul(&xPprime, &invProd) // 1/b = a·(1/(ab))

	// Run the ladder using optimized sparse×dense multiplication
	// Precomputed values are in E2 (sparse), T values are dense in E12
	// Use 2-bit windowed processing for consecutive (0,0) bits
	for idx := 0; idx < len(LoopCounter)-1; {
		i := len(LoopCounter) - 2 - idx

		// Check if we can use 2-bit windowed processing
		if pre.TwoBitValid != nil && idx < len(pre.TwoBitValid) && pre.TwoBitValid[idx] {
			// Both this bit and next bit are 0, use optimized 2-bit cDIFF
			var newT cubicalPointE12
			cDIFF2_SparseQQ(&TX, &TZ, &pre.TwoBitCoeffs[idx], &iXPprime, &pre.FourBTwist, &newT)
			TX.Set(&newT.X)
			TZ.Set(&newT.Z)
			idx += 2 // Skip two bits
		} else if LoopCounter[i] == 0 {
			// Single bit=0: cDIFF(T, precomputed, iXPprime) where precomputed is sparse
			var newT cubicalPointE12
			cDIFFE12_SparseQ(&TX, &TZ, &pre.LadderX[idx], &pre.LadderZ[idx], &iXPprime, &pre.FourBTwist, &newT)
			TX.Set(&newT.X)
			TZ.Set(&newT.Z)
			idx++
		} else {
			// bit=1: cDIFF(precomputed, T, iXQminusP) where precomputed is sparse
			var newT cubicalPointE12
			cDIFFE12_SparseP(&pre.LadderX[idx], &pre.LadderZ[idx], &TX, &TZ, &iXQminusP, &pre.FourBTwist, &newT)
			TX.Set(&newT.X)
			TZ.Set(&newT.Z)
			idx++
		}
	}

	result.Set(&TZ)
	return result, nil
}

// PairCubicalFixedQ computes the cubical pairing e(P, Q)² using precomputed Q values.
func PairCubicalFixedQ(P *G1Affine, pre *G2CubicalPrecompute) (GT, error) {
	result, err := MillerLoopCubicalFixedQ(P, pre)
	if err != nil {
		return result, err
	}

	// Conjugate since x₀ is negative for BLS12-381
	result.Conjugate(&result)

	return FinalExponentiation(&result), nil
}

// PairCubicalFixedQMulti computes ∏ᵢ e(Pᵢ, Q)² for multiple P values with a fixed Q.
func PairCubicalFixedQMulti(P []G1Affine, pre *G2CubicalPrecompute) (GT, error) {
	var result GT
	result.SetOne()

	if pre == nil || len(P) == 0 {
		return result, nil
	}

	// Batch all Miller loops
	for k := 0; k < len(P); k++ {
		if P[k].IsInfinity() {
			continue
		}

		ml, err := MillerLoopCubicalFixedQ(&P[k], pre)
		if err != nil {
			return GT{}, err
		}

		result.Mul(&result, &ml)
	}

	// Conjugate since x₀ is negative for BLS12-381
	result.Conjugate(&result)

	// Single final exponentiation
	return FinalExponentiation(&result), nil
}
