// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package amd64

import (
	"fmt"

	"github.com/consensys/bavard/amd64"
)

// emitIFMAConstants emits the precomputed constants needed for IFMA operations
func (f *FFAmd64) emitIFMAConstants() {
	f.Comment("Permutation index for IFMA transpose: [0, 2, 1, 3, 4, 6, 5, 7]")
	f.Comment("This swaps positions 1<->2 and 5<->6 to fix even/odd interleaving")
	f.DATA("·permuteIdxIFMA<>", 0, 8, "$0")
	f.DATA("·permuteIdxIFMA<>", 8, 8, "$2")
	f.DATA("·permuteIdxIFMA<>", 16, 8, "$1")
	f.DATA("·permuteIdxIFMA<>", 24, 8, "$3")
	f.DATA("·permuteIdxIFMA<>", 32, 8, "$4")
	f.DATA("·permuteIdxIFMA<>", 40, 8, "$6")
	f.DATA("·permuteIdxIFMA<>", 48, 8, "$5")
	f.DATA("·permuteIdxIFMA<>", 56, 8, "$7")
	f.GLOBL("·permuteIdxIFMA<>", "RODATA|NOPTR", 64)
	f.WriteLn("")

	// Note: Previously had 2q, 4q, 8q, 16q constants for binary reduction.
	// Now using Barrett reduction, so these are no longer needed.
}

// generateMulVecIFMA generates AVX-512 IFMA based vector multiplication.
// This uses the radix-52 representation with vpmadd52luq/vpmadd52huq instructions.
//
// For a 4-word (256-bit) field element in radix-64:
//
//	[a0, a1, a2, a3] where each ai is 64 bits
//
// In radix-52 representation:
//
//	[l0, l1, l2, l3, l4] where each li is 52 bits (stored in 64-bit lane)
//
// The conversion:
//
//	l0 = a0 & 0xFFFFFFFFFFFFF (low 52 bits of a0)
//	l1 = (a0 >> 52) | ((a1 & 0xFFFFFFFFF) << 12) (12 bits from a0 + 40 bits from a1)
//	l2 = (a1 >> 40) | ((a2 & 0xFFFFFFF) << 24)   (24 bits from a1 + 28 bits from a2)
//	l3 = (a2 >> 28) | ((a3 & 0xFFFF) << 36)      (36 bits from a2 + 16 bits from a3)
//	l4 = a3 >> 16                                 (48 bits from a3)
//
// Montgomery multiplication using IFMA with BPS (Block Product Scanning):
// For A * B mod q:
// 1. Compute T = A * B (in radix-52, 10 limbs)
// 2. Compute m = T[0] * qInvNeg52 mod 2^52
// 3. Add m * q to T (reduces T[0] to 0)
// 4. Shift right by 52 bits
// 5. Repeat for each limb
// 6. Final conditional subtraction
func (f *FFAmd64) generateMulVecIFMA() {
	// Emit DATA constants before the TEXT block
	f.emitIFMAConstants()

	f.Comment("mulVecIFMA(res, a, b *Element, n uint64)")
	f.Comment("Performs n multiplications using AVX-512 IFMA instructions")
	f.Comment("Processes 8 elements in parallel using radix-52 representation")

	const argSize = 4 * 8
	// We only need 4 GP registers (res, a, b, n) and use ZMM registers for SIMD
	// No stack allocation needed as we have plenty of registers
	stackSize := 0
	registers := f.FnHeader("mulVecIFMA", stackSize, argSize, amd64.AX, amd64.DX)
	defer f.AssertCleanStack(stackSize, 0)

	// Register allocation (4 registers for pointers/counters)
	addrRes := f.Pop(&registers)
	addrA := f.Pop(&registers)
	addrB := f.Pop(&registers)
	n := f.Pop(&registers)

	// Labels
	loop := f.NewLabel("loop")
	done := f.NewLabel("done")

	// Load arguments
	f.MOVQ("res+0(FP)", addrRes)
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("b+16(FP)", addrB)
	f.MOVQ("n+24(FP)", n)

	// Constants for radix-52
	f.Comment("Load constants for radix-52 conversion and reduction")

	// Mask for 52-bit extraction - use R15 as dedicated mask register
	f.MOVQ("$0xFFFFFFFFFFFFF", amd64.R15, "52-bit mask in R15")
	f.VPBROADCASTQ(amd64.R15, amd64.Z31, "Z31 = mask52 for SIMD ops")

	// Load qInvNeg for Montgomery reduction (52-bit version)
	// qInvNeg52 = qInvNeg mod 2^52
	f.MOVQ("$const_qInvNeg", amd64.AX)
	f.ANDQ(amd64.R15, amd64.AX, "keep low 52 bits using mask in R15")
	f.VPBROADCASTQ(amd64.AX, amd64.Z30, "Z30 = qInvNeg52")

	// For IFMA, we need the modulus in radix-52 form
	// q = q0 + q1*2^64 + q2*2^128 + q3*2^192
	// In radix-52: ql0, ql1, ql2, ql3, ql4
	f.Comment("Load modulus in radix-52 form")
	f.loadModulusRadix52()

	f.LABEL(loop)
	f.TESTQ(n, n)
	f.JEQ(done, "n == 0, we are done")

	f.Comment("Process 8 elements in parallel")

	// Load 8 elements from a into radix-52 format
	f.Comment("Load and convert 8 elements from a[] to radix-52")
	f.loadAndConvertToRadix52(addrA, "Z0", "Z1", "Z2", "Z3", "Z4") // a[0..7] in Z0-Z4

	// Load 8 elements from b into radix-52 format
	f.Comment("Load and convert 8 elements from b[] to radix-52")
	f.loadAndConvertToRadix52(addrB, "Z5", "Z6", "Z7", "Z8", "Z9") // b[0..7] in Z5-Z9

	// Perform Montgomery multiplication using IFMA
	f.Comment("Montgomery multiplication using IFMA (BPS method)")
	f.montgomeryMulIFMA()

	// Result is in Z0-Z4 (radix-52, SoA format)
	// x16 correction is already fused into montgomeryMulIFMA
	// Result is in [0, 32q), need Barrett reduction to get to [0, q)
	f.Comment("Barrett reduction from [0, 32q) to [0, q)")
	f.barrettReduction()

	// Convert result back to radix-64
	f.Comment("Convert result from radix-52 back to radix-64")
	f.convertFromRadix52("Z0", "Z1", "Z2", "Z3", "Z4", "Z14", "Z15", "Z16", "Z17")

	// Transpose back and store
	f.Comment("Transpose back to AoS format and store")
	f.transposeAndStore(addrRes)

	f.Comment("Advance pointers")
	f.ADDQ("$256", addrA) // 8 elements * 32 bytes
	f.ADDQ("$256", addrB) // 8 elements * 32 bytes
	f.ADDQ("$256", addrRes)
	f.DECQ(n, "processed 1 group of 8 elements")

	f.JMP(loop)

	f.LABEL(done)
	f.RET()

	f.Push(&registers, addrRes, addrA, addrB, n)
}

