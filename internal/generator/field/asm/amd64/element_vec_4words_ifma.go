// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package amd64

import (
	"fmt"

	"github.com/consensys/bavard/amd64"
)

// ifmaHelper wraps FFAmd64 and caches Define macros for IFMA operations
type ifmaHelper struct {
	*FFAmd64

	// Cached define functions for reuse
	carryProp       defineFn // carry propagation (extract carry, mask, add to next)
	borrowProp      defineFn // borrow propagation (extract sign, add to next, mask)
	ifmaMulAccLH    defineFn // IFMA multiply-accumulate low+high
	toRadix52Limb   defineFn // convert one limb to radix-52
	fromRadix52Limb defineFn // convert one limb from radix-52
	condSelect      defineFn // conditional select using VPTERNLOGQ
	zeroReg         defineFn // zero a ZMM register
}

// newIfmaHelper creates a new IFMA helper with cached defines
func (f *FFAmd64) newIfmaHelper() *ifmaHelper {
	h := &ifmaHelper{FFAmd64: f}
	h.initDefines()
	return h
}

// initDefines initializes all the reusable Define macros
func (h *ifmaHelper) initDefines() {
	// CARRY_PROP: Extract carry from limb, mask it, add carry to next limb
	// args: src, dst_mask, dst_next, mask52, tmp
	h.carryProp = h.Define("CARRY_PROP", 5, func(args ...any) {
		src := args[0]
		dstMask := args[1]
		dstNext := args[2]
		mask52 := args[3]
		tmp := args[4]
		h.VPSRLQ("$52", src, tmp)
		h.VPANDQ(mask52, src, dstMask)
		h.VPADDQ(tmp, dstNext, dstNext)
	}, true)

	// BORROW_PROP: Extract borrow (sign bit), add to next, mask current
	// args: src, dst_next, mask52, tmp
	h.borrowProp = h.Define("BORROW_PROP", 4, func(args ...any) {
		src := args[0]
		dstNext := args[1]
		mask52 := args[2]
		tmp := args[3]
		h.VPSRAQ("$63", src, tmp)
		h.VPANDQ(mask52, src, src)
		h.VPADDQ(tmp, dstNext, dstNext)
	}, true)

	// IFMA_MUL_ACC_LH: IFMA multiply-accumulate both low and high parts
	// args: multiplier, multiplicand, acc_low, acc_high
	h.ifmaMulAccLH = h.Define("IFMA_MUL_ACC_LH", 4, func(args ...any) {
		multiplier := args[0]
		multiplicand := args[1]
		accLow := args[2]
		accHigh := args[3]
		h.VPMADD52LUQ(multiplier, multiplicand, accLow)
		h.VPMADD52HUQ(multiplier, multiplicand, accHigh)
	}, true)

	// COND_SELECT: Conditional select using VPTERNLOGQ
	// if mask == all 1s, keep dst; else use src
	// args: src, mask, dst
	h.condSelect = h.Define("COND_SELECT", 3, func(args ...any) {
		src := args[0]
		mask := args[1]
		dst := args[2]
		h.VPTERNLOGQ("$0xE2", src, mask, dst)
	}, true)

	// ZERO_REG: Zero a ZMM register
	h.zeroReg = h.Define("ZERO_REG", 1, func(args ...any) {
		reg := args[0]
		h.VPXORQ(reg, reg, reg)
	}, true)
}

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
	f.emitIFMAConstants()
	h := f.newIfmaHelper()
	h.generateMulVecIFMABody("mulVec", false)
}

// generateScalarMulVecIFMA generates IFMA-based scalar multiplication
// This is similar to mulVecIFMA but broadcasts the scalar to all lanes
func (f *FFAmd64) generateScalarMulVecIFMA() {
	h := f.newIfmaHelper()
	h.generateMulVecIFMABody("scalarMulVec", true)
}

