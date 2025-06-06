// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package fr

// expBySqrtExp is equivalent to z.Exp(x, 32dbd584953b42564bf8fd939f24f531918901d9cc89c6c833a18bfa01)
//
// uses github.com/mmcloughlin/addchain v0.4.0 to generate a shorter addition chain
func (z *Element) expBySqrtExp(x Element) *Element {
	// addition chain:
	//
	//	_10      = 2*1
	//	_100     = 2*_10
	//	_101     = 1 + _100
	//	_110     = 1 + _101
	//	_1001    = _100 + _101
	//	_1011    = _10 + _1001
	//	_1111    = _100 + _1011
	//	_10011   = _100 + _1111
	//	_10101   = _10 + _10011
	//	_11001   = _100 + _10101
	//	_11011   = _10 + _11001
	//	_100001  = _110 + _11011
	//	_100111  = _110 + _100001
	//	_101011  = _100 + _100111
	//	_110001  = _110 + _101011
	//	_110011  = _10 + _110001
	//	_111001  = _110 + _110011
	//	_111011  = _10 + _111001
	//	_111101  = _10 + _111011
	//	_111111  = _10 + _111101
	//	_1100100 = _100111 + _111101
	//	_1111111 = _11011 + _1100100
	//	i40      = ((_1100100 << 4 + _11011) << 7 + _111101) << 5
	//	i58      = ((_1011 + i40) << 8 + _1001) << 7 + _10101
	//	i83      = ((i58 << 8 + _111011) << 7 + _100001) << 8
	//	i100     = ((_101011 + i83) << 6 + _1001) << 8 + _1111111
	//	i125     = ((i100 << 9 + _111111) << 6 + _11001) << 8
	//	i138     = ((_111001 + i125) << 4 + _1111) << 6 + _1001
	//	i162     = ((i138 << 8 + _111101) << 6 + _10011) << 8
	//	i177     = ((_11001 + i162) << 9 + _110001) << 3 + 1
	//	i204     = ((i177 << 13 + _111011) << 8 + _111001) << 4
	//	i224     = ((_1001 + i204) << 9 + _100111) << 8 + _11011
	//	i243     = ((i224 << 3 + 1) << 11 + _110011) << 3
	//	i264     = ((_101 + i243) << 10 + _110001) << 8 + _1111111
	//	return     (i264 << 2 + 1) << 9 + 1
	//
	// Operations: 225 squares 52 multiplies

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
		t17 = new(Element)
	)

	// var t0,t1,t2,t3,t4,t5,t6,t7,t8,t9,t10,t11,t12,t13,t14,t15,t16,t17 Element
	// Step 1: z = x^0x2
	z.Square(&x)

	// Step 2: t0 = x^0x4
	t0.Square(z)

	// Step 3: t1 = x^0x5
	t1.Mul(&x, t0)

	// Step 4: t6 = x^0x6
	t6.Mul(&x, t1)

	// Step 5: t5 = x^0x9
	t5.Mul(t0, t1)

	// Step 6: t16 = x^0xb
	t16.Mul(z, t5)

	// Step 7: t11 = x^0xf
	t11.Mul(t0, t16)

	// Step 8: t9 = x^0x13
	t9.Mul(t0, t11)

	// Step 9: t15 = x^0x15
	t15.Mul(z, t9)

	// Step 10: t8 = x^0x19
	t8.Mul(t0, t15)

	// Step 11: t3 = x^0x1b
	t3.Mul(z, t8)

	// Step 12: t14 = x^0x21
	t14.Mul(t6, t3)

	// Step 13: t4 = x^0x27
	t4.Mul(t6, t14)

	// Step 14: t13 = x^0x2b
	t13.Mul(t0, t4)

	// Step 15: t0 = x^0x31
	t0.Mul(t6, t13)

	// Step 16: t2 = x^0x33
	t2.Mul(z, t0)

	// Step 17: t6 = x^0x39
	t6.Mul(t6, t2)

	// Step 18: t7 = x^0x3b
	t7.Mul(z, t6)

	// Step 19: t10 = x^0x3d
	t10.Mul(z, t7)

	// Step 20: t12 = x^0x3f
	t12.Mul(z, t10)

	// Step 21: t17 = x^0x64
	t17.Mul(t4, t10)

	// Step 22: z = x^0x7f
	z.Mul(t3, t17)

	// Step 26: t17 = x^0x640
	for s := 0; s < 4; s++ {
		t17.Square(t17)
	}

	// Step 27: t17 = x^0x65b
	t17.Mul(t3, t17)

	// Step 34: t17 = x^0x32d80
	for s := 0; s < 7; s++ {
		t17.Square(t17)
	}

	// Step 35: t17 = x^0x32dbd
	t17.Mul(t10, t17)

	// Step 40: t17 = x^0x65b7a0
	for s := 0; s < 5; s++ {
		t17.Square(t17)
	}

	// Step 41: t16 = x^0x65b7ab
	t16.Mul(t16, t17)

	// Step 49: t16 = x^0x65b7ab00
	for s := 0; s < 8; s++ {
		t16.Square(t16)
	}

	// Step 50: t16 = x^0x65b7ab09
	t16.Mul(t5, t16)

	// Step 57: t16 = x^0x32dbd58480
	for s := 0; s < 7; s++ {
		t16.Square(t16)
	}

	// Step 58: t15 = x^0x32dbd58495
	t15.Mul(t15, t16)

	// Step 66: t15 = x^0x32dbd5849500
	for s := 0; s < 8; s++ {
		t15.Square(t15)
	}

	// Step 67: t15 = x^0x32dbd584953b
	t15.Mul(t7, t15)

	// Step 74: t15 = x^0x196deac24a9d80
	for s := 0; s < 7; s++ {
		t15.Square(t15)
	}

	// Step 75: t14 = x^0x196deac24a9da1
	t14.Mul(t14, t15)

	// Step 83: t14 = x^0x196deac24a9da100
	for s := 0; s < 8; s++ {
		t14.Square(t14)
	}

	// Step 84: t13 = x^0x196deac24a9da12b
	t13.Mul(t13, t14)

	// Step 90: t13 = x^0x65b7ab092a7684ac0
	for s := 0; s < 6; s++ {
		t13.Square(t13)
	}

	// Step 91: t13 = x^0x65b7ab092a7684ac9
	t13.Mul(t5, t13)

	// Step 99: t13 = x^0x65b7ab092a7684ac900
	for s := 0; s < 8; s++ {
		t13.Square(t13)
	}

	// Step 100: t13 = x^0x65b7ab092a7684ac97f
	t13.Mul(z, t13)

	// Step 109: t13 = x^0xcb6f561254ed09592fe00
	for s := 0; s < 9; s++ {
		t13.Square(t13)
	}

	// Step 110: t12 = x^0xcb6f561254ed09592fe3f
	t12.Mul(t12, t13)

	// Step 116: t12 = x^0x32dbd584953b42564bf8fc0
	for s := 0; s < 6; s++ {
		t12.Square(t12)
	}

	// Step 117: t12 = x^0x32dbd584953b42564bf8fd9
	t12.Mul(t8, t12)

	// Step 125: t12 = x^0x32dbd584953b42564bf8fd900
	for s := 0; s < 8; s++ {
		t12.Square(t12)
	}

	// Step 126: t12 = x^0x32dbd584953b42564bf8fd939
	t12.Mul(t6, t12)

	// Step 130: t12 = x^0x32dbd584953b42564bf8fd9390
	for s := 0; s < 4; s++ {
		t12.Square(t12)
	}

	// Step 131: t11 = x^0x32dbd584953b42564bf8fd939f
	t11.Mul(t11, t12)

	// Step 137: t11 = x^0xcb6f561254ed09592fe3f64e7c0
	for s := 0; s < 6; s++ {
		t11.Square(t11)
	}

	// Step 138: t11 = x^0xcb6f561254ed09592fe3f64e7c9
	t11.Mul(t5, t11)

	// Step 146: t11 = x^0xcb6f561254ed09592fe3f64e7c900
	for s := 0; s < 8; s++ {
		t11.Square(t11)
	}

	// Step 147: t10 = x^0xcb6f561254ed09592fe3f64e7c93d
	t10.Mul(t10, t11)

	// Step 153: t10 = x^0x32dbd584953b42564bf8fd939f24f40
	for s := 0; s < 6; s++ {
		t10.Square(t10)
	}

	// Step 154: t9 = x^0x32dbd584953b42564bf8fd939f24f53
	t9.Mul(t9, t10)

	// Step 162: t9 = x^0x32dbd584953b42564bf8fd939f24f5300
	for s := 0; s < 8; s++ {
		t9.Square(t9)
	}

	// Step 163: t8 = x^0x32dbd584953b42564bf8fd939f24f5319
	t8.Mul(t8, t9)

	// Step 172: t8 = x^0x65b7ab092a7684ac97f1fb273e49ea63200
	for s := 0; s < 9; s++ {
		t8.Square(t8)
	}

	// Step 173: t8 = x^0x65b7ab092a7684ac97f1fb273e49ea63231
	t8.Mul(t0, t8)

	// Step 176: t8 = x^0x32dbd584953b42564bf8fd939f24f5319188
	for s := 0; s < 3; s++ {
		t8.Square(t8)
	}

	// Step 177: t8 = x^0x32dbd584953b42564bf8fd939f24f5319189
	t8.Mul(&x, t8)

	// Step 190: t8 = x^0x65b7ab092a7684ac97f1fb273e49ea632312000
	for s := 0; s < 13; s++ {
		t8.Square(t8)
	}

	// Step 191: t7 = x^0x65b7ab092a7684ac97f1fb273e49ea63231203b
	t7.Mul(t7, t8)

	// Step 199: t7 = x^0x65b7ab092a7684ac97f1fb273e49ea63231203b00
	for s := 0; s < 8; s++ {
		t7.Square(t7)
	}

	// Step 200: t6 = x^0x65b7ab092a7684ac97f1fb273e49ea63231203b39
	t6.Mul(t6, t7)

	// Step 204: t6 = x^0x65b7ab092a7684ac97f1fb273e49ea63231203b390
	for s := 0; s < 4; s++ {
		t6.Square(t6)
	}

	// Step 205: t5 = x^0x65b7ab092a7684ac97f1fb273e49ea63231203b399
	t5.Mul(t5, t6)

	// Step 214: t5 = x^0xcb6f561254ed09592fe3f64e7c93d4c6462407673200
	for s := 0; s < 9; s++ {
		t5.Square(t5)
	}

	// Step 215: t4 = x^0xcb6f561254ed09592fe3f64e7c93d4c6462407673227
	t4.Mul(t4, t5)

	// Step 223: t4 = x^0xcb6f561254ed09592fe3f64e7c93d4c646240767322700
	for s := 0; s < 8; s++ {
		t4.Square(t4)
	}

	// Step 224: t3 = x^0xcb6f561254ed09592fe3f64e7c93d4c64624076732271b
	t3.Mul(t3, t4)

	// Step 227: t3 = x^0x65b7ab092a7684ac97f1fb273e49ea63231203b399138d8
	for s := 0; s < 3; s++ {
		t3.Square(t3)
	}

	// Step 228: t3 = x^0x65b7ab092a7684ac97f1fb273e49ea63231203b399138d9
	t3.Mul(&x, t3)

	// Step 239: t3 = x^0x32dbd584953b42564bf8fd939f24f531918901d9cc89c6c800
	for s := 0; s < 11; s++ {
		t3.Square(t3)
	}

	// Step 240: t2 = x^0x32dbd584953b42564bf8fd939f24f531918901d9cc89c6c833
	t2.Mul(t2, t3)

	// Step 243: t2 = x^0x196deac24a9da12b25fc7ec9cf927a98c8c480ece644e364198
	for s := 0; s < 3; s++ {
		t2.Square(t2)
	}

	// Step 244: t1 = x^0x196deac24a9da12b25fc7ec9cf927a98c8c480ece644e36419d
	t1.Mul(t1, t2)

	// Step 254: t1 = x^0x65b7ab092a7684ac97f1fb273e49ea63231203b399138d9067400
	for s := 0; s < 10; s++ {
		t1.Square(t1)
	}

	// Step 255: t0 = x^0x65b7ab092a7684ac97f1fb273e49ea63231203b399138d9067431
	t0.Mul(t0, t1)

	// Step 263: t0 = x^0x65b7ab092a7684ac97f1fb273e49ea63231203b399138d906743100
	for s := 0; s < 8; s++ {
		t0.Square(t0)
	}

	// Step 264: z = x^0x65b7ab092a7684ac97f1fb273e49ea63231203b399138d90674317f
	z.Mul(z, t0)

	// Step 266: z = x^0x196deac24a9da12b25fc7ec9cf927a98c8c480ece644e36419d0c5fc
	for s := 0; s < 2; s++ {
		z.Square(z)
	}

	// Step 267: z = x^0x196deac24a9da12b25fc7ec9cf927a98c8c480ece644e36419d0c5fd
	z.Mul(&x, z)

	// Step 276: z = x^0x32dbd584953b42564bf8fd939f24f531918901d9cc89c6c833a18bfa00
	for s := 0; s < 9; s++ {
		z.Square(z)
	}

	// Step 277: z = x^0x32dbd584953b42564bf8fd939f24f531918901d9cc89c6c833a18bfa01
	z.Mul(&x, z)

	return z
}

