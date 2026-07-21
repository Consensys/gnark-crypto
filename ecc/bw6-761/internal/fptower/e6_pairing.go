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
// (t-1)² = 91893752504881257682351033800651177984 = 0x452217cc900000000000000000000000
//
// Addition chain: ((((69<<7 + 17)<<11 + 69 + 26)<<7 + 25)<<3 + 1)<<92 with
// 17 = 1 + 2*(1<<3), 25 = 1<<3 + 17, 26 = 1 + 25, 69 = 17 + 2*26.
// Operations: 125 squares, 9 multiplications, 1 decompression.
func (z *E6) ExptMinus1Square(x *E6) *E6 {
	var result, x17, x25, x26, x69 E6

	result.CyclotomicSquare(x)
	result.CyclotomicSquare(&result)
	result.CyclotomicSquare(&result) // x^8
	x17.CyclotomicSquare(&result)    // x^16
	x17.Mul(x, &x17)                 // x^17
	x25.Mul(&result, &x17)           // x^25
	x26.Mul(x, &x25)                 // x^26
	x69.CyclotomicSquare(&x26)       // x^52
	x69.Mul(&x17, &x69)              // x^69

	result.Set(&x69)
	result.nSquare(7)
	result.Mul(&result, &x17)
	result.nSquare(11)
	result.Mul(&result, &x69)
	result.Mul(&result, &x26)
	result.nSquare(7)
	result.Mul(&result, &x25)
	result.nSquare(3)
	result.Mul(&result, x)
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
