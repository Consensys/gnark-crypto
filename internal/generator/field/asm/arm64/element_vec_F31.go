package arm64

import (
	"github.com/consensys/bavard/arm64"
)

func (f *FFArm64) generateAddVecF31() {
	f.Comment("addVec(res, a, b *Element, n uint64)")
	f.Comment("n is the number of blocks of 4 uint32 to process")
	registers := f.FnHeader("addVec", 0, 32)
	defer f.AssertCleanStack(0, 0)

	// registers
	resPtr := registers.Pop()
	aPtr := registers.Pop()
	bPtr := registers.Pop()
	n := registers.Pop()

	// load arguments
	f.LDP("res+0(FP)", resPtr, aPtr)
	f.LDP("b+16(FP)", bPtr, n)

	a0 := registers.PopV()
	b0 := registers.PopV()
	q := registers.PopV()

	f.VMOVS("$const_q", q)
	f.VDUP(q.SAt(0), q.S4(), "broadcast q into "+string(q))

	const offset = 4 * 4 // 4 uint32 = 16 bytes per vector

	f.Loop(n, func() {
		f.VLD1_P(offset, aPtr, a0.S4())
		f.VLD1_P(offset, bPtr, b0.S4())

		f.VADD(a0.S4(), b0.S4(), b0.S4(), "b = a + b")
		f.VSUB(q.S4(), b0.S4(), a0.S4(), "a = b - q")
		f.VUMIN(a0.S4(), b0.S4(), b0.S4(), "b = min(a, b)")
		f.VST1_P(b0.S4(), resPtr, offset, "res = b")
	})

	registers.Push(resPtr, aPtr, bPtr, n)
	registers.PushV(a0, b0, q)

	f.RET()

}

