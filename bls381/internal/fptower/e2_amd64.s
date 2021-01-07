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

#define REDUCE_AND_MOVE(ra0, ra1, ra2, ra3, ra4, ra5, rb0, rb1, rb2, rb3, rb4, rb5, res0, res1, res2, res3, res4, res5) \
	MOVQ    ra0, rb0;        \
	MOVQ    ra1, rb1;        \
	MOVQ    ra2, rb2;        \
	MOVQ    ra3, rb3;        \
	MOVQ    ra4, rb4;        \
	MOVQ    ra5, rb5;        \
	SUBQ    q<>(SB), rb0;    \
	SBBQ    q<>+8(SB), rb1;  \
	SBBQ    q<>+16(SB), rb2; \
	SBBQ    q<>+24(SB), rb3; \
	SBBQ    q<>+32(SB), rb4; \
	SBBQ    q<>+40(SB), rb5; \
	CMOVQCC rb0, ra0;        \
	CMOVQCC rb1, ra1;        \
	CMOVQCC rb2, ra2;        \
	CMOVQCC rb3, ra3;        \
	CMOVQCC rb4, ra4;        \
	CMOVQCC rb5, ra5;        \
	MOVQ    ra0, res0;       \
	MOVQ    ra1, res1;       \
	MOVQ    ra2, res2;       \
	MOVQ    ra3, res3;       \
	MOVQ    ra4, res4;       \
	MOVQ    ra5, res5;       \

#define REDUCE(ra0, ra1, ra2, ra3, ra4, ra5, rb0, rb1, rb2, rb3, rb4, rb5) \
	MOVQ    ra0, rb0;        \
	MOVQ    ra1, rb1;        \
	MOVQ    ra2, rb2;        \
	MOVQ    ra3, rb3;        \
	MOVQ    ra4, rb4;        \
	MOVQ    ra5, rb5;        \
	SUBQ    q<>(SB), rb0;    \
	SBBQ    q<>+8(SB), rb1;  \
	SBBQ    q<>+16(SB), rb2; \
	SBBQ    q<>+24(SB), rb3; \
	SBBQ    q<>+32(SB), rb4; \
	SBBQ    q<>+40(SB), rb5; \
	CMOVQCC rb0, ra0;        \
	CMOVQCC rb1, ra1;        \
	CMOVQCC rb2, ra2;        \
	CMOVQCC rb3, ra3;        \
	CMOVQCC rb4, ra4;        \
	CMOVQCC rb5, ra5;        \

TEXT ·addE2(SB), NOSPLIT, $0-24
	MOVQ x+8(FP), AX
	MOVQ 0(AX), BX
	MOVQ 8(AX), BP
	MOVQ 16(AX), SI
	MOVQ 24(AX), DI
	MOVQ 32(AX), R8
	MOVQ 40(AX), R9
	MOVQ y+16(FP), DX
	ADDQ 0(DX), BX
	ADCQ 8(DX), BP
	ADCQ 16(DX), SI
	ADCQ 24(DX), DI
	ADCQ 32(DX), R8
	ADCQ 40(DX), R9
	MOVQ res+0(FP), CX

	// reduce element(BX,BP,SI,DI,R8,R9) using temp registers (R10,R11,R12,R13,R14,R15)
	// stores in (0(CX),8(CX),16(CX),24(CX),32(CX),40(CX))
	REDUCE_AND_MOVE(BX,BP,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,0(CX),8(CX),16(CX),24(CX),32(CX),40(CX))

	MOVQ 48(AX), BX
	MOVQ 56(AX), BP
	MOVQ 64(AX), SI
	MOVQ 72(AX), DI
	MOVQ 80(AX), R8
	MOVQ 88(AX), R9
	ADDQ 48(DX), BX
	ADCQ 56(DX), BP
	ADCQ 64(DX), SI
	ADCQ 72(DX), DI
	ADCQ 80(DX), R8
	ADCQ 88(DX), R9

	// reduce element(BX,BP,SI,DI,R8,R9) using temp registers (R10,R11,R12,R13,R14,R15)
	// stores in (48(CX),56(CX),64(CX),72(CX),80(CX),88(CX))
	REDUCE_AND_MOVE(BX,BP,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,48(CX),56(CX),64(CX),72(CX),80(CX),88(CX))

	RET

TEXT ·doubleE2(SB), NOSPLIT, $0-16
	MOVQ res+0(FP), DX
	MOVQ x+8(FP), AX
	MOVQ 0(AX), CX
	MOVQ 8(AX), BX
	MOVQ 16(AX), BP
	MOVQ 24(AX), SI
	MOVQ 32(AX), DI
	MOVQ 40(AX), R8
	ADDQ CX, CX
	ADCQ BX, BX
	ADCQ BP, BP
	ADCQ SI, SI
	ADCQ DI, DI
	ADCQ R8, R8

	// reduce element(CX,BX,BP,SI,DI,R8) using temp registers (R9,R10,R11,R12,R13,R14)
	// stores in (0(DX),8(DX),16(DX),24(DX),32(DX),40(DX))
	REDUCE_AND_MOVE(CX,BX,BP,SI,DI,R8,R9,R10,R11,R12,R13,R14,0(DX),8(DX),16(DX),24(DX),32(DX),40(DX))

	MOVQ 48(AX), CX
	MOVQ 56(AX), BX
	MOVQ 64(AX), BP
	MOVQ 72(AX), SI
	MOVQ 80(AX), DI
	MOVQ 88(AX), R8
	ADDQ CX, CX
	ADCQ BX, BX
	ADCQ BP, BP
	ADCQ SI, SI
	ADCQ DI, DI
	ADCQ R8, R8

	// reduce element(CX,BX,BP,SI,DI,R8) using temp registers (R15,R9,R10,R11,R12,R13)
	// stores in (48(DX),56(DX),64(DX),72(DX),80(DX),88(DX))
	REDUCE_AND_MOVE(CX,BX,BP,SI,DI,R8,R15,R9,R10,R11,R12,R13,48(DX),56(DX),64(DX),72(DX),80(DX),88(DX))

	RET

TEXT ·subE2(SB), NOSPLIT, $0-24
	MOVQ    x+8(FP), DI
	MOVQ    y+16(FP), R8
	MOVQ    0(DI), AX
	MOVQ    8(DI), DX
	MOVQ    16(DI), CX
	MOVQ    24(DI), BX
	MOVQ    32(DI), BP
	MOVQ    40(DI), SI
	SUBQ    0(R8), AX
	SBBQ    8(R8), DX
	SBBQ    16(R8), CX
	SBBQ    24(R8), BX
	SBBQ    32(R8), BP
	SBBQ    40(R8), SI
	MOVQ    $0xb9feffffffffaaab, R9
	MOVQ    $0x1eabfffeb153ffff, R10
	MOVQ    $0x6730d2a0f6b0f624, R11
	MOVQ    $0x64774b84f38512bf, R12
	MOVQ    $0x4b1ba7b6434bacd7, R13
	MOVQ    $0x1a0111ea397fe69a, R14
	MOVQ    $0, R15
	CMOVQCC R15, R9
	CMOVQCC R15, R10
	CMOVQCC R15, R11
	CMOVQCC R15, R12
	CMOVQCC R15, R13
	CMOVQCC R15, R14
	ADDQ    R9, AX
	ADCQ    R10, DX
	ADCQ    R11, CX
	ADCQ    R12, BX
	ADCQ    R13, BP
	ADCQ    R14, SI
	MOVQ    res+0(FP), R15
	MOVQ    AX, 0(R15)
	MOVQ    DX, 8(R15)
	MOVQ    CX, 16(R15)
	MOVQ    BX, 24(R15)
	MOVQ    BP, 32(R15)
	MOVQ    SI, 40(R15)
	MOVQ    48(DI), AX
	MOVQ    56(DI), DX
	MOVQ    64(DI), CX
	MOVQ    72(DI), BX
	MOVQ    80(DI), BP
	MOVQ    88(DI), SI
	SUBQ    48(R8), AX
	SBBQ    56(R8), DX
	SBBQ    64(R8), CX
	SBBQ    72(R8), BX
	SBBQ    80(R8), BP
	SBBQ    88(R8), SI
	MOVQ    $0xb9feffffffffaaab, R9
	MOVQ    $0x1eabfffeb153ffff, R10
	MOVQ    $0x6730d2a0f6b0f624, R11
	MOVQ    $0x64774b84f38512bf, R12
	MOVQ    $0x4b1ba7b6434bacd7, R13
	MOVQ    $0x1a0111ea397fe69a, R14
	MOVQ    $0, R15
	CMOVQCC R15, R9
	CMOVQCC R15, R10
	CMOVQCC R15, R11
	CMOVQCC R15, R12
	CMOVQCC R15, R13
	CMOVQCC R15, R14
	ADDQ    R9, AX
	ADCQ    R10, DX
	ADCQ    R11, CX
	ADCQ    R12, BX
	ADCQ    R13, BP
	ADCQ    R14, SI
	MOVQ    res+0(FP), DI
	MOVQ    AX, 48(DI)
	MOVQ    DX, 56(DI)
	MOVQ    CX, 64(DI)
	MOVQ    BX, 72(DI)
	MOVQ    BP, 80(DI)
	MOVQ    SI, 88(DI)
	RET

TEXT ·negE2(SB), NOSPLIT, $0-16
	MOVQ  res+0(FP), DX
	MOVQ  x+8(FP), AX
	MOVQ  0(AX), BX
	MOVQ  8(AX), BP
	MOVQ  16(AX), SI
	MOVQ  24(AX), DI
	MOVQ  32(AX), R8
	MOVQ  40(AX), R9
	MOVQ  BX, AX
	ORQ   BP, AX
	ORQ   SI, AX
	ORQ   DI, AX
	ORQ   R8, AX
	ORQ   R9, AX
	TESTQ AX, AX
	JNE   l1
	MOVQ  AX, 48(DX)
	MOVQ  AX, 56(DX)
	MOVQ  AX, 64(DX)
	MOVQ  AX, 72(DX)
	MOVQ  AX, 80(DX)
	MOVQ  AX, 88(DX)
	JMP   l3

