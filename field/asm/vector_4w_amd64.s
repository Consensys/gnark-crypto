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
	MOVQ    $const_q0, R11
	MOVQ    $const_q1, R12
	MOVQ    $const_q2, R13
	MOVQ    $const_q3, R14
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
	MOVQ  $const_qInvNeg, DX
	IMULQ R14, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ ·qElement+0(SB), AX, R13
	ADCXQ R14, AX
	MOVQ  R13, R14

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R15, R14
	MULXQ ·qElement+8(SB), AX, R15
	ADOXQ AX, R14

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ CX, R15
	MULXQ ·qElement+16(SB), AX, CX
	ADOXQ AX, R15

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BX, CX
	MULXQ ·qElement+24(SB), AX, BX
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
	MOVQ  $const_qInvNeg, DX
	IMULQ R14, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ ·qElement+0(SB), AX, R13
	ADCXQ R14, AX
	MOVQ  R13, R14

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R15, R14
	MULXQ ·qElement+8(SB), AX, R15
	ADOXQ AX, R14

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ CX, R15
	MULXQ ·qElement+16(SB), AX, CX
	ADOXQ AX, R15

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BX, CX
	MULXQ ·qElement+24(SB), AX, BX
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
	MOVQ  $const_qInvNeg, DX
	IMULQ R14, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ ·qElement+0(SB), AX, R13
	ADCXQ R14, AX
	MOVQ  R13, R14

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R15, R14
	MULXQ ·qElement+8(SB), AX, R15
	ADOXQ AX, R14

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ CX, R15
	MULXQ ·qElement+16(SB), AX, CX
	ADOXQ AX, R15

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BX, CX
	MULXQ ·qElement+24(SB), AX, BX
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
	MOVQ  $const_qInvNeg, DX
	IMULQ R14, DX

	// clear the flags
	XORQ AX, AX

	// C,_ := t[0] + m*q[0]
	MULXQ ·qElement+0(SB), AX, R13
	ADCXQ R14, AX
	MOVQ  R13, R14

	// (C,t[0]) := t[1] + m*q[1] + C
	ADCXQ R15, R14
	MULXQ ·qElement+8(SB), AX, R15
	ADOXQ AX, R14

	// (C,t[1]) := t[2] + m*q[2] + C
	ADCXQ CX, R15
	MULXQ ·qElement+16(SB), AX, CX
	ADOXQ AX, R15

	// (C,t[2]) := t[3] + m*q[3] + C
	ADCXQ BX, CX
	MULXQ ·qElement+24(SB), AX, BX
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

	// Derived from https://github.com/a16z/vectorized-fields
	// The idea is to use Z registers to accumulate the sum of elements, 8 by 8
	// first, we handle the case where n % 8 != 0
	// then, we loop over the elements 8 by 8 and accumulate the sum in the Z registers
	// finally, we reduce the sum and store it in res
	//
	// when we move an element of a into a Z register, we use VPMOVZXDQ
	// let's note w0...w3 the 4 64bits words of ai: w0 = ai[0], w1 = ai[1], w2 = ai[2], w3 = ai[3]
	// VPMOVZXDQ(ai, Z0) will result in
	// Z0= [hi(w3), lo(w3), hi(w2), lo(w2), hi(w1), lo(w1), hi(w0), lo(w0)]
	// with hi(wi) the high 32 bits of wi and lo(wi) the low 32 bits of wi
	// we can safely add 2^32+1 times Z registers constructed this way without overflow
	// since each of this lo/hi bits are moved into a "64bits" slot
	// N = 2^64-1 / 2^32-1 = 2^32+1
	//
	// we then propagate the carry using ADOXQ and ADCXQ
	// r0 = w0l + lo(woh)
	// r1 = carry + hi(woh) + w1l + lo(w1h)
	// r2 = carry + hi(w1h) + w2l + lo(w2h)
	// r3 = carry + hi(w2h) + w3l + lo(w3h)
	// r4 = carry + hi(w3h)
	// we then reduce the sum using a single-word Barrett reduction
	// we pick mu = 2^288 / q; which correspond to 4.5 words max.
	// meaning we must guarantee that r4 fits in 32bits.
	// To do so, we reduce N to 2^32-1 (since r4 receives 2 carries max)

	MOVQ a+8(FP), R14
	MOVQ n+16(FP), R15

	// initialize accumulators Z0, Z1, Z2, Z3, Z4, Z5, Z6, Z7
	VXORPS    Z0, Z0, Z0
	VMOVDQA64 Z0, Z1
	VMOVDQA64 Z0, Z2
	VMOVDQA64 Z0, Z3
	VMOVDQA64 Z0, Z4
	VMOVDQA64 Z0, Z5
	VMOVDQA64 Z0, Z6
	VMOVDQA64 Z0, Z7

	// n % 8 -> CX
	// n / 8 -> R15
	MOVQ R15, CX
	ANDQ $7, CX
	SHRQ $3, R15