func (f *FFAmd64) loadModulusRadix52() {
	// Load q0..q3 and convert to radix-52
	// This is done at function entry and stays in registers
	f.Comment("q in radix-52: Z25=ql0, Z26=ql1, Z27=ql2, Z28=ql3, Z29=ql4")

	// For BLS12-377 fr: q = [q0, q1, q2, q3] in radix-64
	// q[base16] = 0x12ab655e9a2ca55660b44d1e5c37b00159aa76fed00000010a11800000000001
	// We compute the radix-52 limbs

	// Note: In a full implementation, these would be precomputed constants
	// For prototype, we compute them at runtime (slower but demonstrates the concept)
	// Using R9-R12 to avoid clobbering addrB (CX) and n (BX)

	f.Comment("Load q0-q3 and convert to radix-52")
	f.MOVQ("$const_q0", amd64.R9)
	f.MOVQ("$const_q1", amd64.R10)
	f.MOVQ("$const_q2", amd64.R11)
	f.MOVQ("$const_q3", amd64.R12)

	// ql0 = q0 & mask52  (R15 contains the 52-bit mask)
	f.MOVQ(amd64.R9, amd64.R8)
	f.ANDQ(amd64.R15, amd64.R8)
	f.VPBROADCASTQ(amd64.R8, amd64.Z25)

	// ql1 = (q0 >> 52) | (q1 << 12) & mask52
	f.SHRQ("$52", amd64.R9)
	f.MOVQ(amd64.R10, amd64.R8)
	f.SHLQ("$12", amd64.R8)
	f.ORQ(amd64.R9, amd64.R8)
	f.ANDQ(amd64.R15, amd64.R8)
	f.VPBROADCASTQ(amd64.R8, amd64.Z26)

	// ql2 = (q1 >> 40) | (q2 << 24) & mask52
	f.SHRQ("$40", amd64.R10)
	f.MOVQ(amd64.R11, amd64.R8)
	f.SHLQ("$24", amd64.R8)
	f.ORQ(amd64.R10, amd64.R8)
	f.ANDQ(amd64.R15, amd64.R8)
	f.VPBROADCASTQ(amd64.R8, amd64.Z27)

	// ql3 = (q2 >> 28) | (q3 << 36) & mask52
	f.SHRQ("$28", amd64.R11)
	f.MOVQ(amd64.R12, amd64.R8)
	f.SHLQ("$36", amd64.R8)
	f.ORQ(amd64.R11, amd64.R8)
	f.ANDQ(amd64.R15, amd64.R8)
	f.VPBROADCASTQ(amd64.R8, amd64.Z28)

	// ql4 = q3 >> 16
	f.SHRQ("$16", amd64.R12)
	f.VPBROADCASTQ(amd64.R12, amd64.Z29)
}

