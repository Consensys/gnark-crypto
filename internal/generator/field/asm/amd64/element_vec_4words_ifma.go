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

	// Precomputed multiples of q in radix-52 for binary reduction
	// These are computed as: to_radix52(N * q) for N = 2, 4, 8, 16
	// q = 8444461749428370424248824938781546531375899335154063827935233455917409239041

	f.Comment("2q in radix-52: used for binary reduction after x16 correction")
	f.DATA("·q2Radix52<>", 0, 8, "$0x3000000000002", "2q[0]")
	f.DATA("·q2Radix52<>", 8, 8, "$0xfda0000002142", "2q[1]")
	f.DATA("·q2Radix52<>", 16, 8, "$0x86f6002b354ed", "2q[2]")
	f.DATA("·q2Radix52<>", 24, 8, "$0x4aacc1689a3cb", "2q[3]")
	f.DATA("·q2Radix52<>", 32, 8, "$0x02556cabd3459", "2q[4]")
	f.GLOBL("·q2Radix52<>", "RODATA|NOPTR", 40)
	f.WriteLn("")

	f.Comment("4q in radix-52")
	f.DATA("·q4Radix52<>", 0, 8, "$0x6000000000004", "4q[0]")
	f.DATA("·q4Radix52<>", 8, 8, "$0xfb40000004284", "4q[1]")
	f.DATA("·q4Radix52<>", 16, 8, "$0x0dec00566a9db", "4q[2]")
	f.DATA("·q4Radix52<>", 24, 8, "$0x955982d134797", "4q[3]")
	f.DATA("·q4Radix52<>", 32, 8, "$0x04aad957a68b2", "4q[4]")
	f.GLOBL("·q4Radix52<>", "RODATA|NOPTR", 40)
	f.WriteLn("")

	f.Comment("8q in radix-52")
	f.DATA("·q8Radix52<>", 0, 8, "$0xc000000000008", "8q[0]")
	f.DATA("·q8Radix52<>", 8, 8, "$0xf680000008508", "8q[1]")
	f.DATA("·q8Radix52<>", 16, 8, "$0x1bd800acd53b7", "8q[2]")
	f.DATA("·q8Radix52<>", 24, 8, "$0x2ab305a268f2e", "8q[3]")
	f.DATA("·q8Radix52<>", 32, 8, "$0x0955b2af4d165", "8q[4]")
	f.GLOBL("·q8Radix52<>", "RODATA|NOPTR", 40)
	f.WriteLn("")

	f.Comment("16q in radix-52")
	f.DATA("·q16Radix52<>", 0, 8, "$0x8000000000010", "16q[0]")
	f.DATA("·q16Radix52<>", 8, 8, "$0xed00000010a11", "16q[1]")
	f.DATA("·q16Radix52<>", 16, 8, "$0x37b00159aa76f", "16q[2]")
	f.DATA("·q16Radix52<>", 24, 8, "$0x55660b44d1e5c", "16q[3]")
	f.DATA("·q16Radix52<>", 32, 8, "$0x12ab655e9a2ca", "16q[4]")
	f.GLOBL("·q16Radix52<>", "RODATA|NOPTR", 40)
	f.WriteLn("")
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
	// Correction: multiply by 16 to account for radix difference
	// IFMA uses R=2^260 (5 rounds * 52 bits), but input is in R=2^256 form
	// So result is A*B*2^{-260}, we need A*B*2^{-256}, difference = 2^4 = 16
	// We do this in radix-52 format where l4 has headroom (max ~2^48 after mult)
	f.Comment("Multiply by 16 in radix-52 to correct for radix-260 vs radix-256")
	f.multiplyByConstant16Radix52()

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
	f.Comment("Copy result to Z0-Z4")
	f.VMOVDQA64(amd64.Z10, amd64.Z0)
	f.VMOVDQA64(amd64.Z11, amd64.Z1)
	f.VMOVDQA64(amd64.Z12, amd64.Z2)
	f.VMOVDQA64(amd64.Z13, amd64.Z3)
	f.VMOVDQA64(amd64.Z14, amd64.Z4)

	// Final normalization (ensure all limbs < 2^52)
	f.Comment("Final normalization")
	for i := 0; i < 4; i++ {
		zi := fmt.Sprintf("Z%d", i)
		ziNext := fmt.Sprintf("Z%d", i+1)
		f.VPSRLQ("$52", zi, amd64.Z20)
		f.VPANDQ(amd64.Z31, zi, zi)
		f.VPADDQ(amd64.Z20, ziNext, ziNext)
	}

	// Conditional subtraction of q if result >= q
	f.Comment("Conditional subtraction if >= q")
	f.conditionalSubtractQ()
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
	// For each limb: result = (original & mask) | (subtracted & ~mask)
	f.VPANDQ(amd64.Z20, amd64.Z0, amd64.Z0, "keep original if borrow")
	f.VPANDNQ(amd64.Z10, amd64.Z20, amd64.Z10, "keep subtracted if no borrow")
	f.VPORQ(amd64.Z10, amd64.Z0, amd64.Z0)

	f.VPANDQ(amd64.Z20, amd64.Z1, amd64.Z1)
	f.VPANDNQ(amd64.Z11, amd64.Z20, amd64.Z11)
	f.VPORQ(amd64.Z11, amd64.Z1, amd64.Z1)

	f.VPANDQ(amd64.Z20, amd64.Z2, amd64.Z2)
	f.VPANDNQ(amd64.Z12, amd64.Z20, amd64.Z12)
	f.VPORQ(amd64.Z12, amd64.Z2, amd64.Z2)

	f.VPANDQ(amd64.Z20, amd64.Z3, amd64.Z3)
	f.VPANDNQ(amd64.Z13, amd64.Z20, amd64.Z13)
	f.VPORQ(amd64.Z13, amd64.Z3, amd64.Z3)

	f.VPANDQ(amd64.Z20, amd64.Z4, amd64.Z4)
	f.VPANDNQ(amd64.Z14, amd64.Z20, amd64.Z14)
	f.VPORQ(amd64.Z14, amd64.Z4, amd64.Z4)
}

