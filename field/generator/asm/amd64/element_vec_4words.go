// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package amd64

import (
	"fmt"
	"strconv"

	"github.com/consensys/bavard/amd64"
)

// addVec res = a + b
// func addVec(res, a, b *{{.ElementName}}, n uint64)
func (_f *FFAmd64) generateAddVecW4() {
	_f.Comment("addVec(res, a, b *Element, n uint64) res[0...n] = a[0...n] + b[0...n]")

	const argSize = 4 * 8
	stackSize := _f.StackSize(_f.NbWords*2+4, 0, 0)
	registers := _f.FnHeader("addVec", stackSize, argSize)
	defer _f.AssertCleanStack(stackSize, 0)

	// registers & labels we need
	addrA := _f.Pop(&registers)
	addrB := _f.Pop(&registers)
	addrRes := _f.Pop(&registers)
	len := _f.Pop(&registers)

	a := _f.PopN(&registers)
	t := _f.PopN(&registers)

	loop := _f.NewLabel("loop")
	done := _f.NewLabel("done")

	// load arguments
	_f.MOVQ("res+0(FP)", addrRes)
	_f.MOVQ("a+8(FP)", addrA)
	_f.MOVQ("b+16(FP)", addrB)
	_f.MOVQ("n+24(FP)", len)

	_f.LABEL(loop)

	_f.TESTQ(len, len)
	_f.JEQ(done, "n == 0, we are done")

	// a = a + b
	_f.LabelRegisters("a", a...)
	_f.Mov(addrA, a)
	_f.Add(addrB, a)
	_f.WriteLn(fmt.Sprintf("PREFETCHT0 2048(%[1]s)", addrA))
	_f.WriteLn(fmt.Sprintf("PREFETCHT0 2048(%[1]s)", addrB))

	// reduce a
	_f.ReduceElement(a, t, false)

	// save a into res
	_f.Mov(a, addrRes)
	_f.Comment("increment pointers to visit next element")
	_f.ADDQ("$32", addrA)
	_f.ADDQ("$32", addrB)
	_f.ADDQ("$32", addrRes)
	_f.DECQ(len, "decrement n")
	_f.JMP(loop)

	_f.LABEL(done)

	_f.RET()

	_f.Push(&registers, a...)
	_f.Push(&registers, t...)
	_f.Push(&registers, addrA, addrB, addrRes, len)

}

// subVec res = a - b
// func subVec(res, a, b *{{.ElementName}}, n uint64)
func (_f *FFAmd64) generateSubVecW4() {
	_f.Comment("subVec(res, a, b *Element, n uint64) res[0...n] = a[0...n] - b[0...n]")

	const argSize = 4 * 8
	stackSize := _f.StackSize(_f.NbWords*2+5, 0, 0)
	registers := _f.FnHeader("subVec", stackSize, argSize)
	defer _f.AssertCleanStack(stackSize, 0)

	// registers
	addrA := _f.Pop(&registers)
	addrB := _f.Pop(&registers)
	addrRes := _f.Pop(&registers)
	len := _f.Pop(&registers)
	zero := _f.Pop(&registers)

	a := _f.PopN(&registers)
	q := _f.PopN(&registers)

	loop := _f.NewLabel("loop")
	done := _f.NewLabel("done")

	// load arguments
	_f.MOVQ("res+0(FP)", addrRes)
	_f.MOVQ("a+8(FP)", addrA)
	_f.MOVQ("b+16(FP)", addrB)
	_f.MOVQ("n+24(FP)", len)

	_f.XORQ(zero, zero)

	_f.LABEL(loop)

	_f.TESTQ(len, len)
	_f.JEQ(done, "n == 0, we are done")

	// a = a - b
	_f.LabelRegisters("a", a...)
	_f.Mov(addrA, a)
	_f.Sub(addrB, a)
	_f.WriteLn(fmt.Sprintf("PREFETCHT0 2048(%[1]s)", addrA))
	_f.WriteLn(fmt.Sprintf("PREFETCHT0 2048(%[1]s)", addrB))

	// reduce a
	_f.Comment("reduce (a-b) mod q")
	_f.LabelRegisters("q", q...)
	for i := 0; i < _f.NbWords; i++ {
		_f.MOVQ(fmt.Sprintf("$const_q%d", i), q[i])
	}
	for i := 0; i < _f.NbWords; i++ {
		_f.CMOVQCC(zero, q[i])
	}
	// add registers (q or 0) to a, and set to result
	_f.Comment("add registers (q or 0) to a, and set to result")
	_f.Add(q, a)

	// save a into res
	_f.Mov(a, addrRes)

	_f.Comment("increment pointers to visit next element")
	_f.ADDQ("$32", addrA)
	_f.ADDQ("$32", addrB)
	_f.ADDQ("$32", addrRes)
	_f.DECQ(len, "decrement n")
	_f.JMP(loop)

	_f.LABEL(done)

	_f.RET()

	_f.Push(&registers, a...)
	_f.Push(&registers, q...)
	_f.Push(&registers, addrA, addrB, addrRes, len, zero)

}

