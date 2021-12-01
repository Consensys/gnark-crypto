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
DATA q<>+0(SB)/8, $0x1c0f4efe6b9aeba5
DATA q<>+8(SB)/8, $0x07cc708d578478f9
DATA q<>+16(SB)/8, $0x7a1b80da7e024bc9
DATA q<>+24(SB)/8, $0x458f60da10ef1c57
DATA q<>+32(SB)/8, $0xbebaa6e2636b53a8
DATA q<>+40(SB)/8, $0xf92922e40eef678b
DATA q<>+48(SB)/8, $0x0000000000000003
GLOBL q<>(SB), (RODATA+NOPTR), $56

// qInv0 q'[0]
DATA qInv0<>(SB)/8, $0xd69d9e6dc221fbd3
GLOBL qInv0<>(SB), (RODATA+NOPTR), $8

#define REDUCE(ra0, ra1, ra2, ra3, ra4, ra5, ra6, rb0, rb1, rb2, rb3, rb4, rb5, rb6) \
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
	MOVQ    ra6, rb6;        \
	SBBQ    q<>+48(SB), ra6; \
	CMOVQCS rb0, ra0;        \
	CMOVQCS rb1, ra1;        \
	CMOVQCS rb2, ra2;        \
	CMOVQCS rb3, ra3;        \
	CMOVQCS rb4, ra4;        \
	CMOVQCS rb5, ra5;        \
	CMOVQCS rb6, ra6;        \

// add(res, x, y *Element)
TEXT ·add(SB), NOSPLIT, $0-24
	MOVQ x+8(FP), AX
	MOVQ 0(AX), CX
	MOVQ 8(AX), BX
	MOVQ 16(AX), SI
	MOVQ 24(AX), DI
	MOVQ 32(AX), R8
	MOVQ 40(AX), R9
	MOVQ 48(AX), R10
	MOVQ y+16(FP), DX
	ADDQ 0(DX), CX
	ADCQ 8(DX), BX
	ADCQ 16(DX), SI
	ADCQ 24(DX), DI
	ADCQ 32(DX), R8
	ADCQ 40(DX), R9
	ADCQ 48(DX), R10

	// reduce element(CX,BX,SI,DI,R8,R9,R10) using temp registers (R11,R12,R13,R14,R15,AX,DX)
	REDUCE(CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,AX,DX)

	MOVQ res+0(FP), R11
	MOVQ CX, 0(R11)
	MOVQ BX, 8(R11)
	MOVQ SI, 16(R11)
	MOVQ DI, 24(R11)
	MOVQ R8, 32(R11)
	MOVQ R9, 40(R11)
	MOVQ R10, 48(R11)
	RET

// sub(res, x, y *Element)
TEXT ·sub(SB), NOSPLIT, $0-24
	MOVQ x+8(FP), R9
	MOVQ 0(R9), AX
	MOVQ 8(R9), DX
	MOVQ 16(R9), CX
	MOVQ 24(R9), BX
	MOVQ 32(R9), SI
	MOVQ 40(R9), DI
	MOVQ 48(R9), R8
	MOVQ y+16(FP), R9
	SUBQ 0(R9), AX
	SBBQ 8(R9), DX
	SBBQ 16(R9), CX
	SBBQ 24(R9), BX
	SBBQ 32(R9), SI
	SBBQ 40(R9), DI
	SBBQ 48(R9), R8
	JCC  l1
	MOVQ $0x1c0f4efe6b9aeba5, R10
	ADDQ R10, AX
	MOVQ $0x07cc708d578478f9, R10
	ADCQ R10, DX
	MOVQ $0x7a1b80da7e024bc9, R10
	ADCQ R10, CX
	MOVQ $0x458f60da10ef1c57, R10
	ADCQ R10, BX
	MOVQ $0xbebaa6e2636b53a8, R10
	ADCQ R10, SI
	MOVQ $0xf92922e40eef678b, R10
	ADCQ R10, DI
	MOVQ $0x0000000000000003, R10
	ADCQ R10, R8

l1:
	MOVQ res+0(FP), R11
	MOVQ AX, 0(R11)
	MOVQ DX, 8(R11)
	MOVQ CX, 16(R11)
	MOVQ BX, 24(R11)
	MOVQ SI, 32(R11)
	MOVQ DI, 40(R11)
	MOVQ R8, 48(R11)
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
	MOVQ 48(AX), R9
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI
	ADCQ DI, DI
	ADCQ R8, R8
	ADCQ R9, R9

	// reduce element(DX,CX,BX,SI,DI,R8,R9) using temp registers (R10,R11,R12,R13,R14,R15,AX)
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,AX)

	MOVQ res+0(FP), R10
	MOVQ DX, 0(R10)
	MOVQ CX, 8(R10)
	MOVQ BX, 16(R10)
	MOVQ SI, 24(R10)
	MOVQ DI, 32(R10)
	MOVQ R8, 40(R10)
	MOVQ R9, 48(R10)
	RET

