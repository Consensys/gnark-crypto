// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fptower

import (
	"github.com/consensys/gnark-crypto/ecc/bw6-633/fp"
)

// E3 is a degree-three finite field extension of fp2
type E3 struct {
	A0, A1, A2 fp.Element
}

// Equal returns true if z equals x, false otherwise
// TODO can this be deleted?  Should be able to use == operator instead
func (z *E3) Equal(x *E3) bool {
	return z.A0.Equal(&x.A0) && z.A1.Equal(&x.A1) && z.A2.Equal(&x.A2)
}

// SetString sets a E3 elmt from stringf
func (z *E3) SetString(s1, s2, s3 string) *E3 {
	z.A0.SetString(s1)
	z.A1.SetString(s2)
	z.A2.SetString(s3)
	return z
}

// SetZero sets an E3 elmt to zero
func (z *E3) SetZero() *E3 {
	z.A0.SetZero()
	z.A1.SetZero()
	z.A2.SetZero()
	return z
}

// Clone returns a copy of self
func (z *E3) Clone() *E3 {
	return &E3{
		A0: z.A0,
		A1: z.A1,
		A2: z.A2,
	}
}

// Set Sets a E3 elmt form another E3 elmt
func (z *E3) Set(x *E3) *E3 {
	z.A0 = x.A0
	z.A1 = x.A1
	z.A2 = x.A2
	return z
}

// SetOne sets z to 1 in Montgomery form and returns z
func (z *E3) SetOne() *E3 {
	z.A0.SetOne()
	z.A1.SetZero()
	z.A2.SetZero()
	return z
}

// SetRandom sets z to a uniform random value
func (z *E3) SetRandom() (*E3, error) {
	if _, err := z.A0.SetRandom(); err != nil {
		return nil, err
	}
	if _, err := z.A1.SetRandom(); err != nil {
		return nil, err
	}
	if _, err := z.A2.SetRandom(); err != nil {
		return nil, err
	}
	return z, nil
}

// MustSetRandom sets z to a uniform random value.
// Panics if reading from crypto/rand fails.
func (z *E3) MustSetRandom() *E3 {
	if _, err := z.SetRandom(); err != nil {
		panic(err)
	}
	return z
}

// IsZero returns true if z is zero, false otherwise
func (z *E3) IsZero() bool {
	return z.A0.IsZero() && z.A1.IsZero() && z.A2.IsZero()
}

// IsOne returns true if z is one, false otherwise
func (z *E3) IsOne() bool {
	return z.A0.IsOne() && z.A1.IsZero() && z.A2.IsZero()
}

// Neg negates the E3 number
func (z *E3) Neg(x *E3) *E3 {
	z.A0.Neg(&x.A0)
	z.A1.Neg(&x.A1)
	z.A2.Neg(&x.A2)
	return z
}

// Add adds two elements of E3
func (z *E3) Add(x, y *E3) *E3 {
	z.A0.Add(&x.A0, &y.A0)
	z.A1.Add(&x.A1, &y.A1)
	z.A2.Add(&x.A2, &y.A2)
	return z
}

// Sub subtracts two elements of E3
func (z *E3) Sub(x, y *E3) *E3 {
	z.A0.Sub(&x.A0, &y.A0)
	z.A1.Sub(&x.A1, &y.A1)
	z.A2.Sub(&x.A2, &y.A2)
	return z
}

// Double doubles an element in E3
func (z *E3) Double(x *E3) *E3 {
	z.A0.Double(&x.A0)
	z.A1.Double(&x.A1)
	z.A2.Double(&x.A2)
	return z
}

// String puts E3 elmt in string form
func (z *E3) String() string {
	return (z.A0.String() + "+(" + z.A1.String() + ")*u+(" + z.A2.String() + ")*u**2")
}

// MulByElement multiplies an element in E3 by an element in fp
func (z *E3) MulByElement(x *E3, y *fp.Element) *E3 {
	var yCopy fp.Element
	yCopy.Set(y)
	z.A0.Mul(&x.A0, &yCopy)
	z.A1.Mul(&x.A1, &yCopy)
	z.A2.Mul(&x.A2, &yCopy)
	return z
}

