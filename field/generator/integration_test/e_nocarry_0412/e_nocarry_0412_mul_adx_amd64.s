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
DATA q<>+0(SB)/8, $0x849656216c2cf681
DATA q<>+8(SB)/8, $0x691cc81cc5d9ebea
DATA q<>+16(SB)/8, $0x711b4b5e2ef3e0b9
DATA q<>+24(SB)/8, $0xacafe20d548589a1
DATA q<>+32(SB)/8, $0x293f455fd8a28991
DATA q<>+40(SB)/8, $0xef8c0d8cacbdb898
DATA q<>+48(SB)/8, $0x000000000e45ed61
GLOBL q<>(SB), (RODATA+NOPTR), $56

// qInv0 q'[0]
DATA qInv0<>(SB)/8, $0x2045aaeb1972b67f
GLOBL qInv0<>(SB), (RODATA+NOPTR), $8

#define REDUCE(ra0, ra1, ra2, ra3, ra4, ra5, ra6, rb0, rb1, rb2, rb3, rb4, rb5, rb6) \
	MOVQ    ra0, rb0;        \
	SUBQ    q<>(SB), ra0;    \
	MOVQ    ra1, rb1;        \
	SBBQ    q<>+8(SB), ra1;  \
	MOVQ    ra2, rb2;        \
	SBBQ    q<>+16(SB), ra2; \
	MOVQ    ra3, rb3;        \
	SBBQ    q<>+24(SB), ra3; \
	MOVQ    ra4, rb4;        \
	SBBQ    q<>+32(SB), ra4; \
	MOVQ    ra5, rb5;        \
	SBBQ    q<>+40(SB), ra5; \
	MOVQ    ra6, rb6;        \
	SBBQ    q<>+48(SB), ra6; \
	CMOVQCS rb0, ra0;        \
	CMOVQCS rb1, ra1;        \
	CMOVQCS rb2, ra2;        \
	CMOVQCS rb3, ra3;        \
	CMOVQCS rb4, ra4;        \
	CMOVQCS rb5, ra5;        \
	CMOVQCS rb6, ra6;        \