// sumVec res = sum(a[0...n])
func (_f *FFAmd64) generateSumVecW4() {
	_f.Comment("sumVec(res, a *Element, n uint64) res = sum(a[0...n])")

	const argSize = 3 * 8
	stackSize := _f.StackSize(12, 2, 0)
	registers := _f.FnHeader("sumVec", stackSize, argSize, amd64.DX, amd64.AX)
	defer _f.AssertCleanStack(stackSize, 0)

	_f.WriteLn(`
	// Derived from https://github.com/a16z/vectorized-fields
	// The idea is to use Z registers to accumulate the sum of elements, 8 by 8
	// first, we handle the case where n % 8 != 0
	// then, we loop over the elements 8 by 8 and accumulate the sum in the Z registers
	// finally, we reduce the sum and store it in res
	// 
	// when we move an element of a into a Z register, we use VPMOVZXDQ
	// let's note w0...w3 the 4 64bits words of ai: w0 = ai[0], w1 = ai[1], w2 = ai[2], w3 = ai[3]
	// VPMOVZXDQ(ai, Z0) will result in 
	// Z0= [hi(w3), lo(w3), hi(w2), lo(w2), hi(w1), lo(w1), hi(w0), lo(w0)]
	// with hi(wi) the high 32 bits of wi and lo(wi) the low 32 bits of wi
	// we can safely add 2^32+1 times Z registers constructed this way without overflow
	// since each of this lo/hi bits are moved into a "64bits" slot
	// N = 2^64-1 / 2^32-1 = 2^32+1
	//
	// we then propagate the carry using ADOXQ and ADCXQ
	// r0 = w0l + lo(woh)
	// r1 = carry + hi(woh) + w1l + lo(w1h)
	// r2 = carry + hi(w1h) + w2l + lo(w2h)
	// r3 = carry + hi(w2h) + w3l + lo(w3h)
	// r4 = carry + hi(w3h)
	// we then reduce the sum using a single-word Barrett reduction
	// we pick mu = 2^288 / q; which correspond to 4.5 words max.
	// meaning we must guarantee that r4 fits in 32bits.
	// To do so, we reduce N to 2^32-1 (since r4 receives 2 carries max)
	`)

	// registers & labels we need
	addrA := _f.Pop(&registers)
	n := _f.Pop(&registers)
	nMod8 := _f.Pop(&registers)

	loop := _f.NewLabel("loop8by8")
	done := _f.NewLabel("done")
	loopSingle := _f.NewLabel("loop_single")
	accumulate := _f.NewLabel("accumulate")

	// AVX512 registers
	Z0 := amd64.Register("Z0")
	Z1 := amd64.Register("Z1")
	Z2 := amd64.Register("Z2")
	Z3 := amd64.Register("Z3")
	Z4 := amd64.Register("Z4")
	Z5 := amd64.Register("Z5")
	Z6 := amd64.Register("Z6")
	Z7 := amd64.Register("Z7")
	Z8 := amd64.Register("Z8")

	X0 := amd64.Register("X0")

	// load arguments
	_f.MOVQ("a+8(FP)", addrA)
	_f.MOVQ("n+16(FP)", n)

	_f.Comment("initialize accumulators Z0, Z1, Z2, Z3, Z4, Z5, Z6, Z7")
	_f.VXORPS(Z0, Z0, Z0)
	_f.VMOVDQA64(Z0, Z1)
	_f.VMOVDQA64(Z0, Z2)
	_f.VMOVDQA64(Z0, Z3)
	_f.VMOVDQA64(Z0, Z4)
	_f.VMOVDQA64(Z0, Z5)
	_f.VMOVDQA64(Z0, Z6)
	_f.VMOVDQA64(Z0, Z7)

	// note: we don't need to handle the case n==0; handled by caller already.
	// f.TESTQ(n, n)
	// f.JEQ(done, "n == 0, we are done")

	_f.LabelRegisters("n % 8", nMod8)
	_f.LabelRegisters("n / 8", n)
	_f.MOVQ(n, nMod8)
	_f.ANDQ("$7", nMod8) // nMod8 = n % 8
	_f.SHRQ("$3", n)     // len = n / 8

	_f.LABEL(loopSingle)
	_f.TESTQ(nMod8, nMod8)
	_f.JEQ(loop, "n % 8 == 0, we are going to loop over 8 by 8")

	_f.VPMOVZXDQ("0("+addrA+")", Z8)
	_f.VPADDQ(Z8, Z0, Z0)
	_f.ADDQ("$32", addrA)

	_f.DECQ(nMod8, "decrement nMod8")
	_f.JMP(loopSingle)

	_f.Push(&registers, nMod8) // we don't need tmp0

	_f.LABEL(loop)
	_f.TESTQ(n, n)
	_f.JEQ(accumulate, "n == 0, we are going to accumulate")

	for i := 0; i < 8; i++ {
		r := fmt.Sprintf("Z%d", i+8)
		_f.VPMOVZXDQ(fmt.Sprintf("%d*32("+string(addrA)+")", i), r)
	}

	_f.WriteLn(fmt.Sprintf("PREFETCHT0 4096(%[1]s)", addrA))
	for i := 0; i < 8; i++ {
		r := fmt.Sprintf("Z%d", i)
		_f.VPADDQ(fmt.Sprintf("Z%d", i+8), r, r)
	}

	_f.Comment("increment pointers to visit next 8 elements")
	_f.ADDQ("$256", addrA)
	_f.DECQ(n, "decrement n")
	_f.JMP(loop)

	_f.Push(&registers, n, addrA)

	_f.LABEL(accumulate)

	_f.Comment("accumulate the 8 Z registers into Z0")
	_f.VPADDQ(Z7, Z6, Z6)
	_f.VPADDQ(Z6, Z5, Z5)
	_f.VPADDQ(Z5, Z4, Z4)
	_f.VPADDQ(Z4, Z3, Z3)
	_f.VPADDQ(Z3, Z2, Z2)
	_f.VPADDQ(Z2, Z1, Z1)
	_f.VPADDQ(Z1, Z0, Z0)

	w0l := _f.Pop(&registers)
	w0h := _f.Pop(&registers)
	w1l := _f.Pop(&registers)
	w1h := _f.Pop(&registers)
	w2l := _f.Pop(&registers)
	w2h := _f.Pop(&registers)
	w3l := _f.Pop(&registers)
	w3h := _f.Pop(&registers)
	low0h := _f.Pop(&registers)
	low1h := _f.Pop(&registers)
	low2h := _f.Pop(&registers)
	low3h := _f.Pop(&registers)

	// Propagate carries
	_f.Comment("carry propagation")

	_f.LabelRegisters("lo(w0)", w0l)
	_f.LabelRegisters("hi(w0)", w0h)
	_f.LabelRegisters("lo(w1)", w1l)
	_f.LabelRegisters("hi(w1)", w1h)
	_f.LabelRegisters("lo(w2)", w2l)
	_f.LabelRegisters("hi(w2)", w2h)
	_f.LabelRegisters("lo(w3)", w3l)
	_f.LabelRegisters("hi(w3)", w3h)

	_f.VMOVQ(X0, w0l)
	_f.VALIGNQ("$1", Z0, Z0, Z0)
	_f.VMOVQ(X0, w0h)
	_f.VALIGNQ("$1", Z0, Z0, Z0)
	_f.VMOVQ(X0, w1l)
	_f.VALIGNQ("$1", Z0, Z0, Z0)
	_f.VMOVQ(X0, w1h)
	_f.VALIGNQ("$1", Z0, Z0, Z0)
	_f.VMOVQ(X0, w2l)
	_f.VALIGNQ("$1", Z0, Z0, Z0)
	_f.VMOVQ(X0, w2h)
	_f.VALIGNQ("$1", Z0, Z0, Z0)
	_f.VMOVQ(X0, w3l)
	_f.VALIGNQ("$1", Z0, Z0, Z0)
	_f.VMOVQ(X0, w3h)

	_f.LabelRegisters("lo(hi(wo))", low0h)
	_f.LabelRegisters("lo(hi(w1))", low1h)
	_f.LabelRegisters("lo(hi(w2))", low2h)
	_f.LabelRegisters("lo(hi(w3))", low3h)

	type hilo struct {
		hi, lo amd64.Register
	}

	splitLoHi := _f.Define("SPLIT_LO_HI", 2, func(args ...any) {
		lo := args[0]
		hi := args[1]
		_f.MOVQ(hi, lo)
		_f.ANDQ("$0xffffffff", lo)
		_f.SHLQ("$32", lo)
		_f.SHRQ("$32", hi)
	}, true)

	for _, v := range []hilo{{w0h, low0h}, {w1h, low1h}, {w2h, low2h}, {w3h, low3h}} {
		splitLoHi(v.lo, v.hi)
	}

	_f.WriteLn(`
	// r0 = w0l + lo(woh)
	// r1 = carry + hi(woh) + w1l + lo(w1h)
	// r2 = carry + hi(w1h) + w2l + lo(w2h)
	// r3 = carry + hi(w2h) + w3l + lo(w3h)
	// r4 = carry + hi(w3h)
	`)
	_f.XORQ(amd64.AX, amd64.AX, "clear the flags")
	_f.ADOXQ(low0h, w0l)

	_f.ADOXQ(low1h, w1l)
	_f.ADCXQ(w0h, w1l)

	_f.ADOXQ(low2h, w2l)
	_f.ADCXQ(w1h, w2l)

	_f.ADOXQ(low3h, w3l)
	_f.ADCXQ(w2h, w3l)

	_f.ADOXQ(amd64.AX, w3h)
	_f.ADCXQ(amd64.AX, w3h)

	r0 := w0l
	r1 := w1l
	r2 := w2l
	r3 := w3l
	r4 := w3h

	r := []amd64.Register{r0, r1, r2, r3, r4}
	_f.LabelRegisters("r", r...)
	// we don't need w0h, w1h, w2h anymore
	_f.Push(&registers, w0h, w1h, w2h)
	// we don't need the low bits anymore
	_f.Push(&registers, low0h, low1h, low2h, low3h)

	// Reduce using single-word Barrett
	mu := _f.Pop(&registers)

	_f.Comment("reduce using single-word Barrett")
	_f.Comment("see see Handbook of Applied Cryptography, Algorithm 14.42.")
	_f.LabelRegisters("mu=2^288 / q", mu)
	_f.MOVQ(_f.mu(), mu)
	_f.MOVQ(r3, amd64.AX)
	_f.SHRQw("$32", r4, amd64.AX)
	_f.MULQ(mu, "high bits of res stored in DX")

	_f.MULXQ(_f.qAt(0), amd64.AX, mu)
	_f.SUBQ(amd64.AX, r0)
	_f.SBBQ(mu, r1)

	_f.MULXQ(_f.qAt(2), amd64.AX, mu)
	_f.SBBQ(amd64.AX, r2)
	_f.SBBQ(mu, r3)
	_f.SBBQ("$0", r4)

	_f.MULXQ(_f.qAt(1), amd64.AX, mu)
	_f.SUBQ(amd64.AX, r1)
	_f.SBBQ(mu, r2)

	_f.MULXQ(_f.qAt(3), amd64.AX, mu)
	_f.SBBQ(amd64.AX, r3)
	_f.SBBQ(mu, r4)

	// we need up to 2 conditional substractions to be < q
	modReduced := _f.NewLabel("modReduced")
	t := _f.PopN(&registers)
	_f.Mov(r[:4], t) // backup r0 to r3 (our result)

	// sub modulus
	_f.SUBQ(_f.qAt(0), r0)
	_f.SBBQ(_f.qAt(1), r1)
	_f.SBBQ(_f.qAt(2), r2)
	_f.SBBQ(_f.qAt(3), r3)
	_f.SBBQ("$0", r4)

	// if borrow, we go to mod reduced
	_f.JCS(modReduced)
	_f.Mov(r, t)
	_f.SUBQ(_f.qAt(0), r0)
	_f.SBBQ(_f.qAt(1), r1)
	_f.SBBQ(_f.qAt(2), r2)
	_f.SBBQ(_f.qAt(3), r3)
	_f.SBBQ("$0", r4)

	// if borrow, we skip to the end
	_f.JCS(modReduced)
	_f.Mov(r, t)

	_f.LABEL(modReduced)
	addrRes := mu
	_f.MOVQ("res+0(FP)", addrRes)
	_f.Mov(t, addrRes)

	_f.LABEL(done)

	_f.RET()
	_f.Push(&registers, mu)
	_f.Push(&registers, w0l, w1l, w2l, w3l, w3h)
}

