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
DATA q<>+0(SB)/8, $0x0a11800000000001
DATA q<>+8(SB)/8, $0x59aa76fed0000001
DATA q<>+16(SB)/8, $0x60b44d1e5c37b001
DATA q<>+24(SB)/8, $0x12ab655e9a2ca556
GLOBL q<>(SB), (RODATA+NOPTR), $32

// qInv0 q'[0]
DATA qInv0<>(SB)/8, $0x0a117fffffffffff
GLOBL qInv0<>(SB), (RODATA+NOPTR), $8
// Mu
DATA mu<>(SB)/8, $0x0000000db65247b1
GLOBL mu<>(SB), (RODATA+NOPTR), $8

#define REDUCE(ra0, ra1, ra2, ra3, rb0, rb1, rb2, rb3) \
	MOVQ    ra0, rb0;        \
	SUBQ    q<>(SB), ra0;    \
	MOVQ    ra1, rb1;        \
	SBBQ    q<>+8(SB), ra1;  \
	MOVQ    ra2, rb2;        \
	SBBQ    q<>+16(SB), ra2; \
	MOVQ    ra3, rb3;        \
	SBBQ    q<>+24(SB), ra3; \
	CMOVQCS rb0, ra0;        \
	CMOVQCS rb1, ra1;        \
	CMOVQCS rb2, ra2;        \
	CMOVQCS rb3, ra3;        \

TEXT ·reduce(SB), NOSPLIT, $0-8
	MOVQ res+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), SI

	// reduce element(DX,CX,BX,SI) using temp registers (DI,R8,R9,R10)
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10)

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	MOVQ SI, 24(AX)
	RET

// MulBy3(x *Element)
TEXT ·MulBy3(SB), NOSPLIT, $0-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), SI
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI

	// reduce element(DX,CX,BX,SI) using temp registers (DI,R8,R9,R10)
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10)

	ADDQ 0(AX), DX
	ADCQ 8(AX), CX
	ADCQ 16(AX), BX
	ADCQ 24(AX), SI

	// reduce element(DX,CX,BX,SI) using temp registers (R11,R12,R13,R14)
	REDUCE(DX,CX,BX,SI,R11,R12,R13,R14)

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	MOVQ SI, 24(AX)
	RET

// MulBy5(x *Element)
TEXT ·MulBy5(SB), NOSPLIT, $0-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), SI
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI

	// reduce element(DX,CX,BX,SI) using temp registers (DI,R8,R9,R10)
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10)

	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI

	// reduce element(DX,CX,BX,SI) using temp registers (R11,R12,R13,R14)
	REDUCE(DX,CX,BX,SI,R11,R12,R13,R14)

	ADDQ 0(AX), DX
	ADCQ 8(AX), CX
	ADCQ 16(AX), BX
	ADCQ 24(AX), SI

	// reduce element(DX,CX,BX,SI) using temp registers (R15,DI,R8,R9)
	REDUCE(DX,CX,BX,SI,R15,DI,R8,R9)

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	MOVQ SI, 24(AX)
	RET

// MulBy13(x *Element)
TEXT ·MulBy13(SB), NOSPLIT, $0-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), SI
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI

	// reduce element(DX,CX,BX,SI) using temp registers (DI,R8,R9,R10)
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10)

	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI

	// reduce element(DX,CX,BX,SI) using temp registers (R11,R12,R13,R14)
	REDUCE(DX,CX,BX,SI,R11,R12,R13,R14)

	MOVQ DX, R11
	MOVQ CX, R12
	MOVQ BX, R13
	MOVQ SI, R14
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI

	// reduce element(DX,CX,BX,SI) using temp registers (DI,R8,R9,R10)
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10)

	ADDQ R11, DX
	ADCQ R12, CX
	ADCQ R13, BX
	ADCQ R14, SI

	// reduce element(DX,CX,BX,SI) using temp registers (DI,R8,R9,R10)
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10)

	ADDQ 0(AX), DX
	ADCQ 8(AX), CX
	ADCQ 16(AX), BX
	ADCQ 24(AX), SI

	// reduce element(DX,CX,BX,SI) using temp registers (DI,R8,R9,R10)
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10)

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	MOVQ SI, 24(AX)
	RET

