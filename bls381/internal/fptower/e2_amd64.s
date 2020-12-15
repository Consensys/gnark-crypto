
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

TEXT ·addE2(SB), NOSPLIT, $0-24
    MOVQ x+8(FP), AX
    MOVQ 0(AX), BX
    MOVQ 8(AX), BP
    MOVQ 16(AX), SI
    MOVQ 24(AX), DI
    MOVQ 32(AX), R8
    MOVQ 40(AX), R9
    MOVQ y+16(FP), DX
    ADDQ 0(DX), BX
    ADCQ 8(DX), BP
    ADCQ 16(DX), SI
    ADCQ 24(DX), DI
    ADCQ 32(DX), R8
    ADCQ 40(DX), R9
    MOVQ res+0(FP), CX
    MOVQ BX, R10
    MOVQ BP, R11
    MOVQ SI, R12
    MOVQ DI, R13
    MOVQ R8, R14
    MOVQ R9, R15
    SUBQ ·qE2+0(SB), R10
    SBBQ ·qE2+8(SB), R11
    SBBQ ·qE2+16(SB), R12
    SBBQ ·qE2+24(SB), R13
    SBBQ ·qE2+32(SB), R14
    SBBQ ·qE2+40(SB), R15
    CMOVQCC R10, BX
    CMOVQCC R11, BP
    CMOVQCC R12, SI
    CMOVQCC R13, DI
    CMOVQCC R14, R8
    CMOVQCC R15, R9
    MOVQ BX, 0(CX)
    MOVQ BP, 8(CX)
    MOVQ SI, 16(CX)
    MOVQ DI, 24(CX)
    MOVQ R8, 32(CX)
    MOVQ R9, 40(CX)
    MOVQ 48(AX), BX
    MOVQ 56(AX), BP
    MOVQ 64(AX), SI
    MOVQ 72(AX), DI
    MOVQ 80(AX), R8
    MOVQ 88(AX), R9
    ADDQ 48(DX), BX
    ADCQ 56(DX), BP
    ADCQ 64(DX), SI
    ADCQ 72(DX), DI
    ADCQ 80(DX), R8
    ADCQ 88(DX), R9
    MOVQ BX, R10
    MOVQ BP, R11
    MOVQ SI, R12
    MOVQ DI, R13
    MOVQ R8, R14
    MOVQ R9, R15
    SUBQ ·qE2+0(SB), R10
    SBBQ ·qE2+8(SB), R11
    SBBQ ·qE2+16(SB), R12
    SBBQ ·qE2+24(SB), R13
    SBBQ ·qE2+32(SB), R14
    SBBQ ·qE2+40(SB), R15
    CMOVQCC R10, BX
    CMOVQCC R11, BP
    CMOVQCC R12, SI
    CMOVQCC R13, DI
    CMOVQCC R14, R8
    CMOVQCC R15, R9
    MOVQ BX, 48(CX)
    MOVQ BP, 56(CX)
    MOVQ SI, 64(CX)
    MOVQ DI, 72(CX)
    MOVQ R8, 80(CX)
    MOVQ R9, 88(CX)
    RET

TEXT ·doubleE2(SB), NOSPLIT, $0-16
    MOVQ res+0(FP), DX
    MOVQ x+8(FP), AX
    MOVQ 0(AX), CX
    MOVQ 8(AX), BX
    MOVQ 16(AX), BP
    MOVQ 24(AX), SI
    MOVQ 32(AX), DI
    MOVQ 40(AX), R8
    ADDQ CX, CX
    ADCQ BX, BX
    ADCQ BP, BP
    ADCQ SI, SI
    ADCQ DI, DI
    ADCQ R8, R8
    MOVQ CX, R9
    MOVQ BX, R10
    MOVQ BP, R11
    MOVQ SI, R12
    MOVQ DI, R13
    MOVQ R8, R14
    SUBQ ·qE2+0(SB), R9
    SBBQ ·qE2+8(SB), R10
    SBBQ ·qE2+16(SB), R11
    SBBQ ·qE2+24(SB), R12
    SBBQ ·qE2+32(SB), R13
    SBBQ ·qE2+40(SB), R14
    CMOVQCC R9, CX
    CMOVQCC R10, BX
    CMOVQCC R11, BP
    CMOVQCC R12, SI
    CMOVQCC R13, DI
    CMOVQCC R14, R8
    MOVQ CX, 0(DX)
    MOVQ BX, 8(DX)
    MOVQ BP, 16(DX)
    MOVQ SI, 24(DX)
    MOVQ DI, 32(DX)
    MOVQ R8, 40(DX)
    MOVQ 48(AX), CX
    MOVQ 56(AX), BX
    MOVQ 64(AX), BP
    MOVQ 72(AX), SI
    MOVQ 80(AX), DI
    MOVQ 88(AX), R8
    ADDQ CX, CX
    ADCQ BX, BX
    ADCQ BP, BP
    ADCQ SI, SI
    ADCQ DI, DI
    ADCQ R8, R8
    MOVQ CX, R15
    MOVQ BX, R9
    MOVQ BP, R10
    MOVQ SI, R11
    MOVQ DI, R12
    MOVQ R8, R13
    SUBQ ·qE2+0(SB), R15
    SBBQ ·qE2+8(SB), R9
    SBBQ ·qE2+16(SB), R10
    SBBQ ·qE2+24(SB), R11
    SBBQ ·qE2+32(SB), R12
    SBBQ ·qE2+40(SB), R13
    CMOVQCC R15, CX
    CMOVQCC R9, BX
    CMOVQCC R10, BP
    CMOVQCC R11, SI
    CMOVQCC R12, DI
    CMOVQCC R13, R8
    MOVQ CX, 48(DX)
    MOVQ BX, 56(DX)
    MOVQ BP, 64(DX)
    MOVQ SI, 72(DX)
    MOVQ DI, 80(DX)
    MOVQ R8, 88(DX)
    RET

TEXT ·subE2(SB), NOSPLIT, $0-24
    MOVQ x+8(FP), DI
    MOVQ y+16(FP), R8
    MOVQ 0(DI), AX
    MOVQ 8(DI), DX
    MOVQ 16(DI), CX
    MOVQ 24(DI), BX
    MOVQ 32(DI), BP
    MOVQ 40(DI), SI
    SUBQ 0(R8), AX
    SBBQ 8(R8), DX
    SBBQ 16(R8), CX
    SBBQ 24(R8), BX
    SBBQ 32(R8), BP
    SBBQ 40(R8), SI
    MOVQ $0xb9feffffffffaaab, R9
    MOVQ $0x1eabfffeb153ffff, R10
    MOVQ $0x6730d2a0f6b0f624, R11
    MOVQ $0x64774b84f38512bf, R12
    MOVQ $0x4b1ba7b6434bacd7, R13
    MOVQ $0x1a0111ea397fe69a, R14
    MOVQ $0x0000000000000000, R15
    CMOVQCC R15, R9
    CMOVQCC R15, R10
    CMOVQCC R15, R11
    CMOVQCC R15, R12
    CMOVQCC R15, R13
    CMOVQCC R15, R14
    ADDQ R9, AX
    ADCQ R10, DX
    ADCQ R11, CX
    ADCQ R12, BX
    ADCQ R13, BP
    ADCQ R14, SI
    MOVQ res+0(FP), R15
    MOVQ AX, 0(R15)
    MOVQ DX, 8(R15)
    MOVQ CX, 16(R15)
    MOVQ BX, 24(R15)
    MOVQ BP, 32(R15)
    MOVQ SI, 40(R15)
    MOVQ 48(DI), AX
    MOVQ 56(DI), DX
    MOVQ 64(DI), CX
    MOVQ 72(DI), BX
    MOVQ 80(DI), BP
    MOVQ 88(DI), SI
    SUBQ 48(R8), AX
    SBBQ 56(R8), DX
    SBBQ 64(R8), CX
    SBBQ 72(R8), BX
    SBBQ 80(R8), BP
    SBBQ 88(R8), SI
    MOVQ $0xb9feffffffffaaab, R9
    MOVQ $0x1eabfffeb153ffff, R10
    MOVQ $0x6730d2a0f6b0f624, R11
    MOVQ $0x64774b84f38512bf, R12
    MOVQ $0x4b1ba7b6434bacd7, R13
    MOVQ $0x1a0111ea397fe69a, R14
    MOVQ $0x0000000000000000, R15
    CMOVQCC R15, R9
    CMOVQCC R15, R10
    CMOVQCC R15, R11
    CMOVQCC R15, R12
    CMOVQCC R15, R13
    CMOVQCC R15, R14
    ADDQ R9, AX
    ADCQ R10, DX
    ADCQ R11, CX
    ADCQ R12, BX
    ADCQ R13, BP
    ADCQ R14, SI
    MOVQ res+0(FP), DI
    MOVQ AX, 48(DI)
    MOVQ DX, 56(DI)
    MOVQ CX, 64(DI)
    MOVQ BX, 72(DI)
    MOVQ BP, 80(DI)
    MOVQ SI, 88(DI)
    RET

