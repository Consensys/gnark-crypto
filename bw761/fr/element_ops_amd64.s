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
DATA q<>+0(SB)/8, $0x8508c00000000001
DATA q<>+8(SB)/8, $0x170b5d4430000000
DATA q<>+16(SB)/8, $0x1ef3622fba094800
DATA q<>+24(SB)/8, $0x1a22d9f300f5138f
DATA q<>+32(SB)/8, $0xc63b05c06ca1493b
DATA q<>+40(SB)/8, $0x01ae3a4617c510ea
GLOBL q<>(SB), (RODATA+NOPTR), $48

// qInv0 q'[0]
DATA qInv0<>(SB)/8, $0x8508bfffffffffff
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

// add(res, x, y *Element)
TEXT ·add(SB), NOSPLIT, $0-24
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

	RET

// sub(res, x, y *Element)
TEXT ·sub(SB), NOSPLIT, $0-24
	MOVQ    x+8(FP), DI
	MOVQ    0(DI), AX
	MOVQ    8(DI), DX
	MOVQ    16(DI), CX
	MOVQ    24(DI), BX
	MOVQ    32(DI), BP
	MOVQ    40(DI), SI
	MOVQ    y+16(FP), R8
	SUBQ    0(R8), AX
	SBBQ    8(R8), DX
	SBBQ    16(R8), CX
	SBBQ    24(R8), BX
	SBBQ    32(R8), BP
	SBBQ    40(R8), SI
	MOVQ    $0x8508c00000000001, R9
	MOVQ    $0x170b5d4430000000, R10
	MOVQ    $0x1ef3622fba094800, R11
	MOVQ    $0x1a22d9f300f5138f, R12
	MOVQ    $0xc63b05c06ca1493b, R13
	MOVQ    $0x01ae3a4617c510ea, R14
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
	RET

// double(res, x *Element)
TEXT ·double(SB), NOSPLIT, $0-16
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

	RET

// neg(res, x *Element)
TEXT ·neg(SB), NOSPLIT, $0-16
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
	JEQ   l1
	MOVQ  $0x8508c00000000001, CX
	SUBQ  BX, CX
	MOVQ  CX, 0(DX)
	MOVQ  $0x170b5d4430000000, CX
	SBBQ  BP, CX
	MOVQ  CX, 8(DX)
	MOVQ  $0x1ef3622fba094800, CX
	SBBQ  SI, CX
	MOVQ  CX, 16(DX)
	MOVQ  $0x1a22d9f300f5138f, CX
	SBBQ  DI, CX
	MOVQ  CX, 24(DX)
	MOVQ  $0xc63b05c06ca1493b, CX
	SBBQ  R8, CX
	MOVQ  CX, 32(DX)
	MOVQ  $0x01ae3a4617c510ea, CX
	SBBQ  R9, CX
	MOVQ  CX, 40(DX)
	RET

l1:
	MOVQ AX, 0(DX)
	MOVQ AX, 8(DX)
	MOVQ AX, 16(DX)
	MOVQ AX, 24(DX)
	MOVQ AX, 32(DX)
	MOVQ AX, 40(DX)
	RET

