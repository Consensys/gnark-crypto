// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package extensions

import fr "github.com/consensys/gnark-crypto/field/koalabear"

var (
	cbrtFpOne      fr.Element
	cbrtFpTwo      fr.Element
	cbrtFpThree    fr.Element
	cbrtFpThreeInv fr.Element
	cbrtE2One      E2
	cbrtE2Omega    E2
	cbrtE2Omega2   E2
	cbrtE2NRInv    E2
)

const cbrtE2MuLucasExponent uint64 = 473490319

func init() {
	cbrtFpOne.SetOne()
	cbrtFpTwo.SetUint64(2)
	cbrtFpThree.SetUint64(3)
	cbrtFpThreeInv.Inverse(&cbrtFpThree)

	cbrtE2One.SetOne()
	cbrtE2NRInv.A0.SetZero()
	cbrtE2NRInv.A1.Set(&cbrtFpThreeInv)

	var sqrtMinusThree E2
	sqrtMinusThree.A0.Neg(&cbrtFpThree)
	sqrtMinusThree.Sqrt(&sqrtMinusThree)
	cbrtE2Omega.Sub(&sqrtMinusThree, &cbrtE2One)
	cbrtE2Omega.A0.Halve()
	cbrtE2Omega.A1.Halve()
	cbrtE2Omega2.Square(&cbrtE2Omega)
}

// Cbrt sets z to the cube root of x and returns z.
// It returns nil if x is not a cubic residue.
func (z *E2) Cbrt(x *E2) *E2 {
	if x.A1.IsZero() {
		z.A0.Cbrt(&x.A0)
		z.A1.SetZero()
		return z
	}

	if x.A0.IsZero() {
		z.A0.SetZero()
		z.A1.Mul(&x.A1, &cbrtFpThreeInv)
		z.A1.Cbrt(&z.A1)
		return cbrtVerifyE2(z, x)
	}

	var x0sq, x1sq, betaX1sq, norm fr.Element
	x0sq.Square(&x.A0)
	x1sq.Square(&x.A1)
	betaX1sq.Mul(&x1sq, &cbrtFpThree)
	norm.Sub(&x0sq, &betaX1sq)

	m, normInv, deltaInv, ok := cbrtAndNormInverseE2(&norm, &x0sq, &x1sq)
	if !ok {
		return nil
	}

	var halfTau, tau fr.Element
	halfTau.Add(&x0sq, &betaX1sq)
	halfTau.Mul(&halfTau, &normInv)
	tau.Double(&halfTau)

	te, te1 := lucasV2Fp(&tau)

	var x0x1, imY fr.Element
	x0x1.Mul(&x.A0, &x.A1)
	imY.Double(&x0x1).Neg(&imY)
	imY.Mul(&imY, &normInv)

	var wa0, wa1 fr.Element
	wa0.Mul(&halfTau, &te)
	wa0.Sub(&te1, &wa0)
	wa1.Mul(&imY, &te)

	var sIm, k fr.Element
	sIm.Double(&imY)
	k.Mul(&sIm, &deltaInv)

	var gamma0, gamma1 fr.Element
	gamma0.Mul(&wa1, &k).Mul(&gamma0, &cbrtFpThree)
	gamma1.Mul(&wa0, &k)

	var mInv fr.Element
	mInv.Square(&m).Mul(&mInv, &normInv)

	var t1, t2 fr.Element
	t1.Mul(&x.A0, &gamma0)
	t2.Mul(&x.A1, &gamma1).Mul(&t2, &cbrtFpThree)
	z.A0.Sub(&t1, &t2).Mul(&z.A0, &mInv)
	t1.Mul(&x.A1, &gamma0)
	t2.Mul(&x.A0, &gamma1)
	z.A1.Sub(&t1, &t2).Mul(&z.A1, &mInv)
	if out := cbrtVerifyE2(z, x); out != nil {
		return out
	}

	var sigma fr.Element
	sigma.Set(&te)
	var d0, d1, d0d1, d0d1Inv fr.Element
	d0.Sub(&sigma, &cbrtFpOne)
	d0.Mul(&m, &d0)
	d1.Add(&sigma, &cbrtFpOne)
	d1.Mul(&m, &d1)
	d0d1.Mul(&d0, &d1)
	if d0d1.IsZero() {
		return nil
	}
	d0d1Inv.Inverse(&d0d1)

	z.A0.Mul(&d1, &d0d1Inv).Mul(&z.A0, &x.A0)
	z.A1.Mul(&d0, &d0d1Inv).Mul(&z.A1, &x.A1)
	return cbrtVerifyE2(z, x)
}

func cbrtAndNormInverseE2(norm, x0sq, x1sq *fr.Element) (m, normInv, deltaInv fr.Element, ok bool) {
	var U, x0x1, U2, U3, w fr.Element
	x0x1.Mul(x0sq, x1sq)
	U.Mul(&x0x1, norm)
	U.Double(&U).Double(&U)
	U.Double(&U)
	U.Mul(&U, &cbrtFpThree)
	U2.Square(&U)
	U3.Mul(&U2, &U)
	w.Mul(&U3, norm)

	var t, t2, t4, t6, t8, t9 fr.Element
	t.ExpByCbrtHelperQMinus2Div9(w)
	t2.Square(&t)
	t4.Square(&t2)
	t6.Mul(&t4, &t2)
	t8.Square(&t4)
	t9.Mul(&t8, &t)

	var cbrtW, UInv fr.Element
	cbrtW.Mul(&w, &t6)
	UInv.Mul(&U2, norm).Mul(&UInv, &t9)
	m.Mul(&cbrtW, &UInv)
	normInv.Mul(&U3, &t9)

	var check fr.Element
	check.Square(&m).Mul(&check, &m)
	if !check.Equal(norm) {
		return m, normInv, deltaInv, false
	}

	var norm2, norm3 fr.Element
	norm2.Square(norm)
	norm3.Mul(&norm2, norm)
	deltaInv.Mul(&norm3, &UInv)
	return m, normInv, deltaInv, true
}

func cbrtVerifyE2(z, x *E2) *E2 {
	var check E2
	check.Square(z).Mul(&check, z)
	if !check.Equal(x) {
		return nil
	}
	return z
}

func lucasV2Fp(alpha *fr.Element) (fr.Element, fr.Element) {
	var v0, v1, prod fr.Element
	v0.Set(alpha)
	v1.Square(alpha).Sub(&v1, &cbrtFpTwo)
	for i := 27; i >= 1; i-- {
		bit := (cbrtE2MuLucasExponent >> uint(i)) & 1
		prod.Mul(&v0, &v1).Sub(&prod, alpha)
		if bit == 0 {
			v1.Set(&prod)
			v0.Square(&v0).Sub(&v0, &cbrtFpTwo)
		} else {
			v0.Set(&prod)
			v1.Square(&v1).Sub(&v1, &cbrtFpTwo)
		}
	}
	var te, te1 fr.Element
	te.Mul(&v0, &v1).Sub(&te, alpha)
	te1.Square(&v1).Sub(&te1, &cbrtFpTwo)
	return te, te1
}