func (_f *FFAmd64) generateInnerProductW4() {
	_f.Comment("innerProdVec(res, a,b *Element, n uint64) res = sum(a[0...n] * b[0...n])")

	const argSize = 4 * 8
	stackSize := _f.StackSize(7, 2, 0)
	registers := _f.FnHeader("innerProdVec", stackSize, argSize, amd64.DX, amd64.AX)
	defer _f.AssertCleanStack(stackSize, 0)

	// registers & labels we need
	PX := _f.Pop(&registers)
	PY := _f.Pop(&registers)
	LEN := _f.Pop(&registers)

	loop := _f.NewLabel("loop")
	done := _f.NewLabel("done")
	AddPP := _f.NewLabel("accumulate")

	// AVX512 registers
	PPL := amd64.Register("Z2")
	PPH := amd64.Register("Z3")
	Y := amd64.Register("Z4")
	LSW := amd64.Register("Z5")

	ACC := amd64.Register("Z16")
	A0L := amd64.Register("Z16")
	A1L := amd64.Register("Z17")
	A2L := amd64.Register("Z18")
	A3L := amd64.Register("Z19")
	A4L := amd64.Register("Z20")
	A5L := amd64.Register("Z21")
	A6L := amd64.Register("Z22")
	A7L := amd64.Register("Z23")
	A0H := amd64.Register("Z24")
	A1H := amd64.Register("Z25")
	A2H := amd64.Register("Z26")
	A3H := amd64.Register("Z27")
	A4H := amd64.Register("Z28")
	A5H := amd64.Register("Z29")
	A6H := amd64.Register("Z30")
	A7H := amd64.Register("Z31")

	// load arguments
	_f.MOVQ("a+8(FP)", PX)
	_f.MOVQ("b+16(FP)", PY)
	_f.MOVQ("n+24(FP)", LEN)

	_f.Comment("Create mask for low dword in each qword")
	_f.VPCMPEQB("Y0", "Y0", "Y0")
	_f.VPMOVZXDQ("Y0", LSW)

	// Clear accumulator registers
	_f.VPXORQ(A0L, A0L, A0L)
	_f.VMOVDQA64(A0L, A1L)
	_f.VMOVDQA64(A0L, A2L)
	_f.VMOVDQA64(A0L, A3L)
	_f.VMOVDQA64(A0L, A4L)
	_f.VMOVDQA64(A0L, A5L)
	_f.VMOVDQA64(A0L, A6L)
	_f.VMOVDQA64(A0L, A7L)
	_f.VMOVDQA64(A0L, A0H)
	_f.VMOVDQA64(A0L, A1H)
	_f.VMOVDQA64(A0L, A2H)
	_f.VMOVDQA64(A0L, A3H)
	_f.VMOVDQA64(A0L, A4H)
	_f.VMOVDQA64(A0L, A5H)
	_f.VMOVDQA64(A0L, A6H)
	_f.VMOVDQA64(A0L, A7H)

	// note: we don't need to handle the case n==0; handled by caller already.
	_f.TESTQ(LEN, LEN)
	_f.JEQ(done, "n == 0, we are done")

	_f.LABEL(loop)
	_f.TESTQ(LEN, LEN)
	_f.JEQ(AddPP, "n == 0 we can accumulate")

	_f.VPMOVZXDQ("("+PY+")", Y)

	_f.ADDQ("$32", PY)

	_f.Comment("we multiply and accumulate partial products of 4 bytes * 32 bytes")

	mac := _f.Define("MAC", 3, func(inputs ...any) {
		opLeft := inputs[0]
		lo := inputs[1]
		hi := inputs[2]

		_f.VPMULUDQ_BCST(opLeft, Y, PPL)
		_f.VPSRLQ("$32", PPL, PPH)
		_f.VPANDQ(LSW, PPL, PPL)
		_f.VPADDQ(PPL, lo, lo)
		_f.VPADDQ(PPH, hi, hi)
	}, true)

	mac("0*4("+PX+")", A0L, A0H)
	mac("1*4("+PX+")", A1L, A1H)
	mac("2*4("+PX+")", A2L, A2H)
	mac("3*4("+PX+")", A3L, A3H)
	mac("4*4("+PX+")", A4L, A4H)
	mac("5*4("+PX+")", A5L, A5H)
	mac("6*4("+PX+")", A6L, A6H)
	mac("7*4("+PX+")", A7L, A7H)

	_f.ADDQ("$32", PX)

	_f.DECQ(LEN, "decrement n")
	_f.JMP(loop)

	_f.Push(&registers, LEN, PX, PY)

	_f.LABEL(AddPP)
	_f.Comment("we accumulate the partial products into 544bits in Z1:Z0")

	_f.MOVQ(uint64(0x1555), amd64.AX)
	_f.KMOVD(amd64.AX, "K1")

	_f.MOVQ(uint64(1), amd64.AX)
	_f.KMOVD(amd64.AX, "K2")

	// ACC starts with the value of A0L

	_f.Comment("store the least significant 32 bits of ACC (starts with A0L) in Z0")
	_f.VALIGND_Z("$16", ACC, ACC, "K2", "Z0")
	_f.KSHIFTLW("$1", "K2", "K2")

	_f.VPSRLQ("$32", ACC, PPL)
	_f.VALIGND_Z("$2", ACC, ACC, "K1", ACC)
	_f.VPADDQ(PPL, ACC, ACC)

	_f.VPANDQ(LSW, A0H, PPL)
	_f.VPADDQ(PPL, ACC, ACC)

	_f.VPANDQ(LSW, A1L, PPL)
	_f.VPADDQ(PPL, ACC, ACC)

	// Word 1 of z is ready
	_f.VALIGND("$15", ACC, ACC, "K2", "Z0")
	_f.KSHIFTLW("$1", "K2", "K2")

	_f.Comment("macro to add partial products and store the result in Z0")
	addPP := _f.Define("ADDPP", 5, func(inputs ...any) {
		AxH := inputs[0]
		AyL := inputs[1]
		AyH := inputs[2]
		AzL := inputs[3]
		I := inputs[4]
		_f.VPSRLQ("$32", ACC, PPL)
		_f.VALIGND_Z("$2", ACC, ACC, "K1", ACC)
		_f.VPADDQ(PPL, ACC, ACC)
		_f.VPSRLQ("$32", AxH, AxH)
		_f.VPADDQ(AxH, ACC, ACC)
		_f.VPSRLQ("$32", AyL, AyL)
		_f.VPADDQ(AyL, ACC, ACC)
		_f.VPANDQ(LSW, AyH, PPL)
		_f.VPADDQ(PPL, ACC, ACC)
		_f.VPANDQ(LSW, AzL, PPL)
		_f.VPADDQ(PPL, ACC, ACC)
		_f.VALIGND("$16-"+I.(amd64.Register), ACC, ACC, "K2", "Z0")
		_f.KADDW("K2", "K2", "K2")
	}, true)

	addPP(A0H, A1L, A1H, A2L, "2")
	addPP(A1H, A2L, A2H, A3L, "3")
	addPP(A2H, A3L, A3H, A4L, "4")
	addPP(A3H, A4L, A4H, A5L, "5")
	addPP(A4H, A5L, A5H, A6L, "6")
	addPP(A5H, A6L, A6H, A7L, "7")
	_f.VPSRLQ("$32", ACC, PPL)
	_f.VALIGND_Z("$2", ACC, ACC, "K1", ACC)
	_f.VPADDQ(PPL, ACC, ACC)
	_f.VPSRLQ("$32", A6H, A6H)
	_f.VPADDQ(A6H, ACC, ACC)
	_f.VPSRLQ("$32", A7L, A7L)
	_f.VPADDQ(A7L, ACC, ACC)
	_f.VPANDQ(LSW, A7H, PPL)
	_f.VPADDQ(PPL, ACC, ACC)
	_f.VALIGND("$16-8", ACC, ACC, "K2", "Z0")
	_f.KSHIFTLW("$1", "K2", "K2")

	_f.VPSRLQ("$32", ACC, PPL)
	_f.VALIGND_Z("$2", ACC, ACC, "K1", ACC)
	_f.VPADDQ(PPL, ACC, ACC)
	_f.VPSRLQ("$32", A7H, A7H)
	_f.VPADDQ(A7H, ACC, ACC)
	_f.VALIGND("$16-9", ACC, ACC, "K2", "Z0")
	_f.KSHIFTLW("$1", "K2", "K2")

	addPP2 := _f.Define("ADDPP2", 1, func(args ...any) {
		_f.VPSRLQ("$32", ACC, PPL)
		_f.VALIGND_Z("$2", ACC, ACC, "K1", ACC)
		_f.VPADDQ(PPL, ACC, ACC)
		_f.VALIGND("$16-"+args[0].(amd64.Register), ACC, ACC, "K2", "Z0")
		_f.KSHIFTLW("$1", "K2", "K2")
	}, true)

	addPP2("10")
	addPP2("11")
	addPP2("12")
	addPP2("13")
	addPP2("14")
	addPP2("15")

	_f.VPSRLQ("$32", ACC, PPL)
	_f.VALIGND_Z("$2", ACC, ACC, "K1", ACC)
	_f.VPADDQ(PPL, ACC, ACC)
	_f.VMOVDQA64_Z(ACC, "K1", "Z1")

	T0 := _f.Pop(&registers)
	T1 := _f.Pop(&registers)
	T2 := _f.Pop(&registers)
	T3 := _f.Pop(&registers)
	T4 := _f.Pop(&registers)

	_f.Comment("Extract the 4 least significant qwords of Z0")
	_f.VMOVQ("X0", T1)
	_f.VALIGNQ("$1", "Z0", "Z1", "Z0")
	_f.VMOVQ("X0", T2)
	_f.VALIGNQ("$1", "Z0", "Z0", "Z0")
	_f.VMOVQ("X0", T3)
	_f.VALIGNQ("$1", "Z0", "Z0", "Z0")
	_f.VMOVQ("X0", T4)
	_f.VALIGNQ("$1", "Z0", "Z0", "Z0")
	_f.XORQ(T0, T0)

	PH := _f.Pop(&registers)
	PL := amd64.AX
	_f.MOVQ(_f.qInv0(), amd64.DX)
	_f.MULXQ(T1, amd64.DX, PH)
	_f.MULXQ(_f.qAt(0), PL, PH)
	_f.ADDQ(PL, T1)
	_f.ADCQ(PH, T2)
	_f.MULXQ(_f.qAt(2), PL, PH)
	_f.ADCQ(PL, T3)
	_f.ADCQ(PH, T4)
	_f.ADCQ("$0", T0)
	_f.MULXQ(_f.qAt(1), PL, PH)
	_f.ADDQ(PL, T2)
	_f.ADCQ(PH, T3)
	_f.MULXQ(_f.qAt(3), PL, PH)
	_f.ADCQ(PL, T4)
	_f.ADCQ(PH, T0)
	_f.ADCQ("$0", T1)

	_f.MOVQ(_f.qInv0(), amd64.DX)
	_f.MULXQ(T2, amd64.DX, PH)

	_f.MULXQ(_f.qAt(0), PL, PH)
	_f.ADDQ(PL, T2)
	_f.ADCQ(PH, T3)
	_f.MULXQ(_f.qAt(2), PL, PH)
	_f.ADCQ(PL, T4)
	_f.ADCQ(PH, T0)
	_f.ADCQ("$0", T1)
	_f.MULXQ(_f.qAt(1), PL, PH)
	_f.ADDQ(PL, T3)
	_f.ADCQ(PH, T4)
	_f.MULXQ(_f.qAt(3), PL, PH)
	_f.ADCQ(PL, T0)
	_f.ADCQ(PH, T1)
	_f.ADCQ("$0", T2)

	_f.MOVQ(_f.qInv0(), amd64.DX)

	_f.MULXQ(T3, amd64.DX, PH)

	_f.MULXQ(_f.qAt(0), PL, PH)
	_f.ADDQ(PL, T3)
	_f.ADCQ(PH, T4)
	_f.MULXQ(_f.qAt(2), PL, PH)
	_f.ADCQ(PL, T0)
	_f.ADCQ(PH, T1)
	_f.ADCQ("$0", T2)
	_f.MULXQ(_f.qAt(1), PL, PH)
	_f.ADDQ(PL, T4)
	_f.ADCQ(PH, T0)
	_f.MULXQ(_f.qAt(3), PL, PH)
	_f.ADCQ(PL, T1)
	_f.ADCQ(PH, T2)
	_f.ADCQ("$0", T3)

	_f.MOVQ(_f.qInv0(), amd64.DX)

	_f.MULXQ(T4, amd64.DX, PH)

	_f.MULXQ(_f.qAt(0), PL, PH)
	_f.ADDQ(PL, T4)
	_f.ADCQ(PH, T0)
	_f.MULXQ(_f.qAt(2), PL, PH)
	_f.ADCQ(PL, T1)
	_f.ADCQ(PH, T2)
	_f.ADCQ("$0", T3)
	_f.MULXQ(_f.qAt(1), PL, PH)
	_f.ADDQ(PL, T0)
	_f.ADCQ(PH, T1)
	_f.MULXQ(_f.qAt(3), PL, PH)
	_f.ADCQ(PL, T2)
	_f.ADCQ(PH, T3)
	_f.ADCQ("$0", T4)

	// Add the remaining 5 qwords (9 dwords) from zmm0

	_f.VMOVQ("X0", PL)
	_f.ADDQ(PL, T0)
	_f.VALIGNQ("$1", "Z0", "Z0", "Z0")
	_f.VMOVQ("X0", PL)
	_f.ADCQ(PL, T1)
	_f.VALIGNQ("$1", "Z0", "Z0", "Z0")
	_f.VMOVQ("X0", PL)
	_f.ADCQ(PL, T2)
	_f.VALIGNQ("$1", "Z0", "Z0", "Z0")
	_f.VMOVQ("X0", PL)
	_f.ADCQ(PL, T3)
	_f.VALIGNQ("$1", "Z0", "Z0", "Z0")
	_f.VMOVQ("X0", PL)
	_f.ADCQ(PL, T4)

	_f.Comment("Barrett reduction; see Handbook of Applied Cryptography, Algorithm 14.42.")
	_f.MOVQ(T3, amd64.AX)
	_f.SHRQw("$32", T4, amd64.AX)
	_f.MOVQ(_f.mu(), amd64.DX)
	_f.MULQ(amd64.DX)

	_f.MULXQ(_f.qAt(0), PL, PH)
	_f.SUBQ(PL, T0)
	_f.SBBQ(PH, T1)
	_f.MULXQ(_f.qAt(2), PL, PH)
	_f.SBBQ(PL, T2)
	_f.SBBQ(PH, T3)
	_f.SBBQ("$0", T4)
	_f.MULXQ(_f.qAt(1), PL, PH)
	_f.SUBQ(PL, T1)
	_f.SBBQ(PH, T2)
	_f.MULXQ(_f.qAt(3), PL, PH)
	_f.SBBQ(PL, T3)
	_f.SBBQ(PH, T4)

	_f.Comment("we need up to 2 conditional substractions to be < q")

	PZ := _f.Pop(&registers)
	_f.MOVQ("res+0(FP)", PZ)
	t := []amd64.Register{T0, T1, T2, T3}
	_f.Mov(t, PZ)

	// sub q
	_f.SUBQ(_f.qAt(0), T0)
	_f.SBBQ(_f.qAt(1), T1)
	_f.SBBQ(_f.qAt(2), T2)
	_f.SBBQ(_f.qAt(3), T3)
	_f.SBBQ("$0", T4)

	// if borrow, we go to done
	_f.JCS(done)

	_f.Mov(t, PZ)

	_f.SUBQ(_f.qAt(0), T0)
	_f.SBBQ(_f.qAt(1), T1)
	_f.SBBQ(_f.qAt(2), T2)
	_f.SBBQ(_f.qAt(3), T3)
	_f.SBBQ("$0", T4)

	_f.JCS(done)

	_f.Mov(t, PZ)

	_f.LABEL(done)

	_f.RET()
}

