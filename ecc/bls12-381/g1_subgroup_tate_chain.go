package bls12381

import "github.com/consensys/gnark-crypto/ecc/bls12-381/fp"

// IsInSubGroupTateChain checks whether p is in the correct subgroup using the
// Tate-based test with a chain-based Miller loop (no branching).
//
// It follows "Revisiting subgroup membership testing on pairing-friendly
// curves via the Tate pairing" by Y. Dai et al. with an addition-chain
// optimization that eliminates loop overhead and enables better pipelining.
func (p *G1Affine) IsInSubGroupTateChain(tab chainTable) bool {
	if p.IsInfinity() {
		return false
	}
	if !p.IsOnCurve() {
		return false
	}
	return membershipTestChain(p, &tab.q, tab.tab)
}

// membershipTestChain performs the Tate-based membership test using a
// chain-based Miller loop.
func membershipTestChain(p, q *G1Affine, tab []fp.Element) bool {
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

	// Simultaneous inversion and final exponentiations
	n1.Mul(&n1, &d2)
	n2.Mul(&n2, &d1)
	d1.Mul(&d1, &d2)
	d1.Inverse(&d1)
	n1.Mul(&n1, &d1)
	n2.Mul(&n2, &d1)

	return f1IsOne(&n1) && f2IsOne(&n2)
}

