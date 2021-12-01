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
DATA q<>+0(SB)/8, $0xdd64f813fcb4c2a1
DATA q<>+8(SB)/8, $0x1aa90fd187823ec8
GLOBL q<>(SB), (RODATA+NOPTR), $16

// qInv0 q'[0]
DATA qInv0<>(SB)/8, $0x1378842f09c45e9f
GLOBL qInv0<>(SB), (RODATA+NOPTR), $8

#define REDUCE(ra0, ra1, rb0, rb1) \
	MOVQ    ra0, rb0;       \
	SUBQ    q<>(SB), ra0;   \
	MOVQ    ra1, rb1;       \
	SBBQ    q<>+8(SB), ra1; \
	CMOVQCS rb0, ra0;       \
	CMOVQCS rb1, ra1;       \

// add(res, x, y *Element)
TEXT ·add(SB), NOSPLIT, $0-24
	MOVQ x+8(FP), AX
	MOVQ 0(AX), CX
	MOVQ 8(AX), BX
	MOVQ y+16(FP), DX
	ADDQ 0(DX), CX
	ADCQ 8(DX), BX

	// reduce element(CX,BX) using temp registers (SI,DI)
	REDUCE(CX,BX,SI,DI)

	MOVQ res+0(FP), R8
	MOVQ CX, 0(R8)
	MOVQ BX, 8(R8)
	RET

// sub(res, x, y *Element)
TEXT ·sub(SB), NOSPLIT, $0-24
	XORQ    BX, BX
	MOVQ    x+8(FP), CX
	MOVQ    0(CX), AX
	MOVQ    8(CX), DX
	MOVQ    y+16(FP), CX
	SUBQ    0(CX), AX
	SBBQ    8(CX), DX
	MOVQ    $0xdd64f813fcb4c2a1, SI
	MOVQ    $0x1aa90fd187823ec8, DI
	CMOVQCC BX, SI
	CMOVQCC BX, DI
	ADDQ    SI, AX
	ADCQ    DI, DX
	MOVQ    res+0(FP), R8
	MOVQ    AX, 0(R8)
	MOVQ    DX, 8(R8)
	RET

// double(res, x *Element)
TEXT ·double(SB), NOSPLIT, $0-16
	MOVQ x+8(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	ADDQ DX, DX
	ADCQ CX, CX

	// reduce element(DX,CX) using temp registers (BX,SI)
	REDUCE(DX,CX,BX,SI)

	MOVQ res+0(FP), DI
	MOVQ DX, 0(DI)
	MOVQ CX, 8(DI)
	RET

// neg(res, x *Element)
TEXT ·neg(SB), NOSPLIT, $0-16
	MOVQ  res+0(FP), BX
	MOVQ  x+8(FP), AX
	MOVQ  0(AX), DX
	MOVQ  8(AX), CX
	MOVQ  DX, AX
	ORQ   CX, AX
	TESTQ AX, AX
	JEQ   l1
	MOVQ  $0xdd64f813fcb4c2a1, SI
	SUBQ  DX, SI
	MOVQ  SI, 0(BX)
	MOVQ  $0x1aa90fd187823ec8, SI
	SBBQ  CX, SI
	MOVQ  SI, 8(BX)
	RET

l1:
	MOVQ AX, 0(BX)
	MOVQ AX, 8(BX)
	RET

TEXT ·reduce(SB), NOSPLIT, $0-8
	MOVQ res+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX

	// reduce element(DX,CX) using temp registers (BX,SI)
	REDUCE(DX,CX,BX,SI)

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	RET

// MulBy3(x *Element)
TEXT ·MulBy3(SB), NOSPLIT, $0-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	ADDQ DX, DX
	ADCQ CX, CX

	// reduce element(DX,CX) using temp registers (BX,SI)
	REDUCE(DX,CX,BX,SI)

	ADDQ 0(AX), DX
	ADCQ 8(AX), CX

	// reduce element(DX,CX) using temp registers (DI,R8)
	REDUCE(DX,CX,DI,R8)

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	RET

// MulBy5(x *Element)
TEXT ·MulBy5(SB), NOSPLIT, $0-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	ADDQ DX, DX
	ADCQ CX, CX

	// reduce element(DX,CX) using temp registers (BX,SI)
	REDUCE(DX,CX,BX,SI)

	ADDQ DX, DX
	ADCQ CX, CX

	// reduce element(DX,CX) using temp registers (DI,R8)
	REDUCE(DX,CX,DI,R8)

	ADDQ 0(AX), DX
	ADCQ 8(AX), CX

	// reduce element(DX,CX) using temp registers (R9,R10)
	REDUCE(DX,CX,R9,R10)

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	RET

// MulBy13(x *Element)
TEXT ·MulBy13(SB), NOSPLIT, $0-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	ADDQ DX, DX
	ADCQ CX, CX

	// reduce element(DX,CX) using temp registers (BX,SI)
	REDUCE(DX,CX,BX,SI)

	ADDQ DX, DX
	ADCQ CX, CX

	// reduce element(DX,CX) using temp registers (DI,R8)
	REDUCE(DX,CX,DI,R8)

	MOVQ DX, DI
	MOVQ CX, R8
	ADDQ DX, DX
	ADCQ CX, CX

	// reduce element(DX,CX) using temp registers (BX,SI)
	REDUCE(DX,CX,BX,SI)

	ADDQ DI, DX
	ADCQ R8, CX

	// reduce element(DX,CX) using temp registers (BX,SI)
	REDUCE(DX,CX,BX,SI)

	ADDQ 0(AX), DX
	ADCQ 8(AX), CX

	// reduce element(DX,CX) using temp registers (BX,SI)
	REDUCE(DX,CX,BX,SI)

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	RET

// Butterfly(a, b *Element) sets a = a + b; b = a - b
TEXT ·Butterfly(SB), NOSPLIT, $0-16
	MOVQ    a+0(FP), AX
	MOVQ    0(AX), CX
	MOVQ    8(AX), BX
	MOVQ    CX, SI
	MOVQ    BX, DI
	XORQ    AX, AX
	MOVQ    b+8(FP), DX
	ADDQ    0(DX), CX
	ADCQ    8(DX), BX
	SUBQ    0(DX), SI
	SBBQ    8(DX), DI
	MOVQ    $0xdd64f813fcb4c2a1, R8
	MOVQ    $0x1aa90fd187823ec8, R9
	CMOVQCC AX, R8
	CMOVQCC AX, R9
	ADDQ    R8, SI
	ADCQ    R9, DI
	MOVQ    SI, 0(DX)
	MOVQ    DI, 8(DX)

	// reduce element(CX,BX) using temp registers (SI,DI)
	REDUCE(CX,BX,SI,DI)

	MOVQ a+0(FP), AX
	MOVQ CX, 0(AX)
	MOVQ BX, 8(AX)
	RET
