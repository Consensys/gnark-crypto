// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package koalabear

import (
	"bytes"
	"fmt"
	"testing"
)

// FuzzVectorOperations tests all vector operations through Go's fuzzing framework.
func FuzzVectorOperations(f *testing.F) {
	// Add some seed corpus entries
	f.Add([]byte{0x01, 0x02, 0x03, 0x04})                         // Small vector
	f.Add([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}) // Medium vector

	// Start fuzzing
	f.Fuzz(func(t *testing.T, data []byte) {
		// Need at least enough bytes for a small vector
		if len(data) < 8 {
			return
		}

		// Use first byte to determine vector size
		size := int(data[0]) % 33
		if size == 0 {
			size = 1 // Ensure at least size 1
		}

		// Need enough data for the vectors
		if len(data) < 1+size*8 {
			return
		}

		// Create vectors
		a := make(Vector, size)
		b := make(Vector, size)
		result := make(Vector, size)

		// Fill vectors with data from fuzzer
		r := bytes.NewReader(data[1:])
		for i := 0; i < size; i++ {
			// Read 4 bytes for each element
			var buf [4]byte
			if _, err := r.Read(buf[:]); err != nil {
				return
			}
			a[i].SetBytes(buf[:])

			if _, err := r.Read(buf[:]); err != nil {
				return
			}
			b[i].SetBytes(buf[:])
		}

		// Test 1: Add
		result.Add(a, b)
		for i := 0; i < size; i++ {
			var expected Element
			expected.Add(&a[i], &b[i])
			if !expected.Equal(&result[i]) {
				t.Fatalf("Add failed at index %d: expected %s, got %s", i, expected.String(), result[i].String())
			}
		}

		// Test 2: Sub
		result.Sub(a, b)
		for i := 0; i < size; i++ {
			var expected Element
			expected.Sub(&a[i], &b[i])
			if !expected.Equal(&result[i]) {
				t.Fatalf("Sub failed at index %d: expected %s, got %s", i, expected.String(), result[i].String())
			}
		}

		// Test 3: Mul
		result.Mul(a, b)
		for i := 0; i < size; i++ {
			var expected Element
			expected.Mul(&a[i], &b[i])
			if !expected.Equal(&result[i]) {
				t.Fatalf("Mul failed at index %d: expected %s, got %s", i, expected.String(), result[i].String())
			}
		}

		// Test 4: ScalarMul
		// Use first element of b as scalar
		scalar := b[0]
		result.ScalarMul(a, &scalar)
		for i := 0; i < size; i++ {
			var expected Element
			expected.Mul(&a[i], &scalar)
			if !expected.Equal(&result[i]) {
				t.Fatalf("ScalarMul failed at index %d: expected %s, got %s", i, expected.String(), result[i].String())
			}
		}

		// Test 5: Sum
		computed := a.Sum()
		var expected Element
		for i := 0; i < size; i++ {
			expected.Add(&expected, &a[i])
		}
		if !expected.Equal(&computed) {
			t.Fatalf("Sum failed: expected %s, got %s", expected.String(), computed.String())
		}

		// Test 6: InnerProduct
		computed = a.InnerProduct(b)
		expected.SetZero()
		for i := 0; i < size; i++ {
			var temp Element
			temp.Mul(&a[i], &b[i])
			expected.Add(&expected, &temp)
		}
		if !expected.Equal(&computed) {
			t.Fatalf("InnerProduct failed: expected %s, got %s", expected.String(), computed.String())
		}
	})
}

// TestFuzzVectorManual tests vector operations with predefined inputs
func TestFuzzVectorManual(t *testing.T) {
	testCases := [][]byte{
		{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, // Size 1
		{0x04, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
			0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
			0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f,
			0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3a, 0x3b, 0x3c, 0x3d, 0x3e, 0x3f,
		}, // Size 4
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case_%d", i), func(t *testing.T) {
			// Manually replicate the body of the fuzz function
			data := tc

			// Need at least enough bytes for a small vector
			if len(data) < 8 {
				t.Skip("Not enough data")
				return
			}

			// Use first byte to determine vector size
			size := int(data[0]) % 33
			if size == 0 {
				size = 1 // Ensure at least size 1
			}

			// Need enough data for the vectors
			if len(data) < 1+size*8 {
				t.Skip("Not enough data for vectors")
				return
			}

			// Create vectors
			a := make(Vector, size)
			b := make(Vector, size)
			result := make(Vector, size)

			// Fill vectors with data from test case
			r := bytes.NewReader(data[1:])
			for i := 0; i < size; i++ {
				// Read 4 bytes for each element
				var buf [4]byte
				if _, err := r.Read(buf[:]); err != nil {
					t.Skip("Error reading from buffer")
					return
				}
				a[i].SetBytes(buf[:])

				if _, err := r.Read(buf[:]); err != nil {
					t.Skip("Error reading from buffer")
					return
				}
				b[i].SetBytes(buf[:])
			}

			// Test 1: Add
			result.Add(a, b)
			for i := 0; i < size; i++ {
				var expected Element
				expected.Add(&a[i], &b[i])
				if !expected.Equal(&result[i]) {
					t.Fatalf("Add failed at index %d: expected %s, got %s", i, expected.String(), result[i].String())
				}
			}

			// Test 2: Sub
			result.Sub(a, b)
			for i := 0; i < size; i++ {
				var expected Element
				expected.Sub(&a[i], &b[i])
				if !expected.Equal(&result[i]) {
					t.Fatalf("Sub failed at index %d: expected %s, got %s", i, expected.String(), result[i].String())
				}
			}

			// Test 3: Mul
			result.Mul(a, b)
			for i := 0; i < size; i++ {
				var expected Element
				expected.Mul(&a[i], &b[i])
				if !expected.Equal(&result[i]) {
					t.Fatalf("Mul failed at index %d: expected %s, got %s", i, expected.String(), result[i].String())
				}
			}

			// Test 4: ScalarMul
			// Use first element of b as scalar
			scalar := b[0]
			result.ScalarMul(a, &scalar)
			for i := 0; i < size; i++ {
				var expected Element
				expected.Mul(&a[i], &scalar)
				if !expected.Equal(&result[i]) {
					t.Fatalf("ScalarMul failed at index %d: expected %s, got %s", i, expected.String(), result[i].String())
				}
			}

			// Test 5: Sum
			computed := a.Sum()
			var expected Element
			for i := 0; i < size; i++ {
				expected.Add(&expected, &a[i])
			}
			if !expected.Equal(&computed) {
				t.Fatalf("Sum failed: expected %s, got %s", expected.String(), computed.String())
			}

			// Test 6: InnerProduct
			computed = a.InnerProduct(b)
			expected.SetZero()
			for i := 0; i < size; i++ {
				var temp Element
				temp.Mul(&a[i], &b[i])
				expected.Add(&expected, &temp)
			}
			if !expected.Equal(&computed) {
				t.Fatalf("InnerProduct failed: expected %s, got %s", expected.String(), computed.String())
			}
		})
	}
}
