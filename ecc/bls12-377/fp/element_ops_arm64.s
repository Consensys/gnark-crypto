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
DATA q<>+0(SB)/8, $9586122913090633729
DATA q<>+8(SB)/8, $1660523435060625408
DATA q<>+16(SB)/8, $2230234197602682880
DATA q<>+24(SB)/8, $1883307231910630287
DATA q<>+32(SB)/8, $14284016967150029115
DATA q<>+40(SB)/8, $121098312706494698
GLOBL q<>(SB), (RODATA+NOPTR), $48
// qInv0 q'[0]
DATA qInv0<>(SB)/8, $9586122913090633727
GLOBL qInv0<>(SB), (RODATA+NOPTR), $8
#define storeVector(ePtr, e0, e1, e2, e3, e4, e5) \
	STP (e0, e1), 0(ePtr)  \
	STP (e2, e3), 16(ePtr) \
	STP (e4, e5), 32(ePtr) \

// add(res, x, y *Element)
TEXT ·add(SB), NOSPLIT, $0-24
	LDP x+8(FP), (R6, R7)

	// load operands and add mod 2^r
	LDP  0(R6), (R0, R8)
	LDP  0(R7), (R1, R9)
	ADDS R0, R1, R0
	ADCS R8, R9, R1
	LDP  16(R6), (R2, R8)
	LDP  16(R7), (R3, R9)
	ADCS R2, R3, R2
	ADCS R8, R9, R3
	LDP  32(R6), (R4, R8)
	LDP  32(R7), (R5, R9)
	ADCS R4, R5, R4
	ADCS R8, R9, R5

	// load modulus and subtract
	LDP  q<>+0(SB), (R6, R7)
	SUBS R6, R0, R6
	SBCS R7, R1, R7
	LDP  q<>+16(SB), (R8, R9)
	SBCS R8, R2, R8
	SBCS R9, R3, R9
	LDP  q<>+32(SB), (R10, R11)
	SBCS R10, R4, R10
	SBCS R11, R5, R11

	// reduce if necessary
	CSEL CS, R6, R0, R0
	CSEL CS, R7, R1, R1
	CSEL CS, R8, R2, R2
	CSEL CS, R9, R3, R3
	CSEL CS, R10, R4, R4
	CSEL CS, R11, R5, R5

	// store
	MOVD res+0(FP), R6
	storeVector(R6, R0, R1, R2, R3, R4, R5)
	RET

// sub(res, x, y *Element)
TEXT ·sub(SB), NOSPLIT, $0-24
	LDP x+8(FP), (R6, R7)

	// load operands and subtract mod 2^r
	LDP  0(R6), (R0, R8)
	LDP  0(R7), (R1, R9)
	SUBS R1, R0, R0
	SBCS R9, R8, R1
	LDP  16(R6), (R2, R8)
	LDP  16(R7), (R3, R9)
	SBCS R3, R2, R2
	SBCS R9, R8, R3
	LDP  32(R6), (R4, R8)
	LDP  32(R7), (R5, R9)
	SBCS R5, R4, R4
	SBCS R9, R8, R5

	// load modulus and select
	MOVD $0, R12
	LDP  q<>+0(SB), (R6, R7)
	CSEL CS, R12, R6, R6
	CSEL CS, R12, R7, R7
	LDP  q<>+16(SB), (R8, R9)
	CSEL CS, R12, R8, R8
	CSEL CS, R12, R9, R9
	LDP  q<>+32(SB), (R10, R11)
	CSEL CS, R12, R10, R10
	CSEL CS, R12, R11, R11

	// augment (or not)
	ADDS R0, R6, R0
	ADCS R1, R7, R1
	ADCS R2, R8, R2
	ADCS R3, R9, R3
	ADCS R4, R10, R4
	ADCS R5, R11, R5

	// store
	MOVD res+0(FP), R6
	storeVector(R6, R0, R1, R2, R3, R4, R5)
	RET

// double(res, x *Element)
TEXT ·double(SB), NOSPLIT, $0-16
	LDP res+0(FP), (R7, R6)

	// load operands and add mod 2^r
	LDP  0(R6), (R0, R1)
	ADDS R0, R0, R0
	ADCS R1, R1, R1
	LDP  16(R6), (R2, R3)
	ADCS R2, R2, R2
	ADCS R3, R3, R3
	LDP  32(R6), (R4, R5)
	ADCS R4, R4, R4
	ADCS R5, R5, R5

	// load modulus and subtract
	LDP  q<>+0(SB), (R6, R8)
	SUBS R6, R0, R6
	SBCS R8, R1, R8
	LDP  q<>+16(SB), (R9, R10)
	SBCS R9, R2, R9
	SBCS R10, R3, R10
	LDP  q<>+32(SB), (R11, R12)
	SBCS R11, R4, R11
	SBCS R12, R5, R12

	// reduce if necessary
	CSEL CS, R6, R0, R0
	CSEL CS, R8, R1, R1
	CSEL CS, R9, R2, R2
	CSEL CS, R10, R3, R3
	CSEL CS, R11, R4, R4
	CSEL CS, R12, R5, R5

	// store
	storeVector(R7, R0, R1, R2, R3, R4, R5)
	RET

// neg(res, x *Element)
TEXT ·neg(SB), NOSPLIT, $0-16
	LDP res+0(FP), (R7, R6)

	// load operands and subtract
	MOVD $0, R10
	LDP  0(R6), (R0, R1)
	LDP  q<>+0(SB), (R8, R9)
	ORR  R0, R10, R10             // has x been 0 so far?
	ORR  R1, R10, R10
	SUBS R0, R8, R0
	SBCS R1, R9, R1
	LDP  16(R6), (R2, R3)
	LDP  q<>+16(SB), (R8, R9)
	ORR  R2, R10, R10             // has x been 0 so far?
	ORR  R3, R10, R10
	SBCS R2, R8, R2
	SBCS R3, R9, R3
	LDP  32(R6), (R4, R5)
	LDP  q<>+32(SB), (R8, R9)
	ORR  R4, R10, R10             // has x been 0 so far?
	ORR  R5, R10, R10
	SBCS R4, R8, R4
	SBCS R5, R9, R5
	TST  $0xffffffffffffffff, R10
	CSEL EQ, R10, R0, R0
	CSEL EQ, R10, R1, R1
	CSEL EQ, R10, R2, R2
	CSEL EQ, R10, R3, R3
	CSEL EQ, R10, R4, R4
	CSEL EQ, R10, R5, R5

	// store
	storeVector(R7, R0, R1, R2, R3, R4, R5)
	RET
