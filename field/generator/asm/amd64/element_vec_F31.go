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

func (f *FFAmd64) generateSumVecSmallF31(size int) {
	// size must be 16 or 24
	if size != 16 && size != 24 {
		panic("size must be 16 or 24")
	}
	fName := fmt.Sprintf("sumVec%d_AVX512", size)
	const argSize = 2 * 8
	stackSize := f.StackSize(f.NbWords*3+2, 0, 0)
	registers := f.FnHeader(fName, stackSize, argSize, amd64.DX, amd64.AX)
	defer f.AssertCleanStack(stackSize, 0)

	// registers & labels we need
	addrA := registers.Pop()
	addrT := registers.Pop()

	// AVX512 registers
	a1 := registers.PopV()
	a2 := registers.PopV()
	a3 := registers.PopV()

	// load arguments
	f.MOVQ("t+0(FP)", addrT)
	f.MOVQ("a+8(FP)", addrA)

	f.VPMOVZXDQ(addrA.AtD(0), a1)
	f.VPMOVZXDQ(addrA.AtD(8), a2)
	if size == 24 {
		f.VPMOVZXDQ(addrA.AtD(16), a3)
		f.VPADDQ(a3, a2, a2)
	}

	f.VPADDQ(a1, a2, a1)
	f.VEXTRACTI64X4(1, a1, a2.Y())
	f.VPADDQ(a1.Y(), a2.Y(), a1.Y())
	f.VEXTRACTI64X2(1, a1.Y(), a2.X())
	f.VPADDQ(a1.X(), a2.X(), a1.X())

	f.PEXTRQ(0, a1.X(), amd64.AX)
	f.PEXTRQ(1, a1.X(), amd64.DX)
	f.ADDQ(amd64.DX, amd64.AX)
	f.MOVQ(amd64.AX, addrT.At(0))

	f.RET()

	f.Push(&registers, addrA, addrT)
}

// mulVec res = a * b
func (f *FFAmd64) generateMulVecF31() {
	f.Comment("mulVec(res, a, b *Element, n uint64) res[0...n] = a[0...n] * b[0...n]")
	f.Comment("n is the number of blocks of 16 elements to process")
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

	f.MOVQ(uint64(0b0101010101010101), amd64.AX)
	f.KMOVD(amd64.AX, amd64.K3)

	f.LABEL(loop)

	f.TESTQ(len, len)
	f.JEQ(done, "n == 0, we are done")

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
	f.Comment("scalarMulVec(res, a, b *Element, n uint64) res[0...n] = a[0...n] * b")
	f.Comment("n is the number of blocks of 16 elements to process")
	const argSize = 4 * 8
	stackSize := f.StackSize(f.NbWords*2+4, 0, 0)
	registers := f.FnHeader("scalarMulVec", stackSize, argSize, amd64.AX)
	defer f.AssertCleanStack(stackSize, 0)

	// registers & labels we need
	addrA := registers.Pop()
	addrB := registers.Pop()
	addrRes := registers.Pop()
	len := registers.Pop()

	// AVX512 registers
	a := registers.PopV()
	b := registers.PopV()
	b1 := registers.PopV()
	aOdd := registers.PopV()
	b0 := registers.PopV()
	PL0 := registers.PopV()
	PL1 := registers.PopV()
	q := registers.PopV()
	qInvNeg := registers.PopV()

	// load q in Z3
	f.MOVD("$const_q", amd64.AX)
	f.VPBROADCASTD(amd64.AX, q)
	f.MOVD("$const_qInvNeg", amd64.AX)
	f.VPBROADCASTD(amd64.AX, qInvNeg)

	f.MOVQ(uint64(0b0101010101010101), amd64.AX)
	f.KMOVD(amd64.AX, amd64.K3)

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
	f.VMOVDQU32(addrA.At(0), a)

	f.VPSRLQ("$32", a, aOdd) // keep high 32 bits
	f.VPMULUDQ(a, b, b0)
	f.VPMULUDQ(aOdd, b, b1)
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
	f.ADDQ("$64", addrRes)
	f.DECQ(len, "decrement n")
	f.JMP(loop)

	f.LABEL(done)

	f.RET()

	f.Push(&registers, addrA, addrB, addrRes, len)
}

// innerProdVec res = sum(a * b)
func (f *FFAmd64) generateInnerProdVecF31() {
	f.Comment("innerProdVec(t *uint64, a,b *[]uint32, n uint64) res = sum(a[0...n] * b[0...n])")
	f.Comment("n is the number of blocks of 16 elements to process")
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
	f.MOVQ("t+0(FP)", addrT)
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("b+16(FP)", addrB)
	f.MOVQ("n+24(FP)", len)

	a := registers.PopV()
	b := registers.PopV()
	b1 := registers.PopV()
	aOdd := registers.PopV()
	bOdd := registers.PopV()
	b0 := registers.PopV()
	PL0 := registers.PopV()
	PL1 := registers.PopV()

	acc0 := registers.PopV()
	acc1 := registers.PopV()

	f.MOVQ(uint64(0b0101010101010101), amd64.AX)
	f.KMOVD(amd64.AX, amd64.K3)

	// zeroize the accumulators
	f.VXORPS(acc0, acc0, acc0, "acc0 = 0")
	f.VMOVDQA64(acc0, acc1, "acc1 = 0")

	f.LABEL(loop)

	f.TESTQ(len, len)
	f.JEQ(done, "n == 0, we are done")

	// a = a * b
	f.VMOVDQU32(addrA.At(0), a)
	f.VMOVDQU32(addrB.At(0), b)

	f.VPSRLQ("$32", a, aOdd) // keep high 32 bits
	f.VPSRLQ("$32", b, bOdd) // keep high 32 bits

	f.VPMULUDQ(a, b, b0)
	f.VPMULUDQ(aOdd, bOdd, b1)
	f.VPMULUDQ(b0, qInvNeg, PL0)
	f.VPMULUDQ(b1, qInvNeg, PL1)

	f.VPMULUDQ(PL0, q, PL0)
	f.VPMULUDQ(PL1, q, PL1)

	f.VPADDQ(b0, PL0, b0)
	f.VPSRLQ("$32", b0, b0)
	f.VPADDQ(b0, acc0, acc0)
	f.VPADDQ(b1, PL1, b1)
	f.VPSRLQ("$32", b1, b1)
	f.VPADDQ(b1, acc1, acc1)

	f.Comment("increment pointers to visit next element")
	f.ADDQ("$64", addrA)
	f.ADDQ("$64", addrB)
	f.DECQ(len, "decrement n")
	f.JMP(loop)

	f.LABEL(done)

	// store t into res
	f.VPADDQ(acc1, acc0, acc1, "acc1 += acc0")
	f.VMOVDQU64(acc1, addrT.At(0), "res = acc1")

	f.RET()

	f.Push(&registers, addrA, addrB, addrT, len)
}
