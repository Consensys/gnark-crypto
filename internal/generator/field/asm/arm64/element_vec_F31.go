package arm64

import (
	"github.com/consensys/bavard/arm64"
)

func (f *FFArm64) generateAddVecF31() {
	f.Comment("addVec(res, a, b *Element, n uint64)")
	f.Comment("n is the number of blocks of 16 uint32 to process")
	registers := f.FnHeader("addVec", 0, 32)
	defer f.AssertCleanStack(0, 0)

	// registers
	resPtr := registers.Pop()
	aPtr := registers.Pop()
	bPtr := registers.Pop()
	n := registers.Pop()

	// labels
	loop := f.NewLabel("loop")
	lastBlock := f.NewLabel("lastBlock")

	// load arguments
	f.LDP("res+0(FP)", resPtr, aPtr)
	f.LDP("b+16(FP)", bPtr, n)

	a0 := registers.PopV()
	a1 := registers.PopV()
	a2 := registers.PopV()
	a3 := registers.PopV()
	b0 := registers.PopV()
	b1 := registers.PopV()
	b2 := registers.PopV()
	b3 := registers.PopV()

	q := registers.PopV()

	f.VMOVS("$const_q", q)
	f.VDUP(q.SAt(0), q.S4(), "broadcast q into "+string(q))

	const offset = 4 * 4 // 4 uint32 = 16 bytes per vector

	f.LABEL(loop)
	f.CMP(4, n)
	f.BLT(lastBlock)

	// Load 4 vectors (16 elements) from a and b
	f.VLD1_P_Multi(64, aPtr, a0, a1, a2, a3)
	f.VLD1_P_Multi(64, bPtr, b0, b1, b2, b3)

	// Add: b = a + b
	f.VADD(a0.S4(), b0.S4(), b0.S4())
	f.VADD(a1.S4(), b1.S4(), b1.S4())
	f.VADD(a2.S4(), b2.S4(), b2.S4())
	f.VADD(a3.S4(), b3.S4(), b3.S4())

	// Sub: a = b - q (reuse a registers as temp)
	f.VSUB(q.S4(), b0.S4(), a0.S4())
	f.VSUB(q.S4(), b1.S4(), a1.S4())
	f.VSUB(q.S4(), b2.S4(), a2.S4())
	f.VSUB(q.S4(), b3.S4(), a3.S4())

	// Min: b = min(a, b) = min(b-q, b)
	f.VUMIN(a0.S4(), b0.S4(), b0.S4())
	f.VUMIN(a1.S4(), b1.S4(), b1.S4())
	f.VUMIN(a2.S4(), b2.S4(), b2.S4())
	f.VUMIN(a3.S4(), b3.S4(), b3.S4())

	// Store 4 vectors
	f.VST1_P_Multi(64, resPtr, b0, b1, b2, b3)

	// decrement n by 4
	f.SUB(4, n, n)
	f.JMP(loop)

	// Handle remaining 0-3 blocks
	f.LABEL(lastBlock)
	f.Loop(n, func() {
		f.VLD1_P(offset, aPtr, a0.S4())
		f.VLD1_P(offset, bPtr, b0.S4())

		f.VADD(a0.S4(), b0.S4(), b0.S4(), "b = a + b")
		f.VSUB(q.S4(), b0.S4(), a0.S4(), "a = b - q")
		f.VUMIN(a0.S4(), b0.S4(), b0.S4(), "b = min(a, b)")
		f.VST1_P(b0.S4(), resPtr, offset, "res = b")
	})

	registers.Push(resPtr, aPtr, bPtr, n)
	registers.PushV(a0, a1, a2, a3, b0, b1, b2, b3, q)

	f.RET()

}