func (_f *FFAmd64) generateMulVecW4(funcName string) {
	scalarMul := funcName != "mulVec"

	const argSize = 6 * 8
	const minStackSize = 1*8 + 4*8
	stackSize := _f.StackSize(4-1+4 /* this is incorrect but minStackSize > anyway */, 2, minStackSize)
	registers := _f.FnHeader(funcName, stackSize, argSize, amd64.AX, amd64.DX)
	registers.UnsafePush(amd64.R15)
	defer _f.AssertCleanStack(stackSize, minStackSize)

	// to simplify the generated assembly, we only handle n/16 (and do blocks of 16 muls).
	// that is if n%16 != 0, we let the caller (Go) handle the remaining elements.
	LEN := _f.Pop(&registers, true)
	PZ := _f.Pop(&registers)
	PX := _f.Pop(&registers)
	PY := _f.Pop(&registers)

	// we put q words on the stack, so that we don't need to clobber R15 with global memory access.
	_q := _f.PopN(&registers, true)
	for i := 0; i < _f.NbWords; i++ {
		_f.MOVQ(fmt.Sprintf("$const_q%d", i), amd64.AX)
		_f.MOVQ(amd64.AX, _q[i])
	}
	_f.SetQStack(_q)
	defer func() {
		_f.Push(&registers, _q...)
		_f.UnsetQStack()
	}()

	zi := func(i int) amd64.Register {
		return amd64.Register("Z" + strconv.Itoa(i))
	}

	// AVX_MUL_Q_LO:
	AVX_MUL_Q_LO, err := _f.DefineFn("AVX_MUL_Q_LO")
	if err != nil {
		AVX_MUL_Q_LO = _f.Define("AVX_MUL_Q_LO", 0, func(args ...any) {
			for i := 0; i < 4; i++ {
				_f.VPMULUDQ_BCST(_f.qAt_u32(i), "Z9", zi(10+i))
				_f.VPADDQ(zi(10+i), zi(i), zi(i))
			}
		}, true)
	}

	// AVX_MUL_Q_HI:
	AVX_MUL_Q_HI := _f.Define("AVX_MUL_Q_HI", 0, func(args ...any) {
		for i := 0; i < 4; i++ {
			_f.VPMULUDQ_BCST(_f.qAt_u32(i+4), "Z9", zi(14+i))
			_f.VPADDQ(zi(14+i), zi(i+4), zi(i+4))
		}
	}, true)

	SHIFT_ADD_AND := _f.Define("SHIFT_ADD_AND", 4, func(args ...any) {
		in0 := args[0]
		in1 := args[1]
		in2 := args[2]
		in3 := args[3]
		_f.VPSRLQ("$32", in0, in1)
		_f.VPADDQ(in1, in2, in2)
		_f.VPANDQ(in3, in2, in0)
	}, true)

	// CARRY1:
	CARRY1 := _f.Define("CARRY1", 0, func(args ...any) {
		for i := 0; i < 4; i++ {
			SHIFT_ADD_AND(zi(i), zi(10+i), zi(i+1), "Z8")
		}
	}, true)

	CARRY2 := _f.Define("CARRY2", 0, func(args ...any) {
		for i := 0; i < 3; i++ {
			SHIFT_ADD_AND(zi(i+4), zi(14+i), zi(i+5), "Z8")
		}
		_f.VPSRLQ("$32", "Z7", "Z7")
	}, true)

	// CARRY3:
	CARRY3 := _f.Define("CARRY3", 0, func(args ...any) {
		for i := 0; i < 4; i++ {
			_f.VPSRLQ("$32", zi(i), zi(10+i))
			_f.VPANDQ("Z8", zi(i), zi(i))
			_f.VPADDQ(zi(10+i), zi(i+1), zi(i+1))
		}
	}, true)

	// CARRY4:
	CARRY4 := _f.Define("CARRY4", 0, func(args ...any) {
		for i := 0; i < 3; i++ {
			_f.VPSRLQ("$32", zi(i+4), zi(14+i))
			_f.VPANDQ("Z8", zi(i+4), zi(i+4))
			_f.VPADDQ(zi(14+i), zi(i+5), zi(i+5))
		}
	}, true)

	// we use the same registers as defined in the mul.
	t := _f.PopN(&registers)
	y := _f.PopN(&registers)
	tr := _f.Pop(&registers)
	A := amd64.BP // note, BP is used in the mul defines.

	var mulWord0, mulWordN defineFn
	{
		x := func(i int) amd64.Register {
			return y[i]
		}
		// This part is identical to the element_mul function
		// but we need to redefine to avoid hardcoding the registers values for q, t and x.
		mac, err := _f.DefineFn("MACC")
		if err != nil {
			panic(err)
		}
		divShift := _f.Define("DIV_SHIFT_VEC", 0, func(_ ...any) {
			// m := t[0]*q'[0] mod W
			m := amd64.DX
			_f.MOVQ(_f.qInv0(), m)
			_f.IMULQ(t[0], m)

			// clear the carry flags
			_f.XORQ(amd64.AX, amd64.AX)

			// C,_ := t[0] + m*q[0]
			_f.MULXQ(_f.qAt(0), amd64.AX, tr)
			_f.ADCXQ(t[0], amd64.AX)
			_f.MOVQ(tr, t[0])

			// for j=1 to N-1
			//
			//	(C,t[j-1]) := t[j] + m*q[j] + C
			for j := 1; j < _f.NbWords; j++ {
				mac(t[j], t[j-1], amd64.Register(_f.qAt(j)))
			}

			_f.MOVQ(0, amd64.AX)
			_f.ADCXQ(amd64.AX, t[_f.NbWordsLastIndex])
			_f.ADOXQ(A, t[_f.NbWordsLastIndex])

		}, true)

		mulWord0 = _f.Define("MUL_WORD_0_VEC", 0, func(_ ...any) {
			_f.XORQ(amd64.AX, amd64.AX)
			// for j=0 to N-1
			//    (A,t[j])  := t[j] + x[j]*y[i] + A
			for j := 0; j < _f.NbWords; j++ {
				if j == 0 && _f.NbWords == 1 {
					_f.MULXQ(x(j), t[j], A)
				} else if j == 0 {
					_f.MULXQ(x(j), t[j], t[j+1])
				} else {
					highBits := A
					if j != _f.NbWordsLastIndex {
						highBits = t[j+1]
					}
					_f.MULXQ(x(j), amd64.AX, highBits)
					_f.ADOXQ(amd64.AX, t[j])
				}
			}
			_f.MOVQ(0, amd64.AX)
			_f.ADOXQ(amd64.AX, A)
			divShift()
		}, true)

		mulWordN = _f.Define("MUL_WORD_N_VEC", 0, func(args ...any) {
			_f.XORQ(amd64.AX, amd64.AX)
			// for j=0 to N-1
			//    (A,t[j])  := t[j] + x[j]*y[i] + A
			_f.MULXQ(x(0), amd64.AX, A)
			_f.ADOXQ(amd64.AX, t[0])
			for j := 1; j < _f.NbWords; j++ {
				mac(A, t[j], amd64.Register(x(j)))
			}
			_f.MOVQ(0, amd64.AX)
			_f.ADCXQ(amd64.AX, A)
			_f.ADOXQ(amd64.AX, A)
			divShift()
		}, true)
	}

	zIndex := 0

	loadInput := func() {
		if scalarMul {
			return
		}
		_f.Comment(fmt.Sprintf("load input y[%d]", zIndex))
		_f.Mov(PY, y, zIndex*4)
	}

	mulXi := func(wordIndex int) {
		_f.Comment(fmt.Sprintf("z[%d] -> y * x[%d]", zIndex, wordIndex))
		if wordIndex == 0 {
			mulWord0()
		} else {
			_f.MOVQ(amd64.Register(PX.At(wordIndex)), amd64.DX)
			mulWordN()
		}
	}

	storeOutput := func() {
		scratch := []amd64.Register{A, tr, amd64.AX, amd64.DX}
		_f.ReduceElement(t, scratch, false)

		_f.Comment(fmt.Sprintf("store output z[%d]", zIndex))
		_f.Mov(t, PZ, 0, zIndex*4)
		if zIndex == 7 {
			_f.ADDQ("$288", PX)
		} else {
			_f.ADDQ("$32", PX)
			_f.MOVQ(amd64.Register(PX.At(0)), amd64.DX)
		}
		zIndex++
	}

	done := _f.NewLabel("done")
	loop := _f.NewLabel("loop")

	_f.MOVQ("res+0(FP)", PZ)
	_f.MOVQ("a+8(FP)", PX)
	_f.MOVQ("b+16(FP)", PY)
	_f.MOVQ("n+24(FP)", tr)

	if scalarMul {
		// for scalar mul we move the scalar only once in registers.
		_f.Mov(PY, y)
	}

	// we process 16 elements at a time, Go caller divided len by 16.
	_f.MOVQ(tr, LEN)

	_f.Comment("Create mask for low dword in each qword")

	_f.VPCMPEQB("Y8", "Y8", "Y8")
	_f.VPMOVZXDQ("Y8", "Z8")
	_f.MOVQ("$0x5555", amd64.DX)
	_f.KMOVD(amd64.DX, "K1")

	_f.LABEL(loop)
	// f.MOVQ(LEN, tr)
	_f.TESTQ(tr, tr)
	_f.JEQ(done, "n == 0, we are done")

	_f.MOVQ(amd64.Register(PX.At(0)), amd64.DX)
	_f.VMOVDQU64("256+0*64("+PX+")", "Z16")
	_f.VMOVDQU64("256+1*64("+PX+")", "Z17")
	_f.VMOVDQU64("256+2*64("+PX+")", "Z18")
	_f.VMOVDQU64("256+3*64("+PX+")", "Z19")

	loadInput()
	if scalarMul {
		_f.VMOVDQU64("0("+PY+")", "Z24")
		_f.VMOVDQU64("0("+PY+")", "Z25")
		_f.VMOVDQU64("0("+PY+")", "Z26")
		_f.VMOVDQU64("0("+PY+")", "Z27")
	} else {
		_f.VMOVDQU64("256+0*64("+PY+")", "Z24")
		_f.VMOVDQU64("256+1*64("+PY+")", "Z25")
		_f.VMOVDQU64("256+2*64("+PY+")", "Z26")
		_f.VMOVDQU64("256+3*64("+PY+")", "Z27")
	}

	_f.Comment("Transpose and expand x and y")

	// Step 1

	_f.VSHUFI64X2("$0x88", "Z17", "Z16", "Z20")
	_f.VSHUFI64X2("$0xdd", "Z17", "Z16", "Z22")
	_f.VSHUFI64X2("$0x88", "Z19", "Z18", "Z21")
	_f.VSHUFI64X2("$0xdd", "Z19", "Z18", "Z23")

	_f.VSHUFI64X2("$0x88", "Z25", "Z24", "Z28")
	_f.VSHUFI64X2("$0xdd", "Z25", "Z24", "Z30")
	_f.VSHUFI64X2("$0x88", "Z27", "Z26", "Z29")
	_f.VSHUFI64X2("$0xdd", "Z27", "Z26", "Z31")

	// Step 2

	_f.VPERMQ("$0xd8", "Z20", "Z20")
	_f.VPERMQ("$0xd8", "Z21", "Z21")
	_f.VPERMQ("$0xd8", "Z22", "Z22")
	_f.VPERMQ("$0xd8", "Z23", "Z23")

	mulXi(0)

	_f.VPERMQ("$0xd8", "Z28", "Z28")
	_f.VPERMQ("$0xd8", "Z29", "Z29")
	_f.VPERMQ("$0xd8", "Z30", "Z30")
	_f.VPERMQ("$0xd8", "Z31", "Z31")

	// Step 3

	for i := 20; i <= 23; i++ {
		_f.VSHUFI64X2("$0xd8", zi(i), zi(i), zi(i))
	}

	mulXi(1)

	for i := 28; i <= 31; i++ {
		_f.VSHUFI64X2("$0xd8", zi(i), zi(i), zi(i))
	}

	// Step 4

	_f.VSHUFI64X2("$0x44", "Z21", "Z20", "Z16")
	_f.VSHUFI64X2("$0xee", "Z21", "Z20", "Z18")
	_f.VSHUFI64X2("$0x44", "Z23", "Z22", "Z20")
	_f.VSHUFI64X2("$0xee", "Z23", "Z22", "Z22")

	mulXi(2)
	_f.VSHUFI64X2("$0x44", "Z29", "Z28", "Z24")
	_f.VSHUFI64X2("$0xee", "Z29", "Z28", "Z26")
	_f.VSHUFI64X2("$0x44", "Z31", "Z30", "Z28")
	_f.VSHUFI64X2("$0xee", "Z31", "Z30", "Z30")

	_f.WriteLn("PREFETCHT0 1024(" + string(PX) + ")")

	// Step 5

	_f.VPSRLQ("$32", "Z16", "Z17")
	_f.VPSRLQ("$32", "Z18", "Z19")
	_f.VPSRLQ("$32", "Z20", "Z21")
	_f.VPSRLQ("$32", "Z22", "Z23")

	for i := 24; i <= 30; i += 2 {
		_f.VPSRLQ("$32", zi(i), zi(i+1))
	}
	mulXi(3)

	for i := 16; i <= 30; i += 2 {
		_f.VPANDQ("Z8", zi(i), zi(i))
	}

	storeOutput()

	_f.Comment("For each 256-bit input value, each zmm register now represents a 32-bit input word zero-extended to 64 bits.")
	_f.Comment("Multiply y by doubleword 0 of x")

	for i := 0; i < 8; i++ {
		_f.VPMULUDQ("Z16", zi(24+i), zi(i))
		if i == 4 {
			if !scalarMul {
				_f.WriteLn("PREFETCHT0 1024(" + string(PY) + ")")
			}
		}
	}

	loadInput()

	_f.VPMULUDQ_BCST("qInvNeg+32(FP)", "Z0", "Z9")

	for i := 0; i < 4; i++ {
		_f.VPSRLQ("$32", zi(i), zi(10+i))
		_f.VPANDQ("Z8", zi(i), zi(i))
		_f.VPADDQ(zi(10+i), zi(i+1), zi(i+1))

	}

	mulXi(0)

	for i := 0; i < 3; i++ {
		_f.VPSRLQ("$32", zi(4+i), zi(14+i))
		_f.VPANDQ("Z8", zi(4+i), zi(4+i))
		_f.VPADDQ(zi(14+i), zi(5+i), zi(5+i))

	}

	for i := 0; i < 4; i++ {
		_f.VPMULUDQ_BCST(_f.qAt_u32(i), "Z9", zi(10+i))
		_f.VPADDQ(zi(10+i), zi(i), zi(i))
	}

	mulXi(1)

	_f.VPMULUDQ_BCST(_f.qAt_u32(4), "Z9", "Z14")
	_f.VPADDQ("Z14", "Z4", "Z4")

	_f.VPMULUDQ_BCST(_f.qAt_u32(5), "Z9", "Z15")
	_f.VPADDQ("Z15", "Z5", "Z5")

	_f.VPMULUDQ_BCST(_f.qAt_u32(6), "Z9", "Z16")
	_f.VPADDQ("Z16", "Z6", "Z6")

	_f.VPMULUDQ_BCST(_f.qAt_u32(7), "Z9", "Z10")
	_f.VPADDQ("Z10", "Z7", "Z7")

	CARRY1()

	mulXi(2)

	for i := 0; i < 3; i++ {
		SHIFT_ADD_AND(zi(4+i), zi(14+i), zi(5+i), "Z8")
	}
	_f.VPSRLQ("$32", "Z7", "Z7")

	_f.Comment("Process doubleword 1 of x")

	for i := 0; i < 4; i++ {
		_f.VPMULUDQ("Z17", zi(24+i), zi(10+i))
		_f.VPADDQ(zi(10+i), zi(i), zi(i))

	}

	mulXi(3)

	for i := 0; i < 4; i++ {
		_f.VPMULUDQ("Z17", zi(28+i), zi(14+i))
		_f.VPADDQ(zi(14+i), zi(4+i), zi(4+i))

	}

	_f.VPMULUDQ_BCST("qInvNeg+32(FP)", "Z0", "Z9")

	storeOutput()

	_f.Comment("Move high dwords to zmm10-16, add each to the corresponding low dword (propagate 32-bit carries)")

	for i := 0; i < 3; i++ {
		_f.VPSRLQ("$32", zi(i), zi(10+i))
		_f.VPANDQ("Z8", zi(i), zi(i))
		_f.VPADDQ(zi(10+i), zi(i+1), zi(i+1))
	}
	loadInput()

	_f.VPSRLQ("$32", "Z3", "Z13")
	_f.VPANDQ("Z8", "Z3", "Z3")
	_f.VPADDQ("Z13", "Z4", "Z4")

	CARRY4()
	mulXi(0)

	AVX_MUL_Q_LO()

	AVX_MUL_Q_HI()
	mulXi(1)

	CARRY1()

	CARRY2()
	mulXi(2)

	_f.Comment("Process doubleword 2 of x")

	for i := 0; i < 4; i++ {
		_f.VPMULUDQ("Z18", zi(24+i), zi(10+i))
		_f.VPADDQ(zi(10+i), zi(i), zi(i))
	}

	for i := 0; i < 4; i++ {
		_f.VPMULUDQ("Z18", zi(28+i), zi(14+i))
		_f.VPADDQ(zi(14+i), zi(4+i), zi(4+i))

	}

	_f.VPMULUDQ_BCST("qInvNeg+32(FP)", "Z0", "Z9")

	mulXi(3)

	_f.Comment("Move high dwords to zmm10-16, add each to the corresponding low dword (propagate 32-bit carries)")
	CARRY3()

	storeOutput()
	loadInput()

	CARRY4()

	AVX_MUL_Q_LO()

	mulXi(0)
	AVX_MUL_Q_HI()

	CARRY1()
	CARRY2()

	_f.Comment("Process doubleword 3 of x")

	for i := 0; i < 4; i++ {
		_f.VPMULUDQ("Z19", zi(24+i), zi(10+i))
		_f.VPADDQ(zi(10+i), zi(i), zi(i))

	}

	mulXi(1)

	for i := 0; i < 4; i++ {
		_f.VPMULUDQ("Z19", zi(28+i), zi(14+i))
		_f.VPADDQ(zi(14+i), zi(4+i), zi(4+i))

	}
	mulXi(2)
	_f.VPMULUDQ_BCST("qInvNeg+32(FP)", "Z0", "Z9")

	// Move high dwords to zmm10-16, add each to the corresponding low dword (propagate 32-bit carries)
	CARRY3()
	CARRY4()
	mulXi(3)

	AVX_MUL_Q_LO()

	AVX_MUL_Q_HI()

	storeOutput()

	_f.Comment("Propagate carries and shift down by one dword")
	CARRY1()

	CARRY2()

	loadInput()

	_f.Comment("Process doubleword 4 of x")

	for i := 0; i < 4; i++ {
		_f.VPMULUDQ("Z20", zi(24+i), zi(10+i))
		_f.VPADDQ(zi(10+i), zi(i), zi(i))

	}
	mulXi(0)
	for i := 0; i < 4; i++ {
		_f.VPMULUDQ("Z20", zi(28+i), zi(14+i))
		_f.VPADDQ(zi(14+i), zi(4+i), zi(4+i))

	}

	_f.VPMULUDQ_BCST("qInvNeg+32(FP)", "Z0", "Z9")
	mulXi(1)

	_f.Comment("Move high dwords to zmm10-16, add each to the corresponding low dword (propagate 32-bit carries)")

	CARRY3()

	CARRY4()
	mulXi(2)

	_f.Comment("zmm7 keeps all 64 bits")

	AVX_MUL_Q_LO()

	AVX_MUL_Q_HI()

	mulXi(3)

	_f.Comment("Propagate carries and shift down by one dword")

	CARRY1()

	CARRY2()

	storeOutput()

	_f.Comment("Process doubleword 5 of x")

	for i := 0; i < 4; i++ {
		_f.VPMULUDQ("Z21", zi(24+i), zi(10+i))
		_f.VPADDQ(zi(10+i), zi(i), zi(i))

	}
	loadInput()
	for i := 0; i < 4; i++ {
		_f.VPMULUDQ("Z21", zi(28+i), zi(14+i))
		_f.VPADDQ(zi(14+i), zi(4+i), zi(4+i))

	}

	mulXi(0)

	_f.VPMULUDQ_BCST("qInvNeg+32(FP)", "Z0", "Z9")

	_f.Comment("Move high dwords to zmm10-16, add each to the corresponding low dword (propagate 32-bit carries)")
	CARRY3()

	CARRY4()

	mulXi(1)

	AVX_MUL_Q_LO()

	AVX_MUL_Q_HI()

	mulXi(2)

	CARRY1()

	CARRY2()

	mulXi(3)

	_f.Comment("Process doubleword 6 of x")

	for i := 0; i < 8; i++ {
		_f.VPMULUDQ("Z22", zi(24+i), zi(10+i))
		_f.VPADDQ(zi(10+i), zi(i), zi(i))
	}

	_f.VPMULUDQ_BCST("qInvNeg+32(FP)", "Z0", "Z9")

	storeOutput()

	_f.Comment("Move high dwords to zmm10-16, add each to the corresponding low dword (propagate 32-bit carries)")
	CARRY3()
	loadInput()
	CARRY4()

	mulXi(0)

	AVX_MUL_Q_LO()

	AVX_MUL_Q_HI()

	mulXi(1)

	CARRY1()

	CARRY2()

	mulXi(2)

	_f.Comment("Process doubleword 7 of x")
	for i := 0; i < 4; i++ {
		_f.VPMULUDQ("Z23", zi(24+i), zi(10+i))
		_f.VPADDQ(zi(10+i), zi(i), zi(i))

	}

	for i := 0; i < 4; i++ {
		_f.VPMULUDQ("Z23", zi(28+i), zi(14+i))
		_f.VPADDQ(zi(14+i), zi(4+i), zi(4+i))

	}
	_f.VPMULUDQ_BCST("qInvNeg+32(FP)", "Z0", "Z9")

	mulXi(3)

	CARRY3()
	storeOutput()
	CARRY4()

	loadInput()

	AVX_MUL_Q_LO()

	AVX_MUL_Q_HI()

	mulXi(0)

	CARRY1()

	CARRY2()

	mulXi(1)

	_f.Comment("Conditional subtraction of the modulus")

	for i := 0; i < 8; i++ {
		_f.VPERMD_BCST_Z(_f.qAt_u32(i), "Z8", "K1", zi(10+i))
	}

	for i := 0; i < 8; i++ {
		_f.VPSUBQ(zi(10+i), zi(i), zi(10+i))
		if i > 0 {
			_f.VPSUBQ(zi(20+i-1), zi(10+i), zi(10+i))
		}
		if i != 7 {
			_f.VPSRLQ("$63", zi(10+i), zi(20+i))
			_f.VPANDQ("Z8", zi(10+i), zi(10+i))
		}

	}

	_f.VPMOVQ2M("Z17", "K2")
	_f.KNOTB("K2", "K2")

	for i := 0; i < 8; i++ {
		_f.VMOVDQU64k(zi(10+i), "K2", zi(i))
		if i == 4 {
			mulXi(2)
		}
	}

	_f.Comment("Transpose results back")
	// patterns+40(FP) contains pointer to the patterns array;
	ax := amd64.AX
	_f.MOVQ("patterns+40(FP)", ax)
	_f.VMOVDQU64(ax.At(0), amd64.Z15)
	_f.WriteLn("VALIGND $0, Z15, Z11, Z11")
	_f.VMOVDQU64(ax.At(8), amd64.Z15)
	_f.WriteLn("VALIGND $0, Z15, Z12, Z12")
	_f.VMOVDQU64(ax.At(16), amd64.Z15)
	_f.WriteLn("VALIGND $0, Z15, Z13, Z13")
	_f.VMOVDQU64(ax.At(24), amd64.Z15)
	_f.WriteLn("VALIGND $0, Z15, Z14, Z14")

	for i := 0; i < 4; i++ {
		_f.VPSLLQ("$32", zi(2*i+1), zi(2*i+1))
		_f.VPORQ(zi(2*i+1), zi(2*i), zi(i))
	}

	_f.VMOVDQU64("Z0", "Z4")
	_f.VMOVDQU64("Z2", "Z6")

	mulXi(3)
	_f.VPERMT2Q("Z1", "Z11", "Z0")
	_f.VPERMT2Q("Z4", "Z12", "Z1")
	_f.VPERMT2Q("Z3", "Z11", "Z2")
	_f.VPERMT2Q("Z6", "Z12", "Z3")

	// Step 3
	storeOutput()

	_f.VMOVDQU64("Z0", "Z4")
	_f.VMOVDQU64("Z1", "Z5")
	_f.VPERMT2Q("Z2", "Z13", "Z0")
	_f.VPERMT2Q("Z4", "Z14", "Z2")
	_f.VPERMT2Q("Z3", "Z13", "Z1")
	_f.VPERMT2Q("Z5", "Z14", "Z3")

	_f.Comment("Save AVX-512 results")

	_f.VMOVDQU64("Z0", "256+0*64("+PZ+")")
	_f.VMOVDQU64("Z2", "256+1*64("+PZ+")")
	_f.VMOVDQU64("Z1", "256+2*64("+PZ+")")
	_f.VMOVDQU64("Z3", "256+3*64("+PZ+")")
	_f.ADDQ("$512", PZ)

	if !scalarMul {
		_f.ADDQ("$512", PY)
	}

	_f.MOVQ(LEN, tr)
	_f.DECQ(tr, "decrement n")
	_f.MOVQ(tr, LEN)
	_f.JMP(loop)

	_f.LABEL(done)

	_f.RET()

	_f.UnsafePush(&registers, LEN, PZ)

}
