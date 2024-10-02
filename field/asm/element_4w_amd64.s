// Code generated by gnark-crypto/generator. DO NOT EDIT.
#include "textflag.h"
#include "funcdata.h"
#include "go_asm.h"

#define REDUCE(ra0, ra1, ra2, ra3, rb0, rb1, rb2, rb3) \
	MOVQ    ra0, rb0;              \
	SUBQ    ·qElement(SB), ra0;    \
	MOVQ    ra1, rb1;              \
	SBBQ    ·qElement+8(SB), ra1;  \
	MOVQ    ra2, rb2;              \
	SBBQ    ·qElement+16(SB), ra2; \
	MOVQ    ra3, rb3;              \
	SBBQ    ·qElement+24(SB), ra3; \
	CMOVQCS rb0, ra0;              \
	CMOVQCS rb1, ra1;              \
	CMOVQCS rb2, ra2;              \
	CMOVQCS rb3, ra3;              \

TEXT ·reduce(SB), NOSPLIT, $0-8
	MOVQ res+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), SI

	// reduce element(DX,CX,BX,SI) using temp registers (DI,R8,R9,R10)
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10)

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	MOVQ SI, 24(AX)
	RET

// MulBy3(x *Element)
TEXT ·MulBy3(SB), NOSPLIT, $0-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), SI
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI

	// reduce element(DX,CX,BX,SI) using temp registers (DI,R8,R9,R10)
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10)

	ADDQ 0(AX), DX
	ADCQ 8(AX), CX
	ADCQ 16(AX), BX
	ADCQ 24(AX), SI

	// reduce element(DX,CX,BX,SI) using temp registers (R11,R12,R13,R14)
	REDUCE(DX,CX,BX,SI,R11,R12,R13,R14)

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	MOVQ SI, 24(AX)
	RET

// MulBy5(x *Element)
TEXT ·MulBy5(SB), NOSPLIT, $0-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), SI
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI

	// reduce element(DX,CX,BX,SI) using temp registers (DI,R8,R9,R10)
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10)

	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI

	// reduce element(DX,CX,BX,SI) using temp registers (R11,R12,R13,R14)
	REDUCE(DX,CX,BX,SI,R11,R12,R13,R14)

	ADDQ 0(AX), DX
	ADCQ 8(AX), CX
	ADCQ 16(AX), BX
	ADCQ 24(AX), SI

	// reduce element(DX,CX,BX,SI) using temp registers (R15,DI,R8,R9)
	REDUCE(DX,CX,BX,SI,R15,DI,R8,R9)

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	MOVQ SI, 24(AX)
	RET

// MulBy13(x *Element)
TEXT ·MulBy13(SB), NOSPLIT, $0-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), SI
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI

	// reduce element(DX,CX,BX,SI) using temp registers (DI,R8,R9,R10)
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10)

	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI

	// reduce element(DX,CX,BX,SI) using temp registers (R11,R12,R13,R14)
	REDUCE(DX,CX,BX,SI,R11,R12,R13,R14)

	MOVQ DX, R11
	MOVQ CX, R12
	MOVQ BX, R13
	MOVQ SI, R14
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI

	// reduce element(DX,CX,BX,SI) using temp registers (DI,R8,R9,R10)
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10)

	ADDQ R11, DX
	ADCQ R12, CX
	ADCQ R13, BX
	ADCQ R14, SI

	// reduce element(DX,CX,BX,SI) using temp registers (DI,R8,R9,R10)
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10)

	ADDQ 0(AX), DX
	ADCQ 8(AX), CX
	ADCQ 16(AX), BX
	ADCQ 24(AX), SI

	// reduce element(DX,CX,BX,SI) using temp registers (DI,R8,R9,R10)
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10)

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	MOVQ SI, 24(AX)
	RET

