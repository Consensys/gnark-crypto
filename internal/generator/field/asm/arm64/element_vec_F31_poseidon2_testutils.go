package arm64

import (
	"fmt"
	"io"

	"github.com/consensys/bavard/arm64"
	"github.com/consensys/gnark-crypto/internal/generator/field/asm/amd64"
)

// allVectors contains all 32 vector registers for easy indexing
var allVectors = []arm64.VectorRegister{
	arm64.V0, arm64.V1, arm64.V2, arm64.V3, arm64.V4, arm64.V5, arm64.V6, arm64.V7,
	arm64.V8, arm64.V9, arm64.V10, arm64.V11, arm64.V12, arm64.V13, arm64.V14, arm64.V15,
	arm64.V16, arm64.V17, arm64.V18, arm64.V19, arm64.V20, arm64.V21, arm64.V22, arm64.V23,
	arm64.V24, arm64.V25, arm64.V26, arm64.V27, arm64.V28, arm64.V29, arm64.V30, arm64.V31,
}

// GenerateF31Poseidon2TestUtils generates ARM64 NEON test functions for individual
// Poseidon2 subroutines. Each function processes 16 elements (4 NEON vectors of 4 elements).
func GenerateF31Poseidon2TestUtils(w io.Writer, nbBits int, q, qInvNeg uint64, params []amd64.Poseidon2Parameters) error {
	f := NewFFArm64(w, (nbBits+63)/64)

	f.generateTestMul(q, qInvNeg)
	f.generateTestAdd(q)
	f.generateTestSub(q)
	f.generateTestDouble(q)
	f.generateTestSbox(q, qInvNeg)
	f.generateTestMatMul4(q, qInvNeg)
	f.generateTestMatMulExternal(q, qInvNeg)
	f.generateTestMul2ExpNegN(q, qInvNeg)
	f.generateTestHalve(q)
	f.generateTestTriple(q)
	f.generateTestQuadruple(q)
	f.generateTestAddRoundKey(q)
	f.generateTestRoundKeyAccess(q)
	f.generateTestFullRoundKeyLoad(q)
	f.generateTestMatMulInternal(q, qInvNeg)
	f.generateTestFullRound(q, qInvNeg)
	f.generateTestPartialRound(q, qInvNeg)

	return nil
}

// generateTestMul generates a test function for Montgomery multiplication
// func testMul_arm64(a, b, result *[16]uint32)
func (f *FFArm64) generateTestMul(constQ, constQInvNeg uint64) {
	const fnName = "testMul_arm64"
	const argSize = 3 * 8 // three pointers

	registers := f.FnHeader(fnName, 0, argSize)

	addrA := registers.Pop()
	addrB := registers.Pop()
	addrResult := registers.Pop()

	f.MOVD("a+0(FP)", addrA)
	f.MOVD("b+8(FP)", addrB)
	f.MOVD("result+16(FP)", addrResult)

	// Load constants
	qReg := registers.Pop()
	qInvNegReg := registers.Pop()
	f.MOVD(constQ, qReg)
	f.MOVD(constQInvNeg, qInvNegReg)

	vQ := arm64.V0
	vQInvNeg := arm64.V1
	f.VDUP(qReg, vQ.S4())
	f.VDUP(qInvNegReg, vQInvNeg.S4())

	// Process 4 vectors of 4 elements each (16 total)
	// a[0:4], a[4:8], a[8:12], a[12:16]
	aVecs := []arm64.VectorRegister{arm64.V2, arm64.V3, arm64.V4, arm64.V5}
	bVecs := []arm64.VectorRegister{arm64.V6, arm64.V7, arm64.V8, arm64.V9}
	resultVecs := []arm64.VectorRegister{arm64.V10, arm64.V11, arm64.V12, arm64.V13}

	// Scratch registers
	scratch0 := arm64.V30
	scratch1 := arm64.V31
	mulTmp := arm64.V29
	t8 := arm64.V26
	t9 := arm64.V27

	// Load all input vectors
	for i := 0; i < 4; i++ {
		f.WriteLn(fmt.Sprintf("    VLD1.P 16(%s), [%s.S4]", addrA, aVecs[i]))
	}
	for i := 0; i < 4; i++ {
		f.WriteLn(fmt.Sprintf("    VLD1.P 16(%s), [%s.S4]", addrB, bVecs[i]))
	}

	// Multiply each pair
	for i := 0; i < 4; i++ {
		f.mulMontgomery(aVecs[i], bVecs[i], resultVecs[i], vQ, vQInvNeg, scratch0, scratch1, mulTmp, t8, t9)
	}

	// Store results
	// Reset result pointer by subtracting what we didn't advance
	f.MOVD("result+16(FP)", addrResult)
	for i := 0; i < 4; i++ {
		f.WriteLn(fmt.Sprintf("    VST1.P [%s.S4], 16(%s)", resultVecs[i], addrResult))
	}

	f.RET()
}

// mulMontgomery performs Montgomery multiplication of two vectors
func (f *FFArm64) mulMontgomery(a, b, into, vQ, vQInvNeg, scratch0, scratch1, mulTmp, t8, t9 arm64.VectorRegister) {
	an := vRegNum(a)
	bn := vRegNum(b)
	qn := vRegNum(vQ)
	s0n := vRegNum(scratch0)
	s1n := vRegNum(scratch1)
	mn := vRegNum(mulTmp)
	t8n := vRegNum(t8)
	t9n := vRegNum(t9)

	// Step 1: ab = a * b (64-bit widening multiply)
	// UMULL scratch0.2D, a.2S, b.2S (lanes 0,1)
	encUmullLo := uint32(0x2ea0c000) | (bn << 16) | (an << 5) | s0n
	f.WriteLn(fmt.Sprintf("    WORD $0x%08x // UMULL %s.2D, %s.2S, %s.2S", encUmullLo, baseReg(scratch0), baseReg(a), baseReg(b)))

	// UMULL2 scratch1.2D, a.4S, b.4S (lanes 2,3)
	encUmullHi := uint32(0x6ea0c000) | (bn << 16) | (an << 5) | s1n
	f.WriteLn(fmt.Sprintf("    WORD $0x%08x // UMULL2 %s.2D, %s.4S, %s.4S", encUmullHi, baseReg(scratch1), baseReg(a), baseReg(b)))

	// Step 2: Extract ab_lo (low 32 bits of each 64-bit product) using UZP1
	f.VUZP1(scratch0.S4(), scratch1.S4(), mulTmp.S4())

	// Step 3: m = (ab_lo * qInvNeg) mod 2^32
	f.VMUL_S4(mulTmp.S4(), vQInvNeg.S4(), mulTmp.S4())

	// Step 4: Compute m * q (64-bit)
	encMqLo := uint32(0x2ea0c000) | (qn << 16) | (mn << 5) | t8n
	f.WriteLn(fmt.Sprintf("    WORD $0x%08x // UMULL %s.2D, %s.2S, %s.2S", encMqLo, baseReg(t8), baseReg(mulTmp), baseReg(vQ)))

	encMqHi := uint32(0x6ea0c000) | (qn << 16) | (mn << 5) | t9n
	f.WriteLn(fmt.Sprintf("    WORD $0x%08x // UMULL2 %s.2D, %s.4S, %s.4S", encMqHi, baseReg(t9), baseReg(mulTmp), baseReg(vQ)))

	// Step 5: Add ab + m*q (64-bit addition)
	f.VADD(scratch0.D2(), t8.D2(), scratch0.D2())
	f.VADD(scratch1.D2(), t9.D2(), scratch1.D2())

	// Step 6: Extract high 32 bits (>> 32) using UZP2
	f.VUZP2(scratch0.S4(), scratch1.S4(), into.S4())

	// Step 7: Conditional reduction: if t >= q then t -= q
	// Using unsigned min: result = min(t, t - q)
	// When t < q: t - q wraps to large value, so min(t, t-q) = t
	// When t >= q: t - q is small, so min(t, t-q) = t - q
	// VSUB(a, b, dst) computes dst = b - a
	f.VSUB(vQ.S4(), into.S4(), mulTmp.S4()) // mulTmp = into - vQ = t - q
	f.VUMIN(into.S4(), mulTmp.S4(), into.S4())
}