// sharedMillerloopChain is a straight-line implementation of the shared Miller
// loop for BLS12-381.
//
// Operation sequence (33 iterations total):
//  1. SQPL + SSUB (bits 63→62→61)
//  2. SQPL + SADD (bits 61→60→59)
//  3. SQPL (bits 59→58→57)
//  4. SDADD (bit 57→56)
//     5-8. 4× SQPL (bits 56→48)
//  9. SDADD (bit 48→47)
//     10-24. 15× SQPL (bits 47→17)
//  25. SQPL + SADD (bits 17→16→15)
//     26-33. 8× SQPL (bits 15→-1)
func sharedMillerloopChain(tab []fp.Element, n1, d1, n2, d2 *fp.Element, q, p, p2 *G1Affine) {
	var t0, t1, t2, t3, t4, t5, t6 fp.Element
	var v0, v1, v2 fp.Element

	// Precompute constants
	var u1, u2, u3 fp.Element
	u1.Sub(&p.Y, &q.Y)
	u2.Sub(&p.X, &q.X)
	u3.Sub(&p2.X, &q.X)

	k := 0

	// ========================================
	// 1. SQPL + SSUB (bits 63→62→61)
	// ========================================
	// SQPL part
	t0.Sub(&p.Y, &tab[k+2])
	t1.Sub(&p.X, &tab[k+1])
	t2.Sub(&p2.X, &tab[k+1])
	t3.Mul(&t1, &tab[k])
	t3.Sub(&t0, &t3)
	t4.Mul(&t2, &tab[k])
	t4.Sub(&t0, &t4)
	t5.Mul(&t1, &tab[k+3])
	t5.Sub(&t0, &t5)
	t6.Mul(&t2, &tab[k+3])
	t6.Sub(&t0, &t6)

	n1.Square(n1)
	n1.Square(n1)
	n1.Mul(n1, &t5)
	d1.Square(d1)
	d1.Mul(d1, &t3)
	d1.Square(d1)

	n2.Square(n2)
	n2.Square(n2)
	n2.Mul(n2, &t6)
	d2.Square(d2)
	d2.Mul(d2, &t4)
	d2.Square(d2)
	k += 4

	// SSUB part (naf[62] = -1)
	t1.Sub(&p.X, &tab[k+1])
	t2.Sub(&p2.X, &tab[k+1])
	t3.Mul(&tab[k], &u2)
	t3.Sub(&u1, &t3)
	t4.Mul(&tab[k], &u3)
	t4.Sub(&u1, &t4)

	n1.Mul(n1, &t1)
	d1.Mul(d1, &t3)
	n2.Mul(n2, &t2)
	d2.Mul(d2, &t4)
	k += 2

	// ========================================
	// 2. SQPL + SADD (bits 61→60→59)
	// ========================================
	// SQPL part
	t0.Sub(&p.Y, &tab[k+2])
	t1.Sub(&p.X, &tab[k+1])
	t2.Sub(&p2.X, &tab[k+1])
	t3.Mul(&t1, &tab[k])
	t3.Sub(&t0, &t3)
	t4.Mul(&t2, &tab[k])
	t4.Sub(&t0, &t4)
	t5.Mul(&t1, &tab[k+3])
	t5.Sub(&t0, &t5)
	t6.Mul(&t2, &tab[k+3])
	t6.Sub(&t0, &t6)

	n1.Square(n1)
	n1.Square(n1)
	n1.Mul(n1, &t5)
	d1.Square(d1)
	d1.Mul(d1, &t3)
	d1.Square(d1)

	n2.Square(n2)
	n2.Square(n2)
	n2.Mul(n2, &t6)
	d2.Square(d2)
	d2.Mul(d2, &t4)
	d2.Square(d2)
	k += 4

	// SADD part (naf[60] = 1)
	t1.Mul(&tab[k], &u2)
	t1.Sub(&u1, &t1)
	t2.Mul(&tab[k], &u3)
	t2.Sub(&u1, &t2)
	t3.Sub(&p.X, &tab[k+1])
	t4.Sub(&p2.X, &tab[k+1])

	n1.Mul(n1, &t1)
	d1.Mul(d1, &t3)
	n2.Mul(n2, &t2)
	d2.Mul(d2, &t4)
	k += 2

	// ========================================
	// 3. SQPL (bits 59→58→57), no add
	// ========================================
	t0.Sub(&p.Y, &tab[k+2])
	t1.Sub(&p.X, &tab[k+1])
	t2.Sub(&p2.X, &tab[k+1])
	t3.Mul(&t1, &tab[k])
	t3.Sub(&t0, &t3)
	t4.Mul(&t2, &tab[k])
	t4.Sub(&t0, &t4)
	t5.Mul(&t1, &tab[k+3])
	t5.Sub(&t0, &t5)
	t6.Mul(&t2, &tab[k+3])
	t6.Sub(&t0, &t6)

	n1.Square(n1)
	n1.Square(n1)
	n1.Mul(n1, &t5)
	d1.Square(d1)
	d1.Mul(d1, &t3)
	d1.Square(d1)

	n2.Square(n2)
	n2.Square(n2)
	n2.Mul(n2, &t6)
	d2.Square(d2)
	d2.Mul(d2, &t4)
	d2.Square(d2)
	k += 4

	// ========================================
	// 4. SDADD (bit 57, naf[57] = 1)
	// ========================================
	t0.Sub(&p.X, &tab[k])
	t1.Sub(&p2.X, &tab[k])
	t3.Add(&p.X, &tab[k+2])
	t4.Add(&p2.X, &tab[k+2])
	v0.Mul(&t0, &t3)
	v1.Mul(&t1, &t4)
	t3.Sub(&p.Y, &tab[k+1])
	v2.Mul(&t3, &tab[k+3])
	v0.Sub(&v0, &v2)
	v1.Sub(&v1, &v2)

	n1.Square(n1)
	n1.Mul(n1, &v0)
	d1.Mul(d1, &t0)
	d1.Square(d1)

	n2.Square(n2)
	n2.Mul(n2, &v1)
	d2.Mul(d2, &t1)
	d2.Square(d2)
	k += 4

	// ========================================
	// 5-8. 4× SQPL (bits 56→48)
	// ========================================
	for i := 0; i < 4; i++ {
		t0.Sub(&p.Y, &tab[k+2])
		t1.Sub(&p.X, &tab[k+1])
		t2.Sub(&p2.X, &tab[k+1])
		t3.Mul(&t1, &tab[k])
		t3.Sub(&t0, &t3)
		t4.Mul(&t2, &tab[k])
		t4.Sub(&t0, &t4)
		t5.Mul(&t1, &tab[k+3])
		t5.Sub(&t0, &t5)
		t6.Mul(&t2, &tab[k+3])
		t6.Sub(&t0, &t6)

		n1.Square(n1)
		n1.Square(n1)
		n1.Mul(n1, &t5)
		d1.Square(d1)
		d1.Mul(d1, &t3)
		d1.Square(d1)

		n2.Square(n2)
		n2.Square(n2)
		n2.Mul(n2, &t6)
		d2.Square(d2)
		d2.Mul(d2, &t4)
		d2.Square(d2)
		k += 4
	}

	// ========================================
	// 9. SDADD (bit 48, naf[48] = 1)
	// ========================================
	t0.Sub(&p.X, &tab[k])
	t1.Sub(&p2.X, &tab[k])
	t3.Add(&p.X, &tab[k+2])
	t4.Add(&p2.X, &tab[k+2])
	v0.Mul(&t0, &t3)
	v1.Mul(&t1, &t4)
	t3.Sub(&p.Y, &tab[k+1])
	v2.Mul(&t3, &tab[k+3])
	v0.Sub(&v0, &v2)
	v1.Sub(&v1, &v2)

	n1.Square(n1)
	n1.Mul(n1, &v0)
	d1.Mul(d1, &t0)
	d1.Square(d1)

	n2.Square(n2)
	n2.Mul(n2, &v1)
	d2.Mul(d2, &t1)
	d2.Square(d2)
	k += 4

	// ========================================
	// 10-24. 15× SQPL (bits 47→17)
	// ========================================
	for i := 0; i < 15; i++ {
		t0.Sub(&p.Y, &tab[k+2])
		t1.Sub(&p.X, &tab[k+1])
		t2.Sub(&p2.X, &tab[k+1])
		t3.Mul(&t1, &tab[k])
		t3.Sub(&t0, &t3)
		t4.Mul(&t2, &tab[k])
		t4.Sub(&t0, &t4)
		t5.Mul(&t1, &tab[k+3])
		t5.Sub(&t0, &t5)
		t6.Mul(&t2, &tab[k+3])
		t6.Sub(&t0, &t6)

		n1.Square(n1)
		n1.Square(n1)
		n1.Mul(n1, &t5)
		d1.Square(d1)
		d1.Mul(d1, &t3)
		d1.Square(d1)

		n2.Square(n2)
		n2.Square(n2)
		n2.Mul(n2, &t6)
		d2.Square(d2)
		d2.Mul(d2, &t4)
		d2.Square(d2)
		k += 4
	}

	// ========================================
	// 25. SQPL + SADD (bits 17→16→15)
	// ========================================
	// SQPL part
	t0.Sub(&p.Y, &tab[k+2])
	t1.Sub(&p.X, &tab[k+1])
	t2.Sub(&p2.X, &tab[k+1])
	t3.Mul(&t1, &tab[k])
	t3.Sub(&t0, &t3)
	t4.Mul(&t2, &tab[k])
	t4.Sub(&t0, &t4)
	t5.Mul(&t1, &tab[k+3])
	t5.Sub(&t0, &t5)
	t6.Mul(&t2, &tab[k+3])
	t6.Sub(&t0, &t6)

	n1.Square(n1)
	n1.Square(n1)
	n1.Mul(n1, &t5)
	d1.Square(d1)
	d1.Mul(d1, &t3)
	d1.Square(d1)

	n2.Square(n2)
	n2.Square(n2)
	n2.Mul(n2, &t6)
	d2.Square(d2)
	d2.Mul(d2, &t4)
	d2.Square(d2)
	k += 4

	// SADD part (naf[16] = 1)
	t1.Mul(&tab[k], &u2)
	t1.Sub(&u1, &t1)
	t2.Mul(&tab[k], &u3)
	t2.Sub(&u1, &t2)
	t3.Sub(&p.X, &tab[k+1])
	t4.Sub(&p2.X, &tab[k+1])

	n1.Mul(n1, &t1)
	d1.Mul(d1, &t3)
	n2.Mul(n2, &t2)
	d2.Mul(d2, &t4)
	k += 2

	// ========================================
	// 26-33. 8× SQPL (bits 15→-1)
	// ========================================
	for i := 0; i < 8; i++ {
		t0.Sub(&p.Y, &tab[k+2])
		t1.Sub(&p.X, &tab[k+1])
		t2.Sub(&p2.X, &tab[k+1])
		t3.Mul(&t1, &tab[k])
		t3.Sub(&t0, &t3)
		t4.Mul(&t2, &tab[k])
		t4.Sub(&t0, &t4)
		t5.Mul(&t1, &tab[k+3])
		t5.Sub(&t0, &t5)
		t6.Mul(&t2, &tab[k+3])
		t6.Sub(&t0, &t6)

		n1.Square(n1)
		n1.Square(n1)
		n1.Mul(n1, &t5)
		d1.Square(d1)
		d1.Mul(d1, &t3)
		d1.Square(d1)

		n2.Square(n2)
		n2.Square(n2)
		n2.Mul(n2, &t6)
		d2.Square(d2)
		d2.Mul(d2, &t4)
		d2.Square(d2)
		k += 4
	}
}
