// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

//go:build amd64 && !purego

package fr

import (
	"fmt"
	"testing"

	"github.com/consensys/gnark-crypto/utils/cpu"
)

// TestRadix52Conversion tests that radix-52 conversion is correct
func TestRadix52Conversion(t *testing.T) {
	// Test the radix-52 conversion formulas
	// For a 4-word element [a0, a1, a2, a3]:
	// l0 = a0 & mask52
	// l1 = (a0 >> 52) | ((a1 << 12) & mask52)
	// l2 = (a1 >> 40) | ((a2 << 24) & mask52)
	// l3 = (a2 >> 28) | ((a3 << 36) & mask52)
	// l4 = a3 >> 16

	const mask52 = uint64(0xFFFFFFFFFFFFF)

	// Test with a simple element
	var e Element
	e.SetUint64(1) // This will be in Montgomery form

	a0, a1, a2, a3 := e[0], e[1], e[2], e[3]
	t.Logf("Element (Montgomery form): [%016x, %016x, %016x, %016x]", a0, a1, a2, a3)

	l0 := a0 & mask52
	l1 := ((a0 >> 52) | (a1 << 12)) & mask52
	l2 := ((a1 >> 40) | (a2 << 24)) & mask52
	l3 := ((a2 >> 28) | (a3 << 36)) & mask52
	l4 := a3 >> 16

	t.Logf("Radix-52: [%013x, %013x, %013x, %013x, %013x]", l0, l1, l2, l3, l4)

	// Convert back to radix-64
	b0 := l0 | (l1 << 52)
	b1 := (l1 >> 12) | (l2 << 40)
	b2 := (l2 >> 24) | (l3 << 28)
	b3 := (l3 >> 36) | (l4 << 16)

	t.Logf("Back to radix-64: [%016x, %016x, %016x, %016x]", b0, b1, b2, b3)

	// Verify
	if b0 != a0 || b1 != a1 || b2 != a2 || b3 != a3 {
		t.Errorf("Radix conversion round-trip failed")
		t.Errorf("Original: [%016x, %016x, %016x, %016x]", a0, a1, a2, a3)
		t.Errorf("After:    [%016x, %016x, %016x, %016x]", b0, b1, b2, b3)
	}
}

// DebugMulVecIFMA helps debug the IFMA implementation
func TestDebugMulVecIFMA(t *testing.T) {
	if !cpu.SupportAVX512IFMA {
		t.Skip("IFMA not supported on this CPU")
	}

	// Test with identity-like values (1 * 1 should give 1 in Montgomery form)
	const n = 8
	one := make(Vector, n)
	result := make(Vector, n)
	expected := make(Vector, n)

	// Set all to 1 (in Montgomery form)
	for i := range n {
		one[i].SetOne()
	}

	// 1 * 1 = 1
	mulVecGeneric(expected, one, one)

	// Show what the Montgomery form of 1 looks like
	t.Logf("Montgomery form of 1: %v", one[0])
	t.Logf("Element bytes: [%016x, %016x, %016x, %016x]", one[0][0], one[0][1], one[0][2], one[0][3])

	// Now run IFMA
	mulVec(&result[0], &one[0], &one[0], 1)

	t.Logf("Expected (1*1=1): %v", expected[0])
	t.Logf("IFMA result:      %v", result[0])
	t.Logf("IFMA bytes: [%016x, %016x, %016x, %016x]", result[0][0], result[0][1], result[0][2], result[0][3])

	for i := range n {
		if !result[i].Equal(&expected[i]) {
			t.Errorf("Index %d: expected %v, got %v", i, expected[i], result[i])
		}
	}
}

// Print qElement for debugging
func TestPrintConstants(t *testing.T) {
	fmt.Printf("qElement: [%016x, %016x, %016x, %016x]\n", qElement[0], qElement[1], qElement[2], qElement[3])
	fmt.Printf("qInvNeg: %016x\n", qInvNeg)
}

