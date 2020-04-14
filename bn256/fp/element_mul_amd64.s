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

    MOVQ res+0(FP), DI                                     // dereference x
    CMPB ·supportAdx(SB), $0x0000000000000001             // check if we support MULX and ADOX instructions
    JNE no_adx                                            // no support for MULX or ADOX instructions
    MOVQ y+8(FP), R10                                      // dereference y
    MOVQ 0(DI), R11                                        // R11 = x[0]
    MOVQ 8(DI), R12                                        // R12 = x[1]
    MOVQ 16(DI), R13                                       // R13 = x[2]
    MOVQ 24(DI), R14                                       // R14 = x[3]
    // outter loop 0
    XORQ DX, DX                                            // clear up flags
    MOVQ 0(R10), DX                                        // DX = y[0]
    MULXQ R11, CX, BX                                       // t[0], t[1] = y[0] * x[0]
    MULXQ R12, AX, BP
    ADOXQ AX, BX
    MULXQ R13, AX, SI
    ADOXQ AX, BP
    MULXQ R14, AX, R9
    ADOXQ AX, SI
    // add the last carries to R9
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R9
    ADOXQ DX, R9
    MOVQ $0x87d20782e4866389, DX
    MULXQ CX, R8, DX                                        // m := t[0]*q'[0] mod W
    XORQ DX, DX                                            // clear the flags
    // C,_ := t[0] + m*q[0]
    MOVQ $0x3c208c16d87cfd47, DX
    MULXQ R8, AX, DX
    ADCXQ CX, AX
    MOVQ DX, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    MOVQ $0x97816a916871ca8d, DX
    ADCXQ BX, CX
    MULXQ R8, AX, BX
    ADOXQ AX, CX
    MOVQ $0xb85045b68181585d, DX
    ADCXQ BP, BX
    MULXQ R8, AX, BP
    ADOXQ AX, BX
    MOVQ $0x30644e72e131a029, DX
    ADCXQ SI, BP
    MULXQ R8, AX, SI
    ADOXQ AX, BP
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, SI
    ADOXQ R9, SI
    // outter loop 1
    XORQ DX, DX                                            // clear up flags
    MOVQ 8(R10), DX                                        // DX = y[1]
    MULXQ R11, AX, R9
    ADOXQ AX, CX
    ADCXQ R9, BX                                            // t[1] += regA
    MULXQ R12, AX, R9
    ADOXQ AX, BX
    ADCXQ R9, BP                                            // t[2] += regA
    MULXQ R13, AX, R9
    ADOXQ AX, BP
    ADCXQ R9, SI                                            // t[3] += regA
    MULXQ R14, AX, R9
    ADOXQ AX, SI
    // add the last carries to R9
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R9
    ADOXQ DX, R9
    MOVQ $0x87d20782e4866389, DX
    MULXQ CX, R8, DX                                        // m := t[0]*q'[0] mod W
    XORQ DX, DX                                            // clear the flags
    // C,_ := t[0] + m*q[0]
    MOVQ $0x3c208c16d87cfd47, DX
    MULXQ R8, AX, DX
    ADCXQ CX, AX
    MOVQ DX, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    MOVQ $0x97816a916871ca8d, DX
    ADCXQ BX, CX
    MULXQ R8, AX, BX
    ADOXQ AX, CX
    MOVQ $0xb85045b68181585d, DX
    ADCXQ BP, BX
    MULXQ R8, AX, BP
    ADOXQ AX, BX
    MOVQ $0x30644e72e131a029, DX
    ADCXQ SI, BP
    MULXQ R8, AX, SI
    ADOXQ AX, BP
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, SI
    ADOXQ R9, SI
    // outter loop 2
    XORQ DX, DX                                            // clear up flags
    MOVQ 16(R10), DX                                       // DX = y[2]
    MULXQ R11, AX, R9
    ADOXQ AX, CX
    ADCXQ R9, BX                                            // t[1] += regA
    MULXQ R12, AX, R9
    ADOXQ AX, BX
    ADCXQ R9, BP                                            // t[2] += regA
    MULXQ R13, AX, R9
    ADOXQ AX, BP
    ADCXQ R9, SI                                            // t[3] += regA
    MULXQ R14, AX, R9
    ADOXQ AX, SI
    // add the last carries to R9
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R9
    ADOXQ DX, R9
    MOVQ $0x87d20782e4866389, DX
    MULXQ CX, R8, DX                                        // m := t[0]*q'[0] mod W
    XORQ DX, DX                                            // clear the flags
    // C,_ := t[0] + m*q[0]
    MOVQ $0x3c208c16d87cfd47, DX
    MULXQ R8, AX, DX
    ADCXQ CX, AX
    MOVQ DX, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    MOVQ $0x97816a916871ca8d, DX
    ADCXQ BX, CX
    MULXQ R8, AX, BX
    ADOXQ AX, CX
    MOVQ $0xb85045b68181585d, DX
    ADCXQ BP, BX
    MULXQ R8, AX, BP
    ADOXQ AX, BX
    MOVQ $0x30644e72e131a029, DX
    ADCXQ SI, BP
    MULXQ R8, AX, SI
    ADOXQ AX, BP
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, SI
    ADOXQ R9, SI
    // outter loop 3
    XORQ DX, DX                                            // clear up flags
    MOVQ 24(R10), DX                                       // DX = y[3]
    MULXQ R11, AX, R9
    ADOXQ AX, CX
    ADCXQ R9, BX                                            // t[1] += regA
    MULXQ R12, AX, R9
    ADOXQ AX, BX
    ADCXQ R9, BP                                            // t[2] += regA
    MULXQ R13, AX, R9
    ADOXQ AX, BP
    ADCXQ R9, SI                                            // t[3] += regA
    MULXQ R14, AX, R9
    ADOXQ AX, SI
    // add the last carries to R9
    MOVQ $0x0000000000000000, DX
    ADCXQ DX, R9
    ADOXQ DX, R9
    MOVQ $0x87d20782e4866389, DX
    MULXQ CX, R8, DX                                        // m := t[0]*q'[0] mod W
    XORQ DX, DX                                            // clear the flags
    // C,_ := t[0] + m*q[0]
    MOVQ $0x3c208c16d87cfd47, DX
    MULXQ R8, AX, DX
    ADCXQ CX, AX
    MOVQ DX, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    MOVQ $0x97816a916871ca8d, DX
    ADCXQ BX, CX
    MULXQ R8, AX, BX
    ADOXQ AX, CX
    MOVQ $0xb85045b68181585d, DX
    ADCXQ BP, BX
    MULXQ R8, AX, BP
    ADOXQ AX, BX
    MOVQ $0x30644e72e131a029, DX
    ADCXQ SI, BP
    MULXQ R8, AX, SI
    ADOXQ AX, BP
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, SI
    ADOXQ R9, SI
