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
DATA q<>+0(SB)/8, $1
DATA q<>+8(SB)/8, $3731203976813871104
DATA q<>+16(SB)/8, $15039355238879481536
DATA q<>+24(SB)/8, $4828608925799409630
DATA q<>+32(SB)/8, $16326337093237622437
DATA q<>+40(SB)/8, $756237273905161798
DATA q<>+48(SB)/8, $16934317532427647658
DATA q<>+56(SB)/8, $14755673041361585881
DATA q<>+64(SB)/8, $18154628166362162086
DATA q<>+72(SB)/8, $6671956210750770825
DATA q<>+80(SB)/8, $16333450281447942351
DATA q<>+88(SB)/8, $4352613195430282
GLOBL q<>(SB), (RODATA+NOPTR), $96
// qInv0 q'[0]
DATA qInv0<>(SB)/8, $18446744073709551615
GLOBL qInv0<>(SB), (RODATA+NOPTR), $8
#define storeVector(ePtr, e0, e1, e2, e3, e4, e5, e6, e7, e8, e9, e10, e11) \
	STP (e0, e1), 0(ePtr)    \
	STP (e2, e3), 16(ePtr)   \
	STP (e4, e5), 32(ePtr)   \
	STP (e6, e7), 48(ePtr)   \
	STP (e8, e9), 64(ePtr)   \
	STP (e10, e11), 80(ePtr) \

// add(res, x, y *Element)
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
	storeVector(R12, R0, R1, R2, R3, R4, R5, R6, R7, R8, R9, R10, R11)
	RET

// sub(res, x, y *Element)
TEXT ·sub(SB), NOSPLIT, $0-24
	LDP x+8(FP), (R12, R13)

	// load operands and subtract mod 2^r
	LDP  0(R12), (R0, R14)
	LDP  0(R13), (R1, R15)
	SUBS R1, R0, R0
	SBCS R15, R14, R1
	LDP  16(R12), (R2, R14)
	LDP  16(R13), (R3, R15)
	SBCS R3, R2, R2
	SBCS R15, R14, R3
	LDP  32(R12), (R4, R14)
	LDP  32(R13), (R5, R15)
	SBCS R5, R4, R4
	SBCS R15, R14, R5
	LDP  48(R12), (R6, R14)
	LDP  48(R13), (R7, R15)
	SBCS R7, R6, R6
	SBCS R15, R14, R7
	LDP  64(R12), (R8, R14)
	LDP  64(R13), (R9, R15)
	SBCS R9, R8, R8
	SBCS R15, R14, R9
	LDP  80(R12), (R10, R14)
	LDP  80(R13), (R11, R15)
	SBCS R11, R10, R10
	SBCS R15, R14, R11

	// load modulus and select
	MOVD $0, R25
	LDP  q<>+0(SB), (R12, R13)
	CSEL CS, R25, R12, R12
	CSEL CS, R25, R13, R13
	LDP  q<>+16(SB), (R14, R15)
	CSEL CS, R25, R14, R14
	CSEL CS, R25, R15, R15
	LDP  q<>+32(SB), (R16, R17)
	CSEL CS, R25, R16, R16
	CSEL CS, R25, R17, R17
	LDP  q<>+48(SB), (R19, R20)
	CSEL CS, R25, R19, R19
	CSEL CS, R25, R20, R20
	LDP  q<>+64(SB), (R21, R22)
	CSEL CS, R25, R21, R21
	CSEL CS, R25, R22, R22
	LDP  q<>+80(SB), (R23, R24)
	CSEL CS, R25, R23, R23
	CSEL CS, R25, R24, R24

	// augment (or not)
	ADDS R0, R12, R0
	ADCS R1, R13, R1
	ADCS R2, R14, R2
	ADCS R3, R15, R3
	ADCS R4, R16, R4
	ADCS R5, R17, R5
	ADCS R6, R19, R6
	ADCS R7, R20, R7
	ADCS R8, R21, R8
	ADCS R9, R22, R9
	ADCS R10, R23, R10
	ADCS R11, R24, R11

	// store
	MOVD res+0(FP), R12
	storeVector(R12, R0, R1, R2, R3, R4, R5, R6, R7, R8, R9, R10, R11)
	RET

