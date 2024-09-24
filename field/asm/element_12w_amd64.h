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

#define REDUCE(ra0, ra1, ra2, ra3, ra4, ra5, ra6, ra7, ra8, ra9, ra10, ra11, rb0, rb1, rb2, rb3, rb4, rb5, rb6, rb7, rb8, rb9, rb10, rb11) \
	MOVQ    ra0, rb0;         \
	SUBQ    q<>(SB), ra0;     \
	MOVQ    ra1, rb1;         \
	SBBQ    q<>+8(SB), ra1;   \
	MOVQ    ra2, rb2;         \
	SBBQ    q<>+16(SB), ra2;  \
	MOVQ    ra3, rb3;         \
	SBBQ    q<>+24(SB), ra3;  \
	MOVQ    ra4, rb4;         \
	SBBQ    q<>+32(SB), ra4;  \
	MOVQ    ra5, rb5;         \
	SBBQ    q<>+40(SB), ra5;  \
	MOVQ    ra6, rb6;         \
	SBBQ    q<>+48(SB), ra6;  \
	MOVQ    ra7, rb7;         \
	SBBQ    q<>+56(SB), ra7;  \
	MOVQ    ra8, rb8;         \
	SBBQ    q<>+64(SB), ra8;  \
	MOVQ    ra9, rb9;         \
	SBBQ    q<>+72(SB), ra9;  \
	MOVQ    ra10, rb10;       \
	SBBQ    q<>+80(SB), ra10; \
	MOVQ    ra11, rb11;       \
	SBBQ    q<>+88(SB), ra11; \
	CMOVQCS rb0, ra0;         \
	CMOVQCS rb1, ra1;         \
	CMOVQCS rb2, ra2;         \
	CMOVQCS rb3, ra3;         \
	CMOVQCS rb4, ra4;         \
	CMOVQCS rb5, ra5;         \
	CMOVQCS rb6, ra6;         \
	CMOVQCS rb7, ra7;         \
	CMOVQCS rb8, ra8;         \
	CMOVQCS rb9, ra9;         \
	CMOVQCS rb10, ra10;       \
	CMOVQCS rb11, ra11;       \

TEXT ·reduce(SB), $88-8
	MOVQ res+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), SI
	MOVQ 32(AX), DI
	MOVQ 40(AX), R8
	MOVQ 48(AX), R9
	MOVQ 56(AX), R10
	MOVQ 64(AX), R11
	MOVQ 72(AX), R12
	MOVQ 80(AX), R13
	MOVQ 88(AX), R14

	// reduce element(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14) using temp registers (R15,s0-8(SP),s1-16(SP),s2-24(SP),s3-32(SP),s4-40(SP),s5-48(SP),s6-56(SP),s7-64(SP),s8-72(SP),s9-80(SP),s10-88(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,s0-8(SP),s1-16(SP),s2-24(SP),s3-32(SP),s4-40(SP),s5-48(SP),s6-56(SP),s7-64(SP),s8-72(SP),s9-80(SP),s10-88(SP))

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	MOVQ SI, 24(AX)
	MOVQ DI, 32(AX)
	MOVQ R8, 40(AX)
	MOVQ R9, 48(AX)
	MOVQ R10, 56(AX)
	MOVQ R11, 64(AX)
	MOVQ R12, 72(AX)
	MOVQ R13, 80(AX)
	MOVQ R14, 88(AX)
	RET

// MulBy3(x *Element)
TEXT ·MulBy3(SB), $88-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), SI
	MOVQ 32(AX), DI
	MOVQ 40(AX), R8
	MOVQ 48(AX), R9
	MOVQ 56(AX), R10
	MOVQ 64(AX), R11
	MOVQ 72(AX), R12
	MOVQ 80(AX), R13
	MOVQ 88(AX), R14
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI
	ADCQ DI, DI
	ADCQ R8, R8
	ADCQ R9, R9
	ADCQ R10, R10
	ADCQ R11, R11
	ADCQ R12, R12
	ADCQ R13, R13
	ADCQ R14, R14

	// reduce element(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14) using temp registers (R15,s0-8(SP),s1-16(SP),s2-24(SP),s3-32(SP),s4-40(SP),s5-48(SP),s6-56(SP),s7-64(SP),s8-72(SP),s9-80(SP),s10-88(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,s0-8(SP),s1-16(SP),s2-24(SP),s3-32(SP),s4-40(SP),s5-48(SP),s6-56(SP),s7-64(SP),s8-72(SP),s9-80(SP),s10-88(SP))

	ADDQ 0(AX), DX
	ADCQ 8(AX), CX
	ADCQ 16(AX), BX
	ADCQ 24(AX), SI
	ADCQ 32(AX), DI
	ADCQ 40(AX), R8
	ADCQ 48(AX), R9
	ADCQ 56(AX), R10
	ADCQ 64(AX), R11
	ADCQ 72(AX), R12
	ADCQ 80(AX), R13
	ADCQ 88(AX), R14

	// reduce element(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14) using temp registers (R15,s0-8(SP),s1-16(SP),s2-24(SP),s3-32(SP),s4-40(SP),s5-48(SP),s6-56(SP),s7-64(SP),s8-72(SP),s9-80(SP),s10-88(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,s0-8(SP),s1-16(SP),s2-24(SP),s3-32(SP),s4-40(SP),s5-48(SP),s6-56(SP),s7-64(SP),s8-72(SP),s9-80(SP),s10-88(SP))

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	MOVQ SI, 24(AX)
	MOVQ DI, 32(AX)
	MOVQ R8, 40(AX)
	MOVQ R9, 48(AX)
	MOVQ R10, 56(AX)
	MOVQ R11, 64(AX)
	MOVQ R12, 72(AX)
	MOVQ R13, 80(AX)
	MOVQ R14, 88(AX)
	RET

// MulBy5(x *Element)
TEXT ·MulBy5(SB), $88-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), SI
	MOVQ 32(AX), DI
	MOVQ 40(AX), R8
	MOVQ 48(AX), R9
	MOVQ 56(AX), R10
	MOVQ 64(AX), R11
	MOVQ 72(AX), R12
	MOVQ 80(AX), R13
	MOVQ 88(AX), R14
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI
	ADCQ DI, DI
	ADCQ R8, R8
	ADCQ R9, R9
	ADCQ R10, R10
	ADCQ R11, R11
	ADCQ R12, R12
	ADCQ R13, R13
	ADCQ R14, R14

	// reduce element(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14) using temp registers (R15,s0-8(SP),s1-16(SP),s2-24(SP),s3-32(SP),s4-40(SP),s5-48(SP),s6-56(SP),s7-64(SP),s8-72(SP),s9-80(SP),s10-88(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,s0-8(SP),s1-16(SP),s2-24(SP),s3-32(SP),s4-40(SP),s5-48(SP),s6-56(SP),s7-64(SP),s8-72(SP),s9-80(SP),s10-88(SP))

	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI
	ADCQ DI, DI
	ADCQ R8, R8
	ADCQ R9, R9
	ADCQ R10, R10
	ADCQ R11, R11
	ADCQ R12, R12
	ADCQ R13, R13
	ADCQ R14, R14

	// reduce element(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14) using temp registers (R15,s0-8(SP),s1-16(SP),s2-24(SP),s3-32(SP),s4-40(SP),s5-48(SP),s6-56(SP),s7-64(SP),s8-72(SP),s9-80(SP),s10-88(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,s0-8(SP),s1-16(SP),s2-24(SP),s3-32(SP),s4-40(SP),s5-48(SP),s6-56(SP),s7-64(SP),s8-72(SP),s9-80(SP),s10-88(SP))

	ADDQ 0(AX), DX
	ADCQ 8(AX), CX
	ADCQ 16(AX), BX
	ADCQ 24(AX), SI
	ADCQ 32(AX), DI
	ADCQ 40(AX), R8
	ADCQ 48(AX), R9
	ADCQ 56(AX), R10
	ADCQ 64(AX), R11
	ADCQ 72(AX), R12
	ADCQ 80(AX), R13
	ADCQ 88(AX), R14

	// reduce element(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14) using temp registers (R15,s0-8(SP),s1-16(SP),s2-24(SP),s3-32(SP),s4-40(SP),s5-48(SP),s6-56(SP),s7-64(SP),s8-72(SP),s9-80(SP),s10-88(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,s0-8(SP),s1-16(SP),s2-24(SP),s3-32(SP),s4-40(SP),s5-48(SP),s6-56(SP),s7-64(SP),s8-72(SP),s9-80(SP),s10-88(SP))

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	MOVQ SI, 24(AX)
	MOVQ DI, 32(AX)
	MOVQ R8, 40(AX)
	MOVQ R9, 48(AX)
	MOVQ R10, 56(AX)
	MOVQ R11, 64(AX)
	MOVQ R12, 72(AX)
	MOVQ R13, 80(AX)
	MOVQ R14, 88(AX)
	RET

