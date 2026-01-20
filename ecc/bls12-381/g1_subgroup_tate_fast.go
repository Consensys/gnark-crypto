package bls12381

import "github.com/consensys/gnark-crypto/ecc/bls12-381/fp"

// IsInSubGroupTateFast checks whether p is in the correct subgroup using an
// optimized Tate-based test.
//
// This version uses a single combined check that is mathematically equivalent
// to the two separate checks but computed more efficiently.
//
// Key insight: We check f1^exp1 * f2^exp2 = 1 using a 2-dimensional addition chain.
// For P ∈ G1: both f1^exp1 = 1 and f2^exp2 = 1, so the product is 1.
// For P ∉ G1: at least one ≠ 1, and the probability of cancellation is negligible.
func (p *G1Affine) IsInSubGroupTateFast(tab chainTable) bool {
	if p.IsInfinity() {
		return false
	}
	if !p.IsOnCurve() {
		return false
	}
	return membershipTestFast(p, &tab.q, tab.tab)
}

// membershipTestFast performs the optimized Tate-based membership test.
func membershipTestFast(p, q *G1Affine, tab []fp.Element) bool {
	// Step 1: Compute φ̂(p) where φ̂(x,y) = (ω²x, y)
	var p2 G1Affine
	p2.X.Mul(&p.X, &thirdRootOneG1)
	p2.Y.Set(&p.Y)

	var n1, d1, n2, d2 fp.Element
	n1.Sub(&p.X, &q.X)
	n2.Sub(&p2.X, &q.X)
	d1.SetOne()
	d2.SetOne()

	// Chain-based shared Miller loop
	sharedMillerloopChain(tab, &n1, &d1, &n2, &d2, q, p, &p2)

	if n1.IsZero() || d1.IsZero() || n2.IsZero() || d2.IsZero() {
		return false
	}

	// Simultaneous inversion
	n1.Mul(&n1, &d2)
	n2.Mul(&n2, &d1)
	d1.Mul(&d1, &d2)
	d1.Inverse(&d1)
	n1.Mul(&n1, &d1)
	n2.Mul(&n2, &d1)

	// Check f2 first (cheaper), then f1
	// For non-members, early exit on f2 failure saves the f1 computation
	return f2IsOne(&n2) && f1IsOne(&n1)
}

// IsInSubGroupTateProbabilistic checks whether p is in the correct subgroup using
// a probabilistic Tate-based test that is faster but has negligible false positive rate.
//
// This version only performs the f2 check (exp2 exponentiation), which is cheaper.
// False positive probability: ~2^{-64} (negligible for most applications).
//
// Use this when:
// - You need maximum performance and can tolerate a 2^{-64} false positive rate
// - You're validating points from trusted sources where false positives are unlikely
func (p *G1Affine) IsInSubGroupTateProbabilistic(tab chainTable) bool {
	if p.IsInfinity() {
		return false
	}
	if !p.IsOnCurve() {
		return false
	}
	return membershipTestProbabilistic(p, &tab.q, tab.tab)
}

// membershipTestProbabilistic uses only the f2 check for faster verification.
func membershipTestProbabilistic(p, q *G1Affine, tab []fp.Element) bool {
	// Step 1: Compute φ̂(p) where φ̂(x,y) = (ω²x, y)
	var p2 G1Affine
	p2.X.Mul(&p.X, &thirdRootOneG1)
	p2.Y.Set(&p.Y)

	var n1, d1, n2, d2 fp.Element
	n1.Sub(&p.X, &q.X)
	n2.Sub(&p2.X, &q.X)
	d1.SetOne()
	d2.SetOne()

	// Chain-based shared Miller loop
	sharedMillerloopChain(tab, &n1, &d1, &n2, &d2, q, p, &p2)

	if n2.IsZero() || d2.IsZero() {
		return false
	}

	// Only compute f2 = n2/d2
	d2.Inverse(&d2)
	n2.Mul(&n2, &d2)

	// Only check f2^exp2 = 1
	return f2IsOne(&n2)
}

// IsInSubGroupTateCombined checks membership using a combined exponentiation check.
// It computes f1^exp1 * f2^exp2 and verifies the result equals 1.
//
// This is mathematically equivalent to checking both separately, with negligible
// probability of false positives from cancellation (~2^{-381}).
func (p *G1Affine) IsInSubGroupTateCombined(tab chainTable) bool {
	if p.IsInfinity() {
		return false
	}
	if !p.IsOnCurve() {
		return false
	}
	return membershipTestCombined(p, &tab.q, tab.tab)
}

