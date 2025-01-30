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

	"github.com/consensys/bavard/amd64"
)

func (f *FFAmd64) generateAddVecF31() {
	f.Comment("addVec(res, a, b *Element, n uint64) res[0...n] = a[0...n] + b[0...n]")
	f.Comment("n is the number of blocks of 16 elements to process")

	const argSize = 4 * 8
	stackSize := f.StackSize(f.NbWords*2+4, 0, 0)
	registers := f.FnHeader("addVec", stackSize, argSize)
	defer f.AssertCleanStack(stackSize, 0)

	// registers & labels we need
	addrA := f.Pop(&registers)
	addrB := f.Pop(&registers)
	addrRes := f.Pop(&registers)
	len := f.Pop(&registers)

	// AVX512 registers
	a := amd64.Register("Z0")
	b := amd64.Register("Z1")
	t := amd64.Register("Z2")
	q := amd64.Register("Z3")

	// load q in Z3
	f.MOVD("$const_q", amd64.AX)
	f.VPBROADCASTD(amd64.AX, q)

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
	f.VMOVDQU32(addrA.At(0), a)
	f.VMOVDQU32(addrB.At(0), b)
	f.VPADDD(a, b, a, "a = a + b")
	// t = a - q
	f.VPSUBD(q, a, t, "t = a - q")
	// b = min(t, a)
	f.VPMINUD(a, t, b, "b = min(t, a)")

	// move b to res
	f.VMOVDQU32(b, addrRes.At(0), "res = b")

	f.Comment("increment pointers to visit next element")
	f.ADDQ("$64", addrA)
	f.ADDQ("$64", addrB)
	f.ADDQ("$64", addrRes)
	f.DECQ(len, "decrement n")
	f.JMP(loop)

	f.LABEL(done)

	f.RET()

	f.Push(&registers, addrA, addrB, addrRes, len)

}

func (f *FFAmd64) generateSubVecF31() {
	f.Comment("subVec(res, a, b *Element, n uint64) res[0...n] = a[0...n] - b[0...n]")
	f.Comment("n is the number of blocks of 16 elements to process")

	const argSize = 4 * 8
	stackSize := f.StackSize(f.NbWords*2+4, 0, 0)
	registers := f.FnHeader("subVec", stackSize, argSize)
	defer f.AssertCleanStack(stackSize, 0)

	// registers & labels we need
	addrA := f.Pop(&registers)
	addrB := f.Pop(&registers)
	addrRes := f.Pop(&registers)
	len := f.Pop(&registers)

	// AVX512 registers
	a := amd64.Register("Z0")
	b := amd64.Register("Z1")
	t := amd64.Register("Z2")
	q := amd64.Register("Z3")

	// load q in Z3
	f.MOVD("$const_q", amd64.AX)
	f.VPBROADCASTD(amd64.AX, q)

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

	// a = a - b
	f.VMOVDQU32(addrA.At(0), a)
	f.VMOVDQU32(addrB.At(0), b)

	f.VPSUBD(b, a, a, "a = a - b")

	// t = a + q
	f.VPADDD(q, a, t, "t = a + q")

	// b = min(t, a)
	f.VPMINUD(a, t, b, "b = min(t, a)")

	// move b to res
	f.VMOVDQU32(b, addrRes.At(0), "res = b")

	f.Comment("increment pointers to visit next element")
	f.ADDQ("$64", addrA)
	f.ADDQ("$64", addrB)
	f.ADDQ("$64", addrRes)
	f.DECQ(len, "decrement n")
	f.JMP(loop)

	f.LABEL(done)

	f.RET()

	f.Push(&registers, addrA, addrB, addrRes, len)

}

// sumVec res = sum(a[0...n])
func (f *FFAmd64) generateSumVecF31() {
	f.Comment("sumVec(res *uint64, a *[]uint32, n uint64) res = sum(a[0...n])")
	f.Comment("n is the number of blocks of 16 elements to process")
	const argSize = 3 * 8
	stackSize := f.StackSize(f.NbWords*3+2, 0, 0)
	registers := f.FnHeader("sumVec", stackSize, argSize, amd64.DX, amd64.AX)
	defer f.AssertCleanStack(stackSize, 0)

	f.WriteLn(`
	// We load 8 31bits values at a time and accumulate them into an accumulator of
	// 8 quadwords (64bits). The caller then needs to reduce the result mod q.
	// We can safely accumulate ~2**33 31bits values into a single accumulator.
	// That gives us a maximum of 2**33 * 8 = 2**36 31bits values to sum safely.
	`)

	// registers & labels we need
	addrA := f.Pop(&registers)
	addrT := f.Pop(&registers)
	len := f.Pop(&registers)

	// AVX512 registers
	a1 := amd64.Register("Z0")
	a2 := amd64.Register("Z1")
	acc1 := amd64.Register("Z2")
	acc2 := amd64.Register("Z3")

	loop := f.NewLabel("loop")
	done := f.NewLabel("done")

	// load arguments
	f.MOVQ("t+0(FP)", addrT)
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("n+16(FP)", len)

	// zeroize the accumulators
	f.VXORPS(acc1, acc1, acc1, "acc1 = 0")
	f.VMOVDQA64(acc1, acc2, "acc2 = 0")

	f.LABEL(loop)

	f.TESTQ(len, len)
	f.JEQ(done, "n == 0, we are done")

	// 1 cache line is typically 64 bytes, so we maintain 2 accumulators
	f.VPMOVZXDQ(addrA.At(0), a1, "load 8 31bits values in a1")
	f.VPMOVZXDQ(addrA.At(4), a2, "load 8 31bits values in a2")

	f.VPADDQ(a1, acc1, acc1, "acc1 += a1")
	f.VPADDQ(a2, acc2, acc2, "acc2 += a2")

	f.Comment("increment pointers to visit next element")
	f.ADDQ("$64", addrA)
	f.DECQ(len, "decrement n")
	f.JMP(loop)

	f.LABEL(done)

	// store t into res
	f.VPADDQ(acc1, acc2, acc1, "acc1 += acc2")
	f.VMOVDQU64(acc1, addrT.At(0), "res = acc1")

	f.RET()

	f.Push(&registers, addrA, addrT, len)
}

// mulVec res = a * b
func (f *FFAmd64) generateMulVecF31() {
	f.Comment("mulVec(res, a, b *Element, n uint64) res[0...n] = a[0...n] * b[0...n]")
	f.Comment("n is the number of blocks of 8 elements to process")
	const argSize = 4 * 8
	stackSize := f.StackSize(f.NbWords*2+4, 0, 0)
	registers := f.FnHeader("mulVec", stackSize, argSize, amd64.AX)
	defer f.AssertCleanStack(stackSize, 0)

	f.Comment("code inspired by Plonky3: https://github.com/Plonky3/Plonky3/blob/36e619f3c6526ee86e2e5639a24b3224e1c1700f/monty-31/src/x86_64_avx512/packing.rs#L319")

	// registers & labels we need
	addrA := f.Pop(&registers)
	addrB := f.Pop(&registers)
	addrRes := f.Pop(&registers)
	len := f.Pop(&registers)

	// AVX512 registers
	q := registers.PopV()
	qInvNeg := registers.PopV()

	// load q in Z3
	f.MOVD("$const_q", amd64.AX)
	f.VPBROADCASTD(amd64.AX, q)
	f.MOVD("$const_qInvNeg", amd64.AX)
	f.VPBROADCASTD(amd64.AX, qInvNeg)

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

	const kBlendEven2 = 0b0101010101010101
	f.MOVQ(uint64(kBlendEven2), amd64.AX)
	f.KMOVD(amd64.AX, amd64.K3)

	a := registers.PopV()
	b := registers.PopV()
	b1 := registers.PopV()
	aOdd := registers.PopV()
	bOdd := registers.PopV()
	b0 := registers.PopV()
	PL0 := registers.PopV()
	PL1 := registers.PopV()
	// a = a * b
	f.VMOVDQU32(addrA.At(0), a)
	f.VMOVDQU32(addrB.At(0), b)
	f.VMOVSHDUP(a, aOdd)
	f.VMOVSHDUP(b, bOdd)
	f.VPMULUDQ(a, b, b0)
	f.VPMULUDQ(aOdd, bOdd, b1)
	f.VPMULUDQ(b0, qInvNeg, PL0)
	f.VPMULUDQ(b1, qInvNeg, PL1)

	f.VPMULUDQ(PL0, q, PL0)
	f.VPMULUDQ(PL1, q, PL1)

	f.VPADDQ(b0, PL0, b0)
	f.VPADDQ(b1, PL1, b1)

	f.VMOVSHDUPk(b0, amd64.K3, b1)

	f.VPSUBD(q, b1, PL1)
	f.VPMINUD(b1, PL1, b1)

	f.VMOVDQU32(b1, addrRes.At(0), "res = P")

	f.Comment("increment pointers to visit next element")
	f.ADDQ("$64", addrA)
	f.ADDQ("$64", addrB)
	f.ADDQ("$64", addrRes)
	f.DECQ(len, "decrement n")
	f.JMP(loop)

	f.LABEL(done)

	f.RET()

	f.Push(&registers, addrA, addrB, addrRes, len)

}

