#include "textflag.h"

// func mulAssignElement(res,y *Element)
// montgomery multiplication of res by y 
// stores the result in res
TEXT ·mulAssignElement(SB), NOSPLIT, $0-16
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

    MOVQ res+0(FP), R9                                     // dereference x
    CMPB ·supportAdx(SB), $0x0000000000000001             // check if we support MULX and ADOX instructions
    JNE no_adx                                            // no support for MULX or ADOX instructions
    MOVQ y+8(FP), R12                                      // dereference y
    MOVQ 0(R9), R13                                        // R13 = x[0]
    MOVQ 8(R9), R14                                        // R14 = x[1]
    MOVQ 16(R9), R15                                       // R15 = x[2]
    // outter loop 0
    XORQ DX, DX                                            // clear up flags
    MOVQ 0(R12), DX                                        // DX = y[0]
    MULXQ R13, CX, BX                                       // t[0], t[1] = y[0] * x[0]
    MULXQ R14, AX, BP
    ADOXQ AX, BX
    MULXQ R15, AX, SI
    ADOXQ AX, BP
    MULXQ 24(R9), AX, DI
    ADOXQ AX, SI
    MULXQ 32(R9), AX, R8
    ADOXQ AX, DI
    MULXQ 40(R9), AX, R11
    ADOXQ AX, R8
    // add the last carries to R11
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R11
    ADOXQ DX, R11
    MOVQ $0x8508bfffffffffff, DX
    MULXQ CX, R10, DX                                       // m := t[0]*q'[0] mod W
    XORQ DX, DX                                            // clear the flags
    // C,_ := t[0] + m*q[0]
    MOVQ $0x8508c00000000001, DX
    MULXQ R10, AX, DX
    ADCXQ CX, AX
    MOVQ DX, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    MOVQ $0x170b5d4430000000, DX
    ADCXQ BX, CX
    MULXQ R10, AX, BX
    ADOXQ AX, CX
    MOVQ $0x1ef3622fba094800, DX
    ADCXQ BP, BX
    MULXQ R10, AX, BP
    ADOXQ AX, BX
    MOVQ $0x1a22d9f300f5138f, DX
    ADCXQ SI, BP
    MULXQ R10, AX, SI
    ADOXQ AX, BP
    MOVQ $0xc63b05c06ca1493b, DX
    ADCXQ DI, SI
    MULXQ R10, AX, DI
    ADOXQ AX, SI
    MOVQ $0x01ae3a4617c510ea, DX
    ADCXQ R8, DI
    MULXQ R10, AX, R8
    ADOXQ AX, DI
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R8
    ADOXQ R11, R8
    // outter loop 1
    XORQ DX, DX                                            // clear up flags
    MOVQ 8(R12), DX                                        // DX = y[1]
    MULXQ R13, AX, R11
    ADOXQ AX, CX
    ADCXQ R11, BX                                           // t[1] += regA
    MULXQ R14, AX, R11
    ADOXQ AX, BX
    ADCXQ R11, BP                                           // t[2] += regA
    MULXQ R15, AX, R11
    ADOXQ AX, BP
    ADCXQ R11, SI                                           // t[3] += regA
    MULXQ 24(R9), AX, R11
    ADOXQ AX, SI
    ADCXQ R11, DI                                           // t[4] += regA
    MULXQ 32(R9), AX, R11
    ADOXQ AX, DI
    ADCXQ R11, R8                                           // t[5] += regA
    MULXQ 40(R9), AX, R11
    ADOXQ AX, R8
    // add the last carries to R11
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R11
    ADOXQ DX, R11
    MOVQ $0x8508bfffffffffff, DX
    MULXQ CX, R10, DX                                       // m := t[0]*q'[0] mod W
    XORQ DX, DX                                            // clear the flags
    // C,_ := t[0] + m*q[0]
    MOVQ $0x8508c00000000001, DX
    MULXQ R10, AX, DX
    ADCXQ CX, AX
    MOVQ DX, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    MOVQ $0x170b5d4430000000, DX
    ADCXQ BX, CX
    MULXQ R10, AX, BX
    ADOXQ AX, CX
    MOVQ $0x1ef3622fba094800, DX
    ADCXQ BP, BX
    MULXQ R10, AX, BP
    ADOXQ AX, BX
    MOVQ $0x1a22d9f300f5138f, DX
    ADCXQ SI, BP
    MULXQ R10, AX, SI
    ADOXQ AX, BP
    MOVQ $0xc63b05c06ca1493b, DX
    ADCXQ DI, SI
    MULXQ R10, AX, DI
    ADOXQ AX, SI
    MOVQ $0x01ae3a4617c510ea, DX
    ADCXQ R8, DI
    MULXQ R10, AX, R8
    ADOXQ AX, DI
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R8
    ADOXQ R11, R8
    // outter loop 2
    XORQ DX, DX                                            // clear up flags
    MOVQ 16(R12), DX                                       // DX = y[2]
    MULXQ R13, AX, R11
    ADOXQ AX, CX
    ADCXQ R11, BX                                           // t[1] += regA
    MULXQ R14, AX, R11
    ADOXQ AX, BX
    ADCXQ R11, BP                                           // t[2] += regA
    MULXQ R15, AX, R11
    ADOXQ AX, BP
    ADCXQ R11, SI                                           // t[3] += regA
    MULXQ 24(R9), AX, R11
    ADOXQ AX, SI
    ADCXQ R11, DI                                           // t[4] += regA
    MULXQ 32(R9), AX, R11
    ADOXQ AX, DI
    ADCXQ R11, R8                                           // t[5] += regA
    MULXQ 40(R9), AX, R11
    ADOXQ AX, R8
    // add the last carries to R11
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R11
    ADOXQ DX, R11
    MOVQ $0x8508bfffffffffff, DX
    MULXQ CX, R10, DX                                       // m := t[0]*q'[0] mod W
    XORQ DX, DX                                            // clear the flags
    // C,_ := t[0] + m*q[0]
    MOVQ $0x8508c00000000001, DX
    MULXQ R10, AX, DX
    ADCXQ CX, AX
    MOVQ DX, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    MOVQ $0x170b5d4430000000, DX
    ADCXQ BX, CX
    MULXQ R10, AX, BX
    ADOXQ AX, CX
    MOVQ $0x1ef3622fba094800, DX
    ADCXQ BP, BX
    MULXQ R10, AX, BP
    ADOXQ AX, BX
    MOVQ $0x1a22d9f300f5138f, DX
    ADCXQ SI, BP
    MULXQ R10, AX, SI
    ADOXQ AX, BP
    MOVQ $0xc63b05c06ca1493b, DX
    ADCXQ DI, SI
    MULXQ R10, AX, DI
    ADOXQ AX, SI
    MOVQ $0x01ae3a4617c510ea, DX
    ADCXQ R8, DI
    MULXQ R10, AX, R8
    ADOXQ AX, DI
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R8
    ADOXQ R11, R8
    // outter loop 3
    XORQ DX, DX                                            // clear up flags
    MOVQ 24(R12), DX                                       // DX = y[3]
    MULXQ R13, AX, R11
    ADOXQ AX, CX
    ADCXQ R11, BX                                           // t[1] += regA
    MULXQ R14, AX, R11
    ADOXQ AX, BX
    ADCXQ R11, BP                                           // t[2] += regA
    MULXQ R15, AX, R11
    ADOXQ AX, BP
    ADCXQ R11, SI                                           // t[3] += regA
    MULXQ 24(R9), AX, R11
    ADOXQ AX, SI
    ADCXQ R11, DI                                           // t[4] += regA
    MULXQ 32(R9), AX, R11
    ADOXQ AX, DI
    ADCXQ R11, R8                                           // t[5] += regA
    MULXQ 40(R9), AX, R11
    ADOXQ AX, R8
    // add the last carries to R11
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R11
    ADOXQ DX, R11
    MOVQ $0x8508bfffffffffff, DX
    MULXQ CX, R10, DX                                       // m := t[0]*q'[0] mod W
    XORQ DX, DX                                            // clear the flags
    // C,_ := t[0] + m*q[0]
    MOVQ $0x8508c00000000001, DX
    MULXQ R10, AX, DX
    ADCXQ CX, AX
    MOVQ DX, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    MOVQ $0x170b5d4430000000, DX
    ADCXQ BX, CX
    MULXQ R10, AX, BX
    ADOXQ AX, CX
    MOVQ $0x1ef3622fba094800, DX
    ADCXQ BP, BX
    MULXQ R10, AX, BP
    ADOXQ AX, BX
    MOVQ $0x1a22d9f300f5138f, DX
    ADCXQ SI, BP
    MULXQ R10, AX, SI
    ADOXQ AX, BP
    MOVQ $0xc63b05c06ca1493b, DX
    ADCXQ DI, SI
    MULXQ R10, AX, DI
    ADOXQ AX, SI
    MOVQ $0x01ae3a4617c510ea, DX
    ADCXQ R8, DI
    MULXQ R10, AX, R8
    ADOXQ AX, DI
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R8
    ADOXQ R11, R8
    // outter loop 4
    XORQ DX, DX                                            // clear up flags
    MOVQ 32(R12), DX                                       // DX = y[4]
    MULXQ R13, AX, R11
    ADOXQ AX, CX
    ADCXQ R11, BX                                           // t[1] += regA
    MULXQ R14, AX, R11
    ADOXQ AX, BX
    ADCXQ R11, BP                                           // t[2] += regA
    MULXQ R15, AX, R11
    ADOXQ AX, BP
    ADCXQ R11, SI                                           // t[3] += regA
    MULXQ 24(R9), AX, R11
    ADOXQ AX, SI
    ADCXQ R11, DI                                           // t[4] += regA
    MULXQ 32(R9), AX, R11
    ADOXQ AX, DI
    ADCXQ R11, R8                                           // t[5] += regA
    MULXQ 40(R9), AX, R11
    ADOXQ AX, R8
    // add the last carries to R11
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R11
    ADOXQ DX, R11
    MOVQ $0x8508bfffffffffff, DX
    MULXQ CX, R10, DX                                       // m := t[0]*q'[0] mod W
    XORQ DX, DX                                            // clear the flags
    // C,_ := t[0] + m*q[0]
    MOVQ $0x8508c00000000001, DX
    MULXQ R10, AX, DX
    ADCXQ CX, AX
    MOVQ DX, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    MOVQ $0x170b5d4430000000, DX
    ADCXQ BX, CX
    MULXQ R10, AX, BX
    ADOXQ AX, CX
    MOVQ $0x1ef3622fba094800, DX
    ADCXQ BP, BX
    MULXQ R10, AX, BP
    ADOXQ AX, BX
    MOVQ $0x1a22d9f300f5138f, DX
    ADCXQ SI, BP
    MULXQ R10, AX, SI
    ADOXQ AX, BP
    MOVQ $0xc63b05c06ca1493b, DX
    ADCXQ DI, SI
    MULXQ R10, AX, DI
    ADOXQ AX, SI
    MOVQ $0x01ae3a4617c510ea, DX
    ADCXQ R8, DI
    MULXQ R10, AX, R8
    ADOXQ AX, DI
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R8
    ADOXQ R11, R8
    // outter loop 5
    XORQ DX, DX                                            // clear up flags
    MOVQ 40(R12), DX                                       // DX = y[5]
    MULXQ R13, AX, R11
    ADOXQ AX, CX
    ADCXQ R11, BX                                           // t[1] += regA
    MULXQ R14, AX, R11
    ADOXQ AX, BX
    ADCXQ R11, BP                                           // t[2] += regA
    MULXQ R15, AX, R11
    ADOXQ AX, BP
    ADCXQ R11, SI                                           // t[3] += regA
    MULXQ 24(R9), AX, R11
    ADOXQ AX, SI
    ADCXQ R11, DI                                           // t[4] += regA
    MULXQ 32(R9), AX, R11
    ADOXQ AX, DI
    ADCXQ R11, R8                                           // t[5] += regA
    MULXQ 40(R9), AX, R11
    ADOXQ AX, R8
    // add the last carries to R11
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R11
    ADOXQ DX, R11
    MOVQ $0x8508bfffffffffff, DX
    MULXQ CX, R10, DX                                       // m := t[0]*q'[0] mod W
    XORQ DX, DX                                            // clear the flags
    // C,_ := t[0] + m*q[0]
    MOVQ $0x8508c00000000001, DX
    MULXQ R10, AX, DX
    ADCXQ CX, AX
    MOVQ DX, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    MOVQ $0x170b5d4430000000, DX
    ADCXQ BX, CX
    MULXQ R10, AX, BX
    ADOXQ AX, CX
    MOVQ $0x1ef3622fba094800, DX
    ADCXQ BP, BX
    MULXQ R10, AX, BP
    ADOXQ AX, BX
    MOVQ $0x1a22d9f300f5138f, DX
    ADCXQ SI, BP
    MULXQ R10, AX, SI
    ADOXQ AX, BP
    MOVQ $0xc63b05c06ca1493b, DX
    ADCXQ DI, SI
    MULXQ R10, AX, DI
    ADOXQ AX, SI
    MOVQ $0x01ae3a4617c510ea, DX
    ADCXQ R8, DI
    MULXQ R10, AX, R8
    ADOXQ AX, DI
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R8
    ADOXQ R11, R8
