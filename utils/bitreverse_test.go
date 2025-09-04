// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package utils

import (
	"fmt"
	"testing"
)

const maxSizeBitReverse = 1 << 22

func TestBitReverse(t *testing.T) {
	sizes := []int{2, 4, 8, 16, 32, 64, 128, 256, 512, maxSizeBitReverse}

	t.Run("uint32", func(t *testing.T) {
		for _, size := range sizes {
			t.Run(fmt.Sprintf("size=%d", size), func(t *testing.T) {
				// check that bit-reversing twice is identity
				original := make([]uint32, size)
				for i := range original {
					original[i] = uint32(i)
				}
				a := make([]uint32, size)
				copy(a, original)

				BitReverse(a)
				BitReverse(a)

				for i := range a {
					if a[i] != original[i] {
						t.Fatalf("bit-reversing twice is not identity")
					}
				}
			})
		}

		// check that it panics for non-power of 2
		for _, size := range []int{3, 5, 6, 7, 9, 10, 12} {
			a := make([]uint32, size)
			assertPanic(t, func() {
				BitReverse(a)
			})
		}
	})

	t.Run("[4]uint64", func(t *testing.T) {
		for _, size := range sizes {
			t.Run(fmt.Sprintf("size=%d", size), func(t *testing.T) {
				// check that bit-reversing twice is identity
				original := make([][4]uint64, size)
				for i := range original {
					original[i] = [4]uint64{uint64(i), uint64(i) + 1, uint64(i) + 2, uint64(i) + 3}
				}
				a := make([][4]uint64, size)
				copy(a, original)

				BitReverse(a)
				BitReverse(a)

				for i := range a {
					if a[i] != original[i] {
						t.Fatalf("bit-reversing twice is not identity")
					}
				}
			})
		}

		// check that it panics for non-power of 2
		for _, size := range []int{3, 5, 6, 7, 9, 10, 12} {
			a := make([][4]uint64, size)
			assertPanic(t, func() {
				BitReverse(a)
			})
		}
	})
}

func BenchmarkBitReverse(b *testing.B) {
	sizes := []int{1 << 8, 1 << 9, 1 << 16, 1 << 21, maxSizeBitReverse}

	b.Run("uint32", func(b *testing.B) {
		for _, size := range sizes {
			a := make([]uint32, size)
			for i := range a {
				a[i] = uint32(i)
			}
			b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					BitReverse(a)
				}
			})
		}
	})

	b.Run("[4]uint64", func(b *testing.B) {
		for _, size := range sizes {
			a := make([][4]uint64, size)
			for i := range a {
				a[i] = [4]uint64{uint64(i), uint64(i) + 1, uint64(i) + 2, uint64(i) + 3}
			}
			b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					BitReverse(a)
				}
			})
		}
	})
}

// / test helpers
func assertPanic(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected a panic")
		}
	}()
	f()
}