// TestIFMASimpleIdentity tests with simple integer values
func TestIFMASimpleIdentity(t *testing.T) {
	if !cpu.SupportAVX512IFMA {
		t.Skip("IFMA not supported on this CPU")
	}

	const n = 8
	a := make(Vector, n)
	b := make(Vector, n)
	resultIFMA := make(Vector, n)
	resultGeneric := make(Vector, n)

	// Use simple values: a[i] = i+1, b[i] = 1
	for i := range n {
		a[i].SetUint64(uint64(i + 1))
		b[i].SetOne()
	}

	// a * 1 = a
	mulVecGeneric(resultGeneric, a, b)

	// IFMA
	mulVec(&resultIFMA[0], &a[0], &b[0], 1)

	t.Logf("Input a: %v", a)
	t.Logf("Input b (all 1s): %v", b)
	t.Logf("Expected (a*1=a): %v", resultGeneric)
	t.Logf("IFMA result: %v", resultIFMA)

	for i := range n {
		t.Logf("Index %d: a=%v, generic=%v, ifma=%v",
			i, a[i].String(), resultGeneric[i].String(), resultIFMA[i].String())
		if !resultIFMA[i].Equal(&resultGeneric[i]) {
			t.Errorf("Mismatch at index %d", i)
		}
	}
}

// TestMulVecIFMACorrectness tests that IFMA produces the same results as generic
func TestMulVecIFMACorrectness(t *testing.T) {
	if !cpu.SupportAVX512IFMA {
		t.Skip("IFMA not supported on this CPU")
	}

	// Test with 8 elements (minimum for IFMA path)
	const n = 8
	a := make(Vector, n)
	b := make(Vector, n)
	resultIFMA := make(Vector, n)
	resultGeneric := make(Vector, n)

	// Initialize with simple values
	for i := range n {
		a[i].SetUint64(uint64(i + 1))
		b[i].SetUint64(uint64(i + 2))
	}

	// Compute using generic (the reference)
	mulVecGeneric(resultGeneric, a, b)

	// Compute using IFMA
	mulVec(&resultIFMA[0], &a[0], &b[0], 1) // 1 group of 8 elements

	// Compare results - check both Equal and raw bytes
	mismatchCount := 0
	for i := range n {
		if !resultIFMA[i].Equal(&resultGeneric[i]) {
			mismatchCount++
			t.Errorf("Mismatch at index %d:\n  a[%d]=%v\n  b[%d]=%v\n  IFMA=%v (bytes: %x)\n  Generic=%v (bytes: %x)",
				i, i, a[i].String(), i, b[i].String(),
				resultIFMA[i].String(), resultIFMA[i],
				resultGeneric[i].String(), resultGeneric[i])
		}
	}

	// Also print all values for debugging
	t.Log("Input a:", a)
	t.Log("Input b:", b)
	t.Log("IFMA result:", resultIFMA)
	t.Log("Generic result:", resultGeneric)

	if mismatchCount == 0 {
		t.Log("All 8 elements match!")
	}
}