// scalarMulVec res = a * b
func (f *FFAmd64) generateScalarMulVecF31() {
	f.Comment("scalarMulVec(res, a, b *Element, n uint64) res[0...n] = a[0...n] * b")
	f.Comment("n is the number of blocks of 8 elements to process")
	const argSize = 4 * 8
	stackSize := f.StackSize(f.NbWords*2+4, 0, 0)
	registers := f.FnHeader("scalarMulVec", stackSize, argSize)
	defer f.AssertCleanStack(stackSize, 0)

	// registers & labels we need
	addrA := f.Pop(&registers)
	addrB := f.Pop(&registers)
	addrRes := f.Pop(&registers)
	len := f.Pop(&registers)

	// AVX512 registers
	a := amd64.Register("Z0")
	b := amd64.Register("Z1")
	P := amd64.Register("Z2")
	q := amd64.Register("Z3")
	qInvNeg := amd64.Register("Z4")
	PL := amd64.Register("Z5")
	LSW := amd64.Register("Z6")

	// load q in Z3
	f.MOVD("$const_q", amd64.AX)
	f.VPBROADCASTQ(amd64.AX, q)
	f.MOVD("$const_qInvNeg", amd64.AX)
	f.VPBROADCASTQ(amd64.AX, qInvNeg)

	f.Comment("Create mask for low dword in each qword")
	f.VPCMPEQB("Y0", "Y0", "Y0")
	f.VPMOVZXDQ("Y0", LSW)

	loop := f.NewLabel("loop")
	done := f.NewLabel("done")

	// load arguments
	f.MOVQ("res+0(FP)", addrRes)
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("b+16(FP)", addrB)
	f.MOVQ("n+24(FP)", len)

	f.VPBROADCASTD(addrB.At(0), b)

	f.LABEL(loop)

	f.TESTQ(len, len)
	f.JEQ(done, "n == 0, we are done")

	// a = a * b
	f.VPMOVZXDQ(addrA.At(0), a)

	f.VPMULUDQ(a, b, P, "P = a * b")
	f.VPANDQ(LSW, P, PL, "m = uint32(P)")
	f.VPMULUDQ(PL, qInvNeg, PL, "m = m * qInvNeg")
	f.VPANDQ(LSW, PL, PL, "m = uint32(m)")
	f.VPMULUDQ(PL, q, PL, "m = m * q")
	f.VPADDQ(P, PL, P, "P = P + m")
	f.VPSRLQ("$32", P, P, "P = P >> 32")

	f.VPSUBQ(q, P, PL, "PL = P - q")
	f.VPMINUQ(P, PL, P, "P = min(P, PL)")

	// move P to res
	f.VPMOVQD(P, addrRes.At(0), "res = P")

	f.Comment("increment pointers to visit next element")
	f.ADDQ("$32", addrA)
	f.ADDQ("$32", addrRes)
	f.DECQ(len, "decrement n")
	f.JMP(loop)

	f.LABEL(done)

	f.RET()

	f.Push(&registers, addrA, addrB, addrRes, len)
}

// innerProdVec res = sum(a * b)
func (f *FFAmd64) generateInnerProdVecF31() {
	f.Comment("innerProdVec(t *uint64, a,b *[]uint32, n uint64) res = sum(a[0...n] * b[0...n])")
	f.Comment("n is the number of blocks of 8 elements to process")
	const argSize = 4 * 8
	stackSize := f.StackSize(f.NbWords*4+2, 0, 0)
	registers := f.FnHeader("innerProdVec", stackSize, argSize, amd64.DX, amd64.AX)
	defer f.AssertCleanStack(stackSize, 0)

	f.WriteLn(`
	// Similar to mulVec; we do most of the montgomery multiplication but don't do
	// the final reduction. We accumulate the result like in sumVec and let the caller
	// reduce mod q.
	`)

	// registers & labels we need
	addrA := f.Pop(&registers)
	addrB := f.Pop(&registers)
	addrT := f.Pop(&registers)
	len := f.Pop(&registers)

	// AVX512 registers
	a := amd64.Register("Z0")
	b := amd64.Register("Z1")
	acc := amd64.Register("Z2")
	q := amd64.Register("Z3")
	qInvNeg := amd64.Register("Z4")
	PL := amd64.Register("Z5")
	LSW := amd64.Register("Z6")
	P := amd64.Register("Z7")

	loop := f.NewLabel("loop")
	done := f.NewLabel("done")

	f.MOVD("$const_q", amd64.AX)
	f.VPBROADCASTQ(amd64.AX, q)
	f.MOVD("$const_qInvNeg", amd64.AX)
	f.VPBROADCASTQ(amd64.AX, qInvNeg)

	f.Comment("Create mask for low dword in each qword")
	f.VPCMPEQB("Y0", "Y0", "Y0")
	f.VPMOVZXDQ("Y0", LSW)

	// zeroize the accumulators
	f.VXORPS(acc, acc, acc, "acc = 0")

	// load arguments
	f.MOVQ("t+0(FP)", addrT)
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("b+16(FP)", addrB)
	f.MOVQ("n+24(FP)", len)

	f.LABEL(loop)

	f.TESTQ(len, len)
	f.JEQ(done, "n == 0, we are done")

	f.VPMOVZXDQ(addrA.At(0), a)
	f.VPMOVZXDQ(addrB.At(0), b)

	f.VPMULUDQ(a, b, P, "P = a * b")
	f.VPANDQ(LSW, P, PL, "m = uint32(P)")
	f.VPMULUDQ(PL, qInvNeg, PL, "m = m * qInvNeg")
	f.VPANDQ(LSW, PL, PL, "m = uint32(m)")
	f.VPMULUDQ(PL, q, PL, "m = m * q")
	f.VPADDQ(P, PL, P, "P = P + m")
	f.VPSRLQ("$32", P, P, "P = P >> 32")

	// we can accumulate ~2**32 32bits values into a single accumulator without overflow;
	// that gives us a maximum of 2**32 * 8 = 2**35 32bits values to sum safely.
	f.Comment("accumulate P into acc, P is in [0, 2q] on 32bits max")
	f.VPADDQ(P, acc, acc, "acc += P")

	f.Comment("increment pointers to visit next element")
	f.ADDQ("$32", addrA)
	f.ADDQ("$32", addrB)
	f.DECQ(len, "decrement n")
	f.JMP(loop)

	f.LABEL(done)

	// store t into res
	f.VMOVDQU64(acc, addrT.At(0), "res = acc")

	f.RET()

	f.Push(&registers, addrA, addrT, len)
}