// neg(res, x *Element)
TEXT ·neg(SB), NOSPLIT, $0-16
	MOVQ  res+0(FP), R10
	MOVQ  x+8(FP), AX
	MOVQ  0(AX), DX
	MOVQ  8(AX), CX
	MOVQ  16(AX), BX
	MOVQ  24(AX), SI
	MOVQ  32(AX), DI
	MOVQ  40(AX), R8
	MOVQ  48(AX), R9
	MOVQ  DX, AX
	ORQ   CX, AX
	ORQ   BX, AX
	ORQ   SI, AX
	ORQ   DI, AX
	ORQ   R8, AX
	ORQ   R9, AX
	TESTQ AX, AX
	JEQ   l2
	MOVQ  $0x1c0f4efe6b9aeba5, R11
	SUBQ  DX, R11
	MOVQ  R11, 0(R10)
	MOVQ  $0x07cc708d578478f9, R11
	SBBQ  CX, R11
	MOVQ  R11, 8(R10)
	MOVQ  $0x7a1b80da7e024bc9, R11
	SBBQ  BX, R11
	MOVQ  R11, 16(R10)
	MOVQ  $0x458f60da10ef1c57, R11
	SBBQ  SI, R11
	MOVQ  R11, 24(R10)
	MOVQ  $0xbebaa6e2636b53a8, R11
	SBBQ  DI, R11
	MOVQ  R11, 32(R10)
	MOVQ  $0xf92922e40eef678b, R11
	SBBQ  R8, R11
	MOVQ  R11, 40(R10)
	MOVQ  $0x0000000000000003, R11
	SBBQ  R9, R11
	MOVQ  R11, 48(R10)
	RET

l2:
	MOVQ AX, 0(R10)
	MOVQ AX, 8(R10)
	MOVQ AX, 16(R10)
	MOVQ AX, 24(R10)
	MOVQ AX, 32(R10)
	MOVQ AX, 40(R10)
	MOVQ AX, 48(R10)
	RET

TEXT ·reduce(SB), $8-8
	MOVQ res+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), SI
	MOVQ 32(AX), DI
	MOVQ 40(AX), R8
	MOVQ 48(AX), R9

	// reduce element(DX,CX,BX,SI,DI,R8,R9) using temp registers (R10,R11,R12,R13,R14,R15,s0-8(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,s0-8(SP))

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	MOVQ SI, 24(AX)
	MOVQ DI, 32(AX)
	MOVQ R8, 40(AX)
	MOVQ R9, 48(AX)
	RET

// MulBy3(x *Element)
TEXT ·MulBy3(SB), $8-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), SI
	MOVQ 32(AX), DI
	MOVQ 40(AX), R8
	MOVQ 48(AX), R9
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI
	ADCQ DI, DI
	ADCQ R8, R8
	ADCQ R9, R9

	// reduce element(DX,CX,BX,SI,DI,R8,R9) using temp registers (R10,R11,R12,R13,R14,R15,s0-8(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,s0-8(SP))

	ADDQ 0(AX), DX
	ADCQ 8(AX), CX
	ADCQ 16(AX), BX
	ADCQ 24(AX), SI
	ADCQ 32(AX), DI
	ADCQ 40(AX), R8
	ADCQ 48(AX), R9

	// reduce element(DX,CX,BX,SI,DI,R8,R9) using temp registers (R10,R11,R12,R13,R14,R15,s0-8(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,s0-8(SP))

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	MOVQ SI, 24(AX)
	MOVQ DI, 32(AX)
	MOVQ R8, 40(AX)
	MOVQ R9, 48(AX)
	RET

// MulBy5(x *Element)
TEXT ·MulBy5(SB), $8-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), SI
	MOVQ 32(AX), DI
	MOVQ 40(AX), R8
	MOVQ 48(AX), R9
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI
	ADCQ DI, DI
	ADCQ R8, R8
	ADCQ R9, R9

	// reduce element(DX,CX,BX,SI,DI,R8,R9) using temp registers (R10,R11,R12,R13,R14,R15,s0-8(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,s0-8(SP))

	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI
	ADCQ DI, DI
	ADCQ R8, R8
	ADCQ R9, R9

	// reduce element(DX,CX,BX,SI,DI,R8,R9) using temp registers (R10,R11,R12,R13,R14,R15,s0-8(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,s0-8(SP))

	ADDQ 0(AX), DX
	ADCQ 8(AX), CX
	ADCQ 16(AX), BX
	ADCQ 24(AX), SI
	ADCQ 32(AX), DI
	ADCQ 40(AX), R8
	ADCQ 48(AX), R9

	// reduce element(DX,CX,BX,SI,DI,R8,R9) using temp registers (R10,R11,R12,R13,R14,R15,s0-8(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,s0-8(SP))

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	MOVQ SI, 24(AX)
	MOVQ DI, 32(AX)
	MOVQ R8, 40(AX)
	MOVQ R9, 48(AX)
	RET

