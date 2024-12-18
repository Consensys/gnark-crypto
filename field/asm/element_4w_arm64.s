// Code generated by gnark-crypto/generator. DO NOT EDIT.
#include "textflag.h"
#include "funcdata.h"
#include "go_asm.h"

// butterfly(a, b *Element)
// a, b = a+b, a-b
TEXT ·Butterfly(SB), NOFRAME|NOSPLIT, $0-16
	LDP  x+0(FP), (R16, R17)
	LDP  0(R16), (R0, R1)
	LDP  16(R16), (R2, R3)
	LDP  0(R17), (R4, R5)
	LDP  16(R17), (R6, R7)
	ADDS R0, R4, R8
	ADCS R1, R5, R9
	ADCS R2, R6, R10
	ADC  R3, R7, R11
	SUBS R4, R0, R4
	SBCS R5, R1, R5
	SBCS R6, R2, R6
	SBCS R7, R3, R7
	LDP  ·qElement+0(SB), (R0, R1)
	CSEL CS, ZR, R0, R12
	CSEL CS, ZR, R1, R13
	LDP  ·qElement+16(SB), (R2, R3)
	CSEL CS, ZR, R2, R14
	CSEL CS, ZR, R3, R15

	// add q if underflow, 0 if not
	ADDS R4, R12, R4
	ADCS R5, R13, R5
	STP  (R4, R5), 0(R17)
	ADCS R6, R14, R6
	ADC  R7, R15, R7
	STP  (R6, R7), 16(R17)

	// q = t - q
	SUBS R0, R8, R0
	SBCS R1, R9, R1
	SBCS R2, R10, R2
	SBCS R3, R11, R3

	// if no borrow, return q, else return t
	CSEL CS, R0, R8, R8
	CSEL CS, R1, R9, R9
	STP  (R8, R9), 0(R16)
	CSEL CS, R2, R10, R10
	CSEL CS, R3, R11, R11
	STP  (R10, R11), 16(R16)
	RET

// mul(res, x, y *Element)
// Algorithm 2 of Faster Montgomery Multiplication and Multi-Scalar-Multiplication for SNARKS
// by Y. El Housni and G. Botrel https://doi.org/10.46586/tches.v2023.i3.504-521
TEXT ·mul(SB), NOFRAME|NOSPLIT, $0-24
#define DIVSHIFT() \
	MUL   R13, R12, R0 \
	ADDS  R0, R6, R6   \
	MUL   R14, R12, R0 \
	ADCS  R0, R7, R7   \
	MUL   R15, R12, R0 \
	ADCS  R0, R8, R8   \
	MUL   R16, R12, R0 \
	ADCS  R0, R9, R9   \
	ADC   R10, ZR, R10 \
	UMULH R13, R12, R0 \
	ADDS  R0, R7, R6   \
	UMULH R14, R12, R0 \
	ADCS  R0, R8, R7   \
	UMULH R15, R12, R0 \
	ADCS  R0, R9, R8   \
	UMULH R16, R12, R0 \
	ADCS  R0, R10, R9  \

#define MUL_WORD_N() \
	MUL   R2, R1, R0   \
	ADDS  R0, R6, R6   \
	MUL   R6, R11, R12 \
	MUL   R3, R1, R0   \
	ADCS  R0, R7, R7   \
	MUL   R4, R1, R0   \
	ADCS  R0, R8, R8   \
	MUL   R5, R1, R0   \
	ADCS  R0, R9, R9   \
	ADC   ZR, ZR, R10  \
	UMULH R2, R1, R0   \
	ADDS  R0, R7, R7   \
	UMULH R3, R1, R0   \
	ADCS  R0, R8, R8   \
	UMULH R4, R1, R0   \
	ADCS  R0, R9, R9   \
	UMULH R5, R1, R0   \
	ADC   R0, R10, R10 \
	DIVSHIFT()         \

#define MUL_WORD_0() \
	MUL   R2, R1, R6   \
	MUL   R3, R1, R7   \
	MUL   R4, R1, R8   \
	MUL   R5, R1, R9   \
	UMULH R2, R1, R0   \
	ADDS  R0, R7, R7   \
	UMULH R3, R1, R0   \
	ADCS  R0, R8, R8   \
	UMULH R4, R1, R0   \
	ADCS  R0, R9, R9   \
	UMULH R5, R1, R0   \
	ADC   R0, ZR, R10  \
	MUL   R6, R11, R12 \
	DIVSHIFT()         \

	MOVD y+16(FP), R17
	MOVD x+8(FP), R0
	LDP  0(R0), (R2, R3)
	LDP  16(R0), (R4, R5)
	MOVD 0(R17), R1
	MOVD $const_qInvNeg, R11
	LDP  ·qElement+0(SB), (R13, R14)
	LDP  ·qElement+16(SB), (R15, R16)
	MUL_WORD_0()
	MOVD 8(R17), R1
	MUL_WORD_N()
	MOVD 16(R17), R1
	MUL_WORD_N()
	MOVD 24(R17), R1
	MUL_WORD_N()

	// reduce if necessary
	SUBS R13, R6, R13
	SBCS R14, R7, R14
	SBCS R15, R8, R15
	SBCS R16, R9, R16
	MOVD res+0(FP), R0
	CSEL CS, R13, R6, R6
	CSEL CS, R14, R7, R7
	STP  (R6, R7), 0(R0)
	CSEL CS, R15, R8, R8
	CSEL CS, R16, R9, R9
	STP  (R8, R9), 16(R0)
	RET

// reduce(res *Element)
TEXT ·reduce(SB), NOFRAME|NOSPLIT, $0-8
	LDP  ·qElement+0(SB), (R4, R5)
	LDP  ·qElement+16(SB), (R6, R7)
	MOVD res+0(FP), R8
	LDP  0(R8), (R0, R1)
	LDP  16(R8), (R2, R3)

	// q = t - q
	SUBS R4, R0, R4
	SBCS R5, R1, R5
	SBCS R6, R2, R6
	SBCS R7, R3, R7

	// if no borrow, return q, else return t
	CSEL CS, R4, R0, R0
	CSEL CS, R5, R1, R1
	STP  (R0, R1), 0(R8)
	CSEL CS, R6, R2, R2
	CSEL CS, R7, R3, R3
	STP  (R2, R3), 16(R8)
	RET
