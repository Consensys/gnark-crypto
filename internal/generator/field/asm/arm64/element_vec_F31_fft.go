// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package arm64

// GenerateF31FFTKernels generates the ARM64 FFT kernels for F31 fields
func GenerateF31FFTKernels(f *FFArm64, kernels []int) error {
	f.generateFFTInnerDITF31()
	f.generateFFTInnerDIFF31()
	return nil
}

func (f *FFArm64) generateFFTInnerDITF31() {
	// innerDITWithTwiddles_neon processes 4 elements at a time
	// func innerDITWithTwiddles_neon(a, twiddles *Element, start, end, m int)
	f.Comment("innerDITWithTwiddles_neon(a, twiddles *Element, start, end, m int)")
	f.Comment("Processes 4 elements at a time using NEON")
	f.Comment("For i := start; i < end; i++ { a[i+m].Mul(&a[i+m], &twiddles[i]); Butterfly(&a[i], &a[i+m]) }")
	registers := f.FnHeader("innerDITWithTwiddles_neon", 0, 40)
	defer f.AssertCleanStack(0, 0)

	addrA := registers.Pop()
	addrTwiddles := registers.Pop()
	start := registers.Pop()
	n := registers.Pop()
	m := registers.Pop()
	addrAPlusM := registers.Pop()
	tmp := registers.Pop()

	// Load arguments
	f.LDP("a+0(FP)", addrA, addrTwiddles)
	f.LDP("start+16(FP)", start, n) // start and end
	f.MOVD("m+32(FP)", m)

	// Compute loop count: n = (end - start) / 4
	f.SUB(start, n, n)
	f.WriteLn("\tLSR\t$2, " + string(n) + ", " + string(n))

	// Load constants
	p := registers.PopV()
	mu := registers.PopV()
	f.VMOVS("$const_q", p)
	f.VDUP(p.SAt(0), p.S4(), "broadcast P")
	f.MOVD("$const_mu", tmp)
	f.VDUP(tmp, mu.S4(), "broadcast MU")

	// Advance a and twiddles pointers by start * 4 bytes
	f.WriteLn("\tLSL\t$2, " + string(start) + ", " + string(start))
	f.ADD(start, addrA, addrA)
	f.ADD(start, addrTwiddles, addrTwiddles)

	// offset = m * 4 bytes
	f.WriteLn("\tLSL\t$2, " + string(m) + ", " + string(m))

	f.MOVD(addrA, addrAPlusM)
	f.ADD(m, addrAPlusM, addrAPlusM)

	a := registers.PopV()
	am := registers.PopV()
	tw := registers.PopV()
	t0 := registers.PopV()
	t1 := registers.PopV()
	cHi := registers.PopV()
	qpHi := registers.PopV()
	underflow := registers.PopV()

	lblLoop := f.NewLabel("loop")
	lblDone := f.NewLabel("done")

	f.LABEL(lblLoop)
	f.CBZ(n, lblDone)

	// Load a[i], a[i+m], twiddles
	const offset = 16 // 4 * 4 bytes
	f.VLD1_P(offset, addrA, a.S4())
	f.VLD1_P(offset, addrAPlusM, am.S4())
	f.VLD1_P(offset, addrTwiddles, tw.S4())

	// Montgomery multiply am by twiddle using SQDMULH algorithm
	// 1. c_hi = (2 * am * tw) >> 32
	f.VSQDMULH(am, tw, cHi, "c_hi = (2*am*tw) >> 32")
	// 2. q_lo = am * tw (low 32 bits), then q_lo = q_lo * mu (low 32 bits)
	f.VMUL_S4(am, tw, am, "q_lo = am * tw")
	f.VMUL_S4(am, mu, am, "q_lo = q_lo * mu")
	// 3. qp_hi = (2 * q_lo * P) >> 32
	f.VSQDMULH(am, p, qpHi, "qp_hi = (2*q_lo*P) >> 32")
	// 4. am = (c_hi - qp_hi) / 2
	f.VSHSUB(cHi, qpHi, am, "am = (c_hi - qp_hi) / 2")
	// 5. if c_hi < qp_hi (underflow), add P
	f.VCMLT(cHi, qpHi, underflow, "underflow mask")
	f.VMLS(underflow, p, am, "am += P if underflow")

	// Butterfly: t0 = a + am, am = a - am
	f.VADD(a.S4(), am.S4(), t0.S4(), "t0 = a + am")
	f.VSUB(am.S4(), a.S4(), am.S4(), "am = a - am")
	// Reduce t0: a = min(t0, t0 - p)
	f.VSUB(p.S4(), t0.S4(), a.S4(), "a = t0 - p")
	f.VUMIN(t0.S4(), a.S4(), a.S4(), "a = min(t0, a)")
	// Reduce am: am = min(am, am + p)
	f.VADD(p.S4(), am.S4(), t1.S4(), "t1 = am + p")
	f.VUMIN(t1.S4(), am.S4(), am.S4(), "am = min(t1, am)")

	// Store results (addrA and addrAPlusM were post-incremented, so go back)
	f.WriteLn("\tSUB\t$16, " + string(addrA) + ", " + string(tmp))
	f.VST1_P(a.S4(), tmp, 0, "store a[i]")
	f.WriteLn("\tSUB\t$16, " + string(addrAPlusM) + ", " + string(tmp))
	f.VST1_P(am.S4(), tmp, 0, "store a[i+m]")

	f.WriteLn("\tSUB\t$1, " + string(n) + ", " + string(n))
	f.JMP(lblLoop)

	f.LABEL(lblDone)
	f.RET()

	registers.Push(addrA, addrTwiddles, start, n, m, addrAPlusM, tmp)
	registers.PushV(p, mu, a, am, tw, t0, t1, cHi, qpHi, underflow)
}

