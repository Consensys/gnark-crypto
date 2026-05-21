// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package extensions

var (
	cbrtE8Omega  E8
	cbrtE8Omega2 E8
)

var cbrtE8LucasExponent = [2]uint64{
	2930905110336765953,
	372437575807401643,
}

func init() {
	cbrtE8Omega.C0.B0 = cbrtE2Omega
	cbrtE8Omega2.Square(&cbrtE8Omega)
}

// Cbrt sets z to the cube root of x and returns z.
// It returns nil if x is not a cubic residue.
func (z *E8) Cbrt(x *E8) *E8 {
	if x.C1.IsZero() {
		if z.C0.Cbrt(&x.C0) == nil {
			return nil
		}
		z.C1.SetZero()
		return z
	}

	if x.C0.IsZero() {
		var y E8
		var x1OverNR E4
		x1OverNR.Mul(&x.C1, &cbrtE4NRInv)
		if y.C1.Cbrt(&x1OverNR) == nil {
			return nil
		}
		y.C0.SetZero()
		return cbrtVerifyAndAdjustE8(z.Set(&y), x)
	}

	var x0sq, x1sq, betaX1sq, norm E4
	x0sq.Square(&x.C0)
	x1sq.Square(&x.C1)
	betaX1sq.MulByQuadraticNonResidue(&x1sq)
	norm.Sub(&x0sq, &betaX1sq)

	var m, normInv E4
	if m.Cbrt(&norm) == nil {
		return nil
	}
	normInv.Inverse(&norm)

	var halfTau, tau E4
	halfTau.Add(&x0sq, &betaX1sq)
	halfTau.Mul(&halfTau, &normInv)
	tau.Double(&halfTau)

	var x0x1, imY E4
	x0x1.Mul(&x.C0, &x.C1)
	imY.Double(&x0x1)
	imY.Mul(&imY, &normInv)

	te, te1 := lucasV2E4Cbrt(&tau)

	var wa0, wa1 E4
	wa0.Mul(&halfTau, &te)
	wa0.Sub(&te1, &wa0)
	wa1.Mul(&imY, &te)

	var delta, deltaInv, sIm, k E4
	delta.Square(&tau).Sub(&delta, &cbrtE4One).Sub(&delta, &cbrtE4One).Sub(&delta, &cbrtE4One).Sub(&delta, &cbrtE4One)
	if delta.IsZero() {
		return nil
	}
	deltaInv.Inverse(&delta)
	sIm.Double(&imY)
	k.Mul(&sIm, &deltaInv)

	var gamma0, gamma1 E4
	gamma0.Mul(&wa1, &k)
	gamma0.MulByQuadraticNonResidue(&gamma0)
	gamma1.Mul(&wa0, &k)

	var mInv E4
	mInv.Square(&m).Mul(&mInv, &normInv)

	var y E8
	var t1, t2 E4
	t1.Mul(&x.C0, &gamma0)
	t2.Mul(&x.C1, &gamma1)
	t2.MulByQuadraticNonResidue(&t2)
	y.C0.Sub(&t1, &t2).Mul(&y.C0, &mInv)
	t1.Mul(&x.C1, &gamma0)
	t2.Mul(&x.C0, &gamma1)
	y.C1.Sub(&t1, &t2).Mul(&y.C1, &mInv)
	return cbrtVerifyAndAdjustE8(z.Set(&y), x)
}

func cbrtVerifyAndAdjustE8(z, x *E8) *E8 {
	var check, y E8
	check.Square(z).Mul(&check, z)
	if check.Equal(x) {
		return z
	}

	y.Mul(z, &cbrtE8Omega)
	check.Square(&y).Mul(&check, &y)
	if check.Equal(x) {
		return z.Set(&y)
	}

	y.Mul(z, &cbrtE8Omega2)
	check.Square(&y).Mul(&check, &y)
	if check.Equal(x) {
		return z.Set(&y)
	}

	return nil
}

func lucasV2E4Cbrt(alpha *E4) (E4, E4) {
	var v0, v1, prod E4
	var two E4
	two.B0.A0.SetUint64(2)
	v0.Set(alpha)
	v1.Square(alpha).Sub(&v1, &two)
	for i := 121; i >= 1; i-- {
		bit := (cbrtE8LucasExponent[i/64] >> uint(i%64)) & 1
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

	var te, te1 E4
	te.Mul(&v0, &v1).Sub(&te, alpha)
	te1.Square(&v1).Sub(&te1, &two)
	return te, te1
}