func (f *FFArm64) generateMulVecF31() {
	f.Comment("mulVec(res, a, b *Element, n uint64)")
	f.Comment("n is the number of blocks of 4 uint32 to process")
	f.Comment("")
	f.Comment("Algorithm from plonky3 using SQDMULH for efficient Montgomery reduction:")
	f.Comment("For inputs a, b in [0, P), compute a*b*R^-1 mod P where R = 2^32")
	f.Comment("  1. c_hi = (2 * a * b) >> 32  using SQDMULH")
	f.Comment("  2. q = (a * b * mu) mod 2^32")
	f.Comment("  3. qp_hi = (2 * q * P) >> 32 using SQDMULH")
	f.Comment("  4. d = (c_hi - qp_hi) / 2 using SHSUB")
	f.Comment("  5. if d < 0: d += P")
	registers := f.FnHeader("mulVec", 0, 32)
	defer f.AssertCleanStack(0, 0)

	// registers
	resPtr := registers.Pop()
	aPtr := registers.Pop()
	bPtr := registers.Pop()
	n := registers.Pop()

	// load arguments
	f.LDP("res+0(FP)", resPtr, aPtr)
	f.LDP("b+16(FP)", bPtr, n)

	// Explicit registers for Montgomery multiplication using SQDMULH
	a := arm64.V0
	b := arm64.V1
	cHi := arm64.V2
	q := arm64.V3
	qpHi := arm64.V4
	d := arm64.V5
	p := arm64.V6
	mu := arm64.V7
	underflow := arm64.V8

	// Load constants
	// P (as signed for SQDMULH - values < 2^31 are safe)
	f.VMOVS("$const_q", p)
	f.VDUP(p.SAt(0), p.S4(), "broadcast P")

	// MU
	tmp := registers.Pop()
	f.MOVD("$const_mu", tmp)
	f.VDUP(tmp, mu.S4(), "broadcast MU")
	registers.Push(tmp)

	f.Loop(n, func() {
		const offset = 4 * 4 // we process 4 uint32 at a time

		f.VLD1_P(offset, aPtr, a.S4())
		f.VLD1_P(offset, bPtr, b.S4())

		// Step 1: c_hi = (2 * a * b) >> 32 using SQDMULH
		// SQDMULH computes (2*a*b) >> 32 with signed saturation
		f.VSQDMULH(a, b, cHi, "c_hi = (2*a*b) >> 32")

		// Step 2: q = (a * b * mu) mod 2^32
		// First compute temp = a * b (low 32 bits)
		// Then q = temp * mu (low 32 bits)
		f.VMUL_S4(a, b, q, "q = a * b (low 32 bits)")
		f.VMUL_S4(q, mu, q, "q = q * mu (low 32 bits)")

		// Step 3: qp_hi = (2 * q * P) >> 32 using SQDMULH
		f.VSQDMULH(q, p, qpHi, "qp_hi = (2*q*P) >> 32")

		// Step 4: d = (c_hi - qp_hi) / 2 using SHSUB
		// This computes the halving subtraction which accounts for the 2x factor in SQDMULH
		f.VSHSUB(cHi, qpHi, d, "d = (c_hi - qp_hi) / 2")

		// Step 5: Canonicalize - if d < 0 (negative), add P
		// underflow mask: all 1s if qp_hi > c_hi (which means d is negative), else 0
		f.VCMLT(cHi, qpHi, underflow, "underflow = (c_hi < qp_hi) ? all 1s : 0")

		// d = d - underflow * P = d + P if underflow (since underflow is -1 when true)
		f.VMLS(underflow, p, d, "d = d - underflow * P (adds P when d < 0)")

		f.VST1_P(d.S4(), resPtr, offset, "res = d")
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

	// load arguments
	f.LDP("res+0(FP)", resPtr, aPtr)
	f.LDP("b+16(FP)", bPtr, n)

	a0 := registers.PopV()
	b0 := registers.PopV()
	q := registers.PopV()

	f.VMOVS("$const_q", q)
	f.VDUP(q.SAt(0), q.S4(), "broadcast q into "+string(q))

	const offset = 4 * 4 // we process 4 uint32 at a time

	f.Loop(n, func() {
		f.VLD1_P(offset, aPtr, a0.S4())
		f.VLD1_P(offset, bPtr, b0.S4())

		f.VSUB(b0.S4(), a0.S4(), b0.S4(), "b = a - b")
		f.VADD(b0.S4(), q.S4(), a0.S4(), "t = b + q")
		f.VUMIN(a0.S4(), b0.S4(), b0.S4(), "b = min(t, b)")
		f.VST1_P(b0.S4(), resPtr, offset, "res = b")
	})

	registers.Push(resPtr, aPtr, bPtr, n)
	registers.PushV(a0, b0, q)

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
	acc1V := registers.PopV()
	acc2V := registers.PopV()
	acc3V := registers.PopV()
	acc4V := registers.PopV()

	f.Comment("zeroing accumulators")
	f.VMOVQ_cst(0, 0, acc1V)
	f.VMOVQ_cst(0, 0, acc2V)
	f.VMOVQ_cst(0, 0, acc3V)
	f.VMOVQ_cst(0, 0, acc4V)

	acc1 := acc1V.D2()
	acc2 := acc2V.D2()
	acc3 := acc3V.D2()
	acc4 := acc4V.D2()

	// load arguments
	f.LDP("t+0(FP)", tPtr, aPtr)
	f.MOVD("n+16(FP)", n)

	const offset = 8 * 4

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

	registers.Push(aPtr, tPtr, n)
	registers.PushV(a1, a2, a3, a4, acc1V, acc2V, acc3V, acc4V)

	f.RET()

}

func (f *FFArm64) generateScalarMulVecF31() {
	f.Comment("scalarMulVec(res, a, b *Element, n uint64) res[0...n] = a[0...n] * b")
	f.Comment("n is the number of blocks of 4 uint32 to process")
	f.Comment("")
	f.Comment("Algorithm from plonky3 using SQDMULH for efficient Montgomery reduction")
	registers := f.FnHeader("scalarMulVec", 0, 32)
	defer f.AssertCleanStack(0, 0)

	// registers
	resPtr := registers.Pop()
	aPtr := registers.Pop()
	bPtr := registers.Pop()
	n := registers.Pop()

	// load arguments
	f.LDP("res+0(FP)", resPtr, aPtr)
	f.LDP("b+16(FP)", bPtr, n)

	// Explicit registers for Montgomery multiplication using SQDMULH
	a := arm64.V0
	b := arm64.V1
	cHi := arm64.V2
	q := arm64.V3
	qpHi := arm64.V4
	d := arm64.V5
	p := arm64.V6
	mu := arm64.V7
	underflow := arm64.V8
	muB := arm64.V9 // precomputed mu * b

	// Load constants
	// P (as signed for SQDMULH - values < 2^31 are safe)
	f.VMOVS("$const_q", p)
	f.VDUP(p.SAt(0), p.S4(), "broadcast P")

	// MU
	tmp := registers.Pop()
	f.MOVD("$const_mu", tmp)
	f.VDUP(tmp, mu.S4(), "broadcast MU")

	// Load scalar b and broadcast
	f.MOVWU(bPtr.At2(0), tmp)
	f.VDUP(tmp, b.S4(), "broadcast scalar b")

	// Precompute mu * b for reuse in the loop
	f.VMUL_S4(mu, b, muB, "muB = mu * b (precomputed)")
	registers.Push(tmp)

	f.Loop(n, func() {
		const offset = 4 * 4 // we process 4 uint32 at a time

		f.VLD1_P(offset, aPtr, a.S4())

		// Step 1: c_hi = (2 * a * b) >> 32 using SQDMULH
		f.VSQDMULH(a, b, cHi, "c_hi = (2*a*b) >> 32")

		// Step 2: q = (a * b * mu) mod 2^32 = a * (mu * b) mod 2^32
		// Using precomputed muB = mu * b
		f.VMUL_S4(a, muB, q, "q = a * muB (low 32 bits)")

		// Step 3: qp_hi = (2 * q * P) >> 32 using SQDMULH
		f.VSQDMULH(q, p, qpHi, "qp_hi = (2*q*P) >> 32")

		// Step 4: d = (c_hi - qp_hi) / 2 using SHSUB
		f.VSHSUB(cHi, qpHi, d, "d = (c_hi - qp_hi) / 2")

		// Step 5: Canonicalize - if d < 0 (negative), add P
		f.VCMLT(cHi, qpHi, underflow, "underflow = (c_hi < qp_hi) ? all 1s : 0")

		// d = d - underflow * P = d + P if underflow
		f.VMLS(underflow, p, d, "d = d - underflow * P (adds P when d < 0)")

		f.VST1_P(d.S4(), resPtr, offset, "res = d")
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
