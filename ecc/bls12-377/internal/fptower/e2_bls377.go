// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fptower

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fp"
)

// Mul sets z to the E2-product of x,y, returns z
func (z *E2) Mul(x, y *E2) *E2 {
	var a, b, c fp.Element
	a.Add(&x.A0, &x.A1)
	b.Add(&y.A0, &y.A1)
	a.Mul(&a, &b)
	b.Mul(&x.A0, &y.A0)
	c.Mul(&x.A1, &y.A1)
	z.A1.Sub(&a, &b).Sub(&z.A1, &c)
	fp.MulBy5(&c)
	z.A0.Sub(&b, &c)
	return z
}

// Square sets z to the E2-product of x,x returns z
func (z *E2) Square(x *E2) *E2 {
	//algo 22 https://eprint.iacr.org/2010/354.pdf
	var c0, c2 fp.Element
	c0.Add(&x.A0, &x.A1)
	c2.Neg(&x.A1)
	fp.MulBy5(&c2)
	c2.Add(&c2, &x.A0)

	c0.Mul(&c0, &c2) // (x1+x2)*(x1+(u**2)x2)
	c2.Mul(&x.A0, &x.A1).Double(&c2)
	z.A1 = c2
	c2.Double(&c2)
	z.A0.Add(&c0, &c2)

	return z
}

// MulByNonResidue multiplies a E2 by (0,1)
func (z *E2) MulByNonResidue(x *E2) *E2 {
	a := x.A0
	b := x.A1 // fetching x.A1 in the function below is slower
	fp.MulBy5(&b)
	z.A0.Neg(&b)
	z.A1 = a
	return z
}

// MulByNonResidueInv multiplies a E2 by (0,1)^{-1}
func (z *E2) MulByNonResidueInv(x *E2) *E2 {
	//z.A1.MulByNonResidueInv(&x.A0)
	a := x.A1
	fiveinv := fp.Element{
		330620507644336508,
		9878087358076053079,
		11461392860540703536,
		6973035786057818995,
		8846909097162646007,
		104838758629667239,
	}
	z.A1.Mul(&x.A0, &fiveinv).Neg(&z.A1)
	z.A0 = a
	return z
}

// Inverse sets z to the E2-inverse of x, returns z
func (z *E2) Inverse(x *E2) *E2 {
	// Algorithm 8 from https://eprint.iacr.org/2010/354.pdf
	//var a, b, t0, t1, tmp fp.Element
	var t0, t1, tmp fp.Element
	a := &x.A0 // creating the buffers a, b is faster than querying &x.A0, &x.A1 in the functions call below
	b := &x.A1
	t0.Square(a)
	t1.Square(b)
	tmp.Set(&t1)
	fp.MulBy5(&tmp)
	t0.Add(&t0, &tmp)
	t1.Inverse(&t0)
	z.A0.Mul(a, &t1)
	z.A1.Mul(b, &t1).Neg(&z.A1)

	return z
}

// norm sets x to the norm of z
func (z *E2) norm(x *fp.Element) {
	var tmp fp.Element
	x.Square(&z.A1)
	tmp.Set(x)
	fp.MulBy5(&tmp)
	x.Square(&z.A0).Add(x, &tmp)
}

// MulBybTwistCurveCoeff multiplies by 1/(0,1)
func (z *E2) MulBybTwistCurveCoeff(x *E2) *E2 {

	var res E2
	res.A0.Set(&x.A1)
	res.A1.MulByNonResidueInv(&x.A0)
	z.Set(&res)

	return z
}

// lucasExponent is e = 3⁻¹ mod (p+1) as little-endian uint64 limbs,
// used by the Lucas V-chain in cbrtTorus.
var lucasExponent = [6]uint64{
	0x2c58400000000001,
	0x07ae746c10000000,
	0xb4fbcb653e031800,
	0xb360f3510051b12f,
	0xecbe57402435c313,
	0x008f68c207ec5af8,
}

// cbrtAndNormInverse computes m = cbrt(norm) and normInv = 1/norm from a
// single shared exponentiation, avoiding a separate Fp inversion.
//
// Let t = norm^((q-7)/9) and m₀ = norm · t = norm^((q+2)/9). Then:
//   - m₀³ = norm (unique cube root since q ≡ 7 mod 9)
//   - normInv = m₀⁵ · t⁴ = norm^(q-2)      (Fermat inverse)
func cbrtAndNormInverse(norm *fp.Element) (m, normInv fp.Element) {
	// t = norm^((q-7)/9)
	var t fp.Element
	t.ExpByCbrtHelperQMinus7Div9(*norm)

	// m = norm · t = norm^((q+2)/9)
	m.Mul(norm, &t)

	// normInv = m₀⁵ · t⁴
	// m₀⁵: m² → m⁴ → m⁵ = m⁴·m
	var m2, m4, m5 fp.Element
	m2.Square(&m)
	m4.Square(&m2)
	m5.Mul(&m4, &m)

	// t⁴: t² → t⁴
	var t2, t4 fp.Element
	t2.Square(&t)
	t4.Square(&t2)

	normInv.Mul(&m5, &t4) // m₀⁵ · t⁴ = norm^(q-2)

	return m, normInv
}

