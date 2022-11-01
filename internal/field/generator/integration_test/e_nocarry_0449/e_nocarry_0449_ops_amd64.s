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
DATA q<>+0(SB)/8, $0x2dc691f732641493
DATA q<>+8(SB)/8, $0xc7ed9ab8cad78580
DATA q<>+16(SB)/8, $0x5f727abfa2cebfee
DATA q<>+24(SB)/8, $0x81132f2b390ba83d
DATA q<>+32(SB)/8, $0xaf7a43d236f4ec7f
DATA q<>+40(SB)/8, $0xfeabef75f23a3ee2
DATA q<>+48(SB)/8, $0x8a7dfb4227c099d4
DATA q<>+56(SB)/8, $1
GLOBL q<>(SB), (RODATA+NOPTR), $64

// qInv0 q'[0]
DATA qInv0<>(SB)/8, $0x35750028010bd665
GLOBL qInv0<>(SB), (RODATA+NOPTR), $8

#define REDUCE(ra0, ra1, ra2, ra3, ra4, ra5, ra6, ra7, rb0, rb1, rb2, rb3, rb4, rb5, rb6, rb7) \
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
	MOVQ    ra7, rb7;        \
	SBBQ    q<>+56(SB), ra7; \
	CMOVQCS rb0, ra0;        \
	CMOVQCS rb1, ra1;        \
	CMOVQCS rb2, ra2;        \
	CMOVQCS rb3, ra3;        \
	CMOVQCS rb4, ra4;        \
	CMOVQCS rb5, ra5;        \
	CMOVQCS rb6, ra6;        \
	CMOVQCS rb7, ra7;        \

TEXT ·reduce(SB), $24-8
	MOVQ res+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), SI
	MOVQ 32(AX), DI
	MOVQ 40(AX), R8
	MOVQ 48(AX), R9
	MOVQ 56(AX), R10

	// reduce element(DX,CX,BX,SI,DI,R8,R9,R10) using temp registers (R11,R12,R13,R14,R15,s0-8(SP),s1-16(SP),s2-24(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,s0-8(SP),s1-16(SP),s2-24(SP))

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	MOVQ SI, 24(AX)
	MOVQ DI, 32(AX)
	MOVQ R8, 40(AX)
	MOVQ R9, 48(AX)
	MOVQ R10, 56(AX)
	RET

// MulBy3(x *Element)
TEXT ·MulBy3(SB), $24-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), SI
	MOVQ 32(AX), DI
	MOVQ 40(AX), R8
	MOVQ 48(AX), R9
	MOVQ 56(AX), R10
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI
	ADCQ DI, DI
	ADCQ R8, R8
	ADCQ R9, R9
	ADCQ R10, R10

	// reduce element(DX,CX,BX,SI,DI,R8,R9,R10) using temp registers (R11,R12,R13,R14,R15,s0-8(SP),s1-16(SP),s2-24(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,s0-8(SP),s1-16(SP),s2-24(SP))

	ADDQ 0(AX), DX
	ADCQ 8(AX), CX
	ADCQ 16(AX), BX
	ADCQ 24(AX), SI
	ADCQ 32(AX), DI
	ADCQ 40(AX), R8
	ADCQ 48(AX), R9
	ADCQ 56(AX), R10

	// reduce element(DX,CX,BX,SI,DI,R8,R9,R10) using temp registers (R11,R12,R13,R14,R15,s0-8(SP),s1-16(SP),s2-24(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,s0-8(SP),s1-16(SP),s2-24(SP))

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	MOVQ SI, 24(AX)
	MOVQ DI, 32(AX)
	MOVQ R8, 40(AX)
	MOVQ R9, 48(AX)
	MOVQ R10, 56(AX)
	RET