func (f *FFAmd64) multiplyByConstant16Radix52() {
	// Multiply radix-52 result (Z0-Z4) by 16 with carry propagation
	// This corrects for the 2^260 vs 2^256 Montgomery radix difference
	//
	// For each limb: new_li = (li << 4) + carry_from_lower
	// Since 16 = 2^4, this is a left shift by 4 with carry propagation
	//
	// For 253-bit field, l4 is at most ~2^44, so l4*16 = 2^48 < 2^52 (fits!)
	//
	// Algorithm:
	// tmp0 = l0 << 4;  l0' = tmp0 & mask52;  carry = tmp0 >> 52
	// tmp1 = (l1 << 4) + carry;  l1' = tmp1 & mask52;  carry = tmp1 >> 52
	// etc.
	//
	// Note: we use OR instead of ADD for (li << 4) + carry because
	// (li << 4) has bits 4+ set and carry has only bits 0-3 set (no overlap)

	f.Comment("Multiply by 16 = 2^4 (left shift with carry) in radix-52")

	// Process l0: shift left by 4, extract low 52 bits and carry
	f.VPSLLQ("$4", amd64.Z0, amd64.Z10, "Z10 = l0 << 4")
	f.VPANDQ(amd64.Z31, amd64.Z10, amd64.Z0, "Z0 = (l0 << 4) & mask52")
	f.VPSRLQ("$52", amd64.Z10, amd64.Z15, "Z15 = carry = (l0 << 4) >> 52")

	// Process l1: shift, add carry from l0, extract low 52 bits and new carry
	f.VPSLLQ("$4", amd64.Z1, amd64.Z10, "Z10 = l1 << 4")
	f.VPORQ(amd64.Z15, amd64.Z10, amd64.Z10, "Z10 = (l1 << 4) | carry (no overlap)")
	f.VPANDQ(amd64.Z31, amd64.Z10, amd64.Z1, "Z1 = result & mask52")
	f.VPSRLQ("$52", amd64.Z10, amd64.Z15, "Z15 = new carry")

	// Process l2
	f.VPSLLQ("$4", amd64.Z2, amd64.Z10)
	f.VPORQ(amd64.Z15, amd64.Z10, amd64.Z10)
	f.VPANDQ(amd64.Z31, amd64.Z10, amd64.Z2)
	f.VPSRLQ("$52", amd64.Z10, amd64.Z15)

	// Process l3
	f.VPSLLQ("$4", amd64.Z3, amd64.Z10)
	f.VPORQ(amd64.Z15, amd64.Z10, amd64.Z10)
	f.VPANDQ(amd64.Z31, amd64.Z10, amd64.Z3)
	f.VPSRLQ("$52", amd64.Z10, amd64.Z15)

	// Process l4 (no mask needed, l4 has headroom for the result)
	f.VPSLLQ("$4", amd64.Z4, amd64.Z10)
	f.VPORQ(amd64.Z15, amd64.Z10, amd64.Z4, "Z4 = (l4 << 4) | carry")

	// Now result in Z0-Z4 is 16 * original
	// After Montgomery multiply, result is in [0, 2q), so 16*result can be in [0, 32q)
	//
	// Binary reduction from [0, 32q) to [0, q) using precomputed multiples:
	// 1. If result >= 16q, subtract 16q → [0, 16q)
	// 2. If result >= 8q, subtract 8q → [0, 8q)
	// 3. If result >= 4q, subtract 4q → [0, 4q)
	// 4. If result >= 2q, subtract 2q → [0, 2q)
	// 5. If result >= q, subtract q → [0, q)
	//
	// This is 5 conditional subtractions instead of up to 32 sequential ones.
	// The constants 16q, 8q, 4q, 2q are precomputed in DATA sections.

	f.Comment("Binary reduction: conditionally subtract 16q, 8q, 4q, 2q, q")

	// Load 16q into Z5-Z9 and subtract
	f.VPBROADCASTQ("·q16Radix52<>+0(SB)", amd64.Z5)
	f.VPBROADCASTQ("·q16Radix52<>+8(SB)", amd64.Z6)
	f.VPBROADCASTQ("·q16Radix52<>+16(SB)", amd64.Z7)
	f.VPBROADCASTQ("·q16Radix52<>+24(SB)", amd64.Z8)
	f.VPBROADCASTQ("·q16Radix52<>+32(SB)", amd64.Z9)
	f.conditionalSubtractNQ("Z5", "Z6", "Z7", "Z8", "Z9")

	// Load 8q and subtract
	f.VPBROADCASTQ("·q8Radix52<>+0(SB)", amd64.Z5)
	f.VPBROADCASTQ("·q8Radix52<>+8(SB)", amd64.Z6)
	f.VPBROADCASTQ("·q8Radix52<>+16(SB)", amd64.Z7)
	f.VPBROADCASTQ("·q8Radix52<>+24(SB)", amd64.Z8)
	f.VPBROADCASTQ("·q8Radix52<>+32(SB)", amd64.Z9)
	f.conditionalSubtractNQ("Z5", "Z6", "Z7", "Z8", "Z9")

	// Load 4q and subtract
	f.VPBROADCASTQ("·q4Radix52<>+0(SB)", amd64.Z5)
	f.VPBROADCASTQ("·q4Radix52<>+8(SB)", amd64.Z6)
	f.VPBROADCASTQ("·q4Radix52<>+16(SB)", amd64.Z7)
	f.VPBROADCASTQ("·q4Radix52<>+24(SB)", amd64.Z8)
	f.VPBROADCASTQ("·q4Radix52<>+32(SB)", amd64.Z9)
	f.conditionalSubtractNQ("Z5", "Z6", "Z7", "Z8", "Z9")

	// Load 2q and subtract
	f.VPBROADCASTQ("·q2Radix52<>+0(SB)", amd64.Z5)
	f.VPBROADCASTQ("·q2Radix52<>+8(SB)", amd64.Z6)
	f.VPBROADCASTQ("·q2Radix52<>+16(SB)", amd64.Z7)
	f.VPBROADCASTQ("·q2Radix52<>+24(SB)", amd64.Z8)
	f.VPBROADCASTQ("·q2Radix52<>+32(SB)", amd64.Z9)
	f.conditionalSubtractNQ("Z5", "Z6", "Z7", "Z8", "Z9")

	// Final subtraction of q (using Z25-Z29 which already have q loaded)
	f.conditionalSubtractQ()
}