// MulBy13(x *Element)
TEXT ·MulBy13(SB), $184-8
	MOVQ x+0(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), SI
	MOVQ 32(AX), DI
	MOVQ 40(AX), R8
	MOVQ 48(AX), R9
	MOVQ 56(AX), R10
	MOVQ 64(AX), R11
	MOVQ 72(AX), R12
	MOVQ 80(AX), R13
	MOVQ 88(AX), R14
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI
	ADCQ DI, DI
	ADCQ R8, R8
	ADCQ R9, R9
	ADCQ R10, R10
	ADCQ R11, R11
	ADCQ R12, R12
	ADCQ R13, R13
	ADCQ R14, R14

	// reduce element(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14) using temp registers (R15,s0-8(SP),s1-16(SP),s2-24(SP),s3-32(SP),s4-40(SP),s5-48(SP),s6-56(SP),s7-64(SP),s8-72(SP),s9-80(SP),s10-88(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,s0-8(SP),s1-16(SP),s2-24(SP),s3-32(SP),s4-40(SP),s5-48(SP),s6-56(SP),s7-64(SP),s8-72(SP),s9-80(SP),s10-88(SP))

	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI
	ADCQ DI, DI
	ADCQ R8, R8
	ADCQ R9, R9
	ADCQ R10, R10
	ADCQ R11, R11
	ADCQ R12, R12
	ADCQ R13, R13
	ADCQ R14, R14

	// reduce element(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14) using temp registers (s11-96(SP),s12-104(SP),s13-112(SP),s14-120(SP),s15-128(SP),s16-136(SP),s17-144(SP),s18-152(SP),s19-160(SP),s20-168(SP),s21-176(SP),s22-184(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,s11-96(SP),s12-104(SP),s13-112(SP),s14-120(SP),s15-128(SP),s16-136(SP),s17-144(SP),s18-152(SP),s19-160(SP),s20-168(SP),s21-176(SP),s22-184(SP))

	MOVQ DX, s11-96(SP)
	MOVQ CX, s12-104(SP)
	MOVQ BX, s13-112(SP)
	MOVQ SI, s14-120(SP)
	MOVQ DI, s15-128(SP)
	MOVQ R8, s16-136(SP)
	MOVQ R9, s17-144(SP)
	MOVQ R10, s18-152(SP)
	MOVQ R11, s19-160(SP)
	MOVQ R12, s20-168(SP)
	MOVQ R13, s21-176(SP)
	MOVQ R14, s22-184(SP)
	ADDQ DX, DX
	ADCQ CX, CX
	ADCQ BX, BX
	ADCQ SI, SI
	ADCQ DI, DI
	ADCQ R8, R8
	ADCQ R9, R9
	ADCQ R10, R10
	ADCQ R11, R11
	ADCQ R12, R12
	ADCQ R13, R13
	ADCQ R14, R14

	// reduce element(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14) using temp registers (R15,s0-8(SP),s1-16(SP),s2-24(SP),s3-32(SP),s4-40(SP),s5-48(SP),s6-56(SP),s7-64(SP),s8-72(SP),s9-80(SP),s10-88(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,s0-8(SP),s1-16(SP),s2-24(SP),s3-32(SP),s4-40(SP),s5-48(SP),s6-56(SP),s7-64(SP),s8-72(SP),s9-80(SP),s10-88(SP))

	ADDQ s11-96(SP), DX
	ADCQ s12-104(SP), CX
	ADCQ s13-112(SP), BX
	ADCQ s14-120(SP), SI
	ADCQ s15-128(SP), DI
	ADCQ s16-136(SP), R8
	ADCQ s17-144(SP), R9
	ADCQ s18-152(SP), R10
	ADCQ s19-160(SP), R11
	ADCQ s20-168(SP), R12
	ADCQ s21-176(SP), R13
	ADCQ s22-184(SP), R14

	// reduce element(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14) using temp registers (R15,s0-8(SP),s1-16(SP),s2-24(SP),s3-32(SP),s4-40(SP),s5-48(SP),s6-56(SP),s7-64(SP),s8-72(SP),s9-80(SP),s10-88(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,s0-8(SP),s1-16(SP),s2-24(SP),s3-32(SP),s4-40(SP),s5-48(SP),s6-56(SP),s7-64(SP),s8-72(SP),s9-80(SP),s10-88(SP))

	ADDQ 0(AX), DX
	ADCQ 8(AX), CX
	ADCQ 16(AX), BX
	ADCQ 24(AX), SI
	ADCQ 32(AX), DI
	ADCQ 40(AX), R8
	ADCQ 48(AX), R9
	ADCQ 56(AX), R10
	ADCQ 64(AX), R11
	ADCQ 72(AX), R12
	ADCQ 80(AX), R13
	ADCQ 88(AX), R14

	// reduce element(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14) using temp registers (R15,s0-8(SP),s1-16(SP),s2-24(SP),s3-32(SP),s4-40(SP),s5-48(SP),s6-56(SP),s7-64(SP),s8-72(SP),s9-80(SP),s10-88(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,s0-8(SP),s1-16(SP),s2-24(SP),s3-32(SP),s4-40(SP),s5-48(SP),s6-56(SP),s7-64(SP),s8-72(SP),s9-80(SP),s10-88(SP))

	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	MOVQ SI, 24(AX)
	MOVQ DI, 32(AX)
	MOVQ R8, 40(AX)
	MOVQ R9, 48(AX)
	MOVQ R10, 56(AX)
	MOVQ R11, 64(AX)
	MOVQ R12, 72(AX)
	MOVQ R13, 80(AX)
	MOVQ R14, 88(AX)
	RET

// Butterfly(a, b *Element) sets a = a + b; b = a - b
TEXT ·Butterfly(SB), $88-16
	MOVQ b+8(FP), AX
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), SI
	MOVQ 32(AX), DI
	MOVQ 40(AX), R8
	MOVQ 48(AX), R9
	MOVQ 56(AX), R10
	MOVQ 64(AX), R11
	MOVQ 72(AX), R12
	MOVQ 80(AX), R13
	MOVQ 88(AX), R14
	MOVQ a+0(FP), AX
	ADDQ 0(AX), DX
	ADCQ 8(AX), CX
	ADCQ 16(AX), BX
	ADCQ 24(AX), SI
	ADCQ 32(AX), DI
	ADCQ 40(AX), R8
	ADCQ 48(AX), R9
	ADCQ 56(AX), R10
	ADCQ 64(AX), R11
	ADCQ 72(AX), R12
	ADCQ 80(AX), R13
	ADCQ 88(AX), R14
	MOVQ DX, R15
	MOVQ CX, s0-8(SP)
	MOVQ BX, s1-16(SP)
	MOVQ SI, s2-24(SP)
	MOVQ DI, s3-32(SP)
	MOVQ R8, s4-40(SP)
	MOVQ R9, s5-48(SP)
	MOVQ R10, s6-56(SP)
	MOVQ R11, s7-64(SP)
	MOVQ R12, s8-72(SP)
	MOVQ R13, s9-80(SP)
	MOVQ R14, s10-88(SP)
	MOVQ 0(AX), DX
	MOVQ 8(AX), CX
	MOVQ 16(AX), BX
	MOVQ 24(AX), SI
	MOVQ 32(AX), DI
	MOVQ 40(AX), R8
	MOVQ 48(AX), R9
	MOVQ 56(AX), R10
	MOVQ 64(AX), R11
	MOVQ 72(AX), R12
	MOVQ 80(AX), R13
	MOVQ 88(AX), R14
	MOVQ b+8(FP), AX
	SUBQ 0(AX), DX
	SBBQ 8(AX), CX
	SBBQ 16(AX), BX
	SBBQ 24(AX), SI
	SBBQ 32(AX), DI
	SBBQ 40(AX), R8
	SBBQ 48(AX), R9
	SBBQ 56(AX), R10
	SBBQ 64(AX), R11
	SBBQ 72(AX), R12
	SBBQ 80(AX), R13
	SBBQ 88(AX), R14
	JCC  noReduce_1
	MOVQ q0, AX
	ADDQ AX, DX
	MOVQ q1, AX
	ADCQ AX, CX
	MOVQ q2, AX
	ADCQ AX, BX
	MOVQ q3, AX
	ADCQ AX, SI
	MOVQ q4, AX
	ADCQ AX, DI
	MOVQ q5, AX
	ADCQ AX, R8
	MOVQ q6, AX
	ADCQ AX, R9
	MOVQ q7, AX
	ADCQ AX, R10
	MOVQ q8, AX
	ADCQ AX, R11
	MOVQ q9, AX
	ADCQ AX, R12
	MOVQ q10, AX
	ADCQ AX, R13
	MOVQ q11, AX
	ADCQ AX, R14

