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
DATA q<>+0(SB)/8, $0xaef61a92e0dcf7eb
DATA q<>+8(SB)/8, $1
GLOBL q<>(SB), (RODATA+NOPTR), $16

// qInv0 q'[0]
DATA qInv0<>(SB)/8, $0x17ae998c42d4873d
GLOBL qInv0<>(SB), (RODATA+NOPTR), $8

#define REDUCE(ra0, ra1, rb0, rb1) \
	MOVQ    ra0, rb0;       \
	SUBQ    q<>(SB), ra0;   \
	MOVQ    ra1, rb1;       \
	SBBQ    q<>+8(SB), ra1; \
	CMOVQCS rb0, ra0;       \
	CMOVQCS rb1, ra1;       \

TEXT ·reduce(SB), NOSPLIT, $0-8
	MOVQ res+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX

	// reduce element(DX,CX) using temp registers (BX,SI)
	REDUCE(DX,CX,BX,SI)

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	RET

// MulBy3(x *Element)
TEXT ·MulBy3(SB), NOSPLIT, $0-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	ADDQ DX, DX
	ADCQ CX, CX

	// reduce element(DX,CX) using temp registers (BX,SI)
	REDUCE(DX,CX,BX,SI)

	ADDQ 0(AX), DX
	ADCQ 8(AX), CX

	// reduce element(DX,CX) using temp registers (DI,R8)
	REDUCE(DX,CX,DI,R8)

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	RET

// MulBy5(x *Element)
TEXT ·MulBy5(SB), NOSPLIT, $0-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	ADDQ DX, DX
	ADCQ CX, CX

	// reduce element(DX,CX) using temp registers (BX,SI)
	REDUCE(DX,CX,BX,SI)

	ADDQ DX, DX
	ADCQ CX, CX

	// reduce element(DX,CX) using temp registers (DI,R8)
	REDUCE(DX,CX,DI,R8)

	ADDQ 0(AX), DX
	ADCQ 8(AX), CX

	// reduce element(DX,CX) using temp registers (R9,R10)
	REDUCE(DX,CX,R9,R10)

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	RET

// MulBy13(x *Element)
TEXT ·MulBy13(SB), NOSPLIT, $0-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	ADDQ DX, DX
	ADCQ CX, CX

	// reduce element(DX,CX) using temp registers (BX,SI)
	REDUCE(DX,CX,BX,SI)

	ADDQ DX, DX
	ADCQ CX, CX

	// reduce element(DX,CX) using temp registers (DI,R8)
	REDUCE(DX,CX,DI,R8)

	MOVQ DX, DI
	MOVQ CX, R8
	ADDQ DX, DX
	ADCQ CX, CX

	// reduce element(DX,CX) using temp registers (BX,SI)
	REDUCE(DX,CX,BX,SI)

	ADDQ DI, DX
	ADCQ R8, CX

	// reduce element(DX,CX) using temp registers (BX,SI)
	REDUCE(DX,CX,BX,SI)

	ADDQ 0(AX), DX
	ADCQ 8(AX), CX

	// reduce element(DX,CX) using temp registers (BX,SI)
	REDUCE(DX,CX,BX,SI)

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	RET

// Butterfly(a, b *Element) sets a = a + b; b = a - b
TEXT ·Butterfly(SB), NOSPLIT, $0-16
	MOVQ    a+0(FP), AX
	MOVQ    0(AX), CX
	MOVQ    8(AX), BX
	MOVQ    CX, SI
	MOVQ    BX, DI
	XORQ    AX, AX
	MOVQ    b+8(FP), DX
	ADDQ    0(DX), CX
	ADCQ    8(DX), BX
	SUBQ    0(DX), SI
	SBBQ    8(DX), DI
	MOVQ    $0xaef61a92e0dcf7eb, R8
	MOVQ    $1, R9
	CMOVQCC AX, R8
	CMOVQCC AX, R9
	ADDQ    R8, SI
	ADCQ    R9, DI
	MOVQ    SI, 0(DX)
	MOVQ    DI, 8(DX)

	// reduce element(CX,BX) using temp registers (SI,DI)
	REDUCE(CX,BX,SI,DI)

	MOVQ a+0(FP), AX
	MOVQ CX, 0(AX)
	MOVQ BX, 8(AX)
	RET