func (f *FFAmd64) conditionalSubtractNQ(nq0, nq1, nq2, nq3, nq4 string) {
	// Compare result with N*q and subtract if >= N*q
	// Result in Z0-Z4 (radix-52), N*q in specified registers
	// Z31 contains the 52-bit mask
	//
	// This is the same algorithm as conditionalSubtractQ but uses arbitrary N*q registers

	// Compute result - N*q into Z10-Z14
	f.VPSUBQ(nq0, "Z0", "Z10")
	f.VPSUBQ(nq1, "Z1", "Z11")
	f.VPSUBQ(nq2, "Z2", "Z12")
	f.VPSUBQ(nq3, "Z3", "Z13")
	f.VPSUBQ(nq4, "Z4", "Z14")

	// Propagate borrows through limbs
	f.VPSRAQ("$63", "Z10", "Z20")
	f.VPADDQ("Z20", "Z11", "Z11")

	f.VPSRAQ("$63", "Z11", "Z20")
	f.VPADDQ("Z20", "Z12", "Z12")

	f.VPSRAQ("$63", "Z12", "Z20")
	f.VPADDQ("Z20", "Z13", "Z13")

	f.VPSRAQ("$63", "Z13", "Z20")
	f.VPADDQ("Z20", "Z14", "Z14")

	// Z14's sign bit tells us if result < N*q (borrow occurred)
	f.VPSRAQ("$63", "Z14", "Z20")

	// Mask the subtracted limbs to 52 bits before selection
	f.VPANDQ("Z31", "Z10", "Z10")
	f.VPANDQ("Z31", "Z11", "Z11")
	f.VPANDQ("Z31", "Z12", "Z12")
	f.VPANDQ("Z31", "Z13", "Z13")
	f.VPANDQ("Z31", "Z14", "Z14")

	// Select: if borrow (Z20 = all 1s), keep original; else use subtracted
	f.VPANDQ("Z20", "Z0", "Z0")
	f.VPANDNQ("Z10", "Z20", "Z10")
	f.VPORQ("Z10", "Z0", "Z0")

	f.VPANDQ("Z20", "Z1", "Z1")
	f.VPANDNQ("Z11", "Z20", "Z11")
	f.VPORQ("Z11", "Z1", "Z1")

	f.VPANDQ("Z20", "Z2", "Z2")
	f.VPANDNQ("Z12", "Z20", "Z12")
	f.VPORQ("Z12", "Z2", "Z2")

	f.VPANDQ("Z20", "Z3", "Z3")
	f.VPANDNQ("Z13", "Z20", "Z13")
	f.VPORQ("Z13", "Z3", "Z3")

	f.VPANDQ("Z20", "Z4", "Z4")
	f.VPANDNQ("Z14", "Z20", "Z14")
	f.VPORQ("Z14", "Z4", "Z4")
}