// generateTestAdd generates a test function for modular addition
func (f *FFArm64) generateTestAdd(constQ uint64) {
	const fnName = "testAdd_arm64"
	const argSize = 3 * 8

	registers := f.FnHeader(fnName, 0, argSize)

	addrA := registers.Pop()
	addrB := registers.Pop()
	addrResult := registers.Pop()

	f.MOVD("a+0(FP)", addrA)
	f.MOVD("b+8(FP)", addrB)
	f.MOVD("result+16(FP)", addrResult)

	qReg := registers.Pop()
	f.MOVD(constQ, qReg)
	vQ := arm64.V0
	f.VDUP(qReg, vQ.S4())

	aVecs := []arm64.VectorRegister{arm64.V2, arm64.V3, arm64.V4, arm64.V5}
	bVecs := []arm64.VectorRegister{arm64.V6, arm64.V7, arm64.V8, arm64.V9}
	resultVecs := []arm64.VectorRegister{arm64.V10, arm64.V11, arm64.V12, arm64.V13}
	scratch0 := arm64.V30
	scratch1 := arm64.V31

	for i := 0; i < 4; i++ {
		f.WriteLn(fmt.Sprintf("    VLD1.P 16(%s), [%s.S4]", addrA, aVecs[i]))
	}
	for i := 0; i < 4; i++ {
		f.WriteLn(fmt.Sprintf("    VLD1.P 16(%s), [%s.S4]", addrB, bVecs[i]))
	}

	for i := 0; i < 4; i++ {
		f.VADD(aVecs[i].S4(), bVecs[i].S4(), scratch0.S4())
		// VSUB(a, b, dst) computes dst = b - a
		// We want scratch1 = scratch0 - vQ = (a+b) - q
		f.VSUB(vQ.S4(), scratch0.S4(), scratch1.S4())
		f.VUMIN(scratch0.S4(), scratch1.S4(), resultVecs[i].S4())
	}

	f.MOVD("result+16(FP)", addrResult)
	for i := 0; i < 4; i++ {
		f.WriteLn(fmt.Sprintf("    VST1.P [%s.S4], 16(%s)", resultVecs[i], addrResult))
	}

	f.RET()
}

// generateTestSub generates a test function for modular subtraction
func (f *FFArm64) generateTestSub(constQ uint64) {
	const fnName = "testSub_arm64"
	const argSize = 3 * 8

	registers := f.FnHeader(fnName, 0, argSize)

	addrA := registers.Pop()
	addrB := registers.Pop()
	addrResult := registers.Pop()

	f.MOVD("a+0(FP)", addrA)
	f.MOVD("b+8(FP)", addrB)
	f.MOVD("result+16(FP)", addrResult)

	qReg := registers.Pop()
	f.MOVD(constQ, qReg)
	vQ := arm64.V0
	f.VDUP(qReg, vQ.S4())

	aVecs := []arm64.VectorRegister{arm64.V2, arm64.V3, arm64.V4, arm64.V5}
	bVecs := []arm64.VectorRegister{arm64.V6, arm64.V7, arm64.V8, arm64.V9}
	resultVecs := []arm64.VectorRegister{arm64.V10, arm64.V11, arm64.V12, arm64.V13}
	scratch0 := arm64.V30
	scratch1 := arm64.V31

	for i := 0; i < 4; i++ {
		f.WriteLn(fmt.Sprintf("    VLD1.P 16(%s), [%s.S4]", addrA, aVecs[i]))
	}
	for i := 0; i < 4; i++ {
		f.WriteLn(fmt.Sprintf("    VLD1.P 16(%s), [%s.S4]", addrB, bVecs[i]))
	}

	for i := 0; i < 4; i++ {
		// VSUB(a, b, dst) computes dst = b - a
		// We want scratch0 = a - b
		f.VSUB(bVecs[i].S4(), aVecs[i].S4(), scratch0.S4())
		f.VADD(scratch0.S4(), vQ.S4(), scratch1.S4())
		f.VUMIN(scratch0.S4(), scratch1.S4(), resultVecs[i].S4())
	}

	f.MOVD("result+16(FP)", addrResult)
	for i := 0; i < 4; i++ {
		f.WriteLn(fmt.Sprintf("    VST1.P [%s.S4], 16(%s)", resultVecs[i], addrResult))
	}

	f.RET()
}

// generateTestDouble generates a test function for modular doubling
func (f *FFArm64) generateTestDouble(constQ uint64) {
	const fnName = "testDouble_arm64"
	const argSize = 2 * 8

	registers := f.FnHeader(fnName, 0, argSize)

	addrA := registers.Pop()
	addrResult := registers.Pop()

	f.MOVD("a+0(FP)", addrA)
	f.MOVD("result+8(FP)", addrResult)

	qReg := registers.Pop()
	f.MOVD(constQ, qReg)
	vQ := arm64.V0
	f.VDUP(qReg, vQ.S4())

	aVecs := []arm64.VectorRegister{arm64.V2, arm64.V3, arm64.V4, arm64.V5}
	resultVecs := []arm64.VectorRegister{arm64.V10, arm64.V11, arm64.V12, arm64.V13}
	scratch0 := arm64.V30
	scratch1 := arm64.V31

	for i := 0; i < 4; i++ {
		f.WriteLn(fmt.Sprintf("    VLD1.P 16(%s), [%s.S4]", addrA, aVecs[i]))
	}

	for i := 0; i < 4; i++ {
		f.VSHL("$1", aVecs[i].S4(), scratch0.S4())
		// VSUB(a, b, dst) computes dst = b - a
		// We want scratch1 = scratch0 - vQ = 2a - q
		f.VSUB(vQ.S4(), scratch0.S4(), scratch1.S4())
		f.VUMIN(scratch0.S4(), scratch1.S4(), resultVecs[i].S4())
	}

	f.MOVD("result+8(FP)", addrResult)
	for i := 0; i < 4; i++ {
		f.WriteLn(fmt.Sprintf("    VST1.P [%s.S4], 16(%s)", resultVecs[i], addrResult))
	}

	f.RET()
}

// generateTestSbox generates a test function for the S-box (x^3)
func (f *FFArm64) generateTestSbox(constQ, constQInvNeg uint64) {
	const fnName = "testSbox_arm64"
	const argSize = 2 * 8

	registers := f.FnHeader(fnName, 0, argSize)

	addrA := registers.Pop()
	addrResult := registers.Pop()

	f.MOVD("a+0(FP)", addrA)
	f.MOVD("result+8(FP)", addrResult)

	qReg := registers.Pop()
	qInvNegReg := registers.Pop()
	f.MOVD(constQ, qReg)
	f.MOVD(constQInvNeg, qInvNegReg)

	vQ := arm64.V0
	vQInvNeg := arm64.V1
	f.VDUP(qReg, vQ.S4())
	f.VDUP(qInvNegReg, vQInvNeg.S4())

	aVecs := []arm64.VectorRegister{arm64.V2, arm64.V3, arm64.V4, arm64.V5}
	resultVecs := []arm64.VectorRegister{arm64.V10, arm64.V11, arm64.V12, arm64.V13}
	scratch0 := arm64.V30
	scratch1 := arm64.V31
	mulTmp := arm64.V29
	t8 := arm64.V26
	t9 := arm64.V27
	tmp := arm64.V28

	for i := 0; i < 4; i++ {
		f.WriteLn(fmt.Sprintf("    VLD1.P 16(%s), [%s.S4]", addrA, aVecs[i]))
	}

	// sbox: x^3 = x * x * x
	for i := 0; i < 4; i++ {
		// tmp = x^2
		f.mulMontgomery(aVecs[i], aVecs[i], tmp, vQ, vQInvNeg, scratch0, scratch1, mulTmp, t8, t9)
		// result = x * x^2 = x^3
		f.mulMontgomery(aVecs[i], tmp, resultVecs[i], vQ, vQInvNeg, scratch0, scratch1, mulTmp, t8, t9)
	}

	f.MOVD("result+8(FP)", addrResult)
	for i := 0; i < 4; i++ {
		f.WriteLn(fmt.Sprintf("    VST1.P [%s.S4], 16(%s)", resultVecs[i], addrResult))
	}

	f.RET()
}

