#include "textflag.h"
#include "funcdata.h"
#include "go_asm.h"

#define MULD(in0, in1, in2, in3, in4, in5, in6, in7, in8, in9, in10) \
	VPSRLQ    $32, in0, in2  \
	VPSRLQ    $32, in1, in3  \
	VPMULUDQ  in0, in1, in4  \
	VPMULUDQ  in2, in3, in5  \
	VPMULUDQ  in4, in9, in6  \
	VPMULUDQ  in5, in9, in7  \
	VPMULUDQ  in6, in8, in6  \
	VPADDQ    in4, in6, in4  \
	VPMULUDQ  in7, in8, in7  \
	VPADDQ    in5, in7, in5  \
	VMOVSHDUP in4, K3, in5   \
	VPSUBD    in8, in5, in7  \
	VPMINUD   in5, in7, in10 \

TEXT ·validation(SB), NOSPLIT, $0-48
	MOVQ         $0x0000000000005555, AX
	KMOVD        AX, K3
	MOVD         $const_q, AX
	VPBROADCASTD AX, Z0
	MOVD         $const_qInvNeg, AX
	VPBROADCASTD AX, Z1
	MOVQ         input+0(FP), R15
	VMOVDQU32    0(R15), Z2
	VMOVDQU32    64(R15), Y3

	MOVQ  $0x0000000000000001, AX
	KMOVQ AX, K2

	MULD(Z2, Z2, Z10, Z11, Z4, Z5, Z12, Z13, Z0, Z1, Z6)
	MULD(Z2, Z6, Z10, Z11, Z4, Z5, Z12, Z13, Z0, Z1, Z6)

	// VPBLENDMD Z6, Z2, K2, Z2
	PEXTRD  $0, X6, DX
	VPINSRD $0, DX, X2, X2

	VMOVDQU32 Z2, 0(R15)
	RET