TEXT ·negE2(SB), NOSPLIT, $0-16
    MOVQ res+0(FP), DX
    MOVQ x+8(FP), AX
    MOVQ 0(AX), BX
    MOVQ 8(AX), BP
    MOVQ 16(AX), SI
    MOVQ 24(AX), DI
    MOVQ 32(AX), R8
    MOVQ 40(AX), R9
    MOVQ BX, AX
    ORQ BP, AX
    ORQ SI, AX
    ORQ DI, AX
    ORQ R8, AX
    ORQ R9, AX
    TESTQ AX, AX
    JNE l1
    MOVQ AX, 48(DX)
    MOVQ AX, 56(DX)
    MOVQ AX, 64(DX)
    MOVQ AX, 72(DX)
    MOVQ AX, 80(DX)
    MOVQ AX, 88(DX)
    JMP l3
l1:
    MOVQ $0xb9feffffffffaaab, CX
    SUBQ BX, CX
    MOVQ CX, 0(DX)
    MOVQ $0x1eabfffeb153ffff, CX
    SBBQ BP, CX
    MOVQ CX, 8(DX)
    MOVQ $0x6730d2a0f6b0f624, CX
    SBBQ SI, CX
    MOVQ CX, 16(DX)
    MOVQ $0x64774b84f38512bf, CX
    SBBQ DI, CX
    MOVQ CX, 24(DX)
    MOVQ $0x4b1ba7b6434bacd7, CX
    SBBQ R8, CX
    MOVQ CX, 32(DX)
    MOVQ $0x1a0111ea397fe69a, CX
    SBBQ R9, CX
    MOVQ CX, 40(DX)
l3:
    MOVQ x+8(FP), AX
    MOVQ 48(AX), BX
    MOVQ 56(AX), BP
    MOVQ 64(AX), SI
    MOVQ 72(AX), DI
    MOVQ 80(AX), R8
    MOVQ 88(AX), R9
    MOVQ BX, AX
    ORQ BP, AX
    ORQ SI, AX
    ORQ DI, AX
    ORQ R8, AX
    ORQ R9, AX
    TESTQ AX, AX
    JNE l2
    MOVQ AX, 48(DX)
    MOVQ AX, 56(DX)
    MOVQ AX, 64(DX)
    MOVQ AX, 72(DX)
    MOVQ AX, 80(DX)
    MOVQ AX, 88(DX)
    RET
l2:
    MOVQ $0xb9feffffffffaaab, CX
    SUBQ BX, CX
    MOVQ CX, 48(DX)
    MOVQ $0x1eabfffeb153ffff, CX
    SBBQ BP, CX
    MOVQ CX, 56(DX)
    MOVQ $0x6730d2a0f6b0f624, CX
    SBBQ SI, CX
    MOVQ CX, 64(DX)
    MOVQ $0x64774b84f38512bf, CX
    SBBQ DI, CX
    MOVQ CX, 72(DX)
    MOVQ $0x4b1ba7b6434bacd7, CX
    SBBQ R8, CX
    MOVQ CX, 80(DX)
    MOVQ $0x1a0111ea397fe69a, CX
    SBBQ R9, CX
    MOVQ CX, 88(DX)
    RET

TEXT ·mulNonResE2(SB), NOSPLIT, $0-16
    MOVQ x+8(FP), DI
    MOVQ 0(DI), AX
    MOVQ 8(DI), DX
    MOVQ 16(DI), CX
    MOVQ 24(DI), BX
    MOVQ 32(DI), BP
    MOVQ 40(DI), SI
    SUBQ 48(DI), AX
    SBBQ 56(DI), DX
    SBBQ 64(DI), CX
    SBBQ 72(DI), BX
    SBBQ 80(DI), BP
    SBBQ 88(DI), SI
    MOVQ $0xb9feffffffffaaab, R8
    MOVQ $0x1eabfffeb153ffff, R9
    MOVQ $0x6730d2a0f6b0f624, R10
    MOVQ $0x64774b84f38512bf, R11
    MOVQ $0x4b1ba7b6434bacd7, R12
    MOVQ $0x1a0111ea397fe69a, R13
    MOVQ $0x0000000000000000, R14
    CMOVQCC R14, R8
    CMOVQCC R14, R9
    CMOVQCC R14, R10
    CMOVQCC R14, R11
    CMOVQCC R14, R12
    CMOVQCC R14, R13
    ADDQ R8, AX
    ADCQ R9, DX
    ADCQ R10, CX
    ADCQ R11, BX
    ADCQ R12, BP
    ADCQ R13, SI
    MOVQ 48(DI), R15
    MOVQ 56(DI), R14
    MOVQ 64(DI), R8
    MOVQ 72(DI), R9
    MOVQ 80(DI), R10
    MOVQ 88(DI), R11
    ADDQ 0(DI), R15
    ADCQ 8(DI), R14
    ADCQ 16(DI), R8
    ADCQ 24(DI), R9
    ADCQ 32(DI), R10
    ADCQ 40(DI), R11
    MOVQ res+0(FP), DI
    MOVQ AX, 0(DI)
    MOVQ DX, 8(DI)
    MOVQ CX, 16(DI)
    MOVQ BX, 24(DI)
    MOVQ BP, 32(DI)
    MOVQ SI, 40(DI)
    MOVQ R15, R12
    MOVQ R14, R13
    MOVQ R8, AX
    MOVQ R9, DX
    MOVQ R10, CX
    MOVQ R11, BX
    SUBQ ·qE2+0(SB), R12
    SBBQ ·qE2+8(SB), R13
    SBBQ ·qE2+16(SB), AX
    SBBQ ·qE2+24(SB), DX
    SBBQ ·qE2+32(SB), CX
    SBBQ ·qE2+40(SB), BX
    CMOVQCC R12, R15
    CMOVQCC R13, R14
    CMOVQCC AX, R8
    CMOVQCC DX, R9
    CMOVQCC CX, R10
    CMOVQCC BX, R11
    MOVQ R15, 48(DI)
    MOVQ R14, 56(DI)
    MOVQ R8, 64(DI)
    MOVQ R9, 72(DI)
    MOVQ R10, 80(DI)
    MOVQ R11, 88(DI)
    RET

