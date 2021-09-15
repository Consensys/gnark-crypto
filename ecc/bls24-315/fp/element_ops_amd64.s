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

#include "textflag.h"
#include "funcdata.h"

// modulus q
DATA q<>+0(SB)/8, $0x6fe802ff40300001
DATA q<>+8(SB)/8, $0x421ee5da52bde502
DATA q<>+16(SB)/8, $0xdec1d01aa27a1ae0
DATA q<>+24(SB)/8, $0xd3f7498be97c5eaf
DATA q<>+32(SB)/8, $0x04c23a02b586d650
GLOBL q<>(SB), (RODATA+NOPTR), $40

// qInv0 q'[0]
DATA qInv0<>(SB)/8, $0x702ff9ff402fffff
GLOBL qInv0<>(SB), (RODATA+NOPTR), $8

#define REDUCE(ra0, ra1, ra2, ra3, ra4, rb0, rb1, rb2, rb3, rb4) \
	MOVQ    ra0, rb0;        \
	SUBQ    q<>(SB), ra0;    \
	MOVQ    ra1, rb1;        \
	SBBQ    q<>+8(SB), ra1;  \
	MOVQ    ra2, rb2;        \
	SBBQ    q<>+16(SB), ra2; \
	MOVQ    ra3, rb3;        \
	SBBQ    q<>+24(SB), ra3; \
	MOVQ    ra4, rb4;        \
	SBBQ    q<>+32(SB), ra4; \
	CMOVQCS rb0, ra0;        \
	CMOVQCS rb1, ra1;        \
	CMOVQCS rb2, ra2;        \
	CMOVQCS rb3, ra3;        \
	CMOVQCS rb4, ra4;        \

// add(res, x, y *Element)
TEXT ·add(SB), NOSPLIT, $0-24
	MOVQ x+8(FP), AX
	MOVQ 0(AX), CX
	MOVQ 8(AX), BX
	MOVQ 16(AX), SI
	MOVQ 24(AX), DI
	MOVQ 32(AX), R8
	MOVQ y+16(FP), DX
	ADDQ 0(DX), CX
	ADCQ 8(DX), BX
	ADCQ 16(DX), SI
	ADCQ 24(DX), DI
	ADCQ 32(DX), R8

	// reduce element(CX,BX,SI,DI,R8) using temp registers (R9,R10,R11,R12,R13)
	REDUCE(CX,BX,SI,DI,R8,R9,R10,R11,R12,R13)

	MOVQ res+0(FP), R14
	MOVQ CX, 0(R14)
	MOVQ BX, 8(R14)
	MOVQ SI, 16(R14)
	MOVQ DI, 24(R14)
	MOVQ R8, 32(R14)
	RET

// sub(res, x, y *Element)
TEXT ·sub(SB), NOSPLIT, $0-24
	XORQ    R8, R8
	MOVQ    x+8(FP), DI
	MOVQ    0(DI), AX
	MOVQ    8(DI), DX
	MOVQ    16(DI), CX
	MOVQ    24(DI), BX
	MOVQ    32(DI), SI
	MOVQ    y+16(FP), DI
	SUBQ    0(DI), AX
	SBBQ    8(DI), DX
	SBBQ    16(DI), CX
	SBBQ    24(DI), BX
	SBBQ    32(DI), SI
	MOVQ    $0x6fe802ff40300001, R9
	MOVQ    $0x421ee5da52bde502, R10
	MOVQ    $0xdec1d01aa27a1ae0, R11
	MOVQ    $0xd3f7498be97c5eaf, R12
	MOVQ    $0x04c23a02b586d650, R13
	CMOVQCC R8, R9
	CMOVQCC R8, R10
	CMOVQCC R8, R11
	CMOVQCC R8, R12
	CMOVQCC R8, R13
	ADDQ    R9, AX
	ADCQ    R10, DX
	ADCQ    R11, CX
	ADCQ    R12, BX
	ADCQ    R13, SI
	MOVQ    res+0(FP), R14
	MOVQ    AX, 0(R14)
	MOVQ    DX, 8(R14)
	MOVQ    CX, 16(R14)
	MOVQ    BX, 24(R14)
	MOVQ    SI, 32(R14)
	RET

// double(res, x *Element)
TEXT ·double(SB), NOSPLIT, $0-16
	MOVQ x+8(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), SI
	MOVQ 32(AX), DI
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI
	ADCQ DI, DI

	// reduce element(DX,CX,BX,SI,DI) using temp registers (R8,R9,R10,R11,R12)
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12)

	MOVQ res+0(FP), R13
	MOVQ DX, 0(R13)
	MOVQ CX, 8(R13)
	MOVQ BX, 16(R13)
	MOVQ SI, 24(R13)
	MOVQ DI, 32(R13)
	RET