reduce:
    MOVQ $0x01ae3a4617c510ea, DX
    CMPQ R8, DX                                            // note: this is not constant time, comment out to have constant time mul
    JCC sub_t_q                                           // t > q
t_is_smaller:
    MOVQ CX, 0(R9)
    MOVQ BX, 8(R9)
    MOVQ BP, 16(R9)
    MOVQ SI, 24(R9)
    MOVQ DI, 32(R9)
    MOVQ R8, 40(R9)
    RET
sub_t_q:
    MOVQ CX, R10
    MOVQ $0x8508c00000000001, DX
    SUBQ DX, R10
    MOVQ BX, R12
    MOVQ $0x170b5d4430000000, DX
    SBBQ DX, R12
    MOVQ BP, R11
    MOVQ $0x1ef3622fba094800, DX
    SBBQ DX, R11
    MOVQ SI, R13
    MOVQ $0x1a22d9f300f5138f, DX
    SBBQ DX, R13
    MOVQ DI, R14
    MOVQ $0xc63b05c06ca1493b, DX
    SBBQ DX, R14
    MOVQ R8, R15
    MOVQ $0x01ae3a4617c510ea, DX
    SBBQ DX, R15
    JCS t_is_smaller
    MOVQ R10, 0(R9)
    MOVQ R12, 8(R9)
    MOVQ R11, 16(R9)
    MOVQ R13, 24(R9)
    MOVQ R14, 32(R9)
    MOVQ R15, 40(R9)
    RET