func (f *FFAmd64) loadAndConvertToRadix52(addr amd64.Register, z0, z1, z2, z3, z4 string) {
	// Load 8 elements (8 * 32 bytes = 256 bytes) and convert to radix-52
	// Each element has 4 limbs: [a0, a1, a2, a3]
	// After conversion: 5 limbs in radix-52

	f.Comment(fmt.Sprintf("Load 8 elements from %s", addr))

	// Load a0 for all 8 elements (bytes 0, 32, 64, ... into Z10)
	// This requires a gather operation or sequential loads with shuffling
	// For simplicity, use VGATHERDPD / manual loading

	f.Comment("Load element words using gather pattern")
	f.VMOVDQU64(fmt.Sprintf("0(%s)", addr), amd64.Z10)   // element 0,1 (64 bytes)
	f.VMOVDQU64(fmt.Sprintf("64(%s)", addr), amd64.Z11)  // element 2,3
	f.VMOVDQU64(fmt.Sprintf("128(%s)", addr), amd64.Z12) // element 4,5
	f.VMOVDQU64(fmt.Sprintf("192(%s)", addr), amd64.Z13) // element 6,7

	// Now we have 8 elements in Z10-Z13 (each Z register has 2 elements)
	// We need to transpose to get:
	// Z10' = [a0[0], a0[1], a0[2], a0[3], a0[4], a0[5], a0[6], a0[7]]
	// Z11' = [a1[0], a1[1], a1[2], a1[3], a1[4], a1[5], a1[6], a1[7]]
	// etc.

	f.Comment("Transpose 8 elements for vertical SIMD processing")
	f.transposeForIFMA("Z10", "Z11", "Z12", "Z13", "Z14", "Z15", "Z16", "Z17")

	// Now Z14=all a0, Z15=all a1, Z16=all a2, Z17=all a3
	// Convert to radix-52

	f.Comment("Convert to radix-52")
	// l0 = a0 & mask52
	f.VPANDQ(amd64.Z31, amd64.Z14, z0)

	// l1 = (a0 >> 52) | ((a1 << 12) & mask52)
	f.VPSRLQ("$52", amd64.Z14, amd64.Z18)
	f.VPSLLQ("$12", amd64.Z15, amd64.Z19)
	f.VPORQ(amd64.Z18, amd64.Z19, amd64.Z18)
	f.VPANDQ(amd64.Z31, amd64.Z18, z1)

	// l2 = (a1 >> 40) | ((a2 << 24) & mask52)
	f.VPSRLQ("$40", amd64.Z15, amd64.Z18)
	f.VPSLLQ("$24", amd64.Z16, amd64.Z19)
	f.VPORQ(amd64.Z18, amd64.Z19, amd64.Z18)
	f.VPANDQ(amd64.Z31, amd64.Z18, z2)

	// l3 = (a2 >> 28) | ((a3 << 36) & mask52)
	f.VPSRLQ("$28", amd64.Z16, amd64.Z18)
	f.VPSLLQ("$36", amd64.Z17, amd64.Z19)
	f.VPORQ(amd64.Z18, amd64.Z19, amd64.Z18)
	f.VPANDQ(amd64.Z31, amd64.Z18, z3)

	// l4 = a3 >> 16
	f.VPSRLQ("$16", amd64.Z17, z4)
}

func (f *FFAmd64) transposeForIFMA(in0, in1, in2, in3, out0, out1, out2, out3 string) {
	// Transpose 8 elements from AoS (Array of Structures) to SoA (Structure of Arrays)
	// Input: in0 = [e0.a0, e0.a1, e0.a2, e0.a3, e1.a0, e1.a1, e1.a2, e1.a3]
	//        in1 = [e2.a0, e2.a1, e2.a2, e2.a3, e3.a0, e3.a1, e3.a2, e3.a3]
	//        in2 = [e4.a0, e4.a1, e4.a2, e4.a3, e5.a0, e5.a1, e5.a2, e5.a3]
	//        in3 = [e6.a0, e6.a1, e6.a2, e6.a3, e7.a0, e7.a1, e7.a2, e7.a3]
	// Output: out0 = [e0.a0, e1.a0, e2.a0, e3.a0, e4.a0, e5.a0, e6.a0, e7.a0]
	//         out1 = [e0.a1, e1.a1, e2.a1, e3.a1, e4.a1, e5.a1, e6.a1, e7.a1]
	//         out2 = [e0.a2, e1.a2, e2.a2, e3.a2, e4.a2, e5.a2, e6.a2, e7.a2]
	//         out3 = [e0.a3, e1.a3, e2.a3, e3.a3, e4.a3, e5.a3, e6.a3, e7.a3]

	f.Comment("8x4 transpose using AVX-512 shuffles")

	// Step 1: Interleave low qwords between pairs
	// VPUNPCKLQDQ interleaves elements at even indices (0,2,4,6 within 128-bit lanes)
	// VPUNPCKHQDQ interleaves elements at odd indices (1,3,5,7 within 128-bit lanes)
	f.VPUNPCKLQDQ(in1, in0, amd64.Z18, "[e0.a0, e2.a0, e0.a2, e2.a2, e1.a0, e3.a0, e1.a2, e3.a2]")
	f.VPUNPCKHQDQ(in1, in0, amd64.Z19, "[e0.a1, e2.a1, e0.a3, e2.a3, e1.a1, e3.a1, e1.a3, e3.a3]")
	f.VPUNPCKLQDQ(in3, in2, amd64.Z20, "[e4.a0, e6.a0, e4.a2, e6.a2, e5.a0, e7.a0, e5.a2, e7.a2]")
	f.VPUNPCKHQDQ(in3, in2, amd64.Z21, "[e4.a1, e6.a1, e4.a3, e6.a3, e5.a1, e7.a1, e5.a3, e7.a3]")

	// Step 2: Interleave across the 4 intermediate registers to separate a0,a1,a2,a3
	// Z18 has: a0 at indices 0,1,4,5 and a2 at indices 2,3,6,7
	// Z20 has: a0 at indices 0,1,4,5 and a2 at indices 2,3,6,7
	f.VSHUFI64X2("$0x88", amd64.Z20, amd64.Z18, out0, "a0: lanes 0,2 from Z18 and Z20")
	f.VSHUFI64X2("$0xDD", amd64.Z20, amd64.Z18, out2, "a2: lanes 1,3 from Z18 and Z20")
	f.VSHUFI64X2("$0x88", amd64.Z21, amd64.Z19, out1, "a1: lanes 0,2 from Z19 and Z21")
	f.VSHUFI64X2("$0xDD", amd64.Z21, amd64.Z19, out3, "a3: lanes 1,3 from Z19 and Z21")

	// Step 3: Fix the element ordering within each output register using VPERMQ
	// After step 2, out0 = [e0.a0, e2.a0, e1.a0, e3.a0, e4.a0, e6.a0, e5.a0, e7.a0]
	// We need:    out0 = [e0.a0, e1.a0, e2.a0, e3.a0, e4.a0, e5.a0, e6.a0, e7.a0]
	// Permutation index [0, 2, 1, 3, 4, 6, 5, 7] swaps positions 1<->2 and 5<->6
	// Load permutation index (precomputed constant ·permuteIdxIFMA<>)
	f.VMOVDQU64("·permuteIdxIFMA<>(SB)", amd64.Z22)

	// Apply VPERMQ: Plan9 syntax is VPERMQ src, idx, dst
	f.VPERMQ(out0, amd64.Z22, out0)
	f.VPERMQ(out1, amd64.Z22, out1)
	f.VPERMQ(out2, amd64.Z22, out2)
	f.VPERMQ(out3, amd64.Z22, out3)
}

