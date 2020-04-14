#include "textflag.h"
// func squareElement(res,y *Element)
TEXT ·squareElement(SB), NOSPLIT, $0-16
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

	// if adx and mulx instructions are not available, uses MUL algorithm.
	
    CMPB ·supportAdx(SB), $0x0000000000000001             // check if we support MULX and ADOX instructions
    JNE no_adx                                            // no support for MULX or ADOX instructions
    MOVQ y+8(FP), DI                                       // dereference y
    // outter loop 0
    XORQ AX, AX                                            // clear up flags
    // dx = y[0]
    MOVQ 0(DI), DX
    MULXQ 8(DI), R9, R10
    MULXQ 16(DI), AX, R11
    ADCXQ AX, R10
    MULXQ 24(DI), AX, R8
    ADCXQ AX, R11
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R8
    XORQ AX, AX                                            // clear up flags
    MULXQ DX, CX, DX
    ADCXQ R9, R9
    MOVQ R9, BX
    ADOXQ DX, BX
    ADCXQ R10, R10
    MOVQ R10, BP
    ADOXQ AX, BP
    ADCXQ R11, R11
    MOVQ R11, SI
    ADOXQ AX, SI
    ADCXQ R8, R8
    ADOXQ AX, R8
    MOVQ $0xc2e1f593efffffff, DX
    MULXQ CX, R12, DX
    XORQ DX, DX                                            // clear up flags
    MOVQ $0x43e1f593f0000001, DX
    MULXQ R12, AX, DX
    ADCXQ CX, AX
    MOVQ DX, CX
    MOVQ $0x2833e84879b97091, DX
    ADCXQ BX, CX
    MULXQ R12, AX, BX
    ADOXQ AX, CX
    MOVQ $0xb85045b68181585d, DX
    ADCXQ BP, BX
    MULXQ R12, AX, BP
    ADOXQ AX, BX
    MOVQ $0x30644e72e131a029, DX
    ADCXQ SI, BP
    MULXQ R12, AX, SI
    ADOXQ AX, BP
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, SI
    ADOXQ R8, SI
    // outter loop 1
    XORQ AX, AX                                            // clear up flags
    // dx = y[1]
    MOVQ 8(DI), DX
    MULXQ 16(DI), R13, R14
    MULXQ 24(DI), AX, R8
    ADCXQ AX, R14
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R8
    XORQ AX, AX                                            // clear up flags
    ADCXQ R13, R13
    ADOXQ R13, BP
    ADCXQ R14, R14
    ADOXQ R14, SI
    ADCXQ R8, R8
    ADOXQ AX, R8
    XORQ AX, AX                                            // clear up flags
    MULXQ DX, AX, DX
    ADOXQ AX, BX
    MOVQ $0x0000000000000000, AX
    ADOXQ DX, BP
    ADOXQ AX, SI
    ADOXQ AX, R8
    MOVQ $0xc2e1f593efffffff, DX
    MULXQ CX, R15, DX
    XORQ DX, DX                                            // clear up flags
    MOVQ $0x43e1f593f0000001, DX
    MULXQ R15, AX, DX
    ADCXQ CX, AX
    MOVQ DX, CX
    MOVQ $0x2833e84879b97091, DX
    ADCXQ BX, CX
    MULXQ R15, AX, BX
    ADOXQ AX, CX
    MOVQ $0xb85045b68181585d, DX
    ADCXQ BP, BX
    MULXQ R15, AX, BP
    ADOXQ AX, BX
    MOVQ $0x30644e72e131a029, DX
    ADCXQ SI, BP
    MULXQ R15, AX, SI
    ADOXQ AX, BP
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, SI
    ADOXQ R8, SI
    // outter loop 2
    XORQ AX, AX                                            // clear up flags
    // dx = y[2]
    MOVQ 16(DI), DX
    MULXQ 24(DI), R9, R8
    ADCXQ R9, R9
    ADOXQ R9, SI
    ADCXQ R8, R8
    ADOXQ AX, R8
    XORQ AX, AX                                            // clear up flags
    MULXQ DX, AX, DX
    ADOXQ AX, BP
    MOVQ $0x0000000000000000, AX
    ADOXQ DX, SI
    ADOXQ AX, R8
    MOVQ $0xc2e1f593efffffff, DX
    MULXQ CX, R10, DX
    XORQ DX, DX                                            // clear up flags
    MOVQ $0x43e1f593f0000001, DX
    MULXQ R10, AX, DX
    ADCXQ CX, AX
    MOVQ DX, CX
    MOVQ $0x2833e84879b97091, DX
    ADCXQ BX, CX
    MULXQ R10, AX, BX
    ADOXQ AX, CX
    MOVQ $0xb85045b68181585d, DX
    ADCXQ BP, BX
    MULXQ R10, AX, BP
    ADOXQ AX, BX
    MOVQ $0x30644e72e131a029, DX
    ADCXQ SI, BP
    MULXQ R10, AX, SI
    ADOXQ AX, BP
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, SI
    ADOXQ R8, SI
    // outter loop 3
    XORQ AX, AX                                            // clear up flags
    // dx = y[3]
    MOVQ 24(DI), DX
    MULXQ DX, AX, R8
    ADCXQ AX, SI
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, R8
    MOVQ $0xc2e1f593efffffff, DX
    MULXQ CX, R11, DX
    XORQ DX, DX                                            // clear up flags
    MOVQ $0x43e1f593f0000001, DX
    MULXQ R11, AX, DX
    ADCXQ CX, AX
    MOVQ DX, CX
    MOVQ $0x2833e84879b97091, DX
    ADCXQ BX, CX
    MULXQ R11, AX, BX
    ADOXQ AX, CX
    MOVQ $0xb85045b68181585d, DX
    ADCXQ BP, BX
    MULXQ R11, AX, BP
    ADOXQ AX, BX
    MOVQ $0x30644e72e131a029, DX
    ADCXQ SI, BP
    MULXQ R11, AX, SI
    ADOXQ AX, BP
    MOVQ $0x0000000000000000, AX
    ADCXQ AX, SI
    ADOXQ R8, SI
    // dereference res
    MOVQ res+0(FP), R12