no_adx:
    MOVQ y+8(FP), R14                                      // dereference y
    MOVQ 0(R9), AX
    MOVQ 0(R14), R12
    MULQ R12
    MOVQ AX, CX
    MOVQ DX, R11
    MOVQ $0x8508bfffffffffff, R13
    IMULQ CX, R13
    MOVQ $0x8508c00000000001, AX
    MULQ R13
    ADDQ CX, AX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R10
    MOVQ 8(R9), AX
    MULQ R12
    MOVQ R11, BX
    ADDQ AX, BX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0x170b5d4430000000, AX
    MULQ R13
    ADDQ BX, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, CX
    MOVQ DX, R10
    MOVQ 16(R9), AX
    MULQ R12
    MOVQ R11, BP
    ADDQ AX, BP
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0x1ef3622fba094800, AX
    MULQ R13
    ADDQ BP, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, BX
    MOVQ DX, R10
    MOVQ 24(R9), AX
    MULQ R12
    MOVQ R11, SI
    ADDQ AX, SI
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0x1a22d9f300f5138f, AX
    MULQ R13
    ADDQ SI, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, BP
    MOVQ DX, R10
    MOVQ 32(R9), AX
    MULQ R12
    MOVQ R11, DI
    ADDQ AX, DI
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0xc63b05c06ca1493b, AX
    MULQ R13
    ADDQ DI, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, SI
    MOVQ DX, R10
    MOVQ 40(R9), AX
    MULQ R12
    MOVQ R11, R8
    ADDQ AX, R8
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0x01ae3a4617c510ea, AX
    MULQ R13
    ADDQ R8, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, DI
    MOVQ DX, R10
    ADDQ R10, R11
    MOVQ R11, R8
    MOVQ 0(R9), AX
    MOVQ 8(R14), R12
    MULQ R12
    ADDQ AX, CX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0x8508bfffffffffff, R13
    IMULQ CX, R13
    MOVQ $0x8508c00000000001, AX
    MULQ R13
    ADDQ CX, AX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R10
    MOVQ 8(R9), AX
    MULQ R12
    ADDQ R11, BX
    ADCQ $0x0000000000000000, DX
    ADDQ AX, BX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0x170b5d4430000000, AX
    MULQ R13
    ADDQ BX, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, CX
    MOVQ DX, R10
    MOVQ 16(R9), AX
    MULQ R12
    ADDQ R11, BP
    ADCQ $0x0000000000000000, DX
    ADDQ AX, BP
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0x1ef3622fba094800, AX
    MULQ R13
    ADDQ BP, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, BX
    MOVQ DX, R10
    MOVQ 24(R9), AX
    MULQ R12
    ADDQ R11, SI
    ADCQ $0x0000000000000000, DX
    ADDQ AX, SI
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0x1a22d9f300f5138f, AX
    MULQ R13
    ADDQ SI, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, BP
    MOVQ DX, R10
    MOVQ 32(R9), AX
    MULQ R12
    ADDQ R11, DI
    ADCQ $0x0000000000000000, DX
    ADDQ AX, DI
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0xc63b05c06ca1493b, AX
    MULQ R13
    ADDQ DI, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, SI
    MOVQ DX, R10
    MOVQ 40(R9), AX
    MULQ R12
    ADDQ R11, R8
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R8
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0x01ae3a4617c510ea, AX
    MULQ R13
    ADDQ R8, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, DI
    MOVQ DX, R10
    ADDQ R10, R11
    MOVQ R11, R8
    MOVQ 0(R9), AX
    MOVQ 16(R14), R12
    MULQ R12
    ADDQ AX, CX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0x8508bfffffffffff, R13
    IMULQ CX, R13
    MOVQ $0x8508c00000000001, AX
    MULQ R13
    ADDQ CX, AX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R10
    MOVQ 8(R9), AX
    MULQ R12
    ADDQ R11, BX
    ADCQ $0x0000000000000000, DX
    ADDQ AX, BX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0x170b5d4430000000, AX
    MULQ R13
    ADDQ BX, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, CX
    MOVQ DX, R10
    MOVQ 16(R9), AX
    MULQ R12
    ADDQ R11, BP
    ADCQ $0x0000000000000000, DX
    ADDQ AX, BP
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0x1ef3622fba094800, AX
    MULQ R13
    ADDQ BP, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, BX
    MOVQ DX, R10
    MOVQ 24(R9), AX
    MULQ R12
    ADDQ R11, SI
    ADCQ $0x0000000000000000, DX
    ADDQ AX, SI
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0x1a22d9f300f5138f, AX
    MULQ R13
    ADDQ SI, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, BP
    MOVQ DX, R10
    MOVQ 32(R9), AX
    MULQ R12
    ADDQ R11, DI
    ADCQ $0x0000000000000000, DX
    ADDQ AX, DI
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0xc63b05c06ca1493b, AX
    MULQ R13
    ADDQ DI, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, SI
    MOVQ DX, R10
    MOVQ 40(R9), AX
    MULQ R12
    ADDQ R11, R8
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R8
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0x01ae3a4617c510ea, AX
    MULQ R13
    ADDQ R8, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, DI
    MOVQ DX, R10
    ADDQ R10, R11
    MOVQ R11, R8
    MOVQ 0(R9), AX
    MOVQ 24(R14), R12
    MULQ R12
    ADDQ AX, CX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0x8508bfffffffffff, R13
    IMULQ CX, R13
    MOVQ $0x8508c00000000001, AX
    MULQ R13
    ADDQ CX, AX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R10
    MOVQ 8(R9), AX
    MULQ R12
    ADDQ R11, BX
    ADCQ $0x0000000000000000, DX
    ADDQ AX, BX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0x170b5d4430000000, AX
    MULQ R13
    ADDQ BX, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, CX
    MOVQ DX, R10
    MOVQ 16(R9), AX
    MULQ R12
    ADDQ R11, BP
    ADCQ $0x0000000000000000, DX
    ADDQ AX, BP
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0x1ef3622fba094800, AX
    MULQ R13
    ADDQ BP, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, BX
    MOVQ DX, R10
    MOVQ 24(R9), AX
    MULQ R12
    ADDQ R11, SI
    ADCQ $0x0000000000000000, DX
    ADDQ AX, SI
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0x1a22d9f300f5138f, AX
    MULQ R13
    ADDQ SI, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, BP
    MOVQ DX, R10
    MOVQ 32(R9), AX
    MULQ R12
    ADDQ R11, DI
    ADCQ $0x0000000000000000, DX
    ADDQ AX, DI
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0xc63b05c06ca1493b, AX
    MULQ R13
    ADDQ DI, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, SI
    MOVQ DX, R10
    MOVQ 40(R9), AX
    MULQ R12
    ADDQ R11, R8
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R8
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0x01ae3a4617c510ea, AX
    MULQ R13
    ADDQ R8, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, DI
    MOVQ DX, R10
    ADDQ R10, R11
    MOVQ R11, R8
    MOVQ 0(R9), AX
    MOVQ 32(R14), R12
    MULQ R12
    ADDQ AX, CX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0x8508bfffffffffff, R13
    IMULQ CX, R13
    MOVQ $0x8508c00000000001, AX
    MULQ R13
    ADDQ CX, AX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R10
    MOVQ 8(R9), AX
    MULQ R12
    ADDQ R11, BX
    ADCQ $0x0000000000000000, DX
    ADDQ AX, BX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0x170b5d4430000000, AX
    MULQ R13
    ADDQ BX, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, CX
    MOVQ DX, R10
    MOVQ 16(R9), AX
    MULQ R12
    ADDQ R11, BP
    ADCQ $0x0000000000000000, DX
    ADDQ AX, BP
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0x1ef3622fba094800, AX
    MULQ R13
    ADDQ BP, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, BX
    MOVQ DX, R10
    MOVQ 24(R9), AX
    MULQ R12
    ADDQ R11, SI
    ADCQ $0x0000000000000000, DX
    ADDQ AX, SI
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0x1a22d9f300f5138f, AX
    MULQ R13
    ADDQ SI, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, BP
    MOVQ DX, R10
    MOVQ 32(R9), AX
    MULQ R12
    ADDQ R11, DI
    ADCQ $0x0000000000000000, DX
    ADDQ AX, DI
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0xc63b05c06ca1493b, AX
    MULQ R13
    ADDQ DI, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, SI
    MOVQ DX, R10
    MOVQ 40(R9), AX
    MULQ R12
    ADDQ R11, R8
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R8
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0x01ae3a4617c510ea, AX
    MULQ R13
    ADDQ R8, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, DI
    MOVQ DX, R10
    ADDQ R10, R11
    MOVQ R11, R8
    MOVQ 0(R9), AX
    MOVQ 40(R14), R12
    MULQ R12
    ADDQ AX, CX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0x8508bfffffffffff, R13
    IMULQ CX, R13
    MOVQ $0x8508c00000000001, AX
    MULQ R13
    ADDQ CX, AX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R10
    MOVQ 8(R9), AX
    MULQ R12
    ADDQ R11, BX
    ADCQ $0x0000000000000000, DX
    ADDQ AX, BX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0x170b5d4430000000, AX
    MULQ R13
    ADDQ BX, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, CX
    MOVQ DX, R10
    MOVQ 16(R9), AX
    MULQ R12
    ADDQ R11, BP
    ADCQ $0x0000000000000000, DX
    ADDQ AX, BP
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0x1ef3622fba094800, AX
    MULQ R13
    ADDQ BP, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, BX
    MOVQ DX, R10
    MOVQ 24(R9), AX
    MULQ R12
    ADDQ R11, SI
    ADCQ $0x0000000000000000, DX
    ADDQ AX, SI
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0x1a22d9f300f5138f, AX
    MULQ R13
    ADDQ SI, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, BP
    MOVQ DX, R10
    MOVQ 32(R9), AX
    MULQ R12
    ADDQ R11, DI
    ADCQ $0x0000000000000000, DX
    ADDQ AX, DI
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0xc63b05c06ca1493b, AX
    MULQ R13
    ADDQ DI, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, SI
    MOVQ DX, R10
    MOVQ 40(R9), AX
    MULQ R12
    ADDQ R11, R8
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R8
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0x01ae3a4617c510ea, AX
    MULQ R13
    ADDQ R8, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, DI
    MOVQ DX, R10
    ADDQ R10, R11
    MOVQ R11, R8
    JMP reduce


