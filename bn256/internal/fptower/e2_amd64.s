
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
    MOVQ y+16(FP), DX
    ADDQ 0(DX), BX
    ADCQ 8(DX), BP
    ADCQ 16(DX), SI
    ADCQ 24(DX), DI
    MOVQ res+0(FP), CX
    MOVQ BX, R8
    MOVQ BP, R9
    MOVQ SI, R10
    MOVQ DI, R11
    SUBQ ·qE2+0(SB), R8
    SBBQ ·qE2+8(SB), R9
    SBBQ ·qE2+16(SB), R10
    SBBQ ·qE2+24(SB), R11
    CMOVQCC R8, BX
    CMOVQCC R9, BP
    CMOVQCC R10, SI
    CMOVQCC R11, DI
    MOVQ BX, 0(CX)
    MOVQ BP, 8(CX)
    MOVQ SI, 16(CX)
    MOVQ DI, 24(CX)
    MOVQ 32(AX), BX
    MOVQ 40(AX), BP
    MOVQ 48(AX), SI
    MOVQ 56(AX), DI
    ADDQ 32(DX), BX
    ADCQ 40(DX), BP
    ADCQ 48(DX), SI
    ADCQ 56(DX), DI
    MOVQ BX, R12
    MOVQ BP, R13
    MOVQ SI, R14
    MOVQ DI, R15
    SUBQ ·qE2+0(SB), R12
    SBBQ ·qE2+8(SB), R13
    SBBQ ·qE2+16(SB), R14
    SBBQ ·qE2+24(SB), R15
    CMOVQCC R12, BX
    CMOVQCC R13, BP
    CMOVQCC R14, SI
    CMOVQCC R15, DI
    MOVQ BX, 32(CX)
    MOVQ BP, 40(CX)
    MOVQ SI, 48(CX)
    MOVQ DI, 56(CX)
    RET

TEXT ·doubleE2(SB), NOSPLIT, $0-16
    MOVQ res+0(FP), DX
    MOVQ x+8(FP), AX
    MOVQ 0(AX), CX
    MOVQ 8(AX), BX
    MOVQ 16(AX), BP
    MOVQ 24(AX), SI
    ADDQ CX, CX
    ADCQ BX, BX
    ADCQ BP, BP
    ADCQ SI, SI
    MOVQ CX, DI
    MOVQ BX, R8
    MOVQ BP, R9
    MOVQ SI, R10
    SUBQ ·qE2+0(SB), DI
    SBBQ ·qE2+8(SB), R8
    SBBQ ·qE2+16(SB), R9
    SBBQ ·qE2+24(SB), R10
    CMOVQCC DI, CX
    CMOVQCC R8, BX
    CMOVQCC R9, BP
    CMOVQCC R10, SI
    MOVQ CX, 0(DX)
    MOVQ BX, 8(DX)
    MOVQ BP, 16(DX)
    MOVQ SI, 24(DX)
    MOVQ 32(AX), CX
    MOVQ 40(AX), BX
    MOVQ 48(AX), BP
    MOVQ 56(AX), SI
    ADDQ CX, CX
    ADCQ BX, BX
    ADCQ BP, BP
    ADCQ SI, SI
    MOVQ CX, R11
    MOVQ BX, R12
    MOVQ BP, R13
    MOVQ SI, R14
    SUBQ ·qE2+0(SB), R11
    SBBQ ·qE2+8(SB), R12
    SBBQ ·qE2+16(SB), R13
    SBBQ ·qE2+24(SB), R14
    CMOVQCC R11, CX
    CMOVQCC R12, BX
    CMOVQCC R13, BP
    CMOVQCC R14, SI
    MOVQ CX, 32(DX)
    MOVQ BX, 40(DX)
    MOVQ BP, 48(DX)
    MOVQ SI, 56(DX)
    RET

TEXT ·subE2(SB), NOSPLIT, $0-24
    MOVQ x+8(FP), BP
    MOVQ y+16(FP), SI
    MOVQ 0(BP), AX
    MOVQ 8(BP), DX
    MOVQ 16(BP), CX
    MOVQ 24(BP), BX
    SUBQ 0(SI), AX
    SBBQ 8(SI), DX
    SBBQ 16(SI), CX
    SBBQ 24(SI), BX
    MOVQ $0x3c208c16d87cfd47, DI
    MOVQ $0x97816a916871ca8d, R8
    MOVQ $0xb85045b68181585d, R9
    MOVQ $0x30644e72e131a029, R10
    MOVQ $0x0000000000000000, R11
    CMOVQCC R11, DI
    CMOVQCC R11, R8
    CMOVQCC R11, R9
    CMOVQCC R11, R10
    ADDQ DI, AX
    ADCQ R8, DX
    ADCQ R9, CX
    ADCQ R10, BX
    MOVQ res+0(FP), R12
    MOVQ AX, 0(R12)
    MOVQ DX, 8(R12)
    MOVQ CX, 16(R12)
    MOVQ BX, 24(R12)
    MOVQ 32(BP), AX
    MOVQ 40(BP), DX
    MOVQ 48(BP), CX
    MOVQ 56(BP), BX
    SUBQ 32(SI), AX
    SBBQ 40(SI), DX
    SBBQ 48(SI), CX
    SBBQ 56(SI), BX
    MOVQ $0x3c208c16d87cfd47, R13
    MOVQ $0x97816a916871ca8d, R14
    MOVQ $0xb85045b68181585d, R15
    MOVQ $0x30644e72e131a029, R11
    MOVQ $0x0000000000000000, DI
    CMOVQCC DI, R13
    CMOVQCC DI, R14
    CMOVQCC DI, R15
    CMOVQCC DI, R11
    ADDQ R13, AX
    ADCQ R14, DX
    ADCQ R15, CX
    ADCQ R11, BX
    MOVQ res+0(FP), BP
    MOVQ AX, 32(BP)
    MOVQ DX, 40(BP)
    MOVQ CX, 48(BP)
    MOVQ BX, 56(BP)
    RET

