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
DATA q<>+0(SB)/8, $15512955586897510413
DATA q<>+8(SB)/8, $4410884215886313276
DATA q<>+16(SB)/8, $15543556715411259941
DATA q<>+24(SB)/8, $9083347379620258823
DATA q<>+32(SB)/8, $13320134076191308873
DATA q<>+40(SB)/8, $9318693926755804304
DATA q<>+48(SB)/8, $5645674015335635503
DATA q<>+56(SB)/8, $12176845843281334983
DATA q<>+64(SB)/8, $18165857675053050549
DATA q<>+72(SB)/8, $82862755739295587
GLOBL q<>(SB), (RODATA+NOPTR), $80
// qInv0 q'[0]
DATA qInv0<>(SB)/8, $13046692460116554043
GLOBL qInv0<>(SB), (RODATA+NOPTR), $8
// add(res, xPtr, yPtr *Element)
TEXT ·add(SB), NOSPLIT, $0-24
	LDP x+8(FP), (R10, R11)

	// load operands and add mod 2^r
	LDP  0(R10), (R0, R12)
	LDP  0(R11), (R1, R13)
	ADDS R0, R1, R0
	ADCS R12, R13, R1
	LDP  16(R10), (R2, R12)
	LDP  16(R11), (R3, R13)
	ADCS R2, R3, R2
	ADCS R12, R13, R3
	LDP  32(R10), (R4, R12)
	LDP  32(R11), (R5, R13)
	ADCS R4, R5, R4
	ADCS R12, R13, R5
	LDP  48(R10), (R6, R12)
	LDP  48(R11), (R7, R13)
	ADCS R6, R7, R6
	ADCS R12, R13, R7
	LDP  64(R10), (R8, R12)
	LDP  64(R11), (R9, R13)
	ADCS R8, R9, R8
	ADCS R12, R13, R9

	// load modulus and subtract
	LDP  q<>+0(SB), (R10, R11)
	SUBS R10, R0, R10
	SBCS R11, R1, R11
	LDP  q<>+16(SB), (R12, R13)
	SBCS R12, R2, R12
	SBCS R13, R3, R13
	LDP  q<>+32(SB), (R14, R15)
	SBCS R14, R4, R14
	SBCS R15, R5, R15
	LDP  q<>+48(SB), (R16, R17)
	SBCS R16, R6, R16
	SBCS R17, R7, R17
	LDP  q<>+64(SB), (R19, R20)
	SBCS R19, R8, R19
	SBCS R20, R9, R20

	// reduce if necessary
	CSEL CS, R10, R0, R0
	CSEL CS, R11, R1, R1
	CSEL CS, R12, R2, R2
	CSEL CS, R13, R3, R3
	CSEL CS, R14, R4, R4
	CSEL CS, R15, R5, R5
	CSEL CS, R16, R6, R6
	CSEL CS, R17, R7, R7
	CSEL CS, R19, R8, R8
	CSEL CS, R20, R9, R9

	// store
	MOVD res+0(FP), R10
	STP  (R0, R1), 0(R10)
	STP  (R2, R3), 16(R10)
	STP  (R4, R5), 32(R10)
	STP  (R6, R7), 48(R10)
	STP  (R8, R9), 64(R10)
	RET

// sub(res, xPtr, yPtr *Element)
TEXT ·sub(SB), NOSPLIT, $0-24
	LDP x+8(FP), (R10, R11)

	// load operands and subtract mod 2^r
	LDP  0(R10), (R0, R12)
	LDP  0(R11), (R1, R13)
	SUBS R1, R0, R0
	SBCS R13, R12, R1
	LDP  16(R10), (R2, R12)
	LDP  16(R11), (R3, R13)
	SBCS R3, R2, R2
	SBCS R13, R12, R3
	LDP  32(R10), (R4, R12)
	LDP  32(R11), (R5, R13)
	SBCS R5, R4, R4
	SBCS R13, R12, R5
	LDP  48(R10), (R6, R12)
	LDP  48(R11), (R7, R13)
	SBCS R7, R6, R6
	SBCS R13, R12, R7
	LDP  64(R10), (R8, R12)
	LDP  64(R11), (R9, R13)
	SBCS R9, R8, R8
	SBCS R13, R12, R9

	// Store borrow TODO: Can it be done with one instruction?
	MOVD $0, R10
	ADC  $0, R10, R10

	// load modulus and add
	LDP  q<>+0(SB), (R11, R12)
	ADDS R11, R0, R11
	ADCS R12, R1, R12
	LDP  q<>+16(SB), (R13, R14)
	ADCS R13, R2, R13
	ADCS R14, R3, R14
	LDP  q<>+32(SB), (R15, R16)
	ADCS R15, R4, R15
	ADCS R16, R5, R16
	LDP  q<>+48(SB), (R17, R19)
	ADCS R17, R6, R17
	ADCS R19, R7, R19
	LDP  q<>+64(SB), (R20, R21)
	ADCS R20, R8, R20
	ADCS R21, R9, R21

	// augment if necessary
	CMP  $1, R10         // "recall" the borrow
	CSEL NE, R11, R0, R0
	CSEL NE, R12, R1, R1
	CSEL NE, R13, R2, R2
	CSEL NE, R14, R3, R3
	CSEL NE, R15, R4, R4
	CSEL NE, R16, R5, R5
	CSEL NE, R17, R6, R6
	CSEL NE, R19, R7, R7
	CSEL NE, R20, R8, R8
	CSEL NE, R21, R9, R9

	// store
	MOVD res+0(FP), R10
	STP  (R0, R1), 0(R10)
	STP  (R2, R3), 16(R10)
	STP  (R4, R5), 32(R10)
	STP  (R6, R7), 48(R10)
	STP  (R8, R9), 64(R10)
	RET
