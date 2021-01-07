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

// add(res, x, y *Element)
TEXT ·add(SB), NOSPLIT, $0-24
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

	RET

// sub(res, x, y *Element)
TEXT ·sub(SB), NOSPLIT, $0-24
	MOVQ    x+8(FP), BP
	MOVQ    0(BP), AX
	MOVQ    8(BP), DX
	MOVQ    16(BP), CX
	MOVQ    24(BP), BX
	MOVQ    y+16(FP), SI
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
	RET

// double(res, x *Element)
TEXT ·double(SB), NOSPLIT, $0-16
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

	RET

// neg(res, x *Element)
TEXT ·neg(SB), NOSPLIT, $0-16
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
	JEQ   l1
	MOVQ  $0x3c208c16d87cfd47, CX
	SUBQ  BX, CX
	MOVQ  CX, 0(DX)
	MOVQ  $0x97816a916871ca8d, CX
	SBBQ  BP, CX
	MOVQ  CX, 8(DX)
	MOVQ  $0xb85045b68181585d, CX
	SBBQ  SI, CX
	MOVQ  CX, 16(DX)
	MOVQ  $0x30644e72e131a029, CX
	SBBQ  DI, CX
	MOVQ  CX, 24(DX)
	RET

l1:
	MOVQ AX, 0(DX)
	MOVQ AX, 8(DX)
	MOVQ AX, 16(DX)
	MOVQ AX, 24(DX)
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

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADOXQ AX, DI

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ CX, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, R8
	ADCXQ CX, AX
	MOVQ  R8, CX

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

	// t[3] = C + A
	MOVQ  $0, AX
	ADCXQ AX, SI
	ADOXQ DI, SI

	// clear the flags
	XORQ AX, AX
	MOVQ 8(R15), DX

	// (A,t[0])  := t[0] + x[0]*y[1] + A
	MULXQ 0(R14), AX, DI
	ADOXQ AX, CX

	// (A,t[1])  := t[1] + x[1]*y[1] + A
	ADCXQ DI, BX
	MULXQ 8(R14), AX, DI
	ADOXQ AX, BX

	// (A,t[2])  := t[2] + x[2]*y[1] + A
	ADCXQ DI, BP
	MULXQ 16(R14), AX, DI
	ADOXQ AX, BP

	// (A,t[3])  := t[3] + x[3]*y[1] + A
	ADCXQ DI, SI
	MULXQ 24(R14), AX, DI
	ADOXQ AX, SI

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, DI
	ADOXQ AX, DI

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ CX, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, R9
	ADCXQ CX, AX
	MOVQ  R9, CX

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

	// t[3] = C + A
	MOVQ  $0, AX
	ADCXQ AX, SI
	ADOXQ DI, SI

	// clear the flags
	XORQ AX, AX
	MOVQ 16(R15), DX

	// (A,t[0])  := t[0] + x[0]*y[2] + A
	MULXQ 0(R14), AX, DI
	ADOXQ AX, CX

	// (A,t[1])  := t[1] + x[1]*y[2] + A
	ADCXQ DI, BX
	MULXQ 8(R14), AX, DI
	ADOXQ AX, BX

	// (A,t[2])  := t[2] + x[2]*y[2] + A
	ADCXQ DI, BP
	MULXQ 16(R14), AX, DI
	ADOXQ AX, BP

	// (A,t[3])  := t[3] + x[3]*y[2] + A
	ADCXQ DI, SI
	MULXQ 24(R14), AX, DI
	ADOXQ AX, SI

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, DI
	ADOXQ AX, DI

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

	// t[3] = C + A
	MOVQ  $0, AX
	ADCXQ AX, SI
	ADOXQ DI, SI

	// clear the flags
	XORQ AX, AX
	MOVQ 24(R15), DX

	// (A,t[0])  := t[0] + x[0]*y[3] + A
	MULXQ 0(R14), AX, DI
	ADOXQ AX, CX

	// (A,t[1])  := t[1] + x[1]*y[3] + A
	ADCXQ DI, BX
	MULXQ 8(R14), AX, DI
	ADOXQ AX, BX

	// (A,t[2])  := t[2] + x[2]*y[3] + A
	ADCXQ DI, BP
	MULXQ 16(R14), AX, DI
	ADOXQ AX, BP

	// (A,t[3])  := t[3] + x[3]*y[3] + A
	ADCXQ DI, SI
	MULXQ 24(R14), AX, DI
	ADOXQ AX, SI

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, DI
	ADOXQ AX, DI

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

	// t[3] = C + A
	MOVQ  $0, AX
	ADCXQ AX, SI
	ADOXQ DI, SI
	MOVQ  res+0(FP), R12

	// reduce element(CX,BX,BP,SI) using temp registers (R13,R8,R9,R10)
	// stores in (0(R12),8(R12),16(R12),24(R12))
	REDUCE_AND_MOVE(CX,BX,BP,SI,R13,R8,R9,R10,0(R12),8(R12),16(R12),24(R12))

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
	MOVQ res+0(FP), BP
	MOVQ 0(BP), R14
	MOVQ 8(BP), R15
	MOVQ 16(BP), CX
	MOVQ 24(BP), BX
	XORQ DX, DX

	// m := t[0]*q'[0] mod W
	MOVQ  R14, DX
	MULXQ qInv0<>(SB), DX, AX
	XORQ  AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, SI
	ADCXQ R14, AX
	MOVQ  SI, R14

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
	MOVQ  $0, AX
	ADCXQ AX, BX
	ADOXQ AX, BX
	XORQ  DX, DX

	// m := t[0]*q'[0] mod W
	MOVQ  R14, DX
	MULXQ qInv0<>(SB), DX, AX
	XORQ  AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, SI
	ADCXQ R14, AX
	MOVQ  SI, R14

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
	MOVQ  $0, AX
	ADCXQ AX, BX
	ADOXQ AX, BX
	XORQ  DX, DX

	// m := t[0]*q'[0] mod W
	MOVQ  R14, DX
	MULXQ qInv0<>(SB), DX, AX
	XORQ  AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, SI
	ADCXQ R14, AX
	MOVQ  SI, R14

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
	MOVQ  $0, AX
	ADCXQ AX, BX
	ADOXQ AX, BX
	XORQ  DX, DX

	// m := t[0]*q'[0] mod W
	MOVQ  R14, DX
	MULXQ qInv0<>(SB), DX, AX
	XORQ  AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, SI
	ADCXQ R14, AX
	MOVQ  SI, R14

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
	MOVQ  $0, AX
	ADCXQ AX, BX
	ADOXQ AX, BX

	// reduce element(R14,R15,CX,BX) using temp registers (DI,R8,R9,R10)
	// stores in (0(BP),8(BP),16(BP),24(BP))
	REDUCE_AND_MOVE(R14,R15,CX,BX,DI,R8,R9,R10,0(BP),8(BP),16(BP),24(BP))

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

	// reduce element(DX,CX,BX,BP) using temp registers (SI,DI,R8,R9)
	// stores in (0(AX),8(AX),16(AX),24(AX))
	REDUCE_AND_MOVE(DX,CX,BX,BP,SI,DI,R8,R9,0(AX),8(AX),16(AX),24(AX))

	RET