reduce:
    MOVQ $0x30644e72e131a029, DX
    CMPQ SI, DX                                            // note: this is not constant time, comment out to have constant time mul
    JCC sub_t_q                                           // t > q
t_is_smaller:
    MOVQ CX, 0(DI)
    MOVQ BX, 8(DI)
    MOVQ BP, 16(DI)
    MOVQ SI, 24(DI)
    RET
sub_t_q:
    MOVQ CX, R15
    MOVQ $0x3c208c16d87cfd47, DX
    SUBQ DX, R15
    MOVQ BX, R8
    MOVQ $0x97816a916871ca8d, DX
    SBBQ DX, R8
    MOVQ BP, R10
    MOVQ $0xb85045b68181585d, DX
    SBBQ DX, R10
    MOVQ SI, R9
    MOVQ $0x30644e72e131a029, DX
    SBBQ DX, R9
    JCS t_is_smaller
    MOVQ R15, 0(DI)
    MOVQ R8, 8(DI)
    MOVQ R10, 16(DI)
    MOVQ R9, 24(DI)
    RET
no_adx:
    MOVQ y+8(FP), R15                                      // dereference y
    MOVQ 0(DI), AX
    MOVQ 0(R15), R12
    MULQ R12
    MOVQ AX, CX
    MOVQ DX, R13
    MOVQ $0x87d20782e4866389, R14
    IMULQ CX, R14
    MOVQ $0x3c208c16d87cfd47, AX
    MULQ R14
    ADDQ CX, AX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ 8(DI), AX
    MULQ R12
    MOVQ R13, BX
    ADDQ AX, BX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R13
    MOVQ $0x97816a916871ca8d, AX
    MULQ R14
    ADDQ BX, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, CX
    MOVQ DX, R11
    MOVQ 16(DI), AX
    MULQ R12
    MOVQ R13, BP
    ADDQ AX, BP
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R13
    MOVQ $0xb85045b68181585d, AX
    MULQ R14
    ADDQ BP, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, BX
    MOVQ DX, R11
    MOVQ 24(DI), AX
    MULQ R12
    MOVQ R13, SI
    ADDQ AX, SI
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R13
    MOVQ $0x30644e72e131a029, AX
    MULQ R14
    ADDQ SI, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, BP
    MOVQ DX, R11
    ADDQ R11, R13
    MOVQ R13, SI
    MOVQ 0(DI), AX
    MOVQ 8(R15), R12
    MULQ R12
    ADDQ AX, CX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R13
    MOVQ $0x87d20782e4866389, R14
    IMULQ CX, R14
    MOVQ $0x3c208c16d87cfd47, AX
    MULQ R14
    ADDQ CX, AX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ 8(DI), AX
    MULQ R12
    ADDQ R13, BX
    ADCQ $0x0000000000000000, DX
    ADDQ AX, BX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R13
    MOVQ $0x97816a916871ca8d, AX
    MULQ R14
    ADDQ BX, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, CX
    MOVQ DX, R11
    MOVQ 16(DI), AX
    MULQ R12
    ADDQ R13, BP
    ADCQ $0x0000000000000000, DX
    ADDQ AX, BP
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R13
    MOVQ $0xb85045b68181585d, AX
    MULQ R14
    ADDQ BP, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, BX
    MOVQ DX, R11
    MOVQ 24(DI), AX
    MULQ R12
    ADDQ R13, SI
    ADCQ $0x0000000000000000, DX
    ADDQ AX, SI
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R13
    MOVQ $0x30644e72e131a029, AX
    MULQ R14
    ADDQ SI, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, BP
    MOVQ DX, R11
    ADDQ R11, R13
    MOVQ R13, SI
    MOVQ 0(DI), AX
    MOVQ 16(R15), R12
    MULQ R12
    ADDQ AX, CX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R13
    MOVQ $0x87d20782e4866389, R14
    IMULQ CX, R14
    MOVQ $0x3c208c16d87cfd47, AX
    MULQ R14
    ADDQ CX, AX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ 8(DI), AX
    MULQ R12
    ADDQ R13, BX
    ADCQ $0x0000000000000000, DX
    ADDQ AX, BX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R13
    MOVQ $0x97816a916871ca8d, AX
    MULQ R14
    ADDQ BX, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, CX
    MOVQ DX, R11
    MOVQ 16(DI), AX
    MULQ R12
    ADDQ R13, BP
    ADCQ $0x0000000000000000, DX
    ADDQ AX, BP
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R13
    MOVQ $0xb85045b68181585d, AX
    MULQ R14
    ADDQ BP, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, BX
    MOVQ DX, R11
    MOVQ 24(DI), AX
    MULQ R12
    ADDQ R13, SI
    ADCQ $0x0000000000000000, DX
    ADDQ AX, SI
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R13
    MOVQ $0x30644e72e131a029, AX
    MULQ R14
    ADDQ SI, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, BP
    MOVQ DX, R11
    ADDQ R11, R13
    MOVQ R13, SI
    MOVQ 0(DI), AX
    MOVQ 24(R15), R12
    MULQ R12
    ADDQ AX, CX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R13
    MOVQ $0x87d20782e4866389, R14
    IMULQ CX, R14
    MOVQ $0x3c208c16d87cfd47, AX
    MULQ R14
    ADDQ CX, AX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R11
    MOVQ 8(DI), AX
    MULQ R12
    ADDQ R13, BX
    ADCQ $0x0000000000000000, DX
    ADDQ AX, BX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R13
    MOVQ $0x97816a916871ca8d, AX
    MULQ R14
    ADDQ BX, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, CX
    MOVQ DX, R11
    MOVQ 16(DI), AX
    MULQ R12
    ADDQ R13, BP
    ADCQ $0x0000000000000000, DX
    ADDQ AX, BP
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R13
    MOVQ $0xb85045b68181585d, AX
    MULQ R14
    ADDQ BP, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, BX
    MOVQ DX, R11
    MOVQ 24(DI), AX
    MULQ R12
    ADDQ R13, SI
    ADCQ $0x0000000000000000, DX
    ADDQ AX, SI
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R13
    MOVQ $0x30644e72e131a029, AX
    MULQ R14
    ADDQ SI, R11
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R11
    ADCQ $0x0000000000000000, DX
    MOVQ R11, BP
    MOVQ DX, R11
    ADDQ R11, R13
    MOVQ R13, SI
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


    MOVQ res+0(FP), DI                                     // dereference x
    MOVQ 0(DI), CX                                         // t[0] = x[0]
    MOVQ 8(DI), BX                                         // t[1] = x[1]
    MOVQ 16(DI), BP                                        // t[2] = x[2]
    MOVQ 24(DI), SI                                        // t[3] = x[3]
    CMPB ·supportAdx(SB), $0x0000000000000001             // check if we support MULX and ADOX instructions
    JNE no_adx                                            // no support for MULX or ADOX instructions
    // outter loop 0
    XORQ DX, DX                                            // clear up flags
    MOVQ $0x87d20782e4866389, DX
    MULXQ CX, R8, DX                                        // m := t[0]*q'[0] mod W
    XORQ DX, DX                                            // clear the flags
    // C,_ := t[0] + m*q[0]
    MOVQ $0x3c208c16d87cfd47, DX
    MULXQ R8, AX, DX
    ADCXQ CX, AX
    MOVQ DX, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    MOVQ $0x97816a916871ca8d, DX
    ADCXQ BX, CX
    MULXQ R8, AX, BX
    ADOXQ AX, CX
    MOVQ $0xb85045b68181585d, DX
    ADCXQ BP, BX
    MULXQ R8, AX, BP
    ADOXQ AX, BX
    MOVQ $0x30644e72e131a029, DX
    ADCXQ SI, BP
    MULXQ R8, AX, SI
    ADOXQ AX, BP
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, SI
    ADOXQ AX, SI
    // outter loop 1
    XORQ DX, DX                                            // clear up flags
    MOVQ $0x87d20782e4866389, DX
    MULXQ CX, R8, DX                                        // m := t[0]*q'[0] mod W
    XORQ DX, DX                                            // clear the flags
    // C,_ := t[0] + m*q[0]
    MOVQ $0x3c208c16d87cfd47, DX
    MULXQ R8, AX, DX
    ADCXQ CX, AX
    MOVQ DX, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    MOVQ $0x97816a916871ca8d, DX
    ADCXQ BX, CX
    MULXQ R8, AX, BX
    ADOXQ AX, CX
    MOVQ $0xb85045b68181585d, DX
    ADCXQ BP, BX
    MULXQ R8, AX, BP
    ADOXQ AX, BX
    MOVQ $0x30644e72e131a029, DX
    ADCXQ SI, BP
    MULXQ R8, AX, SI
    ADOXQ AX, BP
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, SI
    ADOXQ AX, SI
    // outter loop 2
    XORQ DX, DX                                            // clear up flags
    MOVQ $0x87d20782e4866389, DX
    MULXQ CX, R8, DX                                        // m := t[0]*q'[0] mod W
    XORQ DX, DX                                            // clear the flags
    // C,_ := t[0] + m*q[0]
    MOVQ $0x3c208c16d87cfd47, DX
    MULXQ R8, AX, DX
    ADCXQ CX, AX
    MOVQ DX, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    MOVQ $0x97816a916871ca8d, DX
    ADCXQ BX, CX
    MULXQ R8, AX, BX
    ADOXQ AX, CX
    MOVQ $0xb85045b68181585d, DX
    ADCXQ BP, BX
    MULXQ R8, AX, BP
    ADOXQ AX, BX
    MOVQ $0x30644e72e131a029, DX
    ADCXQ SI, BP
    MULXQ R8, AX, SI
    ADOXQ AX, BP
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, SI
    ADOXQ AX, SI
    // outter loop 3
    XORQ DX, DX                                            // clear up flags
    MOVQ $0x87d20782e4866389, DX
    MULXQ CX, R8, DX                                        // m := t[0]*q'[0] mod W
    XORQ DX, DX                                            // clear the flags
    // C,_ := t[0] + m*q[0]
    MOVQ $0x3c208c16d87cfd47, DX
    MULXQ R8, AX, DX
    ADCXQ CX, AX
    MOVQ DX, CX
    // for j=1 to N-1
    //     (C,t[j-1]) := t[j] + m*q[j] + C
    MOVQ $0x97816a916871ca8d, DX
    ADCXQ BX, CX
    MULXQ R8, AX, BX
    ADOXQ AX, CX
    MOVQ $0xb85045b68181585d, DX
    ADCXQ BP, BX
    MULXQ R8, AX, BP
    ADOXQ AX, BX
    MOVQ $0x30644e72e131a029, DX
    ADCXQ SI, BP
    MULXQ R8, AX, SI
    ADOXQ AX, BP
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, SI
    ADOXQ AX, SI
