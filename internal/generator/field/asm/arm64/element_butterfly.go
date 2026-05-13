package arm64

func (f *FFArm64) generateButterfly() {
	f.Comment("butterfly(a, b *Element)")
	f.Comment("a, b = a+b, a-b")
	registers := f.FnHeader("Butterfly", 0, 16)
	defer f.AssertCleanStack(0, 0)

	// registers
	a := registers.PopN(f.NbWords)
	b := registers.PopN(f.NbWords)
	r := registers.PopN(f.NbWords)
	t := registers.PopN(f.NbWords)
	aPtr := registers.Pop()
	bPtr := registers.Pop()

	f.LDP("x+0(FP)", aPtr, bPtr)
	f.load(aPtr, a)
	f.load(bPtr, b)

	for i := 0; i < f.NbWords; i++ {
		f.add0n(i)(a[i], b[i], r[i])
	}

	f.SUBS(b[0], a[0], b[0])
	for i := 1; i < f.NbWords; i++ {
		f.SBCS(b[i], a[i], b[i])
	}

	for i := 0; i < f.NbWords; i++ {
		if i%2 == 0 {
			f.LDP(f.qAt(i), a[i], a[i+1])
		}
		f.CSEL("CS", "ZR", a[i], t[i])
	}
	f.Comment("add q if underflow, 0 if not")
	for i := 0; i < f.NbWords; i++ {
		f.add0n(i)(b[i], t[i], b[i])
		if i%2 == 1 {
			f.STP(b[i-1], b[i], bPtr.At(i-1))
		}
	}

	f.reduceAndStore(r, a, aPtr)

	f.RET()
}
