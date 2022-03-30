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
DATA q<>+0(SB)/8, $13402431016077863595
DATA q<>+8(SB)/8, $2210141511517208575
DATA q<>+16(SB)/8, $7435674573564081700
DATA q<>+24(SB)/8, $7239337960414712511
DATA q<>+32(SB)/8, $5412103778470702295
DATA q<>+40(SB)/8, $1873798617647539866
GLOBL q<>(SB), (RODATA+NOPTR), 48
// qInv0 q'[0]
DATA qInv0<>(SB)/8, $9940570264628428797
GLOBL qInv0<>(SB), (RODATA+NOPTR), $8
// add(res, xPtr, yPtr *Element)
TEXT ·add(SB), NOSPLIT, $0-24
	LDP x+8(FP), (R6, R7)

	// load operands and add mod 2^r
	LDP  0(R6), (R0, R8)
	LDP  0(R7), (R1, R9)
	ADDS R0, R8, R0
	ADCS R1, R9, R1
	LDP  16(R6), (R2, R8)
	LDP  16(R7), (R3, R9)
	ADCS R2, R8, R2
	ADCS R3, R9, R3
	LDP  32(R6), (R4, R8)
	LDP  32(R7), (R5, R9)
	ADCS R4, R8, R4
	ADCS R5, R9, R5

	// load modulus and subtract
	LDP  q<>+0(SB), (R6, R7)
	ADCS R6, R0, R6
	ADCS R7, R0, R7
	LDP  q<>+16(SB), (R8, R9)
	ADCS R8, R2, R8
	ADCS R9, R2, R9
	LDP  q<>+32(SB), (R10, R11)
	ADCS R10, R4, R10
	ADCS R11, R4, R11

	// reduce if necessary
	CSEL CS, R6, R0, R0
	CSEL CS, R7, R1, R1
	CSEL CS, R8, R2, R2
	CSEL CS, R9, R3, R3
	CSEL CS, R10, R4, R4
	CSEL CS, R11, R5, R5

	// store
	MOVD z+0(FP), R6
	STP  (R0, R1), 0(R6)
	STP  (R2, R3), 16(R6)
	STP  (R4, R5), 32(R6)
	RET