// generateTestMatMul4 generates a test function for the 4x4 circulant matrix
func (f *FFArm64) generateTestMatMul4(constQ, constQInvNeg uint64) {
	const fnName = "testMatMul4_arm64"
	const argSize = 2 * 8

	registers := f.FnHeader(fnName, 0, argSize)

	addrA := registers.Pop()
	addrResult := registers.Pop()

	f.MOVD("a+0(FP)", addrA)
	f.MOVD("result+8(FP)", addrResult)

	qReg := registers.Pop()
	f.MOVD(constQ, qReg)
	vQ := arm64.V0
	f.VDUP(qReg, vQ.S4())

	// Load 4 vectors (v0, v1, v2, v3) - treating input as 4x4 matrix
	v := []arm64.VectorRegister{arm64.V2, arm64.V3, arm64.V4, arm64.V5}
	t := []arm64.VectorRegister{arm64.V10, arm64.V11, arm64.V12, arm64.V13, arm64.V14, arm64.V15}
	scratch0 := arm64.V30
	scratch1 := arm64.V31

	for i := 0; i < 4; i++ {
		f.WriteLn(fmt.Sprintf("    VLD1.P 16(%s), [%s.S4]", addrA, v[i]))
	}

	// Helper functions
	// VSUB(a, b, dst) computes dst = b - a
	add := func(a, b, into arm64.VectorRegister) {
		f.VADD(a.S4(), b.S4(), scratch0.S4())
		f.VSUB(vQ.S4(), scratch0.S4(), scratch1.S4()) // scratch1 = scratch0 - vQ
		f.VUMIN(scratch0.S4(), scratch1.S4(), into.S4())
	}

	double := func(a, into arm64.VectorRegister) {
		f.VSHL("$1", a.S4(), scratch0.S4())
		f.VSUB(vQ.S4(), scratch0.S4(), scratch1.S4()) // scratch1 = scratch0 - vQ
		f.VUMIN(scratch0.S4(), scratch1.S4(), into.S4())
	}

	// matMul4 algorithm from Plonky3
	// t[0] = v0 + v1
	add(v[0], v[1], t[0])
	// t[1] = v2 + v3
	add(v[2], v[3], t[1])
	// t[2] = t[0] + t[1] = v0+v1+v2+v3
	add(t[0], t[1], t[2])
	// t[3] = t[2] + v1 = v0+2*v1+v2+v3
	add(t[2], v[1], t[3])
	// t[4] = t[2] + v3 = v0+v1+v2+2*v3
	add(t[2], v[3], t[4])
	// t[5] = 2*v0
	double(v[0], t[5])
	// v3 = t[5] + t[4] = 2*v0+v1+v2+2*v3
	add(t[5], t[4], v[3])
	// t[5] = 2*v2
	double(v[2], t[5])
	// v1 = t[5] + t[3] = v0+2*v1+2*v2+v3
	add(t[5], t[3], v[1])
	// v0 = t[0] + t[3] = 2*v0+2*v1+v2+v3
	add(t[0], t[3], v[0])
	// v2 = t[1] + t[4] = v0+v1+2*v2+2*v3
	add(t[1], t[4], v[2])

	// Store results
	f.MOVD("result+8(FP)", addrResult)
	for i := 0; i < 4; i++ {
		f.WriteLn(fmt.Sprintf("    VST1.P [%s.S4], 16(%s)", v[i], addrResult))
	}

	f.RET()
}

// generateTestMatMulExternal generates a test function for the external 16x16 matrix
// This must match the actual matMulExternal implementation in generatePoseidon2_F31_16x16x512
func (f *FFArm64) generateTestMatMulExternal(constQ, constQInvNeg uint64) {
	const fnName = "testMatMulExternal_arm64"
	const argSize = 2 * 8

	registers := f.FnHeader(fnName, 0, argSize)

	addrA := registers.Pop()
	addrResult := registers.Pop()

	f.MOVD("a+0(FP)", addrA)
	f.MOVD("result+8(FP)", addrResult)

	qReg := registers.Pop()
	f.MOVD(constQ, qReg)
	vQ := arm64.V0
	f.VDUP(qReg, vQ.S4())

	// State vectors v[0..15] = V2..V17
	v := make([]arm64.VectorRegister, 16)
	for i := 0; i < 16; i++ {
		v[i] = arm64.VectorRegister(fmt.Sprintf("V%d", i+2))
	}

	// Temporary vectors t[0..4] = V18..V22
	t := []arm64.VectorRegister{arm64.V18, arm64.V19, arm64.V20, arm64.V21, arm64.V22}
	scratch0 := arm64.V30
	scratch1 := arm64.V31

	// Load all 16 vectors
	for i := 0; i < 16; i++ {
		f.WriteLn(fmt.Sprintf("    VLD1.P 16(%s), [%s.S4]", addrA, v[i]))
	}

	// Helper functions matching the main implementation
	add := func(a, b, into arm64.VectorRegister) {
		f.VADD(a.S4(), b.S4(), scratch0.S4())
		f.VSUB(vQ.S4(), scratch0.S4(), scratch1.S4())
		f.VUMIN(scratch0.S4(), scratch1.S4(), into.S4())
	}

	double := func(a, into arm64.VectorRegister) {
		f.VSHL("$1", a.S4(), scratch0.S4())
		f.VSUB(vQ.S4(), scratch0.S4(), scratch1.S4())
		f.VUMIN(scratch0.S4(), scratch1.S4(), into.S4())
	}

	// matMul4 on a slice of 4 vectors - matches the main implementation exactly
	matMul4 := func(s []arm64.VectorRegister) {
		// Algorithm from Plonky3 external.rs
		add(s[0], s[1], t[0]) // t0 = s0 + s1
		add(s[2], s[3], t[1]) // t1 = s2 + s3
		add(t[0], t[1], t[2]) // t2 = s0 + s1 + s2 + s3 (sum)
		add(t[2], s[1], t[3]) // t3 = sum + s1
		add(t[2], s[3], t[4]) // t4 = sum + s3
		double(s[0], s[3])    // new s3 = 2*s0
		add(s[3], t[4], s[3]) // s3 = 2*s0 + sum + s3
		double(s[2], s[1])    // new s1 = 2*s2
		add(s[1], t[3], s[1]) // s1 = 2*s2 + sum + s1
		add(t[0], t[3], s[0]) // s0 = (s0+s1) + (sum+s1)
		add(t[1], t[4], s[2]) // s2 = (s2+s3) + (sum+s3)
	}

	// Apply 4x4 block to each group of 4 state elements
	for i := 0; i < 4; i++ {
		matMul4(v[i*4 : i*4+4])
	}

	// Sum corresponding elements across 4 blocks
	add(v[0], v[4], t[0])
	add(t[0], v[8], t[0])
	add(t[0], v[12], t[0])

	add(v[1], v[5], t[1])
	add(t[1], v[9], t[1])
	add(t[1], v[13], t[1])

	add(v[2], v[6], t[2])
	add(t[2], v[10], t[2])
	add(t[2], v[14], t[2])

	add(v[3], v[7], t[3])
	add(t[3], v[11], t[3])
	add(t[3], v[15], t[3])

	// Add cross-block sum to each element
	for i := 0; i < 16; i++ {
		add(v[i], t[i%4], v[i])
	}

	// Store results
	f.MOVD("result+8(FP)", addrResult)
	for i := 0; i < 16; i++ {
		f.WriteLn(fmt.Sprintf("    VST1.P [%s.S4], 16(%s)", v[i], addrResult))
	}

	f.RET()
}

// generateTestHalve generates a test function for halve (a/2 mod q)
func (f *FFArm64) generateTestHalve(constQ uint64) {
	const fnName = "testHalve_arm64"
	const argSize = 2 * 8

	registers := f.FnHeader(fnName, 0, argSize)

	addrA := registers.Pop()
	addrResult := registers.Pop()

	f.MOVD("a+0(FP)", addrA)
	f.MOVD("result+8(FP)", addrResult)

	qReg := registers.Pop()
	f.MOVD(constQ, qReg)
	vQ := arm64.V0
	f.VDUP(qReg, vQ.S4())

	// Load constant 1
	tmpReg := registers.Pop()
	f.MOVD(1, tmpReg)
	vOneVec := arm64.V28
	f.VDUP(tmpReg, vOneVec.S4())

	aVecs := []arm64.VectorRegister{arm64.V2, arm64.V3, arm64.V4, arm64.V5}
	resultVecs := []arm64.VectorRegister{arm64.V10, arm64.V11, arm64.V12, arm64.V13}
	scratch0 := arm64.V30
	scratch1 := arm64.V31

	for i := 0; i < 4; i++ {
		f.WriteLn(fmt.Sprintf("    VLD1.P 16(%s), [%s.S4]", addrA, aVecs[i]))
	}

	// halve: if a is odd, result = (a + q) >> 1, else result = a >> 1
	// Uses UHADD which computes (a + b) / 2 without overflow
	halve := func(a, into arm64.VectorRegister) {
		// Extract LSB (oddness check) using AND with vOneVec
		f.VAND(a.B16(), vOneVec.B16(), scratch0.B16())
		// Shift bit to MSB position
		f.VSHL("$31", scratch0.S4(), scratch0.S4())
		// Arithmetic shift right to create mask (-1 if odd, 0 if even)
		n := vRegNum(scratch0)
		d := vRegNum(scratch0)
		encoding := uint32(0x4f210400) | (n << 5) | d
		f.WriteLn(fmt.Sprintf("    WORD $0x%08x // SSHR %s.4S, %s.4S, #31", encoding, baseReg(scratch0), baseReg(scratch0)))
		// masked_q = q if odd, 0 if even
		f.VAND(vQ.B16(), scratch0.B16(), scratch1.B16())
		// UHADD Vd.4S, Vn.4S, Vm.4S computes (a + masked_q) / 2 without overflow
		an := vRegNum(a)
		bn := vRegNum(scratch1)
		dn := vRegNum(into)
		uhaddEncoding := uint32(0x6ea00400) | (bn << 16) | (an << 5) | dn
		f.WriteLn(fmt.Sprintf("    WORD $0x%08x // UHADD %s.4S, %s.4S, %s.4S", uhaddEncoding, baseReg(into), baseReg(a), baseReg(scratch1)))
	}

	for i := 0; i < 4; i++ {
		halve(aVecs[i], resultVecs[i])
	}

	f.MOVD("result+8(FP)", addrResult)
	for i := 0; i < 4; i++ {
		f.WriteLn(fmt.Sprintf("    VST1.P [%s.S4], 16(%s)", resultVecs[i], addrResult))
	}

	f.RET()
}