reduce:
    MOVQ $0x30644e72e131a029, DX
    CMPQ SI, DX                                            // note: this is not constant time, comment out to have constant time mul
    JCC sub_t_q                                           // t > q
t_is_smaller:
    MOVQ CX, 0(DI)
    MOVQ BX, 8(DI)
    MOVQ BP, 16(DI)
    MOVQ SI, 24(DI)
    RET
sub_t_q:
    MOVQ CX, R9
    MOVQ $0x3c208c16d87cfd47, DX
    SUBQ DX, R9
    MOVQ BX, R10
    MOVQ $0x97816a916871ca8d, DX
    SBBQ DX, R10
    MOVQ BP, R11
    MOVQ $0xb85045b68181585d, DX
    SBBQ DX, R11
    MOVQ SI, R12
    MOVQ $0x30644e72e131a029, DX
    SBBQ DX, R12
    JCS t_is_smaller
    MOVQ R9, 0(DI)
    MOVQ R10, 8(DI)
    MOVQ R11, 16(DI)
    MOVQ R12, 24(DI)
    RET
no_adx:
    MOVQ $0x87d20782e4866389, R8
    IMULQ CX, R8
    MOVQ $0x3c208c16d87cfd47, AX
    MULQ R8
    ADDQ CX, AX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R13
    MOVQ $0x97816a916871ca8d, AX
    MULQ R8
    ADDQ BX, R13
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R13
    ADCQ $0x0000000000000000, DX
    MOVQ R13, CX
    MOVQ DX, R13
    MOVQ $0xb85045b68181585d, AX
    MULQ R8
    ADDQ BP, R13
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R13
    ADCQ $0x0000000000000000, DX
    MOVQ R13, BX
    MOVQ DX, R13
    MOVQ $0x30644e72e131a029, AX
    MULQ R8
    ADDQ SI, R13
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R13
    ADCQ $0x0000000000000000, DX
    MOVQ R13, BP
    MOVQ DX, R13
    MOVQ R13, SI
    MOVQ $0x87d20782e4866389, R8
    IMULQ CX, R8
    MOVQ $0x3c208c16d87cfd47, AX
    MULQ R8
    ADDQ CX, AX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R13
    MOVQ $0x97816a916871ca8d, AX
    MULQ R8
    ADDQ BX, R13
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R13
    ADCQ $0x0000000000000000, DX
    MOVQ R13, CX
    MOVQ DX, R13
    MOVQ $0xb85045b68181585d, AX
    MULQ R8
    ADDQ BP, R13
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R13
    ADCQ $0x0000000000000000, DX
    MOVQ R13, BX
    MOVQ DX, R13
    MOVQ $0x30644e72e131a029, AX
    MULQ R8
    ADDQ SI, R13
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R13
    ADCQ $0x0000000000000000, DX
    MOVQ R13, BP
    MOVQ DX, R13
    MOVQ R13, SI
    MOVQ $0x87d20782e4866389, R8
    IMULQ CX, R8
    MOVQ $0x3c208c16d87cfd47, AX
    MULQ R8
    ADDQ CX, AX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R13
    MOVQ $0x97816a916871ca8d, AX
    MULQ R8
    ADDQ BX, R13
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R13
    ADCQ $0x0000000000000000, DX
    MOVQ R13, CX
    MOVQ DX, R13
    MOVQ $0xb85045b68181585d, AX
    MULQ R8
    ADDQ BP, R13
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R13
    ADCQ $0x0000000000000000, DX
    MOVQ R13, BX
    MOVQ DX, R13
    MOVQ $0x30644e72e131a029, AX
    MULQ R8
    ADDQ SI, R13
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R13
    ADCQ $0x0000000000000000, DX
    MOVQ R13, BP
    MOVQ DX, R13
    MOVQ R13, SI
    MOVQ $0x87d20782e4866389, R8
    IMULQ CX, R8
    MOVQ $0x3c208c16d87cfd47, AX
    MULQ R8
    ADDQ CX, AX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R13
    MOVQ $0x97816a916871ca8d, AX
    MULQ R8
    ADDQ BX, R13
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R13
    ADCQ $0x0000000000000000, DX
    MOVQ R13, CX
    MOVQ DX, R13
    MOVQ $0xb85045b68181585d, AX
    MULQ R8
    ADDQ BP, R13
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R13
    ADCQ $0x0000000000000000, DX
    MOVQ R13, BX
    MOVQ DX, R13
    MOVQ $0x30644e72e131a029, AX
    MULQ R8
    ADDQ SI, R13
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R13
    ADCQ $0x0000000000000000, DX
    MOVQ R13, BP
    MOVQ DX, R13
    MOVQ R13, SI
    JMP reduce


