// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package extensions

import (
	fr "github.com/consensys/gnark-crypto/field/koalabear"
)

// E5 is a degree 5 finite field extension of fr
type E5 struct {
	A0, A1, A2, A3, A4 fr.Element
}

// Equal returns true if z equals x, false otherwise
func (z *E5) Equal(x *E5) bool {
	return z.A0.Equal(&x.A0) &&
		z.A1.Equal(&x.A1) &&
		z.A2.Equal(&x.A2) &&
		z.A3.Equal(&x.A3) &&
		z.A4.Equal(&x.A4)
}

// Cmp compares (lexicographic order) z and x and returns:
//
//	-1 if z <  x
//	 0 if z == x
//	+1 if z >  x
func (z *E5) Cmp(x *E5) int {
	if a4 := z.A4.Cmp(&x.A4); a4 != 0 {
		return a4
	}
	if a3 := z.A3.Cmp(&x.A3); a3 != 0 {
		return a3
	}
	if a2 := z.A2.Cmp(&x.A2); a2 != 0 {
		return a2
	}
	if a1 := z.A1.Cmp(&x.A1); a1 != 0 {
		return a1
	}
	return z.A0.Cmp(&x.A0)
}

// String puts E5 elmt in string form
func (z *E5) String() string {
	return (z.A0.String() + "+(" + z.A1.String() + ")*u+(" + z.A2.String() + ")*u**2+(" + z.A3.String() + ")*u**3+(" + z.A4.String() + ")*u**4")
}

// SetString sets a E5 elmt from string
func (z *E5) SetString(s1, s2, s3, s4, s5 string) *E5 {
	z.A0.SetString(s1)
	z.A1.SetString(s2)
	z.A2.SetString(s3)
	z.A3.SetString(s4)
	z.A4.SetString(s5)
	return z
}

// Set copies x into z and returns z
func (z *E5) Set(x *E5) *E5 {
	*z = *x
	return z
}

// SetZero sets z to 0 and returns z
func (z *E5) SetZero() *E5 {
	z.A0.SetZero()
	z.A1.SetZero()
	z.A2.SetZero()
	z.A3.SetZero()
	z.A4.SetZero()
	return z
}

// SetOne sets z to 1 and returns z
func (z *E5) SetOne() *E5 {
	z.A0.SetOne()
	z.A1.SetZero()
	z.A2.SetZero()
	z.A3.SetZero()
	z.A4.SetZero()
	return z
}

// Lift sets the A0 component of z to v
func (z *E5) Lift(v *fr.Element) *E5 {
	*z = E5{}
	z.A0.Set(v)
	return z
}

// MulByElement multiplies an element in E5 by an element in fr
func (z *E5) MulByElement(x *E5, y *fr.Element) *E5 {
	z.A0.Mul(&x.A0, y)
	z.A1.Mul(&x.A1, y)
	z.A2.Mul(&x.A2, y)
	z.A3.Mul(&x.A3, y)
	z.A4.Mul(&x.A4, y)
	return z
}

// Add sets z=x+y in E5 and returns z
func (z *E5) Add(x, y *E5) *E5 {
	z.A0.Add(&x.A0, &y.A0)
	z.A1.Add(&x.A1, &y.A1)
	z.A2.Add(&x.A2, &y.A2)
	z.A3.Add(&x.A3, &y.A3)
	z.A4.Add(&x.A4, &y.A4)
	return z
}

// Sub sets z to x-y and returns z
func (z *E5) Sub(x, y *E5) *E5 {
	z.A0.Sub(&x.A0, &y.A0)
	z.A1.Sub(&x.A1, &y.A1)
	z.A2.Sub(&x.A2, &y.A2)
	z.A3.Sub(&x.A3, &y.A3)
	z.A4.Sub(&x.A4, &y.A4)
	return z
}

// Double sets z=2*x and returns z
func (z *E5) Double(x *E5) *E5 {
	z.A0.Double(&x.A0)
	z.A1.Double(&x.A1)
	z.A2.Double(&x.A2)
	z.A3.Double(&x.A3)
	z.A4.Double(&x.A4)
	return z
}

// Neg negates an E5 element
func (z *E5) Neg(x *E5) *E5 {
	z.A0.Neg(&x.A0)
	z.A1.Neg(&x.A1)
	z.A2.Neg(&x.A2)
	z.A3.Neg(&x.A3)
	z.A4.Neg(&x.A4)
	return z
}

// SetRandom used only in tests
func (z *E5) SetRandom() (*E5, error) {
	if _, err := z.A0.SetRandom(); err != nil {
		return nil, err
	}
	if _, err := z.A1.SetRandom(); err != nil {
		return nil, err
	}
	if _, err := z.A2.SetRandom(); err != nil {
		return nil, err
	}
	if _, err := z.A3.SetRandom(); err != nil {
		return nil, err
	}
	if _, err := z.A4.SetRandom(); err != nil {
		return nil, err
	}
	return z, nil
}

