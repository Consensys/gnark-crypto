package fptower

import "github.com/consensys/gnark-crypto/ecc/bw6-633/fp"

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

// Expc1 set z to z^c1 in E6 and return z
// ht, hy = -7, -1
// c1 = (ht-hy)/2 = -3
func (z *E6) Expc1(x *E6) *E6 {
	var result E6
	result.CyclotomicSquare(x)
	result.Mul(x, &result)
	z.Conjugate(&result)

	return z
}

// Expc2 set z to z^c2 in E6 and return z
// ht, hy = -7, -1
// c2 = (ht**2+3*hy**2)/4 = 13
func (z *E6) Expc2(x *E6) *E6 {
	var result E6
	result.CyclotomicSquare(x)
	result.Mul(x, &result)
	result.nSquare(2)
	result.Mul(x, &result)
	z.Set(&result)

	return z
}

// Expt set z to x^t in E6 and return z (t is the seed of the curve)
// t = -3218079743 = -2**32+2**30+2**22-2**20+1
func (z *E6) Expt(x *E6) *E6 {

	var result, x20, x22, x30, x32 E6
	result.Set(x)

	result.nSquareCompressed(20)
	x20.Conjugate(&result)
	result.nSquareCompressed(2)
	x22.Set(&result)
	result.nSquareCompressed(8)
	x30.Set(&result)

	batch := BatchDecompressKarabina([]E6{x20, x22, x30})

	x32.CyclotomicSquare(&batch[2]).
		CyclotomicSquare(&x32).
		Conjugate(&x32)

	z.Mul(x, &batch[0]).
		Mul(z, &batch[1]).
		Mul(z, &batch[2]).
		Mul(z, &x32)

	return z
}

// ExptMinus1 set z to x^(t-1) in E6 and return z
// t-1 = -3218079744
func (z *E6) ExptMinus1(x *E6) *E6 {
	var result, t E6
	result.Expt(x)
	t.Conjugate(x)
	result.Mul(&result, &t)
	z.Set(&result)

	return z
}

// ExptMinus1Squared set z to x^(t-1)^2 in E6 and return z
// (t-1)^2 = 10356037238743105536
func (z *E6) ExptMinus1Squared(x *E6) *E6 {
	var result, t0, t1, t2 E6
	result.CyclotomicSquare(x)
	result.CyclotomicSquare(&result)
	t1.Mul(x, &result)
	result.Mul(&result, &t1)
	t0.CyclotomicSquare(&result)
	t0.Mul(&t1, &t0)
	t2.CyclotomicSquare(&t0)
	t2.Mul(&t0, &t2)
	t2.CyclotomicSquare(&t2)
	t1.Mul(&t1, &t2)
	t1.nSquare(5)
	t0.Mul(&t0, &t1)
	t0.nSquare(11)
	result.Mul(&result, &t0)
	result.nSquareCompressed(40)
	result.DecompressKarabina(&result)
	z.Set(&result)
	return z
}

// ExptPlus1 set z to x^(t+1) in E6 and return z
// t + 1 = -3218079742
func (z *E6) ExptPlus1(x *E6) *E6 {
	var result E6
	result.Expt(x)
	result.Mul(&result, x)
	z.Set(&result)

	return z
}

// ExptSquarePlus1 set z to x^(t^2+1) in E6 and return z
// t^2 + 1 = 10356037232306946050
func (z *E6) ExptSquarePlus1(x *E6) *E6 {
	var result, t0, t1, t2, t3 E6
	t0.CyclotomicSquare(x)
	result.Mul(x, &t0)
	t0.Mul(&t0, &result)
	t1.Mul(x, &t0)
	t0.Mul(&t0, &t1)
	t0.CyclotomicSquare(&t0)
	t2.Mul(x, &t0)
	t0.Mul(&t1, &t2)
	t1.Mul(&t1, &t0)
	t1.nSquare(2)
	t1.Mul(&result, &t1)
	t3.CyclotomicSquare(&t1)
	t3.nSquare(4)
	t2.Mul(&t2, &t3)
	t2.nSquareCompressed(15)
	t2.DecompressKarabina(&t2)
	t1.Mul(&t1, &t2)
	t1.nSquare(5)
	t0.Mul(&t0, &t1)
	t0.nSquare(10)
	result.Mul(&result, &t0)
	result.nSquareCompressed(20)
	result.DecompressKarabina(&result)
	result.Mul(x, &result)
	result.CyclotomicSquare(&result)
	z.Set(&result)

	return z
}

// ExptMinus1Div3 set z to x^((t-1)/3) in E6 and return z
// (t-1)/3 = -1072693248
func (z *E6) ExptMinus1Div3(x *E6) *E6 {
	var result, t0, t1 E6
	result.CyclotomicSquare(x)
	result.Mul(x, &result)
	t0.CyclotomicSquare(&result)
	t0.CyclotomicSquare(&t0)
	t0.Mul(&result, &t0)
	t1.CyclotomicSquare(&t0)
	t1.nSquare(3)
	t0.Mul(&t0, &t1)
	t0.nSquare(2)
	result.Mul(&result, &t0)
	result.nSquareCompressed(20)
	result.DecompressKarabina(&result)
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
