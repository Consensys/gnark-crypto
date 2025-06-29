// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package fr

// expBySqrtExp is equivalent to z.Exp(x, 12ab655e9a2ca55660b44d1e5c37b00159aa76fed00000010a11)
//
// uses github.com/mmcloughlin/addchain v0.4.0 to generate a shorter addition chain
func (z *Element) expBySqrtExp(x Element) *Element {
	// addition chain:
	//
	//	_10       = 2*1
	//	_11       = 1 + _10
	//	_101      = _10 + _11
	//	_1000     = _11 + _101
	//	_1011     = _11 + _1000
	//	_10000    = _101 + _1011
	//	_10001    = 1 + _10000
	//	_10110    = _101 + _10001
	//	_100000   = 2*_10000
	//	_101011   = _1011 + _100000
	//	_101101   = _10 + _101011
	//	_1011010  = 2*_101101
	//	_1011011  = 1 + _1011010
	//	_1111011  = _100000 + _1011011
	//	_10000101 = _101011 + _1011010
	//	_10001011 = _10000 + _1111011
	//	_10100101 = _100000 + _10000101
	//	_10101011 = _100000 + _10001011
	//	_11000001 = _10110 + _10101011
	//	_11000011 = _10 + _11000001
	//	_11010001 = _10000 + _11000001
	//	_11010011 = _10 + _11010001
	//	_11010101 = _10 + _11010011
	//	_11100101 = _10000 + _11010101
	//	_11101101 = _1000 + _11100101
	//	i45       = ((_10000101 + _10100101) << 7 + _1011011) << 10 + _10101011
	//	i74       = ((i45 << 8 + _11010011) << 9 + _10001011) << 10
	//	i94       = ((_10100101 + i74) << 7 + _101011) << 10 + _11000001
	//	i123      = ((i94 << 9 + _11010001) << 10 + _11010001) << 8
	//	i142      = ((_11100101 + i123) << 8 + _11000011) << 8 + _1111011
	//	i181      = ((i142 << 17 + _101011) << 10 + _11010101) << 10
	//	i195      = ((_11101101 + i181) << 8 + _11101101 + _10000) << 3
	//	return      ((_101 + i195) << 35 + _10000101) << 9 + _10001
	//
	// Operations: 199 squares 43 multiplies

	// Allocate Temporaries.
	var (
		t0  = new(Element)
		t1  = new(Element)
		t2  = new(Element)
		t3  = new(Element)
		t4  = new(Element)
		t5  = new(Element)
		t6  = new(Element)
		t7  = new(Element)
		t8  = new(Element)
		t9  = new(Element)
		t10 = new(Element)
		t11 = new(Element)
		t12 = new(Element)
		t13 = new(Element)
		t14 = new(Element)
		t15 = new(Element)
		t16 = new(Element)
	)

	// var t0,t1,t2,t3,t4,t5,t6,t7,t8,t9,t10,t11,t12,t13,t14,t15,t16 Element
	// Step 1: t4 = x^0x2
	t4.Square(&x)

	// Step 2: z = x^0x3
	z.Mul(&x, t4)

	// Step 3: t1 = x^0x5
	t1.Mul(t4, z)

	// Step 4: t3 = x^0x8
	t3.Mul(z, t1)

	// Step 5: t0 = x^0xb
	t0.Mul(z, t3)

	// Step 6: t2 = x^0x10
	t2.Mul(t1, t0)

	// Step 7: z = x^0x11
	z.Mul(&x, t2)

	// Step 8: t7 = x^0x16
	t7.Mul(t1, z)

	// Step 9: t8 = x^0x20
	t8.Square(t2)

	// Step 10: t5 = x^0x2b
	t5.Mul(t0, t8)

	// Step 11: t0 = x^0x2d
	t0.Mul(t4, t5)

	// Step 12: t0 = x^0x5a
	t0.Square(t0)

	// Step 13: t15 = x^0x5b
	t15.Mul(&x, t0)

	// Step 14: t6 = x^0x7b
	t6.Mul(t8, t15)

	// Step 15: t0 = x^0x85
	t0.Mul(t5, t0)

	// Step 16: t12 = x^0x8b
	t12.Mul(t2, t6)

	// Step 17: t11 = x^0xa5
	t11.Mul(t8, t0)

	// Step 18: t14 = x^0xab
	t14.Mul(t8, t12)

	// Step 19: t10 = x^0xc1
	t10.Mul(t7, t14)

	// Step 20: t7 = x^0xc3
	t7.Mul(t4, t10)

	// Step 21: t9 = x^0xd1
	t9.Mul(t2, t10)

	// Step 22: t13 = x^0xd3
	t13.Mul(t4, t9)

	// Step 23: t4 = x^0xd5
	t4.Mul(t4, t13)

	// Step 24: t8 = x^0xe5
	t8.Mul(t2, t4)

	// Step 25: t3 = x^0xed
	t3.Mul(t3, t8)

	// Step 26: t16 = x^0x12a
	t16.Mul(t0, t11)

	// Step 33: t16 = x^0x9500
	for s := 0; s < 7; s++ {
		t16.Square(t16)
	}

	// Step 34: t15 = x^0x955b
	t15.Mul(t15, t16)

	// Step 44: t15 = x^0x2556c00
	for s := 0; s < 10; s++ {
		t15.Square(t15)
	}

	// Step 45: t14 = x^0x2556cab
	t14.Mul(t14, t15)

	// Step 53: t14 = x^0x2556cab00
	for s := 0; s < 8; s++ {
		t14.Square(t14)
	}

	// Step 54: t13 = x^0x2556cabd3
	t13.Mul(t13, t14)

	// Step 63: t13 = x^0x4aad957a600
	for s := 0; s < 9; s++ {
		t13.Square(t13)
	}

	// Step 64: t12 = x^0x4aad957a68b
	t12.Mul(t12, t13)

	// Step 74: t12 = x^0x12ab655e9a2c00
	for s := 0; s < 10; s++ {
		t12.Square(t12)
	}

	// Step 75: t11 = x^0x12ab655e9a2ca5
	t11.Mul(t11, t12)

	// Step 82: t11 = x^0x955b2af4d165280
	for s := 0; s < 7; s++ {
		t11.Square(t11)
	}

	// Step 83: t11 = x^0x955b2af4d1652ab
	t11.Mul(t5, t11)

	// Step 93: t11 = x^0x2556cabd34594aac00
	for s := 0; s < 10; s++ {
		t11.Square(t11)
	}

	// Step 94: t10 = x^0x2556cabd34594aacc1
	t10.Mul(t10, t11)

	// Step 103: t10 = x^0x4aad957a68b295598200
	for s := 0; s < 9; s++ {
		t10.Square(t10)
	}

	// Step 104: t10 = x^0x4aad957a68b2955982d1
	t10.Mul(t9, t10)

	// Step 114: t10 = x^0x12ab655e9a2ca55660b4400
	for s := 0; s < 10; s++ {
		t10.Square(t10)
	}

	// Step 115: t9 = x^0x12ab655e9a2ca55660b44d1
	t9.Mul(t9, t10)

	// Step 123: t9 = x^0x12ab655e9a2ca55660b44d100
	for s := 0; s < 8; s++ {
		t9.Square(t9)
	}

	// Step 124: t8 = x^0x12ab655e9a2ca55660b44d1e5
	t8.Mul(t8, t9)

	// Step 132: t8 = x^0x12ab655e9a2ca55660b44d1e500
	for s := 0; s < 8; s++ {
		t8.Square(t8)
	}

	// Step 133: t7 = x^0x12ab655e9a2ca55660b44d1e5c3
	t7.Mul(t7, t8)

	// Step 141: t7 = x^0x12ab655e9a2ca55660b44d1e5c300
	for s := 0; s < 8; s++ {
		t7.Square(t7)
	}

	// Step 142: t6 = x^0x12ab655e9a2ca55660b44d1e5c37b
	t6.Mul(t6, t7)

	// Step 159: t6 = x^0x2556cabd34594aacc1689a3cb86f60000
	for s := 0; s < 17; s++ {
		t6.Square(t6)
	}

	// Step 160: t5 = x^0x2556cabd34594aacc1689a3cb86f6002b
	t5.Mul(t5, t6)

	// Step 170: t5 = x^0x955b2af4d1652ab305a268f2e1bd800ac00
	for s := 0; s < 10; s++ {
		t5.Square(t5)
	}

	// Step 171: t4 = x^0x955b2af4d1652ab305a268f2e1bd800acd5
	t4.Mul(t4, t5)

	// Step 181: t4 = x^0x2556cabd34594aacc1689a3cb86f6002b35400
	for s := 0; s < 10; s++ {
		t4.Square(t4)
	}

	// Step 182: t4 = x^0x2556cabd34594aacc1689a3cb86f6002b354ed
	t4.Mul(t3, t4)

	// Step 190: t4 = x^0x2556cabd34594aacc1689a3cb86f6002b354ed00
	for s := 0; s < 8; s++ {
		t4.Square(t4)
	}

	// Step 191: t3 = x^0x2556cabd34594aacc1689a3cb86f6002b354eded
	t3.Mul(t3, t4)

	// Step 192: t2 = x^0x2556cabd34594aacc1689a3cb86f6002b354edfd
	t2.Mul(t2, t3)

	// Step 195: t2 = x^0x12ab655e9a2ca55660b44d1e5c37b00159aa76fe8
	for s := 0; s < 3; s++ {
		t2.Square(t2)
	}

	// Step 196: t1 = x^0x12ab655e9a2ca55660b44d1e5c37b00159aa76fed
	t1.Mul(t1, t2)

	// Step 231: t1 = x^0x955b2af4d1652ab305a268f2e1bd800acd53b7f6800000000
	for s := 0; s < 35; s++ {
		t1.Square(t1)
	}

	// Step 232: t0 = x^0x955b2af4d1652ab305a268f2e1bd800acd53b7f6800000085
	t0.Mul(t0, t1)

	// Step 241: t0 = x^0x12ab655e9a2ca55660b44d1e5c37b00159aa76fed00000010a00
	for s := 0; s < 9; s++ {
		t0.Square(t0)
	}

	// Step 242: z = x^0x12ab655e9a2ca55660b44d1e5c37b00159aa76fed00000010a11
	z.Mul(z, t0)

	return z
}
