//go:build !purego
// Code generated by gnark-crypto/generator. DO NOT EDIT.
#include "textflag.h"
#include "funcdata.h"
#include "go_asm.h"

// performs a butterfly between 2 vectors of dwords
// in0 = (in0 + in1) mod q
// in1 = (in0 - in1) mod 2q
// in2: q broadcasted on all dwords lanes
// in3: temporary Z register
#define BUTTERFLYD2Q(in0, in1, in2, in3) \
	VPADDD  in0, in1, in3 \
	VPSUBD  in1, in0, in1 \
	VPSUBD  in2, in3, in0 \
	VPMINUD in3, in0, in0 \
	VPADDD  in2, in1, in1 \

// same as butterflyD2Q but reduces in1 to [0,q)
#define BUTTERFLYD1Q(in0, in1, in2, in3, in4) \
	VPADDD  in0, in1, in3 \
	VPSUBD  in1, in0, in1 \
	VPSUBD  in2, in3, in0 \
	VPMINUD in3, in0, in0 \
	VPADDD  in2, in1, in4 \
	VPMINUD in4, in1, in1 \

// same as butterflyD2Q but for qwords
// in2: must be broadcasted on all qwords lanes
#define BUTTERFLYQ2Q(in0, in1, in2, in3) \
	VPADDQ  in0, in1, in3 \
	VPSUBQ  in1, in0, in1 \
	VPSUBQ  in2, in3, in0 \
	VPMINUQ in3, in0, in0 \
	VPADDQ  in2, in1, in1 \

#define BUTTERFLYQ1Q(in0, in1, in2, in3, in4) \
	VPADDQ  in0, in1, in3 \
	VPSUBQ  in1, in0, in1 \
	VPSUBQ  in2, in3, in0 \
	VPMINUQ in3, in0, in0 \
	VPADDQ  in2, in1, in4 \
	VPMINUQ in4, in1, in1 \

// performs a multiplication in place between 2 vectors of qwords (values should be dwords zero extended)
// in0 = (in0 * in1) mod q
// in1: second operand
// in2: mask for low dword in each qword
// in3: q broadcasted on all qwords lanes
// in4: qInvNeg broadcasted on all qwords lanes
// in5: temporary Z register
// in6: temporary Z register
#define MULQ(in0, in1, in2, in3, in4, in5, in6) \
	VPMULUDQ in0, in1, in5 \
	VPANDQ   in2, in5, in6 \
	VPMULUDQ in6, in4, in6 \
	VPANDQ   in2, in6, in6 \
	VPMULUDQ in6, in3, in6 \
	VPADDQ   in5, in6, in5 \
	VPSRLQ   $32, in5, in5 \
	VPSUBQ   in3, in5, in6 \
	VPMINUQ  in5, in6, in0 \

#define MULD(in0, in1, in2, in3, in4, in5, in6, in7, in8, in9, in10) \
	VMOVSHDUP in0, in2       \
	VMOVSHDUP in1, in3       \
	VPMULUDQ  in0, in1, in4  \
	VPMULUDQ  in2, in3, in5  \
	VPMULUDQ  in4, in9, in6  \
	VPMULUDQ  in5, in9, in7  \
	VPMULUDQ  in6, in8, in6  \
	VPMULUDQ  in7, in8, in7  \
	VPADDQ    in4, in6, in4  \
	VPADDQ    in5, in7, in5  \
	VMOVSHDUP in4, in10, in5 \
	VPSUBD    in8, in5, in7  \
	VPMINUD   in5, in7, in0  \

// goes from
// Z1 = A A A A B B B B
// Z2 = C C C C D D D D
// we want
// Z1 = A A A A C C C C
// Z2 = B B B B D D D D
#define PERMUTE8X8(in0, in1, in2, in3) \
	VSHUFI64X2 $0x000000000000004e, in1, in0, in2 \
	VPBLENDMQ  in0, in2, in3, in0                 \
	VPBLENDMQ  in2, in1, in3, in1                 \

// Z1 = A A B B C C D D
// Z2 = L L M M N N O O
// we want
// Z1 = A A L L C C N N
// Z2 = B B M M D D O O
#define PERMUTE4X4(in0, in1, in2, in3, in4) \
	VMOVDQA64 in2, in3           \
	VPERMI2Q  in1, in0, in3      \
	VPBLENDMQ in0, in3, in4, in0 \
	VPBLENDMQ in3, in1, in4, in1 \

#define PERMUTE2X2(in0, in1, in2, in3) \
	VSHUFPD   $0x0000000000000055, in1, in0, in2 \
	VPBLENDMQ in0, in2, in3, in0                 \
	VPBLENDMQ in2, in1, in3, in1                 \

#define PERMUTE1X1(in0, in1, in2, in3) \
	VPSHRDQ   $32, in1, in0, in2 \
	VPBLENDMD in0, in2, in3, in0 \
	VPBLENDMD in2, in1, in3, in1 \

#define PACK_DWORDS(in0, in1, in2, in3) \
	VPMOVQD      in0, in1          \
	VPMOVQD      in2, in3          \
	VINSERTI64X4 $1, in3, in0, in0 \