// mul(res, x, y *Element)
TEXT ·mul(SB), $24-24

	// the algorithm is described here
	// https://hackmd.io/@zkteam/modular_multiplication
	// however, to benefit from the ADCX and ADOX carry chains
	// we split the inner loops in 2:
	// for i=0 to N-1
	// 		for j=0 to N-1
	// 		    (A,t[j])  := t[j] + x[j]*y[i] + A
	// 		m := t[0]*q'[0] mod W
	// 		C,_ := t[0] + m*q[0]
	// 		for j=1 to N-1
	// 		    (C,t[j-1]) := t[j] + m*q[j] + C
	// 		t[N-1] = C + A

	NO_LOCAL_POINTERS
	CMPB ·supportAdx(SB), $1
	JNE  l2
	MOVQ x+8(FP), R14
	MOVQ y+16(FP), R15

	// t[0] = CX
	// t[1] = BX
	// t[2] = BP
	// t[3] = SI
	// t[4] = DI
	// t[5] = R8

	// clear the flags
	XORQ AX, AX
	MOVQ 0(R15), DX

	// (A,t[0])  := t[0] + x[0]*y[0] + A
	MULXQ 0(R14), CX, BX

	// (A,t[1])  := t[1] + x[1]*y[0] + A
	MULXQ 8(R14), AX, BP
	ADOXQ AX, BX

	// (A,t[2])  := t[2] + x[2]*y[0] + A
	MULXQ 16(R14), AX, SI
	ADOXQ AX, BP

	// (A,t[3])  := t[3] + x[3]*y[0] + A
	MULXQ 24(R14), AX, DI
	ADOXQ AX, SI

	// (A,t[4])  := t[4] + x[4]*y[0] + A
	MULXQ 32(R14), AX, R8
	ADOXQ AX, DI

	// (A,t[5])  := t[5] + x[5]*y[0] + A
	MULXQ 40(R14), AX, R9
	ADOXQ AX, R8

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADOXQ AX, R9

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ CX, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, R10
	ADCXQ CX, AX
	MOVQ  R10, CX

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ BX, CX
	MULXQ q<>+8(SB), AX, BX
	ADOXQ AX, CX

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ BP, BX
	MULXQ q<>+16(SB), AX, BP
	ADOXQ AX, BX

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ SI, BP
	MULXQ q<>+24(SB), AX, SI
	ADOXQ AX, BP

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ DI, SI
	MULXQ q<>+32(SB), AX, DI
	ADOXQ AX, SI

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
	MOVQ 8(R15), DX

	// (A,t[0])  := t[0] + x[0]*y[1] + A
	MULXQ 0(R14), AX, R9
	ADOXQ AX, CX

	// (A,t[1])  := t[1] + x[1]*y[1] + A
	ADCXQ R9, BX
	MULXQ 8(R14), AX, R9
	ADOXQ AX, BX

	// (A,t[2])  := t[2] + x[2]*y[1] + A
	ADCXQ R9, BP
	MULXQ 16(R14), AX, R9
	ADOXQ AX, BP

	// (A,t[3])  := t[3] + x[3]*y[1] + A
	ADCXQ R9, SI
	MULXQ 24(R14), AX, R9
	ADOXQ AX, SI

	// (A,t[4])  := t[4] + x[4]*y[1] + A
	ADCXQ R9, DI
	MULXQ 32(R14), AX, R9
	ADOXQ AX, DI

	// (A,t[5])  := t[5] + x[5]*y[1] + A
	ADCXQ R9, R8
	MULXQ 40(R14), AX, R9
	ADOXQ AX, R8

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R9
	ADOXQ AX, R9

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ CX, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, R11
	ADCXQ CX, AX
	MOVQ  R11, CX

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ BX, CX
	MULXQ q<>+8(SB), AX, BX
	ADOXQ AX, CX

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ BP, BX
	MULXQ q<>+16(SB), AX, BP
	ADOXQ AX, BX

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ SI, BP
	MULXQ q<>+24(SB), AX, SI
	ADOXQ AX, BP

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ DI, SI
	MULXQ q<>+32(SB), AX, DI
	ADOXQ AX, SI

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
	MOVQ 16(R15), DX

	// (A,t[0])  := t[0] + x[0]*y[2] + A
	MULXQ 0(R14), AX, R9
	ADOXQ AX, CX

	// (A,t[1])  := t[1] + x[1]*y[2] + A
	ADCXQ R9, BX
	MULXQ 8(R14), AX, R9
	ADOXQ AX, BX

	// (A,t[2])  := t[2] + x[2]*y[2] + A
	ADCXQ R9, BP
	MULXQ 16(R14), AX, R9
	ADOXQ AX, BP

	// (A,t[3])  := t[3] + x[3]*y[2] + A
	ADCXQ R9, SI
	MULXQ 24(R14), AX, R9
	ADOXQ AX, SI

	// (A,t[4])  := t[4] + x[4]*y[2] + A
	ADCXQ R9, DI
	MULXQ 32(R14), AX, R9
	ADOXQ AX, DI

	// (A,t[5])  := t[5] + x[5]*y[2] + A
	ADCXQ R9, R8
	MULXQ 40(R14), AX, R9
	ADOXQ AX, R8

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R9
	ADOXQ AX, R9

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ CX, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, R12
	ADCXQ CX, AX
	MOVQ  R12, CX

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ BX, CX
	MULXQ q<>+8(SB), AX, BX
	ADOXQ AX, CX

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ BP, BX
	MULXQ q<>+16(SB), AX, BP
	ADOXQ AX, BX

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ SI, BP
	MULXQ q<>+24(SB), AX, SI
	ADOXQ AX, BP

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ DI, SI
	MULXQ q<>+32(SB), AX, DI
	ADOXQ AX, SI

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
	MOVQ 24(R15), DX

	// (A,t[0])  := t[0] + x[0]*y[3] + A
	MULXQ 0(R14), AX, R9
	ADOXQ AX, CX

	// (A,t[1])  := t[1] + x[1]*y[3] + A
	ADCXQ R9, BX
	MULXQ 8(R14), AX, R9
	ADOXQ AX, BX

	// (A,t[2])  := t[2] + x[2]*y[3] + A
	ADCXQ R9, BP
	MULXQ 16(R14), AX, R9
	ADOXQ AX, BP

	// (A,t[3])  := t[3] + x[3]*y[3] + A
	ADCXQ R9, SI
	MULXQ 24(R14), AX, R9
	ADOXQ AX, SI

	// (A,t[4])  := t[4] + x[4]*y[3] + A
	ADCXQ R9, DI
	MULXQ 32(R14), AX, R9
	ADOXQ AX, DI

	// (A,t[5])  := t[5] + x[5]*y[3] + A
	ADCXQ R9, R8
	MULXQ 40(R14), AX, R9
	ADOXQ AX, R8

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R9
	ADOXQ AX, R9

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ CX, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, R13
	ADCXQ CX, AX
	MOVQ  R13, CX

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ BX, CX
	MULXQ q<>+8(SB), AX, BX
	ADOXQ AX, CX

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ BP, BX
	MULXQ q<>+16(SB), AX, BP
	ADOXQ AX, BX

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ SI, BP
	MULXQ q<>+24(SB), AX, SI
	ADOXQ AX, BP

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ DI, SI
	MULXQ q<>+32(SB), AX, DI
	ADOXQ AX, SI

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
	MOVQ 32(R15), DX

	// (A,t[0])  := t[0] + x[0]*y[4] + A
	MULXQ 0(R14), AX, R9
	ADOXQ AX, CX

	// (A,t[1])  := t[1] + x[1]*y[4] + A
	ADCXQ R9, BX
	MULXQ 8(R14), AX, R9
	ADOXQ AX, BX

	// (A,t[2])  := t[2] + x[2]*y[4] + A
	ADCXQ R9, BP
	MULXQ 16(R14), AX, R9
	ADOXQ AX, BP

	// (A,t[3])  := t[3] + x[3]*y[4] + A
	ADCXQ R9, SI
	MULXQ 24(R14), AX, R9
	ADOXQ AX, SI

	// (A,t[4])  := t[4] + x[4]*y[4] + A
	ADCXQ R9, DI
	MULXQ 32(R14), AX, R9
	ADOXQ AX, DI

	// (A,t[5])  := t[5] + x[5]*y[4] + A
	ADCXQ R9, R8
	MULXQ 40(R14), AX, R9
	ADOXQ AX, R8

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R9
	ADOXQ AX, R9

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ CX, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, R10
	ADCXQ CX, AX
	MOVQ  R10, CX

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ BX, CX
	MULXQ q<>+8(SB), AX, BX
	ADOXQ AX, CX

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ BP, BX
	MULXQ q<>+16(SB), AX, BP
	ADOXQ AX, BX

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ SI, BP
	MULXQ q<>+24(SB), AX, SI
	ADOXQ AX, BP

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ DI, SI
	MULXQ q<>+32(SB), AX, DI
	ADOXQ AX, SI

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
	MOVQ 40(R15), DX

	// (A,t[0])  := t[0] + x[0]*y[5] + A
	MULXQ 0(R14), AX, R9
	ADOXQ AX, CX

	// (A,t[1])  := t[1] + x[1]*y[5] + A
	ADCXQ R9, BX
	MULXQ 8(R14), AX, R9
	ADOXQ AX, BX

	// (A,t[2])  := t[2] + x[2]*y[5] + A
	ADCXQ R9, BP
	MULXQ 16(R14), AX, R9
	ADOXQ AX, BP

	// (A,t[3])  := t[3] + x[3]*y[5] + A
	ADCXQ R9, SI
	MULXQ 24(R14), AX, R9
	ADOXQ AX, SI

	// (A,t[4])  := t[4] + x[4]*y[5] + A
	ADCXQ R9, DI
	MULXQ 32(R14), AX, R9
	ADOXQ AX, DI

	// (A,t[5])  := t[5] + x[5]*y[5] + A
	ADCXQ R9, R8
	MULXQ 40(R14), AX, R9
	ADOXQ AX, R8

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, R9
	ADOXQ AX, R9

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ CX, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, R11
	ADCXQ CX, AX
	MOVQ  R11, CX

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ BX, CX
	MULXQ q<>+8(SB), AX, BX
	ADOXQ AX, CX

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ BP, BX
	MULXQ q<>+16(SB), AX, BP
	ADOXQ AX, BX

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ SI, BP
	MULXQ q<>+24(SB), AX, SI
	ADOXQ AX, BP

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ DI, SI
	MULXQ q<>+32(SB), AX, DI
	ADOXQ AX, SI

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ R8, DI
	MULXQ q<>+40(SB), AX, R8
	ADOXQ AX, DI

	// t[5] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R8
	ADOXQ R9, R8
	MOVQ  res+0(FP), R12

	// reduce element(CX,BX,BP,SI,DI,R8) using temp registers (R13,R10,R11,R9,R14,R15)
	// stores in (0(R12),8(R12),16(R12),24(R12),32(R12),40(R12))
	REDUCE_AND_MOVE(CX,BX,BP,SI,DI,R8,R13,R10,R11,R9,R14,R15,0(R12),8(R12),16(R12),24(R12),32(R12),40(R12))

	RET