func (f *FFAmd64) generateFFTDefinesF31() {
	f.Comment("performs a butterfly between 2 vectors of dwords")
	f.Comment("in0 = (in0 + in1) mod q")
	f.Comment("in1 = (in0 - in1) mod 2q")
	f.Comment("in2: q broadcasted on all dwords lanes")
	f.Comment("in3: temporary Z register")
	butterflyD2Q := f.Define("butterflyD2Q", 4, func(args ...any) {
		x := args[0]
		y := args[1]
		qd := args[2]
		b0 := args[3]
		f.VPADDD(x, y, b0)  // b0 = x + y
		f.VPSUBD(y, x, y)   // y = x - y
		f.VPSUBD(qd, b0, x) // x = (x+y) - q
		f.VPMINUD(b0, x, x) // x %= q
		f.VPADDD(qd, y, y)  // y = (x-y) + q --> y in [0,2q)
	})

	// almost the same but we also keep result of the add in [0,2q)
	_ = f.Define("butterflyD2Q2Q", 4, func(args ...any) {
		x := args[0]
		y := args[1]
		qd := args[2]
		b0 := args[3]
		f.VPSUBD(y, x, b0)  // y = x - y
		f.VPADDD(x, y, x)   // b0 = x + y
		f.VPADDD(qd, b0, y) // y = (x-y) + q --> y in [0,2q)
	})

	f.Comment("same as butterflyD2Q but reduces in1 to [0,q)")
	butterflyD1Q := f.Define("butterflyD1Q", 5, func(args ...any) {
		x := args[0]
		y := args[1]
		qd := args[2]
		b0 := args[3]
		b1 := args[4]
		f.VPADDD(x, y, b0)  // b0 = x + y
		f.VPSUBD(y, x, y)   // y = x - y
		f.VPSUBD(qd, b0, x) // x = (x+y) - q
		f.VPMINUD(b0, x, x) // x %= q
		f.VPADDD(qd, y, b1) // y = (x-y) + q --> y in [0,2q)
		f.VPMINUD(b1, y, y) // y %= q
	})

	f.Comment("same as butterflyD2Q but for qwords")
	f.Comment("in2: must be broadcasted on all qwords lanes")
	_ = f.Define("butterflyQ2Q", 4, func(args ...any) {
		x := args[0]
		y := args[1]
		q := args[2]
		b0 := args[3]

		f.VPADDQ(x, y, b0)  // b0 = x + y
		f.VPSUBQ(y, x, y)   // y = x - y
		f.VPSUBQ(q, b0, x)  // x = (x+y) - q
		f.VPMINUQ(b0, x, x) // x %= q
		f.VPADDQ(q, y, y)   // y = (x-y) + q --> y in [0,2q)
	})

	_ = f.Define("butterflyQ1Q", 5, func(args ...any) {
		x := args[0]
		y := args[1]
		q := args[2]
		b0 := args[3]
		b1 := args[4]

		f.VPADDQ(x, y, b0)  // b0 = x + y
		f.VPSUBQ(y, x, y)   // y = x - y
		f.VPSUBQ(q, b0, x)  // x = (x+y) - q
		f.VPMINUQ(b0, x, x) // x %= q
		f.VPADDQ(q, y, b1)  // y = (x-y) + q --> y in [0,2q)
		f.VPMINUQ(b1, y, y) // y %= q
	})

	f.Comment("performs a multiplication in place between 2 vectors of qwords (values should be dwords zero extended)")
	f.Comment("in0 = (in0 * in1) mod q")
	f.Comment("in1: second operand")
	f.Comment("in2: mask for low dword in each qword")
	f.Comment("in3: q broadcasted on all qwords lanes")
	f.Comment("in4: qInvNeg broadcasted on all qwords lanes")
	f.Comment("in5: temporary Z register")
	f.Comment("in6: temporary Z register")
	_ = f.Define("mulQ", 7, func(args ...any) {
		x := args[0]
		y := args[1]
		LSW := args[2]
		q := args[3]
		qInvNeg := args[4]
		P := args[5]
		PL := args[6]

		// TODO @gbotrel can avoid the ANDQ with LSW since MULUDQ ignores the high dword
		f.VPMULUDQ(x, y, P)
		f.VPANDQ(LSW, P, PL)
		f.VPMULUDQ(PL, qInvNeg, PL)
		f.VPANDQ(LSW, PL, PL)
		f.VPMULUDQ(PL, q, PL)
		f.VPADDQ(P, PL, P)
		f.VPSRLQ("$32", P, P)
		f.VPSUBQ(q, P, PL)
		f.VPMINUQ(P, PL, x)
	})

	mulD := f.Define("mulD", 11, func(args ...any) {
		a := args[0]
		b := args[1]
		aOdd := args[2]
		bOdd := args[3]
		b0 := args[4]
		b1 := args[5]
		PL0 := args[6]
		PL1 := args[7]
		q := args[8]
		qInvNeg := args[9]
		kEvens := args[10]

		f.VMOVSHDUP(a, aOdd)
		f.VMOVSHDUP(b, bOdd)
		f.VPMULUDQ(a, b, b0)
		f.VPMULUDQ(aOdd, bOdd, b1)
		f.VPMULUDQ(b0, qInvNeg, PL0)
		f.VPMULUDQ(b1, qInvNeg, PL1)

		f.VPMULUDQ(PL0, q, PL0)
		f.VPMULUDQ(PL1, q, PL1)

		f.VPADDQ(b0, PL0, b0)
		f.VPADDQ(b1, PL1, b1)

		f.VMOVSHDUPk(b0, kEvens, b1)

		// this also works but is slower.
		// f.VPCMPUD(5, b1, q, amd64.K4)
		// f.VPSUBD(q, b1, a)
		// f.VPADDDk(q, a, amd64.K4, a)

		f.VPSUBD(q, b1, PL1)
		f.VPMINUD(b1, PL1, a)
	})

	f.WriteLn(`
	// goes from
	// Z1 = A A A A B B B B
	// Z2 = C C C C D D D D
	// we want
	// Z1 = A A A A C C C C
	// Z2 = B B B B D D D D`)
	_ = f.Define("permute8x8", 4, func(args ...any) {
		x := args[0]
		y := args[1]
		b0 := args[2]
		K := args[3]
		f.VSHUFI64X2(uint64(0b01_00_11_10), y, x, b0)
		f.VPBLENDMQ(x, b0, x, K)
		f.VPBLENDMQ(b0, y, y, K)
	})

	f.WriteLn(`
	// Z1 = A A B B C C D D
	// Z2 = L L M M N N O O
	// we want
	// Z1 = A A L L C C N N
	// Z2 = B B M M D D O O`)
	_ = f.Define("permute4x4", 5, func(args ...any) {
		x := args[0]
		y := args[1]
		vInterleaveIndices := args[2]
		tmp := args[3]
		K := args[4]
		f.VMOVDQA64(vInterleaveIndices, tmp)
		f.VPERMI2Q(y, x, tmp)
		f.VPBLENDMQ(x, tmp, x, K)
		f.VPBLENDMQ(tmp, y, y, K)
	})

	_ = f.Define("permute2x2", 4, func(args ...any) {
		x := args[0]
		y := args[1]
		b0 := args[2]
		K := args[3]

		f.VSHUFPD(0b01010101, y, x, b0)
		f.VPBLENDMQ(x, b0, x, K)
		f.VPBLENDMQ(b0, y, y, K)
	})

	_ = f.Define("permute1x1", 4, func(args ...any) {
		x := args[0]
		y := args[1]
		b0 := args[2]
		K := args[3]

		f.VPSHRDQ("$32", y, x, b0)
		f.VPBLENDMD(x, b0, x, K)
		f.VPBLENDMD(b0, y, y, K)
	})

	_ = f.Define("PACK_DWORDS", 4, func(args ...any) {
		x := args[0]
		xx := args[1]
		y := args[2]
		xy := args[3]

		f.VPMOVQD(x, xx)
		f.VPMOVQD(y, xy)
		f.VINSERTI64X4(1, xy, x, x)
	})

	_ = f.Define("butterfly_mulD", 11+4, func(args ...any) {
		butterflyD2Q(args[0], args[1], args[2], args[3])
		mulD(args[4:]...)
	})

	_ = f.Define("mul_butterflyD", 11+5, func(args ...any) {
		mulD(args[5:]...)
		butterflyD1Q(args[0], args[1], args[2], args[3], args[4])
	})
}