// Butterfly(a, b *Element) sets a = a + b; b = a - b
TEXT ·Butterfly(SB), NOSPLIT, $0-16
	MOVQ    a+0(FP), AX
	MOVQ    0(AX), CX
	MOVQ    8(AX), BX
	MOVQ    16(AX), SI
	MOVQ    24(AX), DI
	MOVQ    CX, R8
	MOVQ    BX, R9
	MOVQ    SI, R10
	MOVQ    DI, R11
	XORQ    AX, AX
	MOVQ    b+8(FP), DX
	ADDQ    0(DX), CX
	ADCQ    8(DX), BX
	ADCQ    16(DX), SI
	ADCQ    24(DX), DI
	SUBQ    0(DX), R8
	SBBQ    8(DX), R9
	SBBQ    16(DX), R10
	SBBQ    24(DX), R11
	MOVQ    $const_q0, R12
	MOVQ    $const_q1, R13
	MOVQ    $const_q2, R14
	MOVQ    $const_q3, R15
	CMOVQCC AX, R12
	CMOVQCC AX, R13
	CMOVQCC AX, R14
	CMOVQCC AX, R15
	ADDQ    R12, R8
	ADCQ    R13, R9
	ADCQ    R14, R10
	ADCQ    R15, R11
	MOVQ    R8, 0(DX)
	MOVQ    R9, 8(DX)
	MOVQ    R10, 16(DX)
	MOVQ    R11, 24(DX)

	// reduce element(CX,BX,SI,DI) using temp registers (R8,R9,R10,R11)
	REDUCE(CX,BX,SI,DI,R8,R9,R10,R11)

	MOVQ a+0(FP), AX
	MOVQ CX, 0(AX)
	MOVQ BX, 8(AX)
	MOVQ SI, 16(AX)
	MOVQ DI, 24(AX)
	RET

// mul(res, x, y *Element)
TEXT ·mul(SB), $24-24

	// Algorithm 2 of "Faster Montgomery Multiplication and Multi-Scalar-Multiplication for SNARKS"
	// by Y. El Housni and G. Botrel https://doi.org/10.46586/tches.v2023.i3.504-521

	NO_LOCAL_POINTERS
	CMPB ·supportAdx(SB), $1
	JNE  noAdx_1
	MOVQ x+8(FP), SI

	// x[0] -> DI
	// x[1] -> R8
	// x[2] -> R9
	// x[3] -> R10
	MOVQ 0(SI), DI
	MOVQ 8(SI), R8
	MOVQ 16(SI), R9
	MOVQ 24(SI), R10
	MOVQ y+16(FP), R11

	// A -> BP
	// t[0] -> R14
	// t[1] -> R13
	// t[2] -> CX
	// t[3] -> BX
#define DIV_SHIFT() \
	MOVQ  $const_qInvNeg, DX       \
	IMULQ R14, DX                  \
	XORQ  AX, AX                   \
	MULXQ ·qElement+0(SB), AX, R12 \
	ADCXQ R14, AX                  \
	MOVQ  R12, R14                 \
	ADCXQ R13, R14                 \
	MULXQ ·qElement+8(SB), AX, R13 \
	ADOXQ AX, R14                  \
	ADCXQ CX, R13                  \
	MULXQ ·qElement+16(SB), AX, CX \
	ADOXQ AX, R13                  \
	ADCXQ BX, CX                   \
	MULXQ ·qElement+24(SB), AX, BX \
	ADOXQ AX, CX                   \
	MOVQ  $0, AX                   \
	ADCXQ AX, BX                   \
	ADOXQ BP, BX                   \

#define MUL_WORD_0() \
	MULXQ DI, R14, R13 \
	MULXQ R8, AX, CX   \
	ADOXQ AX, R13      \
	MULXQ R9, AX, BX   \
	ADOXQ AX, CX       \
	MULXQ R10, AX, BP  \
	ADOXQ AX, BX       \
	MOVQ  $0, AX       \
	ADOXQ AX, BP       \
	DIV_SHIFT()        \

#define MUL_WORD_N() \
	MULXQ DI, AX, BP  \
	ADOXQ AX, R14     \
	ADCXQ BP, R13     \
	MULXQ R8, AX, BP  \
	ADOXQ AX, R13     \
	ADCXQ BP, CX      \
	MULXQ R9, AX, BP  \
	ADOXQ AX, CX      \
	ADCXQ BP, BX      \
	MULXQ R10, AX, BP \
	ADOXQ AX, BX      \
	MOVQ  $0, AX      \
	ADCXQ AX, BP      \
	ADOXQ AX, BP      \
	DIV_SHIFT()       \

	// clear the flags
	XORQ AX, AX
	MOVQ 0(R11), DX
	MUL_WORD_0()

	// clear the flags
	XORQ AX, AX
	MOVQ 8(R11), DX
	MUL_WORD_N()

	// clear the flags
	XORQ AX, AX
	MOVQ 16(R11), DX
	MUL_WORD_N()

	// clear the flags
	XORQ AX, AX
	MOVQ 24(R11), DX
	MUL_WORD_N()

	// reduce element(R14,R13,CX,BX) using temp registers (SI,R12,R11,DI)
	REDUCE(R14,R13,CX,BX,SI,R12,R11,DI)

	MOVQ res+0(FP), AX
	MOVQ R14, 0(AX)
	MOVQ R13, 8(AX)
	MOVQ CX, 16(AX)
	MOVQ BX, 24(AX)
	RET

