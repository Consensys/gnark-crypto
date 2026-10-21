// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fptower

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
//
// Addition-subtraction chain for |t|: (3<<10 - 3)<<20 - 1, computed on x^-1
// to absorb the sign; subtraction is a multiplication by the conjugate.
// Operations: 31 squares, 3 multiplications, 3 conjugates.
//
// Plain Granger-Scott squares throughout: on this curve a Karabina
// decompression costs about 27 cyclotomic squares' worth of savings, so
// compressed squares do not pay off for these run lengths.
func (z *E6) Expt(x *E6) *E6 {
	var result, t0, t1 E6

	result.Conjugate(x) // x^-1; chain below computes (x^-1)^|t| = x^t
	t0.CyclotomicSquare(&result)
	result.Mul(&result, &t0) // ^3
	t1.Conjugate(&result)    // ^-3
	t0.Set(&result)
	t0.nSquare(10)
	t0.Mul(&t0, &t1) // ^(3*2^10 - 3)
	t0.nSquare(20)
	z.Mul(&t0, x) // -1 on x^-1: multiply by x

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
//
// Expanded signed form 9*2^60 - 9*2^51 + 9*2^40: a single compressed run on
// x^9 with snapshots at 40, 51 and 60, decompressed in one batch so the
// three runs share a single inversion.
// Operations: 3 squares, 60 compressed squares, 1 batch decompression,
// 3 multiplications, 1 conjugate.
func (z *E6) ExptMinus1Squared(x *E6) *E6 {
	var c, s40, s51, r E6

	c.CyclotomicSquare(x)
	c.CyclotomicSquare(&c)
	c.CyclotomicSquare(&c)
	c.Mul(&c, x) // x^9
	c.nSquareCompressed(40)
	s40.Set(&c)
	c.nSquareCompressed(11)
	s51.Set(&c)
	c.nSquareCompressed(9)
	batch := BatchDecompressKarabina([]E6{s40, s51, c})
	r.Conjugate(&batch[1]) // x^-(9*2^51)
	r.Mul(&r, &batch[0])   // * x^(9*2^40)
	z.Mul(&r, &batch[2])   // * x^(9*2^60)

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
//
// Addition-subtraction chain:
// 2*(((((9<<9 - 9)<<11 + 9)<<9 - 3)<<10 + 3)<<20 + 1) with 3 = 1 + 2, 9 = 3 + 2*3.
// Operations: 62 squares, 7 multiplications, 2 conjugates.
//
// Plain Granger-Scott squares throughout: on this curve a Karabina
// decompression costs about 27 cyclotomic squares' worth of savings, so
// compressed squares do not pay off for these run lengths.
func (z *E6) ExptSquarePlus1(x *E6) *E6 {
	var result, t0, x3, x9 E6

	t0.CyclotomicSquare(x)
	x3.Mul(x, &t0) // x^3
	t0.CyclotomicSquare(&x3)
	x9.Mul(&x3, &t0) // x^9

	result.Set(&x9)
	result.nSquare(9)
	t0.Conjugate(&x9)
	result.Mul(&result, &t0) // ^(9*2^9 - 9)
	result.nSquare(11)
	result.Mul(&result, &x9) // ^(... + 9)
	result.nSquare(9)
	t0.Conjugate(&x3)
	result.Mul(&result, &t0) // ^(... - 3)
	result.nSquare(10)
	result.Mul(&result, &x3) // ^(... + 3)
	result.nSquare(20)
	result.Mul(&result, x) // ^(... + 1)
	z.CyclotomicSquare(&result)

	return z
}

// ExptMinus1Div3 set z to x^((t-1)/3) in E6 and return z
// (t-1)/3 = -1072693248 = -(2^10-1)*2^20
//
// Subtraction chain for the absolute value: (1<<10 - 1)<<20, computed on x^-1
// to absorb the sign; the subtraction is a multiplication by x itself.
// Operations: 30 squares, 1 multiplication, 1 conjugate.
func (z *E6) ExptMinus1Div3(x *E6) *E6 {
	var result E6
	result.Conjugate(x) // x^-1
	result.nSquare(10)
	result.Mul(&result, x) // (x^-1)^(2^10 - 1)
	result.nSquare(20)
	z.Set(&result)

	return z
}
