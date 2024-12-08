// Code generated by gnark-crypto/generator. DO NOT EDIT.
#include "textflag.h"
#include "funcdata.h"
#include "go_asm.h"

// addVec(res, a, b *Element, n uint64) res[0...n] = a[0...n] + b[0...n]
TEXT ·addVec(SB), NOSPLIT, $0-32
	MOVD         $const_q, AX
	VPBROADCASTD AX, Z3
	MOVQ         res+0(FP), CX
	MOVQ         a+8(FP), AX
	MOVQ         b+16(FP), DX
	MOVQ         n+24(FP), BX

loop_1:
	TESTQ     BX, BX
	JEQ       done_2     // n == 0, we are done
	VMOVDQU32 0(AX), Z0
	VMOVDQU32 0(DX), Z1
	VPADDD    Z0, Z1, Z0 // a = a + b
	VPSUBD    Z3, Z0, Z2 // t = a - q
	VPMINUD   Z0, Z2, Z1 // b = min(t, a)
	VMOVDQU32 Z1, 0(CX)  // res = b

	// increment pointers to visit next element
	ADDQ $64, AX
	ADDQ $64, DX
	ADDQ $64, CX
	DECQ BX      // decrement n
	JMP  loop_1

done_2:
	RET

// subVec(res, a, b *Element, n uint64) res[0...n] = a[0...n] - b[0...n]
TEXT ·subVec(SB), NOSPLIT, $0-32
	MOVD         $const_q, AX
	VPBROADCASTD AX, Z3
	MOVQ         res+0(FP), CX
	MOVQ         a+8(FP), AX
	MOVQ         b+16(FP), DX
	MOVQ         n+24(FP), BX

loop_3:
	TESTQ     BX, BX
	JEQ       done_4     // n == 0, we are done
	VMOVDQU32 0(AX), Z0
	VMOVDQU32 0(DX), Z1
	VPSUBD    Z1, Z0, Z0 // a = a - b
	VPADDD    Z3, Z0, Z2 // t = a + q
	VPMINUD   Z0, Z2, Z1 // b = min(t, a)
	VMOVDQU32 Z1, 0(CX)  // res = b

	// increment pointers to visit next element
	ADDQ $64, AX
	ADDQ $64, DX
	ADDQ $64, CX
	DECQ BX      // decrement n
	JMP  loop_3

done_4:
	RET

// sumVec(res *uint64, a *[]uint32, n uint64) res = sum(a[0...n])

// We are load 8 31bits values at a time and accumulate them into an accumulator of
// 8 quadwords (64bits). The caller then needs to reduce the result mod q.
// We can safely accumulate ~2**33 31bits values into a single accumulator.
// That gives us a maximum of 2**33 * 8 = 2**36 31bits values to sum safely.

TEXT ·sumVec(SB), NOSPLIT, $0-24
	MOVQ      t+0(FP), R15
	MOVQ      a+8(FP), R14
	MOVQ      n+16(FP), CX
	VXORPS    Z2, Z2, Z2   // acc1 = 0
	VMOVDQA64 Z2, Z3       // acc2 = 0

loop_5:
	TESTQ     CX, CX
	JEQ       done_6      // n == 0, we are done
	VPMOVZXDQ 0(R14), Z0  // load 8 31bits values in a1
	VPMOVZXDQ 32(R14), Z1 // load 8 31bits values in a2
	VPADDQ    Z0, Z2, Z2  // acc1 += a1
	VPADDQ    Z1, Z3, Z3  // acc2 += a2

	// increment pointers to visit next element
	ADDQ $64, R14
	DECQ CX       // decrement n
	JMP  loop_5

done_6:
	VPADDQ    Z2, Z3, Z2 // acc1 += acc2
	VMOVDQU64 Z2, 0(R15) // res = acc1
	RET

// mulVec(res, a, b *Element, n uint64) res[0...n] = a[0...n] * b[0...n]
TEXT ·mulVec(SB), NOSPLIT, $0-32
	MOVD         $const_q, AX
	VPBROADCASTQ AX, Z3
	MOVD         $const_qInvNeg, AX
	VPBROADCASTQ AX, Z4

	// Create mask for low dword in each qword
	VPCMPEQB  Y0, Y0, Y0
	VPMOVZXDQ Y0, Z6
	MOVQ      res+0(FP), CX
	MOVQ      a+8(FP), AX
	MOVQ      b+16(FP), DX
	MOVQ      n+24(FP), BX