// cbrtTorus computes the cube root of x in E2 using the algebraic torus T₂(Fp).
//
// For BLS12-377, Fp2 = Fp[u]/(u²+5), so beta = -5.
// norm(x) = x₀² + 5·x₁², α_t = 2·(x₀² - 5·x₁²)/N(x).
// q ≡ 7 (mod 9) means unique cube root (no ζ-adjustment needed).
func (z *E2) cbrtTorus(x *E2) *E2 {
	if x.A1.IsZero() {
		if z.A0.Cbrt(&x.A0) == nil {
			return nil
		}
		z.A1.SetZero()
		return z
	}

	if x.A0.IsZero() {
		// x = x₁·u, so x³ = x₁³·u³ = x₁³·(-5u) = (-5x₁³)·u
		// We need y such that y³ = x₁·u
		// Try y = a·u: y³ = a³·u³ = -5a³·u. So -5a³ = x₁, a³ = -x₁/5.
		var negA1Over5 fp.Element
		fiveinv := fp.Element{
			330620507644336508,
			9878087358076053079,
			11461392860540703536,
			6973035786057818995,
			8846909097162646007,
			104838758629667239,
		}
		negA1Over5.Neg(&x.A1)
		negA1Over5.Mul(&negA1Over5, &fiveinv)
		var y E2
		if y.A1.Cbrt(&negA1Over5) == nil {
			return nil
		}
		y.A0.SetZero()
		return z.cbrtVerifyAndAdjust(x, &y)
	}

	// x₀², x₁² — reused for both norm and α_t
	var x0sq, x1sq fp.Element
	x0sq.Square(&x.A0)
	x1sq.Square(&x.A1)

	// N = x₀² + 5·x₁² (norm of x, since beta = -5)
	var norm, fiveX1sq fp.Element
	fiveX1sq.Set(&x1sq)
	fp.MulBy5(&fiveX1sq)
	norm.Add(&x0sq, &fiveX1sq)

	// m = cbrt(N) and normInv = 1/N from shared exponentiation
	m, normInv := cbrtAndNormInverse(&norm)

	// α_t = 2·(x₀² - 5·x₁²)/N = trace of x^{p-1} on T₂
	var alphaT fp.Element
	alphaT.Sub(&x0sq, &fiveX1sq)
	alphaT.Double(&alphaT)
	alphaT.Mul(&alphaT, &normInv)

	// s₁ = V_e(α_t, 1) where e = 3⁻¹ mod (p+1), Q = 1
	sp := lucasV(&alphaT)

	// Recovery: z₀ = x₀/(m·(s₁-1)), z₁ = x₁/(m·(s₁+1))
	var one, s1m1, s1p1, d0, d1, d0d1, d0d1Inv fp.Element
	one.SetOne()
	s1m1.Sub(&sp, &one)
	s1p1.Add(&sp, &one)
	d0.Mul(&m, &s1m1)
	d1.Mul(&m, &s1p1)

	d0d1.Mul(&d0, &d1)
	d0d1Inv.Inverse(&d0d1)

	var y E2
	y.A0.Mul(&d1, &d0d1Inv).Mul(&y.A0, &x.A0)
	y.A1.Mul(&d0, &d0d1Inv).Mul(&y.A1, &x.A1)

	return z.cbrtVerifyAndAdjust(x, &y)
}

// lucasV computes V_e(alpha, 1) where e = 3⁻¹ mod (p+1), using the
// Lucas V-sequence with Q=1 and a Montgomery ladder on precomputed bits.
func lucasV(alpha *fp.Element) fp.Element {
	var v0, v1, two fp.Element
	two.SetUint64(2)
	v0.Set(alpha)
	v1.Square(alpha).Sub(&v1, &two)

	var prod fp.Element

	// Process bits 374 down to 1 (bit 375 is the leading 1, already consumed by init)
	for i := 374; i >= 1; i-- {
		bit := (lucasExponent[i/64] >> uint(i%64)) & 1

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