func (f *FFAmd64) montgomeryMulIFMA() {
	// Montgomery multiplication for 5-limb radix-52 numbers
	// A in Z0-Z4, B in Z5-Z9
	// Result in Z0-Z4
	//
	// Algorithm: CIOS (Coarsely Integrated Operand Scanning) variant
	// This interleaves multiplication with Montgomery reduction for better efficiency.
	//
	// For each limb i of B:
	//   1. Multiply A by B[i] and add to T
	//   2. Compute Montgomery quotient m and add m*q to T
	//   3. "Shift" by discarding the lowest limb

	f.Comment("Montgomery multiplication using CIOS variant")
	f.Comment("A = [Z0, Z1, Z2, Z3, Z4], B = [Z5, Z6, Z7, Z8, Z9]")

	// Initialize accumulators (6 limbs: T0-T5)
	// We only need 6 limbs because we process one B limb at a time
	f.VPXORQ(amd64.Z10, amd64.Z10, amd64.Z10, "T0")
	f.VPXORQ(amd64.Z11, amd64.Z11, amd64.Z11, "T1")
	f.VPXORQ(amd64.Z12, amd64.Z12, amd64.Z12, "T2")
	f.VPXORQ(amd64.Z13, amd64.Z13, amd64.Z13, "T3")
	f.VPXORQ(amd64.Z14, amd64.Z14, amd64.Z14, "T4")
	f.VPXORQ(amd64.Z15, amd64.Z15, amd64.Z15, "T5 (overflow)")

	// Process each limb of B
	for i := 0; i < 5; i++ {
		bi := fmt.Sprintf("Z%d", i+5) // B[i] is in Z5+i

		f.Comment(fmt.Sprintf("Round %d: process B[%d]", i, i))

		// Step 1: T += A * B[i]
		f.Comment("T += A * B[i]")
		for j := 0; j < 5; j++ {
			aj := fmt.Sprintf("Z%d", j) // A[j] is in Zj
			tLow := fmt.Sprintf("Z%d", j+10)
			tHigh := fmt.Sprintf("Z%d", j+11)
			f.VPMADD52LUQ(bi, aj, tLow)
			f.VPMADD52HUQ(bi, aj, tHigh)
		}

		// Step 2: Normalize T[0] before computing m
		// Propagate any overflow from T[0] to T[1]
		f.Comment("Normalize T[0]")
		f.VPSRLQ("$52", amd64.Z10, amd64.Z20, "carry = T[0] >> 52")
		f.VPANDQ(amd64.Z31, amd64.Z10, amd64.Z10, "T[0] &= mask52")
		f.VPADDQ(amd64.Z20, amd64.Z11, amd64.Z11, "T[1] += carry")

		// Step 3: Compute m = T[0] * qInvNeg52 mod 2^52
		// Since T[0] is now < 2^52, we can use VPMADD52LUQ
		f.Comment("m = T[0] * qInvNeg52 mod 2^52")
		f.VPXORQ(amd64.Z20, amd64.Z20, amd64.Z20, "clear Z20")
		f.VPMADD52LUQ(amd64.Z30, amd64.Z10, amd64.Z20, "Z20 = low52(T[0] * qInvNeg52)")
		f.VPANDQ(amd64.Z31, amd64.Z20, amd64.Z20, "mask to 52 bits (m in Z20)")

		// Step 4: T += m * q
		f.Comment("T += m * q")
		for j := 0; j < 5; j++ {
			qj := fmt.Sprintf("Z%d", j+25) // q[j] is in Z25+j
			tLow := fmt.Sprintf("Z%d", j+10)
			tHigh := fmt.Sprintf("Z%d", j+11)
			f.VPMADD52LUQ(qj, amd64.Z20, tLow)
			f.VPMADD52HUQ(qj, amd64.Z20, tHigh)
		}

		// Step 5: Shift right - T[0] is now 0 (mod 2^52), discard it
		// T[j] = T[j+1] for j = 0..4
		f.Comment("Shift: T[j] = T[j+1]")
		f.VPSRLQ("$52", amd64.Z10, amd64.Z20, "carry from T[0] (should be the only content)")
		f.VPADDQ(amd64.Z20, amd64.Z11, amd64.Z10, "T[0] = T[1] + carry")
		f.VMOVDQA64(amd64.Z12, amd64.Z11, "T[1] = T[2]")
		f.VMOVDQA64(amd64.Z13, amd64.Z12, "T[2] = T[3]")
		f.VMOVDQA64(amd64.Z14, amd64.Z13, "T[3] = T[4]")
		f.VMOVDQA64(amd64.Z15, amd64.Z14, "T[4] = T[5]")
		f.VPXORQ(amd64.Z15, amd64.Z15, amd64.Z15, "T[5] = 0")
	}

	// Result is in T[0..4] (Z10-Z14), copy to Z0-Z4
	// FUSED: Copy + normalization + x16 in one pass
	// Instead of: copy Z10-Z14 -> Z0-Z4, normalize, then x16
	// We do: shift Z10-Z14 by 4 directly, then normalize (handles both Montgomery and x16 carries)
	// This saves 5 VMOVDQA64 + some redundant operations
	f.Comment("Fused: x16 shift + normalization in one pass")

	// Step 1: Shift by 4 (x16) directly from Z10-Z14 to Z0-Z4
	f.VPSLLQ("$4", amd64.Z10, amd64.Z0, "Z0 = T[0] << 4")
	f.VPSLLQ("$4", amd64.Z11, amd64.Z1, "Z1 = T[1] << 4")
	f.VPSLLQ("$4", amd64.Z12, amd64.Z2, "Z2 = T[2] << 4")
	f.VPSLLQ("$4", amd64.Z13, amd64.Z3, "Z3 = T[3] << 4")
	f.VPSLLQ("$4", amd64.Z14, amd64.Z4, "Z4 = T[4] << 4")

	// Step 2: Extract all carries in parallel (up to 12 bits each: 8 from Mont + 4 from x16)
	f.VPSRLQ("$52", amd64.Z0, amd64.Z20, "carry0")
	f.VPSRLQ("$52", amd64.Z1, amd64.Z21, "carry1")
	f.VPSRLQ("$52", amd64.Z2, amd64.Z22, "carry2")
	f.VPSRLQ("$52", amd64.Z3, amd64.Z23, "carry3")

	// Step 3: Mask all limbs in parallel
	f.VPANDQ(amd64.Z31, amd64.Z0, amd64.Z0)
	f.VPANDQ(amd64.Z31, amd64.Z1, amd64.Z1)
	f.VPANDQ(amd64.Z31, amd64.Z2, amd64.Z2)
	f.VPANDQ(amd64.Z31, amd64.Z3, amd64.Z3)
	// Z4 doesn't need masking (has headroom for up to 52+12=64 bits)

	// Step 4: Add all carries to next limbs in parallel
	f.VPADDQ(amd64.Z20, amd64.Z1, amd64.Z1)
	f.VPADDQ(amd64.Z21, amd64.Z2, amd64.Z2)
	f.VPADDQ(amd64.Z22, amd64.Z3, amd64.Z3)
	f.VPADDQ(amd64.Z23, amd64.Z4, amd64.Z4)

	// AMM: Result is in [0, 32q), skip conditional subtraction - Barrett handles it
	f.Comment("AMM: result in [0, 32q) after x16, Barrett reduction follows")
}

