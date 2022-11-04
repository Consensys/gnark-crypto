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
DATA q<>+0(SB)/8, $0x3291440000000001
DATA q<>+8(SB)/8, $0xeae77f3da0940001
DATA q<>+16(SB)/8, $0x87787fb4e3dbb0ff
DATA q<>+24(SB)/8, $0x20e7b9c8ef7b2eb1
GLOBL q<>(SB), (RODATA+NOPTR), $32

// Butterfly(a, b *Element) sets a = a + b; b = a - b
TEXT ·Butterfly(SB), NOSPLIT, $0-16
	// a[0] -> R0
	// a[1] -> R1
	// a[2] -> R2
	// a[3] -> R3
	// b[0] -> R0
	// b[1] -> R1
	// b[2] -> R2
	// b[3] -> R3
	// t[0] -> R12
	// t[1] -> R13
	// t[2] -> R14
	// t[3] -> R15
	LDP  a+0(FP), (R16, R17)
	LDP  0(R16), (R0, R1)
	LDP  16(R16), (R2, R3)
	MOVD R0, R12
	MOVD R1, R13
	MOVD R2, R14
	MOVD R3, R15
	LDP  0(R17), (R4, R5)
	LDP  16(R17), (R6, R7)

	// q[0] -> R8
	// q[1] -> R9
	// q[2] -> R10
	// q[3] -> R11
	LDP  q<>+0(SB), (R8, R9)
	LDP  q<>+16(SB), (R10, R11)
	ADDS R0, R4, R0
	ADCS R1, R5, R1
	ADCS R2, R6, R2
	ADCS R3, R7, R3
	SUBS R8, R0, R8
	SBCS R9, R1, R9
	SBCS R10, R2, R10
	SBCS R11, R3, R11
	CSEL CS, R8, R0, R0
	CSEL CS, R9, R1, R1
	CSEL CS, R10, R2, R2
	CSEL CS, R11, R3, R3
	LDP  q<>+0(SB), (R8, R9)
	LDP  q<>+16(SB), (R10, R11)
	STP  (R0, R1), 0(R16)
	STP  (R2, R3), 16(R16)
	SUBS R4, R12, R4
	SBCS R5, R13, R5
	SBCS R6, R14, R6
	SBCS R7, R15, R7
	CSEL CS, ZR, R8, R0
	CSEL CS, ZR, R9, R1
	CSEL CS, ZR, R10, R2
	CSEL CS, ZR, R11, R3
	ADDS R0, R4, R4
	ADCS R1, R5, R5
	ADCS R2, R6, R6
	ADCS R3, R7, R7
	STP  (R4, R5), 0(R17)
	STP  (R6, R7), 16(R17)
	RET
