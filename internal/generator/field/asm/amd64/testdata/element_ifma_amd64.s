    // AVX-512 IFMA Vector Multiplication for 4-word Fields
    // Generated for prototype testing - DO NOT USE IN PRODUCTION

#include "textflag.h"
#include "funcdata.h"
#include "go_asm.h"

    // Permutation index for IFMA transpose: [0, 2, 1, 3, 4, 6, 5, 7]
    // This swaps positions 1<->2 and 5<->6 to fix even/odd interleaving
    DATA ·permuteIdxIFMA<>+0(SB)/8, $0
    DATA ·permuteIdxIFMA<>+8(SB)/8, $2
    DATA ·permuteIdxIFMA<>+16(SB)/8, $1
    DATA ·permuteIdxIFMA<>+24(SB)/8, $3
    DATA ·permuteIdxIFMA<>+32(SB)/8, $4
    DATA ·permuteIdxIFMA<>+40(SB)/8, $6
    DATA ·permuteIdxIFMA<>+48(SB)/8, $5
    DATA ·permuteIdxIFMA<>+56(SB)/8, $7
    GLOBL ·permuteIdxIFMA<>(SB), RODATA|NOPTR, $64

    // 2q in radix-52: used for binary reduction after x16 correction
    DATA ·q2Radix52<>+0(SB)/8, $0x3000000000002           // 2q[0]
    DATA ·q2Radix52<>+8(SB)/8, $0xfda0000002142           // 2q[1]
    DATA ·q2Radix52<>+16(SB)/8, $0x86f6002b354ed          // 2q[2]
    DATA ·q2Radix52<>+24(SB)/8, $0x4aacc1689a3cb          // 2q[3]
    DATA ·q2Radix52<>+32(SB)/8, $0x02556cabd3459          // 2q[4]
    GLOBL ·q2Radix52<>(SB), RODATA|NOPTR, $40

    // 4q in radix-52
    DATA ·q4Radix52<>+0(SB)/8, $0x6000000000004           // 4q[0]
    DATA ·q4Radix52<>+8(SB)/8, $0xfb40000004284           // 4q[1]
    DATA ·q4Radix52<>+16(SB)/8, $0x0dec00566a9db          // 4q[2]
    DATA ·q4Radix52<>+24(SB)/8, $0x955982d134797          // 4q[3]
    DATA ·q4Radix52<>+32(SB)/8, $0x04aad957a68b2          // 4q[4]
    GLOBL ·q4Radix52<>(SB), RODATA|NOPTR, $40

    // 8q in radix-52
    DATA ·q8Radix52<>+0(SB)/8, $0xc000000000008           // 8q[0]
    DATA ·q8Radix52<>+8(SB)/8, $0xf680000008508           // 8q[1]
    DATA ·q8Radix52<>+16(SB)/8, $0x1bd800acd53b7          // 8q[2]
    DATA ·q8Radix52<>+24(SB)/8, $0x2ab305a268f2e          // 8q[3]
    DATA ·q8Radix52<>+32(SB)/8, $0x0955b2af4d165          // 8q[4]
    GLOBL ·q8Radix52<>(SB), RODATA|NOPTR, $40

    // 16q in radix-52
    DATA ·q16Radix52<>+0(SB)/8, $0x8000000000010          // 16q[0]
    DATA ·q16Radix52<>+8(SB)/8, $0xed00000010a11          // 16q[1]
    DATA ·q16Radix52<>+16(SB)/8, $0x37b00159aa76f         // 16q[2]
    DATA ·q16Radix52<>+24(SB)/8, $0x55660b44d1e5c         // 16q[3]
    DATA ·q16Radix52<>+32(SB)/8, $0x12ab655e9a2ca         // 16q[4]
    GLOBL ·q16Radix52<>(SB), RODATA|NOPTR, $40

    // mulVecIFMA(res, a, b *Element, n uint64)
    // Performs n multiplications using AVX-512 IFMA instructions
    // Processes 8 elements in parallel using radix-52 representation
TEXT ·mulVecIFMA(SB), NOSPLIT, $0-32
    MOVQ res+0(FP), R14
    MOVQ a+8(FP), R13
    MOVQ b+16(FP), CX
    MOVQ n+24(FP), BX
    // Load constants for radix-52 conversion and reduction
    MOVQ $0xFFFFFFFFFFFFF, R15                             // 52-bit mask in R15
    VPBROADCASTQ R15, Z31                                          // Z31 = mask52 for SIMD ops
    MOVQ $const_qInvNeg, AX
    ANDQ R15, AX                                           // keep low 52 bits using mask in R15
    VPBROADCASTQ AX, Z30                                           // Z30 = qInvNeg52
    // Load modulus in radix-52 form
    // q in radix-52: Z25=ql0, Z26=ql1, Z27=ql2, Z28=ql3, Z29=ql4
    // Load q0-q3 and convert to radix-52
    MOVQ $const_q0, R9
    MOVQ $const_q1, R10
    MOVQ $const_q2, R11
    MOVQ $const_q3, R12
    MOVQ R9, R8
    ANDQ R15, R8
    VPBROADCASTQ R8, Z25
    SHRQ $52, R9
    MOVQ R10, R8
    SHLQ $12, R8
    ORQ R9, R8
    ANDQ R15, R8
    VPBROADCASTQ R8, Z26
    SHRQ $40, R10
    MOVQ R11, R8
    SHLQ $24, R8
    ORQ R10, R8
    ANDQ R15, R8
    VPBROADCASTQ R8, Z27
    SHRQ $28, R11
    MOVQ R12, R8
    SHLQ $36, R8
    ORQ R11, R8
    ANDQ R15, R8
    VPBROADCASTQ R8, Z28
    SHRQ $16, R12
    VPBROADCASTQ R12, Z29
