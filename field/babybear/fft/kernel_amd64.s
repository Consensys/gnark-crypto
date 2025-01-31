//go:build !purego

// Code generated by gnark-crypto/generator. DO NOT EDIT.
// Refer to the generator for more documentation.
// Some sub-functions are derived from Plonky3:
// https://github.com/Plonky3/Plonky3/blob/36e619f3c6526ee86e2e5639a24b3224e1c1700f/monty-31/src/x86_64_avx512/packing.rs#L319

#include "textflag.h"
#include "funcdata.h"
#include "go_asm.h"

#define BUTTERFLYD1Q(in0, in1, in2, in3, in4) \
	VPADDD  in0, in1, in3 \
	VPSUBD  in1, in0, in1 \
	VPSUBD  in2, in3, in0 \
	VPMINUD in3, in0, in0 \
	VPADDD  in2, in1, in4 \
	VPMINUD in4, in1, in1 \

#define BUTTERFLYD2Q(in0, in1, in2, in3, in4) \
	VPSUBD  in1, in0, in4 \
	VPADDD  in0, in1, in3 \
	VPADDD  in2, in4, in1 \
	VPSUBD  in2, in3, in0 \
	VPMINUD in3, in0, in0 \

#define BUTTERFLYD2Q2Q(in0, in1, in2, in3) \
	VPSUBD in1, in0, in3 \
	VPADDD in0, in1, in0 \
	VPADDD in2, in3, in1 \

#define MULD(in0, in1, in2, in3, in4, in5, in6, in7, in8, in9) \
	VPSRLQ    $32, in0, in2 \
	VPSRLQ    $32, in1, in3 \
	VPMULUDQ  in0, in1, in4 \
	VPMULUDQ  in2, in3, in5 \
	VPMULUDQ  in4, in9, in6 \
	VPMULUDQ  in5, in9, in7 \
	VPMULUDQ  in6, in8, in6 \
	VPADDQ    in4, in6, in4 \
	VPMULUDQ  in7, in8, in7 \
	VPADDQ    in5, in7, in5 \
	VMOVSHDUP in4, K3, in5  \
	VPSUBD    in8, in5, in7 \
	VPMINUD   in5, in7, in0 \

#define PERMUTE8X8(in0, in1, in2) \
	VSHUFI64X2 $0x000000000000004e, in1, in0, in2 \
	VPBLENDMQ  in0, in2, K1, in0                  \
	VPBLENDMQ  in2, in1, K1, in1                  \

#define PERMUTE4X4(in0, in1, in2, in3) \
	VMOVDQA64 in2, in3          \
	VPERMI2Q  in1, in0, in3     \
	VPBLENDMQ in0, in3, K2, in0 \
	VPBLENDMQ in3, in1, K2, in1 \

#define PERMUTE2X2(in0, in1, in2) \
	VSHUFPD   $0x0000000000000055, in1, in0, in2 \
	VPBLENDMQ in0, in2, K3, in0                  \
	VPBLENDMQ in2, in1, K3, in1                  \

#define PERMUTE1X1(in0, in1, in2) \
	VPSHRDQ   $32, in1, in0, in2 \
	VPBLENDMD in0, in2, K3, in0  \
	VPBLENDMD in2, in1, K3, in1  \

#define LOAD_Q(in0, in1) \
	MOVD         $const_q, AX       \
	VPBROADCASTD AX, in0            \
	MOVD         $const_qInvNeg, AX \
	VPBROADCASTD AX, in1            \

#define LOAD_MASKS() \
	MOVQ  $0x0000000000000f0f, AX \
	KMOVQ AX, K1                  \
	MOVQ  $0x0000000000000033, AX \
	KMOVQ AX, K2                  \
	MOVQ  $0x0000000000005555, AX \
	KMOVD AX, K3                  \

#define BUTTERFLY_MULD(in0, in1, in2, in3, in4, in5, in6, in7, in8, in9, in10, in11, in12, in13, in14) \
BUTTERFLYD2Q(in0, in1, in2, in3, in4)                       \
MULD(in5, in6, in7, in8, in9, in10, in11, in12, in13, in14) \

TEXT ·innerDITWithTwiddles_avx512(SB), NOSPLIT, $0-72
	LOAD_Q(Z4, Z5)
	LOAD_MASKS()

	// load arguments
	MOVQ a+0(FP), R15
	MOVQ twiddles+24(FP), CX
	MOVQ end+56(FP), SI
	MOVQ m+64(FP), BX
	CMPQ BX, $0x0000000000000010
	JL   smallerThan16_1         // m < 16
	SHRQ $4, SI                  // we are processing 16 elements at a time
	SHLQ $2, BX                  // offset = m * 4bytes
	MOVQ R15, DX
	ADDQ BX, DX

