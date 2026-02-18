// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fptower

import (
	"math/bits"

	"github.com/consensys/gnark-crypto/ecc/bn254/fp"
)

// declaring nonResInverse as global makes MulByNonResInv inlinable
var nonResInverse E2 = E2{
	A0: fp.Element{
		10477841894441615122,
		7327163185667482322,
		3635199979766503006,
		3215324977242306624,
	},
	A1: fp.Element{
		7515750141297360845,
		14746352163864140223,
		11319968037783994424,
		30185921062296004,
	},
}

// mulGenericE2 sets z to the E2-product of x,y, returns z
// note: do not rename, this is referenced in the x86 assembly impl
func mulGenericE2(z, x, y *E2) {
	var a, b, c fp.Element
	a.Add(&x.A0, &x.A1)
	b.Add(&y.A0, &y.A1)
	a.Mul(&a, &b)
	b.Mul(&x.A0, &y.A0)
	c.Mul(&x.A1, &y.A1)
	z.A1.Sub(&a, &b).Sub(&z.A1, &c)
	z.A0.Sub(&b, &c) // z.A0.MulByNonResidue(&c).Add(&z.A0, &b)
}

// squareGenericE2 sets z to the E2-product of x,x returns z
// note: do not rename, this is referenced in the x86 assembly impl
func squareGenericE2(z, x *E2) {
	// adapted from algo 22 https://eprint.iacr.org/2010/354.pdf
	var a, b fp.Element
	a.Add(&x.A0, &x.A1)
	b.Sub(&x.A0, &x.A1)
	a.Mul(&a, &b)
	b.Mul(&x.A0, &x.A1).Double(&b)
	z.A0.Set(&a)
	z.A1.Set(&b)
}

// MulByNonResidueInv multiplies a E2 by (9,1)^{-1}
func (z *E2) MulByNonResidueInv(x *E2) *E2 {
	z.Mul(x, &nonResInverse)
	return z
}

// Inverse sets z to the E2-inverse of x, returns z
//
// if x == 0, sets and returns z = x
func (z *E2) Inverse(x *E2) *E2 {
	// Algorithm 8 from https://eprint.iacr.org/2010/354.pdf
	var t0, t1 fp.Element
	t0.Square(&x.A0)
	t1.Square(&x.A1)
	t0.Add(&t0, &t1)
	t1.Inverse(&t0)
	z.A0.Mul(&x.A0, &t1)
	z.A1.Mul(&x.A1, &t1).Neg(&z.A1)

	return z
}

// norm sets x to the norm of z
func (z *E2) norm(x *fp.Element) {
	var tmp fp.Element
	x.Square(&z.A0)
	tmp.Square(&z.A1)
	x.Add(x, &tmp)
}

// MulBybTwistCurveCoeff multiplies by 3/(9,1)
func (z *E2) MulBybTwistCurveCoeff(x *E2) *E2 {

	var res E2
	res.MulByNonResidueInv(x)
	z.Double(&res).
		Add(&res, z)

	return z
}

// lucasExponent is the bit decomposition of e = 3⁻¹ mod (p+1),
// excluding the leading 1-bit and the trailing bit (which is 1).
// e = 7296080957279758407415468581752425029565437052432607887563012631548408736195
// 253 bits total; stored MSB-first, 251 inner bits.
var lucasExponent = [251]uint8{
	0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 1, 1, 0, 1, 1, 1, 1, 0, 1, 1, 1, 1, 0, 1,
	1, 1, 0, 1, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 1, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 1, 1, 0, 1, 1, 1, 1, 0, 1, 0, 0, 0, 0, 0, 0, 1, 1, 0, 1, 0, 1, 1, 0, 0, 0,
	0, 0, 1, 1, 1, 1, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0,
	0, 1, 1, 1, 0, 0, 1, 0, 1, 1, 0, 0, 1, 0, 0, 1, 1, 1, 0, 1, 1, 1, 0, 1, 0, 0, 1,
	0, 1, 0, 1, 1, 0, 0, 1, 0, 0, 0, 1, 1, 1, 0, 0, 0, 0, 1, 0, 1, 1, 1, 0, 0, 1, 1,
	0, 1, 0, 1, 1, 1, 1, 0, 1, 1, 0, 1, 0, 0, 0, 0, 1, 1, 1, 0, 0, 0, 0, 1, 0, 0, 0,
	1, 1, 0, 1, 0, 0, 1, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 1, 1, 1, 0, 1, 0, 1, 1,
	0, 0, 1, 0, 0, 1, 0, 0, 1, 0, 0, 0, 0, 0, 1, 0, 1, 0, 0, 1, 1, 0, 1, 0, 1, 0, 0,
	1, 1, 1, 0, 0, 0, 0, 1,
}