loop_1:
    TESTQ BX, BX
    JEQ done_2                                            // n == 0, we are done
    // Process 8 elements in parallel
    // Load and convert 8 elements from a[] to radix-52
    // Load 8 elements from R13
    // Load element words using gather pattern
    VMOVDQU64 0(R13), Z10
    VMOVDQU64 64(R13), Z11
    VMOVDQU64 128(R13), Z12
    VMOVDQU64 192(R13), Z13
    // Transpose 8 elements for vertical SIMD processing
    // 8x4 transpose using AVX-512 shuffles
    VPUNPCKLQDQ Z11, Z10, Z18                                     // [e0.a0, e2.a0, e0.a2, e2.a2, e1.a0, e3.a0, e1.a2, e3.a2]
    VPUNPCKHQDQ Z11, Z10, Z19                                     // [e0.a1, e2.a1, e0.a3, e2.a3, e1.a1, e3.a1, e1.a3, e3.a3]
    VPUNPCKLQDQ Z13, Z12, Z20                                     // [e4.a0, e6.a0, e4.a2, e6.a2, e5.a0, e7.a0, e5.a2, e7.a2]
    VPUNPCKHQDQ Z13, Z12, Z21                                     // [e4.a1, e6.a1, e4.a3, e6.a3, e5.a1, e7.a1, e5.a3, e7.a3]
    VSHUFI64X2 $0x88, Z20, Z18, Z14                              // a0: lanes 0,2 from Z18 and Z20
    VSHUFI64X2 $0xDD, Z20, Z18, Z16                              // a2: lanes 1,3 from Z18 and Z20
    VSHUFI64X2 $0x88, Z21, Z19, Z15                              // a1: lanes 0,2 from Z19 and Z21
    VSHUFI64X2 $0xDD, Z21, Z19, Z17                              // a3: lanes 1,3 from Z19 and Z21
    VMOVDQU64 ·permuteIdxIFMA<>(SB), Z22
    VPERMQ Z14, Z22, Z14
    VPERMQ Z15, Z22, Z15
    VPERMQ Z16, Z22, Z16
    VPERMQ Z17, Z22, Z17
    // Convert to radix-52
    VPANDQ Z31, Z14, Z0
    VPSRLQ $52, Z14, Z18
    VPSLLQ $12, Z15, Z19
    VPORQ Z18, Z19, Z18
    VPANDQ Z31, Z18, Z1
    VPSRLQ $40, Z15, Z18
    VPSLLQ $24, Z16, Z19
    VPORQ Z18, Z19, Z18
    VPANDQ Z31, Z18, Z2
    VPSRLQ $28, Z16, Z18
    VPSLLQ $36, Z17, Z19
    VPORQ Z18, Z19, Z18
    VPANDQ Z31, Z18, Z3
    VPSRLQ $16, Z17, Z4
    // Load and convert 8 elements from b[] to radix-52
    // Load 8 elements from CX
    // Load element words using gather pattern
    VMOVDQU64 0(CX), Z10
    VMOVDQU64 64(CX), Z11
    VMOVDQU64 128(CX), Z12
    VMOVDQU64 192(CX), Z13
    // Transpose 8 elements for vertical SIMD processing
    // 8x4 transpose using AVX-512 shuffles
    VPUNPCKLQDQ Z11, Z10, Z18                                     // [e0.a0, e2.a0, e0.a2, e2.a2, e1.a0, e3.a0, e1.a2, e3.a2]
    VPUNPCKHQDQ Z11, Z10, Z19                                     // [e0.a1, e2.a1, e0.a3, e2.a3, e1.a1, e3.a1, e1.a3, e3.a3]
    VPUNPCKLQDQ Z13, Z12, Z20                                     // [e4.a0, e6.a0, e4.a2, e6.a2, e5.a0, e7.a0, e5.a2, e7.a2]
    VPUNPCKHQDQ Z13, Z12, Z21                                     // [e4.a1, e6.a1, e4.a3, e6.a3, e5.a1, e7.a1, e5.a3, e7.a3]
    VSHUFI64X2 $0x88, Z20, Z18, Z14                              // a0: lanes 0,2 from Z18 and Z20
    VSHUFI64X2 $0xDD, Z20, Z18, Z16                              // a2: lanes 1,3 from Z18 and Z20
    VSHUFI64X2 $0x88, Z21, Z19, Z15                              // a1: lanes 0,2 from Z19 and Z21
    VSHUFI64X2 $0xDD, Z21, Z19, Z17                              // a3: lanes 1,3 from Z19 and Z21
    VMOVDQU64 ·permuteIdxIFMA<>(SB), Z22
    VPERMQ Z14, Z22, Z14
    VPERMQ Z15, Z22, Z15
    VPERMQ Z16, Z22, Z16
    VPERMQ Z17, Z22, Z17
    // Convert to radix-52
    VPANDQ Z31, Z14, Z5
    VPSRLQ $52, Z14, Z18
    VPSLLQ $12, Z15, Z19
    VPORQ Z18, Z19, Z18
    VPANDQ Z31, Z18, Z6
    VPSRLQ $40, Z15, Z18
    VPSLLQ $24, Z16, Z19
    VPORQ Z18, Z19, Z18
    VPANDQ Z31, Z18, Z7
    VPSRLQ $28, Z16, Z18
    VPSLLQ $36, Z17, Z19
    VPORQ Z18, Z19, Z18
    VPANDQ Z31, Z18, Z8
    VPSRLQ $16, Z17, Z9
    // Montgomery multiplication using IFMA (BPS method)
    // Montgomery multiplication using CIOS variant
    // A = [Z0, Z1, Z2, Z3, Z4], B = [Z5, Z6, Z7, Z8, Z9]
    VPXORQ Z10, Z10, Z10                                     // T0
    VPXORQ Z11, Z11, Z11                                     // T1
    VPXORQ Z12, Z12, Z12                                     // T2
    VPXORQ Z13, Z13, Z13                                     // T3
    VPXORQ Z14, Z14, Z14                                     // T4
    VPXORQ Z15, Z15, Z15                                     // T5 (overflow)
    // Round 0: process B[0]
    // T += A * B[i]
    VPMADD52LUQ Z5, Z0, Z10
    VPMADD52HUQ Z5, Z0, Z11
    VPMADD52LUQ Z5, Z1, Z11
    VPMADD52HUQ Z5, Z1, Z12
    VPMADD52LUQ Z5, Z2, Z12
    VPMADD52HUQ Z5, Z2, Z13
    VPMADD52LUQ Z5, Z3, Z13
    VPMADD52HUQ Z5, Z3, Z14
    VPMADD52LUQ Z5, Z4, Z14
    VPMADD52HUQ Z5, Z4, Z15
    // Normalize T[0]
    VPSRLQ $52, Z10, Z20                                     // carry = T[0] >> 52
    VPANDQ Z31, Z10, Z10                                     // T[0] &= mask52
    VPADDQ Z20, Z11, Z11                                     // T[1] += carry
    // m = T[0] * qInvNeg52 mod 2^52
    VPXORQ Z20, Z20, Z20                                     // clear Z20
    VPMADD52LUQ Z30, Z10, Z20                                     // Z20 = low52(T[0] * qInvNeg52)
    VPANDQ Z31, Z20, Z20                                     // mask to 52 bits (m in Z20)
    // T += m * q
    VPMADD52LUQ Z25, Z20, Z10
    VPMADD52HUQ Z25, Z20, Z11
    VPMADD52LUQ Z26, Z20, Z11
    VPMADD52HUQ Z26, Z20, Z12
    VPMADD52LUQ Z27, Z20, Z12
    VPMADD52HUQ Z27, Z20, Z13
    VPMADD52LUQ Z28, Z20, Z13
    VPMADD52HUQ Z28, Z20, Z14
    VPMADD52LUQ Z29, Z20, Z14
    VPMADD52HUQ Z29, Z20, Z15
    // Shift: T[j] = T[j+1]
    VPSRLQ $52, Z10, Z20                                     // carry from T[0] (should be the only content)
    VPADDQ Z20, Z11, Z10                                     // T[0] = T[1] + carry
    VMOVDQA64 Z12, Z11                                          // T[1] = T[2]
    VMOVDQA64 Z13, Z12                                          // T[2] = T[3]
    VMOVDQA64 Z14, Z13                                          // T[3] = T[4]
    VMOVDQA64 Z15, Z14                                          // T[4] = T[5]
    VPXORQ Z15, Z15, Z15                                     // T[5] = 0
    // Round 1: process B[1]
    // T += A * B[i]
    VPMADD52LUQ Z6, Z0, Z10
    VPMADD52HUQ Z6, Z0, Z11
    VPMADD52LUQ Z6, Z1, Z11
    VPMADD52HUQ Z6, Z1, Z12
    VPMADD52LUQ Z6, Z2, Z12
    VPMADD52HUQ Z6, Z2, Z13
    VPMADD52LUQ Z6, Z3, Z13
    VPMADD52HUQ Z6, Z3, Z14
    VPMADD52LUQ Z6, Z4, Z14
    VPMADD52HUQ Z6, Z4, Z15
    // Normalize T[0]
    VPSRLQ $52, Z10, Z20                                     // carry = T[0] >> 52
    VPANDQ Z31, Z10, Z10                                     // T[0] &= mask52
    VPADDQ Z20, Z11, Z11                                     // T[1] += carry
    // m = T[0] * qInvNeg52 mod 2^52
    VPXORQ Z20, Z20, Z20                                     // clear Z20
    VPMADD52LUQ Z30, Z10, Z20                                     // Z20 = low52(T[0] * qInvNeg52)
    VPANDQ Z31, Z20, Z20                                     // mask to 52 bits (m in Z20)
    // T += m * q
    VPMADD52LUQ Z25, Z20, Z10
    VPMADD52HUQ Z25, Z20, Z11
    VPMADD52LUQ Z26, Z20, Z11
    VPMADD52HUQ Z26, Z20, Z12
    VPMADD52LUQ Z27, Z20, Z12
    VPMADD52HUQ Z27, Z20, Z13
    VPMADD52LUQ Z28, Z20, Z13
    VPMADD52HUQ Z28, Z20, Z14
    VPMADD52LUQ Z29, Z20, Z14
    VPMADD52HUQ Z29, Z20, Z15
    // Shift: T[j] = T[j+1]
    VPSRLQ $52, Z10, Z20                                     // carry from T[0] (should be the only content)
    VPADDQ Z20, Z11, Z10                                     // T[0] = T[1] + carry
    VMOVDQA64 Z12, Z11                                          // T[1] = T[2]
    VMOVDQA64 Z13, Z12                                          // T[2] = T[3]
    VMOVDQA64 Z14, Z13                                          // T[3] = T[4]
    VMOVDQA64 Z15, Z14                                          // T[4] = T[5]
    VPXORQ Z15, Z15, Z15                                     // T[5] = 0
    // Round 2: process B[2]
    // T += A * B[i]
    VPMADD52LUQ Z7, Z0, Z10
    VPMADD52HUQ Z7, Z0, Z11
    VPMADD52LUQ Z7, Z1, Z11
    VPMADD52HUQ Z7, Z1, Z12
    VPMADD52LUQ Z7, Z2, Z12
    VPMADD52HUQ Z7, Z2, Z13
    VPMADD52LUQ Z7, Z3, Z13
    VPMADD52HUQ Z7, Z3, Z14
    VPMADD52LUQ Z7, Z4, Z14
    VPMADD52HUQ Z7, Z4, Z15
    // Normalize T[0]
    VPSRLQ $52, Z10, Z20                                     // carry = T[0] >> 52
    VPANDQ Z31, Z10, Z10                                     // T[0] &= mask52
    VPADDQ Z20, Z11, Z11                                     // T[1] += carry
    // m = T[0] * qInvNeg52 mod 2^52
    VPXORQ Z20, Z20, Z20                                     // clear Z20
    VPMADD52LUQ Z30, Z10, Z20                                     // Z20 = low52(T[0] * qInvNeg52)
    VPANDQ Z31, Z20, Z20                                     // mask to 52 bits (m in Z20)
    // T += m * q
    VPMADD52LUQ Z25, Z20, Z10
    VPMADD52HUQ Z25, Z20, Z11
    VPMADD52LUQ Z26, Z20, Z11
    VPMADD52HUQ Z26, Z20, Z12
    VPMADD52LUQ Z27, Z20, Z12
    VPMADD52HUQ Z27, Z20, Z13
    VPMADD52LUQ Z28, Z20, Z13
    VPMADD52HUQ Z28, Z20, Z14
    VPMADD52LUQ Z29, Z20, Z14
    VPMADD52HUQ Z29, Z20, Z15
    // Shift: T[j] = T[j+1]
    VPSRLQ $52, Z10, Z20                                     // carry from T[0] (should be the only content)
    VPADDQ Z20, Z11, Z10                                     // T[0] = T[1] + carry
    VMOVDQA64 Z12, Z11                                          // T[1] = T[2]
    VMOVDQA64 Z13, Z12                                          // T[2] = T[3]
    VMOVDQA64 Z14, Z13                                          // T[3] = T[4]
    VMOVDQA64 Z15, Z14                                          // T[4] = T[5]
    VPXORQ Z15, Z15, Z15                                     // T[5] = 0
    // Round 3: process B[3]
    // T += A * B[i]
    VPMADD52LUQ Z8, Z0, Z10
    VPMADD52HUQ Z8, Z0, Z11
    VPMADD52LUQ Z8, Z1, Z11
    VPMADD52HUQ Z8, Z1, Z12
    VPMADD52LUQ Z8, Z2, Z12
    VPMADD52HUQ Z8, Z2, Z13
    VPMADD52LUQ Z8, Z3, Z13
    VPMADD52HUQ Z8, Z3, Z14
    VPMADD52LUQ Z8, Z4, Z14
    VPMADD52HUQ Z8, Z4, Z15
    // Normalize T[0]
    VPSRLQ $52, Z10, Z20                                     // carry = T[0] >> 52
    VPANDQ Z31, Z10, Z10                                     // T[0] &= mask52
    VPADDQ Z20, Z11, Z11                                     // T[1] += carry
    // m = T[0] * qInvNeg52 mod 2^52
    VPXORQ Z20, Z20, Z20                                     // clear Z20
    VPMADD52LUQ Z30, Z10, Z20                                     // Z20 = low52(T[0] * qInvNeg52)
    VPANDQ Z31, Z20, Z20                                     // mask to 52 bits (m in Z20)
    // T += m * q
    VPMADD52LUQ Z25, Z20, Z10
    VPMADD52HUQ Z25, Z20, Z11
    VPMADD52LUQ Z26, Z20, Z11
    VPMADD52HUQ Z26, Z20, Z12
    VPMADD52LUQ Z27, Z20, Z12
    VPMADD52HUQ Z27, Z20, Z13
    VPMADD52LUQ Z28, Z20, Z13
    VPMADD52HUQ Z28, Z20, Z14
    VPMADD52LUQ Z29, Z20, Z14
    VPMADD52HUQ Z29, Z20, Z15
    // Shift: T[j] = T[j+1]
    VPSRLQ $52, Z10, Z20                                     // carry from T[0] (should be the only content)
    VPADDQ Z20, Z11, Z10                                     // T[0] = T[1] + carry
    VMOVDQA64 Z12, Z11                                          // T[1] = T[2]
    VMOVDQA64 Z13, Z12                                          // T[2] = T[3]
    VMOVDQA64 Z14, Z13                                          // T[3] = T[4]
    VMOVDQA64 Z15, Z14                                          // T[4] = T[5]
    VPXORQ Z15, Z15, Z15                                     // T[5] = 0
    // Round 4: process B[4]
    // T += A * B[i]
    VPMADD52LUQ Z9, Z0, Z10
    VPMADD52HUQ Z9, Z0, Z11
    VPMADD52LUQ Z9, Z1, Z11
    VPMADD52HUQ Z9, Z1, Z12
    VPMADD52LUQ Z9, Z2, Z12
    VPMADD52HUQ Z9, Z2, Z13
    VPMADD52LUQ Z9, Z3, Z13
    VPMADD52HUQ Z9, Z3, Z14
    VPMADD52LUQ Z9, Z4, Z14
    VPMADD52HUQ Z9, Z4, Z15
    // Normalize T[0]
    VPSRLQ $52, Z10, Z20                                     // carry = T[0] >> 52
    VPANDQ Z31, Z10, Z10                                     // T[0] &= mask52
    VPADDQ Z20, Z11, Z11                                     // T[1] += carry
    // m = T[0] * qInvNeg52 mod 2^52
    VPXORQ Z20, Z20, Z20                                     // clear Z20
    VPMADD52LUQ Z30, Z10, Z20                                     // Z20 = low52(T[0] * qInvNeg52)
    VPANDQ Z31, Z20, Z20                                     // mask to 52 bits (m in Z20)
    // T += m * q
    VPMADD52LUQ Z25, Z20, Z10
    VPMADD52HUQ Z25, Z20, Z11
    VPMADD52LUQ Z26, Z20, Z11
    VPMADD52HUQ Z26, Z20, Z12
    VPMADD52LUQ Z27, Z20, Z12
    VPMADD52HUQ Z27, Z20, Z13
    VPMADD52LUQ Z28, Z20, Z13
    VPMADD52HUQ Z28, Z20, Z14
    VPMADD52LUQ Z29, Z20, Z14
    VPMADD52HUQ Z29, Z20, Z15
    // Shift: T[j] = T[j+1]
    VPSRLQ $52, Z10, Z20                                     // carry from T[0] (should be the only content)
    VPADDQ Z20, Z11, Z10                                     // T[0] = T[1] + carry
    VMOVDQA64 Z12, Z11                                          // T[1] = T[2]
    VMOVDQA64 Z13, Z12                                          // T[2] = T[3]
    VMOVDQA64 Z14, Z13                                          // T[3] = T[4]
    VMOVDQA64 Z15, Z14                                          // T[4] = T[5]
    VPXORQ Z15, Z15, Z15                                     // T[5] = 0
    // Copy result to Z0-Z4
    VMOVDQA64 Z10, Z0
    VMOVDQA64 Z11, Z1
    VMOVDQA64 Z12, Z2
    VMOVDQA64 Z13, Z3
    VMOVDQA64 Z14, Z4
    // Final normalization
    VPSRLQ $52, Z0, Z20
    VPANDQ Z31, Z0, Z0
    VPADDQ Z20, Z1, Z1
    VPSRLQ $52, Z1, Z20
    VPANDQ Z31, Z1, Z1
    VPADDQ Z20, Z2, Z2
    VPSRLQ $52, Z2, Z20
    VPANDQ Z31, Z2, Z2
    VPADDQ Z20, Z3, Z3
    VPSRLQ $52, Z3, Z20
    VPANDQ Z31, Z3, Z3
    VPADDQ Z20, Z4, Z4
    // Conditional subtraction if >= q
    VPSUBQ Z25, Z0, Z10
    VPSUBQ Z26, Z1, Z11
    VPSUBQ Z27, Z2, Z12
    VPSUBQ Z28, Z3, Z13
    VPSUBQ Z29, Z4, Z14
    VPSRAQ $63, Z10, Z20                                     // Z20 = -1 if borrow, 0 otherwise
    VPADDQ Z20, Z11, Z11                                     // Z11 -= borrow
    VPSRAQ $63, Z11, Z20
    VPADDQ Z20, Z12, Z12
    VPSRAQ $63, Z12, Z20
    VPADDQ Z20, Z13, Z13
    VPSRAQ $63, Z13, Z20
    VPADDQ Z20, Z14, Z14
    VPSRAQ $63, Z14, Z20                                     // Z20 = all 1s if borrow (result < q), all 0s if no borrow
    VPANDQ Z31, Z10, Z10
    VPANDQ Z31, Z11, Z11
    VPANDQ Z31, Z12, Z12
    VPANDQ Z31, Z13, Z13
    VPANDQ Z31, Z14, Z14
    VPANDQ Z20, Z0, Z0                                       // keep original if borrow
    VPANDNQ Z10, Z20, Z10                                     // keep subtracted if no borrow
    VPORQ Z10, Z0, Z0
    VPANDQ Z20, Z1, Z1
    VPANDNQ Z11, Z20, Z11
    VPORQ Z11, Z1, Z1
    VPANDQ Z20, Z2, Z2
    VPANDNQ Z12, Z20, Z12
    VPORQ Z12, Z2, Z2
    VPANDQ Z20, Z3, Z3
    VPANDNQ Z13, Z20, Z13
    VPORQ Z13, Z3, Z3
    VPANDQ Z20, Z4, Z4
    VPANDNQ Z14, Z20, Z14
    VPORQ Z14, Z4, Z4
    // Multiply by 16 in radix-52 to correct for radix-260 vs radix-256
    // Multiply by 16 = 2^4 (left shift with carry) in radix-52
    VPSLLQ $4, Z0, Z10                                       // Z10 = l0 << 4
    VPANDQ Z31, Z10, Z0                                      // Z0 = (l0 << 4) & mask52
    VPSRLQ $52, Z10, Z15                                     // Z15 = carry = (l0 << 4) >> 52
    VPSLLQ $4, Z1, Z10                                       // Z10 = l1 << 4
    VPORQ Z15, Z10, Z10                                     // Z10 = (l1 << 4) | carry (no overlap)
    VPANDQ Z31, Z10, Z1                                      // Z1 = result & mask52
    VPSRLQ $52, Z10, Z15                                     // Z15 = new carry
    VPSLLQ $4, Z2, Z10
    VPORQ Z15, Z10, Z10
    VPANDQ Z31, Z10, Z2
    VPSRLQ $52, Z10, Z15
    VPSLLQ $4, Z3, Z10
    VPORQ Z15, Z10, Z10
    VPANDQ Z31, Z10, Z3
    VPSRLQ $52, Z10, Z15
    VPSLLQ $4, Z4, Z10
    VPORQ Z15, Z10, Z4                                      // Z4 = (l4 << 4) | carry
    // Binary reduction: conditionally subtract 16q, 8q, 4q, 2q, q
    VPBROADCASTQ ·q16Radix52<>+0(SB), Z5
    VPBROADCASTQ ·q16Radix52<>+8(SB), Z6
    VPBROADCASTQ ·q16Radix52<>+16(SB), Z7
    VPBROADCASTQ ·q16Radix52<>+24(SB), Z8
    VPBROADCASTQ ·q16Radix52<>+32(SB), Z9
    VPSUBQ Z5, Z0, Z10
    VPSUBQ Z6, Z1, Z11
    VPSUBQ Z7, Z2, Z12
    VPSUBQ Z8, Z3, Z13
    VPSUBQ Z9, Z4, Z14
    VPSRAQ $63, Z10, Z20
    VPADDQ Z20, Z11, Z11
    VPSRAQ $63, Z11, Z20
    VPADDQ Z20, Z12, Z12
    VPSRAQ $63, Z12, Z20
    VPADDQ Z20, Z13, Z13
    VPSRAQ $63, Z13, Z20
    VPADDQ Z20, Z14, Z14
    VPSRAQ $63, Z14, Z20
    VPANDQ Z31, Z10, Z10
    VPANDQ Z31, Z11, Z11
    VPANDQ Z31, Z12, Z12
    VPANDQ Z31, Z13, Z13
    VPANDQ Z31, Z14, Z14
    VPANDQ Z20, Z0, Z0
    VPANDNQ Z10, Z20, Z10
    VPORQ Z10, Z0, Z0
    VPANDQ Z20, Z1, Z1
    VPANDNQ Z11, Z20, Z11
    VPORQ Z11, Z1, Z1
    VPANDQ Z20, Z2, Z2
    VPANDNQ Z12, Z20, Z12
    VPORQ Z12, Z2, Z2
    VPANDQ Z20, Z3, Z3
    VPANDNQ Z13, Z20, Z13
    VPORQ Z13, Z3, Z3
    VPANDQ Z20, Z4, Z4
    VPANDNQ Z14, Z20, Z14
    VPORQ Z14, Z4, Z4
    VPBROADCASTQ ·q8Radix52<>+0(SB), Z5
    VPBROADCASTQ ·q8Radix52<>+8(SB), Z6
    VPBROADCASTQ ·q8Radix52<>+16(SB), Z7
    VPBROADCASTQ ·q8Radix52<>+24(SB), Z8
    VPBROADCASTQ ·q8Radix52<>+32(SB), Z9
    VPSUBQ Z5, Z0, Z10
    VPSUBQ Z6, Z1, Z11
    VPSUBQ Z7, Z2, Z12
    VPSUBQ Z8, Z3, Z13
    VPSUBQ Z9, Z4, Z14
    VPSRAQ $63, Z10, Z20
    VPADDQ Z20, Z11, Z11
    VPSRAQ $63, Z11, Z20
    VPADDQ Z20, Z12, Z12
    VPSRAQ $63, Z12, Z20
    VPADDQ Z20, Z13, Z13
    VPSRAQ $63, Z13, Z20
    VPADDQ Z20, Z14, Z14
    VPSRAQ $63, Z14, Z20
    VPANDQ Z31, Z10, Z10
    VPANDQ Z31, Z11, Z11
    VPANDQ Z31, Z12, Z12
    VPANDQ Z31, Z13, Z13
    VPANDQ Z31, Z14, Z14
    VPANDQ Z20, Z0, Z0
    VPANDNQ Z10, Z20, Z10
    VPORQ Z10, Z0, Z0
    VPANDQ Z20, Z1, Z1
    VPANDNQ Z11, Z20, Z11
    VPORQ Z11, Z1, Z1
    VPANDQ Z20, Z2, Z2
    VPANDNQ Z12, Z20, Z12
    VPORQ Z12, Z2, Z2
    VPANDQ Z20, Z3, Z3
    VPANDNQ Z13, Z20, Z13
    VPORQ Z13, Z3, Z3
    VPANDQ Z20, Z4, Z4
    VPANDNQ Z14, Z20, Z14
    VPORQ Z14, Z4, Z4
    VPBROADCASTQ ·q4Radix52<>+0(SB), Z5
    VPBROADCASTQ ·q4Radix52<>+8(SB), Z6
    VPBROADCASTQ ·q4Radix52<>+16(SB), Z7
    VPBROADCASTQ ·q4Radix52<>+24(SB), Z8
    VPBROADCASTQ ·q4Radix52<>+32(SB), Z9
    VPSUBQ Z5, Z0, Z10
    VPSUBQ Z6, Z1, Z11
    VPSUBQ Z7, Z2, Z12
    VPSUBQ Z8, Z3, Z13
    VPSUBQ Z9, Z4, Z14
    VPSRAQ $63, Z10, Z20
    VPADDQ Z20, Z11, Z11
    VPSRAQ $63, Z11, Z20
    VPADDQ Z20, Z12, Z12
    VPSRAQ $63, Z12, Z20
    VPADDQ Z20, Z13, Z13
    VPSRAQ $63, Z13, Z20
    VPADDQ Z20, Z14, Z14
    VPSRAQ $63, Z14, Z20
    VPANDQ Z31, Z10, Z10
    VPANDQ Z31, Z11, Z11
    VPANDQ Z31, Z12, Z12
    VPANDQ Z31, Z13, Z13
    VPANDQ Z31, Z14, Z14
    VPANDQ Z20, Z0, Z0
    VPANDNQ Z10, Z20, Z10
    VPORQ Z10, Z0, Z0
    VPANDQ Z20, Z1, Z1
    VPANDNQ Z11, Z20, Z11
    VPORQ Z11, Z1, Z1
    VPANDQ Z20, Z2, Z2
    VPANDNQ Z12, Z20, Z12
    VPORQ Z12, Z2, Z2
    VPANDQ Z20, Z3, Z3
    VPANDNQ Z13, Z20, Z13
    VPORQ Z13, Z3, Z3
    VPANDQ Z20, Z4, Z4
    VPANDNQ Z14, Z20, Z14
    VPORQ Z14, Z4, Z4
    VPBROADCASTQ ·q2Radix52<>+0(SB), Z5
    VPBROADCASTQ ·q2Radix52<>+8(SB), Z6
    VPBROADCASTQ ·q2Radix52<>+16(SB), Z7
    VPBROADCASTQ ·q2Radix52<>+24(SB), Z8
    VPBROADCASTQ ·q2Radix52<>+32(SB), Z9
    VPSUBQ Z5, Z0, Z10
    VPSUBQ Z6, Z1, Z11
    VPSUBQ Z7, Z2, Z12
    VPSUBQ Z8, Z3, Z13
    VPSUBQ Z9, Z4, Z14
    VPSRAQ $63, Z10, Z20
    VPADDQ Z20, Z11, Z11
    VPSRAQ $63, Z11, Z20
    VPADDQ Z20, Z12, Z12
    VPSRAQ $63, Z12, Z20
    VPADDQ Z20, Z13, Z13
    VPSRAQ $63, Z13, Z20
    VPADDQ Z20, Z14, Z14
    VPSRAQ $63, Z14, Z20
    VPANDQ Z31, Z10, Z10
    VPANDQ Z31, Z11, Z11
    VPANDQ Z31, Z12, Z12
    VPANDQ Z31, Z13, Z13
    VPANDQ Z31, Z14, Z14
    VPANDQ Z20, Z0, Z0
    VPANDNQ Z10, Z20, Z10
    VPORQ Z10, Z0, Z0
    VPANDQ Z20, Z1, Z1
    VPANDNQ Z11, Z20, Z11
    VPORQ Z11, Z1, Z1
    VPANDQ Z20, Z2, Z2
    VPANDNQ Z12, Z20, Z12
    VPORQ Z12, Z2, Z2
    VPANDQ Z20, Z3, Z3
    VPANDNQ Z13, Z20, Z13
    VPORQ Z13, Z3, Z3
    VPANDQ Z20, Z4, Z4
    VPANDNQ Z14, Z20, Z14
    VPORQ Z14, Z4, Z4
    VPSUBQ Z25, Z0, Z10
    VPSUBQ Z26, Z1, Z11
    VPSUBQ Z27, Z2, Z12
    VPSUBQ Z28, Z3, Z13
    VPSUBQ Z29, Z4, Z14
    VPSRAQ $63, Z10, Z20                                     // Z20 = -1 if borrow, 0 otherwise
    VPADDQ Z20, Z11, Z11                                     // Z11 -= borrow
    VPSRAQ $63, Z11, Z20
    VPADDQ Z20, Z12, Z12
    VPSRAQ $63, Z12, Z20
    VPADDQ Z20, Z13, Z13
    VPSRAQ $63, Z13, Z20
    VPADDQ Z20, Z14, Z14
    VPSRAQ $63, Z14, Z20                                     // Z20 = all 1s if borrow (result < q), all 0s if no borrow
    VPANDQ Z31, Z10, Z10
    VPANDQ Z31, Z11, Z11
    VPANDQ Z31, Z12, Z12
    VPANDQ Z31, Z13, Z13
    VPANDQ Z31, Z14, Z14
    VPANDQ Z20, Z0, Z0                                       // keep original if borrow
    VPANDNQ Z10, Z20, Z10                                     // keep subtracted if no borrow
    VPORQ Z10, Z0, Z0
    VPANDQ Z20, Z1, Z1
    VPANDNQ Z11, Z20, Z11
    VPORQ Z11, Z1, Z1
    VPANDQ Z20, Z2, Z2
    VPANDNQ Z12, Z20, Z12
    VPORQ Z12, Z2, Z2
    VPANDQ Z20, Z3, Z3
    VPANDNQ Z13, Z20, Z13
    VPORQ Z13, Z3, Z3
    VPANDQ Z20, Z4, Z4
    VPANDNQ Z14, Z20, Z14
    VPORQ Z14, Z4, Z4
    // Convert result from radix-52 back to radix-64
    // Convert from radix-52 to radix-64
    VPSLLQ $52, Z1, Z18
    VPORQ Z18, Z0, Z14
    VPSRLQ $12, Z1, Z18
    VPSLLQ $40, Z2, Z19
    VPORQ Z19, Z18, Z15
    VPSRLQ $24, Z2, Z18
    VPSLLQ $28, Z3, Z19
    VPORQ Z19, Z18, Z16
    VPSRLQ $36, Z3, Z18
    VPSLLQ $16, Z4, Z19
    VPORQ Z19, Z18, Z17
    // Transpose back to AoS format and store