func (f *FFArm64) generateMulVecF31() {
	f.Comment("mulVec(res, a, b *Element, n uint64)")
	f.Comment("n is the number of blocks of 4 uint32 to process")
	registers := f.FnHeader("mulVec", 0, 32)
	defer f.AssertCleanStack(0, 0)

	// registers
	resPtr := registers.Pop()
	aPtr := registers.Pop()
	bPtr := registers.Pop()
	n := registers.Pop()

	// labels
	// loop := f.NewLabel("loop")
	// done := f.NewLabel("done")

	// load arguments
	f.LDP("res+0(FP)", resPtr, aPtr)
	f.LDP("b+16(FP)", bPtr, n)

	// Explicit registers
	a := arm64.V0
	b := arm64.V1
	cLow := arm64.V2
	cHigh := arm64.V3
	q := arm64.V4
	mLow := arm64.V5
	mHigh := arm64.V6
	p := arm64.V7
	mu := arm64.V8
	zero := arm64.V9
	mask := arm64.V10
	corr := arm64.V11
	temp := arm64.V12

	// Load constants
	// P
	f.VMOVS("$const_q", p)
	f.VDUP(p.SAt(0), p.S4(), "broadcast P")

	// MU = 0x81000001
	tmp := registers.Pop()
	f.MOVD("$const_mu", tmp)
	f.VDUP(tmp, mu.S4(), "broadcast MU")
	registers.Push(tmp)

	// Zero
	f.VMOVQ_cst(0, 0, zero)

	f.Loop(n, func() {
		const offset = 4 * 4 // we process 4 uint32 at a time

		f.VLD1_P(offset, aPtr, a.S4())
		f.VLD1_P(offset, bPtr, b.S4())

		// C = a * b
		f.VUMULL(a, b, cLow, "cLow = a * b (lower halves)")
		f.VUMULL2(a, b, cHigh, "cHigh = a * b (upper halves)")

		// Q = (a * b * MU) mod 2^32
		f.VMUL_S4(a, b, temp, "temp = a * b (low 32 bits)")
		f.VMUL_S4(temp, mu, q, "q = temp * mu (low 32 bits)")

		// M = Q * P
		f.VUMULL(q, p, mLow, "mLow = q * p (lower halves)")
		f.VUMULL2(q, p, mHigh, "mHigh = q * p (upper halves)")

		// X = C - M
		f.VSUB(mLow.D2(), cLow.D2(), cLow.D2(), "cLow = cLow - mLow")
		f.VSUB(mHigh.D2(), cHigh.D2(), cHigh.D2(), "cHigh = cHigh - mHigh")

		// D = X >> 32 (take high parts using UZP2)
		f.VUZP2(cLow, cHigh, a, "a = high 32 bits of [cLow, cHigh]")

		// Correction: if D < 0 (signed comparison with 0)
		f.VCMGT(zero, a, mask, "mask = (0 > a) ? all 1s : 0")

		// corr = mask & P
		f.VAND(p.B16(), mask.B16(), corr.B16(), "corr = mask & P")

		// res = D + corr
		f.VADD(a.S4(), corr.S4(), a.S4())

		f.VST1_P(a.S4(), resPtr, offset, "res = a")
	})

	registers.Push(resPtr, aPtr, bPtr, n)

	f.RET()

}