// neg(res, x *Element)
TEXT ·neg(SB), NOSPLIT, $0-16
	MOVQ  res+0(FP), R8
	MOVQ  x+8(FP), AX
	MOVQ  0(AX), DX
	MOVQ  8(AX), CX
	MOVQ  16(AX), BX
	MOVQ  24(AX), SI
	MOVQ  32(AX), DI
	MOVQ  DX, AX
	ORQ   CX, AX
	ORQ   BX, AX
	ORQ   SI, AX
	ORQ   DI, AX
	TESTQ AX, AX
	JEQ   l1
	MOVQ  $0x6fe802ff40300001, R9
	SUBQ  DX, R9
	MOVQ  R9, 0(R8)
	MOVQ  $0x421ee5da52bde502, R9
	SBBQ  CX, R9
	MOVQ  R9, 8(R8)
	MOVQ  $0xdec1d01aa27a1ae0, R9
	SBBQ  BX, R9
	MOVQ  R9, 16(R8)
	MOVQ  $0xd3f7498be97c5eaf, R9
	SBBQ  SI, R9
	MOVQ  R9, 24(R8)
	MOVQ  $0x04c23a02b586d650, R9
	SBBQ  DI, R9
	MOVQ  R9, 32(R8)
	RET

l1:
	MOVQ AX, 0(R8)
	MOVQ AX, 8(R8)
	MOVQ AX, 16(R8)
	MOVQ AX, 24(R8)
	MOVQ AX, 32(R8)
	RET

TEXT ·reduce(SB), NOSPLIT, $0-8
	MOVQ res+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), SI
	MOVQ 32(AX), DI

	// reduce element(DX,CX,BX,SI,DI) using temp registers (R8,R9,R10,R11,R12)
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12)

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	MOVQ SI, 24(AX)
	MOVQ DI, 32(AX)
	RET

// MulBy3(x *Element)
TEXT ·MulBy3(SB), NOSPLIT, $0-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), SI
	MOVQ 32(AX), DI
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI
	ADCQ DI, DI

	// reduce element(DX,CX,BX,SI,DI) using temp registers (R8,R9,R10,R11,R12)
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12)

	ADDQ 0(AX), DX
	ADCQ 8(AX), CX
	ADCQ 16(AX), BX
	ADCQ 24(AX), SI
	ADCQ 32(AX), DI

	// reduce element(DX,CX,BX,SI,DI) using temp registers (R13,R14,R15,R8,R9)
	REDUCE(DX,CX,BX,SI,DI,R13,R14,R15,R8,R9)

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	MOVQ SI, 24(AX)
	MOVQ DI, 32(AX)
	RET

// MulBy5(x *Element)
TEXT ·MulBy5(SB), NOSPLIT, $0-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), SI
	MOVQ 32(AX), DI
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI
	ADCQ DI, DI

	// reduce element(DX,CX,BX,SI,DI) using temp registers (R8,R9,R10,R11,R12)
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12)

	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI
	ADCQ DI, DI

	// reduce element(DX,CX,BX,SI,DI) using temp registers (R13,R14,R15,R8,R9)
	REDUCE(DX,CX,BX,SI,DI,R13,R14,R15,R8,R9)

	ADDQ 0(AX), DX
	ADCQ 8(AX), CX
	ADCQ 16(AX), BX
	ADCQ 24(AX), SI
	ADCQ 32(AX), DI

	// reduce element(DX,CX,BX,SI,DI) using temp registers (R10,R11,R12,R13,R14)
	REDUCE(DX,CX,BX,SI,DI,R10,R11,R12,R13,R14)

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	MOVQ SI, 24(AX)
	MOVQ DI, 32(AX)
	RET

