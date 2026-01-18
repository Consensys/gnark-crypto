// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

//go:build amd64 && !purego

package fr

import (
	"testing"

	"github.com/consensys/gnark-crypto/utils/cpu"
)

// This file contains unit tests for verifying correct IFMA instruction usage
// and Plan9 assembly operand ordering. These tests serve as reference for
// future AVX-512 IFMA implementations.

// TestVPMADD52LUQOperandOrder verifies the operand ordering for VPMADD52LUQ
// in Plan9 assembly. VPMADD52LUQ computes: dest = dest + (src1 * src2)[51:0]
//
// Intel syntax: VPMADD52LUQ zmm1, zmm2, zmm3  means zmm1 += low52(zmm2 * zmm3)
// Plan9 syntax needs verification - this test helps determine correct order.
func TestVPMADD52LUQOperandOrder(t *testing.T) {
	if !cpu.SupportAVX512IFMA {
		t.Skip("IFMA not supported on this CPU")
	}

	// Test VPMADD52LUQ: acc += (a * b) & ((1<<52)-1) (low 52 bits)
	// We'll test with acc=10, a=3, b=5, expecting result = 10 + 15 = 25

	acc := []uint64{10, 0, 0, 0, 0, 0, 0, 0}
	a := []uint64{3, 0, 0, 0, 0, 0, 0, 0}
	b := []uint64{5, 0, 0, 0, 0, 0, 0, 0}
	result := make([]uint64, 8)

	testVPMADD52LUQ(&acc[0], &a[0], &b[0], &result[0])

	t.Logf("VPMADD52LUQ test:")
	t.Logf("  acc = %d", acc[0])
	t.Logf("  a   = %d", a[0])
	t.Logf("  b   = %d", b[0])
	t.Logf("  result = %d (expected %d)", result[0], acc[0]+(a[0]*b[0]))

	expected := acc[0] + (a[0]*b[0])&0xFFFFFFFFFFFFF
	if result[0] != expected {
		t.Errorf("VPMADD52LUQ: expected %d, got %d", expected, result[0])
	}

	// Also test with larger values to ensure 52-bit truncation works
	// Use values that produce a product > 52 bits
	acc2 := []uint64{0, 0, 0, 0, 0, 0, 0, 0}
	a2 := []uint64{0xFFFFFFFFFFF, 0, 0, 0, 0, 0, 0, 0} // 44 bits
	b2 := []uint64{0xFFF, 0, 0, 0, 0, 0, 0, 0}         // 12 bits
	result2 := make([]uint64, 8)
	// Product = 0xFFFFFFFFFFF * 0xFFF = 56 bits, should be truncated to 52 bits

	testVPMADD52LUQ(&acc2[0], &a2[0], &b2[0], &result2[0])

	product := a2[0] * b2[0]
	expectedLow := product & 0xFFFFFFFFFFFFF // low 52 bits
	t.Logf("VPMADD52LUQ large test:")
	t.Logf("  a * b = 0x%X", product)
	t.Logf("  low 52 bits = 0x%X", expectedLow)
	t.Logf("  result = 0x%X", result2[0])

	if result2[0] != expectedLow {
		t.Errorf("VPMADD52LUQ large: expected 0x%X, got 0x%X", expectedLow, result2[0])
	}
}

// TestVPMADD52HUQOperandOrder verifies the operand ordering for VPMADD52HUQ
// VPMADD52HUQ computes: dest = dest + (src1 * src2)[103:52] (high 52 bits, shifted)
func TestVPMADD52HUQOperandOrder(t *testing.T) {
	if !cpu.SupportAVX512IFMA {
		t.Skip("IFMA not supported on this CPU")
	}

	// Test VPMADD52HUQ: acc += (a * b) >> 52 (high bits of 104-bit product)
	// Use values that produce a product > 52 bits to see the high part

	acc := []uint64{0, 0, 0, 0, 0, 0, 0, 0}
	a := []uint64{0xFFFFFFFFFFF, 0, 0, 0, 0, 0, 0, 0} // 44 bits
	b := []uint64{0xFFF, 0, 0, 0, 0, 0, 0, 0}         // 12 bits
	result := make([]uint64, 8)
	// Product = 0xFFFFFFFFFFF * 0xFFF = 56 bits

	testVPMADD52HUQ(&acc[0], &a[0], &b[0], &result[0])

	product := a[0] * b[0]
	expectedHigh := product >> 52 // high bits
	t.Logf("VPMADD52HUQ test:")
	t.Logf("  a * b = 0x%X", product)
	t.Logf("  high bits (>>52) = 0x%X", expectedHigh)
	t.Logf("  result = 0x%X", result[0])

	if result[0] != expectedHigh {
		t.Errorf("VPMADD52HUQ: expected 0x%X, got 0x%X", expectedHigh, result[0])
	}
}

//go:noescape
func testVPMADD52LUQ(acc, a, b, result *uint64)

