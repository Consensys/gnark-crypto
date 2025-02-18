// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

#include "textflag.h"
#include "funcdata.h"
#include "go_asm.h"

#define REDUCE(ra0, rb0) \
	MOVQ    ra0, rb0;           \
	SUBQ    ·qElement(SB), ra0; \
	CMOVQCS rb0, ra0;           \

// this code is generated and identical to fr.Mul(...)
// A -> BP
// t -> CX
#define MACC(in0, in1, in2) \
	ADCXQ in0, in1     \
	MULXQ in2, AX, in0 \
	ADOXQ AX, in1      \

#define DIV_SHIFT() \
	MOVQ  $const_qInvNeg, DX      \
	IMULQ CX, DX                  \
	XORQ  AX, AX                  \
	MULXQ ·qElement+0(SB), AX, BX \
	ADCXQ CX, AX                  \
	MOVQ  BX, CX                  \
	MOVQ  $0, AX                  \
	ADCXQ AX, CX                  \
	ADOXQ BP, CX                  \

#define MUL_WORD_0() \
	XORQ  AX, AX      \
	MULXQ R14, CX, BP \
	MOVQ  $0, AX      \
	ADOXQ AX, BP      \
	DIV_SHIFT()       \

#define MUL_WORD_N() \
	XORQ  AX, AX      \
	MULXQ R14, AX, BP \
	ADOXQ AX, CX      \
	MOVQ  $0, AX      \
	ADCXQ AX, BP      \
	ADOXQ AX, BP      \
	DIV_SHIFT()       \

#define MUL() \
	MOVQ R15, DX; \
	MUL_WORD_0(); \

TEXT ·addE2(SB), NOSPLIT, $0-24
	MOVQ x+8(FP), AX
	MOVQ 0(AX), BX
	MOVQ y+16(FP), DX
	ADDQ 0(DX), BX

	// reduce element(BX) using temp registers (SI)
	REDUCE(BX,SI)

	MOVQ res+0(FP), CX
	MOVQ BX, 0(CX)
	MOVQ 8(AX), BX
	ADDQ 8(DX), BX

	// reduce element(BX) using temp registers (DI)
	REDUCE(BX,DI)

	MOVQ BX, 8(CX)
	RET

TEXT ·doubleE2(SB), NOSPLIT, $0-16
	MOVQ res+0(FP), DX
	MOVQ x+8(FP), AX
	MOVQ 0(AX), CX
	ADDQ CX, CX

	// reduce element(CX) using temp registers (BX)
	REDUCE(CX,BX)

	MOVQ CX, 0(DX)
	MOVQ 8(AX), CX
	ADDQ CX, CX

	// reduce element(CX) using temp registers (SI)
	REDUCE(CX,SI)

	MOVQ CX, 8(DX)
	RET

TEXT ·subE2(SB), NOSPLIT, $0-24
	XORQ    CX, CX
	MOVQ    x+8(FP), DX
	MOVQ    0(DX), AX
	MOVQ    y+16(FP), DX
	SUBQ    0(DX), AX
	MOVQ    x+8(FP), DX
	MOVQ    $0x0000000078000001, BX
	CMOVQCC CX, BX
	ADDQ    BX, AX
	MOVQ    res+0(FP), SI
	MOVQ    AX, 0(SI)
	MOVQ    8(DX), AX
	MOVQ    y+16(FP), DX
	SUBQ    8(DX), AX
	MOVQ    $0x0000000078000001, DI
	CMOVQCC CX, DI
	ADDQ    DI, AX
	MOVQ    res+0(FP), DX
	MOVQ    AX, 8(DX)
	RET

TEXT ·negE2(SB), NOSPLIT, $0-16
	MOVQ  res+0(FP), DX
	MOVQ  x+8(FP), AX
	MOVQ  0(AX), BX
	MOVQ  BX, AX
	TESTQ AX, AX
	JNE   l1
	MOVQ  AX, 0(DX)
	JMP   l3

l1:
	MOVQ $0x0000000078000001, CX
	SUBQ BX, CX
	MOVQ CX, 0(DX)

l3:
	MOVQ  x+8(FP), AX
	MOVQ  8(AX), BX
	MOVQ  BX, AX
	TESTQ AX, AX
	JNE   l2
	MOVQ  AX, 8(DX)
	RET

l2:
	MOVQ $0x0000000078000001, CX
	SUBQ BX, CX
	MOVQ CX, 8(DX)
	RET

