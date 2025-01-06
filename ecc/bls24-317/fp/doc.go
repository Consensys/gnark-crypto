// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

// Package fp contains field arithmetic operations for modulus = 0x1058ca...ab2aab.
//
// The API is similar to math/big (big.Int), but the operations are significantly faster (up to 20x).
//
// Additionally fp.Vector offers an API to manipulate []Element.
//
// The modulus is hardcoded in all the operations.
//
// Field elements are represented as an array, and assumed to be in Montgomery form in all methods:
//
//	type Element [5]uint64
//
// # Usage
//
// Example API signature:
//
//	// Mul z = x * y (mod q)
//	func (z *Element) Mul(x, y *Element) *Element
//
// and can be used like so:
//
//	var a, b Element
//	a.SetUint64(2)
//	b.SetString("984896738")
//	a.Mul(a, b)
//	a.Sub(a, a)
//	 .Add(a, b)
//	 .Inv(a)
//	b.Exp(b, new(big.Int).SetUint64(42))
//
// Modulus q =
//
//	q[base10] = 136393071104295911515099765908274057061945112121419593977210139303905973197232025618026156731051
//	q[base16] = 0x1058ca226f60892cf28fc5a0b7f9d039169a61e684c73446d6f339e43424bf7e8d512e565dab2aab
//
// # Warning
//
// There is no security guarantees such as constant time implementation or side-channel attack resistance.
// This code is provided as-is. Partially audited, see https://github.com/Consensys/gnark/tree/master/audits
// for more details.
package fp