// func fromMontElement(res *Element)
// montgomery multiplication of res by 1 
// stores the result in res
TEXT ·fromMontElement(SB), NOSPLIT, $0-8
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


    MOVQ res+0(FP), R9                                     // dereference x
    MOVQ 0(R9), CX                                         // t[0] = x[0]
    MOVQ 8(R9), BX                                         // t[1] = x[1]
    MOVQ 16(R9), BP                                        // t[2] = x[2]
    MOVQ 24(R9), SI                                        // t[3] = x[3]
    MOVQ 32(R9), DI                                        // t[4] = x[4]
    MOVQ 40(R9), R8                                        // t[5] = x[5]
    CMPB ·supportAdx(SB), $0x0000000000000001             // check if we support MULX and ADOX instructions
    JNE no_adx                                            // no support for MULX or ADOX instructions
    // outter loop 0
    XORQ DX, DX                                            // clear up flags
    MOVQ $0x8508bfffffffffff, DX
    MULXQ CX, R10, DX                                       // m := t[0]*q'[0] mod W
    XORQ DX, DX                                            // clear the flags
    // C,_ := t[0] + m*q[0]
    MOVQ $0x8508c00000000001, DX
    MULXQ R10, AX, DX
    ADCXQ CX, AX
    MOVQ DX, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    MOVQ $0x170b5d4430000000, DX
    ADCXQ BX, CX
    MULXQ R10, AX, BX
    ADOXQ AX, CX
    MOVQ $0x1ef3622fba094800, DX
    ADCXQ BP, BX
    MULXQ R10, AX, BP
    ADOXQ AX, BX
    MOVQ $0x1a22d9f300f5138f, DX
    ADCXQ SI, BP
    MULXQ R10, AX, SI
    ADOXQ AX, BP
    MOVQ $0xc63b05c06ca1493b, DX
    ADCXQ DI, SI
    MULXQ R10, AX, DI
    ADOXQ AX, SI
    MOVQ $0x01ae3a4617c510ea, DX
    ADCXQ R8, DI
    MULXQ R10, AX, R8
    ADOXQ AX, DI
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R8
    ADOXQ AX, R8
    // outter loop 1
    XORQ DX, DX                                            // clear up flags
    MOVQ $0x8508bfffffffffff, DX
    MULXQ CX, R10, DX                                       // m := t[0]*q'[0] mod W
    XORQ DX, DX                                            // clear the flags
    // C,_ := t[0] + m*q[0]
    MOVQ $0x8508c00000000001, DX
    MULXQ R10, AX, DX
    ADCXQ CX, AX
    MOVQ DX, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    MOVQ $0x170b5d4430000000, DX
    ADCXQ BX, CX
    MULXQ R10, AX, BX
    ADOXQ AX, CX
    MOVQ $0x1ef3622fba094800, DX
    ADCXQ BP, BX
    MULXQ R10, AX, BP
    ADOXQ AX, BX
    MOVQ $0x1a22d9f300f5138f, DX
    ADCXQ SI, BP
    MULXQ R10, AX, SI
    ADOXQ AX, BP
    MOVQ $0xc63b05c06ca1493b, DX
    ADCXQ DI, SI
    MULXQ R10, AX, DI
    ADOXQ AX, SI
    MOVQ $0x01ae3a4617c510ea, DX
    ADCXQ R8, DI
    MULXQ R10, AX, R8
    ADOXQ AX, DI
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R8
    ADOXQ AX, R8
    // outter loop 2
    XORQ DX, DX                                            // clear up flags
    MOVQ $0x8508bfffffffffff, DX
    MULXQ CX, R10, DX                                       // m := t[0]*q'[0] mod W
    XORQ DX, DX                                            // clear the flags
    // C,_ := t[0] + m*q[0]
    MOVQ $0x8508c00000000001, DX
    MULXQ R10, AX, DX
    ADCXQ CX, AX
    MOVQ DX, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    MOVQ $0x170b5d4430000000, DX
    ADCXQ BX, CX
    MULXQ R10, AX, BX
    ADOXQ AX, CX
    MOVQ $0x1ef3622fba094800, DX
    ADCXQ BP, BX
    MULXQ R10, AX, BP
    ADOXQ AX, BX
    MOVQ $0x1a22d9f300f5138f, DX
    ADCXQ SI, BP
    MULXQ R10, AX, SI
    ADOXQ AX, BP
    MOVQ $0xc63b05c06ca1493b, DX
    ADCXQ DI, SI
    MULXQ R10, AX, DI
    ADOXQ AX, SI
    MOVQ $0x01ae3a4617c510ea, DX
    ADCXQ R8, DI
    MULXQ R10, AX, R8
    ADOXQ AX, DI
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R8
    ADOXQ AX, R8
    // outter loop 3
    XORQ DX, DX                                            // clear up flags
    MOVQ $0x8508bfffffffffff, DX
    MULXQ CX, R10, DX                                       // m := t[0]*q'[0] mod W
    XORQ DX, DX                                            // clear the flags
    // C,_ := t[0] + m*q[0]
    MOVQ $0x8508c00000000001, DX
    MULXQ R10, AX, DX
    ADCXQ CX, AX
    MOVQ DX, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    MOVQ $0x170b5d4430000000, DX
    ADCXQ BX, CX
    MULXQ R10, AX, BX
    ADOXQ AX, CX
    MOVQ $0x1ef3622fba094800, DX
    ADCXQ BP, BX
    MULXQ R10, AX, BP
    ADOXQ AX, BX
    MOVQ $0x1a22d9f300f5138f, DX
    ADCXQ SI, BP
    MULXQ R10, AX, SI
    ADOXQ AX, BP
    MOVQ $0xc63b05c06ca1493b, DX
    ADCXQ DI, SI
    MULXQ R10, AX, DI
    ADOXQ AX, SI
    MOVQ $0x01ae3a4617c510ea, DX
    ADCXQ R8, DI
    MULXQ R10, AX, R8
    ADOXQ AX, DI
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R8
    ADOXQ AX, R8
    // outter loop 4
    XORQ DX, DX                                            // clear up flags
    MOVQ $0x8508bfffffffffff, DX
    MULXQ CX, R10, DX                                       // m := t[0]*q'[0] mod W
    XORQ DX, DX                                            // clear the flags
    // C,_ := t[0] + m*q[0]
    MOVQ $0x8508c00000000001, DX
    MULXQ R10, AX, DX
    ADCXQ CX, AX
    MOVQ DX, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    MOVQ $0x170b5d4430000000, DX
    ADCXQ BX, CX
    MULXQ R10, AX, BX
    ADOXQ AX, CX
    MOVQ $0x1ef3622fba094800, DX
    ADCXQ BP, BX
    MULXQ R10, AX, BP
    ADOXQ AX, BX
    MOVQ $0x1a22d9f300f5138f, DX
    ADCXQ SI, BP
    MULXQ R10, AX, SI
    ADOXQ AX, BP
    MOVQ $0xc63b05c06ca1493b, DX
    ADCXQ DI, SI
    MULXQ R10, AX, DI
    ADOXQ AX, SI
    MOVQ $0x01ae3a4617c510ea, DX
    ADCXQ R8, DI
    MULXQ R10, AX, R8
    ADOXQ AX, DI
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R8
    ADOXQ AX, R8
    // outter loop 5
    XORQ DX, DX                                            // clear up flags
    MOVQ $0x8508bfffffffffff, DX
    MULXQ CX, R10, DX                                       // m := t[0]*q'[0] mod W
    XORQ DX, DX                                            // clear the flags
    // C,_ := t[0] + m*q[0]
    MOVQ $0x8508c00000000001, DX
    MULXQ R10, AX, DX
    ADCXQ CX, AX
    MOVQ DX, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    MOVQ $0x170b5d4430000000, DX
    ADCXQ BX, CX
    MULXQ R10, AX, BX
    ADOXQ AX, CX
    MOVQ $0x1ef3622fba094800, DX
    ADCXQ BP, BX
    MULXQ R10, AX, BP
    ADOXQ AX, BX
    MOVQ $0x1a22d9f300f5138f, DX
    ADCXQ SI, BP
    MULXQ R10, AX, SI
    ADOXQ AX, BP
    MOVQ $0xc63b05c06ca1493b, DX
    ADCXQ DI, SI
    MULXQ R10, AX, DI
    ADOXQ AX, SI
    MOVQ $0x01ae3a4617c510ea, DX
    ADCXQ R8, DI
    MULXQ R10, AX, R8
    ADOXQ AX, DI
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R8
    ADOXQ AX, R8