// MulBy5(x *Element)
TEXT ·MulBy5(SB), $24-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), SI
	MOVQ 32(AX), DI
	MOVQ 40(AX), R8
	MOVQ 48(AX), R9
	MOVQ 56(AX), R10
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI
	ADCQ DI, DI
	ADCQ R8, R8
	ADCQ R9, R9
	ADCQ R10, R10

	// reduce element(DX,CX,BX,SI,DI,R8,R9,R10) using temp registers (R11,R12,R13,R14,R15,s0-8(SP),s1-16(SP),s2-24(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,s0-8(SP),s1-16(SP),s2-24(SP))

	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI
	ADCQ DI, DI
	ADCQ R8, R8
	ADCQ R9, R9
	ADCQ R10, R10

	// reduce element(DX,CX,BX,SI,DI,R8,R9,R10) using temp registers (R11,R12,R13,R14,R15,s0-8(SP),s1-16(SP),s2-24(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,s0-8(SP),s1-16(SP),s2-24(SP))

	ADDQ 0(AX), DX
	ADCQ 8(AX), CX
	ADCQ 16(AX), BX
	ADCQ 24(AX), SI
	ADCQ 32(AX), DI
	ADCQ 40(AX), R8
	ADCQ 48(AX), R9
	ADCQ 56(AX), R10

	// reduce element(DX,CX,BX,SI,DI,R8,R9,R10) using temp registers (R11,R12,R13,R14,R15,s0-8(SP),s1-16(SP),s2-24(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,s0-8(SP),s1-16(SP),s2-24(SP))

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	MOVQ SI, 24(AX)
	MOVQ DI, 32(AX)
	MOVQ R8, 40(AX)
	MOVQ R9, 48(AX)
	MOVQ R10, 56(AX)
	RET

// MulBy13(x *Element)
TEXT ·MulBy13(SB), $88-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), SI
	MOVQ 32(AX), DI
	MOVQ 40(AX), R8
	MOVQ 48(AX), R9
	MOVQ 56(AX), R10
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI
	ADCQ DI, DI
	ADCQ R8, R8
	ADCQ R9, R9
	ADCQ R10, R10

	// reduce element(DX,CX,BX,SI,DI,R8,R9,R10) using temp registers (R11,R12,R13,R14,R15,s0-8(SP),s1-16(SP),s2-24(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,s0-8(SP),s1-16(SP),s2-24(SP))

	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI
	ADCQ DI, DI
	ADCQ R8, R8
	ADCQ R9, R9
	ADCQ R10, R10

	// reduce element(DX,CX,BX,SI,DI,R8,R9,R10) using temp registers (s3-32(SP),s4-40(SP),s5-48(SP),s6-56(SP),s7-64(SP),s8-72(SP),s9-80(SP),s10-88(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,s3-32(SP),s4-40(SP),s5-48(SP),s6-56(SP),s7-64(SP),s8-72(SP),s9-80(SP),s10-88(SP))

	MOVQ DX, s3-32(SP)
	MOVQ CX, s4-40(SP)
	MOVQ BX, s5-48(SP)
	MOVQ SI, s6-56(SP)
	MOVQ DI, s7-64(SP)
	MOVQ R8, s8-72(SP)
	MOVQ R9, s9-80(SP)
	MOVQ R10, s10-88(SP)
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI
	ADCQ DI, DI
	ADCQ R8, R8
	ADCQ R9, R9
	ADCQ R10, R10

	// reduce element(DX,CX,BX,SI,DI,R8,R9,R10) using temp registers (R11,R12,R13,R14,R15,s0-8(SP),s1-16(SP),s2-24(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,s0-8(SP),s1-16(SP),s2-24(SP))

	ADDQ s3-32(SP), DX
	ADCQ s4-40(SP), CX
	ADCQ s5-48(SP), BX
	ADCQ s6-56(SP), SI
	ADCQ s7-64(SP), DI
	ADCQ s8-72(SP), R8
	ADCQ s9-80(SP), R9
	ADCQ s10-88(SP), R10

	// reduce element(DX,CX,BX,SI,DI,R8,R9,R10) using temp registers (R11,R12,R13,R14,R15,s0-8(SP),s1-16(SP),s2-24(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,s0-8(SP),s1-16(SP),s2-24(SP))

	ADDQ 0(AX), DX
	ADCQ 8(AX), CX
	ADCQ 16(AX), BX
	ADCQ 24(AX), SI
	ADCQ 32(AX), DI
	ADCQ 40(AX), R8
	ADCQ 48(AX), R9
	ADCQ 56(AX), R10

	// reduce element(DX,CX,BX,SI,DI,R8,R9,R10) using temp registers (R11,R12,R13,R14,R15,s0-8(SP),s1-16(SP),s2-24(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,s0-8(SP),s1-16(SP),s2-24(SP))

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	MOVQ SI, 24(AX)
	MOVQ DI, 32(AX)
	MOVQ R8, 40(AX)
	MOVQ R9, 48(AX)
	MOVQ R10, 56(AX)
	RET

