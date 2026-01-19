// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package amd64

import (
	"fmt"

	"github.com/consensys/bavard/amd64"
)

// addVec res = a + b
// func addVec(res, a, b *{{.ElementName}}, n uint64)
func (f *FFAmd64) generateAddVecW4() {
	f.Comment("addVec(res, a, b *Element, n uint64) res[0...n] = a[0...n] + b[0...n]")

	const argSize = 4 * 8
	stackSize := f.StackSize(f.NbWords*2+4, 0, 0)
	registers := f.FnHeader("addVec", stackSize, argSize)
	defer f.AssertCleanStack(stackSize, 0)

	// registers & labels we need
	addrA := f.Pop(&registers)
	addrB := f.Pop(&registers)
	addrRes := f.Pop(&registers)
	len := f.Pop(&registers)

	a := f.PopN(&registers)
	t := f.PopN(&registers)

	loop := f.NewLabel("loop")
	done := f.NewLabel("done")

	// load arguments
	f.MOVQ("res+0(FP)", addrRes)
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("b+16(FP)", addrB)
	f.MOVQ("n+24(FP)", len)

	f.LABEL(loop)

	f.TESTQ(len, len)
	f.JEQ(done, "n == 0, we are done")

	// a = a + b
	f.LabelRegisters("a", a...)
	f.Mov(addrA, a)
	f.Add(addrB, a)
	f.PREFETCHT0(fmt.Sprintf("2048(%s)", addrA))
	f.PREFETCHT0(fmt.Sprintf("2048(%s)", addrB))

	// reduce a
	f.ReduceElement(a, t, false)

	// save a into res
	f.Mov(a, addrRes)
	f.Comment("increment pointers to visit next element")
	f.ADDQ("$32", addrA)
	f.ADDQ("$32", addrB)
	f.ADDQ("$32", addrRes)
	f.DECQ(len, "decrement n")
	f.JMP(loop)

	f.LABEL(done)

	f.RET()

	f.Push(&registers, a...)
	f.Push(&registers, t...)
	f.Push(&registers, addrA, addrB, addrRes, len)

}

// subVec res = a - b
// func subVec(res, a, b *{{.ElementName}}, n uint64)
func (f *FFAmd64) generateSubVecW4() {
	f.Comment("subVec(res, a, b *Element, n uint64) res[0...n] = a[0...n] - b[0...n]")

	const argSize = 4 * 8
	stackSize := f.StackSize(f.NbWords*2+5, 0, 0)
	registers := f.FnHeader("subVec", stackSize, argSize)
	defer f.AssertCleanStack(stackSize, 0)

	// registers
	addrA := f.Pop(&registers)
	addrB := f.Pop(&registers)
	addrRes := f.Pop(&registers)
	len := f.Pop(&registers)
	zero := f.Pop(&registers)

	a := f.PopN(&registers)
	q := f.PopN(&registers)

	loop := f.NewLabel("loop")
	done := f.NewLabel("done")

	// load arguments
	f.MOVQ("res+0(FP)", addrRes)
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("b+16(FP)", addrB)
	f.MOVQ("n+24(FP)", len)

	f.XORQ(zero, zero)

	f.LABEL(loop)

	f.TESTQ(len, len)
	f.JEQ(done, "n == 0, we are done")

	// a = a - b
	f.LabelRegisters("a", a...)
	f.Mov(addrA, a)
	f.Sub(addrB, a)
	f.PREFETCHT0(fmt.Sprintf("2048(%s)", addrA))
	f.PREFETCHT0(fmt.Sprintf("2048(%s)", addrB))

	// reduce a
	f.Comment("reduce (a-b) mod q")
	f.LabelRegisters("q", q...)
	for i := 0; i < f.NbWords; i++ {
		f.MOVQ(fmt.Sprintf("$const_q%d", i), q[i])
	}
	for i := 0; i < f.NbWords; i++ {
		f.CMOVQCC(zero, q[i])
	}
	// add registers (q or 0) to a, and set to result
	f.Comment("add registers (q or 0) to a, and set to result")
	f.Add(q, a)

	// save a into res
	f.Mov(a, addrRes)

	f.Comment("increment pointers to visit next element")
	f.ADDQ("$32", addrA)
	f.ADDQ("$32", addrB)
	f.ADDQ("$32", addrRes)
	f.DECQ(len, "decrement n")
	f.JMP(loop)

	f.LABEL(done)

	f.RET()

	f.Push(&registers, a...)
	f.Push(&registers, q...)
	f.Push(&registers, addrA, addrB, addrRes, len, zero)

}