// func reduceElement(res *Element)
TEXT ·reduceElement(SB), NOSPLIT, $0-8
	// test purposes

    MOVQ res+0(FP), DI                                     // dereference x
    MOVQ 0(DI), CX                                         // t[0] = x[0]
    MOVQ 8(DI), BX                                         // t[1] = x[1]
    MOVQ 16(DI), BP                                        // t[2] = x[2]
    MOVQ 24(DI), SI                                        // t[3] = x[3]
reduce:
    MOVQ $0x30644e72e131a029, DX
    CMPQ SI, DX                                            // note: this is not constant time, comment out to have constant time mul
    JCC sub_t_q                                           // t > q
t_is_smaller:
    MOVQ CX, 0(DI)
    MOVQ BX, 8(DI)
    MOVQ BP, 16(DI)
    MOVQ SI, 24(DI)
    RET
sub_t_q:
    MOVQ CX, R8
    MOVQ $0x3c208c16d87cfd47, DX
    SUBQ DX, R8
    MOVQ BX, R9
    MOVQ $0x97816a916871ca8d, DX
    SBBQ DX, R9
    MOVQ BP, R10
    MOVQ $0xb85045b68181585d, DX
    SBBQ DX, R10
    MOVQ SI, R11
    MOVQ $0x30644e72e131a029, DX
    SBBQ DX, R11
    JCS t_is_smaller
    MOVQ R8, 0(DI)
    MOVQ R9, 8(DI)
    MOVQ R10, 16(DI)
    MOVQ R11, 24(DI)
    RET