// MulBy13(x *Element)
TEXT ·MulBy13(SB), $16-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), SI
	MOVQ 32(AX), DI
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI
	ADCQ DI, DI

	// reduce element(DX,CX,BX,SI,DI) using temp registers (R8,R9,R10,R11,R12)
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12)

	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI
	ADCQ DI, DI

	// reduce element(DX,CX,BX,SI,DI) using temp registers (R13,R14,R15,s0-8(SP),s1-16(SP))
	REDUCE(DX,CX,BX,SI,DI,R13,R14,R15,s0-8(SP),s1-16(SP))

	MOVQ DX, R13
	MOVQ CX, R14
	MOVQ BX, R15
	MOVQ SI, s0-8(SP)
	MOVQ DI, s1-16(SP)
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI
	ADCQ DI, DI

	// reduce element(DX,CX,BX,SI,DI) using temp registers (R8,R9,R10,R11,R12)
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12)

	ADDQ R13, DX
	ADCQ R14, CX
	ADCQ R15, BX
	ADCQ s0-8(SP), SI
	ADCQ s1-16(SP), DI

	// reduce element(DX,CX,BX,SI,DI) using temp registers (R8,R9,R10,R11,R12)
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12)

	ADDQ 0(AX), DX
	ADCQ 8(AX), CX
	ADCQ 16(AX), BX
	ADCQ 24(AX), SI
	ADCQ 32(AX), DI

	// reduce element(DX,CX,BX,SI,DI) using temp registers (R8,R9,R10,R11,R12)
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12)

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	MOVQ SI, 24(AX)
	MOVQ DI, 32(AX)
	RET

// Butterfly(a, b *Element) sets a = a + b; b = a - b
TEXT ·Butterfly(SB), $24-16
	MOVQ    a+0(FP), AX
	MOVQ    0(AX), CX
	MOVQ    8(AX), BX
	MOVQ    16(AX), SI
	MOVQ    24(AX), DI
	MOVQ    32(AX), R8
	MOVQ    CX, R9
	MOVQ    BX, R10
	MOVQ    SI, R11
	MOVQ    DI, R12
	MOVQ    R8, R13
	XORQ    AX, AX
	MOVQ    b+8(FP), DX
	ADDQ    0(DX), CX
	ADCQ    8(DX), BX
	ADCQ    16(DX), SI
	ADCQ    24(DX), DI
	ADCQ    32(DX), R8
	SUBQ    0(DX), R9
	SBBQ    8(DX), R10
	SBBQ    16(DX), R11
	SBBQ    24(DX), R12
	SBBQ    32(DX), R13
	MOVQ    CX, R14
	MOVQ    BX, R15
	MOVQ    SI, s0-8(SP)
	MOVQ    DI, s1-16(SP)
	MOVQ    R8, s2-24(SP)
	MOVQ    $0x6fe802ff40300001, CX
	MOVQ    $0x421ee5da52bde502, BX
	MOVQ    $0xdec1d01aa27a1ae0, SI
	MOVQ    $0xd3f7498be97c5eaf, DI
	MOVQ    $0x04c23a02b586d650, R8
	CMOVQCC AX, CX
	CMOVQCC AX, BX
	CMOVQCC AX, SI
	CMOVQCC AX, DI
	CMOVQCC AX, R8
	ADDQ    CX, R9
	ADCQ    BX, R10
	ADCQ    SI, R11
	ADCQ    DI, R12
	ADCQ    R8, R13
	MOVQ    R14, CX
	MOVQ    R15, BX
	MOVQ    s0-8(SP), SI
	MOVQ    s1-16(SP), DI
	MOVQ    s2-24(SP), R8
	MOVQ    R9, 0(DX)
	MOVQ    R10, 8(DX)
	MOVQ    R11, 16(DX)
	MOVQ    R12, 24(DX)
	MOVQ    R13, 32(DX)

	// reduce element(CX,BX,SI,DI,R8) using temp registers (R9,R10,R11,R12,R13)
	REDUCE(CX,BX,SI,DI,R8,R9,R10,R11,R12,R13)

	MOVQ a+0(FP), AX
	MOVQ CX, 0(AX)
	MOVQ BX, 8(AX)
	MOVQ SI, 16(AX)
	MOVQ DI, 24(AX)
	MOVQ R8, 32(AX)
	RET