reduce:
    MOVQ $0x30644e72e131a029, DX
    CMPQ SI, DX                                            // note: this is not constant time, comment out to have constant time mul
    JCC sub_t_q                                           // t > q
t_is_smaller:
    MOVQ CX, 0(R12)
    MOVQ BX, 8(R12)
    MOVQ BP, 16(R12)
    MOVQ SI, 24(R12)
    RET
sub_t_q:
    MOVQ CX, R13
    MOVQ $0x43e1f593f0000001, DX
    SUBQ DX, R13
    MOVQ BX, R14
    MOVQ $0x2833e84879b97091, DX
    SBBQ DX, R14
    MOVQ BP, R15
    MOVQ $0xb85045b68181585d, DX
    SBBQ DX, R15
    MOVQ SI, R9
    MOVQ $0x30644e72e131a029, DX
    SBBQ DX, R9
    JCS t_is_smaller
    MOVQ R13, 0(R12)
    MOVQ R14, 8(R12)
    MOVQ R15, 16(R12)
    MOVQ R9, 24(R12)
    RET
no_adx:
    // dereference y
    MOVQ y+8(FP), R13
    MOVQ 0(R13), AX
    MOVQ 0(R13), R11
    MULQ R11
    MOVQ AX, CX
    MOVQ DX, DI
    MOVQ $0xc2e1f593efffffff, R8
    IMULQ CX, R8
    MOVQ $0x43e1f593f0000001, AX
    MULQ R8
    ADDQ CX, AX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R10
    MOVQ 8(R13), AX
    MULQ R11
    MOVQ DI, BX
    ADDQ AX, BX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, DI
    MOVQ $0x2833e84879b97091, AX
    MULQ R8
    ADDQ BX, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, CX
    MOVQ DX, R10
    MOVQ 16(R13), AX
    MULQ R11
    MOVQ DI, BP
    ADDQ AX, BP
    ADCQ $0x0000000000000000, DX
    MOVQ DX, DI
    MOVQ $0xb85045b68181585d, AX
    MULQ R8
    ADDQ BP, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, BX
    MOVQ DX, R10
    MOVQ 24(R13), AX
    MULQ R11
    MOVQ DI, SI
    ADDQ AX, SI
    ADCQ $0x0000000000000000, DX
    MOVQ DX, DI
    MOVQ $0x30644e72e131a029, AX
    MULQ R8
    ADDQ SI, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, BP
    MOVQ DX, R10
    ADDQ R10, DI
    MOVQ DI, SI
    MOVQ 0(R13), AX
    MOVQ 8(R13), R11
    MULQ R11
    ADDQ AX, CX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, DI
    MOVQ $0xc2e1f593efffffff, R8
    IMULQ CX, R8
    MOVQ $0x43e1f593f0000001, AX
    MULQ R8
    ADDQ CX, AX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R10
    MOVQ 8(R13), AX
    MULQ R11
    ADDQ DI, BX
    ADCQ $0x0000000000000000, DX
    ADDQ AX, BX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, DI
    MOVQ $0x2833e84879b97091, AX
    MULQ R8
    ADDQ BX, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, CX
    MOVQ DX, R10
    MOVQ 16(R13), AX
    MULQ R11
    ADDQ DI, BP
    ADCQ $0x0000000000000000, DX
    ADDQ AX, BP
    ADCQ $0x0000000000000000, DX
    MOVQ DX, DI
    MOVQ $0xb85045b68181585d, AX
    MULQ R8
    ADDQ BP, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, BX
    MOVQ DX, R10
    MOVQ 24(R13), AX
    MULQ R11
    ADDQ DI, SI
    ADCQ $0x0000000000000000, DX
    ADDQ AX, SI
    ADCQ $0x0000000000000000, DX
    MOVQ DX, DI
    MOVQ $0x30644e72e131a029, AX
    MULQ R8
    ADDQ SI, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, BP
    MOVQ DX, R10
    ADDQ R10, DI
    MOVQ DI, SI
    MOVQ 0(R13), AX
    MOVQ 16(R13), R11
    MULQ R11
    ADDQ AX, CX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, DI
    MOVQ $0xc2e1f593efffffff, R8
    IMULQ CX, R8
    MOVQ $0x43e1f593f0000001, AX
    MULQ R8
    ADDQ CX, AX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R10
    MOVQ 8(R13), AX
    MULQ R11
    ADDQ DI, BX
    ADCQ $0x0000000000000000, DX
    ADDQ AX, BX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, DI
    MOVQ $0x2833e84879b97091, AX
    MULQ R8
    ADDQ BX, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, CX
    MOVQ DX, R10
    MOVQ 16(R13), AX
    MULQ R11
    ADDQ DI, BP
    ADCQ $0x0000000000000000, DX
    ADDQ AX, BP
    ADCQ $0x0000000000000000, DX
    MOVQ DX, DI
    MOVQ $0xb85045b68181585d, AX
    MULQ R8
    ADDQ BP, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, BX
    MOVQ DX, R10
    MOVQ 24(R13), AX
    MULQ R11
    ADDQ DI, SI
    ADCQ $0x0000000000000000, DX
    ADDQ AX, SI
    ADCQ $0x0000000000000000, DX
    MOVQ DX, DI
    MOVQ $0x30644e72e131a029, AX
    MULQ R8
    ADDQ SI, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, BP
    MOVQ DX, R10
    ADDQ R10, DI
    MOVQ DI, SI
    MOVQ 0(R13), AX
    MOVQ 24(R13), R11
    MULQ R11
    ADDQ AX, CX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, DI
    MOVQ $0xc2e1f593efffffff, R8
    IMULQ CX, R8
    MOVQ $0x43e1f593f0000001, AX
    MULQ R8
    ADDQ CX, AX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, R10
    MOVQ 8(R13), AX
    MULQ R11
    ADDQ DI, BX
    ADCQ $0x0000000000000000, DX
    ADDQ AX, BX
    ADCQ $0x0000000000000000, DX
    MOVQ DX, DI
    MOVQ $0x2833e84879b97091, AX
    MULQ R8
    ADDQ BX, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, CX
    MOVQ DX, R10
    MOVQ 16(R13), AX
    MULQ R11
    ADDQ DI, BP
    ADCQ $0x0000000000000000, DX
    ADDQ AX, BP
    ADCQ $0x0000000000000000, DX
    MOVQ DX, DI
    MOVQ $0xb85045b68181585d, AX
    MULQ R8
    ADDQ BP, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, BX
    MOVQ DX, R10
    MOVQ 24(R13), AX
    MULQ R11
    ADDQ DI, SI
    ADCQ $0x0000000000000000, DX
    ADDQ AX, SI
    ADCQ $0x0000000000000000, DX
    MOVQ DX, DI
    MOVQ $0x30644e72e131a029, AX
    MULQ R8
    ADDQ SI, R10
    ADCQ $0x0000000000000000, DX
    ADDQ AX, R10
    ADCQ $0x0000000000000000, DX
    MOVQ R10, BP
    MOVQ DX, R10
    ADDQ R10, DI
    MOVQ DI, SI
    // dereference res
    MOVQ res+0(FP), R12
    JMP reduce