loop_single_10:
	TESTQ     CX, CX
	JEQ       loop8by8_8     // n % 8 == 0, we are going to loop over 8 by 8
	VPMOVZXDQ 0(R14), Z8
	VPADDQ    Z8, Z0, Z0
	ADDQ      $32, R14
	DECQ      CX             // decrement nMod8
	JMP       loop_single_10

loop8by8_8:
	TESTQ      R15, R15
	JEQ        accumulate_11  // n == 0, we are going to accumulate
	VPMOVZXDQ  0*32(R14), Z8
	VPMOVZXDQ  1*32(R14), Z9
	VPMOVZXDQ  2*32(R14), Z10
	VPMOVZXDQ  3*32(R14), Z11
	VPMOVZXDQ  4*32(R14), Z12
	VPMOVZXDQ  5*32(R14), Z13
	VPMOVZXDQ  6*32(R14), Z14
	VPMOVZXDQ  7*32(R14), Z15
	PREFETCHT0 256(R14)
	VPADDQ     Z8, Z0, Z0
	VPADDQ     Z9, Z1, Z1
	VPADDQ     Z10, Z2, Z2
	VPADDQ     Z11, Z3, Z3
	VPADDQ     Z12, Z4, Z4
	VPADDQ     Z13, Z5, Z5
	VPADDQ     Z14, Z6, Z6
	VPADDQ     Z15, Z7, Z7

	// increment pointers to visit next 8 elements
	ADDQ $256, R14
	DECQ R15        // decrement n
	JMP  loop8by8_8

accumulate_11:
	// accumulate the 8 Z registers into Z0
	VPADDQ Z7, Z6, Z6
	VPADDQ Z6, Z5, Z5
	VPADDQ Z5, Z4, Z4
	VPADDQ Z4, Z3, Z3
	VPADDQ Z3, Z2, Z2
	VPADDQ Z2, Z1, Z1
	VPADDQ Z1, Z0, Z0

	// carry propagation
	// lo(w0) -> BX
	// hi(w0) -> SI
	// lo(w1) -> DI
	// hi(w1) -> R8
	// lo(w2) -> R9
	// hi(w2) -> R10
	// lo(w3) -> R11
	// hi(w3) -> R12
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

	// lo(hi(wo)) -> R13
	// lo(hi(w1)) -> CX
	// lo(hi(w2)) -> R15
	// lo(hi(w3)) -> R14
