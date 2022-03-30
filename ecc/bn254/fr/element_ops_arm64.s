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
DATA q<>+0(SB)/8, $4891460686036598785
DATA q<>+8(SB)/8, $2896914383306846353
DATA q<>+16(SB)/8, $13281191951274694749
DATA q<>+24(SB)/8, $3486998266802970665
GLOBL q<>(SB), (RODATA+NOPTR), $32
// qInv0 q'[0]
DATA qInv0<>(SB)/8, $14042775128853446655
GLOBL qInv0<>(SB), (RODATA+NOPTR), $8
// add(res, xPtr, yPtr *Element)
TEXT ·add(SB), NOSPLIT, $0-24
	LDP x+8(FP), (R4, R5)

	// load operands and add mod 2^r
	LDP  0(R4), (R0, R6)
	LDP  0(R5), (R1, R7)
	ADDS R0, R1, R0
	ADCS R6, R7, R1
	LDP  16(R4), (R2, R6)
	LDP  16(R5), (R3, R7)
	ADCS R2, R3, R2
	ADCS R6, R7, R3

	// load modulus and subtract
	LDP  q<>+0(SB), (R4, R5)
	SUBS R4, R0, R4
	SBCS R5, R1, R5
	LDP  q<>+16(SB), (R6, R7)
	SBCS R6, R2, R6
	SBCS R7, R3, R7

	// reduce if necessary
	CSEL CS, R4, R0, R0
	CSEL CS, R5, R1, R1
	CSEL CS, R6, R2, R2
	CSEL CS, R7, R3, R3

	// store
	MOVD res+0(FP), R4
	STP  (R0, R1), 0(R4)
	STP  (R2, R3), 16(R4)
	RET

// sub(res, xPtr, yPtr *Element)
TEXT ·sub(SB), NOSPLIT, $0-24
	LDP x+8(FP), (R4, R5)

	// load operands and subtract mod 2^r
	LDP  0(R4), (R0, R6)
	LDP  0(R5), (R1, R7)
	SUBS R0, R1, R0
	SBCS R6, R7, R1
	LDP  16(R4), (R2, R6)
	LDP  16(R5), (R3, R7)
	SBCS R2, R3, R2
	SBCS R6, R7, R3

	// Store borrow TODO: Can it be done with one instruction?
	MOVD $0, R4
	ADC  $0, R4, R4

	// load modulus and add
	LDP  q<>+0(SB), (R5, R6)
	ADDS R5, R0, R5
	ADCS R6, R1, R6
	LDP  q<>+16(SB), (R7, R8)
	ADCS R7, R2, R7
	ADCS R8, R3, R8

	// augment if necessary
	CMP  $1, R4         // "recall" the borrow
	CSEL EQ, R5, R0, R0
	CSEL EQ, R6, R1, R1
	CSEL EQ, R7, R2, R2
	CSEL EQ, R8, R3, R3

	// store
	MOVD res+0(FP), R4
	STP  (R0, R1), 0(R4)
	STP  (R2, R3), 16(R4)
	RET
