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

// MulByNonResidue1Power1 set z=x*(2,1)^(1*(p^1-1)/6) and return z
func (z *E2) MulByNonResidue1Power1(x *E2) *E2 {
	var a E2
	a.A0.SetString("125727767022531944802882402103319481348859762347463543974374015693148713445069532645742307214154121370636188162845")
	a.A1.SetString("123462512926791639928528100498369354423274226030963990368186865357298339918512805417597545136207631028366565348647")
	z.Mul(x, &a)
	return z
}

// MulByNonResidue1Power2 set z=x*(2,1)^(2*(p^1-1)/6) and return z
func (z *E2) MulByNonResidue1Power2(x *E2) *E2 {
	var a E2
	a.A0.SetString("88899026867714575188751882972325772064086217814419839293293529345724897417332022887401511902040845984717680389861")
	a.A1.SetString("176618080069251975695932762281816448616069709450241741294490249889740395629078583406915317197028257142967053061951")
	z.Mul(x, &a)
	return z
}

// MulByNonResidue1Power3 set z=x*(3,1)^(3*(p^1-1)/6) and return z
func (z *E2) MulByNonResidue1Power3(x *E2) *E2 {
	var a E2
	a.A0.SetString("137764902161523724326024409884254444150371201229781842718963172888657649467106698224478747992403404446630035153285")
	a.A1.SetString("70311022050158941927701660579722878126128079702520884352918338625171444218636495686162913580612066007627520089883")
	z.Mul(x, &a)
	return z
}

// MulByNonResidue1Power4 set z=x*(2,1)^(4*(p^1-1)/6) and return z
func (z *E2) MulByNonResidue1Power4(x *E2) *E2 {
	var a E2
	a.A0.SetString("34732284740132355627735192228232544713496148837914996228895876473809357265661937159603222209943097214023504872468")
	a.A1.SetString("11738325102981970909903475540228683675145941045083188041844660024960312224636080526302380743722458880609435809443")
	z.Mul(x, &a)
	return z
}

// MulByNonResidue1Power5 set z=x*(2,1)^(5*(p^1-1)/6) and return z
func (z *E2) MulByNonResidue1Power5(x *E2) *E2 {
	var a E2
	a.A0.SetString("191929863109124368524854013324891348967149251259147557480058022342158207832954861144348824542013365603692629661146")
	a.A1.SetString("26503915771196494510228209740886042592697998502129558820554157718109827400586151390629860340245053385636979209387")
	z.Mul(x, &a)
	return z
}

// MulByNonResidue2Power1 set z=x*(2,1)^(1*(p^2-1)/6) and return z
func (z *E2) MulByNonResidue2Power1(x *E2) *E2 {
	var b fp.Element
	b.SetString("66749594872528601112692535115452694730463020851273681873215020777094334903430823628450258725296")
	z.A0.Mul(&x.A0, &b)
	z.A1.Mul(&x.A1, &b)
	return z
}

// MulByNonResidue2Power2 set z=x*(2,1)^(2*(p^2-1)/6) and return z
func (z *E2) MulByNonResidue2Power2(x *E2) *E2 {
	var b fp.Element
	b.SetString("66749594872528601112692535115452694730463020851273681873215020777094334903430823628450258725295")
	z.A0.Mul(&x.A0, &b)
	z.A1.Mul(&x.A1, &b)
	return z
}

// MulByNonResidue2Power3 set z=x*(2,1)^(3*(p^2-1)/6) and return z
func (z *E2) MulByNonResidue2Power3(x *E2) *E2 {
	var b fp.Element
	b.SetString("205218782272888506724347159188786010174614322757042801085008007152143854715576900762794582404194742885632550216686")
	z.A0.Mul(&x.A0, &b)
	z.A1.Mul(&x.A1, &b)
	return z
}

// MulByNonResidue2Power4 set z=x*(2,1)^(4*(p^2-1)/6) and return z
func (z *E2) MulByNonResidue2Power4(x *E2) *E2 {
	var b fp.Element
	b.SetString("205218782272888506657597564316257409061921787641590106354544986300870172842361879985700247500763919257182291491391")
	z.A0.Mul(&x.A0, &b)
	z.A1.Mul(&x.A1, &b)
	return z
}

// MulByNonResidue2Power5 set z=x*(0,1)^(5*(p^2-1)/6) and return z
func (z *E2) MulByNonResidue2Power5(x *E2) *E2 {
	var b fp.Element
	b.SetString("205218782272888506657597564316257409061921787641590106354544986300870172842361879985700247500763919257182291491392")
	z.A0.Mul(&x.A0, &b)
	z.A1.Mul(&x.A1, &b)
	return z
}