func (f *FFArm64) generateSubVecF31() {
	f.Comment("subVec(res, a, b *Element, n uint64)")
	f.Comment("n is the number of blocks of 4 uint32 to process")
	registers := f.FnHeader("subVec", 0, 32)
	defer f.AssertCleanStack(0, 0)

	// registers
	resPtr := registers.Pop()
	aPtr := registers.Pop()
	bPtr := registers.Pop()
	n := registers.Pop()

	// labels
	loop := f.NewLabel("loop")
	lastBlock := f.NewLabel("lastBlock")

	// load arguments
	f.LDP("res+0(FP)", resPtr, aPtr)
	f.LDP("b+16(FP)", bPtr, n)

	a0 := registers.PopV()
	a1 := registers.PopV()
	a2 := registers.PopV()
	a3 := registers.PopV()
	b0 := registers.PopV()
	b1 := registers.PopV()
	b2 := registers.PopV()
	b3 := registers.PopV()
	q := registers.PopV()

	f.VMOVS("$const_q", q)
	f.VDUP(q.SAt(0), q.S4(), "broadcast q into "+string(q))

	f.LABEL(loop)
	f.CMP(4, n)
	f.BLT(lastBlock)

	const offset = 4 * 4 // we process 4 uint32 at a time

	// Load 4 vectors (16 elements) from a and b
	f.VLD1_P_Multi(64, aPtr, a0, a1, a2, a3)
	f.VLD1_P_Multi(64, bPtr, b0, b1, b2, b3)

	// b = a - b
	f.VSUB(b0.S4(), a0.S4(), b0.S4())
	f.VSUB(b1.S4(), a1.S4(), b1.S4())
	f.VSUB(b2.S4(), a2.S4(), b2.S4())
	f.VSUB(b3.S4(), a3.S4(), b3.S4())

	// t = b + q (store in a)
	f.VADD(b0.S4(), q.S4(), a0.S4())
	f.VADD(b1.S4(), q.S4(), a1.S4())
	f.VADD(b2.S4(), q.S4(), a2.S4())
	f.VADD(b3.S4(), q.S4(), a3.S4())

	// b = min(t, b) = min(a, b)
	f.VUMIN(a0.S4(), b0.S4(), b0.S4())
	f.VUMIN(a1.S4(), b1.S4(), b1.S4())
	f.VUMIN(a2.S4(), b2.S4(), b2.S4())
	f.VUMIN(a3.S4(), b3.S4(), b3.S4())

	// Store
	f.VST1_P_Multi(64, resPtr, b0, b1, b2, b3)

	// decrement n
	f.SUB(4, n, n)
	f.JMP(loop)

	f.LABEL(lastBlock)
	f.Loop(n, func() {
		f.VLD1_P(offset, aPtr, a0.S4())
		f.VLD1_P(offset, bPtr, b0.S4())

		f.VSUB(b0.S4(), a0.S4(), b0.S4(), "b = a - b")
		f.VADD(b0.S4(), q.S4(), a0.S4(), "t = b + q")
		f.VUMIN(a0.S4(), b0.S4(), b0.S4(), "b = min(t, b)")
		f.VST1_P(b0.S4(), resPtr, offset, "res = b")
	})

	registers.Push(resPtr, aPtr, bPtr, n)
	registers.PushV(a0, a1, a2, a3, b0, b1, b2, b3, q)

	f.RET()

}

func (f *FFArm64) generateSumVecF31() {
	f.Comment("sumVec(t *uint64, a *[]uint32, n uint64) res = sum(a[0...n])")
	f.Comment("n is the number of blocks of 16 uint32 to process")
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
	lastBlock := f.NewLabel("lastBlock")

	// load arguments
	f.LDP("t+0(FP)", tPtr, aPtr)
	f.MOVD("n+16(FP)", n)

	f.LABEL(loop)
	f.CMP(4, n)
	f.BLT(lastBlock)

	f.Comment("blockSize is 16 uint32; we load 4 vectors of 4 uint32 at a time")
	f.Comment("(4*4)*4 = 64 bytes ~= 1 cache line")
	f.Comment("since our values are 31 bits, we can add 2 by 2 these vectors")
	f.Comment("we are left with 2 vectors of 4x32 bits values")
	f.Comment("that we accumulate in 4*2*64bits accumulators")
	f.Comment("the caller will reduce mod q the accumulators.")

	const offset = 8 * 4

	// Unrolled loop body (4x)
	// Block 1
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

	// Block 2
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

	// Block 3
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

	// Block 4
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
	f.SUB(4, n, n)
	f.JMP(loop)

	f.LABEL(lastBlock)
	f.Loop(n, func() {
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
	})

	f.VADD(acc1, acc3, acc1, "acc1 += acc3")
	f.VADD(acc2, acc4, acc2, "acc2 += acc4")

	f.VST2_P(acc1, acc2, tPtr, 0, "store acc1 and acc2")

	f.RET()

}