// sumVec res = sum(a[0...n])
func (f *FFAmd64) generateSumVecW4() {
	f.Comment("sumVec(res *Element, a *Element, n uint64) res = sum(a[0...n])")
	f.Comment("")
	f.Comment("Algorithm: Accumulate elements into 8 qword accumulators (radix-2^32),")
	f.Comment("then normalize to radix-2^64 and reduce mod q.")
	f.Comment("")
	f.Comment("When we move an element into a Z register using VPMOVZXDQ,")
	f.Comment("each 32-bit dword is zero-extended to a 64-bit qword.")
	f.Comment("We can safely add up to 2^32 elements without overflow.")

	const argSize = 3 * 8
	// Need 64 bytes of stack for temporary storage (8 qwords)
	stackSize := f.StackSize(10, 2, 64)
	registers := f.FnHeader("sumVec", stackSize, argSize, amd64.DX, amd64.AX)
	defer f.AssertCleanStack(stackSize, 64) // We use 64 bytes directly for temp storage

	// registers & labels we need
	addrA := f.Pop(&registers)
	addrRes := f.Pop(&registers)
	n := f.Pop(&registers)
	nMod8 := f.Pop(&registers)

	done := f.NewLabel("done")

	// AVX512 registers
	acc := registers.PopVN(8)
	zvec := registers.PopVN(8)

	// load arguments
	f.MOVQ("res+0(FP)", addrRes)
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("n+16(FP)", n)

	f.Comment("initialize accumulators to zero")
	f.VXORPS(acc[0], acc[0], acc[0])
	for i := 1; i < 8; i++ {
		f.VMOVDQA64(acc[0], acc[i])
	}

	f.LabelRegisters("n % 8", nMod8)
	f.LabelRegisters("n / 8", n)
	f.MOVQ(n, nMod8)
	f.ANDQ("$7", nMod8) // nMod8 = n % 8
	f.SHRQ("$3", n)     // len = n / 8

	// handle n % 8 first
	f.Loop(nMod8, func() {
		f.VPMOVZXDQ("0("+addrA+")", zvec[0])
		f.VPADDQ(zvec[0], acc[0], acc[0])
		f.ADDQ("$32", addrA)
	})

	f.Loop(n, func() {
		for i := 0; i < 8; i++ {
			f.VPMOVZXDQ(addrA.At(4*i), zvec[i])
		}

		f.WriteLn(fmt.Sprintf("PREFETCHT0 4096(%[1]s)", addrA))
		for i := 0; i < 8; i++ {
			f.VPADDQ(zvec[i], acc[i], acc[i])
		}

		f.ADDQ("$256", addrA, "increment pointer by 8 elements")
	})

	f.Comment("accumulate the 8 Z registers into acc[0]")
	for i := 7; i > 0; i-- {
		f.VPADDQ(acc[i], acc[i-1], acc[i-1])
	}

	f.Comment("Now acc[0] contains 8 qwords in radix-2^32:")
	f.Comment("  sum = t0 + t1*2^32 + t2*2^64 + t3*2^96 + t4*2^128 + t5*2^160 + t6*2^192 + t7*2^224")
	f.Comment("Each ti can be up to 64 bits. We need to normalize and reduce mod q.")

	f.Comment("Extract the 8 qwords from acc[0] to stack for processing")
	// Use 64 bytes of stack space for the 8 qwords
	f.VMOVDQU64(acc[0], "0(SP)")

	// Now we have t[0..7] at SP+0, SP+8, ..., SP+56
	// We need to normalize to radix-2^64 and reduce mod q

	f.Comment("Normalize radix-2^32 to radix-2^64")
	f.Comment("sum = t0 + t1*2^32 + t2*2^64 + t3*2^96 + t4*2^128 + t5*2^160 + t6*2^192 + t7*2^224")
	f.Comment("Word i = t_{2i} + (t_{2i+1} << 32), with carries propagated")

	// Result registers: T[0..4] for the 5-limb normalized value (up to 288 bits)
	T := make([]amd64.Register, 5)
	for i := 0; i < 5; i++ {
		T[i] = f.Pop(&registers)
	}
	f.LabelRegisters("T", T...)

	tmp := f.Pop(&registers)

	// Word 0: t0 + (t1_lo << 32), carry = t1_hi + CF
	f.Comment("Word 0: t0 + (t1 << 32)")
	f.MOVQ("0(SP)", T[0]) // T[0] = t0
	f.MOVQ("8(SP)", T[1]) // T[1] = t1
	f.MOVQ(T[1], tmp)     // tmp = t1
	f.SHLQ("$32", tmp)    // tmp = t1_lo << 32
	f.SHRQ("$32", T[1])   // T[1] = t1_hi
	f.ADDQ(tmp, T[0])     // T[0] = t0 + (t1_lo << 32)
	f.ADCQ("$0", T[1])    // T[1] = t1_hi + CF (carry for word 1)

	// Word 1: t2 + (t3_lo << 32) + carry
	f.Comment("Word 1: t2 + (t3 << 32) + carry")
	f.MOVQ("16(SP)", T[2]) // T[2] = t2
	f.MOVQ("24(SP)", T[3]) // T[3] = t3
	f.MOVQ(T[3], tmp)      // tmp = t3
	f.SHLQ("$32", tmp)     // tmp = t3_lo << 32
	f.SHRQ("$32", T[3])    // T[3] = t3_hi
	f.ADDQ(tmp, T[2])      // T[2] = t2 + (t3_lo << 32)
	f.ADCQ("$0", T[3])     // T[3] = t3_hi + CF
	f.ADDQ(T[1], T[2])     // T[2] += carry from word 0
	f.ADCQ("$0", T[3])     // T[3] += CF
	f.MOVQ(T[2], T[1])     // T[1] = word 1
	f.MOVQ(T[3], T[2])     // T[2] = carry for word 2

	// Word 2: t4 + (t5_lo << 32) + carry
	f.Comment("Word 2: t4 + (t5 << 32) + carry")
	f.MOVQ("32(SP)", T[3]) // T[3] = t4
	f.MOVQ("40(SP)", T[4]) // T[4] = t5
	f.MOVQ(T[4], tmp)      // tmp = t5
	f.SHLQ("$32", tmp)     // tmp = t5_lo << 32
	f.SHRQ("$32", T[4])    // T[4] = t5_hi
	f.ADDQ(tmp, T[3])      // T[3] = t4 + (t5_lo << 32)
	f.ADCQ("$0", T[4])     // T[4] = t5_hi + CF
	f.ADDQ(T[2], T[3])     // T[3] += carry from word 1
	f.ADCQ("$0", T[4])     // T[4] += CF
	f.MOVQ(T[3], T[2])     // T[2] = word 2
	f.MOVQ(T[4], T[3])     // T[3] = carry for word 3

	// Word 3: t6 + (t7_lo << 32) + carry
	f.Comment("Word 3: t6 + (t7 << 32) + carry")
	f.MOVQ("48(SP)", T[4]) // T[4] = t6
	f.MOVQ("56(SP)", tmp)  // tmp = t7
	f.MOVQ(tmp, addrA)     // addrA = t7
	f.SHLQ("$32", addrA)   // addrA = t7_lo << 32
	f.SHRQ("$32", tmp)     // tmp = t7_hi
	f.ADDQ(addrA, T[4])    // T[4] = t6 + (t7_lo << 32)
	f.ADCQ("$0", tmp)      // tmp = t7_hi + CF
	f.ADDQ(T[3], T[4])     // T[4] += carry from word 2
	f.ADCQ("$0", tmp)      // tmp += CF
	f.MOVQ(T[4], T[3])     // T[3] = word 3
	f.MOVQ(tmp, T[4])      // T[4] = word 4 (overflow)

	f.Comment("Now T[0..4] contains the normalized sum (up to 288 bits)")
	f.Comment("T[0] = word 0, T[1] = word 1, T[2] = word 2, T[3] = word 3, T[4] = overflow")

	f.Comment("Barrett reduction to get result < 2q")
	f.Comment("k = floor(T / 2^224) * mu >> 64, where mu = floor(2^288 / q)")
	f.Comment("Then result = T - k*q")

	// For Barrett reduction, we need: k ≈ (T >> 224) * mu / 2^64
	// T >> 224 = T[3] >> 32 | T[4] << 32
	// But T[4] can be up to ~33 bits, so T >> 224 is up to ~65 bits

	PL := amd64.AX
	PH := n // reuse n register
	f.MOVQ(T[3], PL)
	f.SHRQw("$32", T[4], PL) // PL = (T[4] << 32) | (T[3] >> 32) = T >> 224 (low 64 bits)
	f.MOVQ(f.mu(), amd64.DX)
	f.MULXQ(PL, PL, PH) // PH:PL = (T >> 224) * mu
	// k = PH (we want the high 64 bits of the product)
	f.MOVQ(PH, amd64.DX) // DX = k

	f.Comment("Subtract k*q from T")
	// We need to compute T - k*q
	// k*q is at most ~2^64 * 2^256 = 2^320, but since k ≈ T/q, k*q ≈ T, so the difference is small

	f.MULXQ(f.qAt(0), PL, PH)
	f.SUBQ(PL, T[0])
	f.SBBQ(PH, T[1])
	f.MULXQ(f.qAt(2), PL, PH)
	f.SBBQ(PL, T[2])
	f.SBBQ(PH, T[3])
	f.SBBQ("$0", T[4])
	f.MULXQ(f.qAt(1), PL, PH)
	f.SUBQ(PL, T[1])
	f.SBBQ(PH, T[2])
	f.MULXQ(f.qAt(3), PL, PH)
	f.SBBQ(PL, T[3])
	f.SBBQ(PH, T[4])

	f.Comment("Store result (may still be >= q, need conditional subtraction)")
	result := T[:4]
	f.Mov(result, addrRes)

	f.Comment("Conditional subtraction: if T >= q, subtract q")
	for i := 0; i < 2; i++ {
		f.SUBQ(f.qAt(0), T[0])
		f.SBBQ(f.qAt(1), T[1])
		f.SBBQ(f.qAt(2), T[2])
		f.SBBQ(f.qAt(3), T[3])
		f.SBBQ("$0", T[4])
		f.JCS(done)
		f.Mov(result, addrRes)
	}

	f.LABEL(done)
	f.RET()
}

