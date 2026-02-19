// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fptower

import "github.com/consensys/gnark-crypto/ecc/bls12-381/fp"

// used with !amd64, make staticcheck happier.
var (
	_ = mulGenericE2
	_ = squareGenericE2
)

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
	z.A0.Sub(&b, &c)
}

// Square sets z to the E2-product of x,x returns z
func squareGenericE2(z, x *E2) *E2 {
	// adapted from algo 22 https://eprint.iacr.org/2010/354.pdf
	var a, b fp.Element
	a.Add(&x.A0, &x.A1)
	b.Sub(&x.A0, &x.A1)
	a.Mul(&a, &b)
	b.Mul(&x.A0, &x.A1).Double(&b)
	z.A0.Set(&a)
	z.A1.Set(&b)
	return z
}

var twoInv = fp.Element{
	1730508156817200468,
	9606178027640717313,
	7150789853162776431,
	7936136305760253186,
	15245073033536294050,
	1728177566264616342,
}

// MulByNonResidueInv multiplies a E2 by (1,1)^{-1}
func (z *E2) MulByNonResidueInv(x *E2) *E2 {

	var tmp fp.Element
	tmp.Add(&x.A0, &x.A1)
	z.A1.Sub(&x.A1, &x.A0).Mul(&z.A1, &twoInv)
	z.A0.Set(&tmp).Mul(&z.A0, &twoInv)

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

// MulBybTwistCurveCoeff multiplies by 4(1,1)
func (z *E2) MulBybTwistCurveCoeff(x *E2) *E2 {

	var res E2
	res.A0.Sub(&x.A0, &x.A1)
	res.A1.Add(&x.A0, &x.A1)
	z.Double(&res).
		Double(z)

	return z
}

// lucasExponent is e = 3⁻¹ mod (p+1) as little-endian uint64 limbs,
// used by the Lucas V-chain in cbrtTorus.
var lucasExponent = [6]uint64{
	0x9354ffffffffe38f,
	0x0a395554e5c6aaaa,
	0xcd104635a790520c,
	0xcc27c3d6fbd7063f,
	0x190937e76bc3e447,
	0x08ab05f8bdd54cde,
}

// cbrtAndNormInverse computes m = cbrt(norm) and normInv = 1/norm from a
// single shared exponentiation, avoiding a separate Fp inversion.
//
// Let t = norm^((q-10)/27) and m₀ = norm · t² = norm^((q+17)/27). Then:
//   - m₀³ should equal norm (possibly after ζ-adjustment)
//   - normInv = m₀⁸ · t¹¹ = norm^(q-2)      (Fermat inverse)
//
// The ζ-adjustment for the cube root is identical to fp.Cbrt.
func cbrtAndNormInverse(norm *fp.Element) (m, normInv fp.Element, ok bool) {
	// t = norm^((q-10)/27)
	var t fp.Element
	t.ExpByCbrtHelperQMinus10Div27(*norm)

	// t² = t^2
	var t2 fp.Element
	t2.Square(&t)

	// m = norm · t² = norm^((q+17)/27)
	m.Mul(norm, &t2)

	// normInv = m₀⁸ · t¹¹ = norm^(q-2)
	// m⁸: m² → m⁴ → m⁸
	var m2, m4, m8 fp.Element
	m2.Square(&m)
	m4.Square(&m2)
	m8.Square(&m4)

	// t¹¹ = t⁸ · t² · t
	// t⁴ = (t²)²
	var t4, t8 fp.Element
	t4.Square(&t2)
	t8.Square(&t4)
	normInv.Mul(&t8, &t2)      // t¹⁰
	normInv.Mul(&normInv, &t)  // t¹¹
	normInv.Mul(&normInv, &m8) // m⁸ · t¹¹

	// Verify m³ = norm, adjust by ζ if needed (same logic as fp.Cbrt)
	var c fp.Element
	c.Mul(&m2, &m)
	if !c.Equal(norm) {
		var zeta = fp.Element{
			13616190144799058984,
			9227582506135211912,
			4426607408274926740,
			7455198167498346307,
			10794825842164118204,
			335101026345095675,
		}
		var zeta2 = fp.Element{
			3828863564860874189,
			5918733612565202776,
			16843310164143221096,
			16127847466718491017,
			17435063908385505950,
			407112797415018074,
		}
		var omega = fp.Element{
			14772873186050699377,
			6749526151121446354,
			6372666795664677781,
			10283423008382700446,
			286397964926079186,
			1796971870900422465,
		}
		var omega2 = fp.Element{
			3526659474838938856,
			17562030475567847978,
			1632777218702014455,
			14009062335050482331,
			3906511377122991214,
			368068849512964448,
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

	return m, normInv, true
}

// cbrtTorus computes the cube root of x in E2 using the algebraic torus T₂(Fp).
//
// Let r = z^{p-1} where z³ = x. Then r³ = x^{p-1} lies on T₂(Fp) (norm 1).
// The trace of r on T₂ is s₁ = V_e(α_t, 1) where:
//   - α_t = 2·(x₀² - x₁²)/N(x) is the trace of x^{p-1} on T₂
//   - e = 3⁻¹ mod (p+1)
//
// Recovery: z₀ = x₀/(m·(s₁-1)), z₁ = x₁/(m·(s₁+1)), where m = cbrt(N(x)).
func (z *E2) cbrtTorus(x *E2) *E2 {
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

	// N = x₀² + x₁² (norm of x, since beta = -1)
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

	// Process bits 378 down to 1 (bit 379 is the leading 1, already consumed by init)
	for i := 378; i >= 1; i-- {
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
