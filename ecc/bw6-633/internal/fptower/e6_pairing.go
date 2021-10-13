package fptower

import "github.com/consensys/gnark-crypto/ecc/bw6-633/fp"

func (z *E6) nSquare(n int) {
	for i := 0; i < n; i++ {
		z.CyclotomicSquare(z)
	}
}

// Expt set z to x^t in E6 and return z (t is the seed of the curve)
func (z *E6) Expt(x *E6) *E6 {

	var result, xInv E6
	result.Set(x)
	xInv.Conjugate(x)

	result.nSquare(2)
	result.Mul(&result, &xInv)
	result.nSquare(8)
	result.Mul(&result, &xInv)
	result.nSquare(2)
	result.Mul(&result, x)
	result.nSquare(20)
	result.Mul(&result, &xInv)

	z.Conjugate(&result)

	return z
}

// MulBy034 multiplication by sparse element (c0,0,0,c3,c4,0)
func (z *E6) MulBy034(c0, c3, c4 *fp.Element) *E6 {

	var a, b, d E3

	a.MulByElement(&z.B0, c0)

	b.Set(&z.B1)
	b.MulBy01(c3, c4)

	c0.Add(c0, c3)
	d.Add(&z.B0, &z.B1)
	d.MulBy01(c0, c4)

	z.B1.Add(&a, &b).Neg(&z.B1).Add(&z.B1, &d)
	z.B0.MulByNonResidue(&b).Add(&z.B0, &a)

	return z
}

// Mul034By034 multiplication of sparse element (c0,0,0,c3,c4,0) by sparse element (d0,0,0,d3,d4,0)
func (z *E6) Mul034By034(d0, d3, d4, c0, c3, c4 *fp.Element) *E6 {
	var tmp, x0, x3, x4, x04, x03, x34 fp.Element
	x0.Mul(c0, d0)
	x3.Mul(c3, d3)
	x4.Mul(c4, d4)
	tmp.Add(c0, c4)
	x04.Add(d0, d4).
		Mul(&x04, &tmp).
		Sub(&x04, &x0).
		Sub(&x04, &x4)
	tmp.Add(c0, c3)
	x03.Add(d0, d3).
		Mul(&x03, &tmp).
		Sub(&x03, &x0).
		Sub(&x03, &x3)
	tmp.Add(c3, c4)
	x34.Add(d3, d4).
		Mul(&x34, &tmp).
		Sub(&x34, &x3).
		Sub(&x34, &x4)

	z.B0.A0.MulByNonResidue(&x4).
		Add(&z.B0.A0, &x0)
	z.B0.A1.Set(&x3)
	z.B0.A2.Set(&x34)
	z.B1.A0.Set(&x03)
	z.B1.A1.Set(&x04)
	z.B1.A2.SetZero()

	return z
}
