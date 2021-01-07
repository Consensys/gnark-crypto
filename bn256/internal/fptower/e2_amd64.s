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
DATA q<>+0(SB)/8, $0x3c208c16d87cfd47
DATA q<>+8(SB)/8, $0x97816a916871ca8d
DATA q<>+16(SB)/8, $0xb85045b68181585d
DATA q<>+24(SB)/8, $0x30644e72e131a029
GLOBL q<>(SB), (RODATA+NOPTR), $32

// qInv0 q'[0]
DATA qInv0<>(SB)/8, $0x87d20782e4866389
GLOBL qInv0<>(SB), (RODATA+NOPTR), $8

#define REDUCE_AND_MOVE(ra0, ra1, ra2, ra3, rb0, rb1, rb2, rb3, res0, res1, res2, res3) \
	MOVQ    ra0, rb0;        \
	MOVQ    ra1, rb1;        \
	MOVQ    ra2, rb2;        \
	MOVQ    ra3, rb3;        \
	SUBQ    q<>(SB), rb0;    \
	SBBQ    q<>+8(SB), rb1;  \
	SBBQ    q<>+16(SB), rb2; \
	SBBQ    q<>+24(SB), rb3; \
	CMOVQCC rb0, ra0;        \
	CMOVQCC rb1, ra1;        \
	CMOVQCC rb2, ra2;        \
	CMOVQCC rb3, ra3;        \
	MOVQ    ra0, res0;       \
	MOVQ    ra1, res1;       \
	MOVQ    ra2, res2;       \
	MOVQ    ra3, res3;       \

#define REDUCE(ra0, ra1, ra2, ra3, rb0, rb1, rb2, rb3) \
	MOVQ    ra0, rb0;        \
	MOVQ    ra1, rb1;        \
	MOVQ    ra2, rb2;        \
	MOVQ    ra3, rb3;        \
	SUBQ    q<>(SB), rb0;    \
	SBBQ    q<>+8(SB), rb1;  \
	SBBQ    q<>+16(SB), rb2; \
	SBBQ    q<>+24(SB), rb3; \
	CMOVQCC rb0, ra0;        \
	CMOVQCC rb1, ra1;        \
	CMOVQCC rb2, ra2;        \
	CMOVQCC rb3, ra3;        \

TEXT ·addE2(SB), NOSPLIT, $0-24
	MOVQ x+8(FP), AX
	MOVQ 0(AX), BX
	MOVQ 8(AX), BP
	MOVQ 16(AX), SI
	MOVQ 24(AX), DI
	MOVQ y+16(FP), DX
	ADDQ 0(DX), BX
	ADCQ 8(DX), BP
	ADCQ 16(DX), SI
	ADCQ 24(DX), DI
	MOVQ res+0(FP), CX

	// reduce element(BX,BP,SI,DI) using temp registers (R8,R9,R10,R11)
	// stores in (0(CX),8(CX),16(CX),24(CX))
	REDUCE_AND_MOVE(BX,BP,SI,DI,R8,R9,R10,R11,0(CX),8(CX),16(CX),24(CX))

	MOVQ 32(AX), BX
	MOVQ 40(AX), BP
	MOVQ 48(AX), SI
	MOVQ 56(AX), DI
	ADDQ 32(DX), BX
	ADCQ 40(DX), BP
	ADCQ 48(DX), SI
	ADCQ 56(DX), DI

	// reduce element(BX,BP,SI,DI) using temp registers (R12,R13,R14,R15)
	// stores in (32(CX),40(CX),48(CX),56(CX))
	REDUCE_AND_MOVE(BX,BP,SI,DI,R12,R13,R14,R15,32(CX),40(CX),48(CX),56(CX))

	RET

TEXT ·doubleE2(SB), NOSPLIT, $0-16
	MOVQ res+0(FP), DX
	MOVQ x+8(FP), AX
	MOVQ 0(AX), CX
	MOVQ 8(AX), BX
	MOVQ 16(AX), BP
	MOVQ 24(AX), SI
	ADDQ CX, CX
	ADCQ BX, BX
	ADCQ BP, BP
	ADCQ SI, SI

	// reduce element(CX,BX,BP,SI) using temp registers (DI,R8,R9,R10)
	// stores in (0(DX),8(DX),16(DX),24(DX))
	REDUCE_AND_MOVE(CX,BX,BP,SI,DI,R8,R9,R10,0(DX),8(DX),16(DX),24(DX))

	MOVQ 32(AX), CX
	MOVQ 40(AX), BX
	MOVQ 48(AX), BP
	MOVQ 56(AX), SI
	ADDQ CX, CX
	ADCQ BX, BX
	ADCQ BP, BP
	ADCQ SI, SI

	// reduce element(CX,BX,BP,SI) using temp registers (R11,R12,R13,R14)
	// stores in (32(DX),40(DX),48(DX),56(DX))
	REDUCE_AND_MOVE(CX,BX,BP,SI,R11,R12,R13,R14,32(DX),40(DX),48(DX),56(DX))

	RET

