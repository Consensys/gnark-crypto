package arm64

import "github.com/consensys/bavard/arm64"

func (f *FFArm64) generateAddVecF31() {
	f.Comment("addVec(res, a, b *Element, n uint64)")
	registers := f.FnHeader("addVec", 0, 32)
	defer f.AssertCleanStack(0, 0)
	defer registers.AssertCleanState()

	// registers
	resPtr := registers.Pop()
	// qqPtr := registers.Pop()
	aPtr := registers.Pop()
	bPtr := registers.Pop()
	n := registers.Pop()

	// labels
	loop := f.NewLabel("loop")
	done := f.NewLabel("done")

	// load arguments
	f.LDP("res+0(FP)", resPtr, aPtr)
	f.LDP("b+16(FP)", bPtr, n)

	a := registers.PopV("a")
	b := registers.PopV("b")
	t := registers.PopV("t")
	q := registers.PopV("q")

	f.VMOVS("$const_q", q)
	f.VDUP(q.SAt(0), q.S4(), "broadcast q into "+string(q))

	f.LABEL(loop)

	f.CBZ(n, done)

	const offset = 4 * 4 // we process 4 uint32 at a time

	f.VLD1_P(offset, aPtr, a.S4())
	f.VLD1_P(offset, bPtr, b.S4())

	f.VADD(a.S4(), b.S4(), b.S4(), "b = a + b")
	f.VSUB(q.S4(), b.S4(), t.S4(), "t = q - b")
	f.VUMIN(t.S4(), b.S4(), b.S4(), "b = min(t, b)")
	f.VST1_P(b.S4(), resPtr, offset, "res = b")

	// decrement n
	f.SUB(1, n, n)
	f.JMP(loop)

	f.LABEL(done)

	registers.Push(resPtr, aPtr, bPtr, n)
	registers.PushV(a, b, t, q)

	f.RET()

}

func (f *FFArm64) generateSubVecF31() {
	f.Comment("subVec(res, a, b *Element, n uint64)")
	registers := f.FnHeader("subVec", 0, 32)
	defer f.AssertCleanStack(0, 0)
	defer registers.AssertCleanState()

	// registers
	resPtr := registers.Pop()
	aPtr := registers.Pop()
	bPtr := registers.Pop()
	n := registers.Pop()

	// labels
	loop := f.NewLabel("loop")
	done := f.NewLabel("done")

	// load arguments
	f.LDP("res+0(FP)", resPtr, aPtr)
	f.LDP("b+16(FP)", bPtr, n)

	a := registers.PopV("a")
	b := registers.PopV("b")
	t := registers.PopV("t")
	q := registers.PopV("q")

	f.VMOVS("$const_q", q)
	f.VDUP(q.SAt(0), q.S4(), "broadcast q into "+string(q))

	f.LABEL(loop)

	f.CBZ(n, done)

	const offset = 4 * 4 // we process 4 uint32 at a time

	f.VLD1_P(offset, aPtr, a.S4())
	f.VLD1_P(offset, bPtr, b.S4())

	f.VSUB(b.S4(), a.S4(), b.S4(), "b = a - b")
	f.VADD(b.S4(), q.S4(), t.S4(), "t = b + q")
	f.VUMIN(t.S4(), b.S4(), b.S4(), "b = min(t, b)")
	f.VST1_P(b.S4(), resPtr, offset, "res = b")

	// decrement n
	f.SUB(1, n, n)
	f.JMP(loop)

	f.LABEL(done)

	registers.Push(resPtr, aPtr, bPtr, n)
	registers.PushV(a, b, q, t)

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

	a1 := registers.PopV()
	a2 := registers.PopV()
	a3 := registers.PopV()
	a4 := registers.PopV()
	acc1 := registers.PopV()
	acc2 := registers.PopV()
	acc3 := registers.PopV()
	acc4 := registers.PopV()

	f.Comment("zeroing accumulators")
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

	f.WriteLn(`
	// blockSize is 16 uint32; we load 4 vectors of 4 uint32 at a time
	// (4*4)*4 = 64 bytes ~= 1 cache line
	// since our values are 31 bits, we can add 2 by 2 these vectors
	// we are left with 2 vectors of 4x32 bits values
	// that we accumulate in 4*2*64bits accumulators
	// the caller will reduce mod q the accumulators.
	`)

	const offset = 8 * 4
	f.VLD2_P(offset, aPtr, a1.S4(), a2.S4())
	f.VADD(a1.S4(), a2.S4(), a1.S4(), "a1 += a2")

	f.VLD2_P(offset, aPtr, a3.S4(), a4.S4())
	f.VADD(a3.S4(), a4.S4(), a3.S4(), "a3 += a4")

	f.VUSHLL(0, a1.S2(), a2.D2(), "convert low words to 64 bits")
	f.VADD(a2.D2(), acc2, acc2, "acc2 += a2")
	f.VUSHLL2(0, a1.S4(), a1.D2(), "convert high words to 64 bits")
	f.VADD(a1.D2(), acc1, acc1, "acc1 += a1")

	f.VUSHLL(0, a3.S2(), a4.D2(), "convert low words to 64 bits")
	f.VADD(a4.D2(), acc4, acc4, "acc4 += a4")
	f.VUSHLL2(0, a3.S4(), a3.D2(), "convert high words to 64 bits")
	f.VADD(a3.D2(), acc3, acc3, "acc3 += a3")

	// decrement n
	f.SUB(1, n, n)
	f.JMP(loop)

	f.LABEL(done)

	f.VADD(acc1, acc3, acc1, "acc1 += acc3")
	f.VADD(acc2, acc4, acc2, "acc2 += acc4")

	f.VST2_P(acc1, acc2, tPtr, 0, "store acc1 and acc2")

	f.RET()

}