TEXT ·squareAdxE2(SB), $56-16
NO_LOCAL_POINTERS
    CMPB ·supportAdx(SB), $0x0000000000000001
    JNE l4
    MOVQ x+8(FP), DX
    MOVQ 0(DX), R14
    MOVQ 8(DX), R15
    MOVQ 16(DX), CX
    MOVQ 24(DX), BX
    MOVQ 32(DX), BP
    MOVQ 40(DX), SI
    SUBQ 48(DX), R14
    SBBQ 56(DX), R15
    SBBQ 64(DX), CX
    SBBQ 72(DX), BX
    SBBQ 80(DX), BP
    SBBQ 88(DX), SI
    MOVQ $0xb9feffffffffaaab, DI
    MOVQ $0x1eabfffeb153ffff, R8
    MOVQ $0x6730d2a0f6b0f624, R9
    MOVQ $0x64774b84f38512bf, R10
    MOVQ $0x4b1ba7b6434bacd7, R11
    MOVQ $0x1a0111ea397fe69a, R12
    MOVQ $0x0000000000000000, R13
    CMOVQCC R13, DI
    CMOVQCC R13, R8
    CMOVQCC R13, R9
    CMOVQCC R13, R10
    CMOVQCC R13, R11
    CMOVQCC R13, R12
    ADDQ DI, R14
    ADCQ R8, R15
    ADCQ R9, CX
    ADCQ R10, BX
    ADCQ R11, BP
    ADCQ R12, SI
    MOVQ R14, -16(SP)
    MOVQ R15, -24(SP)
    MOVQ CX, -32(SP)
    MOVQ BX, -40(SP)
    MOVQ BP, -48(SP)
    MOVQ SI, -56(SP)
    MOVQ 0(DX), R14
    MOVQ 8(DX), R15
    MOVQ 16(DX), CX
    MOVQ 24(DX), BX
    MOVQ 32(DX), BP
    MOVQ 40(DX), SI
    MOVQ 48(DX), R13
    MOVQ 56(DX), DI
    MOVQ 64(DX), R8
    MOVQ 72(DX), R9
    MOVQ 80(DX), R10
    MOVQ 88(DX), R11
    ADDQ R13, R14
    ADCQ DI, R15
    ADCQ R8, CX
    ADCQ R9, BX
    ADCQ R10, BP
    ADCQ R11, SI
    MOVQ R14, R12
    MOVQ R15, R13
    MOVQ CX, DI
    MOVQ BX, R8
    MOVQ BP, R9
    MOVQ SI, R10
    SUBQ ·qE2+0(SB), R12
    SBBQ ·qE2+8(SB), R13
    SBBQ ·qE2+16(SB), DI
    SBBQ ·qE2+24(SB), R8
    SBBQ ·qE2+32(SB), R9
    SBBQ ·qE2+40(SB), R10
    CMOVQCC R12, R14
    CMOVQCC R13, R15
    CMOVQCC DI, CX
    CMOVQCC R8, BX
    CMOVQCC R9, BP
    CMOVQCC R10, SI
    XORQ DX, DX
    MOVQ -16(SP), DX
    MULXQ R14, R11, R12
    MULXQ R15, AX, R13
    ADOXQ AX, R12
    MULXQ CX, AX, DI
    ADOXQ AX, R13
    MULXQ BX, AX, R8
    ADOXQ AX, DI
    MULXQ BP, AX, R9
    ADOXQ AX, R8
    MULXQ SI, AX, R10
    ADOXQ AX, R9
    // add the last carries to R10
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R10
    ADOXQ DX, R10
    MOVQ R11, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    PUSHQ R10
    MULXQ ·qE2+0(SB), AX, R10
    ADCXQ R11, AX
    MOVQ R10, R11
    POPQ R10
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R12, R11
    MULXQ ·qE2+8(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qE2+16(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ DI, R13
    MULXQ ·qE2+24(SB), AX, DI
    ADOXQ AX, R13
    ADCXQ R8, DI
    MULXQ ·qE2+32(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qE2+40(SB), AX, R9
    ADOXQ AX, R8
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R9
    ADOXQ R10, R9
    XORQ DX, DX
    MOVQ -24(SP), DX
    MULXQ R14, AX, R10
    ADOXQ AX, R11
    ADCXQ R10, R12
    MULXQ R15, AX, R10
    ADOXQ AX, R12
    ADCXQ R10, R13
    MULXQ CX, AX, R10
    ADOXQ AX, R13
    ADCXQ R10, DI
    MULXQ BX, AX, R10
    ADOXQ AX, DI
    ADCXQ R10, R8
    MULXQ BP, AX, R10
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ SI, AX, R10
    ADOXQ AX, R9
    // add the last carries to R10
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R10
    ADOXQ DX, R10
    MOVQ R11, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    PUSHQ R10
    MULXQ ·qE2+0(SB), AX, R10
    ADCXQ R11, AX
    MOVQ R10, R11
    POPQ R10
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R12, R11
    MULXQ ·qE2+8(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qE2+16(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ DI, R13
    MULXQ ·qE2+24(SB), AX, DI
    ADOXQ AX, R13
    ADCXQ R8, DI
    MULXQ ·qE2+32(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qE2+40(SB), AX, R9
    ADOXQ AX, R8
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R9
    ADOXQ R10, R9
    XORQ DX, DX
    MOVQ -32(SP), DX
    MULXQ R14, AX, R10
    ADOXQ AX, R11
    ADCXQ R10, R12
    MULXQ R15, AX, R10
    ADOXQ AX, R12
    ADCXQ R10, R13
    MULXQ CX, AX, R10
    ADOXQ AX, R13
    ADCXQ R10, DI
    MULXQ BX, AX, R10
    ADOXQ AX, DI
    ADCXQ R10, R8
    MULXQ BP, AX, R10
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ SI, AX, R10
    ADOXQ AX, R9
    // add the last carries to R10
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R10
    ADOXQ DX, R10
    MOVQ R11, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    PUSHQ R10
    MULXQ ·qE2+0(SB), AX, R10
    ADCXQ R11, AX
    MOVQ R10, R11
    POPQ R10
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R12, R11
    MULXQ ·qE2+8(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qE2+16(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ DI, R13
    MULXQ ·qE2+24(SB), AX, DI
    ADOXQ AX, R13
    ADCXQ R8, DI
    MULXQ ·qE2+32(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qE2+40(SB), AX, R9
    ADOXQ AX, R8
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R9
    ADOXQ R10, R9
    XORQ DX, DX
    MOVQ -40(SP), DX
    MULXQ R14, AX, R10
    ADOXQ AX, R11
    ADCXQ R10, R12
    MULXQ R15, AX, R10
    ADOXQ AX, R12
    ADCXQ R10, R13
    MULXQ CX, AX, R10
    ADOXQ AX, R13
    ADCXQ R10, DI
    MULXQ BX, AX, R10
    ADOXQ AX, DI
    ADCXQ R10, R8
    MULXQ BP, AX, R10
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ SI, AX, R10
    ADOXQ AX, R9
    // add the last carries to R10
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R10
    ADOXQ DX, R10
    MOVQ R11, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    PUSHQ R10
    MULXQ ·qE2+0(SB), AX, R10
    ADCXQ R11, AX
    MOVQ R10, R11
    POPQ R10
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R12, R11
    MULXQ ·qE2+8(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qE2+16(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ DI, R13
    MULXQ ·qE2+24(SB), AX, DI
    ADOXQ AX, R13
    ADCXQ R8, DI
    MULXQ ·qE2+32(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qE2+40(SB), AX, R9
    ADOXQ AX, R8
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R9
    ADOXQ R10, R9
    XORQ DX, DX
    MOVQ -48(SP), DX
    MULXQ R14, AX, R10
    ADOXQ AX, R11
    ADCXQ R10, R12
    MULXQ R15, AX, R10
    ADOXQ AX, R12
    ADCXQ R10, R13
    MULXQ CX, AX, R10
    ADOXQ AX, R13
    ADCXQ R10, DI
    MULXQ BX, AX, R10
    ADOXQ AX, DI
    ADCXQ R10, R8
    MULXQ BP, AX, R10
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ SI, AX, R10
    ADOXQ AX, R9
    // add the last carries to R10
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R10
    ADOXQ DX, R10
    MOVQ R11, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    PUSHQ R10
    MULXQ ·qE2+0(SB), AX, R10
    ADCXQ R11, AX
    MOVQ R10, R11
    POPQ R10
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R12, R11
    MULXQ ·qE2+8(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qE2+16(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ DI, R13
    MULXQ ·qE2+24(SB), AX, DI
    ADOXQ AX, R13
    ADCXQ R8, DI
    MULXQ ·qE2+32(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qE2+40(SB), AX, R9
    ADOXQ AX, R8
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R9
    ADOXQ R10, R9
    XORQ DX, DX
    MOVQ -56(SP), DX
    MULXQ R14, AX, R10
    ADOXQ AX, R11
    ADCXQ R10, R12
    MULXQ R15, AX, R10
    ADOXQ AX, R12
    ADCXQ R10, R13
    MULXQ CX, AX, R10
    ADOXQ AX, R13
    ADCXQ R10, DI
    MULXQ BX, AX, R10
    ADOXQ AX, DI
    ADCXQ R10, R8
    MULXQ BP, AX, R10
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ SI, AX, R10
    ADOXQ AX, R9
    // add the last carries to R10
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R10
    ADOXQ DX, R10
    MOVQ R11, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    PUSHQ R10
    MULXQ ·qE2+0(SB), AX, R10
    ADCXQ R11, AX
    MOVQ R10, R11
    POPQ R10
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R12, R11
    MULXQ ·qE2+8(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qE2+16(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ DI, R13
    MULXQ ·qE2+24(SB), AX, DI
    ADOXQ AX, R13
    ADCXQ R8, DI
    MULXQ ·qE2+32(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qE2+40(SB), AX, R9
    ADOXQ AX, R8
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R9
    ADOXQ R10, R9
    MOVQ R11, R10
    MOVQ R12, R14
    MOVQ R13, R15
    MOVQ DI, CX
    MOVQ R8, BX
    MOVQ R9, BP
    SUBQ ·qE2+0(SB), R10
    SBBQ ·qE2+8(SB), R14
    SBBQ ·qE2+16(SB), R15
    SBBQ ·qE2+24(SB), CX
    SBBQ ·qE2+32(SB), BX
    SBBQ ·qE2+40(SB), BP
    CMOVQCC R10, R11
    CMOVQCC R14, R12
    CMOVQCC R15, R13
    CMOVQCC CX, DI
    CMOVQCC BX, R8
    CMOVQCC BP, R9
    MOVQ res+0(FP), DX
    MOVQ x+8(FP), SI
    MOVQ 0(SI), R10
    MOVQ 8(SI), R14
    MOVQ 16(SI), R15
    MOVQ 24(SI), CX
    MOVQ 32(SI), BX
    MOVQ 40(SI), BP
    MOVQ R11, 0(DX)
    MOVQ R12, 8(DX)
    MOVQ R13, 16(DX)
    MOVQ DI, 24(DX)
    MOVQ R8, 32(DX)
    MOVQ R9, 40(DX)
    XORQ DX, DX
    MOVQ x+8(FP), DX
    MOVQ 48(DX), DX
    MULXQ R10, SI, R11
    MULXQ R14, AX, R12
    ADOXQ AX, R11
    MULXQ R15, AX, R13
    ADOXQ AX, R12
    MULXQ CX, AX, DI
    ADOXQ AX, R13
    MULXQ BX, AX, R8
    ADOXQ AX, DI
    MULXQ BP, AX, R9
    ADOXQ AX, R8
    // add the last carries to R9
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R9
    ADOXQ DX, R9
    MOVQ SI, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    PUSHQ R9
    MULXQ ·qE2+0(SB), AX, R9
    ADCXQ SI, AX
    MOVQ R9, SI
    POPQ R9
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R11, SI
    MULXQ ·qE2+8(SB), AX, R11
    ADOXQ AX, SI
    ADCXQ R12, R11
    MULXQ ·qE2+16(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qE2+24(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ DI, R13
    MULXQ ·qE2+32(SB), AX, DI
    ADOXQ AX, R13
    ADCXQ R8, DI
    MULXQ ·qE2+40(SB), AX, R8
    ADOXQ AX, DI
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R8
    ADOXQ R9, R8
    XORQ DX, DX
    MOVQ x+8(FP), DX
    MOVQ 56(DX), DX
    MULXQ R10, AX, R9
    ADOXQ AX, SI
    ADCXQ R9, R11
    MULXQ R14, AX, R9
    ADOXQ AX, R11
    ADCXQ R9, R12
    MULXQ R15, AX, R9
    ADOXQ AX, R12
    ADCXQ R9, R13
    MULXQ CX, AX, R9
    ADOXQ AX, R13
    ADCXQ R9, DI
    MULXQ BX, AX, R9
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ BP, AX, R9
    ADOXQ AX, R8
    // add the last carries to R9
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R9
    ADOXQ DX, R9
    MOVQ SI, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    PUSHQ R9
    MULXQ ·qE2+0(SB), AX, R9
    ADCXQ SI, AX
    MOVQ R9, SI
    POPQ R9
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R11, SI
    MULXQ ·qE2+8(SB), AX, R11
    ADOXQ AX, SI
    ADCXQ R12, R11
    MULXQ ·qE2+16(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qE2+24(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ DI, R13
    MULXQ ·qE2+32(SB), AX, DI
    ADOXQ AX, R13
    ADCXQ R8, DI
    MULXQ ·qE2+40(SB), AX, R8
    ADOXQ AX, DI
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R8
    ADOXQ R9, R8
    XORQ DX, DX
    MOVQ x+8(FP), DX
    MOVQ 64(DX), DX
    MULXQ R10, AX, R9
    ADOXQ AX, SI
    ADCXQ R9, R11
    MULXQ R14, AX, R9
    ADOXQ AX, R11
    ADCXQ R9, R12
    MULXQ R15, AX, R9
    ADOXQ AX, R12
    ADCXQ R9, R13
    MULXQ CX, AX, R9
    ADOXQ AX, R13
    ADCXQ R9, DI
    MULXQ BX, AX, R9
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ BP, AX, R9
    ADOXQ AX, R8
    // add the last carries to R9
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R9
    ADOXQ DX, R9
    MOVQ SI, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    PUSHQ R9
    MULXQ ·qE2+0(SB), AX, R9
    ADCXQ SI, AX
    MOVQ R9, SI
    POPQ R9
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R11, SI
    MULXQ ·qE2+8(SB), AX, R11
    ADOXQ AX, SI
    ADCXQ R12, R11
    MULXQ ·qE2+16(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qE2+24(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ DI, R13
    MULXQ ·qE2+32(SB), AX, DI
    ADOXQ AX, R13
    ADCXQ R8, DI
    MULXQ ·qE2+40(SB), AX, R8
    ADOXQ AX, DI
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R8
    ADOXQ R9, R8
    XORQ DX, DX
    MOVQ x+8(FP), DX
    MOVQ 72(DX), DX
    MULXQ R10, AX, R9
    ADOXQ AX, SI
    ADCXQ R9, R11
    MULXQ R14, AX, R9
    ADOXQ AX, R11
    ADCXQ R9, R12
    MULXQ R15, AX, R9
    ADOXQ AX, R12
    ADCXQ R9, R13
    MULXQ CX, AX, R9
    ADOXQ AX, R13
    ADCXQ R9, DI
    MULXQ BX, AX, R9
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ BP, AX, R9
    ADOXQ AX, R8
    // add the last carries to R9
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R9
    ADOXQ DX, R9
    MOVQ SI, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    PUSHQ R9
    MULXQ ·qE2+0(SB), AX, R9
    ADCXQ SI, AX
    MOVQ R9, SI
    POPQ R9
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R11, SI
    MULXQ ·qE2+8(SB), AX, R11
    ADOXQ AX, SI
    ADCXQ R12, R11
    MULXQ ·qE2+16(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qE2+24(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ DI, R13
    MULXQ ·qE2+32(SB), AX, DI
    ADOXQ AX, R13
    ADCXQ R8, DI
    MULXQ ·qE2+40(SB), AX, R8
    ADOXQ AX, DI
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R8
    ADOXQ R9, R8
    XORQ DX, DX
    MOVQ x+8(FP), DX
    MOVQ 80(DX), DX
    MULXQ R10, AX, R9
    ADOXQ AX, SI
    ADCXQ R9, R11
    MULXQ R14, AX, R9
    ADOXQ AX, R11
    ADCXQ R9, R12
    MULXQ R15, AX, R9
    ADOXQ AX, R12
    ADCXQ R9, R13
    MULXQ CX, AX, R9
    ADOXQ AX, R13
    ADCXQ R9, DI
    MULXQ BX, AX, R9
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ BP, AX, R9
    ADOXQ AX, R8
    // add the last carries to R9
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R9
    ADOXQ DX, R9
    MOVQ SI, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    PUSHQ R9
    MULXQ ·qE2+0(SB), AX, R9
    ADCXQ SI, AX
    MOVQ R9, SI
    POPQ R9
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R11, SI
    MULXQ ·qE2+8(SB), AX, R11
    ADOXQ AX, SI
    ADCXQ R12, R11
    MULXQ ·qE2+16(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qE2+24(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ DI, R13
    MULXQ ·qE2+32(SB), AX, DI
    ADOXQ AX, R13
    ADCXQ R8, DI
    MULXQ ·qE2+40(SB), AX, R8
    ADOXQ AX, DI
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R8
    ADOXQ R9, R8
    XORQ DX, DX
    MOVQ x+8(FP), DX
    MOVQ 88(DX), DX
    MULXQ R10, AX, R9
    ADOXQ AX, SI
    ADCXQ R9, R11
    MULXQ R14, AX, R9
    ADOXQ AX, R11
    ADCXQ R9, R12
    MULXQ R15, AX, R9
    ADOXQ AX, R12
    ADCXQ R9, R13
    MULXQ CX, AX, R9
    ADOXQ AX, R13
    ADCXQ R9, DI
    MULXQ BX, AX, R9
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ BP, AX, R9
    ADOXQ AX, R8
    // add the last carries to R9
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R9
    ADOXQ DX, R9
    MOVQ SI, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    PUSHQ R9
    MULXQ ·qE2+0(SB), AX, R9
    ADCXQ SI, AX
    MOVQ R9, SI
    POPQ R9
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R11, SI
    MULXQ ·qE2+8(SB), AX, R11
    ADOXQ AX, SI
    ADCXQ R12, R11
    MULXQ ·qE2+16(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qE2+24(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ DI, R13
    MULXQ ·qE2+32(SB), AX, DI
    ADOXQ AX, R13
    ADCXQ R8, DI
    MULXQ ·qE2+40(SB), AX, R8
    ADOXQ AX, DI
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R8
    ADOXQ R9, R8
    MOVQ SI, R9
    MOVQ R11, R10
    MOVQ R12, R14
    MOVQ R13, R15
    MOVQ DI, CX
    MOVQ R8, BX
    SUBQ ·qE2+0(SB), R9
    SBBQ ·qE2+8(SB), R10
    SBBQ ·qE2+16(SB), R14
    SBBQ ·qE2+24(SB), R15
    SBBQ ·qE2+32(SB), CX
    SBBQ ·qE2+40(SB), BX
    CMOVQCC R9, SI
    CMOVQCC R10, R11
    CMOVQCC R14, R12
    CMOVQCC R15, R13
    CMOVQCC CX, DI
    CMOVQCC BX, R8
    ADDQ SI, SI
    ADCQ R11, R11
    ADCQ R12, R12
    ADCQ R13, R13
    ADCQ DI, DI
    ADCQ R8, R8
    MOVQ res+0(FP), DX
    MOVQ SI, BP
    MOVQ R11, R9
    MOVQ R12, R10
    MOVQ R13, R14
    MOVQ DI, R15
    MOVQ R8, CX
    SUBQ ·qE2+0(SB), BP
    SBBQ ·qE2+8(SB), R9
    SBBQ ·qE2+16(SB), R10
    SBBQ ·qE2+24(SB), R14
    SBBQ ·qE2+32(SB), R15
    SBBQ ·qE2+40(SB), CX
    CMOVQCC BP, SI
    CMOVQCC R9, R11
    CMOVQCC R10, R12
    CMOVQCC R14, R13
    CMOVQCC R15, DI
    CMOVQCC CX, R8
    MOVQ SI, 48(DX)
    MOVQ R11, 56(DX)
    MOVQ R12, 64(DX)
    MOVQ R13, 72(DX)
    MOVQ DI, 80(DX)
    MOVQ R8, 88(DX)
    RET
l4:
    MOVQ res+0(FP), AX
    MOVQ AX, (SP)
    MOVQ x+8(FP), AX
    MOVQ AX, 8(SP)
CALL ·squareGenericE2(SB)
    RET

TEXT ·mulAdxE2(SB), $152-24
NO_LOCAL_POINTERS
    CMPB ·supportAdx(SB), $0x0000000000000001
    JNE l5
    MOVQ x+8(FP), AX
    MOVQ 0(AX), R14
    MOVQ 8(AX), R15
    MOVQ 16(AX), CX
    MOVQ 24(AX), BX
    MOVQ 32(AX), BP
    MOVQ 40(AX), SI
    XORQ DX, DX
    MOVQ y+16(FP), DX
    MOVQ 0(DX), DX
    MULXQ R14, DI, R8
    MULXQ R15, AX, R9
    ADOXQ AX, R8
    MULXQ CX, AX, R10
    ADOXQ AX, R9
    MULXQ BX, AX, R11
    ADOXQ AX, R10
    MULXQ BP, AX, R12
    ADOXQ AX, R11
    MULXQ SI, AX, R13
    ADOXQ AX, R12
    // add the last carries to R13
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R13
    ADOXQ DX, R13
    MOVQ DI, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    PUSHQ R13
    MULXQ ·qE2+0(SB), AX, R13
    ADCXQ DI, AX
    MOVQ R13, DI
    POPQ R13
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R8, DI
    MULXQ ·qE2+8(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qE2+16(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qE2+24(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qE2+32(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qE2+40(SB), AX, R12
    ADOXQ AX, R11
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ R13, R12
    XORQ DX, DX
    MOVQ y+16(FP), DX
    MOVQ 8(DX), DX
    MULXQ R14, AX, R13
    ADOXQ AX, DI
    ADCXQ R13, R8
    MULXQ R15, AX, R13
    ADOXQ AX, R8
    ADCXQ R13, R9
    MULXQ CX, AX, R13
    ADOXQ AX, R9
    ADCXQ R13, R10
    MULXQ BX, AX, R13
    ADOXQ AX, R10
    ADCXQ R13, R11
    MULXQ BP, AX, R13
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ SI, AX, R13
    ADOXQ AX, R12
    // add the last carries to R13
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R13
    ADOXQ DX, R13
    MOVQ DI, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    PUSHQ R13
    MULXQ ·qE2+0(SB), AX, R13
    ADCXQ DI, AX
    MOVQ R13, DI
    POPQ R13
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R8, DI
    MULXQ ·qE2+8(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qE2+16(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qE2+24(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qE2+32(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qE2+40(SB), AX, R12
    ADOXQ AX, R11
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ R13, R12
    XORQ DX, DX
    MOVQ y+16(FP), DX
    MOVQ 16(DX), DX
    MULXQ R14, AX, R13
    ADOXQ AX, DI
    ADCXQ R13, R8
    MULXQ R15, AX, R13
    ADOXQ AX, R8
    ADCXQ R13, R9
    MULXQ CX, AX, R13
    ADOXQ AX, R9
    ADCXQ R13, R10
    MULXQ BX, AX, R13
    ADOXQ AX, R10
    ADCXQ R13, R11
    MULXQ BP, AX, R13
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ SI, AX, R13
    ADOXQ AX, R12
    // add the last carries to R13
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R13
    ADOXQ DX, R13
    MOVQ DI, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    PUSHQ R13
    MULXQ ·qE2+0(SB), AX, R13
    ADCXQ DI, AX
    MOVQ R13, DI
    POPQ R13
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R8, DI
    MULXQ ·qE2+8(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qE2+16(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qE2+24(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qE2+32(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qE2+40(SB), AX, R12
    ADOXQ AX, R11
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ R13, R12
    XORQ DX, DX
    MOVQ y+16(FP), DX
    MOVQ 24(DX), DX
    MULXQ R14, AX, R13
    ADOXQ AX, DI
    ADCXQ R13, R8
    MULXQ R15, AX, R13
    ADOXQ AX, R8
    ADCXQ R13, R9
    MULXQ CX, AX, R13
    ADOXQ AX, R9
    ADCXQ R13, R10
    MULXQ BX, AX, R13
    ADOXQ AX, R10
    ADCXQ R13, R11
    MULXQ BP, AX, R13
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ SI, AX, R13
    ADOXQ AX, R12
    // add the last carries to R13
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R13
    ADOXQ DX, R13
    MOVQ DI, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    PUSHQ R13
    MULXQ ·qE2+0(SB), AX, R13
    ADCXQ DI, AX
    MOVQ R13, DI
    POPQ R13
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R8, DI
    MULXQ ·qE2+8(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qE2+16(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qE2+24(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qE2+32(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qE2+40(SB), AX, R12
    ADOXQ AX, R11
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ R13, R12
    XORQ DX, DX
    MOVQ y+16(FP), DX
    MOVQ 32(DX), DX
    MULXQ R14, AX, R13
    ADOXQ AX, DI
    ADCXQ R13, R8
    MULXQ R15, AX, R13
    ADOXQ AX, R8
    ADCXQ R13, R9
    MULXQ CX, AX, R13
    ADOXQ AX, R9
    ADCXQ R13, R10
    MULXQ BX, AX, R13
    ADOXQ AX, R10
    ADCXQ R13, R11
    MULXQ BP, AX, R13
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ SI, AX, R13
    ADOXQ AX, R12
    // add the last carries to R13
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R13
    ADOXQ DX, R13
    MOVQ DI, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    PUSHQ R13
    MULXQ ·qE2+0(SB), AX, R13
    ADCXQ DI, AX
    MOVQ R13, DI
    POPQ R13
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R8, DI
    MULXQ ·qE2+8(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qE2+16(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qE2+24(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qE2+32(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qE2+40(SB), AX, R12
    ADOXQ AX, R11
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ R13, R12
    XORQ DX, DX
    MOVQ y+16(FP), DX
    MOVQ 40(DX), DX
    MULXQ R14, AX, R13
    ADOXQ AX, DI
    ADCXQ R13, R8
    MULXQ R15, AX, R13
    ADOXQ AX, R8
    ADCXQ R13, R9
    MULXQ CX, AX, R13
    ADOXQ AX, R9
    ADCXQ R13, R10
    MULXQ BX, AX, R13
    ADOXQ AX, R10
    ADCXQ R13, R11
    MULXQ BP, AX, R13
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ SI, AX, R13
    ADOXQ AX, R12
    // add the last carries to R13
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R13
    ADOXQ DX, R13
    MOVQ DI, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    PUSHQ R13
    MULXQ ·qE2+0(SB), AX, R13
    ADCXQ DI, AX
    MOVQ R13, DI
    POPQ R13
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R8, DI
    MULXQ ·qE2+8(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qE2+16(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qE2+24(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qE2+32(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qE2+40(SB), AX, R12
    ADOXQ AX, R11
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ R13, R12
    MOVQ DI, R13
    MOVQ R8, R14
    MOVQ R9, R15
    MOVQ R10, CX
    MOVQ R11, BX
    MOVQ R12, BP
    SUBQ ·qE2+0(SB), R13
    SBBQ ·qE2+8(SB), R14
    SBBQ ·qE2+16(SB), R15
    SBBQ ·qE2+24(SB), CX
    SBBQ ·qE2+32(SB), BX
    SBBQ ·qE2+40(SB), BP
    CMOVQCC R13, DI
    CMOVQCC R14, R8
    CMOVQCC R15, R9
    CMOVQCC CX, R10
    CMOVQCC BX, R11
    CMOVQCC BP, R12
    MOVQ DI, -16(SP)
    MOVQ R8, -24(SP)
    MOVQ R9, -32(SP)
    MOVQ R10, -40(SP)
    MOVQ R11, -48(SP)
    MOVQ R12, -56(SP)
    MOVQ x+8(FP), AX
    MOVQ 48(AX), SI
    MOVQ 56(AX), R13
    MOVQ 64(AX), R14
    MOVQ 72(AX), R15
    MOVQ 80(AX), CX
    MOVQ 88(AX), BX
    XORQ DX, DX
    MOVQ y+16(FP), DX
    MOVQ 48(DX), DX
    MULXQ SI, BP, DI
    MULXQ R13, AX, R8
    ADOXQ AX, DI
    MULXQ R14, AX, R9
    ADOXQ AX, R8
    MULXQ R15, AX, R10
    ADOXQ AX, R9
    MULXQ CX, AX, R11
    ADOXQ AX, R10
    MULXQ BX, AX, R12
    ADOXQ AX, R11
    // add the last carries to R12
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R12
    ADOXQ DX, R12
    MOVQ BP, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    PUSHQ R12
    MULXQ ·qE2+0(SB), AX, R12
    ADCXQ BP, AX
    MOVQ R12, BP
    POPQ R12
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ DI, BP
    MULXQ ·qE2+8(SB), AX, DI
    ADOXQ AX, BP
    ADCXQ R8, DI
    MULXQ ·qE2+16(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qE2+24(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qE2+32(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qE2+40(SB), AX, R11
    ADOXQ AX, R10
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R11
    ADOXQ R12, R11
    XORQ DX, DX
    MOVQ y+16(FP), DX
    MOVQ 56(DX), DX
    MULXQ SI, AX, R12
    ADOXQ AX, BP
    ADCXQ R12, DI
    MULXQ R13, AX, R12
    ADOXQ AX, DI
    ADCXQ R12, R8
    MULXQ R14, AX, R12
    ADOXQ AX, R8
    ADCXQ R12, R9
    MULXQ R15, AX, R12
    ADOXQ AX, R9
    ADCXQ R12, R10
    MULXQ CX, AX, R12
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ BX, AX, R12
    ADOXQ AX, R11
    // add the last carries to R12
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R12
    ADOXQ DX, R12
    MOVQ BP, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    PUSHQ R12
    MULXQ ·qE2+0(SB), AX, R12
    ADCXQ BP, AX
    MOVQ R12, BP
    POPQ R12
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ DI, BP
    MULXQ ·qE2+8(SB), AX, DI
    ADOXQ AX, BP
    ADCXQ R8, DI
    MULXQ ·qE2+16(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qE2+24(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qE2+32(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qE2+40(SB), AX, R11
    ADOXQ AX, R10
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R11
    ADOXQ R12, R11
    XORQ DX, DX
    MOVQ y+16(FP), DX
    MOVQ 64(DX), DX
    MULXQ SI, AX, R12
    ADOXQ AX, BP
    ADCXQ R12, DI
    MULXQ R13, AX, R12
    ADOXQ AX, DI
    ADCXQ R12, R8
    MULXQ R14, AX, R12
    ADOXQ AX, R8
    ADCXQ R12, R9
    MULXQ R15, AX, R12
    ADOXQ AX, R9
    ADCXQ R12, R10
    MULXQ CX, AX, R12
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ BX, AX, R12
    ADOXQ AX, R11
    // add the last carries to R12
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R12
    ADOXQ DX, R12
    MOVQ BP, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    PUSHQ R12
    MULXQ ·qE2+0(SB), AX, R12
    ADCXQ BP, AX
    MOVQ R12, BP
    POPQ R12
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ DI, BP
    MULXQ ·qE2+8(SB), AX, DI
    ADOXQ AX, BP
    ADCXQ R8, DI
    MULXQ ·qE2+16(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qE2+24(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qE2+32(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qE2+40(SB), AX, R11
    ADOXQ AX, R10
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R11
    ADOXQ R12, R11
    XORQ DX, DX
    MOVQ y+16(FP), DX
    MOVQ 72(DX), DX
    MULXQ SI, AX, R12
    ADOXQ AX, BP
    ADCXQ R12, DI
    MULXQ R13, AX, R12
    ADOXQ AX, DI
    ADCXQ R12, R8
    MULXQ R14, AX, R12
    ADOXQ AX, R8
    ADCXQ R12, R9
    MULXQ R15, AX, R12
    ADOXQ AX, R9
    ADCXQ R12, R10
    MULXQ CX, AX, R12
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ BX, AX, R12
    ADOXQ AX, R11
    // add the last carries to R12
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R12
    ADOXQ DX, R12
    MOVQ BP, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    PUSHQ R12
    MULXQ ·qE2+0(SB), AX, R12
    ADCXQ BP, AX
    MOVQ R12, BP
    POPQ R12
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ DI, BP
    MULXQ ·qE2+8(SB), AX, DI
    ADOXQ AX, BP
    ADCXQ R8, DI
    MULXQ ·qE2+16(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qE2+24(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qE2+32(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qE2+40(SB), AX, R11
    ADOXQ AX, R10
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R11
    ADOXQ R12, R11
    XORQ DX, DX
    MOVQ y+16(FP), DX
    MOVQ 80(DX), DX
    MULXQ SI, AX, R12
    ADOXQ AX, BP
    ADCXQ R12, DI
    MULXQ R13, AX, R12
    ADOXQ AX, DI
    ADCXQ R12, R8
    MULXQ R14, AX, R12
    ADOXQ AX, R8
    ADCXQ R12, R9
    MULXQ R15, AX, R12
    ADOXQ AX, R9
    ADCXQ R12, R10
    MULXQ CX, AX, R12
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ BX, AX, R12
    ADOXQ AX, R11
    // add the last carries to R12
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R12
    ADOXQ DX, R12
    MOVQ BP, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    PUSHQ R12
    MULXQ ·qE2+0(SB), AX, R12
    ADCXQ BP, AX
    MOVQ R12, BP
    POPQ R12
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ DI, BP
    MULXQ ·qE2+8(SB), AX, DI
    ADOXQ AX, BP
    ADCXQ R8, DI
    MULXQ ·qE2+16(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qE2+24(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qE2+32(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qE2+40(SB), AX, R11
    ADOXQ AX, R10
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R11
    ADOXQ R12, R11
    XORQ DX, DX
    MOVQ y+16(FP), DX
    MOVQ 88(DX), DX
    MULXQ SI, AX, R12
    ADOXQ AX, BP
    ADCXQ R12, DI
    MULXQ R13, AX, R12
    ADOXQ AX, DI
    ADCXQ R12, R8
    MULXQ R14, AX, R12
    ADOXQ AX, R8
    ADCXQ R12, R9
    MULXQ R15, AX, R12
    ADOXQ AX, R9
    ADCXQ R12, R10
    MULXQ CX, AX, R12
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ BX, AX, R12
    ADOXQ AX, R11
    // add the last carries to R12
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R12
    ADOXQ DX, R12
    MOVQ BP, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    PUSHQ R12
    MULXQ ·qE2+0(SB), AX, R12
    ADCXQ BP, AX
    MOVQ R12, BP
    POPQ R12
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ DI, BP
    MULXQ ·qE2+8(SB), AX, DI
    ADOXQ AX, BP
    ADCXQ R8, DI
    MULXQ ·qE2+16(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qE2+24(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qE2+32(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qE2+40(SB), AX, R11
    ADOXQ AX, R10
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R11
    ADOXQ R12, R11
    MOVQ BP, R12
    MOVQ DI, SI
    MOVQ R8, R13
    MOVQ R9, R14
    MOVQ R10, R15
    MOVQ R11, CX
    SUBQ ·qE2+0(SB), R12
    SBBQ ·qE2+8(SB), SI
    SBBQ ·qE2+16(SB), R13
    SBBQ ·qE2+24(SB), R14
    SBBQ ·qE2+32(SB), R15
    SBBQ ·qE2+40(SB), CX
    CMOVQCC R12, BP
    CMOVQCC SI, DI
    CMOVQCC R13, R8
    CMOVQCC R14, R9
    CMOVQCC R15, R10
    CMOVQCC CX, R11
    MOVQ BP, -64(SP)
    MOVQ DI, -72(SP)
    MOVQ R8, -80(SP)
    MOVQ R9, -88(SP)
    MOVQ R10, -96(SP)
    MOVQ R11, -104(SP)
    MOVQ x+8(FP), AX
    MOVQ 0(AX), BX
    MOVQ 8(AX), R12
    MOVQ 16(AX), SI
    MOVQ 24(AX), R13
    MOVQ 32(AX), R14
    MOVQ 40(AX), R15
    ADDQ 48(AX), BX
    ADCQ 56(AX), R12
    ADCQ 64(AX), SI
    ADCQ 72(AX), R13
    ADCQ 80(AX), R14
    ADCQ 88(AX), R15
    MOVQ BX, CX
    MOVQ R12, BP
    MOVQ SI, DI
    MOVQ R13, R8
    MOVQ R14, R9
    MOVQ R15, R10
    SUBQ ·qE2+0(SB), CX
    SBBQ ·qE2+8(SB), BP
    SBBQ ·qE2+16(SB), DI
    SBBQ ·qE2+24(SB), R8
    SBBQ ·qE2+32(SB), R9
    SBBQ ·qE2+40(SB), R10
    CMOVQCC CX, BX
    CMOVQCC BP, R12
    CMOVQCC DI, SI
    CMOVQCC R8, R13
    CMOVQCC R9, R14
    CMOVQCC R10, R15
    MOVQ BX, -112(SP)
    MOVQ R12, -120(SP)
    MOVQ SI, -128(SP)
    MOVQ R13, -136(SP)
    MOVQ R14, -144(SP)
    MOVQ R15, -152(SP)
    MOVQ y+16(FP), DX
    MOVQ 0(DX), BX
    MOVQ 8(DX), R12
    MOVQ 16(DX), SI
    MOVQ 24(DX), R13
    MOVQ 32(DX), R14
    MOVQ 40(DX), R15
    ADDQ 48(DX), BX
    ADCQ 56(DX), R12
    ADCQ 64(DX), SI
    ADCQ 72(DX), R13
    ADCQ 80(DX), R14
    ADCQ 88(DX), R15
    MOVQ BX, R11
    MOVQ R12, CX
    MOVQ SI, BP
    MOVQ R13, DI
    MOVQ R14, R8
    MOVQ R15, R9
    SUBQ ·qE2+0(SB), R11
    SBBQ ·qE2+8(SB), CX
    SBBQ ·qE2+16(SB), BP
    SBBQ ·qE2+24(SB), DI
    SBBQ ·qE2+32(SB), R8
    SBBQ ·qE2+40(SB), R9
    CMOVQCC R11, BX
    CMOVQCC CX, R12
    CMOVQCC BP, SI
    CMOVQCC DI, R13
    CMOVQCC R8, R14
    CMOVQCC R9, R15
    XORQ DX, DX
    MOVQ -112(SP), DX
    MULXQ BX, R10, R11
    MULXQ R12, AX, CX
    ADOXQ AX, R11
    MULXQ SI, AX, BP
    ADOXQ AX, CX
    MULXQ R13, AX, DI
    ADOXQ AX, BP
    MULXQ R14, AX, R8
    ADOXQ AX, DI
    MULXQ R15, AX, R9
    ADOXQ AX, R8
    // add the last carries to R9
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R9
    ADOXQ DX, R9
    MOVQ R10, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    PUSHQ R9
    MULXQ ·qE2+0(SB), AX, R9
    ADCXQ R10, AX
    MOVQ R9, R10
    POPQ R9
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R11, R10
    MULXQ ·qE2+8(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ CX, R11
    MULXQ ·qE2+16(SB), AX, CX
    ADOXQ AX, R11
    ADCXQ BP, CX
    MULXQ ·qE2+24(SB), AX, BP
    ADOXQ AX, CX
    ADCXQ DI, BP
    MULXQ ·qE2+32(SB), AX, DI
    ADOXQ AX, BP
    ADCXQ R8, DI
    MULXQ ·qE2+40(SB), AX, R8
    ADOXQ AX, DI
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R8
    ADOXQ R9, R8
    XORQ DX, DX
    MOVQ -120(SP), DX
    MULXQ BX, AX, R9
    ADOXQ AX, R10
    ADCXQ R9, R11
    MULXQ R12, AX, R9
    ADOXQ AX, R11
    ADCXQ R9, CX
    MULXQ SI, AX, R9
    ADOXQ AX, CX
    ADCXQ R9, BP
    MULXQ R13, AX, R9
    ADOXQ AX, BP
    ADCXQ R9, DI
    MULXQ R14, AX, R9
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ R15, AX, R9
    ADOXQ AX, R8
    // add the last carries to R9
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R9
    ADOXQ DX, R9
    MOVQ R10, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    PUSHQ R9
    MULXQ ·qE2+0(SB), AX, R9
    ADCXQ R10, AX
    MOVQ R9, R10
    POPQ R9
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R11, R10
    MULXQ ·qE2+8(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ CX, R11
    MULXQ ·qE2+16(SB), AX, CX
    ADOXQ AX, R11
    ADCXQ BP, CX
    MULXQ ·qE2+24(SB), AX, BP
    ADOXQ AX, CX
    ADCXQ DI, BP
    MULXQ ·qE2+32(SB), AX, DI
    ADOXQ AX, BP
    ADCXQ R8, DI
    MULXQ ·qE2+40(SB), AX, R8
    ADOXQ AX, DI
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R8
    ADOXQ R9, R8
    XORQ DX, DX
    MOVQ -128(SP), DX
    MULXQ BX, AX, R9
    ADOXQ AX, R10
    ADCXQ R9, R11
    MULXQ R12, AX, R9
    ADOXQ AX, R11
    ADCXQ R9, CX
    MULXQ SI, AX, R9
    ADOXQ AX, CX
    ADCXQ R9, BP
    MULXQ R13, AX, R9
    ADOXQ AX, BP
    ADCXQ R9, DI
    MULXQ R14, AX, R9
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ R15, AX, R9
    ADOXQ AX, R8
    // add the last carries to R9
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R9
    ADOXQ DX, R9
    MOVQ R10, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    PUSHQ R9
    MULXQ ·qE2+0(SB), AX, R9
    ADCXQ R10, AX
    MOVQ R9, R10
    POPQ R9
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R11, R10
    MULXQ ·qE2+8(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ CX, R11
    MULXQ ·qE2+16(SB), AX, CX
    ADOXQ AX, R11
    ADCXQ BP, CX
    MULXQ ·qE2+24(SB), AX, BP
    ADOXQ AX, CX
    ADCXQ DI, BP
    MULXQ ·qE2+32(SB), AX, DI
    ADOXQ AX, BP
    ADCXQ R8, DI
    MULXQ ·qE2+40(SB), AX, R8
    ADOXQ AX, DI
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R8
    ADOXQ R9, R8
    XORQ DX, DX
    MOVQ -136(SP), DX
    MULXQ BX, AX, R9
    ADOXQ AX, R10
    ADCXQ R9, R11
    MULXQ R12, AX, R9
    ADOXQ AX, R11
    ADCXQ R9, CX
    MULXQ SI, AX, R9
    ADOXQ AX, CX
    ADCXQ R9, BP
    MULXQ R13, AX, R9
    ADOXQ AX, BP
    ADCXQ R9, DI
    MULXQ R14, AX, R9
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ R15, AX, R9
    ADOXQ AX, R8
    // add the last carries to R9
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R9
    ADOXQ DX, R9
    MOVQ R10, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    PUSHQ R9
    MULXQ ·qE2+0(SB), AX, R9
    ADCXQ R10, AX
    MOVQ R9, R10
    POPQ R9
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R11, R10
    MULXQ ·qE2+8(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ CX, R11
    MULXQ ·qE2+16(SB), AX, CX
    ADOXQ AX, R11
    ADCXQ BP, CX
    MULXQ ·qE2+24(SB), AX, BP
    ADOXQ AX, CX
    ADCXQ DI, BP
    MULXQ ·qE2+32(SB), AX, DI
    ADOXQ AX, BP
    ADCXQ R8, DI
    MULXQ ·qE2+40(SB), AX, R8
    ADOXQ AX, DI
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R8
    ADOXQ R9, R8
    XORQ DX, DX
    MOVQ -144(SP), DX
    MULXQ BX, AX, R9
    ADOXQ AX, R10
    ADCXQ R9, R11
    MULXQ R12, AX, R9
    ADOXQ AX, R11
    ADCXQ R9, CX
    MULXQ SI, AX, R9
    ADOXQ AX, CX
    ADCXQ R9, BP
    MULXQ R13, AX, R9
    ADOXQ AX, BP
    ADCXQ R9, DI
    MULXQ R14, AX, R9
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ R15, AX, R9
    ADOXQ AX, R8
    // add the last carries to R9
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R9
    ADOXQ DX, R9
    MOVQ R10, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    PUSHQ R9
    MULXQ ·qE2+0(SB), AX, R9
    ADCXQ R10, AX
    MOVQ R9, R10
    POPQ R9
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R11, R10
    MULXQ ·qE2+8(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ CX, R11
    MULXQ ·qE2+16(SB), AX, CX
    ADOXQ AX, R11
    ADCXQ BP, CX
    MULXQ ·qE2+24(SB), AX, BP
    ADOXQ AX, CX
    ADCXQ DI, BP
    MULXQ ·qE2+32(SB), AX, DI
    ADOXQ AX, BP
    ADCXQ R8, DI
    MULXQ ·qE2+40(SB), AX, R8
    ADOXQ AX, DI
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R8
    ADOXQ R9, R8
    XORQ DX, DX
    MOVQ -152(SP), DX
    MULXQ BX, AX, R9
    ADOXQ AX, R10
    ADCXQ R9, R11
    MULXQ R12, AX, R9
    ADOXQ AX, R11
    ADCXQ R9, CX
    MULXQ SI, AX, R9
    ADOXQ AX, CX
    ADCXQ R9, BP
    MULXQ R13, AX, R9
    ADOXQ AX, BP
    ADCXQ R9, DI
    MULXQ R14, AX, R9
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ R15, AX, R9
    ADOXQ AX, R8
    // add the last carries to R9
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R9
    ADOXQ DX, R9
    MOVQ R10, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    PUSHQ R9
    MULXQ ·qE2+0(SB), AX, R9
    ADCXQ R10, AX
    MOVQ R9, R10
    POPQ R9
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R11, R10
    MULXQ ·qE2+8(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ CX, R11
    MULXQ ·qE2+16(SB), AX, CX
    ADOXQ AX, R11
    ADCXQ BP, CX
    MULXQ ·qE2+24(SB), AX, BP
    ADOXQ AX, CX
    ADCXQ DI, BP
    MULXQ ·qE2+32(SB), AX, DI
    ADOXQ AX, BP
    ADCXQ R8, DI
    MULXQ ·qE2+40(SB), AX, R8
    ADOXQ AX, DI
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R8
    ADOXQ R9, R8
    MOVQ R10, R9
    MOVQ R11, BX
    MOVQ CX, R12
    MOVQ BP, SI
    MOVQ DI, R13
    MOVQ R8, R14
    SUBQ ·qE2+0(SB), R9
    SBBQ ·qE2+8(SB), BX
    SBBQ ·qE2+16(SB), R12
    SBBQ ·qE2+24(SB), SI
    SBBQ ·qE2+32(SB), R13
    SBBQ ·qE2+40(SB), R14
    CMOVQCC R9, R10
    CMOVQCC BX, R11
    CMOVQCC R12, CX
    CMOVQCC SI, BP
    CMOVQCC R13, DI
    CMOVQCC R14, R8
    MOVQ z+0(FP), DX
    SUBQ -16(SP), R10
    SBBQ -24(SP), R11
    SBBQ -32(SP), CX
    SBBQ -40(SP), BP
    SBBQ -48(SP), DI
    SBBQ -56(SP), R8
    MOVQ $0xb9feffffffffaaab, R15
    MOVQ $0x1eabfffeb153ffff, R9
    MOVQ $0x6730d2a0f6b0f624, BX
    MOVQ $0x64774b84f38512bf, R12
    MOVQ $0x4b1ba7b6434bacd7, SI
    MOVQ $0x1a0111ea397fe69a, R13
    MOVQ $0x0000000000000000, R14
    CMOVQCC R14, R15
    CMOVQCC R14, R9
    CMOVQCC R14, BX
    CMOVQCC R14, R12
    CMOVQCC R14, SI
    CMOVQCC R14, R13
    ADDQ R15, R10
    ADCQ R9, R11
    ADCQ BX, CX
    ADCQ R12, BP
    ADCQ SI, DI
    ADCQ R13, R8
    SUBQ -64(SP), R10
    SBBQ -72(SP), R11
    SBBQ -80(SP), CX
    SBBQ -88(SP), BP
    SBBQ -96(SP), DI
    SBBQ -104(SP), R8
    MOVQ $0xb9feffffffffaaab, R14
    MOVQ $0x1eabfffeb153ffff, R15
    MOVQ $0x6730d2a0f6b0f624, R9
    MOVQ $0x64774b84f38512bf, BX
    MOVQ $0x4b1ba7b6434bacd7, R12
    MOVQ $0x1a0111ea397fe69a, SI
    MOVQ $0x0000000000000000, R13
    CMOVQCC R13, R14
    CMOVQCC R13, R15
    CMOVQCC R13, R9
    CMOVQCC R13, BX
    CMOVQCC R13, R12
    CMOVQCC R13, SI
    ADDQ R14, R10
    ADCQ R15, R11
    ADCQ R9, CX
    ADCQ BX, BP
    ADCQ R12, DI
    ADCQ SI, R8
    MOVQ R10, 48(DX)
    MOVQ R11, 56(DX)
    MOVQ CX, 64(DX)
    MOVQ BP, 72(DX)
    MOVQ DI, 80(DX)
    MOVQ R8, 88(DX)
    MOVQ -16(SP), R10
    MOVQ -24(SP), R11
    MOVQ -32(SP), CX
    MOVQ -40(SP), BP
    MOVQ -48(SP), DI
    MOVQ -56(SP), R8
    SUBQ -64(SP), R10
    SBBQ -72(SP), R11
    SBBQ -80(SP), CX
    SBBQ -88(SP), BP
    SBBQ -96(SP), DI
    SBBQ -104(SP), R8
    MOVQ $0xb9feffffffffaaab, R13
    MOVQ $0x1eabfffeb153ffff, R14
    MOVQ $0x6730d2a0f6b0f624, R15
    MOVQ $0x64774b84f38512bf, R9
    MOVQ $0x4b1ba7b6434bacd7, BX
    MOVQ $0x1a0111ea397fe69a, R12
    MOVQ $0x0000000000000000, SI
    CMOVQCC SI, R13
    CMOVQCC SI, R14
    CMOVQCC SI, R15
    CMOVQCC SI, R9
    CMOVQCC SI, BX
    CMOVQCC SI, R12
    ADDQ R13, R10
    ADCQ R14, R11
    ADCQ R15, CX
    ADCQ R9, BP
    ADCQ BX, DI
    ADCQ R12, R8
    MOVQ R10, 0(DX)
    MOVQ R11, 8(DX)
    MOVQ CX, 16(DX)
    MOVQ BP, 24(DX)
    MOVQ DI, 32(DX)
    MOVQ R8, 40(DX)
    RET
l5:
    MOVQ z+0(FP), AX
    MOVQ AX, (SP)
    MOVQ x+8(FP), AX
    MOVQ AX, 8(SP)
    MOVQ y+16(FP), AX
    MOVQ AX, 16(SP)
CALL ·mulGenericE2(SB)
    RET