TEXT ·subE2(SB), NOSPLIT, $0-24
	MOVQ    x+8(FP), BP
	MOVQ    y+16(FP), SI
	MOVQ    0(BP), AX
	MOVQ    8(BP), DX
	MOVQ    16(BP), CX
	MOVQ    24(BP), BX
	SUBQ    0(SI), AX
	SBBQ    8(SI), DX
	SBBQ    16(SI), CX
	SBBQ    24(SI), BX
	MOVQ    $0x3c208c16d87cfd47, DI
	MOVQ    $0x97816a916871ca8d, R8
	MOVQ    $0xb85045b68181585d, R9
	MOVQ    $0x30644e72e131a029, R10
	MOVQ    $0, R11
	CMOVQCC R11, DI
	CMOVQCC R11, R8
	CMOVQCC R11, R9
	CMOVQCC R11, R10
	ADDQ    DI, AX
	ADCQ    R8, DX
	ADCQ    R9, CX
	ADCQ    R10, BX
	MOVQ    res+0(FP), R12
	MOVQ    AX, 0(R12)
	MOVQ    DX, 8(R12)
	MOVQ    CX, 16(R12)
	MOVQ    BX, 24(R12)
	MOVQ    32(BP), AX
	MOVQ    40(BP), DX
	MOVQ    48(BP), CX
	MOVQ    56(BP), BX
	SUBQ    32(SI), AX
	SBBQ    40(SI), DX
	SBBQ    48(SI), CX
	SBBQ    56(SI), BX
	MOVQ    $0x3c208c16d87cfd47, R13
	MOVQ    $0x97816a916871ca8d, R14
	MOVQ    $0xb85045b68181585d, R15
	MOVQ    $0x30644e72e131a029, R11
	MOVQ    $0, DI
	CMOVQCC DI, R13
	CMOVQCC DI, R14
	CMOVQCC DI, R15
	CMOVQCC DI, R11
	ADDQ    R13, AX
	ADCQ    R14, DX
	ADCQ    R15, CX
	ADCQ    R11, BX
	MOVQ    res+0(FP), BP
	MOVQ    AX, 32(BP)
	MOVQ    DX, 40(BP)
	MOVQ    CX, 48(BP)
	MOVQ    BX, 56(BP)
	RET

TEXT ·negE2(SB), NOSPLIT, $0-16
	MOVQ  res+0(FP), DX
	MOVQ  x+8(FP), AX
	MOVQ  0(AX), BX
	MOVQ  8(AX), BP
	MOVQ  16(AX), SI
	MOVQ  24(AX), DI
	MOVQ  BX, AX
	ORQ   BP, AX
	ORQ   SI, AX
	ORQ   DI, AX
	TESTQ AX, AX
	JNE   l1
	MOVQ  AX, 32(DX)
	MOVQ  AX, 40(DX)
	MOVQ  AX, 48(DX)
	MOVQ  AX, 56(DX)
	JMP   l3

l1:
	MOVQ $0x3c208c16d87cfd47, CX
	SUBQ BX, CX
	MOVQ CX, 0(DX)
	MOVQ $0x97816a916871ca8d, CX
	SBBQ BP, CX
	MOVQ CX, 8(DX)
	MOVQ $0xb85045b68181585d, CX
	SBBQ SI, CX
	MOVQ CX, 16(DX)
	MOVQ $0x30644e72e131a029, CX
	SBBQ DI, CX
	MOVQ CX, 24(DX)

l3:
	MOVQ  x+8(FP), AX
	MOVQ  32(AX), BX
	MOVQ  40(AX), BP
	MOVQ  48(AX), SI
	MOVQ  56(AX), DI
	MOVQ  BX, AX
	ORQ   BP, AX
	ORQ   SI, AX
	ORQ   DI, AX
	TESTQ AX, AX
	JNE   l2
	MOVQ  AX, 32(DX)
	MOVQ  AX, 40(DX)
	MOVQ  AX, 48(DX)
	MOVQ  AX, 56(DX)
	RET

l2:
	MOVQ $0x3c208c16d87cfd47, CX
	SUBQ BX, CX
	MOVQ CX, 32(DX)
	MOVQ $0x97816a916871ca8d, CX
	SBBQ BP, CX
	MOVQ CX, 40(DX)
	MOVQ $0xb85045b68181585d, CX
	SBBQ SI, CX
	MOVQ CX, 48(DX)
	MOVQ $0x30644e72e131a029, CX
	SBBQ DI, CX
	MOVQ CX, 56(DX)
	RET