// Butterfly(a, b *Element) sets a = a + b; b = a - b
TEXT ·Butterfly(SB), NOSPLIT, $0-16
	MOVQ    a+0(FP), AX
	MOVQ    0(AX), CX
	MOVQ    8(AX), BX
	MOVQ    16(AX), SI
	MOVQ    24(AX), DI
	MOVQ    CX, R8
	MOVQ    BX, R9
	MOVQ    SI, R10
	MOVQ    DI, R11
	XORQ    AX, AX
	MOVQ    b+8(FP), DX
	ADDQ    0(DX), CX
	ADCQ    8(DX), BX
	ADCQ    16(DX), SI
	ADCQ    24(DX), DI
	SUBQ    0(DX), R8
	SBBQ    8(DX), R9
	SBBQ    16(DX), R10
	SBBQ    24(DX), R11
	MOVQ    $0x0a11800000000001, R12
	MOVQ    $0x59aa76fed0000001, R13
	MOVQ    $0x60b44d1e5c37b001, R14
	MOVQ    $0x12ab655e9a2ca556, R15
	CMOVQCC AX, R12
	CMOVQCC AX, R13
	CMOVQCC AX, R14
	CMOVQCC AX, R15
	ADDQ    R12, R8
	ADCQ    R13, R9
	ADCQ    R14, R10
	ADCQ    R15, R11
	MOVQ    R8, 0(DX)
	MOVQ    R9, 8(DX)
	MOVQ    R10, 16(DX)
	MOVQ    R11, 24(DX)

	// reduce element(CX,BX,SI,DI) using temp registers (R8,R9,R10,R11)
	REDUCE(CX,BX,SI,DI,R8,R9,R10,R11)

	MOVQ a+0(FP), AX
	MOVQ CX, 0(AX)
	MOVQ BX, 8(AX)
	MOVQ SI, 16(AX)
	MOVQ DI, 24(AX)
	RET

// addVec(res, a, b *Element, n uint64) res[0...n] = a[0...n] + b[0...n]
TEXT ·addVec(SB), NOSPLIT, $0-32
	MOVQ res+0(FP), CX
	MOVQ a+8(FP), AX
	MOVQ b+16(FP), DX
	MOVQ n+24(FP), BX

loop_1:
	TESTQ BX, BX
	JEQ   done_2 // n == 0, we are done

	// a[0] -> SI
	// a[1] -> DI
	// a[2] -> R8
	// a[3] -> R9
	MOVQ 0(AX), SI
	MOVQ 8(AX), DI
	MOVQ 16(AX), R8
	MOVQ 24(AX), R9
	ADDQ 0(DX), SI
	ADCQ 8(DX), DI
	ADCQ 16(DX), R8
	ADCQ 24(DX), R9

	// reduce element(SI,DI,R8,R9) using temp registers (R10,R11,R12,R13)
	REDUCE(SI,DI,R8,R9,R10,R11,R12,R13)

	MOVQ SI, 0(CX)
	MOVQ DI, 8(CX)
	MOVQ R8, 16(CX)
	MOVQ R9, 24(CX)

	// increment pointers to visit next element
	ADDQ $32, AX
	ADDQ $32, DX
	ADDQ $32, CX
	DECQ BX      // decrement n
	JMP  loop_1

done_2:
	RET

// subVec(res, a, b *Element, n uint64) res[0...n] = a[0...n] - b[0...n]
TEXT ·subVec(SB), NOSPLIT, $0-32
	MOVQ res+0(FP), CX
	MOVQ a+8(FP), AX
	MOVQ b+16(FP), DX
	MOVQ n+24(FP), BX
	XORQ SI, SI

