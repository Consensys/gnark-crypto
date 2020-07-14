#include "textflag.h"
#include "funcdata.h"

TEXT ·_mulLargeADXElement(SB), $96-200

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
	
    XORQ DX, DX
    MOVQ y0+104(FP), DX
    MULXQ x0+8(FP), CX, BX
    MULXQ x1+16(FP), AX, BP
    ADOXQ AX, BX
    MULXQ x2+24(FP), AX, SI
    ADOXQ AX, BP
    MULXQ x3+32(FP), AX, DI
    ADOXQ AX, SI
    MULXQ x4+40(FP), AX, R8
    ADOXQ AX, DI
    MULXQ x5+48(FP), AX, R9
    ADOXQ AX, R8
    MULXQ x6+56(FP), AX, R10
    ADOXQ AX, R9
    MULXQ x7+64(FP), AX, R11
    ADOXQ AX, R10
    MULXQ x8+72(FP), AX, R12
    ADOXQ AX, R11
    MULXQ x9+80(FP), AX, R13
    ADOXQ AX, R12
    MULXQ x10+88(FP), AX, R14
    ADOXQ AX, R13
    MULXQ x11+96(FP), AX, R15
    ADOXQ AX, R14
    // add the last carries to R15
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R15
    ADOXQ DX, R15
    PUSHQ R15
    MOVQ CX, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R15
    ADCXQ CX, AX
    MOVQ R15, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ BX, CX
    MULXQ ·qElement+8(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+16(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+24(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+32(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+40(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+48(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+56(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+64(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+72(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qElement+80(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ R14, R13
    MULXQ ·qElement+88(SB), AX, R14
    ADOXQ AX, R13
    POPQ R15
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R14
    ADOXQ R15, R14
    XORQ DX, DX
    MOVQ y1+112(FP), DX
    MULXQ x0+8(FP), AX, R15
    ADOXQ AX, CX
    ADCXQ R15, BX
    MULXQ x1+16(FP), AX, R15
    ADOXQ AX, BX
    ADCXQ R15, BP
    MULXQ x2+24(FP), AX, R15
    ADOXQ AX, BP
    ADCXQ R15, SI
    MULXQ x3+32(FP), AX, R15
    ADOXQ AX, SI
    ADCXQ R15, DI
    MULXQ x4+40(FP), AX, R15
    ADOXQ AX, DI
    ADCXQ R15, R8
    MULXQ x5+48(FP), AX, R15
    ADOXQ AX, R8
    ADCXQ R15, R9
    MULXQ x6+56(FP), AX, R15
    ADOXQ AX, R9
    ADCXQ R15, R10
    MULXQ x7+64(FP), AX, R15
    ADOXQ AX, R10
    ADCXQ R15, R11
    MULXQ x8+72(FP), AX, R15
    ADOXQ AX, R11
    ADCXQ R15, R12
    MULXQ x9+80(FP), AX, R15
    ADOXQ AX, R12
    ADCXQ R15, R13
    MULXQ x10+88(FP), AX, R15
    ADOXQ AX, R13
    ADCXQ R15, R14
    MULXQ x11+96(FP), AX, R15
    ADOXQ AX, R14
    // add the last carries to R15
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R15
    ADOXQ DX, R15
    PUSHQ R15
    MOVQ CX, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R15
    ADCXQ CX, AX
    MOVQ R15, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ BX, CX
    MULXQ ·qElement+8(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+16(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+24(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+32(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+40(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+48(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+56(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+64(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+72(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qElement+80(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ R14, R13
    MULXQ ·qElement+88(SB), AX, R14
    ADOXQ AX, R13
    POPQ R15
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R14
    ADOXQ R15, R14
    XORQ DX, DX
    MOVQ y2+120(FP), DX
    MULXQ x0+8(FP), AX, R15
    ADOXQ AX, CX
    ADCXQ R15, BX
    MULXQ x1+16(FP), AX, R15
    ADOXQ AX, BX
    ADCXQ R15, BP
    MULXQ x2+24(FP), AX, R15
    ADOXQ AX, BP
    ADCXQ R15, SI
    MULXQ x3+32(FP), AX, R15
    ADOXQ AX, SI
    ADCXQ R15, DI
    MULXQ x4+40(FP), AX, R15
    ADOXQ AX, DI
    ADCXQ R15, R8
    MULXQ x5+48(FP), AX, R15
    ADOXQ AX, R8
    ADCXQ R15, R9
    MULXQ x6+56(FP), AX, R15
    ADOXQ AX, R9
    ADCXQ R15, R10
    MULXQ x7+64(FP), AX, R15
    ADOXQ AX, R10
    ADCXQ R15, R11
    MULXQ x8+72(FP), AX, R15
    ADOXQ AX, R11
    ADCXQ R15, R12
    MULXQ x9+80(FP), AX, R15
    ADOXQ AX, R12
    ADCXQ R15, R13
    MULXQ x10+88(FP), AX, R15
    ADOXQ AX, R13
    ADCXQ R15, R14
    MULXQ x11+96(FP), AX, R15
    ADOXQ AX, R14
    // add the last carries to R15
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R15
    ADOXQ DX, R15
    PUSHQ R15
    MOVQ CX, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R15
    ADCXQ CX, AX
    MOVQ R15, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ BX, CX
    MULXQ ·qElement+8(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+16(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+24(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+32(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+40(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+48(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+56(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+64(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+72(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qElement+80(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ R14, R13
    MULXQ ·qElement+88(SB), AX, R14
    ADOXQ AX, R13
    POPQ R15
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R14
    ADOXQ R15, R14
    XORQ DX, DX
    MOVQ y3+128(FP), DX
    MULXQ x0+8(FP), AX, R15
    ADOXQ AX, CX
    ADCXQ R15, BX
    MULXQ x1+16(FP), AX, R15
    ADOXQ AX, BX
    ADCXQ R15, BP
    MULXQ x2+24(FP), AX, R15
    ADOXQ AX, BP
    ADCXQ R15, SI
    MULXQ x3+32(FP), AX, R15
    ADOXQ AX, SI
    ADCXQ R15, DI
    MULXQ x4+40(FP), AX, R15
    ADOXQ AX, DI
    ADCXQ R15, R8
    MULXQ x5+48(FP), AX, R15
    ADOXQ AX, R8
    ADCXQ R15, R9
    MULXQ x6+56(FP), AX, R15
    ADOXQ AX, R9
    ADCXQ R15, R10
    MULXQ x7+64(FP), AX, R15
    ADOXQ AX, R10
    ADCXQ R15, R11
    MULXQ x8+72(FP), AX, R15
    ADOXQ AX, R11
    ADCXQ R15, R12
    MULXQ x9+80(FP), AX, R15
    ADOXQ AX, R12
    ADCXQ R15, R13
    MULXQ x10+88(FP), AX, R15
    ADOXQ AX, R13
    ADCXQ R15, R14
    MULXQ x11+96(FP), AX, R15
    ADOXQ AX, R14
    // add the last carries to R15
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R15
    ADOXQ DX, R15
    PUSHQ R15
    MOVQ CX, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R15
    ADCXQ CX, AX
    MOVQ R15, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ BX, CX
    MULXQ ·qElement+8(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+16(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+24(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+32(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+40(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+48(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+56(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+64(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+72(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qElement+80(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ R14, R13
    MULXQ ·qElement+88(SB), AX, R14
    ADOXQ AX, R13
    POPQ R15
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R14
    ADOXQ R15, R14
    XORQ DX, DX
    MOVQ y4+136(FP), DX
    MULXQ x0+8(FP), AX, R15
    ADOXQ AX, CX
    ADCXQ R15, BX
    MULXQ x1+16(FP), AX, R15
    ADOXQ AX, BX
    ADCXQ R15, BP
    MULXQ x2+24(FP), AX, R15
    ADOXQ AX, BP
    ADCXQ R15, SI
    MULXQ x3+32(FP), AX, R15
    ADOXQ AX, SI
    ADCXQ R15, DI
    MULXQ x4+40(FP), AX, R15
    ADOXQ AX, DI
    ADCXQ R15, R8
    MULXQ x5+48(FP), AX, R15
    ADOXQ AX, R8
    ADCXQ R15, R9
    MULXQ x6+56(FP), AX, R15
    ADOXQ AX, R9
    ADCXQ R15, R10
    MULXQ x7+64(FP), AX, R15
    ADOXQ AX, R10
    ADCXQ R15, R11
    MULXQ x8+72(FP), AX, R15
    ADOXQ AX, R11
    ADCXQ R15, R12
    MULXQ x9+80(FP), AX, R15
    ADOXQ AX, R12
    ADCXQ R15, R13
    MULXQ x10+88(FP), AX, R15
    ADOXQ AX, R13
    ADCXQ R15, R14
    MULXQ x11+96(FP), AX, R15
    ADOXQ AX, R14
    // add the last carries to R15
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R15
    ADOXQ DX, R15
    PUSHQ R15
    MOVQ CX, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R15
    ADCXQ CX, AX
    MOVQ R15, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ BX, CX
    MULXQ ·qElement+8(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+16(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+24(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+32(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+40(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+48(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+56(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+64(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+72(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qElement+80(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ R14, R13
    MULXQ ·qElement+88(SB), AX, R14
    ADOXQ AX, R13
    POPQ R15
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R14
    ADOXQ R15, R14
    XORQ DX, DX
    MOVQ y5+144(FP), DX
    MULXQ x0+8(FP), AX, R15
    ADOXQ AX, CX
    ADCXQ R15, BX
    MULXQ x1+16(FP), AX, R15
    ADOXQ AX, BX
    ADCXQ R15, BP
    MULXQ x2+24(FP), AX, R15
    ADOXQ AX, BP
    ADCXQ R15, SI
    MULXQ x3+32(FP), AX, R15
    ADOXQ AX, SI
    ADCXQ R15, DI
    MULXQ x4+40(FP), AX, R15
    ADOXQ AX, DI
    ADCXQ R15, R8
    MULXQ x5+48(FP), AX, R15
    ADOXQ AX, R8
    ADCXQ R15, R9
    MULXQ x6+56(FP), AX, R15
    ADOXQ AX, R9
    ADCXQ R15, R10
    MULXQ x7+64(FP), AX, R15
    ADOXQ AX, R10
    ADCXQ R15, R11
    MULXQ x8+72(FP), AX, R15
    ADOXQ AX, R11
    ADCXQ R15, R12
    MULXQ x9+80(FP), AX, R15
    ADOXQ AX, R12
    ADCXQ R15, R13
    MULXQ x10+88(FP), AX, R15
    ADOXQ AX, R13
    ADCXQ R15, R14
    MULXQ x11+96(FP), AX, R15
    ADOXQ AX, R14
    // add the last carries to R15
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R15
    ADOXQ DX, R15
    PUSHQ R15
    MOVQ CX, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R15
    ADCXQ CX, AX
    MOVQ R15, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ BX, CX
    MULXQ ·qElement+8(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+16(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+24(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+32(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+40(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+48(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+56(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+64(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+72(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qElement+80(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ R14, R13
    MULXQ ·qElement+88(SB), AX, R14
    ADOXQ AX, R13
    POPQ R15
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R14
    ADOXQ R15, R14
    XORQ DX, DX
    MOVQ y6+152(FP), DX
    MULXQ x0+8(FP), AX, R15
    ADOXQ AX, CX
    ADCXQ R15, BX
    MULXQ x1+16(FP), AX, R15
    ADOXQ AX, BX
    ADCXQ R15, BP
    MULXQ x2+24(FP), AX, R15
    ADOXQ AX, BP
    ADCXQ R15, SI
    MULXQ x3+32(FP), AX, R15
    ADOXQ AX, SI
    ADCXQ R15, DI
    MULXQ x4+40(FP), AX, R15
    ADOXQ AX, DI
    ADCXQ R15, R8
    MULXQ x5+48(FP), AX, R15
    ADOXQ AX, R8
    ADCXQ R15, R9
    MULXQ x6+56(FP), AX, R15
    ADOXQ AX, R9
    ADCXQ R15, R10
    MULXQ x7+64(FP), AX, R15
    ADOXQ AX, R10
    ADCXQ R15, R11
    MULXQ x8+72(FP), AX, R15
    ADOXQ AX, R11
    ADCXQ R15, R12
    MULXQ x9+80(FP), AX, R15
    ADOXQ AX, R12
    ADCXQ R15, R13
    MULXQ x10+88(FP), AX, R15
    ADOXQ AX, R13
    ADCXQ R15, R14
    MULXQ x11+96(FP), AX, R15
    ADOXQ AX, R14
    // add the last carries to R15
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R15
    ADOXQ DX, R15
    PUSHQ R15
    MOVQ CX, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R15
    ADCXQ CX, AX
    MOVQ R15, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ BX, CX
    MULXQ ·qElement+8(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+16(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+24(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+32(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+40(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+48(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+56(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+64(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+72(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qElement+80(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ R14, R13
    MULXQ ·qElement+88(SB), AX, R14
    ADOXQ AX, R13
    POPQ R15
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R14
    ADOXQ R15, R14
    XORQ DX, DX
    MOVQ y7+160(FP), DX
    MULXQ x0+8(FP), AX, R15
    ADOXQ AX, CX
    ADCXQ R15, BX
    MULXQ x1+16(FP), AX, R15
    ADOXQ AX, BX
    ADCXQ R15, BP
    MULXQ x2+24(FP), AX, R15
    ADOXQ AX, BP
    ADCXQ R15, SI
    MULXQ x3+32(FP), AX, R15
    ADOXQ AX, SI
    ADCXQ R15, DI
    MULXQ x4+40(FP), AX, R15
    ADOXQ AX, DI
    ADCXQ R15, R8
    MULXQ x5+48(FP), AX, R15
    ADOXQ AX, R8
    ADCXQ R15, R9
    MULXQ x6+56(FP), AX, R15
    ADOXQ AX, R9
    ADCXQ R15, R10
    MULXQ x7+64(FP), AX, R15
    ADOXQ AX, R10
    ADCXQ R15, R11
    MULXQ x8+72(FP), AX, R15
    ADOXQ AX, R11
    ADCXQ R15, R12
    MULXQ x9+80(FP), AX, R15
    ADOXQ AX, R12
    ADCXQ R15, R13
    MULXQ x10+88(FP), AX, R15
    ADOXQ AX, R13
    ADCXQ R15, R14
    MULXQ x11+96(FP), AX, R15
    ADOXQ AX, R14
    // add the last carries to R15
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R15
    ADOXQ DX, R15
    PUSHQ R15
    MOVQ CX, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R15
    ADCXQ CX, AX
    MOVQ R15, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ BX, CX
    MULXQ ·qElement+8(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+16(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+24(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+32(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+40(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+48(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+56(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+64(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+72(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qElement+80(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ R14, R13
    MULXQ ·qElement+88(SB), AX, R14
    ADOXQ AX, R13
    POPQ R15
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R14
    ADOXQ R15, R14
    XORQ DX, DX
    MOVQ y8+168(FP), DX
    MULXQ x0+8(FP), AX, R15
    ADOXQ AX, CX
    ADCXQ R15, BX
    MULXQ x1+16(FP), AX, R15
    ADOXQ AX, BX
    ADCXQ R15, BP
    MULXQ x2+24(FP), AX, R15
    ADOXQ AX, BP
    ADCXQ R15, SI
    MULXQ x3+32(FP), AX, R15
    ADOXQ AX, SI
    ADCXQ R15, DI
    MULXQ x4+40(FP), AX, R15
    ADOXQ AX, DI
    ADCXQ R15, R8
    MULXQ x5+48(FP), AX, R15
    ADOXQ AX, R8
    ADCXQ R15, R9
    MULXQ x6+56(FP), AX, R15
    ADOXQ AX, R9
    ADCXQ R15, R10
    MULXQ x7+64(FP), AX, R15
    ADOXQ AX, R10
    ADCXQ R15, R11
    MULXQ x8+72(FP), AX, R15
    ADOXQ AX, R11
    ADCXQ R15, R12
    MULXQ x9+80(FP), AX, R15
    ADOXQ AX, R12
    ADCXQ R15, R13
    MULXQ x10+88(FP), AX, R15
    ADOXQ AX, R13
    ADCXQ R15, R14
    MULXQ x11+96(FP), AX, R15
    ADOXQ AX, R14
    // add the last carries to R15
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R15
    ADOXQ DX, R15
    PUSHQ R15
    MOVQ CX, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R15
    ADCXQ CX, AX
    MOVQ R15, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ BX, CX
    MULXQ ·qElement+8(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+16(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+24(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+32(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+40(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+48(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+56(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+64(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+72(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qElement+80(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ R14, R13
    MULXQ ·qElement+88(SB), AX, R14
    ADOXQ AX, R13
    POPQ R15
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R14
    ADOXQ R15, R14
    XORQ DX, DX
    MOVQ y9+176(FP), DX
    MULXQ x0+8(FP), AX, R15
    ADOXQ AX, CX
    ADCXQ R15, BX
    MULXQ x1+16(FP), AX, R15
    ADOXQ AX, BX
    ADCXQ R15, BP
    MULXQ x2+24(FP), AX, R15
    ADOXQ AX, BP
    ADCXQ R15, SI
    MULXQ x3+32(FP), AX, R15
    ADOXQ AX, SI
    ADCXQ R15, DI
    MULXQ x4+40(FP), AX, R15
    ADOXQ AX, DI
    ADCXQ R15, R8
    MULXQ x5+48(FP), AX, R15
    ADOXQ AX, R8
    ADCXQ R15, R9
    MULXQ x6+56(FP), AX, R15
    ADOXQ AX, R9
    ADCXQ R15, R10
    MULXQ x7+64(FP), AX, R15
    ADOXQ AX, R10
    ADCXQ R15, R11
    MULXQ x8+72(FP), AX, R15
    ADOXQ AX, R11
    ADCXQ R15, R12
    MULXQ x9+80(FP), AX, R15
    ADOXQ AX, R12
    ADCXQ R15, R13
    MULXQ x10+88(FP), AX, R15
    ADOXQ AX, R13
    ADCXQ R15, R14
    MULXQ x11+96(FP), AX, R15
    ADOXQ AX, R14
    // add the last carries to R15
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R15
    ADOXQ DX, R15
    PUSHQ R15
    MOVQ CX, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R15
    ADCXQ CX, AX
    MOVQ R15, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ BX, CX
    MULXQ ·qElement+8(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+16(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+24(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+32(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+40(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+48(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+56(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+64(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+72(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qElement+80(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ R14, R13
    MULXQ ·qElement+88(SB), AX, R14
    ADOXQ AX, R13
    POPQ R15
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R14
    ADOXQ R15, R14
    XORQ DX, DX
    MOVQ y10+184(FP), DX
    MULXQ x0+8(FP), AX, R15
    ADOXQ AX, CX
    ADCXQ R15, BX
    MULXQ x1+16(FP), AX, R15
    ADOXQ AX, BX
    ADCXQ R15, BP
    MULXQ x2+24(FP), AX, R15
    ADOXQ AX, BP
    ADCXQ R15, SI
    MULXQ x3+32(FP), AX, R15
    ADOXQ AX, SI
    ADCXQ R15, DI
    MULXQ x4+40(FP), AX, R15
    ADOXQ AX, DI
    ADCXQ R15, R8
    MULXQ x5+48(FP), AX, R15
    ADOXQ AX, R8
    ADCXQ R15, R9
    MULXQ x6+56(FP), AX, R15
    ADOXQ AX, R9
    ADCXQ R15, R10
    MULXQ x7+64(FP), AX, R15
    ADOXQ AX, R10
    ADCXQ R15, R11
    MULXQ x8+72(FP), AX, R15
    ADOXQ AX, R11
    ADCXQ R15, R12
    MULXQ x9+80(FP), AX, R15
    ADOXQ AX, R12
    ADCXQ R15, R13
    MULXQ x10+88(FP), AX, R15
    ADOXQ AX, R13
    ADCXQ R15, R14
    MULXQ x11+96(FP), AX, R15
    ADOXQ AX, R14
    // add the last carries to R15
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R15
    ADOXQ DX, R15
    PUSHQ R15
    MOVQ CX, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R15
    ADCXQ CX, AX
    MOVQ R15, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ BX, CX
    MULXQ ·qElement+8(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+16(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+24(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+32(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+40(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+48(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+56(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+64(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+72(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qElement+80(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ R14, R13
    MULXQ ·qElement+88(SB), AX, R14
    ADOXQ AX, R13
    POPQ R15
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R14
    ADOXQ R15, R14
    XORQ DX, DX
    MOVQ y11+192(FP), DX
    MULXQ x0+8(FP), AX, R15
    ADOXQ AX, CX
    ADCXQ R15, BX
    MULXQ x1+16(FP), AX, R15
    ADOXQ AX, BX
    ADCXQ R15, BP
    MULXQ x2+24(FP), AX, R15
    ADOXQ AX, BP
    ADCXQ R15, SI
    MULXQ x3+32(FP), AX, R15
    ADOXQ AX, SI
    ADCXQ R15, DI
    MULXQ x4+40(FP), AX, R15
    ADOXQ AX, DI
    ADCXQ R15, R8
    MULXQ x5+48(FP), AX, R15
    ADOXQ AX, R8
    ADCXQ R15, R9
    MULXQ x6+56(FP), AX, R15
    ADOXQ AX, R9
    ADCXQ R15, R10
    MULXQ x7+64(FP), AX, R15
    ADOXQ AX, R10
    ADCXQ R15, R11
    MULXQ x8+72(FP), AX, R15
    ADOXQ AX, R11
    ADCXQ R15, R12
    MULXQ x9+80(FP), AX, R15
    ADOXQ AX, R12
    ADCXQ R15, R13
    MULXQ x10+88(FP), AX, R15
    ADOXQ AX, R13
    ADCXQ R15, R14
    MULXQ x11+96(FP), AX, R15
    ADOXQ AX, R14
    // add the last carries to R15
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R15
    ADOXQ DX, R15
    PUSHQ R15
    MOVQ CX, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R15
    ADCXQ CX, AX
    MOVQ R15, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ BX, CX
    MULXQ ·qElement+8(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+16(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+24(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+32(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+40(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+48(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+56(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+64(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+72(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qElement+80(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ R14, R13
    MULXQ ·qElement+88(SB), AX, R14
    ADOXQ AX, R13
    POPQ R15
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R14
    ADOXQ R15, R14
    MOVQ res+0(FP), R15
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
    MOVQ CX, 0(R15)
    MOVQ BX, 8(R15)
    MOVQ BP, 16(R15)
    MOVQ SI, 24(R15)
    MOVQ DI, 32(R15)
    MOVQ R8, 40(R15)
    MOVQ R9, 48(R15)
    MOVQ R10, 56(R15)
    MOVQ R11, 64(R15)
    MOVQ R12, 72(R15)
    MOVQ R13, 80(R15)
    MOVQ R14, 88(R15)
    RET

TEXT ·_fromMontADXElement(SB), $96-8
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
    JNE no_adx
    MOVQ res+0(FP), R15
    MOVQ 0(R15), CX
    MOVQ 8(R15), BX
    MOVQ 16(R15), BP
    MOVQ 24(R15), SI
    MOVQ 32(R15), DI
    MOVQ 40(R15), R8
    MOVQ 48(R15), R9
    MOVQ 56(R15), R10
    MOVQ 64(R15), R11
    MOVQ 72(R15), R12
    MOVQ 80(R15), R13
    MOVQ 88(R15), R14
    XORQ DX, DX
    MOVQ CX, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R15
    ADCXQ CX, AX
    MOVQ R15, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ BX, CX
    MULXQ ·qElement+8(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+16(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+24(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+32(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+40(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+48(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+56(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+64(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+72(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qElement+80(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ R14, R13
    MULXQ ·qElement+88(SB), AX, R14
    ADOXQ AX, R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R14
    ADOXQ AX, R14
    XORQ DX, DX
    MOVQ CX, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R15
    ADCXQ CX, AX
    MOVQ R15, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ BX, CX
    MULXQ ·qElement+8(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+16(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+24(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+32(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+40(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+48(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+56(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+64(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+72(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qElement+80(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ R14, R13
    MULXQ ·qElement+88(SB), AX, R14
    ADOXQ AX, R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R14
    ADOXQ AX, R14
    XORQ DX, DX
    MOVQ CX, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R15
    ADCXQ CX, AX
    MOVQ R15, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ BX, CX
    MULXQ ·qElement+8(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+16(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+24(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+32(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+40(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+48(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+56(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+64(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+72(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qElement+80(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ R14, R13
    MULXQ ·qElement+88(SB), AX, R14
    ADOXQ AX, R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R14
    ADOXQ AX, R14
    XORQ DX, DX
    MOVQ CX, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R15
    ADCXQ CX, AX
    MOVQ R15, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ BX, CX
    MULXQ ·qElement+8(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+16(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+24(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+32(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+40(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+48(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+56(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+64(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+72(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qElement+80(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ R14, R13
    MULXQ ·qElement+88(SB), AX, R14
    ADOXQ AX, R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R14
    ADOXQ AX, R14
    XORQ DX, DX
    MOVQ CX, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R15
    ADCXQ CX, AX
    MOVQ R15, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ BX, CX
    MULXQ ·qElement+8(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+16(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+24(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+32(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+40(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+48(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+56(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+64(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+72(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qElement+80(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ R14, R13
    MULXQ ·qElement+88(SB), AX, R14
    ADOXQ AX, R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R14
    ADOXQ AX, R14
    XORQ DX, DX
    MOVQ CX, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R15
    ADCXQ CX, AX
    MOVQ R15, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ BX, CX
    MULXQ ·qElement+8(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+16(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+24(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+32(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+40(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+48(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+56(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+64(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+72(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qElement+80(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ R14, R13
    MULXQ ·qElement+88(SB), AX, R14
    ADOXQ AX, R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R14
    ADOXQ AX, R14
    XORQ DX, DX
    MOVQ CX, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R15
    ADCXQ CX, AX
    MOVQ R15, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ BX, CX
    MULXQ ·qElement+8(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+16(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+24(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+32(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+40(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+48(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+56(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+64(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+72(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qElement+80(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ R14, R13
    MULXQ ·qElement+88(SB), AX, R14
    ADOXQ AX, R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R14
    ADOXQ AX, R14
    XORQ DX, DX
    MOVQ CX, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R15
    ADCXQ CX, AX
    MOVQ R15, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ BX, CX
    MULXQ ·qElement+8(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+16(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+24(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+32(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+40(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+48(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+56(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+64(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+72(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qElement+80(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ R14, R13
    MULXQ ·qElement+88(SB), AX, R14
    ADOXQ AX, R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R14
    ADOXQ AX, R14
    XORQ DX, DX
    MOVQ CX, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R15
    ADCXQ CX, AX
    MOVQ R15, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ BX, CX
    MULXQ ·qElement+8(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+16(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+24(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+32(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+40(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+48(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+56(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+64(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+72(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qElement+80(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ R14, R13
    MULXQ ·qElement+88(SB), AX, R14
    ADOXQ AX, R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R14
    ADOXQ AX, R14
    XORQ DX, DX
    MOVQ CX, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R15
    ADCXQ CX, AX
    MOVQ R15, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ BX, CX
    MULXQ ·qElement+8(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+16(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+24(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+32(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+40(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+48(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+56(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+64(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+72(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qElement+80(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ R14, R13
    MULXQ ·qElement+88(SB), AX, R14
    ADOXQ AX, R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R14
    ADOXQ AX, R14
    XORQ DX, DX
    MOVQ CX, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R15
    ADCXQ CX, AX
    MOVQ R15, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ BX, CX
    MULXQ ·qElement+8(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+16(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+24(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+32(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+40(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+48(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+56(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+64(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+72(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qElement+80(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ R14, R13
    MULXQ ·qElement+88(SB), AX, R14
    ADOXQ AX, R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R14
    ADOXQ AX, R14
    XORQ DX, DX
    MOVQ CX, DX
    MULXQ ·qElementInv0(SB), DX, AX                        // m := t[0]*q'[0] mod W
    XORQ AX, AX
    // C,_ := t[0] + m*q[0]
    MULXQ ·qElement+0(SB), AX, R15
    ADCXQ CX, AX
    MOVQ R15, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    ADCXQ BX, CX
    MULXQ ·qElement+8(SB), AX, BX
    ADOXQ AX, CX
    ADCXQ BP, BX
    MULXQ ·qElement+16(SB), AX, BP
    ADOXQ AX, BX
    ADCXQ SI, BP
    MULXQ ·qElement+24(SB), AX, SI
    ADOXQ AX, BP
    ADCXQ DI, SI
    MULXQ ·qElement+32(SB), AX, DI
    ADOXQ AX, SI
    ADCXQ R8, DI
    MULXQ ·qElement+40(SB), AX, R8
    ADOXQ AX, DI
    ADCXQ R9, R8
    MULXQ ·qElement+48(SB), AX, R9
    ADOXQ AX, R8
    ADCXQ R10, R9
    MULXQ ·qElement+56(SB), AX, R10
    ADOXQ AX, R9
    ADCXQ R11, R10
    MULXQ ·qElement+64(SB), AX, R11
    ADOXQ AX, R10
    ADCXQ R12, R11
    MULXQ ·qElement+72(SB), AX, R12
    ADOXQ AX, R11
    ADCXQ R13, R12
    MULXQ ·qElement+80(SB), AX, R13
    ADOXQ AX, R12
    ADCXQ R14, R13
    MULXQ ·qElement+88(SB), AX, R14
    ADOXQ AX, R13
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R14
    ADOXQ AX, R14
    MOVQ res+0(FP), R15
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
    MOVQ CX, 0(R15)
    MOVQ BX, 8(R15)
    MOVQ BP, 16(R15)
    MOVQ SI, 24(R15)
    MOVQ DI, 32(R15)
    MOVQ R8, 40(R15)
    MOVQ R9, 48(R15)
    MOVQ R10, 56(R15)
    MOVQ R11, 64(R15)
    MOVQ R12, 72(R15)
    MOVQ R13, 80(R15)
    MOVQ R14, 88(R15)
    RET
no_adx:
    MOVQ res+0(FP), AX
    MOVQ AX, (SP)
CALL ·_fromMontGenericElement(SB)
    RET

TEXT ·reduceElement(SB), $96-8
    MOVQ res+0(FP), CX
    MOVQ 0(CX), BX
    MOVQ 8(CX), BP
    MOVQ 16(CX), SI
    MOVQ 24(CX), DI
    MOVQ 32(CX), R8
    MOVQ 40(CX), R9
    MOVQ 48(CX), R10
    MOVQ 56(CX), R11
    MOVQ 64(CX), R12
    MOVQ 72(CX), R13
    MOVQ 80(CX), R14
    MOVQ 88(CX), R15
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

TEXT ·addElement(SB), $96-24
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
    MOVQ y+16(FP), AX
    ADDQ 0(AX), CX
    ADCQ 8(AX), BX
    ADCQ 16(AX), BP
    ADCQ 24(AX), SI
    ADCQ 32(AX), DI
    ADCQ 40(AX), R8
    ADCQ 48(AX), R9
    ADCQ 56(AX), R10
    ADCQ 64(AX), R11
    ADCQ 72(AX), R12
    ADCQ 80(AX), R13
    ADCQ 88(AX), R14
    // note that we don't check for the carry here, as this code was generated assuming F.NoCarry condition is set
    // (see goff for more details)
    MOVQ res+0(FP), AX
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
    MOVQ CX, 0(AX)
    MOVQ BX, 8(AX)
    MOVQ BP, 16(AX)
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

TEXT ·doubleElement(SB), $96-16
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
    // note that we don't check for the carry here, as this code was generated assuming F.NoCarry condition is set
    // (see goff for more details)
    MOVQ res+0(FP), AX
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
    MOVQ CX, 0(AX)
    MOVQ BX, 8(AX)
    MOVQ BP, 16(AX)
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