func (f *FFAmd64) generateInnerProductW4() {
	f.Comment("innerProdVec(res, a, b *Element, n uint64) res = sum(a[0...n] * b[0...n])")
	f.Comment("")
	f.Comment("Algorithm: Accumulate 32-bit x 32-bit partial products into 16 Z registers")
	f.Comment("(8 low halves + 8 high halves), then combine into 544-bit result and reduce.")
	f.Comment("")
	f.Comment("Register allocation:")
	f.Comment("  Z0, Z1 = output result (544-bit in Z1:Z0)")
	f.Comment("  Z2 = PPL (partial product low), Z3 = PPH (partial product high)")
	f.Comment("  Z4 = Y (loaded b element), Z5 = LSW (low 32-bit mask)")
	f.Comment("  Z16-Z23 = accL[0-7], Z24-Z31 = accH[0-7]")

	const argSize = 4 * 8
	stackSize := f.StackSize(7, 2, 0)
	registers := f.FnHeader("innerProdVec", stackSize, argSize, amd64.DX, amd64.AX)
	defer f.AssertCleanStack(stackSize, 0)

	// Scalar registers
	addrA := f.Pop(&registers)
	addrB := f.Pop(&registers)
	n := f.Pop(&registers)

	done := f.NewLabel("done")
	accumulate := f.NewLabel("accumulate")

	// Explicit AVX512 register assignments to avoid conflicts
	// Z0, Z1 are reserved for output (544-bit result)
	PPL := amd64.Register("Z2")
	PPH := amd64.Register("Z3")
	Y := amd64.Register("Z4")
	LSW := amd64.Register("Z5")

	// 16 accumulators: use Z16-Z23 for low halves, Z24-Z31 for high halves
	// This keeps them separate from the output and temp registers
	ACC := amd64.Register("Z16") // ACC aliases accL[0]
	accL := []amd64.Register{"Z16", "Z17", "Z18", "Z19", "Z20", "Z21", "Z22", "Z23"}
	accH := []amd64.Register{"Z24", "Z25", "Z26", "Z27", "Z28", "Z29", "Z30", "Z31"}

	f.Comment("Load arguments")
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("b+16(FP)", addrB)
	f.MOVQ("n+24(FP)", n)

	f.Comment("Create mask for low dword in each qword: 0x00000000FFFFFFFF")
	f.VPCMPEQB("Y0", "Y0", "Y0")
	f.VPMOVZXDQ("Y0", LSW)

	f.Comment("Initialize all 16 accumulators to zero")
	f.VPXORQ(accL[0], accL[0], accL[0])
	for i := 1; i < 8; i++ {
		f.VMOVDQA64(accL[0], accL[i])
	}
	for i := 0; i < 8; i++ {
		f.VMOVDQA64(accL[0], accH[i])
	}

	f.Comment("Main loop: multiply and accumulate partial products")
	f.Comment("For each element pair (a[i], b[i]):")
	f.Comment("  - Load b[i] as 8 dwords zero-extended to qwords")
	f.Comment("  - Multiply each dword of a[i] by all dwords of b[i]")
	f.Comment("  - Split results into low/high 32 bits and accumulate")

	// Define MAC (multiply-accumulate) macro
	mac := f.Define("MAC", 3, func(inputs ...any) {
		opLeft := inputs[0] // memory operand for dword of a
		lo := inputs[1]     // accumulator for low 32 bits
		hi := inputs[2]     // accumulator for high 32 bits

		f.VPMULUDQ_BCST(opLeft, Y, PPL) // PPL = a_dword * b_dwords (64-bit results)
		f.VPSRLQ("$32", PPL, PPH)       // PPH = high 32 bits
		f.VPANDQ(LSW, PPL, PPL)         // PPL = low 32 bits
		f.VPADDQ(PPL, lo, lo)           // accumulate low
		f.VPADDQ(PPH, hi, hi)           // accumulate high
	}, true)

	f.Loop(n, func() {
		f.Comment("Load b[i]: 8 dwords -> 8 qwords via zero-extension")
		f.VPMOVZXDQ("0("+addrB+")", Y)

		f.Comment("Multiply each dword of a[i] (8 dwords) by b[i] and accumulate")
		for i := 0; i < 8; i++ {
			mac(fmt.Sprintf("%d(%s)", i*4, addrA), accL[i], accH[i])
		}

		f.ADDQ("$32", addrA)
		f.ADDQ("$32", addrB)
	})

	f.LABEL(accumulate)
	f.Comment("Combine 16 partial product accumulators into 544-bit result in Z1:Z0")
	f.Comment("Each accumulator has 8 qwords; we reduce across qwords using VALIGND")

	// Setup masks for VALIGND operations
	f.MOVQ(uint64(0x1555), amd64.AX)
	f.KMOVD(amd64.AX, "K1")
	f.MOVQ(uint64(1), amd64.AX)
	f.KMOVD(amd64.AX, "K2")

	f.Comment("Word 0: lowest 32 bits from accL[0]")
	f.VALIGND_Z("$16", ACC, ACC, "K2", "Z0")
	f.KSHIFTLW("$1", "K2", "K2")

	f.Comment("Propagate carries and add contributions for word 1")
	f.VPSRLQ("$32", ACC, PPL)
	f.VALIGND_Z("$2", ACC, ACC, "K1", ACC)
	f.VPADDQ(PPL, ACC, ACC)
	f.VPANDQ(LSW, accH[0], PPL)
	f.VPADDQ(PPL, ACC, ACC)
	f.VPANDQ(LSW, accL[1], PPL)
	f.VPADDQ(PPL, ACC, ACC)
	f.VALIGNDk("$15", ACC, ACC, "K2", "Z0")
	f.KSHIFTLW("$1", "K2", "K2")

	f.Comment("Words 2-7: combine partial products with carry propagation")
	addPP := f.Define("ADDPP", 5, func(inputs ...any) {
		axH := inputs[0] // accH[x]
		ayL := inputs[1] // accL[y]
		ayH := inputs[2] // accH[y]
		azL := inputs[3] // accL[z]
		idx := inputs[4] // word index

		f.VPSRLQ("$32", ACC, PPL)
		f.VALIGND_Z("$2", ACC, ACC, "K1", ACC)
		f.VPADDQ(PPL, ACC, ACC)
		f.VPSRLQ("$32", axH, axH)
		f.VPADDQ(axH, ACC, ACC)
		f.VPSRLQ("$32", ayL, ayL)
		f.VPADDQ(ayL, ACC, ACC)
		f.VPANDQ(LSW, ayH, PPL)
		f.VPADDQ(PPL, ACC, ACC)
		f.VPANDQ(LSW, azL, PPL)
		f.VPADDQ(PPL, ACC, ACC)
		f.VALIGNDk("$16-"+idx.(amd64.Register), ACC, ACC, "K2", "Z0")
		f.KADDW("K2", "K2", "K2")
	}, true)

	addPP(accH[0], accL[1], accH[1], accL[2], "2")
	addPP(accH[1], accL[2], accH[2], accL[3], "3")
	addPP(accH[2], accL[3], accH[3], accL[4], "4")
	addPP(accH[3], accL[4], accH[4], accL[5], "5")
	addPP(accH[4], accL[5], accH[5], accL[6], "6")
	addPP(accH[5], accL[6], accH[6], accL[7], "7")

	f.Comment("Word 8: final contributions from accH[6], accL[7], accH[7]")
	f.VPSRLQ("$32", ACC, PPL)
	f.VALIGND_Z("$2", ACC, ACC, "K1", ACC)
	f.VPADDQ(PPL, ACC, ACC)
	f.VPSRLQ("$32", accH[6], accH[6])
	f.VPADDQ(accH[6], ACC, ACC)
	f.VPSRLQ("$32", accL[7], accL[7])
	f.VPADDQ(accL[7], ACC, ACC)
	f.VPANDQ(LSW, accH[7], PPL)
	f.VPADDQ(PPL, ACC, ACC)
	f.VALIGNDk("$16-8", ACC, ACC, "K2", "Z0")
	f.KSHIFTLW("$1", "K2", "K2")

	f.Comment("Word 9: remaining from accH[7]")
	f.VPSRLQ("$32", ACC, PPL)
	f.VALIGND_Z("$2", ACC, ACC, "K1", ACC)
	f.VPADDQ(PPL, ACC, ACC)
	f.VPSRLQ("$32", accH[7], accH[7])
	f.VPADDQ(accH[7], ACC, ACC)
	f.VALIGNDk("$16-9", ACC, ACC, "K2", "Z0")
	f.KSHIFTLW("$1", "K2", "K2")

	f.Comment("Words 10-15: propagate remaining carries")
	addPP2 := f.Define("ADDPP2", 1, func(args ...any) {
		idx := args[0]
		f.VPSRLQ("$32", ACC, PPL)
		f.VALIGND_Z("$2", ACC, ACC, "K1", ACC)
		f.VPADDQ(PPL, ACC, ACC)
		f.VALIGNDk("$16-"+idx.(amd64.Register), ACC, ACC, "K2", "Z0")
		f.KSHIFTLW("$1", "K2", "K2")
	}, true)

	for i := 10; i <= 15; i++ {
		addPP2(fmt.Sprintf("%d", i))
	}

	f.Comment("Final carry into Z1")
	f.VPSRLQ("$32", ACC, PPL)
	f.VALIGND_Z("$2", ACC, ACC, "K1", ACC)
	f.VPADDQ(PPL, ACC, ACC)
	f.VMOVDQA64_Z(ACC, "K1", "Z1")

	f.Comment("Montgomery reduction of the 544-bit result")
	f.Comment("Extract low 4 qwords from Z0 into scalar registers")
	T := make([]amd64.Register, 5)
	for i := 0; i < 5; i++ {
		T[i] = f.Pop(&registers)
	}
	f.LabelRegisters("T", T...)

	f.VMOVQ("X0", T[1])
	f.VALIGNQ("$1", "Z0", "Z1", "Z0")
	f.VMOVQ("X0", T[2])
	f.VALIGNQ("$1", "Z0", "Z0", "Z0")
	f.VMOVQ("X0", T[3])
	f.VALIGNQ("$1", "Z0", "Z0", "Z0")
	f.VMOVQ("X0", T[4])
	f.VALIGNQ("$1", "Z0", "Z0", "Z0")
	f.XORQ(T[0], T[0])

	f.Comment("4 rounds of Montgomery reduction")
	PH := f.Pop(&registers)
	PL := amd64.AX

	for round := 0; round < 4; round++ {
		f.Comment(fmt.Sprintf("Montgomery reduction round %d", round+1))
		src := T[1+round%4]
		f.MOVQ(f.qInv0(), amd64.DX)
		f.MULXQ(src, amd64.DX, PH)

		// Add q * m to T (4-word addition with carry)
		f.MULXQ(f.qAt(0), PL, PH)
		f.ADDQ(PL, T[(1+round)%5])
		f.ADCQ(PH, T[(2+round)%5])
		f.MULXQ(f.qAt(2), PL, PH)
		f.ADCQ(PL, T[(3+round)%5])
		f.ADCQ(PH, T[(4+round)%5])
		f.ADCQ("$0", T[(0+round)%5])
		f.MULXQ(f.qAt(1), PL, PH)
		f.ADDQ(PL, T[(2+round)%5])
		f.ADCQ(PH, T[(3+round)%5])
		f.MULXQ(f.qAt(3), PL, PH)
		f.ADCQ(PL, T[(4+round)%5])
		f.ADCQ(PH, T[(0+round)%5])
		f.ADCQ("$0", T[(1+round)%5])
	}

	f.Comment("Add remaining 5 qwords from Z0")
	for i := 0; i < 5; i++ {
		f.VMOVQ("X0", PL)
		if i == 0 {
			f.ADDQ(PL, T[0])
		} else {
			f.ADCQ(PL, T[i])
		}
		if i < 4 {
			f.VALIGNQ("$1", "Z0", "Z0", "Z0")
		}
	}

	f.Comment("Barrett reduction; see Handbook of Applied Cryptography, Algorithm 14.42.")
	f.MOVQ(T[3], amd64.AX)
	f.SHRQw("$32", T[4], amd64.AX)
	f.MOVQ(f.mu(), amd64.DX)
	f.MULQ(amd64.DX)

	f.Comment("Subtract k*q from T")
	f.MULXQ(f.qAt(0), PL, PH)
	f.SUBQ(PL, T[0])
	f.SBBQ(PH, T[1])
	f.MULXQ(f.qAt(2), PL, PH)
	f.SBBQ(PL, T[2])
	f.SBBQ(PH, T[3])
	f.SBBQ("$0", T[4])
	f.MULXQ(f.qAt(1), PL, PH)
	f.SUBQ(PL, T[1])
	f.SBBQ(PH, T[2])
	f.MULXQ(f.qAt(3), PL, PH)
	f.SBBQ(PL, T[3])
	f.SBBQ(PH, T[4])

	f.Comment("Conditional subtraction: up to 2 subtractions needed to get result < q")
	addrRes := f.Pop(&registers)
	f.MOVQ("res+0(FP)", addrRes)
	result := T[:4]
	f.Mov(result, addrRes)

	for i := 0; i < 2; i++ {
		f.SUBQ(f.qAt(0), T[0])
		f.SBBQ(f.qAt(1), T[1])
		f.SBBQ(f.qAt(2), T[2])
		f.SBBQ(f.qAt(3), T[3])
		f.SBBQ("$0", T[4])
		f.JCS(done)
		f.Mov(result, addrRes)
	}

	f.LABEL(done)
	f.RET()
}