l2:
	MOVQ res+0(FP), AX
	MOVQ AX, (SP)
	MOVQ x+8(FP), AX
	MOVQ AX, 8(SP)
	MOVQ y+16(FP), AX
	MOVQ AX, 16(SP)
	CALL ·_mulGeneric(SB)
	RET

TEXT ·fromMont(SB), $8-8
	NO_LOCAL_POINTERS

	// the algorithm is described here
	// https://hackmd.io/@zkteam/modular_multiplication
	// when y = 1 we have:
	// for i=0 to N-1
	// 		t[i] = x[i]
	// for i=0 to N-1
	// 		m := t[0]*q'[0] mod W
	// 		C,_ := t[0] + m*q[0]
	// 		for j=1 to N-1
	// 		    (C,t[j-1]) := t[j] + m*q[j] + C
	// 		t[N-1] = C
	CMPB ·supportAdx(SB), $1
	JNE  l3
	MOVQ res+0(FP), DI
	MOVQ 0(DI), R14
	MOVQ 8(DI), R15
	MOVQ 16(DI), CX
	MOVQ 24(DI), BX
	MOVQ 32(DI), BP
	MOVQ 40(DI), SI
	XORQ DX, DX

	// m := t[0]*q'[0] mod W
	MOVQ  R14, DX
	MULXQ qInv0<>(SB), DX, AX
	XORQ  AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, R8
	ADCXQ R14, AX
	MOVQ  R8, R14

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R15, R14
	MULXQ q<>+8(SB), AX, R15
	ADOXQ AX, R14

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ CX, R15
	MULXQ q<>+16(SB), AX, CX
	ADOXQ AX, R15

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BX, CX
	MULXQ q<>+24(SB), AX, BX
	ADOXQ AX, CX

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ BP, BX
	MULXQ q<>+32(SB), AX, BP
	ADOXQ AX, BX

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ SI, BP
	MULXQ q<>+40(SB), AX, SI
	ADOXQ AX, BP
	MOVQ  $0, AX
	ADCXQ AX, SI
	ADOXQ AX, SI
	XORQ  DX, DX

	// m := t[0]*q'[0] mod W
	MOVQ  R14, DX
	MULXQ qInv0<>(SB), DX, AX
	XORQ  AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, R8
	ADCXQ R14, AX
	MOVQ  R8, R14

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R15, R14
	MULXQ q<>+8(SB), AX, R15
	ADOXQ AX, R14

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ CX, R15
	MULXQ q<>+16(SB), AX, CX
	ADOXQ AX, R15

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BX, CX
	MULXQ q<>+24(SB), AX, BX
	ADOXQ AX, CX

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ BP, BX
	MULXQ q<>+32(SB), AX, BP
	ADOXQ AX, BX

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ SI, BP
	MULXQ q<>+40(SB), AX, SI
	ADOXQ AX, BP
	MOVQ  $0, AX
	ADCXQ AX, SI
	ADOXQ AX, SI
	XORQ  DX, DX

	// m := t[0]*q'[0] mod W
	MOVQ  R14, DX
	MULXQ qInv0<>(SB), DX, AX
	XORQ  AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, R8
	ADCXQ R14, AX
	MOVQ  R8, R14

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R15, R14
	MULXQ q<>+8(SB), AX, R15
	ADOXQ AX, R14

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ CX, R15
	MULXQ q<>+16(SB), AX, CX
	ADOXQ AX, R15

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BX, CX
	MULXQ q<>+24(SB), AX, BX
	ADOXQ AX, CX

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ BP, BX
	MULXQ q<>+32(SB), AX, BP
	ADOXQ AX, BX

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ SI, BP
	MULXQ q<>+40(SB), AX, SI
	ADOXQ AX, BP
	MOVQ  $0, AX
	ADCXQ AX, SI
	ADOXQ AX, SI
	XORQ  DX, DX

	// m := t[0]*q'[0] mod W
	MOVQ  R14, DX
	MULXQ qInv0<>(SB), DX, AX
	XORQ  AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, R8
	ADCXQ R14, AX
	MOVQ  R8, R14

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R15, R14
	MULXQ q<>+8(SB), AX, R15
	ADOXQ AX, R14

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ CX, R15
	MULXQ q<>+16(SB), AX, CX
	ADOXQ AX, R15

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BX, CX
	MULXQ q<>+24(SB), AX, BX
	ADOXQ AX, CX

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ BP, BX
	MULXQ q<>+32(SB), AX, BP
	ADOXQ AX, BX

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ SI, BP
	MULXQ q<>+40(SB), AX, SI
	ADOXQ AX, BP
	MOVQ  $0, AX
	ADCXQ AX, SI
	ADOXQ AX, SI
	XORQ  DX, DX

	// m := t[0]*q'[0] mod W
	MOVQ  R14, DX
	MULXQ qInv0<>(SB), DX, AX
	XORQ  AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, R8
	ADCXQ R14, AX
	MOVQ  R8, R14

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R15, R14
	MULXQ q<>+8(SB), AX, R15
	ADOXQ AX, R14

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ CX, R15
	MULXQ q<>+16(SB), AX, CX
	ADOXQ AX, R15

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BX, CX
	MULXQ q<>+24(SB), AX, BX
	ADOXQ AX, CX

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ BP, BX
	MULXQ q<>+32(SB), AX, BP
	ADOXQ AX, BX

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ SI, BP
	MULXQ q<>+40(SB), AX, SI
	ADOXQ AX, BP
	MOVQ  $0, AX
	ADCXQ AX, SI
	ADOXQ AX, SI
	XORQ  DX, DX

	// m := t[0]*q'[0] mod W
	MOVQ  R14, DX
	MULXQ qInv0<>(SB), DX, AX
	XORQ  AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, R8
	ADCXQ R14, AX
	MOVQ  R8, R14

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R15, R14
	MULXQ q<>+8(SB), AX, R15
	ADOXQ AX, R14

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ CX, R15
	MULXQ q<>+16(SB), AX, CX
	ADOXQ AX, R15

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BX, CX
	MULXQ q<>+24(SB), AX, BX
	ADOXQ AX, CX

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ BP, BX
	MULXQ q<>+32(SB), AX, BP
	ADOXQ AX, BX

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ SI, BP
	MULXQ q<>+40(SB), AX, SI
	ADOXQ AX, BP
	MOVQ  $0, AX
	ADCXQ AX, SI
	ADOXQ AX, SI

	// reduce element(R14,R15,CX,BX,BP,SI) using temp registers (R9,R10,R11,R12,R13,R8)
	// stores in (0(DI),8(DI),16(DI),24(DI),32(DI),40(DI))
	REDUCE_AND_MOVE(R14,R15,CX,BX,BP,SI,R9,R10,R11,R12,R13,R8,0(DI),8(DI),16(DI),24(DI),32(DI),40(DI))

	RET

