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

	splitLoHi := f.Define("SPLIT_LO_HI", 2, func(args ...amd64.Register) {
		lo := args[0]
		hi := args[1]
		f.MOVQ(hi, lo)
		f.ANDQ("$0xffffffff", lo)
		f.SHLQ("$32", lo)
		f.SHRQ("$32", hi)
	})

	for _, v := range []hilo{{w0h, low0h}, {w1h, low1h}, {w2h, low2h}, {w3h, low3h}} {
		splitLoHi(v.lo, v.hi)
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

	// load arguments
	f.MOVQ("a+8(FP)", PX)
	f.MOVQ("b+16(FP)", PY)
	f.MOVQ("n+24(FP)", LEN)

	f.Comment("Create mask for low dword in each qword")
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

	f.Comment("we multiply and accumulate partial products of 4 bytes * 32 bytes")

	mac := f.Define("MAC", 3, func(inputs ...amd64.Register) {
		opLeft := inputs[0]
		lo := inputs[1]
		hi := inputs[2]

		f.VPMULUDQ_BCST(opLeft, Y, PPL)
		f.VPSRLQ("$32", PPL, PPH)
		f.VPANDQ(LSW, PPL, PPL)
		f.VPADDQ(PPL, lo, lo)
		f.VPADDQ(PPH, hi, hi)
	})

	mac("0*4("+PX+")", A0L, A0H)
	mac("1*4("+PX+")", A1L, A1H)
	mac("2*4("+PX+")", A2L, A2H)
	mac("3*4("+PX+")", A3L, A3H)
	mac("4*4("+PX+")", A4L, A4H)
	mac("5*4("+PX+")", A5L, A5H)
	mac("6*4("+PX+")", A6L, A6H)
	mac("7*4("+PX+")", A7L, A7H)

	f.ADDQ("$32", PX)

	f.DECQ(LEN, "decrement n")
	f.JMP(loop)

	f.Push(&registers, LEN, PX, PY)

	f.LABEL(AddPP)
	f.Comment("we accumulate the partial products into 544bits in Z1:Z0")

	f.MOVQ(uint64(0x1555), amd64.AX)
	f.KMOVD(amd64.AX, "K1")

	f.MOVQ(uint64(1), amd64.AX)
	f.KMOVD(amd64.AX, "K2")

	// ACC starts with the value of A0L

	f.Comment("store the least significant 32 bits of ACC (starts with A0L) in Z0")
	f.VALIGND_Z("$16", ACC, ACC, "K2", "Z0")
	f.KSHIFTLW("$1", "K2", "K2")

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

	f.Comment("macro to add partial products and store the result in Z0")
	addPP := f.Define("ADDPP", 5, func(inputs ...amd64.Register) {
		AxH := inputs[0]
		AyL := inputs[1]
		AyH := inputs[2]
		AzL := inputs[3]
		I := inputs[4]
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
	})

	addPP(A0H, A1L, A1H, A2L, "2")
	addPP(A1H, A2L, A2H, A3L, "3")
	addPP(A2H, A3L, A3H, A4L, "4")
	addPP(A3H, A4L, A4H, A5L, "5")
	addPP(A4H, A5L, A5H, A6L, "6")
	addPP(A5H, A6L, A6H, A7L, "7")

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

	f.VPSRLQ("$32", ACC, PPL)
	f.VALIGND_Z("$2", ACC, ACC, "K1", ACC)
	f.VPADDQ(PPL, ACC, ACC)
	f.VPSRLQ("$32", A7H, A7H)
	f.VPADDQ(A7H, ACC, ACC)
	f.VALIGND("$16-9", ACC, ACC, "K2", "Z0")
	f.KSHIFTLW("$1", "K2", "K2")

	addPP2 := f.Define("ADDPP2", 1, func(args ...amd64.Register) {
		f.VPSRLQ("$32", ACC, PPL)
		f.VALIGND_Z("$2", ACC, ACC, "K1", ACC)
		f.VPADDQ(PPL, ACC, ACC)
		f.VALIGND("$16-"+args[0], ACC, ACC, "K2", "Z0")
		f.KSHIFTLW("$1", "K2", "K2")
	})

	addPP2("10")
	addPP2("11")
	addPP2("12")
	addPP2("13")
	addPP2("14")
	addPP2("15")

	f.VPSRLQ("$32", ACC, PPL)
	f.VALIGND_Z("$2", ACC, ACC, "K1", ACC)
	f.VPADDQ(PPL, ACC, ACC)
	f.VMOVDQA64_Z(ACC, "K1", "Z1")

	T0 := f.Pop(&registers)
	T1 := f.Pop(&registers)
	T2 := f.Pop(&registers)
	T3 := f.Pop(&registers)
	T4 := f.Pop(&registers)

	f.Comment("Extract the 4 least significant qwords of Z0")
	f.VMOVQ("X0", T1)
	f.VALIGNQ("$1", "Z0", "Z1", "Z0")
	f.VMOVQ("X0", T2)
	f.VALIGNQ("$1", "Z0", "Z0", "Z0")
	f.VMOVQ("X0", T3)
	f.VALIGNQ("$1", "Z0", "Z0", "Z0")
	f.VMOVQ("X0", T4)
	f.VALIGNQ("$1", "Z0", "Z0", "Z0")
	f.XORQ(T0, T0)

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

	f.Comment("Barrett reduction; see Handbook of Applied Cryptography, Algorithm 14.42.")
	f.MOVQ(T3, amd64.AX)
	f.SHRQw("$32", T4, amd64.AX)
	f.MOVQ(f.mu(), amd64.DX)
	f.MULQ(amd64.DX)

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

	f.Comment("we need up to 2 conditional substractions to be < q")

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

func (f *FFAmd64) generateMulVec() {
	f.Comment("mulVec(res, a,b *Element, n uint64, qInvNeg uint64) res = a[0...n] * b[0...n]")

	const argSize = 5 * 8
	stackSize := f.StackSize(13, 1, 16)
	registers := f.FnHeader("mulVec", stackSize, argSize, amd64.DX)
	defer f.AssertCleanStack(stackSize, 0)

	// follows field/asm/modmul256.S
	// to simplify the generated assembly, we only handle n/16 (and do blocks of 16 muls).
	// that is if n%16 != 0, we let the caller (Go) handle the remaining elements.
	LEN := f.Pop(&registers)
	PZ := f.Pop(&registers, true)
	PX := amd64.BP
	PY := f.Pop(&registers, true)
	r := f.Pop(&registers)
	MUL := amd64.DX
	Y0 := f.Pop(&registers)
	Y1 := f.Pop(&registers)
	Y2 := f.Pop(&registers)
	Y3 := f.Pop(&registers)
	T0 := f.Pop(&registers)
	T1 := f.Pop(&registers)
	T2 := f.Pop(&registers)
	T3 := f.Pop(&registers)
	T4 := f.Pop(&registers)
	PL := f.Pop(&registers)
	PH := f.Pop(&registers)

	y := []amd64.Register{Y0, Y1, Y2, Y3}
	t := []amd64.Register{T0, T1, T2, T3}

	f.Comment("couple of defines")

	// 	INNER_MUL_0:
	// 	mulxq	Y0, T1, T2
	// 	mulxq	Y1, PL, T3;		addq	PL, T2
	// 	mulxq	Y2, PL, T4;		adcq	PL, T3
	// 	mulxq	Y3, PL, T0;		adcq	PL, T4;	adcq	$0, T0
	INNER_MUL_0 := f.Define("INNER_MUL_0", 0, func(args ...amd64.Register) {
		f.MULXQ(Y0, T1, T2)
		f.MULXQ(Y1, PL, T3)
		f.ADDQ(PL, T2)
		f.MULXQ(Y2, PL, T4)
		f.ADCQ(PL, T3)
		f.MULXQ(Y3, PL, T0)
		f.ADCQ(PL, T4)
		f.ADCQ("$0", T0)
	})

	// INNER_MUL_1(in0, in1, in2, in3):
	// 	mulxq	in0, PL, PH;	addq	PL, T1;	adcq	PH, T2
	// 	mulxq	in1, PL, PH;	adcq	PL, T3;	adcq	PH, T4;	adcq	$0, T0
	// 	mulxq	in2, PL, PH;	addq	PL, T2;	adcq	PH, T3
	// 	mulxq	in3, PL, PH;	adcq	PL, T4;	adcq	PH, T0;	adcq	$0, T1
	INNER_MUL_1 := f.Define("INNER_MUL_1", 4, func(args ...amd64.Register) {
		in0 := args[0]
		in1 := args[1]
		in2 := args[2]
		in3 := args[3]

		f.MULXQ(in0, PL, PH)
		f.ADDQ(PL, T1)
		f.ADCQ(PH, T2)
		f.MULXQ(in1, PL, PH)
		f.ADCQ(PL, T3)
		f.ADCQ(PH, T4)
		f.ADCQ("$0", T0)
		f.MULXQ(in2, PL, PH)
		f.ADDQ(PL, T2)
		f.ADCQ(PH, T3)
		f.MULXQ(in3, PL, PH)
		f.ADCQ(PL, T4)
		f.ADCQ(PH, T0)
		f.ADCQ("$0", T1)
	})

	// INNER_MUL_2:
	// 	mulxq	0*8(PM), PL, PH;	addq	PL, T2;	adcq	PH, T3
	// 	mulxq	2*8(PM), PL, PH;	adcq	PL, T4;	adcq	PH, T0;	adcq	$0, T1
	// 	mulxq	1*8(PM), PL, PH;	addq	PL, T3;	adcq	PH, T4
	// 	mulxq	3*8(PM), PL, PH;	adcq	PL, T0;	adcq	PH, T1;	adcq	$0, T2
	INNER_MUL_2 := f.Define("INNER_MUL_2", 0, func(args ...amd64.Register) {
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
	})

	// INNER_MUL_3(in0, in1, in2, in3):
	//     mulxq	in0, PL, PH;		addq	PL, T3;	adcq	PH, T4
	// 	mulxq	in1, PL, PH;		adcq	PL, T0;	adcq	PH, T1;	adcq	$0, T2
	// 	mulxq	in2, PL, PH;		addq	PL, T4;	adcq	PH, T0
	// 	mulxq	in3, PL, PH;		adcq	PL, T1;	adcq	PH, T2;	adcq	$0, T3
	INNER_MUL_3 := f.Define("INNER_MUL_3", 4, func(args ...amd64.Register) {
		in0 := args[0]
		in1 := args[1]
		in2 := args[2]
		in3 := args[3]

		f.MULXQ(in0, PL, PH)
		f.ADDQ(PL, T3)
		f.ADCQ(PH, T4)
		f.MULXQ(in1, PL, PH)
		f.ADCQ(PL, T0)
		f.ADCQ(PH, T1)
		f.ADCQ("$0", T2)
		f.MULXQ(in2, PL, PH)
		f.ADDQ(PL, T4)
		f.ADCQ(PH, T0)
		f.MULXQ(in3, PL, PH)
		f.ADCQ(PL, T1)
		f.ADCQ(PH, T2)
		f.ADCQ("$0", T3)
	})

	// INNER_MUL_4(in0, in1, in2, in3):
	//     mulxq	in0, PL, PH;		addq	PL, T4;	adcq	PH, T0
	// 	mulxq	in1, PL, PH;		adcq	PL, T1;	adcq	PH, T2;	adcq	$0, T3
	// 	mulxq	in2, PL, PH;		addq	PL, T0;	adcq	PH, T1
	// 	mulxq	in3, PL, PH;		adcq	PL, T2;	adcq	PH, T3;	adcq	$0, T4
	INNER_MUL_4 := f.Define("INNER_MUL_4", 4, func(args ...amd64.Register) {
		in0 := args[0]
		in1 := args[1]
		in2 := args[2]
		in3 := args[3]

		f.MULXQ(in0, PL, PH)
		f.ADDQ(PL, T4)
		f.ADCQ(PH, T0)
		f.MULXQ(in1, PL, PH)
		f.ADCQ(PL, T1)
		f.ADCQ(PH, T2)
		f.ADCQ("$0", T3)
		f.MULXQ(in2, PL, PH)
		f.ADDQ(PL, T0)
		f.ADCQ(PH, T1)
		f.MULXQ(in3, PL, PH)
		f.ADCQ(PL, T2)
		f.ADCQ(PH, T3)
		f.ADCQ("$0", T4)
	})

	// MUL_W_Q_LO:
	// 	vpmuludq	0*4(PM){1to8}, %zmm9, %zmm10;	vpaddq	%zmm10, %zmm0, %zmm0;	// Low dword of zmm0 is zero
	// 	vpmuludq	1*4(PM){1to8}, %zmm9, %zmm11;	vpaddq	%zmm11, %zmm1, %zmm1;
	// 	vpmuludq	2*4(PM){1to8}, %zmm9, %zmm12;	vpaddq	%zmm12, %zmm2, %zmm2;
	// 	vpmuludq	3*4(PM){1to8}, %zmm9, %zmm13;	vpaddq	%zmm13, %zmm3, %zmm3;
	MUL_W_Q_LO := f.Define("MUL_W_Q_LO", 0, func(args ...amd64.Register) {
		for i := 0; i < 4; i++ {
			f.VPMULUDQ_BCST(f.qAt_bcst(i), "Z9", "Z"+strconv.Itoa(10+i))
			f.VPADDQ("Z"+strconv.Itoa(10+i), "Z"+strconv.Itoa(i), "Z"+strconv.Itoa(i))
		}
	})

	// MUL_W_Q_HI:
	// 	vpmuludq	4*4(PM){1to8}, %zmm9, %zmm14;	vpaddq	%zmm14, %zmm4, %zmm4;
	// 	vpmuludq	5*4(PM){1to8}, %zmm9, %zmm15;	vpaddq	%zmm15, %zmm5, %zmm5;
	// 	vpmuludq	6*4(PM){1to8}, %zmm9, %zmm16;	vpaddq	%zmm16, %zmm6, %zmm6;
	// 	vpmuludq	7*4(PM){1to8}, %zmm9, %zmm17;	vpaddq	%zmm17, %zmm7, %zmm7;
	MUL_W_Q_HI := f.Define("MUL_W_Q_HI", 0, func(args ...amd64.Register) {
		for i := 0; i < 4; i++ {
			f.VPMULUDQ_BCST(f.qAt_bcst(i+4), "Z9", "Z"+strconv.Itoa(14+i))
			f.VPADDQ("Z"+strconv.Itoa(14+i), "Z"+strconv.Itoa(i+4), "Z"+strconv.Itoa(i+4))
		}
	})

	// CARRY1:
	// 	vpsrlq		$32, %zmm0, %zmm10;		vpaddq	%zmm10, %zmm1, %zmm1;	vpandq	%zmm8, %zmm1, %zmm0
	// 	vpsrlq		$32, %zmm1, %zmm11;		vpaddq	%zmm11, %zmm2, %zmm2;	vpandq	%zmm8, %zmm2, %zmm1
	// 	vpsrlq		$32, %zmm2, %zmm12;		vpaddq	%zmm12, %zmm3, %zmm3;	vpandq	%zmm8, %zmm3, %zmm2
	// 	vpsrlq		$32, %zmm3, %zmm13;		vpaddq	%zmm13, %zmm4, %zmm4;	vpandq	%zmm8, %zmm4, %zmm3
	CARRY1 := f.Define("CARRY1", 0, func(args ...amd64.Register) {
		for i := 0; i < 4; i++ {
			f.VPSRLQ("$32", "Z"+strconv.Itoa(i), "Z"+strconv.Itoa(10+i))
			f.VPADDQ("Z"+strconv.Itoa(10+i), "Z"+strconv.Itoa(i+1), "Z"+strconv.Itoa(i+1))
			f.VPANDQ("Z8", "Z"+strconv.Itoa(i+1), "Z"+strconv.Itoa(i))
		}
	})

	// CARRY2:
	// 	vpsrlq		$32, %zmm4, %zmm14;		vpaddq	%zmm14, %zmm5, %zmm5;	vpandq	%zmm8, %zmm5, %zmm4
	// 	vpsrlq		$32, %zmm5, %zmm15;		vpaddq	%zmm15, %zmm6, %zmm6;	vpandq	%zmm8, %zmm6, %zmm5
	// 	vpsrlq		$32, %zmm6, %zmm16;		vpaddq	%zmm16, %zmm7, %zmm7;	vpandq	%zmm8, %zmm7, %zmm6
	// 	vpsrlq		$32, %zmm7, %zmm7
	CARRY2 := f.Define("CARRY2", 0, func(args ...amd64.Register) {
		for i := 0; i < 3; i++ {
			f.VPSRLQ("$32", "Z"+strconv.Itoa(i+4), "Z"+strconv.Itoa(14+i))
			f.VPADDQ("Z"+strconv.Itoa(14+i), "Z"+strconv.Itoa(i+5), "Z"+strconv.Itoa(i+5))
			f.VPANDQ("Z8", "Z"+strconv.Itoa(i+5), "Z"+strconv.Itoa(i+4))
		}
		f.VPSRLQ("$32", "Z7", "Z7")
	})

	// CARRY3:
	// 	vpsrlq		$32, %zmm0, %zmm10;		vpandq	%zmm8, %zmm0, %zmm0;	vpaddq	%zmm10, %zmm1, %zmm1;
	// 	vpsrlq		$32, %zmm1, %zmm11;		vpandq	%zmm8, %zmm1, %zmm1;	vpaddq	%zmm11, %zmm2, %zmm2;
	// 	vpsrlq		$32, %zmm2, %zmm12;		vpandq	%zmm8, %zmm2, %zmm2;	vpaddq	%zmm12, %zmm3, %zmm3;
	// 	vpsrlq		$32, %zmm3, %zmm13;		vpandq	%zmm8, %zmm3, %zmm3;	vpaddq	%zmm13, %zmm4, %zmm4;
	CARRY3 := f.Define("CARRY3", 0, func(args ...amd64.Register) {
		for i := 0; i < 4; i++ {
			f.VPSRLQ("$32", "Z"+strconv.Itoa(i), "Z"+strconv.Itoa(10+i))
			f.VPANDQ("Z8", "Z"+strconv.Itoa(i), "Z"+strconv.Itoa(i))
			f.VPADDQ("Z"+strconv.Itoa(10+i), "Z"+strconv.Itoa(i+1), "Z"+strconv.Itoa(i+1))
		}
	})

	// CARRY4:
	// 	vpsrlq		$32, %zmm4, %zmm14;		vpandq	%zmm8, %zmm4, %zmm4;	vpaddq	%zmm14, %zmm5, %zmm5;
	// 	vpsrlq		$32, %zmm5, %zmm15;		vpandq	%zmm8, %zmm5, %zmm5;	vpaddq	%zmm15, %zmm6, %zmm6;
	// 	vpsrlq		$32, %zmm6, %zmm16;		vpandq	%zmm8, %zmm6, %zmm6;	vpaddq	%zmm16, %zmm7, %zmm7;
	CARRY4 := f.Define("CARRY4", 0, func(args ...amd64.Register) {
		for i := 0; i < 3; i++ {
			f.VPSRLQ("$32", "Z"+strconv.Itoa(i+4), "Z"+strconv.Itoa(14+i))
			f.VPANDQ("Z8", "Z"+strconv.Itoa(i+4), "Z"+strconv.Itoa(i+4))
			f.VPADDQ("Z"+strconv.Itoa(14+i), "Z"+strconv.Itoa(i+5), "Z"+strconv.Itoa(i+5))
		}
	})

	storeOutLoadIn := func() {
		f.MOVQ(PZ, r)
		f.Mov(t, r)
		f.ADDQ("$32", r)
		f.MOVQ(r, PZ)

		f.ADDQ("$32", PX)
		f.MOVQ("0*8("+PX+")", MUL)

		f.MOVQ(PY, r)
		f.ADDQ("$32", r)
		f.Mov(r, y)
		f.MOVQ(r, PY)
	}

	done := f.NewLabel("done")
	loop := f.NewLabel("loop")

	f.MOVQ("res+0(FP)", r)
	f.MOVQ(r, PZ)
	f.MOVQ("a+8(FP)", PX)
	f.MOVQ("b+16(FP)", r)
	f.MOVQ(r, PY)
	f.MOVQ("n+24(FP)", LEN)

	// shift LEN by 4 to get the number of blocks of 16 elements
	f.SHRQ("$4", LEN)
	// f.SHLQ("$5", LEN) // 32 bytes per element

	// Create mask for low dword in each qword
	// vpcmpeqb	%ymm8, %ymm8, %ymm8
	// vpmovzxdq	%ymm8, %zmm8
	// mov	$0x5555, %edx
	// kmovd	%edx, %k1

	f.VPCMPEQB("Y8", "Y8", "Y8")
	f.VPMOVZXDQ("Y8", "Z8")
	f.MOVQ(uint64(0x5555), amd64.DX)
	f.KMOVD(amd64.DX, "K1")

	f.LABEL(loop)
	f.TESTQ(LEN, LEN)
	f.JEQ(done, "n == 0, we are done")

	f.VMOVDQU64("256+0*64("+PX+")", "Z16")
	f.VMOVDQU64("256+1*64("+PX+")", "Z17")
	f.VMOVDQU64("256+2*64("+PX+")", "Z18")
	f.VMOVDQU64("256+3*64("+PX+")", "Z19")

	f.MOVQ("0*8("+PX+")", MUL)

	f.MOVQ(PY, r)
	f.VMOVDQU64("256+0*64("+r+")", "Z24")
	f.VMOVDQU64("256+1*64("+r+")", "Z25")
	f.VMOVDQU64("256+2*64("+r+")", "Z26")
	f.VMOVDQU64("256+3*64("+r+")", "Z27")

	f.MOVQ("0*8("+r+")", Y0)
	f.MOVQ("1*8("+r+")", Y1)
	f.MOVQ("2*8("+r+")", Y2)
	f.MOVQ("3*8("+r+")", Y3)

	//////////////////////////////////////////////////
	// Transpose and expand x and y
	//////////////////////////////////////////////////

	// Step 1

	f.VSHUFI64X2("$0x88", "Z17", "Z16", "Z20")
	f.VSHUFI64X2("$0xdd", "Z17", "Z16", "Z22")
	f.VSHUFI64X2("$0x88", "Z19", "Z18", "Z21")
	f.VSHUFI64X2("$0xdd", "Z19", "Z18", "Z23")

	f.VSHUFI64X2("$0x88", "Z25", "Z24", "Z28")
	f.VSHUFI64X2("$0xdd", "Z25", "Z24", "Z30")
	f.VSHUFI64X2("$0x88", "Z27", "Z26", "Z29")
	f.VSHUFI64X2("$0xdd", "Z27", "Z26", "Z31")

	// INNER_MUL_0
	// movq	T1, MUL
	INNER_MUL_0()
	f.MOVQ(T1, MUL)

	// Step 2

	f.VPERMQ("$0xd8", "Z20", "Z20")
	f.VPERMQ("$0xd8", "Z21", "Z21")
	f.VPERMQ("$0xd8", "Z22", "Z22")
	f.VPERMQ("$0xd8", "Z23", "Z23")

	f.MULXQ("qInvNeg+32(FP)", MUL, PH)
	INNER_MUL_1(amd64.Register(f.qAt(0)), amd64.Register(f.qAt(2)), amd64.Register(f.qAt(1)), amd64.Register(f.qAt(3)))

	f.MOVQ("1*8("+PX+")", MUL)

	f.VPERMQ("$0xd8", "Z28", "Z28")
	f.VPERMQ("$0xd8", "Z29", "Z29")
	f.VPERMQ("$0xd8", "Z30", "Z30")
	f.VPERMQ("$0xd8", "Z31", "Z31")

	INNER_MUL_1(Y0, Y2, Y1, Y3)
	f.MOVQ(T2, MUL)

	// Step 3

	for i := 20; i <= 23; i++ {
		f.VSHUFI64X2("$0xd8", "Z"+strconv.Itoa(i), "Z"+strconv.Itoa(i), "Z"+strconv.Itoa(i))
	}

	f.MULXQ("qInvNeg+32(FP)", MUL, PH)
	INNER_MUL_2()

	f.MOVQ("2*8("+PX+")", MUL)

	for i := 28; i <= 31; i++ {
		f.VSHUFI64X2("$0xd8", "Z"+strconv.Itoa(i), "Z"+strconv.Itoa(i), "Z"+strconv.Itoa(i))
	}

	INNER_MUL_3(Y0, Y2, Y1, Y3)
	f.MOVQ(T3, MUL)

	// Step 4

	// vshufi64x2	$0x44, %zmm21, %zmm20, %zmm16	// 0x44 = 0b01_00_01_00: low half of each input
	// vshufi64x2	$0xee, %zmm21, %zmm20, %zmm18	// 0xee = 0b11_10_11_10: high half of each input
	// vshufi64x2	$0x44, %zmm23, %zmm22, %zmm20
	// vshufi64x2	$0xee, %zmm23, %zmm22, %zmm22

	// mulxq	4*8(PM), MUL, PH
	// mulxq	0*8(PM), PL, PH;	addq	PL, T3;	adcq	PH, T4
	// mulxq	2*8(PM), PL, PH;	adcq	PL, T0;	adcq	PH, T1;	adcq	$0, T2
	// mulxq	1*8(PM), PL, PH;	addq	PL, T4;	adcq	PH, T0
	// mulxq	3*8(PM), PL, PH;	adcq	PL, T1;	adcq	PH, T2;	adcq	$0, T3

	// movq	3*8(PX, LEN), MUL

	// vshufi64x2	$0x44, %zmm29, %zmm28, %zmm24
	// vshufi64x2	$0xee, %zmm29, %zmm28, %zmm26
	// vshufi64x2	$0x44, %zmm31, %zmm30, %zmm28
	// vshufi64x2	$0xee, %zmm31, %zmm30, %zmm30

	f.VSHUFI64X2("$0x44", "Z21", "Z20", "Z16")
	f.VSHUFI64X2("$0xee", "Z21", "Z20", "Z18")
	f.VSHUFI64X2("$0x44", "Z23", "Z22", "Z20")
	f.VSHUFI64X2("$0xee", "Z23", "Z22", "Z22")

	f.MULXQ("qInvNeg+32(FP)", MUL, PH)
	INNER_MUL_3(amd64.Register(f.qAt(0)), amd64.Register(f.qAt(2)), amd64.Register(f.qAt(1)), amd64.Register(f.qAt(3)))

	f.MOVQ("3*8("+PX+")", MUL)

	f.VSHUFI64X2("$0x44", "Z29", "Z28", "Z24")
	f.VSHUFI64X2("$0xee", "Z29", "Z28", "Z26")
	f.VSHUFI64X2("$0x44", "Z31", "Z30", "Z28")
	f.VSHUFI64X2("$0xee", "Z31", "Z30", "Z30")

	// Step 5

	// vpsrlq		$32, %zmm16, %zmm17
	// vpsrlq		$32, %zmm18, %zmm19
	// vpsrlq		$32, %zmm20, %zmm21
	// vpsrlq		$32, %zmm22, %zmm23

	// mulxq	Y0, PL, PH;		addq	PL, T4;	adcq	PH, T0
	// mulxq	Y2, PL, PH;		adcq	PL, T1;	adcq	PH, T2;	adcq	$0, T3
	// mulxq	Y1, PL, PH;		addq	PL, T0;	adcq	PH, T1
	// mulxq	Y3, PL, PH;		adcq	PL, T2;	adcq	PH, T3;	adcq	$0, T4
	// movq	T4, MUL

	// vpsrlq		$32, %zmm24, %zmm25
	// vpsrlq		$32, %zmm26, %zmm27
	// vpsrlq		$32, %zmm28, %zmm29
	// vpsrlq		$32, %zmm30, %zmm31

	// mulxq	4*8(PM), MUL, PH
	// mulxq	0*8(PM), PL, PH;	addq	PL, T4;	adcq	PH, T0
	// mulxq	2*8(PM), PL, PH;	adcq	PL, T1;	adcq	PH, T2;	adcq	$0, T3
	// mulxq	1*8(PM), PL, PH;	addq	PL, T0;	adcq	PH, T1
	// mulxq	3*8(PM), PL, PH;	adcq	PL, T2;	adcq	PH, T3;	adcq	$0, T4

	// vpandq		%zmm8, %zmm16, %zmm16
	// vpandq		%zmm8, %zmm18, %zmm18
	// vpandq		%zmm8, %zmm20, %zmm20
	// vpandq		%zmm8, %zmm22, %zmm22

	// vpandq		%zmm8, %zmm24, %zmm24
	// vpandq		%zmm8, %zmm26, %zmm26
	// vpandq		%zmm8, %zmm28, %zmm28
	// vpandq		%zmm8, %zmm30, %zmm30

	f.VPSRLQ("$32", "Z16", "Z17")
	f.VPSRLQ("$32", "Z18", "Z19")
	f.VPSRLQ("$32", "Z20", "Z21")
	f.VPSRLQ("$32", "Z22", "Z23")

	INNER_MUL_4(Y0, Y2, Y1, Y3)
	f.MOVQ(T4, MUL)

	for i := 24; i <= 30; i += 2 {
		f.VPSRLQ("$32", "Z"+strconv.Itoa(i), "Z"+strconv.Itoa(i+1))
	}

	f.MULXQ("qInvNeg+32(FP)", MUL, PH)
	INNER_MUL_4(amd64.Register(f.qAt(0)), amd64.Register(f.qAt(2)), amd64.Register(f.qAt(1)), amd64.Register(f.qAt(3)))

	for i := 16; i <= 30; i += 2 {
		f.VPANDQ("Z8", "Z"+strconv.Itoa(i), "Z"+strconv.Itoa(i))
	}

	f.ReduceElement(t, y)

	// TODO @gbotrel add offset
	// we processed one element; offset by 32 PX, PY, PZ
	storeOutLoadIn()

	// For each 256-bit input value, each zmm register now represents a 32-bit input word zero-extended to 64 bits.

	//////////////////////////////////////////////////
	// Multiply y by doubleword 0 of x
	//////////////////////////////////////////////////

	// vpmuludq	%zmm16, %zmm24, %zmm0
	// vpmuludq	%zmm16, %zmm25, %zmm1
	// vpmuludq	%zmm16, %zmm26, %zmm2
	// vpmuludq	%zmm16, %zmm27, %zmm3
	// vpmuludq	%zmm16, %zmm28, %zmm4
	// vpmuludq	%zmm16, %zmm29, %zmm5
	// vpmuludq	%zmm16, %zmm30, %zmm6
	// vpmuludq	%zmm16, %zmm31, %zmm7

	// vpmuludq	8*4(PM){1to8}, %zmm0, %zmm9	// Reduction multiplier

	for i := 0; i < 8; i++ {
		f.VPMULUDQ("Z16", "Z"+strconv.Itoa(24+i), "Z"+strconv.Itoa(i))
	}
	f.VPMULUDQ_BCST("qInvNeg+32(FP)", "Z0", "Z9")

	// mulxq	Y0, T1, T2
	// mulxq	Y1, PL, T3;		addq	PL, T2
	// mulxq	Y2, PL, T4;		adcq	PL, T3
	// mulxq	Y3, PL, T0;		adcq	PL, T4;	adcq	$0, T0
	// movq	T1, MUL

	INNER_MUL_0()
	f.MOVQ(T1, MUL)

	// vpsrlq		$32, %zmm0, %zmm10;	vpandq	%zmm8, %zmm0, %zmm0;	vpaddq	%zmm10, %zmm1, %zmm1
	// vpsrlq		$32, %zmm1, %zmm11;	vpandq	%zmm8, %zmm1, %zmm1;	vpaddq	%zmm11, %zmm2, %zmm2
	// vpsrlq		$32, %zmm2, %zmm12;	vpandq	%zmm8, %zmm2, %zmm2;	vpaddq	%zmm12, %zmm3, %zmm3
	// vpsrlq		$32, %zmm3, %zmm13;	vpandq	%zmm8, %zmm3, %zmm3;	vpaddq	%zmm13, %zmm4, %zmm4

	for i := 0; i < 4; i++ {
		f.VPSRLQ("$32", "Z"+strconv.Itoa(i), "Z"+strconv.Itoa(10+i))
		f.VPANDQ("Z8", "Z"+strconv.Itoa(i), "Z"+strconv.Itoa(i))
		f.VPADDQ("Z"+strconv.Itoa(10+i), "Z"+strconv.Itoa(i+1), "Z"+strconv.Itoa(i+1))
	}

	// mulxq	4*8(PM), MUL, PH
	// mulxq	0*8(PM), PL, PH;	addq	PL, T1;	adcq	PH, T2
	// mulxq	2*8(PM), PL, PH;	adcq	PL, T3;	adcq	PH, T4;	adcq	$0, T0
	// mulxq	1*8(PM), PL, PH;	addq	PL, T2;	adcq	PH, T3
	// mulxq	3*8(PM), PL, PH;	adcq	PL, T4;	adcq	PH, T0;	adcq	$0, T1

	f.MULXQ("qInvNeg+32(FP)", MUL, PH)
	INNER_MUL_1(amd64.Register(f.qAt(0)), amd64.Register(f.qAt(2)), amd64.Register(f.qAt(1)), amd64.Register(f.qAt(3)))

	// movq	1*8(PX, LEN), MUL

	f.MOVQ("1*8("+PX+")", MUL)

	// vpsrlq		$32, %zmm4, %zmm14;	vpandq	%zmm8, %zmm4, %zmm4;	vpaddq	%zmm14, %zmm5, %zmm5
	// vpsrlq		$32, %zmm5, %zmm15;	vpandq	%zmm8, %zmm5, %zmm5;	vpaddq	%zmm15, %zmm6, %zmm6
	// vpsrlq		$32, %zmm6, %zmm16;	vpandq	%zmm8, %zmm6, %zmm6;	vpaddq	%zmm16, %zmm7, %zmm7

	for i := 0; i < 3; i++ {
		f.VPSRLQ("$32", "Z"+strconv.Itoa(4+i), "Z"+strconv.Itoa(14+i))
		f.VPANDQ("Z8", "Z"+strconv.Itoa(4+i), "Z"+strconv.Itoa(4+i))
		f.VPADDQ("Z"+strconv.Itoa(14+i), "Z"+strconv.Itoa(5+i), "Z"+strconv.Itoa(5+i))
	}

	// mulxq	Y0, PL, PH;		addq	PL, T2;	adcq	PH, T3
	// mulxq	Y2, PL, PH;		adcq	PL, T4;	adcq	PH, T0;	adcq	$0, T1
	// mulxq	Y1, PL, PH;		addq	PL, T3;	adcq	PH, T4
	// mulxq	Y3, PL, PH;		adcq	PL, T0;	adcq	PH, T1;	adcq	$0, T2
	// movq	T2, MUL

	INNER_MUL_1(Y0, Y2, Y1, Y3)
	f.MOVQ(T2, MUL)

	//////////////////////////////////////////////////
	// Reduce
	//////////////////////////////////////////////////
	f.Comment("Reduce")

	// vpmuludq	0*4(PM){1to8}, %zmm9, %zmm10;	vpaddq	%zmm10, %zmm0, %zmm0
	// vpmuludq	1*4(PM){1to8}, %zmm9, %zmm11;	vpaddq	%zmm11, %zmm1, %zmm1
	// vpmuludq	2*4(PM){1to8}, %zmm9, %zmm12;	vpaddq	%zmm12, %zmm2, %zmm2
	// vpmuludq	3*4(PM){1to8}, %zmm9, %zmm13;	vpaddq	%zmm13, %zmm3, %zmm3

	for i := 0; i < 4; i++ {
		f.VPMULUDQ_BCST(f.qAt_bcst(i), "Z9", "Z"+strconv.Itoa(10+i))
		f.VPADDQ("Z"+strconv.Itoa(10+i), "Z"+strconv.Itoa(i), "Z"+strconv.Itoa(i))
	}

	// mulxq	f.qInv0(), MUL, PH
	// INNER_MUL_2
	// movq	2*8(PX, LEN), MUL
	f.MULXQ("qInvNeg+32(FP)", MUL, PH)
	INNER_MUL_2()

	f.MOVQ("2*8("+PX+")", MUL)

	// vpmuludq	4*4(PM){1to8}, %zmm9, %zmm14;	vpaddq	%zmm14, %zmm4, %zmm4
	// vpmuludq	5*4(PM){1to8}, %zmm9, %zmm15;	vpaddq	%zmm15, %zmm5, %zmm5
	// vpmuludq	6*4(PM){1to8}, %zmm9, %zmm16;	vpaddq	%zmm16, %zmm6, %zmm6
	// vpmuludq	7*4(PM){1to8}, %zmm9, %zmm10;	vpaddq	%zmm10, %zmm7, %zmm7

	f.VPMULUDQ_BCST(f.qAt_bcst(4), "Z9", "Z14")
	f.VPADDQ("Z14", "Z4", "Z4")

	f.VPMULUDQ_BCST(f.qAt_bcst(5), "Z9", "Z15")
	f.VPADDQ("Z15", "Z5", "Z5")

	f.VPMULUDQ_BCST(f.qAt_bcst(6), "Z9", "Z16")
	f.VPADDQ("Z16", "Z6", "Z6")

	f.VPMULUDQ_BCST(f.qAt_bcst(7), "Z9", "Z10")
	f.VPADDQ("Z10", "Z7", "Z7")

	// INNER_MUL_3(Y0, Y2, Y1, Y3)
	// movq	T3, MUL
	INNER_MUL_3(Y0, Y2, Y1, Y3)
	f.MOVQ(T3, MUL)

	// vpsrlq	$32, %zmm0, %zmm10;	vpaddq	%zmm10, %zmm1, %zmm1;	vpandq	%zmm8, %zmm1, %zmm0
	// vpsrlq	$32, %zmm1, %zmm11;	vpaddq	%zmm11, %zmm2, %zmm2;	vpandq	%zmm8, %zmm2, %zmm1
	// vpsrlq	$32, %zmm2, %zmm12;	vpaddq	%zmm12, %zmm3, %zmm3;	vpandq	%zmm8, %zmm3, %zmm2
	// vpsrlq	$32, %zmm3, %zmm13;	vpaddq	%zmm13, %zmm4, %zmm4;	vpandq	%zmm8, %zmm4, %zmm3

	for i := 0; i < 4; i++ {
		f.VPSRLQ("$32", "Z"+strconv.Itoa(i), "Z"+strconv.Itoa(10+i))
		f.VPADDQ("Z"+strconv.Itoa(10+i), "Z"+strconv.Itoa(i+1), "Z"+strconv.Itoa(i+1))
		f.VPANDQ("Z8", "Z"+strconv.Itoa(i+1), "Z"+strconv.Itoa(i))
	}

	// mulxq	f.qInv0(), MUL, PH
	// INNER_MUL_3(f.qAt(0), f.qAt(2), f.qAt(1), f.qAt(3))
	// movq	3*8(PX, LEN), MUL

	f.MULXQ("qInvNeg+32(FP)", MUL, PH)
	INNER_MUL_3(amd64.Register(f.qAt(0)), amd64.Register(f.qAt(2)), amd64.Register(f.qAt(1)), amd64.Register(f.qAt(3)))

	f.MOVQ("3*8("+PX+")", MUL)

	// vpsrlq	$32, %zmm4, %zmm14;	vpaddq	%zmm14, %zmm5, %zmm5;	vpandq	%zmm8, %zmm5, %zmm4
	// vpsrlq	$32, %zmm5, %zmm15;	vpaddq	%zmm15, %zmm6, %zmm6;	vpandq	%zmm8, %zmm6, %zmm5
	// vpsrlq	$32, %zmm6, %zmm16;	vpaddq	%zmm16, %zmm7, %zmm7;	vpandq	%zmm8, %zmm7, %zmm6
	// vpsrlq	$32, %zmm7, %zmm7

	for i := 0; i < 3; i++ {
		f.VPSRLQ("$32", "Z"+strconv.Itoa(4+i), "Z"+strconv.Itoa(14+i))
		f.VPADDQ("Z"+strconv.Itoa(14+i), "Z"+strconv.Itoa(4+i+1), "Z"+strconv.Itoa(4+i+1))
		f.VPANDQ("Z8", "Z"+strconv.Itoa(4+i+1), "Z"+strconv.Itoa(4+i))
	}
	f.VPSRLQ("$32", "Z7", "Z7")

	// INNER_MUL_4(Y0, Y2, Y1, Y3)
	// movq	T4, MUL
	INNER_MUL_4(Y0, Y2, Y1, Y3)
	f.MOVQ(T4, MUL)

	//////////////////////////////////////////////////
	// Process doubleword 1 of x
	//////////////////////////////////////////////////

	// vpmuludq	%zmm17, %zmm24, %zmm10;		vpaddq	%zmm10, %zmm0, %zmm0;
	// vpmuludq	%zmm17, %zmm25, %zmm11;		vpaddq	%zmm11, %zmm1, %zmm1;
	// vpmuludq	%zmm17, %zmm26, %zmm12;		vpaddq	%zmm12, %zmm2, %zmm2;
	// vpmuludq	%zmm17, %zmm27, %zmm13;		vpaddq	%zmm13, %zmm3, %zmm3;

	for i := 0; i < 4; i++ {
		f.VPMULUDQ("Z17", "Z"+strconv.Itoa(24+i), "Z"+strconv.Itoa(10+i))
		f.VPADDQ("Z"+strconv.Itoa(10+i), "Z"+strconv.Itoa(i), "Z"+strconv.Itoa(i))
	}

	// mulxq	f.qInv0(), MUL, PH
	// INNER_MUL_4(f.qAt(0), f.qAt(2), f.qAt(1), f.qAt(3))

	f.MULXQ("qInvNeg+32(FP)", MUL, PH)
	INNER_MUL_4(amd64.Register(f.qAt(0)), amd64.Register(f.qAt(2)), amd64.Register(f.qAt(1)), amd64.Register(f.qAt(3)))

	// vpmuludq	%zmm17, %zmm28, %zmm14;		vpaddq	%zmm14, %zmm4, %zmm4;
	// vpmuludq	%zmm17, %zmm29, %zmm15;		vpaddq	%zmm15, %zmm5, %zmm5;
	// vpmuludq	%zmm17, %zmm30, %zmm16;		vpaddq	%zmm16, %zmm6, %zmm6;
	// vpmuludq	%zmm17, %zmm31, %zmm17;		vpaddq	%zmm17, %zmm7, %zmm7;

	// vpmuludq	f.qInv0(){1to8}, %zmm0, %zmm9;

	for i := 0; i < 4; i++ {
		f.VPMULUDQ("Z17", "Z"+strconv.Itoa(28+i), "Z"+strconv.Itoa(14+i))
		f.VPADDQ("Z"+strconv.Itoa(14+i), "Z"+strconv.Itoa(4+i), "Z"+strconv.Itoa(4+i))
	}
	f.VPMULUDQ_BCST("qInvNeg+32(FP)", "Z0", "Z9")

	f.ReduceElement(t, y)

	storeOutLoadIn()

	// Move high dwords to zmm10-16, add each to the corresponding low dword (propagate 32-bit carries)

	// vpsrlq		$32, %zmm0, %zmm10;		vpandq	%zmm8, %zmm0, %zmm0;	vpaddq	%zmm10, %zmm1, %zmm1;
	// vpsrlq		$32, %zmm1, %zmm11;		vpandq	%zmm8, %zmm1, %zmm1;	vpaddq	%zmm11, %zmm2, %zmm2;
	// vpsrlq		$32, %zmm2, %zmm12;		vpandq	%zmm8, %zmm2, %zmm2;	vpaddq	%zmm12, %zmm3, %zmm3;

	for i := 0; i < 3; i++ {
		f.VPSRLQ("$32", "Z"+strconv.Itoa(i), "Z"+strconv.Itoa(10+i))
		f.VPANDQ("Z8", "Z"+strconv.Itoa(i), "Z"+strconv.Itoa(i))
		f.VPADDQ("Z"+strconv.Itoa(10+i), "Z"+strconv.Itoa(i+1), "Z"+strconv.Itoa(i+1))
	}

	// INNER_MUL_0
	// movq	T1, MUL
	INNER_MUL_0()
	f.MOVQ(T1, MUL)

	// vpsrlq		$32, %zmm3, %zmm13;		vpandq	%zmm8, %zmm3, %zmm3;	vpaddq	%zmm13, %zmm4, %zmm4;
	// CARRY4
	f.VPSRLQ("$32", "Z3", "Z13")
	f.VPANDQ("Z8", "Z3", "Z3")
	f.VPADDQ("Z13", "Z4", "Z4")
	CARRY4()

	// zmm7 keeps all 64 bits

	// mulxq	f.qInv0(), MUL, PH
	// INNER_MUL_1(f.qAt(0), f.qAt(2), f.qAt(1), f.qAt(3))
	// movq	1*8(PX, LEN), MUL
	// MUL_W_Q_LO
	// INNER_MUL_1(Y0, Y2, Y1, Y3)
	// movq	T2, MUL
	// MUL_W_Q_HI
	// mulxq	f.qInv0(), MUL, PH
	// INNER_MUL_2
	// movq	2*8(PX, LEN), MUL

	f.MULXQ("qInvNeg+32(FP)", MUL, PH)
	INNER_MUL_1(amd64.Register(f.qAt(0)), amd64.Register(f.qAt(2)), amd64.Register(f.qAt(1)), amd64.Register(f.qAt(3)))

	f.MOVQ("1*8("+PX+")", MUL)
	MUL_W_Q_LO()
	INNER_MUL_1(Y0, Y2, Y1, Y3)
	f.MOVQ(T2, MUL)
	MUL_W_Q_HI()
	f.MULXQ("qInvNeg+32(FP)", MUL, PH)
	INNER_MUL_2()

	f.MOVQ("2*8("+PX+")", MUL)

	// Propagate carries and shift down by one dword
	// CARRY1
	// INNER_MUL_3(Y0, Y2, Y1, Y3)
	// movq	T3, MUL
	// CARRY2
	// mulxq	f.qInv0(), MUL, PH
	// INNER_MUL_3(f.qAt(0), f.qAt(2), f.qAt(1), f.qAt(3))
	// movq	3*8(PX, LEN), MUL
	CARRY1()
	INNER_MUL_3(Y0, Y2, Y1, Y3)
	f.MOVQ(T3, MUL)
	CARRY2()
	f.MULXQ("qInvNeg+32(FP)", MUL, PH)
	INNER_MUL_3(amd64.Register(f.qAt(0)), amd64.Register(f.qAt(2)), amd64.Register(f.qAt(1)), amd64.Register(f.qAt(3)))

	f.MOVQ("3*8("+PX+")", MUL)

	//////////////////////////////////////////////////
	// Process doubleword 2 of x
	//////////////////////////////////////////////////

	// vpmuludq	%zmm18, %zmm24, %zmm10;		vpaddq	%zmm10, %zmm0, %zmm0;
	// vpmuludq	%zmm18, %zmm25, %zmm11;		vpaddq	%zmm11, %zmm1, %zmm1;
	// vpmuludq	%zmm18, %zmm26, %zmm12;		vpaddq	%zmm12, %zmm2, %zmm2;
	// vpmuludq	%zmm18, %zmm27, %zmm13;		vpaddq	%zmm13, %zmm3, %zmm3;

	for i := 0; i < 4; i++ {
		f.VPMULUDQ("Z18", "Z"+strconv.Itoa(24+i), "Z"+strconv.Itoa(10+i))
		f.VPADDQ("Z"+strconv.Itoa(10+i), "Z"+strconv.Itoa(i), "Z"+strconv.Itoa(i))
	}

	// INNER_MUL_4(Y0, Y2, Y1, Y3)
	// movq	T4, MUL
	INNER_MUL_4(Y0, Y2, Y1, Y3)
	f.MOVQ(T4, MUL)

	// vpmuludq	%zmm18, %zmm28, %zmm14;		vpaddq	%zmm14, %zmm4, %zmm4;
	// vpmuludq	%zmm18, %zmm29, %zmm15;		vpaddq	%zmm15, %zmm5, %zmm5;
	// vpmuludq	%zmm18, %zmm30, %zmm16;		vpaddq	%zmm16, %zmm6, %zmm6;
	// vpmuludq	%zmm18, %zmm31, %zmm17;		vpaddq	%zmm17, %zmm7, %zmm7;

	// vpmuludq	f.qInv0(){1to8}, %zmm0, %zmm9;	// Compute reduction multipliers

	// mulxq	f.qInv0(), MUL, PH
	// INNER_MUL_4(f.qAt(0), f.qAt(2), f.qAt(1), f.qAt(3))

	for i := 0; i < 4; i++ {
		f.VPMULUDQ("Z18", "Z"+strconv.Itoa(28+i), "Z"+strconv.Itoa(14+i))
		f.VPADDQ("Z"+strconv.Itoa(14+i), "Z"+strconv.Itoa(4+i), "Z"+strconv.Itoa(4+i))
	}

	f.VPMULUDQ_BCST("qInvNeg+32(FP)", "Z0", "Z9")
	f.MULXQ("qInvNeg+32(FP)", MUL, PH)
	INNER_MUL_4(amd64.Register(f.qAt(0)), amd64.Register(f.qAt(2)), amd64.Register(f.qAt(1)), amd64.Register(f.qAt(3)))

	// Move high dwords to zmm10-16, add each to the corresponding low dword (propagate 32-bit carries)
	CARRY3()
	f.ReduceElement(t, y)

	CARRY4()

	storeOutLoadIn()

	// MUL_W_Q_LO
	// INNER_MUL_0
	// movq	T1, MUL
	// MUL_W_Q_HI
	// mulxq	f.qInv0(), MUL, PH
	// INNER_MUL_1(f.qAt(0), f.qAt(2), f.qAt(1), f.qAt(3))
	// movq	1*8(PX, LEN), MUL

	MUL_W_Q_LO()
	INNER_MUL_0()
	f.MOVQ(T1, MUL)
	MUL_W_Q_HI()
	f.MULXQ("qInvNeg+32(FP)", MUL, PH)
	INNER_MUL_1(amd64.Register(f.qAt(0)), amd64.Register(f.qAt(2)), amd64.Register(f.qAt(1)), amd64.Register(f.qAt(3)))

	f.MOVQ("1*8("+PX+")", MUL)

	// Propagate carries and shift down by one dword
	// CARRY1
	// INNER_MUL_1(Y0, Y2, Y1, Y3)
	// movq	T2, MUL
	// CARRY2
	// mulxq	f.qInv0(), MUL, PH
	// INNER_MUL_2
	// movq	2*8(PX, LEN), MUL
	CARRY1()
	INNER_MUL_1(Y0, Y2, Y1, Y3)
	f.MOVQ(T2, MUL)
	CARRY2()
	f.MULXQ("qInvNeg+32(FP)", MUL, PH)
	INNER_MUL_2()

	f.MOVQ("2*8("+PX+")", MUL)

	//////////////////////////////////////////////////
	// Process doubleword 3 of x
	//////////////////////////////////////////////////

	// vpmuludq	%zmm19, %zmm24, %zmm10;		vpaddq	%zmm10, %zmm0, %zmm0;
	// vpmuludq	%zmm19, %zmm25, %zmm11;		vpaddq	%zmm11, %zmm1, %zmm1;
	// vpmuludq	%zmm19, %zmm26, %zmm12;		vpaddq	%zmm12, %zmm2, %zmm2;
	// vpmuludq	%zmm19, %zmm27, %zmm13;		vpaddq	%zmm13, %zmm3, %zmm3;

	// INNER_MUL_3(Y0, Y2, Y1, Y3)
	// movq	T3, MUL

	// vpmuludq	%zmm19, %zmm28, %zmm14;		vpaddq	%zmm14, %zmm4, %zmm4;
	// vpmuludq	%zmm19, %zmm29, %zmm15;		vpaddq	%zmm15, %zmm5, %zmm5;
	// vpmuludq	%zmm19, %zmm30, %zmm16;		vpaddq	%zmm16, %zmm6, %zmm6;
	// vpmuludq	%zmm19, %zmm31, %zmm17;		vpaddq	%zmm17, %zmm7, %zmm7;

	// vpmuludq	f.qInv0(){1to8}, %zmm0, %zmm9;	// Compute reduction multipliers

	// mulxq	f.qInv0(), MUL, PH
	// INNER_MUL_3(f.qAt(0), f.qAt(2), f.qAt(1), f.qAt(3))

	// movq	3*8(PX, LEN), MUL

	for i := 0; i < 4; i++ {
		f.VPMULUDQ("Z19", "Z"+strconv.Itoa(24+i), "Z"+strconv.Itoa(10+i))
		f.VPADDQ("Z"+strconv.Itoa(10+i), "Z"+strconv.Itoa(i), "Z"+strconv.Itoa(i))
	}

	INNER_MUL_3(Y0, Y2, Y1, Y3)
	f.MOVQ(T3, MUL)

	for i := 0; i < 4; i++ {
		f.VPMULUDQ("Z19", "Z"+strconv.Itoa(28+i), "Z"+strconv.Itoa(14+i))
		f.VPADDQ("Z"+strconv.Itoa(14+i), "Z"+strconv.Itoa(4+i), "Z"+strconv.Itoa(4+i))
	}

	f.VPMULUDQ_BCST("qInvNeg+32(FP)", "Z0", "Z9")
	f.MULXQ("qInvNeg+32(FP)", MUL, PH)

	INNER_MUL_3(amd64.Register(f.qAt(0)), amd64.Register(f.qAt(2)), amd64.Register(f.qAt(1)), amd64.Register(f.qAt(3)))

	f.MOVQ("3*8("+PX+")", MUL)

	// Move high dwords to zmm10-16, add each to the corresponding low dword (propagate 32-bit carries)
	// CARRY3
	// INNER_MUL_4(Y0, Y2, Y1, Y3)
	// movq	T4, MUL
	// CARRY4
	CARRY3()
	INNER_MUL_4(Y0, Y2, Y1, Y3)
	f.MOVQ(T4, MUL)
	CARRY4()

	MUL_W_Q_LO()

	// mulxq	f.qInv0(), MUL, PH
	// INNER_MUL_4(f.qAt(0), f.qAt(2), f.qAt(1), f.qAt(3))
	f.MULXQ("qInvNeg+32(FP)", MUL, PH)
	INNER_MUL_4(amd64.Register(f.qAt(0)), amd64.Register(f.qAt(2)), amd64.Register(f.qAt(1)), amd64.Register(f.qAt(3)))

	f.ReduceElement(t, y)

	MUL_W_Q_HI()

	storeOutLoadIn()

	// // Propagate carries and shift down by one dword

	// CARRY1
	// INNER_MUL_0
	// movq	T1, MUL
	// CARRY2
	// mulxq	f.qInv0(), MUL, PH
	// INNER_MUL_1(f.qAt(0), f.qAt(2), f.qAt(1), f.qAt(3))
	// movq	1*8(PX, LEN), MUL
	CARRY1()
	INNER_MUL_0()
	f.MOVQ(T1, MUL)
	CARRY2()
	f.MULXQ("qInvNeg+32(FP)", MUL, PH)
	INNER_MUL_1(amd64.Register(f.qAt(0)), amd64.Register(f.qAt(2)), amd64.Register(f.qAt(1)), amd64.Register(f.qAt(3)))

	f.MOVQ("1*8("+PX+")", MUL)

	//////////////////////////////////////////////////
	// Process doubleword 4 of x
	//////////////////////////////////////////////////

	// vpmuludq	%zmm20, %zmm24, %zmm10;		vpaddq	%zmm10, %zmm0, %zmm0;
	// vpmuludq	%zmm20, %zmm25, %zmm11;		vpaddq	%zmm11, %zmm1, %zmm1;
	// vpmuludq	%zmm20, %zmm26, %zmm12;		vpaddq	%zmm12, %zmm2, %zmm2;
	// vpmuludq	%zmm20, %zmm27, %zmm13;		vpaddq	%zmm13, %zmm3, %zmm3;

	// INNER_MUL_1(Y0, Y2, Y1, Y3)
	// movq	T2, MUL

	// vpmuludq	%zmm20, %zmm28, %zmm14;		vpaddq	%zmm14, %zmm4, %zmm4;
	// vpmuludq	%zmm20, %zmm29, %zmm15;		vpaddq	%zmm15, %zmm5, %zmm5;
	// vpmuludq	%zmm20, %zmm30, %zmm16;		vpaddq	%zmm16, %zmm6, %zmm6;
	// vpmuludq	%zmm20, %zmm31, %zmm17;		vpaddq	%zmm17, %zmm7, %zmm7;

	// vpmuludq	f.qInv0(){1to8}, %zmm0, %zmm9;	// Compute reduction multipliers

	// mulxq	f.qInv0(), MUL, PH
	// INNER_MUL_2

	// movq	2*8(PX, LEN), MUL

	for i := 0; i < 4; i++ {
		f.VPMULUDQ("Z20", "Z"+strconv.Itoa(24+i), "Z"+strconv.Itoa(10+i))
		f.VPADDQ("Z"+strconv.Itoa(10+i), "Z"+strconv.Itoa(i), "Z"+strconv.Itoa(i))
	}

	INNER_MUL_1(Y0, Y2, Y1, Y3)
	f.MOVQ(T2, MUL)

	for i := 0; i < 4; i++ {
		f.VPMULUDQ("Z20", "Z"+strconv.Itoa(28+i), "Z"+strconv.Itoa(14+i))
		f.VPADDQ("Z"+strconv.Itoa(14+i), "Z"+strconv.Itoa(4+i), "Z"+strconv.Itoa(4+i))
	}

	f.VPMULUDQ_BCST("qInvNeg+32(FP)", "Z0", "Z9")
	f.MULXQ("qInvNeg+32(FP)", MUL, PH)
	INNER_MUL_2()

	f.MOVQ("2*8("+PX+")", MUL)

	// // Move high dwords to zmm10-16, add each to the corresponding low dword (propagate 32-bit carries)

	// CARRY3

	// INNER_MUL_3(Y0, Y2, Y1, Y3)
	// movq	T3, MUL
	// mulxq	f.qInv0(), MUL, PH

	// CARRY4

	CARRY3()
	INNER_MUL_3(Y0, Y2, Y1, Y3)
	f.MOVQ(T3, MUL)
	f.MULXQ("qInvNeg+32(FP)", MUL, PH)
	CARRY4()

	// zmm7 keeps all 64 bits

	// INNER_MUL_3(f.qAt(0), f.qAt(2), f.qAt(1), f.qAt(3))

	// movq	3*8(PX, LEN), MUL

	// MUL_W_Q_LO

	// INNER_MUL_4(Y0, Y2, Y1, Y3)
	// movq	T4, MUL
	// mulxq	f.qInv0(), MUL, PH

	// MUL_W_Q_HI

	// INNER_MUL_4(f.qAt(0), f.qAt(2), f.qAt(1), f.qAt(3))

	INNER_MUL_3(amd64.Register(f.qAt(0)), amd64.Register(f.qAt(2)), amd64.Register(f.qAt(1)), amd64.Register(f.qAt(3)))

	f.MOVQ("3*8("+PX+")", MUL)
	MUL_W_Q_LO()
	INNER_MUL_4(Y0, Y2, Y1, Y3)
	f.MOVQ(T4, MUL)
	f.MULXQ("qInvNeg+32(FP)", MUL, PH)

	MUL_W_Q_HI()
	INNER_MUL_4(amd64.Register(f.qAt(0)), amd64.Register(f.qAt(2)), amd64.Register(f.qAt(1)), amd64.Register(f.qAt(3)))

	// Propagate carries and shift down by one dword

	CARRY1()
	CARRY2()
	f.ReduceElement(t, y)

	storeOutLoadIn()

	// 	//////////////////////////////////////////////////
	// // Process doubleword 5 of x
	// //////////////////////////////////////////////////

	// vpmuludq	%zmm21, %zmm24, %zmm10;		vpaddq	%zmm10, %zmm0, %zmm0;
	// vpmuludq	%zmm21, %zmm25, %zmm11;		vpaddq	%zmm11, %zmm1, %zmm1;
	// vpmuludq	%zmm21, %zmm26, %zmm12;		vpaddq	%zmm12, %zmm2, %zmm2;
	// vpmuludq	%zmm21, %zmm27, %zmm13;		vpaddq	%zmm13, %zmm3, %zmm3;

	for i := 0; i < 4; i++ {
		f.VPMULUDQ("Z21", "Z"+strconv.Itoa(24+i), "Z"+strconv.Itoa(10+i))
		f.VPADDQ("Z"+strconv.Itoa(10+i), "Z"+strconv.Itoa(i), "Z"+strconv.Itoa(i))
	}

	// INNER_MUL_0
	// movq	T1, MUL
	// mulxq	f.qInv0(), MUL, PH

	INNER_MUL_0()
	f.MOVQ(T1, MUL)
	f.MULXQ("qInvNeg+32(FP)", MUL, PH)

	// vpmuludq	%zmm21, %zmm28, %zmm14;		vpaddq	%zmm14, %zmm4, %zmm4;
	// vpmuludq	%zmm21, %zmm29, %zmm15;		vpaddq	%zmm15, %zmm5, %zmm5;
	// vpmuludq	%zmm21, %zmm30, %zmm16;		vpaddq	%zmm16, %zmm6, %zmm6;
	// vpmuludq	%zmm21, %zmm31, %zmm17;		vpaddq	%zmm17, %zmm7, %zmm7;

	for i := 0; i < 4; i++ {
		f.VPMULUDQ("Z21", "Z"+strconv.Itoa(28+i), "Z"+strconv.Itoa(14+i))
		f.VPADDQ("Z"+strconv.Itoa(14+i), "Z"+strconv.Itoa(4+i), "Z"+strconv.Itoa(4+i))
	}

	// vpmuludq	f.qInv0(){1to8}, %zmm0, %zmm9;	// Compute reduction multipliers

	f.VPMULUDQ_BCST("qInvNeg+32(FP)", "Z0", "Z9")

	// INNER_MUL_1(f.qAt(0), f.qAt(2), f.qAt(1), f.qAt(3))

	INNER_MUL_1(amd64.Register(f.qAt(0)), amd64.Register(f.qAt(2)), amd64.Register(f.qAt(1)), amd64.Register(f.qAt(3)))

	// movq	1*8(PX, LEN), MUL
	f.MOVQ("1*8("+PX+")", MUL)

	// Move high dwords to zmm10-16, add each to the corresponding low dword (propagate 32-bit carries)
	// CARRY3
	// INNER_MUL_1(Y0, Y2, Y1, Y3)
	// movq	T2, MUL
	// mulxq	f.qInv0(), MUL, PH
	// CARRY4

	CARRY3()
	INNER_MUL_1(Y0, Y2, Y1, Y3)
	f.MOVQ(T2, MUL)
	f.MULXQ("qInvNeg+32(FP)", MUL, PH)
	CARRY4()

	// zmm7 keeps all 64 bits
	// INNER_MUL_2
	// movq	2*8(PX, LEN), MUL
	// MUL_W_Q_LO
	// INNER_MUL_3(Y0, Y2, Y1, Y3)
	// movq	T3, MUL
	// mulxq	f.qInv0(), MUL, PH
	// MUL_W_Q_HI
	// INNER_MUL_3(f.qAt(0), f.qAt(2), f.qAt(1), f.qAt(3))
	// movq	3*8(PX, LEN), MUL
	INNER_MUL_2()

	f.MOVQ("2*8("+PX+")", MUL)
	MUL_W_Q_LO()
	INNER_MUL_3(Y0, Y2, Y1, Y3)
	f.MOVQ(T3, MUL)
	f.MULXQ("qInvNeg+32(FP)", MUL, PH)
	MUL_W_Q_HI()
	INNER_MUL_3(amd64.Register(f.qAt(0)), amd64.Register(f.qAt(2)), amd64.Register(f.qAt(1)), amd64.Register(f.qAt(3)))
	f.MOVQ("3*8("+PX+")", MUL)

	// Propagate carries and shift down by one dword
	// CARRY1
	// INNER_MUL_4(Y0, Y2, Y1, Y3)
	// movq	T4, MUL
	// mulxq	f.qInv0(), MUL, PH
	// CARRY2
	// INNER_MUL_4(f.qAt(0), f.qAt(2), f.qAt(1), f.qAt(3))
	CARRY1()
	INNER_MUL_4(Y0, Y2, Y1, Y3)
	f.MOVQ(T4, MUL)
	f.MULXQ("qInvNeg+32(FP)", MUL, PH)
	CARRY2()
	INNER_MUL_4(amd64.Register(f.qAt(0)), amd64.Register(f.qAt(2)), amd64.Register(f.qAt(1)), amd64.Register(f.qAt(3)))

	//////////////////////////////////////////////////
	// Process doubleword 6 of x
	//////////////////////////////////////////////////

	// vpmuludq	%zmm22, %zmm24, %zmm10;		vpaddq	%zmm10, %zmm0, %zmm0;
	// vpmuludq	%zmm22, %zmm25, %zmm11;		vpaddq	%zmm11, %zmm1, %zmm1;
	// vpmuludq	%zmm22, %zmm26, %zmm12;		vpaddq	%zmm12, %zmm2, %zmm2;
	// vpmuludq	%zmm22, %zmm27, %zmm13;		vpaddq	%zmm13, %zmm3, %zmm3;
	// vpmuludq	%zmm22, %zmm28, %zmm14;		vpaddq	%zmm14, %zmm4, %zmm4;
	// vpmuludq	%zmm22, %zmm29, %zmm15;		vpaddq	%zmm15, %zmm5, %zmm5;
	// vpmuludq	%zmm22, %zmm30, %zmm16;		vpaddq	%zmm16, %zmm6, %zmm6;
	// vpmuludq	%zmm22, %zmm31, %zmm17;		vpaddq	%zmm17, %zmm7, %zmm7;

	for i := 0; i < 8; i++ {
		f.VPMULUDQ("Z22", "Z"+strconv.Itoa(24+i), "Z"+strconv.Itoa(10+i))
		f.VPADDQ("Z"+strconv.Itoa(10+i), "Z"+strconv.Itoa(i), "Z"+strconv.Itoa(i))
	}

	// vpmuludq	f.qInv0(){1to8}, %zmm0, %zmm9;	// Compute reduction multipliers
	f.VPMULUDQ_BCST("qInvNeg+32(FP)", "Z0", "Z9")

	// REDUCE(T0,T1,T2,T3,Y0,Y1,Y2,Y3)
	f.ReduceElement(t, y)

	storeOutLoadIn()

	// Move high dwords to zmm10-16, add each to the corresponding low dword (propagate 32-bit carries)
	// CARRY3
	// INNER_MUL_0
	// movq	T1, MUL
	// mulxq	f.qInv0(), MUL, PH
	// CARRY4
	CARRY3()
	INNER_MUL_0()
	f.MOVQ(T1, MUL)
	f.MULXQ("qInvNeg+32(FP)", MUL, PH)
	CARRY4()

	// zmm7 keeps all 64 bits
	// INNER_MUL_1(f.qAt(0), f.qAt(2), f.qAt(1), f.qAt(3))
	// movq	1*8(PX, LEN), MUL
	// MUL_W_Q_LO
	// INNER_MUL_1(Y0, Y2, Y1, Y3)
	// movq	T2, MUL
	// mulxq	f.qInv0(), MUL, PH
	// MUL_W_Q_HI
	// INNER_MUL_2
	// movq	2*8(PX, LEN), MUL
	INNER_MUL_1(amd64.Register(f.qAt(0)), amd64.Register(f.qAt(2)), amd64.Register(f.qAt(1)), amd64.Register(f.qAt(3)))
	f.MOVQ("1*8("+PX+")", MUL)
	MUL_W_Q_LO()
	INNER_MUL_1(Y0, Y2, Y1, Y3)
	f.MOVQ(T2, MUL)
	f.MULXQ("qInvNeg+32(FP)", MUL, PH)
	MUL_W_Q_HI()
	INNER_MUL_2()
	f.MOVQ("2*8("+PX+")", MUL)

	// Propagate carries and shift down by one dword
	// CARRY1
	// INNER_MUL_3(Y0, Y2, Y1, Y3)
	// movq	T3, MUL
	// mulxq	f.qInv0(), MUL, PH
	// CARRY2
	// INNER_MUL_3(f.qAt(0), f.qAt(2), f.qAt(1), f.qAt(3))
	// movq	3*8(PX, LEN), MUL
	CARRY1()
	INNER_MUL_3(Y0, Y2, Y1, Y3)
	f.MOVQ(T3, MUL)
	f.MULXQ("qInvNeg+32(FP)", MUL, PH)
	CARRY2()
	INNER_MUL_3(amd64.Register(f.qAt(0)), amd64.Register(f.qAt(2)), amd64.Register(f.qAt(1)), amd64.Register(f.qAt(3)))

	f.MOVQ("3*8("+PX+")", MUL)

	//////////////////////////////////////////////////
	// Process doubleword 7 of x
	//////////////////////////////////////////////////

	// vpmuludq	%zmm23, %zmm24, %zmm10;		vpaddq	%zmm10, %zmm0, %zmm0;
	// vpmuludq	%zmm23, %zmm25, %zmm11;		vpaddq	%zmm11, %zmm1, %zmm1;
	// vpmuludq	%zmm23, %zmm26, %zmm12;		vpaddq	%zmm12, %zmm2, %zmm2;
	// vpmuludq	%zmm23, %zmm27, %zmm13;		vpaddq	%zmm13, %zmm3, %zmm3;

	// INNER_MUL_4(Y0, Y2, Y1, Y3)
	// movq	T4, MUL
	// mulxq	f.qInv0(), MUL, PH

	// vpmuludq	%zmm23, %zmm28, %zmm14;		vpaddq	%zmm14, %zmm4, %zmm4;
	// vpmuludq	%zmm23, %zmm29, %zmm15;		vpaddq	%zmm15, %zmm5, %zmm5;
	// vpmuludq	%zmm23, %zmm30, %zmm16;		vpaddq	%zmm16, %zmm6, %zmm6;
	// vpmuludq	%zmm23, %zmm31, %zmm17;		vpaddq	%zmm17, %zmm7, %zmm7;

	// vpmuludq	f.qInv0(){1to8}, %zmm0, %zmm9;	// Compute reduction multipliers

	// INNER_MUL_4(f.qAt(0), f.qAt(2), f.qAt(1), f.qAt(3))

	for i := 0; i < 4; i++ {
		f.VPMULUDQ("Z23", "Z"+strconv.Itoa(24+i), "Z"+strconv.Itoa(10+i))
		f.VPADDQ("Z"+strconv.Itoa(10+i), "Z"+strconv.Itoa(i), "Z"+strconv.Itoa(i))
	}

	INNER_MUL_4(Y0, Y2, Y1, Y3)
	f.MOVQ(T4, MUL)
	f.MULXQ("qInvNeg+32(FP)", MUL, PH)

	for i := 0; i < 4; i++ {
		f.VPMULUDQ("Z23", "Z"+strconv.Itoa(28+i), "Z"+strconv.Itoa(14+i))
		f.VPADDQ("Z"+strconv.Itoa(14+i), "Z"+strconv.Itoa(4+i), "Z"+strconv.Itoa(4+i))
	}

	f.VPMULUDQ_BCST("qInvNeg+32(FP)", "Z0", "Z9")

	INNER_MUL_4(amd64.Register(f.qAt(0)), amd64.Register(f.qAt(2)), amd64.Register(f.qAt(1)), amd64.Register(f.qAt(3)))

	CARRY3()
	CARRY4()
	f.ReduceElement(t, y)
	storeOutLoadIn()

	// MUL_W_Q_LO
	// INNER_MUL_0
	// movq	T1, MUL
	// mulxq	f.qInv0(), MUL, PH
	// MUL_W_Q_HI
	// INNER_MUL_1(f.qAt(0), f.qAt(2), f.qAt(1), f.qAt(3))
	// movq	1*8(PX, LEN), MUL

	MUL_W_Q_LO()
	INNER_MUL_0()
	f.MOVQ(T1, MUL)
	f.MULXQ("qInvNeg+32(FP)", MUL, PH)
	MUL_W_Q_HI()
	INNER_MUL_1(amd64.Register(f.qAt(0)), amd64.Register(f.qAt(2)), amd64.Register(f.qAt(1)), amd64.Register(f.qAt(3)))

	f.MOVQ("1*8("+PX+")", MUL)

	// Propagate carries and shift down by one dword

	// CARRY1
	// CARRY2

	// INNER_MUL_1(Y0, Y2, Y1, Y3)
	// movq	T2, MUL
	// mulxq	f.qInv0(), MUL, PH
	// INNER_MUL_2

	// movq	2*8(PX, LEN), MUL

	CARRY1()
	CARRY2()
	INNER_MUL_1(Y0, Y2, Y1, Y3)
	f.MOVQ(T2, MUL)
	f.MULXQ("qInvNeg+32(FP)", MUL, PH)
	INNER_MUL_2()

	f.MOVQ("2*8("+PX+")", MUL)

	//////////////////////////////////////////////////
	// Conditional subtraction of the modulus
	//////////////////////////////////////////////////

	// vpermd	0*4(PM){1to16}, %zmm8, %zmm10{%k1}{z}
	// vpermd	1*4(PM){1to16}, %zmm8, %zmm11{%k1}{z}
	// vpermd	2*4(PM){1to16}, %zmm8, %zmm12{%k1}{z}
	// vpermd	3*4(PM){1to16}, %zmm8, %zmm13{%k1}{z}
	// vpermd	4*4(PM){1to16}, %zmm8, %zmm14{%k1}{z}
	// vpermd	5*4(PM){1to16}, %zmm8, %zmm15{%k1}{z}
	// vpermd	6*4(PM){1to16}, %zmm8, %zmm16{%k1}{z}
	// vpermd	7*4(PM){1to16}, %zmm8, %zmm17{%k1}{z}

	// vpsubq	%zmm10, %zmm0, %zmm10;									vpsrlq	$63, %zmm10, %zmm20;	vpandq	%zmm8, %zmm10, %zmm10
	// vpsubq	%zmm11, %zmm1, %zmm11;	vpsubq	%zmm20, %zmm11, %zmm11;	vpsrlq	$63, %zmm11, %zmm21;	vpandq	%zmm8, %zmm11, %zmm11
	// vpsubq	%zmm12, %zmm2, %zmm12;	vpsubq	%zmm21, %zmm12, %zmm12;	vpsrlq	$63, %zmm12, %zmm22;	vpandq	%zmm8, %zmm12, %zmm12
	// vpsubq	%zmm13, %zmm3, %zmm13;	vpsubq	%zmm22, %zmm13, %zmm13;	vpsrlq	$63, %zmm13, %zmm23;	vpandq	%zmm8, %zmm13, %zmm13
	// vpsubq	%zmm14, %zmm4, %zmm14;	vpsubq	%zmm23, %zmm14, %zmm14;	vpsrlq	$63, %zmm14, %zmm24;	vpandq	%zmm8, %zmm14, %zmm14
	// vpsubq	%zmm15, %zmm5, %zmm15;	vpsubq	%zmm24, %zmm15, %zmm15;	vpsrlq	$63, %zmm15, %zmm25;	vpandq	%zmm8, %zmm15, %zmm15
	// vpsubq	%zmm16, %zmm6, %zmm16;	vpsubq	%zmm25, %zmm16, %zmm16;	vpsrlq	$63, %zmm16, %zmm26;	vpandq	%zmm8, %zmm16, %zmm16
	// vpsubq	%zmm17, %zmm7, %zmm17;	vpsubq	%zmm26, %zmm17, %zmm17;

	for i := 0; i < 8; i++ {
		f.VPERMD_BCST_Z(f.qAt_bcst(i), "Z8", "K1", "Z"+strconv.Itoa(10+i))
	}

	for i := 0; i < 8; i++ {
		f.VPSUBQ("Z"+strconv.Itoa(10+i), "Z"+strconv.Itoa(i), "Z"+strconv.Itoa(10+i))
		if i > 0 {
			f.VPSUBQ("Z"+strconv.Itoa(20+i-1), "Z"+strconv.Itoa(10+i), "Z"+strconv.Itoa(10+i))
		}
		if i != 7 {
			f.VPSRLQ("$63", "Z"+strconv.Itoa(10+i), "Z"+strconv.Itoa(20+i))
			f.VPANDQ("Z8", "Z"+strconv.Itoa(10+i), "Z"+strconv.Itoa(10+i))
		}
	}

	// INNER_MUL_3(Y0, Y2, Y1, Y3)
	// movq	T3, MUL
	// vpmovq2m	%zmm17, %k2
	// mulxq	f.qInv0(), MUL, PH
	// knotb		%k2, %k2

	INNER_MUL_3(Y0, Y2, Y1, Y3)
	f.MOVQ(T3, MUL)
	f.VPMOVQ2M("Z17", "K2")
	f.MULXQ("qInvNeg+32(FP)", MUL, PH)
	f.KNOTB("K2", "K2")

	// vmovdqu64	%zmm10, %zmm0{%k2}
	// vmovdqu64	%zmm11, %zmm1{%k2}
	// vmovdqu64	%zmm12, %zmm2{%k2}
	// vmovdqu64	%zmm13, %zmm3{%k2}
	// vmovdqu64	%zmm14, %zmm4{%k2}
	// vmovdqu64	%zmm15, %zmm5{%k2}
	// vmovdqu64	%zmm16, %zmm6{%k2}
	// vmovdqu64	%zmm17, %zmm7{%k2}
	for i := 0; i < 8; i++ {
		f.VMOVDQU64k("Z"+strconv.Itoa(10+i), "K2", "Z"+strconv.Itoa(i))
	}

	// INNER_MUL_3(f.qAt(0), f.qAt(2), f.qAt(1), f.qAt(3))
	// movq	3*8(PX, LEN), MUL
	INNER_MUL_3(amd64.Register(f.qAt(0)), amd64.Register(f.qAt(2)), amd64.Register(f.qAt(1)), amd64.Register(f.qAt(3)))

	f.MOVQ("3*8("+PX+")", MUL)

	//////////////////////////////////////////////////
	// Transpose results back
	//////////////////////////////////////////////////

	// vmovdqa64	pattern1(%rip), %zmm11
	// vmovdqa64	pattern2(%rip), %zmm12
	// vmovdqa64	pattern3(%rip), %zmm13
	// vmovdqa64	pattern4(%rip), %zmm14

	// Patterns use for transposing the vectors in mulVec
	// var (
	// 	pattern1 = [8]uint64{0, 8, 1, 9, 2, 10, 3, 11}
	// 	pattern2 = [8]uint64{12, 4, 13, 5, 14, 6, 15, 7}
	// 	pattern3 = [8]uint64{0, 1, 8, 9, 2, 3, 10, 11}
	// 	pattern4 = [8]uint64{12, 13, 4, 5, 14, 15, 6, 7}
	// )

	// take care of pattern1:
	for i := 0; i < 8; i++ {
		f.MOVQ(fmt.Sprintf("·pattern1+%d(SB)", i*8), r)
		f.WriteLn(fmt.Sprintf("VPINSRQ $%d, %s, %s, %s", i, r, "X11", "X11"))
	}
	// pattern2
	for i := 0; i < 8; i++ {
		f.MOVQ(fmt.Sprintf("·pattern2+%d(SB)", i*8), r)
		f.WriteLn(fmt.Sprintf("VPINSRQ $%d, %s, %s, %s", i, r, "X12", "X12"))
	}
	// pattern3
	for i := 0; i < 8; i++ {
		f.MOVQ(fmt.Sprintf("·pattern3+%d(SB)", i*8), r)
		f.WriteLn(fmt.Sprintf("VPINSRQ $%d, %s, %s, %s", i, r, "X13", "X13"))
	}
	// pattern4
	for i := 0; i < 8; i++ {
		f.MOVQ(fmt.Sprintf("·pattern4+%d(SB)", i*8), r)
		f.WriteLn(fmt.Sprintf("VPINSRQ $%d, %s, %s, %s", i, r, "X14", "X14"))
	}

	// Step 1
	// vpsllq		$32, %zmm1, %zmm1;	vporq	%zmm1, %zmm0, %zmm0
	// vpsllq		$32, %zmm3, %zmm3;	vporq	%zmm3, %zmm2, %zmm1
	// vpsllq		$32, %zmm5, %zmm5;	vporq	%zmm5, %zmm4, %zmm2
	// vpsllq		$32, %zmm7, %zmm7;	vporq	%zmm7, %zmm6, %zmm3

	for i := 0; i < 4; i++ {
		f.VPSLLQ("$32", "Z"+strconv.Itoa(2*i+1), "Z"+strconv.Itoa(2*i+1))
		f.VPORQ("Z"+strconv.Itoa(2*i+1), "Z"+strconv.Itoa(2*i), "Z"+strconv.Itoa(i))
	}

	// INNER_MUL_4(Y0, Y2, Y1, Y3)
	// movq	T4, MUL
	INNER_MUL_4(Y0, Y2, Y1, Y3)
	f.MOVQ(T4, MUL)

	// vmovdqu64	%zmm0, %zmm4
	// vmovdqu64	%zmm2, %zmm6

	// vpermt2q	%zmm1, %zmm11, %zmm0
	// vpermt2q	%zmm4, %zmm12, %zmm1
	// vpermt2q	%zmm3, %zmm11, %zmm2
	// vpermt2q	%zmm6, %zmm12, %zmm3

	f.VMOVDQU64("Z0", "Z4")
	f.VMOVDQU64("Z2", "Z6")

	f.VPERMT2Q("Z1", "Z11", "Z0")
	f.VPERMT2Q("Z4", "Z12", "Z1")
	f.VPERMT2Q("Z3", "Z11", "Z2")
	f.VPERMT2Q("Z6", "Z12", "Z3")

	// 	mulxq	f.qInv0(), MUL, PH
	f.MULXQ("qInvNeg+32(FP)", MUL, PH)

	// Step 3

	// vmovdqu64	%zmm0, %zmm4
	// vmovdqu64	%zmm1, %zmm5

	// vpermt2q	%zmm2, %zmm13, %zmm0
	// vpermt2q	%zmm4, %zmm14, %zmm2
	// vpermt2q	%zmm3, %zmm13, %zmm1
	// vpermt2q	%zmm5, %zmm14, %zmm3

	f.VMOVDQU64("Z0", "Z4")
	f.VMOVDQU64("Z1", "Z5")

	f.VPERMT2Q("Z2", "Z13", "Z0")
	f.VPERMT2Q("Z4", "Z14", "Z2")
	f.VPERMT2Q("Z3", "Z13", "Z1")
	f.VPERMT2Q("Z5", "Z14", "Z3")

	// INNER_MUL_4(f.qAt(0), f.qAt(2), f.qAt(1), f.qAt(3))
	// REDUCE(T0,T1,T2,T3,Y0,Y1,Y2,Y3)
	INNER_MUL_4(amd64.Register(f.qAt(0)), amd64.Register(f.qAt(2)), amd64.Register(f.qAt(1)), amd64.Register(f.qAt(3)))
	f.ReduceElement(t, y)

	// Store output

	// movq	T0, 0*8(PZ, LEN)
	// movq	T1, 1*8(PZ, LEN)
	// movq	T2, 2*8(PZ, LEN)
	// movq	T3, 3*8(PZ, LEN)

	// addq	$32, LEN

	// TODO @gbotrel check that.
	f.ADDQ("$32", PX)
	f.MOVQ(PY, r)
	f.ADDQ("$32", r)
	f.MOVQ(r, PY)

	f.MOVQ(PZ, r)
	f.Mov(t, r)
	f.ADDQ("$32", r)
	f.MOVQ(r, PZ)

	// //////////////////////////////////////////////////
	// // Save AVX-512 results
	// //////////////////////////////////////////////////

	// vmovdqu64	%zmm0, 0*64(PZ, LEN)
	// vmovdqu64	%zmm2, 1*64(PZ, LEN)
	// vmovdqu64	%zmm1, 2*64(PZ, LEN)
	// vmovdqu64	%zmm3, 3*64(PZ, LEN)

	f.VMOVDQU64("Z0", "0*64("+r+")")
	f.VMOVDQU64("Z2", "1*64("+r+")")
	f.VMOVDQU64("Z1", "2*64("+r+")")
	f.VMOVDQU64("Z3", "3*64("+r+")")

	// TODO @gbotrel probably not.
	f.DECQ(LEN, "decrement n")
	f.JMP(loop)

	f.Comment("available registers:" + strconv.Itoa(registers.Available()))

	f.LABEL(done)

	f.RET()

	f.Push(&registers, PY, PZ)

}