noReduce_1:
	MOVQ b+8(FP), AX
	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	MOVQ SI, 24(AX)
	MOVQ DI, 32(AX)
	MOVQ R8, 40(AX)
	MOVQ R9, 48(AX)
	MOVQ R10, 56(AX)
	MOVQ R11, 64(AX)
	MOVQ R12, 72(AX)
	MOVQ R13, 80(AX)
	MOVQ R14, 88(AX)
	MOVQ R15, DX
	MOVQ s0-8(SP), CX
	MOVQ s1-16(SP), BX
	MOVQ s2-24(SP), SI
	MOVQ s3-32(SP), DI
	MOVQ s4-40(SP), R8
	MOVQ s5-48(SP), R9
	MOVQ s6-56(SP), R10
	MOVQ s7-64(SP), R11
	MOVQ s8-72(SP), R12
	MOVQ s9-80(SP), R13
	MOVQ s10-88(SP), R14

	// reduce element(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14) using temp registers (R15,s0-8(SP),s1-16(SP),s2-24(SP),s3-32(SP),s4-40(SP),s5-48(SP),s6-56(SP),s7-64(SP),s8-72(SP),s9-80(SP),s10-88(SP))
	REDUCE(DX,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,R14,R15,s0-8(SP),s1-16(SP),s2-24(SP),s3-32(SP),s4-40(SP),s5-48(SP),s6-56(SP),s7-64(SP),s8-72(SP),s9-80(SP),s10-88(SP))

	MOVQ a+0(FP), AX
	MOVQ DX, 0(AX)
	MOVQ CX, 8(AX)
	MOVQ BX, 16(AX)
	MOVQ SI, 24(AX)
	MOVQ DI, 32(AX)
	MOVQ R8, 40(AX)
	MOVQ R9, 48(AX)
	MOVQ R10, 56(AX)
	MOVQ R11, 64(AX)
	MOVQ R12, 72(AX)
	MOVQ R13, 80(AX)
	MOVQ R14, 88(AX)
	RET