loop_3:
	TESTQ BX, BX
	JEQ   done_4 // n == 0, we are done

	// a[0] -> DI
	// a[1] -> R8
	// a[2] -> R9
	// a[3] -> R10
	MOVQ 0(AX), DI
	MOVQ 8(AX), R8
	MOVQ 16(AX), R9
	MOVQ 24(AX), R10
	SUBQ 0(DX), DI
	SBBQ 8(DX), R8
	SBBQ 16(DX), R9
	SBBQ 24(DX), R10

	// reduce (a-b) mod q
	// q[0] -> R11
	// q[1] -> R12
	// q[2] -> R13
	// q[3] -> R14
	MOVQ    $0x0a11800000000001, R11
	MOVQ    $0x59aa76fed0000001, R12
	MOVQ    $0x60b44d1e5c37b001, R13
	MOVQ    $0x12ab655e9a2ca556, R14
	CMOVQCC SI, R11
	CMOVQCC SI, R12
	CMOVQCC SI, R13
	CMOVQCC SI, R14

	// add registers (q or 0) to a, and set to result
	ADDQ R11, DI
	ADCQ R12, R8
	ADCQ R13, R9
	ADCQ R14, R10
	MOVQ DI, 0(CX)
	MOVQ R8, 8(CX)
	MOVQ R9, 16(CX)
	MOVQ R10, 24(CX)

	// increment pointers to visit next element
	ADDQ $32, AX
	ADDQ $32, DX
	ADDQ $32, CX
	DECQ BX      // decrement n
	JMP  loop_3

done_4:
	RET

// scalarMulVec(res, a, b *Element, n uint64) res[0...n] = a[0...n] * b
TEXT ·scalarMulVec(SB), $56-32
	CMPB ·supportAdx(SB), $1
	JNE  noAdx_5
	MOVQ a+8(FP), R11
	MOVQ b+16(FP), R10
	MOVQ n+24(FP), R12

	// scalar[0] -> SI
	// scalar[1] -> DI
	// scalar[2] -> R8
	// scalar[3] -> R9
	MOVQ 0(R10), SI
	MOVQ 8(R10), DI
	MOVQ 16(R10), R8
	MOVQ 24(R10), R9
	MOVQ res+0(FP), R10

loop_6:
	TESTQ R12, R12
	JEQ   done_7   // n == 0, we are done

	// TODO @gbotrel this is generated from the same macro as the unit mul, we should refactor this in a single asm function
	// A -> BP
	// t[0] -> R14
	// t[1] -> R15
	// t[2] -> CX
	// t[3] -> BX
	// clear the flags
	XORQ AX, AX
	MOVQ 0(R11), DX

	// (A,t[0])  := x[0]*y[0] + A
	MULXQ SI, R14, R15

	// (A,t[1])  := x[1]*y[0] + A
	MULXQ DI, AX, CX
	ADOXQ AX, R15

	// (A,t[2])  := x[2]*y[0] + A
	MULXQ R8, AX, BX
	ADOXQ AX, CX

	// (A,t[3])  := x[3]*y[0] + A
	MULXQ R9, AX, BP
	ADOXQ AX, BX

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADOXQ AX, BP

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R14, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, R13
	ADCXQ R14, AX
	MOVQ  R13, R14

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

	// t[3] = C + A
	MOVQ  $0, AX
	ADCXQ AX, BX
	ADOXQ BP, BX

	// clear the flags
	XORQ AX, AX
	MOVQ 8(R11), DX

	// (A,t[0])  := t[0] + x[0]*y[1] + A
	MULXQ SI, AX, BP
	ADOXQ AX, R14

	// (A,t[1])  := t[1] + x[1]*y[1] + A
	ADCXQ BP, R15
	MULXQ DI, AX, BP
	ADOXQ AX, R15

	// (A,t[2])  := t[2] + x[2]*y[1] + A
	ADCXQ BP, CX
	MULXQ R8, AX, BP
	ADOXQ AX, CX

	// (A,t[3])  := t[3] + x[3]*y[1] + A
	ADCXQ BP, BX
	MULXQ R9, AX, BP
	ADOXQ AX, BX

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
	MULXQ q<>+0(SB), AX, R13
	ADCXQ R14, AX
	MOVQ  R13, R14

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

	// t[3] = C + A
	MOVQ  $0, AX
	ADCXQ AX, BX
	ADOXQ BP, BX

	// clear the flags
	XORQ AX, AX
	MOVQ 16(R11), DX

	// (A,t[0])  := t[0] + x[0]*y[2] + A
	MULXQ SI, AX, BP
	ADOXQ AX, R14

	// (A,t[1])  := t[1] + x[1]*y[2] + A
	ADCXQ BP, R15
	MULXQ DI, AX, BP
	ADOXQ AX, R15

	// (A,t[2])  := t[2] + x[2]*y[2] + A
	ADCXQ BP, CX
	MULXQ R8, AX, BP
	ADOXQ AX, CX

	// (A,t[3])  := t[3] + x[3]*y[2] + A
	ADCXQ BP, BX
	MULXQ R9, AX, BP
	ADOXQ AX, BX

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
	MULXQ q<>+0(SB), AX, R13
	ADCXQ R14, AX
	MOVQ  R13, R14

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

	// t[3] = C + A
	MOVQ  $0, AX
	ADCXQ AX, BX
	ADOXQ BP, BX

	// clear the flags
	XORQ AX, AX
	MOVQ 24(R11), DX

	// (A,t[0])  := t[0] + x[0]*y[3] + A
	MULXQ SI, AX, BP
	ADOXQ AX, R14

	// (A,t[1])  := t[1] + x[1]*y[3] + A
	ADCXQ BP, R15
	MULXQ DI, AX, BP
	ADOXQ AX, R15

	// (A,t[2])  := t[2] + x[2]*y[3] + A
	ADCXQ BP, CX
	MULXQ R8, AX, BP
	ADOXQ AX, CX

	// (A,t[3])  := t[3] + x[3]*y[3] + A
	ADCXQ BP, BX
	MULXQ R9, AX, BP
	ADOXQ AX, BX

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
	MULXQ q<>+0(SB), AX, R13
	ADCXQ R14, AX
	MOVQ  R13, R14

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

	// t[3] = C + A
	MOVQ  $0, AX
	ADCXQ AX, BX
	ADOXQ BP, BX

	// reduce t mod q
	// reduce element(R14,R15,CX,BX) using temp registers (R13,AX,DX,s0-8(SP))
	REDUCE(R14,R15,CX,BX,R13,AX,DX,s0-8(SP))

	MOVQ R14, 0(R10)
	MOVQ R15, 8(R10)
	MOVQ CX, 16(R10)
	MOVQ BX, 24(R10)

	// increment pointers to visit next element
	ADDQ $32, R11
	ADDQ $32, R10
	DECQ R12      // decrement n
	JMP  loop_6

