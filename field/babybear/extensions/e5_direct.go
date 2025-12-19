// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package extensions

import fr "github.com/consensys/gnark-crypto/field/babybear"

// E5D is a degree 5 finite field extension of fr
type E5D struct {
	A0, A1, A2, A3, A4 fr.Element
}

// Equal returns true if z equals x, false otherwise.
func (z *E5D) Equal(x *E5D) bool {
	return z.A0.Equal(&x.A0) &&
		z.A1.Equal(&x.A1) &&
		z.A2.Equal(&x.A2) &&
		z.A3.Equal(&x.A3) &&
		z.A4.Equal(&x.A4)
}

// String puts E5D elmt in string form.
func (z *E5D) String() string {
	return (z.A0.String() + "+(" + z.A1.String() + ")*u+(" + z.A2.String() + ")*u**2+(" + z.A3.String() + ")*u**3+(" + z.A4.String() + ")*u**4")
}

// SetString sets a E5D elmt from string.
func (z *E5D) SetString(s1, s2, s3, s4, s5 string) *E5D {
	z.A0.SetString(s1)
	z.A1.SetString(s2)
	z.A2.SetString(s3)
	z.A3.SetString(s4)
	z.A4.SetString(s5)
	return z
}

// Set copies x into z and returns z.
func (z *E5D) Set(x *E5D) *E5D {
	*z = *x
	return z
}

// SetOne sets z to 1 and returns z.
func (z *E5D) SetOne() *E5D {
	z.A0.SetOne()
	z.A1.SetZero()
	z.A2.SetZero()
	z.A3.SetZero()
	z.A4.SetZero()
	return z
}

// Add sets z=x+y in E5D and returns z.
func (z *E5D) Add(x, y *E5D) *E5D {
	z.A0.Add(&x.A0, &y.A0)
	z.A1.Add(&x.A1, &y.A1)
	z.A2.Add(&x.A2, &y.A2)
	z.A3.Add(&x.A3, &y.A3)
	z.A4.Add(&x.A4, &y.A4)
	return z
}

// Sub sets z to x-y and returns z.
func (z *E5D) Sub(x, y *E5D) *E5D {
	z.A0.Sub(&x.A0, &y.A0)
	z.A1.Sub(&x.A1, &y.A1)
	z.A2.Sub(&x.A2, &y.A2)
	z.A3.Sub(&x.A3, &y.A3)
	z.A4.Sub(&x.A4, &y.A4)
	return z
}

// Double sets z=2*x and returns z.
func (z *E5D) Double(x *E5D) *E5D {
	z.A0.Double(&x.A0)
	z.A1.Double(&x.A1)
	z.A2.Double(&x.A2)
	z.A3.Double(&x.A3)
	z.A4.Double(&x.A4)
	return z
}

