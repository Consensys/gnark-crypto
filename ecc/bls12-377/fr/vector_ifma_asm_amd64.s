// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

//go:build amd64 && !purego

#include "textflag.h"

// Test functions for verifying AVX-512 IFMA instruction usage and transpose operations.
// These serve as reference implementations for correct Plan9 assembly operand ordering.

// Permutation index for reordering [0,2,1,3,4,6,5,7] -> [0,1,2,3,4,5,6,7]
// This fixes the even/odd interleaving from VPUNPCKLQDQ/VPUNPCKHQDQ
DATA ·permuteIdx<>+0(SB)/8, $0 // position 0 -> 0
DATA ·permuteIdx<>+8(SB)/8, $2 // position 1 -> 2
DATA ·permuteIdx<>+16(SB)/8, $1 // position 2 -> 1
DATA ·permuteIdx<>+24(SB)/8, $3 // position 3 -> 3
DATA ·permuteIdx<>+32(SB)/8, $4 // position 4 -> 4
DATA ·permuteIdx<>+40(SB)/8, $6 // position 5 -> 6
DATA ·permuteIdx<>+48(SB)/8, $5 // position 6 -> 5
DATA ·permuteIdx<>+56(SB)/8, $7 // position 7 -> 7
GLOBL ·permuteIdx<>(SB), RODATA|NOPTR, $64

// func testVPERMQ(in, out *uint64)
// Test function to verify VPERMQ operand ordering in Plan9 assembly.
// Uses index vector [0, 2, 1, 3, 4, 6, 5, 7] to permute input.
TEXT ·testVPERMQ(SB), NOSPLIT, $0-16
	MOVQ in+0(FP), AX
	MOVQ out+8(FP), BX

	// Load input data
	VMOVDQU64 0(AX), Z0

	// Load permutation indices
	VMOVDQU64 ·permuteIdx<>(SB), Z1

	// Try VPERMQ with operand order: src, idx, dst
	// Intel syntax: VPERMQ dst, idx, src
	// Plan9 syntax (hypothesis 2): VPERMQ src, idx, dst
	VPERMQ Z0, Z1, Z2

	// Store result
	VMOVDQU64 Z2, 0(BX)

	RET

// func testVPMADD52LUQ(acc, a, b, result *uint64)
// Test function for VPMADD52LUQ operand ordering.
// Computes: result = acc + (a * b) & ((1<<52)-1) (low 52 bits of product)
//
// Intel syntax: VPMADD52LUQ zmm1, zmm2, zmm3  means zmm1 += low52(zmm2 * zmm3)
// We need to verify Plan9 operand order.
TEXT ·testVPMADD52LUQ(SB), NOSPLIT, $0-32
	MOVQ acc+0(FP), AX
	MOVQ a+8(FP), BX
	MOVQ b+16(FP), CX
	MOVQ result+24(FP), DX

	// Load operands
	VMOVDQU64 0(AX), Z0 // accumulator
	VMOVDQU64 0(BX), Z1 // multiplier a
	VMOVDQU64 0(CX), Z2 // multiplier b

	// VPMADD52LUQ: Intel syntax zmm1 += low52(zmm2 * zmm3)
	// Plan9 syntax (hypothesis): VPMADD52LUQ src2, src1, dst
	// Where: dst += low52(src1 * src2)
	// So for Z0 += low52(Z1 * Z2), we try: VPMADD52LUQ Z2, Z1, Z0
	VPMADD52LUQ Z2, Z1, Z0

	// Store result
	VMOVDQU64 Z0, 0(DX)

	RET

// func testVPMADD52HUQ(acc, a, b, result *uint64)
// Test function for VPMADD52HUQ operand ordering.
// Computes: result = acc + (a * b) >> 52 (high 52 bits of product, shifted)
//
// Intel syntax: VPMADD52HUQ zmm1, zmm2, zmm3  means zmm1 += high52(zmm2 * zmm3)
TEXT ·testVPMADD52HUQ(SB), NOSPLIT, $0-32
	MOVQ acc+0(FP), AX
	MOVQ a+8(FP), BX
	MOVQ b+16(FP), CX
	MOVQ result+24(FP), DX

	// Load operands
	VMOVDQU64 0(AX), Z0 // accumulator
	VMOVDQU64 0(BX), Z1 // multiplier a
	VMOVDQU64 0(CX), Z2 // multiplier b

	// VPMADD52HUQ: Intel syntax zmm1 += high52(zmm2 * zmm3)
	// Plan9 syntax (hypothesis): VPMADD52HUQ src2, src1, dst
	VPMADD52HUQ Z2, Z1, Z0

	// Store result
	VMOVDQU64 Z0, 0(DX)

	RET