TEXT ·negE2(SB), NOSPLIT, $0-16
    MOVQ res+0(FP), DX
    MOVQ x+8(FP), AX
    MOVQ 0(AX), BX
    MOVQ 8(AX), BP
    MOVQ 16(AX), SI
    MOVQ 24(AX), DI
    MOVQ BX, AX
    ORQ BP, AX
    ORQ SI, AX
    ORQ DI, AX
    TESTQ AX, AX
    JNE l1
    MOVQ AX, 32(DX)
    MOVQ AX, 40(DX)
    MOVQ AX, 48(DX)
    MOVQ AX, 56(DX)
    JMP l3
l1:
    MOVQ $0x3c208c16d87cfd47, CX
    SUBQ BX, CX
    MOVQ CX, 0(DX)
    MOVQ $0x97816a916871ca8d, CX
    SBBQ BP, CX
    MOVQ CX, 8(DX)
    MOVQ $0xb85045b68181585d, CX
    SBBQ SI, CX
    MOVQ CX, 16(DX)
    MOVQ $0x30644e72e131a029, CX
    SBBQ DI, CX
    MOVQ CX, 24(DX)
l3:
    MOVQ x+8(FP), AX
    MOVQ 32(AX), BX
    MOVQ 40(AX), BP
    MOVQ 48(AX), SI
    MOVQ 56(AX), DI
    MOVQ BX, AX
    ORQ BP, AX
    ORQ SI, AX
    ORQ DI, AX
    TESTQ AX, AX
    JNE l2
    MOVQ AX, 32(DX)
    MOVQ AX, 40(DX)
    MOVQ AX, 48(DX)
    MOVQ AX, 56(DX)
    RET
l2:
    MOVQ $0x3c208c16d87cfd47, CX
    SUBQ BX, CX
    MOVQ CX, 32(DX)
    MOVQ $0x97816a916871ca8d, CX
    SBBQ BP, CX
    MOVQ CX, 40(DX)
    MOVQ $0xb85045b68181585d, CX
    SBBQ SI, CX
    MOVQ CX, 48(DX)
    MOVQ $0x30644e72e131a029, CX
    SBBQ DI, CX
    MOVQ CX, 56(DX)
    RET