loop_3:
	TESTQ     SI, SI
	JEQ       done_2     // n == 0, we are done
	VMOVDQU32 0(R15), Z0 // load a[i]
	VMOVDQU32 0(DX), Z1  // load a[i+m]
	VMOVDQU32 0(CX), Z6
	MULD(Z1, Z6, Z7, Z8, Z2, Z3, Z9, Z10, Z4, Z5)
	BUTTERFLYD1Q(Z0, Z1, Z4, Z2, Z3)
	VMOVDQU32 Z0, 0(R15) // store a[i]
	VMOVDQU32 Z1, 0(DX)  // store a[i+m]
	ADDQ      $64, R15
	ADDQ      $64, DX
	ADDQ      $64, CX
	DECQ      SI         // decrement n
	JMP       loop_3

done_2:
	RET

smallerThan16_1:
	// m < 16, we call the generic one
	// note that this should happen only when doing a FFT smaller than the smallest generated kernel
	MOVQ a+0(FP), AX
	MOVQ AX, (SP)
	MOVQ twiddles+24(FP), AX
	MOVQ AX, 24(SP)
	MOVQ start+48(FP), AX
	MOVQ AX, 48(SP)
	MOVQ end+56(FP), AX
	MOVQ AX, 56(SP)
	MOVQ m+64(FP), AX
	MOVQ AX, 64(SP)
	CALL ·innerDITWithTwiddlesGeneric(SB)
	RET

TEXT ·innerDIFWithTwiddles_avx512(SB), NOSPLIT, $0-72
	LOAD_Q(Z2, Z4)
	LOAD_MASKS()

	// load arguments
	MOVQ a+0(FP), R15
	MOVQ twiddles+24(FP), CX
	MOVQ end+56(FP), SI
	MOVQ m+64(FP), BX
	CMPQ BX, $0x0000000000000010
	JL   smallerThan16_4         // m < 16
	SHRQ $4, SI                  // we are processing 16 elements at a time
	SHLQ $2, BX                  // offset = m * 4bytes
	MOVQ R15, DX
	ADDQ BX, DX

loop_6:
	TESTQ     SI, SI
	JEQ       done_5     // n == 0, we are done
	VMOVDQU32 0(R15), Z0 // load a[i]
	VMOVDQU32 0(DX), Z1  // load a[i+m]
	VMOVDQU32 0(CX), Z5
	BUTTERFLY_MULD(Z0, Z1, Z2, Z3, Z8, Z1, Z5, Z6, Z7, Z3, Z8, Z9, Z10, Z2, Z4)
	VMOVDQU32 Z0, 0(R15) // store a[i]
	VMOVDQU32 Z1, 0(DX)
	ADDQ      $64, R15
	ADDQ      $64, DX
	ADDQ      $64, CX
	DECQ      SI         // decrement n
	JMP       loop_6

done_5:
	RET

smallerThan16_4:
	// m < 16, we call the generic one
	// note that this should happen only when doing a FFT smaller than the smallest generated kernel
	MOVQ a+0(FP), AX
	MOVQ AX, (SP)
	MOVQ twiddles+24(FP), AX
	MOVQ AX, 24(SP)
	MOVQ start+48(FP), AX
	MOVQ AX, 48(SP)
	MOVQ end+56(FP), AX
	MOVQ AX, 56(SP)
	MOVQ m+64(FP), AX
	MOVQ AX, 64(SP)
	CALL ·innerDIFWithTwiddlesGeneric(SB)
	RET

