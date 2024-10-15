// Code generated by gnark-crypto/generator. DO NOT EDIT.
#include "textflag.h"
#include "funcdata.h"
#include "go_asm.h"

// add(res, x, y *Element)
TEXT ·add(SB), NOSPLIT, $0-24
	LDP x+8(FP), (R12, R13)

	// load operands and add mod 2^r
	LDP  0(R12), (R8, R9)
	LDP  0(R13), (R4, R5)
	LDP  16(R12), (R10, R11)
	LDP  16(R13), (R6, R7)
	ADDS R8, R4, R4
	ADCS R9, R5, R5
	ADCS R10, R6, R6
	ADCS R11, R7, R7

	// load modulus and subtract
	MOVD $const_q0, R0
	MOVD $const_q1, R1
	MOVD $const_q2, R2
	MOVD $const_q3, R3
	SUBS R0, R4, R0
	SBCS R1, R5, R1
	SBCS R2, R6, R2
	SBCS R3, R7, R3

	// reduce if necessary
	CSEL CS, R0, R4, R4
	CSEL CS, R1, R5, R5
	CSEL CS, R2, R6, R6
	CSEL CS, R3, R7, R7

	// store
	MOVD res+0(FP), R14
	STP  (R4, R5), 0(R14)
	STP  (R6, R7), 16(R14)
	RET
