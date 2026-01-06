package arm64

import (
	"fmt"
	"io"

	"github.com/consensys/bavard/arm64"
	"github.com/consensys/gnark-crypto/internal/generator/field/asm/amd64"
)

// vRegNum extracts the numeric register ID from a VectorRegister (V0 -> 0, V31 -> 31)
func vRegNum(v arm64.VectorRegister) uint32 {
	s := string(v)
	// Remove any suffix like .S4, .D2, etc.
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			s = s[:i]
			break
		}
	}
	// Parse "Vn" where n is 0-31
	if len(s) < 2 || s[0] != 'V' {
		panic("invalid vector register: " + string(v))
	}
	var n uint32
	for i := 1; i < len(s); i++ {
		n = n*10 + uint32(s[i]-'0')
	}
	return n
}

// baseReg returns the base register name (e.g., "V0" from "V0.S4")
func baseReg(v arm64.VectorRegister) string {
	s := string(v)
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			return s[:i]
		}
	}
	return s
}

func GenerateF31Poseidon2(w io.Writer, nbBits int, q, qInvNeg uint64, params []amd64.Poseidon2Parameters) error {
	f := NewFFArm64(w, (nbBits+63)/64)
	for _, p := range params {
		if p.Width == 16 {
			f.generatePoseidon2_F31_16x16x512(p, q, qInvNeg)
		}
	}
	return nil
}