//go:noescape
func testVPMADD52HUQ(acc, a, b, result *uint64)

// TestVPERMQOperandOrder verifies the operand ordering for VPERMQ
// This is a minimal test to determine correct Plan9 operand order.
func TestVPERMQOperandOrder(t *testing.T) {
	if !cpu.SupportAVX512 {
		t.Skip("AVX-512 not supported on this CPU")
	}

	// Input: [0, 1, 2, 3, 4, 5, 6, 7]
	// Permute to: [0, 2, 1, 3, 4, 6, 5, 7] (swap positions 1<->2 and 5<->6)
	// Index vector: [0, 2, 1, 3, 4, 6, 5, 7]

	input := []uint64{10, 20, 30, 40, 50, 60, 70, 80}
	output := make([]uint64, 8)

	testVPERMQ(&input[0], &output[0])

	t.Logf("Input:  %v", input)
	t.Logf("Output: %v", output)

	// With index [0, 2, 1, 3, 4, 6, 5, 7], we expect:
	// output[0] = input[0] = 10
	// output[1] = input[2] = 30
	// output[2] = input[1] = 20
	// output[3] = input[3] = 40
	// output[4] = input[4] = 50
	// output[5] = input[6] = 70
	// output[6] = input[5] = 60
	// output[7] = input[7] = 80
	expected := []uint64{10, 30, 20, 40, 50, 70, 60, 80}

	for i := 0; i < 8; i++ {
		if output[i] != expected[i] {
			t.Errorf("Mismatch at position %d: expected %d, got %d", i, expected[i], output[i])
		}
	}
}

//go:noescape
func testVPERMQ(in, out *uint64)

// TestTransposeAoSToSoA tests the 8x4 matrix transpose from Array-of-Structures
// to Structure-of-Arrays format.
//
// Input (8 elements, each with 4 uint64 limbs):
//
//	Memory: [e0.a0, e0.a1, e0.a2, e0.a3, e1.a0, e1.a1, e1.a2, e1.a3, ...]
//
// Output (4 registers, each with 8 uint64 values):
//
//	Z0 = [e0.a0, e1.a0, e2.a0, e3.a0, e4.a0, e5.a0, e6.a0, e7.a0]
//	Z1 = [e0.a1, e1.a1, e2.a1, e3.a1, e4.a1, e5.a1, e6.a1, e7.a1]
//	Z2 = [e0.a2, e1.a2, e2.a2, e3.a2, e4.a2, e5.a2, e6.a2, e7.a2]
//	Z3 = [e0.a3, e1.a3, e2.a3, e3.a3, e4.a3, e5.a3, e6.a3, e7.a3]
func TestTransposeAoSToSoA(t *testing.T) {
	if !cpu.SupportAVX512 {
		t.Skip("AVX-512 not supported on this CPU")
	}

	// Create test data: 8 elements with recognizable patterns
	// Element i has limbs [i*4+0, i*4+1, i*4+2, i*4+3]
	input := make([]uint64, 32) // 8 elements * 4 limbs
	for i := 0; i < 8; i++ {
		for j := 0; j < 4; j++ {
			input[i*4+j] = uint64(i*4 + j)
		}
	}

	t.Logf("Input (AoS format):")
	for i := 0; i < 8; i++ {
		t.Logf("  Element %d: [%d, %d, %d, %d]", i,
			input[i*4+0], input[i*4+1], input[i*4+2], input[i*4+3])
	}

	// Expected output after transpose (SoA format)
	expectedSoA := make([][]uint64, 4)
	for j := 0; j < 4; j++ {
		expectedSoA[j] = make([]uint64, 8)
		for i := 0; i < 8; i++ {
			expectedSoA[j][i] = uint64(i*4 + j)
		}
	}

	t.Logf("Expected output (SoA format):")
	for j := 0; j < 4; j++ {
		t.Logf("  Limb %d across all elements: %v", j, expectedSoA[j])
	}

	// Call the transpose test function (to be implemented in assembly)
	output := make([]uint64, 32)
	testTransposeAoSToSoA(&input[0], &output[0])

	// Verify output
	t.Logf("Actual output (SoA format):")
	for j := 0; j < 4; j++ {
		actual := output[j*8 : (j+1)*8]
		t.Logf("  Limb %d: %v", j, actual)
		for i := 0; i < 8; i++ {
			if actual[i] != expectedSoA[j][i] {
				t.Errorf("Mismatch at limb %d, element %d: expected %d, got %d",
					j, i, expectedSoA[j][i], actual[i])
			}
		}
	}
}