func (f *FFArm64) generateFFTInnerDIFF31() {
	// innerDIFWithTwiddles_neon processes 4 elements at a time
	// func innerDIFWithTwiddles_neon(a, twiddles *Element, start, end, m int)
	f.Comment("innerDIFWithTwiddles_neon(a, twiddles *Element, start, end, m int)")
	f.Comment("Processes 4 elements at a time using NEON")
	f.Comment("For i := start; i < end; i++ { Butterfly(&a[i], &a[i+m]); a[i+m].Mul(&a[i+m], &twiddles[i]) }")
	registers := f.FnHeader("innerDIFWithTwiddles_neon", 0, 40)
	defer f.AssertCleanStack(0, 0)

	addrA := registers.Pop()
	addrTwiddles := registers.Pop()
	start := registers.Pop()
	n := registers.Pop()
	m := registers.Pop()
	addrAPlusM := registers.Pop()
	tmp := registers.Pop()

	// Load arguments
	f.LDP("a+0(FP)", addrA, addrTwiddles)
	f.LDP("start+16(FP)", start, n) // start and end
	f.MOVD("m+32(FP)", m)

	// Compute loop count: n = (end - start) / 4
	f.SUB(start, n, n)
	f.WriteLn("\tLSR\t$2, " + string(n) + ", " + string(n))

	// Load constants
	p := registers.PopV()
	mu := registers.PopV()
	f.VMOVS("$const_q", p)
	f.VDUP(p.SAt(0), p.S4(), "broadcast P")
	f.MOVD("$const_mu", tmp)
	f.VDUP(tmp, mu.S4(), "broadcast MU")

	// Advance a and twiddles pointers by start * 4 bytes
	f.WriteLn("\tLSL\t$2, " + string(start) + ", " + string(start))
	f.ADD(start, addrA, addrA)
	f.ADD(start, addrTwiddles, addrTwiddles)

	// offset = m * 4 bytes
	f.WriteLn("\tLSL\t$2, " + string(m) + ", " + string(m))

	f.MOVD(addrA, addrAPlusM)
	f.ADD(m, addrAPlusM, addrAPlusM)

	a := registers.PopV()
	am := registers.PopV()
	tw := registers.PopV()
	t0 := registers.PopV()
	t1 := registers.PopV()
	cHi := registers.PopV()
	qpHi := registers.PopV()
	underflow := registers.PopV()

	lblLoop := f.NewLabel("loop")
	lblDone := f.NewLabel("done")

	f.LABEL(lblLoop)
	f.CBZ(n, lblDone)

	// Load a[i], a[i+m], twiddles
	const offset = 16 // 4 * 4 bytes
	f.VLD1_P(offset, addrA, a.S4())
	f.VLD1_P(offset, addrAPlusM, am.S4())
	f.VLD1_P(offset, addrTwiddles, tw.S4())

	// Butterfly: t0 = a + am, t1 = a - am
	f.VSUB(am.S4(), a.S4(), t1.S4(), "t1 = a - am")
	f.VADD(a.S4(), am.S4(), t0.S4(), "t0 = a + am")
	// Reduce t0: a = min(t0, t0 - p)
	f.VSUB(p.S4(), t0.S4(), a.S4(), "a = t0 - p")
	f.VUMIN(t0.S4(), a.S4(), a.S4(), "a = min(t0, a)")
	// Reduce t1: am = min(t1, t1 + p) - handles negative case
	f.VADD(p.S4(), t1.S4(), am.S4(), "am_temp = t1 + p")
	f.VUMIN(t1.S4(), am.S4(), am.S4(), "am = min(t1, t1+p)")

	// Montgomery multiply am by twiddle
	f.VSQDMULH(am, tw, cHi, "c_hi = (2*am*tw) >> 32")
	f.VMUL_S4(am, tw, am, "q_lo = am * tw")
	f.VMUL_S4(am, mu, am, "q_lo = q_lo * mu")
	f.VSQDMULH(am, p, qpHi, "qp_hi = (2*q_lo*P) >> 32")
	f.VSHSUB(cHi, qpHi, am, "am = (c_hi - qp_hi) / 2")
	f.VCMLT(cHi, qpHi, underflow, "underflow mask")
	f.VMLS(underflow, p, am, "am += P if underflow")

	// Store results
	f.WriteLn("\tSUB\t$16, " + string(addrA) + ", " + string(tmp))
	f.VST1_P(a.S4(), tmp, 0, "store a[i]")
	f.WriteLn("\tSUB\t$16, " + string(addrAPlusM) + ", " + string(tmp))
	f.VST1_P(am.S4(), tmp, 0, "store a[i+m]")

	f.WriteLn("\tSUB\t$1, " + string(n) + ", " + string(n))
	f.JMP(lblLoop)

	f.LABEL(lblDone)
	f.RET()

	registers.Push(addrA, addrTwiddles, start, n, m, addrAPlusM, tmp)
	registers.PushV(p, mu, a, am, tw, t0, t1, cHi, qpHi, underflow)
}
