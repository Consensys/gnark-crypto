// Copyright 2020-2024 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fptower

// MulBybTwistCurveCoeff multiplies by 4*(0,1)
func (z *E4) MulBybTwistCurveCoeff(x *E4) *E4 {

	var res E4
	res.B1.Set(&x.B0)
	res.B0.MulByNonResidue(&x.B1)

	z.Double(&res).
		Double(z)

	return z
}
