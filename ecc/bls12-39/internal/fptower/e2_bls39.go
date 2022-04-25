// Copyright 2020 ConsenSys AG
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fptower

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-39/fp"
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
	fp.MulBy3(&c)
	z.A0.Add(&b, &c)
	return z
}

// Square sets z to the E2-product of x,x returns z
func (z *E2) Square(x *E2) *E2 {
	//algo 22 https://eprint.iacr.org/2010/354.pdf
	var c0, c1, c2, c3 fp.Element
	c0.Sub(&x.A0, &x.A1)
	c1.Neg(&x.A1)
	fp.MulBy3(&c1)
	c3.Add(&x.A0, &c1)
	c2.Mul(&x.A0, &x.A1)
	c0.Mul(&c0, &c3).
		Add(&c0, &c2)
	z.A1.Double(&c2)
	fp.MulBy3(&c2)
	z.A0.Add(&c0, &c2)

	return z
}

// MulByNonResidue multiplies a E2 by (1,1)
func (z *E2) MulByNonResidue(x *E2) *E2 {
	var res E2
	res.A0.Set(&x.A1)
	fp.MulBy3(&res.A0)
	res.A0.Add(&x.A0, &res.A0)
	res.A1.Add(&x.A0, &x.A1)

	z.Set(&res)
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
	fp.MulBy3(&tmp)
	t0.Sub(&t0, &tmp)
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
	fp.MulBy3(&tmp)
	x.Square(&z.A0).Sub(x, &tmp)
}

// MulByNonResidueInv multiplies a E2 by (1,1)^{-1}
func (z *E2) MulByNonResidueInv(x *E2) *E2 {
	var a E2
	a.A0.SetString("163333666683")
	a.A1.SetString("163333666684")

	z.Mul(x, &a)
	return z
}

// MulBybTwistCurveCoeff multiplies by 2/(1,1)
func (z *E2) MulBybTwistCurveCoeff(x *E2) *E2 {

	z.MulByNonResidueInv(x).Double(z)

	return z
}