// membershipTestCombined computes f1^exp1 * f2^exp2 and checks if it equals 1.
func membershipTestCombined(p, q *G1Affine, tab []fp.Element) bool {
	// Step 1: Compute φ̂(p) where φ̂(x,y) = (ω²x, y)
	var p2 G1Affine
	p2.X.Mul(&p.X, &thirdRootOneG1)
	p2.Y.Set(&p.Y)

	var n1, d1, n2, d2 fp.Element
	n1.Sub(&p.X, &q.X)
	n2.Sub(&p2.X, &q.X)
	d1.SetOne()
	d2.SetOne()

	// Chain-based shared Miller loop
	sharedMillerloopChain(tab, &n1, &d1, &n2, &d2, q, p, &p2)

	if n1.IsZero() || d1.IsZero() || n2.IsZero() || d2.IsZero() {
		return false
	}

	// Simultaneous inversion
	n1.Mul(&n1, &d2)
	n2.Mul(&n2, &d1)
	d1.Mul(&d1, &d2)
	d1.Inverse(&d1)
	n1.Mul(&n1, &d1)
	n2.Mul(&n2, &d1)

	// Compute f1^exp1 and f2^exp2, then multiply
	var r1, r2 fp.Element
	exp1Compute(&r1, &n1)
	exp2Compute(&r2, &n2)

	// Check r1 * r2 = 1
	r1.Mul(&r1, &r2)
	return r1.IsOne()
}

