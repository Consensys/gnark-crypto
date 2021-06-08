// Copyright 2020 ConsenSys Software Inc.
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
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls24-315/fp"
)

// E8 is a degree two finite field extension of fp4
type E8 struct {
	C0, C1 E4
}

// Equal returns true if z equals x, fasle otherwise
func (z *E8) Equal(x *E8) bool {
	return z.C0.Equal(&x.C0) && z.C1.Equal(&x.C1)
}

// Cmp compares (lexicographic order) z and x and returns:
//
//   -1 if z <  x
//    0 if z == x
//   +1 if z >  x
//
func (z *E8) Cmp(x *E8) int {
	if a1 := z.C1.Cmp(&x.C1); a1 != 0 {
		return a1
	}
	return z.C0.Cmp(&x.C0)
}

// LexicographicallyLargest returns true if this element is strictly lexicographically
// larger than its negation, false otherwise
func (z *E8) LexicographicallyLargest() bool {
	// adapted from github.com/zkcrypto/bls12_381
	if z.C1.IsZero() {
		return z.C0.LexicographicallyLargest()
	}
	return z.C1.LexicographicallyLargest()
}

// String puts E8 in string form
func (z *E8) String() string {
	return (z.C0.String() + "+(" + z.C1.String() + ")*w")
}

// SetString sets a E8 from string
func (z *E8) SetString(s0, s1, s2, s3, s4, s5, s6, s7 string) *E8 {
	z.C0.SetString(s0, s1, s2, s3)
	z.C1.SetString(s4, s5, s6, s7)
	return z
}

// Set copies x into z and returns z
func (z *E8) Set(x *E8) *E8 {
	z.C0 = x.C0
	z.C1 = x.C1
	return z
}

// SetZero sets an E8 elmt to zero
func (z *E8) SetZero() *E8 {
	z.C0.SetZero()
	z.C1.SetZero()
	return z
}

// SetOne sets z to 1 in Montgomery form and returns z
func (z *E8) SetOne() *E8 {
	*z = E8{}
	z.C0.B0.A0.SetOne()
	return z
}

// ToMont converts to Mont form
func (z *E8) ToMont() *E8 {
	z.C0.ToMont()
	z.C1.ToMont()
	return z
}

// FromMont converts from Mont form
func (z *E8) FromMont() *E8 {
	z.C0.FromMont()
	z.C1.FromMont()
	return z
}

// MulByElement multiplies an element in E8 by an element in fp
func (z *E8) MulByElement(x *E8, y *fp.Element) *E8 {
	var yCopy fp.Element
	yCopy.Set(y)
	z.C0.MulByElement(&x.C0, &yCopy)
	z.C1.MulByElement(&x.C1, &yCopy)
	return z
}

// Add set z=x+y in E8 and return z
func (z *E8) Add(x, y *E8) *E8 {
	z.C0.Add(&x.C0, &y.C0)
	z.C1.Add(&x.C1, &y.C1)
	return z
}

// Sub sets z to x sub y and return z
func (z *E8) Sub(x, y *E8) *E8 {
	z.C0.Sub(&x.C0, &y.C0)
	z.C1.Sub(&x.C1, &y.C1)
	return z
}

// Double sets z=2*x and returns z
func (z *E8) Double(x *E8) *E8 {
	z.C0.Double(&x.C0)
	z.C1.Double(&x.C1)
	return z
}

// Neg negates an E8 element
func (z *E8) Neg(x *E8) *E8 {
	z.C0.Neg(&x.C0)
	z.C1.Neg(&x.C1)
	return z
}

// SetRandom used only in tests
func (z *E8) SetRandom() (*E8, error) {
	if _, err := z.C0.SetRandom(); err != nil {
		return nil, err
	}
	if _, err := z.C1.SetRandom(); err != nil {
		return nil, err
	}
	return z, nil
}

// IsZero returns true if the element is zero, fasle otherwise
func (z *E8) IsZero() bool {
	return z.C0.IsZero() && z.C1.IsZero()
}

// MulByNonResidue mul x by (0,1)
func (z *E8) MulByNonResidue(x *E8) *E8 {
	z.C1, z.C0 = x.C0, x.C1
	z.C0.MulByNonResidue(&z.C0)
	return z
}

// Mul set z=x*y in E8 and return z
func (z *E8) Mul(x, y *E8) *E8 {
	var a, b, c E4
	a.Add(&x.C0, &x.C1)
	b.Add(&y.C0, &y.C1)
	a.Mul(&a, &b)
	b.Mul(&x.C0, &y.C0)
	c.Mul(&x.C1, &y.C1)
	z.C1.Sub(&a, &b).Sub(&z.C1, &c)
	z.C0.MulByNonResidue(&c).Add(&z.C0, &b)
	return z
}

// Square set z=x*x in E8 and return z
func (z *E8) Square(x *E8) *E8 {

	//Algorithm 22 from https://eprint.iacr.org/2010/354.pdf
	c0 := x.C0
	c1 := x.C1

	var c2 E4
	c2.Mul(&c0, &c1)
	c0.Sub(&c0, &c1)

	z.C1.Double(&c2)

	c1.MulByNonResidue(&c1)
	c1.Sub(&x.C0, &c1)

	c0.Mul(&c0, &c1)
	c0.Add(&c0, &c2)

	c2.MulByNonResidue(&c2)
	z.C0.Add(&c0, &c2)

	return z
}

// Inverse set z to the inverse of x in E8 and return z
func (z *E8) Inverse(x *E8) *E8 {
	// Algorithm 23 from https://eprint.iacr.org/2010/354.pdf

	var t0, t1, tmp E4
	t0.Square(&x.C0)
	t1.Square(&x.C1)
	tmp.MulByNonResidue(&t1)
	t0.Sub(&t0, &tmp)
	t1.Inverse(&t0)
	z.C0.Mul(&x.C0, &t1)
	z.C1.Mul(&x.C1, &t1).Neg(&z.C1)

	return z
}

// Exp sets z=x**e and returns it
func (z *E8) Exp(x *E8, e big.Int) *E8 {
	var res E8
	res.SetOne()
	b := e.Bytes()
	for i := range b {
		w := b[i]
		mask := byte(0x80)
		for j := 7; j >= 0; j-- {
			res.Square(&res)
			if (w&mask)>>j != 0 {
				res.Mul(&res, x)
			}
			mask = mask >> 1
		}
	}
	z.Set(&res)
	return z
}

// Conjugate set z to x conjugated and return z
func (z *E8) Conjugate(x *E8) *E8 {
	z.C0 = x.C0
	z.C1.Neg(&x.C1)
	return z
}