// func testTransposeAoSToSoA(in, out *uint64)
// Transposes 8 elements from Array-of-Structures to Structure-of-Arrays format.
//
// Input layout (32 uint64 values, 8 elements × 4 limbs each):
//   in[0:8]   = [e0.a0, e0.a1, e0.a2, e0.a3, e1.a0, e1.a1, e1.a2, e1.a3]
//   in[8:16]  = [e2.a0, e2.a1, e2.a2, e2.a3, e3.a0, e3.a1, e3.a2, e3.a3]
//   in[16:24] = [e4.a0, e4.a1, e4.a2, e4.a3, e5.a0, e5.a1, e5.a2, e5.a3]
//   in[24:32] = [e6.a0, e6.a1, e6.a2, e6.a3, e7.a0, e7.a1, e7.a2, e7.a3]
//
// Output layout (32 uint64 values, 4 limbs × 8 elements each):
//   out[0:8]   = [e0.a0, e1.a0, e2.a0, e3.a0, e4.a0, e5.a0, e6.a0, e7.a0]  // all a0
//   out[8:16]  = [e0.a1, e1.a1, e2.a1, e3.a1, e4.a1, e5.a1, e6.a1, e7.a1]  // all a1
//   out[16:24] = [e0.a2, e1.a2, e2.a2, e3.a2, e4.a2, e5.a2, e6.a2, e7.a2]  // all a2
//   out[24:32] = [e0.a3, e1.a3, e2.a3, e3.a3, e4.a3, e5.a3, e6.a3, e7.a3]  // all a3
TEXT ·testTransposeAoSToSoA(SB), NOSPLIT, $0-16
	MOVQ in+0(FP), AX
	MOVQ out+8(FP), BX

	// Load 8 elements (4 ZMM registers, 2 elements each)
	// Z0 = [e0.a0, e0.a1, e0.a2, e0.a3, e1.a0, e1.a1, e1.a2, e1.a3]
	// Z1 = [e2.a0, e2.a1, e2.a2, e2.a3, e3.a0, e3.a1, e3.a2, e3.a3]
	// Z2 = [e4.a0, e4.a1, e4.a2, e4.a3, e5.a0, e5.a1, e5.a2, e5.a3]
	// Z3 = [e6.a0, e6.a1, e6.a2, e6.a3, e7.a0, e7.a1, e7.a2, e7.a3]
	VMOVDQU64 0(AX), Z0
	VMOVDQU64 64(AX), Z1
	VMOVDQU64 128(AX), Z2
	VMOVDQU64 192(AX), Z3

	// Step 1: Use VPUNPCKLQDQ/VPUNPCKHQDQ to interleave within 128-bit lanes
	// VPUNPCKLQDQ interleaves low qwords from each 128-bit lane
	// VPUNPCKHQDQ interleaves high qwords from each 128-bit lane
	//
	// Plan9 syntax: VPUNPCKLQDQ src2, src1, dst
	// Intel syntax: VPUNPCKLQDQ dst, src1, src2
	// Result per lane: [src1.low, src2.low]
	//
	// Z0 lanes: [e0.a0,e0.a1 | e0.a2,e0.a3 | e1.a0,e1.a1 | e1.a2,e1.a3]
	// Z1 lanes: [e2.a0,e2.a1 | e2.a2,e2.a3 | e3.a0,e3.a1 | e3.a2,e3.a3]
	//
	// After VPUNPCKLQDQ Z1, Z0, Z4:
	//   Lane 0: [e0.a0, e2.a0]
	//   Lane 1: [e0.a2, e2.a2]
	//   Lane 2: [e1.a0, e3.a0]
	//   Lane 3: [e1.a2, e3.a2]
	VPUNPCKLQDQ Z1, Z0, Z4
	VPUNPCKHQDQ Z1, Z0, Z5 // [e0.a1,e2.a1 | e0.a3,e2.a3 | e1.a1,e3.a1 | e1.a3,e3.a3]
	VPUNPCKLQDQ Z3, Z2, Z6 // [e4.a0,e6.a0 | e4.a2,e6.a2 | e5.a0,e7.a0 | e5.a2,e7.a2]
	VPUNPCKHQDQ Z3, Z2, Z7 // [e4.a1,e6.a1 | e4.a3,e6.a3 | e5.a1,e7.a1 | e5.a3,e7.a3]

	// Step 2: Use VSHUFI64X2 to combine 128-bit lanes
	// Plan9: VSHUFI64X2 $imm, src2, src1, dst
	// Intel: VSHUFI64X2 dst, src1, src2, imm
	// For imm8: lanes 0,1 select from src1, lanes 2,3 select from src2
	//
	// $0x88 = binary 10_00_10_00:
	//   dst.lane0 = src1.lane[0], dst.lane1 = src1.lane[2]
	//   dst.lane2 = src2.lane[0], dst.lane3 = src2.lane[2]
	//
	// Z4 lanes: [e0.a0,e2.a0 | e0.a2,e2.a2 | e1.a0,e3.a0 | e1.a2,e3.a2]
	// Z6 lanes: [e4.a0,e6.a0 | e4.a2,e6.a2 | e5.a0,e7.a0 | e5.a2,e7.a2]
	// After VSHUFI64X2 $0x88, Z6, Z4, Z8:
	//   Z8 = [e0.a0,e2.a0 | e1.a0,e3.a0 | e4.a0,e6.a0 | e5.a0,e7.a0]
	//   As qwords: [e0.a0, e2.a0, e1.a0, e3.a0, e4.a0, e6.a0, e5.a0, e7.a0]
	VSHUFI64X2 $0x88, Z6, Z4, Z8  // all a0 (unordered)
	VSHUFI64X2 $0xDD, Z6, Z4, Z9  // all a2 (unordered)
	VSHUFI64X2 $0x88, Z7, Z5, Z10 // all a1 (unordered)
	VSHUFI64X2 $0xDD, Z7, Z5, Z11 // all a3 (unordered)

	// Step 3: Fix element ordering using VPERMQ
	// Current: [e0, e2, e1, e3, e4, e6, e5, e7]
	// Desired: [e0, e1, e2, e3, e4, e5, e6, e7]
	// Permutation: swap positions 1<->2 and 5<->6
	// Use VPERMQ with index vector [0, 2, 1, 3, 4, 6, 5, 7]
	VMOVDQU64 ·permuteIdx<>(SB), Z15

	// VPERMQ in Plan9: VPERMQ src, idx, dst
	// This permutes src using indices from idx and stores in dst
	VPERMQ Z8, Z15, Z8
	VPERMQ Z9, Z15, Z9
	VPERMQ Z10, Z15, Z10
	VPERMQ Z11, Z15, Z11

	// Store results: out[0:8]=a0, out[8:16]=a1, out[16:24]=a2, out[24:32]=a3
	VMOVDQU64 Z8, 0(BX)    // all a0
	VMOVDQU64 Z10, 64(BX)  // all a1
	VMOVDQU64 Z9, 128(BX)  // all a2
	VMOVDQU64 Z11, 192(BX) // all a3

	RET

