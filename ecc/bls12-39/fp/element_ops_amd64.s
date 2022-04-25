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

// add(res, x, y *Element)
TEXT ·add(SB), NOSPLIT, $0-24
	MOVQ x+8(FP), AX
	MOVQ 0(AX), CX
	MOVQ y+16(FP), DX
	ADDQ 0(DX), CX

	// reduce element(CX) using temp registers (BX)
	REDUCE(CX,BX)

	MOVQ res+0(FP), SI
	MOVQ CX, 0(SI)
	RET

// sub(res, x, y *Element)
TEXT ·sub(SB), NOSPLIT, $0-24
	XORQ    CX, CX
	MOVQ    x+8(FP), DX
	MOVQ    0(DX), AX
	MOVQ    y+16(FP), DX
	SUBQ    0(DX), AX
	MOVQ    $0x0000004c0ee3eef7, BX
	CMOVQCC CX, BX
	ADDQ    BX, AX
	MOVQ    res+0(FP), SI
	MOVQ    AX, 0(SI)
	RET

// double(res, x *Element)
TEXT ·double(SB), NOSPLIT, $0-16
	MOVQ x+8(FP), AX
	MOVQ 0(AX), DX
	ADDQ DX, DX

	// reduce element(DX) using temp registers (CX)
	REDUCE(DX,CX)

	MOVQ res+0(FP), BX
	MOVQ DX, 0(BX)
	RET

// neg(res, x *Element)
TEXT ·neg(SB), NOSPLIT, $0-16
	MOVQ  res+0(FP), CX
	MOVQ  x+8(FP), AX
	MOVQ  0(AX), DX
	MOVQ  DX, AX
	TESTQ AX, AX
	JEQ   l1
	MOVQ  $0x0000004c0ee3eef7, BX
	SUBQ  DX, BX
	MOVQ  BX, 0(CX)
	RET

l1:
	MOVQ AX, 0(CX)
	RET

TEXT ·reduce(SB), NOSPLIT, $0-8
	MOVQ res+0(FP), AX
	MOVQ 0(AX), DX

	// reduce element(DX) using temp registers (CX)
	REDUCE(DX,CX)

	MOVQ DX, 0(AX)
	RET

// MulBy3(x *Element)
TEXT ·MulBy3(SB), NOSPLIT, $0-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	ADDQ DX, DX

	// reduce element(DX) using temp registers (CX)
	REDUCE(DX,CX)

	ADDQ 0(AX), DX

	// reduce element(DX) using temp registers (BX)
	REDUCE(DX,BX)

	MOVQ DX, 0(AX)
	RET

// MulBy5(x *Element)
TEXT ·MulBy5(SB), NOSPLIT, $0-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	ADDQ DX, DX

	// reduce element(DX) using temp registers (CX)
	REDUCE(DX,CX)

	ADDQ DX, DX

	// reduce element(DX) using temp registers (BX)
	REDUCE(DX,BX)

	ADDQ 0(AX), DX

	// reduce element(DX) using temp registers (SI)
	REDUCE(DX,SI)

	MOVQ DX, 0(AX)
	RET

// MulBy13(x *Element)
TEXT ·MulBy13(SB), NOSPLIT, $0-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	ADDQ DX, DX

	// reduce element(DX) using temp registers (CX)
	REDUCE(DX,CX)

	ADDQ DX, DX

	// reduce element(DX) using temp registers (BX)
	REDUCE(DX,BX)

	MOVQ DX, BX
	ADDQ DX, DX

	// reduce element(DX) using temp registers (CX)
	REDUCE(DX,CX)

	ADDQ BX, DX

	// reduce element(DX) using temp registers (CX)
	REDUCE(DX,CX)

	ADDQ 0(AX), DX

	// reduce element(DX) using temp registers (CX)
	REDUCE(DX,CX)

	MOVQ DX, 0(AX)
	RET

// Butterfly(a, b *Element) sets a = a + b; b = a - b
TEXT ·Butterfly(SB), NOSPLIT, $0-16
	MOVQ    a+0(FP), AX
	MOVQ    0(AX), CX
	MOVQ    CX, BX
	XORQ    AX, AX
	MOVQ    b+8(FP), DX
	ADDQ    0(DX), CX
	SUBQ    0(DX), BX
	MOVQ    $0x0000004c0ee3eef7, SI
	CMOVQCC AX, SI
	ADDQ    SI, BX
	MOVQ    BX, 0(DX)

	// reduce element(CX) using temp registers (BX)
	REDUCE(CX,BX)

	MOVQ a+0(FP), AX
	MOVQ CX, 0(AX)
	RET