// mul(res, x, y *Element)
TEXT ·mul(SB), $96-24

	// Algorithm 2 of "Faster Montgomery Multiplication and Multi-Scalar-Multiplication for SNARKS"
	// by Y. El Housni and G. Botrel https://doi.org/10.46586/tches.v2023.i3.504-521

	NO_LOCAL_POINTERS
	CMPB ·supportAdx(SB), $1
	JNE  noAdx_2
	MOVQ x+8(FP), AX

	// x[0] -> s0-8(SP)
	// x[1] -> s1-16(SP)
	// x[2] -> s2-24(SP)
	// x[3] -> s3-32(SP)
	// x[4] -> s4-40(SP)
	// x[5] -> s5-48(SP)
	// x[6] -> s6-56(SP)
	// x[7] -> s7-64(SP)
	// x[8] -> s8-72(SP)
	// x[9] -> s9-80(SP)
	// x[10] -> s10-88(SP)
	// x[11] -> s11-96(SP)
	MOVQ 0(AX), R14
	MOVQ 8(AX), R15
	MOVQ 16(AX), CX
	MOVQ 24(AX), BX
	MOVQ 32(AX), SI
	MOVQ 40(AX), DI
	MOVQ 48(AX), R8
	MOVQ 56(AX), R9
	MOVQ 64(AX), R10
	MOVQ 72(AX), R11
	MOVQ 80(AX), R12
	MOVQ 88(AX), R13
	MOVQ R14, s0-8(SP)
	MOVQ R15, s1-16(SP)
	MOVQ CX, s2-24(SP)
	MOVQ BX, s3-32(SP)
	MOVQ SI, s4-40(SP)
	MOVQ DI, s5-48(SP)
	MOVQ R8, s6-56(SP)
	MOVQ R9, s7-64(SP)
	MOVQ R10, s8-72(SP)
	MOVQ R11, s9-80(SP)
	MOVQ R12, s10-88(SP)
	MOVQ R13, s11-96(SP)

	// A -> BP
	// t[0] -> R14
	// t[1] -> R15
	// t[2] -> CX
	// t[3] -> BX
	// t[4] -> SI
	// t[5] -> DI
	// t[6] -> R8
	// t[7] -> R9
	// t[8] -> R10
	// t[9] -> R11
	// t[10] -> R12
	// t[11] -> R13
	// clear the flags
	XORQ AX, AX
	MOVQ y+16(FP), AX
	MOVQ 0(AX), DX

	// (A,t[0])  := x[0]*y[0] + A
	MULXQ s0-8(SP), R14, R15

	// (A,t[1])  := x[1]*y[0] + A
	MULXQ s1-16(SP), AX, CX
	ADOXQ AX, R15

	// (A,t[2])  := x[2]*y[0] + A
	MULXQ s2-24(SP), AX, BX
	ADOXQ AX, CX

	// (A,t[3])  := x[3]*y[0] + A
	MULXQ s3-32(SP), AX, SI
	ADOXQ AX, BX

	// (A,t[4])  := x[4]*y[0] + A
	MULXQ s4-40(SP), AX, DI
	ADOXQ AX, SI

	// (A,t[5])  := x[5]*y[0] + A
	MULXQ s5-48(SP), AX, R8
	ADOXQ AX, DI

	// (A,t[6])  := x[6]*y[0] + A
	MULXQ s6-56(SP), AX, R9
	ADOXQ AX, R8

	// (A,t[7])  := x[7]*y[0] + A
	MULXQ s7-64(SP), AX, R10
	ADOXQ AX, R9

	// (A,t[8])  := x[8]*y[0] + A
	MULXQ s8-72(SP), AX, R11
	ADOXQ AX, R10

	// (A,t[9])  := x[9]*y[0] + A
	MULXQ s9-80(SP), AX, R12
	ADOXQ AX, R11

	// (A,t[10])  := x[10]*y[0] + A
	MULXQ s10-88(SP), AX, R13
	ADOXQ AX, R12

	// (A,t[11])  := x[11]*y[0] + A
	MULXQ s11-96(SP), AX, BP
	ADOXQ AX, R13

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADOXQ AX, BP
	PUSHQ BP

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R14, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, BP
	ADCXQ R14, AX
	MOVQ  BP, R14
	POPQ  BP

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

	// (C,t[6]) := t[7] + m*q[7] + C
	ADCXQ R9, R8
	MULXQ q<>+56(SB), AX, R9
	ADOXQ AX, R8

	// (C,t[7]) := t[8] + m*q[8] + C
	ADCXQ R10, R9
	MULXQ q<>+64(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[8]) := t[9] + m*q[9] + C
	ADCXQ R11, R10
	MULXQ q<>+72(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[9]) := t[10] + m*q[10] + C
	ADCXQ R12, R11
	MULXQ q<>+80(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[10]) := t[11] + m*q[11] + C
	ADCXQ R13, R12
	MULXQ q<>+88(SB), AX, R13
	ADOXQ AX, R12

	// t[11] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R13
	ADOXQ BP, R13

	// clear the flags
	XORQ AX, AX
	MOVQ y+16(FP), AX
	MOVQ 8(AX), DX

	// (A,t[0])  := t[0] + x[0]*y[1] + A
	MULXQ s0-8(SP), AX, BP
	ADOXQ AX, R14

	// (A,t[1])  := t[1] + x[1]*y[1] + A
	ADCXQ BP, R15
	MULXQ s1-16(SP), AX, BP
	ADOXQ AX, R15

	// (A,t[2])  := t[2] + x[2]*y[1] + A
	ADCXQ BP, CX
	MULXQ s2-24(SP), AX, BP
	ADOXQ AX, CX

	// (A,t[3])  := t[3] + x[3]*y[1] + A
	ADCXQ BP, BX
	MULXQ s3-32(SP), AX, BP
	ADOXQ AX, BX

	// (A,t[4])  := t[4] + x[4]*y[1] + A
	ADCXQ BP, SI
	MULXQ s4-40(SP), AX, BP
	ADOXQ AX, SI

	// (A,t[5])  := t[5] + x[5]*y[1] + A
	ADCXQ BP, DI
	MULXQ s5-48(SP), AX, BP
	ADOXQ AX, DI

	// (A,t[6])  := t[6] + x[6]*y[1] + A
	ADCXQ BP, R8
	MULXQ s6-56(SP), AX, BP
	ADOXQ AX, R8

	// (A,t[7])  := t[7] + x[7]*y[1] + A
	ADCXQ BP, R9
	MULXQ s7-64(SP), AX, BP
	ADOXQ AX, R9

	// (A,t[8])  := t[8] + x[8]*y[1] + A
	ADCXQ BP, R10
	MULXQ s8-72(SP), AX, BP
	ADOXQ AX, R10

	// (A,t[9])  := t[9] + x[9]*y[1] + A
	ADCXQ BP, R11
	MULXQ s9-80(SP), AX, BP
	ADOXQ AX, R11

	// (A,t[10])  := t[10] + x[10]*y[1] + A
	ADCXQ BP, R12
	MULXQ s10-88(SP), AX, BP
	ADOXQ AX, R12

	// (A,t[11])  := t[11] + x[11]*y[1] + A
	ADCXQ BP, R13
	MULXQ s11-96(SP), AX, BP
	ADOXQ AX, R13

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, BP
	ADOXQ AX, BP
	PUSHQ BP

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R14, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, BP
	ADCXQ R14, AX
	MOVQ  BP, R14
	POPQ  BP

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

	// (C,t[6]) := t[7] + m*q[7] + C
	ADCXQ R9, R8
	MULXQ q<>+56(SB), AX, R9
	ADOXQ AX, R8

	// (C,t[7]) := t[8] + m*q[8] + C
	ADCXQ R10, R9
	MULXQ q<>+64(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[8]) := t[9] + m*q[9] + C
	ADCXQ R11, R10
	MULXQ q<>+72(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[9]) := t[10] + m*q[10] + C
	ADCXQ R12, R11
	MULXQ q<>+80(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[10]) := t[11] + m*q[11] + C
	ADCXQ R13, R12
	MULXQ q<>+88(SB), AX, R13
	ADOXQ AX, R12

	// t[11] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R13
	ADOXQ BP, R13

	// clear the flags
	XORQ AX, AX
	MOVQ y+16(FP), AX
	MOVQ 16(AX), DX

	// (A,t[0])  := t[0] + x[0]*y[2] + A
	MULXQ s0-8(SP), AX, BP
	ADOXQ AX, R14

	// (A,t[1])  := t[1] + x[1]*y[2] + A
	ADCXQ BP, R15
	MULXQ s1-16(SP), AX, BP
	ADOXQ AX, R15

	// (A,t[2])  := t[2] + x[2]*y[2] + A
	ADCXQ BP, CX
	MULXQ s2-24(SP), AX, BP
	ADOXQ AX, CX

	// (A,t[3])  := t[3] + x[3]*y[2] + A
	ADCXQ BP, BX
	MULXQ s3-32(SP), AX, BP
	ADOXQ AX, BX

	// (A,t[4])  := t[4] + x[4]*y[2] + A
	ADCXQ BP, SI
	MULXQ s4-40(SP), AX, BP
	ADOXQ AX, SI

	// (A,t[5])  := t[5] + x[5]*y[2] + A
	ADCXQ BP, DI
	MULXQ s5-48(SP), AX, BP
	ADOXQ AX, DI

	// (A,t[6])  := t[6] + x[6]*y[2] + A
	ADCXQ BP, R8
	MULXQ s6-56(SP), AX, BP
	ADOXQ AX, R8

	// (A,t[7])  := t[7] + x[7]*y[2] + A
	ADCXQ BP, R9
	MULXQ s7-64(SP), AX, BP
	ADOXQ AX, R9

	// (A,t[8])  := t[8] + x[8]*y[2] + A
	ADCXQ BP, R10
	MULXQ s8-72(SP), AX, BP
	ADOXQ AX, R10

	// (A,t[9])  := t[9] + x[9]*y[2] + A
	ADCXQ BP, R11
	MULXQ s9-80(SP), AX, BP
	ADOXQ AX, R11

	// (A,t[10])  := t[10] + x[10]*y[2] + A
	ADCXQ BP, R12
	MULXQ s10-88(SP), AX, BP
	ADOXQ AX, R12

	// (A,t[11])  := t[11] + x[11]*y[2] + A
	ADCXQ BP, R13
	MULXQ s11-96(SP), AX, BP
	ADOXQ AX, R13

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, BP
	ADOXQ AX, BP
	PUSHQ BP

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R14, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, BP
	ADCXQ R14, AX
	MOVQ  BP, R14
	POPQ  BP

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

	// (C,t[6]) := t[7] + m*q[7] + C
	ADCXQ R9, R8
	MULXQ q<>+56(SB), AX, R9
	ADOXQ AX, R8

	// (C,t[7]) := t[8] + m*q[8] + C
	ADCXQ R10, R9
	MULXQ q<>+64(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[8]) := t[9] + m*q[9] + C
	ADCXQ R11, R10
	MULXQ q<>+72(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[9]) := t[10] + m*q[10] + C
	ADCXQ R12, R11
	MULXQ q<>+80(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[10]) := t[11] + m*q[11] + C
	ADCXQ R13, R12
	MULXQ q<>+88(SB), AX, R13
	ADOXQ AX, R12

	// t[11] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R13
	ADOXQ BP, R13

	// clear the flags
	XORQ AX, AX
	MOVQ y+16(FP), AX
	MOVQ 24(AX), DX

	// (A,t[0])  := t[0] + x[0]*y[3] + A
	MULXQ s0-8(SP), AX, BP
	ADOXQ AX, R14

	// (A,t[1])  := t[1] + x[1]*y[3] + A
	ADCXQ BP, R15
	MULXQ s1-16(SP), AX, BP
	ADOXQ AX, R15

	// (A,t[2])  := t[2] + x[2]*y[3] + A
	ADCXQ BP, CX
	MULXQ s2-24(SP), AX, BP
	ADOXQ AX, CX

	// (A,t[3])  := t[3] + x[3]*y[3] + A
	ADCXQ BP, BX
	MULXQ s3-32(SP), AX, BP
	ADOXQ AX, BX

	// (A,t[4])  := t[4] + x[4]*y[3] + A
	ADCXQ BP, SI
	MULXQ s4-40(SP), AX, BP
	ADOXQ AX, SI

	// (A,t[5])  := t[5] + x[5]*y[3] + A
	ADCXQ BP, DI
	MULXQ s5-48(SP), AX, BP
	ADOXQ AX, DI

	// (A,t[6])  := t[6] + x[6]*y[3] + A
	ADCXQ BP, R8
	MULXQ s6-56(SP), AX, BP
	ADOXQ AX, R8

	// (A,t[7])  := t[7] + x[7]*y[3] + A
	ADCXQ BP, R9
	MULXQ s7-64(SP), AX, BP
	ADOXQ AX, R9

	// (A,t[8])  := t[8] + x[8]*y[3] + A
	ADCXQ BP, R10
	MULXQ s8-72(SP), AX, BP
	ADOXQ AX, R10

	// (A,t[9])  := t[9] + x[9]*y[3] + A
	ADCXQ BP, R11
	MULXQ s9-80(SP), AX, BP
	ADOXQ AX, R11

	// (A,t[10])  := t[10] + x[10]*y[3] + A
	ADCXQ BP, R12
	MULXQ s10-88(SP), AX, BP
	ADOXQ AX, R12

	// (A,t[11])  := t[11] + x[11]*y[3] + A
	ADCXQ BP, R13
	MULXQ s11-96(SP), AX, BP
	ADOXQ AX, R13

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, BP
	ADOXQ AX, BP
	PUSHQ BP

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R14, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, BP
	ADCXQ R14, AX
	MOVQ  BP, R14
	POPQ  BP

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

	// (C,t[6]) := t[7] + m*q[7] + C
	ADCXQ R9, R8
	MULXQ q<>+56(SB), AX, R9
	ADOXQ AX, R8

	// (C,t[7]) := t[8] + m*q[8] + C
	ADCXQ R10, R9
	MULXQ q<>+64(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[8]) := t[9] + m*q[9] + C
	ADCXQ R11, R10
	MULXQ q<>+72(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[9]) := t[10] + m*q[10] + C
	ADCXQ R12, R11
	MULXQ q<>+80(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[10]) := t[11] + m*q[11] + C
	ADCXQ R13, R12
	MULXQ q<>+88(SB), AX, R13
	ADOXQ AX, R12

	// t[11] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R13
	ADOXQ BP, R13

	// clear the flags
	XORQ AX, AX
	MOVQ y+16(FP), AX
	MOVQ 32(AX), DX

	// (A,t[0])  := t[0] + x[0]*y[4] + A
	MULXQ s0-8(SP), AX, BP
	ADOXQ AX, R14

	// (A,t[1])  := t[1] + x[1]*y[4] + A
	ADCXQ BP, R15
	MULXQ s1-16(SP), AX, BP
	ADOXQ AX, R15

	// (A,t[2])  := t[2] + x[2]*y[4] + A
	ADCXQ BP, CX
	MULXQ s2-24(SP), AX, BP
	ADOXQ AX, CX

	// (A,t[3])  := t[3] + x[3]*y[4] + A
	ADCXQ BP, BX
	MULXQ s3-32(SP), AX, BP
	ADOXQ AX, BX

	// (A,t[4])  := t[4] + x[4]*y[4] + A
	ADCXQ BP, SI
	MULXQ s4-40(SP), AX, BP
	ADOXQ AX, SI

	// (A,t[5])  := t[5] + x[5]*y[4] + A
	ADCXQ BP, DI
	MULXQ s5-48(SP), AX, BP
	ADOXQ AX, DI

	// (A,t[6])  := t[6] + x[6]*y[4] + A
	ADCXQ BP, R8
	MULXQ s6-56(SP), AX, BP
	ADOXQ AX, R8

	// (A,t[7])  := t[7] + x[7]*y[4] + A
	ADCXQ BP, R9
	MULXQ s7-64(SP), AX, BP
	ADOXQ AX, R9

	// (A,t[8])  := t[8] + x[8]*y[4] + A
	ADCXQ BP, R10
	MULXQ s8-72(SP), AX, BP
	ADOXQ AX, R10

	// (A,t[9])  := t[9] + x[9]*y[4] + A
	ADCXQ BP, R11
	MULXQ s9-80(SP), AX, BP
	ADOXQ AX, R11

	// (A,t[10])  := t[10] + x[10]*y[4] + A
	ADCXQ BP, R12
	MULXQ s10-88(SP), AX, BP
	ADOXQ AX, R12

	// (A,t[11])  := t[11] + x[11]*y[4] + A
	ADCXQ BP, R13
	MULXQ s11-96(SP), AX, BP
	ADOXQ AX, R13

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, BP
	ADOXQ AX, BP
	PUSHQ BP

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R14, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, BP
	ADCXQ R14, AX
	MOVQ  BP, R14
	POPQ  BP

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

	// (C,t[6]) := t[7] + m*q[7] + C
	ADCXQ R9, R8
	MULXQ q<>+56(SB), AX, R9
	ADOXQ AX, R8

	// (C,t[7]) := t[8] + m*q[8] + C
	ADCXQ R10, R9
	MULXQ q<>+64(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[8]) := t[9] + m*q[9] + C
	ADCXQ R11, R10
	MULXQ q<>+72(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[9]) := t[10] + m*q[10] + C
	ADCXQ R12, R11
	MULXQ q<>+80(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[10]) := t[11] + m*q[11] + C
	ADCXQ R13, R12
	MULXQ q<>+88(SB), AX, R13
	ADOXQ AX, R12

	// t[11] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R13
	ADOXQ BP, R13

	// clear the flags
	XORQ AX, AX
	MOVQ y+16(FP), AX
	MOVQ 40(AX), DX

	// (A,t[0])  := t[0] + x[0]*y[5] + A
	MULXQ s0-8(SP), AX, BP
	ADOXQ AX, R14

	// (A,t[1])  := t[1] + x[1]*y[5] + A
	ADCXQ BP, R15
	MULXQ s1-16(SP), AX, BP
	ADOXQ AX, R15

	// (A,t[2])  := t[2] + x[2]*y[5] + A
	ADCXQ BP, CX
	MULXQ s2-24(SP), AX, BP
	ADOXQ AX, CX

	// (A,t[3])  := t[3] + x[3]*y[5] + A
	ADCXQ BP, BX
	MULXQ s3-32(SP), AX, BP
	ADOXQ AX, BX

	// (A,t[4])  := t[4] + x[4]*y[5] + A
	ADCXQ BP, SI
	MULXQ s4-40(SP), AX, BP
	ADOXQ AX, SI

	// (A,t[5])  := t[5] + x[5]*y[5] + A
	ADCXQ BP, DI
	MULXQ s5-48(SP), AX, BP
	ADOXQ AX, DI

	// (A,t[6])  := t[6] + x[6]*y[5] + A
	ADCXQ BP, R8
	MULXQ s6-56(SP), AX, BP
	ADOXQ AX, R8

	// (A,t[7])  := t[7] + x[7]*y[5] + A
	ADCXQ BP, R9
	MULXQ s7-64(SP), AX, BP
	ADOXQ AX, R9

	// (A,t[8])  := t[8] + x[8]*y[5] + A
	ADCXQ BP, R10
	MULXQ s8-72(SP), AX, BP
	ADOXQ AX, R10

	// (A,t[9])  := t[9] + x[9]*y[5] + A
	ADCXQ BP, R11
	MULXQ s9-80(SP), AX, BP
	ADOXQ AX, R11

	// (A,t[10])  := t[10] + x[10]*y[5] + A
	ADCXQ BP, R12
	MULXQ s10-88(SP), AX, BP
	ADOXQ AX, R12

	// (A,t[11])  := t[11] + x[11]*y[5] + A
	ADCXQ BP, R13
	MULXQ s11-96(SP), AX, BP
	ADOXQ AX, R13

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, BP
	ADOXQ AX, BP
	PUSHQ BP

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R14, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, BP
	ADCXQ R14, AX
	MOVQ  BP, R14
	POPQ  BP

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

	// (C,t[6]) := t[7] + m*q[7] + C
	ADCXQ R9, R8
	MULXQ q<>+56(SB), AX, R9
	ADOXQ AX, R8

	// (C,t[7]) := t[8] + m*q[8] + C
	ADCXQ R10, R9
	MULXQ q<>+64(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[8]) := t[9] + m*q[9] + C
	ADCXQ R11, R10
	MULXQ q<>+72(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[9]) := t[10] + m*q[10] + C
	ADCXQ R12, R11
	MULXQ q<>+80(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[10]) := t[11] + m*q[11] + C
	ADCXQ R13, R12
	MULXQ q<>+88(SB), AX, R13
	ADOXQ AX, R12

	// t[11] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R13
	ADOXQ BP, R13

	// clear the flags
	XORQ AX, AX
	MOVQ y+16(FP), AX
	MOVQ 48(AX), DX

	// (A,t[0])  := t[0] + x[0]*y[6] + A
	MULXQ s0-8(SP), AX, BP
	ADOXQ AX, R14

	// (A,t[1])  := t[1] + x[1]*y[6] + A
	ADCXQ BP, R15
	MULXQ s1-16(SP), AX, BP
	ADOXQ AX, R15

	// (A,t[2])  := t[2] + x[2]*y[6] + A
	ADCXQ BP, CX
	MULXQ s2-24(SP), AX, BP
	ADOXQ AX, CX

	// (A,t[3])  := t[3] + x[3]*y[6] + A
	ADCXQ BP, BX
	MULXQ s3-32(SP), AX, BP
	ADOXQ AX, BX

	// (A,t[4])  := t[4] + x[4]*y[6] + A
	ADCXQ BP, SI
	MULXQ s4-40(SP), AX, BP
	ADOXQ AX, SI

	// (A,t[5])  := t[5] + x[5]*y[6] + A
	ADCXQ BP, DI
	MULXQ s5-48(SP), AX, BP
	ADOXQ AX, DI

	// (A,t[6])  := t[6] + x[6]*y[6] + A
	ADCXQ BP, R8
	MULXQ s6-56(SP), AX, BP
	ADOXQ AX, R8

	// (A,t[7])  := t[7] + x[7]*y[6] + A
	ADCXQ BP, R9
	MULXQ s7-64(SP), AX, BP
	ADOXQ AX, R9

	// (A,t[8])  := t[8] + x[8]*y[6] + A
	ADCXQ BP, R10
	MULXQ s8-72(SP), AX, BP
	ADOXQ AX, R10

	// (A,t[9])  := t[9] + x[9]*y[6] + A
	ADCXQ BP, R11
	MULXQ s9-80(SP), AX, BP
	ADOXQ AX, R11

	// (A,t[10])  := t[10] + x[10]*y[6] + A
	ADCXQ BP, R12
	MULXQ s10-88(SP), AX, BP
	ADOXQ AX, R12

	// (A,t[11])  := t[11] + x[11]*y[6] + A
	ADCXQ BP, R13
	MULXQ s11-96(SP), AX, BP
	ADOXQ AX, R13

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, BP
	ADOXQ AX, BP
	PUSHQ BP

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R14, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, BP
	ADCXQ R14, AX
	MOVQ  BP, R14
	POPQ  BP

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

	// (C,t[6]) := t[7] + m*q[7] + C
	ADCXQ R9, R8
	MULXQ q<>+56(SB), AX, R9
	ADOXQ AX, R8

	// (C,t[7]) := t[8] + m*q[8] + C
	ADCXQ R10, R9
	MULXQ q<>+64(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[8]) := t[9] + m*q[9] + C
	ADCXQ R11, R10
	MULXQ q<>+72(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[9]) := t[10] + m*q[10] + C
	ADCXQ R12, R11
	MULXQ q<>+80(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[10]) := t[11] + m*q[11] + C
	ADCXQ R13, R12
	MULXQ q<>+88(SB), AX, R13
	ADOXQ AX, R12

	// t[11] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R13
	ADOXQ BP, R13

	// clear the flags
	XORQ AX, AX
	MOVQ y+16(FP), AX
	MOVQ 56(AX), DX

	// (A,t[0])  := t[0] + x[0]*y[7] + A
	MULXQ s0-8(SP), AX, BP
	ADOXQ AX, R14

	// (A,t[1])  := t[1] + x[1]*y[7] + A
	ADCXQ BP, R15
	MULXQ s1-16(SP), AX, BP
	ADOXQ AX, R15

	// (A,t[2])  := t[2] + x[2]*y[7] + A
	ADCXQ BP, CX
	MULXQ s2-24(SP), AX, BP
	ADOXQ AX, CX

	// (A,t[3])  := t[3] + x[3]*y[7] + A
	ADCXQ BP, BX
	MULXQ s3-32(SP), AX, BP
	ADOXQ AX, BX

	// (A,t[4])  := t[4] + x[4]*y[7] + A
	ADCXQ BP, SI
	MULXQ s4-40(SP), AX, BP
	ADOXQ AX, SI

	// (A,t[5])  := t[5] + x[5]*y[7] + A
	ADCXQ BP, DI
	MULXQ s5-48(SP), AX, BP
	ADOXQ AX, DI

	// (A,t[6])  := t[6] + x[6]*y[7] + A
	ADCXQ BP, R8
	MULXQ s6-56(SP), AX, BP
	ADOXQ AX, R8

	// (A,t[7])  := t[7] + x[7]*y[7] + A
	ADCXQ BP, R9
	MULXQ s7-64(SP), AX, BP
	ADOXQ AX, R9

	// (A,t[8])  := t[8] + x[8]*y[7] + A
	ADCXQ BP, R10
	MULXQ s8-72(SP), AX, BP
	ADOXQ AX, R10

	// (A,t[9])  := t[9] + x[9]*y[7] + A
	ADCXQ BP, R11
	MULXQ s9-80(SP), AX, BP
	ADOXQ AX, R11

	// (A,t[10])  := t[10] + x[10]*y[7] + A
	ADCXQ BP, R12
	MULXQ s10-88(SP), AX, BP
	ADOXQ AX, R12

	// (A,t[11])  := t[11] + x[11]*y[7] + A
	ADCXQ BP, R13
	MULXQ s11-96(SP), AX, BP
	ADOXQ AX, R13

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, BP
	ADOXQ AX, BP
	PUSHQ BP

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R14, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, BP
	ADCXQ R14, AX
	MOVQ  BP, R14
	POPQ  BP

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

	// (C,t[6]) := t[7] + m*q[7] + C
	ADCXQ R9, R8
	MULXQ q<>+56(SB), AX, R9
	ADOXQ AX, R8

	// (C,t[7]) := t[8] + m*q[8] + C
	ADCXQ R10, R9
	MULXQ q<>+64(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[8]) := t[9] + m*q[9] + C
	ADCXQ R11, R10
	MULXQ q<>+72(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[9]) := t[10] + m*q[10] + C
	ADCXQ R12, R11
	MULXQ q<>+80(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[10]) := t[11] + m*q[11] + C
	ADCXQ R13, R12
	MULXQ q<>+88(SB), AX, R13
	ADOXQ AX, R12

	// t[11] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R13
	ADOXQ BP, R13

	// clear the flags
	XORQ AX, AX
	MOVQ y+16(FP), AX
	MOVQ 64(AX), DX

	// (A,t[0])  := t[0] + x[0]*y[8] + A
	MULXQ s0-8(SP), AX, BP
	ADOXQ AX, R14

	// (A,t[1])  := t[1] + x[1]*y[8] + A
	ADCXQ BP, R15
	MULXQ s1-16(SP), AX, BP
	ADOXQ AX, R15

	// (A,t[2])  := t[2] + x[2]*y[8] + A
	ADCXQ BP, CX
	MULXQ s2-24(SP), AX, BP
	ADOXQ AX, CX

	// (A,t[3])  := t[3] + x[3]*y[8] + A
	ADCXQ BP, BX
	MULXQ s3-32(SP), AX, BP
	ADOXQ AX, BX

	// (A,t[4])  := t[4] + x[4]*y[8] + A
	ADCXQ BP, SI
	MULXQ s4-40(SP), AX, BP
	ADOXQ AX, SI

	// (A,t[5])  := t[5] + x[5]*y[8] + A
	ADCXQ BP, DI
	MULXQ s5-48(SP), AX, BP
	ADOXQ AX, DI

	// (A,t[6])  := t[6] + x[6]*y[8] + A
	ADCXQ BP, R8
	MULXQ s6-56(SP), AX, BP
	ADOXQ AX, R8

	// (A,t[7])  := t[7] + x[7]*y[8] + A
	ADCXQ BP, R9
	MULXQ s7-64(SP), AX, BP
	ADOXQ AX, R9

	// (A,t[8])  := t[8] + x[8]*y[8] + A
	ADCXQ BP, R10
	MULXQ s8-72(SP), AX, BP
	ADOXQ AX, R10

	// (A,t[9])  := t[9] + x[9]*y[8] + A
	ADCXQ BP, R11
	MULXQ s9-80(SP), AX, BP
	ADOXQ AX, R11

	// (A,t[10])  := t[10] + x[10]*y[8] + A
	ADCXQ BP, R12
	MULXQ s10-88(SP), AX, BP
	ADOXQ AX, R12

	// (A,t[11])  := t[11] + x[11]*y[8] + A
	ADCXQ BP, R13
	MULXQ s11-96(SP), AX, BP
	ADOXQ AX, R13

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, BP
	ADOXQ AX, BP
	PUSHQ BP

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R14, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, BP
	ADCXQ R14, AX
	MOVQ  BP, R14
	POPQ  BP

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

	// (C,t[6]) := t[7] + m*q[7] + C
	ADCXQ R9, R8
	MULXQ q<>+56(SB), AX, R9
	ADOXQ AX, R8

	// (C,t[7]) := t[8] + m*q[8] + C
	ADCXQ R10, R9
	MULXQ q<>+64(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[8]) := t[9] + m*q[9] + C
	ADCXQ R11, R10
	MULXQ q<>+72(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[9]) := t[10] + m*q[10] + C
	ADCXQ R12, R11
	MULXQ q<>+80(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[10]) := t[11] + m*q[11] + C
	ADCXQ R13, R12
	MULXQ q<>+88(SB), AX, R13
	ADOXQ AX, R12

	// t[11] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R13
	ADOXQ BP, R13

	// clear the flags
	XORQ AX, AX
	MOVQ y+16(FP), AX
	MOVQ 72(AX), DX

	// (A,t[0])  := t[0] + x[0]*y[9] + A
	MULXQ s0-8(SP), AX, BP
	ADOXQ AX, R14

	// (A,t[1])  := t[1] + x[1]*y[9] + A
	ADCXQ BP, R15
	MULXQ s1-16(SP), AX, BP
	ADOXQ AX, R15

	// (A,t[2])  := t[2] + x[2]*y[9] + A
	ADCXQ BP, CX
	MULXQ s2-24(SP), AX, BP
	ADOXQ AX, CX

	// (A,t[3])  := t[3] + x[3]*y[9] + A
	ADCXQ BP, BX
	MULXQ s3-32(SP), AX, BP
	ADOXQ AX, BX

	// (A,t[4])  := t[4] + x[4]*y[9] + A
	ADCXQ BP, SI
	MULXQ s4-40(SP), AX, BP
	ADOXQ AX, SI

	// (A,t[5])  := t[5] + x[5]*y[9] + A
	ADCXQ BP, DI
	MULXQ s5-48(SP), AX, BP
	ADOXQ AX, DI

	// (A,t[6])  := t[6] + x[6]*y[9] + A
	ADCXQ BP, R8
	MULXQ s6-56(SP), AX, BP
	ADOXQ AX, R8

	// (A,t[7])  := t[7] + x[7]*y[9] + A
	ADCXQ BP, R9
	MULXQ s7-64(SP), AX, BP
	ADOXQ AX, R9

	// (A,t[8])  := t[8] + x[8]*y[9] + A
	ADCXQ BP, R10
	MULXQ s8-72(SP), AX, BP
	ADOXQ AX, R10

	// (A,t[9])  := t[9] + x[9]*y[9] + A
	ADCXQ BP, R11
	MULXQ s9-80(SP), AX, BP
	ADOXQ AX, R11

	// (A,t[10])  := t[10] + x[10]*y[9] + A
	ADCXQ BP, R12
	MULXQ s10-88(SP), AX, BP
	ADOXQ AX, R12

	// (A,t[11])  := t[11] + x[11]*y[9] + A
	ADCXQ BP, R13
	MULXQ s11-96(SP), AX, BP
	ADOXQ AX, R13

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, BP
	ADOXQ AX, BP
	PUSHQ BP

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R14, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, BP
	ADCXQ R14, AX
	MOVQ  BP, R14
	POPQ  BP

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

	// (C,t[6]) := t[7] + m*q[7] + C
	ADCXQ R9, R8
	MULXQ q<>+56(SB), AX, R9
	ADOXQ AX, R8

	// (C,t[7]) := t[8] + m*q[8] + C
	ADCXQ R10, R9
	MULXQ q<>+64(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[8]) := t[9] + m*q[9] + C
	ADCXQ R11, R10
	MULXQ q<>+72(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[9]) := t[10] + m*q[10] + C
	ADCXQ R12, R11
	MULXQ q<>+80(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[10]) := t[11] + m*q[11] + C
	ADCXQ R13, R12
	MULXQ q<>+88(SB), AX, R13
	ADOXQ AX, R12

	// t[11] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R13
	ADOXQ BP, R13

	// clear the flags
	XORQ AX, AX
	MOVQ y+16(FP), AX
	MOVQ 80(AX), DX

	// (A,t[0])  := t[0] + x[0]*y[10] + A
	MULXQ s0-8(SP), AX, BP
	ADOXQ AX, R14

	// (A,t[1])  := t[1] + x[1]*y[10] + A
	ADCXQ BP, R15
	MULXQ s1-16(SP), AX, BP
	ADOXQ AX, R15

	// (A,t[2])  := t[2] + x[2]*y[10] + A
	ADCXQ BP, CX
	MULXQ s2-24(SP), AX, BP
	ADOXQ AX, CX

	// (A,t[3])  := t[3] + x[3]*y[10] + A
	ADCXQ BP, BX
	MULXQ s3-32(SP), AX, BP
	ADOXQ AX, BX

	// (A,t[4])  := t[4] + x[4]*y[10] + A
	ADCXQ BP, SI
	MULXQ s4-40(SP), AX, BP
	ADOXQ AX, SI

	// (A,t[5])  := t[5] + x[5]*y[10] + A
	ADCXQ BP, DI
	MULXQ s5-48(SP), AX, BP
	ADOXQ AX, DI

	// (A,t[6])  := t[6] + x[6]*y[10] + A
	ADCXQ BP, R8
	MULXQ s6-56(SP), AX, BP
	ADOXQ AX, R8

	// (A,t[7])  := t[7] + x[7]*y[10] + A
	ADCXQ BP, R9
	MULXQ s7-64(SP), AX, BP
	ADOXQ AX, R9

	// (A,t[8])  := t[8] + x[8]*y[10] + A
	ADCXQ BP, R10
	MULXQ s8-72(SP), AX, BP
	ADOXQ AX, R10

	// (A,t[9])  := t[9] + x[9]*y[10] + A
	ADCXQ BP, R11
	MULXQ s9-80(SP), AX, BP
	ADOXQ AX, R11

	// (A,t[10])  := t[10] + x[10]*y[10] + A
	ADCXQ BP, R12
	MULXQ s10-88(SP), AX, BP
	ADOXQ AX, R12

	// (A,t[11])  := t[11] + x[11]*y[10] + A
	ADCXQ BP, R13
	MULXQ s11-96(SP), AX, BP
	ADOXQ AX, R13

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, BP
	ADOXQ AX, BP
	PUSHQ BP

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R14, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, BP
	ADCXQ R14, AX
	MOVQ  BP, R14
	POPQ  BP

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

	// (C,t[6]) := t[7] + m*q[7] + C
	ADCXQ R9, R8
	MULXQ q<>+56(SB), AX, R9
	ADOXQ AX, R8

	// (C,t[7]) := t[8] + m*q[8] + C
	ADCXQ R10, R9
	MULXQ q<>+64(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[8]) := t[9] + m*q[9] + C
	ADCXQ R11, R10
	MULXQ q<>+72(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[9]) := t[10] + m*q[10] + C
	ADCXQ R12, R11
	MULXQ q<>+80(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[10]) := t[11] + m*q[11] + C
	ADCXQ R13, R12
	MULXQ q<>+88(SB), AX, R13
	ADOXQ AX, R12

	// t[11] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R13
	ADOXQ BP, R13

	// clear the flags
	XORQ AX, AX
	MOVQ y+16(FP), AX
	MOVQ 88(AX), DX

	// (A,t[0])  := t[0] + x[0]*y[11] + A
	MULXQ s0-8(SP), AX, BP
	ADOXQ AX, R14

	// (A,t[1])  := t[1] + x[1]*y[11] + A
	ADCXQ BP, R15
	MULXQ s1-16(SP), AX, BP
	ADOXQ AX, R15

	// (A,t[2])  := t[2] + x[2]*y[11] + A
	ADCXQ BP, CX
	MULXQ s2-24(SP), AX, BP
	ADOXQ AX, CX

	// (A,t[3])  := t[3] + x[3]*y[11] + A
	ADCXQ BP, BX
	MULXQ s3-32(SP), AX, BP
	ADOXQ AX, BX

	// (A,t[4])  := t[4] + x[4]*y[11] + A
	ADCXQ BP, SI
	MULXQ s4-40(SP), AX, BP
	ADOXQ AX, SI

	// (A,t[5])  := t[5] + x[5]*y[11] + A
	ADCXQ BP, DI
	MULXQ s5-48(SP), AX, BP
	ADOXQ AX, DI

	// (A,t[6])  := t[6] + x[6]*y[11] + A
	ADCXQ BP, R8
	MULXQ s6-56(SP), AX, BP
	ADOXQ AX, R8

	// (A,t[7])  := t[7] + x[7]*y[11] + A
	ADCXQ BP, R9
	MULXQ s7-64(SP), AX, BP
	ADOXQ AX, R9

	// (A,t[8])  := t[8] + x[8]*y[11] + A
	ADCXQ BP, R10
	MULXQ s8-72(SP), AX, BP
	ADOXQ AX, R10

	// (A,t[9])  := t[9] + x[9]*y[11] + A
	ADCXQ BP, R11
	MULXQ s9-80(SP), AX, BP
	ADOXQ AX, R11

	// (A,t[10])  := t[10] + x[10]*y[11] + A
	ADCXQ BP, R12
	MULXQ s10-88(SP), AX, BP
	ADOXQ AX, R12

	// (A,t[11])  := t[11] + x[11]*y[11] + A
	ADCXQ BP, R13
	MULXQ s11-96(SP), AX, BP
	ADOXQ AX, R13

	// A += carries from ADCXQ and ADOXQ
	MOVQ  $0, AX
	ADCXQ AX, BP
	ADOXQ AX, BP
	PUSHQ BP

	// m := t[0]*q'[0] mod W
	MOVQ  qInv0<>(SB), DX
	IMULQ R14, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ q<>+0(SB), AX, BP
	ADCXQ R14, AX
	MOVQ  BP, R14
	POPQ  BP

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

	// (C,t[6]) := t[7] + m*q[7] + C
	ADCXQ R9, R8
	MULXQ q<>+56(SB), AX, R9
	ADOXQ AX, R8

	// (C,t[7]) := t[8] + m*q[8] + C
	ADCXQ R10, R9
	MULXQ q<>+64(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[8]) := t[9] + m*q[9] + C
	ADCXQ R11, R10
	MULXQ q<>+72(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[9]) := t[10] + m*q[10] + C
	ADCXQ R12, R11
	MULXQ q<>+80(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[10]) := t[11] + m*q[11] + C
	ADCXQ R13, R12
	MULXQ q<>+88(SB), AX, R13
	ADOXQ AX, R12

	// t[11] = C + A
	MOVQ  $0, AX
	ADCXQ AX, R13
	ADOXQ BP, R13

	// reduce element(R14,R15,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13) using temp registers (s0-8(SP),s1-16(SP),s2-24(SP),s3-32(SP),s4-40(SP),s5-48(SP),s6-56(SP),s7-64(SP),s8-72(SP),s9-80(SP),s10-88(SP),s11-96(SP))
	REDUCE(R14,R15,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,s0-8(SP),s1-16(SP),s2-24(SP),s3-32(SP),s4-40(SP),s5-48(SP),s6-56(SP),s7-64(SP),s8-72(SP),s9-80(SP),s10-88(SP),s11-96(SP))

	MOVQ res+0(FP), AX
	MOVQ R14, 0(AX)
	MOVQ R15, 8(AX)
	MOVQ CX, 16(AX)
	MOVQ BX, 24(AX)
	MOVQ SI, 32(AX)
	MOVQ DI, 40(AX)
	MOVQ R8, 48(AX)
	MOVQ R9, 56(AX)
	MOVQ R10, 64(AX)
	MOVQ R11, 72(AX)
	MOVQ R12, 80(AX)
	MOVQ R13, 88(AX)
	RET

