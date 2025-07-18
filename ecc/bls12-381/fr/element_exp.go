// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package fr

// expBySqrtExp is equivalent to z.Exp(x, 39f6d3a994cebea4199cec0404d0ec02a9ded2017fff2dff7fffffff)
//
// uses github.com/mmcloughlin/addchain v0.4.0 to generate a shorter addition chain
func (z *Element) expBySqrtExp(x Element) *Element {
	// addition chain:
	//
	//	_10       = 2*1
	//	_100      = 2*_10
	//	_110      = _10 + _100
	//	_1100     = 2*_110
	//	_10010    = _110 + _1100
	//	_10011    = 1 + _10010
	//	_10110    = _100 + _10010
	//	_11000    = _10 + _10110
	//	_11010    = _10 + _11000
	//	_100010   = _1100 + _10110
	//	_110101   = _10011 + _100010
	//	_111011   = _110 + _110101
	//	_1001011  = _10110 + _110101
	//	_1001101  = _10 + _1001011
	//	_1010101  = _11010 + _111011
	//	_1100111  = _10010 + _1010101
	//	_1101001  = _10 + _1100111
	//	_10000011 = _11010 + _1101001
	//	_10011001 = _10110 + _10000011
	//	_10011101 = _100 + _10011001
	//	_10111111 = _100010 + _10011101
	//	_11010111 = _11000 + _10111111
	//	_11011011 = _100 + _11010111
	//	_11100111 = _1100 + _11011011
	//	_11101111 = _11000 + _11010111
	//	_11111111 = _11000 + _11100111
	//	i54       = ((_11100111 << 8 + _11011011) << 9 + _10011101) << 9
	//	i74       = ((_10011001 + i54) << 9 + _10011001) << 8 + _11010111
	//	i101      = ((i74 << 6 + _110101) << 10 + _10000011) << 9
	//	i120      = ((_1100111 + i101) << 8 + _111011) << 8 + 1
	//	i161      = ((i120 << 14 + _1001101) << 10 + _111011) << 15
	//	i182      = ((_1010101 + i161) << 10 + _11101111) << 8 + _1101001
	//	i215      = ((i182 << 16 + _10111111) << 8 + _11111111) << 7
	//	i235      = ((_1001011 + i215) << 9 + _11111111) << 8 + _10111111
	//	i261      = ((i235 << 8 + _11111111) << 8 + _11111111) << 8
	//	return      2*(_11111111 + i261) + 1
	//
	// Operations: 217 squares 47 multiplies

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
	)

	// var t0,t1,t2,t3,t4,t5,t6,t7,t8,t9,t10,t11,t12,t13,t14 Element
	// Step 1: t2 = x^0x2
	t2.Square(&x)

	// Step 2: t13 = x^0x4
	t13.Square(t2)

	// Step 3: t1 = x^0x6
	t1.Mul(t2, t13)

	// Step 4: t3 = x^0xc
	t3.Square(t1)

	// Step 5: t7 = x^0x12
	t7.Mul(t1, t3)

	// Step 6: t4 = x^0x13
	t4.Mul(&x, t7)

	// Step 7: t10 = x^0x16
	t10.Mul(t13, t7)

	// Step 8: z = x^0x18
	z.Mul(t2, t10)

	// Step 9: t8 = x^0x1a
	t8.Mul(t2, z)

	// Step 10: t0 = x^0x22
	t0.Mul(t3, t10)

	// Step 11: t9 = x^0x35
	t9.Mul(t4, t0)

	// Step 12: t5 = x^0x3b
	t5.Mul(t1, t9)

	// Step 13: t1 = x^0x4b
	t1.Mul(t10, t9)

	// Step 14: t6 = x^0x4d
	t6.Mul(t2, t1)

	// Step 15: t4 = x^0x55
	t4.Mul(t8, t5)

	// Step 16: t7 = x^0x67
	t7.Mul(t7, t4)

	// Step 17: t2 = x^0x69
	t2.Mul(t2, t7)

	// Step 18: t8 = x^0x83
	t8.Mul(t8, t2)

	// Step 19: t11 = x^0x99
	t11.Mul(t10, t8)

	// Step 20: t12 = x^0x9d
	t12.Mul(t13, t11)

	// Step 21: t0 = x^0xbf
	t0.Mul(t0, t12)

	// Step 22: t10 = x^0xd7
	t10.Mul(z, t0)

	// Step 23: t13 = x^0xdb
	t13.Mul(t13, t10)

	// Step 24: t14 = x^0xe7
	t14.Mul(t3, t13)

	// Step 25: t3 = x^0xef
	t3.Mul(z, t10)

	// Step 26: z = x^0xff
	z.Mul(z, t14)

	// Step 34: t14 = x^0xe700
	for s := 0; s < 8; s++ {
		t14.Square(t14)
	}

	// Step 35: t13 = x^0xe7db
	t13.Mul(t13, t14)

	// Step 44: t13 = x^0x1cfb600
	for s := 0; s < 9; s++ {
		t13.Square(t13)
	}

	// Step 45: t12 = x^0x1cfb69d
	t12.Mul(t12, t13)

	// Step 54: t12 = x^0x39f6d3a00
	for s := 0; s < 9; s++ {
		t12.Square(t12)
	}

	// Step 55: t12 = x^0x39f6d3a99
	t12.Mul(t11, t12)

	// Step 64: t12 = x^0x73eda753200
	for s := 0; s < 9; s++ {
		t12.Square(t12)
	}

	// Step 65: t11 = x^0x73eda753299
	t11.Mul(t11, t12)

	// Step 73: t11 = x^0x73eda75329900
	for s := 0; s < 8; s++ {
		t11.Square(t11)
	}

	// Step 74: t10 = x^0x73eda753299d7
	t10.Mul(t10, t11)

	// Step 80: t10 = x^0x1cfb69d4ca675c0
	for s := 0; s < 6; s++ {
		t10.Square(t10)
	}

	// Step 81: t9 = x^0x1cfb69d4ca675f5
	t9.Mul(t9, t10)

	// Step 91: t9 = x^0x73eda753299d7d400
	for s := 0; s < 10; s++ {
		t9.Square(t9)
	}

	// Step 92: t8 = x^0x73eda753299d7d483
	t8.Mul(t8, t9)

	// Step 101: t8 = x^0xe7db4ea6533afa90600
	for s := 0; s < 9; s++ {
		t8.Square(t8)
	}

	// Step 102: t7 = x^0xe7db4ea6533afa90667
	t7.Mul(t7, t8)

	// Step 110: t7 = x^0xe7db4ea6533afa9066700
	for s := 0; s < 8; s++ {
		t7.Square(t7)
	}

	// Step 111: t7 = x^0xe7db4ea6533afa906673b
	t7.Mul(t5, t7)

	// Step 119: t7 = x^0xe7db4ea6533afa906673b00
	for s := 0; s < 8; s++ {
		t7.Square(t7)
	}

	// Step 120: t7 = x^0xe7db4ea6533afa906673b01
	t7.Mul(&x, t7)

	// Step 134: t7 = x^0x39f6d3a994cebea4199cec04000
	for s := 0; s < 14; s++ {
		t7.Square(t7)
	}

	// Step 135: t6 = x^0x39f6d3a994cebea4199cec0404d
	t6.Mul(t6, t7)

	// Step 145: t6 = x^0xe7db4ea6533afa906673b01013400
	for s := 0; s < 10; s++ {
		t6.Square(t6)
	}

	// Step 146: t5 = x^0xe7db4ea6533afa906673b0101343b
	t5.Mul(t5, t6)

	// Step 161: t5 = x^0x73eda753299d7d483339d80809a1d8000
	for s := 0; s < 15; s++ {
		t5.Square(t5)
	}

	// Step 162: t4 = x^0x73eda753299d7d483339d80809a1d8055
	t4.Mul(t4, t5)

	// Step 172: t4 = x^0x1cfb69d4ca675f520cce7602026876015400
	for s := 0; s < 10; s++ {
		t4.Square(t4)
	}

	// Step 173: t3 = x^0x1cfb69d4ca675f520cce76020268760154ef
	t3.Mul(t3, t4)

	// Step 181: t3 = x^0x1cfb69d4ca675f520cce76020268760154ef00
	for s := 0; s < 8; s++ {
		t3.Square(t3)
	}

	// Step 182: t2 = x^0x1cfb69d4ca675f520cce76020268760154ef69
	t2.Mul(t2, t3)

	// Step 198: t2 = x^0x1cfb69d4ca675f520cce76020268760154ef690000
	for s := 0; s < 16; s++ {
		t2.Square(t2)
	}

	// Step 199: t2 = x^0x1cfb69d4ca675f520cce76020268760154ef6900bf
	t2.Mul(t0, t2)

	// Step 207: t2 = x^0x1cfb69d4ca675f520cce76020268760154ef6900bf00
	for s := 0; s < 8; s++ {
		t2.Square(t2)
	}

	// Step 208: t2 = x^0x1cfb69d4ca675f520cce76020268760154ef6900bfff
	t2.Mul(z, t2)

	// Step 215: t2 = x^0xe7db4ea6533afa906673b0101343b00aa77b4805fff80
	for s := 0; s < 7; s++ {
		t2.Square(t2)
	}

	// Step 216: t1 = x^0xe7db4ea6533afa906673b0101343b00aa77b4805fffcb
	t1.Mul(t1, t2)

	// Step 225: t1 = x^0x1cfb69d4ca675f520cce76020268760154ef6900bfff9600
	for s := 0; s < 9; s++ {
		t1.Square(t1)
	}

	// Step 226: t1 = x^0x1cfb69d4ca675f520cce76020268760154ef6900bfff96ff
	t1.Mul(z, t1)

	// Step 234: t1 = x^0x1cfb69d4ca675f520cce76020268760154ef6900bfff96ff00
	for s := 0; s < 8; s++ {
		t1.Square(t1)
	}

	// Step 235: t0 = x^0x1cfb69d4ca675f520cce76020268760154ef6900bfff96ffbf
	t0.Mul(t0, t1)

	// Step 243: t0 = x^0x1cfb69d4ca675f520cce76020268760154ef6900bfff96ffbf00
	for s := 0; s < 8; s++ {
		t0.Square(t0)
	}

	// Step 244: t0 = x^0x1cfb69d4ca675f520cce76020268760154ef6900bfff96ffbfff
	t0.Mul(z, t0)

	// Step 252: t0 = x^0x1cfb69d4ca675f520cce76020268760154ef6900bfff96ffbfff00
	for s := 0; s < 8; s++ {
		t0.Square(t0)
	}

	// Step 253: t0 = x^0x1cfb69d4ca675f520cce76020268760154ef6900bfff96ffbfffff
	t0.Mul(z, t0)

	// Step 261: t0 = x^0x1cfb69d4ca675f520cce76020268760154ef6900bfff96ffbfffff00
	for s := 0; s < 8; s++ {
		t0.Square(t0)
	}

	// Step 262: z = x^0x1cfb69d4ca675f520cce76020268760154ef6900bfff96ffbfffffff
	z.Mul(z, t0)

	// Step 263: z = x^0x39f6d3a994cebea4199cec0404d0ec02a9ded2017fff2dff7ffffffe
	z.Square(z)

	// Step 264: z = x^0x39f6d3a994cebea4199cec0404d0ec02a9ded2017fff2dff7fffffff
	z.Mul(&x, z)

	return z
}