func (f *FFAmd64) generateFFTInnerDITF31() {
	// func innerDITWithTwiddles(a []Element, twiddles []Element, start, end, m int) {
	// 	for i := start; i < end; i++ {
	// 		a[i+m].Mul(&a[i+m], &twiddles[i])
	// 		Butterfly(&a[i], &a[i+m])
	// 	}
	// }
	const argSize = 9 * 8
	stackSize := f.StackSize(f.NbWords*2+4, 1, 0)
	registers := f.FnHeader("innerDITWithTwiddles_avx512", stackSize, argSize, amd64.AX)
	defer f.AssertCleanStack(stackSize, 0)

	addrA := f.Pop(&registers)
	addrAPlusM := f.Pop(&registers)
	addrTwiddles := f.Pop(&registers)
	m := f.Pop(&registers)
	len := f.Pop(&registers)

	a := registers.PopV()
	am := registers.PopV()
	b0 := registers.PopV()
	b1 := registers.PopV()
	qd := registers.PopV()
	qInvNeg := registers.PopV()
	t0 := registers.PopV()
	aOdd := registers.PopV()
	bOdd := registers.PopV()
	PL0 := registers.PopV()
	PL1 := registers.PopV()

	f.Comment("prepare constants needed for mul and reduce ops")
	f.MOVD("$const_q", amd64.AX)
	f.VPBROADCASTD(amd64.AX, qd)
	f.MOVD("$const_qInvNeg", amd64.AX)
	f.VPBROADCASTD(amd64.AX, qInvNeg)

	f.Comment("load arguments")
	f.MOVQ("a+0(FP)", addrA)
	f.MOVQ("twiddles+24(FP)", addrTwiddles)
	f.MOVQ("end+56(FP)", len)
	f.MOVQ("m+64(FP)", m)

	f.MOVQ(uint64(0b0101010101010101), amd64.AX)
	f.KMOVD(amd64.AX, amd64.K3)

	// get defines
	butterflyD1Q, _ := f.DefineFn("butterflyD1Q")
	mulD, _ := f.DefineFn("mulD")

	// we do only m >= 8;
	// if m < 8, we call the generic one; this can be called when doing a FFT
	// smaller than the smallest generated kernel
	lblSmallerThan16 := f.NewLabel("smallerThan16")
	f.CMPQ(m, 16)
	f.JL(lblSmallerThan16, "m < 16")

	f.SHRQ("$4", len, "we are processing 16 elements at a time")

	// offset we want to add to a is m*4bytes
	f.SHLQ("$2", m, "offset = m * 4bytes")

	f.MOVQ(addrA, addrAPlusM)
	f.ADDQ(m, addrAPlusM)

	lblDone := f.NewLabel("done")
	lblLoop := f.NewLabel("loop")

	f.LABEL(lblLoop)

	f.TESTQ(len, len)
	f.JEQ(lblDone, "n == 0, we are done")

	f.VMOVDQU32(addrA.At(0), a, "load a[i]")
	f.VMOVDQU32(addrAPlusM.At(0), am, "load a[i+m]")
	f.VMOVDQU32(addrTwiddles.At(0), t0)

	mulD(am, t0, aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg, amd64.K3)
	butterflyD1Q(a, am, qd, b0, b1)

	// a is ready to be stored, but we need to scale am by twiddles.
	f.VMOVDQU32(a, addrA.At(0), "store a[i]")
	f.VMOVDQU32(am, addrAPlusM.At(0), "store a[i+m]")

	f.ADDQ("$64", addrA)
	f.ADDQ("$64", addrAPlusM)
	f.ADDQ("$64", addrTwiddles)
	f.DECQ(len, "decrement n")
	f.JMP(lblLoop)

	f.LABEL(lblDone)

	f.RET()

	f.LABEL(lblSmallerThan16)
	f.Comment("m < 16, we call the generic one")
	f.Comment("note that this should happen only when doing a FFT smaller than the smallest generated kernel")

	// TODO @gbotrel should have dedicated tests
	f.MOVQ("a+0(FP)", amd64.AX)
	f.MOVQ(amd64.AX, "(SP)")
	f.MOVQ("twiddles+24(FP)", amd64.AX)
	f.MOVQ(amd64.AX, "24(SP)") // go vet says 24(SP) should be a_cap+16(FP)
	f.MOVQ("start+48(FP)", amd64.AX)
	f.MOVQ(amd64.AX, "48(SP)") // go vet says 48(SP) should be twiddles_cap+40(FP)
	f.MOVQ("end+56(FP)", amd64.AX)
	f.MOVQ(amd64.AX, "56(SP)")
	f.MOVQ("m+64(FP)", amd64.AX)
	f.MOVQ(amd64.AX, "64(SP)")

	f.WriteLn("CALL ·innerDITWithTwiddlesGeneric(SB)")
	f.RET()

}

func (f *FFAmd64) generateFFTInnerDIFF31() {
	// func innerDIFWithTwiddles(a []Element, twiddles []Element, start, end, m int) {
	// 	for i := start; i < end; i++ {
	// 		Butterfly(&a[i], &a[i+m])
	// 		a[i+m].Mul(&a[i+m], &twiddles[i])
	// 	}
	// }
	const argSize = 9 * 8
	stackSize := f.StackSize(f.NbWords*2+4, 1, 0)
	registers := f.FnHeader("innerDIFWithTwiddles_avx512", stackSize, argSize, amd64.AX)
	defer f.AssertCleanStack(stackSize, 0)

	addrA := f.Pop(&registers)
	addrAPlusM := f.Pop(&registers)
	addrTwiddles := f.Pop(&registers)
	m := f.Pop(&registers)
	len := f.Pop(&registers)

	a := registers.PopV()
	am := registers.PopV()
	qd := registers.PopV()
	b0 := registers.PopV()
	q := registers.PopV()
	qInvNeg := registers.PopV()
	t0 := registers.PopV()

	aOdd := registers.PopV()
	bOdd := registers.PopV()
	b1 := registers.PopV()
	PL0 := registers.PopV()
	PL1 := registers.PopV()

	f.Comment("prepare constants needed for mul and reduce ops")
	f.MOVD("$const_q", amd64.AX)
	f.VPBROADCASTD(amd64.AX, qd)
	f.VPBROADCASTD(amd64.AX, q)
	f.MOVD("$const_qInvNeg", amd64.AX)
	f.VPBROADCASTD(amd64.AX, qInvNeg)

	f.Comment("load arguments")
	f.MOVQ("a+0(FP)", addrA)
	f.MOVQ("twiddles+24(FP)", addrTwiddles)
	f.MOVQ("end+56(FP)", len)
	f.MOVQ("m+64(FP)", m)

	// get defines
	butterflyMulD, _ := f.DefineFn("butterfly_mulD")

	// we do only m >= 16;
	// if m < 16, we call the generic one; this can be called when doing a FFT
	// smaller than the smallest generated kernel
	lblSmallerThan16 := f.NewLabel("smallerThan16")
	f.CMPQ(m, 16)
	f.JL(lblSmallerThan16, "m < 16")

	// we are processing elements 16x16 so we divide len by 16
	f.SHRQ("$4", len, "we are processing 16 elements at a time")

	// offset we want to add to a is m*4bytes
	f.SHLQ("$2", m, "offset = m * 4bytes")

	f.MOVQ(addrA, addrAPlusM)
	f.ADDQ(m, addrAPlusM)

	const kBlendEven2 = 0b0101010101010101
	f.MOVQ(uint64(kBlendEven2), amd64.AX)
	f.KMOVD(amd64.AX, amd64.K3)

	lblDone := f.NewLabel("done")
	lblLoop := f.NewLabel("loop")

	f.LABEL(lblLoop)

	f.TESTQ(len, len)
	f.JEQ(lblDone, "n == 0, we are done")

	f.VMOVDQU32(addrA.At(0), a, "load a[i]")
	f.VMOVDQU32(addrAPlusM.At(0), am, "load a[i+m]")

	f.VMOVDQU32(addrTwiddles.At(0), t0)
	butterflyMulD(a, am, qd, b0,
		am, t0, aOdd, bOdd, b0, b1, PL0, PL1, q, qInvNeg, amd64.K3)

	f.VMOVDQU32(a, addrA.At(0), "store a[i]")
	f.VMOVDQU32(am, addrAPlusM.At(0))

	f.ADDQ("$64", addrA)
	f.ADDQ("$64", addrAPlusM)
	f.ADDQ("$64", addrTwiddles)
	f.DECQ(len, "decrement n")
	f.JMP(lblLoop)

	f.LABEL(lblDone)

	f.RET()

	f.LABEL(lblSmallerThan16)
	f.Comment("m < 16, we call the generic one")
	f.Comment("note that this should happen only when doing a FFT smaller than the smallest generated kernel")

	// TODO @gbotrel should have dedicated tests
	f.MOVQ("a+0(FP)", amd64.AX)
	f.MOVQ(amd64.AX, "(SP)")
	f.MOVQ("twiddles+24(FP)", amd64.AX)
	f.MOVQ(amd64.AX, "24(SP)") // go vet says 24(SP) should be a_cap+16(FP)
	f.MOVQ("start+48(FP)", amd64.AX)
	f.MOVQ(amd64.AX, "48(SP)") // go vet says 48(SP) should be twiddles_cap+40(FP)
	f.MOVQ("end+56(FP)", amd64.AX)
	f.MOVQ(amd64.AX, "56(SP)")
	f.MOVQ("m+64(FP)", amd64.AX)
	f.MOVQ(amd64.AX, "64(SP)")

	f.WriteLn("CALL ·innerDIFWithTwiddlesGeneric(SB)")
	f.RET()

}