noAdx_1:
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

	// Algorithm 2 of "Faster Montgomery Multiplication and Multi-Scalar-Multiplication for SNARKS"
	// by Y. El Housni and G. Botrel https://doi.org/10.46586/tches.v2023.i3.504-521
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
	JNE  noAdx_2
	MOVQ res+0(FP), DX
	MOVQ 0(DX), R14
	MOVQ 8(DX), R13
	MOVQ 16(DX), CX
	MOVQ 24(DX), BX
	XORQ DX, DX

	// m := t[0]*q'[0] mod W
	MOVQ  $const_qInvNeg, DX
	IMULQ R14, DX
	XORQ  AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ ·qElement+0(SB), AX, BP
	ADCXQ R14, AX
	MOVQ  BP, R14

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R13, R14
	MULXQ ·qElement+8(SB), AX, R13
	ADOXQ AX, R14

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ CX, R13
	MULXQ ·qElement+16(SB), AX, CX
	ADOXQ AX, R13

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BX, CX
	MULXQ ·qElement+24(SB), AX, BX
	ADOXQ AX, CX
	MOVQ  $0, AX
	ADCXQ AX, BX
	ADOXQ AX, BX
	XORQ  DX, DX

	// m := t[0]*q'[0] mod W
	MOVQ  $const_qInvNeg, DX
	IMULQ R14, DX
	XORQ  AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ ·qElement+0(SB), AX, BP
	ADCXQ R14, AX
	MOVQ  BP, R14

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R13, R14
	MULXQ ·qElement+8(SB), AX, R13
	ADOXQ AX, R14

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ CX, R13
	MULXQ ·qElement+16(SB), AX, CX
	ADOXQ AX, R13

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BX, CX
	MULXQ ·qElement+24(SB), AX, BX
	ADOXQ AX, CX
	MOVQ  $0, AX
	ADCXQ AX, BX
	ADOXQ AX, BX
	XORQ  DX, DX

	// m := t[0]*q'[0] mod W
	MOVQ  $const_qInvNeg, DX
	IMULQ R14, DX
	XORQ  AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ ·qElement+0(SB), AX, BP
	ADCXQ R14, AX
	MOVQ  BP, R14

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R13, R14
	MULXQ ·qElement+8(SB), AX, R13
	ADOXQ AX, R14

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ CX, R13
	MULXQ ·qElement+16(SB), AX, CX
	ADOXQ AX, R13

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BX, CX
	MULXQ ·qElement+24(SB), AX, BX
	ADOXQ AX, CX
	MOVQ  $0, AX
	ADCXQ AX, BX
	ADOXQ AX, BX
	XORQ  DX, DX

	// m := t[0]*q'[0] mod W
	MOVQ  $const_qInvNeg, DX
	IMULQ R14, DX
	XORQ  AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ ·qElement+0(SB), AX, BP
	ADCXQ R14, AX
	MOVQ  BP, R14

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R13, R14
	MULXQ ·qElement+8(SB), AX, R13
	ADOXQ AX, R14

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ CX, R13
	MULXQ ·qElement+16(SB), AX, CX
	ADOXQ AX, R13

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BX, CX
	MULXQ ·qElement+24(SB), AX, BX
	ADOXQ AX, CX
	MOVQ  $0, AX
	ADCXQ AX, BX
	ADOXQ AX, BX

	// reduce element(R14,R13,CX,BX) using temp registers (SI,DI,R8,R9)
	REDUCE(R14,R13,CX,BX,SI,DI,R8,R9)

	MOVQ res+0(FP), AX
	MOVQ R14, 0(AX)
	MOVQ R13, 8(AX)
	MOVQ CX, 16(AX)
	MOVQ BX, 24(AX)
	RET

noAdx_2:
	MOVQ res+0(FP), AX
	MOVQ AX, (SP)
	CALL ·_fromMontGeneric(SB)
	RET
