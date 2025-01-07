// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package babybear

// MulBy3 x *= 3 (mod q)
func MulBy3(x *Element) {
	var y Element
	y.SetUint64(3)
	x.Mul(x, &y)
}

// MulBy5 x *= 5 (mod q)
func MulBy5(x *Element) {
	var y Element
	y.SetUint64(5)
	x.Mul(x, &y)
}

// MulBy13 x *= 13 (mod q)
func MulBy13(x *Element) {
	var y Element
	y.SetUint64(13)
	x.Mul(x, &y)
}

func fromMont(z *Element) {
	_fromMontGeneric(z)
}

func reduce(z *Element) {
	_reduceGeneric(z)
}
func montReduce(v uint64) uint32 {
	m := uint32(v) * qInvNeg
	t := uint32((v + uint64(m)*q) >> 32)
	if t >= q {
		t -= q
	}
	return t
}

// Mul z = x * y (mod q)
//
// x and y must be less than q
func (z *Element) Mul(x, y *Element) *Element {
	v := uint64(x[0]) * uint64(y[0])
	z[0] = montReduce(v)
	return z
}

// Square z = x * x (mod q)
//
// x must be less than q
func (z *Element) Square(x *Element) *Element {
	// see Mul for algorithm documentation
	v := uint64(x[0]) * uint64(x[0])
	z[0] = montReduce(v)
	return z
}

// Butterfly sets
//
//	a = a + b (mod q)
//	b = a - b (mod q)
func Butterfly(a, b *Element) {
	_butterflyGeneric(a, b)
}
