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
DATA q<>+0(SB)/8, $0x6220bb2726ec502d
DATA q<>+8(SB)/8, $0x1287818c76907fc8
DATA q<>+16(SB)/8, $0x0300dd268771ce96
GLOBL q<>(SB), (RODATA+NOPTR), $24

// qInv0 q'[0]
DATA qInv0<>(SB)/8, $0xe6d76790d22d805b
GLOBL qInv0<>(SB), (RODATA+NOPTR), $8

#define REDUCE(ra0, ra1, ra2, rb0, rb1, rb2) \
	MOVQ    ra0, rb0;        \
	SUBQ    q<>(SB), ra0;    \
	MOVQ    ra1, rb1;        \
	SBBQ    q<>+8(SB), ra1;  \
	MOVQ    ra2, rb2;        \
	SBBQ    q<>+16(SB), ra2; \
	CMOVQCS rb0, ra0;        \
	CMOVQCS rb1, ra1;        \
	CMOVQCS rb2, ra2;        \

// add(res, x, y *Element)
TEXT ·add(SB), NOSPLIT, $0-24
	MOVQ x+8(FP), AX
	MOVQ 0(AX), CX
	MOVQ 8(AX), BX
	MOVQ 16(AX), SI
	MOVQ y+16(FP), DX
	ADDQ 0(DX), CX
	ADCQ 8(DX), BX
	ADCQ 16(DX), SI

	// reduce element(CX,BX,SI) using temp registers (DI,R8,R9)
	REDUCE(CX,BX,SI,DI,R8,R9)

	MOVQ res+0(FP), R10
	MOVQ CX, 0(R10)
	MOVQ BX, 8(R10)
	MOVQ SI, 16(R10)
	RET

// sub(res, x, y *Element)
TEXT ·sub(SB), NOSPLIT, $0-24
	XORQ    SI, SI
	MOVQ    x+8(FP), BX
	MOVQ    0(BX), AX
	MOVQ    8(BX), DX
	MOVQ    16(BX), CX
	MOVQ    y+16(FP), BX
	SUBQ    0(BX), AX
	SBBQ    8(BX), DX
	SBBQ    16(BX), CX
	MOVQ    $0x6220bb2726ec502d, DI
	MOVQ    $0x1287818c76907fc8, R8
	MOVQ    $0x0300dd268771ce96, R9
	CMOVQCC SI, DI
	CMOVQCC SI, R8
	CMOVQCC SI, R9
	ADDQ    DI, AX
	ADCQ    R8, DX
	ADCQ    R9, CX
	MOVQ    res+0(FP), R10
	MOVQ    AX, 0(R10)
	MOVQ    DX, 8(R10)
	MOVQ    CX, 16(R10)
	RET