reduce:
    MOVQ $0x01ae3a4617c510ea, DX
    CMPQ R8, DX                                            // note: this is not constant time, comment out to have constant time mul
    JCC sub_t_q                                           // t > q
t_is_smaller:
    MOVQ CX, 0(R9)
    MOVQ BX, 8(R9)
    MOVQ BP, 16(R9)
    MOVQ SI, 24(R9)
    MOVQ DI, 32(R9)
    MOVQ R8, 40(R9)
    RET
sub_t_q:
    MOVQ CX, R11
    MOVQ $0x8508c00000000001, DX
    SUBQ DX, R11
    MOVQ BX, R12
    MOVQ $0x170b5d4430000000, DX
    SBBQ DX, R12
    MOVQ BP, R13
    MOVQ $0x1ef3622fba094800, DX
    SBBQ DX, R13
    MOVQ SI, R14
    MOVQ $0x1a22d9f300f5138f, DX
    SBBQ DX, R14
    MOVQ DI, R15
    MOVQ $0xc63b05c06ca1493b, DX
    SBBQ DX, R15
    MOVQ R8, R10
    MOVQ $0x01ae3a4617c510ea, DX
    SBBQ DX, R10
    JCS t_is_smaller
    MOVQ R11, 0(R9)
    MOVQ R12, 8(R9)
    MOVQ R13, 16(R9)
    MOVQ R14, 24(R9)
    MOVQ R15, 32(R9)
    MOVQ R10, 40(R9)
    RET
