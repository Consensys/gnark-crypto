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
DATA q<>+0(SB)/8, $0x0000004c0ee3eef7
GLOBL q<>(SB), (RODATA+NOPTR), $8

// qInv0 q'[0]
DATA qInv0<>(SB)/8, $0xcce1bac4513ccd39
GLOBL qInv0<>(SB), (RODATA+NOPTR), $8

#define REDUCE(ra0, rb0) \
	MOVQ    ra0, rb0;     \
	SUBQ    q<>(SB), ra0; \
	CMOVQCS rb0, ra0;     \

TEXT ·addE2(SB), NOSPLIT, $0-24
	MOVQ x+8(FP), AX
	MOVQ 0(AX), BX
	MOVQ y+16(FP), DX
	ADDQ 0(DX), BX

	// reduce element(BX) using temp registers (SI)
	REDUCE(BX,SI)

	MOVQ res+0(FP), CX
	MOVQ BX, 0(CX)
	MOVQ 8(AX), BX
	ADDQ 8(DX), BX

	// reduce element(BX) using temp registers (DI)
	REDUCE(BX,DI)

	MOVQ BX, 8(CX)
	RET

TEXT ·doubleE2(SB), NOSPLIT, $0-16
	MOVQ res+0(FP), DX
	MOVQ x+8(FP), AX
	MOVQ 0(AX), CX
	ADDQ CX, CX

	// reduce element(CX) using temp registers (BX)
	REDUCE(CX,BX)

	MOVQ CX, 0(DX)
	MOVQ 8(AX), CX
	ADDQ CX, CX

	// reduce element(CX) using temp registers (SI)
	REDUCE(CX,SI)

	MOVQ CX, 8(DX)
	RET

TEXT ·subE2(SB), NOSPLIT, $0-24
	XORQ    CX, CX
	MOVQ    x+8(FP), DX
	MOVQ    0(DX), AX
	MOVQ    y+16(FP), DX
	SUBQ    0(DX), AX
	MOVQ    x+8(FP), DX
	MOVQ    $0x0000004c0ee3eef7, BX
	CMOVQCC CX, BX
	ADDQ    BX, AX
	MOVQ    res+0(FP), SI
	MOVQ    AX, 0(SI)
	MOVQ    8(DX), AX
	MOVQ    y+16(FP), DX
	SUBQ    8(DX), AX
	MOVQ    $0x0000004c0ee3eef7, DI
	CMOVQCC CX, DI
	ADDQ    DI, AX
	MOVQ    res+0(FP), DX
	MOVQ    AX, 8(DX)
	RET

TEXT ·negE2(SB), NOSPLIT, $0-16
	MOVQ  res+0(FP), DX
	MOVQ  x+8(FP), AX
	MOVQ  0(AX), BX
	MOVQ  BX, AX
	TESTQ AX, AX
	JNE   l1
	MOVQ  AX, 0(DX)
	JMP   l3

l1:
	MOVQ $0x0000004c0ee3eef7, CX
	SUBQ BX, CX
	MOVQ CX, 0(DX)

l3:
	MOVQ  x+8(FP), AX
	MOVQ  8(AX), BX
	MOVQ  BX, AX
	TESTQ AX, AX
	JNE   l2
	MOVQ  AX, 8(DX)
	RET

l2:
	MOVQ $0x0000004c0ee3eef7, CX
	SUBQ BX, CX
	MOVQ CX, 8(DX)
	RET