// MustSetRandom sets z to a random value.
// It panics if reading from crypto/rand fails.
func (z *E5) MustSetRandom() *E5 {
	if _, err := z.SetRandom(); err != nil {
		panic(err)
	}
	return z
}

// IsZero returns true if z is zero, false otherwise
func (z *E5) IsZero() bool {
	return z.A0.IsZero() &&
		z.A1.IsZero() &&
		z.A2.IsZero() &&
		z.A3.IsZero() &&
		z.A4.IsZero()
}

// IsOne returns true if z is one, false otherwise
func (z *E5) IsOne() bool {
	return z.A0.IsOne() &&
		z.A1.IsZero() &&
		z.A2.IsZero() &&
		z.A3.IsZero() &&
		z.A4.IsZero()
}

// Mul sets z=x*y in E5 and returns z
func (z *E5) Mul(x, y *E5) *E5 {
	return z.mulMontgomery5(x, y)
}

func (z *E5) mulMontgomery5(x, y *E5) *E5 {
	// Ref.: Peter L. Montgomery. Five, six, and seven-term Karatsuba-like formulae.
	// IEEE Transactions on Computers, 54(3):362–369, 2005.
	//
	// We first compute the interpolation points:
	//
	//		v0 = (a0 + a1 + a2 + a3 + a4)(b0 + b1 + b2 + b3 + b4)
	//		v1 = (a0 - a2 - a3 - a4)(b0 - b2 - b3 - b4)
	//		v2 = (a0 + a1 + a2 - a4)(b0 + b1 + b2 - b4)
	//		v3 = (a0 + a1 - a3 - a4)(b0 + b1 - b3 - b4)
	//		v4 = (a0 - a2 − a3)(b0 - b2 − b3)
	//		v5 = (a1 + a2 - a4)(b1 + b2 - b4)
	//		v6 = (a3 + a4)(b3 + b4)
	//		v7 = (a0 + a1)(b0 + b1)
	//		v8 = (a0 − a4)(b0 − b4)
	//		v9 = a4b4
	//		v10 = a3b3
	//		v11 = a1b1
	//		v12 = a0b0

	var v13, v12, v11, v10, v9, v8, v7, v6, v5, v4, v3, v2, v1, v0 fr.Element
	var t0, t1, t2, t3 fr.Element
	v13.Mul(&x.A0, &y.A0)
	v12.Mul(&x.A1, &y.A1)
	v11.Mul(&x.A3, &y.A3)
	v10.Mul(&x.A4, &y.A4)
	t0.Sub(&x.A0, &x.A4)
	t1.Sub(&y.A0, &y.A4)
	v9.Mul(&t0, &t1)
	t0.Add(&x.A0, &x.A1)
	t1.Add(&y.A0, &y.A1)
	v8.Mul(&t0, &t1)
	t2.Add(&x.A3, &x.A4)
	t3.Add(&y.A3, &y.A4)
	v7.Mul(&t2, &t3)
	v0.Add(&t0, &t2)
	v1.Add(&t1, &t3)
	v0.Mul(&v0, &v1)
	t2.Sub(&t0, &t2)
	t3.Sub(&t1, &t3)
	v4.Mul(&t2, &t3)
	v6.Add(&x.A1, &x.A2).Sub(&v6, &x.A4)
	v5.Add(&y.A1, &y.A2).Sub(&v5, &y.A4)
	v6.Mul(&v6, &v5)
	v5.Sub(&x.A0, &x.A2).Sub(&v5, &x.A3)
	t0.Sub(&y.A0, &y.A2).Sub(&t0, &y.A3)
	v5.Mul(&v5, &t0)
	v3.Add(&x.A0, &x.A1).Add(&v3, &x.A2).Sub(&v3, &x.A4)
	t1.Add(&y.A0, &y.A1).Add(&t1, &y.A2).Sub(&t1, &y.A4)
	v3.Mul(&v3, &t1)
	v2.Sub(&x.A0, &x.A2).Sub(&v2, &x.A3).Sub(&v2, &x.A4)
	t2.Sub(&y.A0, &y.A2).Sub(&t2, &y.A3).Sub(&t2, &y.A4)
	v2.Mul(&v2, &t2)

	//  By re-arranging the formula in function of the degree of X, we obtain
	//  the following expression:
	//
	// 	c0 = v0
	//  c1 = -v0 - v1 + v5
	// 	c2 = v0 + 2v1 + v3 - v4 - v5 - v7 + v10
	// 	c3 = −2v0 − 2v1 − 3v3 + 3v4 − v6 + 2v7 + v9 − 2v10 − v11 + v12
	// 	c4 = 3v0 + v1 + v2 + 3v3 − 4v4 + v5 + v6 − v7 − v8 − 2v9 + 2v10 + 2v11 − v12
	//  c5 = −3v0 − 2v2 − 2v3 + 3v4 − v5 + 2v8 + v9 − v10 − 2v11 + v12
	//  c6 = v0 + 2v2 + v3 − v4 − v6 − v8 + v11
	//  c7 = -v2 - v3 + v6
	//  c8 = v3
	//
	//  Given that 𝔽r⁵[i] = 𝔽r/i⁵-i²+1, we replace X⁵ by X²-1
	//
	//  c0 <-- c0 - c5 - c8
	//  c1 <-- c1 - c6
	//  c2 <-- c2 + c5 - c7 + c8
	//  c3 <-- c3 + c6 - c8
	//  c4 <-- c4 + c7
	//
	//  So the final formula for x*y is:
	//
	//  c0 = 4v0 + 2v2 + v3 - 3v4 + v5 - 2v8 - v9 + v10 + 2v11 - v12
	//     = 2(2v0 + v2 + v11 - v8) + v3 + v5 + v10 - v9 - v12 - 3v4
	//  c1 = v4 + v5 + v6 + v8 - v11 - v1 - 2v0
	//  c2 = 2v0 + 3v1 + 4v3 - 4v4 + 2v5 - v6 + v7 - v10 + v12
	//     = 2(v0 + v5 + 2v3 - 2v4) + 3v1 + v7 + v12 - v6 - v10
	//  c3 = -v0 -2v1 + 2v2 - 2v3 + 2v4 - 2v6 + 2v7 + v8 + v9 - 2v10 - 2v11 + v12
	//     = 2(v2 + v4 + v7 - v6 - v1 - v3 - v10 - v11) + v12 + v8 + v9 - v0
	//  c4 = 3v0 + v1 + 2v3 − 4v4 + v5 + 2v6 − v7 − v8 − 2v9 + 2v10 + 2v11 − v12
	//     = 3v0 + 2(v3 + v6 + v10 + v11 − v9 − 2v4) + v5 + v1 − v12 − v7 − v8
	var c0, c1, c2, c3, c4 fr.Element
	c0.Double(&v0).
		Add(&c0, &v2).
		Add(&c0, &v11).
		Sub(&c0, &v8).
		Double(&c0).
		Add(&c0, &v3).
		Add(&c0, &v5).
		Add(&c0, &v10).
		Sub(&c0, &v9).
		Add(&c0, &v12)
	t0 = v4
	fr.MulBy3(&t0)
	c0.Sub(&c0, &t0)

	c1.Add(&v4, &v5).
		Add(&c1, &v6).
		Add(&c1, &v8).
		Sub(&c1, &v11).
		Sub(&c1, &v1).
		Sub(&c1, &v0).
		Sub(&c1, &v0)

	c2.Add(&v0, &v5)
	t0.Sub(&v3, &v4).Double(&t0)
	c2.Add(&c2, &t0).
		Double(&c2).
		Add(&c2, &v7).
		Add(&c2, &v12).
		Sub(&c2, &v6).
		Sub(&c2, &v10)
	t0 = v1
	fr.MulBy3(&t0)
	c2.Add(&c2, &t0)

	c3.Add(&v2, &v4).
		Add(&c3, &v7).
		Sub(&c3, &v6).
		Sub(&c3, &v1).
		Sub(&c3, &v3).
		Sub(&c3, &v10).
		Sub(&c3, &v11).
		Double(&c3).
		Add(&c3, &v12).
		Add(&c3, &v8).
		Add(&c3, &v9).
		Sub(&c3, &v0)

	//     = 3v0 + 2(v3 + v6 + v10 + v11 − v9 − 2v4) + v5 + v1 − v12 − v7 − v8
	c4.Add(&v3, &v6).
		Add(&c4, &v10).
		Add(&c4, &v11).
		Sub(&c4, &v9).
		Sub(&c4, &v4).Sub(&c4, &v4).
		Double(&c4).
		Add(&c4, &v5).
		Add(&c4, &v1).
		Sub(&c4, &v12).
		Sub(&c4, &v7).
		Sub(&c4, &v8)
	t0 = v0
	fr.MulBy3(&t0)
	c4.Add(&c4, &t0)

	z.A0 = c0
	z.A1 = c1
	z.A2 = c2
	z.A3 = c3
	z.A4 = c4

	return z
}

// Square sets z=x*x in E5 and returns z
func (z *E5) Square(x *E5) *E5 {
	return z.Mul(x, x)
}

// Inverse sets z to the inverse of x in E5 and returns z
//
// if x == 0, sets and returns z = x
func (z *E5) Inverse(x *E5) *E5 {
	// Implement
	return z
}