// TestTransposeSoAToAoS tests the 4x8 matrix transpose from Structure-of-Arrays
// back to Array-of-Structures format (reverse of TestTransposeAoSToSoA).
func TestTransposeSoAToAoS(t *testing.T) {
	if !cpu.SupportAVX512 {
		t.Skip("AVX-512 not supported on this CPU")
	}

	// Create test data in SoA format
	// Limb j has values [0+j, 4+j, 8+j, 12+j, 16+j, 20+j, 24+j, 28+j]
	input := make([]uint64, 32)
	for j := 0; j < 4; j++ {
		for i := 0; i < 8; i++ {
			input[j*8+i] = uint64(i*4 + j)
		}
	}

	t.Logf("Input (SoA format):")
	for j := 0; j < 4; j++ {
		t.Logf("  Limb %d: %v", j, input[j*8:(j+1)*8])
	}

	// Expected output after reverse transpose (AoS format)
	expectedAoS := make([]uint64, 32)
	for i := 0; i < 8; i++ {
		for j := 0; j < 4; j++ {
			expectedAoS[i*4+j] = uint64(i*4 + j)
		}
	}

	t.Logf("Expected output (AoS format):")
	for i := 0; i < 8; i++ {
		t.Logf("  Element %d: [%d, %d, %d, %d]", i,
			expectedAoS[i*4+0], expectedAoS[i*4+1], expectedAoS[i*4+2], expectedAoS[i*4+3])
	}

	// Call the reverse transpose test function
	output := make([]uint64, 32)
	testTransposeSoAToAoS(&input[0], &output[0])

	// Verify output
	t.Logf("Actual output (AoS format):")
	for i := 0; i < 8; i++ {
		actual := output[i*4 : (i+1)*4]
		t.Logf("  Element %d: %v", i, actual)
		for j := 0; j < 4; j++ {
			if actual[j] != expectedAoS[i*4+j] {
				t.Errorf("Mismatch at element %d, limb %d: expected %d, got %d",
					i, j, expectedAoS[i*4+j], actual[j])
			}
		}
	}
}

// TestTransposeRoundTrip verifies that AoS->SoA->AoS is identity
func TestTransposeRoundTrip(t *testing.T) {
	if !cpu.SupportAVX512 {
		t.Skip("AVX-512 not supported on this CPU")
	}

	// Create random-ish test data
	input := make([]uint64, 32)
	for i := range input {
		input[i] = uint64(i*7 + 13) // arbitrary pattern
	}

	t.Logf("Original input:")
	for i := 0; i < 8; i++ {
		t.Logf("  Element %d: [%d, %d, %d, %d]", i,
			input[i*4+0], input[i*4+1], input[i*4+2], input[i*4+3])
	}

	// AoS -> SoA
	intermediate := make([]uint64, 32)
	testTransposeAoSToSoA(&input[0], &intermediate[0])

	// SoA -> AoS
	output := make([]uint64, 32)
	testTransposeSoAToAoS(&intermediate[0], &output[0])

	// Verify round-trip
	t.Logf("After round-trip:")
	for i := 0; i < 8; i++ {
		t.Logf("  Element %d: [%d, %d, %d, %d]", i,
			output[i*4+0], output[i*4+1], output[i*4+2], output[i*4+3])
	}

	for i := range input {
		if input[i] != output[i] {
			t.Errorf("Round-trip mismatch at index %d: expected %d, got %d",
				i, input[i], output[i])
		}
	}
}

// Assembly function declarations for transpose tests
//
//go:noescape
func testTransposeAoSToSoA(in, out *uint64)

//go:noescape
func testTransposeSoAToAoS(in, out *uint64)

//go:noescape
func testRadix52RoundTrip(in, out *uint64)

// TestRadix52RoundTrip tests the complete radix-52 conversion round-trip
// including transpose to SoA, radix conversion, and back.
func TestRadix52RoundTrip(t *testing.T) {
	if !cpu.SupportAVX512 {
		t.Skip("AVX-512 not supported on this CPU")
	}

	// Create test data: 8 elements with recognizable patterns
	// Use actual field element values (Montgomery form of 1)
	input := make([]uint64, 32) // 8 elements * 4 limbs
	var one Element
	one.SetOne()
	for i := 0; i < 8; i++ {
		input[i*4+0] = one[0]
		input[i*4+1] = one[1]
		input[i*4+2] = one[2]
		input[i*4+3] = one[3]
	}

	t.Logf("Input (8 copies of Montgomery form of 1):")
	for i := 0; i < 8; i++ {
		t.Logf("  Element %d: [%016x, %016x, %016x, %016x]", i,
			input[i*4+0], input[i*4+1], input[i*4+2], input[i*4+3])
	}

	output := make([]uint64, 32)
	testRadix52RoundTrip(&input[0], &output[0])

	t.Logf("Output after radix-52 round-trip:")
	for i := 0; i < 8; i++ {
		t.Logf("  Element %d: [%016x, %016x, %016x, %016x]", i,
			output[i*4+0], output[i*4+1], output[i*4+2], output[i*4+3])
	}

	// Verify round-trip preserves data
	for i := 0; i < 32; i++ {
		if input[i] != output[i] {
			t.Errorf("Radix-52 round-trip mismatch at index %d: expected %016x, got %016x",
				i, input[i], output[i])
		}
	}
}