// TestMontgomeryMulRadix52 tests the radix-52 Montgomery multiplication in Go
// This mirrors the assembly algorithm to help debug
func TestMontgomeryMulRadix52(t *testing.T) {
	const mask52 = uint64(0xFFFFFFFFFFFFF)

	// Convert radix-64 to radix-52
	toRadix52 := func(a [4]uint64) [5]uint64 {
		return [5]uint64{
			a[0] & mask52,
			((a[0] >> 52) | (a[1] << 12)) & mask52,
			((a[1] >> 40) | (a[2] << 24)) & mask52,
			((a[2] >> 28) | (a[3] << 36)) & mask52,
			a[3] >> 16,
		}
	}

	// Convert radix-52 to radix-64
	fromRadix52 := func(l [5]uint64) [4]uint64 {
		return [4]uint64{
			l[0] | (l[1] << 52),
			(l[1] >> 12) | (l[2] << 40),
			(l[2] >> 24) | (l[3] << 28),
			(l[3] >> 36) | (l[4] << 16),
		}
	}

	// Get q in radix-52
	q52 := toRadix52(qElement)
	t.Logf("q in radix-52: [%013x, %013x, %013x, %013x, %013x]",
		q52[0], q52[1], q52[2], q52[3], q52[4])

	// Compute qInvNeg52
	qInvNeg52 := qInvNeg & mask52
	t.Logf("qInvNeg52: %013x", qInvNeg52)

	// Verify qInvNeg52: qInvNeg52 * q52[0] should be -1 mod 2^52
	check := (qInvNeg52 * q52[0]) & mask52
	t.Logf("qInvNeg52 * q52[0] mod 2^52 = %013x (expected %013x)", check, mask52)

	// Test input: Montgomery form of 1
	var one Element
	one.SetOne()
	t.Logf("Montgomery 1: [%016x, %016x, %016x, %016x]", one[0], one[1], one[2], one[3])

	a52 := toRadix52(one)
	b52 := toRadix52(one)
	t.Logf("A in radix-52: [%013x, %013x, %013x, %013x, %013x]",
		a52[0], a52[1], a52[2], a52[3], a52[4])

	// CIOS Montgomery multiplication (5 rounds)
	T := [6]uint64{0, 0, 0, 0, 0, 0}

	for round := 0; round < 5; round++ {
		bi := b52[round]
		t.Logf("Round %d: B[%d] = %013x", round, round, bi)

		// T += A * B[i]
		for j := 0; j < 5; j++ {
			prod := uint128Mul(a52[j], bi)
			T[j] += prod.lo & mask52
			T[j+1] += prod.lo >> 52
			T[j+1] += prod.hi << (64 - 52)
		}
		t.Logf("  After A*B[%d]: T = [%x, %x, %x, %x, %x, %x]",
			round, T[0], T[1], T[2], T[3], T[4], T[5])

		// Normalize T[0]
		carry := T[0] >> 52
		T[0] &= mask52
		T[1] += carry

		// Compute m = T[0] * qInvNeg52 mod 2^52
		m := (T[0] * qInvNeg52) & mask52
		t.Logf("  m = %013x", m)

		// T += m * q
		for j := 0; j < 5; j++ {
			prod := uint128Mul(m, q52[j])
			T[j] += prod.lo & mask52
			T[j+1] += prod.lo >> 52
			T[j+1] += prod.hi << (64 - 52)
		}
		t.Logf("  After m*q: T = [%x, %x, %x, %x, %x, %x]",
			T[0], T[1], T[2], T[3], T[4], T[5])

		// Shift: T[0] should be 0 mod 2^52
		carry = T[0] >> 52
		T = [6]uint64{T[1] + carry, T[2], T[3], T[4], T[5], 0}
		t.Logf("  After shift: T = [%x, %x, %x, %x, %x, %x]",
			T[0], T[1], T[2], T[3], T[4], T[5])
	}

	// Final normalization
	for i := 0; i < 4; i++ {
		carry := T[i] >> 52
		T[i] &= mask52
		T[i+1] += carry
	}
	t.Logf("After final normalize: T = [%013x, %013x, %013x, %013x, %013x]",
		T[0], T[1], T[2], T[3], T[4])

	// Convert result back to radix-64 (this is the raw result before x16 correction)
	rawResult := fromRadix52([5]uint64{T[0], T[1], T[2], T[3], T[4]})
	t.Logf("Raw result (radix-64): [%016x, %016x, %016x, %016x]",
		rawResult[0], rawResult[1], rawResult[2], rawResult[3])

	// The raw result is A*B*R^{-1} where R=2^260, but we want R=2^256
	// So we need to multiply by 2^4 = 16
	var rawElem Element
	rawElem[0], rawElem[1], rawElem[2], rawElem[3] = rawResult[0], rawResult[1], rawResult[2], rawResult[3]

	// Multiply by 16 using field multiplication
	var sixteen, corrected Element
	sixteen.SetUint64(16)
	corrected.Mul(&rawElem, &sixteen)

	t.Logf("After x16 correction: [%016x, %016x, %016x, %016x]",
		corrected[0], corrected[1], corrected[2], corrected[3])
	t.Logf("Expected (1*1=1): [%016x, %016x, %016x, %016x]",
		one[0], one[1], one[2], one[3])

	if corrected != one {
		t.Errorf("Montgomery multiplication mismatch")
	}
}

// uint128 represents a 128-bit unsigned integer as two 64-bit parts
type uint128 struct {
	lo, hi uint64
}