// cbrtHybrid computes the cube root of x in E2 using the algebraic torus T₂(Fp).
//
// Let r = z^{p-1} where z³ = x. Then r³ = x^{p-1} lies on T₂(Fp) (norm 1).
// The trace of r on T₂ is s₁ = V_e(α_t, 1) where:
//   - α_t = 2·(x₀² - x₁²)/N(x) is the trace of x^{p-1} on T₂
//   - e = 3⁻¹ mod (p+1)
//
// Recovery: z₀ = x₀/(m·(s₁-1)), z₁ = x₁/(m·(s₁+1)), where m = cbrt(N(x)).
func (z *E2) cbrtHybrid(x *E2) *E2 {
	if x.A1.IsZero() {
		if z.A0.Cbrt(&x.A0) == nil {
			return nil
		}
		z.A1.SetZero()
		return z
	}

	if x.A0.IsZero() {
		var negA1 fp.Element
		negA1.Neg(&x.A1)
		var y E2
		if y.A1.Cbrt(&negA1) == nil {
			return nil
		}
		y.A0.SetZero()
		return z.cbrtVerifyAndAdjust(x, &y)
	}

	// N = x₀² + x₁² (norm of x)
	var norm fp.Element
	x.norm(&norm)

	// m = cbrt(N) in Fp
	var m fp.Element
	if m.Cbrt(&norm) == nil {
		return nil
	}

	// α_t = 2·(x₀² - x₁²)/N = trace of x^{p-1} on T₂
	var x0sq, x1sq, alphaT fp.Element
	x0sq.Square(&x.A0)
	x1sq.Square(&x.A1)
	alphaT.Sub(&x0sq, &x1sq)
	alphaT.Double(&alphaT)
	var normInv fp.Element
	normInv.Inverse(&norm)
	alphaT.Mul(&alphaT, &normInv)

	// s₁ = V_e(α_t, 1) where e = 3⁻¹ mod (p+1), Q = 1
	sp := lucasV(&alphaT)

	// Recovery: z₀ = x₀/(m·(s₁ - 1)), z₁ = x₁/(m·(s₁ + 1))
	// Use a single inversion via Montgomery's trick: 1/(a·b) then multiply out.
	var z0, z1 fp.Element
	var one, s1m1, s1p1, d0, d1, d0d1, d0d1Inv fp.Element

	one.SetOne()
	s1m1.Sub(&sp, &one)
	s1p1.Add(&sp, &one)
	d0.Mul(&m, &s1m1) // m·(s₁-1)
	d1.Mul(&m, &s1p1) // m·(s₁+1)

	// single inversion: 1/(d0·d1)
	d0d1.Mul(&d0, &d1)
	d0d1Inv.Inverse(&d0d1)

	// 1/d0 = d1 · 1/(d0·d1), 1/d1 = d0 · 1/(d0·d1)
	z0.Mul(&d1, &d0d1Inv)
	z0.Mul(&z0, &x.A0) // x₀/d0
	z1.Mul(&d0, &d0d1Inv)
	z1.Mul(&z1, &x.A1) // x₁/d1

	var y E2
	y.A0.Set(&z0)
	y.A1.Set(&z1)

	return z.cbrtVerifyAndAdjust(x, &y)
}

