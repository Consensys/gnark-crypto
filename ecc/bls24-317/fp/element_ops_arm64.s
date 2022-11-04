// +build !purego

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
DATA q<>+0(SB)/8, $0x8d512e565dab2aab
DATA q<>+8(SB)/8, $0xd6f339e43424bf7e
DATA q<>+16(SB)/8, $0x169a61e684c73446
DATA q<>+24(SB)/8, $0xf28fc5a0b7f9d039
DATA q<>+32(SB)/8, $0x1058ca226f60892c
GLOBL q<>(SB), (RODATA+NOPTR), $40

// Butterfly(a, b *Element) sets a = a + b; b = a - b
TEXT ·Butterfly(SB), NOSPLIT, $0-16
	// a[0] -> R0
	// a[1] -> R1
	// a[2] -> R2
	// a[3] -> R3
	// a[4] -> R4
	// b[0] -> R0
	// b[1] -> R1
	// b[2] -> R2
	// b[3] -> R3
	// b[4] -> R4
	// t[0] -> R15
	// t[1] -> R16
	// t[2] -> R17
	// t[3] -> R19
	// t[4] -> R20
	LDP  a+0(FP), (R21, R22)
	LDP  0(R21), (R0, R1)
	LDP  16(R21), (R2, R3)
	MOVD 32(R21), R4
	MOVD R0, R15
	MOVD R1, R16
	MOVD R2, R17
	MOVD R3, R19
	MOVD R4, R20
	LDP  0(R22), (R5, R6)
	LDP  16(R22), (R7, R8)
	MOVD 32(R22), R9

	// q[0] -> R10
	// q[1] -> R11
	// q[2] -> R12
	// q[3] -> R13
	// q[4] -> R14
	LDP  q<>+0(SB), (R10, R11)
	LDP  q<>+16(SB), (R12, R13)
	MOVD q<>+32(SB), R14
	ADDS R0, R5, R0
	ADCS R1, R6, R1
	ADCS R2, R7, R2
	ADCS R3, R8, R3
	ADCS R4, R9, R4
	SUBS R10, R0, R10
	SBCS R11, R1, R11
	SBCS R12, R2, R12
	SBCS R13, R3, R13
	SBCS R14, R4, R14
	CSEL CS, R10, R0, R0
	CSEL CS, R11, R1, R1
	CSEL CS, R12, R2, R2
	CSEL CS, R13, R3, R3
	CSEL CS, R14, R4, R4
	LDP  q<>+0(SB), (R10, R11)
	LDP  q<>+16(SB), (R12, R13)
	MOVD q<>+32(SB), R14
	STP  (R0, R1), 0(R21)
	STP  (R2, R3), 16(R21)
	MOVD R4, 32(R21)
	SUBS R5, R15, R5
	SBCS R6, R16, R6
	SBCS R7, R17, R7
	SBCS R8, R19, R8
	SBCS R9, R20, R9
	CSEL CS, ZR, R10, R0
	CSEL CS, ZR, R11, R1
	CSEL CS, ZR, R12, R2
	CSEL CS, ZR, R13, R3
	CSEL CS, ZR, R14, R4
	ADDS R0, R5, R5
	ADCS R1, R6, R6
	ADCS R2, R7, R7
	ADCS R3, R8, R8
	ADCS R4, R9, R9
	STP  (R5, R6), 0(R22)
	STP  (R7, R8), 16(R22)
	MOVD R9, 32(R22)
	RET