l3:
	MOVQ res+0(FP), AX
	MOVQ AX, (SP)
	CALL ·_fromMontGeneric(SB)
	RET

TEXT ·reduce(SB), NOSPLIT, $0-8
	MOVQ res+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), BP
	MOVQ 32(AX), SI
	MOVQ 40(AX), DI

	// reduce element(DX,CX,BX,BP,SI,DI) using temp registers (R8,R9,R10,R11,R12,R13)
	// stores in (0(AX),8(AX),16(AX),24(AX),32(AX),40(AX))
	REDUCE_AND_MOVE(DX,CX,BX,BP,SI,DI,R8,R9,R10,R11,R12,R13,0(AX),8(AX),16(AX),24(AX),32(AX),40(AX))

	RET

// MulBy3(x *Element)
TEXT ·MulBy3(SB), NOSPLIT, $0-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), BP
	MOVQ 32(AX), SI
	MOVQ 40(AX), DI
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ BP, BP
	ADCQ SI, SI
	ADCQ DI, DI

	// reduce element(DX,CX,BX,BP,SI,DI) using temp registers (R8,R9,R10,R11,R12,R13)
	REDUCE(DX,CX,BX,BP,SI,DI,R8,R9,R10,R11,R12,R13)

	ADDQ 0(AX), DX
	ADCQ 8(AX), CX
	ADCQ 16(AX), BX
	ADCQ 24(AX), BP
	ADCQ 32(AX), SI
	ADCQ 40(AX), DI

	// reduce element(DX,CX,BX,BP,SI,DI) using temp registers (R14,R15,R8,R9,R10,R11)
	// stores in (0(AX),8(AX),16(AX),24(AX),32(AX),40(AX))
	REDUCE_AND_MOVE(DX,CX,BX,BP,SI,DI,R14,R15,R8,R9,R10,R11,0(AX),8(AX),16(AX),24(AX),32(AX),40(AX))

	RET