TEXT ·mulAdxE2(SB), $24-24
NO_LOCAL_POINTERS
    CMPB ·supportAdx(SB), $0x0000000000000001
    JNE l4
    MOVQ x+8(FP), R9
    MOVQ 32(R9), R14
    MOVQ 40(R9), R15
    MOVQ 48(R9), CX
    MOVQ 56(R9), BX
    ADDQ 0(R9), R14
    ADCQ 8(R9), R15
    ADCQ 16(R9), CX
    ADCQ 24(R9), BX
    MOVQ R14, R10
    MOVQ R15, R11
    MOVQ CX, R12
    MOVQ BX, R13
    SUBQ ·qE2+0(SB), R10
    SBBQ ·qE2+8(SB), R11
    SBBQ ·qE2+16(SB), R12
    SBBQ ·qE2+24(SB), R13
    CMOVQCC R10, R14
    CMOVQCC R11, R15
    CMOVQCC R12, CX
    CMOVQCC R13, BX
    MOVQ y+16(FP), R9
    MOVQ 32(R9), BP
    MOVQ 40(R9), SI
    MOVQ 48(R9), DI
    MOVQ 56(R9), R8
    ADDQ 0(R9), BP
    ADCQ 8(R9), SI
    ADCQ 16(R9), DI
    ADCQ 24(R9), R8
    MOVQ BP, R10
    MOVQ SI, R11
    MOVQ DI, R12
    MOVQ R8, R13
    SUBQ ·qE2+0(SB), R10
    SBBQ ·qE2+8(SB), R11
    SBBQ ·qE2+16(SB), R12
    SBBQ ·qE2+24(SB), R13
    CMOVQCC R10, BP
    CMOVQCC R11, SI
    CMOVQCC R12, DI
    CMOVQCC R13, R8
    XORQ DX, DX
    MOVQ BP, DX
    MULXQ R14, R10, R11
    MULXQ R15, AX, R12
    ADOXQ AX, R11
    MULXQ CX, AX, R13
    ADOXQ AX, R12
    MULXQ BX, AX, R9
    ADOXQ AX, R13
    // add the last carries to R9
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R9
    ADOXQ DX, R9
    MOVQ R10, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qE2+0(SB), AX, BP
    ADCXQ R10, AX
    MOVQ BP, R10
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R11, R10
    MULXQ ·qE2+8(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qE2+16(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qE2+24(SB), AX, R13
    ADOXQ AX, R12
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R13
    ADOXQ R9, R13
    XORQ DX, DX
    MOVQ SI, DX
    MULXQ R14, AX, R9
    ADOXQ AX, R10
    ADCXQ R9, R11
    MULXQ R15, AX, R9
    ADOXQ AX, R11
    ADCXQ R9, R12
    MULXQ CX, AX, R9
    ADOXQ AX, R12
    ADCXQ R9, R13
    MULXQ BX, AX, R9
    ADOXQ AX, R13
    // add the last carries to R9
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R9
    ADOXQ DX, R9
    MOVQ R10, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qE2+0(SB), AX, BP
    ADCXQ R10, AX
    MOVQ BP, R10
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R11, R10
    MULXQ ·qE2+8(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qE2+16(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qE2+24(SB), AX, R13
    ADOXQ AX, R12
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R13
    ADOXQ R9, R13
    XORQ DX, DX
    MOVQ DI, DX
    MULXQ R14, AX, R9
    ADOXQ AX, R10
    ADCXQ R9, R11
    MULXQ R15, AX, R9
    ADOXQ AX, R11
    ADCXQ R9, R12
    MULXQ CX, AX, R9
    ADOXQ AX, R12
    ADCXQ R9, R13
    MULXQ BX, AX, R9
    ADOXQ AX, R13
    // add the last carries to R9
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R9
    ADOXQ DX, R9
    MOVQ R10, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qE2+0(SB), AX, SI
    ADCXQ R10, AX
    MOVQ SI, R10
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R11, R10
    MULXQ ·qE2+8(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qE2+16(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qE2+24(SB), AX, R13
    ADOXQ AX, R12
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R13
    ADOXQ R9, R13
    XORQ DX, DX
    MOVQ R8, DX
    MULXQ R14, AX, R9
    ADOXQ AX, R10
    ADCXQ R9, R11
    MULXQ R15, AX, R9
    ADOXQ AX, R11
    ADCXQ R9, R12
    MULXQ CX, AX, R9
    ADOXQ AX, R12
    ADCXQ R9, R13
    MULXQ BX, AX, R9
    ADOXQ AX, R13
    // add the last carries to R9
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R9
    ADOXQ DX, R9
    MOVQ R10, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qE2+0(SB), AX, BP
    ADCXQ R10, AX
    MOVQ BP, R10
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R11, R10
    MULXQ ·qE2+8(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qE2+16(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qE2+24(SB), AX, R13
    ADOXQ AX, R12
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R13
    ADOXQ R9, R13
    MOVQ R10, DI
    MOVQ R11, SI
    MOVQ R12, R8
    MOVQ R13, BP
    SUBQ ·qE2+0(SB), DI
    SBBQ ·qE2+8(SB), SI
    SBBQ ·qE2+16(SB), R8
    SBBQ ·qE2+24(SB), BP
    CMOVQCC DI, R10
    CMOVQCC SI, R11
    CMOVQCC R8, R12
    CMOVQCC BP, R13
    MOVQ R10, R14
    MOVQ R11, R15
    MOVQ R12, CX
    MOVQ R13, BX
    XORQ DX, DX
    MOVQ y+16(FP), R9
    MOVQ 0(R9), DX
    MOVQ x+8(FP), R9
    MULXQ 0(R9), DI, SI
    MOVQ x+8(FP), R9
    MULXQ 8(R9), AX, R8
    ADOXQ AX, SI
    MOVQ x+8(FP), R9
    MULXQ 16(R9), AX, BP
    ADOXQ AX, R8
    MOVQ x+8(FP), R9
    MULXQ 24(R9), AX, R10
    ADOXQ AX, BP
    // add the last carries to R10
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R10
    ADOXQ DX, R10
    MOVQ DI, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qE2+0(SB), AX, R11
    ADCXQ DI, AX
    MOVQ R11, DI
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ SI, DI
    MULXQ ·qE2+8(SB), AX, SI
    ADOXQ AX, DI
    ADCXQ R8, SI
    MULXQ ·qE2+16(SB), AX, R8
    ADOXQ AX, SI
    ADCXQ BP, R8
    MULXQ ·qE2+24(SB), AX, BP
    ADOXQ AX, R8
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, BP
    ADOXQ R10, BP
    XORQ DX, DX
    MOVQ y+16(FP), R9
    MOVQ 8(R9), DX
    MOVQ x+8(FP), R9
    MULXQ 0(R9), AX, R10
    ADOXQ AX, DI
    MOVQ x+8(FP), R9
    ADCXQ R10, SI
    MULXQ 8(R9), AX, R10
    ADOXQ AX, SI
    MOVQ x+8(FP), R9
    ADCXQ R10, R8
    MULXQ 16(R9), AX, R10
    ADOXQ AX, R8
    MOVQ x+8(FP), R9
    ADCXQ R10, BP
    MULXQ 24(R9), AX, R10
    ADOXQ AX, BP
    // add the last carries to R10
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R10
    ADOXQ DX, R10
    MOVQ DI, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qE2+0(SB), AX, R12
    ADCXQ DI, AX
    MOVQ R12, DI
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ SI, DI
    MULXQ ·qE2+8(SB), AX, SI
    ADOXQ AX, DI
    ADCXQ R8, SI
    MULXQ ·qE2+16(SB), AX, R8
    ADOXQ AX, SI
    ADCXQ BP, R8
    MULXQ ·qE2+24(SB), AX, BP
    ADOXQ AX, R8
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, BP
    ADOXQ R10, BP
    XORQ DX, DX
    MOVQ y+16(FP), R9
    MOVQ 16(R9), DX
    MOVQ x+8(FP), R9
    MULXQ 0(R9), AX, R10
    ADOXQ AX, DI
    MOVQ x+8(FP), R9
    ADCXQ R10, SI
    MULXQ 8(R9), AX, R10
    ADOXQ AX, SI
    MOVQ x+8(FP), R9
    ADCXQ R10, R8
    MULXQ 16(R9), AX, R10
    ADOXQ AX, R8
    MOVQ x+8(FP), R9
    ADCXQ R10, BP
    MULXQ 24(R9), AX, R10
    ADOXQ AX, BP
    // add the last carries to R10
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R10
    ADOXQ DX, R10
    MOVQ DI, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qE2+0(SB), AX, R13
    ADCXQ DI, AX
    MOVQ R13, DI
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ SI, DI
    MULXQ ·qE2+8(SB), AX, SI
    ADOXQ AX, DI
    ADCXQ R8, SI
    MULXQ ·qE2+16(SB), AX, R8
    ADOXQ AX, SI
    ADCXQ BP, R8
    MULXQ ·qE2+24(SB), AX, BP
    ADOXQ AX, R8
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, BP
    ADOXQ R10, BP
    XORQ DX, DX
    MOVQ y+16(FP), R9
    MOVQ 24(R9), DX
    MOVQ x+8(FP), R9
    MULXQ 0(R9), AX, R10
    ADOXQ AX, DI
    MOVQ x+8(FP), R9
    ADCXQ R10, SI
    MULXQ 8(R9), AX, R10
    ADOXQ AX, SI
    MOVQ x+8(FP), R9
    ADCXQ R10, R8
    MULXQ 16(R9), AX, R10
    ADOXQ AX, R8
    MOVQ x+8(FP), R9
    ADCXQ R10, BP
    MULXQ 24(R9), AX, R10
    ADOXQ AX, BP
    // add the last carries to R10
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R10
    ADOXQ DX, R10
    MOVQ DI, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qE2+0(SB), AX, R11
    ADCXQ DI, AX
    MOVQ R11, DI
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ SI, DI
    MULXQ ·qE2+8(SB), AX, SI
    ADOXQ AX, DI
    ADCXQ R8, SI
    MULXQ ·qE2+16(SB), AX, R8
    ADOXQ AX, SI
    ADCXQ BP, R8
    MULXQ ·qE2+24(SB), AX, BP
    ADOXQ AX, R8
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, BP
    ADOXQ R10, BP
    MOVQ DI, R12
    MOVQ SI, R13
    MOVQ R8, R11
    MOVQ BP, R10
    SUBQ ·qE2+0(SB), R12
    SBBQ ·qE2+8(SB), R13
    SBBQ ·qE2+16(SB), R11
    SBBQ ·qE2+24(SB), R10
    CMOVQCC R12, DI
    CMOVQCC R13, SI
    CMOVQCC R11, R8
    CMOVQCC R10, BP
    SUBQ DI, R14
    SBBQ SI, R15
    SBBQ R8, CX
    SBBQ BP, BX
    MOVQ $0x3c208c16d87cfd47, R9
    MOVQ $0x97816a916871ca8d, R12
    MOVQ $0xb85045b68181585d, R13
    MOVQ $0x30644e72e131a029, R11
    MOVQ $0x0000000000000000, R10
    CMOVQCC R10, R9
    CMOVQCC R10, R12
    CMOVQCC R10, R13
    CMOVQCC R10, R11
    ADDQ R9, R14
    ADCQ R12, R15
    ADCQ R13, CX
    ADCQ R11, BX
    PUSHQ R14
    PUSHQ R15
    PUSHQ CX
    PUSHQ BX
    XORQ DX, DX
    MOVQ y+16(FP), R10
    MOVQ 32(R10), DX
    MOVQ x+8(FP), R10
    MULXQ 32(R10), R9, R12
    MOVQ x+8(FP), R10
    MULXQ 40(R10), AX, R13
    ADOXQ AX, R12
    MOVQ x+8(FP), R10
    MULXQ 48(R10), AX, R11
    ADOXQ AX, R13
    MOVQ x+8(FP), R10
    MULXQ 56(R10), AX, R14
    ADOXQ AX, R11
    // add the last carries to R14
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R14
    ADOXQ DX, R14
    MOVQ R9, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qE2+0(SB), AX, R15
    ADCXQ R9, AX
    MOVQ R15, R9
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R12, R9
    MULXQ ·qE2+8(SB), AX, R12
    ADOXQ AX, R9
    ADCXQ R13, R12
    MULXQ ·qE2+16(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ R11, R13
    MULXQ ·qE2+24(SB), AX, R11
    ADOXQ AX, R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R11
    ADOXQ R14, R11
    XORQ DX, DX
    MOVQ y+16(FP), R10
    MOVQ 40(R10), DX
    MOVQ x+8(FP), R10
    MULXQ 32(R10), AX, R14
    ADOXQ AX, R9
    MOVQ x+8(FP), R10
    ADCXQ R14, R12
    MULXQ 40(R10), AX, R14
    ADOXQ AX, R12
    MOVQ x+8(FP), R10
    ADCXQ R14, R13
    MULXQ 48(R10), AX, R14
    ADOXQ AX, R13
    MOVQ x+8(FP), R10
    ADCXQ R14, R11
    MULXQ 56(R10), AX, R14
    ADOXQ AX, R11
    // add the last carries to R14
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R14
    ADOXQ DX, R14
    MOVQ R9, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qE2+0(SB), AX, CX
    ADCXQ R9, AX
    MOVQ CX, R9
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R12, R9
    MULXQ ·qE2+8(SB), AX, R12
    ADOXQ AX, R9
    ADCXQ R13, R12
    MULXQ ·qE2+16(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ R11, R13
    MULXQ ·qE2+24(SB), AX, R11
    ADOXQ AX, R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R11
    ADOXQ R14, R11
    XORQ DX, DX
    MOVQ y+16(FP), R10
    MOVQ 48(R10), DX
    MOVQ x+8(FP), R10
    MULXQ 32(R10), AX, R14
    ADOXQ AX, R9
    MOVQ x+8(FP), R10
    ADCXQ R14, R12
    MULXQ 40(R10), AX, R14
    ADOXQ AX, R12
    MOVQ x+8(FP), R10
    ADCXQ R14, R13
    MULXQ 48(R10), AX, R14
    ADOXQ AX, R13
    MOVQ x+8(FP), R10
    ADCXQ R14, R11
    MULXQ 56(R10), AX, R14
    ADOXQ AX, R11
    // add the last carries to R14
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R14
    ADOXQ DX, R14
    MOVQ R9, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qE2+0(SB), AX, BX
    ADCXQ R9, AX
    MOVQ BX, R9
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R12, R9
    MULXQ ·qE2+8(SB), AX, R12
    ADOXQ AX, R9
    ADCXQ R13, R12
    MULXQ ·qE2+16(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ R11, R13
    MULXQ ·qE2+24(SB), AX, R11
    ADOXQ AX, R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R11
    ADOXQ R14, R11
    XORQ DX, DX
    MOVQ y+16(FP), R10
    MOVQ 56(R10), DX
    MOVQ x+8(FP), R10
    MULXQ 32(R10), AX, R14
    ADOXQ AX, R9
    MOVQ x+8(FP), R10
    ADCXQ R14, R12
    MULXQ 40(R10), AX, R14
    ADOXQ AX, R12
    MOVQ x+8(FP), R10
    ADCXQ R14, R13
    MULXQ 48(R10), AX, R14
    ADOXQ AX, R13
    MOVQ x+8(FP), R10
    ADCXQ R14, R11
    MULXQ 56(R10), AX, R14
    ADOXQ AX, R11
    // add the last carries to R14
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R14
    ADOXQ DX, R14
    MOVQ R9, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qE2+0(SB), AX, R15
    ADCXQ R9, AX
    MOVQ R15, R9
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R12, R9
    MULXQ ·qE2+8(SB), AX, R12
    ADOXQ AX, R9
    ADCXQ R13, R12
    MULXQ ·qE2+16(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ R11, R13
    MULXQ ·qE2+24(SB), AX, R11
    ADOXQ AX, R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R11
    ADOXQ R14, R11
    MOVQ R9, CX
    MOVQ R12, BX
    MOVQ R13, R15
    MOVQ R11, R14
    SUBQ ·qE2+0(SB), CX
    SBBQ ·qE2+8(SB), BX
    SBBQ ·qE2+16(SB), R15
    SBBQ ·qE2+24(SB), R14
    CMOVQCC CX, R9
    CMOVQCC BX, R12
    CMOVQCC R15, R13
    CMOVQCC R14, R11
    SUBQ R9, DI
    SBBQ R12, SI
    SBBQ R13, R8
    SBBQ R11, BP
    MOVQ $0x3c208c16d87cfd47, R10
    MOVQ $0x97816a916871ca8d, CX
    MOVQ $0xb85045b68181585d, BX
    MOVQ $0x30644e72e131a029, R15
    MOVQ $0x0000000000000000, R14
    CMOVQCC R14, R10
    CMOVQCC R14, CX
    CMOVQCC R14, BX
    CMOVQCC R14, R15
    ADDQ R10, DI
    ADCQ CX, SI
    ADCQ BX, R8
    ADCQ R15, BP
    MOVQ res+0(FP), R14
    MOVQ DI, 0(R14)
    MOVQ SI, 8(R14)
    MOVQ R8, 16(R14)
    MOVQ BP, 24(R14)
    POPQ BP
    POPQ R8
    POPQ SI
    POPQ DI
    SUBQ R9, DI
    SBBQ R12, SI
    SBBQ R13, R8
    SBBQ R11, BP
    MOVQ $0x3c208c16d87cfd47, R10
    MOVQ $0x97816a916871ca8d, CX
    MOVQ $0xb85045b68181585d, BX
    MOVQ $0x30644e72e131a029, R15
    MOVQ $0x0000000000000000, R9
    CMOVQCC R9, R10
    CMOVQCC R9, CX
    CMOVQCC R9, BX
    CMOVQCC R9, R15
    ADDQ R10, DI
    ADCQ CX, SI
    ADCQ BX, R8
    ADCQ R15, BP
    MOVQ DI, 32(R14)
    MOVQ SI, 40(R14)
    MOVQ R8, 48(R14)
    MOVQ BP, 56(R14)
    RET
l4:
    MOVQ res+0(FP), AX
    MOVQ AX, (SP)
    MOVQ x+8(FP), AX
    MOVQ AX, 8(SP)
    MOVQ y+16(FP), AX
    MOVQ AX, 16(SP)
CALL ·mulGenericE2(SB)
    RET

TEXT ·squareAdxE2(SB), $16-16
NO_LOCAL_POINTERS
    CMPB ·supportAdx(SB), $0x0000000000000001
    JNE l5
    MOVQ x+8(FP), R9
    MOVQ 32(R9), R14
    MOVQ 40(R9), R15
    MOVQ 48(R9), CX
    MOVQ 56(R9), BX
    MOVQ 0(R9), BP
    MOVQ 8(R9), SI
    MOVQ 16(R9), DI
    MOVQ 24(R9), R8
    ADDQ BP, R14
    ADCQ SI, R15
    ADCQ DI, CX
    ADCQ R8, BX
    MOVQ R14, R10
    MOVQ R15, R11
    MOVQ CX, R12
    MOVQ BX, R13
    SUBQ ·qE2+0(SB), R10
    SBBQ ·qE2+8(SB), R11
    SBBQ ·qE2+16(SB), R12
    SBBQ ·qE2+24(SB), R13
    CMOVQCC R10, R14
    CMOVQCC R11, R15
    CMOVQCC R12, CX
    CMOVQCC R13, BX
    SUBQ 32(R9), BP
    SBBQ 40(R9), SI
    SBBQ 48(R9), DI
    SBBQ 56(R9), R8
    MOVQ $0x3c208c16d87cfd47, R10
    MOVQ $0x97816a916871ca8d, R11
    MOVQ $0xb85045b68181585d, R12
    MOVQ $0x30644e72e131a029, R13
    MOVQ $0x0000000000000000, R9
    CMOVQCC R9, R10
    CMOVQCC R9, R11
    CMOVQCC R9, R12
    CMOVQCC R9, R13
    ADDQ R10, BP
    ADCQ R11, SI
    ADCQ R12, DI
    ADCQ R13, R8
    XORQ DX, DX
    MOVQ BP, DX
    MULXQ R14, R9, R10
    MULXQ R15, AX, R11
    ADOXQ AX, R10
    MULXQ CX, AX, R12
    ADOXQ AX, R11
    MULXQ BX, AX, R13
    ADOXQ AX, R12
    // add the last carries to R13
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R13
    ADOXQ DX, R13
    MOVQ R9, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qE2+0(SB), AX, BP
    ADCXQ R9, AX
    MOVQ BP, R9
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R10, R9
    MULXQ ·qE2+8(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qE2+16(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qE2+24(SB), AX, R12
    ADOXQ AX, R11
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ R13, R12
    XORQ DX, DX
    MOVQ SI, DX
    MULXQ R14, AX, R13
    ADOXQ AX, R9
    ADCXQ R13, R10
    MULXQ R15, AX, R13
    ADOXQ AX, R10
    ADCXQ R13, R11
    MULXQ CX, AX, R13
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ BX, AX, R13
    ADOXQ AX, R12
    // add the last carries to R13
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R13
    ADOXQ DX, R13
    MOVQ R9, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qE2+0(SB), AX, BP
    ADCXQ R9, AX
    MOVQ BP, R9
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R10, R9
    MULXQ ·qE2+8(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qE2+16(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qE2+24(SB), AX, R12
    ADOXQ AX, R11
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ R13, R12
    XORQ DX, DX
    MOVQ DI, DX
    MULXQ R14, AX, R13
    ADOXQ AX, R9
    ADCXQ R13, R10
    MULXQ R15, AX, R13
    ADOXQ AX, R10
    ADCXQ R13, R11
    MULXQ CX, AX, R13
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ BX, AX, R13
    ADOXQ AX, R12
    // add the last carries to R13
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R13
    ADOXQ DX, R13
    MOVQ R9, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qE2+0(SB), AX, SI
    ADCXQ R9, AX
    MOVQ SI, R9
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R10, R9
    MULXQ ·qE2+8(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qE2+16(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qE2+24(SB), AX, R12
    ADOXQ AX, R11
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ R13, R12
    XORQ DX, DX
    MOVQ R8, DX
    MULXQ R14, AX, R13
    ADOXQ AX, R9
    ADCXQ R13, R10
    MULXQ R15, AX, R13
    ADOXQ AX, R10
    ADCXQ R13, R11
    MULXQ CX, AX, R13
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ BX, AX, R13
    ADOXQ AX, R12
    // add the last carries to R13
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R13
    ADOXQ DX, R13
    MOVQ R9, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qE2+0(SB), AX, BP
    ADCXQ R9, AX
    MOVQ BP, R9
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R10, R9
    MULXQ ·qE2+8(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qE2+16(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qE2+24(SB), AX, R12
    ADOXQ AX, R11
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ R13, R12
    MOVQ R9, DI
    MOVQ R10, SI
    MOVQ R11, R8
    MOVQ R12, BP
    SUBQ ·qE2+0(SB), DI
    SBBQ ·qE2+8(SB), SI
    SBBQ ·qE2+16(SB), R8
    SBBQ ·qE2+24(SB), BP
    CMOVQCC DI, R9
    CMOVQCC SI, R10
    CMOVQCC R8, R11
    CMOVQCC BP, R12
    MOVQ R9, R14
    MOVQ R10, R15
    MOVQ R11, CX
    MOVQ R12, BX
    MOVQ x+8(FP), R13
    XORQ DX, DX
    MOVQ 32(R13), DX
    MULXQ 0(R13), DI, SI
    MULXQ 8(R13), AX, R8
    ADOXQ AX, SI
    MULXQ 16(R13), AX, BP
    ADOXQ AX, R8
    MULXQ 24(R13), AX, R9
    ADOXQ AX, BP
    // add the last carries to R9
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R9
    ADOXQ DX, R9
    MOVQ DI, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qE2+0(SB), AX, R10
    ADCXQ DI, AX
    MOVQ R10, DI
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ SI, DI
    MULXQ ·qE2+8(SB), AX, SI
    ADOXQ AX, DI
    ADCXQ R8, SI
    MULXQ ·qE2+16(SB), AX, R8
    ADOXQ AX, SI
    ADCXQ BP, R8
    MULXQ ·qE2+24(SB), AX, BP
    ADOXQ AX, R8
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, BP
    ADOXQ R9, BP
    XORQ DX, DX
    MOVQ 40(R13), DX
    MULXQ 0(R13), AX, R9
    ADOXQ AX, DI
    ADCXQ R9, SI
    MULXQ 8(R13), AX, R9
    ADOXQ AX, SI
    ADCXQ R9, R8
    MULXQ 16(R13), AX, R9
    ADOXQ AX, R8
    ADCXQ R9, BP
    MULXQ 24(R13), AX, R9
    ADOXQ AX, BP
    // add the last carries to R9
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R9
    ADOXQ DX, R9
    MOVQ DI, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qE2+0(SB), AX, R11
    ADCXQ DI, AX
    MOVQ R11, DI
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ SI, DI
    MULXQ ·qE2+8(SB), AX, SI
    ADOXQ AX, DI
    ADCXQ R8, SI
    MULXQ ·qE2+16(SB), AX, R8
    ADOXQ AX, SI
    ADCXQ BP, R8
    MULXQ ·qE2+24(SB), AX, BP
    ADOXQ AX, R8
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, BP
    ADOXQ R9, BP
    XORQ DX, DX
    MOVQ 48(R13), DX
    MULXQ 0(R13), AX, R9
    ADOXQ AX, DI
    ADCXQ R9, SI
    MULXQ 8(R13), AX, R9
    ADOXQ AX, SI
    ADCXQ R9, R8
    MULXQ 16(R13), AX, R9
    ADOXQ AX, R8
    ADCXQ R9, BP
    MULXQ 24(R13), AX, R9
    ADOXQ AX, BP
    // add the last carries to R9
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R9
    ADOXQ DX, R9
    MOVQ DI, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qE2+0(SB), AX, R12
    ADCXQ DI, AX
    MOVQ R12, DI
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ SI, DI
    MULXQ ·qE2+8(SB), AX, SI
    ADOXQ AX, DI
    ADCXQ R8, SI
    MULXQ ·qE2+16(SB), AX, R8
    ADOXQ AX, SI
    ADCXQ BP, R8
    MULXQ ·qE2+24(SB), AX, BP
    ADOXQ AX, R8
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, BP
    ADOXQ R9, BP
    XORQ DX, DX
    MOVQ 56(R13), DX
    MULXQ 0(R13), AX, R9
    ADOXQ AX, DI
    ADCXQ R9, SI
    MULXQ 8(R13), AX, R9
    ADOXQ AX, SI
    ADCXQ R9, R8
    MULXQ 16(R13), AX, R9
    ADOXQ AX, R8
    ADCXQ R9, BP
    MULXQ 24(R13), AX, R9
    ADOXQ AX, BP
    // add the last carries to R9
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R9
    ADOXQ DX, R9
    MOVQ DI, DX
    MULXQ ·qE2Inv0(SB), DX, AX                             // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qE2+0(SB), AX, R10
    ADCXQ DI, AX
    MOVQ R10, DI
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ SI, DI
    MULXQ ·qE2+8(SB), AX, SI
    ADOXQ AX, DI
    ADCXQ R8, SI
    MULXQ ·qE2+16(SB), AX, R8
    ADOXQ AX, SI
    ADCXQ BP, R8
    MULXQ ·qE2+24(SB), AX, BP
    ADOXQ AX, R8
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, BP
    ADOXQ R9, BP
    MOVQ DI, R11
    MOVQ SI, R12
    MOVQ R8, R10
    MOVQ BP, R9
    SUBQ ·qE2+0(SB), R11
    SBBQ ·qE2+8(SB), R12
    SBBQ ·qE2+16(SB), R10
    SBBQ ·qE2+24(SB), R9
    CMOVQCC R11, DI
    CMOVQCC R12, SI
    CMOVQCC R10, R8
    CMOVQCC R9, BP
    ADDQ DI, DI
    ADCQ SI, SI
    ADCQ R8, R8
    ADCQ BP, BP
    MOVQ res+0(FP), R13
    MOVQ DI, R11
    MOVQ SI, R12
    MOVQ R8, R10
    MOVQ BP, R9
    SUBQ ·qE2+0(SB), R11
    SBBQ ·qE2+8(SB), R12
    SBBQ ·qE2+16(SB), R10
    SBBQ ·qE2+24(SB), R9
    CMOVQCC R11, DI
    CMOVQCC R12, SI
    CMOVQCC R10, R8
    CMOVQCC R9, BP
    MOVQ DI, 32(R13)
    MOVQ SI, 40(R13)
    MOVQ R8, 48(R13)
    MOVQ BP, 56(R13)
    MOVQ R14, 0(R13)
    MOVQ R15, 8(R13)
    MOVQ CX, 16(R13)
    MOVQ BX, 24(R13)
    RET
l5:
    MOVQ res+0(FP), AX
    MOVQ AX, (SP)
    MOVQ x+8(FP), AX
    MOVQ AX, 8(SP)
CALL ·squareGenericE2(SB)
    RET

TEXT ·mulNonResE2(SB), NOSPLIT, $0-16
    MOVQ x+8(FP), R9
    MOVQ 0(R9), AX
    MOVQ 8(R9), DX
    MOVQ 16(R9), CX
    MOVQ 24(R9), BX
    ADDQ AX, AX
    ADCQ DX, DX
    ADCQ CX, CX
    ADCQ BX, BX
    MOVQ AX, R10
    MOVQ DX, R11
    MOVQ CX, R12
    MOVQ BX, R13
    SUBQ ·qE2+0(SB), R10
    SBBQ ·qE2+8(SB), R11
    SBBQ ·qE2+16(SB), R12
    SBBQ ·qE2+24(SB), R13
    CMOVQCC R10, AX
    CMOVQCC R11, DX
    CMOVQCC R12, CX
    CMOVQCC R13, BX
    ADDQ AX, AX
    ADCQ DX, DX
    ADCQ CX, CX
    ADCQ BX, BX
    MOVQ AX, R14
    MOVQ DX, R15
    MOVQ CX, R10
    MOVQ BX, R11
    SUBQ ·qE2+0(SB), R14
    SBBQ ·qE2+8(SB), R15
    SBBQ ·qE2+16(SB), R10
    SBBQ ·qE2+24(SB), R11
    CMOVQCC R14, AX
    CMOVQCC R15, DX
    CMOVQCC R10, CX
    CMOVQCC R11, BX
    ADDQ AX, AX
    ADCQ DX, DX
    ADCQ CX, CX
    ADCQ BX, BX
    MOVQ AX, R12
    MOVQ DX, R13
    MOVQ CX, R14
    MOVQ BX, R15
    SUBQ ·qE2+0(SB), R12
    SBBQ ·qE2+8(SB), R13
    SBBQ ·qE2+16(SB), R14
    SBBQ ·qE2+24(SB), R15
    CMOVQCC R12, AX
    CMOVQCC R13, DX
    CMOVQCC R14, CX
    CMOVQCC R15, BX
    ADDQ 0(R9), AX
    ADCQ 8(R9), DX
    ADCQ 16(R9), CX
    ADCQ 24(R9), BX
    MOVQ AX, R10
    MOVQ DX, R11
    MOVQ CX, R12
    MOVQ BX, R13
    SUBQ ·qE2+0(SB), R10
    SBBQ ·qE2+8(SB), R11
    SBBQ ·qE2+16(SB), R12
    SBBQ ·qE2+24(SB), R13
    CMOVQCC R10, AX
    CMOVQCC R11, DX
    CMOVQCC R12, CX
    CMOVQCC R13, BX
    MOVQ 32(R9), BP
    MOVQ 40(R9), SI
    MOVQ 48(R9), DI
    MOVQ 56(R9), R8
    SUBQ BP, AX
    SBBQ SI, DX
    SBBQ DI, CX
    SBBQ R8, BX
    MOVQ $0x3c208c16d87cfd47, R14
    MOVQ $0x97816a916871ca8d, R15
    MOVQ $0xb85045b68181585d, R10
    MOVQ $0x30644e72e131a029, R11
    MOVQ $0x0000000000000000, R12
    CMOVQCC R12, R14
    CMOVQCC R12, R15
    CMOVQCC R12, R10
    CMOVQCC R12, R11
    ADDQ R14, AX
    ADCQ R15, DX
    ADCQ R10, CX
    ADCQ R11, BX
    ADDQ BP, BP
    ADCQ SI, SI
    ADCQ DI, DI
    ADCQ R8, R8
    MOVQ BP, R13
    MOVQ SI, R12
    MOVQ DI, R14
    MOVQ R8, R15
    SUBQ ·qE2+0(SB), R13
    SBBQ ·qE2+8(SB), R12
    SBBQ ·qE2+16(SB), R14
    SBBQ ·qE2+24(SB), R15
    CMOVQCC R13, BP
    CMOVQCC R12, SI
    CMOVQCC R14, DI
    CMOVQCC R15, R8
    ADDQ BP, BP
    ADCQ SI, SI
    ADCQ DI, DI
    ADCQ R8, R8
    MOVQ BP, R10
    MOVQ SI, R11
    MOVQ DI, R13
    MOVQ R8, R12
    SUBQ ·qE2+0(SB), R10
    SBBQ ·qE2+8(SB), R11
    SBBQ ·qE2+16(SB), R13
    SBBQ ·qE2+24(SB), R12
    CMOVQCC R10, BP
    CMOVQCC R11, SI
    CMOVQCC R13, DI
    CMOVQCC R12, R8
    ADDQ BP, BP
    ADCQ SI, SI
    ADCQ DI, DI
    ADCQ R8, R8
    MOVQ BP, R14
    MOVQ SI, R15
    MOVQ DI, R10
    MOVQ R8, R11
    SUBQ ·qE2+0(SB), R14
    SBBQ ·qE2+8(SB), R15
    SBBQ ·qE2+16(SB), R10
    SBBQ ·qE2+24(SB), R11
    CMOVQCC R14, BP
    CMOVQCC R15, SI
    CMOVQCC R10, DI
    CMOVQCC R11, R8
    ADDQ 32(R9), BP
    ADCQ 40(R9), SI
    ADCQ 48(R9), DI
    ADCQ 56(R9), R8
    MOVQ BP, R13
    MOVQ SI, R12
    MOVQ DI, R14
    MOVQ R8, R15
    SUBQ ·qE2+0(SB), R13
    SBBQ ·qE2+8(SB), R12
    SBBQ ·qE2+16(SB), R14
    SBBQ ·qE2+24(SB), R15
    CMOVQCC R13, BP
    CMOVQCC R12, SI
    CMOVQCC R14, DI
    CMOVQCC R15, R8
    ADDQ 0(R9), BP
    ADCQ 8(R9), SI
    ADCQ 16(R9), DI
    ADCQ 24(R9), R8
    MOVQ BP, R10
    MOVQ SI, R11
    MOVQ DI, R13
    MOVQ R8, R12
    SUBQ ·qE2+0(SB), R10
    SBBQ ·qE2+8(SB), R11
    SBBQ ·qE2+16(SB), R13
    SBBQ ·qE2+24(SB), R12
    CMOVQCC R10, BP
    CMOVQCC R11, SI
    CMOVQCC R13, DI
    CMOVQCC R12, R8
    MOVQ res+0(FP), R9
    MOVQ AX, 0(R9)
    MOVQ DX, 8(R9)
    MOVQ CX, 16(R9)
    MOVQ BX, 24(R9)
    MOVQ BP, 32(R9)
    MOVQ SI, 40(R9)
    MOVQ DI, 48(R9)
    MOVQ R8, 56(R9)
    RET