// double(res, x *Element)
TEXT ·double(SB), NOSPLIT, $0-16
	MOVQ x+8(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX

	// reduce element(DX,CX,BX) using temp registers (SI,DI,R8)
	REDUCE(DX,CX,BX,SI,DI,R8)

	MOVQ res+0(FP), R9
	MOVQ DX, 0(R9)
	MOVQ CX, 8(R9)
	MOVQ BX, 16(R9)
	RET

// neg(res, x *Element)
TEXT ·neg(SB), NOSPLIT, $0-16
	MOVQ  res+0(FP), SI
	MOVQ  x+8(FP), AX
	MOVQ  0(AX), DX
	MOVQ  8(AX), CX
	MOVQ  16(AX), BX
	MOVQ  DX, AX
	ORQ   CX, AX
	ORQ   BX, AX
	TESTQ AX, AX
	JEQ   l1
	MOVQ  $0x6220bb2726ec502d, DI
	SUBQ  DX, DI
	MOVQ  DI, 0(SI)
	MOVQ  $0x1287818c76907fc8, DI
	SBBQ  CX, DI
	MOVQ  DI, 8(SI)
	MOVQ  $0x0300dd268771ce96, DI
	SBBQ  BX, DI
	MOVQ  DI, 16(SI)
	RET

l1:
	MOVQ AX, 0(SI)
	MOVQ AX, 8(SI)
	MOVQ AX, 16(SI)
	RET

TEXT ·reduce(SB), NOSPLIT, $0-8
	MOVQ res+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX

	// reduce element(DX,CX,BX) using temp registers (SI,DI,R8)
	REDUCE(DX,CX,BX,SI,DI,R8)

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	RET

// MulBy3(x *Element)
TEXT ·MulBy3(SB), NOSPLIT, $0-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX

	// reduce element(DX,CX,BX) using temp registers (SI,DI,R8)
	REDUCE(DX,CX,BX,SI,DI,R8)

	ADDQ 0(AX), DX
	ADCQ 8(AX), CX
	ADCQ 16(AX), BX

	// reduce element(DX,CX,BX) using temp registers (R9,R10,R11)
	REDUCE(DX,CX,BX,R9,R10,R11)

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	RET

// MulBy5(x *Element)
TEXT ·MulBy5(SB), NOSPLIT, $0-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX

	// reduce element(DX,CX,BX) using temp registers (SI,DI,R8)
	REDUCE(DX,CX,BX,SI,DI,R8)

	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX

	// reduce element(DX,CX,BX) using temp registers (R9,R10,R11)
	REDUCE(DX,CX,BX,R9,R10,R11)

	ADDQ 0(AX), DX
	ADCQ 8(AX), CX
	ADCQ 16(AX), BX

	// reduce element(DX,CX,BX) using temp registers (R12,R13,R14)
	REDUCE(DX,CX,BX,R12,R13,R14)

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	RET

// MulBy13(x *Element)
TEXT ·MulBy13(SB), NOSPLIT, $0-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX

	// reduce element(DX,CX,BX) using temp registers (SI,DI,R8)
	REDUCE(DX,CX,BX,SI,DI,R8)

	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX

	// reduce element(DX,CX,BX) using temp registers (R9,R10,R11)
	REDUCE(DX,CX,BX,R9,R10,R11)

	MOVQ DX, R9
	MOVQ CX, R10
	MOVQ BX, R11
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX

	// reduce element(DX,CX,BX) using temp registers (SI,DI,R8)
	REDUCE(DX,CX,BX,SI,DI,R8)

	ADDQ R9, DX
	ADCQ R10, CX
	ADCQ R11, BX

	// reduce element(DX,CX,BX) using temp registers (SI,DI,R8)
	REDUCE(DX,CX,BX,SI,DI,R8)

	ADDQ 0(AX), DX
	ADCQ 8(AX), CX
	ADCQ 16(AX), BX

	// reduce element(DX,CX,BX) using temp registers (SI,DI,R8)
	REDUCE(DX,CX,BX,SI,DI,R8)

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	RET

// Butterfly(a, b *Element) sets a = a + b; b = a - b
TEXT ·Butterfly(SB), NOSPLIT, $0-16
	MOVQ    a+0(FP), AX
	MOVQ    0(AX), CX
	MOVQ    8(AX), BX
	MOVQ    16(AX), SI
	MOVQ    CX, DI
	MOVQ    BX, R8
	MOVQ    SI, R9
	XORQ    AX, AX
	MOVQ    b+8(FP), DX
	ADDQ    0(DX), CX
	ADCQ    8(DX), BX
	ADCQ    16(DX), SI
	SUBQ    0(DX), DI
	SBBQ    8(DX), R8
	SBBQ    16(DX), R9
	MOVQ    $0x6220bb2726ec502d, R10
	MOVQ    $0x1287818c76907fc8, R11
	MOVQ    $0x0300dd268771ce96, R12
	CMOVQCC AX, R10
	CMOVQCC AX, R11
	CMOVQCC AX, R12
	ADDQ    R10, DI
	ADCQ    R11, R8
	ADCQ    R12, R9
	MOVQ    DI, 0(DX)
	MOVQ    R8, 8(DX)
	MOVQ    R9, 16(DX)

	// reduce element(CX,BX,SI) using temp registers (DI,R8,R9)
	REDUCE(CX,BX,SI,DI,R8,R9)

	MOVQ a+0(FP), AX
	MOVQ CX, 0(AX)
	MOVQ BX, 8(AX)
	MOVQ SI, 16(AX)
	RET