l1:
	MOVQ $0xb9feffffffffaaab, CX
	SUBQ BX, CX
	MOVQ CX, 0(DX)
	MOVQ $0x1eabfffeb153ffff, CX
	SBBQ BP, CX
	MOVQ CX, 8(DX)
	MOVQ $0x6730d2a0f6b0f624, CX
	SBBQ SI, CX
	MOVQ CX, 16(DX)
	MOVQ $0x64774b84f38512bf, CX
	SBBQ DI, CX
	MOVQ CX, 24(DX)
	MOVQ $0x4b1ba7b6434bacd7, CX
	SBBQ R8, CX
	MOVQ CX, 32(DX)
	MOVQ $0x1a0111ea397fe69a, CX
	SBBQ R9, CX
	MOVQ CX, 40(DX)

l3:
	MOVQ  x+8(FP), AX
	MOVQ  48(AX), BX
	MOVQ  56(AX), BP
	MOVQ  64(AX), SI
	MOVQ  72(AX), DI
	MOVQ  80(AX), R8
	MOVQ  88(AX), R9
	MOVQ  BX, AX
	ORQ   BP, AX
	ORQ   SI, AX
	ORQ   DI, AX
	ORQ   R8, AX
	ORQ   R9, AX
	TESTQ AX, AX
	JNE   l2
	MOVQ  AX, 48(DX)
	MOVQ  AX, 56(DX)
	MOVQ  AX, 64(DX)
	MOVQ  AX, 72(DX)
	MOVQ  AX, 80(DX)
	MOVQ  AX, 88(DX)
	RET

l2:
	MOVQ $0xb9feffffffffaaab, CX
	SUBQ BX, CX
	MOVQ CX, 48(DX)
	MOVQ $0x1eabfffeb153ffff, CX
	SBBQ BP, CX
	MOVQ CX, 56(DX)
	MOVQ $0x6730d2a0f6b0f624, CX
	SBBQ SI, CX
	MOVQ CX, 64(DX)
	MOVQ $0x64774b84f38512bf, CX
	SBBQ DI, CX
	MOVQ CX, 72(DX)
	MOVQ $0x4b1ba7b6434bacd7, CX
	SBBQ R8, CX
	MOVQ CX, 80(DX)
	MOVQ $0x1a0111ea397fe69a, CX
	SBBQ R9, CX
	MOVQ CX, 88(DX)
	RET

TEXT ·mulNonResE2(SB), NOSPLIT, $0-16
	MOVQ    x+8(FP), DI
	MOVQ    0(DI), AX
	MOVQ    8(DI), DX
	MOVQ    16(DI), CX
	MOVQ    24(DI), BX
	MOVQ    32(DI), BP
	MOVQ    40(DI), SI
	SUBQ    48(DI), AX
	SBBQ    56(DI), DX
	SBBQ    64(DI), CX
	SBBQ    72(DI), BX
	SBBQ    80(DI), BP
	SBBQ    88(DI), SI
	MOVQ    $0xb9feffffffffaaab, R8
	MOVQ    $0x1eabfffeb153ffff, R9
	MOVQ    $0x6730d2a0f6b0f624, R10
	MOVQ    $0x64774b84f38512bf, R11
	MOVQ    $0x4b1ba7b6434bacd7, R12
	MOVQ    $0x1a0111ea397fe69a, R13
	MOVQ    $0, R14
	CMOVQCC R14, R8
	CMOVQCC R14, R9
	CMOVQCC R14, R10
	CMOVQCC R14, R11
	CMOVQCC R14, R12
	CMOVQCC R14, R13
	ADDQ    R8, AX
	ADCQ    R9, DX
	ADCQ    R10, CX
	ADCQ    R11, BX
	ADCQ    R12, BP
	ADCQ    R13, SI
	MOVQ    48(DI), R15
	MOVQ    56(DI), R14
	MOVQ    64(DI), R8
	MOVQ    72(DI), R9
	MOVQ    80(DI), R10
	MOVQ    88(DI), R11
	ADDQ    0(DI), R15
	ADCQ    8(DI), R14
	ADCQ    16(DI), R8
	ADCQ    24(DI), R9
	ADCQ    32(DI), R10
	ADCQ    40(DI), R11
	MOVQ    res+0(FP), DI
	MOVQ    AX, 0(DI)
	MOVQ    DX, 8(DI)
	MOVQ    CX, 16(DI)
	MOVQ    BX, 24(DI)
	MOVQ    BP, 32(DI)
	MOVQ    SI, 40(DI)

	// reduce element(R15,R14,R8,R9,R10,R11) using temp registers (R12,R13,AX,DX,CX,BX)
	// stores in (48(DI),56(DI),64(DI),72(DI),80(DI),88(DI))
	REDUCE_AND_MOVE(R15,R14,R8,R9,R10,R11,R12,R13,AX,DX,CX,BX,48(DI),56(DI),64(DI),72(DI),80(DI),88(DI))

	RET