// MulBy12 multiplication by sparse element (0,b1,b2)
func (z *E3) MulBy12(b1, b2 *fp.Element) *E3 {
	var t1, t2, c0, tmp, c1, c2 fp.Element
	t1.Mul(&z.A1, b1)
	t2.Mul(&z.A2, b2)
	c0.Add(&z.A1, &z.A2)
	tmp.Add(b1, b2)
	c0.Mul(&c0, &tmp)
	c0.Sub(&c0, &t1)
	c0.Sub(&c0, &t2)
	c0.MulByNonResidue(&c0)
	c1.Add(&z.A0, &z.A1)
	c1.Mul(&c1, b1)
	c1.Sub(&c1, &t1)
	tmp.MulByNonResidue(&t2)
	c1.Add(&c1, &tmp)
	tmp.Add(&z.A0, &z.A2)
	c2.Mul(b2, &tmp)
	c2.Sub(&c2, &t2)
	c2.Add(&c2, &t1)

	z.A0 = c0
	z.A1 = c1
	z.A2 = c2

	return z
}

// MulBy01 multiplication by sparse element (c0,c1,0)
func (z *E3) MulBy01(c0, c1 *fp.Element) *E3 {

	var a, b, tmp, t0, t1, t2 fp.Element

	a.Mul(&z.A0, c0)
	b.Mul(&z.A1, c1)

	tmp.Add(&z.A1, &z.A2)
	t0.Mul(c1, &tmp)
	t0.Sub(&t0, &b)
	t0.MulByNonResidue(&t0)
	t0.Add(&t0, &a)

	tmp.Add(&z.A0, &z.A2)
	t2.Mul(c0, &tmp)
	t2.Sub(&t2, &a)
	t2.Add(&t2, &b)

	t1.Add(c0, c1)
	tmp.Add(&z.A0, &z.A1)
	t1.Mul(&t1, &tmp)
	t1.Sub(&t1, &a)
	t1.Sub(&t1, &b)

	z.A0.Set(&t0)
	z.A1.Set(&t1)
	z.A2.Set(&t2)

	return z
}

// MulBy1 multiplication of E6 by sparse element (0, c1, 0)
func (z *E3) MulBy1(c1 *fp.Element) *E3 {

	var b, tmp, t0, t1 fp.Element
	b.Mul(&z.A1, c1)

	tmp.Add(&z.A1, &z.A2)
	t0.Mul(c1, &tmp)
	t0.Sub(&t0, &b)
	t0.MulByNonResidue(&t0)

	tmp.Add(&z.A0, &z.A1)
	t1.Mul(c1, &tmp)
	t1.Sub(&t1, &b)

	z.A0.Set(&t0)
	z.A1.Set(&t1)
	z.A2.Set(&b)

	return z
}

// Mul sets z to the E3-product of x,y, returns z
func (z *E3) Mul(x, y *E3) *E3 {
	// Karatsuba method for cubic extensions
	// https://eprint.iacr.org/2006/471.pdf (section 4)
	var t0, t1, t2, c0, c1, c2, tmp fp.Element
	t0.Mul(&x.A0, &y.A0)
	t1.Mul(&x.A1, &y.A1)
	t2.Mul(&x.A2, &y.A2)

	c0.Add(&x.A1, &x.A2)
	tmp.Add(&y.A1, &y.A2)
	c0.Mul(&c0, &tmp).Sub(&c0, &t1).Sub(&c0, &t2).MulByNonResidue(&c0).Add(&c0, &t0)

	c1.Add(&x.A0, &x.A1)
	tmp.Add(&y.A0, &y.A1)
	c1.Mul(&c1, &tmp).Sub(&c1, &t0).Sub(&c1, &t1)
	tmp.MulByNonResidue(&t2)
	c1.Add(&c1, &tmp)

	tmp.Add(&x.A0, &x.A2)
	c2.Add(&y.A0, &y.A2).Mul(&c2, &tmp).Sub(&c2, &t0).Sub(&c2, &t2).Add(&c2, &t1)

	z.A0.Set(&c0)
	z.A1.Set(&c1)
	z.A2.Set(&c2)

	return z
}