// func testTransposeSoAToAoS(in, out *uint64)
// Transposes from Structure-of-Arrays back to Array-of-Structures format.
// This is the inverse of testTransposeAoSToSoA.
//
// Input layout (SoA - 4 limbs × 8 elements):
//   in[0:8]   = [e0.a0, e1.a0, e2.a0, e3.a0, e4.a0, e5.a0, e6.a0, e7.a0]
//   in[8:16]  = [e0.a1, e1.a1, e2.a1, e3.a1, e4.a1, e5.a1, e6.a1, e7.a1]
//   in[16:24] = [e0.a2, e1.a2, e2.a2, e3.a2, e4.a2, e5.a2, e6.a2, e7.a2]
//   in[24:32] = [e0.a3, e1.a3, e2.a3, e3.a3, e4.a3, e5.a3, e6.a3, e7.a3]
//
// Output layout (AoS - 8 elements × 4 limbs):
//   out[0:8]   = [e0.a0, e0.a1, e0.a2, e0.a3, e1.a0, e1.a1, e1.a2, e1.a3]
//   out[8:16]  = [e2.a0, e2.a1, e2.a2, e2.a3, e3.a0, e3.a1, e3.a2, e3.a3]
//   out[16:24] = [e4.a0, e4.a1, e4.a2, e4.a3, e5.a0, e5.a1, e5.a2, e5.a3]
//   out[24:32] = [e6.a0, e6.a1, e6.a2, e6.a3, e7.a0, e7.a1, e7.a2, e7.a3]
TEXT ·testTransposeSoAToAoS(SB), NOSPLIT, $0-16
	MOVQ in+0(FP), AX
	MOVQ out+8(FP), BX

	// Load SoA data
	// Z0 = [e0.a0, e1.a0, e2.a0, e3.a0, e4.a0, e5.a0, e6.a0, e7.a0]  (all a0)
	// Z1 = [e0.a1, e1.a1, e2.a1, e3.a1, e4.a1, e5.a1, e6.a1, e7.a1]  (all a1)
	// Z2 = [e0.a2, e1.a2, e2.a2, e3.a2, e4.a2, e5.a2, e6.a2, e7.a2]  (all a2)
	// Z3 = [e0.a3, e1.a3, e2.a3, e3.a3, e4.a3, e5.a3, e6.a3, e7.a3]  (all a3)
	VMOVDQU64 0(AX), Z0
	VMOVDQU64 64(AX), Z1
	VMOVDQU64 128(AX), Z2
	VMOVDQU64 192(AX), Z3

	// Load permutation index to fix ordering before interleaving
	// We need to pre-permute to account for how VPUNPCKLQDQ pairs elements
	// Input is [e0,e1,e2,e3,e4,e5,e6,e7], we want it reordered so that
	// after VPUNPCKLQDQ pairs (0,2), (1,3), etc. we get the right grouping
	// Permute: [0,2,1,3,4,6,5,7] so pairs become (e0,e1), (e2,e3), etc.
	// VPERMQ in Plan9: VPERMQ src, idx, dst
	VMOVDQU64 ·permuteIdx<>(SB), Z15
	VPERMQ    Z0, Z15, Z0            // [e0.a0, e2.a0, e1.a0, e3.a0, e4.a0, e6.a0, e5.a0, e7.a0]
	VPERMQ    Z1, Z15, Z1
	VPERMQ    Z2, Z15, Z2
	VPERMQ    Z3, Z15, Z3

	// Now Z0 lanes: [e0.a0,e2.a0 | e1.a0,e3.a0 | e4.a0,e6.a0 | e5.a0,e7.a0]
	// And Z1 lanes: [e0.a1,e2.a1 | e1.a1,e3.a1 | e4.a1,e6.a1 | e5.a1,e7.a1]

	// Step 1: Interleave a0 with a1 to get (a0,a1) pairs
	// VPUNPCKLQDQ Z1, Z0, Z4: pairs low qwords within lanes
	//   Lane 0: [e0.a0, e0.a1]
	//   Lane 1: [e1.a0, e1.a1]
	//   Lane 2: [e4.a0, e4.a1]
	//   Lane 3: [e5.a0, e5.a1]
	VPUNPCKLQDQ Z1, Z0, Z4

	// VPUNPCKHQDQ Z1, Z0, Z5: pairs high qwords within lanes
	//   Lane 0: [e2.a0, e2.a1]
	//   Lane 1: [e3.a0, e3.a1]
	//   Lane 2: [e6.a0, e6.a1]
	//   Lane 3: [e7.a0, e7.a1]
	VPUNPCKHQDQ Z1, Z0, Z5

	// Similarly for a2 and a3
	// Z6: [e0.a2,e0.a3 | e1.a2,e1.a3 | e4.a2,e4.a3 | e5.a2,e5.a3]
	VPUNPCKLQDQ Z3, Z2, Z6

	// Z7: [e2.a2,e2.a3 | e3.a2,e3.a3 | e6.a2,e6.a3 | e7.a2,e7.a3]
	VPUNPCKHQDQ Z3, Z2, Z7

	// Step 2: Combine (a0,a1) with (a2,a3) to form complete elements
	// Use VSHUFI64X2 to select 128-bit lanes
	//
	// Z4 = [e0.a0,e0.a1 | e1.a0,e1.a1 | e4.a0,e4.a1 | e5.a0,e5.a1]
	// Z6 = [e0.a2,e0.a3 | e1.a2,e1.a3 | e4.a2,e4.a3 | e5.a2,e5.a3]
	//
	// VSHUFI64X2 $0x44 = 01_00_01_00: [src1[0], src1[1], src2[0], src2[1]]
	// This gives: [e0.a0,e0.a1 | e1.a0,e1.a1 | e0.a2,e0.a3 | e1.a2,e1.a3]
	// We need lanes reordered to: [e0.a0,e0.a1 | e0.a2,e0.a3 | e1.a0,e1.a1 | e1.a2,e1.a3]
	// Use VSHUFI64X2 $0xD8 = 11_01_10_00 to swap lanes 1 and 2: [0,2,1,3]
	VSHUFI64X2 $0x44, Z6, Z4, Z8
	VSHUFI64X2 $0xD8, Z8, Z8, Z8 // [e0 | e1]

	VSHUFI64X2 $0x44, Z7, Z5, Z9
	VSHUFI64X2 $0xD8, Z9, Z9, Z9 // [e2 | e3]

	// For output [e4, e5]: take lanes 2,3 from Z4 and Z6
	// VSHUFI64X2 $0xEE = 11_10_11_10: [src1[2], src1[3], src2[2], src2[3]]
	// This gives: [e4.a0,e4.a1 | e5.a0,e5.a1 | e4.a2,e4.a3 | e5.a2,e5.a3]
	// Then reorder lanes with $0xD8 to get [e4 | e5]
	VSHUFI64X2 $0xEE, Z6, Z4, Z10
	VSHUFI64X2 $0xD8, Z10, Z10, Z10 // [e4 | e5]

	VSHUFI64X2 $0xEE, Z7, Z5, Z11
	VSHUFI64X2 $0xD8, Z11, Z11, Z11 // [e6 | e7]

	// Store results
	VMOVDQU64 Z8, 0(BX)    // [e0, e1]
	VMOVDQU64 Z9, 64(BX)   // [e2, e3]
	VMOVDQU64 Z10, 128(BX) // [e4, e5]
	VMOVDQU64 Z11, 192(BX) // [e6, e7]

	RET