// generatePoseidon2_F31_16x16x512 generates ARM64 NEON assembly for Poseidon2 permutation
// on F31 fields with width=16, processing 4 permutations in parallel using NEON vectors.
//
// Memory layout:
//   - matrix: input data, 16 rows × 512 field elements each (total 16 × 512 × 4 = 32768 bytes)
//   - roundKeys: slice header pointing to [][]fr.Element round keys
//   - result: output buffer, 16 rows × 8 field elements each (total 16 × 8 × 4 = 512 bytes)
//
// Algorithm:
//   - We process 4 rows in parallel (4 NEON lanes), so we need 4 batches to cover all 16 rows
//   - Each batch processes N=64 steps (512 elements / 8 elements per step)
//   - The state v[0..15] holds 16 field elements for each of 4 parallel permutations
//   - Feed-forward: after all steps, state[j] += input[j] for j in [0,8)
//
// Register allocation:
//   - V0: constant q (field modulus broadcast to all lanes)
//   - V1: constant mu (Montgomery constant for reduction)
//   - V2-V17: state vectors (16 vectors, each holds 4 field elements from parallel permutations)
//   - V18-V27: temporary vectors for arithmetic operations
//   - V28: constant 1s for AND operations
//   - V29: used by mul as private temp
//   - V30-V31: scratch for modular arithmetic
//   - R0-R12: general purpose (addresses, counters, etc.)
func (f *FFArm64) generatePoseidon2_F31_16x16x512(params amd64.Poseidon2Parameters, constQ, constQInvNeg uint64) {
	fullRounds := params.FullRounds
	partialRounds := params.PartialRounds
	rf := fullRounds / 2 // half rounds before and after partial rounds

	if params.Width != 16 {
		panic("only width 16 is supported")
	}

	const fnName = "permutation16x16x512_arm64"
	const argSize = 8 + 24 + 8 // matrix ptr + roundKeys slice header + result ptr

	// Stack frame for temporary storage during each step (8 vectors × 16 bytes = 128 bytes)
	const stackSize = 128
	registers := f.FnHeader(fnName, stackSize, argSize)

	// =========================================================================
	// Register Allocation
	// =========================================================================

	// General purpose registers for addresses and counters
	addrMatrix := registers.Pop()    // base address of input matrix
	addrRoundKeys := registers.Pop() // address of roundKeys slice header
	addrResult := registers.Pop()    // base address of result buffer

	f.MOVD("matrix+0(FP)", addrMatrix)
	f.MOVD("roundKeys+8(FP)", addrRoundKeys)
	f.MOVD("result+32(FP)", addrResult)

	// Constants in scalar registers for VDUP
	qReg := registers.Pop()       // field modulus q
	qInvNegReg := registers.Pop() // qInvNeg = -q^{-1} mod 2^32

	f.MOVD(constQ, qReg)
	f.MOVD(constQInvNeg, qInvNegReg)

	// Vector constants (broadcast to all 4 lanes)
	vQ := arm64.V0  // q broadcast
	vMu := arm64.V1 // qInvNeg broadcast (misnamed vMu for historical reasons)
	f.VDUP(qReg, vQ.S4())
	f.VDUP(qInvNegReg, vMu.S4())

	// Load constant 1 for LSB extraction in halve operation
	tmpReg := registers.Pop()
	f.MOVD(1, tmpReg)
	vOneVec := arm64.V28 // V28 = {1, 1, 1, 1} for AND operations
	f.VDUP(tmpReg, vOneVec.S4())

	// State vectors: v[0..15] = V2..V17
	// Each vector holds 4 field elements from 4 parallel permutations
	v := []arm64.VectorRegister{
		arm64.V2, arm64.V3, arm64.V4, arm64.V5,
		arm64.V6, arm64.V7, arm64.V8, arm64.V9,
		arm64.V10, arm64.V11, arm64.V12, arm64.V13,
		arm64.V14, arm64.V15, arm64.V16, arm64.V17,
	}

	// Temporary vectors for arithmetic: t[0..9] = V18..V27
	// V28 is reserved for vOneVec (constant 1s)
	t := []arm64.VectorRegister{
		arm64.V18, arm64.V19, arm64.V20, arm64.V21,
		arm64.V22, arm64.V23, arm64.V24, arm64.V25,
		arm64.V26, arm64.V27,
	}

	// Scratch registers (used within macros, can be overwritten freely)
	scratch0 := arm64.V30
	scratch1 := arm64.V31

	// Additional GP registers for loop control and addresses
	rKeyPtr := registers.Pop()  // current round key pointer
	batchIdx := registers.Pop() // outer loop counter (0..3)
	stepIdx := registers.Pop()  // inner loop counter (0..N-1)
	// Pointers to 4 rows for current batch
	ptr0 := registers.Pop()    // data pointer for batch row 0
	ptr1 := registers.Pop()    // data pointer for batch row 1
	ptr2 := registers.Pop()    // data pointer for batch row 2
	ptr3 := registers.Pop()    // data pointer for batch row 3
	tmpCalc := registers.Pop() // temporary for address calculations

	// =========================================================================
	// Modular Arithmetic Macros
	// =========================================================================

	// add computes (a + b) mod q using conditional subtraction
	// Result is in [0, q) assuming inputs are in [0, q)
	// Note: VSUB(a, b, dst) computes dst = b - a
	add := func(a, b, into arm64.VectorRegister) {
		f.VADD(a.S4(), b.S4(), scratch0.S4())            // scratch0 = a + b
		f.VSUB(vQ.S4(), scratch0.S4(), scratch1.S4())    // scratch1 = scratch0 - vQ = (a + b) - q
		f.VUMIN(scratch0.S4(), scratch1.S4(), into.S4()) // into = min(a+b, a+b-q)
	}

	// sub computes (a - b) mod q using conditional addition
	// Result is in [0, q) assuming inputs are in [0, q)
	// Note: VSUB(a, b, dst) computes dst = b - a
	sub := func(a, b, into arm64.VectorRegister) {
		f.VSUB(b.S4(), a.S4(), scratch0.S4())            // scratch0 = a - b (may underflow)
		f.VADD(vQ.S4(), scratch0.S4(), scratch1.S4())    // scratch1 = (a - b) + q
		f.VUMIN(scratch0.S4(), scratch1.S4(), into.S4()) // into = min(a-b, a-b+q)
	}
	_ = sub // reserved for future optimization of matMulInternal

	// mul computes Montgomery multiplication: (a * b * R^-1) mod q
	// where R = 2^32. Uses widening unsigned multiplication.
	// Algorithm (standard Montgomery):
	//   1. ab = a * b (64-bit)
	//   2. m = (ab_lo * qInvNeg) mod 2^32
	//   3. result = (ab + m * q) >> 32
	//   4. if result >= q: result -= q
	//
	// We process 4 lanes in parallel. Since NEON widening multiplies produce
	// 2 x 64-bit results, we need separate UMULL/UMULL2 for low and high lane pairs.
	mulTmp := arm64.V29
	mul := func(a, b, into arm64.VectorRegister) {
		an := vRegNum(a)
		bn := vRegNum(b)
		qn := vRegNum(vQ)
		_ = vRegNum(vMu) // qInvNeg used via vMu directly
		s0n := vRegNum(scratch0)
		s1n := vRegNum(scratch1)
		mn := vRegNum(mulTmp)
		_ = vRegNum(into) // used via method calls

		// Step 1: ab = a * b (64-bit widening multiply)
		// UMULL scratch0.2D, a.2S, b.2S (lanes 0,1)
		encUmullLo := uint32(0x2ea0c000) | (bn << 16) | (an << 5) | s0n
		f.WriteLn(fmt.Sprintf("    WORD $0x%08x // UMULL %s.2D, %s.2S, %s.2S", encUmullLo, baseReg(scratch0), baseReg(a), baseReg(b)))

		// UMULL2 scratch1.2D, a.4S, b.4S (lanes 2,3)
		encUmullHi := uint32(0x6ea0c000) | (bn << 16) | (an << 5) | s1n
		f.WriteLn(fmt.Sprintf("    WORD $0x%08x // UMULL2 %s.2D, %s.4S, %s.4S", encUmullHi, baseReg(scratch1), baseReg(a), baseReg(b)))

		// Step 2: Extract ab_lo (low 32 bits of each 64-bit product)
		// UZP1 gives us the even (low) 32-bit elements
		f.VUZP1(scratch0.S4(), scratch1.S4(), mulTmp.S4()) // mulTmp = ab_lo

		// Step 3: m = (ab_lo * qInvNeg) mod 2^32
		// Note: We use qInvNeg here (stored in vMu), not mu
		f.VMUL_S4(mulTmp.S4(), vMu.S4(), mulTmp.S4()) // mulTmp = m = ab_lo * qInvNeg

		// Step 4: Compute m * q (64-bit)
		// UMULL t8.2D, mulTmp.2S, vQ.2S
		t8n := vRegNum(t[8])
		t9n := vRegNum(t[9])
		encMqLo := uint32(0x2ea0c000) | (qn << 16) | (mn << 5) | t8n
		f.WriteLn(fmt.Sprintf("    WORD $0x%08x // UMULL %s.2D, %s.2S, %s.2S", encMqLo, baseReg(t[8]), baseReg(mulTmp), baseReg(vQ)))

		// UMULL2 t9.2D, mulTmp.4S, vQ.4S
		encMqHi := uint32(0x6ea0c000) | (qn << 16) | (mn << 5) | t9n
		f.WriteLn(fmt.Sprintf("    WORD $0x%08x // UMULL2 %s.2D, %s.4S, %s.4S", encMqHi, baseReg(t[9]), baseReg(mulTmp), baseReg(vQ)))

		// Step 5: Add ab + m*q (64-bit addition)
		// ADD scratch0.2D, scratch0.2D, t8.2D
		f.VADD(scratch0.D2(), t[8].D2(), scratch0.D2())
		// ADD scratch1.2D, scratch1.2D, t9.2D
		f.VADD(scratch1.D2(), t[9].D2(), scratch1.D2())

		// Step 6: Extract high 32 bits (>> 32) using UZP2
		f.VUZP2(scratch0.S4(), scratch1.S4(), into.S4()) // into = (ab + m*q) >> 32

		// Step 7: Reduce if result >= q
		// Note: VSUB(a, b, dst) computes dst = b - a
		f.VSUB(vQ.S4(), into.S4(), mulTmp.S4())    // mulTmp = into - vQ = result - q
		f.VUMIN(into.S4(), mulTmp.S4(), into.S4()) // into = min(result, result - q)
	}

	// double computes 2*a mod q
	// Note: VSUB(a, b, dst) computes dst = b - a
	double := func(a, into arm64.VectorRegister) {
		f.VSHL("$1", a.S4(), scratch0.S4())
		f.VSUB(vQ.S4(), scratch0.S4(), scratch1.S4())    // scratch1 = scratch0 - vQ = 2a - q
		f.VUMIN(scratch0.S4(), scratch1.S4(), into.S4()) // into = min(2a, 2a-q)
	}

	// sbox applies x^3 (cubic S-box for Koalabear)
	sbox := func(state arm64.VectorRegister) {
		mul(state, state, t[0]) // t[0] = state^2
		mul(state, t[0], state) // state = state^3
	}

	// =========================================================================
	// Matrix Multiplication
	// =========================================================================

	// matMul4 computes 4x4 circulant matrix multiplication with matrix:
	//   (2 3 1 1)
	//   (1 2 3 1)
	//   (1 1 2 3)
	//   (3 1 1 2)
	// This is the efficient Plonky3 MDS matrix.
	matMul4 := func(s []arm64.VectorRegister) {
		// Algorithm from Plonky3 external.rs
		add(s[0], s[1], t[0]) // t0 = s0 + s1
		add(s[2], s[3], t[1]) // t1 = s2 + s3
		add(t[0], t[1], t[2]) // t2 = s0 + s1 + s2 + s3 (sum)
		add(t[2], s[1], t[3]) // t3 = sum + s1
		add(t[2], s[3], t[4]) // t4 = sum + s3
		double(s[0], s[3])    // new s3 = 2*s0
		add(s[3], t[4], s[3]) // s3 = 2*s0 + sum + s3 = 2*s0 + (s0+s1+s2+2*s3)
		double(s[2], s[1])    // new s1 = 2*s2
		add(s[1], t[3], s[1]) // s1 = 2*s2 + sum + s1 = (s0+2*s1+s2+s3) + s2
		add(t[0], t[3], s[0]) // s0 = (s0+s1) + (sum+s1) = 2*s0 + 3*s1 + s2 + s3
		add(t[1], t[4], s[2]) // s2 = (s2+s3) + (sum+s3) = s0 + s1 + 2*s2 + 3*s3
	}

	// matMulExternal computes the external (full) matrix multiplication
	// The 16x16 matrix is circ(2*M4, M4, M4, M4) applied in blocks
	matMulExternal := func() {
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
	}

	// =========================================================================
	// Internal Matrix (Partial Round)
	// =========================================================================

	// halve computes a/2 mod q using conditional add of q for odd values
	// Algorithm: if a is odd, result = (a + q) >> 1, else result = a >> 1
	// Uses UHADD which computes (a + b) / 2 without overflow
	// Optimized approach from plonky3: uses VSHL + SSHR to create mask, then UHADD
	halve := func(a, into arm64.VectorRegister) {
		an := vRegNum(a)
		s0n := vRegNum(scratch0)
		s1n := vRegNum(scratch1)
		dn := vRegNum(into)

		// mask = (a & 1) << 31
		f.VAND(a.B16(), vOneVec.B16(), scratch0.B16())
		f.VSHL("$31", scratch0.S4(), scratch0.S4())

		// mask = mask >> 31 (arithmetic shift, creates all 1s or all 0s)
		// SSHR Vd.4S, Vn.4S, #31 - encoding: 0 1 0 01111 00100001 000001 Rn Rd
		encoding := uint32(0x4f210400) | (s0n << 5) | s0n
		f.WriteLn(fmt.Sprintf("    WORD $0x%08x // SSHR V%d.4S, V%d.4S, #31", encoding, s0n, s0n))

		// masked_q = mask & q (q if odd, 0 if even)
		f.VAND(vQ.B16(), scratch0.B16(), scratch1.B16())

		// result = UHADD(a, masked_q) = (a + masked_q) / 2
		// UHADD Vd.4S, Vn.4S, Vm.4S - Encoding: 0 1 1 01110 10 1 Rm 0000 01 Rn Rd = 0x6ea00400
		uhaddEncoding := uint32(0x6ea00400) | (s1n << 16) | (an << 5) | dn
		f.WriteLn(fmt.Sprintf("    WORD $0x%08x // UHADD V%d.4S, V%d.4S, V%d.4S", uhaddEncoding, dn, an, s1n))
	}

	// halveInPlace is a variant that operates in-place on the input register
	halveInPlace := func(a arm64.VectorRegister) {
		halve(a, a)
	}
	_ = halveInPlace

	// mul2ExpNegN computes a * 2^(-n) mod q
	// For small n (n <= 4), use repeated halving (simpler, fewer dependencies)
	// For large n (n > 4), use Montgomery reduction (fewer total operations)
	//
	// Montgomery reduction approach:
	// a * 2^{-n} = (a * 2^{32-n}) * 2^{-32} mod q
	// where 2^{-32} mod q is computed via Montgomery reduction
	mul2ExpNegN := func(a, into arm64.VectorRegister, n int) {
		if n <= 4 {
			// For small n, repeated halving is efficient
			halve(a, into)
			for i := 1; i < n; i++ {
				halve(into, into)
			}
			return
		}

		// For larger n, use Montgomery reduction
		// v = a << (32 - n) gives us a 64-bit value
		// Then Montgomery reduce: result = (v + (v_lo * mu mod 2^32) * q) >> 32
		shift := 32 - n

		an := vRegNum(a)
		s0n := vRegNum(scratch0)
		s1n := vRegNum(scratch1)
		mn := vRegNum(mulTmp)
		qn := vRegNum(vQ)
		_ = vRegNum(into) // used below via direct register access

		// Step 1: Widen and shift left: v = a << shift (64-bit result)
		// USHLL scratch0.2D, a.2S, #shift (low 2 lanes)
		encLo := uint32(0x2f00a400) | (uint32(32+shift) << 16) | (an << 5) | s0n
		f.WriteLn(fmt.Sprintf("    WORD $0x%08x // USHLL V%d.2D, V%d.2S, #%d", encLo, s0n, an, shift))

		// USHLL2 scratch1.2D, a.4S, #shift (high 2 lanes)
		encHi := uint32(0x6f00a400) | (uint32(32+shift) << 16) | (an << 5) | s1n
		f.WriteLn(fmt.Sprintf("    WORD $0x%08x // USHLL2 V%d.2D, V%d.4S, #%d", encHi, s1n, an, shift))

		// Step 2: Extract v_lo (low 32 bits of each 64-bit lane)
		f.VUZP1(scratch0.S4(), scratch1.S4(), mulTmp.S4())

		// Step 3: m = v_lo * mu (mod 2^32)
		f.VMUL_S4(mulTmp.S4(), vMu.S4(), mulTmp.S4())

		// Step 4: Compute m * q (64-bit widening multiply)
		t8n := vRegNum(t[8])
		t9n := vRegNum(t[9])
		// UMULL t8.2D, mulTmp.2S, vQ.2S
		umullLoEnc := uint32(0x2ea0c000) | (qn << 16) | (mn << 5) | t8n
		f.WriteLn(fmt.Sprintf("    WORD $0x%08x // UMULL V%d.2D, V%d.2S, V%d.2S", umullLoEnc, t8n, mn, qn))

		// UMULL2 t9.2D, mulTmp.4S, vQ.4S
		umullHiEnc := uint32(0x6ea0c000) | (qn << 16) | (mn << 5) | t9n
		f.WriteLn(fmt.Sprintf("    WORD $0x%08x // UMULL2 V%d.2D, V%d.4S, V%d.4S", umullHiEnc, t9n, mn, qn))

		// Step 5: Add v + m*q (64-bit)
		f.VADD(scratch0.D2(), t[8].D2(), scratch0.D2())
		f.VADD(scratch1.D2(), t[9].D2(), scratch1.D2())

		// Step 6: Extract high 32 bits (>> 32) using UZP2
		f.VUZP2(scratch0.S4(), scratch1.S4(), into.S4())

		// Step 7: Final reduction: if result >= q, subtract q
		f.VSUB(vQ.S4(), into.S4(), mulTmp.S4())
		f.VUMIN(into.S4(), mulTmp.S4(), into.S4())
	}

	// triple computes 3*a mod q = 2*a + a
	triple := func(a, into arm64.VectorRegister) {
		double(a, scratch0)    // scratch0 = 2*a
		add(scratch0, a, into) // into = 2*a + a = 3*a
	}

	// quadruple computes 4*a mod q = 2*(2*a)
	quadruple := func(a, into arm64.VectorRegister) {
		double(a, scratch0)    // scratch0 = 2*a
		double(scratch0, into) // into = 4*a
	}

	// matMulInternal computes the internal matrix multiplication for partial rounds
	// The matrix is M = I + Diag(V) where I is the all-ones matrix
	// Diagonal V = [-2, 1, 2, 1/2, 3, 4, -1/2, -3, -4, 1/2^8, 1/8, 1/2^24, -1/2^8, -1/8, -1/16, -1/2^24]
	//
	// Algorithm (from Plonky3):
	// 1. Compute sum = sum of all state elements
	// 2. For each element: new_v[i] = sum + v[i] * diag[i]
	// 3. Special case for v[0]: new_v[0] = sum - 2*v[0] = sum_tail - v[0] (since sum_tail = sum - v[0])
	//
	// For negative diagonal coefficients, we compute the positive multiplication and subtract from sum
	matMulInternal := func() {
		// Step 1: Compute sum of all state elements
		// We do this in a tree-reduction pattern for efficiency
		add(v[0], v[1], t[0])
		add(v[2], v[3], t[1])
		add(v[4], v[5], t[2])
		add(v[6], v[7], t[3])
		add(t[0], t[1], t[0])
		add(t[2], t[3], t[2])
		add(t[0], t[2], t[0]) // t[0] = sum of v[0..7]

		add(v[8], v[9], t[4])
		add(v[10], v[11], t[5])
		add(v[12], v[13], t[6])
		add(v[14], v[15], t[7])
		add(t[4], t[5], t[4])
		add(t[6], t[7], t[6])
		add(t[4], t[6], t[4]) // t[4] = sum of v[8..15]

		add(t[0], t[4], t[0]) // t[0] = sum of all 16 elements

		// Step 2: Apply diagonal multiplication and add/sub sum
		// Diagonal: [-2, 1, 2, 1/2, 3, 4, -1/2, -3, -4, 1/2^8, 1/8, 1/2^24, -1/2^8, -1/8, -1/16, -1/2^24]

		// v[0]: diag=-2, formula: sum - 2*v[0]
		double(v[0], v[0])
		sub(t[0], v[0], v[0])

		// v[1]: diag=1, formula: sum + v[1]
		add(t[0], v[1], v[1])

		// v[2]: diag=2, formula: sum + 2*v[2]
		double(v[2], v[2])
		add(t[0], v[2], v[2])

		// v[3]: diag=1/2, formula: sum + v[3]/2
		halve(v[3], v[3])
		add(t[0], v[3], v[3])

		// v[4]: diag=3, formula: sum + 3*v[4]
		triple(v[4], v[4])
		add(t[0], v[4], v[4])

		// v[5]: diag=4, formula: sum + 4*v[5]
		quadruple(v[5], v[5])
		add(t[0], v[5], v[5])

		// v[6]: diag=-1/2, formula: sum - v[6]/2
		halve(v[6], v[6])
		sub(t[0], v[6], v[6])

		// v[7]: diag=-3, formula: sum - 3*v[7]
		triple(v[7], v[7])
		sub(t[0], v[7], v[7])

		// v[8]: diag=-4, formula: sum - 4*v[8]
		quadruple(v[8], v[8])
		sub(t[0], v[8], v[8])

		// v[9]: diag=1/2^8, formula: sum + v[9]/256
		mul2ExpNegN(v[9], v[9], 8)
		add(t[0], v[9], v[9])

		// v[10]: diag=1/8, formula: sum + v[10]/8
		mul2ExpNegN(v[10], v[10], 3)
		add(t[0], v[10], v[10])

		// v[11]: diag=1/2^24, formula: sum + v[11]/2^24
		mul2ExpNegN(v[11], v[11], 24)
		add(t[0], v[11], v[11])

		// v[12]: diag=-1/2^8, formula: sum - v[12]/256
		mul2ExpNegN(v[12], v[12], 8)
		sub(t[0], v[12], v[12])

		// v[13]: diag=-1/8, formula: sum - v[13]/8
		mul2ExpNegN(v[13], v[13], 3)
		sub(t[0], v[13], v[13])

		// v[14]: diag=-1/16, formula: sum - v[14]/16
		mul2ExpNegN(v[14], v[14], 4)
		sub(t[0], v[14], v[14])

		// v[15]: diag=-1/2^24, formula: sum - v[15]/2^24
		mul2ExpNegN(v[15], v[15], 24)
		sub(t[0], v[15], v[15])
	}

	// =========================================================================
	// Round Functions
	// =========================================================================

	// Round key access:
	// roundKeys is [][]fr.Element - an array of slice headers
	// Each slice header is 24 bytes: [ptr, len, cap]
	// For round r, the data pointer is at addrRoundKeys + r*24
	// Full rounds (r < 3 or r >= 24): 16 keys, accessed as *(dataPtr + j*4)
	// Partial rounds (3 <= r < 24): 1 key, accessed as *(dataPtr)

	roundIdx := 0 // track which round we're on

	// addRoundKeyFull loads round key j for current full round and adds to state
	addRoundKeyFull := func(j int) {
		// Load data pointer for this round from slice header
		f.MOVD(fmt.Sprintf("%d(%s)", roundIdx*24, addrRoundKeys), rKeyPtr)
		// Load key[j] and broadcast
		f.ADD(uint64(j*4), rKeyPtr, tmpCalc)
		f.WriteLn(fmt.Sprintf("    VLD1R (%s), [%s]", tmpCalc, scratch0.S4()))
		add(v[j], scratch0, v[j])
	}

	// addRoundKeyPartial loads the single round key for current partial round
	addRoundKeyPartial := func() {
		// Load data pointer for this round from slice header
		f.MOVD(fmt.Sprintf("%d(%s)", roundIdx*24, addrRoundKeys), rKeyPtr)
		// Load key[0] and broadcast
		f.WriteLn(fmt.Sprintf("    VLD1R (%s), [%s]", rKeyPtr, scratch0.S4()))
		add(v[0], scratch0, v[0])
	}

	// fullRound applies round key, S-box to all elements, then external matrix
	fullRound := func() {
		for j := 0; j < 16; j++ {
			addRoundKeyFull(j)
			sbox(v[j])
		}
		matMulExternal()
		roundIdx++
	}

	// partialRound applies round key and S-box only to v[0], then internal matrix
	partialRound := func() {
		addRoundKeyPartial()
		sbox(v[0])
		roundIdx++
		matMulInternal()
	}

	// =========================================================================
	// Main Loop Structure
	// =========================================================================
	// Process 4 batches, each containing N=64 steps
	// Each step processes 8 input elements and accumulates into state

	f.MOVD(0, batchIdx)
	f.LABEL("batch_loop")

	// Zero state vectors
	for i := 0; i < 16; i++ {
		f.VEOR(v[i].B16(), v[i].B16(), v[i].B16())
	}

	// Initialize pointers for 4 parallel inputs (one per NEON lane)
	const N = 512 / 8 // 64 steps per batch
	f.MOVD(0, stepIdx)

	// Calculate base addresses for 4 lanes
	// Each row has 512 elements × 4 bytes = 2048 bytes
	// Batch b processes rows 4*b, 4*b+1, 4*b+2, 4*b+3
	// Base offset = b * 4 * 2048 = b * 8192 = b << 13
	f.WriteLn(fmt.Sprintf("    LSL $13, %s, %s", batchIdx, tmpCalc))
	f.ADD(addrMatrix, tmpCalc, ptr0)
	f.ADD(2048, ptr0, ptr1) // next row
	f.ADD(2048, ptr1, ptr2) // next row
	f.ADD(2048, ptr2, ptr3) // next row

	f.LABEL("step_loop")

	// Load 8 elements from each of 4 lanes into t[0..7]
	// This forms a 4x8 matrix in transposed form
	for j := 0; j < 8; j++ {
		// Load one 32-bit element from each of 4 lanes
		f.MOVWU(fmt.Sprintf("(%s)", ptr0), tmpCalc)
		f.WriteLn(fmt.Sprintf("    VMOV %s, %s", tmpCalc, t[j].SAt(0)))
		f.MOVWU(fmt.Sprintf("(%s)", ptr1), tmpCalc)
		f.WriteLn(fmt.Sprintf("    VMOV %s, %s", tmpCalc, t[j].SAt(1)))
		f.MOVWU(fmt.Sprintf("(%s)", ptr2), tmpCalc)
		f.WriteLn(fmt.Sprintf("    VMOV %s, %s", tmpCalc, t[j].SAt(2)))
		f.MOVWU(fmt.Sprintf("(%s)", ptr3), tmpCalc)
		f.WriteLn(fmt.Sprintf("    VMOV %s, %s", tmpCalc, t[j].SAt(3)))
		f.ADD(4, ptr0, ptr0)
		f.ADD(4, ptr1, ptr1)
		f.ADD(4, ptr2, ptr2)
		f.ADD(4, ptr3, ptr3)
	}

	// Copy input into state[8..15] (replacing, not accumulating)
	for j := 0; j < 8; j++ {
		f.VMOV(t[j].B16(), v[8+j].B16())
	}

	// Store t[0..7] on stack for feed-forward at the end
	// Use the frame's stack space at offsets 0-127 (8 vectors × 16 bytes = 128 bytes)
	f.MOVD("RSP", tmpCalc)
	for j := 0; j < 8; j++ {
		f.VST1_P(t[j].S4(), tmpCalc, 16)
	}

	// Reset round index for this permutation
	roundIdx = 0

	// Initial external matrix multiplication (before any rounds)
	// This mixes the state v[0..7] with the new inputs v[8..15]
	matMulExternal()

	// Apply rf full rounds (rounds 0, 1, 2)
	for i := 0; i < rf; i++ {
		fullRound()
	}

	// Apply partial rounds (rounds 3 to 23)
	for i := 0; i < partialRounds; i++ {
		partialRound()
	}

	// Apply rf more full rounds (rounds 24, 25, 26)
	for i := 0; i < rf; i++ {
		fullRound()
	}

	// Restore t[0..7] from stack
	f.MOVD("RSP", tmpCalc)
	for j := 0; j < 8; j++ {
		f.VLD1_P(16, tmpCalc, t[j].S4())
	}

	// Feed-forward: state[j] = state[8+j] + original_input[j]
	// Note: this overwrites v[0:8] with the feed-forward result
	for j := 0; j < 8; j++ {
		add(v[8+j], t[j], v[j])
	}

	// Loop control
	f.ADD(1, stepIdx, stepIdx)
	f.CMP(N, stepIdx)
	f.WriteLn("    BNE step_loop")

	// Store results for this batch
	// Output is 4 x 8 elements (32 elements total per batch)
	// tmpCalc = batchIdx * 128 = batchIdx << 7
	f.WriteLn(fmt.Sprintf("    LSL $7, %s, %s", batchIdx, tmpCalc))
	f.ADD(addrResult, tmpCalc, ptr0)
	f.ADD(32, ptr0, ptr1)
	f.ADD(32, ptr1, ptr2)
	f.ADD(32, ptr2, ptr3)

	for j := 0; j < 8; j++ {
		f.WriteLn(fmt.Sprintf("    VMOV %s, %s", v[j].SAt(0), tmpCalc))
		f.MOVWU(tmpCalc, fmt.Sprintf("(%s)", ptr0))
		f.WriteLn(fmt.Sprintf("    VMOV %s, %s", v[j].SAt(1), tmpCalc))
		f.MOVWU(tmpCalc, fmt.Sprintf("(%s)", ptr1))
		f.WriteLn(fmt.Sprintf("    VMOV %s, %s", v[j].SAt(2), tmpCalc))
		f.MOVWU(tmpCalc, fmt.Sprintf("(%s)", ptr2))
		f.WriteLn(fmt.Sprintf("    VMOV %s, %s", v[j].SAt(3), tmpCalc))
		f.MOVWU(tmpCalc, fmt.Sprintf("(%s)", ptr3))
		f.ADD(4, ptr0, ptr0)
		f.ADD(4, ptr1, ptr1)
		f.ADD(4, ptr2, ptr2)
		f.ADD(4, ptr3, ptr3)
	}

	// Batch loop control
	f.ADD(1, batchIdx, batchIdx)
	f.CMP(4, batchIdx)
	f.WriteLn("    BNE batch_loop")

	f.RET()
}