TEXT ·squareAdxE2(SB), $56-16
	NO_LOCAL_POINTERS
	CMPB    ·supportAdx(SB), $1
	JNE     l4
	MOVQ    x+8(FP), DX
	MOVQ    0(DX), R14
	MOVQ    8(DX), R15
	MOVQ    16(DX), CX
	MOVQ    24(DX), BX
	MOVQ    32(DX), BP
	MOVQ    40(DX), SI
	SUBQ    48(DX), R14
	SBBQ    56(DX), R15
	SBBQ    64(DX), CX
	SBBQ    72(DX), BX
	SBBQ    80(DX), BP
	SBBQ    88(DX), SI
	MOVQ    $0xb9feffffffffaaab, DI
	MOVQ    $0x1eabfffeb153ffff, R8
	MOVQ    $0x6730d2a0f6b0f624, R9
	MOVQ    $0x64774b84f38512bf, R10
	MOVQ    $0x4b1ba7b6434bacd7, R11
	MOVQ    $0x1a0111ea397fe69a, R12
	MOVQ    $0, R13
	CMOVQCC R13, DI
	CMOVQCC R13, R8
	CMOVQCC R13, R9
	CMOVQCC R13, R10
	CMOVQCC R13, R11
	CMOVQCC R13, R12
	ADDQ    DI, R14
	ADCQ    R8, R15
	ADCQ    R9, CX
	ADCQ    R10, BX
	ADCQ    R11, BP
	ADCQ    R12, SI
	MOVQ    R14, -16(SP)
	MOVQ    R15, -24(SP)
	MOVQ    CX, -32(SP)
	MOVQ    BX, -40(SP)
	MOVQ    BP, -48(SP)
	MOVQ    SI, -56(SP)
	MOVQ    0(DX), R14
	MOVQ    8(DX), R15
	MOVQ    16(DX), CX
	MOVQ    24(DX), BX
	MOVQ    32(DX), BP
	MOVQ    40(DX), SI
	MOVQ    48(DX), R13
	MOVQ    56(DX), DI
	MOVQ    64(DX), R8
	MOVQ    72(DX), R9
	MOVQ    80(DX), R10
	MOVQ    88(DX), R11
	ADDQ    R13, R14
	ADCQ    DI, R15
	ADCQ    R8, CX
	ADCQ    R9, BX
	ADCQ    R10, BP
	ADCQ    R11, SI

	// reduce element(R14,R15,CX,BX,BP,SI) using temp registers (R12,R13,DI,R8,R9,R10)
	REDUCE(R14,R15,CX,BX,BP,SI,R12,R13,DI,R8,R9,R10)

	// t[0] = R11
	// t[1] = R12
	// t[2] = R13
	// t[3] = DI
	// t[4] = R8
	// t[5] = R9

	// clear the flags
	XORQ AX, AX
	MOVQ -16(SP), DX

	// (A,t[0])  := t[0] + x[0]*y[0] + A
	MULXQ R14, R11, R12

	// (A,t[1])  := t[1] + x[1]*y[0] + A
	MULXQ R15, AX, R13
	ADOXQ AX, R12

	// (A,t[2])  := t[2] + x[2]*y[0] + A
	MULXQ CX, AX, DI
	ADOXQ AX, R13

	// (A,t[3])  := t[3] + x[3]*y[0] + A
	MULXQ BX, AX, R8
	ADOXQ AX, DI

	// (A,t[4])  := t[4] + x[4]*y[0] + A
	MULXQ BP, AX, R9
	ADOXQ AX, R8

	// (A,t[5])  := t[5] + x[5]*y[0] + A
	MULXQ SI, AX, R10
	ADOXQ AX, R9

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADOXQ AX, R10

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R11, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	PUSHQ R10
	MULXQ q<>+0(SB), AX, R10
	ADCXQ R11, AX
	MOVQ  R10, R11
	POPQ  R10

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R12, R11
	MULXQ q<>+8(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R13, R12
	MULXQ q<>+16(SB), AX, R13
	ADOXQ AX, R12

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ DI, R13
	MULXQ q<>+24(SB), AX, DI
	ADOXQ AX, R13

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ R8, DI
	MULXQ q<>+32(SB), AX, R8
	ADOXQ AX, DI

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ R9, R8
	MULXQ q<>+40(SB), AX, R9
	ADOXQ AX, R8

	// t[5] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R9
	ADOXQ R10, R9

	// clear the flags
	XORQ AX, AX
	MOVQ -24(SP), DX

	// (A,t[0])  := t[0] + x[0]*y[1] + A
	MULXQ R14, AX, R10
	ADOXQ AX, R11

	// (A,t[1])  := t[1] + x[1]*y[1] + A
	ADCXQ R10, R12
	MULXQ R15, AX, R10
	ADOXQ AX, R12

	// (A,t[2])  := t[2] + x[2]*y[1] + A
	ADCXQ R10, R13
	MULXQ CX, AX, R10
	ADOXQ AX, R13

	// (A,t[3])  := t[3] + x[3]*y[1] + A
	ADCXQ R10, DI
	MULXQ BX, AX, R10
	ADOXQ AX, DI

	// (A,t[4])  := t[4] + x[4]*y[1] + A
	ADCXQ R10, R8
	MULXQ BP, AX, R10
	ADOXQ AX, R8

	// (A,t[5])  := t[5] + x[5]*y[1] + A
	ADCXQ R10, R9
	MULXQ SI, AX, R10
	ADOXQ AX, R9

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R10
	ADOXQ AX, R10

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R11, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	PUSHQ R10
	MULXQ q<>+0(SB), AX, R10
	ADCXQ R11, AX
	MOVQ  R10, R11
	POPQ  R10

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R12, R11
	MULXQ q<>+8(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R13, R12
	MULXQ q<>+16(SB), AX, R13
	ADOXQ AX, R12

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ DI, R13
	MULXQ q<>+24(SB), AX, DI
	ADOXQ AX, R13

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ R8, DI
	MULXQ q<>+32(SB), AX, R8
	ADOXQ AX, DI

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ R9, R8
	MULXQ q<>+40(SB), AX, R9
	ADOXQ AX, R8

	// t[5] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R9
	ADOXQ R10, R9

	// clear the flags
	XORQ AX, AX
	MOVQ -32(SP), DX

	// (A,t[0])  := t[0] + x[0]*y[2] + A
	MULXQ R14, AX, R10
	ADOXQ AX, R11

	// (A,t[1])  := t[1] + x[1]*y[2] + A
	ADCXQ R10, R12
	MULXQ R15, AX, R10
	ADOXQ AX, R12

	// (A,t[2])  := t[2] + x[2]*y[2] + A
	ADCXQ R10, R13
	MULXQ CX, AX, R10
	ADOXQ AX, R13

	// (A,t[3])  := t[3] + x[3]*y[2] + A
	ADCXQ R10, DI
	MULXQ BX, AX, R10
	ADOXQ AX, DI

	// (A,t[4])  := t[4] + x[4]*y[2] + A
	ADCXQ R10, R8
	MULXQ BP, AX, R10
	ADOXQ AX, R8

	// (A,t[5])  := t[5] + x[5]*y[2] + A
	ADCXQ R10, R9
	MULXQ SI, AX, R10
	ADOXQ AX, R9

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R10
	ADOXQ AX, R10

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R11, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	PUSHQ R10
	MULXQ q<>+0(SB), AX, R10
	ADCXQ R11, AX
	MOVQ  R10, R11
	POPQ  R10

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R12, R11
	MULXQ q<>+8(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R13, R12
	MULXQ q<>+16(SB), AX, R13
	ADOXQ AX, R12

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ DI, R13
	MULXQ q<>+24(SB), AX, DI
	ADOXQ AX, R13

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ R8, DI
	MULXQ q<>+32(SB), AX, R8
	ADOXQ AX, DI

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ R9, R8
	MULXQ q<>+40(SB), AX, R9
	ADOXQ AX, R8

	// t[5] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R9
	ADOXQ R10, R9

	// clear the flags
	XORQ AX, AX
	MOVQ -40(SP), DX

	// (A,t[0])  := t[0] + x[0]*y[3] + A
	MULXQ R14, AX, R10
	ADOXQ AX, R11

	// (A,t[1])  := t[1] + x[1]*y[3] + A
	ADCXQ R10, R12
	MULXQ R15, AX, R10
	ADOXQ AX, R12

	// (A,t[2])  := t[2] + x[2]*y[3] + A
	ADCXQ R10, R13
	MULXQ CX, AX, R10
	ADOXQ AX, R13

	// (A,t[3])  := t[3] + x[3]*y[3] + A
	ADCXQ R10, DI
	MULXQ BX, AX, R10
	ADOXQ AX, DI

	// (A,t[4])  := t[4] + x[4]*y[3] + A
	ADCXQ R10, R8
	MULXQ BP, AX, R10
	ADOXQ AX, R8

	// (A,t[5])  := t[5] + x[5]*y[3] + A
	ADCXQ R10, R9
	MULXQ SI, AX, R10
	ADOXQ AX, R9

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R10
	ADOXQ AX, R10

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R11, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	PUSHQ R10
	MULXQ q<>+0(SB), AX, R10
	ADCXQ R11, AX
	MOVQ  R10, R11
	POPQ  R10

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R12, R11
	MULXQ q<>+8(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R13, R12
	MULXQ q<>+16(SB), AX, R13
	ADOXQ AX, R12

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ DI, R13
	MULXQ q<>+24(SB), AX, DI
	ADOXQ AX, R13

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ R8, DI
	MULXQ q<>+32(SB), AX, R8
	ADOXQ AX, DI

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ R9, R8
	MULXQ q<>+40(SB), AX, R9
	ADOXQ AX, R8

	// t[5] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R9
	ADOXQ R10, R9

	// clear the flags
	XORQ AX, AX
	MOVQ -48(SP), DX

	// (A,t[0])  := t[0] + x[0]*y[4] + A
	MULXQ R14, AX, R10
	ADOXQ AX, R11

	// (A,t[1])  := t[1] + x[1]*y[4] + A
	ADCXQ R10, R12
	MULXQ R15, AX, R10
	ADOXQ AX, R12

	// (A,t[2])  := t[2] + x[2]*y[4] + A
	ADCXQ R10, R13
	MULXQ CX, AX, R10
	ADOXQ AX, R13

	// (A,t[3])  := t[3] + x[3]*y[4] + A
	ADCXQ R10, DI
	MULXQ BX, AX, R10
	ADOXQ AX, DI

	// (A,t[4])  := t[4] + x[4]*y[4] + A
	ADCXQ R10, R8
	MULXQ BP, AX, R10
	ADOXQ AX, R8

	// (A,t[5])  := t[5] + x[5]*y[4] + A
	ADCXQ R10, R9
	MULXQ SI, AX, R10
	ADOXQ AX, R9

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R10
	ADOXQ AX, R10

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R11, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	PUSHQ R10
	MULXQ q<>+0(SB), AX, R10
	ADCXQ R11, AX
	MOVQ  R10, R11
	POPQ  R10

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R12, R11
	MULXQ q<>+8(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R13, R12
	MULXQ q<>+16(SB), AX, R13
	ADOXQ AX, R12

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ DI, R13
	MULXQ q<>+24(SB), AX, DI
	ADOXQ AX, R13

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ R8, DI
	MULXQ q<>+32(SB), AX, R8
	ADOXQ AX, DI

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ R9, R8
	MULXQ q<>+40(SB), AX, R9
	ADOXQ AX, R8

	// t[5] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R9
	ADOXQ R10, R9

	// clear the flags
	XORQ AX, AX
	MOVQ -56(SP), DX

	// (A,t[0])  := t[0] + x[0]*y[5] + A
	MULXQ R14, AX, R10
	ADOXQ AX, R11

	// (A,t[1])  := t[1] + x[1]*y[5] + A
	ADCXQ R10, R12
	MULXQ R15, AX, R10
	ADOXQ AX, R12

	// (A,t[2])  := t[2] + x[2]*y[5] + A
	ADCXQ R10, R13
	MULXQ CX, AX, R10
	ADOXQ AX, R13

	// (A,t[3])  := t[3] + x[3]*y[5] + A
	ADCXQ R10, DI
	MULXQ BX, AX, R10
	ADOXQ AX, DI

	// (A,t[4])  := t[4] + x[4]*y[5] + A
	ADCXQ R10, R8
	MULXQ BP, AX, R10
	ADOXQ AX, R8

	// (A,t[5])  := t[5] + x[5]*y[5] + A
	ADCXQ R10, R9
	MULXQ SI, AX, R10
	ADOXQ AX, R9

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R10
	ADOXQ AX, R10

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R11, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	PUSHQ R10
	MULXQ q<>+0(SB), AX, R10
	ADCXQ R11, AX
	MOVQ  R10, R11
	POPQ  R10

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R12, R11
	MULXQ q<>+8(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R13, R12
	MULXQ q<>+16(SB), AX, R13
	ADOXQ AX, R12

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ DI, R13
	MULXQ q<>+24(SB), AX, DI
	ADOXQ AX, R13

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ R8, DI
	MULXQ q<>+32(SB), AX, R8
	ADOXQ AX, DI

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ R9, R8
	MULXQ q<>+40(SB), AX, R9
	ADOXQ AX, R8

	// t[5] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R9
	ADOXQ R10, R9

	// reduce element(R11,R12,R13,DI,R8,R9) using temp registers (R10,R14,R15,CX,BX,BP)
	REDUCE(R11,R12,R13,DI,R8,R9,R10,R14,R15,CX,BX,BP)

	MOVQ res+0(FP), DX
	MOVQ x+8(FP), SI
	MOVQ 0(SI), R10
	MOVQ 8(SI), R14
	MOVQ 16(SI), R15
	MOVQ 24(SI), CX
	MOVQ 32(SI), BX
	MOVQ 40(SI), BP
	MOVQ R11, 0(DX)
	MOVQ R12, 8(DX)
	MOVQ R13, 16(DX)
	MOVQ DI, 24(DX)
	MOVQ R8, 32(DX)
	MOVQ R9, 40(DX)

	// t[0] = SI
	// t[1] = R11
	// t[2] = R12
	// t[3] = R13
	// t[4] = DI
	// t[5] = R8

	// clear the flags
	XORQ AX, AX
	MOVQ x+8(FP), DX
	MOVQ 48(DX), DX

	// (A,t[0])  := t[0] + x[0]*y[0] + A
	MULXQ R10, SI, R11

	// (A,t[1])  := t[1] + x[1]*y[0] + A
	MULXQ R14, AX, R12
	ADOXQ AX, R11

	// (A,t[2])  := t[2] + x[2]*y[0] + A
	MULXQ R15, AX, R13
	ADOXQ AX, R12

	// (A,t[3])  := t[3] + x[3]*y[0] + A
	MULXQ CX, AX, DI
	ADOXQ AX, R13

	// (A,t[4])  := t[4] + x[4]*y[0] + A
	MULXQ BX, AX, R8
	ADOXQ AX, DI

	// (A,t[5])  := t[5] + x[5]*y[0] + A
	MULXQ BP, AX, R9
	ADOXQ AX, R8

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADOXQ AX, R9

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ SI, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	PUSHQ R9
	MULXQ q<>+0(SB), AX, R9
	ADCXQ SI, AX
	MOVQ  R9, SI
	POPQ  R9

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R11, SI
	MULXQ q<>+8(SB), AX, R11
	ADOXQ AX, SI

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R12, R11
	MULXQ q<>+16(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ R13, R12
	MULXQ q<>+24(SB), AX, R13
	ADOXQ AX, R12

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ DI, R13
	MULXQ q<>+32(SB), AX, DI
	ADOXQ AX, R13

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ R8, DI
	MULXQ q<>+40(SB), AX, R8
	ADOXQ AX, DI

	// t[5] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R8
	ADOXQ R9, R8

	// clear the flags
	XORQ AX, AX
	MOVQ x+8(FP), DX
	MOVQ 56(DX), DX

	// (A,t[0])  := t[0] + x[0]*y[1] + A
	MULXQ R10, AX, R9
	ADOXQ AX, SI

	// (A,t[1])  := t[1] + x[1]*y[1] + A
	ADCXQ R9, R11
	MULXQ R14, AX, R9
	ADOXQ AX, R11

	// (A,t[2])  := t[2] + x[2]*y[1] + A
	ADCXQ R9, R12
	MULXQ R15, AX, R9
	ADOXQ AX, R12

	// (A,t[3])  := t[3] + x[3]*y[1] + A
	ADCXQ R9, R13
	MULXQ CX, AX, R9
	ADOXQ AX, R13

	// (A,t[4])  := t[4] + x[4]*y[1] + A
	ADCXQ R9, DI
	MULXQ BX, AX, R9
	ADOXQ AX, DI

	// (A,t[5])  := t[5] + x[5]*y[1] + A
	ADCXQ R9, R8
	MULXQ BP, AX, R9
	ADOXQ AX, R8

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R9
	ADOXQ AX, R9

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ SI, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	PUSHQ R9
	MULXQ q<>+0(SB), AX, R9
	ADCXQ SI, AX
	MOVQ  R9, SI
	POPQ  R9

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R11, SI
	MULXQ q<>+8(SB), AX, R11
	ADOXQ AX, SI

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R12, R11
	MULXQ q<>+16(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ R13, R12
	MULXQ q<>+24(SB), AX, R13
	ADOXQ AX, R12

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ DI, R13
	MULXQ q<>+32(SB), AX, DI
	ADOXQ AX, R13

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ R8, DI
	MULXQ q<>+40(SB), AX, R8
	ADOXQ AX, DI

	// t[5] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R8
	ADOXQ R9, R8

	// clear the flags
	XORQ AX, AX
	MOVQ x+8(FP), DX
	MOVQ 64(DX), DX

	// (A,t[0])  := t[0] + x[0]*y[2] + A
	MULXQ R10, AX, R9
	ADOXQ AX, SI

	// (A,t[1])  := t[1] + x[1]*y[2] + A
	ADCXQ R9, R11
	MULXQ R14, AX, R9
	ADOXQ AX, R11

	// (A,t[2])  := t[2] + x[2]*y[2] + A
	ADCXQ R9, R12
	MULXQ R15, AX, R9
	ADOXQ AX, R12

	// (A,t[3])  := t[3] + x[3]*y[2] + A
	ADCXQ R9, R13
	MULXQ CX, AX, R9
	ADOXQ AX, R13

	// (A,t[4])  := t[4] + x[4]*y[2] + A
	ADCXQ R9, DI
	MULXQ BX, AX, R9
	ADOXQ AX, DI

	// (A,t[5])  := t[5] + x[5]*y[2] + A
	ADCXQ R9, R8
	MULXQ BP, AX, R9
	ADOXQ AX, R8

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R9
	ADOXQ AX, R9

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ SI, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	PUSHQ R9
	MULXQ q<>+0(SB), AX, R9
	ADCXQ SI, AX
	MOVQ  R9, SI
	POPQ  R9

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R11, SI
	MULXQ q<>+8(SB), AX, R11
	ADOXQ AX, SI

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R12, R11
	MULXQ q<>+16(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ R13, R12
	MULXQ q<>+24(SB), AX, R13
	ADOXQ AX, R12

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ DI, R13
	MULXQ q<>+32(SB), AX, DI
	ADOXQ AX, R13

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ R8, DI
	MULXQ q<>+40(SB), AX, R8
	ADOXQ AX, DI

	// t[5] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R8
	ADOXQ R9, R8

	// clear the flags
	XORQ AX, AX
	MOVQ x+8(FP), DX
	MOVQ 72(DX), DX

	// (A,t[0])  := t[0] + x[0]*y[3] + A
	MULXQ R10, AX, R9
	ADOXQ AX, SI

	// (A,t[1])  := t[1] + x[1]*y[3] + A
	ADCXQ R9, R11
	MULXQ R14, AX, R9
	ADOXQ AX, R11

	// (A,t[2])  := t[2] + x[2]*y[3] + A
	ADCXQ R9, R12
	MULXQ R15, AX, R9
	ADOXQ AX, R12

	// (A,t[3])  := t[3] + x[3]*y[3] + A
	ADCXQ R9, R13
	MULXQ CX, AX, R9
	ADOXQ AX, R13

	// (A,t[4])  := t[4] + x[4]*y[3] + A
	ADCXQ R9, DI
	MULXQ BX, AX, R9
	ADOXQ AX, DI

	// (A,t[5])  := t[5] + x[5]*y[3] + A
	ADCXQ R9, R8
	MULXQ BP, AX, R9
	ADOXQ AX, R8

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R9
	ADOXQ AX, R9

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ SI, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	PUSHQ R9
	MULXQ q<>+0(SB), AX, R9
	ADCXQ SI, AX
	MOVQ  R9, SI
	POPQ  R9

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R11, SI
	MULXQ q<>+8(SB), AX, R11
	ADOXQ AX, SI

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R12, R11
	MULXQ q<>+16(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ R13, R12
	MULXQ q<>+24(SB), AX, R13
	ADOXQ AX, R12

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ DI, R13
	MULXQ q<>+32(SB), AX, DI
	ADOXQ AX, R13

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ R8, DI
	MULXQ q<>+40(SB), AX, R8
	ADOXQ AX, DI

	// t[5] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R8
	ADOXQ R9, R8

	// clear the flags
	XORQ AX, AX
	MOVQ x+8(FP), DX
	MOVQ 80(DX), DX

	// (A,t[0])  := t[0] + x[0]*y[4] + A
	MULXQ R10, AX, R9
	ADOXQ AX, SI

	// (A,t[1])  := t[1] + x[1]*y[4] + A
	ADCXQ R9, R11
	MULXQ R14, AX, R9
	ADOXQ AX, R11

	// (A,t[2])  := t[2] + x[2]*y[4] + A
	ADCXQ R9, R12
	MULXQ R15, AX, R9
	ADOXQ AX, R12

	// (A,t[3])  := t[3] + x[3]*y[4] + A
	ADCXQ R9, R13
	MULXQ CX, AX, R9
	ADOXQ AX, R13

	// (A,t[4])  := t[4] + x[4]*y[4] + A
	ADCXQ R9, DI
	MULXQ BX, AX, R9
	ADOXQ AX, DI

	// (A,t[5])  := t[5] + x[5]*y[4] + A
	ADCXQ R9, R8
	MULXQ BP, AX, R9
	ADOXQ AX, R8

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R9
	ADOXQ AX, R9

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ SI, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	PUSHQ R9
	MULXQ q<>+0(SB), AX, R9
	ADCXQ SI, AX
	MOVQ  R9, SI
	POPQ  R9

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R11, SI
	MULXQ q<>+8(SB), AX, R11
	ADOXQ AX, SI

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R12, R11
	MULXQ q<>+16(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ R13, R12
	MULXQ q<>+24(SB), AX, R13
	ADOXQ AX, R12

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ DI, R13
	MULXQ q<>+32(SB), AX, DI
	ADOXQ AX, R13

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ R8, DI
	MULXQ q<>+40(SB), AX, R8
	ADOXQ AX, DI

	// t[5] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R8
	ADOXQ R9, R8

	// clear the flags
	XORQ AX, AX
	MOVQ x+8(FP), DX
	MOVQ 88(DX), DX

	// (A,t[0])  := t[0] + x[0]*y[5] + A
	MULXQ R10, AX, R9
	ADOXQ AX, SI

	// (A,t[1])  := t[1] + x[1]*y[5] + A
	ADCXQ R9, R11
	MULXQ R14, AX, R9
	ADOXQ AX, R11

	// (A,t[2])  := t[2] + x[2]*y[5] + A
	ADCXQ R9, R12
	MULXQ R15, AX, R9
	ADOXQ AX, R12

	// (A,t[3])  := t[3] + x[3]*y[5] + A
	ADCXQ R9, R13
	MULXQ CX, AX, R9
	ADOXQ AX, R13

	// (A,t[4])  := t[4] + x[4]*y[5] + A
	ADCXQ R9, DI
	MULXQ BX, AX, R9
	ADOXQ AX, DI

	// (A,t[5])  := t[5] + x[5]*y[5] + A
	ADCXQ R9, R8
	MULXQ BP, AX, R9
	ADOXQ AX, R8

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R9
	ADOXQ AX, R9

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ SI, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	PUSHQ R9
	MULXQ q<>+0(SB), AX, R9
	ADCXQ SI, AX
	MOVQ  R9, SI
	POPQ  R9

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R11, SI
	MULXQ q<>+8(SB), AX, R11
	ADOXQ AX, SI

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R12, R11
	MULXQ q<>+16(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ R13, R12
	MULXQ q<>+24(SB), AX, R13
	ADOXQ AX, R12

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ DI, R13
	MULXQ q<>+32(SB), AX, DI
	ADOXQ AX, R13

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ R8, DI
	MULXQ q<>+40(SB), AX, R8
	ADOXQ AX, DI

	// t[5] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R8
	ADOXQ R9, R8

	// reduce element(SI,R11,R12,R13,DI,R8) using temp registers (R9,R10,R14,R15,CX,BX)
	REDUCE(SI,R11,R12,R13,DI,R8,R9,R10,R14,R15,CX,BX)

	ADDQ SI, SI
	ADCQ R11, R11
	ADCQ R12, R12
	ADCQ R13, R13
	ADCQ DI, DI
	ADCQ R8, R8
	MOVQ res+0(FP), DX

	// reduce element(SI,R11,R12,R13,DI,R8) using temp registers (BP,R9,R10,R14,R15,CX)
	// stores in (48(DX),56(DX),64(DX),72(DX),80(DX),88(DX))
	REDUCE_AND_MOVE(SI,R11,R12,R13,DI,R8,BP,R9,R10,R14,R15,CX,48(DX),56(DX),64(DX),72(DX),80(DX),88(DX))

	RET

l4:
	MOVQ res+0(FP), AX
	MOVQ AX, (SP)
	MOVQ x+8(FP), AX
	MOVQ AX, 8(SP)
	CALL ·squareGenericE2(SB)
	RET

TEXT ·mulAdxE2(SB), $152-24
	NO_LOCAL_POINTERS
	CMPB ·supportAdx(SB), $1
	JNE  l5
	MOVQ x+8(FP), AX
	MOVQ 0(AX), R14
	MOVQ 8(AX), R15
	MOVQ 16(AX), CX
	MOVQ 24(AX), BX
	MOVQ 32(AX), BP
	MOVQ 40(AX), SI

	// t[0] = DI
	// t[1] = R8
	// t[2] = R9
	// t[3] = R10
	// t[4] = R11
	// t[5] = R12

	// clear the flags
	XORQ AX, AX
	MOVQ y+16(FP), DX
	MOVQ 0(DX), DX

	// (A,t[0])  := t[0] + x[0]*y[0] + A
	MULXQ R14, DI, R8

	// (A,t[1])  := t[1] + x[1]*y[0] + A
	MULXQ R15, AX, R9
	ADOXQ AX, R8

	// (A,t[2])  := t[2] + x[2]*y[0] + A
	MULXQ CX, AX, R10
	ADOXQ AX, R9

	// (A,t[3])  := t[3] + x[3]*y[0] + A
	MULXQ BX, AX, R11
	ADOXQ AX, R10

	// (A,t[4])  := t[4] + x[4]*y[0] + A
	MULXQ BP, AX, R12
	ADOXQ AX, R11

	// (A,t[5])  := t[5] + x[5]*y[0] + A
	MULXQ SI, AX, R13
	ADOXQ AX, R12

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADOXQ AX, R13

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ DI, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	PUSHQ R13
	MULXQ q<>+0(SB), AX, R13
	ADCXQ DI, AX
	MOVQ  R13, DI
	POPQ  R13

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R8, DI
	MULXQ q<>+8(SB), AX, R8
	ADOXQ AX, DI

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R9, R8
	MULXQ q<>+16(SB), AX, R9
	ADOXQ AX, R8

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ R10, R9
	MULXQ q<>+24(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ R11, R10
	MULXQ q<>+32(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ R12, R11
	MULXQ q<>+40(SB), AX, R12
	ADOXQ AX, R11

	// t[5] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R12
	ADOXQ R13, R12

	// clear the flags
	XORQ AX, AX
	MOVQ y+16(FP), DX
	MOVQ 8(DX), DX

	// (A,t[0])  := t[0] + x[0]*y[1] + A
	MULXQ R14, AX, R13
	ADOXQ AX, DI

	// (A,t[1])  := t[1] + x[1]*y[1] + A
	ADCXQ R13, R8
	MULXQ R15, AX, R13
	ADOXQ AX, R8

	// (A,t[2])  := t[2] + x[2]*y[1] + A
	ADCXQ R13, R9
	MULXQ CX, AX, R13
	ADOXQ AX, R9

	// (A,t[3])  := t[3] + x[3]*y[1] + A
	ADCXQ R13, R10
	MULXQ BX, AX, R13
	ADOXQ AX, R10

	// (A,t[4])  := t[4] + x[4]*y[1] + A
	ADCXQ R13, R11
	MULXQ BP, AX, R13
	ADOXQ AX, R11

	// (A,t[5])  := t[5] + x[5]*y[1] + A
	ADCXQ R13, R12
	MULXQ SI, AX, R13
	ADOXQ AX, R12

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R13
	ADOXQ AX, R13

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ DI, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	PUSHQ R13
	MULXQ q<>+0(SB), AX, R13
	ADCXQ DI, AX
	MOVQ  R13, DI
	POPQ  R13

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R8, DI
	MULXQ q<>+8(SB), AX, R8
	ADOXQ AX, DI

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R9, R8
	MULXQ q<>+16(SB), AX, R9
	ADOXQ AX, R8

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ R10, R9
	MULXQ q<>+24(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ R11, R10
	MULXQ q<>+32(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ R12, R11
	MULXQ q<>+40(SB), AX, R12
	ADOXQ AX, R11

	// t[5] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R12
	ADOXQ R13, R12

	// clear the flags
	XORQ AX, AX
	MOVQ y+16(FP), DX
	MOVQ 16(DX), DX

	// (A,t[0])  := t[0] + x[0]*y[2] + A
	MULXQ R14, AX, R13
	ADOXQ AX, DI

	// (A,t[1])  := t[1] + x[1]*y[2] + A
	ADCXQ R13, R8
	MULXQ R15, AX, R13
	ADOXQ AX, R8

	// (A,t[2])  := t[2] + x[2]*y[2] + A
	ADCXQ R13, R9
	MULXQ CX, AX, R13
	ADOXQ AX, R9

	// (A,t[3])  := t[3] + x[3]*y[2] + A
	ADCXQ R13, R10
	MULXQ BX, AX, R13
	ADOXQ AX, R10

	// (A,t[4])  := t[4] + x[4]*y[2] + A
	ADCXQ R13, R11
	MULXQ BP, AX, R13
	ADOXQ AX, R11

	// (A,t[5])  := t[5] + x[5]*y[2] + A
	ADCXQ R13, R12
	MULXQ SI, AX, R13
	ADOXQ AX, R12

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R13
	ADOXQ AX, R13

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ DI, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	PUSHQ R13
	MULXQ q<>+0(SB), AX, R13
	ADCXQ DI, AX
	MOVQ  R13, DI
	POPQ  R13

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R8, DI
	MULXQ q<>+8(SB), AX, R8
	ADOXQ AX, DI

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R9, R8
	MULXQ q<>+16(SB), AX, R9
	ADOXQ AX, R8

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ R10, R9
	MULXQ q<>+24(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ R11, R10
	MULXQ q<>+32(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ R12, R11
	MULXQ q<>+40(SB), AX, R12
	ADOXQ AX, R11

	// t[5] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R12
	ADOXQ R13, R12

	// clear the flags
	XORQ AX, AX
	MOVQ y+16(FP), DX
	MOVQ 24(DX), DX

	// (A,t[0])  := t[0] + x[0]*y[3] + A
	MULXQ R14, AX, R13
	ADOXQ AX, DI

	// (A,t[1])  := t[1] + x[1]*y[3] + A
	ADCXQ R13, R8
	MULXQ R15, AX, R13
	ADOXQ AX, R8

	// (A,t[2])  := t[2] + x[2]*y[3] + A
	ADCXQ R13, R9
	MULXQ CX, AX, R13
	ADOXQ AX, R9

	// (A,t[3])  := t[3] + x[3]*y[3] + A
	ADCXQ R13, R10
	MULXQ BX, AX, R13
	ADOXQ AX, R10

	// (A,t[4])  := t[4] + x[4]*y[3] + A
	ADCXQ R13, R11
	MULXQ BP, AX, R13
	ADOXQ AX, R11

	// (A,t[5])  := t[5] + x[5]*y[3] + A
	ADCXQ R13, R12
	MULXQ SI, AX, R13
	ADOXQ AX, R12

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R13
	ADOXQ AX, R13

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ DI, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	PUSHQ R13
	MULXQ q<>+0(SB), AX, R13
	ADCXQ DI, AX
	MOVQ  R13, DI
	POPQ  R13

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R8, DI
	MULXQ q<>+8(SB), AX, R8
	ADOXQ AX, DI

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R9, R8
	MULXQ q<>+16(SB), AX, R9
	ADOXQ AX, R8

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ R10, R9
	MULXQ q<>+24(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ R11, R10
	MULXQ q<>+32(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ R12, R11
	MULXQ q<>+40(SB), AX, R12
	ADOXQ AX, R11

	// t[5] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R12
	ADOXQ R13, R12

	// clear the flags
	XORQ AX, AX
	MOVQ y+16(FP), DX
	MOVQ 32(DX), DX

	// (A,t[0])  := t[0] + x[0]*y[4] + A
	MULXQ R14, AX, R13
	ADOXQ AX, DI

	// (A,t[1])  := t[1] + x[1]*y[4] + A
	ADCXQ R13, R8
	MULXQ R15, AX, R13
	ADOXQ AX, R8

	// (A,t[2])  := t[2] + x[2]*y[4] + A
	ADCXQ R13, R9
	MULXQ CX, AX, R13
	ADOXQ AX, R9

	// (A,t[3])  := t[3] + x[3]*y[4] + A
	ADCXQ R13, R10
	MULXQ BX, AX, R13
	ADOXQ AX, R10

	// (A,t[4])  := t[4] + x[4]*y[4] + A
	ADCXQ R13, R11
	MULXQ BP, AX, R13
	ADOXQ AX, R11

	// (A,t[5])  := t[5] + x[5]*y[4] + A
	ADCXQ R13, R12
	MULXQ SI, AX, R13
	ADOXQ AX, R12

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R13
	ADOXQ AX, R13

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ DI, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	PUSHQ R13
	MULXQ q<>+0(SB), AX, R13
	ADCXQ DI, AX
	MOVQ  R13, DI
	POPQ  R13

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R8, DI
	MULXQ q<>+8(SB), AX, R8
	ADOXQ AX, DI

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R9, R8
	MULXQ q<>+16(SB), AX, R9
	ADOXQ AX, R8

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ R10, R9
	MULXQ q<>+24(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ R11, R10
	MULXQ q<>+32(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ R12, R11
	MULXQ q<>+40(SB), AX, R12
	ADOXQ AX, R11

	// t[5] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R12
	ADOXQ R13, R12

	// clear the flags
	XORQ AX, AX
	MOVQ y+16(FP), DX
	MOVQ 40(DX), DX

	// (A,t[0])  := t[0] + x[0]*y[5] + A
	MULXQ R14, AX, R13
	ADOXQ AX, DI

	// (A,t[1])  := t[1] + x[1]*y[5] + A
	ADCXQ R13, R8
	MULXQ R15, AX, R13
	ADOXQ AX, R8

	// (A,t[2])  := t[2] + x[2]*y[5] + A
	ADCXQ R13, R9
	MULXQ CX, AX, R13
	ADOXQ AX, R9

	// (A,t[3])  := t[3] + x[3]*y[5] + A
	ADCXQ R13, R10
	MULXQ BX, AX, R13
	ADOXQ AX, R10

	// (A,t[4])  := t[4] + x[4]*y[5] + A
	ADCXQ R13, R11
	MULXQ BP, AX, R13
	ADOXQ AX, R11

	// (A,t[5])  := t[5] + x[5]*y[5] + A
	ADCXQ R13, R12
	MULXQ SI, AX, R13
	ADOXQ AX, R12

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R13
	ADOXQ AX, R13

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ DI, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	PUSHQ R13
	MULXQ q<>+0(SB), AX, R13
	ADCXQ DI, AX
	MOVQ  R13, DI
	POPQ  R13

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R8, DI
	MULXQ q<>+8(SB), AX, R8
	ADOXQ AX, DI

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R9, R8
	MULXQ q<>+16(SB), AX, R9
	ADOXQ AX, R8

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ R10, R9
	MULXQ q<>+24(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ R11, R10
	MULXQ q<>+32(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ R12, R11
	MULXQ q<>+40(SB), AX, R12
	ADOXQ AX, R11

	// t[5] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R12
	ADOXQ R13, R12

	// reduce element(DI,R8,R9,R10,R11,R12) using temp registers (R13,R14,R15,CX,BX,BP)
	REDUCE(DI,R8,R9,R10,R11,R12,R13,R14,R15,CX,BX,BP)

	MOVQ DI, -16(SP)
	MOVQ R8, -24(SP)
	MOVQ R9, -32(SP)
	MOVQ R10, -40(SP)
	MOVQ R11, -48(SP)
	MOVQ R12, -56(SP)
	MOVQ x+8(FP), AX
	MOVQ 48(AX), SI
	MOVQ 56(AX), R13
	MOVQ 64(AX), R14
	MOVQ 72(AX), R15
	MOVQ 80(AX), CX
	MOVQ 88(AX), BX

	// t[0] = BP
	// t[1] = DI
	// t[2] = R8
	// t[3] = R9
	// t[4] = R10
	// t[5] = R11

	// clear the flags
	XORQ AX, AX
	MOVQ y+16(FP), DX
	MOVQ 48(DX), DX

	// (A,t[0])  := t[0] + x[0]*y[0] + A
	MULXQ SI, BP, DI

	// (A,t[1])  := t[1] + x[1]*y[0] + A
	MULXQ R13, AX, R8
	ADOXQ AX, DI

	// (A,t[2])  := t[2] + x[2]*y[0] + A
	MULXQ R14, AX, R9
	ADOXQ AX, R8

	// (A,t[3])  := t[3] + x[3]*y[0] + A
	MULXQ R15, AX, R10
	ADOXQ AX, R9

	// (A,t[4])  := t[4] + x[4]*y[0] + A
	MULXQ CX, AX, R11
	ADOXQ AX, R10

	// (A,t[5])  := t[5] + x[5]*y[0] + A
	MULXQ BX, AX, R12
	ADOXQ AX, R11

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADOXQ AX, R12

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ BP, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	PUSHQ R12
	MULXQ q<>+0(SB), AX, R12
	ADCXQ BP, AX
	MOVQ  R12, BP
	POPQ  R12

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ DI, BP
	MULXQ q<>+8(SB), AX, DI
	ADOXQ AX, BP

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R8, DI
	MULXQ q<>+16(SB), AX, R8
	ADOXQ AX, DI

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ R9, R8
	MULXQ q<>+24(SB), AX, R9
	ADOXQ AX, R8

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ R10, R9
	MULXQ q<>+32(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ R11, R10
	MULXQ q<>+40(SB), AX, R11
	ADOXQ AX, R10

	// t[5] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R11
	ADOXQ R12, R11

	// clear the flags
	XORQ AX, AX
	MOVQ y+16(FP), DX
	MOVQ 56(DX), DX

	// (A,t[0])  := t[0] + x[0]*y[1] + A
	MULXQ SI, AX, R12
	ADOXQ AX, BP

	// (A,t[1])  := t[1] + x[1]*y[1] + A
	ADCXQ R12, DI
	MULXQ R13, AX, R12
	ADOXQ AX, DI

	// (A,t[2])  := t[2] + x[2]*y[1] + A
	ADCXQ R12, R8
	MULXQ R14, AX, R12
	ADOXQ AX, R8

	// (A,t[3])  := t[3] + x[3]*y[1] + A
	ADCXQ R12, R9
	MULXQ R15, AX, R12
	ADOXQ AX, R9

	// (A,t[4])  := t[4] + x[4]*y[1] + A
	ADCXQ R12, R10
	MULXQ CX, AX, R12
	ADOXQ AX, R10

	// (A,t[5])  := t[5] + x[5]*y[1] + A
	ADCXQ R12, R11
	MULXQ BX, AX, R12
	ADOXQ AX, R11

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R12
	ADOXQ AX, R12

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ BP, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	PUSHQ R12
	MULXQ q<>+0(SB), AX, R12
	ADCXQ BP, AX
	MOVQ  R12, BP
	POPQ  R12

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ DI, BP
	MULXQ q<>+8(SB), AX, DI
	ADOXQ AX, BP

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R8, DI
	MULXQ q<>+16(SB), AX, R8
	ADOXQ AX, DI

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ R9, R8
	MULXQ q<>+24(SB), AX, R9
	ADOXQ AX, R8

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ R10, R9
	MULXQ q<>+32(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ R11, R10
	MULXQ q<>+40(SB), AX, R11
	ADOXQ AX, R10

	// t[5] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R11
	ADOXQ R12, R11

	// clear the flags
	XORQ AX, AX
	MOVQ y+16(FP), DX
	MOVQ 64(DX), DX

	// (A,t[0])  := t[0] + x[0]*y[2] + A
	MULXQ SI, AX, R12
	ADOXQ AX, BP

	// (A,t[1])  := t[1] + x[1]*y[2] + A
	ADCXQ R12, DI
	MULXQ R13, AX, R12
	ADOXQ AX, DI

	// (A,t[2])  := t[2] + x[2]*y[2] + A
	ADCXQ R12, R8
	MULXQ R14, AX, R12
	ADOXQ AX, R8

	// (A,t[3])  := t[3] + x[3]*y[2] + A
	ADCXQ R12, R9
	MULXQ R15, AX, R12
	ADOXQ AX, R9

	// (A,t[4])  := t[4] + x[4]*y[2] + A
	ADCXQ R12, R10
	MULXQ CX, AX, R12
	ADOXQ AX, R10

	// (A,t[5])  := t[5] + x[5]*y[2] + A
	ADCXQ R12, R11
	MULXQ BX, AX, R12
	ADOXQ AX, R11

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R12
	ADOXQ AX, R12

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ BP, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	PUSHQ R12
	MULXQ q<>+0(SB), AX, R12
	ADCXQ BP, AX
	MOVQ  R12, BP
	POPQ  R12

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ DI, BP
	MULXQ q<>+8(SB), AX, DI
	ADOXQ AX, BP

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R8, DI
	MULXQ q<>+16(SB), AX, R8
	ADOXQ AX, DI

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ R9, R8
	MULXQ q<>+24(SB), AX, R9
	ADOXQ AX, R8

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ R10, R9
	MULXQ q<>+32(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ R11, R10
	MULXQ q<>+40(SB), AX, R11
	ADOXQ AX, R10

	// t[5] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R11
	ADOXQ R12, R11

	// clear the flags
	XORQ AX, AX
	MOVQ y+16(FP), DX
	MOVQ 72(DX), DX

	// (A,t[0])  := t[0] + x[0]*y[3] + A
	MULXQ SI, AX, R12
	ADOXQ AX, BP

	// (A,t[1])  := t[1] + x[1]*y[3] + A
	ADCXQ R12, DI
	MULXQ R13, AX, R12
	ADOXQ AX, DI

	// (A,t[2])  := t[2] + x[2]*y[3] + A
	ADCXQ R12, R8
	MULXQ R14, AX, R12
	ADOXQ AX, R8

	// (A,t[3])  := t[3] + x[3]*y[3] + A
	ADCXQ R12, R9
	MULXQ R15, AX, R12
	ADOXQ AX, R9

	// (A,t[4])  := t[4] + x[4]*y[3] + A
	ADCXQ R12, R10
	MULXQ CX, AX, R12
	ADOXQ AX, R10

	// (A,t[5])  := t[5] + x[5]*y[3] + A
	ADCXQ R12, R11
	MULXQ BX, AX, R12
	ADOXQ AX, R11

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R12
	ADOXQ AX, R12

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ BP, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	PUSHQ R12
	MULXQ q<>+0(SB), AX, R12
	ADCXQ BP, AX
	MOVQ  R12, BP
	POPQ  R12

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ DI, BP
	MULXQ q<>+8(SB), AX, DI
	ADOXQ AX, BP

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R8, DI
	MULXQ q<>+16(SB), AX, R8
	ADOXQ AX, DI

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ R9, R8
	MULXQ q<>+24(SB), AX, R9
	ADOXQ AX, R8

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ R10, R9
	MULXQ q<>+32(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ R11, R10
	MULXQ q<>+40(SB), AX, R11
	ADOXQ AX, R10

	// t[5] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R11
	ADOXQ R12, R11

	// clear the flags
	XORQ AX, AX
	MOVQ y+16(FP), DX
	MOVQ 80(DX), DX

	// (A,t[0])  := t[0] + x[0]*y[4] + A
	MULXQ SI, AX, R12
	ADOXQ AX, BP

	// (A,t[1])  := t[1] + x[1]*y[4] + A
	ADCXQ R12, DI
	MULXQ R13, AX, R12
	ADOXQ AX, DI

	// (A,t[2])  := t[2] + x[2]*y[4] + A
	ADCXQ R12, R8
	MULXQ R14, AX, R12
	ADOXQ AX, R8

	// (A,t[3])  := t[3] + x[3]*y[4] + A
	ADCXQ R12, R9
	MULXQ R15, AX, R12
	ADOXQ AX, R9

	// (A,t[4])  := t[4] + x[4]*y[4] + A
	ADCXQ R12, R10
	MULXQ CX, AX, R12
	ADOXQ AX, R10

	// (A,t[5])  := t[5] + x[5]*y[4] + A
	ADCXQ R12, R11
	MULXQ BX, AX, R12
	ADOXQ AX, R11

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R12
	ADOXQ AX, R12

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ BP, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	PUSHQ R12
	MULXQ q<>+0(SB), AX, R12
	ADCXQ BP, AX
	MOVQ  R12, BP
	POPQ  R12

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ DI, BP
	MULXQ q<>+8(SB), AX, DI
	ADOXQ AX, BP

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R8, DI
	MULXQ q<>+16(SB), AX, R8
	ADOXQ AX, DI

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ R9, R8
	MULXQ q<>+24(SB), AX, R9
	ADOXQ AX, R8

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ R10, R9
	MULXQ q<>+32(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ R11, R10
	MULXQ q<>+40(SB), AX, R11
	ADOXQ AX, R10

	// t[5] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R11
	ADOXQ R12, R11

	// clear the flags
	XORQ AX, AX
	MOVQ y+16(FP), DX
	MOVQ 88(DX), DX

	// (A,t[0])  := t[0] + x[0]*y[5] + A
	MULXQ SI, AX, R12
	ADOXQ AX, BP

	// (A,t[1])  := t[1] + x[1]*y[5] + A
	ADCXQ R12, DI
	MULXQ R13, AX, R12
	ADOXQ AX, DI

	// (A,t[2])  := t[2] + x[2]*y[5] + A
	ADCXQ R12, R8
	MULXQ R14, AX, R12
	ADOXQ AX, R8

	// (A,t[3])  := t[3] + x[3]*y[5] + A
	ADCXQ R12, R9
	MULXQ R15, AX, R12
	ADOXQ AX, R9

	// (A,t[4])  := t[4] + x[4]*y[5] + A
	ADCXQ R12, R10
	MULXQ CX, AX, R12
	ADOXQ AX, R10

	// (A,t[5])  := t[5] + x[5]*y[5] + A
	ADCXQ R12, R11
	MULXQ BX, AX, R12
	ADOXQ AX, R11

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R12
	ADOXQ AX, R12

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ BP, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	PUSHQ R12
	MULXQ q<>+0(SB), AX, R12
	ADCXQ BP, AX
	MOVQ  R12, BP
	POPQ  R12

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ DI, BP
	MULXQ q<>+8(SB), AX, DI
	ADOXQ AX, BP

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R8, DI
	MULXQ q<>+16(SB), AX, R8
	ADOXQ AX, DI

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ R9, R8
	MULXQ q<>+24(SB), AX, R9
	ADOXQ AX, R8

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ R10, R9
	MULXQ q<>+32(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ R11, R10
	MULXQ q<>+40(SB), AX, R11
	ADOXQ AX, R10

	// t[5] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R11
	ADOXQ R12, R11

	// reduce element(BP,DI,R8,R9,R10,R11) using temp registers (R12,SI,R13,R14,R15,CX)
	REDUCE(BP,DI,R8,R9,R10,R11,R12,SI,R13,R14,R15,CX)

	MOVQ BP, -64(SP)
	MOVQ DI, -72(SP)
	MOVQ R8, -80(SP)
	MOVQ R9, -88(SP)
	MOVQ R10, -96(SP)
	MOVQ R11, -104(SP)
	MOVQ x+8(FP), AX
	MOVQ 0(AX), BX
	MOVQ 8(AX), R12
	MOVQ 16(AX), SI
	MOVQ 24(AX), R13
	MOVQ 32(AX), R14
	MOVQ 40(AX), R15
	ADDQ 48(AX), BX
	ADCQ 56(AX), R12
	ADCQ 64(AX), SI
	ADCQ 72(AX), R13
	ADCQ 80(AX), R14
	ADCQ 88(AX), R15

	// reduce element(BX,R12,SI,R13,R14,R15) using temp registers (CX,BP,DI,R8,R9,R10)
	REDUCE(BX,R12,SI,R13,R14,R15,CX,BP,DI,R8,R9,R10)

	MOVQ BX, -112(SP)
	MOVQ R12, -120(SP)
	MOVQ SI, -128(SP)
	MOVQ R13, -136(SP)
	MOVQ R14, -144(SP)
	MOVQ R15, -152(SP)
	MOVQ y+16(FP), DX
	MOVQ 0(DX), BX
	MOVQ 8(DX), R12
	MOVQ 16(DX), SI
	MOVQ 24(DX), R13
	MOVQ 32(DX), R14
	MOVQ 40(DX), R15
	ADDQ 48(DX), BX
	ADCQ 56(DX), R12
	ADCQ 64(DX), SI
	ADCQ 72(DX), R13
	ADCQ 80(DX), R14
	ADCQ 88(DX), R15

	// reduce element(BX,R12,SI,R13,R14,R15) using temp registers (R11,CX,BP,DI,R8,R9)
	REDUCE(BX,R12,SI,R13,R14,R15,R11,CX,BP,DI,R8,R9)

	// t[0] = R10
	// t[1] = R11
	// t[2] = CX
	// t[3] = BP
	// t[4] = DI
	// t[5] = R8

	// clear the flags
	XORQ AX, AX
	MOVQ -112(SP), DX

	// (A,t[0])  := t[0] + x[0]*y[0] + A
	MULXQ BX, R10, R11

	// (A,t[1])  := t[1] + x[1]*y[0] + A
	MULXQ R12, AX, CX
	ADOXQ AX, R11

	// (A,t[2])  := t[2] + x[2]*y[0] + A
	MULXQ SI, AX, BP
	ADOXQ AX, CX

	// (A,t[3])  := t[3] + x[3]*y[0] + A
	MULXQ R13, AX, DI
	ADOXQ AX, BP

	// (A,t[4])  := t[4] + x[4]*y[0] + A
	MULXQ R14, AX, R8
	ADOXQ AX, DI

	// (A,t[5])  := t[5] + x[5]*y[0] + A
	MULXQ R15, AX, R9
	ADOXQ AX, R8

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADOXQ AX, R9

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R10, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	PUSHQ R9
	MULXQ q<>+0(SB), AX, R9
	ADCXQ R10, AX
	MOVQ  R9, R10
	POPQ  R9

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R11, R10
	MULXQ q<>+8(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ CX, R11
	MULXQ q<>+16(SB), AX, CX
	ADOXQ AX, R11

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BP, CX
	MULXQ q<>+24(SB), AX, BP
	ADOXQ AX, CX

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ DI, BP
	MULXQ q<>+32(SB), AX, DI
	ADOXQ AX, BP

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ R8, DI
	MULXQ q<>+40(SB), AX, R8
	ADOXQ AX, DI

	// t[5] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R8
	ADOXQ R9, R8

	// clear the flags
	XORQ AX, AX
	MOVQ -120(SP), DX

	// (A,t[0])  := t[0] + x[0]*y[1] + A
	MULXQ BX, AX, R9
	ADOXQ AX, R10

	// (A,t[1])  := t[1] + x[1]*y[1] + A
	ADCXQ R9, R11
	MULXQ R12, AX, R9
	ADOXQ AX, R11

	// (A,t[2])  := t[2] + x[2]*y[1] + A
	ADCXQ R9, CX
	MULXQ SI, AX, R9
	ADOXQ AX, CX

	// (A,t[3])  := t[3] + x[3]*y[1] + A
	ADCXQ R9, BP
	MULXQ R13, AX, R9
	ADOXQ AX, BP

	// (A,t[4])  := t[4] + x[4]*y[1] + A
	ADCXQ R9, DI
	MULXQ R14, AX, R9
	ADOXQ AX, DI

	// (A,t[5])  := t[5] + x[5]*y[1] + A
	ADCXQ R9, R8
	MULXQ R15, AX, R9
	ADOXQ AX, R8

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R9
	ADOXQ AX, R9

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R10, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	PUSHQ R9
	MULXQ q<>+0(SB), AX, R9
	ADCXQ R10, AX
	MOVQ  R9, R10
	POPQ  R9

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R11, R10
	MULXQ q<>+8(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ CX, R11
	MULXQ q<>+16(SB), AX, CX
	ADOXQ AX, R11

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BP, CX
	MULXQ q<>+24(SB), AX, BP
	ADOXQ AX, CX

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ DI, BP
	MULXQ q<>+32(SB), AX, DI
	ADOXQ AX, BP

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ R8, DI
	MULXQ q<>+40(SB), AX, R8
	ADOXQ AX, DI

	// t[5] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R8
	ADOXQ R9, R8

	// clear the flags
	XORQ AX, AX
	MOVQ -128(SP), DX

	// (A,t[0])  := t[0] + x[0]*y[2] + A
	MULXQ BX, AX, R9
	ADOXQ AX, R10

	// (A,t[1])  := t[1] + x[1]*y[2] + A
	ADCXQ R9, R11
	MULXQ R12, AX, R9
	ADOXQ AX, R11

	// (A,t[2])  := t[2] + x[2]*y[2] + A
	ADCXQ R9, CX
	MULXQ SI, AX, R9
	ADOXQ AX, CX

	// (A,t[3])  := t[3] + x[3]*y[2] + A
	ADCXQ R9, BP
	MULXQ R13, AX, R9
	ADOXQ AX, BP

	// (A,t[4])  := t[4] + x[4]*y[2] + A
	ADCXQ R9, DI
	MULXQ R14, AX, R9
	ADOXQ AX, DI

	// (A,t[5])  := t[5] + x[5]*y[2] + A
	ADCXQ R9, R8
	MULXQ R15, AX, R9
	ADOXQ AX, R8

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R9
	ADOXQ AX, R9

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R10, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	PUSHQ R9
	MULXQ q<>+0(SB), AX, R9
	ADCXQ R10, AX
	MOVQ  R9, R10
	POPQ  R9

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R11, R10
	MULXQ q<>+8(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ CX, R11
	MULXQ q<>+16(SB), AX, CX
	ADOXQ AX, R11

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BP, CX
	MULXQ q<>+24(SB), AX, BP
	ADOXQ AX, CX

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ DI, BP
	MULXQ q<>+32(SB), AX, DI
	ADOXQ AX, BP

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ R8, DI
	MULXQ q<>+40(SB), AX, R8
	ADOXQ AX, DI

	// t[5] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R8
	ADOXQ R9, R8

	// clear the flags
	XORQ AX, AX
	MOVQ -136(SP), DX

	// (A,t[0])  := t[0] + x[0]*y[3] + A
	MULXQ BX, AX, R9
	ADOXQ AX, R10

	// (A,t[1])  := t[1] + x[1]*y[3] + A
	ADCXQ R9, R11
	MULXQ R12, AX, R9
	ADOXQ AX, R11

	// (A,t[2])  := t[2] + x[2]*y[3] + A
	ADCXQ R9, CX
	MULXQ SI, AX, R9
	ADOXQ AX, CX

	// (A,t[3])  := t[3] + x[3]*y[3] + A
	ADCXQ R9, BP
	MULXQ R13, AX, R9
	ADOXQ AX, BP

	// (A,t[4])  := t[4] + x[4]*y[3] + A
	ADCXQ R9, DI
	MULXQ R14, AX, R9
	ADOXQ AX, DI

	// (A,t[5])  := t[5] + x[5]*y[3] + A
	ADCXQ R9, R8
	MULXQ R15, AX, R9
	ADOXQ AX, R8

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R9
	ADOXQ AX, R9

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R10, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	PUSHQ R9
	MULXQ q<>+0(SB), AX, R9
	ADCXQ R10, AX
	MOVQ  R9, R10
	POPQ  R9

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R11, R10
	MULXQ q<>+8(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ CX, R11
	MULXQ q<>+16(SB), AX, CX
	ADOXQ AX, R11

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BP, CX
	MULXQ q<>+24(SB), AX, BP
	ADOXQ AX, CX

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ DI, BP
	MULXQ q<>+32(SB), AX, DI
	ADOXQ AX, BP

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ R8, DI
	MULXQ q<>+40(SB), AX, R8
	ADOXQ AX, DI

	// t[5] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R8
	ADOXQ R9, R8

	// clear the flags
	XORQ AX, AX
	MOVQ -144(SP), DX

	// (A,t[0])  := t[0] + x[0]*y[4] + A
	MULXQ BX, AX, R9
	ADOXQ AX, R10

	// (A,t[1])  := t[1] + x[1]*y[4] + A
	ADCXQ R9, R11
	MULXQ R12, AX, R9
	ADOXQ AX, R11

	// (A,t[2])  := t[2] + x[2]*y[4] + A
	ADCXQ R9, CX
	MULXQ SI, AX, R9
	ADOXQ AX, CX

	// (A,t[3])  := t[3] + x[3]*y[4] + A
	ADCXQ R9, BP
	MULXQ R13, AX, R9
	ADOXQ AX, BP

	// (A,t[4])  := t[4] + x[4]*y[4] + A
	ADCXQ R9, DI
	MULXQ R14, AX, R9
	ADOXQ AX, DI

	// (A,t[5])  := t[5] + x[5]*y[4] + A
	ADCXQ R9, R8
	MULXQ R15, AX, R9
	ADOXQ AX, R8

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R9
	ADOXQ AX, R9

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R10, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	PUSHQ R9
	MULXQ q<>+0(SB), AX, R9
	ADCXQ R10, AX
	MOVQ  R9, R10
	POPQ  R9

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R11, R10
	MULXQ q<>+8(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ CX, R11
	MULXQ q<>+16(SB), AX, CX
	ADOXQ AX, R11

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BP, CX
	MULXQ q<>+24(SB), AX, BP
	ADOXQ AX, CX

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ DI, BP
	MULXQ q<>+32(SB), AX, DI
	ADOXQ AX, BP

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ R8, DI
	MULXQ q<>+40(SB), AX, R8
	ADOXQ AX, DI

	// t[5] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R8
	ADOXQ R9, R8

	// clear the flags
	XORQ AX, AX
	MOVQ -152(SP), DX

	// (A,t[0])  := t[0] + x[0]*y[5] + A
	MULXQ BX, AX, R9
	ADOXQ AX, R10

	// (A,t[1])  := t[1] + x[1]*y[5] + A
	ADCXQ R9, R11
	MULXQ R12, AX, R9
	ADOXQ AX, R11

	// (A,t[2])  := t[2] + x[2]*y[5] + A
	ADCXQ R9, CX
	MULXQ SI, AX, R9
	ADOXQ AX, CX

	// (A,t[3])  := t[3] + x[3]*y[5] + A
	ADCXQ R9, BP
	MULXQ R13, AX, R9
	ADOXQ AX, BP

	// (A,t[4])  := t[4] + x[4]*y[5] + A
	ADCXQ R9, DI
	MULXQ R14, AX, R9
	ADOXQ AX, DI

	// (A,t[5])  := t[5] + x[5]*y[5] + A
	ADCXQ R9, R8
	MULXQ R15, AX, R9
	ADOXQ AX, R8

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R9
	ADOXQ AX, R9

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R10, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	PUSHQ R9
	MULXQ q<>+0(SB), AX, R9
	ADCXQ R10, AX
	MOVQ  R9, R10
	POPQ  R9

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R11, R10
	MULXQ q<>+8(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ CX, R11
	MULXQ q<>+16(SB), AX, CX
	ADOXQ AX, R11

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BP, CX
	MULXQ q<>+24(SB), AX, BP
	ADOXQ AX, CX

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ DI, BP
	MULXQ q<>+32(SB), AX, DI
	ADOXQ AX, BP

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ R8, DI
	MULXQ q<>+40(SB), AX, R8
	ADOXQ AX, DI

	// t[5] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R8
	ADOXQ R9, R8

	// reduce element(R10,R11,CX,BP,DI,R8) using temp registers (R9,BX,R12,SI,R13,R14)
	REDUCE(R10,R11,CX,BP,DI,R8,R9,BX,R12,SI,R13,R14)

	MOVQ    z+0(FP), DX
	SUBQ    -16(SP), R10
	SBBQ    -24(SP), R11
	SBBQ    -32(SP), CX
	SBBQ    -40(SP), BP
	SBBQ    -48(SP), DI
	SBBQ    -56(SP), R8
	MOVQ    $0xb9feffffffffaaab, R15
	MOVQ    $0x1eabfffeb153ffff, R9
	MOVQ    $0x6730d2a0f6b0f624, BX
	MOVQ    $0x64774b84f38512bf, R12
	MOVQ    $0x4b1ba7b6434bacd7, SI
	MOVQ    $0x1a0111ea397fe69a, R13
	MOVQ    $0, R14
	CMOVQCC R14, R15
	CMOVQCC R14, R9
	CMOVQCC R14, BX
	CMOVQCC R14, R12
	CMOVQCC R14, SI
	CMOVQCC R14, R13
	ADDQ    R15, R10
	ADCQ    R9, R11
	ADCQ    BX, CX
	ADCQ    R12, BP
	ADCQ    SI, DI
	ADCQ    R13, R8
	SUBQ    -64(SP), R10
	SBBQ    -72(SP), R11
	SBBQ    -80(SP), CX
	SBBQ    -88(SP), BP
	SBBQ    -96(SP), DI
	SBBQ    -104(SP), R8
	MOVQ    $0xb9feffffffffaaab, R14
	MOVQ    $0x1eabfffeb153ffff, R15
	MOVQ    $0x6730d2a0f6b0f624, R9
	MOVQ    $0x64774b84f38512bf, BX
	MOVQ    $0x4b1ba7b6434bacd7, R12
	MOVQ    $0x1a0111ea397fe69a, SI
	MOVQ    $0, R13
	CMOVQCC R13, R14
	CMOVQCC R13, R15
	CMOVQCC R13, R9
	CMOVQCC R13, BX
	CMOVQCC R13, R12
	CMOVQCC R13, SI
	ADDQ    R14, R10
	ADCQ    R15, R11
	ADCQ    R9, CX
	ADCQ    BX, BP
	ADCQ    R12, DI
	ADCQ    SI, R8
	MOVQ    R10, 48(DX)
	MOVQ    R11, 56(DX)
	MOVQ    CX, 64(DX)
	MOVQ    BP, 72(DX)
	MOVQ    DI, 80(DX)
	MOVQ    R8, 88(DX)
	MOVQ    -16(SP), R10
	MOVQ    -24(SP), R11
	MOVQ    -32(SP), CX
	MOVQ    -40(SP), BP
	MOVQ    -48(SP), DI
	MOVQ    -56(SP), R8
	SUBQ    -64(SP), R10
	SBBQ    -72(SP), R11
	SBBQ    -80(SP), CX
	SBBQ    -88(SP), BP
	SBBQ    -96(SP), DI
	SBBQ    -104(SP), R8
	MOVQ    $0xb9feffffffffaaab, R13
	MOVQ    $0x1eabfffeb153ffff, R14
	MOVQ    $0x6730d2a0f6b0f624, R15
	MOVQ    $0x64774b84f38512bf, R9
	MOVQ    $0x4b1ba7b6434bacd7, BX
	MOVQ    $0x1a0111ea397fe69a, R12
	MOVQ    $0, SI
	CMOVQCC SI, R13
	CMOVQCC SI, R14
	CMOVQCC SI, R15
	CMOVQCC SI, R9
	CMOVQCC SI, BX
	CMOVQCC SI, R12
	ADDQ    R13, R10
	ADCQ    R14, R11
	ADCQ    R15, CX
	ADCQ    R9, BP
	ADCQ    BX, DI
	ADCQ    R12, R8
	MOVQ    R10, 0(DX)
	MOVQ    R11, 8(DX)
	MOVQ    CX, 16(DX)
	MOVQ    BP, 24(DX)
	MOVQ    DI, 32(DX)
	MOVQ    R8, 40(DX)
	RET

l5:
	MOVQ z+0(FP), AX
	MOVQ AX, (SP)
	MOVQ x+8(FP), AX
	MOVQ AX, 8(SP)
	MOVQ y+16(FP), AX
	MOVQ AX, 16(SP)
	CALL ·mulGenericE2(SB)
	RET