// uint128Mul multiplies two 64-bit integers and returns a 128-bit result
func uint128Mul(a, b uint64) uint128 {
	// Use the standard decomposition: a = a_hi * 2^32 + a_lo, etc.
	aLo := a & 0xFFFFFFFF
	aHi := a >> 32
	bLo := b & 0xFFFFFFFF
	bHi := b >> 32

	// Four partial products
	llProd := aLo * bLo
	lhProd := aLo * bHi
	hlProd := aHi * bLo
	hhProd := aHi * bHi

	// Combine: lo = llProd + (lhProd << 32) + (hlProd << 32)
	//          hi = hhProd + carries
	mid := lhProd + hlProd // This could overflow, but we handle it below
	midLo := mid << 32
	midHi := mid >> 32

	lo := llProd + midLo
	carry := uint64(0)
	if lo < llProd {
		carry = 1
	}
	hi := hhProd + midHi + carry

	// Handle overflow of lhProd + hlProd
	if mid < lhProd {
		hi += 1 << 32
	}

	return uint128{lo: lo, hi: hi}
}

// TestConditionalSubtraction tests the radix-52 conditional subtraction logic
func TestConditionalSubtraction(t *testing.T) {
	const mask52 = uint64(0xFFFFFFFFFFFFF)

	// Get q in radix-52
	q52 := [5]uint64{
		qElement[0] & mask52,
		((qElement[0] >> 52) | (qElement[1] << 12)) & mask52,
		((qElement[1] >> 40) | (qElement[2] << 24)) & mask52,
		((qElement[2] >> 28) | (qElement[3] << 36)) & mask52,
		qElement[3] >> 16,
	}

	// Test case 1: value < q (should not subtract)
	// Use q - 1 which should remain as q - 1
	v1 := [5]uint64{q52[0] - 1, q52[1], q52[2], q52[3], q52[4]}

	result1 := conditionalSubtractRadix52(v1, q52, mask52)
	t.Logf("Test 1: v = q-1 (should remain unchanged)")
	t.Logf("  Input:  %x", v1)
	t.Logf("  Output: %x", result1)
	if result1 != v1 {
		t.Errorf("Test 1 failed: expected %x, got %x", v1, result1)
	}

	// Test case 2: value = q (should become 0)
	v2 := q52
	result2 := conditionalSubtractRadix52(v2, q52, mask52)
	expected2 := [5]uint64{0, 0, 0, 0, 0}
	t.Logf("Test 2: v = q (should become 0)")
	t.Logf("  Input:  %x", v2)
	t.Logf("  Output: %x", result2)
	if result2 != expected2 {
		t.Errorf("Test 2 failed: expected %x, got %x", expected2, result2)
	}

	// Test case 3: value = 2q - 1 (should become q - 1)
	// This requires two subtractions but our function does one at a time
	// Just test that 2q - 1 - q = q - 1
	// First, compute 2q - 1 in radix-52 (with potential overflow)
	v3 := [5]uint64{
		(q52[0]*2 - 1) & mask52,
		(q52[1]*2 + ((q52[0]*2 - 1) >> 52)) & mask52,
		(q52[2]*2 + ((q52[1]*2 + ((q52[0]*2 - 1) >> 52)) >> 52)) & mask52,
		(q52[3]*2 + ((q52[2]*2 + ((q52[1]*2 + ((q52[0]*2 - 1) >> 52)) >> 52)) >> 52)) & mask52,
		(q52[4] * 2) + ((q52[3]*2 + ((q52[2]*2 + ((q52[1]*2 + ((q52[0]*2 - 1) >> 52)) >> 52)) >> 52)) >> 52),
	}
	result3 := conditionalSubtractRadix52(v3, q52, mask52)
	// After one subtraction, should be q - 1
	expected3 := v1 // q - 1
	t.Logf("Test 3: v = 2q-1 (should become q-1 after one subtraction)")
	t.Logf("  Input:  %x", v3)
	t.Logf("  Output: %x", result3)
	if result3 != expected3 {
		t.Errorf("Test 3 failed: expected %x, got %x", expected3, result3)
	}
}