// generateTestTriple generates a test function for triple (3*a mod q)
func (f *FFArm64) generateTestTriple(constQ uint64) {
	const fnName = "testTriple_arm64"
	const argSize = 2 * 8

	registers := f.FnHeader(fnName, 0, argSize)

	addrA := registers.Pop()
	addrResult := registers.Pop()

	f.MOVD("a+0(FP)", addrA)
	f.MOVD("result+8(FP)", addrResult)

	qReg := registers.Pop()
	f.MOVD(constQ, qReg)
	vQ := arm64.V0
	f.VDUP(qReg, vQ.S4())

	aVecs := []arm64.VectorRegister{arm64.V2, arm64.V3, arm64.V4, arm64.V5}
	resultVecs := []arm64.VectorRegister{arm64.V10, arm64.V11, arm64.V12, arm64.V13}
	scratch0 := arm64.V30
	scratch1 := arm64.V31
	tmp := arm64.V29

	for i := 0; i < 4; i++ {
		f.WriteLn(fmt.Sprintf("    VLD1.P 16(%s), [%s.S4]", addrA, aVecs[i]))
	}

	add := func(a, b, into arm64.VectorRegister) {
		f.VADD(a.S4(), b.S4(), scratch0.S4())
		f.VSUB(vQ.S4(), scratch0.S4(), scratch1.S4())
		f.VUMIN(scratch0.S4(), scratch1.S4(), into.S4())
	}

	double := func(a, into arm64.VectorRegister) {
		f.VSHL("$1", a.S4(), scratch0.S4())
		f.VSUB(vQ.S4(), scratch0.S4(), scratch1.S4())
		f.VUMIN(scratch0.S4(), scratch1.S4(), into.S4())
	}

	// triple: 3*a = 2*a + a
	triple := func(a, into arm64.VectorRegister) {
		double(a, tmp)
		add(tmp, a, into)
	}

	for i := 0; i < 4; i++ {
		triple(aVecs[i], resultVecs[i])
	}

	f.MOVD("result+8(FP)", addrResult)
	for i := 0; i < 4; i++ {
		f.WriteLn(fmt.Sprintf("    VST1.P [%s.S4], 16(%s)", resultVecs[i], addrResult))
	}

	f.RET()
}

// generateTestQuadruple generates a test function for quadruple (4*a mod q)
func (f *FFArm64) generateTestQuadruple(constQ uint64) {
	const fnName = "testQuadruple_arm64"
	const argSize = 2 * 8

	registers := f.FnHeader(fnName, 0, argSize)

	addrA := registers.Pop()
	addrResult := registers.Pop()

	f.MOVD("a+0(FP)", addrA)
	f.MOVD("result+8(FP)", addrResult)

	qReg := registers.Pop()
	f.MOVD(constQ, qReg)
	vQ := arm64.V0
	f.VDUP(qReg, vQ.S4())

	aVecs := []arm64.VectorRegister{arm64.V2, arm64.V3, arm64.V4, arm64.V5}
	resultVecs := []arm64.VectorRegister{arm64.V10, arm64.V11, arm64.V12, arm64.V13}
	scratch0 := arm64.V30
	scratch1 := arm64.V31
	tmp := arm64.V29

	for i := 0; i < 4; i++ {
		f.WriteLn(fmt.Sprintf("    VLD1.P 16(%s), [%s.S4]", addrA, aVecs[i]))
	}

	double := func(a, into arm64.VectorRegister) {
		f.VSHL("$1", a.S4(), scratch0.S4())
		f.VSUB(vQ.S4(), scratch0.S4(), scratch1.S4())
		f.VUMIN(scratch0.S4(), scratch1.S4(), into.S4())
	}

	// quadruple: 4*a = 2*(2*a)
	quadruple := func(a, into arm64.VectorRegister) {
		double(a, tmp)
		double(tmp, into)
	}

	for i := 0; i < 4; i++ {
		quadruple(aVecs[i], resultVecs[i])
	}

	f.MOVD("result+8(FP)", addrResult)
	for i := 0; i < 4; i++ {
		f.WriteLn(fmt.Sprintf("    VST1.P [%s.S4], 16(%s)", resultVecs[i], addrResult))
	}

	f.RET()
}

// generateTestAddRoundKey generates a test function for round key addition
// func testAddRoundKey_arm64(state *[64]uint32, roundKey *uint32, result *[64]uint32)
// This tests adding a single broadcasted round key to 16 vectors
func (f *FFArm64) generateTestAddRoundKey(constQ uint64) {
	const fnName = "testAddRoundKey_arm64"
	const argSize = 3 * 8

	registers := f.FnHeader(fnName, 0, argSize)

	addrState := registers.Pop()
	addrKey := registers.Pop()
	addrResult := registers.Pop()

	f.MOVD("state+0(FP)", addrState)
	f.MOVD("roundKey+8(FP)", addrKey)
	f.MOVD("result+16(FP)", addrResult)

	qReg := registers.Pop()
	f.MOVD(constQ, qReg)
	vQ := arm64.V0
	f.VDUP(qReg, vQ.S4())

	// Load round key and broadcast
	scratch0 := arm64.V30
	scratch1 := arm64.V31
	f.WriteLn(fmt.Sprintf("    VLD1R (%s), [%s]", addrKey, scratch0.S4()))

	// State vectors
	v := make([]arm64.VectorRegister, 16)
	for i := 0; i < 16; i++ {
		v[i] = arm64.VectorRegister(fmt.Sprintf("V%d", i+2))
	}

	// Load state
	for i := 0; i < 16; i++ {
		f.WriteLn(fmt.Sprintf("    VLD1.P 16(%s), [%s.S4]", addrState, v[i]))
	}

	// Add round key to each state element
	add := func(a, b, into arm64.VectorRegister) {
		f.VADD(a.S4(), b.S4(), into.S4())
		f.VSUB(vQ.S4(), into.S4(), scratch1.S4())
		f.VUMIN(into.S4(), scratch1.S4(), into.S4())
	}

	for i := 0; i < 16; i++ {
		add(v[i], scratch0, v[i])
	}

	// Store results
	f.MOVD("result+16(FP)", addrResult)
	for i := 0; i < 16; i++ {
		f.WriteLn(fmt.Sprintf("    VST1.P [%s.S4], 16(%s)", v[i], addrResult))
	}

	f.RET()
}

// generateTestRoundKeyAccess tests the round key memory access pattern
// func testRoundKeyAccess_arm64(roundKeys *[27][16]uint32, roundIdx uint64, keyIdx uint64, result *uint32)
// This simulates how round keys are accessed: roundKeys[roundIdx][keyIdx]
// In the actual code, roundKeys is [][]fr.Element with slice headers at 24-byte intervals
func (f *FFArm64) generateTestRoundKeyAccess(constQ uint64) {
	const fnName = "testRoundKeyAccess_arm64"
	const argSize = 4 * 8

	registers := f.FnHeader(fnName, 0, argSize)

	addrRoundKeys := registers.Pop()
	roundIdx := registers.Pop()
	keyIdx := registers.Pop()
	addrResult := registers.Pop()
	tmpCalc := registers.Pop()
	rKeyPtr := registers.Pop()

	f.MOVD("roundKeys+0(FP)", addrRoundKeys)
	f.MOVD("roundIdx+8(FP)", roundIdx)
	f.MOVD("keyIdx+16(FP)", keyIdx)
	f.MOVD("result+24(FP)", addrResult)

	// Compute offset: roundIdx * 24 to get slice header
	// Slice header: [ptr (8), len (8), cap (8)]
	f.WriteLn(fmt.Sprintf("    MOVD $24, %s", tmpCalc))
	f.WriteLn(fmt.Sprintf("    MUL %s, %s, %s", roundIdx, tmpCalc, tmpCalc))
	f.ADD(addrRoundKeys, tmpCalc, tmpCalc)

	// Load data pointer from slice header
	f.WriteLn(fmt.Sprintf("    MOVD (%s), %s", tmpCalc, rKeyPtr))

	// Compute key offset: keyIdx * 4 bytes
	f.WriteLn(fmt.Sprintf("    LSL $2, %s, %s", keyIdx, tmpCalc))
	f.ADD(rKeyPtr, tmpCalc, rKeyPtr)

	// Load and store the key
	f.MOVWU(fmt.Sprintf("(%s)", rKeyPtr), tmpCalc)
	f.MOVWU(tmpCalc, fmt.Sprintf("(%s)", addrResult))

	f.RET()
}