func (f *FFAmd64) conditionalSubtractQ() {
	// Compare result with q and subtract if >= q
	// Result in Z0-Z4 (radix-52), q in Z25-Z29 (radix-52)
	// Z31 contains the 52-bit mask
	//
	// Algorithm:
	// 1. Compute limb-wise subtraction result - q
	// 2. Propagate borrows through the limbs
	// 3. If final borrow occurred (result < q), keep original; else use subtracted
	//
	// Key insight: In radix-52 with 64-bit registers, when we do VPSUBQ:
	// - If limb[i] >= q[i], result is correct (non-negative, fits in 52 bits)
	// - If limb[i] < q[i], result wraps to 2^64 + limb[i] - q[i]
	//   The low 52 bits are (limb[i] - q[i] + 2^52) mod 2^52 which is correct
	//   after accounting for the borrow from the next limb
	// So we MUST mask to 52 bits before using the result.

	// Compute result - q into Z10-Z14
	f.VPSUBQ(amd64.Z25, amd64.Z0, amd64.Z10)
	f.VPSUBQ(amd64.Z26, amd64.Z1, amd64.Z11)
	f.VPSUBQ(amd64.Z27, amd64.Z2, amd64.Z12)
	f.VPSUBQ(amd64.Z28, amd64.Z3, amd64.Z13)
	f.VPSUBQ(amd64.Z29, amd64.Z4, amd64.Z14)

	// Propagate borrows through limbs
	// If Z10 is negative (borrow from limb 0), subtract 1 from Z11
	f.VPSRAQ("$63", amd64.Z10, amd64.Z20, "Z20 = -1 if borrow, 0 otherwise")
	f.VPADDQ(amd64.Z20, amd64.Z11, amd64.Z11, "Z11 -= borrow")

	f.VPSRAQ("$63", amd64.Z11, amd64.Z20)
	f.VPADDQ(amd64.Z20, amd64.Z12, amd64.Z12)

	f.VPSRAQ("$63", amd64.Z12, amd64.Z20)
	f.VPADDQ(amd64.Z20, amd64.Z13, amd64.Z13)

	f.VPSRAQ("$63", amd64.Z13, amd64.Z20)
	f.VPADDQ(amd64.Z20, amd64.Z14, amd64.Z14)

	// Z14's sign bit tells us if result < q (borrow occurred)
	f.VPSRAQ("$63", amd64.Z14, amd64.Z20, "Z20 = all 1s if borrow (result < q), all 0s if no borrow")

	// Mask the subtracted limbs to 52 bits before selection
	// This is necessary because underflowed limbs have garbage in bits 52-63
	f.VPANDQ(amd64.Z31, amd64.Z10, amd64.Z10)
	f.VPANDQ(amd64.Z31, amd64.Z11, amd64.Z11)
	f.VPANDQ(amd64.Z31, amd64.Z12, amd64.Z12)
	f.VPANDQ(amd64.Z31, amd64.Z13, amd64.Z13)
	f.VPANDQ(amd64.Z31, amd64.Z14, amd64.Z14)

	// Select: if borrow (Z20 = all 1s), keep original; else use subtracted
	// Using VPTERNLOGQ: dst = f(dst, src1, src2) where imm8 is truth table
	// For VPTERNLOGQ $imm, Z10, Z20, Z0:
	//   dst=Z0 (original), src1=Z20 (mask), src2=Z10 (subtracted)
	// We want: if mask then original else subtracted
	// Truth table index = dst*4 + src1*2 + src2
	// imm8 = 0xE2 = 0b11100010: selects dst when src1=1, src2 when src1=0
	f.Comment("Conditional select using VPTERNLOGQ (saves 10 instructions)")
	f.VPTERNLOGQ("$0xE2", amd64.Z10, amd64.Z20, amd64.Z0)
	f.VPTERNLOGQ("$0xE2", amd64.Z11, amd64.Z20, amd64.Z1)
	f.VPTERNLOGQ("$0xE2", amd64.Z12, amd64.Z20, amd64.Z2)
	f.VPTERNLOGQ("$0xE2", amd64.Z13, amd64.Z20, amd64.Z3)
	f.VPTERNLOGQ("$0xE2", amd64.Z14, amd64.Z20, amd64.Z4)
}

