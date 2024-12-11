package arm64

import "github.com/consensys/bavard/arm64"

func (f *FFArm64) generateAddVecF31() {
	f.Comment("addVec(qq *uint32, res, a, b *Element, n uint64)")
	registers := f.FnHeader("addVec", 0, 40)
	defer f.AssertCleanStack(0, 0)

	// registers
	resPtr := registers.Pop()
	qqPtr := registers.Pop()
	aPtr := registers.Pop()
	bPtr := registers.Pop()
	n := registers.Pop()

	a := arm64.V0.S4()
	b := arm64.V1.S4()
	q := arm64.V2.S4()
	t := arm64.V3.S4()

	// labels
	loop := f.NewLabel("loop")
	done := f.NewLabel("done")

	// load arguments
	f.LDP("qq+0(FP)", qqPtr, resPtr)
	f.LDP("a+16(FP)", aPtr, bPtr)
	f.MOVD("n+32(FP)", n)

	f.VLD1(0, qqPtr, q, "broadcast q into "+string(q))

	f.LABEL(loop)

	f.CBZ(n, done)

	const offset = 4 * 4 // we process 4 uint32 at a time

	f.VLD1_P(offset, aPtr, a)
	f.VLD1_P(offset, bPtr, b)

	f.VADD(a, b, b, "b = a + b")
	f.VSUB(q, b, t, "t = q - b")
	f.VUMIN(t, b, b, "b = min(t, b)")
	f.VST1_P(b, resPtr, offset, "res = b")

	// decrement n
	f.SUB(1, n, n)
	f.JMP(loop)

	f.LABEL(done)
	f.RET()

}

func (f *FFArm64) generateSubVecF31() {
	f.Comment("subVec(qq *uint32, res, a, b *Element, n uint64)")
	registers := f.FnHeader("subVec", 0, 40)
	defer f.AssertCleanStack(0, 0)

	// registers
	resPtr := registers.Pop()
	qPtr := registers.Pop()
	aPtr := registers.Pop()
	bPtr := registers.Pop()
	n := registers.Pop()

	a := arm64.V0.S4()
	b := arm64.V1.S4()
	q := arm64.V2.S4()
	t := arm64.V3.S4()

	// labels
	loop := f.NewLabel("loop")
	done := f.NewLabel("done")

	// load arguments
	f.LDP("qLane+0(FP)", qPtr, resPtr)
	f.LDP("a+16(FP)", aPtr, bPtr)
	f.MOVD("n+32(FP)", n)

	f.VLD1(0, qPtr, q, "broadcast q into "+string(q))

	f.LABEL(loop)

	f.CBZ(n, done)

	const offset = 4 * 4 // we process 4 uint32 at a time

	f.VLD1_P(offset, aPtr, a)
	f.VLD1_P(offset, bPtr, b)

	f.VSUB(b, a, b, "b = a - b")
	f.VADD(b, q, t, "t = b + q")
	f.VUMIN(t, b, b, "b = min(t, b)")
	f.VST1_P(b, resPtr, offset, "res = b")

	// decrement n
	f.SUB(1, n, n)
	f.JMP(loop)

	f.LABEL(done)
	f.RET()

}

func (f *FFArm64) generateSumVecF31() {
	f.Comment("sumVec(t *uint64, a *[]uint32, n uint64) res = sum(a[0...n])")
	registers := f.FnHeader("sumVec", 0, 3*8)
	defer f.AssertCleanStack(0, 0)

	// registers
	aPtr := registers.Pop()
	tPtr := registers.Pop()
	n := registers.Pop()

	a1 := arm64.V0
	a2 := arm64.V1
	a3 := arm64.V2
	a4 := arm64.V3
	acc1 := arm64.V4
	acc2 := arm64.V5
	acc3 := arm64.V6
	acc4 := arm64.V7

	// zero out accumulators
	f.VMOVQ_cst(0, 0, acc1)
	f.VMOVQ_cst(0, 0, acc2)
	f.VMOVQ_cst(0, 0, acc3)
	f.VMOVQ_cst(0, 0, acc4)

	acc1 = arm64.V4.D2()
	acc2 = arm64.V5.D2()
	acc3 = arm64.V6.D2()
	acc4 = arm64.V7.D2()

	// labels
	loop := f.NewLabel("loop")
	done := f.NewLabel("done")

	// load arguments
	f.LDP("t+0(FP)", tPtr, aPtr)
	f.MOVD("n+16(FP)", n)

	f.LABEL(loop)

	f.CBZ(n, done)

	const offset = 8 * 4 // we process 4 uint32 at a time

	f.VLD2_P(offset/2, aPtr, a1.S2(), a2.S2()) // load 2*2 uint32
	f.VLD2_P(offset/2, aPtr, a3.S2(), a4.S2()) // load 2*2 uint32

	f.VUSHLL(0, a1.S2(), a1.D2(), "convert to 64 bits")
	f.VUSHLL(0, a2.S2(), a2.D2(), "convert to 64 bits")
	f.VADD(a1.D2(), acc1, acc1, "acc1 += a1")
	f.VADD(a2.D2(), acc2, acc2, "acc2 += a2")

	f.VUSHLL(0, a3.S2(), a3.D2(), "convert to 64 bits")
	f.VUSHLL(0, a4.S2(), a4.D2(), "convert to 64 bits")
	f.VADD(a3.D2(), acc3, acc3, "acc3 += a3")
	f.VADD(a4.D2(), acc4, acc4, "acc4 += a4")

	// decrement n
	f.SUB(1, n, n)
	f.JMP(loop)

	f.LABEL(done)

	f.VADD(acc1, acc3, acc1, "acc1 += acc3")
	f.VADD(acc2, acc4, acc2, "acc2 += acc4")

	f.VST2_P(acc1, acc2, tPtr, 0, "store acc1 and acc2")

	f.RET()

}
