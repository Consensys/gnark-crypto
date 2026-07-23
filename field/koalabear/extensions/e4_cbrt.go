// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package extensions

var (
	cbrtE4One   E4
	cbrtE4NRInv E4
)

var cbrtE4LucasExponent = [1]uint64{
	1513303301209194497,
}

func init() {
	cbrtE4One.SetOne()
	cbrtE4NRInv.B1.SetOne()
	cbrtE4NRInv.Inverse(&cbrtE4NRInv)
}

// Cbrt sets z to the cube root of x and returns z.
// It returns nil if x is not a cubic residue.
//
// The candidate is computed into a local E4 and only written to z at the
// end, so x.Cbrt(&x) is safe (does not corrupt x mid-computation).
func (z *E4) Cbrt(x *E4) *E4 {
	if x.B1.IsZero() {
		if z.B0.Cbrt(&x.B0) == nil {
			return nil
		}
		z.B1.SetZero()
		return z
	}

	var y E4

	if x.B0.IsZero() {
		var x1OverNR E2
		x1OverNR.Mul(&x.B1, &cbrtE2NRInv)
		if y.B1.Cbrt(&x1OverNR) == nil {
			return nil
		}
		y.B0.SetZero()
		if cbrtVerifyE4(&y, x) == nil {
			return nil
		}
		return z.Set(&y)
	}

	var x0sq, x1sq, betaX1sq, norm E2
	x0sq.Square(&x.B0)
	x1sq.Square(&x.B1)
	betaX1sq.MulByQuadraticNonResidue(&x1sq)
	norm.Sub(&x0sq, &betaX1sq)

	var m, normInv E2
	if m.Cbrt(&norm) == nil {
		return nil
	}
	normInv.Inverse(&norm)

	var halfTau, tau E2
	halfTau.Add(&x0sq, &betaX1sq)
	halfTau.Mul(&halfTau, &normInv)
	tau.Double(&halfTau)

	var x0x1, imY E2
	x0x1.Mul(&x.B0, &x.B1)
	imY.Double(&x0x1)
	imY.Mul(&imY, &normInv)

	te, te1 := lucasV2E2Cbrt(&tau)

	var wa0, wa1 E2
	wa0.Mul(&halfTau, &te)
	wa0.Sub(&te1, &wa0)
	wa1.Mul(&imY, &te)

	var delta, deltaInv, sIm, k E2
	delta.Square(&tau).Sub(&delta, &cbrtE2One).Sub(&delta, &cbrtE2One).Sub(&delta, &cbrtE2One).Sub(&delta, &cbrtE2One)
	if delta.IsZero() {
		return nil
	}
	deltaInv.Inverse(&delta)
	sIm.Double(&imY)
	k.Mul(&sIm, &deltaInv)

	var gamma0, gamma1 E2
	gamma0.Mul(&wa1, &k)
	gamma0.MulByQuadraticNonResidue(&gamma0)
	gamma1.Mul(&wa0, &k)

	var mInv E2
	mInv.Square(&m).Mul(&mInv, &normInv)

	var t1, t2 E2
	t1.Mul(&x.B0, &gamma0)
	t2.Mul(&x.B1, &gamma1)
	t2.MulByQuadraticNonResidue(&t2)
	y.B0.Sub(&t1, &t2).Mul(&y.B0, &mInv)
	t1.Mul(&x.B1, &gamma0)
	t2.Mul(&x.B0, &gamma1)
	y.B1.Sub(&t1, &t2).Mul(&y.B1, &mInv)
	if cbrtVerifyE4(&y, x) == nil {
		return nil
	}
	return z.Set(&y)
}

// cbrtVerifyE4 returns z iff z³ == x, else nil.
//
// The E4.Cbrt construction always returns a true cube root on the first try
// when x is a cubic residue (verified empirically over many random cubes), so
// no cube-root-of-unity rotation is needed — the verify is just a safety
// check against non-residue inputs.
func cbrtVerifyE4(z, x *E4) *E4 {
	var check E4
	check.Square(z).Mul(&check, z)
	if !check.Equal(x) {
		return nil
	}
	return z
}

func lucasV2E2Cbrt(alpha *E2) (E2, E2) {
	var v0, v1, prod E2
	var two E2
	two.A0.SetUint64(2)
	v0.Set(alpha)
	v1.Square(alpha).Sub(&v1, &two)
	for i := 59; i >= 1; i-- {
		bit := (cbrtE4LucasExponent[0] >> uint(i)) & 1
		if bit == 0 {
			prod.Mul(&v0, &v1).Sub(&prod, alpha)
			v1.Set(&prod)
			v0.Square(&v0).Sub(&v0, &two)
		} else {
			prod.Mul(&v0, &v1).Sub(&prod, alpha)
			v0.Set(&prod)
			v1.Square(&v1).Sub(&v1, &two)
		}
	}

	var te, te1 E2
	te.Mul(&v0, &v1).Sub(&te, alpha)
	te1.Square(&v1).Sub(&te1, &two)
	return te, te1
}