// Butterfly(a, b *Element) sets a = a + b; b = a - b
TEXT ·Butterfly(SB), $24-16
	MOVQ b+8(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), SI
	MOVQ 32(AX), DI
	MOVQ 40(AX), R8
	MOVQ 48(AX), R9
	MOVQ 56(AX), R10
	MOVQ a+0(FP), AX
	ADDQ 0(AX), DX
	ADCQ 8(AX), CX
	ADCQ 16(AX), BX
	ADCQ 24(AX), SI
	ADCQ 32(AX), DI
	ADCQ 40(AX), R8
	ADCQ 48(AX), R9
	ADCQ 56(AX), R10
	MOVQ DX, R11
	MOVQ CX, R12
	MOVQ BX, R13
	MOVQ SI, R14
	MOVQ DI, R15
	MOVQ R8, s0-8(SP)
	MOVQ R9, s1-16(SP)
	MOVQ R10, s2-24(SP)
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), SI
	MOVQ 32(AX), DI
	MOVQ 40(AX), R8
	MOVQ 48(AX), R9
	MOVQ 56(AX), R10
	MOVQ b+8(FP), AX
	SUBQ 0(AX), DX
	SBBQ 8(AX), CX
	SBBQ 16(AX), BX
	SBBQ 24(AX), SI
	SBBQ 32(AX), DI
	SBBQ 40(AX), R8
	SBBQ 48(AX), R9
	SBBQ 56(AX), R10
	JCC  l1
	MOVQ $0x2dc691f732641493, AX
	ADDQ AX, DX
	MOVQ $0xc7ed9ab8cad78580, AX
	ADCQ AX, CX
	MOVQ $0x5f727abfa2cebfee, AX
	ADCQ AX, BX
	MOVQ $0x81132f2b390ba83d, AX
	ADCQ AX, SI
	MOVQ $0xaf7a43d236f4ec7f, AX
	ADCQ AX, DI
	MOVQ $0xfeabef75f23a3ee2, AX
	ADCQ AX, R8
	MOVQ $0x8a7dfb4227c099d4, AX
	ADCQ AX, R9
	MOVQ $1, AX
	ADCQ AX, R10

l1:
	MOVQ b+8(FP), AX
	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	MOVQ SI, 24(AX)
	MOVQ DI, 32(AX)
	MOVQ R8, 40(AX)
	MOVQ R9, 48(AX)
	MOVQ R10, 56(AX)
	MOVQ R11, DX
	MOVQ R12, CX
	MOVQ R13, BX
	MOVQ R14, SI
	MOVQ R15, DI
	MOVQ s0-8(SP), R8
	MOVQ s1-16(SP), R9
	MOVQ s2-24(SP), R10

	// reduce element(DX,CX,BX,SI,DI,R8,R9,R10) using temp registers (R11,R12,R13,R14,R15,s0-8(SP),s1-16(SP),s2-24(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,s0-8(SP),s1-16(SP),s2-24(SP))

	MOVQ a+0(FP), AX
	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	MOVQ SI, 24(AX)
	MOVQ DI, 32(AX)
	MOVQ R8, 40(AX)
	MOVQ R9, 48(AX)
	MOVQ R10, 56(AX)
	RET