// generateMulVecIFMABody is the common implementation for mul and scalarMul
func (h *ifmaHelper) generateMulVecIFMABody(funcName string, scalarMul bool) {
	if scalarMul {
		h.Comment(fmt.Sprintf("%s(res, a, b *Element, n uint64)", funcName))
		h.Comment("Performs n scalar multiplications using AVX-512 IFMA instructions")
	} else {
		h.Comment(fmt.Sprintf("%s(res, a, b *Element, n uint64)", funcName))
		h.Comment("Performs n multiplications using AVX-512 IFMA instructions")
	}
	h.Comment("Processes 8 elements in parallel using radix-52 representation")

	const argSize = 4 * 8
	stackSize := 0
	registers := h.FnHeader(funcName, stackSize, argSize, amd64.AX, amd64.DX)
	defer h.AssertCleanStack(stackSize, 0)

	addrRes := h.Pop(&registers)
	addrA := h.Pop(&registers)
	addrB := h.Pop(&registers)
	n := h.Pop(&registers)

	loop := h.NewLabel("loop")
	done := h.NewLabel("done")

	h.MOVQ("res+0(FP)", addrRes)
	h.MOVQ("a+8(FP)", addrA)
	h.MOVQ("b+16(FP)", addrB)
	h.MOVQ("n+24(FP)", n)

	// Initialize constants
	h.Comment("Load constants for radix-52 conversion and reduction")
	h.initIFMAConstants()

	h.LABEL(loop)
	h.TESTQ(n, n)
	h.JEQ(done, "n == 0, we are done")

	h.Comment("Process 8 elements in parallel")

	h.Comment("Load and convert 8 elements from a[] to radix-52")
	h.loadAndConvertToRadix52(addrA, "Z0", "Z1", "Z2", "Z3", "Z4")

	if scalarMul {
		// Reload scalar each iteration since Barrett reduction clobbers Z5-Z9
		h.Comment("Load scalar and convert to radix-52 (broadcast)")
		h.loadScalarToRadix52(addrB)
	} else {
		h.Comment("Load and convert 8 elements from b[] to radix-52")
		h.loadAndConvertToRadix52(addrB, "Z5", "Z6", "Z7", "Z8", "Z9")
	}

	h.Comment("Montgomery multiplication using IFMA (CIOS variant)")
	h.montgomeryMulIFMAWithDefines()

	h.Comment("Barrett reduction from [0, 32q) to [0, q)")
	h.barrettReductionWithDefines()

	h.Comment("Convert result from radix-52 back to radix-64")
	h.convertFromRadix52("Z0", "Z1", "Z2", "Z3", "Z4", "Z14", "Z15", "Z16", "Z17")

	h.Comment("Transpose back to AoS format and store")
	h.transposeAndStore(addrRes)

	h.Comment("Advance pointers")
	h.ADDQ("$256", addrA)
	if !scalarMul {
		h.ADDQ("$256", addrB)
	}
	h.ADDQ("$256", addrRes)
	h.DECQ(n, "processed 1 group of 8 elements")

	h.JMP(loop)

	h.LABEL(done)
	h.RET()

	h.Push(&registers, addrRes, addrA, addrB, n)
}

// initIFMAConstants initializes the constants needed for IFMA operations
func (h *ifmaHelper) initIFMAConstants() {
	h.MOVQ("$0xFFFFFFFFFFFFF", amd64.R15, "52-bit mask in R15")
	h.VPBROADCASTQ(amd64.R15, amd64.Z31, "Z31 = mask52")

	h.MOVQ("$const_qInvNeg", amd64.AX)
	h.ANDQ(amd64.R15, amd64.AX, "keep low 52 bits")
	h.VPBROADCASTQ(amd64.AX, amd64.Z30, "Z30 = qInvNeg52")

	h.Comment("Load modulus in radix-52 form")
	h.loadModulusRadix52()
}

// loadScalarToRadix52 loads a scalar and broadcasts it to all lanes in radix-52
func (h *ifmaHelper) loadScalarToRadix52(addr amd64.Register) {
	// Load the 4 words of the scalar
	h.MOVQ(fmt.Sprintf("0(%s)", addr), amd64.R9)
	h.MOVQ(fmt.Sprintf("8(%s)", addr), amd64.R10)
	h.MOVQ(fmt.Sprintf("16(%s)", addr), amd64.R11)
	h.MOVQ(fmt.Sprintf("24(%s)", addr), amd64.R12)

	// Convert to radix-52 and broadcast
	// l0 = a0 & mask52
	h.MOVQ(amd64.R9, amd64.R8)
	h.ANDQ(amd64.R15, amd64.R8)
	h.VPBROADCASTQ(amd64.R8, amd64.Z5)

	// l1 = (a0 >> 52) | (a1 << 12) & mask52
	h.SHRQ("$52", amd64.R9)
	h.MOVQ(amd64.R10, amd64.R8)
	h.SHLQ("$12", amd64.R8)
	h.ORQ(amd64.R9, amd64.R8)
	h.ANDQ(amd64.R15, amd64.R8)
	h.VPBROADCASTQ(amd64.R8, amd64.Z6)

	// l2 = (a1 >> 40) | (a2 << 24) & mask52
	h.SHRQ("$40", amd64.R10)
	h.MOVQ(amd64.R11, amd64.R8)
	h.SHLQ("$24", amd64.R8)
	h.ORQ(amd64.R10, amd64.R8)
	h.ANDQ(amd64.R15, amd64.R8)
	h.VPBROADCASTQ(amd64.R8, amd64.Z7)

	// l3 = (a2 >> 28) | (a3 << 36) & mask52
	h.SHRQ("$28", amd64.R11)
	h.MOVQ(amd64.R12, amd64.R8)
	h.SHLQ("$36", amd64.R8)
	h.ORQ(amd64.R11, amd64.R8)
	h.ANDQ(amd64.R15, amd64.R8)
	h.VPBROADCASTQ(amd64.R8, amd64.Z8)

	// l4 = a3 >> 16
	h.SHRQ("$16", amd64.R12)
	h.VPBROADCASTQ(amd64.R12, amd64.Z9)
}