func (f *FFAmd64) barrettReduction() {
	// Barrett reduction from [0, 32q) to [0, q) using single quotient estimation:
	// 1. k = (l4 * mu) >> 58, where mu = 0x36d9 (precomputed for this field)
	// 2. Subtract k*q from result (k is at most 31)
	// 3. One final conditional subtraction to handle rounding error
	//
	// This replaces 5 conditional subtractions (~175 instructions) with
	// 1 multiply + 5 multiplies (for k*q) + 1 conditional subtract (~30 instructions)

	f.Comment("Barrett reduction: k = (l4 * mu) >> 58, subtract k*q, then conditional subtract q")

	// Load Barrett constant mu (field-specific, defined in element.go as muBarrett52)
	// CRITICAL: VPMULUDQ only uses the low 32 bits of each operand, but l4 can be up to 50 bits
	// after x16. We must pre-shift l4 to fit in 32 bits.
	//
	// Original formula: k = (l4 * mu) >> 58
	// Since l4 can be up to 50 bits, we compute: k = ((l4 >> 20) * mu) >> 38
	// This ensures (l4 >> 20) fits in 30 bits, making VPMULUDQ valid.
	f.MOVQ("$const_muBarrett52", amd64.AX, "Barrett mu constant (field-specific)")
	f.VPBROADCASTQ(amd64.AX, amd64.Z5, "Z5 = mu broadcast")

	// Pre-shift l4 to fit in 32 bits for VPMULUDQ
	f.VPSRLQ("$20", amd64.Z4, amd64.Z6, "Z6 = l4 >> 20 (fits in 30 bits)")

	// Compute k = ((l4 >> 20) * mu) >> 38
	f.VPMULUDQ(amd64.Z5, amd64.Z6, amd64.Z5, "Z5 = (l4 >> 20) * mu")
	f.VPSRLQ("$38", amd64.Z5, amd64.Z5, "Z5 = k = ((l4 >> 20) * mu) >> 38")

	// Compute k*q and subtract from result
	// k*q needs 5 limb multiplications: k * q[i] for i=0..4
	// CRITICAL: VPMULUDQ only uses low 32 bits, but q[i] is 52 bits!
	// Use VPMADD52LUQ/HUQ which properly handle 52-bit operands.
	// k is at most 31 (5 bits), q[i] is at most 52 bits, so k*q[i] < 2^57.
	f.Comment("Compute k*q using VPMADD52 (handles 52-bit operands correctly)")

	// Use VPMADD52LUQ with zero accumulator to get k * q[i]
	// VPMADD52LUQ computes: dst += low52(src1 * src2)
	f.VPXORQ(amd64.Z6, amd64.Z6, amd64.Z6, "Z6 = 0")
	f.VPXORQ(amd64.Z7, amd64.Z7, amd64.Z7, "Z7 = 0")
	f.VPXORQ(amd64.Z8, amd64.Z8, amd64.Z8, "Z8 = 0")
	f.VPXORQ(amd64.Z9, amd64.Z9, amd64.Z9, "Z9 = 0")
	f.VPXORQ(amd64.Z10, amd64.Z10, amd64.Z10, "Z10 = 0")
	f.VPXORQ(amd64.Z15, amd64.Z15, amd64.Z15, "Z15 = 0 (for high parts)")

	// Compute k*q[i] low 52 bits
	f.VPMADD52LUQ(amd64.Z25, amd64.Z5, amd64.Z6, "Z6 = low52(k * q[0])")
	f.VPMADD52LUQ(amd64.Z26, amd64.Z5, amd64.Z7, "Z7 = low52(k * q[1])")
	f.VPMADD52LUQ(amd64.Z27, amd64.Z5, amd64.Z8, "Z8 = low52(k * q[2])")
	f.VPMADD52LUQ(amd64.Z28, amd64.Z5, amd64.Z9, "Z9 = low52(k * q[3])")
	f.VPMADD52LUQ(amd64.Z29, amd64.Z5, amd64.Z10, "Z10 = low52(k * q[4])")

	// Compute k*q[i] high 52 bits (actually just the carry, ~5 bits max)
	// Use Z15, Z16, Z17, Z18, Z19 for high parts
	f.VPXORQ(amd64.Z16, amd64.Z16, amd64.Z16)
	f.VPXORQ(amd64.Z17, amd64.Z17, amd64.Z17)
	f.VPXORQ(amd64.Z18, amd64.Z18, amd64.Z18)
	f.VPXORQ(amd64.Z19, amd64.Z19, amd64.Z19)

	f.VPMADD52HUQ(amd64.Z25, amd64.Z5, amd64.Z15, "Z15 = high52(k * q[0])")
	f.VPMADD52HUQ(amd64.Z26, amd64.Z5, amd64.Z16, "Z16 = high52(k * q[1])")
	f.VPMADD52HUQ(amd64.Z27, amd64.Z5, amd64.Z17, "Z17 = high52(k * q[2])")
	f.VPMADD52HUQ(amd64.Z28, amd64.Z5, amd64.Z18, "Z18 = high52(k * q[3])")
	f.VPMADD52HUQ(amd64.Z29, amd64.Z5, amd64.Z19, "Z19 = high52(k * q[4])")

	// Subtract k*q from result with carry propagation
	// k*q[i] = Z[6+i] + Z[15+i] * 2^52
	// We need to do: result[i] -= k*q[i]_low, then propagate high part as carry
	f.Comment("Subtract k*q with carry propagation")

	// Subtract low parts
	f.VPSUBQ(amd64.Z6, amd64.Z0, amd64.Z0, "Z0 -= k*q[0]_low")
	f.VPSUBQ(amd64.Z7, amd64.Z1, amd64.Z1, "Z1 -= k*q[1]_low")
	f.VPSUBQ(amd64.Z8, amd64.Z2, amd64.Z2, "Z2 -= k*q[2]_low")
	f.VPSUBQ(amd64.Z9, amd64.Z3, amd64.Z3, "Z3 -= k*q[3]_low")
	f.VPSUBQ(amd64.Z10, amd64.Z4, amd64.Z4, "Z4 -= k*q[4]_low")

	// Subtract high parts (carries) from next limbs
	f.VPSUBQ(amd64.Z15, amd64.Z1, amd64.Z1, "Z1 -= carry from k*q[0]")
	f.VPSUBQ(amd64.Z16, amd64.Z2, amd64.Z2, "Z2 -= carry from k*q[1]")
	f.VPSUBQ(amd64.Z17, amd64.Z3, amd64.Z3, "Z3 -= carry from k*q[2]")
	f.VPSUBQ(amd64.Z18, amd64.Z4, amd64.Z4, "Z4 -= carry from k*q[3]")
	// Note: Z19 (carry from k*q[4]) should be 0 since k*q[4] < 2^52 for valid inputs

	// Now propagate borrows through the result limbs
	// If result[i] underflowed (negative), we need to borrow from result[i+1]
	f.Comment("Propagate borrows through result")
	f.VPSRAQ("$63", amd64.Z0, amd64.Z15, "Z15 = -1 if Z0 underflowed, 0 otherwise")
	f.VPANDQ(amd64.Z31, amd64.Z0, amd64.Z0, "Z0 &= mask52")
	f.VPADDQ(amd64.Z15, amd64.Z1, amd64.Z1, "Z1 += borrow (borrow is -1 or 0)")

	f.VPSRAQ("$63", amd64.Z1, amd64.Z15)
	f.VPANDQ(amd64.Z31, amd64.Z1, amd64.Z1)
	f.VPADDQ(amd64.Z15, amd64.Z2, amd64.Z2)

	f.VPSRAQ("$63", amd64.Z2, amd64.Z15)
	f.VPANDQ(amd64.Z31, amd64.Z2, amd64.Z2)
	f.VPADDQ(amd64.Z15, amd64.Z3, amd64.Z3)

	f.VPSRAQ("$63", amd64.Z3, amd64.Z15)
	f.VPANDQ(amd64.Z31, amd64.Z3, amd64.Z3)
	f.VPADDQ(amd64.Z15, amd64.Z4, amd64.Z4)

	f.VPANDQ(amd64.Z31, amd64.Z4, amd64.Z4, "Z4 &= mask52")

	// Result is now in [0, 2q) due to Barrett rounding error
	// One final conditional subtraction of q to get result in [0, q)
	f.Comment("Final conditional subtraction of q")
	f.conditionalSubtractQ()
}