// generateTestFullRoundKeyLoad tests loading all 16 round keys for a full round
// func testFullRoundKeyLoad_arm64(roundKeys *[][]uint32, roundIdx uint64, result *[64]uint32)
func (f *FFArm64) generateTestFullRoundKeyLoad(constQ uint64) {
	const fnName = "testFullRoundKeyLoad_arm64"
	const argSize = 3 * 8

	registers := f.FnHeader(fnName, 0, argSize)

	addrRoundKeys := registers.Pop()
	roundIdxReg := registers.Pop()
	addrResult := registers.Pop()
	tmpCalc := registers.Pop()
	rKeyPtr := registers.Pop()

	f.MOVD("roundKeys+0(FP)", addrRoundKeys)
	f.MOVD("roundIdx+8(FP)", roundIdxReg)
	f.MOVD("result+16(FP)", addrResult)

	qReg := registers.Pop()
	f.MOVD(constQ, qReg)
	vQ := arm64.V0
	f.VDUP(qReg, vQ.S4())

	// Result vectors
	v := make([]arm64.VectorRegister, 16)
	for i := 0; i < 16; i++ {
		v[i] = arm64.VectorRegister(fmt.Sprintf("V%d", i+2))
	}

	// Compute offset: roundIdx * 24 to get slice header
	f.WriteLn(fmt.Sprintf("    MOVD $24, %s", tmpCalc))
	f.WriteLn(fmt.Sprintf("    MUL %s, %s, %s", roundIdxReg, tmpCalc, tmpCalc))
	f.ADD(addrRoundKeys, tmpCalc, tmpCalc)

	// Load data pointer from slice header
	f.WriteLn(fmt.Sprintf("    MOVD (%s), %s", tmpCalc, rKeyPtr))

	// Load all 16 round keys using VLD1R (load and replicate)
	for j := 0; j < 16; j++ {
		f.ADD(uint64(j*4), rKeyPtr, tmpCalc)
		f.WriteLn(fmt.Sprintf("    VLD1R (%s), [%s]", tmpCalc, v[j].S4()))
	}

	// Store results
	for i := 0; i < 16; i++ {
		f.WriteLn(fmt.Sprintf("    VST1.P [%s.S4], 16(%s)", v[i], addrResult))
	}

	f.RET()
}

// generateTestMul2ExpNegN generates test functions for multiplication by 2^(-n)
func (f *FFArm64) generateTestMul2ExpNegN(constQ, constQInvNeg uint64) {
	// Generate test for each shift value used in matMulInternal
	for _, n := range []int{8, 16, 24} {
		f.generateTestMul2ExpNegNSingle(constQ, constQInvNeg, n)
	}
}

func (f *FFArm64) generateTestMul2ExpNegNSingle(constQ, constQInvNeg uint64, n int) {
	fnName := fmt.Sprintf("testMul2ExpNeg%d_arm64", n)
	const argSize = 2 * 8

	registers := f.FnHeader(fnName, 0, argSize)

	addrA := registers.Pop()
	addrResult := registers.Pop()

	f.MOVD("a+0(FP)", addrA)
	f.MOVD("result+8(FP)", addrResult)

	qReg := registers.Pop()
	qInvNegReg := registers.Pop()
	f.MOVD(constQ, qReg)
	f.MOVD(constQInvNeg, qInvNegReg)

	vQ := arm64.V0
	vQInvNeg := arm64.V1
	f.VDUP(qReg, vQ.S4())
	f.VDUP(qInvNegReg, vQInvNeg.S4())

	aVecs := []arm64.VectorRegister{arm64.V2, arm64.V3, arm64.V4, arm64.V5}
	resultVecs := []arm64.VectorRegister{arm64.V10, arm64.V11, arm64.V12, arm64.V13}
	scratch0 := arm64.V30
	scratch1 := arm64.V31
	mulTmp := arm64.V29
	t8 := arm64.V26
	t9 := arm64.V27

	for i := 0; i < 4; i++ {
		f.WriteLn(fmt.Sprintf("    VLD1.P 16(%s), [%s.S4]", addrA, aVecs[i]))
	}

	shift := 32 - n

	for i := 0; i < 4; i++ {
		f.mul2ExpNegNImpl(aVecs[i], resultVecs[i], shift, vQ, vQInvNeg, scratch0, scratch1, mulTmp, t8, t9)
	}

	f.MOVD("result+8(FP)", addrResult)
	for i := 0; i < 4; i++ {
		f.WriteLn(fmt.Sprintf("    VST1.P [%s.S4], 16(%s)", resultVecs[i], addrResult))
	}

	f.RET()
}

// mul2ExpNegNImpl performs a * 2^{-n} mod q using Montgomery reduction
// Reference: v := uint64(x[0]) << (32 - n); z[0] = montReduce(v)
// where montReduce(v) = (v + (uint32(v) * qInvNeg) * q) >> 32, with conditional reduction
func (f *FFArm64) mul2ExpNegNImpl(a, into arm64.VectorRegister, shift int, vQ, vQInvNeg, scratch0, scratch1, mulTmp, t8, t9 arm64.VectorRegister) {
	an := vRegNum(a)
	s0n := vRegNum(scratch0)
	s1n := vRegNum(scratch1)
	mn := vRegNum(mulTmp)
	qn := vRegNum(vQ)

	// v = a << shift (widening to 64-bit)
	// USHLL scratch0.2D, a.2S, #shift
	// Encoding: 0 Q U 011110 0 immh immb 101001 Rn Rd
	// For 32-bit elements: immh:immb = shift + 32
	encLo := uint32(0x2f00a400) | (uint32(32+shift) << 16) | (an << 5) | s0n
	f.WriteLn(fmt.Sprintf("    WORD $0x%08x // USHLL %s.2D, %s.2S, #%d", encLo, baseReg(scratch0), baseReg(a), shift))

	// USHLL2 scratch1.2D, a.4S, #shift
	// Encoding: same as USHLL but with Q=1
	encHi := uint32(0x6f00a400) | (uint32(32+shift) << 16) | (an << 5) | s1n
	f.WriteLn(fmt.Sprintf("    WORD $0x%08x // USHLL2 %s.2D, %s.4S, #%d", encHi, baseReg(scratch1), baseReg(a), shift))

	// Extract low 32 bits (v_lo) into mulTmp
	f.VUZP1(scratch0.S4(), scratch1.S4(), mulTmp.S4())

	// m = v_lo * qInvNeg
	f.VMUL_S4(mulTmp.S4(), vQInvNeg.S4(), mulTmp.S4())

	// m * q (64-bit) -> overwrites scratch0, scratch1
	encMqLo := uint32(0x2ea0c000) | (qn << 16) | (mn << 5) | s0n
	f.WriteLn(fmt.Sprintf("    WORD $0x%08x // UMULL %s.2D, %s.2S, %s.2S", encMqLo, baseReg(scratch0), baseReg(mulTmp), baseReg(vQ)))

	encMqHi := uint32(0x6ea0c000) | (qn << 16) | (mn << 5) | s1n
	f.WriteLn(fmt.Sprintf("    WORD $0x%08x // UMULL2 %s.2D, %s.4S, %s.4S", encMqHi, baseReg(scratch1), baseReg(mulTmp), baseReg(vQ)))

	// Now we need to compute (v + m*q) >> 32
	// Re-compute v = a << shift (we overwrote scratch0/scratch1 with m*q)
	// Actually, we need to do the 64-bit add: v + m*q, then extract high 32 bits
	// Let's re-compute v into t8, t9

	// v = a << shift again into t8/t9
	t8n := vRegNum(t8)
	t9n := vRegNum(t9)
	encVLo := uint32(0x2f00a400) | (uint32(32+shift) << 16) | (an << 5) | t8n
	f.WriteLn(fmt.Sprintf("    WORD $0x%08x // USHLL %s.2D, %s.2S, #%d", encVLo, baseReg(t8), baseReg(a), shift))

	encVHi := uint32(0x6f00a400) | (uint32(32+shift) << 16) | (an << 5) | t9n
	f.WriteLn(fmt.Sprintf("    WORD $0x%08x // USHLL2 %s.2D, %s.4S, #%d", encVHi, baseReg(t9), baseReg(a), shift))

	// 64-bit add: v + m*q
	// scratch0.D2 = scratch0.D2 + t8.D2 (lanes 0,1)
	// scratch1.D2 = scratch1.D2 + t9.D2 (lanes 2,3)
	f.VADD(scratch0.D2(), t8.D2(), scratch0.D2())
	f.VADD(scratch1.D2(), t9.D2(), scratch1.D2())

	// Extract high 32 bits (>> 32) using UZP2
	f.VUZP2(scratch0.S4(), scratch1.S4(), into.S4())

	// Conditional reduction: if t >= q then t -= q
	// VSUB(a, b, dst) computes dst = b - a
	// We want mulTmp = into - vQ
	f.VSUB(vQ.S4(), into.S4(), mulTmp.S4())
	f.VUMIN(into.S4(), mulTmp.S4(), into.S4())
}

