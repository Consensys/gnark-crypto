// Copyright 2020 ConsenSys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package amd64

import (
	"fmt"
	"strconv"

	"github.com/consensys/bavard/amd64"
)

// addVec res = a + b
// func addVec(res, a, b *{{.ElementName}}, n uint64)
func (f *FFAmd64) generateAddVec() {
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

	// reduce a
	f.ReduceElement(a, t)

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
func (f *FFAmd64) generateSubVec() {
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

	// reduce a
	f.Comment("reduce (a-b) mod q")
	f.LabelRegisters("q", q...)
	for i := 0; i < f.NbWords; i++ {
		f.MOVQ(fmt.Sprintf("q%d", i), q[i])
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

// scalarMulVec res = a * b
// func scalarMulVec(res, a, b *{{.ElementName}}, n uint64)
func (f *FFAmd64) generateScalarMulVec() {
	f.Comment("scalarMulVec(res, a, b *Element, n uint64) res[0...n] = a[0...n] * b")

	const argSize = 4 * 8
	const minStackSize = 7 * 8 // 2 slices (3 words each) + pointer to the scalar
	stackSize := f.StackSize(f.NbWords*2+3, 2, minStackSize)
	reserved := []amd64.Register{amd64.DX, amd64.AX}
	registers := f.FnHeader("scalarMulVec", stackSize, argSize, reserved...)
	defer f.AssertCleanStack(stackSize, minStackSize)

	// labels & registers we need
	noAdx := f.NewLabel("noAdx")
	loop := f.NewLabel("loop")
	done := f.NewLabel("done")

	t := registers.PopN(f.NbWords)
	scalar := registers.PopN(f.NbWords)

	addrB := registers.Pop()
	addrA := registers.Pop()
	addrRes := addrB
	len := registers.Pop()

	// check ADX instruction support
	f.CMPB("·supportAdx(SB)", 1)
	f.JNE(noAdx)

	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("b+16(FP)", addrB)
	f.MOVQ("n+24(FP)", len)

	// we store b, the scalar, fully in registers
	f.LabelRegisters("scalar", scalar...)
	f.Mov(addrB, scalar)

	xat := func(i int) string {
		return string(scalar[i])
	}

	f.MOVQ("res+0(FP)", addrRes)

	f.LABEL(loop)
	f.TESTQ(len, len)
	f.JEQ(done, "n == 0, we are done")

	yat := func(i int) string {
		return addrA.At(i)
	}

	f.Comment("TODO @gbotrel this is generated from the same macro as the unit mul, we should refactor this in a single asm function")

	f.MulADX(&registers, xat, yat, t)

	// registers.Push(addrA)

	// reduce; we need at least 4 extra registers
	registers.Push(amd64.AX, amd64.DX)
	f.Comment("reduce t mod q")
	f.Reduce(&registers, t)
	f.Mov(t, addrRes)

	f.Comment("increment pointers to visit next element")
	f.ADDQ("$32", addrA)
	f.ADDQ("$32", addrRes)
	f.DECQ(len, "decrement n")
	f.JMP(loop)

	f.LABEL(done)
	f.RET()

	// no ADX support
	f.LABEL(noAdx)

	f.MOVQ("n+24(FP)", amd64.DX)

	f.MOVQ("res+0(FP)", amd64.AX)
	f.MOVQ(amd64.AX, "(SP)")
	f.MOVQ(amd64.DX, "8(SP)")  // len
	f.MOVQ(amd64.DX, "16(SP)") // cap
	f.MOVQ("a+8(FP)", amd64.AX)
	f.MOVQ(amd64.AX, "24(SP)")
	f.MOVQ(amd64.DX, "32(SP)") // len
	f.MOVQ(amd64.DX, "40(SP)") // cap
	f.MOVQ("b+16(FP)", amd64.AX)
	f.MOVQ(amd64.AX, "48(SP)")
	f.WriteLn("CALL ·scalarMulVecGeneric(SB)")
	f.RET()

}

// sumVec res = sum(a[0...n])
func (f *FFAmd64) generateSumVec() {
	f.Comment("sumVec(res, a *Element, n uint64) res = sum(a[0...n])")

	const argSize = 3 * 8
	stackSize := f.StackSize(f.NbWords*3+2, 0, 0)
	registers := f.FnHeader("sumVec", stackSize, argSize, amd64.DX, amd64.AX)
	defer f.AssertCleanStack(stackSize, 0)

	f.WriteLn(`
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
	addrA := f.Pop(&registers)
	n := f.Pop(&registers)
	nMod8 := f.Pop(&registers)

	loop := f.NewLabel("loop8by8")
	done := f.NewLabel("done")
	loopSingle := f.NewLabel("loop_single")
	accumulate := f.NewLabel("accumulate")

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
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("n+16(FP)", n)

	f.Comment("initialize accumulators Z0, Z1, Z2, Z3, Z4, Z5, Z6, Z7")
	f.VXORPS(Z0, Z0, Z0)
	f.VMOVDQA64(Z0, Z1)
	f.VMOVDQA64(Z0, Z2)
	f.VMOVDQA64(Z0, Z3)
	f.VMOVDQA64(Z0, Z4)
	f.VMOVDQA64(Z0, Z5)
	f.VMOVDQA64(Z0, Z6)
	f.VMOVDQA64(Z0, Z7)

	// note: we don't need to handle the case n==0; handled by caller already.
	// f.TESTQ(n, n)
	// f.JEQ(done, "n == 0, we are done")

	f.LabelRegisters("n % 8", nMod8)
	f.LabelRegisters("n / 8", n)
	f.MOVQ(n, nMod8)
	f.ANDQ("$7", nMod8) // nMod8 = n % 8
	f.SHRQ("$3", n)     // len = n / 8

	f.LABEL(loopSingle)
	f.TESTQ(nMod8, nMod8)
	f.JEQ(loop, "n % 8 == 0, we are going to loop over 8 by 8")

	f.VPMOVZXDQ("0("+addrA+")", Z8)
	f.VPADDQ(Z8, Z0, Z0)
	f.ADDQ("$32", addrA)

	f.DECQ(nMod8, "decrement nMod8")
	f.JMP(loopSingle)

	f.Push(&registers, nMod8) // we don't need tmp0

	f.LABEL(loop)
	f.TESTQ(n, n)
	f.JEQ(accumulate, "n == 0, we are going to accumulate")

	for i := 0; i < 8; i++ {
		r := fmt.Sprintf("Z%d", i+8)
		f.VPMOVZXDQ(fmt.Sprintf("%d*32("+string(addrA)+")", i), r)
	}
	f.WriteLn(fmt.Sprintf("PREFETCHT0 256(%[1]s)", addrA))
	for i := 0; i < 8; i++ {
		r := fmt.Sprintf("Z%d", i)
		f.VPADDQ(fmt.Sprintf("Z%d", i+8), r, r)
	}

	f.Comment("increment pointers to visit next 8 elements")
	f.ADDQ("$256", addrA)
	f.DECQ(n, "decrement n")
	f.JMP(loop)

	f.Push(&registers, n, addrA)

	f.LABEL(accumulate)

	f.Comment("accumulate the 8 Z registers into Z0")
	f.VPADDQ(Z7, Z6, Z6)
	f.VPADDQ(Z6, Z5, Z5)
	f.VPADDQ(Z5, Z4, Z4)
	f.VPADDQ(Z4, Z3, Z3)
	f.VPADDQ(Z3, Z2, Z2)
	f.VPADDQ(Z2, Z1, Z1)
	f.VPADDQ(Z1, Z0, Z0)

	w0l := f.Pop(&registers)
	w0h := f.Pop(&registers)
	w1l := f.Pop(&registers)
	w1h := f.Pop(&registers)
	w2l := f.Pop(&registers)
	w2h := f.Pop(&registers)
	w3l := f.Pop(&registers)
	w3h := f.Pop(&registers)
	low0h := f.Pop(&registers)
	low1h := f.Pop(&registers)
	low2h := f.Pop(&registers)
	low3h := f.Pop(&registers)

	// Propagate carries
	f.Comment("carry propagation")

	f.LabelRegisters("lo(w0)", w0l)
	f.LabelRegisters("hi(w0)", w0h)
	f.LabelRegisters("lo(w1)", w1l)
	f.LabelRegisters("hi(w1)", w1h)
	f.LabelRegisters("lo(w2)", w2l)
	f.LabelRegisters("hi(w2)", w2h)
	f.LabelRegisters("lo(w3)", w3l)
	f.LabelRegisters("hi(w3)", w3h)

	f.VMOVQ(X0, w0l)
	f.VALIGNQ("$1", Z0, Z0, Z0)
	f.VMOVQ(X0, w0h)
	f.VALIGNQ("$1", Z0, Z0, Z0)
	f.VMOVQ(X0, w1l)
	f.VALIGNQ("$1", Z0, Z0, Z0)
	f.VMOVQ(X0, w1h)
	f.VALIGNQ("$1", Z0, Z0, Z0)
	f.VMOVQ(X0, w2l)
	f.VALIGNQ("$1", Z0, Z0, Z0)
	f.VMOVQ(X0, w2h)
	f.VALIGNQ("$1", Z0, Z0, Z0)
	f.VMOVQ(X0, w3l)
	f.VALIGNQ("$1", Z0, Z0, Z0)
	f.VMOVQ(X0, w3h)

	f.LabelRegisters("lo(hi(wo))", low0h)
	f.LabelRegisters("lo(hi(w1))", low1h)
	f.LabelRegisters("lo(hi(w2))", low2h)
	f.LabelRegisters("lo(hi(w3))", low3h)

	type hilo struct {
		hi, lo amd64.Register
	}

	f.WriteLn(`#define SPLIT_LO_HI(lo, hi) \
		MOVQ hi, lo; \
		ANDQ $0xffffffff, lo; \
		SHLQ $32, lo; \
		SHRQ $32, hi; \
	`)

	for _, v := range []hilo{{w0h, low0h}, {w1h, low1h}, {w2h, low2h}, {w3h, low3h}} {
		f.WriteLn(`SPLIT_LO_HI(` + string(v.lo) + `, ` + string(v.hi) + `)`)
	}

	f.WriteLn(`
	// r0 = w0l + lo(woh)
	// r1 = carry + hi(woh) + w1l + lo(w1h)
	// r2 = carry + hi(w1h) + w2l + lo(w2h)
	// r3 = carry + hi(w2h) + w3l + lo(w3h)
	// r4 = carry + hi(w3h)
	`)
	f.XORQ(amd64.AX, amd64.AX, "clear the flags")
	f.ADOXQ(low0h, w0l)

	f.ADOXQ(low1h, w1l)
	f.ADCXQ(w0h, w1l)

	f.ADOXQ(low2h, w2l)
	f.ADCXQ(w1h, w2l)

	f.ADOXQ(low3h, w3l)
	f.ADCXQ(w2h, w3l)

	f.ADOXQ(amd64.AX, w3h)
	f.ADCXQ(amd64.AX, w3h)

	r0 := w0l
	r1 := w1l
	r2 := w2l
	r3 := w3l
	r4 := w3h

	r := []amd64.Register{r0, r1, r2, r3, r4}
	f.LabelRegisters("r", r...)
	// we don't need w0h, w1h, w2h anymore
	f.Push(&registers, w0h, w1h, w2h)
	// we don't need the low bits anymore
	f.Push(&registers, low0h, low1h, low2h, low3h)

	// Reduce using single-word Barrett
	mu := f.Pop(&registers)

	f.Comment("reduce using single-word Barrett")
	f.Comment("see see Handbook of Applied Cryptography, Algorithm 14.42.")
	f.LabelRegisters("mu=2^288 / q", mu)
	f.MOVQ(f.mu(), mu)
	f.MOVQ(r3, amd64.AX)
	f.SHRQw("$32", r4, amd64.AX)
	f.MULQ(mu, "high bits of res stored in DX")

	f.MULXQ(f.qAt(0), amd64.AX, mu)
	f.SUBQ(amd64.AX, r0)
	f.SBBQ(mu, r1)

	f.MULXQ(f.qAt(2), amd64.AX, mu)
	f.SBBQ(amd64.AX, r2)
	f.SBBQ(mu, r3)
	f.SBBQ("$0", r4)

	f.MULXQ(f.qAt(1), amd64.AX, mu)
	f.SUBQ(amd64.AX, r1)
	f.SBBQ(mu, r2)

	f.MULXQ(f.qAt(3), amd64.AX, mu)
	f.SBBQ(amd64.AX, r3)
	f.SBBQ(mu, r4)

	// we need up to 2 conditional substractions to be < q
	modReduced := f.NewLabel("modReduced")
	t := f.PopN(&registers)
	f.Mov(r[:4], t) // backup r0 to r3 (our result)

	// sub modulus
	f.SUBQ(f.qAt(0), r0)
	f.SBBQ(f.qAt(1), r1)
	f.SBBQ(f.qAt(2), r2)
	f.SBBQ(f.qAt(3), r3)
	f.SBBQ("$0", r4)

	// if borrow, we go to mod reduced
	f.JCS(modReduced)
	f.Mov(r, t)
	f.SUBQ(f.qAt(0), r0)
	f.SBBQ(f.qAt(1), r1)
	f.SBBQ(f.qAt(2), r2)
	f.SBBQ(f.qAt(3), r3)
	f.SBBQ("$0", r4)

	// if borrow, we skip to the end
	f.JCS(modReduced)
	f.Mov(r, t)

	f.LABEL(modReduced)
	addrRes := mu
	f.MOVQ("res+0(FP)", addrRes)
	f.Mov(t, addrRes)

	f.LABEL(done)

	f.RET()
	f.Push(&registers, mu)
	f.Push(&registers, w0l, w1l, w2l, w3l, w3h)
}

func (f *FFAmd64) generateInnerProduct() {
	f.Comment("innerProdVec(res, a,b *Element, n uint64) res = sum(a[0...n] * b[0...n])")

	const argSize = 4 * 8
	stackSize := f.StackSize(f.NbWords*3+2, 0, 0)
	registers := f.FnHeader("innerProdVec", stackSize, argSize, amd64.DX, amd64.AX)
	defer f.AssertCleanStack(stackSize, 0)

	// registers & labels we need
	PX := f.Pop(&registers)
	PY := f.Pop(&registers)
	LEN := f.Pop(&registers)

	loop := f.NewLabel("loop")
	done := f.NewLabel("done")
	AddPP := f.NewLabel("accumulate")

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

	// X0 := amd64.Register("X0")

	// load arguments
	f.MOVQ("a+8(FP)", PX)
	f.MOVQ("b+16(FP)", PY)
	f.MOVQ("n+24(FP)", LEN)

	// Create mask for low dword in each qword
	// vpmovzxdq	%ymm0, LSW
	f.VPCMPEQB("Y0", "Y0", "Y0")
	f.VPMOVZXDQ("Y0", LSW)

	// Clear accumulator registers
	f.VPXORQ(A0L, A0L, A0L)
	f.VMOVDQA64(A0L, A1L)
	f.VMOVDQA64(A0L, A2L)
	f.VMOVDQA64(A0L, A3L)
	f.VMOVDQA64(A0L, A4L)
	f.VMOVDQA64(A0L, A5L)
	f.VMOVDQA64(A0L, A6L)
	f.VMOVDQA64(A0L, A7L)
	f.VMOVDQA64(A0L, A0H)
	f.VMOVDQA64(A0L, A1H)
	f.VMOVDQA64(A0L, A2H)
	f.VMOVDQA64(A0L, A3H)
	f.VMOVDQA64(A0L, A4H)
	f.VMOVDQA64(A0L, A5H)
	f.VMOVDQA64(A0L, A6H)
	f.VMOVDQA64(A0L, A7H)

	// note: we don't need to handle the case n==0; handled by caller already.
	f.TESTQ(LEN, LEN)
	f.JEQ(done, "n == 0, we are done")

	f.LABEL(loop)
	f.TESTQ(LEN, LEN)
	f.JEQ(AddPP, "n == 0 we can accumulate")

	f.VPMOVZXDQ("("+PY+")", Y)

	f.ADDQ("$32", PY)

	f.VPMULUDQ_BCST("0*4("+PX+")", Y, PPL)
	f.VPSRLQ("$32", PPL, PPH)
	f.VPANDQ(LSW, PPL, PPL)
	f.VPADDQ(PPL, A0L, A0L)
	f.VPADDQ(PPH, A0H, A0H)

	f.VPMULUDQ_BCST("1*4("+PX+")", Y, PPL)
	f.VPSRLQ("$32", PPL, PPH)
	f.VPANDQ(LSW, PPL, PPL)
	f.VPADDQ(PPL, A1L, A1L)
	f.VPADDQ(PPH, A1H, A1H)

	f.VPMULUDQ_BCST("2*4("+PX+")", Y, PPL)
	f.VPSRLQ("$32", PPL, PPH)
	f.VPANDQ(LSW, PPL, PPL)
	f.VPADDQ(PPL, A2L, A2L)
	f.VPADDQ(PPH, A2H, A2H)

	f.VPMULUDQ_BCST("3*4("+PX+")", Y, PPL)
	f.VPSRLQ("$32", PPL, PPH)
	f.VPANDQ(LSW, PPL, PPL)
	f.VPADDQ(PPL, A3L, A3L)
	f.VPADDQ(PPH, A3H, A3H)

	f.VPMULUDQ_BCST("4*4("+PX+")", Y, PPL)
	f.VPSRLQ("$32", PPL, PPH)
	f.VPANDQ(LSW, PPL, PPL)
	f.VPADDQ(PPL, A4L, A4L)
	f.VPADDQ(PPH, A4H, A4H)

	f.VPMULUDQ_BCST("5*4("+PX+")", Y, PPL)
	f.VPSRLQ("$32", PPL, PPH)
	f.VPANDQ(LSW, PPL, PPL)
	f.VPADDQ(PPL, A5L, A5L)
	f.VPADDQ(PPH, A5H, A5H)

	f.VPMULUDQ_BCST("6*4("+PX+")", Y, PPL)
	f.VPSRLQ("$32", PPL, PPH)
	f.VPANDQ(LSW, PPL, PPL)
	f.VPADDQ(PPL, A6L, A6L)
	f.VPADDQ(PPH, A6H, A6H)

	f.VPMULUDQ_BCST("7*4("+PX+")", Y, PPL)
	f.VPSRLQ("$32", PPL, PPH)
	f.VPANDQ(LSW, PPL, PPL)
	f.VPADDQ(PPL, A7L, A7L)
	f.VPADDQ(PPH, A7H, A7H)

	f.ADDQ("$32", PX)

	f.DECQ(LEN, "decrement n")
	f.JMP(loop)

	f.Push(&registers, LEN, PX, PY)

	f.LABEL(AddPP)
	// Load mask register values

	f.MOVQ(uint64(0x1555), amd64.AX)
	f.KMOVD(amd64.AX, "K1")

	f.MOVQ(uint64(1), amd64.AX)
	f.KMOVD(amd64.AX, "K2")

	// ACC starts with the value of A0L

	f.VALIGND_Z("$16", ACC, ACC, "K2", "Z0") // Store least significant 32 bits of ACC
	f.KSHIFTLW("$1", "K2", "K2")

	// vpsrlq		$32, ACC, PPL
	// valignd		$2, ACC, ACC, ACC{%k1}{z}
	// vpaddq		PPL, ACC, ACC

	// vpandq		LSW, A0H, PPL
	// vpaddq		PPL, ACC, ACC

	// vpandq		LSW, A1L, PPL
	// vpaddq		PPL, ACC, ACC

	// Word 1 of z is ready
	// valignd		$15, ACC, ACC, %zmm0{%k2}
	// kshiftlw	$1, %k2, %k2

	f.VPSRLQ("$32", ACC, PPL)
	f.VALIGND_Z("$2", ACC, ACC, "K1", ACC)
	f.VPADDQ(PPL, ACC, ACC)

	f.VPANDQ(LSW, A0H, PPL)
	f.VPADDQ(PPL, ACC, ACC)

	f.VPANDQ(LSW, A1L, PPL)
	f.VPADDQ(PPL, ACC, ACC)

	// Word 1 of z is ready
	f.VALIGND("$15", ACC, ACC, "K2", "Z0")
	f.KSHIFTLW("$1", "K2", "K2")

	ADDPP := func(AxH, AyL, AyH, AzL, I amd64.Register) {
		f.VPSRLQ("$32", ACC, PPL)
		f.VALIGND_Z("$2", ACC, ACC, "K1", ACC)
		f.VPADDQ(PPL, ACC, ACC)
		f.VPSRLQ("$32", AxH, AxH)
		f.VPADDQ(AxH, ACC, ACC)
		f.VPSRLQ("$32", AyL, AyL)
		f.VPADDQ(AyL, ACC, ACC)
		f.VPANDQ(LSW, AyH, PPL)
		f.VPADDQ(PPL, ACC, ACC)
		f.VPANDQ(LSW, AzL, PPL)
		f.VPADDQ(PPL, ACC, ACC)
		f.VALIGND("$16-"+I, ACC, ACC, "K2", "Z0")
		f.KADDW("K2", "K2", "K2")
	}

	ADDPP(A0H, A1L, A1H, A2L, "2")
	ADDPP(A1H, A2L, A2H, A3L, "3")
	ADDPP(A2H, A3L, A3H, A4L, "4")
	ADDPP(A3H, A4L, A4H, A5L, "5")
	ADDPP(A4H, A5L, A5H, A6L, "6")
	ADDPP(A5H, A6L, A6H, A7L, "7")

	// vpsrlq		$32, ACC, PPL;
	// valignd		$2, ACC, ACC, ACC{%k1}{z};
	// vpaddq		PPL, ACC, ACC;
	// vpsrlq		$32, A6H, A6H; vpaddq	A6H, ACC, ACC;
	// vpsrlq		$32, A7L, A7L; vpaddq	A7L, ACC, ACC;
	// vpandq		LSW, A7H, PPL; vpaddq	PPL, ACC, ACC;
	// valignd		$16-8, ACC, ACC, %zmm0{%k2}
	// kshiftlw	$1, %k2, %k2

	f.VPSRLQ("$32", ACC, PPL)
	f.VALIGND_Z("$2", ACC, ACC, "K1", ACC)
	f.VPADDQ(PPL, ACC, ACC)
	f.VPSRLQ("$32", A6H, A6H)
	f.VPADDQ(A6H, ACC, ACC)
	f.VPSRLQ("$32", A7L, A7L)
	f.VPADDQ(A7L, ACC, ACC)
	f.VPANDQ(LSW, A7H, PPL)
	f.VPADDQ(PPL, ACC, ACC)
	f.VALIGND("$16-8", ACC, ACC, "K2", "Z0")
	f.KSHIFTLW("$1", "K2", "K2")

	// vpsrlq		$32, ACC, PPL;
	// valignd		$2, ACC, ACC, ACC{%k1}{z};
	// vpaddq		PPL, ACC, ACC;
	// vpsrlq		$32, A7H, A7H; vpaddq	A7H, ACC, ACC;
	// valignd		$16-9, ACC, ACC, %zmm0{%k2}
	// kshiftlw	$1, %k2, %k2

	f.VPSRLQ("$32", ACC, PPL)
	f.VALIGND_Z("$2", ACC, ACC, "K1", ACC)
	f.VPADDQ(PPL, ACC, ACC)
	f.VPSRLQ("$32", A7H, A7H)
	f.VPADDQ(A7H, ACC, ACC)
	f.VALIGND("$16-9", ACC, ACC, "K2", "Z0")
	f.KSHIFTLW("$1", "K2", "K2")

	// #define ADDPP(I) \
	// vpsrlq		$32, ACC, PPL; \
	// valignd		$2, ACC, ACC, ACC{%k1}{z}; \
	// vpaddq		PPL, ACC, ACC; \
	// valignd		$16-I, ACC, ACC, %zmm0{%k2}; \
	// kshiftlw	$1, %k2, %k2
	ADDPP_2 := func(I int) {
		f.VPSRLQ("$32", ACC, PPL)
		f.VALIGND_Z("$2", ACC, ACC, "K1", ACC)
		f.VPADDQ(PPL, ACC, ACC)
		f.VALIGND("$16-"+strconv.Itoa(I), ACC, ACC, "K2", "Z0")
		f.KSHIFTLW("$1", "K2", "K2")
	}

	ADDPP_2(10)
	ADDPP_2(11)
	ADDPP_2(12)
	ADDPP_2(13)
	ADDPP_2(14)
	ADDPP_2(15)

	// vpsrlq		$32, ACC, PPL;
	// valignd		$2, ACC, ACC, ACC{%k1}{z};
	// vpaddq		PPL, ACC, ACC;
	// vmovdqa64	ACC, %zmm1{%k1}{z}

	f.VPSRLQ("$32", ACC, PPL)
	f.VALIGND_Z("$2", ACC, ACC, "K1", ACC)
	f.VPADDQ(PPL, ACC, ACC)
	f.VMOVDQA64_Z(ACC, "K1", "Z1")

	// We have 544-bit (72-byte) result in Z1:Z0.
	// Only the modular reduction remains to be computed.

	T0 := f.Pop(&registers)
	T1 := f.Pop(&registers)
	T2 := f.Pop(&registers)
	T3 := f.Pop(&registers)
	T4 := f.Pop(&registers)

	// Extract the 4 least significant qwords of %zmm0

	// vmovq	%xmm0, T1; valignq	$1, %zmm0, %zmm1, %zmm0	// Shift in low word from zmm1
	// vmovq	%xmm0, T2; valignq	$1, %zmm0, %zmm0, %zmm0
	// vmovq	%xmm0, T3; valignq	$1, %zmm0, %zmm0, %zmm0
	// vmovq	%xmm0, T4; valignq	$1, %zmm0, %zmm0, %zmm0
	// xorq	T0, T0

	f.VMOVQ("X0", T1)
	f.VALIGNQ("$1", "Z0", "Z1", "Z0")
	f.VMOVQ("X0", T2)
	f.VALIGNQ("$1", "Z0", "Z0", "Z0")
	f.VMOVQ("X0", T3)
	f.VALIGNQ("$1", "Z0", "Z0", "Z0")
	f.VMOVQ("X0", T4)
	f.VALIGNQ("$1", "Z0", "Z0", "Z0")
	f.XORQ(T0, T0)

	// movq	INV, %rdx	// Load negative inverse mod 2^64

	// mulx	T1, %rdx, PH

	// mulx	0*8(PM), PL, PH; add	PL, T1; adc	PH, T2
	// mulx	2*8(PM), PL, PH; adc	PL, T3; adc	PH, T4; adc	$0, T0
	// mulx	1*8(PM), PL, PH; add	PL, T2; adc	PH, T3
	// mulx	3*8(PM), PL, PH; adc	PL, T4; adc	PH, T0; adc	$0, T1
	PH := f.Pop(&registers)
	PL := amd64.AX
	f.MOVQ(f.qInv0(), amd64.DX)
	f.MULXQ(T1, amd64.DX, PH)
	f.MULXQ(f.qAt(0), PL, PH)
	f.ADDQ(PL, T1)
	f.ADCQ(PH, T2)
	f.MULXQ(f.qAt(2), PL, PH)
	f.ADCQ(PL, T3)
	f.ADCQ(PH, T4)
	f.ADCQ("$0", T0)
	f.MULXQ(f.qAt(1), PL, PH)
	f.ADDQ(PL, T2)
	f.ADCQ(PH, T3)
	f.MULXQ(f.qAt(3), PL, PH)
	f.ADCQ(PL, T4)
	f.ADCQ(PH, T0)
	f.ADCQ("$0", T1)

	// movq	INV, %rdx

	// mulx	T2, %rdx, PH

	// mulx	0*8(PM), PL, PH; add	PL, T2; adc	PH, T3
	// mulx	2*8(PM), PL, PH; adc	PL, T4; adc	PH, T0; adc	$0, T1
	// mulx	1*8(PM), PL, PH; add	PL, T3; adc	PH, T4
	// mulx	3*8(PM), PL, PH; adc	PL, T0; adc	PH, T1; adc	$0, T2

	f.MOVQ(f.qInv0(), amd64.DX)
	f.MULXQ(T2, amd64.DX, PH)

	f.MULXQ(f.qAt(0), PL, PH)
	f.ADDQ(PL, T2)
	f.ADCQ(PH, T3)
	f.MULXQ(f.qAt(2), PL, PH)
	f.ADCQ(PL, T4)
	f.ADCQ(PH, T0)
	f.ADCQ("$0", T1)
	f.MULXQ(f.qAt(1), PL, PH)
	f.ADDQ(PL, T3)
	f.ADCQ(PH, T4)
	f.MULXQ(f.qAt(3), PL, PH)
	f.ADCQ(PL, T0)
	f.ADCQ(PH, T1)
	f.ADCQ("$0", T2)

	// movq	INV, %rdx

	// mulx	T3, %rdx, PH

	// mulx	0*8(PM), PL, PH; add	PL, T3; adc	PH, T4
	// mulx	2*8(PM), PL, PH; adc	PL, T0; adc	PH, T1; adc	$0, T2
	// mulx	1*8(PM), PL, PH; add	PL, T4; adc	PH, T0
	// mulx	3*8(PM), PL, PH; adc	PL, T1; adc	PH, T2; adc	$0, T3

	f.MOVQ(f.qInv0(), amd64.DX)

	f.MULXQ(T3, amd64.DX, PH)

	f.MULXQ(f.qAt(0), PL, PH)
	f.ADDQ(PL, T3)
	f.ADCQ(PH, T4)
	f.MULXQ(f.qAt(2), PL, PH)
	f.ADCQ(PL, T0)
	f.ADCQ(PH, T1)
	f.ADCQ("$0", T2)
	f.MULXQ(f.qAt(1), PL, PH)
	f.ADDQ(PL, T4)
	f.ADCQ(PH, T0)
	f.MULXQ(f.qAt(3), PL, PH)
	f.ADCQ(PL, T1)
	f.ADCQ(PH, T2)
	f.ADCQ("$0", T3)

	// movq	INV, %rdx

	// mulx	T4, %rdx, PH

	// mulx	0*8(PM), PL, PH; add	PL, T4; adc	PH, T0
	// mulx	2*8(PM), PL, PH; adc	PL, T1; adc	PH, T2; adc	$0, T3
	// mulx	1*8(PM), PL, PH; add	PL, T0; adc	PH, T1
	// mulx	3*8(PM), PL, PH; adc	PL, T2; adc	PH, T3; adc	$0, T4

	f.MOVQ(f.qInv0(), amd64.DX)

	f.MULXQ(T4, amd64.DX, PH)

	f.MULXQ(f.qAt(0), PL, PH)
	f.ADDQ(PL, T4)
	f.ADCQ(PH, T0)
	f.MULXQ(f.qAt(2), PL, PH)
	f.ADCQ(PL, T1)
	f.ADCQ(PH, T2)
	f.ADCQ("$0", T3)
	f.MULXQ(f.qAt(1), PL, PH)
	f.ADDQ(PL, T0)
	f.ADCQ(PH, T1)
	f.MULXQ(f.qAt(3), PL, PH)
	f.ADCQ(PL, T2)
	f.ADCQ(PH, T3)
	f.ADCQ("$0", T4)

	// Add the remaining 5 qwords (9 dwords) from zmm0

	// vmovq	%xmm0, PL; add	PL, T0;	valignq	$1, %zmm0, %zmm0, %zmm0
	// vmovq	%xmm0, PL; adc	PL, T1;	valignq	$1, %zmm0, %zmm0, %zmm0
	// vmovq	%xmm0, PL; adc	PL, T2;	valignq	$1, %zmm0, %zmm0, %zmm0
	// vmovq	%xmm0, PL; adc	PL, T3;	valignq	$1, %zmm0, %zmm0, %zmm0
	// vmovq	%xmm0, PL; adc	PL, T4	// T4 < 2^32

	f.VMOVQ("X0", PL)
	f.ADDQ(PL, T0)
	f.VALIGNQ("$1", "Z0", "Z0", "Z0")
	f.VMOVQ("X0", PL)
	f.ADCQ(PL, T1)
	f.VALIGNQ("$1", "Z0", "Z0", "Z0")
	f.VMOVQ("X0", PL)
	f.ADCQ(PL, T2)
	f.VALIGNQ("$1", "Z0", "Z0", "Z0")
	f.VMOVQ("X0", PL)
	f.ADCQ(PL, T3)
	f.VALIGNQ("$1", "Z0", "Z0", "Z0")
	f.VMOVQ("X0", PL)
	f.ADCQ(PL, T4)

	//////////////////////////////////////////////////
	// Barrett reduction
	//////////////////////////////////////////////////

	// // For explanation of mu, q1, q2, q3, r1, r2, see Handbook of
	// // Applied Cryptography, Algorithm 14.42.

	// // q1 is low 32 bits of T4 and high 32 bits of T3

	// movq	T3, %rax
	// shrd	$32, T4, %rax	// q1
	// mulq	MU		// Multiply by mu. q2 in rdx:rax, q3 in rdx

	f.MOVQ(T3, amd64.AX)
	f.SHRDw("$32", T4, amd64.AX)
	f.MULQ(f.mu())

	// // Subtract r2 from r1

	// mulx	0*8(PM), PL, PH; sub	PL, T0; sbb	PH, T1;
	// mulx	2*8(PM), PL, PH; sbb	PL, T2; sbb	PH, T3;	sbb	$0, T4
	// mulx	1*8(PM), PL, PH; sub	PL, T1; sbb	PH, T2;
	// mulx	3*8(PM), PL, PH; sbb	PL, T3; sbb	PH, T4

	f.MULXQ(f.qAt(0), PL, PH)
	f.SUBQ(PL, T0)
	f.SBBQ(PH, T1)
	f.MULXQ(f.qAt(2), PL, PH)
	f.SBBQ(PL, T2)
	f.SBBQ(PH, T3)
	f.SBBQ("$0", T4)
	f.MULXQ(f.qAt(1), PL, PH)
	f.SUBQ(PL, T1)
	f.SBBQ(PH, T2)
	f.MULXQ(f.qAt(3), PL, PH)
	f.SBBQ(PL, T3)
	f.SBBQ(PH, T4)

	PZ := f.Pop(&registers)
	f.MOVQ("res+0(FP)", PZ)
	t := []amd64.Register{T0, T1, T2, T3}
	f.Mov(t, PZ)

	// sub q
	f.SUBQ(f.qAt(0), T0)
	f.SBBQ(f.qAt(1), T1)
	f.SBBQ(f.qAt(2), T2)
	f.SBBQ(f.qAt(3), T3)
	f.SBBQ("$0", T4)

	// if borrow, we go to done
	f.JCS(done)

	f.Mov(t, PZ)

	f.SUBQ(f.qAt(0), T0)
	f.SBBQ(f.qAt(1), T1)
	f.SBBQ(f.qAt(2), T2)
	f.SBBQ(f.qAt(3), T3)
	f.SBBQ("$0", T4)

	f.JCS(done)

	f.Mov(t, PZ)

	f.LABEL(done)

	f.RET()
}