func (f *FFAmd64) convertAndStoreRadix64(addr amd64.Register, z0, z1, z2, z3, z4 string) {
	// Convert from radix-52 (Z0-Z4) back to radix-64 and store

	f.Comment("Convert from radix-52 to radix-64")

	// a0 = l0 | (l1 << 52)
	f.VPSLLQ("$52", z1, "Z18")
	f.VPORQ("Z18", z0, "Z14", "Z14 = a0")

	// a1 = (l1 >> 12) | (l2 << 40)
	f.VPSRLQ("$12", z1, "Z18")
	f.VPSLLQ("$40", z2, "Z19")
	f.VPORQ("Z19", "Z18", "Z15", "Z15 = a1")

	// a2 = (l2 >> 24) | (l3 << 28)
	f.VPSRLQ("$24", z2, "Z18")
	f.VPSLLQ("$28", z3, "Z19")
	f.VPORQ("Z19", "Z18", "Z16", "Z16 = a2")

	// a3 = (l3 >> 36) | (l4 << 16)
	f.VPSRLQ("$36", z3, "Z18")
	f.VPSLLQ("$16", z4, "Z19")
	f.VPORQ("Z19", "Z18", "Z17", "Z17 = a3")

	// Transpose back from SoA to AoS and store
	f.Comment("Transpose back to AoS format and store")
	f.transposeFromIFMA("Z14", "Z15", "Z16", "Z17", "Z10", "Z11", "Z12", "Z13")

	f.VMOVDQU64("Z10", fmt.Sprintf("0(%s)", addr))
	f.VMOVDQU64("Z11", fmt.Sprintf("64(%s)", addr))
	f.VMOVDQU64("Z12", fmt.Sprintf("128(%s)", addr))
	f.VMOVDQU64("Z13", fmt.Sprintf("192(%s)", addr))
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