// conditionalSubtractRadix52 simulates the assembly conditional subtraction
func conditionalSubtractRadix52(v, q [5]uint64, mask52 uint64) [5]uint64 {
	// Compute v - q with borrow propagation
	d := [5]uint64{}
	d[0] = v[0] - q[0]
	d[1] = v[1] - q[1]
	d[2] = v[2] - q[2]
	d[3] = v[3] - q[3]
	d[4] = v[4] - q[4]

	// Propagate borrows (using signed arithmetic shift to detect borrows)
	borrow0 := int64(d[0]) >> 63 // -1 if d[0] is negative (as signed), 0 otherwise
	d[1] = uint64(int64(d[1]) + borrow0)

	borrow1 := int64(d[1]) >> 63
	d[2] = uint64(int64(d[2]) + borrow1)

	borrow2 := int64(d[2]) >> 63
	d[3] = uint64(int64(d[3]) + borrow2)

	borrow3 := int64(d[3]) >> 63
	d[4] = uint64(int64(d[4]) + borrow3)

	// Check if final borrow occurred
	finalBorrow := int64(d[4]) >> 63 // -1 if v < q, 0 if v >= q

	// Mask d to 52 bits
	d[0] &= mask52
	d[1] &= mask52
	d[2] &= mask52
	d[3] &= mask52
	d[4] &= mask52

	// Select: if finalBorrow == -1 (v < q), keep v; else use d (v - q)
	result := [5]uint64{}
	mask := uint64(finalBorrow) // all 1s if borrow, all 0s otherwise
	for i := range 5 {
		result[i] = (v[i] & mask) | (d[i] & ^mask)
	}

	return result
}

// TestMulVecIFMAWithMontgomery tests IFMA with Montgomery form inputs
func TestMulVecIFMAWithMontgomery(t *testing.T) {
	if !cpu.SupportAVX512IFMA {
		t.Skip("IFMA not supported on this CPU")
	}

	const n = 8
	a := make(Vector, n)
	b := make(Vector, n)
	resultIFMA := make(Vector, n)
	resultGeneric := make(Vector, n)

	// Initialize with random Montgomery form values
	for i := range n {
		a[i].MustSetRandom()
		b[i].MustSetRandom()
	}

	// Compute using generic (the reference)
	mulVecGeneric(resultGeneric, a, b)

	// Compute using IFMA
	mulVec(&resultIFMA[0], &a[0], &b[0], 1)

	// Compare results
	mismatches := 0
	for i := range n {
		if !resultIFMA[i].Equal(&resultGeneric[i]) {
			mismatches++
			if mismatches <= 3 {
				t.Errorf("Mismatch at index %d:\n  a[%d]=%v\n  b[%d]=%v\n  IFMA=%v\n  Generic=%v",
					i, i, a[i].String(), i, b[i].String(),
					resultIFMA[i].String(), resultGeneric[i].String())
			}
		}
	}
	if mismatches > 0 {
		t.Errorf("Total mismatches: %d/%d", mismatches, n)
	}
}

// BenchmarkMulVecIFMA benchmarks IFMA vector multiplication
func BenchmarkMulVecIFMA(b *testing.B) {
	if !cpu.SupportAVX512IFMA {
		b.Skip("IFMA not supported on this CPU")
	}

	sizes := []int{8, 64, 256, 1024, 4096}

	for _, size := range sizes {
		a := make(Vector, size)
		bVec := make(Vector, size)
		result := make(Vector, size)

		for i := range size {
			a[i].MustSetRandom()
			bVec[i].MustSetRandom()
		}

		// Benchmark IFMA
		b.Run(fmt.Sprintf("IFMA/%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				mulVec(&result[0], &a[0], &bVec[0], uint64(size/8))
			}
		})

		// Benchmark generic for comparison
		b.Run(fmt.Sprintf("Generic/%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				mulVecGeneric(result, a, bVec)
			}
		})
	}
}

// BenchmarkMulVecComparison provides a direct side-by-side comparison
func BenchmarkMulVecComparison(b *testing.B) {
	if !cpu.SupportAVX512IFMA {
		b.Skip("IFMA not supported on this CPU")
	}

	const size = 1024
	a := make(Vector, size)
	bVec := make(Vector, size)
	result := make(Vector, size)

	for i := range size {
		a[i].MustSetRandom()
		bVec[i].MustSetRandom()
	}

	b.Run("IFMA", func(b *testing.B) {
		b.SetBytes(int64(size * 32 * 3)) // 3 vectors * 32 bytes each
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mulVec(&result[0], &a[0], &bVec[0], uint64(size/8))
		}
	})

	b.Run("Generic", func(b *testing.B) {
		b.SetBytes(int64(size * 32 * 3))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mulVecGeneric(result, a, bVec)
		}
	})
}

