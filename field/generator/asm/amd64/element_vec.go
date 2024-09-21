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
	f.Mov(f.Q, q)
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
	stackSize := f.StackSize(f.NbWords*2+4, 0, 0)
	registers := f.FnHeader("sumVec", stackSize, argSize, amd64.DX, amd64.AX)
	defer f.AssertCleanStack(stackSize, 0)

	// registers & labels we need
	addrA := f.Pop(&registers)
	len := f.Pop(&registers)
	tmp0 := f.Pop(&registers)

	t := f.PopN(&registers)
	s := f.PopN(&registers)
	t4 := f.Pop(&registers)

	loop := f.NewLabel("loop")
	done := f.NewLabel("done")
	rr1 := f.NewLabel("rr1")
	rr2 := f.NewLabel("rr2")
	accumulate := f.NewLabel("accumulate")
	// propagate := f.NewLabel("propagate")

	// AVX512 registers
	Z0 := amd64.Register("Z0")
	Z1 := amd64.Register("Z1")
	Z2 := amd64.Register("Z2")
	Z3 := amd64.Register("Z3")
	Z4 := amd64.Register("Z4")
	X0 := amd64.Register("X0")

	K1 := amd64.Register("K1")
	K2 := amd64.Register("K2")
	K3 := amd64.Register("K3")

	f.MOVQ("$0x1555", tmp0)
	f.KMOVW(tmp0, K1)

	f.MOVQ("$0xff80", tmp0)
	f.KMOVW(tmp0, K2)

	f.MOVQ("$0x01ff", tmp0)
	f.KMOVW(tmp0, K3)

	// load arguments
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("n+16(FP)", len)

	// initialize accumulators to zero (zmm0, zmm1, zmm2, zmm3)
	f.VXORPS(Z0, Z0, Z0)
	f.VMOVDQA64(Z0, Z1)
	f.VMOVDQA64(Z0, Z2)
	f.VMOVDQA64(Z0, Z3)

	f.TESTQ(len, len)
	f.JEQ(done, "n == 0, we are done")

	f.MOVQ(len, tmp0)
	f.ANDQ("$3", tmp0) // t0 = n % 4
	f.SHRQ("$2", len)  // len = n / 4

	// if len % 4 != 0, we need to handle the remaining elements
	f.CMPB(tmp0, "$1")
	f.JEQ(rr1, "we have 1 remaining element")

	f.CMPB(tmp0, "$2")
	f.JEQ(rr2, "we have 2 remaining elements")

	f.CMPB(tmp0, "$3")
	f.JNE(loop, "== 0; we have 0 remaining elements")

	f.Comment("we have 3 remaining elements")
	// vpmovzxdq 	2*32(PX), %zmm4;	vpaddq	%zmm4, %zmm0, %zmm0
	f.VPMOVZXDQ("2*32("+addrA+")", Z4)
	f.VPADDQ(Z4, Z0, Z0)

	f.LABEL(rr2)
	f.Comment("we have 2 remaining elements")
	// vpmovzxdq 	1*32(PX), %zmm4;	vpaddq	%zmm4, %zmm1, %zmm1
	f.VPMOVZXDQ("1*32("+addrA+")", Z4)
	f.VPADDQ(Z4, Z1, Z1)

	f.LABEL(rr1)
	f.Comment("we have 1 remaining element")
	// vpmovzxdq 	0*32(PX), %zmm4;	vpaddq	%zmm4, %zmm2, %zmm2
	f.VPMOVZXDQ("0*32("+addrA+")", Z4)
	f.VPADDQ(Z4, Z2, Z2)

	f.LABEL(loop)
	f.TESTQ(len, len)
	f.JEQ(accumulate, "n == 0, we are going to accumulate")

	f.VPMOVZXDQ("0*32("+addrA+")", Z4)
	f.VPADDQ(Z4, Z0, Z0)

	f.VPMOVZXDQ("1*32("+addrA+")", Z4)
	f.VPADDQ(Z4, Z1, Z1)

	f.VPMOVZXDQ("2*32("+addrA+")", Z4)
	f.VPADDQ(Z4, Z2, Z2)

	f.VPMOVZXDQ("3*32("+addrA+")", Z4)
	f.VPADDQ(Z4, Z3, Z3)

	f.Comment("increment pointers to visit next 4 elements")
	f.ADDQ("$128", addrA)
	f.DECQ(len, "decrement n")
	f.JMP(loop)

	f.LABEL(accumulate)

	f.VPADDQ(Z1, Z0, Z0)
	f.VPADDQ(Z3, Z2, Z2)
	f.VPADDQ(Z2, Z0, Z0)

	// Propagate carries
	f.VMOVQ(X0, t[0])
	f.VALIGNQ("$1", Z0, Z0, Z0)
	f.VMOVQ(X0, t[1])
	f.VALIGNQ("$1", Z0, Z0, Z0)
	f.VMOVQ(X0, t[2])
	f.VALIGNQ("$1", Z0, Z0, Z0)
	f.VMOVQ(X0, t[3])
	f.VALIGNQ("$1", Z0, Z0, Z0)
	f.VMOVQ(X0, s[0])
	f.VALIGNQ("$1", Z0, Z0, Z0)
	f.VMOVQ(X0, s[1])
	f.VALIGNQ("$1", Z0, Z0, Z0)
	f.VMOVQ(X0, s[2])
	f.VALIGNQ("$1", Z0, Z0, Z0)
	f.VMOVQ(X0, s[3])

	w0l := t[0]
	w0h := t[1]
	w1l := t[2]
	w1h := t[3]
	w2l := s[0]
	w2h := s[1]
	w3l := s[2]
	w3h := s[3]

	// r0 = w0l + lo(woh)
	// r1 = carry + hi(woh) + w1l + lo(w1h)
	// r2 = carry + hi(w1h) + w2l + lo(w2h)
	// r3 = carry + hi(w2h) + w3l + lo(w3h)
	r0 := w0l
	r1 := w1l
	r2 := w2l
	r3 := w3l

	// we need 2 carry so we use ADOXQ and ADCXQ
	f.XORQ(amd64.AX, amd64.AX)

	// get low bits of w0h
	f.MOVQ(w0h, amd64.AX)
	f.ANDQ("$0xffffffff", amd64.AX)
	f.SHLQ("$32", amd64.AX)
	f.SHRQ("$32", w0h)

	// start the carry chain
	f.ADOXQ(amd64.AX, w0l) // w0l is good.

	// get low bits of w1h
	f.MOVQ(w1h, amd64.AX)
	f.ANDQ("$0xffffffff", amd64.AX)
	f.SHLQ("$32", amd64.AX)
	f.SHRQ("$32", w1h)

	f.ADOXQ(amd64.AX, w1l)
	f.ADCXQ(w0h, w1l)

	// get low bits of w2h
	f.MOVQ(w2h, amd64.AX)
	f.ANDQ("$0xffffffff", amd64.AX)
	f.SHLQ("$32", amd64.AX)
	f.SHRQ("$32", w2h)

	f.ADOXQ(amd64.AX, w2l)
	f.ADCXQ(w1h, w2l)

	// get low bits of w3h
	f.MOVQ(w3h, amd64.AX)
	f.ANDQ("$0xffffffff", amd64.AX)
	f.SHLQ("$32", amd64.AX)
	f.SHRQ("$32", w3h)

	f.ADOXQ(amd64.AX, w3l)
	f.ADCXQ(w2h, w3l)
	r4 := w3h
	f.MOVQ("$0", amd64.AX)
	f.ADOXQ(amd64.AX, r4)
	f.ADCXQ(amd64.AX, r4)

	// // we use AX for low 32bits
	// f.MOVQ(t[1], amd64.AX)
	// f.ANDQ("$0xffffffff", amd64.AX)
	// f.SHRQ("$32", t[1])

	// // start the carry chain
	// f.ADDQ(amd64.AX, t[0]) // t0 is good.
	// // now t1, we have to add t1 + low(t2)

	// // // Propagate carries
	// // mov	$8, %eax
	// // valignd	$1, %zmm3, %zmm0, %zmm3{%k2}{z}	// Shift lowest dword of zmm0 into zmm3
	// f.MOVQ("$8", tmp0)
	// f.VALIGND("$1", Z3, Z0, K2, Z3)

	// f.LABEL(propagate)
	// f.VPSRLQ("$32", Z0, Z1)
	// f.VALIGND("$2", Z0, Z0, K1, Z0)
	// f.VPADDQ(Z1, Z0, Z0)
	// f.VALIGND("$1", Z3, Z0, K2, Z3)

	// f.DECQ(tmp0)
	// f.JNE(propagate)

	// // The top 9 dwords of zmm3 now contain the sum
	// // we shift by 224 bits to get the result in the low 32bytes

	// // // Move intermediate result to integer registers
	// // The top 9 dwords of zmm3 now contain the sum. Copy them to the low end of zmm0.
	// // valignd	$7, %zmm3, %zmm3, %zmm0{%k3}{z}
	// // // Copy to integer registers
	// // vmovq	%xmm0, T0;	valignq	$1, %zmm0, %zmm0, %zmm0
	// // vmovq	%xmm0, T1;	valignq	$1, %zmm0, %zmm0, %zmm0
	// // vmovq	%xmm0, T2;	valignq	$1, %zmm0, %zmm0, %zmm0
	// // vmovq	%xmm0, T3;	valignq	$1, %zmm0, %zmm0, %zmm0
	// // vmovq	%xmm0, T4

	// f.VALIGND("$7", Z3, Z3, K3, Z0)

	// f.VMOVQ(X0, t[0])
	// f.VALIGNQ("$1", Z0, Z0, Z0)
	// f.VMOVQ(X0, t[1])
	// f.VALIGNQ("$1", Z0, Z0, Z0)
	// f.VMOVQ(X0, t[2])
	// f.VALIGNQ("$1", Z0, Z0, Z0)
	// f.VMOVQ(X0, t[3])
	// f.VALIGNQ("$1", Z0, Z0, Z0)
	// f.VMOVQ(X0, t4)

	f.MOVQ("res+0(FP)", addrA)
	r := []amd64.Register{r0, r1, r2, r3}
	f.Mov(r, addrA)

	f.RET()

	// Reduce using single-word Barrett
	// q1 is low 32 bits of T4 and high 32 bits of T3
	// movq	T3, %rax
	// shrd	$32, T4, %rax
	// mulq	MU		// Multiply by mu. q2 in rdx:rax, q3 in rdx
	f.MOVQ(f.mu(), tmp0)
	f.MOVQ(t[3], amd64.AX)
	f.SHRQw("$32", t4, amd64.AX)
	f.MULQ(tmp0)

	// Subtract r2 from r1
	// mulx	0*8(PM), PL, PH; sub	PL, T0; sbb	PH, T1;
	// mulx	2*8(PM), PL, PH; sbb	PL, T2; sbb	PH, T3;	sbb	$0, T4
	// mulx	1*8(PM), PL, PH; sub	PL, T1; sbb	PH, T2;
	// mulx	3*8(PM), PL, PH; sbb	PL, T3; sbb	PH, T4
	f.MULXQ(f.qAt(0), amd64.AX, tmp0)
	f.SUBQ(amd64.AX, t[0])
	f.SBBQ(tmp0, t[1])

	f.MULXQ(f.qAt(2), amd64.AX, tmp0)
	f.SBBQ(amd64.AX, t[2])
	f.SBBQ(tmp0, t[3])
	f.SBBQ("$0", t4)

	f.MULXQ(f.qAt(1), amd64.AX, tmp0)
	f.SUBQ(amd64.AX, t[1])
	f.SBBQ(tmp0, t[2])

	f.MULXQ(f.qAt(3), amd64.AX, tmp0)
	f.SBBQ(amd64.AX, t[3])
	f.SBBQ(tmp0, t4)

	// Two conditional subtractions to guarantee canonicity of the result
	// substract modulus from t
	f.Mov(f.Q, s)

	f.MOVQ("res+0(FP)", addrA)
	f.Mov(t, addrA)

	f.Sub(s, t)
	f.SBBQ("$0", t4)
	// if borrow, skip to end
	f.JCS(done)

	f.Mov(t, addrA)
	f.Sub(s, t)
	f.SBBQ("$0", t4)
	// if borrow, skip to end
	f.JCS(done)

	f.Mov(t, addrA)

	// save t into res

	f.LABEL(done)

	// save t into res

	// f.Mov(t, addrA)

	f.RET()
	f.Push(&registers, addrA, len, tmp0, t4)
	f.Push(&registers, t...)
	f.Push(&registers, s...)
}