// double(res, x *Element)
TEXT ·double(SB), NOSPLIT, $0-16
	LDP res+0(FP), (R13, R12)

	// load operands and add mod 2^r
	LDP  0(R12), (R0, R1)
	ADDS R0, R0, R0
	ADCS R1, R1, R1
	LDP  16(R12), (R2, R3)
	ADCS R2, R2, R2
	ADCS R3, R3, R3
	LDP  32(R12), (R4, R5)
	ADCS R4, R4, R4
	ADCS R5, R5, R5
	LDP  48(R12), (R6, R7)
	ADCS R6, R6, R6
	ADCS R7, R7, R7
	LDP  64(R12), (R8, R9)
	ADCS R8, R8, R8
	ADCS R9, R9, R9
	LDP  80(R12), (R10, R11)
	ADCS R10, R10, R10
	ADCS R11, R11, R11

	// load modulus and subtract
	LDP  q<>+0(SB), (R12, R14)
	SUBS R12, R0, R12
	SBCS R14, R1, R14
	LDP  q<>+16(SB), (R15, R16)
	SBCS R15, R2, R15
	SBCS R16, R3, R16
	LDP  q<>+32(SB), (R17, R19)
	SBCS R17, R4, R17
	SBCS R19, R5, R19
	LDP  q<>+48(SB), (R20, R21)
	SBCS R20, R6, R20
	SBCS R21, R7, R21
	LDP  q<>+64(SB), (R22, R23)
	SBCS R22, R8, R22
	SBCS R23, R9, R23
	LDP  q<>+80(SB), (R24, R25)
	SBCS R24, R10, R24
	SBCS R25, R11, R25

	// reduce if necessary
	CSEL CS, R12, R0, R0
	CSEL CS, R14, R1, R1
	CSEL CS, R15, R2, R2
	CSEL CS, R16, R3, R3
	CSEL CS, R17, R4, R4
	CSEL CS, R19, R5, R5
	CSEL CS, R20, R6, R6
	CSEL CS, R21, R7, R7
	CSEL CS, R22, R8, R8
	CSEL CS, R23, R9, R9
	CSEL CS, R24, R10, R10
	CSEL CS, R25, R11, R11

	// store
	storeVector(R13, R0, R1, R2, R3, R4, R5, R6, R7, R8, R9, R10, R11)
	RET

// neg(res, x *Element)
TEXT ·neg(SB), NOSPLIT, $0-16
	LDP res+0(FP), (R13, R12)

	// load operands and subtract
	MOVD $0, R16
	LDP  0(R12), (R0, R1)
	LDP  q<>+0(SB), (R14, R15)
	ORR  R0, R16, R16             // has x been 0 so far?
	ORR  R1, R16, R16
	SUBS R0, R14, R0
	SBCS R1, R15, R1
	LDP  16(R12), (R2, R3)
	LDP  q<>+16(SB), (R14, R15)
	ORR  R2, R16, R16             // has x been 0 so far?
	ORR  R3, R16, R16
	SBCS R2, R14, R2
	SBCS R3, R15, R3
	LDP  32(R12), (R4, R5)
	LDP  q<>+32(SB), (R14, R15)
	ORR  R4, R16, R16             // has x been 0 so far?
	ORR  R5, R16, R16
	SBCS R4, R14, R4
	SBCS R5, R15, R5
	LDP  48(R12), (R6, R7)
	LDP  q<>+48(SB), (R14, R15)
	ORR  R6, R16, R16             // has x been 0 so far?
	ORR  R7, R16, R16
	SBCS R6, R14, R6
	SBCS R7, R15, R7
	LDP  64(R12), (R8, R9)
	LDP  q<>+64(SB), (R14, R15)
	ORR  R8, R16, R16             // has x been 0 so far?
	ORR  R9, R16, R16
	SBCS R8, R14, R8
	SBCS R9, R15, R9
	LDP  80(R12), (R10, R11)
	LDP  q<>+80(SB), (R14, R15)
	ORR  R10, R16, R16            // has x been 0 so far?
	ORR  R11, R16, R16
	SBCS R10, R14, R10
	SBCS R11, R15, R11
	TST  $0xffffffffffffffff, R16
	CSEL EQ, R16, R0, R0
	CSEL EQ, R16, R1, R1
	CSEL EQ, R16, R2, R2
	CSEL EQ, R16, R3, R3
	CSEL EQ, R16, R4, R4
	CSEL EQ, R16, R5, R5
	CSEL EQ, R16, R6, R6
	CSEL EQ, R16, R7, R7
	CSEL EQ, R16, R8, R8
	CSEL EQ, R16, R9, R9
	CSEL EQ, R16, R10, R10
	CSEL EQ, R16, R11, R11

	// store
	storeVector(R13, R0, R1, R2, R3, R4, R5, R6, R7, R8, R9, R10, R11)
	RET
