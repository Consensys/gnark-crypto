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
	f.VSUB(q.S4(), b.S4(), t.S4(), "t = b - q")
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

func (f *FFArm64) generateMulVecF31() {
	f.Comment("mulVec(res, a, b *Element, n uint64)")
	registers := f.FnHeader("mulVec", 0, 32)
	defer f.AssertCleanStack(0, 0)

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

	// a := registers.PopV()
	// b := registers.PopV()
	// t := registers.PopV()
	// q := registers.PopV()
	// qInvNeg := registers.PopV()

	// f.VMOVS("$const_q", q)
	// f.VDUP(q.SAt(0), q.S4(), "broadcast q into "+string(q))

	// f.VMOVS("$const_qInvNeg", qInvNeg)
	// f.VDUP(qInvNeg.SAt(0), qInvNeg.S4(), "broadcast qInvNeg into "+string(qInvNeg))

	q := registers.PopV()

	f.VMOVS("$const_q", q)
	f.VDUP(q.DAt(0), q.D2(), "broadcast q into "+string(q))

	const maxUint32 = 0xFFFFFFFF
	mask := registers.PopV()
	f.VMOVQ_cst(maxUint32, maxUint32, mask)

	f.LABEL(loop)

	f.CBZ(n, done)

	a0 := registers.Pop()
	a1 := registers.Pop()
	b0 := registers.Pop()
	b1 := registers.Pop()
	r0 := registers.Pop()
	r1 := registers.Pop()

	v := registers.PopV()
	m := registers.PopV()
	t1 := registers.PopV()
	t2 := registers.PopV()

	a0_2 := registers.Pop()
	a1_2 := registers.Pop()
	b0_2 := registers.Pop()
	b1_2 := registers.Pop()
	r0_2 := registers.Pop()
	r1_2 := registers.Pop()
	v_2 := registers.PopV()
	m_2 := registers.PopV()
	t1_2 := registers.PopV()
	t2_2 := registers.PopV()

	// let's do 2 by 2 to start with;
	f.MOVWUP_Load(4, aPtr, a0)
	f.MOVWUP_Load(4, aPtr, a1)
	f.MOVWUP_Load(4, bPtr, b0)
	f.MOVWUP_Load(4, bPtr, b1)

	f.MUL(a0, b0, r0)
	f.MUL(a1, b1, r1)

	f.VMOV(r0, v.DAt(0))
	f.VMOV(r1, v.DAt(1))

	// qInvNeg == 2**31 - 2**24 -1
	// so we shift left by 31, store in a vector
	// we shift left by 24, store in a vector
	// we subtract the two vectors
	f.VSHL(31, v.D2(), t1.D2())
	f.VSHL(24, v.D2(), t2.D2())
	f.MOVWUP_Load(4, aPtr, a0_2)
	f.MOVWUP_Load(4, aPtr, a1_2)

	f.VSUB(t2.D2(), t1.D2(), t1.D2())
	f.VSUB(v.D2(), t1.D2(), m.D2())
	f.MOVWUP_Load(4, bPtr, b0_2)
	f.MOVWUP_Load(4, bPtr, b1_2)

	// here we just want to keep m=low bits(vRes)
	f.VAND(m.B16(), mask.B16(), m.B16())

	// q == 2**31 - 2**24 + 1
	f.VSHL(31, m.D2(), t1.D2())
	f.VSHL(24, m.D2(), t2.D2())
	f.VSUB(t2.D2(), t1.D2(), t1.D2())
	f.VADD(m.D2(), t1.D2(), m.D2())

	f.VADD(m.D2(), v.D2(), m.D2())
	f.VUSHR(32, m.D2(), m.D2())

	// now we do mod q if needed
	f.VSUB(q.D2(), m.D2(), t1.D2(), "t = q - m")
	f.VUMIN(t1.S4(), m.S4(), m.S4(), "m = min(t, m)")

	f.VSHL(32, m.D2(), m.D2())

	// f.VMOV(m.DAt(0), r0)
	// f.VMOV(m.DAt(1), r1)

	// f.MOVWUP_Store(r0, resPtr, 4)
	// f.MOVWUP_Store(r1, resPtr, 4)

	f.MUL(a0_2, b0_2, r0_2)
	f.MUL(a1_2, b1_2, r1_2)

	f.VMOV(r0_2, v_2.DAt(0))
	f.VMOV(r1_2, v_2.DAt(1))

	// qInvNeg == 2**31 - 2**24 -1
	// so we shift left by 31, store in a vector
	// we shift left by 24, store in a vector
	// we subtract the two vectors
	f.VSHL(31, v_2.D2(), t1_2.D2())
	f.VSHL(24, v_2.D2(), t2_2.D2())
	f.VSUB(t2_2.D2(), t1_2.D2(), t1_2.D2())
	f.VSUB(v_2.D2(), t1_2.D2(), m_2.D2())

	// here we just want to keep m=low bits(vRes)
	f.VAND(m_2.B16(), mask.B16(), m_2.B16())

	// q == 2**31 - 2**24 + 1
	f.VSHL(31, m_2.D2(), t1_2.D2())
	f.VSHL(24, m_2.D2(), t2_2.D2())
	f.VSUB(t2_2.D2(), t1_2.D2(), t1_2.D2())
	f.VADD(m_2.D2(), t1_2.D2(), m_2.D2())

	f.VADD(m_2.D2(), v_2.D2(), m_2.D2())
	f.VUSHR(32, m_2.D2(), m_2.D2())

	// now we do mod q if needed
	f.VSUB(q.D2(), m_2.D2(), t1_2.D2(), "t = q - m")
	f.VUMIN(t1_2.S4(), m_2.S4(), m_2.S4(), "m = min(t, m)")

	f.VADD(m_2.S4(), m.S4(), m.S4())
	// f.VREV64(m.B16(), m.B16())

	f.VST1_P(m.S4(), resPtr, 4*4, "res = b")

	// f.VMOV(m_2.DAt(0), r0_2)
	// f.VMOV(m_2.DAt(1), r1_2)

	// f.MOVWUP_Store(r0_2, resPtr, 4)
	// f.MOVWUP_Store(r1_2, resPtr, 4)

	// func montReduce(v uint64) uint32 {
	// 	m := uint32(v) * qInvNeg
	// 	t := uint32((v + uint64(m)*q) >> 32)
	// 	if t >= q {
	// 		t -= q
	// 	}
	// 	return t
	// }

	// 		g.VST1_P(vRes.D2(), resPtr, 0)

	// const offset = 4 * 4 // we process 4 uint32 at a time

	// f.VLD1_P(offset, aPtr, a.S4())
	// f.VLD1_P(offset, bPtr, b.S4())

	// // let's compute p1 := a1 * b1
	// // f.VPMULL(a.S4(), b.S4(), p1.D2())
	// // let's move the low words in t
	// // f.VMOV(p1.D2(), t.D2())

	// f.VUSHLL2(0, a.S4(), a.D2(), "convert high words to 64 bits")
	// f.VUSHLL2(0, b.S4(), b.D2(), "convert high words to 64 bits")

	// // f.VMUL(a.S4(), b.S4(), b.S4(), "b = a * b")
	// f.VSUB(q.S4(), b.S4(), t.S4(), "t = q - b")
	// f.VUMIN(t.S4(), b.S4(), b.S4(), "b = min(t, b)")
	// f.VST1_P(b.S4(), resPtr, offset, "res = b")

	// decrement n
	f.SUB(1, n, n)
	f.JMP(loop)

	f.LABEL(done)

	f.RET()
}