no_adx:
    MOVQ $0x8508bfffffffffff, R14
    IMULQ CX, R14
    MOVQ $0x8508c00000000001, AX
    MULQ R14
    ADDQ CX, AX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0x170b5d4430000000, AX
    MULQ R14
    ADDQ BX, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, CX
    MOVQ DX, R11
    MOVQ $0x1ef3622fba094800, AX
    MULQ R14
    ADDQ BP, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, BX
    MOVQ DX, R11
    MOVQ $0x1a22d9f300f5138f, AX
    MULQ R14
    ADDQ SI, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, BP
    MOVQ DX, R11
    MOVQ $0xc63b05c06ca1493b, AX
    MULQ R14
    ADDQ DI, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, SI
    MOVQ DX, R11
    MOVQ $0x01ae3a4617c510ea, AX
    MULQ R14
    ADDQ R8, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, DI
    MOVQ DX, R11
    MOVQ R11, R8
    MOVQ $0x8508bfffffffffff, R14
    IMULQ CX, R14
    MOVQ $0x8508c00000000001, AX
    MULQ R14
    ADDQ CX, AX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0x170b5d4430000000, AX
    MULQ R14
    ADDQ BX, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, CX
    MOVQ DX, R11
    MOVQ $0x1ef3622fba094800, AX
    MULQ R14
    ADDQ BP, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, BX
    MOVQ DX, R11
    MOVQ $0x1a22d9f300f5138f, AX
    MULQ R14
    ADDQ SI, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, BP
    MOVQ DX, R11
    MOVQ $0xc63b05c06ca1493b, AX
    MULQ R14
    ADDQ DI, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, SI
    MOVQ DX, R11
    MOVQ $0x01ae3a4617c510ea, AX
    MULQ R14
    ADDQ R8, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, DI
    MOVQ DX, R11
    MOVQ R11, R8
    MOVQ $0x8508bfffffffffff, R14
    IMULQ CX, R14
    MOVQ $0x8508c00000000001, AX
    MULQ R14
    ADDQ CX, AX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0x170b5d4430000000, AX
    MULQ R14
    ADDQ BX, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, CX
    MOVQ DX, R11
    MOVQ $0x1ef3622fba094800, AX
    MULQ R14
    ADDQ BP, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, BX
    MOVQ DX, R11
    MOVQ $0x1a22d9f300f5138f, AX
    MULQ R14
    ADDQ SI, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, BP
    MOVQ DX, R11
    MOVQ $0xc63b05c06ca1493b, AX
    MULQ R14
    ADDQ DI, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, SI
    MOVQ DX, R11
    MOVQ $0x01ae3a4617c510ea, AX
    MULQ R14
    ADDQ R8, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, DI
    MOVQ DX, R11
    MOVQ R11, R8
    MOVQ $0x8508bfffffffffff, R14
    IMULQ CX, R14
    MOVQ $0x8508c00000000001, AX
    MULQ R14
    ADDQ CX, AX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0x170b5d4430000000, AX
    MULQ R14
    ADDQ BX, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, CX
    MOVQ DX, R11
    MOVQ $0x1ef3622fba094800, AX
    MULQ R14
    ADDQ BP, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, BX
    MOVQ DX, R11
    MOVQ $0x1a22d9f300f5138f, AX
    MULQ R14
    ADDQ SI, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, BP
    MOVQ DX, R11
    MOVQ $0xc63b05c06ca1493b, AX
    MULQ R14
    ADDQ DI, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, SI
    MOVQ DX, R11
    MOVQ $0x01ae3a4617c510ea, AX
    MULQ R14
    ADDQ R8, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, DI
    MOVQ DX, R11
    MOVQ R11, R8
    MOVQ $0x8508bfffffffffff, R14
    IMULQ CX, R14
    MOVQ $0x8508c00000000001, AX
    MULQ R14
    ADDQ CX, AX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0x170b5d4430000000, AX
    MULQ R14
    ADDQ BX, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, CX
    MOVQ DX, R11
    MOVQ $0x1ef3622fba094800, AX
    MULQ R14
    ADDQ BP, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, BX
    MOVQ DX, R11
    MOVQ $0x1a22d9f300f5138f, AX
    MULQ R14
    ADDQ SI, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, BP
    MOVQ DX, R11
    MOVQ $0xc63b05c06ca1493b, AX
    MULQ R14
    ADDQ DI, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, SI
    MOVQ DX, R11
    MOVQ $0x01ae3a4617c510ea, AX
    MULQ R14
    ADDQ R8, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, DI
    MOVQ DX, R11
    MOVQ R11, R8
    MOVQ $0x8508bfffffffffff, R14
    IMULQ CX, R14
    MOVQ $0x8508c00000000001, AX
    MULQ R14
    ADDQ CX, AX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ $0x170b5d4430000000, AX
    MULQ R14
    ADDQ BX, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, CX
    MOVQ DX, R11
    MOVQ $0x1ef3622fba094800, AX
    MULQ R14
    ADDQ BP, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, BX
    MOVQ DX, R11
    MOVQ $0x1a22d9f300f5138f, AX
    MULQ R14
    ADDQ SI, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, BP
    MOVQ DX, R11
    MOVQ $0xc63b05c06ca1493b, AX
    MULQ R14
    ADDQ DI, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, SI
    MOVQ DX, R11
    MOVQ $0x01ae3a4617c510ea, AX
    MULQ R14
    ADDQ R8, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, DI
    MOVQ DX, R11
    MOVQ R11, R8
    JMP reduce