// generateTestMatMulInternal generates a test function for the internal matrix multiplication
// func testMatMulInternal_arm64(state *[64]uint32, result *[64]uint32)
func (f *FFArm64) generateTestMatMulInternal(constQ, constQInvNeg uint64) {
	const fnName = "testMatMulInternal_arm64"
	const argSize = 2 * 8

	registers := f.FnHeader(fnName, 0, argSize)

	addrA := registers.Pop()
	addrResult := registers.Pop()

	f.MOVD("a+0(FP)", addrA)
	f.MOVD("result+8(FP)", addrResult)

	qReg := registers.Pop()
	qInvNegReg := registers.Pop()
	f.MOVD(constQ, qReg)
	f.MOVD(constQInvNeg, qInvNegReg)

	vQ := arm64.V0
	vQInvNeg := arm64.V1
	f.VDUP(qReg, vQ.S4())
	f.VDUP(qInvNegReg, vQInvNeg.S4())

	// Load constant 1
	tmpReg := registers.Pop()
	f.MOVD(1, tmpReg)
	vOneVec := arm64.V28
	f.VDUP(tmpReg, vOneVec.S4())

	// State vectors v[0..15] = V2..V17
	v := make([]arm64.VectorRegister, 16)
	for i := 0; i < 16; i++ {
		v[i] = arm64.VectorRegister(fmt.Sprintf("V%d", i+2))
	}

	// Temporary vectors t[0..9] = V18..V27
	t := make([]arm64.VectorRegister, 10)
	for i := 0; i < 10; i++ {
		t[i] = arm64.VectorRegister(fmt.Sprintf("V%d", i+18))
	}

	scratch0 := arm64.V30
	scratch1 := arm64.V31
	mulTmp := arm64.V29

	// Load all 16 vectors
	for i := 0; i < 16; i++ {
		f.WriteLn(fmt.Sprintf("    VLD1.P 16(%s), [%s.S4]", addrA, v[i]))
	}

	// Helper functions
	add := func(a, b, into arm64.VectorRegister) {
		f.VADD(a.S4(), b.S4(), scratch0.S4())
		f.VSUB(vQ.S4(), scratch0.S4(), scratch1.S4())
		f.VUMIN(scratch0.S4(), scratch1.S4(), into.S4())
	}

	sub := func(a, b, into arm64.VectorRegister) {
		f.VSUB(b.S4(), a.S4(), scratch0.S4())
		f.VADD(scratch0.S4(), vQ.S4(), scratch1.S4())
		f.VUMIN(scratch0.S4(), scratch1.S4(), into.S4())
	}

	double := func(a, into arm64.VectorRegister) {
		f.VSHL("$1", a.S4(), scratch0.S4())
		f.VSUB(vQ.S4(), scratch0.S4(), scratch1.S4())
		f.VUMIN(scratch0.S4(), scratch1.S4(), into.S4())
	}

	halve := func(a, into arm64.VectorRegister) {
		f.VAND(a.B16(), vOneVec.B16(), scratch0.B16())
		f.VSHL("$31", scratch0.S4(), scratch0.S4())
		n := vRegNum(scratch0)
		d := vRegNum(scratch0)
		encoding := uint32(0x4f210400) | (n << 5) | d
		f.WriteLn(fmt.Sprintf("    WORD $0x%08x // SSHR %s.4S, %s.4S, #31", encoding, baseReg(scratch0), baseReg(scratch0)))
		f.VAND(vQ.B16(), scratch0.B16(), scratch1.B16())
		an := vRegNum(a)
		bn := vRegNum(scratch1)
		dn := vRegNum(into)
		uhaddEncoding := uint32(0x6ea00400) | (bn << 16) | (an << 5) | dn
		f.WriteLn(fmt.Sprintf("    WORD $0x%08x // UHADD %s.4S, %s.4S, %s.4S", uhaddEncoding, baseReg(into), baseReg(a), baseReg(scratch1)))
	}

	triple := func(a, into arm64.VectorRegister) {
		double(a, scratch0)
		add(scratch0, a, into)
	}

	quadruple := func(a, into arm64.VectorRegister) {
		double(a, scratch0)
		double(scratch0, into)
	}

	mul2ExpNegN := func(a, into arm64.VectorRegister, n int) {
		shift := 32 - n
		an := vRegNum(a)
		s0n := vRegNum(scratch0)
		s1n := vRegNum(scratch1)
		encLo := uint32(0x2f00a400) | (uint32(32+shift) << 16) | (an << 5) | s0n
		f.WriteLn(fmt.Sprintf("    WORD $0x%08x // USHLL %s.2D, %s.2S, #%d", encLo, baseReg(scratch0), baseReg(a), shift))
		encHi := uint32(0x6f00a400) | (uint32(32+shift) << 16) | (an << 5) | s1n
		f.WriteLn(fmt.Sprintf("    WORD $0x%08x // USHLL2 %s.2D, %s.4S, #%d", encHi, baseReg(scratch1), baseReg(a), shift))
		f.VUZP1(scratch0.S4(), scratch1.S4(), mulTmp.S4())
		f.VMUL_S4(mulTmp.S4(), vQInvNeg.S4(), mulTmp.S4())
		mn := vRegNum(mulTmp)
		qn := vRegNum(vQ)
		umullLoEnc := uint32(0x2ea0c000) | (qn << 16) | (mn << 5) | s0n
		f.WriteLn(fmt.Sprintf("    WORD $0x%08x // UMULL %s.2D, %s.2S, %s.2S", umullLoEnc, baseReg(scratch0), baseReg(mulTmp), baseReg(vQ)))
		umullHiEnc := uint32(0x6ea0c000) | (qn << 16) | (mn << 5) | s1n
		f.WriteLn(fmt.Sprintf("    WORD $0x%08x // UMULL2 %s.2D, %s.4S, %s.4S", umullHiEnc, baseReg(scratch1), baseReg(mulTmp), baseReg(vQ)))
		t8n := vRegNum(t[8])
		t9n := vRegNum(t[9])
		encVLo := uint32(0x2f00a400) | (uint32(32+shift) << 16) | (an << 5) | t8n
		f.WriteLn(fmt.Sprintf("    WORD $0x%08x // USHLL %s.2D, %s.2S, #%d", encVLo, baseReg(t[8]), baseReg(a), shift))
		encVHi := uint32(0x6f00a400) | (uint32(32+shift) << 16) | (an << 5) | t9n
		f.WriteLn(fmt.Sprintf("    WORD $0x%08x // USHLL2 %s.2D, %s.4S, #%d", encVHi, baseReg(t[9]), baseReg(a), shift))
		f.VADD(scratch0.D2(), t[8].D2(), scratch0.D2())
		f.VADD(scratch1.D2(), t[9].D2(), scratch1.D2())
		f.VUZP2(scratch0.S4(), scratch1.S4(), into.S4())
		f.VSUB(vQ.S4(), into.S4(), mulTmp.S4())
		f.VUMIN(into.S4(), mulTmp.S4(), into.S4())
	}

	// matMulInternal: compute sum of all elements and apply diagonal
	// Step 1: Compute sum
	add(v[0], v[1], t[0])
	add(v[2], v[3], t[1])
	add(v[4], v[5], t[2])
	add(v[6], v[7], t[3])
	add(t[0], t[1], t[0])
	add(t[2], t[3], t[2])
	add(t[0], t[2], t[0])

	add(v[8], v[9], t[4])
	add(v[10], v[11], t[5])
	add(v[12], v[13], t[6])
	add(v[14], v[15], t[7])
	add(t[4], t[5], t[4])
	add(t[6], t[7], t[6])
	add(t[4], t[6], t[4])

	add(t[0], t[4], t[0]) // t[0] = sum

	// Step 2: Apply diagonal
	// v[0]: diag=-2, sum - 2*v[0]
	double(v[0], v[0])
	sub(t[0], v[0], v[0])

	// v[1]: diag=1, sum + v[1]
	add(t[0], v[1], v[1])

	// v[2]: diag=2, sum + 2*v[2]
	double(v[2], v[2])
	add(t[0], v[2], v[2])

	// v[3]: diag=1/2, sum + v[3]/2
	halve(v[3], v[3])
	add(t[0], v[3], v[3])

	// v[4]: diag=3, sum + 3*v[4]
	triple(v[4], v[4])
	add(t[0], v[4], v[4])

	// v[5]: diag=4, sum + 4*v[5]
	quadruple(v[5], v[5])
	add(t[0], v[5], v[5])

	// v[6]: diag=-1/2, sum - v[6]/2
	halve(v[6], v[6])
	sub(t[0], v[6], v[6])

	// v[7]: diag=-3, sum - 3*v[7]
	triple(v[7], v[7])
	sub(t[0], v[7], v[7])

	// v[8]: diag=-4, sum - 4*v[8]
	quadruple(v[8], v[8])
	sub(t[0], v[8], v[8])

	// v[9]: diag=1/2^8, sum + v[9]/256
	mul2ExpNegN(v[9], v[9], 8)
	add(t[0], v[9], v[9])

	// v[10]: diag=1/8, sum + v[10]/8
	mul2ExpNegN(v[10], v[10], 3)
	add(t[0], v[10], v[10])

	// v[11]: diag=1/2^24, sum + v[11]/2^24
	mul2ExpNegN(v[11], v[11], 24)
	add(t[0], v[11], v[11])

	// v[12]: diag=-1/2^8, sum - v[12]/256
	mul2ExpNegN(v[12], v[12], 8)
	sub(t[0], v[12], v[12])

	// v[13]: diag=-1/8, sum - v[13]/8
	mul2ExpNegN(v[13], v[13], 3)
	sub(t[0], v[13], v[13])

	// v[14]: diag=-1/16, sum - v[14]/16
	mul2ExpNegN(v[14], v[14], 4)
	sub(t[0], v[14], v[14])

	// v[15]: diag=-1/2^24, sum - v[15]/2^24
	mul2ExpNegN(v[15], v[15], 24)
	sub(t[0], v[15], v[15])

	// Store results
	f.MOVD("result+8(FP)", addrResult)
	for i := 0; i < 16; i++ {
		f.WriteLn(fmt.Sprintf("    VST1.P [%s.S4], 16(%s)", v[i], addrResult))
	}

	f.RET()
}

