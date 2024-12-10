// Copyright 2020 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fptower

// MulBybTwistCurveCoeff multiplies by 1/(0,1)
func (z *E4) MulBybTwistCurveCoeff(x *E4) *E4 {

	var res E4
	res.B0.Set(&x.B1)
	res.B1.MulByNonResidueInv(&x.B0)
	z.Set(&res)

	return z
}
