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