// inverse(res, x *Element)
TEXT ·inverse(SB), $96-16
	// u = q
	// u[0] -> R8
	// u[1] -> R9
	// u[2] -> R10
	// u[3] -> R11
	// u[4] -> R12
	MOVQ q<>+0(SB), R8
	MOVQ q<>+8(SB), R9
	MOVQ q<>+16(SB), R10
	MOVQ q<>+24(SB), R11
	MOVQ q<>+32(SB), R12

	// s = r^2
	// s[0] -> s7-64(SP)
	// s[1] -> s8-72(SP)
	// s[2] -> s9-80(SP)
	// s[3] -> s10-88(SP)
	// s[4] -> s11-96(SP)
	MOVQ $0x6b817891fe329c16, DI
	MOVQ DI, s7-64(SP)
	MOVQ $0x599ce86eec6e2c35, DI
	MOVQ DI, s8-72(SP)
	MOVQ $0xc338890f540d5ad6, DI
	MOVQ DI, s9-80(SP)
	MOVQ $0xcc160f6924c81f32, DI
	MOVQ DI, s10-88(SP)
	MOVQ $0x0215d8d4607a88d5, DI
	MOVQ DI, s11-96(SP)

	// v = x
	// v[0] -> R13
	// v[1] -> R14
	// v[2] -> R15
	// v[3] -> s0-8(SP)
	// v[4] -> s1-16(SP)
	MOVQ x+8(FP), DI
	MOVQ 0(DI), AX
	MOVQ 8(DI), DX
	MOVQ 16(DI), CX
	MOVQ 24(DI), BX
	MOVQ 32(DI), SI
	MOVQ AX, R13
	MOVQ DX, R14
	MOVQ CX, R15
	MOVQ BX, s0-8(SP)
	MOVQ SI, s1-16(SP)

	// if x is 0, returns 0
	MOVQ AX, DI
	ORQ  DX, DI
	ORQ  CX, DI
	ORQ  BX, DI
	ORQ  SI, DI
	JEQ  l7

	// r = 0
	// r[0] -> s2-24(SP)
	// r[1] -> s3-32(SP)
	// r[2] -> s4-40(SP)
	// r[3] -> s5-48(SP)
	// r[4] -> s6-56(SP)
	MOVQ $0, s2-24(SP)
	MOVQ $0, s3-32(SP)
	MOVQ $0, s4-40(SP)
	MOVQ $0, s5-48(SP)
	MOVQ $0, s6-56(SP)

l2:
	BTQ  $0, AX
	JCS  l8
	MOVQ $0, BP
	XORQ DI, DI

l9:
	INCQ BP
	SHRQ $1, AX, DI
	SHRQ $1, DX, AX
	SHRQ $1, CX, DX
	SHRQ $1, BX, CX
	SHRQ $1, SI, BX
	SHRQ $1, SI
	BTQ  $0, AX
	JCC  l9
	MOVQ AX, R13
	MOVQ DX, R14
	MOVQ CX, R15
	MOVQ BX, s0-8(SP)
	MOVQ SI, s1-16(SP)
	MOVQ s7-64(SP), AX
	MOVQ s8-72(SP), DX
	MOVQ s9-80(SP), CX
	MOVQ s10-88(SP), BX
	MOVQ s11-96(SP), SI

l10:
	BTQ  $0, AX
	JCC  l11
	ADDQ q<>+0(SB), AX
	ADCQ q<>+8(SB), DX
	ADCQ q<>+16(SB), CX
	ADCQ q<>+24(SB), BX
	ADCQ q<>+32(SB), SI

l11:
	SHRQ $1, AX, DI
	SHRQ $1, DX, AX
	SHRQ $1, CX, DX
	SHRQ $1, BX, CX
	SHRQ $1, SI, BX
	SHRQ $1, SI
	DECQ BP
	JNE  l10
	MOVQ AX, s7-64(SP)
	MOVQ DX, s8-72(SP)
	MOVQ CX, s9-80(SP)
	MOVQ BX, s10-88(SP)
	MOVQ SI, s11-96(SP)

l8:
	MOVQ R8, AX
	MOVQ R9, DX
	MOVQ R10, CX
	MOVQ R11, BX
	MOVQ R12, SI
	BTQ  $0, AX
	JCS  l12
	MOVQ $0, BP
	XORQ DI, DI

l13:
	INCQ BP
	SHRQ $1, AX, DI
	SHRQ $1, DX, AX
	SHRQ $1, CX, DX
	SHRQ $1, BX, CX
	SHRQ $1, SI, BX
	SHRQ $1, SI
	BTQ  $0, AX
	JCC  l13
	MOVQ AX, R8
	MOVQ DX, R9
	MOVQ CX, R10
	MOVQ BX, R11
	MOVQ SI, R12
	MOVQ s2-24(SP), AX
	MOVQ s3-32(SP), DX
	MOVQ s4-40(SP), CX
	MOVQ s5-48(SP), BX
	MOVQ s6-56(SP), SI

l14:
	BTQ  $0, AX
	JCC  l15
	ADDQ q<>+0(SB), AX
	ADCQ q<>+8(SB), DX
	ADCQ q<>+16(SB), CX
	ADCQ q<>+24(SB), BX
	ADCQ q<>+32(SB), SI

l15:
	SHRQ $1, AX, DI
	SHRQ $1, DX, AX
	SHRQ $1, CX, DX
	SHRQ $1, BX, CX
	SHRQ $1, SI, BX
	SHRQ $1, SI
	DECQ BP
	JNE  l14
	MOVQ AX, s2-24(SP)
	MOVQ DX, s3-32(SP)
	MOVQ CX, s4-40(SP)
	MOVQ BX, s5-48(SP)
	MOVQ SI, s6-56(SP)

