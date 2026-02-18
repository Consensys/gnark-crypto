// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fptower

import (
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

// lucasExponent is e = 3⁻¹ mod (p+1) as little-endian uint64 limbs,
// used by the Lucas V-chain in cbrtHybrid.
// e = 7296080957279758407415468581752425029565437052432607887563012631548408736195
var lucasExponent = [4]uint64{
	7593120314996402627,
	15936870763965662084,
	16724893366231265993,
	1162332755600990221,
}

// cbrtAndNormInverse computes m = cbrt(norm) and normInv = 1/norm from a
// single shared exponentiation, avoiding a separate Fp inversion.
//
// Let t = norm^((q-19)/27). Then:
//   - m = norm · t  = norm^((q+8)/27)          (cube root candidate)
//   - normInv = t^27 · norm^17 = norm^(q-2)    (Fermat inverse)
//
// The ζ-adjustment for the cube root is identical to fp.Cbrt.
func cbrtAndNormInverse(norm *fp.Element) (m, normInv fp.Element, ok bool) {
	// t = norm^((q-19)/27)
	var t fp.Element
	t.ExpByCbrtHelperQMinus19Div27(*norm)

	// m = norm · t = norm^((q+8)/27)
	m.Mul(norm, &t)

	// Verify m³ = norm, adjust by ζ if needed (same logic as fp.Cbrt)
	var c fp.Element
	c.Cube(&m)
	if !c.Equal(norm) {
		// Precomputed constants (same as in fp.Cbrt)
		var zeta = fp.Element{
			9092840637269024442,
			11284133545212953584,
			7919372827184455520,
			1596114425137527684,
		}
		var zeta2 = fp.Element{
			1735008219140503419,
			10465829585049341007,
			6017168831245289042,
			1570250484855163800,
		}
		var omega = fp.Element{
			8183898218631979349,
			12014359695528440611,
			12263358156045030468,
			3187210487005268291,
		}
		var omega2 = fp.Element{
			3697675806616062876,
			9065277094688085689,
			6918009208039626314,
			2775033306905974752,
		}

		var cw2 fp.Element
		cw2.Mul(&c, &omega2)
		if cw2.Equal(norm) {
			m.Mul(&m, &zeta)
		} else {
			var cw fp.Element
			cw.Mul(&c, &omega)
			if cw.Equal(norm) {
				m.Mul(&m, &zeta2)
			} else {
				return m, normInv, false
			}
		}
	}

	// normInv = t^27 · norm^17 = norm^(q-2)
	// t^27: t² → t³ → (t³)² = t⁶ → t⁶·t³ = t⁹ → (t⁹)² = t¹⁸ → t¹⁸·t⁹ = t²⁷
	var t2, t3, t9 fp.Element
	t2.Square(&t)
	t3.Mul(&t2, &t)
	t9.Square(&t3)   // t⁶
	t9.Mul(&t9, &t3) // t⁹
	t2.Square(&t9)   // t¹⁸
	t2.Mul(&t2, &t9) // t²⁷

	// norm^17: norm² → norm⁴ → norm⁸ → norm¹⁶ → norm¹⁶·norm = norm¹⁷
	var n2, n4, n8, n16 fp.Element
	n2.Square(norm)
	n4.Square(&n2)
	n8.Square(&n4)
	n16.Square(&n8)
	normInv.Mul(&n16, norm)    // norm^17
	normInv.Mul(&normInv, &t2) // t^27 · norm^17 = norm^(q-2)

	return m, normInv, true
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

	// x₀², x₁² — reused for both norm and α_t
	var x0sq, x1sq fp.Element
	x0sq.Square(&x.A0)
	x1sq.Square(&x.A1)

	// N = x₀² + x₁² (norm of x)
	var norm fp.Element
	norm.Add(&x0sq, &x1sq)

	// m = cbrt(N) and normInv = 1/N from shared exponentiation
	m, normInv, ok := cbrtAndNormInverse(&norm)
	if !ok {
		return nil
	}

	// α_t = 2·(x₀² - x₁²)/N = trace of x^{p-1} on T₂
	var alphaT fp.Element
	alphaT.Sub(&x0sq, &x1sq)
	alphaT.Double(&alphaT)
	alphaT.Mul(&alphaT, &normInv)

	// s₁ = V_e(α_t, 1) where e = 3⁻¹ mod (p+1), Q = 1
	sp := lucasV(&alphaT)

	// Recovery: z₀ = x₀/(m·(s₁-1)), z₁ = x₁/(m·(s₁+1))
	// Use a single inversion via Montgomery's trick: 1/(a·b) then multiply out.
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
	var y E2
	y.A0.Mul(&d1, &d0d1Inv).Mul(&y.A0, &x.A0) // x₀/d0
	y.A1.Mul(&d0, &d0d1Inv).Mul(&y.A1, &x.A1) // x₁/d1

	return z.cbrtVerifyAndAdjust(x, &y)
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

	// Process bits 251 down to 1 (bit 252 is the leading 1, bit 0 handled separately)
	for i := 251; i >= 1; i-- {
		bit := (lucasExponent[i/64] >> uint(i%64)) & 1

		// prod = V_k · V_{k+1} - alpha
		prod.Mul(&v0, &v1).Sub(&prod, alpha)

		if bit == 0 {
			v1.Set(&prod)
			v0.Square(&v0).Sub(&v0, &two)
		} else {
			v0.Set(&prod)
			v1.Square(&v1).Sub(&v1, &two)
		}
	}

	// Last bit (bit 0) is 1: only compute v0, skip unused v1
	v0.Mul(&v0, &v1).Sub(&v0, alpha)

	return v0
}
