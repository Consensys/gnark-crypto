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

package bw761

import (
	"github.com/consensys/gurvy/bw761/fp"
)

// e2 is a degree-two finite field extension of fp.Element
type e2 struct {
	A0, A1 fp.Element
}

// Equal returns true if z equals x, fasle otherwise
func (z *e2) Equal(x *e2) bool {
	return z.A0.Equal(&x.A0) && z.A1.Equal(&x.A1)
}

// SetString sets a e2 element from strings
func (z *e2) SetString(s1, s2 string) *e2 {
	z.A0.SetString(s1)
	z.A1.SetString(s2)
	return z
}

// SetZero sets an e2 elmt to zero
func (z *e2) SetZero() *e2 {
	z.A0.SetZero()
	z.A1.SetZero()
	return z
}

// Clone returns a copy of self
func (z *e2) Clone() *e2 {
	return &e2{
		A0: z.A0,
		A1: z.A1,
	}
}

// Set sets an e2 from x
func (z *e2) Set(x *e2) *e2 {
	z.A0.Set(&x.A0)
	z.A1.Set(&x.A1)
	return z
}

// SetOne sets z to 1 in Montgomery form and returns z
func (z *e2) SetOne() *e2 {
	z.A0.SetOne()
	z.A1.SetZero()
	return z
}

// SetRandom sets a0 and a1 to random values
func (z *e2) SetRandom() *e2 {
	z.A0.SetRandom()
	z.A1.SetRandom()
	return z
}

// IsZero returns true if the two elements are equal, fasle otherwise
func (z *e2) IsZero() bool {
	return z.A0.IsZero() && z.A1.IsZero()
}

// Neg negates an e2 element
func (z *e2) Neg(x *e2) *e2 {
	z.A0.Neg(&x.A0)
	z.A1.Neg(&x.A1)
	return z
}

// String implements Stringer interface for fancy printing
func (z *e2) String() string {
	return (z.A0.String() + "+" + z.A1.String() + "*u")
}

// ToMont converts to mont form
func (z *e2) ToMont() *e2 {
	z.A0.ToMont()
	z.A1.ToMont()
	return z
}

// FromMont converts from mont form
func (z *e2) FromMont() *e2 {
	z.A0.FromMont()
	z.A1.FromMont()
	return z
}

// Add adds two elements of e2
func (z *e2) Add(x, y *e2) *e2 {
	z.A0.Add(&x.A0, &y.A0)
	z.A1.Add(&x.A1, &y.A1)
	return z
}

// Sub two elements of e2
func (z *e2) Sub(x, y *e2) *e2 {
	z.A0.Sub(&x.A0, &y.A0)
	z.A1.Sub(&x.A1, &y.A1)
	return z
}

// Double doubles an e2 element
func (z *e2) Double(x *e2) *e2 {
	z.A0.Double(&x.A0)
	z.A1.Double(&x.A1)
	return z
}

// Mul sets z to the e2-product of x,y, returns z
func (z *e2) Mul(x, y *e2) *e2 {
	var a, b, c fp.Element
	a.Add(&x.A0, &x.A1)
	b.Add(&y.A0, &y.A1)
	a.Mul(&a, &b)
	b.Mul(&x.A0, &y.A0)
	c.Mul(&x.A1, &y.A1)
	z.A1.Sub(&a, &b).Sub(&z.A1, &c)
	z.A0.Double(&c).Double(&z.A0).Neg(&z.A0).Add(&z.A0, &b)
	return z
}

// MulAssign sets z to the e2-product of z,x returns z
func (z *e2) MulAssign(x *e2) *e2 {
	var t e2
	t.Mul(z, x)
	z.Set(&t)
	return z
}

// Square sets z to the e2-product of x,x returns z
func (z *e2) Square(x *e2) *e2 {
	// algo 22 https://eprint.iacr.org/2010/354.pdf
	var c0, c2 fp.Element
	c2.Double(&x.A1).Double(&c2).Neg(&c2).AddAssign(&x.A0)
	c0.Add(&x.A0, &x.A1)
	c0.Mul(&c0, &c2) // (x1+x2)*(x1+(u**2)x2) = x1**2+(u**2)x2**2+(u**2+1)x1x2
	c2.Mul(&x.A0, &x.A1)
	z.A1.Double(&c2)
	z.A0.Add(&c0, &z.A1).AddAssign(&c2)
	return z
}

// MulByNonResidue multiplies a e2 by (0,1)
func (z *e2) MulByNonResidue(x *e2) *e2 {
	a := x.A0
	b := x.A1 // fetching x.A1 in the function below is slower
	z.A0.Double(&b).Double(&z.A0).Neg(&z.A0)
	z.A1 = a
	return z
}

// Inverse sets z to the e2-inverse of x, returns z
func (z *e2) Inverse(x *e2) *e2 {
	// Algorithm 8 from https://eprint.iacr.org/2010/354.pdf
	var t0, t1, tmp fp.Element
	t0.Square(&x.A0)
	t1.Square(&x.A1)
	tmp.Double(&t1).Double(&tmp).Neg(&tmp)
	t0.Sub(&t0, &tmp)
	t1.Inverse(&t0)
	z.A0.Mul(&x.A0, &t1)
	z.A1.Mul(&x.A1, &t1).Neg(&z.A1)

	return z
}

// MulByElement multiplies an element in e2 by an element in fp
func (z *e2) MulByElement(x *e2, y *fp.Element) *e2 {
	var yCopy fp.Element
	yCopy.Set(y)
	z.A0.Mul(&x.A0, &yCopy)
	z.A1.Mul(&x.A1, &yCopy)
	return z
}

// Conjugate conjugates an element in e2
func (z *e2) Conjugate(x *e2) *e2 {
	z.A0.Set(&x.A0)
	z.A1.Neg(&x.A1)
	return z
}