#define SPLIT_LO_HI(lo, hi) \
	MOVQ hi, lo;          \
	ANDQ $0xffffffff, lo; \
	SHLQ $32, lo;         \
	SHRQ $32, hi;         \

	SPLIT_LO_HI(R13, SI)
	SPLIT_LO_HI(CX, R8)
	SPLIT_LO_HI(R15, R10)
	SPLIT_LO_HI(R14, R12)

	// r0 = w0l + lo(woh)
	// r1 = carry + hi(woh) + w1l + lo(w1h)
	// r2 = carry + hi(w1h) + w2l + lo(w2h)
	// r3 = carry + hi(w2h) + w3l + lo(w3h)
	// r4 = carry + hi(w3h)

	XORQ  AX, AX   // clear the flags
	ADOXQ R13, BX
	ADOXQ CX, DI
	ADCXQ SI, DI
	ADOXQ R15, R9
	ADCXQ R8, R9
	ADOXQ R14, R11
	ADCXQ R10, R11
	ADOXQ AX, R12
	ADCXQ AX, R12

	// r[0] -> BX
	// r[1] -> DI
	// r[2] -> R9
	// r[3] -> R11
	// r[4] -> R12
	// reduce using single-word Barrett
	// see see Handbook of Applied Cryptography, Algorithm 14.42.
	// mu=2^288 / q -> SI
	MOVQ  $const_mu, SI
	MOVQ  R11, AX
	SHRQ  $32, R12, AX
	MULQ  SI                       // high bits of res stored in DX
	MULXQ ·qElement+0(SB), AX, SI
	SUBQ  AX, BX
	SBBQ  SI, DI
	MULXQ ·qElement+16(SB), AX, SI
	SBBQ  AX, R9
	SBBQ  SI, R11
	SBBQ  $0, R12
	MULXQ ·qElement+8(SB), AX, SI
	SUBQ  AX, DI
	SBBQ  SI, R9
	MULXQ ·qElement+24(SB), AX, SI
	SBBQ  AX, R11
	SBBQ  SI, R12
	MOVQ  BX, R8
	MOVQ  DI, R10
	MOVQ  R9, R13
	MOVQ  R11, CX
	SUBQ  ·qElement+0(SB), BX
	SBBQ  ·qElement+8(SB), DI
	SBBQ  ·qElement+16(SB), R9
	SBBQ  ·qElement+24(SB), R11
	SBBQ  $0, R12
	JCS   modReduced_12
	MOVQ  BX, R8
	MOVQ  DI, R10
	MOVQ  R9, R13
	MOVQ  R11, CX
	SUBQ  ·qElement+0(SB), BX
	SBBQ  ·qElement+8(SB), DI
	SBBQ  ·qElement+16(SB), R9
	SBBQ  ·qElement+24(SB), R11
	SBBQ  $0, R12
	JCS   modReduced_12
	MOVQ  BX, R8
	MOVQ  DI, R10
	MOVQ  R9, R13
	MOVQ  R11, CX

modReduced_12:
	MOVQ res+0(FP), SI
	MOVQ R8, 0(SI)
	MOVQ R10, 8(SI)
	MOVQ R13, 16(SI)
	MOVQ CX, 24(SI)

done_9:
	RET

// innerProdVec(res, a,b *Element, n uint64) res = sum(a[0...n] * b[0...n])
TEXT ·innerProdVec(SB), NOSPLIT, $0-32
	MOVQ      a+8(FP), R14
	MOVQ      b+16(FP), R15
	MOVQ      n+24(FP), CX
	VPCMPEQB  Y0, Y0, Y0
	VPMOVZXDQ Y0, Z5
	VPXORQ    Z16, Z16, Z16
	VMOVDQA64 Z16, Z17
	VMOVDQA64 Z16, Z18
	VMOVDQA64 Z16, Z19
	VMOVDQA64 Z16, Z20
	VMOVDQA64 Z16, Z21
	VMOVDQA64 Z16, Z22
	VMOVDQA64 Z16, Z23
	VMOVDQA64 Z16, Z24
	VMOVDQA64 Z16, Z25
	VMOVDQA64 Z16, Z26
	VMOVDQA64 Z16, Z27
	VMOVDQA64 Z16, Z28
	VMOVDQA64 Z16, Z29
	VMOVDQA64 Z16, Z30
	VMOVDQA64 Z16, Z31
	TESTQ     CX, CX
	JEQ       done_14       // n == 0, we are done

