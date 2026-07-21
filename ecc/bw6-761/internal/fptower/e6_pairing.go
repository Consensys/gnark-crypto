// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fptower

// ExptMinus1 set z to x^(t-1) in E6 and return z
// t-1 = 91893752504881257682351033800651177983
func (z *E6) ExptMinus1(x *E6) *E6 {

	var result, x33 E6
	result.Set(x)
	result.nSquare(5)
	result.Mul(&result, x)
	x33.Set(&result)
	result.nSquare(7)
	result.Mul(&result, &x33)
	result.nSquare(4)
	result.Mul(&result, x)
	result.CyclotomicSquare(&result)
	result.Mul(&result, x)
	result.nSquareCompressed(46)
	result.DecompressKarabina(&result)
	z.Set(&result)

	return z
}

// ExptMinus1Square set z to x^{(t-1)²} in E6 and return z
// (t-1)² = 91893752504881257682351033800651177984
func (z *E6) ExptMinus1Square(x *E6) *E6 {
	var result, t0, t1, t2 E6

	result.Set(x)
	result.CyclotomicSquare(&result)
	t0.Mul(x, &result)
	t1.CyclotomicSquare(&t0)
	t0.Mul(&t0, &t1)
	result.Mul(&result, &t0)
	t1.Mul(&t1, &result)
	t0.Mul(&t0, &t1)
	t2.CyclotomicSquare(&t0)
	t2.Mul(&t1, &t2)
	t0.Mul(&t0, &t2)
	t2.nSquare(7)
	t1.Mul(&t1, &t2)
	t1.nSquare(11)
	t1.Mul(&t0, &t1)
	t1.nSquare(9)
	t0.Mul(&t0, &t1)
	t0.CyclotomicSquare(&t0)
	result.Mul(&result, &t0)
	result.nSquareCompressed(92)
	result.DecompressKarabina(&result)
	z.Set(&result)

	return z
}

// Expt set z to x^t in E6 and return z
// t = 91893752504881257682351033800651177984
func (z *E6) Expt(x *E6) *E6 {
	var result E6

	result.ExptMinus1(x)
	result.Mul(&result, x)
	z.Set(&result)

	return z
}

// ExptPlus1 set z to x^(t+1) in E6 and return z
// t+1 = 91893752504881257682351033800651177985
func (z *E6) ExptPlus1(x *E6) *E6 {
	var result, t E6

	result.ExptMinus1(x)
	t.CyclotomicSquare(x)
	result.Mul(&result, &t)
	z.Set(&result)

	return z
}

// ExptMinus1Div3 set z to x^(t-1)/3 in E6 and return z
// (t-1)/3 = 3195374304363544576
func (z *E6) ExptMinus1Div3(x *E6) *E6 {
	var result, t0 E6

	result.Set(x)
	result.CyclotomicSquare(&result)
	result.Mul(&result, x)
	t0.Mul(&result, x)
	t0.CyclotomicSquare(&t0)
	result.Mul(&result, &t0)
	t0.Set(&result)
	t0.nSquare(7)
	result.Mul(&result, &t0)
	result.nSquare(5)
	result.Mul(&result, x)
	result.nSquareCompressed(46)
	result.DecompressKarabina(&result)
	z.Set(&result)

	return z
}

// Expc1 set z to z^c1 in E6 and return z
// ht, hy = 13, 9
// c1 = (ht+hy)/2 = 11
func (z *E6) Expc1(x *E6) *E6 {
	var result, t0 E6

	result.CyclotomicSquare(x)
	result.Mul(&result, x)
	t0.Mul(x, &result)
	t0.CyclotomicSquare(&t0)
	result.Mul(&result, &t0)
	z.Set(&result)

	return z
}

// Expc2 set z to z^c2 in E6 and return z
// ht, hy = 13, 9
// c2 = (ht**2+3*hy**2)/4 = 103
func (z *E6) Expc2(x *E6) *E6 {
	var result, t0 E6

	result.CyclotomicSquare(x)
	result.Mul(&result, x)
	t0.Set(&result)
	t0.nSquare(4)
	result.Mul(&result, &t0)
	result.CyclotomicSquare(&result)
	result.Mul(&result, x)
	z.Set(&result)

	return z
}
