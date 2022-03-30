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
DATA q<>+0(SB)/8, $8063698428123676673
DATA q<>+8(SB)/8, $4764498181658371330
DATA q<>+16(SB)/8, $16051339359738796768
DATA q<>+24(SB)/8, $15273757526516850351
DATA q<>+32(SB)/8, $342900304943437392
GLOBL q<>(SB), (RODATA+NOPTR), 40
// qInv0 q'[0]
DATA qInv0<>(SB)/8, $8083954730842193919
GLOBL qInv0<>(SB), (RODATA+NOPTR), $8
// add(res, xPtr, yPtr *Element)
TEXT ·add(SB), NOSPLIT, $0-24
	LDP x+8(FP), (R5, R6)

	// load operands and add mod 2^r
	LDP  0(R5), (R0, R7)
	LDP  0(R6), (R1, R8)
	ADDS R0, R7, R0
	ADCS R1, R8, R1
	LDP  16(R5), (R2, R7)
	LDP  16(R6), (R3, R8)
	ADCS R2, R7, R2
	ADCS R3, R8, R3
	MOVD 32(R5), R4       // can't import these in pairs
	MOVD 32(R6), R7
	ADCS R4, R7, R4

	// load modulus and subtract
	LDP  q<>+0(SB), (R5, R6)
	ADCS R5, R0, R5
	ADCS R6, R0, R6
	LDP  q<>+16(SB), (R7, R8)
	ADCS R7, R2, R7
	ADCS R8, R2, R8
	MOVD q<>+32(SB), R9
	ADCS R9, R4, R9

	// reduce if necessary
	CSEL CS, R5, R0, R0
	CSEL CS, R6, R1, R1
	CSEL CS, R7, R2, R2
	CSEL CS, R8, R3, R3
	CSEL CS, R9, R4, R4

	// store
	MOVD z+0(FP), R5
	STP  (R0, R1), 0(R5)
	STP  (R2, R3), 16(R5)
	MOVD R4, 32(R5)
	RET
