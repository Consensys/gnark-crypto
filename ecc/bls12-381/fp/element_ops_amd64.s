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
DATA q<>+0(SB)/8, $0xb9feffffffffaaab
DATA q<>+8(SB)/8, $0x1eabfffeb153ffff
DATA q<>+16(SB)/8, $0x6730d2a0f6b0f624
DATA q<>+24(SB)/8, $0x64774b84f38512bf
DATA q<>+32(SB)/8, $0x4b1ba7b6434bacd7
DATA q<>+40(SB)/8, $0x1a0111ea397fe69a
GLOBL q<>(SB), (RODATA+NOPTR), $48

// qInv0 q'[0]
DATA qInv0<>(SB)/8, $0x89f3fffcfffcfffd
GLOBL qInv0<>(SB), (RODATA+NOPTR), $8

#define REDUCE(ra0, ra1, ra2, ra3, ra4, ra5, rb0, rb1, rb2, rb3, rb4, rb5) \
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
	MOVQ    ra5, rb5;        \
	SBBQ    q<>+40(SB), ra5; \
	CMOVQCS rb0, ra0;        \
	CMOVQCS rb1, ra1;        \
	CMOVQCS rb2, ra2;        \
	CMOVQCS rb3, ra3;        \
	CMOVQCS rb4, ra4;        \
	CMOVQCS rb5, ra5;        \

// add(res, x, y *Element)
TEXT ·add(SB), NOSPLIT, $0-24
	MOVQ x+8(FP), AX
	MOVQ 0(AX), CX
	MOVQ 8(AX), BX
	MOVQ 16(AX), SI
	MOVQ 24(AX), DI
	MOVQ 32(AX), R8
	MOVQ 40(AX), R9
	MOVQ y+16(FP), DX
	ADDQ 0(DX), CX
	ADCQ 8(DX), BX
	ADCQ 16(DX), SI
	ADCQ 24(DX), DI
	ADCQ 32(DX), R8
	ADCQ 40(DX), R9

	// reduce element(CX,BX,SI,DI,R8,R9) using temp registers (R10,R11,R12,R13,R14,R15)
	REDUCE(CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15)

	MOVQ res+0(FP), AX
	MOVQ CX, 0(AX)
	MOVQ BX, 8(AX)
	MOVQ SI, 16(AX)
	MOVQ DI, 24(AX)
	MOVQ R8, 32(AX)
	MOVQ R9, 40(AX)
	RET

// sub(res, x, y *Element)
TEXT ·sub(SB), NOSPLIT, $0-24
	XORQ    R9, R9
	MOVQ    x+8(FP), R8
	MOVQ    0(R8), AX
	MOVQ    8(R8), DX
	MOVQ    16(R8), CX
	MOVQ    24(R8), BX
	MOVQ    32(R8), SI
	MOVQ    40(R8), DI
	MOVQ    y+16(FP), R8
	SUBQ    0(R8), AX
	SBBQ    8(R8), DX
	SBBQ    16(R8), CX
	SBBQ    24(R8), BX
	SBBQ    32(R8), SI
	SBBQ    40(R8), DI
	MOVQ    $0xb9feffffffffaaab, R10
	MOVQ    $0x1eabfffeb153ffff, R11
	MOVQ    $0x6730d2a0f6b0f624, R12
	MOVQ    $0x64774b84f38512bf, R13
	MOVQ    $0x4b1ba7b6434bacd7, R14
	MOVQ    $0x1a0111ea397fe69a, R15
	CMOVQCC R9, R10
	CMOVQCC R9, R11
	CMOVQCC R9, R12
	CMOVQCC R9, R13
	CMOVQCC R9, R14
	CMOVQCC R9, R15
	ADDQ    R10, AX
	ADCQ    R11, DX
	ADCQ    R12, CX
	ADCQ    R13, BX
	ADCQ    R14, SI
	ADCQ    R15, DI
	MOVQ    res+0(FP), R8
	MOVQ    AX, 0(R8)
	MOVQ    DX, 8(R8)
	MOVQ    CX, 16(R8)
	MOVQ    BX, 24(R8)
	MOVQ    SI, 32(R8)
	MOVQ    DI, 40(R8)
	RET

// double(res, x *Element)
TEXT ·double(SB), NOSPLIT, $0-16
	MOVQ x+8(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), SI
	MOVQ 32(AX), DI
	MOVQ 40(AX), R8
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI
	ADCQ DI, DI
	ADCQ R8, R8

	// reduce element(DX,CX,BX,SI,DI,R8) using temp registers (R9,R10,R11,R12,R13,R14)
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14)

	MOVQ res+0(FP), R15
	MOVQ DX, 0(R15)
	MOVQ CX, 8(R15)
	MOVQ BX, 16(R15)
	MOVQ SI, 24(R15)
	MOVQ DI, 32(R15)
	MOVQ R8, 40(R15)
	RET

// neg(res, x *Element)
TEXT ·neg(SB), NOSPLIT, $0-16
	MOVQ  res+0(FP), R9
	MOVQ  x+8(FP), AX
	MOVQ  0(AX), DX
	MOVQ  8(AX), CX
	MOVQ  16(AX), BX
	MOVQ  24(AX), SI
	MOVQ  32(AX), DI
	MOVQ  40(AX), R8
	MOVQ  DX, AX
	ORQ   CX, AX
	ORQ   BX, AX
	ORQ   SI, AX
	ORQ   DI, AX
	ORQ   R8, AX
	TESTQ AX, AX
	JEQ   l1
	MOVQ  $0xb9feffffffffaaab, R10
	SUBQ  DX, R10
	MOVQ  R10, 0(R9)
	MOVQ  $0x1eabfffeb153ffff, R10
	SBBQ  CX, R10
	MOVQ  R10, 8(R9)
	MOVQ  $0x6730d2a0f6b0f624, R10
	SBBQ  BX, R10
	MOVQ  R10, 16(R9)
	MOVQ  $0x64774b84f38512bf, R10
	SBBQ  SI, R10
	MOVQ  R10, 24(R9)
	MOVQ  $0x4b1ba7b6434bacd7, R10
	SBBQ  DI, R10
	MOVQ  R10, 32(R9)
	MOVQ  $0x1a0111ea397fe69a, R10
	SBBQ  R8, R10
	MOVQ  R10, 40(R9)
	RET

