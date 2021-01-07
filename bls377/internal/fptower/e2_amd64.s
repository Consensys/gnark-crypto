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
	MOVQ $0x8508c00000000001, CX
	SUBQ BX, CX
	MOVQ CX, 0(DX)
	MOVQ $0x170b5d4430000000, CX
	SBBQ BP, CX
	MOVQ CX, 8(DX)
	MOVQ $0x1ef3622fba094800, CX
	SBBQ SI, CX
	MOVQ CX, 16(DX)
	MOVQ $0x1a22d9f300f5138f, CX
	SBBQ DI, CX
	MOVQ CX, 24(DX)
	MOVQ $0xc63b05c06ca1493b, CX
	SBBQ R8, CX
	MOVQ CX, 32(DX)
	MOVQ $0x01ae3a4617c510ea, CX
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
	MOVQ $0x8508c00000000001, CX
	SUBQ BX, CX
	MOVQ CX, 48(DX)
	MOVQ $0x170b5d4430000000, CX
	SBBQ BP, CX
	MOVQ CX, 56(DX)
	MOVQ $0x1ef3622fba094800, CX
	SBBQ SI, CX
	MOVQ CX, 64(DX)
	MOVQ $0x1a22d9f300f5138f, CX
	SBBQ DI, CX
	MOVQ CX, 72(DX)
	MOVQ $0xc63b05c06ca1493b, CX
	SBBQ R8, CX
	MOVQ CX, 80(DX)
	MOVQ $0x01ae3a4617c510ea, CX
	SBBQ R9, CX
	MOVQ CX, 88(DX)
	RET