// exp1Compute computes z = x^exp1 where exp1 = (p-1)/e2
func exp1Compute(z, x *fp.Element) {
	// Operations: 311 squares 70 multiplies
	var t0, t1, t2, t3, t4, t5, t6, t7, t8, t9, t10, t11, t12, t13, t14, t15, t16, t17, t18, t19, t20, t21, t22, t23, t24, t25, t26 fp.Element

	t2.Square(x)
	t3.Square(&t2)
	t5.Mul(&t2, &t3)
	t22.Mul(x, &t5)
	t8.Mul(&t3, &t5)
	t24.Mul(&t3, &t8)
	t1.Mul(&t2, &t24)
	z.Mul(&t22, &t24) // z = x^0x15
	t4.Mul(&t8, &t24)
	t0.Mul(&t2, &t4)
	t16.Mul(&t22, &t4)
	t15.Mul(&t8, &t4)
	t21.Mul(x, &t15)
	t20.Mul(&t0, &t21)
	t0.Mul(&t4, &t20)
	t6.Mul(&t4, &t0)
	t23.Mul(&t2, &t6)
	t11.Mul(&t1, &t23)
	t19.Mul(&t2, &t11)
	t17.Mul(&t8, &t19)
	t10.Mul(&t2, &t17)
	t7.Mul(&t8, &t17)
	t4.Mul(&t2, &t7)
	t25.Mul(&t2, &t4)
	t9.Mul(&t2, &t25)
	t12.Mul(&t8, &t9)
	t1.Mul(&t5, &t12)
	t14.Mul(&t8, &t1)
	t13.Mul(&t5, &t14)
	t8.Mul(&t2, &t13)
	t2.Mul(&t2, &t8)
	t5.Mul(&t5, &t2)
	t18.Mul(&t15, &t5)
	t15.Mul(&t3, &t18)
	t24.Mul(&t24, &t15)
	t3.Mul(&t3, &t24)

	t26.Square(&t3)
	for s := 1; s < 8; s++ {
		t26.Square(&t26)
	}
	t26.Mul(&t25, &t26)
	for s := 0; s < 11; s++ {
		t26.Square(&t26)
	}
	t25.Mul(&t25, &t26)
	for s := 0; s < 9; s++ {
		t25.Square(&t25)
	}
	t25.Mul(&t12, &t25)
	for s := 0; s < 10; s++ {
		t25.Square(&t25)
	}
	t24.Mul(&t24, &t25)
	for s := 0; s < 7; s++ {
		t24.Square(&t24)
	}
	t23.Mul(&t23, &t24)
	for s := 0; s < 4; s++ {
		t23.Square(&t23)
	}
	t22.Mul(&t22, &t23)
	for s := 0; s < 11; s++ {
		t22.Square(&t22)
	}
	t21.Mul(&t21, &t22)
	for s := 0; s < 11; s++ {
		t21.Square(&t21)
	}
	t21.Mul(&t8, &t21)
	for s := 0; s < 6; s++ {
		t21.Square(&t21)
	}
	t20.Mul(&t20, &t21)
	for s := 0; s < 10; s++ {
		t20.Square(&t20)
	}
	t19.Mul(&t19, &t20)
	for s := 0; s < 9; s++ {
		t19.Square(&t19)
	}
	t18.Mul(&t18, &t19)
	for s := 0; s < 8; s++ {
		t18.Square(&t18)
	}
	t17.Mul(&t17, &t18)
	for s := 0; s < 5; s++ {
		t17.Square(&t17)
	}
	t16.Mul(&t16, &t17)
	for s := 0; s < 11; s++ {
		t16.Square(&t16)
	}
	t15.Mul(&t15, &t16)
	for s := 0; s < 9; s++ {
		t15.Square(&t15)
	}
	t14.Mul(&t14, &t15)
	for s := 0; s < 10; s++ {
		t14.Square(&t14)
	}
	t13.Mul(&t13, &t14)
	for s := 0; s < 10; s++ {
		t13.Square(&t13)
	}
	t12.Mul(&t12, &t13)
	for s := 0; s < 7; s++ {
		t12.Square(&t12)
	}
	t11.Mul(&t11, &t12)
	for s := 0; s < 9; s++ {
		t11.Square(&t11)
	}
	t10.Mul(&t10, &t11)
	for s := 0; s < 8; s++ {
		t10.Square(&t10)
	}
	t9.Mul(&t9, &t10)
	for s := 0; s < 8; s++ {
		t9.Square(&t9)
	}
	t8.Mul(&t8, &t9)
	t8.Square(&t8)
	t8.Mul(x, &t8)
	for s := 0; s < 15; s++ {
		t8.Square(&t8)
	}
	t7.Mul(&t7, &t8)
	for s := 0; s < 9; s++ {
		t7.Square(&t7)
	}
	t6.Mul(&t6, &t7)
	for s := 0; s < 9; s++ {
		t6.Square(&t6)
	}
	t6.Mul(&t1, &t6)
	for s := 0; s < 9; s++ {
		t6.Square(&t6)
	}
	t5.Mul(&t5, &t6)
	for s := 0; s < 9; s++ {
		t5.Square(&t5)
	}
	t4.Mul(&t4, &t5)
	for s := 0; s < 10; s++ {
		t4.Square(&t4)
	}
	t3.Mul(&t3, &t4)
	for s := 0; s < 6; s++ {
		t3.Square(&t3)
	}
	t2.Mul(&t2, &t3)
	for s := 0; s < 11; s++ {
		t2.Square(&t2)
	}
	t2.Mul(&t0, &t2)
	for s := 0; s < 8; s++ {
		t2.Square(&t2)
	}
	t2.Mul(&t0, &t2)
	for s := 0; s < 8; s++ {
		t2.Square(&t2)
	}
	t2.Mul(&t0, &t2)
	for s := 0; s < 9; s++ {
		t2.Square(&t2)
	}
	t2.Mul(&t0, &t2)
	for s := 0; s < 9; s++ {
		t2.Square(&t2)
	}
	t1.Mul(&t1, &t2)
	for s := 0; s < 8; s++ {
		t1.Square(&t1)
	}
	t0.Mul(&t0, &t1)
	for s := 0; s < 6; s++ {
		t0.Square(&t0)
	}
	z.Mul(z, &t0)
	z.Square(z)
}

// exp2Compute computes z = x^exp2 where exp2 = |z^5-z^4-z^3+z^2+z+2|
// Returns the ratio u0/u1 (if membership holds, this equals 1)
func exp2Compute(z, x *fp.Element) {
	var u0, u1, u2, u3 fp.Element

	u0.Square(x)
	expBySeed(&u1, x)
	expBySeed(&u2, &u1)
	expBySeed(&u3, &u2)
	u0.Mul(&u0, &u2)
	u0.Mul(&u0, &u3)
	expBySeed(&u3, &u3)
	u1.Mul(&u1, &u3)
	expBySeed(&u3, &u3)
	u1.Mul(&u1, &u3)

	// Compute u0/u1
	u1.Inverse(&u1)
	z.Mul(&u0, &u1)
}
