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
DATA q<>+0(SB)/8, $17626244516597989515
DATA q<>+8(SB)/8, $16614129118623039618
DATA q<>+16(SB)/8, $1588918198704579639
DATA q<>+24(SB)/8, $10998096788944562424
DATA q<>+32(SB)/8, $8204665564953313070
DATA q<>+40(SB)/8, $9694500593442880912
DATA q<>+48(SB)/8, $274362232328168196
DATA q<>+56(SB)/8, $8105254717682411801
DATA q<>+64(SB)/8, $5945444129596489281
DATA q<>+72(SB)/8, $13341377791855249032
DATA q<>+80(SB)/8, $15098257552581525310
DATA q<>+88(SB)/8, $81882988782276106
GLOBL q<>(SB), (RODATA+NOPTR), $96
// qInv0 q'[0]
DATA qInv0<>(SB)/8, $744663313386281181
GLOBL qInv0<>(SB), (RODATA+NOPTR), $8
// add(res, xPtr, yPtr *Element)
TEXT ·add(SB), NOSPLIT, $0-24
	LDP x+8(FP), (R12, R13)

	// load operands and add mod 2^r
	LDP  0(R12), (R0, R14)
	LDP  0(R13), (R1, R15)
	ADDS R0, R1, R0
	ADCS R14, R15, R1
	LDP  16(R12), (R2, R14)
	LDP  16(R13), (R3, R15)
	ADCS R2, R3, R2
	ADCS R14, R15, R3
	LDP  32(R12), (R4, R14)
	LDP  32(R13), (R5, R15)
	ADCS R4, R5, R4
	ADCS R14, R15, R5
	LDP  48(R12), (R6, R14)
	LDP  48(R13), (R7, R15)
	ADCS R6, R7, R6
	ADCS R14, R15, R7
	LDP  64(R12), (R8, R14)
	LDP  64(R13), (R9, R15)
	ADCS R8, R9, R8
	ADCS R14, R15, R9
	LDP  80(R12), (R10, R14)
	LDP  80(R13), (R11, R15)
	ADCS R10, R11, R10
	ADCS R14, R15, R11

	// load modulus and subtract
	LDP  q<>+0(SB), (R12, R13)
	SUBS R12, R0, R12
	SBCS R13, R1, R13
	LDP  q<>+16(SB), (R14, R15)
	SBCS R14, R2, R14
	SBCS R15, R3, R15
	LDP  q<>+32(SB), (R16, R17)
	SBCS R16, R4, R16
	SBCS R17, R5, R17
	LDP  q<>+48(SB), (R19, R20)
	SBCS R19, R6, R19
	SBCS R20, R7, R20
	LDP  q<>+64(SB), (R21, R22)
	SBCS R21, R8, R21
	SBCS R22, R9, R22
	LDP  q<>+80(SB), (R23, R24)
	SBCS R23, R10, R23
	SBCS R24, R11, R24

	// reduce if necessary
	CSEL CS, R12, R0, R0
	CSEL CS, R13, R1, R1
	CSEL CS, R14, R2, R2
	CSEL CS, R15, R3, R3
	CSEL CS, R16, R4, R4
	CSEL CS, R17, R5, R5
	CSEL CS, R19, R6, R6
	CSEL CS, R20, R7, R7
	CSEL CS, R21, R8, R8
	CSEL CS, R22, R9, R9
	CSEL CS, R23, R10, R10
	CSEL CS, R24, R11, R11

	// store
	MOVD res+0(FP), R12
	STP  (R0, R1), 0(R12)
	STP  (R2, R3), 16(R12)
	STP  (R4, R5), 32(R12)
	STP  (R6, R7), 48(R12)
	STP  (R8, R9), 64(R12)
	STP  (R10, R11), 80(R12)
	RET

// sub(res, xPtr, yPtr *Element)
TEXT ·sub(SB), NOSPLIT, $0-24
	LDP x+8(FP), (R12, R13)

	// load operands and subtract mod 2^r
	LDP  0(R12), (R0, R14)
	LDP  0(R13), (R1, R15)
	SUBS R0, R1, R0
	SBCS R14, R15, R1
	LDP  16(R12), (R2, R14)
	LDP  16(R13), (R3, R15)
	SBCS R2, R3, R2
	SBCS R14, R15, R3
	LDP  32(R12), (R4, R14)
	LDP  32(R13), (R5, R15)
	SBCS R4, R5, R4
	SBCS R14, R15, R5
	LDP  48(R12), (R6, R14)
	LDP  48(R13), (R7, R15)
	SBCS R6, R7, R6
	SBCS R14, R15, R7
	LDP  64(R12), (R8, R14)
	LDP  64(R13), (R9, R15)
	SBCS R8, R9, R8
	SBCS R14, R15, R9
	LDP  80(R12), (R10, R14)
	LDP  80(R13), (R11, R15)
	SBCS R10, R11, R10
	SBCS R14, R15, R11

	// Store borrow TODO: Can it be done with one instruction?
	MOVD $0, R12
	ADC  $0, R12, R12

	// load modulus and add
	LDP  q<>+0(SB), (R13, R14)
	ADDS R13, R0, R13
	ADCS R14, R1, R14
	LDP  q<>+16(SB), (R15, R16)
	ADCS R15, R2, R15
	ADCS R16, R3, R16
	LDP  q<>+32(SB), (R17, R19)
	ADCS R17, R4, R17
	ADCS R19, R5, R19
	LDP  q<>+48(SB), (R20, R21)
	ADCS R20, R6, R20
	ADCS R21, R7, R21
	LDP  q<>+64(SB), (R22, R23)
	ADCS R22, R8, R22
	ADCS R23, R9, R23
	LDP  q<>+80(SB), (R24, R25)
	ADCS R24, R10, R24
	ADCS R25, R11, R25

	// augment if necessary
	CMP  $1, R12           // "recall" the borrow
	CSEL EQ, R13, R0, R0
	CSEL EQ, R14, R1, R1
	CSEL EQ, R15, R2, R2
	CSEL EQ, R16, R3, R3
	CSEL EQ, R17, R4, R4
	CSEL EQ, R19, R5, R5
	CSEL EQ, R20, R6, R6
	CSEL EQ, R21, R7, R7
	CSEL EQ, R22, R8, R8
	CSEL EQ, R23, R9, R9
	CSEL EQ, R24, R10, R10
	CSEL EQ, R25, R11, R11

	// store
	MOVD res+0(FP), R12
	STP  (R0, R1), 0(R12)
	STP  (R2, R3), 16(R12)
	STP  (R4, R5), 32(R12)
	STP  (R6, R7), 48(R12)
	STP  (R8, R9), 64(R12)
	STP  (R10, R11), 80(R12)
	RET