func (f *FFAmd64) convertFromRadix52(l0, l1, l2, l3, l4, a0, a1, a2, a3 string) {
	// Convert from radix-52 (l0-l4) to radix-64 (a0-a3)
	// Same as first part of convertAndStoreRadix64 but outputs to specified registers

	f.Comment("Convert from radix-52 to radix-64")

	// a0 = l0 | (l1 << 52)
	f.VPSLLQ("$52", l1, "Z18")
	f.VPORQ("Z18", l0, a0)

	// a1 = (l1 >> 12) | (l2 << 40)
	f.VPSRLQ("$12", l1, "Z18")
	f.VPSLLQ("$40", l2, "Z19")
	f.VPORQ("Z19", "Z18", a1)

	// a2 = (l2 >> 24) | (l3 << 28)
	f.VPSRLQ("$24", l2, "Z18")
	f.VPSLLQ("$28", l3, "Z19")
	f.VPORQ("Z19", "Z18", a2)

	// a3 = (l3 >> 36) | (l4 << 16)
	f.VPSRLQ("$36", l3, "Z18")
	f.VPSLLQ("$16", l4, "Z19")
	f.VPORQ("Z19", "Z18", a3)
}

func (f *FFAmd64) transposeAndStore(addr amd64.Register) {
	// Transpose from SoA (Z14-Z17) to AoS format and store
	// Z14 = [a0[0], a0[1], ..., a0[7]]
	// Z15 = [a1[0], a1[1], ..., a1[7]]
	// Z16 = [a2[0], a2[1], ..., a2[7]]
	// Z17 = [a3[0], a3[1], ..., a3[7]]

	f.transposeFromIFMA("Z14", "Z15", "Z16", "Z17", "Z10", "Z11", "Z12", "Z13")

	f.VMOVDQU64("Z10", fmt.Sprintf("0(%s)", addr))
	f.VMOVDQU64("Z11", fmt.Sprintf("64(%s)", addr))
	f.VMOVDQU64("Z12", fmt.Sprintf("128(%s)", addr))
	f.VMOVDQU64("Z13", fmt.Sprintf("192(%s)", addr))
}