// mul(res, x, y *Element)
TEXT ·mul(SB), $16-24

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

	NO_LOCAL_POINTERS
	MOVQ x+8(FP), R9

	// x[0] -> R11
	// x[1] -> R12
	MOVQ 0(R9), R11
	MOVQ 8(R9), R12
	MOVQ y+16(FP), R13

	// A -> BP
	// t[0] -> R14
	// t[1] -> R15
	// t[2] -> CX
	// t[3] -> BX
	// t[4] -> SI
	// t[5] -> DI
	// t[6] -> R8
	// clear the flags
	XORQ AX, AX
	MOVQ 0(R13), DX

	// (A,t[0])  := x[0]*y[0] + A
	MULXQ R11, R14, R15

	// (A,t[1])  := x[1]*y[0] + A
	MULXQ R12, AX, CX
	ADOXQ AX, R15

	// (A,t[2])  := x[2]*y[0] + A
	MULXQ 16(R9), AX, BX
	ADOXQ AX, CX

	// (A,t[3])  := x[3]*y[0] + A
	MULXQ 24(R9), AX, SI
	ADOXQ AX, BX

	// (A,t[4])  := x[4]*y[0] + A
	MULXQ 32(R9), AX, DI
	ADOXQ AX, SI

	// (A,t[5])  := x[5]*y[0] + A
	MULXQ 40(R9), AX, R8
	ADOXQ AX, DI

	// (A,t[6])  := x[6]*y[0] + A
	MULXQ 48(R9), AX, BP
	ADOXQ AX, R8

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADOXQ AX, BP

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R14, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, R10
	ADCXQ R14, AX
	MOVQ  R10, R14

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R15, R14
	MULXQ q<>+8(SB), AX, R15
	ADOXQ AX, R14

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ CX, R15
	MULXQ q<>+16(SB), AX, CX
	ADOXQ AX, R15

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BX, CX
	MULXQ q<>+24(SB), AX, BX
	ADOXQ AX, CX

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ SI, BX
	MULXQ q<>+32(SB), AX, SI
	ADOXQ AX, BX

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ DI, SI
	MULXQ q<>+40(SB), AX, DI
	ADOXQ AX, SI

	// (C,t[5]) := t[6] + m*q[6] + C
	ADCXQ R8, DI
	MULXQ q<>+48(SB), AX, R8
	ADOXQ AX, DI

	// t[6] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R8
	ADOXQ BP, R8

	// clear the flags
	XORQ AX, AX
	MOVQ 8(R13), DX

	// (A,t[0])  := t[0] + x[0]*y[1] + A
	MULXQ R11, AX, BP
	ADOXQ AX, R14

	// (A,t[1])  := t[1] + x[1]*y[1] + A
	ADCXQ BP, R15
	MULXQ R12, AX, BP
	ADOXQ AX, R15

	// (A,t[2])  := t[2] + x[2]*y[1] + A
	ADCXQ BP, CX
	MULXQ 16(R9), AX, BP
	ADOXQ AX, CX

	// (A,t[3])  := t[3] + x[3]*y[1] + A
	ADCXQ BP, BX
	MULXQ 24(R9), AX, BP
	ADOXQ AX, BX

	// (A,t[4])  := t[4] + x[4]*y[1] + A
	ADCXQ BP, SI
	MULXQ 32(R9), AX, BP
	ADOXQ AX, SI

	// (A,t[5])  := t[5] + x[5]*y[1] + A
	ADCXQ BP, DI
	MULXQ 40(R9), AX, BP
	ADOXQ AX, DI

	// (A,t[6])  := t[6] + x[6]*y[1] + A
	ADCXQ BP, R8
	MULXQ 48(R9), AX, BP
	ADOXQ AX, R8

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
	MULXQ q<>+0(SB), AX, R10
	ADCXQ R14, AX
	MOVQ  R10, R14

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R15, R14
	MULXQ q<>+8(SB), AX, R15
	ADOXQ AX, R14

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ CX, R15
	MULXQ q<>+16(SB), AX, CX
	ADOXQ AX, R15

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BX, CX
	MULXQ q<>+24(SB), AX, BX
	ADOXQ AX, CX

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ SI, BX
	MULXQ q<>+32(SB), AX, SI
	ADOXQ AX, BX

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ DI, SI
	MULXQ q<>+40(SB), AX, DI
	ADOXQ AX, SI

	// (C,t[5]) := t[6] + m*q[6] + C
	ADCXQ R8, DI
	MULXQ q<>+48(SB), AX, R8
	ADOXQ AX, DI

	// t[6] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R8
	ADOXQ BP, R8

	// clear the flags
	XORQ AX, AX
	MOVQ 16(R13), DX

	// (A,t[0])  := t[0] + x[0]*y[2] + A
	MULXQ R11, AX, BP
	ADOXQ AX, R14

	// (A,t[1])  := t[1] + x[1]*y[2] + A
	ADCXQ BP, R15
	MULXQ R12, AX, BP
	ADOXQ AX, R15

	// (A,t[2])  := t[2] + x[2]*y[2] + A
	ADCXQ BP, CX
	MULXQ 16(R9), AX, BP
	ADOXQ AX, CX

	// (A,t[3])  := t[3] + x[3]*y[2] + A
	ADCXQ BP, BX
	MULXQ 24(R9), AX, BP
	ADOXQ AX, BX

	// (A,t[4])  := t[4] + x[4]*y[2] + A
	ADCXQ BP, SI
	MULXQ 32(R9), AX, BP
	ADOXQ AX, SI

	// (A,t[5])  := t[5] + x[5]*y[2] + A
	ADCXQ BP, DI
	MULXQ 40(R9), AX, BP
	ADOXQ AX, DI

	// (A,t[6])  := t[6] + x[6]*y[2] + A
	ADCXQ BP, R8
	MULXQ 48(R9), AX, BP
	ADOXQ AX, R8

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
	MULXQ q<>+0(SB), AX, R10
	ADCXQ R14, AX
	MOVQ  R10, R14

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R15, R14
	MULXQ q<>+8(SB), AX, R15
	ADOXQ AX, R14

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ CX, R15
	MULXQ q<>+16(SB), AX, CX
	ADOXQ AX, R15

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BX, CX
	MULXQ q<>+24(SB), AX, BX
	ADOXQ AX, CX

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ SI, BX
	MULXQ q<>+32(SB), AX, SI
	ADOXQ AX, BX

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ DI, SI
	MULXQ q<>+40(SB), AX, DI
	ADOXQ AX, SI

	// (C,t[5]) := t[6] + m*q[6] + C
	ADCXQ R8, DI
	MULXQ q<>+48(SB), AX, R8
	ADOXQ AX, DI

	// t[6] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R8
	ADOXQ BP, R8

	// clear the flags
	XORQ AX, AX
	MOVQ 24(R13), DX

	// (A,t[0])  := t[0] + x[0]*y[3] + A
	MULXQ R11, AX, BP
	ADOXQ AX, R14

	// (A,t[1])  := t[1] + x[1]*y[3] + A
	ADCXQ BP, R15
	MULXQ R12, AX, BP
	ADOXQ AX, R15

	// (A,t[2])  := t[2] + x[2]*y[3] + A
	ADCXQ BP, CX
	MULXQ 16(R9), AX, BP
	ADOXQ AX, CX

	// (A,t[3])  := t[3] + x[3]*y[3] + A
	ADCXQ BP, BX
	MULXQ 24(R9), AX, BP
	ADOXQ AX, BX

	// (A,t[4])  := t[4] + x[4]*y[3] + A
	ADCXQ BP, SI
	MULXQ 32(R9), AX, BP
	ADOXQ AX, SI

	// (A,t[5])  := t[5] + x[5]*y[3] + A
	ADCXQ BP, DI
	MULXQ 40(R9), AX, BP
	ADOXQ AX, DI

	// (A,t[6])  := t[6] + x[6]*y[3] + A
	ADCXQ BP, R8
	MULXQ 48(R9), AX, BP
	ADOXQ AX, R8

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
	MULXQ q<>+0(SB), AX, R10
	ADCXQ R14, AX
	MOVQ  R10, R14

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R15, R14
	MULXQ q<>+8(SB), AX, R15
	ADOXQ AX, R14

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ CX, R15
	MULXQ q<>+16(SB), AX, CX
	ADOXQ AX, R15

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BX, CX
	MULXQ q<>+24(SB), AX, BX
	ADOXQ AX, CX

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ SI, BX
	MULXQ q<>+32(SB), AX, SI
	ADOXQ AX, BX

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ DI, SI
	MULXQ q<>+40(SB), AX, DI
	ADOXQ AX, SI

	// (C,t[5]) := t[6] + m*q[6] + C
	ADCXQ R8, DI
	MULXQ q<>+48(SB), AX, R8
	ADOXQ AX, DI

	// t[6] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R8
	ADOXQ BP, R8

	// clear the flags
	XORQ AX, AX
	MOVQ 32(R13), DX

	// (A,t[0])  := t[0] + x[0]*y[4] + A
	MULXQ R11, AX, BP
	ADOXQ AX, R14

	// (A,t[1])  := t[1] + x[1]*y[4] + A
	ADCXQ BP, R15
	MULXQ R12, AX, BP
	ADOXQ AX, R15

	// (A,t[2])  := t[2] + x[2]*y[4] + A
	ADCXQ BP, CX
	MULXQ 16(R9), AX, BP
	ADOXQ AX, CX

	// (A,t[3])  := t[3] + x[3]*y[4] + A
	ADCXQ BP, BX
	MULXQ 24(R9), AX, BP
	ADOXQ AX, BX

	// (A,t[4])  := t[4] + x[4]*y[4] + A
	ADCXQ BP, SI
	MULXQ 32(R9), AX, BP
	ADOXQ AX, SI

	// (A,t[5])  := t[5] + x[5]*y[4] + A
	ADCXQ BP, DI
	MULXQ 40(R9), AX, BP
	ADOXQ AX, DI

	// (A,t[6])  := t[6] + x[6]*y[4] + A
	ADCXQ BP, R8
	MULXQ 48(R9), AX, BP
	ADOXQ AX, R8

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
	MULXQ q<>+0(SB), AX, R10
	ADCXQ R14, AX
	MOVQ  R10, R14

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R15, R14
	MULXQ q<>+8(SB), AX, R15
	ADOXQ AX, R14

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ CX, R15
	MULXQ q<>+16(SB), AX, CX
	ADOXQ AX, R15

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BX, CX
	MULXQ q<>+24(SB), AX, BX
	ADOXQ AX, CX

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ SI, BX
	MULXQ q<>+32(SB), AX, SI
	ADOXQ AX, BX

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ DI, SI
	MULXQ q<>+40(SB), AX, DI
	ADOXQ AX, SI

	// (C,t[5]) := t[6] + m*q[6] + C
	ADCXQ R8, DI
	MULXQ q<>+48(SB), AX, R8
	ADOXQ AX, DI

	// t[6] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R8
	ADOXQ BP, R8

	// clear the flags
	XORQ AX, AX
	MOVQ 40(R13), DX

	// (A,t[0])  := t[0] + x[0]*y[5] + A
	MULXQ R11, AX, BP
	ADOXQ AX, R14

	// (A,t[1])  := t[1] + x[1]*y[5] + A
	ADCXQ BP, R15
	MULXQ R12, AX, BP
	ADOXQ AX, R15

	// (A,t[2])  := t[2] + x[2]*y[5] + A
	ADCXQ BP, CX
	MULXQ 16(R9), AX, BP
	ADOXQ AX, CX

	// (A,t[3])  := t[3] + x[3]*y[5] + A
	ADCXQ BP, BX
	MULXQ 24(R9), AX, BP
	ADOXQ AX, BX

	// (A,t[4])  := t[4] + x[4]*y[5] + A
	ADCXQ BP, SI
	MULXQ 32(R9), AX, BP
	ADOXQ AX, SI

	// (A,t[5])  := t[5] + x[5]*y[5] + A
	ADCXQ BP, DI
	MULXQ 40(R9), AX, BP
	ADOXQ AX, DI

	// (A,t[6])  := t[6] + x[6]*y[5] + A
	ADCXQ BP, R8
	MULXQ 48(R9), AX, BP
	ADOXQ AX, R8

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
	MULXQ q<>+0(SB), AX, R10
	ADCXQ R14, AX
	MOVQ  R10, R14

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R15, R14
	MULXQ q<>+8(SB), AX, R15
	ADOXQ AX, R14

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ CX, R15
	MULXQ q<>+16(SB), AX, CX
	ADOXQ AX, R15

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BX, CX
	MULXQ q<>+24(SB), AX, BX
	ADOXQ AX, CX

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ SI, BX
	MULXQ q<>+32(SB), AX, SI
	ADOXQ AX, BX

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ DI, SI
	MULXQ q<>+40(SB), AX, DI
	ADOXQ AX, SI

	// (C,t[5]) := t[6] + m*q[6] + C
	ADCXQ R8, DI
	MULXQ q<>+48(SB), AX, R8
	ADOXQ AX, DI

	// t[6] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R8
	ADOXQ BP, R8

	// clear the flags
	XORQ AX, AX
	MOVQ 48(R13), DX

	// (A,t[0])  := t[0] + x[0]*y[6] + A
	MULXQ R11, AX, BP
	ADOXQ AX, R14

	// (A,t[1])  := t[1] + x[1]*y[6] + A
	ADCXQ BP, R15
	MULXQ R12, AX, BP
	ADOXQ AX, R15

	// (A,t[2])  := t[2] + x[2]*y[6] + A
	ADCXQ BP, CX
	MULXQ 16(R9), AX, BP
	ADOXQ AX, CX

	// (A,t[3])  := t[3] + x[3]*y[6] + A
	ADCXQ BP, BX
	MULXQ 24(R9), AX, BP
	ADOXQ AX, BX

	// (A,t[4])  := t[4] + x[4]*y[6] + A
	ADCXQ BP, SI
	MULXQ 32(R9), AX, BP
	ADOXQ AX, SI

	// (A,t[5])  := t[5] + x[5]*y[6] + A
	ADCXQ BP, DI
	MULXQ 40(R9), AX, BP
	ADOXQ AX, DI

	// (A,t[6])  := t[6] + x[6]*y[6] + A
	ADCXQ BP, R8
	MULXQ 48(R9), AX, BP
	ADOXQ AX, R8

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
	MULXQ q<>+0(SB), AX, R10
	ADCXQ R14, AX
	MOVQ  R10, R14

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R15, R14
	MULXQ q<>+8(SB), AX, R15
	ADOXQ AX, R14

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ CX, R15
	MULXQ q<>+16(SB), AX, CX
	ADOXQ AX, R15

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BX, CX
	MULXQ q<>+24(SB), AX, BX
	ADOXQ AX, CX

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ SI, BX
	MULXQ q<>+32(SB), AX, SI
	ADOXQ AX, BX

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ DI, SI
	MULXQ q<>+40(SB), AX, DI
	ADOXQ AX, SI

	// (C,t[5]) := t[6] + m*q[6] + C
	ADCXQ R8, DI
	MULXQ q<>+48(SB), AX, R8
	ADOXQ AX, DI

	// t[6] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R8
	ADOXQ BP, R8

	// reduce element(R14,R15,CX,BX,SI,DI,R8) using temp registers (R10,R9,R13,R11,R12,s0-8(SP),s1-16(SP))
	REDUCE(R14,R15,CX,BX,SI,DI,R8,R10,R9,R13,R11,R12,s0-8(SP),s1-16(SP))

	MOVQ res+0(FP), AX
	MOVQ R14, 0(AX)
	MOVQ R15, 8(AX)
	MOVQ CX, 16(AX)
	MOVQ BX, 24(AX)
	MOVQ SI, 32(AX)
	MOVQ DI, 40(AX)
	MOVQ R8, 48(AX)
	RET