l1:
	MOVQ AX, 0(R9)
	MOVQ AX, 8(R9)
	MOVQ AX, 16(R9)
	MOVQ AX, 24(R9)
	MOVQ AX, 32(R9)
	MOVQ AX, 40(R9)
	RET

TEXT ·reduce(SB), NOSPLIT, $0-8
	MOVQ res+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), SI
	MOVQ 32(AX), DI
	MOVQ 40(AX), R8

	// reduce element(DX,CX,BX,SI,DI,R8) using temp registers (R9,R10,R11,R12,R13,R14)
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14)

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	MOVQ SI, 24(AX)
	MOVQ DI, 32(AX)
	MOVQ R8, 40(AX)
	RET

// MulBy3(x *Element)
TEXT ·MulBy3(SB), NOSPLIT, $0-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), SI
	MOVQ 32(AX), DI
	MOVQ 40(AX), R8
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI
	ADCQ DI, DI
	ADCQ R8, R8

	// reduce element(DX,CX,BX,SI,DI,R8) using temp registers (R9,R10,R11,R12,R13,R14)
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14)

	ADDQ 0(AX), DX
	ADCQ 8(AX), CX
	ADCQ 16(AX), BX
	ADCQ 24(AX), SI
	ADCQ 32(AX), DI
	ADCQ 40(AX), R8

	// reduce element(DX,CX,BX,SI,DI,R8) using temp registers (R15,R9,R10,R11,R12,R13)
	REDUCE(DX,CX,BX,SI,DI,R8,R15,R9,R10,R11,R12,R13)

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	MOVQ SI, 24(AX)
	MOVQ DI, 32(AX)
	MOVQ R8, 40(AX)
	RET

// MulBy5(x *Element)
TEXT ·MulBy5(SB), NOSPLIT, $0-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), SI
	MOVQ 32(AX), DI
	MOVQ 40(AX), R8
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI
	ADCQ DI, DI
	ADCQ R8, R8

	// reduce element(DX,CX,BX,SI,DI,R8) using temp registers (R9,R10,R11,R12,R13,R14)
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14)

	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI
	ADCQ DI, DI
	ADCQ R8, R8

	// reduce element(DX,CX,BX,SI,DI,R8) using temp registers (R15,R9,R10,R11,R12,R13)
	REDUCE(DX,CX,BX,SI,DI,R8,R15,R9,R10,R11,R12,R13)

	ADDQ 0(AX), DX
	ADCQ 8(AX), CX
	ADCQ 16(AX), BX
	ADCQ 24(AX), SI
	ADCQ 32(AX), DI
	ADCQ 40(AX), R8

	// reduce element(DX,CX,BX,SI,DI,R8) using temp registers (R14,R15,R9,R10,R11,R12)
	REDUCE(DX,CX,BX,SI,DI,R8,R14,R15,R9,R10,R11,R12)

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	MOVQ SI, 24(AX)
	MOVQ DI, 32(AX)
	MOVQ R8, 40(AX)
	RET

// MulBy13(x *Element)
TEXT ·MulBy13(SB), $40-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), SI
	MOVQ 32(AX), DI
	MOVQ 40(AX), R8
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI
	ADCQ DI, DI
	ADCQ R8, R8

	// reduce element(DX,CX,BX,SI,DI,R8) using temp registers (R9,R10,R11,R12,R13,R14)
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14)

	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI
	ADCQ DI, DI
	ADCQ R8, R8

	// reduce element(DX,CX,BX,SI,DI,R8) using temp registers (R15,s0-8(SP),s1-16(SP),s2-24(SP),s3-32(SP),s4-40(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R15,s0-8(SP),s1-16(SP),s2-24(SP),s3-32(SP),s4-40(SP))

	MOVQ DX, R15
	MOVQ CX, s0-8(SP)
	MOVQ BX, s1-16(SP)
	MOVQ SI, s2-24(SP)
	MOVQ DI, s3-32(SP)
	MOVQ R8, s4-40(SP)
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI
	ADCQ DI, DI
	ADCQ R8, R8

	// reduce element(DX,CX,BX,SI,DI,R8) using temp registers (R9,R10,R11,R12,R13,R14)
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14)

	ADDQ R15, DX
	ADCQ s0-8(SP), CX
	ADCQ s1-16(SP), BX
	ADCQ s2-24(SP), SI
	ADCQ s3-32(SP), DI
	ADCQ s4-40(SP), R8

	// reduce element(DX,CX,BX,SI,DI,R8) using temp registers (R9,R10,R11,R12,R13,R14)
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14)

	ADDQ 0(AX), DX
	ADCQ 8(AX), CX
	ADCQ 16(AX), BX
	ADCQ 24(AX), SI
	ADCQ 32(AX), DI
	ADCQ 40(AX), R8

	// reduce element(DX,CX,BX,SI,DI,R8) using temp registers (R9,R10,R11,R12,R13,R14)
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14)

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	MOVQ SI, 24(AX)
	MOVQ DI, 32(AX)
	MOVQ R8, 40(AX)
	RET