// TestMulVecIFMAStress performs extensive testing with random values
func TestMulVecIFMAStress(t *testing.T) {
	if !cpu.SupportAVX512IFMA {
		t.Skip("IFMA not supported on this CPU")
	}

	const iterations = 100
	const size = 64 // Test with 64 elements (8 groups of 8)

	for iter := range iterations {
		a := make(Vector, size)
		bVec := make(Vector, size)
		resultIFMA := make(Vector, size)
		resultGeneric := make(Vector, size)

		// Generate random inputs
		for i := range size {
			a[i].MustSetRandom()
			bVec[i].MustSetRandom()
		}

		// Compute using both methods
		mulVecGeneric(resultGeneric, a, bVec)
		mulVec(&resultIFMA[0], &a[0], &bVec[0], uint64(size/8))

		// Compare results
		for i := range size {
			if !resultIFMA[i].Equal(&resultGeneric[i]) {
				t.Errorf("Iteration %d, index %d: mismatch\n  a=%v\n  b=%v\n  IFMA=%v\n  Generic=%v",
					iter, i, a[i].String(), bVec[i].String(),
					resultIFMA[i].String(), resultGeneric[i].String())
			}
		}
	}
	t.Logf("Stress test passed: %d iterations, %d elements each", iterations, size)
}

// TestMulVecIFMAEdgeCases tests edge cases near modulus boundaries
func TestMulVecIFMAEdgeCases(t *testing.T) {
	if !cpu.SupportAVX512IFMA {
		t.Skip("IFMA not supported on this CPU")
	}

	const n = 8
	a := make(Vector, n)
	bVec := make(Vector, n)
	resultIFMA := make(Vector, n)
	resultGeneric := make(Vector, n)

	// Test case 1: All zeros
	t.Run("AllZeros", func(t *testing.T) {
		for i := range n {
			a[i].SetZero()
			bVec[i].SetOne()
		}
		mulVecGeneric(resultGeneric, a, bVec)
		mulVec(&resultIFMA[0], &a[0], &bVec[0], 1)
		for i := range n {
			if !resultIFMA[i].Equal(&resultGeneric[i]) {
				t.Errorf("Index %d: expected %v, got %v", i, resultGeneric[i], resultIFMA[i])
			}
		}
	})

	// Test case 2: Large values near q-1
	t.Run("NearModulus", func(t *testing.T) {
		for i := range n {
			// Set a[i] = q - 1 - i
			a[i].SetUint64(uint64(i + 1))
			a[i].Neg(&a[i])      // a[i] = -(i+1) = q - (i+1)
			bVec[i].SetUint64(2) // multiply by 2
		}
		mulVecGeneric(resultGeneric, a, bVec)
		mulVec(&resultIFMA[0], &a[0], &bVec[0], 1)
		for i := range n {
			if !resultIFMA[i].Equal(&resultGeneric[i]) {
				t.Errorf("Index %d: expected %v, got %v", i, resultGeneric[i].String(), resultIFMA[i].String())
			}
		}
	})

	// Test case 3: Self-multiplication (squaring)
	t.Run("Squaring", func(t *testing.T) {
		for i := range n {
			a[i].MustSetRandom()
			bVec[i] = a[i]
		}
		mulVecGeneric(resultGeneric, a, bVec)
		mulVec(&resultIFMA[0], &a[0], &bVec[0], 1)
		for i := range n {
			if !resultIFMA[i].Equal(&resultGeneric[i]) {
				t.Errorf("Index %d: expected %v, got %v", i, resultGeneric[i].String(), resultIFMA[i].String())
			}
		}
	})

	// Test case 4: Powers of 2
	t.Run("PowersOf2", func(t *testing.T) {
		for i := range n {
			a[i].SetUint64(1 << uint(i))
			bVec[i].SetUint64(1 << uint(7-i))
		}
		mulVecGeneric(resultGeneric, a, bVec)
		mulVec(&resultIFMA[0], &a[0], &bVec[0], 1)
		for i := range n {
			if !resultIFMA[i].Equal(&resultGeneric[i]) {
				t.Errorf("Index %d: expected %v, got %v", i, resultGeneric[i].String(), resultIFMA[i].String())
			}
		}
	})

	t.Log("Edge case tests passed")
}