done_7:
	RET

noAdx_5:
	MOVQ n+24(FP), DX
	MOVQ res+0(FP), AX
	MOVQ AX, (SP)
	MOVQ DX, 8(SP)
	MOVQ DX, 16(SP)
	MOVQ a+8(FP), AX
	MOVQ AX, 24(SP)
	MOVQ DX, 32(SP)
	MOVQ DX, 40(SP)
	MOVQ b+16(FP), AX
	MOVQ AX, 48(SP)
	CALL ·scalarMulVecGeneric(SB)
	RET

// sumVec(res, a *Element, n uint64) res = sum(a[0...n])
TEXT ·sumVec(SB), NOSPLIT, $0-24
	XORQ      AX, AX
	MOVQ      a+8(FP), R14
	MOVQ      n+16(FP), R15
	VXORPS    Z0, Z0, Z0
	VMOVDQA64 Z0, Z1
	VMOVDQA64 Z0, Z2
	VMOVDQA64 Z0, Z3
	TESTQ     R15, R15
	JEQ       done_9        // n == 0, we are done
	MOVQ      R15, CX
	ANDQ      $3, CX
	SHRQ      $2, R15
	CMPB      CX, $1
	JEQ       rr1_10        // we have 1 remaining element
	CMPB      CX, $2
	JEQ       rr2_11        // we have 2 remaining elements
	CMPB      CX, $3
	JNE       loop_8        // == 0; we have 0 remaining elements

	// we have 3 remaining elements
	VPMOVZXDQ 2*32(R14), Z4
	VPADDQ    Z4, Z0, Z0

rr2_11:
	// we have 2 remaining elements
	VPMOVZXDQ 1*32(R14), Z4
	VPADDQ    Z4, Z1, Z1

rr1_10:
	// we have 1 remaining element
	VPMOVZXDQ 0*32(R14), Z4
	VPADDQ    Z4, Z2, Z2

