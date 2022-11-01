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
DATA q<>+0(SB)/8, $0x6e08691dd8d00acf
DATA q<>+8(SB)/8, $0x72bb251bfc52ed43
DATA q<>+16(SB)/8, $0x1d819dc22c277791
GLOBL q<>(SB), (RODATA+NOPTR), $24

// qInv0 q'[0]
DATA qInv0<>(SB)/8, $0x34caaffd883d43d1
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
	MOVQ    $0x6e08691dd8d00acf, R10
	MOVQ    $0x72bb251bfc52ed43, R11
	MOVQ    $0x1d819dc22c277791, R12
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