func (f *FFAmd64) transposeFromIFMA(in0, in1, in2, in3, out0, out1, out2, out3 string) {
	// Reverse transpose from SoA (Structure of Arrays) back to AoS (Array of Structures)
	// Input: in0 = [e0.a0, e1.a0, e2.a0, e3.a0, e4.a0, e5.a0, e6.a0, e7.a0]
	//        in1 = [e0.a1, e1.a1, e2.a1, e3.a1, e4.a1, e5.a1, e6.a1, e7.a1]
	//        in2 = [e0.a2, e1.a2, e2.a2, e3.a2, e4.a2, e5.a2, e6.a2, e7.a2]
	//        in3 = [e0.a3, e1.a3, e2.a3, e3.a3, e4.a3, e5.a3, e6.a3, e7.a3]
	// Output: out0 = [e0.a0, e0.a1, e0.a2, e0.a3, e1.a0, e1.a1, e1.a2, e1.a3]
	//         out1 = [e2.a0, e2.a1, e2.a2, e2.a3, e3.a0, e3.a1, e3.a2, e3.a3]
	//         out2 = [e4.a0, e4.a1, e4.a2, e4.a3, e5.a0, e5.a1, e5.a2, e5.a3]
	//         out3 = [e6.a0, e6.a1, e6.a2, e6.a3, e7.a0, e7.a1, e7.a2, e7.a3]

	f.WriteLn("// 4x8 reverse transpose (SoA to AoS)")

	// Step 1: Pre-permute inputs to account for VPUNPCKLQDQ pairing behavior
	// VPUNPCKLQDQ pairs elements at indices (0,2), (1,3) within each 256-bit half
	// We need to reorder inputs so consecutive elements get paired correctly
	// Permute: [0,2,1,3,4,6,5,7] -> after VPUNPCKLQDQ we get correct pairing
	// Load permutation index (precomputed constant ·permuteIdxIFMA<>)
	f.VMOVDQU64("·permuteIdxIFMA<>(SB)", "Z22")
	f.VPERMQ(in0, "Z22", in0)
	f.VPERMQ(in1, "Z22", in1)
	f.VPERMQ(in2, "Z22", in2)
	f.VPERMQ(in3, "Z22", in3)

	// Step 2: Pair a0 with a1 and a2 with a3 using VPUNPCKLQDQ/VPUNPCKHQDQ
	f.VPUNPCKLQDQ(in1, in0, "Z18", "pairs (a0,a1) for elements 0,1,4,5")
	f.VPUNPCKHQDQ(in1, in0, "Z19", "pairs (a0,a1) for elements 2,3,6,7")
	f.VPUNPCKLQDQ(in3, in2, "Z20", "pairs (a2,a3) for elements 0,1,4,5")
	f.VPUNPCKHQDQ(in3, in2, "Z21", "pairs (a2,a3) for elements 2,3,6,7")

	// Step 3: Combine (a0,a1) with (a2,a3) to get complete 4-limb elements
	// VSHUFI64X2 $0x44 takes lanes 0,1 from both sources
	// VSHUFI64X2 $0xEE takes lanes 2,3 from both sources
	f.VSHUFI64X2("$0x44", "Z20", "Z18", out0)
	f.VSHUFI64X2("$0x44", "Z21", "Z19", out1)
	f.VSHUFI64X2("$0xEE", "Z20", "Z18", out2)
	f.VSHUFI64X2("$0xEE", "Z21", "Z19", out3)

	// Step 4: Fix lane ordering with VSHUFI64X2 $0xD8 to swap lanes 1 and 2
	f.VSHUFI64X2("$0xD8", out0, out0, out0)
	f.VSHUFI64X2("$0xD8", out1, out1, out1)
	f.VSHUFI64X2("$0xD8", out2, out2, out2)
	f.VSHUFI64X2("$0xD8", out3, out3, out3)
}
