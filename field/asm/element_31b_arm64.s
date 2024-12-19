// Code generated by gnark-crypto/generator. DO NOT EDIT.
#include "textflag.h"
#include "funcdata.h"
#include "go_asm.h"

// addVec(res, a, b *Element, n uint64)
TEXT ·addVec(SB), NOFRAME|NOSPLIT, $0-32
	LDP res+0(FP), (R0, R1)
	LDP b+16(FP), (R2, R3)

#define a V0
#define b V1
#define t V2
#define q V3
	VMOVS $const_q, q
	VDUP  q.S[0], q.S4 // broadcast q into q

loop1:
	CBZ    R3, done2
	VLD1.P 16(R1), [a.S4]
	VLD1.P 16(R2), [b.S4]
	VADD   a.S4, b.S4, b.S4 // b = a + b
	VSUB   q.S4, b.S4, t.S4 // t = b - q
	VUMIN  t.S4, b.S4, b.S4 // b = min(t, b)
	VST1.P [b.S4], 16(R0)   // res = b
	SUB    $1, R3, R3
	JMP    loop1

done2:
#undef a
#undef b
#undef t
#undef q
	RET

// subVec(res, a, b *Element, n uint64)
TEXT ·subVec(SB), NOFRAME|NOSPLIT, $0-32
	LDP res+0(FP), (R0, R1)
	LDP b+16(FP), (R2, R3)

#define a V0
#define b V1
#define t V2
#define q V3
	VMOVS $const_q, q
	VDUP  q.S[0], q.S4 // broadcast q into q

loop3:
	CBZ    R3, done4
	VLD1.P 16(R1), [a.S4]
	VLD1.P 16(R2), [b.S4]
	VSUB   b.S4, a.S4, b.S4 // b = a - b
	VADD   b.S4, q.S4, t.S4 // t = b + q
	VUMIN  t.S4, b.S4, b.S4 // b = min(t, b)
	VST1.P [b.S4], 16(R0)   // res = b
	SUB    $1, R3, R3
	JMP    loop3

done4:
#undef a
#undef b
#undef q
#undef t
	RET

// sumVec(t *uint64, a *[]uint32, n uint64) res = sum(a[0...n])
TEXT ·sumVec(SB), NOFRAME|NOSPLIT, $0-24
	// zeroing accumulators
	VMOVQ $0, $0, V4
	VMOVQ $0, $0, V5
	VMOVQ $0, $0, V6
	VMOVQ $0, $0, V7
	LDP   t+0(FP), (R1, R0)
	MOVD  n+16(FP), R2

loop5:
	CBZ R2, done6

	// blockSize is 16 uint32; we load 4 vectors of 4 uint32 at a time
	// (4*4)*4 = 64 bytes ~= 1 cache line
	// since our values are 31 bits, we can add 2 by 2 these vectors
	// we are left with 2 vectors of 4x32 bits values
	// that we accumulate in 4*2*64bits accumulators
	// the caller will reduce mod q the accumulators.

	VLD2.P  32(R0), [V0.S4, V1.S4]
	VADD    V0.S4, V1.S4, V0.S4    // a1 += a2
	VLD2.P  32(R0), [V2.S4, V3.S4]
	VADD    V2.S4, V3.S4, V2.S4    // a3 += a4
	VUSHLL  $0, V0.S2, V1.D2       // convert low words to 64 bits
	VADD    V1.D2, V5.D2, V5.D2    // acc2 += a2
	VUSHLL2 $0, V0.S4, V0.D2       // convert high words to 64 bits
	VADD    V0.D2, V4.D2, V4.D2    // acc1 += a1
	VUSHLL  $0, V2.S2, V3.D2       // convert low words to 64 bits
	VADD    V3.D2, V7.D2, V7.D2    // acc4 += a4
	VUSHLL2 $0, V2.S4, V2.D2       // convert high words to 64 bits
	VADD    V2.D2, V6.D2, V6.D2    // acc3 += a3
	SUB     $1, R2, R2
	JMP     loop5

done6:
	VADD   V4.D2, V6.D2, V4.D2   // acc1 += acc3
	VADD   V5.D2, V7.D2, V5.D2   // acc2 += acc4
	VST2.P [V4.D2, V5.D2], 0(R1) // store acc1 and acc2
	RET

// mulVec(res, a, b *Element, n uint64)
TEXT ·mulVec(SB), NOFRAME|NOSPLIT, $0-32
	LDP   res+0(FP), (R0, R1)
	LDP   b+16(FP), (R2, R3)
	VMOVS $const_q, V0
	VDUP  V0.D[0], V0.D2               // broadcast q into V0
	VMOVQ $0xffffffff, $0xffffffff, V1

loop7:
	CBZ     R3, done8
	MOVWU.P 4(R1), R4
	MOVWU.P 4(R1), R5
	MOVWU.P 4(R2), R6
	MOVWU.P 4(R2), R7
	MUL     R4, R6, R8
	MUL     R5, R7, R9
	VMOV    R8, V2.D[0]
	VMOV    R9, V2.D[1]
	VSHL    $0x1f, V2.D2, V4.D2
	VSHL    $0x18, V2.D2, V5.D2
	MOVWU.P 4(R1), R10
	MOVWU.P 4(R1), R11
	VSUB    V5.D2, V4.D2, V4.D2
	VSUB    V2.D2, V4.D2, V3.D2
	MOVWU.P 4(R2), R12
	MOVWU.P 4(R2), R13
	VAND    V3.B16, V1.B16, V3.B16
	VSHL    $0x1f, V3.D2, V4.D2
	VSHL    $0x18, V3.D2, V5.D2
	VSUB    V5.D2, V4.D2, V4.D2
	VADD    V3.D2, V4.D2, V3.D2
	VADD    V3.D2, V2.D2, V3.D2
	VUSHR   $0x20, V3.D2, V3.D2
	VSUB    V0.D2, V3.D2, V4.D2    // t = q - m
	VUMIN   V4.S4, V3.S4, V3.S4    // m = min(t, m)
	VSHL    $0x20, V3.D2, V3.D2
	MUL     R10, R12, R14
	MUL     R11, R13, R15
	VMOV    R14, V6.D[0]
	VMOV    R15, V6.D[1]
	VSHL    $0x1f, V6.D2, V8.D2
	VSHL    $0x18, V6.D2, V9.D2
	VSUB    V9.D2, V8.D2, V8.D2
	VSUB    V6.D2, V8.D2, V7.D2
	VAND    V7.B16, V1.B16, V7.B16
	VSHL    $0x1f, V7.D2, V8.D2
	VSHL    $0x18, V7.D2, V9.D2
	VSUB    V9.D2, V8.D2, V8.D2
	VADD    V7.D2, V8.D2, V7.D2
	VADD    V7.D2, V6.D2, V7.D2
	VUSHR   $0x20, V7.D2, V7.D2
	VSUB    V0.D2, V7.D2, V8.D2    // t = q - m
	VUMIN   V8.S4, V7.S4, V7.S4    // m = min(t, m)
	VADD    V7.S4, V3.S4, V3.S4
	VST1.P  [V3.S4], 16(R0)        // res = b
	SUB     $1, R3, R3
	JMP     loop7

done8:
	RET