func (f *FFAmd64) generateFFTKernelF31(klog2 int, dif bool) {
	if klog2 > 8 || klog2 < 7 {
		panic("not implemented")
	}
	// for now we generate kernels of size 1 << 7 (128) only
	// as we can keep the input and twiddles in registers and avoid round trips with memory.
	// perf note: we could generate a larger kernel, maybe up to 512 and process the "left" part of the FFT
	// fully in registers. may not be clearly worth it since it would only save 3 calls to the assembly
	// innerDIFWithTwiddles ; + the latency to write a to L1 cache.
	n := 1 << klog2
	name := fmt.Sprintf("kerDIFNP_%d_avx512", n)
	if !dif {
		name = fmt.Sprintf("kerDITNP_%d_avx512", n)
	}
	const argSize = 7 * 8
	stackSize := f.StackSize(f.NbWords*2+4, 1, 0)
	registers := f.FnHeader(name, stackSize, argSize, amd64.AX)
	defer f.AssertCleanStack(stackSize, 0)

	// registers & labels we need
	addrA := f.Pop(&registers)
	addrAPlusM := f.Pop(&registers)
	addrTwiddlesRoot := f.Pop(&registers)
	addrTwiddles := f.Pop(&registers)
	innerLen := f.Pop(&registers)

	// AVX512 registers
	a := registers.PopVN(n / 16)
	qd := registers.PopV()
	qInvNeg := registers.PopV()

	// load q and qInvNeg
	f.Comment("prepare constants needed for mul and reduce ops")
	f.MOVD("$const_q", amd64.AX)
	f.VPBROADCASTD(amd64.AX, qd)
	f.MOVD("$const_qInvNeg", amd64.AX)
	f.VPBROADCASTD(amd64.AX, qInvNeg)

	f.Comment("load arguments")
	f.MOVQ("a+0(FP)", addrA)
	f.MOVQ("twiddles+24(FP)", addrTwiddlesRoot)
	f.MOVQ("stage+48(FP)", amd64.AX)
	f.IMULQ("$24", amd64.AX)
	f.ADDQ(amd64.AX, addrTwiddlesRoot, "we want twiddles[stage] as starting point")

	f.MOVQ(uint64(0x0f0f), amd64.AX)
	f.KMOVQ(amd64.AX, amd64.K1)

	f.MOVQ(uint64(0b00110011), amd64.AX)
	f.KMOVQ(amd64.AX, amd64.K2)

	f.MOVQ(uint64(0b0101010101010101), amd64.AX)
	f.KMOVD(amd64.AX, amd64.K3)

	for i := range a {
		// we want to advance by 32bytes to have 8 uint32 element loaded at a time.
		f.VMOVDQU32(addrA.AtD(i*16), a[i], fmt.Sprintf("load a[%d]", i))
	}

	permute4x4, _ := f.DefineFn("permute4x4")
	permute8x8, _ := f.DefineFn("permute8x8")

	if dif {
		vInterleaveIndices := f.generateCoreDIFKernel(n, &registers, addrTwiddlesRoot, a, qd, qInvNeg, true)

		b0 := registers.PopV()

		for i := 0; i < len(a); i += 2 {
			// ok let's say now each pair of vector v0 v1
			// such that
			// v0 = [a0 a2 a4 a6 a8 a10 a12 a14 | b0 b2 b4 b6 b8 b10 b12 b14]
			// v1 = [a1 a3 a5 a7 a9 a11 a13 a15 | b1 b3 b5 b7 b9 b11 b13 b15]

			// we need to repack them; let's do it the naive way for now
			f.VPUNPCKLDQ(a[i+1], a[i], b0)
			f.VPUNPCKHDQ(a[i+1], a[i], a[i+1])
			f.VMOVDQA32(b0, a[i])
			permute4x4(a[i], a[i+1], vInterleaveIndices, b0, amd64.K2)
			permute8x8(a[i], a[i+1], b0, amd64.K1)

			// store the result
			f.VMOVDQU32(a[i], addrA.AtD(i*16))
			f.VMOVDQU32(a[i+1], addrA.AtD((i+1)*16))
		}
	} else {
		f.generateCoreDITKernel(n, &registers, addrTwiddlesRoot, a, qd, qInvNeg, true, klog2-2)
		for i := 0; i < len(a); i += 2 {

			// store the result
			f.VMOVDQU32(a[i], addrA.AtD(i*16))
			f.VMOVDQU32(a[i+1], addrA.AtD((i+1)*16))
		}
	}

	f.RET()

	f.Push(&registers, addrA, addrTwiddles, addrAPlusM, innerLen)

}