// expByLegendreExp is equivalent to z.Exp(x, cb6f561254ed09592fe3f64e7c93d4c64624076732271b20ce862fe80600000)
//
// uses github.com/mmcloughlin/addchain v0.4.0 to generate a shorter addition chain
func (z *Element) expByLegendreExp(x Element) *Element {
	// addition chain:
	//
	//	_10      = 2*1
	//	_11      = 1 + _10
	//	_100     = 1 + _11
	//	_101     = 1 + _100
	//	_110     = 1 + _101
	//	_1001    = _11 + _110
	//	_1011    = _10 + _1001
	//	_1111    = _100 + _1011
	//	_10011   = _100 + _1111
	//	_10101   = _10 + _10011
	//	_11001   = _100 + _10101
	//	_11011   = _10 + _11001
	//	_100001  = _110 + _11011
	//	_100111  = _110 + _100001
	//	_101011  = _100 + _100111
	//	_110001  = _110 + _101011
	//	_110011  = _10 + _110001
	//	_111001  = _110 + _110011
	//	_111011  = _10 + _111001
	//	_111101  = _10 + _111011
	//	_111111  = _10 + _111101
	//	_1100100 = _100111 + _111101
	//	_1111111 = _11011 + _1100100
	//	i41      = ((_1100100 << 4 + _11011) << 7 + _111101) << 5
	//	i59      = ((_1011 + i41) << 8 + _1001) << 7 + _10101
	//	i84      = ((i59 << 8 + _111011) << 7 + _100001) << 8
	//	i101     = ((_101011 + i84) << 6 + _1001) << 8 + _1111111
	//	i126     = ((i101 << 9 + _111111) << 6 + _11001) << 8
	//	i139     = ((_111001 + i126) << 4 + _1111) << 6 + _1001
	//	i163     = ((i139 << 8 + _111101) << 6 + _10011) << 8
	//	i178     = ((_11001 + i163) << 9 + _110001) << 3 + 1
	//	i205     = ((i178 << 13 + _111011) << 8 + _111001) << 4
	//	i225     = ((_1001 + i205) << 9 + _100111) << 8 + _11011
	//	i244     = ((i225 << 3 + 1) << 11 + _110011) << 3
	//	i265     = ((_101 + i244) << 10 + _110001) << 8 + _1111111
	//	return     ((i265 << 2 + 1) << 10 + _11) << 21
	//
	// Operations: 246 squares 54 multiplies

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
		t17 = new(Element)
		t18 = new(Element)
	)

	// var t0,t1,t2,t3,t4,t5,t6,t7,t8,t9,t10,t11,t12,t13,t14,t15,t16,t17,t18 Element
	// Step 1: t0 = x^0x2
	t0.Square(&x)

	// Step 2: z = x^0x3
	z.Mul(&x, t0)

	// Step 3: t1 = x^0x4
	t1.Mul(&x, z)

	// Step 4: t2 = x^0x5
	t2.Mul(&x, t1)

	// Step 5: t7 = x^0x6
	t7.Mul(&x, t2)

	// Step 6: t6 = x^0x9
	t6.Mul(z, t7)

	// Step 7: t17 = x^0xb
	t17.Mul(t0, t6)

	// Step 8: t12 = x^0xf
	t12.Mul(t1, t17)

	// Step 9: t10 = x^0x13
	t10.Mul(t1, t12)

	// Step 10: t16 = x^0x15
	t16.Mul(t0, t10)

	// Step 11: t9 = x^0x19
	t9.Mul(t1, t16)

	// Step 12: t4 = x^0x1b
	t4.Mul(t0, t9)

	// Step 13: t15 = x^0x21
	t15.Mul(t7, t4)

	// Step 14: t5 = x^0x27
	t5.Mul(t7, t15)

	// Step 15: t14 = x^0x2b
	t14.Mul(t1, t5)

	// Step 16: t1 = x^0x31
	t1.Mul(t7, t14)

	// Step 17: t3 = x^0x33
	t3.Mul(t0, t1)

	// Step 18: t7 = x^0x39
	t7.Mul(t7, t3)

	// Step 19: t8 = x^0x3b
	t8.Mul(t0, t7)

	// Step 20: t11 = x^0x3d
	t11.Mul(t0, t8)

	// Step 21: t13 = x^0x3f
	t13.Mul(t0, t11)

	// Step 22: t18 = x^0x64
	t18.Mul(t5, t11)

	// Step 23: t0 = x^0x7f
	t0.Mul(t4, t18)

	// Step 27: t18 = x^0x640
	for s := 0; s < 4; s++ {
		t18.Square(t18)
	}

	// Step 28: t18 = x^0x65b
	t18.Mul(t4, t18)

	// Step 35: t18 = x^0x32d80
	for s := 0; s < 7; s++ {
		t18.Square(t18)
	}

	// Step 36: t18 = x^0x32dbd
	t18.Mul(t11, t18)

	// Step 41: t18 = x^0x65b7a0
	for s := 0; s < 5; s++ {
		t18.Square(t18)
	}

	// Step 42: t17 = x^0x65b7ab
	t17.Mul(t17, t18)

	// Step 50: t17 = x^0x65b7ab00
	for s := 0; s < 8; s++ {
		t17.Square(t17)
	}

	// Step 51: t17 = x^0x65b7ab09
	t17.Mul(t6, t17)

	// Step 58: t17 = x^0x32dbd58480
	for s := 0; s < 7; s++ {
		t17.Square(t17)
	}

	// Step 59: t16 = x^0x32dbd58495
	t16.Mul(t16, t17)

	// Step 67: t16 = x^0x32dbd5849500
	for s := 0; s < 8; s++ {
		t16.Square(t16)
	}

	// Step 68: t16 = x^0x32dbd584953b
	t16.Mul(t8, t16)

	// Step 75: t16 = x^0x196deac24a9d80
	for s := 0; s < 7; s++ {
		t16.Square(t16)
	}

	// Step 76: t15 = x^0x196deac24a9da1
	t15.Mul(t15, t16)

	// Step 84: t15 = x^0x196deac24a9da100
	for s := 0; s < 8; s++ {
		t15.Square(t15)
	}

	// Step 85: t14 = x^0x196deac24a9da12b
	t14.Mul(t14, t15)

	// Step 91: t14 = x^0x65b7ab092a7684ac0
	for s := 0; s < 6; s++ {
		t14.Square(t14)
	}

	// Step 92: t14 = x^0x65b7ab092a7684ac9
	t14.Mul(t6, t14)

	// Step 100: t14 = x^0x65b7ab092a7684ac900
	for s := 0; s < 8; s++ {
		t14.Square(t14)
	}

	// Step 101: t14 = x^0x65b7ab092a7684ac97f
	t14.Mul(t0, t14)

	// Step 110: t14 = x^0xcb6f561254ed09592fe00
	for s := 0; s < 9; s++ {
		t14.Square(t14)
	}

	// Step 111: t13 = x^0xcb6f561254ed09592fe3f
	t13.Mul(t13, t14)

	// Step 117: t13 = x^0x32dbd584953b42564bf8fc0
	for s := 0; s < 6; s++ {
		t13.Square(t13)
	}

	// Step 118: t13 = x^0x32dbd584953b42564bf8fd9
	t13.Mul(t9, t13)

	// Step 126: t13 = x^0x32dbd584953b42564bf8fd900
	for s := 0; s < 8; s++ {
		t13.Square(t13)
	}

	// Step 127: t13 = x^0x32dbd584953b42564bf8fd939
	t13.Mul(t7, t13)

	// Step 131: t13 = x^0x32dbd584953b42564bf8fd9390
	for s := 0; s < 4; s++ {
		t13.Square(t13)
	}

	// Step 132: t12 = x^0x32dbd584953b42564bf8fd939f
	t12.Mul(t12, t13)

	// Step 138: t12 = x^0xcb6f561254ed09592fe3f64e7c0
	for s := 0; s < 6; s++ {
		t12.Square(t12)
	}

	// Step 139: t12 = x^0xcb6f561254ed09592fe3f64e7c9
	t12.Mul(t6, t12)

	// Step 147: t12 = x^0xcb6f561254ed09592fe3f64e7c900
	for s := 0; s < 8; s++ {
		t12.Square(t12)
	}

	// Step 148: t11 = x^0xcb6f561254ed09592fe3f64e7c93d
	t11.Mul(t11, t12)

	// Step 154: t11 = x^0x32dbd584953b42564bf8fd939f24f40
	for s := 0; s < 6; s++ {
		t11.Square(t11)
	}

	// Step 155: t10 = x^0x32dbd584953b42564bf8fd939f24f53
	t10.Mul(t10, t11)

	// Step 163: t10 = x^0x32dbd584953b42564bf8fd939f24f5300
	for s := 0; s < 8; s++ {
		t10.Square(t10)
	}

	// Step 164: t9 = x^0x32dbd584953b42564bf8fd939f24f5319
	t9.Mul(t9, t10)

	// Step 173: t9 = x^0x65b7ab092a7684ac97f1fb273e49ea63200
	for s := 0; s < 9; s++ {
		t9.Square(t9)
	}

	// Step 174: t9 = x^0x65b7ab092a7684ac97f1fb273e49ea63231
	t9.Mul(t1, t9)

	// Step 177: t9 = x^0x32dbd584953b42564bf8fd939f24f5319188
	for s := 0; s < 3; s++ {
		t9.Square(t9)
	}

	// Step 178: t9 = x^0x32dbd584953b42564bf8fd939f24f5319189
	t9.Mul(&x, t9)

	// Step 191: t9 = x^0x65b7ab092a7684ac97f1fb273e49ea632312000
	for s := 0; s < 13; s++ {
		t9.Square(t9)
	}

	// Step 192: t8 = x^0x65b7ab092a7684ac97f1fb273e49ea63231203b
	t8.Mul(t8, t9)

	// Step 200: t8 = x^0x65b7ab092a7684ac97f1fb273e49ea63231203b00
	for s := 0; s < 8; s++ {
		t8.Square(t8)
	}

	// Step 201: t7 = x^0x65b7ab092a7684ac97f1fb273e49ea63231203b39
	t7.Mul(t7, t8)

	// Step 205: t7 = x^0x65b7ab092a7684ac97f1fb273e49ea63231203b390
	for s := 0; s < 4; s++ {
		t7.Square(t7)
	}

	// Step 206: t6 = x^0x65b7ab092a7684ac97f1fb273e49ea63231203b399
	t6.Mul(t6, t7)

	// Step 215: t6 = x^0xcb6f561254ed09592fe3f64e7c93d4c6462407673200
	for s := 0; s < 9; s++ {
		t6.Square(t6)
	}

	// Step 216: t5 = x^0xcb6f561254ed09592fe3f64e7c93d4c6462407673227
	t5.Mul(t5, t6)

	// Step 224: t5 = x^0xcb6f561254ed09592fe3f64e7c93d4c646240767322700
	for s := 0; s < 8; s++ {
		t5.Square(t5)
	}

	// Step 225: t4 = x^0xcb6f561254ed09592fe3f64e7c93d4c64624076732271b
	t4.Mul(t4, t5)

	// Step 228: t4 = x^0x65b7ab092a7684ac97f1fb273e49ea63231203b399138d8
	for s := 0; s < 3; s++ {
		t4.Square(t4)
	}

	// Step 229: t4 = x^0x65b7ab092a7684ac97f1fb273e49ea63231203b399138d9
	t4.Mul(&x, t4)

	// Step 240: t4 = x^0x32dbd584953b42564bf8fd939f24f531918901d9cc89c6c800
	for s := 0; s < 11; s++ {
		t4.Square(t4)
	}

	// Step 241: t3 = x^0x32dbd584953b42564bf8fd939f24f531918901d9cc89c6c833
	t3.Mul(t3, t4)

	// Step 244: t3 = x^0x196deac24a9da12b25fc7ec9cf927a98c8c480ece644e364198
	for s := 0; s < 3; s++ {
		t3.Square(t3)
	}

	// Step 245: t2 = x^0x196deac24a9da12b25fc7ec9cf927a98c8c480ece644e36419d
	t2.Mul(t2, t3)

	// Step 255: t2 = x^0x65b7ab092a7684ac97f1fb273e49ea63231203b399138d9067400
	for s := 0; s < 10; s++ {
		t2.Square(t2)
	}

	// Step 256: t1 = x^0x65b7ab092a7684ac97f1fb273e49ea63231203b399138d9067431
	t1.Mul(t1, t2)

	// Step 264: t1 = x^0x65b7ab092a7684ac97f1fb273e49ea63231203b399138d906743100
	for s := 0; s < 8; s++ {
		t1.Square(t1)
	}

	// Step 265: t0 = x^0x65b7ab092a7684ac97f1fb273e49ea63231203b399138d90674317f
	t0.Mul(t0, t1)

	// Step 267: t0 = x^0x196deac24a9da12b25fc7ec9cf927a98c8c480ece644e36419d0c5fc
	for s := 0; s < 2; s++ {
		t0.Square(t0)
	}

	// Step 268: t0 = x^0x196deac24a9da12b25fc7ec9cf927a98c8c480ece644e36419d0c5fd
	t0.Mul(&x, t0)

	// Step 278: t0 = x^0x65b7ab092a7684ac97f1fb273e49ea63231203b399138d90674317f400
	for s := 0; s < 10; s++ {
		t0.Square(t0)
	}

	// Step 279: z = x^0x65b7ab092a7684ac97f1fb273e49ea63231203b399138d90674317f403
	z.Mul(z, t0)

	// Step 300: z = x^0xcb6f561254ed09592fe3f64e7c93d4c64624076732271b20ce862fe80600000
	for s := 0; s < 21; s++ {
		z.Square(z)
	}

	return z
}