// MulBy13(x *Element)
TEXT ·MulBy13(SB), $64-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), SI
	MOVQ 32(AX), DI
	MOVQ 40(AX), R8
	MOVQ 48(AX), R9
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI
	ADCQ DI, DI
	ADCQ R8, R8
	ADCQ R9, R9

	// reduce element(DX,CX,BX,SI,DI,R8,R9) using temp registers (R10,R11,R12,R13,R14,R15,s0-8(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,s0-8(SP))

	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI
	ADCQ DI, DI
	ADCQ R8, R8
	ADCQ R9, R9

	// reduce element(DX,CX,BX,SI,DI,R8,R9) using temp registers (s1-16(SP),s2-24(SP),s3-32(SP),s4-40(SP),s5-48(SP),s6-56(SP),s7-64(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R9,s1-16(SP),s2-24(SP),s3-32(SP),s4-40(SP),s5-48(SP),s6-56(SP),s7-64(SP))

	MOVQ DX, s1-16(SP)
	MOVQ CX, s2-24(SP)
	MOVQ BX, s3-32(SP)
	MOVQ SI, s4-40(SP)
	MOVQ DI, s5-48(SP)
	MOVQ R8, s6-56(SP)
	MOVQ R9, s7-64(SP)
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI
	ADCQ DI, DI
	ADCQ R8, R8
	ADCQ R9, R9

	// reduce element(DX,CX,BX,SI,DI,R8,R9) using temp registers (R10,R11,R12,R13,R14,R15,s0-8(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,s0-8(SP))

	ADDQ s1-16(SP), DX
	ADCQ s2-24(SP), CX
	ADCQ s3-32(SP), BX
	ADCQ s4-40(SP), SI
	ADCQ s5-48(SP), DI
	ADCQ s6-56(SP), R8
	ADCQ s7-64(SP), R9

	// reduce element(DX,CX,BX,SI,DI,R8,R9) using temp registers (R10,R11,R12,R13,R14,R15,s0-8(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,s0-8(SP))

	ADDQ 0(AX), DX
	ADCQ 8(AX), CX
	ADCQ 16(AX), BX
	ADCQ 24(AX), SI
	ADCQ 32(AX), DI
	ADCQ 40(AX), R8
	ADCQ 48(AX), R9

	// reduce element(DX,CX,BX,SI,DI,R8,R9) using temp registers (R10,R11,R12,R13,R14,R15,s0-8(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,s0-8(SP))

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	MOVQ SI, 24(AX)
	MOVQ DI, 32(AX)
	MOVQ R8, 40(AX)
	MOVQ R9, 48(AX)
	RET

// Butterfly(a, b *Element) sets a = a + b; b = a - b
TEXT ·Butterfly(SB), $8-16
	MOVQ b+8(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), SI
	MOVQ 32(AX), DI
	MOVQ 40(AX), R8
	MOVQ 48(AX), R9
	MOVQ a+0(FP), AX
	ADDQ 0(AX), DX
	ADCQ 8(AX), CX
	ADCQ 16(AX), BX
	ADCQ 24(AX), SI
	ADCQ 32(AX), DI
	ADCQ 40(AX), R8
	ADCQ 48(AX), R9
	MOVQ DX, R10
	MOVQ CX, R11
	MOVQ BX, R12
	MOVQ SI, R13
	MOVQ DI, R14
	MOVQ R8, R15
	MOVQ R9, s0-8(SP)
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), SI
	MOVQ 32(AX), DI
	MOVQ 40(AX), R8
	MOVQ 48(AX), R9
	MOVQ b+8(FP), AX
	SUBQ 0(AX), DX
	SBBQ 8(AX), CX
	SBBQ 16(AX), BX
	SBBQ 24(AX), SI
	SBBQ 32(AX), DI
	SBBQ 40(AX), R8
	SBBQ 48(AX), R9
	JCC  l3
	MOVQ $0x1c0f4efe6b9aeba5, AX
	ADDQ AX, DX
	MOVQ $0x07cc708d578478f9, AX
	ADCQ AX, CX
	MOVQ $0x7a1b80da7e024bc9, AX
	ADCQ AX, BX
	MOVQ $0x458f60da10ef1c57, AX
	ADCQ AX, SI
	MOVQ $0xbebaa6e2636b53a8, AX
	ADCQ AX, DI
	MOVQ $0xf92922e40eef678b, AX
	ADCQ AX, R8
	MOVQ $0x0000000000000003, AX
	ADCQ AX, R9

l3:
	MOVQ b+8(FP), AX
	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	MOVQ SI, 24(AX)
	MOVQ DI, 32(AX)
	MOVQ R8, 40(AX)
	MOVQ R9, 48(AX)
	MOVQ R10, DX
	MOVQ R11, CX
	MOVQ R12, BX
	MOVQ R13, SI
	MOVQ R14, DI
	MOVQ R15, R8
	MOVQ s0-8(SP), R9

	// reduce element(DX,CX,BX,SI,DI,R8,R9) using temp registers (R10,R11,R12,R13,R14,R15,s0-8(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,s0-8(SP))

	MOVQ a+0(FP), AX
	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	MOVQ SI, 24(AX)
	MOVQ DI, 32(AX)
	MOVQ R8, 40(AX)
	MOVQ R9, 48(AX)
	RET
