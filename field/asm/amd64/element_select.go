package amd64

func (f *FFAmd64) generateSelect() {
	f.Comment("select_(res *Element, c int64, x0, x1 *Element)")
	registers := f.FnHeader("select_", 0, 32)
	defer f.AssertCleanStack(0, 0)

	x0 := registers.Pop()
	x1 := registers.Pop()
	t0 := registers.Pop()
	t1 := registers.Pop()
	r := registers.Pop()

	f.MOVQ("x1+24(FP)", x1)
	f.MOVQ("x0+16(FP)", x0)
	f.MOVQ("c+8(FP)", t0)
	f.MOVQ("res+0(FP)", r)

	f.TESTQ(t0, t0)
	for i := 0; i < f.NbWords; i++ {
		f.MOVQ(x0.At(i), t0)
		f.MOVQ(x1.At(i), t1)
		f.CMOVQEQ(t0, t1)
		f.MOVQ(t1, r.At(i))
	}
	registers.Push(x1, t0, t1, r)

	f.RET()
}