l12:
	// v = v - u
	MOVQ R13, AX
	MOVQ R14, DX
	MOVQ R15, CX
	MOVQ s0-8(SP), BX
	MOVQ s1-16(SP), SI
	SUBQ R8, AX
	SBBQ R9, DX
	SBBQ R10, CX
	SBBQ R11, BX
	SBBQ R12, SI
	JCC  l3
	SUBQ R13, R8
	SBBQ R14, R9
	SBBQ R15, R10
	SBBQ s0-8(SP), R11
	SBBQ s1-16(SP), R12
	MOVQ s2-24(SP), AX
	MOVQ s3-32(SP), DX
	MOVQ s4-40(SP), CX
	MOVQ s5-48(SP), BX
	MOVQ s6-56(SP), SI
	SUBQ s7-64(SP), AX
	SBBQ s8-72(SP), DX
	SBBQ s9-80(SP), CX
	SBBQ s10-88(SP), BX
	SBBQ s11-96(SP), SI
	JCC  l16
	ADDQ q<>+0(SB), AX
	ADCQ q<>+8(SB), DX
	ADCQ q<>+16(SB), CX
	ADCQ q<>+24(SB), BX
	ADCQ q<>+32(SB), SI

l16:
	MOVQ AX, s2-24(SP)
	MOVQ DX, s3-32(SP)
	MOVQ CX, s4-40(SP)
	MOVQ BX, s5-48(SP)
	MOVQ SI, s6-56(SP)
	JMP  l4

l3:
	MOVQ AX, R13
	MOVQ DX, R14
	MOVQ CX, R15
	MOVQ BX, s0-8(SP)
	MOVQ SI, s1-16(SP)
	MOVQ s7-64(SP), AX
	MOVQ s8-72(SP), DX
	MOVQ s9-80(SP), CX
	MOVQ s10-88(SP), BX
	MOVQ s11-96(SP), SI
	SUBQ s2-24(SP), AX
	SBBQ s3-32(SP), DX
	SBBQ s4-40(SP), CX
	SBBQ s5-48(SP), BX
	SBBQ s6-56(SP), SI
	JCC  l17
	ADDQ q<>+0(SB), AX
	ADCQ q<>+8(SB), DX
	ADCQ q<>+16(SB), CX
	ADCQ q<>+24(SB), BX
	ADCQ q<>+32(SB), SI

l17:
	MOVQ AX, s7-64(SP)
	MOVQ DX, s8-72(SP)
	MOVQ CX, s9-80(SP)
	MOVQ BX, s10-88(SP)
	MOVQ SI, s11-96(SP)

l4:
	MOVQ R8, DI
	SUBQ $1, DI
	ORQ  R9, DI
	ORQ  R10, DI
	ORQ  R11, DI
	ORQ  R12, DI
	JEQ  l5
	MOVQ R13, AX
	MOVQ R14, DX
	MOVQ R15, CX
	MOVQ s0-8(SP), BX
	MOVQ s1-16(SP), SI
	MOVQ AX, DI
	SUBQ $1, DI
	JNE  l2
	ORQ  DX, DI
	ORQ  CX, DI
	ORQ  BX, DI
	ORQ  SI, DI
	JEQ  l6
	JMP  l2

l5:
	MOVQ res+0(FP), DI
	MOVQ s2-24(SP), AX
	MOVQ s3-32(SP), DX
	MOVQ s4-40(SP), CX
	MOVQ s5-48(SP), BX
	MOVQ s6-56(SP), SI
	MOVQ AX, 0(DI)
	MOVQ DX, 8(DI)
	MOVQ CX, 16(DI)
	MOVQ BX, 24(DI)
	MOVQ SI, 32(DI)
	RET

l6:
	MOVQ res+0(FP), DI
	MOVQ s7-64(SP), AX
	MOVQ s8-72(SP), DX
	MOVQ s9-80(SP), CX
	MOVQ s10-88(SP), BX
	MOVQ s11-96(SP), SI
	MOVQ AX, 0(DI)
	MOVQ DX, 8(DI)
	MOVQ CX, 16(DI)
	MOVQ BX, 24(DI)
	MOVQ SI, 32(DI)
	RET

l7:
	MOVQ res+0(FP), DI
	MOVQ $0, 0(DI)
	MOVQ $0, 8(DI)
	MOVQ $0, 16(DI)
	MOVQ $0, 24(DI)
	MOVQ $0, 32(DI)
	RET