loop_13:
	TESTQ     CX, CX
	JEQ       accumulate_15 // n == 0 we can accumulate
	VPMOVZXDQ (R15), Z4
	ADDQ      $32, R15

	// we multiply and accumulate partial products of 4 bytes * 32 bytes
	VPMULUDQ.BCST 0*4(R14), Z4, Z2
	VPSRLQ        $32, Z2, Z3
	VPANDQ        Z5, Z2, Z2
	VPADDQ        Z2, Z16, Z16
	VPADDQ        Z3, Z24, Z24
	VPMULUDQ.BCST 1*4(R14), Z4, Z2
	VPSRLQ        $32, Z2, Z3
	VPANDQ        Z5, Z2, Z2
	VPADDQ        Z2, Z17, Z17
	VPADDQ        Z3, Z25, Z25
	VPMULUDQ.BCST 2*4(R14), Z4, Z2
	VPSRLQ        $32, Z2, Z3
	VPANDQ        Z5, Z2, Z2
	VPADDQ        Z2, Z18, Z18
	VPADDQ        Z3, Z26, Z26
	VPMULUDQ.BCST 3*4(R14), Z4, Z2
	VPSRLQ        $32, Z2, Z3
	VPANDQ        Z5, Z2, Z2
	VPADDQ        Z2, Z19, Z19
	VPADDQ        Z3, Z27, Z27
	VPMULUDQ.BCST 4*4(R14), Z4, Z2
	VPSRLQ        $32, Z2, Z3
	VPANDQ        Z5, Z2, Z2
	VPADDQ        Z2, Z20, Z20
	VPADDQ        Z3, Z28, Z28
	VPMULUDQ.BCST 5*4(R14), Z4, Z2
	VPSRLQ        $32, Z2, Z3
	VPANDQ        Z5, Z2, Z2
	VPADDQ        Z2, Z21, Z21
	VPADDQ        Z3, Z29, Z29
	VPMULUDQ.BCST 6*4(R14), Z4, Z2
	VPSRLQ        $32, Z2, Z3
	VPANDQ        Z5, Z2, Z2
	VPADDQ        Z2, Z22, Z22
	VPADDQ        Z3, Z30, Z30
	VPMULUDQ.BCST 7*4(R14), Z4, Z2
	VPSRLQ        $32, Z2, Z3
	VPANDQ        Z5, Z2, Z2
	VPADDQ        Z2, Z23, Z23
	VPADDQ        Z3, Z31, Z31
	ADDQ          $32, R14
	DECQ          CX               // decrement n
	JMP           loop_13