TEXT ·fromMont(SB), $16-8
	NO_LOCAL_POINTERS

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
	MOVQ 16(DX), CX
	MOVQ 24(DX), BX
	MOVQ 32(DX), SI
	MOVQ 40(DX), DI
	MOVQ 48(DX), R8
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

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ CX, R15
	MULXQ q<>+16(SB), AX, CX
	ADOXQ AX, R15

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BX, CX
	MULXQ q<>+24(SB), AX, BX
	ADOXQ AX, CX

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ SI, BX
	MULXQ q<>+32(SB), AX, SI
	ADOXQ AX, BX

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ DI, SI
	MULXQ q<>+40(SB), AX, DI
	ADOXQ AX, SI

	// (C,t[5]) := t[6] + m*q[6] + C
	ADCXQ R8, DI
	MULXQ q<>+48(SB), AX, R8
	ADOXQ AX, DI
	MOVQ  $0, AX
	ADCXQ AX, R8
	ADOXQ AX, R8
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

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ CX, R15
	MULXQ q<>+16(SB), AX, CX
	ADOXQ AX, R15

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BX, CX
	MULXQ q<>+24(SB), AX, BX
	ADOXQ AX, CX

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ SI, BX
	MULXQ q<>+32(SB), AX, SI
	ADOXQ AX, BX

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ DI, SI
	MULXQ q<>+40(SB), AX, DI
	ADOXQ AX, SI

	// (C,t[5]) := t[6] + m*q[6] + C
	ADCXQ R8, DI
	MULXQ q<>+48(SB), AX, R8
	ADOXQ AX, DI
	MOVQ  $0, AX
	ADCXQ AX, R8
	ADOXQ AX, R8
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

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ CX, R15
	MULXQ q<>+16(SB), AX, CX
	ADOXQ AX, R15

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BX, CX
	MULXQ q<>+24(SB), AX, BX
	ADOXQ AX, CX

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ SI, BX
	MULXQ q<>+32(SB), AX, SI
	ADOXQ AX, BX

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ DI, SI
	MULXQ q<>+40(SB), AX, DI
	ADOXQ AX, SI

	// (C,t[5]) := t[6] + m*q[6] + C
	ADCXQ R8, DI
	MULXQ q<>+48(SB), AX, R8
	ADOXQ AX, DI
	MOVQ  $0, AX
	ADCXQ AX, R8
	ADOXQ AX, R8
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

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ CX, R15
	MULXQ q<>+16(SB), AX, CX
	ADOXQ AX, R15

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BX, CX
	MULXQ q<>+24(SB), AX, BX
	ADOXQ AX, CX

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ SI, BX
	MULXQ q<>+32(SB), AX, SI
	ADOXQ AX, BX

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ DI, SI
	MULXQ q<>+40(SB), AX, DI
	ADOXQ AX, SI

	// (C,t[5]) := t[6] + m*q[6] + C
	ADCXQ R8, DI
	MULXQ q<>+48(SB), AX, R8
	ADOXQ AX, DI
	MOVQ  $0, AX
	ADCXQ AX, R8
	ADOXQ AX, R8
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

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ CX, R15
	MULXQ q<>+16(SB), AX, CX
	ADOXQ AX, R15

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BX, CX
	MULXQ q<>+24(SB), AX, BX
	ADOXQ AX, CX

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ SI, BX
	MULXQ q<>+32(SB), AX, SI
	ADOXQ AX, BX

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ DI, SI
	MULXQ q<>+40(SB), AX, DI
	ADOXQ AX, SI

	// (C,t[5]) := t[6] + m*q[6] + C
	ADCXQ R8, DI
	MULXQ q<>+48(SB), AX, R8
	ADOXQ AX, DI
	MOVQ  $0, AX
	ADCXQ AX, R8
	ADOXQ AX, R8
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

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ CX, R15
	MULXQ q<>+16(SB), AX, CX
	ADOXQ AX, R15

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BX, CX
	MULXQ q<>+24(SB), AX, BX
	ADOXQ AX, CX

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ SI, BX
	MULXQ q<>+32(SB), AX, SI
	ADOXQ AX, BX

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ DI, SI
	MULXQ q<>+40(SB), AX, DI
	ADOXQ AX, SI

	// (C,t[5]) := t[6] + m*q[6] + C
	ADCXQ R8, DI
	MULXQ q<>+48(SB), AX, R8
	ADOXQ AX, DI
	MOVQ  $0, AX
	ADCXQ AX, R8
	ADOXQ AX, R8
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

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ CX, R15
	MULXQ q<>+16(SB), AX, CX
	ADOXQ AX, R15

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BX, CX
	MULXQ q<>+24(SB), AX, BX
	ADOXQ AX, CX

	// (C,t[3]) := t[4] + m*q[4] + C
	ADCXQ SI, BX
	MULXQ q<>+32(SB), AX, SI
	ADOXQ AX, BX

	// (C,t[4]) := t[5] + m*q[5] + C
	ADCXQ DI, SI
	MULXQ q<>+40(SB), AX, DI
	ADOXQ AX, SI

	// (C,t[5]) := t[6] + m*q[6] + C
	ADCXQ R8, DI
	MULXQ q<>+48(SB), AX, R8
	ADOXQ AX, DI
	MOVQ  $0, AX
	ADCXQ AX, R8
	ADOXQ AX, R8

	// reduce element(R14,R15,CX,BX,SI,DI,R8) using temp registers (R9,R10,R11,R12,R13,s0-8(SP),s1-16(SP))
	REDUCE(R14,R15,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,s0-8(SP),s1-16(SP))

	MOVQ res+0(FP), AX
	MOVQ R14, 0(AX)
	MOVQ R15, 8(AX)
	MOVQ CX, 16(AX)
	MOVQ BX, 24(AX)
	MOVQ SI, 32(AX)
	MOVQ DI, 40(AX)
	MOVQ R8, 48(AX)
	RET
