package fptower

import "github.com/consensys/gnark-crypto/ecc/bw6-756/fp"

func (z *E6) nSquare(n int) {
	for i := 0; i < n; i++ {
		z.CyclotomicSquare(z)
	}
}

func (z *E6) nSquareCompressed(n int) {
	for i := 0; i < n; i++ {
		z.CyclotomicSquareCompressed(z)
	}
}

// ExptMinus1 set z to x^t in E6 and return z
// t-1 = 11045256207009841152
func (z *E6) ExptMinus1(x *E6) *E6 {

	var result, t0, t1 E6

	result.CyclotomicSquare(x)
	result.nSquare(2)
	t0.Mul(x, &result)
	t1.CyclotomicSquare(&t0)
	t1.nSquare(2)
	result.Mul(&t0, &t1)
	t1.Mul(&t1, &result)
	t1.nSquare(5)
	t0.Mul(&t0, &t1)
	t0.nSquare(10)
	result.Mul(&result, &t0)
	result.nSquareCompressed(41)
	result.DecompressKarabina(&result)
	z.Set(&result)

	return z
}

// ExptMinus1Square set z to x^t in E6 and return z
// (t-1)^2 = 121997684678489422939424157776272687104
func (z *E6) ExptMinus1Square(x *E6) *E6 {
	var result, t0, t1, t2, t3 E6
	result.CyclotomicSquare(x)
	t0.Mul(x, &result)
	result.CyclotomicSquare(&t0)
	t0.Mul(&t0, &result)
	result.Mul(&result, &t0)
	t1.CyclotomicSquare(&result)
	t1.nSquare(2)
	t0.Mul(&t0, &t1)
	result.Mul(&result, &t0)
	t0.Mul(&t0, &result)
	result.Mul(&result, &t0)
	t1.CyclotomicSquare(&result)
	t1.Mul(&result, &t1)
	t1.CyclotomicSquare(&t1)
	t2.CyclotomicSquare(&t1)
	t1.Mul(&t1, &t2)
	t2.Mul(&t2, &t1)
	t1.Mul(&t1, &t2)
	t3.CyclotomicSquare(&t1)
	t2.Mul(&t2, &t3)
	t1.Mul(&t1, &t2)
	t3.CyclotomicSquare(&t1)
	t3.CyclotomicSquare(&t3)
	t3.Mul(&t1, &t3)
	t3.nSquare(2)
	t2.Mul(&t2, &t3)
	t2.nSquare(11)
	t1.Mul(&t1, &t2)
	t0.Mul(&t0, &t1)
	t0.nSquare(13)
	result.Mul(&result, &t0)
	result.nSquareCompressed(82)
	result.DecompressKarabina(&result)
	z.Set(&result)

	return z
}

// Expt set z to x^t in E6 and return z
// t = 11045256207009841153
func (z *E6) Expt(x *E6) *E6 {
	var result E6

	result.ExptMinus1(x)
	result.Mul(&result, x)
	z.Set(&result)

	return z
}

// ExptPlus1 set z to x^(t+1) in E6 and return z
// t+1 = 11045256207009841154
func (z *E6) ExptPlus1(x *E6) *E6 {
	var result, t E6

	result.ExptMinus1(x)
	t.CyclotomicSquare(x)
	result.Mul(&result, &t)
	z.Set(&result)

	return z
}

// ExptMinus1Div3 set z to x^(t-1)/3 in E6 and return z
// (t-1)/3 = 3681752069003280384
func (z *E6) ExptMinus1Div3(x *E6) *E6 {
	var result, t0, t1 E6
	result.CyclotomicSquare(x)
	t0.Mul(x, &result)
	t1.CyclotomicSquare(&t0)
	t1.nSquare(2)
	result.Mul(&t0, &t1)
	t1.Mul(&t1, &result)
	t1.nSquare(5)
	t0.Mul(&t0, &t1)
	t0.nSquare(10)
	result.Mul(&result, &t0)
	result.nSquareCompressed(41)
	result.DecompressKarabina(&result)
	z.Set(&result)

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
func Mul034By034(d0, d3, d4, c0, c3, c4 *fp.Element) [5]fp.Element {
	var z00, tmp, x0, x3, x4, x04, x03, x34 fp.Element
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

	z00.MulByNonResidue(&x4).
		Add(&z00, &x0)

	return [5]fp.Element{z00, x3, x34, x03, x04}
}

// MulBy01234 multiplies z by an E12 sparse element of the form (x0, x1, x2, x3, x4, 0)
func (z *E6) MulBy01234(x *[5]fp.Element) *E6 {
	var c1, a, b, c, z0, z1 E3
	c0 := &E3{A0: x[0], A1: x[1], A2: x[2]}
	c1.A0 = x[3]
	c1.A1 = x[4]
	a.Add(&z.B0, &z.B1)
	b.Add(c0, &c1)
	a.Mul(&a, &b)
	b.Mul(&z.B0, c0)
	c.Set(&z.B1).MulBy01(&x[3], &x[4])
	z1.Sub(&a, &b)
	z1.Sub(&z1, &c)
	z0.MulByNonResidue(&c)
	z0.Add(&z0, &b)

	z.B0 = z0
	z.B1 = z1

	return z
}