func (f *FFAmd64) generateCoreDIFKernel(n int, registers *amd64.Registers, addrTwiddlesRoot amd64.Register, a []amd64.VectorRegister, qd, qInvNeg amd64.VectorRegister, reduceModQ bool) amd64.VectorRegister {

	m := n >> 1
	kk := n

	// get the defines
	butterflyMulD, _ := f.DefineFn("butterfly_mulD")
	permute1x1, _ := f.DefineFn("permute1x1")
	permute2x2, _ := f.DefineFn("permute2x2")
	permute4x4, _ := f.DefineFn("permute4x4")
	permute8x8, _ := f.DefineFn("permute8x8")
	butterflyD1Q, _ := f.DefineFn("butterflyD1Q")
	butterflyD2Q2Q, _ := f.DefineFn("butterflyD2Q2Q")

	t := registers.PopVN(n / 32)
	b0 := registers.PopV()
	b1 := registers.PopV()
	aOdd := registers.PopV()
	bOdd := registers.PopV()
	PL0 := registers.PopV()
	PL1 := registers.PopV()

	addrTwiddles := registers.Pop()

	for m >= 16 {

		f.MOVQ(addrTwiddlesRoot.At(0), addrTwiddles)
		nbTwiddles := m
		for i := 0; i < nbTwiddles/16; i++ {
			f.VMOVDQU32(addrTwiddles.AtD(i*16), t[i])
		}
		am := m / 16
		for offset := 0; offset < kk; offset += n {
			// for offset := 0; offset < 128; offset += n {
			aa := a[offset/16:]
			for i := 0; i < am; i++ {
				butterflyMulD(aa[i], aa[i+am], qd, b0,
					aa[i+am], t[i], aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg, amd64.K3)
			}
		}

		n >>= 1
		m = n >> 1

		// increment addrTwiddlesRoot
		f.ADDQ("$24", addrTwiddlesRoot)
	}

	for i := 4; i < len(t); i++ {
		registers.PushV(t[i])
	}

	addrVInterleaveIndices := f.Pop(registers)
	vInterleaveIndices := registers.PopV()
	f.MOVQ("·vInterleaveIndices+0(SB)", addrVInterleaveIndices)
	f.VMOVDQU64(addrVInterleaveIndices.At(0), vInterleaveIndices)

	// here we have m == 8; we can't process
	// vector of 16 elements anymore without some interleaving

	f.MOVQ(addrTwiddlesRoot.At(0), addrTwiddles)
	f.VMOVDQU32(addrTwiddles.At(0), t[0].Y())
	f.VINSERTI64X4(1, t[0].Y(), t[0], t[0])
	f.MOVQ(addrTwiddlesRoot.At(3), addrTwiddles)
	f.VMOVDQU32(addrTwiddles.At(0), t[1].X())
	f.VINSERTI64X2(1, t[1].X(), t[1], t[1])
	f.VINSERTI64X2(2, t[1].X(), t[1], t[1])
	f.VINSERTI64X2(3, t[1].X(), t[1], t[1])
	f.MOVQ(addrTwiddlesRoot.At(6), addrTwiddles)
	f.VPBROADCASTD(addrTwiddles.AtD(0), t[2])
	f.VPBROADCASTD(addrTwiddles.AtD(1), t[3])
	f.VPBLENDMD(t[2], t[3], t[2], amd64.K3)

	for j := 0; j < 3; j++ {
		for i := 0; i < len(a); i += 2 {
			switch j {
			case 0:
				// m == 8
				permute8x8(a[i], a[i+1], b0, amd64.K1)
			case 1:
				// m == 4
				permute4x4(a[i], a[i+1], vInterleaveIndices, b0, amd64.K2)
			case 2:
				// m == 2
				permute2x2(a[i], a[i+1], b0, amd64.K3)
			}

			// perf note:
			// we can optimize a bit further here by having a
			// mulD version that takes b and bOdd as input;
			// will save couple of VMOVSHDUP since the twiddles are known.

			butterflyMulD(a[i], a[i+1], qd, b0,
				a[i+1], t[j], aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg, amd64.K3)
		}
	}

	for i := 0; i < len(a); i += 2 {
		// m == 1
		permute1x1(a[i], a[i+1], b0, amd64.K3)
		// the last butterfly we reduce everything in [0, q)

		if reduceModQ {
			butterflyD1Q(a[i], a[i+1], qd, b0, b1)
		} else {
			butterflyD2Q2Q(a[i], a[i+1], qd, b0)
		}
	}

	for i := 0; i < 4; i++ {
		registers.PushV(t[i])
	}
	registers.PushV(b0, b1, aOdd, bOdd, PL0, PL1)
	registers.Push(addrVInterleaveIndices, addrTwiddles)

	return vInterleaveIndices
}

func (f *FFAmd64) generateCoreDITKernel(n int, registers *amd64.Registers, addrTwiddlesRoot amd64.Register, a []amd64.VectorRegister, qd, qInvNeg amd64.VectorRegister, reduceModQ bool, startStage int) {
	// perf note: this is less optimized than the DIF one and unrolled a bit naively.
	// not on a hot path at the moment.

	// get the defines
	permute1x1, _ := f.DefineFn("permute1x1")
	permute2x2, _ := f.DefineFn("permute2x2")
	permute4x4, _ := f.DefineFn("permute4x4")
	permute8x8, _ := f.DefineFn("permute8x8")
	butterflyD1Q, _ := f.DefineFn("butterflyD1Q")
	mulD, _ := f.DefineFn("mulD")

	t := registers.PopVN(4)
	b0 := registers.PopV()
	b1 := registers.PopV()
	aOdd := registers.PopV()
	bOdd := registers.PopV()
	PL0 := registers.PopV()
	PL1 := registers.PopV()

	addrTwiddles := registers.Pop()

	addrVInterleaveIndices := f.Pop(registers)
	vInterleaveIndices := registers.PopV()
	f.MOVQ("·vInterleaveIndices+0(SB)", addrVInterleaveIndices)
	f.VMOVDQU64(addrVInterleaveIndices.At(0), vInterleaveIndices)

	// m == 1
	for i := 0; i < len(a); i += 2 {
		// m == 1
		permute1x1(a[i], a[i+1], b0, amd64.K3)
		butterflyD1Q(a[i], a[i+1], qd, b0, b1)
		permute1x1(a[i], a[i+1], b0, amd64.K3)
	}

	// m == 2
	f.MOVQ(startStage, amd64.AX)
	f.IMULQ("$24", amd64.AX)
	f.ADDQ(amd64.AX, addrTwiddlesRoot)
	f.MOVQ(addrTwiddlesRoot.At(0), addrTwiddles)
	f.VPBROADCASTD(addrTwiddles.AtD(0), t[2])
	f.VPBROADCASTD(addrTwiddles.AtD(1), t[3])
	f.VPBLENDMD(t[2], t[3], t[2], amd64.K3)
	for i := 0; i < len(a); i += 2 {
		permute2x2(a[i], a[i+1], b0, amd64.K3)
		mulD(a[i+1], t[2], aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg, amd64.K3)
		butterflyD1Q(a[i], a[i+1], qd, b0, b1)
		permute2x2(a[i], a[i+1], b0, amd64.K3)
	}

	// m == 4
	f.SUBQ("$24", addrTwiddlesRoot)
	f.MOVQ(addrTwiddlesRoot.At(0), addrTwiddles)
	f.VMOVDQU32(addrTwiddles.At(0), t[1].X())
	f.VINSERTI64X2(1, t[1].X(), t[1], t[1])
	f.VINSERTI64X2(2, t[1].X(), t[1], t[1])
	f.VINSERTI64X2(3, t[1].X(), t[1], t[1])
	for i := 0; i < len(a); i += 2 {
		permute4x4(a[i], a[i+1], vInterleaveIndices, b0, amd64.K2)

		mulD(a[i+1], t[1], aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg, amd64.K3)
		butterflyD1Q(a[i], a[i+1], qd, b0, b1)

		permute4x4(a[i], a[i+1], vInterleaveIndices, b0, amd64.K2)
	}

	// m == 8
	f.SUBQ("$24", addrTwiddlesRoot)
	f.MOVQ(addrTwiddlesRoot.At(0), addrTwiddles)
	f.VMOVDQU32(addrTwiddles.At(0), t[0].Y())
	f.VINSERTI64X4(1, t[0].Y(), t[0], t[0])
	for i := 0; i < len(a); i += 2 {
		permute8x8(a[i], a[i+1], b0, amd64.K1)

		mulD(a[i+1], t[0], aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg, amd64.K3)
		butterflyD1Q(a[i], a[i+1], qd, b0, b1)

		permute8x8(a[i], a[i+1], b0, amd64.K1)
	}

	registers.PushV(t...)
	registers.PushV(vInterleaveIndices)

	t = registers.PopVN(n / 32)

	targetM := n >> 1
	kk := n
	m := 16
	_n := 32

	for m <= targetM {
		// increment addrTwiddlesRoot
		f.SUBQ("$24", addrTwiddlesRoot)
		f.MOVQ(addrTwiddlesRoot.At(0), addrTwiddles)
		nbTwiddles := m
		for i := 0; i < nbTwiddles/16; i++ {
			f.VMOVDQU32(addrTwiddles.AtD(i*16), t[i])
		}
		am := m / 16
		f.Comment(fmt.Sprintf("for offset := 0; offset < %d; offset += %d {\n", kk, _n))
		for offset := 0; offset < kk; offset += _n {
			// for offset := 0; offset < 128; offset += n {
			aa := a[offset/16:]
			for i := 0; i < am; i++ {
				f.Comment(fmt.Sprintf("a[i=%d] = %v, a[i+am=%d] = %v\n", i+offset/16, aa[i], i+offset/16+am, aa[i+am]))
				mulD(aa[i+am], t[i], aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg, amd64.K3)
				butterflyD1Q(aa[i], aa[i+am], qd, b0, b1)
			}
		}

		_n <<= 1
		m = _n >> 1

	}

	for i := range t {
		registers.PushV(t[i])
	}
	registers.PushV(b0, b1, aOdd, bOdd, PL0, PL1)
	registers.Push(addrVInterleaveIndices, addrTwiddles)

}