// generateTestFullRound generates a test function for a complete full round:
// add round keys to all 16 elements, apply sbox to all 16, then matMulExternal
// func testFullRound_arm64(state *[64]uint32, roundKeys *[16]uint32, result *[64]uint32)
func (f *FFArm64) generateTestFullRound(constQ, constQInvNeg uint64) {
	const fnName = "testFullRound_arm64"
	const argSize = 3 * 8

	registers := f.FnHeader(fnName, 0, argSize)

	addrState := registers.Pop()
	addrRoundKeys := registers.Pop()
	addrResult := registers.Pop()

	f.MOVD("state+0(FP)", addrState)
	f.MOVD("roundKeys+8(FP)", addrRoundKeys)
	f.MOVD("result+16(FP)", addrResult)

	qReg := registers.Pop()
	qInvNegReg := registers.Pop()
	f.MOVD(constQ, qReg)
	f.MOVD(constQInvNeg, qInvNegReg)

	vQ := arm64.V0
	vQInvNeg := arm64.V1
	f.VDUP(qReg, vQ.S4())
	f.VDUP(qInvNegReg, vQInvNeg.S4())

	// v[0..15] for state - uses V2-V17
	v := make([]arm64.VectorRegister, 16)
	for i := 0; i < 16; i++ {
		v[i] = allVectors[2+i] // V2 through V17
	}

	// Scratch registers - be careful not to overlap
	scratch0 := arm64.V18
	scratch1 := arm64.V19
	scratch2 := arm64.V20 // extra for sbox temps
	scratch3 := arm64.V21 // extra for sbox temps
	mulTmp := arm64.V22
	t8 := arm64.V23
	t9 := arm64.V24
	tmp := registers.Pop()

	// Load state (16 vectors)
	for i := 0; i < 16; i++ {
		f.WriteLn(fmt.Sprintf("    VLD1.P 16(%s), [%s.S4]", addrState, v[i]))
	}

	// Define helper operations
	add := func(a, b, into arm64.VectorRegister) {
		f.VADD(a.S4(), b.S4(), scratch0.S4())
		f.VSUB(vQ.S4(), scratch0.S4(), scratch1.S4())
		f.VUMIN(scratch0.S4(), scratch1.S4(), into.S4())
	}

	double := func(a, into arm64.VectorRegister) {
		f.VSHL("$1", a.S4(), scratch0.S4())
		f.VSUB(vQ.S4(), scratch0.S4(), scratch1.S4())
		f.VUMIN(scratch0.S4(), scratch1.S4(), into.S4())
	}

	mul := func(a, b, into arm64.VectorRegister) {
		f.mulMontgomery(a, b, into, vQ, vQInvNeg, scratch0, scratch1, mulTmp, t8, t9)
	}

	sbox := func(a arm64.VectorRegister) {
		// x^7 = x * (x^2)^3 = x * x^6
		// Use scratch2, scratch3 as temps for sbox (not used by mul)
		mul(a, a, scratch2)               // scratch2 = x^2
		mul(scratch2, scratch2, scratch3) // scratch3 = x^4
		mul(scratch3, scratch2, scratch2) // scratch2 = x^6
		mul(a, scratch2, a)               // a = x^7
	}

	// Step 1: Add round keys to all 16 elements
	// Round keys are stored as 16 consecutive uint32 values
	for j := 0; j < 16; j++ {
		// Load key[j] and broadcast (use scratch3 as it's not being used here)
		f.ADD(uint64(j*4), addrRoundKeys, tmp)
		f.WriteLn(fmt.Sprintf("    VLD1R (%s), [%s]", tmp, scratch3.S4()))
		add(v[j], scratch3, v[j])
	}

	// Step 2: Apply sbox to all 16 elements
	for j := 0; j < 16; j++ {
		sbox(v[j])
	}

	// Step 3: matMulExternal
	// matMul4 on each of 4 blocks
	matMul4 := func(a0, a1, a2, a3 arm64.VectorRegister) {
		// t0 = a0 + a1
		add(a0, a1, scratch0)
		// t1 = a2 + a3
		add(a2, a3, scratch1)
		// t2 = 2*a1 + t1
		double(a1, scratch2)
		add(scratch2, scratch1, scratch2)
		// t3 = 2*a3 + t0
		double(a3, scratch3)
		add(scratch3, scratch0, scratch3)
		// a0 = t0 + t1
		add(scratch0, scratch1, a0)
		// a2 = t2 + t3
		add(scratch2, scratch3, a2)
		// a1 = t3
		f.VMOV(scratch3.B16(), a1.B16())
		// a3 = t2
		f.VMOV(scratch2.B16(), a3.B16())
	}

	// Apply matMul4 to each block
	matMul4(v[0], v[1], v[2], v[3])
	matMul4(v[4], v[5], v[6], v[7])
	matMul4(v[8], v[9], v[10], v[11])
	matMul4(v[12], v[13], v[14], v[15])

	// Cross-block mixing: double each second block
	for i := 4; i < 8; i++ {
		double(v[i], v[i])
	}
	for i := 12; i < 16; i++ {
		double(v[i], v[i])
	}

	// Add across blocks
	add(v[0], v[4], scratch0)
	add(scratch0, v[8], scratch0)
	add(scratch0, v[12], v[0])

	add(v[1], v[5], scratch0)
	add(scratch0, v[9], scratch0)
	add(scratch0, v[13], v[1])

	add(v[2], v[6], scratch0)
	add(scratch0, v[10], scratch0)
	add(scratch0, v[14], v[2])

	add(v[3], v[7], scratch0)
	add(scratch0, v[11], scratch0)
	add(scratch0, v[15], v[3])

	// Add v[0:4] back to v[4:8], v[8:12], v[12:16]
	add(v[0], v[4], v[4])
	add(v[1], v[5], v[5])
	add(v[2], v[6], v[6])
	add(v[3], v[7], v[7])

	add(v[0], v[8], v[8])
	add(v[1], v[9], v[9])
	add(v[2], v[10], v[10])
	add(v[3], v[11], v[11])

	add(v[0], v[12], v[12])
	add(v[1], v[13], v[13])
	add(v[2], v[14], v[14])
	add(v[3], v[15], v[15])

	// Store results
	f.MOVD("result+16(FP)", addrResult)
	for i := 0; i < 16; i++ {
		f.WriteLn(fmt.Sprintf("    VST1.P [%s.S4], 16(%s)", v[i], addrResult))
	}

	f.RET()
}

