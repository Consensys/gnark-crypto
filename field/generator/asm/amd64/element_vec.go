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
	reserved = append(reserved, mul4Registers...)
	registers := f.FnHeader("scalarMulVec", stackSize, argSize, reserved...)
	defer f.AssertCleanStack(stackSize, minStackSize)

	// labels & registers we need
	noAdx := f.NewLabel("noAdx")
	loop := f.NewLabel("loop")
	done := f.NewLabel("done")

	t := mul4Registers[:f.NbWords]
	scalar := mul4Registers[f.NbWords : f.NbWords*2]

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

	f.MOVQ("res+0(FP)", addrRes)

	f.LABEL(loop)
	f.TESTQ(len, len)
	f.JEQ(done, "n == 0, we are done")

	// reuse defines from the mul function
	mulWord0 := f.DefineFn("MUL_WORD_0")
	mulWordN := f.DefineFn("MUL_WORD_N")
	for i := 0; i < f.NbWords; i++ {
		f.MOVQ(addrA.At(i), amd64.DX)
		if i == 0 {
			mulWord0()
		} else {
			mulWordN()
		}
	}

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

	// TODO @gbotrel prefetch
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
	stackSize := f.StackSize(6+f.NbWords, 2, 8)
	reserved := make([]amd64.Register, len(mul4Registers)+2)
	copy(reserved, mul4Registers)
	reserved[len(mul4Registers)] = amd64.AX
	reserved[len(mul4Registers)+1] = amd64.DX
	registers := f.FnHeader("mulVec", stackSize, argSize, reserved...)
	defer f.AssertCleanStack(stackSize, 0)

	// to simplify the generated assembly, we only handle n/16 (and do blocks of 16 muls).
	// that is if n%16 != 0, we let the caller (Go) handle the remaining elements.
	LEN := f.Pop(&registers, true)
	PZ := f.Pop(&registers)
	PX := f.Pop(&registers)
	PY := f.Pop(&registers)

	f.Comment("couple of defines")

	zi := func(i int) string {
		return "Z" + strconv.Itoa(i)
	}

	// AVX_MUL_Q_LO:
	// 	vpmuludq	0*4(PM){1to8}, %zmm9, %zmm10;	vpaddq	%zmm10, %zmm0, %zmm0;	// Low dword of zmm0 is zero
	// 	vpmuludq	1*4(PM){1to8}, %zmm9, %zmm11;	vpaddq	%zmm11, %zmm1, %zmm1;
	// 	vpmuludq	2*4(PM){1to8}, %zmm9, %zmm12;	vpaddq	%zmm12, %zmm2, %zmm2;
	// 	vpmuludq	3*4(PM){1to8}, %zmm9, %zmm13;	vpaddq	%zmm13, %zmm3, %zmm3;
	AVX_MUL_Q_LO := f.Define("AVX_MUL_Q_LO", 0, func(args ...amd64.Register) {
		for i := 0; i < 4; i++ {
			f.VPMULUDQ_BCST(f.qAt_bcst(i), "Z9", zi(10+i))
			f.VPADDQ(zi(10+i), zi(i), zi(i))
		}
	})

	// AVX_MUL_Q_HI:
	// 	vpmuludq	4*4(PM){1to8}, %zmm9, %zmm14;	vpaddq	%zmm14, %zmm4, %zmm4;
	// 	vpmuludq	5*4(PM){1to8}, %zmm9, %zmm15;	vpaddq	%zmm15, %zmm5, %zmm5;
	// 	vpmuludq	6*4(PM){1to8}, %zmm9, %zmm16;	vpaddq	%zmm16, %zmm6, %zmm6;
	// 	vpmuludq	7*4(PM){1to8}, %zmm9, %zmm17;	vpaddq	%zmm17, %zmm7, %zmm7;
	AVX_MUL_Q_HI := f.Define("AVX_MUL_Q_HI", 0, func(args ...amd64.Register) {
		for i := 0; i < 4; i++ {
			f.VPMULUDQ_BCST(f.qAt_bcst(i+4), "Z9", zi(14+i))
			f.VPADDQ(zi(14+i), zi(i+4), zi(i+4))
		}
	})

	// CARRY1:
	// 	vpsrlq		$32, %zmm0, %zmm10;		vpaddq	%zmm10, %zmm1, %zmm1;	vpandq	%zmm8, %zmm1, %zmm0
	// 	vpsrlq		$32, %zmm1, %zmm11;		vpaddq	%zmm11, %zmm2, %zmm2;	vpandq	%zmm8, %zmm2, %zmm1
	// 	vpsrlq		$32, %zmm2, %zmm12;		vpaddq	%zmm12, %zmm3, %zmm3;	vpandq	%zmm8, %zmm3, %zmm2
	// 	vpsrlq		$32, %zmm3, %zmm13;		vpaddq	%zmm13, %zmm4, %zmm4;	vpandq	%zmm8, %zmm4, %zmm3
	CARRY1 := f.Define("CARRY1", 0, func(args ...amd64.Register) {
		for i := 0; i < 4; i++ {
			f.VPSRLQ("$32", zi(i), zi(10+i))
			f.VPADDQ(zi(10+i), zi(i+1), zi(i+1))
			f.VPANDQ("Z8", zi(i+1), zi(i))
		}
	})

	// CARRY2:
	// 	vpsrlq		$32, %zmm4, %zmm14;		vpaddq	%zmm14, %zmm5, %zmm5;	vpandq	%zmm8, %zmm5, %zmm4
	// 	vpsrlq		$32, %zmm5, %zmm15;		vpaddq	%zmm15, %zmm6, %zmm6;	vpandq	%zmm8, %zmm6, %zmm5
	// 	vpsrlq		$32, %zmm6, %zmm16;		vpaddq	%zmm16, %zmm7, %zmm7;	vpandq	%zmm8, %zmm7, %zmm6
	// 	vpsrlq		$32, %zmm7, %zmm7
	CARRY2 := f.Define("CARRY2", 0, func(args ...amd64.Register) {
		for i := 0; i < 3; i++ {
			f.VPSRLQ("$32", zi(i+4), zi(14+i))
			f.VPADDQ(zi(14+i), zi(i+5), zi(i+5))
			f.VPANDQ("Z8", zi(i+5), zi(i+4))
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
			f.VPSRLQ("$32", zi(i), zi(10+i))
			f.VPANDQ("Z8", zi(i), zi(i))
			f.VPADDQ(zi(10+i), zi(i+1), zi(i+1))
		}
	})

	// CARRY4:
	// 	vpsrlq		$32, %zmm4, %zmm14;		vpandq	%zmm8, %zmm4, %zmm4;	vpaddq	%zmm14, %zmm5, %zmm5;
	// 	vpsrlq		$32, %zmm5, %zmm15;		vpandq	%zmm8, %zmm5, %zmm5;	vpaddq	%zmm15, %zmm6, %zmm6;
	// 	vpsrlq		$32, %zmm6, %zmm16;		vpandq	%zmm8, %zmm6, %zmm6;	vpaddq	%zmm16, %zmm7, %zmm7;
	CARRY4 := f.Define("CARRY4", 0, func(args ...amd64.Register) {
		for i := 0; i < 3; i++ {
			f.VPSRLQ("$32", zi(i+4), zi(14+i))
			f.VPANDQ("Z8", zi(i+4), zi(i+4))
			f.VPADDQ(zi(14+i), zi(i+5), zi(i+5))
		}
	})

	// we use the same registers as defined in the mul.
	t := mul4Registers[:4]
	f.LabelRegisters("t", t...)
	y := mul4Registers[4:8]
	f.LabelRegisters("y", y...)
	tr := mul4Registers[8]
	A := amd64.BP // note, BP is used in the mul defines.

	// reuse defines from the mul function
	mulWord0 := f.DefineFn("MUL_WORD_0")
	mulWordN := f.DefineFn("MUL_WORD_N")

	zIndex := 0

	loadInput := func() {
		f.Comment(fmt.Sprintf("load input y[%d]", zIndex))
		f.Mov(PY, y, zIndex*4)
	}

	mulXi := func(wordIndex int) {
		f.Comment(fmt.Sprintf("z[%d] -> y * x[%d]", zIndex, wordIndex))
		if wordIndex == 0 {
			mulWord0()
		} else {
			f.MOVQ(amd64.Register(PX.At(wordIndex)), amd64.DX)
			mulWordN()
		}
	}

	storeOutput := func() {
		scratch := []amd64.Register{A, tr, amd64.AX, amd64.DX}
		f.ReduceElement(t, scratch)

		f.Comment(fmt.Sprintf("store output z[%d]", zIndex))
		f.Mov(t, PZ, 0, zIndex*4)
		if zIndex == 7 {
			f.ADDQ("$288", PX)
		} else {
			f.ADDQ("$32", PX)
			f.MOVQ(amd64.Register(PX.At(0)), amd64.DX)
		}
		zIndex++
	}

	done := f.NewLabel("done")
	loop := f.NewLabel("loop")

	f.MOVQ("res+0(FP)", PZ)
	f.MOVQ("a+8(FP)", PX)
	f.MOVQ("b+16(FP)", PY)
	f.MOVQ("n+24(FP)", tr)

	// we process 16 elements at a time, so we divide by 16
	// f.SHRQ("$4", tr)
	f.MOVQ(tr, LEN)

	// Create mask for low dword in each qword
	// vpcmpeqb	%ymm8, %ymm8, %ymm8
	// vpmovzxdq	%ymm8, %zmm8
	// mov	$0x5555, %edx
	// kmovd	%edx, %k1

	f.VPCMPEQB("Y8", "Y8", "Y8")
	f.VPMOVZXDQ("Y8", "Z8")
	f.MOVQ("$0x5555", amd64.DX)
	f.KMOVD(amd64.DX, "K1")

	f.LABEL(loop)
	// f.MOVQ(LEN, tr)
	f.TESTQ(tr, tr)
	f.JEQ(done, "n == 0, we are done")

	f.MOVQ(amd64.Register(PX.At(0)), amd64.DX)
	f.VMOVDQU64("256+0*64("+PX+")", "Z16")
	f.VMOVDQU64("256+1*64("+PX+")", "Z17")
	f.VMOVDQU64("256+2*64("+PX+")", "Z18")
	f.VMOVDQU64("256+3*64("+PX+")", "Z19")

	loadInput()
	f.VMOVDQU64("256+0*64("+PY+")", "Z24")
	f.VMOVDQU64("256+1*64("+PY+")", "Z25")
	f.VMOVDQU64("256+2*64("+PY+")", "Z26")
	f.VMOVDQU64("256+3*64("+PY+")", "Z27")

	f.Comment("Transpose and expand x and y")

	// Step 1

	f.VSHUFI64X2("$0x88", "Z17", "Z16", "Z20")
	f.VSHUFI64X2("$0xdd", "Z17", "Z16", "Z22")
	f.VSHUFI64X2("$0x88", "Z19", "Z18", "Z21")
	f.VSHUFI64X2("$0xdd", "Z19", "Z18", "Z23")

	f.VSHUFI64X2("$0x88", "Z25", "Z24", "Z28")
	f.VSHUFI64X2("$0xdd", "Z25", "Z24", "Z30")
	f.VSHUFI64X2("$0x88", "Z27", "Z26", "Z29")
	f.VSHUFI64X2("$0xdd", "Z27", "Z26", "Z31")

	// Step 2

	f.VPERMQ("$0xd8", "Z20", "Z20")
	f.VPERMQ("$0xd8", "Z21", "Z21")
	f.VPERMQ("$0xd8", "Z22", "Z22")
	f.VPERMQ("$0xd8", "Z23", "Z23")

	mulXi(0)

	f.VPERMQ("$0xd8", "Z28", "Z28")
	f.VPERMQ("$0xd8", "Z29", "Z29")
	f.VPERMQ("$0xd8", "Z30", "Z30")
	f.VPERMQ("$0xd8", "Z31", "Z31")

	// Step 3

	for i := 20; i <= 23; i++ {
		f.VSHUFI64X2("$0xd8", zi(i), zi(i), zi(i))
	}

	mulXi(1)

	for i := 28; i <= 31; i++ {
		f.VSHUFI64X2("$0xd8", zi(i), zi(i), zi(i))
	}

	// Step 4

	f.VSHUFI64X2("$0x44", "Z21", "Z20", "Z16")
	f.VSHUFI64X2("$0xee", "Z21", "Z20", "Z18")
	f.VSHUFI64X2("$0x44", "Z23", "Z22", "Z20")
	f.VSHUFI64X2("$0xee", "Z23", "Z22", "Z22")

	mulXi(2)
	f.VSHUFI64X2("$0x44", "Z29", "Z28", "Z24")
	f.VSHUFI64X2("$0xee", "Z29", "Z28", "Z26")
	f.VSHUFI64X2("$0x44", "Z31", "Z30", "Z28")
	f.VSHUFI64X2("$0xee", "Z31", "Z30", "Z30")

	f.WriteLn("PREFETCHT0 1024(" + string(PX) + ")")

	// Step 5

	f.VPSRLQ("$32", "Z16", "Z17")
	f.VPSRLQ("$32", "Z18", "Z19")
	f.VPSRLQ("$32", "Z20", "Z21")
	f.VPSRLQ("$32", "Z22", "Z23")

	for i := 24; i <= 30; i += 2 {
		f.VPSRLQ("$32", zi(i), zi(i+1))
	}
	mulXi(3)

	for i := 16; i <= 30; i += 2 {
		f.VPANDQ("Z8", zi(i), zi(i))
	}

	storeOutput()
	loadInput()

	f.Comment("For each 256-bit input value, each zmm register now represents a 32-bit input word zero-extended to 64 bits.")
	f.Comment("Multiply y by doubleword 0 of x")

	for i := 0; i < 8; i++ {
		f.VPMULUDQ("Z16", zi(24+i), zi(i))
		if i == 4 {
			f.WriteLn("PREFETCHT0 1024(" + string(PY) + ")")
		}
	}

	f.VPMULUDQ_BCST("qInvNeg+32(FP)", "Z0", "Z9")

	for i := 0; i < 4; i++ {
		f.VPSRLQ("$32", zi(i), zi(10+i))
		f.VPANDQ("Z8", zi(i), zi(i))
		f.VPADDQ(zi(10+i), zi(i+1), zi(i+1))

	}

	mulXi(0)

	for i := 0; i < 3; i++ {
		f.VPSRLQ("$32", zi(4+i), zi(14+i))
		f.VPANDQ("Z8", zi(4+i), zi(4+i))
		f.VPADDQ(zi(14+i), zi(5+i), zi(5+i))

	}

	for i := 0; i < 4; i++ {
		f.VPMULUDQ_BCST(f.qAt_bcst(i), "Z9", zi(10+i))
		f.VPADDQ(zi(10+i), zi(i), zi(i))
	}

	mulXi(1)

	f.VPMULUDQ_BCST(f.qAt_bcst(4), "Z9", "Z14")
	f.VPADDQ("Z14", "Z4", "Z4")

	f.VPMULUDQ_BCST(f.qAt_bcst(5), "Z9", "Z15")
	f.VPADDQ("Z15", "Z5", "Z5")

	f.VPMULUDQ_BCST(f.qAt_bcst(6), "Z9", "Z16")
	f.VPADDQ("Z16", "Z6", "Z6")

	f.VPMULUDQ_BCST(f.qAt_bcst(7), "Z9", "Z10")
	f.VPADDQ("Z10", "Z7", "Z7")

	for i := 0; i < 4; i++ {
		f.VPSRLQ("$32", zi(i), zi(10+i))
		f.VPADDQ(zi(10+i), zi(i+1), zi(i+1))
		f.VPANDQ("Z8", zi(i+1), zi(i))

	}

	mulXi(2)

	for i := 0; i < 3; i++ {
		f.VPSRLQ("$32", zi(4+i), zi(14+i))
		f.VPADDQ(zi(14+i), zi(4+i+1), zi(4+i+1))
		f.VPANDQ("Z8", zi(4+i+1), zi(4+i))

	}
	f.VPSRLQ("$32", "Z7", "Z7")

	f.Comment("Process doubleword 1 of x")

	for i := 0; i < 4; i++ {
		f.VPMULUDQ("Z17", zi(24+i), zi(10+i))
		f.VPADDQ(zi(10+i), zi(i), zi(i))

	}

	mulXi(3)

	for i := 0; i < 4; i++ {
		f.VPMULUDQ("Z17", zi(28+i), zi(14+i))
		f.VPADDQ(zi(14+i), zi(4+i), zi(4+i))

	}

	f.VPMULUDQ_BCST("qInvNeg+32(FP)", "Z0", "Z9")

	storeOutput()
	loadInput()

	f.Comment("Move high dwords to zmm10-16, add each to the corresponding low dword (propagate 32-bit carries)")

	for i := 0; i < 3; i++ {
		f.VPSRLQ("$32", zi(i), zi(10+i))
		f.VPANDQ("Z8", zi(i), zi(i))
		f.VPADDQ(zi(10+i), zi(i+1), zi(i+1))
	}

	f.VPSRLQ("$32", "Z3", "Z13")
	f.VPANDQ("Z8", "Z3", "Z3")
	f.VPADDQ("Z13", "Z4", "Z4")

	CARRY4()
	mulXi(0)

	AVX_MUL_Q_LO()

	AVX_MUL_Q_HI()
	mulXi(1)

	CARRY1()

	CARRY2()
	mulXi(2)

	f.Comment("Process doubleword 2 of x")

	for i := 0; i < 4; i++ {
		f.VPMULUDQ("Z18", zi(24+i), zi(10+i))
		f.VPADDQ(zi(10+i), zi(i), zi(i))
	}

	for i := 0; i < 4; i++ {
		f.VPMULUDQ("Z18", zi(28+i), zi(14+i))
		f.VPADDQ(zi(14+i), zi(4+i), zi(4+i))

	}

	f.VPMULUDQ_BCST("qInvNeg+32(FP)", "Z0", "Z9")

	mulXi(3)

	f.Comment("Move high dwords to zmm10-16, add each to the corresponding low dword (propagate 32-bit carries)")
	CARRY3()

	storeOutput()
	loadInput()

	CARRY4()

	AVX_MUL_Q_LO()

	mulXi(0)
	AVX_MUL_Q_HI()

	CARRY1()
	mulXi(1)

	CARRY2()

	f.Comment("Process doubleword 3 of x")

	for i := 0; i < 4; i++ {
		f.VPMULUDQ("Z19", zi(24+i), zi(10+i))
		f.VPADDQ(zi(10+i), zi(i), zi(i))

	}
	mulXi(2)

	for i := 0; i < 4; i++ {
		f.VPMULUDQ("Z19", zi(28+i), zi(14+i))
		f.VPADDQ(zi(14+i), zi(4+i), zi(4+i))

	}

	f.VPMULUDQ_BCST("qInvNeg+32(FP)", "Z0", "Z9")

	// Move high dwords to zmm10-16, add each to the corresponding low dword (propagate 32-bit carries)
	CARRY3()
	mulXi(3)
	CARRY4()
	storeOutput()
	loadInput()
	AVX_MUL_Q_LO()

	AVX_MUL_Q_HI()

	mulXi(0)

	f.Comment("Propagate carries and shift down by one dword")
	CARRY1()

	CARRY2()

	mulXi(1)

	f.Comment("Process doubleword 4 of x")

	for i := 0; i < 4; i++ {
		f.VPMULUDQ("Z20", zi(24+i), zi(10+i))
		f.VPADDQ(zi(10+i), zi(i), zi(i))

	}

	for i := 0; i < 4; i++ {
		f.VPMULUDQ("Z20", zi(28+i), zi(14+i))
		f.VPADDQ(zi(14+i), zi(4+i), zi(4+i))

	}

	f.VPMULUDQ_BCST("qInvNeg+32(FP)", "Z0", "Z9")

	mulXi(2)

	f.Comment("Move high dwords to zmm10-16, add each to the corresponding low dword (propagate 32-bit carries)")

	CARRY3()

	CARRY4()

	mulXi(3)

	f.Comment("zmm7 keeps all 64 bits")

	AVX_MUL_Q_LO()

	AVX_MUL_Q_HI()

	storeOutput()

	f.Comment("Propagate carries and shift down by one dword")

	CARRY1()

	CARRY2()

	loadInput()

	f.Comment("Process doubleword 5 of x")

	for i := 0; i < 4; i++ {
		f.VPMULUDQ("Z21", zi(24+i), zi(10+i))
		f.VPADDQ(zi(10+i), zi(i), zi(i))

	}

	for i := 0; i < 4; i++ {
		f.VPMULUDQ("Z21", zi(28+i), zi(14+i))
		f.VPADDQ(zi(14+i), zi(4+i), zi(4+i))

	}

	mulXi(0)

	f.VPMULUDQ_BCST("qInvNeg+32(FP)", "Z0", "Z9")

	f.Comment("Move high dwords to zmm10-16, add each to the corresponding low dword (propagate 32-bit carries)")
	CARRY3()

	CARRY4()

	mulXi(1)

	AVX_MUL_Q_LO()

	AVX_MUL_Q_HI()

	mulXi(2)

	CARRY1()

	CARRY2()

	mulXi(3)

	f.Comment("Process doubleword 6 of x")

	for i := 0; i < 8; i++ {
		f.VPMULUDQ("Z22", zi(24+i), zi(10+i))
		f.VPADDQ(zi(10+i), zi(i), zi(i))
	}

	f.VPMULUDQ_BCST("qInvNeg+32(FP)", "Z0", "Z9")

	storeOutput()
	loadInput()

	f.Comment("Move high dwords to zmm10-16, add each to the corresponding low dword (propagate 32-bit carries)")
	CARRY3()

	CARRY4()

	mulXi(0)

	AVX_MUL_Q_LO()

	AVX_MUL_Q_HI()

	mulXi(1)

	CARRY1()

	CARRY2()

	mulXi(2)

	f.Comment("Process doubleword 7 of x")
	for i := 0; i < 4; i++ {
		f.VPMULUDQ("Z23", zi(24+i), zi(10+i))
		f.VPADDQ(zi(10+i), zi(i), zi(i))

	}

	for i := 0; i < 4; i++ {
		f.VPMULUDQ("Z23", zi(28+i), zi(14+i))
		f.VPADDQ(zi(14+i), zi(4+i), zi(4+i))

	}
	f.VPMULUDQ_BCST("qInvNeg+32(FP)", "Z0", "Z9")

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

	f.Comment("Conditional subtraction of the modulus")

	for i := 0; i < 8; i++ {
		f.VPERMD_BCST_Z(f.qAt_bcst(i), "Z8", "K1", zi(10+i))
	}

	for i := 0; i < 8; i++ {
		f.VPSUBQ(zi(10+i), zi(i), zi(10+i))
		if i > 0 {
			f.VPSUBQ(zi(20+i-1), zi(10+i), zi(10+i))
		}
		if i != 7 {
			f.VPSRLQ("$63", zi(10+i), zi(20+i))
			f.VPANDQ("Z8", zi(10+i), zi(10+i))
		}

	}

	f.VPMOVQ2M("Z17", "K2")
	f.KNOTB("K2", "K2")

	for i := 0; i < 8; i++ {
		f.VMOVDQU64k(zi(10+i), "K2", zi(i))
	}

	mulXi(2)

	f.Comment("Transpose results back")

	f.WriteLn("VALIGND $0, ·pattern1+0(SB), Z11, Z11")
	f.WriteLn("VALIGND $0, ·pattern2+0(SB), Z12, Z12")
	f.WriteLn("VALIGND $0, ·pattern3+0(SB), Z13, Z13")
	f.WriteLn("VALIGND $0, ·pattern4+0(SB), Z14, Z14")

	for i := 0; i < 4; i++ {
		f.VPSLLQ("$32", zi(2*i+1), zi(2*i+1))
		f.VPORQ(zi(2*i+1), zi(2*i), zi(i))
	}

	f.VMOVDQU64("Z0", "Z4")
	f.VMOVDQU64("Z2", "Z6")

	f.VPERMT2Q("Z1", "Z11", "Z0")
	f.VPERMT2Q("Z4", "Z12", "Z1")
	f.VPERMT2Q("Z3", "Z11", "Z2")
	f.VPERMT2Q("Z6", "Z12", "Z3")

	// Step 3

	mulXi(3)

	f.VMOVDQU64("Z0", "Z4")
	f.VMOVDQU64("Z1", "Z5")
	f.VPERMT2Q("Z2", "Z13", "Z0")
	f.VPERMT2Q("Z4", "Z14", "Z2")
	f.VPERMT2Q("Z3", "Z13", "Z1")
	f.VPERMT2Q("Z5", "Z14", "Z3")

	storeOutput()

	f.Comment("Save AVX-512 results")

	f.VMOVDQU64("Z0", "256+0*64("+PZ+")")
	f.VMOVDQU64("Z2", "256+1*64("+PZ+")")
	f.VMOVDQU64("Z1", "256+2*64("+PZ+")")
	f.VMOVDQU64("Z3", "256+3*64("+PZ+")")

	f.ADDQ("$512", PZ)
	f.ADDQ("$512", PY)

	f.MOVQ(LEN, tr)
	f.DECQ(tr, "decrement n")
	f.MOVQ(tr, LEN)
	f.JMP(loop)

	f.LABEL(done)

	f.RET()

	f.Push(&registers, LEN)

}