func (f *FFAmd64) generateSISShuffleF31() {
	const argSize = 1 * 3 * 8
	stackSize := f.StackSize(f.NbWords*2+4, 1, 0) // we reserve 512*4bytes and some extra because we want to "align" SP
	registers := f.FnHeader("SISShuffle", stackSize, argSize, amd64.AX, amd64.DI)

	addrA := registers.Pop()
	lenA := registers.Pop()

	b0 := registers.PopV()
	a0 := registers.PopV()
	a1 := registers.PopV()

	f.MOVQ("a+0(FP)", addrA)
	f.MOVQ("a_len+8(FP)", lenA)

	// divide len by 32
	f.SHRQ("$5", lenA)

	lblDone := f.NewLabel("done")
	lblLoop := f.NewLabel("loop")

	permute1x1, _ := f.DefineFn("permute1x1")
	permute2x2, _ := f.DefineFn("permute2x2")
	permute4x4, _ := f.DefineFn("permute4x4")
	permute8x8, _ := f.DefineFn("permute8x8")

	f.MOVQ(uint64(0x0f0f), amd64.AX)
	f.KMOVQ(amd64.AX, amd64.K1)

	f.MOVQ(uint64(0b00110011), amd64.AX)
	f.KMOVQ(amd64.AX, amd64.K2)

	f.MOVQ(uint64(0b0101010101010101), amd64.AX)
	f.KMOVD(amd64.AX, amd64.K3)

	addrVInterleaveIndices := registers.Pop()
	vInterleaveIndices := registers.PopV()
	f.MOVQ("·vInterleaveIndices+0(SB)", addrVInterleaveIndices)
	f.VMOVDQU64(addrVInterleaveIndices.At(0), vInterleaveIndices)

	f.LABEL(lblLoop)

	f.TESTQ(lenA, lenA)
	f.JEQ(lblDone, "n == 0, we are done")

	f.VMOVDQU32(addrA.AtD(0), a0, "load a[i]")
	f.VMOVDQU32(addrA.AtD(16), a1, "load a[i+16]")

	// probably a faster way to do this, but let's do it the naive way for now
	permute8x8(a0, a1, b0, amd64.K1)
	permute4x4(a0, a1, vInterleaveIndices, b0, amd64.K2)
	permute2x2(a0, a1, b0, amd64.K3)
	permute1x1(a0, a1, b0, amd64.K3)

	f.VMOVDQU32(a0, addrA.AtD(0), "store a[i]")
	f.VMOVDQU32(a1, addrA.AtD(16), "store a[i+16]")

	f.ADDQ("$128", addrA)
	f.DECQ(lenA, "decrement n")
	f.JMP(lblLoop)
	f.LABEL(lblDone)
	f.RET()
}

func (f *FFAmd64) generateSISUnhuffleF31() {
	const argSize = 1 * 3 * 8
	stackSize := f.StackSize(f.NbWords*2+4, 1, 0) // we reserve 512*4bytes and some extra because we want to "align" SP
	registers := f.FnHeader("SISUnshuffle", stackSize, argSize, amd64.AX, amd64.DI)

	addrA := registers.Pop()
	lenA := registers.Pop()

	b0 := registers.PopV()
	a0 := registers.PopV()
	a1 := registers.PopV()

	f.MOVQ("a+0(FP)", addrA)
	f.MOVQ("a_len+8(FP)", lenA)

	// divide len by 32
	f.SHRQ("$5", lenA)

	lblDone := f.NewLabel("done")
	lblLoop := f.NewLabel("loop")

	permute4x4, _ := f.DefineFn("permute4x4")
	permute8x8, _ := f.DefineFn("permute8x8")

	f.MOVQ(uint64(0x0f0f), amd64.AX)
	f.KMOVQ(amd64.AX, amd64.K1)

	f.MOVQ(uint64(0b00110011), amd64.AX)
	f.KMOVQ(amd64.AX, amd64.K2)

	f.MOVQ(uint64(0b0101010101010101), amd64.AX)
	f.KMOVD(amd64.AX, amd64.K3)

	addrVInterleaveIndices := registers.Pop()
	vInterleaveIndices := registers.PopV()
	f.MOVQ("·vInterleaveIndices+0(SB)", addrVInterleaveIndices)
	f.VMOVDQU64(addrVInterleaveIndices.At(0), vInterleaveIndices)

	f.LABEL(lblLoop)

	f.TESTQ(lenA, lenA)
	f.JEQ(lblDone, "n == 0, we are done")

	f.VMOVDQU32(addrA.AtD(0), a0, "load a[i]")
	f.VMOVDQU32(addrA.AtD(16), a1, "load a[i+16]")

	// ok let's say now each pair of vector v0 v1
	// such that
	// v0 = [a0 a2 a4 a6 a8 a10 a12 a14 | b0 b2 b4 b6 b8 b10 b12 b14]
	// v1 = [a1 a3 a5 a7 a9 a11 a13 a15 | b1 b3 b5 b7 b9 b11 b13 b15]

	// we need to repack them; let's do it the naive way for now
	f.VPUNPCKLDQ(a1, a0, b0)
	f.VPUNPCKHDQ(a1, a0, a1)
	f.VMOVDQA32(b0, a0)
	permute4x4(a0, a1, vInterleaveIndices, b0, amd64.K2)
	permute8x8(a0, a1, b0, amd64.K1)

	f.VMOVDQU32(a0, addrA.AtD(0), "store a[i]")
	f.VMOVDQU32(a1, addrA.AtD(16), "store a[i+16]")

	f.ADDQ("$128", addrA)
	f.DECQ(lenA, "decrement n")
	f.JMP(lblLoop)
	f.LABEL(lblDone)
	f.RET()

}