loop_8:
	TESTQ     R15, R15
	JEQ       accumulate_12 // n == 0, we are going to accumulate
	VPMOVZXDQ 0*32(R14), Z4
	VPADDQ    Z4, Z0, Z0
	VPMOVZXDQ 1*32(R14), Z4
	VPADDQ    Z4, Z1, Z1
	VPMOVZXDQ 2*32(R14), Z4
	VPADDQ    Z4, Z2, Z2
	VPMOVZXDQ 3*32(R14), Z4
	VPADDQ    Z4, Z3, Z3

	// increment pointers to visit next 4 elements
	ADDQ $128, R14
	DECQ R15       // decrement n
	JMP  loop_8

accumulate_12:
	VPADDQ  Z1, Z0, Z0
	VPADDQ  Z3, Z2, Z2
	VPADDQ  Z2, Z0, Z0
	VMOVQ   X0, BX
	VALIGNQ $1, Z0, Z0, Z0
	VMOVQ   X0, SI
	VALIGNQ $1, Z0, Z0, Z0
	VMOVQ   X0, DI
	VALIGNQ $1, Z0, Z0, Z0
	VMOVQ   X0, R8
	VALIGNQ $1, Z0, Z0, Z0
	VMOVQ   X0, R9
	VALIGNQ $1, Z0, Z0, Z0
	VMOVQ   X0, R10
	VALIGNQ $1, Z0, Z0, Z0
	VMOVQ   X0, R11
	VALIGNQ $1, Z0, Z0, Z0
	VMOVQ   X0, R12
	XORQ    AX, AX
	MOVQ    SI, R13
	ANDQ    $0xffffffff, R13
	SHLQ    $32, R13
	SHRQ    $32, SI
	MOVQ    R8, CX
	ANDQ    $0xffffffff, CX
	SHLQ    $32, CX
	SHRQ    $32, R8
	MOVQ    R10, R15
	ANDQ    $0xffffffff, R15
	SHLQ    $32, R15
	SHRQ    $32, R10
	MOVQ    R12, R14
	ANDQ    $0xffffffff, R14
	SHLQ    $32, R14
	SHRQ    $32, R12
	XORQ    AX, AX
	ADOXQ   R13, BX
	ADOXQ   CX, DI
	ADCXQ   SI, DI
	ADOXQ   R15, R9
	ADCXQ   R8, R9
	ADOXQ   R14, R11
	ADCXQ   R10, R11
	ADOXQ   AX, R12
	ADCXQ   AX, R12
	XORQ    AX, AX
	MOVQ    mu<>(SB), SI
	MOVQ    R11, AX
	SHRQ    $32, R12, AX
	MULQ    SI
	MULXQ   q<>+0(SB), AX, SI
	SUBQ    AX, BX
	SBBQ    SI, DI
	MULXQ   q<>+16(SB), AX, SI
	SBBQ    AX, R9
	SBBQ    SI, R11
	SBBQ    $0, R12
	MULXQ   q<>+8(SB), AX, SI
	SUBQ    AX, DI
	SBBQ    SI, R9
	MULXQ   q<>+24(SB), AX, SI
	SBBQ    AX, R11
	SBBQ    SI, R12
	MOVQ    res+0(FP), SI
	MOVQ    BX, 0(SI)
	MOVQ    DI, 8(SI)
	MOVQ    R9, 16(SI)
	MOVQ    R11, 24(SI)
	SUBQ    q<>+0(SB), BX
	SBBQ    q<>+8(SB), DI
	SBBQ    q<>+16(SB), R9
	SBBQ    q<>+24(SB), R11
	SBBQ    $0, R12
	JCS     done_9
	MOVQ    BX, 0(SI)
	MOVQ    DI, 8(SI)
	MOVQ    R9, 16(SI)
	MOVQ    R11, 24(SI)
	SUBQ    q<>+0(SB), BX
	SBBQ    q<>+8(SB), DI
	SBBQ    q<>+16(SB), R9
	SBBQ    q<>+24(SB), R11
	SBBQ    $0, R12
	JCS     done_9
	MOVQ    BX, 0(SI)
	MOVQ    DI, 8(SI)
	MOVQ    R9, 16(SI)
	MOVQ    R11, 24(SI)

done_9:
	RET
