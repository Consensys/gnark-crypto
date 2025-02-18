// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package fr

import "math/bits"

// MulBy3 x *= 3 (mod q)
func MulBy3(x *Element) {
	_x := *x
	x.Double(x).Add(x, &_x)
}

// MulBy5 x *= 5 (mod q)
func MulBy5(x *Element) {
	_x := *x
	x.Double(x).Double(x).Add(x, &_x)
}

// MulBy11 x *= 11 (mod q)
func MulBy11(x *Element) {
	var y = Element{
		13976340845799617077,
		18049444439606238574,
		13,
		0,
	}
	x.Mul(x, &y)
}

// MulBy13 x *= 13 (mod q)
func MulBy13(x *Element) {
	var y = Element{
		4778656589038923699,
		9592324472628567287,
		16,
		0,
	}
	x.Mul(x, &y)
}

func fromMont(z *Element) {
	_fromMontGeneric(z)
}

func reduce(z *Element) {
	_reduceGeneric(z)
}

// Mul z = x * y (mod q)
func (z *Element) Mul(x, y *Element) *Element {

	// Algorithm 2 of "Faster Montgomery Multiplication and Multi-Scalar-Multiplication for SNARKS"
	// by Y. El Housni and G. Botrel https://doi.org/10.46586/tches.v2023.i3.504-521

	var t [5]uint64
	var D uint64
	var m, C uint64
	// -----------------------------------
	// First loop

	C, t[0] = bits.Mul64(y[0], x[0])
	C, t[1] = madd1(y[0], x[1], C)
	C, t[2] = madd1(y[0], x[2], C)
	C, t[3] = madd1(y[0], x[3], C)

	t[4], D = bits.Add64(t[4], C, 0)

	// m = t[0]n'[0] mod W
	m = t[0] * qInvNeg

	// -----------------------------------
	// Second loop
	C = madd0(m, q0, t[0])
	C, t[0] = madd2(m, q1, t[1], C)
	C, t[1] = madd2(m, q2, t[2], C)
	C, t[2] = madd2(m, q3, t[3], C)

	t[3], C = bits.Add64(t[4], C, 0)
	t[4], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(y[1], x[0], t[0])
	C, t[1] = madd2(y[1], x[1], t[1], C)
	C, t[2] = madd2(y[1], x[2], t[2], C)
	C, t[3] = madd2(y[1], x[3], t[3], C)

	t[4], D = bits.Add64(t[4], C, 0)

	// m = t[0]n'[0] mod W
	m = t[0] * qInvNeg

	// -----------------------------------
	// Second loop
	C = madd0(m, q0, t[0])
	C, t[0] = madd2(m, q1, t[1], C)
	C, t[1] = madd2(m, q2, t[2], C)
	C, t[2] = madd2(m, q3, t[3], C)

	t[3], C = bits.Add64(t[4], C, 0)
	t[4], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(y[2], x[0], t[0])
	C, t[1] = madd2(y[2], x[1], t[1], C)
	C, t[2] = madd2(y[2], x[2], t[2], C)
	C, t[3] = madd2(y[2], x[3], t[3], C)

	t[4], D = bits.Add64(t[4], C, 0)

	// m = t[0]n'[0] mod W
	m = t[0] * qInvNeg

	// -----------------------------------
	// Second loop
	C = madd0(m, q0, t[0])
	C, t[0] = madd2(m, q1, t[1], C)
	C, t[1] = madd2(m, q2, t[2], C)
	C, t[2] = madd2(m, q3, t[3], C)

	t[3], C = bits.Add64(t[4], C, 0)
	t[4], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(y[3], x[0], t[0])
	C, t[1] = madd2(y[3], x[1], t[1], C)
	C, t[2] = madd2(y[3], x[2], t[2], C)
	C, t[3] = madd2(y[3], x[3], t[3], C)

	t[4], D = bits.Add64(t[4], C, 0)

	// m = t[0]n'[0] mod W
	m = t[0] * qInvNeg

	// -----------------------------------
	// Second loop
	C = madd0(m, q0, t[0])
	C, t[0] = madd2(m, q1, t[1], C)
	C, t[1] = madd2(m, q2, t[2], C)
	C, t[2] = madd2(m, q3, t[3], C)

	t[3], C = bits.Add64(t[4], C, 0)
	t[4], _ = bits.Add64(0, D, C)

	if t[4] != 0 {
		// we need to reduce, we have a result on 5 words
		var b uint64
		z[0], b = bits.Sub64(t[0], q0, 0)
		z[1], b = bits.Sub64(t[1], q1, b)
		z[2], b = bits.Sub64(t[2], q2, b)
		z[3], _ = bits.Sub64(t[3], q3, b)
		return z
	}

	// copy t into z
	z[0] = t[0]
	z[1] = t[1]
	z[2] = t[2]
	z[3] = t[3]

	// if z ⩾ q → z -= q
	if !z.smallerThanModulus() {
		var b uint64
		z[0], b = bits.Sub64(z[0], q0, 0)
		z[1], b = bits.Sub64(z[1], q1, b)
		z[2], b = bits.Sub64(z[2], q2, b)
		z[3], _ = bits.Sub64(z[3], q3, b)
	}
	return z
}

