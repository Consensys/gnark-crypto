// +build amd64_adx

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
DATA q<>+0(SB)/8, $0xdd64f813fcb4c2a1
DATA q<>+8(SB)/8, $0x1aa90fd187823ec8
GLOBL q<>(SB), (RODATA+NOPTR), $16

// qInv0 q'[0]
DATA qInv0<>(SB)/8, $0x1378842f09c45e9f
GLOBL qInv0<>(SB), (RODATA+NOPTR), $8

#define REDUCE(ra0, ra1, rb0, rb1) \
	MOVQ    ra0, rb0;       \
	SUBQ    q<>(SB), ra0;   \
	MOVQ    ra1, rb1;       \
	SBBQ    q<>+8(SB), ra1; \
	CMOVQCS rb0, ra0;       \
	CMOVQCS rb1, ra1;       \

// mul(res, x, y *Element)
TEXT ·mul(SB), NOSPLIT, $0-24

	// the algorithm is described here
	// https://hackmd.io/@zkteam/modular_multiplication
	// however, to benefit from the ADCX and ADOX carry chains
	// we split the inner loops in 2:
	// for i=0 to N-1
	// 		for j=0 to N-1
	// 		    (A,t[j])  := t[j] + x[j]*y[i] + A
	// 		m := t[0]*q'[0] mod W
	// 		C,_ := t[0] + m*q[0]
	// 		for j=1 to N-1
	// 		    (C,t[j-1]) := t[j] + m*q[j] + C
	// 		t[N-1] = C + A

	MOVQ x+8(FP), CX

	// x[0] -> BX
	// x[1] -> SI
	MOVQ 0(CX), BX
	MOVQ 8(CX), SI
	MOVQ y+16(FP), DI

	// A -> BP
	// t[0] -> R14
	// t[1] -> R15
	// clear the flags
	XORQ AX, AX
	MOVQ 0(DI), DX

	// (A,t[0])  := x[0]*y[0] + A
	MULXQ BX, R14, R15

	// (A,t[1])  := x[1]*y[0] + A
	MULXQ SI, AX, BP
	ADOXQ AX, R15

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADOXQ AX, BP

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R14, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, R8
	ADCXQ R14, AX
	MOVQ  R8, R14

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R15, R14
	MULXQ q<>+8(SB), AX, R15
	ADOXQ AX, R14

	// t[1] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R15
	ADOXQ BP, R15

	// clear the flags
	XORQ AX, AX
	MOVQ 8(DI), DX

	// (A,t[0])  := t[0] + x[0]*y[1] + A
	MULXQ BX, AX, BP
	ADOXQ AX, R14

	// (A,t[1])  := t[1] + x[1]*y[1] + A
	ADCXQ BP, R15
	MULXQ SI, AX, BP
	ADOXQ AX, R15

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, BP
	ADOXQ AX, BP

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R14, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, R8
	ADCXQ R14, AX
	MOVQ  R8, R14

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R15, R14
	MULXQ q<>+8(SB), AX, R15
	ADOXQ AX, R14

	// t[1] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R15
	ADOXQ BP, R15

	// reduce element(R14,R15) using temp registers (R9,R10)
	REDUCE(R14,R15,R9,R10)

	MOVQ res+0(FP), AX
	MOVQ R14, 0(AX)
	MOVQ R15, 8(AX)
	RET

TEXT ·fromMont(SB), NOSPLIT, $0-8

	// the algorithm is described here
	// https://hackmd.io/@zkteam/modular_multiplication
	// when y = 1 we have:
	// for i=0 to N-1
	// 		t[i] = x[i]
	// for i=0 to N-1
	// 		m := t[0]*q'[0] mod W
	// 		C,_ := t[0] + m*q[0]
	// 		for j=1 to N-1
	// 		    (C,t[j-1]) := t[j] + m*q[j] + C
	// 		t[N-1] = C
	MOVQ res+0(FP), DX
	MOVQ 0(DX), R14
	MOVQ 8(DX), R15
	XORQ DX, DX

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R14, DX
	XORQ  AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, BP
	ADCXQ R14, AX
	MOVQ  BP, R14

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R15, R14
	MULXQ q<>+8(SB), AX, R15
	ADOXQ AX, R14
	MOVQ  $0, AX
	ADCXQ AX, R15
	ADOXQ AX, R15
	XORQ  DX, DX

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R14, DX
	XORQ  AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, BP
	ADCXQ R14, AX
	MOVQ  BP, R14

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R15, R14
	MULXQ q<>+8(SB), AX, R15
	ADOXQ AX, R14
	MOVQ  $0, AX
	ADCXQ AX, R15
	ADOXQ AX, R15

	// reduce element(R14,R15) using temp registers (CX,BX)
	REDUCE(R14,R15,CX,BX)

	MOVQ res+0(FP), AX
	MOVQ R14, 0(AX)
	MOVQ R15, 8(AX)
	RET
