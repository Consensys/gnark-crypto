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