// MulAssign sets z to the E3-product of z,y, returns z
func (z *E3) MulAssign(x *E3) *E3 {
	z.Mul(z, x)
	return z
}

// Square sets z to the E3-product of x,x, returns z
func (z *E3) Square(x *E3) *E3 {

	// Algorithm 16 from https://eprint.iacr.org/2010/354.pdf
	var c4, c5, c1, c2, c3, c0 fp.Element
	c4.Mul(&x.A0, &x.A1).Double(&c4)
	c5.Square(&x.A2)
	c1.MulByNonResidue(&c5).Add(&c1, &c4)
	c2.Sub(&c4, &c5)
	c3.Square(&x.A0)
	c4.Sub(&x.A0, &x.A1).Add(&c4, &x.A2)
	c5.Mul(&x.A1, &x.A2).Double(&c5)
	c4.Square(&c4)
	c0.MulByNonResidue(&c5).Add(&c0, &c3)
	z.A2.Add(&c2, &c4).Add(&z.A2, &c5).Sub(&z.A2, &c3)
	z.A0.Set(&c0)
	z.A1.Set(&c1)

	return z
}

// MulByNonResidue mul x by (0,1,0)
func (z *E3) MulByNonResidue(x *E3) *E3 {
	z.A2, z.A1, z.A0 = x.A1, x.A0, x.A2
	z.A0.MulByNonResidue(&z.A0)
	return z
}

// Inverse an element in E3
//
// if x == 0, sets and returns z = x
func (z *E3) Inverse(x *E3) *E3 {
	// Algorithm 17 from https://eprint.iacr.org/2010/354.pdf
	// step 9 is wrong in the paper it's t1-t4
	var t0, t1, t2, t3, t4, t5, t6, c0, c1, c2, d1, d2 fp.Element
	t0.Square(&x.A0)
	t1.Square(&x.A1)
	t2.Square(&x.A2)
	t3.Mul(&x.A0, &x.A1)
	t4.Mul(&x.A0, &x.A2)
	t5.Mul(&x.A1, &x.A2)
	c0.MulByNonResidue(&t5).Neg(&c0).Add(&c0, &t0)
	c1.MulByNonResidue(&t2).Sub(&c1, &t3)
	c2.Sub(&t1, &t4)
	t6.Mul(&x.A0, &c0)
	d1.Mul(&x.A2, &c1)
	d2.Mul(&x.A1, &c2)
	d1.Add(&d1, &d2).MulByNonResidue(&d1)
	t6.Add(&t6, &d1)
	t6.Inverse(&t6)
	z.A0.Mul(&c0, &t6)
	z.A1.Mul(&c1, &t6)
	z.A2.Mul(&c2, &t6)

	return z
}

// BatchInvertE3 returns a new slice with every element in a inverted.
// It uses Montgomery batch inversion trick.
//
// if a[i] == 0, returns result[i] = a[i]
func BatchInvertE3(a []E3) []E3 {
	res := make([]E3, len(a))
	if len(a) == 0 {
		return res
	}

	zeroes := make([]bool, len(a))
	var accumulator E3
	accumulator.SetOne()

	for i := 0; i < len(a); i++ {
		if a[i].IsZero() {
			zeroes[i] = true
			continue
		}
		res[i].Set(&accumulator)
		accumulator.Mul(&accumulator, &a[i])
	}

	accumulator.Inverse(&accumulator)

	for i := len(a) - 1; i >= 0; i-- {
		if zeroes[i] {
			continue
		}
		res[i].Mul(&res[i], &accumulator)
		accumulator.Mul(&accumulator, &a[i])
	}

	return res
}
