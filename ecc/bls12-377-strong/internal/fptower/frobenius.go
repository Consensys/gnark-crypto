// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fptower

import "github.com/consensys/gnark-crypto/ecc/bls12-377-strong/fp"

// Frobenius set z to Frobenius(x), return z
func (z *E12) Frobenius(x *E12) *E12 {
	// Algorithm 28 from https://eprint.iacr.org/2010/354.pdf (beware typos!)
	var t [6]E2

	// Frobenius acts on fp2 by conjugation
	t[0].Conjugate(&x.C0.B0)
	t[1].Conjugate(&x.C0.B1)
	t[2].Conjugate(&x.C0.B2)
	t[3].Conjugate(&x.C1.B0)
	t[4].Conjugate(&x.C1.B1)
	t[5].Conjugate(&x.C1.B2)

	t[1].MulByNonResidue1Power2(&t[1])
	t[2].MulByNonResidue1Power4(&t[2])
	t[3].MulByNonResidue1Power1(&t[3])
	t[4].MulByNonResidue1Power3(&t[4])
	t[5].MulByNonResidue1Power5(&t[5])

	z.C0.B0 = t[0]
	z.C0.B1 = t[1]
	z.C0.B2 = t[2]
	z.C1.B0 = t[3]
	z.C1.B1 = t[4]
	z.C1.B2 = t[5]

	return z
}

// FrobeniusSquare set z to Frobenius^2(x), and return z
func (z *E12) FrobeniusSquare(x *E12) *E12 {
	// Algorithm 29 from https://eprint.iacr.org/2010/354.pdf (beware typos!)
	var t [6]E2

	t[1].MulByNonResidue2Power2(&x.C0.B1)
	t[2].MulByNonResidue2Power4(&x.C0.B2)
	t[3].MulByNonResidue2Power1(&x.C1.B0)
	t[4].MulByNonResidue2Power3(&x.C1.B1)
	t[5].MulByNonResidue2Power5(&x.C1.B2)

	z.C0.B0 = x.C0.B0
	z.C0.B1 = t[1]
	z.C0.B2 = t[2]
	z.C1.B0 = t[3]
	z.C1.B1 = t[4]
	z.C1.B2 = t[5]

	return z
}

// MulByNonResidue1Power1 set z=x*(1,1)^(1*(p^1-1)/6) and return z
func (z *E2) MulByNonResidue1Power1(x *E2) *E2 {
	var b E2
	b.A0.SetString("118022375362511969752646055245697019732962424165626948670975149478625353268059691701138390100656439435356726754607")
	b.A1.SetString("100838761346247805436245255368139017013448653656674880346872577520281859135749442941515208726974485834436634520956")
	z.Mul(x, &b)
	return z
}

// MulByNonResidue1Power2 set z=x*(1,1)^(2*(p^1-1)/6) and return z
func (z *E2) MulByNonResidue1Power2(x *E2) *E2 {
	var b E2
	b.A1.SetString("70427388980041189209419192134717659880491069728330620094944516061531890781668663842312340635652")
	z.Mul(x, &b)
	return z
}

// MulByNonResidue1Power3 set z=x*(1,1)^(3*(p^1-1)/6) and return z
func (z *E2) MulByNonResidue1Power3(x *E2) *E2 {
	var b E2
	b.A0.SetString("144394925554309751978230035908754812058838028871212916374968753301545726589628552395959492401580323225873322644219")
	b.A1.SetString("144394925554309751978230035908754812058838028871212916374968753301545726589628552395959492401580323225873322644219")
	z.Mul(x, &b)
	return z
}

// MulByNonResidue1Power4 set z=x*(1,1)^(4*(p^1-1)/6) and return z
func (z *E2) MulByNonResidue1Power4(x *E2) *E2 {
	var b fp.Element
	b.SetString("70427388980041189209419192134717659880491069728330620094944516061531890781668663842312340635653")
	z.A0.Mul(&x.A0, &b)
	z.A1.Mul(&x.A1, &b)
	return z
}

// MulByNonResidue1Power5 set z=x*(1,1)^(5*(p^1-1)/6) and return z
func (z *E2) MulByNonResidue1Power5(x *E2) *E2 {
	var b E2
	b.A0.SetString("43556164208061946541984780540615795045389375214538036028096175781263867453879109454444283674605837391436688123263")
	b.A1.SetString("175304972500697828646906530073220241701021702607763792989751551217643344949930025188209315153025087878356673152300")
	z.Mul(x, &b)
	return z
}

// MulByNonResidue2Power1 set z=x*(1,1)^(1*(p^2-1)/6) and return z
func (z *E2) MulByNonResidue2Power1(x *E2) *E2 {
	var b fp.Element
	b.SetString("218861136708759775118463921633794847536991885687584169137356657270576592308864618581121708045962261427481020639911")
	z.A0.Mul(&x.A0, &b)
	z.A1.Mul(&x.A1, &b)
	return z
}

// MulByNonResidue2Power2 set z=x*(1,1)^(2*(p^2-1)/6) and return z
func (z *E2) MulByNonResidue2Power2(x *E2) *E2 {
	var b fp.Element
	b.SetString("218861136708759775118463921633794847536991885687584169137356657270576592308864618581121708045962261427481020639910")
	z.A0.Mul(&x.A0, &b)
	z.A1.Mul(&x.A1, &b)
	return z
}

// MulByNonResidue2Power3 set z=x*(1,1)^(3*(p^2-1)/6) and return z
func (z *E2) MulByNonResidue2Power3(x *E2) *E2 {
	var b fp.Element
	b.SetString("218861136708759775188891310613836036746411077822301829017847726998907212403809134642653598827630925269793361275562")
	z.A0.Mul(&x.A0, &b)
	z.A1.Mul(&x.A1, &b)
	return z
}

// MulByNonResidue2Power4 set z=x*(1,1)^(4*(p^2-1)/6) and return z
func (z *E2) MulByNonResidue2Power4(x *E2) *E2 {
	var b fp.Element
	b.SetString("70427388980041189209419192134717659880491069728330620094944516061531890781668663842312340635652")
	z.A0.Mul(&x.A0, &b)
	z.A1.Mul(&x.A1, &b)
	return z
}

// MulByNonResidue2Power5 set z=x*(1,1)^(5*(p^2-1)/6) and return z
func (z *E2) MulByNonResidue2Power5(x *E2) *E2 {
	var b fp.Element
	b.SetString("70427388980041189209419192134717659880491069728330620094944516061531890781668663842312340635653")
	z.A0.Mul(&x.A0, &b)
	z.A1.Mul(&x.A1, &b)
	return z
}
