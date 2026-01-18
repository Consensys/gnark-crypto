// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

//go:build !purego

package fr

import (
	"fmt"
	"testing"

	"github.com/consensys/gnark-crypto/utils/cpu"
)

// BenchmarkMulVecIFMAComparison benchmarks vector multiplication operations
// to provide a baseline for comparing against IFMA implementation.
//
// Run with: go test -bench=BenchmarkMulVecIFMAComparison -benchmem ./ecc/bls12-377/fr/
func BenchmarkMulVecIFMAComparison(b *testing.B) {
	b.Logf("CPU Features: AVX512=%v, AVX512IFMA=%v", cpu.SupportAVX512, cpu.SupportAVX512IFMA)

	const N = 1 << 20 // 1M elements
	a := make(Vector, N)
	bvec := make(Vector, N)
	c := make(Vector, N)

	// Initialize with random values
	var mixer Element
	mixer.MustSetRandom()
	for i := range N {
		a[i].SetUint64(uint64(i+1)).Mul(&a[i], &mixer)
		bvec[i].SetUint64(uint64(N-i)).Mul(&bvec[i], &mixer)
	}

	// Benchmark various sizes
	sizes := []int{8, 16, 32, 64, 128, 256, 512, 1024, 4096, 16384, 65536, 262144, 1 << 20}

	for _, n := range sizes {
		if n > N {
			continue
		}

		b.Run(fmt.Sprintf("current_mul_%d", n), func(b *testing.B) {
			_a := a[:n]
			_b := bvec[:n]
			_c := c[:n]
			b.ResetTimer()
			for range b.N {
				_c.Mul(_a, _b)
			}
			b.ReportMetric(float64(n)/float64(b.Elapsed().Nanoseconds())*1e9, "elem/s")
		})

		// Also benchmark scalar multiplication for comparison
		b.Run(fmt.Sprintf("current_scalarMul_%d", n), func(b *testing.B) {
			_a := a[:n]
			_c := c[:n]
			b.ResetTimer()
			for range b.N {
				_c.ScalarMul(_a, &mixer)
			}
			b.ReportMetric(float64(n)/float64(b.Elapsed().Nanoseconds())*1e9, "elem/s")
		})
	}
}

// BenchmarkMulVecGenericVsAVX512 compares the generic Go implementation
// against the AVX-512 optimized version to measure the current speedup.
func BenchmarkMulVecGenericVsAVX512(b *testing.B) {
	b.Logf("CPU Features: AVX512=%v, AVX512IFMA=%v", cpu.SupportAVX512, cpu.SupportAVX512IFMA)

	const N = 1 << 16 // 64K elements
	a := make(Vector, N)
	bvec := make(Vector, N)
	cGeneric := make(Vector, N)
	cAVX512 := make(Vector, N)

	// Initialize with random values
	var mixer Element
	mixer.MustSetRandom()
	for i := range N {
		a[i].SetUint64(uint64(i+1)).Mul(&a[i], &mixer)
		bvec[i].SetUint64(uint64(N-i)).Mul(&bvec[i], &mixer)
	}

	sizes := []int{16, 64, 256, 1024, 4096, 16384, 65536}

	for _, n := range sizes {
		if n > N {
			continue
		}

		// Benchmark generic implementation
		b.Run(fmt.Sprintf("generic_%d", n), func(b *testing.B) {
			_a := a[:n]
			_b := bvec[:n]
			_c := cGeneric[:n]
			b.ResetTimer()
			for range b.N {
				mulVecGeneric(_c, _a, _b)
			}
			b.ReportMetric(float64(n)/float64(b.Elapsed().Nanoseconds())*1e9, "elem/s")
		})

		// Benchmark AVX-512 implementation (via Mul which uses it when available)
		if cpu.SupportAVX512 {
			b.Run(fmt.Sprintf("avx512_%d", n), func(b *testing.B) {
				_a := a[:n]
				_b := bvec[:n]
				_c := cAVX512[:n]
				b.ResetTimer()
				for range b.N {
					_c.Mul(_a, _b)
				}
				b.ReportMetric(float64(n)/float64(b.Elapsed().Nanoseconds())*1e9, "elem/s")
			})
		}
	}
}

// BenchmarkIFMAvsGeneric directly compares IFMA vs generic implementations
func BenchmarkIFMAvsGeneric(b *testing.B) {
	if !cpu.SupportAVX512IFMA {
		b.Skip("IFMA not supported on this CPU")
	}

	b.Logf("CPU Features: AVX512=%v, AVX512IFMA=%v", cpu.SupportAVX512, cpu.SupportAVX512IFMA)

	const N = 1 << 16 // 64K elements
	a := make(Vector, N)
	bvec := make(Vector, N)
	cIFMA := make(Vector, N)
	cGeneric := make(Vector, N)

	// Initialize with random values
	var mixer Element
	mixer.MustSetRandom()
	for i := range N {
		a[i].SetUint64(uint64(i+1)).Mul(&a[i], &mixer)
		bvec[i].SetUint64(uint64(N-i)).Mul(&bvec[i], &mixer)
	}

	sizes := []int{8, 16, 32, 64, 128, 256, 512, 1024, 4096, 16384, 65536}

	for _, n := range sizes {
		if n > N {
			continue
		}

		// Benchmark IFMA implementation directly
		b.Run(fmt.Sprintf("ifma_%d", n), func(b *testing.B) {
			_a := a[:n]
			_b := bvec[:n]
			_c := cIFMA[:n]
			nGroups := uint64(n / 8)
			b.ResetTimer()
			for range b.N {
				mulVec(&_c[0], &_a[0], &_b[0], nGroups)
			}
			b.ReportMetric(float64(n)/float64(b.Elapsed().Nanoseconds())*1e9, "elem/s")
		})

		// Benchmark generic implementation for comparison
		b.Run(fmt.Sprintf("generic_%d", n), func(b *testing.B) {
			_a := a[:n]
			_b := bvec[:n]
			_c := cGeneric[:n]
			b.ResetTimer()
			for range b.N {
				mulVecGeneric(_c, _a, _b)
			}
			b.ReportMetric(float64(n)/float64(b.Elapsed().Nanoseconds())*1e9, "elem/s")
		})
	}
}

// BenchmarkSingleMulThroughput measures the throughput of single element multiplication
// to understand the baseline cost without vectorization overhead.
func BenchmarkSingleMulThroughput(b *testing.B) {
	var a, bval, c Element
	a.MustSetRandom()
	bval.MustSetRandom()

	b.Run("single_mul", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			c.Mul(&a, &bval)
		}
	})

	// Measure 8 sequential multiplications (what IFMA would do in parallel)
	var a8, b8, c8 [8]Element
	for i := range 8 {
		a8[i].MustSetRandom()
		b8[i].MustSetRandom()
	}

	b.Run("sequential_8_muls", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			for j := range 8 {
				c8[j].Mul(&a8[j], &b8[j])
			}
		}
		b.ReportMetric(float64(8)/float64(b.Elapsed().Nanoseconds())*1e9, "elem/s")
	})
}
