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
	registers := f.FnHeader("addVec", stackSize, argSize, amd64.AX)
	defer f.AssertCleanStack(stackSize, 0)

	// registers & labels we need
	addrA := registers.Pop()
	addrB := registers.Pop()
	addrRes := registers.Pop()
	len := registers.Pop()

	// AVX512 registers
	a := registers.PopV()
	b := registers.PopV()
	t := registers.PopV()
	q := registers.PopV()

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
	registers := f.FnHeader("subVec", stackSize, argSize, amd64.AX)
	defer f.AssertCleanStack(stackSize, 0)

	// registers & labels we need
	addrA := registers.Pop()
	addrB := registers.Pop()
	addrRes := registers.Pop()
	len := registers.Pop()

	// AVX512 registers
	a := registers.PopV()
	b := registers.PopV()
	t := registers.PopV()
	q := registers.PopV()

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
	addrA := registers.Pop()
	addrT := registers.Pop()
	len := registers.Pop()

	// AVX512 registers
	a1 := registers.PopV()
	a2 := registers.PopV()
	acc1 := registers.PopV()
	acc2 := registers.PopV()

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
	addrA := registers.Pop()
	addrB := registers.Pop()
	addrRes := registers.Pop()
	len := registers.Pop()

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
	// f.VMOVSHDUP(a, aOdd)
	// f.VMOVSHDUP(b, bOdd)

	f.VPSRLQ("$32", a, aOdd) // keep high 32 bits
	f.VPSRLQ("$32", b, bOdd) // keep high 32 bits

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
	// TODO @gbotrel update to process 16 by 16 like mul.
	f.Comment("scalarMulVec(res, a, b *Element, n uint64) res[0...n] = a[0...n] * b")
	f.Comment("n is the number of blocks of 8 elements to process")
	const argSize = 4 * 8
	stackSize := f.StackSize(f.NbWords*2+4, 0, 0)
	registers := f.FnHeader("scalarMulVec", stackSize, argSize)
	defer f.AssertCleanStack(stackSize, 0)

	// registers & labels we need
	addrA := registers.Pop()
	addrB := registers.Pop()
	addrRes := registers.Pop()
	len := registers.Pop()

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
	// TODO @gbotrel update to process 16 by 16 like mul.
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
	addrA := registers.Pop()
	addrB := registers.Pop()
	addrT := registers.Pop()
	len := registers.Pop()

	// AVX512 registers
	a := registers.PopV()
	b := registers.PopV()
	acc := registers.PopV()
	q := registers.PopV()
	qInvNeg := registers.PopV()
	PL := registers.PopV()
	LSW := registers.PopV()
	P := registers.PopV()

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

	// computes a = a + b and b = a - b,
	// leaves a in [0, q)
	// leaves b in [0,q)
	_ = f.Define("butterflyD1Q", 5, func(args ...any) {
		x := args[0]
		y := args[1]
		qd := args[2]
		b0 := args[3]
		b1 := args[4]
		f.VPADDD(x, y, b0)
		f.VPSUBD(y, x, y)
		f.VPSUBD(qd, b0, x)
		f.VPMINUD(b0, x, x)
		f.VPADDD(qd, y, b1)
		f.VPMINUD(b1, y, y)
	})

	// computes a = a + b and b = a - b,
	// leaves a in [0, q)
	// leaves b in [0,2q)
	butterflyD2Q := f.Define("butterflyD2Q", 4, func(args ...any) {
		x := args[0]
		y := args[1]
		qd := args[2]
		b0 := args[3]
		f.VPADDD(x, y, b0)
		f.VPSUBD(y, x, y)
		f.VPSUBD(qd, b0, x)
		f.VPMINUD(b0, x, x)
		f.VPADDD(qd, y, y)
	})

	// computes a = a + b and b = a - b,
	// leaves a in [0,2q)
	// leaves b in [0,2q)
	_ = f.Define("butterflyD2Q2Q", 4, func(args ...any) {
		x := args[0]
		y := args[1]
		qd := args[2]
		b0 := args[3]
		f.VPSUBD(y, x, b0)
		f.VPADDD(x, y, x)
		f.VPADDD(qd, b0, y)
	})

	// computes a = a * b mod q
	// a and b can be in [0, 2q)
	mulD := f.Define("mulD", 10, func(args ...any) {
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

		f.VPSRLQ("$32", a, aOdd) // keep high 32 bits
		f.VPSRLQ("$32", b, bOdd) // keep high 32 bits
		f.VPMULUDQ(a, b, b0)
		f.VPMULUDQ(aOdd, bOdd, b1)
		f.VPMULUDQ(b0, qInvNeg, PL0)
		f.VPMULUDQ(b1, qInvNeg, PL1)

		f.VPMULUDQ(PL0, q, PL0)
		f.VPADDQ(b0, PL0, b0)

		f.VPMULUDQ(PL1, q, PL1)
		f.VPADDQ(b1, PL1, b1)

		f.VMOVSHDUPk(b0, amd64.K3, b1)

		f.VPSUBD(q, b1, PL1)
		f.VPMINUD(b1, PL1, a)
	})

	// goes from
	// in0 = [a0 a1 a2 a3 a4 a5 a6 a7 a8 a9 a10 a11 a12 a13 a14 a15]
	// in1 = [b0 b1 b2 b3 b4 b5 b6 b7 b8 b9 b10 b11 b12 b13 b14 b15]
	// to
	// in0 = [a0 a1 a2 a3 a4 a5 a6 a7 b0 b1 b2 b3 b4 b5 b6 b7]
	// in1 = [a8 a9 a10 a11 a12 a13 a14 a15 b8 b9 b10 b11 b12 b13 b14 b15]
	_ = f.Define("permute8x8", 3, func(args ...any) {
		x := args[0]
		y := args[1]
		b0 := args[2]
		f.VSHUFI64X2(uint64(0b01_00_11_10), y, x, b0)
		f.VPBLENDMQ(x, b0, x, amd64.K1)
		f.VPBLENDMQ(b0, y, y, amd64.K1)
	})

	// goes from
	// in0 = [a0 a1 a2 a3 a4 a5 a6 a7 a8 a9 a10 a11 a12 a13 a14 a15]
	// in1 = [b0 b1 b2 b3 b4 b5 b6 b7 b8 b9 b10 b11 b12 b13 b14 b15]
	// to
	// in0 = [a0 a1 a2 a3 b0 b1 b2 b3 a8 a9 a10 a11 b8 b9 b10 b11]
	// in1 = [a4 a5 a6 a7 b4 b5 b6 b7 a12 a13 a14 a15 b12 b13 b14 b15]
	_ = f.Define("permute4x4", 4, func(args ...any) {
		x := args[0]
		y := args[1]
		vInterleaveIndices := args[2]
		tmp := args[3]
		f.VMOVDQA64(vInterleaveIndices, tmp)
		f.VPERMI2Q(y, x, tmp)
		f.VPBLENDMQ(x, tmp, x, amd64.K2)
		f.VPBLENDMQ(tmp, y, y, amd64.K2)
	})

	// goes from
	// in0 = [a0 a1 a2 a3 a4 a5 a6 a7 a8 a9 a10 a11 a12 a13 a14 a15]
	// in1 = [b0 b1 b2 b3 b4 b5 b6 b7 b8 b9 b10 b11 b12 b13 b14 b15]
	// to
	// in0 = [a0 a1 b0 b1 a4 a5 b4 b5 a8 a9 b8 b9 a12 a13 b12 b13]
	// in1 = [a2 a3 b2 b3 a6 a7 b6 b7 a10 a11 b10 b11 a14 a15 b14 b15]
	_ = f.Define("permute2x2", 3, func(args ...any) {
		x := args[0]
		y := args[1]
		b0 := args[2]

		f.VSHUFPD(0b01010101, y, x, b0)
		f.VPBLENDMQ(x, b0, x, amd64.K3)
		f.VPBLENDMQ(b0, y, y, amd64.K3)
	})

	// goes from
	// in0 = [a0 a1 a2 a3 a4 a5 a6 a7 a8 a9 a10 a11 a12 a13 a14 a15]
	// in1 = [b0 b1 b2 b3 b4 b5 b6 b7 b8 b9 b10 b11 b12 b13 b14 b15]
	// to
	// in0 = [a0 b0 a2 b2 a4 b4 a6 b6 a8 b8 a10 b10 a12 b12 a14 b14]
	// in1 = [a1 b1 a3 b3 a5 b5 a7 b7 a9 b9 a11 b11 a13 b13 a15 b15]
	_ = f.Define("permute1x1", 3, func(args ...any) {
		x := args[0]
		y := args[1]
		b0 := args[2]

		f.VPSHRDQ("$32", y, x, b0)
		f.VPBLENDMD(x, b0, x, amd64.K3)
		f.VPBLENDMD(b0, y, y, amd64.K3)
	})

	_ = f.Define("load_q", 2, func(args ...any) {
		q := args[0]
		qInv := args[1]
		f.MOVD("$const_q", amd64.AX)
		f.VPBROADCASTD(amd64.AX, q)
		f.MOVD("$const_qInvNeg", amd64.AX)
		f.VPBROADCASTD(amd64.AX, qInv)
	})

	_ = f.Define("load_masks", 0, func(_ ...any) {
		f.MOVQ(uint64(0b0000_1111_0000_1111), amd64.AX)
		f.KMOVQ(amd64.AX, amd64.K1)

		f.MOVQ(uint64(0b00_11_00_11), amd64.AX)
		f.KMOVQ(amd64.AX, amd64.K2)

		f.MOVQ(uint64(0b01_01_01_01_01_01_01_01), amd64.AX)
		f.KMOVD(amd64.AX, amd64.K3)
	})

	_ = f.Define("butterfly_mulD", 10+4, func(args ...any) {
		butterflyD2Q(args[0], args[1], args[2], args[3])
		mulD(args[4:]...)
	})
}