TEXT ·mulNonResE2(SB), NOSPLIT, $0-16
	MOVQ x+8(FP), CX
	MOVQ 0(CX), AX
	ADDQ AX, AX

	// reduce element(AX) using temp registers (BX)
	REDUCE(AX,BX)

	ADDQ AX, AX

	// reduce element(AX) using temp registers (SI)
	REDUCE(AX,SI)

	ADDQ AX, AX

	// reduce element(AX) using temp registers (DI)
	REDUCE(AX,DI)

	ADDQ 0(CX), AX

	// reduce element(AX) using temp registers (R8)
	REDUCE(AX,R8)

	MOVQ    8(CX), DX
	XORQ    R9, R9
	SUBQ    DX, AX
	MOVQ    $0x0000000078000001, R10
	CMOVQCC R9, R10
	ADDQ    R10, AX
	ADDQ    DX, DX

	// reduce element(DX) using temp registers (R11)
	REDUCE(DX,R11)

	ADDQ DX, DX

	// reduce element(DX) using temp registers (R12)
	REDUCE(DX,R12)

	ADDQ DX, DX

	// reduce element(DX) using temp registers (R13)
	REDUCE(DX,R13)

	ADDQ 8(CX), DX

	// reduce element(DX) using temp registers (R14)
	REDUCE(DX,R14)

	ADDQ 0(CX), DX

	// reduce element(DX) using temp registers (R15)
	REDUCE(DX,R15)

	MOVQ res+0(FP), CX
	MOVQ AX, 0(CX)
	MOVQ DX, 8(CX)
	RET

TEXT ·mulAdxE2(SB), $24-24
	NO_LOCAL_POINTERS

	// var a, b, c fr.Element
	// a.Add(&x.A0, &x.A1)
	// b.Add(&y.A0, &y.A1)
	// a.Mul(&a, &b)
	// b.Mul(&x.A0, &y.A0)
	// c.Mul(&x.A1, &y.A1)
	// z.A1.Sub(&a, &b).Sub(&z.A1, &c)
	// z.A0.Sub(&b, &c)

	CMPB ·supportAdx(SB), $1
	JNE  l4
	MOVQ x+8(FP), AX
	MOVQ y+16(FP), DX
	MOVQ 8(AX), R14
	MOVQ 8(DX), R15

	// mul (R14) with (R15) into (CX)
	MUL()

	// reduce element(CX) using temp registers (R15)
	REDUCE(CX,R15)

	MOVQ CX, s1-16(SP)
	MOVQ x+8(FP), AX
	MOVQ y+16(FP), DX
	ADDQ 0(AX), R14
	MOVQ 0(DX), R15
	ADDQ 8(DX), R15

	// mul (R14) with (R15) into (CX)
	MUL()

	// reduce element(CX) using temp registers (R15)
	REDUCE(CX,R15)

	MOVQ CX, s0-8(SP)
	MOVQ x+8(FP), AX
	MOVQ y+16(FP), DX
	MOVQ 0(AX), R14
	MOVQ 0(DX), R15

	// mul (R14) with (R15) into (CX)
	MUL()

	// reduce element(CX) using temp registers (R15)
	REDUCE(CX,R15)

	XORQ    DX, DX
	MOVQ    s0-8(SP), R14
	SUBQ    CX, R14
	MOVQ    $0x0000000078000001, R15
	CMOVQCC DX, R15
	ADDQ    R15, R14
	SUBQ    s1-16(SP), R14
	MOVQ    $0x0000000078000001, R15
	CMOVQCC DX, R15
	ADDQ    R15, R14
	MOVQ    res+0(FP), AX
	MOVQ    R14, 8(AX)
	MOVQ    s1-16(SP), R15
	SUBQ    R15, CX
	MOVQ    $0x0000000078000001, R14
	CMOVQCC DX, R14
	ADDQ    R14, CX
	MOVQ    CX, 0(AX)
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

	// z.A0 = (x.A0 + x.A1) * (x.A0 - x.A1)
	// z.A1 = 2 * x.A0 * x.A1

	CMPB ·supportAdx(SB), $1
	JNE  l5

	// 2 * x.A0 * x.A1
	MOVQ x+8(FP), AX

	// x.A0 -> R15
	MOVQ 0(AX), R15

	// 2 * x.A1 -> R14
	MOVQ 8(AX), R14
	ADDQ R14, R14

	// mul (R14) with (R15) into (CX)
	MUL()

	// reduce element(CX) using temp registers (R14)
	REDUCE(CX,R14)

	MOVQ x+8(FP), AX

	// x.A1 -> R14
	MOVQ 8(AX), R14
	MOVQ res+0(FP), DX
	MOVQ CX, 8(DX)
	MOVQ R14, CX

	// Add(&x.A0, &x.A1)
	ADDQ R15, R14
	XORQ BP, BP

	// Sub(&x.A0, &x.A1)
	SUBQ    CX, R15
	MOVQ    $0x0000000078000001, CX
	CMOVQCC BP, CX
	ADDQ    CX, R15

	// mul (R14) with (R15) into (CX)
	MUL()

	// reduce element(CX) using temp registers (R14)
	REDUCE(CX,R14)

	MOVQ res+0(FP), AX
	MOVQ CX, 0(AX)
	RET

l5:
	MOVQ res+0(FP), AX
	MOVQ AX, (SP)
	MOVQ x+8(FP), AX
	MOVQ AX, 8(SP)
	CALL ·squareGenericE2(SB)
	RET
