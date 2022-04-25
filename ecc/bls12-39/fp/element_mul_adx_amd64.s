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
DATA q<>+0(SB)/8, $0x0000004c0ee3eef7
GLOBL q<>(SB), (RODATA+NOPTR), $8

// qInv0 q'[0]
DATA qInv0<>(SB)/8, $0xcce1bac4513ccd39
GLOBL qInv0<>(SB), (RODATA+NOPTR), $8

#define REDUCE(ra0, rb0) \
	MOVQ    ra0, rb0;     \
	SUBQ    q<>(SB), ra0; \
	CMOVQCS rb0, ra0;     \

// mul(res, x, y *Element)
TEXT ·mul(SB), NOSPLIT, $0-24

	// the algorithm is described here
	// https://hackmd.io/@gnark/modular_multiplication
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

	MOVQ x+8(FP), R13

	// x -> CX
	MOVQ 0(R13), CX
	MOVQ y+16(FP), BX

	// A -> BP
	// t -> R14
	// clear the flags
	XORQ AX, AX
	MOVQ 0(BX), DX

	// (A,t[0])  := x[0]*y[0] + A
	MULXQ CX, R14, BP

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADOXQ AX, BP

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R14, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, SI
	ADCXQ R14, AX
	MOVQ  SI, R14

	// t[0] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R14
	ADOXQ BP, R14

	// reduce element(R14) using temp registers (DI)
	REDUCE(R14,DI)

	MOVQ res+0(FP), AX
	MOVQ R14, 0(AX)
	RET

TEXT ·fromMont(SB), NOSPLIT, $0-8

	// the algorithm is described here
	// https://hackmd.io/@gnark/modular_multiplication
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
	XORQ DX, DX

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R14, DX
	XORQ  AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, BP
	ADCXQ R14, AX
	MOVQ  BP, R14
	MOVQ  $0, AX
	ADCXQ AX, R14
	ADOXQ AX, R14

	// reduce element(R14) using temp registers (R13)
	REDUCE(R14,R13)

	MOVQ res+0(FP), AX
	MOVQ R14, 0(AX)
	RET