func (_f *FFAmd64) generateFFTInnerDITF31() {
	f := &fftHelper{_f}

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

	addrA := registers.Pop()
	addrAPlusM := registers.Pop()
	addrTwiddles := registers.Pop()
	m := registers.Pop()
	len := registers.Pop()

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

	f.loadQ(qd, qInvNeg)
	f.loadMasks()

	f.Comment("load arguments")
	f.MOVQ("a+0(FP)", addrA)
	f.MOVQ("twiddles+24(FP)", addrTwiddles)
	f.MOVQ("end+56(FP)", len)
	f.MOVQ("m+64(FP)", m)

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

	f.mulD(am, t0, aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg)
	f.butterflyD1Q(a, am, qd, b0, b1)

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

func (_f *FFAmd64) generateFFTInnerDIFF31() {
	f := &fftHelper{_f}
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

	addrA := registers.Pop()
	addrAPlusM := registers.Pop()
	addrTwiddles := registers.Pop()
	m := registers.Pop()
	len := registers.Pop()

	a := registers.PopV()
	am := registers.PopV()
	qd := registers.PopV()
	b0 := registers.PopV()
	qInvNeg := registers.PopV()
	t0 := registers.PopV()

	aOdd := registers.PopV()
	bOdd := registers.PopV()
	b1 := registers.PopV()
	PL0 := registers.PopV()
	PL1 := registers.PopV()

	f.loadQ(qd, qInvNeg)
	f.loadMasks()

	f.Comment("load arguments")
	f.MOVQ("a+0(FP)", addrA)
	f.MOVQ("twiddles+24(FP)", addrTwiddles)
	f.MOVQ("end+56(FP)", len)
	f.MOVQ("m+64(FP)", m)

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

	lblDone := f.NewLabel("done")
	lblLoop := f.NewLabel("loop")

	f.LABEL(lblLoop)

	f.TESTQ(len, len)
	f.JEQ(lblDone, "n == 0, we are done")

	f.VMOVDQU32(addrA.At(0), a, "load a[i]")
	f.VMOVDQU32(addrAPlusM.At(0), am, "load a[i+m]")

	f.VMOVDQU32(addrTwiddles.At(0), t0)
	f.butterfly_mulD(a, am, qd, b0,
		am, t0, aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg)

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

func (_f *FFAmd64) generateFFTKernelF31(klog2 int, dif bool) {
	f := &fftHelper{_f}
	if klog2 > 8 || klog2 < 7 {
		panic("not implemented")
	}
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
	addrA := registers.Pop()
	addrAPlusM := registers.Pop()
	addrTwiddlesRoot := registers.Pop()
	addrTwiddles := registers.Pop()
	innerLen := registers.Pop()

	// AVX512 registers
	a := registers.PopVN(n / 16)
	qd := registers.PopV()
	qInvNeg := registers.PopV()

	f.loadQ(qd, qInvNeg)
	f.loadMasks()

	f.Comment("load arguments")
	f.MOVQ("a+0(FP)", addrA)
	f.MOVQ("twiddles+24(FP)", addrTwiddlesRoot)
	f.MOVQ("stage+48(FP)", amd64.AX)
	f.IMULQ("$24", amd64.AX)
	f.ADDQ(amd64.AX, addrTwiddlesRoot, "we want twiddles[stage] as starting point")

	for i := range a {
		// we want to advance by 32bytes to have 8 uint32 element loaded at a time.
		f.VMOVDQU32(addrA.AtD(i*16), a[i], fmt.Sprintf("load a[%d]", i))
	}

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
			f.permute4x4(a[i], a[i+1], vInterleaveIndices, b0)
			f.permute8x8(a[i], a[i+1], b0)

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

func (f *fftHelper) generateCoreDIFKernel(n int, registers *amd64.Registers, addrTwiddlesRoot amd64.Register, a []amd64.VectorRegister, qd, qInvNeg amd64.VectorRegister, reduceModQ bool) amd64.VectorRegister {

	m := n >> 1
	kk := n

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
			aa := a[offset/16:]
			for i := 0; i < am; i++ {
				f.butterfly_mulD(aa[i], aa[i+am], qd, b0,
					aa[i+am], t[i], aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg)
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
				f.permute8x8(a[i], a[i+1], b0)
			case 1:
				// m == 4
				f.permute4x4(a[i], a[i+1], vInterleaveIndices, b0)
			case 2:
				// m == 2
				f.permute2x2(a[i], a[i+1], b0)
			}

			// perf note:
			// we can optimize a bit further here by having a
			// mulD version that takes b and bOdd as input;
			// will save couple of high bit extraction

			f.butterfly_mulD(a[i], a[i+1], qd, b0,
				a[i+1], t[j], aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg)
		}
	}

	for i := 0; i < len(a); i += 2 {
		// m == 1
		f.permute1x1(a[i], a[i+1], b0)
		if reduceModQ {
			// the last butterfly we reduce everything in [0, q)
			f.butterflyD1Q(a[i], a[i+1], qd, b0, b1)
		} else {
			// the last butterfly we reduce everything in [0, 2q)
			// this is useful in SIS when we scale the result afterwards;
			// the last mul will reduce in [0, q)
			f.butterflyD2Q2Q(a[i], a[i+1], qd, b0)
		}
	}

	for i := 0; i < 4; i++ {
		registers.PushV(t[i])
	}
	registers.PushV(b0, b1, aOdd, bOdd, PL0, PL1)
	registers.Push(addrVInterleaveIndices, addrTwiddles)

	return vInterleaveIndices
}

func (f *fftHelper) generateCoreDITKernel(n int, registers *amd64.Registers, addrTwiddlesRoot amd64.Register, a []amd64.VectorRegister, qd, qInvNeg amd64.VectorRegister, reduceModQ bool, startStage int) {
	// perf note: this is less optimized than the DIF one and unrolled a bit naively.
	// not on a hot path at the moment.

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
		f.permute1x1(a[i], a[i+1], b0)
		f.butterflyD1Q(a[i], a[i+1], qd, b0, b1)
		f.permute1x1(a[i], a[i+1], b0)
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
		f.permute2x2(a[i], a[i+1], b0)
		f.mulD(a[i+1], t[2], aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg)
		f.butterflyD1Q(a[i], a[i+1], qd, b0, b1)
		f.permute2x2(a[i], a[i+1], b0)
	}

	// m == 4
	f.SUBQ("$24", addrTwiddlesRoot)
	f.MOVQ(addrTwiddlesRoot.At(0), addrTwiddles)
	f.VMOVDQU32(addrTwiddles.At(0), t[1].X())
	f.VINSERTI64X2(1, t[1].X(), t[1], t[1])
	f.VINSERTI64X2(2, t[1].X(), t[1], t[1])
	f.VINSERTI64X2(3, t[1].X(), t[1], t[1])
	for i := 0; i < len(a); i += 2 {
		f.permute4x4(a[i], a[i+1], vInterleaveIndices, b0)
		f.mulD(a[i+1], t[1], aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg)
		f.butterflyD1Q(a[i], a[i+1], qd, b0, b1)
		f.permute4x4(a[i], a[i+1], vInterleaveIndices, b0)
	}

	// m == 8
	f.SUBQ("$24", addrTwiddlesRoot)
	f.MOVQ(addrTwiddlesRoot.At(0), addrTwiddles)
	f.VMOVDQU32(addrTwiddles.At(0), t[0].Y())
	f.VINSERTI64X4(1, t[0].Y(), t[0], t[0])
	for i := 0; i < len(a); i += 2 {
		f.permute8x8(a[i], a[i+1], b0)
		f.mulD(a[i+1], t[0], aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg)
		f.butterflyD1Q(a[i], a[i+1], qd, b0, b1)
		f.permute8x8(a[i], a[i+1], b0)
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
		for offset := 0; offset < kk; offset += _n {
			// for offset := 0; offset < 128; offset += n {
			aa := a[offset/16:]
			for i := 0; i < am; i++ {
				f.mulD(aa[i+am], t[i], aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg)
				f.butterflyD1Q(aa[i], aa[i+am], qd, b0, b1)
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

func (_f *FFAmd64) generateSISShuffleF31() {
	f := &fftHelper{_f}
	const argSize = 1 * 3 * 8
	stackSize := f.StackSize(f.NbWords*2+4, 1, 0)
	registers := f.FnHeader("sisShuffle_avx512", stackSize, argSize, amd64.AX, amd64.DI)

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

	f.loadMasks()

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
	// it's not on the hot path.
	// the idea here is to "shuffle" a vector the way the avx512 fft DIF would.
	f.permute8x8(a0, a1, b0)
	f.permute4x4(a0, a1, vInterleaveIndices, b0)
	f.permute2x2(a0, a1, b0)
	f.permute1x1(a0, a1, b0)

	f.VMOVDQU32(a0, addrA.AtD(0), "store a[i]")
	f.VMOVDQU32(a1, addrA.AtD(16), "store a[i+16]")

	f.ADDQ("$128", addrA)
	f.DECQ(lenA, "decrement n")
	f.JMP(lblLoop)
	f.LABEL(lblDone)
	f.RET()
}

func (_f *FFAmd64) generateSISUnhuffleF31() {
	f := &fftHelper{_f}
	const argSize = 1 * 3 * 8
	stackSize := f.StackSize(f.NbWords*2+4, 1, 0)
	registers := f.FnHeader("sisUnshuffle_avx512", stackSize, argSize, amd64.AX, amd64.DI)

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

	f.loadMasks()

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
	// this is what the FFT avx512 DIF would do as last step;
	// but for SIS we can skip it and do it only once at the end, very useful
	// when hashing lots of vectors.
	f.VPUNPCKLDQ(a1, a0, b0)
	f.VPUNPCKHDQ(a1, a0, a1)
	f.VMOVDQA32(b0, a0)
	f.permute4x4(a0, a1, vInterleaveIndices, b0)
	f.permute8x8(a0, a1, b0)

	f.VMOVDQU32(a0, addrA.AtD(0), "store a[i]")
	f.VMOVDQU32(a1, addrA.AtD(16), "store a[i+16]")

	f.ADDQ("$128", addrA)
	f.DECQ(lenA, "decrement n")
	f.JMP(lblLoop)
	f.LABEL(lblDone)
	f.RET()

}

func (_f *FFAmd64) generateSIS512_16F31() {
	f := &fftHelper{_f}
	const argSize = 5 * 3 * 8
	stackSize := f.StackSize(f.NbWords*2+4, 1, 256*4)
	registers := f.FnHeader("sis512_16_avx512", stackSize, argSize, amd64.AX)
	sp := amd64.Register("SP")

	addrK256 := registers.Pop()
	addrK256m := registers.Pop()
	addrCosets := registers.Pop()
	addrCosetsm := registers.Pop()
	addrTwiddlesRoot := registers.Pop()
	addrTwiddles := registers.Pop()
	addrRag := registers.Pop()
	addrRes := registers.Pop()

	qd := registers.PopV()
	qInvNeg := registers.PopV()

	// load q and qInvNeg
	f.loadQ(qd, qInvNeg)
	f.loadMasks()

	f.MOVQ("k256+0(FP)", addrK256)
	f.MOVQ("cosets+24(FP)", addrCosets)
	f.MOVQ("twiddles+48(FP)", addrTwiddlesRoot)
	f.MOVQ("rag+72(FP)", addrRag)
	f.MOVQ("res+96(FP)", addrRes)
	f.MOVQ(addrTwiddlesRoot.At(0), addrTwiddles, "twiddles[0]") // stage 0

	f.MOVQ(addrK256, addrK256m)
	f.MOVQ(addrCosets, addrCosetsm)

	f.ADDQ(512, addrK256m)
	f.ADDQ(1024, addrCosetsm)

	// this batch convert vectors from montgomery form to regular form
	// it is a mulD(a, 1)
	fromMont := f.Define("fromMontgomery", 7, func(args ...any) {
		a := args[0]
		b0 := args[1]
		b1 := args[2]
		PL0 := args[3]
		PL1 := args[4]
		q := args[5]
		qInvNeg := args[6]

		f.VPSRLQ("$32", a, b1) // keep high 32 bits

		f.VPMULUDQ(a, qInvNeg, PL0)
		f.VPMULUDQ(b1, qInvNeg, PL1)

		f.VPMULUDQ(PL0, q, PL0)
		f.VPMULUDQ(PL1, q, PL1)

		f.VPANDDkz(a, a, amd64.K3, b0) // keep low 32 bits
		f.VPADDQ(b0, PL0, b0)
		f.VPADDQ(b1, PL1, b1)

		f.VMOVSHDUPk(b0, amd64.K3, b1)

		f.VPSUBD(q, b1, PL1)
		f.VPMINUD(b1, PL1, a)
	})

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

	// essentially, the algorithm is as follows:
	// 1. load 256 values from k256 (uint32);
	// 		note that we have 2 pointers; 1 at the beginning "x" and 1 at the middle "xm"
	// 		this enables us to do the first stage of the FFT directly, and save in registers the first half
	// 		for the next stage of the FFT. the second half is stored on the stack.
	// 2. convert to regular form
	// 3. split the limbs into 512 values (uint32)
	// 4. multiply by cosets
	// 5. perform the FFT first stage (512)
	//		that is butterfly and multiply by twiddles
	// 6. at the end of this first unrolled loop, we have the first 256 values in registers
	// and the second 256 values on the stack.
	// we still need to do the FFT on these 2 halves, then multiply by rag, and accumulate in res.
	//
	// The result is "shuffled", and before calling the FFT inverse, caller need to call sisUnshuffle

	n := 256 / 16
	for i := 0; i < n/2; i++ {
		a0 := a[i*2]
		a1 := a[i*2+1]
		// load 8 uint32 from k256 into a zmm register (zero extended)
		f.VMOVDQU32(addrK256.AtD(i*16), a0)

		// convert to regular form
		fromMont(a0, b0, b1, PL0, PL1, qd, qInvNeg)

		// split the limbs
		f.VEXTRACTI64X4(1, a0, a1.Y())
		f.VPMOVZXWD(a0.Y(), a0)
		f.VPMOVZXWD(a1.Y(), a1)

		// mul by cosets
		f.VMOVDQU32(addrCosets.AtD((i*2)*16), c0)
		f.VMOVDQU32(addrCosets.AtD((i*2+1)*16), c1)

		f.mulD(a0, c0, aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg)
		f.mulD(a1, c1, aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg)

		f.VMOVDQU32(addrK256m.AtD(i*16), am0)
		fromMont(am0, b0, b1, PL0, PL1, qd, qInvNeg)

		f.VEXTRACTI64X4(1, am0, am1.Y())
		f.VPMOVZXWD(am0.Y(), am0)
		f.VPMOVZXWD(am1.Y(), am1)

		// mul by cosets
		f.VMOVDQU32(addrCosetsm.AtD((i*2)*16), c0)
		f.VMOVDQU32(addrCosetsm.AtD((i*2+1)*16), c1)

		f.mulD(am0, c0, aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg)
		f.mulD(am1, c1, aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg)

		// now we can do the first layer of the fft easily
		f.butterflyD2Q(a0, am0, qd, b0)
		f.butterflyD2Q(a1, am1, qd, b0)

		// scale am0 and am1 by twiddles
		f.VMOVDQU32(addrTwiddles.AtD((i*2)*16), c0)
		f.VMOVDQU32(addrTwiddles.AtD((i*2+1)*16), c1)

		f.mulD(am0, c0, aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg)
		f.mulD(am1, c1, aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg)

		f.VMOVDQU32(am0, sp.AtD((i*2)*16))
		f.VMOVDQU32(am1, sp.AtD((i*2+1)*16))
	}

	registers.PushV(b0, b1, aOdd, bOdd, PL0, PL1, c1, am1, c0, am0)

	// next stage of twiddles
	f.ADDQ("$24", addrTwiddlesRoot)
	f.MOVQ(addrTwiddlesRoot, addrTwiddles) // backup addr twiddles for the other half

	// we unroll the FFT 256 here
	// and call it twice by modifying the "pointers" to the data and the twiddles.
	lblDone := f.NewLabel("done")
	lblFFT256 := f.NewLabel("fft256")
	o := registers.Pop()
	f.MOVQ("$2", o)
	f.LABEL(lblFFT256)
	// we do the fft on the first half.
	_ = f.generateCoreDIFKernel(256, &registers, addrTwiddlesRoot, a, qd, qInvNeg, false)

	b0 = registers.PopV()
	b1 = registers.PopV()
	aOdd = registers.PopV()
	bOdd = registers.PopV()
	PL0 = registers.PopV()
	PL1 = registers.PopV()
	c0 = registers.PopV()
	c1 = registers.PopV()

	// we can now mul by RAG the result of the FFT256
	for i := 0; i < len(a); i += 2 {
		f.VMOVDQU32(addrRag.AtD(i*16), c0)
		f.VMOVDQU32(addrRag.AtD((i+1)*16), c1)

		f.mulD(a[i], c0, aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg)
		f.mulD(a[i+1], c1, aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg)
	}

	// accumulate in res
	for i := 0; i < len(a); i += 2 {
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

	// decrement the count of FFT256 to do
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

	f.LABEL(lblDone)

	f.RET()
}

type fftHelper struct {
	*FFAmd64
}

func (f *fftHelper) loadQ(q, qInv amd64.VectorRegister) {
	loadQ, _ := f.DefineFn("load_q")
	loadQ(q, qInv)
}

func (f *fftHelper) loadMasks() {
	loadMasks, _ := f.DefineFn("load_masks")
	loadMasks()
}

func (f *fftHelper) butterflyD1Q(args ...any) {
	butterflyD1Q, _ := f.DefineFn("butterflyD1Q")
	butterflyD1Q(args...)
}

func (f *fftHelper) butterflyD2Q(args ...any) {
	butterflyD2Q, _ := f.DefineFn("butterflyD2Q")
	butterflyD2Q(args...)
}

func (f *fftHelper) butterflyD2Q2Q(args ...any) {
	butterflyD2Q2Q, _ := f.DefineFn("butterflyD2Q2Q")
	butterflyD2Q2Q(args...)
}

func (f *fftHelper) mulD(args ...any) {
	mulD, _ := f.DefineFn("mulD")
	mulD(args...)
}

func (f *fftHelper) butterfly_mulD(args ...any) {
	butterfly_mulD, _ := f.DefineFn("butterfly_mulD")
	butterfly_mulD(args...)
}

func (f *fftHelper) permute1x1(args ...any) {
	permute1x1, _ := f.DefineFn("permute1x1")
	permute1x1(args...)
}

func (f *fftHelper) permute2x2(args ...any) {
	permute2x2, _ := f.DefineFn("permute2x2")
	permute2x2(args...)
}

func (f *fftHelper) permute4x4(args ...any) {
	permute4x4, _ := f.DefineFn("permute4x4")
	permute4x4(args...)
}

func (f *fftHelper) permute8x8(args ...any) {
	permute8x8, _ := f.DefineFn("permute8x8")
	permute8x8(args...)
}
