
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

TEXT ·mul(SB), $96-24

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
    CMPB ·supportAdx(SB), $0x0000000000000001
    JNE l1
    XORQ DX, DX
    MOVQ y+16(FP), DX
    MOVQ 0(DX), DX
    MOVQ x+8(FP), AX
    MOVQ 0(AX), AX
    MULXQ AX, R14, R15
    MOVQ x+8(FP), AX
    MOVQ 8(AX), AX
    MULXQ AX, AX, CX
    ADOXQ AX, R15
    MOVQ x+8(FP), AX
    MOVQ 16(AX), AX
    MULXQ AX, AX, BX
    ADOXQ AX, CX
    MOVQ x+8(FP), AX
    MOVQ 24(AX), AX
    MULXQ AX, AX, BP
    ADOXQ AX, BX
    MOVQ x+8(FP), AX
    MOVQ 32(AX), AX
    MULXQ AX, AX, SI
    ADOXQ AX, BP
    MOVQ x+8(FP), AX
    MOVQ 40(AX), AX
    MULXQ AX, AX, DI
    ADOXQ AX, SI
    MOVQ x+8(FP), AX
    MOVQ 48(AX), AX
    MULXQ AX, AX, R8
    ADOXQ AX, DI
    MOVQ x+8(FP), AX
    MOVQ 56(AX), AX
    MULXQ AX, AX, R9
    ADOXQ AX, R8
    MOVQ x+8(FP), AX
    MOVQ 64(AX), AX
    MULXQ AX, AX, R10
    ADOXQ AX, R9
    MOVQ x+8(FP), AX
    MOVQ 72(AX), AX
    MULXQ AX, AX, R11
    ADOXQ AX, R10
    MOVQ x+8(FP), AX
    MOVQ 80(AX), AX
    MULXQ AX, AX, R12
    ADOXQ AX, R11
    MOVQ x+8(FP), AX
    MOVQ 88(AX), AX
    MULXQ AX, AX, R13
    ADOXQ AX, R12
    // add the last carries to R13
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R13
    ADOXQ DX, R13
    PUSHQ R13
    MOVQ R14, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R13
    ADCXQ R14, AX
    MOVQ R13, R14
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R15, R14
    MULXQ ·qElement+8(SB), AX, R15
    ADOXQ AX, R14
    ADCXQ CX, R15
    MULXQ ·qElement+16(SB), AX, CX
    ADOXQ AX, R15
    ADCXQ BX, CX
    MULXQ ·qElement+24(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+32(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+40(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+48(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+56(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+64(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+72(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+80(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+88(SB), AX, R12
    ADOXQ AX, R11
    POPQ R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ R13, R12
    XORQ DX, DX
    MOVQ y+16(FP), DX
    MOVQ 8(DX), DX
    MOVQ x+8(FP), AX
    MOVQ 0(AX), AX
    MULXQ AX, AX, R13
    ADOXQ AX, R14
    MOVQ x+8(FP), AX
    MOVQ 8(AX), AX
    ADCXQ R13, R15
    MULXQ AX, AX, R13
    ADOXQ AX, R15
    MOVQ x+8(FP), AX
    MOVQ 16(AX), AX
    ADCXQ R13, CX
    MULXQ AX, AX, R13
    ADOXQ AX, CX
    MOVQ x+8(FP), AX
    MOVQ 24(AX), AX
    ADCXQ R13, BX
    MULXQ AX, AX, R13
    ADOXQ AX, BX
    MOVQ x+8(FP), AX
    MOVQ 32(AX), AX
    ADCXQ R13, BP
    MULXQ AX, AX, R13
    ADOXQ AX, BP
    MOVQ x+8(FP), AX
    MOVQ 40(AX), AX
    ADCXQ R13, SI
    MULXQ AX, AX, R13
    ADOXQ AX, SI
    MOVQ x+8(FP), AX
    MOVQ 48(AX), AX
    ADCXQ R13, DI
    MULXQ AX, AX, R13
    ADOXQ AX, DI
    MOVQ x+8(FP), AX
    MOVQ 56(AX), AX
    ADCXQ R13, R8
    MULXQ AX, AX, R13
    ADOXQ AX, R8
    MOVQ x+8(FP), AX
    MOVQ 64(AX), AX
    ADCXQ R13, R9
    MULXQ AX, AX, R13
    ADOXQ AX, R9
    MOVQ x+8(FP), AX
    MOVQ 72(AX), AX
    ADCXQ R13, R10
    MULXQ AX, AX, R13
    ADOXQ AX, R10
    MOVQ x+8(FP), AX
    MOVQ 80(AX), AX
    ADCXQ R13, R11
    MULXQ AX, AX, R13
    ADOXQ AX, R11
    MOVQ x+8(FP), AX
    MOVQ 88(AX), AX
    ADCXQ R13, R12
    MULXQ AX, AX, R13
    ADOXQ AX, R12
    // add the last carries to R13
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R13
    ADOXQ DX, R13
    PUSHQ R13
    MOVQ R14, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R13
    ADCXQ R14, AX
    MOVQ R13, R14
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R15, R14
    MULXQ ·qElement+8(SB), AX, R15
    ADOXQ AX, R14
    ADCXQ CX, R15
    MULXQ ·qElement+16(SB), AX, CX
    ADOXQ AX, R15
    ADCXQ BX, CX
    MULXQ ·qElement+24(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+32(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+40(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+48(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+56(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+64(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+72(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+80(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+88(SB), AX, R12
    ADOXQ AX, R11
    POPQ R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ R13, R12
    XORQ DX, DX
    MOVQ y+16(FP), DX
    MOVQ 16(DX), DX
    MOVQ x+8(FP), AX
    MOVQ 0(AX), AX
    MULXQ AX, AX, R13
    ADOXQ AX, R14
    MOVQ x+8(FP), AX
    MOVQ 8(AX), AX
    ADCXQ R13, R15
    MULXQ AX, AX, R13
    ADOXQ AX, R15
    MOVQ x+8(FP), AX
    MOVQ 16(AX), AX
    ADCXQ R13, CX
    MULXQ AX, AX, R13
    ADOXQ AX, CX
    MOVQ x+8(FP), AX
    MOVQ 24(AX), AX
    ADCXQ R13, BX
    MULXQ AX, AX, R13
    ADOXQ AX, BX
    MOVQ x+8(FP), AX
    MOVQ 32(AX), AX
    ADCXQ R13, BP
    MULXQ AX, AX, R13
    ADOXQ AX, BP
    MOVQ x+8(FP), AX
    MOVQ 40(AX), AX
    ADCXQ R13, SI
    MULXQ AX, AX, R13
    ADOXQ AX, SI
    MOVQ x+8(FP), AX
    MOVQ 48(AX), AX
    ADCXQ R13, DI
    MULXQ AX, AX, R13
    ADOXQ AX, DI
    MOVQ x+8(FP), AX
    MOVQ 56(AX), AX
    ADCXQ R13, R8
    MULXQ AX, AX, R13
    ADOXQ AX, R8
    MOVQ x+8(FP), AX
    MOVQ 64(AX), AX
    ADCXQ R13, R9
    MULXQ AX, AX, R13
    ADOXQ AX, R9
    MOVQ x+8(FP), AX
    MOVQ 72(AX), AX
    ADCXQ R13, R10
    MULXQ AX, AX, R13
    ADOXQ AX, R10
    MOVQ x+8(FP), AX
    MOVQ 80(AX), AX
    ADCXQ R13, R11
    MULXQ AX, AX, R13
    ADOXQ AX, R11
    MOVQ x+8(FP), AX
    MOVQ 88(AX), AX
    ADCXQ R13, R12
    MULXQ AX, AX, R13
    ADOXQ AX, R12
    // add the last carries to R13
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R13
    ADOXQ DX, R13
    PUSHQ R13
    MOVQ R14, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R13
    ADCXQ R14, AX
    MOVQ R13, R14
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R15, R14
    MULXQ ·qElement+8(SB), AX, R15
    ADOXQ AX, R14
    ADCXQ CX, R15
    MULXQ ·qElement+16(SB), AX, CX
    ADOXQ AX, R15
    ADCXQ BX, CX
    MULXQ ·qElement+24(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+32(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+40(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+48(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+56(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+64(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+72(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+80(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+88(SB), AX, R12
    ADOXQ AX, R11
    POPQ R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ R13, R12
    XORQ DX, DX
    MOVQ y+16(FP), DX
    MOVQ 24(DX), DX
    MOVQ x+8(FP), AX
    MOVQ 0(AX), AX
    MULXQ AX, AX, R13
    ADOXQ AX, R14
    MOVQ x+8(FP), AX
    MOVQ 8(AX), AX
    ADCXQ R13, R15
    MULXQ AX, AX, R13
    ADOXQ AX, R15
    MOVQ x+8(FP), AX
    MOVQ 16(AX), AX
    ADCXQ R13, CX
    MULXQ AX, AX, R13
    ADOXQ AX, CX
    MOVQ x+8(FP), AX
    MOVQ 24(AX), AX
    ADCXQ R13, BX
    MULXQ AX, AX, R13
    ADOXQ AX, BX
    MOVQ x+8(FP), AX
    MOVQ 32(AX), AX
    ADCXQ R13, BP
    MULXQ AX, AX, R13
    ADOXQ AX, BP
    MOVQ x+8(FP), AX
    MOVQ 40(AX), AX
    ADCXQ R13, SI
    MULXQ AX, AX, R13
    ADOXQ AX, SI
    MOVQ x+8(FP), AX
    MOVQ 48(AX), AX
    ADCXQ R13, DI
    MULXQ AX, AX, R13
    ADOXQ AX, DI
    MOVQ x+8(FP), AX
    MOVQ 56(AX), AX
    ADCXQ R13, R8
    MULXQ AX, AX, R13
    ADOXQ AX, R8
    MOVQ x+8(FP), AX
    MOVQ 64(AX), AX
    ADCXQ R13, R9
    MULXQ AX, AX, R13
    ADOXQ AX, R9
    MOVQ x+8(FP), AX
    MOVQ 72(AX), AX
    ADCXQ R13, R10
    MULXQ AX, AX, R13
    ADOXQ AX, R10
    MOVQ x+8(FP), AX
    MOVQ 80(AX), AX
    ADCXQ R13, R11
    MULXQ AX, AX, R13
    ADOXQ AX, R11
    MOVQ x+8(FP), AX
    MOVQ 88(AX), AX
    ADCXQ R13, R12
    MULXQ AX, AX, R13
    ADOXQ AX, R12
    // add the last carries to R13
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R13
    ADOXQ DX, R13
    PUSHQ R13
    MOVQ R14, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R13
    ADCXQ R14, AX
    MOVQ R13, R14
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R15, R14
    MULXQ ·qElement+8(SB), AX, R15
    ADOXQ AX, R14
    ADCXQ CX, R15
    MULXQ ·qElement+16(SB), AX, CX
    ADOXQ AX, R15
    ADCXQ BX, CX
    MULXQ ·qElement+24(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+32(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+40(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+48(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+56(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+64(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+72(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+80(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+88(SB), AX, R12
    ADOXQ AX, R11
    POPQ R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ R13, R12
    XORQ DX, DX
    MOVQ y+16(FP), DX
    MOVQ 32(DX), DX
    MOVQ x+8(FP), AX
    MOVQ 0(AX), AX
    MULXQ AX, AX, R13
    ADOXQ AX, R14
    MOVQ x+8(FP), AX
    MOVQ 8(AX), AX
    ADCXQ R13, R15
    MULXQ AX, AX, R13
    ADOXQ AX, R15
    MOVQ x+8(FP), AX
    MOVQ 16(AX), AX
    ADCXQ R13, CX
    MULXQ AX, AX, R13
    ADOXQ AX, CX
    MOVQ x+8(FP), AX
    MOVQ 24(AX), AX
    ADCXQ R13, BX
    MULXQ AX, AX, R13
    ADOXQ AX, BX
    MOVQ x+8(FP), AX
    MOVQ 32(AX), AX
    ADCXQ R13, BP
    MULXQ AX, AX, R13
    ADOXQ AX, BP
    MOVQ x+8(FP), AX
    MOVQ 40(AX), AX
    ADCXQ R13, SI
    MULXQ AX, AX, R13
    ADOXQ AX, SI
    MOVQ x+8(FP), AX
    MOVQ 48(AX), AX
    ADCXQ R13, DI
    MULXQ AX, AX, R13
    ADOXQ AX, DI
    MOVQ x+8(FP), AX
    MOVQ 56(AX), AX
    ADCXQ R13, R8
    MULXQ AX, AX, R13
    ADOXQ AX, R8
    MOVQ x+8(FP), AX
    MOVQ 64(AX), AX
    ADCXQ R13, R9
    MULXQ AX, AX, R13
    ADOXQ AX, R9
    MOVQ x+8(FP), AX
    MOVQ 72(AX), AX
    ADCXQ R13, R10
    MULXQ AX, AX, R13
    ADOXQ AX, R10
    MOVQ x+8(FP), AX
    MOVQ 80(AX), AX
    ADCXQ R13, R11
    MULXQ AX, AX, R13
    ADOXQ AX, R11
    MOVQ x+8(FP), AX
    MOVQ 88(AX), AX
    ADCXQ R13, R12
    MULXQ AX, AX, R13
    ADOXQ AX, R12
    // add the last carries to R13
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R13
    ADOXQ DX, R13
    PUSHQ R13
    MOVQ R14, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R13
    ADCXQ R14, AX
    MOVQ R13, R14
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R15, R14
    MULXQ ·qElement+8(SB), AX, R15
    ADOXQ AX, R14
    ADCXQ CX, R15
    MULXQ ·qElement+16(SB), AX, CX
    ADOXQ AX, R15
    ADCXQ BX, CX
    MULXQ ·qElement+24(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+32(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+40(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+48(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+56(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+64(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+72(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+80(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+88(SB), AX, R12
    ADOXQ AX, R11
    POPQ R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ R13, R12
    XORQ DX, DX
    MOVQ y+16(FP), DX
    MOVQ 40(DX), DX
    MOVQ x+8(FP), AX
    MOVQ 0(AX), AX
    MULXQ AX, AX, R13
    ADOXQ AX, R14
    MOVQ x+8(FP), AX
    MOVQ 8(AX), AX
    ADCXQ R13, R15
    MULXQ AX, AX, R13
    ADOXQ AX, R15
    MOVQ x+8(FP), AX
    MOVQ 16(AX), AX
    ADCXQ R13, CX
    MULXQ AX, AX, R13
    ADOXQ AX, CX
    MOVQ x+8(FP), AX
    MOVQ 24(AX), AX
    ADCXQ R13, BX
    MULXQ AX, AX, R13
    ADOXQ AX, BX
    MOVQ x+8(FP), AX
    MOVQ 32(AX), AX
    ADCXQ R13, BP
    MULXQ AX, AX, R13
    ADOXQ AX, BP
    MOVQ x+8(FP), AX
    MOVQ 40(AX), AX
    ADCXQ R13, SI
    MULXQ AX, AX, R13
    ADOXQ AX, SI
    MOVQ x+8(FP), AX
    MOVQ 48(AX), AX
    ADCXQ R13, DI
    MULXQ AX, AX, R13
    ADOXQ AX, DI
    MOVQ x+8(FP), AX
    MOVQ 56(AX), AX
    ADCXQ R13, R8
    MULXQ AX, AX, R13
    ADOXQ AX, R8
    MOVQ x+8(FP), AX
    MOVQ 64(AX), AX
    ADCXQ R13, R9
    MULXQ AX, AX, R13
    ADOXQ AX, R9
    MOVQ x+8(FP), AX
    MOVQ 72(AX), AX
    ADCXQ R13, R10
    MULXQ AX, AX, R13
    ADOXQ AX, R10
    MOVQ x+8(FP), AX
    MOVQ 80(AX), AX
    ADCXQ R13, R11
    MULXQ AX, AX, R13
    ADOXQ AX, R11
    MOVQ x+8(FP), AX
    MOVQ 88(AX), AX
    ADCXQ R13, R12
    MULXQ AX, AX, R13
    ADOXQ AX, R12
    // add the last carries to R13
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R13
    ADOXQ DX, R13
    PUSHQ R13
    MOVQ R14, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R13
    ADCXQ R14, AX
    MOVQ R13, R14
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R15, R14
    MULXQ ·qElement+8(SB), AX, R15
    ADOXQ AX, R14
    ADCXQ CX, R15
    MULXQ ·qElement+16(SB), AX, CX
    ADOXQ AX, R15
    ADCXQ BX, CX
    MULXQ ·qElement+24(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+32(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+40(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+48(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+56(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+64(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+72(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+80(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+88(SB), AX, R12
    ADOXQ AX, R11
    POPQ R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ R13, R12
    XORQ DX, DX
    MOVQ y+16(FP), DX
    MOVQ 48(DX), DX
    MOVQ x+8(FP), AX
    MOVQ 0(AX), AX
    MULXQ AX, AX, R13
    ADOXQ AX, R14
    MOVQ x+8(FP), AX
    MOVQ 8(AX), AX
    ADCXQ R13, R15
    MULXQ AX, AX, R13
    ADOXQ AX, R15
    MOVQ x+8(FP), AX
    MOVQ 16(AX), AX
    ADCXQ R13, CX
    MULXQ AX, AX, R13
    ADOXQ AX, CX
    MOVQ x+8(FP), AX
    MOVQ 24(AX), AX
    ADCXQ R13, BX
    MULXQ AX, AX, R13
    ADOXQ AX, BX
    MOVQ x+8(FP), AX
    MOVQ 32(AX), AX
    ADCXQ R13, BP
    MULXQ AX, AX, R13
    ADOXQ AX, BP
    MOVQ x+8(FP), AX
    MOVQ 40(AX), AX
    ADCXQ R13, SI
    MULXQ AX, AX, R13
    ADOXQ AX, SI
    MOVQ x+8(FP), AX
    MOVQ 48(AX), AX
    ADCXQ R13, DI
    MULXQ AX, AX, R13
    ADOXQ AX, DI
    MOVQ x+8(FP), AX
    MOVQ 56(AX), AX
    ADCXQ R13, R8
    MULXQ AX, AX, R13
    ADOXQ AX, R8
    MOVQ x+8(FP), AX
    MOVQ 64(AX), AX
    ADCXQ R13, R9
    MULXQ AX, AX, R13
    ADOXQ AX, R9
    MOVQ x+8(FP), AX
    MOVQ 72(AX), AX
    ADCXQ R13, R10
    MULXQ AX, AX, R13
    ADOXQ AX, R10
    MOVQ x+8(FP), AX
    MOVQ 80(AX), AX
    ADCXQ R13, R11
    MULXQ AX, AX, R13
    ADOXQ AX, R11
    MOVQ x+8(FP), AX
    MOVQ 88(AX), AX
    ADCXQ R13, R12
    MULXQ AX, AX, R13
    ADOXQ AX, R12
    // add the last carries to R13
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R13
    ADOXQ DX, R13
    PUSHQ R13
    MOVQ R14, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R13
    ADCXQ R14, AX
    MOVQ R13, R14
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R15, R14
    MULXQ ·qElement+8(SB), AX, R15
    ADOXQ AX, R14
    ADCXQ CX, R15
    MULXQ ·qElement+16(SB), AX, CX
    ADOXQ AX, R15
    ADCXQ BX, CX
    MULXQ ·qElement+24(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+32(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+40(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+48(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+56(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+64(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+72(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+80(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+88(SB), AX, R12
    ADOXQ AX, R11
    POPQ R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ R13, R12
    XORQ DX, DX
    MOVQ y+16(FP), DX
    MOVQ 56(DX), DX
    MOVQ x+8(FP), AX
    MOVQ 0(AX), AX
    MULXQ AX, AX, R13
    ADOXQ AX, R14
    MOVQ x+8(FP), AX
    MOVQ 8(AX), AX
    ADCXQ R13, R15
    MULXQ AX, AX, R13
    ADOXQ AX, R15
    MOVQ x+8(FP), AX
    MOVQ 16(AX), AX
    ADCXQ R13, CX
    MULXQ AX, AX, R13
    ADOXQ AX, CX
    MOVQ x+8(FP), AX
    MOVQ 24(AX), AX
    ADCXQ R13, BX
    MULXQ AX, AX, R13
    ADOXQ AX, BX
    MOVQ x+8(FP), AX
    MOVQ 32(AX), AX
    ADCXQ R13, BP
    MULXQ AX, AX, R13
    ADOXQ AX, BP
    MOVQ x+8(FP), AX
    MOVQ 40(AX), AX
    ADCXQ R13, SI
    MULXQ AX, AX, R13
    ADOXQ AX, SI
    MOVQ x+8(FP), AX
    MOVQ 48(AX), AX
    ADCXQ R13, DI
    MULXQ AX, AX, R13
    ADOXQ AX, DI
    MOVQ x+8(FP), AX
    MOVQ 56(AX), AX
    ADCXQ R13, R8
    MULXQ AX, AX, R13
    ADOXQ AX, R8
    MOVQ x+8(FP), AX
    MOVQ 64(AX), AX
    ADCXQ R13, R9
    MULXQ AX, AX, R13
    ADOXQ AX, R9
    MOVQ x+8(FP), AX
    MOVQ 72(AX), AX
    ADCXQ R13, R10
    MULXQ AX, AX, R13
    ADOXQ AX, R10
    MOVQ x+8(FP), AX
    MOVQ 80(AX), AX
    ADCXQ R13, R11
    MULXQ AX, AX, R13
    ADOXQ AX, R11
    MOVQ x+8(FP), AX
    MOVQ 88(AX), AX
    ADCXQ R13, R12
    MULXQ AX, AX, R13
    ADOXQ AX, R12
    // add the last carries to R13
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R13
    ADOXQ DX, R13
    PUSHQ R13
    MOVQ R14, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R13
    ADCXQ R14, AX
    MOVQ R13, R14
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R15, R14
    MULXQ ·qElement+8(SB), AX, R15
    ADOXQ AX, R14
    ADCXQ CX, R15
    MULXQ ·qElement+16(SB), AX, CX
    ADOXQ AX, R15
    ADCXQ BX, CX
    MULXQ ·qElement+24(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+32(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+40(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+48(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+56(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+64(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+72(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+80(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+88(SB), AX, R12
    ADOXQ AX, R11
    POPQ R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ R13, R12
    XORQ DX, DX
    MOVQ y+16(FP), DX
    MOVQ 64(DX), DX
    MOVQ x+8(FP), AX
    MOVQ 0(AX), AX
    MULXQ AX, AX, R13
    ADOXQ AX, R14
    MOVQ x+8(FP), AX
    MOVQ 8(AX), AX
    ADCXQ R13, R15
    MULXQ AX, AX, R13
    ADOXQ AX, R15
    MOVQ x+8(FP), AX
    MOVQ 16(AX), AX
    ADCXQ R13, CX
    MULXQ AX, AX, R13
    ADOXQ AX, CX
    MOVQ x+8(FP), AX
    MOVQ 24(AX), AX
    ADCXQ R13, BX
    MULXQ AX, AX, R13
    ADOXQ AX, BX
    MOVQ x+8(FP), AX
    MOVQ 32(AX), AX
    ADCXQ R13, BP
    MULXQ AX, AX, R13
    ADOXQ AX, BP
    MOVQ x+8(FP), AX
    MOVQ 40(AX), AX
    ADCXQ R13, SI
    MULXQ AX, AX, R13
    ADOXQ AX, SI
    MOVQ x+8(FP), AX
    MOVQ 48(AX), AX
    ADCXQ R13, DI
    MULXQ AX, AX, R13
    ADOXQ AX, DI
    MOVQ x+8(FP), AX
    MOVQ 56(AX), AX
    ADCXQ R13, R8
    MULXQ AX, AX, R13
    ADOXQ AX, R8
    MOVQ x+8(FP), AX
    MOVQ 64(AX), AX
    ADCXQ R13, R9
    MULXQ AX, AX, R13
    ADOXQ AX, R9
    MOVQ x+8(FP), AX
    MOVQ 72(AX), AX
    ADCXQ R13, R10
    MULXQ AX, AX, R13
    ADOXQ AX, R10
    MOVQ x+8(FP), AX
    MOVQ 80(AX), AX
    ADCXQ R13, R11
    MULXQ AX, AX, R13
    ADOXQ AX, R11
    MOVQ x+8(FP), AX
    MOVQ 88(AX), AX
    ADCXQ R13, R12
    MULXQ AX, AX, R13
    ADOXQ AX, R12
    // add the last carries to R13
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R13
    ADOXQ DX, R13
    PUSHQ R13
    MOVQ R14, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R13
    ADCXQ R14, AX
    MOVQ R13, R14
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R15, R14
    MULXQ ·qElement+8(SB), AX, R15
    ADOXQ AX, R14
    ADCXQ CX, R15
    MULXQ ·qElement+16(SB), AX, CX
    ADOXQ AX, R15
    ADCXQ BX, CX
    MULXQ ·qElement+24(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+32(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+40(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+48(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+56(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+64(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+72(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+80(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+88(SB), AX, R12
    ADOXQ AX, R11
    POPQ R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ R13, R12
    XORQ DX, DX
    MOVQ y+16(FP), DX
    MOVQ 72(DX), DX
    MOVQ x+8(FP), AX
    MOVQ 0(AX), AX
    MULXQ AX, AX, R13
    ADOXQ AX, R14
    MOVQ x+8(FP), AX
    MOVQ 8(AX), AX
    ADCXQ R13, R15
    MULXQ AX, AX, R13
    ADOXQ AX, R15
    MOVQ x+8(FP), AX
    MOVQ 16(AX), AX
    ADCXQ R13, CX
    MULXQ AX, AX, R13
    ADOXQ AX, CX
    MOVQ x+8(FP), AX
    MOVQ 24(AX), AX
    ADCXQ R13, BX
    MULXQ AX, AX, R13
    ADOXQ AX, BX
    MOVQ x+8(FP), AX
    MOVQ 32(AX), AX
    ADCXQ R13, BP
    MULXQ AX, AX, R13
    ADOXQ AX, BP
    MOVQ x+8(FP), AX
    MOVQ 40(AX), AX
    ADCXQ R13, SI
    MULXQ AX, AX, R13
    ADOXQ AX, SI
    MOVQ x+8(FP), AX
    MOVQ 48(AX), AX
    ADCXQ R13, DI
    MULXQ AX, AX, R13
    ADOXQ AX, DI
    MOVQ x+8(FP), AX
    MOVQ 56(AX), AX
    ADCXQ R13, R8
    MULXQ AX, AX, R13
    ADOXQ AX, R8
    MOVQ x+8(FP), AX
    MOVQ 64(AX), AX
    ADCXQ R13, R9
    MULXQ AX, AX, R13
    ADOXQ AX, R9
    MOVQ x+8(FP), AX
    MOVQ 72(AX), AX
    ADCXQ R13, R10
    MULXQ AX, AX, R13
    ADOXQ AX, R10
    MOVQ x+8(FP), AX
    MOVQ 80(AX), AX
    ADCXQ R13, R11
    MULXQ AX, AX, R13
    ADOXQ AX, R11
    MOVQ x+8(FP), AX
    MOVQ 88(AX), AX
    ADCXQ R13, R12
    MULXQ AX, AX, R13
    ADOXQ AX, R12
    // add the last carries to R13
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R13
    ADOXQ DX, R13
    PUSHQ R13
    MOVQ R14, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R13
    ADCXQ R14, AX
    MOVQ R13, R14
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R15, R14
    MULXQ ·qElement+8(SB), AX, R15
    ADOXQ AX, R14
    ADCXQ CX, R15
    MULXQ ·qElement+16(SB), AX, CX
    ADOXQ AX, R15
    ADCXQ BX, CX
    MULXQ ·qElement+24(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+32(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+40(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+48(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+56(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+64(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+72(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+80(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+88(SB), AX, R12
    ADOXQ AX, R11
    POPQ R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ R13, R12
    XORQ DX, DX
    MOVQ y+16(FP), DX
    MOVQ 80(DX), DX
    MOVQ x+8(FP), AX
    MOVQ 0(AX), AX
    MULXQ AX, AX, R13
    ADOXQ AX, R14
    MOVQ x+8(FP), AX
    MOVQ 8(AX), AX
    ADCXQ R13, R15
    MULXQ AX, AX, R13
    ADOXQ AX, R15
    MOVQ x+8(FP), AX
    MOVQ 16(AX), AX
    ADCXQ R13, CX
    MULXQ AX, AX, R13
    ADOXQ AX, CX
    MOVQ x+8(FP), AX
    MOVQ 24(AX), AX
    ADCXQ R13, BX
    MULXQ AX, AX, R13
    ADOXQ AX, BX
    MOVQ x+8(FP), AX
    MOVQ 32(AX), AX
    ADCXQ R13, BP
    MULXQ AX, AX, R13
    ADOXQ AX, BP
    MOVQ x+8(FP), AX
    MOVQ 40(AX), AX
    ADCXQ R13, SI
    MULXQ AX, AX, R13
    ADOXQ AX, SI
    MOVQ x+8(FP), AX
    MOVQ 48(AX), AX
    ADCXQ R13, DI
    MULXQ AX, AX, R13
    ADOXQ AX, DI
    MOVQ x+8(FP), AX
    MOVQ 56(AX), AX
    ADCXQ R13, R8
    MULXQ AX, AX, R13
    ADOXQ AX, R8
    MOVQ x+8(FP), AX
    MOVQ 64(AX), AX
    ADCXQ R13, R9
    MULXQ AX, AX, R13
    ADOXQ AX, R9
    MOVQ x+8(FP), AX
    MOVQ 72(AX), AX
    ADCXQ R13, R10
    MULXQ AX, AX, R13
    ADOXQ AX, R10
    MOVQ x+8(FP), AX
    MOVQ 80(AX), AX
    ADCXQ R13, R11
    MULXQ AX, AX, R13
    ADOXQ AX, R11
    MOVQ x+8(FP), AX
    MOVQ 88(AX), AX
    ADCXQ R13, R12
    MULXQ AX, AX, R13
    ADOXQ AX, R12
    // add the last carries to R13
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R13
    ADOXQ DX, R13
    PUSHQ R13
    MOVQ R14, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R13
    ADCXQ R14, AX
    MOVQ R13, R14
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R15, R14
    MULXQ ·qElement+8(SB), AX, R15
    ADOXQ AX, R14
    ADCXQ CX, R15
    MULXQ ·qElement+16(SB), AX, CX
    ADOXQ AX, R15
    ADCXQ BX, CX
    MULXQ ·qElement+24(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+32(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+40(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+48(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+56(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+64(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+72(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+80(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+88(SB), AX, R12
    ADOXQ AX, R11
    POPQ R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ R13, R12
    XORQ DX, DX
    MOVQ y+16(FP), DX
    MOVQ 88(DX), DX
    MOVQ x+8(FP), AX
    MOVQ 0(AX), AX
    MULXQ AX, AX, R13
    ADOXQ AX, R14
    MOVQ x+8(FP), AX
    MOVQ 8(AX), AX
    ADCXQ R13, R15
    MULXQ AX, AX, R13
    ADOXQ AX, R15
    MOVQ x+8(FP), AX
    MOVQ 16(AX), AX
    ADCXQ R13, CX
    MULXQ AX, AX, R13
    ADOXQ AX, CX
    MOVQ x+8(FP), AX
    MOVQ 24(AX), AX
    ADCXQ R13, BX
    MULXQ AX, AX, R13
    ADOXQ AX, BX
    MOVQ x+8(FP), AX
    MOVQ 32(AX), AX
    ADCXQ R13, BP
    MULXQ AX, AX, R13
    ADOXQ AX, BP
    MOVQ x+8(FP), AX
    MOVQ 40(AX), AX
    ADCXQ R13, SI
    MULXQ AX, AX, R13
    ADOXQ AX, SI
    MOVQ x+8(FP), AX
    MOVQ 48(AX), AX
    ADCXQ R13, DI
    MULXQ AX, AX, R13
    ADOXQ AX, DI
    MOVQ x+8(FP), AX
    MOVQ 56(AX), AX
    ADCXQ R13, R8
    MULXQ AX, AX, R13
    ADOXQ AX, R8
    MOVQ x+8(FP), AX
    MOVQ 64(AX), AX
    ADCXQ R13, R9
    MULXQ AX, AX, R13
    ADOXQ AX, R9
    MOVQ x+8(FP), AX
    MOVQ 72(AX), AX
    ADCXQ R13, R10
    MULXQ AX, AX, R13
    ADOXQ AX, R10
    MOVQ x+8(FP), AX
    MOVQ 80(AX), AX
    ADCXQ R13, R11
    MULXQ AX, AX, R13
    ADOXQ AX, R11
    MOVQ x+8(FP), AX
    MOVQ 88(AX), AX
    ADCXQ R13, R12
    MULXQ AX, AX, R13
    ADOXQ AX, R12
    // add the last carries to R13
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R13
    ADOXQ DX, R13
    PUSHQ R13
    MOVQ R14, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R13
    ADCXQ R14, AX
    MOVQ R13, R14
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R15, R14
    MULXQ ·qElement+8(SB), AX, R15
    ADOXQ AX, R14
    ADCXQ CX, R15
    MULXQ ·qElement+16(SB), AX, CX
    ADOXQ AX, R15
    ADCXQ BX, CX
    MULXQ ·qElement+24(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+32(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+40(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+48(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+56(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+64(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+72(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+80(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+88(SB), AX, R12
    ADOXQ AX, R11
    POPQ R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ R13, R12
    MOVQ res+0(FP), R13
    MOVQ R14, t0-8(SP)
    SUBQ ·qElement+0(SB), R14
    MOVQ R15, t1-16(SP)
    SBBQ ·qElement+8(SB), R15
    MOVQ CX, t2-24(SP)
    SBBQ ·qElement+16(SB), CX
    MOVQ BX, t3-32(SP)
    SBBQ ·qElement+24(SB), BX
    MOVQ BP, t4-40(SP)
    SBBQ ·qElement+32(SB), BP
    MOVQ SI, t5-48(SP)
    SBBQ ·qElement+40(SB), SI
    MOVQ DI, t6-56(SP)
    SBBQ ·qElement+48(SB), DI
    MOVQ R8, t7-64(SP)
    SBBQ ·qElement+56(SB), R8
    MOVQ R9, t8-72(SP)
    SBBQ ·qElement+64(SB), R9
    MOVQ R10, t9-80(SP)
    SBBQ ·qElement+72(SB), R10
    MOVQ R11, t10-88(SP)
    SBBQ ·qElement+80(SB), R11
    MOVQ R12, t11-96(SP)
    SBBQ ·qElement+88(SB), R12
    CMOVQCS t0-8(SP), R14
    CMOVQCS t1-16(SP), R15
    CMOVQCS t2-24(SP), CX
    CMOVQCS t3-32(SP), BX
    CMOVQCS t4-40(SP), BP
    CMOVQCS t5-48(SP), SI
    CMOVQCS t6-56(SP), DI
    CMOVQCS t7-64(SP), R8
    CMOVQCS t8-72(SP), R9
    CMOVQCS t9-80(SP), R10
    CMOVQCS t10-88(SP), R11
    CMOVQCS t11-96(SP), R12
    MOVQ R14, 0(R13)
    MOVQ R15, 8(R13)
    MOVQ CX, 16(R13)
    MOVQ BX, 24(R13)
    MOVQ BP, 32(R13)
    MOVQ SI, 40(R13)
    MOVQ DI, 48(R13)
    MOVQ R8, 56(R13)
    MOVQ R9, 64(R13)
    MOVQ R10, 72(R13)
    MOVQ R11, 80(R13)
    MOVQ R12, 88(R13)
    RET
l1:
    MOVQ res+0(FP), AX
    MOVQ AX, (SP)
    MOVQ x+8(FP), AX
    MOVQ AX, 8(SP)
    MOVQ y+16(FP), AX
    MOVQ AX, 16(SP)
CALL ·_mulGeneric(SB)
    RET

TEXT ·square(SB), $96-16

	// the algorithm is described here
	// https://hackmd.io/@zkteam/modular_multiplication
	// for i=0 to N-1
	// A, t[i] = x[i] * x[i] + t[i]
	// p = 0
	// for j=i+1 to N-1
	//     p,A,t[j] = 2*x[j]*x[i] + t[j] + (p,A)
	// m = t[0] * q'[0]
	// C, _ = t[0] + q[0]*m
	// for j=1 to N-1
	//     C, t[j-1] = q[j]*m +  t[j] + C
	// t[N-1] = C + A

	
NO_LOCAL_POINTERS
    CMPB ·supportAdx(SB), $0x0000000000000001
    JNE l2
    XORQ DX, DX
    MOVQ x+8(FP), DX
    MOVQ 0(DX), DX
    MOVQ x+8(FP), AX
    MOVQ 0(AX), AX
    MULXQ AX, R14, R15
    MOVQ x+8(FP), AX
    MOVQ 8(AX), AX
    MULXQ AX, AX, CX
    ADOXQ AX, R15
    MOVQ x+8(FP), AX
    MOVQ 16(AX), AX
    MULXQ AX, AX, BX
    ADOXQ AX, CX
    MOVQ x+8(FP), AX
    MOVQ 24(AX), AX
    MULXQ AX, AX, BP
    ADOXQ AX, BX
    MOVQ x+8(FP), AX
    MOVQ 32(AX), AX
    MULXQ AX, AX, SI
    ADOXQ AX, BP
    MOVQ x+8(FP), AX
    MOVQ 40(AX), AX
    MULXQ AX, AX, DI
    ADOXQ AX, SI
    MOVQ x+8(FP), AX
    MOVQ 48(AX), AX
    MULXQ AX, AX, R8
    ADOXQ AX, DI
    MOVQ x+8(FP), AX
    MOVQ 56(AX), AX
    MULXQ AX, AX, R9
    ADOXQ AX, R8
    MOVQ x+8(FP), AX
    MOVQ 64(AX), AX
    MULXQ AX, AX, R10
    ADOXQ AX, R9
    MOVQ x+8(FP), AX
    MOVQ 72(AX), AX
    MULXQ AX, AX, R11
    ADOXQ AX, R10
    MOVQ x+8(FP), AX
    MOVQ 80(AX), AX
    MULXQ AX, AX, R12
    ADOXQ AX, R11
    MOVQ x+8(FP), AX
    MOVQ 88(AX), AX
    MULXQ AX, AX, R13
    ADOXQ AX, R12
    // add the last carries to R13
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R13
    ADOXQ DX, R13
    PUSHQ R13
    MOVQ R14, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R13
    ADCXQ R14, AX
    MOVQ R13, R14
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R15, R14
    MULXQ ·qElement+8(SB), AX, R15
    ADOXQ AX, R14
    ADCXQ CX, R15
    MULXQ ·qElement+16(SB), AX, CX
    ADOXQ AX, R15
    ADCXQ BX, CX
    MULXQ ·qElement+24(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+32(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+40(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+48(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+56(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+64(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+72(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+80(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+88(SB), AX, R12
    ADOXQ AX, R11
    POPQ R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ R13, R12
    XORQ DX, DX
    MOVQ x+8(FP), DX
    MOVQ 8(DX), DX
    MOVQ x+8(FP), AX
    MOVQ 0(AX), AX
    MULXQ AX, AX, R13
    ADOXQ AX, R14
    MOVQ x+8(FP), AX
    MOVQ 8(AX), AX
    ADCXQ R13, R15
    MULXQ AX, AX, R13
    ADOXQ AX, R15
    MOVQ x+8(FP), AX
    MOVQ 16(AX), AX
    ADCXQ R13, CX
    MULXQ AX, AX, R13
    ADOXQ AX, CX
    MOVQ x+8(FP), AX
    MOVQ 24(AX), AX
    ADCXQ R13, BX
    MULXQ AX, AX, R13
    ADOXQ AX, BX
    MOVQ x+8(FP), AX
    MOVQ 32(AX), AX
    ADCXQ R13, BP
    MULXQ AX, AX, R13
    ADOXQ AX, BP
    MOVQ x+8(FP), AX
    MOVQ 40(AX), AX
    ADCXQ R13, SI
    MULXQ AX, AX, R13
    ADOXQ AX, SI
    MOVQ x+8(FP), AX
    MOVQ 48(AX), AX
    ADCXQ R13, DI
    MULXQ AX, AX, R13
    ADOXQ AX, DI
    MOVQ x+8(FP), AX
    MOVQ 56(AX), AX
    ADCXQ R13, R8
    MULXQ AX, AX, R13
    ADOXQ AX, R8
    MOVQ x+8(FP), AX
    MOVQ 64(AX), AX
    ADCXQ R13, R9
    MULXQ AX, AX, R13
    ADOXQ AX, R9
    MOVQ x+8(FP), AX
    MOVQ 72(AX), AX
    ADCXQ R13, R10
    MULXQ AX, AX, R13
    ADOXQ AX, R10
    MOVQ x+8(FP), AX
    MOVQ 80(AX), AX
    ADCXQ R13, R11
    MULXQ AX, AX, R13
    ADOXQ AX, R11
    MOVQ x+8(FP), AX
    MOVQ 88(AX), AX
    ADCXQ R13, R12
    MULXQ AX, AX, R13
    ADOXQ AX, R12
    // add the last carries to R13
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R13
    ADOXQ DX, R13
    PUSHQ R13
    MOVQ R14, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R13
    ADCXQ R14, AX
    MOVQ R13, R14
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R15, R14
    MULXQ ·qElement+8(SB), AX, R15
    ADOXQ AX, R14
    ADCXQ CX, R15
    MULXQ ·qElement+16(SB), AX, CX
    ADOXQ AX, R15
    ADCXQ BX, CX
    MULXQ ·qElement+24(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+32(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+40(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+48(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+56(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+64(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+72(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+80(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+88(SB), AX, R12
    ADOXQ AX, R11
    POPQ R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ R13, R12
    XORQ DX, DX
    MOVQ x+8(FP), DX
    MOVQ 16(DX), DX
    MOVQ x+8(FP), AX
    MOVQ 0(AX), AX
    MULXQ AX, AX, R13
    ADOXQ AX, R14
    MOVQ x+8(FP), AX
    MOVQ 8(AX), AX
    ADCXQ R13, R15
    MULXQ AX, AX, R13
    ADOXQ AX, R15
    MOVQ x+8(FP), AX
    MOVQ 16(AX), AX
    ADCXQ R13, CX
    MULXQ AX, AX, R13
    ADOXQ AX, CX
    MOVQ x+8(FP), AX
    MOVQ 24(AX), AX
    ADCXQ R13, BX
    MULXQ AX, AX, R13
    ADOXQ AX, BX
    MOVQ x+8(FP), AX
    MOVQ 32(AX), AX
    ADCXQ R13, BP
    MULXQ AX, AX, R13
    ADOXQ AX, BP
    MOVQ x+8(FP), AX
    MOVQ 40(AX), AX
    ADCXQ R13, SI
    MULXQ AX, AX, R13
    ADOXQ AX, SI
    MOVQ x+8(FP), AX
    MOVQ 48(AX), AX
    ADCXQ R13, DI
    MULXQ AX, AX, R13
    ADOXQ AX, DI
    MOVQ x+8(FP), AX
    MOVQ 56(AX), AX
    ADCXQ R13, R8
    MULXQ AX, AX, R13
    ADOXQ AX, R8
    MOVQ x+8(FP), AX
    MOVQ 64(AX), AX
    ADCXQ R13, R9
    MULXQ AX, AX, R13
    ADOXQ AX, R9
    MOVQ x+8(FP), AX
    MOVQ 72(AX), AX
    ADCXQ R13, R10
    MULXQ AX, AX, R13
    ADOXQ AX, R10
    MOVQ x+8(FP), AX
    MOVQ 80(AX), AX
    ADCXQ R13, R11
    MULXQ AX, AX, R13
    ADOXQ AX, R11
    MOVQ x+8(FP), AX
    MOVQ 88(AX), AX
    ADCXQ R13, R12
    MULXQ AX, AX, R13
    ADOXQ AX, R12
    // add the last carries to R13
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R13
    ADOXQ DX, R13
    PUSHQ R13
    MOVQ R14, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R13
    ADCXQ R14, AX
    MOVQ R13, R14
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R15, R14
    MULXQ ·qElement+8(SB), AX, R15
    ADOXQ AX, R14
    ADCXQ CX, R15
    MULXQ ·qElement+16(SB), AX, CX
    ADOXQ AX, R15
    ADCXQ BX, CX
    MULXQ ·qElement+24(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+32(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+40(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+48(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+56(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+64(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+72(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+80(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+88(SB), AX, R12
    ADOXQ AX, R11
    POPQ R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ R13, R12
    XORQ DX, DX
    MOVQ x+8(FP), DX
    MOVQ 24(DX), DX
    MOVQ x+8(FP), AX
    MOVQ 0(AX), AX
    MULXQ AX, AX, R13
    ADOXQ AX, R14
    MOVQ x+8(FP), AX
    MOVQ 8(AX), AX
    ADCXQ R13, R15
    MULXQ AX, AX, R13
    ADOXQ AX, R15
    MOVQ x+8(FP), AX
    MOVQ 16(AX), AX
    ADCXQ R13, CX
    MULXQ AX, AX, R13
    ADOXQ AX, CX
    MOVQ x+8(FP), AX
    MOVQ 24(AX), AX
    ADCXQ R13, BX
    MULXQ AX, AX, R13
    ADOXQ AX, BX
    MOVQ x+8(FP), AX
    MOVQ 32(AX), AX
    ADCXQ R13, BP
    MULXQ AX, AX, R13
    ADOXQ AX, BP
    MOVQ x+8(FP), AX
    MOVQ 40(AX), AX
    ADCXQ R13, SI
    MULXQ AX, AX, R13
    ADOXQ AX, SI
    MOVQ x+8(FP), AX
    MOVQ 48(AX), AX
    ADCXQ R13, DI
    MULXQ AX, AX, R13
    ADOXQ AX, DI
    MOVQ x+8(FP), AX
    MOVQ 56(AX), AX
    ADCXQ R13, R8
    MULXQ AX, AX, R13
    ADOXQ AX, R8
    MOVQ x+8(FP), AX
    MOVQ 64(AX), AX
    ADCXQ R13, R9
    MULXQ AX, AX, R13
    ADOXQ AX, R9
    MOVQ x+8(FP), AX
    MOVQ 72(AX), AX
    ADCXQ R13, R10
    MULXQ AX, AX, R13
    ADOXQ AX, R10
    MOVQ x+8(FP), AX
    MOVQ 80(AX), AX
    ADCXQ R13, R11
    MULXQ AX, AX, R13
    ADOXQ AX, R11
    MOVQ x+8(FP), AX
    MOVQ 88(AX), AX
    ADCXQ R13, R12
    MULXQ AX, AX, R13
    ADOXQ AX, R12
    // add the last carries to R13
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R13
    ADOXQ DX, R13
    PUSHQ R13
    MOVQ R14, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R13
    ADCXQ R14, AX
    MOVQ R13, R14
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R15, R14
    MULXQ ·qElement+8(SB), AX, R15
    ADOXQ AX, R14
    ADCXQ CX, R15
    MULXQ ·qElement+16(SB), AX, CX
    ADOXQ AX, R15
    ADCXQ BX, CX
    MULXQ ·qElement+24(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+32(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+40(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+48(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+56(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+64(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+72(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+80(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+88(SB), AX, R12
    ADOXQ AX, R11
    POPQ R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ R13, R12
    XORQ DX, DX
    MOVQ x+8(FP), DX
    MOVQ 32(DX), DX
    MOVQ x+8(FP), AX
    MOVQ 0(AX), AX
    MULXQ AX, AX, R13
    ADOXQ AX, R14
    MOVQ x+8(FP), AX
    MOVQ 8(AX), AX
    ADCXQ R13, R15
    MULXQ AX, AX, R13
    ADOXQ AX, R15
    MOVQ x+8(FP), AX
    MOVQ 16(AX), AX
    ADCXQ R13, CX
    MULXQ AX, AX, R13
    ADOXQ AX, CX
    MOVQ x+8(FP), AX
    MOVQ 24(AX), AX
    ADCXQ R13, BX
    MULXQ AX, AX, R13
    ADOXQ AX, BX
    MOVQ x+8(FP), AX
    MOVQ 32(AX), AX
    ADCXQ R13, BP
    MULXQ AX, AX, R13
    ADOXQ AX, BP
    MOVQ x+8(FP), AX
    MOVQ 40(AX), AX
    ADCXQ R13, SI
    MULXQ AX, AX, R13
    ADOXQ AX, SI
    MOVQ x+8(FP), AX
    MOVQ 48(AX), AX
    ADCXQ R13, DI
    MULXQ AX, AX, R13
    ADOXQ AX, DI
    MOVQ x+8(FP), AX
    MOVQ 56(AX), AX
    ADCXQ R13, R8
    MULXQ AX, AX, R13
    ADOXQ AX, R8
    MOVQ x+8(FP), AX
    MOVQ 64(AX), AX
    ADCXQ R13, R9
    MULXQ AX, AX, R13
    ADOXQ AX, R9
    MOVQ x+8(FP), AX
    MOVQ 72(AX), AX
    ADCXQ R13, R10
    MULXQ AX, AX, R13
    ADOXQ AX, R10
    MOVQ x+8(FP), AX
    MOVQ 80(AX), AX
    ADCXQ R13, R11
    MULXQ AX, AX, R13
    ADOXQ AX, R11
    MOVQ x+8(FP), AX
    MOVQ 88(AX), AX
    ADCXQ R13, R12
    MULXQ AX, AX, R13
    ADOXQ AX, R12
    // add the last carries to R13
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R13
    ADOXQ DX, R13
    PUSHQ R13
    MOVQ R14, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R13
    ADCXQ R14, AX
    MOVQ R13, R14
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R15, R14
    MULXQ ·qElement+8(SB), AX, R15
    ADOXQ AX, R14
    ADCXQ CX, R15
    MULXQ ·qElement+16(SB), AX, CX
    ADOXQ AX, R15
    ADCXQ BX, CX
    MULXQ ·qElement+24(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+32(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+40(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+48(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+56(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+64(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+72(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+80(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+88(SB), AX, R12
    ADOXQ AX, R11
    POPQ R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ R13, R12
    XORQ DX, DX
    MOVQ x+8(FP), DX
    MOVQ 40(DX), DX
    MOVQ x+8(FP), AX
    MOVQ 0(AX), AX
    MULXQ AX, AX, R13
    ADOXQ AX, R14
    MOVQ x+8(FP), AX
    MOVQ 8(AX), AX
    ADCXQ R13, R15
    MULXQ AX, AX, R13
    ADOXQ AX, R15
    MOVQ x+8(FP), AX
    MOVQ 16(AX), AX
    ADCXQ R13, CX
    MULXQ AX, AX, R13
    ADOXQ AX, CX
    MOVQ x+8(FP), AX
    MOVQ 24(AX), AX
    ADCXQ R13, BX
    MULXQ AX, AX, R13
    ADOXQ AX, BX
    MOVQ x+8(FP), AX
    MOVQ 32(AX), AX
    ADCXQ R13, BP
    MULXQ AX, AX, R13
    ADOXQ AX, BP
    MOVQ x+8(FP), AX
    MOVQ 40(AX), AX
    ADCXQ R13, SI
    MULXQ AX, AX, R13
    ADOXQ AX, SI
    MOVQ x+8(FP), AX
    MOVQ 48(AX), AX
    ADCXQ R13, DI
    MULXQ AX, AX, R13
    ADOXQ AX, DI
    MOVQ x+8(FP), AX
    MOVQ 56(AX), AX
    ADCXQ R13, R8
    MULXQ AX, AX, R13
    ADOXQ AX, R8
    MOVQ x+8(FP), AX
    MOVQ 64(AX), AX
    ADCXQ R13, R9
    MULXQ AX, AX, R13
    ADOXQ AX, R9
    MOVQ x+8(FP), AX
    MOVQ 72(AX), AX
    ADCXQ R13, R10
    MULXQ AX, AX, R13
    ADOXQ AX, R10
    MOVQ x+8(FP), AX
    MOVQ 80(AX), AX
    ADCXQ R13, R11
    MULXQ AX, AX, R13
    ADOXQ AX, R11
    MOVQ x+8(FP), AX
    MOVQ 88(AX), AX
    ADCXQ R13, R12
    MULXQ AX, AX, R13
    ADOXQ AX, R12
    // add the last carries to R13
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R13
    ADOXQ DX, R13
    PUSHQ R13
    MOVQ R14, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R13
    ADCXQ R14, AX
    MOVQ R13, R14
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R15, R14
    MULXQ ·qElement+8(SB), AX, R15
    ADOXQ AX, R14
    ADCXQ CX, R15
    MULXQ ·qElement+16(SB), AX, CX
    ADOXQ AX, R15
    ADCXQ BX, CX
    MULXQ ·qElement+24(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+32(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+40(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+48(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+56(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+64(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+72(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+80(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+88(SB), AX, R12
    ADOXQ AX, R11
    POPQ R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ R13, R12
    XORQ DX, DX
    MOVQ x+8(FP), DX
    MOVQ 48(DX), DX
    MOVQ x+8(FP), AX
    MOVQ 0(AX), AX
    MULXQ AX, AX, R13
    ADOXQ AX, R14
    MOVQ x+8(FP), AX
    MOVQ 8(AX), AX
    ADCXQ R13, R15
    MULXQ AX, AX, R13
    ADOXQ AX, R15
    MOVQ x+8(FP), AX
    MOVQ 16(AX), AX
    ADCXQ R13, CX
    MULXQ AX, AX, R13
    ADOXQ AX, CX
    MOVQ x+8(FP), AX
    MOVQ 24(AX), AX
    ADCXQ R13, BX
    MULXQ AX, AX, R13
    ADOXQ AX, BX
    MOVQ x+8(FP), AX
    MOVQ 32(AX), AX
    ADCXQ R13, BP
    MULXQ AX, AX, R13
    ADOXQ AX, BP
    MOVQ x+8(FP), AX
    MOVQ 40(AX), AX
    ADCXQ R13, SI
    MULXQ AX, AX, R13
    ADOXQ AX, SI
    MOVQ x+8(FP), AX
    MOVQ 48(AX), AX
    ADCXQ R13, DI
    MULXQ AX, AX, R13
    ADOXQ AX, DI
    MOVQ x+8(FP), AX
    MOVQ 56(AX), AX
    ADCXQ R13, R8
    MULXQ AX, AX, R13
    ADOXQ AX, R8
    MOVQ x+8(FP), AX
    MOVQ 64(AX), AX
    ADCXQ R13, R9
    MULXQ AX, AX, R13
    ADOXQ AX, R9
    MOVQ x+8(FP), AX
    MOVQ 72(AX), AX
    ADCXQ R13, R10
    MULXQ AX, AX, R13
    ADOXQ AX, R10
    MOVQ x+8(FP), AX
    MOVQ 80(AX), AX
    ADCXQ R13, R11
    MULXQ AX, AX, R13
    ADOXQ AX, R11
    MOVQ x+8(FP), AX
    MOVQ 88(AX), AX
    ADCXQ R13, R12
    MULXQ AX, AX, R13
    ADOXQ AX, R12
    // add the last carries to R13
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R13
    ADOXQ DX, R13
    PUSHQ R13
    MOVQ R14, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R13
    ADCXQ R14, AX
    MOVQ R13, R14
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R15, R14
    MULXQ ·qElement+8(SB), AX, R15
    ADOXQ AX, R14
    ADCXQ CX, R15
    MULXQ ·qElement+16(SB), AX, CX
    ADOXQ AX, R15
    ADCXQ BX, CX
    MULXQ ·qElement+24(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+32(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+40(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+48(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+56(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+64(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+72(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+80(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+88(SB), AX, R12
    ADOXQ AX, R11
    POPQ R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ R13, R12
    XORQ DX, DX
    MOVQ x+8(FP), DX
    MOVQ 56(DX), DX
    MOVQ x+8(FP), AX
    MOVQ 0(AX), AX
    MULXQ AX, AX, R13
    ADOXQ AX, R14
    MOVQ x+8(FP), AX
    MOVQ 8(AX), AX
    ADCXQ R13, R15
    MULXQ AX, AX, R13
    ADOXQ AX, R15
    MOVQ x+8(FP), AX
    MOVQ 16(AX), AX
    ADCXQ R13, CX
    MULXQ AX, AX, R13
    ADOXQ AX, CX
    MOVQ x+8(FP), AX
    MOVQ 24(AX), AX
    ADCXQ R13, BX
    MULXQ AX, AX, R13
    ADOXQ AX, BX
    MOVQ x+8(FP), AX
    MOVQ 32(AX), AX
    ADCXQ R13, BP
    MULXQ AX, AX, R13
    ADOXQ AX, BP
    MOVQ x+8(FP), AX
    MOVQ 40(AX), AX
    ADCXQ R13, SI
    MULXQ AX, AX, R13
    ADOXQ AX, SI
    MOVQ x+8(FP), AX
    MOVQ 48(AX), AX
    ADCXQ R13, DI
    MULXQ AX, AX, R13
    ADOXQ AX, DI
    MOVQ x+8(FP), AX
    MOVQ 56(AX), AX
    ADCXQ R13, R8
    MULXQ AX, AX, R13
    ADOXQ AX, R8
    MOVQ x+8(FP), AX
    MOVQ 64(AX), AX
    ADCXQ R13, R9
    MULXQ AX, AX, R13
    ADOXQ AX, R9
    MOVQ x+8(FP), AX
    MOVQ 72(AX), AX
    ADCXQ R13, R10
    MULXQ AX, AX, R13
    ADOXQ AX, R10
    MOVQ x+8(FP), AX
    MOVQ 80(AX), AX
    ADCXQ R13, R11
    MULXQ AX, AX, R13
    ADOXQ AX, R11
    MOVQ x+8(FP), AX
    MOVQ 88(AX), AX
    ADCXQ R13, R12
    MULXQ AX, AX, R13
    ADOXQ AX, R12
    // add the last carries to R13
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R13
    ADOXQ DX, R13
    PUSHQ R13
    MOVQ R14, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R13
    ADCXQ R14, AX
    MOVQ R13, R14
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R15, R14
    MULXQ ·qElement+8(SB), AX, R15
    ADOXQ AX, R14
    ADCXQ CX, R15
    MULXQ ·qElement+16(SB), AX, CX
    ADOXQ AX, R15
    ADCXQ BX, CX
    MULXQ ·qElement+24(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+32(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+40(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+48(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+56(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+64(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+72(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+80(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+88(SB), AX, R12
    ADOXQ AX, R11
    POPQ R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ R13, R12
    XORQ DX, DX
    MOVQ x+8(FP), DX
    MOVQ 64(DX), DX
    MOVQ x+8(FP), AX
    MOVQ 0(AX), AX
    MULXQ AX, AX, R13
    ADOXQ AX, R14
    MOVQ x+8(FP), AX
    MOVQ 8(AX), AX
    ADCXQ R13, R15
    MULXQ AX, AX, R13
    ADOXQ AX, R15
    MOVQ x+8(FP), AX
    MOVQ 16(AX), AX
    ADCXQ R13, CX
    MULXQ AX, AX, R13
    ADOXQ AX, CX
    MOVQ x+8(FP), AX
    MOVQ 24(AX), AX
    ADCXQ R13, BX
    MULXQ AX, AX, R13
    ADOXQ AX, BX
    MOVQ x+8(FP), AX
    MOVQ 32(AX), AX
    ADCXQ R13, BP
    MULXQ AX, AX, R13
    ADOXQ AX, BP
    MOVQ x+8(FP), AX
    MOVQ 40(AX), AX
    ADCXQ R13, SI
    MULXQ AX, AX, R13
    ADOXQ AX, SI
    MOVQ x+8(FP), AX
    MOVQ 48(AX), AX
    ADCXQ R13, DI
    MULXQ AX, AX, R13
    ADOXQ AX, DI
    MOVQ x+8(FP), AX
    MOVQ 56(AX), AX
    ADCXQ R13, R8
    MULXQ AX, AX, R13
    ADOXQ AX, R8
    MOVQ x+8(FP), AX
    MOVQ 64(AX), AX
    ADCXQ R13, R9
    MULXQ AX, AX, R13
    ADOXQ AX, R9
    MOVQ x+8(FP), AX
    MOVQ 72(AX), AX
    ADCXQ R13, R10
    MULXQ AX, AX, R13
    ADOXQ AX, R10
    MOVQ x+8(FP), AX
    MOVQ 80(AX), AX
    ADCXQ R13, R11
    MULXQ AX, AX, R13
    ADOXQ AX, R11
    MOVQ x+8(FP), AX
    MOVQ 88(AX), AX
    ADCXQ R13, R12
    MULXQ AX, AX, R13
    ADOXQ AX, R12
    // add the last carries to R13
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R13
    ADOXQ DX, R13
    PUSHQ R13
    MOVQ R14, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R13
    ADCXQ R14, AX
    MOVQ R13, R14
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R15, R14
    MULXQ ·qElement+8(SB), AX, R15
    ADOXQ AX, R14
    ADCXQ CX, R15
    MULXQ ·qElement+16(SB), AX, CX
    ADOXQ AX, R15
    ADCXQ BX, CX
    MULXQ ·qElement+24(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+32(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+40(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+48(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+56(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+64(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+72(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+80(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+88(SB), AX, R12
    ADOXQ AX, R11
    POPQ R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ R13, R12
    XORQ DX, DX
    MOVQ x+8(FP), DX
    MOVQ 72(DX), DX
    MOVQ x+8(FP), AX
    MOVQ 0(AX), AX
    MULXQ AX, AX, R13
    ADOXQ AX, R14
    MOVQ x+8(FP), AX
    MOVQ 8(AX), AX
    ADCXQ R13, R15
    MULXQ AX, AX, R13
    ADOXQ AX, R15
    MOVQ x+8(FP), AX
    MOVQ 16(AX), AX
    ADCXQ R13, CX
    MULXQ AX, AX, R13
    ADOXQ AX, CX
    MOVQ x+8(FP), AX
    MOVQ 24(AX), AX
    ADCXQ R13, BX
    MULXQ AX, AX, R13
    ADOXQ AX, BX
    MOVQ x+8(FP), AX
    MOVQ 32(AX), AX
    ADCXQ R13, BP
    MULXQ AX, AX, R13
    ADOXQ AX, BP
    MOVQ x+8(FP), AX
    MOVQ 40(AX), AX
    ADCXQ R13, SI
    MULXQ AX, AX, R13
    ADOXQ AX, SI
    MOVQ x+8(FP), AX
    MOVQ 48(AX), AX
    ADCXQ R13, DI
    MULXQ AX, AX, R13
    ADOXQ AX, DI
    MOVQ x+8(FP), AX
    MOVQ 56(AX), AX
    ADCXQ R13, R8
    MULXQ AX, AX, R13
    ADOXQ AX, R8
    MOVQ x+8(FP), AX
    MOVQ 64(AX), AX
    ADCXQ R13, R9
    MULXQ AX, AX, R13
    ADOXQ AX, R9
    MOVQ x+8(FP), AX
    MOVQ 72(AX), AX
    ADCXQ R13, R10
    MULXQ AX, AX, R13
    ADOXQ AX, R10
    MOVQ x+8(FP), AX
    MOVQ 80(AX), AX
    ADCXQ R13, R11
    MULXQ AX, AX, R13
    ADOXQ AX, R11
    MOVQ x+8(FP), AX
    MOVQ 88(AX), AX
    ADCXQ R13, R12
    MULXQ AX, AX, R13
    ADOXQ AX, R12
    // add the last carries to R13
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R13
    ADOXQ DX, R13
    PUSHQ R13
    MOVQ R14, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R13
    ADCXQ R14, AX
    MOVQ R13, R14
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R15, R14
    MULXQ ·qElement+8(SB), AX, R15
    ADOXQ AX, R14
    ADCXQ CX, R15
    MULXQ ·qElement+16(SB), AX, CX
    ADOXQ AX, R15
    ADCXQ BX, CX
    MULXQ ·qElement+24(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+32(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+40(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+48(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+56(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+64(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+72(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+80(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+88(SB), AX, R12
    ADOXQ AX, R11
    POPQ R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ R13, R12
    XORQ DX, DX
    MOVQ x+8(FP), DX
    MOVQ 80(DX), DX
    MOVQ x+8(FP), AX
    MOVQ 0(AX), AX
    MULXQ AX, AX, R13
    ADOXQ AX, R14
    MOVQ x+8(FP), AX
    MOVQ 8(AX), AX
    ADCXQ R13, R15
    MULXQ AX, AX, R13
    ADOXQ AX, R15
    MOVQ x+8(FP), AX
    MOVQ 16(AX), AX
    ADCXQ R13, CX
    MULXQ AX, AX, R13
    ADOXQ AX, CX
    MOVQ x+8(FP), AX
    MOVQ 24(AX), AX
    ADCXQ R13, BX
    MULXQ AX, AX, R13
    ADOXQ AX, BX
    MOVQ x+8(FP), AX
    MOVQ 32(AX), AX
    ADCXQ R13, BP
    MULXQ AX, AX, R13
    ADOXQ AX, BP
    MOVQ x+8(FP), AX
    MOVQ 40(AX), AX
    ADCXQ R13, SI
    MULXQ AX, AX, R13
    ADOXQ AX, SI
    MOVQ x+8(FP), AX
    MOVQ 48(AX), AX
    ADCXQ R13, DI
    MULXQ AX, AX, R13
    ADOXQ AX, DI
    MOVQ x+8(FP), AX
    MOVQ 56(AX), AX
    ADCXQ R13, R8
    MULXQ AX, AX, R13
    ADOXQ AX, R8
    MOVQ x+8(FP), AX
    MOVQ 64(AX), AX
    ADCXQ R13, R9
    MULXQ AX, AX, R13
    ADOXQ AX, R9
    MOVQ x+8(FP), AX
    MOVQ 72(AX), AX
    ADCXQ R13, R10
    MULXQ AX, AX, R13
    ADOXQ AX, R10
    MOVQ x+8(FP), AX
    MOVQ 80(AX), AX
    ADCXQ R13, R11
    MULXQ AX, AX, R13
    ADOXQ AX, R11
    MOVQ x+8(FP), AX
    MOVQ 88(AX), AX
    ADCXQ R13, R12
    MULXQ AX, AX, R13
    ADOXQ AX, R12
    // add the last carries to R13
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R13
    ADOXQ DX, R13
    PUSHQ R13
    MOVQ R14, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R13
    ADCXQ R14, AX
    MOVQ R13, R14
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R15, R14
    MULXQ ·qElement+8(SB), AX, R15
    ADOXQ AX, R14
    ADCXQ CX, R15
    MULXQ ·qElement+16(SB), AX, CX
    ADOXQ AX, R15
    ADCXQ BX, CX
    MULXQ ·qElement+24(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+32(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+40(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+48(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+56(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+64(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+72(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+80(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+88(SB), AX, R12
    ADOXQ AX, R11
    POPQ R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ R13, R12
    XORQ DX, DX
    MOVQ x+8(FP), DX
    MOVQ 88(DX), DX
    MOVQ x+8(FP), AX
    MOVQ 0(AX), AX
    MULXQ AX, AX, R13
    ADOXQ AX, R14
    MOVQ x+8(FP), AX
    MOVQ 8(AX), AX
    ADCXQ R13, R15
    MULXQ AX, AX, R13
    ADOXQ AX, R15
    MOVQ x+8(FP), AX
    MOVQ 16(AX), AX
    ADCXQ R13, CX
    MULXQ AX, AX, R13
    ADOXQ AX, CX
    MOVQ x+8(FP), AX
    MOVQ 24(AX), AX
    ADCXQ R13, BX
    MULXQ AX, AX, R13
    ADOXQ AX, BX
    MOVQ x+8(FP), AX
    MOVQ 32(AX), AX
    ADCXQ R13, BP
    MULXQ AX, AX, R13
    ADOXQ AX, BP
    MOVQ x+8(FP), AX
    MOVQ 40(AX), AX
    ADCXQ R13, SI
    MULXQ AX, AX, R13
    ADOXQ AX, SI
    MOVQ x+8(FP), AX
    MOVQ 48(AX), AX
    ADCXQ R13, DI
    MULXQ AX, AX, R13
    ADOXQ AX, DI
    MOVQ x+8(FP), AX
    MOVQ 56(AX), AX
    ADCXQ R13, R8
    MULXQ AX, AX, R13
    ADOXQ AX, R8
    MOVQ x+8(FP), AX
    MOVQ 64(AX), AX
    ADCXQ R13, R9
    MULXQ AX, AX, R13
    ADOXQ AX, R9
    MOVQ x+8(FP), AX
    MOVQ 72(AX), AX
    ADCXQ R13, R10
    MULXQ AX, AX, R13
    ADOXQ AX, R10
    MOVQ x+8(FP), AX
    MOVQ 80(AX), AX
    ADCXQ R13, R11
    MULXQ AX, AX, R13
    ADOXQ AX, R11
    MOVQ x+8(FP), AX
    MOVQ 88(AX), AX
    ADCXQ R13, R12
    MULXQ AX, AX, R13
    ADOXQ AX, R12
    // add the last carries to R13
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R13
    ADOXQ DX, R13
    PUSHQ R13
    MOVQ R14, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R13
    ADCXQ R14, AX
    MOVQ R13, R14
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R15, R14
    MULXQ ·qElement+8(SB), AX, R15
    ADOXQ AX, R14
    ADCXQ CX, R15
    MULXQ ·qElement+16(SB), AX, CX
    ADOXQ AX, R15
    ADCXQ BX, CX
    MULXQ ·qElement+24(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+32(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+40(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+48(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+56(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+64(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+72(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+80(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+88(SB), AX, R12
    ADOXQ AX, R11
    POPQ R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ R13, R12
    MOVQ res+0(FP), R13
    MOVQ R14, t0-8(SP)
    SUBQ ·qElement+0(SB), R14
    MOVQ R15, t1-16(SP)
    SBBQ ·qElement+8(SB), R15
    MOVQ CX, t2-24(SP)
    SBBQ ·qElement+16(SB), CX
    MOVQ BX, t3-32(SP)
    SBBQ ·qElement+24(SB), BX
    MOVQ BP, t4-40(SP)
    SBBQ ·qElement+32(SB), BP
    MOVQ SI, t5-48(SP)
    SBBQ ·qElement+40(SB), SI
    MOVQ DI, t6-56(SP)
    SBBQ ·qElement+48(SB), DI
    MOVQ R8, t7-64(SP)
    SBBQ ·qElement+56(SB), R8
    MOVQ R9, t8-72(SP)
    SBBQ ·qElement+64(SB), R9
    MOVQ R10, t9-80(SP)
    SBBQ ·qElement+72(SB), R10
    MOVQ R11, t10-88(SP)
    SBBQ ·qElement+80(SB), R11
    MOVQ R12, t11-96(SP)
    SBBQ ·qElement+88(SB), R12
    CMOVQCS t0-8(SP), R14
    CMOVQCS t1-16(SP), R15
    CMOVQCS t2-24(SP), CX
    CMOVQCS t3-32(SP), BX
    CMOVQCS t4-40(SP), BP
    CMOVQCS t5-48(SP), SI
    CMOVQCS t6-56(SP), DI
    CMOVQCS t7-64(SP), R8
    CMOVQCS t8-72(SP), R9
    CMOVQCS t9-80(SP), R10
    CMOVQCS t10-88(SP), R11
    CMOVQCS t11-96(SP), R12
    MOVQ R14, 0(R13)
    MOVQ R15, 8(R13)
    MOVQ CX, 16(R13)
    MOVQ BX, 24(R13)
    MOVQ BP, 32(R13)
    MOVQ SI, 40(R13)
    MOVQ DI, 48(R13)
    MOVQ R8, 56(R13)
    MOVQ R9, 64(R13)
    MOVQ R10, 72(R13)
    MOVQ R11, 80(R13)
    MOVQ R12, 88(R13)
    RET
l2:
    MOVQ res+0(FP), AX
    MOVQ AX, (SP)
    MOVQ x+8(FP), AX
    MOVQ AX, 8(SP)
CALL ·_squareGeneric(SB)
    RET

TEXT ·fromMont(SB), $96-8
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
    CMPB ·supportAdx(SB), $0x0000000000000001
    JNE l3
    MOVQ res+0(FP), R13
    MOVQ 0(R13), R14
    MOVQ 8(R13), R15
    MOVQ 16(R13), CX
    MOVQ 24(R13), BX
    MOVQ 32(R13), BP
    MOVQ 40(R13), SI
    MOVQ 48(R13), DI
    MOVQ 56(R13), R8
    MOVQ 64(R13), R9
    MOVQ 72(R13), R10
    MOVQ 80(R13), R11
    MOVQ 88(R13), R12
    XORQ DX, DX
    MOVQ R14, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R13
    ADCXQ R14, AX
    MOVQ R13, R14
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R15, R14
    MULXQ ·qElement+8(SB), AX, R15
    ADOXQ AX, R14
    ADCXQ CX, R15
    MULXQ ·qElement+16(SB), AX, CX
    ADOXQ AX, R15
    ADCXQ BX, CX
    MULXQ ·qElement+24(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+32(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+40(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+48(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+56(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+64(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+72(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+80(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+88(SB), AX, R12
    ADOXQ AX, R11
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ AX, R12
    XORQ DX, DX
    MOVQ R14, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R13
    ADCXQ R14, AX
    MOVQ R13, R14
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R15, R14
    MULXQ ·qElement+8(SB), AX, R15
    ADOXQ AX, R14
    ADCXQ CX, R15
    MULXQ ·qElement+16(SB), AX, CX
    ADOXQ AX, R15
    ADCXQ BX, CX
    MULXQ ·qElement+24(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+32(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+40(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+48(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+56(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+64(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+72(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+80(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+88(SB), AX, R12
    ADOXQ AX, R11
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ AX, R12
    XORQ DX, DX
    MOVQ R14, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R13
    ADCXQ R14, AX
    MOVQ R13, R14
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R15, R14
    MULXQ ·qElement+8(SB), AX, R15
    ADOXQ AX, R14
    ADCXQ CX, R15
    MULXQ ·qElement+16(SB), AX, CX
    ADOXQ AX, R15
    ADCXQ BX, CX
    MULXQ ·qElement+24(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+32(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+40(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+48(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+56(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+64(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+72(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+80(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+88(SB), AX, R12
    ADOXQ AX, R11
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ AX, R12
    XORQ DX, DX
    MOVQ R14, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R13
    ADCXQ R14, AX
    MOVQ R13, R14
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R15, R14
    MULXQ ·qElement+8(SB), AX, R15
    ADOXQ AX, R14
    ADCXQ CX, R15
    MULXQ ·qElement+16(SB), AX, CX
    ADOXQ AX, R15
    ADCXQ BX, CX
    MULXQ ·qElement+24(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+32(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+40(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+48(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+56(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+64(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+72(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+80(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+88(SB), AX, R12
    ADOXQ AX, R11
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ AX, R12
    XORQ DX, DX
    MOVQ R14, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R13
    ADCXQ R14, AX
    MOVQ R13, R14
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R15, R14
    MULXQ ·qElement+8(SB), AX, R15
    ADOXQ AX, R14
    ADCXQ CX, R15
    MULXQ ·qElement+16(SB), AX, CX
    ADOXQ AX, R15
    ADCXQ BX, CX
    MULXQ ·qElement+24(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+32(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+40(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+48(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+56(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+64(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+72(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+80(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+88(SB), AX, R12
    ADOXQ AX, R11
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ AX, R12
    XORQ DX, DX
    MOVQ R14, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R13
    ADCXQ R14, AX
    MOVQ R13, R14
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R15, R14
    MULXQ ·qElement+8(SB), AX, R15
    ADOXQ AX, R14
    ADCXQ CX, R15
    MULXQ ·qElement+16(SB), AX, CX
    ADOXQ AX, R15
    ADCXQ BX, CX
    MULXQ ·qElement+24(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+32(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+40(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+48(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+56(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+64(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+72(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+80(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+88(SB), AX, R12
    ADOXQ AX, R11
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ AX, R12
    XORQ DX, DX
    MOVQ R14, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R13
    ADCXQ R14, AX
    MOVQ R13, R14
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R15, R14
    MULXQ ·qElement+8(SB), AX, R15
    ADOXQ AX, R14
    ADCXQ CX, R15
    MULXQ ·qElement+16(SB), AX, CX
    ADOXQ AX, R15
    ADCXQ BX, CX
    MULXQ ·qElement+24(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+32(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+40(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+48(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+56(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+64(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+72(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+80(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+88(SB), AX, R12
    ADOXQ AX, R11
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ AX, R12
    XORQ DX, DX
    MOVQ R14, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R13
    ADCXQ R14, AX
    MOVQ R13, R14
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R15, R14
    MULXQ ·qElement+8(SB), AX, R15
    ADOXQ AX, R14
    ADCXQ CX, R15
    MULXQ ·qElement+16(SB), AX, CX
    ADOXQ AX, R15
    ADCXQ BX, CX
    MULXQ ·qElement+24(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+32(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+40(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+48(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+56(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+64(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+72(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+80(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+88(SB), AX, R12
    ADOXQ AX, R11
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ AX, R12
    XORQ DX, DX
    MOVQ R14, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R13
    ADCXQ R14, AX
    MOVQ R13, R14
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R15, R14
    MULXQ ·qElement+8(SB), AX, R15
    ADOXQ AX, R14
    ADCXQ CX, R15
    MULXQ ·qElement+16(SB), AX, CX
    ADOXQ AX, R15
    ADCXQ BX, CX
    MULXQ ·qElement+24(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+32(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+40(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+48(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+56(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+64(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+72(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+80(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+88(SB), AX, R12
    ADOXQ AX, R11
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ AX, R12
    XORQ DX, DX
    MOVQ R14, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R13
    ADCXQ R14, AX
    MOVQ R13, R14
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R15, R14
    MULXQ ·qElement+8(SB), AX, R15
    ADOXQ AX, R14
    ADCXQ CX, R15
    MULXQ ·qElement+16(SB), AX, CX
    ADOXQ AX, R15
    ADCXQ BX, CX
    MULXQ ·qElement+24(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+32(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+40(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+48(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+56(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+64(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+72(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+80(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+88(SB), AX, R12
    ADOXQ AX, R11
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ AX, R12
    XORQ DX, DX
    MOVQ R14, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R13
    ADCXQ R14, AX
    MOVQ R13, R14
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R15, R14
    MULXQ ·qElement+8(SB), AX, R15
    ADOXQ AX, R14
    ADCXQ CX, R15
    MULXQ ·qElement+16(SB), AX, CX
    ADOXQ AX, R15
    ADCXQ BX, CX
    MULXQ ·qElement+24(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+32(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+40(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+48(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+56(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+64(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+72(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+80(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+88(SB), AX, R12
    ADOXQ AX, R11
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ AX, R12
    XORQ DX, DX
    MOVQ R14, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R13
    ADCXQ R14, AX
    MOVQ R13, R14
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ R15, R14
    MULXQ ·qElement+8(SB), AX, R15
    ADOXQ AX, R14
    ADCXQ CX, R15
    MULXQ ·qElement+16(SB), AX, CX
    ADOXQ AX, R15
    ADCXQ BX, CX
    MULXQ ·qElement+24(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+32(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+40(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+48(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+56(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+64(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+72(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+80(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+88(SB), AX, R12
    ADOXQ AX, R11
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R12
    ADOXQ AX, R12
    MOVQ res+0(FP), R13
    MOVQ R14, t0-8(SP)
    SUBQ ·qElement+0(SB), R14
    MOVQ R15, t1-16(SP)
    SBBQ ·qElement+8(SB), R15
    MOVQ CX, t2-24(SP)
    SBBQ ·qElement+16(SB), CX
    MOVQ BX, t3-32(SP)
    SBBQ ·qElement+24(SB), BX
    MOVQ BP, t4-40(SP)
    SBBQ ·qElement+32(SB), BP
    MOVQ SI, t5-48(SP)
    SBBQ ·qElement+40(SB), SI
    MOVQ DI, t6-56(SP)
    SBBQ ·qElement+48(SB), DI
    MOVQ R8, t7-64(SP)
    SBBQ ·qElement+56(SB), R8
    MOVQ R9, t8-72(SP)
    SBBQ ·qElement+64(SB), R9
    MOVQ R10, t9-80(SP)
    SBBQ ·qElement+72(SB), R10
    MOVQ R11, t10-88(SP)
    SBBQ ·qElement+80(SB), R11
    MOVQ R12, t11-96(SP)
    SBBQ ·qElement+88(SB), R12
    CMOVQCS t0-8(SP), R14
    CMOVQCS t1-16(SP), R15
    CMOVQCS t2-24(SP), CX
    CMOVQCS t3-32(SP), BX
    CMOVQCS t4-40(SP), BP
    CMOVQCS t5-48(SP), SI
    CMOVQCS t6-56(SP), DI
    CMOVQCS t7-64(SP), R8
    CMOVQCS t8-72(SP), R9
    CMOVQCS t9-80(SP), R10
    CMOVQCS t10-88(SP), R11
    CMOVQCS t11-96(SP), R12
    MOVQ R14, 0(R13)
    MOVQ R15, 8(R13)
    MOVQ CX, 16(R13)
    MOVQ BX, 24(R13)
    MOVQ BP, 32(R13)
    MOVQ SI, 40(R13)
    MOVQ DI, 48(R13)
    MOVQ R8, 56(R13)
    MOVQ R9, 64(R13)
    MOVQ R10, 72(R13)
    MOVQ R11, 80(R13)
    MOVQ R12, 88(R13)
    RET
l3:
    MOVQ res+0(FP), AX
    MOVQ AX, (SP)
CALL ·_fromMontGeneric(SB)
    RET

TEXT ·reduce(SB), $96-8
    MOVQ res+0(FP), AX
    MOVQ 0(AX), DX
    MOVQ 8(AX), CX
    MOVQ 16(AX), BX
    MOVQ 24(AX), BP
    MOVQ 32(AX), SI
    MOVQ 40(AX), DI
    MOVQ 48(AX), R8
    MOVQ 56(AX), R9
    MOVQ 64(AX), R10
    MOVQ 72(AX), R11
    MOVQ 80(AX), R12
    MOVQ 88(AX), R13
    MOVQ DX, t0-8(SP)
    SUBQ ·qElement+0(SB), DX
    MOVQ CX, t1-16(SP)
    SBBQ ·qElement+8(SB), CX
    MOVQ BX, t2-24(SP)
    SBBQ ·qElement+16(SB), BX
    MOVQ BP, t3-32(SP)
    SBBQ ·qElement+24(SB), BP
    MOVQ SI, t4-40(SP)
    SBBQ ·qElement+32(SB), SI
    MOVQ DI, t5-48(SP)
    SBBQ ·qElement+40(SB), DI
    MOVQ R8, t6-56(SP)
    SBBQ ·qElement+48(SB), R8
    MOVQ R9, t7-64(SP)
    SBBQ ·qElement+56(SB), R9
    MOVQ R10, t8-72(SP)
    SBBQ ·qElement+64(SB), R10
    MOVQ R11, t9-80(SP)
    SBBQ ·qElement+72(SB), R11
    MOVQ R12, t10-88(SP)
    SBBQ ·qElement+80(SB), R12
    MOVQ R13, t11-96(SP)
    SBBQ ·qElement+88(SB), R13
    CMOVQCS t0-8(SP), DX
    CMOVQCS t1-16(SP), CX
    CMOVQCS t2-24(SP), BX
    CMOVQCS t3-32(SP), BP
    CMOVQCS t4-40(SP), SI
    CMOVQCS t5-48(SP), DI
    CMOVQCS t6-56(SP), R8
    CMOVQCS t7-64(SP), R9
    CMOVQCS t8-72(SP), R10
    CMOVQCS t9-80(SP), R11
    CMOVQCS t10-88(SP), R12
    CMOVQCS t11-96(SP), R13
    MOVQ DX, 0(AX)
    MOVQ CX, 8(AX)
    MOVQ BX, 16(AX)
    MOVQ BP, 24(AX)
    MOVQ SI, 32(AX)
    MOVQ DI, 40(AX)
    MOVQ R8, 48(AX)
    MOVQ R9, 56(AX)
    MOVQ R10, 64(AX)
    MOVQ R11, 72(AX)
    MOVQ R12, 80(AX)
    MOVQ R13, 88(AX)
    RET

TEXT ·add(SB), $96-24
    MOVQ x+8(FP), AX
    MOVQ 0(AX), BX
    MOVQ 8(AX), BP
    MOVQ 16(AX), SI
    MOVQ 24(AX), DI
    MOVQ 32(AX), R8
    MOVQ 40(AX), R9
    MOVQ 48(AX), R10
    MOVQ 56(AX), R11
    MOVQ 64(AX), R12
    MOVQ 72(AX), R13
    MOVQ 80(AX), R14
    MOVQ 88(AX), R15
    MOVQ y+16(FP), DX
    ADDQ 0(DX), BX
    ADCQ 8(DX), BP
    ADCQ 16(DX), SI
    ADCQ 24(DX), DI
    ADCQ 32(DX), R8
    ADCQ 40(DX), R9
    ADCQ 48(DX), R10
    ADCQ 56(DX), R11
    ADCQ 64(DX), R12
    ADCQ 72(DX), R13
    ADCQ 80(DX), R14
    ADCQ 88(DX), R15
    MOVQ res+0(FP), CX
    MOVQ BX, t0-8(SP)
    SUBQ ·qElement+0(SB), BX
    MOVQ BP, t1-16(SP)
    SBBQ ·qElement+8(SB), BP
    MOVQ SI, t2-24(SP)
    SBBQ ·qElement+16(SB), SI
    MOVQ DI, t3-32(SP)
    SBBQ ·qElement+24(SB), DI
    MOVQ R8, t4-40(SP)
    SBBQ ·qElement+32(SB), R8
    MOVQ R9, t5-48(SP)
    SBBQ ·qElement+40(SB), R9
    MOVQ R10, t6-56(SP)
    SBBQ ·qElement+48(SB), R10
    MOVQ R11, t7-64(SP)
    SBBQ ·qElement+56(SB), R11
    MOVQ R12, t8-72(SP)
    SBBQ ·qElement+64(SB), R12
    MOVQ R13, t9-80(SP)
    SBBQ ·qElement+72(SB), R13
    MOVQ R14, t10-88(SP)
    SBBQ ·qElement+80(SB), R14
    MOVQ R15, t11-96(SP)
    SBBQ ·qElement+88(SB), R15
    CMOVQCS t0-8(SP), BX
    CMOVQCS t1-16(SP), BP
    CMOVQCS t2-24(SP), SI
    CMOVQCS t3-32(SP), DI
    CMOVQCS t4-40(SP), R8
    CMOVQCS t5-48(SP), R9
    CMOVQCS t6-56(SP), R10
    CMOVQCS t7-64(SP), R11
    CMOVQCS t8-72(SP), R12
    CMOVQCS t9-80(SP), R13
    CMOVQCS t10-88(SP), R14
    CMOVQCS t11-96(SP), R15
    MOVQ BX, 0(CX)
    MOVQ BP, 8(CX)
    MOVQ SI, 16(CX)
    MOVQ DI, 24(CX)
    MOVQ R8, 32(CX)
    MOVQ R9, 40(CX)
    MOVQ R10, 48(CX)
    MOVQ R11, 56(CX)
    MOVQ R12, 64(CX)
    MOVQ R13, 72(CX)
    MOVQ R14, 80(CX)
    MOVQ R15, 88(CX)
    RET

TEXT ·sub(SB), NOSPLIT, $0-24
    MOVQ x+8(FP), R13
    MOVQ 0(R13), AX
    MOVQ 8(R13), DX
    MOVQ 16(R13), CX
    MOVQ 24(R13), BX
    MOVQ 32(R13), BP
    MOVQ 40(R13), SI
    MOVQ 48(R13), DI
    MOVQ 56(R13), R8
    MOVQ 64(R13), R9
    MOVQ 72(R13), R10
    MOVQ 80(R13), R11
    MOVQ 88(R13), R12
    MOVQ y+16(FP), R14
    SUBQ 0(R14), AX
    SBBQ 8(R14), DX
    SBBQ 16(R14), CX
    SBBQ 24(R14), BX
    SBBQ 32(R14), BP
    SBBQ 40(R14), SI
    SBBQ 48(R14), DI
    SBBQ 56(R14), R8
    SBBQ 64(R14), R9
    SBBQ 72(R14), R10
    SBBQ 80(R14), R11
    SBBQ 88(R14), R12
    JCC l4
    ADDQ ·qElement+0(SB), AX
    ADCQ ·qElement+8(SB), DX
    ADCQ ·qElement+16(SB), CX
    ADCQ ·qElement+24(SB), BX
    ADCQ ·qElement+32(SB), BP
    ADCQ ·qElement+40(SB), SI
    ADCQ ·qElement+48(SB), DI
    ADCQ ·qElement+56(SB), R8
    ADCQ ·qElement+64(SB), R9
    ADCQ ·qElement+72(SB), R10
    ADCQ ·qElement+80(SB), R11
    ADCQ ·qElement+88(SB), R12
l4:
    MOVQ res+0(FP), R15
    MOVQ AX, 0(R15)
    MOVQ DX, 8(R15)
    MOVQ CX, 16(R15)
    MOVQ BX, 24(R15)
    MOVQ BP, 32(R15)
    MOVQ SI, 40(R15)
    MOVQ DI, 48(R15)
    MOVQ R8, 56(R15)
    MOVQ R9, 64(R15)
    MOVQ R10, 72(R15)
    MOVQ R11, 80(R15)
    MOVQ R12, 88(R15)
    RET

TEXT ·double(SB), $96-16
    MOVQ res+0(FP), DX
    MOVQ x+8(FP), AX
    MOVQ 0(AX), CX
    MOVQ 8(AX), BX
    MOVQ 16(AX), BP
    MOVQ 24(AX), SI
    MOVQ 32(AX), DI
    MOVQ 40(AX), R8
    MOVQ 48(AX), R9
    MOVQ 56(AX), R10
    MOVQ 64(AX), R11
    MOVQ 72(AX), R12
    MOVQ 80(AX), R13
    MOVQ 88(AX), R14
    ADDQ CX, CX
    ADCQ BX, BX
    ADCQ BP, BP
    ADCQ SI, SI
    ADCQ DI, DI
    ADCQ R8, R8
    ADCQ R9, R9
    ADCQ R10, R10
    ADCQ R11, R11
    ADCQ R12, R12
    ADCQ R13, R13
    ADCQ R14, R14
    MOVQ CX, t0-8(SP)
    SUBQ ·qElement+0(SB), CX
    MOVQ BX, t1-16(SP)
    SBBQ ·qElement+8(SB), BX
    MOVQ BP, t2-24(SP)
    SBBQ ·qElement+16(SB), BP
    MOVQ SI, t3-32(SP)
    SBBQ ·qElement+24(SB), SI
    MOVQ DI, t4-40(SP)
    SBBQ ·qElement+32(SB), DI
    MOVQ R8, t5-48(SP)
    SBBQ ·qElement+40(SB), R8
    MOVQ R9, t6-56(SP)
    SBBQ ·qElement+48(SB), R9
    MOVQ R10, t7-64(SP)
    SBBQ ·qElement+56(SB), R10
    MOVQ R11, t8-72(SP)
    SBBQ ·qElement+64(SB), R11
    MOVQ R12, t9-80(SP)
    SBBQ ·qElement+72(SB), R12
    MOVQ R13, t10-88(SP)
    SBBQ ·qElement+80(SB), R13
    MOVQ R14, t11-96(SP)
    SBBQ ·qElement+88(SB), R14
    CMOVQCS t0-8(SP), CX
    CMOVQCS t1-16(SP), BX
    CMOVQCS t2-24(SP), BP
    CMOVQCS t3-32(SP), SI
    CMOVQCS t4-40(SP), DI
    CMOVQCS t5-48(SP), R8
    CMOVQCS t6-56(SP), R9
    CMOVQCS t7-64(SP), R10
    CMOVQCS t8-72(SP), R11
    CMOVQCS t9-80(SP), R12
    CMOVQCS t10-88(SP), R13
    CMOVQCS t11-96(SP), R14
    MOVQ CX, 0(DX)
    MOVQ BX, 8(DX)
    MOVQ BP, 16(DX)
    MOVQ SI, 24(DX)
    MOVQ DI, 32(DX)
    MOVQ R8, 40(DX)
    MOVQ R9, 48(DX)
    MOVQ R10, 56(DX)
    MOVQ R11, 64(DX)
    MOVQ R12, 72(DX)
    MOVQ R13, 80(DX)
    MOVQ R14, 88(DX)
    RET

TEXT ·neg(SB), NOSPLIT, $0-16
    MOVQ res+0(FP), DX
    MOVQ x+8(FP), AX
    MOVQ 0(AX), BX
    MOVQ 8(AX), BP
    MOVQ 16(AX), SI
    MOVQ 24(AX), DI
    MOVQ 32(AX), R8
    MOVQ 40(AX), R9
    MOVQ 48(AX), R10
    MOVQ 56(AX), R11
    MOVQ 64(AX), R12
    MOVQ 72(AX), R13
    MOVQ 80(AX), R14
    MOVQ 88(AX), R15
    MOVQ BX, AX
    ORQ BP, AX
    ORQ SI, AX
    ORQ DI, AX
    ORQ R8, AX
    ORQ R9, AX
    ORQ R10, AX
    ORQ R11, AX
    ORQ R12, AX
    ORQ R13, AX
    ORQ R14, AX
    ORQ R15, AX
    TESTQ AX, AX
    JNE l5
    MOVQ AX, 0(DX)
    MOVQ AX, 8(DX)
    MOVQ AX, 16(DX)
    MOVQ AX, 24(DX)
    MOVQ AX, 32(DX)
    MOVQ AX, 40(DX)
    RET
l5:
    MOVQ $0xf49d00000000008b, CX
    SUBQ BX, CX
    MOVQ CX, 0(DX)
    MOVQ $0xe6913e6870000082, CX
    SBBQ BP, CX
    MOVQ CX, 8(DX)
    MOVQ $0x160cf8aeeaf0a437, CX
    SBBQ SI, CX
    MOVQ CX, 16(DX)
    MOVQ $0x98a116c25667a8f8, CX
    SBBQ DI, CX
    MOVQ CX, 24(DX)
    MOVQ $0x71dcd3dc73ebff2e, CX
    SBBQ R8, CX
    MOVQ CX, 32(DX)
    MOVQ $0x8689c8ed12f9fd90, CX
    SBBQ R9, CX
    MOVQ CX, 40(DX)
    MOVQ $0x03cebaff25b42304, CX
    SBBQ R10, CX
    MOVQ CX, 48(DX)
    MOVQ $0x707ba638e584e919, CX
    SBBQ R11, CX
    MOVQ CX, 56(DX)
    MOVQ $0x528275ef8087be41, CX
    SBBQ R12, CX
    MOVQ CX, 64(DX)
    MOVQ $0xb926186a81d14688, CX
    SBBQ R13, CX
    MOVQ CX, 72(DX)
    MOVQ $0xd187c94004faff3e, CX
    SBBQ R14, CX
    MOVQ CX, 80(DX)
    MOVQ $0x0122e824fb83ce0a, CX
    SBBQ R15, CX
    MOVQ CX, 88(DX)
    RET