TEXT ·mulAdxE2(SB), $24-24
	NO_LOCAL_POINTERS
	CMPB ·supportAdx(SB), $1
	JNE  l4
	MOVQ x+8(FP), R9
	MOVQ 32(R9), R14
	MOVQ 40(R9), R15
	MOVQ 48(R9), CX
	MOVQ 56(R9), BX
	ADDQ 0(R9), R14
	ADCQ 8(R9), R15
	ADCQ 16(R9), CX
	ADCQ 24(R9), BX

	// reduce element(R14,R15,CX,BX) using temp registers (R10,R11,R12,R13)
	REDUCE(R14,R15,CX,BX,R10,R11,R12,R13)

	MOVQ y+16(FP), R9
	MOVQ 32(R9), BP
	MOVQ 40(R9), SI
	MOVQ 48(R9), DI
	MOVQ 56(R9), R8
	ADDQ 0(R9), BP
	ADCQ 8(R9), SI
	ADCQ 16(R9), DI
	ADCQ 24(R9), R8

	// reduce element(BP,SI,DI,R8) using temp registers (R10,R11,R12,R13)
	REDUCE(BP,SI,DI,R8,R10,R11,R12,R13)

	// t[0] = R10
	// t[1] = R11
	// t[2] = R12
	// t[3] = R13

	// clear the flags
	XORQ AX, AX
	MOVQ BP, DX

	// (A,t[0])  := t[0] + x[0]*y[0] + A
	MULXQ R14, R10, R11

	// (A,t[1])  := t[1] + x[1]*y[0] + A
	MULXQ R15, AX, R12
	ADOXQ AX, R11

	// (A,t[2])  := t[2] + x[2]*y[0] + A
	MULXQ CX, AX, R13
	ADOXQ AX, R12

	// (A,t[3])  := t[3] + x[3]*y[0] + A
	MULXQ BX, AX, R9
	ADOXQ AX, R13

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADOXQ AX, R9

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R10, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, BP
	ADCXQ R10, AX
	MOVQ  BP, R10

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R11, R10
	MULXQ q<>+8(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R12, R11
	MULXQ q<>+16(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ R13, R12
	MULXQ q<>+24(SB), AX, R13
	ADOXQ AX, R12

	// t[3] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R13
	ADOXQ R9, R13

	// clear the flags
	XORQ AX, AX
	MOVQ SI, DX

	// (A,t[0])  := t[0] + x[0]*y[1] + A
	MULXQ R14, AX, R9
	ADOXQ AX, R10

	// (A,t[1])  := t[1] + x[1]*y[1] + A
	ADCXQ R9, R11
	MULXQ R15, AX, R9
	ADOXQ AX, R11

	// (A,t[2])  := t[2] + x[2]*y[1] + A
	ADCXQ R9, R12
	MULXQ CX, AX, R9
	ADOXQ AX, R12

	// (A,t[3])  := t[3] + x[3]*y[1] + A
	ADCXQ R9, R13
	MULXQ BX, AX, R9
	ADOXQ AX, R13

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
	MULXQ q<>+0(SB), AX, BP
	ADCXQ R10, AX
	MOVQ  BP, R10

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R11, R10
	MULXQ q<>+8(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R12, R11
	MULXQ q<>+16(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ R13, R12
	MULXQ q<>+24(SB), AX, R13
	ADOXQ AX, R12

	// t[3] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R13
	ADOXQ R9, R13

	// clear the flags
	XORQ AX, AX
	MOVQ DI, DX

	// (A,t[0])  := t[0] + x[0]*y[2] + A
	MULXQ R14, AX, R9
	ADOXQ AX, R10

	// (A,t[1])  := t[1] + x[1]*y[2] + A
	ADCXQ R9, R11
	MULXQ R15, AX, R9
	ADOXQ AX, R11

	// (A,t[2])  := t[2] + x[2]*y[2] + A
	ADCXQ R9, R12
	MULXQ CX, AX, R9
	ADOXQ AX, R12

	// (A,t[3])  := t[3] + x[3]*y[2] + A
	ADCXQ R9, R13
	MULXQ BX, AX, R9
	ADOXQ AX, R13

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
	MULXQ q<>+0(SB), AX, SI
	ADCXQ R10, AX
	MOVQ  SI, R10

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R11, R10
	MULXQ q<>+8(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R12, R11
	MULXQ q<>+16(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ R13, R12
	MULXQ q<>+24(SB), AX, R13
	ADOXQ AX, R12

	// t[3] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R13
	ADOXQ R9, R13

	// clear the flags
	XORQ AX, AX
	MOVQ R8, DX

	// (A,t[0])  := t[0] + x[0]*y[3] + A
	MULXQ R14, AX, R9
	ADOXQ AX, R10

	// (A,t[1])  := t[1] + x[1]*y[3] + A
	ADCXQ R9, R11
	MULXQ R15, AX, R9
	ADOXQ AX, R11

	// (A,t[2])  := t[2] + x[2]*y[3] + A
	ADCXQ R9, R12
	MULXQ CX, AX, R9
	ADOXQ AX, R12

	// (A,t[3])  := t[3] + x[3]*y[3] + A
	ADCXQ R9, R13
	MULXQ BX, AX, R9
	ADOXQ AX, R13

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
	MULXQ q<>+0(SB), AX, BP
	ADCXQ R10, AX
	MOVQ  BP, R10

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R11, R10
	MULXQ q<>+8(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R12, R11
	MULXQ q<>+16(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ R13, R12
	MULXQ q<>+24(SB), AX, R13
	ADOXQ AX, R12

	// t[3] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R13
	ADOXQ R9, R13

	// reduce element(R10,R11,R12,R13) using temp registers (DI,SI,R8,BP)
	// stores in (R14,R15,CX,BX)
	REDUCE_AND_MOVE(R10,R11,R12,R13,DI,SI,R8,BP,R14,R15,CX,BX)

	// t[0] = DI
	// t[1] = SI
	// t[2] = R8
	// t[3] = BP

	// clear the flags
	XORQ AX, AX
	MOVQ y+16(FP), R9
	MOVQ 0(R9), DX

	// (A,t[0])  := t[0] + x[0]*y[0] + A
	MOVQ  x+8(FP), R9
	MULXQ 0(R9), DI, SI

	// (A,t[1])  := t[1] + x[1]*y[0] + A
	MOVQ  x+8(FP), R9
	MULXQ 8(R9), AX, R8
	ADOXQ AX, SI

	// (A,t[2])  := t[2] + x[2]*y[0] + A
	MOVQ  x+8(FP), R9
	MULXQ 16(R9), AX, BP
	ADOXQ AX, R8

	// (A,t[3])  := t[3] + x[3]*y[0] + A
	MOVQ  x+8(FP), R9
	MULXQ 24(R9), AX, R10
	ADOXQ AX, BP

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADOXQ AX, R10

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ DI, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, R11
	ADCXQ DI, AX
	MOVQ  R11, DI

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ SI, DI
	MULXQ q<>+8(SB), AX, SI
	ADOXQ AX, DI

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R8, SI
	MULXQ q<>+16(SB), AX, R8
	ADOXQ AX, SI

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BP, R8
	MULXQ q<>+24(SB), AX, BP
	ADOXQ AX, R8

	// t[3] = C + A
	MOVQ  $0, AX
	ADCXQ AX, BP
	ADOXQ R10, BP

	// clear the flags
	XORQ AX, AX
	MOVQ y+16(FP), R9
	MOVQ 8(R9), DX

	// (A,t[0])  := t[0] + x[0]*y[1] + A
	MOVQ  x+8(FP), R9
	MULXQ 0(R9), AX, R10
	ADOXQ AX, DI

	// (A,t[1])  := t[1] + x[1]*y[1] + A
	MOVQ  x+8(FP), R9
	ADCXQ R10, SI
	MULXQ 8(R9), AX, R10
	ADOXQ AX, SI

	// (A,t[2])  := t[2] + x[2]*y[1] + A
	MOVQ  x+8(FP), R9
	ADCXQ R10, R8
	MULXQ 16(R9), AX, R10
	ADOXQ AX, R8

	// (A,t[3])  := t[3] + x[3]*y[1] + A
	MOVQ  x+8(FP), R9
	ADCXQ R10, BP
	MULXQ 24(R9), AX, R10
	ADOXQ AX, BP

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R10
	ADOXQ AX, R10

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ DI, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, R12
	ADCXQ DI, AX
	MOVQ  R12, DI

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ SI, DI
	MULXQ q<>+8(SB), AX, SI
	ADOXQ AX, DI

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R8, SI
	MULXQ q<>+16(SB), AX, R8
	ADOXQ AX, SI

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BP, R8
	MULXQ q<>+24(SB), AX, BP
	ADOXQ AX, R8

	// t[3] = C + A
	MOVQ  $0, AX
	ADCXQ AX, BP
	ADOXQ R10, BP

	// clear the flags
	XORQ AX, AX
	MOVQ y+16(FP), R9
	MOVQ 16(R9), DX

	// (A,t[0])  := t[0] + x[0]*y[2] + A
	MOVQ  x+8(FP), R9
	MULXQ 0(R9), AX, R10
	ADOXQ AX, DI

	// (A,t[1])  := t[1] + x[1]*y[2] + A
	MOVQ  x+8(FP), R9
	ADCXQ R10, SI
	MULXQ 8(R9), AX, R10
	ADOXQ AX, SI

	// (A,t[2])  := t[2] + x[2]*y[2] + A
	MOVQ  x+8(FP), R9
	ADCXQ R10, R8
	MULXQ 16(R9), AX, R10
	ADOXQ AX, R8

	// (A,t[3])  := t[3] + x[3]*y[2] + A
	MOVQ  x+8(FP), R9
	ADCXQ R10, BP
	MULXQ 24(R9), AX, R10
	ADOXQ AX, BP

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R10
	ADOXQ AX, R10

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ DI, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, R13
	ADCXQ DI, AX
	MOVQ  R13, DI

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ SI, DI
	MULXQ q<>+8(SB), AX, SI
	ADOXQ AX, DI

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R8, SI
	MULXQ q<>+16(SB), AX, R8
	ADOXQ AX, SI

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BP, R8
	MULXQ q<>+24(SB), AX, BP
	ADOXQ AX, R8

	// t[3] = C + A
	MOVQ  $0, AX
	ADCXQ AX, BP
	ADOXQ R10, BP

	// clear the flags
	XORQ AX, AX
	MOVQ y+16(FP), R9
	MOVQ 24(R9), DX

	// (A,t[0])  := t[0] + x[0]*y[3] + A
	MOVQ  x+8(FP), R9
	MULXQ 0(R9), AX, R10
	ADOXQ AX, DI

	// (A,t[1])  := t[1] + x[1]*y[3] + A
	MOVQ  x+8(FP), R9
	ADCXQ R10, SI
	MULXQ 8(R9), AX, R10
	ADOXQ AX, SI

	// (A,t[2])  := t[2] + x[2]*y[3] + A
	MOVQ  x+8(FP), R9
	ADCXQ R10, R8
	MULXQ 16(R9), AX, R10
	ADOXQ AX, R8

	// (A,t[3])  := t[3] + x[3]*y[3] + A
	MOVQ  x+8(FP), R9
	ADCXQ R10, BP
	MULXQ 24(R9), AX, R10
	ADOXQ AX, BP

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R10
	ADOXQ AX, R10

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ DI, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, R11
	ADCXQ DI, AX
	MOVQ  R11, DI

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ SI, DI
	MULXQ q<>+8(SB), AX, SI
	ADOXQ AX, DI

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R8, SI
	MULXQ q<>+16(SB), AX, R8
	ADOXQ AX, SI

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BP, R8
	MULXQ q<>+24(SB), AX, BP
	ADOXQ AX, R8

	// t[3] = C + A
	MOVQ  $0, AX
	ADCXQ AX, BP
	ADOXQ R10, BP

	// reduce element(DI,SI,R8,BP) using temp registers (R12,R13,R11,R10)
	REDUCE(DI,SI,R8,BP,R12,R13,R11,R10)

	SUBQ    DI, R14
	SBBQ    SI, R15
	SBBQ    R8, CX
	SBBQ    BP, BX
	MOVQ    $0x3c208c16d87cfd47, R9
	MOVQ    $0x97816a916871ca8d, R12
	MOVQ    $0xb85045b68181585d, R13
	MOVQ    $0x30644e72e131a029, R11
	MOVQ    $0, R10
	CMOVQCC R10, R9
	CMOVQCC R10, R12
	CMOVQCC R10, R13
	CMOVQCC R10, R11
	ADDQ    R9, R14
	ADCQ    R12, R15
	ADCQ    R13, CX
	ADCQ    R11, BX
	PUSHQ   R14
	PUSHQ   R15
	PUSHQ   CX
	PUSHQ   BX

	// t[0] = R9
	// t[1] = R12
	// t[2] = R13
	// t[3] = R11

	// clear the flags
	XORQ AX, AX
	MOVQ y+16(FP), R10
	MOVQ 32(R10), DX

	// (A,t[0])  := t[0] + x[0]*y[0] + A
	MOVQ  x+8(FP), R10
	MULXQ 32(R10), R9, R12

	// (A,t[1])  := t[1] + x[1]*y[0] + A
	MOVQ  x+8(FP), R10
	MULXQ 40(R10), AX, R13
	ADOXQ AX, R12

	// (A,t[2])  := t[2] + x[2]*y[0] + A
	MOVQ  x+8(FP), R10
	MULXQ 48(R10), AX, R11
	ADOXQ AX, R13

	// (A,t[3])  := t[3] + x[3]*y[0] + A
	MOVQ  x+8(FP), R10
	MULXQ 56(R10), AX, R14
	ADOXQ AX, R11

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADOXQ AX, R14

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R9, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, R15
	ADCXQ R9, AX
	MOVQ  R15, R9

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R12, R9
	MULXQ q<>+8(SB), AX, R12
	ADOXQ AX, R9

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R13, R12
	MULXQ q<>+16(SB), AX, R13
	ADOXQ AX, R12

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ R11, R13
	MULXQ q<>+24(SB), AX, R11
	ADOXQ AX, R13

	// t[3] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R11
	ADOXQ R14, R11

	// clear the flags
	XORQ AX, AX
	MOVQ y+16(FP), R10
	MOVQ 40(R10), DX

	// (A,t[0])  := t[0] + x[0]*y[1] + A
	MOVQ  x+8(FP), R10
	MULXQ 32(R10), AX, R14
	ADOXQ AX, R9

	// (A,t[1])  := t[1] + x[1]*y[1] + A
	MOVQ  x+8(FP), R10
	ADCXQ R14, R12
	MULXQ 40(R10), AX, R14
	ADOXQ AX, R12

	// (A,t[2])  := t[2] + x[2]*y[1] + A
	MOVQ  x+8(FP), R10
	ADCXQ R14, R13
	MULXQ 48(R10), AX, R14
	ADOXQ AX, R13

	// (A,t[3])  := t[3] + x[3]*y[1] + A
	MOVQ  x+8(FP), R10
	ADCXQ R14, R11
	MULXQ 56(R10), AX, R14
	ADOXQ AX, R11

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R14
	ADOXQ AX, R14

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R9, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, CX
	ADCXQ R9, AX
	MOVQ  CX, R9

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R12, R9
	MULXQ q<>+8(SB), AX, R12
	ADOXQ AX, R9

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R13, R12
	MULXQ q<>+16(SB), AX, R13
	ADOXQ AX, R12

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ R11, R13
	MULXQ q<>+24(SB), AX, R11
	ADOXQ AX, R13

	// t[3] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R11
	ADOXQ R14, R11

	// clear the flags
	XORQ AX, AX
	MOVQ y+16(FP), R10
	MOVQ 48(R10), DX

	// (A,t[0])  := t[0] + x[0]*y[2] + A
	MOVQ  x+8(FP), R10
	MULXQ 32(R10), AX, R14
	ADOXQ AX, R9

	// (A,t[1])  := t[1] + x[1]*y[2] + A
	MOVQ  x+8(FP), R10
	ADCXQ R14, R12
	MULXQ 40(R10), AX, R14
	ADOXQ AX, R12

	// (A,t[2])  := t[2] + x[2]*y[2] + A
	MOVQ  x+8(FP), R10
	ADCXQ R14, R13
	MULXQ 48(R10), AX, R14
	ADOXQ AX, R13

	// (A,t[3])  := t[3] + x[3]*y[2] + A
	MOVQ  x+8(FP), R10
	ADCXQ R14, R11
	MULXQ 56(R10), AX, R14
	ADOXQ AX, R11

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R14
	ADOXQ AX, R14

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R9, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, BX
	ADCXQ R9, AX
	MOVQ  BX, R9

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R12, R9
	MULXQ q<>+8(SB), AX, R12
	ADOXQ AX, R9

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R13, R12
	MULXQ q<>+16(SB), AX, R13
	ADOXQ AX, R12

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ R11, R13
	MULXQ q<>+24(SB), AX, R11
	ADOXQ AX, R13

	// t[3] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R11
	ADOXQ R14, R11

	// clear the flags
	XORQ AX, AX
	MOVQ y+16(FP), R10
	MOVQ 56(R10), DX

	// (A,t[0])  := t[0] + x[0]*y[3] + A
	MOVQ  x+8(FP), R10
	MULXQ 32(R10), AX, R14
	ADOXQ AX, R9

	// (A,t[1])  := t[1] + x[1]*y[3] + A
	MOVQ  x+8(FP), R10
	ADCXQ R14, R12
	MULXQ 40(R10), AX, R14
	ADOXQ AX, R12

	// (A,t[2])  := t[2] + x[2]*y[3] + A
	MOVQ  x+8(FP), R10
	ADCXQ R14, R13
	MULXQ 48(R10), AX, R14
	ADOXQ AX, R13

	// (A,t[3])  := t[3] + x[3]*y[3] + A
	MOVQ  x+8(FP), R10
	ADCXQ R14, R11
	MULXQ 56(R10), AX, R14
	ADOXQ AX, R11

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R14
	ADOXQ AX, R14

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R9, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, R15
	ADCXQ R9, AX
	MOVQ  R15, R9

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R12, R9
	MULXQ q<>+8(SB), AX, R12
	ADOXQ AX, R9

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R13, R12
	MULXQ q<>+16(SB), AX, R13
	ADOXQ AX, R12

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ R11, R13
	MULXQ q<>+24(SB), AX, R11
	ADOXQ AX, R13

	// t[3] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R11
	ADOXQ R14, R11

	// reduce element(R9,R12,R13,R11) using temp registers (CX,BX,R15,R14)
	REDUCE(R9,R12,R13,R11,CX,BX,R15,R14)

	SUBQ    R9, DI
	SBBQ    R12, SI
	SBBQ    R13, R8
	SBBQ    R11, BP
	MOVQ    $0x3c208c16d87cfd47, R10
	MOVQ    $0x97816a916871ca8d, CX
	MOVQ    $0xb85045b68181585d, BX
	MOVQ    $0x30644e72e131a029, R15
	MOVQ    $0, R14
	CMOVQCC R14, R10
	CMOVQCC R14, CX
	CMOVQCC R14, BX
	CMOVQCC R14, R15
	ADDQ    R10, DI
	ADCQ    CX, SI
	ADCQ    BX, R8
	ADCQ    R15, BP
	MOVQ    res+0(FP), R14
	MOVQ    DI, 0(R14)
	MOVQ    SI, 8(R14)
	MOVQ    R8, 16(R14)
	MOVQ    BP, 24(R14)
	POPQ    BP
	POPQ    R8
	POPQ    SI
	POPQ    DI
	SUBQ    R9, DI
	SBBQ    R12, SI
	SBBQ    R13, R8
	SBBQ    R11, BP
	MOVQ    $0x3c208c16d87cfd47, R10
	MOVQ    $0x97816a916871ca8d, CX
	MOVQ    $0xb85045b68181585d, BX
	MOVQ    $0x30644e72e131a029, R15
	MOVQ    $0, R9
	CMOVQCC R9, R10
	CMOVQCC R9, CX
	CMOVQCC R9, BX
	CMOVQCC R9, R15
	ADDQ    R10, DI
	ADCQ    CX, SI
	ADCQ    BX, R8
	ADCQ    R15, BP
	MOVQ    DI, 32(R14)
	MOVQ    SI, 40(R14)
	MOVQ    R8, 48(R14)
	MOVQ    BP, 56(R14)
	RET

l4:
	MOVQ res+0(FP), AX
	MOVQ AX, (SP)
	MOVQ x+8(FP), AX
	MOVQ AX, 8(SP)
	MOVQ y+16(FP), AX
	MOVQ AX, 16(SP)
	CALL ·mulGenericE2(SB)
	RET

TEXT ·squareAdxE2(SB), $16-16
	NO_LOCAL_POINTERS
	CMPB ·supportAdx(SB), $1
	JNE  l5
	MOVQ x+8(FP), R9
	MOVQ 32(R9), R14
	MOVQ 40(R9), R15
	MOVQ 48(R9), CX
	MOVQ 56(R9), BX
	MOVQ 0(R9), BP
	MOVQ 8(R9), SI
	MOVQ 16(R9), DI
	MOVQ 24(R9), R8
	ADDQ BP, R14
	ADCQ SI, R15
	ADCQ DI, CX
	ADCQ R8, BX

	// reduce element(R14,R15,CX,BX) using temp registers (R10,R11,R12,R13)
	REDUCE(R14,R15,CX,BX,R10,R11,R12,R13)

	SUBQ    32(R9), BP
	SBBQ    40(R9), SI
	SBBQ    48(R9), DI
	SBBQ    56(R9), R8
	MOVQ    $0x3c208c16d87cfd47, R10
	MOVQ    $0x97816a916871ca8d, R11
	MOVQ    $0xb85045b68181585d, R12
	MOVQ    $0x30644e72e131a029, R13
	MOVQ    $0, R9
	CMOVQCC R9, R10
	CMOVQCC R9, R11
	CMOVQCC R9, R12
	CMOVQCC R9, R13
	ADDQ    R10, BP
	ADCQ    R11, SI
	ADCQ    R12, DI
	ADCQ    R13, R8

	// t[0] = R9
	// t[1] = R10
	// t[2] = R11
	// t[3] = R12

	// clear the flags
	XORQ AX, AX
	MOVQ BP, DX

	// (A,t[0])  := t[0] + x[0]*y[0] + A
	MULXQ R14, R9, R10

	// (A,t[1])  := t[1] + x[1]*y[0] + A
	MULXQ R15, AX, R11
	ADOXQ AX, R10

	// (A,t[2])  := t[2] + x[2]*y[0] + A
	MULXQ CX, AX, R12
	ADOXQ AX, R11

	// (A,t[3])  := t[3] + x[3]*y[0] + A
	MULXQ BX, AX, R13
	ADOXQ AX, R12

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADOXQ AX, R13

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R9, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, BP
	ADCXQ R9, AX
	MOVQ  BP, R9

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R10, R9
	MULXQ q<>+8(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R11, R10
	MULXQ q<>+16(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ R12, R11
	MULXQ q<>+24(SB), AX, R12
	ADOXQ AX, R11

	// t[3] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R12
	ADOXQ R13, R12

	// clear the flags
	XORQ AX, AX
	MOVQ SI, DX

	// (A,t[0])  := t[0] + x[0]*y[1] + A
	MULXQ R14, AX, R13
	ADOXQ AX, R9

	// (A,t[1])  := t[1] + x[1]*y[1] + A
	ADCXQ R13, R10
	MULXQ R15, AX, R13
	ADOXQ AX, R10

	// (A,t[2])  := t[2] + x[2]*y[1] + A
	ADCXQ R13, R11
	MULXQ CX, AX, R13
	ADOXQ AX, R11

	// (A,t[3])  := t[3] + x[3]*y[1] + A
	ADCXQ R13, R12
	MULXQ BX, AX, R13
	ADOXQ AX, R12

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R13
	ADOXQ AX, R13

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R9, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, BP
	ADCXQ R9, AX
	MOVQ  BP, R9

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R10, R9
	MULXQ q<>+8(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R11, R10
	MULXQ q<>+16(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ R12, R11
	MULXQ q<>+24(SB), AX, R12
	ADOXQ AX, R11

	// t[3] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R12
	ADOXQ R13, R12

	// clear the flags
	XORQ AX, AX
	MOVQ DI, DX

	// (A,t[0])  := t[0] + x[0]*y[2] + A
	MULXQ R14, AX, R13
	ADOXQ AX, R9

	// (A,t[1])  := t[1] + x[1]*y[2] + A
	ADCXQ R13, R10
	MULXQ R15, AX, R13
	ADOXQ AX, R10

	// (A,t[2])  := t[2] + x[2]*y[2] + A
	ADCXQ R13, R11
	MULXQ CX, AX, R13
	ADOXQ AX, R11

	// (A,t[3])  := t[3] + x[3]*y[2] + A
	ADCXQ R13, R12
	MULXQ BX, AX, R13
	ADOXQ AX, R12

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R13
	ADOXQ AX, R13

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R9, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, SI
	ADCXQ R9, AX
	MOVQ  SI, R9

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R10, R9
	MULXQ q<>+8(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R11, R10
	MULXQ q<>+16(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ R12, R11
	MULXQ q<>+24(SB), AX, R12
	ADOXQ AX, R11

	// t[3] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R12
	ADOXQ R13, R12

	// clear the flags
	XORQ AX, AX
	MOVQ R8, DX

	// (A,t[0])  := t[0] + x[0]*y[3] + A
	MULXQ R14, AX, R13
	ADOXQ AX, R9

	// (A,t[1])  := t[1] + x[1]*y[3] + A
	ADCXQ R13, R10
	MULXQ R15, AX, R13
	ADOXQ AX, R10

	// (A,t[2])  := t[2] + x[2]*y[3] + A
	ADCXQ R13, R11
	MULXQ CX, AX, R13
	ADOXQ AX, R11

	// (A,t[3])  := t[3] + x[3]*y[3] + A
	ADCXQ R13, R12
	MULXQ BX, AX, R13
	ADOXQ AX, R12

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R13
	ADOXQ AX, R13

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R9, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, BP
	ADCXQ R9, AX
	MOVQ  BP, R9

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R10, R9
	MULXQ q<>+8(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R11, R10
	MULXQ q<>+16(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ R12, R11
	MULXQ q<>+24(SB), AX, R12
	ADOXQ AX, R11

	// t[3] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R12
	ADOXQ R13, R12

	// reduce element(R9,R10,R11,R12) using temp registers (DI,SI,R8,BP)
	// stores in (R14,R15,CX,BX)
	REDUCE_AND_MOVE(R9,R10,R11,R12,DI,SI,R8,BP,R14,R15,CX,BX)

	MOVQ x+8(FP), R13

	// t[0] = DI
	// t[1] = SI
	// t[2] = R8
	// t[3] = BP

	// clear the flags
	XORQ AX, AX
	MOVQ 32(R13), DX

	// (A,t[0])  := t[0] + x[0]*y[0] + A
	MULXQ 0(R13), DI, SI

	// (A,t[1])  := t[1] + x[1]*y[0] + A
	MULXQ 8(R13), AX, R8
	ADOXQ AX, SI

	// (A,t[2])  := t[2] + x[2]*y[0] + A
	MULXQ 16(R13), AX, BP
	ADOXQ AX, R8

	// (A,t[3])  := t[3] + x[3]*y[0] + A
	MULXQ 24(R13), AX, R9
	ADOXQ AX, BP

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADOXQ AX, R9

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ DI, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, R10
	ADCXQ DI, AX
	MOVQ  R10, DI

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ SI, DI
	MULXQ q<>+8(SB), AX, SI
	ADOXQ AX, DI

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R8, SI
	MULXQ q<>+16(SB), AX, R8
	ADOXQ AX, SI

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BP, R8
	MULXQ q<>+24(SB), AX, BP
	ADOXQ AX, R8

	// t[3] = C + A
	MOVQ  $0, AX
	ADCXQ AX, BP
	ADOXQ R9, BP

	// clear the flags
	XORQ AX, AX
	MOVQ 40(R13), DX

	// (A,t[0])  := t[0] + x[0]*y[1] + A
	MULXQ 0(R13), AX, R9
	ADOXQ AX, DI

	// (A,t[1])  := t[1] + x[1]*y[1] + A
	ADCXQ R9, SI
	MULXQ 8(R13), AX, R9
	ADOXQ AX, SI

	// (A,t[2])  := t[2] + x[2]*y[1] + A
	ADCXQ R9, R8
	MULXQ 16(R13), AX, R9
	ADOXQ AX, R8

	// (A,t[3])  := t[3] + x[3]*y[1] + A
	ADCXQ R9, BP
	MULXQ 24(R13), AX, R9
	ADOXQ AX, BP

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R9
	ADOXQ AX, R9

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ DI, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, R11
	ADCXQ DI, AX
	MOVQ  R11, DI

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ SI, DI
	MULXQ q<>+8(SB), AX, SI
	ADOXQ AX, DI

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R8, SI
	MULXQ q<>+16(SB), AX, R8
	ADOXQ AX, SI

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BP, R8
	MULXQ q<>+24(SB), AX, BP
	ADOXQ AX, R8

	// t[3] = C + A
	MOVQ  $0, AX
	ADCXQ AX, BP
	ADOXQ R9, BP

	// clear the flags
	XORQ AX, AX
	MOVQ 48(R13), DX

	// (A,t[0])  := t[0] + x[0]*y[2] + A
	MULXQ 0(R13), AX, R9
	ADOXQ AX, DI

	// (A,t[1])  := t[1] + x[1]*y[2] + A
	ADCXQ R9, SI
	MULXQ 8(R13), AX, R9
	ADOXQ AX, SI

	// (A,t[2])  := t[2] + x[2]*y[2] + A
	ADCXQ R9, R8
	MULXQ 16(R13), AX, R9
	ADOXQ AX, R8

	// (A,t[3])  := t[3] + x[3]*y[2] + A
	ADCXQ R9, BP
	MULXQ 24(R13), AX, R9
	ADOXQ AX, BP

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R9
	ADOXQ AX, R9

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ DI, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, R12
	ADCXQ DI, AX
	MOVQ  R12, DI

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ SI, DI
	MULXQ q<>+8(SB), AX, SI
	ADOXQ AX, DI

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R8, SI
	MULXQ q<>+16(SB), AX, R8
	ADOXQ AX, SI

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BP, R8
	MULXQ q<>+24(SB), AX, BP
	ADOXQ AX, R8

	// t[3] = C + A
	MOVQ  $0, AX
	ADCXQ AX, BP
	ADOXQ R9, BP

	// clear the flags
	XORQ AX, AX
	MOVQ 56(R13), DX

	// (A,t[0])  := t[0] + x[0]*y[3] + A
	MULXQ 0(R13), AX, R9
	ADOXQ AX, DI

	// (A,t[1])  := t[1] + x[1]*y[3] + A
	ADCXQ R9, SI
	MULXQ 8(R13), AX, R9
	ADOXQ AX, SI

	// (A,t[2])  := t[2] + x[2]*y[3] + A
	ADCXQ R9, R8
	MULXQ 16(R13), AX, R9
	ADOXQ AX, R8

	// (A,t[3])  := t[3] + x[3]*y[3] + A
	ADCXQ R9, BP
	MULXQ 24(R13), AX, R9
	ADOXQ AX, BP

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R9
	ADOXQ AX, R9

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ DI, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, R10
	ADCXQ DI, AX
	MOVQ  R10, DI

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ SI, DI
	MULXQ q<>+8(SB), AX, SI
	ADOXQ AX, DI

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ R8, SI
	MULXQ q<>+16(SB), AX, R8
	ADOXQ AX, SI

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BP, R8
	MULXQ q<>+24(SB), AX, BP
	ADOXQ AX, R8

	// t[3] = C + A
	MOVQ  $0, AX
	ADCXQ AX, BP
	ADOXQ R9, BP

	// reduce element(DI,SI,R8,BP) using temp registers (R11,R12,R10,R9)
	REDUCE(DI,SI,R8,BP,R11,R12,R10,R9)

	ADDQ DI, DI
	ADCQ SI, SI
	ADCQ R8, R8
	ADCQ BP, BP
	MOVQ res+0(FP), R13

	// reduce element(DI,SI,R8,BP) using temp registers (R11,R12,R10,R9)
	// stores in (32(R13),40(R13),48(R13),56(R13))
	REDUCE_AND_MOVE(DI,SI,R8,BP,R11,R12,R10,R9,32(R13),40(R13),48(R13),56(R13))

	MOVQ R14, 0(R13)
	MOVQ R15, 8(R13)
	MOVQ CX, 16(R13)
	MOVQ BX, 24(R13)
	RET

l5:
	MOVQ res+0(FP), AX
	MOVQ AX, (SP)
	MOVQ x+8(FP), AX
	MOVQ AX, 8(SP)
	CALL ·squareGenericE2(SB)
	RET

TEXT ·mulNonResE2(SB), NOSPLIT, $0-16
	MOVQ x+8(FP), R9
	MOVQ 0(R9), AX
	MOVQ 8(R9), DX
	MOVQ 16(R9), CX
	MOVQ 24(R9), BX
	ADDQ AX, AX
	ADCQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX

	// reduce element(AX,DX,CX,BX) using temp registers (R10,R11,R12,R13)
	REDUCE(AX,DX,CX,BX,R10,R11,R12,R13)

	ADDQ AX, AX
	ADCQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX

	// reduce element(AX,DX,CX,BX) using temp registers (R14,R15,R10,R11)
	REDUCE(AX,DX,CX,BX,R14,R15,R10,R11)

	ADDQ AX, AX
	ADCQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX

	// reduce element(AX,DX,CX,BX) using temp registers (R12,R13,R14,R15)
	REDUCE(AX,DX,CX,BX,R12,R13,R14,R15)

	ADDQ 0(R9), AX
	ADCQ 8(R9), DX
	ADCQ 16(R9), CX
	ADCQ 24(R9), BX

	// reduce element(AX,DX,CX,BX) using temp registers (R10,R11,R12,R13)
	REDUCE(AX,DX,CX,BX,R10,R11,R12,R13)

	MOVQ    32(R9), BP
	MOVQ    40(R9), SI
	MOVQ    48(R9), DI
	MOVQ    56(R9), R8
	SUBQ    BP, AX
	SBBQ    SI, DX
	SBBQ    DI, CX
	SBBQ    R8, BX
	MOVQ    $0x3c208c16d87cfd47, R14
	MOVQ    $0x97816a916871ca8d, R15
	MOVQ    $0xb85045b68181585d, R10
	MOVQ    $0x30644e72e131a029, R11
	MOVQ    $0, R12
	CMOVQCC R12, R14
	CMOVQCC R12, R15
	CMOVQCC R12, R10
	CMOVQCC R12, R11
	ADDQ    R14, AX
	ADCQ    R15, DX
	ADCQ    R10, CX
	ADCQ    R11, BX
	ADDQ    BP, BP
	ADCQ    SI, SI
	ADCQ    DI, DI
	ADCQ    R8, R8

	// reduce element(BP,SI,DI,R8) using temp registers (R13,R12,R14,R15)
	REDUCE(BP,SI,DI,R8,R13,R12,R14,R15)

	ADDQ BP, BP
	ADCQ SI, SI
	ADCQ DI, DI
	ADCQ R8, R8

	// reduce element(BP,SI,DI,R8) using temp registers (R10,R11,R13,R12)
	REDUCE(BP,SI,DI,R8,R10,R11,R13,R12)

	ADDQ BP, BP
	ADCQ SI, SI
	ADCQ DI, DI
	ADCQ R8, R8

	// reduce element(BP,SI,DI,R8) using temp registers (R14,R15,R10,R11)
	REDUCE(BP,SI,DI,R8,R14,R15,R10,R11)

	ADDQ 32(R9), BP
	ADCQ 40(R9), SI
	ADCQ 48(R9), DI
	ADCQ 56(R9), R8

	// reduce element(BP,SI,DI,R8) using temp registers (R13,R12,R14,R15)
	REDUCE(BP,SI,DI,R8,R13,R12,R14,R15)

	ADDQ 0(R9), BP
	ADCQ 8(R9), SI
	ADCQ 16(R9), DI
	ADCQ 24(R9), R8

	// reduce element(BP,SI,DI,R8) using temp registers (R10,R11,R13,R12)
	REDUCE(BP,SI,DI,R8,R10,R11,R13,R12)

	MOVQ res+0(FP), R9
	MOVQ AX, 0(R9)
	MOVQ DX, 8(R9)
	MOVQ CX, 16(R9)
	MOVQ BX, 24(R9)
	MOVQ BP, 32(R9)
	MOVQ SI, 40(R9)
	MOVQ DI, 48(R9)
	MOVQ R8, 56(R9)
	RET