// 4x8 reverse transpose (SoA to AoS)
    VMOVDQU64 ·permuteIdxIFMA<>(SB), Z22
    VPERMQ Z14, Z22, Z14
    VPERMQ Z15, Z22, Z15
    VPERMQ Z16, Z22, Z16
    VPERMQ Z17, Z22, Z17
    VPUNPCKLQDQ Z15, Z14, Z18                                     // pairs (a0,a1) for elements 0,1,4,5
    VPUNPCKHQDQ Z15, Z14, Z19                                     // pairs (a0,a1) for elements 2,3,6,7
    VPUNPCKLQDQ Z17, Z16, Z20                                     // pairs (a2,a3) for elements 0,1,4,5
    VPUNPCKHQDQ Z17, Z16, Z21                                     // pairs (a2,a3) for elements 2,3,6,7
    VSHUFI64X2 $0x44, Z20, Z18, Z10
    VSHUFI64X2 $0x44, Z21, Z19, Z11
    VSHUFI64X2 $0xEE, Z20, Z18, Z12
    VSHUFI64X2 $0xEE, Z21, Z19, Z13
    VSHUFI64X2 $0xD8, Z10, Z10, Z10
    VSHUFI64X2 $0xD8, Z11, Z11, Z11
    VSHUFI64X2 $0xD8, Z12, Z12, Z12
    VSHUFI64X2 $0xD8, Z13, Z13, Z13
    VMOVDQU64 Z10, 0(R14)
    VMOVDQU64 Z11, 64(R14)
    VMOVDQU64 Z12, 128(R14)
    VMOVDQU64 Z13, 192(R14)
    // Advance pointers
    ADDQ $256, R13
    ADDQ $256, CX
    ADDQ $256, R14
    DECQ BX                                                // processed 1 group of 8 elements
    JMP loop_1
done_2:
    RET
