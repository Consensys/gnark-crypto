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
	f.WriteLn("MOVD $const_q, AX")
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
	f.WriteLn("MOVD $const_q, AX")
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
	registers := f.FnHeader("mulVec", stackSize, argSize)
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
	f.WriteLn("MOVD $const_q, AX")
	f.VPBROADCASTQ(amd64.AX, q)
	f.WriteLn("MOVD $const_qInvNeg, AX")
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

	f.LABEL(loop)

	f.TESTQ(len, len)
	f.JEQ(done, "n == 0, we are done")

	// a = a * b
	f.VPMOVZXDQ(addrA.At(0), a)
	f.VPMOVZXDQ(addrB.At(0), b)
	f.VPMULUDQ(a, b, P, "P = a * b")
	f.VPANDQ(LSW, P, PL, "m = uint32(P)")
	f.VPMULUDQ(PL, qInvNeg, PL, "m = m * qInvNeg")
	f.VPANDQ(LSW, PL, PL, "m = uint32(m)")
	f.VPMULUDQ(PL, q, PL, "m = m * q")
	f.VPADDQ(P, PL, P, "P = P + m")
	f.VPSRLQ("$32", P, P, "P = P >> 32")

	f.VPSUBD(q, P, PL, "PL = P - q")
	f.VPMINUD(P, PL, P, "P = min(P, PL)")

	// move P to res
	f.VPMOVQD(P, addrRes.At(0), "res = P")

	f.Comment("increment pointers to visit next element")
	f.ADDQ("$32", addrA)
	f.ADDQ("$32", addrB)
	f.ADDQ("$32", addrRes)
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
	f.WriteLn("MOVD $const_q, AX")
	f.VPBROADCASTQ(amd64.AX, q)
	f.WriteLn("MOVD $const_qInvNeg, AX")
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

	f.VPSUBD(q, P, PL, "PL = P - q")
	f.VPMINUD(P, PL, P, "P = min(P, PL)")

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

	f.WriteLn("MOVD $const_q, AX")
	f.VPBROADCASTQ(amd64.AX, q)
	f.WriteLn("MOVD $const_qInvNeg, AX")
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
	_ = f.Define("butterflyD2Q", 4, func(args ...amd64.Register) {
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

	f.Comment("same as butterflyD2Q but reduces in1 to [0,q)")
	_ = f.Define("butterflyD1Q", 5, func(args ...amd64.Register) {
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
	_ = f.Define("butterflyQ2Q", 4, func(args ...amd64.Register) {
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

	f.Comment("performs a multiplication in place between 2 vectors of qwords (values should be dwords zero extended)")
	f.Comment("in0 = (in0 * in1) mod q")
	f.Comment("in1: second operand")
	f.Comment("in2: mask for low dword in each qword")
	f.Comment("in3: q broadcasted on all qwords lanes")
	f.Comment("in4: qInvNeg broadcasted on all qwords lanes")
	f.Comment("in5: temporary Z register")
	f.Comment("in6: temporary Z register")
	_ = f.Define("mul", 7, func(args ...amd64.Register) {
		x := args[0]
		y := args[1]
		LSW := args[2]
		q := args[3]
		qInvNeg := args[4]
		P := args[5]
		PL := args[6]

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

	f.WriteLn(`
	// goes from
	// Z1 = A A A A B B B B
	// Z2 = C C C C D D D D
	// we want
	// Z1 = A A A A C C C C
	// Z2 = B B B B D D D D`)
	_ = f.Define("permute4x4", 4, func(args ...amd64.Register) {
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
	_ = f.Define("permute2x2", 5, func(args ...amd64.Register) {
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

	_ = f.Define("permute1x1", 4, func(args ...amd64.Register) {
		x := args[0]
		y := args[1]
		b0 := args[2]
		K := args[3]

		f.VPSHRDQ("$32", y, x, b0)
		f.VPBLENDMD(x, b0, x, K)
		f.VPBLENDMD(b0, y, y, K)
	})

	_ = f.Define("PACK_DWORDS", 4, func(args ...amd64.Register) {
		x := args[0]
		xx := args[1]
		y := args[2]
		xy := args[3]

		f.VPMOVQD(x, xx)
		f.VPMOVQD(y, xy)
		f.VINSERTI64X4(1, xy, x, x)
	})
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

	a := amd64.Register("Z0")
	am := amd64.Register("Z1")
	qd := amd64.Register("Z2")
	b0 := amd64.Register("Z3")
	q := amd64.Register("Z8")
	qInvNeg := amd64.Register("Z9")
	PL := amd64.Register("Z10")
	LSW := amd64.Register("Z11")
	P := amd64.Register("Z12")
	m1 := amd64.Register("Z13")
	m2 := amd64.Register("Z14")
	t0 := amd64.Register("Z15")
	t1 := amd64.Register("Z16")
	y1 := amd64.Register("Y20")
	y2 := amd64.Register("Y21")

	f.Comment("prepare constants needed for mul and reduce ops")
	f.WriteLn("MOVD $const_q, AX")
	f.VPBROADCASTD(amd64.AX, qd)
	f.VPBROADCASTQ(amd64.AX, q)
	f.WriteLn("MOVD $const_qInvNeg, AX")
	f.VPBROADCASTQ(amd64.AX, qInvNeg)
	f.VPCMPEQB("Y0", "Y0", "Y0")
	f.VPMOVZXDQ("Y0", LSW)

	f.Comment("load arguments")
	f.MOVQ("a+0(FP)", addrA)
	f.MOVQ("twiddles+24(FP)", addrTwiddles)
	f.MOVQ("end+56(FP)", len)
	f.MOVQ("m+64(FP)", m)

	// get defines
	butterflyD2Q, _ := f.DefineFn("butterflyD2Q")
	mul, _ := f.DefineFn("mul")

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

	f.VMOVDQA32(addrA.At(0), a, "load a[i]")
	f.VMOVDQA32(addrAPlusM.At(0), am, "load a[i+m]")

	butterflyD2Q(a, am, qd, b0)

	// a is ready to be stored, but we need to scale am by twiddles.
	f.VMOVDQA32(a, addrA.At(0), "store a[i]")

	// we split am into m1 and m2;
	// that is am contains 16 uint32
	// but we want that to be 2x8 uint64
	f.VEXTRACTI32X8(0, am, y1)
	f.VEXTRACTI32X8(1, am, y2)
	f.VPMOVZXDQ(y1, m1)
	f.VPMOVZXDQ(y2, m2)

	// load twiddles
	f.VPMOVZXDQ(addrTwiddles.At(0), t0)
	f.VPMOVZXDQ(addrTwiddles.At(4), t1)

	mul(m1, t0, LSW, q, qInvNeg, P, PL)
	mul(m2, t1, LSW, q, qInvNeg, P, PL)

	// store m1 and m2
	f.VPMOVQD(m1, addrAPlusM.At(0))
	f.VPMOVQD(m2, addrAPlusM.At(4))

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

func (f *FFAmd64) generateFFTKernelF31(klog2 int) {
	if klog2 != 7 {
		panic("not implemented")
	}
	// for now we generate kernels of size 1 << 7 (128) only
	// as we can keep the input and twiddles in registers and avoid round trips with memory.
	// perf note: we could generate a larger kernel, maybe up to 512 and process the "left" part of the FFT
	// fully in registers. may not be clearly worth it since it would only save 3 calls to the assembly
	// innerDIFWithTwiddles ; + the latency to write a to L1 cache.
	n := 1 << klog2
	f.Comment(fmt.Sprintf("kerDIFNP_%d_avx512(a []{{ .FF }}.Element, twiddles [][]{{ .FF }}.Element, stage int)", n))
	const argSize = 7 * 8
	stackSize := f.StackSize(f.NbWords*2+4, 1, 0)
	registers := f.FnHeader(fmt.Sprintf("kerDIFNP_%d_avx512", n), stackSize, argSize, amd64.AX)
	defer f.AssertCleanStack(stackSize, 0)

	// registers & labels we need
	addrA := f.Pop(&registers)
	addrAPlusM := f.Pop(&registers)
	addrTwiddlesRoot := f.Pop(&registers)
	addrTwiddles := f.Pop(&registers)
	innerLen := f.Pop(&registers)
	addrVInterleaveIndices := f.Pop(&registers)

	// AVX512 registers
	// Z0-Z15 taken by a
	a := make([]amd64.Register, 16)
	for i := range a {
		a[i] = amd64.Register(fmt.Sprintf("Z%d", i))
	}

	P := amd64.Register("Z16")
	q := amd64.Register("Z17")
	qInvNeg := amd64.Register("Z18")
	PL := amd64.Register("Z19")
	LSW := amd64.Register("Z20")
	b0 := amd64.Register("Z21")
	b1 := amd64.Register("Z22")
	_ = b1

	// t takes Z23 -> Z31
	t := make([]amd64.Register, 8)
	for i := range t {
		t[i] = amd64.Register(fmt.Sprintf("Z%d", 23+i))
	}
	tx0 := amd64.Register("X23")
	ty0 := amd64.Register("Y23")

	// load q and qInvNeg
	f.Comment("prepare constants needed for mul and reduce ops")
	f.WriteLn("MOVD $const_q, AX")
	f.VPBROADCASTQ(amd64.AX, q)
	f.WriteLn("MOVD $const_qInvNeg, AX")
	f.VPBROADCASTQ(amd64.AX, qInvNeg)
	f.VPCMPEQB("Y0", "Y0", "Y0")
	f.VPMOVZXDQ("Y0", LSW)

	f.Comment("load arguments")
	f.MOVQ("a+0(FP)", addrA)
	f.MOVQ("twiddles+24(FP)", addrTwiddlesRoot)
	f.MOVQ("stage+48(FP)", amd64.AX)
	f.IMULQ("$24", amd64.AX)
	f.ADDQ(amd64.AX, addrTwiddlesRoot, "we want twiddles[stage] as starting point")

	f.Comment("load a[:128] in registers")
	for i := range a {
		// addr.At gives i*8 (word size);
		// we want to advance by 32bytes to have 8 uint32 element loaded at a time.
		f.VPMOVZXDQ(addrA.At(i*4), a[i])
	}

	// get the defines
	butterflyQ2Q, _ := f.DefineFn("butterflyQ2Q")
	mul, _ := f.DefineFn("mul")

	const kBlendEven4 = 0x0f0f
	f.MOVQ(uint64(kBlendEven4), amd64.AX)
	f.KMOVQ(amd64.AX, "K1")
	permute4x4, _ := f.DefineFn("permute4x4")

	m := n >> 1

	// perf note: we could handle the case m == 16 a bit differently
	// (see innerDIFWithTwiddles)
	// and likely save some cycles; keeping as is for now for simplicity.
	for m >= 8 {

		f.MOVQ(addrTwiddlesRoot.At(0), addrTwiddles)
		nbTwiddles := m
		for i := 0; i < nbTwiddles/8; i++ {
			f.VPMOVZXDQ(addrTwiddles.At(i*4), t[i])
		}

		am := m / 8
		for offset := 0; offset < 128; offset += n {
			aa := a[offset/8:]
			for i := 0; i < am; i++ {
				butterflyQ2Q(aa[i], aa[i+am], q, b0)
				mul(aa[i+am], t[i], LSW, q, qInvNeg, P, PL)
			}
		}

		n >>= 1
		m = n >> 1

		// increment addrTwiddlesRoot
		f.ADDQ("$24", addrTwiddlesRoot)
	}

	// here we should have m == 2
	if m != 4 {
		panic("unexpected m value")
	}

	// for m == 4, we are going to permute some lanes;
	// we have for example
	// Z1 = A A A A B B B B
	// Z2 = C C C C D D D D
	// we want
	// Z1 = A A A A C C C C
	// Z2 = B B B B D D D D
	// and then we can do our butterfly ops,
	// our scaling by twiddles
	// and permute back.
	//
	// similarly, we need to "pack" 2x4 twiddles
	// into a single Z register
	f.MOVQ(addrTwiddlesRoot.At(0), addrTwiddles)
	f.VPMOVZXDQ(addrTwiddles.At(0), ty0, "zero extend 4x uint32 to 4x uint64")
	f.VINSERTI64X4(1, ty0, t[0], t[0])

	// now we process the a[i] 2 by 2 and permute before / after the ops.
	for offset := 0; offset < 128; offset += n * 2 {
		// note that we advance by 2*n, that is 16 uint32
		// that is 2 ZMM vectors
		x := a[offset/8]
		y := a[(offset/8)+1]

		// first we need to permute 4 last of x with 4 first of y
		permute4x4(x, y, b0, "K1")
		butterflyQ2Q(x, y, q, b0)
		mul(y, t[0], LSW, q, qInvNeg, P, PL)

		// permute back
		permute4x4(x, y, b0, "K1")
	}

	n >>= 1

	// now m == 2 our permutation may cost a bit more but let's see.

	// increment addrTwiddlesRoot
	f.ADDQ("$24", addrTwiddlesRoot)
	// we could probably extract the twiddles from t[0] with a stride.
	// we know we want t[0] = 1 t2 1 t2 1 t2 1 t2
	// and we have from m == 4
	// t[0] = 1 t1 t2 t3 1 t1 t2 t3
	f.MOVQ(addrTwiddlesRoot.At(0), addrTwiddles)
	f.VPMOVZXDQ(addrTwiddles.At(0), tx0, "zero extend 2x uint32 to 2x uint64")
	f.VINSERTI64X2(1, tx0, t[0], t[0])
	f.VINSERTI64X2(2, tx0, t[0], t[0])
	f.VINSERTI64X2(3, tx0, t[0], t[0])

	const kBlendEven = 0b00110011
	f.MOVQ(uint64(kBlendEven), amd64.AX)
	f.KMOVQ(amd64.AX, "K2")

	vInterleaveIndices := t[7]
	f.MOVQ("·vInterleaveIndices+0(SB)", addrVInterleaveIndices)
	f.VMOVDQU64(addrVInterleaveIndices.At(0), vInterleaveIndices)

	permute2x2, _ := f.DefineFn("permute2x2")

	// for offset := 0; offset < 128; offset += 4 {
	// 	innerDIFWithTwiddles(a[offset:offset+4], twiddles[stage+5], 0, 2, 2)
	// }
	for offset := 0; offset < 128; offset += n * 4 {
		// note that we advance by 4*n, that is 16 uint32
		// that is 2 ZMM vectors
		x := a[offset/8]
		y := a[(offset/8)+1]

		// first we need to permute 4 last of x with 4 first of y
		permute2x2(x, y, vInterleaveIndices, t[6], "K2")
		butterflyQ2Q(x, y, q, b0)
		mul(y, t[0], LSW, q, qInvNeg, P, PL)

		// invert back
		permute2x2(x, y, vInterleaveIndices, t[6], "K2")
	}

	const kBlendEven2 = 0b0101010101010101

	f.MOVQ(uint64(kBlendEven2), amd64.AX)
	f.KMOVD(amd64.AX, "K3")

	permute1x1, _ := f.DefineFn("permute1x1")

	f.WriteLn("MOVD $const_q, AX")
	f.VPBROADCASTD(amd64.AX, q, "rebroadcast q, but on dword lanes")

	butterflyD1Q, _ := f.DefineFn("butterflyD1Q")
	packDWORDS, _ := f.DefineFn("PACK_DWORDS")

	// now m == 1, last step is only butterflies like so
	// for offset := 0; offset < 128; offset += 2 {
	// 	koalabear.Butterfly(&a[offset], &a[offset+1])
	// }
	// our a vectors are on QWORDS lanes, we can pack them into DWORD lanes to reduce nb ops
	for i := 0; i < len(a); i += 4 {
		u, v, w, x := a[i], a[i+1], a[i+2], a[i+3]
		packDWORDS(u, zToy(u), v, zToy(v))
		packDWORDS(w, zToy(w), x, zToy(x))

		permute1x1(u, w, b0, "K3")
		butterflyD1Q(u, w, q, b0, b1)
		permute1x1(u, w, b0, "K3")
	}

	// end we store back a
	f.Comment("store a[:128] in memory")
	for i := 0; i < len(a); i += 2 {
		f.VMOVDQA32(a[i], addrA.At(i*4))
	}

	f.RET()

	f.Push(&registers, addrA, addrTwiddles, addrAPlusM, innerLen)

}

func zToy(r amd64.Register) amd64.Register {
	v := string(r) // v will be Z1, Z2, ... etc, we want to return Y1, Y2, ...
	vr := "Y" + v[1:]
	return amd64.Register(vr)
}