// func testRadix52RoundTrip(in, out *uint64)
// Tests the radix-52 conversion round-trip:
// 1. Load 8 elements in radix-64 format
// 2. Transpose to SoA layout
// 3. Convert to radix-52 (5 limbs)
// 4. Convert back to radix-64 (4 limbs)
// 5. Transpose back to AoS layout
// 6. Store
// If all conversions are correct, out should equal in.
TEXT ·testRadix52RoundTrip(SB), NOSPLIT, $0-16
	MOVQ in+0(FP), AX
	MOVQ out+8(FP), BX

	// Load mask52 constant
	MOVQ         $0xFFFFFFFFFFFFF, R15
	VPBROADCASTQ R15, Z31

	// ========== LOAD AND TRANSPOSE TO SoA ==========
	// Load 8 elements (256 bytes = 8 elements × 4 limbs × 8 bytes)
	VMOVDQU64 0(AX), Z0   // elements 0,1
	VMOVDQU64 64(AX), Z1  // elements 2,3
	VMOVDQU64 128(AX), Z2 // elements 4,5
	VMOVDQU64 192(AX), Z3 // elements 6,7

	// Transpose AoS to SoA
	VPUNPCKLQDQ Z1, Z0, Z4
	VPUNPCKHQDQ Z1, Z0, Z5
	VPUNPCKLQDQ Z3, Z2, Z6
	VPUNPCKHQDQ Z3, Z2, Z7

	VSHUFI64X2 $0x88, Z6, Z4, Z8  // all a0 (unordered)
	VSHUFI64X2 $0xDD, Z6, Z4, Z9  // all a2 (unordered)
	VSHUFI64X2 $0x88, Z7, Z5, Z10 // all a1 (unordered)
	VSHUFI64X2 $0xDD, Z7, Z5, Z11 // all a3 (unordered)

	// Fix element ordering with VPERMQ
	VMOVDQU64 ·permuteIdx<>(SB), Z15
	VPERMQ    Z8, Z15, Z8            // Z8 = all a0, correctly ordered
	VPERMQ    Z10, Z15, Z10          // Z10 = all a1, correctly ordered
	VPERMQ    Z9, Z15, Z9            // Z9 = all a2, correctly ordered
	VPERMQ    Z11, Z15, Z11          // Z11 = all a3, correctly ordered

	// Now: Z8=a0, Z10=a1, Z9=a2, Z11=a3

	// ========== CONVERT TO RADIX-52 ==========
	// Input:  a0, a1, a2, a3 (each 64-bit)
	// Output: l0, l1, l2, l3, l4 (each 52-bit in 64-bit lane)

	// l0 = a0 & mask52
	VPANDQ Z31, Z8, Z0

	// l1 = (a0 >> 52) | ((a1 << 12) & mask52)
	VPSRLQ $52, Z8, Z18
	VPSLLQ $12, Z10, Z19
	VPORQ  Z18, Z19, Z18
	VPANDQ Z31, Z18, Z1

	// l2 = (a1 >> 40) | ((a2 << 24) & mask52)
	VPSRLQ $40, Z10, Z18
	VPSLLQ $24, Z9, Z19
	VPORQ  Z18, Z19, Z18
	VPANDQ Z31, Z18, Z2

	// l3 = (a2 >> 28) | ((a3 << 36) & mask52)
	VPSRLQ $28, Z9, Z18
	VPSLLQ $36, Z11, Z19
	VPORQ  Z18, Z19, Z18
	VPANDQ Z31, Z18, Z3

	// l4 = a3 >> 16
	VPSRLQ $16, Z11, Z4

	// Now: Z0=l0, Z1=l1, Z2=l2, Z3=l3, Z4=l4 (radix-52)

	// ========== CONVERT BACK TO RADIX-64 ==========
	// Input:  l0, l1, l2, l3, l4 (each 52-bit)
	// Output: a0, a1, a2, a3 (each 64-bit)

	// a0 = l0 | (l1 << 52)
	VPSLLQ $52, Z1, Z18
	VPORQ  Z18, Z0, Z8  // Z8 = a0

	// a1 = (l1 >> 12) | (l2 << 40)
	VPSRLQ $12, Z1, Z18
	VPSLLQ $40, Z2, Z19
	VPORQ  Z19, Z18, Z10 // Z10 = a1

	// a2 = (l2 >> 24) | (l3 << 28)
	VPSRLQ $24, Z2, Z18
	VPSLLQ $28, Z3, Z19
	VPORQ  Z19, Z18, Z9 // Z9 = a2

	// a3 = (l3 >> 36) | (l4 << 16)
	VPSRLQ $36, Z3, Z18
	VPSLLQ $16, Z4, Z19
	VPORQ  Z19, Z18, Z11 // Z11 = a3

	// Now: Z8=a0, Z10=a1, Z9=a2, Z11=a3 (radix-64)

	// ========== TRANSPOSE BACK TO AoS ==========
	// Pre-permute inputs
	VMOVDQU64 ·permuteIdx<>(SB), Z15
	VPERMQ    Z8, Z15, Z8
	VPERMQ    Z10, Z15, Z10
	VPERMQ    Z9, Z15, Z9
	VPERMQ    Z11, Z15, Z11

	// Interleave a0 with a1, a2 with a3
	VPUNPCKLQDQ Z10, Z8, Z4
	VPUNPCKHQDQ Z10, Z8, Z5
	VPUNPCKLQDQ Z11, Z9, Z6
	VPUNPCKHQDQ Z11, Z9, Z7

	// Combine to form complete elements
	VSHUFI64X2 $0x44, Z6, Z4, Z0
	VSHUFI64X2 $0xD8, Z0, Z0, Z0 // [e0 | e1]

	VSHUFI64X2 $0x44, Z7, Z5, Z1
	VSHUFI64X2 $0xD8, Z1, Z1, Z1 // [e2 | e3]

	VSHUFI64X2 $0xEE, Z6, Z4, Z2
	VSHUFI64X2 $0xD8, Z2, Z2, Z2 // [e4 | e5]

	VSHUFI64X2 $0xEE, Z7, Z5, Z3
	VSHUFI64X2 $0xD8, Z3, Z3, Z3 // [e6 | e7]

	// Store results
	VMOVDQU64 Z0, 0(BX)
	VMOVDQU64 Z1, 64(BX)
	VMOVDQU64 Z2, 128(BX)
	VMOVDQU64 Z3, 192(BX)

	RET