// func reduceElement(res *Element)
TEXT ·reduceElement(SB), NOSPLIT, $0-8
	// test purposes

    MOVQ res+0(FP), R9                                     // dereference x
    MOVQ 0(R9), CX                                         // t[0] = x[0]
    MOVQ 8(R9), BX                                         // t[1] = x[1]
    MOVQ 16(R9), BP                                        // t[2] = x[2]
    MOVQ 24(R9), SI                                        // t[3] = x[3]
    MOVQ 32(R9), DI                                        // t[4] = x[4]
    MOVQ 40(R9), R8                                        // t[5] = x[5]
reduce:
    MOVQ $0x01ae3a4617c510ea, DX
    CMPQ R8, DX                                            // note: this is not constant time, comment out to have constant time mul
    JCC sub_t_q                                           // t > q
t_is_smaller:
    MOVQ CX, 0(R9)
    MOVQ BX, 8(R9)
    MOVQ BP, 16(R9)
    MOVQ SI, 24(R9)
    MOVQ DI, 32(R9)
    MOVQ R8, 40(R9)
    RET
sub_t_q:
    MOVQ CX, R10
    MOVQ $0x8508c00000000001, DX
    SUBQ DX, R10
    MOVQ BX, R11
    MOVQ $0x170b5d4430000000, DX
    SBBQ DX, R11
    MOVQ BP, R12
    MOVQ $0x1ef3622fba094800, DX
    SBBQ DX, R12
    MOVQ SI, R13
    MOVQ $0x1a22d9f300f5138f, DX
    SBBQ DX, R13
    MOVQ DI, R14
    MOVQ $0xc63b05c06ca1493b, DX
    SBBQ DX, R14
    MOVQ R8, R15
    MOVQ $0x01ae3a4617c510ea, DX
    SBBQ DX, R15
    JCS t_is_smaller
    MOVQ R10, 0(R9)
    MOVQ R11, 8(R9)
    MOVQ R12, 16(R9)
    MOVQ R13, 24(R9)
    MOVQ R14, 32(R9)
    MOVQ R15, 40(R9)
    RET
