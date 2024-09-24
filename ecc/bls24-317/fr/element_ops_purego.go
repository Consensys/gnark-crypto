//go:build !amd64 || purego
// +build !amd64 purego

// Copyright 2020 ConsenSys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

// MulBy13 x *= 13 (mod q)
func MulBy13(x *Element) {
	var y = Element{
		18446744073709551568,
		10999079689622735090,
		16060824205876888138,
		3752826977836272504,
	}
	x.Mul(x, &y)
}

// Butterfly sets
//
//	a = a + b (mod q)
//	b = a - b (mod q)
func Butterfly(a, b *Element) {
	_butterflyGeneric(a, b)
}

func fromMont(z *Element) {
	_fromMontGeneric(z)
}

func reduce(z *Element) {
	_reduceGeneric(z)
}

// Add adds two vectors element-wise and stores the result in self.
// It panics if the vectors don't have the same length.
func (vector *Vector) Add(a, b Vector) {
	addVecGeneric(*vector, a, b)
}

// Sub subtracts two vectors element-wise and stores the result in self.
// It panics if the vectors don't have the same length.
func (vector *Vector) Sub(a, b Vector) {
	subVecGeneric(*vector, a, b)
}

// ScalarMul multiplies a vector by a scalar element-wise and stores the result in self.
// It panics if the vectors don't have the same length.
func (vector *Vector) ScalarMul(a Vector, b *Element) {
	scalarMulVecGeneric(*vector, a, b)
}

// Sum computes the sum of all elements in the vector.
func (vector *Vector) Sum() (res Element) {
	sumVecGeneric(&res, *vector)
	return
}