func (f *FFAmd64) generateSISToRefactorF31() {
	const argSize = 5 * 3 * 8
	// func SISToRefactor(k256,  cosets, twiddles, rag, res []{{ .FF }}.Element)
	stackSize := f.StackSize(f.NbWords*2+4, 1, 256*4+64) // we reserve 512*4bytes and some extra because we want to "align" SP
	registers := f.FnHeader("SISToRefactor", stackSize, argSize, amd64.AX, amd64.DI)
	// defer f.AssertCleanStack(stackSize, 0)
	sp := amd64.DI
	f.MOVQ(amd64.Register("SP"), sp)

	// if sp is not aligned, we add an offset to it to align it;
	// TODO @gbotrel double check this.
	f.ANDQ("$-64", sp)

	addrK256 := f.Pop(&registers)

	addrK256m := f.Pop(&registers)

	addrCosets := f.Pop(&registers)
	addrTwiddlesRoot := f.Pop(&registers)
	addrTwiddles := f.Pop(&registers)

	addrRag := f.Pop(&registers)
	addrRes := f.Pop(&registers)

	qd := registers.PopV()
	qInvNeg := registers.PopV()

	// load q and qInvNeg
	f.Comment("prepare constants needed for mul and reduce ops")
	f.MOVD("$const_q", amd64.AX)
	f.VPBROADCASTD(amd64.AX, qd)
	f.MOVD("$const_qInvNeg", amd64.AX)
	f.VPBROADCASTD(amd64.AX, qInvNeg)

	f.MOVQ("k256+0(FP)", addrK256)
	f.MOVQ("cosets+24(FP)", addrCosets)
	f.MOVQ("twiddles+48(FP)", addrTwiddlesRoot)
	f.MOVQ("rag+72(FP)", addrRag)
	f.MOVQ("res+96(FP)", addrRes)
	f.MOVQ(addrTwiddlesRoot.At(0), addrTwiddles, "twiddles[0]") // stage 0

	addrCosetsm := registers.Pop()

	f.MOVQ(addrK256, addrK256m)
	f.MOVQ(addrCosets, addrCosetsm)

	f.ADDQ(512, addrK256m)
	f.ADDQ(1024, addrCosetsm)

	// ok let's go step by step during refactor to test...

	// we are going to blend the FFT 512 into that loop;
	// so we want to iterate on the halves of the 512 vector.

	f.MOVQ(uint64(0x0f0f), amd64.AX)
	f.KMOVQ(amd64.AX, amd64.K1)

	f.MOVQ(uint64(0b00110011), amd64.AX)
	f.KMOVQ(amd64.AX, amd64.K2)

	f.MOVQ(uint64(0b0101010101010101), amd64.AX)
	f.KMOVD(amd64.AX, amd64.K3)

	fromMont := f.Define("fromMontgomery", 8, func(args ...any) {
		a := args[0]
		b0 := args[1]
		b1 := args[2]
		PL0 := args[3]
		PL1 := args[4]
		q := args[5]
		qInvNeg := args[6]
		kEvens := args[7]

		f.VPANDDkz(a, a, amd64.K3, b0) // keep low 32 bits
		f.VPSRLQ("$32", a, b1)         // keep high 32 bits

		f.VPMULUDQ(b0, qInvNeg, PL0)
		f.VPMULUDQ(b1, qInvNeg, PL1)

		f.VPMULUDQ(PL0, q, PL0)
		f.VPMULUDQ(PL1, q, PL1)

		f.VPADDQ(b0, PL0, b0)
		f.VPADDQ(b1, PL1, b1)

		f.VMOVSHDUPk(b0, kEvens, b1)

		f.VPSUBD(q, b1, PL1)
		f.VPMINUD(b1, PL1, a)
	})

	mulD, _ := f.DefineFn("mulD")
	butterflyD2Q, _ := f.DefineFn("butterflyD2Q")

	b0 := registers.PopV()
	b1 := registers.PopV()
	aOdd := registers.PopV()
	bOdd := registers.PopV()
	PL0 := registers.PopV()
	PL1 := registers.PopV()

	c1 := registers.PopV()
	am1 := registers.PopV()

	c0 := registers.PopV()
	am0 := registers.PopV()

	// we store the first 256 values directly in register
	a := registers.PopVN(16)

	n := 256 / 16
	for i := 0; i < n/2; i++ {
		a0 := a[i*2]
		a1 := a[i*2+1]
		// load 8 uint32 from k256 into a zmm register (zero extended)
		f.VMOVDQU32(addrK256.AtD(i*16), a0)

		// convert to regular form
		fromMont(a0, b0, b1, PL0, PL1, qd, qInvNeg, amd64.K3)

		// split the limbs
		f.VEXTRACTI64X4(1, a0, a1.Y())
		f.VPMOVZXWD(a0.Y(), a0)
		f.VPMOVZXWD(a1.Y(), a1)

		// mul by cosets
		f.VMOVDQU32(addrCosets.AtD((i*2)*16), c0)
		f.VMOVDQU32(addrCosets.AtD((i*2+1)*16), c1)

		mulD(a0, c0, aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg, amd64.K3)
		mulD(a1, c1, aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg, amd64.K3)

		f.VMOVDQU32(addrK256m.AtD(i*16), am0)
		// convert to regular form
		fromMont(am0, b0, b1, PL0, PL1, qd, qInvNeg, amd64.K3)

		// split the limbs
		f.VEXTRACTI64X4(1, am0, am1.Y())
		f.VPMOVZXWD(am0.Y(), am0)
		f.VPMOVZXWD(am1.Y(), am1)

		// mul by cosets
		f.VMOVDQU32(addrCosetsm.AtD((i*2)*16), c0)
		f.VMOVDQU32(addrCosetsm.AtD((i*2+1)*16), c1)

		mulD(am0, c0, aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg, amd64.K3)
		mulD(am1, c1, aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg, amd64.K3)

		// now we can do the first layer of the fft easily
		butterflyD2Q(a0, am0, qd, b0)
		butterflyD2Q(a1, am1, qd, b0)

		// scale am0 and am1 by twiddles
		f.VMOVDQU32(addrTwiddles.AtD((i*2)*16), c0)
		f.VMOVDQU32(addrTwiddles.AtD((i*2+1)*16), c1)

		mulD(am0, c0, aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg, amd64.K3)
		mulD(am1, c1, aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg, amd64.K3)

		// f.VMOVDQU32(a0, addrK512.AtD((i*2)*16))
		// f.VMOVDQU32(a1, addrK512.AtD((i*2+1)*16))
		f.VMOVDQA32(am0, sp.AtD((i*2)*16))
		f.VMOVDQA32(am1, sp.AtD((i*2+1)*16))
	}

	registers.PushV(b0, b1, aOdd, bOdd, PL0, PL1, c1, am1, c0, am0)

	// next stage of twiddles
	f.ADDQ("$24", addrTwiddlesRoot)
	f.MOVQ(addrTwiddlesRoot, addrTwiddles) // backup addr twiddles for the other half

	lblDone := f.NewLabel("done")
	lblFFT256 := f.NewLabel("fft256")
	o := registers.Pop()
	f.MOVQ("$2", o)
	f.LABEL(lblFFT256)
	// we do the fft on the first half.
	vInterleaveIndices := f.generateCoreDIFKernel(256, &registers, addrTwiddlesRoot, a, qd, qInvNeg, false)

	b0 = registers.PopV()
	b1 = registers.PopV()
	aOdd = registers.PopV()
	bOdd = registers.PopV()
	PL0 = registers.PopV()
	PL1 = registers.PopV()
	c0 = registers.PopV()
	c1 = registers.PopV()

	// permute4x4, _ := f.DefineFn("permute4x4")
	// permute8x8, _ := f.DefineFn("permute8x8")
	// for i := 0; i < len(a); i += 2 {
	// 	// ok let's say now each pair of vector v0 v1
	// 	// such that
	// 	// v0 = [a0 a2 a4 a6 a8 a10 a12 a14 | b0 b2 b4 b6 b8 b10 b12 b14]
	// 	// v1 = [a1 a3 a5 a7 a9 a11 a13 a15 | b1 b3 b5 b7 b9 b11 b13 b15]

	// 	// we need to repack them; let's do it the naive way for now
	// 	f.VPUNPCKLDQ(a[i+1], a[i], b0)
	// 	f.VPUNPCKHDQ(a[i+1], a[i], a[i+1])
	// 	f.VMOVDQA32(b0, a[i])
	// 	permute4x4(a[i], a[i+1], vInterleaveIndices, b0, amd64.K2)
	// 	permute8x8(a[i], a[i+1], b0, amd64.K1)
	// }
	for i := 0; i < len(a); i += 2 {
		// mul by rag
		f.VMOVDQU32(addrRag.AtD(i*16), c0)
		f.VMOVDQU32(addrRag.AtD((i+1)*16), c1)

		mulD(a[i], c0, aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg, amd64.K3)
		mulD(a[i+1], c1, aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg, amd64.K3)

	}
	for i := 0; i < len(a); i += 2 {
		// accumulate in res
		f.VMOVDQU32(addrRes.AtD(i*16), c0)
		f.VMOVDQU32(addrRes.AtD((i+1)*16), c1)

		f.VPADDD(c0, a[i], a[i])
		f.VPSUBD(qd, a[i], PL0)
		f.VPMINUD(PL0, a[i], a[i])
		f.VPADDD(c1, a[i+1], a[i+1])
		f.VPSUBD(qd, a[i+1], PL1)
		f.VPMINUD(PL1, a[i+1], a[i+1])

		// store the result
		f.VMOVDQU32(a[i], addrRes.AtD(i*16))
		f.VMOVDQU32(a[i+1], addrRes.AtD((i+1)*16))
	}

	f.DECQ(o)
	f.TESTQ(o, o)
	f.JEQ(lblDone)

	// here we are not done, so we setup the other half of the fft.
	f.MOVQ(addrTwiddles, addrTwiddlesRoot)
	for i := range a {
		f.VMOVDQU32(sp.AtD(i*16), a[i])
	}
	f.ADDQ(1024, addrRag)
	f.ADDQ(1024, addrRes)

	f.JMP(lblFFT256)

	registers.PushV(vInterleaveIndices)

	// for i := range a {
	// 	f.VMOVDQU32(a[i], addrK512.AtD(i*16))
	// }

	f.LABEL(lblDone)

	f.RET()
}