// MulBy5(x *Element)
TEXT ·MulBy5(SB), NOSPLIT, $0-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), BP
	MOVQ 32(AX), SI
	MOVQ 40(AX), DI
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ BP, BP
	ADCQ SI, SI
	ADCQ DI, DI

	// reduce element(DX,CX,BX,BP,SI,DI) using temp registers (R8,R9,R10,R11,R12,R13)
	REDUCE(DX,CX,BX,BP,SI,DI,R8,R9,R10,R11,R12,R13)

	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ BP, BP
	ADCQ SI, SI
	ADCQ DI, DI

	// reduce element(DX,CX,BX,BP,SI,DI) using temp registers (R14,R15,R8,R9,R10,R11)
	REDUCE(DX,CX,BX,BP,SI,DI,R14,R15,R8,R9,R10,R11)

	ADDQ 0(AX), DX
	ADCQ 8(AX), CX
	ADCQ 16(AX), BX
	ADCQ 24(AX), BP
	ADCQ 32(AX), SI
	ADCQ 40(AX), DI

	// reduce element(DX,CX,BX,BP,SI,DI) using temp registers (R12,R13,R14,R15,R8,R9)
	// stores in (0(AX),8(AX),16(AX),24(AX),32(AX),40(AX))
	REDUCE_AND_MOVE(DX,CX,BX,BP,SI,DI,R12,R13,R14,R15,R8,R9,0(AX),8(AX),16(AX),24(AX),32(AX),40(AX))

	RET
