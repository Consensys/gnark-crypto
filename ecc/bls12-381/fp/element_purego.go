//go:build purego || (!amd64 && !arm64)

// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package fp

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
		9830232086645309404,
		1112389714365644829,
		8603885298299447491,
		11361495444721768256,
		5788602283869803809,
		543934104870762216,
	}
	x.Mul(x, &y)
}

// MulBy13 x *= 13 (mod q)
func MulBy13(x *Element) {
	var y = Element{
		13438459813099623723,
		14459933216667336738,
		14900020990258308116,
		2941282712809091851,
		13639094935183769893,
		1835248516986607988,
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
//
// x and y must be less than q
func (z *Element) Mul(x, y *Element) *Element {

	// Algorithm 2 of "Faster Montgomery Multiplication and Multi-Scalar-Multiplication for SNARKS"
	// by Y. El Housni and G. Botrel https://doi.org/10.46586/tches.v2023.i3.504-521

	var t0, t1, t2, t3, t4, t5 uint64
	var u0, u1, u2, u3, u4, u5 uint64
	{
		var c0, c1, c2 uint64
		v := x[0]
		u0, t0 = bits.Mul64(v, y[0])
		u1, t1 = bits.Mul64(v, y[1])
		u2, t2 = bits.Mul64(v, y[2])
		u3, t3 = bits.Mul64(v, y[3])
		u4, t4 = bits.Mul64(v, y[4])
		u5, t5 = bits.Mul64(v, y[5])
		t1, c0 = bits.Add64(u0, t1, 0)
		t2, c0 = bits.Add64(u1, t2, c0)
		t3, c0 = bits.Add64(u2, t3, c0)
		t4, c0 = bits.Add64(u3, t4, c0)
		t5, c0 = bits.Add64(u4, t5, c0)
		c2, _ = bits.Add64(u5, 0, c0)

		m := qInvNeg * t0

		u0, c1 = bits.Mul64(m, q0)
		_, c0 = bits.Add64(t0, c1, 0)
		u1, c1 = bits.Mul64(m, q1)
		t0, c0 = bits.Add64(t1, c1, c0)
		u2, c1 = bits.Mul64(m, q2)
		t1, c0 = bits.Add64(t2, c1, c0)
		u3, c1 = bits.Mul64(m, q3)
		t2, c0 = bits.Add64(t3, c1, c0)
		u4, c1 = bits.Mul64(m, q4)
		t3, c0 = bits.Add64(t4, c1, c0)
		u5, c1 = bits.Mul64(m, q5)

		t4, c0 = bits.Add64(0, c1, c0)
		u5, _ = bits.Add64(u5, 0, c0)
		t0, c0 = bits.Add64(u0, t0, 0)
		t1, c0 = bits.Add64(u1, t1, c0)
		t2, c0 = bits.Add64(u2, t2, c0)
		t3, c0 = bits.Add64(u3, t3, c0)
		t4, c0 = bits.Add64(u4, t4, c0)
		c2, _ = bits.Add64(c2, 0, c0)
		t4, c0 = bits.Add64(t5, t4, 0)
		t5, _ = bits.Add64(u5, c2, c0)

	}
	{
		var c0, c1, c2 uint64
		v := x[1]
		u0, c1 = bits.Mul64(v, y[0])
		t0, c0 = bits.Add64(c1, t0, 0)
		u1, c1 = bits.Mul64(v, y[1])
		t1, c0 = bits.Add64(c1, t1, c0)
		u2, c1 = bits.Mul64(v, y[2])
		t2, c0 = bits.Add64(c1, t2, c0)
		u3, c1 = bits.Mul64(v, y[3])
		t3, c0 = bits.Add64(c1, t3, c0)
		u4, c1 = bits.Mul64(v, y[4])
		t4, c0 = bits.Add64(c1, t4, c0)
		u5, c1 = bits.Mul64(v, y[5])
		t5, c0 = bits.Add64(c1, t5, c0)

		c2, _ = bits.Add64(0, 0, c0)
		t1, c0 = bits.Add64(u0, t1, 0)
		t2, c0 = bits.Add64(u1, t2, c0)
		t3, c0 = bits.Add64(u2, t3, c0)
		t4, c0 = bits.Add64(u3, t4, c0)
		t5, c0 = bits.Add64(u4, t5, c0)
		c2, _ = bits.Add64(u5, c2, c0)

		m := qInvNeg * t0

		u0, c1 = bits.Mul64(m, q0)
		_, c0 = bits.Add64(t0, c1, 0)
		u1, c1 = bits.Mul64(m, q1)
		t0, c0 = bits.Add64(t1, c1, c0)
		u2, c1 = bits.Mul64(m, q2)
		t1, c0 = bits.Add64(t2, c1, c0)
		u3, c1 = bits.Mul64(m, q3)
		t2, c0 = bits.Add64(t3, c1, c0)
		u4, c1 = bits.Mul64(m, q4)
		t3, c0 = bits.Add64(t4, c1, c0)
		u5, c1 = bits.Mul64(m, q5)

		t4, c0 = bits.Add64(0, c1, c0)
		u5, _ = bits.Add64(u5, 0, c0)
		t0, c0 = bits.Add64(u0, t0, 0)
		t1, c0 = bits.Add64(u1, t1, c0)
		t2, c0 = bits.Add64(u2, t2, c0)
		t3, c0 = bits.Add64(u3, t3, c0)
		t4, c0 = bits.Add64(u4, t4, c0)
		c2, _ = bits.Add64(c2, 0, c0)
		t4, c0 = bits.Add64(t5, t4, 0)
		t5, _ = bits.Add64(u5, c2, c0)

	}
	{
		var c0, c1, c2 uint64
		v := x[2]
		u0, c1 = bits.Mul64(v, y[0])
		t0, c0 = bits.Add64(c1, t0, 0)
		u1, c1 = bits.Mul64(v, y[1])
		t1, c0 = bits.Add64(c1, t1, c0)
		u2, c1 = bits.Mul64(v, y[2])
		t2, c0 = bits.Add64(c1, t2, c0)
		u3, c1 = bits.Mul64(v, y[3])
		t3, c0 = bits.Add64(c1, t3, c0)
		u4, c1 = bits.Mul64(v, y[4])
		t4, c0 = bits.Add64(c1, t4, c0)
		u5, c1 = bits.Mul64(v, y[5])
		t5, c0 = bits.Add64(c1, t5, c0)

		c2, _ = bits.Add64(0, 0, c0)
		t1, c0 = bits.Add64(u0, t1, 0)
		t2, c0 = bits.Add64(u1, t2, c0)
		t3, c0 = bits.Add64(u2, t3, c0)
		t4, c0 = bits.Add64(u3, t4, c0)
		t5, c0 = bits.Add64(u4, t5, c0)
		c2, _ = bits.Add64(u5, c2, c0)

		m := qInvNeg * t0

		u0, c1 = bits.Mul64(m, q0)
		_, c0 = bits.Add64(t0, c1, 0)
		u1, c1 = bits.Mul64(m, q1)
		t0, c0 = bits.Add64(t1, c1, c0)
		u2, c1 = bits.Mul64(m, q2)
		t1, c0 = bits.Add64(t2, c1, c0)
		u3, c1 = bits.Mul64(m, q3)
		t2, c0 = bits.Add64(t3, c1, c0)
		u4, c1 = bits.Mul64(m, q4)
		t3, c0 = bits.Add64(t4, c1, c0)
		u5, c1 = bits.Mul64(m, q5)

		t4, c0 = bits.Add64(0, c1, c0)
		u5, _ = bits.Add64(u5, 0, c0)
		t0, c0 = bits.Add64(u0, t0, 0)
		t1, c0 = bits.Add64(u1, t1, c0)
		t2, c0 = bits.Add64(u2, t2, c0)
		t3, c0 = bits.Add64(u3, t3, c0)
		t4, c0 = bits.Add64(u4, t4, c0)
		c2, _ = bits.Add64(c2, 0, c0)
		t4, c0 = bits.Add64(t5, t4, 0)
		t5, _ = bits.Add64(u5, c2, c0)

	}
	{
		var c0, c1, c2 uint64
		v := x[3]
		u0, c1 = bits.Mul64(v, y[0])
		t0, c0 = bits.Add64(c1, t0, 0)
		u1, c1 = bits.Mul64(v, y[1])
		t1, c0 = bits.Add64(c1, t1, c0)
		u2, c1 = bits.Mul64(v, y[2])
		t2, c0 = bits.Add64(c1, t2, c0)
		u3, c1 = bits.Mul64(v, y[3])
		t3, c0 = bits.Add64(c1, t3, c0)
		u4, c1 = bits.Mul64(v, y[4])
		t4, c0 = bits.Add64(c1, t4, c0)
		u5, c1 = bits.Mul64(v, y[5])
		t5, c0 = bits.Add64(c1, t5, c0)

		c2, _ = bits.Add64(0, 0, c0)
		t1, c0 = bits.Add64(u0, t1, 0)
		t2, c0 = bits.Add64(u1, t2, c0)
		t3, c0 = bits.Add64(u2, t3, c0)
		t4, c0 = bits.Add64(u3, t4, c0)
		t5, c0 = bits.Add64(u4, t5, c0)
		c2, _ = bits.Add64(u5, c2, c0)

		m := qInvNeg * t0

		u0, c1 = bits.Mul64(m, q0)
		_, c0 = bits.Add64(t0, c1, 0)
		u1, c1 = bits.Mul64(m, q1)
		t0, c0 = bits.Add64(t1, c1, c0)
		u2, c1 = bits.Mul64(m, q2)
		t1, c0 = bits.Add64(t2, c1, c0)
		u3, c1 = bits.Mul64(m, q3)
		t2, c0 = bits.Add64(t3, c1, c0)
		u4, c1 = bits.Mul64(m, q4)
		t3, c0 = bits.Add64(t4, c1, c0)
		u5, c1 = bits.Mul64(m, q5)

		t4, c0 = bits.Add64(0, c1, c0)
		u5, _ = bits.Add64(u5, 0, c0)
		t0, c0 = bits.Add64(u0, t0, 0)
		t1, c0 = bits.Add64(u1, t1, c0)
		t2, c0 = bits.Add64(u2, t2, c0)
		t3, c0 = bits.Add64(u3, t3, c0)
		t4, c0 = bits.Add64(u4, t4, c0)
		c2, _ = bits.Add64(c2, 0, c0)
		t4, c0 = bits.Add64(t5, t4, 0)
		t5, _ = bits.Add64(u5, c2, c0)

	}
	{
		var c0, c1, c2 uint64
		v := x[4]
		u0, c1 = bits.Mul64(v, y[0])
		t0, c0 = bits.Add64(c1, t0, 0)
		u1, c1 = bits.Mul64(v, y[1])
		t1, c0 = bits.Add64(c1, t1, c0)
		u2, c1 = bits.Mul64(v, y[2])
		t2, c0 = bits.Add64(c1, t2, c0)
		u3, c1 = bits.Mul64(v, y[3])
		t3, c0 = bits.Add64(c1, t3, c0)
		u4, c1 = bits.Mul64(v, y[4])
		t4, c0 = bits.Add64(c1, t4, c0)
		u5, c1 = bits.Mul64(v, y[5])
		t5, c0 = bits.Add64(c1, t5, c0)

		c2, _ = bits.Add64(0, 0, c0)
		t1, c0 = bits.Add64(u0, t1, 0)
		t2, c0 = bits.Add64(u1, t2, c0)
		t3, c0 = bits.Add64(u2, t3, c0)
		t4, c0 = bits.Add64(u3, t4, c0)
		t5, c0 = bits.Add64(u4, t5, c0)
		c2, _ = bits.Add64(u5, c2, c0)

		m := qInvNeg * t0

		u0, c1 = bits.Mul64(m, q0)
		_, c0 = bits.Add64(t0, c1, 0)
		u1, c1 = bits.Mul64(m, q1)
		t0, c0 = bits.Add64(t1, c1, c0)
		u2, c1 = bits.Mul64(m, q2)
		t1, c0 = bits.Add64(t2, c1, c0)
		u3, c1 = bits.Mul64(m, q3)
		t2, c0 = bits.Add64(t3, c1, c0)
		u4, c1 = bits.Mul64(m, q4)
		t3, c0 = bits.Add64(t4, c1, c0)
		u5, c1 = bits.Mul64(m, q5)

		t4, c0 = bits.Add64(0, c1, c0)
		u5, _ = bits.Add64(u5, 0, c0)
		t0, c0 = bits.Add64(u0, t0, 0)
		t1, c0 = bits.Add64(u1, t1, c0)
		t2, c0 = bits.Add64(u2, t2, c0)
		t3, c0 = bits.Add64(u3, t3, c0)
		t4, c0 = bits.Add64(u4, t4, c0)
		c2, _ = bits.Add64(c2, 0, c0)
		t4, c0 = bits.Add64(t5, t4, 0)
		t5, _ = bits.Add64(u5, c2, c0)

	}
	{
		var c0, c1, c2 uint64
		v := x[5]
		u0, c1 = bits.Mul64(v, y[0])
		t0, c0 = bits.Add64(c1, t0, 0)
		u1, c1 = bits.Mul64(v, y[1])
		t1, c0 = bits.Add64(c1, t1, c0)
		u2, c1 = bits.Mul64(v, y[2])
		t2, c0 = bits.Add64(c1, t2, c0)
		u3, c1 = bits.Mul64(v, y[3])
		t3, c0 = bits.Add64(c1, t3, c0)
		u4, c1 = bits.Mul64(v, y[4])
		t4, c0 = bits.Add64(c1, t4, c0)
		u5, c1 = bits.Mul64(v, y[5])
		t5, c0 = bits.Add64(c1, t5, c0)

		c2, _ = bits.Add64(0, 0, c0)
		t1, c0 = bits.Add64(u0, t1, 0)
		t2, c0 = bits.Add64(u1, t2, c0)
		t3, c0 = bits.Add64(u2, t3, c0)
		t4, c0 = bits.Add64(u3, t4, c0)
		t5, c0 = bits.Add64(u4, t5, c0)
		c2, _ = bits.Add64(u5, c2, c0)

		m := qInvNeg * t0

		u0, c1 = bits.Mul64(m, q0)
		_, c0 = bits.Add64(t0, c1, 0)
		u1, c1 = bits.Mul64(m, q1)
		t0, c0 = bits.Add64(t1, c1, c0)
		u2, c1 = bits.Mul64(m, q2)
		t1, c0 = bits.Add64(t2, c1, c0)
		u3, c1 = bits.Mul64(m, q3)
		t2, c0 = bits.Add64(t3, c1, c0)
		u4, c1 = bits.Mul64(m, q4)
		t3, c0 = bits.Add64(t4, c1, c0)
		u5, c1 = bits.Mul64(m, q5)

		t4, c0 = bits.Add64(0, c1, c0)
		u5, _ = bits.Add64(u5, 0, c0)
		t0, c0 = bits.Add64(u0, t0, 0)
		t1, c0 = bits.Add64(u1, t1, c0)
		t2, c0 = bits.Add64(u2, t2, c0)
		t3, c0 = bits.Add64(u3, t3, c0)
		t4, c0 = bits.Add64(u4, t4, c0)
		c2, _ = bits.Add64(c2, 0, c0)
		t4, c0 = bits.Add64(t5, t4, 0)
		t5, _ = bits.Add64(u5, c2, c0)

	}
	z[0] = t0
	z[1] = t1
	z[2] = t2
	z[3] = t3
	z[4] = t4
	z[5] = t5

	// if z ⩾ q → z -= q
	if !z.smallerThanModulus() {
		var b uint64
		z[0], b = bits.Sub64(z[0], q0, 0)
		z[1], b = bits.Sub64(z[1], q1, b)
		z[2], b = bits.Sub64(z[2], q2, b)
		z[3], b = bits.Sub64(z[3], q3, b)
		z[4], b = bits.Sub64(z[4], q4, b)
		z[5], _ = bits.Sub64(z[5], q5, b)
	}
	return z
}

// Square z = x * x (mod q)
//
// x must be less than q
func (z *Element) Square(x *Element) *Element {
	// see Mul for algorithm documentation

	var t0, t1, t2, t3, t4, t5 uint64
	var u0, u1, u2, u3, u4, u5 uint64
	{
		var c0, c1, c2 uint64
		v := x[0]
		u0, t0 = bits.Mul64(v, x[0])
		u1, t1 = bits.Mul64(v, x[1])
		u2, t2 = bits.Mul64(v, x[2])
		u3, t3 = bits.Mul64(v, x[3])
		u4, t4 = bits.Mul64(v, x[4])
		u5, t5 = bits.Mul64(v, x[5])
		t1, c0 = bits.Add64(u0, t1, 0)
		t2, c0 = bits.Add64(u1, t2, c0)
		t3, c0 = bits.Add64(u2, t3, c0)
		t4, c0 = bits.Add64(u3, t4, c0)
		t5, c0 = bits.Add64(u4, t5, c0)
		c2, _ = bits.Add64(u5, 0, c0)

		m := qInvNeg * t0

		u0, c1 = bits.Mul64(m, q0)
		_, c0 = bits.Add64(t0, c1, 0)
		u1, c1 = bits.Mul64(m, q1)
		t0, c0 = bits.Add64(t1, c1, c0)
		u2, c1 = bits.Mul64(m, q2)
		t1, c0 = bits.Add64(t2, c1, c0)
		u3, c1 = bits.Mul64(m, q3)
		t2, c0 = bits.Add64(t3, c1, c0)
		u4, c1 = bits.Mul64(m, q4)
		t3, c0 = bits.Add64(t4, c1, c0)
		u5, c1 = bits.Mul64(m, q5)

		t4, c0 = bits.Add64(0, c1, c0)
		u5, _ = bits.Add64(u5, 0, c0)
		t0, c0 = bits.Add64(u0, t0, 0)
		t1, c0 = bits.Add64(u1, t1, c0)
		t2, c0 = bits.Add64(u2, t2, c0)
		t3, c0 = bits.Add64(u3, t3, c0)
		t4, c0 = bits.Add64(u4, t4, c0)
		c2, _ = bits.Add64(c2, 0, c0)
		t4, c0 = bits.Add64(t5, t4, 0)
		t5, _ = bits.Add64(u5, c2, c0)

	}
	{
		var c0, c1, c2 uint64
		v := x[1]
		u0, c1 = bits.Mul64(v, x[0])
		t0, c0 = bits.Add64(c1, t0, 0)
		u1, c1 = bits.Mul64(v, x[1])
		t1, c0 = bits.Add64(c1, t1, c0)
		u2, c1 = bits.Mul64(v, x[2])
		t2, c0 = bits.Add64(c1, t2, c0)
		u3, c1 = bits.Mul64(v, x[3])
		t3, c0 = bits.Add64(c1, t3, c0)
		u4, c1 = bits.Mul64(v, x[4])
		t4, c0 = bits.Add64(c1, t4, c0)
		u5, c1 = bits.Mul64(v, x[5])
		t5, c0 = bits.Add64(c1, t5, c0)

		c2, _ = bits.Add64(0, 0, c0)
		t1, c0 = bits.Add64(u0, t1, 0)
		t2, c0 = bits.Add64(u1, t2, c0)
		t3, c0 = bits.Add64(u2, t3, c0)
		t4, c0 = bits.Add64(u3, t4, c0)
		t5, c0 = bits.Add64(u4, t5, c0)
		c2, _ = bits.Add64(u5, c2, c0)

		m := qInvNeg * t0

		u0, c1 = bits.Mul64(m, q0)
		_, c0 = bits.Add64(t0, c1, 0)
		u1, c1 = bits.Mul64(m, q1)
		t0, c0 = bits.Add64(t1, c1, c0)
		u2, c1 = bits.Mul64(m, q2)
		t1, c0 = bits.Add64(t2, c1, c0)
		u3, c1 = bits.Mul64(m, q3)
		t2, c0 = bits.Add64(t3, c1, c0)
		u4, c1 = bits.Mul64(m, q4)
		t3, c0 = bits.Add64(t4, c1, c0)
		u5, c1 = bits.Mul64(m, q5)

		t4, c0 = bits.Add64(0, c1, c0)
		u5, _ = bits.Add64(u5, 0, c0)
		t0, c0 = bits.Add64(u0, t0, 0)
		t1, c0 = bits.Add64(u1, t1, c0)
		t2, c0 = bits.Add64(u2, t2, c0)
		t3, c0 = bits.Add64(u3, t3, c0)
		t4, c0 = bits.Add64(u4, t4, c0)
		c2, _ = bits.Add64(c2, 0, c0)
		t4, c0 = bits.Add64(t5, t4, 0)
		t5, _ = bits.Add64(u5, c2, c0)

	}
	{
		var c0, c1, c2 uint64
		v := x[2]
		u0, c1 = bits.Mul64(v, x[0])
		t0, c0 = bits.Add64(c1, t0, 0)
		u1, c1 = bits.Mul64(v, x[1])
		t1, c0 = bits.Add64(c1, t1, c0)
		u2, c1 = bits.Mul64(v, x[2])
		t2, c0 = bits.Add64(c1, t2, c0)
		u3, c1 = bits.Mul64(v, x[3])
		t3, c0 = bits.Add64(c1, t3, c0)
		u4, c1 = bits.Mul64(v, x[4])
		t4, c0 = bits.Add64(c1, t4, c0)
		u5, c1 = bits.Mul64(v, x[5])
		t5, c0 = bits.Add64(c1, t5, c0)

		c2, _ = bits.Add64(0, 0, c0)
		t1, c0 = bits.Add64(u0, t1, 0)
		t2, c0 = bits.Add64(u1, t2, c0)
		t3, c0 = bits.Add64(u2, t3, c0)
		t4, c0 = bits.Add64(u3, t4, c0)
		t5, c0 = bits.Add64(u4, t5, c0)
		c2, _ = bits.Add64(u5, c2, c0)

		m := qInvNeg * t0

		u0, c1 = bits.Mul64(m, q0)
		_, c0 = bits.Add64(t0, c1, 0)
		u1, c1 = bits.Mul64(m, q1)
		t0, c0 = bits.Add64(t1, c1, c0)
		u2, c1 = bits.Mul64(m, q2)
		t1, c0 = bits.Add64(t2, c1, c0)
		u3, c1 = bits.Mul64(m, q3)
		t2, c0 = bits.Add64(t3, c1, c0)
		u4, c1 = bits.Mul64(m, q4)
		t3, c0 = bits.Add64(t4, c1, c0)
		u5, c1 = bits.Mul64(m, q5)

		t4, c0 = bits.Add64(0, c1, c0)
		u5, _ = bits.Add64(u5, 0, c0)
		t0, c0 = bits.Add64(u0, t0, 0)
		t1, c0 = bits.Add64(u1, t1, c0)
		t2, c0 = bits.Add64(u2, t2, c0)
		t3, c0 = bits.Add64(u3, t3, c0)
		t4, c0 = bits.Add64(u4, t4, c0)
		c2, _ = bits.Add64(c2, 0, c0)
		t4, c0 = bits.Add64(t5, t4, 0)
		t5, _ = bits.Add64(u5, c2, c0)

	}
	{
		var c0, c1, c2 uint64
		v := x[3]
		u0, c1 = bits.Mul64(v, x[0])
		t0, c0 = bits.Add64(c1, t0, 0)
		u1, c1 = bits.Mul64(v, x[1])
		t1, c0 = bits.Add64(c1, t1, c0)
		u2, c1 = bits.Mul64(v, x[2])
		t2, c0 = bits.Add64(c1, t2, c0)
		u3, c1 = bits.Mul64(v, x[3])
		t3, c0 = bits.Add64(c1, t3, c0)
		u4, c1 = bits.Mul64(v, x[4])
		t4, c0 = bits.Add64(c1, t4, c0)
		u5, c1 = bits.Mul64(v, x[5])
		t5, c0 = bits.Add64(c1, t5, c0)

		c2, _ = bits.Add64(0, 0, c0)
		t1, c0 = bits.Add64(u0, t1, 0)
		t2, c0 = bits.Add64(u1, t2, c0)
		t3, c0 = bits.Add64(u2, t3, c0)
		t4, c0 = bits.Add64(u3, t4, c0)
		t5, c0 = bits.Add64(u4, t5, c0)
		c2, _ = bits.Add64(u5, c2, c0)

		m := qInvNeg * t0

		u0, c1 = bits.Mul64(m, q0)
		_, c0 = bits.Add64(t0, c1, 0)
		u1, c1 = bits.Mul64(m, q1)
		t0, c0 = bits.Add64(t1, c1, c0)
		u2, c1 = bits.Mul64(m, q2)
		t1, c0 = bits.Add64(t2, c1, c0)
		u3, c1 = bits.Mul64(m, q3)
		t2, c0 = bits.Add64(t3, c1, c0)
		u4, c1 = bits.Mul64(m, q4)
		t3, c0 = bits.Add64(t4, c1, c0)
		u5, c1 = bits.Mul64(m, q5)

		t4, c0 = bits.Add64(0, c1, c0)
		u5, _ = bits.Add64(u5, 0, c0)
		t0, c0 = bits.Add64(u0, t0, 0)
		t1, c0 = bits.Add64(u1, t1, c0)
		t2, c0 = bits.Add64(u2, t2, c0)
		t3, c0 = bits.Add64(u3, t3, c0)
		t4, c0 = bits.Add64(u4, t4, c0)
		c2, _ = bits.Add64(c2, 0, c0)
		t4, c0 = bits.Add64(t5, t4, 0)
		t5, _ = bits.Add64(u5, c2, c0)

	}
	{
		var c0, c1, c2 uint64
		v := x[4]
		u0, c1 = bits.Mul64(v, x[0])
		t0, c0 = bits.Add64(c1, t0, 0)
		u1, c1 = bits.Mul64(v, x[1])
		t1, c0 = bits.Add64(c1, t1, c0)
		u2, c1 = bits.Mul64(v, x[2])
		t2, c0 = bits.Add64(c1, t2, c0)
		u3, c1 = bits.Mul64(v, x[3])
		t3, c0 = bits.Add64(c1, t3, c0)
		u4, c1 = bits.Mul64(v, x[4])
		t4, c0 = bits.Add64(c1, t4, c0)
		u5, c1 = bits.Mul64(v, x[5])
		t5, c0 = bits.Add64(c1, t5, c0)

		c2, _ = bits.Add64(0, 0, c0)
		t1, c0 = bits.Add64(u0, t1, 0)
		t2, c0 = bits.Add64(u1, t2, c0)
		t3, c0 = bits.Add64(u2, t3, c0)
		t4, c0 = bits.Add64(u3, t4, c0)
		t5, c0 = bits.Add64(u4, t5, c0)
		c2, _ = bits.Add64(u5, c2, c0)

		m := qInvNeg * t0

		u0, c1 = bits.Mul64(m, q0)
		_, c0 = bits.Add64(t0, c1, 0)
		u1, c1 = bits.Mul64(m, q1)
		t0, c0 = bits.Add64(t1, c1, c0)
		u2, c1 = bits.Mul64(m, q2)
		t1, c0 = bits.Add64(t2, c1, c0)
		u3, c1 = bits.Mul64(m, q3)
		t2, c0 = bits.Add64(t3, c1, c0)
		u4, c1 = bits.Mul64(m, q4)
		t3, c0 = bits.Add64(t4, c1, c0)
		u5, c1 = bits.Mul64(m, q5)

		t4, c0 = bits.Add64(0, c1, c0)
		u5, _ = bits.Add64(u5, 0, c0)
		t0, c0 = bits.Add64(u0, t0, 0)
		t1, c0 = bits.Add64(u1, t1, c0)
		t2, c0 = bits.Add64(u2, t2, c0)
		t3, c0 = bits.Add64(u3, t3, c0)
		t4, c0 = bits.Add64(u4, t4, c0)
		c2, _ = bits.Add64(c2, 0, c0)
		t4, c0 = bits.Add64(t5, t4, 0)
		t5, _ = bits.Add64(u5, c2, c0)

	}
	{
		var c0, c1, c2 uint64
		v := x[5]
		u0, c1 = bits.Mul64(v, x[0])
		t0, c0 = bits.Add64(c1, t0, 0)
		u1, c1 = bits.Mul64(v, x[1])
		t1, c0 = bits.Add64(c1, t1, c0)
		u2, c1 = bits.Mul64(v, x[2])
		t2, c0 = bits.Add64(c1, t2, c0)
		u3, c1 = bits.Mul64(v, x[3])
		t3, c0 = bits.Add64(c1, t3, c0)
		u4, c1 = bits.Mul64(v, x[4])
		t4, c0 = bits.Add64(c1, t4, c0)
		u5, c1 = bits.Mul64(v, x[5])
		t5, c0 = bits.Add64(c1, t5, c0)

		c2, _ = bits.Add64(0, 0, c0)
		t1, c0 = bits.Add64(u0, t1, 0)
		t2, c0 = bits.Add64(u1, t2, c0)
		t3, c0 = bits.Add64(u2, t3, c0)
		t4, c0 = bits.Add64(u3, t4, c0)
		t5, c0 = bits.Add64(u4, t5, c0)
		c2, _ = bits.Add64(u5, c2, c0)

		m := qInvNeg * t0

		u0, c1 = bits.Mul64(m, q0)
		_, c0 = bits.Add64(t0, c1, 0)
		u1, c1 = bits.Mul64(m, q1)
		t0, c0 = bits.Add64(t1, c1, c0)
		u2, c1 = bits.Mul64(m, q2)
		t1, c0 = bits.Add64(t2, c1, c0)
		u3, c1 = bits.Mul64(m, q3)
		t2, c0 = bits.Add64(t3, c1, c0)
		u4, c1 = bits.Mul64(m, q4)
		t3, c0 = bits.Add64(t4, c1, c0)
		u5, c1 = bits.Mul64(m, q5)

		t4, c0 = bits.Add64(0, c1, c0)
		u5, _ = bits.Add64(u5, 0, c0)
		t0, c0 = bits.Add64(u0, t0, 0)
		t1, c0 = bits.Add64(u1, t1, c0)
		t2, c0 = bits.Add64(u2, t2, c0)
		t3, c0 = bits.Add64(u3, t3, c0)
		t4, c0 = bits.Add64(u4, t4, c0)
		c2, _ = bits.Add64(c2, 0, c0)
		t4, c0 = bits.Add64(t5, t4, 0)
		t5, _ = bits.Add64(u5, c2, c0)

	}
	z[0] = t0
	z[1] = t1
	z[2] = t2
	z[3] = t3
	z[4] = t4
	z[5] = t5

	// if z ⩾ q → z -= q
	if !z.smallerThanModulus() {
		var b uint64
		z[0], b = bits.Sub64(z[0], q0, 0)
		z[1], b = bits.Sub64(z[1], q1, b)
		z[2], b = bits.Sub64(z[2], q2, b)
		z[3], b = bits.Sub64(z[3], q3, b)
		z[4], b = bits.Sub64(z[4], q4, b)
		z[5], _ = bits.Sub64(z[5], q5, b)
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
