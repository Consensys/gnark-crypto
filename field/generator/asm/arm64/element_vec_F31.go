package arm64

func (f *FFArm64) generateAddVecF31() {
	f.Comment("addVec(res, a, b *Element, n uint64)")
	registers := f.FnHeader("addVec", 0, 32)
	defer f.AssertCleanStack(0, 0)

	// registers
	resPtr := registers.Pop()
	aPtr := registers.Pop()
	bPtr := registers.Pop()
	n := registers.Pop()

	b := registers.Pop()
	a := registers.Pop()
	q := registers.Pop()
	t := registers.Pop()

	f.MOVD("$const_q", q)

	// labels
	loop := f.NewLabel("loop")
	done := f.NewLabel("done")

	// load arguments
	f.LDP("res+0(FP)", resPtr, aPtr)
	f.LDP("b+16(FP)", bPtr, n)

	f.LABEL(loop)

	f.CBZ(n, done)

	// load a
	f.MOVWUP_Load(4, aPtr, a)
	// load b
	f.MOVWUP_Load(4, bPtr, b)

	// res = a + b
	f.ADD(a, b, b)

	// t = res - q
	f.SUBS(q, b, t)

	// t = min(t, res)
	f.CSEL("CS", t, b, t)

	// res = t
	f.MOVWUP_Store(b, resPtr, 4)

	// decrement n
	f.SUB(1, n, n)
	f.JMP(loop)

	f.LABEL(done)
	f.RET()

}