TEXT ·innerDITWithTwiddles_avx512(SB), NOSPLIT, $0-72
	// prepare constants needed for mul and reduce ops
	MOVD         $const_q, AX
	VPBROADCASTQ AX, Z8
	MOVD         $const_qInvNeg, AX
	VPBROADCASTQ AX, Z9
	VPCMPEQB     Y0, Y0, Y0
	VPMOVZXDQ    Y0, Z11

	// load arguments
	MOVQ a+0(FP), R15
	MOVQ twiddles+24(FP), CX
	MOVQ end+56(FP), SI
	MOVQ m+64(FP), BX
	CMPQ BX, $0x0000000000000008
	JL   smallerThan8_1          // m < 8
	SHRQ $3, SI                  // we are processing 8 elements at a time
	SHLQ $2, BX                  // offset = m * 4bytes
	MOVQ R15, DX
	ADDQ BX, DX

loop_3:
	TESTQ     SI, SI
	JEQ       done_2     // n == 0, we are done
	VPMOVZXDQ 0(R15), Z0 // load a[i]
	VPMOVZXDQ 0(DX), Z1  // load a[i+m]
	VPMOVZXDQ 0(CX), Z15
	MULQ(Z1, Z15, Z11, Z8, Z9, Z12, Z10)
	BUTTERFLYQ1Q(Z0, Z1, Z8, Z3, Z4)
	VPMOVQD   Z0, 0(R15) // store a[i]
	VPMOVQD   Z1, 0(DX)  // store a[i+m]
	ADDQ      $32, R15
	ADDQ      $32, DX
	ADDQ      $32, CX
	DECQ      SI         // decrement n
	JMP       loop_3

done_2:
	RET

smallerThan8_1:
	// m < 8, we call the generic one
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
	// prepare constants needed for mul and reduce ops
	MOVD         $const_q, AX
	VPBROADCASTD AX, Z2
	VPBROADCASTD AX, Z4
	MOVD         $const_qInvNeg, AX
	VPBROADCASTD AX, Z5

	// load arguments
	MOVQ  a+0(FP), R15
	MOVQ  twiddles+24(FP), CX
	MOVQ  end+56(FP), SI
	MOVQ  m+64(FP), BX
	CMPQ  BX, $0x0000000000000010
	JL    smallerThan16_4         // m < 16
	SHRQ  $4, SI                  // we are processing 16 elements at a time
	SHLQ  $2, BX                  // offset = m * 4bytes
	MOVQ  R15, DX
	ADDQ  BX, DX
	MOVQ  $0x0000000000005555, AX
	KMOVD AX, K3

loop_6:
	TESTQ     SI, SI
	JEQ       done_5     // n == 0, we are done
	VMOVDQU32 0(R15), Z0 // load a[i]
	VMOVDQU32 0(DX), Z1  // load a[i+m]
	BUTTERFLYD2Q(Z0, Z1, Z2, Z3)
	VMOVDQU32 Z0, 0(R15) // store a[i]
	VMOVDQU32 0(CX), Z6
	MULD(Z1, Z6, Z7, Z8, Z3, Z9, Z10, Z11, Z4, Z5, K3)
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