// Mul z = x * y (mod q)
//
// x and y must be less than q
func (z *Element) Mul(x, y *Element) *Element {

	// Algorithm 2 of "Faster Montgomery Multiplication and Multi-Scalar-Multiplication for SNARKS"
	// by Y. El Housni and G. Botrel https://doi.org/10.46586/tches.v2023.i3.504-521

	var t0, t1, t2, t3 uint64
	var u0, u1, u2, u3 uint64
	{
		var c0, c1, c2 uint64
		v := x[0]
		u0, t0 = bits.Mul64(v, y[0])
		u1, t1 = bits.Mul64(v, y[1])
		u2, t2 = bits.Mul64(v, y[2])
		u3, t3 = bits.Mul64(v, y[3])
		t1, c0 = bits.Add64(u0, t1, 0)
		t2, c0 = bits.Add64(u1, t2, c0)
		t3, c0 = bits.Add64(u2, t3, c0)
		c2, _ = bits.Add64(u3, 0, c0)

		m := qInvNeg * t0

		u0, c1 = bits.Mul64(m, q0)
		_, c0 = bits.Add64(t0, c1, 0)
		u1, c1 = bits.Mul64(m, q1)
		t0, c0 = bits.Add64(t1, c1, c0)
		u2, c1 = bits.Mul64(m, q2)
		t1, c0 = bits.Add64(t2, c1, c0)
		u3, c1 = bits.Mul64(m, q3)

		t2, c0 = bits.Add64(0, c1, c0)
		u3, _ = bits.Add64(u3, 0, c0)
		t0, c0 = bits.Add64(u0, t0, 0)
		t1, c0 = bits.Add64(u1, t1, c0)
		t2, c0 = bits.Add64(u2, t2, c0)
		c2, _ = bits.Add64(c2, 0, c0)
		t2, c0 = bits.Add64(t3, t2, 0)
		t3, _ = bits.Add64(u3, c2, c0)

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

		c2, _ = bits.Add64(0, 0, c0)
		t1, c0 = bits.Add64(u0, t1, 0)
		t2, c0 = bits.Add64(u1, t2, c0)
		t3, c0 = bits.Add64(u2, t3, c0)
		c2, _ = bits.Add64(u3, c2, c0)

		m := qInvNeg * t0

		u0, c1 = bits.Mul64(m, q0)
		_, c0 = bits.Add64(t0, c1, 0)
		u1, c1 = bits.Mul64(m, q1)
		t0, c0 = bits.Add64(t1, c1, c0)
		u2, c1 = bits.Mul64(m, q2)
		t1, c0 = bits.Add64(t2, c1, c0)
		u3, c1 = bits.Mul64(m, q3)

		t2, c0 = bits.Add64(0, c1, c0)
		u3, _ = bits.Add64(u3, 0, c0)
		t0, c0 = bits.Add64(u0, t0, 0)
		t1, c0 = bits.Add64(u1, t1, c0)
		t2, c0 = bits.Add64(u2, t2, c0)
		c2, _ = bits.Add64(c2, 0, c0)
		t2, c0 = bits.Add64(t3, t2, 0)
		t3, _ = bits.Add64(u3, c2, c0)

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

		c2, _ = bits.Add64(0, 0, c0)
		t1, c0 = bits.Add64(u0, t1, 0)
		t2, c0 = bits.Add64(u1, t2, c0)
		t3, c0 = bits.Add64(u2, t3, c0)
		c2, _ = bits.Add64(u3, c2, c0)

		m := qInvNeg * t0

		u0, c1 = bits.Mul64(m, q0)
		_, c0 = bits.Add64(t0, c1, 0)
		u1, c1 = bits.Mul64(m, q1)
		t0, c0 = bits.Add64(t1, c1, c0)
		u2, c1 = bits.Mul64(m, q2)
		t1, c0 = bits.Add64(t2, c1, c0)
		u3, c1 = bits.Mul64(m, q3)

		t2, c0 = bits.Add64(0, c1, c0)
		u3, _ = bits.Add64(u3, 0, c0)
		t0, c0 = bits.Add64(u0, t0, 0)
		t1, c0 = bits.Add64(u1, t1, c0)
		t2, c0 = bits.Add64(u2, t2, c0)
		c2, _ = bits.Add64(c2, 0, c0)
		t2, c0 = bits.Add64(t3, t2, 0)
		t3, _ = bits.Add64(u3, c2, c0)

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

		c2, _ = bits.Add64(0, 0, c0)
		t1, c0 = bits.Add64(u0, t1, 0)
		t2, c0 = bits.Add64(u1, t2, c0)
		t3, c0 = bits.Add64(u2, t3, c0)
		c2, _ = bits.Add64(u3, c2, c0)

		m := qInvNeg * t0

		u0, c1 = bits.Mul64(m, q0)
		_, c0 = bits.Add64(t0, c1, 0)
		u1, c1 = bits.Mul64(m, q1)
		t0, c0 = bits.Add64(t1, c1, c0)
		u2, c1 = bits.Mul64(m, q2)
		t1, c0 = bits.Add64(t2, c1, c0)
		u3, c1 = bits.Mul64(m, q3)

		t2, c0 = bits.Add64(0, c1, c0)
		u3, _ = bits.Add64(u3, 0, c0)
		t0, c0 = bits.Add64(u0, t0, 0)
		t1, c0 = bits.Add64(u1, t1, c0)
		t2, c0 = bits.Add64(u2, t2, c0)
		c2, _ = bits.Add64(c2, 0, c0)
		t2, c0 = bits.Add64(t3, t2, 0)
		t3, _ = bits.Add64(u3, c2, c0)

	}
	z[0] = t0
	z[1] = t1
	z[2] = t2
	z[3] = t3

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
//
// x must be less than q
func (z *Element) Square(x *Element) *Element {
	// see Mul for algorithm documentation

	var t0, t1, t2, t3 uint64
	var u0, u1, u2, u3 uint64
	{
		var c0, c1, c2 uint64
		v := x[0]
		u0, t0 = bits.Mul64(v, x[0])
		u1, t1 = bits.Mul64(v, x[1])
		u2, t2 = bits.Mul64(v, x[2])
		u3, t3 = bits.Mul64(v, x[3])
		t1, c0 = bits.Add64(u0, t1, 0)
		t2, c0 = bits.Add64(u1, t2, c0)
		t3, c0 = bits.Add64(u2, t3, c0)
		c2, _ = bits.Add64(u3, 0, c0)

		m := qInvNeg * t0

		u0, c1 = bits.Mul64(m, q0)
		_, c0 = bits.Add64(t0, c1, 0)
		u1, c1 = bits.Mul64(m, q1)
		t0, c0 = bits.Add64(t1, c1, c0)
		u2, c1 = bits.Mul64(m, q2)
		t1, c0 = bits.Add64(t2, c1, c0)
		u3, c1 = bits.Mul64(m, q3)

		t2, c0 = bits.Add64(0, c1, c0)
		u3, _ = bits.Add64(u3, 0, c0)
		t0, c0 = bits.Add64(u0, t0, 0)
		t1, c0 = bits.Add64(u1, t1, c0)
		t2, c0 = bits.Add64(u2, t2, c0)
		c2, _ = bits.Add64(c2, 0, c0)
		t2, c0 = bits.Add64(t3, t2, 0)
		t3, _ = bits.Add64(u3, c2, c0)

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

		c2, _ = bits.Add64(0, 0, c0)
		t1, c0 = bits.Add64(u0, t1, 0)
		t2, c0 = bits.Add64(u1, t2, c0)
		t3, c0 = bits.Add64(u2, t3, c0)
		c2, _ = bits.Add64(u3, c2, c0)

		m := qInvNeg * t0

		u0, c1 = bits.Mul64(m, q0)
		_, c0 = bits.Add64(t0, c1, 0)
		u1, c1 = bits.Mul64(m, q1)
		t0, c0 = bits.Add64(t1, c1, c0)
		u2, c1 = bits.Mul64(m, q2)
		t1, c0 = bits.Add64(t2, c1, c0)
		u3, c1 = bits.Mul64(m, q3)

		t2, c0 = bits.Add64(0, c1, c0)
		u3, _ = bits.Add64(u3, 0, c0)
		t0, c0 = bits.Add64(u0, t0, 0)
		t1, c0 = bits.Add64(u1, t1, c0)
		t2, c0 = bits.Add64(u2, t2, c0)
		c2, _ = bits.Add64(c2, 0, c0)
		t2, c0 = bits.Add64(t3, t2, 0)
		t3, _ = bits.Add64(u3, c2, c0)

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

		c2, _ = bits.Add64(0, 0, c0)
		t1, c0 = bits.Add64(u0, t1, 0)
		t2, c0 = bits.Add64(u1, t2, c0)
		t3, c0 = bits.Add64(u2, t3, c0)
		c2, _ = bits.Add64(u3, c2, c0)

		m := qInvNeg * t0

		u0, c1 = bits.Mul64(m, q0)
		_, c0 = bits.Add64(t0, c1, 0)
		u1, c1 = bits.Mul64(m, q1)
		t0, c0 = bits.Add64(t1, c1, c0)
		u2, c1 = bits.Mul64(m, q2)
		t1, c0 = bits.Add64(t2, c1, c0)
		u3, c1 = bits.Mul64(m, q3)

		t2, c0 = bits.Add64(0, c1, c0)
		u3, _ = bits.Add64(u3, 0, c0)
		t0, c0 = bits.Add64(u0, t0, 0)
		t1, c0 = bits.Add64(u1, t1, c0)
		t2, c0 = bits.Add64(u2, t2, c0)
		c2, _ = bits.Add64(c2, 0, c0)
		t2, c0 = bits.Add64(t3, t2, 0)
		t3, _ = bits.Add64(u3, c2, c0)

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

		c2, _ = bits.Add64(0, 0, c0)
		t1, c0 = bits.Add64(u0, t1, 0)
		t2, c0 = bits.Add64(u1, t2, c0)
		t3, c0 = bits.Add64(u2, t3, c0)
		c2, _ = bits.Add64(u3, c2, c0)

		m := qInvNeg * t0

		u0, c1 = bits.Mul64(m, q0)
		_, c0 = bits.Add64(t0, c1, 0)
		u1, c1 = bits.Mul64(m, q1)
		t0, c0 = bits.Add64(t1, c1, c0)
		u2, c1 = bits.Mul64(m, q2)
		t1, c0 = bits.Add64(t2, c1, c0)
		u3, c1 = bits.Mul64(m, q3)

		t2, c0 = bits.Add64(0, c1, c0)
		u3, _ = bits.Add64(u3, 0, c0)
		t0, c0 = bits.Add64(u0, t0, 0)
		t1, c0 = bits.Add64(u1, t1, c0)
		t2, c0 = bits.Add64(u2, t2, c0)
		c2, _ = bits.Add64(c2, 0, c0)
		t2, c0 = bits.Add64(t3, t2, 0)
		t3, _ = bits.Add64(u3, c2, c0)

	}
	z[0] = t0
	z[1] = t1
	z[2] = t2
	z[3] = t3

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
