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
DATA q<>+0(SB)/8, $0x8508c00000000001
DATA q<>+8(SB)/8, $0x170b5d4430000000
DATA q<>+16(SB)/8, $0x1ef3622fba094800
DATA q<>+24(SB)/8, $0x1a22d9f300f5138f
DATA q<>+32(SB)/8, $0xc63b05c06ca1493b
DATA q<>+40(SB)/8, $0x01ae3a4617c510ea
GLOBL q<>(SB), (RODATA+NOPTR), $48

// Butterfly(a, b *Element) sets a = a + b; b = a - b
TEXT ·Butterfly(SB), NOSPLIT, $0-16
	// a[0] -> R0
	// a[1] -> R1
	// a[2] -> R2
	// a[3] -> R3
	// a[4] -> R4
	// a[5] -> R5
	// b[0] -> R0
	// b[1] -> R1
	// b[2] -> R2
	// b[3] -> R3
	// b[4] -> R4
	// b[5] -> R5
	// t[0] -> R19
	// t[1] -> R20
	// t[2] -> R21
	// t[3] -> R22
	// t[4] -> R23
	// t[5] -> R24
	LDP  a+0(FP), (R25, R26)
	LDP  0(R25), (R0, R1)
	LDP  16(R25), (R2, R3)
	LDP  32(R25), (R4, R5)
	MOVD R0, R19
	MOVD R1, R20
	MOVD R2, R21
	MOVD R3, R22
	MOVD R4, R23
	MOVD R5, R24
	LDP  0(R26), (R6, R7)
	LDP  16(R26), (R8, R9)
	LDP  32(R26), (R10, R11)

	// q[0] -> R12
	// q[1] -> R13
	// q[2] -> R14
	// q[3] -> R15
	// q[4] -> R16
	// q[5] -> R17
	LDP  q<>+0(SB), (R12, R13)
	LDP  q<>+16(SB), (R14, R15)
	LDP  q<>+32(SB), (R16, R17)
	ADDS R0, R6, R0
	ADCS R1, R7, R1
	ADCS R2, R8, R2
	ADCS R3, R9, R3
	ADCS R4, R10, R4
	ADCS R5, R11, R5
	SUBS R12, R0, R12
	SBCS R13, R1, R13
	SBCS R14, R2, R14
	SBCS R15, R3, R15
	SBCS R16, R4, R16
	SBCS R17, R5, R17
	CSEL CS, R12, R0, R0
	CSEL CS, R13, R1, R1
	CSEL CS, R14, R2, R2
	CSEL CS, R15, R3, R3
	CSEL CS, R16, R4, R4
	CSEL CS, R17, R5, R5
	LDP  q<>+0(SB), (R12, R13)
	LDP  q<>+16(SB), (R14, R15)
	LDP  q<>+32(SB), (R16, R17)
	STP  (R0, R1), 0(R25)
	STP  (R2, R3), 16(R25)
	STP  (R4, R5), 32(R25)
	SUBS R6, R19, R6
	SBCS R7, R20, R7
	SBCS R8, R21, R8
	SBCS R9, R22, R9
	SBCS R10, R23, R10
	SBCS R11, R24, R11
	CSEL CS, ZR, R12, R0
	CSEL CS, ZR, R13, R1
	CSEL CS, ZR, R14, R2
	CSEL CS, ZR, R15, R3
	CSEL CS, ZR, R16, R4
	CSEL CS, ZR, R17, R5
	ADDS R0, R6, R6
	ADCS R1, R7, R7
	ADCS R2, R8, R8
	ADCS R3, R9, R9
	ADCS R4, R10, R10
	ADCS R5, R11, R11
	STP  (R6, R7), 0(R26)
	STP  (R8, R9), 16(R26)
	STP  (R10, R11), 32(R26)
	RET