// MulBy3(x *Element)
TEXT ·MulBy3(SB), NOSPLIT, $0-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), BP
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ BP, BP

	// reduce element(DX,CX,BX,BP) using temp registers (SI,DI,R8,R9)
	REDUCE(DX,CX,BX,BP,SI,DI,R8,R9)

	ADDQ 0(AX), DX
	ADCQ 8(AX), CX
	ADCQ 16(AX), BX
	ADCQ 24(AX), BP

	// reduce element(DX,CX,BX,BP) using temp registers (R10,R11,R12,R13)
	// stores in (0(AX),8(AX),16(AX),24(AX))
	REDUCE_AND_MOVE(DX,CX,BX,BP,R10,R11,R12,R13,0(AX),8(AX),16(AX),24(AX))

	RET

// MulBy5(x *Element)
TEXT ·MulBy5(SB), NOSPLIT, $0-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), BP
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ BP, BP

	// reduce element(DX,CX,BX,BP) using temp registers (SI,DI,R8,R9)
	REDUCE(DX,CX,BX,BP,SI,DI,R8,R9)

	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ BP, BP

	// reduce element(DX,CX,BX,BP) using temp registers (R10,R11,R12,R13)
	REDUCE(DX,CX,BX,BP,R10,R11,R12,R13)

	ADDQ 0(AX), DX
	ADCQ 8(AX), CX
	ADCQ 16(AX), BX
	ADCQ 24(AX), BP

	// reduce element(DX,CX,BX,BP) using temp registers (R14,R15,SI,DI)
	// stores in (0(AX),8(AX),16(AX),24(AX))
	REDUCE_AND_MOVE(DX,CX,BX,BP,R14,R15,SI,DI,0(AX),8(AX),16(AX),24(AX))

	RET
