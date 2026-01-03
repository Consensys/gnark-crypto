package arm64

import (
	"fmt"

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
	done := f.NewLabel("done")
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
	f.WriteLn(fmt.Sprintf("\tBLT %s", lastBlock))

	// Load 4 vectors (16 elements) from a and b
	// Using VLD1.P with multiple registers - format: VLD1.P offset(Rn), [Vt.S4, Vt2.S4, ...]
	f.WriteLn(fmt.Sprintf("\tVLD1.P 64(%s), [%s, %s, %s, %s]", aPtr, a0.S4(), a1.S4(), a2.S4(), a3.S4()))
	f.WriteLn(fmt.Sprintf("\tVLD1.P 64(%s), [%s, %s, %s, %s]", bPtr, b0.S4(), b1.S4(), b2.S4(), b3.S4()))

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
	f.WriteLn(fmt.Sprintf("\tVST1.P [%s, %s, %s, %s], 64(%s)", b0.S4(), b1.S4(), b2.S4(), b3.S4(), resPtr))

	// decrement n by 4
	f.SUB(4, n, n)
	f.JMP(loop)

	// Handle remaining 0-3 blocks
	f.LABEL(lastBlock)
	f.CBZ(n, done)

	f.VLD1_P(offset, aPtr, a0.S4())
	f.VLD1_P(offset, bPtr, b0.S4())

	f.VADD(a0.S4(), b0.S4(), b0.S4(), "b = a + b")
	f.VSUB(q.S4(), b0.S4(), a0.S4(), "a = b - q")
	f.VUMIN(a0.S4(), b0.S4(), b0.S4(), "b = min(a, b)")
	f.VST1_P(b0.S4(), resPtr, offset, "res = b")

	f.SUB(1, n, n)
	f.JMP(lastBlock)

	f.LABEL(done)

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
	loop := f.NewLabel("loop")
	done := f.NewLabel("done")

	// load arguments
	f.LDP("res+0(FP)", resPtr, aPtr)
	f.LDP("b+16(FP)", bPtr, n)

	// Explicit registers
	a := arm64.V0
	b := arm64.V1
	cLow := arm64.V2
	cHigh := arm64.V3
	// q := arm64.V4
	mLow := arm64.V5
	mHigh := arm64.V6
	p := arm64.V7
	mu := arm64.V8
	zero := arm64.V9
	// mask := arm64.V10
	corr := arm64.V11
	// temp := arm64.V12

	// Load constants
	// P
	f.VMOVS("$const_q", p)
	f.VDUP(p.SAt(0), p.S4(), "broadcast P")

	// MU = 0x81000001
	tmp := registers.Pop()
	f.MOVD("$0x81000001", tmp)
	f.VDUP(tmp, mu.S4(), "broadcast MU")
	registers.Push(tmp)

	// Zero
	f.VMOVQ_cst(0, 0, zero)

	f.LABEL(loop)

	f.CBZ(n, done)

	const offset = 4 * 4 // we process 4 uint32 at a time

	f.VLD1_P(offset, aPtr, a.S4())
	f.VLD1_P(offset, bPtr, b.S4())

	// C = a * b
	// UMULL V0.2S, V1.2S, V2.2D
	f.WriteLn("WORD $0x2ea1c002 // UMULL V2.2D, V0.2S, V1.2S")
	// UMULL2 V0.4S, V1.4S, V3.2D
	f.WriteLn("WORD $0x6ea1c003 // UMULL2 V3.2D, V0.4S, V1.4S")

	// Q = (a * b * MU) mod 2^32
	// MUL a, b -> temp (V12)
	f.WriteLn("WORD $0x4ea19c0c // MUL V12.4S, V0.4S, V1.4S")
	// MUL temp, mu -> q (V4)
	f.WriteLn("WORD $0x4ea89d84 // MUL V4.4S, V12.4S, V8.4S")

	// M = Q * P
	// UMULL q.2S, p.2S, mLow.2D
	f.WriteLn("WORD $0x2ea7c085 // UMULL V5.2D, V4.2S, V7.2S")
	// UMULL2 q.4S, p.4S, mHigh.2D
	f.WriteLn("WORD $0x6ea7c086 // UMULL2 V6.2D, V4.4S, V7.4S")

	// X = C - M
	f.VSUB(mLow.D2(), cLow.D2(), cLow.D2())    // X_low
	f.VSUB(mHigh.D2(), cHigh.D2(), cHigh.D2()) // X_high

	// D = X >> 32 (take high parts)
	// UZP2 cLow.4S, cHigh.4S, a.4S (Dest is a)
	f.WriteLn("WORD $0x4e835840 // UZP2 V0.4S, V2.4S, V3.4S")

	// Correction
	// if D < 0 (signed comparison with 0)
	// CMGT zero, a, mask (mask = (0 > a))
	f.WriteLn("WORD $0x4ea0352a // CMGT V10.4S, V9.4S, V0.4S")

	// corr = mask & P
	f.WriteLn("WORD $0x4e2a1ceb // AND V11.16B, V7.16B, V10.16B")

	// res = D + corr
	f.VADD(a.S4(), corr.S4(), a.S4())

	f.VST1_P(a.S4(), resPtr, offset, "res = a")

	// decrement n
	f.SUB(1, n, n)
	f.JMP(loop)

	f.LABEL(done)

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
	done := f.NewLabel("done")
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
	f.WriteLn(fmt.Sprintf("\tBLT %s", lastBlock))

	const offset = 4 * 4 // we process 4 uint32 at a time

	// Load 4 vectors (16 elements) from a and b
	f.WriteLn(fmt.Sprintf("\tVLD1.P 64(%s), [%s, %s, %s, %s]", aPtr, a0.S4(), a1.S4(), a2.S4(), a3.S4()))
	f.WriteLn(fmt.Sprintf("\tVLD1.P 64(%s), [%s, %s, %s, %s]", bPtr, b0.S4(), b1.S4(), b2.S4(), b3.S4()))

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
	f.WriteLn(fmt.Sprintf("\tVST1.P [%s, %s, %s, %s], 64(%s)", b0.S4(), b1.S4(), b2.S4(), b3.S4(), resPtr))

	// decrement n
	f.SUB(4, n, n)
	f.JMP(loop)

	f.LABEL(lastBlock)
	f.CBZ(n, done)

	f.VLD1_P(offset, aPtr, a0.S4())
	f.VLD1_P(offset, bPtr, b0.S4())

	f.VSUB(b0.S4(), a0.S4(), b0.S4(), "b = a - b")
	f.VADD(b0.S4(), q.S4(), a0.S4(), "t = b + q")
	f.VUMIN(a0.S4(), b0.S4(), b0.S4(), "b = min(t, b)")
	f.VST1_P(b0.S4(), resPtr, offset, "res = b")

	// decrement n
	f.SUB(1, n, n)
	f.JMP(lastBlock)

	f.LABEL(done)

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
	done := f.NewLabel("done")
	lastBlock := f.NewLabel("lastBlock")

	// load arguments
	f.LDP("t+0(FP)", tPtr, aPtr)
	f.MOVD("n+16(FP)", n)

	f.LABEL(loop)
	f.CMP(4, n)
	f.WriteLn(fmt.Sprintf("\tBLT %s", lastBlock))

	f.WriteLn(`
	// blockSize is 16 uint32; we load 4 vectors of 4 uint32 at a time
	// (4*4)*4 = 64 bytes ~= 1 cache line
	// since our values are 31 bits, we can add 2 by 2 these vectors
	// we are left with 2 vectors of 4x32 bits values
	// that we accumulate in 4*2*64bits accumulators
	// the caller will reduce mod q the accumulators.
	`)

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
	f.CBZ(n, done)

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
	f.JMP(lastBlock)

	f.LABEL(done)

	f.VADD(acc1, acc3, acc1, "acc1 += acc3")
	f.VADD(acc2, acc4, acc2, "acc2 += acc4")

	f.VST2_P(acc1, acc2, tPtr, 0, "store acc1 and acc2")

	f.RET()

}
