// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

// Package fp contains field arithmetic operations for modulus = 0x122e82...00008b.
//
// The API is similar to math/big (big.Int), but the operations are significantly faster (up to 20x).
//
// Additionally fp.Vector offers an API to manipulate []Element.
//
// The modulus is hardcoded in all the operations.
//
// Field elements are represented as an array, and assumed to be in Montgomery form in all methods:
//
//	type Element [12]uint64
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
//	q[base10] = 6891450384315732539396789682275657542479668912536150109513790160209623422243491736087683183289411687640864567753786613451161759120554247759349511699125301598951605099378508850372543631423596795951899700429969112842764913119068299
//	q[base16] = 0x122e824fb83ce0ad187c94004faff3eb926186a81d14688528275ef8087be41707ba638e584e91903cebaff25b423048689c8ed12f9fd9071dcd3dc73ebff2e98a116c25667a8f8160cf8aeeaf0a437e6913e6870000082f49d00000000008b
//
// # Warning
//
// There is no security guarantees such as constant time implementation or side-channel attack resistance.
// This code is provided as-is. Partially audited, see https://github.com/Consensys/gnark/tree/master/audits
// for more details.
package fp
