// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package fp

// expBySqrtExp is equivalent to z.Exp(x, 400000000000008)
//
// uses github.com/mmcloughlin/addchain v0.4.0 to generate a shorter addition chain
func (z *Element) expBySqrtExp(x Element) *Element {
	// addition chain:
	//
	//	return  (1 << 55 + 1) << 3
	//
	// Operations: 58 squares 1 multiplies

	// Allocate Temporaries.
	var ()

	// var
	// Step 55: z = x^0x80000000000000
	z.Square(&x)
	for s := 1; s < 55; s++ {
		z.Square(z)
	}

	// Step 56: z = x^0x80000000000001
	z.Mul(&x, z)

	// Step 59: z = x^0x400000000000008
	for s := 0; s < 3; s++ {
		z.Square(z)
	}

	return z
}

// expByLegendreExp is equivalent to z.Exp(x, 400000000000008800000000000000000000000000000000000000000000000)
//
// uses github.com/mmcloughlin/addchain v0.4.0 to generate a shorter addition chain
func (z *Element) expByLegendreExp(x Element) *Element {
	// addition chain:
	//
	//	return  ((1 << 55 + 1) << 4 + 1) << 191
	//
	// Operations: 250 squares 2 multiplies

	// Allocate Temporaries.
	var ()

	// var
	// Step 55: z = x^0x80000000000000
	z.Square(&x)
	for s := 1; s < 55; s++ {
		z.Square(z)
	}

	// Step 56: z = x^0x80000000000001
	z.Mul(&x, z)

	// Step 60: z = x^0x800000000000010
	for s := 0; s < 4; s++ {
		z.Square(z)
	}

	// Step 61: z = x^0x800000000000011
	z.Mul(&x, z)

	// Step 252: z = x^0x400000000000008800000000000000000000000000000000000000000000000
	for s := 0; s < 191; s++ {
		z.Square(z)
	}

	return z
}
