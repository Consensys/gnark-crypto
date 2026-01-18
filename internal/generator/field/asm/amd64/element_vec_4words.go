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
	f.Comment("sumVec(t *[8]uint64, a *Element, n uint64) t = raw accumulators")
	f.Comment("Accumulates elements into 8 qword accumulators for reduction in Go")
	f.Comment("")
	f.Comment("When we move an element into a Z register using VPMOVZXDQ,")
	f.Comment("each 32-bit dword is zero-extended to a 64-bit qword.")
	f.Comment("We can safely add up to 2^32 elements without overflow.")

	const argSize = 3 * 8
	stackSize := f.StackSize(4, 2, 0)
	registers := f.FnHeader("sumVec", stackSize, argSize, amd64.DX, amd64.AX)
	defer f.AssertCleanStack(stackSize, 0)

	// registers & labels we need
	addrA := f.Pop(&registers)
	addrT := f.Pop(&registers)
	n := f.Pop(&registers)
	nMod8 := f.Pop(&registers)

	// AVX512 registers
	acc := registers.PopVN(8)
	t := registers.PopVN(8)

	// load arguments
	f.MOVQ("t+0(FP)", addrT)
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("n+16(FP)", n)

	f.Comment("initialize accumulators")
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
		f.VPMOVZXDQ("0("+addrA+")", t[0])
		f.VPADDQ(t[0], acc[0], acc[0])
		f.ADDQ("$32", addrA)
	})

	f.Loop(n, func() {
		for i := 0; i < 8; i++ {
			f.VPMOVZXDQ(addrA.At(4*i), t[i])
		}

		f.WriteLn(fmt.Sprintf("PREFETCHT0 4096(%[1]s)", addrA))
		for i := 0; i < 8; i++ {
			f.VPADDQ(t[i], acc[i], acc[i])
		}

		f.ADDQ("$256", addrA, "increment pointer by 8 elements")
	})

	f.Comment("accumulate the 8 Z registers into Z0")
	for i := 7; i > 0; i-- {
		f.VPADDQ(acc[i], acc[i-1], acc[i-1])
	}

	f.Comment("store the 8 qwords to the output buffer")
	f.VMOVDQU64(acc[0], "0("+addrT+")")

	f.RET()
}

func (f *FFAmd64) generateInnerProductW4() {
	f.Comment("innerProdVec(res, a,b *Element, n uint64) res = sum(a[0...n] * b[0...n])")

	const argSize = 4 * 8
	stackSize := f.StackSize(7, 2, 0)
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

	mac := f.Define("MAC", 3, func(inputs ...any) {
		opLeft := inputs[0]
		lo := inputs[1]
		hi := inputs[2]

		f.VPMULUDQ_BCST(opLeft, Y, PPL)
		f.VPSRLQ("$32", PPL, PPH)
		f.VPANDQ(LSW, PPL, PPL)
		f.VPADDQ(PPL, lo, lo)
		f.VPADDQ(PPH, hi, hi)
	}, true)

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
	f.VALIGNDk("$15", ACC, ACC, "K2", "Z0")
	f.KSHIFTLW("$1", "K2", "K2")

	f.Comment("macro to add partial products and store the result in Z0")
	addPP := f.Define("ADDPP", 5, func(inputs ...any) {
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
		f.VALIGNDk("$16-"+I.(amd64.Register), ACC, ACC, "K2", "Z0")
		f.KADDW("K2", "K2", "K2")
	}, true)

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
	f.VALIGNDk("$16-8", ACC, ACC, "K2", "Z0")
	f.KSHIFTLW("$1", "K2", "K2")

	f.VPSRLQ("$32", ACC, PPL)
	f.VALIGND_Z("$2", ACC, ACC, "K1", ACC)
	f.VPADDQ(PPL, ACC, ACC)
	f.VPSRLQ("$32", A7H, A7H)
	f.VPADDQ(A7H, ACC, ACC)
	f.VALIGNDk("$16-9", ACC, ACC, "K2", "Z0")
	f.KSHIFTLW("$1", "K2", "K2")

	addPP2 := f.Define("ADDPP2", 1, func(args ...any) {
		f.VPSRLQ("$32", ACC, PPL)
		f.VALIGND_Z("$2", ACC, ACC, "K1", ACC)
		f.VPADDQ(PPL, ACC, ACC)
		f.VALIGNDk("$16-"+args[0].(amd64.Register), ACC, ACC, "K2", "Z0")
		f.KSHIFTLW("$1", "K2", "K2")
	}, true)

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