// SetRandom used only in tests.
func (z *E5D) SetRandom() (*E5D, error) {
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
func (z *E5D) MustSetRandom() *E5D {
	if _, err := z.SetRandom(); err != nil {
		panic(err)
	}
	return z
}

// IsZero returns true if z is zero, false otherwise.
func (z *E5D) IsZero() bool {
	return z.A0.IsZero() && z.A1.IsZero() && z.A2.IsZero() && z.A3.IsZero() && z.A4.IsZero()
}

// IsOne returns true if z is one, false otherwise.
func (z *E5D) IsOne() bool {
	return z.A0.IsOne() && z.A1.IsZero() && z.A2.IsZero() && z.A3.IsZero() && z.A4.IsZero()
}

// Mul sets z=x*y in E5D and returns z.
func (z *E5D) Mul(x, y *E5D) *E5D {
	return z.mulMontgomery5(x, y)
}

// mulNaive is the schoolbook multiplication followed by reduction.
func (z *E5D) mulNaive(a, b *E5D) *E5D {
	var d [9]fr.Element
	var t fr.Element

	// degree 0
	d[0].Mul(&a.A0, &b.A0)

	// degree 1
	d[1].Mul(&a.A0, &b.A1)
	t.Mul(&a.A1, &b.A0)
	d[1].Add(&d[1], &t)

	// degree 2
	d[2].Mul(&a.A0, &b.A2)
	t.Mul(&a.A1, &b.A1)
	d[2].Add(&d[2], &t)
	t.Mul(&a.A2, &b.A0)
	d[2].Add(&d[2], &t)

	// degree 3
	d[3].Mul(&a.A0, &b.A3)
	t.Mul(&a.A1, &b.A2)
	d[3].Add(&d[3], &t)
	t.Mul(&a.A2, &b.A1)
	d[3].Add(&d[3], &t)
	t.Mul(&a.A3, &b.A0)
	d[3].Add(&d[3], &t)

	// degree 4
	d[4].Mul(&a.A0, &b.A4)
	t.Mul(&a.A1, &b.A3)
	d[4].Add(&d[4], &t)
	t.Mul(&a.A2, &b.A2)
	d[4].Add(&d[4], &t)
	t.Mul(&a.A3, &b.A1)
	d[4].Add(&d[4], &t)
	t.Mul(&a.A4, &b.A0)
	d[4].Add(&d[4], &t)

	// degree 5
	d[5].Mul(&a.A1, &b.A4)
	t.Mul(&a.A2, &b.A3)
	d[5].Add(&d[5], &t)
	t.Mul(&a.A3, &b.A2)
	d[5].Add(&d[5], &t)
	t.Mul(&a.A4, &b.A1)
	d[5].Add(&d[5], &t)

	// degree 6
	d[6].Mul(&a.A2, &b.A4)
	t.Mul(&a.A3, &b.A3)
	d[6].Add(&d[6], &t)
	t.Mul(&a.A4, &b.A2)
	d[6].Add(&d[6], &t)

	// degree 7
	d[7].Mul(&a.A3, &b.A4)
	t.Mul(&a.A4, &b.A3)
	d[7].Add(&d[7], &t)

	// degree 8
	d[8].Mul(&a.A4, &b.A4)

	return z.reduceDegree8(&d)
}

// mulMontgomery5 implements Montgomery's 5-term formula
//
// Ref.: Peter L. Montgomery. Five, six, and seven-term Karatsuba-like formulae. IEEE
// Transactions on Computers, 54(3):362–369, 2005.
func (z *E5D) mulMontgomery5(a, b *E5D) *E5D {
	// -------------------------------------------------------------------------
	// Phase 1: evaluation points (13 muls in fr)
	// -------------------------------------------------------------------------
	var u [13]fr.Element
	var t [12]fr.Element

	// Common sums/diffs
	t[0].Add(&a.A0, &a.A1)  // a0+a1
	t[1].Add(&b.A0, &b.A1)  // b0+b1
	t[2].Add(&a.A2, &a.A3)  // a2+a3
	t[3].Add(&b.A2, &b.A3)  // b2+b3
	t[4].Add(&t[2], &a.A4)  // a2+a3+a4
	t[5].Add(&t[3], &b.A4)  // b2+b3+b4
	t[6].Add(&a.A3, &a.A4)  // a3+a4
	t[7].Add(&b.A3, &b.A4)  // b3+b4
	t[8].Add(&t[0], &t[4])  // a0+a1+a2+a3+a4
	t[9].Add(&t[1], &t[5])  // b0+b1+b2+b3+b4
	t[10].Add(&t[0], &a.A2) // a0+a1+a2
	t[11].Add(&t[1], &b.A2) // b0+b1+b2

	u[12].Mul(&t[8], &t[9]) // (a0+a1+a2+a3+a4)(b0+b1+b2+b3+b4)

	u[11].Sub(&a.A0, &t[4]) // a0-(a2+a3+a4)
	t[4].Sub(&b.A0, &t[5])  // b0-(b2+b3+b4)
	u[11].Mul(&u[11], &t[4])

	t[4].Set(&t[10])
	t[4].Sub(&t[4], &a.A4) // a0+a1+a2-a4
	t[5].Set(&t[11])
	t[5].Sub(&t[5], &b.A4) // b0+b1+b2-b4
	u[10].Mul(&t[4], &t[5])

	t[4].Sub(&t[0], &t[6]) // a0+a1-(a3+a4)
	t[5].Sub(&t[1], &t[7]) // b0+b1-(b3+b4)
	u[9].Mul(&t[4], &t[5])

	t[4].Sub(&a.A0, &t[2]) // a0-(a2+a3)
	t[5].Sub(&b.A0, &t[3]) // b0-(b2+b3)
	u[8].Mul(&t[4], &t[5])

	t[4].Add(&a.A1, &a.A2) // a1+a2
	t[4].Sub(&t[4], &a.A4) // a1+a2-a4
	t[5].Add(&b.A1, &b.A2) // b1+b2
	t[5].Sub(&t[5], &b.A4) // b1+b2-b4
	u[7].Mul(&t[4], &t[5])

	u[6].Mul(&t[6], &t[7]) // (a3+a4)(b3+b4)
	u[5].Mul(&t[0], &t[1]) // (a0+a1)(b0+b1)

	t[4].Sub(&a.A0, &a.A4) // a0-a4
	t[5].Sub(&b.A0, &b.A4) // b0-b4
	u[4].Mul(&t[4], &t[5])

	u[3].Mul(&a.A4, &b.A4)
	u[2].Mul(&a.A3, &b.A3)
	u[1].Mul(&a.A1, &b.A1)
	u[0].Mul(&a.A0, &b.A0)

	// -------------------------------------------------------------------------
	// Phase 2: reconstruction of degree-8 coefficients (only adds/subs)
	// -------------------------------------------------------------------------
	// Shared helpers:
	// t0 = u0 + u3 - u4
	// t1 = u6 - u2
	// t2 = u2 - u8 + u11
	// t3 = u9 - u10 - u11
	// t4 = u1 - u7 + u10
	var t0, t1, t2, t3, t4 fr.Element
	t0.Add(&u[0], &u[3])
	t0.Sub(&t0, &u[4])

	t1.Sub(&u[6], &u[2])

	t2.Sub(&u[2], &u[8])
	t2.Add(&t2, &u[11])

	t3.Sub(&u[9], &u[10])
	t3.Sub(&t3, &u[11])

	t4.Sub(&u[1], &u[7])
	t4.Add(&t4, &u[10])

	var d [9]fr.Element
	// d8
	d[8].Set(&u[3])

	// d7 = t1 - u3
	d[7].Sub(&t1, &u[3])

	// d6 = t0 - t1 + t2
	d[6].Sub(&t0, &t1)
	d[6].Add(&d[6], &t2)

	// d5 = u3 - u5 + u9 - u10 + u12 - 2*t2 - 3*t0
	d[5].Sub(&u[3], &u[5])
	d[5].Add(&d[5], &u[9])
	d[5].Sub(&d[5], &u[10])
	d[5].Add(&d[5], &u[12])
	var tmp fr.Element
	tmp.Add(&t2, &t2) // 2*t2
	d[5].Sub(&d[5], &tmp)
	tmp.Add(&t0, &t0) // 2*t0
	tmp.Add(&tmp, &t0)
	d[5].Sub(&d[5], &tmp) // -3*t0

	// d4 = u1 + u2 - u4 + u5 + u6 - u7 - u8 - u12 - 2*t3 + 3*t0
	d[4].Add(&u[1], &u[2])
	d[4].Sub(&d[4], &u[4])
	d[4].Add(&d[4], &u[5])
	d[4].Add(&d[4], &u[6])
	d[4].Sub(&d[4], &u[7])
	d[4].Sub(&d[4], &u[8])
	d[4].Sub(&d[4], &u[12])
	tmp.Add(&t3, &t3) // 2*t3
	d[4].Sub(&d[4], &tmp)
	tmp.Add(&t0, &t0) // 2*t0
	tmp.Add(&tmp, &t0)
	d[4].Add(&d[4], &tmp) // +3*t0

	// d3 = u0 - u6 + u9 - u11 + u12 - 2*t4 - 3*t0
	d[3].Sub(&u[0], &u[6])
	d[3].Add(&d[3], &u[9])
	d[3].Sub(&d[3], &u[11])
	d[3].Add(&d[3], &u[12])
	tmp.Add(&t4, &t4) // 2*t4
	d[3].Sub(&d[3], &tmp)
	tmp.Add(&t0, &t0) // 2*t0
	tmp.Add(&tmp, &t0)
	d[3].Sub(&d[3], &tmp) // -3*t0

	// d2 = t0 + u1 - u5 + t4
	d[2].Add(&t0, &u[1])
	d[2].Sub(&d[2], &u[5])
	d[2].Add(&d[2], &t4)

	// d1 = -u0 - u1 + u5
	d[1].Add(&u[0], &u[1])
	d[1].Neg(&d[1])
	d[1].Add(&d[1], &u[5])

	// d0 = u0
	d[0].Set(&u[0])

	return z.reduceDegree8(&d)
}

// mulElMGuiIon5 implements the 9-multiplication interpolation formula.
//
// Ref. "Efficient Multiplication in Finite Field Extensions of Degree 5" (Sect. 3)
// https://link.springer.com/chapter/10.1007/978-3-642-21969-6_12
func (z *E5D) mulElMGuiIon5(a, b *E5D) *E5D {
	var u [9]fr.Element

	// -------------------------------------------------------------------------
	// Phase 1: evaluations (no base field muls)
	// -------------------------------------------------------------------------
	u[0].Set(&a.A0) // A(0)
	u[1].Add(&u[0], &a.A1)
	u[1].Add(&u[1], &a.A2)
	u[1].Add(&u[1], &a.A3)
	u[1].Add(&u[1], &a.A4) // A(1)
	u[2] = evalAt(a, -1)   // A(-1)
	u[3] = evalAt(a, 2)    // A(2)
	u[4] = evalAt(a, -2)   // A(-2)
	u[5] = evalAt(a, 4)    // A(4)
	u[6] = evalAt(a, -4)   // A(-4)
	u[7] = evalAt(a, 3)    // A(3)
	u[8].Set(&a.A4)        // A(∞)

	// reuse u as A values; compute B values in v then pointwise mul
	var v [9]fr.Element
	v[0].Set(&b.A0)
	v[1].Add(&v[0], &b.A1)
	v[1].Add(&v[1], &b.A2)
	v[1].Add(&v[1], &b.A3)
	v[1].Add(&v[1], &b.A4)
	v[2] = evalAt(b, -1)
	v[3] = evalAt(b, 2)
	v[4] = evalAt(b, -2)
	v[5] = evalAt(b, 4)
	v[6] = evalAt(b, -4)
	v[7] = evalAt(b, 3)
	v[8].Set(&b.A4)

	for i := 0; i < 9; i++ {
		u[i].Mul(&u[i], &v[i])
	}

	// -------------------------------------------------------------------------
	// Phase 2: divided differences (c coefficients)
	// -------------------------------------------------------------------------
	var c [9]fr.Element
	c[0].Set(&u[0])

	c[1].Sub(&u[1], &c[0])

	c[2].Sub(&u[2], &c[0])
	c[2].Add(&c[2], &c[1])
	c[2].Mul(&c[2], &inv2)

	c[3].Sub(&u[3], &c[0])
	c[3].Mul(&c[3], &inv2)
	c[3].Sub(&c[3], &c[1])
	c[3].Sub(&c[3], &c[2])
	c[3].Mul(&c[3], &inv3)

	c[4].Sub(&u[4], &c[0])
	c[4].Mul(&c[4], &inv2)
	c[4].Add(&c[4], &c[1])
	c[4].Mul(&c[4], &inv3)
	c[4].Sub(&c[4], &c[2])
	c[4].Add(&c[4], &c[3])
	c[4].Mul(&c[4], &inv4)

	c[5].Sub(&u[5], &c[0])
	c[5].Mul(&c[5], &inv4)
	c[5].Sub(&c[5], &c[1])
	c[5].Mul(&c[5], &inv3)
	c[5].Sub(&c[5], &c[2])
	c[5].Mul(&c[5], &inv5)
	c[5].Sub(&c[5], &c[3])
	c[5].Mul(&c[5], &inv2)
	c[5].Sub(&c[5], &c[4])
	c[5].Mul(&c[5], &inv6)

	c[6].Sub(&u[6], &c[0])
	c[6].Mul(&c[6], &inv4)
	c[6].Add(&c[6], &c[1])
	c[6].Mul(&c[6], &inv5)
	c[6].Sub(&c[6], &c[2])
	c[6].Mul(&c[6], &inv3)
	c[6].Add(&c[6], &c[3])
	c[6].Mul(&c[6], &inv6)
	c[6].Sub(&c[6], &c[4])
	c[6].Mul(&c[6], &inv2)
	c[6].Add(&c[6], &c[5])
	c[6].Mul(&c[6], &inv8)

	c[7].Neg(&u[7])
	c[7].Add(&c[7], &c[0])
	c[7].Mul(&c[7], &inv3)
	c[7].Add(&c[7], &c[1])
	c[7].Mul(&c[7], &inv2)
	c[7].Add(&c[7], &c[2])
	c[7].Mul(&c[7], &inv4)
	c[7].Add(&c[7], &c[3])
	c[7].Add(&c[7], &c[4])
	c[7].Mul(&c[7], &inv5)
	c[7].Add(&c[7], &c[5])
	c[7].Sub(&c[7], &c[6])
	c[7].Mul(&c[7], &inv7)

	c[8].Set(&u[8]) // leading coefficient

	// -------------------------------------------------------------------------
	// Phase 3: Horner reconstruction to monomial basis
	// -------------------------------------------------------------------------
	poly := []fr.Element{c[8]}
	for i := 7; i >= 0; i-- {
		poly = multiplyPolyByMonomial(poly, alphaInterp[i])
		poly[0].Add(&poly[0], &c[i])
	}

	var d [9]fr.Element
	copy(d[:], poly)

	return z.reduceDegree8(&d)
}

// Square sets z=x*x in E5D and returns z.
func (z *E5D) Square(x *E5D) *E5D {
	// Use the 9-multiplication interpolation with squares.
	var u [9]fr.Element

	u[0].Set(&x.A0)
	u[1].Add(&u[0], &x.A1)
	u[1].Add(&u[1], &x.A2)
	u[1].Add(&u[1], &x.A3)
	u[1].Add(&u[1], &x.A4)
	u[2] = evalAt(x, -1)
	u[3] = evalAt(x, 2)
	u[4] = evalAt(x, -2)
	u[5] = evalAt(x, 4)
	u[6] = evalAt(x, -4)
	u[7] = evalAt(x, 3)
	u[8].Set(&x.A4)

	for i := 0; i < 9; i++ {
		u[i].Square(&u[i])
	}

	var c [9]fr.Element
	c[0].Set(&u[0])

	c[1].Sub(&u[1], &c[0])

	c[2].Sub(&u[2], &c[0])
	c[2].Add(&c[2], &c[1])
	c[2].Mul(&c[2], &inv2)

	c[3].Sub(&u[3], &c[0])
	c[3].Mul(&c[3], &inv2)
	c[3].Sub(&c[3], &c[1])
	c[3].Sub(&c[3], &c[2])
	c[3].Mul(&c[3], &inv3)

	c[4].Sub(&u[4], &c[0])
	c[4].Mul(&c[4], &inv2)
	c[4].Add(&c[4], &c[1])
	c[4].Mul(&c[4], &inv3)
	c[4].Sub(&c[4], &c[2])
	c[4].Add(&c[4], &c[3])
	c[4].Mul(&c[4], &inv4)

	c[5].Sub(&u[5], &c[0])
	c[5].Mul(&c[5], &inv4)
	c[5].Sub(&c[5], &c[1])
	c[5].Mul(&c[5], &inv3)
	c[5].Sub(&c[5], &c[2])
	c[5].Mul(&c[5], &inv5)
	c[5].Sub(&c[5], &c[3])
	c[5].Mul(&c[5], &inv2)
	c[5].Sub(&c[5], &c[4])
	c[5].Mul(&c[5], &inv6)

	c[6].Sub(&u[6], &c[0])
	c[6].Mul(&c[6], &inv4)
	c[6].Add(&c[6], &c[1])
	c[6].Mul(&c[6], &inv5)
	c[6].Sub(&c[6], &c[2])
	c[6].Mul(&c[6], &inv3)
	c[6].Add(&c[6], &c[3])
	c[6].Mul(&c[6], &inv6)
	c[6].Sub(&c[6], &c[4])
	c[6].Mul(&c[6], &inv2)
	c[6].Add(&c[6], &c[5])
	c[6].Mul(&c[6], &inv8)

	c[7].Neg(&u[7])
	c[7].Add(&c[7], &c[0])
	c[7].Mul(&c[7], &inv3)
	c[7].Add(&c[7], &c[1])
	c[7].Mul(&c[7], &inv2)
	c[7].Add(&c[7], &c[2])
	c[7].Mul(&c[7], &inv4)
	c[7].Add(&c[7], &c[3])
	c[7].Add(&c[7], &c[4])
	c[7].Mul(&c[7], &inv5)
	c[7].Add(&c[7], &c[5])
	c[7].Sub(&c[7], &c[6])
	c[7].Mul(&c[7], &inv7)

	c[8].Set(&u[8])

	poly := []fr.Element{c[8]}
	for i := 7; i >= 0; i-- {
		poly = multiplyPolyByMonomial(poly, alphaInterp[i])
		poly[0].Add(&poly[0], &c[i])
	}

	var d [9]fr.Element
	copy(d[:], poly)

	return z.reduceDegree8(&d)
}

var (
	inv2 fr.Element
	inv3 fr.Element
	inv4 fr.Element
	inv5 fr.Element
	inv6 fr.Element
	inv7 fr.Element
	inv8 fr.Element

	alphaInterp = [...]int64{0, 1, -1, 2, -2, 4, -4, 3}
)

func init() {
	var tmp fr.Element
	tmp.SetUint64(2)
	inv2.Inverse(&tmp)
	tmp.SetUint64(3)
	inv3.Inverse(&tmp)
	tmp.SetUint64(4)
	inv4.Inverse(&tmp)
	tmp.SetUint64(5)
	inv5.Inverse(&tmp)
	tmp.SetUint64(6)
	inv6.Inverse(&tmp)
	tmp.SetUint64(7)
	inv7.Inverse(&tmp)
	tmp.SetUint64(8)
	inv8.Inverse(&tmp)
}

// reduceDegree8 applies w^5 = 2 to bring degree-8 coeffs back to degree < 5.
func (z *E5D) reduceDegree8(d *[9]fr.Element) *E5D {
	var c0, c1, c2, c3 fr.Element

	c0.Double(&d[5])
	c0.Add(&c0, &d[0])

	c1.Double(&d[6])
	c1.Add(&c1, &d[1])

	c2.Double(&d[7])
	c2.Add(&c2, &d[2])

	c3.Double(&d[8])
	c3.Add(&c3, &d[3])

	z.A0.Set(&c0)
	z.A1.Set(&c1)
	z.A2.Set(&c2)
	z.A3.Set(&c3)
	z.A4.Set(&d[4])
	return z
}

// evalAt evaluates a at a small integer alpha using only additions/doubles.
func evalAt(a *E5D, alpha int64) fr.Element {
	var r fr.Element
	switch alpha {
	case 0:
		return r
	case 1:
		r.Add(&a.A1, &a.A0)
		r.Add(&r, &a.A2)
		r.Add(&r, &a.A3)
		r.Add(&r, &a.A4)
		return r
	case -1:
		r.Sub(&a.A4, &a.A3)
		r.Add(&r, &a.A2)
		r.Sub(&r, &a.A1)
		r.Add(&r, &a.A0)
		return r
	case 2:
		r.Set(&a.A4)
		r.Double(&r)
		r.Add(&r, &a.A3)
		r.Double(&r)
		r.Add(&r, &a.A2)
		r.Double(&r)
		r.Add(&r, &a.A1)
		r.Double(&r)
		r.Add(&r, &a.A0)
		return r
	case 4:
		r.Set(&a.A4)
		r.Double(&r)
		r.Double(&r) // *4
		r.Add(&r, &a.A3)
		r.Double(&r)
		r.Double(&r)
		r.Add(&r, &a.A2)
		r.Double(&r)
		r.Double(&r)
		r.Add(&r, &a.A1)
		r.Double(&r)
		r.Double(&r)
		r.Add(&r, &a.A0)
		return r
	default:
		// fallback Horner using mulBySmall (covers -1, -2, -4, 3, etc.).
		r.Set(&a.A4)
		r = mulBySmall(&r, alpha)
		r.Add(&r, &a.A3)
		r = mulBySmall(&r, alpha)
		r.Add(&r, &a.A2)
		r = mulBySmall(&r, alpha)
		r.Add(&r, &a.A1)
		r = mulBySmall(&r, alpha)
		r.Add(&r, &a.A0)
		return r
	}
}

// mulBySmall multiplies x by a small signed integer using only doubles/adds.
func mulBySmall(x *fr.Element, k int64) fr.Element {
	if k == 0 {
		return fr.Element{}
	}
	var res fr.Element
	abs := k
	if k < 0 {
		abs = -k
	}
	switch abs {
	case 1:
		res.Set(x)
	case 2:
		res.Double(x)
	case 3:
		res.Double(x)
		res.Add(&res, x)
	case 4:
		res.Double(x)
		res.Double(&res)
	default:
		// fallback (should not hit with our fixed set)
		var tmp fr.Element
		res.SetZero()
		tmp.Set(x)
		n := abs
		for n > 0 {
			if n&1 == 1 {
				res.Add(&res, &tmp)
			}
			tmp.Double(&tmp)
			n >>= 1
		}
	}
	if k < 0 {
		res.Neg(&res)
	}
	return res
}

// multiplyPolyByMonomial computes (X - alpha)*poly.
func multiplyPolyByMonomial(poly []fr.Element, alpha int64) []fr.Element {
	newPoly := make([]fr.Element, len(poly)+1)
	// newPoly[0] = -alpha * poly[0]
	tmp := mulBySmall(&poly[0], alpha)
	newPoly[0].Neg(&tmp)
	for i := 1; i < len(poly); i++ {
		tmp = mulBySmall(&poly[i], alpha)
		newPoly[i].Sub(&poly[i-1], &tmp)
	}
	newPoly[len(poly)].Set(&poly[len(poly)-1])
	return newPoly
}