func (f *FFAmd64) loadModulusRadix52() {
	f.Comment("q in radix-52: Z25=ql0, Z26=ql1, Z27=ql2, Z28=ql3, Z29=ql4")
	f.Comment("Load q0-q3 and convert to radix-52")
	f.MOVQ("$const_q0", amd64.R9)
	f.MOVQ("$const_q1", amd64.R10)
	f.MOVQ("$const_q2", amd64.R11)
	f.MOVQ("$const_q3", amd64.R12)

	// ql0 = q0 & mask52
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

// montgomeryMulIFMAWithDefines performs Montgomery multiplication using defines for compact output
func (h *ifmaHelper) montgomeryMulIFMAWithDefines() {
	h.Comment("A = [Z0-Z4], B = [Z5-Z9], result in [Z0-Z4]")

	// Initialize accumulators
	for i := 10; i <= 15; i++ {
		h.zeroReg(fmt.Sprintf("Z%d", i))
	}

	// Process each limb of B (CIOS rounds)
	for i := 0; i < 5; i++ {
		h.Comment(fmt.Sprintf("CIOS Round %d", i))
		h.ciosRound(i)
	}

	// Fused x16 shift + normalization
	h.Comment("Fused x16 shift + normalization")
	for i := 0; i < 5; i++ {
		h.VPSLLQ("$4", fmt.Sprintf("Z%d", i+10), fmt.Sprintf("Z%d", i))
	}

	// Extract carries in parallel
	for i := 0; i < 4; i++ {
		h.VPSRLQ("$52", fmt.Sprintf("Z%d", i), fmt.Sprintf("Z%d", i+20))
	}

	// Mask limbs
	for i := 0; i < 4; i++ {
		h.VPANDQ(amd64.Z31, fmt.Sprintf("Z%d", i), fmt.Sprintf("Z%d", i))
	}

	// Add carries
	for i := 0; i < 4; i++ {
		h.VPADDQ(fmt.Sprintf("Z%d", i+20), fmt.Sprintf("Z%d", i+1), fmt.Sprintf("Z%d", i+1))
	}
}

// ciosRound generates one round of the CIOS Montgomery multiplication
func (h *ifmaHelper) ciosRound(i int) {
	bi := fmt.Sprintf("Z%d", i+5)

	// T += A * B[i]
	for j := 0; j < 5; j++ {
		aj := fmt.Sprintf("Z%d", j)
		tLow := fmt.Sprintf("Z%d", j+10)
		tHigh := fmt.Sprintf("Z%d", j+11)
		h.VPMADD52LUQ(bi, aj, tLow)
		h.VPMADD52HUQ(bi, aj, tHigh)
	}

	// Normalize T[0]
	h.VPSRLQ("$52", amd64.Z10, amd64.Z20)
	h.VPANDQ(amd64.Z31, amd64.Z10, amd64.Z10)
	h.VPADDQ(amd64.Z20, amd64.Z11, amd64.Z11)

	// m = T[0] * qInvNeg52 mod 2^52
	h.VPXORQ(amd64.Z20, amd64.Z20, amd64.Z20)
	h.VPMADD52LUQ(amd64.Z30, amd64.Z10, amd64.Z20)
	h.VPANDQ(amd64.Z31, amd64.Z20, amd64.Z20)

	// T += m * q
	for j := 0; j < 5; j++ {
		qj := fmt.Sprintf("Z%d", j+25)
		tLow := fmt.Sprintf("Z%d", j+10)
		tHigh := fmt.Sprintf("Z%d", j+11)
		h.VPMADD52LUQ(qj, amd64.Z20, tLow)
		h.VPMADD52HUQ(qj, amd64.Z20, tHigh)
	}

	// Shift: T[j] = T[j+1]
	h.VPSRLQ("$52", amd64.Z10, amd64.Z20)
	h.VPADDQ(amd64.Z20, amd64.Z11, amd64.Z10)
	h.VMOVDQA64(amd64.Z12, amd64.Z11)
	h.VMOVDQA64(amd64.Z13, amd64.Z12)
	h.VMOVDQA64(amd64.Z14, amd64.Z13)
	h.VMOVDQA64(amd64.Z15, amd64.Z14)
	h.VPXORQ(amd64.Z15, amd64.Z15, amd64.Z15)
}

// barrettReductionWithDefines performs Barrett reduction using defines
func (h *ifmaHelper) barrettReductionWithDefines() {
	h.Comment("k = (l4 * mu) >> 58, subtract k*q, then conditional subtract q")

	// Load Barrett constant and compute k
	h.MOVQ("$const_muBarrett52", amd64.AX)
	h.VPBROADCASTQ(amd64.AX, amd64.Z5)
	h.VPSRLQ("$20", amd64.Z4, amd64.Z6)
	h.VPMULUDQ(amd64.Z5, amd64.Z6, amd64.Z5)
	h.VPSRLQ("$38", amd64.Z5, amd64.Z5)

	// Compute k*q using VPMADD52
	h.Comment("k*q using VPMADD52")
	for i := 6; i <= 10; i++ {
		h.zeroReg(fmt.Sprintf("Z%d", i))
	}
	for i := 15; i <= 19; i++ {
		h.zeroReg(fmt.Sprintf("Z%d", i))
	}

	// Low parts
	for i := 0; i < 5; i++ {
		h.VPMADD52LUQ(fmt.Sprintf("Z%d", i+25), amd64.Z5, fmt.Sprintf("Z%d", i+6))
	}

	// High parts
	for i := 0; i < 5; i++ {
		h.VPMADD52HUQ(fmt.Sprintf("Z%d", i+25), amd64.Z5, fmt.Sprintf("Z%d", i+15))
	}

	// Subtract k*q from result
	h.Comment("Subtract k*q")
	for i := 0; i < 5; i++ {
		h.VPSUBQ(fmt.Sprintf("Z%d", i+6), fmt.Sprintf("Z%d", i), fmt.Sprintf("Z%d", i))
	}

	// Subtract carries
	for i := 0; i < 4; i++ {
		h.VPSUBQ(fmt.Sprintf("Z%d", i+15), fmt.Sprintf("Z%d", i+1), fmt.Sprintf("Z%d", i+1))
	}

	// Propagate borrows
	h.Comment("Propagate borrows")
	for i := 0; i < 4; i++ {
		h.borrowProp(fmt.Sprintf("Z%d", i), fmt.Sprintf("Z%d", i+1), amd64.Z31, amd64.Z15)
	}
	h.VPANDQ(amd64.Z31, amd64.Z4, amd64.Z4)

	// Final conditional subtraction
	h.Comment("Final conditional subtraction of q")
	h.conditionalSubtractQWithDefines()
}

// conditionalSubtractQWithDefines performs conditional subtraction using defines
func (h *ifmaHelper) conditionalSubtractQWithDefines() {
	// Compute result - q
	for i := 0; i < 5; i++ {
		h.VPSUBQ(fmt.Sprintf("Z%d", i+25), fmt.Sprintf("Z%d", i), fmt.Sprintf("Z%d", i+10))
	}

	// Propagate borrows
	for i := 0; i < 4; i++ {
		h.VPSRAQ("$63", fmt.Sprintf("Z%d", i+10), amd64.Z20)
		h.VPADDQ(amd64.Z20, fmt.Sprintf("Z%d", i+11), fmt.Sprintf("Z%d", i+11))
	}

	// Get final borrow mask
	h.VPSRAQ("$63", amd64.Z14, amd64.Z20)

	// Mask subtracted limbs
	for i := 0; i < 5; i++ {
		h.VPANDQ(amd64.Z31, fmt.Sprintf("Z%d", i+10), fmt.Sprintf("Z%d", i+10))
	}

	// Conditional select
	for i := 0; i < 5; i++ {
		h.condSelect(fmt.Sprintf("Z%d", i+10), amd64.Z20, fmt.Sprintf("Z%d", i))
	}
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