// generateTestPartialRound generates a test function for a complete partial round:
// add round key to v[0] only, apply sbox to v[0], then matMulInternal
// func testPartialRound_arm64(state *[64]uint32, roundKey *uint32, result *[64]uint32)
func (f *FFArm64) generateTestPartialRound(constQ, constQInvNeg uint64) {
	const fnName = "testPartialRound_arm64"
	const argSize = 3 * 8

	registers := f.FnHeader(fnName, 0, argSize)

	addrState := registers.Pop()
	addrRoundKey := registers.Pop()
	addrResult := registers.Pop()

	f.MOVD("state+0(FP)", addrState)
	f.MOVD("roundKey+8(FP)", addrRoundKey)
	f.MOVD("result+16(FP)", addrResult)

	qReg := registers.Pop()
	qInvNegReg := registers.Pop()
	f.MOVD(constQ, qReg)
	f.MOVD(constQInvNeg, qInvNegReg)

	vQ := arm64.V0
	vQInvNeg := arm64.V1
	f.VDUP(qReg, vQ.S4())
	f.VDUP(qInvNegReg, vQInvNeg.S4())

	// v[0..15] for state - uses V2-V17
	v := make([]arm64.VectorRegister, 16)
	for i := 0; i < 16; i++ {
		v[i] = allVectors[2+i] // V2 through V17
	}

	// t[0..7] for temps - uses V18-V25 (only 8 temps needed for matMulInternal sum)
	t := make([]arm64.VectorRegister, 8)
	for i := 0; i < 8; i++ {
		t[i] = allVectors[18+i] // V18 through V25
	}

	// More scratch registers for mul2ExpNegN and mul
	scratch0 := arm64.V26
	scratch1 := arm64.V27
	mulTmp := arm64.V28
	t8 := arm64.V29      // for mul
	t9 := arm64.V30      // for mul
	sboxTmp := arm64.V31 // for sbox intermediate

	// Load state (16 vectors)
	for i := 0; i < 16; i++ {
		f.WriteLn(fmt.Sprintf("    VLD1.P 16(%s), [%s.S4]", addrState, v[i]))
	}

	// Define helper operations
	add := func(a, b, into arm64.VectorRegister) {
		f.VADD(a.S4(), b.S4(), scratch0.S4())
		f.VSUB(vQ.S4(), scratch0.S4(), scratch1.S4())
		f.VUMIN(scratch0.S4(), scratch1.S4(), into.S4())
	}

	sub := func(a, b, into arm64.VectorRegister) {
		f.VSUB(b.S4(), a.S4(), scratch0.S4())
		f.VADD(scratch0.S4(), vQ.S4(), scratch1.S4())
		f.VUMIN(scratch0.S4(), scratch1.S4(), into.S4())
	}

	double := func(a, into arm64.VectorRegister) {
		f.VSHL("$1", a.S4(), scratch0.S4())
		f.VSUB(vQ.S4(), scratch0.S4(), scratch1.S4())
		f.VUMIN(scratch0.S4(), scratch1.S4(), into.S4())
	}

	mul := func(a, b, into arm64.VectorRegister) {
		f.mulMontgomery(a, b, into, vQ, vQInvNeg, scratch0, scratch1, mulTmp, t8, t9)
	}

	halve := func(a, into arm64.VectorRegister) {
		f.WriteLn(fmt.Sprintf("    VSHL $31, %s.S4, %s.S4", a, scratch0))
		f.WriteLn(fmt.Sprintf("    VUSHR $1, %s.S4, %s.S4", a, into))
		f.WriteLn(fmt.Sprintf("    VUSRA $31, %s.S4, %s.S4", scratch0, into))
	}

	triple := func(a, into arm64.VectorRegister) {
		double(a, sboxTmp)
		add(sboxTmp, a, into)
	}

	quadruple := func(a, into arm64.VectorRegister) {
		double(a, sboxTmp)
		double(sboxTmp, into)
	}

	mul2ExpNegN := func(a, into arm64.VectorRegister, n int) {
		shift := 32 - n
		an := vRegNum(a)
		s0n := vRegNum(scratch0)
		s1n := vRegNum(scratch1)
		encLo := uint32(0x2f00a400) | (uint32(32+shift) << 16) | (an << 5) | s0n
		f.WriteLn(fmt.Sprintf("    WORD $0x%08x // USHLL %s.2D, %s.2S, #%d", encLo, baseReg(scratch0), baseReg(a), shift))
		encHi := uint32(0x6f00a400) | (uint32(32+shift) << 16) | (an << 5) | s1n
		f.WriteLn(fmt.Sprintf("    WORD $0x%08x // USHLL2 %s.2D, %s.4S, #%d", encHi, baseReg(scratch1), baseReg(a), shift))
		f.VUZP1(scratch0.S4(), scratch1.S4(), mulTmp.S4())
		f.VMUL_S4(mulTmp.S4(), vQInvNeg.S4(), mulTmp.S4())
		mn := vRegNum(mulTmp)
		qn := vRegNum(vQ)
		umullLoEnc := uint32(0x2ea0c000) | (qn << 16) | (mn << 5) | s0n
		f.WriteLn(fmt.Sprintf("    WORD $0x%08x // UMULL %s.2D, %s.2S, %s.2S", umullLoEnc, baseReg(scratch0), baseReg(mulTmp), baseReg(vQ)))
		umullHiEnc := uint32(0x6ea0c000) | (qn << 16) | (mn << 5) | s1n
		f.WriteLn(fmt.Sprintf("    WORD $0x%08x // UMULL2 %s.2D, %s.4S, %s.4S", umullHiEnc, baseReg(scratch1), baseReg(mulTmp), baseReg(vQ)))
		t8n := vRegNum(t8)
		t9n := vRegNum(t9)
		encVLo := uint32(0x2f00a400) | (uint32(32+shift) << 16) | (an << 5) | t8n
		f.WriteLn(fmt.Sprintf("    WORD $0x%08x // USHLL %s.2D, %s.2S, #%d", encVLo, baseReg(t8), baseReg(a), shift))
		encVHi := uint32(0x6f00a400) | (uint32(32+shift) << 16) | (an << 5) | t9n
		f.WriteLn(fmt.Sprintf("    WORD $0x%08x // USHLL2 %s.2D, %s.4S, #%d", encVHi, baseReg(t9), baseReg(a), shift))
		f.VADD(scratch0.D2(), t8.D2(), scratch0.D2())
		f.VADD(scratch1.D2(), t9.D2(), scratch1.D2())
		f.VUZP2(scratch0.S4(), scratch1.S4(), into.S4())
		f.VSUB(vQ.S4(), into.S4(), mulTmp.S4())
		f.VUMIN(into.S4(), mulTmp.S4(), into.S4())
	}

	sbox := func(a arm64.VectorRegister) {
		// x^7 = x * (x^2)^3 = x * x^6
		// Use t[0], t[1] as temps since they'll be overwritten later anyway
		mul(a, a, t[0])       // t[0] = x^2
		mul(t[0], t[0], t[1]) // t[1] = x^4
		mul(t[1], t[0], t[0]) // t[0] = x^6
		mul(a, t[0], a)       // a = x^7
	}

	// Step 1: Add round key to v[0] only
	f.WriteLn(fmt.Sprintf("    VLD1R (%s), [%s]", addrRoundKey, sboxTmp.S4()))
	add(v[0], sboxTmp, v[0])

	// Step 2: Apply sbox to v[0] only
	sbox(v[0])

	// Step 3: matMulInternal
	// Compute sum
	add(v[0], v[1], t[0])
	add(v[2], v[3], t[1])
	add(v[4], v[5], t[2])
	add(v[6], v[7], t[3])
	add(t[0], t[1], t[0])
	add(t[2], t[3], t[2])
	add(t[0], t[2], t[0])

	add(v[8], v[9], t[4])
	add(v[10], v[11], t[5])
	add(v[12], v[13], t[6])
	add(v[14], v[15], t[7])
	add(t[4], t[5], t[4])
	add(t[6], t[7], t[6])
	add(t[4], t[6], t[4])

	add(t[0], t[4], t[0]) // t[0] = sum

	// Apply diagonal
	double(v[0], v[0])
	sub(t[0], v[0], v[0])

	add(t[0], v[1], v[1])

	double(v[2], v[2])
	add(t[0], v[2], v[2])

	halve(v[3], v[3])
	add(t[0], v[3], v[3])

	triple(v[4], v[4])
	add(t[0], v[4], v[4])

	quadruple(v[5], v[5])
	add(t[0], v[5], v[5])

	halve(v[6], v[6])
	sub(t[0], v[6], v[6])

	triple(v[7], v[7])
	sub(t[0], v[7], v[7])

	quadruple(v[8], v[8])
	sub(t[0], v[8], v[8])

	mul2ExpNegN(v[9], v[9], 8)
	add(t[0], v[9], v[9])

	mul2ExpNegN(v[10], v[10], 3)
	add(t[0], v[10], v[10])

	mul2ExpNegN(v[11], v[11], 24)
	add(t[0], v[11], v[11])

	mul2ExpNegN(v[12], v[12], 8)
	sub(t[0], v[12], v[12])

	mul2ExpNegN(v[13], v[13], 3)
	sub(t[0], v[13], v[13])

	mul2ExpNegN(v[14], v[14], 4)
	sub(t[0], v[14], v[14])

	mul2ExpNegN(v[15], v[15], 24)
	sub(t[0], v[15], v[15])

	// Store results
	f.MOVD("result+16(FP)", addrResult)
	for i := 0; i < 16; i++ {
		f.WriteLn(fmt.Sprintf("    VST1.P [%s.S4], 16(%s)", v[i], addrResult))
	}

	f.RET()
}