func (f *FFArm64) generateScalarMulVecF31() {
	f.Comment("scalarMulVec(res, a, b *Element, n uint64) res[0...n] = a[0...n] * b")
	f.Comment("n is the number of blocks of 4 uint32 to process")
	registers := f.FnHeader("scalarMulVec", 0, 32)
	defer f.AssertCleanStack(0, 0)

	// registers
	resPtr := registers.Pop()
	aPtr := registers.Pop()
	bPtr := registers.Pop()
	n := registers.Pop()

	// labels
	// loop := f.NewLabel("loop")
	// done := f.NewLabel("done")

	// load arguments
	f.LDP("res+0(FP)", resPtr, aPtr)
	f.LDP("b+16(FP)", bPtr, n)

	// Explicit registers for Montgomery multiplication
	a := arm64.V0
	b := arm64.V1
	cLow := arm64.V2
	cHigh := arm64.V3
	q := arm64.V4
	mLow := arm64.V5
	mHigh := arm64.V6
	p := arm64.V7
	mu := arm64.V8
	zero := arm64.V9
	mask := arm64.V10
	corr := arm64.V11
	temp := arm64.V12

	// Load constants
	// P
	f.VMOVS("$const_q", p)
	f.VDUP(p.SAt(0), p.S4(), "broadcast P")

	// MU
	tmp := registers.Pop()
	f.MOVD("$const_mu", tmp)
	f.VDUP(tmp, mu.S4(), "broadcast MU")

	// Load scalar b and broadcast
	f.MOVWU(bPtr.At2(0), tmp)
	f.VDUP(tmp, b.S4(), "broadcast scalar b")
	registers.Push(tmp)

	// Zero
	f.VMOVQ_cst(0, 0, zero)

	f.Loop(n, func() {
		const offset = 4 * 4 // we process 4 uint32 at a time

		f.VLD1_P(offset, aPtr, a.S4())

		// C = a * b
		f.VUMULL(a, b, cLow, "cLow = a * b (lower halves)")
		f.VUMULL2(a, b, cHigh, "cHigh = a * b (upper halves)")

		// Q = (a * b * MU) mod 2^32
		f.VMUL_S4(a, b, temp, "temp = a * b (low 32 bits)")
		f.VMUL_S4(temp, mu, q, "q = temp * mu (low 32 bits)")

		// M = Q * P
		f.VUMULL(q, p, mLow, "mLow = q * p (lower halves)")
		f.VUMULL2(q, p, mHigh, "mHigh = q * p (upper halves)")

		// X = C - M
		f.VSUB(mLow.D2(), cLow.D2(), cLow.D2(), "cLow = cLow - mLow")
		f.VSUB(mHigh.D2(), cHigh.D2(), cHigh.D2(), "cHigh = cHigh - mHigh")

		// D = X >> 32 (take high parts using UZP2)
		f.VUZP2(cLow, cHigh, a, "a = high 32 bits of [cLow, cHigh]")

		// Correction: if D < 0 (signed comparison with 0)
		f.VCMGT(zero, a, mask, "mask = (0 > a) ? all 1s : 0")

		// corr = mask & P
		f.VAND(p.B16(), mask.B16(), corr.B16(), "corr = mask & P")

		// res = D + corr
		f.VADD(a.S4(), corr.S4(), a.S4())

		f.VST1_P(a.S4(), resPtr, offset, "res = a")
	})

	registers.Push(resPtr, aPtr, bPtr, n)

	f.RET()
}