accumulate_15:
	MOVQ  $0x0000000000001555, AX
	KMOVD AX, K1
	MOVQ  $1, AX
	KMOVD AX, K2

	// store the least significant 32 bits of ACC (starts with A0L) in Z0
	VALIGND.Z   $16, Z16, Z16, K2, Z0
	KSHIFTLW    $1, K2, K2
	VPSRLQ      $32, Z16, Z2
	VALIGND.Z   $2, Z16, Z16, K1, Z16
	VPADDQ      Z2, Z16, Z16
	VPANDQ      Z5, Z24, Z2
	VPADDQ      Z2, Z16, Z16
	VPANDQ      Z5, Z17, Z2
	VPADDQ      Z2, Z16, Z16
	VALIGND     $15, Z16, Z16, K2, Z0
	KSHIFTLW    $1, K2, K2
	VPSRLQ      $32, Z16, Z2
	VALIGND.Z   $2, Z16, Z16, K1, Z16
	VPADDQ      Z2, Z16, Z16
	VPSRLQ      $32, Z24, Z24
	VPADDQ      Z24, Z16, Z16
	VPSRLQ      $32, Z17, Z17
	VPADDQ      Z17, Z16, Z16
	VPANDQ      Z5, Z25, Z2
	VPADDQ      Z2, Z16, Z16
	VPANDQ      Z5, Z18, Z2
	VPADDQ      Z2, Z16, Z16
	VALIGND     $16-2, Z16, Z16, K2, Z0
	KADDW       K2, K2, K2
	VPSRLQ      $32, Z16, Z2
	VALIGND.Z   $2, Z16, Z16, K1, Z16
	VPADDQ      Z2, Z16, Z16
	VPSRLQ      $32, Z25, Z25
	VPADDQ      Z25, Z16, Z16
	VPSRLQ      $32, Z18, Z18
	VPADDQ      Z18, Z16, Z16
	VPANDQ      Z5, Z26, Z2
	VPADDQ      Z2, Z16, Z16
	VPANDQ      Z5, Z19, Z2
	VPADDQ      Z2, Z16, Z16
	VALIGND     $16-3, Z16, Z16, K2, Z0
	KADDW       K2, K2, K2
	VPSRLQ      $32, Z16, Z2
	VALIGND.Z   $2, Z16, Z16, K1, Z16
	VPADDQ      Z2, Z16, Z16
	VPSRLQ      $32, Z26, Z26
	VPADDQ      Z26, Z16, Z16
	VPSRLQ      $32, Z19, Z19
	VPADDQ      Z19, Z16, Z16
	VPANDQ      Z5, Z27, Z2
	VPADDQ      Z2, Z16, Z16
	VPANDQ      Z5, Z20, Z2
	VPADDQ      Z2, Z16, Z16
	VALIGND     $16-4, Z16, Z16, K2, Z0
	KADDW       K2, K2, K2
	VPSRLQ      $32, Z16, Z2
	VALIGND.Z   $2, Z16, Z16, K1, Z16
	VPADDQ      Z2, Z16, Z16
	VPSRLQ      $32, Z27, Z27
	VPADDQ      Z27, Z16, Z16
	VPSRLQ      $32, Z20, Z20
	VPADDQ      Z20, Z16, Z16
	VPANDQ      Z5, Z28, Z2
	VPADDQ      Z2, Z16, Z16
	VPANDQ      Z5, Z21, Z2
	VPADDQ      Z2, Z16, Z16
	VALIGND     $16-5, Z16, Z16, K2, Z0
	KADDW       K2, K2, K2
	VPSRLQ      $32, Z16, Z2
	VALIGND.Z   $2, Z16, Z16, K1, Z16
	VPADDQ      Z2, Z16, Z16
	VPSRLQ      $32, Z28, Z28
	VPADDQ      Z28, Z16, Z16
	VPSRLQ      $32, Z21, Z21
	VPADDQ      Z21, Z16, Z16
	VPANDQ      Z5, Z29, Z2
	VPADDQ      Z2, Z16, Z16
	VPANDQ      Z5, Z22, Z2
	VPADDQ      Z2, Z16, Z16
	VALIGND     $16-6, Z16, Z16, K2, Z0
	KADDW       K2, K2, K2
	VPSRLQ      $32, Z16, Z2
	VALIGND.Z   $2, Z16, Z16, K1, Z16
	VPADDQ      Z2, Z16, Z16
	VPSRLQ      $32, Z29, Z29
	VPADDQ      Z29, Z16, Z16
	VPSRLQ      $32, Z22, Z22
	VPADDQ      Z22, Z16, Z16
	VPANDQ      Z5, Z30, Z2
	VPADDQ      Z2, Z16, Z16
	VPANDQ      Z5, Z23, Z2
	VPADDQ      Z2, Z16, Z16
	VALIGND     $16-7, Z16, Z16, K2, Z0
	KADDW       K2, K2, K2
	VPSRLQ      $32, Z16, Z2
	VALIGND.Z   $2, Z16, Z16, K1, Z16
	VPADDQ      Z2, Z16, Z16
	VPSRLQ      $32, Z30, Z30
	VPADDQ      Z30, Z16, Z16
	VPSRLQ      $32, Z23, Z23
	VPADDQ      Z23, Z16, Z16
	VPANDQ      Z5, Z31, Z2
	VPADDQ      Z2, Z16, Z16
	VALIGND     $16-8, Z16, Z16, K2, Z0
	KSHIFTLW    $1, K2, K2
	VPSRLQ      $32, Z16, Z2
	VALIGND.Z   $2, Z16, Z16, K1, Z16
	VPADDQ      Z2, Z16, Z16
	VPSRLQ      $32, Z31, Z31
	VPADDQ      Z31, Z16, Z16
	VALIGND     $16-9, Z16, Z16, K2, Z0
	KSHIFTLW    $1, K2, K2
	VPSRLQ      $32, Z16, Z2
	VALIGND.Z   $2, Z16, Z16, K1, Z16
	VPADDQ      Z2, Z16, Z16
	VALIGND     $16-10, Z16, Z16, K2, Z0
	KSHIFTLW    $1, K2, K2
	VPSRLQ      $32, Z16, Z2
	VALIGND.Z   $2, Z16, Z16, K1, Z16
	VPADDQ      Z2, Z16, Z16
	VALIGND     $16-11, Z16, Z16, K2, Z0
	KSHIFTLW    $1, K2, K2
	VPSRLQ      $32, Z16, Z2
	VALIGND.Z   $2, Z16, Z16, K1, Z16
	VPADDQ      Z2, Z16, Z16
	VALIGND     $16-12, Z16, Z16, K2, Z0
	KSHIFTLW    $1, K2, K2
	VPSRLQ      $32, Z16, Z2
	VALIGND.Z   $2, Z16, Z16, K1, Z16
	VPADDQ      Z2, Z16, Z16
	VALIGND     $16-13, Z16, Z16, K2, Z0
	KSHIFTLW    $1, K2, K2
	VPSRLQ      $32, Z16, Z2
	VALIGND.Z   $2, Z16, Z16, K1, Z16
	VPADDQ      Z2, Z16, Z16
	VALIGND     $16-14, Z16, Z16, K2, Z0
	KSHIFTLW    $1, K2, K2
	VPSRLQ      $32, Z16, Z2
	VALIGND.Z   $2, Z16, Z16, K1, Z16
	VPADDQ      Z2, Z16, Z16
	VALIGND     $16-15, Z16, Z16, K2, Z0
	KSHIFTLW    $1, K2, K2
	VPSRLQ      $32, Z16, Z2
	VALIGND.Z   $2, Z16, Z16, K1, Z16
	VPADDQ      Z2, Z16, Z16
	VMOVDQA64.Z Z16, K1, Z1
	VMOVQ       X0, SI
	VALIGNQ     $1, Z0, Z1, Z0
	VMOVQ       X0, DI
	VALIGNQ     $1, Z0, Z0, Z0
	VMOVQ       X0, R8
	VALIGNQ     $1, Z0, Z0, Z0
	VMOVQ       X0, R9
	VALIGNQ     $1, Z0, Z0, Z0
	XORQ        BX, BX
	MOVQ        $const_qInvNeg, DX
	MULXQ       SI, DX, R10
	MULXQ       ·qElement+0(SB), AX, R10
	ADDQ        AX, SI
	ADCQ        R10, DI
	MULXQ       ·qElement+16(SB), AX, R10
	ADCQ        AX, R8
	ADCQ        R10, R9
	ADCQ        $0, BX
	MULXQ       ·qElement+8(SB), AX, R10
	ADDQ        AX, DI
	ADCQ        R10, R8
	MULXQ       ·qElement+24(SB), AX, R10
	ADCQ        AX, R9
	ADCQ        R10, BX
	ADCQ        $0, SI
	MOVQ        $const_qInvNeg, DX
	MULXQ       DI, DX, R10
	MULXQ       ·qElement+0(SB), AX, R10
	ADDQ        AX, DI
	ADCQ        R10, R8
	MULXQ       ·qElement+16(SB), AX, R10
	ADCQ        AX, R9
	ADCQ        R10, BX
	ADCQ        $0, SI
	MULXQ       ·qElement+8(SB), AX, R10
	ADDQ        AX, R8
	ADCQ        R10, R9
	MULXQ       ·qElement+24(SB), AX, R10
	ADCQ        AX, BX
	ADCQ        R10, SI
	ADCQ        $0, DI
	MOVQ        $const_qInvNeg, DX
	MULXQ       R8, DX, R10
	MULXQ       ·qElement+0(SB), AX, R10
	ADDQ        AX, R8
	ADCQ        R10, R9
	MULXQ       ·qElement+16(SB), AX, R10
	ADCQ        AX, BX
	ADCQ        R10, SI
	ADCQ        $0, DI
	MULXQ       ·qElement+8(SB), AX, R10
	ADDQ        AX, R9
	ADCQ        R10, BX
	MULXQ       ·qElement+24(SB), AX, R10
	ADCQ        AX, SI
	ADCQ        R10, DI
	ADCQ        $0, R8
	MOVQ        $const_qInvNeg, DX
	MULXQ       R9, DX, R10
	MULXQ       ·qElement+0(SB), AX, R10
	ADDQ        AX, R9
	ADCQ        R10, BX
	MULXQ       ·qElement+16(SB), AX, R10
	ADCQ        AX, SI
	ADCQ        R10, DI
	ADCQ        $0, R8
	MULXQ       ·qElement+8(SB), AX, R10
	ADDQ        AX, BX
	ADCQ        R10, SI
	MULXQ       ·qElement+24(SB), AX, R10
	ADCQ        AX, DI
	ADCQ        R10, R8
	ADCQ        $0, R9
	VMOVQ       X0, AX
	ADDQ        AX, BX
	VALIGNQ     $1, Z0, Z0, Z0
	VMOVQ       X0, AX
	ADCQ        AX, SI
	VALIGNQ     $1, Z0, Z0, Z0
	VMOVQ       X0, AX
	ADCQ        AX, DI
	VALIGNQ     $1, Z0, Z0, Z0
	VMOVQ       X0, AX
	ADCQ        AX, R8
	VALIGNQ     $1, Z0, Z0, Z0
	VMOVQ       X0, AX
	ADCQ        AX, R9
	MOVQ        R8, AX
	SHRQ        $32, R9, AX
	MOVQ        $const_mu, DX
	MULQ        DX
	MULXQ       ·qElement+0(SB), AX, R10
	SUBQ        AX, BX
	SBBQ        R10, SI
	MULXQ       ·qElement+16(SB), AX, R10
	SBBQ        AX, DI
	SBBQ        R10, R8
	SBBQ        $0, R9
	MULXQ       ·qElement+8(SB), AX, R10
	SUBQ        AX, SI
	SBBQ        R10, DI
	MULXQ       ·qElement+24(SB), AX, R10
	SBBQ        AX, R8
	SBBQ        R10, R9
	MOVQ        res+0(FP), R11
	MOVQ        BX, 0(R11)
	MOVQ        SI, 8(R11)
	MOVQ        DI, 16(R11)
	MOVQ        R8, 24(R11)
	SUBQ        ·qElement+0(SB), BX
	SBBQ        ·qElement+8(SB), SI
	SBBQ        ·qElement+16(SB), DI
	SBBQ        ·qElement+24(SB), R8
	SBBQ        $0, R9
	JCS         done_14
	MOVQ        BX, 0(R11)
	MOVQ        SI, 8(R11)
	MOVQ        DI, 16(R11)
	MOVQ        R8, 24(R11)
	SUBQ        ·qElement+0(SB), BX
	SBBQ        ·qElement+8(SB), SI
	SBBQ        ·qElement+16(SB), DI
	SBBQ        ·qElement+24(SB), R8
	SBBQ        $0, R9
	JCS         done_14
	MOVQ        BX, 0(R11)
	MOVQ        SI, 8(R11)
	MOVQ        DI, 16(R11)
	MOVQ        R8, 24(R11)

done_14:
	RET