// lucasVChainQ computes V_e(alpha) with product parameter Q using a
// Montgomery ladder for generalized Lucas V-sequences.
// The recurrence is V_{n+1} = alpha·V_n - Q·V_{n-1}, V_0 = 2, V_1 = alpha.
// e is given as little-endian uint64 limbs.
//
// We maintain (V_k, V_{k+1}, Q^k). For each bit of e from MSB-1 down to 0:
//
//	bit=0: (V_{2k}, V_{2k+1}, Q^{2k})
//	  V_{2k}   = V_k² - 2·Q^k
//	  V_{2k+1} = V_k·V_{k+1} - alpha·Q^k
//	  Q^{2k}   = (Q^k)²
//
//	bit=1: (V_{2k+1}, V_{2k+2}, Q^{2k+1})
//	  V_{2k+1} = V_k·V_{k+1} - alpha·Q^k
//	  V_{2k+2} = V_{k+1}² - 2·Q^{k+1}
//	  Q^{2k+1} = (Q^k)²·Q
func lucasVChainQ(e [4]uint64, alpha, Q *fp.Element) fp.Element {
	// Find MSB
	totalBits := 0
	for i := 3; i >= 0; i-- {
		if e[i] != 0 {
			totalBits = i*64 + 64 - bits.LeadingZeros64(e[i])
			break
		}
	}
	if totalBits == 0 {
		var two fp.Element
		two.SetUint64(2)
		return two
	}

	// Initialize for MSB=1: (V_1, V_2, Q^1)
	var v0, v1, qk fp.Element
	v0.Set(alpha) // V_1 = alpha
	v1.Square(alpha)
	var twoQ fp.Element
	twoQ.Double(Q)
	v1.Sub(&v1, &twoQ) // V_2 = alpha² - 2Q
	qk.Set(Q)          // Q^1

	var prod, aqk, qk1 fp.Element

	for i := totalBits - 2; i >= 0; i-- {
		bit := (e[i/64] >> uint(i%64)) & 1

		// common: prod = V_k · V_{k+1} - alpha · Q^k
		prod.Mul(&v0, &v1)
		aqk.Mul(alpha, &qk)
		prod.Sub(&prod, &aqk)

		if bit == 0 {
			v1.Set(&prod)
			// v0 = V_k² - 2·Q^k
			aqk.Double(&qk) // reuse aqk as tmp
			v0.Square(&v0)
			v0.Sub(&v0, &aqk)
			qk.Square(&qk)
		} else {
			v0.Set(&prod)
			// v1 = V_{k+1}² - 2·Q^{k+1}
			qk1.Mul(&qk, Q)  // Q^{k+1}
			aqk.Double(&qk1) // 2·Q^{k+1}
			v1.Square(&v1)
			v1.Sub(&v1, &aqk)
			// Q^{2k+1} = (Q^k)² · Q
			qk.Square(&qk)
			qk.Mul(&qk, Q)
		}
	}

	return v0
}

// lucasV computes V_e(alpha, 1) where e = 3⁻¹ mod (p+1), using the
// Lucas V-sequence with Q=1 and a Montgomery ladder on precomputed bits.
//
// Since Q=1, Q^k=1 for all k, so we don't need to track it.
// Recurrence: V_{n+1} = alpha·V_n - V_{n-1}, V_0 = 2, V_1 = alpha.
//
// Per-bit step (maintaining V_k, V_{k+1}):
//
//	prod    = V_k·V_{k+1} - alpha
//	bit=0: V_{2k}   = V_k² - 2,      V_{2k+1} = prod
//	bit=1: V_{2k+1} = prod,           V_{2k+2} = V_{k+1}² - 2
func lucasV(alpha *fp.Element) fp.Element {
	// Initialize for MSB=1: V_1 = alpha, V_2 = alpha² - 2
	var v0, v1, two fp.Element
	two.SetUint64(2)
	v0.Set(alpha)
	v1.Square(alpha).Sub(&v1, &two)

	var prod fp.Element

	// Process 251 precomputed bits (MSB-first, excluding leading 1 and trailing bit)
	for i := 0; i < 251; i++ {
		// prod = V_k · V_{k+1} - alpha
		prod.Mul(&v0, &v1).Sub(&prod, alpha)

		if lucasExponent[i] == 0 {
			v1.Set(&prod)
			v0.Square(&v0).Sub(&v0, &two)
		} else {
			v0.Set(&prod)
			v1.Square(&v1).Sub(&v1, &two)
		}
	}

	// Last bit of exponent is 1: only compute v0 = v0·v1 - alpha
	v0.Mul(&v0, &v1).Sub(&v0, alpha)

	return v0
}
