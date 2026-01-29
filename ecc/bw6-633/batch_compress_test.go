// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package bw6633

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bw6-633/fp"
	"github.com/consensys/gnark-crypto/ecc/bw6-633/fr"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

// randomG1 generates a random G1Affine point via scalar multiplication
func randomG1() G1Affine {
	var s fr.Element
	s.SetRandom()
	var sInt big.Int
	s.BigInt(&sInt)
	var p G1Affine
	p.ScalarMultiplication(&g1GenAff, &sInt)
	return p
}

// randomG2 generates a random G2Affine point via scalar multiplication
func randomG2() G2Affine {
	var s fr.Element
	s.SetRandom()
	var sInt big.Int
	s.BigInt(&sInt)
	var p G2Affine
	p.ScalarMultiplication(&g2GenAff, &sInt)
	return p
}

func TestBatchCompress2G1(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Test random point pairs
	properties.Property("BatchCompress2G1/BatchDecompress2G1 roundtrip (random points)", prop.ForAll(
		func() bool {
			p0 := randomG1()
			p1 := randomG1()

			z0, z1, flags, err := BatchCompress2G1(&p0, &p1)
			if err != nil {
				t.Logf("compression failed: %v", err)
				return false
			}

			q0, q1, err := BatchDecompress2G1(z0, z1, flags)
			if err != nil {
				t.Logf("decompression failed: %v", err)
				return false
			}

			return p0.Equal(&q0) && p1.Equal(&q1)
		},
	))

	// Test with infinity points
	properties.Property("BatchCompress2G1/BatchDecompress2G1 (both infinity)", prop.ForAll(
		func() bool {
			var p0, p1 G1Affine
			p0.SetInfinity()
			p1.SetInfinity()

			z0, z1, flags, err := BatchCompress2G1(&p0, &p1)
			if err != nil {
				return false
			}

			q0, q1, err := BatchDecompress2G1(z0, z1, flags)
			if err != nil {
				return false
			}

			return q0.IsInfinity() && q1.IsInfinity()
		},
	))

	properties.Property("BatchCompress2G1/BatchDecompress2G1 (p0 infinity)", prop.ForAll(
		func() bool {
			var p0 G1Affine
			p0.SetInfinity()
			p1 := randomG1()

			z0, z1, flags, err := BatchCompress2G1(&p0, &p1)
			if err != nil {
				return false
			}

			q0, q1, err := BatchDecompress2G1(z0, z1, flags)
			if err != nil {
				return false
			}

			return q0.IsInfinity() && p1.Equal(&q1)
		},
	))

	properties.Property("BatchCompress2G1/BatchDecompress2G1 (p1 infinity)", prop.ForAll(
		func() bool {
			p0 := randomG1()
			var p1 G1Affine
			p1.SetInfinity()

			z0, z1, flags, err := BatchCompress2G1(&p0, &p1)
			if err != nil {
				return false
			}

			q0, q1, err := BatchDecompress2G1(z0, z1, flags)
			if err != nil {
				return false
			}

			return p0.Equal(&q0) && q1.IsInfinity()
		},
	))

	// Test same point
	properties.Property("BatchCompress2G1/BatchDecompress2G1 (same point)", prop.ForAll(
		func() bool {
			p0 := randomG1()
			p1 := p0

			z0, z1, flags, err := BatchCompress2G1(&p0, &p1)
			if err != nil {
				return false
			}

			q0, q1, err := BatchDecompress2G1(z0, z1, flags)
			if err != nil {
				return false
			}

			return p0.Equal(&q0) && p0.Equal(&q1)
		},
	))

	// Test negation
	properties.Property("BatchCompress2G1/BatchDecompress2G1 (negation)", prop.ForAll(
		func() bool {
			p0 := randomG1()
			var p1 G1Affine
			p1.Neg(&p0)

			z0, z1, flags, err := BatchCompress2G1(&p0, &p1)
			if err != nil {
				return false
			}

			q0, q1, err := BatchDecompress2G1(z0, z1, flags)
			if err != nil {
				return false
			}

			return p0.Equal(&q0) && p1.Equal(&q1)
		},
	))

	// Test automorphism (y0² == y1²)
	properties.Property("BatchCompress2G1/BatchDecompress2G1 (automorphism)", prop.ForAll(
		func() bool {
			p0 := randomG1()
			// Apply ω automorphism: (x, y) -> (ω·x, -y)
			var p1 G1Affine
			p1.X.Mul(&p0.X, &thirdRootOneG1)
			p1.Y.Neg(&p0.Y)

			z0, z1, flags, err := BatchCompress2G1(&p0, &p1)
			if err != nil {
				t.Logf("compression failed: %v", err)
				return false
			}

			q0, q1, err := BatchDecompress2G1(z0, z1, flags)
			if err != nil {
				t.Logf("decompression failed: %v", err)
				return false
			}

			return p0.Equal(&q0) && p1.Equal(&q1)
		},
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestBatchCompress2G2(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Test random point pairs
	properties.Property("BatchCompress2G2/BatchDecompress2G2 roundtrip (random points)", prop.ForAll(
		func() bool {
			p0 := randomG2()
			p1 := randomG2()

			z0, z1, flags, err := BatchCompress2G2(&p0, &p1)
			if err != nil {
				t.Logf("compression failed: %v", err)
				return false
			}

			q0, q1, err := BatchDecompress2G2(z0, z1, flags)
			if err != nil {
				t.Logf("decompression failed: %v", err)
				return false
			}

			return p0.Equal(&q0) && p1.Equal(&q1)
		},
	))

	// Test with infinity points
	properties.Property("BatchCompress2G2/BatchDecompress2G2 (both infinity)", prop.ForAll(
		func() bool {
			var p0, p1 G2Affine
			p0.SetInfinity()
			p1.SetInfinity()

			z0, z1, flags, err := BatchCompress2G2(&p0, &p1)
			if err != nil {
				return false
			}

			q0, q1, err := BatchDecompress2G2(z0, z1, flags)
			if err != nil {
				return false
			}

			return q0.IsInfinity() && q1.IsInfinity()
		},
	))

	// Test same point
	properties.Property("BatchCompress2G2/BatchDecompress2G2 (same point)", prop.ForAll(
		func() bool {
			p0 := randomG2()
			p1 := p0

			z0, z1, flags, err := BatchCompress2G2(&p0, &p1)
			if err != nil {
				return false
			}

			q0, q1, err := BatchDecompress2G2(z0, z1, flags)
			if err != nil {
				return false
			}

			return p0.Equal(&q0) && p0.Equal(&q1)
		},
	))

	// Test negation
	properties.Property("BatchCompress2G2/BatchDecompress2G2 (negation)", prop.ForAll(
		func() bool {
			p0 := randomG2()
			var p1 G2Affine
			p1.Neg(&p0)

			z0, z1, flags, err := BatchCompress2G2(&p0, &p1)
			if err != nil {
				return false
			}

			q0, q1, err := BatchDecompress2G2(z0, z1, flags)
			if err != nil {
				return false
			}

			return p0.Equal(&q0) && p1.Equal(&q1)
		},
	))

	// Test automorphism for G2 (using thirdRootOneG2)
	properties.Property("BatchCompress2G2/BatchDecompress2G2 (automorphism)", prop.ForAll(
		func() bool {
			p0 := randomG2()
			// Apply ω automorphism for G2: (x, y) -> (ω·x, -y) where ω = thirdRootOneG2
			var p1 G2Affine
			p1.X.Mul(&p0.X, &thirdRootOneG2)
			p1.Y.Neg(&p0.Y)

			z0, z1, flags, err := BatchCompress2G2(&p0, &p1)
			if err != nil {
				t.Logf("compression failed: %v", err)
				return false
			}

			q0, q1, err := BatchDecompress2G2(z0, z1, flags)
			if err != nil {
				t.Logf("decompression failed: %v", err)
				return false
			}

			return p0.Equal(&q0) && p1.Equal(&q1)
		},
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestBatchCompressG1Slice(t *testing.T) {
	// Test various slice lengths
	for _, n := range []int{0, 1, 2, 3, 4, 5, 10, 11, 100, 101} {
		t.Run("", func(t *testing.T) {
			points := make([]G1Affine, n)
			for i := range points {
				points[i] = randomG1()
			}

			compressed, err := BatchCompressG1Slice(points)
			if err != nil {
				t.Fatalf("compression failed: %v", err)
			}

			decompressed, err := BatchDecompressG1Slice(compressed, n)
			if err != nil {
				t.Fatalf("decompression failed: %v", err)
			}

			if len(decompressed) != n {
				t.Fatalf("wrong length: got %d, want %d", len(decompressed), n)
			}

			for i := range points {
				if !points[i].Equal(&decompressed[i]) {
					t.Fatalf("point %d mismatch", i)
				}
			}
		})
	}
}

func TestBatchCompressG2Slice(t *testing.T) {
	// Test various slice lengths
	for _, n := range []int{0, 1, 2, 3, 4, 5, 10, 11, 100, 101} {
		t.Run("", func(t *testing.T) {
			points := make([]G2Affine, n)
			for i := range points {
				points[i] = randomG2()
			}

			compressed, err := BatchCompressG2Slice(points)
			if err != nil {
				t.Fatalf("compression failed: %v", err)
			}

			decompressed, err := BatchDecompressG2Slice(compressed, n)
			if err != nil {
				t.Fatalf("decompression failed: %v", err)
			}

			if len(decompressed) != n {
				t.Fatalf("wrong length: got %d, want %d", len(decompressed), n)
			}

			for i := range points {
				if !points[i].Equal(&decompressed[i]) {
					t.Fatalf("point %d mismatch", i)
				}
			}
		})
	}
}

func BenchmarkBatchCompress2G1(b *testing.B) {
	p0 := randomG1()
	p1 := randomG1()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BatchCompress2G1(&p0, &p1)
	}
}

func BenchmarkBatchDecompress2G1(b *testing.B) {
	p0 := randomG1()
	p1 := randomG1()

	z0, z1, flags, _ := BatchCompress2G1(&p0, &p1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BatchDecompress2G1(z0, z1, flags)
	}
}

func BenchmarkStandardDecompress2G1(b *testing.B) {
	p0 := randomG1()
	p1 := randomG1()

	b0 := p0.Bytes()
	b1 := p1.Bytes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var q0, q1 G1Affine
		q0.SetBytes(b0[:])
		q1.SetBytes(b1[:])
	}
}

func BenchmarkBatchCompressG1Slice(b *testing.B) {
	const n = 1000
	points := make([]G1Affine, n)
	for i := range points {
		points[i] = randomG1()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BatchCompressG1Slice(points)
	}
}

func BenchmarkBatchDecompressG1Slice(b *testing.B) {
	const n = 1000
	points := make([]G1Affine, n)
	for i := range points {
		points[i] = randomG1()
	}

	compressed, _ := BatchCompressG1Slice(points)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BatchDecompressG1Slice(compressed, n)
	}
}

// Unused variables to satisfy compiler
var _ fp.Element
var _ fr.Element