noAdx_2:
	MOVQ res+0(FP), AX
	MOVQ AX, (SP)
	MOVQ x+8(FP), AX
	MOVQ AX, 8(SP)
	MOVQ y+16(FP), AX
	MOVQ AX, 16(SP)
	CALL ·_mulGeneric(SB)
	RET

TEXT ·fromMont(SB), $96-8
	NO_LOCAL_POINTERS

	// Algorithm 2 of "Faster Montgomery Multiplication and Multi-Scalar-Multiplication for SNARKS"
	// by Y. El Housni and G. Botrel https://doi.org/10.46586/tches.v2023.i3.504-521
	// when y = 1 we have:
	// for i=0 to N-1
	// 		t[i] = x[i]
	// for i=0 to N-1
	// 		m := t[0]*q'[0] mod W
	// 		C,_ := t[0] + m*q[0]
	// 		for j=1 to N-1
	// 		    (C,t[j-1]) := t[j] + m*q[j] + C
	// 		t[N-1] = C
	CMPB ·supportAdx(SB), $1
	JNE  noAdx_3
	MOVQ res+0(FP), DX
	MOVQ 0(DX), R14
	MOVQ 8(DX), R15
	MOVQ 16(DX), CX
	MOVQ 24(DX), BX
	MOVQ 32(DX), SI
	MOVQ 40(DX), DI
	MOVQ 48(DX), R8
	MOVQ 56(DX), R9
	MOVQ 64(DX), R10
	MOVQ 72(DX), R11
	MOVQ 80(DX), R12
	MOVQ 88(DX), R13
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

	// (C,t[6]) := t[7] + m*q[7] + C
	ADCXQ R9, R8
	MULXQ q<>+56(SB), AX, R9
	ADOXQ AX, R8

	// (C,t[7]) := t[8] + m*q[8] + C
	ADCXQ R10, R9
	MULXQ q<>+64(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[8]) := t[9] + m*q[9] + C
	ADCXQ R11, R10
	MULXQ q<>+72(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[9]) := t[10] + m*q[10] + C
	ADCXQ R12, R11
	MULXQ q<>+80(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[10]) := t[11] + m*q[11] + C
	ADCXQ R13, R12
	MULXQ q<>+88(SB), AX, R13
	ADOXQ AX, R12
	MOVQ  $0, AX
	ADCXQ AX, R13
	ADOXQ AX, R13
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

	// (C,t[6]) := t[7] + m*q[7] + C
	ADCXQ R9, R8
	MULXQ q<>+56(SB), AX, R9
	ADOXQ AX, R8

	// (C,t[7]) := t[8] + m*q[8] + C
	ADCXQ R10, R9
	MULXQ q<>+64(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[8]) := t[9] + m*q[9] + C
	ADCXQ R11, R10
	MULXQ q<>+72(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[9]) := t[10] + m*q[10] + C
	ADCXQ R12, R11
	MULXQ q<>+80(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[10]) := t[11] + m*q[11] + C
	ADCXQ R13, R12
	MULXQ q<>+88(SB), AX, R13
	ADOXQ AX, R12
	MOVQ  $0, AX
	ADCXQ AX, R13
	ADOXQ AX, R13
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

	// (C,t[6]) := t[7] + m*q[7] + C
	ADCXQ R9, R8
	MULXQ q<>+56(SB), AX, R9
	ADOXQ AX, R8

	// (C,t[7]) := t[8] + m*q[8] + C
	ADCXQ R10, R9
	MULXQ q<>+64(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[8]) := t[9] + m*q[9] + C
	ADCXQ R11, R10
	MULXQ q<>+72(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[9]) := t[10] + m*q[10] + C
	ADCXQ R12, R11
	MULXQ q<>+80(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[10]) := t[11] + m*q[11] + C
	ADCXQ R13, R12
	MULXQ q<>+88(SB), AX, R13
	ADOXQ AX, R12
	MOVQ  $0, AX
	ADCXQ AX, R13
	ADOXQ AX, R13
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

	// (C,t[6]) := t[7] + m*q[7] + C
	ADCXQ R9, R8
	MULXQ q<>+56(SB), AX, R9
	ADOXQ AX, R8

	// (C,t[7]) := t[8] + m*q[8] + C
	ADCXQ R10, R9
	MULXQ q<>+64(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[8]) := t[9] + m*q[9] + C
	ADCXQ R11, R10
	MULXQ q<>+72(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[9]) := t[10] + m*q[10] + C
	ADCXQ R12, R11
	MULXQ q<>+80(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[10]) := t[11] + m*q[11] + C
	ADCXQ R13, R12
	MULXQ q<>+88(SB), AX, R13
	ADOXQ AX, R12
	MOVQ  $0, AX
	ADCXQ AX, R13
	ADOXQ AX, R13
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

	// (C,t[6]) := t[7] + m*q[7] + C
	ADCXQ R9, R8
	MULXQ q<>+56(SB), AX, R9
	ADOXQ AX, R8

	// (C,t[7]) := t[8] + m*q[8] + C
	ADCXQ R10, R9
	MULXQ q<>+64(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[8]) := t[9] + m*q[9] + C
	ADCXQ R11, R10
	MULXQ q<>+72(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[9]) := t[10] + m*q[10] + C
	ADCXQ R12, R11
	MULXQ q<>+80(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[10]) := t[11] + m*q[11] + C
	ADCXQ R13, R12
	MULXQ q<>+88(SB), AX, R13
	ADOXQ AX, R12
	MOVQ  $0, AX
	ADCXQ AX, R13
	ADOXQ AX, R13
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

	// (C,t[6]) := t[7] + m*q[7] + C
	ADCXQ R9, R8
	MULXQ q<>+56(SB), AX, R9
	ADOXQ AX, R8

	// (C,t[7]) := t[8] + m*q[8] + C
	ADCXQ R10, R9
	MULXQ q<>+64(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[8]) := t[9] + m*q[9] + C
	ADCXQ R11, R10
	MULXQ q<>+72(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[9]) := t[10] + m*q[10] + C
	ADCXQ R12, R11
	MULXQ q<>+80(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[10]) := t[11] + m*q[11] + C
	ADCXQ R13, R12
	MULXQ q<>+88(SB), AX, R13
	ADOXQ AX, R12
	MOVQ  $0, AX
	ADCXQ AX, R13
	ADOXQ AX, R13
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

	// (C,t[6]) := t[7] + m*q[7] + C
	ADCXQ R9, R8
	MULXQ q<>+56(SB), AX, R9
	ADOXQ AX, R8

	// (C,t[7]) := t[8] + m*q[8] + C
	ADCXQ R10, R9
	MULXQ q<>+64(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[8]) := t[9] + m*q[9] + C
	ADCXQ R11, R10
	MULXQ q<>+72(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[9]) := t[10] + m*q[10] + C
	ADCXQ R12, R11
	MULXQ q<>+80(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[10]) := t[11] + m*q[11] + C
	ADCXQ R13, R12
	MULXQ q<>+88(SB), AX, R13
	ADOXQ AX, R12
	MOVQ  $0, AX
	ADCXQ AX, R13
	ADOXQ AX, R13
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

	// (C,t[6]) := t[7] + m*q[7] + C
	ADCXQ R9, R8
	MULXQ q<>+56(SB), AX, R9
	ADOXQ AX, R8

	// (C,t[7]) := t[8] + m*q[8] + C
	ADCXQ R10, R9
	MULXQ q<>+64(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[8]) := t[9] + m*q[9] + C
	ADCXQ R11, R10
	MULXQ q<>+72(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[9]) := t[10] + m*q[10] + C
	ADCXQ R12, R11
	MULXQ q<>+80(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[10]) := t[11] + m*q[11] + C
	ADCXQ R13, R12
	MULXQ q<>+88(SB), AX, R13
	ADOXQ AX, R12
	MOVQ  $0, AX
	ADCXQ AX, R13
	ADOXQ AX, R13
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

	// (C,t[6]) := t[7] + m*q[7] + C
	ADCXQ R9, R8
	MULXQ q<>+56(SB), AX, R9
	ADOXQ AX, R8

	// (C,t[7]) := t[8] + m*q[8] + C
	ADCXQ R10, R9
	MULXQ q<>+64(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[8]) := t[9] + m*q[9] + C
	ADCXQ R11, R10
	MULXQ q<>+72(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[9]) := t[10] + m*q[10] + C
	ADCXQ R12, R11
	MULXQ q<>+80(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[10]) := t[11] + m*q[11] + C
	ADCXQ R13, R12
	MULXQ q<>+88(SB), AX, R13
	ADOXQ AX, R12
	MOVQ  $0, AX
	ADCXQ AX, R13
	ADOXQ AX, R13
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

	// (C,t[6]) := t[7] + m*q[7] + C
	ADCXQ R9, R8
	MULXQ q<>+56(SB), AX, R9
	ADOXQ AX, R8

	// (C,t[7]) := t[8] + m*q[8] + C
	ADCXQ R10, R9
	MULXQ q<>+64(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[8]) := t[9] + m*q[9] + C
	ADCXQ R11, R10
	MULXQ q<>+72(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[9]) := t[10] + m*q[10] + C
	ADCXQ R12, R11
	MULXQ q<>+80(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[10]) := t[11] + m*q[11] + C
	ADCXQ R13, R12
	MULXQ q<>+88(SB), AX, R13
	ADOXQ AX, R12
	MOVQ  $0, AX
	ADCXQ AX, R13
	ADOXQ AX, R13
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

	// (C,t[6]) := t[7] + m*q[7] + C
	ADCXQ R9, R8
	MULXQ q<>+56(SB), AX, R9
	ADOXQ AX, R8

	// (C,t[7]) := t[8] + m*q[8] + C
	ADCXQ R10, R9
	MULXQ q<>+64(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[8]) := t[9] + m*q[9] + C
	ADCXQ R11, R10
	MULXQ q<>+72(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[9]) := t[10] + m*q[10] + C
	ADCXQ R12, R11
	MULXQ q<>+80(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[10]) := t[11] + m*q[11] + C
	ADCXQ R13, R12
	MULXQ q<>+88(SB), AX, R13
	ADOXQ AX, R12
	MOVQ  $0, AX
	ADCXQ AX, R13
	ADOXQ AX, R13
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

	// (C,t[6]) := t[7] + m*q[7] + C
	ADCXQ R9, R8
	MULXQ q<>+56(SB), AX, R9
	ADOXQ AX, R8

	// (C,t[7]) := t[8] + m*q[8] + C
	ADCXQ R10, R9
	MULXQ q<>+64(SB), AX, R10
	ADOXQ AX, R9

	// (C,t[8]) := t[9] + m*q[9] + C
	ADCXQ R11, R10
	MULXQ q<>+72(SB), AX, R11
	ADOXQ AX, R10

	// (C,t[9]) := t[10] + m*q[10] + C
	ADCXQ R12, R11
	MULXQ q<>+80(SB), AX, R12
	ADOXQ AX, R11

	// (C,t[10]) := t[11] + m*q[11] + C
	ADCXQ R13, R12
	MULXQ q<>+88(SB), AX, R13
	ADOXQ AX, R12
	MOVQ  $0, AX
	ADCXQ AX, R13
	ADOXQ AX, R13

	// reduce element(R14,R15,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13) using temp registers (s0-8(SP),s1-16(SP),s2-24(SP),s3-32(SP),s4-40(SP),s5-48(SP),s6-56(SP),s7-64(SP),s8-72(SP),s9-80(SP),s10-88(SP),s11-96(SP))
	REDUCE(R14,R15,CX,BX,SI,DI,R8,R9,R10,R11,R12,R13,s0-8(SP),s1-16(SP),s2-24(SP),s3-32(SP),s4-40(SP),s5-48(SP),s6-56(SP),s7-64(SP),s8-72(SP),s9-80(SP),s10-88(SP),s11-96(SP))

	MOVQ res+0(FP), AX
	MOVQ R14, 0(AX)
	MOVQ R15, 8(AX)
	MOVQ CX, 16(AX)
	MOVQ BX, 24(AX)
	MOVQ SI, 32(AX)
	MOVQ DI, 40(AX)
	MOVQ R8, 48(AX)
	MOVQ R9, 56(AX)
	MOVQ R10, 64(AX)
	MOVQ R11, 72(AX)
	MOVQ R12, 80(AX)
	MOVQ R13, 88(AX)
	RET

noAdx_3:
	MOVQ res+0(FP), AX
	MOVQ AX, (SP)
	CALL ·_fromMontGeneric(SB)
	RET