TEXT ·kerDIFNP_256_avx512(SB), NOSPLIT, $0-56
	LOAD_Q(Z16, Z17)
	LOAD_MASKS()

	// load arguments
	MOVQ         a+0(FP), R15
	MOVQ         twiddles+24(FP), CX
	MOVQ         stage+48(FP), AX
	IMULQ        $24, AX
	ADDQ         AX, CX                             // we want twiddles[stage] as starting point
	VMOVDQU32    0(R15), Z0                         // load a[0]
	VMOVDQU32    64(R15), Z1                        // load a[1]
	VMOVDQU32    128(R15), Z2                       // load a[2]
	VMOVDQU32    192(R15), Z3                       // load a[3]
	VMOVDQU32    256(R15), Z4                       // load a[4]
	VMOVDQU32    320(R15), Z5                       // load a[5]
	VMOVDQU32    384(R15), Z6                       // load a[6]
	VMOVDQU32    448(R15), Z7                       // load a[7]
	VMOVDQU32    512(R15), Z8                       // load a[8]
	VMOVDQU32    576(R15), Z9                       // load a[9]
	VMOVDQU32    640(R15), Z10                      // load a[10]
	VMOVDQU32    704(R15), Z11                      // load a[11]
	VMOVDQU32    768(R15), Z12                      // load a[12]
	VMOVDQU32    832(R15), Z13                      // load a[13]
	VMOVDQU32    896(R15), Z14                      // load a[14]
	VMOVDQU32    960(R15), Z15                      // load a[15]
	MOVQ         0(CX), DI
	VMOVDQU32    0(DI), Z18
	VMOVDQU32    64(DI), Z19
	VMOVDQU32    128(DI), Z20
	VMOVDQU32    192(DI), Z21
	VMOVDQU32    256(DI), Z22
	VMOVDQU32    320(DI), Z23
	VMOVDQU32    384(DI), Z24
	VMOVDQU32    448(DI), Z25
	BUTTERFLYD2Q(Z0, Z8, Z16, Z31, Z27)
	BUTTERFLYD2Q(Z1, Z9, Z16, Z31, Z27)
	BUTTERFLYD2Q(Z2, Z10, Z16, Z31, Z27)
	BUTTERFLYD2Q(Z3, Z11, Z16, Z31, Z27)
	BUTTERFLYD2Q(Z4, Z12, Z16, Z31, Z27)
	BUTTERFLYD2Q(Z5, Z13, Z16, Z31, Z27)
	BUTTERFLYD2Q(Z6, Z14, Z16, Z31, Z27)
	BUTTERFLYD2Q(Z7, Z15, Z16, Z31, Z27)
	MULD(Z8, Z18, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z9, Z19, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z10, Z20, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z11, Z21, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z12, Z22, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z13, Z23, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z14, Z24, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z15, Z25, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	ADDQ         $24, CX
	MOVQ         0(CX), DI
	VMOVDQU32    0(DI), Z18
	VMOVDQU32    64(DI), Z19
	VMOVDQU32    128(DI), Z20
	VMOVDQU32    192(DI), Z21
	BUTTERFLYD2Q(Z0, Z4, Z16, Z31, Z27)
	BUTTERFLYD2Q(Z1, Z5, Z16, Z31, Z27)
	BUTTERFLYD2Q(Z2, Z6, Z16, Z31, Z27)
	BUTTERFLYD2Q(Z3, Z7, Z16, Z31, Z27)
	BUTTERFLYD2Q(Z8, Z12, Z16, Z31, Z27)
	BUTTERFLYD2Q(Z9, Z13, Z16, Z31, Z27)
	BUTTERFLYD2Q(Z10, Z14, Z16, Z31, Z27)
	BUTTERFLYD2Q(Z11, Z15, Z16, Z31, Z27)
	MULD(Z4, Z18, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z5, Z19, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z6, Z20, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z7, Z21, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z12, Z18, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z13, Z19, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z14, Z20, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z15, Z21, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	ADDQ         $24, CX
	MOVQ         0(CX), DI
	VMOVDQU32    0(DI), Z18
	VMOVDQU32    64(DI), Z19
	BUTTERFLYD2Q(Z0, Z2, Z16, Z31, Z27)
	BUTTERFLYD2Q(Z1, Z3, Z16, Z31, Z27)
	BUTTERFLYD2Q(Z4, Z6, Z16, Z31, Z27)
	BUTTERFLYD2Q(Z5, Z7, Z16, Z31, Z27)
	BUTTERFLYD2Q(Z8, Z10, Z16, Z31, Z27)
	BUTTERFLYD2Q(Z9, Z11, Z16, Z31, Z27)
	BUTTERFLYD2Q(Z12, Z14, Z16, Z31, Z27)
	BUTTERFLYD2Q(Z13, Z15, Z16, Z31, Z27)
	MULD(Z2, Z18, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z3, Z19, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z6, Z18, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z7, Z19, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z10, Z18, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z11, Z19, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z14, Z18, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z15, Z19, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	ADDQ         $24, CX
	MOVQ         0(CX), DI
	VMOVDQU32    0(DI), Z18
	BUTTERFLYD2Q(Z0, Z1, Z16, Z31, Z27)
	BUTTERFLYD2Q(Z2, Z3, Z16, Z31, Z27)
	BUTTERFLYD2Q(Z4, Z5, Z16, Z31, Z27)
	BUTTERFLYD2Q(Z6, Z7, Z16, Z31, Z27)
	BUTTERFLYD2Q(Z8, Z9, Z16, Z31, Z27)
	BUTTERFLYD2Q(Z10, Z11, Z16, Z31, Z27)
	BUTTERFLYD2Q(Z12, Z13, Z16, Z31, Z27)
	BUTTERFLYD2Q(Z14, Z15, Z16, Z31, Z27)
	MULD(Z1, Z18, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z3, Z18, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z5, Z18, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z7, Z18, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z9, Z18, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z11, Z18, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z13, Z18, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z15, Z18, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	ADDQ         $24, CX
	MOVQ         ·vInterleaveIndices+0(SB), R8
	VMOVDQU64    0(R8), Z22
	MOVQ         0(CX), DI
	VMOVDQU32    0(DI), Y18
	VINSERTI64X4 $1, Y18, Z18, Z18
	MOVQ         24(CX), DI
	VMOVDQU32    0(DI), X19
	VINSERTI64X2 $1, X19, Z19, Z19
	VINSERTI64X2 $0x0000000000000002, X19, Z19, Z19
	VINSERTI64X2 $0x0000000000000003, X19, Z19, Z19
	MOVQ         48(CX), DI
	VPBROADCASTD 0(DI), Z20
	VPBROADCASTD 4(DI), Z21
	VPBLENDMD    Z20, Z21, K3, Z20
	PERMUTE8X8(Z0, Z1, Z26)
	BUTTERFLYD2Q(Z0, Z1, Z16, Z31, Z27)
	PERMUTE8X8(Z2, Z3, Z26)
	BUTTERFLYD2Q(Z2, Z3, Z16, Z31, Z27)
	PERMUTE8X8(Z4, Z5, Z26)
	BUTTERFLYD2Q(Z4, Z5, Z16, Z31, Z27)
	PERMUTE8X8(Z6, Z7, Z26)
	BUTTERFLYD2Q(Z6, Z7, Z16, Z31, Z27)
	PERMUTE8X8(Z8, Z9, Z26)
	BUTTERFLYD2Q(Z8, Z9, Z16, Z31, Z27)
	PERMUTE8X8(Z10, Z11, Z26)
	BUTTERFLYD2Q(Z10, Z11, Z16, Z31, Z27)
	PERMUTE8X8(Z12, Z13, Z26)
	BUTTERFLYD2Q(Z12, Z13, Z16, Z31, Z27)
	PERMUTE8X8(Z14, Z15, Z26)
	BUTTERFLYD2Q(Z14, Z15, Z16, Z31, Z27)
	MULD(Z1, Z18, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z3, Z18, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z5, Z18, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z7, Z18, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z9, Z18, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z11, Z18, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z13, Z18, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z15, Z18, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	PERMUTE4X4(Z0, Z1, Z22, Z26)
	BUTTERFLYD2Q(Z0, Z1, Z16, Z31, Z27)
	PERMUTE4X4(Z2, Z3, Z22, Z26)
	BUTTERFLYD2Q(Z2, Z3, Z16, Z31, Z27)
	PERMUTE4X4(Z4, Z5, Z22, Z26)
	BUTTERFLYD2Q(Z4, Z5, Z16, Z31, Z27)
	PERMUTE4X4(Z6, Z7, Z22, Z26)
	BUTTERFLYD2Q(Z6, Z7, Z16, Z31, Z27)
	PERMUTE4X4(Z8, Z9, Z22, Z26)
	BUTTERFLYD2Q(Z8, Z9, Z16, Z31, Z27)
	PERMUTE4X4(Z10, Z11, Z22, Z26)
	BUTTERFLYD2Q(Z10, Z11, Z16, Z31, Z27)
	PERMUTE4X4(Z12, Z13, Z22, Z26)
	BUTTERFLYD2Q(Z12, Z13, Z16, Z31, Z27)
	PERMUTE4X4(Z14, Z15, Z22, Z26)
	BUTTERFLYD2Q(Z14, Z15, Z16, Z31, Z27)
	MULD(Z1, Z19, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z3, Z19, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z5, Z19, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z7, Z19, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z9, Z19, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z11, Z19, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z13, Z19, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	MULD(Z15, Z19, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	PERMUTE2X2(Z0, Z1, Z26)
	BUTTERFLYD2Q(Z0, Z1, Z16, Z31, Z27)
	PERMUTE2X2(Z2, Z3, Z26)
	BUTTERFLYD2Q(Z2, Z3, Z16, Z31, Z27)
	PERMUTE2X2(Z4, Z5, Z26)
	BUTTERFLYD2Q(Z4, Z5, Z16, Z31, Z27)
	PERMUTE2X2(Z6, Z7, Z26)
	BUTTERFLYD2Q(Z6, Z7, Z16, Z31, Z27)
	PERMUTE2X2(Z8, Z9, Z26)
	BUTTERFLYD2Q(Z8, Z9, Z16, Z31, Z27)
	PERMUTE2X2(Z10, Z11, Z26)
	BUTTERFLYD2Q(Z10, Z11, Z16, Z31, Z27)
	PERMUTE2X2(Z12, Z13, Z26)
	BUTTERFLYD2Q(Z12, Z13, Z16, Z31, Z27)
	PERMUTE2X2(Z14, Z15, Z26)
	BUTTERFLYD2Q(Z14, Z15, Z16, Z31, Z27)
	MULD(Z1, Z20, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	PERMUTE1X1(Z0, Z1, Z26)
	MULD(Z3, Z20, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	PERMUTE1X1(Z2, Z3, Z26)
	MULD(Z5, Z20, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	PERMUTE1X1(Z4, Z5, Z26)
	MULD(Z7, Z20, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	PERMUTE1X1(Z6, Z7, Z26)
	MULD(Z9, Z20, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	PERMUTE1X1(Z8, Z9, Z26)
	MULD(Z11, Z20, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	PERMUTE1X1(Z10, Z11, Z26)
	MULD(Z13, Z20, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	PERMUTE1X1(Z12, Z13, Z26)
	MULD(Z15, Z20, Z28, Z29, Z26, Z27, Z30, Z31, Z16, Z17)
	PERMUTE1X1(Z14, Z15, Z26)
	BUTTERFLYD1Q(Z0, Z1, Z16, Z26, Z27)
	BUTTERFLYD1Q(Z2, Z3, Z16, Z26, Z27)
	BUTTERFLYD1Q(Z4, Z5, Z16, Z26, Z27)
	BUTTERFLYD1Q(Z6, Z7, Z16, Z26, Z27)
	BUTTERFLYD1Q(Z8, Z9, Z16, Z26, Z27)
	BUTTERFLYD1Q(Z10, Z11, Z16, Z26, Z27)
	BUTTERFLYD1Q(Z12, Z13, Z16, Z26, Z27)
	BUTTERFLYD1Q(Z14, Z15, Z16, Z26, Z27)
	VPUNPCKLDQ   Z1, Z0, Z23
	VPUNPCKHDQ   Z1, Z0, Z1
	VMOVDQA32    Z23, Z0
	PERMUTE4X4(Z0, Z1, Z22, Z23)
	PERMUTE8X8(Z0, Z1, Z23)
	VMOVDQU32    Z0, 0(R15)
	VMOVDQU32    Z1, 64(R15)
	VPUNPCKLDQ   Z3, Z2, Z23
	VPUNPCKHDQ   Z3, Z2, Z3
	VMOVDQA32    Z23, Z2
	PERMUTE4X4(Z2, Z3, Z22, Z23)
	PERMUTE8X8(Z2, Z3, Z23)
	VMOVDQU32    Z2, 128(R15)
	VMOVDQU32    Z3, 192(R15)
	VPUNPCKLDQ   Z5, Z4, Z23
	VPUNPCKHDQ   Z5, Z4, Z5
	VMOVDQA32    Z23, Z4
	PERMUTE4X4(Z4, Z5, Z22, Z23)
	PERMUTE8X8(Z4, Z5, Z23)
	VMOVDQU32    Z4, 256(R15)
	VMOVDQU32    Z5, 320(R15)
	VPUNPCKLDQ   Z7, Z6, Z23
	VPUNPCKHDQ   Z7, Z6, Z7
	VMOVDQA32    Z23, Z6
	PERMUTE4X4(Z6, Z7, Z22, Z23)
	PERMUTE8X8(Z6, Z7, Z23)
	VMOVDQU32    Z6, 384(R15)
	VMOVDQU32    Z7, 448(R15)
	VPUNPCKLDQ   Z9, Z8, Z23
	VPUNPCKHDQ   Z9, Z8, Z9
	VMOVDQA32    Z23, Z8
	PERMUTE4X4(Z8, Z9, Z22, Z23)
	PERMUTE8X8(Z8, Z9, Z23)
	VMOVDQU32    Z8, 512(R15)
	VMOVDQU32    Z9, 576(R15)
	VPUNPCKLDQ   Z11, Z10, Z23
	VPUNPCKHDQ   Z11, Z10, Z11
	VMOVDQA32    Z23, Z10
	PERMUTE4X4(Z10, Z11, Z22, Z23)
	PERMUTE8X8(Z10, Z11, Z23)
	VMOVDQU32    Z10, 640(R15)
	VMOVDQU32    Z11, 704(R15)
	VPUNPCKLDQ   Z13, Z12, Z23
	VPUNPCKHDQ   Z13, Z12, Z13
	VMOVDQA32    Z23, Z12
	PERMUTE4X4(Z12, Z13, Z22, Z23)
	PERMUTE8X8(Z12, Z13, Z23)
	VMOVDQU32    Z12, 768(R15)
	VMOVDQU32    Z13, 832(R15)
	VPUNPCKLDQ   Z15, Z14, Z23
	VPUNPCKHDQ   Z15, Z14, Z15
	VMOVDQA32    Z23, Z14
	PERMUTE4X4(Z14, Z15, Z22, Z23)
	PERMUTE8X8(Z14, Z15, Z23)
	VMOVDQU32    Z14, 896(R15)
	VMOVDQU32    Z15, 960(R15)
	RET

TEXT ·kerDITNP_256_avx512(SB), NOSPLIT, $0-56
	LOAD_Q(Z16, Z17)
	LOAD_MASKS()

	// load arguments
	MOVQ         a+0(FP), R15
	MOVQ         twiddles+24(FP), CX
	MOVQ         stage+48(FP), AX
	IMULQ        $24, AX
	ADDQ         AX, CX                             // we want twiddles[stage] as starting point
	VMOVDQU32    0(R15), Z0                         // load a[0]
	VMOVDQU32    64(R15), Z1                        // load a[1]
	VMOVDQU32    128(R15), Z2                       // load a[2]
	VMOVDQU32    192(R15), Z3                       // load a[3]
	VMOVDQU32    256(R15), Z4                       // load a[4]
	VMOVDQU32    320(R15), Z5                       // load a[5]
	VMOVDQU32    384(R15), Z6                       // load a[6]
	VMOVDQU32    448(R15), Z7                       // load a[7]
	VMOVDQU32    512(R15), Z8                       // load a[8]
	VMOVDQU32    576(R15), Z9                       // load a[9]
	VMOVDQU32    640(R15), Z10                      // load a[10]
	VMOVDQU32    704(R15), Z11                      // load a[11]
	VMOVDQU32    768(R15), Z12                      // load a[12]
	VMOVDQU32    832(R15), Z13                      // load a[13]
	VMOVDQU32    896(R15), Z14                      // load a[14]
	VMOVDQU32    960(R15), Z15                      // load a[15]
	MOVQ         ·vInterleaveIndices+0(SB), R8
	VMOVDQU64    0(R8), Z28
	PERMUTE1X1(Z0, Z1, Z22)
	BUTTERFLYD1Q(Z0, Z1, Z16, Z22, Z23)
	PERMUTE1X1(Z0, Z1, Z22)
	PERMUTE1X1(Z2, Z3, Z22)
	BUTTERFLYD1Q(Z2, Z3, Z16, Z22, Z23)
	PERMUTE1X1(Z2, Z3, Z22)
	PERMUTE1X1(Z4, Z5, Z22)
	BUTTERFLYD1Q(Z4, Z5, Z16, Z22, Z23)
	PERMUTE1X1(Z4, Z5, Z22)
	PERMUTE1X1(Z6, Z7, Z22)
	BUTTERFLYD1Q(Z6, Z7, Z16, Z22, Z23)
	PERMUTE1X1(Z6, Z7, Z22)
	PERMUTE1X1(Z8, Z9, Z22)
	BUTTERFLYD1Q(Z8, Z9, Z16, Z22, Z23)
	PERMUTE1X1(Z8, Z9, Z22)
	PERMUTE1X1(Z10, Z11, Z22)
	BUTTERFLYD1Q(Z10, Z11, Z16, Z22, Z23)
	PERMUTE1X1(Z10, Z11, Z22)
	PERMUTE1X1(Z12, Z13, Z22)
	BUTTERFLYD1Q(Z12, Z13, Z16, Z22, Z23)
	PERMUTE1X1(Z12, Z13, Z22)
	PERMUTE1X1(Z14, Z15, Z22)
	BUTTERFLYD1Q(Z14, Z15, Z16, Z22, Z23)
	PERMUTE1X1(Z14, Z15, Z22)
	MOVQ         $0x0000000000000006, AX
	IMULQ        $24, AX
	ADDQ         AX, CX
	MOVQ         0(CX), DI
	VPBROADCASTD 0(DI), Z20
	VPBROADCASTD 4(DI), Z21
	VPBLENDMD    Z20, Z21, K3, Z20
	PERMUTE2X2(Z0, Z1, Z22)
	MULD(Z1, Z20, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z0, Z1, Z16, Z22, Z23)
	PERMUTE2X2(Z0, Z1, Z22)
	PERMUTE2X2(Z2, Z3, Z22)
	MULD(Z3, Z20, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z2, Z3, Z16, Z22, Z23)
	PERMUTE2X2(Z2, Z3, Z22)
	PERMUTE2X2(Z4, Z5, Z22)
	MULD(Z5, Z20, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z4, Z5, Z16, Z22, Z23)
	PERMUTE2X2(Z4, Z5, Z22)
	PERMUTE2X2(Z6, Z7, Z22)
	MULD(Z7, Z20, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z6, Z7, Z16, Z22, Z23)
	PERMUTE2X2(Z6, Z7, Z22)
	PERMUTE2X2(Z8, Z9, Z22)
	MULD(Z9, Z20, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z8, Z9, Z16, Z22, Z23)
	PERMUTE2X2(Z8, Z9, Z22)
	PERMUTE2X2(Z10, Z11, Z22)
	MULD(Z11, Z20, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z10, Z11, Z16, Z22, Z23)
	PERMUTE2X2(Z10, Z11, Z22)
	PERMUTE2X2(Z12, Z13, Z22)
	MULD(Z13, Z20, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z12, Z13, Z16, Z22, Z23)
	PERMUTE2X2(Z12, Z13, Z22)
	PERMUTE2X2(Z14, Z15, Z22)
	MULD(Z15, Z20, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z14, Z15, Z16, Z22, Z23)
	PERMUTE2X2(Z14, Z15, Z22)
	SUBQ         $24, CX
	MOVQ         0(CX), DI
	VMOVDQU32    0(DI), X19
	VINSERTI64X2 $1, X19, Z19, Z19
	VINSERTI64X2 $0x0000000000000002, X19, Z19, Z19
	VINSERTI64X2 $0x0000000000000003, X19, Z19, Z19
	PERMUTE4X4(Z0, Z1, Z28, Z22)
	MULD(Z1, Z19, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z0, Z1, Z16, Z22, Z23)
	PERMUTE4X4(Z0, Z1, Z28, Z22)
	PERMUTE4X4(Z2, Z3, Z28, Z22)
	MULD(Z3, Z19, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z2, Z3, Z16, Z22, Z23)
	PERMUTE4X4(Z2, Z3, Z28, Z22)
	PERMUTE4X4(Z4, Z5, Z28, Z22)
	MULD(Z5, Z19, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z4, Z5, Z16, Z22, Z23)
	PERMUTE4X4(Z4, Z5, Z28, Z22)
	PERMUTE4X4(Z6, Z7, Z28, Z22)
	MULD(Z7, Z19, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z6, Z7, Z16, Z22, Z23)
	PERMUTE4X4(Z6, Z7, Z28, Z22)
	PERMUTE4X4(Z8, Z9, Z28, Z22)
	MULD(Z9, Z19, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z8, Z9, Z16, Z22, Z23)
	PERMUTE4X4(Z8, Z9, Z28, Z22)
	PERMUTE4X4(Z10, Z11, Z28, Z22)
	MULD(Z11, Z19, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z10, Z11, Z16, Z22, Z23)
	PERMUTE4X4(Z10, Z11, Z28, Z22)
	PERMUTE4X4(Z12, Z13, Z28, Z22)
	MULD(Z13, Z19, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z12, Z13, Z16, Z22, Z23)
	PERMUTE4X4(Z12, Z13, Z28, Z22)
	PERMUTE4X4(Z14, Z15, Z28, Z22)
	MULD(Z15, Z19, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z14, Z15, Z16, Z22, Z23)
	PERMUTE4X4(Z14, Z15, Z28, Z22)
	SUBQ         $24, CX
	MOVQ         0(CX), DI
	VMOVDQU32    0(DI), Y18
	VINSERTI64X4 $1, Y18, Z18, Z18
	PERMUTE8X8(Z0, Z1, Z22)
	MULD(Z1, Z18, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z0, Z1, Z16, Z22, Z23)
	PERMUTE8X8(Z0, Z1, Z22)
	PERMUTE8X8(Z2, Z3, Z22)
	MULD(Z3, Z18, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z2, Z3, Z16, Z22, Z23)
	PERMUTE8X8(Z2, Z3, Z22)
	PERMUTE8X8(Z4, Z5, Z22)
	MULD(Z5, Z18, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z4, Z5, Z16, Z22, Z23)
	PERMUTE8X8(Z4, Z5, Z22)
	PERMUTE8X8(Z6, Z7, Z22)
	MULD(Z7, Z18, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z6, Z7, Z16, Z22, Z23)
	PERMUTE8X8(Z6, Z7, Z22)
	PERMUTE8X8(Z8, Z9, Z22)
	MULD(Z9, Z18, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z8, Z9, Z16, Z22, Z23)
	PERMUTE8X8(Z8, Z9, Z22)
	PERMUTE8X8(Z10, Z11, Z22)
	MULD(Z11, Z18, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z10, Z11, Z16, Z22, Z23)
	PERMUTE8X8(Z10, Z11, Z22)
	PERMUTE8X8(Z12, Z13, Z22)
	MULD(Z13, Z18, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z12, Z13, Z16, Z22, Z23)
	PERMUTE8X8(Z12, Z13, Z22)
	PERMUTE8X8(Z14, Z15, Z22)
	MULD(Z15, Z18, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z14, Z15, Z16, Z22, Z23)
	PERMUTE8X8(Z14, Z15, Z22)
	SUBQ         $24, CX
	MOVQ         0(CX), DI
	VMOVDQU32    0(DI), Z29
	MULD(Z1, Z29, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z0, Z1, Z16, Z22, Z23)
	MULD(Z3, Z29, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z2, Z3, Z16, Z22, Z23)
	MULD(Z5, Z29, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z4, Z5, Z16, Z22, Z23)
	MULD(Z7, Z29, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z6, Z7, Z16, Z22, Z23)
	MULD(Z9, Z29, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z8, Z9, Z16, Z22, Z23)
	MULD(Z11, Z29, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z10, Z11, Z16, Z22, Z23)
	MULD(Z13, Z29, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z12, Z13, Z16, Z22, Z23)
	MULD(Z15, Z29, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z14, Z15, Z16, Z22, Z23)
	SUBQ         $24, CX
	MOVQ         0(CX), DI
	VMOVDQU32    0(DI), Z29
	VMOVDQU32    64(DI), Z30
	MULD(Z2, Z29, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z0, Z2, Z16, Z22, Z23)
	MULD(Z3, Z30, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z1, Z3, Z16, Z22, Z23)
	MULD(Z6, Z29, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z4, Z6, Z16, Z22, Z23)
	MULD(Z7, Z30, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z5, Z7, Z16, Z22, Z23)
	MULD(Z10, Z29, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z8, Z10, Z16, Z22, Z23)
	MULD(Z11, Z30, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z9, Z11, Z16, Z22, Z23)
	MULD(Z14, Z29, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z12, Z14, Z16, Z22, Z23)
	MULD(Z15, Z30, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z13, Z15, Z16, Z22, Z23)
	SUBQ         $24, CX
	MOVQ         0(CX), DI
	VMOVDQU32    0(DI), Z29
	VMOVDQU32    64(DI), Z30
	VMOVDQU32    128(DI), Z31
	VMOVDQU32    192(DI), Z18
	MULD(Z4, Z29, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z0, Z4, Z16, Z22, Z23)
	MULD(Z5, Z30, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z1, Z5, Z16, Z22, Z23)
	MULD(Z6, Z31, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z2, Z6, Z16, Z22, Z23)
	MULD(Z7, Z18, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z3, Z7, Z16, Z22, Z23)
	MULD(Z12, Z29, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z8, Z12, Z16, Z22, Z23)
	MULD(Z13, Z30, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z9, Z13, Z16, Z22, Z23)
	MULD(Z14, Z31, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z10, Z14, Z16, Z22, Z23)
	MULD(Z15, Z18, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z11, Z15, Z16, Z22, Z23)
	SUBQ         $24, CX
	MOVQ         0(CX), DI
	VMOVDQU32    0(DI), Z29
	VMOVDQU32    64(DI), Z30
	VMOVDQU32    128(DI), Z31
	VMOVDQU32    192(DI), Z18
	VMOVDQU32    256(DI), Z19
	VMOVDQU32    320(DI), Z20
	VMOVDQU32    384(DI), Z21
	VMOVDQU32    448(DI), Z28
	MULD(Z8, Z29, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z0, Z8, Z16, Z22, Z23)
	MULD(Z9, Z30, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z1, Z9, Z16, Z22, Z23)
	MULD(Z10, Z31, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z2, Z10, Z16, Z22, Z23)
	MULD(Z11, Z18, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z3, Z11, Z16, Z22, Z23)
	MULD(Z12, Z19, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z4, Z12, Z16, Z22, Z23)
	MULD(Z13, Z20, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z5, Z13, Z16, Z22, Z23)
	MULD(Z14, Z21, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z6, Z14, Z16, Z22, Z23)
	MULD(Z15, Z28, Z24, Z25, Z22, Z23, Z26, Z27, Z16, Z17)
	BUTTERFLYD1Q(Z7, Z15, Z16, Z22, Z23)
	VMOVDQU32    Z0, 0(R15)
	VMOVDQU32    Z1, 64(R15)
	VMOVDQU32    Z2, 128(R15)
	VMOVDQU32    Z3, 192(R15)
	VMOVDQU32    Z4, 256(R15)
	VMOVDQU32    Z5, 320(R15)
	VMOVDQU32    Z6, 384(R15)
	VMOVDQU32    Z7, 448(R15)
	VMOVDQU32    Z8, 512(R15)
	VMOVDQU32    Z9, 576(R15)
	VMOVDQU32    Z10, 640(R15)
	VMOVDQU32    Z11, 704(R15)
	VMOVDQU32    Z12, 768(R15)
	VMOVDQU32    Z13, 832(R15)
	VMOVDQU32    Z14, 896(R15)
	VMOVDQU32    Z15, 960(R15)
	RET