// Square z = x * x (mod q)
func (z *Element) Square(x *Element) *Element {
	// see Mul for algorithm documentation

	var t [5]uint64
	var D uint64
	var m, C uint64
	// -----------------------------------
	// First loop

	C, t[0] = bits.Mul64(x[0], x[0])
	C, t[1] = madd1(x[0], x[1], C)
	C, t[2] = madd1(x[0], x[2], C)
	C, t[3] = madd1(x[0], x[3], C)

	t[4], D = bits.Add64(t[4], C, 0)

	// m = t[0]n'[0] mod W
	m = t[0] * qInvNeg

	// -----------------------------------
	// Second loop
	C = madd0(m, q0, t[0])
	C, t[0] = madd2(m, q1, t[1], C)
	C, t[1] = madd2(m, q2, t[2], C)
	C, t[2] = madd2(m, q3, t[3], C)

	t[3], C = bits.Add64(t[4], C, 0)
	t[4], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[1], x[0], t[0])
	C, t[1] = madd2(x[1], x[1], t[1], C)
	C, t[2] = madd2(x[1], x[2], t[2], C)
	C, t[3] = madd2(x[1], x[3], t[3], C)

	t[4], D = bits.Add64(t[4], C, 0)

	// m = t[0]n'[0] mod W
	m = t[0] * qInvNeg

	// -----------------------------------
	// Second loop
	C = madd0(m, q0, t[0])
	C, t[0] = madd2(m, q1, t[1], C)
	C, t[1] = madd2(m, q2, t[2], C)
	C, t[2] = madd2(m, q3, t[3], C)

	t[3], C = bits.Add64(t[4], C, 0)
	t[4], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[2], x[0], t[0])
	C, t[1] = madd2(x[2], x[1], t[1], C)
	C, t[2] = madd2(x[2], x[2], t[2], C)
	C, t[3] = madd2(x[2], x[3], t[3], C)

	t[4], D = bits.Add64(t[4], C, 0)

	// m = t[0]n'[0] mod W
	m = t[0] * qInvNeg

	// -----------------------------------
	// Second loop
	C = madd0(m, q0, t[0])
	C, t[0] = madd2(m, q1, t[1], C)
	C, t[1] = madd2(m, q2, t[2], C)
	C, t[2] = madd2(m, q3, t[3], C)

	t[3], C = bits.Add64(t[4], C, 0)
	t[4], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[3], x[0], t[0])
	C, t[1] = madd2(x[3], x[1], t[1], C)
	C, t[2] = madd2(x[3], x[2], t[2], C)
	C, t[3] = madd2(x[3], x[3], t[3], C)

	t[4], D = bits.Add64(t[4], C, 0)

	// m = t[0]n'[0] mod W
	m = t[0] * qInvNeg

	// -----------------------------------
	// Second loop
	C = madd0(m, q0, t[0])
	C, t[0] = madd2(m, q1, t[1], C)
	C, t[1] = madd2(m, q2, t[2], C)
	C, t[2] = madd2(m, q3, t[3], C)

	t[3], C = bits.Add64(t[4], C, 0)
	t[4], _ = bits.Add64(0, D, C)

	if t[4] != 0 {
		// we need to reduce, we have a result on 5 words
		var b uint64
		z[0], b = bits.Sub64(t[0], q0, 0)
		z[1], b = bits.Sub64(t[1], q1, b)
		z[2], b = bits.Sub64(t[2], q2, b)
		z[3], _ = bits.Sub64(t[3], q3, b)
		return z
	}

	// copy t into z
	z[0] = t[0]
	z[1] = t[1]
	z[2] = t[2]
	z[3] = t[3]

	// if z ⩾ q → z -= q
	if !z.smallerThanModulus() {
		var b uint64
		z[0], b = bits.Sub64(z[0], q0, 0)
		z[1], b = bits.Sub64(z[1], q1, b)
		z[2], b = bits.Sub64(z[2], q2, b)
		z[3], _ = bits.Sub64(z[3], q3, b)
	}
	return z
}

// Butterfly sets
//
//	a = a + b (mod q)
//	b = a - b (mod q)
func Butterfly(a, b *Element) {
	_butterflyGeneric(a, b)
}
