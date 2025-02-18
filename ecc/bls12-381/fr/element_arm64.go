//go:build !purego

// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package fr

import (
	_ "github.com/consensys/gnark-crypto/field/asm/element_4w"
)

// Butterfly sets
//
//	a = a + b (mod q)
//	b = a - b (mod q)
//
//go:noescape
func Butterfly(a, b *Element)

//go:noescape
func mul(res, x, y *Element)

// Mul z = x * y (mod q)
//
// x and y must be less than q
func (z *Element) Mul(x, y *Element) *Element {
	mul(z, x, y)
	return z
}

// Square z = x * x (mod q)
//
// x must be less than q
func (z *Element) Square(x *Element) *Element {
	// see Mul for doc.
	mul(z, x, x)
	return z
}

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
		103079215080,
		2754126784762437656,
		3644466801193238400,
		2429780183658283323,
	}
	x.Mul(x, &y)
}

// MulBy13 x *= 13 (mod q)
func MulBy13(x *Element) {
	var y = Element{
		120259084260,
		15510977298029211676,
		7326335280343703402,
		5909200893219589146,
	}
	x.Mul(x, &y)
}

func fromMont(z *Element) {
	_fromMontGeneric(z)
}

//go:noescape
func reduce(res *Element)