// kerDIFNP_256_avx512(a []{{ .FF }}.Element, twiddles [][]{{ .FF }}.Element, stage int)
TEXT ·kerDIFNP_256_avx512(SB), NOSPLIT, $0-56
	// prepare constants needed for mul and reduce ops
	MOVD         $const_q, AX
	VPBROADCASTD AX, Z24
	MOVD         $const_qInvNeg, AX
	VPBROADCASTD AX, Z25

	// load arguments
	MOVQ         a+0(FP), R15
	MOVQ         twiddles+24(FP), CX
	MOVQ         stage+48(FP), AX
	IMULQ        $24, AX
	ADDQ         AX, CX                             // we want twiddles[stage] as starting point
	MOVQ         $0x0000000000000f0f, AX
	KMOVQ        AX, K1
	MOVQ         $0x0000000000000033, AX
	KMOVQ        AX, K2
	MOVQ         $0x0000000000005555, AX
	KMOVD        AX, K3
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
	MOVQ         0(CX), BX
	VMOVDQU32    0(BX), Z16
	VMOVDQU32    64(BX), Z17
	VMOVDQU32    128(BX), Z18
	VMOVDQU32    192(BX), Z19
	VMOVDQU32    256(BX), Z20
	VMOVDQU32    320(BX), Z21
	VMOVDQU32    384(BX), Z22
	VMOVDQU32    448(BX), Z23
	BUTTERFLYD2Q(Z0, Z8, Z24, Z26)
	MULD(Z8, Z16, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	BUTTERFLYD2Q(Z1, Z9, Z24, Z26)
	MULD(Z9, Z17, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	BUTTERFLYD2Q(Z2, Z10, Z24, Z26)
	MULD(Z10, Z18, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	BUTTERFLYD2Q(Z3, Z11, Z24, Z26)
	MULD(Z11, Z19, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	BUTTERFLYD2Q(Z4, Z12, Z24, Z26)
	MULD(Z12, Z20, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	BUTTERFLYD2Q(Z5, Z13, Z24, Z26)
	MULD(Z13, Z21, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	BUTTERFLYD2Q(Z6, Z14, Z24, Z26)
	MULD(Z14, Z22, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	BUTTERFLYD2Q(Z7, Z15, Z24, Z26)
	MULD(Z15, Z23, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	ADDQ         $24, CX
	MOVQ         0(CX), BX
	VMOVDQU32    0(BX), Z16
	VMOVDQU32    64(BX), Z17
	VMOVDQU32    128(BX), Z18
	VMOVDQU32    192(BX), Z19
	BUTTERFLYD2Q(Z0, Z4, Z24, Z26)
	MULD(Z4, Z16, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	BUTTERFLYD2Q(Z1, Z5, Z24, Z26)
	MULD(Z5, Z17, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	BUTTERFLYD2Q(Z2, Z6, Z24, Z26)
	MULD(Z6, Z18, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	BUTTERFLYD2Q(Z3, Z7, Z24, Z26)
	MULD(Z7, Z19, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	BUTTERFLYD2Q(Z8, Z12, Z24, Z26)
	MULD(Z12, Z16, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	BUTTERFLYD2Q(Z9, Z13, Z24, Z26)
	MULD(Z13, Z17, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	BUTTERFLYD2Q(Z10, Z14, Z24, Z26)
	MULD(Z14, Z18, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	BUTTERFLYD2Q(Z11, Z15, Z24, Z26)
	MULD(Z15, Z19, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	ADDQ         $24, CX
	MOVQ         0(CX), BX
	VMOVDQU32    0(BX), Z16
	VMOVDQU32    64(BX), Z17
	BUTTERFLYD2Q(Z0, Z2, Z24, Z26)
	MULD(Z2, Z16, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	BUTTERFLYD2Q(Z1, Z3, Z24, Z26)
	MULD(Z3, Z17, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	BUTTERFLYD2Q(Z4, Z6, Z24, Z26)
	MULD(Z6, Z16, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	BUTTERFLYD2Q(Z5, Z7, Z24, Z26)
	MULD(Z7, Z17, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	BUTTERFLYD2Q(Z8, Z10, Z24, Z26)
	MULD(Z10, Z16, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	BUTTERFLYD2Q(Z9, Z11, Z24, Z26)
	MULD(Z11, Z17, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	BUTTERFLYD2Q(Z12, Z14, Z24, Z26)
	MULD(Z14, Z16, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	BUTTERFLYD2Q(Z13, Z15, Z24, Z26)
	MULD(Z15, Z17, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	ADDQ         $24, CX
	MOVQ         0(CX), BX
	VMOVDQU32    0(BX), Z16
	BUTTERFLYD2Q(Z0, Z1, Z24, Z26)
	MULD(Z1, Z16, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	BUTTERFLYD2Q(Z2, Z3, Z24, Z26)
	MULD(Z3, Z16, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	BUTTERFLYD2Q(Z4, Z5, Z24, Z26)
	MULD(Z5, Z16, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	BUTTERFLYD2Q(Z6, Z7, Z24, Z26)
	MULD(Z7, Z16, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	BUTTERFLYD2Q(Z8, Z9, Z24, Z26)
	MULD(Z9, Z16, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	BUTTERFLYD2Q(Z10, Z11, Z24, Z26)
	MULD(Z11, Z16, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	BUTTERFLYD2Q(Z12, Z13, Z24, Z26)
	MULD(Z13, Z16, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	BUTTERFLYD2Q(Z14, Z15, Z24, Z26)
	MULD(Z15, Z16, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	ADDQ         $24, CX
	MOVQ         ·vInterleaveIndices+0(SB), DI
	VMOVDQU64    0(DI), Z20
	MOVQ         0(CX), BX
	VMOVDQU32    0(BX), Y16
	VINSERTI64X4 $1, Y16, Z16, Z16
	MOVQ         24(CX), BX
	VMOVDQU32    0(BX), X17
	VINSERTI64X2 $1, X17, Z17, Z17
	VINSERTI64X2 $0x0000000000000002, X17, Z17, Z17
	VINSERTI64X2 $0x0000000000000003, X17, Z17, Z17
	MOVQ         48(CX), BX
	VPBROADCASTD 0(BX), Z18
	VPBROADCASTD 4(BX), Z19
	VPBLENDMD    Z18, Z19, K3, Z18
	PERMUTE8X8(Z0, Z1, Z26, K1)
	BUTTERFLYD2Q(Z0, Z1, Z24, Z26)
	MULD(Z1, Z16, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	PERMUTE8X8(Z2, Z3, Z26, K1)
	BUTTERFLYD2Q(Z2, Z3, Z24, Z26)
	MULD(Z3, Z16, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	PERMUTE8X8(Z4, Z5, Z26, K1)
	BUTTERFLYD2Q(Z4, Z5, Z24, Z26)
	MULD(Z5, Z16, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	PERMUTE8X8(Z6, Z7, Z26, K1)
	BUTTERFLYD2Q(Z6, Z7, Z24, Z26)
	MULD(Z7, Z16, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	PERMUTE8X8(Z8, Z9, Z26, K1)
	BUTTERFLYD2Q(Z8, Z9, Z24, Z26)
	MULD(Z9, Z16, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	PERMUTE8X8(Z10, Z11, Z26, K1)
	BUTTERFLYD2Q(Z10, Z11, Z24, Z26)
	MULD(Z11, Z16, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	PERMUTE8X8(Z12, Z13, Z26, K1)
	BUTTERFLYD2Q(Z12, Z13, Z24, Z26)
	MULD(Z13, Z16, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	PERMUTE8X8(Z14, Z15, Z26, K1)
	BUTTERFLYD2Q(Z14, Z15, Z24, Z26)
	MULD(Z15, Z16, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	PERMUTE4X4(Z0, Z1, Z20, Z26, K2)
	BUTTERFLYD2Q(Z0, Z1, Z24, Z26)
	MULD(Z1, Z17, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	PERMUTE4X4(Z2, Z3, Z20, Z26, K2)
	BUTTERFLYD2Q(Z2, Z3, Z24, Z26)
	MULD(Z3, Z17, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	PERMUTE4X4(Z4, Z5, Z20, Z26, K2)
	BUTTERFLYD2Q(Z4, Z5, Z24, Z26)
	MULD(Z5, Z17, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	PERMUTE4X4(Z6, Z7, Z20, Z26, K2)
	BUTTERFLYD2Q(Z6, Z7, Z24, Z26)
	MULD(Z7, Z17, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	PERMUTE4X4(Z8, Z9, Z20, Z26, K2)
	BUTTERFLYD2Q(Z8, Z9, Z24, Z26)
	MULD(Z9, Z17, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	PERMUTE4X4(Z10, Z11, Z20, Z26, K2)
	BUTTERFLYD2Q(Z10, Z11, Z24, Z26)
	MULD(Z11, Z17, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	PERMUTE4X4(Z12, Z13, Z20, Z26, K2)
	BUTTERFLYD2Q(Z12, Z13, Z24, Z26)
	MULD(Z13, Z17, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	PERMUTE4X4(Z14, Z15, Z20, Z26, K2)
	BUTTERFLYD2Q(Z14, Z15, Z24, Z26)
	MULD(Z15, Z17, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	PERMUTE2X2(Z0, Z1, Z26, K3)
	BUTTERFLYD2Q(Z0, Z1, Z24, Z26)
	MULD(Z1, Z18, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	PERMUTE2X2(Z2, Z3, Z26, K3)
	BUTTERFLYD2Q(Z2, Z3, Z24, Z26)
	MULD(Z3, Z18, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	PERMUTE2X2(Z4, Z5, Z26, K3)
	BUTTERFLYD2Q(Z4, Z5, Z24, Z26)
	MULD(Z5, Z18, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	PERMUTE2X2(Z6, Z7, Z26, K3)
	BUTTERFLYD2Q(Z6, Z7, Z24, Z26)
	MULD(Z7, Z18, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	PERMUTE2X2(Z8, Z9, Z26, K3)
	BUTTERFLYD2Q(Z8, Z9, Z24, Z26)
	MULD(Z9, Z18, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	PERMUTE2X2(Z10, Z11, Z26, K3)
	BUTTERFLYD2Q(Z10, Z11, Z24, Z26)
	MULD(Z11, Z18, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	PERMUTE2X2(Z12, Z13, Z26, K3)
	BUTTERFLYD2Q(Z12, Z13, Z24, Z26)
	MULD(Z13, Z18, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	PERMUTE2X2(Z14, Z15, Z26, K3)
	BUTTERFLYD2Q(Z14, Z15, Z24, Z26)
	MULD(Z15, Z18, Z28, Z29, Z26, Z27, Z30, Z31, Z24, Z25, K3)
	PERMUTE1X1(Z0, Z1, Z26, K3)
	BUTTERFLYD1Q(Z0, Z1, Z24, Z26, Z27)
	VPUNPCKLDQ   Z1, Z0, Z26
	VPUNPCKHDQ   Z1, Z0, Z1
	VMOVDQA32    Z26, Z0
	PERMUTE4X4(Z0, Z1, Z20, Z26, K2)
	PERMUTE8X8(Z0, Z1, Z26, K1)
	VMOVDQU32    Z0, 0(R15)
	VMOVDQU32    Z1, 64(R15)
	PERMUTE1X1(Z2, Z3, Z26, K3)
	BUTTERFLYD1Q(Z2, Z3, Z24, Z26, Z27)
	VPUNPCKLDQ   Z3, Z2, Z26
	VPUNPCKHDQ   Z3, Z2, Z3
	VMOVDQA32    Z26, Z2
	PERMUTE4X4(Z2, Z3, Z20, Z26, K2)
	PERMUTE8X8(Z2, Z3, Z26, K1)
	VMOVDQU32    Z2, 128(R15)
	VMOVDQU32    Z3, 192(R15)
	PERMUTE1X1(Z4, Z5, Z26, K3)
	BUTTERFLYD1Q(Z4, Z5, Z24, Z26, Z27)
	VPUNPCKLDQ   Z5, Z4, Z26
	VPUNPCKHDQ   Z5, Z4, Z5
	VMOVDQA32    Z26, Z4
	PERMUTE4X4(Z4, Z5, Z20, Z26, K2)
	PERMUTE8X8(Z4, Z5, Z26, K1)
	VMOVDQU32    Z4, 256(R15)
	VMOVDQU32    Z5, 320(R15)
	PERMUTE1X1(Z6, Z7, Z26, K3)
	BUTTERFLYD1Q(Z6, Z7, Z24, Z26, Z27)
	VPUNPCKLDQ   Z7, Z6, Z26
	VPUNPCKHDQ   Z7, Z6, Z7
	VMOVDQA32    Z26, Z6
	PERMUTE4X4(Z6, Z7, Z20, Z26, K2)
	PERMUTE8X8(Z6, Z7, Z26, K1)
	VMOVDQU32    Z6, 384(R15)
	VMOVDQU32    Z7, 448(R15)
	PERMUTE1X1(Z8, Z9, Z26, K3)
	BUTTERFLYD1Q(Z8, Z9, Z24, Z26, Z27)
	VPUNPCKLDQ   Z9, Z8, Z26
	VPUNPCKHDQ   Z9, Z8, Z9
	VMOVDQA32    Z26, Z8
	PERMUTE4X4(Z8, Z9, Z20, Z26, K2)
	PERMUTE8X8(Z8, Z9, Z26, K1)
	VMOVDQU32    Z8, 512(R15)
	VMOVDQU32    Z9, 576(R15)
	PERMUTE1X1(Z10, Z11, Z26, K3)
	BUTTERFLYD1Q(Z10, Z11, Z24, Z26, Z27)
	VPUNPCKLDQ   Z11, Z10, Z26
	VPUNPCKHDQ   Z11, Z10, Z11
	VMOVDQA32    Z26, Z10
	PERMUTE4X4(Z10, Z11, Z20, Z26, K2)
	PERMUTE8X8(Z10, Z11, Z26, K1)
	VMOVDQU32    Z10, 640(R15)
	VMOVDQU32    Z11, 704(R15)
	PERMUTE1X1(Z12, Z13, Z26, K3)
	BUTTERFLYD1Q(Z12, Z13, Z24, Z26, Z27)
	VPUNPCKLDQ   Z13, Z12, Z26
	VPUNPCKHDQ   Z13, Z12, Z13
	VMOVDQA32    Z26, Z12
	PERMUTE4X4(Z12, Z13, Z20, Z26, K2)
	PERMUTE8X8(Z12, Z13, Z26, K1)
	VMOVDQU32    Z12, 768(R15)
	VMOVDQU32    Z13, 832(R15)
	PERMUTE1X1(Z14, Z15, Z26, K3)
	BUTTERFLYD1Q(Z14, Z15, Z24, Z26, Z27)
	VPUNPCKLDQ   Z15, Z14, Z26
	VPUNPCKHDQ   Z15, Z14, Z15
	VMOVDQA32    Z26, Z14
	PERMUTE4X4(Z14, Z15, Z20, Z26, K2)
	PERMUTE8X8(Z14, Z15, Z26, K1)
	VMOVDQU32    Z14, 896(R15)
	VMOVDQU32    Z15, 960(R15)
	RET

TEXT ·SISToRefactor(SB), $2112-144
	MOVQ SP, DI
	ANDQ $-64, DI

	// prepare constants needed for mul and reduce ops
	VPCMPEQB     Y0, Y0, Y0
	VPMOVZXDQ    Y0, Z3
	MOVD         $const_q, AX
	VPBROADCASTQ AX, Z0
	VPBROADCASTD AX, Z1
	MOVD         $const_qInvNeg, AX
	VPBROADCASTQ AX, Z2
	MOVQ         k256+0(FP), R15
	MOVQ         k512+24(FP), DX
	MOVQ         cosets+48(FP), SI
	MOVQ         twiddles+72(FP), R14
	MOVQ         0(R14), R8              // twiddles[0]
	MOVQ         R15, CX
	MOVQ         DX, BX
	MOVQ         SI, R9
	ADDQ         $0x0000000000000200, CX
	ADDQ         $0x0000000000000400, BX
	ADDQ         $0x0000000000000400, R9

#define FROMMONTGOMERY(in0) \
	VPMULUDQ in0, Z2, Z5   \
	VPANDQ   Z3, Z5, Z5    \
	VPMULUDQ Z5, Z0, Z5    \
	VPADDQ   in0, Z5, in0  \
	VPSRLQ   $32, in0, in0 \
	VPSUBQ   Z0, in0, Z5   \
	VPMINUQ  in0, Z5, in0  \

#define LIMBSPLIT(in0) \
	VPSHUFLW $0x00000000000000dc, in0, in0 \
	VPSHUFHW $0x00000000000000dc, in0, in0 \

#define SPLITDWORDS(in0, in1, in2, in3) \
	VEXTRACTI32X8 $1, in0, in3 \
	VPMOVZXDQ     in3, in2     \
	VPMOVZXDQ     in1, in0     \

	MOVD         0(SI), AX
	VPBROADCASTQ AX, Z12
	VPMOVZXDQ    0(R15), Z6
	FROMMONTGOMERY(Z6)
	LIMBSPLIT(Z6)
	VPMOVZXDQ    0(SI), Z8
	VPMOVZXDQ    32(SI), Z9
	SPLITDWORDS(Z6, Y6, Z7, Y7)
	MULQ(Z6, Z8, Z3, Z0, Z2, Z4, Z5)
	MULQ(Z7, Z9, Z3, Z0, Z2, Z4, Z5)
	VPMOVQD      Z6, 0(DX)
	VPMOVQD      Z7, 32(DX)
	VPMOVZXDQ    0(CX), Z10
	FROMMONTGOMERY(Z10)
	LIMBSPLIT(Z10)
	VPMOVZXDQ    0(R9), Z8
	VPMOVZXDQ    32(R9), Z9
	SPLITDWORDS(Z10, Y10, Z11, Y11)
	MULQ(Z10, Z8, Z3, Z0, Z2, Z4, Z5)
	MULQ(Z11, Z9, Z3, Z0, Z2, Z4, Z5)
	VPMOVQD      Z10, 0(BX)
	VPMOVQD      Z11, 32(BX)
	VPMOVZXDQ    32(R15), Z6
	FROMMONTGOMERY(Z6)
	LIMBSPLIT(Z6)
	VPMOVZXDQ    64(SI), Z8
	VPMOVZXDQ    96(SI), Z9
	SPLITDWORDS(Z6, Y6, Z7, Y7)
	MULQ(Z6, Z8, Z3, Z0, Z2, Z4, Z5)
	MULQ(Z7, Z9, Z3, Z0, Z2, Z4, Z5)
	VPMOVQD      Z6, 64(DX)
	VPMOVQD      Z7, 96(DX)
	VPMOVZXDQ    32(CX), Z10
	FROMMONTGOMERY(Z10)
	LIMBSPLIT(Z10)
	VPMOVZXDQ    64(R9), Z8
	VPMOVZXDQ    96(R9), Z9
	SPLITDWORDS(Z10, Y10, Z11, Y11)
	MULQ(Z10, Z8, Z3, Z0, Z2, Z4, Z5)
	MULQ(Z11, Z9, Z3, Z0, Z2, Z4, Z5)
	VPMOVQD      Z10, 64(BX)
	VPMOVQD      Z11, 96(BX)
	VPMOVZXDQ    64(R15), Z6
	FROMMONTGOMERY(Z6)
	LIMBSPLIT(Z6)
	VPMOVZXDQ    128(SI), Z8
	VPMOVZXDQ    160(SI), Z9
	SPLITDWORDS(Z6, Y6, Z7, Y7)
	MULQ(Z6, Z8, Z3, Z0, Z2, Z4, Z5)
	MULQ(Z7, Z9, Z3, Z0, Z2, Z4, Z5)
	VPMOVQD      Z6, 128(DX)
	VPMOVQD      Z7, 160(DX)
	VPMOVZXDQ    64(CX), Z10
	FROMMONTGOMERY(Z10)
	LIMBSPLIT(Z10)
	VPMOVZXDQ    128(R9), Z8
	VPMOVZXDQ    160(R9), Z9
	SPLITDWORDS(Z10, Y10, Z11, Y11)
	MULQ(Z10, Z8, Z3, Z0, Z2, Z4, Z5)
	MULQ(Z11, Z9, Z3, Z0, Z2, Z4, Z5)
	VPMOVQD      Z10, 128(BX)
	VPMOVQD      Z11, 160(BX)
	VPMOVZXDQ    96(R15), Z6
	FROMMONTGOMERY(Z6)
	LIMBSPLIT(Z6)
	VPMOVZXDQ    192(SI), Z8
	VPMOVZXDQ    224(SI), Z9
	SPLITDWORDS(Z6, Y6, Z7, Y7)
	MULQ(Z6, Z8, Z3, Z0, Z2, Z4, Z5)
	MULQ(Z7, Z9, Z3, Z0, Z2, Z4, Z5)
	VPMOVQD      Z6, 192(DX)
	VPMOVQD      Z7, 224(DX)
	VPMOVZXDQ    96(CX), Z10
	FROMMONTGOMERY(Z10)
	LIMBSPLIT(Z10)
	VPMOVZXDQ    192(R9), Z8
	VPMOVZXDQ    224(R9), Z9
	SPLITDWORDS(Z10, Y10, Z11, Y11)
	MULQ(Z10, Z8, Z3, Z0, Z2, Z4, Z5)
	MULQ(Z11, Z9, Z3, Z0, Z2, Z4, Z5)
	VPMOVQD      Z10, 192(BX)
	VPMOVQD      Z11, 224(BX)
	VPMOVZXDQ    128(R15), Z6
	FROMMONTGOMERY(Z6)
	LIMBSPLIT(Z6)
	VPMOVZXDQ    256(SI), Z8
	VPMOVZXDQ    288(SI), Z9
	SPLITDWORDS(Z6, Y6, Z7, Y7)
	MULQ(Z6, Z8, Z3, Z0, Z2, Z4, Z5)
	MULQ(Z7, Z9, Z3, Z0, Z2, Z4, Z5)
	VPMOVQD      Z6, 256(DX)
	VPMOVQD      Z7, 288(DX)
	VPMOVZXDQ    128(CX), Z10
	FROMMONTGOMERY(Z10)
	LIMBSPLIT(Z10)
	VPMOVZXDQ    256(R9), Z8
	VPMOVZXDQ    288(R9), Z9
	SPLITDWORDS(Z10, Y10, Z11, Y11)
	MULQ(Z10, Z8, Z3, Z0, Z2, Z4, Z5)
	MULQ(Z11, Z9, Z3, Z0, Z2, Z4, Z5)
	VPMOVQD      Z10, 256(BX)
	VPMOVQD      Z11, 288(BX)
	VPMOVZXDQ    160(R15), Z6
	FROMMONTGOMERY(Z6)
	LIMBSPLIT(Z6)
	VPMOVZXDQ    320(SI), Z8
	VPMOVZXDQ    352(SI), Z9
	SPLITDWORDS(Z6, Y6, Z7, Y7)
	MULQ(Z6, Z8, Z3, Z0, Z2, Z4, Z5)
	MULQ(Z7, Z9, Z3, Z0, Z2, Z4, Z5)
	VPMOVQD      Z6, 320(DX)
	VPMOVQD      Z7, 352(DX)
	VPMOVZXDQ    160(CX), Z10
	FROMMONTGOMERY(Z10)
	LIMBSPLIT(Z10)
	VPMOVZXDQ    320(R9), Z8
	VPMOVZXDQ    352(R9), Z9
	SPLITDWORDS(Z10, Y10, Z11, Y11)
	MULQ(Z10, Z8, Z3, Z0, Z2, Z4, Z5)
	MULQ(Z11, Z9, Z3, Z0, Z2, Z4, Z5)
	VPMOVQD      Z10, 320(BX)
	VPMOVQD      Z11, 352(BX)
	VPMOVZXDQ    192(R15), Z6
	FROMMONTGOMERY(Z6)
	LIMBSPLIT(Z6)
	VPMOVZXDQ    384(SI), Z8
	VPMOVZXDQ    416(SI), Z9
	SPLITDWORDS(Z6, Y6, Z7, Y7)
	MULQ(Z6, Z8, Z3, Z0, Z2, Z4, Z5)
	MULQ(Z7, Z9, Z3, Z0, Z2, Z4, Z5)
	VPMOVQD      Z6, 384(DX)
	VPMOVQD      Z7, 416(DX)
	VPMOVZXDQ    192(CX), Z10
	FROMMONTGOMERY(Z10)
	LIMBSPLIT(Z10)
	VPMOVZXDQ    384(R9), Z8
	VPMOVZXDQ    416(R9), Z9
	SPLITDWORDS(Z10, Y10, Z11, Y11)
	MULQ(Z10, Z8, Z3, Z0, Z2, Z4, Z5)
	MULQ(Z11, Z9, Z3, Z0, Z2, Z4, Z5)
	VPMOVQD      Z10, 384(BX)
	VPMOVQD      Z11, 416(BX)
	VPMOVZXDQ    224(R15), Z6
	FROMMONTGOMERY(Z6)
	LIMBSPLIT(Z6)
	VPMOVZXDQ    448(SI), Z8
	VPMOVZXDQ    480(SI), Z9
	SPLITDWORDS(Z6, Y6, Z7, Y7)
	MULQ(Z6, Z8, Z3, Z0, Z2, Z4, Z5)
	MULQ(Z7, Z9, Z3, Z0, Z2, Z4, Z5)
	VPMOVQD      Z6, 448(DX)
	VPMOVQD      Z7, 480(DX)
	VPMOVZXDQ    224(CX), Z10
	FROMMONTGOMERY(Z10)
	LIMBSPLIT(Z10)
	VPMOVZXDQ    448(R9), Z8
	VPMOVZXDQ    480(R9), Z9
	SPLITDWORDS(Z10, Y10, Z11, Y11)
	MULQ(Z10, Z8, Z3, Z0, Z2, Z4, Z5)
	MULQ(Z11, Z9, Z3, Z0, Z2, Z4, Z5)
	VPMOVQD      Z10, 448(BX)
	VPMOVQD      Z11, 480(BX)
	VPMOVZXDQ    256(R15), Z6
	FROMMONTGOMERY(Z6)
	LIMBSPLIT(Z6)
	VPMOVZXDQ    512(SI), Z8
	VPMOVZXDQ    544(SI), Z9
	SPLITDWORDS(Z6, Y6, Z7, Y7)
	MULQ(Z6, Z8, Z3, Z0, Z2, Z4, Z5)
	MULQ(Z7, Z9, Z3, Z0, Z2, Z4, Z5)
	VPMOVQD      Z6, 512(DX)
	VPMOVQD      Z7, 544(DX)
	VPMOVZXDQ    256(CX), Z10
	FROMMONTGOMERY(Z10)
	LIMBSPLIT(Z10)
	VPMOVZXDQ    512(R9), Z8
	VPMOVZXDQ    544(R9), Z9
	SPLITDWORDS(Z10, Y10, Z11, Y11)
	MULQ(Z10, Z8, Z3, Z0, Z2, Z4, Z5)
	MULQ(Z11, Z9, Z3, Z0, Z2, Z4, Z5)
	VPMOVQD      Z10, 512(BX)
	VPMOVQD      Z11, 544(BX)
	VPMOVZXDQ    288(R15), Z6
	FROMMONTGOMERY(Z6)
	LIMBSPLIT(Z6)
	VPMOVZXDQ    576(SI), Z8
	VPMOVZXDQ    608(SI), Z9
	SPLITDWORDS(Z6, Y6, Z7, Y7)
	MULQ(Z6, Z8, Z3, Z0, Z2, Z4, Z5)
	MULQ(Z7, Z9, Z3, Z0, Z2, Z4, Z5)
	VPMOVQD      Z6, 576(DX)
	VPMOVQD      Z7, 608(DX)
	VPMOVZXDQ    288(CX), Z10
	FROMMONTGOMERY(Z10)
	LIMBSPLIT(Z10)
	VPMOVZXDQ    576(R9), Z8
	VPMOVZXDQ    608(R9), Z9
	SPLITDWORDS(Z10, Y10, Z11, Y11)
	MULQ(Z10, Z8, Z3, Z0, Z2, Z4, Z5)
	MULQ(Z11, Z9, Z3, Z0, Z2, Z4, Z5)
	VPMOVQD      Z10, 576(BX)
	VPMOVQD      Z11, 608(BX)
	VPMOVZXDQ    320(R15), Z6
	FROMMONTGOMERY(Z6)
	LIMBSPLIT(Z6)
	VPMOVZXDQ    640(SI), Z8
	VPMOVZXDQ    672(SI), Z9
	SPLITDWORDS(Z6, Y6, Z7, Y7)
	MULQ(Z6, Z8, Z3, Z0, Z2, Z4, Z5)
	MULQ(Z7, Z9, Z3, Z0, Z2, Z4, Z5)
	VPMOVQD      Z6, 640(DX)
	VPMOVQD      Z7, 672(DX)
	VPMOVZXDQ    320(CX), Z10
	FROMMONTGOMERY(Z10)
	LIMBSPLIT(Z10)
	VPMOVZXDQ    640(R9), Z8
	VPMOVZXDQ    672(R9), Z9
	SPLITDWORDS(Z10, Y10, Z11, Y11)
	MULQ(Z10, Z8, Z3, Z0, Z2, Z4, Z5)
	MULQ(Z11, Z9, Z3, Z0, Z2, Z4, Z5)
	VPMOVQD      Z10, 640(BX)
	VPMOVQD      Z11, 672(BX)
	VPMOVZXDQ    352(R15), Z6
	FROMMONTGOMERY(Z6)
	LIMBSPLIT(Z6)
	VPMOVZXDQ    704(SI), Z8
	VPMOVZXDQ    736(SI), Z9
	SPLITDWORDS(Z6, Y6, Z7, Y7)
	MULQ(Z6, Z8, Z3, Z0, Z2, Z4, Z5)
	MULQ(Z7, Z9, Z3, Z0, Z2, Z4, Z5)
	VPMOVQD      Z6, 704(DX)
	VPMOVQD      Z7, 736(DX)
	VPMOVZXDQ    352(CX), Z10
	FROMMONTGOMERY(Z10)
	LIMBSPLIT(Z10)
	VPMOVZXDQ    704(R9), Z8
	VPMOVZXDQ    736(R9), Z9
	SPLITDWORDS(Z10, Y10, Z11, Y11)
	MULQ(Z10, Z8, Z3, Z0, Z2, Z4, Z5)
	MULQ(Z11, Z9, Z3, Z0, Z2, Z4, Z5)
	VPMOVQD      Z10, 704(BX)
	VPMOVQD      Z11, 736(BX)
	VPMOVZXDQ    384(R15), Z6
	FROMMONTGOMERY(Z6)
	LIMBSPLIT(Z6)
	VPMOVZXDQ    768(SI), Z8
	VPMOVZXDQ    800(SI), Z9
	SPLITDWORDS(Z6, Y6, Z7, Y7)
	MULQ(Z6, Z8, Z3, Z0, Z2, Z4, Z5)
	MULQ(Z7, Z9, Z3, Z0, Z2, Z4, Z5)
	VPMOVQD      Z6, 768(DX)
	VPMOVQD      Z7, 800(DX)
	VPMOVZXDQ    384(CX), Z10
	FROMMONTGOMERY(Z10)
	LIMBSPLIT(Z10)
	VPMOVZXDQ    768(R9), Z8
	VPMOVZXDQ    800(R9), Z9
	SPLITDWORDS(Z10, Y10, Z11, Y11)
	MULQ(Z10, Z8, Z3, Z0, Z2, Z4, Z5)
	MULQ(Z11, Z9, Z3, Z0, Z2, Z4, Z5)
	VPMOVQD      Z10, 768(BX)
	VPMOVQD      Z11, 800(BX)
	VPMOVZXDQ    416(R15), Z6
	FROMMONTGOMERY(Z6)
	LIMBSPLIT(Z6)
	VPMOVZXDQ    832(SI), Z8
	VPMOVZXDQ    864(SI), Z9
	SPLITDWORDS(Z6, Y6, Z7, Y7)
	MULQ(Z6, Z8, Z3, Z0, Z2, Z4, Z5)
	MULQ(Z7, Z9, Z3, Z0, Z2, Z4, Z5)
	VPMOVQD      Z6, 832(DX)
	VPMOVQD      Z7, 864(DX)
	VPMOVZXDQ    416(CX), Z10
	FROMMONTGOMERY(Z10)
	LIMBSPLIT(Z10)
	VPMOVZXDQ    832(R9), Z8
	VPMOVZXDQ    864(R9), Z9
	SPLITDWORDS(Z10, Y10, Z11, Y11)
	MULQ(Z10, Z8, Z3, Z0, Z2, Z4, Z5)
	MULQ(Z11, Z9, Z3, Z0, Z2, Z4, Z5)
	VPMOVQD      Z10, 832(BX)
	VPMOVQD      Z11, 864(BX)
	VPMOVZXDQ    448(R15), Z6
	FROMMONTGOMERY(Z6)
	LIMBSPLIT(Z6)
	VPMOVZXDQ    896(SI), Z8
	VPMOVZXDQ    928(SI), Z9
	SPLITDWORDS(Z6, Y6, Z7, Y7)
	MULQ(Z6, Z8, Z3, Z0, Z2, Z4, Z5)
	MULQ(Z7, Z9, Z3, Z0, Z2, Z4, Z5)
	VPMOVQD      Z6, 896(DX)
	VPMOVQD      Z7, 928(DX)
	VPMOVZXDQ    448(CX), Z10
	FROMMONTGOMERY(Z10)
	LIMBSPLIT(Z10)
	VPMOVZXDQ    896(R9), Z8
	VPMOVZXDQ    928(R9), Z9
	SPLITDWORDS(Z10, Y10, Z11, Y11)
	MULQ(Z10, Z8, Z3, Z0, Z2, Z4, Z5)
	MULQ(Z11, Z9, Z3, Z0, Z2, Z4, Z5)
	VPMOVQD      Z10, 896(BX)
	VPMOVQD      Z11, 928(BX)
	VPMOVZXDQ    480(R15), Z6
	FROMMONTGOMERY(Z6)
	LIMBSPLIT(Z6)
	VPMOVZXDQ    960(SI), Z8
	VPMOVZXDQ    992(SI), Z9
	SPLITDWORDS(Z6, Y6, Z7, Y7)
	MULQ(Z6, Z8, Z3, Z0, Z2, Z4, Z5)
	MULQ(Z7, Z9, Z3, Z0, Z2, Z4, Z5)
	VPMOVQD      Z6, 960(DX)
	VPMOVQD      Z7, 992(DX)
	VPMOVZXDQ    480(CX), Z10
	FROMMONTGOMERY(Z10)
	LIMBSPLIT(Z10)
	VPMOVZXDQ    960(R9), Z8
	VPMOVZXDQ    992(R9), Z9
	SPLITDWORDS(Z10, Y10, Z11, Y11)
	MULQ(Z10, Z8, Z3, Z0, Z2, Z4, Z5)
	MULQ(Z11, Z9, Z3, Z0, Z2, Z4, Z5)
	VPMOVQD      Z10, 960(BX)
	VPMOVQD      Z11, 992(BX)
	RET