func (f *FFArm64) generateInnerProdVecF31() {
	f.Comment("innerProdVec(t *uint64, a, b *[]uint32, n uint64) res = sum(a[0...n] * b[0...n])")
	f.Comment("n is the number of blocks of 4 uint32 to process")
	f.Comment("We do most of the montgomery multiplication but accumulate the")
	f.Comment("temporary result (without final reduction) and let the caller reduce.")
	registers := f.FnHeader("innerProdVec", 0, 32)
	defer f.AssertCleanStack(0, 0)

	// registers
	tPtr := registers.Pop()
	aPtr := registers.Pop()
	bPtr := registers.Pop()
	n := registers.Pop()

	// labels
	// loop := f.NewLabel("loop")
	// done := f.NewLabel("done")

	// load arguments
	f.LDP("t+0(FP)", tPtr, aPtr)
	f.LDP("b+16(FP)", bPtr, n)

	// Explicit registers for Montgomery multiplication
	a := arm64.V0
	b := arm64.V1
	cLow := arm64.V2
	cHigh := arm64.V3
	q := arm64.V4
	mLow := arm64.V5
	mHigh := arm64.V6
	p := arm64.V7
	mu := arm64.V8
	temp := arm64.V12

	// Accumulators (64-bit)
	acc0 := arm64.V9.D2()
	acc1 := arm64.V10.D2()

	// Load constants
	// P
	f.VMOVS("$const_q", p)
	f.VDUP(p.SAt(0), p.S4(), "broadcast P")

	// MU
	tmp := registers.Pop()
	f.MOVD("$const_mu", tmp)
	f.VDUP(tmp, mu.S4(), "broadcast MU")
	registers.Push(tmp)

	// Zero accumulators
	f.VMOVQ_cst(0, 0, arm64.V9)
	f.VMOVQ_cst(0, 0, arm64.V10)

	// Zero register for correction
	zero := arm64.V11
	f.VEOR(zero.B16(), zero.B16(), zero.B16(), "zero = 0")

	f.Loop(n, func() {
		const offset = 4 * 4 // we process 4 uint32 at a time

		f.VLD1_P(offset, aPtr, a.S4())
		f.VLD1_P(offset, bPtr, b.S4())

		// C = a * b (full 64-bit product)
		f.VUMULL(a, b, cLow, "cLow = a * b (lower halves)")
		f.VUMULL2(a, b, cHigh, "cHigh = a * b (upper halves)")

		// Q = (a * b * MU) mod 2^32
		f.VMUL_S4(a, b, temp, "temp = a * b (low 32 bits)")
		f.VMUL_S4(temp, mu, q, "q = temp * mu (low 32 bits)")

		// M = Q * P
		f.VUMULL(q, p, mLow, "mLow = q * p (lower halves)")
		f.VUMULL2(q, p, mHigh, "mHigh = q * p (upper halves)")

		// X = C - M (Montgomery reduction step)
		f.VSUB(mLow.D2(), cLow.D2(), cLow.D2(), "cLow = cLow - mLow")
		f.VSUB(mHigh.D2(), cHigh.D2(), cHigh.D2(), "cHigh = cHigh - mHigh")

		// Extract high 32 bits
		// cLow = [L0, H0, L1, H1], cHigh = [L2, H2, L3, H3]
		// VUZP2(cLow, cHigh) -> [H0, H1, H2, H3]
		// We store it in cLow (reusing register)
		f.VUZP2(cLow.S4(), cHigh.S4(), cLow.S4(), "cLow = high 32 bits of [cLow, cHigh]")

		// Correction: if D < 0 (signed comparison with 0)
		mask := arm64.V13
		f.VCMGT(zero.S4(), cLow.S4(), mask.S4(), "mask = (0 > cLow) ? all 1s : 0")

		// corr = mask & P
		corr := arm64.V14
		f.VAND(p.B16(), mask.B16(), corr.B16(), "corr = mask & P")

		// res = D + corr
		f.VADD(cLow.S4(), corr.S4(), cLow.S4(), "cLow = cLow + corr")

		// Extend to 64 bits and accumulate
		// We need two 64-bit registers. We can reuse mLow and mHigh as temps.
		f.VUSHLL(0, cLow.S2(), mLow.D2(), "mLow = extend(cLow[0,1])")
		f.VUSHLL2(0, cLow.S4(), mHigh.D2(), "mHigh = extend(cLow[2,3])")

		// Accumulate
		f.VADD(mLow.D2(), acc0, acc0, "acc0 += mLow")
		f.VADD(mHigh.D2(), acc1, acc1, "acc1 += mHigh")
	})

	// Combine accumulators
	f.VADD(acc0, acc1, acc0, "acc0 += acc1")

	// Store result
	f.VST1_P(acc0, tPtr, 16, "store accumulator")

	registers.Push(tPtr, aPtr, bPtr, n)

	f.RET()
}
