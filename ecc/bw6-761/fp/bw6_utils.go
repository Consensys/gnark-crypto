// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fp

// MulByNonResidue multiplies a fp.Element by -4
func (z *Element) MulByNonResidue(x *Element) *Element {
	z.Double(x).Double(z).Neg(z)
	return z
}

// MulByNonResidueInv multiplies a fp.Element by (-4)**-1
func (z *Element) MulByNonResidueInv(x *Element) *Element {

	nrInv := Element{
		8571757465769615091,
		6221412002326125864,
		16781361031322833010,
		18148962537424854844,
		6497335359600054623,
		17630955688667215145,
		15638647242705587201,
		830917065158682257,
		6848922060227959954,
		4142027113657578586,
		12050453106507568375,
		55644342162350184,
	}

	z.Mul(x, &nrInv)

	return z
}