loop_7:
	TESTQ     BX, BX
	JEQ       done_8      // n == 0, we are done
	VPMOVZXDQ 0(AX), Z0
	VPMOVZXDQ 0(DX), Z1
	VPMULUDQ  Z0, Z1, Z2  // P = a * b
	VPANDQ    Z6, Z2, Z5  // m = uint32(P)
	VPMULUDQ  Z5, Z4, Z5  // m = m * qInvNeg
	VPANDQ    Z6, Z5, Z5  // m = uint32(m)
	VPMULUDQ  Z5, Z3, Z5  // m = m * q
	VPADDQ    Z2, Z5, Z2  // P = P + m
	VPSRLQ    $32, Z2, Z2 // P = P >> 32
	VPSUBD    Z3, Z2, Z5  // PL = P - q
	VPMINUD   Z2, Z5, Z2  // P = min(P, PL)
	VPMOVQD   Z2, 0(CX)   // res = P

	// increment pointers to visit next element
	ADDQ $32, AX
	ADDQ $32, DX
	ADDQ $32, CX
	DECQ BX      // decrement n
	JMP  loop_7

done_8:
	RET

// scalarMulVec(res, a, b *Element, n uint64) res[0...n] = a[0...n] * b
TEXT ·scalarMulVec(SB), NOSPLIT, $0-32
	MOVD         $const_q, AX
	VPBROADCASTQ AX, Z3
	MOVD         $const_qInvNeg, AX
	VPBROADCASTQ AX, Z4

	// Create mask for low dword in each qword
	VPCMPEQB     Y0, Y0, Y0
	VPMOVZXDQ    Y0, Z6
	MOVQ         res+0(FP), CX
	MOVQ         a+8(FP), AX
	MOVQ         b+16(FP), DX
	MOVQ         n+24(FP), BX
	VPBROADCASTD 0(DX), Z1

loop_9:
	TESTQ     BX, BX
	JEQ       done_10     // n == 0, we are done
	VPMOVZXDQ 0(AX), Z0
	VPMULUDQ  Z0, Z1, Z2  // P = a * b
	VPANDQ    Z6, Z2, Z5  // m = uint32(P)
	VPMULUDQ  Z5, Z4, Z5  // m = m * qInvNeg
	VPANDQ    Z6, Z5, Z5  // m = uint32(m)
	VPMULUDQ  Z5, Z3, Z5  // m = m * q
	VPADDQ    Z2, Z5, Z2  // P = P + m
	VPSRLQ    $32, Z2, Z2 // P = P >> 32
	VPSUBD    Z3, Z2, Z5  // PL = P - q
	VPMINUD   Z2, Z5, Z2  // P = min(P, PL)
	VPMOVQD   Z2, 0(CX)   // res = P

	// increment pointers to visit next element
	ADDQ $32, AX
	ADDQ $32, CX
	DECQ BX      // decrement n
	JMP  loop_9

done_10:
	RET

// innerProdVec(t *uint64, a,b *[]uint32, n uint64) res = sum(a[0...n] * b[0...n])
TEXT ·innerProdVec(SB), NOSPLIT, $0-32

	// Similar to mulVec; we do most of the montgomery multiplication but don't do
	// the final reduction. We accumulate the result like in sumVec and let the caller
	// reduce mod q.

	MOVD         $const_q, AX
	VPBROADCASTQ AX, Z3
	MOVD         $const_qInvNeg, AX
	VPBROADCASTQ AX, Z4

	// Create mask for low dword in each qword
	VPCMPEQB  Y0, Y0, Y0
	VPMOVZXDQ Y0, Z6
	VXORPS    Z2, Z2, Z2    // acc = 0
	MOVQ      t+0(FP), CX
	MOVQ      a+8(FP), R14
	MOVQ      b+16(FP), R15
	MOVQ      n+24(FP), BX

loop_11:
	TESTQ     BX, BX
	JEQ       done_12     // n == 0, we are done
	VPMOVZXDQ 0(R14), Z0
	VPMOVZXDQ 0(R15), Z1
	VPMULUDQ  Z0, Z1, Z7  // P = a * b
	VPANDQ    Z6, Z7, Z5  // m = uint32(P)
	VPMULUDQ  Z5, Z4, Z5  // m = m * qInvNeg
	VPANDQ    Z6, Z5, Z5  // m = uint32(m)
	VPMULUDQ  Z5, Z3, Z5  // m = m * q
	VPADDQ    Z7, Z5, Z7  // P = P + m
	VPSRLQ    $32, Z7, Z7 // P = P >> 32

	// accumulate P into acc, P is in [0, 2q] on 32bits max
	VPADDQ Z7, Z2, Z2 // acc += P

	// increment pointers to visit next element
	ADDQ $32, R14
	ADDQ $32, R15
	DECQ BX       // decrement n
	JMP  loop_11

done_12:
	VMOVDQU64 Z2, 0(CX) // res = acc
	RET
